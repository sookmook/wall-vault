// Package setup: interactive setup wizard for beginners
package setup

import (
	"bufio"
	"crypto/rand"
	"encoding/hex"
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/sookmook/wall-vault/internal/config"
	"github.com/sookmook/wall-vault/internal/i18n"
	"github.com/sookmook/wall-vault/internal/models"
)

var reader = bufio.NewReader(os.Stdin)

// Run: wall-vault setup
func Run(_ []string) {
	// auto-detect system language then prompt selection
	i18n.Init()
	selectLanguage()

	fmt.Println()
	fmt.Println("🌸 wall-vault " + i18n.T("setup_welcome"))
	fmt.Println("──────────────────────────────────────────")
	fmt.Println()

	cfg := config.Default()

	// ─── Theme ───────────────────────────────────────────────────────────────
	fmt.Println("테마 / Theme / テーマ:")
	fmt.Println("  1) dark    🌑")
	fmt.Println("  2) light   ☀️")
	fmt.Println("  3) cherry  🌸 (기본/default)")
	fmt.Println("  4) ocean   🌊")
	themeChoice := ask("번호/number", "3")
	themes := map[string]string{"1": "dark", "2": "light", "3": "cherry", "4": "ocean"}
	if t, ok := themes[themeChoice]; ok {
		cfg.Theme = t
	}

	// ─── Mode ────────────────────────────────────────────────────────────────
	fmt.Println("\n운용 방식 / Mode:")
	fmt.Println("  1) standalone  — 이 기기 하나에서 프록시+금고 모두 실행 (권장)")
	fmt.Println("  2) distributed — 금고는 다른 기기, 여기서는 프록시만")
	modeChoice := ask("번호/number", "1")
	if modeChoice == "2" {
		cfg.Mode = "distributed"
		cfg.Proxy.VaultURL = ask("금고 서버 URL / Vault URL", "http://192.168.x.x:56243")
		cfg.Proxy.VaultToken = ask("프록시 인증 토큰 / Vault token", "")
	}

	// ─── Client ID ───────────────────────────────────────────────────────────
	cfg.Proxy.ClientID = ask("\n봇 이름 / Bot name (e.g. my-bot)", "my-bot")

	// ─── Port ────────────────────────────────────────────────────────────────
	portStr := ask("프록시 포트 / Proxy port", "56244")
	fmt.Sscanf(portStr, "%d", &cfg.Proxy.Port)
	if cfg.Mode == "standalone" {
		vaultPortStr := ask("금고 포트 / Vault port", "56243")
		fmt.Sscanf(vaultPortStr, "%d", &cfg.Vault.Port)
	}

	// ─── Service Selection ────────────────────────────────────────────────────
	fmt.Println("\nAI 서비스 선택 / Select AI services:")

	useGoogle := askYN("Google Gemini", false)
	useOpenRouter := askYN("OpenRouter", false)
	useOllama := askYN("Ollama (로컬/local)", true)

	cfg.Proxy.Services = []string{}
	if useGoogle {
		cfg.Proxy.Services = append(cfg.Proxy.Services, "google")
	}
	if useOpenRouter {
		cfg.Proxy.Services = append(cfg.Proxy.Services, "openrouter")
	}
	if useOllama {
		cfg.Proxy.Services = append(cfg.Proxy.Services, "ollama")
	}
	if len(cfg.Proxy.Services) == 0 {
		cfg.Proxy.Services = []string{"ollama"}
		fmt.Println("  → 서비스 미선택, Ollama로 설정합니다.")
	}

	// ─── Ollama Configuration ─────────────────────────────────────────────────
	if useOllama {
		ollamaURL := ask("\nOllama 서버 URL", "http://localhost:11434")
		fmt.Printf("  → %s 에서 모델 목록 조회 중...\n", ollamaURL)
		if ollamaModels, err := models.FetchOllamaPublic(ollamaURL); err == nil && len(ollamaModels) > 0 {
			fmt.Printf("  발견된 모델 %d개:\n", len(ollamaModels))
			for i, m := range ollamaModels {
				if i >= 10 {
					fmt.Printf("    ... 외 %d개\n", len(ollamaModels)-10)
					break
				}
				fmt.Printf("    %d) %s\n", i+1, m.ID)
			}
			defaultModel := ollamaModels[0].ID
			_ = ask("기본 모델 / Default model", defaultModel)
		} else {
			fmt.Println("  (Ollama 미연결 — 나중에 시작 후 자동 조회됩니다)")
		}
		if ollamaURL != "http://localhost:11434" {
			fmt.Printf("\n  환경변수 설정 필요:\n    export WV_OLLAMA_URL=%s\n", ollamaURL)
		}
	}

	// ─── Tool Filter ──────────────────────────────────────────────────────────
	fmt.Println("\n도구 보안 필터 / Tool filter:")
	fmt.Println("  1) strip_all   — 외부 도구 전부 차단 (권장)")
	fmt.Println("  2) passthrough — 필터 없음")
	filterChoice := ask("번호/number", "1")
	if filterChoice == "2" {
		cfg.Proxy.ToolFilter = "passthrough"
	}

	// ─── Vault Security Settings ──────────────────────────────────────────────
	if cfg.Mode == "standalone" {
		fmt.Println("\n금고 보안 설정 / Vault security:")

		adminToken := ask("관리자 토큰 (엔터=자동생성) / Admin token", "")
		if adminToken == "" {
			adminToken = generateToken(24)
			fmt.Printf("  → 자동 생성됨: %s\n", adminToken)
		}
		cfg.Vault.AdminToken = adminToken

		masterPass := ask("API 키 암호화 비밀번호 (엔터=암호화 없음) / Master password", "")
		if masterPass != "" {
			cfg.Vault.MasterPass = masterPass
			fmt.Println("  → API 키가 AES-GCM으로 암호화됩니다")
		} else {
			fmt.Println("  → 암호화 없음 (나중에 설정 파일에서 추가 가능)")
		}
	}

	// ─── Save ────────────────────────────────────────────────────────────────
	savePath := ask("\n설정 파일 저장 경로 / Save path", "wall-vault.yaml")

	if err := config.Save(cfg, savePath); err != nil {
		fmt.Fprintf(os.Stderr, "저장 실패: %v\n", err)
		os.Exit(1)
	}

	absPath, _ := filepath.Abs(savePath)
	fmt.Println()
	fmt.Println("✅ " + i18n.T("setup_done"))
	fmt.Printf("   파일: %s\n", absPath)
	fmt.Println()
	fmt.Println("다음 단계 / Next steps:")
	if useGoogle {
		fmt.Println("  1. Google API 키:")
		fmt.Println("       export WV_KEY_GOOGLE=AIzaSy...")
		fmt.Println("     또는 대시보드 → 🔑 API 키 추가")
	}
	if useOpenRouter {
		fmt.Println("  1. OpenRouter API 키:")
		fmt.Println("       export WV_KEY_OPENROUTER=sk-or-...")
	}
	fmt.Println("  2. wall-vault start")
	fmt.Printf("  3. 대시보드: http://localhost:%d\n", cfg.Vault.Port)
	if cfg.Vault.AdminToken != "" {
		fmt.Printf("  4. 관리자 토큰: %s\n", cfg.Vault.AdminToken)
	}
}

// selectLanguage: language selection prompt (list auto-generated from locales/*.json files)
func selectLanguage() {
	fmt.Println("언어 선택 / Select Language:")
	supported := i18n.Supported
	for idx, code := range supported {
		fmt.Printf("  %2d) %s (%s)\n", idx+1, i18n.LangLabel(code), code)
	}
	// use currently detected language as default
	currentIdx := 1
	for i, code := range supported {
		if code == i18n.Lang() {
			currentIdx = i + 1
			break
		}
	}
	choice := ask(fmt.Sprintf("번호/number (현재 감지: %s)", i18n.Lang()), fmt.Sprintf("%d", currentIdx))
	var n int
	fmt.Sscanf(choice, "%d", &n)
	if n >= 1 && n <= len(supported) {
		i18n.SetLang(supported[n-1])
	}
	cfg := config.Default()
	cfg.Lang = i18n.Lang()
	_ = cfg
}

// ─── Util ─────────────────────────────────────────────────────────────────────

func ask(prompt, defaultVal string) string {
	if defaultVal != "" {
		fmt.Printf("? %s [%s]: ", prompt, defaultVal)
	} else {
		fmt.Printf("? %s: ", prompt)
	}
	line, _ := reader.ReadString('\n')
	line = strings.TrimSpace(line)
	if line == "" {
		return defaultVal
	}
	return line
}

func askYN(prompt string, defaultYes bool) bool {
	hint := "y/N"
	if defaultYes {
		hint = "Y/n"
	}
	fmt.Printf("? %s [%s]: ", prompt, hint)
	line, _ := reader.ReadString('\n')
	line = strings.ToLower(strings.TrimSpace(line))
	if line == "" {
		return defaultYes
	}
	return line == "y" || line == "yes" || line == "네" || line == "是" || line == "oui" || line == "ja" || line == "si" || line == "sí"
}

// generateToken: generate a cryptographically secure random token
func generateToken(bytes int) string {
	b := make([]byte, bytes)
	if _, err := rand.Read(b); err != nil {
		// fallback: timestamp-based (rare case)
		return fmt.Sprintf("wv-%x", b)
	}
	return hex.EncodeToString(b)
}
