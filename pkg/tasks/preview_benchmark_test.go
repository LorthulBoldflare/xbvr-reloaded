//go:build benchmark

package tasks

import (
	"fmt"
	"os"
	"path/filepath"
	"testing"
	"time"

	"github.com/xbapps/xbvr/pkg/common"
	"github.com/xbapps/xbvr/pkg/ffprobe"
)

// BenchmarkPreviewPipelines faces off pure CPU-based preview generation
// against GPU-based preview generation for the same input file.
//
// It is only compiled with the `benchmark` build tag and reads the input from
// an environment variable, so no real filename is stored in the repository:
//
//	XBVR_BENCH_INPUT=/path/to/video.mp4 go test -tags benchmark -vet=off \
//	    -run '^$' -bench BenchmarkPreviewPipelines -benchtime 1x ./pkg/tasks/
//
// XBVR_BENCH_PROJECTION optionally sets the video projection (default "LR").
func BenchmarkPreviewPipelines(b *testing.B) {
	input := os.Getenv("XBVR_BENCH_INPUT")
	if input == "" {
		b.Skip("XBVR_BENCH_INPUT not set")
	}
	projection := os.Getenv("XBVR_BENCH_PROJECTION")
	if projection == "" {
		projection = "LR"
	}

	const (
		startTime     = 10
		snippetLength = 0.4
		snippetAmount = 20
		resolution    = 400
	)

	ffmpegPath, err := resolveBinary("ffmpeg")
	if err != nil {
		b.Fatal(err)
	}

	ffdata, err := ffprobe.GetProbeData(input, 10*time.Second)
	if err != nil {
		b.Fatal(err)
	}
	vs := ffdata.GetFirstVideoStream()
	if vs == nil {
		b.Fatal("no video stream found")
	}
	dur := ffdata.Format.DurationSeconds

	crop := "iw/2:ih:iw/2:ih"
	rect := cropRect{w: vs.Width / 2, h: vs.Height, x: vs.Width / 2, y: 0}
	if vs.Height == vs.Width {
		crop = "iw/2:ih/2:iw/4:ih/2"
		rect = cropRect{w: vs.Width / 2, h: vs.Height / 2, x: vs.Width / 4, y: vs.Height / 2}
	}
	flat := projection == "flat"
	if flat {
		crop = "iw:ih:iw:ih"
		rect = cropRect{w: vs.Width, h: vs.Height, x: 0, y: 0}
	}

	newTranscoder := func() *previewTranscoder {
		tr, err := newPreviewTranscoder(
			ffmpegPath, input, vs.CodecName, vs.PixFmt, vs.Width, vs.Height, flat, crop, rect, resolution, snippetLength, startTime,
		)
		if err != nil {
			b.Fatal(err)
		}
		return tr
	}

	// CPU: software candidate only.
	cpuTranscoder := newTranscoder()
	cpuTranscoder.candidates = []previewCandidate{{name: "software", build: cpuTranscoder.softwareArgs}}

	// GPU: first validated hardware candidate only (no software fallback, so
	// the measurement is not contaminated by a silent fallback).
	gpuTranscoder := newTranscoder()
	gpu := gpuTranscoder.candidates[0]
	if gpu.name == "software" {
		b.Skip("no hardware pipeline validated for this input")
	}
	gpuTranscoder.candidates = []previewCandidate{gpu}

	b.Logf("input: %dx%d %s %s, CPU pipeline: software, GPU pipeline: %s", vs.Width, vs.Height, vs.CodecName, projection, gpu.name)

	for _, bench := range []struct {
		name string
		tr   *previewTranscoder
	}{
		{"cpu", cpuTranscoder},
		{"gpu", gpuTranscoder},
	} {
		b.Run(bench.name, func(b *testing.B) {
			for i := 0; i < b.N; i++ {
				tmp := b.TempDir()
				common.VideoPreviewDir = tmp
				dest := filepath.Join(tmp, bench.name+".mp4")
				renderBenchmarkPreview(b, ffmpegPath, bench.tr, dur, startTime, snippetLength, snippetAmount, dest)
				info, err := os.Stat(dest)
				if err != nil || info.Size() == 0 {
					b.Fatalf("preview missing or empty: %v", err)
				}
			}
		})
	}
}

// renderBenchmarkPreview mirrors the RenderPreview snippet loop and concat so
// the benchmark measures the complete preview generation, not just encoding.
func renderBenchmarkPreview(b *testing.B, ffmpegPath string, tr *previewTranscoder, dur float64, startTime int, snippetLength float64, snippetAmount int, dest string) {
	b.Helper()

	tmpPath := filepath.Join(common.VideoPreviewDir, "tmp")
	if err := os.MkdirAll(tmpPath, os.ModePerm); err != nil {
		b.Fatal(err)
	}

	interval := (dur - float64(startTime)) / float64(snippetAmount)
	for i := 1; i <= snippetAmount; i++ {
		start := time.Duration(float64(i)*interval+float64(startTime)) * time.Second
		if err := tr.renderSnippet(start, filepath.Join(tmpPath, fmt.Sprintf("%v.mp4", i))); err != nil {
			b.Fatalf("snippet %d via %s failed: %v", i, tr.candidates[tr.current].name, err)
		}
	}

	concatFile := filepath.Join(tmpPath, "concat.txt")
	f, err := os.Create(concatFile)
	if err != nil {
		b.Fatal(err)
	}
	for i := 1; i <= snippetAmount; i++ {
		f.WriteString(fmt.Sprintf("file '%v.mp4'\n", i))
	}
	f.Close()

	if err := runFFmpeg(ffmpegPath,
		"-y", "-f", "concat", "-safe", "0",
		"-i", filepath.ToSlash(concatFile),
		"-c", "copy", filepath.ToSlash(dest),
	); err != nil {
		b.Fatalf("concat failed: %v", err)
	}
}
