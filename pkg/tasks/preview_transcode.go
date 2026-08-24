package tasks

import (
	"bytes"
	"context"
	"fmt"
	"os"
	"runtime"
	"strconv"
	"strings"
	"sync"
	"time"

	"github.com/darwayne/go-timecode/timecode"
)

// ffmpegCapabilities describes the hardware-relevant features advertised by a
// specific ffmpeg binary. Advertised features only prove build support; actual
// usability is established by runtime probes.
type ffmpegCapabilities struct {
	path     string
	hwaccels string
	decoders string
	encoders string
	filters  string
}

func (c *ffmpegCapabilities) hasHwaccel(name string) bool { return hasToken(c.hwaccels, name) }
func (c *ffmpegCapabilities) hasDecoder(name string) bool { return hasToken(c.decoders, name) }
func (c *ffmpegCapabilities) hasEncoder(name string) bool { return hasToken(c.encoders, name) }
func (c *ffmpegCapabilities) hasFilter(name string) bool  { return hasToken(c.filters, name) }

// hasToken reports whether name occurs as a whitespace-separated token in the
// given ffmpeg listing output.
func hasToken(output, name string) bool {
	for _, field := range strings.Fields(output) {
		if field == name {
			return true
		}
	}
	return false
}

// queryFFmpeg runs an ffmpeg information command and returns its combined
// output. It is a variable so tests can inject fixture output.
var queryFFmpeg = func(path string, args ...string) (string, error) {
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()
	cmd := buildCmdContext(ctx, path, args...)
	var out bytes.Buffer
	cmd.Stdout = &out
	cmd.Stderr = &out
	if err := cmd.Run(); err != nil {
		return "", fmt.Errorf("ffmpeg %s failed: %w", strings.Join(args, " "), err)
	}
	return out.String(), nil
}

var (
	capsMu    sync.Mutex
	capsCache = map[string]*ffmpegCapabilities{}
)

// detectFFmpegCapabilities inspects the given ffmpeg binary once per path and
// caches the result.
func detectFFmpegCapabilities(path string) (*ffmpegCapabilities, error) {
	capsMu.Lock()
	defer capsMu.Unlock()

	if caps, ok := capsCache[path]; ok {
		return caps, nil
	}

	caps := &ffmpegCapabilities{path: path}
	queries := []struct {
		args []string
		dest *string
	}{
		{[]string{"-hide_banner", "-hwaccels"}, &caps.hwaccels},
		{[]string{"-hide_banner", "-decoders"}, &caps.decoders},
		{[]string{"-hide_banner", "-encoders"}, &caps.encoders},
		{[]string{"-hide_banner", "-filters"}, &caps.filters},
	}
	for _, q := range queries {
		out, err := queryFFmpeg(path, q.args...)
		if err != nil {
			return nil, err
		}
		*q.dest = out
	}

	capsCache[path] = caps
	return caps, nil
}

// ffmpegSnippetTimeout bounds a single ffmpeg invocation (probe or snippet
// render). Snippets are a few seconds of video, so this is generous; it
// exists so a hung ffmpeg cannot stall the preview queue forever.
const ffmpegSnippetTimeout = 5 * time.Minute

// runFFmpeg executes ffmpeg at lowered process priority and returns an error
// including a bounded stderr tail on failure. It is a variable so tests can
// inject failures. The process is killed when it exceeds
// ffmpegSnippetTimeout or when the user requests the preview queue to stop.
var runFFmpeg = func(path string, args ...string) error {
	ctx, cancel := context.WithTimeout(context.Background(), ffmpegSnippetTimeout)
	defer cancel()

	// Cancel early when the user stops preview generation, so a running
	// ffmpeg does not block the queue shutdown.
	stopWatch := make(chan struct{})
	defer close(stopWatch)
	go func() {
		ticker := time.NewTicker(500 * time.Millisecond)
		defer ticker.Stop()
		for {
			select {
			case <-stopWatch:
				return
			case <-ctx.Done():
				return
			case <-ticker.C:
				if previewStopRequested() {
					cancel()
					return
				}
			}
		}
	}()

	cmd := buildCmdContext(ctx, path, args...)

	var stderr bytes.Buffer
	cmd.Stderr = &stderr
	if err := cmd.Start(); err != nil {
		return fmt.Errorf("ffmpeg failed to start: %w", err)
	}
	lowerProcessPriority(cmd)
	if err := cmd.Wait(); err != nil {
		if ctx.Err() == context.DeadlineExceeded {
			return fmt.Errorf("ffmpeg timed out after %v: %s", ffmpegSnippetTimeout, tailString(stderr.String(), 500))
		}
		if ctx.Err() == context.Canceled {
			return fmt.Errorf("ffmpeg stopped by user request")
		}
		return fmt.Errorf("ffmpeg failed: %w: %s", err, tailString(stderr.String(), 500))
	}
	return nil
}

func tailString(s string, max int) string {
	s = strings.TrimSpace(s)
	if len(s) > max {
		return "..." + s[len(s)-max:]
	}
	return s
}

// cropRect is the numeric pixel rectangle matching the crop expression.
type cropRect struct {
	w, h, x, y int
}

// previewCandidate is one complete transcode pipeline that can render a
// snippet, ordered from most to least hardware accelerated.
type previewCandidate struct {
	name  string
	build func(start time.Duration, dest string) []string
}

// backendDef describes one hardware backend and the pipelines it can provide.
type backendDef struct {
	name          string
	hwaccel       string
	encoder       string
	filter        string // full-GPU crop/scale filter; scale-only unless qsv/vulkan
	fullFlatOnly  bool   // filter cannot crop; full GPU only for identity crops
	partialPixFmt string // software pixel format accepted by the encoder
	codecs        []string
	inputOpts     []string // extra per-input options, e.g. thread limits
	outputOpts    []string // extra encoder/output options, e.g. rate caps
	optInEnv      string   // if set, backend is only used when this env var is truthy
	fullInputArgs func(t *previewTranscoder) []string
	fullFilter    func(t *previewTranscoder) string
}

func (b backendDef) supportsCodec(codec string) bool {
	for _, c := range b.codecs {
		if c == codec {
			return true
		}
	}
	return false
}

var codecSets = struct {
	common       []string
	videotoolbox []string
	vulkan       []string
}{
	common:       []string{"h264", "hevc", "mpeg2video", "vc1", "vp8", "vp9", "av1", "mjpeg"},
	videotoolbox: []string{"h264", "hevc", "mpeg2video", "vp9", "prores"},
	// Vulkan video decode supports far fewer codecs than the other backends.
	vulkan: []string{"h264", "hevc", "av1"},
}

var (
	qsvBackend = backendDef{
		name:          "qsv",
		hwaccel:       "qsv",
		encoder:       "h264_qsv",
		filter:        "vpp_qsv",
		partialPixFmt: "nv12",
		codecs:        codecSets.common,
		fullInputArgs: func(t *previewTranscoder) []string {
			return []string{"-init_hw_device", "qsv=qsv", "-filter_hw_device", "qsv", "-hwaccel", "qsv", "-hwaccel_output_format", "qsv"}
		},
		fullFilter: func(t *previewTranscoder) string {
			return fmt.Sprintf("vpp_qsv=cw=%d:ch=%d:cx=%d:cy=%d:w=%d:h=%d:format=nv12",
				t.rect.w, t.rect.h, t.rect.x, t.rect.y, t.resolution, t.resolution)
		},
	}
	cudaBackend = backendDef{
		name:          "cuda",
		hwaccel:       "cuda",
		encoder:       "h264_nvenc",
		filter:        "scale_cuda",
		fullFlatOnly:  true,
		partialPixFmt: "yuv420p",
		codecs:        codecSets.common,
		fullInputArgs: func(t *previewTranscoder) []string {
			return []string{"-init_hw_device", "cuda=cuda", "-filter_hw_device", "cuda", "-hwaccel", "cuda", "-hwaccel_output_format", "cuda"}
		},
		fullFilter: func(t *previewTranscoder) string {
			return fmt.Sprintf("scale_cuda=w=%d:h=%d:format=yuv420p", t.resolution, t.resolution)
		},
	}
	vaapiBackend = backendDef{
		name:          "vaapi",
		hwaccel:       "vaapi",
		encoder:       "h264_vaapi",
		filter:        "scale_vaapi",
		fullFlatOnly:  true,
		partialPixFmt: "nv12",
		codecs:        codecSets.common,
		fullInputArgs: func(t *previewTranscoder) []string {
			return []string{"-init_hw_device", "vaapi=vaapi:/dev/dri/renderD128", "-filter_hw_device", "vaapi", "-hwaccel", "vaapi", "-hwaccel_output_format", "vaapi"}
		},
		fullFilter: func(t *previewTranscoder) string {
			return fmt.Sprintf("scale_vaapi=w=%d:h=%d:format=nv12", t.resolution, t.resolution)
		},
	}
	videoToolboxBackend = backendDef{
		name:          "videotoolbox",
		hwaccel:       "videotoolbox",
		encoder:       "h264_videotoolbox",
		filter:        "scale_vt",
		fullFlatOnly:  true,
		partialPixFmt: "yuv420p",
		codecs:        codecSets.videotoolbox,
		// Keep the host responsive: cap decoder threads so the media engine
		// queue is not swamped, and bound the encoder rate because
		// VideoToolbox over-allocates buffers on Apple silicon when left
		// unbounded. 2 Mbit/s is generous for a small square H.264 preview.
		inputOpts:  []string{"-threads", "4"},
		outputOpts: []string{"-maxrate", "2M", "-bufsize", "4M"},
		// VideoToolbox offload is opt-in: it needs a modern ffmpeg (the
		// bundled 4.2.1 lacks scale_vt) and benefits depend on the host CPU.
		optInEnv: "XBVR_PREVIEW_VIDEOTOOLBOX",
		fullInputArgs: func(t *previewTranscoder) []string {
			// The hardware pixel format is named videotoolbox_vld in FFmpeg
			// versions new enough to provide scale_vt.
			return []string{"-hwaccel", "videotoolbox", "-hwaccel_output_format", "videotoolbox_vld"}
		},
		fullFilter: func(t *previewTranscoder) string {
			return fmt.Sprintf("scale_vt=w=%d:h=%d", t.resolution, t.resolution)
		},
	}
	amfBackend = backendDef{
		name:          "amf",
		hwaccel:       "d3d11va",
		encoder:       "h264_amf",
		filter:        "vpp_amf",
		fullFlatOnly:  true,
		partialPixFmt: "yuv420p",
		codecs:        codecSets.common,
		fullInputArgs: func(t *previewTranscoder) []string {
			return []string{"-init_hw_device", "d3d11va=dx11", "-init_hw_device", "amf=amf@dx11", "-filter_hw_device", "amf", "-hwaccel", "d3d11va", "-hwaccel_output_format", "d3d11va"}
		},
		fullFilter: func(t *previewTranscoder) string {
			return fmt.Sprintf("vpp_amf=w=%d:h=%d:format=nv12", t.resolution, t.resolution)
		},
	}
	// vulkanBackend is cross-platform but only usable when the host has a
	// working Vulkan driver; the runtime probe rejects it otherwise.
	vulkanBackend = backendDef{
		name:          "vulkan",
		hwaccel:       "vulkan",
		encoder:       "h264_vulkan",
		filter:        "scale_vulkan",
		fullFlatOnly:  true,
		partialPixFmt: "nv12",
		codecs:        codecSets.vulkan,
		fullInputArgs: func(t *previewTranscoder) []string {
			return []string{"-init_hw_device", "vulkan=vk", "-filter_hw_device", "vk", "-hwaccel", "vulkan", "-hwaccel_output_format", "vulkan"}
		},
		fullFilter: func(t *previewTranscoder) string {
			return fmt.Sprintf("scale_vulkan=w=%d:h=%d:format=nv12", t.resolution, t.resolution)
		},
	}
)

// platformBackends returns backends in platform-native priority order.
func platformBackends() []backendDef {
	switch runtime.GOOS {
	case "darwin":
		return []backendDef{videoToolboxBackend, cudaBackend, vulkanBackend}
	case "windows":
		return []backendDef{amfBackend, qsvBackend, cudaBackend, vulkanBackend}
	default:
		return []backendDef{vaapiBackend, qsvBackend, cudaBackend, vulkanBackend}
	}
}

// previewTranscoder renders preview snippets for one input file using the
// first working candidate pipeline, falling back to the next candidate on
// failure.
type previewTranscoder struct {
	ffmpegPath    string
	inputFile     string
	codec         string
	pixFmt        string
	width, height int
	flat          bool
	crop          string
	rect          cropRect
	resolution    int
	snippetLength float64
	startTime     int
	candidates    []previewCandidate
	current       int
}

func newPreviewTranscoder(ffmpegPath, inputFile, codec, pixFmt string, width, height int, flat bool, crop string, rect cropRect, resolution int, snippetLength float64, startTime int) (*previewTranscoder, error) {
	t := &previewTranscoder{
		ffmpegPath:    ffmpegPath,
		inputFile:     inputFile,
		codec:         codec,
		pixFmt:        pixFmt,
		width:         width,
		height:        height,
		flat:          flat,
		crop:          crop,
		rect:          rect,
		resolution:    resolution,
		snippetLength: snippetLength,
		startTime:     startTime,
	}

	caps, err := detectFFmpegCapabilities(ffmpegPath)
	if err != nil {
		log.Warnf("FFmpeg capability detection failed, using software encoding: %v", err)
		caps = &ffmpegCapabilities{path: ffmpegPath}
	}

	t.candidates = t.buildCandidates(caps)

	validated := make([]previewCandidate, 0, len(t.candidates))
	for _, c := range t.candidates {
		if t.probe(c) {
			validated = append(validated, c)
		} else {
			log.Debugf("Preview pipeline %s failed validation", c.name)
		}
	}
	if len(validated) == 0 {
		return nil, fmt.Errorf("no working transcode pipeline found for %s with %s", inputFile, ffmpegPath)
	}
	t.candidates = validated
	log.Infof("Preview pipeline for %s: %s (ffmpeg: %s)", inputFile, t.candidates[0].name, ffmpegPath)
	return t, nil
}

// buildCandidates returns all candidate pipelines in priority order, gated by
// the advertised capabilities of the selected ffmpeg binary. The software
// pipeline is always the final candidate.
func (t *previewTranscoder) buildCandidates(caps *ffmpegCapabilities) []previewCandidate {
	var out []previewCandidate

	// Most hardware encoders require even dimensions; use software for odd ones.
	if t.resolution%2 == 0 {
		for _, b := range platformBackends() {
			out = append(out, t.backendCandidates(b, caps)...)
		}
	}

	out = append(out, previewCandidate{name: "software", build: t.softwareArgs})
	return out
}

func (t *previewTranscoder) backendCandidates(b backendDef, caps *ffmpegCapabilities) []previewCandidate {
	if b.optInEnv != "" && !envEnabled(b.optInEnv) {
		return nil
	}
	if !caps.hasEncoder(b.encoder) {
		return nil
	}
	decode := caps.hasHwaccel(b.hwaccel) && b.supportsCodec(t.codec)

	outputOpts := append([]string{"-c:v", b.encoder}, b.outputOpts...)

	var out []previewCandidate
	if decode && caps.hasFilter(b.filter) && (!b.fullFlatOnly || t.flat) {
		inputOpts := append(b.fullInputArgs(t), b.inputOpts...)
		out = append(out, previewCandidate{
			name: b.name + "-full",
			build: func(start time.Duration, dest string) []string {
				return t.snippetArgs(inputOpts, b.fullFilter(t), outputOpts, start, dest)
			},
		})
	}
	if decode {
		inputOpts := append([]string{"-hwaccel", b.hwaccel}, b.inputOpts...)
		out = append(out, previewCandidate{
			name: b.name + "-partial",
			build: func(start time.Duration, dest string) []string {
				return t.snippetArgs(inputOpts, t.cpuFilter(b.partialPixFmt), outputOpts, start, dest)
			},
		})
	}
	out = append(out, previewCandidate{
		name: b.name + "-encode",
		build: func(start time.Duration, dest string) []string {
			return t.snippetArgs(b.inputOpts, t.cpuFilter(b.partialPixFmt), outputOpts, start, dest)
		},
	})
	return out
}

// envEnabled reports whether an opt-in environment variable is set to a
// truthy value ("1", "true", ...).
func envEnabled(name string) bool {
	v, err := strconv.ParseBool(os.Getenv(name))
	return err == nil && v
}

func (t *previewTranscoder) cpuFilter(pixFmt string) string {
	return fmt.Sprintf("crop=%s,scale=%d:%d,format=%s", t.crop, t.resolution, t.resolution, pixFmt)
}

func (t *previewTranscoder) softwareArgs(start time.Duration, dest string) []string {
	return t.snippetArgs(nil, fmt.Sprintf("crop=%s,scale=%d:%d", t.crop, t.resolution, t.resolution), []string{"-pix_fmt", "yuv420p"}, start, dest)
}

// snippetArgs assembles a full ffmpeg command line. inputOpts must be input
// options (before -i); outputOpts are output options.
func (t *previewTranscoder) snippetArgs(inputOpts []string, filter string, outputOpts []string, start time.Duration, dest string) []string {
	args := []string{"-y"}
	args = append(args, inputOpts...)
	args = append(args, "-ss", seekString(start), "-i", t.inputFile)
	if filter != "" {
		args = append(args, "-vf", filter)
	}
	args = append(args, outputOpts...)
	args = append(args, "-t", fmt.Sprintf("%v", t.snippetLength), "-an", dest)
	return args
}

func seekString(start time.Duration) string {
	return strings.TrimSuffix(timecode.New(start, timecode.IdentityRate).String(), ":00")
}

type probeKey struct {
	path, candidate, codec, pixFmt string
	width, height                  int
}

// probeCache remembers successful pipeline probes across scenes. Failures are
// deliberately not cached so a candidate that failed for one input may still
// be attempted for another.
var probeCache sync.Map

// probe validates a candidate by running it briefly against the real input
// with the null muxer.
func (t *previewTranscoder) probe(c previewCandidate) bool {
	if c.name == "software" {
		return true
	}

	key := probeKey{t.ffmpegPath, c.name, t.codec, t.pixFmt, t.width, t.height}
	if _, ok := probeCache.Load(key); ok {
		return true
	}

	args := c.build(time.Duration(t.startTime)*time.Second, "-")
	// Replace the output destination with the null muxer.
	args = append(args[:len(args)-1], "-f", "null", "-")
	if err := runFFmpeg(t.ffmpegPath, args...); err != nil {
		log.Debugf("Preview pipeline %s probe failed: %v", c.name, err)
		return false
	}

	probeCache.Store(key, true)
	return true
}

// renderSnippet renders one snippet, retrying with the next validated
// candidate on failure. The successful candidate is used for the remaining
// snippets of this scene to keep stream parameters concat-compatible.
func (t *previewTranscoder) renderSnippet(start time.Duration, dest string) error {
	var lastErr error
	for t.current < len(t.candidates) {
		c := t.candidates[t.current]
		if err := runFFmpeg(t.ffmpegPath, c.build(start, dest)...); err != nil {
			_ = os.Remove(dest)
			log.Warnf("Preview snippet via %s failed, trying next pipeline: %v", c.name, err)
			lastErr = err
			t.current++
			continue
		}
		return nil
	}
	return fmt.Errorf("all preview pipelines failed for %s: %w", t.inputFile, lastErr)
}
