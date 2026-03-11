package vault

import (
	"encoding/json"
	"fmt"
	"net/http"
	"sync"
)

// Broker: SSE 이벤트 브로드캐스터
type Broker struct {
	mu      sync.RWMutex
	clients map[chan string]struct{}
}

func NewBroker() *Broker {
	return &Broker{
		clients: make(map[chan string]struct{}),
	}
}

// Subscribe: 새 SSE 클라이언트 채널 등록
func (b *Broker) Subscribe() chan string {
	ch := make(chan string, 8)
	b.mu.Lock()
	b.clients[ch] = struct{}{}
	b.mu.Unlock()
	return ch
}

// Unsubscribe: 클라이언트 채널 해제
func (b *Broker) Unsubscribe(ch chan string) {
	b.mu.Lock()
	delete(b.clients, ch)
	b.mu.Unlock()
	close(ch)
}

// Broadcast: 모든 구독 클라이언트에 이벤트 전송
func (b *Broker) Broadcast(evt SSEEvent) {
	data, err := json.Marshal(evt)
	if err != nil {
		return
	}
	msg := fmt.Sprintf("data: %s\n\n", data)

	b.mu.RLock()
	defer b.mu.RUnlock()
	for ch := range b.clients {
		select {
		case ch <- msg:
		default:
			// 꽉 찬 채널은 스킵 (클라이언트가 느린 경우)
		}
	}
}

// Count: 연결된 클라이언트 수
func (b *Broker) Count() int {
	b.mu.RLock()
	defer b.mu.RUnlock()
	return len(b.clients)
}

// ServeHTTP: SSE 엔드포인트 핸들러
func (b *Broker) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	flusher, ok := w.(http.Flusher)
	if !ok {
		http.Error(w, "SSE 미지원", http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "text/event-stream")
	w.Header().Set("Cache-Control", "no-cache")
	w.Header().Set("Connection", "keep-alive")
	w.Header().Set("Access-Control-Allow-Origin", "*")

	ch := b.Subscribe()
	defer b.Unsubscribe(ch)

	// 연결 즉시 ping 전송
	fmt.Fprintf(w, "data: {\"type\":\"connected\",\"clients\":%d}\n\n", b.Count())
	flusher.Flush()

	ctx := r.Context()
	for {
		select {
		case <-ctx.Done():
			return
		case msg, ok := <-ch:
			if !ok {
				return
			}
			fmt.Fprint(w, msg)
			flusher.Flush()
		}
	}
}
