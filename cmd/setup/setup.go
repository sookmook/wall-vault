// Package setup: 초보자용 대화형 설치 마법사
package setup

import (
	"bufio"
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/sookmook/wall-vault/internal/config"
	"github.com/sookmook/wall-vault/internal/models"
)

var reader = bufio.NewReader(os.Stdin)

// Run: wall-vault setup
func Run(_ []string) {
	fmt.Println()
	fmt.Println("🌸 wall-vault 설치 마법사")
	fmt.Println("──────────────────────────")
	fmt.Println("설정 파일(wall-vault.yaml)을 만들어드립니다.")
	fmt.Println("모르면 그냥 엔터를 누르면 됩니다.")
	fmt.Println()

	cfg := config.Default()

	// ─── 언어 ────────────────────────────────────────────────────────────────
	cfg.Lang = ask("언어 선택 [ko/en/ja]", "ko")

	// ─── 테마 ────────────────────────────────────────────────────────────────
	fmt.Println("\n테마 선택:")
	fmt.Println("  1) sakura  🌸 벚꽃 (기본)")
	fmt.Println("  2) dark    어두운 테마")
	fmt.Println("  3) light   밝은 테마")
	fmt.Println("  4) ocean   바다 테마")
	themeChoice := ask("번호 입력", "1")
	themes := map[string]string{"1": "sakura", "2": "dark", "3": "light", "4": "ocean"}
	if t, ok := themes[themeChoice]; ok {
		cfg.Theme = t
	}

	// ─── 모드 ────────────────────────────────────────────────────────────────
	fmt.Println("\n운용 방식:")
	fmt.Println("  1) standalone  이 기기 하나에서 프록시+금고 모두 실행 (권장)")
	fmt.Println("  2) distributed 금고는 다른 기기에 있고 여기서는 프록시만")
	modeChoice := ask("번호 입력", "1")
	if modeChoice == "2" {
		cfg.Mode = "distributed"
		cfg.Proxy.VaultURL = ask("금고 서버 URL", "http://192.168.0.6:56243")
		cfg.Proxy.VaultToken = ask("프록시 인증 토큰 (금고에서 발급)", "")
	}

	// ─── 클라이언트 ID ────────────────────────────────────────────────────────
	cfg.Proxy.ClientID = ask("\n봇 이름 (영문, 예: motoko)", "my-bot")

	// ─── 포트 ────────────────────────────────────────────────────────────────
	portStr := ask("프록시 포트", "56244")
	fmt.Sscanf(portStr, "%d", &cfg.Proxy.Port)
	vaultPortStr := ask("금고 포트", "56243")
	fmt.Sscanf(vaultPortStr, "%d", &cfg.Vault.Port)

	// ─── 서비스 선택 ──────────────────────────────────────────────────────────
	fmt.Println("\n사용할 AI 서비스 (모두 엔터로 건너뛰어도 됩니다):")

	useGoogle := askYN("Google Gemini 사용?", false)
	useOpenRouter := askYN("OpenRouter 사용?", false)
	useOllama := askYN("Ollama 로컬 모델 사용?", true)

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

	// ─── Ollama 설정 ──────────────────────────────────────────────────────────
	if useOllama {
		ollamaURL := ask("\nOllama 서버 URL", "http://localhost:11434")
		fmt.Printf("  → %s 에서 모델 목록 조회 중...\n", ollamaURL)
		if ollamaModels, err := models.FetchOllamaPublic(ollamaURL); err == nil && len(ollamaModels) > 0 {
			fmt.Println("  발견된 Ollama 모델:")
			for i, m := range ollamaModels {
				if i >= 10 {
					fmt.Printf("    ... 외 %d개\n", len(ollamaModels)-10)
					break
				}
				fmt.Printf("    %d) %s\n", i+1, m.ID)
			}
			if len(ollamaModels) > 0 {
				defaultModel := ollamaModels[0].ID
				chosen := ask(fmt.Sprintf("기본 Ollama 모델", ), defaultModel)
				_ = chosen // TODO: 클라이언트 기본 모델에 반영
			}
		} else {
			fmt.Println("  (Ollama 미연결 — 나중에 시작 후 자동 조회됩니다)")
		}
		if ollamaURL != "http://localhost:11434" {
			fmt.Printf("\n환경변수 설정 필요:\n  export WV_OLLAMA_URL=%s\n", ollamaURL)
		}
	}

	// ─── 도구 필터 ────────────────────────────────────────────────────────────
	fmt.Println("\n도구 보안 필터:")
	fmt.Println("  1) strip_all    외부 도구 전부 차단 (권장 — 보안상 안전)")
	fmt.Println("  2) passthrough  필터 없음")
	filterChoice := ask("번호 입력", "1")
	if filterChoice == "2" {
		cfg.Proxy.ToolFilter = "passthrough"
	}

	// ─── 금고 토큰 ────────────────────────────────────────────────────────────
	if cfg.Mode == "standalone" {
		fmt.Println("\n금고 관리자 토큰 설정 (대시보드 접근용):")
		adminToken := ask("관리자 토큰 (엔터=자동생성)", "")
		if adminToken == "" {
			adminToken = generateToken()
			fmt.Printf("  → 자동 생성됨: %s\n", adminToken)
		}
		cfg.Vault.AdminToken = adminToken
	}

	// ─── 저장 ────────────────────────────────────────────────────────────────
	savePath := ask("\n설정 파일 저장 위치", "wall-vault.yaml")

	if err := config.Save(cfg, savePath); err != nil {
		fmt.Fprintf(os.Stderr, "저장 실패: %v\n", err)
		os.Exit(1)
	}

	absPath, _ := filepath.Abs(savePath)
	fmt.Println()
	fmt.Println("✅ 설정 완료!")
	fmt.Printf("   파일: %s\n", absPath)
	fmt.Println()
	fmt.Println("다음 단계:")
	if useGoogle || useOpenRouter {
		fmt.Println("  1. API 키 등록 (금고 시작 후 대시보드에서):")
		if useGoogle {
			fmt.Println("     export WV_KEY_GOOGLE=AIza...")
		}
		if useOpenRouter {
			fmt.Println("     export WV_KEY_OPENROUTER=sk-or-...")
		}
	}
	fmt.Println("  2. wall-vault start")
	fmt.Printf("  3. 브라우저: http://localhost:%d\n", cfg.Vault.Port)
}

// ─── 유틸 ────────────────────────────────────────────────────────────────────

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
	return line == "y" || line == "yes" || line == "네"
}

func generateToken() string {
	b := make([]byte, 16)
	// crypto/rand 없이 간단히
	for i := range b {
		b[i] = "abcdefghijklmnopqrstuvwxyz0123456789"[i%36]
	}
	return string(b)
}
