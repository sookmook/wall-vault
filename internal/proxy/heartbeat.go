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
	s.mu.RUnlock()

	sseConn := s.sse != nil && s.sse.IsConnected()

	// Collect last-used key ID per service
	activeKeys := make(map[string]string)
	for _, service := range []string{"google", "openrouter", "anthropic", "ollama"} {
		if id := s.keyMgr.LastUsedID(service); id != "" {
			activeKeys[service] = id
		}
	}

	payload := heartbeatPayload{
		ClientID:     s.cfg.Proxy.ClientID,
		Version:      Version,
		Service:      svc,
		Model:        mdl,
		SSE:          sseConn,
		Avatar:       readLocalAvatar(s.cfg.Proxy.Avatar),
		ActiveKeys:   activeKeys,
		KeyUsage:     s.keyMgr.UsageSnapshot(),
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
