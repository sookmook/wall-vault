# wall-vault

> **एक ही Go बाइनरी में API key vault + AI proxy।**
> कुंजियों को स्थानीय रूप से AES-GCM से संग्रहीत करता है, उन्हें providers के बीच रोटेट करता है, किसी एक के विफल होने पर fallback करता है, और real-time dashboard के साथ आता है।

[![License: GPL v3](https://img.shields.io/badge/License-GPLv3-blue.svg)](LICENSE)
[![Go](https://img.shields.io/badge/Go-1.25-00ADD8.svg)](go.mod)

[English](README.md) · [한국어](README.ko.md) · [中文](README.zh.md) · [日本語](README.ja.md) · [Español](README.es.md) · [Français](README.fr.md) · [Deutsch](README.de.md) · [Português](README.pt.md) · [العربية](README.ar.md) · **हिन्दी** · [Bahasa Indonesia](README.id.md) · [ภาษาไทย](README.th.md) · [Kiswahili](README.sw.md) · [Hausa](README.ha.md) · [नेपाली](README.ne.md) · [Монгол](README.mn.md) · [isiZulu](README.zu.md)

---

## यह क्या है

wall-vault एक AI agent (OpenClaw, Claude Code, Cursor, Continue, या आपका अपना script) और cloud या local AI providers के बीच बैठता है जिनसे वह संवाद करता है। एक बाइनरी में दो चीज़ें:

- **Vault** — API keys को विश्राम पर encrypted (मास्टर पासवर्ड के साथ AES-GCM) संग्रहीत करता है, उन्हें रोटेट करता है, प्रति-key उपयोग और cooldowns को ट्रैक करता है, परिवर्तनों को SSE पर broadcast करता है, और `:56243` पर एक web dashboard सर्व करता है।
- **Proxy** — `:56244` पर Gemini, Anthropic, और OpenAI-compatible endpoints को expose करता है, vault से एक key चुनता है, आपके configure किए गए upstream को dispatch करता है, और किसी एक के विफल होने पर अगले provider पर fallback करता है।

यह चार request आकारों (Gemini `:generateContent`, Anthropic `/v1/messages`, OpenAI `/v1/chat/completions`, और Ollama-native `/api/chat`) और upstream की पाँच श्रेणियों का समर्थन करता है:

| Provider | टिप्पणियाँ |
|----------|-------|
| Google Gemini | Native API; प्रति project key rotation |
| Anthropic | Native `/v1/messages` passthrough |
| OpenAI | Native `/v1/chat/completions` |
| OpenRouter | 340+ models, `:free` variants पर auto-fallback |
| Ollama / LM Studio / vLLM / llama.cpp / Jan / KoboldCpp / TabbyAPI / mlx-server / LiteLLM | OpenAI-compat local backends; plugin yaml के माध्यम से drop-in |

एक नया OpenAI-compatible backend जोड़ना `~/.wall-vault/services/` के अंतर्गत एक yaml file है — कोई कोड परिवर्तन नहीं।

## आप इसे क्यों चाह सकते हैं

- आप तीन या चार AI सेवाओं को संभाल रहे हैं और एक URL चाहते हैं जिससे agent बात करे।
- आप चाहते हैं कि cooldown पर एक free-tier key अगले के लिए रास्ता बना दे बिना session को तोड़े।
- आप चाहते हैं कि वही keys एक ही LAN पर कई bots / IDEs / scripts को बिना credentials कॉपी किए बिना संचालित करें।
- आप API keys संपादित करने के लिए environment variables नहीं, बल्कि एक dashboard चाहते हैं।
- जब cloud limits समाप्त हो जाते हैं तो आप एक local-first विकल्प (Ollama / LM Studio) चाहते हैं।

## त्वरित प्रारंभ

### Install (Linux / macOS)

```bash
curl -fsSL https://raw.githubusercontent.com/sookmook/wall-vault/main/install.sh | sh
```

या सीधे एक pre-built binary डाउनलोड करें:

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

### Install (Windows)

```powershell
Invoke-WebRequest -Uri `
  "https://github.com/sookmook/wall-vault/releases/latest/download/wall-vault-windows-amd64.exe" `
  -OutFile wall-vault.exe
```

### पहली बार चलाना

```bash
wall-vault setup    # interactive wizard — port, services, admin token, master password चुनता है
wall-vault start    # vault और proxy दोनों launch करता है
```

ब्राउज़र में `http://localhost:56243` (या TLS चालू होने पर `https://...` — नीचे देखें) खोलें। Dashboard `setup` द्वारा printed admin token के लिए prompt करता है। वहाँ से आप API keys जोड़ते हैं, clients पंजीकृत करते हैं, और बिना restart किए models बदलते हैं।

---

## TLS (अनुशंसित)

डिफ़ॉल्ट रूप से `wall-vault setup` बिना TLS के config लिखता है, इसलिए दोनों listeners plain HTTP का जवाब देते हैं। इस README में उदाहरण URLs `https://localhost:56244` का उपयोग करते हैं क्योंकि अधिकांश agents (OpenClaw, Claude Code, Cursor) एक ही TLS-fronted endpoint चाहते हैं जो बाद में proxy को किसी दूसरे host पर ले जाने पर भी न टूटे। उन उदाहरणों से मेल खाने के लिए, बंडल किए गए internal CA के साथ TLS को एक बार enable करें:

```bash
# 1. wall-vault internal CA बनाएँ (एक बार, ~/.wall-vault/ca.{crt,key} में रहता है)
wall-vault cert init

# 2. इस machine के लिए एक host certificate जारी करें
#    SANs में hostname, localhost, 127.0.0.1, और कोई भी detected LAN IP शामिल हैं
wall-vault cert issue $(hostname)

# 3. local OS keychain में CA पर भरोसा करें
wall-vault cert install-trust

# 4. listeners को TLS पर switch करें
export WV_PROXY_TLS_ENABLED=1
export WV_PROXY_TLS_CERT="$HOME/.wall-vault/$(hostname).crt"
export WV_PROXY_TLS_KEY="$HOME/.wall-vault/$(hostname).key"
export WV_VAULT_TLS_ENABLED=1
export WV_VAULT_TLS_CERT="$HOME/.wall-vault/$(hostname).crt"
export WV_VAULT_TLS_KEY="$HOME/.wall-vault/$(hostname).key"
wall-vault start
```

अपने LAN पर किसी अन्य machine के लिए: `~/.wall-vault/ca.crt` कॉपी करें और वहाँ `wall-vault cert install-trust --ca <path>` चलाएँ। एक बार CA हर जगह trusted हो जाने पर, network पर हर machine certificate warnings के बिना `https://<host>:56244` पर proxy तक पहुँच सकती है।

यदि आप plain HTTP पर रहना पसंद करते हैं, तो config को वैसा ही छोड़ दें और नीचे client snippets में `https://` को `http://` से बदलें। दोनों schemes काम करते हैं; अंतर यह है कि कौन-सा port TLS handshake का जवाब देता है।

**Loopback fallback.** Same-host clients जो wall-vault CA का सम्मान नहीं कर सकते (विशेष रूप से OpenClaw का बंडल किया गया Node runtime, जो spawn पर `NODE_EXTRA_CA_CERTS` को rewrite करता है) `127.0.0.1:56245` पर एक loopback-only plain-HTTP companion के माध्यम से proxy तक पहुँचते हैं। TLS चालू होने पर wall-vault इसे automatically enable करता है।

---

## Clients को connect करना

किसी भी AI client को `https://<host>:56244` पर इंगित करें (या यदि TLS बंद है तो `http://...`)। Proxy चार आकारों का जवाब देता है:

| Format | Path | उदाहरण clients |
|--------|------|-----------------|
| Gemini | `/google/v1beta/models/<model>:generateContent` | OpenClaw, Gemini CLI, Antigravity |
| Anthropic | `/v1/messages` | Claude Code, Anthropic SDKs |
| OpenAI | `/v1/chat/completions` | Cursor, Continue, custom scripts, अधिकांश LLM apps |
| Ollama-native | `/api/chat` | Ollama clients passing through |

### Claude Code

```bash
export ANTHROPIC_BASE_URL=https://localhost:56244
export ANTHROPIC_API_KEY=<your-vault-client-token>
claude
```

जब upstream Anthropic credits समाप्त हो जाते हैं, तो dispatch उन providers पर fallback करता है जो आपने इस client के लिए `fallback_services` में सेट किए हैं। non-Claude fallback के लिए स्पष्ट रूप से opt in करने के लिए:

```yaml
proxy:
  anthropic_fallback_model: "claude-haiku-4-5-20251001"
```

(खाली default के कारण dispatch एक error लौटाता है ताकि misrouting तुरंत surface हो जाए।)

### Cursor / Continue

Cursor में **Settings → AI → OpenAI API**:

```
Base URL:  https://localhost:56244
API Key:   <your-vault-client-token>
Model:     gemini-2.5-flash    # या कोई भी model जिसे wall-vault जानता है
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

OpenClaw एक TUI agent framework है जिसकी सेवा के लिए wall-vault मूल रूप से बनाया गया था। Dashboard का **Add Agent** modal agent type को `openclaw` (या `nanoclaw`) पर सेट करता है; फिर wall-vault सीधे `~/.openclaw/openclaw.json` लिखता है, जिसमें provider URLs, vault token, और model entries शामिल हैं:

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

## Configuration

`wall-vault setup` या तो `./wall-vault.yaml` या `~/.wall-vault/config.yaml` लिखता है। उन fields के लिए जो wizard नहीं पूछता, हाथ से संपादित करें।

```yaml
mode: standalone     # standalone | distributed
lang: en             # en | ko | zh | ja | es | fr | de | pt | ar | hi | id | th | sw | ha | ne | mn | zu
theme: light         # light | dark | cherry | ocean | gold | autumn | winter

proxy:
  port: 56244
  host: ""           # default: standalone में 127.0.0.1, distributed में 0.0.0.0
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
  plain_port: 56245              # TLS चालू होने पर loopback-only HTTP companion
  ollama_keep_alive: "30m"       # "-1" कभी unload न करें, "0" तुरंत unload करें
  ollama_num_ctx: 8192
  oai_stream_forward: false      # opt-in real backend SSE passthrough
  anthropic_fallback_model: ""   # anthropic dispatch पर opt-in non-Claude rewrite

vault:
  port: 56243
  host: ""
  admin_token: ""
  admin_ip_whitelist: []
  master_password: ""            # AES-GCM key encryption password
  data_dir: ~/.wall-vault/data
  services_dir: ~/.wall-vault/services
  tls:
    enabled: false
    cert_file: ""
    key_file: ""
  bootstrap_port: 56247          # केवल ca.crt सर्व करने वाला plain-HTTP listener

doctor:
  interval: 5m
  auto_fix: true
  log_file: /tmp/wall-vault-doctor.log

hooks:
  on_model_change: ""            # shell command (env: SERVICE, MODEL)
  on_key_exhausted: ""
  on_service_down: ""
  on_doctor_fix: ""
  openclaw_socket: ""
```

### Environment variables

प्रत्येक YAML field का एक env override है जो file पर जीतता है। सामान्य:

| Variable | विवरण |
|----------|-------------|
| `WV_LANG`, `WV_THEME` | भाषा और theme |
| `WV_PROXY_PORT`, `WV_PROXY_HOST` | Proxy listen address |
| `WV_VAULT_PORT`, `WV_VAULT_HOST` | Vault listen address |
| `WV_VAULT_URL`, `WV_VAULT_TOKEN` | Distributed-mode endpoints |
| `WV_ADMIN_TOKEN`, `WV_MASTER_PASS` | Vault credentials |
| `WV_KEY_GOOGLE`, `WV_KEY_OPENROUTER`, `WV_KEY_ANTHROPIC`, `WV_KEY_OPENAI` | API keys (कई के लिए comma-separated) |
| `WV_PROXY_TLS_ENABLED`, `WV_PROXY_TLS_CERT`, `WV_PROXY_TLS_KEY` | Proxy TLS |
| `WV_VAULT_TLS_ENABLED`, `WV_VAULT_TLS_CERT`, `WV_VAULT_TLS_KEY` | Vault TLS |
| `WV_PROXY_PLAIN_PORT` | Loopback HTTP companion (disable के लिए `0`) |
| `WV_VAULT_BOOTSTRAP_PORT` | CA bootstrap listener (disable के लिए `0`) |
| `WV_OLLAMA_URL`, `WV_OLLAMA_KEEP_ALIVE`, `WV_OLLAMA_NUM_CTX` | Ollama tuning |
| `WV_LMSTUDIO_URL`, `WV_VLLM_URL`, `WV_LLAMACPP_URL` | Local backend overrides |
| `WV_TOKEN_SENTINEL_FALLBACK` | Loopback "proxy-managed" sentinel substitution |
| `WV_OAI_STREAM_FORWARD` | OpenAI-compat real backend SSE passthrough |
| `WV_ANTHROPIC_FALLBACK_MODEL` | anthropic पर opt-in non-Claude rewrite |

---

## Modes

### Standalone (default)

Vault और proxy एक ही process में चलते हैं। एक ही host के लिए सर्वश्रेष्ठ जो keys और agent दोनों को host करता है। Default रूप से केवल loopback पर सुनता है।

```bash
wall-vault start    # दोनों चलाता है
```

### Distributed

Vault एक host पर चलता है (**vault host**) और सभी keys संग्रहीत करता है; अन्य hosts पर कई proxies प्रत्येक एक per-client token के साथ authenticate करते हैं। तब उपयोगी जब कई machines को बिना उन्हें इधर-उधर copy किए समान keys की आवश्यकता हो।

**Vault host:**

```bash
WV_VAULT_HOST=0.0.0.0 wall-vault vault
```

**प्रत्येक proxy host:**

```bash
WV_VAULT_URL=http://<vault-host>:56243 \
WV_VAULT_TOKEN=<this-client-token> \
WV_PROXY_HOST=0.0.0.0 \
wall-vault proxy
```

Dashboard का **Add Client** modal एक token mint करता है, एक agent type पंजीकृत करता है, और proxy बिना restart के SSE पर अपना config उठा लेता है।

---

## Plugin yaml (drop-in backend)

किसी भी OpenAI-compatible backend को `~/.wall-vault/services/` के अंतर्गत एक yaml के रूप में जोड़ा जा सकता है। wall-vault इसे start पर उठा लेता है, इसे एक routable service के रूप में पंजीकृत करता है, और dispatch + OAI-compat detection set + Gemini-stream bridge सभी इसे कोड परिवर्तन के बिना देखते हैं।

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
inline_no_think_for_qwen3: false   # opt in यदि आपका backend marker को strip करता है
```

Hub topology (एक wall-vault दूसरे के सामने) `tls_internal_ca: true`, `auth.type: bearer`, और `preserve_model_id: true` के माध्यम से समर्थित है।

---

## Source से build करें

```bash
git clone https://github.com/sookmook/wall-vault
cd wall-vault
go build -o wall-vault .
./wall-vault setup
./wall-vault start
```

पूरे supported set के लिए cross-compile:

```bash
make build-all   # linux-amd64, linux-arm64, darwin-amd64, darwin-arm64, windows-amd64
```

Versions `v{major}.{minor}.{patch}.{YYYYMMDD}.{HHmmss}` का अनुसरण करती हैं; Makefile में `BASE_VERSION` prefix सेट करता है।

### Project layout

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

## Documentation

- [User manual](docs/MANUAL.en.md) — installation, dashboard, agents, troubleshooting
- [API reference](docs/API.en.md) — request/response आकारों के साथ हर endpoint
- [CHANGELOG](CHANGELOG.md)

---

## Tech stack

- Go 1.25, single static binary
- Server-rendered dashboard के लिए [templ](https://templ.guide), partial updates के लिए [HTMX](https://htmx.org)
- विश्राम पर key encryption के लिए AES-GCM (PBKDF2-derived key)
- vault और proxies के बीच live config sync के लिए Server-Sent Events
- Self-signed internal CA + per-host certs (कोई public DNS / Let's Encrypt आवश्यक नहीं)

## License

GPL-3.0. [LICENSE](LICENSE) देखें।

## योगदान

Pull requests का स्वागत है। [CONTRIBUTING.md](CONTRIBUTING.md) देखें। बड़े बदलावों के लिए कृपया पहले एक issue खोलें ताकि design पर चर्चा हो सके।
