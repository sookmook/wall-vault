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
	"time"
)

// SSEClient: vault SSE stream subscription client
type SSEClient struct {
	mu              sync.RWMutex
	vaultURL        string
	clientID        string
	connected       bool
	onConfig        func(service, model string) // config change callback
	onKeyChange     func()                      // key added/deleted callback
	onUsageReset    func()                      // midnight daily-counter reset callback
	onServiceChange func([]string)              // proxy service list change callback
}

func NewSSEClient(vaultURL, clientID string, onConfig func(service, model string), onKeyChange func(), onUsageReset func(), onServiceChange func([]string)) *SSEClient {
	return &SSEClient{
		vaultURL:        vaultURL,
		clientID:        clientID,
		onConfig:        onConfig,
		onKeyChange:     onKeyChange,
		onUsageReset:    onUsageReset,
		onServiceChange: onServiceChange,
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
	}
}

func (c *SSEClient) connect() error {
	if c.vaultURL == "" {
		return fmt.Errorf("vault URL 없음")
	}
	url := c.vaultURL + "/api/events"
	resp, err := http.Get(url)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	c.mu.Lock()
	c.connected = true
	c.mu.Unlock()
	log.Printf("[SSE] ✅ vault connected: %s", url)

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
			ProxyServices []string `json:"proxy_services"`
		} `json:"data"`
	}
	if err := json.Unmarshal([]byte(data), &evt); err != nil {
		return
	}
	switch evt.Type {
	case "config_change":
		if evt.Data.ClientID == c.clientID || evt.Data.ClientID == "" {
			log.Printf("[SSE] 🔔 config change received: %s/%s", evt.Data.Service, evt.Data.Model)
			if c.onConfig != nil {
				c.onConfig(evt.Data.Service, evt.Data.Model)
			}
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
