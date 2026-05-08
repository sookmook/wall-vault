package proxy

import (
	"context"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"
)

// fakeOllamaServer returns an httptest.Server whose /api/tags responds with
// the given model id list. Other paths return 404. Used by the pool tests
// so we don't depend on a live Ollama install.
func fakeOllamaServer(models ...string) *httptest.Server {
	mux := http.NewServeMux()
	mux.HandleFunc("/api/tags", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(200)
		w.Write([]byte(`{"models":[`))
		for i, m := range models {
			if i > 0 {
				w.Write([]byte(","))
			}
			w.Write([]byte(`{"name":"` + m + `"}`))
		}
		w.Write([]byte(`]}`))
	})
	return httptest.NewServer(mux)
}

// fakeOAICompatServer returns an httptest.Server whose /v1/models responds
// with the given model id list in OpenAI envelope shape.
func fakeOAICompatServer(models ...string) *httptest.Server {
	mux := http.NewServeMux()
	mux.HandleFunc("/v1/models", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(200)
		w.Write([]byte(`{"data":[`))
		for i, m := range models {
			if i > 0 {
				w.Write([]byte(","))
			}
			w.Write([]byte(`{"id":"` + m + `"}`))
		}
		w.Write([]byte(`]}`))
	})
	return httptest.NewServer(mux)
}

func TestDedupURLs(t *testing.T) {
	cases := []struct {
		in   []string
		want []string
	}{
		{nil, []string{}},
		{[]string{}, []string{}},
		{[]string{"a", "b"}, []string{"a", "b"}},
		{[]string{"a", "a", "b"}, []string{"a", "b"}},
		{[]string{"", "a", "", "b"}, []string{"a", "b"}},
		{[]string{"a", "b", "a"}, []string{"a", "b"}},
	}
	for _, c := range cases {
		got := dedupURLs(c.in)
		if len(got) != len(c.want) {
			t.Errorf("dedupURLs(%v): len=%d want=%d (%v)", c.in, len(got), len(c.want), got)
			continue
		}
		for i := range got {
			if got[i] != c.want[i] {
				t.Errorf("dedupURLs(%v)[%d]=%q want=%q", c.in, i, got[i], c.want[i])
			}
		}
	}
}

func TestServicePool_OllamaFetcher_FirstMatchWins(t *testing.T) {
	srvA := fakeOllamaServer("alpha:7b", "shared:13b")
	defer srvA.Close()
	srvB := fakeOllamaServer("beta:7b", "shared:13b")
	defer srvB.Close()

	p := newServicePool("ollama", []string{srvA.URL, srvB.URL}, fetchOllamaTags, &http.Client{Timeout: 2 * time.Second})
	p.Refresh(context.Background())

	if got := p.URLForModel("alpha:7b"); got != srvA.URL {
		t.Errorf("alpha:7b → %q want %q", got, srvA.URL)
	}
	if got := p.URLForModel("beta:7b"); got != srvB.URL {
		t.Errorf("beta:7b → %q want %q", got, srvB.URL)
	}
	if got := p.URLForModel("shared:13b"); got != srvA.URL {
		t.Errorf("shared:13b → %q want %q (first-match wins)", got, srvA.URL)
	}
}

func TestServicePool_OAICompatFetcher_FirstMatchWins(t *testing.T) {
	srvA := fakeOAICompatServer("qwen3-coder:30b", "shared/big")
	defer srvA.Close()
	srvB := fakeOAICompatServer("gemma3:12b", "shared/big")
	defer srvB.Close()

	p := newServicePool("lmstudio", []string{srvA.URL, srvB.URL}, fetchOpenAICompatModels, &http.Client{Timeout: 2 * time.Second})
	p.Refresh(context.Background())

	if got := p.URLForModel("qwen3-coder:30b"); got != srvA.URL {
		t.Errorf("qwen3-coder:30b → %q want %q", got, srvA.URL)
	}
	if got := p.URLForModel("gemma3:12b"); got != srvB.URL {
		t.Errorf("gemma3:12b → %q want %q", got, srvB.URL)
	}
	if got := p.URLForModel("shared/big"); got != srvA.URL {
		t.Errorf("shared/big → %q want %q (first-match wins)", got, srvA.URL)
	}
}

func TestServicePool_MissReturnsFirstURL(t *testing.T) {
	srvA := fakeOllamaServer("alpha:7b")
	defer srvA.Close()
	srvB := fakeOllamaServer("beta:7b")
	defer srvB.Close()

	p := newServicePool("ollama", []string{srvA.URL, srvB.URL}, fetchOllamaTags, &http.Client{Timeout: 2 * time.Second})
	p.Refresh(context.Background())

	if got := p.URLForModel("missing:1b"); got != srvA.URL {
		t.Errorf("missing model → %q want %q (first URL)", got, srvA.URL)
	}
}

func TestServicePool_EmptyPoolReturnsEmpty(t *testing.T) {
	p := newServicePool("custom", nil, fetchOpenAICompatModels, nil)
	if got := p.URLForModel("anything"); got != "" {
		t.Errorf("empty pool → %q want \"\"", got)
	}
}

func TestServicePool_SetURLs_DropsStaleModels(t *testing.T) {
	srvA := fakeOllamaServer("alpha:7b")
	defer srvA.Close()
	srvB := fakeOllamaServer("beta:7b")
	defer srvB.Close()

	p := newServicePool("ollama", []string{srvA.URL}, fetchOllamaTags, &http.Client{Timeout: 2 * time.Second})
	p.Refresh(context.Background())
	if got := p.URLForModel("alpha:7b"); got != srvA.URL {
		t.Fatalf("pre-swap: alpha:7b → %q want %q", got, srvA.URL)
	}

	p.SetURLs([]string{srvB.URL})
	if got := p.URLForModel("alpha:7b"); got != srvB.URL {
		t.Errorf("post-swap miss: alpha:7b → %q want %q (first URL of new list)", got, srvB.URL)
	}
	p.Refresh(context.Background())
	if got := p.URLForModel("beta:7b"); got != srvB.URL {
		t.Errorf("post-swap hit: beta:7b → %q want %q", got, srvB.URL)
	}
}

func TestSplitOllamaURLs(t *testing.T) {
	cases := []struct {
		in   string
		want []string
	}{
		{"", nil},
		{"http://a:11434", []string{"http://a:11434"}},
		{"http://a:11434,http://b:11434", []string{"http://a:11434", "http://b:11434"}},
		{"http://a:11434, http://b:11434", []string{"http://a:11434", "http://b:11434"}},
		{"http://a:11434,,http://b:11434", []string{"http://a:11434", "http://b:11434"}},
		{"  http://a:11434  ", []string{"http://a:11434"}},
	}
	for _, c := range cases {
		got := splitOllamaURLs(c.in)
		if len(got) != len(c.want) {
			t.Errorf("splitOllamaURLs(%q): len=%d want=%d (%v)", c.in, len(got), len(c.want), got)
			continue
		}
		for i := range got {
			if got[i] != c.want[i] {
				t.Errorf("splitOllamaURLs(%q)[%d]=%q want=%q", c.in, i, got[i], c.want[i])
			}
		}
	}
}

func TestSplitInstanceAlias(t *testing.T) {
	cases := []struct {
		in        string
		wantClean string
		wantAlias string
	}{
		{"qwen3:32b", "qwen3:32b", ""},
		{"qwen3:32b@local1", "qwen3:32b", "local1"},
		{"qwen3:32b@local12", "qwen3:32b", "local12"},
		{"publisher/model@local2", "publisher/model", "local2"},
		// Not a recognized alias — stays in the model id.
		{"model@variant", "model@variant", ""},
		{"qwen3:32b@local", "qwen3:32b@local", ""},
		{"qwen3:32b@localabc", "qwen3:32b@localabc", ""},
		{"@local1", "", "local1"},
	}
	for _, c := range cases {
		clean, alias := splitInstanceAlias(c.in)
		if clean != c.wantClean || alias != c.wantAlias {
			t.Errorf("splitInstanceAlias(%q) = (%q, %q), want (%q, %q)", c.in, clean, alias, c.wantClean, c.wantAlias)
		}
	}
}

func TestServicePool_AliasOverridesModelLookup(t *testing.T) {
	srvA := fakeOllamaServer("qwen3:32b")
	defer srvA.Close()
	srvB := fakeOllamaServer("qwen3:32b") // same model on both URLs
	defer srvB.Close()

	p := newServicePool("ollama", []string{srvA.URL, srvB.URL}, fetchOllamaTags, &http.Client{Timeout: 2 * time.Second})
	p.Refresh(context.Background())

	// No alias → first match wins.
	if got := p.URLForModel("qwen3:32b"); got != srvA.URL {
		t.Errorf("plain qwen3:32b → %q want %q", got, srvA.URL)
	}
	// @local1 pins to URL[0] = srvA.
	if got := p.URLForModel("qwen3:32b@local1"); got != srvA.URL {
		t.Errorf("qwen3:32b@local1 → %q want %q", got, srvA.URL)
	}
	// @local2 pins to URL[1] = srvB even though qwen3:32b is in the cache for srvA.
	if got := p.URLForModel("qwen3:32b@local2"); got != srvB.URL {
		t.Errorf("qwen3:32b@local2 → %q want %q", got, srvB.URL)
	}
	// Unknown alias → falls back to first URL (caller will get a real 404).
	if got := p.URLForModel("qwen3:32b@local99"); got != srvA.URL {
		t.Errorf("qwen3:32b@local99 → %q want %q (fallback to first)", got, srvA.URL)
	}
}

func TestServicePool_TolerateOneDeadURL(t *testing.T) {
	srvA := fakeOllamaServer("alpha:7b")
	defer srvA.Close()
	dead := "http://127.0.0.1:1"

	p := newServicePool("ollama", []string{dead, srvA.URL}, fetchOllamaTags, &http.Client{Timeout: 500 * time.Millisecond})
	p.Refresh(context.Background())

	if got := p.URLForModel("alpha:7b"); got != srvA.URL {
		t.Errorf("alpha:7b with one dead URL → %q want %q", got, srvA.URL)
	}
}
