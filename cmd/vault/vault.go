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
	ivault "github.com/sookmook/wall-vault/internal/vault"
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

	runVault(cfg, *cfgPath)
}

func runVault(cfg *config.Config, cfgPath string) {
	srv, err := ivault.NewServer(cfg)
	if err != nil {
		log.Fatalf("[vault] 초기화 오류: %v", err)
	}
	if cfgPath != "" {
		srv.SetConfigPath(cfgPath)
	} else {
		srv.SetConfigPath("wall-vault.yaml")
	}

	addr := fmt.Sprintf("%s:%d", cfg.Vault.Host, cfg.Vault.Port)
	log.Printf("[vault] 시작 :%d (theme=%s, mode=%s)", cfg.Vault.Port, cfg.Theme, cfg.Mode)

	httpSrv := &http.Server{
		Addr:    addr,
		Handler: srv.Handler(),
	}

	go func() {
		if err := httpSrv.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			log.Fatalf("[vault] 서버 오류: %v", err)
		}
	}()

	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
	<-quit
	log.Println("[vault] 종료 중...")
}
