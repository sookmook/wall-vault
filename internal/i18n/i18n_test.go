package i18n

import (
	"testing"
)

func TestT_Korean(t *testing.T) {
	SetLang("ko")
	if got := T("health_ok"); got != "정상" {
		t.Errorf("ko health_ok = %q, want %q", got, "정상")
	}
}

func TestT_English(t *testing.T) {
	SetLang("en")
	if got := T("health_ok"); got != "ok" {
		t.Errorf("en health_ok = %q, want %q", got, "ok")
	}
}

func TestT_FallbackToEnglish(t *testing.T) {
	// 지원하지 않는 언어 → 영어 폴백
	SetLang("xx")
	if got := T("health_ok"); got != "ok" {
		t.Errorf("fallback health_ok = %q, want %q", got, "ok")
	}
}

func TestT_MissingKey(t *testing.T) {
	SetLang("ko")
	// 존재하지 않는 키 → 키 이름 반환
	if got := T("__nonexistent_key__"); got != "__nonexistent_key__" {
		t.Errorf("missing key = %q, want key itself", got)
	}
}

func TestSetLang_LocaleString(t *testing.T) {
	cases := []struct {
		input string
		want  string
	}{
		{"ko_KR.UTF-8", "ko"},
		{"en_US.UTF-8", "en"},
		{"zh_CN.UTF-8", "zh"},
		{"ja_JP.UTF-8", "ja"},
		{"de_DE.UTF-8", "de"},
		{"fr_FR.UTF-8", "fr"},
		{"es_ES.UTF-8", "es"},
		{"pt_BR.UTF-8", "pt"},
		{"hi_IN.UTF-8", "hi"},
		{"ar_SA.UTF-8", "ar"},
	}
	for _, c := range cases {
		SetLang(c.input)
		if got := Lang(); got != c.want {
			t.Errorf("SetLang(%q) = %q, want %q", c.input, got, c.want)
		}
	}
}

func TestSupported(t *testing.T) {
	if len(Supported) != 10 {
		t.Errorf("Supported 언어 수 = %d, want 10", len(Supported))
	}
	// 모든 지원 언어에 setup_welcome 키가 있는지 확인
	for _, code := range Supported {
		SetLang(code)
		msg := T("setup_welcome")
		if msg == "setup_welcome" {
			t.Errorf("언어 %s: setup_welcome 번역 없음", code)
		}
	}
}

func TestAllLanguagesHaveAllKeys(t *testing.T) {
	// 영어 키 목록을 기준으로 다른 언어 누락 키 검사
	enKeys := make([]string, 0, len(messages["en"]))
	for k := range messages["en"] {
		enKeys = append(enKeys, k)
	}

	for lang, m := range messages {
		if lang == "en" {
			continue
		}
		for _, key := range enKeys {
			if _, ok := m[key]; !ok {
				t.Errorf("언어 %s: 키 %q 누락", lang, key)
			}
		}
	}
}

func TestSetupMessages(t *testing.T) {
	langs := []string{"ko", "en", "ja", "zh", "de"}
	for _, l := range langs {
		SetLang(l)
		welcome := T("setup_welcome")
		done := T("setup_done")
		if welcome == "" || welcome == "setup_welcome" {
			t.Errorf("언어 %s: setup_welcome 비어있음", l)
		}
		if done == "" || done == "setup_done" {
			t.Errorf("언어 %s: setup_done 비어있음", l)
		}
	}
}
