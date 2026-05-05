<p align="center">
  <img src="docs/logo.png" alt="wall-vault" width="800">
</p>

# wall-vault

> **Hifadhi ya funguo za API + AI proxy katika faili moja la Go.**
> Huhifadhi funguo ndani ya kifaa kwa AES-GCM, huzizungusha kati ya watoa huduma, hutumia mbadala wakati moja inashindwa, na huja na dashibodi ya wakati halisi.

[![License: GPL v3](https://img.shields.io/badge/License-GPLv3-blue.svg)](LICENSE)
[![Go](https://img.shields.io/badge/Go-1.25-00ADD8.svg)](go.mod)

[English](README.md) · [한국어](README.ko.md) · [中文](README.zh.md) · [日本語](README.ja.md) · [Español](README.es.md) · [Français](README.fr.md) · [Deutsch](README.de.md) · [Português](README.pt.md) · [العربية](README.ar.md) · [हिन्दी](README.hi.md) · [Bahasa Indonesia](README.id.md) · [ภาษาไทย](README.th.md) · **Kiswahili** · [Hausa](README.ha.md) · [नेपाली](README.ne.md) · [Монгол](README.mn.md) · [isiZulu](README.zu.md)

---

## Ni nini

wall-vault hukaa kati ya AI agent (OpenClaw, Claude Code, Cursor, Continue, hati yako mwenyewe) na watoa huduma za AI wa wingu au wa ndani anaowasiliana nao. Mambo mawili katika faili moja:

- **Vault** — huhifadhi funguo za API zilizofichwa zikiwa zimepumzika (AES-GCM ikiwa na nenosiri kuu), huzizungusha, hufuatilia matumizi na nyakati za kupumzika za kila ufunguo, hutangaza mabadiliko kupitia SSE, na hutoa dashibodi ya wavuti kwenye `:56243`.
- **Proxy** — hufichua sehemu zinazolingana na Gemini, Anthropic, na OpenAI kwenye `:56244`, huchagua ufunguo kutoka kwa vault, hutuma kwa upstream uliyoiweka, na hubadilisha kwa mtoa huduma anayefuata wakati moja inashindwa.

Huunga mkono maumbo manne ya maombi (Gemini `:generateContent`, Anthropic `/v1/messages`, OpenAI `/v1/chat/completions`, na Ollama-native `/api/chat`) na makundi matano ya upstream:

| Mtoa Huduma | Maelezo |
|----------|-------|
| Google Gemini | API ya asili; mzunguko wa funguo kwa kila mradi |
| Anthropic | Upitishaji wa asili `/v1/messages` |
| OpenAI | `/v1/chat/completions` ya asili |
| OpenRouter | Modeli 340+, mbadala wa kiotomatiki kwa lahaja za `:free` |
| Ollama / LM Studio / vLLM / llama.cpp / Jan / KoboldCpp / TabbyAPI / mlx-server / LiteLLM | Backend za ndani zinazolingana na OpenAI; sakinisha haraka kupitia plugin yaml |

Kuongeza backend mpya inayolingana na OpenAI ni faili moja la yaml chini ya `~/.wall-vault/services/` — bila kubadilisha msimbo.

## Kwa nini unaweza kuitaka

- Unashughulikia huduma tatu au nne za AI na unataka URL moja ambayo agent inawasiliana nayo.
- Unataka ufunguo wa free-tier uliyo kwenye kupumzika upishe nafasi kwa unaofuata bila kuvunja kipindi.
- Unataka funguo zilezile zilete nguvu kwa bot / IDE / hati nyingi kwenye LAN moja bila kunakili sifa za kuingia.
- Unataka dashibodi, sio environment variables, kwa kuhariri funguo za API.
- Unataka chaguo la kwanza-la-ndani (Ollama / LM Studio) wakati mipaka ya wingu inaisha.

## Anza haraka

### Sakinisha (Linux / macOS)

```bash
curl -fsSL https://raw.githubusercontent.com/sookmook/wall-vault/main/install.sh | sh
```

Au pakua faili lililojengwa moja kwa moja:

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

### Sakinisha (Windows)

```powershell
Invoke-WebRequest -Uri `
  "https://github.com/sookmook/wall-vault/releases/latest/download/wall-vault-windows-amd64.exe" `
  -OutFile wall-vault.exe
```

### Uendeshaji wa kwanza

```bash
wall-vault setup    # mchawi wa mwingiliano — huchagua bandari, huduma, admin token, nenosiri kuu
wall-vault start    # huzindua vault na proxy
```

Fungua `http://localhost:56243` (au `https://...` mara TLS imewashwa — angalia chini) kwenye kivinjari. Dashibodi inaomba admin token iliyochapishwa na `setup`. Kutoka pale unaweza kuongeza funguo za API, kusajili wateja, na kubadilisha modeli bila kuanzisha upya.

---

## TLS (inayopendekezwa)

Kwa chaguo-msingi `wall-vault setup` huandika usanidi bila TLS, kwa hivyo wasikilizaji wote wawili hujibu HTTP rahisi. Mifano ya URL katika README hii hutumia `https://localhost:56244` kwa sababu agents wengi (OpenClaw, Claude Code, Cursor) wanataka sehemu moja iliyo na TLS mbele ambayo haitavunjika ukihamisha proxy kwa host nyingine baadaye. Ili kufanana na mifano hiyo, washa TLS mara moja ukitumia CA ya ndani iliyopo:

```bash
# 1. Tengeneza CA ya ndani ya wall-vault (mara moja, inaishi katika ~/.wall-vault/ca.{crt,key})
wall-vault cert init

# 2. Toa cheti cha host kwa kifaa HIKI
#    SAN zinajumuisha hostname, localhost, 127.0.0.1, na IP yoyote ya LAN iliyogunduliwa
wall-vault cert issue $(hostname)

# 3. Amini CA katika keychain ya OS ya ndani
wall-vault cert install-trust

# 4. Hamisha wasikilizaji kwa TLS
export WV_PROXY_TLS_ENABLED=1
export WV_PROXY_TLS_CERT="$HOME/.wall-vault/$(hostname).crt"
export WV_PROXY_TLS_KEY="$HOME/.wall-vault/$(hostname).key"
export WV_VAULT_TLS_ENABLED=1
export WV_VAULT_TLS_CERT="$HOME/.wall-vault/$(hostname).crt"
export WV_VAULT_TLS_KEY="$HOME/.wall-vault/$(hostname).key"
wall-vault start
```

Kwa kifaa kingine kwenye LAN yako: nakili `~/.wall-vault/ca.crt` na uendeshe `wall-vault cert install-trust --ca <path>` huko. Mara CA inapoaminiwa kila mahali, kila kifaa kwenye mtandao kinaweza kufikia proxy kupitia `https://<host>:56244` bila maonyo ya cheti.

Ikiwa unapendelea kubaki kwenye HTTP rahisi, acha usanidi kama ulivyo na ubadilishe `https://` na `http://` katika vipande vya client hapo chini. Mipangilio yote miwili inafanya kazi; tofauti ni bandari ipi inajibu TLS handshake.

**Mbadala wa Loopback.** Wateja walio kwenye host moja ambao hawawezi kuheshimu CA ya wall-vault (haswa Node runtime ya OpenClaw iliyopo, ambayo huandika upya `NODE_EXTRA_CA_CERTS` wakati wa spawn) hufikia proxy kupitia mwenza wa HTTP-rahisi wa loopback-pekee kwenye `127.0.0.1:56245`. wall-vault huiwasha kiotomatiki wakati TLS imewashwa.

---

## Kuunganisha wateja

Elekeza AI client yoyote kwa `https://<host>:56244` (au `http://...` ikiwa TLS imezimwa). Proxy hujibu maumbo manne:

| Umbo | Njia | Mifano ya wateja |
|--------|------|-----------------|
| Gemini | `/google/v1beta/models/<model>:generateContent` | OpenClaw, Gemini CLI, Antigravity |
| Anthropic | `/v1/messages` | Claude Code, Anthropic SDKs |
| OpenAI | `/v1/chat/completions` | Cursor, Continue, hati za kibinafsi, programu nyingi za LLM |
| Ollama-native | `/api/chat` | Wateja wa Ollama wakipitia |

### Claude Code

```bash
export ANTHROPIC_BASE_URL=https://localhost:56244
export ANTHROPIC_API_KEY=<your-vault-client-token>
claude
```

Wakati mikopo ya upstream ya Anthropic inaisha, utumaji hubadilika kwa watoa huduma wowote uliowekwa katika `fallback_services` kwa client huyu. Ili kuchagua mbadala usio wa Claude waziwazi:

```yaml
proxy:
  anthropic_fallback_model: "claude-haiku-4-5-20251001"
```

(Chaguo-msingi tupu hufanya utumaji urudishe kosa ili upotofu wa njia uonekane mara moja.)

### Cursor / Continue

Katika Cursor **Settings → AI → OpenAI API**:

```
Base URL:  https://localhost:56244
API Key:   <your-vault-client-token>
Model:     gemini-2.5-flash    # au modeli yoyote ambayo wall-vault inaijua
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

OpenClaw ni mfumo wa agent wa TUI ambao wall-vault iliundwa awali kuhudumia. Modal ya **Add Agent** ya dashibodi huweka aina ya agent kuwa `openclaw` (au `nanoclaw`); kisha wall-vault huandika `~/.openclaw/openclaw.json` moja kwa moja, ikijumuisha URL za watoa huduma, vault token, na vipengele vya modeli:

```bash
VAULT_CLIENT_ID=my-bot \
VAULT_URL=http://<vault-host>:56243 \
VAULT_TOKEN=<your-client-token> \
wall-vault proxy
```

### curl / hati

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

## Usanidi

`wall-vault setup` huandika ama `./wall-vault.yaml` au `~/.wall-vault/config.yaml`. Hariri kwa mkono kwa sehemu ambazo mchawi hauulizi.

```yaml
mode: standalone     # standalone | distributed
lang: en             # en | ko | zh | ja | es | fr | de | pt | ar | hi | id | th | sw | ha | ne | mn | zu
theme: light         # light | dark | cherry | ocean | gold | autumn | winter

proxy:
  port: 56244
  host: ""           # chaguo-msingi: 127.0.0.1 standalone, 0.0.0.0 distributed
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
  plain_port: 56245              # mwenza wa HTTP wa loopback-pekee wakati TLS imewashwa
  ollama_keep_alive: "30m"       # "-1" usipakue kamwe, "0" pakua mara moja
  ollama_num_ctx: 8192
  oai_stream_forward: false      # ushiriki wa hiari wa SSE wa backend halisi
  anthropic_fallback_model: ""   # chaguo la kuandika upya kisicho cha Claude kwenye anthropic dispatch

vault:
  port: 56243
  host: ""
  admin_token: ""
  admin_ip_whitelist: []
  master_password: ""            # nenosiri la usimbaji fiche la AES-GCM
  data_dir: ~/.wall-vault/data
  services_dir: ~/.wall-vault/services
  tls:
    enabled: false
    cert_file: ""
    key_file: ""
  bootstrap_port: 56247          # msikilizaji wa HTTP rahisi anayehudumia ca.crt tu

doctor:
  interval: 5m
  auto_fix: true
  log_file: /tmp/wall-vault-doctor.log

hooks:
  on_model_change: ""            # amri ya shell (env: SERVICE, MODEL)
  on_key_exhausted: ""
  on_service_down: ""
  on_doctor_fix: ""
  openclaw_socket: ""
```

### Environment variables

Kila sehemu ya YAML ina env override inayoshinda faili. Zile za kawaida:

| Kibadala | Maelezo |
|----------|-------------|
| `WV_LANG`, `WV_THEME` | Lugha na mandhari |
| `WV_PROXY_PORT`, `WV_PROXY_HOST` | Anwani ya kusikiliza ya Proxy |
| `WV_VAULT_PORT`, `WV_VAULT_HOST` | Anwani ya kusikiliza ya Vault |
| `WV_VAULT_URL`, `WV_VAULT_TOKEN` | Sehemu za hali ya distributed |
| `WV_ADMIN_TOKEN`, `WV_MASTER_PASS` | Sifa za Vault |
| `WV_KEY_GOOGLE`, `WV_KEY_OPENROUTER`, `WV_KEY_ANTHROPIC`, `WV_KEY_OPENAI` | Funguo za API (zilizotenganishwa kwa koma kwa nyingi) |
| `WV_PROXY_TLS_ENABLED`, `WV_PROXY_TLS_CERT`, `WV_PROXY_TLS_KEY` | Proxy TLS |
| `WV_VAULT_TLS_ENABLED`, `WV_VAULT_TLS_CERT`, `WV_VAULT_TLS_KEY` | Vault TLS |
| `WV_PROXY_PLAIN_PORT` | Mwenza wa HTTP wa loopback (`0` kuzima) |
| `WV_VAULT_BOOTSTRAP_PORT` | Msikilizaji wa CA bootstrap (`0` kuzima) |
| `WV_OLLAMA_URL`, `WV_OLLAMA_KEEP_ALIVE`, `WV_OLLAMA_NUM_CTX` | Marekebisho ya Ollama |
| `WV_LMSTUDIO_URL`, `WV_VLLM_URL`, `WV_LLAMACPP_URL` | Mabadiliko ya backend ya ndani |
| `WV_TOKEN_SENTINEL_FALLBACK` | Mbadala wa sentinel "proxy-managed" wa loopback |
| `WV_OAI_STREAM_FORWARD` | Upitishaji wa SSE wa backend halisi inayolingana na OpenAI |
| `WV_ANTHROPIC_FALLBACK_MODEL` | Chaguo la kuandika upya kisicho cha Claude kwenye anthropic |

---

## Hali

### Standalone (chaguo-msingi)

Vault na proxy huendesha katika mchakato sawa. Bora kwa host moja inayohifadhi funguo na agent. Husikiliza tu kwenye loopback kwa chaguo-msingi.

```bash
wall-vault start    # huendesha zote mbili
```

### Distributed

Vault huendesha kwenye host moja (**vault host**) na huhifadhi funguo zote; proxies nyingi kwenye hosts nyingine kila moja inathibitisha kwa token ya kila client. Inafaa wakati mashine kadhaa zinahitaji funguo zilezile bila kuzinakili.

**Vault host:**

```bash
WV_VAULT_HOST=0.0.0.0 wall-vault vault
```

**Kila proxy host:**

```bash
WV_VAULT_URL=http://<vault-host>:56243 \
WV_VAULT_TOKEN=<this-client-token> \
WV_PROXY_HOST=0.0.0.0 \
wall-vault proxy
```

Modal ya **Add Client** ya dashibodi hutengeneza token, husajili aina ya agent, na proxy huchukua usanidi wake kupitia SSE bila kuanza upya.

---

## Plugin yaml (backend ya kuingiza)

Backend yoyote inayolingana na OpenAI inaweza kuongezwa kama yaml chini ya `~/.wall-vault/services/`. wall-vault huichukua wakati wa kuanza, huisajili kama huduma inayoweza kuelekezwa, na utumaji + seti ya kugundua OAI-compat + daraja la stream la Gemini zote zinaiona bila mabadiliko ya msimbo.

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
inline_no_think_for_qwen3: false   # chagua ikiwa backend yako huondoa marker
```

Topolojia ya Hub (wall-vault moja inakaa mbele ya nyingine) inaungwa mkono kupitia `tls_internal_ca: true`, `auth.type: bearer`, na `preserve_model_id: true`.

---

## Jenga kutoka kwa chanzo

```bash
git clone https://github.com/sookmook/wall-vault
cd wall-vault
go build -o wall-vault .
./wall-vault setup
./wall-vault start
```

Cross-compile kwa seti yote inayoungwa mkono:

```bash
make build-all   # linux-amd64, linux-arm64, darwin-amd64, darwin-arm64, windows-amd64
```

Matoleo yanafuata `v{major}.{minor}.{patch}.{YYYYMMDD}.{HHmmss}`; `BASE_VERSION` katika Makefile huweka kiambishi awali.

### Mpangilio wa mradi

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

## Nyaraka

- [Mwongozo wa mtumiaji](docs/MANUAL.en.md) — usakinishaji, dashibodi, agents, utatuzi
- [API reference](docs/API.en.md) — kila endpoint na maumbo ya request/response
- [CHANGELOG](CHANGELOG.md)

---

## Mfumo wa kiufundi

- Go 1.25, faili moja la statiki
- [templ](https://templ.guide) kwa dashibodi inayotolewa upande wa server, [HTMX](https://htmx.org) kwa masasisho ya sehemu
- AES-GCM (ufunguo unaotokana na PBKDF2) kwa usimbaji fiche wa funguo zilizopumzika
- Server-Sent Events kwa usawazishaji wa moja kwa moja wa usanidi kati ya vault na proxies
- CA ya ndani iliyojisaini + vyeti vya kila host (hakuna DNS ya umma / Let's Encrypt inayohitajika)

## Leseni

GPL-3.0. Angalia [LICENSE](LICENSE).

## Kuchangia

Pull requests zinakaribishwa. Angalia [CONTRIBUTING.md](CONTRIBUTING.md). Kwa mabadiliko makubwa tafadhali fungua issue kwanza ili kujadili muundo.
