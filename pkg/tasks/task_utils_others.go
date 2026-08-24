//go:build !windows
// +build !windows

package tasks

import (
	"context"
	"os/exec"

	"golang.org/x/sys/unix"
)

// lowPriorityNice matches `nice -n 20` to keep ffmpeg preview renders from
// affecting host responsiveness. Values above the platform maximum are
// clamped by the kernel.
const lowPriorityNice = 20

func buildCmd(name string, arg ...string) *exec.Cmd {
	return exec.Command(name, arg...)
}

func buildCmdContext(ctx context.Context, name string, arg ...string) *exec.Cmd {
	return exec.CommandContext(ctx, name, arg...)
}

// lowerProcessPriority demotes a started process to background priority.
// Errors are ignored: the render must proceed even if demotion fails.
func lowerProcessPriority(cmd *exec.Cmd) {
	if cmd.Process != nil {
		_ = unix.Setpriority(unix.PRIO_PROCESS, cmd.Process.Pid, lowPriorityNice)
	}
}
