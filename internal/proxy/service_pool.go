// Package proxy: multi-instance dispatch pool for local AI services.
//
// servicePool keeps a list of base URLs for one logical service (ollama,
// lmstudio, vllm, llamacpp, custom OpenAI-compat plugins…) and a model→URL
// map populated by polling each URL's discovery endpoint. Dispatch picks a
// URL per request based on the requested model id. First-match wins on
// duplicates; a miss returns the first URL in the list so the caller still
// gets a real 404 from the upstream rather than a fabricated error.
//
// The discovery endpoint differs per service, so the pool is constructed
// with a modelFetcher chosen by kind:
//
//   - "ollama"        → /api/tags
//   - "openai_compat" → /v1/models  (lmstudio, vllm, llamacpp, custom)
//
// Resolution rule for the URL list (rebuilt on vault sync):
//
//	urls = dedup(env <SERVICE>_URLS list + env <SERVICE>_URL + vault serviceURLs[id])
//
// A pool with a single URL behaves identically to the pre-pool single-URL
// dispatch path — every model resolves to that one URL.
package proxy

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"strings"
	"sync"
	"time"
)

const (
	servicePoolDefaultRefresh = 5 * time.Minute
	servicePoolFetchTimeout   = 5 * time.Second
)

// modelFetcher is the discovery probe a pool runs against each URL. Returning
// nil means "this URL contributed no models" — a transient outage on one
// instance must not erase entries from the others. Errors are swallowed so
// the refresh stays best-effort and never blocks dispatch.
type modelFetcher func(ctx context.Context, baseURL string, client *http.Client) []string

type servicePool struct {
	mu          sync.RWMutex
	serviceID   string            // for logging only
	urls        []string          // immutable list; rebuilt by SetURLs
	aliasToURL  map[string]string // local1..localN → urls[0..N-1]; rebuilt by SetURLs
	modelToURL  map[string]string // populated by Refresh
	fetcher     modelFetcher
	httpClient  *http.Client
	refreshIv   time.Duration
	stopCh      chan struct{}
	stopOnce    sync.Once
	missTrigger chan struct{} // buffered cap-1; signals refresh on cache miss
}

// newServicePool builds a servicePool with the given initial URL list and a
// model-discovery fetcher. A nil or short HTTP client is acceptable — the
// poller reuses long-lived idle conns when the supplied client has them, but
// works without keep-alive too. initialURLs may be empty; the caller is
// expected to call SetURLs once it has resolved the env/vault inputs.
func newServicePool(serviceID string, initialURLs []string, fetcher modelFetcher, httpClient *http.Client) *servicePool {
	if httpClient == nil {
		httpClient = &http.Client{Timeout: servicePoolFetchTimeout}
	}
	if fetcher == nil {
		fetcher = func(context.Context, string, *http.Client) []string { return nil }
	}
	clean := dedupURLs(initialURLs)
	return &servicePool{
		serviceID:   serviceID,
		urls:        clean,
		aliasToURL:  buildAliasMap(clean),
		modelToURL:  map[string]string{},
		fetcher:     fetcher,
		httpClient:  httpClient,
		refreshIv:   servicePoolDefaultRefresh,
		stopCh:      make(chan struct{}),
		missTrigger: make(chan struct{}, 1),
	}
}

// buildAliasMap pairs each URL with a stable positional alias "local<N>"
// (1-indexed) so callers can pin a request to a specific instance via
// "<model>@local<N>". The 1-based numbering matches operator-friendly
// ordering: the first URL in the dashboard becomes local1, the second
// becomes local2, and so on. Reordering the URL list rotates the
// aliases, which is intentional — the alias is a positional reference,
// not an opaque identifier. Operators who want stable names should pin
// instances by URL list ordering when editing the dashboard.
func buildAliasMap(urls []string) map[string]string {
	out := make(map[string]string, len(urls))
	for i, u := range urls {
		out[fmt.Sprintf("local%d", i+1)] = u
	}
	return out
}

// dedupURLs removes empty/duplicate entries while preserving order. First
// occurrence wins, so an env-supplied URL takes priority over a later
// vault-supplied one when SetURLs concatenates them.
func dedupURLs(in []string) []string {
	seen := map[string]bool{}
	out := make([]string, 0, len(in))
	for _, u := range in {
		if u == "" || seen[u] {
			continue
		}
		seen[u] = true
		out = append(out, u)
	}
	return out
}

// SetURLs replaces the URL list. Called once at startup and whenever vault
// sync changes the upstream targets. Drops the existing model map so stale
// model→URL entries from a removed instance can't misroute the next request;
// the next Refresh repopulates from the current URL set. A non-blocking
// nudge to missTrigger asks the running poller to refresh immediately
// instead of waiting up to five minutes for the next tick — users who flip
// the URL list in the dashboard see the new model catalog within seconds.
func (p *servicePool) SetURLs(urls []string) {
	clean := dedupURLs(urls)
	p.mu.Lock()
	if equalStringSlices(p.urls, clean) {
		p.mu.Unlock()
		return
	}
	p.urls = clean
	p.aliasToURL = buildAliasMap(clean)
	p.modelToURL = map[string]string{}
	p.mu.Unlock()
	select {
	case p.missTrigger <- struct{}{}:
	default:
	}
}

func equalStringSlices(a, b []string) bool {
	if len(a) != len(b) {
		return false
	}
	for i := range a {
		if a[i] != b[i] {
			return false
		}
	}
	return true
}

// URLs returns a copy of the current URL list. Defensive copy so a caller
// can range over it without holding the lock or mutating internal state.
func (p *servicePool) URLs() []string {
	p.mu.RLock()
	defer p.mu.RUnlock()
	out := make([]string, len(p.urls))
	copy(out, p.urls)
	return out
}

// URLForModel returns the base URL that hosts the given model id. The
// model id may carry a positional alias suffix ("<model>@local<N>"); when
// present, the alias maps directly to a URL via aliasToURL — the
// model-discovery cache is bypassed so the caller's instance pin always
// wins, even if that instance hasn't been polled yet. Without an alias,
// the model→URL cache is consulted; on miss returns the first configured
// URL (so dispatch still hits a real upstream and gets back a real 404)
// and triggers a non-blocking refresh in case the model just appeared.
// With an empty pool returns "" — the caller is responsible for
// surfacing that as a config error.
func (p *servicePool) URLForModel(model string) string {
	_, alias := splitInstanceAlias(model)
	p.mu.RLock()
	if alias != "" {
		if u, ok := p.aliasToURL[alias]; ok {
			p.mu.RUnlock()
			return u
		}
	}
	cleanModel := stripInstanceAlias(model)
	if u, ok := p.modelToURL[cleanModel]; ok {
		p.mu.RUnlock()
		return u
	}
	first := ""
	if len(p.urls) > 0 {
		first = p.urls[0]
	}
	p.mu.RUnlock()
	select {
	case p.missTrigger <- struct{}{}:
	default:
	}
	return first
}

// splitInstanceAlias returns (cleanModel, alias) for a model id that
// may carry a "@local<N>" suffix. Models with no suffix yield an empty
// alias. The split is conservative: only the suffix matching
// /^@local\d+$/ is recognized as an alias so an "@" inside the model id
// itself (rare in Ollama, common in OpenRouter route ids like
// "publisher/model@variant") is not accidentally consumed.
func splitInstanceAlias(model string) (string, string) {
	idx := strings.LastIndex(model, "@")
	if idx < 0 {
		return model, ""
	}
	suffix := model[idx+1:]
	if !looksLikeAlias(suffix) {
		return model, ""
	}
	return model[:idx], suffix
}

// stripInstanceAlias returns the model id with any recognized alias
// suffix removed. Used at the boundary where wall-vault forwards the
// request to the upstream backend — Ollama and OpenAI-compatible
// servers must not see the alias suffix.
func stripInstanceAlias(model string) string {
	clean, _ := splitInstanceAlias(model)
	return clean
}

func looksLikeAlias(s string) bool {
	if !strings.HasPrefix(s, "local") {
		return false
	}
	rest := s[len("local"):]
	if rest == "" {
		return false
	}
	for _, r := range rest {
		if r < '0' || r > '9' {
			return false
		}
	}
	return true
}

// Models returns the union of model ids known across all URLs in the pool.
// The caller can use this to decorate dashboards or pre-warm registry
// caches; ordering is map-iteration order (caller sorts if needed).
func (p *servicePool) Models() []string {
	p.mu.RLock()
	defer p.mu.RUnlock()
	out := make([]string, 0, len(p.modelToURL))
	for m := range p.modelToURL {
		out = append(out, m)
	}
	return out
}

// Refresh polls each URL via the configured fetcher and rebuilds modelToURL.
// URL order is honoured: the first URL that lists a model wins, so a stable
// dispatch target across refreshes is the URL list ordering. A failed poll
// for one URL doesn't drop entries from the others; we just skip its
// contribution.
func (p *servicePool) Refresh(ctx context.Context) {
	urls := p.URLs()
	if len(urls) == 0 {
		return
	}
	merged := map[string]string{}
	for _, u := range urls {
		for _, m := range p.fetcher(ctx, u, p.httpClient) {
			if _, exists := merged[m]; !exists {
				merged[m] = u
			}
		}
	}
	p.mu.Lock()
	p.modelToURL = merged
	p.mu.Unlock()
}

// Start kicks off the background poller. Runs an initial refresh
// synchronously so the first dispatch after Start() already sees populated
// state, then loops on a ticker. Cache-miss triggers from URLForModel and
// SetURLs collapse into the next iteration so a hot loop of misses doesn't
// melt the upstream.
func (p *servicePool) Start(ctx context.Context) {
	p.Refresh(ctx)
	go p.run(ctx)
}

func (p *servicePool) run(ctx context.Context) {
	ticker := time.NewTicker(p.refreshIv)
	defer ticker.Stop()
	for {
		select {
		case <-ctx.Done():
			return
		case <-p.stopCh:
			return
		case <-ticker.C:
			p.Refresh(ctx)
		case <-p.missTrigger:
			p.Refresh(ctx)
		}
	}
}

// Stop signals the background goroutine to exit. Safe to call multiple times.
func (p *servicePool) Stop() {
	p.stopOnce.Do(func() { close(p.stopCh) })
}

// String renders a compact summary for log lines and dashboards.
func (p *servicePool) String() string {
	p.mu.RLock()
	defer p.mu.RUnlock()
	return fmt.Sprintf("servicePool{id=%s urls=%d models=%d}", p.serviceID, len(p.urls), len(p.modelToURL))
}

// ─── Discovery fetchers ─────────────────────────────────────────────────────

type ollamaTagsResponse struct {
	Models []struct {
		Name string `json:"name"`
	} `json:"models"`
}

// fetchOllamaTags hits <url>/api/tags. Used for the "ollama" service.
func fetchOllamaTags(ctx context.Context, baseURL string, client *http.Client) []string {
	reqCtx, cancel := context.WithTimeout(ctx, servicePoolFetchTimeout)
	defer cancel()
	req, err := http.NewRequestWithContext(reqCtx, "GET", baseURL+"/api/tags", nil)
	if err != nil {
		return nil
	}
	resp, err := client.Do(req)
	if err != nil {
		return nil
	}
	defer resp.Body.Close()
	if resp.StatusCode != http.StatusOK {
		return nil
	}
	var parsed ollamaTagsResponse
	if err := json.NewDecoder(resp.Body).Decode(&parsed); err != nil {
		return nil
	}
	out := make([]string, 0, len(parsed.Models))
	for _, m := range parsed.Models {
		if m.Name != "" {
			out = append(out, m.Name)
		}
	}
	return out
}

type oaiCompatModelsResponse struct {
	Data []struct {
		ID string `json:"id"`
	} `json:"data"`
}

// fetchOpenAICompatModels hits <url>/v1/models. Used for lmstudio, vllm,
// llamacpp, and any custom OpenAI-compatible plugin. The /v1/models reply
// is a {data:[{id, ...}]} envelope per the OpenAI spec.
func fetchOpenAICompatModels(ctx context.Context, baseURL string, client *http.Client) []string {
	reqCtx, cancel := context.WithTimeout(ctx, servicePoolFetchTimeout)
	defer cancel()
	req, err := http.NewRequestWithContext(reqCtx, "GET", baseURL+"/v1/models", nil)
	if err != nil {
		return nil
	}
	resp, err := client.Do(req)
	if err != nil {
		return nil
	}
	defer resp.Body.Close()
	if resp.StatusCode != http.StatusOK {
		return nil
	}
	var parsed oaiCompatModelsResponse
	if err := json.NewDecoder(resp.Body).Decode(&parsed); err != nil {
		return nil
	}
	out := make([]string, 0, len(parsed.Data))
	for _, m := range parsed.Data {
		if m.ID != "" {
			out = append(out, m.ID)
		}
	}
	return out
}
