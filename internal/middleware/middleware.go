// Package middleware: common HTTP middleware
package middleware

import (
	"log"
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
	// strip port
	if idx := strings.LastIndex(host, ":"); idx != -1 {
		host = host[:idx]
	}
	if host == "localhost" || host == "127.0.0.1" {
		return true
	}
	if strings.HasPrefix(host, "192.168.") {
		return true
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
