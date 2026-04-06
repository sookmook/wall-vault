// Package rtk implements the "wall-vault rtk" subcommand for token reduction.
package rtk

import (
	"fmt"
	"os"
	"strings"

	irtk "github.com/sookmook/wall-vault/internal/rtk"
)

// Run is the entry point for the rtk subcommand.
func Run(args []string) {
	if len(args) == 0 {
		printUsage()
		os.Exit(0)
	}

	switch args[0] {
	case "rewrite":
		runRewrite(args[1:])
	case "help", "--help", "-h":
		printUsage()
	default:
		// default mode: execute + filter
		exitCode := irtk.Run(args)
		os.Exit(exitCode)
	}
}

// runRewrite outputs the rewritten command for hook integration.
// Input: the original command as a single string or separate args.
// Output: "wall-vault rtk <command>" to stdout.
func runRewrite(args []string) {
	if len(args) == 0 {
		os.Exit(1)
	}
	// join args into the original command string
	origCmd := strings.Join(args, " ")
	// output the rewritten command
	exe, _ := os.Executable()
	if exe == "" {
		exe = "wall-vault"
	}
	fmt.Printf("%s rtk %s\n", exe, origCmd)
}

func printUsage() {
	fmt.Println(`wall-vault rtk — 명령 출력 축소 (토큰 절약)

사용법:
  wall-vault rtk <command> [args...]   명령 실행 후 출력 필터링
  wall-vault rtk rewrite "<command>"   리라이트된 명령 출력 (훅용)

예시:
  wall-vault rtk git status            git status 축소 출력
  wall-vault rtk go test ./...         실패한 테스트만 표시
  wall-vault rtk git diff HEAD~1       diff 컨텍스트 축소

지원 명령:
  git     status, diff, log, push, pull, fetch, branch
  go      test, build, vet
  기타    패스스루 + 자동 절삭 (50줄 앞/뒤)`)
}
