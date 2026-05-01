package proxy

// Standalone sanitizer for ~/.openclaw/openclaw.json.
//
// applyOpenClawConfig already filters empty-id model entries on every write,
// but that path only fires when wall-vault has reason to rewrite the file
// (vault SSE config_change events, /agent/apply calls, etc.). Hosts that
// never trigger such an event keep whatever pre-guard config they had —
// and OpenClaw 2026.4.29 made schema validation strict enough that a single
// historic empty-id entry crash-loops the gateway on the next restart
// (observed on raspi 2026-05-01: "models.providers.custom.models.0.id:
// Too small: expected string to have >=1 characters").
//
// SanitizeOpenClawConfig provides a one-shot fix that the proxy fires at
// boot. It only edits the file when it actually changes anything, so hosts
// that already have a clean config see no write churn (no mtime bump, no
// backup proliferation).

import (
	"encoding/json"
	"fmt"
	"log"
	"os"
	"path/filepath"
	"time"
)

// SanitizeOpenClawConfig opens ~/.openclaw/openclaw.json (or returns silently
// when the file isn't present — most fleet hosts don't run OpenClaw), removes
// any models.providers.<provider>.models[] entry with an empty id, and
// rewrites the file when something changed. The caller gets back a one-line
// summary suitable for a startup log; nil/no-op cases return ("", nil).
func SanitizeOpenClawConfig() (string, error) {
	home, err := os.UserHomeDir()
	if err != nil {
		return "", nil // can't sanitize without a home dir; not fatal
	}
	path := filepath.Join(home, ".openclaw", "openclaw.json")
	data, err := os.ReadFile(path)
	if err != nil {
		// Most hosts don't run OpenClaw — this is the common case.
		if os.IsNotExist(err) {
			return "", nil
		}
		return "", fmt.Errorf("read %s: %w", path, err)
	}
	var cfg map[string]interface{}
	if err := json.Unmarshal(data, &cfg); err != nil {
		return "", fmt.Errorf("parse %s: %w", path, err)
	}

	removed := sanitizeProviderModels(cfg)
	if removed == 0 {
		return "", nil
	}

	// Backup before writing so an operator can roll back if our sanitize
	// rule ever turns out to be too aggressive. Single suffixed file —
	// don't accumulate multiple backups across restarts.
	bak := path + ".bak.sanitize"
	if err := os.WriteFile(bak, data, 0o600); err != nil {
		log.Printf("[openclaw-sanitize] warning: backup write failed: %v", err)
	}
	out, err := json.MarshalIndent(cfg, "", "  ")
	if err != nil {
		return "", fmt.Errorf("marshal: %w", err)
	}
	if err := os.WriteFile(path, out, 0o600); err != nil {
		return "", fmt.Errorf("write %s: %w", path, err)
	}
	return fmt.Sprintf("removed %d empty-id model entries from %s (backup: %s)",
		removed, path, bak), nil
}

// sanitizeProviderModels walks cfg.models.providers.* and drops any models[]
// entry whose id is empty. Returns the total number of entries removed.
// Mutates cfg in place; callers should marshal cfg back when this returns
// > 0.
func sanitizeProviderModels(cfg map[string]interface{}) int {
	removed := 0
	models, ok := cfg["models"].(map[string]interface{})
	if !ok {
		return 0
	}
	providers, ok := models["providers"].(map[string]interface{})
	if !ok {
		return 0
	}
	for _, pv := range providers {
		provider, ok := pv.(map[string]interface{})
		if !ok {
			continue
		}
		raw, ok := provider["models"].([]interface{})
		if !ok {
			continue
		}
		filtered := make([]interface{}, 0, len(raw))
		for _, m := range raw {
			mm, ok := m.(map[string]interface{})
			if !ok {
				// Non-object entry — preserve so we don't silently
				// reshape data we don't recognise.
				filtered = append(filtered, m)
				continue
			}
			if id, _ := mm["id"].(string); id == "" {
				removed++
				continue
			}
			filtered = append(filtered, m)
		}
		if len(filtered) != len(raw) {
			provider["models"] = filtered
		}
	}
	return removed
}

// runStartupSanitize fires SanitizeOpenClawConfig once at proxy boot in a
// goroutine so a slow disk doesn't delay startup. Logs the result; never
// fatal because the proxy itself doesn't depend on OpenClaw running.
func runStartupSanitize() {
	go func() {
		// Tiny delay so the boot banner gets flushed first — not strictly
		// required but keeps the log ordering readable.
		time.Sleep(500 * time.Millisecond)
		summary, err := SanitizeOpenClawConfig()
		if err != nil {
			log.Printf("[openclaw-sanitize] %v", err)
			return
		}
		if summary != "" {
			log.Printf("[openclaw-sanitize] %s", summary)
		}
	}()
}
