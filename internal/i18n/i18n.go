// Package i18n: multi-language support based on automatic locale file loading
// adding locales/*.json files automatically includes them in supported languages.
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

// Supported: list of supported language codes (auto-determined from locales/*.json files)
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
	// sort order: place ko, en at front
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

// Init: determine language from env var or system locale
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

// SetLang: set language code (callable from external packages)
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

// T: return message in current language (English fallback)
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

// Lang: return current language code
func Lang() string { return lang }

// LangLabel: return language display string (for dropdowns)
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

// AllLangs: return full language map — for building web UI JS I18N object
func AllLangs() map[string]map[string]string {
	return messages
}
