package tasks

import (
	"archive/zip"
	"crypto/sha256"
	"encoding/hex"
	"fmt"
	"io"
	"net/http"
	"os"
	"os/exec"
	"path/filepath"
	"runtime"
	"strings"
	"sync"
	"time"

	"github.com/go-resty/resty/v2"
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

// ffbinariesSHA256 pins the expected SHA-256 of every ffbinaries 4.2.1 zip
// (github.com/ffbinaries/ffbinaries-prebuilt releases), keyed by
// "<platform>/<tool>". Downloaded archives are verified against these hashes
// before extraction; a tampered or corrupted download is rejected.
var ffbinariesSHA256 = map[string]string{
	"windows-32/ffmpeg":   "0756d28826d7f89afff43fb831b4452eac745d12868cf988a09be8052c564215",
	"windows-32/ffprobe":  "0aff9d7a3ae9ab287f29068657eb0321890e386fdac3401521cb5a94cc2f3fd7",
	"windows-64/ffmpeg":   "6894ad0ffe2dba571d670649e7bfca99713e26c4b6765726e1445bda0cf676de",
	"windows-64/ffprobe":  "bd50167b75c2996ab8c488304188bd18f61da083c5f9502eff585acc2c81ada1",
	"osx-64/ffmpeg":       "dc1b0fb67bb24b877fe3d776c884160def4965ef9a41987a7dc90ea5267e8323",
	"osx-64/ffprobe":      "a84e66f24bbfe2ff104f117fae552e6f22bf440d034cc99c83c60ef8fd516ef3",
	"linux-32/ffmpeg":     "3b3097dc9b5b995b3f80b002fed0992eaa022f31502f85bff6be35b1c3eeed95",
	"linux-32/ffprobe":    "8ad12cfacaa7dc3eb672fc0014b7ac16391680b2e99efc56276e3ec593ea28fe",
	"linux-64/ffmpeg":     "b66df6c2a0cc442c8fa870420620c2e39f4f541a8d6fe64f385c56b206579525",
	"linux-64/ffprobe":    "7e09f1bcc043b828e237d8063df2dc0a48e7f96478d5543d9f3574473adb9722",
	"linux-armhf/ffmpeg":  "7a207ea3af0278ddc3f68d00822cfbf9f94eb8e8537d19f6ab3fab5cc095769b",
	"linux-armhf/ffprobe": "7d632e19b60ab769a4e0ad6dfc85f962a78acf56e3434df70533db7de9b0f6bf",
	"linux-arm64/ffmpeg":  "f63df1a4477fdcc25932a232689d44b9702fc5bd1866b8e7472e0d8a8362ce2f",
	"linux-arm64/ffprobe": "22fd2230958c41f6405dc1fc0ede5fb0dcc993f648a977912e72e358a70a788a",
}

// verifySHA256 checks the file at path against an expected hex SHA-256.
func verifySHA256(path, expectedHex string) error {
	f, err := os.Open(path)
	if err != nil {
		return err
	}
	defer f.Close()

	h := sha256.New()
	if _, err := io.Copy(h, f); err != nil {
		return err
	}
	if got := hex.EncodeToString(h.Sum(nil)); got != expectedHex {
		return fmt.Errorf("SHA-256 mismatch for %s: got %s, expected %s (download may be tampered with or corrupt)", filepath.Base(path), got, expectedHex)
	}
	return nil
}

// extractZipSafe extracts a zip archive into destDir, rejecting entries that
// would escape the destination (ZipSlip). Replaces mholt/archiver v3, which
// predates the ZipSlip fix.
func extractZipSafe(zipPath, destDir string) error {
	zr, err := zip.OpenReader(zipPath)
	if err != nil {
		return err
	}
	defer zr.Close()

	base := filepath.Clean(destDir) + string(os.PathSeparator)
	for _, f := range zr.File {
		name := filepath.Clean(f.Name)
		if name == ".." || strings.HasPrefix(name, ".."+string(os.PathSeparator)) || filepath.IsAbs(name) {
			return fmt.Errorf("zip entry %q escapes destination directory", f.Name)
		}
		dest := filepath.Join(destDir, name)
		if !strings.HasPrefix(dest, base) {
			return fmt.Errorf("zip entry %q escapes destination directory", f.Name)
		}

		if f.FileInfo().IsDir() {
			if err := os.MkdirAll(dest, 0o755); err != nil {
				return err
			}
			continue
		}

		mode := f.Mode()
		if mode == 0 {
			mode = 0o755 // archives for executables
		}
		rc, err := f.Open()
		if err != nil {
			return err
		}
		out, err := os.OpenFile(dest, os.O_CREATE|os.O_WRONLY|os.O_TRUNC, mode)
		if err != nil {
			rc.Close()
			return err
		}
		_, copyErr := io.Copy(out, rc)
		closeErr := out.Close()
		rc.Close()
		if copyErr != nil {
			return copyErr
		}
		if closeErr != nil {
			return closeErr
		}
	}
	return nil
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

	expectedHash, ok := ffbinariesSHA256[platformId+"/"+tool]
	if !ok {
		return errors.Errorf("No pinned checksum for %s on %s — refusing to download an unverified binary", tool, platformId)
	}

	resp, err := resty.New().SetTimeout(30 * time.Second).R().Get("https://ffbinaries.com/api/v1/version/4.2.1")
	if err != nil {
		return err
	}

	if resp.StatusCode() != 200 {
		return errors.Errorf("HTTP status code %d", resp.StatusCode())
	}

	url := gjson.Get(resp.String(), "bin."+platformId+"."+tool)

	zipPath := filepath.Join(common.BinDir, tool+".zip")
	err = downloadFile(url.String(), zipPath)
	if err != nil {
		return err
	}
	defer os.Remove(zipPath)

	if err := verifySHA256(zipPath, expectedHash); err != nil {
		return err
	}

	if err := extractZipSafe(zipPath, common.BinDir); err != nil {
		return err
	}

	return nil
}

func downloadFile(url, destPath string) error {
	client := &http.Client{Timeout: 5 * time.Minute}
	resp, err := client.Get(url)
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
