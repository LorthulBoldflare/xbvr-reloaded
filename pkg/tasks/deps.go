package tasks

import (
	"io"
	"net/http"
	"os"
	"os/exec"
	"path/filepath"
	"runtime"
	"sync"

	"github.com/go-resty/resty/v2"
	"github.com/mholt/archiver"
	"github.com/pkg/errors"
	"github.com/tidwall/gjson"
	"github.com/xbapps/xbvr/pkg/common"
	"github.com/xbapps/xbvr/pkg/ffprobe"
)

var depsOnce sync.Once

// CheckDependencies ensures ffmpeg and ffprobe are available. Tools installed
// on the system PATH take precedence over the application-local binaries; only
// tools missing from both locations are downloaded. It runs at most once per
// process and concurrent callers block until the first run completes.
func CheckDependencies() {
	depsOnce.Do(checkDependencies)
}

func checkDependencies() {
	for _, tool := range []string{"ffprobe", "ffmpeg"} {
		if path, found := findSystemBinary(tool); found {
			log.Infof("Using %s from %s", tool, path)
			continue
		}
		if _, err := os.Stat(getLocalBinPath(tool)); err == nil {
			continue
		}
		log.Infof("%s not installed, downloading now...", tool)
		if err := downloadFfbinaries(tool); err != nil {
			log.Warnf("Failed to download %s: %v", tool, err)
		}
	}

	// Set path for go-ffprobe
	ffprobe.SetFFProbeBinPath(GetBinPath("ffprobe"))
}

// resolveBinary ensures dependencies are set up and returns the path to the
// requested tool, preferring PATH over the application-local binary.
func resolveBinary(tool string) (string, error) {
	CheckDependencies()
	path := GetBinPath(tool)
	if _, err := os.Stat(path); err != nil {
		return "", errors.Wrapf(err, "%s not found in PATH or %s", tool, common.BinDir)
	}
	return path, nil
}

// GetBinPath resolves a tool from PATH first, then well-known system bin
// directories, falling back to the application-local binary in common.BinDir.
func GetBinPath(tool string) string {
	if path, found := findSystemBinary(tool); found {
		return path
	}
	return getLocalBinPath(tool)
}

// findSystemBinary looks for a tool on PATH and in well-known bin
// directories. Symlinks are accepted: os.Stat follows them, and the symlink
// path itself is returned so the indirection is preserved.
func findSystemBinary(tool string) (string, bool) {
	if path, err := exec.LookPath(tool); err == nil {
		return path, true
	}
	name := tool
	if runtime.GOOS == "windows" {
		name += ".exe"
	}
	for _, dir := range binSearchDirs() {
		path := filepath.Join(dir, name)
		if info, err := os.Stat(path); err == nil && !info.IsDir() && info.Mode()&0o111 != 0 {
			return path, true
		}
	}
	return "", false
}

// binSearchDirs lists well-known bin directories searched in addition to
// PATH, for platforms where those paths are valid. It is a variable so tests
// can substitute temporary directories.
var binSearchDirs = func() []string {
	var dirs []string
	switch runtime.GOOS {
	case "darwin":
		dirs = append(dirs, "/opt/homebrew/bin", "/usr/local/bin", "/usr/bin")
	case "linux":
		dirs = append(dirs, "/usr/local/bin", "/usr/bin", "/snap/bin")
	}
	if runtime.GOOS != "windows" {
		if home, err := os.UserHomeDir(); err == nil {
			dirs = append(dirs, filepath.Join(home, ".local", "bin"))
		}
	}
	return dirs
}

func getLocalBinPath(tool string) string {
	path := filepath.Join(common.BinDir, tool)
	if runtime.GOOS == "windows" {
		path = path + ".exe"
	}
	return path
}

func downloadFfbinaries(tool string) error {
	var platformId = ""
	if runtime.GOOS == "windows" {
		switch runtime.GOARCH {
		case "386":
			platformId = "windows-32"
		default:
			platformId = "windows-64"
		}
	}
	if runtime.GOOS == "darwin" {
		platformId = "osx-64"
	}
	if runtime.GOOS == "linux" {
		switch runtime.GOARCH {
		case "386":
			platformId = "linux-32"
		case "amd64":
			platformId = "linux-64"
		case "arm":
			platformId = "linux-armhf"
		case "arm64":
			platformId = "linux-arm64"
		}
	}

	if platformId == "" {
		return errors.Errorf("Unknown architecture: %v/%v", runtime.GOOS, runtime.GOARCH)
	}

	resp, err := resty.New().R().Get("https://ffbinaries.com/api/v1/version/4.2.1")
	if err != nil {
		return err
	}

	if resp.StatusCode() != 200 {
		return errors.Errorf("HTTP status code %d", resp.StatusCode())
	}

	url := gjson.Get(resp.String(), "bin."+platformId+"."+tool)

	err = downloadFile(url.String(), filepath.Join(common.BinDir, tool+".zip"))
	if err != nil {
		return err
	}

	err = archiver.Unarchive(filepath.Join(common.BinDir, tool+".zip"), common.BinDir)
	if err != nil {
		return err
	}

	_ = os.Remove(filepath.Join(common.BinDir, tool+".zip"))

	return nil
}

func downloadFile(url, destPath string) error {
	resp, err := http.Get(url)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	if resp.StatusCode != 200 {
		return errors.Errorf("HTTP status code %d", resp.StatusCode)
	}

	out, err := os.Create(destPath)
	if err != nil {
		return err
	}
	defer out.Close()

	_, err = io.Copy(out, resp.Body)
	if err != nil {
		return err
	}

	return nil
}
