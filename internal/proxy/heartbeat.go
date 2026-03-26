package proxy

import (
	"bytes"
	"encoding/base64"
	"encoding/json"
	"log"
	"net/http"
	"os"
	"path/filepath"
	"strings"
	"time"
)

type heartbeatPayload struct {
	ClientID      string            `json:"client_id"`
	Version       string            `json:"version"`
	Service       string            `json:"service"`
	Model         string            `json:"model"`
	SSE           bool              `json:"sse_connected"`
	Host          string            `json:"host,omitempty"`
	Avatar        string            `json:"avatar,omitempty"`         // base64 data URI of local avatar file
	ActiveKeys    map[string]string `json:"active_keys,omitempty"`    // service → key ID
	KeyUsage      map[string]int    `json:"key_usage,omitempty"`      // key ID → successful tokens today
	KeyAttempts   map[string]int    `json:"key_attempts,omitempty"`   // key ID → total requests today (including rate-limited)
	KeyCooldowns  map[string]string `json:"key_cooldowns,omitempty"`  // key ID → cooldown RFC3339
	ActiveClients []activeClientItem `json:"active_clients,omitempty"` // recently-served non-proxy clients
}

// activeClientItem: activity record for a client served through this proxy
type activeClientItem struct {
	ClientID string `json:"client_id"`
	Service  string `json:"service"`
	Model    string `json:"model"`
}

// readLocalAvatar reads the configured avatar file and returns a base64 data URI.
// avatarPath is relative to ~/.openclaw/ (e.g. "workspace/avatars/bot-a.png").
// Falls back to workspace/avatar.png if avatarPath is empty.
func readLocalAvatar(avatarPath string) string {
	home, err := os.UserHomeDir()
	if err != nil {
		return ""
	}
	relPath := avatarPath
	if relPath == "" {
		relPath = filepath.Join("workspace", "avatar.png")
	}
	data, err := os.ReadFile(filepath.Join(home, ".openclaw", relPath))
	if err != nil {
		return ""
	}
	mime := "image/png"
	switch strings.ToLower(filepath.Ext(relPath)) {
	case ".jpg", ".jpeg", ".hpg":
		mime = "image/jpeg"
	case ".webp":
		mime = "image/webp"
	case ".gif":
		mime = "image/gif"
	}
	return "data:" + mime + ";base64," + base64.StdEncoding.EncodeToString(data)
}

// startHeartbeat: send status to vault every 60 seconds (async)
func (s *Server) startHeartbeat() {
	if s.cfg.Proxy.VaultURL == "" {
		return
	}
	go func() {
		// first send after 5 seconds (allow service to start up)
		time.Sleep(5 * time.Second)
		s.sendHeartbeat()

		ticker := time.NewTicker(20 * time.Second)
		defer ticker.Stop()
		for range ticker.C {
			s.sendHeartbeat()
		}
	}()
}

func (s *Server) sendHeartbeat() {
	s.mu.RLock()
	svc := s.service
	mdl := s.model
	// Report the actual service/model that last handled a request,
	// so the dashboard shows what's really being used (may be a fallback).
	if s.lastActualSvc != "" {
		svc = s.lastActualSvc
		mdl = s.lastActualMdl
	}
	s.mu.RUnlock()

	sseConn := s.sse != nil && s.sse.IsConnected()

	// Collect last-used key ID per service
	activeKeys := make(map[string]string)
	for _, service := range []string{"google", "openrouter", "anthropic", "ollama", "openai", "lmstudio", "vllm"} {
		if id := s.keyMgr.LastUsedID(service); id != "" {
			activeKeys[service] = id
		}
	}

	// Collect recently-active non-proxy clients.
	// Applied entries get a longer grace period (3 min) so newly-applied
	// agents stay visible on the dashboard even before the user starts using them.
	// Regular (request-based) entries time out after 90s of inactivity.
	var activeClients []activeClientItem
	now := time.Now()
	cutoff := now.Add(-90 * time.Second)
	appliedCutoff := now.Add(-3 * time.Minute)
	s.clientActMu.Lock()
	for cid, act := range s.clientActs {
		alive := act.lastSeen.After(cutoff)
		if !alive && act.applied {
			alive = act.lastSeen.After(appliedCutoff)
		}
		if alive {
			activeClients = append(activeClients, activeClientItem{
				ClientID: cid,
				Service:  act.service,
				Model:    act.model,
			})
		} else {
			delete(s.clientActs, cid) // evict stale entries
		}
	}
	s.clientActMu.Unlock()

	payload := heartbeatPayload{
		ClientID:      s.cfg.Proxy.ClientID,
		Version:       Version,
		Service:       svc,
		Model:         mdl,
		SSE:           sseConn,
		Avatar:        readLocalAvatar(s.cfg.Proxy.Avatar),
		ActiveKeys:    activeKeys,
		ActiveClients: activeClients,
		KeyUsage:      s.keyMgr.UsageSnapshot(),
		KeyAttempts:  s.keyMgr.AttemptsSnapshot(),
		KeyCooldowns: s.keyMgr.CooldownSnapshot(),
	}

	body, _ := json.Marshal(payload)
	req, err := http.NewRequest("POST", s.cfg.Proxy.VaultURL+"/api/heartbeat", bytes.NewReader(body))
	if err != nil {
		return
	}
	req.Header.Set("Content-Type", "application/json")
	if s.cfg.Proxy.VaultToken != "" {
		req.Header.Set("Authorization", "Bearer "+s.cfg.Proxy.VaultToken)
	}

	client := &http.Client{Timeout: 5 * time.Second}
	resp, err := client.Do(req)
	if err != nil {
		log.Printf("[heartbeat] 전송 실패: %v", err)
		return
	}
	resp.Body.Close()
}
