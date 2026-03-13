// Package doctor: health check and auto-recovery logic
package doctor

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"os"
	"os/exec"
	"path/filepath"
	"runtime"
	"strings"
	"time"

	"github.com/sookmook/wall-vault/internal/config"
)

// ─── Status Structure ─────────────────────────────────────────────────────────

type ServiceStatus struct {
	Name    string
	Port    int
	URL     string
	Running bool
	Version string
	Detail  string
}

// ─── Status Check ─────────────────────────────────────────────────────────────

func Check(cfg *config.Config) []ServiceStatus {
	services := []ServiceStatus{
		{Name: "proxy", Port: cfg.Proxy.Port, URL: fmt.Sprintf("http://localhost:%d/health", cfg.Proxy.Port)},
	}
	if cfg.Mode == "standalone" {
		services = append(services, ServiceStatus{
			Name: "vault", Port: cfg.Vault.Port, URL: fmt.Sprintf("http://localhost:%d/api/status", cfg.Vault.Port),
		})
	}

	client := &http.Client{Timeout: 3 * time.Second}
	for i := range services {
		s := &services[i]
		resp, err := client.Get(s.URL)
		if err != nil {
			s.Running = false
			s.Detail = "응답 없음"
			continue
		}
		defer resp.Body.Close()
		body, _ := io.ReadAll(resp.Body)
		s.Running = resp.StatusCode == http.StatusOK
		s.Version = extractVersion(body)
		if s.Running {
			s.Detail = "정상"
		} else {
			s.Detail = fmt.Sprintf("HTTP %d", resp.StatusCode)
		}
	}
	return services
}

// ─── Auto-Recovery ────────────────────────────────────────────────────────────

func Fix(cfg *config.Config) error {
	statuses := Check(cfg)

	allOK := true
	for _, s := range statuses {
		if !s.Running {
			allOK = false
		}
	}
	if allOK {
		fmt.Println("[fix] 모든 서비스 정상 — 복구 불필요")
		return nil
	}

	// 1st priority: systemd (Linux/WSL)
	if isSystemd() {
		return fixSystemd(cfg)
	}

	// 2nd priority: launchd (macOS)
	if runtime.GOOS == "darwin" {
		return fixLaunchd()
	}

	// 3rd priority: NSSM (Windows native)
	if runtime.GOOS == "windows" {
		if err := fixNSSM(); err == nil {
			return nil
		}
	}

	// 4th priority: start process directly
	return fixDirect(cfg)
}

// ─── systemd Recovery ─────────────────────────────────────────────────────────

func isSystemd() bool {
	_, err := exec.LookPath("systemctl")
	if err != nil {
		return false
	}
	// check service file existence
	home, _ := os.UserHomeDir()
	svcFile := filepath.Join(home, ".config", "systemd", "user", "wall-vault.service")
	_, err = os.Stat(svcFile)
	return err == nil
}

func fixSystemd(cfg *config.Config) error {
	home, _ := os.UserHomeDir()
	svcFile := filepath.Join(home, ".config", "systemd", "user", "wall-vault.service")

	fmt.Printf("[fix] systemd 서비스 재시작: %s\n", svcFile)
	cmd := exec.Command("systemctl", "--user", "restart", "wall-vault")
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	if err := cmd.Run(); err != nil {
		return fmt.Errorf("systemctl restart 실패: %w", err)
	}
	return nil
}

// ─── launchd Recovery (macOS) ────────────────────────────────────────────────

func fixLaunchd() error {
	home, _ := os.UserHomeDir()
	plist := filepath.Join(home, "Library", "LaunchAgents", "com.wall-vault.plist")
	if _, err := os.Stat(plist); err != nil {
		return fmt.Errorf("launchd plist 없음: %s", plist)
	}
	fmt.Printf("[fix] launchd 재시작: %s\n", plist)
	exec.Command("launchctl", "unload", plist).Run()  //nolint:errcheck
	cmd := exec.Command("launchctl", "load", plist)
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	return cmd.Run()
}

// ─── NSSM Recovery (Windows) ──────────────────────────────────────────────────

func fixNSSM() error {
	if _, err := exec.LookPath("nssm"); err != nil {
		return fmt.Errorf("nssm 없음")
	}
	fmt.Println("[fix] NSSM 서비스 재시작: wall-vault")
	cmd := exec.Command("nssm", "restart", "wall-vault")
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	return cmd.Run()
}

// ─── Direct Process Start ─────────────────────────────────────────────────────

func fixDirect(cfg *config.Config) error {
	// locate wall-vault binary
	bin := findBinary()
	if bin == "" {
		return fmt.Errorf("wall-vault 바이너리를 찾을 수 없음 — PATH 또는 ~/.local/bin 확인")
	}

	fmt.Printf("[fix] 프로세스 시작: %s start\n", bin)
	cmd := exec.Command(bin, "start")
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	if err := cmd.Start(); err != nil {
		return fmt.Errorf("프로세스 시작 실패: %w", err)
	}
	fmt.Printf("[fix] PID %d 로 시작됨\n", cmd.Process.Pid)
	return nil
}

func findBinary() string {
	// 1. current executable
	if exe, err := os.Executable(); err == nil {
		return exe
	}
	// 2. platform-specific candidate paths
	home, _ := os.UserHomeDir()
	candidates := []string{
		// Linux / WSL
		filepath.Join(home, ".local", "bin", "wall-vault"),
		filepath.Join(home, "go", "bin", "wall-vault"),
		"/usr/local/bin/wall-vault",
		// macOS (Homebrew prefix)

		"/opt/homebrew/bin/wall-vault",
		"/usr/local/bin/wall-vault",
	}
	// Windows
	if runtime.GOOS == "windows" {
		candidates = append(candidates,
			filepath.Join(home, "AppData", "Local", "Programs", "wall-vault", "wall-vault.exe"),
			`C:\Program Files\wall-vault\wall-vault.exe`,
		)
	}
	for _, c := range candidates {
		if _, err := os.Stat(c); err == nil {
			return c
		}
	}
	// 3. PATH
	bin := "wall-vault"
	if runtime.GOOS == "windows" {
		bin = "wall-vault.exe"
	}
	if p, err := exec.LookPath(bin); err == nil {
		return p
	}
	return ""
}

// ─── Service File Generation ──────────────────────────────────────────────────

// GenerateSystemdService: generate ~/.config/systemd/user/wall-vault.service
func GenerateSystemdService(cfg *config.Config) error {
	bin := findBinary()
	if bin == "" {
		return fmt.Errorf("바이너리를 찾을 수 없음 — 먼저 make install 실행")
	}
	home, _ := os.UserHomeDir()
	dir := filepath.Join(home, ".config", "systemd", "user")
	if err := os.MkdirAll(dir, 0755); err != nil {
		return err
	}
	path := filepath.Join(dir, "wall-vault.service")

	content := fmt.Sprintf(`[Unit]
Description=wall-vault AI Proxy + Key Vault
After=network.target

[Service]
Type=simple
ExecStart=%s start
Restart=on-failure
RestartSec=5s
Environment=WV_LANG=ko
WorkingDirectory=%s

[Install]
WantedBy=default.target
`, bin, home)

	if err := os.WriteFile(path, []byte(content), 0644); err != nil {
		return err
	}
	fmt.Printf("[deploy] systemd 서비스 파일 생성: %s\n", path)
	fmt.Println("[deploy] 다음 명령으로 등록:")
	fmt.Println("  systemctl --user daemon-reload")
	fmt.Println("  systemctl --user enable --now wall-vault")
	return nil
}

// GenerateLaunchdPlist: generate ~/Library/LaunchAgents/com.wall-vault.plist (macOS)
func GenerateLaunchdPlist(cfg *config.Config) error {
	bin := findBinary()
	if bin == "" {
		return fmt.Errorf("바이너리를 찾을 수 없음")
	}
	home, _ := os.UserHomeDir()
	dir := filepath.Join(home, "Library", "LaunchAgents")
	if err := os.MkdirAll(dir, 0755); err != nil {
		return err
	}
	path := filepath.Join(dir, "com.wall-vault.plist")

	content := fmt.Sprintf(`<?xml version="1.0" encoding="UTF-8"?>
<!DOCTYPE plist PUBLIC "-//Apple//DTD PLIST 1.0//EN"
  "http://www.apple.com/DTDs/PropertyList-1.0.dtd">
<plist version="1.0">
<dict>
  <key>Label</key>
  <string>com.wall-vault</string>
  <key>ProgramArguments</key>
  <array>
    <string>%s</string>
    <string>start</string>
  </array>
  <key>RunAtLoad</key>
  <true/>
  <key>KeepAlive</key>
  <true/>
  <key>WorkingDirectory</key>
  <string>%s</string>
  <key>EnvironmentVariables</key>
  <dict>
    <key>WV_LANG</key>
    <string>ko</string>
  </dict>
  <key>StandardOutPath</key>
  <string>/tmp/wall-vault.log</string>
  <key>StandardErrorPath</key>
  <string>/tmp/wall-vault.log</string>
</dict>
</plist>
`, bin, home)

	if err := os.WriteFile(path, []byte(content), 0644); err != nil {
		return err
	}
	fmt.Printf("[deploy] launchd plist 생성: %s\n", path)
	fmt.Println("[deploy] 다음 명령으로 등록:")
	fmt.Printf("  launchctl load %s\n", path)
	return nil
}

// GenerateNSSMScript: generate Windows NSSM service registration script
func GenerateNSSMScript(cfg *config.Config) error {
	bin := findBinary()
	if bin == "" {
		bin = `C:\Program Files\wall-vault\wall-vault.exe`
	}
	home, _ := os.UserHomeDir()
	path := filepath.Join(home, "install-wall-vault-service.bat")

	content := fmt.Sprintf(`@echo off
REM wall-vault Windows 서비스 설치 스크립트 (NSSM 필요)
REM https://nssm.cc 에서 다운로드 후 PATH에 추가

SET BIN=%s
SET SVC=wall-vault

nssm install %s "%s" start
nssm set %s AppDirectory "%s"
nssm set %s AppEnvironmentExtra WV_LANG=ko
nssm set %s Start SERVICE_AUTO_START
nssm start %s

echo.
echo 서비스 등록 완료: %s
echo 관리: nssm start/stop/restart %s
`, bin, "%SVC%", "%BIN%", "%SVC%", home, "%SVC%", "%SVC%", "%SVC%", "%SVC%", "%SVC%")

	if err := os.WriteFile(path, []byte(content), 0644); err != nil {
		return err
	}
	fmt.Printf("[deploy] Windows 서비스 스크립트 생성: %s\n", path)
	fmt.Println("[deploy] 관리자 권한으로 실행하세요:")
	fmt.Printf("  %s\n", path)
	return nil
}

// ─── Status Report Output ─────────────────────────────────────────────────────

func PrintStatus(cfg *config.Config) {
	now := time.Now().Format("2006-01-02 15:04:05")
	host, _ := os.Hostname()

	fmt.Printf(`
╔══════════════════════════════════════════════════════╗
║       wall-vault 시스템 상태 보고서                   ║
╚══════════════════════════════════════════════════════╝

📍 호스트: %s
📅 시간:   %s
⚙️  모드:   %s
🎨 테마:   %s

`, host, now, cfg.Mode, cfg.Theme)

	statuses := Check(cfg)
	fmt.Println("─── 서비스 상태 ───")
	for _, s := range statuses {
		icon := "●"
		color := "\033[32m" // 녹색
		if !s.Running {
			icon = "○"
			color = "\033[31m" // 빨간색
		}
		fmt.Printf("  %-8s %s%s\033[0m :%d %s %s\n",
			s.Name, color, icon, s.Port, s.Version, s.Detail)
	}

	fmt.Printf(`
─── 설정 ───
  프록시 포트:  %d (client=%s, filter=%s)
  금고 포트:    %d
  서비스:       %s
`, cfg.Proxy.Port, cfg.Proxy.ClientID, cfg.Proxy.ToolFilter,
		cfg.Vault.Port, strings.Join(cfg.Proxy.Services, ", "))
}

// ─── Util ─────────────────────────────────────────────────────────────────────

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

func Printline(level, name, msg string) {
	colors := map[string]string{
		"OK":    "\033[32m",
		"WARN":  "\033[33m",
		"ERROR": "\033[31m",
		"INFO":  "\033[34m",
	}
	reset := "\033[0m"
	color := colors[level]
	fmt.Printf("  %-12s %s[%-5s]%s %s\n", name, color, level, reset, msg)
}
