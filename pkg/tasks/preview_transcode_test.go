package tasks

import (
	"errors"
	"strings"
	"testing"
	"time"
)

const sampleFFmpegOutput = `
Hardware acceleration methods:
videotoolbox
`

const sampleEncodersOutput = `
Encoders:
 V....D libx264              libx264 H.264 / AVC / MPEG-4 AVC / MPEG-4 part 10 (codec h264)
 V....D h264_videotoolbox    VideoToolbox H.264 Encoder (codec h264)
 V....D h264_nvenc           NVIDIA NVENC H.264 encoder (codec h264)
`

const sampleDecodersOutput = `
Decoders:
 VFS..D h264                 H.264 / AVC / MPEG-4 AVC / MPEG-4 part 10
 VFS..D hevc                 HEVC (High Efficiency Video Coding)
`

const sampleFiltersOutput = `
Filters:
 .. crop              V->V       Crop the input video.
 .. scale             V->V       Scale the input video size and/or convert the image format.
 .. scale_vt          V->V       Scale Videotoolbox frames
`

func testCaps() *ffmpegCapabilities {
	return &ffmpegCapabilities{
		path:     "/fake/ffmpeg",
		hwaccels: sampleFFmpegOutput,
		decoders: sampleDecodersOutput,
		encoders: sampleEncodersOutput,
		filters:  sampleFiltersOutput,
	}
}

func TestHasToken(t *testing.T) {
	caps := testCaps()
	if !caps.hasHwaccel("videotoolbox") {
		t.Error("expected videotoolbox hwaccel")
	}
	if !caps.hasEncoder("h264_videotoolbox") {
		t.Error("expected h264_videotoolbox encoder")
	}
	if !caps.hasDecoder("h264") {
		t.Error("expected h264 decoder")
	}
	if !caps.hasFilter("scale_vt") {
		t.Error("expected scale_vt filter")
	}
	for _, missing := range []string{"cuda", "h264_nvenc_fake", "vp8", "scale_cuda"} {
		if caps.hasHwaccel(missing) || caps.hasFilter(missing) || caps.hasDecoder(missing) {
			t.Errorf("unexpected token match for %q", missing)
		}
	}
	if caps.hasEncoder("h264_nvenc_fake") {
		t.Error("unexpected encoder match")
	}
}

func newTestTranscoder(flat bool, resolution int) *previewTranscoder {
	return &previewTranscoder{
		ffmpegPath:    "/fake/ffmpeg",
		inputFile:     "/fake/input.mp4",
		codec:         "h264",
		pixFmt:        "yuv420p",
		width:         3840,
		height:        1920,
		flat:          flat,
		crop:          "iw/2:ih:iw/2:ih",
		rect:          cropRect{w: 1920, h: 1920, x: 1920, y: 0},
		resolution:    resolution,
		snippetLength: 0.4,
		startTime:     10,
	}
}

func TestSoftwareCandidateAlwaysLast(t *testing.T) {
	tr := newTestTranscoder(false, 400)
	candidates := tr.buildCandidates(testCaps())
	last := candidates[len(candidates)-1]
	if last.name != "software" {
		t.Errorf("last candidate = %q, want software", last.name)
	}
}

func TestOddResolutionSkipsHardware(t *testing.T) {
	tr := newTestTranscoder(false, 401)
	candidates := tr.buildCandidates(testCaps())
	if len(candidates) != 1 || candidates[0].name != "software" {
		t.Errorf("odd resolution should yield only software, got %v", candidateNames(candidates))
	}
}

func TestEncodeOnlyCandidateWithoutHwaccel(t *testing.T) {
	caps := testCaps()
	caps.hwaccels = "" // encoder present, hwaccel method absent
	tr := newTestTranscoder(false, 400)

	var names []string
	for _, c := range tr.buildCandidates(caps) {
		names = append(names, c.name)
	}
	for _, n := range names {
		if strings.Contains(n, "videotoolbox-full") || strings.Contains(n, "videotoolbox-partial") {
			t.Errorf("full/partial candidates must require the hwaccel method, got %q", n)
		}
	}
}

func candidateNames(candidates []previewCandidate) []string {
	names := make([]string, len(candidates))
	for i, c := range candidates {
		names[i] = c.name
	}
	return names
}

func TestUnsupportedCodecSkipsHardwareDecode(t *testing.T) {
	tr := newTestTranscoder(false, 400)
	tr.codec = "wmv3"
	for _, c := range tr.buildCandidates(testCaps()) {
		if strings.HasSuffix(c.name, "-full") || strings.HasSuffix(c.name, "-partial") {
			t.Errorf("codec wmv3 must not produce decode candidates, got %q", c.name)
		}
	}
}

func TestSoftwareArgsPreserveExistingBehavior(t *testing.T) {
	tr := newTestTranscoder(false, 400)
	args := tr.softwareArgs(70*time.Second, "/tmp/out.mp4")
	joined := strings.Join(args, " ")

	for _, want := range []string{
		"-y",
		"-ss 00:01:10",
		"-i /fake/input.mp4",
		"-vf crop=iw/2:ih:iw/2:ih,scale=400:400",
		"-pix_fmt yuv420p",
		"-t 0.4",
		"-an /tmp/out.mp4",
	} {
		if !strings.Contains(joined, want) {
			t.Errorf("software args missing %q: %s", want, joined)
		}
	}
	if strings.Contains(joined, "-c:v") {
		t.Errorf("software args must not pin an encoder: %s", joined)
	}
}

func TestPartialArgsUseHwaccelAndEncoder(t *testing.T) {
	tr := newTestTranscoder(false, 400)
	args := tr.snippetArgs(
		[]string{"-hwaccel", "videotoolbox"},
		tr.cpuFilter("yuv420p"),
		[]string{"-c:v", "h264_videotoolbox"},
		70*time.Second, "/tmp/out.mp4",
	)
	joined := strings.Join(args, " ")

	// hwaccel must be an input option (before -i)
	if strings.Index(joined, "-hwaccel") > strings.Index(joined, "-i ") {
		t.Errorf("hwaccel option must precede -i: %s", joined)
	}
	for _, want := range []string{"-hwaccel videotoolbox", "-c:v h264_videotoolbox", "format=yuv420p"} {
		if !strings.Contains(joined, want) {
			t.Errorf("partial args missing %q: %s", want, joined)
		}
	}
}

func TestQSVFullFilterIncludesCrop(t *testing.T) {
	tr := newTestTranscoder(false, 400)
	filter := qsvBackend.fullFilter(tr)
	if filter != "vpp_qsv=cw=1920:ch=1920:cx=1920:cy=0:w=400:h=400:format=nv12" {
		t.Errorf("unexpected qsv filter: %s", filter)
	}
}

func TestVideoToolboxOptIn(t *testing.T) {
	tr := newTestTranscoder(false, 400)

	if got := tr.backendCandidates(videoToolboxBackend, testCaps()); got != nil {
		t.Errorf("videotoolbox must be opt-in, got %v without env var", candidateNames(got))
	}

	t.Setenv("XBVR_PREVIEW_VIDEOTOOLBOX", "1")
	if got := tr.backendCandidates(videoToolboxBackend, testCaps()); len(got) == 0 {
		t.Error("videotoolbox candidates expected when XBVR_PREVIEW_VIDEOTOOLBOX=1")
	}
}

func TestVideoToolboxCandidatesLimitThreadsAndRate(t *testing.T) {
	t.Setenv("XBVR_PREVIEW_VIDEOTOOLBOX", "true")
	tr := newTestTranscoder(false, 400)
	candidates := tr.backendCandidates(videoToolboxBackend, testCaps())
	if len(candidates) == 0 {
		t.Fatal("expected videotoolbox candidates")
	}
	for _, c := range candidates {
		joined := strings.Join(c.build(70*time.Second, "/tmp/out.mp4"), " ")
		for _, want := range []string{"-threads 4", "-maxrate 2M", "-bufsize 4M"} {
			if !strings.Contains(joined, want) {
				t.Errorf("candidate %q missing %q: %s", c.name, want, joined)
			}
		}
	}
	// Thread cap must be an input option (before -i).
	partial := strings.Join(candidates[1].build(70*time.Second, "/tmp/out.mp4"), " ")
	if strings.Index(partial, "-threads 4") > strings.Index(partial, "-i ") {
		t.Errorf("thread cap must precede -i: %s", partial)
	}
}

func TestRenderSnippetFallsBackAndSticks(t *testing.T) {
	defer func(orig func(string, ...string) error) { runFFmpeg = orig }(runFFmpeg)

	tr := newTestTranscoder(false, 400)
	ran := ""
	track := func(name string) func(time.Duration, string) []string {
		return func(start time.Duration, dest string) []string {
			ran = name
			return tr.softwareArgs(start, dest)
		}
	}
	runFFmpeg = func(path string, args ...string) error {
		if ran == "broken" {
			return errors.New("boom")
		}
		return nil
	}

	tr.candidates = []previewCandidate{
		{name: "broken", build: track("broken")},
		{name: "working", build: track("working")},
	}

	if err := tr.renderSnippet(70*time.Second, t.TempDir()+"/1.mp4"); err != nil {
		t.Fatalf("renderSnippet should succeed via fallback: %v", err)
	}
	if tr.current != 1 {
		t.Fatalf("current candidate = %d, want 1", tr.current)
	}

	// Subsequent snippets keep using the successful candidate.
	if err := tr.renderSnippet(80*time.Second, t.TempDir()+"/2.mp4"); err != nil {
		t.Fatal(err)
	}
	if tr.current != 1 || ran != "working" {
		t.Errorf("candidate switched after success, current = %d ran = %q", tr.current, ran)
	}
}

func TestRenderSnippetExhaustsCandidates(t *testing.T) {
	defer func(orig func(string, ...string) error) { runFFmpeg = orig }(runFFmpeg)

	runFFmpeg = func(path string, args ...string) error {
		return errors.New("boom")
	}

	tr := newTestTranscoder(false, 400)
	tr.candidates = []previewCandidate{
		{name: "a", build: tr.softwareArgs},
		{name: "b", build: tr.softwareArgs},
	}

	if err := tr.renderSnippet(70*time.Second, t.TempDir()+"/1.mp4"); err == nil {
		t.Fatal("expected error when all candidates fail")
	}
}

func TestProbeUsesNullMuxerAndCachesSuccess(t *testing.T) {
	defer func(orig func(string, ...string) error) { runFFmpeg = orig }(runFFmpeg)

	var gotArgs []string
	runFFmpeg = func(path string, args ...string) error {
		gotArgs = args
		return nil
	}

	tr := newTestTranscoder(false, 400)
	candidate := previewCandidate{name: "probe-test-vt", build: tr.softwareArgs}

	if !tr.probe(candidate) {
		t.Fatal("probe should succeed")
	}
	joined := strings.Join(gotArgs, " ")
	if !strings.HasSuffix(joined, "-f null -") {
		t.Errorf("probe output should be the null muxer: %s", joined)
	}
	if strings.Contains(joined, "-an - -f") {
		t.Errorf("probe must not keep the placeholder destination: %s", joined)
	}

	// A cached success must not invoke ffmpeg again.
	runFFmpeg = func(path string, args ...string) error {
		t.Error("cached probe should not run ffmpeg")
		return nil
	}
	if !tr.probe(candidate) {
		t.Error("cached probe should succeed")
	}
	probeCache.Delete(probeKey{tr.ffmpegPath, candidate.name, tr.codec, tr.pixFmt, tr.width, tr.height})
}

func TestProbeFailureNotCached(t *testing.T) {
	defer func(orig func(string, ...string) error) { runFFmpeg = orig }(runFFmpeg)

	calls := 0
	runFFmpeg = func(path string, args ...string) error {
		calls++
		return errors.New("boom")
	}

	tr := newTestTranscoder(false, 400)
	candidate := previewCandidate{name: "probe-test-fail", build: tr.softwareArgs}

	if tr.probe(candidate) {
		t.Fatal("probe should fail")
	}
	tr.probe(candidate)
	if calls != 2 {
		t.Errorf("failed probe should not be cached, ffmpeg ran %d times", calls)
	}
}
