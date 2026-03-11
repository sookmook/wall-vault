// wall-vault: AI 프록시 + 키 금고 통합 시스템
// GitHub: https://github.com/sookmook/wall-vault
package main

import (
	"fmt"
	"log"
	"net/http"
	"os"
	"os/signal"
	"syscall"

	"github.com/sookmook/wall-vault/cmd/doctor"
	"github.com/sookmook/wall-vault/cmd/proxy"
	"github.com/sookmook/wall-vault/cmd/setup"
	"github.com/sookmook/wall-vault/cmd/vault"
	"github.com/sookmook/wall-vault/internal/config"
	"github.com/sookmook/wall-vault/internal/i18n"
	iproxy "github.com/sookmook/wall-vault/internal/proxy"
	ivault "github.com/sookmook/wall-vault/internal/vault"
)

const version = "v0.1.0"

func main() {
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
	case "doctor":
		doctor.Run(os.Args[2:])
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

// runAll: proxy + vault 동시 시작 (wall-vault start)
func runAll() {
	cfg, err := config.Load("")
	if err != nil {
		log.Fatalf("설정 오류: %v", err)
	}

	// 금고 시작
	vaultSrv, err := ivault.NewServer(cfg)
	if err != nil {
		log.Fatalf("[vault] 초기화 오류: %v", err)
	}
	vaultAddr := fmt.Sprintf("%s:%d", cfg.Vault.Host, cfg.Vault.Port)
	go func() {
		log.Printf("[vault] 시작 :%d → http://localhost:%d", cfg.Vault.Port, cfg.Vault.Port)
		if err := http.ListenAndServe(vaultAddr, vaultSrv.Handler()); err != nil {
			log.Fatalf("[vault] 오류: %v", err)
		}
	}()

	// 프록시 시작
	proxySrv := iproxy.NewServer(cfg)
	proxyAddr := fmt.Sprintf("%s:%d", cfg.Proxy.Host, cfg.Proxy.Port)
	go func() {
		log.Printf("[proxy] 시작 :%d (client=%s, filter=%s)",
			cfg.Proxy.Port, cfg.Proxy.ClientID, cfg.Proxy.ToolFilter)
		if err := http.ListenAndServe(proxyAddr, proxySrv.Handler()); err != nil {
			log.Fatalf("[proxy] 오류: %v", err)
		}
	}()

	log.Printf("wall-vault %s 실행 중", version)
	log.Printf("  대시보드: http://localhost:%d", cfg.Vault.Port)
	log.Printf("  프록시:   http://localhost:%d", cfg.Proxy.Port)
	log.Printf("Ctrl+C로 종료")

	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
	<-quit
	log.Println("종료 중...")
}

func printHelp() {
	fmt.Printf(`wall-vault %s — AI 프록시 + 키 금고

사용법:
  wall-vault setup            대화형 설치 마법사 (처음 시작)
  wall-vault start            모든 서비스 시작
  wall-vault proxy [flags]    프록시 서버 단독 실행
  wall-vault vault [flags]    키 금고 서버 단독 실행
  wall-vault doctor [cmd]     헬스체크·자동복구 (fix/deploy)

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
