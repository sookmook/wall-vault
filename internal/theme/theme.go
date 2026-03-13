// Package theme: UI 테마 (dark/light/cherry/ocean)
package theme

import "os"

// Theme: CSS 변수 세트
type Theme struct {
	Name        string
	Background  string
	Surface     string
	Border      string
	Text        string
	TextMuted   string
	Green       string
	Yellow      string
	Red         string
	Blue        string
	Accent      string
	AccentHover string
}

var themes = map[string]*Theme{
	// 체리 테마 — 봄날 벚꽃, 밝고 화사한 핑크 계열
	"cherry": {
		Name:        "cherry",
		Background:  "#fff5f8",
		Surface:     "#ffffff",
		Border:      "#f8b8d0",
		Text:        "#3a0a28",
		TextMuted:   "#a05878",
		Green:       "#1a9a40",
		Yellow:      "#d07000",
		Red:         "#d81848",
		Blue:        "#1060b8",
		Accent:      "#f0106a",
		AccentHover: "#ff3d85",
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
	// 골드 테마 — 밝고 화려한 황금빛
	"gold": {
		Name:        "gold",
		Background:  "#fffce8",
		Surface:     "#ffffff",
		Border:      "#e8c840",
		Text:        "#2a1e00",
		TextMuted:   "#8a6e10",
		Green:       "#2a8020",
		Yellow:      "#c08000",
		Red:         "#c83020",
		Blue:        "#2060b0",
		Accent:      "#b87800",
		AccentHover: "#d49000",
	},
	// 라이트 테마
	"light": {
		Name:        "light",
		Background:  "#f0f2f5",
		Surface:     "#ffffff",
		Border:      "#dde1e9",
		Text:        "#1d2433",
		TextMuted:   "#6b7385",
		Green:       "#1a8a3c",
		Yellow:      "#c47d00",
		Red:         "#d93025",
		Blue:        "#1a6fd4",
		Accent:      "#5c5fc4",
		AccentHover: "#4a4db0",
	},
	// 오션 테마 — 맑은 하늘과 에메랄드 바다
	"ocean": {
		Name:        "ocean",
		Background:  "#e8f8ff",
		Surface:     "#ffffff",
		Border:      "#80d0f0",
		Text:        "#062040",
		TextMuted:   "#2878a0",
		Green:       "#008060",
		Yellow:      "#c07000",
		Red:         "#d03030",
		Blue:        "#0080c0",
		Accent:      "#0098d8",
		AccentHover: "#20b8f8",
	},
	// 가을 테마 — 밝고 옅은 브라운 계통, 단풍잎 가을
	"autumn": {
		Name:        "autumn",
		Background:  "#fff8f0",
		Surface:     "#ffffff",
		Border:      "#e8c8a0",
		Text:        "#3a2010",
		TextMuted:   "#9a6030",
		Green:       "#6a9a20",
		Yellow:      "#c08000",
		Red:         "#c83020",
		Blue:        "#4060a0",
		Accent:      "#d04010",
		AccentHover: "#e85820",
	},
	// 겨울 테마 — 하얀 배경, 눈사람과 트리 뱅글벵글
	"winter": {
		Name:        "winter",
		Background:  "#f4f8ff",
		Surface:     "#ffffff",
		Border:      "#b8d4f0",
		Text:        "#1a2840",
		TextMuted:   "#5878a0",
		Green:       "#087850",
		Yellow:      "#d89800",
		Red:         "#c82020",
		Blue:        "#1868c0",
		Accent:      "#1870d8",
		AccentHover: "#3090f8",
	},
}

// Get: 테마 반환 (없으면 cherry)
func Get(name string) *Theme {
	if t, ok := themes[name]; ok {
		return t
	}
	return themes["cherry"]
}

// Current: 환경변수 또는 기본값으로 테마 반환
func Current() *Theme {
	name := os.Getenv("WV_THEME")
	if name == "" {
		name = "cherry"
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
	return []string{"light", "dark", "gold", "cherry", "ocean", "autumn", "winter"}
}
