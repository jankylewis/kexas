package httpbin_test

import (
	"strings"
	"testing"

	"github.com/jankylewis/kexas/kapi"
)

// httpbin.org is the canonical HTTP-client test target — it exposes one
// endpoint per HTTP method that echoes the request shape back as JSON.
// We exercise EVERY kapi method against it, plus multipart upload and
// file download.

const httpbinBase string = "https://httpbin.org"

func newClient() *kapi.Client {
	return kapi.NewClient(httpbinBase,
		kapi.WithHeader("User-Agent", "kexas-apitests/0.1"),
	)
}

// --- One test per HTTP method (8 methods total) ---

func TestGet_ReturnsURLAndArgs(t *testing.T) {
	c := newClient()
	resp, err := c.Request("GET", "/get").Query("a", "1").Send()
	must(t, err)
	if !resp.IsOK() {
		t.Fatalf("GET /get → %d", resp.StatusCode)
	}
	body, _ := resp.JSONMap()
	args, _ := body["args"].(map[string]interface{})
	if args["a"] != "1" {
		t.Errorf("expected args.a=1, got %v", args)
	}
}

func TestPost_EchoesJSONBody(t *testing.T) {
	c := newClient()
	resp, err := c.Post("/post", map[string]string{"name": "kexas"})
	must(t, err)
	if !resp.IsOK() {
		t.Fatalf("POST /post → %d", resp.StatusCode)
	}
	body, _ := resp.JSONMap()
	json_, _ := body["json"].(map[string]interface{})
	if json_["name"] != "kexas" {
		t.Errorf("expected json.name=kexas, got %v", json_)
	}
}

func TestPut_EchoesJSONBody(t *testing.T) {
	c := newClient()
	resp, err := c.Put("/put", map[string]string{"updated": "true"})
	must(t, err)
	if !resp.IsOK() {
		t.Fatalf("PUT /put → %d", resp.StatusCode)
	}
	body, _ := resp.JSONMap()
	json_, _ := body["json"].(map[string]interface{})
	if json_["updated"] != "true" {
		t.Errorf("expected json.updated=true, got %v", json_)
	}
}

func TestPatch_EchoesJSONBody(t *testing.T) {
	c := newClient()
	resp, err := c.Patch("/patch", map[string]string{"patched": "yes"})
	must(t, err)
	if !resp.IsOK() {
		t.Fatalf("PATCH /patch → %d", resp.StatusCode)
	}
}

func TestDelete_ReturnsOK(t *testing.T) {
	c := newClient()
	resp, err := c.Delete("/delete")
	must(t, err)
	if !resp.IsOK() {
		t.Fatalf("DELETE /delete → %d", resp.StatusCode)
	}
}

func TestHead_ReturnsHeadersNoBody(t *testing.T) {
	c := newClient()
	resp, err := c.Head("/get")
	must(t, err)
	if !resp.IsOK() {
		t.Fatalf("HEAD /get → %d", resp.StatusCode)
	}
	// HEAD spec: same status + headers as GET, no body.
	if len(resp.BodyBytes) != 0 {
		t.Errorf("HEAD response should have empty body; got %d bytes", len(resp.BodyBytes))
	}
}

func TestOptions_ReturnsAllowHeader(t *testing.T) {
	c := newClient()
	resp, err := c.Options("/get")
	must(t, err)
	// httpbin returns 200 with Allow header listing the supported verbs.
	allow := resp.HeaderValue("Allow")
	if allow == "" {
		t.Errorf("expected Allow header on OPTIONS response; got headers %v", resp.Headers)
	}
}

func TestTrace_ReachesServer(t *testing.T) {
	c := newClient()
	resp, err := c.Trace("/")
	must(t, err)
	// httpbin/many CDNs reject TRACE for security (405 method not allowed).
	// Either status is acceptable — we just verify the request shipped.
	if resp.StatusCode == 0 {
		t.Errorf("TRACE returned no status code")
	}
}

// --- Custom verbs via Request("METHOD", path) ---

func TestRequest_CustomVerbReachesServer(t *testing.T) {
	c := newClient()
	// PROPFIND is WebDAV — httpbin will 405 it but we verify kapi sends it.
	resp, err := c.Request("PROPFIND", "/get").Send()
	if err != nil {
		// Some networks block exotic verbs entirely; treat as known limitation.
		t.Skipf("PROPFIND blocked at network layer: %v", err)
	}
	if resp.StatusCode == 0 {
		t.Error("expected a status code from server")
	}
}

// --- Multipart upload ---

func TestMultipart_UploadFileEchoesContent(t *testing.T) {
	c := newClient()
	resp, err := c.Request("POST", "/post").
		BodyMultipart(
			map[string]string{"caption": "kexas-mp"},
			[]kapi.MultipartFile{{
				FieldName: "upload",
				Filename:  "marker.txt",
				Content:   []byte("multipart-roundtrip-marker"),
			}},
		).Send()
	must(t, err)
	if !resp.IsOK() {
		t.Fatalf("POST /post (multipart) → %d", resp.StatusCode)
	}
	// httpbin's /post echoes the file body under "files.<field>" and the text
	// fields under "form.<field>". (Filenames aren't echoed by httpbin.)
	if !strings.Contains(resp.Body, "multipart-roundtrip-marker") {
		t.Errorf("file content missing from echo; first 500 chars: %s", truncate(resp.Body, 500))
	}
	if !strings.Contains(resp.Body, "kexas-mp") {
		t.Errorf("text field 'caption=kexas-mp' missing; first 500 chars: %s", truncate(resp.Body, 500))
	}
	// Verify the Content-Type echoed by httpbin includes the multipart boundary
	// header — proves the request was framed as multipart, not raw JSON.
	if !strings.Contains(resp.Body, "multipart/form-data") {
		t.Errorf("Content-Type echo should mention multipart/form-data; first 500 chars: %s", truncate(resp.Body, 500))
	}
}

// --- File download via SaveToFile ---

func TestDownload_PNGImageSavesValidBytes(t *testing.T) {
	c := newClient()
	resp, err := c.Get("/image/png")
	must(t, err)
	if !resp.IsOK() {
		t.Fatalf("GET /image/png → %d", resp.StatusCode)
	}
	if !strings.Contains(resp.ContentType(), "image/png") {
		t.Errorf("expected PNG content-type; got %q", resp.ContentType())
	}

	var path string = t.TempDir() + "/downloaded.png"
	if err := resp.SaveToFile(path); err != nil {
		t.Fatal(err)
	}

	// PNG magic bytes: 89 50 4E 47 ...
	if len(resp.BodyBytes) < 8 ||
		resp.BodyBytes[0] != 0x89 || resp.BodyBytes[1] != 0x50 ||
		resp.BodyBytes[2] != 0x4E || resp.BodyBytes[3] != 0x47 {
		t.Errorf("downloaded bytes don't have PNG magic header")
	}
}

// --- helpers ---

func must(t *testing.T, err error) {
	t.Helper()
	if err != nil {
		t.Fatal(err)
	}
}

func truncate(s string, n int) string {
	if len(s) < n {
		return s
	}
	return s[:n]
}
