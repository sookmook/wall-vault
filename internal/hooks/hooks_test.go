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
	// 파일 존재 여부로 셸 명령 실행 확인
	tmpFile := t.TempDir() + "/hooks_test_flag"

	m := NewManager(map[EventType]string{
		EventModelChanged: "touch " + tmpFile,
	}, "")

	m.Fire(EventModelChanged, map[string]string{"model": "gemini-2.5-flash"})

	// 비동기 실행이므로 짧게 대기
	time.Sleep(200 * time.Millisecond)

	if _, err := os.Stat(tmpFile); os.IsNotExist(err) {
		t.Error("셸 명령이 실행되지 않음 (파일 없음)")
	}
}

func TestFire_NoCommand(t *testing.T) {
	// 등록되지 않은 이벤트 — 패닉 없이 처리
	m := NewManager(map[EventType]string{}, "")
	m.Fire(EventKeyExhausted, nil)
	time.Sleep(50 * time.Millisecond)
}

func TestFire_EmptyCommand(t *testing.T) {
	// 빈 문자열 명령 — 실행 안 함
	m := NewManager(map[EventType]string{
		EventServiceDown: "",
	}, "")
	m.Fire(EventServiceDown, nil)
	time.Sleep(50 * time.Millisecond)
}

func TestNotifySocket(t *testing.T) {
	// Unix 소켓 서버 시작
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
		// JSON 파싱 검증
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
	// 소켓 없이 호출 — 패닉 없어야 함
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
