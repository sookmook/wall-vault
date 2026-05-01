package vault

// Browser-side authentication (claim, login, session cookie).
//
// Pre-v0.2.39 the dashboard auth model was: the server embedded admin_token
// directly into the HTML at GET /, and the in-page JS read it back from the
// meta tag and attached it to every HTMX request as Bearer. Anyone who could
// reach GET / on the LAN therefore held the admin token. That worked when the
// vault ran on a strictly trusted LAN but is unworkable for public GitHub
// users who would not know how to provision an admin token before first run.
//
// This module replaces that flow with a click-to-claim + session cookie path:
//   - First boot (admin_token unset): GET / from loopback redirects to /setup.
//     Clicking "초기화" generates admin_token + proxy.vault_token + a new
//     master_password, persists them to the config path, and sets a session
//     cookie. Non-loopback callers see a "claim from localhost" page so an
//     attacker on the LAN cannot front-run the legitimate operator.
//   - Subsequent boot: GET / without cookie redirects to /login. Operator
//     pastes admin_token, server validates, sets cookie. The dashboard's HTMX
//     calls now ride the cookie; the meta-tag admin token is kept for back-
//     compat with API/CLI users but the unauthenticated GET / no longer
//     ships it.
//
// Bearer-token API access (/v1/*, /admin/*, /api/*) is unchanged. The cookie
// is purely an additional accept path for browser sessions.

import (
	"crypto/rand"
	"encoding/hex"
	"fmt"
	"html"
	"io"
	"log"
	"net"
	"net/http"
	"strings"
	"sync"
	"time"

	"github.com/sookmook/wall-vault/internal/config"
)

const (
	sessionCookieName = "wv_session"
	sessionTTL        = 12 * time.Hour
)

// sessionStore is an in-memory map of session id → expiry. Cleared on restart,
// so operators are forced to re-login after the binary restarts; that matches
// the secrets-in-config threat model where a session that outlived the
// process would be a needless surface.
type sessionStore struct {
	mu       sync.Mutex
	sessions map[string]time.Time
}

func newSessionStore() *sessionStore {
	return &sessionStore{sessions: make(map[string]time.Time)}
}

func (s *sessionStore) issue() (string, error) {
	b := make([]byte, 32)
	if _, err := io.ReadFull(rand.Reader, b); err != nil {
		return "", err
	}
	id := hex.EncodeToString(b)
	exp := time.Now().Add(sessionTTL)
	s.mu.Lock()
	// Opportunistically GC expired entries on every issue so the map stays
	// bounded without a separate ticker goroutine.
	now := time.Now()
	for k, t := range s.sessions {
		if t.Before(now) {
			delete(s.sessions, k)
		}
	}
	s.sessions[id] = exp
	s.mu.Unlock()
	return id, nil
}

func (s *sessionStore) valid(id string) bool {
	if id == "" {
		return false
	}
	s.mu.Lock()
	defer s.mu.Unlock()
	exp, ok := s.sessions[id]
	if !ok {
		return false
	}
	if time.Now().After(exp) {
		delete(s.sessions, id)
		return false
	}
	return true
}

func (s *sessionStore) revoke(id string) {
	s.mu.Lock()
	delete(s.sessions, id)
	s.mu.Unlock()
}

// hasValidSession returns true if the request carries a session cookie that
// matches a live entry in the in-memory store. Used by adminAuth/clientAuth
// as an alternative to Bearer.
func (s *Server) hasValidSession(r *http.Request) bool {
	if s.sessions == nil {
		return false
	}
	c, err := r.Cookie(sessionCookieName)
	if err != nil {
		return false
	}
	return s.sessions.valid(c.Value)
}

// setSessionCookie issues a fresh session id and writes it back to the
// browser. Caller is responsible for redirecting / rendering the post-login
// page.
func (s *Server) setSessionCookie(w http.ResponseWriter, r *http.Request) error {
	id, err := s.sessions.issue()
	if err != nil {
		return err
	}
	// Secure flag mirrors the listener: if vault runs over TLS the cookie
	// must not leak over plain HTTP, but on localhost-only HTTP boots the
	// cookie has to be deliverable so a browser will actually attach it.
	secure := r.TLS != nil
	http.SetCookie(w, &http.Cookie{
		Name:     sessionCookieName,
		Value:    id,
		Path:     "/",
		HttpOnly: true,
		SameSite: http.SameSiteLaxMode,
		Secure:   secure,
		Expires:  time.Now().Add(sessionTTL),
	})
	return nil
}

func (s *Server) clearSessionCookie(w http.ResponseWriter) {
	http.SetCookie(w, &http.Cookie{
		Name:     sessionCookieName,
		Value:    "",
		Path:     "/",
		HttpOnly: true,
		SameSite: http.SameSiteLaxMode,
		MaxAge:   -1,
	})
}

// isLoopback reports whether the request originated from the same host. Only
// loopback requests are allowed to claim an unconfigured vault, so an
// attacker on the LAN can't front-run the legitimate operator.
func isLoopback(r *http.Request) bool {
	ip, _, err := net.SplitHostPort(r.RemoteAddr)
	if err != nil {
		ip = r.RemoteAddr
	}
	parsed := net.ParseIP(ip)
	if parsed == nil {
		return false
	}
	return parsed.IsLoopback()
}

// ─── Handlers ────────────────────────────────────────────────────────────────

// handleSetup serves the first-run claim flow. GET shows the page (or a
// notice that claim is locked because admin_token is already set), POST
// generates the secrets and persists them. Loopback-only on POST.
func (s *Server) handleSetup(w http.ResponseWriter, r *http.Request) {
	switch r.Method {
	case http.MethodGet:
		s.renderSetupPage(w, r, "")
	case http.MethodPost:
		if s.cfg.Vault.AdminToken != "" {
			s.renderSetupPage(w, r, "이미 초기화되어 있습니다. /login 으로 이동하세요.")
			return
		}
		if !isLoopback(r) {
			http.Error(w, "claim must originate from localhost", http.StatusForbidden)
			return
		}
		if err := s.claim(); err != nil {
			log.Printf("[vault] claim failed: %v", err)
			http.Error(w, "claim failed: "+err.Error(), http.StatusInternalServerError)
			return
		}
		if err := s.setSessionCookie(w, r); err != nil {
			http.Error(w, "session error", http.StatusInternalServerError)
			return
		}
		http.Redirect(w, r, "/", http.StatusSeeOther)
	default:
		http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
	}
}

// claim provisions admin_token + proxy.vault_token (+ master_password if
// missing) and writes them back to the config file. Only called from /setup
// after the request has been validated as loopback.
func (s *Server) claim() error {
	cfg := s.cfg
	if cfg.Vault.AdminToken != "" {
		return fmt.Errorf("already claimed")
	}
	adminTok, err := randomToken(24)
	if err != nil {
		return err
	}
	cfg.Vault.AdminToken = adminTok
	if cfg.Proxy.VaultToken == "" {
		proxyTok, err := randomToken(24)
		if err != nil {
			return err
		}
		cfg.Proxy.VaultToken = proxyTok
	}
	if cfg.Vault.MasterPass == "" {
		mp, err := randomToken(32)
		if err != nil {
			return err
		}
		cfg.Vault.MasterPass = mp
	}
	path := s.cfgPath
	if path == "" {
		path = "wall-vault.yaml"
	}
	if err := config.Save(cfg, path); err != nil {
		return fmt.Errorf("save %s: %w", path, err)
	}
	log.Printf("[vault] claimed via web: admin_token + proxy.vault_token written to %s", path)
	return nil
}

// handleLogin serves the password form (GET) or validates a submitted token
// and sets a session cookie (POST).
func (s *Server) handleLogin(w http.ResponseWriter, r *http.Request) {
	switch r.Method {
	case http.MethodGet:
		s.renderLoginPage(w, r, "")
	case http.MethodPost:
		if err := r.ParseForm(); err != nil {
			s.renderLoginPage(w, r, "form parse error")
			return
		}
		ip := realIP(r)
		if s.limiter.blocked(ip) {
			http.Error(w, "too many failed attempts", http.StatusTooManyRequests)
			return
		}
		token := strings.TrimSpace(r.PostForm.Get("token"))
		if !secureCompare(token, s.cfg.Vault.AdminToken) {
			s.limiter.record(ip)
			s.renderLoginPage(w, r, "invalid token")
			return
		}
		if err := s.setSessionCookie(w, r); err != nil {
			http.Error(w, "session error", http.StatusInternalServerError)
			return
		}
		http.Redirect(w, r, "/", http.StatusSeeOther)
	default:
		http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
	}
}

// handleLogout clears the session cookie and redirects to /login.
func (s *Server) handleLogout(w http.ResponseWriter, r *http.Request) {
	if c, err := r.Cookie(sessionCookieName); err == nil {
		s.sessions.revoke(c.Value)
	}
	s.clearSessionCookie(w)
	http.Redirect(w, r, "/login", http.StatusSeeOther)
}

// ─── HTML pages (inline so we don't depend on templ regeneration) ────────────

func (s *Server) renderSetupPage(w http.ResponseWriter, r *http.Request, errMsg string) {
	w.Header().Set("Content-Type", "text/html; charset=utf-8")
	if s.cfg.Vault.AdminToken != "" {
		_, _ = w.Write([]byte(authPage("이미 초기화됨", `
<p>이 벽금고 인스턴스는 이미 초기화되어 있습니다.</p>
<p><a href="/login">로그인 페이지로 이동</a></p>`)))
		return
	}
	if !isLoopback(r) {
		_, _ = w.Write([]byte(authPage("초기화는 localhost 에서", `
<p>처음 사용을 시작하시려면 벽금고/프록시가 실행 중인 머신에서
직접 <code>http://localhost:`+fmt.Sprintf("%d", s.cfg.Vault.Port)+`/setup</code> 을 여세요.</p>
<p>외부 IP 에서의 첫 클레임은 보안상 차단됩니다.</p>`)))
		return
	}
	body := `
<p>벽금고/프록시 첫 부팅입니다. 아래 버튼을 누르면:</p>
<ul>
<li>관리자 토큰 (<code>admin_token</code>) 자동 생성</li>
<li>프록시 인증 토큰 (<code>proxy.vault_token</code>) 자동 생성</li>
<li>API 키 암호화 비밀번호 (<code>master_password</code>) 자동 생성</li>
<li>위 값들을 설정 파일에 안전하게 저장</li>
<li>이 브라우저에 세션 쿠키 발급 → 대시보드 진입</li>
</ul>
<form method="POST" action="/setup">
<button type="submit">초기화 진행</button>
</form>`
	if errMsg != "" {
		body = `<p class="err">` + html.EscapeString(errMsg) + `</p>` + body
	}
	_, _ = w.Write([]byte(authPage("벽금고 초기 설정", body)))
}

func (s *Server) renderLoginPage(w http.ResponseWriter, r *http.Request, errMsg string) {
	w.Header().Set("Content-Type", "text/html; charset=utf-8")
	body := `
<p>관리자 토큰을 입력하세요. 토큰은 첫 부팅 후 대시보드의 "API 토큰" 카드 또는
설정 파일 (<code>vault.admin_token</code>) 에서 확인할 수 있습니다.</p>
<form method="POST" action="/login">
<input type="password" name="token" autocomplete="current-password" required autofocus placeholder="admin_token"/>
<button type="submit">로그인</button>
</form>`
	if errMsg != "" {
		body = `<p class="err">` + html.EscapeString(errMsg) + `</p>` + body
	}
	_, _ = w.Write([]byte(authPage("벽금고 로그인", body)))
}

// authPage wraps body HTML in a minimal page chrome shared between /setup
// and /login. Inline CSS only — these pages render before the dashboard
// stylesheet is available, and we don't want them to depend on the templ
// build pipeline.
func authPage(title, body string) string {
	return `<!DOCTYPE html>
<html lang="ko">
<head>
<meta charset="utf-8"/>
<meta name="viewport" content="width=device-width,initial-scale=1"/>
<title>` + html.EscapeString(title) + `</title>
<style>
body{font-family:system-ui,-apple-system,sans-serif;background:#f7f7f8;color:#1f2937;margin:0;padding:0;display:flex;align-items:center;justify-content:center;min-height:100vh;}
.box{background:#fff;border:1px solid #e5e7eb;border-radius:8px;padding:32px;max-width:520px;width:90%;box-shadow:0 4px 16px rgba(0,0,0,0.04);}
h1{margin:0 0 16px;font-size:20px;}
ul{padding-left:18px;}
code{background:#f3f4f6;padding:1px 5px;border-radius:3px;font-size:90%;}
input[type=password]{width:100%;padding:10px;border:1px solid #d1d5db;border-radius:6px;margin:8px 0;box-sizing:border-box;}
button{background:#2563eb;color:#fff;border:0;border-radius:6px;padding:10px 18px;font-weight:600;cursor:pointer;}
button:hover{background:#1d4ed8;}
.err{color:#b91c1c;background:#fef2f2;border:1px solid #fecaca;padding:8px 10px;border-radius:6px;}
a{color:#2563eb;}
</style>
</head>
<body>
<div class="box">
<h1>` + html.EscapeString(title) + `</h1>` + body + `
</div>
</body>
</html>`
}

// ─── Helpers ─────────────────────────────────────────────────────────────────

func randomToken(byteLen int) (string, error) {
	b := make([]byte, byteLen)
	if _, err := io.ReadFull(rand.Reader, b); err != nil {
		return "", err
	}
	return hex.EncodeToString(b), nil
}
