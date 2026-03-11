// Package hooks: OpenClaw 연동 훅 및 이벤트 시스템
package hooks

import (
	"encoding/json"
	"fmt"
	"net"
	"os/exec"
	"time"
)

// EventType: 훅 이벤트 종류
type EventType string

const (
	EventModelChanged  EventType = "model_changed"
	EventKeyExhausted  EventType = "key_exhausted"
	EventServiceDown   EventType = "service_down"
	EventDoctorFix     EventType = "doctor_fix"
	EventOllamaWaiting EventType = "ollama_waiting"
	EventOllamaDone    EventType = "ollama_done"
)

// Event: 훅 이벤트 데이터
type Event struct {
	Type      EventType         `json:"type"`
	Timestamp time.Time         `json:"timestamp"`
	Data      map[string]string `json:"data,omitempty"`
}

// Manager: 훅 관리자
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

// Fire: 이벤트 발생 (비동기)
func (m *Manager) Fire(evt EventType, data map[string]string) {
	go m.fire(evt, data)
}

func (m *Manager) fire(evt EventType, data map[string]string) {
	e := Event{
		Type:      evt,
		Timestamp: time.Now(),
		Data:      data,
	}

	// 1. 셸 명령 실행
	if cmd, ok := m.shellCmds[evt]; ok && cmd != "" {
		exec.Command("sh", "-c", cmd).Run() //nolint:errcheck
	}

	// 2. OpenClaw TUI 소켓 알림
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

// TUIFooter: OpenClaw TUI 푸터 메시지 전송 (짧은 상태 표시용)
// 예: "⏳ Ollama 대기 중..." 또는 "✅ 모델 변경됨: gemini-2.5-flash"
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

// DoctorFixGuard: doctor --fix가 설정을 덮어쓰지 않도록 보호
// OpenClaw doctor가 wall-vault 설정 파일을 수정하려 할 때 차단
func (m *Manager) DoctorFixGuard(configPath string) {
	// doctor fix 이벤트 수신 시 현재 설정 백업 후 복원
	m.fire(EventDoctorFix, map[string]string{
		"config": configPath,
		"action": "guard",
	})
}
