// Package hooks: OpenClaw integration hooks and event system
package hooks

import (
	"encoding/json"
	"fmt"
	"net"
	"os/exec"
	"time"
)

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

	// 1. execute shell command
	if cmd, ok := m.shellCmds[evt]; ok && cmd != "" {
		exec.Command("sh", "-c", cmd).Run() //nolint:errcheck
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
