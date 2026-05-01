package vault

// Plain-HTTP bootstrap listener for first-time clients that need to download
// the wall-vault internal CA before they can speak HTTPS to the main vault
// listener (catch-22: they need to trust our cert, but the cert itself is
// only reachable over HTTPS we don't trust yet).
//
// Listens on a separate port (default 56247) and serves only a small set of
// public-information endpoints:
//
//   GET /          — minimal install-instruction page (per-OS one-liners)
//   GET /ca.crt    — the CA certificate as PEM
//   GET /health    — liveness probe
//
// The CA certificate is public information by design — it is the trust anchor
// we want every client to ship — so exposing it without auth is safe. Anything
// secret (admin tokens, API keys, vault data) stays on the main HTTPS
// listener. Operators who don't want the extra port can disable it with
// vault.bootstrap_port=0 in the config.

import (
	"fmt"
	"html"
	"io"
	"log"
	"net"
	"net/http"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/sookmook/wall-vault/internal/config"
)

// BootstrapHandler builds the plain-HTTP handler used by NewBootstrapServer.
// Exposed for tests so they can hit individual routes via httptest without
// binding a real port.
func BootstrapHandler(caPath string) http.Handler {
	mux := http.NewServeMux()
	mux.HandleFunc("/health", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		_, _ = w.Write([]byte(`{"status":"ok"}`))
	})
	mux.HandleFunc("/ca.crt", func(w http.ResponseWriter, r *http.Request) {
		f, err := os.Open(caPath)
		if err != nil {
			http.Error(w, "ca.crt not provisioned — run `wall-vault cert init` on the vault host", http.StatusNotFound)
			return
		}
		defer f.Close()
		w.Header().Set("Content-Type", "application/x-pem-file")
		w.Header().Set("Content-Disposition", `attachment; filename="wall-vault-ca.crt"`)
		_, _ = io.Copy(w, f)
	})
	mux.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/" {
			http.NotFound(w, r)
			return
		}
		host := r.Host // includes port
		w.Header().Set("Content-Type", "text/html; charset=utf-8")
		_, _ = w.Write([]byte(bootstrapIndexHTML(host)))
	})
	return mux
}

// NewBootstrapServer returns an *http.Server pre-wired to BootstrapHandler.
// Caller is responsible for ListenAndServe + Shutdown so it integrates with
// the rest of main.go's lifecycle the same way the vault/proxy listeners do.
func NewBootstrapServer(addr, caPath string) *http.Server {
	return &http.Server{
		Addr:              addr,
		Handler:           BootstrapHandler(caPath),
		ReadHeaderTimeout: 5 * time.Second,
	}
}

// ResolveCAPath picks the ca.crt the bootstrap listener will serve. Mirrors
// the cert tooling's search order: explicit override → WV_CERT_DIR →
// ~/.wall-vault/ca.crt. Returns empty string if no candidate exists; the
// /ca.crt handler then 404s with an instructional message.
func ResolveCAPath(override string) string {
	if override != "" {
		if _, err := os.Stat(override); err == nil {
			return override
		}
	}
	if env := os.Getenv("WV_CERT_DIR"); env != "" {
		p := filepath.Join(env, "ca.crt")
		if _, err := os.Stat(p); err == nil {
			return p
		}
	}
	if home, err := os.UserHomeDir(); err == nil {
		p := filepath.Join(home, ".wall-vault", "ca.crt")
		if _, err := os.Stat(p); err == nil {
			return p
		}
	}
	return ""
}

// BootstrapAddr formats the listen address for the bootstrap server. We
// always reuse cfg.Vault.Host so the bootstrap listener is reachable from
// the same callers that already speak to the vault — applyHostDefaults has
// already filled this in (loopback in standalone, 0.0.0.0 in distributed).
// Hard-coding a separate policy here drifted out of sync on hosts where the
// vault was bound to 0.0.0.0 but cfg.Mode happened to read "standalone",
// leaving the bootstrap listener stuck on 127.0.0.1 and unreachable from
// the LAN clients it was supposed to bootstrap.
func BootstrapAddr(cfg *config.Config) string {
	host := cfg.Vault.Host
	if host == "" {
		host = "127.0.0.1"
	}
	return net.JoinHostPort(host, fmt.Sprintf("%d", cfg.Vault.BootstrapPort))
}

// bootstrapIndexHTML is rendered by GET / on the bootstrap listener. The
// install-instruction snippets are the same ones used by `wall-vault cert
// install-trust`, but here we hand them to the user copy-pasteable so the
// flow works on machines that don't have the wall-vault binary (e.g. a
// Windows RDP host that just needs to trust the CA in Chrome).
func bootstrapIndexHTML(host string) string {
	dl := "http://" + host + "/ca.crt"
	dlEsc := html.EscapeString(dl)
	return `<!DOCTYPE html>
<html lang="ko">
<head>
<meta charset="utf-8"/>
<title>wall-vault CA bootstrap</title>
<style>
body{font-family:system-ui,sans-serif;max-width:720px;margin:40px auto;padding:0 20px;color:#1f2937;}
h1{font-size:22px;margin:0 0 10px;}
h2{font-size:16px;margin:24px 0 8px;}
code,pre{background:#f3f4f6;border-radius:4px;font-size:90%;}
code{padding:1px 5px;}
pre{padding:10px;overflow:auto;}
a{color:#2563eb;}
.dl{display:inline-block;background:#2563eb;color:#fff;padding:8px 14px;border-radius:6px;text-decoration:none;font-weight:600;margin:6px 0;}
.warn{background:#fef3c7;border:1px solid #fde68a;padding:8px 12px;border-radius:6px;font-size:90%;}
</style>
</head>
<body>
<h1>wall-vault CA — 신뢰 저장소 등록</h1>
<p>이 페이지는 평문 HTTP 로 노출되는 부트스트랩 페이지입니다. CA 인증서 자체는 공개 정보라 인증 없이 다운로드할 수 있도록 의도적으로 열어둔 경로입니다.</p>

<p><a class="dl" href="/ca.crt" download="wall-vault-ca.crt">📥 ca.crt 다운로드</a></p>
<p>또는 명령줄에서: <code>curl -O ` + dlEsc + `</code></p>

<h2>Linux (Debian/Ubuntu)</h2>
<pre>sudo cp ca.crt /usr/local/share/ca-certificates/wall-vault-ca.crt
sudo update-ca-certificates</pre>

<h2>macOS</h2>
<pre>sudo security add-trusted-cert -d -r trustRoot \
  -k /Library/Keychains/System.keychain ca.crt</pre>
<p>키체인 다이얼로그가 뜨면 관리자 비밀번호 입력.</p>

<h2>Windows (관리자 PowerShell)</h2>
<pre>certutil -addstore -f Root .\ca.crt</pre>
<p>관리자 권한이 없으면 사용자 저장소만:</p>
<pre>certutil -user -addstore -f Root .\ca.crt</pre>

<h2>Python / OpenAI SDK / requests / httpx</h2>
<p>시스템 신뢰 저장소 대신 환경변수로 지정:</p>
<pre>export SSL_CERT_FILE=/path/to/ca.crt</pre>

<h2>Node.js</h2>
<pre>export NODE_EXTRA_CA_CERTS=/path/to/ca.crt</pre>

<h2>wall-vault 바이너리가 깔린 머신</h2>
<p>위 단계를 자동으로 처리:</p>
<pre>wall-vault cert install-trust</pre>

<p class="warn">⚠️ 이 부트스트랩 listener 는 평문 HTTP 로 ca.crt 만 노출합니다. admin token, API 키 등 일체의 비밀은 메인 HTTPS 리스너 (포트 56243) 에서만 제공됩니다.</p>
</body>
</html>`
}

// LogStartupHint mirrors the main.go startup hint pattern: a short
// announcement so operators staring at the terminal see where to point
// first-time clients. Called by main.go after the bootstrap listener boots.
func LogStartupHint(addr string) {
	if strings.HasSuffix(addr, ":0") {
		// disabled
		return
	}
	host, port, _ := net.SplitHostPort(addr)
	if host == "" {
		host = "localhost"
	}
	if host == "0.0.0.0" || host == "::" {
		host = "<host>"
	}
	log.Printf("[bootstrap] CA distribution: http://%s:%s/ca.crt", host, port)
}
