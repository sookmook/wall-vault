// Package middleware: common HTTP middleware
package middleware

import (
	"log"
	"net"
	"net/http"
	"strings"
	"time"
)

// responseWriter: captures response status code
type responseWriter struct {
	http.ResponseWriter
	status int
	size   int
}

func (rw *responseWriter) WriteHeader(status int) {
	rw.status = status
	rw.ResponseWriter.WriteHeader(status)
}

func (rw *responseWriter) Write(b []byte) (int, error) {
	n, err := rw.ResponseWriter.Write(b)
	rw.size += n
	return n, err
}

// Flush: implements http.Flusher — required for SSE to work correctly
func (rw *responseWriter) Flush() {
	if f, ok := rw.ResponseWriter.(http.Flusher); ok {
		f.Flush()
	}
}

// Logger: request logging middleware
func Logger(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		start := time.Now()
		rw := &responseWriter{ResponseWriter: w, status: http.StatusOK}
		next.ServeHTTP(rw, r)
		elapsed := time.Since(start)

		// skip logging for /api/events (SSE) and /health
		if r.URL.Path == "/api/events" || r.URL.Path == "/health" {
			return
		}
		log.Printf("[%s] %s %s %d %v",
			time.Now().Format("15:04:05"),
			r.Method, r.URL.Path,
			rw.status, elapsed.Round(time.Millisecond),
		)
	})
}

// isAllowedOrigin checks if the origin is from a local network address.
// Accepts loopback (localhost/127.0.0.1/::1) and any RFC1918 private address
// (10/8, 172.16/12, 192.168/16) so corporate networks using 10.x or Docker-ish
// 172.16-31.x subnets aren't refused when all traffic is already LAN-scoped.
func isAllowedOrigin(origin string) bool {
	if origin == "" {
		return false
	}
	// strip scheme to get host:port
	host := origin
	for _, prefix := range []string{"http://", "https://"} {
		if strings.HasPrefix(host, prefix) {
			host = strings.TrimPrefix(host, prefix)
			break
		}
	}
	// strip port — handle [::1]:port form by splitting via SplitHostPort first.
	if h, _, err := net.SplitHostPort(host); err == nil {
		host = h
	} else if idx := strings.LastIndex(host, ":"); idx != -1 {
		host = host[:idx]
	}
	host = strings.TrimPrefix(strings.TrimSuffix(host, "]"), "[")
	if host == "localhost" {
		return true
	}
	if ip := net.ParseIP(host); ip != nil {
		return ip.IsLoopback() || ip.IsPrivate() || ip.IsLinkLocalUnicast()
	}
	return false
}

// CORS: CORS header middleware — only allows local network origins
func CORS(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		origin := r.Header.Get("Origin")
		if isAllowedOrigin(origin) {
			w.Header().Set("Access-Control-Allow-Origin", origin)
			w.Header().Set("Vary", "Origin")
		}
		w.Header().Set("Access-Control-Allow-Methods", "GET, POST, PUT, DELETE, OPTIONS")
		w.Header().Set("Access-Control-Allow-Headers", "Content-Type, Authorization")
		if r.Method == http.MethodOptions {
			w.WriteHeader(http.StatusNoContent)
			return
		}
		next.ServeHTTP(w, r)
	})
}

// SecurityHeaders sets conservative HTTP response headers on every response.
// Kept to the ones that are safe under wall-vault's embedded HTMX script and
// small inline styles (no external CDN frames, no third-party scripts). TLS
// / HSTS is left to the reverse proxy since wall-vault typically runs on HTTP
// behind nginx/caddy in production.
func SecurityHeaders(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		h := w.Header()
		h.Set("X-Content-Type-Options", "nosniff")
		h.Set("X-Frame-Options", "SAMEORIGIN")
		h.Set("Referrer-Policy", "strict-origin-when-cross-origin")
		// htmx + templ inject inline scripts/styles; the CDN htmx script is the
		// only third-party source. Tight enough to block XSS from untrusted
		// content without breaking the current UI.
		h.Set("Content-Security-Policy",
			"default-src 'self'; "+
				"script-src 'self' 'unsafe-inline' https://unpkg.com; "+
				"style-src 'self' 'unsafe-inline'; "+
				"img-src 'self' data:; "+
				"connect-src 'self'; "+
				"frame-ancestors 'self'; "+
				"base-uri 'self'")
		next.ServeHTTP(w, r)
	})
}

// Recovery: panic recovery middleware
func Recovery(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		defer func() {
			if err := recover(); err != nil {
				log.Printf("[panic] %v — %s %s", err, r.Method, r.URL.Path)
				http.Error(w, `{"error":"internal server error"}`, http.StatusInternalServerError)
			}
		}()
		next.ServeHTTP(w, r)
	})
}

// Chain: middleware chain (applied in reverse order)
func Chain(h http.Handler, middlewares ...func(http.Handler) http.Handler) http.Handler {
	for i := len(middlewares) - 1; i >= 0; i-- {
		h = middlewares[i](h)
	}
	return h
}
