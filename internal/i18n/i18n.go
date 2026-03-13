// Package i18n: 언어 파일 자동 로드 기반 다국어 지원
// locales/*.json 파일을 추가하면 자동으로 지원 언어에 포함됩니다.
package i18n

import (
	"embed"
	"encoding/json"
	"os"
	"sort"
	"strings"
)

//go:embed locales/*.json
var localeFS embed.FS

var lang = "en"

// Supported: 지원 언어 코드 목록 (locales/*.json 파일에서 자동 결정)
var Supported []string

var messages = map[string]map[string]string{}

func init() {
	entries, err := localeFS.ReadDir("locales")
	if err != nil {
		return
	}
	for _, e := range entries {
		if e.IsDir() || !strings.HasSuffix(e.Name(), ".json") {
			continue
		}
		code := strings.TrimSuffix(e.Name(), ".json")
		data, err := localeFS.ReadFile("locales/" + e.Name())
		if err != nil {
			continue
		}
		var m map[string]string
		if err := json.Unmarshal(data, &m); err != nil {
			continue
		}
		messages[code] = m
		Supported = append(Supported, code)
	}
	// 우선 순서: ko, en 앞에 배치
	sort.Slice(Supported, func(i, j int) bool {
		order := map[string]int{"ko": 0, "en": 1}
		oi, ok1 := order[Supported[i]]
		oj, ok2 := order[Supported[j]]
		if ok1 && ok2 {
			return oi < oj
		}
		if ok1 {
			return true
		}
		if ok2 {
			return false
		}
		return Supported[i] < Supported[j]
	})
}

// Init: 환경변수 또는 시스템 로케일에서 언어 결정
func Init() {
	if v := os.Getenv("WV_LANG"); v != "" {
		SetLang(v)
		return
	}
	if v := os.Getenv("LANG"); v != "" {
		SetLang(v)
		return
	}
}

// SetLang: 언어 코드 설정 (외부 패키지에서 호출 가능)
func SetLang(v string) {
	v = strings.ToLower(v)
	for _, code := range Supported {
		if strings.HasPrefix(v, code) {
			lang = code
			return
		}
	}
	switch {
	case strings.HasPrefix(v, "zh"):
		lang = "zh"
	case strings.HasPrefix(v, "he"), strings.HasPrefix(v, "iw"):
		lang = "ar"
	default:
		lang = "en"
	}
}

// T: 현재 언어로 메시지 반환 (영어 폴백)
func T(key string) string {
	if m, ok := messages[lang]; ok {
		if s, ok := m[key]; ok {
			return s
		}
	}
	if m, ok := messages["en"]; ok {
		if s, ok := m[key]; ok {
			return s
		}
	}
	return key
}

// Lang: 현재 언어 코드 반환
func Lang() string { return lang }

// LangLabel: 언어 표시 문자열 반환 (드롭다운용)
func LangLabel(code string) string {
	if m, ok := messages[code]; ok {
		emoji := m["lang_emoji"]
		label := m["lang_label"]
		if emoji != "" && label != "" {
			return emoji + " " + label
		}
	}
	return "🌍 " + code
}

// AllLangs: 웹 UI JS I18N 객체 생성용 — 전체 언어 맵 반환
func AllLangs() map[string]map[string]string {
	return messages
}
