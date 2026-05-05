<p align="center">
  <img src="docs/logo.png" alt="wall-vault" width="200">
</p>

# wall-vault

> **Ma'aji na maɓallin API + AI proxy a cikin binary ɗaya na Go.**
> Yana adana maɓallai a cikin na'ura tare da AES-GCM, yana juya su tsakanin masu samar da ayyuka, yana juyawa lokacin da ɗaya ya gaza, kuma yana zuwa da dashboard na ainihin lokaci.

[![License: GPL v3](https://img.shields.io/badge/License-GPLv3-blue.svg)](LICENSE)
[![Go](https://img.shields.io/badge/Go-1.25-00ADD8.svg)](go.mod)

[English](README.md) · [한국어](README.ko.md) · [中文](README.zh.md) · [日本語](README.ja.md) · [Español](README.es.md) · [Français](README.fr.md) · [Deutsch](README.de.md) · [Português](README.pt.md) · [العربية](README.ar.md) · [हिन्दी](README.hi.md) · [Bahasa Indonesia](README.id.md) · [ภาษาไทย](README.th.md) · [Kiswahili](README.sw.md) · **Hausa** · [नेपाली](README.ne.md) · [Монгол](README.mn.md) · [isiZulu](README.zu.md)

---

## Menene shi

wall-vault yana zaune tsakanin AI agent (OpenClaw, Claude Code, Cursor, Continue, naka rubutaccen rubutu) da masu samar da AI na cloud ko na cikin gida da yake magana da su. Abubuwa biyu a cikin binary ɗaya:

- **Vault** — yana adana maɓallin API da aka boye yayin hutawa (AES-GCM tare da babban kalmar wucewa), yana juya su, yana bin diddigin amfani da hutu na kowane maɓalli, yana watsa canje-canje ta SSE, kuma yana ba da dashboard na yanar gizo a `:56243`.
- **Proxy** — yana fallasa wuraren da suka dace da Gemini, Anthropic, da OpenAI a `:56244`, yana zaɓen maɓalli daga vault, yana aikawa zuwa upstream da ka saita, kuma yana juyawa zuwa mai ba da sabis na gaba lokacin da ɗaya ya gaza.

Yana goyon bayan siffofi huɗu na buƙata (Gemini `:generateContent`, Anthropic `/v1/messages`, OpenAI `/v1/chat/completions`, da Ollama-native `/api/chat`) da nau'ikan upstream guda biyar:

| Mai Ba da Sabis | Bayanan kula |
|----------|-------|
| Google Gemini | API na asali; juyawar maɓalli kowane aikin |
| Anthropic | Wucewa ta asali `/v1/messages` |
| OpenAI | `/v1/chat/completions` na asali |
| OpenRouter | Samfura 340+, juyawa ta atomatik zuwa ire-iren `:free` |
| Ollama / LM Studio / vLLM / llama.cpp / Jan / KoboldCpp / TabbyAPI / mlx-server / LiteLLM | Backend na cikin gida masu jituwa da OpenAI; saka cikin sauƙi ta plugin yaml |

Ƙara sabon backend mai jituwa da OpenAI shine fayil ɗin yaml ɗaya a ƙarƙashin `~/.wall-vault/services/` — babu canjin code.

## Me yasa kuke iya so shi

- Kuna sarrafa ayyuka uku ko huɗu na AI kuma kuna son URL ɗaya da agent yake magana da shi.
- Kuna son maɓallin free-tier akan hutu ya bayar da hanya ga na gaba ba tare da karya zama ba.
- Kuna son maɓallai iri ɗaya su ƙarfafa bots / IDEs / scripts da yawa akan LAN ɗaya ba tare da kwafin sirrin shaida ba.
- Kuna son dashboard, ba environment variables ba, don gyara maɓallin API.
- Kuna son zaɓi na farko-na-cikin-gida (Ollama / LM Studio) lokacin da iyakokin cloud sun gushe.

## Farawa cikin sauri

### Sakawa (Linux / macOS)

```bash
curl -fsSL https://raw.githubusercontent.com/sookmook/wall-vault/main/install.sh | sh
```

Ko zazzage binary da aka riga aka gina kai tsaye:

```bash
# Linux amd64
curl -L https://github.com/sookmook/wall-vault/releases/latest/download/wall-vault-linux-amd64 \
  -o wall-vault && chmod +x wall-vault

# macOS Apple Silicon
curl -L https://github.com/sookmook/wall-vault/releases/latest/download/wall-vault-darwin-arm64 \
  -o wall-vault && chmod +x wall-vault

# Linux arm64 (Raspberry Pi, ARM servers)
curl -L https://github.com/sookmook/wall-vault/releases/latest/download/wall-vault-linux-arm64 \
  -o wall-vault && chmod +x wall-vault
```

### Sakawa (Windows)

```powershell
Invoke-WebRequest -Uri `
  "https://github.com/sookmook/wall-vault/releases/latest/download/wall-vault-windows-amd64.exe" `
  -OutFile wall-vault.exe
```

### Gudanarwa ta farko

```bash
wall-vault setup    # mai shawara mai hulɗa — yana zaɓen tashar jiragen ruwa, ayyuka, admin token, babban kalmar wucewa
wall-vault start    # yana ƙaddamar da vault da proxy
```

Buɗe `http://localhost:56243` (ko `https://...` da zarar an kunna TLS — duba ƙasa) a cikin browser. Dashboard yana neman admin token da `setup` ya buga. Daga can kuna ƙara maɓallin API, yin rajistar abokan ciniki, da canza samfura ba tare da sake farawa ba.

---

## TLS (an ba da shawarar)

Ta tsohuwa `wall-vault setup` yana rubuta saiti ba tare da TLS ba, don haka masu sauraro biyu suna amsawa da HTTP a fili. URLs na misali a cikin wannan README suna amfani da `https://localhost:56244` saboda yawancin agents (OpenClaw, Claude Code, Cursor) suna son endpoint guda ɗaya da TLS a gaba wanda ba zai karya ba idan ka motsa proxy zuwa wani host daga baya. Don dacewa da waɗancan misalai, kunna TLS sau ɗaya tare da CA na ciki da aka haɗa:

```bash
# 1. Ƙirƙira CA na ciki na wall-vault (sau ɗaya, yana zaune a ~/.wall-vault/ca.{crt,key})
wall-vault cert init

# 2. Bayar da takardar shaidar host don na'ura WANNAN
#    SAN sun haɗa da hostname, localhost, 127.0.0.1, da kowane LAN IP da aka gano
wall-vault cert issue $(hostname)

# 3. Amince da CA a cikin keychain na OS na cikin gida
wall-vault cert install-trust

# 4. Sauya masu sauraro zuwa TLS
export WV_PROXY_TLS_ENABLED=1
export WV_PROXY_TLS_CERT="$HOME/.wall-vault/$(hostname).crt"
export WV_PROXY_TLS_KEY="$HOME/.wall-vault/$(hostname).key"
export WV_VAULT_TLS_ENABLED=1
export WV_VAULT_TLS_CERT="$HOME/.wall-vault/$(hostname).crt"
export WV_VAULT_TLS_KEY="$HOME/.wall-vault/$(hostname).key"
wall-vault start
```

Don wata na'ura akan LAN ɗinka: kwafa `~/.wall-vault/ca.crt` zuwa can kuma gudanar da `wall-vault cert install-trust --ca <path>` a wurin. Da zarar an amince da CA a ko'ina, kowane na'ura akan hanyar sadarwa zai iya isa proxy ta `https://<host>:56244` ba tare da gargaɗin takardar shaida ba.

Idan ka gwammace ka kasance a kan HTTP a fili, bar saiti kamar yadda yake kuma maye gurbin `https://` da `http://` a cikin sassan client a ƙasa. Dukkanin tsare-tsaren suna aiki; bambance-bambancen shine wace tashar jiragen ruwa ce ke amsa TLS handshake.

**Loopback fallback.** Abokan ciniki masu zaune a host iri ɗaya waɗanda ba za su iya girmama CA na wall-vault ba (musamman runtime na Node da aka haɗa da OpenClaw, wanda yake sake rubuta `NODE_EXTRA_CA_CERTS` lokacin spawn) suna isa proxy ta abokin tafiya na HTTP a fili kawai-loopback a `127.0.0.1:56245`. wall-vault yana kunna shi ta atomatik lokacin da TLS yake aiki.

---

## Haɗa abokan ciniki

Nuna kowane AI client zuwa `https://<host>:56244` (ko `http://...` idan TLS yana kashe). Proxy yana amsa siffofi huɗu:

| Tsari | Hanya | Misalan abokan ciniki |
|--------|------|-----------------|
| Gemini | `/google/v1beta/models/<model>:generateContent` | OpenClaw, Gemini CLI, Antigravity |
| Anthropic | `/v1/messages` | Claude Code, Anthropic SDKs |
| OpenAI | `/v1/chat/completions` | Cursor, Continue, custom scripts, yawancin LLM apps |
| Ollama-native | `/api/chat` | Abokan ciniki na Ollama suna wucewa |

### Claude Code

```bash
export ANTHROPIC_BASE_URL=https://localhost:56244
export ANTHROPIC_API_KEY=<your-vault-client-token>
claude
```

Lokacin da credit ɗin upstream Anthropic ya gushe, dispatch yana juya zuwa ga masu samar da sabis duk inda ka saita su a `fallback_services` don wannan client. Don zaɓen juyawa wadda ba ta Claude ba a fili:

```yaml
proxy:
  anthropic_fallback_model: "claude-haiku-4-5-20251001"
```

(Tsohuwar shi babu komai yana sa dispatch ya dawo da kuskure ta yadda misrouting ya bayyana nan take.)

### Cursor / Continue

A cikin Cursor **Settings → AI → OpenAI API**:

```
Base URL:  https://localhost:56244
API Key:   <your-vault-client-token>
Model:     gemini-2.5-flash    # ko kowane samfurin da wall-vault ya san
```

Continue (`config.json`):

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

### OpenClaw

OpenClaw shine tsarin agent na TUI wanda aka gina wall-vault don bayar da sabis a asali. Modal ɗin **Add Agent** na dashboard yana saita nau'in agent zuwa `openclaw` (ko `nanoclaw`); sannan wall-vault yana rubuta `~/.openclaw/openclaw.json` kai tsaye, gami da URLs na masu samar da sabis, vault token, da shigarwar samfurin:

```bash
VAULT_CLIENT_ID=my-bot \
VAULT_URL=http://<vault-host>:56243 \
VAULT_TOKEN=<your-client-token> \
wall-vault proxy
```

### curl / scripts

```bash
curl -X POST https://localhost:56244/v1/chat/completions \
  -H "Authorization: Bearer <your-vault-client-token>" \
  -H "Content-Type: application/json" \
  -d '{
    "model": "gemini-2.5-flash",
    "messages": [{"role": "user", "content": "ping"}]
  }'
```

---

## Saiti

`wall-vault setup` yana rubuta ko `./wall-vault.yaml` ko `~/.wall-vault/config.yaml`. Gyara da hannu don filaye da mai shawara baya tambaya akai.

```yaml
mode: standalone     # standalone | distributed
lang: en             # en | ko | zh | ja | es | fr | de | pt | ar | hi | id | th | sw | ha | ne | mn | zu
theme: light         # light | dark | cherry | ocean | gold | autumn | winter

proxy:
  port: 56244
  host: ""           # tsoho: 127.0.0.1 standalone, 0.0.0.0 distributed
  client_id: my-bot
  vault_url: ""      # distributed: http://vault-host:56243
  vault_token: ""    # distributed: client token
  tool_filter: strip_all   # strip_all | whitelist | passthrough
  allowed_tools: []
  services: [google, openrouter, ollama]
  timeout: 300s
  tls:
    enabled: false
    cert_file: ""
    key_file: ""
  plain_port: 56245              # abokin tafiya na HTTP na loopback-kawai lokacin da TLS yake aiki
  ollama_keep_alive: "30m"       # "-1" kar a sauke, "0" sauke nan take
  ollama_num_ctx: 8192
  oai_stream_forward: false      # zaɓi na ainihin backend SSE passthrough
  anthropic_fallback_model: ""   # zaɓi na sake rubutawa wadda ba ta Claude ba a anthropic dispatch

vault:
  port: 56243
  host: ""
  admin_token: ""
  admin_ip_whitelist: []
  master_password: ""            # kalmar wucewa don boye maɓalli na AES-GCM
  data_dir: ~/.wall-vault/data
  services_dir: ~/.wall-vault/services
  tls:
    enabled: false
    cert_file: ""
    key_file: ""
  bootstrap_port: 56247          # mai sauraron HTTP a fili wanda ke ba da ca.crt kawai

doctor:
  interval: 5m
  auto_fix: true
  log_file: /tmp/wall-vault-doctor.log

hooks:
  on_model_change: ""            # umarnin shell (env: SERVICE, MODEL)
  on_key_exhausted: ""
  on_service_down: ""
  on_doctor_fix: ""
  openclaw_socket: ""
```

### Environment variables

Kowane filin YAML yana da env override wanda ke nasara akan fayil. Wadanda aka saba:

| Mai canzawa | Bayani |
|----------|-------------|
| `WV_LANG`, `WV_THEME` | Harshe da jigon |
| `WV_PROXY_PORT`, `WV_PROXY_HOST` | Adireshin sauraro na Proxy |
| `WV_VAULT_PORT`, `WV_VAULT_HOST` | Adireshin sauraro na Vault |
| `WV_VAULT_URL`, `WV_VAULT_TOKEN` | Endpoints na yanayin distributed |
| `WV_ADMIN_TOKEN`, `WV_MASTER_PASS` | Sirrin shaida na Vault |
| `WV_KEY_GOOGLE`, `WV_KEY_OPENROUTER`, `WV_KEY_ANTHROPIC`, `WV_KEY_OPENAI` | Maɓallin API (rabuwa da waƙafi don da yawa) |
| `WV_PROXY_TLS_ENABLED`, `WV_PROXY_TLS_CERT`, `WV_PROXY_TLS_KEY` | Proxy TLS |
| `WV_VAULT_TLS_ENABLED`, `WV_VAULT_TLS_CERT`, `WV_VAULT_TLS_KEY` | Vault TLS |
| `WV_PROXY_PLAIN_PORT` | Abokin tafiya na HTTP na loopback (`0` don kashe) |
| `WV_VAULT_BOOTSTRAP_PORT` | Mai sauraro na CA bootstrap (`0` don kashe) |
| `WV_OLLAMA_URL`, `WV_OLLAMA_KEEP_ALIVE`, `WV_OLLAMA_NUM_CTX` | Daidaita Ollama |
| `WV_LMSTUDIO_URL`, `WV_VLLM_URL`, `WV_LLAMACPP_URL` | Sauya backend na cikin gida |
| `WV_TOKEN_SENTINEL_FALLBACK` | Maye gurbin sentinel "proxy-managed" na loopback |
| `WV_OAI_STREAM_FORWARD` | Wucewar SSE na ainihin backend mai jituwa da OpenAI |
| `WV_ANTHROPIC_FALLBACK_MODEL` | Zaɓi na sake rubutawa wadda ba ta Claude ba a anthropic |

---

## Yanayi

### Standalone (tsoho)

Vault da proxy suna aiki a cikin tsari ɗaya. Mafi kyau don host ɗaya wanda ke ɗaukar maɓallai da agent. Yana sauraro kawai akan loopback ta tsohuwa.

```bash
wall-vault start    # yana gudana duka biyu
```

### Distributed

Vault yana aiki akan host ɗaya (**vault host**) kuma yana adana dukkan maɓallai; proxies da yawa a kan wasu hosts kowannensu yana tantance kansa da token kowane-client. Yana da amfani lokacin da na'urori da yawa suna buƙatar maɓallai iri ɗaya ba tare da kwafa su ba.

**Vault host:**

```bash
WV_VAULT_HOST=0.0.0.0 wall-vault vault
```

**Kowane proxy host:**

```bash
WV_VAULT_URL=http://<vault-host>:56243 \
WV_VAULT_TOKEN=<this-client-token> \
WV_PROXY_HOST=0.0.0.0 \
wall-vault proxy
```

Modal ɗin **Add Client** na dashboard yana ƙirƙirar token, yana yin rajistar nau'in agent, kuma proxy yana ɗaukar saitin sa ta SSE ba tare da sake farawa ba.

---

## Plugin yaml (drop-in backend)

Kowane backend mai jituwa da OpenAI ana iya ƙarawa azaman yaml a ƙarƙashin `~/.wall-vault/services/`. wall-vault yana ɗaukar shi a farko, yana yin rajistarsa a matsayin sabis mai juyawa, kuma dispatch + saitin gano OAI-compat + gadar Gemini-stream duk suna ganin sa ba tare da canjin code ba.

```yaml
# ~/.wall-vault/services/llamacpp.yaml
id: llamacpp
name: llama.cpp
enabled: true
default_url: http://localhost:8080
endpoints:
  generate: /v1/chat/completions
  list_models: /v1/models
auth:
  type: none
request_format: openai
inline_no_think_for_qwen3: false   # zaɓi idan backend ɗinka yana cire alama
```

Hub topology (wall-vault ɗaya yana gaban wani) ana goyon bayan ta `tls_internal_ca: true`, `auth.type: bearer`, da `preserve_model_id: true`.

---

## Gina daga tushe

```bash
git clone https://github.com/sookmook/wall-vault
cd wall-vault
go build -o wall-vault .
./wall-vault setup
./wall-vault start
```

Cross-compile don dukkanin saitin da ake goyan baya:

```bash
make build-all   # linux-amd64, linux-arm64, darwin-amd64, darwin-arm64, windows-amd64
```

Sigogi suna bin `v{major}.{minor}.{patch}.{YYYYMMDD}.{HHmmss}`; `BASE_VERSION` a cikin Makefile yana saita prefix.

### Tsarin aikin

```
wall-vault/
├── main.go                     # CLI dispatch (start/proxy/vault/setup/cert/doctor)
├── cmd/
│   ├── setup/                  # interactive setup wizard
│   └── cert/                   # internal CA + per-host TLS certificate issuer
├── internal/
│   ├── config/                 # YAML + env loader, plugin loader
│   ├── proxy/                  # request dispatch, key rotation, format converters
│   ├── vault/                  # AES-GCM store, dashboard, SSE broker
│   ├── doctor/                 # health probe + auto-fix
│   ├── hooks/                  # shell-command event triggers
│   └── i18n/                   # 17-language UI strings
├── configs/services/           # bundled plugin yamls (lmstudio, vllm, ollama, …)
└── docs/                       # MANUAL, API reference, 16 locale variants
```

---

## Takardun shaida

- [Manual na mai amfani](docs/MANUAL.en.md) — sakawa, dashboard, agents, gyara matsala
- [API reference](docs/API.en.md) — kowane endpoint tare da siffofin request/response
- [CHANGELOG](CHANGELOG.md)

---

## Tsarin fasaha

- Go 1.25, binary statik ɗaya
- [templ](https://templ.guide) don dashboard da aka bada a server, [HTMX](https://htmx.org) don sabuntawa na ɓangare
- AES-GCM (maɓalli wanda PBKDF2 ya samar) don boye maɓallin a hutu
- Server-Sent Events don sync na saiti kai tsaye tsakanin vault da proxies
- CA na ciki mai sa hannu — kai + takaddun shaida kowane-host (babu DNS na jama'a / Let's Encrypt da ake buƙata)

## Lasisi

GPL-3.0. Duba [LICENSE](LICENSE).

## Bayar da gudunmawa

Pull requests ana maraba da su. Duba [CONTRIBUTING.md](CONTRIBUTING.md). Don manyan canje-canje, da fatan a buɗe issue da farko don tattaunawa kan ƙira.
