# wall-vault User Manual

[English](MANUAL.md) · [한국어](MANUAL.ko.md) · [中文](MANUAL.zh.md) · [日本語](MANUAL.ja.md) · [Español](MANUAL.es.md) · [Français](MANUAL.fr.md) · [Deutsch](MANUAL.de.md) · [Português](MANUAL.pt.md) · [العربية](MANUAL.ar.md) · [हिन्दी](MANUAL.hi.md) · [Bahasa Indonesia](MANUAL.id.md) · [ภาษาไทย](MANUAL.th.md) · [Kiswahili](MANUAL.sw.md) · [Hausa](MANUAL.ha.md) · नेपाली · [Монгол](MANUAL.mn.md) · [isiZulu](MANUAL.zu.md)

यो निर्देशिकाले wall-vault स्थापना, कन्फिगर र सञ्चालन गर्ने तरिका समेट्छ। एक नजरमा अवलोकनको लागि [README](../README.md) हेर्नुहोस्। HTTP API विवरणको लागि [API reference](API.md) हेर्नुहोस्।

## विषयसूची

1. [wall-vault के गर्छ](#wall-vault-के-गर्छ)
2. [स्थापना](#स्थापना)
3. [setup wizard सँग पहिलो रन](#setup-wizard-सँग-पहिलो-रन)
4. [TLS सक्षम पार्ने](#tls-सक्षम-पार्ने)
5. [API key दर्ता गर्ने](#api-key-दर्ता-गर्ने)
6. [Agents जोड्ने](#agents-जोड्ने)
7. [Dashboard](#dashboard)
8. [Distributed मोड](#distributed-मोड)
9. [Auto-start](#auto-start)
10. [Plugin yamls](#plugin-yamls)
11. [Doctor](#doctor)
12. [Hooks](#hooks)
13. [वातावरण चरहरू](#वातावरण-चरहरू)
14. [समस्या समाधान](#समस्या-समाधान)

---

## wall-vault के गर्छ

wall-vault एकल Go binary हो जसले दुई सहयोगी सेवाहरू बन्डल गर्छ:

- **Vault** ले API keys विश्रामको समयमा एन्क्रिप्ट गरी राख्छ (मास्टर पासवर्डसँग AES-GCM), प्रति key प्रयोग र cooldown ट्र्याक गर्छ, Server-Sent Events (SSE) मार्फत परिवर्तनहरू प्रसारण गर्छ, र मानव सञ्चालकहरूको लागि `:56243` मा वेब dashboard सेवा प्रदान गर्छ।
- **Proxy** ले `:56244` मा Gemini, Anthropic, OpenAI-compatible, र Ollama-native endpoints expose गर्छ। Proxy तर्फ इङ्गित गर्ने कुनै पनि AI client ले vault मा भएको keys प्रयोग गर्छ — clients ले तिनीहरूलाई कहिल्यै देख्दैन। जब एक upstream विफल हुन्छ, dispatch क्रममा अर्को provider मा फलब्याक हुन्छ।

यो उपयोगी हुन्छ जब:

- तपाईंसँग धेरै providers का keys छन् र agent ले कुरा गर्ने एउटै URL चाहनुहुन्छ।
- तपाईं cooldown मा रहेको free-tier key लाई session नभाँचेरै बाहिर निस्कन दिन चाहनुहुन्छ।
- तपाईं credentials कपी नगरीकनै एउटै LAN मा धेरै bots, IDEs, वा scripts लाई एउटै keys ले शक्ति दिन चाहनुहुन्छ।
- तपाईं keys सम्पादन गर्न र models स्विच गर्न environment variables होइन dashboard चाहनुहुन्छ।
- Cloud सीमाहरू सकिएपछि तपाईं स्थानीय fallback (Ollama, LM Studio, vLLM) चाहनुहुन्छ।

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

## स्थापना

### Linux / macOS one-liner

```bash
curl -fsSL https://raw.githubusercontent.com/sookmook/wall-vault/main/install.sh | sh
```

स्क्रिप्टले OS र architecture स्वतः पत्ता लगाउँछ, सही binary लाई `~/.local/bin/wall-vault` मा डाउनलोड गर्छ, र यसलाई executable बनाउँछ। यदि `~/.local/bin` तपाईंको `PATH` मा छैन भने, यसलाई थप्नुहोस्:

```bash
export PATH="$HOME/.local/bin:$PATH"
```

### म्यानुअल डाउनलोड

प्रत्येक release मा प्रि-बिल्ट binaries `https://github.com/sookmook/wall-vault/releases` मा प्रकाशित हुन्छन्।

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

### Source बाट build गर्ने

Go 1.25 वा नयाँ चाहिन्छ।

```bash
git clone https://github.com/sookmook/wall-vault
cd wall-vault
go build -o wall-vault .
```

`make build-all` ले सबै पाँच समर्थित प्लेटफर्महरूमा cross-compile गर्छ। Binaries `bin/` मा अवतरण गर्छन्।

---

## setup wizard सँग पहिलो रन

```bash
wall-vault setup
```

Wizard ले तपाईंलाई क्रममा सोध्छ:

1. **भाषा** — 17 UI locales मध्ये एक छनोट गर्छ। `$LANG` बाट स्वतः पत्ता लगाइन्छ; wizard ले जे भए पनि सूची प्रस्तुत गर्छ।
2. **Theme** — `light` (पूर्वनिर्धारित), `dark`, `cherry`, `ocean`, `gold`, `autumn`, `winter`। केवल कस्मेटिक।
3. **मोड** — `standalone` (एकल host, पूर्वनिर्धारित) वा `distributed` (एउटा host मा vault, अरूमा proxies)।
4. **Bot नाम** — एउटा फ्रि-फर्म `client_id` slug। Vault ले यसलाई प्रति-client config (model overrides, fallback chains) स्कोप गर्न प्रयोग गर्छ।
5. **Proxy port** — पूर्वनिर्धारित `56244`।
6. **Vault port** — पूर्वनिर्धारित `56243` (केवल standalone)।
7. **सेवा छनोट** — प्रत्येकको लागि y/N: Google Gemini, OpenRouter, Anthropic, OpenAI, Ollama, LM Studio, vLLM। बहु छनोटहरू ठीक छन्; प्रत्येकले अन्त्यमा यसको env-var hint लेख्छ।
8. **Tool filter** — `strip_all` (पूर्वनिर्धारित; सुरक्षाको लागि सबै आगमन tool definitions ब्लक गर्छ) वा `passthrough` (कुनै पनि tool पास हुन दिन्छ)।
9. **Admin token** — स्वतः उत्पन्न गर्न खाली छोड्नुहोस्। Dashboard मा लग इन गर्न यो token चाहिन्छ।
10. **Master password** — एन्क्रिप्शन नचाहिएको खण्डमा खाली छोड्नुहोस् (सिफारिस गरिएको छैन); विश्रामको समयमा key store लाई AES-GCM एन्क्रिप्ट गर्न मान सेट गर्नुहोस्।
11. **Save path** — हालको directory मा `wall-vault.yaml` मा पूर्वनिर्धारित। Loader ले `~/.wall-vault/config.yaml` मा पनि हेर्छ।

बचत पछि, wizard ले `doctor.FixTrust` चलाउँछ ताकि कुनै पनि स्थानीय रूपमा-स्थापना गरिएको agent (OpenClaw, Claude Code, Cline) ले wall-vault आन्तरिक CA स्वचालित रूपमा यसको trust store मा थपिएको पाओस्। यदि त्यस्तो कुनै agent स्थापना गरिएको छैन भने, चरणले `SKIP` प्रिन्ट गर्छ र केही पनि लेख्दैन।

त्यसपछि binary सुरु गर्नुहोस्:

```bash
wall-vault start
```

`start` ले एक process मा vault र proxy दुवै चलाउँछ (standalone मोड)। Distributed मोडको लागि vault host मा `wall-vault vault` र प्रत्येक proxy host मा `wall-vault proxy` प्रयोग गर्नुहोस्।

ब्राउजरमा `http://localhost:56243` खोल्नुहोस्। Wizard ले प्रिन्ट गरेको admin token सँग लग इन गर्नुहोस्।

---

## TLS सक्षम पार्ने

Wizard को पूर्वनिर्धारितहरूले दुवै listeners लाई plain HTTP मा छोड्छ। धेरैजसो agents (OpenClaw, Claude Code, Cursor) एकल HTTPS endpoint विरुद्ध राम्रोसँग काम गर्छन्, त्यसैले स्थानीय मेसिन भन्दा बढी फैलिने कुनै पनि deployment मा TLS सिफारिस गरिन्छ।

wall-vault आफ्नै आन्तरिक CA सहित आउँछ त्यसैले तपाईंलाई सार्वजनिक DNS नाम वा Let's Encrypt चाहिँदैन।

```bash
# 1. आन्तरिक CA सिर्जना गर्नुहोस् — ~/.wall-vault/ca.{crt,key} मा लेखिन्छ।
#    CA पूर्वनिर्धारित रूपमा 10 वर्षको लागि राम्रो छ; --ca-years सँग override गर्नुहोस्।
wall-vault cert init

# 2. Host certificate जारी गर्नुहोस्। Subject Alternative Names ले स्वतः समावेश गर्छ:
#       hostname, "localhost", "127.0.0.1", र पत्ता लगाइएको कुनै पनि गैर-loopback LAN IP।
#    --dir सँग issuer dir override गर्नुहोस्, --host-years सँग validity।
wall-vault cert issue $(hostname)

# 3. यो मेसिनको OS keychain मा CA लाई trust गर्नुहोस्।
#    Linux: update-ca-certificates मार्फत /etc/ssl/certs/ मा लेख्छ (sudo चाहिन्छ)।
#    macOS: security add-trusted-cert मार्फत System keychain मा थप्छ (sudo चाहिन्छ)।
#    Windows: certutil मार्फत CurrentUser\Root मा import गर्छ (admin आवश्यक छैन)।
wall-vault cert install-trust

# 4. दुवै listeners मा TLS सक्षम गर्नुहोस्।
export WV_PROXY_TLS_ENABLED=1
export WV_PROXY_TLS_CERT="$HOME/.wall-vault/$(hostname).crt"
export WV_PROXY_TLS_KEY="$HOME/.wall-vault/$(hostname).key"
export WV_VAULT_TLS_ENABLED=1
export WV_VAULT_TLS_CERT="$HOME/.wall-vault/$(hostname).crt"
export WV_VAULT_TLS_KEY="$HOME/.wall-vault/$(hostname).key"

wall-vault start
```

अन्य LAN मेसिनहरूमा trust विस्तार गर्न, `~/.wall-vault/ca.crt` लाई कपी गर्नुहोस् र प्रत्येकमा `wall-vault cert install-trust --ca <path>` चलाउनुहोस्। Vault ले पनि `:56247` (**bootstrap port**) मा एउटा सानो plain-HTTP listener मार्फत `ca.crt` expose गर्छ catch-22 केसको लागि जहाँ नयाँ client लाई HTTPS कुरा गर्न CA चाहिन्छ।

### Loopback HTTP companion

केही agents — विशेष गरी OpenClaw को बन्डल गरिएको Node runtime — process spawn मा `NODE_EXTRA_CA_CERTS` पुनः लेख्छन्, सञ्चालक-प्रदान CA hint छाडेर। तिनीहरूले `cert install-trust` पछि पनि daemon भित्रबाट wall-vault CA लाई सम्मान गर्न सक्दैनन्। wall-vault ले TLS सक्षम भएको जुनसुकै बेला `127.0.0.1:56245` मा अतिरिक्त **loopback-only plain-HTTP listener** बाइन्ड गरेर यो वरिपरि काम गर्छ। Same-host clients ले त्यो port मार्फत TLS बिना नै proxy मा पुग्छन्; LAN clients ले TLS listener प्रयोग गर्न जारी राख्छन्।

यदि चाहिँदैन भने `WV_PROXY_PLAIN_PORT=0` सँग असक्षम गर्नुहोस्।

### `wall-vault cert list`

`~/.wall-vault/` अन्तर्गत प्रत्येक cert लाई subject, validity window, र SANs सहित देखाउँछ।

```
$ wall-vault cert list
ca.crt          subject=wall-vault internal CA   not-after=2036-05-05
hostname.crt    subject=hostname                 not-after=2031-05-05   SAN=hostname,localhost,127.0.0.1,192.168.…
```

---

## API key दर्ता गर्ने

दुई तरिका: dashboard, वा environment variables।

### Dashboard (सिफारिस गरिएको)

1. Admin token सँग `https://localhost:56243` मा लग इन गर्नुहोस्।
2. Keys card मा **+ API key** क्लिक गर्नुहोस्।
3. एक सेवा छान्नुहोस् (Google, OpenRouter, Anthropic, OpenAI, …)।
4. Key पेस्ट गर्नुहोस्। बचत गर्नुहोस्।

प्रति सेवा धेरै keys ठीक छन्; proxy ले तिनीहरू बीच round-robin गर्छ र per-key cooldown मा पुगेकाहरूलाई स्किप गर्छ।

### Environment variables (one-shot bootstrap)

```bash
export WV_KEY_GOOGLE="AIzaSyA1...,AIzaSyB2...,AIzaSyC3..."   # comma-separated
export WV_KEY_OPENROUTER="sk-or-v1-…"
export WV_KEY_ANTHROPIC="sk-ant-…"
export WV_KEY_OPENAI="sk-…"
wall-vault start
```

यो तरिकाले प्रदान गरिएका keys लाई पहिलो launch मा encrypted store मा लेखिन्छ। पछिका starts ले तिनीहरूलाई disk बाट पढ्छ; तपाईंले पहिलो run पछि env vars unset गर्न सक्नुहुन्छ।

### Cooldowns र rotation

प्रत्येक सफल call ले key को `usage_count` बढाउँछ र `last_used` ताजा बनाउँछ। HTTP 429 / 402 / 403 मा, proxy ले key लाई **cooldown** मा राख्छ (पूर्वनिर्धारित: 429 को लागि 60 मिनेट, 402 को लागि 24 घण्टा, 403 को लागि 12 घण्टा)। अर्को dispatch ले त्यो सेवाको लागि फरक key छान्छ। जब सेवाको सबै keys cooldown मा हुन्छन्, proxy ले त्यो सेवालाई पूरै तीव्र-स्किप गर्छ र fallback chain मा अर्को provider प्रयास गर्छ।

Cooldowns countdown सँग dashboard मा प्रति-key देखिने हुन्छन्।

---

## Agents जोड्ने

### OpenClaw

OpenClaw मूल लक्षित client हो। Dashboard को **+ Add agent** modal प्रयोग गर्नुहोस्:

- **Agent type** लाई `openclaw` वा `nanoclaw` मा सेट गर्नुहोस्।
- **Work directory** सेट गर्नुहोस् — OpenClaw को लागि यो स्वतः `~/.openclaw` को रूपमा भरिन्छ।
- एक **preferred service** र वैकल्पिक रूपमा **model override** छान्नुहोस्।
- **Apply** क्लिक गर्नुहोस्। wall-vault ले `~/.openclaw/openclaw.json` सीधै लेख्छ (provider URLs, vault token, model entries)।

जब तपाईंले dashboard बाट model परिवर्तन गर्नुहुन्छ, OpenClaw ले 1–3 सेकेन्ड भित्र SSE मार्फत परिवर्तन उठाउँछ — restart छैन।

### Claude Code

```bash
export ANTHROPIC_BASE_URL=https://localhost:56244
export ANTHROPIC_API_KEY=<your-vault-client-token>
claude
```

जब upstream Anthropic credits सकिन्छन्, dispatch यस client को `fallback_services` मा सूचीबद्ध भएका जुनसुकै सेवामा फलब्याक हुन्छ। पूर्वनिर्धारित रूपमा, anthropic dispatch मा पठाइएको गैर-Claude model id ले error फिर्ता गर्छ ताकि misrouting तुरुन्तै सतहमा आओस्। स्वचालित पुनर्लेखनमा opt in गर्नुहोस्:

```yaml
proxy:
  anthropic_fallback_model: "claude-haiku-4-5-20251001"
```

### Cursor

Cursor **Settings → AI → OpenAI API** मा:

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

`proxy.oai_stream_forward: true` सेट हुँदा उही endpoint ले streaming (`"stream": true`) स्वीकार गर्छ।

---

## Dashboard

`https://localhost:56243`। Home grid मा पाँच cards:

- **Keys** — प्रत्येक API key, सेवाद्वारा समूहीकृत। थप्नुहोस्, सम्पादन गर्नुहोस्, मेटाउनुहोस्; प्रयोग र cooldown हेर्नुहोस्।
- **Services** — Google / OpenRouter / Anthropic / OpenAI / Ollama / LM Studio / vLLM / llama.cpp, साथै `~/.wall-vault/services/` मा कुनै पनि plugin yaml। प्रति-सेवा `default_model`, `allowed_models`, base URL, reasoning toggle सेट गर्नुहोस्।
- **Clients (agents)** — प्रत्येक दर्ता गरिएको client (OpenClaw bot, Claude Code session, Cursor instance, …)। Preferred service, model override, fallback chain असाइन गर्नुहोस्।
- **Proxies** — यस vault विरुद्ध authenticate गरेका प्रत्येक proxy। प्रत्यक्ष स्थिति (online/offline), अन्तिम पटक देखिएको, हालको model।
- **Settings** — admin token, master password rotation, theme, language।

प्रत्येक card मा सम्पादन slideover छ (दायाँ छेउ)। बाहिर-क्लिक वा `Esc` ले बन्द गर्छ। परिवर्तनहरू सेकेन्ड भित्र SSE मार्फत सबै जोडिएका proxies मा push गरिन्छ।

**Footer** ले SSE indicator (हरियो = जोडिएको, सुन्तला = पुन: जोड्दै, खैरो = असम्बद्ध) र प्रत्यक्ष build version बोक्छ।

---

## Distributed मोड

जब तपाईंसँग धेरै मेसिनहरू छन् जसलाई सबैलाई उही keys चाहिन्छ, एउटा host मा vault र अरू प्रत्येकमा proxies चलाउनुहोस्।

### Vault host

```bash
WV_VAULT_HOST=0.0.0.0 \
WV_ADMIN_TOKEN=<admin> \
WV_MASTER_PASS=<master> \
wall-vault vault
```

Dashboard अब `https://<vault-host>:56243` मा पहुँचयोग्य छ। **Clients** card मा प्रत्येक रिमोट proxy को लागि एउटा agent थप्नुहोस्; प्रत्येकले एउटा अद्वितीय `vault_token` बनाउँछ।

### Proxy hosts

```bash
WV_VAULT_URL=http://<vault-host>:56243 \
WV_VAULT_TOKEN=<that-client-token> \
WV_PROXY_HOST=0.0.0.0 \
wall-vault proxy
```

Proxy ले vault विरुद्ध authenticate गर्छ, SSE stream खोल्छ, र यसले प्राप्त गर्ने कुनै पनि config लागू गर्छ (preferred service, model override, fallback chain)। पछिल्ला vault edits सेकेन्डमा बिना restart अवतरण गर्छन्।

LAN-फैलिने installs को लागि, vault host मा TLS सक्षम गर्नुहोस् (`WV_VAULT_TLS_ENABLED=1` + cert/key env vars) र प्रत्येक proxy host लाई उही `wall-vault cert install-trust` चरणबाट चलाउनुहोस् ताकि vault मा proxy को HTTPS कलहरू trusted होउन्।

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

उही host मा vault को लागि, समानान्तर `wall-vault-vault.service` लेख्नुहोस्। Standalone मोडको लागि, `wall-vault start` कल गर्ने एक unit पर्याप्त छ।

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

`wall-vault.exe start` लाई Windows service को रूपमा wrap गर्न `nssm` प्रयोग गर्नुहोस्, वा user logon मा चल्ने `schtasks` entry।

---

## Plugin yamls

कुनै पनि OpenAI-compatible backend लाई `~/.wall-vault/services/` अन्तर्गत yaml छाडेर code परिवर्तन बिना थप्न सकिन्छ। wall-vault ले यसलाई startup मा लोड गर्छ र dispatch, OAI-compat detection set, र Gemini-stream bridge को लागि सेवा दर्ता गर्छ।

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

`configs/services/` मा बन्डल गरिएको सेट (lmstudio, vllm, llamacpp, tgwui, localai, jan, koboldcpp, tabbyapi, mlx-server, litellm-proxy, ollama, google, openrouter) पूर्वनिर्धारित रूपमा असक्षम पठाउँछ। तपाईंले चाहेको एउटालाई `~/.wall-vault/services/` मा कपी गर्नुहोस्, `enabled: true` सेट गर्नुहोस्, restart गर्नुहोस्।

---

## Doctor

`wall-vault doctor` ले सम्पूर्ण install भर एक-शट health probe चलाउँछ:

```
✓ vault listener  (https://localhost:56243)
✓ proxy listener  (https://localhost:56244)
✓ master password set
⚠ Google: 2 keys, all on cooldown
✓ Anthropic: 1 key healthy
✗ Ollama: not reachable at http://localhost:11434
```

प्रत्येक लाइन निम्न मध्ये एक हो:

- `✓` — स्वस्थ
- `⚠` — खराब तर कार्यशील (एक key cooldown मा परेको, low quota, इत्यादि)
- `✗` — टुटेको
- `SKIP` — कन्फिगर नगरिएको / यस host मा लागू नहुने

दोस्रो daemon मोडले उही probe लाई प्रत्येक `doctor.interval` (पूर्वनिर्धारित 5 मिनेट) मा चलाउँछ र परिणामहरू `doctor.log_file` (पूर्वनिर्धारित `/tmp/wall-vault-doctor.log`) मा लेख्छ। जब `doctor.auto_fix` true हुन्छ, यसले सामान्य drift पनि मर्मत गर्ने प्रयास गर्छ (पुरानो OpenClaw config, हराइरहेको TLS trust, restartable services)।

Dashboard बाट **Doctor** card वा `wall-vault doctor` मार्फत one-shot ट्रिगर गर्नुहोस्।

---

## Hooks

Key events मा shell command चलाउनुहोस्:

```yaml
hooks:
  on_model_change:   "logger 'wall-vault: $SERVICE/$MODEL'"
  on_key_exhausted:  "notify-send 'wall-vault' '$SERVICE keys all on cooldown'"
  on_service_down:   "/usr/local/bin/page-oncall.sh $SERVICE '$ERROR'"
  on_doctor_fix:     "echo \"$AGENT: $LEVEL $MSG\" >> ~/wall-vault.audit.log"
  openclaw_socket:   ""    # if set, OpenClaw TUI receives events over this Unix socket
```

प्रत्येक hook ले event-specific environment variables (`SERVICE`, `MODEL`, `ERROR`, `AGENT`, `LEVEL`, `MSG`) पाउँछ। Hooks 5-सेकेन्ड timeout सँग async चल्छन् — proxy ले कहिल्यै ढिलो hook मा block गर्दैन।

---

## वातावरण चरहरू

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

प्रत्येक env var, सेट गर्दा, YAML file माथि जित्छ।

---

## समस्या समाधान

### `:56244` मा `connection refused`

या त proxy चलिरहेको छैन वा यो फरक host मा बाँधिएको छ। जाँच गर्नुहोस्:

```bash
ss -lnp | grep 56244
systemctl --user status wall-vault-proxy   # Linux
launchctl list | grep wall-vault           # macOS
```

यदि यो फरक port मा चलिरहेको छ भने, तपाईंको config मा `proxy.port` override गरिएको छ — `~/.wall-vault/config.yaml` जाँच गर्नुहोस्।

### `x509: certificate signed by unknown authority`

Client ले wall-vault आन्तरिक CA लाई trust गर्दैन। Client मेसिनमा `wall-vault cert install-trust` चलाउनुहोस्। OS trust store लाई बेवास्ता गर्ने runtime भएका agents को लागि (जस्तै hardcoded `NODE_EXTRA_CA_CERTS` सहितको Node), `127.0.0.1:56245` (केवल same-host) मा loopback HTTP companion प्रयोग गर्नुहोस् वा plain HTTP मा फलब्याक गर्न `WV_PROXY_TLS_ENABLED=0` सेट गर्नुहोस्।

### `token not registered with vault`

Client को `Authorization: Bearer <token>` कुनै पनि दर्ता गरिएको client सँग मेल खाँदैन। Dashboard मा **Clients** अन्तर्गत token प्रमाणित गर्नुहोस्। यदि तपाईंले पुरानो config बाट `proxy-managed`, `dummy`, वा `""` जस्तो literal token कपी गर्नुभयो भने, यसलाई वास्तविक client token सँग बदल्नुहोस्।

### `Anthropic dispatch needs a Claude model id`

v0.2.63 अनुसार पूर्वनिर्धारित व्यवहार: anthropic dispatch मा पठाइएको गैर-Claude model id ले error फिर्ता गर्छ। या त routing सच्याउनुहोस् (anthropic मा `gemini-2.5-flash` नपठाउनुहोस्) वा `proxy.anthropic_fallback_model` मार्फत स्वचालित पुनर्लेखनमा opt in गर्नुहोस्।

### `unknown service: <id>`

Dispatch ले एउटा service id देख्यो जुन कुनै plugin yaml ले दाबी गरेन। जाँच गर्नुहोस्:

```bash
ls ~/.wall-vault/services/        # any plugin yaml present?
cat ~/.wall-vault/services/<id>.yaml | grep enabled
```

यदि yaml अवस्थित छ तर `enabled: false` छ भने, यसलाई पल्टाउनुहोस्। यदि यो पूरै हराएको छ भने, source tree मा `configs/services/` बाट कपी गर्नुहोस्।

### Reasoning model मा खाली प्रतिक्रिया

`qwen3.6`, `deepseek-r1`, र GPT-`o1` परिवारले कहिलेकाहीं `reasoning_content` मात्र emit गर्छ र `content` खाली छोड्छ। v0.2.63 अनुसार wall-vault ले स्वचालित रूपमा reasoning text मा फलब्याक गर्छ — यदि तपाईंले अझै खाली responses देख्नुहुन्छ भने, backend ले कुनै पनि field फिर्ता गरिरहेको छैन। Upstream को logs जाँच गर्नुहोस्।

LM Studio सँग qwen3 को लागि विशेष रूपमा, plugin yaml मा `inline_no_think_for_qwen3: true` सेट गर्नुहोस् ताकि reasoning inline रूपमा असक्षम होस्। Built-in lmstudio.yaml र ollama.yaml ले पहिले नै यो गर्छन्।

### Dashboard ले "all keys on cooldown" देखाउँछ तर मैले भर्खरै एउटा थपेँ

नयाँ key स्वस्थ छ तर dispatch path अझै पुरानो key को लागि cooldown मा हुन सक्छ। एक नयाँ अनुरोध गर्नुहोस् — proxy प्रति कल round-robins गर्छ, र एक स्वस्थ key अर्को छानिनेछ।

### Vault master password सँग unlock हुँदैन

गलत पासवर्ड। कुनै रिकभरी छैन — wall-vault ले जानाजानी backdoor समावेश गर्दैन। यदि तपाईंले साँच्चै master password गुमाउनुभयो भने, एक मात्र मार्ग `~/.wall-vault/data/vault.json` मेटाउने, नयाँ पासवर्डसँग restart गर्ने, र keys फेरि थप्ने हो।

### Free-tier OpenRouter सीमाहरू पुगियो

`proxy.services` मा `openrouter` समावेश गर्न सेट गर्नुहोस् र कम्तिमा एउटा OpenRouter key थप्नुहोस्। जब paid path ले 402 / 429 फिर्ता गर्छ, proxy ले paid model बाट यसको `:free` variant मा स्वतः फलब्याक गर्छ।

### `journalctl --user -u wall-vault-proxy` खाली छ

systemd `--user` logs यो चलाइरहेको user को journal मा जान्छ। यदि तपाईंले unit लाई `root` को रूपमा वा `sudo` मार्फत सुरु गर्नुभयो भने, journal यसको सट्टा system instance मा छ — `--user` बिना `journalctl -u wall-vault-proxy` प्रयास गर्नुहोस्।

---

## थप

- HTTP API reference — [API.md](API.md) हेर्नुहोस्
- Source — `https://github.com/sookmook/wall-vault`
- Bug reports / feature requests — GitHub Issues
- Release इतिहास — [CHANGELOG.md](../CHANGELOG.md)
