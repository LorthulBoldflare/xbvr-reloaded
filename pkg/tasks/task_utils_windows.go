//go:build windows
// +build windows

package tasks

import (
	"context"
	"os/exec"
	"syscall"
)

// belowNormalPriorityClass keeps ffmpeg preview renders from affecting host
// responsiveness, the Windows equivalent of `nice -n 20`.
const belowNormalPriorityClass = 0x00004000

func buildCmd(name string, arg ...string) *exec.Cmd {
	cmd := exec.Command(name, arg...)
	cmd.SysProcAttr = &syscall.SysProcAttr{HideWindow: true, CreationFlags: belowNormalPriorityClass}
	return cmd
}

func buildCmdContext(ctx context.Context, name string, arg ...string) *exec.Cmd {
	cmd := exec.CommandContext(ctx, name, arg...)
	cmd.SysProcAttr = &syscall.SysProcAttr{HideWindow: true, CreationFlags: belowNormalPriorityClass}
	return cmd
}

// lowerProcessPriority is a no-op on Windows: the priority class is set at
// process creation via CreationFlags.
func lowerProcessPriority(cmd *exec.Cmd) {}
