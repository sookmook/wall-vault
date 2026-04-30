// wall-vault: AI proxy + key vault integrated system
// GitHub: https://github.com/sookmook/wall-vault
package main

import (
	"context"
	"errors"
	"fmt"
	"log"
	"log/slog"
	"net/http"
	"os"
	"os/signal"
	"strings"
	"syscall"
	"time"

	"github.com/sookmook/wall-vault/cmd/cert"
	"github.com/sookmook/wall-vault/cmd/doctor"
	"github.com/sookmook/wall-vault/cmd/proxy"
	crtk "github.com/sookmook/wall-vault/cmd/rtk"
	"github.com/sookmook/wall-vault/cmd/setup"
	"github.com/sookmook/wall-vault/cmd/vault"
	"github.com/sookmook/wall-vault/internal/config"
	"github.com/sookmook/wall-vault/internal/i18n"
	iproxy "github.com/sookmook/wall-vault/internal/proxy"
	ivault "github.com/sookmook/wall-vault/internal/vault"
)

var version = "dev" // overridden at build time via -ldflags "-X main.version=..."

// initSlog installs a console-friendly slog TextHandler as the default.
// Level is controlled via WV_LOG_LEVEL (debug|info|warn|error) and falls back
// to info. The classic `log` package keeps working in parallel — existing
// log.Printf calls still write with their own prefixes — while newly added
// code can use slog.Info / slog.Warn / slog.Error with structured key=value
// attributes for better grep/parse.
func initSlog() {
	level := slog.LevelInfo
	switch strings.ToLower(os.Getenv("WV_LOG_LEVEL")) {
	case "debug":
		level = slog.LevelDebug
	case "warn", "warning":
		level = slog.LevelWarn
	case "error":
		level = slog.LevelError
	}
	h := slog.NewTextHandler(os.Stderr, &slog.HandlerOptions{Level: level})
	slog.SetDefault(slog.New(h))
}

func main() {
	// propagate build-injected version to all packages
	iproxy.Version = version
	ivault.Version = version

	initSlog()
	i18n.Init()

	if len(os.Args) < 2 {
		printHelp()
		os.Exit(0)
	}

	switch os.Args[1] {
	case "start":
		runAll()
	case "proxy":
		proxy.Run(os.Args[2:])
	case "vault":
		vault.Run(os.Args[2:])
	case "setup":
		setup.Run(os.Args[2:])
	case "cert":
		cert.Run(os.Args[2:])
	case "doctor":
		doctor.Run(os.Args[2:])
	case "rtk":
		crtk.Run(os.Args[2:])
	case "version", "--version", "-v":
		fmt.Printf("wall-vault %s\n", version)
	case "help", "--help", "-h":
		printHelp()
	default:
		fmt.Fprintf(os.Stderr, i18n.T("unknown_command")+": %s\n", os.Args[1])
		printHelp()
		os.Exit(1)
	}
}

// runAll: start proxy + vault simultaneously (wall-vault start)
func runAll() {
	cfg, err := config.Load("")
	if err != nil {
		log.Fatalf("설정 오류: %v", err)
	}

	// start vault
	vaultSrv, err := ivault.NewServer(cfg)
	if err != nil {
		log.Fatalf("[vault] 초기화 오류: %v", err)
	}
	vaultSrv.SetConfigPath("wall-vault.yaml")
	vaultAddr := fmt.Sprintf("%s:%d", cfg.Vault.Host, cfg.Vault.Port)
	vaultHTTP := &http.Server{Addr: vaultAddr, Handler: vaultSrv.Handler()}
	vaultScheme := "http"
	if cfg.Vault.TLS.Enabled {
		vaultScheme = "https"
	}
	go func() {
		log.Printf("[vault] 시작 :%d → %s://localhost:%d", cfg.Vault.Port, vaultScheme, cfg.Vault.Port)
		err := serveHTTP(vaultHTTP, cfg.Vault.TLS)
		if err != nil && !errors.Is(err, http.ErrServerClosed) {
			log.Fatalf("[vault] 오류: %v", err)
		}
	}()

	// start proxy
	proxySrv := iproxy.NewServer(cfg)
	proxyAddr := fmt.Sprintf("%s:%d", cfg.Proxy.Host, cfg.Proxy.Port)
	proxyHTTP := &http.Server{Addr: proxyAddr, Handler: proxySrv.Handler()}
	proxyScheme := "http"
	if cfg.Proxy.TLS.Enabled {
		proxyScheme = "https"
	}
	go func() {
		log.Printf("[proxy] 시작 :%d (client=%s, filter=%s, scheme=%s)",
			cfg.Proxy.Port, cfg.Proxy.ClientID, cfg.Proxy.ToolFilter, proxyScheme)
		err := serveHTTP(proxyHTTP, cfg.Proxy.TLS)
		if err != nil && !errors.Is(err, http.ErrServerClosed) {
			log.Fatalf("[proxy] 오류: %v", err)
		}
	}()

	slog.Info("wall-vault running",
		"version", version,
		"dashboard", fmt.Sprintf("%s://localhost:%d", vaultScheme, cfg.Vault.Port),
		"proxy", fmt.Sprintf("%s://localhost:%d", proxyScheme, cfg.Proxy.Port),
	)
	log.Printf("Ctrl+C로 종료")

	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
	<-quit
	slog.Info("wall-vault shutting down")

	// Drain in-flight HTTP requests, then tear down each server's background
	// goroutines. 10s is enough for streaming responses to finish their current
	// chunk while still keeping systemd stop bounded.
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()
	if err := proxyHTTP.Shutdown(ctx); err != nil {
		slog.Warn("proxy http shutdown", "err", err)
	}
	if err := vaultHTTP.Shutdown(ctx); err != nil {
		slog.Warn("vault http shutdown", "err", err)
	}
	proxySrv.Stop()
	vaultSrv.Stop()
}

// serveHTTP starts an HTTP listener, switching to TLS when the per-listener
// config says so. CertFile/KeyFile must already exist; provision them with
// `wall-vault cert init` + `wall-vault cert issue <hostname>`.
func serveHTTP(srv *http.Server, tls config.TLSConfig) error {
	if tls.Enabled {
		if tls.CertFile == "" || tls.KeyFile == "" {
			return fmt.Errorf("tls.enabled=true requires tls.cert_file and tls.key_file")
		}
		return srv.ListenAndServeTLS(tls.CertFile, tls.KeyFile)
	}
	return srv.ListenAndServe()
}

func printHelp() {
	fmt.Printf(`wall-vault %s — AI 프록시 + 키 금고

사용법:
  wall-vault setup            대화형 설치 마법사 (처음 시작)
  wall-vault start            모든 서비스 시작
  wall-vault proxy [flags]    프록시 서버 단독 실행
  wall-vault vault [flags]    키 금고 서버 단독 실행
  wall-vault doctor [cmd]     헬스체크·자동복구 (fix/deploy)
  wall-vault cert <cmd>       내부 CA·호스트 인증서 (init/issue/list)
  wall-vault rtk <cmd> [args]  명령 출력 축소 (토큰 절약)

옵션:
  -h, --help                  도움말
  -v, --version               버전 정보

설정 파일:
  ./wall-vault.yaml           프로젝트 루트 설정
  ~/.wall-vault/config.yaml   사용자 설정

더 보기:
  wall-vault proxy --help
  wall-vault vault --help
  wall-vault doctor --help
`, version)
}
