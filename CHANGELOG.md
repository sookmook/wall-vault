# Changelog

All notable changes to wall-vault are documented here.
Format follows [Keep a Changelog](https://keepachangelog.com/en/1.0.0/).

wall-vault의 모든 주요 변경 사항을 기록합니다.
형식은 [Keep a Changelog](https://keepachangelog.com/ko/1.0.0/)를 따릅니다.

---

## [Unreleased]

### i18n
- Added 20 new i18n keys to all 17 locale files — previously hardcoded Korean strings in dashboard JS/HTML are now fully translated: `proxy_use`, `lbl_avatar`, `st_claude_hint`, `st_editor_hint`, `st_gemini_hint`, `toggle_model`, `cfg_gemini_cli`, `cfg_gemini_cli_title`, `cfg_antigravity`, `cfg_antigravity_title`, `err_name`, `ph_token_edit`, `cfg_ok`, `cfg_manual`, `cfg_openclaw_hint`, `cfg_claude_hint`, `cfg_cursor_hint`, `cfg_vscode_hint`, `cfg_gemini_cli_hint`, `cfg_antigravity_hint`
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

### Changed
- `Makefile`: `VERSION` assignment changed from recursive (`=`) to immediate (`:=`) — `$(shell date)` is now evaluated once at `make` start, preventing version mismatch between build and verify steps
- `Makefile.local` + `Makefile.local.example`: deploy targets hardened with kill→wait→verify pattern:
  - `pkill -x "wall-vault"` after service stop (exact process name, not `-f`)
  - 10-second wait loop using `pgrep -x "wall-vault"`
  - Error exit if process still alive after 10 seconds
  - Binary replacement only proceeds after confirming old process is dead

### Fixed
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

[Unreleased]: https://github.com/sookmook/wall-vault/compare/v0.1.6...HEAD
[0.1.6]: https://github.com/sookmook/wall-vault/compare/v0.1.5...v0.1.6
[0.1.5]: https://github.com/sookmook/wall-vault/compare/v0.1.3...v0.1.5
[0.1.3]: https://github.com/sookmook/wall-vault/compare/v0.1.2...v0.1.3
[0.1.2]: https://github.com/sookmook/wall-vault/compare/v0.1.1...v0.1.2
[0.1.1]: https://github.com/sookmook/wall-vault/compare/v0.1.0...v0.1.1
[0.1.0]: https://github.com/sookmook/wall-vault/releases/tag/v0.1.0
