# Mwongozo wa API ya wall-vault

Hati hii inaelezea kwa undani vituo vyote vya API za HTTP vya wall-vault.

---

## Yaliyomo

- [Uthibitishaji](#uthibitishaji)
- [API ya Proksi (:56244)](#api-ya-proksi-56244)
  - [Ukaguzi wa afya](#get-health)
  - [Hali ya kina](#get-status)
  - [Orodha ya modeli](#get-apimodels)
  - [Kubadilisha modeli](#put-apiconfigmodel)
  - [Hali ya mawazo](#put-apiconfigthink-mode)
  - [Kusasisha mipangilio](#post-reload)
  - [Gemini API](#post-googlev1betamodelsmgeneratecontent)
  - [Gemini utiririshaji](#post-googlev1betamodelsmstreamgeneratecontent)
  - [API inayolingana na OpenAI](#post-v1chatcompletions)
- [API ya Hifadhi ya Funguo (:56243)](#api-ya-hifadhi-ya-funguo-56243)
  - [API ya umma](#api-ya-umma-hakuna-uthibitishaji-unaohitajika)
  - [Mtiririko wa matukio ya SSE](#get-apievents)
  - [API ya proksi pekee](#api-ya-proksi-pekee-tokeni-ya-mteja)
  - [API ya msimamizi — Funguo](#api-ya-msimamizi--funguo-za-api)
  - [API ya msimamizi — Wateja](#api-ya-msimamizi--wateja)
  - [API ya msimamizi — Huduma](#api-ya-msimamizi--huduma)
  - [API ya msimamizi — Orodha ya modeli](#api-ya-msimamizi--orodha-ya-modeli)
  - [API ya msimamizi — Hali ya proksi](#api-ya-msimamizi--hali-ya-proksi)
- [Aina za matukio ya SSE](#aina-za-matukio-ya-sse)
- [Uelekezaji wa mtoa huduma·modeli](#uelekezaji-wa-mtoa-hudumamodeli)
- [Muundo wa data](#muundo-wa-data)
- [Majibu ya makosa](#majibu-ya-makosa)
- [Mkusanyiko wa mifano ya cURL](#mkusanyiko-wa-mifano-ya-curl)

---

## Uthibitishaji

| Eneo | Njia | Kichwa |
|------|------|------|
| API ya msimamizi | Tokeni ya Bearer | `Authorization: Bearer <admin_token>` |
| Proksi → Hifadhi | Tokeni ya Bearer | `Authorization: Bearer <client_token>` |
| API ya proksi | Hakuna (ndani) | — |

Ikiwa `admin_token` haijawekwa (mfuatano tupu) API zote za msimamizi zinaweza kufikiwa bila uthibitishaji.

### Sera ya usalama

- **Kizuizi cha Kiwango**: Ikiwa uthibitishaji wa API ya msimamizi unashindwa zaidi ya mara 10/dakika 15, IP hiyo itazuiwa kwa muda (`429 Too Many Requests`)
- **Orodha ya IP Zilizoidhinishwa**: IP/CIDR zilizosajiliwa katika sehemu ya `ip_whitelist` ya wakala (`Client`) pekee ndizo zinazoweza kufikia `/api/keys`. Ikiwa safu ni tupu, wote wanaruhusiwa.
- **Ulinzi wa theme·lang**: `/admin/theme`, `/admin/lang` pia zinahitaji uthibitishaji wa tokeni ya msimamizi

---

## API ya Proksi (:56244)

Seva ambayo proksi inaendesha. Bandari chaguo-msingi `56244`.

---

### `GET /health`

Ukaguzi wa afya. Daima inarudisha 200 OK.

**Mfano wa jibu:**
```json
{
  "status": "ok",
  "version": "v0.1.6.20260314.231308",
  "client": "bot-a"
}
```

---

### `GET /status`

Uchunguzi wa kina wa hali ya proksi.

**Mfano wa jibu:**
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

| Sehemu | Aina | Maelezo |
|------|------|------|
| `service` | string | Huduma ya sasa chaguo-msingi |
| `model` | string | Modeli ya sasa chaguo-msingi |
| `sse` | bool | Hali ya muunganisho wa SSE na hifadhi |
| `filter` | string | Hali ya kichujio cha zana |
| `services` | []string | Orodha ya huduma zinazofanya kazi |
| `mode` | string | `standalone` \| `distributed` |

---

### `GET /api/models`

Orodha ya modeli zinazopatikana. Inatumia akiba ya TTL (chaguo-msingi dakika 10).

**Vigezo vya hoja:**

| Kigezo | Maelezo | Mfano |
|---------|------|------|
| `service` | Kichujio cha huduma | `?service=google` |
| `q` | Utafutaji wa ID/jina la modeli | `?q=gemini` |

**Mfano wa jibu:**
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

| Sehemu | Aina | Maelezo |
|------|------|------|
| `id` | string | ID ya modeli |
| `name` | string | Jina la kuonyesha la modeli |
| `service` | string | `google` \| `openai` \| `anthropic` \| `openrouter` \| `ollama` na kadhalika |
| `context_length` | int | Ukubwa wa dirisha la muktadha |
| `free` | bool | Iwapo modeli ni bure (OpenRouter) |

---

### `PUT /api/config/model`

Kubadilisha huduma·modeli ya sasa.

**Mwili wa ombi:**
```json
{
  "service": "google",
  "model": "gemini-2.5-flash"
}
```

**Jibu:**
```json
{
  "status": "ok",
  "service": "google",
  "model": "gemini-2.5-flash"
}
```

> **Kumbuka:** Katika hali ya distributed, inashauriwa kutumia `PUT /admin/clients/{id}` ya hifadhi badala ya API hii. Mabadiliko yaliyofanywa kwenye hifadhi yanaonyeshwa kiotomatiki kupitia SSE ndani ya sekunde 1–3.

---

### `PUT /api/config/think-mode`

Kubadilisha hali ya mawazo (no-op, kwa upanuzi wa baadaye).

**Jibu:**
```json
{"status": "ok"}
```

---

### `POST /reload`

Kusawazisha upya mipangilio ya mteja·funguo kutoka hifadhi mara moja.

**Jibu:**
```json
{"status": "reloading"}
```

Usawazishaji unaendesha kwa njia isiyo ya moja kwa moja, kwa hivyo unakamilika ndani ya sekunde 1–2 baada ya kupokea jibu.

---

### `POST /google/v1beta/models/{model}:generateContent`

Proksi ya Gemini API (bila utiririshaji).

**Kigezo cha njia:**
- `{model}`: ID ya modeli. Ikiwa ina kiambishi awali cha `gemini-` huduma ya Google inachaguliwa kiotomatiki.

**Mwili wa ombi:** [Muundo wa ombi la Gemini generateContent](https://ai.google.dev/api/generate-content)

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

**Mwili wa jibu:** Muundo wa jibu la Gemini generateContent

**Kichujio cha zana:** Ikiwa `tool_filter: strip_all` imewekwa, safu ya `tools` ya ombi inaondolewa kiotomatiki.

**Mnyororo wa kurudisha:** Huduma iliyoainishwa inashindwa → kurudisha kwa mpangilio wa huduma zilizowekwa → Ollama (ya mwisho).

---

### `POST /google/v1beta/models/{model}:streamGenerateContent`

Proksi ya utiririshaji wa Gemini API. Muundo wa ombi ni sawa na ule usio wa utiririshaji. Jibu linakuja kama mtiririko wa SSE:

```
data: {"candidates":[{"content":{"parts":[{"text":"안"}],...},...}]}

data: [DONE]
```

---

### `POST /v1/chat/completions`

API inayolingana na OpenAI. Inabadilishwa ndani hadi muundo wa Gemini kabla ya kuchakatwa.

**Mwili wa ombi:**
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

**Usaidizi wa kiambishi awali cha mtoa huduma katika sehemu ya `model` (OpenClaw 3.11+):**

| Mfano wa modeli | Uelekezaji |
|-----------|--------|
| `gemini-2.5-flash` | Huduma ya mipangilio ya sasa |
| `google/gemini-2.5-pro` | Google moja kwa moja |
| `openai/gpt-4o` | OpenAI moja kwa moja |
| `anthropic/claude-opus-4-6` | Kupitia OpenRouter |
| `openrouter/meta-llama/llama-3.3-70b` | OpenRouter moja kwa moja |
| `wall-vault/gemini-2.5-flash` | Kugundua kiotomatiki → Google |
| `wall-vault/claude-opus-4-6` | Kugundua kiotomatiki → OpenRouter Anthropic |
| `wall-vault/gpt-4o` | Kugundua kiotomatiki → OpenAI |
| `wall-vault/hunter-alpha` | OpenRouter (bure 1M context) |
| `moonshot/kimi-k2.5` | Kupitia OpenRouter |
| `opencode-go/model` | Kupitia OpenRouter |
| `kimi-k2.5:cloud` | Kiambishi tamati `:cloud` → OpenRouter |

Kwa maelezo zaidi tazama [Uelekezaji wa mtoa huduma·modeli](#uelekezaji-wa-mtoa-hudumamodeli).

**Mwili wa jibu:**
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

> **Kuondoa kiotomatiki tokeni za udhibiti wa modeli:** Ikiwa jibu lina vitenganishi vya GLM-5 / DeepSeek / ChatML (`<|im_start|>`, `[gMASK]`, `[sop]` na kadhalika) vinaondolewa kiotomatiki.

---

## API ya Hifadhi ya Funguo (:56243)

Seva ambayo hifadhi ya funguo inaendesha. Bandari chaguo-msingi `56243`.

---

### API ya umma (hakuna uthibitishaji unaohitajika)

#### `GET /`

UI ya dashibodi ya wavuti. Fikia kupitia kivinjari.

---

#### `GET /api/status`

Ukaguzi wa hali ya hifadhi.

**Mfano wa jibu:**
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

Orodha ya wateja waliosajiliwa (taarifa za umma pekee, bila tokeni).

---

### `GET /api/events`

Mtiririko wa matukio ya SSE (Server-Sent Events) kwa wakati halisi.

**Vichwa:**
```
Content-Type: text/event-stream
Cache-Control: no-cache
Connection: keep-alive
```

**Unapokea mara baada ya kuunganishwa:**
```
data: {"type":"connected","clients":2}
```

**Mfano wa matukio:**
```
data: {"type":"config_change","data":{"client_id":"bot-a","service":"google","model":"gemini-2.5-pro"}}

data: {"type":"key_added","data":{"service":"openrouter"}}

data: {"type":"key_deleted","data":{"service":"google"}}

data: {"type":"service_changed","data":{"action":"updated","id":"ollama"}}

data: {"type":"usage_reset","data":{"time":"2026-03-13T00:00:30Z"}}
```

Kwa aina kamili za matukio tazama [Aina za matukio ya SSE](#aina-za-matukio-ya-sse).

---

### API ya proksi pekee (tokeni ya mteja)

Kichwa cha `Authorization: Bearer <client_token>` kinahitajika. Uthibitishaji kupitia tokeni ya msimamizi pia unawezekana.

#### `GET /api/keys`

Orodha ya funguo za API zilizosimbuliwa ambazo zinatolewa kwa proksi.

**Vigezo vya hoja:**

| Kigezo | Maelezo |
|---------|------|
| `service` | Kichujio cha huduma (mfano: `?service=google`) |

**Mfano wa jibu:**
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

> **Usalama:** Inarudisha funguo kwa maandishi wazi. Funguo za huduma zilizoidhinishwa pekee ndizo zinarudishwa kulingana na mipangilio ya `allowed_services` ya mteja.

---

#### `GET /api/services`

Orodha ya huduma ambazo proksi inatumia. Inarudisha safu ya ID za huduma ambapo `proxy_enabled=true`.

**Mfano wa jibu:**
```json
["google", "ollama"]
```

Ikiwa safu ni tupu, proksi inatumia huduma zote bila vikwazo.

---

#### `POST /api/heartbeat`

Kutuma hali ya proksi (inaendeshwa kiotomatiki kila sekunde 20).

**Mwili wa ombi:**
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

| Sehemu | Aina | Maelezo |
|------|------|------|
| `client_id` | string | ID ya mteja |
| `version` | string | Toleo la proksi (ikiwa na muhuri wa wakati wa ujenzi, mfano `v0.1.6.20260314.231308`) |
| `service` | string | Huduma ya sasa |
| `model` | string | Modeli ya sasa |
| `sse_connected` | bool | Hali ya muunganisho wa SSE |
| `host` | string | Jina la mwenyeji |
| `avatar` | string | base64 data URI of the proxy's local avatar image (sent by proxy, auto-persisted to client record by vault). Set via `WV_AVATAR` env var (relative path under `~/.openclaw/`). |
| `key_usage` | map[string]int | Key ID → successful tokens used today. Vault calls `SetKeyUsage()` (absolute value sync). 429/402/582 errors do **not** increment this. |
| `key_attempts` | map[string]int | Key ID → total API calls today (success + rate-limited). Vault calls `SetKeyAttempts()`. 429/402/582 errors only increment this, not `key_usage`. |
| `key_cooldowns` | map[string]string | Key ID → cooldown end time (RFC3339). Vault calls `SetKeyCooldownIfLater()`. |

**Jibu:**
```json
{"status": "ok"}
```

---

### API ya msimamizi — Funguo za API

Kichwa cha `Authorization: Bearer <admin_token>` kinahitajika.

#### `GET /admin/keys`

Orodha ya funguo zote za API zilizosajiliwa (bila funguo za maandishi wazi).

**Mfano wa jibu:**
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

| Sehemu | Aina | Maelezo |
|------|------|------|
| `today_usage` | int | Idadi ya tokeni za ombi zilizofanikiwa leo (makosa ya 429/402/582 hayajumuishwi) |
| `today_attempts` | int | Jumla ya simu za API leo (zilizofanikiwa + zilizozuiliwa na kiwango) |
| `available` | bool | Inapatikana kutumika bila kupoa·kikomo |
| `usage_pct` | int | Asilimia ya matumizi dhidi ya kikomo cha kila siku (`daily_limit=0` ni 0) |
| `cooldown_until` | RFC3339 | Wakati wa mwisho wa kupoa (ikiwa thamani sifuri hakuna) |
| `last_error` | int | Msimbo wa kosa wa HTTP wa mwisho |

---

#### `POST /admin/keys`

Kusajili ufunguo mpya wa API. Mara baada ya kusajili tukio la SSE `key_added` linatangazwa.

**Mwili wa ombi:**
```json
{
  "service": "google",
  "key": "AIzaSy...",
  "label": "메인 키",
  "daily_limit": 1000
}
```

| Sehemu | Inahitajika | Maelezo |
|------|------|------|
| `service` | ✅ | `google` \| `openai` \| `anthropic` \| `openrouter` \| `ollama` \| maalum |
| `key` | ✅ | Ufunguo wa API kwa maandishi wazi |
| `label` | — | Lebo ya kutambulisha |
| `daily_limit` | — | Kikomo cha matumizi kwa siku (0 = bila kikomo) |

---

#### `DELETE /admin/keys/{id}`

Kufuta ufunguo wa API. Baada ya kufuta tukio la SSE `key_deleted` linatangazwa.

**Jibu:**
```json
{"status": "deleted"}
```

---

#### `POST /admin/keys/reset`

Kuseti upya kiasi cha matumizi ya kila siku cha funguo zote. Tukio la SSE `usage_reset` linatangazwa.

**Jibu:**
```json
{
  "status": "reset",
  "time": "2026-03-13T15:00:00Z"
}
```

---

### API ya msimamizi — Wateja

#### `GET /admin/clients`

Orodha ya wateja wote (ikiwa na tokeni).

---

#### `POST /admin/clients`

Kusajili mteja mpya.

**Mwili wa ombi:**
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

| Sehemu | Inahitajika | Maelezo |
|------|------|------|
| `id` | ✅ | ID ya kipekee ya mteja |
| `name` | — | Jina la kuonyesha |
| `token` | — | Tokeni ya uthibitishaji (ikiachiwa inazalishwa kiotomatiki) |
| `default_service` | — | Huduma chaguo-msingi |
| `default_model` | — | Modeli chaguo-msingi |
| `allowed_services` | — | Orodha ya huduma zilizoidhinishwa (safu tupu = zote) |
| `agent_type` | — | `openclaw` \| `claude-code` \| `cursor` \| `vscode` \| `custom` |
| `work_dir` | — | Saraka ya kazi ya wakala |
| `description` | — | Maelezo ya wakala |
| `ip_whitelist` | — | Orodha ya IP zilizoidhinishwa (safu tupu = zote, CIDR inasaidiwa) |
| `enabled` | — | Hali ya kuwezesha (chaguo-msingi `true`) |

---

#### `GET /admin/clients/{id}`

Kuchunguza mteja mahususi (ikiwa na tokeni).

---

#### `PUT /admin/clients/{id}`

Kubadilisha mipangilio ya mteja. **Tukio la SSE `config_change` linatangazwa → linaonyeshwa kwenye proksi ndani ya sekunde 1–3.**

**Mwili wa ombi (sehemu za kubadilisha pekee):**
```json
{
  "default_service": "openrouter",
  "default_model": "anthropic/claude-opus-4-6",
  "enabled": true
}
```

**Jibu:**
```json
{"status": "updated"}
```

---

#### `DELETE /admin/clients/{id}`

Kufuta mteja.

---

### API ya msimamizi — Huduma

#### `GET /admin/services`

Orodha ya huduma zilizosajiliwa.

**Mfano wa jibu:**
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

Huduma 8 zinazotolewa kwa chaguo-msingi: `google`, `openai`, `anthropic`, `openrouter`, `github-copilot`, `ollama`, `lmstudio`, `vllm`

---

#### `POST /admin/services`

Kuongeza huduma maalum. Baada ya kuongeza tukio la SSE `service_changed` linatangazwa → **orodha ya kushuka ya dashibodi inasasishwa mara moja**.

**Mwili wa ombi:**
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

Kusasisha mipangilio ya huduma. Baada ya mabadiliko tukio la SSE `service_changed` linatangazwa.

**Mwili wa ombi:**
```json
{
  "local_url": "http://192.168.x.x:11434",
  "enabled": true
}
```

---

#### `DELETE /admin/services/{id}`

Kufuta huduma maalum. Baada ya kufuta tukio la SSE `service_changed` linatangazwa.

Jaribio la kufuta huduma ya chaguo-msingi (`custom: false`):
```json
{"error": "기본 서비스는 삭제할 수 없습니다: google"}
```

---

### API ya msimamizi — Orodha ya modeli

#### `GET /admin/models`

Orodha ya modeli kwa kila huduma. Inatumia akiba ya TTL (dakika 10).

**Vigezo vya hoja:**

| Kigezo | Maelezo | Mfano |
|---------|------|------|
| `service` | Kichujio cha huduma | `?service=google` |
| `q` | Utafutaji wa modeli | `?q=gemini` |

**Njia ya kupata modeli kwa kila huduma:**

| Huduma | Njia | Idadi |
|--------|------|------|
| `google` | Orodha tuli | 8 (ikiwa na embedding) |
| `openai` | Orodha tuli | 9 |
| `anthropic` | Orodha tuli | 6 |
| `github-copilot` | Orodha tuli | 6 |
| `openrouter` | Hoja ya API yenye nguvu (ikishindwa curated fallback 14) | 340+ |
| `ollama` | Hoja ya seva ya ndani yenye nguvu (isipojibu mapendekezo 7) | inabadilika |
| `lmstudio` | Hoja ya seva ya ndani yenye nguvu | inabadilika |
| `vllm` | Hoja ya seva ya ndani yenye nguvu | inabadilika |
| maalum | OpenAI inayolingana `/v1/models` | inabadilika |

**Orodha ya modeli za kurudisha za OpenRouter (API isipojibu):**

| Modeli | Maelezo maalum |
|------|----------|
| `openrouter/hunter-alpha` | bure, 1M context |
| `openrouter/healer-alpha` | bure, omni-modal |
| `moonshot/kimi-k2.5` | 256K context |
| `z-ai/glm-5`, `z-ai/glm-4.7-flash` | — |
| `deepseek/deepseek-r1`, `deepseek/deepseek-chat` | — |
| `qwen/qwen-2.5-72b-instruct` | 131K context |
| `minimax/minimax-m2.5` | — |
| `meta-llama/llama-3.3-70b-instruct` | 131K context |

---

### API ya msimamizi — Hali ya proksi

#### `GET /admin/proxies`

Hali ya Heartbeat ya mwisho ya proksi zote zilizounganishwa.

---

## Aina za matukio ya SSE

Matukio yanayopokelewa kutoka mtiririko wa `/api/events` wa hifadhi:

| `type` | Hali ya kutokea | Yaliyomo ya `data` | Mwitikio wa dashibodi |
|--------|-----------|-------------|--------------|
| `connected` | Mara baada ya muunganisho wa SSE | `{"clients": N}` | — |
| `config_change` | Mabadiliko ya mipangilio ya mteja | `{"client_id","service","model"}` | Orodha ya kushuka ya modeli ya kadi ya wakala inasasishwa |
| `key_added` | Usajili mpya wa ufunguo wa API | `{"service": "google"}` | Orodha ya kushuka ya modeli inasasishwa |
| `key_deleted` | Ufunguo wa API umefutwa | `{"service": "google"}` | Orodha ya kushuka ya modeli inasasishwa |
| `service_changed` | Kuongeza/kuhariri/kufuta huduma | `{"action":"added"\|"updated"\|"deleted","id":"...","proxy_services":["google","ollama"]}` | Kuchagua huduma + orodha ya kushuka ya modeli inasasishwa mara moja; orodha ya huduma za kutuma ya proksi inasasishwa kwa wakati halisi |
| `usage_update` | Heartbeat ya proksi inapopokelewa (kila sekunde 20) | `{"keys":[{"id","service","today_usage","today_attempts","daily_limit","cooldown_until"},...]}` | Mwamba·nambari ya matumizi ya ufunguo inasasishwa mara moja, kuhesabu kupoa kunaanza. Data ya SSE inatumika moja kwa moja bila fetch. Mwamba unatumia uwiano wa sehemu-ya-jumla (funguo zisizo na kikomo). |
| `usage_reset` | Kiasi cha matumizi ya kila siku kimeseti upya | `{"time": "RFC3339"}` | Ukurasa unasasishwa |

**Uchakataji wa matukio ambayo proksi inapokea:**

```
config_change imepokelewa
  → ikiwa client_id inalingana na yake
    → service, model inasasishwa mara moja
    → hooksMgr.Fire(EventModelChanged)
```

---

## Uelekezaji wa mtoa huduma·modeli

Ukiainisha muundo wa `provider/model` katika sehemu ya `model` ya `/v1/chat/completions` uelekezaji otomatiki unafanywa (inaendana na OpenClaw 3.11).

### Kanuni za uelekezaji wa kiambishi awali

| Kiambishi awali | Lengo la uelekezaji | Mfano |
|--------|------------|------|
| `google/` | Google moja kwa moja | `google/gemini-2.5-pro` |
| `openai/` | OpenAI moja kwa moja | `openai/gpt-4o` |
| `anthropic/` | Kupitia OpenRouter | `anthropic/claude-opus-4-6` |
| `ollama/` | Ollama moja kwa moja | `ollama/qwen3.5:35b` |
| `custom/` | Kuchambua upya kwa kujirudia (ondoa `custom/` kisha elekezisha upya) | `custom/google/gemini-2.5-flash` → Google |
| `openrouter/` | OpenRouter (njia wazi imehifadhiwa) | `openrouter/meta-llama/llama-3.3-70b` |
| `opencode/`, `opencode-go/`, `opencode-zen/` | OpenRouter | `opencode-go/model` |
| `moonshot/`, `kimi-coding/` | OpenRouter (njia kamili imehifadhiwa) | `moonshot/kimi-k2.5` |
| `groq/`, `mistral/`, `minimax/`, `cohere/`, `perplexity/` | OpenRouter | `groq/llama-3.1-70b` |
| `together/`, `huggingface/`, `nvidia/`, `venice/` | OpenRouter | — |
| `meta-llama/`, `qwen/`, `deepseek/`, `01-ai/` | OpenRouter (njia kamili) | `deepseek/deepseek-r1` |

### Kugundua kiotomatiki kwa kiambishi awali cha `wall-vault/`

wall-vault inatambua huduma kiotomatiki kutoka ID ya modeli kupitia kiambishi chake.

| Muundo wa ID ya modeli | Uelekezaji |
|-------------|--------|
| `wall-vault/gemini-*` | Google |
| `wall-vault/gpt-*`, `wall-vault/o1`, `wall-vault/o3`, `wall-vault/o4*` | OpenAI |
| `wall-vault/claude-*` | OpenRouter (njia ya Anthropic) |
| `wall-vault/hunter-alpha`, `wall-vault/healer-alpha` | OpenRouter (bure 1M ctx) |
| `wall-vault/kimi-*`, `wall-vault/glm-*`, `wall-vault/deepseek-*`, `wall-vault/qwen*` | OpenRouter |
| Nyingine | OpenRouter |

### Uchakataji wa kiambishi tamati `:cloud`

Kiambishi tamati `:cloud` cha muundo wa lebo wa Ollama kinaondolewa kiotomatiki kisha kuelekezwa kwa OpenRouter.

```
kimi-k2.5:cloud  →  OpenRouter, ID ya modeli: kimi-k2.5
glm-5:cloud      →  OpenRouter, ID ya modeli: glm-5
```

### Mfano wa muunganisho wa OpenClaw openclaw.json

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

Kubofya kitufe cha **🐾** kwenye kadi ya wakala kunakili kiotomatiki snippet ya mipangilio ya wakala huyo kwenye ubao wa kunakili.

---

## Muundo wa data

### APIKey

| Sehemu | Aina | Maelezo |
|------|------|------|
| `id` | string | ID ya kipekee ya muundo wa UUID |
| `service` | string | `google` \| `openai` \| `anthropic` \| `openrouter` \| `ollama` \| maalum |
| `encrypted_key` | string | Ufunguo uliofichwa kwa AES-GCM (Base64) |
| `label` | string | Lebo ya kutambulisha |
| `today_usage` | int | Idadi ya tokeni za ombi zilizofanikiwa leo (makosa ya 429/402/582 hayajumuishwi) |
| `today_attempts` | int | Jumla ya simu za API leo (zilizofanikiwa + zilizozuiliwa na kiwango; inaseti upya usiku wa manane) |
| `daily_limit` | int | Kikomo cha kila siku (0 = bila kikomo) |
| `cooldown_until` | time.Time | Wakati wa mwisho wa kupoa |
| `last_error` | int | Msimbo wa kosa wa HTTP wa mwisho |
| `created_at` | time.Time | Wakati wa usajili |

**Sera ya kupoa:**

| Kosa la HTTP | Kupoa |
|-----------|--------|
| 429 (Too Many Requests) | Dakika 30 |
| 402 (Payment Required) | Saa 24 |
| 400 / 401 / 403 | Saa 24 |
| 582 (Gateway Overload) | Dakika 5 |
| Kosa la mtandao | Dakika 10 |

> **429·402·582**: Kupoa kunawekwa + `today_attempts` inaongezeka. `today_usage` haibadiliki (tokeni zilizofanikiwa pekee ndizo zinazohesabiwa).
> **Ollama (huduma ya ndani)**: `callOllama` inatumia mteja wa HTTP uliotengwa `Timeout: 0` (bila kikomo). Makisio ya modeli kubwa yanaweza kuchukua sekunde kadhaa hadi dakika kadhaa, kwa hivyo muda wa kuisha wa sekunde 60 chaguo-msingi hautekelezwi.

### Client

| Sehemu | Aina | Maelezo |
|------|------|------|
| `id` | string | ID ya kipekee ya mteja |
| `name` | string | Jina la kuonyesha |
| `token` | string | Tokeni ya uthibitishaji |
| `default_service` | string | Huduma chaguo-msingi |
| `default_model` | string | Modeli chaguo-msingi (muundo wa `provider/model` unawezekana) |
| `allowed_services` | []string | Huduma zilizoidhinishwa (safu tupu = zote) |
| `agent_type` | string | `openclaw` \| `claude-code` \| `cursor` \| `vscode` \| `custom` |
| `work_dir` | string | Saraka ya kazi ya wakala |
| `description` | string | Maelezo |
| `ip_whitelist` | []string | Orodha ya IP zilizoidhinishwa (CIDR inasaidiwa) |
| `avatar` | string | Agent avatar — relative path under `~/.openclaw/` (e.g. `workspace/avatar.png`, `workspace/avatars/profile.hpg`) OR base64 data URI (`data:image/png;base64,...`). Supported extensions: `.png`, `.jpg`/`.jpeg`/`.hpg`, `.webp`, `.gif`. |
| `enabled` | bool | Ikiwa `false` kufikia `/api/keys` kunarudisha `403` |
| `created_at` | time.Time | Wakati wa usajili |

### ServiceConfig

| Sehemu | Aina | Maelezo |
|------|------|------|
| `id` | string | ID ya kipekee ya huduma |
| `name` | string | Jina la kuonyesha |
| `local_url` | string | URL ya seva ya ndani (Ollama/LMStudio/vLLM/maalum) |
| `enabled` | bool | Hali ya kuwezesha |
| `custom` | bool | Iwapo ni huduma iliyoongezwa na mtumiaji |
| `proxy_enabled` | bool | Whether this service is included in proxy dispatch and agent model dropdowns. Controlled by the "프록시 사용" checkbox in the Services card. Only `proxy_enabled: true` services appear in agent service/model selectors. |

### ProxyStatus (Heartbeat)

| Sehemu | Aina | Maelezo |
|------|------|------|
| `client_id` | string | ID ya mteja |
| `version` | string | Toleo la proksi (mfano `v0.1.6.20260314.231308`) |
| `service` | string | Huduma ya sasa |
| `model` | string | Modeli ya sasa |
| `sse_connected` | bool | Hali ya muunganisho wa SSE |
| `host` | string | Jina la mwenyeji |
| `avatar` | string | base64 data URI of local avatar (auto-synced from proxy via heartbeat) |
| `updated_at` | time.Time | Sasisho la mwisho |
| `vault.today_usage` | int | Matumizi ya tokeni ya leo |
| `vault.daily_limit` | int | Kikomo cha kila siku |
| `vault.key_status` | string | `active` \| `cooldown` \| `exhausted` |

---

## Majibu ya makosa

```json
{"error": "오류 메시지"}
```

| Msimbo | Maana |
|------|------|
| 200 | Imefanikiwa |
| 400 | Ombi batili |
| 401 | Uthibitishaji umeshindwa |
| 403 | Ufikiaji umekataliwa (mteja asiyefanya kazi, IP imezuiwa) |
| 404 | Rasilimali haijapatikana |
| 405 | Njia isiyoruhusiwa |
| 429 | Kiwango cha kikomo kimezidishwa |
| 500 | Kosa la ndani la seva |
| 502 | Kosa la API ya juu (kurudisha kwote kumeshindwa) |

---

## Mkusanyiko wa mifano ya cURL

```bash
# ─── Proksi ───────────────────────────────────────────────────────────────────

# Ukaguzi wa afya
curl https://localhost:56244/health

# Hali ya kina
curl https://localhost:56244/status

# Orodha ya modeli (zote)
curl https://localhost:56244/api/models

# Modeli za Google pekee
curl "https://localhost:56244/api/models?service=google"

# Utafutaji wa modeli bure
curl "https://localhost:56244/api/models?q=alpha"

# Kubadilisha modeli (ndani)
curl -X PUT https://localhost:56244/api/config/model \
  -H "Content-Type: application/json" \
  -d '{"service":"google","model":"gemini-2.5-pro"}'

# Kusasisha mipangilio
curl -X POST https://localhost:56244/reload

# Kupiga simu Gemini API moja kwa moja
curl -X POST "https://localhost:56244/google/v1beta/models/gemini-2.5-flash:generateContent" \
  -H "Content-Type: application/json" \
  -d '{"contents":[{"role":"user","parts":[{"text":"안녕하세요"}]}]}'

# Inayolingana na OpenAI (modeli chaguo-msingi)
curl -X POST https://localhost:56244/v1/chat/completions \
  -H "Content-Type: application/json" \
  -d '{"model":"gemini-2.5-flash","messages":[{"role":"user","content":"안녕"}]}'

# Muundo wa provider/model wa OpenClaw
curl -X POST https://localhost:56244/v1/chat/completions \
  -H "Authorization: Bearer my-agent-token" \
  -H "Content-Type: application/json" \
  -d '{"model":"wall-vault/gemini-2.5-flash","messages":[{"role":"user","content":"안녕"}]}'

# Kutumia modeli ya 1M context bure
curl -X POST https://localhost:56244/v1/chat/completions \
  -H "Authorization: Bearer my-agent-token" \
  -H "Content-Type: application/json" \
  -d '{"model":"wall-vault/hunter-alpha","messages":[{"role":"user","content":"안녕"}]}'

# ─── Hifadhi ya funguo (umma) ───────────────────────────────────────────────────────────

curl https://localhost:56243/api/status
curl https://localhost:56243/api/clients
curl -s https://localhost:56243/api/events --max-time 3

# ─── Hifadhi ya funguo (msimamizi) ─────────────────────────────────────────────────────────

ADMIN="Authorization: Bearer admin-token"

# Orodha ya funguo
curl -H "$ADMIN" https://localhost:56243/admin/keys

# Kuongeza ufunguo wa Google
curl -X POST https://localhost:56243/admin/keys \
  -H "$ADMIN" -H "Content-Type: application/json" \
  -d '{"service":"google","key":"AIzaSy...","label":"메인 키","daily_limit":1000}'

# Kuongeza ufunguo wa OpenAI
curl -X POST https://localhost:56243/admin/keys \
  -H "$ADMIN" -H "Content-Type: application/json" \
  -d '{"service":"openai","key":"sk-...","label":"GPT 키"}'

# Kuongeza ufunguo wa OpenRouter
curl -X POST https://localhost:56243/admin/keys \
  -H "$ADMIN" -H "Content-Type: application/json" \
  -d '{"service":"openrouter","key":"sk-or-v1-...","label":"OR 키"}'

# Kufuta ufunguo (SSE key_deleted inatangazwa)
curl -X DELETE https://localhost:56243/admin/keys/key-abc123 -H "$ADMIN"

# Kuseti upya matumizi ya kila siku
curl -X POST https://localhost:56243/admin/keys/reset -H "$ADMIN"

# Orodha ya wateja
curl -H "$ADMIN" https://localhost:56243/admin/clients

# Kuongeza mteja (OpenClaw)
curl -X POST https://localhost:56243/admin/clients \
  -H "$ADMIN" -H "Content-Type: application/json" \
  -d '{"id":"bot-a","name":"봇 A","agent_type":"openclaw","work_dir":"~/.openclaw","default_service":"google","default_model":"gemini-2.5-flash"}'

# Kubadilisha modeli ya mteja (SSE mara moja)
curl -X PUT https://localhost:56243/admin/clients/bot-a \
  -H "$ADMIN" -H "Content-Type: application/json" \
  -d '{"default_service":"openrouter","default_model":"wall-vault/hunter-alpha"}'

# Kuzima mteja
curl -X PUT https://localhost:56243/admin/clients/my-bot \
  -H "$ADMIN" -H "Content-Type: application/json" \
  -d '{"enabled":false}'

# Kufuta mteja
curl -X DELETE https://localhost:56243/admin/clients/my-bot -H "$ADMIN"

# Orodha ya huduma
curl -H "$ADMIN" https://localhost:56243/admin/services

# Kuweka URL ya ndani ya Ollama (SSE service_changed inatangazwa)
curl -X PUT https://localhost:56243/admin/services/ollama \
  -H "$ADMIN" -H "Content-Type: application/json" \
  -d '{"local_url":"http://192.168.x.x:11434","enabled":true}'

# Kuwezesha huduma ya OpenAI
curl -X PUT https://localhost:56243/admin/services/openai \
  -H "$ADMIN" -H "Content-Type: application/json" \
  -d '{"enabled":true}'

# Kuongeza huduma maalum (SSE service_changed inatangazwa)
curl -X POST https://localhost:56243/admin/services \
  -H "$ADMIN" -H "Content-Type: application/json" \
  -d '{"id":"my-llm","name":"사내 LLM","local_url":"http://10.0.0.50:8080","enabled":true}'

# Kufuta huduma maalum
curl -X DELETE https://localhost:56243/admin/services/my-llm -H "$ADMIN"

# Kukagua orodha ya modeli
curl -H "$ADMIN" https://localhost:56243/admin/models
curl -H "$ADMIN" "https://localhost:56243/admin/models?service=openrouter"
curl -H "$ADMIN" "https://localhost:56243/admin/models?q=hunter"

# Hali ya proksi (heartbeat)
curl -H "$ADMIN" https://localhost:56243/admin/proxies

# ─── Hali ya usambazaji — Proksi → Hifadhi ───────────────────────────────────────────────

# Kukagua funguo zilizosimbuliwa
curl https://localhost:56243/api/keys \
  -H "Authorization: Bearer your-bot-a-token"

# Kutuma Heartbeat
curl -X POST https://localhost:56243/api/heartbeat \
  -H "Authorization: Bearer your-bot-a-token" \
  -H "Content-Type: application/json" \
  -d '{"client_id":"bot-a","version":"v0.1.6.20260314.231308","service":"google","model":"gemini-2.5-flash","sse_connected":true}'
```

---

## Kati-programu

Inatumika kiotomatiki kwa maombi yote:

| Kati-programu | Kazi |
|---------|------|
| **Logger** | Kuweka kumbukumbu katika muundo `[method] path status latencyms` |
| **CORS** | `Access-Control-Allow-Origin: *` |
| **Recovery** | Kurejesha kutoka hofu, kurudisha jibu la 500 |

---

*Ilisasishwa mwisho: 2026-03-16 — v0.1.7: dashboard title renamed to "벽금고(wall-vault) 대시보드", logo moved to sticky topbar, today_attempts tracking, HTTP 582 cooldown (5min), share-of-total bar scaling, custom/ routing fix, Ollama Timeout:0, key_att i18n, avatar heartbeat sync, build timestamp versioning, proxy-only service filter*
