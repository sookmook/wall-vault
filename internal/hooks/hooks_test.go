package hooks

import (
	"encoding/json"
	"net"
	"os"
	"strings"
	"testing"
	"time"
)

func TestFire_ShellCommand(t *testing.T) {
	// verify shell command execution by checking file existence
	tmpFile := t.TempDir() + "/hooks_test_flag"

	m := NewManager(map[EventType]string{
		EventModelChanged: "touch " + tmpFile,
	}, "")

	m.Fire(EventModelChanged, map[string]string{"model": "gemini-2.5-flash"})

	// brief wait since execution is async
	time.Sleep(200 * time.Millisecond)

	if _, err := os.Stat(tmpFile); os.IsNotExist(err) {
		t.Error("셸 명령이 실행되지 않음 (파일 없음)")
	}
}

func TestFire_NoCommand(t *testing.T) {
	// unregistered event — handled without panic
	m := NewManager(map[EventType]string{}, "")
	m.Fire(EventKeyExhausted, nil)
	time.Sleep(50 * time.Millisecond)
}

func TestFire_EmptyCommand(t *testing.T) {
	// empty string command — not executed
	m := NewManager(map[EventType]string{
		EventServiceDown: "",
	}, "")
	m.Fire(EventServiceDown, nil)
	time.Sleep(50 * time.Millisecond)
}

func TestNotifySocket(t *testing.T) {
	// start Unix socket server
	sockPath := t.TempDir() + "/test.sock"
	ln, err := net.Listen("unix", sockPath)
	if err != nil {
		t.Fatalf("소켓 생성 실패: %v", err)
	}
	defer ln.Close()

	received := make(chan string, 1)
	go func() {
		conn, err := ln.Accept()
		if err != nil {
			return
		}
		defer conn.Close()
		buf := make([]byte, 4096)
		n, _ := conn.Read(buf)
		received <- string(buf[:n])
	}()

	m := NewManager(map[EventType]string{}, sockPath)
	m.Fire(EventModelChanged, map[string]string{"model": "test-model"})

	select {
	case msg := <-received:
		// validate JSON parsing
		var evt Event
		if err := json.Unmarshal([]byte(strings.TrimSpace(msg)), &evt); err != nil {
			t.Fatalf("소켓 메시지 파싱 실패: %v — msg: %q", err, msg)
		}
		if evt.Type != EventModelChanged {
			t.Errorf("type = %q, want %q", evt.Type, EventModelChanged)
		}
		if evt.Data["model"] != "test-model" {
			t.Errorf("data.model = %q", evt.Data["model"])
		}
	case <-time.After(time.Second):
		t.Error("소켓 메시지 수신 타임아웃")
	}
}

func TestTUIFooter(t *testing.T) {
	sockPath := t.TempDir() + "/tui.sock"
	ln, err := net.Listen("unix", sockPath)
	if err != nil {
		t.Fatalf("소켓 생성 실패: %v", err)
	}
	defer ln.Close()

	received := make(chan string, 1)
	go func() {
		conn, err := ln.Accept()
		if err != nil {
			return
		}
		defer conn.Close()
		buf := make([]byte, 4096)
		n, _ := conn.Read(buf)
		received <- string(buf[:n])
	}()

	m := NewManager(map[EventType]string{}, sockPath)
	m.TUIFooter("⏳ Ollama 대기 중...")

	select {
	case msg := <-received:
		var evt Event
		json.Unmarshal([]byte(strings.TrimSpace(msg)), &evt)
		if evt.Data["message"] != "⏳ Ollama 대기 중..." {
			t.Errorf("TUIFooter 메시지 = %q", evt.Data["message"])
		}
	case <-time.After(time.Second):
		t.Error("TUIFooter 타임아웃")
	}
}

func TestTUIFooter_NoSocket(t *testing.T) {
	// called without socket — must not panic
	m := NewManager(map[EventType]string{}, "")
	m.TUIFooter("테스트")
}

func TestEventTypes(t *testing.T) {
	types := []EventType{
		EventModelChanged,
		EventKeyExhausted,
		EventServiceDown,
		EventDoctorFix,
		EventOllamaWaiting,
		EventOllamaDone,
	}
	for _, et := range types {
		if string(et) == "" {
			t.Errorf("빈 EventType 있음")
		}
	}
}
