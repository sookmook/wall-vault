// Package theme: UI themes (dark/light/cherry/ocean/gold/autumn/winter)
package theme

import "os"

// Theme: CSS variable set
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
	// cherry theme — spring cherry blossoms, bright and vivid pink tones
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
	// dark theme
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
	// gold theme — bright and dazzling golden tones
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
	// light theme
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
	// ocean theme — clear sky and emerald sea
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
	// autumn theme — deep amber-brown, fallen leaves in late autumn
	"autumn": {
		Name:        "autumn",
		Background:  "#ede0c8",
		Surface:     "#f5e8d2",
		Border:      "#b88848",
		Text:        "#201008",
		TextMuted:   "#6a3e18",
		Green:       "#4a7810",
		Yellow:      "#986000",
		Red:         "#a82010",
		Blue:        "#284070",
		Accent:      "#b03008",
		AccentHover: "#cc4818",
	},
	// winter theme — white background, spinning snowmen and trees
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

// Get: return theme (falls back to cherry if not found)
func Get(name string) *Theme {
	if t, ok := themes[name]; ok {
		return t
	}
	return themes["cherry"]
}

// Current: return theme from env var or default
func Current() *Theme {
	name := os.Getenv("WV_THEME")
	if name == "" {
		name = "cherry"
	}
	return Get(name)
}

// CSSVars: generate CSS variable string for HTML <style> tag
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

// List: list of available themes
func List() []string {
	return []string{"light", "dark", "gold", "cherry", "ocean", "autumn", "winter"}
}
