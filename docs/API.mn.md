# wall-vault API гарын авлага

Энэ баримт бичиг нь wall-vault-ийн бүх HTTP API эндпойнтуудыг нарийвчлан тайлбарласан болно.

---

## Агуулга

- [Нэвтрэлт баталгаажуулалт](#нэвтрэлт-баталгаажуулалт)
- [Прокси API (:56244)](#прокси-api-56244)
  - [Эрүүл мэндийн шалгалт](#get-health)
  - [Төлөв лавлах](#get-status)
  - [Моделийн жагсаалт](#get-apimodels)
  - [Модель солих](#put-apiconfigmodel)
  - [Бодолтын горим](#put-apiconfigthink-mode)
  - [Тохиргоо шинэчлэх](#post-reload)
  - [Gemini API](#post-googlev1betamodelsmgeneratecontent)
  - [Gemini стриминг](#post-googlev1betamodelsmstreamgeneratecontent)
  - [OpenAI-тай нийцтэй API](#post-v1chatcompletions)
- [Түлхүүрийн сейф API (:56243)](#түлхүүрийн-сейф-api-56243)
  - [Нээлттэй API](#нээлттэй-api-баталгаажуулалт-шаардлагагүй)
  - [SSE үйл явдлын стрим](#get-apievents)
  - [Зөвхөн прокси API](#зөвхөн-прокси-api-клиентийн-токен)
  - [Админ API — Түлхүүрүүд](#админ-api--api-түлхүүрүүд)
  - [Админ API — Клиентүүд](#админ-api--клиентүүд)
  - [Админ API — Үйлчилгээнүүд](#админ-api--үйлчилгээнүүд)
  - [Админ API — Моделийн жагсаалт](#админ-api--моделийн-жагсаалт)
  - [Админ API — Проксийн төлөв](#админ-api--проксийн-төлөв)
- [SSE үйл явдлын төрлүүд](#sse-үйл-явдлын-төрлүүд)
- [Үйлчилгээ үзүүлэгч ба модель чиглүүлэлт](#үйлчилгээ-үзүүлэгч-ба-модель-чиглүүлэлт)
- [Өгөгдлийн схем](#өгөгдлийн-схем)
- [Алдааны хариу](#алдааны-хариу)
- [cURL жишээний цуглуулга](#curl-жишээний-цуглуулга)

---

## Нэвтрэлт баталгаажуулалт

| Хүрээ | Арга | Толгой |
|-------|------|--------|
| Админ API | Bearer токен | `Authorization: Bearer <admin_token>` |
| Прокси → Сейф | Bearer токен | `Authorization: Bearer <client_token>` |
| Прокси API | Байхгүй (локаль) | — |

Хэрэв `admin_token` тохируулаагүй бол (хоосон тэмдэгт мөр) бүх админ API-д баталгаажуулалтгүйгээр хандаж болно.

### Аюулгүй байдлын бодлого

- **Хурд хязгаарлалт**: Админ API-ийн баталгаажуулалтын алдаа 15 минутад 10 удаагаас давсан тохиолдолд тухайн IP-г түр блоклоно (`429 Too Many Requests`)
- **IP цагаан жагсаалт**: Агент (`Client`)-ийн `ip_whitelist` талбарт бүртгэгдсэн IP/CIDR-ээс зөвхөн `/api/keys`-д хандаж болно. Хоосон массив бол бүгдийг зөвшөөрнө.
- **theme·lang хамгаалалт**: `/admin/theme`, `/admin/lang` ч мөн админ токен баталгаажуулалт шаардана

---

## Прокси API (:56244)

Прокси ажилладаг сервер. Анхдагч порт `56244`.

---

### `GET /health`

Эрүүл мэндийн шалгалт. Үргэлж 200 OK буцаана.

**Хариуны жишээ:**
```json
{
  "status": "ok",
  "version": "v0.1.6.20260314.231308",
  "client": "bot-a"
}
```

---

### `GET /status`

Проксийн төлвийн дэлгэрэнгүй лавлагаа.

**Хариуны жишээ:**
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

| Талбар | Төрөл | Тайлбар |
|--------|-------|---------|
| `service` | string | Одоогийн анхдагч үйлчилгээ |
| `model` | string | Одоогийн анхдагч модель |
| `sse` | bool | Сейфтэй SSE холболтын төлөв |
| `filter` | string | Хэрэгслийн шүүлтүүрийн горим |
| `services` | []string | Идэвхтэй үйлчилгээнүүдийн жагсаалт |
| `mode` | string | `standalone` \| `distributed` |

---

### `GET /api/models`

Боломжтой моделүүдийн жагсаалт лавлах. TTL кэш (анхдагч 10 минут) ашигладаг.

**Асуулгын параметрүүд:**

| Параметр | Тайлбар | Жишээ |
|----------|---------|-------|
| `service` | Үйлчилгээний шүүлтүүр | `?service=google` |
| `q` | Моделийн ID/нэрээр хайх | `?q=gemini` |

**Хариуны жишээ:**
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

| Талбар | Төрөл | Тайлбар |
|--------|-------|---------|
| `id` | string | Моделийн ID |
| `name` | string | Моделийн харуулах нэр |
| `service` | string | `google` \| `openai` \| `anthropic` \| `openrouter` \| `ollama` гэх мэт |
| `context_length` | int | Контекст цонхны хэмжээ |
| `free` | bool | Үнэгүй модель эсэх (OpenRouter) |

---

### `PUT /api/config/model`

Одоогийн үйлчилгээ ба модель солих.

**Хүсэлтийн бие:**
```json
{
  "service": "google",
  "model": "gemini-2.5-flash"
}
```

**Хариу:**
```json
{
  "status": "ok",
  "service": "google",
  "model": "gemini-2.5-flash"
}
```

> **Тэмдэглэл:** Тархсан горимд энэ API-ийн оронд сейфийн `PUT /admin/clients/{id}` ашиглахыг зөвлөж байна. Сейфийн өөрчлөлтүүд SSE-ээр дамжуулан 1-3 секундын дотор автоматаар тусгагдана.

---

### `PUT /api/config/think-mode`

Бодолтын горимыг сэлгэх (no-op, ирээдүйд өргөтгөхөд зориулсан).

**Хариу:**
```json
{"status": "ok"}
```

---

### `POST /reload`

Сейфээс клиентийн тохиргоо болон түлхүүрүүдийг шууд дахин синхрончлох.

**Хариу:**
```json
{"status": "reloading"}
```

Дахин синхрончлол нь асинхрон ажиллах тул хариу хүлээн авсны дараа 1-2 секундын дотор дуусна.

---

### `POST /google/v1beta/models/{model}:generateContent`

Gemini API прокси (стримингүй).

**Замын параметр:**
- `{model}`: Моделийн ID. `gemini-` угтвартай бол Google үйлчилгээ автоматаар сонгогдоно.

**Хүсэлтийн бие:** [Gemini generateContent хүсэлтийн формат](https://ai.google.dev/api/generate-content)

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

**Хариуны бие:** Gemini generateContent хариуны формат

**Хэрэгслийн шүүлтүүр:** `tool_filter: strip_all` тохиргоотой үед хүсэлтийн `tools` массив автоматаар устгагдана.

**Нөөц гинж:** Заасан үйлчилгээ амжилтгүй → тохируулсан үйлчилгээний дарааллаар нөөцлөх → Ollama (хамгийн сүүлд).

---

### `POST /google/v1beta/models/{model}:streamGenerateContent`

Gemini API стриминг прокси. Хүсэлтийн формат стримингүйтэй ижил. Хариу нь SSE стрим:

```
data: {"candidates":[{"content":{"parts":[{"text":"안"}],...},...}]}

data: [DONE]
```

---

### `POST /v1/chat/completions`

OpenAI-тай нийцтэй API. Дотооддоо Gemini формат руу хөрвүүлэн боловсруулна.

**Хүсэлтийн бие:**
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

**`model` талбарт үйлчилгээ үзүүлэгчийн угтвар дэмжлэг (OpenClaw 3.11+):**

| Моделийн жишээ | Чиглүүлэлт |
|---------------|------------|
| `gemini-2.5-flash` | Одоогийн тохируулсан үйлчилгээ |
| `google/gemini-2.5-pro` | Шууд Google руу |
| `openai/gpt-4o` | Шууд OpenAI руу |
| `anthropic/claude-opus-4-6` | OpenRouter-ээр дамжуулан |
| `openrouter/meta-llama/llama-3.3-70b` | Шууд OpenRouter руу |
| `wall-vault/gemini-2.5-flash` | Автомат илрүүлэлт → Google |
| `wall-vault/claude-opus-4-6` | Автомат илрүүлэлт → OpenRouter Anthropic |
| `wall-vault/gpt-4o` | Автомат илрүүлэлт → OpenAI |
| `wall-vault/hunter-alpha` | OpenRouter (үнэгүй 1M context) |
| `moonshot/kimi-k2.5` | OpenRouter-ээр дамжуулан |
| `opencode-go/model` | OpenRouter-ээр дамжуулан |
| `kimi-k2.5:cloud` | `:cloud` дагавар → OpenRouter |

Дэлгэрэнгүй мэдээллийг [Үйлчилгээ үзүүлэгч ба модель чиглүүлэлт](#үйлчилгээ-үзүүлэгч-ба-модель-чиглүүлэлт) хэсгээс үзнэ үү.

**Хариуны бие:**
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

> **Моделийн удирдлагын токенуудын автомат устгал:** Хариу нь GLM-5 / DeepSeek / ChatML тусгаарлагч (`<|im_start|>`, `[gMASK]`, `[sop]` гэх мэт) агуулж байвал автоматаар устгагдана.

---

## Түлхүүрийн сейф API (:56243)

Түлхүүрийн сейф ажилладаг сервер. Анхдагч порт `56243`.

---

### Нээлттэй API (Баталгаажуулалт шаардлагагүй)

#### `GET /`

Вэб хянах самбарын UI. Хөтчөөр хандана.

---

#### `GET /api/status`

Сейфийн төлөв лавлах.

**Хариуны жишээ:**
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

Бүртгэгдсэн клиентүүдийн жагсаалт (зөвхөн нээлттэй мэдээлэл, токенгүй).

---

### `GET /api/events`

SSE (Server-Sent Events) бодит цагийн үйл явдлын стрим.

**Толгой:**
```
Content-Type: text/event-stream
Cache-Control: no-cache
Connection: keep-alive
```

**Холбогдсон даруйд хүлээн авна:**
```
data: {"type":"connected","clients":2}
```

**Үйл явдлын жишээнүүд:**
```
data: {"type":"config_change","data":{"client_id":"bot-a","service":"google","model":"gemini-2.5-pro"}}

data: {"type":"key_added","data":{"service":"openrouter"}}

data: {"type":"key_deleted","data":{"service":"google"}}

data: {"type":"service_changed","data":{"action":"updated","id":"ollama"}}

data: {"type":"usage_reset","data":{"time":"2026-03-13T00:00:30Z"}}
```

Үйл явдлын төрлүүдийн дэлгэрэнгүйг [SSE үйл явдлын төрлүүд](#sse-үйл-явдлын-төрлүүд) хэсгээс үзнэ үү.

---

### Зөвхөн прокси API (Клиентийн токен)

`Authorization: Bearer <client_token>` толгой шаардлагатай. Админ токеноор ч баталгаажуулах боломжтой.

#### `GET /api/keys`

Проксид зориулсан нууцлалгүй болсон API түлхүүрүүдийн жагсаалт.

**Асуулгын параметрүүд:**

| Параметр | Тайлбар |
|----------|---------|
| `service` | Үйлчилгээний шүүлтүүр (жишээ: `?service=google`) |

**Хариуны жишээ:**
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

> **Аюулгүй байдал:** Түлхүүрүүдийг нууцлалгүй буцаана. Клиентийн `allowed_services` тохиргооны дагуу зөвшөөрөгдсөн үйлчилгээний түлхүүрүүдийг л буцаана.

---

#### `GET /api/services`

Проксийн ашигладаг үйлчилгээнүүдийн жагсаалт лавлах. `proxy_enabled=true` үйлчилгээний ID массив буцаана.

**Хариуны жишээ:**
```json
["google", "ollama"]
```

Хоосон массив бол прокси бүх үйлчилгээг хязгаарлалтгүйгээр ашиглана.

---

#### `POST /api/heartbeat`

Проксийн төлөв илгээх (20 секунд тутам автоматаар ажиллана).

**Хүсэлтийн бие:**
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

| Талбар | Төрөл | Тайлбар |
|--------|-------|---------|
| `client_id` | string | Клиентийн ID |
| `version` | string | Проксийн хувилбар (build timestamp-тай, жишээ: `v0.1.6.20260314.231308`) |
| `service` | string | Одоогийн үйлчилгээ |
| `model` | string | Одоогийн модель |
| `sse_connected` | bool | SSE холболтын төлөв |
| `host` | string | Хостын нэр |
| `avatar` | string | base64 data URI of the proxy's local avatar image (sent by proxy, auto-persisted to client record by vault). Set via `WV_AVATAR` env var (relative path under `~/.openclaw/`). |
| `key_usage` | map[string]int | Key ID → successful tokens used today. Vault calls `SetKeyUsage()` (absolute value sync). 429/402/582 errors do **not** increment this. |
| `key_attempts` | map[string]int | Key ID → total API calls today (success + rate-limited). Vault calls `SetKeyAttempts()`. 429/402/582 errors only increment this, not `key_usage`. |
| `key_cooldowns` | map[string]string | Key ID → cooldown end time (RFC3339). Vault calls `SetKeyCooldownIfLater()`. |

**Хариу:**
```json
{"status": "ok"}
```

---

### Админ API — API түлхүүрүүд

`Authorization: Bearer <admin_token>` толгой шаардлагатай.

#### `GET /admin/keys`

Бүртгэгдсэн бүх API түлхүүрүүдийн жагсаалт (нууцлалгүй түлхүүргүй).

**Хариуны жишээ:**
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

| Талбар | Төрөл | Тайлбар |
|--------|-------|---------|
| `today_usage` | int | Өнөөдрийн амжилттай хүсэлтийн токенуудын тоо (429/402/582 алдаа тооцохгүй) |
| `today_attempts` | int | Өнөөдрийн нийт API дуудлагын тоо (амжилттай + хурд хязгаарлалттай) |
| `available` | bool | Хөргөлт болон хязгааргүйгээр ашиглах боломжтой эсэх |
| `usage_pct` | int | Өдрийн хязгаараас ашигласан хувь % (`daily_limit=0` бол 0) |
| `cooldown_until` | RFC3339 | Хөргөлтийн дуусах хугацаа (тэг утга бол байхгүй) |
| `last_error` | int | Сүүлийн HTTP алдааны код |

---

#### `POST /admin/keys`

Шинэ API түлхүүр бүртгэх. Бүртгэгдсэн даруйд SSE `key_added` үйл явдал цацагдана.

**Хүсэлтийн бие:**
```json
{
  "service": "google",
  "key": "AIzaSy...",
  "label": "메인 키",
  "daily_limit": 1000
}
```

| Талбар | Заавал | Тайлбар |
|--------|--------|---------|
| `service` | ✅ | `google` \| `openai` \| `anthropic` \| `openrouter` \| `ollama` \| тусгай |
| `key` | ✅ | API түлхүүрийн нууцлалгүй текст |
| `label` | — | Тодорхойлох шошго |
| `daily_limit` | — | Өдрийн хэрэглээний хязгаар (0 = хязгааргүй) |

---

#### `DELETE /admin/keys/{id}`

API түлхүүр устгах. Устгасны дараа SSE `key_deleted` үйл явдал цацагдана.

**Хариу:**
```json
{"status": "deleted"}
```

---

#### `POST /admin/keys/reset`

Бүх түлхүүрүүдийн өдрийн хэрэглээг эхний байдалд оруулах. SSE `usage_reset` үйл явдал цацагдана.

**Хариу:**
```json
{
  "status": "reset",
  "time": "2026-03-13T15:00:00Z"
}
```

---

### Админ API — Клиентүүд

#### `GET /admin/clients`

Бүх клиентүүдийн жагсаалт (токен багтсан).

---

#### `POST /admin/clients`

Шинэ клиент бүртгэх.

**Хүсэлтийн бие:**
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

| Талбар | Заавал | Тайлбар |
|--------|--------|---------|
| `id` | ✅ | Клиентийн өвөрмөц ID |
| `name` | — | Харуулах нэр |
| `token` | — | Баталгаажуулалтын токен (орхивол автоматаар үүсгэнэ) |
| `default_service` | — | Анхдагч үйлчилгээ |
| `default_model` | — | Анхдагч модель |
| `allowed_services` | — | Зөвшөөрөгдсөн үйлчилгээнүүдийн жагсаалт (хоосон массив = бүгдийг зөвшөөрөх) |
| `agent_type` | — | `openclaw` \| `claude-code` \| `cursor` \| `vscode` \| `custom` |
| `work_dir` | — | Агентын ажлын директор |
| `description` | — | Агентын тайлбар |
| `ip_whitelist` | — | Зөвшөөрөгдсөн IP жагсаалт (хоосон массив = бүгдийг зөвшөөрөх, CIDR дэмжинэ) |
| `enabled` | — | Идэвхтэй эсэх (анхдагч `true`) |

---

#### `GET /admin/clients/{id}`

Тодорхой клиент лавлах (токен багтсан).

---

#### `PUT /admin/clients/{id}`

Клиентийн тохиргоо өөрчлөх. **SSE `config_change` цацагдана → проксид 1-3 секундын дотор тусгагдана.**

**Хүсэлтийн бие (зөвхөн өөрчлөх талбарууд):**
```json
{
  "default_service": "openrouter",
  "default_model": "anthropic/claude-opus-4-6",
  "enabled": true
}
```

**Хариу:**
```json
{"status": "updated"}
```

---

#### `DELETE /admin/clients/{id}`

Клиент устгах.

---

### Админ API — Үйлчилгээнүүд

#### `GET /admin/services`

Бүртгэгдсэн үйлчилгээнүүдийн жагсаалт.

**Хариуны жишээ:**
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

Анхдагч 8 үйлчилгээ: `google`, `openai`, `anthropic`, `openrouter`, `github-copilot`, `ollama`, `lmstudio`, `vllm`

---

#### `POST /admin/services`

Тусгай үйлчилгээ нэмэх. Нэмсний дараа SSE `service_changed` үйл явдал цацагдана → **хянах самбарын dropdown шууд шинэчлэгдэнэ**.

**Хүсэлтийн бие:**
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

Үйлчилгээний тохиргоо шинэчлэх. Өөрчлөлтийн дараа SSE `service_changed` үйл явдал цацагдана.

**Хүсэлтийн бие:**
```json
{
  "local_url": "http://192.168.x.x:11434",
  "enabled": true
}
```

---

#### `DELETE /admin/services/{id}`

Тусгай үйлчилгээ устгах. Устгасны дараа SSE `service_changed` үйл явдал цацагдана.

Анхдагч үйлчилгээг (`custom: false`) устгах оролдлого:
```json
{"error": "기본 서비스는 삭제할 수 없습니다: google"}
```

---

### Админ API — Моделийн жагсаалт

#### `GET /admin/models`

Үйлчилгээ тус бүрээр моделийн жагсаалт лавлах. TTL кэш (10 минут) ашиглана.

**Асуулгын параметрүүд:**

| Параметр | Тайлбар | Жишээ |
|----------|---------|-------|
| `service` | Үйлчилгээний шүүлтүүр | `?service=google` |
| `q` | Моделийн хайлт | `?q=gemini` |

**Үйлчилгээ тус бүрийн моделийн лавлагааны арга:**

| Үйлчилгээ | Арга | Тоо |
|-----------|------|-----|
| `google` | Тогтмол жагсаалт | 8 (embedding багтсан) |
| `openai` | Тогтмол жагсаалт | 9 |
| `anthropic` | Тогтмол жагсаалт | 6 |
| `github-copilot` | Тогтмол жагсаалт | 6 |
| `openrouter` | API-ээр динамик лавлагаа (алдаа гарвал сонгосон 14 модель нөөц) | 340+ |
| `ollama` | Локаль серверээс динамик лавлагаа (хариу өгөхгүй бол 7 санал) | Хувьсах |
| `lmstudio` | Локаль серверээс динамик лавлагаа | Хувьсах |
| `vllm` | Локаль серверээс динамик лавлагаа | Хувьсах |
| Тусгай | OpenAI-тай нийцтэй `/v1/models` | Хувьсах |

**OpenRouter нөөц моделийн жагсаалт (API хариу өгөхгүй үед):**

| Модель | Тэмдэглэл |
|--------|-----------|
| `openrouter/hunter-alpha` | Үнэгүй, 1M context |
| `openrouter/healer-alpha` | Үнэгүй, omni-modal |
| `moonshot/kimi-k2.5` | 256K context |
| `z-ai/glm-5`, `z-ai/glm-4.7-flash` | — |
| `deepseek/deepseek-r1`, `deepseek/deepseek-chat` | — |
| `qwen/qwen-2.5-72b-instruct` | 131K context |
| `minimax/minimax-m2.5` | — |
| `meta-llama/llama-3.3-70b-instruct` | 131K context |

---

### Админ API — Проксийн төлөв

#### `GET /admin/proxies`

Холбогдсон бүх проксигийн сүүлийн Heartbeat төлөв.

---

## SSE үйл явдлын төрлүүд

Сейфийн `/api/events` стримээс хүлээн авах үйл явдлууд:

| `type` | Үүсэх нөхцөл | `data` агуулга | Хянах самбарын хариу үйлдэл |
|--------|-------------|---------------|---------------------------|
| `connected` | SSE холбогдсон даруйд | `{"clients": N}` | — |
| `config_change` | Клиентийн тохиргоо өөрчлөгдсөн | `{"client_id","service","model"}` | Агент картын модель dropdown шинэчлэлт |
| `key_added` | Шинэ API түлхүүр бүртгэгдсэн | `{"service": "google"}` | Модель dropdown шинэчлэлт |
| `key_deleted` | API түлхүүр устгагдсан | `{"service": "google"}` | Модель dropdown шинэчлэлт |
| `service_changed` | Үйлчилгээ нэмэгдсэн/өөрчлөгдсөн/устгагдсан | `{"action":"added"\|"updated"\|"deleted","id":"...","proxy_services":["google","ollama"]}` | Үйлчилгээний select + модель dropdown шууд шинэчлэлт; проксийн dispatch үйлчилгээний жагсаалт бодит цагийн шинэчлэлт |
| `usage_update` | Проксийн heartbeat хүлээн авах үед (20 секунд тутам) | `{"keys":[{"id","service","today_usage","today_attempts","daily_limit","cooldown_until"},...]}` | Түлхүүрийн хэрэглээний мөр ба тоонуудын шууд шинэчлэлт, хөргөлтийн тоолол эхлэнэ. fetch-гүйгээр SSE өгөгдлийг шууд ашиглана. Мөрүүд share-of-total масштаблалт ашигладаг (хязгааргүй түлхүүрүүд). |
| `usage_reset` | Өдрийн хэрэглээ эхний байдалд оруулсан | `{"time": "RFC3339"}` | Хуудас шинэчлэх |

**Прокси хүлээн авах үйл явдлуудын боловсруулалт:**

```
config_change хүлээн авсан
  → client_id нь өөрийнхтэй таарвал
    → service, model шууд шинэчлэнэ
    → hooksMgr.Fire(EventModelChanged)
```

---

## Үйлчилгээ үзүүлэгч ба модель чиглүүлэлт

`/v1/chat/completions`-ийн `model` талбарт `provider/model` форматаар зааж өгвөл автомат чиглүүлэлт хийгдэнэ (OpenClaw 3.11 нийцтэй).

### Угтвар чиглүүлэлтийн дүрмүүд

| Угтвар | Чиглүүлэлтийн зорилго | Жишээ |
|--------|----------------------|-------|
| `google/` | Шууд Google руу | `google/gemini-2.5-pro` |
| `openai/` | Шууд OpenAI руу | `openai/gpt-4o` |
| `anthropic/` | OpenRouter-ээр дамжуулан | `anthropic/claude-opus-4-6` |
| `ollama/` | Шууд Ollama руу | `ollama/qwen3.5:35b` |
| `custom/` | Рекурсив дахин задлан шинжлэх (`custom/` устгаад дахин чиглүүлэх) | `custom/google/gemini-2.5-flash` → Google |
| `openrouter/` | OpenRouter (bare path хадгална) | `openrouter/meta-llama/llama-3.3-70b` |
| `opencode/`, `opencode-go/`, `opencode-zen/` | OpenRouter | `opencode-go/model` |
| `moonshot/`, `kimi-coding/` | OpenRouter (full path хадгална) | `moonshot/kimi-k2.5` |
| `groq/`, `mistral/`, `minimax/`, `cohere/`, `perplexity/` | OpenRouter | `groq/llama-3.1-70b` |
| `together/`, `huggingface/`, `nvidia/`, `venice/` | OpenRouter | — |
| `meta-llama/`, `qwen/`, `deepseek/`, `01-ai/` | OpenRouter (full path) | `deepseek/deepseek-r1` |

### `wall-vault/` угтварын автомат илрүүлэлт

wall-vault-ийн өөрийн угтвар нь моделийн ID-аас үйлчилгээг автоматаар тодорхойлно.

| Моделийн ID хэв маяг | Чиглүүлэлт |
|---------------------|------------|
| `wall-vault/gemini-*` | Google |
| `wall-vault/gpt-*`, `wall-vault/o1`, `wall-vault/o3`, `wall-vault/o4*` | OpenAI |
| `wall-vault/claude-*` | OpenRouter (Anthropic зам) |
| `wall-vault/hunter-alpha`, `wall-vault/healer-alpha` | OpenRouter (үнэгүй 1M ctx) |
| `wall-vault/kimi-*`, `wall-vault/glm-*`, `wall-vault/deepseek-*`, `wall-vault/qwen*` | OpenRouter |
| Бусад | OpenRouter |

### `:cloud` дагаварын боловсруулалт

Ollama тэг форматын `:cloud` дагавар автоматаар устгагдаж OpenRouter руу чиглүүлэгдэнэ.

```
kimi-k2.5:cloud  →  OpenRouter, моделийн ID: kimi-k2.5
glm-5:cloud      →  OpenRouter, моделийн ID: glm-5
```

### OpenClaw openclaw.json холболтын жишээ

```json5
// ~/.openclaw/openclaw.json
{
  models: {
    providers: {
      "wall-vault": {
        baseUrl: "http://localhost:56244/v1",
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

Агент картын **🐾 товчлуур** дээр дарснаар тухайн агентын тохиргооны хэсэг санах ойд автоматаар хуулагдана.

---

## Өгөгдлийн схем

### APIKey

| Талбар | Төрөл | Тайлбар |
|--------|-------|---------|
| `id` | string | UUID форматтай өвөрмөц ID |
| `service` | string | `google` \| `openai` \| `anthropic` \| `openrouter` \| `ollama` \| тусгай |
| `encrypted_key` | string | AES-GCM-ээр шифрлэгдсэн түлхүүр (Base64) |
| `label` | string | Тодорхойлох шошго |
| `today_usage` | int | Өнөөдрийн амжилттай хүсэлтийн токенуудын тоо (429/402/582 алдаа тооцохгүй) |
| `today_attempts` | int | Өнөөдрийн нийт API дуудлагын тоо (амжилттай + хурд хязгаарлалт; шөнө дунд цэвэрлэгдэнэ) |
| `daily_limit` | int | Өдрийн хязгаар (0 = хязгааргүй) |
| `cooldown_until` | time.Time | Хөргөлтийн дуусах хугацаа |
| `last_error` | int | Сүүлийн HTTP алдааны код |
| `created_at` | time.Time | Бүртгэгдсэн цаг |

**Хөргөлтийн бодлого:**

| HTTP алдаа | Хөргөлт |
|-----------|---------|
| 429 (Too Many Requests) | 30 минут |
| 402 (Payment Required) | 24 цаг |
| 400 / 401 / 403 | 24 цаг |
| 582 (Gateway Overload) | 5 минут |
| Сүлжээний алдаа | 10 минут |

> **429·402·582**: Хөргөлт тохируулах + `today_attempts` нэмэгдэнэ. `today_usage` өөрчлөгдөхгүй (зөвхөн амжилттай токенуудыг тоолно).
> **Ollama (локаль үйлчилгээ)**: `callOllama` нь `Timeout: 0` (хязгааргүй) тусгай HTTP клиент ашигладаг. Том моделийн дүгнэлт хэдэн арваас хэдэн минут үргэлжилж болохоор анхдагч 60 секундын хугацааны хязгаар хэрэглэгдэхгүй.

### Client

| Талбар | Төрөл | Тайлбар |
|--------|-------|---------|
| `id` | string | Клиентийн өвөрмөц ID |
| `name` | string | Харуулах нэр |
| `token` | string | Баталгаажуулалтын токен |
| `default_service` | string | Анхдагч үйлчилгээ |
| `default_model` | string | Анхдагч модель (`provider/model` формат байж болно) |
| `allowed_services` | []string | Зөвшөөрөгдсөн үйлчилгээнүүд (хоосон массив = бүгд) |
| `agent_type` | string | `openclaw` \| `claude-code` \| `cursor` \| `vscode` \| `custom` |
| `work_dir` | string | Агентын ажлын директор |
| `description` | string | Тайлбар |
| `ip_whitelist` | []string | Зөвшөөрөгдсөн IP жагсаалт (CIDR дэмжинэ) |
| `avatar` | string | Agent avatar — relative path under `~/.openclaw/` (e.g. `workspace/avatar.png`, `workspace/avatars/profile.hpg`) OR base64 data URI (`data:image/png;base64,...`). Supported extensions: `.png`, `.jpg`/`.jpeg`/`.hpg`, `.webp`, `.gif`. |
| `enabled` | bool | `false` бол `/api/keys`-д хандахад `403` буцаана |
| `created_at` | time.Time | Бүртгэгдсэн цаг |

### ServiceConfig

| Талбар | Төрөл | Тайлбар |
|--------|-------|---------|
| `id` | string | Үйлчилгээний өвөрмөц ID |
| `name` | string | Харуулах нэр |
| `local_url` | string | Локаль серверийн URL (Ollama/LMStudio/vLLM/тусгай) |
| `enabled` | bool | Идэвхтэй эсэх |
| `custom` | bool | Хэрэглэгчийн нэмсэн үйлчилгээ эсэх |
| `proxy_enabled` | bool | Whether this service is included in proxy dispatch and agent model dropdowns. Controlled by the "프록시 사용" checkbox in the Services card. Only `proxy_enabled: true` services appear in agent service/model selectors. |

### ProxyStatus (Heartbeat)

| Талбар | Төрөл | Тайлбар |
|--------|-------|---------|
| `client_id` | string | Клиентийн ID |
| `version` | string | Проксийн хувилбар (жишээ: `v0.1.6.20260314.231308`) |
| `service` | string | Одоогийн үйлчилгээ |
| `model` | string | Одоогийн модель |
| `sse_connected` | bool | SSE холболтын төлөв |
| `host` | string | Хостын нэр |
| `avatar` | string | base64 data URI of local avatar (auto-synced from proxy via heartbeat) |
| `updated_at` | time.Time | Сүүлийн шинэчлэлт |
| `vault.today_usage` | int | Өнөөдрийн токен хэрэглээ |
| `vault.daily_limit` | int | Өдрийн хязгаар |
| `vault.key_status` | string | `active` \| `cooldown` \| `exhausted` |

---

## Алдааны хариу

```json
{"error": "Алдааны мессеж"}
```

| Код | Утга |
|-----|------|
| 200 | Амжилттай |
| 400 | Буруу хүсэлт |
| 401 | Баталгаажуулалт амжилтгүй |
| 403 | Хандалт хориглогдсон (идэвхгүй клиент, IP хориглолт) |
| 404 | Нөөц олдсонгүй |
| 405 | Зөвшөөрөгдөөгүй арга |
| 429 | Хурд хязгаарлалт давсан |
| 500 | Серверийн дотоод алдаа |
| 502 | Дээд API алдаа (бүх нөөц амжилтгүй) |

---

## cURL жишээний цуглуулга

```bash
# ─── Прокси ───────────────────────────────────────────────────────────────────

# Эрүүл мэндийн шалгалт
curl http://localhost:56244/health

# Төлөв лавлах
curl http://localhost:56244/status

# Моделийн жагсаалт (бүгд)
curl http://localhost:56244/api/models

# Зөвхөн Google модель
curl "http://localhost:56244/api/models?service=google"

# Үнэгүй модель хайх
curl "http://localhost:56244/api/models?q=alpha"

# Модель солих (локаль)
curl -X PUT http://localhost:56244/api/config/model \
  -H "Content-Type: application/json" \
  -d '{"service":"google","model":"gemini-2.5-pro"}'

# Тохиргоо шинэчлэх
curl -X POST http://localhost:56244/reload

# Gemini API шууд дуудлага
curl -X POST "http://localhost:56244/google/v1beta/models/gemini-2.5-flash:generateContent" \
  -H "Content-Type: application/json" \
  -d '{"contents":[{"role":"user","parts":[{"text":"안녕하세요"}]}]}'

# OpenAI нийцтэй (анхдагч модель)
curl -X POST http://localhost:56244/v1/chat/completions \
  -H "Content-Type: application/json" \
  -d '{"model":"gemini-2.5-flash","messages":[{"role":"user","content":"안녕"}]}'

# OpenClaw provider/model формат
curl -X POST http://localhost:56244/v1/chat/completions \
  -H "Authorization: Bearer my-agent-token" \
  -H "Content-Type: application/json" \
  -d '{"model":"wall-vault/gemini-2.5-flash","messages":[{"role":"user","content":"안녕"}]}'

# Үнэгүй 1M context модель ашиглах
curl -X POST http://localhost:56244/v1/chat/completions \
  -H "Authorization: Bearer my-agent-token" \
  -H "Content-Type: application/json" \
  -d '{"model":"wall-vault/hunter-alpha","messages":[{"role":"user","content":"안녕"}]}'

# ─── Түлхүүрийн сейф (нээлттэй) ─────────────────────────────────────────────

curl http://localhost:56243/api/status
curl http://localhost:56243/api/clients
curl -s http://localhost:56243/api/events --max-time 3

# ─── Түлхүүрийн сейф (админ) ─────────────────────────────────────────────────

ADMIN="Authorization: Bearer admin-token"

# Түлхүүрүүдийн жагсаалт
curl -H "$ADMIN" http://localhost:56243/admin/keys

# Google түлхүүр нэмэх
curl -X POST http://localhost:56243/admin/keys \
  -H "$ADMIN" -H "Content-Type: application/json" \
  -d '{"service":"google","key":"AIzaSy...","label":"메인 키","daily_limit":1000}'

# OpenAI түлхүүр нэмэх
curl -X POST http://localhost:56243/admin/keys \
  -H "$ADMIN" -H "Content-Type: application/json" \
  -d '{"service":"openai","key":"sk-...","label":"GPT 키"}'

# OpenRouter түлхүүр нэмэх
curl -X POST http://localhost:56243/admin/keys \
  -H "$ADMIN" -H "Content-Type: application/json" \
  -d '{"service":"openrouter","key":"sk-or-v1-...","label":"OR 키"}'

# Түлхүүр устгах (SSE key_deleted цацалт)
curl -X DELETE http://localhost:56243/admin/keys/key-abc123 -H "$ADMIN"

# Өдрийн хэрэглээ цэвэрлэх
curl -X POST http://localhost:56243/admin/keys/reset -H "$ADMIN"

# Клиентүүдийн жагсаалт
curl -H "$ADMIN" http://localhost:56243/admin/clients

# Клиент нэмэх (OpenClaw)
curl -X POST http://localhost:56243/admin/clients \
  -H "$ADMIN" -H "Content-Type: application/json" \
  -d '{"id":"bot-a","name":"봇 A","agent_type":"openclaw","work_dir":"~/.openclaw","default_service":"google","default_model":"gemini-2.5-flash"}'

# Клиентийн модель солих (SSE-ээр шууд тусгагдана)
curl -X PUT http://localhost:56243/admin/clients/bot-a \
  -H "$ADMIN" -H "Content-Type: application/json" \
  -d '{"default_service":"openrouter","default_model":"wall-vault/hunter-alpha"}'

# Клиент идэвхгүй болгох
curl -X PUT http://localhost:56243/admin/clients/my-bot \
  -H "$ADMIN" -H "Content-Type: application/json" \
  -d '{"enabled":false}'

# Клиент устгах
curl -X DELETE http://localhost:56243/admin/clients/my-bot -H "$ADMIN"

# Үйлчилгээнүүдийн жагсаалт
curl -H "$ADMIN" http://localhost:56243/admin/services

# Ollama локаль URL тохируулах (SSE service_changed цацалт)
curl -X PUT http://localhost:56243/admin/services/ollama \
  -H "$ADMIN" -H "Content-Type: application/json" \
  -d '{"local_url":"http://192.168.x.x:11434","enabled":true}'

# OpenAI үйлчилгээ идэвхжүүлэх
curl -X PUT http://localhost:56243/admin/services/openai \
  -H "$ADMIN" -H "Content-Type: application/json" \
  -d '{"enabled":true}'

# Тусгай үйлчилгээ нэмэх (SSE service_changed цацалт)
curl -X POST http://localhost:56243/admin/services \
  -H "$ADMIN" -H "Content-Type: application/json" \
  -d '{"id":"my-llm","name":"사내 LLM","local_url":"http://10.0.0.50:8080","enabled":true}'

# Тусгай үйлчилгээ устгах
curl -X DELETE http://localhost:56243/admin/services/my-llm -H "$ADMIN"

# Моделийн жагсаалт лавлах
curl -H "$ADMIN" http://localhost:56243/admin/models
curl -H "$ADMIN" "http://localhost:56243/admin/models?service=openrouter"
curl -H "$ADMIN" "http://localhost:56243/admin/models?q=hunter"

# Проксийн төлөв (heartbeat)
curl -H "$ADMIN" http://localhost:56243/admin/proxies

# ─── Тархсан горим — Прокси → Сейф ───────────────────────────────────────────

# Нууцлал тайлсан түлхүүрүүд лавлах
curl http://localhost:56243/api/keys \
  -H "Authorization: Bearer your-bot-a-token"

# Heartbeat илгээх
curl -X POST http://localhost:56243/api/heartbeat \
  -H "Authorization: Bearer your-bot-a-token" \
  -H "Content-Type: application/json" \
  -d '{"client_id":"bot-a","version":"v0.1.6.20260314.231308","service":"google","model":"gemini-2.5-flash","sse_connected":true}'
```

---

## Зуучлагч програм хангамж

Бүх хүсэлтэд автоматаар хэрэглэгдэнэ:

| Зуучлагч | Үүрэг |
|----------|-------|
| **Logger** | `[method] path status latencyms` форматаар лог бичих |
| **CORS** | `Access-Control-Allow-Origin: *` |
| **Recovery** | Panic-ээс сэргэх, 500 хариу буцаах |

---

*Сүүлд шинэчлэгдсэн: 2026-03-16 — v0.1.7: dashboard title renamed to "벽금고(wall-vault) 대시보드", logo moved to sticky topbar, today_attempts tracking, HTTP 582 cooldown (5min), share-of-total bar scaling, custom/ routing fix, Ollama Timeout:0, key_att i18n, avatar heartbeat sync, build timestamp versioning, proxy-only service filter*
