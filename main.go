// wall-vault: AI 프록시 + 키 금고 통합 시스템
// GitHub: https://github.com/sookmook/wall-vault
package main

import (
	"fmt"
	"os"

	"github.com/sookmook/wall-vault/cmd/doctor"
	"github.com/sookmook/wall-vault/cmd/proxy"
	"github.com/sookmook/wall-vault/cmd/vault"
	"github.com/sookmook/wall-vault/internal/i18n"
)

const version = "v0.1.0"

func main() {
	// 언어 초기화 (환경변수 LANG 또는 WV_LANG 우선)
	i18n.Init()

	if len(os.Args) < 2 {
		printHelp()
		os.Exit(0)
	}

	switch os.Args[1] {
	case "start":
		// 초보자용: proxy + vault 동시 시작 (standalone 모드)
		proxy.RunStandalone()
	case "proxy":
		proxy.Run(os.Args[2:])
	case "vault":
		vault.Run(os.Args[2:])
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

func printHelp() {
	fmt.Printf(`wall-vault %s — AI 프록시 + 키 금고

사용법:
  wall-vault start            모든 서비스 시작 (초보자용)
  wall-vault proxy [flags]    프록시 서버 단독 실행
  wall-vault vault [flags]    키 금고 서버 단독 실행
  wall-vault doctor [cmd]     헬스체크 및 자동 복구

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
