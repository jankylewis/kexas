package kcore

import (
	"archive/zip"
	"fmt"
	"io"
	"net/http"
	"os"
	"os/exec"
	"path/filepath"
	"runtime"
	"strings"

	"github.com/jankylewis/kexas/internal/logger"
)

// EnsureFFmpeg returns a usable ffmpeg binary path, installing one to
// ~/.kexas/bin/ on first call if neither that location nor PATH has one.
//
// Resolution order:
//  1. ~/.kexas/bin/ffmpeg      (kexas-managed install — checked first so a
//     locally-cached download is reused before scanning PATH)
//  2. exec.LookPath("ffmpeg")  (system install)
//  3. download + extract       (writes to managed dir, cached for next call)
//
// Modelled after Playwright's bundled-Chromium approach: zero user setup, the
// SDK fetches its own binary on first use.
func EnsureFFmpeg(log *logger.Logger) (string, error) {
	var binDir string
	var err error
	binDir, err = ffmpegManagedBinDir()
	if err != nil {
		return "", err
	}

	var managedPath string = filepath.Join(binDir, ffmpegBinaryName())
	_, err = os.Stat(managedPath)
	if err == nil {
		return managedPath, nil
	}

	var systemPath string
	systemPath, err = exec.LookPath("ffmpeg")
	if err == nil {
		return systemPath, nil
	}

	log.Info("ffmpeg not found on PATH, downloading static build", "target", binDir)
	err = installFFmpeg(binDir, log)
	if err != nil {
		return "", fmt.Errorf("ffmpeg auto-install: %w", err)
	}
	return managedPath, nil
}

// ffmpegManagedBinDir returns ~/.kexas/bin/ — the directory kexas uses for
// auto-installed binaries.
func ffmpegManagedBinDir() (string, error) {
	var home string
	var err error
	home, err = os.UserHomeDir()
	if err != nil {
		return "", fmt.Errorf("user home dir: %w", err)
	}
	return filepath.Join(home, ".kexas", "bin"), nil
}

// ffmpegBinaryName returns the platform-specific ffmpeg binary name.
func ffmpegBinaryName() string {
	if runtime.GOOS == "windows" {
		return "ffmpeg.exe"
	}
	return "ffmpeg"
}

// installFFmpeg downloads the static ffmpeg build for the current platform and
// extracts the ffmpeg binary into binDir.
func installFFmpeg(binDir string, log *logger.Logger) error {
	var err error = os.MkdirAll(binDir, 0755)
	if err != nil {
		return fmt.Errorf("create %s: %w", binDir, err)
	}

	var downloadURL string
	downloadURL, err = ffmpegDownloadURL()
	if err != nil {
		return err
	}

	log.Info("downloading ffmpeg", "url", downloadURL)
	var archivePath string
	archivePath, err = downloadFFmpegArchive(downloadURL)
	if err != nil {
		return err
	}
	defer os.Remove(archivePath)

	var targetPath string = filepath.Join(binDir, ffmpegBinaryName())
	err = extractFFmpegBinary(archivePath, targetPath)
	if err != nil {
		return err
	}
	log.Info("ffmpeg installed", "path", targetPath)
	return nil
}

// ffmpegDownloadURL returns the static-build URL for the running platform.
// Returns an error for combinations we don't pre-package (e.g., 32-bit ARM).
//
// Sources:
//   - macOS:   evermeet.cx (universal binary, Intel + Apple Silicon)
//   - Linux:   github.com/BtbN/FFmpeg-Builds (GPL static)
//   - Windows: github.com/BtbN/FFmpeg-Builds (GPL static)
func ffmpegDownloadURL() (string, error) {
	switch runtime.GOOS {
	case "darwin":
		return "https://evermeet.cx/ffmpeg/getrelease/zip", nil
	case "linux":
		switch runtime.GOARCH {
		case "amd64":
			return "https://github.com/BtbN/FFmpeg-Builds/releases/latest/download/ffmpeg-master-latest-linux64-gpl.tar.xz", nil
		case "arm64":
			return "https://github.com/BtbN/FFmpeg-Builds/releases/latest/download/ffmpeg-master-latest-linuxarm64-gpl.tar.xz", nil
		}
	case "windows":
		if runtime.GOARCH == "amd64" {
			return "https://github.com/BtbN/FFmpeg-Builds/releases/latest/download/ffmpeg-master-latest-win64-gpl.zip", nil
		}
	}
	return "", fmt.Errorf("no ffmpeg auto-install build for %s/%s", runtime.GOOS, runtime.GOARCH)
}

// downloadFFmpegArchive downloads the URL to a tempfile preserving the original
// extension (.zip / .tar.xz) and returns the temp path.
func downloadFFmpegArchive(downloadURL string) (string, error) {
	var resp *http.Response
	var err error
	resp, err = http.Get(downloadURL)
	if err != nil {
		return "", fmt.Errorf("GET %s: %w", downloadURL, err)
	}
	defer resp.Body.Close()
	if resp.StatusCode != http.StatusOK {
		return "", fmt.Errorf("GET %s: status %d", downloadURL, resp.StatusCode)
	}

	var ext string = archiveExtension(downloadURL)
	var f *os.File
	f, err = os.CreateTemp("", "kexas-ffmpeg-*"+ext)
	if err != nil {
		return "", err
	}
	defer f.Close()

	_, err = io.Copy(f, resp.Body)
	if err != nil {
		os.Remove(f.Name())
		return "", fmt.Errorf("write archive: %w", err)
	}
	return f.Name(), nil
}

// archiveExtension returns ".tar.xz" or ".zip" based on the URL suffix.
// We need to detect ".tar.xz" specifically because filepath.Ext only sees ".xz".
// evermeet's getrelease URL has no extension; it serves a zip.
func archiveExtension(downloadURL string) string {
	if strings.HasSuffix(downloadURL, ".tar.xz") {
		return ".tar.xz"
	}
	if strings.HasSuffix(downloadURL, ".zip") {
		return ".zip"
	}
	return ".zip"
}

// extractFFmpegBinary copies just the ffmpeg binary out of archivePath into
// targetPath, dispatching on archive type.
func extractFFmpegBinary(archivePath, targetPath string) error {
	if strings.HasSuffix(archivePath, ".zip") {
		return extractFFmpegFromZip(archivePath, targetPath)
	}
	if strings.HasSuffix(archivePath, ".tar.xz") {
		return extractFFmpegFromTarXZ(archivePath, targetPath)
	}
	return fmt.Errorf("unrecognised archive: %s", archivePath)
}

// extractFFmpegFromZip walks the zip, finds the entry whose basename matches
// the ffmpeg binary, and writes it to targetPath with executable perms.
func extractFFmpegFromZip(archivePath, targetPath string) error {
	var r *zip.ReadCloser
	var err error
	r, err = zip.OpenReader(archivePath)
	if err != nil {
		return fmt.Errorf("open zip: %w", err)
	}
	defer r.Close()

	var binaryName string = ffmpegBinaryName()
	for _, entry := range r.File {
		if filepath.Base(entry.Name) != binaryName {
			continue
		}
		return copyZipEntryToFile(entry, targetPath)
	}
	return fmt.Errorf("%s not found in %s", binaryName, archivePath)
}

// copyZipEntryToFile copies a zip entry's contents to dst with 0755 perms.
func copyZipEntryToFile(entry *zip.File, dst string) error {
	var rc io.ReadCloser
	var err error
	rc, err = entry.Open()
	if err != nil {
		return fmt.Errorf("open zip entry: %w", err)
	}
	defer rc.Close()

	var out *os.File
	out, err = os.OpenFile(dst, os.O_WRONLY|os.O_CREATE|os.O_TRUNC, 0755)
	if err != nil {
		return fmt.Errorf("create %s: %w", dst, err)
	}
	defer out.Close()

	_, err = io.Copy(out, rc)
	if err != nil {
		return fmt.Errorf("copy zip entry: %w", err)
	}
	return os.Chmod(dst, 0755)
}

// extractFFmpegFromTarXZ shells out to the system tar to extract archivePath,
// then copies the ffmpeg binary out of the result. Avoids adding a pure-Go xz
// decoder dependency — `tar -xf` auto-detects xz on every modern Linux.
func extractFFmpegFromTarXZ(archivePath, targetPath string) error {
	var tmpDir string
	var err error
	tmpDir, err = os.MkdirTemp("", "kexas-ffmpeg-extract-")
	if err != nil {
		return err
	}
	defer os.RemoveAll(tmpDir)

	var cmd *exec.Cmd = exec.Command("tar", "-xf", archivePath, "-C", tmpDir)
	var output []byte
	output, err = cmd.CombinedOutput()
	if err != nil {
		return fmt.Errorf("tar -xf %s: %w (%s)", archivePath, err, string(output))
	}

	var foundPath string
	foundPath, err = findFFmpegBinaryInDir(tmpDir)
	if err != nil {
		return err
	}
	return copyFileWithMode(foundPath, targetPath, 0755)
}

// findFFmpegBinaryInDir walks dir and returns the first file whose basename is
// the ffmpeg binary.
func findFFmpegBinaryInDir(dir string) (string, error) {
	var binaryName string = ffmpegBinaryName()
	var foundPath string
	var walkErr error = filepath.Walk(dir, func(p string, info os.FileInfo, e error) error {
		if e != nil {
			return e
		}
		if info.IsDir() {
			return nil
		}
		if filepath.Base(p) != binaryName {
			return nil
		}
		foundPath = p
		return filepath.SkipAll
	})
	if walkErr != nil {
		return "", walkErr
	}
	if foundPath == "" {
		return "", fmt.Errorf("%s not found in %s", binaryName, dir)
	}
	return foundPath, nil
}

// copyFileWithMode copies src to dst, setting dst's mode bits.
func copyFileWithMode(src, dst string, mode os.FileMode) error {
	var in *os.File
	var err error
	in, err = os.Open(src)
	if err != nil {
		return err
	}
	defer in.Close()

	var out *os.File
	out, err = os.OpenFile(dst, os.O_WRONLY|os.O_CREATE|os.O_TRUNC, mode)
	if err != nil {
		return err
	}
	defer out.Close()

	_, err = io.Copy(out, in)
	if err != nil {
		return err
	}
	return os.Chmod(dst, mode)
}
