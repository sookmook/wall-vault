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
	if err := fixDirect(cfg); err != nil {
		return err
	}

	// after process recovery, also fix vault service proxy_enabled flags
	time.Sleep(2 * time.Second) // give the vault a moment to start
	_ = FixVaultServices(cfg)
	return nil
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

// ─── Vault Service Auto-fix ───────────────────────────────────────────────────

// FixVaultServices: for each enabled service that has keys but proxy_enabled=false,
// automatically enables proxy_enabled via the vault admin API.
// Only runs if cfg.Vault.AdminToken is set (vault machine).
func FixVaultServices(cfg *config.Config) error {
	if cfg.Vault.AdminToken == "" {
		fmt.Println("[fix-services] 관리자 토큰 없음 — 건너뜀")
		return nil
	}
	vaultBase := fmt.Sprintf("http://localhost:%d", cfg.Vault.Port)
	client := &http.Client{Timeout: 5 * time.Second}

	svcRaw, err := vaultGetJSON(client, vaultBase+"/admin/services", cfg.Vault.AdminToken)
	if err != nil {
		return fmt.Errorf("서비스 목록 조회 실패: %w", err)
	}
	keyRaw, err := vaultGetJSON(client, vaultBase+"/admin/keys", cfg.Vault.AdminToken)
	if err != nil {
		return fmt.Errorf("키 목록 조회 실패: %w", err)
	}

	// count keys per service
	keyCounts := map[string]int{}
	if keyArr, ok := keyRaw.([]interface{}); ok {
		for _, k := range keyArr {
			if km, ok := k.(map[string]interface{}); ok {
				if svc, ok := km["service"].(string); ok {
					keyCounts[svc]++
				}
			}
		}
	}

	fixed := 0
	svcArr, _ := svcRaw.([]interface{})
	for _, s := range svcArr {
		sm, ok := s.(map[string]interface{})
		if !ok {
			continue
		}
		id, _ := sm["id"].(string)
		enabled, _ := sm["enabled"].(bool)
		proxyEnabled, _ := sm["proxy_enabled"].(bool)

		if enabled && keyCounts[id] > 0 && !proxyEnabled {
			body := strings.NewReader(`{"proxy_enabled":true}`)
			req, _ := http.NewRequest(http.MethodPut, vaultBase+"/admin/services/"+id, body)
			req.Header.Set("Authorization", "Bearer "+cfg.Vault.AdminToken)
			req.Header.Set("Content-Type", "application/json")
			resp, err := client.Do(req)
			if err != nil || resp.StatusCode != http.StatusOK {
				fmt.Printf("[fix-services] ✗ %s proxy_enabled 활성화 실패\n", id)
				if resp != nil {
					resp.Body.Close()
				}
				continue
			}
			resp.Body.Close()
			fmt.Printf("[fix-services] ✓ %s proxy_enabled 활성화\n", id)
			fixed++
		}
	}
	if fixed == 0 {
		fmt.Println("[fix-services] 수정할 서비스 없음")
	} else {
		fmt.Printf("[fix-services] %d개 서비스 proxy_enabled 활성화 완료\n", fixed)
	}
	return nil
}

func vaultGetJSON(client *http.Client, url, token string) (interface{}, error) {
	req, err := http.NewRequest(http.MethodGet, url, nil)
	if err != nil {
		return nil, err
	}
	req.Header.Set("Authorization", "Bearer "+token)
	resp, err := client.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()
	body, _ := io.ReadAll(resp.Body)
	var v interface{}
	if err := json.Unmarshal(body, &v); err != nil {
		return nil, err
	}
	return v, nil
}

// ─── NanoClaw env check / fix ────────────────────────────────────────────────

// CheckNanoclawEnv: verifies ~/nanoclaw/.env has ANTHROPIC_BASE_URL pointing to
// the local wall-vault proxy.
func CheckNanoclawEnv(proxyPort int) (ok bool, detail string) {
	home, _ := os.UserHomeDir()
	envFile := filepath.Join(home, "nanoclaw", ".env")
	data, err := os.ReadFile(envFile)
	if err != nil {
		return false, "~/nanoclaw/.env 없음"
	}
	expected := fmt.Sprintf("ANTHROPIC_BASE_URL=http://localhost:%d", proxyPort)
	content := string(data)
	if strings.Contains(content, expected) {
		return true, "ANTHROPIC_BASE_URL 정상"
	}
	if strings.Contains(content, "ANTHROPIC_BASE_URL=") {
		return false, fmt.Sprintf("ANTHROPIC_BASE_URL가 localhost:%d 가 아님", proxyPort)
	}
	return false, "ANTHROPIC_BASE_URL 없음"
}

// FixNanoclawEnv: sets ANTHROPIC_BASE_URL=http://localhost:{proxyPort} in
// ~/nanoclaw/.env, replacing an existing entry if present.
// Also replaces any expired sk-ant-... CLAUDE_CODE_OAUTH_TOKEN with vaultToken.
func FixNanoclawEnv(proxyPort int, vaultToken string) error {
	home, _ := os.UserHomeDir()
	envFile := filepath.Join(home, "nanoclaw", ".env")
	data, err := os.ReadFile(envFile)
	if err != nil {
		return fmt.Errorf("~/nanoclaw/.env 없음: %w", err)
	}

	expected := fmt.Sprintf("ANTHROPIC_BASE_URL=http://localhost:%d", proxyPort)
	lines := strings.Split(strings.TrimRight(string(data), "\n"), "\n")
	foundBase, foundToken := false, false
	for i, line := range lines {
		if strings.HasPrefix(line, "ANTHROPIC_BASE_URL=") {
			lines[i] = expected
			foundBase = true
		}
		// replace expired sk-ant API keys with vault token
		if strings.HasPrefix(line, "CLAUDE_CODE_OAUTH_TOKEN=sk-ant-") && vaultToken != "" {
			lines[i] = "CLAUDE_CODE_OAUTH_TOKEN=" + vaultToken
			foundToken = true
		}
	}
	if !foundBase {
		lines = append(lines, expected)
	}
	if foundToken {
		fmt.Println("[fix-nanoclaw] CLAUDE_CODE_OAUTH_TOKEN sk-ant key → vault token 교체")
	}

	return os.WriteFile(envFile, []byte(strings.Join(lines, "\n")+"\n"), 0600)
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
