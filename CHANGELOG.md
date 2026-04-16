# Changelog

All notable changes to wall-vault are documented here.
Format follows [Keep a Changelog](https://keepachangelog.com/en/1.0.0/).

wall-vault의 모든 주요 변경 사항을 기록합니다.
형식은 [Keep a Changelog](https://keepachangelog.com/ko/1.0.0/)를 따릅니다.

---

## [0.2.9] — 2026-04-17

Surface the **active model** in `GET /status` so polling consumers
(EconoWorld analyzer, dashboards) can tell which model the proxy
will actually send upstream — not just whether the client happens
to have a `model_override`.

### Fixed

- **`/status` `model` field no longer empty when `model_override` is
  unset**: the field now falls back to the vault-synced
  `serviceDefaults[service]` (the current service's
  `default_model`). The raw `model_override` still wins when set, so
  an explicit override always shows through. Test:
  `TestHandleStatus_ModelFallsBackToServiceDefault`.

Requested by the EconoWorld analyzer team (option 1 of three
proposed) — their analyzer polls `/status` every 5 s and treated an
empty `model` as a signal that the proxy was mid-reconfigure.

---

## [0.2.8] — 2026-04-16

Defensive cleanup to stop OpenClaw's config validator from crashing
the gateway on a corrupt `custom.models` entry left over from older
proxy writes.

### Fixed

- **`agent_apply.go` sanitizes `models.providers.custom.models` on
  every write**: any entry with an empty `id` is dropped before the
  current update runs. A stale `{"id":""}` entry (from a pre-guard
  version of this function, or an external editor) would otherwise
  trip OpenClaw's zod validator with
  `models.providers.custom.models.1.id: Too small: expected string to
  have >=1 characters` and push the gateway into a SIGTERM crash
  loop. Observed on mini after the v0.2.7 deploy wave — `openclaw.json`
  had a bare `{"id":""}` at index 1 that prevented the Telegram bot
  from ever starting.
- No behavior change for a clean config; the only side effect is
  that a previously-bad entry gets silently removed the next time
  the proxy pushes any client config.

---

## [0.2.7] — 2026-04-16

Stop trapping admins when a client's `model_override` is stale.

### Fixed

- **`PUT /admin/clients/{id}` no longer 422s on stale `model_override`**:
  previously, if the submitted `model_override` wasn't in the target
  service's `allowed_models`, the whole update was rejected. This
  created a trap: switching an agent's `preferred_service` from
  `google` to `openrouter` (which uses prefixed model IDs like
  `google/gemini-3.1-flash-lite-preview`) would fail, because the
  unprefixed override carried over and no longer matched. The admin
  had to manually wipe `model_override` before any other edit could
  go through. The handler now silently clears the override in this
  case — dispatch already falls back to `DefaultModel` per
  `proxy.ResolveModel` invariant — and logs the event for audit:
  `vault: client=X model_override=Y cleared on save (not in service=Z allowed_models)`.
- Updated the corresponding test
  (`TestAdminPutClient_ModelOverrideWhitelistViolationIsSoftCleared`)
  to assert the new 200-OK-with-cleared-override contract.

### Added

- **Stale-override warning in the agent edit form**: when a saved
  `model_override` is not found in the current service's default or
  allowed_models, `client_edit.templ` now decorates the "(현재 값)"
  option as `(⚠ 현재 값 · 저장 시 초기화)` and shows an amber hint
  line below the dropdown spelling out what will happen on save.
- New `.wv-stale-override` style in `theme.templ` (fixed amber
  `#b45309` across all 7 themes so the warning stays visible).

---

## [0.2.6] — 2026-04-16

Service edit gains a **catalog picker** so admins can populate
`allowed_models` by checkbox instead of typing model IDs by hand —
closes the v0.2.4 asymmetry where the only way to add curated models
was to know their IDs in advance.

### Added

- **카탈로그에서 추가 picker** in the service edit slideover: a
  `<details>` block listing every registry-known model that is NOT
  already in `default_model` or `allowed_models`, each with a
  checkbox. "선택 항목 허용 목록에 추가" appends the checked rows to
  the `allowed_models` textarea (skipping dupes) and hides the
  just-added rows. "체크 해제" clears the selection.
- **Inline search filter** over the catalog list. Typing in the
  search box hides any row whose ID doesn't contain the (case-
  insensitive) query — makes the OpenRouter list (341 items)
  usable.
- Catalog picker is scrollable (max-height 220px) and row labels
  render in a monospace font for readability.

### Internal

- `ServiceVM` gains `CatalogUnused []string` (registry minus
  default_model and AllowedModels).
- `toSlideoverService` computes `CatalogUnused` once per slideover
  render using the already-`ensureRegistry()`'d catalog.
- New click handler in `base.templ` dispatches on
  `[data-wv-add-catalog]` / `[data-wv-catalog-clear]`; a separate
  input handler dispatches on `[data-wv-catalog-filter]`.
- New styles in `theme.templ`: `.wv-catalog-picker`,
  `.wv-catalog-filter`, `.wv-catalog-list`, `.wv-catalog-item`,
  `.wv-catalog-actions`.

---

## [0.2.5] — 2026-04-16

Server-render the agent model_override dropdown so users see the full
list (기본 + 허용 목록) immediately, without depending on a JS hydrate
that sometimes doesn't run in time after an HTMX OOB slideover swap.

### Fixed

- **Dropdown showing only 1 model after opening the edit slideover**:
  the initial `<select>` used to ship with just a blank option plus the
  current `model_override` value, relying on `wvInitModelOverride` to
  fill in the optgroups on `htmx:afterSwap` / `htmx:oobAfterSwap`. When
  the swap used OOB semantics the hydrator occasionally never applied
  to the new form — the user then saw only one model, making the
  feature look broken. `ClientEdit` now renders `<optgroup>기본</optgroup>`
  and `<optgroup>허용 목록</optgroup>` server-side from
  `ClientVM.CurrentGroup`, so the full list is present before any JS
  runs. The existing JS still updates the list when `preferred_service`
  changes.

### Internal

- `ClientVM` gains `CurrentGroup ServiceModelGroup` (pre-resolved
  group for `PreferredService`) and an `OverrideInCurrentGroup()`
  helper that suppresses duplicate "(현재 값)" rendering when the
  saved value already appears in the default/allowed list.

---

## [0.2.4] — 2026-04-16

Scope correction for the agent `model_override` dropdown: drop the
registry catalog that v0.2.3 added, restore admin curation via
`allowed_models` only.

### Changed

- **Remove 카탈로그 optgroup from agent model_override**: the
  auto-populated provider-registry list introduced in v0.2.3 is gone.
  The dropdown now shows only the `기본` (service default_model) and
  `허용 목록` (admin-curated allowed_models) groups, filtered by the
  current preferred_service. Admins who want more choices add them via
  the service edit page's `allowed_models` textarea, keeping the
  agent-facing list curated rather than dumping the entire provider
  catalog into every agent form.

### Internal

- `ServiceModelGroup` drops the `Catalog` field.
- `toSlideoverClient` no longer calls `ensureRegistry()` or
  `registry.All()` — only default + allowed_models are plumbed through.
- `wvInitModelOverride` renders two optgroups instead of three.

---

## [0.2.3] — 2026-04-14

UX fix: agent `model_override` dropdown now surfaces the full service model
catalog instead of just the one `default_model`.

### Fixed

- **Empty agent model dropdown**: the agent edit slideover's
  `model_override` `<select>` was previously fed only
  `default_model + allowed_models` from the vault. In most deployments
  `allowed_models` is empty, so the dropdown collapsed to a single
  (default) option — indistinguishable from "not working" for users.
  Now `toSlideoverClient` also pulls `s.registry.All(svc)`, so every
  model known to the provider (Google 15, OpenRouter 345, etc.) is
  available. Service edit already behaved this way; agent edit now
  matches.

### Changed

- `ClientVM.ServiceModelMap` is now
  `map[string]ServiceModelGroup` where `ServiceModelGroup` splits models
  into `{default, allowed[], catalog[]}`. `wvInitModelOverride` renders
  the three buckets as separate `<optgroup>`s (`기본` / `허용 목록` /
  `카탈로그`) so users can tell admin-curated choices from registry
  overflow. Dedup preserves priority (default → allowed → catalog).

### Internal

- `toSlideoverClient` now calls `ensureRegistry()` before building the
  map — same cold-cache path as service edit.
- `serviceModelJSON(map[string]ServiceModelGroup) string` replaces the
  old flat-list marshaller; `ClientCreate` signature updated to match.

---

## [0.2.2] — 2026-04-16

Audit-driven polish: dispatch reliability, model-selection UX, ClientInput
v0.2 field migration, and documentation refresh.

### Fixed

- **Dispatch fast-skip for cooled services**: `KeyManager.CanServe(svc)`
  predicate lets dispatch bypass cloud services whose keys are all on
  cooldown/exhausted, eliminating the forced-retry loops that caused
  ~15 s caller timeouts. Local services (ollama/lmstudio/vllm) are
  always tried.
- **Dispatch fallback model swap**: each fallback step now applies the
  target service's `default_model` (synced from vault via the new
  `ProxyService.DefaultModel` field). Previously fallback sent the
  caller's original model name to every service (e.g. `gemini-2.5-flash`
  to Anthropic → 400, then Ollama → 404).
- **Anthropic 400 "credit balance" → cooldown**: Anthropic returns HTTP
  400 (not 402) when the account balance is depleted. Detect
  "credit balance" / "billing" in the 400 body and promote to 402-level
  30 min cooldown so subsequent dispatches fast-skip.
- **Service edit default_model dropdown — server-render full list**:
  previously the `<select>` shipped with only the current value, relying
  on a second `/admin/models` round-trip to populate. Cold cache / OOB
  swap edge cases left the dropdown with a single option, effectively
  locking the user out of changing models. New `Server.ensureRegistry()`
  refreshes the registry before rendering the slideover, so the HTML
  now arrives with every available model pre-populated (Google 15,
  OpenRouter 345, Anthropic 6, etc.).
- **OOB swap hydration**: htmx doesn't fire `htmx:afterSwap` on
  OOB-swapped nodes. All four hydrate helpers (`wvHydrateModels`,
  `wvInitModelOverride`, `wvInitReorder`, `wvHydrateProgress`) now also
  listen on `htmx:oobAfterSwap` + `htmx:afterOnLoad`. Model refresh,
  drag handles, and progress bars re-initialise on slideover open.

### Added

- **Service edit — default_model swap UX**:
  - `↓ 허용으로` / `↓ Move to Allowed` button demotes the current
    default_model into the `allowed_models` textarea on click
  - `✕ 지움` / `✕ Clear` button empties the default_model in place
  - Collapsible `직접 입력` / `Custom input` details block as a
    fallback when the dropdown can't be populated (offline / registry
    failure). Submit-time override logic swaps the custom value into
    the serialised JSON.
- **Agent edit/create — model_override dropdown**: `<input>` replaced
  by a `<select>` populated from `ClientVM.ServiceModelMap`
  (`service → [default_model, ...allowed_models]`). Changing
  preferred_service auto-repopulates the override select. Free-text
  values kept via auto-added "(현재 값)" option so legacy records
  stay round-trippable.

### Changed

- `ClientInput` (POST /admin/clients) now accepts v0.2 canonical fields
  `preferred_service` / `model_override` alongside legacy
  `default_service` / `default_model`. V0.2 takes precedence.
  `EffectiveService()` / `EffectiveModel()` helpers unify the two.
- Ollama `default_model` configurable via dashboard; previously
  hard-coded in dispatch.
- `toSlideoverClient` becomes a `Server` method so it can fan out into
  `store.ListServices()` when assembling `ServiceModelMap`.

### Internal

- CLAUDE.md refreshed: views/ architecture, multimodal pass-through,
  EconoWorld agent type, dispatch fast-skip / fallback model swap,
  CanServe predicate.
- Full codebase audit: no critical bugs; dispatch_v2.go remains test-only
  harness (design decision); 5 TODOs for Stage 2 legacy field removal
  tracked.

---

## [0.2.1] — 2026-04-16

Post-RC1 polish round: dashboard becomes substantially more capable, and the
proxy gains multimodal pass-through so OpenAI-format clients (e.g. EconoWorld)
can stream audio / video / image / arbitrary file blobs straight to Gemini
without the proxy stripping them.

### Added — Proxy

- **Multimodal pass-through (OpenAI → Gemini `inlineData`)**:
  `OpenAIToGemini` now recognises six content part types — `text`,
  `input_audio`, `input_video`, `input_image`, `input_file`, and
  `image_url` (data URI). Each maps to a `GeminiPart` with a `BlobData`
  carrying `{mimeType, data}`. Audio/video/image format-to-mime helpers
  cover wav/mp3/ogg/flac/webm/m4a, mp4/mov/webm/mkv/avi, and
  jpg/png/gif/webp/heic respectively. Body limit `maxAIBodySize=50 MB`
  is unchanged. External http(s) URLs in `image_url` remain unsupported
  in this round (data URI only).
- **EconoWorld agent type** (`agentType: "econoworld"`) for
  `POST /agent/apply`: writes wall-vault settings into the project's
  `analyzer/ai_config.json` (`provider` flipped to `openai_compatible`,
  base URL / api_key / model populated). Other provider sections are
  preserved (partial merge). `workDir` accepts a comma-separated list of
  candidate paths and picks the first whose `analyzer/` directory
  exists; Windows drive paths are converted to WSL mounts.

### Added — Dashboard

- **Header bar**: logo image, Korean title, version stamp, theme switcher
  (7 themes), language switcher (17 locales auto-discovered).
- **Footer**: GitHub / sookmook.org / email links + live uptime ticker
  reading `data-wv-started` and refreshing every second.
- **Service / agent / key cards** redesigned: chips, status dots
  (active / cooldown / off / online / ready), avatar previews, key usage
  progress bars (rendered via `[data-pct]` JS hydration to side-step
  templ's strict `style={…}` parser).
- **Drag-and-drop reorder** for agent cards via native HTML5 DnD with a
  dedicated `⋮⋮` handle so plain card clicks still open the slideover.
  PUTs `/admin/clients/reorder` on drop.
- **Keys grid UI** (replacing the "Keys list UI coming in a later round."
  placeholder): 11 keys render as compact cards with status dots,
  per-service usage / limit, attempt counter, cooldown remaining label,
  and progress bar when limit is set.
- **Per-theme animation layer** ported from v0.1: cherry petals
  (per-petal zigzag keyframes), ocean (3 wave bands + drifting clouds +
  rising bubbles + sparkles), gold (32 ✦✧⋆ twinkle/drift), autumn
  (28 leaves with per-leaf rotation), winter (snowman/tree/snowflake +
  20 falling flakes). Layer sits at `z-index:0`, `.shell` lifted to
  `z-index:1` so the effect plays between background and cards.
- **Slideover form polish**: 4-section layout (Basic / AI Routing /
  Access / Appearance) for client edit + create; 3-section (Basic /
  Routing / Advanced) for service edit. Each field gets a labelled
  `<small class="hint">` tip. Service edit drops the `local_url` input
  for cloud services and presents `default_model` as a `<select>` with
  `<optgroup>` separating Free / Paid (filled async via
  `/admin/models?service=…`).
- **Avatar upload**: file input + live 64px preview, client-side
  Canvas-resize to ≤256px, embedded as data URI in the JSON body.
- **Native fetch form submit**: forms carry `data-wv-submit="<URL>"` /
  `data-wv-method="POST|PUT"`; a delegated `submit` listener serialises
  FormData to JSON, attaches `Authorization: Bearer <admin_token>`, and
  reloads on success / alerts on failure. All HTMX requests gain the
  same auth header via `htmx:configRequest`.

### Changed

- `ClientInput.IPWhitelist` and `…AllowedServices` use a new
  `StringOrList` JSON type that accepts both arrays and comma-separated
  strings — single-line dashboard inputs are normalised on the server
  side.
- Sidebar Keys link is now an in-page anchor (`#keys`) since the keys
  grid renders inline below Services and Agents.
- Sidebar widened to 260px and section headers separated by a divider
  for readability.
- Card grids set to fixed 2 columns (single column under 900px).

### Fixed

- v0.2 migration `MigrateV1ToV2` previously reused the legacy `keys`
  field name in `api_keys`, leading to all encrypted API keys being
  dropped on upgrade. Migrator now reads `keys` correctly; affected
  installs can manually merge from `vault.json.pre-v02.*.bak`.
- Service `enabled` flag was being toggled off during migration in some
  vault states, causing dispatch to fall through to Ollama with the
  wrong model. Re-enabling via `PUT /admin/services/{id}` is the
  manual fix for already-migrated installs.

### Internal

- `i18n` locales: 40 form-label / section / placeholder / hint keys
  added to ko + en. Other 15 locales backfilled with the en value to
  satisfy `TestAllLanguagesHaveAllKeys`.
- Convert tests: 8 new unit tests exercise text-only regression,
  `input_audio`, `input_video`, image data URI, video data URI via
  `image_url`, explicit `input_video`, `audioFormatToMime`, and
  `parseDataURI`.

---

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

---

## [0.1.29] — 2026-04-13

### Fixed
- **Anthropic `/v1/messages` dispatch drops tool_use / tool_result turns**:
  `AnthropicToGemini` extracted text only (`ContentText()`), so Claude Code's
  content-block messages — which often carry `tool_use` (no text) or
  `tool_result` blocks — collapsed into empty-text Gemini parts. Every
  dispatch backend then rejected the request:
    * Google Gemini → HTTP 400 "contents is not specified"
    * OpenRouter / Ollama → HTTP 400 "[] is too short - 'messages'"
    * Anthropic fallback → "Anthropic: 변환할 메시지 없음" (0 messages after filter)
  Fix: route Anthropic → OAI intermediate → Gemini so tool blocks become
  proper `functionCall` / `functionResponse` parts and empty-content turns
  are dropped cleanly (reusing the existing `anthropicToOpenAIReq` +
  `OpenAIToGemini` path). Generation params (temperature / max_tokens)
  still applied on top.
- **Anthropic fallback turn with no plain text**: `doAnthropicRequest` used
  `extractText` per turn and produced empty-content messages (which the
  Anthropic API itself accepts but that combined with the bug above led to
  `0 messages` when all turns were tool-only). Now JSON-serializes the raw
  parts as fallback content for tool turns and skips genuinely empty turns
  before the "변환할 메시지 없음" guard.

### Why this only surfaced now
OpenClaw 2026.4.10's Active Memory plugin and nanoclaw's Claude Code agent
(via `ANTHROPIC_BASE_URL` pointing at wall-vault) both emit tool_use /
tool_result heavy `/v1/messages` requests far more often than the earlier
text-only chat turns, making the latent conversion bug routine instead of
edge-case.

---

## [0.1.28] — 2026-04-11

### Fixed
- **PNG avatar upload failures in agents section**: PNG files — typically
  much larger than JPG at the same resolution — frequently exceeded the 1 MB
  body limit on `POST/PUT /admin/clients`, causing silent rejections.
  Now downscaled client-side and the body limit raised specifically for
  client CRUD.
- **`.hpg` typo in extension switch**: Both `internal/vault/ui.go` and
  `internal/proxy/heartbeat.go` listed `.hpg` alongside `.jpg`/`.jpeg` —
  almost certainly a fat-fingered `.png`. Removed the non-existent extension.
  No behavior change (the default MIME was already `image/png`), but the
  dead case was misleading.

### Changed
- **Client-side avatar downscale** (`loadAvatarPreview` in `ui.go`): uploads
  are now resized to at most 256×256 via a `<canvas>` before being sent as
  a base64 data URI. PNG inputs stay PNG (transparency preserved); other
  formats are re-encoded as JPEG quality 0.9. Invalid files trigger a
  localized alert and clear the input. Dashboard renders avatars at 48×48
  and 5.28 rem anyway, so the original resolution was wasted bytes.
- **Dedicated body limit for client CRUD** (`server.go`): introduced
  `maxAvatarBodySize = 3 << 20` (3 MB). `POST /admin/clients` and
  `PUT /admin/clients/{id}` now use this limit instead of the generic
  1 MB, providing headroom even without the client-side resize (e.g. for
  direct API callers).

---

## [0.1.27] — 2026-04-09

### Fixed
- **Ollama model name mismatch on fallback**: When dispatch falls back to
  Ollama from another service (e.g. OpenRouter), the provider-prefixed model
  name (e.g. `google/gemini-3.1-pro-preview`) was sent to Ollama which doesn't
  recognize it. Now detects `/` in model name and falls back to env var or
  default Ollama model. Same fix applied to `resolveActualModel()`.

### Changed
- **Cooldown durations shortened**: 429 rate limit reduced from 30 minutes to
  5 minutes. 402 payment required from 1 hour to 30 minutes. 401/403 from 24
  hours to 6 hours. Default cooldown from 10 minutes to 5 minutes. Prevents
  total proxy lockout when all keys hit rate limits simultaneously.
- **Force-retry on total cooldown**: When all keys for a service are on
  cooldown, the proxy now clears the soonest-expiring key and retries instead
  of returning an error. Eliminates the "all services failed" dead-end.
- **Status endpoint shows actual services**: `/status` now returns
  `allowedServices` (vault-synced list) instead of `cfg.Proxy.Services`
  (static config), correctly showing anthropic and other dynamically enabled
  services.

---

## [0.1.26] — 2026-04-08

### Fixed
- **RTK filterGitLog panic**: Fixed index-out-of-bounds crash when git log
  output has lines shorter than 19 chars or commit messages before any hash.
- **Doctor nil response dereference**: Separated nil check and status code
  check to prevent panic when HTTP request fails entirely.
- **XSS in agent card status**: Service and model names in dashboard agent
  cards are now set via `textContent` instead of `innerHTML`, preventing
  potential script injection.
- **Broker broadcast race condition**: SSE broadcast now copies the client
  channel list before iterating, preventing concurrent map read/write panics.
- **Hook command goroutine leak**: Shell hook commands now have a 30-second
  context timeout, preventing indefinite goroutine accumulation.
- **Local service infinite timeout**: Ollama and LM Studio/vLLM HTTP clients
  now use 10-minute timeout instead of unbounded (Timeout: 0).
- **Silent Anthropic model fallback**: Non-Claude models routed to Anthropic
  now log a warning when falling back to claude-haiku.
- **Cache TTL comment mismatch**: Fixed misleading comment (said 30s, actual 5s).

---

## [0.1.25] — 2026-04-08

### Added
- **Agent process health detection**: Proxy heartbeat now detects whether the
  local agent process (nanoclaw via `systemctl`, openclaw via `pgrep`) is alive.
  When the agent dies while the proxy is still running, the dashboard agent card
  shows an orange pulsing traffic light with "⚠ Agent process stopped" status.
- **Drag handle on agent cards**: Drag-and-drop reordering now uses the traffic
  light dot area as the grab handle, preventing accidental drags from input
  fields and buttons.
- **i18n keys**: Added `drag_reorder` and `agent_dead` to all 17 locales.

---

## [0.1.24] — 2026-04-06

### Added
- **RTK subcommand (`wall-vault rtk`)**: Token reduction for shell command
  output, ported from RTK (Rust Token Killer) concept. Filters and compresses
  command output before it reaches the LLM context window, reducing token
  usage by 60-90% on common development operations.
  - 3-tier filter pipeline: command-specific → regex post-processing → passthrough
  - Git filters: `status` (87% reduction), `diff` (context trimming), `log`
    (hash+message only), `push/pull/fetch` (summary only)
  - Go filters: `test` (failure-focus, hide passing), `build/vet` (errors only)
  - General: passthrough with auto-truncation (head 50 + tail 50 lines, 32KB max)
  - `LC_ALL=C` forced for consistent English output parsing
  - Exit code preservation for LLM error detection
  - `wall-vault rtk rewrite` for Claude Code PreToolUse hook integration
  - Zero external dependencies (stdlib only)

---

## [0.1.23] — 2026-04-06

### Fixed
- **Ollama model change from vault dashboard had no effect**: `callOllama()`
  ignored the vault-configured model and always read from environment variables
  (`OLLAMA_MODEL` / `WV_OLLAMA_MODEL`) or hardcoded default `qwen3.5:35b`.
  Same issue in `resolveActualModel()`. Now uses the vault model first, falling
  back to env vars only when unset.

### Changed
- **Local service auto-toggle based on connectivity**: Local services (Ollama,
  LM Studio, vLLM) now auto-enable when reachable and auto-disable when
  unreachable, mirroring cloud services' key-based auto-toggle. Both the
  initial `autoCheckServices()` and the 15-second periodic ping update the
  enabled checkbox. Previously, only the status dot (●) color was updated
  while the checkbox required manual toggling (v0.1.21 behavior).

---

## [0.1.22] — 2026-04-05

### Fixed
- **Empty content field dropped in OpenAI/Anthropic responses**: When a
  thinking model (gemini-3.1-pro, claude-opus-4-thinking, o1, etc.) exhausts
  `max_tokens` on reasoning before producing visible output, the response
  carries empty text — but our proxy silently dropped the `content` /
  `text` JSON field via `omitempty`. OpenAI and Anthropic SDKs expect the
  field to always be present (per official API spec), and crashed with
  `Cannot read properties of undefined (reading 'trim')` when it was missing.
  - `OpenAIMessage.MarshalJSON` now always emits `"content":""` for empty
    assistant messages.
  - `AnthropicContent.Text` removed `omitempty` so `"text":""` is always
    emitted in text blocks.

---

## [0.1.21] — 2026-04-05

### Added
- **Gemma 4 model support**: Proxy now routes `gemma-*` prefixed models to
  Google Gemini API alongside `gemini-*` models. Added `gemma-4-31b-it` and
  `gemma-4-26b-a4b-it` to the model registry (256K context). Streaming handler
  and `parseProviderModel()` updated accordingly.
- **LM Studio / vLLM dispatch**: Added `callLocalService()` handler for
  `lmstudio` and `vllm` services in the dispatch switch. Previously these
  services were silently skipped (`default: continue`), causing all requests
  to fall back to Ollama.
- **`WV_TOOL_FILTER` environment variable**: Tool filter mode (`strip_all`,
  `whitelist`, `passthrough`) can now be set via environment variable,
  in addition to the YAML config file.

### Fixed
- **Dashboard shows fallback service instead of configured service**: Heartbeat
  reported `lastActualSvc` (e.g. Ollama fallback) instead of the user's
  configured service (e.g. LM Studio). Removed `lastActualSvc/lastActualMdl`
  tracking entirely — heartbeat now always reports the configured service/model.
- **Local service auto-probe overrides user setting**: `autoCheckServices()`
  on dashboard load pinged local services (Ollama, LM Studio, vLLM) and
  auto-disabled them if unreachable. Now `_checkLocalSvc()` only updates the
  status dot (●) color without changing the enabled checkbox or saving to server.

---

## [0.1.20] — 2026-03-28

### Security
- **Stored XSS prevention**: All user-controlled data (client names, IDs,
  descriptions, models, services, agent types, IP whitelists, key labels) is
  now HTML-escaped via `html.EscapeString` (41 injection points fixed in ui.go).
- **Constant-time token comparison**: Replaced all `==` token checks with
  `crypto/subtle.ConstantTimeCompare` to prevent timing-based token extraction
  (6 comparisons in vault server + 1 in store).
- **CORS restriction**: Changed `Access-Control-Allow-Origin` from wildcard `*`
  to whitelist: localhost, 127.0.0.1, and 192.168.* LAN origins only.
- **Request body size limits**: Added `http.MaxBytesReader` to all endpoints
  (1 MB general, 5 MB heartbeat, 50 MB AI proxy) to prevent OOM DoS.
- **Path traversal prevention**: Avatar file path now rejects `..` segments and
  absolute paths, with `filepath.Clean` + boundary verification (vault ui.go +
  proxy heartbeat.go).
- **Agent apply token validation**: `/agent/apply` endpoint now validates
  Authorization token against vault token or registered client tokens (was
  accepting any non-empty header).
- **SSE endpoint authentication**: When admin token is configured, SSE `/api/events`
  now requires valid admin or client token (via header or `?token=` query param).
- **Rate limiter hardening**: Removed X-Forwarded-For trust in `realIP()`, always
  using `r.RemoteAddr` to prevent rate-limiter bypass via spoofed headers.
- **JSON injection prevention**: Model name in `openclaw_sync.go` now serialized
  via `json.Marshal` instead of string concatenation.
- **tmux command injection prevention**: Added `sanitizeModelForTmux()` to strip
  control characters from model names before `tmux send-keys`.
- **Empty admin token warning**: Prominent multi-line warning logged at startup
  when no admin token is configured.
- **Info leak reduction**: Unauthenticated `/api/status` returns only version;
  `/api/clients` hides `agent_type` for unauthenticated callers.

---

## [0.1.19] — 2026-03-27

### Added
- **Claude Code online detection**: Proxy detects locally-running Claude Code
  processes via `pgrep -x claude` and reports them in the heartbeat's
  `ActiveClients`. Since Claude Code uses Anthropic OAuth directly (bypassing
  the proxy), it was always shown as offline on the dashboard. Now the proxy
  injects a synthetic activity entry every heartbeat cycle (20s), so the
  dashboard correctly shows Claude Code as online with its current model.
- **`agent_type` in public clients API**: `/api/clients` response now includes
  `agent_type` field, allowing proxies to discover which vault client
  corresponds to a local claude-code agent.

---

## [0.1.18] — 2026-03-26

### Fixed
- **Fallback service stuck on Ollama**: When a primary service (e.g. Google) hit a
  transient error and fell back to Ollama, the proxy permanently overwrote its
  configured service/model and pushed the change to vault — making recovery impossible
  even after the primary key's cooldown expired. Now the user's preferred
  service/model is immutable during fallback; the proxy retries the preferred service
  first on every request and automatically recovers when keys become available again.
- **Dashboard online/offline detection**: Agent cards stayed "live" forever when a
  proxy died because the dashboard only received status updates inside heartbeat
  handlers — no heartbeats meant no offline transition. Redesigned with a unified
  `agents_sync` SSE event model:
  - Server computes status once, broadcasts ONE event with status + service/model/version.
  - Replaced dual `proxy_update` + `clients_status` events that could race each other.
  - Added 15-second periodic status ticker so the dashboard detects proxy death
    independently of incoming heartbeats.
  - SSE reconnect sync: new connections receive full state immediately via
    `Broker.OnConnect` callback.
  - Client-side watchdog monitors SSE health; `visibilitychange` handler catches
    stale state when returning to a background tab.

### Changed
- **Anthropic handler passthrough**: Non-Claude models (e.g. `google/gemini-*`)
  routed through Anthropic endpoint now correctly dispatch to the appropriate
  backend instead of silently falling back to `claude-haiku`.
- **Anthropic → OpenAI conversion**: Added `anthropicToOpenAIReq()` converter that
  preserves tool_use / tool_result content blocks when dispatching via OpenRouter.
- **Claude Code model sync**: `updateClaudeCodeModel()` now skips non-Claude models
  to avoid "There's an issue with the selected model" errors; `isClaudeModel()`
  helper added. OpenAI models endpoint includes Claude aliases when the configured
  model is non-Claude.
- **Claude Code apply**: Now passes `model` parameter through to `settings.json`.
- **Client identification**: `pushConfigToVault` includes `client_id` query param
  for unambiguous proxy identification. `handleClientConfig` prioritizes explicit
  `client_id` over token-based lookup.
- **Heartbeat active-client tracking**: `clientActs` TTL tightened from 5min/7min
  to 90s/3min; `applied` flag preserves longer grace for freshly-applied agents;
  `refreshClientAct` keeps long-running streaming requests visible.
- **Status thresholds**: Unified to 90s (live→delay) / 3min (delay→offline) across
  server render, SSE broadcast, and uptime reset (was 3min/10min).
- **Version bump**: `BASE_VERSION` → `v0.1.18`.

---

## [0.1.17] — 2026-03-25

### Added
- **Drag-and-drop agent card reordering**: Agent cards on the vault dashboard can
  now be reordered by dragging. Order is persisted in `Client.SortOrder` and saved
  to `vault.json` via `PUT /admin/clients/reorder`. Existing clients auto-migrate
  with sequential SortOrder on first load.
- **Inline apply buttons for disconnected agents**: When an agent (claude-code,
  cline, openclaw, nanoclaw) is not connected, the status area now shows a
  clickable [⚡ 설정 적용] button instead of a cryptic env-var instruction.
  Clicking it auto-writes the agent's local config via the proxy's `/agent/apply`.
- **`cokacdir` agent type**: New agent type for cokacdir (AI terminal file manager).
  Badge: 📂, color: #2d8659. Config copy provides `ANTHROPIC_BASE_URL` and
  `OPENAI_BASE_URL` environment variable templates.

### Removed
- **`vscode` agent type**: VS Code Continue extension — config format has diverged
  and the generated YAML no longer matches current Continue versions.
- **`antigravity` agent type**: Low adoption. Gemini CLI covers the same use case.

### Changed
- **Claude Code apply success message**: Now says "(새 대화 시작 시 적용)" to
  clarify that Claude Code picks up the new settings on the next conversation.
- **Version bump**: `BASE_VERSION` → `v0.1.17`.

---

## [0.1.16] — 2026-03-25

### Added
- **Bidirectional model sync for Cline**: When bot-b_vsc's model is changed in the
  vault dashboard, the proxy automatically updates Cline's `globalState.json` with the
  correct fields (`actModeOpenAiModelId`, `planModeOpenAiModelId`, `openAiModelId`).
  Preserves `openAiBaseUrl` — only model fields are touched.
- **Bidirectional model sync for Claude Code**: `updateClaudeCodeModel()` writes the
  `model` field to `~/.claude/settings.json`. Searches both WSL and Windows paths
  (`/mnt/c/Users/*/.claude/settings.json`).
- **SSE `onAnyConfig` callback**: `SSEClient` now receives config changes for *any*
  client (not just its own). Vault broadcasts `agent_type` in `config_change` events
  so proxies can dispatch to the correct local agent updater (`cline`, `claude-code`).
- **`agent_type` in `ConfigChangeEvent`**: Vault reads the client's `AgentType` after
  update and includes it in the SSE broadcast.

### Fixed
- **Cline field name mismatch**: Previously wrote `actModeApiModelId` /
  `planModeApiModelId` which Cline ignores. Now writes to the correct
  `actModeOpenAiModelId` / `planModeOpenAiModelId` fields.
- **Proxy's own model overwrote Cline config**: `onConfig` and `syncFromVault`
  called `updateClineModel()` with the proxy's own model (e.g. `qwen3.5:35b`),
  polluting Cline's settings. Removed — Cline is now only updated via `onAnyConfig`
  when a `cline`-typed client changes.
- **Unrelated client changes polluted Cline**: `onAnyConfig` previously fired for
  *all* foreign clients. When bot-c's model changed to `gemini-3.1-pro-preview`,
  it overwrote Cline's config. Now filtered by `agent_type`.

### Changed
- **Faster foreign client disconnect detection**: `clientActs` TTL reduced from
  5 minutes to 90 seconds. Combined with tighter dashboard thresholds for
  SSE=false clients (2min live / 5min delay vs 3min/10min for native proxies),
  VS Code closure is reflected in ~2.5 minutes instead of ~8 minutes.
- **Version bump**: `BASE_VERSION` → `v0.1.16`.

---

## [0.1.15] — 2026-03-22 (patch 11)

### Changed
- **Fallback order follows vault UI service list**: `dispatch()` now builds `tryOrder`
  directly from `allowedServices` (the proxy-enabled service list from the vault dashboard,
  in the order they appear). The primary configured service is moved to the front; all
  remaining proxy-enabled services follow in dashboard order (including vLLM, LM Studio,
  etc.). Falls back to `s.cfg.Proxy.Services` only when the vault list is not yet available.

### Fixed
- **Ollama HTTP 400 on tool-call conversations**: `callOllama` used `/api/chat` which requires
  `tool_calls.function.arguments` as a JSON object, but our internal format (OpenAI-compatible)
  stores arguments as a JSON string. Switched to Ollama's `/v1/chat/completions` OpenAI-compat
  endpoint which accepts the standard OpenAI format including arguments-as-string. Response is
  now parsed as `OpenAIResponse` via `OpenAIRespToGemini` instead of the native Ollama format.
- **Fallback chain incomplete when proxy starts without `-services` flag**: `dispatch()` built
  `tryOrder` from `s.cfg.Proxy.Services` (local config) and filtered by `allowedServices`
  (from vault). When no local services are configured (common in distributed mode), only the
  primary service + Anthropic ended up in `tryOrder` — Ollama and OpenRouter were silently
  skipped. Fixed by appending any vault-allowed services not already in `tryOrder` after the
  filter step, so all vault-configured services are tried in order.
- **Anthropic HTTP 400 incorrectly marked as key failure**: `callAnthropic` called
  `RecordError(key, 400)` on Bad Request responses, potentially triggering key cooldowns
  for request-format errors (wrong model name, unsupported parameters). 400 is a request
  error, not a key error; fixed to skip without cooldown so dispatch falls through normally.
- **Multi-turn tool calling broken for OpenRouter and Ollama backends**: `GeminiToOpenAI()`
  was reconstructing messages from Gemini `Contents` using `extractText()`, which turned
  `FunctionCall` and `FunctionResponse` parts into empty strings. When a cloud fallback
  (OpenRouter/Ollama) handled the second turn (with tool results), the tool call history
  was entirely lost and the model generated confused "I tried but failed" responses.
  Fixed by using `req.RawOAI.Messages` directly when available — the original OAI messages
  faithfully preserve `tool_calls`, `tool_call_id`, and `role=tool` fields without any
  round-trip conversion loss. Falls back to Gemini-content reconstruction only when
  `RawOAI` is not set (e.g. `handleGemini`/`handleAnthropic` paths).
- **Fallback model reflected in vault UI and openclaw TUI**: When `dispatch()` succeeded on a
  fallback service (e.g. Google keys exhausted → OpenRouter), `s.service`/`s.model` was never
  updated, so the vault UI and openclaw TUI continued showing the original (now-failing) model.
  Added `onFallback()` which updates `s.service`/`s.model`, calls `pushConfigToVault()` to
  persist the change, calls `updateOpenClawJSON()` to update the TUI, and fires the
  `EventModelChanged` hook — same path as a manual model change.
  Added `resolveActualModel()` to correctly report the Ollama model name (from env vars)
  instead of the upstream model string when Ollama is the fallback.

---

## [0.1.15] — 2026-03-22 (patch 5)

### Fixed
- **Tool calling end-to-end through proxy**: Multiple bugs prevented tool calls from working
  when OpenClaw/Claude Code routed through the proxy in OpenAI-compat mode (`/v1/chat/completions`):
  - `OpenAIToGemini` dropped the `tools` array entirely — Gemini and OpenRouter backends
    never received tool definitions. Fixed by converting OAI `tools` to Gemini
    `functionDeclarations` format and carrying the original `tools`/`tool_choice` through
    `RawOAI` for OAI-native backends.
  - `OpenAIMessage.UnmarshalJSON` only parsed `role` and `content`, silently dropping
    `tool_calls`, `tool_call_id`, and `name`. Fixed by adding these fields to the raw struct.
  - `function_response.name` was always empty for tool result messages: OAI tool results
    use `tool_call_id` (not `name`). Added `toolCallNames` map in `OpenAIToGemini` to track
    `tool_call_id → function_name` from preceding assistant `tool_calls`, then look up the
    name when converting role=tool messages to Gemini `functionResponse`.
  - Empty text `GeminiPart` created for assistant messages with `tool_calls` but no content
    (`msg.Content == ""`). Gemini rejects parts with no data field. Fixed by skipping empty
    content parts.
  - `GeminiCandidate.RawToolCalls` and `GeminiRequest.RawOAI` carry-through fields added to
    `models.go` so tool_calls survive the Gemini response → OAI response conversion.
  - `stripGeminiUnsupported` added to remove JSON Schema fields Gemini rejects
    (`additionalProperties`, `patternProperties`, `$schema`, `$ref`, `$defs`, `definitions`,
    `unevaluatedProperties`, `strict`) from function declaration parameters.
- **`parseProviderModel` ignoring configured OpenRouter service**: when `svc=openrouter` and
  the model had a provider prefix (e.g. `google/gemini-2.5-flash`), the switch matched
  `"google"` and re-routed to the native Google handler, bypassing OpenRouter. Fixed by
  returning `("openrouter", mdl)` early when `svc == "openrouter"`.
- **`callGoogle` swallowed HTTP error body**: non-200 responses discarded the body before
  logging, making 400 errors undiagnosable. Fixed by reading body before closing.
- **`openclaw_sync` Anthropic provider routing**: when `service == "anthropic"`, now
  configures `models.providers.anthropic` with `baseUrl = "http://localhost:56244"` so
  OpenClaw sends tool-aware `/v1/messages` requests through the proxy passthrough path
  instead of calling the real Anthropic API directly.

---

## [0.1.15] — 2026-03-22 (patch 4)

### Fixed
- **`ResetDailyUsage` not resetting `UsageDate`**: daily usage reset cleared `TodayUsage` and
  `TodayAttempts` but left `UsageDate` stale, causing the auto-reset guard in `SetKeyUsage` /
  `SetKeyAttempts` to skip the reset on the next heartbeat.
- **`handleAdminKeys` missing `today_attempts` field**: the safe struct for `GET /admin/keys`
  was omitting `TodayAttempts`, so the REST response always returned `0` for attempts.
- **Sub-minute cooldown shown as "0분 후"**: `%.0f` of `remain.Minutes()` rounds sub-60s
  values to 0. Now shows seconds (e.g. "45초 후") when cooldown is under 60 seconds. Added
  `key_in_sec` i18n key to all 17 locale files.
- **`_keyCache` empty on page load**: countdown ticker had no data for the first ~20s after
  page load. Added `_seedKeyCache()` IIFE that populates `_keyCache` from server-rendered
  `data-cd-ms` DOM attributes on load so countdowns start ticking immediately.
- **Model change in vault UI not immediately reflected**: `lookupTokenConfig` cached
  token→model mappings for 30 seconds; changing a client's model didn't invalidate the cache.
  Now any `config_change` SSE event immediately flushes the entire token cache via the new
  `onConfigFlush` callback on `SSEClient`, so the new model takes effect within one request.
- **heartbeat `activeKeys` missing `openai`, `lmstudio`, `vllm`**: the service list for
  last-used key tracking only included `google`, `openrouter`, `anthropic`, `ollama`.

### Changed
- **`BatchUpdateKeyMetrics` replaces 3-loop heartbeat key sync**: vault's `handleHeartbeat`
  previously called `SetKeyUsage`, `SetKeyAttempts`, and `SetKeyCooldownIfLater` in separate
  loops (up to 3N `save()` calls per heartbeat). Replaced with a single
  `BatchUpdateKeyMetrics` that updates all keys in one lock and one `save()`.
- **O(n²) → O(n) service ordering in `buildKeysCard`**: replaced inner linear scan with a
  `map[string]bool` set.

---

## [0.1.15] — 2026-03-21 (patch 3)

### Added
- **One-click config apply for all local agents** (`/agent/apply` proxy endpoint): replaced
  the Cline-only `/cline/apply` endpoint with a unified `/agent/apply` that dispatches to the
  correct config writer based on `agentType`. Supported types:
  - `cline` → `~/.cline/data/globalState.json` + `secrets.json` (WSL-aware path detection)
  - `claude-code` → `~/.claude/settings.json`
  - `openclaw` / `nanoclaw` → `~/.openclaw/openclaw.json` (updates `models.providers.custom`
    with `baseUrl` / `apiKey` and sets `agents.defaults.model.primary`)
- **`⚡ 설정 적용` buttons for openclaw, nanoclaw, and claude-code** agent types in the
  dashboard: clicking the button calls `applyAgentConfig(clientId, agentType)` on the local
  proxy and patches the config file in one step. Previously these types only had a copy button.
- **Generic `applyAgentConfig(clientId, agentType)` JS function**: replaced the Cline-specific
  `applyClineConfig` with a single function that handles all agent types and shows a
  type-appropriate success message.

---

## [0.1.15] — 2026-03-21 (patch 2)

### Added
- **Token-based model override for third-party clients (Cline, Cursor, etc.)**: proxy now
  resolves the Bearer token on each `/v1/chat/completions` request via the new vault
  `GET /api/token/config` endpoint and overrides the requested model with the agent's
  dashboard-configured `default_service`/`default_model`. This lets any OpenAI-compatible
  client be controlled from the wall-vault dashboard without changing its local settings.
  Results are cached for 30 seconds to avoid per-request vault round-trips.
- **`cline` agent type**: new agent type in the dashboard with 🔧 icon, status hint, and
  "Cline 설정 복사" button that outputs provider/base URL/API key for Cline's UI settings.

### Fixed
- **Double close bug in `lookupTokenConfig`**: `resp.Body.Close()` was called both explicitly
  in the error branch and via `defer`, causing a panic on non-200 vault responses. Moved the
  defer after the nil check so only one close path is taken.
- **Unbounded token cache**: added eviction of expired entries when the cache exceeds 500
  entries to prevent memory growth from unique tokens.
- **Bearer token extraction duplication**: extracted `bearerToken(r)` helper in vault/server.go
  to replace four copies of `strings.TrimPrefix(r.Header.Get("Authorization"), "Bearer ")`.
- **Missing Cline JS handler**: `copyAgentConfig` had the button but no `cline` branch, so
  clicking the button silently did nothing. Added the config generation logic.

---

## [0.1.15] — 2026-03-21 (patch)

### Fixed
- **Clipboard copy buttons on HTTP + IP access**: `navigator.clipboard.writeText` is only
  available in secure contexts (HTTPS or localhost). Accessing the dashboard via
  `http://192.168.x.x:56243` caused `TypeError: Cannot read properties of undefined
  (reading 'writeText')` before any Promise was created, so the existing `.catch()` fallback
  never fired. Added `_copyText()` helper that checks `navigator.clipboard` availability
  first; if absent, it falls back to `prompt()` so the user can manually copy with Ctrl+A →
  Ctrl+C. Applied to all three copy buttons: OpenClaw config, agent config, and deploy script.

---

## [0.1.15] — 2026-03-20

### Fixed
- **Local service auto-enable (Ollama / LM Studio / vLLM)**: `_checkLocalSvc` now uses
  `GET /admin/services/{id}/ping` instead of `GET /admin/models?service={id}` to decide
  whether to enable a local service. The previous approach never enabled the service
  because `handleAdminModels` only queries models for already-enabled services
  (chicken-and-egg). The ping endpoint connects directly to the configured `local_url`
  (or the default port) regardless of enabled state, so services are correctly
  auto-checked when the local server is reachable.
- **Service checkbox override loop**: manual checkbox toggles were being immediately
  reverted by `autoCheckServices`. Two root causes fixed:
  1. `service_changed` SSE handler no longer calls `autoCheckServices` — it only
     refreshes the agent service/model selects. Calling `autoCheckServices` on every
     toggle created a ping→enable→SSE→autoCheck→ping loop that undid user intent.
  2. `checkLocalService` (15 s dot-ping loop) no longer calls `_setSvcEnabled` — it
     updates only the status dot (●). Checkbox state is now controlled exclusively by
     `autoCheckServices` on page load and `key_added`/`key_deleted` SSE events.

---

## [0.1.14] — 2026-03-20

### Added
- **Local service status indicator**: green/grey dot (●) next to local service names
  (Ollama, LM Studio, vLLM) in the Services card — auto-pings on page load and after
  saving the URL. Endpoint: `GET /admin/services/{id}/ping` (3 s timeout).

### Fixed
- **Ollama distributed routing**: proxy now receives `local_url` per service from vault
  (`/api/services` returns `[{id, local_url}]` instead of `[]string`). Ollama URL
  priority: env var → vault config → `localhost:11434`.
- **Google model fallback to OpenRouter**: `google/X` model names no longer lose the
  `google/` prefix when falling back to OpenRouter. `callGoogle` strips the prefix
  internally; all other services receive the full `google/X` form.
- **UpsertService partial update**: PUT `/admin/services/{id}` now uses map-based
  partial update so toggling `proxy_enabled` does not accidentally reset `enabled` to
  `false` (Go JSON zero-value bug).
- **AnthropicRequest.System array**: Claude Code ≥ 2026-03 sends `system` as
  `[{type, text}]` array instead of a plain string. `System` field changed to
  `json.RawMessage`; new `SystemText()` method handles both formats.
- **Anthropic native passthrough**: `/v1/messages` handler now forwards the original
  request body directly to Anthropic (skipping GeminiRequest round-trip) to preserve
  tool calls, tool_results, and multi-block content.

---

## [0.1.8] — 2026-03-20

### Added
- `nanoclaw` agent type in dashboard agent modal — lightweight OpenClaw-compatible agent
  - 🦀 teal badge (`.atb-nanoclaw`, `#16a085`)
  - Work directory auto-fills as `~/nanoclaw` on type selection
  - **🦀 NanoClaw 설정 복사** config copy button — reuses OpenClaw `~/.openclaw/openclaw.json` format
  - `cfg_nanoclaw` / `cfg_nanoclaw_title` i18n keys added to all 17 locale files
- `install.sh`: one-liner installer — auto-detects OS/arch, downloads correct binary, installs to `~/.local/bin`

---

## [0.1.8-pre] — 2026-03-17

### Fixed

#### Midnight daily reset — stale counter feedback loop
Previously, after the vault broadcasted `usage_reset` at 00:00:30, a running proxy would
push its stale yesterday-values back to the vault via the next heartbeat. The vault stored
them with today's `usage_date`, the next proxy startup loaded them as today's usage, and
the loop repeated indefinitely.

Four-part fix:

1. **`proxy/keymgr.go` — `ResetDailyCounters()`**: new method that zeroes all
   `todayUsage` / `todayAttempts` in the local key map immediately. Called on `usage_reset`
   SSE before the follow-up `SyncFromVault()`.

2. **`proxy/keymgr.go` — `lastSyncDate` rollover detection**: `KeyManager` now tracks
   `lastSyncDate string` ("YYYY-MM-DD"). In `SyncFromVault()`, if the current date differs
   from `lastSyncDate`, locally accumulated counters are discarded entirely so that
   `max(vault=0, local=yesterday)` cannot keep yesterday's values alive.

3. **`proxy/sseconn.go` — `onUsageReset` callback**: separated `usage_reset` event from
   `key_added`/`key_deleted` in the SSE dispatcher. Proxy registers a dedicated callback
   that calls `ResetDailyCounters()` then `SyncFromVault()`.

4. **`vault/store.go` — `UsageDate` auto-reset**: `SetKeyUsage` and `SetKeyAttempts` now
   check `k.UsageDate` against today. When a new day is detected, they reset
   `TodayUsage`/`TodayAttempts` to zero before applying the new value, preventing
   cross-day stale accumulation even without the proxy-side fix.

#### `UsageDate` field propagation
- `vault/models.go` `APIKey`: added `UsageDate string` (`"usage_date"` JSON) — records the
  YYYY-MM-DD of the last `today_usage` write.
- `vault/server.go` `safeKey` struct and `/api/keys` response: `UsageDate` now included so
  proxy can compare against `time.Now().Format("2006-01-02")`.
- `proxy/keymgr.go` `SyncFromVault()`: if `k.UsageDate != today`, vault's reported
  `TodayUsage`/`TodayAttempts` are treated as zero (stale — proxy's local counters win).

#### Streaming token count always reported as 1
OpenRouter's final SSE chunk carries `usage` with total token count but has an empty
`choices` array and/or empty `delta.content`. `streamOpenRouter` was checking `choices`
length and `delta.content` with `continue` **before** reading `chunk.Usage`, so the
token count was always discarded and the fallback value of `1` was used.

Fix in `proxy/stream.go`: moved `chunk.Usage` extraction before all `continue` guards.
Both `streamOpenRouter` and `streamGoogle` now apply a `min=1` fallback only when the
API genuinely returns no usage metadata.

### Changed

#### Key selection: round-robin → drain-first
`proxy/keymgr.go` `KeyManager.Get()` previously advanced `idx` by `+1` after every
successful key acquisition (round-robin), distributing load evenly across all keys.

Changed to drain-first: `idx` now stays on the current key after a successful call
(`(start+i)%n` instead of `(start+i+1)%n`). The next key is only selected when
`isAvailable()` returns false (cooldown or daily limit reached). This ensures a single
key absorbs all traffic until it is exhausted before moving to the next.

#### Dashboard UI
- Dashboard title updated to **"벽금고(wall-vault) 대시보드"** across all 17 locale files
  (`title` key). Previously service-specific titles per locale.
- HTML `<title>` tag updated to match.
- Logo moved from `.header` (non-sticky, scrolled under topbar) to `.topbar-brand`
  (sticky, always visible). Height fixed at `38px`; `border-radius` and `object-fit`
  removed.
- `.header` section simplified to h1 title only (`font-size:1.5rem`, `font-weight:700`).

---

## [0.1.7] — 2026-03-16

### i18n
- Added 20 new i18n keys to all 17 locale files — previously hardcoded Korean strings in dashboard JS/HTML are now fully translated: `proxy_use`, `lbl_avatar`, `st_claude_hint`, `st_editor_hint`, `st_gemini_hint`, `toggle_model`, `cfg_gemini_cli`, `cfg_gemini_cli_title`, `cfg_antigravity`, `cfg_antigravity_title`, `err_name`, `ph_token_edit`, `cfg_ok`, `cfg_manual`, `cfg_openclaw_hint`, `cfg_claude_hint`, `cfg_cursor_hint`, `cfg_vscode_hint`, `cfg_gemini_cli_hint`, `cfg_antigravity_hint`
- Added `key_att` i18n key to all 17 locale files — used in key usage panel to display attempt count label (e.g. `"시도"` / `"att"` / `"Vers."` / `"試行"` etc.)
- Fixed `ko.json` time unit values: `uph` `"h"` → `"시간"`, `upm` `"m"` → `"분"`, `ups` `"s"` → `"초"` — countdown timer and key usage panel were showing raw English letters in Korean UI
- `internal/i18n/i18n_test.go` `TestSupported`: updated expected language count from 10 → 17

### Added
- Agent modal default model selection UI: full service-specific model list via dropdown, with manual text input fallback (`onAgentServiceChange`, `onModelSelect`)
- Agent status 4 states: 🟢 Online (<3min) / 🟡 Delayed (3–10min) / 🔴 Offline (>10min) / ⚫ Inactive / Disconnected
- `.dot-yellow` CSS class (+ glow effect)
- `.dot-red` CSS glow effect
- `vscode` agent type option in agent modal
- Work directory auto-hint on agent type change (`onAgentTypeChange` JS function): `openclaw` → `~/.openclaw`, `claude-code` → `~/.claude`, `cursor`/`vscode` → `~/projects`
- `docs/logo.png` logo image
- README: origin story + full rewrite (MuJaMae style)

### Fixed (proxy routing)
- `parseProviderModel`: added `custom/` prefix handler — `custom/google/gemini-...` and similar paths sent by the OpenClaw model picker were falling through to the `default` case and routing to OpenRouter instead of the correct provider. Now strips `custom/` and re-parses the remainder recursively.
- Ollama non-streaming call used `http.Client{Timeout: 60s}` which expired before large-model inference completed, producing a misleading "Ollama connection failed: context deadline exceeded" error even when Ollama was healthy. Changed to `Timeout: 0` (no deadline) since Ollama is a local service and generation time is unbounded.

### Changed
- Key usage section fully redesigned — `handleHeartbeat` in `server.go` now always broadcasts `usage_update` SSE with full key state snapshot `{keys: [{id, service, today_usage, today_attempts, daily_limit, cooldown_until}]}` after every heartbeat (previously conditional on non-empty usage map). Dashboard JS `refreshKeyUsage` no longer fetches `/admin/keys`; it updates DOM directly from SSE payload. `_keyCache` changed from array to object (id → state) for O(1) lookup. Keys with `daily_limit=0` now use **share-of-total** bar scaling (key activity / sum of all keys in that service group × 100%).
- `Makefile`: `VERSION` assignment changed from recursive (`=`) to immediate (`:=`) — `$(shell date)` is now evaluated once at `make` start, preventing version mismatch between build and verify steps
- `Makefile.local` + `Makefile.local.example`: deploy targets hardened with kill→wait→verify pattern:
  - `pkill -x "wall-vault"` after service stop (exact process name, not `-f`)
  - 10-second wait loop using `pgrep -x "wall-vault"`
  - Error exit if process still alive after 10 seconds
  - Binary replacement only proceeds after confirming old process is dead

### Fixed
- Key usage display now separates successful token count (`today_usage`) from total attempt count (`today_attempts`). Bar for unlimited keys uses **share-of-total** scaling (each key's activity / sum of all keys in service × 100%), not max-relative, to avoid the "all-100% or all-0%" binary appearance. Server-side initial render (`buildKeysCard`) uses the same formula with `TodayAttempts`. Rate-limited requests (429, 402, 582) increment `today_attempts` only; `today_usage` tracks successful tokens/requests only. Dashboard label shows `"N req (M att)"` or `"M att"` (all failed) when attempts differ from usage.
- HTTP 582 (upstream gateway overload) added to cooldown table with 5-minute backoff; previously fell through to the 10-minute default
- `today_attempts` field added to `APIKey` (vault.json), heartbeat payload (`key_attempts`), SSE `usage_update`, and `/api/keys` response so vault, proxy, and dashboard all stay in sync
- Countdown timer in key status panel was hardcoded Korean (`분`, `초`) — now uses `T('upm')` / `T('ups')`
- Request count label `'요청'` in key usage was hardcoded — now uses `T('key_reqs')`
- `copyOpenClawConfig` / `copyAgentConfig` alert/prompt messages were hardcoded Korean — now fully i18n via `T()` keys
- `pgrep -f "wall-vault"` in deploy scripts self-matched the shell process running `make` — replaced with `pgrep -x "wall-vault"` throughout `Makefile.local` and `Makefile.local.example`
- Version mismatch during deploy verify: `VERSION =` re-evaluated `$(shell date)` at verify time (seconds after build), producing a different timestamp — fixed with `VERSION :=`
- Agent modal field order: ID → Name → Agent type → Work dir → Service → Model → Description → IP whitelist → Token → Enabled
- `buildClientModalBody` `fmt.Sprintf` argument count mismatch (19 verbs / 20 args → 20/20)
- Offline state (`dot-red`) CSS class was not actually applied

### Security
- Applied `adminAuth` middleware to `/admin/theme` and `/admin/lang` endpoints (was unauthenticated)
- `/api/keys` handler now enforces IP whitelist — CIDR notation supported (`net.ParseCIDR`), `X-Forwarded-For` header handled
- Admin auth failure rate limiting: `429 Too Many Requests` after 10 failures within 15 minutes
- Added `realIP()`, `ipAllowed()` helper functions

---

## [0.1.6] — 2026-03-14

### Added
- `resolveAvatarDataURI(avatarVal string)`: avatar field now accepts relative paths under `~/.openclaw/` (e.g. `workspace/avatars/profile.hpg`), in addition to base64 data URIs — per-agent avatar file support
- Supported avatar extensions: `.png`, `.jpg`/`.jpeg`/`.hpg`, `.webp`, `.gif`; MIME type auto-detected from extension
- Avatar auto-sync via heartbeat: proxy reads local avatar file (`WV_AVATAR` env var, relative to `~/.openclaw/`) and sends as base64 data URI in heartbeat payload. Vault auto-updates client avatar record on receive. `readLocalAvatar()` in `proxy/heartbeat.go`.
- `WV_AVATAR` environment variable: relative path under `~/.openclaw/` for the proxy's local avatar file. Set per machine in systemd unit or launchd plist.

### Changed
- Agent model dropdowns now show **only proxy-enabled services** — both Go `buildServiceOptions()` and JS `refreshServiceSelects()` filter by `proxy_enabled: true`. Agents can only select from services that have the "프록시 사용" checkbox enabled in the Services card.
- Workspace avatar (default `~/.openclaw/workspace/avatar.png`) is now shown for agents with `agent_type == ""` (unset) as well as `openclaw` — fixes avatar not displaying for existing agents without type set
- Version format now includes build timestamp: `v0.1.6.YYYYMMDD.HHmmss` (e.g. `v0.1.6.20260314.231308`) — generated automatically by Makefile at build time. `var version` in `main.go` defaults to `"dev"` for non-Makefile builds.

### Fixed
- `SyncFromVault()` in `proxy/keymgr.go` was resetting locally accumulated `today_usage` to vault value (often 0 in standalone mode) every 5 minutes — fixed by preserving the higher of local vs. vault usage across syncs
- Agent avatar not showing when `agent_type` field is blank (vault clients created before v0.1.3 had no type set)
- `buildClientModalBody`: `border-radius:50%` inside `fmt.Sprintf` format string consumed an extra argument, rendering `submitModal('%!s(MISSING)')` — agent add/edit save buttons were completely broken. Fixed by escaping to `50%%`.

---

## [0.1.5] — 2026-03-14

### Added
- Proxy service selection: "프록시 사용" checkbox per service card → only checked services are used by OpenClaw proxy
- `/api/services` endpoint (client-auth): returns list of proxy-enabled service IDs
- `service_changed` SSE now includes `proxy_services []string` — proxy updates `allowedServices` in real-time
- `proxy/sseconn.go`: `onServiceChange` callback for live proxy service filtering
- `proxy/server.go`: `allowedServices` field + `syncAllowedServices()` on startup

### Changed
- Service card UI: removed redundant ID label below service name (name already identifies the service)
- Documentation reframed around OpenClaw as primary use case (README, MANUAL, API)

### Fixed
- Binary on mini was v0.1.3; redeployed v0.1.4+ darwin/arm64 build

---

## [0.1.3] — 2026-03-13

### Added
- Agent card redesign — per-type icons & config copy buttons:
  - `openclaw` → 🦞 (red lobster) + "OpenClaw 설정 복사" button
  - `claude-code` → 🟠 + "Claude Code 설정 복사" (copies `~/.claude/settings.json` snippet)
  - `cursor` → ⌨ + "Cursor 설정 복사" (copies Cursor AI API settings)
  - `vscode` → 💻 + "VSCode 설정 복사" (copies Continue extension `config.json` snippet)
  - generic/custom → 📋 "설정 복사" (OpenClaw format)
- `copyAgentConfig(clientId, agentType)`: per-type proxy config generator (JS)
- Connection status chip with context hint ("● 프록시 미연결" + heartbeat explanation)
- `💾 저장` button replaces bare "적용" — intent is now explicit
- After save: `✓ 저장됨` inline indicator in status area (3s) + `✓` on button (2s)
- New CSS: `.atbadge`, `.atb-openclaw/claude/cursor/vscode/custom`, `.agent-status`,
  `.status-live/delay/offline/dc/hint/version`, `.btn-cfg`, `.btn-cfg-openclaw/claude`, `.btn-save`

### Changed
- `buildAgentsCard()`: fully rewritten with per-item `strings.Builder` (no more single large `Sprintf`)
- Agent type badge: colored pill per type (red=openclaw, orange=claude-code, blue=cursor/vscode)
- Status display: `미연결` → `● 프록시 미연결 — heartbeat 미수신` with guidance

### Fixed
- "미연결" ambiguity: users can now distinguish proxy connection state from config-save result

---

## [0.1.2] — 2026-03-13

### Added
- `callOpenAI()`: direct OpenAI API handler (separate from OpenRouter)
- `dispatch()`: `openai` case (direct), `anthropic` case (via OpenRouter with `anthropic/model` path)
- `parseProviderModel()` comprehensive rewrite (OpenClaw 3.11 compatibility):
  - `anthropic/` → OpenRouter `anthropic/model` (Anthropic API format differs)
  - `openai/` → direct OpenAI
  - `:cloud` suffix (Ollama cloud tags) → strip + route to OpenRouter
  - New prefixes: `opencode`, `opencode-go`, `opencode-zen`, `moonshot`, `kimi-coding`,
    `groq`, `mistral`, `cohere`, `perplexity`, `minimax`, `together`, `huggingface`,
    `nvidia`, `venice`, `meta-llama`, `qwen`, `deepseek`, `01-ai`
  - `wall-vault/claude-*` → OpenRouter `anthropic/model` (was incorrectly routing to `anthropic` service)
- `stripControlTokens()`: removes GLM-5 / DeepSeek / ChatML control tokens from responses (`<|im_start|>`, `[gMASK]`, `[sop]`, etc.)
- `fetchOpenRouterKnown()`: curated fallback model list — Hunter Alpha (1M ctx, free), Healer Alpha, Kimi K2.5, GLM-5, GLM-4.7 Flash, DeepSeek R1/V3, Qwen 2.5, MiniMax M2.5, Llama 3.3
- `OllamaRecommended()`: fallback model list when Ollama server is unreachable (glm-4.7-flash, qwen3.5:35b, deepseek-r1:7b, etc.)
- Google model list: `gemini-2.5-flash-8b`, `gemini-embedding-2-preview` (OpenClaw 3.11 memorySearch)
- OpenAI model list: `o3`

### Changed
- OpenRouter fetch failure → fall back to `fetchOpenRouterKnown()`
- Ollama server unreachable → fall back to `OllamaRecommended()`
- Response text in `/v1/chat/completions` now passes through `stripControlTokens()`

### Fixed
- `anthropic` / `openai` services were silently ignored in `dispatch()`
- `wall-vault/claude-*` models were never actually called

---

## [0.1.1] — 2026-03-13

### Added
- Agent card: model dropdown + manual input combo (same as modal) with auto-load on page
- `onAgentServiceChange()`, `onModelSelect()` JS functions for agent service/model combo
- DOMContentLoaded initializer pre-loads model lists for all agent cards on page load
- README: OpenClaw integration section (KO + EN) — socket events, SSE sync, dir layout
- README: multilingual sections (zh, ja, es, fr, de)
- README: copyright/license notice (GPL-3.0)

### Changed
- License: MIT → GPL-3.0
- Theme order unified to light/dark/gold/cherry/ocean across all code and docs
- Agent modal: model field upgraded from datalist to select+input combo
- All commit messages in English going forward

### Fixed
- `setTheme()` / `setLang()` missing `Authorization` header → 401 on theme/lang change
- `server.go` theme error message updated to reflect correct order

---

## [0.1.0] — 2026-03-11

### Post-release additions
- `cmd/proxy`: `--key-google`, `--key-openrouter`, `--vault`, `--vault-token`, `--filter` flags
- `internal/models`: `Registry.NeedsRefresh()`, `Registry.Search(query)`
- `internal/proxy/server_test.go`: 12 proxy HTTP handler tests
- `internal/vault/server_test.go`: 15 vault HTTP handler tests
- `internal/middleware/middleware_test.go`: 8 middleware chain tests
- `internal/hooks/hooks_test.go`: 7 hook system tests
- `docs/API.md`: full API endpoint reference
- `docs/MANUAL.md`: user guide (install → distributed mode → troubleshooting)
- `CONTRIBUTING.md`: contributor guide
- GitHub Actions CI/Release workflows (ready locally)

### Initial release (single Go binary)

#### Architecture
- **Single binary** `wall-vault` — subcommand pattern (start / proxy / vault / doctor / setup)
- **standalone / distributed** two operating modes
- **SSE (Server-Sent Events)** real-time config sync (vault → proxy, within 1–3s)
- **AES-GCM encryption** — master-password-based API key persistence

#### Subcommands

| Command | Description |
|---------|-------------|
| `wall-vault start` | Run proxy + vault together (standalone) |
| `wall-vault proxy` | Run proxy only |
| `wall-vault vault` | Run vault only |
| `wall-vault doctor` | Health check and auto-recovery |
| `wall-vault setup` | Interactive setup wizard |

#### Proxy features
- **Google Gemini / OpenRouter / Ollama** simultaneous support
- **Round-robin key management** — per-service index tracking with `idx map[string]int`
- **Cooldown management** — 429: 30min, 400/401/403: 24h, network error: 10min
- **Tool security filter** — strip_all / whitelist / passthrough
- **Fallback chain** — Google → OpenRouter → Ollama
- **Hook system** — shell commands on model change, key exhaustion, service down
- **OpenClaw socket** event integration

#### Vault
- **REST API** — `/api/keys`, `/api/clients`, `/api/status`
- **SSE broadcast** — `/api/events` endpoint
- **Web dashboard** — themes (sakura/dark/light/ocean), key CRUD, client management
- **Admin token** authentication

#### Doctor
- `doctor check` / `fix` / `status` / `all` / `deploy` subcommands
- Auto-recovery priority: **systemd → launchd → NSSM (Windows) → direct process**
- `deploy` — auto-generates systemd / launchd / NSSM service files

#### Setup wizard
- **Top 10 world languages** — ko/en/zh/es/hi/ar/pt/fr/de/ja
- Interactive configuration: theme, mode, ports, services, tool filter, security tokens
- Ollama server auto-connect and model list fetch
- Secure admin token auto-generation via `crypto/rand`

#### i18n
- Top 10 world languages supported
- Auto-detect from LANG / WV_LANG environment variables
- Locale string parsing (e.g. `ko_KR.UTF-8` → `ko`)
- English fallback guaranteed

#### Platform support
- **Linux** (amd64 / arm64)
- **macOS** (amd64 / arm64, Apple Silicon)
- **Windows** (amd64, NSSM service support)
- **WSL** fully supported

#### Model registry
- Google: 6 fixed models (Gemini 1.5/2.0/2.5)
- OpenRouter: 346+ dynamic fetch
- Ollama: local server auto-discovery
- TTL-based cache (default 10min)
- Case-insensitive model ID/name search

#### Service plugins
- External service plugin loader from `~/.wall-vault/services/*.yaml`
- `enabled: true/false` field for runtime activation control

#### Tests (39, all PASS)
- `crypto_test.go` — AES-GCM encrypt/decrypt/random nonce (5)
- `toolfilter_test.go` — strip_all/whitelist/passthrough (5)
- `convert_test.go` — Gemini↔OpenAI↔Ollama format conversion (6)
- `services_test.go` — plugin loader edge cases (5)
- `keymgr_test.go` — round-robin/cooldown/daily limit (8)
- `store_test.go` — key/client CRUD/persistence (10)

#### CI/CD
- GitHub Actions CI — vet + test + 4-platform cross-compile on push/PR
- GitHub Actions Release — auto GitHub Release on `v*` tag

---

[Unreleased]: https://github.com/sookmook/wall-vault/compare/v0.1.8...HEAD
[0.1.8]: https://github.com/sookmook/wall-vault/compare/v0.1.7...v0.1.8
[0.1.7]: https://github.com/sookmook/wall-vault/compare/v0.1.6...v0.1.7
[0.1.6]: https://github.com/sookmook/wall-vault/compare/v0.1.5...v0.1.6
[0.1.5]: https://github.com/sookmook/wall-vault/compare/v0.1.3...v0.1.5
[0.1.3]: https://github.com/sookmook/wall-vault/compare/v0.1.2...v0.1.3
[0.1.2]: https://github.com/sookmook/wall-vault/compare/v0.1.1...v0.1.2
[0.1.1]: https://github.com/sookmook/wall-vault/compare/v0.1.0...v0.1.1
[0.1.0]: https://github.com/sookmook/wall-vault/releases/tag/v0.1.0
