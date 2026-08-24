package tasks

import (
	"os"
	"os/exec"
	"path/filepath"
	"testing"
	"time"

	"github.com/xbapps/xbvr/pkg/common"
	"github.com/xbapps/xbvr/pkg/ffprobe"
)

// TestRenderPreviewIntegration renders a real preview with the ffmpeg/ffprobe
// found on PATH. It skips when the tools are not installed.
func TestRenderPreviewIntegration(t *testing.T) {
	if _, err := exec.LookPath("ffmpeg"); err != nil {
		t.Skip("ffmpeg not available")
	}
	if _, err := exec.LookPath("ffprobe"); err != nil {
		t.Skip("ffprobe not available")
	}

	common.VideoPreviewDir = t.TempDir()

	tests := []struct {
		name       string
		size       string
		projection string
	}{
		{"lr", "3840x1920", "lr"},
		{"flat", "1920x960", "flat"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			input := filepath.Join(t.TempDir(), "input.mp4")
			gen := exec.Command("ffmpeg", "-y", "-f", "lavfi", "-i",
				"testsrc=size="+tt.size+":duration=20:rate=30",
				"-c:v", "libx264", "-pix_fmt", "yuv420p", input)
			if out, err := gen.CombinedOutput(); err != nil {
				t.Fatalf("failed to generate input: %v\n%s", err, out)
			}

			dest := filepath.Join(common.VideoPreviewDir, tt.name+".mp4")
			if err := RenderPreview(input, dest, tt.projection, 2, 0.4, 4, 400, true); err != nil {
				t.Fatalf("RenderPreview failed: %v", err)
			}

			info, err := os.Stat(dest)
			if err != nil || info.Size() == 0 {
				t.Fatalf("preview missing or empty: %v", err)
			}

			data, err := ffprobe.GetProbeData(dest, 10*time.Second)
			if err != nil {
				t.Fatalf("failed to probe preview: %v", err)
			}
			vs := data.GetFirstVideoStream()
			if vs == nil {
				t.Fatal("preview has no video stream")
			}
			if vs.Width != 400 || vs.Height != 400 {
				t.Errorf("preview resolution = %dx%d, want 400x400", vs.Width, vs.Height)
			}
			if vs.CodecName != "h264" {
				t.Errorf("preview codec = %q, want h264", vs.CodecName)
			}
		})
	}
}
