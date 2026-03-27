package vault

import (
	"encoding/json"
	"fmt"
	"net/http"
	"sync"
)

// Broker: SSE event broadcaster
type Broker struct {
	mu        sync.RWMutex
	clients   map[chan string]struct{}
	OnConnect func() // called (in a goroutine) after a new SSE client connects
}

func NewBroker() *Broker {
	return &Broker{
		clients: make(map[chan string]struct{}),
	}
}

// Subscribe: register a new SSE client channel
func (b *Broker) Subscribe() chan string {
	ch := make(chan string, 8)
	b.mu.Lock()
	b.clients[ch] = struct{}{}
	b.mu.Unlock()
	return ch
}

// Unsubscribe: deregister a client channel
func (b *Broker) Unsubscribe(ch chan string) {
	b.mu.Lock()
	delete(b.clients, ch)
	b.mu.Unlock()
	close(ch)
}

// Broadcast: send event to all subscribed clients
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
			// skip full channels (slow client)
		}
	}
}

// Count: number of connected clients
func (b *Broker) Count() int {
	b.mu.RLock()
	defer b.mu.RUnlock()
	return len(b.clients)
}

// ServeHTTP: SSE endpoint handler
func (b *Broker) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	flusher, ok := w.(http.Flusher)
	if !ok {
		http.Error(w, "SSE 미지원", http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "text/event-stream")
	w.Header().Set("Cache-Control", "no-cache")
	w.Header().Set("Connection", "keep-alive")

	ch := b.Subscribe()
	defer b.Unsubscribe(ch)

	// send ping immediately on connect
	fmt.Fprintf(w, "data: {\"type\":\"connected\",\"clients\":%d}\n\n", b.Count())
	flusher.Flush()

	// broadcast full state so newly connected / reconnected clients are in sync
	if b.OnConnect != nil {
		go b.OnConnect()
	}

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
