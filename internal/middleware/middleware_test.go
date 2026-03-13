package middleware

import (
	"net/http"
	"net/http/httptest"
	"testing"
)

func okHandler(w http.ResponseWriter, r *http.Request) {
	w.WriteHeader(http.StatusOK)
	w.Write([]byte("ok")) //nolint:errcheck
}

func panicHandler(w http.ResponseWriter, r *http.Request) {
	panic("테스트 패닉")
}

func TestCORS_Headers(t *testing.T) {
	h := CORS(http.HandlerFunc(okHandler))

	req := httptest.NewRequest("GET", "/", nil)
	w := httptest.NewRecorder()
	h.ServeHTTP(w, req)

	if w.Header().Get("Access-Control-Allow-Origin") != "*" {
		t.Errorf("ACAO = %q, want *", w.Header().Get("Access-Control-Allow-Origin"))
	}
	if w.Header().Get("Access-Control-Allow-Methods") == "" {
		t.Error("ACAM 헤더 없음")
	}
	if w.Header().Get("Access-Control-Allow-Headers") == "" {
		t.Error("ACAH 헤더 없음")
	}
}

func TestCORS_Preflight(t *testing.T) {
	h := CORS(http.HandlerFunc(okHandler))

	req := httptest.NewRequest("OPTIONS", "/api/keys", nil)
	req.Header.Set("Origin", "http://localhost:3000")
	w := httptest.NewRecorder()
	h.ServeHTTP(w, req)

	if w.Code != http.StatusNoContent {
		t.Errorf("OPTIONS status = %d, want 204", w.Code)
	}
}

func TestRecovery_Panic(t *testing.T) {
	h := Recovery(http.HandlerFunc(panicHandler))

	req := httptest.NewRequest("GET", "/", nil)
	w := httptest.NewRecorder()

	// panic must not propagate outside
	h.ServeHTTP(w, req)

	if w.Code != http.StatusInternalServerError {
		t.Errorf("panic recovery status = %d, want 500", w.Code)
	}
}

func TestRecovery_NoPanic(t *testing.T) {
	h := Recovery(http.HandlerFunc(okHandler))

	req := httptest.NewRequest("GET", "/", nil)
	w := httptest.NewRecorder()
	h.ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Errorf("status = %d, want 200", w.Code)
	}
}

func TestLogger_SkipsHealth(t *testing.T) {
	called := false
	inner := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		called = true
		w.WriteHeader(http.StatusOK)
	})
	h := Logger(inner)

	req := httptest.NewRequest("GET", "/health", nil)
	w := httptest.NewRecorder()
	h.ServeHTTP(w, req)

	if !called {
		t.Error("내부 핸들러가 호출되지 않음")
	}
	if w.Code != http.StatusOK {
		t.Errorf("status = %d", w.Code)
	}
}

func TestLogger_Normal(t *testing.T) {
	h := Logger(http.HandlerFunc(okHandler))

	req := httptest.NewRequest("GET", "/status", nil)
	w := httptest.NewRecorder()
	h.ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Errorf("status = %d", w.Code)
	}
}

func TestChain_Order(t *testing.T) {
	// middleware application order: Chain(h, A, B, C) → A(B(C(h)))
	order := []string{}

	makeMiddleware := func(name string) func(http.Handler) http.Handler {
		return func(next http.Handler) http.Handler {
			return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				order = append(order, name+"-before")
				next.ServeHTTP(w, r)
				order = append(order, name+"-after")
			})
		}
	}

	base := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		order = append(order, "handler")
		w.WriteHeader(http.StatusOK)
	})

	h := Chain(base, makeMiddleware("A"), makeMiddleware("B"), makeMiddleware("C"))

	req := httptest.NewRequest("GET", "/", nil)
	w := httptest.NewRecorder()
	h.ServeHTTP(w, req)

	// expected order: A→B→C→handler→C→B→A
	expected := []string{
		"A-before", "B-before", "C-before",
		"handler",
		"C-after", "B-after", "A-after",
	}
	for i, got := range order {
		if i >= len(expected) {
			t.Errorf("예상 외 항목: %s", got)
			continue
		}
		if got != expected[i] {
			t.Errorf("order[%d] = %q, want %q", i, got, expected[i])
		}
	}
}

func TestChain_Empty(t *testing.T) {
	h := Chain(http.HandlerFunc(okHandler))

	req := httptest.NewRequest("GET", "/", nil)
	w := httptest.NewRecorder()
	h.ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Errorf("status = %d", w.Code)
	}
}
