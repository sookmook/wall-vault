// Package vault: 키 금고 서버 커맨드
package vault

import (
	"flag"
	"fmt"
	"log"
	"net/http"
	"os"
	"os/signal"
	"syscall"

	"github.com/sookmook/wall-vault/internal/config"
	"github.com/sookmook/wall-vault/internal/i18n"
	"github.com/sookmook/wall-vault/internal/theme"
)

// Run: wall-vault vault [flags]
func Run(args []string) {
	fs := flag.NewFlagSet("vault", flag.ExitOnError)
	cfgPath := fs.String("config", "", "설정 파일 경로")
	port := fs.Int("port", 0, "포트 (기본 56243)")

	fs.Usage = func() {
		fmt.Fprintln(os.Stderr, `wall-vault vault — 키 금고 서버

사용법:
  wall-vault vault [flags]

플래그:`)
		fs.PrintDefaults()
	}

	if err := fs.Parse(args); err != nil {
		os.Exit(1)
	}

	cfg, err := config.Load(*cfgPath)
	if err != nil {
		log.Fatalf("설정 오류: %v", err)
	}

	if *port > 0 {
		cfg.Vault.Port = *port
	}

	runVault(cfg)
}

func runVault(cfg *config.Config) {
	addr := fmt.Sprintf("%s:%d", cfg.Vault.Host, cfg.Vault.Port)
	fmt.Printf("[vault] %s :%d\n", i18n.T("vault_started"), cfg.Vault.Port)

	t := theme.Get(cfg.Theme)
	mux := http.NewServeMux()
	registerVaultRoutes(mux, cfg, t)

	srv := &http.Server{
		Addr:    addr,
		Handler: mux,
	}

	go func() {
		if err := srv.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			log.Fatalf("키 금고 서버 오류: %v", err)
		}
	}()

	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
	<-quit

	fmt.Printf("[vault] %s\n", i18n.T("stopping"))
}

func registerVaultRoutes(mux *http.ServeMux, cfg *config.Config, t *theme.Theme) {
	// 공개 경로
	mux.HandleFunc("/api/status", handleStatus())
	mux.HandleFunc("/api/events", handleSSE())    // SSE 스트림
	mux.HandleFunc("/api/clients", handleClients())

	// 관리자 경로 (Admin Token 필요)
	mux.HandleFunc("/admin/", adminMiddleware(cfg, mux))

	// 대시보드 UI
	mux.HandleFunc("/", handleDashboard(cfg, t))
}

func handleStatus() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		fmt.Fprintf(w, `{"status":"ok","version":"v0.1.0"}`)
	}
}

func handleSSE() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "text/event-stream")
		w.Header().Set("Cache-Control", "no-cache")
		w.Header().Set("Connection", "keep-alive")
		// TODO: Phase 2에서 SSE 브로드캐스터 구현
		fmt.Fprintf(w, "data: {\"type\":\"connected\"}\n\n")
	}
}

func handleClients() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		fmt.Fprintf(w, `{"clients":[]}`)
	}
}

func adminMiddleware(cfg *config.Config, mux *http.ServeMux) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		token := r.Header.Get("Authorization")
		expected := "Bearer " + cfg.Vault.AdminToken
		if cfg.Vault.AdminToken != "" && token != expected {
			http.Error(w, `{"error":"unauthorized"}`, http.StatusUnauthorized)
			return
		}
		// TODO: Phase 2 관리자 라우터
		http.Error(w, "not implemented yet", http.StatusNotImplemented)
	}
}

func handleDashboard(cfg *config.Config, t *theme.Theme) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "text/html; charset=utf-8")
		fmt.Fprintf(w, dashboardHTML(cfg, t))
	}
}

func dashboardHTML(cfg *config.Config, t *theme.Theme) string {
	return `<!DOCTYPE html>
<html lang="ko">
<head>
<meta charset="UTF-8">
<meta name="viewport" content="width=device-width, initial-scale=1.0">
<title>wall-vault 키 금고</title>
<style>
:root {` + t.CSSVars() + `}
* { box-sizing: border-box; margin: 0; padding: 0; }
body {
  background: var(--bg);
  color: var(--text);
  font-family: 'Courier New', monospace;
  padding: 2rem;
  min-height: 100vh;
}
.header {
  text-align: center;
  margin-bottom: 2rem;
  border-bottom: 1px solid var(--border);
  padding-bottom: 1rem;
}
.header h1 { color: var(--accent); font-size: 1.5rem; }
.header p { color: var(--text-muted); font-size: 0.85rem; margin-top: 0.3rem; }
.status-badge {
  display: inline-block;
  background: var(--surface);
  border: 1px solid var(--green);
  color: var(--green);
  padding: 0.2rem 0.8rem;
  border-radius: 4px;
  font-size: 0.8rem;
  margin-top: 0.5rem;
}
.card {
  background: var(--surface);
  border: 1px solid var(--border);
  border-radius: 6px;
  padding: 1.2rem;
  margin-bottom: 1rem;
}
.card h2 { color: var(--accent); font-size: 1rem; margin-bottom: 0.8rem; }
.info-row { display: flex; justify-content: space-between; margin: 0.3rem 0; font-size: 0.85rem; }
.info-label { color: var(--text-muted); }
.info-value { color: var(--text); }
.footer {
  text-align: center;
  color: var(--text-muted);
  font-size: 0.75rem;
  margin-top: 2rem;
  padding-top: 1rem;
  border-top: 1px solid var(--border);
}
</style>
</head>
<body>
<div class="header">
  <h1>🔐 wall-vault 키 금고</h1>
  <p>AI 프록시 키 관리 시스템</p>
  <span class="status-badge">● 실행 중</span>
</div>

<div class="card">
  <h2>서버 정보</h2>
  <div class="info-row"><span class="info-label">버전</span><span class="info-value">v0.1.0</span></div>
  <div class="info-row"><span class="info-label">포트</span><span class="info-value">` + fmt.Sprintf("%d", cfg.Vault.Port) + `</span></div>
  <div class="info-row"><span class="info-label">모드</span><span class="info-value">` + cfg.Mode + `</span></div>
  <div class="info-row"><span class="info-label">테마</span><span class="info-value">` + cfg.Theme + `</span></div>
</div>

<div class="card">
  <h2>⚙️ Phase 2 예정</h2>
  <div class="info-row"><span class="info-label">클라이언트 관리</span><span class="info-value" style="color:var(--text-muted)">개발 중</span></div>
  <div class="info-row"><span class="info-label">키 관리 + 바 차트</span><span class="info-value" style="color:var(--text-muted)">개발 중</span></div>
  <div class="info-row"><span class="info-label">모델 선택 UI</span><span class="info-value" style="color:var(--text-muted)">개발 중</span></div>
  <div class="info-row"><span class="info-label">SSE 실시간 동기화</span><span class="info-value" style="color:var(--text-muted)">개발 중</span></div>
</div>

<div class="footer">
  wall-vault — <a href="https://github.com/sookmook/wall-vault" style="color:var(--accent)">github.com/sookmook/wall-vault</a>
</div>

<script>
// SSE 연결
const es = new EventSource('/api/events');
es.onmessage = (e) => {
  try {
    const data = JSON.parse(e.data);
    if (data.type === 'config_change') {
      location.reload();
    }
  } catch {}
};
</script>
</body>
</html>`
}
