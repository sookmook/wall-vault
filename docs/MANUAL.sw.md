# wall-vault User Manual

[English](MANUAL.md) · [한국어](MANUAL.ko.md) · [中文](MANUAL.zh.md) · [日本語](MANUAL.ja.md) · [Español](MANUAL.es.md) · [Français](MANUAL.fr.md) · [Deutsch](MANUAL.de.md) · [Português](MANUAL.pt.md) · [العربية](MANUAL.ar.md) · [हिन्दी](MANUAL.hi.md) · [Bahasa Indonesia](MANUAL.id.md) · [ภาษาไทย](MANUAL.th.md) · Kiswahili · [Hausa](MANUAL.ha.md) · [नेपाली](MANUAL.ne.md) · [Монгол](MANUAL.mn.md) · [isiZulu](MANUAL.zu.md)

Mwongozo huu unahusu kusakinisha, kusanidi, na kuendesha wall-vault. Kwa muhtasari wa haraka, angalia [README](../README.md). Kwa maelezo ya HTTP API, angalia [API reference](API.md).

## Yaliyomo

1. [wall-vault inafanya nini](#wall-vault-inafanya-nini)
2. [Usakinishaji](#usakinishaji)
3. [Kuendesha kwa mara ya kwanza na setup wizard](#kuendesha-kwa-mara-ya-kwanza-na-setup-wizard)
4. [Kuwezesha TLS](#kuwezesha-tls)
5. [Kusajili API key](#kusajili-api-key)
6. [Kuunganisha agent](#kuunganisha-agent)
7. [Dashboard](#dashboard)
8. [Hali ya distributed](#hali-ya-distributed)
9. [Kuanza kiotomatiki](#kuanza-kiotomatiki)
10. [Plugin yamls](#plugin-yamls)
11. [Doctor](#doctor)
12. [Hooks](#hooks)
13. [Vigezo vya mazingira](#vigezo-vya-mazingira)
14. [Utatuzi wa matatizo](#utatuzi-wa-matatizo)

---

## wall-vault inafanya nini

wall-vault ni Go binary moja inayobeba huduma mbili zinazoshirikiana:

- **vault** huhifadhi API key zilizosimbwa wakati wa kupumzika (AES-GCM kwa nenosiri kuu), hufuatilia matumizi na cooldown kwa kila key, hutangaza mabadiliko kupitia Server-Sent Events (SSE), na hutoa web dashboard kwenye `:56243` kwa waendeshaji wa kibinadamu.
- **proxy** hufunua endpoint zinazoendana na Gemini, Anthropic, OpenAI, na Ollama-native kwenye `:56244`. AI client yoyote inayoelekeza kwenye proxy hutumia key zilizo katika vault — wateja kamwe hawazioni. Wakati upstream moja inashindwa, dispatch huangukia kwa provider inayofuata kwa mpangilio.

Hii ni muhimu wakati:

- Una key za provider kadhaa na unataka URL moja ambayo agent inazungumza nayo.
- Unataka key ya free-tier iliyo kwenye cooldown ipishe bila kuvunja session.
- Unataka key zile zile ziwashe bot, IDE, au script kadhaa kwenye LAN moja bila kunakili credentials.
- Unataka dashboard, sio vigezo vya mazingira, kwa kuhariri key na kubadilisha model.
- Unataka fallback ya ndani (Ollama, LM Studio, vLLM) wakati ukomo wa cloud unapoisha.

```
   AI client (OpenClaw, Claude Code, Cursor, …)
            │
            ▼
   wall-vault proxy  :56244
            │  (selects key, dispatches, falls back on failure)
            ├──► Google Gemini
            ├──► Anthropic
            ├──► OpenAI
            ├──► OpenRouter (340+ models, auto :free fallback)
            └──► Local OAI-compat backends (Ollama / LM Studio / vLLM / …)

   vault (AES-GCM key store + dashboard)  :56243
            ▲
            │  SSE broadcast on change
   Multiple proxies on different hosts can share one vault.
```

---

## Usakinishaji

### Linux / macOS one-liner

```bash
curl -fsSL https://raw.githubusercontent.com/sookmook/wall-vault/main/install.sh | sh
```

Script hugundua OS na architecture kiotomatiki, hupakua binary sahihi katika `~/.local/bin/wall-vault`, na huifanya iwe inayoweza kutekelezwa. Ikiwa `~/.local/bin` haiko kwenye `PATH` yako, iongeze:

```bash
export PATH="$HOME/.local/bin:$PATH"
```

### Kupakua kwa mkono

Binary zilizojengwa awali huchapishwa kila release katika `https://github.com/sookmook/wall-vault/releases`.

```bash
# Linux amd64
curl -L https://github.com/sookmook/wall-vault/releases/latest/download/wall-vault-linux-amd64 \
  -o wall-vault && chmod +x wall-vault

# Linux arm64 (Raspberry Pi, ARM servers)
curl -L https://github.com/sookmook/wall-vault/releases/latest/download/wall-vault-linux-arm64 \
  -o wall-vault && chmod +x wall-vault

# macOS Apple Silicon
curl -L https://github.com/sookmook/wall-vault/releases/latest/download/wall-vault-darwin-arm64 \
  -o wall-vault && chmod +x wall-vault

# macOS Intel
curl -L https://github.com/sookmook/wall-vault/releases/latest/download/wall-vault-darwin-amd64 \
  -o wall-vault && chmod +x wall-vault
```

### Windows

```powershell
Invoke-WebRequest -Uri `
  "https://github.com/sookmook/wall-vault/releases/latest/download/wall-vault-windows-amd64.exe" `
  -OutFile wall-vault.exe
```

### Kujenga kutoka source

Inahitaji Go 1.25 au mpya zaidi.

```bash
git clone https://github.com/sookmook/wall-vault
cd wall-vault
go build -o wall-vault .
```

`make build-all` hu-cross-compile kwa platforms zote tano zinazoungwa mkono. Binary huishia kwa `bin/`.

---

## Kuendesha kwa mara ya kwanza na setup wizard

```bash
wall-vault setup
```

Wizard hukuuliza, kwa mpangilio:

1. **Lugha** — huchagua moja ya locale 17 za UI. Hugunduliwa kiotomatiki kutoka `$LANG`; wizard hutoa orodha hata hivyo.
2. **Mandhari** — `light` (chaguo-msingi), `dark`, `cherry`, `ocean`, `gold`, `autumn`, `winter`. Mapambo tu.
3. **Hali** — `standalone` (host moja, chaguo-msingi) au `distributed` (vault kwenye host moja, proxy kwenye nyingine).
4. **Jina la bot** — slug ya `client_id` ya bure. Vault hutumia hii kufungia config kwa kila client (model overrides, fallback chains).
5. **Proxy port** — chaguo-msingi `56244`.
6. **Vault port** — chaguo-msingi `56243` (standalone tu).
7. **Uchaguzi wa huduma** — y/N kwa kila moja ya: Google Gemini, OpenRouter, Anthropic, OpenAI, Ollama, LM Studio, vLLM. Chaguo nyingi ni sawa; kila moja huandika dokezo lake la env-var mwishoni.
8. **Tool filter** — `strip_all` (chaguo-msingi; huzuia tool definitions zote zinazoingia kwa usalama) au `passthrough` (huruhusu tool yoyote ipite).
9. **Admin token** — acha tupu ili kuzalisha kiotomatiki. Dashboard inahitaji token hii kuingia.
10. **Nenosiri kuu** — acha tupu kwa kutotumia encryption (HAIPENDEKEZWI); weka thamani ili AES-GCM isimbe key store wakati wa kupumzika.
11. **Njia ya kuhifadhi** — chaguo-msingi ni `wall-vault.yaml` katika directory ya sasa. Loader pia huangalia katika `~/.wall-vault/config.yaml`.

Baada ya kuhifadhi, wizard huendesha `doctor.FixTrust` ili agent yoyote iliyosakinishwa kienyeji (OpenClaw, Claude Code, Cline) ipate CA ya ndani ya wall-vault iliyoongezwa kwenye trust store yake kiotomatiki. Ikiwa hakuna agent kama hiyo iliyosakinishwa, hatua hii huchapisha `SKIP` na haiandiki chochote.

Kisha anza binary:

```bash
wall-vault start
```

`start` huendesha vault na proxy katika process moja (hali ya standalone). Kwa hali ya distributed tumia `wall-vault vault` kwenye vault host na `wall-vault proxy` kwenye kila proxy host.

Fungua `http://localhost:56243` kwenye browser. Ingia na admin token ambayo wizard ilichapisha.

---

## Kuwezesha TLS

Vifani-msingi vya wizard huacha listener zote mbili kwenye HTTP wazi. Agent nyingi (OpenClaw, Claude Code, Cursor) hufanya kazi vyema dhidi ya HTTPS endpoint moja, hivyo TLS inapendekezwa katika deployment yoyote inayoenea zaidi ya mashine ya kienyeji.

wall-vault inakuja na CA yake ya ndani hivyo huhitaji jina la DNS la umma au Let's Encrypt.

```bash
# 1. Unda CA ya ndani — imeandikwa kwa ~/.wall-vault/ca.{crt,key}.
#    CA ni nzuri kwa miaka 10 kwa chaguo-msingi; pindua na --ca-years.
wall-vault cert init

# 2. Toa cheti cha host. Subject Alternative Names hujumuisha kiotomatiki:
#       hostname, "localhost", "127.0.0.1", na LAN IP yoyote isiyo ya loopback iliyogunduliwa.
#    Pindua issuer dir na --dir, validity na --host-years.
wall-vault cert issue $(hostname)

# 3. Amini CA katika OS keychain ya mashine hii.
#    Linux: huandika kwa /etc/ssl/certs/ kupitia update-ca-certificates (inahitaji sudo).
#    macOS: huongeza kwa System keychain kupitia security add-trusted-cert (inahitaji sudo).
#    Windows: hu-import katika CurrentUser\Root kupitia certutil (haiitaji admin).
wall-vault cert install-trust

# 4. Wezesha TLS kwenye listener zote mbili.
export WV_PROXY_TLS_ENABLED=1
export WV_PROXY_TLS_CERT="$HOME/.wall-vault/$(hostname).crt"
export WV_PROXY_TLS_KEY="$HOME/.wall-vault/$(hostname).key"
export WV_VAULT_TLS_ENABLED=1
export WV_VAULT_TLS_CERT="$HOME/.wall-vault/$(hostname).crt"
export WV_VAULT_TLS_KEY="$HOME/.wall-vault/$(hostname).key"

wall-vault start
```

Ili kupanua trust kwa mashine zingine za LAN, nakili `~/.wall-vault/ca.crt` na endesha `wall-vault cert install-trust --ca <path>` kwenye kila moja. Vault pia hufunua `ca.crt` kupitia plain-HTTP listener ndogo kwenye `:56247` (**bootstrap port**) kwa kesi ya catch-22 ambapo client mpya inahitaji CA kuzungumza HTTPS.

### Loopback HTTP companion

Agent zingine — hasa Node runtime ya OpenClaw — huandika upya `NODE_EXTRA_CA_CERTS` wakati wa kuanzisha process, zikidondosha CA hint yoyote iliyotolewa na operator. Haziwezi kuheshimu CA ya wall-vault kutoka ndani ya daemon, hata baada ya `cert install-trust`. wall-vault hupata njia ya kuzunguka hili kwa kufunga **plain-HTTP listener ya loopback-tu** ya ziada kwenye `127.0.0.1:56245` wakati wowote TLS imewezeshwa. Wateja wa host moja hufikia proxy kupitia port hiyo bila TLS kabisa; wateja wa LAN huendelea kutumia TLS listener.

Lemaza na `WV_PROXY_PLAIN_PORT=0` ikiwa huitaji.

### `wall-vault cert list`

Huonyesha kila cheti chini ya `~/.wall-vault/` na subject, validity window, na SANs.

```
$ wall-vault cert list
ca.crt          subject=wall-vault internal CA   not-after=2036-05-05
hostname.crt    subject=hostname                 not-after=2031-05-05   SAN=hostname,localhost,127.0.0.1,192.168.…
```

---

## Kusajili API key

Njia mbili: dashboard, au vigezo vya mazingira.

### Dashboard (inapendekezwa)

1. Ingia katika `https://localhost:56243` na admin token.
2. Bonyeza **+ API key** katika kadi ya keys.
3. Chagua huduma (Google, OpenRouter, Anthropic, OpenAI, …).
4. Bandika key. Hifadhi.

Key nyingi kwa kila huduma ni sawa; proxy hu-round-robin kati yao na kuruka zile zinazogonga cooldown ya per-key.

### Vigezo vya mazingira (bootstrap ya mara moja)

```bash
export WV_KEY_GOOGLE="AIzaSyA1...,AIzaSyB2...,AIzaSyC3..."   # comma-separated
export WV_KEY_OPENROUTER="sk-or-v1-…"
export WV_KEY_ANTHROPIC="sk-ant-…"
export WV_KEY_OPENAI="sk-…"
wall-vault start
```

Key zilizotolewa kwa njia hii huandikwa kwenye encrypted store wakati wa kuanza kwa mara ya kwanza. Kuanza kwa baadaye huzisoma kutoka diski; unaweza ku-unset env var baada ya run ya kwanza.

### Cooldown na rotation

Kila call inayofanikiwa huongeza `usage_count` ya key na kuonyesha upya `last_used`. Kwenye HTTP 429 / 402 / 403, proxy huweka key kwenye **cooldown** (chaguo-msingi: dakika 60 kwa 429, masaa 24 kwa 402, masaa 12 kwa 403). Dispatch inayofuata huchagua key tofauti kwa huduma hiyo. Wakati key zote za huduma ziko kwenye cooldown, proxy huruka huduma hiyo haraka kabisa na kujaribu provider inayofuata katika fallback chain.

Cooldown zinaonekana kwa kila key katika dashboard na hesabu ya kushuka.

---

## Kuunganisha agent

### OpenClaw

OpenClaw ni client ya asili inayolengwa. Tumia modal ya **+ Add agent** ya dashboard:

- Weka **Agent type** kuwa `openclaw` au `nanoclaw`.
- Weka **Work directory** — kwa OpenClaw hii hujaza kiotomatiki kama `~/.openclaw`.
- Chagua **preferred service** na ikiwezekana **model override**.
- Bonyeza **Apply**. wall-vault huandika `~/.openclaw/openclaw.json` moja kwa moja (provider URLs, vault token, model entries).

Unapobadilisha model kutoka dashboard, OpenClaw huchukua mabadiliko kupitia SSE ndani ya sekunde 1–3 — hakuna restart.

### Claude Code

```bash
export ANTHROPIC_BASE_URL=https://localhost:56244
export ANTHROPIC_API_KEY=<your-vault-client-token>
claude
```

Wakati credit za upstream Anthropic zinapoisha, dispatch huangukia kwa huduma yoyote iliyoorodheshwa katika `fallback_services` ya client hii. Kwa chaguo-msingi, model id isiyo ya Claude iliyotumwa kwa anthropic dispatch hurudisha error ili misrouting ionekane mara moja. Chagua opt in kwa kuandika upya kiotomatiki:

```yaml
proxy:
  anthropic_fallback_model: "claude-haiku-4-5-20251001"
```

### Cursor

Katika Cursor **Settings → AI → OpenAI API**:

```
Base URL:  https://localhost:56244
API Key:   <your-vault-client-token>
Model:     gemini-2.5-flash    # or any model wall-vault knows
```

### Continue (VS Code, JetBrains)

`config.json`:

```json
{
  "models": [
    {
      "title": "wall-vault",
      "provider": "openai",
      "model": "gemini-2.5-flash",
      "apiBase": "https://localhost:56244/v1",
      "apiKey": "<your-vault-client-token>"
    }
  ]
}
```

### Custom HTTP

```bash
curl -X POST https://localhost:56244/v1/chat/completions \
  -H "Authorization: Bearer <your-vault-client-token>" \
  -H "Content-Type: application/json" \
  -d '{
    "model": "gemini-2.5-flash",
    "messages": [{"role": "user", "content": "hello"}]
  }'
```

Endpoint hiyo hiyo hukubali streaming (`"stream": true`) wakati `proxy.oai_stream_forward: true` imewekwa.

---

## Dashboard

`https://localhost:56243`. Kadi tano kwenye home grid:

- **Keys** — kila API key, iliyowekwa kwa huduma. Ongeza, hariri, futa; ona matumizi na cooldown.
- **Services** — Google / OpenRouter / Anthropic / OpenAI / Ollama / LM Studio / vLLM / llama.cpp, pamoja na plugin yaml yoyote katika `~/.wall-vault/services/`. Weka per-service `default_model`, `allowed_models`, base URL, kifungo cha reasoning.
- **Clients (agents)** — kila client iliyosajiliwa (OpenClaw bot, Claude Code session, Cursor instance, …). Toa preferred service, model override, fallback chain.
- **Proxies** — kila proxy iliyo-authenticate dhidi ya vault hii. Hali ya live (online/offline), iliyoonekana mwisho, model ya sasa.
- **Settings** — admin token, mzunguko wa nenosiri kuu, mandhari, lugha.

Kila kadi ina edit slideover (upande wa kulia). Bonyeza nje au `Esc` kufunga. Mabadiliko husukumwa kwa proxy zote zilizounganishwa kupitia SSE ndani ya sekunde.

**Footer** hubeba kiashiria cha SSE (kijani = imeunganishwa, machungwa = inaunganisha tena, kijivu = imekatika) na live build version.

---

## Hali ya distributed

Una mashine kadhaa zote zinazohitaji key zile zile, endesha vault kwenye host moja na proxy kwenye kila moja ya zingine.

### Vault host

```bash
WV_VAULT_HOST=0.0.0.0 \
WV_ADMIN_TOKEN=<admin> \
WV_MASTER_PASS=<master> \
wall-vault vault
```

Sasa dashboard inafikika katika `https://<vault-host>:56243`. Ongeza agent kwa kila proxy ya mbali katika kadi ya **Clients**; kila moja huzalisha `vault_token` ya kipekee.

### Proxy hosts

```bash
WV_VAULT_URL=http://<vault-host>:56243 \
WV_VAULT_TOKEN=<that-client-token> \
WV_PROXY_HOST=0.0.0.0 \
wall-vault proxy
```

Proxy hu-authenticate dhidi ya vault, hufungua SSE stream, na hutekeleza config yoyote inayopokea (preferred service, model override, fallback chain). Edit za vault za baadaye hutua ndani ya sekunde bila restart.

Kwa usakinishaji unaoenea LAN, wezesha TLS kwenye vault host (`WV_VAULT_TLS_ENABLED=1` + cert/key env vars) na endesha kila proxy host kupitia hatua ile ile ya `wall-vault cert install-trust` ili HTTPS calls za proxy ndani ya vault ziaminike.

---

## Kuanza kiotomatiki

### systemd (Linux)

```ini
# ~/.config/systemd/user/wall-vault-proxy.service
[Unit]
Description=wall-vault proxy
After=network-online.target

[Service]
Type=simple
ExecStart=%h/.local/bin/wall-vault proxy
Restart=on-failure
RestartSec=5

[Install]
WantedBy=default.target
```

```bash
systemctl --user enable --now wall-vault-proxy
loginctl enable-linger $USER       # so the unit keeps running after logout
```

Kwa vault kwenye host moja, andika `wall-vault-vault.service` sambamba. Kwa hali ya standalone, unit moja inayoita `wall-vault start` inatosha.

### launchd (macOS)

```xml
<!-- ~/Library/LaunchAgents/com.wall-vault.proxy.plist -->
<?xml version="1.0" encoding="UTF-8"?>
<!DOCTYPE plist PUBLIC "-//Apple//DTD PLIST 1.0//EN"
  "http://www.apple.com/DTDs/PropertyList-1.0.dtd">
<plist version="1.0">
<dict>
  <key>Label</key><string>com.wall-vault.proxy</string>
  <key>ProgramArguments</key>
  <array><string>/usr/local/bin/wall-vault</string><string>proxy</string></array>
  <key>RunAtLoad</key><true/>
  <key>KeepAlive</key><true/>
  <key>StandardOutPath</key><string>/tmp/wall-vault.proxy.log</string>
  <key>StandardErrorPath</key><string>/tmp/wall-vault.proxy.err</string>
</dict>
</plist>
```

```bash
launchctl load ~/Library/LaunchAgents/com.wall-vault.proxy.plist
```

### Windows

Tumia `nssm` ku-wrap `wall-vault.exe start` kama Windows service, au entry ya `schtasks` inayoendesha wakati wa user logon.

---

## Plugin yamls

Backend yoyote inayoendana na OpenAI inaweza kuongezwa bila mabadiliko ya code kwa kudondosha yaml chini ya `~/.wall-vault/services/`. wall-vault huipakia wakati wa startup na husajili huduma kwa dispatch, set ya OAI-compat detection, na Gemini-stream bridge.

```yaml
# ~/.wall-vault/services/llamacpp.yaml
id: llamacpp                 # unique service id
name: llama.cpp              # human label
enabled: true                # disabled plugins are skipped at load

default_url: http://localhost:8080   # operator override; env wins (WV_LLAMACPP_URL)
endpoints:
  generate: /v1/chat/completions
  list_models: /v1/models

auth:
  type: none                 # none | bearer | query_param | header
  param: ""                  # for query_param: the param name (e.g. "key")

request_format: openai       # openai | gemini | ollama | raw

model_fetch:
  enabled: true              # let the dashboard auto-detect models
  dynamic: true              # re-fetch on every dashboard open
  auto_detect_url: true      # try /v1/models even when not declared

concurrency:
  max: 1                     # max concurrent requests to this backend
  queue_size: 10
  wait_notify: true          # show "queued" hint to TUI agents

error_codes:
  503:
    cooldown: 5m
    message: "llama.cpp not responding"

# Opt in to qwen3-family inline /no_think directive when reasoning is off.
# Set true if your backend's chat template strips the marker (LM Studio's
# jinja, Ollama's /v1 layer). Other backends typically echo the literal
# text back, so this stays opt-in per yaml.
inline_no_think_for_qwen3: false

# Hub topology — point at another wall-vault. Required when this plugin
# fronts a remote wall-vault (so the receiving wall-vault sees the
# publisher prefix and routes correctly) and so the bearer token in
# proxy.vault_token is sent as Authorization.
preserve_model_id: false
tls_internal_ca: false       # add ~/.wall-vault/ca.crt to client trust pool
```

Set iliyojumuishwa katika `configs/services/` (lmstudio, vllm, llamacpp, tgwui, localai, jan, koboldcpp, tabbyapi, mlx-server, litellm-proxy, ollama, google, openrouter) huja imelemazwa kwa chaguo-msingi. Nakili ile unayotaka kwa `~/.wall-vault/services/`, weka `enabled: true`, anza upya.

---

## Doctor

`wall-vault doctor` huendesha health probe ya mara moja kwenye usakinishaji wote:

```
✓ vault listener  (https://localhost:56243)
✓ proxy listener  (https://localhost:56244)
✓ master password set
⚠ Google: 2 keys, all on cooldown
✓ Anthropic: 1 key healthy
✗ Ollama: not reachable at http://localhost:11434
```

Kila line ni moja ya:

- `✓` — afya nzuri
- `⚠` — imepunguzwa lakini inafanya kazi (key moja imepoa, kiwango cha chini, n.k.)
- `✗` — imevunjika
- `SKIP` — haijasanidiwa / haitumiki kwenye host hii

Hali ya pili ya daemon huendesha probe ile ile kila `doctor.interval` (chaguo-msingi dakika 5) na huandika matokeo kwa `doctor.log_file` (chaguo-msingi `/tmp/wall-vault-doctor.log`). Wakati `doctor.auto_fix` ni true, pia hujaribu kurekebisha drift za kawaida (config ya OpenClaw iliyochakaa, TLS trust iliyokosekana, huduma zinazoweza kuanzwa upya).

Anzisha mara moja kutoka dashboard kupitia kadi ya **Doctor** au `wall-vault doctor`.

---

## Hooks

Endesha amri ya shell kwenye matukio ya key:

```yaml
hooks:
  on_model_change:   "logger 'wall-vault: $SERVICE/$MODEL'"
  on_key_exhausted:  "notify-send 'wall-vault' '$SERVICE keys all on cooldown'"
  on_service_down:   "/usr/local/bin/page-oncall.sh $SERVICE '$ERROR'"
  on_doctor_fix:     "echo \"$AGENT: $LEVEL $MSG\" >> ~/wall-vault.audit.log"
  openclaw_socket:   ""    # if set, OpenClaw TUI receives events over this Unix socket
```

Kila hook hupata vigezo vya mazingira maalum kwa tukio (`SERVICE`, `MODEL`, `ERROR`, `AGENT`, `LEVEL`, `MSG`). Hooks huendesha async na timeout ya sekunde 5 — proxy kamwe haisubiri hook ya polepole.

---

## Vigezo vya mazingira

| Variable | YAML field |
|----------|------------|
| `WV_LANG` | `lang` |
| `WV_THEME` | `theme` |
| `WV_PROXY_PORT` | `proxy.port` |
| `WV_PROXY_HOST` | `proxy.host` |
| `WV_VAULT_PORT` | `vault.port` |
| `WV_VAULT_HOST` | `vault.host` |
| `WV_VAULT_URL` | `proxy.vault_url` (distributed) |
| `WV_VAULT_TOKEN` | `proxy.vault_token` |
| `WV_ADMIN_TOKEN` | `vault.admin_token` |
| `WV_MASTER_PASS` | `vault.master_password` |
| `WV_AVATAR` | `proxy.avatar` |
| `WV_TOOL_FILTER` | `proxy.tool_filter` |
| `WV_CC_CLIENT_ID` | `proxy.claude_code_client_id` |
| `WV_PROXY_TLS_ENABLED` | `proxy.tls.enabled` |
| `WV_PROXY_TLS_CERT` | `proxy.tls.cert_file` |
| `WV_PROXY_TLS_KEY` | `proxy.tls.key_file` |
| `WV_VAULT_TLS_ENABLED` | `vault.tls.enabled` |
| `WV_VAULT_TLS_CERT` | `vault.tls.cert_file` |
| `WV_VAULT_TLS_KEY` | `vault.tls.key_file` |
| `WV_VAULT_BOOTSTRAP_PORT` | `vault.bootstrap_port` |
| `WV_PROXY_PLAIN_PORT` | `proxy.plain_port` |
| `WV_KEY_GOOGLE` | One-shot import: comma-separated Google keys |
| `WV_KEY_OPENROUTER` | One-shot import: OpenRouter keys |
| `WV_KEY_ANTHROPIC` | One-shot import: Anthropic keys |
| `WV_KEY_OPENAI` | One-shot import: OpenAI keys |
| `WV_OLLAMA_URL` | Per-host Ollama URL override |
| `WV_OLLAMA_KEEP_ALIVE` | `proxy.ollama_keep_alive` |
| `WV_OLLAMA_NUM_CTX` | `proxy.ollama_num_ctx` |
| `WV_LMSTUDIO_URL`, `WV_VLLM_URL`, `WV_LLAMACPP_URL` | Per-backend URL override |
| `WV_TOKEN_SENTINEL_FALLBACK` | `proxy.token_sentinel_fallback` |
| `WV_OAI_STREAM_FORWARD` | `proxy.oai_stream_forward` |
| `WV_ANTHROPIC_FALLBACK_MODEL` | `proxy.anthropic_fallback_model` |
| `WV_ECONOWORLD_MAX_TOKENS` | `proxy.econoworld_max_tokens` |
| `WV_ECONOWORLD_STREAM` | `proxy.econoworld_stream` |
| `WV_ECONOWORLD_REQUEST_TIMEOUT` | `proxy.econoworld_request_timeout` |

Kila env var, ikiwekwa, hushinda faili ya YAML.

---

## Utatuzi wa matatizo

### `connection refused` kwenye `:56244`

Ama proxy haiendi au imefungwa kwenye host tofauti. Angalia:

```bash
ss -lnp | grep 56244
systemctl --user status wall-vault-proxy   # Linux
launchctl list | grep wall-vault           # macOS
```

Ikiwa inaendesha kwenye port tofauti, config yako ina `proxy.port` iliyobadilishwa — angalia `~/.wall-vault/config.yaml`.

### `x509: certificate signed by unknown authority`

Client haiamini CA ya ndani ya wall-vault. Endesha `wall-vault cert install-trust` kwenye mashine ya client. Kwa agent ambazo runtime yake huipuuza OS trust store (k.m. Node yenye `NODE_EXTRA_CA_CERTS` iliyowekwa hard-coded), tumia loopback HTTP companion kwenye `127.0.0.1:56245` (host moja tu) au weka `WV_PROXY_TLS_ENABLED=0` ili kuangukia kwa HTTP wazi.

### `token not registered with vault`

`Authorization: Bearer <token>` ya client hailingani na client yoyote iliyosajiliwa. Thibitisha token chini ya **Clients** kwenye dashboard. Ikiwa ulinakili token literal kama `proxy-managed`, `dummy`, au `""` kutoka config iliyochakaa, ibadilishe na client token halisi.

### `Anthropic dispatch needs a Claude model id`

Tabia ya chaguo-msingi tangu v0.2.63: model id isiyo ya Claude iliyotumwa kwa anthropic dispatch hurudisha error. Ama rekebisha routing (usitume `gemini-2.5-flash` kwa anthropic) au chagua opt in kwa kuandika upya kiotomatiki kupitia `proxy.anthropic_fallback_model`.

### `unknown service: <id>`

Dispatch iliona service id ambayo hakuna plugin yaml iliyoidai. Angalia:

```bash
ls ~/.wall-vault/services/        # any plugin yaml present?
cat ~/.wall-vault/services/<id>.yaml | grep enabled
```

Ikiwa yaml ipo lakini ni `enabled: false`, igeuze. Ikiwa imepotea kabisa, nakili kutoka `configs/services/` katika source tree.

### Response tupu kwenye reasoning model

`qwen3.6`, `deepseek-r1`, na familia ya GPT-`o1` wakati mwingine hutoa `reasoning_content` tu na kuacha `content` tupu. Tangu v0.2.63 wall-vault huangukia kwa reasoning text kiotomatiki — ikiwa bado unaona response tupu, backend hairudishi field yoyote. Angalia logs za upstream.

Kwa LM Studio na qwen3 hasa, weka `inline_no_think_for_qwen3: true` katika plugin yaml ili reasoning ilemazwe inline. lmstudio.yaml na ollama.yaml zilizojengwa-ndani tayari zinafanya hivi.

### Dashboard inaonyesha "all keys on cooldown" lakini nimeongeza moja tu

Key mpya ina afya nzuri lakini dispatch path bado inaweza kuwa katika cooldown ya key ya zamani. Jaribu request mpya — proxy hu-round-robin kwa kila call, na key yenye afya itachaguliwa baadaye.

### Vault haitafungua na nenosiri kuu

Nenosiri sio sahihi. Hakuna recovery — wall-vault kwa makusudi haitumii backdoor. Ikiwa umepoteza nenosiri kuu kweli, njia pekee ni kufuta `~/.wall-vault/data/vault.json`, anza upya na nenosiri jipya, na uongeze key tena.

### Ukomo wa Free-tier OpenRouter umefikiwa

Weka `proxy.services` ijumuishe `openrouter` na ongeza angalau OpenRouter key moja. Proxy hu-auto-fall-back kutoka model ya kulipia hadi `:free` variant yake wakati path ya kulipia inarudisha 402 / 429.

### `journalctl --user -u wall-vault-proxy` ni tupu

systemd `--user` logs huenda kwa journal ya user anayeziendesha. Ikiwa ulianza unit kama `root` au kupitia `sudo`, journal iko katika system instance badala yake — jaribu `journalctl -u wall-vault-proxy` bila `--user`.

---

## Zaidi

- HTTP API reference — angalia [API.md](API.md)
- Source — `https://github.com/sookmook/wall-vault`
- Ripoti za bug / maombi ya feature — GitHub Issues
- Historia ya release — [CHANGELOG.md](../CHANGELOG.md)
