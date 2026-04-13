package vault

import (
	"encoding/json"
	"fmt"
	"time"
)

// MigrateV1ToV2 converts the legacy v0.1.x decrypted vault envelope JSON
// into a *storeData (v2). Field mapping per spec §7.2:
//
//   services[].default_model      ← most frequent clients[].default_model whose
//                                   default_service matches this id (ties broken
//                                   by lowest sort_order client); "" otherwise.
//   services[].allowed_models     ← nil (empty = unrestricted).
//   services[].sort_order         ← preserve v1 if present, else 0.
//   clients[].preferred_service   ← clients[].default_service
//   clients[].model_override      ← clients[].default_model
//
// api_keys, theme, lang, ip_whitelist, work_dir, avatar, etc. carry over verbatim.
func MigrateV1ToV2(raw []byte) (*storeData, error) {
	var legacy struct {
		Services []struct {
			ID           string `json:"id"`
			Name         string `json:"name"`
			LocalURL     string `json:"local_url,omitempty"`
			Enabled      bool   `json:"enabled"`
			ProxyEnabled bool   `json:"proxy_enabled"`
			SortOrder    int    `json:"sort_order,omitempty"`
		} `json:"services"`
		Clients []struct {
			ID              string    `json:"id"`
			Name            string    `json:"name"`
			Token           string    `json:"token"`
			DefaultService  string    `json:"default_service"`
			DefaultModel    string    `json:"default_model"`
			AllowedServices []string  `json:"allowed_services,omitempty"`
			AgentType       string    `json:"agent_type,omitempty"`
			WorkDir         string    `json:"work_dir,omitempty"`
			IPWhitelist     []string  `json:"ip_whitelist,omitempty"`
			Avatar          string    `json:"avatar,omitempty"`
			Enabled         bool      `json:"enabled"`
			SortOrder       int       `json:"sort_order"`
			CreatedAt       time.Time `json:"created_at"`
		} `json:"clients"`
		APIKeys  []*APIKey `json:"api_keys"`
		Theme    string    `json:"theme,omitempty"`
		Lang     string    `json:"lang,omitempty"`
	}
	if err := json.Unmarshal(raw, &legacy); err != nil {
		return nil, fmt.Errorf("migrate: decode v1: %w", err)
	}

	// Per-service majority vote for default_model, tie-broken by
	// lowest sort_order client.
	type tally struct{ count, minSort int }
	stats := make(map[string]map[string]tally) // svc → model → tally
	for _, c := range legacy.Clients {
		if c.DefaultService == "" || c.DefaultModel == "" {
			continue
		}
		if _, ok := stats[c.DefaultService]; !ok {
			stats[c.DefaultService] = make(map[string]tally)
		}
		entry, exists := stats[c.DefaultService][c.DefaultModel]
		if !exists {
			entry = tally{count: 0, minSort: c.SortOrder}
		}
		entry.count++
		if c.SortOrder < entry.minSort {
			entry.minSort = c.SortOrder
		}
		stats[c.DefaultService][c.DefaultModel] = entry
	}
	pickModel := func(svc string) string {
		best := ""
		var bestT tally
		for mdl, t := range stats[svc] {
			if best == "" ||
				t.count > bestT.count ||
				(t.count == bestT.count && t.minSort < bestT.minSort) {
				best, bestT = mdl, t
			}
		}
		return best
	}

	env := &storeData{
		SchemaVersion: CurrentSchemaVersion,
		Keys:          legacy.APIKeys,
	}

	// Migrate services into ServiceConfig (v0.2 adds DefaultModel field).
	for _, s := range legacy.Services {
		env.Services = append(env.Services, &ServiceConfig{
			ID:           s.ID,
			Name:         s.Name,
			LocalURL:     s.LocalURL,
			Enabled:      s.Enabled,
			ProxyEnabled: s.ProxyEnabled,
			DefaultModel: pickModel(s.ID),
		})
	}

	// Migrate clients: default_service → preferred_service, default_model → model_override.
	for _, c := range legacy.Clients {
		env.Clients = append(env.Clients, &Client{
			ID:               c.ID,
			Name:             c.Name,
			Token:            c.Token,
			PreferredService: c.DefaultService,
			ModelOverride:    c.DefaultModel,
			AllowedServices:  c.AllowedServices,
			AgentType:        c.AgentType,
			WorkDir:          c.WorkDir,
			IPWhitelist:      c.IPWhitelist,
			Avatar:           c.Avatar,
			Enabled:          c.Enabled,
			SortOrder:        c.SortOrder,
			CreatedAt:        c.CreatedAt,
		})
	}

	// Carry over theme/lang into Settings.
	if legacy.Theme != "" || legacy.Lang != "" {
		env.Settings = &StoreSettings{
			Theme: legacy.Theme,
			Lang:  legacy.Lang,
		}
	}

	return env, nil
}
