// Package theme: UI 테마 (sakura/dark/light/ocean)
package theme

import "os"

// Theme: CSS 변수 세트
type Theme struct {
	Name       string
	Background string
	Surface    string
	Border     string
	Text       string
	TextMuted  string
	Green      string
	Yellow     string
	Red        string
	Blue       string
	Accent     string
	AccentHover string
}

var themes = map[string]*Theme{
	// 벚꽃 테마 (기본) — 화사하고 따뜻한 핑크 계열
	"sakura": {
		Name:        "sakura",
		Background:  "#1a0a0f",
		Surface:     "#2d1520",
		Border:      "#5c2d3e",
		Text:        "#f7d4e0",
		TextMuted:   "#a06070",
		Green:       "#a8d8a8",
		Yellow:      "#f5d76e",
		Red:         "#f07070",
		Blue:        "#8ec5e6",
		Accent:      "#e88fa8",
		AccentHover: "#f0aac0",
	},
	// 다크 테마
	"dark": {
		Name:        "dark",
		Background:  "#0d0d0d",
		Surface:     "#1a1a1a",
		Border:      "#333333",
		Text:        "#e0e0e0",
		TextMuted:   "#666666",
		Green:       "#4caf50",
		Yellow:      "#ffc107",
		Red:         "#f44336",
		Blue:        "#2196f3",
		Accent:      "#bb86fc",
		AccentHover: "#cf9fff",
	},
	// 라이트 테마
	"light": {
		Name:        "light",
		Background:  "#f5f5f5",
		Surface:     "#ffffff",
		Border:      "#dddddd",
		Text:        "#212121",
		TextMuted:   "#757575",
		Green:       "#388e3c",
		Yellow:      "#f57c00",
		Red:         "#d32f2f",
		Blue:        "#1976d2",
		Accent:      "#6200ee",
		AccentHover: "#7c2cff",
	},
	// 오션 테마
	"ocean": {
		Name:        "ocean",
		Background:  "#0a1628",
		Surface:     "#0d2040",
		Border:      "#1a4060",
		Text:        "#c8e0f0",
		TextMuted:   "#507090",
		Green:       "#00e5a0",
		Yellow:      "#ffd060",
		Red:         "#ff6060",
		Blue:        "#40c0f0",
		Accent:      "#00b4d8",
		AccentHover: "#48cae4",
	},
}

// Get: 테마 반환 (없으면 sakura)
func Get(name string) *Theme {
	if t, ok := themes[name]; ok {
		return t
	}
	return themes["sakura"]
}

// Current: 환경변수 또는 기본값으로 테마 반환
func Current() *Theme {
	name := os.Getenv("WV_THEME")
	if name == "" {
		name = "sakura"
	}
	return Get(name)
}

// CSSVars: HTML <style> 태그용 CSS 변수 문자열 생성
func (t *Theme) CSSVars() string {
	return `
  --bg: ` + t.Background + `;
  --surface: ` + t.Surface + `;
  --border: ` + t.Border + `;
  --text: ` + t.Text + `;
  --text-muted: ` + t.TextMuted + `;
  --green: ` + t.Green + `;
  --yellow: ` + t.Yellow + `;
  --red: ` + t.Red + `;
  --blue: ` + t.Blue + `;
  --accent: ` + t.Accent + `;
  --accent-hover: ` + t.AccentHover + `;`
}

// List: 사용 가능한 테마 목록
func List() []string {
	return []string{"sakura", "dark", "light", "ocean"}
}
