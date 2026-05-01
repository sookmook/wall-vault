package proxy

// OpenClaw version detection + version-aware config application.
//
// Today (2026.4.29) all OpenClaw versions wall-vault talks to share the same
// `models.providers.custom` schema, so applyOpenClawConfig writes the same
// fields regardless of version. Past upgrades have not broken the write path
// (see CHANGELOG fork research, 2026-05-01). This file exists so when the
// schema *does* fork — and based on the OpenClaw release cadence (~weekly
// patch versions) it eventually will — we can switch on detected version
// without rewriting the dispatcher.
//
// Two near-term values it provides regardless of schema divergence:
//   1. Diagnostics: agent_apply logs `openclaw_version=2026.4.29` so future
//      bug reports show which schema fork the writer expected to satisfy.
//   2. `meta.lastTouchedVersion` in the written openclaw.json — operators
//      grepping the file can tell whether it was last touched by a writer
//      that thought the user was on 4.26 vs 4.29.

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"strconv"
	"strings"
)

// openClawVersion is the parsed CalVer version of a local OpenClaw install.
// Empty Raw means detection failed (binary not installed, package.json
// unreadable). Callers should treat empty as "use the latest known schema".
type openClawVersion struct {
	Raw   string // "2026.4.29"
	Year  int
	Month int
	Patch int
}

// detectOpenClawVersion reads the OpenClaw npm package's package.json from
// the well-known npm-global path. Returns an empty struct (no error) when
// OpenClaw isn't installed — the caller should fall back to "latest known
// schema" behaviour rather than refusing to write the config.
//
// Search order:
//   1. ~/.npm-global/lib/node_modules/openclaw/package.json   (user-scoped npm)
//   2. /usr/lib/node_modules/openclaw/package.json             (Debian apt)
//   3. /usr/local/lib/node_modules/openclaw/package.json       (system npm)
//   4. /opt/homebrew/lib/node_modules/openclaw/package.json    (Apple Silicon brew)
func detectOpenClawVersion() openClawVersion {
	candidates := []string{}
	if home, err := os.UserHomeDir(); err == nil {
		candidates = append(candidates,
			filepath.Join(home, ".npm-global", "lib", "node_modules", "openclaw", "package.json"))
	}
	candidates = append(candidates,
		"/usr/lib/node_modules/openclaw/package.json",
		"/usr/local/lib/node_modules/openclaw/package.json",
		"/opt/homebrew/lib/node_modules/openclaw/package.json",
	)
	for _, p := range candidates {
		data, err := os.ReadFile(p)
		if err != nil {
			continue
		}
		var pkg struct {
			Version string `json:"version"`
		}
		if err := json.Unmarshal(data, &pkg); err != nil {
			continue
		}
		if pkg.Version == "" {
			continue
		}
		return parseOpenClawVersion(pkg.Version)
	}
	return openClawVersion{}
}

// parseOpenClawVersion turns "2026.4.29" into a structured version. Bad
// strings return an empty struct rather than panicking; the writer treats
// detection failure the same as "latest known schema".
func parseOpenClawVersion(s string) openClawVersion {
	s = strings.TrimSpace(s)
	parts := strings.Split(s, ".")
	if len(parts) != 3 {
		return openClawVersion{Raw: s}
	}
	year, errY := strconv.Atoi(parts[0])
	month, errM := strconv.Atoi(parts[1])
	patch, errP := strconv.Atoi(parts[2])
	if errY != nil || errM != nil || errP != nil {
		return openClawVersion{Raw: s}
	}
	return openClawVersion{Raw: s, Year: year, Month: month, Patch: patch}
}

// gte returns true if v is at least year.month.patch. Empty Raw (detection
// failed) returns true so callers default to the newest schema branch.
func (v openClawVersion) gte(year, month, patch int) bool {
	if v.Raw == "" {
		return true
	}
	if v.Year != year {
		return v.Year > year
	}
	if v.Month != month {
		return v.Month > month
	}
	return v.Patch >= patch
}

// schemaTag returns a short identifier for the config-schema branch this
// version maps to. Today there is exactly one branch ("v1") covering every
// 2026.x release the writer has been tested against. When OpenClaw breaks
// the schema we add a new tag here and route the writer accordingly.
func (v openClawVersion) schemaTag() string {
	// Hypothetical future fork: if 2026.6.0 introduced a new providers
	// layout we would return "v2" for v.gte(2026, 6, 0). Until then every
	// reachable version maps to v1.
	return "v1"
}

// describe formats the version for log lines and meta.lastTouchedVersion.
// Returns "unknown" rather than empty so log greps stay readable.
func (v openClawVersion) describe() string {
	if v.Raw == "" {
		return "unknown"
	}
	return fmt.Sprintf("%s (schema=%s)", v.Raw, v.schemaTag())
}
