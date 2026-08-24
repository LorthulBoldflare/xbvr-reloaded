package tasks

import (
	"os"
	"path/filepath"
	"runtime"
	"testing"

	"github.com/xbapps/xbvr/pkg/common"
)

func TestGetBinPathPrefersPath(t *testing.T) {
	if runtime.GOOS == "windows" {
		t.Skip("test uses unix executable semantics")
	}

	dir := t.TempDir()
	fakeTool := filepath.Join(dir, "xbvr-test-tool")
	if err := os.WriteFile(fakeTool, []byte("#!/bin/sh\n"), 0755); err != nil {
		t.Fatal(err)
	}
	t.Setenv("PATH", dir)

	common.BinDir = filepath.Join(dir, "appbin")

	if got := GetBinPath("xbvr-test-tool"); got != fakeTool {
		t.Errorf("GetBinPath should prefer PATH, got %q want %q", got, fakeTool)
	}
}

func TestGetBinPathSearchesWellKnownDirs(t *testing.T) {
	if runtime.GOOS == "windows" {
		t.Skip("test uses unix executable semantics")
	}

	dir := t.TempDir()
	tool := filepath.Join(dir, "xbvr-wellknown-tool")
	if err := os.WriteFile(tool, []byte("#!/bin/sh\n"), 0755); err != nil {
		t.Fatal(err)
	}

	defer func(orig func() []string) { binSearchDirs = orig }(binSearchDirs)
	binSearchDirs = func() []string { return []string{dir} }
	t.Setenv("PATH", t.TempDir()) // not on PATH
	common.BinDir = filepath.Join(dir, "appbin")

	if got := GetBinPath("xbvr-wellknown-tool"); got != tool {
		t.Errorf("GetBinPath should find well-known dir, got %q want %q", got, tool)
	}
}

func TestGetBinPathAcceptsSymlink(t *testing.T) {
	if runtime.GOOS == "windows" {
		t.Skip("test uses unix executable semantics")
	}

	realDir := t.TempDir()
	linkDir := t.TempDir()
	realTool := filepath.Join(realDir, "xbvr-link-tool")
	if err := os.WriteFile(realTool, []byte("#!/bin/sh\n"), 0755); err != nil {
		t.Fatal(err)
	}
	linkTool := filepath.Join(linkDir, "xbvr-link-tool")
	if err := os.Symlink(realTool, linkTool); err != nil {
		t.Fatal(err)
	}

	defer func(orig func() []string) { binSearchDirs = orig }(binSearchDirs)
	binSearchDirs = func() []string { return []string{linkDir} }
	t.Setenv("PATH", t.TempDir())
	common.BinDir = filepath.Join(linkDir, "appbin")

	// The symlink path itself is returned, preserving the indirection.
	if got := GetBinPath("xbvr-link-tool"); got != linkTool {
		t.Errorf("GetBinPath should accept symlink, got %q want %q", got, linkTool)
	}
}

func TestGetBinPathRejectsDanglingSymlinkAndNonExecutable(t *testing.T) {
	if runtime.GOOS == "windows" {
		t.Skip("test uses unix executable semantics")
	}

	dir := t.TempDir()
	if err := os.Symlink(filepath.Join(dir, "missing"), filepath.Join(dir, "xbvr-dangling-tool")); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(dir, "xbvr-noexec-tool"), []byte("#!/bin/sh\n"), 0644); err != nil {
		t.Fatal(err)
	}

	defer func(orig func() []string) { binSearchDirs = orig }(binSearchDirs)
	binSearchDirs = func() []string { return []string{dir} }
	t.Setenv("PATH", t.TempDir())
	common.BinDir = filepath.Join(dir, "appbin")

	for _, tool := range []string{"xbvr-dangling-tool", "xbvr-noexec-tool"} {
		want := filepath.Join(common.BinDir, tool)
		if got := GetBinPath(tool); got != want {
			t.Errorf("GetBinPath(%q) should fall back to BinDir, got %q want %q", tool, got, want)
		}
	}
}

func TestGetBinPathFallsBackToLocalBin(t *testing.T) {
	dir := t.TempDir()
	common.BinDir = dir

	tool := "xbvr-missing-tool"
	want := filepath.Join(dir, tool)
	if runtime.GOOS == "windows" {
		want += ".exe"
	}

	if got := GetBinPath(tool); got != want {
		t.Errorf("GetBinPath should fall back to BinDir, got %q want %q", got, want)
	}
}
