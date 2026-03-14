package proxy

import (
	"bytes"
	"encoding/json"
	"log"
	"net/http"
	"time"
)

type heartbeatPayload struct {
	ClientID      string            `json:"client_id"`
	Version       string            `json:"version"`
	Service       string            `json:"service"`
	Model         string            `json:"model"`
	SSE           bool              `json:"sse_connected"`
	Host          string            `json:"host,omitempty"`
	ActiveKeys    map[string]string `json:"active_keys,omitempty"`    // service → key ID
	KeyUsage      map[string]int    `json:"key_usage,omitempty"`      // key ID → tokens used today
	KeyCooldowns  map[string]string `json:"key_cooldowns,omitempty"`  // key ID → cooldown RFC3339
}

// startHeartbeat: send status to vault every 60 seconds (async)
func (s *Server) startHeartbeat() {
	if s.cfg.Proxy.VaultURL == "" {
		return
	}
	go func() {
		// first send after 15 seconds (allow service to start up)
		time.Sleep(15 * time.Second)
		s.sendHeartbeat()

		ticker := time.NewTicker(60 * time.Second)
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
		Version:      "v0.1.4",
		Service:      svc,
		Model:        mdl,
		SSE:          sseConn,
		ActiveKeys:   activeKeys,
		KeyUsage:     s.keyMgr.UsageSnapshot(),
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
