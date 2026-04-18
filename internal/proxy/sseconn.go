package proxy

import (
	"bufio"
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net/http"
	"strings"
	"sync"
	"sync/atomic"
	"time"
)

// SSEClient: vault SSE stream subscription client
type SSEClient struct {
	mu              sync.RWMutex
	vaultURL        string
	clientID        string
	token           string // vault token for SSE authentication
	connected       bool
	onConfig        func(service, model string)              // config change callback (own client only)
	onAnyConfig     func(clientID, agentType, service, model string) // config change callback (any other client)
	onConfigFlush   func()                                   // flush token cache on any config_change event
	onKeyChange     func()                                   // key added/deleted callback
	onUsageReset    func()                                   // midnight daily-counter reset callback
	onServiceChange func([]string)                           // proxy service list change callback
	onReconnect     func()                                   // called after SSE reconnect to re-sync state

	// reconnectInFlight ensures at most one onReconnect handler runs at a time.
	// Without this, a vault that keeps dropping the SSE stream would spawn a
	// fresh sync goroutine on every reconnect, piling up indefinitely when the
	// sync itself is slow or stuck on a hung HTTP call.
	reconnectInFlight atomic.Bool
}

func NewSSEClient(vaultURL, clientID, token string, onConfig func(service, model string), onAnyConfig func(clientID, agentType, service, model string), onConfigFlush func(), onKeyChange func(), onUsageReset func(), onServiceChange func([]string), onReconnect func()) *SSEClient {
	return &SSEClient{
		vaultURL:        vaultURL,
		clientID:        clientID,
		token:           token,
		onConfig:        onConfig,
		onAnyConfig:     onAnyConfig,
		onConfigFlush:   onConfigFlush,
		onKeyChange:     onKeyChange,
		onUsageReset:    onUsageReset,
		onServiceChange: onServiceChange,
		onReconnect:     onReconnect,
	}
}

// Start: start background SSE connection (auto-reconnect)
func (c *SSEClient) Start() {
	go c.loop()
}

func (c *SSEClient) IsConnected() bool {
	c.mu.RLock()
	defer c.mu.RUnlock()
	return c.connected
}

func (c *SSEClient) loop() {
	backoff := time.Second
	first := true
	for {
		err := c.connect()
		c.mu.Lock()
		c.connected = false
		c.mu.Unlock()
		if err != nil && err != io.EOF {
			log.Printf("[SSE] connection error: %v — retrying in %v", err, backoff)
		}
		time.Sleep(backoff)
		if backoff < 30*time.Second {
			backoff *= 2
		}
		// On reconnect (not first connect): immediately re-sync model/config from vault
		// to pick up any changes that occurred while the SSE connection was down.
		// Throttled so a flapping vault can't pile up concurrent syncs.
		if !first && c.onReconnect != nil {
			if c.reconnectInFlight.CompareAndSwap(false, true) {
				go func() {
					defer c.reconnectInFlight.Store(false)
					c.onReconnect()
				}()
			} else {
				log.Printf("[SSE] skip onReconnect — previous sync still running")
			}
		}
		first = false
	}
}

func (c *SSEClient) connect() error {
	if c.vaultURL == "" {
		return fmt.Errorf("vault URL 없음")
	}
	sseURL := c.vaultURL + "/api/events"
	req, err := http.NewRequest("GET", sseURL, nil)
	if err != nil {
		return err
	}
	// Authenticate with the vault token (required when admin_token is configured)
	if c.token != "" {
		req.Header.Set("Authorization", "Bearer "+c.token)
	}
	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return err
	}
	defer resp.Body.Close()
	if resp.StatusCode == http.StatusUnauthorized {
		return fmt.Errorf("SSE auth failed (401) — check vault_token config")
	}

	c.mu.Lock()
	c.connected = true
	c.mu.Unlock()
	log.Printf("[SSE] ✅ vault connected: %s", sseURL)

	scanner := bufio.NewScanner(resp.Body)
	for scanner.Scan() {
		line := scanner.Text()
		if !strings.HasPrefix(line, "data: ") {
			continue
		}
		data := strings.TrimPrefix(line, "data: ")
		c.handleEvent(data)
	}
	return scanner.Err()
}

func (c *SSEClient) handleEvent(data string) {
	var evt struct {
		Type string `json:"type"`
		Data struct {
			ClientID      string   `json:"client_id"`
			Service       string   `json:"service"`
			Model         string   `json:"model"`
			AgentType     string   `json:"agent_type"`
			ProxyServices []string `json:"proxy_services"`
		} `json:"data"`
	}
	if err := json.Unmarshal([]byte(data), &evt); err != nil {
		return
	}
	switch evt.Type {
	case "config_change":
		// Always flush token cache so model changes take effect within one request
		if c.onConfigFlush != nil {
			c.onConfigFlush()
		}
		if evt.Data.ClientID == c.clientID || evt.Data.ClientID == "" {
			log.Printf("[SSE] 🔔 config change received: %s/%s", evt.Data.Service, evt.Data.Model)
			if c.onConfig != nil {
				c.onConfig(evt.Data.Service, evt.Data.Model)
			}
		} else if c.onAnyConfig != nil {
			log.Printf("[SSE] 🔔 foreign config change: %s(%s) → %s/%s", evt.Data.ClientID, evt.Data.AgentType, evt.Data.Service, evt.Data.Model)
			c.onAnyConfig(evt.Data.ClientID, evt.Data.AgentType, evt.Data.Service, evt.Data.Model)
		}
	case "key_added", "key_deleted":
		log.Printf("[SSE] 🔑 key event: %s — re-syncing keys", evt.Type)
		if c.onKeyChange != nil {
			go c.onKeyChange()
		}
	case "usage_reset":
		log.Printf("[SSE] 🌅 usage_reset — resetting local counters and re-syncing")
		if c.onUsageReset != nil {
			go c.onUsageReset()
		}
	case "service_changed":
		if evt.Data.ProxyServices != nil && c.onServiceChange != nil {
			log.Printf("[SSE] 🔧 service changed: proxy_services=%v", evt.Data.ProxyServices)
			go c.onServiceChange(evt.Data.ProxyServices)
		}
	case "connected":
		log.Printf("[SSE] connection confirmed")
	}
}
