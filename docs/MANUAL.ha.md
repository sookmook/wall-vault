# wall-vault User Manual

[English](MANUAL.md) · [한국어](MANUAL.ko.md) · [中文](MANUAL.zh.md) · [日本語](MANUAL.ja.md) · [Español](MANUAL.es.md) · [Français](MANUAL.fr.md) · [Deutsch](MANUAL.de.md) · [Português](MANUAL.pt.md) · [العربية](MANUAL.ar.md) · [हिन्दी](MANUAL.hi.md) · [Bahasa Indonesia](MANUAL.id.md) · [ภาษาไทย](MANUAL.th.md) · [Kiswahili](MANUAL.sw.md) · Hausa · [नेपाली](MANUAL.ne.md) · [Монгол](MANUAL.mn.md) · [isiZulu](MANUAL.zu.md)

Wannan jagora ya ƙunshi shigar da, sanyawa, da kuma sarrafa wall-vault. Don taƙaitaccen bayyani, duba [README](../README.md). Don cikakken HTTP API, duba [API reference](API.md).

## Abubuwan da ke Ciki

1. [Abin da wall-vault ke yi](#abin-da-wall-vault-ke-yi)
2. [Shigarwa](#shigarwa)
3. [Gudanarwa ta farko da setup wizard](#gudanarwa-ta-farko-da-setup-wizard)
4. [Kunna TLS](#kunna-tls)
5. [Yi rejista da API key](#yi-rejista-da-api-key)
6. [Haɗa agent](#haɗa-agent)
7. [Dashboard](#dashboard)
8. [Yanayin distributed](#yanayin-distributed)
9. [Auto-start](#auto-start)
10. [Plugin yamls](#plugin-yamls)
11. [Doctor](#doctor)
12. [Hooks](#hooks)
13. [Manyan masu canji](#manyan-masu-canji)
14. [Magance matsala](#magance-matsala)

---

## Abin da wall-vault ke yi

wall-vault wani Go binary ɗaya ne wanda ya ƙunshi ayyuka biyu masu haɗin gwiwa:

- **Vault** yana adana API key da aka ɓoye yayin hutawa (AES-GCM tare da master password), yana bin diddigin amfani da cooldown ga kowane key, yana watsa canje-canje ta hanyar Server-Sent Events (SSE), kuma yana samar da web dashboard a `:56243` ga masu kula da mutum.
- **Proxy** yana fitar da endpoints masu jituwa da Gemini, Anthropic, OpenAI, da Ollama-native a `:56244`. Duk wani AI client da ya nuna ga proxy yana amfani da key da ke cikin vault — clients ba sa ganinsu. Idan upstream ɗaya ya gaza, dispatch yana komawa zuwa provider na gaba a tsari.

Wannan yana da amfani lokacin:

- Kuna da key na masu samar da yawa kuma kuna son URL ɗaya wanda agent ke magana da shi.
- Kuna son key na free-tier mai cooldown ya tafi gefe ba tare da ya karya session ba.
- Kuna son key ɗaya ya bautar da bots, IDEs, ko scripts da yawa a kan LAN ɗaya ba tare da kwafi credentials ba.
- Kuna son dashboard, ba environment variables ba, don gyara key da canza models.
- Kuna son fallback na gida (Ollama, LM Studio, vLLM) lokacin da iyakokin cloud suka ƙare.

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

## Shigarwa

### Linux / macOS one-liner

```bash
curl -fsSL https://raw.githubusercontent.com/sookmook/wall-vault/main/install.sh | sh
```

Script ɗin yana gano OS da architecture ta atomatik, yana saukar da daidai binary cikin `~/.local/bin/wall-vault`, kuma yana sanya shi mai aiwatuwa. Idan `~/.local/bin` ba ya kan `PATH` ɗinka, ƙara shi:

```bash
export PATH="$HOME/.local/bin:$PATH"
```

### Saukewa da hannu

An buga binaries da aka gina a kowane release a `https://github.com/sookmook/wall-vault/releases`.

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

### Gina daga source

Yana buƙatar Go 1.25 ko sabo.

```bash
git clone https://github.com/sookmook/wall-vault
cd wall-vault
go build -o wall-vault .
```

`make build-all` yana cross-compile zuwa duka platforms guda biyar da ake tallafa. Binaries suna sauka cikin `bin/`.

---

## Gudanarwa ta farko da setup wizard

```bash
wall-vault setup
```

Wizard yana tambayarka, a tsari:

1. **Harshe** — yana zaɓar ɗaya daga UI locales 17. Ana gano shi ta atomatik daga `$LANG`; wizard yana ba da jeri duk da haka.
2. **Theme** — `light` (default), `dark`, `cherry`, `ocean`, `gold`, `autumn`, `winter`. Cosmetic kawai.
3. **Mode** — `standalone` (host ɗaya, default) ko `distributed` (vault a host ɗaya, proxies a kan wasu).
4. **Sunan bot** — `client_id` slug na kyauta. Vault yana amfani da wannan don iyakance config kowane client (model overrides, fallback chains).
5. **Proxy port** — default `56244`.
6. **Vault port** — default `56243` (standalone kawai).
7. **Zaɓin sabis** — y/N ga kowane: Google Gemini, OpenRouter, Anthropic, OpenAI, Ollama, LM Studio, vLLM. Zaɓuɓɓuka da yawa suna da kyau; kowanne yana rubuta env-var hint a ƙarshe.
8. **Tool filter** — `strip_all` (default; yana toshe duk shigowar tool definitions don tsaro) ko `passthrough` (yana barin kowane tool ya wuce).
9. **Admin token** — bar fanko don auto-generate. Dashboard yana buƙatar wannan token don shiga.
10. **Master password** — bar fanko don rashin encryption (BA A BA SHAWARA BA); saita ƙimar don AES-GCM ɓoye key store yayin hutawa.
11. **Hanyar adanawa** — defaults zuwa `wall-vault.yaml` a directory na yanzu. Loader kuma yana duba `~/.wall-vault/config.yaml`.

Bayan adanawa, wizard yana gudanar da `doctor.FixTrust` don kowane agent da aka shigar a gida (OpenClaw, Claude Code, Cline) ya sami CA na ciki na wall-vault aka ƙara zuwa trust store ɗinsa ta atomatik. Idan babu agent kamar wancan da aka shigar, mataki yana buga `SKIP` kuma ba ya rubuta komai.

Sa'an nan ka fara binary:

```bash
wall-vault start
```

`start` yana gudanar da vault da proxy duka cikin process ɗaya (yanayin standalone). Don yanayin distributed yi amfani da `wall-vault vault` a vault host kuma `wall-vault proxy` a kowane proxy host.

Buɗe `http://localhost:56243` cikin browser. Shiga da admin token wanda wizard ya buga.

---

## Kunna TLS

Defaults na wizard suna barin masu sauraro biyu a HTTP fanko. Yawancin agents (OpenClaw, Claude Code, Cursor) suna aiki da kyau a kan HTTPS endpoint ɗaya, don haka ana ba da shawarar TLS a kowane deployment da ya wuce na'urar gida.

wall-vault yana zuwa tare da CA na ciki nasa don haka ba kwa buƙatar sunan DNS na jama'a ko Let's Encrypt.

```bash
# 1. Ƙirƙiri CA na ciki — an rubuta zuwa ~/.wall-vault/ca.{crt,key}.
#    CA yana da kyau na shekaru 10 ta tsohon; sake rubutawa da --ca-years.
wall-vault cert init

# 2. Bayar da host certificate. Subject Alternative Names sun haɗa da ta atomatik:
#       hostname, "localhost", "127.0.0.1", da kowane LAN IP da aka gano wanda ba loopback ba.
#    Sake rubuta issuer dir da --dir, validity da --host-years.
wall-vault cert issue $(hostname)

# 3. Amince da CA a OS keychain na wannan na'ura.
#    Linux: yana rubutawa zuwa /etc/ssl/certs/ ta hanyar update-ca-certificates (yana buƙatar sudo).
#    macOS: yana ƙara zuwa System keychain ta hanyar security add-trusted-cert (yana buƙatar sudo).
#    Windows: yana shiga cikin CurrentUser\Root ta hanyar certutil (babu admin da ake buƙata).
wall-vault cert install-trust

# 4. Kunna TLS a masu sauraro duka biyu.
export WV_PROXY_TLS_ENABLED=1
export WV_PROXY_TLS_CERT="$HOME/.wall-vault/$(hostname).crt"
export WV_PROXY_TLS_KEY="$HOME/.wall-vault/$(hostname).key"
export WV_VAULT_TLS_ENABLED=1
export WV_VAULT_TLS_CERT="$HOME/.wall-vault/$(hostname).crt"
export WV_VAULT_TLS_KEY="$HOME/.wall-vault/$(hostname).key"

wall-vault start
```

Don faɗaɗa amincewa zuwa wasu na'urorin LAN, kwafi `~/.wall-vault/ca.crt` kuma gudanar da `wall-vault cert install-trust --ca <path>` a kowane ɗaya. Vault kuma yana fitar da `ca.crt` ta hanyar ƙaramin plain-HTTP listener a `:56247` (**bootstrap port**) don yanayin catch-22 inda sabon client ke buƙatar CA don magana da HTTPS.

### Loopback HTTP companion

Wasu agents — musamman Node runtime na OpenClaw — suna sake rubuta `NODE_EXTRA_CA_CERTS` a process spawn, suna jefa duk wani CA hint da operator ya bayar. Ba za su iya girmama wall-vault CA daga ciki na daemon ba, ko bayan `cert install-trust`. wall-vault yana zagaya wannan ta hanyar ɗaure wani **loopback-only plain-HTTP listener** a `127.0.0.1:56245` a kowane lokaci da TLS aka kunna. Same-host clients suna isa proxy ta wannan port ba tare da TLS ba; LAN clients suna ci gaba da amfani da TLS listener.

Kashe da `WV_PROXY_PLAIN_PORT=0` idan ba ku buƙatar shi.

### `wall-vault cert list`

Yana nuna kowane cert ƙarƙashin `~/.wall-vault/` tare da subject, validity window, da SANs.

```
$ wall-vault cert list
ca.crt          subject=wall-vault internal CA   not-after=2036-05-05
hostname.crt    subject=hostname                 not-after=2031-05-05   SAN=hostname,localhost,127.0.0.1,192.168.…
```

---

## Yi rejista da API key

Hanyoyi biyu: dashboard, ko environment variables.

### Dashboard (ana ba da shawara)

1. Shiga `https://localhost:56243` tare da admin token.
2. Danna **+ API key** a cikin keys card.
3. Zaɓi sabis (Google, OpenRouter, Anthropic, OpenAI, …).
4. Manna key. Adana.

Yawan keys kowane sabis suna da kyau; proxy yana round-robin tsakanin su kuma yana tsallake waɗanda suka buga per-key cooldown.

### Environment variables (one-shot bootstrap)

```bash
export WV_KEY_GOOGLE="AIzaSyA1...,AIzaSyB2...,AIzaSyC3..."   # comma-separated
export WV_KEY_OPENROUTER="sk-or-v1-…"
export WV_KEY_ANTHROPIC="sk-ant-…"
export WV_KEY_OPENAI="sk-…"
wall-vault start
```

Keys da aka bayar ta wannan hanya ana rubuta su cikin encrypted store a launch na farko. Saukowa da ke biyowa suna karanta su daga disk; zaka iya unset env vars bayan run na farko.

### Cooldowns da rotation

Kowane call mai nasara yana ƙara `usage_count` na key kuma yana sabunta `last_used`. A HTTP 429 / 402 / 403, proxy yana sa key a **cooldown** (defaults: minti 60 ga 429, sa'o'i 24 ga 402, sa'o'i 12 ga 403). Dispatch na gaba yana zaɓar wani key dabam don wannan sabis. Lokacin da duk keys na sabis suka kasance a cooldown, proxy yana saurin tsallake wannan sabis kuma yana gwada provider na gaba a fallback chain.

Cooldowns suna ganuwa per-key a dashboard tare da ƙidaya.

---

## Haɗa agent

### OpenClaw

OpenClaw shine asalin client da ake nufi. Yi amfani da modal ɗin **+ Add agent** na dashboard:

- Saita **Agent type** zuwa `openclaw` ko `nanoclaw`.
- Saita **Work directory** — ga OpenClaw wannan yana cika atomatik kamar `~/.openclaw`.
- Zaɓi **preferred service** kuma ƙwarai **model override**.
- Danna **Apply**. wall-vault yana rubuta `~/.openclaw/openclaw.json` kai tsaye (provider URLs, vault token, model entries).

Lokacin da kuka canza model daga dashboard, OpenClaw yana ɗaukar canjin ta SSE cikin sakanni 1–3 — babu restart.

### Claude Code

```bash
export ANTHROPIC_BASE_URL=https://localhost:56244
export ANTHROPIC_API_KEY=<your-vault-client-token>
claude
```

Lokacin da credit ɗin Anthropic upstream ya ƙare, dispatch yana komawa zuwa duk wanene aka jera a cikin `fallback_services` na wannan client. Ta tsohon, model id wanda ba na Claude ba da aka aika zuwa anthropic dispatch yana mayar da error don misrouting ya bayyana nan da nan. Zaɓi opt in zuwa rubutu na atomatik:

```yaml
proxy:
  anthropic_fallback_model: "claude-haiku-4-5-20251001"
```

### Cursor

A Cursor **Settings → AI → OpenAI API**:

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

Endpoint ɗaya yana karɓar streaming (`"stream": true`) lokacin da `proxy.oai_stream_forward: true` aka saita.

---

## Dashboard

`https://localhost:56243`. Cards biyar a kan home grid:

- **Keys** — kowane API key, da aka tara ta sabis. Ƙara, gyara, gogewa; ga amfani da cooldown.
- **Services** — Google / OpenRouter / Anthropic / OpenAI / Ollama / LM Studio / vLLM / llama.cpp, tare da kowane plugin yaml a `~/.wall-vault/services/`. Saita per-service `default_model`, `allowed_models`, base URL, reasoning toggle.
- **Clients (agents)** — kowane client da aka yi rejista (OpenClaw bot, Claude Code session, Cursor instance, …). Sanya preferred service, model override, fallback chain.
- **Proxies** — kowane proxy da ya tabbatar da wannan vault. Halayen rai (online/offline), an gani na ƙarshe, model na yanzu.
- **Settings** — admin token, juyawar master password, theme, harshe.

Kowane card yana da edit slideover (gefen dama). Latsa waje ko `Esc` yana rufewa. Ana tura canje-canje ga duk proxies da aka haɗa ta SSE cikin sakanni.

**Footer** yana ɗauke da SSE indicator (kore = an haɗa, lemu = sake haɗawa, toka = an cire) da live build version.

---

## Yanayin distributed

Lokacin da kuke da na'urori da yawa duk suna buƙatar keys ɗaya, gudanar da vault a host ɗaya kuma proxies a kowane na sauran.

### Vault host

```bash
WV_VAULT_HOST=0.0.0.0 \
WV_ADMIN_TOKEN=<admin> \
WV_MASTER_PASS=<master> \
wall-vault vault
```

Yanzu dashboard zai iya kaiwa a `https://<vault-host>:56243`. Ƙara agent ga kowane proxy mai nesa a card ɗin **Clients**; kowanne yana yin `vault_token` na musamman.

### Proxy hosts

```bash
WV_VAULT_URL=http://<vault-host>:56243 \
WV_VAULT_TOKEN=<that-client-token> \
WV_PROXY_HOST=0.0.0.0 \
wall-vault proxy
```

Proxy yana tabbatarwa a vault, yana buɗe SSE stream, kuma yana amfani da kowane config da ya samu (preferred service, model override, fallback chain). Edits na vault na gaba suna saukowa cikin sakanni ba tare da restart ba.

Don shigarwa da ke yawo a LAN, kunna TLS a vault host (`WV_VAULT_TLS_ENABLED=1` + cert/key env vars) kuma gudanar da kowane proxy host ta hanyar mataki ɗaya na `wall-vault cert install-trust` don kiran HTTPS na proxy zuwa vault ya zama mai amincewa.

---

## Auto-start

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

Ga vault a host ɗaya, rubuta `wall-vault-vault.service` mai daidaitawa. Don yanayin standalone, unit ɗaya da ke kira `wall-vault start` ya isa.

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

Yi amfani da `nssm` don nade `wall-vault.exe start` a matsayin Windows service, ko shigarwar `schtasks` da ke gudana a user logon.

---

## Plugin yamls

Kowane backend mai jituwa da OpenAI ana iya ƙara shi ba tare da canjin code ba ta hanyar zubar yaml ƙarƙashin `~/.wall-vault/services/`. wall-vault yana loaded shi a startup kuma yana yi rejista da sabis don dispatch, OAI-compat detection set, da Gemini-stream bridge.

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

Saitin da aka haɗa a `configs/services/` (lmstudio, vllm, llamacpp, tgwui, localai, jan, koboldcpp, tabbyapi, mlx-server, litellm-proxy, ollama, google, openrouter) yana zuwa kashe ta tsohon. Kwafi wanda kuke so zuwa `~/.wall-vault/services/`, saita `enabled: true`, sake fara.

---

## Doctor

`wall-vault doctor` yana gudanar da one-shot health probe akan duk shigarwa:

```
✓ vault listener  (https://localhost:56243)
✓ proxy listener  (https://localhost:56244)
✓ master password set
⚠ Google: 2 keys, all on cooldown
✓ Anthropic: 1 key healthy
✗ Ollama: not reachable at http://localhost:11434
```

Kowane layi shine ɗaya daga cikin:

- `✓` — lafiya
- `⚠` — ya ragu amma yana aiki (key ɗaya ya yi cooldown, low quota, da sauransu)
- `✗` — ya karye
- `SKIP` — ba a saita ba / ba ya dacewa a kan wannan host

Yanayin daemon na biyu yana gudanar da probe ɗaya a kowane `doctor.interval` (tsohon minti 5) kuma yana rubuta sakamako zuwa `doctor.log_file` (tsohon `/tmp/wall-vault-doctor.log`). Lokacin da `doctor.auto_fix` ya zama gaskiya, yana kuma ƙoƙarin gyara drift na kowa (stale OpenClaw config, missing TLS trust, restartable services).

Faɗakar da one-shot daga dashboard ta card ɗin **Doctor** ko `wall-vault doctor`.

---

## Hooks

Gudanar da shell command akan key events:

```yaml
hooks:
  on_model_change:   "logger 'wall-vault: $SERVICE/$MODEL'"
  on_key_exhausted:  "notify-send 'wall-vault' '$SERVICE keys all on cooldown'"
  on_service_down:   "/usr/local/bin/page-oncall.sh $SERVICE '$ERROR'"
  on_doctor_fix:     "echo \"$AGENT: $LEVEL $MSG\" >> ~/wall-vault.audit.log"
  openclaw_socket:   ""    # if set, OpenClaw TUI receives events over this Unix socket
```

Kowane hook yana samun event-specific environment variables (`SERVICE`, `MODEL`, `ERROR`, `AGENT`, `LEVEL`, `MSG`). Hooks suna gudana async tare da timeout sakanni 5 — proxy ba ya toshewa a kan hook mai jinkiri.

---

## Manyan masu canji

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
| `WV_PROXY_TLS_REQUIRED` | `proxy.tls.required` (refuse to start with TLS off — fails closed when set) |
| `WV_PROXY_ALLOW_CIDRS` | `proxy.allow_cidrs` (comma-separated list, e.g. `192.168.0.0/16,10.0.0.0/8`; loopback always passes) |
| `WV_VAULT_TLS_ENABLED` | `vault.tls.enabled` |
| `WV_VAULT_TLS_CERT` | `vault.tls.cert_file` |
| `WV_VAULT_TLS_KEY` | `vault.tls.key_file` |
| `WV_VAULT_BOOTSTRAP_PORT` | `vault.bootstrap_port` |
| `WV_PROXY_PLAIN_PORT` | `proxy.plain_port` |
| `WV_KEY_GOOGLE` | One-shot import: comma-separated Google keys |
| `WV_KEY_OPENROUTER` | One-shot import: OpenRouter keys |
| `WV_KEY_ANTHROPIC` | One-shot import: Anthropic keys |
| `WV_KEY_OPENAI` | One-shot import: OpenAI keys |
| `WV_OLLAMA_URL` | Per-host Ollama URL override (single instance) |
| `WV_OLLAMA_URLS` | Comma-separated Ollama URLs (multi-instance dispatch) |
| `WV_OLLAMA_KEEP_ALIVE` | `proxy.ollama_keep_alive` |
| `WV_OLLAMA_NUM_CTX` | `proxy.ollama_num_ctx` |
| `WV_LMSTUDIO_URL`, `WV_VLLM_URL`, `WV_LLAMACPP_URL` | Per-backend URL override (single instance) |
| `WV_TOKEN_SENTINEL_FALLBACK` | `proxy.token_sentinel_fallback` |
| `WV_OAI_STREAM_FORWARD` | `proxy.oai_stream_forward` |
| `WV_ANTHROPIC_FALLBACK_MODEL` | `proxy.anthropic_fallback_model` |
| `WV_INJECT_MODEL_IDENTITY` | `proxy.inject_model_identity` (system-message identity guard, off by default) |
| `WV_PROMPT_TOKEN_CAP` | Per-host auto-truncate cap for local OAI-compat prompts (positive int = enable, 0 = off) |
| `WV_DISPATCH_TRACE` | Set to `1` to log every dispatch's resolved service / model with reason (off by default) |
| `WV_ECONOWORLD_MAX_TOKENS` | `proxy.econoworld_max_tokens` |
| `WV_ECONOWORLD_STREAM` | `proxy.econoworld_stream` |
| `WV_ECONOWORLD_REQUEST_TIMEOUT` | `proxy.econoworld_request_timeout` |

Kowane env var, lokacin da aka saita, yana cin nasara akan YAML file.

---

## Magance matsala

### `connection refused` a `:56244`

Ko dai proxy ba ya gudana ko an ɗaure shi a host dabam. Duba:

```bash
ss -lnp | grep 56244
systemctl --user status wall-vault-proxy   # Linux
launchctl list | grep wall-vault           # macOS
```

Idan yana gudana a port dabam, config ɗinka yana da `proxy.port` da aka rubuta — duba `~/.wall-vault/config.yaml`.

### `x509: certificate signed by unknown authority`

Client baya amincewa da CA na ciki na wall-vault. Gudanar da `wall-vault cert install-trust` a na'urar client. Ga agents waɗanda runtime ɗinsu yake yin watsi da OS trust store (misali Node tare da `NODE_EXTRA_CA_CERTS` mai hardcoded), yi amfani da loopback HTTP companion a `127.0.0.1:56245` (host ɗaya kawai) ko saita `WV_PROXY_TLS_ENABLED=0` don komawa zuwa HTTP fanko.

### `token not registered with vault`

`Authorization: Bearer <token>` na client baya dacewa da kowane client da aka yi rejista. Tabbatar da token a ƙarƙashin **Clients** a dashboard. Idan ka kwafi token literal kamar `proxy-managed`, `dummy`, ko `""` daga config tsoho, maye gurbinsa da real client token.

### `Anthropic dispatch needs a Claude model id`

Hali na tsohon a v0.2.63: model id wanda ba na Claude ba da aka aika zuwa anthropic dispatch yana mayar da error. Ko dai gyara routing (kar a aika `gemini-2.5-flash` zuwa anthropic) ko zaɓa opt in zuwa rubutu atomatik ta `proxy.anthropic_fallback_model`.

### `unknown service: <id>`

Dispatch ya ga service id wanda babu plugin yaml da ya da'awa. Duba:

```bash
ls ~/.wall-vault/services/        # any plugin yaml present?
cat ~/.wall-vault/services/<id>.yaml | grep enabled
```

Idan yaml yana wanzu amma `enabled: false`, juya shi. Idan ya ɓace gaba ɗaya, kwafi daga `configs/services/` a source tree.

### Amsa fanko a kan reasoning model

`qwen3.6`, `deepseek-r1`, da iyalin GPT-`o1` wani lokaci suna fitar da `reasoning_content` kawai kuma su bar `content` fanko. A v0.2.63 wall-vault yana komawa zuwa reasoning text ta atomatik — idan har yanzu kana ganin fanko amsa, backend baya mayar da kowane filin. Duba logs na upstream.

Ga LM Studio tare da qwen3 musamman, saita `inline_no_think_for_qwen3: true` a plugin yaml don a kashe reasoning inline. Built-in lmstudio.yaml da ollama.yaml suna riga sun yi haka.

### Dashboard yana nuna "all keys on cooldown" amma kawai na ƙara ɗaya

Sabuwar key ta yi lafiya amma dispatch path na iya zama har yanzu a cooldown na key mafi tsufa. Gwada sabon request — proxy yana round-robins per call, kuma za a zaɓi key mai lafiya na gaba.

### Vault baya buɗewa da master password

Kalma sirri ba daidai ba. Babu farfaɗowa — wall-vault da gangan baya jigilar backdoor. Idan kun rasa master password da gaske, hanyar kawai ita ce gogewa `~/.wall-vault/data/vault.json`, sake fara da sabon kalma sirri, kuma sake ƙara keys.

### An buga iyakokin OpenRouter na free-tier

Saita `proxy.services` ya haɗa da `openrouter` kuma ƙara aƙalla OpenRouter key ɗaya. Proxy yana auto-falls-back daga model mai biyan kuɗi zuwa `:free` variant ɗinsa lokacin da paid path ya mayar da 402 / 429.

### `journalctl --user -u wall-vault-proxy` ba shi da komai

systemd `--user` logs suna zuwa journal na user ɗin da ke gudanar da shi. Idan ka fara unit a matsayin `root` ko ta `sudo`, journal yana cikin system instance maimakon — gwada `journalctl -u wall-vault-proxy` ba tare da `--user` ba.

---

## Ƙari

- HTTP API reference — duba [API.md](API.md)
- Source — `https://github.com/sookmook/wall-vault`
- Bug reports / feature requests — GitHub Issues
- Tarihin release — [CHANGELOG.md](../CHANGELOG.md)
