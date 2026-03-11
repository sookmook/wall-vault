// Package i18n: 다국어 지원 (ko/en/ja)
package i18n

import (
	"os"
	"strings"
)

var lang = "ko"

var messages = map[string]map[string]string{
	"ko": {
		"unknown_command":      "알 수 없는 명령",
		"starting":             "시작 중",
		"stopping":             "종료 중",
		"proxy_started":        "프록시 서버 시작됨",
		"vault_started":        "키 금고 서버 시작됨",
		"health_ok":            "정상",
		"health_err":           "오류",
		"ollama_fetching":      "Ollama 모델 목록 조회 중...",
		"ollama_found":         "Ollama 모델 발견",
		"ollama_unavailable":   "Ollama 미사용 가능",
		"key_cooldown":         "키 쿨다운",
		"key_active":           "키 활성",
		"key_exhausted":        "키 소진",
		"waiting_ollama":       "Ollama 대기 중...",
		"service_not_found":    "서비스를 찾을 수 없음",
		"model_changed":        "모델 변경됨",
		"config_loaded":        "설정 로드 완료",
		"config_not_found":     "설정 파일 없음 — 기본값 사용",
	},
	"en": {
		"unknown_command":      "unknown command",
		"starting":             "starting",
		"stopping":             "stopping",
		"proxy_started":        "proxy server started",
		"vault_started":        "key vault server started",
		"health_ok":            "ok",
		"health_err":           "error",
		"ollama_fetching":      "fetching Ollama model list...",
		"ollama_found":         "Ollama models found",
		"ollama_unavailable":   "Ollama not available",
		"key_cooldown":         "key on cooldown",
		"key_active":           "key active",
		"key_exhausted":        "key exhausted",
		"waiting_ollama":       "waiting for Ollama...",
		"service_not_found":    "service not found",
		"model_changed":        "model changed",
		"config_loaded":        "config loaded",
		"config_not_found":     "config not found — using defaults",
	},
	"ja": {
		"unknown_command":      "不明なコマンド",
		"starting":             "起動中",
		"stopping":             "停止中",
		"proxy_started":        "プロキシサーバー起動済み",
		"vault_started":        "キーボルト起動済み",
		"health_ok":            "正常",
		"health_err":           "エラー",
		"ollama_fetching":      "Ollamaモデル一覧取得中...",
		"ollama_found":         "Ollamaモデル発見",
		"ollama_unavailable":   "Ollama利用不可",
		"key_cooldown":         "キークールダウン",
		"key_active":           "キーアクティブ",
		"key_exhausted":        "キー枯渇",
		"waiting_ollama":       "Ollama待機中...",
		"service_not_found":    "サービスが見つかりません",
		"model_changed":        "モデル変更済み",
		"config_loaded":        "設定読み込み完了",
		"config_not_found":     "設定ファイルなし — デフォルト使用",
	},
}

// Init: 환경변수에서 언어 결정
func Init() {
	if v := os.Getenv("WV_LANG"); v != "" {
		setLang(v)
		return
	}
	if v := os.Getenv("LANG"); v != "" {
		setLang(v)
	}
}

func setLang(v string) {
	v = strings.ToLower(v)
	switch {
	case strings.HasPrefix(v, "ko"):
		lang = "ko"
	case strings.HasPrefix(v, "ja"):
		lang = "ja"
	default:
		lang = "en"
	}
}

// T: 현재 언어로 메시지 반환
func T(key string) string {
	if m, ok := messages[lang]; ok {
		if s, ok := m[key]; ok {
			return s
		}
	}
	// 영어 폴백
	if m, ok := messages["en"]; ok {
		if s, ok := m[key]; ok {
			return s
		}
	}
	return key
}

// Lang: 현재 언어 코드 반환
func Lang() string { return lang }
