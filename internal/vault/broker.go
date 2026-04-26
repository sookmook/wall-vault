package vault

import (
	"encoding/json"
	"fmt"
	"net/http"
	"sync"
	"sync/atomic"
)

// Broker: SSE event broadcaster
type Broker struct {
	mu        sync.RWMutex
	clients   map[chan string]struct{}
	OnConnect func() // called (in a goroutine) after a new SSE client connects

	// droppedEvents counts events that had to be dropped because a subscriber's
	// channel was full. Surfaced via /api/status so operators can spot slow
	// clients instead of silently losing agents_sync / config_change updates.
	droppedEvents atomic.Int64
}

// subscribeBufferSize is the per-subscriber channel buffer. 8 turned out to be
// too tight for the 15s agents_sync cadence + burstier config_change bursts —
// slow tabs dropped events silently. 64 keeps roughly a minute of peak traffic
// in memory per subscriber; memory footprint is negligible for the few tabs we
// ever have open.
const subscribeBufferSize = 64

func NewBroker() *Broker {
	return &Broker{
		clients: make(map[chan string]struct{}),
	}
}

// Subscribe: register a new SSE client channel
func (b *Broker) Subscribe() chan string {
	ch := make(chan string, subscribeBufferSize)
	b.mu.Lock()
	b.clients[ch] = struct{}{}
	b.mu.Unlock()
	return ch
}

// DroppedEvents returns how many SSE events have been dropped because a
// subscriber's buffer was full.
func (b *Broker) DroppedEvents() int64 {
	return b.droppedEvents.Load()
}

// Unsubscribe: deregister a client channel
func (b *Broker) Unsubscribe(ch chan string) {
	b.mu.Lock()
	delete(b.clients, ch)
	b.mu.Unlock()
	close(ch)
}

// Broadcast: send event to all subscribed clients.
//
// We hold the read lock for the entire send loop. The previous snapshot-then-
// send pattern released the lock before iterating, which let Unsubscribe close
// a channel that Broadcast was about to write to — sending on a closed channel
// panics. Each iteration is non-blocking thanks to the default branch, so
// holding RLock here cannot deadlock or stall a slow subscriber.
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
			// slow subscriber — count the drop so it shows up in /api/status
			b.droppedEvents.Add(1)
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
