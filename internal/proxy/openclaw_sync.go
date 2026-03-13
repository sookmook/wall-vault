package proxy

import (
	"encoding/json"
	"log"
	"os"
	"path/filepath"
	"time"
)

// updateOpenClawJSON updates ~/.openclaw/openclaw.json when the model changes via SSE.
// This keeps OpenClaw TUI in sync with the vault config without restarting.
//
// Supports openclaw.json v2026.3.12+ format:
//   models.providers.<name>.api = "openai-completions"
//   models.providers.<name>.models[].{id, name, reasoning, input, contextWindow, maxTokens}
func updateOpenClawJSON(service, model string) {
	home, err := os.UserHomeDir()
	if err != nil {
		return
	}
	path := filepath.Join(home, ".openclaw", "openclaw.json")

	data, err := os.ReadFile(path)
	if err != nil {
		return // file doesn't exist — skip silently
	}

	var cfg map[string]interface{}
	if err := json.Unmarshal(data, &cfg); err != nil {
		log.Printf("[openclaw-sync] failed to parse openclaw.json: %v", err)
		return
	}

	// Build the provider/model string OpenClaw expects (e.g. "custom/gemini-2.5-flash")
	primaryModel := "custom/" + model

	changed := false

	// ── 1. agents.defaults.model.primary ──────────────────────────────────────
	if agents, ok := cfg["agents"].(map[string]interface{}); ok {
		if defaults, ok := agents["defaults"].(map[string]interface{}); ok {
			mdl, _ := defaults["model"].(map[string]interface{})
			if mdl == nil {
				mdl = map[string]interface{}{}
				defaults["model"] = mdl
			}
			if mdl["primary"] != primaryModel {
				mdl["primary"] = primaryModel
				changed = true
			}
		}
	}

	// ── 2. models.providers.custom — v2026.3.12 format ────────────────────────
	// Path: cfg["models"]["providers"]["custom"]["models"]
	modelsSection, _ := cfg["models"].(map[string]interface{})
	if modelsSection == nil {
		modelsSection = map[string]interface{}{}
		cfg["models"] = modelsSection
	}
	providers, _ := modelsSection["providers"].(map[string]interface{})
	if providers == nil {
		providers = map[string]interface{}{}
		modelsSection["providers"] = providers
	}
	custom, _ := providers["custom"].(map[string]interface{})
	if custom == nil {
		custom = map[string]interface{}{
			"baseUrl":    "http://localhost:56244/v1",
			"apiKey":     "proxy-managed",
			"api":        "openai-completions",
			"authHeader": false,
		}
		providers["custom"] = custom
		changed = true
	}

	// Ensure provider-level api field is set (migration from per-model api)
	if custom["api"] == nil {
		custom["api"] = "openai-completions"
		changed = true
	}
	// Migrate old baseUrl from Gemini path to OpenAI path
	if bu, _ := custom["baseUrl"].(string); bu == "http://localhost:56244/google/v1beta" {
		custom["baseUrl"] = "http://localhost:56244/v1"
		changed = true
	}

	// Add model entry if missing
	models, _ := custom["models"].([]interface{})
	found := false
	for _, m := range models {
		if mm, ok := m.(map[string]interface{}); ok {
			if mm["id"] == model {
				found = true
				// Remove legacy per-model api field if present
				if _, hasAPI := mm["api"]; hasAPI {
					delete(mm, "api")
					changed = true
				}
				break
			}
		}
	}
	if !found {
		entry := map[string]interface{}{
			"id":            model,
			"name":          service + " / " + model,
			"reasoning":     false,
			"input":         []interface{}{"text"},
			"contextWindow": 1048576,
			"maxTokens":     65536,
		}
		models = append([]interface{}{entry}, models...)
		custom["models"] = models
		changed = true
	}

	// ── 3. meta version bump ───────────────────────────────────────────────────
	meta, _ := cfg["meta"].(map[string]interface{})
	if meta == nil {
		meta = map[string]interface{}{}
		cfg["meta"] = meta
	}
	meta["lastTouchedVersion"] = "2026.3.12"
	meta["lastTouchedAt"] = time.Now().UTC().Format(time.RFC3339Nano)
	changed = true

	if !changed {
		return
	}

	out, err := json.MarshalIndent(cfg, "", "  ")
	if err != nil {
		log.Printf("[openclaw-sync] failed to marshal openclaw.json: %v", err)
		return
	}

	if err := os.WriteFile(path, out, 0644); err != nil {
		log.Printf("[openclaw-sync] failed to write openclaw.json: %v", err)
		return
	}

	log.Printf("[openclaw-sync] openclaw.json updated: primary=%s (v2026.3.12 format)", primaryModel)
}
