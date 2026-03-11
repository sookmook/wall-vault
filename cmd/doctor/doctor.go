// Package doctor: 헬스체크 및 자동 복구
package doctor

import (
	"encoding/json"
	"flag"
	"fmt"
	"io"
	"net/http"
	"os"
	"time"

	"github.com/sookmook/wall-vault/internal/config"
)

type ServiceStatus struct {
	Name    string
	Port    int
	URL     string
	Running bool
	Detail  string
}

// Run: wall-vault doctor [check|fix|status|all]
func Run(args []string) {
	fs := flag.NewFlagSet("doctor", flag.ExitOnError)
	cfgPath := fs.String("config", "", "설정 파일 경로")

	fs.Usage = func() {
		fmt.Fprintln(os.Stderr, `wall-vault doctor — 헬스체크 및 자동 복구

사용법:
  wall-vault doctor [command]

명령:
  check     상태 확인만
  fix       복구만
  status    상세 보고서 (기본)
  all       확인 후 필요시 복구

플래그:`)
		fs.PrintDefaults()
	}

	if err := fs.Parse(args); err != nil {
		os.Exit(1)
	}

	cfg, err := config.Load(*cfgPath)
	if err != nil {
		fmt.Fprintf(os.Stderr, "설정 오류: %v\n", err)
		cfg = config.Default()
	}

	cmd := "status"
	if fs.NArg() > 0 {
		cmd = fs.Arg(0)
	}

	switch cmd {
	case "check":
		runCheck(cfg)
	case "fix":
		runFix(cfg)
	case "all":
		statuses := runCheck(cfg)
		needFix := false
		for _, s := range statuses {
			if !s.Running {
				needFix = true
				break
			}
		}
		if needFix {
			fmt.Println("\n[fix] 복구 시도...")
			runFix(cfg)
		}
	default:
		printStatus(cfg)
	}
}

func runCheck(cfg *config.Config) []ServiceStatus {
	services := []ServiceStatus{
		{
			Name: "proxy",
			Port: cfg.Proxy.Port,
			URL:  fmt.Sprintf("http://localhost:%d/health", cfg.Proxy.Port),
		},
	}

	// standalone 모드에서는 vault도 체크
	if cfg.Mode == "standalone" {
		services = append(services, ServiceStatus{
			Name: "vault",
			Port: cfg.Vault.Port,
			URL:  fmt.Sprintf("http://localhost:%d/api/status", cfg.Vault.Port),
		})
	}

	client := &http.Client{Timeout: 3 * time.Second}
	for i := range services {
		s := &services[i]
		resp, err := client.Get(s.URL)
		if err != nil {
			s.Running = false
			s.Detail = err.Error()
			printLine("ERROR", s.Name, fmt.Sprintf("포트 %d 응답 없음", s.Port))
			continue
		}
		defer resp.Body.Close()

		body, _ := io.ReadAll(resp.Body)
		if resp.StatusCode == 200 {
			s.Running = true
			s.Detail = extractVersion(body)
			printLine("OK", s.Name, fmt.Sprintf("포트 %d 정상 %s", s.Port, s.Detail))
		} else {
			s.Running = false
			printLine("WARN", s.Name, fmt.Sprintf("포트 %d HTTP %d", s.Port, resp.StatusCode))
		}
	}
	return services
}

func runFix(cfg *config.Config) {
	fmt.Println("[fix] 자동 복구 기능은 Phase 2에서 구현됩니다.")
	fmt.Println("[fix] 수동 복구: wall-vault start")
}

func printStatus(cfg *config.Config) {
	now := time.Now().Format("2006-01-02 15:04:05")
	host, _ := os.Hostname()

	fmt.Printf(`
╔══════════════════════════════════════════════════════════╗
║          wall-vault 시스템 상태 보고서                    ║
╚══════════════════════════════════════════════════════════╝

📍 호스트: %s
📅 시간: %s
⚙️  모드: %s
🎨 테마: %s

`, host, now, cfg.Mode, cfg.Theme)

	runCheck(cfg)

	fmt.Printf(`
─── 설정 ───
  프록시    포트 %d (client=%s, filter=%s)
  금고      포트 %d
`, cfg.Proxy.Port, cfg.Proxy.ClientID, cfg.Proxy.ToolFilter, cfg.Vault.Port)
}

func printLine(level, name, msg string) {
	// ANSI 색상
	colors := map[string]string{
		"OK":    "\033[32m",
		"WARN":  "\033[33m",
		"ERROR": "\033[31m",
		"INFO":  "\033[34m",
	}
	reset := "\033[0m"
	color := colors[level]
	if color == "" {
		color = ""
	}
	fmt.Printf("  %-12s %s[%-5s]%s %s\n", name, color, level, reset, msg)
}

func extractVersion(body []byte) string {
	var m map[string]interface{}
	if err := json.Unmarshal(body, &m); err != nil {
		return ""
	}
	if v, ok := m["version"]; ok {
		return fmt.Sprintf("(%v)", v)
	}
	return ""
}
