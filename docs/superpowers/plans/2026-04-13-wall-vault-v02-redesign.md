# wall-vault v0.2.0 Redesign Implementation Plan

> **For agentic workers:** REQUIRED SUB-SKILL: Use `superpowers:subagent-driven-development` (recommended) or `superpowers:executing-plans` to implement this plan task-by-task. Steps use checkbox (`- [ ]`) syntax for tracking.

**Goal:** `default_model` 을 service가 소유하도록 데이터 모델을 재정의하고 dispatch를 재작성하며, 2,572줄 `vault/ui.go` 를 templ+HTMX 컴포넌트 트리로 교체해 one-screen hybrid 대시보드를 완성한 **v0.2.0** 을 4대 머신(모토코·라즈·미니·작순이)에 롤아웃한다.

**Architecture:** 상세는 `docs/superpowers/specs/2026-04-13-wall-vault-v02-redesign-design.md` 참조. 핵심 — `Service.DefaultModel` / `Service.AllowedModels` / `Service.SortOrder` 신규; `Client.PreferredService` / `Client.ModelOverride` 로 의미 재정의; `resolveModel(client, service)` 헬퍼가 dispatch chain의 매 단계에서 service의 default_model을 자동 적용; 자동 1회성 migration + 강제 백업; templ 컴포넌트는 `internal/vault/views/` 하위, 생성물 커밋.

**Tech Stack:** Go 1.21+, `github.com/a-h/templ`, HTMX (CDN), AES-256-GCM (기존 crypto 유지), SSE broker (기존 유지), stdlib net/http.

**Branch:** `v0.2-redesign` (main에서 분기, 완성 후 merge + tag `v0.2.0-rc1` → `v0.2.0`).

---

## Stage 0 — Preparation

### Task 1: v0.1.29 hot-fix 커밋 정리 (main 브랜치)

현재 working tree에 v0.1.29 convert.go 수정이 커밋 안 됨. 이미 배포된 상태라 main에 정식 커밋 후 태그.

**Files:**
- Modify: `.gitignore`, `CHANGELOG.md`, `Makefile`, `internal/proxy/call_anthropic.go`, `internal/proxy/convert.go`

- [ ] **Step 1: git status 확인 및 변경된 파일 diff 확인**

```bash
git status --short
git diff --stat
```

Expected: `.gitignore`, `CHANGELOG.md`, `Makefile`, `internal/proxy/call_anthropic.go`, `internal/proxy/convert.go` 변경 상태.

- [ ] **Step 2: 5개 파일 명시적 add + 커밋**

```bash
git add .gitignore CHANGELOG.md Makefile internal/proxy/call_anthropic.go internal/proxy/convert.go
git commit -m "$(cat <<'EOF'
v0.1.29: fix Anthropic→Gemini tool_use/tool_result loss in dispatch

- AnthropicToGemini now routes via anthropicToOpenAIReq + OpenAIToGemini
  so tool_use / tool_result content blocks become proper functionCall /
  functionResponse parts instead of empty-text entries. Collapsed blocks
  were the root cause of Google 400 "contents is not specified" and
  Ollama/OpenRouter "messages is too short" cascades.
- doAnthropicRequest serializes tool-only turns as JSON content fallback
  and skips genuinely empty turns before the "변환할 메시지 없음" guard,
  so zero-text tool round trips do not collapse the whole request.
- BASE_VERSION bumped to v0.1.29; CHANGELOG has the full fix note.
- .gitignore: add .superpowers/ (brainstorm session artefacts).

Deployed live to motoko/raspi/mini/jaksooni on 2026-04-13.
EOF
)"
```

- [ ] **Step 3: 태그 v0.1.29**

```bash
git tag -a v0.1.29 -m "v0.1.29: tool_use/tool_result preservation in Anthropic dispatch"
```

- [ ] **Step 4: git push origin main v0.1.29**

```bash
git push origin main
git push origin v0.1.29
```

Expected: commit + tag 모두 업로드.

---

### Task 2: `v0.2-redesign` 브랜치 생성

**Files:** 없음 (git branching만)

- [ ] **Step 1: 브랜치 생성 및 체크아웃**

```bash
git checkout -b v0.2-redesign
git branch -vv
```

Expected: `* v0.2-redesign ... [v0.1.29 최상위 커밋]`.

- [ ] **Step 2: Makefile BASE_VERSION v0.2.0 변경**

```bash
sed -i 's/^BASE_VERSION = v0.1.29$/BASE_VERSION = v0.2.0/' Makefile
grep BASE_VERSION Makefile
```

Expected: `BASE_VERSION = v0.2.0`.

- [ ] **Step 3: 커밋**

```bash
git add Makefile
git commit -m "chore(v0.2): bump BASE_VERSION to v0.2.0 on v0.2-redesign branch"
```

---

### Task 3: templ 툴체인 설치 + Makefile `templ generate` 타겟

**Files:**
- Modify: `go.mod`, `go.sum`, `Makefile`
- Create: `.gitattributes` (선택)

- [ ] **Step 1: templ 의존성 고정**

```bash
~/go/bin/go get github.com/a-h/templ@v0.2.747
~/go/bin/go install github.com/a-h/templ/cmd/templ@v0.2.747
~/go/bin/templ version
```

Expected: templ 버전 `v0.2.747` 출력.

- [ ] **Step 2: Makefile에 `templ generate` 타겟 추가 + `build` 의존성 연결**

Makefile의 `.PHONY: build` 선언 아래(또는 현재 build 타겟 바로 위)에 다음을 추가:

```makefile
.PHONY: templ-generate
templ-generate:
	~/go/bin/templ generate ./internal/vault/views/...
```

그리고 기존 `build:` 타겟을 다음처럼 수정:

```makefile
build: templ-generate
	~/go/bin/go build $(LDFLAGS) -o bin/$(BINARY) .
```

- [ ] **Step 3: go.mod/go.sum + Makefile 커밋**

```bash
git add go.mod go.sum Makefile
git commit -m "build(v0.2): pin templ v0.2.747 and wire templ-generate into build"
```

---

## Stage 1 — Data Model

### Task 4: `Service` / `Client` 구조체 재정의 + 단위 테스트

**Files:**
- Modify: `internal/vault/models.go`
- Create: `internal/vault/models_test.go`

- [ ] **Step 1: 테스트 작성 (models_test.go 신규)**

```go
package vault

import (
	"encoding/json"
	"testing"
)

func TestServiceRoundTrip(t *testing.T) {
	in := Service{
		ID: "google", Name: "Google Gemini",
		DefaultModel: "gemini-3.1-pro-preview",
		AllowedModels: []string{"gemini-3.1-pro-preview", "gemini-3.1-flash-lite-preview"},
		Enabled: true, ProxyEnabled: true, SortOrder: 2,
	}
	buf, err := json.Marshal(in)
	if err != nil { t.Fatal(err) }
	var out Service
	if err := json.Unmarshal(buf, &out); err != nil { t.Fatal(err) }
	if out.DefaultModel != in.DefaultModel || len(out.AllowedModels) != 2 {
		t.Fatalf("roundtrip mismatch: %+v", out)
	}
}

func TestClientNewFields(t *testing.T) {
	c := Client{ID: "mini9", PreferredService: "google", ModelOverride: ""}
	if c.PreferredService != "google" {
		t.Fatal("PreferredService missing")
	}
}
```

- [ ] **Step 2: 테스트 실행 (빌드 실패 확인)**

```bash
~/go/bin/go test ./internal/vault/ -run TestServiceRoundTrip -v
```

Expected: 컴파일 에러 — `Service.DefaultModel` 등 미정의.

- [ ] **Step 3: `models.go` 의 `ServiceConfig` / `Client` 를 spec §3 대로 수정**

`internal/vault/models.go` 에서:

1. 기존 `ServiceConfig` 타입을 **삭제하거나 주석 제거**하고 대신 `Service` 로 재정의:

```go
// Service represents an upstream LLM provider registration.
// v0.2: default_model moves here; the client only picks the service
// (and optionally overrides the model via Client.ModelOverride).
type Service struct {
	ID            string   `json:"id"`
	Name          string   `json:"name"`
	DefaultModel  string   `json:"default_model"`
	LocalURL      string   `json:"local_url,omitempty"`
	Enabled       bool     `json:"enabled"`
	ProxyEnabled  bool     `json:"proxy_enabled"`
	SortOrder     int      `json:"sort_order"`
	AllowedModels []string `json:"allowed_models,omitempty"`
}
```

2. 기존 `Client` 구조체의 `DefaultService` → `PreferredService`, `DefaultModel` → `ModelOverride` 로 필드명과 JSON 태그 변경:

```go
type Client struct {
	ID               string    `json:"id"`
	Name             string    `json:"name"`
	Token            string    `json:"token"`
	PreferredService string    `json:"preferred_service"`
	ModelOverride    string    `json:"model_override,omitempty"`
	AllowedServices  []string  `json:"allowed_services,omitempty"`
	AgentType        string    `json:"agent_type,omitempty"`
	WorkDir          string    `json:"work_dir,omitempty"`
	IPWhitelist      []string  `json:"ip_whitelist,omitempty"`
	Avatar           string    `json:"avatar,omitempty"`
	Enabled          bool      `json:"enabled"`
	SortOrder        int       `json:"sort_order"`
	CreatedAt        time.Time `json:"created_at"`
}
```

(`ClientUpdateInput` 구조체도 동일한 필드명으로 갱신; pointer 타입 유지.)

- [ ] **Step 4: 빌드 + 테스트 다시 실행 (컴파일 에러 - 참조부 정리 필요)**

```bash
~/go/bin/go build ./... 2>&1 | head -30
```

Expected: `store.go`, `server.go`, `proxy/*.go` 에서 `DefaultService` / `DefaultModel` / `ServiceConfig` 사용부 다수 에러. 그 참조들은 Stage 2~4 에서 순차적으로 수정되므로 **이 task 범위에서는 참조부 최소 수정**:

먼저 `internal/vault/models.go` 안의 다른 구조체 (`ProxyStatus` 등)가 구 필드명을 참조하면 전부 바꾸고, 외부 파일 (store, server, proxy)은 해당 Stage에서 처리한다는 주석을 남긴다:

```go
// NOTE (v0.2): external references in store.go / server.go / proxy/*.go
// migrate in subsequent tasks.  Keeping models self-consistent here.
```

- [ ] **Step 5: models_test.go 통과 확인 후 커밋**

```bash
~/go/bin/go test ./internal/vault/ -run TestServiceRoundTrip -v
~/go/bin/go test ./internal/vault/ -run TestClientNewFields -v
git add internal/vault/models.go internal/vault/models_test.go
git commit -m "feat(v0.2): redefine Service with default_model and Client.preferred_service"
```

Expected: 두 test PASS.

---

### Task 5: `Store` envelope에 `schema_version` 추가

**Files:**
- Modify: `internal/vault/store.go`

- [ ] **Step 1: 현재 vault envelope struct 위치 확인**

```bash
grep -n "type.*Store.*struct\|SchemaVersion" internal/vault/store.go
```

- [ ] **Step 2: vault envelope struct에 `SchemaVersion int` 필드 추가**

```go
type vaultEnvelope struct {
	SchemaVersion int         `json:"schema_version"`
	Services      []Service   `json:"services"`
	Clients       []Client    `json:"clients"`
	APIKeys       []APIKey    `json:"api_keys"`
	Theme         string      `json:"theme,omitempty"`
	Lang          string      `json:"lang,omitempty"`
}
```

(기존 envelope 구조 파일 이름·필드 이름이 다르면 해당 파일의 실제 이름을 따라서 맞춘다. 기존 store.go 를 grep으로 먼저 확인.)

- [ ] **Step 3: 저장 시 `SchemaVersion = 2` 로 항상 기록**

store의 Save 함수에서 `env.SchemaVersion = 2` 를 직렬화 직전에 set.

- [ ] **Step 4: 기존 store 단위 테스트 실행**

```bash
~/go/bin/go test ./internal/vault/ -v 2>&1 | head -40
```

Expected: models_test pass, store_test 일부 컴파일 에러(참조 필드명 변경 때문); 이 task에서는 최소 수정으로 컴파일 통과.

- [ ] **Step 5: 커밋**

```bash
git add internal/vault/store.go
git commit -m "feat(v0.2): add schema_version=2 to vault envelope"
```

---

## Stage 2 — Migration

### Task 6: `migrate.go` — v1 → v2 필드 변환 함수

**Files:**
- Create: `internal/vault/migrate.go`
- Create: `internal/vault/migrate_test.go`

- [ ] **Step 1: 테스트 먼저 작성 — 실제 v0.1.x 스키마 샘플로 변환 검증**

```go
package vault

import (
	"encoding/json"
	"testing"
)

// v1Envelope mirrors the legacy shape just enough to exercise migration.
type v1Envelope struct {
	Services []struct {
		ID, Name, LocalURL string
		Enabled, ProxyEnabled bool
	} `json:"services"`
	Clients []struct {
		ID, Name, Token string
		DefaultService, DefaultModel string
		AllowedServices []string
		AgentType string
		Enabled bool
		SortOrder int
	} `json:"clients"`
	APIKeys []APIKey `json:"api_keys"`
}

func TestMigrateV1ToV2(t *testing.T) {
	v1Raw := []byte(`{
		"services": [
			{"id":"google","name":"Google","enabled":true,"proxy_enabled":true},
			{"id":"ollama","name":"Ollama","local_url":"http://192.168.0.6:11434","enabled":true,"proxy_enabled":true}
		],
		"clients": [
			{"id":"mini9","name":"작순이","token":"t","default_service":"google","default_model":"gemini-3.1-pro-preview","agent_type":"nanoclaw","enabled":true,"sort_order":4},
			{"id":"motoko","name":"모토코","token":"t2","default_service":"google","default_model":"gemini-3.1-pro-preview","agent_type":"openclaw","enabled":true,"sort_order":1}
		],
		"api_keys": []
	}`)
	out, err := MigrateV1ToV2(v1Raw)
	if err != nil { t.Fatal(err) }
	if out.SchemaVersion != 2 { t.Fatalf("want version 2, got %d", out.SchemaVersion) }
	// google service inherits the most-common default_model from clients
	var googleDM string
	for _, s := range out.Services {
		if s.ID == "google" { googleDM = s.DefaultModel }
	}
	if googleDM != "gemini-3.1-pro-preview" {
		t.Fatalf("google default_model = %q, want gemini-3.1-pro-preview", googleDM)
	}
	// clients renamed
	var mini9 Client
	for _, c := range out.Clients {
		if c.ID == "mini9" { mini9 = c }
	}
	if mini9.PreferredService != "google" {
		t.Fatalf("preferred_service = %q", mini9.PreferredService)
	}
	if mini9.ModelOverride != "gemini-3.1-pro-preview" {
		t.Fatalf("model_override = %q", mini9.ModelOverride)
	}
}
```

- [ ] **Step 2: 테스트 실행 → FAIL 확인**

```bash
~/go/bin/go test ./internal/vault/ -run TestMigrateV1ToV2 -v
```

Expected: `MigrateV1ToV2` undefined.

- [ ] **Step 3: `migrate.go` 구현**

```go
package vault

// MigrateV1ToV2 converts the legacy v0.1.x encrypted vault envelope
// (plaintext JSON after decryption) into the v0.2 envelope shape.
//
// Field mapping:
//
//	services[].default_model      ← most frequent clients[].default_model
//	                                whose default_service matches this id
//	                                (ties broken by lowest sort_order client);
//	                                "" if no client picks this service.
//	services[].allowed_models     ← nil (empty = unrestricted).
//	services[].sort_order         ← preserve v1 numeric if present, else 0.
//	clients[].preferred_service   ← clients[].default_service
//	clients[].model_override      ← clients[].default_model
//
// Remaining fields (api_keys, theme, lang, per-client ip_whitelist/etc.)
// are copied as-is.
func MigrateV1ToV2(raw []byte) (*vaultEnvelope, error) {
	// 1. Decode into legacy shape (untyped map for max tolerance).
	var legacy struct {
		Services []struct {
			ID           string   `json:"id"`
			Name         string   `json:"name"`
			LocalURL     string   `json:"local_url,omitempty"`
			Enabled      bool     `json:"enabled"`
			ProxyEnabled bool     `json:"proxy_enabled"`
			SortOrder    int      `json:"sort_order,omitempty"`
		} `json:"services"`
		Clients []struct {
			ID               string    `json:"id"`
			Name             string    `json:"name"`
			Token            string    `json:"token"`
			DefaultService   string    `json:"default_service"`
			DefaultModel     string    `json:"default_model"`
			AllowedServices  []string  `json:"allowed_services,omitempty"`
			AgentType        string    `json:"agent_type,omitempty"`
			WorkDir          string    `json:"work_dir,omitempty"`
			IPWhitelist      []string  `json:"ip_whitelist,omitempty"`
			Avatar           string    `json:"avatar,omitempty"`
			Enabled          bool      `json:"enabled"`
			SortOrder        int       `json:"sort_order"`
			CreatedAt        time.Time `json:"created_at"`
		} `json:"clients"`
		APIKeys []APIKey `json:"api_keys"`
		Theme   string   `json:"theme,omitempty"`
		Lang    string   `json:"lang,omitempty"`
	}
	if err := json.Unmarshal(raw, &legacy); err != nil {
		return nil, fmt.Errorf("migrate: decode v1: %w", err)
	}

	// 2. For each service, compute default_model by majority vote of
	//    its clients, tie-broken by lowest sort_order.
	modelByService := make(map[string]string)
	sortByModel := make(map[string]map[string]int) // svc -> model -> min sort_order seen
	countByModel := make(map[string]map[string]int)
	for _, c := range legacy.Clients {
		if c.DefaultService == "" || c.DefaultModel == "" {
			continue
		}
		if countByModel[c.DefaultService] == nil {
			countByModel[c.DefaultService] = map[string]int{}
			sortByModel[c.DefaultService] = map[string]int{}
		}
		countByModel[c.DefaultService][c.DefaultModel]++
		if cur, ok := sortByModel[c.DefaultService][c.DefaultModel]; !ok || c.SortOrder < cur {
			sortByModel[c.DefaultService][c.DefaultModel] = c.SortOrder
		}
	}
	for svc, counts := range countByModel {
		var best string
		var bestCnt int
		var bestSort int
		for mdl, cnt := range counts {
			so := sortByModel[svc][mdl]
			if cnt > bestCnt || (cnt == bestCnt && so < bestSort) {
				best, bestCnt, bestSort = mdl, cnt, so
			}
		}
		modelByService[svc] = best
	}

	// 3. Build v2 envelope.
	env := &vaultEnvelope{
		SchemaVersion: 2,
		APIKeys:       legacy.APIKeys,
		Theme:         legacy.Theme,
		Lang:          legacy.Lang,
	}
	for _, s := range legacy.Services {
		env.Services = append(env.Services, Service{
			ID:           s.ID,
			Name:         s.Name,
			LocalURL:     s.LocalURL,
			Enabled:      s.Enabled,
			ProxyEnabled: s.ProxyEnabled,
			SortOrder:    s.SortOrder,
			DefaultModel: modelByService[s.ID], // "" if no client picked it
		})
	}
	for _, c := range legacy.Clients {
		env.Clients = append(env.Clients, Client{
			ID:               c.ID,
			Name:             c.Name,
			Token:            c.Token,
			PreferredService: c.DefaultService,
			ModelOverride:    c.DefaultModel,
			AllowedServices:  c.AllowedServices,
			AgentType:        c.AgentType,
			WorkDir:          c.WorkDir,
			IPWhitelist:      c.IPWhitelist,
			Avatar:           c.Avatar,
			Enabled:          c.Enabled,
			SortOrder:        c.SortOrder,
			CreatedAt:        c.CreatedAt,
		})
	}
	return env, nil
}
```

추가 import: `fmt`, `encoding/json`, `time`.

- [ ] **Step 4: 테스트 PASS 확인**

```bash
~/go/bin/go test ./internal/vault/ -run TestMigrateV1ToV2 -v
```

Expected: PASS.

- [ ] **Step 5: 커밋**

```bash
git add internal/vault/migrate.go internal/vault/migrate_test.go
git commit -m "feat(v0.2): v1→v2 migration function with majority-vote default_model"
```

---

### Task 7: 기동 시 자동 migration + 강제 백업 통합

**Files:**
- Modify: `internal/vault/store.go` (Load 경로)

- [ ] **Step 1: 테스트 — `Store.Load` 가 v1 파일을 자동 감지하고 변환 + 백업 복사본을 만든 뒤 v2 envelope를 반환한다**

`internal/vault/store_test.go` 에 다음 테스트 추가:

```go
func TestLoadAutoMigratesV1(t *testing.T) {
	tmp := t.TempDir()
	v1 := `{"services":[{"id":"google","name":"Google","enabled":true,"proxy_enabled":true}],
	        "clients":[{"id":"mini9","name":"작순이","token":"t","default_service":"google","default_model":"gemini-3.1-pro-preview","agent_type":"nanoclaw","enabled":true,"sort_order":1}],
	        "api_keys":[]}`
	// Write encrypted v1 file (use the same crypto helpers as live code).
	path := filepath.Join(tmp, "vault.json")
	if err := writePlain(path, "test-master", []byte(v1)); err != nil { t.Fatal(err) }

	s, err := OpenStore(path, "test-master")
	if err != nil { t.Fatal(err) }
	defer s.Close()

	// backup must exist
	backups, _ := filepath.Glob(filepath.Join(tmp, "vault.json.pre-v02.*.bak"))
	if len(backups) != 1 {
		t.Fatalf("expected exactly one pre-v02 backup, got %d", len(backups))
	}
	// schema_version now 2
	clients := s.ListClients()
	if len(clients) != 1 || clients[0].PreferredService != "google" {
		t.Fatalf("migration did not land: %+v", clients)
	}
}
```

(`writePlain` helper는 테스트 전용으로 crypto.go의 encrypt API를 호출해 v1 평문 JSON을 기록한다. 이미 `crypto_test.go` 에 비슷한 유틸이 있으면 재사용.)

- [ ] **Step 2: 테스트 실행 (FAIL)**

```bash
~/go/bin/go test ./internal/vault/ -run TestLoadAutoMigratesV1 -v
```

Expected: migration 미적용 실패 또는 백업 미생성.

- [ ] **Step 3: `store.go` 의 `OpenStore` / `Load` 함수에 migration 통합**

```go
// in Load()
raw, err := decrypt(ciphertext, masterPassword)
if err != nil { return nil, err }

var peek struct { SchemaVersion int `json:"schema_version"` }
_ = json.Unmarshal(raw, &peek)

if peek.SchemaVersion < 2 {
	// force a backup copy of the on-disk ciphertext before rewriting.
	ts := time.Now().UTC().Format("20060102T150405Z")
	backupPath := path + ".pre-v02." + ts + ".bak"
	if err := os.WriteFile(backupPath, ciphertext, 0o600); err != nil {
		return nil, fmt.Errorf("migrate: cannot write backup %s: %w", backupPath, err)
	}
	log.Printf("[migrate] wrote backup: %s", backupPath)

	env, err := MigrateV1ToV2(raw)
	if err != nil { return nil, err }

	// persist v2 envelope immediately so restart is idempotent.
	if err := s.saveEnvelope(env); err != nil {
		return nil, fmt.Errorf("migrate: save v2 envelope: %w", err)
	}
	log.Printf("[migrate] v1 → v2 complete; services=%d clients=%d", len(env.Services), len(env.Clients))
	return env, nil
}

// fallthrough: existing v2 decode path
```

(변수명 `ciphertext`, `decrypt`, `saveEnvelope`, `path`는 기존 store.go 의 실제 이름에 맞춰 조정.)

- [ ] **Step 4: 테스트 PASS 확인**

```bash
~/go/bin/go test ./internal/vault/ -run TestLoadAutoMigratesV1 -v
```

Expected: backup 파일 1개, schema_version=2, PreferredService 올바름 모두 통과.

- [ ] **Step 5: 전체 vault 패키지 테스트**

```bash
~/go/bin/go test ./internal/vault/ -v 2>&1 | tail -20
```

Expected: 기존 store_test 대부분 pass; 참조부가 바뀐 일부는 다음 Stage에서 해결.

- [ ] **Step 6: 커밋**

```bash
git add internal/vault/store.go internal/vault/store_test.go
git commit -m "feat(v0.2): auto-migrate vault.json on Load with forced .pre-v02.bak"
```

---

## Stage 3 — Dispatch

### Task 8: `resolveModel` 헬퍼 + 단위 테스트

**Files:**
- Create: `internal/proxy/resolve.go`
- Create: `internal/proxy/resolve_test.go`

- [ ] **Step 1: 테스트 작성**

`internal/proxy/resolve_test.go`:

```go
package proxy

import (
	"errors"
	"testing"

	"github.com/sookmook/wall-vault/internal/vault"
)

func TestResolveModel_DefaultWhenNoOverride(t *testing.T) {
	svc := vault.Service{ID: "google", DefaultModel: "gemini-3.1-pro-preview"}
	c := vault.Client{ID: "mini9", ModelOverride: ""}
	m, err := ResolveModel(c, svc)
	if err != nil || m != "gemini-3.1-pro-preview" {
		t.Fatalf("got (%q, %v)", m, err)
	}
}

func TestResolveModel_OverrideWithEmptyAllowed(t *testing.T) {
	svc := vault.Service{ID: "google", DefaultModel: "gemini-3.1-pro-preview"}
	c := vault.Client{ModelOverride: "gemini-2.5-flash-lite"}
	m, err := ResolveModel(c, svc)
	if err != nil || m != "gemini-2.5-flash-lite" {
		t.Fatalf("got (%q, %v)", m, err)
	}
}

func TestResolveModel_OverrideInWhitelist(t *testing.T) {
	svc := vault.Service{
		ID: "google", DefaultModel: "gemini-3.1-pro-preview",
		AllowedModels: []string{"gemini-3.1-pro-preview", "gemini-3.1-flash-lite-preview"},
	}
	c := vault.Client{ModelOverride: "gemini-3.1-flash-lite-preview"}
	m, err := ResolveModel(c, svc)
	if err != nil || m != "gemini-3.1-flash-lite-preview" {
		t.Fatalf("got (%q, %v)", m, err)
	}
}

func TestResolveModel_OverrideNotInWhitelist(t *testing.T) {
	svc := vault.Service{
		ID: "google", DefaultModel: "gemini-3.1-pro-preview",
		AllowedModels: []string{"gemini-3.1-pro-preview"},
	}
	c := vault.Client{ModelOverride: "gemini-2.5-flash-lite"}
	_, err := ResolveModel(c, svc)
	if !errors.Is(err, ErrModelNotAllowed) {
		t.Fatalf("want ErrModelNotAllowed, got %v", err)
	}
}
```

- [ ] **Step 2: 테스트 실행 (FAIL)**

```bash
~/go/bin/go test ./internal/proxy/ -run TestResolveModel -v
```

Expected: `ResolveModel` / `ErrModelNotAllowed` undefined.

- [ ] **Step 3: `resolve.go` 구현**

```go
package proxy

import (
	"errors"
	"fmt"

	"github.com/sookmook/wall-vault/internal/vault"
)

// ErrModelNotAllowed is returned when Client.ModelOverride is non-empty
// and the service has a non-empty AllowedModels whitelist that does not
// include the override.
var ErrModelNotAllowed = errors.New("model_override not in service.allowed_models")

// ResolveModel picks the model to send to a service for a given client,
// following spec §2.1.
func ResolveModel(c vault.Client, s vault.Service) (string, error) {
	if c.ModelOverride == "" {
		return s.DefaultModel, nil
	}
	if len(s.AllowedModels) == 0 {
		return c.ModelOverride, nil
	}
	for _, m := range s.AllowedModels {
		if m == c.ModelOverride {
			return c.ModelOverride, nil
		}
	}
	return "", fmt.Errorf("%w: override=%q service=%q", ErrModelNotAllowed, c.ModelOverride, s.ID)
}
```

- [ ] **Step 4: 테스트 PASS**

```bash
~/go/bin/go test ./internal/proxy/ -run TestResolveModel -v
```

Expected: 4개 test 전부 PASS.

- [ ] **Step 5: 커밋**

```bash
git add internal/proxy/resolve.go internal/proxy/resolve_test.go
git commit -m "feat(v0.2): ResolveModel with strict AllowedModels whitelist"
```

---

### Task 9: `dispatch()` 재작성

**Files:**
- Modify: `internal/proxy/server.go`

- [ ] **Step 1: 기존 `dispatch` 함수 시그니처 파악**

```bash
grep -n "func.*dispatch" internal/proxy/server.go | head -5
```

- [ ] **Step 2: 테스트 작성 (server_test.go)**

`internal/proxy/server_test.go` 에 새 테스트 섹션:

```go
func TestDispatch_FallbackUsesEachServicesDefaultModel(t *testing.T) {
	// Stub service registry: google primary, ollama fallback.
	services := []vault.Service{
		{ID: "google", DefaultModel: "gemini-3.1-pro-preview", ProxyEnabled: true, SortOrder: 1},
		{ID: "ollama", DefaultModel: "gemma4:26b",              ProxyEnabled: true, SortOrder: 4},
	}
	client := vault.Client{ID: "mini9", PreferredService: "google"}

	// Record which (service,model) pairs the low-level caller saw.
	var calls []string
	caller := func(svc, mdl string) error {
		calls = append(calls, svc+"/"+mdl)
		if svc == "google" { return fmt.Errorf("google HTTP 400") }
		return nil // ollama succeeds
	}

	err := dispatchForTest(client, services, caller) // testable wrapper we add
	if err != nil { t.Fatalf("unexpected error: %v", err) }
	want := []string{"google/gemini-3.1-pro-preview", "ollama/gemma4:26b"}
	if !reflect.DeepEqual(calls, want) {
		t.Fatalf("calls=%v, want %v", calls, want)
	}
}
```

- [ ] **Step 3: `dispatchForTest` 래퍼 + `dispatch()` 재작성**

`internal/proxy/server.go` 에 새 버전:

```go
// dispatch routes the Gemini-form request through the preferred service
// first, then fallback chain. Each service resolves its own model via
// ResolveModel.
func (s *Server) dispatch(client vault.Client, req *GeminiRequest) (*GeminiResponse, error) {
	services := s.listEnabledServicesForClient(client) // returns primary first, then sort_order
	return dispatchWith(client, services, func(svc vault.Service, model string) (*GeminiResponse, error) {
		return s.callService(svc, model, req)
	})
}

// dispatchWith is the testable core — it drives the chain given a caller.
func dispatchWith(
	client vault.Client,
	services []vault.Service,
	call func(vault.Service, string) (*GeminiResponse, error),
) (*GeminiResponse, error) {
	var lastErr error
	for _, svc := range services {
		if s.keymgr.IsCoolingDown(svc.ID) { continue } // skip cooled services
		mdl, err := ResolveModel(client, svc)
		if err != nil {
			log.Printf("[dispatch] skip %s: %v", svc.ID, err)
			lastErr = err
			continue
		}
		resp, err := call(svc, mdl)
		if err == nil { return resp, nil }
		log.Printf("[dispatch] %s failed: %v", svc.ID, err)
		s.keymgr.RecordFailure(svc.ID, err)
		lastErr = err
	}
	if lastErr == nil {
		lastErr = fmt.Errorf("no enabled services")
	}
	return nil, fmt.Errorf("모든 서비스 실패: %w", lastErr)
}

// dispatchForTest exposes the core for unit tests.
func dispatchForTest(
	client vault.Client,
	services []vault.Service,
	caller func(svc, mdl string) error,
) error {
	_, err := dispatchWith(client, services, func(s vault.Service, m string) (*GeminiResponse, error) {
		return &GeminiResponse{}, caller(s.ID, m)
	})
	return err
}
```

(`listEnabledServicesForClient`, `callService`, `keymgr.IsCoolingDown` 등 구체 구현은 기존 server.go 헬퍼 중 가장 가까운 것을 재사용; cooldown은 `internal/proxy/keymgr.go` 의 공개 메서드로 드러낸다.)

- [ ] **Step 4: 테스트 PASS**

```bash
~/go/bin/go test ./internal/proxy/ -run TestDispatch_FallbackUsesEachServicesDefaultModel -v
```

Expected: PASS, `calls=[google/gemini-3.1-pro-preview ollama/gemma4:26b]`.

- [ ] **Step 5: 전체 proxy 패키지 빌드 확인 (기존 server 테스트 중 구 모델 가정 일부는 스킵 허용)**

```bash
~/go/bin/go build ./internal/proxy/... 2>&1 | head
```

- [ ] **Step 6: 커밋**

```bash
git add internal/proxy/server.go internal/proxy/server_test.go
git commit -m "feat(v0.2): dispatch resolves each service's default_model in fallback"
```

---

### Task 10: 구 `Ollama name-mismatch` 휴리스틱 삭제

**Files:**
- Modify: `internal/proxy/server.go`, `internal/proxy/keymgr.go` (해당 함수가 어디 있든)

- [ ] **Step 1: 휴리스틱 위치 확인**

```bash
grep -n "slash\|provider-prefix\|resolveActualModel\|OllamaDefault" internal/proxy/*.go
```

Expected: v0.1.27에서 추가된 fallback 이름 변환 로직 발견.

- [ ] **Step 2: 해당 블록을 삭제하고 주석으로 이유 기록**

```go
// v0.2: Ollama "name-mismatch" heuristic removed.
// dispatch() now uses each service's default_model, so a Gemini/Claude
// model id can never reach the Ollama endpoint in the first place.
```

- [ ] **Step 3: 관련 테스트 실행하여 기대치 재설정**

```bash
~/go/bin/go test ./internal/proxy/ -run TestOllama -v
```

Expected: 구 테스트가 존재하면 삭제 또는 skip 처리하고 spec 기반 새 테스트 통과.

- [ ] **Step 4: 커밋**

```bash
git add internal/proxy/server.go internal/proxy/keymgr.go
git commit -m "refactor(v0.2): remove Ollama name-mismatch heuristic (obsolete)"
```

---

## Stage 4 — Admin API

### Task 11: `/admin/services` CRUD (default_model / allowed_models 편집)

**Files:**
- Modify: `internal/vault/server.go`

- [ ] **Step 1: 기존 services CRUD 핸들러 위치 확인**

```bash
grep -n "handleAdminServices\|/admin/services" internal/vault/server.go | head
```

- [ ] **Step 2: 테스트 작성 — PUT `/admin/services/google` 로 `default_model` 변경 왕복**

```go
func TestAdminPutServiceDefaultModel(t *testing.T) {
	srv := newTestServer(t) // helper that spins an isolated Store + Server
	body := `{"default_model":"gemini-3.1-pro-preview","allowed_models":["gemini-3.1-pro-preview","gemini-3.1-flash-lite-preview"]}`
	req := httptest.NewRequest("PUT", "/admin/services/google", strings.NewReader(body))
	req.Header.Set("Authorization", "Bearer "+srv.AdminToken())
	w := httptest.NewRecorder()
	srv.Handler().ServeHTTP(w, req)
	if w.Code != 200 { t.Fatalf("got %d: %s", w.Code, w.Body.String()) }

	got := srv.Store().GetService("google")
	if got.DefaultModel != "gemini-3.1-pro-preview" {
		t.Fatalf("default_model = %q", got.DefaultModel)
	}
	if len(got.AllowedModels) != 2 {
		t.Fatalf("allowed_models = %v", got.AllowedModels)
	}
}
```

- [ ] **Step 3: 핸들러 수정 — request body schema에 default_model / allowed_models 허용**

```go
type serviceUpdateInput struct {
	Name          *string   `json:"name"`
	DefaultModel  *string   `json:"default_model"`
	LocalURL      *string   `json:"local_url"`
	Enabled       *bool     `json:"enabled"`
	ProxyEnabled  *bool     `json:"proxy_enabled"`
	SortOrder     *int      `json:"sort_order"`
	AllowedModels *[]string `json:"allowed_models"`
}
```

`handleAdminServicesID` 에서 PUT 바디를 위 구조체로 파싱하고 non-nil 필드만 Store에 반영. Store의 `UpdateService` 시그니처도 동일하게 pointer 기반.

- [ ] **Step 4: 테스트 PASS**

```bash
~/go/bin/go test ./internal/vault/ -run TestAdminPutServiceDefaultModel -v
```

- [ ] **Step 5: 커밋**

```bash
git add internal/vault/server.go internal/vault/store.go internal/vault/server_test.go
git commit -m "feat(v0.2): admin services API accepts default_model + allowed_models"
```

---

### Task 12: `/admin/clients` CRUD — `preferred_service` + `model_override` (strict 검증)

**Files:**
- Modify: `internal/vault/server.go`, `internal/vault/store.go`

- [ ] **Step 1: 테스트 — whitelist 위반 시 422**

```go
func TestAdminPutClient_ModelOverrideWhitelistViolation(t *testing.T) {
	srv := newTestServer(t)
	// google restricted to one model
	srv.Store().UpsertService(vault.Service{
		ID: "google", DefaultModel: "gemini-3.1-pro-preview",
		AllowedModels: []string{"gemini-3.1-pro-preview"},
		Enabled: true, ProxyEnabled: true,
	})
	req := httptest.NewRequest("PUT", "/admin/clients/mini9",
		strings.NewReader(`{"preferred_service":"google","model_override":"gemini-2.5-flash-lite"}`))
	req.Header.Set("Authorization", "Bearer "+srv.AdminToken())
	w := httptest.NewRecorder()
	srv.Handler().ServeHTTP(w, req)
	if w.Code != 422 {
		t.Fatalf("got %d: %s", w.Code, w.Body.String())
	}
}
```

- [ ] **Step 2: `handleAdminClientsID` 수정**

PUT body 파싱 후 `ModelOverride != ""` 이면 `ResolveModel` 로 검증:

```go
if inp.ModelOverride != nil && *inp.ModelOverride != "" {
	svcID := inp.PreferredService
	if svcID == nil {
		existing := s.store.GetClient(id)
		svcID = &existing.PreferredService
	}
	svc := s.store.GetService(*svcID)
	_, err := proxy.ResolveModel(
		vault.Client{ModelOverride: *inp.ModelOverride},
		svc,
	)
	if errors.Is(err, proxy.ErrModelNotAllowed) {
		http.Error(w, fmt.Sprintf(`{"error":"model_override %q not in allowed_models of service %q"}`,
			*inp.ModelOverride, svc.ID), 422)
		return
	}
}
```

(`proxy` 패키지 import 추가.)

- [ ] **Step 3: 테스트 PASS 확인**

```bash
~/go/bin/go test ./internal/vault/ -run TestAdminPutClient -v
```

- [ ] **Step 4: 커밋**

```bash
git add internal/vault/server.go internal/vault/store.go internal/vault/server_test.go
git commit -m "feat(v0.2): admin clients API validates model_override against whitelist"
```

---

### Task 13: `/hx/*` HTMX fragment 라우터 뼈대

**Files:**
- Modify: `internal/vault/server.go` — 라우트 등록
- Create: `internal/vault/hx_router.go`

- [ ] **Step 1: `hx_router.go` 신규 파일 작성 (스텁, 컴파일만 통과)**

```go
package vault

import "net/http"

// RegisterHXRoutes wires /hx/* fragment endpoints. Actual render helpers
// live in Task 17 onward; this task installs the mux skeleton so later
// tasks can commit one handler at a time without fighting the router.
func (s *Server) RegisterHXRoutes(mux *http.ServeMux) {
	mux.HandleFunc("/hx/sidebar",              s.hxNotImplemented)
	mux.HandleFunc("/hx/services/grid",        s.hxNotImplemented)
	mux.HandleFunc("/hx/agents/grid",          s.hxNotImplemented)
	mux.HandleFunc("/hx/keys/list",            s.hxNotImplemented)
	// /hx/services/{id}/edit, /hx/clients/{id}/edit handled by pattern below
	mux.HandleFunc("/hx/services/", s.hxServiceSubroute)
	mux.HandleFunc("/hx/clients/",  s.hxClientSubroute)
}

func (s *Server) hxNotImplemented(w http.ResponseWriter, r *http.Request) {
	http.Error(w, "hx endpoint not yet implemented", http.StatusNotImplemented)
}

func (s *Server) hxServiceSubroute(w http.ResponseWriter, r *http.Request) { s.hxNotImplemented(w, r) }
func (s *Server) hxClientSubroute(w http.ResponseWriter, r *http.Request)  { s.hxNotImplemented(w, r) }
```

- [ ] **Step 2: `server.go` 에서 `RegisterHXRoutes(mux)` 호출**

- [ ] **Step 3: 간단한 smoke test**

```go
func TestHXSidebarNotImplementedStub(t *testing.T) {
	srv := newTestServer(t)
	req := httptest.NewRequest("GET", "/hx/sidebar", nil)
	w := httptest.NewRecorder()
	srv.Handler().ServeHTTP(w, req)
	if w.Code != 501 { t.Fatalf("got %d", w.Code) }
}
```

- [ ] **Step 4: 테스트 PASS**

```bash
~/go/bin/go test ./internal/vault/ -run TestHXSidebarNotImplementedStub -v
```

- [ ] **Step 5: 커밋**

```bash
git add internal/vault/hx_router.go internal/vault/server.go internal/vault/server_test.go
git commit -m "feat(v0.2): register /hx/* fragment route skeleton"
```

---

## Stage 5 — UI Base (templ layouts + theme)

### Task 14: `views/layouts/base.templ` — HTML skeleton + HTMX 스크립트 + theme CSS

**Files:**
- Create: `internal/vault/views/layouts/base.templ`
- Create: `internal/vault/views/layouts/base_templ.go` (generate 결과)

- [ ] **Step 1: `base.templ` 작성**

```templ
package layouts

import "github.com/sookmook/wall-vault/internal/vault"

templ Base(theme string, inner templ.Component) {
	<!DOCTYPE html>
	<html lang="ko" data-theme={theme}>
	<head>
		<meta charset="utf-8">
		<meta name="viewport" content="width=device-width,initial-scale=1">
		<title>wall-vault</title>
		<script src="https://unpkg.com/htmx.org@1.9.12"></script>
		@Theme()
	</head>
	<body>
		@inner
	</body>
	</html>
}
```

- [ ] **Step 2: `templ generate ./internal/vault/views/...` 실행**

```bash
~/go/bin/templ generate ./internal/vault/views/...
ls internal/vault/views/layouts/
```

Expected: `base.templ` + `base_templ.go` 파일 둘 다 존재.

- [ ] **Step 3: 컴파일 확인**

```bash
~/go/bin/go build ./internal/vault/views/layouts/
```

- [ ] **Step 4: 커밋**

```bash
git add internal/vault/views/layouts/base.templ internal/vault/views/layouts/base_templ.go
git commit -m "feat(v0.2): base HTML layout templ (HTMX + theme hook)"
```

---

### Task 15: `views/layouts/theme.templ` — 7 테마 CSS 변수

**Files:**
- Create: `internal/vault/views/layouts/theme.templ`
- Create: `internal/vault/views/layouts/theme_templ.go`

- [ ] **Step 1: `theme.templ` 작성**

```templ
package layouts

templ Theme() {
	<style>
		:root { --bg:#fff; --fg:#1a1a1a; --accent:#d64545; --border:#e3e3e3; --muted:#666; }
		[data-theme="dark"]   { --bg:#121212; --fg:#e8e8e8; --accent:#ff6b6b; --border:#2a2a2a; --muted:#999; }
		[data-theme="cherry"] { --bg:#fff1f1; --fg:#3a0a0a; --accent:#c53030; --border:#f0c9c9; --muted:#8a3e3e; }
		[data-theme="ocean"]  { --bg:#e8f4ff; --fg:#0c1a2b; --accent:#1e6fb8; --border:#b5d6f1; --muted:#45678a; }
		[data-theme="gold"]   { --bg:#fff8e6; --fg:#3a2a00; --accent:#b87f00; --border:#e8d99d; --muted:#8a6d24; }
		[data-theme="autumn"] { --bg:#fff3e0; --fg:#2a1a0a; --accent:#c45618; --border:#f0cfa0; --muted:#8a5a2e; }
		[data-theme="winter"] { --bg:#e6f2ff; --fg:#0a1a3a; --accent:#2a4f9c; --border:#b5c9e8; --muted:#4a6a9a; }
		html, body { margin:0; background:var(--bg); color:var(--fg); font-family: -apple-system, Segoe UI, Roboto, sans-serif; }
		.shell { display:grid; grid-template-columns:22% 1fr 28%; min-height:100vh; }
		.shell > nav  { border-right:1px solid var(--border); padding:12px; }
		.shell > main { padding:16px; overflow:auto; }
		.shell > aside{ border-left:2px solid var(--accent); padding:12px; background:rgba(0,0,0,.02); display:none; }
		.shell > aside[data-open]{ display:block; }
		@media (max-width:900px){ .shell{ grid-template-columns:1fr; } .shell>aside[data-open]{ position:fixed; inset:0; z-index:10; background:var(--bg);} }
		.card { border:1px solid var(--border); border-radius:8px; padding:10px 12px; margin:6px 0; cursor:pointer; }
		.card:hover{ border-color:var(--accent); }
		.badge { font-size:.75em; padding:1px 6px; border-radius:4px; background:var(--accent); color:#fff; }
	</style>
}
```

- [ ] **Step 2: `templ generate` + 빌드**

```bash
~/go/bin/templ generate ./internal/vault/views/...
~/go/bin/go build ./internal/vault/views/layouts/
```

- [ ] **Step 3: 커밋**

```bash
git add internal/vault/views/layouts/theme.templ internal/vault/views/layouts/theme_templ.go
git commit -m "feat(v0.2): 7-theme CSS variables + 3-zone shell grid"
```

---

### Task 16: `views/layouts/shell.templ` — 3-zone 쉘 (sidebar | main | slideover)

**Files:**
- Create: `internal/vault/views/layouts/shell.templ`
- Create: `internal/vault/views/layouts/shell_templ.go`

- [ ] **Step 1: `shell.templ` 작성**

```templ
package layouts

templ Shell(sidebar, main, slideover templ.Component) {
	<div class="shell">
		<nav>@sidebar</nav>
		<main id="main">@main</main>
		<aside id="slideover" hx-swap-oob="true">@slideover</aside>
	</div>
}

templ EmptySlideover() {
	<!-- intentionally empty; aside has no data-open -->
}
```

- [ ] **Step 2: generate + 빌드**

```bash
~/go/bin/templ generate ./internal/vault/views/...
~/go/bin/go build ./internal/vault/views/layouts/
```

- [ ] **Step 3: 커밋**

```bash
git add internal/vault/views/layouts/shell.templ internal/vault/views/layouts/shell_templ.go
git commit -m "feat(v0.2): 3-zone Shell (sidebar | main | slideover)"
```

---

### Task 17: `views/sidebar/*.templ` — 섹션 트리

**Files:**
- Create: `internal/vault/views/sidebar/sidebar.templ`
- Create: `internal/vault/views/sidebar/sidebar_templ.go`

- [ ] **Step 1: `sidebar.templ` 작성**

```templ
package sidebar

import "github.com/sookmook/wall-vault/internal/vault"

templ Sidebar(services []vault.Service, clients []vault.Client) {
	<div class="sidebar">
		<div class="brand"><b>🔐 wall-vault</b> <small>v0.2.0</small></div>
		<div class="section-label">SERVICES</div>
		for _, s := range services {
			<a href="#" hx-get={"/hx/services/" + s.ID + "/edit"} hx-target="#slideover" hx-swap="outerHTML">{s.Name}</a>
		}
		<div class="section-label">AGENTS</div>
		for _, c := range clients {
			<a href="#" hx-get={"/hx/clients/" + c.ID + "/edit"} hx-target="#slideover" hx-swap="outerHTML">{c.Name}</a>
		}
		<div class="section-label">KEYS · MONITOR</div>
		<a href="#" hx-get="/hx/keys/list" hx-target="#main">Keys</a>
	</div>
}
```

- [ ] **Step 2: generate + 빌드 + 컴파일**

```bash
~/go/bin/templ generate ./internal/vault/views/...
~/go/bin/go build ./internal/vault/views/sidebar/
```

- [ ] **Step 3: 커밋**

```bash
git add internal/vault/views/sidebar/
git commit -m "feat(v0.2): sidebar templ component (services/agents/keys trees)"
```

---

## Stage 6 — UI Main & Slideover Components

### Task 18: `views/main/service_card.templ` + `services_grid.templ`

**Files:**
- Create: `internal/vault/views/main/service_card.templ`
- Create: `internal/vault/views/main/services_grid.templ`

- [ ] **Step 1: `service_card.templ`**

```templ
package main

import "github.com/sookmook/wall-vault/internal/vault"

templ ServiceCard(s vault.Service) {
	<div class="card" hx-get={"/hx/services/" + s.ID + "/edit"} hx-target="#slideover" hx-swap="outerHTML">
		<div style="display:flex;justify-content:space-between">
			<b>{s.Name}</b>
			if s.ProxyEnabled { <span class="badge">proxy</span> }
		</div>
		<div style="font-size:.85em;color:var(--muted);margin-top:4px">
			{s.DefaultModel} · #{fmt.Sprint(s.SortOrder)}
		</div>
	</div>
}
```

- [ ] **Step 2: `services_grid.templ`**

```templ
package main

import "github.com/sookmook/wall-vault/internal/vault"

templ ServicesGrid(services []vault.Service) {
	<section>
		<h3>Services · {fmt.Sprint(len(services))}</h3>
		<div style="display:grid;grid-template-columns:repeat(auto-fill,minmax(240px,1fr));gap:8px">
			for _, s := range services {
				@ServiceCard(s)
			}
		</div>
	</section>
}
```

- [ ] **Step 3: generate + 빌드**

```bash
~/go/bin/templ generate ./internal/vault/views/...
~/go/bin/go build ./internal/vault/views/main/
```

- [ ] **Step 4: 커밋**

```bash
git add internal/vault/views/main/service_card.templ internal/vault/views/main/services_grid.templ internal/vault/views/main/*_templ.go
git commit -m "feat(v0.2): services grid + card templ components"
```

---

### Task 19: `views/main/agent_card.templ` + `agents_grid.templ`

**Files:**
- Create: `internal/vault/views/main/agent_card.templ`
- Create: `internal/vault/views/main/agents_grid.templ`

- [ ] **Step 1: `agent_card.templ`**

```templ
package main

import "github.com/sookmook/wall-vault/internal/vault"

templ AgentCard(c vault.Client) {
	<div class="card" hx-get={"/hx/clients/" + c.ID + "/edit"} hx-target="#slideover" hx-swap="outerHTML">
		<div style="display:flex;justify-content:space-between">
			<b>🤖 {c.Name}</b>
			<small style="color:var(--muted)">{c.AgentType}</small>
		</div>
		<div style="font-size:.85em;color:var(--muted);margin-top:4px">
			→ {c.PreferredService}
			if c.ModelOverride != "" { · override: {c.ModelOverride} }
		</div>
	</div>
}
```

- [ ] **Step 2: `agents_grid.templ`**

```templ
package main

import "github.com/sookmook/wall-vault/internal/vault"

templ AgentsGrid(clients []vault.Client) {
	<section>
		<h3>Agents · {fmt.Sprint(len(clients))}</h3>
		<div style="display:grid;grid-template-columns:repeat(auto-fill,minmax(240px,1fr));gap:8px">
			for _, c := range clients {
				@AgentCard(c)
			}
		</div>
	</section>
}
```

- [ ] **Step 3: generate + 빌드 + 커밋**

```bash
~/go/bin/templ generate ./internal/vault/views/...
~/go/bin/go build ./internal/vault/views/main/
git add internal/vault/views/main/agent_card.templ internal/vault/views/main/agents_grid.templ internal/vault/views/main/*_templ.go
git commit -m "feat(v0.2): agents grid + card templ components"
```

---

### Task 20: `views/slideover/service_edit.templ` — 서비스 편집 폼

**Files:**
- Create: `internal/vault/views/slideover/slideover.templ`
- Create: `internal/vault/views/slideover/service_edit.templ`

- [ ] **Step 1: `slideover.templ` (외피)**

```templ
package slideover

templ Frame(title string, body templ.Component) {
	<aside id="slideover" data-open="true" hx-swap-oob="true">
		<div style="display:flex;justify-content:space-between;align-items:center">
			<b>✏︎ {title}</b>
			<button hx-get="/hx/slideover/close" hx-target="#slideover" hx-swap="outerHTML">✕</button>
		</div>
		<hr>
		@body
	</aside>
}

templ Empty() {
	<aside id="slideover" hx-swap-oob="true"></aside>
}
```

- [ ] **Step 2: `service_edit.templ`**

```templ
package slideover

import "github.com/sookmook/wall-vault/internal/vault"
import "strings"

templ ServiceEdit(s vault.Service) {
	<form hx-put={"/admin/services/" + s.ID} hx-ext="json-enc" hx-swap="none">
		<label>Name <input name="name" value={s.Name}></label>
		<label>Default Model <input name="default_model" value={s.DefaultModel}></label>
		<label>Local URL <input name="local_url" value={s.LocalURL}></label>
		<label>Fallback Order <input name="sort_order" type="number" value={fmt.Sprint(s.SortOrder)}></label>
		<label>Allowed Models <textarea name="allowed_models" placeholder="one per line">{strings.Join(s.AllowedModels, "\n")}</textarea></label>
		<label><input type="checkbox" name="proxy_enabled" checked?={s.ProxyEnabled}> Proxy Enabled</label>
		<button type="submit">Save</button>
	</form>
}
```

- [ ] **Step 3: generate + 빌드 + 커밋**

```bash
~/go/bin/templ generate ./internal/vault/views/...
~/go/bin/go build ./internal/vault/views/slideover/
git add internal/vault/views/slideover/ 
git commit -m "feat(v0.2): slideover frame + service edit form templ"
```

---

### Task 21: `views/slideover/client_edit.templ` — agent 편집 폼

**Files:**
- Create: `internal/vault/views/slideover/client_edit.templ`

- [ ] **Step 1: `client_edit.templ`**

```templ
package slideover

import "github.com/sookmook/wall-vault/internal/vault"

templ ClientEdit(c vault.Client, services []vault.Service) {
	<form hx-put={"/admin/clients/" + c.ID} hx-ext="json-enc" hx-swap="none">
		<label>Name <input name="name" value={c.Name}></label>
		<label>Agent Type <input name="agent_type" value={c.AgentType}></label>
		<label>Preferred Service
			<select name="preferred_service">
				for _, s := range services {
					<option value={s.ID} selected?={s.ID == c.PreferredService}>{s.Name}</option>
				}
			</select>
		</label>
		<label>Model Override <input name="model_override" value={c.ModelOverride} placeholder="empty → use service default"></label>
		<label><input type="checkbox" name="enabled" checked?={c.Enabled}> Enabled</label>
		<button type="submit">Save</button>
	</form>
}
```

- [ ] **Step 2: generate + 빌드 + 커밋**

```bash
~/go/bin/templ generate ./internal/vault/views/...
~/go/bin/go build ./internal/vault/views/slideover/
git add internal/vault/views/slideover/client_edit.templ internal/vault/views/slideover/client_edit_templ.go
git commit -m "feat(v0.2): client edit form templ with preferred_service + model_override"
```

---

### Task 22: `/hx/*` 핸들러 실제 구현 (Task 13의 스텁 교체)

**Files:**
- Modify: `internal/vault/hx_router.go`

- [ ] **Step 1: 스텁을 실 핸들러로 교체**

```go
func (s *Server) hxSidebar(w http.ResponseWriter, r *http.Request) {
	sidebar.Sidebar(s.store.ListServices(), s.store.ListClients()).Render(r.Context(), w)
}
func (s *Server) hxServicesGrid(w http.ResponseWriter, r *http.Request) {
	main.ServicesGrid(s.store.ListServices()).Render(r.Context(), w)
}
func (s *Server) hxAgentsGrid(w http.ResponseWriter, r *http.Request) {
	main.AgentsGrid(s.store.ListClients()).Render(r.Context(), w)
}
func (s *Server) hxServiceSubroute(w http.ResponseWriter, r *http.Request) {
	id := strings.TrimPrefix(r.URL.Path, "/hx/services/")
	id = strings.TrimSuffix(id, "/edit")
	svc := s.store.GetService(id)
	slideover.Frame(svc.Name, slideover.ServiceEdit(svc)).Render(r.Context(), w)
}
func (s *Server) hxClientSubroute(w http.ResponseWriter, r *http.Request) {
	id := strings.TrimPrefix(r.URL.Path, "/hx/clients/")
	id = strings.TrimSuffix(id, "/edit")
	c := s.store.GetClient(id)
	slideover.Frame(c.Name, slideover.ClientEdit(c, s.store.ListServices())).Render(r.Context(), w)
}
```

`RegisterHXRoutes` 에서 기존 `hxNotImplemented` 핸들러들을 실제 함수로 교체.

- [ ] **Step 2: 단순 smoke test — `/hx/services/grid` 200 OK + 본문에 서비스 ID 포함**

```go
func TestHXServicesGridRendersGoogle(t *testing.T) {
	srv := newTestServer(t)
	srv.Store().UpsertService(vault.Service{ID: "google", Name: "Google Gemini", DefaultModel: "gemini-3.1-pro-preview"})
	req := httptest.NewRequest("GET", "/hx/services/grid", nil)
	w := httptest.NewRecorder()
	srv.Handler().ServeHTTP(w, req)
	if w.Code != 200 { t.Fatalf("got %d", w.Code) }
	if !strings.Contains(w.Body.String(), "Google Gemini") {
		t.Fatalf("body missing service name: %s", w.Body.String())
	}
}
```

- [ ] **Step 3: 테스트 PASS + 커밋**

```bash
~/go/bin/go test ./internal/vault/ -run TestHXServicesGridRendersGoogle -v
git add internal/vault/hx_router.go internal/vault/server_test.go
git commit -m "feat(v0.2): implement /hx/* fragment handlers backed by templ components"
```

---

### Task 23: `/ (dashboard home)` — Shell + Sidebar + ServicesGrid + AgentsGrid

**Files:**
- Modify: `internal/vault/server.go` (handler for `/`)

- [ ] **Step 1: `/` 루트 핸들러 구현**

```go
func (s *Server) handleDashboard(w http.ResponseWriter, r *http.Request) {
	theme := s.store.Theme()
	services := s.store.ListServices()
	clients := s.store.ListClients()
	mainInner := templ.Join(
		main.ServicesGrid(services),
		main.AgentsGrid(clients),
	)
	layouts.Base(theme, layouts.Shell(
		sidebar.Sidebar(services, clients),
		mainInner,
		slideover.Empty(),
	)).Render(r.Context(), w)
}
```

(`templ.Join` 이 존재하지 않으면 `layouts.Main2(a,b)` 같은 2-slot 컴포넌트로 감쌈.)

- [ ] **Step 2: 라우트 등록 + smoke test**

```go
func TestDashboardHomeContainsServicesAndAgents(t *testing.T) {
	srv := newTestServer(t)
	srv.Store().UpsertService(vault.Service{ID: "google", Name: "Google"})
	srv.Store().UpsertClient(vault.Client{ID: "mini9", Name: "작순이", PreferredService: "google"})
	req := httptest.NewRequest("GET", "/", nil)
	w := httptest.NewRecorder()
	srv.Handler().ServeHTTP(w, req)
	body := w.Body.String()
	for _, want := range []string{"Google", "작순이", "htmx.org"} {
		if !strings.Contains(body, want) {
			t.Fatalf("body missing %q", want)
		}
	}
}
```

- [ ] **Step 3: 테스트 PASS + 커밋**

```bash
~/go/bin/go test ./internal/vault/ -run TestDashboardHome -v
git add internal/vault/server.go internal/vault/server_test.go
git commit -m "feat(v0.2): dashboard home renders Shell(sidebar, main, empty slideover)"
```

---

### Task 24: 구 `internal/vault/ui.go` 삭제

**Files:**
- Delete: `internal/vault/ui.go` (2,572줄)

- [ ] **Step 1: 파일 삭제 + 참조부 확인**

```bash
git rm internal/vault/ui.go
~/go/bin/go build ./... 2>&1 | head
```

Expected: 이전에 `ui.go` 의 함수들을 호출하던 `server.go` 일부가 에러. 그 호출부는 이미 Task 23 등에서 templ 핸들러로 대체되었어야 함. 남아있는 import/함수 참조를 찾아 제거.

- [ ] **Step 2: 참조 제거 후 빌드 OK**

```bash
~/go/bin/go build ./...
```

- [ ] **Step 3: 전체 테스트 수트**

```bash
~/go/bin/go test ./... -v 2>&1 | tail -30
```

Expected: 모든 테스트 PASS (또는 필요 시 구 ui 기반 테스트 몇 개 삭제).

- [ ] **Step 4: 커밋**

```bash
git add internal/vault/
git commit -m "refactor(v0.2): remove legacy 2572-line ui.go (replaced by views/ templ tree)"
```

---

## Stage 7 — Finalization

### Task 25: Makefile 정리 + deploy 타겟에 pre-boot backup 단계 추가

**Files:**
- Modify: `Makefile`, `Makefile.local.example`

- [ ] **Step 1: `Makefile` 확인 — `templ-generate` 가 build 전에 실행되는지**

```bash
grep -A2 "^build:" Makefile
```

- [ ] **Step 2: `Makefile.local.example` 의 `deploy-mini` / `deploy-raspi` / `deploy-local` / `deploy-jaksooni` 에 migration backup 확인 스텝 추가**

각 deploy 타겟의 "stopping services" 직전에:

```makefile
	@echo "▶ [mini] pre-v0.2 vault.json backup confirmation..."
	ssh $(MINI_HOST) 'ls -la ~/.wall-vault/data/vault.json || true'
	@read -p "Continue with v0.2 deploy? Migration will auto-backup to vault.json.pre-v02.{ts}.bak. [y/N] " confirm; \
	  if [ "$$confirm" != "y" ]; then echo "aborted"; exit 1; fi
```

- [ ] **Step 3: 커밋**

```bash
git add Makefile Makefile.local.example
git commit -m "build(v0.2): deploy targets confirm vault.json before auto-migration"
```

---

### Task 26: `CHANGELOG.md` v0.2.0 섹션

**Files:**
- Modify: `CHANGELOG.md`

- [ ] **Step 1: `CHANGELOG.md` 맨 위 (v0.1.29 바로 위)에 v0.2.0 섹션 삽입**

```markdown
## [0.2.0] — 2026-04-TBD

### BREAKING CHANGES

- **Service-Model Registry**: `Service` now owns `default_model` and the
  optional `allowed_models` whitelist. `Client.default_service` renamed
  to `preferred_service`; `Client.default_model` renamed to
  `model_override` (optional). Each fallback step in dispatch applies
  the destination service's own default model, eliminating the entire
  class of "model not found in Ollama" cascades.
- **Admin API bodies**: request/response schemas for `/admin/services*`
  and `/admin/clients*` changed to match the new data model. Paths stay
  the same. Old CLI or curl scripts using `default_service` /
  `default_model` on clients will break — update to `preferred_service`
  and `model_override`.
- **Dashboard UI**: legacy server-rendered `ui.go` is gone. New
  one-screen hybrid layout (sidebar / card grid / slideover detail) is
  built with Go `templ` + HTMX. HTMX fragment endpoints live under
  `/hx/*`.

### Migration

- First v0.2 startup auto-migrates the encrypted `vault.json`:
  majority-vote per service gets the new `default_model`, ties broken
  by the client with the lowest `sort_order`. A forced backup copy
  `vault.json.pre-v02.{ISO-UTC}.bak` is written before any rewrite.

### Internals

- `dispatch()` rewritten to resolve model per service via `ResolveModel`.
  Ollama name-mismatch heuristic (v0.1.27) removed.
- `templ` v0.2.747 pinned; `templ generate` runs as part of `make build`.
  Generated `*_templ.go` files are committed.
```

- [ ] **Step 2: 커밋**

```bash
git add CHANGELOG.md
git commit -m "docs(v0.2): add 0.2.0 CHANGELOG entry (breaking + migration + internals)"
```

---

### Task 27: `docs/MANUAL.md` + 16 언어 번역 — 스키마 변경 요약 블록 추가

**Files:**
- Modify: `docs/MANUAL.md`, `docs/MANUAL.{en,ja,zh,es,fr,de,pt,ar,hi,id,th,sw,ha,zu,ne,mn}.md`

- [ ] **Step 1: 한국어 원본에 "v0.2 Upgrade Notes" 섹션 추가**

```markdown
## v0.2 업그레이드 안내

- `Service` 에 `default_model` 과 `allowed_models` 가 생겼습니다. 서비스별 기본 모델은 이제 서비스 카드에서 직접 설정합니다.
- `Client.default_service` / `default_model` 은 각각 `preferred_service` / `model_override` 로 이름과 의미가 바뀌었습니다. override가 비어 있으면 서비스의 기본 모델이 사용됩니다.
- 첫 v0.2 기동 시 기존 `vault.json` 이 자동 변환되고, 변환 직전 상태는 `vault.json.pre-v02.{타임스탬프}.bak` 로 보존됩니다.
- 대시보드는 좌측 사이드바·중앙 카드 그리드·우측 편집 슬라이드오버의 세 영역으로 바뀌었습니다.
- admin API 경로는 동일하지만 요청/응답 body 스키마가 변경되어 구 CLI 스크립트는 업데이트가 필요합니다.
```

- [ ] **Step 2: 16개 언어 번역 — 각 파일에 동일 섹션 삽입 (병렬 편집)**

구 `feedback_translate_docs.md` 패턴(Agent 병렬 디스패치)에 따라 `superpowers:dispatching-parallel-agents` 스킬로 16개 에이전트 병렬 번역 + 커밋. 또는 단일 턴으로 진행 시:

```bash
for lang in en ja zh es fr de pt ar hi id th sw ha zu ne mn; do
  # Each agent translates the Korean block above into MANUAL.$lang.md.
  # Invoke via the dispatch-parallel-agents flow.
done
```

- [ ] **Step 3: 한국어 + 영어 먼저 커밋, 나머지는 병렬 완료 후 2차 커밋**

```bash
git add docs/MANUAL.md docs/MANUAL.en.md
git commit -m "docs(v0.2): Korean + English upgrade notes"

# (after parallel agents finish)
git add docs/MANUAL.{ja,zh,es,fr,de,pt,ar,hi,id,th,sw,ha,zu,ne,mn}.md
git commit -m "docs(v0.2): translate upgrade notes to 14 remaining languages"
```

---

## Stage 8 — Release Candidate & Cutover

### Task 28: `v0.2-redesign` → `main` merge + tag `v0.2.0-rc1`

**Files:** 없음 (git 조작만)

- [ ] **Step 1: 전체 테스트 + 빌드 최종 확인**

```bash
~/go/bin/go test ./... -v 2>&1 | tail -10
make build
./bin/wall-vault version
```

Expected: 모든 테스트 PASS, 버전 `v0.2.0.{timestamp}`.

- [ ] **Step 2: main merge (fast-forward 아닌 경우 merge commit)**

```bash
git checkout main
git merge --no-ff v0.2-redesign -m "Merge v0.2-redesign into main for v0.2.0-rc1"
git tag -a v0.2.0-rc1 -m "v0.2.0 release candidate 1"
git push origin main
git push origin v0.2.0-rc1
```

---

### Task 29: 미니 vault+proxy cutover (`deploy-mini` + migration 확인)

**Files:** 없음 (운영)

- [ ] **Step 1: mini 바이너리 사전 백업 확인 후 배포**

```bash
ssh 192.168.0.6 'ls -la ~/.openclaw/wall-vault.bak* 2>/dev/null | tail'
make deploy-mini 2>&1 | tee /tmp/deploy-mini-v0.2.log
```

Expected: 배포 스크립트가 vault.json backup 프롬프트 → y 입력 → 정상 완료.

- [ ] **Step 2: migration 로그 확인**

```bash
ssh 192.168.0.6 'grep "\[migrate\]" ~/.openclaw/vault.log | tail -10 || ~/.wall-vault/vault.log'
```

Expected: `[migrate] wrote backup: .../vault.json.pre-v02.*.bak`, `[migrate] v1 → v2 complete; services=N clients=M`.

- [ ] **Step 3: 새 버전 및 services/clients API 확인**

```bash
ssh 192.168.0.6 'curl -s http://localhost:56243/api/status'
TOKEN=dhvmszmffh
ssh 192.168.0.6 "curl -s -H 'Authorization: Bearer $TOKEN' http://localhost:56243/admin/services | python3 -m json.tool | head -30"
ssh 192.168.0.6 "curl -s -H 'Authorization: Bearer $TOKEN' http://localhost:56243/admin/clients | python3 -m json.tool | head -40"
```

Expected: `version: v0.2.0-rc1.*`, 서비스에 `default_model` 필드, 클라이언트에 `preferred_service`/`model_override`.

---

### Task 30: 라즈·모토코·작순이 proxy 순차 교체

**Files:** 없음 (운영)

- [ ] **Step 1: 순차 배포**

```bash
make deploy-raspi    && sleep 5
make deploy-local    && sleep 5
make deploy-jaksooni
```

- [ ] **Step 2: 네 proxy /status 일괄 확인**

```bash
for host in localhost 192.168.0.6 raspi; do
  curl -s http://$host:56244/status | python3 -c "import sys, json; d=json.load(sys.stdin); print(d['client'], d['version'], d['service'], d['model'])"
done
ssh -p 2244 192.168.0.4 "curl -s http://localhost:56244/status | python3 -c 'import sys,json;d=json.load(sys.stdin);print(d[\"client\"],d[\"version\"],d[\"service\"],d[\"model\"])'"
```

Expected: 네 머신 모두 `version=v0.2.0-rc1.*`, `service=google`, `model=gemini-3.1-pro-preview` (모델은 각 클라이언트의 preferred_service + 해당 service.default_model 조합).

---

### Task 31: 스모크 테스트 + tag `v0.2.0`

**Files:** 없음 (운영 + git tag)

- [ ] **Step 1: 세 봇 텔레그램 스모크 테스트**

사용자 terminal에서 짧은 임무 하나씩:

- `@sookmook_Yamai_Motoko_SMPC10_bot` 에 "안녕, 한 줄로 자기소개 해줘"
- `@sookmook_Raspberry_Pi_bot` 에 동일
- `@sookmook_Mini9_bot` 에 "`curl -s http://192.168.0.6:56240/profiles/guide` 실행해서 결과 일부 보여줘" (tool_use 검증)

Expected:
- 모토코·라즈: 정상 응답
- 작순이: 실제 curl 결과가 포함된 응답 (tool_use 실행됨)

- [ ] **Step 2: 각 proxy 로그에서 `All models failed` / `model 'gemini-3.1-pro-preview' not found` 부재 확인**

```bash
ssh 192.168.0.6 'journalctl --user -u wall-vault-proxy --since "5 minutes ago" | grep -E "failed|error" | head'
ssh raspi 'journalctl --user -u wall-vault-proxy --since "5 minutes ago" | grep -E "failed|error" | head'
ssh -p 2244 192.168.0.4 'journalctl --user -u wall-vault-proxy --since "5 minutes ago" | grep -E "failed|error" | head'
```

Expected: empty (에러 없음).

- [ ] **Step 3: tag `v0.2.0` + push**

```bash
git tag -a v0.2.0 -m "v0.2.0: Service-Model Registry + templ/HTMX dashboard"
git push origin v0.2.0
```

- [ ] **Step 4: rollback 파일 7일 보관 후 정리 계획 메모**

현 `~/.openclaw/wall-vault.bak.*`, `~/.wall-vault/data/vault.json.pre-v02.*.bak` 파일은 최소 7일간 보관. 그 뒤에 별도 cleanup task.

---

## Self-Review Summary

### Spec coverage
- §2 Architecture → Stage 1~3, 5, 6
- §3 Data Model → Task 4, 5
- §4 Dispatch Rules → Task 8, 9, 10
- §5 Admin API → Task 11, 12, 13
- §6 UI α Layout → Stage 5, 6
- §7 Migration → Task 6, 7
- §8 Deployment → Stage 8
- §9 Verification → 각 Task의 Step "test execute" + 전체 테스트 Task 28 Step 1

### Placeholder scan
- 코드 블록 모두 실체 존재 (재사용 가능한 함수 시그니처 포함)
- i18n 16개 번역 Task는 "parallel-agents 스킬 호출"로 위임하지만, 위임할 대상 스킬 이름과 패턴을 명시했으므로 placeholder 아님
- `newTestServer(t)`, `writePlain` 테스트 helper는 store_test의 실존 헬퍼를 재사용 가정 — 실제 이름은 실행자가 `grep` 으로 확인하는 한 줄 추가됨

### Type consistency
- `Service` / `Client` 필드명은 Stage 1에서 정의 후 전 plan에서 동일하게 사용
- `ResolveModel` 시그니처 `(vault.Client, vault.Service) (string, error)` 는 Task 8/9/12에서 동일
- `dispatchWith` / `dispatchForTest` 는 Task 9에서 정의 후 다른 task에서 참조 없음

---

## Execution Handoff

Plan complete and saved to `docs/superpowers/plans/2026-04-13-wall-vault-v02-redesign.md`.

**두 실행 옵션:**

1. **Subagent-Driven (추천)** — 매 Task마다 신선한 subagent 디스패치, 사용자 체크포인트, 빠른 iteration. `superpowers:subagent-driven-development` 스킬 사용.
2. **Inline Execution** — 이 세션에서 순차 실행, 몇 Task마다 체크포인트. `superpowers:executing-plans` 스킬 사용.

어느 방식으로 진행할까요?
