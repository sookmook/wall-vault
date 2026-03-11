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

// SSEClient: 금고 SSE 스트림 구독 클라이언트
type SSEClient struct {
	mu        sync.RWMutex
	vaultURL  string
	clientID  string
	connected bool
	onConfig  func(service, model string) // 설정 변경 콜백
}

func NewSSEClient(vaultURL, clientID string, onConfig func(service, model string)) *SSEClient {
	return &SSEClient{
		vaultURL: vaultURL,
		clientID: clientID,
		onConfig: onConfig,
	}
}

// Start: 백그라운드 SSE 연결 시작 (자동 재연결)
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
			log.Printf("[SSE] 연결 오류: %v — %v 후 재시도", err, backoff)
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
	log.Printf("[SSE] ✅ 금고 연결됨: %s", url)

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
			ClientID string `json:"client_id"`
			Service  string `json:"service"`
			Model    string `json:"model"`
		} `json:"data"`
	}
	if err := json.Unmarshal([]byte(data), &evt); err != nil {
		return
	}
	switch evt.Type {
	case "config_change":
		if evt.Data.ClientID == c.clientID || evt.Data.ClientID == "" {
			log.Printf("[SSE] 🔔 설정 변경 수신: %s/%s", evt.Data.Service, evt.Data.Model)
			if c.onConfig != nil {
				c.onConfig(evt.Data.Service, evt.Data.Model)
			}
		}
	case "connected":
		log.Printf("[SSE] 연결 확인됨")
	}
}
