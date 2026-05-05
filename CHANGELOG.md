# Changelog

All notable changes to wall-vault are documented here.
Format follows [Keep a Changelog](https://keepachangelog.com/en/1.0.0/).

wall-vault의 모든 주요 변경 사항을 기록합니다.
형식은 [Keep a Changelog](https://keepachangelog.com/ko/1.0.0/)를 따릅니다.

---

## [0.2.64] — 2026-05-05

Documentation rewrite + repository sanitisation pass. No proxy or vault
behaviour changes; the binary built from this tag is functionally
equivalent to v0.2.63.

### Changed

- **README rewritten as English master** (`README.md`, ~390 lines), focused
  on what wall-vault does for an external operator: install, TLS setup,
  client connections, configuration, modes, plugin yamls, and build from
  source. The previous monolith mixed origin-story prose, fleet
  narrative, and seven languages in a single 1370-line file.
- **MANUAL rewritten as English master** (`docs/MANUAL.md`, 621 lines)
  covering install, setup wizard, TLS, key registration, agent
  connections, dashboard, distributed mode, auto-start, plugin yamls,
  doctor, hooks, environment variables, and troubleshooting. Removed
  outdated v0.2.x upgrade notes that CHANGELOG already covers.
- **Per-language files** for both README and MANUAL: `*.ko.md` Korean,
  `*.zh.md` Simplified Chinese, `*.ja.md` Japanese, `*.es.md` Spanish,
  `*.fr.md` French, `*.de.md` German, `*.pt.md` Portuguese, `*.ar.md`
  Arabic, `*.hi.md` Hindi, `*.id.md` Indonesian, `*.th.md` Thai, `*.sw.md`
  Swahili, `*.ha.md` Hausa, `*.ne.md` Nepali, `*.mn.md` Mongolian (modern
  Cyrillic), `*.zu.md` isiZulu. 17 locales total per document. The English
  master moves to `README.md` / `docs/MANUAL.md`; previous Korean masters
  become `README.ko.md` / `docs/MANUAL.ko.md` to mirror the README/MANUAL
  pattern. Old `docs/MANUAL.en.md` removed.
- **Source comments and test files anonymised.** Per-host references in
  `internal/proxy/{server,openclaw_sync,openclaw_sanitize,local_dispatch_test,openclaw_heal_test,token_sentinel_test}.go`
  were rewritten from concrete hostnames (`raspi`, `motoko`, `jaksooni`,
  `mini9`) to generic placeholders (`host-A`/`host-B`/`host-C`/`host-D`).
  Test fixture IPs that pinned a specific LAN address moved to neutral
  placeholders inside the RFC1918 range. CHANGELOG entries that
  described per-host incidents were rewritten the same way. The author's
  personal site/email links were already removed from the dashboard
  footer in v0.2.63; the corresponding CHANGELOG narrative was sanitised
  here to match.

### Notes

- A fresh `git clone` shows 17-locale READMEs and manuals; existing
  installs get the same when they pull this tag.
- No code-path change. `make build` produces a binary stamped
  `v0.2.64.…` but the proxy + vault dispatch is identical to v0.2.63.

---

## [0.2.63] — 2026-05-05

External-user usability hardening pass. Three diagnostic agents flagged
several places where wall-vault assumed a single-operator deployment;
this release fixes the ones that block a fresh non-default install.

### Added

- **`proxy.anthropic_fallback_model` config / `WV_ANTHROPIC_FALLBACK_MODEL`
  env var.** Opt-in rewrite when an anthropic dispatch arrives with a
  non-Claude model id. Empty (default) makes dispatch return an error
  instead — silent rewrites burned upstream credits on a model the
  caller never asked for. Operators who relied on the v0.2.62-and-earlier
  silent rewrite to `claude-haiku-4-5-20251001` can opt back in by
  setting this field.
- **Per-plugin `inline_no_think_for_qwen3` yaml field.** Backends
  whose chat template strips the qwen3-family `/no_think` inline
  marker (LM Studio's jinja, Ollama's /v1 layer) opt in via yaml.
  Replaces three duplicated copies of the rule that previously
  were scoped to hardcoded `serviceID == "lmstudio"` and only
  worked for backends the Go switch already knew about.
- **`WV_LMSTUDIO_URL` / `WV_VLLM_URL` / `WV_LLAMACPP_URL` env vars.**
  Operator-side override of the local backend URL without editing a
  plugin yaml. Resolution order is now env > plugin.DefaultURL >
  vault SSE > built-in default; env wins so a fleet operator can
  hot-swap the backend host without redeploying yamls.
- **`DefaultProxyOrigin(port, tlsEnabled)` helper.** Single source of
  truth for the canonical `scheme://localhost:<port>` origin
  configuration writers (openclaw_sync, agent_apply, setup) target.
  Replaces ~7 duplicated `https://localhost:56244` literals.
- **Plugin-driven dispatch + stream extension.** A plugin yaml with
  `request_format: openai` (or unset) is now picked up by
  `dispatchWithChain` and the Gemini-stream handler without
  requiring a Go-side `case` edit. New OAI-compat backends drop in
  as a yaml file.
- **OpenAI `reasoning_content` fallback in response converters.**
  Reasoning-only responses (qwen3.6 / deepseek-r1 / gpt-o1 family
  on backends that emit only `reasoning_content` and leave `content`
  empty) now surface the chain-of-thought as the response text
  instead of returning an empty string. Applies to both the
  non-stream (`OpenAIRespToGemini`) and stream (`streamOpenRouter`,
  `streamPluginAsGemini`) paths.
- **README TLS / cert setup section.** `wall-vault cert init` →
  `cert issue $(hostname)` → `cert install-trust` → enable TLS via
  env vars. Documents the loopback-only HTTP companion port and the
  http-vs-https tradeoff explicitly so a fresh installer doesn't
  read every `https://localhost:56244` example as the only mode.
- **README full config reference.** All YAML fields under
  `proxy:` / `vault:` / `doctor:` / `hooks:` listed with defaults +
  inline comments, plus all `WV_*` env vars in the env-var table.
  Was missing 16+ fields previously.
- **setup wizard now offers Anthropic / OpenAI / LM Studio / vLLM.**
  Previously asked only Google / OpenRouter / Ollama. Each option
  prints its corresponding env-var setup hint at the end.

### Changed

- **`openclaw_sync.go` no longer hardcodes `https://localhost:56244`.**
  `updateOpenClawJSON`, `healOpenClawConfig`, `healAgentSpecificModels`,
  `runStartupOpenClawHeal`, and `providerHealURLs` now take a
  `defaultOrigin string` arg derived from `cfg.Proxy.Port` and
  `cfg.Proxy.TLS.Enabled`. An operator who moves the proxy off port
  56244 or disables TLS no longer has the openclaw.json heal pass
  silently rewrite their config back to the wrong origin.
- **`dispatchWithChain` switch + `oaiCompatServices` set generalised.**
  The 6-case switch falls through to `callLocalService` when the
  service id matches an enabled plugin yaml; the OAI-compat set is
  merged with all `request_format: openai` plugins at boot. Both
  changes mean a drop-in plugin yaml is enough — no Go edit
  required for new backends.
- **`serviceNeedsKey` replaces `svc != "ollama" && …` chain.** Plugin
  yamls with `auth.type: none` / `bearer` / `""` skip the keyMgr
  cooldown gate the same way the historic local backends do.
- **Local backend URL resolution unified into `resolveLocalServiceURL`.**
  `callLocalService` / `streamLocalService` / `streamPluginAsGemini`
  share the same lookup path so future tweaks land in one place.

### Removed

- **Author's personal site/email links from the vault dashboard
  footer.** The footer now shows only the wall-vault GitHub link, so
  external operators no longer see the author's personal contact
  details on every install.

### Notes

- Anthropic silent rewrite is a behaviour change: any deployment
  that depended on the proxy quietly switching a non-Claude id to
  `claude-haiku-4-5-20251001` must set `proxy.anthropic_fallback_model`
  (or `WV_ANTHROPIC_FALLBACK_MODEL`) to opt back in.

---

## [0.2.62] — 2026-05-04

### Added

- **OAI-compat backend stream passthrough (opt-in).** A caller's
  `stream:true` against `/v1/chat/completions` now triggers a real
  backend stream forward when `WV_OAI_STREAM_FORWARD=1` is set and
  the resolved service is in the OAI-compat set (lmstudio, vllm,
  llamacpp, tgwui, localai, jan, koboldcpp, tabbyapi, mlx-server,
  litellm-proxy). Backend SSE chunks pipe straight to the caller
  with no aggregate buffering, and the chunk's `model` field is
  rewritten to the caller-resolved id so hub-mode topologies
  remain transparent. Stream-mode callers do NOT consult the
  fallback chain — design tradeoff documented in the spec at
  `docs/superpowers/specs/2026-05-04-oai-stream-passthrough-design.md`.
  Default off; existing v0.2.61 fake-chunk replay path stays
  unchanged when the flag is off.

---

## [0.2.61] — 2026-05-04

### Fixed

- **Hub-topology lmstudio plugin no longer drops the publisher
  prefix on the way out.** With a hub-mode plugin yaml (default_url
  pointing at another wall-vault) callLocalService used to forward
  the model id through its leading-prefix-strip step — so a request
  for "google/gemma-4-26b-a4b" arrived at the receiving wall-vault
  as a bare "gemma-4-26b-a4b". The hub then ran
  inferServiceFromBareModel on it, recognised the gemma- prefix as
  Google's family, and dispatched to the native Google handler,
  which 502'd with "Google: 모델 없음 (gemma-4-26b-a4b)". Symptom on
  host-A: telegram replies stopped landing because OpenClaw → host-A
  wall-vault → mini wall-vault hub all 502'd before reaching mini's
  LM Studio. New plugin field `preserve_model_id: true` opts out of
  the strip; host-A's lmstudio.yaml now sets it.

---

## [0.2.60] — 2026-05-04

### Added

- **Dispatch trace logging (opt-in).** `dispatchWithChain` now emits a
  single-line `[dispatch] requested=<svc>/<mdl> resolved=<svc>/<mdl>
  reason=<primary|fallback>` log on each successful dispatch when
  `WV_DISPATCH_TRACE=1` is set. Designed for live diagnosis of
  routing decisions (which service the call actually landed on, and
  whether a fallback fired) without grepping multiple log lines.
  Off by default — production hosts pay no log-volume cost.

---

## [0.2.59] — 2026-05-04

### Added

- **EconoWorld workDir cache (process-local).** /agent/apply now seeds
  an `atomic.Pointer[string]` cache with the resolved EconoWorld
  install dir, so subsequent SSE-driven `updateEconoWorldModel` calls
  and `runStartupEconoWorldHeal` operate on exactly the directory the
  operator pinned via the dashboard's `work_dir` field rather than the
  hard-coded fallback list. Hosts whose install lives outside the
  fallback list now stay in lockstep with the dashboard. Disable knob:
  `WV_ECONOWORLD_WORKDIR_CACHE_DISABLE=1`.

### Changed

- **wslHomeCandidates extracted as a single helper.**
  `findClineDataDir` and `findClaudeSettingsPaths` previously
  enumerated the same three-source candidate list (`$HOME`,
  `$USERPROFILE`-as-WSL, `/mnt/c/Users/*`) inline. The order and
  filtering rules had begun to drift between the two; consolidated
  into one helper so adding a fourth root or tightening the filter
  is a single edit.

---

## [0.2.58] — 2026-05-03

### Added

- **Self-managed sentinel token substitute (loopback only, opt-in).**
  Fresh agent installs frequently ship with a placeholder token like
  `proxy-managed` or `dummy` in their config (OpenClaw's
  `models.providers.custom.apiKey`, NanoClaw's equivalent) and rely on
  the proxy's heal pass to rewrite the field to the real `VaultToken`.
  When the heal hadn't yet run — for example, a freshly installed
  client whose gateway started before the proxy's first heal cycle —
  every request from that client was a hard 401 dead-end. The proxy
  now recognises these sentinels at the request boundary and, when the
  operator sets `WV_TOKEN_SENTINEL_FALLBACK=1` and the caller is on the
  loopback interface, substitutes the proxy's own `VaultToken` so the
  request can resolve normally. Both `Authorization: Bearer …` and
  Anthropic-style `x-api-key` headers are covered. Loopback is the
  security boundary: only a process already on the same host can use
  these sentinels, which matches the trust level local agents already
  enjoy via filesystem access to the same config. New env knob:
  `WV_TOKEN_SENTINEL_FALLBACK` (default off).

### Changed

- **Provider-prefix hijack policy consolidated into a single set.**
  The list of OpenAI-compatible backends that publish path-style model
  ids (`publisher/model`) was previously enumerated in three places
  (two case-by-case `if` blocks per backend in
  `parseProviderModelDepth`, plus the dispatch case label). It is now
  one source of truth — `oaiCompatServices` — used for both directions
  of the parser: caller chose one of these as `svc` (honour the choice
  regardless of the body model's prefix), and caller wrote one of
  these as the model prefix (route to that backend). Adding a new
  OAI-compat plugin yaml only needs that single set updated.

---

## [0.2.57] — 2026-05-03

### Fixed

- **Dashboard "use service default" now actually applies on dispatch.**
  Previously, picking "(use service default)" in a client's edit panel
  cleared `entry.model` in vault, but the proxy's request-handling path
  still let the request body's `model` win — so a client whose
  `ai_config.json` carried a stale model id silently bypassed the
  operator's dashboard knob. Now, when a token-resolved client has an
  empty `default_model` and a non-empty `default_service`, the proxy
  forces the request body's model to the service-level
  `serviceDefaults[svc]`. Surfaced when an OAI-compat client kept
  a stale path-style model id in its body even after the operator
  switched the LM Studio service default to a different model. Both
  the OpenAI-compat and Anthropic dispatch paths apply the same
  fallback.
- **LM Studio model ids stop getting hijacked into other services'
  handlers.** LM Studio publishes models as `publisher/model-id`
  (e.g. `google/gemma-4-26b-a4b`, `qwen/qwen3.6-27b`,
  `anthropic/claude-…`), which collided with the proxy's existing
  `provider/...` parsing — a request to `lmstudio` carrying
  `google/gemma-4-26b-a4b` was force-routed into the native Google
  service and 502'd with `모델 없음`. Now, when the caller explicitly
  chose `lmstudio`, the publisher prefix is honoured as part of the
  model id and the request stays on lmstudio (mirrors the existing
  OpenRouter exemption).
- **LM Studio Qwen3 inline `/no_think` token now mirrors the ollama
  path.** LM Studio's `/v1/chat/completions` silently ignores the
  top-level `think` field, same as ollama, so Qwen3 burns
  `max_tokens` on hidden reasoning when reasoning is off in vault. The
  inline `/no_think` marker is appended to the last user message — but
  only when `serviceID=="lmstudio"` and the model id starts with
  `qwen3*`, so other OAI-compat backends (vllm/llamacpp/…) whose jinja
  templates may not strip the marker are not affected.
- **`<script id="wv-i18n-data">` no longer ships a literal
  `@templ.Raw(...)` to the browser.** templ's parser does not evaluate
  expressions inside raw text elements like `<script>`, so the JSON
  blob expression was emitted verbatim into the page; the bootstrap
  parser then caught the JSON.parse error and left `WV_I18N` as `{}`,
  which made every JS-side i18n key (optgroup labels, model dropdown
  placeholder, etc.) fall through to the English literal even in
  Korean mode. The whole `<script>` element is now emitted via
  `@templ.Raw("<script…>" + I18nJSONBlob(lang) + "</script>")`.
- **`I18nJSONBlob` is no longer dropped from non-`js` builds.** The
  helper lived in `i18n_js.go`, which Go interprets as a `GOARCH=js`
  filename suffix and quietly excludes from every other build target —
  the function was missing entirely from the linux/darwin/windows
  binaries and the `wv-i18n-data` blob would have been broken even
  with the templ fix above. Renamed to `i18n_jsdata.go`.

### Added

- **EconoWorld ai_config.json gains stream / request_timeout_seconds
  fields and a configurable max_tokens floor.** The previous bootstrap
  hard-coded `max_tokens=4096`, which truncated long Korean analyses
  mid-sentence, and never wrote a `stream` flag — so the openai-compat
  client paid the full ollama cold-load latency as a single silent wait
  before getting any output. New `ProxyConfig.EconoWorldMaxTokens`
  (default 8192), `EconoWorldStream` (default true), and
  `EconoWorldRequestTimeout` (default 300s) are written on bootstrap
  and survive subsequent `/agent/apply` calls — operator-tuned values
  in the file are never clobbered. Override per host with
  `WV_ECONOWORLD_MAX_TOKENS` / `WV_ECONOWORLD_STREAM` /
  `WV_ECONOWORLD_REQUEST_TIMEOUT`.
- **Boot heal pass for EconoWorld config.** Mirrors the OpenClaw heal
  added in v0.2.51-54: at proxy startup, any same-host (localhost /
  127.0.0.1) `base_url` in `analyzer/ai_config.json` is rewritten to
  the loopback plain-HTTP companion when one is enabled, sidestepping
  the self-signed-CA trust problem the same way OpenClaw does. Heal
  also fills missing `stream` / `request_timeout_seconds` fields and
  bumps the legacy `max_tokens=4096` to the configured floor; any
  other operator value is left alone. External (LAN) `base_url`s are
  never touched. Hosts without an EconoWorld install no-op silently.
  Set `WV_ECONOWORLD_HEAL_DISABLE=1` to skip the heal entirely.

## [0.2.56] — 2026-05-03

### Fixed

- **Heal pass relaxes the gateway's channel-stale threshold to 60
  minutes.** Mini gateway 2026-05-03 spent the night in a 300-second
  SIGTERM-restart loop that survived even after v0.2.55 aligned
  active-memory's model: every five minutes the gateway logged
  `[gateway] signal SIGTERM received` and launchd brought it back up.
  Tracing schema docs revealed
  `gateway.channelStaleEventThresholdMinutes` — "How many minutes a
  connected channel can go without provider-proven transport activity
  before the health monitor treats it as a stale socket." With the
  threshold left short (effectively the 5-minute health-check
  interval), an idle telegram bot triggered restarts the moment a
  human paused typing. Boot heal now floors the value at 60 minutes,
  preserving OpenClaw's failure-detection logic for sockets that
  genuinely die while letting quiet bots stay connected. Idempotent —
  operator settings of 60+ are left alone.

---

## [0.2.55] — 2026-05-02

### Fixed

- **Heal pass aligns active-memory plugin model with the agent's
  primary.** Mini gateway 2026-05-02 was caught in a 300-second
  SIGTERM-restart loop: OpenClaw's `active-memory` plugin shipped
  with `model: custom/gemini-2.5-flash-lite`, but the host's vault
  has no google credentials in its services list, so every plugin
  tick failed to dispatch (`No callable tools remain after resolving
  explicit tool allowlist`). The gateway's health-monitor reads that
  failure as an unhealthy signal and SIGTERMs itself; launchd
  restarts; 300s later the same fail; restart again — every
  in-flight TUI/Telegram turn dies with the gateway.

  Boot heal now walks `plugins.entries.active-memory.config.model` and
  rewrites it to match `agents.defaults.model.primary` whenever the
  two diverge. Idempotent: change the agent primary and the plugin
  resyncs on next boot. No-op when active-memory isn't configured or
  when the agent has no primary set, so we never materialize plugin
  config the operator deliberately left empty.

---

## [0.2.54] — 2026-05-02

### Added

- **Loopback-only plain-HTTP companion listener for the proxy.** When
  `Proxy.TLS.Enabled` is true and `Proxy.PlainPort` is non-zero
  (default `56245`, env `WV_PROXY_PLAIN_PORT`), wall-vault binds a
  second HTTP server to `127.0.0.1:<PlainPort>` with TLS off and the
  exact same handler. Same-host clients that cannot honour the proxy's
  self-signed CA — most pressingly OpenClaw, whose macOS daemon
  rewrites `NODE_EXTRA_CA_CERTS` from the operator's value to
  `/etc/ssl/cert.pem` at spawn ((operator host, earlier)), but the same class
  of issue exists on any client that uses the system trust store
  exclusively — now have a path to the proxy that doesn't depend on
  trust-store configuration. LAN callers (other fleet machines using
  `ca.crt`) continue to use the TLS listener untouched, and the
  loopback bind keeps the vault token's exposure inside the same OS
  user's loopback interface (same trust boundary as a Unix socket).
  Set `WV_PROXY_PLAIN_PORT=0` to disable.

### Fixed

- **Heal pass routes same-host OpenClaw to the plain companion when
  one is active.** When the proxy boots with the plain companion
  enabled, `runStartupOpenClawHeal` rewrites
  `models.providers.{custom,anthropic,google}.baseUrl` from
  `https://localhost:56244...` to `http://127.0.0.1:56245...` on both
  `~/.openclaw/openclaw.json` and every per-agent
  `~/.openclaw/agents/<id>/agent/models.json` cache. Without the
  companion (TLS off, or plain port disabled) the heal preserves its
  v0.2.51 behaviour and leaves any localhost URL alone. Operators
  running OpenClaw on the same host as wall-vault no longer have to
  install the wall-vault CA into their system trust store, and the
  fix is OS- and client-agnostic — the `request.tls.ca` heal
  (v0.2.53) is kept around as a hint for clients that do honour it,
  but the plain companion is the load-bearing path.

---

## [0.2.53] — 2026-05-02

### Fixed

- **Heal pass also writes provider-level `request.tls.ca`.** Mini
  gateway 2026-05-02 surfaced the next layer of breakage after the
  v0.2.52 `data:` chunk fix: TUI and Telegram dispatches still failed
  with `Connection error`, and `wall-vault-proxy.err` revealed the real
  cause — `http: TLS handshake error from 127.0.0.1: EOF`. The
  OpenClaw daemon wrapper rewrites `NODE_EXTRA_CA_CERTS` from the
  plist value (`~/.wall-vault/ca.crt`) to the system bundle
  (`/etc/ssl/cert.pem`) before spawning the gateway Node process, and
  the gateway also sets `NODE_USE_SYSTEM_CA=1`. Our self-signed CA is
  in neither place, so every embedded fetch to the proxy hits the
  system trust store, fails verification, and surfaces as the generic
  `Connection error` from undici.

  Boot heal now writes the proxy's CA bundle path into
  `models.providers.{custom,anthropic,google}.request.tls.ca` on both
  the main `openclaw.json` and every per-agent cache. OpenClaw reads
  that field directly from config and threads it into the per-provider
  fetch's TLS connect options, bypassing the env-rewriting behaviour
  entirely. The bundle is discovered alongside the proxy's TLS cert
  (`<dir-of-Proxy.TLS.CertFile>/ca.crt`); when the file isn't there
  (HTTP-only deployments, missing companion bundle), heal silently
  skips the TLS-CA write and otherwise behaves identically.

  Operator-set sibling TLS fields (`request.tls.serverName`,
  `request.tls.cert`, etc.) are preserved — the heal only adds/updates
  the `ca` key.

---

## [0.2.52] — 2026-05-02

### Fixed

- **Stream early-flush emits a parsable `data:` chunk instead of an SSE
  comment.** v0.2.49 wrote `: warming up` and `: keepalive` comment
  frames to defeat first-byte timeouts; raw `fetch` accepted them but
  the OpenAI Node SDK that OpenClaw embeds rejected the stream as
  `Connection error` ~14s into the call ((operator host, earlier) gateway
  gateway.err.log: 4× retry / lane durationMs ≈ 75s — TUI and
  Telegram dispatches both failed before any data could arrive). The
  SDK seems to require the first frame to be a parsable `data:` chunk
  before it considers the stream live. v0.2.52 ships a stable empty-
  delta no-op chunk (`data: {"id":"chatcmpl-warmup",...,
  "choices":[{"index":0,"delta":{},"finish_reason":null}]}`) for both
  the early-flush frame and the keepalive ticks. Empty deltas merge to
  the empty string and are treated as no-ops by well-behaved SSE
  consumers, but they count as real frames for the SDK's stream-start
  and idle counters.

  Tick interval reduced from 15s to **8s** — still well below typical
  60s+ idle timeouts, but short enough that even a 14s SDK quirk gets
  two frames before tripping.

---

## [0.2.51] — 2026-05-02

### Fixed

- **Heal pass also normalizes the per-agent provider cache.** OpenClaw
  maintains a parallel provider config at
  `~/.openclaw/agents/<id>/agent/models.json` — same providers section
  as the main `openclaw.json` but at the top level (no
  `models.providers` wrapper) — and consults it during embedded-agent
  dispatch. On hosts upgraded from pre-v0.2.37 the cache lags the main
  config: even after `healOpenClawConfig` rewrites
  `baseUrl=https://...` and `apiKey=<vault token>` on the main file,
  the cache still holds `baseUrl=http://localhost:56244/v1` +
  `apiKey="dummy"` + `authHeader=false` (observed (operator host, earlier)).
  Every embedded dispatch then either fails the TLS handshake
  (HTTP→HTTPS mismatch on the local proxy) or trips the post-v0.2.39
  token-auth gate before reaching an LLM, and OpenClaw silently
  rotates to a fallback model that is not even present locally.

  Boot heal now sweeps every per-agent cache it finds and applies the
  same normalization pass (baseUrl + apiKey + authHeader) using a new
  strict `forceExactBaseURL` helper that, unlike the existing
  `forceLocalhostBaseURL`, also rewrites local-proxy URLs that
  disagree on scheme — so `http://localhost:56244/v1` is healed to
  `https://localhost:56244/v1` when wall-vault is running over TLS.
  The main `openclaw.json` heal keeps its existing behaviour
  (rewrites only non-localhost URLs) so we don't accidentally flip a
  deliberately-HTTP localhost setup on standalone hosts.

  Third-party providers in the cache (e.g. native ollama) are
  untouched, same rationale as the main heal — we only normalize
  providers known to be wall-vault-fronted.

---

## [0.2.50] — 2026-05-02

### Added

- **"Remember this device" toggle on the login page.** The dashboard
  login form now carries a toggle switch (`이 기기 기억하기 (30일)`)
  that, when enabled, issues an HMAC-signed long-lived session cookie
  instead of the default 12-hour in-memory session. The cookie value
  is `<expiryUnix>.<hmacHex>` where the HMAC key is the admin token
  itself, so it survives process restarts (no server-side store
  required) and rotating the admin token instantly invalidates every
  outstanding remember cookie — the only revocation path, by design.
  Plain non-remember sessions keep the existing 12-hour in-memory
  behaviour. The toggle is i18n'd across all 17 supported locales.

  Operators no longer have to re-paste the admin token after every
  binary restart or 12-hour wall clock — tick the toggle once per
  trusted browser and the dashboard stays logged in for 30 days.

---

## [0.2.49] — 2026-05-02

### Fixed

- **SSE early-flush for `stream:true` /v1/chat/completions.** When an
  upstream local model takes 60-180s on a cold/large prompt (e.g.
  qwen3.6:27b prompt-eval on a fresh process), the proxy used to hold
  the wire silent until the upstream returned, which caused
  first-byte-timeout-bound callers — most notably OpenClaw embedded
  agents (~75s default) — to abort the request well before any data
  was produced. From v0.2.49 the proxy now commits status 200 + SSE
  headers and emits a `: warming up` comment frame *before* dispatch
  starts, then runs a 15s `: keepalive` ticker until the first real
  chunk is ready. SSE comments (lines starting with `:`) are ignored
  by parsers but reset the client's idle timer, so OpenClaw and any
  other SSE consumer waits for the full upstream latency without
  aborting.

  Error path: when dispatch fails after the early-flush, the proxy
  emits the error as a final SSE chunk + `data: [DONE]` instead of
  trying to switch to a JSON 502 (whose status would already be
  committed). Non-stream callers (`stream:false`) are unaffected — the
  early-flush block only runs when the request asked for streaming.

---

## [0.2.47] — 2026-05-02

### Added

- **Heal pass also normalizes provider apiKey + authHeader.** The boot
  heal in `internal/proxy/openclaw_sync.go` now writes the proxy's own
  vault token into `models.providers.{custom,anthropic,google}.apiKey`
  and forces `authHeader: true` whenever the existing values look
  pre-v0.2.37 (literal `"dummy"` / `"proxy-managed"` / empty + the
  default `authHeader: false`). Triggered by host-A + host-B
  2026-05-02: both still carried `apiKey: "dummy"` from a year-old
  install, so every OpenClaw → wall-vault call after the v0.2.39
  token-auth gate landed surfaced as `401 token not registered with
  vault`. With this heal, a stale config self-corrects on next proxy
  boot — operators don't need to know what changed or run `agent apply`
  by hand.

  Third-party providers (anything outside the wall-vault-fronted
  custom/anthropic/google trio) are left untouched — we don't know
  their auth scheme.

  No-op when `cfg.Proxy.VaultToken` is empty (standalone mode without
  vault), so the auth heal can't blank a working apiKey.

---

## [0.2.46] — 2026-05-02

### Added

- **Heal pass also normalizes the `google` provider.**
  `normalizeOpenClawProviders` now runs the same upstream-host →
  localhost rewrite on `models.providers.google.baseUrl` that v0.2.43
  introduced for `custom` and `anthropic`. Triggered by host-B
  2026-05-02: the google provider had `http://<internal-host>:11434/v1`
  written into its baseUrl slot — an ollama URL accidentally landed in
  the google provider — so every OpenClaw call addressed
  `custom/gemini-2.5-flash` (which `parseProviderModel` correctly
  routed to the google provider) ended up on ollama and 404'd with
  `model 'gemini-2.5-flash' not found`. Heal forces it back to
  `https://localhost:56244` (OpenClaw appends its own `/google/v1beta`
  path).

---

## [0.2.45] — 2026-05-02

### Added

- **`doctor fix-trust` — auto-inject the wall-vault internal CA into
  local AI agent runtimes.** A new doctor subcommand walks every known
  agent supervisor on the host (systemd `--user` units on Linux,
  launchd plists on macOS) and writes a `wall-vault-trust.conf`
  drop-in (Linux) or `EnvironmentVariables` entry (macOS) that sets
  `NODE_EXTRA_CA_CERTS=~/.wall-vault/ca.crt`. After the agent is
  bounced, its Node process trusts the wall-vault HTTPS endpoint
  natively — no `NODE_TLS_REJECT_UNAUTHORIZED=0`, no system-wide cert
  install, no TLS downgrade.

  Covers OpenClaw gateway, Claude Code, Cline today; new agents are a
  one-line addition to `knownAgents` in
  `internal/doctor/fix_trust.go`. Skips silently when an agent's unit
  isn't present on the current host (most fleet members), so the same
  command is safe to run on every machine.

  Triggered by (operator host, earlier): OpenClaw failed every LLM call with
  `network connection error` because Node could not validate the
  self-signed wall-vault cert. The right fix could not be a
  systemd-level `WV_PROXY_TLS_ENABLED=0` override (which weakens TLS
  on shared infrastructure); it had to be a generic "teach the local
  agent to trust the local CA" path that any third-party deployment
  could reuse without touching their OS trust store.

- **Auto-trust on first setup.** `wall-vault setup` now calls
  `doctor.FixTrust` after writing the config and prints the per-agent
  result. A fresh install therefore configures the agent trust as part
  of the standard setup flow — operators don't have to know
  `fix-trust` exists.

- **`doctor all` includes fix-trust.** The catch-all health command
  runs the trust pass at the end so periodic doctoring keeps agents'
  drop-ins current after a CA rotation.

---

## [0.2.44] — 2026-05-01

### Added

- **Generalized local LLM dispatch via plugin yaml.** The
  `internal/config/services.go` `ServicePlugin` schema (which already
  carried `auth`, `endpoints`, `request_format`, etc. but was never
  consulted by the dispatch path) is now wired into
  `internal/proxy/server.go` `callLocalService`. Every OpenAI-compat
  local backend — LM Studio, vLLM, llama.cpp, text-generation-webui,
  LocalAI, Jan, KoboldCpp, TabbyAPI, mlx_lm.server, LiteLLM proxy, and
  *another wall-vault instance running in hub mode* — flows through one
  code path; the per-backend differences are declared in yaml.

  New schema fields:
  - `default_url` — fallback URL when the SSE-distributed
    `serviceURLs[id]` hasn't been populated yet (handy for fresh
    machines / standalone hosts).
  - `default_model` — fallback model id with the same semantics.
  - `tls_internal_ca` — when `true`, outbound calls use the proxy's
    `internalHTTPClient`, which trusts `~/.wall-vault/ca.crt`. Lets a
    client reach a self-signed wall-vault hub over HTTPS without
    OS-level CA installation.
  - `auth.type: bearer` — adds `Authorization: Bearer <vault_token>`
    using the proxy's already-resolved vault token.

  Backward compat: every existing plugin yaml without the new fields
  loads unchanged and behaves exactly as before. Plugin-less hosts
  (no `~/.wall-vault/services/`) keep the pre-v0.2.44 dispatch path.

  Resolution rules (top wins):
  1. Plugin `default_url` beats vault-distributed `serviceURLs[id]`.
     A plugin yaml is the operator's explicit local override; without
     this rule a vault default URL pointing at an unreachable
     localhost backend (e.g. LM Studio listening on 127.0.0.1 on a
     different host) silently masks the operator's hub plugin.
  2. A plugin with `enabled: true` whose id is not yet in
     `proxy.services` is auto-appended at boot, so an installed
     plugin actually gets dispatched to (otherwise the service-list
     gate would reject it before reaching `callLocalService`).
  3. `parseProviderModel` returns the full `<id>/<rest>` model string
     (not just the bare tail) for every local backend, so org-scoped
     ids like `qwen/qwen3.6-27b` survive the prefix-stripper inside
     `callLocalService`. New backend prefixes (`lmstudio`, `vllm`,
     `llamacpp`, `tgwui`, `localai`, `jan`, `koboldcpp`, `tabbyapi`,
     `mlx-server`, `litellm-proxy`) all share the same path.

- **Reference plugin yamls under `configs/services/`.** Added 9 new
  examples + 1 commented hub template (`wall-vault-hub.yaml.example`),
  all `enabled: false` so they sit dormant until an operator copies
  them into `~/.wall-vault/services/` and flips the flag. Covers the
  major OpenAI-compat local backends in current use (2026-05).

- **Plain-HTTP-to-remote-host warning at boot.** A plugin whose
  `default_url` or `endpoints.generate` resolves to a non-localhost
  host with `http://` scheme produces a single
  `[plugin] warn: <id> url=… reaches remote host over plain HTTP …`
  log line. Local hosts (`localhost`, `127.0.0.1`, `::1`,
  `*.local`, `*.localhost`) are exempt — an http LM Studio on the
  same machine is the normal case.

### Triggered by

- (operator host, earlier) incident: telegram bot couldn't reach LM Studio
  because the vault-distributed service URL `http://<internal-host>:1234`
  pointed at LM Studio's localhost-only listener on a different
  machine, with no path through the mini's wall-vault hub. The fix
  could not be a hardcoded host-A/mini IP edit; it had to be a generic
  "client wall-vault forwards to hub wall-vault" pattern that any
  third-party deployment (cloud GPU box + laptop, office LM Studio +
  N seats, etc.) could reuse without code changes. v0.2.44 makes that
  pattern a one-line plugin yaml.

---

## [0.2.43] — 2026-05-01

### Added

- **OpenClaw config heal pass at boot.** `internal/proxy/openclaw_sync.go`
  now ships `healOpenClawConfig` (runs once at proxy startup alongside
  the existing sanitize pass) plus a stronger `updateOpenClawJSON`
  steady-state writer. Both share three normalization rules:
  1. `models.providers.{custom,anthropic}.baseUrl` is forced back to
     `https://localhost:56244{,/v1}` whenever it points at any non-local
     host. This catches stale configs where an upstream URL was written
     directly into OpenClaw and the proxy was being bypassed entirely
     ((operator host, earlier): `http://<internal-host>:11434/v1` written into
     `providers.custom.baseUrl`, breaking auth and routing for every
     non-ollama service).
  2. `models[]` entries with empty `id`, dangling-name (e.g.
     `"openrouter / "`), or duplicate `id` within the same provider are
     pruned. The host-A snapshot carried 11 entries with identical
     `id="qwen3.6:27b"` and differing names, which collapsed OpenClaw's
     model selector into a single resolvable model and surfaced as the
     `Model "custom/" could not be resolved` log spam.
  3. `agents.defaults.model.primary` is rewritten when it holds a
     dangling `"<provider>/"` reference ((operator host, earlier): SSE
     config_change wrote `primary="custom/"` and OpenClaw rejected
     every gateway start with `Invalid model reference: custom/`). The
     first usable entry from `fallbacks` takes the primary slot; if no
     usable fallback exists, the primary key is removed so OpenClaw
     falls back to its own default resolution.
  4. Sanitize pass (empty-id dropper) still runs first so heal sees a
     pre-cleaned input.

- **Empty-model SSE guard.** `updateOpenClawJSON` now returns early
  when `model == ""`. A vault soft-clear or pre-resolution
  config_change used to fire with empty model, materialize
  `primary="<service>/"`, and break OpenClaw's model selector until
  the next non-empty change arrived (often hours later, after the
  next user-driven model switch).

  No-op for hosts without `~/.openclaw/openclaw.json` (most fleet
  members). Writes the file only when something actually changes.

---

## [0.2.42] — 2026-05-01

### Added

- **OpenClaw config sanitizer.** `internal/proxy/openclaw_sanitize.go`
  filters empty-id model entries (`models.providers.<provider>.models[]`)
  from `~/.openclaw/openclaw.json` so OpenClaw 2026.4.29's stricter
  schema validation doesn't crash-loop the gateway. Fires once at proxy
  boot (no-op for hosts without OpenClaw) and is also exposed as
  `wall-vault doctor sanitize-openclaw`. Backs up the original to
  `*.bak.sanitize` before rewriting; only writes when something actually
  changes so clean configs see no churn. Triggered by host-A observation
  on 2026-05-01: a single empty-id entry left over from a pre-v0.2.32
  applyOpenClawConfig caller crash-looped the gateway after restart.
- **17-language i18n on auth + bootstrap pages.** `/setup`, `/login`,
  the bootstrap CA-distribution index page, and the dashboard header's
  theme dropdown all now route every prose string through `i18n.TFor`.
  Lang resolution honours `?lang=xx` override → `Accept-Language` →
  server default. Adds a small `lang_match.go` helper that parses the
  q-weighted Accept-Language list. New keys (`auth_*`, `bs_*`,
  `theme_*`) ship in all 17 locales (ko, en, ja, zh, es, de, fr, pt,
  id, th, hi, ar, mn, ne, sw, zu, ha) — code tokens and proper nouns
  stay verbatim in every locale.

---

## [0.2.41] — 2026-05-01

### Added — CA bootstrap listener + token-auth diagnostic messages

- **Plain-HTTP bootstrap listener (default :56247).** Breaks the catch-22
  where new clients need the wall-vault CA to speak HTTPS to the main
  vault listener but the CA itself was only reachable behind that HTTPS.
  Listener serves only `/ca.crt` (PEM download), `/` (per-OS install
  instructions for Linux/macOS/Windows + Python `SSL_CERT_FILE` /
  Node.js `NODE_EXTRA_CA_CERTS` snippets), and `/health`. CA cert is
  public information by design — exposing it without auth is safe.
  Disable with `vault.bootstrap_port: 0` (or `WV_VAULT_BOOTSTRAP_PORT=0`).
- **`tokenAuthFail` replaces the catch-all `"invalid token"` 401.**
  `requireProxyToken` and `requireAnthropicToken` now distinguish:
  - 401 `"token not registered with vault"` — vault returned 401/403/404
  - 503 `"vault unreachable"` — network / dial / timeout
  - 502 `"vault returned an unexpected response"` — vault 5xx or
    malformed JSON
  - 401 `"proxy.vault_token mismatch"` — proxy token configured but
    didn't match (no vault fallback)
  - 503 `"no auth configured"` — neither proxy.vault_token nor
    proxy.vault_url set
  Real ops cost prompted this — board #43 documented operators chasing
  IP whitelist for an hour when the cause was a stale token cache.
- Tests: bootstrap handler routes (ca.crt + index + 404 hint), CA path
  resolution fallback chain, every tokenAuthFail branch's wire-format.

---

## [0.2.40] — 2026-05-01

### Added — Ollama latency reduction

- **`keep_alive` on every Ollama call (default 30m).** Ollama's own default
  unloads the model after 5 minutes idle, which on the 27B fleet model on
  mini meant every call separated by more than 5 minutes paid an 80-113s
  cold reload. The proxy now sends `keep_alive=30m` on every Ollama
  request so the model stays warm between sparse calls. Tunable via
  `WV_OLLAMA_KEEP_ALIVE` (e.g. `-1` for forever, `10m` for tighter RAM,
  `0` to disable). Recent Ollama (>=0.3.x) honours top-level keep_alive
  on the OpenAI compat endpoint; older versions silently ignore — no
  regression either way.
- **`num_ctx` (default 8192).** Pin Ollama's context window so long
  Korean conversations + tool-call payloads don't get truncated by the
  2048-token built-in default. Tunable via `WV_OLLAMA_NUM_CTX`.
- **HTTP connection pooling for Ollama.** Pre-v0.2.40 every callOllama
  built a fresh `http.Client` whose connection died straight to TIME_WAIT
  after the response. The proxy now keeps one long-lived client with
  `MaxIdleConns=10`, `MaxIdleConnsPerHost=10`, `IdleConnTimeout=120s`
  per Server instance, saving the syscall+handshake cost on every
  follow-up request.
- Tests cover the JSON wire-format invariants (keep_alive/num_ctx
  appear when set, are absent when nil) and env-var override behaviour.

Note: timeout was already 600s (10 min) in `internal/proxy/server.go` —
the user's "raise to 600s" intuition was right but the lever was already
pulled. The actual bottleneck was the missing keep_alive hint.

---

## [0.2.39] — 2026-05-01

### Added — OpenClaw version-aware config writes

- **`internal/proxy/openclaw_version.go`** — detects the local OpenClaw
  install via `package.json` (search order: ~/.npm-global, /usr/lib,
  /usr/local/lib, /opt/homebrew/lib), parses CalVer (year.month.patch),
  and exposes a `schemaTag()` so future config-schema forks plug into
  `applyOpenClawConfig` without rewiring the dispatcher.
- **`agent_apply.go`** — logs detected version on every apply
  (`[agent-apply] openclaw version=2026.4.29 (schema=v1)`) and writes
  `meta.lastTouchedVersion` + `meta.lastTouchedSchema` into
  `~/.openclaw/openclaw.json` so an operator grepping the file later
  knows which writer touched it last. Today every reachable version
  shares schema=v1 so the actual write path is unchanged; the scaffold
  is in place for the day OpenClaw breaks the providers layout.
- Tests for parser + gte() + schemaTag() invariant.

Note: the gateway-restart failures users reported on host-B/mini after
`openclaw update` to 2026.4.29 are unrelated to wall-vault — research
confirmed the providers schema is unchanged; the failures are
OpenClaw's own lifecycle bugs (stale PID holding port 56242 on mini,
health-probe race on host-B). Recovery is operator-side: stop the
stale gateway, retry update.

### Added — Click-to-claim onboarding

- **Web-based first-run claim.** Fresh installs no longer need `wall-vault
  setup` ahead of time. The vault dashboard at `/` now redirects browsers
  to `/setup` when `admin_token` is unset; clicking "초기화 진행" generates
  `admin_token`, `proxy.vault_token`, and `master_password`, persists them
  to the loaded config file, and issues a session cookie. Loopback-only on
  the claim POST so an attacker on the LAN cannot front-run the legitimate
  operator.
- **Session cookie auth.** `adminAuth`/`clientAuth` accept either a
  Bearer token (CLI/API path, unchanged) or a `wv_session` cookie issued
  by `/login` or `/setup`. Cookies last 12 hours and are wiped on process
  restart so a stolen cookie cannot outlive the binary that issued it.
  HTMX requests from the dashboard now ride the cookie automatically.
- **`/login` page.** Subsequent visits without a session redirect to a
  password form that takes the admin token. Failed attempts feed the
  existing per-IP rate limiter (10 fails / 15 min → 429).
- **`/logout`.** Revokes the cookie and bounces back to `/login`.
- **Dashboard auth gate on `/`.** Pre-v0.2.39 the unauthenticated GET `/`
  shipped `admin_token` baked into the HTML meta tag — anyone on the LAN
  who could reach the dashboard could grab it. `/` now requires either
  the cookie or the Bearer token before rendering, closing that leak.
- **`doctor check` security audit.** Surfaces empty `admin_token` /
  `proxy.vault_token` / `master_password`, plus `tls.enabled=true` with
  missing cert files. Same lines also rendered by `doctor status`.
- **Startup hint.** `wall-vault start` against an empty config now logs
  a banner pointing the operator at `http://localhost:<port>/setup`.

### Changed

- Removed the warning banner about an unset admin token from the vault
  startup logs — the hint above replaces it and points at the fix
  instead of just naming the problem.

---

## [0.2.38] — 2026-04-29

### Added

- **`wall-vault cert install-trust` subcommand.** Detects current OS
  (Linux/macOS/Windows) and registers `~/.wall-vault/ca.crt` into the
  system trust store: Linux installs to `/usr/local/share/ca-certificates/`
  + `update-ca-certificates`, macOS uses `security add-trusted-cert -d -r
  trustRoot -k /Library/Keychains/System.keychain`, Windows uses `certutil
  -addstore -f Root`. Linux/macOS auto-prepend `sudo` when not root;
  Windows must run from an elevated prompt. Replaces the per-OS manual
  install instructions that previously had to be run by hand on each
  machine.

---

## [0.2.37] — 2026-05-01

### Added — Security hardening (Phase 1+2+3)

- **Admin endpoint gate.** `PUT /api/config/model`, `POST /reload`,
  `POST /api/config/think-mode` now require `Authorization: Bearer <proxy
  vault_token>`. Client tokens are explicitly rejected for these — they
  mutate proxy-wide state, which a per-client token has no authority over.
  Fail-closed when `proxy.vault_token` is unset (503).
- **Proxy gate on `/v1/*`, `/api/models`, `/google/*`.** All AI routing
  endpoints now require either the proxy's own VaultToken or a
  vault-registered client token. `/v1/messages` additionally accepts
  Anthropic-native `x-api-key` and BYO `sk-ant-*` (Claude Code OAuth via
  NanoClaw). Pre-v0.2.37 the proxy fell back to its own client config when
  no token was present, which let any LAN-reachable caller use proxy
  credentials.
- **Internal CA + per-host certificate tooling.** New `wall-vault cert`
  subcommand: `init` generates an ECDSA P-256 CA at `~/.wall-vault/ca.{crt,key}`
  (10-year), `issue <host> [ip…]` mints host certs (5-year, SAN = hostname
  + localhost + 127.0.0.1 + ::1 + extra IPs), `list` enumerates issued
  certs. CA never re-signs intermediate CAs (`MaxPathLenZero`).
- **TLS for proxy + vault listeners.** New `proxy.tls.{enabled,cert_file,
  key_file}` and `vault.tls.{...}` config blocks plus
  `WV_PROXY_TLS_ENABLED` / `WV_PROXY_TLS_CERT` / `WV_PROXY_TLS_KEY` (and
  `WV_VAULT_TLS_*`) environment overrides. Default is plain HTTP for
  backwards compatibility; the fleet enables TLS by setting the env vars
  in the systemd / launchd unit.
- **cokacdir signal-light = solid blue (`#2563eb`).** On-demand agents
  previously rendered as a hollow ring; now they show as a distinct
  blue dot, separate from green (active), amber (warn), grey (disabled).

### Notes — Phase 3 fleet rollout caveats

- Each host needs the issued cert/key under `~/.wall-vault/<host>.{crt,key}`
  and `ca.crt` registered in the OS trust store (Linux:
  `update-ca-certificates`, macOS: `security add-trusted-cert -d -r
  trustRoot -k /Library/Keychains/System.keychain`). macOS GUI confirms
  cannot be bypassed via SSH.
- Node clients (lab-sns) need `NODE_EXTRA_CA_CERTS` since Node ships its
  own ca-bundle and ignores the system trust store.
- macOS `/usr/bin/curl` uses LibreSSL with `/etc/ssl/cert.pem`, not the
  Keychain — callers must pass `--cacert ~/.wall-vault/ca.crt` explicitly.
- Python `urllib` and Linux `curl` pick up the system trust automatically
  once `update-ca-certificates` runs.

### Changed

- `openclaw_sync.go` writes `https://localhost:56244` baseUrls into
  OpenClaw's `models.json` (was `http://`).

---

## [0.2.34] — 2026-04-27

### Fixed

- **Proxy upstream timeout raised from 60s → 5m to survive Ollama cold-starts.**
  On the mini (Apple M4 Pro / 64 GB) loading qwen3.6:27b cold takes ≈ 80s and
  gemma4:26b ≈ 6m for the first request after the model unloads. Every
  minute-cron caller (machine-status push, voice_assistant's OpenClaw flow,
  etc.) was disconnecting after 60s, which canceled the in-flight ollama
  request server-side; ollama then dropped the half-loaded model, the next
  call started cold again, and the loop reproduced indefinitely — surfacing
  on `:56240/machines` as `"LLM 프로바이더 다운: 모든 서비스 실패: Ollama 연결
  실패…"` from speaker=mini. With 5m the cold-start completes, keep_alive=3m
  keeps the model resident, and subsequent calls return in 1-3s.

---

## [0.2.33] — 2026-04-26

### Fixed

- **Vault now shuts down cleanly on SIGTERM.** `wall-vault start` previously
  held no reference to the vault HTTP server and never stopped its background
  `startDailyReset` / `startStatusTicker` goroutines, so `systemctl stop` (or
  Ctrl-C in `start` mode) tore the process down mid-write of `vault.json`.
  `runAll` now wraps each handler in `*http.Server`, calls `Shutdown(ctx)` on
  both before exit, and a new `vault.Server.Stop()` closes a `stopCh` that the
  ticker goroutines now select on. `proxySrv.Stop()` already existed but was
  reached after the vault routine had been killed; both stops now run in the
  shutdown ordering the rest of the codebase already expected.

- **Broker no longer panics when a slow SSE subscriber disconnects mid-broadcast.**
  `broker.Broadcast` snapshotted the channel slice under `RLock`, released the
  lock, and then iterated. If `Unsubscribe` ran during that gap it could close a
  channel that `Broadcast` was about to write to, panicking the vault on
  send-on-closed-channel. The send loop now stays inside `RLock`; each
  iteration is non-blocking thanks to the existing `default:` drop branch, so
  there is no deadlock concern.

- **`clientAuth` no longer accepts arbitrary tokens when admin_token is unset.**
  The dev-mode short-circuit was an `OR` clause that fired before the client
  token lookup, so `Authorization: Bearer literally-anything` passed through
  on `/api/keys`, `/api/heartbeat`, etc. The handler now mirrors the
  `adminAuth` / `sseAuth` pattern (open mode = early return, no header
  inspection), so the only behaviour change is that bogus tokens are no longer
  rewarded with success when running without an admin_token.

- **`vault.json` writes are now durable across crashes.** The atomic
  write-then-rename pattern relied on `os.WriteFile`, which on Linux leaves the
  bytes buffered in the page cache. A power loss between WriteFile and the
  next fsync corrupted `vault.json`. `Store.save` now `O_TRUNC|O_CREATE`s the
  tmp file, `f.Sync()`s before close, renames, and best-effort fsyncs the
  parent directory so the rename's metadata is also durable.

- **Hot-path saves surface their errors.** `RecordKeyUsage`,
  `SetKeyUsage`, and `SetKeyCooldownIfLater` discarded the `s.save()` error
  with `_ =`, so a full disk during heartbeat handling silently lost usage
  counters. Each callsite now logs the failure with the key id.

- **Migration of legacy keys is no longer silent.** `migrateLegacyKeys`
  `continue`d past keys whose legacy decrypt failed (typically wrong master
  password) without logging — operators had no way to notice that a
  master_password change had stranded entries on the SHA-256 scheme. Both the
  per-key reason and a tail summary line are now logged.

- **Internal vault errors no longer leak through the API.** Six handlers used
  `jsonError(w, err.Error(), 500)` which echoed messages like
  `"duplicate ID: xyz"` straight to the caller, exposing storage internals.
  A new `jsonInternalError(w, where, err)` helper logs the full error
  server-side and returns a generic `"internal error"` to the client. All 500
  paths in `handleAdmin{Clients,ClientsReorder,ServicesReorder,Keys,Services{,ID}}`
  now route through it.

- **Service plugin loader logs every skip.** `LoadPlugins` swallowed file-read
  errors, YAML parse errors, and missing-id entries with bare `continue`s, so
  a typo in `~/.wall-vault/services/*.yaml` made the corresponding service
  silently disappear. Each skip path now writes a `[plugins] skip <file>:
  <reason>` line.

- **Proxy → vault key sync no longer accepts a partial body.**
  `keymgr.SyncFromVault` ignored the `io.ReadAll` error, which meant a
  truncated response could yield an empty `keys` slice that subsequent
  Unmarshal happily parsed as "no keys" — with no diagnostic for the operator.
  The error is now propagated.

- **`streamOllama` no longer hardcodes `qwen3.5:35b`.** Falling back to a
  hardcoded model name produced a 404 on any Ollama daemon that didn't have
  that exact tag pulled. The fallback now reads `s.model`, and if that is also
  empty the handler emits a clear SSE error frame instead of silently calling
  the wrong model.

- **`parseProviderModel` is now bounded against pathological `custom/` chains.**
  The recursive arm at the `case "custom":` branch had no depth limit, so a
  request body containing `model: "custom/custom/.../foo"` could theoretically
  unwind through Go's stack. The recursion now caps at 8 levels via an
  internal `parseProviderModelDepth` helper.

- **Dashboard delete failure alert is i18n-aware.** The DELETE button handler
  in `base.templ` hardcoded `'삭제 실패: '` regardless of the user's selected
  language; the matching `js_delete_fail_fmt` key was already present in
  `jsI18nKeys` but never referenced. The handler now uses `wvFmt(WV_I18N.
  js_delete_fail_fmt || 'Delete failed: {err}', {err: t})`, matching how
  save/reorder failures are already reported.

- **Unknown theme names log a warning instead of failing silently.**
  `theme.Get` quietly fell back to `cherry` whenever an unknown name arrived,
  which made config-file typos hard to diagnose. The fallback now logs once
  per call.

- **OpenAI multipart messages no longer carry the Korean placeholder
  `"[이미지]"`.** The `image_url` handler in `models.go` dropped a localized
  string into request bodies regardless of locale; the placeholder is now the
  locale-neutral `"[image]"`.

---

## [0.2.32] — 2026-04-26

### Fixed

- **Korean dashboard no longer leaks English `<optgroup>` labels into the agent
  edit slideover.** `I18nJSONBlob()` was reading the JS-facing strings via
  `i18n.T()`, which itself read a process-global `lang` variable mutated by
  every dashboard request through `SetLang`. With multiple users (or a single
  user toggling languages) the global could be "en" at the moment a Korean
  page was being rendered, so `WV_I18N.lbl_group_default` and friends shipped
  empty and the front-end JS fell through to its `'Default' / 'Allowed' /
  '(use service default)'` literals — producing English entries inside an
  otherwise-Korean model_override `<select>`. Adds `i18n.TFor(lang, key)`,
  a stateless variant of `T`, and threads the per-request locale from
  `layouts.Base(..., lang, ...)` into `I18nJSONBlob(lang)` so the JSON blob
  always reflects the locale the surrounding HTML rendered in. `T()` retains
  its old signature for back-compat (it now delegates to `TFor(lang, key)`).

---

## [0.2.31] — 2026-04-26

### Fixed

- **Pin Ollama's per-request `think` switch to stop hidden-reasoning blow-ups.**
  Thinking-capable models on the fleet's shared Ollama (qwen3.5/3.6, gemma3+,
  deepseek-r1, …) default to `think=true`. The proxy never overrode this, so
  every chat completion silently entered reasoning mode: the model burned the
  full `num_predict` budget on hidden thought, returned `content=""` with
  `finish_reason="length"`, and held the model resident in wired memory.
  Across four fleet machines hammering one mini Ollama daemon this accumulated
  to ~42 GB wired and a hung daemon within minutes — the symptom the operator
  described as "프록시 안 쓰고 미니 자체에서 큰 모델도 잘만 된다", because
  `ollama run --think=false` from the CLI bypassed the same setting.
  `OllamaRequest` and `OpenAIRequest` now expose a `Think *bool` field, and
  `callOllama` / `streamOllama` / `callLocalService` pin it to the same value
  as `Reasoning` (vault `reasoning_mode=true` → `think=true`, default → `false`).
  Pointer type so the field is only emitted when explicitly set, leaving
  Ollama's behaviour untouched on builds that haven't synced from vault yet.

---

## [0.2.30] — 2026-04-26

### Fixed

- **Token-less callers now inherit the proxy's own `model_override`.**
  v0.2.29 made token-less calls pick up the proxy's own `fallback_services`
  but the model itself still came from the request body — so an operator
  who switched host-B to OpenRouter via `vault.model_override` was still
  served whatever local OpenClaw / Claude Code's primary model said.
  Surfaced when host-B's OpenClaw kept emitting `custom/gemma4:26b`,
  driving every chat to the (chronically stuck) mini Ollama instead of
  the freshly-funded OpenRouter. v0.2.30 mirrors `client.ModelOverride`
  into a new `s.ownModelOverride` field on every `syncFromVault` and
  applies it to token-less calls before request-body model is consulted.
  Vault's `/api/clients` now exposes `model_override` separately from
  the legacy `default_model` field so the proxy can distinguish operator
  enforcement (override) from the legacy fallback. Token-resolved calls
  unchanged — they continue to follow the v0.2.27 priority
  `vault override > body > proxy default`. The fix means vault is now
  the single source of truth for routing on every host.

---

## [0.2.29] — 2026-04-26

### Fixed

- **Token-less callers now inherit the proxy's own fallback chain.** Local
  OpenClaw / Claude Code processes call `localhost:56244` without an
  Authorization header, and the proxy's own `VAULT_TOKEN` is explicitly
  excluded from `lookupTokenConfig`. The result, after v0.2.27/v0.2.28's
  strict-by-default and bare-name routing changes shipped, was that
  every fleet machine's local agent would 502 the moment its inferred
  service hit a cooldown — `claude-opus-4-7` → anthropic → all keys
  cooldown → 502 with no chain to fall back through. v0.2.29 loads the
  proxy's own client `fallback_services` from `/api/clients` into a new
  `s.ownFallback` field on every `syncFromVault` and uses it when the
  request carries no token (or its token resolves to nil). Vault's
  `/api/clients` now serialises `fallback_services` for authenticated
  callers (the proxies). v0.2.27's strict-by-default policy is preserved
  for clients whose vault config explicitly leaves `fallback_services`
  empty.

---

## [0.2.28] — 2026-04-25

### Fixed

- **Bare-model-name routing — `gemini-2.5-flash` no longer goes to ollama.**
  `parseProviderModel` previously short-circuited on any model name without a
  `provider/` prefix, returning the caller's preferred service unchanged. A
  request like `{"model": "gemini-2.5-flash"}` addressed to a client whose
  `preferred_service` is `ollama` was therefore forwarded to ollama as-is,
  producing a 404 from ollama (it does not host google models) and a noisy
  cascade of downstream errors visible in voice_api logs as
  `entity.parse.failed` (<separate incident>, "mini 카드 로그 진단 — LLM/TTS 흐름").

  v0.2.28 adds `inferServiceFromBareModel` and consults it whenever the model
  name has no `/`. Mapping rules:

    | bare name pattern               | service     |
    |---------------------------------|-------------|
    | contains `:` (e.g. `qwen:27b`) | `ollama`    |
    | `claude-*`                      | `anthropic` |
    | `gemini-*`, `gemma-*`           | `google`    |
    | `gpt-*`, `o1*`, `o3*`, `o4*`    | `openai`    |
    | anything else                   | (caller's choice stands) |

  When the inferred service differs from the caller's preferred service the
  inferred one wins — that's the whole point. Ambiguous or unknown bare names
  (`qwen3.5-32b` without a colon, `deepseek-r1` bare, custom internal names)
  leave the caller's preferred service untouched, preserving the v0.2.27
  priority order (`vault model_override > request body > proxy default`).

  Verified by `internal/proxy/parse_provider_test.go` covering 19 bare-model
  inference cases and 9 end-to-end `parseProviderModel` cases.

---

## [0.2.27] — 2026-04-25

### Changed

- **Request-body model is now respected when vault has no model override.**
  Pre-v0.2.27 priority was `proxy.s.model > request body > nothing`, so
  a token-auth'd client whose `model_override` was empty had its
  request-body model silently replaced by the proxy's own default
  (e.g. an econoworld request for `qwen3.6:27b` routed through
  host-C's `host-D` proxy was rewritten to `anthropic/claude-opus-4-7`).
  New priority: `vault token override > request body > proxy default`.
  Vault override is still final — operators can still pin a model on a
  client by setting its `model_override`. Affects both `/v1/chat/completions`
  (handleOpenAI) and `/v1/messages` (handleAnthropic).
- **Model routing redesign — strict-by-default fallback + visible
  routing decisions.** v0.2.21–v0.2.26 worked through a series of
  patches (host-based attribution, per-type liveness, env-priority
  ollama URL) that incrementally exposed how silent the dispatch chain
  was: a request for `qwen3.6:27b` could quietly return
  `google/gemini-3.1-flash-lite-preview` because the previous dispatch
  fell over to the next service with that service's `default_model`,
  with no signal to the caller other than the response body's `model`
  field. v0.2.27 splits the decision:

  - `Client.FallbackServices` (new) — ordered list of services to try
    if the primary fails, **separate from** the existing
    `AllowedServices` (security whitelist). Empty by default.
  - Empty `FallbackServices` = strict primary-only. Primary failure
    returns 502 with the primary error preserved as the headline; no
    silent model substitution.
  - When `FallbackServices` is set and dispatch falls over, the
    response carries three new headers so the caller never has to
    reverse-engineer a substitution:
      `X-WV-Used-Service`, `X-WV-Used-Model`, `X-WV-Fallback-Reason`.
  - The implicit "fall back through `proxy.services` order" path that
    used `s.allowedServices` / `s.cfg.Proxy.Services` as the default
    fallback chain is removed. Operators wanting fallback now opt in
    explicitly per client via the dashboard's new "Fallback service
    chain" field (i18n: `f_fallback_services`, `ph_fallback_services`,
    `hint_fallback_services` across all 17 locales).
  - `dispatchWithChain` is the new core; the old `dispatch(ctx, svc,
    model, req)` is kept as a convenience wrapper that forwards `nil`
    for the chain (= strict primary-only).

  Surfaced by earlier reviewer feedback — an econoworld-token call to
  `qwen3.6:27b` returned `google/gemini-3.1-flash-lite-preview`.

- **Restored per-call `AgentOffset + FallbackJitter` on local-inference
  paths.** v0.2.23 removed the v0.2.21 entry-distribution logic citing
  "0 queue overflow events in 24h"; v0.2.26 revealed that data point
  was an artifact of a separate `ollamaURL()` priority bug — the
  fleet's traffic was reaching `127.0.0.1:11434` and bouncing on
  connection-refused before ever queueing. With v0.2.26's URL fix,
  four proxies actually fan in to mini's Ollama and the host hung in
  practice. v0.2.27 puts the deterministic `AgentOffset(client_id,
  500ms)` + uniform `FallbackJitter(0–200ms)` back at all three
  acquire sites (`callOllama`, `streamOllama`, `callLocalService`).

### Added

- **Agent card "default model" indicator.** When `ModelOverride` is
  empty the card now renders the configured service's `default_model`
  alongside a muted "default" chip (i18n: `chip_default_model`,
  `tip_default_model` across all 17 locales). Previously the model
  column would simply be blank, leaving operators unsure whether the
  agent was misconfigured or correctly using the service default.

---

## [0.2.26] — 2026-04-25

### Fixed

- **`ollamaURL()` priority — per-machine env var now wins over vault
  serviceURLs.** Each fleet host knows its own topology (where the
  GPU/Ollama is reachable) better than a single global vault default.
  In a multi-machine fleet only one box runs Ollama and the other
  proxies must reach it remotely; the previous priority (vault > env)
  ignored the systemd `Environment=WV_OLLAMA_URL=...` override, so
  even when the operator pinned the correct address per-host the proxy
  still used the vault-provided `127.0.0.1`, producing `connection
  refused` and a silent fallback to the cloud chain. New order:
  `WV_OLLAMA_URL` env > `OLLAMA_URL` env > vault `serviceURLs` >
  `http://localhost:11434`. Surfaced by <separate incident>: an
  econoworld-token call to `qwen3.6:27b` was returning
  `google/gemini-3.1-flash-lite-preview` because host-C's proxy
  could not reach its env-pinned Ollama.

---

## [0.2.25] — 2026-04-25

### Changed

- **Split `detectClientAlive` policy by agent type.** v0.2.24 ran the
  same pgrep gate on every agent type, which fixed cline's false-green
  but introduced a false-red for claude-code clients that simply hadn't
  been invoked in the last few minutes (operator complaint:
  "babi2 잘 하고 있는데 왜 빨간불"). The two cases differ in nature —
  Cline only runs while VSCode is open, while a claude-code client is
  a fleet member that gets typed into intermittently and uses
  Anthropic OAuth bypassing this proxy. Updated probe table:

  | agent_type | probe |
  |---|---|
  | claude-code | always true (trust Host — operator-assigned membership is the source of truth) |
  | cline | `pgrep -x code` (VSCode must be open for the extension to be alive) |
  | openclaw | `pgrep -f openclaw-gateway` |
  | nanoclaw | `systemctl --user is-active nanoclaw` |
  | econoworld | always false (self-reports separately) |
  | other | false (don't fake green for unknown types) |

  Net effect: `claude` / `babi` / `babi2` / `macclaude` / `saweol`
  light up as long as their host is heartbeating; `motoko_vsc` stays
  red while VSCode is closed; openclaw/nanoclaw signals are unchanged.

---

## [0.2.24] — 2026-04-25

### Fixed

- **Signal light no longer shows green for an agent whose process isn't
  running.** v0.2.22 simplified the dashboard liveness model to "any
  vault Client whose Host field equals the reporting proxy's
  `os.Hostname()` is broadcast as active in every heartbeat" — a clean
  rule that incidentally meant a host-matched cline (or any other
  agent type) stayed permanently green even after VSCode was closed,
  because the host was still up. Heartbeat now runs `detectClientAlive`
  on every host-matched client before emitting it, so the operator-
  assigned Host field decides *who may be claimed by this proxy* and
  the per-agent-type process probe decides *who is actually up right
  now*. Initial probe table:

  | agent_type | probe |
  |---|---|
  | claude-code | `pgrep -x claude` |
  | cline | `pgrep -x code` (VSCode binary) |
  | openclaw | `pgrep -f openclaw-gateway` |
  | nanoclaw | `systemctl --user is-active nanoclaw` |
  | econoworld | always false (self-reports via its own heartbeat) |
  | other | false (don't fake green for unknown types) |

  Multiple claude-code clients sharing one Host all match the same
  pgrep, so a single running CLI still lights all of them up — fully
  disambiguating WSL-side from Windows-side claude-code on the same
  physical box would need a cwd-based or OS-interop probe and is left
  for a future iteration.

---

## [0.2.23] — 2026-04-25

### Changed

- **Drop per-call AgentOffset + FallbackJitter on local-inference paths.**
  v0.2.21 added a 500 ms deterministic + 200 ms random pre-acquire delay
  (~350 ms avg) to spread fan-in across a fleet of proxies sharing one
  Ollama host. With the current operating point (4 proxies, single
  model, ~1000 daily calls, zero queue-overflow events observed in 24h)
  the defence carries measurable cost (latency on every chat completion
  and stream first-token) without measurable benefit. Per-proxy cap-1
  semaphore is kept as the cheap GPU-memory serialiser, and the
  one-shot phase-shift on heartbeat / vault-sync tickers stays — those
  run once at boot and the cost is invisible. Removed at three sites:
  `callOllama`, `streamOllama`, `callLocalService`. The
  `localAgentOffsetMs` / `localFallbackJitterMs` constants are gone;
  the `AgentOffset` and `FallbackJitter` helpers themselves remain
  (used by heartbeat / vault-sync boot phase + test coverage).

  Re-introduce the per-call offset/jitter when any of the following
  is observed:
    - fleet grows to 6+ proxies sharing the same upstream local backend
    - /admin/proxies dashboard logs 5+ Ollama 503 / queue-overflow
      events in a rolling 24h window
    - a multi-model upgrade (concurrent loads on the Ollama host)
      reintroduces the cross-proxy thundering-herd pattern

---

## [0.2.22] — 2026-04-24

### Fixed

- **Signal light attribution across a multi-agent host**:
  `proxy.syncFromVault` selected the claude-code client to report
  activity for by iterating the vault's client list and letting the
  last match win, so every proxy pinned the same id (whichever sorted
  last) and every other claude-code client's signal light was
  permanently offline even while its host was active. Additionally,
  only a single claude-code client was ever emitted per heartbeat, so
  a host that legitimately runs multiple claude-code agents (WSL +
  Windows, multiple profiles, or a mix of claude-code / cline /
  openclaw in one box) could not light all of them up at once. The
  new scheme adds a `Host` field to the vault `Client` record plus a
  `Hostname` input in the admin slideover (i18n keys `f_host`,
  `ph_host`, `hint_host` across all 17 locales); each proxy then
  caches the full set of clients whose Host equals `os.Hostname()`
  and emits every one of them in each heartbeat's `ActiveClients`
  list, so N co-hosted agents produce N green lights. An explicit
  `proxy.claude_code_client_id` config (env `WV_CC_CLIENT_ID`)
  overrides the match for hosts where `os.Hostname()` is unreliable
  (renamed boxes, WSL). The previous per-agent pgrep liveness probe
  is dropped in favour of the operator-assigned Host field, because
  liveness detection isn't uniform across agent types (VSCode
  extensions have no binary to pgrep; Windows-side claude-code is
  invisible to WSL pgrep).

---

## [0.2.21] — 2026-04-24

### Added

- **Per-agent time distribution across the fleet**: multi-proxy
  deployments that share an upstream local inference host and a shared
  vault exhibited two collision patterns — several proxies failing
  over to the same local backend at the same instant after an upstream
  4xx and driving the server queue toward the host's per-model queue
  limits, and several proxies booted within the same second running
  their heartbeat / vault-sync tickers on the same wall-clock edge.
  A new `internal/proxy/timing.go` introduces two helpers: `AgentOffset`
  returns a deterministic delay derived from `sha256(client_id)`, so
  two proxies in the same fleet land on different phase positions;
  `FallbackJitter` returns a uniform `crypto/rand` delay that smooths
  residual hash-bucket collisions. The offset+jitter pair (~700 ms
  worst case, ctx-cancellable) is now applied at every local-inference
  call entry (`callOllama`, `streamOllama`, `callLocalService`); the
  deterministic offset alone (bounded by each goroutine's period) is
  applied at the initial tick of the two periodic goroutines
  (`startHeartbeat` and the vault-sync loop in `NewServer`).
  Standalone deployments with an empty `client_id` see no offset
  change; the jitter component remains small enough to be cosmetic
  there.

### Changed

- **`Server.ollamaSem` replaced with `Server.localSems` map**: the
  previous single Ollama-only semaphore is now a cap-1 channel per
  local inference backend (`ollama`, `llamacpp`, `lmstudio`, `vllm`),
  initialised at construction time. `callOllama` and `streamOllama`
  acquire `s.localSems["ollama"]`; `callLocalService` — which
  previously had no concurrency control at all — now acquires the
  service's own slot. A fleet that migrates primary off Ollama to any
  other local backend inherits the same proxy-side serialisation
  Ollama had.

---

## [0.2.20] — 2026-04-24

### Fixed

- **`streamOllama` now honours caller context, uses a 10-minute client
  timeout, and acquires the concurrency semaphore in a cancellable
  select** (`internal/proxy/stream.go`). Previously it built requests
  with `http.NewRequest` (no context), took the `ollamaSem` slot via a
  bare `<-` send with no way to unwind on client disconnect, and ran
  under `cfg.Proxy.Timeout` (default 60s) — too tight once
  `OLLAMA_KEEP_ALIVE` is tuned down, since the first post-idle call has
  to cover a cold model reload that can take tens of seconds for large
  (>8 GB) models. The fix threads `r.Context()` from
  `handleGeminiStream` into `streamOllama`, swaps to
  `NewRequestWithContext`, wraps the semaphore acquire in a `select`
  that aborts on `ctx.Done()`, and bumps the HTTP client timeout to
  match `callOllama`'s 10-minute budget. The practical path
  (`/v1/chat/completions` → non-streaming `callOllama`) already behaves
  correctly; this aligns the Gemini streaming path
  (`/v1beta/.../streamGenerateContent`) with the same guarantees.

---

## [0.2.19] — 2026-04-24

### Fixed

- **`/api/clients` now advertises v0.2 canonical routing fields to
  proxies**: the response still uses the `default_service` /
  `default_model` wire names (backwards-compat unchanged), but their
  value is now the **effective** one — `PreferredService` if set,
  falling back to the legacy `DefaultService`; same pattern for
  `ModelOverride` vs `DefaultModel`. Before this fix a client whose
  dashboard `preferred_service` had been changed to `ollama` would
  still ship its old legacy `default_service=openrouter` to every
  proxy's `syncFromVault`, yielding `[sync] 설정 로드: openrouter/`
  and a silent service–model mismatch. EconoWorld (earlier reviewer) and other
  hosts whose proxies had migrated their config to canonical fields
  were trapped in the 6–10 minute fallback timeout loop this caused.
- **`ollamaURL()` prefers vault-synced URL over environment / default**:
  the resolution order changed from `env > vault sync > localhost`
  to **`vault sync > env > localhost`**. A proxy that started before
  its `syncAllowedServices` had finished would previously fall through
  straight to `http://localhost:11434` on hosts with no local Ollama,
  producing `dial tcp 127.0.0.1:11434: connect: connection refused`
  even though the vault had published the correct fleet URL
  (`http://<internal-host>:11434`). Env vars still work as explicit
  overrides when no `ollama` service entry is registered in vault.

---

## [0.2.18] — 2026-04-23

### Changed

- **Uptime ticker + SSE indicator promoted from footer to topbar**:
  `HeaderVM` gained a `StartedAt` field and the header template now
  renders a `topbar-meta` cluster holding the `⏱ <uptime>` ticker and
  the `● SSE` dot. `FooterVM` dropped `StartedAt` and the two indicator
  spans — the footer is now identity/attribution only (version +
  github / domain / email links). `theme.templ` gained
  `.topbar-meta`/`.topbar-uptime`/`.topbar-sse` CSS rules. Existing JS
  driving `#wv-uptime` and `#wv-sse-dot` works unchanged (DOM IDs
  preserved), so operators see connection health without scrolling to
  the footer.
- **Dispatch fallback chain reordered by reliability**: `dispatch()`
  previously iterated `allowedServices` in dashboard sort order, which
  meant a primary failure would spend minutes waiting on metered
  clouds (openrouter → anthropic → openai) whose keys might all be
  exhausted before finally reaching a local backend. A new
  `fallbackPriority` constant pins the order to
  `ollama → llamacpp → lmstudio → vllm → google → github-copilot →
  openrouter → anthropic → openai`. Primary (`preferred_service`) is
  still tried first; custom services absent from the priority list
  keep their dashboard order at the tail.
- **Anthropic passthrough honours caller OAuth / API-key (BYO auth)**:
  `handleAnthropic` now detects `sk-ant-*` tokens in `Authorization:
  Bearer` or `x-api-key` and forwards them to Anthropic verbatim,
  bypassing the vault-key rotation / cooldown / token-accounting path.
  Single-shot call, no retry, no vault-key side effects. Lets upstream
  OAuth sessions reach Anthropic even when the proxy's own vault key
  is exhausted.
- **SSE bridge for buffered Anthropic responses**: handlers used to
  return a single JSON blob even for `stream: true` requests, which
  made the Claude Code SDK hang waiting for the first SSE chunk. A new
  `WriteAnthropicSSEFromJSON` splits the buffered response into the
  spec'd event sequence (`message_start` → `content_block_start` /
  `content_block_delta` / `content_block_stop` per block →
  `message_delta` → `message_stop`), flushes between events, and
  preserves the upstream Content-Type. Used by both the passthrough
  and dispatch paths whenever the original request had
  `stream: true`.

### Fixed

- **Network / context-cancel no longer parks keys in cooldown**:
  `cooldownDurations[0]` was falling through to the default 5-minute
  bucket, so a client disconnect (errCode=0) mid-request was treated
  as a key fault and every key on the service cycled into cooldown.
  `cooldownDurations[0] = 0` now explicitly marks transport-layer
  errors as non-key faults — the request fails, the key stays
  available, the next retry proceeds normally.

### Security

- **Pre-release leak sweep**: grep pattern across all tracked files
  for real fleet hostnames / LAN IPs / persistent tokens — zero hits
  after prior `v0.2.17` scrub remains zero after the v0.2.18 diffs.
  Only the intentionally-public owner email + github/domain links
  remain in `footer.templ`.

---

## [0.2.17] — 2026-04-19

Full-surface audit rollup: rounds A (security), B (reliability), C (UX /
i18n / observability), and D (hardening). Single release so existing
deployments can move off v0.2.16 in one step.

### Added

- **`llama.cpp` as a local service**: new `llamacpp` entry in the default
  service list and dispatch switch. Shares `callLocalService` with
  `lmstudio`/`vllm` (OpenAI-compatible `/v1/chat/completions`). Dashboard
  treats it as a local service (no key required). Users configure
  `local_url` and `default_model` via the service edit slideover.
- **Reasoning mode toggle for local services** (`ServiceConfig.ReasoningMode`):
  new per-service checkbox shown in the edit slideover for Ollama /
  lmstudio / vLLM / llama.cpp. When enabled, the proxy sets
  `OpenAIRequest.Reasoning = true` before marshalling, so forwarded
  chat-completions bodies include `"reasoning": true`. Servers that
  understand the flag emit thinking/chain-of-thought output; others
  ignore the unknown field. The toggle is synced from vault to proxies
  via `/api/services` (now returns `reasoning_mode` alongside
  `local_url` / `default_model`) and stored in `serviceReasoning` on
  the proxy side. New i18n keys `f_reasoning_mode` /
  `hint_reasoning_mode` across all 17 locales.
- **Service edit dropdown falls back to on-demand model fetch**: opening
  the edit slideover for a *disabled* local service (e.g. a freshly
  added llama.cpp) used to show an empty `default_model` dropdown
  because `ensureRegistry` skipped disabled services. A new
  `Registry.RefreshService(svcID, localURL, orKey)` method fetches
  models for a single service and upserts them into the cache; the
  slideover renderer now triggers it whenever the registry returns
  zero entries for the service being edited.

### Changed

- **SSE broker channel size 8 → 64** and **drop counter**: per-subscriber
  buffer was too tight for the 15s `agents_sync` cadence plus
  `config_change` bursts — slow tabs silently dropped events. Now holds
  ~1 min of peak traffic, and `/api/status` (authenticated) exposes
  `sse_dropped` so operators can spot slow clients.
- **SSEClient reconnect throttle**: proxy's SSE client used to spawn a
  fresh `onReconnect` goroutine on every reconnect. A flapping vault
  could pile up concurrent sync calls. Now guarded by an `atomic.Bool`
  so only one sync runs at a time; further reconnects log a skip line.
- **Daily-usage date comparison unified to UTC** (`internal/proxy/keymgr.go`,
  `internal/vault/store.go`, `internal/vault/server.go`): proxies and the
  vault were both formatting `time.Now().Format("2006-01-02")` against
  local time. Different time zones or clock drift could flag fresh usage
  as stale. All three sites now use `time.Now().UTC().Format(...)`.
- **Token cache periodic eviction** (`internal/proxy/server.go`): the
  proxy's in-memory token→config cache previously only evicted when it
  crossed 500 entries. A 30-second background ticker now trims expired
  rows, and the inline safety-valve cap was tightened to 200.
- **Config `Validate()` on load**: `config.Load` now rejects an invalid
  `mode` / out-of-range ports / non-positive `proxy.timeout` / unknown
  `tool_filter` / empty `proxy.services` instead of silently starting a
  service that will misbehave at runtime.
- **Dispatch returns actual used service/model**: `dispatch` now returns
  a `DispatchResult{Response, UsedService, UsedModel}`. Handlers
  (OpenAI, Anthropic) populate the response `model` field with the
  actual backend used, so a fallback from `claude-sonnet-4-6` to
  `google/gemini-flash` is no longer misleadingly labeled as the
  originally-requested model.
- **HTTP context propagation through dispatch chain**: `handleOpenAI` /
  `handleGemini` / `handleAnthropic` now forward `r.Context()` through
  `dispatch` → `callGoogle` / `callOpenRouter` / `callOpenAI` /
  `callAnthropic` / `callOllama` / `callLocalService` /
  `callAnthropicPassthrough` / `doRequest` / `doAnthropicRequest`. All
  upstream HTTP calls use `NewRequestWithContext`, so a client
  disconnect cancels the outbound request instead of letting it run to
  completion and leak a socket.
- **Ollama concurrency: mutex → context-aware semaphore**: the plain
  `sync.Mutex` guarding Ollama requests is replaced by a buffered
  channel (`ollamaSem`, capacity 1). `callOllama` acquires via `select`
  with `ctx.Done()`, so a caller whose request is cancelled while
  another is mid-inference no longer holds a slot for up to 10 minutes.
  `streamOllama` keeps the same blocking behavior (no ctx plumbed yet).
- **Tool-call argument parse failures now log**: `convert.go` used to
  silently replace unparseable `tool_call.arguments` with an empty map,
  hiding misbehaving agents. Both OpenAI `tool_calls` and Anthropic
  `tool_result` paths now emit a log line when they fall back, naming
  the function so traces point at the offending caller.
- **OpenClaw tmux model injection fails safely**: `sanitizeModelForTmux`
  now returns a `dropped` flag. `injectModelToTUI` logs when non-ASCII
  runes are stripped and aborts when the sanitized value is empty, so a
  Korean-only alias no longer becomes a bare `/model` command that
  confuses the OpenClaw TUI.
- **CORS private-IP coverage**: `isAllowedOrigin` now accepts any
  RFC1918 range (`10/8`, `172.16/12`, `192.168/16`) plus loopback and
  link-local, via `net.IP.IsPrivate()` / `IsLoopback()` /
  `IsLinkLocalUnicast()`. Corporate `10.x` networks no longer hit CORS
  errors when hosting wall-vault. Also handles `[::1]:port` / IPv6
  origins through `net.SplitHostPort`.
- **Doctor degraded-state detection**: a 200 response is no longer
  auto-classified "정상". The body is now validated as JSON with a
  `status` field; otherwise Doctor surfaces "응답 형식 오류" so a
  wrong-binary or crashed-handler port that still returns 200 on
  `/health` doesn't mask a broken service.
- **Proxy graceful-shutdown signal**: added `Server.stopCh` + `Stop()`.
  Background goroutines (initial vault load, 5-min key sync, 30-s
  token-cache evict) now exit on `select { case <-stopCh: return }`
  instead of leaking past the owner's Stop call.
- **writeJSON parent-dir TOCTOU guard**: `agent_apply.writeJSON` now
  `Stat`s the target parent directory before writing, wraps errors with
  the path involved (`write %s`, `rename %s → %s`), and surfaces a
  clear `parent missing` message when the agent directory was removed
  between discovery and write.
- **Slideover accessibility**: `Frame` now renders
  `role="dialog" aria-modal="true" aria-labelledby="wv-slideover-title"`
  with the title element carrying the matching id; the close button
  has an explicit `aria-label`. Sidebar nav gets `aria-label`. Empty
  slideover state gets `aria-hidden="true"`.
- **Default UI theme → light**: `config.Default().Theme` is now
  `"light"` so fresh installs start with the most accessible palette.
  Existing installs are unaffected — `Settings.Theme` from vault.json
  continues to take precedence via `handleAdminTheme`'s persistence
  path, so user selections stay put across restarts.
- **Docs: v0.2 canonical field resolution**: `.claude/docs/endpoints.md`
  gained "Client Model Resolution (v0.2)" (preferred_service/model_override
  vs legacy default_*; empty string = "use service default"; stale-override
  soft-clear) and "SSE Authentication" (`?ticket=`/Authorization/`?token=`
  deprecation) sections.
- **CI version stamp aligned with Makefile**: `.github/workflows/ci.yml`
  now derives `BASE_VERSION` from the Makefile and appends a UTC
  timestamp + short SHA, matching `make build` output instead of the
  unrelated `git describe --tags --always --dirty` that previously
  produced diverging version strings between local and CI artifacts.
- **i18n hardcoded UI strings externalized**: 20 new keys
  (`act_close`, `opt_service_default`, `warn_stale_override_*`,
  `lbl_group_*`, `tip_*`, `act_*`, `sum_*`, `ph_*`, `nav_label_sections`)
  replace raw Korean literals in `slideover.templ` / `client_edit.templ`
  / `client_create.templ` / `service_edit.templ` / `service_create.templ`
  / `agent_card.templ` / `key_card.templ` / `service_card.templ` /
  `shell.templ`. All **17 locale JSONs** received native translations
  (ko/en authored in-place; ar/de/es/fr/ha/hi/id/ja/mn/ne/pt/sw/th/zh/zu
  translated by localization pass). `{val}` / `{n}` placeholders and
  unicode symbols (`↓`, `✕`, `⚠`, `·`, `…`) are preserved across all
  languages. Remaining Korean in JS alerts / `confirm()` prompts is
  scheduled for a later pass that plumbs the i18n dictionary into the
  base template as a JS object.
- **slog introduced (console-friendly default)**: `main.initSlog`
  installs a `slog.TextHandler` on stderr honoring `WV_LOG_LEVEL`
  (debug|info|warn|error, default info). New code paths — starting with
  the `wall-vault start` banner — use `slog.Info` with structured
  key/value attrs, while existing `log.Printf` calls continue to work
  unchanged. `runAll` now calls `proxySrv.Stop()` on SIGINT/SIGTERM so
  the stopCh goroutine-shutdown wiring from round B actually fires.
- **Default UI theme → light**: `config.Default().Theme` is now
  `"light"` so fresh installs pick the most accessible palette.
  Existing installs retain their saved `Settings.Theme` — persistence
  was already wired end-to-end via `handleAdminTheme` / `store.SetTheme`
  / `handleDashboard`.
- **HTTP security headers**: new `middleware.SecurityHeaders` sets
  `X-Content-Type-Options`, `X-Frame-Options`, `Referrer-Policy`, and a
  conservative `Content-Security-Policy` (allows the unpkg htmx script
  + inline style/script needed by templ) on every response. Wired into
  both vault and proxy chains. TLS/HSTS is left to whatever reverse
  proxy terminates TLS.
- **Config file permission hardening**: `config.Load` now warns on
  stderr when the file it read is group/world readable (`000077`
  mask). `config.Save` enforces `chmod 0600` after write so a pre-
  existing loose-mode file gets tightened by any subsequent save.
- **Model registry capped at 2000 entries**: `models.Registry` gained
  a `maxSize` (default 2000) so a runaway upstream catalog can't blow
  memory across refreshes. Exceeding it logs a truncation note. Also
  registered the new `llamacpp` service in `Refresh` (defaulting to
  `http://localhost:8080`) and in the `compatFallback` list.
- **Proxy per-IP rate limit**: new `middleware.RateLimiter` (100 req/s
  per IP, burst 20) shields the proxy's public endpoints
  (`/v1/chat/completions`, `/v1/messages`, `/v1/models`, `/health`,
  `/status`) from accidental loops and scanners. Idle visitor entries
  are reaped every 5 min. No external dependency — small token-bucket
  implementation kept in-tree.
- **Healthcheck endpoints normalized**: both proxy and vault now serve
  `GET /health` returning `{status, readiness, version, …}`. Proxy
  also reports `sse_connected` so K8s / doctor probes can distinguish
  "running but not sync'd" from fully ready. Vault's endpoint is
  auth-free so probes don't need admin tokens.
- **Hooks shell command: stdout/stderr capture + tunable timeout**:
  `hooks.Manager.fire` now captures both streams via `bytes.Buffer`
  and logs a truncated tail + elapsed time on failure, instead of
  silently swallowing errors. Timeout defaults stay at 30s but can be
  overridden with `WV_HOOK_TIMEOUT` (e.g. `2m`) for slow backup or
  webhook hooks.
- **Deploy hardening guide**: new `.claude/docs/deploy-hardening.md`
  collects the systemd unit options (`KillSignal`, `RestartSec`,
  `ProtectSystem`, `PrivateTmp`, `NoNewPrivileges`,
  `MemoryDenyWriteExecute`, `IPAddressAllow`), the launchd plist
  counterparts, the **`launchctl kickstart -k` env-reload gotcha**
  documented from the 2026-04-18 incident, and the new `/health`
  probe recipes.
- **templ form layout extracted**: new `FormSection(title)` and
  `FormActions()` components in
  `internal/vault/views/slideover/form.templ` replace the repeated
  `<div class="wv-form-section"><div class="wv-form-section-title">…`
  / `<div class="wv-form-actions">…` plumbing across
  `client_create`, `client_edit`, `service_create`, `service_edit`,
  `key_create`. CSS untouched — same DOM, far cleaner templ sources
  so future form tweaks happen in one place.
- **JS alerts / SSE status / confirm dialogs externalized**: the
  remaining hardcoded Korean in `base.templ`'s inline JavaScript
  (`confirm('정말 삭제하시겠습니까?')`, `alert('순서 저장 실패: …')`,
  SSE status tooltip map, avatar preview alt, `(사용자 지정)`,
  `무료/유료` optgroup labels, etc.) now come from a new
  `window.WV_I18N` object. `layouts.I18nJSONBlob()` emits the 16-key
  subset as inline `<script type="application/json">`; a bootstrap
  IIFE parses it before any other script runs, and a tiny
  `wvFmt(tpl, vars)` helper handles `{err}` / `{val}` / `{n}`
  placeholder substitution. 12 new `js_*` keys authored in ko/en and
  translated into all 15 other locales (ar/de/es/fr/ha/hi/id/ja/mn/ne/
  pt/sw/th/zh/zu) respecting language-specific typography (French
  spaces before `:` and `?`, Chinese full-width parentheses, Arabic
  RTL wording, Hausa hooked-ɗ spelling, etc.).

### Added

- **Anthropic live model-list fetch + Claude Opus 4.7**: `fetchAnthropic`
  now tries `GET https://api.anthropic.com/v1/models` first (using the
  configured Anthropic API key from the vault) and only falls back to
  the hand-curated static list when no key is set or the endpoint is
  unreachable. The static list gained `claude-opus-4-7` so the newest
  Opus release shows up in the dashboard dropdown even before the live
  fetch lands. `Registry.Refresh` / `Registry.RefreshService` now take
  a `ServiceKeys map[string]string` so the vault can hand both the
  OpenRouter and Anthropic keys through the same channel (same pattern
  already used for `ServiceURLs`). Callers in `vault/server.go`,
  `vault/hx_router.go`, and `proxy/server.go` were updated; the
  proxy-side call passes `nil` because that path only cares about
  local URLs.

### Fixed

- **Anthropic passthrough honours caller-supplied Anthropic credentials**
  (Bring-Your-Own auth): `callAnthropicPassthrough` used to unconditionally
  overwrite the upstream auth with a vault `x-api-key`, so a Claude Code
  session whose OAuth token had been injected by an intermediary
  (e.g. NanoClaw's credential-proxy) never actually reached Anthropic
  with that token — every call was billed against the vault key instead.
  When that vault key ran out of credits the upstream returned HTTP 400
  "credit balance too low", wall-vault fell through openrouter → google
  → ollama → llamacpp (6–10 minutes per request), Claude Code's SDK
  retried, and after a handful of rounds hit its own 30-minute ceiling
  with no reply on the Telegram side. `handleAnthropic` now detects an
  Anthropic-formatted token in the request (`sk-ant-*` in either
  `Authorization: Bearer` or `x-api-key`) and forwards it upstream
  verbatim; the BYO branch does a single-shot call, skips vault-key
  rotation / cooldown / usage tracking, and returns the upstream
  Content-Type so streaming (`stream: true`) responses aren't mislabelled
  as `application/json`. Vault-key requests keep their existing retry +
  cooldown + token-accounting semantics untouched.
- **Multimodal content silently dropped on OpenAI-compat upstreams**:
  `OpenAIMessage.RawContent` was declared `json:"-"` so the multi-part
  array parked by `UnmarshalJSON` (text + image_url + input_audio + …)
  never made it back onto the wire — outbound `MarshalJSON` only emitted
  the flat `Content` string extracted for legacy consumers. Every
  chat-completions-style upstream (Ollama, lmstudio, vLLM, llama.cpp,
  OpenAI direct, OpenRouter) therefore received a text-only body even
  when the client sent images. `MarshalJSON` now re-emits `RawContent`
  verbatim as the `content` field whenever it's populated, falling back
  to the existing assistant empty-content guard otherwise. Gemini path
  was never affected because `openaiPartsToGemini` materialises parts
  into `InlineData` before marshal.
- **Anthropic `image` / `document` blocks dropped on non-Claude
  dispatch**: `anthropicToOpenAIReq`'s user-message loop handled only
  `text` and `tool_result`, so any Claude-format client sending an
  `image` (base64 or URL) or `document` source block had the block
  silently skipped before the request reached Gemini / Ollama / etc.
  New helpers `anthropicImageSourceToURL` and
  `anthropicDocumentSourceToPart` convert them to OpenAI `image_url`
  (data URI for base64, passthrough URL for http(s)) and `input_file`
  (for base64 PDFs) respectively, packed into `RawContent` so Fix 1
  carries them the rest of the way.
- **Gemini image responses collapsed to text on the Anthropic wire**:
  `GeminiRespToAnthropic` previously ran every candidate through
  `extractText`, discarding any `InlineData` that image-generation
  models (e.g. `gemini-3.1-flash-image-preview`) returned. It now walks
  each part in order, emitting `{type:"text", text:"…"}` blocks for
  text spans and `{type:"image", source:{type:"base64", media_type,
  data}}` blocks for `InlineData`. A new `AnthropicSource` struct plus
  an optional `Source` field on `AnthropicContent` carry the payload.
- **Gemini image responses invisible to OpenAI clients**: the response
  builder in `handleOpenAI` called `extractText`, so the binary blob
  disappeared without a trace. A new `extractTextAndMediaNotes`
  preserves the text while appending a one-line placeholder
  (`[media attached: <mime>, ~N bytes …]`) whenever `InlineData` is
  present, making the dropped blob at least observable on the
  string-only OpenAI `content` field.
- **Remote `image_url` fetch failures invisible**: `fetchAsBase64`
  returned `ok=false` on timeout / >5 MB body / non-2xx, which
  `openaiPartsToGemini` silently ignored — the request proceeded as if
  the image never existed. The caller now substitutes a
  `[image_url fetch failed: <url> …]` text part so the upstream model
  and downstream logs surface the drop.
- **Slideover i18n locale leak**: edit slideovers rendered in English even
  after the dashboard was set to Korean, because only `handleDashboard`
  called `i18n.SetLang` while `/hx/*` fragment endpoints inherited the
  global. Added a `langMiddleware` that resolves the active language
  (vault settings → config → current) on every request before any templ
  render. Mitigates the most visible symptom; a proper per-request i18n
  context is scheduled for a later round.

### Security

- **Anthropic tool filter**: `handleAnthropic` now runs the request
  through `ToolFilter.FilterAnthropic` (new) right after decode, so
  `tool_filter: strip_all` and `whitelist` modes are enforced for
  Claude-format clients. Previously only OpenAI and Gemini paths were
  filtered — Anthropic requests bypassed the policy.
- **One-shot SSE tickets**: new `POST /api/sse-ticket` exchanges an
  admin or client bearer token for a 24-byte random ticket (5 min TTL,
  single-use). The dashboard connects to `/api/events?ticket=…` so the
  admin token no longer appears in URLs, Referer headers, or access
  logs. Legacy `?token=` is still accepted but logs a deprecation line
  with the caller IP.
- **Admin IP whitelist**: new `vault.admin_ip_whitelist` setting —
  when populated, `adminAuth` and the admin-scope path of
  `handleProxyKeys` both reject callers whose IP is not in the list.
  Unset keeps the existing behavior.
- **Mode-aware host defaults**: `ProxyConfig.Host` /
  `VaultConfig.Host` are now empty by default; `applyHostDefaults`
  fills them with `127.0.0.1` in `standalone` mode and `0.0.0.0` in
  `distributed`. YAML and the new `WV_PROXY_HOST` / `WV_VAULT_HOST`
  env vars still win. Stops single-host installs from exposing the
  vault to the LAN unintentionally.
- **Fail-fast on entropy exhaustion**: `newID()` now uses
  `io.ReadFull(rand.Reader, …)` and panics on error instead of
  silently producing a partially-initialized hex ID.

---

## [0.2.16] — 2026-04-17

### Fixed

- **Agent card uptime**: showed "last heartbeat" (e.g. `3s ago`) with
  ⏱ icon, which was confused with uptime. Now displays actual uptime
  from `ProxyStatus.StartedAt` (e.g. `2d 3h`, `15h 22m`). Only
  shown for online agents.

### Added

- **Key card usage bar for unlimited keys**: the progress bar
  previously only appeared for keys with a `DailyLimit`. Unlimited
  keys now show a relative-usage bar based on the maximum usage
  across all unlimited keys (busiest key = 100%). Bar height bumped
  from 4px → 6px for visibility.

### Security

- **Redacted admin token from plan doc**: a plaintext vault admin
  token (`TOKEN=…`) was committed in the v0.2 implementation plan.
  Replaced with `<ADMIN_TOKEN>` placeholder.
- **Redacted concrete LAN IPs from committed examples**: replaced
  deployment-specific addresses with generic `<internal-host>` /
  `<internal-host>` across 17 locale JSON files, 2 test files, 17
  MANUAL translations, and the plan doc.
- **Redacted Telegram bot usernames**: replaced 3 real bot usernames
  in the plan doc with `@<bot>` placeholders.
- `git filter-repo` recommended to purge history (see below).

---

## [0.2.15] — 2026-04-17

### Fixed

- **Atomic config file writes prevent 0-byte corruption on deploy**:
  `writeJSON` (used by all agent config writers — OpenClaw, EconoWorld,
  Cline, Claude Code) now writes to a `.tmp` sidecar and `os.Rename`s
  into place. Previously `os.WriteFile` truncated the file first, so
  a `pkill -x wall-vault` during deploy could catch the goroutine
  between truncate and write, leaving a 0-byte file that crashed
  OpenClaw's config validator. Observed on a deployed host's
  `~/.openclaw/openclaw.json` (Apr 16 02:30). `updateOpenClawJSON`
  in `openclaw_sync.go` also switched from inline `os.WriteFile` to
  the shared `writeJSON` for the same protection.

---

## [0.2.14] — 2026-04-17

### Fixed

- **Agent card reflects saved `model_override` immediately after
  save** — the card previously showed `RemoteModel` (the proxy's
  last-reported model via heartbeat) with priority, so right after
  a dashboard save the card kept displaying the old value for up
  to the next heartbeat cycle (~20 s). Priority is now
  `ModelOverride` → `RemoteModel`, so the just-saved configured
  model appears instantly on reload. `RemoteModel` still shows when
  there is no explicit override (e.g. the agent is on the service
  default).

### Changed

- **Unified model chip color on agent cards** — both paths
  (`ModelOverride` and `RemoteModel`) now render as
  `chip chip-accent`. Previously only one was `chip-accent` and the
  other plain `chip`, which made some cards' model text a different
  color than others for no user-facing reason. `agent_type` chip
  stays plain to preserve the information-tier distinction.

---

## [0.2.13] — 2026-04-17

### Fixed

- **Agent edit `preferred_service` change no longer leaves stale
  `model_override` options** — when the initial hydrate helper didn't
  run in time on an OOB slideover swap (race already documented in
  v0.2.5), the per-select `change` listener was never bound either.
  Result: switching the service dropdown from, say, openrouter to
  ollama kept the 6 openrouter options visible in the model_override
  dropdown. The JS now uses **document-level event delegation** for
  the change event — no per-form binding needed, so the timing of
  the hydrate pass is irrelevant. Shared rebuild logic is extracted
  into `wvRebuildModelOverride(form)`; the old `wvInitModelOverride`
  stays as a back-compat initial hydrate but only needs to run once
  per form lifetime.

---

## [0.2.12] — 2026-04-17

### Fixed

- **`/status` is now token-aware** — previously it always returned
  the proxy's own client config, so an observability consumer polling
  `/status` via a different client's proxy saw that proxy's own model
  (e.g. `gemini-3.1-flash-lite-preview`) instead of the caller's model
  (e.g. `google/gemini-3.1-pro-preview`). The consumer's model badge
  reported the wrong value even though its `ai_config.json` was
  correct. With a Bearer token header, `/status` now looks up the
  caller's client config via `lookupTokenConfig` and returns that
  client's `client/service/model` instead. Without a token it still
  returns the proxy's own config (backward compatible).
- **`/api/token/config` prefers v0.2 canonical fields** — the vault
  endpoint now returns `preferred_service` / `model_override` when
  set, falling back to legacy `default_service` / `default_model`.
  JSON field names stay `default_*` for wire-format back-compat.
  Without this, the `model_override` a user sets via the dashboard
  wouldn't propagate to token-lookup consumers.

### Notes for consumers

- EconoWorld analyzer (or any external poller that wants per-client
  status): include `Authorization: Bearer <your-token>` on `/status`
  GETs. The response will reflect your client's routing, not the
  proxy's own.

---

## [0.2.11] — 2026-04-17

### Fixed

- **EconoWorld `ai_config.json` now follows service defaults after a
  soft-cleared `model_override`**: v0.2.10 wired the SSE path, but
  `updateEconoWorldModel` short-circuits on an empty model — which is
  exactly the state vault arrives at via v0.2.7's soft-clear when
  the user switches `preferred_service` and the old override is no
  longer in the new service's `allowed_models`. Net effect: vault
  recorded the change, `/status.model` updated (via v0.2.9's
  `serviceDefaults` fallback), but `ai_config.json` kept the stale
  override and drifted from the proxy's actual routing. The SSE
  handler's `case "econoworld"` now does the same fallback —
  `effective := mdl; if effective == "" && svc != "" { effective =
  s.serviceDefaults[svc] }` — before calling
  `updateEconoWorldModel`. `cline` and `claude-code` keep their
  existing empty-model short-circuit.

---

## [0.2.10] — 2026-04-17

### Fixed

- **EconoWorld model changes no longer lost on the SSE path**: when a
  user changed the `econoworld` client's `model_override` in the vault
  dashboard, the proxy received the `config_change` event but
  silently dropped it because `server.go`'s `onAnyConfig` switch only
  handled `cline` and `claude-code`. `ai_config.json` kept the old
  bootstrap model until someone triggered another `POST /agent/apply`.
  The switch now has a `case "econoworld": go updateEconoWorldModel(mdl)`
  branch that rewrites only `openai_compatible.model` in the local
  `ai_config.json`, preserving `base_url`/`api_key`/`max_tokens` from
  the prior bootstrap.

### Internal

- `updateEconoWorldModel(model string)` + `updateEconoWorldModelAt(path, model string)`
  added to `agent_apply.go`. The `At` variant is file-path-parameterised
  so it can be unit-tested; both short-circuit on an empty model
  (vault clears shouldn't clobber bootstrap), silent-skip on missing
  `ai_config.json` (proxies on hosts without EconoWorld installed),
  and leave files alone when they're missing the `openai_compatible`
  section (bootstrap responsibility).
- 4 new tests in `agent_apply_test.go`:
  - `TestUpdateEconoWorldModelAt_UpdatesModelField`
  - `TestUpdateEconoWorldModelAt_MissingFileIsSilent`
  - `TestUpdateEconoWorldModelAt_NoOpenAICompatSectionIsSilent`
  - `TestUpdateEconoWorldModel_EmptyModelIsNoop`

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
- **Footer**: GitHub / author site / email links + live uptime ticker
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
- **Bidirectional model sync for Cline**: When a Cline client's model is changed in the
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
  *all* foreign clients. When another client's model changed to `gemini-3.1-pro-preview`,
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
