# wall-vault उपयोगकर्ता मैनुअल

[English](MANUAL.md) · [한국어](MANUAL.ko.md) · [中文](MANUAL.zh.md) · [日本語](MANUAL.ja.md) · [Español](MANUAL.es.md) · [Français](MANUAL.fr.md) · [Deutsch](MANUAL.de.md) · [Português](MANUAL.pt.md) · [العربية](MANUAL.ar.md) · **हिन्दी** · [Bahasa Indonesia](MANUAL.id.md) · [ภาษาไทย](MANUAL.th.md) · [Kiswahili](MANUAL.sw.md) · [Hausa](MANUAL.ha.md) · [नेपाली](MANUAL.ne.md) · [Монгол](MANUAL.mn.md) · [isiZulu](MANUAL.zu.md)

यह मैनुअल wall-vault को इंस्टॉल करने, कॉन्फ़िगर करने और संचालित करने को कवर करता है। एक नज़र में अवलोकन के लिए [README](../README.md) देखें। HTTP API विवरण के लिए [API reference](API.md) देखें।

## विषय-सूची

1. [wall-vault क्या करता है](#wall-vault-क्या-करता-है)
2. [इंस्टॉलेशन](#इंस्टॉलेशन)
3. [सेटअप विज़ार्ड के साथ पहली बार चलाना](#सेटअप-विज़ार्ड-के-साथ-पहली-बार-चलाना)
4. [TLS सक्षम करना](#tls-सक्षम-करना)
5. [API कुंजियाँ रजिस्टर करना](#api-कुंजियाँ-रजिस्टर-करना)
6. [एजेंट कनेक्ट करना](#एजेंट-कनेक्ट-करना)
7. [डैशबोर्ड](#डैशबोर्ड)
8. [वितरित मोड](#वितरित-मोड)
9. [ऑटो-स्टार्ट](#ऑटो-स्टार्ट)
10. [प्लगइन yamls](#प्लगइन-yamls)
11. [Doctor](#doctor)
12. [Hooks](#hooks)
13. [पर्यावरण चर](#पर्यावरण-चर)
14. [समस्या निवारण](#समस्या-निवारण)

---

## wall-vault क्या करता है

wall-vault एक एकल Go बायनरी है जो दो सहयोगी सेवाओं को एक साथ जोड़ता है:

- **vault** API कुंजियों को विश्राम पर एन्क्रिप्टेड (मास्टर पासवर्ड के साथ AES-GCM) संग्रहीत करता है, प्रति कुंजी उपयोग और कूलडाउन को ट्रैक करता है, परिवर्तनों को Server-Sent Events (SSE) पर प्रसारित करता है, और मानव ऑपरेटरों के लिए `:56243` पर एक वेब डैशबोर्ड प्रदान करता है।
- **proxy** Gemini, Anthropic, OpenAI-संगत और Ollama-नेटिव एंडपॉइंट्स को `:56244` पर एक्सपोज़ करता है। कोई भी AI क्लाइंट जो प्रॉक्सी की ओर इशारा करता है वह vault की कुंजियों का उपयोग कर रहा है — क्लाइंट उन्हें कभी नहीं देखते। जब एक upstream विफल होता है, तो डिस्पैच क्रम में अगले प्रदाता पर वापस आ जाता है।

यह उपयोगी है जब:

- आपके पास कई प्रदाताओं के लिए कुंजियाँ हैं और आप एक URL चाहते हैं जिससे एजेंट बात करे।
- आप चाहते हैं कि कूलडाउन पर मुफ्त-स्तरीय कुंजी सत्र को बाधित किए बिना एक तरफ हो जाए।
- आप चाहते हैं कि वही कुंजियाँ क्रेडेंशियल कॉपी किए बिना एक ही LAN पर कई बॉट्स, IDEs, या स्क्रिप्ट्स को चलाएँ।
- आप कुंजियों को संपादित करने और मॉडल बदलने के लिए डैशबोर्ड चाहते हैं, पर्यावरण चर नहीं।
- आप क्लाउड सीमाएँ समाप्त होने पर एक स्थानीय फ़ॉलबैक (Ollama, LM Studio, vLLM) चाहते हैं।

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

## इंस्टॉलेशन

### Linux / macOS एक-लाइनर

```bash
curl -fsSL https://raw.githubusercontent.com/sookmook/wall-vault/main/install.sh | sh
```

स्क्रिप्ट OS और आर्किटेक्चर का स्वचालित रूप से पता लगाती है, सही बायनरी को `~/.local/bin/wall-vault` में डाउनलोड करती है, और इसे निष्पादन योग्य बनाती है। यदि `~/.local/bin` आपके `PATH` पर नहीं है, तो इसे जोड़ें:

```bash
export PATH="$HOME/.local/bin:$PATH"
```

### मैनुअल डाउनलोड

प्रत्येक रिलीज़ पर पूर्व-निर्मित बायनरी `https://github.com/sookmook/wall-vault/releases` पर प्रकाशित की जाती हैं।

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

### स्रोत से बनाएँ

Go 1.25 या नया आवश्यक है।

```bash
git clone https://github.com/sookmook/wall-vault
cd wall-vault
go build -o wall-vault .
```

`make build-all` सभी पाँच समर्थित प्लेटफ़ॉर्म के लिए क्रॉस-कंपाइल करता है। बायनरी `bin/` में आती हैं।

---

## सेटअप विज़ार्ड के साथ पहली बार चलाना

```bash
wall-vault setup
```

विज़ार्ड आपसे क्रम में पूछता है:

1. **भाषा** — 17 UI लोकेल में से एक चुनता है। `$LANG` से स्वचालित रूप से पता लगाया जाता है; विज़ार्ड फिर भी एक सूची प्रदान करता है।
2. **थीम** — `light` (डिफ़ॉल्ट), `dark`, `cherry`, `ocean`, `gold`, `autumn`, `winter`. केवल कॉस्मेटिक।
3. **मोड** — `standalone` (एकल होस्ट, डिफ़ॉल्ट) या `distributed` (एक होस्ट पर vault, अन्य पर प्रॉक्सी)।
4. **बॉट का नाम** — एक स्वतंत्र-रूप `client_id` स्लग। vault इसका उपयोग प्रति-क्लाइंट कॉन्फ़िगरेशन (मॉडल ओवरराइड, फ़ॉलबैक चेन) को स्कोप करने के लिए करता है।
5. **प्रॉक्सी पोर्ट** — डिफ़ॉल्ट `56244`।
6. **vault पोर्ट** — डिफ़ॉल्ट `56243` (केवल standalone)।
7. **सेवा चयन** — प्रत्येक के लिए y/N: Google Gemini, OpenRouter, Anthropic, OpenAI, Ollama, LM Studio, vLLM. कई विकल्प ठीक हैं; प्रत्येक अंत में अपना env-var संकेत लिखता है।
8. **टूल फ़िल्टर** — `strip_all` (डिफ़ॉल्ट; सुरक्षा के लिए सभी आने वाली टूल परिभाषाओं को ब्लॉक करता है) या `passthrough` (किसी भी टूल को गुज़रने दें)।
9. **एडमिन टोकन** — स्वत: उत्पन्न करने के लिए खाली छोड़ दें। डैशबोर्ड को लॉग इन करने के लिए इस टोकन की आवश्यकता है।
10. **मास्टर पासवर्ड** — कोई एन्क्रिप्शन नहीं के लिए खाली छोड़ दें (अनुशंसित नहीं); विश्राम पर कुंजी संग्रह को AES-GCM एन्क्रिप्ट करने के लिए एक मान सेट करें।
11. **सहेजने का पथ** — वर्तमान निर्देशिका में डिफ़ॉल्ट रूप से `wall-vault.yaml`। लोडर `~/.wall-vault/config.yaml` को भी देखता है।

सहेजने के बाद, विज़ार्ड `doctor.FixTrust` चलाता है ताकि कोई भी स्थानीय रूप से इंस्टॉल किया गया एजेंट (OpenClaw, Claude Code, Cline) स्वचालित रूप से अपने ट्रस्ट स्टोर में जोड़ा गया wall-vault आंतरिक CA प्राप्त करे। यदि ऐसा कोई एजेंट इंस्टॉल नहीं है, तो चरण `SKIP` प्रिंट करता है और कुछ नहीं लिखता।

फिर बायनरी शुरू करें:

```bash
wall-vault start
```

`start` एक प्रक्रिया (standalone मोड) में vault और प्रॉक्सी दोनों चलाता है। वितरित मोड के लिए vault होस्ट पर `wall-vault vault` और प्रत्येक प्रॉक्सी होस्ट पर `wall-vault proxy` का उपयोग करें।

ब्राउज़र में `http://localhost:56243` खोलें। विज़ार्ड द्वारा प्रिंट किए गए एडमिन टोकन के साथ लॉग इन करें।

---

## TLS सक्षम करना

विज़ार्ड के डिफ़ॉल्ट दोनों listeners को सादे HTTP पर छोड़ देते हैं। अधिकांश एजेंट (OpenClaw, Claude Code, Cursor) एकल HTTPS एंडपॉइंट के विरुद्ध बेहतर काम करते हैं, इसलिए स्थानीय मशीन से अधिक तक फैले किसी भी डिप्लॉयमेंट में TLS की अनुशंसा की जाती है।

wall-vault अपने स्वयं के आंतरिक CA के साथ शिप करता है ताकि आपको सार्वजनिक DNS नाम या Let's Encrypt की आवश्यकता न हो।

```bash
# 1. Create the internal CA — written to ~/.wall-vault/ca.{crt,key}.
#    The CA is good for 10 years by default; override with --ca-years.
wall-vault cert init

# 2. Issue a host certificate. Subject Alternative Names automatically include:
#       hostname, "localhost", "127.0.0.1", and any non-loopback LAN IP detected.
#    Override the issuer dir with --dir, validity with --host-years.
wall-vault cert issue $(hostname)

# 3. Trust the CA in this machine's OS keychain.
#    Linux: writes to /etc/ssl/certs/ via update-ca-certificates (needs sudo).
#    macOS: adds to the System keychain via security add-trusted-cert (needs sudo).
#    Windows: imports into CurrentUser\Root via certutil (no admin needed).
wall-vault cert install-trust

# 4. Enable TLS on both listeners.
export WV_PROXY_TLS_ENABLED=1
export WV_PROXY_TLS_CERT="$HOME/.wall-vault/$(hostname).crt"
export WV_PROXY_TLS_KEY="$HOME/.wall-vault/$(hostname).key"
export WV_VAULT_TLS_ENABLED=1
export WV_VAULT_TLS_CERT="$HOME/.wall-vault/$(hostname).crt"
export WV_VAULT_TLS_KEY="$HOME/.wall-vault/$(hostname).key"

wall-vault start
```

अन्य LAN मशीनों तक विश्वास का विस्तार करने के लिए, `~/.wall-vault/ca.crt` को कॉपी करें और प्रत्येक पर `wall-vault cert install-trust --ca <path>` चलाएँ। vault `:56247` पर एक छोटे प्लेन-HTTP listener के माध्यम से `ca.crt` को भी एक्सपोज़ करता है (**bootstrap port**) catch-22 मामले के लिए जहाँ एक नए क्लाइंट को HTTPS बात करने के लिए CA की आवश्यकता होती है।

### Loopback HTTP साथी

कुछ एजेंट — विशेष रूप से OpenClaw का बंडल किया गया Node रनटाइम — प्रोसेस स्पॉन पर `NODE_EXTRA_CA_CERTS` को फिर से लिखते हैं, किसी भी ऑपरेटर-आपूर्ति किए गए CA संकेत को छोड़ देते हैं। वे `cert install-trust` के बाद भी, daemon के अंदर से wall-vault CA का सम्मान नहीं कर सकते। wall-vault इसके चारों ओर TLS सक्षम होने पर `127.0.0.1:56245` पर एक अतिरिक्त **loopback-only plain-HTTP listener** बाँधकर काम करता है। एक ही होस्ट के क्लाइंट उस पोर्ट के माध्यम से बिना किसी TLS के प्रॉक्सी तक पहुँचते हैं; LAN क्लाइंट TLS listener का उपयोग करना जारी रखते हैं।

यदि आपको इसकी आवश्यकता नहीं है तो `WV_PROXY_PLAIN_PORT=0` के साथ अक्षम करें।

### `wall-vault cert list`

`~/.wall-vault/` के अंतर्गत प्रत्येक प्रमाणपत्र को विषय, वैधता विंडो और SANs के साथ दिखाता है।

```
$ wall-vault cert list
ca.crt          subject=wall-vault internal CA   not-after=2036-05-05
hostname.crt    subject=hostname                 not-after=2031-05-05   SAN=hostname,localhost,127.0.0.1,192.168.…
```

---

## API कुंजियाँ रजिस्टर करना

दो तरीके: डैशबोर्ड, या पर्यावरण चर।

### डैशबोर्ड (अनुशंसित)

1. एडमिन टोकन के साथ `https://localhost:56243` पर लॉग इन करें।
2. कुंजी कार्ड में **+ API key** क्लिक करें।
3. एक सेवा चुनें (Google, OpenRouter, Anthropic, OpenAI, …)।
4. कुंजी पेस्ट करें। सहेजें।

प्रति सेवा कई कुंजियाँ ठीक हैं; प्रॉक्सी उनके बीच राउंड-रॉबिन करता है और प्रति-कुंजी कूलडाउन से टकराने वालों को छोड़ देता है।

### पर्यावरण चर (वन-शॉट bootstrap)

```bash
export WV_KEY_GOOGLE="AIzaSyA1...,AIzaSyB2...,AIzaSyC3..."   # comma-separated
export WV_KEY_OPENROUTER="sk-or-v1-…"
export WV_KEY_ANTHROPIC="sk-ant-…"
export WV_KEY_OPENAI="sk-…"
wall-vault start
```

इस तरह प्रदान की गई कुंजियाँ पहले लॉन्च पर एन्क्रिप्टेड स्टोर में लिखी जाती हैं। बाद के स्टार्ट उन्हें डिस्क से पढ़ते हैं; पहले रन के बाद आप env चर को unset कर सकते हैं।

### कूलडाउन और रोटेशन

प्रत्येक सफल कॉल कुंजी के `usage_count` को बढ़ाती है और `last_used` को रिफ्रेश करती है। HTTP 429 / 402 / 403 पर, प्रॉक्सी कुंजी को **कूलडाउन** पर रखता है (डिफ़ॉल्ट: 429 के लिए 60 मिनट, 402 के लिए 24 घंटे, 403 के लिए 12 घंटे)। अगला डिस्पैच उस सेवा के लिए एक अलग कुंजी चुनता है। जब किसी सेवा के लिए सभी कुंजियाँ कूलडाउन पर होती हैं, तो प्रॉक्सी उस सेवा को पूरी तरह से तेज़ी से छोड़ देता है और फ़ॉलबैक चेन में अगले प्रदाता को आज़माता है।

कूलडाउन डैशबोर्ड में काउंटडाउन के साथ प्रति-कुंजी दिखाई देते हैं।

---

## एजेंट कनेक्ट करना

### OpenClaw

OpenClaw मूल लक्षित क्लाइंट है। डैशबोर्ड के **+ Add agent** मोडल का उपयोग करें:

- **Agent type** को `openclaw` या `nanoclaw` पर सेट करें।
- **Work directory** सेट करें — OpenClaw के लिए यह स्वचालित रूप से `~/.openclaw` के रूप में भर जाता है।
- एक **preferred service** चुनें और वैकल्पिक रूप से एक **model override**।
- **Apply** क्लिक करें। wall-vault सीधे `~/.openclaw/openclaw.json` लिखता है (प्रदाता URLs, vault token, मॉडल entries)।

जब आप डैशबोर्ड से मॉडल बदलते हैं, तो OpenClaw 1-3 सेकंड के भीतर SSE पर परिवर्तन उठा लेता है — कोई पुनः आरंभ नहीं।

### Claude Code

```bash
export ANTHROPIC_BASE_URL=https://localhost:56244
export ANTHROPIC_API_KEY=<your-vault-client-token>
claude
```

जब upstream Anthropic क्रेडिट समाप्त हो जाते हैं, तो डिस्पैच इस क्लाइंट के `fallback_services` में सूचीबद्ध जो भी सेवाएँ हैं, उन पर वापस आ जाता है। डिफ़ॉल्ट रूप से, anthropic डिस्पैच को भेजी गई एक गैर-Claude मॉडल आईडी एक त्रुटि लौटाती है ताकि गलत मार्गनिर्देशन तुरंत सतह पर आ जाए। स्वचालित पुनर्लेखन में ऑप्ट-इन करें:

```yaml
proxy:
  anthropic_fallback_model: "claude-haiku-4-5-20251001"
```

### Cursor

Cursor **Settings → AI → OpenAI API** में:

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

### कस्टम HTTP

```bash
curl -X POST https://localhost:56244/v1/chat/completions \
  -H "Authorization: Bearer <your-vault-client-token>" \
  -H "Content-Type: application/json" \
  -d '{
    "model": "gemini-2.5-flash",
    "messages": [{"role": "user", "content": "hello"}]
  }'
```

जब `proxy.oai_stream_forward: true` सेट है तो वही एंडपॉइंट स्ट्रीमिंग (`"stream": true`) स्वीकार करता है।

---

## डैशबोर्ड

`https://localhost:56243`. होम ग्रिड पर पाँच कार्ड:

- **Keys** — प्रत्येक API कुंजी, सेवा द्वारा समूहीकृत। जोड़ें, संपादित करें, हटाएँ; उपयोग और कूलडाउन देखें।
- **Services** — Google / OpenRouter / Anthropic / OpenAI / Ollama / LM Studio / vLLM / llama.cpp, और `~/.wall-vault/services/` में कोई भी प्लगइन yaml। प्रति-सेवा `default_model`, `allowed_models`, बेस URL, reasoning टॉगल सेट करें।
- **Clients (agents)** — प्रत्येक रजिस्टर्ड क्लाइंट (OpenClaw बॉट, Claude Code सत्र, Cursor इंस्टेंस, …)। पसंदीदा सेवा, मॉडल ओवरराइड, फ़ॉलबैक चेन असाइन करें।
- **Proxies** — हर प्रॉक्सी जिसने इस vault के विरुद्ध प्रमाणित किया है। लाइव स्थिति (online/offline), अंतिम बार देखा गया, वर्तमान मॉडल।
- **Settings** — एडमिन टोकन, मास्टर पासवर्ड रोटेशन, थीम, भाषा।

प्रत्येक कार्ड में एक संपादन slideover (दाहिनी ओर) है। बाहर-क्लिक या `Esc` इसे बंद कर देता है। परिवर्तन कुछ सेकंड के भीतर SSE पर सभी कनेक्ट किए गए प्रॉक्सी पर पुश किए जाते हैं।

**फ़ुटर** एक SSE संकेतक (हरा = कनेक्टेड, नारंगी = फिर से कनेक्ट हो रहा है, ग्रे = डिस्कनेक्ट) और लाइव बिल्ड संस्करण रखता है।

---

## वितरित मोड

जब आपके पास कई मशीनें हैं जिन्हें सभी को समान कुंजियों की आवश्यकता है, तो एक होस्ट पर vault और बाकी प्रत्येक पर प्रॉक्सी चलाएँ।

### vault होस्ट

```bash
WV_VAULT_HOST=0.0.0.0 \
WV_ADMIN_TOKEN=<admin> \
WV_MASTER_PASS=<master> \
wall-vault vault
```

डैशबोर्ड अब `https://<vault-host>:56243` पर पहुँच योग्य है। **Clients** कार्ड में प्रत्येक रिमोट प्रॉक्सी के लिए एक एजेंट जोड़ें; प्रत्येक एक अद्वितीय `vault_token` बनाता है।

### प्रॉक्सी होस्ट

```bash
WV_VAULT_URL=http://<vault-host>:56243 \
WV_VAULT_TOKEN=<that-client-token> \
WV_PROXY_HOST=0.0.0.0 \
wall-vault proxy
```

प्रॉक्सी vault के विरुद्ध प्रमाणित करता है, एक SSE स्ट्रीम खोलता है, और प्राप्त किसी भी कॉन्फ़िगरेशन को लागू करता है (पसंदीदा सेवा, मॉडल ओवरराइड, फ़ॉलबैक चेन)। बाद के vault संपादन बिना पुनः आरंभ के सेकंडों में आते हैं।

LAN-व्यापी इंस्टॉल के लिए, vault होस्ट पर TLS सक्षम करें (`WV_VAULT_TLS_ENABLED=1` + cert/key env चर) और प्रत्येक प्रॉक्सी होस्ट को उसी `wall-vault cert install-trust` चरण से चलाएँ ताकि vault में प्रॉक्सी की HTTPS कॉल पर भरोसा किया जा सके।

---

## ऑटो-स्टार्ट

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

उसी होस्ट पर vault के लिए, एक समानांतर `wall-vault-vault.service` लिखें। standalone मोड के लिए, एक यूनिट `wall-vault start` को कॉल करना पर्याप्त है।

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

`wall-vault.exe start` को Windows सेवा के रूप में wrap करने के लिए `nssm` का उपयोग करें, या एक `schtasks` प्रविष्टि जो उपयोगकर्ता लॉगऑन पर चलती है।

---

## प्लगइन yamls

`~/.wall-vault/services/` के अंतर्गत एक yaml ड्रॉप करके किसी भी OpenAI-संगत बैकएंड को बिना कोड परिवर्तन के जोड़ा जा सकता है। wall-vault इसे स्टार्टअप पर लोड करता है और सेवा को डिस्पैच, OAI-compat डिटेक्शन सेट और Gemini-stream ब्रिज के लिए रजिस्टर करता है।

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

`configs/services/` में बंडल किया गया सेट (lmstudio, vllm, llamacpp, tgwui, localai, jan, koboldcpp, tabbyapi, mlx-server, litellm-proxy, ollama, google, openrouter) डिफ़ॉल्ट रूप से अक्षम शिप होता है। आप जो चाहते हैं उसे `~/.wall-vault/services/` में कॉपी करें, `enabled: true` सेट करें, पुनः आरंभ करें।

---

## Doctor

`wall-vault doctor` पूरे इंस्टॉल पर एक बार का स्वास्थ्य परीक्षण चलाता है:

```
✓ vault listener  (https://localhost:56243)
✓ proxy listener  (https://localhost:56244)
✓ master password set
⚠ Google: 2 keys, all on cooldown
✓ Anthropic: 1 key healthy
✗ Ollama: not reachable at http://localhost:11434
```

प्रत्येक पंक्ति इनमें से एक है:

- `✓` — स्वस्थ
- `⚠` — खराब लेकिन कार्य कर रहा है (एक कुंजी कूल हो गई, कम कोटा, आदि।)
- `✗` — टूटा हुआ
- `SKIP` — कॉन्फ़िगर नहीं किया गया / इस होस्ट पर लागू नहीं

एक दूसरा daemon मोड हर `doctor.interval` (डिफ़ॉल्ट 5 मिनट) समान परीक्षण चलाता है और परिणामों को `doctor.log_file` (डिफ़ॉल्ट `/tmp/wall-vault-doctor.log`) में लिखता है। जब `doctor.auto_fix` true है, तो यह आम drift (बासी OpenClaw कॉन्फ़िग, गायब TLS विश्वास, पुनः आरंभ करने योग्य सेवाएँ) की मरम्मत करने का भी प्रयास करता है।

डैशबोर्ड से **Doctor** कार्ड या `wall-vault doctor` के माध्यम से वन-शॉट ट्रिगर करें।

---

## Hooks

मुख्य घटनाओं पर एक shell कमांड चलाएँ:

```yaml
hooks:
  on_model_change:   "logger 'wall-vault: $SERVICE/$MODEL'"
  on_key_exhausted:  "notify-send 'wall-vault' '$SERVICE keys all on cooldown'"
  on_service_down:   "/usr/local/bin/page-oncall.sh $SERVICE '$ERROR'"
  on_doctor_fix:     "echo \"$AGENT: $LEVEL $MSG\" >> ~/wall-vault.audit.log"
  openclaw_socket:   ""    # if set, OpenClaw TUI receives events over this Unix socket
```

प्रत्येक hook को घटना-विशिष्ट पर्यावरण चर मिलते हैं (`SERVICE`, `MODEL`, `ERROR`, `AGENT`, `LEVEL`, `MSG`)। Hooks 5-सेकंड टाइमआउट के साथ async चलते हैं — प्रॉक्सी कभी भी धीमे hook पर ब्लॉक नहीं होता।

---

## पर्यावरण चर

| चर | YAML फ़ील्ड |
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

प्रत्येक env चर, जब सेट किया जाता है, YAML फ़ाइल पर जीतता है।

---

## समस्या निवारण

### `:56244` पर `connection refused`

या तो प्रॉक्सी नहीं चल रहा है या यह एक अलग होस्ट से बाउंड है। जाँच करें:

```bash
ss -lnp | grep 56244
systemctl --user status wall-vault-proxy   # Linux
launchctl list | grep wall-vault           # macOS
```

यदि यह एक अलग पोर्ट पर चल रहा है, तो आपके कॉन्फ़िगरेशन में `proxy.port` ओवरराइड किया गया है — `~/.wall-vault/config.yaml` जाँचें।

### `x509: certificate signed by unknown authority`

क्लाइंट wall-vault आंतरिक CA पर भरोसा नहीं करता। क्लाइंट मशीन पर `wall-vault cert install-trust` चलाएँ। उन एजेंटों के लिए जिनका रनटाइम OS ट्रस्ट स्टोर को अनदेखा करता है (जैसे हार्डकोडेड `NODE_EXTRA_CA_CERTS` के साथ Node), `127.0.0.1:56245` (केवल same-host) पर loopback HTTP साथी का उपयोग करें या प्लेन HTTP पर वापस जाने के लिए `WV_PROXY_TLS_ENABLED=0` सेट करें।

### `token not registered with vault`

क्लाइंट का `Authorization: Bearer <token>` किसी रजिस्टर्ड क्लाइंट से मेल नहीं खाता। डैशबोर्ड में **Clients** के अंतर्गत टोकन सत्यापित करें। यदि आपने एक बासी कॉन्फ़िग से `proxy-managed`, `dummy`, या `""` जैसा एक टोकन शाब्दिक कॉपी किया है, तो इसे वास्तविक क्लाइंट टोकन से बदलें।

### `Anthropic dispatch needs a Claude model id`

v0.2.63 के अनुसार डिफ़ॉल्ट व्यवहार: anthropic डिस्पैच को भेजी गई एक गैर-Claude मॉडल आईडी एक त्रुटि लौटाती है। या तो रूटिंग को ठीक करें (`gemini-2.5-flash` को anthropic पर न भेजें) या `proxy.anthropic_fallback_model` के माध्यम से स्वचालित पुनर्लेखन में ऑप्ट-इन करें।

### `unknown service: <id>`

डिस्पैच ने एक सेवा आईडी देखी जिसका कोई प्लगइन yaml दावा नहीं करता। जाँच करें:

```bash
ls ~/.wall-vault/services/        # any plugin yaml present?
cat ~/.wall-vault/services/<id>.yaml | grep enabled
```

यदि yaml मौजूद है लेकिन `enabled: false` है, तो इसे फ्लिप करें। यदि यह पूरी तरह से गायब है, तो स्रोत ट्री में `configs/services/` से कॉपी करें।

### एक reasoning मॉडल पर खाली प्रतिक्रिया

`qwen3.6`, `deepseek-r1`, और GPT-`o1` परिवार कभी-कभी केवल `reasoning_content` उत्सर्जित करते हैं और `content` को खाली छोड़ देते हैं। v0.2.63 के अनुसार wall-vault स्वचालित रूप से reasoning टेक्स्ट पर वापस आता है — यदि आप अभी भी खाली प्रतिक्रियाएँ देखते हैं, तो बैकएंड कोई भी फ़ील्ड नहीं लौटा रहा है। upstream के लॉग जाँचें।

विशेष रूप से qwen3 के साथ LM Studio के लिए, प्लगइन yaml में `inline_no_think_for_qwen3: true` सेट करें ताकि reasoning inline अक्षम हो जाए। बिल्ट-इन lmstudio.yaml और ollama.yaml पहले से ही ऐसा करते हैं।

### डैशबोर्ड "all keys on cooldown" दिखाता है लेकिन मैंने अभी एक जोड़ा

नई कुंजी स्वस्थ है लेकिन डिस्पैच पथ अभी भी एक पुरानी कुंजी के लिए कूलडाउन में हो सकता है। एक नया अनुरोध आज़माएँ — प्रॉक्सी प्रति कॉल राउंड-रॉबिन करता है, और एक स्वस्थ कुंजी अगली बार चुनी जाएगी।

### vault मास्टर पासवर्ड के साथ अनलॉक नहीं होगा

गलत पासवर्ड। कोई पुनर्प्राप्ति नहीं है — wall-vault जानबूझकर बैकडोर शिप नहीं करता। यदि आपने वास्तव में मास्टर पासवर्ड खो दिया है, तो एकमात्र मार्ग `~/.wall-vault/data/vault.json` को हटाना, एक नए पासवर्ड के साथ पुनः आरंभ करना और कुंजियों को फिर से जोड़ना है।

### फ्री-टियर OpenRouter सीमाएँ हिट

`openrouter` को शामिल करने के लिए `proxy.services` सेट करें और कम से कम एक OpenRouter कुंजी जोड़ें। प्रॉक्सी एक भुगतान किए गए मॉडल से उसके `:free` संस्करण पर स्वत: फ़ॉलबैक करता है जब भुगतान किया गया पथ 402 / 429 लौटाता है।

### `journalctl --user -u wall-vault-proxy` खाली है

systemd `--user` लॉग इसे चलाने वाले उपयोगकर्ता के journal में जाते हैं। यदि आपने यूनिट को `root` के रूप में या `sudo` के माध्यम से शुरू किया है, तो journal इसके बजाय सिस्टम इंस्टेंस में है — `--user` के बिना `journalctl -u wall-vault-proxy` आज़माएँ।

---

## अधिक

- HTTP API रेफरेंस — [API.md](API.md) देखें
- स्रोत — `https://github.com/sookmook/wall-vault`
- बग रिपोर्ट / फ़ीचर अनुरोध — GitHub Issues
- रिलीज़ इतिहास — [CHANGELOG.md](../CHANGELOG.md)
