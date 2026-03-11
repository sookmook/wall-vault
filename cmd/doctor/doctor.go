// Package doctor: 헬스체크 및 자동 복구 커맨드
package doctor

import (
	"flag"
	"fmt"
	"os"

	"github.com/sookmook/wall-vault/internal/config"
	idoctor "github.com/sookmook/wall-vault/internal/doctor"
)

// Run: wall-vault doctor [check|fix|status|all|deploy]
func Run(args []string) {
	fs := flag.NewFlagSet("doctor", flag.ExitOnError)
	cfgPath := fs.String("config", "", "설정 파일 경로")

	fs.Usage = func() {
		fmt.Fprintln(os.Stderr, `wall-vault doctor — 헬스체크 및 자동 복구

사용법:
  wall-vault doctor [command]

명령:
  check     상태 확인만 (기본)
  fix       자동 복구 실행
  status    상세 보고서
  all       확인 후 필요시 복구
  deploy    systemd/launchd 서비스 파일 생성

플래그:`)
		fs.PrintDefaults()
	}

	if err := fs.Parse(args); err != nil {
		os.Exit(1)
	}

	cfg, err := config.Load(*cfgPath)
	if err != nil {
		cfg = config.Default()
	}

	cmd := "check"
	if fs.NArg() > 0 {
		cmd = fs.Arg(0)
	}

	switch cmd {
	case "check":
		statuses := idoctor.Check(cfg)
		for _, s := range statuses {
			level := "OK"
			if !s.Running {
				level = "ERROR"
			}
			idoctor.Printline(level, s.Name, fmt.Sprintf(":%d %s", s.Port, s.Detail))
		}

	case "fix", "--fix":
		if err := idoctor.Fix(cfg); err != nil {
			fmt.Fprintf(os.Stderr, "[fix] 오류: %v\n", err)
			os.Exit(1)
		}

	case "status":
		idoctor.PrintStatus(cfg)

	case "all":
		statuses := idoctor.Check(cfg)
		needFix := false
		for _, s := range statuses {
			level := "OK"
			if !s.Running {
				level = "ERROR"
				needFix = true
			}
			idoctor.Printline(level, s.Name, fmt.Sprintf(":%d %s", s.Port, s.Detail))
		}
		if needFix {
			fmt.Println()
			if err := idoctor.Fix(cfg); err != nil {
				fmt.Fprintf(os.Stderr, "[fix] 오류: %v\n", err)
				os.Exit(1)
			}
		}

	case "deploy":
		deployCmd := "systemd"
		if fs.NArg() > 1 {
			deployCmd = fs.Arg(1)
		}
		switch deployCmd {
		case "launchd", "macos":
			if err := idoctor.GenerateLaunchdPlist(cfg); err != nil {
				fmt.Fprintf(os.Stderr, "[deploy] 오류: %v\n", err)
				os.Exit(1)
			}
		case "nssm", "windows":
			if err := idoctor.GenerateNSSMScript(cfg); err != nil {
				fmt.Fprintf(os.Stderr, "[deploy] 오류: %v\n", err)
				os.Exit(1)
			}
		default: // systemd (Linux/WSL)
			if err := idoctor.GenerateSystemdService(cfg); err != nil {
				fmt.Fprintf(os.Stderr, "[deploy] 오류: %v\n", err)
				os.Exit(1)
			}
		}

	default:
		fmt.Fprintf(os.Stderr, "알 수 없는 명령: %s\n", cmd)
		fs.Usage()
		os.Exit(1)
	}
}
