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
	iproxy "github.com/sookmook/wall-vault/internal/proxy"
)

// Run: wall-vault proxy [flags]
func Run(args []string) {
	fs := flag.NewFlagSet("proxy", flag.ExitOnError)
	cfgPath    := fs.String("config",       "",  "설정 파일 경로")
	port       := fs.Int("port",            0,   "포트 (기본 56244)")
	clientID   := fs.String("id",           "",  "클라이언트 ID")
	keyGoogle  := fs.String("key-google",   "",  "Google API 키 (env: WV_KEY_GOOGLE)")
	keyOR      := fs.String("key-openrouter","", "OpenRouter API 키 (env: WV_KEY_OPENROUTER)")
	vaultURL   := fs.String("vault",        "",  "금고 서버 URL (env: WV_VAULT_URL)")
	vaultToken := fs.String("vault-token",  "",  "금고 인증 토큰 (env: WV_VAULT_TOKEN)")
	filter     := fs.String("filter",       "",  "도구 필터: strip_all|whitelist|passthrough")

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

	// flag override (priority: flag > env var > config file)
	if *port > 0 {
		cfg.Proxy.Port = *port
	}
	if *clientID != "" {
		cfg.Proxy.ClientID = *clientID
	}
	if *vaultURL != "" {
		cfg.Proxy.VaultURL = *vaultURL
	}
	if *vaultToken != "" {
		cfg.Proxy.VaultToken = *vaultToken
	}
	if *filter != "" {
		cfg.Proxy.ToolFilter = *filter
	}

	// API keys: flag → env var order of precedence
	if v := *keyGoogle; v == "" {
		v = os.Getenv("WV_KEY_GOOGLE")
		_ = v
	} else {
		os.Setenv("WV_KEY_GOOGLE", v)
	}
	if v := *keyOR; v == "" {
		v = os.Getenv("WV_KEY_OPENROUTER")
		_ = v
	} else {
		os.Setenv("WV_KEY_OPENROUTER", v)
	}

	runProxy(cfg)
}

// RunStandalone: wall-vault start (proxy portion)
func RunStandalone() {
	cfg, err := config.Load("")
	if err != nil {
		log.Fatalf("설정 오류: %v", err)
	}
	runProxy(cfg)
}

func runProxy(cfg *config.Config) {
	srv := iproxy.NewServer(cfg)
	addr := fmt.Sprintf("%s:%d", cfg.Proxy.Host, cfg.Proxy.Port)

	log.Printf("[proxy] 시작 :%d (client=%s, filter=%s)",
		cfg.Proxy.Port, cfg.Proxy.ClientID, cfg.Proxy.ToolFilter)

	httpSrv := &http.Server{
		Addr:    addr,
		Handler: srv.Handler(),
	}

	go func() {
		var err error
		if cfg.Proxy.TLS.Enabled {
			if cfg.Proxy.TLS.CertFile == "" || cfg.Proxy.TLS.KeyFile == "" {
				log.Fatalf("[proxy] tls.enabled=true requires tls.cert_file and tls.key_file")
			}
			err = httpSrv.ListenAndServeTLS(cfg.Proxy.TLS.CertFile, cfg.Proxy.TLS.KeyFile)
		} else {
			err = httpSrv.ListenAndServe()
		}
		if err != nil && err != http.ErrServerClosed {
			log.Fatalf("[proxy] 서버 오류: %v", err)
		}
	}()

	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
	<-quit
	log.Println("[proxy] 종료 중...")
}
