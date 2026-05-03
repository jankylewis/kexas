package launcher

import (
	"archive/zip"
	"fmt"
	"io"
	"net/http"
	"os"
	"path/filepath"
	"strings"
)

// downloadFile downloads a file from a URL to a local path.
func downloadFile(filepath string, url string) error {
	var resp *http.Response
	var err error
	resp, err = http.Get(url)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("bad status: %s", resp.Status)
	}

	var out *os.File
	out, err = os.Create(filepath)
	if err != nil {
		return err
	}
	defer out.Close()

	_, err = io.Copy(out, resp.Body)
	return err
}

// unzip extracts a zip file to a destination directory.
// Guards against ZipSlip by rejecting entries whose target path escapes dest.
func unzip(src string, dest string) error {
	var r *zip.ReadCloser
	var err error
	r, err = zip.OpenReader(src)
	if err != nil {
		return err
	}
	defer r.Close()

	err = os.MkdirAll(dest, 0755)
	if err != nil {
		return err
	}

	for _, f := range r.File {
		err = unzipEntry(f, dest)
		if err != nil {
			return err
		}
	}
	return nil
}

// unzipEntry writes a single zip entry into dest, with a ZipSlip guard.
func unzipEntry(f *zip.File, dest string) error {
	var fpath string = filepath.Join(dest, f.Name)

	if !strings.HasPrefix(fpath, filepath.Clean(dest)+string(os.PathSeparator)) {
		return fmt.Errorf("illegal file path: %s", fpath)
	}

	if f.FileInfo().IsDir() {
		os.MkdirAll(fpath, 0755)
		return nil
	}

	var err error = os.MkdirAll(filepath.Dir(fpath), 0755)
	if err != nil {
		return err
	}

	var outFile *os.File
	outFile, err = os.OpenFile(fpath, os.O_WRONLY|os.O_CREATE|os.O_TRUNC, f.Mode())
	if err != nil {
		return err
	}

	var rc io.ReadCloser
	rc, err = f.Open()
	if err != nil {
		outFile.Close()
		return err
	}

	_, err = io.Copy(outFile, rc)
	outFile.Close()
	rc.Close()
	return err
}
