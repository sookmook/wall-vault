package proxy

// Internal HTTPS trust for proxy → vault calls.
//
// When vault.tls.enabled flips to true the vault listener serves its
// per-host certificate signed by the wall-vault internal CA. The proxy
// running on the same fleet talks to that listener via HTTPS, but Go's
// default http.Client uses the OS trust store — which on macOS doesn't
// always pick up CA certs added via the GUI keychain dialog, and on Linux
// requires update-ca-certificates to have run with the cert in
// /usr/local/share/ca-certificates.
//
// Rather than depending on per-host trust-store configuration we ship the
// CA inside wall-vault: every host the cert tooling provisions has
// ~/.wall-vault/ca.crt, and that's the same trust anchor the proxy needs
// to verify the vault listener. This file builds an http.Client whose
// TLS config has RootCAs explicitly seeded with that CA, falling back to
// the OS pool when the file isn't present so existing fleet hosts that
// disabled TLS keep working.
//
// Cloud API calls (Anthropic / OpenAI / Google / OpenRouter) deliberately
// keep using the bare http.Client — those endpoints are signed by public
// CAs and adding a private root to that path would only widen the trust
// surface without benefit.

import (
	"crypto/tls"
	"crypto/x509"
	"net/http"
	"os"
	"path/filepath"
	"sync"
	"time"
)

var (
	internalTrustOnce sync.Once
	internalTrustPool *x509.CertPool // nil → fall back to system pool
)

// internalRoots loads ~/.wall-vault/ca.crt into a CertPool the first time
// it's called and caches the result. Returns nil when the CA file is
// missing or unreadable so callers fall back to system trust.
func internalRoots() *x509.CertPool {
	internalTrustOnce.Do(func() {
		caPath := internalCAPath()
		if caPath == "" {
			return
		}
		pem, err := os.ReadFile(caPath)
		if err != nil {
			return
		}
		// Start from the system pool when available so wall-vault hosts can
		// still verify Let's Encrypt etc. when the same client gets reused
		// for a non-vault target. SystemCertPool() can fail on minimal
		// containers; an empty pool with only our CA is still better than
		// blanket InsecureSkipVerify.
		pool, err := x509.SystemCertPool()
		if err != nil || pool == nil {
			pool = x509.NewCertPool()
		}
		if !pool.AppendCertsFromPEM(pem) {
			// Bad PEM — leave internalTrustPool nil so we fall back rather
			// than silently breaking every vault call.
			return
		}
		internalTrustPool = pool
	})
	return internalTrustPool
}

// internalCAPath mirrors the cert tooling's resolution order:
//   1. WV_CERT_DIR/ca.crt
//   2. ~/.wall-vault/ca.crt
// Returns empty when neither file exists; callers then fall back to OS trust.
func internalCAPath() string {
	if env := os.Getenv("WV_CERT_DIR"); env != "" {
		p := filepath.Join(env, "ca.crt")
		if _, err := os.Stat(p); err == nil {
			return p
		}
	}
	home, err := os.UserHomeDir()
	if err != nil {
		return ""
	}
	p := filepath.Join(home, ".wall-vault", "ca.crt")
	if _, err := os.Stat(p); err == nil {
		return p
	}
	return ""
}

// internalHTTPClient returns an http.Client safe to point at the wall-vault
// vault HTTPS listener. Reuses one Transport per call site so callers don't
// each build their own pool — if RAM ever becomes a concern this is the
// place to pool, but per-call construction was already the pre-existing
// pattern so the cost is bounded.
func internalHTTPClient(timeout time.Duration) *http.Client {
	tr := &http.Transport{
		TLSClientConfig: &tls.Config{
			RootCAs:    internalRoots(),
			MinVersion: tls.VersionTLS12,
		},
	}
	return &http.Client{Timeout: timeout, Transport: tr}
}
