// Package hooks: OpenClaw integration hooks and event system
package hooks

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"log"
	"net"
	"os"
	"os/exec"
	"strings"
	"time"
)

// hookTimeout resolves the per-invocation shell command deadline. Defaults to
// 30s but can be stretched via WV_HOOK_TIMEOUT (e.g. "2m") when a legitimate
// hook runs a slow backup or webhook with external dependencies.
func hookTimeout() time.Duration {
	if raw := os.Getenv("WV_HOOK_TIMEOUT"); raw != "" {
		if d, err := time.ParseDuration(strings.TrimSpace(raw)); err == nil && d > 0 {
			return d
		}
	}
	return 30 * time.Second
}

// EventType: hook event type
type EventType string

const (
	EventModelChanged  EventType = "model_changed"
	EventKeyExhausted  EventType = "key_exhausted"
	EventServiceDown   EventType = "service_down"
	EventDoctorFix     EventType = "doctor_fix"
	EventOllamaWaiting EventType = "ollama_waiting"
	EventOllamaDone    EventType = "ollama_done"
)

// Event: hook event data
type Event struct {
	Type      EventType         `json:"type"`
	Timestamp time.Time         `json:"timestamp"`
	Data      map[string]string `json:"data,omitempty"`
}

// Manager: hook manager
type Manager struct {
	shellCmds    map[EventType]string
	openClawSock string
}

func NewManager(shellCmds map[EventType]string, sockPath string) *Manager {
	return &Manager{
		shellCmds:    shellCmds,
		openClawSock: sockPath,
	}
}

// Fire: fire event (async)
func (m *Manager) Fire(evt EventType, data map[string]string) {
	go m.fire(evt, data)
}

func (m *Manager) fire(evt EventType, data map[string]string) {
	e := Event{
		Type:      evt,
		Timestamp: time.Now(),
		Data:      data,
	}

	// 1. execute shell command with timeout + stdout/stderr capture so a
	// failing hook doesn't vanish into /dev/null. We log a truncated tail on
	// failure (full output can be huge and contain secrets).
	if cmd, ok := m.shellCmds[evt]; ok && cmd != "" {
		timeout := hookTimeout()
		ctx, cancel := context.WithTimeout(context.Background(), timeout)
		defer cancel()
		var outBuf, errBuf bytes.Buffer
		c := exec.CommandContext(ctx, "sh", "-c", cmd)
		c.Stdout = &outBuf
		c.Stderr = &errBuf
		start := time.Now()
		runErr := c.Run()
		elapsed := time.Since(start)
		if runErr != nil {
			log.Printf("[hooks] %s command failed after %s: %v | stdout=%q stderr=%q",
				evt, elapsed.Round(time.Millisecond), runErr,
				truncate(outBuf.String(), 400), truncate(errBuf.String(), 400))
		}
	}

	// 2. notify OpenClaw TUI socket
	if m.openClawSock != "" {
		m.notifySocket(e)
	}
}

func (m *Manager) notifySocket(e Event) {
	conn, err := net.DialTimeout("unix", m.openClawSock, 500*time.Millisecond)
	if err != nil {
		return
	}
	defer conn.Close()

	data, _ := json.Marshal(e)
	fmt.Fprintf(conn, "%s\n", data)
}

// TUIFooter: send footer message to OpenClaw TUI (for brief status display)
// e.g. "⏳ Ollama waiting..." or "✅ model changed: gemini-2.5-flash"
func (m *Manager) TUIFooter(msg string) {
	if m.openClawSock == "" {
		return
	}
	m.notifySocket(Event{
		Type:      "tui_footer",
		Timestamp: time.Now(),
		Data:      map[string]string{"message": msg},
	})
}

// DoctorFixGuard: protect against doctor --fix overwriting config
// blocks OpenClaw doctor from modifying wall-vault config files
func (m *Manager) DoctorFixGuard(configPath string) {
	// on doctor fix event: backup and restore current config
	m.fire(EventDoctorFix, map[string]string{
		"config": configPath,
		"action": "guard",
	})
}

// truncate returns s if shorter than n, otherwise s[:n] + "…". Used to keep
// hook stdout/stderr log lines readable when a hook spews many kilobytes.
func truncate(s string, n int) string {
	if len(s) <= n {
		return s
	}
	return s[:n] + "…"
}
