# wall-vault API म्यानुअल

यो कागजातले wall-vault का सबै HTTP API एन्डपोइन्टहरूको विस्तृत विवरण गर्छ।

---

## विषयसूची

- [प्रमाणीकरण](#प्रमाणीकरण)
- [प्रोक्सी API (:56244)](#प्रोक्सी-api-56244)
  - [हेल्थचेक](#get-health)
  - [स्थिति जाँच](#get-status)
  - [मोडेल सूची](#get-apimodels)
  - [मोडेल परिवर्तन](#put-apiconfigmodel)
  - [सोच मोड](#put-apiconfigthink-mode)
  - [सेटिङ रिफ्रेस](#post-reload)
  - [Gemini API](#post-googlev1betamodelsmgeneratecontent)
  - [Gemini स्ट्रिमिङ](#post-googlev1betamodelsmstreamgeneratecontent)
  - [OpenAI मिल्दो API](#post-v1chatcompletions)
- [कुञ्जी भण्डार API (:56243)](#कुञ्जी-भण्डार-api-56243)
  - [सार्वजनिक API](#सार्वजनिक-api-प्रमाणीकरण-आवश्यक-छैन)
  - [SSE इभेन्ट स्ट्रिम](#get-apievents)
  - [प्रोक्सी-मात्र API](#प्रोक्सी-मात्र-api-क्लाइन्ट-टोकन)
  - [प्रशासक API — कुञ्जीहरू](#प्रशासक-api--api-कुञ्जी)
  - [प्रशासक API — क्लाइन्टहरू](#प्रशासक-api--क्लाइन्ट)
  - [प्रशासक API — सेवाहरू](#प्रशासक-api--सेवा)
  - [प्रशासक API — मोडेल सूची](#प्रशासक-api--मोडेल-सूची)
  - [प्रशासक API — प्रोक्सी स्थिति](#प्रशासक-api--प्रोक्सी-स्थिति)
- [SSE इभेन्ट प्रकारहरू](#sse-इभेन्ट-प्रकारहरू)
- [प्रोभाइडर·मोडेल राउटिङ](#प्रोभाइडरमोडेल-राउटिङ)
- [डाटा स्किमा](#डाटा-स्किमा)
- [त्रुटि प्रतिक्रिया](#त्रुटि-प्रतिक्रिया)
- [cURL उदाहरण संग्रह](#curl-उदाहरण-संग्रह)

---

## प्रमाणीकरण

| क्षेत्र | विधि | हेडर |
|------|------|------|
| प्रशासक API | Bearer टोकन | `Authorization: Bearer <admin_token>` |
| प्रोक्सी → भण्डार | Bearer टोकन | `Authorization: Bearer <client_token>` |
| प्रोक्सी API | छैन (स्थानीय) | — |

`admin_token` सेट नगरिएको अवस्थामा (खाली स्ट्रिङ) सबै प्रशासक API बिना प्रमाणीकरण पहुँचयोग्य हुन्छन्।

### सुरक्षा नीति

- **Rate Limiting**: प्रशासक API प्रमाणीकरण असफल 10 पटक/15 मिनेट भन्दा बढी हुँदा सम्बन्धित IP अस्थायी रूपमा ब्लक हुन्छ (`429 Too Many Requests`)
- **IP ह्वाइटलिस्ट**: एजेन्ट (`Client`) को `ip_whitelist` फिल्डमा दर्ता गरिएका IP/CIDR मात्र `/api/keys` पहुँच गर्न सक्छन्। खाली एरे भए सबैलाई अनुमति दिइन्छ।
- **theme·lang सुरक्षा**: `/admin/theme`, `/admin/lang` मा पनि प्रशासक टोकन प्रमाणीकरण आवश्यक छ

---

## प्रोक्सी API (:56244)

प्रोक्सी चल्ने सर्भर। पूर्वनिर्धारित पोर्ट `56244`।

---

### `GET /health`

हेल्थचेक। सधैँ 200 OK फर्काउँछ।

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

प्रोक्सी स्थितिको विस्तृत जाँच।

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

| फिल्ड | प्रकार | विवरण |
|------|------|------|
| `service` | string | हालको पूर्वनिर्धारित सेवा |
| `model` | string | हालको पूर्वनिर्धारित मोडेल |
| `sse` | bool | भण्डार SSE जडान स्थिति |
| `filter` | string | उपकरण फिल्टर मोड |
| `services` | []string | सक्रिय सेवा सूची |
| `mode` | string | `standalone` \| `distributed` |

---

### `GET /api/models`

उपलब्ध मोडेलहरूको सूची। TTL क्यास (पूर्वनिर्धारित 10 मिनेट) प्रयोग गर्छ।

**क्वेरी प्यारामिटर:**

| प्यारामिटर | विवरण | उदाहरण |
|---------|------|------|
| `service` | सेवा फिल्टर | `?service=google` |
| `q` | मोडेल ID/नाम खोजी | `?q=gemini` |

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

| फिल्ड | प्रकार | विवरण |
|------|------|------|
| `id` | string | मोडेल ID |
| `name` | string | मोडेल प्रदर्शन नाम |
| `service` | string | `google` \| `openai` \| `anthropic` \| `openrouter` \| `ollama` आदि |
| `context_length` | int | सन्दर्भ विन्डो आकार |
| `free` | bool | निःशुल्क मोडेल हो वा होइन (OpenRouter) |

---

### `PUT /api/config/model`

हालको सेवा·मोडेल परिवर्तन।

**अनुरोध बडी:**
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

> **नोट:** distributed मोडमा यो API को सट्टा भण्डारको `PUT /admin/clients/{id}` प्रयोग गर्नु सिफारिस गरिन्छ। भण्डारमा गरिएको परिवर्तन SSE मार्फत 1–3 सेकेन्डमा स्वचालित रूपमा प्रतिबिम्बित हुन्छ।

---

### `PUT /api/config/think-mode`

सोच मोड टगल (no-op, भविष्य विस्तारका लागि)।

**प्रतिक्रिया:**
```json
{"status": "ok"}
```

---

### `POST /reload`

भण्डारबाट क्लाइन्ट सेटिङ·कुञ्जीहरू तुरुन्तै पुन:सिंक गर्छ।

**प्रतिक्रिया:**
```json
{"status": "reloading"}
```

पुन:सिंक एसिन्क्रोनस रूपमा चल्छ, त्यसैले प्रतिक्रिया प्राप्त भएपछि 1–2 सेकेन्डमा पूरा हुन्छ।

---

### `POST /google/v1beta/models/{model}:generateContent`

Gemini API प्रोक्सी (गैर-स्ट्रिमिङ)।

**पथ प्यारामिटर:**
- `{model}`: मोडेल ID। `gemini-` प्रिफिक्स भएमा स्वचालित रूपमा Google सेवा छनोट हुन्छ।

**अनुरोध बडी:** [Gemini generateContent अनुरोध ढाँचा](https://ai.google.dev/api/generate-content)

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

**प्रतिक्रिया बडी:** Gemini generateContent प्रतिक्रिया ढाँचा

**उपकरण फिल्टर:** `tool_filter: strip_all` सेटिङ गरिएमा अनुरोधको `tools` एरे स्वचालित रूपमा हटाइन्छ।

**फलब्याक चेन:** निर्दिष्ट सेवा असफल → सेट गरिएको सेवा क्रमानुसार फलब्याक → Ollama (अन्तिम)।

---

### `POST /google/v1beta/models/{model}:streamGenerateContent`

Gemini API स्ट्रिमिङ प्रोक्सी। अनुरोध ढाँचा गैर-स्ट्रिमिङसँग उस्तै। प्रतिक्रिया SSE स्ट्रिममा आउँछ:

```
data: {"candidates":[{"content":{"parts":[{"text":"안"}],...},...}]}

data: [DONE]
```

---

### `POST /v1/chat/completions`

OpenAI मिल्दो API। आन्तरिक रूपमा Gemini ढाँचामा रूपान्तरण गरेर प्रशोधन गर्छ।

**अनुरोध बडी:**
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

**`model` फिल्डमा प्रोभाइडर प्रिफिक्स समर्थन (OpenClaw 3.11+):**

| मोडेल उदाहरण | राउटिङ |
|-----------|--------|
| `gemini-2.5-flash` | हालको सेटिङ सेवा |
| `google/gemini-2.5-pro` | Google प्रत्यक्ष |
| `openai/gpt-4o` | OpenAI प्रत्यक्ष |
| `anthropic/claude-opus-4-6` | OpenRouter मार्फत |
| `openrouter/meta-llama/llama-3.3-70b` | OpenRouter प्रत्यक्ष |
| `wall-vault/gemini-2.5-flash` | स्वचालित पत्ता लगाउने → Google |
| `wall-vault/claude-opus-4-6` | स्वचालित पत्ता लगाउने → OpenRouter Anthropic |
| `wall-vault/gpt-4o` | स्वचालित पत्ता लगाउने → OpenAI |
| `wall-vault/hunter-alpha` | OpenRouter (निःशुल्क 1M context) |
| `moonshot/kimi-k2.5` | OpenRouter मार्फत |
| `opencode-go/model` | OpenRouter मार्फत |
| `kimi-k2.5:cloud` | `:cloud` प्रत्यय → OpenRouter |

विस्तृत जानकारीको लागि [प्रोभाइडर·मोडेल राउटिङ](#प्रोभाइडरमोडेल-राउटिङ) हेर्नुहोस्।

**प्रतिक्रिया बडी:**
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

> **मोडेल नियन्त्रण टोकन स्वचालित हटाउने:** प्रतिक्रियामा GLM-5 / DeepSeek / ChatML विभाजक (`<|im_start|>`, `[gMASK]`, `[sop]` आदि) समावेश भएमा स्वचालित रूपमा हटाइन्छ।

---

## कुञ्जी भण्डार API (:56243)

कुञ्जी भण्डार चल्ने सर्भर। पूर्वनिर्धारित पोर्ट `56243`।

---

### सार्वजनिक API (प्रमाणीकरण आवश्यक छैन)

#### `GET /`

वेब ड्यासबोर्ड UI। ब्राउजरबाट पहुँच गर्नुहोस्।

---

#### `GET /api/status`

भण्डार स्थिति जाँच।

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

दर्ता गरिएका क्लाइन्ट सूची (सार्वजनिक जानकारी मात्र, टोकन बाहेक)।

---

### `GET /api/events`

SSE (Server-Sent Events) रियल-टाइम इभेन्ट स्ट्रिम।

**हेडर:**
```
Content-Type: text/event-stream
Cache-Control: no-cache
Connection: keep-alive
```

**जडान हुनासाथ प्राप्त:**
```
data: {"type":"connected","clients":2}
```

**इभेन्ट उदाहरण:**
```
data: {"type":"config_change","data":{"client_id":"bot-a","service":"google","model":"gemini-2.5-pro"}}

data: {"type":"key_added","data":{"service":"openrouter"}}

data: {"type":"key_deleted","data":{"service":"google"}}

data: {"type":"service_changed","data":{"action":"updated","id":"ollama"}}

data: {"type":"usage_reset","data":{"time":"2026-03-13T00:00:30Z"}}
```

विस्तृत इभेन्ट प्रकारहरूको लागि [SSE इभेन्ट प्रकारहरू](#sse-इभेन्ट-प्रकारहरू) हेर्नुहोस्।

---

### प्रोक्सी-मात्र API (क्लाइन्ट टोकन)

`Authorization: Bearer <client_token>` हेडर आवश्यक। प्रशासक टोकनबाट पनि प्रमाणीकरण सम्भव छ।

#### `GET /api/keys`

प्रोक्सीलाई प्रदान गरिने डिक्रिप्ट गरिएको API कुञ्जी सूची।

**क्वेरी प्यारामिटर:**

| प्यारामिटर | विवरण |
|---------|------|
| `service` | सेवा फिल्टर (उदा: `?service=google`) |

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

> **सुरक्षा:** प्लेनटेक्स्ट कुञ्जी फर्काउँछ। क्लाइन्टको `allowed_services` सेटिङ अनुसार अनुमति दिइएका सेवा कुञ्जीहरू मात्र फर्काइन्छ।

---

#### `GET /api/services`

प्रोक्सीले प्रयोग गर्ने सेवा सूची। `proxy_enabled=true` भएका सेवा ID एरे फर्काउँछ।

**प्रतिक्रिया उदाहरण:**
```json
["google", "ollama"]
```

खाली एरे भए प्रोक्सीले प्रतिबन्ध बिना सबै सेवा प्रयोग गर्छ।

---

#### `POST /api/heartbeat`

प्रोक्सी स्थिति पठाउने (हरेक 20 सेकेन्डमा स्वचालित रूपमा चल्छ)।

**अनुरोध बडी:**
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

| फिल्ड | प्रकार | विवरण |
|------|------|------|
| `client_id` | string | क्लाइन्ट ID |
| `version` | string | प्रोक्सी संस्करण (build timestamp सहित, उदा. `v0.1.6.20260314.231308`) |
| `service` | string | हालको सेवा |
| `model` | string | हालको मोडेल |
| `sse_connected` | bool | SSE जडान स्थिति |
| `host` | string | होस्टनाम |
| `avatar` | string | base64 data URI of the proxy's local avatar image (sent by proxy, auto-persisted to client record by vault). Set via `WV_AVATAR` env var (relative path under `~/.openclaw/`). |
| `key_usage` | map[string]int | Key ID → successful tokens used today. Vault calls `SetKeyUsage()` (absolute value sync). 429/402/582 errors do **not** increment this. |
| `key_attempts` | map[string]int | Key ID → total API calls today (success + rate-limited). Vault calls `SetKeyAttempts()`. 429/402/582 errors only increment this, not `key_usage`. |
| `key_cooldowns` | map[string]string | Key ID → cooldown end time (RFC3339). Vault calls `SetKeyCooldownIfLater()`. |

**प्रतिक्रिया:**
```json
{"status": "ok"}
```

---

### प्रशासक API — API कुञ्जी

`Authorization: Bearer <admin_token>` हेडर आवश्यक।

#### `GET /admin/keys`

दर्ता गरिएका सबै API कुञ्जी सूची (प्लेनटेक्स्ट कुञ्जी बाहेक)।

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

| फिल्ड | प्रकार | विवरण |
|------|------|------|
| `today_usage` | int | आज सफल अनुरोध टोकन संख्या (429/402/582 त्रुटि समावेश छैन) |
| `today_attempts` | int | आज कुल API कल संख्या (सफल + rate-limited सहित) |
| `available` | bool | कूलडाउन·सीमा बिना प्रयोगयोग्य छ वा छैन |
| `usage_pct` | int | दैनिक सीमा अनुपातमा प्रयोग % (`daily_limit=0` भए 0) |
| `cooldown_until` | RFC3339 | कूलडाउन समाप्ति समय (शून्य मान भए छैन) |
| `last_error` | int | अन्तिम HTTP त्रुटि कोड |

---

#### `POST /admin/keys`

नयाँ API कुञ्जी दर्ता। दर्ता हुनासाथ SSE `key_added` इभेन्ट ब्रोडकास्ट हुन्छ।

**अनुरोध बडी:**
```json
{
  "service": "google",
  "key": "AIzaSy...",
  "label": "메인 키",
  "daily_limit": 1000
}
```

| फिल्ड | आवश्यक | विवरण |
|------|------|------|
| `service` | ✅ | `google` \| `openai` \| `anthropic` \| `openrouter` \| `ollama` \| कस्टम |
| `key` | ✅ | API कुञ्जी प्लेनटेक्स्ट |
| `label` | — | पहिचान लेबल |
| `daily_limit` | — | दैनिक प्रयोग सीमा (0 = असीमित) |

---

#### `DELETE /admin/keys/{id}`

API कुञ्जी मेटाउने। मेटाएपछि SSE `key_deleted` इभेन्ट ब्रोडकास्ट हुन्छ।

**प्रतिक्रिया:**
```json
{"status": "deleted"}
```

---

#### `POST /admin/keys/reset`

सबै कुञ्जीहरूको दैनिक प्रयोग मात्रा रिसेट। SSE `usage_reset` इभेन्ट ब्रोडकास्ट।

**प्रतिक्रिया:**
```json
{
  "status": "reset",
  "time": "2026-03-13T15:00:00Z"
}
```

---

### प्रशासक API — क्लाइन्ट

#### `GET /admin/clients`

सबै क्लाइन्ट सूची (टोकन सहित)।

---

#### `POST /admin/clients`

नयाँ क्लाइन्ट दर्ता।

**अनुरोध बडी:**
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

| फिल्ड | आवश्यक | विवरण |
|------|------|------|
| `id` | ✅ | क्लाइन्ट अद्वितीय ID |
| `name` | — | प्रदर्शन नाम |
| `token` | — | प्रमाणीकरण टोकन (छोडेमा स्वचालित उत्पन्न) |
| `default_service` | — | पूर्वनिर्धारित सेवा |
| `default_model` | — | पूर्वनिर्धारित मोडेल |
| `allowed_services` | — | अनुमति दिइएको सेवा सूची (खाली एरे = सबै अनुमति) |
| `agent_type` | — | `openclaw` \| `claude-code` \| `cursor` \| `vscode` \| `custom` |
| `work_dir` | — | एजेन्ट कार्य डाइरेक्टरी |
| `description` | — | एजेन्ट विवरण |
| `ip_whitelist` | — | अनुमति IP सूची (खाली एरे = सबै अनुमति, CIDR समर्थित) |
| `enabled` | — | सक्रिय स्थिति (पूर्वनिर्धारित `true`) |

---

#### `GET /admin/clients/{id}`

विशिष्ट क्लाइन्ट जाँच (टोकन सहित)।

---

#### `PUT /admin/clients/{id}`

क्लाइन्ट सेटिङ परिवर्तन। **SSE `config_change` ब्रोडकास्ट → प्रोक्सीमा 1–3 सेकेन्डमा प्रतिबिम्बित।**

**अनुरोध बडी (परिवर्तन गर्ने फिल्ड मात्र):**
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

क्लाइन्ट मेटाउने।

---

### प्रशासक API — सेवा

#### `GET /admin/services`

दर्ता गरिएको सेवा सूची।

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

पूर्वनिर्धारित सेवा 8 वटा: `google`, `openai`, `anthropic`, `openrouter`, `github-copilot`, `ollama`, `lmstudio`, `vllm`

---

#### `POST /admin/services`

कस्टम सेवा थप्ने। थपेपछि SSE `service_changed` इभेन्ट ब्रोडकास्ट → **ड्यासबोर्ड ड्रपडाउन तुरुन्तै अपडेट**।

**अनुरोध बडी:**
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

सेवा सेटिङ अपडेट। परिवर्तनपछि SSE `service_changed` इभेन्ट ब्रोडकास्ट।

**अनुरोध बडी:**
```json
{
  "local_url": "http://192.168.x.x:11434",
  "enabled": true
}
```

---

#### `DELETE /admin/services/{id}`

कस्टम सेवा मेटाउने। मेटाएपछि SSE `service_changed` इभेन्ट ब्रोडकास्ट।

पूर्वनिर्धारित सेवा (`custom: false`) मेटाउने प्रयास:
```json
{"error": "기본 서비스는 삭제할 수 없습니다: google"}
```

---

### प्रशासक API — मोडेल सूची

#### `GET /admin/models`

सेवा अनुसार मोडेल सूची। TTL क्यास (10 मिनेट) प्रयोग।

**क्वेरी प्यारामिटर:**

| प्यारामिटर | विवरण | उदाहरण |
|---------|------|------|
| `service` | सेवा फिल्टर | `?service=google` |
| `q` | मोडेल खोजी | `?q=gemini` |

**सेवा अनुसार मोडेल प्राप्ति विधि:**

| सेवा | विधि | संख्या |
|--------|------|------|
| `google` | स्थिर सूची | 8 वटा (embedding सहित) |
| `openai` | स्थिर सूची | 9 वटा |
| `anthropic` | स्थिर सूची | 6 वटा |
| `github-copilot` | स्थिर सूची | 6 वटा |
| `openrouter` | API गतिशील क्वेरी (असफल भए curated फलब्याक 14 वटा) | 340+ वटा |
| `ollama` | स्थानीय सर्भर गतिशील क्वेरी (प्रतिक्रिया नभए सिफारिस 7 वटा) | परिवर्तनशील |
| `lmstudio` | स्थानीय सर्भर गतिशील क्वेरी | परिवर्तनशील |
| `vllm` | स्थानीय सर्भर गतिशील क्वेरी | परिवर्तनशील |
| कस्टम | OpenAI मिल्दो `/v1/models` | परिवर्तनशील |

**OpenRouter फलब्याक मोडेल सूची (API प्रतिक्रिया नभएमा):**

| मोडेल | विशेष नोट |
|------|----------|
| `openrouter/hunter-alpha` | निःशुल्क, 1M context |
| `openrouter/healer-alpha` | निःशुल्क, omni-modal |
| `moonshot/kimi-k2.5` | 256K context |
| `z-ai/glm-5`, `z-ai/glm-4.7-flash` | — |
| `deepseek/deepseek-r1`, `deepseek/deepseek-chat` | — |
| `qwen/qwen-2.5-72b-instruct` | 131K context |
| `minimax/minimax-m2.5` | — |
| `meta-llama/llama-3.3-70b-instruct` | 131K context |

---

### प्रशासक API — प्रोक्सी स्थिति

#### `GET /admin/proxies`

जडान गरिएका सबै प्रोक्सीको अन्तिम Heartbeat स्थिति।

---

## SSE इभेन्ट प्रकारहरू

भण्डार `/api/events` स्ट्रिमबाट प्राप्त हुने इभेन्टहरू:

| `type` | उत्पन्न हुने अवस्था | `data` सामग्री | ड्यासबोर्ड प्रतिक्रिया |
|--------|-----------|-------------|--------------|
| `connected` | SSE जडान हुनासाथ | `{"clients": N}` | — |
| `config_change` | क्लाइन्ट सेटिङ परिवर्तन | `{"client_id","service","model"}` | एजेन्ट कार्ड मोडेल ड्रपडाउन अपडेट |
| `key_added` | नयाँ API कुञ्जी दर्ता | `{"service": "google"}` | मोडेल ड्रपडाउन अपडेट |
| `key_deleted` | API कुञ्जी मेटाइयो | `{"service": "google"}` | मोडेल ड्रपडाउन अपडेट |
| `service_changed` | सेवा थप/सम्पादन/मेटाउने | `{"action":"added"\|"updated"\|"deleted","id":"...","proxy_services":["google","ollama"]}` | सेवा select + मोडेल ड्रपडाउन तुरुन्तै अपडेट; प्रोक्सीको dispatch सेवा सूची रियल-टाइम अपडेट |
| `usage_update` | प्रोक्सी heartbeat प्राप्त हुँदा (हरेक 20 सेकेन्ड) | `{"keys":[{"id","service","today_usage","today_attempts","daily_limit","cooldown_until"},...]}` | कुञ्जी प्रयोग बार·अंक तुरुन्तै अपडेट, कूलडाउन काउन्टडाउन सुरु। fetch बिना SSE डाटा प्रत्यक्ष प्रयोग। बार share-of-total स्केलिङ (असीमित कुञ्जी)। |
| `usage_reset` | दैनिक प्रयोग मात्रा रिसेट | `{"time": "RFC3339"}` | पृष्ठ रिफ्रेस |

**प्रोक्सीले प्राप्त गर्ने इभेन्ट प्रशोधन:**

```
config_change प्राप्त
  → client_id आफ्नोसँग मिल्ने अवस्थामा
    → service, model तुरुन्तै अपडेट
    → hooksMgr.Fire(EventModelChanged)
```

---

## प्रोभाइडर·मोडेल राउटिङ

`/v1/chat/completions` को `model` फिल्डमा `provider/model` ढाँचा निर्दिष्ट गर्दा स्वचालित राउटिङ हुन्छ (OpenClaw 3.11 मिल्दो)।

### प्रिफिक्स राउटिङ नियमहरू

| प्रिफिक्स | राउटिङ गन्तव्य | उदाहरण |
|--------|------------|------|
| `google/` | Google प्रत्यक्ष | `google/gemini-2.5-pro` |
| `openai/` | OpenAI प्रत्यक्ष | `openai/gpt-4o` |
| `anthropic/` | OpenRouter मार्फत | `anthropic/claude-opus-4-6` |
| `ollama/` | Ollama प्रत्यक्ष | `ollama/qwen3.5:35b` |
| `custom/` | पुनरावर्ती पुन:पार्सिङ (`custom/` हटाएर पुन:राउटिङ) | `custom/google/gemini-2.5-flash` → Google |
| `openrouter/` | OpenRouter (bare पथ कायम) | `openrouter/meta-llama/llama-3.3-70b` |
| `opencode/`, `opencode-go/`, `opencode-zen/` | OpenRouter | `opencode-go/model` |
| `moonshot/`, `kimi-coding/` | OpenRouter (full path कायम) | `moonshot/kimi-k2.5` |
| `groq/`, `mistral/`, `minimax/`, `cohere/`, `perplexity/` | OpenRouter | `groq/llama-3.1-70b` |
| `together/`, `huggingface/`, `nvidia/`, `venice/` | OpenRouter | — |
| `meta-llama/`, `qwen/`, `deepseek/`, `01-ai/` | OpenRouter (full path) | `deepseek/deepseek-r1` |

### `wall-vault/` प्रिफिक्स स्वचालित पत्ता लगाउने

wall-vault आफ्नै प्रिफिक्सबाट मोडेल ID मा सेवा स्वचालित रूपमा पहिचान गर्छ।

| मोडेल ID ढाँचा | राउटिङ |
|-------------|--------|
| `wall-vault/gemini-*` | Google |
| `wall-vault/gpt-*`, `wall-vault/o1`, `wall-vault/o3`, `wall-vault/o4*` | OpenAI |
| `wall-vault/claude-*` | OpenRouter (Anthropic पथ) |
| `wall-vault/hunter-alpha`, `wall-vault/healer-alpha` | OpenRouter (निःशुल्क 1M ctx) |
| `wall-vault/kimi-*`, `wall-vault/glm-*`, `wall-vault/deepseek-*`, `wall-vault/qwen*` | OpenRouter |
| अन्य | OpenRouter |

### `:cloud` प्रत्यय प्रशोधन

Ollama ट्याग ढाँचाको `:cloud` प्रत्यय स्वचालित रूपमा हटाएर OpenRouter मा राउट गरिन्छ।

```
kimi-k2.5:cloud  →  OpenRouter, मोडेल ID: kimi-k2.5
glm-5:cloud      →  OpenRouter, मोडेल ID: glm-5
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

एजेन्ट कार्डको **🐾 बटन** क्लिक गर्दा सम्बन्धित एजेन्टको सेटिङ स्निपेट क्लिपबोर्डमा स्वचालित रूपमा प्रतिलिपि हुन्छ।

---

## डाटा स्किमा

### APIKey

| फिल्ड | प्रकार | विवरण |
|------|------|------|
| `id` | string | UUID ढाँचाको अद्वितीय ID |
| `service` | string | `google` \| `openai` \| `anthropic` \| `openrouter` \| `ollama` \| कस्टम |
| `encrypted_key` | string | AES-GCM इन्क्रिप्ट गरिएको कुञ्जी (Base64) |
| `label` | string | पहिचान लेबल |
| `today_usage` | int | आज सफल अनुरोध टोकन संख्या (429/402/582 त्रुटि समावेश छैन) |
| `today_attempts` | int | आज कुल API कल संख्या (सफल + rate-limited; मध्यरात रिसेट) |
| `daily_limit` | int | दैनिक सीमा (0 = असीमित) |
| `cooldown_until` | time.Time | कूलडाउन समाप्ति समय |
| `last_error` | int | अन्तिम HTTP त्रुटि कोड |
| `created_at` | time.Time | दर्ता समय |

**कूलडाउन नीति:**

| HTTP त्रुटि | कूलडाउन |
|-----------|--------|
| 429 (Too Many Requests) | 30 मिनेट |
| 402 (Payment Required) | 24 घण्टा |
| 400 / 401 / 403 | 24 घण्टा |
| 582 (Gateway Overload) | 5 मिनेट |
| नेटवर्क त्रुटि | 10 मिनेट |

> **429·402·582**: कूलडाउन सेट + `today_attempts` वृद्धि। `today_usage` परिवर्तन हुँदैन (सफल टोकन मात्र गणना)।
> **Ollama (स्थानीय सेवा)**: `callOllama` ले `Timeout: 0` (असीमित) समर्पित HTTP क्लाइन्ट प्रयोग गर्छ। ठूला मोडेल अनुमान दशौं सेकेन्ड देखि केही मिनेट लाग्न सक्छ, त्यसैले पूर्वनिर्धारित 60 सेकेन्ड टाइमआउट लागू हुँदैन।

### Client

| फिल्ड | प्रकार | विवरण |
|------|------|------|
| `id` | string | क्लाइन्ट अद्वितीय ID |
| `name` | string | प्रदर्शन नाम |
| `token` | string | प्रमाणीकरण टोकन |
| `default_service` | string | पूर्वनिर्धारित सेवा |
| `default_model` | string | पूर्वनिर्धारित मोडेल (`provider/model` ढाँचा सम्भव) |
| `allowed_services` | []string | अनुमति सेवा (खाली एरे = सबै) |
| `agent_type` | string | `openclaw` \| `claude-code` \| `cursor` \| `vscode` \| `custom` |
| `work_dir` | string | एजेन्ट कार्य डाइरेक्टरी |
| `description` | string | विवरण |
| `ip_whitelist` | []string | अनुमति IP सूची (CIDR समर्थित) |
| `avatar` | string | Agent avatar — relative path under `~/.openclaw/` (e.g. `workspace/avatar.png`, `workspace/avatars/profile.hpg`) OR base64 data URI (`data:image/png;base64,...`). Supported extensions: `.png`, `.jpg`/`.jpeg`/`.hpg`, `.webp`, `.gif`. |
| `enabled` | bool | `false` भए `/api/keys` पहुँचमा `403` |
| `created_at` | time.Time | दर्ता समय |

### ServiceConfig

| फिल्ड | प्रकार | विवरण |
|------|------|------|
| `id` | string | सेवा अद्वितीय ID |
| `name` | string | प्रदर्शन नाम |
| `local_url` | string | स्थानीय सर्भर URL (Ollama/LMStudio/vLLM/कस्टम) |
| `enabled` | bool | सक्रिय स्थिति |
| `custom` | bool | प्रयोगकर्ताले थपेको सेवा हो वा होइन |
| `proxy_enabled` | bool | Whether this service is included in proxy dispatch and agent model dropdowns. Controlled by the "프록시 사용" checkbox in the Services card. Only `proxy_enabled: true` services appear in agent service/model selectors. |

### ProxyStatus (Heartbeat)

| फिल्ड | प्रकार | विवरण |
|------|------|------|
| `client_id` | string | क्लाइन्ट ID |
| `version` | string | प्रोक्सी संस्करण (उदा. `v0.1.6.20260314.231308`) |
| `service` | string | हालको सेवा |
| `model` | string | हालको मोडेल |
| `sse_connected` | bool | SSE जडान स्थिति |
| `host` | string | होस्टनाम |
| `avatar` | string | base64 data URI of local avatar (auto-synced from proxy via heartbeat) |
| `updated_at` | time.Time | अन्तिम अपडेट |
| `vault.today_usage` | int | आजको टोकन प्रयोग |
| `vault.daily_limit` | int | दैनिक सीमा |
| `vault.key_status` | string | `active` \| `cooldown` \| `exhausted` |

---

## त्रुटि प्रतिक्रिया

```json
{"error": "오류 메시지"}
```

| कोड | अर्थ |
|------|------|
| 200 | सफल |
| 400 | गलत अनुरोध |
| 401 | प्रमाणीकरण असफल |
| 403 | पहुँच अस्वीकृत (निष्क्रिय क्लाइन्ट, IP ब्लक) |
| 404 | संसाधन फेला परेन |
| 405 | अनुमति नभएको विधि |
| 429 | Rate limit अतिक्रमण |
| 500 | सर्भर आन्तरिक त्रुटि |
| 502 | अपस्ट्रिम API त्रुटि (सबै फलब्याक असफल) |

---

## cURL उदाहरण संग्रह

```bash
# ─── प्रोक्सी ───────────────────────────────────────────────────────────────────

# हेल्थचेक
curl https://localhost:56244/health

# स्थिति जाँच
curl https://localhost:56244/status

# मोडेल सूची (सबै)
curl https://localhost:56244/api/models

# Google मोडेल मात्र
curl "https://localhost:56244/api/models?service=google"

# निःशुल्क मोडेल खोजी
curl "https://localhost:56244/api/models?q=alpha"

# मोडेल परिवर्तन (स्थानीय)
curl -X PUT https://localhost:56244/api/config/model \
  -H "Content-Type: application/json" \
  -d '{"service":"google","model":"gemini-2.5-pro"}'

# सेटिङ रिफ्रेस
curl -X POST https://localhost:56244/reload

# Gemini API प्रत्यक्ष कल
curl -X POST "https://localhost:56244/google/v1beta/models/gemini-2.5-flash:generateContent" \
  -H "Content-Type: application/json" \
  -d '{"contents":[{"role":"user","parts":[{"text":"안녕하세요"}]}]}'

# OpenAI मिल्दो (पूर्वनिर्धारित मोडेल)
curl -X POST https://localhost:56244/v1/chat/completions \
  -H "Content-Type: application/json" \
  -d '{"model":"gemini-2.5-flash","messages":[{"role":"user","content":"안녕"}]}'

# OpenClaw provider/model ढाँचा
curl -X POST https://localhost:56244/v1/chat/completions \
  -H "Authorization: Bearer my-agent-token" \
  -H "Content-Type: application/json" \
  -d '{"model":"wall-vault/gemini-2.5-flash","messages":[{"role":"user","content":"안녕"}]}'

# निःशुल्क 1M context मोडेल प्रयोग
curl -X POST https://localhost:56244/v1/chat/completions \
  -H "Authorization: Bearer my-agent-token" \
  -H "Content-Type: application/json" \
  -d '{"model":"wall-vault/hunter-alpha","messages":[{"role":"user","content":"안녕"}]}'

# ─── कुञ्जी भण्डार (सार्वजनिक) ───────────────────────────────────────────────────────────

curl https://localhost:56243/api/status
curl https://localhost:56243/api/clients
curl -s https://localhost:56243/api/events --max-time 3

# ─── कुञ्जी भण्डार (प्रशासक) ─────────────────────────────────────────────────────────

ADMIN="Authorization: Bearer admin-token"

# कुञ्जी सूची
curl -H "$ADMIN" https://localhost:56243/admin/keys

# Google कुञ्जी थप्ने
curl -X POST https://localhost:56243/admin/keys \
  -H "$ADMIN" -H "Content-Type: application/json" \
  -d '{"service":"google","key":"AIzaSy...","label":"메인 키","daily_limit":1000}'

# OpenAI कुञ्जी थप्ने
curl -X POST https://localhost:56243/admin/keys \
  -H "$ADMIN" -H "Content-Type: application/json" \
  -d '{"service":"openai","key":"sk-...","label":"GPT 키"}'

# OpenRouter कुञ्जी थप्ने
curl -X POST https://localhost:56243/admin/keys \
  -H "$ADMIN" -H "Content-Type: application/json" \
  -d '{"service":"openrouter","key":"sk-or-v1-...","label":"OR 키"}'

# कुञ्जी मेटाउने (SSE key_deleted ब्रोडकास्ट)
curl -X DELETE https://localhost:56243/admin/keys/key-abc123 -H "$ADMIN"

# दैनिक प्रयोग मात्रा रिसेट
curl -X POST https://localhost:56243/admin/keys/reset -H "$ADMIN"

# क्लाइन्ट सूची
curl -H "$ADMIN" https://localhost:56243/admin/clients

# क्लाइन्ट थप्ने (OpenClaw)
curl -X POST https://localhost:56243/admin/clients \
  -H "$ADMIN" -H "Content-Type: application/json" \
  -d '{"id":"bot-a","name":"봇 A","agent_type":"openclaw","work_dir":"~/.openclaw","default_service":"google","default_model":"gemini-2.5-flash"}'

# क्लाइन्ट मोडेल परिवर्तन (SSE तुरुन्तै प्रतिबिम्बित)
curl -X PUT https://localhost:56243/admin/clients/bot-a \
  -H "$ADMIN" -H "Content-Type: application/json" \
  -d '{"default_service":"openrouter","default_model":"wall-vault/hunter-alpha"}'

# क्लाइन्ट निष्क्रिय गर्ने
curl -X PUT https://localhost:56243/admin/clients/my-bot \
  -H "$ADMIN" -H "Content-Type: application/json" \
  -d '{"enabled":false}'

# क्लाइन्ट मेटाउने
curl -X DELETE https://localhost:56243/admin/clients/my-bot -H "$ADMIN"

# सेवा सूची
curl -H "$ADMIN" https://localhost:56243/admin/services

# Ollama स्थानीय URL सेट गर्ने (SSE service_changed ब्रोडकास्ट)
curl -X PUT https://localhost:56243/admin/services/ollama \
  -H "$ADMIN" -H "Content-Type: application/json" \
  -d '{"local_url":"http://192.168.x.x:11434","enabled":true}'

# OpenAI सेवा सक्रिय गर्ने
curl -X PUT https://localhost:56243/admin/services/openai \
  -H "$ADMIN" -H "Content-Type: application/json" \
  -d '{"enabled":true}'

# कस्टम सेवा थप्ने (SSE service_changed ब्रोडकास्ट)
curl -X POST https://localhost:56243/admin/services \
  -H "$ADMIN" -H "Content-Type: application/json" \
  -d '{"id":"my-llm","name":"사내 LLM","local_url":"http://10.0.0.50:8080","enabled":true}'

# कस्टम सेवा मेटाउने
curl -X DELETE https://localhost:56243/admin/services/my-llm -H "$ADMIN"

# मोडेल सूची जाँच
curl -H "$ADMIN" https://localhost:56243/admin/models
curl -H "$ADMIN" "https://localhost:56243/admin/models?service=openrouter"
curl -H "$ADMIN" "https://localhost:56243/admin/models?q=hunter"

# प्रोक्सी स्थिति (heartbeat)
curl -H "$ADMIN" https://localhost:56243/admin/proxies

# ─── वितरित मोड — प्रोक्सी → भण्डार ───────────────────────────────────────────────

# डिक्रिप्ट गरिएको कुञ्जी जाँच
curl https://localhost:56243/api/keys \
  -H "Authorization: Bearer your-bot-a-token"

# Heartbeat पठाउने
curl -X POST https://localhost:56243/api/heartbeat \
  -H "Authorization: Bearer your-bot-a-token" \
  -H "Content-Type: application/json" \
  -d '{"client_id":"bot-a","version":"v0.1.6.20260314.231308","service":"google","model":"gemini-2.5-flash","sse_connected":true}'
```

---

## मिडलवेयर

सबै अनुरोधमा स्वचालित रूपमा लागू हुन्छ:

| मिडलवेयर | कार्य |
|---------|------|
| **Logger** | `[method] path status latencyms` ढाँचामा लगिङ |
| **CORS** | `Access-Control-Allow-Origin: *` |
| **Recovery** | प्यानिक रिकभरी, 500 प्रतिक्रिया फर्काउने |

---

*अन्तिम अपडेट: 2026-03-16 — v0.1.7: dashboard title renamed to "벽금고(wall-vault) 대시보드", logo moved to sticky topbar, today_attempts tracking, HTTP 582 cooldown (5min), share-of-total bar scaling, custom/ routing fix, Ollama Timeout:0, key_att i18n, avatar heartbeat sync, build timestamp versioning, proxy-only service filter*
