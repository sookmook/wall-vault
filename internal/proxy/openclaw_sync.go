package proxy

import (
	"encoding/json"
	"log"
	"os"
	"path/filepath"
)

// updateOpenClawJSON updates ~/.openclaw/openclaw.json when the model changes via SSE.
// This keeps OpenClaw TUI in sync with the vault config without restarting.
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
	// We always route through the "custom" provider (wall-vault proxy endpoint)
	primaryModel := "custom/" + model

	changed := false

	// 1. Update agents.defaults.model.primary
	if agents, ok := cfg["agents"].(map[string]interface{}); ok {
		if defaults, ok := agents["defaults"].(map[string]interface{}); ok {
			if mdl, ok := defaults["model"].(map[string]interface{}); ok {
				if mdl["primary"] != primaryModel {
					mdl["primary"] = primaryModel
					changed = true
				}
			}
		}
	}

	// 2. Ensure the model exists in custom provider models list
	if modelProviders, ok := cfg["modelProviders"].(map[string]interface{}); ok {
		if custom, ok := modelProviders["custom"].(map[string]interface{}); ok {
			models, _ := custom["models"].([]interface{})
			found := false
			for _, m := range models {
				if mm, ok := m.(map[string]interface{}); ok {
					if mm["id"] == model {
						found = true
						break
					}
				}
			}
			if !found {
				models = append([]interface{}{map[string]interface{}{
					"id":   model,
					"name": service + "/" + model,
					"api":  "google-generative-ai",
				}}, models...)
				custom["models"] = models
				changed = true
			}
		}
	}

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

	log.Printf("[openclaw-sync] openclaw.json updated: primary=%s", primaryModel)
}
