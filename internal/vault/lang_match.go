package vault

// Accept-Language → supported locale matcher.
//
// Used by /setup, /login, and the bootstrap listener to pick a UI language
// when no session/cookie is present yet. Browsers send Accept-Language
// strings like "ko,en-US;q=0.9,en;q=0.8" — we honour the q-weighted order
// and stop at the first entry whose primary subtag we ship a locale for.

import (
	"sort"
	"strconv"
	"strings"

	"github.com/sookmook/wall-vault/internal/i18n"
)

// i18nSupported reports whether a locale code is among the embedded
// locales/*.json files. Lets callers reject explicit ?lang overrides for
// codes we have no translations for so they fall back gracefully.
func i18nSupported(code string) bool {
	for _, c := range i18n.Supported {
		if c == code {
			return true
		}
	}
	return false
}

// matchAcceptLanguage parses a single Accept-Language header value and
// returns the highest-q locale we actually support. Empty string means no
// supported locale found; the caller should then use its own default.
func matchAcceptLanguage(header string) string {
	type cand struct {
		code string
		q    float64
	}
	var cands []cand
	for _, raw := range strings.Split(header, ",") {
		raw = strings.TrimSpace(raw)
		if raw == "" {
			continue
		}
		code := raw
		q := 1.0
		if i := strings.Index(raw, ";"); i != -1 {
			code = strings.TrimSpace(raw[:i])
			rest := raw[i+1:]
			for _, p := range strings.Split(rest, ";") {
				p = strings.TrimSpace(p)
				if strings.HasPrefix(p, "q=") {
					if v, err := strconv.ParseFloat(p[2:], 64); err == nil {
						q = v
					}
				}
			}
		}
		// Reduce "ko-KR" → "ko" since our locales are language-only.
		if i := strings.Index(code, "-"); i != -1 {
			code = code[:i]
		}
		code = strings.ToLower(code)
		cands = append(cands, cand{code: code, q: q})
	}
	// Stable sort by descending q so the browser's preference order wins
	// when q ties (the user agent emits them in preference order already).
	sort.SliceStable(cands, func(i, j int) bool { return cands[i].q > cands[j].q })
	for _, c := range cands {
		if i18nSupported(c.code) {
			return c.code
		}
	}
	return ""
}
