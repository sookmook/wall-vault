# wall-vault API मैनुअल

यह दस्तावेज़ wall-vault के सभी HTTP API एंडपॉइंट का विस्तृत विवरण प्रदान करता है।

---

## विषय सूची

- [प्रमाणीकरण](#प्रमाणीकरण)
- [प्रॉक्सी API (:56244)](#प्रॉक्सी-api-56244)
  - [हेल्थ चेक](#get-health)
  - [स्थिति जाँच](#get-status)
  - [मॉडल सूची](#get-apimodels)
  - [मॉडल बदलें](#put-apiconfigmodel)
  - [थिंकिंग मोड](#put-apiconfigthink-mode)
  - [कॉन्फ़िगरेशन रीलोड](#post-reload)
  - [Gemini API](#post-googlev1betamodelsmgeneratecontent)
  - [Gemini स्ट्रीमिंग](#post-googlev1betamodelsmstreamgeneratecontent)
  - [OpenAI संगत API](#post-v1chatcompletions)
- [की वॉल्ट API (:56243)](#की-वॉल्ट-api-56243)
  - [सार्वजनिक API](#सार्वजनिक-apiप्रमाणीकरण-अनावश्यक)
  - [SSE इवेंट स्ट्रीम](#get-apievents)
  - [प्रॉक्सी-विशिष्ट API](#प्रॉक्सी-विशिष्ट-apiक्लाइंट-टोकन)
  - [एडमिन API — कीज़](#एडमिन-api--api-कीज़)
  - [एडमिन API — क्लाइंट](#एडमिन-api--क्लाइंट)
  - [एडमिन API — सर्विसेज़](#एडमिन-api--सर्विसेज़)
  - [एडमिन API — मॉडल सूची](#एडमिन-api--मॉडल-सूची)
  - [एडमिन API — प्रॉक्सी स्थिति](#एडमिन-api--प्रॉक्सी-स्थिति)
- [SSE इवेंट प्रकार](#sse-इवेंट-प्रकार)
- [प्रोवाइडर·मॉडल राउटिंग](#प्रोवाइडरमॉडल-राउटिंग)
- [डेटा स्कीमा](#डेटा-स्कीमा)
- [त्रुटि प्रतिक्रिया](#त्रुटि-प्रतिक्रिया)
- [cURL उदाहरण संग्रह](#curl-उदाहरण-संग्रह)

---

## प्रमाणीकरण

| क्षेत्र | विधि | हेडर |
|--------|-------|------|
| एडमिन API | Bearer टोकन | `Authorization: Bearer <admin_token>` |
| प्रॉक्सी → वॉल्ट | Bearer टोकन | `Authorization: Bearer <client_token>` |
| प्रॉक्सी API | कोई नहीं (लोकल) | — |

जब `admin_token` सेट नहीं है (खाली स्ट्रिंग), सभी एडमिन API बिना प्रमाणीकरण के एक्सेस किए जा सकते हैं।

### सुरक्षा नीति

- **रेट लिमिटिंग**: एडमिन API प्रमाणीकरण 15 मिनट में 10 बार से अधिक विफल होने पर संबंधित IP को अस्थायी रूप से ब्लॉक किया जाता है (`429 Too Many Requests`)
- **IP व्हाइटलिस्ट**: एजेंट (`Client`) के `ip_whitelist` फ़ील्ड में पंजीकृत IP/CIDR ही `/api/keys` तक पहुँच सकते हैं। खाली ऐरे का मतलब सभी को अनुमति।
- **theme·lang सुरक्षा**: `/admin/theme`, `/admin/lang` के लिए भी एडमिन टोकन प्रमाणीकरण आवश्यक है

---

## प्रॉक्सी API (:56244)

प्रॉक्सी जिस सर्वर पर चलती है। डिफ़ॉल्ट पोर्ट `56244`।

---

### `GET /health`

हेल्थ चेक। हमेशा 200 OK लौटाता है।

**प्रतिक्रिया उदाहरण:**
```json
{
  "status": "ok",
  "version": "v0.1.6.20260314.231308",
  "client": "bot-a"
}
```

---

### `GET /status`

प्रॉक्सी स्थिति का विस्तृत विवरण।

**प्रतिक्रिया उदाहरण:**
```json
{
  "status": "ok",
  "version": "v0.1.6.20260314.231308",
  "client": "bot-a",
  "service": "google",
  "model": "gemini-2.5-flash",
  "sse": true,
  "filter": "strip_all",
  "services": ["google", "openrouter", "ollama"],
  "mode": "distributed"
}
```

| फ़ील्ड | प्रकार | विवरण |
|--------|--------|-------|
| `service` | string | वर्तमान डिफ़ॉल्ट सर्विस |
| `model` | string | वर्तमान डिफ़ॉल्ट मॉडल |
| `sse` | bool | वॉल्ट SSE कनेक्शन स्थिति |
| `filter` | string | टूल फ़िल्टर मोड |
| `services` | []string | सक्रिय सर्विसेज़ की सूची |
| `mode` | string | `standalone` \| `distributed` |

---

### `GET /api/models`

उपलब्ध मॉडल सूची की जाँच। TTL कैश (डिफ़ॉल्ट 10 मिनट) का उपयोग करता है।

**क्वेरी पैरामीटर:**

| पैरामीटर | विवरण | उदाहरण |
|---------|-------|--------|
| `service` | सर्विस फ़िल्टर | `?service=google` |
| `q` | मॉडल ID/नाम खोज | `?q=gemini` |

**प्रतिक्रिया उदाहरण:**
```json
{
  "models": [
    {
      "id": "gemini-2.5-pro",
      "name": "Gemini 2.5 Pro",
      "service": "google",
      "context_length": 1048576,
      "free": false
    },
    {
      "id": "openrouter/hunter-alpha",
      "name": "Hunter Alpha (1M ctx, free)",
      "service": "openrouter",
      "context_length": 1048576,
      "free": true
    }
  ],
  "count": 2
}
```

| फ़ील्ड | प्रकार | विवरण |
|--------|--------|-------|
| `id` | string | मॉडल ID |
| `name` | string | मॉडल प्रदर्शन नाम |
| `service` | string | `google` \| `openai` \| `anthropic` \| `openrouter` \| `ollama` आदि |
| `context_length` | int | कॉन्टेक्स्ट विंडो आकार |
| `free` | bool | मुफ़्त मॉडल है या नहीं (OpenRouter) |

---

### `PUT /api/config/model`

वर्तमान सर्विस·मॉडल बदलें।

**अनुरोध बॉडी:**
```json
{
  "service": "google",
  "model": "gemini-2.5-flash"
}
```

**प्रतिक्रिया:**
```json
{
  "status": "ok",
  "service": "google",
  "model": "gemini-2.5-flash"
}
```

> **नोट:** distributed मोड में, इस API के बजाय वॉल्ट के `PUT /admin/clients/{id}` का उपयोग करने की अनुशंसा की जाती है। वॉल्ट में किए गए परिवर्तन SSE के माध्यम से 1-3 सेकंड में स्वचालित रूप से प्रतिबिंबित होते हैं।

---

### `PUT /api/config/think-mode`

थिंकिंग मोड टॉगल (no-op, भविष्य के विस्तार के लिए आरक्षित)।

**प्रतिक्रिया:**
```json
{"status": "ok"}
```

---

### `POST /reload`

वॉल्ट से क्लाइंट कॉन्फ़िगरेशन·कीज़ तुरंत पुन: सिंक करता है।

**प्रतिक्रिया:**
```json
{"status": "reloading"}
```

पुन: सिंक एसिंक्रोनस रूप से निष्पादित होता है, इसलिए प्रतिक्रिया प्राप्त होने के बाद 1-2 सेकंड में पूरा हो जाता है।

---

### `POST /google/v1beta/models/{model}:generateContent`

Gemini API प्रॉक्सी (नॉन-स्ट्रीमिंग)।

**पाथ पैरामीटर:**
- `{model}`: मॉडल ID। `gemini-` प्रीफ़िक्स होने पर स्वचालित रूप से Google सर्विस चुनी जाती है।

**अनुरोध बॉडी:** [Gemini generateContent अनुरोध प्रारूप](https://ai.google.dev/api/generate-content)

```json
{
  "contents": [
    {
      "role": "user",
      "parts": [{"text": "안녕하세요"}]
    }
  ],
  "generationConfig": {
    "temperature": 0.7,
    "maxOutputTokens": 1024
  }
}
```

**प्रतिक्रिया बॉडी:** Gemini generateContent प्रतिक्रिया प्रारूप

**टूल फ़िल्टर:** `tool_filter: strip_all` सेट होने पर अनुरोध का `tools` ऐरे स्वचालित रूप से हटा दिया जाता है।

**फ़ॉलबैक चेन:** निर्दिष्ट सर्विस विफल → कॉन्फ़िगर सर्विस क्रम में फ़ॉलबैक → Ollama (अंतिम)।

---

### `POST /google/v1beta/models/{model}:streamGenerateContent`

Gemini API स्ट्रीमिंग प्रॉक्सी। अनुरोध प्रारूप नॉन-स्ट्रीमिंग के समान। प्रतिक्रिया SSE स्ट्रीम है:

```
data: {"candidates":[{"content":{"parts":[{"text":"안"}],...},...}]}

data: [DONE]
```

---

### `POST /v1/chat/completions`

OpenAI संगत API। आंतरिक रूप से Gemini प्रारूप में रूपांतरित करके प्रोसेस करता है।

**अनुरोध बॉडी:**
```json
{
  "model": "gemini-2.5-flash",
  "messages": [
    {"role": "system", "content": "당신은 도움이 되는 어시스턴트입니다."},
    {"role": "user", "content": "안녕하세요"}
  ],
  "temperature": 0.7,
  "max_tokens": 1024,
  "stream": false
}
```

**`model` फ़ील्ड में प्रोवाइडर प्रीफ़िक्स समर्थन (OpenClaw 3.11+):**

| मॉडल उदाहरण | राउटिंग |
|-------------|---------|
| `gemini-2.5-flash` | वर्तमान कॉन्फ़िगर सर्विस |
| `google/gemini-2.5-pro` | सीधे Google |
| `openai/gpt-4o` | सीधे OpenAI |
| `anthropic/claude-opus-4-6` | OpenRouter के माध्यम से |
| `openrouter/meta-llama/llama-3.3-70b` | सीधे OpenRouter |
| `wall-vault/gemini-2.5-flash` | ऑटो-डिटेक्ट → Google |
| `wall-vault/claude-opus-4-6` | ऑटो-डिटेक्ट → OpenRouter Anthropic |
| `wall-vault/gpt-4o` | ऑटो-डिटेक्ट → OpenAI |
| `wall-vault/hunter-alpha` | OpenRouter (मुफ़्त 1M context) |
| `moonshot/kimi-k2.5` | OpenRouter के माध्यम से |
| `opencode-go/model` | OpenRouter के माध्यम से |
| `kimi-k2.5:cloud` | `:cloud` सफ़िक्स → OpenRouter |

विवरण के लिए [प्रोवाइडर·मॉडल राउटिंग](#प्रोवाइडरमॉडल-राउटिंग) देखें।

**प्रतिक्रिया बॉडी:**
```json
{
  "choices": [
    {
      "message": {
        "role": "assistant",
        "content": "안녕하세요! 무엇을 도와드릴까요?"
      },
      "finish_reason": "stop",
      "index": 0
    }
  ]
}
```

> **मॉडल कंट्रोल टोकन स्वचालित हटाना:** यदि प्रतिक्रिया में GLM-5 / DeepSeek / ChatML डिलिमिटर (`<|im_start|>`, `[gMASK]`, `[sop]` आदि) शामिल हैं, तो वे स्वचालित रूप से हटा दिए जाते हैं।

---

## की वॉल्ट API (:56243)

की वॉल्ट जिस सर्वर पर चलता है। डिफ़ॉल्ट पोर्ट `56243`।

---

### सार्वजनिक API (प्रमाणीकरण अनावश्यक)

#### `GET /`

वेब डैशबोर्ड UI। ब्राउज़र से एक्सेस करें।

---

#### `GET /api/status`

वॉल्ट स्थिति जाँच।

**प्रतिक्रिया उदाहरण:**
```json
{
  "status": "ok",
  "version": "v0.1.6.20260314.231308",
  "keys": 3,
  "clients": 2,
  "sse": 2
}
```

---

#### `GET /api/clients`

पंजीकृत क्लाइंट सूची (केवल सार्वजनिक जानकारी, टोकन शामिल नहीं)।

---

### `GET /api/events`

SSE (Server-Sent Events) रीयल-टाइम इवेंट स्ट्रीम।

**हेडर:**
```
Content-Type: text/event-stream
Cache-Control: no-cache
Connection: keep-alive
```

**कनेक्शन होते ही प्राप्त:**
```
data: {"type":"connected","clients":2}
```

**इवेंट उदाहरण:**
```
data: {"type":"config_change","data":{"client_id":"bot-a","service":"google","model":"gemini-2.5-pro"}}

data: {"type":"key_added","data":{"service":"openrouter"}}

data: {"type":"key_deleted","data":{"service":"google"}}

data: {"type":"service_changed","data":{"action":"updated","id":"ollama"}}

data: {"type":"usage_reset","data":{"time":"2026-03-13T00:00:30Z"}}
```

इवेंट प्रकारों का विवरण [SSE इवेंट प्रकार](#sse-इवेंट-प्रकार) में देखें।

---

### प्रॉक्सी-विशिष्ट API (क्लाइंट टोकन)

`Authorization: Bearer <client_token>` हेडर आवश्यक। एडमिन टोकन से भी प्रमाणीकरण संभव।

#### `GET /api/keys`

प्रॉक्सी को प्रदान की जाने वाली डिक्रिप्ट की गई API कीज़ की सूची।

**क्वेरी पैरामीटर:**

| पैरामीटर | विवरण |
|---------|-------|
| `service` | सर्विस फ़िल्टर (उदा: `?service=google`) |

**प्रतिक्रिया उदाहरण:**
```json
[
  {
    "id": "key-abc123",
    "service": "google",
    "plain_key": "AIzaSy...",
    "daily_limit": 1000,
    "today_usage": 42,
    "today_attempts": 45
  }
]
```

> **सुरक्षा:** प्लेनटेक्स्ट कीज़ लौटाता है। क्लाइंट की `allowed_services` सेटिंग के अनुसार केवल अनुमत सर्विस कीज़ ही लौटाई जाती हैं।

---

#### `GET /api/services`

प्रॉक्सी द्वारा उपयोग की जाने वाली सर्विस सूची। `proxy_enabled=true` वाले सर्विस ID का ऐरे लौटाता है।

**प्रतिक्रिया उदाहरण:**
```json
["google", "ollama"]
```

खाली ऐरे का मतलब प्रॉक्सी बिना प्रतिबंध के सभी सर्विसेज़ का उपयोग कर सकती है।

---

#### `POST /api/heartbeat`

प्रॉक्सी स्थिति भेजना (हर 20 सेकंड स्वचालित रूप से निष्पादित)।

**अनुरोध बॉडी:**
```json
{
  "client_id": "bot-a",
  "version": "v0.1.6.20260314.231308",
  "service": "google",
  "model": "gemini-2.5-flash",
  "sse_connected": true,
  "host": "bot-a-host",
  "avatar": "data:image/png;base64,...",
  "key_usage":     {"key-abc123": 42, "key-def456": 0},
  "key_attempts":  {"key-abc123": 45, "key-def456": 3},
  "key_cooldowns": {"key-abc123": "2026-03-15T14:30:00Z"}
}
```

| फ़ील्ड | प्रकार | विवरण |
|--------|--------|-------|
| `client_id` | string | क्लाइंट ID |
| `version` | string | प्रॉक्सी वर्जन (बिल्ड टाइमस्टैम्प सहित, उदा: `v0.1.6.20260314.231308`) |
| `service` | string | वर्तमान सर्विस |
| `model` | string | वर्तमान मॉडल |
| `sse_connected` | bool | SSE कनेक्शन स्थिति |
| `host` | string | होस्टनेम |
| `avatar` | string | base64 data URI of the proxy's local avatar image (sent by proxy, auto-persisted to client record by vault). Set via `WV_AVATAR` env var (relative path under `~/.openclaw/`). |
| `key_usage` | map[string]int | Key ID → successful tokens used today. Vault calls `SetKeyUsage()` (absolute value sync). 429/402/582 errors do **not** increment this. |
| `key_attempts` | map[string]int | Key ID → total API calls today (success + rate-limited). Vault calls `SetKeyAttempts()`. 429/402/582 errors only increment this, not `key_usage`. |
| `key_cooldowns` | map[string]string | Key ID → cooldown end time (RFC3339). Vault calls `SetKeyCooldownIfLater()`. |

**प्रतिक्रिया:**
```json
{"status": "ok"}
```

---

### एडमिन API — API कीज़

`Authorization: Bearer <admin_token>` हेडर आवश्यक।

#### `GET /admin/keys`

सभी पंजीकृत API कीज़ की सूची (प्लेनटेक्स्ट की शामिल नहीं)।

**प्रतिक्रिया उदाहरण:**
```json
[
  {
    "id": "key-abc123",
    "service": "google",
    "label": "메인 키",
    "today_usage": 42,
    "today_attempts": 45,
    "daily_limit": 1000,
    "cooldown_until": "0001-01-01T00:00:00Z",
    "last_error": 0,
    "created_at": "2026-03-13T12:00:00Z",
    "available": true,
    "usage_pct": 4
  }
]
```

| फ़ील्ड | प्रकार | विवरण |
|--------|--------|-------|
| `today_usage` | int | आज सफल अनुरोध टोकन संख्या (429/402/582 त्रुटियाँ शामिल नहीं) |
| `today_attempts` | int | आज कुल API कॉल संख्या (सफल + rate-limited सहित) |
| `available` | bool | कूलडाउन·सीमा के बिना उपयोग योग्य है या नहीं |
| `usage_pct` | int | दैनिक सीमा के विरुद्ध उपयोग प्रतिशत (`daily_limit=0` होने पर 0) |
| `cooldown_until` | RFC3339 | कूलडाउन समाप्ति समय (शून्य मान का अर्थ कोई नहीं) |
| `last_error` | int | अंतिम HTTP त्रुटि कोड |

---

#### `POST /admin/keys`

नई API की पंजीकरण। पंजीकरण के तुरंत बाद SSE `key_added` इवेंट ब्रॉडकास्ट होता है।

**अनुरोध बॉडी:**
```json
{
  "service": "google",
  "key": "AIzaSy...",
  "label": "메인 키",
  "daily_limit": 1000
}
```

| फ़ील्ड | आवश्यक | विवरण |
|--------|---------|-------|
| `service` | ✅ | `google` \| `openai` \| `anthropic` \| `openrouter` \| `ollama` \| कस्टम |
| `key` | ✅ | API की प्लेनटेक्स्ट |
| `label` | — | पहचान लेबल |
| `daily_limit` | — | दैनिक उपयोग सीमा (0 = असीमित) |

---

#### `DELETE /admin/keys/{id}`

API की हटाना। हटाने के बाद SSE `key_deleted` इवेंट ब्रॉडकास्ट होता है।

**प्रतिक्रिया:**
```json
{"status": "deleted"}
```

---

#### `POST /admin/keys/reset`

सभी कीज़ का दैनिक उपयोग रीसेट। SSE `usage_reset` इवेंट ब्रॉडकास्ट।

**प्रतिक्रिया:**
```json
{
  "status": "reset",
  "time": "2026-03-13T15:00:00Z"
}
```

---

### एडमिन API — क्लाइंट

#### `GET /admin/clients`

सभी क्लाइंट सूची (टोकन सहित)।

---

#### `POST /admin/clients`

नया क्लाइंट पंजीकरण।

**अनुरोध बॉडी:**
```json
{
  "id": "my-bot",
  "name": "내 봇",
  "token": "my-secret-token",
  "default_service": "google",
  "default_model": "gemini-2.5-flash",
  "allowed_services": ["google", "openrouter"],
  "agent_type": "openclaw",
  "work_dir": "~/.openclaw",
  "description": "OpenClaw 에이전트",
  "ip_whitelist": ["10.0.0.1", "10.0.0.0/24"],
  "enabled": true
}
```

| फ़ील्ड | आवश्यक | विवरण |
|--------|---------|-------|
| `id` | ✅ | क्लाइंट अद्वितीय ID |
| `name` | — | प्रदर्शन नाम |
| `token` | — | प्रमाणीकरण टोकन (छोड़ने पर स्वचालित जनरेट) |
| `default_service` | — | डिफ़ॉल्ट सर्विस |
| `default_model` | — | डिफ़ॉल्ट मॉडल |
| `allowed_services` | — | अनुमत सर्विसेज़ की सूची (खाली ऐरे = सभी अनुमत) |
| `agent_type` | — | `openclaw` \| `claude-code` \| `cursor` \| `vscode` \| `custom` |
| `work_dir` | — | एजेंट कार्य निर्देशिका |
| `description` | — | एजेंट विवरण |
| `ip_whitelist` | — | अनुमत IP सूची (खाली ऐरे = सभी अनुमत, CIDR समर्थित) |
| `enabled` | — | सक्रिय है या नहीं (डिफ़ॉल्ट `true`) |

---

#### `GET /admin/clients/{id}`

विशिष्ट क्लाइंट की जाँच (टोकन सहित)।

---

#### `PUT /admin/clients/{id}`

क्लाइंट कॉन्फ़िगरेशन बदलें। **SSE `config_change` ब्रॉडकास्ट → 1-3 सेकंड में प्रॉक्सी पर प्रतिबिंबित।**

**अनुरोध बॉडी (केवल बदले जाने वाले फ़ील्ड):**
```json
{
  "default_service": "openrouter",
  "default_model": "anthropic/claude-opus-4-6",
  "enabled": true
}
```

**प्रतिक्रिया:**
```json
{"status": "updated"}
```

---

#### `DELETE /admin/clients/{id}`

क्लाइंट हटाना।

---

### एडमिन API — सर्विसेज़

#### `GET /admin/services`

पंजीकृत सर्विसेज़ की सूची।

**प्रतिक्रिया उदाहरण:**
```json
[
  {"id": "google",      "name": "Google Gemini",   "enabled": true,  "custom": false},
  {"id": "openai",      "name": "OpenAI",          "enabled": true,  "custom": false},
  {"id": "anthropic",   "name": "Anthropic",       "enabled": false, "custom": false},
  {"id": "openrouter",  "name": "OpenRouter",      "enabled": true,  "custom": false},
  {"id": "ollama",      "name": "Ollama (Local)",  "enabled": true,  "custom": false,
   "local_url": "http://localhost:11434"},
  {"id": "lmstudio",    "name": "LM Studio",       "enabled": false, "custom": false},
  {"id": "vllm",        "name": "vLLM",            "enabled": false, "custom": false},
  {"id": "github-copilot","name":"GitHub Copilot", "enabled": false, "custom": false}
]
```

डिफ़ॉल्ट प्रदान सर्विसेज़ 8: `google`, `openai`, `anthropic`, `openrouter`, `github-copilot`, `ollama`, `lmstudio`, `vllm`

---

#### `POST /admin/services`

कस्टम सर्विस जोड़ना। जोड़ने के बाद SSE `service_changed` इवेंट ब्रॉडकास्ट → **डैशबोर्ड ड्रॉपडाउन तुरंत अपडेट**।

**अनुरोध बॉडी:**
```json
{
  "id": "my-llm",
  "name": "사내 LLM 서버",
  "local_url": "http://10.0.0.50:8080",
  "enabled": true
}
```

---

#### `PUT /admin/services/{id}`

सर्विस कॉन्फ़िगरेशन अपडेट। परिवर्तन के बाद SSE `service_changed` इवेंट ब्रॉडकास्ट।

**अनुरोध बॉडी:**
```json
{
  "local_url": "http://192.168.x.x:11434",
  "enabled": true
}
```

---

#### `DELETE /admin/services/{id}`

कस्टम सर्विस हटाना। हटाने के बाद SSE `service_changed` इवेंट ब्रॉडकास्ट।

डिफ़ॉल्ट सर्विस (`custom: false`) हटाने का प्रयास करने पर:
```json
{"error": "기본 서비스는 삭제할 수 없습니다: google"}
```

---

### एडमिन API — मॉडल सूची

#### `GET /admin/models`

सर्विस के अनुसार मॉडल सूची जाँच। TTL कैश (10 मिनट) का उपयोग।

**क्वेरी पैरामीटर:**

| पैरामीटर | विवरण | उदाहरण |
|---------|-------|--------|
| `service` | सर्विस फ़िल्टर | `?service=google` |
| `q` | मॉडल खोज | `?q=gemini` |

**सर्विस के अनुसार मॉडल जाँच विधि:**

| सर्विस | विधि | संख्या |
|--------|-------|--------|
| `google` | स्थिर सूची | 8 (embedding सहित) |
| `openai` | स्थिर सूची | 9 |
| `anthropic` | स्थिर सूची | 6 |
| `github-copilot` | स्थिर सूची | 6 |
| `openrouter` | API डायनामिक जाँच (विफल होने पर curated फ़ॉलबैक 14) | 340+ |
| `ollama` | लोकल सर्वर डायनामिक जाँच (अनुत्तरदायी होने पर अनुशंसित 7) | परिवर्तनीय |
| `lmstudio` | लोकल सर्वर डायनामिक जाँच | परिवर्तनीय |
| `vllm` | लोकल सर्वर डायनामिक जाँच | परिवर्तनीय |
| कस्टम | OpenAI संगत `/v1/models` | परिवर्तनीय |

**OpenRouter फ़ॉलबैक मॉडल सूची (API अनुत्तरदायी होने पर):**

| मॉडल | विशेष नोट |
|------|----------|
| `openrouter/hunter-alpha` | मुफ़्त, 1M context |
| `openrouter/healer-alpha` | मुफ़्त, omni-modal |
| `moonshot/kimi-k2.5` | 256K context |
| `z-ai/glm-5`, `z-ai/glm-4.7-flash` | — |
| `deepseek/deepseek-r1`, `deepseek/deepseek-chat` | — |
| `qwen/qwen-2.5-72b-instruct` | 131K context |
| `minimax/minimax-m2.5` | — |
| `meta-llama/llama-3.3-70b-instruct` | 131K context |

---

### एडमिन API — प्रॉक्सी स्थिति

#### `GET /admin/proxies`

सभी कनेक्टेड प्रॉक्सी का नवीनतम Heartbeat स्थिति।

---

## SSE इवेंट प्रकार

वॉल्ट `/api/events` स्ट्रीम में प्राप्त इवेंट:

| `type` | ट्रिगर शर्त | `data` सामग्री | डैशबोर्ड प्रतिक्रिया |
|--------|-------------|----------------|---------------------|
| `connected` | SSE कनेक्शन होते ही | `{"clients": N}` | — |
| `config_change` | क्लाइंट कॉन्फ़िगरेशन परिवर्तन | `{"client_id","service","model"}` | एजेंट कार्ड मॉडल ड्रॉपडाउन अपडेट |
| `key_added` | नई API की पंजीकरण | `{"service": "google"}` | मॉडल ड्रॉपडाउन अपडेट |
| `key_deleted` | API की हटाना | `{"service": "google"}` | मॉडल ड्रॉपडाउन अपडेट |
| `service_changed` | सर्विस जोड़ना/संशोधन/हटाना | `{"action":"added"\|"updated"\|"deleted","id":"...","proxy_services":["google","ollama"]}` | सर्विस select + मॉडल ड्रॉपडाउन तुरंत अपडेट; प्रॉक्सी dispatch सर्विस सूची रीयल-टाइम अपडेट |
| `usage_update` | प्रॉक्सी heartbeat प्राप्त होने पर (हर 20 सेकंड) | `{"keys":[{"id","service","today_usage","today_attempts","daily_limit","cooldown_until"},...]}` | की उपयोग बार·संख्या तुरंत अपडेट, कूलडाउन काउंटडाउन शुरू। fetch के बिना SSE डेटा सीधे उपयोग। बार share-of-total स्केलिंग (असीमित कीज़)। |
| `usage_reset` | दैनिक उपयोग रीसेट | `{"time": "RFC3339"}` | पेज रीफ़्रेश |

**प्रॉक्सी द्वारा प्राप्त इवेंट प्रोसेसिंग:**

```
config_change प्राप्त
  → client_id अपने से मेल खाने पर
    → service, model तुरंत अपडेट
    → hooksMgr.Fire(EventModelChanged)
```

---

## प्रोवाइडर·मॉडल राउटिंग

`/v1/chat/completions` के `model` फ़ील्ड में `provider/model` प्रारूप निर्दिष्ट करने पर स्वचालित राउटिंग होती है (OpenClaw 3.11 संगत)।

### प्रीफ़िक्स राउटिंग नियम

| प्रीफ़िक्स | राउटिंग लक्ष्य | उदाहरण |
|-----------|---------------|--------|
| `google/` | सीधे Google | `google/gemini-2.5-pro` |
| `openai/` | सीधे OpenAI | `openai/gpt-4o` |
| `anthropic/` | OpenRouter के माध्यम से | `anthropic/claude-opus-4-6` |
| `ollama/` | सीधे Ollama | `ollama/qwen3.5:35b` |
| `custom/` | पुनरावर्ती पुन: पार्स (`custom/` हटाकर पुन: राउटिंग) | `custom/google/gemini-2.5-flash` → Google |
| `openrouter/` | OpenRouter (bare पाथ बनाए रखें) | `openrouter/meta-llama/llama-3.3-70b` |
| `opencode/`, `opencode-go/`, `opencode-zen/` | OpenRouter | `opencode-go/model` |
| `moonshot/`, `kimi-coding/` | OpenRouter (full path बनाए रखें) | `moonshot/kimi-k2.5` |
| `groq/`, `mistral/`, `minimax/`, `cohere/`, `perplexity/` | OpenRouter | `groq/llama-3.1-70b` |
| `together/`, `huggingface/`, `nvidia/`, `venice/` | OpenRouter | — |
| `meta-llama/`, `qwen/`, `deepseek/`, `01-ai/` | OpenRouter (full path) | `deepseek/deepseek-r1` |

### `wall-vault/` प्रीफ़िक्स ऑटो-डिटेक्शन

wall-vault का अपना प्रीफ़िक्स जो मॉडल ID से सर्विस स्वचालित रूप से पहचानता है।

| मॉडल ID पैटर्न | राउटिंग |
|---------------|---------|
| `wall-vault/gemini-*` | Google |
| `wall-vault/gpt-*`, `wall-vault/o1`, `wall-vault/o3`, `wall-vault/o4*` | OpenAI |
| `wall-vault/claude-*` | OpenRouter (Anthropic पाथ) |
| `wall-vault/hunter-alpha`, `wall-vault/healer-alpha` | OpenRouter (मुफ़्त 1M ctx) |
| `wall-vault/kimi-*`, `wall-vault/glm-*`, `wall-vault/deepseek-*`, `wall-vault/qwen*` | OpenRouter |
| अन्य | OpenRouter |

### `:cloud` सफ़िक्स प्रोसेसिंग

Ollama टैग प्रारूप का `:cloud` सफ़िक्स स्वचालित रूप से हटाकर OpenRouter पर राउट किया जाता है।

```
kimi-k2.5:cloud  →  OpenRouter, मॉडल ID: kimi-k2.5
glm-5:cloud      →  OpenRouter, मॉडल ID: glm-5
```

### OpenClaw openclaw.json एकीकरण उदाहरण

```json5
// ~/.openclaw/openclaw.json
{
  models: {
    providers: {
      "wall-vault": {
        baseUrl: "https://localhost:56244/v1",
        apiKey: "YOUR_AGENT_TOKEN",
        api: "openai-completions",
        models: [
          { id: "wall-vault/gemini-2.5-flash" },
          { id: "wall-vault/hunter-alpha" },
          { id: "wall-vault/claude-opus-4-6" }
        ]
      }
    }
  },
  agents: {
    defaults: {
      model: {
        primary: "wall-vault/gemini-2.5-flash",
        fallbacks: ["wall-vault/hunter-alpha"]
      }
    }
  }
}
```

एजेंट कार्ड पर **🐾 बटन** क्लिक करने पर उस एजेंट के लिए कॉन्फ़िगरेशन स्निपेट स्वचालित रूप से क्लिपबोर्ड पर कॉपी हो जाता है।

---

## डेटा स्कीमा

### APIKey

| फ़ील्ड | प्रकार | विवरण |
|--------|--------|-------|
| `id` | string | UUID प्रारूप अद्वितीय ID |
| `service` | string | `google` \| `openai` \| `anthropic` \| `openrouter` \| `ollama` \| कस्टम |
| `encrypted_key` | string | AES-GCM एन्क्रिप्टेड की (Base64) |
| `label` | string | पहचान लेबल |
| `today_usage` | int | आज सफल अनुरोध टोकन संख्या (429/402/582 त्रुटियाँ शामिल नहीं) |
| `today_attempts` | int | आज कुल API कॉल संख्या (सफल + rate-limited; मध्यरात्रि रीसेट) |
| `daily_limit` | int | दैनिक सीमा (0 = असीमित) |
| `cooldown_until` | time.Time | कूलडाउन समाप्ति समय |
| `last_error` | int | अंतिम HTTP त्रुटि कोड |
| `created_at` | time.Time | पंजीकरण समय |

**कूलडाउन नीति:**

| HTTP त्रुटि | कूलडाउन |
|-------------|---------|
| 429 (Too Many Requests) | 30 मिनट |
| 402 (Payment Required) | 24 घंटे |
| 400 / 401 / 403 | 24 घंटे |
| 582 (Gateway Overload) | 5 मिनट |
| नेटवर्क त्रुटि | 10 मिनट |

> **429·402·582**: कूलडाउन सेट + `today_attempts` बढ़ता है। `today_usage` नहीं बदलता (केवल सफल टोकन गिने जाते हैं)।
> **Ollama (लोकल सर्विस)**: `callOllama` `Timeout: 0` (असीमित) वाला विशेष HTTP क्लाइंट उपयोग करता है। बड़े मॉडल इन्फ़ेरेंस में दसियों सेकंड से कई मिनट लग सकते हैं, इसलिए डिफ़ॉल्ट 60 सेकंड का टाइमआउट लागू नहीं किया जाता।

### Client

| फ़ील्ड | प्रकार | विवरण |
|--------|--------|-------|
| `id` | string | क्लाइंट अद्वितीय ID |
| `name` | string | प्रदर्शन नाम |
| `token` | string | प्रमाणीकरण टोकन |
| `default_service` | string | डिफ़ॉल्ट सर्विस |
| `default_model` | string | डिफ़ॉल्ट मॉडल (`provider/model` प्रारूप संभव) |
| `allowed_services` | []string | अनुमत सर्विसेज़ (खाली ऐरे = सभी) |
| `agent_type` | string | `openclaw` \| `claude-code` \| `cursor` \| `vscode` \| `custom` |
| `work_dir` | string | एजेंट कार्य निर्देशिका |
| `description` | string | विवरण |
| `ip_whitelist` | []string | अनुमत IP सूची (CIDR समर्थित) |
| `avatar` | string | Agent avatar — relative path under `~/.openclaw/` (e.g. `workspace/avatar.png`, `workspace/avatars/profile.hpg`) OR base64 data URI (`data:image/png;base64,...`). Supported extensions: `.png`, `.jpg`/`.jpeg`/`.hpg`, `.webp`, `.gif`. |
| `enabled` | bool | `false` होने पर `/api/keys` एक्सेस करते समय `403` |
| `created_at` | time.Time | पंजीकरण समय |

### ServiceConfig

| फ़ील्ड | प्रकार | विवरण |
|--------|--------|-------|
| `id` | string | सर्विस अद्वितीय ID |
| `name` | string | प्रदर्शन नाम |
| `local_url` | string | लोकल सर्वर URL (Ollama/LMStudio/vLLM/कस्टम) |
| `enabled` | bool | सक्रिय है या नहीं |
| `custom` | bool | उपयोगकर्ता द्वारा जोड़ी गई सर्विस है या नहीं |
| `proxy_enabled` | bool | Whether this service is included in proxy dispatch and agent model dropdowns. Controlled by the "프록시 사용" checkbox in the Services card. Only `proxy_enabled: true` services appear in agent service/model selectors. |

### ProxyStatus (Heartbeat)

| फ़ील्ड | प्रकार | विवरण |
|--------|--------|-------|
| `client_id` | string | क्लाइंट ID |
| `version` | string | प्रॉक्सी वर्जन (उदा: `v0.1.6.20260314.231308`) |
| `service` | string | वर्तमान सर्विस |
| `model` | string | वर्तमान मॉडल |
| `sse_connected` | bool | SSE कनेक्शन स्थिति |
| `host` | string | होस्टनेम |
| `avatar` | string | base64 data URI of local avatar (auto-synced from proxy via heartbeat) |
| `updated_at` | time.Time | अंतिम अपडेट समय |
| `vault.today_usage` | int | आज का टोकन उपयोग |
| `vault.daily_limit` | int | दैनिक सीमा |
| `vault.key_status` | string | `active` \| `cooldown` \| `exhausted` |

---

## त्रुटि प्रतिक्रिया

```json
{"error": "오류 메시지"}
```

| कोड | अर्थ |
|-----|------|
| 200 | सफल |
| 400 | अमान्य अनुरोध |
| 401 | प्रमाणीकरण विफल |
| 403 | एक्सेस अस्वीकृत (निष्क्रिय क्लाइंट, IP ब्लॉक) |
| 404 | संसाधन नहीं मिला |
| 405 | अनुमत नहीं विधि |
| 429 | रेट लिमिट पार |
| 500 | सर्वर आंतरिक त्रुटि |
| 502 | अपस्ट्रीम API त्रुटि (सभी फ़ॉलबैक विफल) |

---

## cURL उदाहरण संग्रह

```bash
# ─── प्रॉक्सी ──────────────────────────────────────────────────────────────────

# हेल्थ चेक
curl https://localhost:56244/health

# स्थिति जाँच
curl https://localhost:56244/status

# मॉडल सूची (सभी)
curl https://localhost:56244/api/models

# केवल Google मॉडल
curl "https://localhost:56244/api/models?service=google"

# मुफ़्त मॉडल खोज
curl "https://localhost:56244/api/models?q=alpha"

# मॉडल बदलें (लोकल)
curl -X PUT https://localhost:56244/api/config/model \
  -H "Content-Type: application/json" \
  -d '{"service":"google","model":"gemini-2.5-pro"}'

# कॉन्फ़िगरेशन रीलोड
curl -X POST https://localhost:56244/reload

# Gemini API सीधे कॉल
curl -X POST "https://localhost:56244/google/v1beta/models/gemini-2.5-flash:generateContent" \
  -H "Content-Type: application/json" \
  -d '{"contents":[{"role":"user","parts":[{"text":"안녕하세요"}]}]}'

# OpenAI संगत (डिफ़ॉल्ट मॉडल)
curl -X POST https://localhost:56244/v1/chat/completions \
  -H "Content-Type: application/json" \
  -d '{"model":"gemini-2.5-flash","messages":[{"role":"user","content":"안녕"}]}'

# OpenClaw provider/model प्रारूप
curl -X POST https://localhost:56244/v1/chat/completions \
  -H "Authorization: Bearer my-agent-token" \
  -H "Content-Type: application/json" \
  -d '{"model":"wall-vault/gemini-2.5-flash","messages":[{"role":"user","content":"안녕"}]}'

# मुफ़्त 1M context मॉडल उपयोग
curl -X POST https://localhost:56244/v1/chat/completions \
  -H "Authorization: Bearer my-agent-token" \
  -H "Content-Type: application/json" \
  -d '{"model":"wall-vault/hunter-alpha","messages":[{"role":"user","content":"안녕"}]}'

# ─── की वॉल्ट (सार्वजनिक) ──────────────────────────────────────────────────────

curl https://localhost:56243/api/status
curl https://localhost:56243/api/clients
curl -s https://localhost:56243/api/events --max-time 3

# ─── की वॉल्ट (एडमिन) ────────────────────────────────────────────────────────

ADMIN="Authorization: Bearer admin-token"

# की सूची
curl -H "$ADMIN" https://localhost:56243/admin/keys

# Google की जोड़ें
curl -X POST https://localhost:56243/admin/keys \
  -H "$ADMIN" -H "Content-Type: application/json" \
  -d '{"service":"google","key":"AIzaSy...","label":"메인 키","daily_limit":1000}'

# OpenAI की जोड़ें
curl -X POST https://localhost:56243/admin/keys \
  -H "$ADMIN" -H "Content-Type: application/json" \
  -d '{"service":"openai","key":"sk-...","label":"GPT 키"}'

# OpenRouter की जोड़ें
curl -X POST https://localhost:56243/admin/keys \
  -H "$ADMIN" -H "Content-Type: application/json" \
  -d '{"service":"openrouter","key":"sk-or-v1-...","label":"OR 키"}'

# की हटाएँ (SSE key_deleted ब्रॉडकास्ट)
curl -X DELETE https://localhost:56243/admin/keys/key-abc123 -H "$ADMIN"

# दैनिक उपयोग रीसेट
curl -X POST https://localhost:56243/admin/keys/reset -H "$ADMIN"

# क्लाइंट सूची
curl -H "$ADMIN" https://localhost:56243/admin/clients

# क्लाइंट जोड़ें (OpenClaw)
curl -X POST https://localhost:56243/admin/clients \
  -H "$ADMIN" -H "Content-Type: application/json" \
  -d '{"id":"bot-a","name":"봇 A","agent_type":"openclaw","work_dir":"~/.openclaw","default_service":"google","default_model":"gemini-2.5-flash"}'

# क्लाइंट मॉडल बदलें (SSE तुरंत प्रतिबिंबन)
curl -X PUT https://localhost:56243/admin/clients/bot-a \
  -H "$ADMIN" -H "Content-Type: application/json" \
  -d '{"default_service":"openrouter","default_model":"wall-vault/hunter-alpha"}'

# क्लाइंट निष्क्रिय करें
curl -X PUT https://localhost:56243/admin/clients/my-bot \
  -H "$ADMIN" -H "Content-Type: application/json" \
  -d '{"enabled":false}'

# क्लाइंट हटाएँ
curl -X DELETE https://localhost:56243/admin/clients/my-bot -H "$ADMIN"

# सर्विस सूची
curl -H "$ADMIN" https://localhost:56243/admin/services

# Ollama लोकल URL सेट करें (SSE service_changed ब्रॉडकास्ट)
curl -X PUT https://localhost:56243/admin/services/ollama \
  -H "$ADMIN" -H "Content-Type: application/json" \
  -d '{"local_url":"http://192.168.x.x:11434","enabled":true}'

# OpenAI सर्विस सक्रिय करें
curl -X PUT https://localhost:56243/admin/services/openai \
  -H "$ADMIN" -H "Content-Type: application/json" \
  -d '{"enabled":true}'

# कस्टम सर्विस जोड़ें (SSE service_changed ब्रॉडकास्ट)
curl -X POST https://localhost:56243/admin/services \
  -H "$ADMIN" -H "Content-Type: application/json" \
  -d '{"id":"my-llm","name":"사내 LLM","local_url":"http://10.0.0.50:8080","enabled":true}'

# कस्टम सर्विस हटाएँ
curl -X DELETE https://localhost:56243/admin/services/my-llm -H "$ADMIN"

# मॉडल सूची जाँच
curl -H "$ADMIN" https://localhost:56243/admin/models
curl -H "$ADMIN" "https://localhost:56243/admin/models?service=openrouter"
curl -H "$ADMIN" "https://localhost:56243/admin/models?q=hunter"

# प्रॉक्सी स्थिति (heartbeat)
curl -H "$ADMIN" https://localhost:56243/admin/proxies

# ─── वितरित मोड — प्रॉक्सी → वॉल्ट ──────────────────────────────────────────

# डिक्रिप्ट की गई कीज़ जाँचें
curl https://localhost:56243/api/keys \
  -H "Authorization: Bearer your-bot-a-token"

# Heartbeat भेजें
curl -X POST https://localhost:56243/api/heartbeat \
  -H "Authorization: Bearer your-bot-a-token" \
  -H "Content-Type: application/json" \
  -d '{"client_id":"bot-a","version":"v0.1.6.20260314.231308","service":"google","model":"gemini-2.5-flash","sse_connected":true}'
```

---

## मिडलवेयर

सभी अनुरोधों पर स्वचालित रूप से लागू:

| मिडलवेयर | कार्य |
|----------|-------|
| **Logger** | `[method] path status latencyms` प्रारूप में लॉगिंग |
| **CORS** | `Access-Control-Allow-Origin: *` |
| **Recovery** | पैनिक रिकवरी, 500 प्रतिक्रिया लौटाना |

---

*अंतिम अपडेट: 2026-03-16 — v0.1.7: dashboard title renamed to "벽금고(wall-vault) 대시보드", logo moved to sticky topbar, today_attempts tracking, HTTP 582 cooldown (5min), share-of-total bar scaling, custom/ routing fix, Ollama Timeout:0, key_att i18n, avatar heartbeat sync, build timestamp versioning, proxy-only service filter*
