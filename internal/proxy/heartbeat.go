package proxy

import (
	"bytes"
	"encoding/json"
	"log"
	"net/http"
	"time"
)

type heartbeatPayload struct {
	ClientID string `json:"client_id"`
	Version  string `json:"version"`
	Service  string `json:"service"`
	Model    string `json:"model"`
	SSE      bool   `json:"sse_connected"`
	Host     string `json:"host,omitempty"`
}

// startHeartbeat: 60초마다 금고에 상태 전송 (비동기)
func (s *Server) startHeartbeat() {
	if s.cfg.Proxy.VaultURL == "" {
		return
	}
	go func() {
		// 최초 전송은 15초 후 (서비스가 뜰 시간)
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

	payload := heartbeatPayload{
		ClientID: s.cfg.Proxy.ClientID,
		Version:  "v0.1.2",
		Service:  svc,
		Model:    mdl,
		SSE:      sseConn,
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
