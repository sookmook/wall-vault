<p align="center">
  <img src="docs/logo.png" alt="wall-vault" width="800">
</p>

# wall-vault

> **एउटै Go binary मा API key vault + AI proxy।**
> AES-GCM सँग स्थानीय रूपमा कुञ्जीहरू भण्डारण गर्छ, प्रदायकहरू बीच घुमाउँछ, एउटा असफल हुँदा फलब्याक गर्छ, र वास्तविक-समय ड्यासबोर्डसँग आउँछ।

[![License: GPL v3](https://img.shields.io/badge/License-GPLv3-blue.svg)](LICENSE)
[![Go](https://img.shields.io/badge/Go-1.25-00ADD8.svg)](go.mod)

[English](README.md) · [한국어](README.ko.md) · [中文](README.zh.md) · [日本語](README.ja.md) · [Español](README.es.md) · [Français](README.fr.md) · [Deutsch](README.de.md) · [Português](README.pt.md) · [العربية](README.ar.md) · [हिन्दी](README.hi.md) · [Bahasa Indonesia](README.id.md) · [ภาษาไทย](README.th.md) · [Kiswahili](README.sw.md) · [Hausa](README.ha.md) · **नेपाली** · [Монгол](README.mn.md) · [isiZulu](README.zu.md)

---

## यो के हो

wall-vault एउटा AI एजेन्ट (OpenClaw, Claude Code, Cursor, Continue, तपाईंको आफ्नै स्क्रिप्ट) र यसले कुरा गर्ने क्लाउड वा स्थानीय AI प्रदायकहरूको बीचमा बस्छ। एउटै binary मा दुई कुरा:

- **Vault** — AES-GCM (मास्टर पासवर्ड सँग) मार्फत आराममा एन्क्रिप्ट गरिएका API कुञ्जीहरू भण्डारण गर्छ, घुमाउँछ, प्रति-कुञ्जी प्रयोग र कूलडाउनहरू ट्र्याक गर्छ, SSE मार्फत परिवर्तनहरू प्रसारण गर्छ, र `:56243` मा एक वेब ड्यासबोर्ड प्रदान गर्छ।
- **Proxy** — `:56244` मा Gemini, Anthropic, र OpenAI-compatible endpoints प्रदर्शन गर्छ, vault बाट कुञ्जी छनोट गर्छ, तपाईंले कन्फिगर गरेको upstream मा पठाउँछ, र एउटा असफल हुँदा अर्को प्रदायकमा फलब्याक गर्छ।

यसले अनुरोधका चार आकारहरू (Gemini `:generateContent`, Anthropic `/v1/messages`, OpenAI `/v1/chat/completions`, र Ollama-native `/api/chat`) र upstream का पाँच श्रेणीहरू समर्थन गर्छ:

| प्रदायक | टिप्पणीहरू |
|----------|-------|
| Google Gemini | नेटिभ API; प्रति परियोजना कुञ्जी घुमाव |
| Anthropic | नेटिभ `/v1/messages` पासथ्रू |
| OpenAI | नेटिभ `/v1/chat/completions` |
| OpenRouter | 340+ मोडेलहरू, `:free` संस्करणहरूमा स्वत:-फलब्याक |
| Ollama / LM Studio / vLLM / llama.cpp / Jan / KoboldCpp / TabbyAPI / mlx-server / LiteLLM | OpenAI-compat स्थानीय backend हरू; plugin yaml मार्फत drop-in |

OpenAI-compatible नयाँ backend थप्न `~/.wall-vault/services/` मुनि एउटा yaml फाइल मात्र हो — कुनै कोड परिवर्तन छैन।

## तपाईंलाई किन चाहिन सक्छ

- तपाईं तीन वा चार AI सेवाहरू व्यवस्थापन गरिरहनुभएको छ र एजेन्टले कुरा गर्ने एउटा URL चाहनुहुन्छ।
- तपाईं कूलडाउनमा रहेको free-tier कुञ्जीले सत्र भङ्ग नगरी अर्कोलाई बाटो दिओस् भन्ने चाहनुहुन्छ।
- तपाईं उही कुञ्जीहरूले एउटै LAN मा प्रमाणीकरण कपी नगरी धेरै bots / IDEs / scripts लाई शक्ति दिओस् भन्ने चाहनुहुन्छ।
- API कुञ्जीहरू सम्पादन गर्नका लागि तपाईं ड्यासबोर्ड चाहनुहुन्छ, environment variables होइन।
- क्लाउड सीमाहरू समाप्त हुँदा तपाईं स्थानीय-पहिले विकल्प (Ollama / LM Studio) चाहनुहुन्छ।

## द्रुत सुरुवात

### स्थापना (Linux / macOS)

```bash
curl -fsSL https://raw.githubusercontent.com/sookmook/wall-vault/main/install.sh | sh
```

वा पूर्व-निर्मित binary सिधै डाउनलोड गर्नुहोस्:

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

### स्थापना (Windows)

```powershell
Invoke-WebRequest -Uri `
  "https://github.com/sookmook/wall-vault/releases/latest/download/wall-vault-windows-amd64.exe" `
  -OutFile wall-vault.exe
```

### पहिलो रन

```bash
wall-vault setup    # अन्तरक्रियात्मक wizard — पोर्ट, सेवाहरू, admin token, master password छनोट गर्छ
wall-vault start    # vault र proxy दुवै सुरु गर्छ
```

ब्राउजरमा `http://localhost:56243` (वा TLS सक्रिय भएपछि `https://...` — तल हेर्नुहोस्) खोल्नुहोस्। ड्यासबोर्डले `setup` ले छापेको admin token माग्छ। त्यहाँबाट तपाईं API कुञ्जीहरू थप्नुहुन्छ, क्लाइन्टहरू दर्ता गर्नुहुन्छ, र पुन: सुरु नगरी मोडेलहरू स्विच गर्नुहुन्छ।

---

## TLS (सिफारिस गरिएको)

पूर्वनिर्धारित रूपमा `wall-vault setup` ले TLS बिनाको कन्फिग लेख्छ, त्यसैले दुवै सुन्नेहरू सादा HTTP जवाफ दिन्छन्। यस README का उदाहरण URL हरूले `https://localhost:56244` प्रयोग गर्छन् किनभने धेरै एजेन्टहरू (OpenClaw, Claude Code, Cursor) ले एउटै TLS-अग्र endpoint चाहन्छन् जुन तपाईंले पछि proxy लाई अर्को host मा सार्दा नटुट्ने हुन्छ। ती उदाहरणहरू मिलाउन, बन्डल गरिएको आन्तरिक CA सँग एकपटक TLS सक्रिय गर्नुहोस्:

```bash
# 1. wall-vault आन्तरिक CA सिर्जना गर्नुहोस् (एकपटक मात्र, ~/.wall-vault/ca.{crt,key} मा रहन्छ)
wall-vault cert init

# 2. यो मेसिनको लागि host certificate जारी गर्नुहोस्
#    SAN हरूमा hostname, localhost, 127.0.0.1, र पत्ता लागेको कुनै पनि LAN IP समावेश छ
wall-vault cert issue $(hostname)

# 3. स्थानीय OS keychain मा CA विश्वास गर्नुहोस्
wall-vault cert install-trust

# 4. सुन्नेहरूलाई TLS मा स्विच गर्नुहोस्
export WV_PROXY_TLS_ENABLED=1
export WV_PROXY_TLS_CERT="$HOME/.wall-vault/$(hostname).crt"
export WV_PROXY_TLS_KEY="$HOME/.wall-vault/$(hostname).key"
export WV_VAULT_TLS_ENABLED=1
export WV_VAULT_TLS_CERT="$HOME/.wall-vault/$(hostname).crt"
export WV_VAULT_TLS_KEY="$HOME/.wall-vault/$(hostname).key"
wall-vault start
```

तपाईंको LAN मा अर्को मेसिनको लागि: `~/.wall-vault/ca.crt` त्यहाँ कपी गर्नुहोस् र त्यहाँ `wall-vault cert install-trust --ca <path>` चलाउनुहोस्। CA सबैतिर विश्वास गरिएपछि, नेटवर्कमा प्रत्येक मेसिनले प्रमाणपत्र चेतावनी बिना `https://<host>:56244` मार्फत proxy सम्म पुग्न सक्छ।

यदि तपाईं सादा HTTP मा रहन चाहनुहुन्छ भने, कन्फिग जस्ताको त्यस्तै छोड्नुहोस् र तलका client snippets मा `https://` लाई `http://` ले प्रतिस्थापन गर्नुहोस्। दुवै schemes काम गर्छन्; भिन्नता कुन port ले TLS handshake को जवाफ दिन्छ भन्ने मा हो।

**Loopback fallback।** wall-vault CA सम्मान गर्न नसक्ने एउटै-host क्लाइन्टहरू (विशेष गरी OpenClaw को बन्डल गरिएको Node runtime, जसले spawn मा `NODE_EXTRA_CA_CERTS` पुनर्लेखन गर्छ) `127.0.0.1:56245` मा loopback-only सादा-HTTP companion मार्फत proxy सम्म पुग्छन्। TLS सक्रिय हुँदा wall-vault ले यसलाई स्वत: सक्षम बनाउँछ।

---

## क्लाइन्टहरू जोड्ने

कुनै पनि AI client लाई `https://<host>:56244` मा देखाउनुहोस् (वा TLS बन्द छ भने `http://...`)। Proxy ले चार आकारहरूमा जवाफ दिन्छ:

| ढाँचा | पथ | उदाहरण क्लाइन्टहरू |
|--------|------|-----------------|
| Gemini | `/google/v1beta/models/<model>:generateContent` | OpenClaw, Gemini CLI, Antigravity |
| Anthropic | `/v1/messages` | Claude Code, Anthropic SDKs |
| OpenAI | `/v1/chat/completions` | Cursor, Continue, अनुकूलन स्क्रिप्टहरू, धेरै LLM apps |
| Ollama-native | `/api/chat` | Ollama क्लाइन्टहरू पासथ्रू |

### Claude Code

```bash
export ANTHROPIC_BASE_URL=https://localhost:56244
export ANTHROPIC_API_KEY=<your-vault-client-token>
claude
```

जब upstream Anthropic क्रेडिट सकिन्छ, dispatch तपाईंले यो client को लागि `fallback_services` मा सेट गरेका जुनसुकै प्रदायकहरूमा फलब्याक हुन्छ। गैर-Claude फलब्याकमा स्पष्ट रूपमा छनोट गर्न:

```yaml
proxy:
  anthropic_fallback_model: "claude-haiku-4-5-20251001"
```

(खाली पूर्वनिर्धारितले dispatch लाई त्रुटि फिर्ता गर्न लगाउँछ ताकि गलत-राउटिङ तुरुन्तै देखा परोस्।)

### Cursor / Continue

Cursor **Settings → AI → OpenAI API** मा:

```
Base URL:  https://localhost:56244
API Key:   <your-vault-client-token>
Model:     gemini-2.5-flash    # वा wall-vault लाई थाहा भएको कुनै पनि मोडेल
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

OpenClaw एक TUI एजेन्ट फ्रेमवर्क हो जसको लागि wall-vault मूल रूपमा सेवा प्रदान गर्न निर्मित गरिएको थियो। ड्यासबोर्डको **Add Agent** मोडलले एजेन्ट प्रकारलाई `openclaw` (वा `nanoclaw`) मा सेट गर्छ; त्यसपछि wall-vault ले प्रदायक URL हरू, vault token, र मोडेल प्रविष्टिहरू सहित `~/.openclaw/openclaw.json` लाई सिधै लेख्छ:

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

## कन्फिगरेसन

`wall-vault setup` ले या त `./wall-vault.yaml` वा `~/.wall-vault/config.yaml` लेख्छ। wizard ले नसोधेका फिल्डहरूको लागि हातैले सम्पादन गर्नुहोस्।

```yaml
mode: standalone     # standalone | distributed
lang: en             # en | ko | zh | ja | es | fr | de | pt | ar | hi | id | th | sw | ha | ne | mn | zu
theme: light         # light | dark | cherry | ocean | gold | autumn | winter

proxy:
  port: 56244
  host: ""           # पूर्वनिर्धारित: 127.0.0.1 standalone, 0.0.0.0 distributed
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
  plain_port: 56245              # TLS सक्रिय हुँदा loopback-only HTTP companion
  ollama_keep_alive: "30m"       # "-1" कहिल्यै unload नगर्ने, "0" तुरुन्तै unload गर्ने
  ollama_num_ctx: 8192
  oai_stream_forward: false      # opt-in वास्तविक backend SSE पासथ्रू
  anthropic_fallback_model: ""   # anthropic dispatch मा opt-in non-Claude पुनर्लेखन

vault:
  port: 56243
  host: ""
  admin_token: ""
  admin_ip_whitelist: []
  master_password: ""            # AES-GCM कुञ्जी एन्क्रिप्शन पासवर्ड
  data_dir: ~/.wall-vault/data
  services_dir: ~/.wall-vault/services
  tls:
    enabled: false
    cert_file: ""
    key_file: ""
  bootstrap_port: 56247          # केवल ca.crt सेवा गर्ने सादा-HTTP listener

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

प्रत्येक YAML फिल्डसँग env override हुन्छ जसले फाइलमा जित्छ। सामान्य:

| चर | विवरण |
|----------|-------------|
| `WV_LANG`, `WV_THEME` | भाषा र थिम |
| `WV_PROXY_PORT`, `WV_PROXY_HOST` | Proxy सुन्ने ठेगाना |
| `WV_VAULT_PORT`, `WV_VAULT_HOST` | Vault सुन्ने ठेगाना |
| `WV_VAULT_URL`, `WV_VAULT_TOKEN` | Distributed-मोड endpoints |
| `WV_ADMIN_TOKEN`, `WV_MASTER_PASS` | Vault क्रेडेन्सियलहरू |
| `WV_KEY_GOOGLE`, `WV_KEY_OPENROUTER`, `WV_KEY_ANTHROPIC`, `WV_KEY_OPENAI` | API कुञ्जीहरू (धेरैको लागि अल्पविरामले छुट्याइएको) |
| `WV_PROXY_TLS_ENABLED`, `WV_PROXY_TLS_CERT`, `WV_PROXY_TLS_KEY` | Proxy TLS |
| `WV_VAULT_TLS_ENABLED`, `WV_VAULT_TLS_CERT`, `WV_VAULT_TLS_KEY` | Vault TLS |
| `WV_PROXY_PLAIN_PORT` | Loopback HTTP companion (अक्षम गर्न `0`) |
| `WV_VAULT_BOOTSTRAP_PORT` | CA bootstrap listener (अक्षम गर्न `0`) |
| `WV_OLLAMA_URL`, `WV_OLLAMA_KEEP_ALIVE`, `WV_OLLAMA_NUM_CTX` | Ollama ट्युनिङ |
| `WV_LMSTUDIO_URL`, `WV_VLLM_URL`, `WV_LLAMACPP_URL` | स्थानीय backend overrides |
| `WV_TOKEN_SENTINEL_FALLBACK` | Loopback "proxy-managed" sentinel प्रतिस्थापन |
| `WV_OAI_STREAM_FORWARD` | OpenAI-compat वास्तविक backend SSE पासथ्रू |
| `WV_ANTHROPIC_FALLBACK_MODEL` | Anthropic मा opt-in non-Claude पुनर्लेखन |

---

## मोडहरू

### Standalone (पूर्वनिर्धारित)

Vault र proxy एउटै process मा चल्छन्। कुञ्जीहरू र एजेन्ट दुवै राख्ने एउटै host को लागि उत्तम। पूर्वनिर्धारित रूपमा loopback मा मात्र सुन्छ।

```bash
wall-vault start    # दुवै चलाउँछ
```

### Distributed

Vault एउटा host (**vault host**) मा चल्छ र सबै कुञ्जीहरू भण्डारण गर्छ; अन्य hosts मा भएका धेरै proxies हरूले प्रति-client token सँग प्रमाणीकरण गर्छन्। धेरै मेसिनहरूलाई कुञ्जीहरू कपी नगरी उही कुञ्जीहरू चाहिएको बेलामा उपयोगी।

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

ड्यासबोर्डको **Add Client** मोडलले token बनाउँछ, एजेन्ट प्रकार दर्ता गर्छ, र proxy ले पुन: सुरु बिना SSE मार्फत आफ्नो कन्फिग लिन्छ।

---

## Plugin yaml (drop-in backend)

कुनै पनि OpenAI-compatible backend `~/.wall-vault/services/` मुनि yaml को रूपमा थप्न सकिन्छ। wall-vault ले यसलाई सुरुमा लिन्छ, यसलाई एक मार्ग योग्य सेवाको रूपमा दर्ता गर्छ, र dispatch + OAI-compat detection set + Gemini-stream bridge सबैले कोड परिवर्तन बिना यसलाई देख्छन्।

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
inline_no_think_for_qwen3: false   # तपाईंको backend ले मार्कर हटाउँछ भने opt in गर्नुहोस्
```

Hub topology (एउटा wall-vault ले अर्कोको अगाडि बस्छ) `tls_internal_ca: true`, `auth.type: bearer`, र `preserve_model_id: true` मार्फत समर्थित छ।

---

## स्रोतबाट build गर्ने

```bash
git clone https://github.com/sookmook/wall-vault
cd wall-vault
go build -o wall-vault .
./wall-vault setup
./wall-vault start
```

समर्थित सम्पूर्ण सेटको लागि cross-compile गर्नुहोस्:

```bash
make build-all   # linux-amd64, linux-arm64, darwin-amd64, darwin-arm64, windows-amd64
```

संस्करणहरूले `v{major}.{minor}.{patch}.{YYYYMMDD}.{HHmmss}` अनुसरण गर्छन्; Makefile मा `BASE_VERSION` ले prefix सेट गर्छ।

### परियोजना लेआउट

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

## कागजात

- [प्रयोगकर्ता पुस्तिका](docs/MANUAL.en.md) — स्थापना, ड्यासबोर्ड, एजेन्टहरू, समस्या निवारण
- [API सन्दर्भ](docs/API.en.md) — request/response आकारहरू सहित प्रत्येक endpoint
- [CHANGELOG](CHANGELOG.md)

---

## प्रविधि स्ट्याक

- Go 1.25, एकल स्थिर binary
- सर्भर-रेन्डर गरिएको ड्यासबोर्डको लागि [templ](https://templ.guide), आंशिक अपडेटहरूको लागि [HTMX](https://htmx.org)
- आरामको कुञ्जी एन्क्रिप्शनको लागि AES-GCM (PBKDF2-व्युत्पन्न कुञ्जी)
- vault र proxies बीच लाइभ कन्फिग सिङ्कको लागि Server-Sent Events
- स्व-हस्ताक्षरित आन्तरिक CA + प्रति-host certs (कुनै सार्वजनिक DNS / Let's Encrypt आवश्यक छैन)

## Licens

GPL-3.0। [LICENSE](LICENSE) हेर्नुहोस्।

## योगदान

Pull requests स्वागत छन्। [CONTRIBUTING.md](CONTRIBUTING.md) हेर्नुहोस्। ठूला परिवर्तनहरूको लागि कृपया डिजाइन छलफल गर्न पहिले एउटा issue खोल्नुहोस्।
