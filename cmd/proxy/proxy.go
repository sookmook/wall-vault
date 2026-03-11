// Package proxy: 프록시 서버 커맨드
package proxy

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
)

// Run: wall-vault proxy [flags]
func Run(args []string) {
	fs := flag.NewFlagSet("proxy", flag.ExitOnError)
	cfgPath := fs.String("config", "", "설정 파일 경로")
	port := fs.Int("port", 0, "포트 (기본 56244)")
	clientID := fs.String("id", "", "클라이언트 ID")

	fs.Usage = func() {
		fmt.Fprintln(os.Stderr, `wall-vault proxy — AI API 프록시 서버

사용법:
  wall-vault proxy [flags]

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

	// 플래그 덮어쓰기
	if *port > 0 {
		cfg.Proxy.Port = *port
	}
	if *clientID != "" {
		cfg.Proxy.ClientID = *clientID
	}

	runProxy(cfg)
}

// RunStandalone: wall-vault start (proxy 부분)
func RunStandalone() {
	cfg, err := config.Load("")
	if err != nil {
		log.Fatalf("설정 오류: %v", err)
	}
	runProxy(cfg)
}

func runProxy(cfg *config.Config) {
	addr := fmt.Sprintf("%s:%d", cfg.Proxy.Host, cfg.Proxy.Port)
	fmt.Printf("[proxy] %s :%d (client=%s, filter=%s)\n",
		i18n.T("proxy_started"), cfg.Proxy.Port, cfg.Proxy.ClientID, cfg.Proxy.ToolFilter)

	mux := http.NewServeMux()
	registerRoutes(mux, cfg)

	srv := &http.Server{
		Addr:    addr,
		Handler: mux,
	}

	go func() {
		if err := srv.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			log.Fatalf("프록시 서버 오류: %v", err)
		}
	}()

	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
	<-quit

	fmt.Printf("[proxy] %s\n", i18n.T("stopping"))
}

func registerRoutes(mux *http.ServeMux, cfg *config.Config) {
	// TODO: 핸들러 구현 (Phase 2)
	mux.HandleFunc("/health", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		fmt.Fprintf(w, `{"status":"ok","version":"v0.1.0","client":"%s"}`, cfg.Proxy.ClientID)
	})

	mux.HandleFunc("/status", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		fmt.Fprintf(w, `{"status":"ok","port":%d,"client":"%s","filter":"%s"}`,
			cfg.Proxy.Port, cfg.Proxy.ClientID, cfg.Proxy.ToolFilter)
	})

	// Gemini API 호환 엔드포인트
	mux.HandleFunc("/google/", handleGemini(cfg))

	// OpenAI 호환 엔드포인트
	mux.HandleFunc("/v1/chat/completions", handleOpenAI(cfg))

	// 모델 목록
	mux.HandleFunc("/api/models", handleModels(cfg))

	// 설정 변경
	mux.HandleFunc("/api/config/model", handleConfigModel(cfg))
}

// 핸들러 스텁 — Phase 2에서 구현
func handleGemini(cfg *config.Config) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		http.Error(w, "not implemented yet", http.StatusNotImplemented)
	}
}

func handleOpenAI(cfg *config.Config) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		http.Error(w, "not implemented yet", http.StatusNotImplemented)
	}
}

func handleModels(cfg *config.Config) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		fmt.Fprintf(w, `{"models":[],"services":%q}`, cfg.Proxy.Services)
	}
}

func handleConfigModel(cfg *config.Config) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		http.Error(w, "not implemented yet", http.StatusNotImplemented)
	}
}
