package proxy

import (
	"bytes"
	"encoding/base64"
	"encoding/json"
	"log"
	"net/http"
	"os"
	"os/exec"
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
	AgentAlive    *bool             `json:"agent_alive,omitempty"`    // local agent process is running (nanoclaw/openclaw)
}

// activeClientItem: activity record for a client served through this proxy
type activeClientItem struct {
	ClientID string `json:"client_id"`
	Service  string `json:"service"`
	Model    string `json:"model"`
}

// readLocalAvatar reads the configured avatar file and returns a base64 data URI.
// avatarPath is relative to ~/.openclaw/ (e.g. "workspace/avatars/<client-id>.png").
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
	// path traversal guard: reject absolute paths and ".." components
	if strings.HasPrefix(relPath, "/") || strings.Contains(relPath, "..") {
		return ""
	}
	cleaned := filepath.Clean(relPath)
	if strings.HasPrefix(cleaned, "..") || filepath.IsAbs(cleaned) {
		return ""
	}
	baseDir := filepath.Join(home, ".openclaw")
	fullPath := filepath.Join(baseDir, cleaned)
	// verify the resolved path is still under the base directory
	if !strings.HasPrefix(fullPath, baseDir+string(filepath.Separator)) {
		return ""
	}
	data, err := os.ReadFile(fullPath)
	if err != nil {
		return ""
	}
	mime := "image/png"
	switch strings.ToLower(filepath.Ext(cleaned)) {
	case ".jpg", ".jpeg":
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
		// Phase-shift the heartbeat cadence by a client-specific offset
		// so proxies booted within the same second do not send synchronised
		// bursts to the vault. Bounded by the 20-second heartbeat period.
		const heartbeatPeriodMs = 20 * 1000
		if offset := AgentOffset(s.cfg.Proxy.ClientID, heartbeatPeriodMs); offset > 0 {
			select {
			case <-time.After(offset):
			case <-s.stopCh:
				return
			}
		}

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
	// Always report the user's configured service/model.
	// Fallback is transient — showing it in the dashboard confuses users
	// (e.g. "LM Studio configured but dashboard says Ollama").
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

	// Emit every co-hosted agent (Client.Host == os.Hostname()) that also
	// passes a per-agent-type liveness probe. syncFromVault curates which
	// clients we *may* claim (Host match); detectClientAlive decides which
	// of those are *actually running right now*. Without the second check
	// the dashboard would show green for an agent whose host is up but
	// whose process isn't (e.g. VSCode closed → cline not running).
	already := make(map[string]bool, len(activeClients))
	for _, ac := range activeClients {
		already[ac.ClientID] = true
	}
	s.mu.RLock()
	hosted := make([]hostAgent, len(s.hostAgents))
	copy(hosted, s.hostAgents)
	s.mu.RUnlock()
	for _, ha := range hosted {
		if already[ha.ClientID] {
			continue
		}
		if !detectClientAlive(ha.AgentType) {
			continue
		}
		activeClients = append(activeClients, activeClientItem{
			ClientID: ha.ClientID,
			Service:  ha.Service,
			Model:    ha.Model,
		})
	}

	// detect local agent process health
	s.mu.RLock()
	ownAgentType := s.ownAgentType
	s.mu.RUnlock()
	agentAlive := detectAgentProcess(ownAgentType)

	payload := heartbeatPayload{
		ClientID:      s.cfg.Proxy.ClientID,
		Version:       Version,
		Service:       svc,
		Model:         mdl,
		SSE:           sseConn,
		Avatar:        readLocalAvatar(s.cfg.Proxy.Avatar),
		ActiveKeys:    activeKeys,
		ActiveClients: activeClients,
		AgentAlive:    agentAlive,
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

// detectClientAlive reports whether a process consistent with the given
// agent_type is running on this host. Used to gate which co-hosted clients
// (those whose Client.Host matches os.Hostname()) get reported as ACTIVE
// in the heartbeat — host membership decides who *may* be claimed, this
// function decides who *actually is up right now*.
//
// Detection per type:
//   - claude-code   pgrep -x claude              (Claude Code CLI binary)
//   - cline         pgrep -x code                (VSCode binary; Cline is an
//                                                 extension and can't run
//                                                 without VSCode)
//   - openclaw      pgrep -f openclaw-gateway
//   - nanoclaw      systemctl --user is-active nanoclaw
//   - econoworld    always false                 (self-reports via its own
//                                                 heartbeat — should not
//                                                 appear in hostAgents)
//   - other         false                        (be honest — don't fake
//                                                 green for unknown types)
//
// Multiple claude-code clients sharing one Host all match the same pgrep,
// so a single running CLI lights all of them up. Disambiguating across
// OS namespaces (e.g. Windows-side claude-code visible from a WSL proxy)
// would need cwd or interop probes — out of scope for this iteration.
func detectClientAlive(agentType string) bool {
	switch agentType {
	case "claude-code":
		return exec.Command("pgrep", "-x", "claude").Run() == nil
	case "cline":
		return exec.Command("pgrep", "-x", "code").Run() == nil
	case "openclaw":
		return exec.Command("pgrep", "-f", "openclaw-gateway").Run() == nil
	case "nanoclaw":
		return exec.Command("systemctl", "--user", "is-active", "--quiet", "nanoclaw").Run() == nil
	}
	return false
}

// detectAgentProcess checks if the local agent process matching the given
// agent_type is alive. Returns nil if agent_type has no process to check,
// or a *bool indicating alive/dead.
func detectAgentProcess(agentType string) *bool {
	alive := true
	dead := false
	switch agentType {
	case "nanoclaw":
		// systemctl --user is-active nanoclaw (exit 0 = active)
		if err := exec.Command("systemctl", "--user", "is-active", "--quiet", "nanoclaw").Run(); err != nil {
			return &dead
		}
		return &alive
	case "openclaw":
		if err := exec.Command("pgrep", "-f", "openclaw-gateway").Run(); err != nil {
			return &dead
		}
		return &alive
	}
	// agent types without a detectable process (cursor, cline, etc.)
	return nil
}
