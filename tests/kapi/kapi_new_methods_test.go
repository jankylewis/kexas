package kapi_test

import (
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/jankylewis/kexas/kapi"
)

// 12 tests for new kapi methods: Trace, BodyMultipart, SaveToFile.

// --- TRACE ---

func TestClient_Trace_SendsTraceMethod(t *testing.T) {
	var seen string
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		seen = r.Method
		w.WriteHeader(200)
	}))
	defer srv.Close()

	resp, err := kapi.NewClient(srv.URL).Trace("/")
	if err != nil {
		t.Fatal(err)
	}
	if seen != "TRACE" {
		t.Errorf("expected TRACE, server saw %q", seen)
	}
	if !resp.IsOK() {
		t.Errorf("expected IsOK, got %d", resp.StatusCode)
	}
}

func TestClient_Trace_RespectsBaseURL(t *testing.T) {
	var seenPath string
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		seenPath = r.URL.Path
	}))
	defer srv.Close()

	_, err := kapi.NewClient(srv.URL).Trace("/echo-trace")
	if err != nil {
		t.Fatal(err)
	}
	if seenPath != "/echo-trace" {
		t.Errorf("expected /echo-trace, got %q", seenPath)
	}
}

func TestClient_Trace_ReachableViaGenericRequest(t *testing.T) {
	// Verifying the convenience method matches the generic Request("TRACE",...) form.
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(200)
		_, _ = w.Write([]byte(r.Method))
	}))
	defer srv.Close()

	resp, err := kapi.NewClient(srv.URL).Request("TRACE", "/").Send()
	if err != nil {
		t.Fatal(err)
	}
	if !strings.Contains(resp.Body, "TRACE") {
		t.Errorf("body should echo TRACE; got %q", resp.Body)
	}
}

// --- Multipart ---

func TestRequestBuilder_BodyMultipart_FieldsAndFiles(t *testing.T) {
	var contentType string
	var body []byte
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		contentType = r.Header.Get("Content-Type")
		var buf [4096]byte
		var n int
		n, _ = r.Body.Read(buf[:])
		body = buf[:n]
	}))
	defer srv.Close()

	_, err := kapi.NewClient(srv.URL).Request("POST", "/upload").
		BodyMultipart(
			map[string]string{"caption": "kexas-test"},
			[]kapi.MultipartFile{{
				FieldName: "file",
				Filename:  "hello.txt",
				Content:   []byte("kexas multipart marker"),
			}},
		).
		Send()
	if err != nil {
		t.Fatal(err)
	}

	if !strings.HasPrefix(contentType, "multipart/form-data; boundary=") {
		t.Errorf("Content-Type should be multipart/form-data; got %q", contentType)
	}
	if !strings.Contains(string(body), "kexas-test") {
		t.Errorf("field 'caption' missing from body; got %q", string(body))
	}
	if !strings.Contains(string(body), "hello.txt") {
		t.Errorf("filename 'hello.txt' missing; got %q", string(body))
	}
	if !strings.Contains(string(body), "kexas multipart marker") {
		t.Errorf("file content missing; got %q", string(body))
	}
}

func TestRequestBuilder_BodyMultipart_DefaultContentType(t *testing.T) {
	var body []byte
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		var buf [4096]byte
		var n int
		n, _ = r.Body.Read(buf[:])
		body = buf[:n]
	}))
	defer srv.Close()

	_, err := kapi.NewClient(srv.URL).Request("POST", "/upload").
		BodyMultipart(nil, []kapi.MultipartFile{{
			FieldName: "f",
			Filename:  "a.bin",
			Content:   []byte{0xDE, 0xAD},
			// ContentType deliberately omitted — should default.
		}}).
		Send()
	if err != nil {
		t.Fatal(err)
	}
	// Default content-type for files w/o explicit ContentType is application/octet-stream.
	if !strings.Contains(string(body), "application/octet-stream") {
		t.Errorf("expected default octet-stream content-type; got %q", string(body))
	}
}

func TestRequestBuilder_BodyMultipart_MultipleFiles(t *testing.T) {
	var body []byte
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		var buf [8192]byte
		var n int
		n, _ = r.Body.Read(buf[:])
		body = buf[:n]
	}))
	defer srv.Close()

	_, err := kapi.NewClient(srv.URL).Request("POST", "/upload").
		BodyMultipart(nil, []kapi.MultipartFile{
			{FieldName: "first", Filename: "a.txt", Content: []byte("alpha")},
			{FieldName: "second", Filename: "b.txt", Content: []byte("beta")},
		}).
		Send()
	if err != nil {
		t.Fatal(err)
	}
	for _, marker := range []string{"alpha", "beta", "a.txt", "b.txt"} {
		if !strings.Contains(string(body), marker) {
			t.Errorf("missing marker %q in body", marker)
		}
	}
}

func TestRequestBuilder_BodyMultipart_NoFilesNoFields(t *testing.T) {
	// Edge: empty multipart should still serialise to something valid (just the
	// boundary), not crash.
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(200)
	}))
	defer srv.Close()

	resp, err := kapi.NewClient(srv.URL).Request("POST", "/").
		BodyMultipart(nil, nil).Send()
	if err != nil {
		t.Fatalf("empty multipart should not error: %v", err)
	}
	if !resp.IsOK() {
		t.Errorf("server should accept empty multipart; got %d", resp.StatusCode)
	}
}

// --- SaveToFile ---

func TestResponse_SaveToFile_WritesBytesAtPath(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		_, _ = w.Write([]byte("kexas-download-marker"))
	}))
	defer srv.Close()

	resp, err := kapi.NewClient(srv.URL).Get("/")
	if err != nil {
		t.Fatal(err)
	}

	var dir string = t.TempDir()
	var path string = filepath.Join(dir, "downloaded.txt")
	err = resp.SaveToFile(path)
	if err != nil {
		t.Fatal(err)
	}

	var contents []byte
	contents, err = os.ReadFile(path)
	if err != nil {
		t.Fatal(err)
	}
	if string(contents) != "kexas-download-marker" {
		t.Errorf("expected marker in file; got %q", string(contents))
	}
}

func TestResponse_SaveToFile_CreatesParentDirs(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		_, _ = w.Write([]byte("data"))
	}))
	defer srv.Close()

	resp, err := kapi.NewClient(srv.URL).Get("/")
	if err != nil {
		t.Fatal(err)
	}

	var dir string = t.TempDir()
	var nested string = filepath.Join(dir, "nested", "deep", "file.bin")
	err = resp.SaveToFile(nested)
	if err != nil {
		t.Fatalf("SaveToFile should mkdir -p; got %v", err)
	}
	if _, err := os.Stat(nested); err != nil {
		t.Errorf("file not created: %v", err)
	}
}

func TestResponse_SaveToFile_BinaryContentRoundtrips(t *testing.T) {
	var binary []byte = []byte{0x00, 0xFF, 0x89, 0x50, 0x4E, 0x47}
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/octet-stream")
		_, _ = w.Write(binary)
	}))
	defer srv.Close()

	resp, err := kapi.NewClient(srv.URL).Get("/")
	if err != nil {
		t.Fatal(err)
	}

	var dir string = t.TempDir()
	var path string = filepath.Join(dir, "binary.bin")
	if err := resp.SaveToFile(path); err != nil {
		t.Fatal(err)
	}

	var got []byte
	got, err = os.ReadFile(path)
	if err != nil {
		t.Fatal(err)
	}
	if string(got) != string(binary) {
		t.Errorf("binary content not preserved")
	}
}

func TestResponse_SaveToFile_OverwritesExisting(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		_, _ = w.Write([]byte("new content"))
	}))
	defer srv.Close()

	var dir string = t.TempDir()
	var path string = filepath.Join(dir, "f.txt")
	_ = os.WriteFile(path, []byte("old content"), 0644)

	resp, err := kapi.NewClient(srv.URL).Get("/")
	if err != nil {
		t.Fatal(err)
	}
	if err := resp.SaveToFile(path); err != nil {
		t.Fatal(err)
	}

	var got []byte
	got, _ = os.ReadFile(path)
	if string(got) != "new content" {
		t.Errorf("expected overwrite; got %q", string(got))
	}
}

func TestResponse_SaveToFile_BadPath_ReturnsError(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		_, _ = w.Write([]byte("x"))
	}))
	defer srv.Close()

	resp, _ := kapi.NewClient(srv.URL).Get("/")
	// /dev/null/sub is not a valid path under any sane filesystem.
	err := resp.SaveToFile("/dev/null/cannot-create")
	if err == nil {
		t.Error("expected error for invalid path")
	}
}
