package vault

import (
	"io"
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

// TestBootstrapHandlerCacrt verifies that GET /ca.crt streams the cert file
// with PEM content-type + an attachment filename. The whole point of the
// bootstrap listener is letting first-time clients curl it down without
// auth, so this is the load-bearing route.
func TestBootstrapHandlerCacrt(t *testing.T) {
	dir := t.TempDir()
	caPath := filepath.Join(dir, "ca.crt")
	pem := "-----BEGIN CERTIFICATE-----\nfake-ca-data\n-----END CERTIFICATE-----\n"
	if err := os.WriteFile(caPath, []byte(pem), 0o644); err != nil {
		t.Fatalf("seed ca.crt: %v", err)
	}

	srv := httptest.NewServer(BootstrapHandler(caPath))
	defer srv.Close()

	resp, err := http.Get(srv.URL + "/ca.crt")
	if err != nil {
		t.Fatalf("GET /ca.crt: %v", err)
	}
	defer resp.Body.Close()
	if resp.StatusCode != 200 {
		t.Fatalf("status = %d, want 200", resp.StatusCode)
	}
	if ct := resp.Header.Get("Content-Type"); ct != "application/x-pem-file" {
		t.Errorf("Content-Type = %q, want application/x-pem-file", ct)
	}
	if cd := resp.Header.Get("Content-Disposition"); !strings.Contains(cd, "wall-vault-ca.crt") {
		t.Errorf("Content-Disposition missing filename: %q", cd)
	}
	body, _ := io.ReadAll(resp.Body)
	if string(body) != pem {
		t.Errorf("body mismatch: %q", body)
	}
}

// TestBootstrapHandlerCacrtMissing — when ca.crt isn't provisioned the
// listener should 404 with an instructional message rather than 500. New
// users running `wall-vault start` before `wall-vault cert init` shouldn't
// see a server error.
func TestBootstrapHandlerCacrtMissing(t *testing.T) {
	srv := httptest.NewServer(BootstrapHandler("/nonexistent/ca.crt"))
	defer srv.Close()

	resp, err := http.Get(srv.URL + "/ca.crt")
	if err != nil {
		t.Fatalf("GET: %v", err)
	}
	defer resp.Body.Close()
	if resp.StatusCode != 404 {
		t.Errorf("status = %d, want 404", resp.StatusCode)
	}
	body, _ := io.ReadAll(resp.Body)
	if !strings.Contains(string(body), "wall-vault cert init") {
		t.Errorf("missing actionable hint in 404: %s", body)
	}
}

// TestBootstrapHandlerIndex renders the install help page with per-OS
// snippets. We only assert the headings and the download link are present —
// the snippet content is template-rendered and changing it shouldn't break
// this test.
func TestBootstrapHandlerIndex(t *testing.T) {
	srv := httptest.NewServer(BootstrapHandler("/tmp/whatever"))
	defer srv.Close()

	resp, err := http.Get(srv.URL + "/")
	if err != nil {
		t.Fatalf("GET /: %v", err)
	}
	defer resp.Body.Close()
	body, _ := io.ReadAll(resp.Body)
	s := string(body)
	for _, want := range []string{"ca.crt", "Linux", "macOS", "Windows", "SSL_CERT_FILE", "NODE_EXTRA_CA_CERTS"} {
		if !strings.Contains(s, want) {
			t.Errorf("index missing %q", want)
		}
	}
}

// TestResolveCAPath_Override — the explicit override wins when the file
// exists. Mirrors the cert tooling's --dir convention.
func TestResolveCAPath_Override(t *testing.T) {
	dir := t.TempDir()
	p := filepath.Join(dir, "explicit.crt")
	_ = os.WriteFile(p, []byte("x"), 0o644)
	if got := ResolveCAPath(p); got != p {
		t.Errorf("override ignored: got %q, want %q", got, p)
	}
}

// TestResolveCAPath_NotFound — every candidate missing returns "" so the
// /ca.crt handler can render the actionable 404 instead of failing.
func TestResolveCAPath_NotFound(t *testing.T) {
	t.Setenv("WV_CERT_DIR", t.TempDir())            // exists but no ca.crt
	t.Setenv("HOME", filepath.Join(t.TempDir(), "no-home"))
	if got := ResolveCAPath(""); got != "" {
		t.Errorf("missing CA should return empty, got %q", got)
	}
}
