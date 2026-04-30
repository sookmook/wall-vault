# Jagorar API na wall-vault

Wannan takarda tana bayyana dukkan wuraren API na HTTP na wall-vault cikin cikakken bayani.

---

## Abubuwan da ke ciki

- [Tantancewa](#tantancewa)
- [API na Proxy (:56244)](#api-na-proxy-56244)
  - [Duba lafiya](#get-health)
  - [Duba yanayi](#get-status)
  - [Jerin model](#get-apimodels)
  - [Canza model](#put-apiconfigmodel)
  - [Yanayin tunani](#put-apiconfigthink-mode)
  - [Sabunta saiti](#post-reload)
  - [Gemini API](#post-googlev1betamodelsmgeneratecontent)
  - [Gemini streaming](#post-googlev1betamodelsmstreamgeneratecontent)
  - [API mai dacewa da OpenAI](#post-v1chatcompletions)
- [API na Makullin Asiri (:56243)](#api-na-makullin-asiri-56243)
  - [API na jama'a](#api-na-jamaa-ba-a-bukata-tantancewa)
  - [Rafin abubuwan SSE](#get-apievents)
  - [API na proxy kadai](#api-na-proxy-kadai-alamar-abokin-ciniki)
  - [API na gudanarwa — Makullai](#api-na-gudanarwa--makullai-api)
  - [API na gudanarwa — Abokan ciniki](#api-na-gudanarwa--abokan-ciniki)
  - [API na gudanarwa — Ayyuka](#api-na-gudanarwa--ayyuka)
  - [API na gudanarwa — Jerin model](#api-na-gudanarwa--jerin-model)
  - [API na gudanarwa — Yanayin proxy](#api-na-gudanarwa--yanayin-proxy)
- [Nau'o'in abubuwan SSE](#nau-o-in-abubuwan-sse)
- [Jagorar hanyar mai bada sabis·model](#jagoran-hanyar-mai-bada-sabismodel)
- [Tsarin bayanan](#tsarin-bayanan)
- [Amsar kuskure](#amsar-kuskure)
- [Tarin misalan cURL](#tarin-misalan-curl)

---

## Tantancewa

| Yanki | Hanya | Taken kai |
|------|------|------|
| API na gudanarwa | Alamar Bearer | `Authorization: Bearer <admin_token>` |
| Proxy → Makulli | Alamar Bearer | `Authorization: Bearer <client_token>` |
| API na proxy | Babu (na gida) | — |

Idan ba a saita `admin_token` ba (kirtani mai fanko) dukkan API na gudanarwa za a iya shiga ba tare da tantancewa ba.

### Tsarin tsaro

- **Rate Limiting**: Idan tantancewar API na gudanarwa ta gaza sau 10/minti 15, za a toshe wannan IP na ɗan lokaci (`429 Too Many Requests`)
- **Jerin IP da aka yarda**: IP/CIDR da aka yi rajista a filin `ip_whitelist` na wakili (`Client`) ne kawai za su iya shiga `/api/keys`. Idan jeri fanko ne, duka za su iya shiga.
- **Kariyar theme·lang**: `/admin/theme`, `/admin/lang` ma suna buƙatar tantancewar alamar gudanarwa

---

## API na Proxy (:56244)

Uwar garken da proxy ke gudana. Tashar da ba ta canzawa `56244`.

---

### `GET /health`

Duba lafiya. Koyaushe tana mayar da 200 OK.

**Misalin amsa:**
```json
{
  "status": "ok",
  "version": "v0.1.6.20260314.231308",
  "client": "bot-a"
}
```

---

### `GET /status`

Cikakken bayani akan yanayin proxy.

**Misalin amsa:**
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

| Filin | Iri | Bayani |
|------|------|------|
| `service` | string | Sabis na yanzu |
| `model` | string | Model na yanzu |
| `sse` | bool | Yanayin haɗin SSE na makulli |
| `filter` | string | Yanayin tacen kayan aiki |
| `services` | []string | Jerin ayyuka masu aiki |
| `mode` | string | `standalone` \| `distributed` |

---

### `GET /api/models`

Jerin model ɗin da ake samu. Yana amfani da TTL cache (minti 10 da ba ta canzawa).

**Sigogin tambaya:**

| Sigar | Bayani | Misali |
|---------|------|------|
| `service` | Tacen sabis | `?service=google` |
| `q` | Binciken ID/sunan model | `?q=gemini` |

**Misalin amsa:**
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

| Filin | Iri | Bayani |
|------|------|------|
| `id` | string | ID na model |
| `name` | string | Sunan nuni na model |
| `service` | string | `google` \| `openai` \| `anthropic` \| `openrouter` \| `ollama` da sauransu |
| `context_length` | int | Girman tagar mahallin |
| `free` | bool | Shin model ɗin kyauta ne (OpenRouter) |

---

### `PUT /api/config/model`

Canza sabis·model na yanzu.

**Jikin buƙata:**
```json
{
  "service": "google",
  "model": "gemini-2.5-flash"
}
```

**Amsa:**
```json
{
  "status": "ok",
  "service": "google",
  "model": "gemini-2.5-flash"
}
```

> **Lura:** A cikin yanayin distributed, ana ba da shawarar amfani da `PUT /admin/clients/{id}` na makulli maimakon wannan API. Canje-canjen makulli na nuna ta hanyar SSE cikin dakika 1–3 kai tsaye.

---

### `PUT /api/config/think-mode`

Canza yanayin tunani (no-op, don faɗaɗawa nan gaba).

**Amsa:**
```json
{"status": "ok"}
```

---

### `POST /reload`

Sake daidaita saitunan abokin ciniki·makullai daga makulli nan take.

**Amsa:**
```json
{"status": "reloading"}
```

Sake daidaitawar tana gudana a baya, don haka tana ƙarewa cikin dakika 1–2 bayan karɓar amsa.

---

### `POST /google/v1beta/models/{model}:generateContent`

Wakili na Gemini API (ba streaming ba).

**Sigogin hanya:**
- `{model}`: ID na model. Idan yana da prefix `gemini-` za a zaɓi sabis na Google kai tsaye.

**Jikin buƙata:** [Tsarin buƙatar Gemini generateContent](https://ai.google.dev/api/generate-content)

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

**Jikin amsa:** Tsarin amsar Gemini generateContent

**Tacen kayan aiki:** Idan aka saita `tool_filter: strip_all` jerin `tools` na buƙata za a cire shi kai tsaye.

**Sarkar madadin:** Sabis da aka fayyace ya gaza → madadin bisa tsarin sabis da aka saita → Ollama (na ƙarshe).

---

### `POST /google/v1beta/models/{model}:streamGenerateContent`

Wakili na Gemini API streaming. Tsarin buƙata iri ɗaya ne da ba streaming ba. Amsa ta zo a matsayin rafin SSE:

```
data: {"candidates":[{"content":{"parts":[{"text":"안"}],...},...}]}

data: [DONE]
```

---

### `POST /v1/chat/completions`

API mai dacewa da OpenAI. Ana canza zuwa tsarin Gemini a ciki kafin sarrafawa.

**Jikin buƙata:**
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

**Goyon bayan prefix na mai bada sabis a filin `model` (OpenClaw 3.11+):**

| Misalin model | Jagorar hanya |
|-----------|--------|
| `gemini-2.5-flash` | Sabis na saitin yanzu |
| `google/gemini-2.5-pro` | Google kai tsaye |
| `openai/gpt-4o` | OpenAI kai tsaye |
| `anthropic/claude-opus-4-6` | Ta hanyar OpenRouter |
| `openrouter/meta-llama/llama-3.3-70b` | OpenRouter kai tsaye |
| `wall-vault/gemini-2.5-flash` | Gano kai tsaye → Google |
| `wall-vault/claude-opus-4-6` | Gano kai tsaye → OpenRouter Anthropic |
| `wall-vault/gpt-4o` | Gano kai tsaye → OpenAI |
| `wall-vault/hunter-alpha` | OpenRouter (kyauta 1M context) |
| `moonshot/kimi-k2.5` | Ta hanyar OpenRouter |
| `opencode-go/model` | Ta hanyar OpenRouter |
| `kimi-k2.5:cloud` | Ƙarshen `:cloud` → OpenRouter |

Don cikakken bayani duba [Jagorar hanyar mai bada sabis·model](#jagoran-hanyar-mai-bada-sabismodel).

**Jikin amsa:**
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

> **Cire alamomin sarrafa model kai tsaye:** Idan amsa ta ƙunshi rarrabuwar GLM-5 / DeepSeek / ChatML (`<|im_start|>`, `[gMASK]`, `[sop]` da sauransu) za a cire su kai tsaye.

---

## API na Makullin Asiri (:56243)

Uwar garken da makullin asiri ke gudana. Tashar da ba ta canzawa `56243`.

---

### API na jama'a (ba a buƙata tantancewa)

#### `GET /`

UI na dashboard na yanar gizo. Shiga ta hanyar mashigar yanar gizo.

---

#### `GET /api/status`

Duba yanayin makulli.

**Misalin amsa:**
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

Jerin abokan ciniki da aka yi rajista (bayanin jama'a kawai, ba tare da alama ba).

---

### `GET /api/events`

Rafin abubuwan SSE (Server-Sent Events) na lokaci na gaske.

**Taken kai:**
```
Content-Type: text/event-stream
Cache-Control: no-cache
Connection: keep-alive
```

**Ana karɓa nan take bayan haɗawa:**
```
data: {"type":"connected","clients":2}
```

**Misalin abubuwa:**
```
data: {"type":"config_change","data":{"client_id":"bot-a","service":"google","model":"gemini-2.5-pro"}}

data: {"type":"key_added","data":{"service":"openrouter"}}

data: {"type":"key_deleted","data":{"service":"google"}}

data: {"type":"service_changed","data":{"action":"updated","id":"ollama"}}

data: {"type":"usage_reset","data":{"time":"2026-03-13T00:00:30Z"}}
```

Don cikakken nau'o'in abubuwa duba [Nau'o'in abubuwan SSE](#nau-o-in-abubuwan-sse).

---

### API na proxy kadai (alamar abokin ciniki)

Ana buƙatar taken kai `Authorization: Bearer <client_token>`. Ana iya tantancewa da alamar gudanarwa ma.

#### `GET /api/keys`

Jerin makullan API da aka ɓuɓɓuka da ake bai wa proxy.

**Sigogin tambaya:**

| Sigar | Bayani |
|---------|------|
| `service` | Tacen sabis (misali: `?service=google`) |

**Misalin amsa:**
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

> **Tsaro:** Yana mayar da makulli a fayyace. Ana mayar da makullan sabis da aka yarda kawai bisa saitin `allowed_services` na abokin ciniki.

---

#### `GET /api/services`

Jerin ayyukan da proxy ke amfani da su. Yana mayar da jerin ID na sabis inda `proxy_enabled=true`.

**Misalin amsa:**
```json
["google", "ollama"]
```

Idan jeri fanko ne, proxy yana amfani da dukkan ayyuka ba tare da takurawa ba.

---

#### `POST /api/heartbeat`

Aika yanayin proxy (ana aiwatar da shi kai tsaye kowane dakika 20).

**Jikin buƙata:**
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

| Filin | Iri | Bayani |
|------|------|------|
| `client_id` | string | ID na abokin ciniki |
| `version` | string | Sigar proxy (tare da alamar lokacin gini, misali `v0.1.6.20260314.231308`) |
| `service` | string | Sabis na yanzu |
| `model` | string | Model na yanzu |
| `sse_connected` | bool | Yanayin haɗin SSE |
| `host` | string | Sunan mai masauki |
| `avatar` | string | base64 data URI of the proxy's local avatar image (sent by proxy, auto-persisted to client record by vault). Set via `WV_AVATAR` env var (relative path under `~/.openclaw/`). |
| `key_usage` | map[string]int | Key ID → successful tokens used today. Vault calls `SetKeyUsage()` (absolute value sync). 429/402/582 errors do **not** increment this. |
| `key_attempts` | map[string]int | Key ID → total API calls today (success + rate-limited). Vault calls `SetKeyAttempts()`. 429/402/582 errors only increment this, not `key_usage`. |
| `key_cooldowns` | map[string]string | Key ID → cooldown end time (RFC3339). Vault calls `SetKeyCooldownIfLater()`. |

**Amsa:**
```json
{"status": "ok"}
```

---

### API na gudanarwa — Makullai API

Ana buƙatar taken kai `Authorization: Bearer <admin_token>`.

#### `GET /admin/keys`

Jerin dukkan makullan API da aka yi rajista (ban da makulli a fayyace).

**Misalin amsa:**
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

| Filin | Iri | Bayani |
|------|------|------|
| `today_usage` | int | Adadin alamomin buƙatu da suka yi nasara a yau (429/402/582 kuskure ba a haɗa ba) |
| `today_attempts` | int | Jimlar kiran API a yau (nasara + rate-limited haɗe) |
| `available` | bool | Ana iya amfani ba tare da cooldown·iyaka ba |
| `usage_pct` | int | Kashi na % na amfani dangane da iyakar yau da kullum (`daily_limit=0` ya zama 0) |
| `cooldown_until` | RFC3339 | Lokacin ƙarshen cooldown (idan ƙimar sifili babu) |
| `last_error` | int | Lambar kuskuren HTTP na ƙarshe |

---

#### `POST /admin/keys`

Yin rajistar sabon makullin API. Nan take bayan rajista SSE `key_added` abin da ya faru ya yi watsa.

**Jikin buƙata:**
```json
{
  "service": "google",
  "key": "AIzaSy...",
  "label": "메인 키",
  "daily_limit": 1000
}
```

| Filin | Wajibi | Bayani |
|------|------|------|
| `service` | ✅ | `google` \| `openai` \| `anthropic` \| `openrouter` \| `ollama` \| na musamman |
| `key` | ✅ | Makullin API a fayyace |
| `label` | — | Alamar ganewa |
| `daily_limit` | — | Iyakar amfani ta yau da kullum (0 = marar iyaka) |

---

#### `DELETE /admin/keys/{id}`

Share makullin API. Bayan sharewa SSE `key_deleted` abin da ya faru ya yi watsa.

**Amsa:**
```json
{"status": "deleted"}
```

---

#### `POST /admin/keys/reset`

Sake saita adadin amfani na yau da kullum na dukkan makullai. SSE `usage_reset` abin da ya faru ya yi watsa.

**Amsa:**
```json
{
  "status": "reset",
  "time": "2026-03-13T15:00:00Z"
}
```

---

### API na gudanarwa — Abokan ciniki

#### `GET /admin/clients`

Jerin dukkan abokan ciniki (tare da alama).

---

#### `POST /admin/clients`

Yin rajistar sabon abokin ciniki.

**Jikin buƙata:**
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

| Filin | Wajibi | Bayani |
|------|------|------|
| `id` | ✅ | ID na musamman na abokin ciniki |
| `name` | — | Sunan nuni |
| `token` | — | Alamar tantancewa (idan aka bari za a ƙirƙira kai tsaye) |
| `default_service` | — | Sabis na asali |
| `default_model` | — | Model na asali |
| `allowed_services` | — | Jerin ayyukan da aka yarda (jeri fanko = duka) |
| `agent_type` | — | `openclaw` \| `claude-code` \| `cursor` \| `vscode` \| `custom` |
| `work_dir` | — | Jakar aikin wakili |
| `description` | — | Bayani kan wakili |
| `ip_whitelist` | — | Jerin IP da aka yarda (jeri fanko = duka, CIDR a goye) |
| `enabled` | — | Yanayin aiki (ƙimar asali `true`) |

---

#### `GET /admin/clients/{id}`

Duba takamaiman abokin ciniki (tare da alama).

---

#### `PUT /admin/clients/{id}`

Canza saitin abokin ciniki. **SSE `config_change` watsa → ana nuna a cikin proxy cikin dakika 1–3.**

**Jikin buƙata (filayen da za a canza kawai):**
```json
{
  "default_service": "openrouter",
  "default_model": "anthropic/claude-opus-4-6",
  "enabled": true
}
```

**Amsa:**
```json
{"status": "updated"}
```

---

#### `DELETE /admin/clients/{id}`

Share abokin ciniki.

---

### API na gudanarwa — Ayyuka

#### `GET /admin/services`

Jerin ayyukan da aka yi rajista.

**Misalin amsa:**
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

Ayyuka 8 da aka bayar tun farko: `google`, `openai`, `anthropic`, `openrouter`, `github-copilot`, `ollama`, `lmstudio`, `vllm`

---

#### `POST /admin/services`

Ƙara sabis na musamman. Bayan ƙarawa SSE `service_changed` abin da ya faru ya yi watsa → **ana sabunta dropdown na dashboard nan take**.

**Jikin buƙata:**
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

Sabunta saitin sabis. Bayan canzawa SSE `service_changed` abin da ya faru ya yi watsa.

**Jikin buƙata:**
```json
{
  "local_url": "http://192.168.x.x:11434",
  "enabled": true
}
```

---

#### `DELETE /admin/services/{id}`

Share sabis na musamman. Bayan sharewa SSE `service_changed` abin da ya faru ya yi watsa.

Yunƙurin share sabis na asali (`custom: false`):
```json
{"error": "기본 서비스는 삭제할 수 없습니다: google"}
```

---

### API na gudanarwa — Jerin model

#### `GET /admin/models`

Jerin model bisa sabis. Ana amfani da TTL cache (minti 10).

**Sigogin tambaya:**

| Sigar | Bayani | Misali |
|---------|------|------|
| `service` | Tacen sabis | `?service=google` |
| `q` | Binciken model | `?q=gemini` |

**Hanyar samun model bisa sabis:**

| Sabis | Hanya | Adadi |
|--------|------|------|
| `google` | Jeri a tsaye | 8 (tare da embedding) |
| `openai` | Jeri a tsaye | 9 |
| `anthropic` | Jeri a tsaye | 6 |
| `github-copilot` | Jeri a tsaye | 6 |
| `openrouter` | Tambayar API mai motsi (idan ya gaza curated fallback 14) | 340+ |
| `ollama` | Tambayar uwar garken gida mai motsi (idan bai amsa ba shawarwari 7) | mai canzawa |
| `lmstudio` | Tambayar uwar garken gida mai motsi | mai canzawa |
| `vllm` | Tambayar uwar garken gida mai motsi | mai canzawa |
| na musamman | OpenAI mai dacewa `/v1/models` | mai canzawa |

**Jerin model na madadin OpenRouter (idan API bai amsa ba):**

| Model | Bayani na musamman |
|------|----------|
| `openrouter/hunter-alpha` | kyauta, 1M context |
| `openrouter/healer-alpha` | kyauta, omni-modal |
| `moonshot/kimi-k2.5` | 256K context |
| `z-ai/glm-5`, `z-ai/glm-4.7-flash` | — |
| `deepseek/deepseek-r1`, `deepseek/deepseek-chat` | — |
| `qwen/qwen-2.5-72b-instruct` | 131K context |
| `minimax/minimax-m2.5` | — |
| `meta-llama/llama-3.3-70b-instruct` | 131K context |

---

### API na gudanarwa — Yanayin proxy

#### `GET /admin/proxies`

Yanayin Heartbeat na ƙarshe na dukkan proxy da suka haɗa.

---

## Nau'o'in abubuwan SSE

Abubuwan da ake karɓa daga rafin `/api/events` na makulli:

| `type` | Yanayin faruwa | Abubuwan `data` | Amsar dashboard |
|--------|-----------|-------------|--------------|
| `connected` | Nan take bayan haɗin SSE | `{"clients": N}` | — |
| `config_change` | Canjin saitin abokin ciniki | `{"client_id","service","model"}` | Ana sabunta dropdown na model na katin wakili |
| `key_added` | Sabon rajistar makullin API | `{"service": "google"}` | Ana sabunta dropdown na model |
| `key_deleted` | An share makullin API | `{"service": "google"}` | Ana sabunta dropdown na model |
| `service_changed` | Ƙara/gyara/share sabis | `{"action":"added"\|"updated"\|"deleted","id":"...","proxy_services":["google","ollama"]}` | Ana sabunta select na sabis + dropdown na model nan take; ana sabunta jerin ayyukan dispatch na proxy a lokaci na gaske |
| `usage_update` | Lokacin da aka karɓi heartbeat na proxy (kowane dakika 20) | `{"keys":[{"id","service","today_usage","today_attempts","daily_limit","cooldown_until"},...]}` | Ana sabunta sandar·lambar amfani nan take, ana fara ƙidayar lokacin cooldown. Ana amfani da bayanan SSE kai tsaye ba tare da fetch ba. Sandar tana amfani da ma'aunin raba-jimla (makullai marar iyaka). |
| `usage_reset` | Sake saitar adadin amfani na yau da kullum | `{"time": "RFC3339"}` | Ana sabunta shafin |

**Sarrafar abubuwan da proxy ke karɓa:**

```
config_change an karɓa
  → idan client_id ya dace da na kansa
    → ana sabunta service, model nan take
    → hooksMgr.Fire(EventModelChanged)
```

---

## Jagorar hanyar mai bada sabis·model

Idan aka fayyace tsarin `provider/model` a filin `model` na `/v1/chat/completions` za a yi jagorar hanya kai tsaye (mai dacewa da OpenClaw 3.11).

### Dokokin jagorar hanya ta prefix

| Prefix | Makasudin jagorar hanya | Misali |
|--------|------------|------|
| `google/` | Google kai tsaye | `google/gemini-2.5-pro` |
| `openai/` | OpenAI kai tsaye | `openai/gpt-4o` |
| `anthropic/` | Ta hanyar OpenRouter | `anthropic/claude-opus-4-6` |
| `ollama/` | Ollama kai tsaye | `ollama/qwen3.5:35b` |
| `custom/` | Sake fasara (cire `custom/` sai a sake jagorar hanya) | `custom/google/gemini-2.5-flash` → Google |
| `openrouter/` | OpenRouter (hanyar bare a kiyaye) | `openrouter/meta-llama/llama-3.3-70b` |
| `opencode/`, `opencode-go/`, `opencode-zen/` | OpenRouter | `opencode-go/model` |
| `moonshot/`, `kimi-coding/` | OpenRouter (hanyar cikakkiya a kiyaye) | `moonshot/kimi-k2.5` |
| `groq/`, `mistral/`, `minimax/`, `cohere/`, `perplexity/` | OpenRouter | `groq/llama-3.1-70b` |
| `together/`, `huggingface/`, `nvidia/`, `venice/` | OpenRouter | — |
| `meta-llama/`, `qwen/`, `deepseek/`, `01-ai/` | OpenRouter (hanyar cikakkiya) | `deepseek/deepseek-r1` |

### Gano `wall-vault/` prefix kai tsaye

wall-vault yana gane sabis kai tsaye daga ID na model ta hanyar prefix ɗin kansa.

| Tsarin ID na model | Jagorar hanya |
|-------------|--------|
| `wall-vault/gemini-*` | Google |
| `wall-vault/gpt-*`, `wall-vault/o1`, `wall-vault/o3`, `wall-vault/o4*` | OpenAI |
| `wall-vault/claude-*` | OpenRouter (hanyar Anthropic) |
| `wall-vault/hunter-alpha`, `wall-vault/healer-alpha` | OpenRouter (kyauta 1M ctx) |
| `wall-vault/kimi-*`, `wall-vault/glm-*`, `wall-vault/deepseek-*`, `wall-vault/qwen*` | OpenRouter |
| Sauran | OpenRouter |

### Sarrafar ƙarshen `:cloud`

Ana cire ƙarshen `:cloud` na tsarin tag na Ollama kai tsaye sai a aika zuwa OpenRouter.

```
kimi-k2.5:cloud  →  OpenRouter, ID na model: kimi-k2.5
glm-5:cloud      →  OpenRouter, ID na model: glm-5
```

### Misalin haɗin OpenClaw openclaw.json

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

Danna maballin **🐾** na katin wakili zai kwafa snippet na saiti na wannan wakili zuwa clipboard kai tsaye.

---

## Tsarin bayanan

### APIKey

| Filin | Iri | Bayani |
|------|------|------|
| `id` | string | ID na musamman na tsarin UUID |
| `service` | string | `google` \| `openai` \| `anthropic` \| `openrouter` \| `ollama` \| na musamman |
| `encrypted_key` | string | Makulli da aka ɓoye ta AES-GCM (Base64) |
| `label` | string | Alamar ganewa |
| `today_usage` | int | Adadin alamomin buƙatu da suka yi nasara a yau (429/402/582 kuskure ba a haɗa ba) |
| `today_attempts` | int | Jimlar kiran API a yau (nasara + rate-limited; ana sake saita tsakar dare) |
| `daily_limit` | int | Iyakar yau da kullum (0 = marar iyaka) |
| `cooldown_until` | time.Time | Lokacin ƙarshen cooldown |
| `last_error` | int | Lambar kuskuren HTTP na ƙarshe |
| `created_at` | time.Time | Lokacin rajista |

**Tsarin cooldown:**

| Kuskuren HTTP | Cooldown |
|-----------|--------|
| 429 (Too Many Requests) | Minti 30 |
| 402 (Payment Required) | Awa 24 |
| 400 / 401 / 403 | Awa 24 |
| 582 (Gateway Overload) | Minti 5 |
| Kuskuren hanyar sadarwa | Minti 10 |

> **429·402·582**: Ana saita cooldown + ana ƙara `today_attempts`. `today_usage` ba ya canzawa (alamomin nasara ne kawai ake lissafa).
> **Ollama (sabis na gida)**: `callOllama` yana amfani da abokin ciniki HTTP na musamman `Timeout: 0` (marar iyaka). Yin lissafin model ɗin manyan iya ɗaukar dakiku da yawa, don haka ba a aiwatar da lokacin jiran dakika 60 na asali ba.

### Client

| Filin | Iri | Bayani |
|------|------|------|
| `id` | string | ID na musamman na abokin ciniki |
| `name` | string | Sunan nuni |
| `token` | string | Alamar tantancewa |
| `default_service` | string | Sabis na asali |
| `default_model` | string | Model na asali (tsarin `provider/model` zai iya aiki) |
| `allowed_services` | []string | Ayyukan da aka yarda (jeri fanko = duka) |
| `agent_type` | string | `openclaw` \| `claude-code` \| `cursor` \| `vscode` \| `custom` |
| `work_dir` | string | Jakar aikin wakili |
| `description` | string | Bayani |
| `ip_whitelist` | []string | Jerin IP da aka yarda (CIDR a goye) |
| `avatar` | string | Agent avatar — relative path under `~/.openclaw/` (e.g. `workspace/avatar.png`, `workspace/avatars/profile.hpg`) OR base64 data URI (`data:image/png;base64,...`). Supported extensions: `.png`, `.jpg`/`.jpeg`/`.hpg`, `.webp`, `.gif`. |
| `enabled` | bool | Idan `false` za a samu `403` lokacin shiga `/api/keys` |
| `created_at` | time.Time | Lokacin rajista |

### ServiceConfig

| Filin | Iri | Bayani |
|------|------|------|
| `id` | string | ID na musamman na sabis |
| `name` | string | Sunan nuni |
| `local_url` | string | URL na uwar garken gida (Ollama/LMStudio/vLLM/na musamman) |
| `enabled` | bool | Yanayin aiki |
| `custom` | bool | Shin sabis ne da mai amfani ya ƙara |
| `proxy_enabled` | bool | Whether this service is included in proxy dispatch and agent model dropdowns. Controlled by the "프록시 사용" checkbox in the Services card. Only `proxy_enabled: true` services appear in agent service/model selectors. |

### ProxyStatus (Heartbeat)

| Filin | Iri | Bayani |
|------|------|------|
| `client_id` | string | ID na abokin ciniki |
| `version` | string | Sigar proxy (misali `v0.1.6.20260314.231308`) |
| `service` | string | Sabis na yanzu |
| `model` | string | Model na yanzu |
| `sse_connected` | bool | Yanayin haɗin SSE |
| `host` | string | Sunan mai masauki |
| `avatar` | string | base64 data URI of local avatar (auto-synced from proxy via heartbeat) |
| `updated_at` | time.Time | Sabuntawar ƙarshe |
| `vault.today_usage` | int | Adadin alamomin da aka yi amfani da su a yau |
| `vault.daily_limit` | int | Iyakar yau da kullum |
| `vault.key_status` | string | `active` \| `cooldown` \| `exhausted` |

---

## Amsar kuskure

```json
{"error": "오류 메시지"}
```

| Lamba | Ma'ana |
|------|------|
| 200 | Nasara |
| 400 | Buƙatar da ba daidai ba |
| 401 | Tantancewa ta gaza |
| 403 | An hana shiga (abokin ciniki da ba ya aiki, an toshe IP) |
| 404 | Ba a sami albarkatun ba |
| 405 | Hanyar da ba a yarda da ita ba |
| 429 | An wuce iyakar rate limit |
| 500 | Kuskuren uwar garke na ciki |
| 502 | Kuskuren API na sama (dukkan fallback sun gaza) |

---

## Tarin misalan cURL

```bash
# ─── Proxy ───────────────────────────────────────────────────────────────────

# Duba lafiya
curl https://localhost:56244/health

# Duba yanayi
curl https://localhost:56244/status

# Jerin model (duka)
curl https://localhost:56244/api/models

# Model na Google kawai
curl "https://localhost:56244/api/models?service=google"

# Binciken model na kyauta
curl "https://localhost:56244/api/models?q=alpha"

# Canza model (na gida)
curl -X PUT https://localhost:56244/api/config/model \
  -H "Content-Type: application/json" \
  -d '{"service":"google","model":"gemini-2.5-pro"}'

# Sabunta saiti
curl -X POST https://localhost:56244/reload

# Kiran Gemini API kai tsaye
curl -X POST "https://localhost:56244/google/v1beta/models/gemini-2.5-flash:generateContent" \
  -H "Content-Type: application/json" \
  -d '{"contents":[{"role":"user","parts":[{"text":"안녕하세요"}]}]}'

# Mai dacewa da OpenAI (model na asali)
curl -X POST https://localhost:56244/v1/chat/completions \
  -H "Content-Type: application/json" \
  -d '{"model":"gemini-2.5-flash","messages":[{"role":"user","content":"안녕"}]}'

# Tsarin provider/model na OpenClaw
curl -X POST https://localhost:56244/v1/chat/completions \
  -H "Authorization: Bearer my-agent-token" \
  -H "Content-Type: application/json" \
  -d '{"model":"wall-vault/gemini-2.5-flash","messages":[{"role":"user","content":"안녕"}]}'

# Amfani da model na 1M context kyauta
curl -X POST https://localhost:56244/v1/chat/completions \
  -H "Authorization: Bearer my-agent-token" \
  -H "Content-Type: application/json" \
  -d '{"model":"wall-vault/hunter-alpha","messages":[{"role":"user","content":"안녕"}]}'

# ─── Makullin asiri (jama'a) ───────────────────────────────────────────────────────────

curl https://localhost:56243/api/status
curl https://localhost:56243/api/clients
curl -s https://localhost:56243/api/events --max-time 3

# ─── Makullin asiri (gudanarwa) ─────────────────────────────────────────────────────────

ADMIN="Authorization: Bearer admin-token"

# Jerin makullai
curl -H "$ADMIN" https://localhost:56243/admin/keys

# Ƙara makullin Google
curl -X POST https://localhost:56243/admin/keys \
  -H "$ADMIN" -H "Content-Type: application/json" \
  -d '{"service":"google","key":"AIzaSy...","label":"메인 키","daily_limit":1000}'

# Ƙara makullin OpenAI
curl -X POST https://localhost:56243/admin/keys \
  -H "$ADMIN" -H "Content-Type: application/json" \
  -d '{"service":"openai","key":"sk-...","label":"GPT 키"}'

# Ƙara makullin OpenRouter
curl -X POST https://localhost:56243/admin/keys \
  -H "$ADMIN" -H "Content-Type: application/json" \
  -d '{"service":"openrouter","key":"sk-or-v1-...","label":"OR 키"}'

# Share makulli (SSE key_deleted watsa)
curl -X DELETE https://localhost:56243/admin/keys/key-abc123 -H "$ADMIN"

# Sake saita adadin amfani na yau da kullum
curl -X POST https://localhost:56243/admin/keys/reset -H "$ADMIN"

# Jerin abokan ciniki
curl -H "$ADMIN" https://localhost:56243/admin/clients

# Ƙara abokin ciniki (OpenClaw)
curl -X POST https://localhost:56243/admin/clients \
  -H "$ADMIN" -H "Content-Type: application/json" \
  -d '{"id":"bot-a","name":"봇 A","agent_type":"openclaw","work_dir":"~/.openclaw","default_service":"google","default_model":"gemini-2.5-flash"}'

# Canza model na abokin ciniki (SSE nan take)
curl -X PUT https://localhost:56243/admin/clients/bot-a \
  -H "$ADMIN" -H "Content-Type: application/json" \
  -d '{"default_service":"openrouter","default_model":"wall-vault/hunter-alpha"}'

# Kashe abokin ciniki
curl -X PUT https://localhost:56243/admin/clients/my-bot \
  -H "$ADMIN" -H "Content-Type: application/json" \
  -d '{"enabled":false}'

# Share abokin ciniki
curl -X DELETE https://localhost:56243/admin/clients/my-bot -H "$ADMIN"

# Jerin ayyuka
curl -H "$ADMIN" https://localhost:56243/admin/services

# Saita URL na gida na Ollama (SSE service_changed watsa)
curl -X PUT https://localhost:56243/admin/services/ollama \
  -H "$ADMIN" -H "Content-Type: application/json" \
  -d '{"local_url":"http://192.168.x.x:11434","enabled":true}'

# Kunna sabis na OpenAI
curl -X PUT https://localhost:56243/admin/services/openai \
  -H "$ADMIN" -H "Content-Type: application/json" \
  -d '{"enabled":true}'

# Ƙara sabis na musamman (SSE service_changed watsa)
curl -X POST https://localhost:56243/admin/services \
  -H "$ADMIN" -H "Content-Type: application/json" \
  -d '{"id":"my-llm","name":"사내 LLM","local_url":"http://10.0.0.50:8080","enabled":true}'

# Share sabis na musamman
curl -X DELETE https://localhost:56243/admin/services/my-llm -H "$ADMIN"

# Duba jerin model
curl -H "$ADMIN" https://localhost:56243/admin/models
curl -H "$ADMIN" "https://localhost:56243/admin/models?service=openrouter"
curl -H "$ADMIN" "https://localhost:56243/admin/models?q=hunter"

# Yanayin proxy (heartbeat)
curl -H "$ADMIN" https://localhost:56243/admin/proxies

# ─── Yanayin rarraba — Proxy → Makulli ───────────────────────────────────────────────

# Duba makulli da aka ɓuɓɓuka
curl https://localhost:56243/api/keys \
  -H "Authorization: Bearer your-bot-a-token"

# Aika Heartbeat
curl -X POST https://localhost:56243/api/heartbeat \
  -H "Authorization: Bearer your-bot-a-token" \
  -H "Content-Type: application/json" \
  -d '{"client_id":"bot-a","version":"v0.1.6.20260314.231308","service":"google","model":"gemini-2.5-flash","sse_connected":true}'
```

---

## Middleware

Ana aiwatarwa kai tsaye ga dukkan buƙatu:

| Middleware | Aiki |
|---------|------|
| **Logger** | Ana yin log a tsarin `[method] path status latencyms` |
| **CORS** | `Access-Control-Allow-Origin: *` |
| **Recovery** | Farfaɗo daga firgita, mayar da amsa 500 |

---

*An sabunta ƙarshe: 2026-03-16 — v0.1.7: dashboard title renamed to "벽금고(wall-vault) 대시보드", logo moved to sticky topbar, today_attempts tracking, HTTP 582 cooldown (5min), share-of-total bar scaling, custom/ routing fix, Ollama Timeout:0, key_att i18n, avatar heartbeat sync, build timestamp versioning, proxy-only service filter*
