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
	"crypto/hmac"
	"crypto/rand"
	"crypto/sha256"
	"encoding/hex"
	"fmt"
	"html"
	"io"
	"log"
	"net"
	"net/http"
	"strconv"
	"strings"
	"sync"
	"time"

	"github.com/sookmook/wall-vault/internal/config"
	"github.com/sookmook/wall-vault/internal/i18n"
)

const (
	sessionCookieName  = "wv_session"
	sessionTTL         = 12 * time.Hour
	rememberSessionTTL = 30 * 24 * time.Hour
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
// matches a live entry in the in-memory store, OR a valid HMAC-signed
// "remember this device" cookie. Used by adminAuth/clientAuth as an
// alternative to Bearer.
func (s *Server) hasValidSession(r *http.Request) bool {
	c, err := r.Cookie(sessionCookieName)
	if err != nil {
		return false
	}
	// HMAC-signed remember cookies always contain a "." separating the
	// expiry from the signature; plain in-memory session ids are pure hex
	// so they cannot collide with this format.
	if strings.Contains(c.Value, ".") {
		return s.validRememberCookie(c.Value)
	}
	if s.sessions == nil {
		return false
	}
	return s.sessions.valid(c.Value)
}

// signRememberCookie produces an HMAC-signed cookie value of the form
// `<expiryUnix>.<hexHmac>`. The HMAC key is the admin token, so changing the
// admin token instantly invalidates every outstanding remember cookie — that
// is the only revocation path (there is no server-side store of remember
// cookies, by design: a stolen value cannot survive an admin-token rotation).
func (s *Server) signRememberCookie(expiry int64) string {
	msg := strconv.FormatInt(expiry, 10)
	mac := hmac.New(sha256.New, []byte(s.cfg.Vault.AdminToken))
	mac.Write([]byte(msg))
	return msg + "." + hex.EncodeToString(mac.Sum(nil))
}

// validRememberCookie verifies an HMAC-signed remember cookie and checks its
// embedded expiry. Returns false on any malformed input — never panic.
func (s *Server) validRememberCookie(val string) bool {
	if s.cfg.Vault.AdminToken == "" {
		return false
	}
	dot := strings.IndexByte(val, '.')
	if dot <= 0 || dot == len(val)-1 {
		return false
	}
	expiry, err := strconv.ParseInt(val[:dot], 10, 64)
	if err != nil {
		return false
	}
	if time.Now().Unix() > expiry {
		return false
	}
	expected := s.signRememberCookie(expiry)
	return hmac.Equal([]byte(val), []byte(expected))
}

// setRememberCookie writes a long-lived HMAC-signed cookie that survives
// process restarts. Used when the operator ticked the "remember this device"
// toggle on the login page.
func (s *Server) setRememberCookie(w http.ResponseWriter, r *http.Request) {
	expiry := time.Now().Add(rememberSessionTTL).Unix()
	val := s.signRememberCookie(expiry)
	secure := r.TLS != nil
	http.SetCookie(w, &http.Cookie{
		Name:     sessionCookieName,
		Value:    val,
		Path:     "/",
		HttpOnly: true,
		SameSite: http.SameSiteLaxMode,
		Secure:   secure,
		Expires:  time.Unix(expiry, 0),
	})
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
		lang := s.authLang(r)
		if err := r.ParseForm(); err != nil {
			s.renderLoginPage(w, r, i18n.TFor(lang, "auth_err_form"))
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
			s.renderLoginPage(w, r, i18n.TFor(lang, "auth_err_invalid_token"))
			return
		}
		if r.PostForm.Get("remember") != "" {
			s.setRememberCookie(w, r)
		} else if err := s.setSessionCookie(w, r); err != nil {
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

// authLang picks the locale for an unauthenticated auth page. Sessionless
// callers don't carry the dashboard's lang cookie yet, so we honour
// (1) the URL ?lang=xx override (lets the operator force a locale),
// (2) the browser's Accept-Language preferences,
// (3) the server's configured default.
func (s *Server) authLang(r *http.Request) string {
	if v := r.URL.Query().Get("lang"); v != "" && i18nSupported(v) {
		return v
	}
	if al := r.Header.Get("Accept-Language"); al != "" {
		if code := matchAcceptLanguage(al); code != "" {
			return code
		}
	}
	return s.currentLang()
}

func (s *Server) renderSetupPage(w http.ResponseWriter, r *http.Request, errMsg string) {
	w.Header().Set("Content-Type", "text/html; charset=utf-8")
	lang := s.authLang(r)
	if s.cfg.Vault.AdminToken != "" {
		body := `<p>` + html.EscapeString(i18n.TFor(lang, "auth_already_body")) + `</p>` +
			`<p><a href="/login">` + html.EscapeString(i18n.TFor(lang, "auth_already_login_link")) + `</a></p>`
		_, _ = w.Write([]byte(authPage(i18n.TFor(lang, "auth_already_title"), body)))
		return
	}
	if !isLoopback(r) {
		setupURL := fmt.Sprintf("http://localhost:%d/setup", s.cfg.Vault.Port)
		body := `<p>` + html.EscapeString(i18n.TFor(lang, "auth_loopback_body")) + ` <code>` + html.EscapeString(setupURL) + `</code></p>` +
			`<p>` + html.EscapeString(i18n.TFor(lang, "auth_loopback_warn")) + `</p>`
		_, _ = w.Write([]byte(authPage(i18n.TFor(lang, "auth_loopback_title"), body)))
		return
	}
	body := `<p>` + html.EscapeString(i18n.TFor(lang, "auth_setup_intro")) + `</p>
<ul>
<li>` + html.EscapeString(i18n.TFor(lang, "auth_setup_b1")) + `</li>
<li>` + html.EscapeString(i18n.TFor(lang, "auth_setup_b2")) + `</li>
<li>` + html.EscapeString(i18n.TFor(lang, "auth_setup_b3")) + `</li>
<li>` + html.EscapeString(i18n.TFor(lang, "auth_setup_b4")) + `</li>
<li>` + html.EscapeString(i18n.TFor(lang, "auth_setup_b5")) + `</li>
</ul>
<form method="POST" action="/setup">
<button type="submit">` + html.EscapeString(i18n.TFor(lang, "auth_setup_button")) + `</button>
</form>`
	if errMsg != "" {
		body = `<p class="err">` + html.EscapeString(errMsg) + `</p>` + body
	}
	_, _ = w.Write([]byte(authPage(i18n.TFor(lang, "auth_setup_title"), body)))
}

func (s *Server) renderLoginPage(w http.ResponseWriter, r *http.Request, errMsg string) {
	w.Header().Set("Content-Type", "text/html; charset=utf-8")
	lang := s.authLang(r)
	body := `<p>` + html.EscapeString(i18n.TFor(lang, "auth_login_intro")) + `</p>
<form method="POST" action="/login">
<input type="password" name="token" autocomplete="current-password" required autofocus placeholder="` + html.EscapeString(i18n.TFor(lang, "auth_login_field_ph")) + `"/>
<label class="toggle">
  <input type="checkbox" name="remember" value="1"/>
  <span class="slider"></span>
  <span class="toggle-label">` + html.EscapeString(i18n.TFor(lang, "auth_login_remember")) + `</span>
</label>
<button type="submit">` + html.EscapeString(i18n.TFor(lang, "auth_login_submit")) + `</button>
</form>`
	if errMsg != "" {
		body = `<p class="err">` + html.EscapeString(errMsg) + `</p>` + body
	}
	_, _ = w.Write([]byte(authPage(i18n.TFor(lang, "auth_login_title"), body)))
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
.toggle{display:inline-flex;align-items:center;gap:10px;margin:12px 0 16px;cursor:pointer;user-select:none;}
.toggle input{position:absolute;opacity:0;width:0;height:0;pointer-events:none;}
.toggle .slider{position:relative;flex:0 0 auto;width:36px;height:20px;background:#cbd5e1;border-radius:10px;transition:background .15s;}
.toggle .slider::after{content:"";position:absolute;left:2px;top:2px;width:16px;height:16px;background:#fff;border-radius:50%;box-shadow:0 1px 2px rgba(0,0,0,.15);transition:left .15s;}
.toggle input:checked + .slider{background:#2563eb;}
.toggle input:checked + .slider::after{left:18px;}
.toggle input:focus-visible + .slider{outline:2px solid #2563eb;outline-offset:2px;}
.toggle-label{font-size:14px;color:#374151;}
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
