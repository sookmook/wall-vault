// Package doctor: health check and auto-recovery command
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
  check           상태 확인만 (기본)
  fix             자동 복구 실행 (프로세스 + proxy_enabled 수정)
  fix-services    vault 서비스 proxy_enabled 자동 활성화
  fix-nanoclaw    ~/nanoclaw/.env ANTHROPIC_BASE_URL 수정
  status          상세 보고서
  all             확인 후 필요시 복구
  deploy          systemd/launchd 서비스 파일 생성

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
		// also check nanoclaw env if present
		if ok, detail := idoctor.CheckNanoclawEnv(cfg.Proxy.Port); !ok {
			idoctor.Printline("WARN", "nanoclaw-env", detail)
		} else {
			idoctor.Printline("OK", "nanoclaw-env", detail)
		}

	case "fix", "--fix":
		if err := idoctor.Fix(cfg); err != nil {
			fmt.Fprintf(os.Stderr, "[fix] 오류: %v\n", err)
			os.Exit(1)
		}
		// also fix vault services (no-op if no admin token)
		if err := idoctor.FixVaultServices(cfg); err != nil {
			fmt.Fprintf(os.Stderr, "[fix-services] 오류: %v\n", err)
		}

	case "fix-services":
		if err := idoctor.FixVaultServices(cfg); err != nil {
			fmt.Fprintf(os.Stderr, "[fix-services] 오류: %v\n", err)
			os.Exit(1)
		}

	case "fix-nanoclaw":
		if err := idoctor.FixNanoclawEnv(cfg.Proxy.Port, cfg.Proxy.VaultToken); err != nil {
			fmt.Fprintf(os.Stderr, "[fix-nanoclaw] 오류: %v\n", err)
			os.Exit(1)
		}
		fmt.Printf("[fix-nanoclaw] ✓ ANTHROPIC_BASE_URL=http://localhost:%d 설정 완료\n", cfg.Proxy.Port)

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
		if ok, detail := idoctor.CheckNanoclawEnv(cfg.Proxy.Port); !ok {
			idoctor.Printline("WARN", "nanoclaw-env", detail)
			fmt.Println()
			_ = idoctor.FixNanoclawEnv(cfg.Proxy.Port, cfg.Proxy.VaultToken)
		} else {
			idoctor.Printline("OK", "nanoclaw-env", detail)
		}
		if needFix {
			fmt.Println()
			if err := idoctor.Fix(cfg); err != nil {
				fmt.Fprintf(os.Stderr, "[fix] 오류: %v\n", err)
				os.Exit(1)
			}
		}
		// always try to fix services (no-op when all is well)
		_ = idoctor.FixVaultServices(cfg)

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
