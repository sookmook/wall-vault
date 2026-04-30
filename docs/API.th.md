# คู่มือ API ของ wall-vault

เอกสารนี้อธิบาย HTTP API endpoint ทั้งหมดของ wall-vault อย่างละเอียด

---

## สารบัญ

- [การยืนยันตัวตน](#การยืนยันตัวตน)
- [Proxy API (:56244)](#proxy-api-56244)
  - [ตรวจสอบสุขภาพ](#get-health)
  - [สอบถามสถานะ](#get-status)
  - [รายการโมเดล](#get-apimodels)
  - [เปลี่ยนโมเดล](#put-apiconfigmodel)
  - [โหมดคิด](#put-apiconfigthink-mode)
  - [รีโหลดการตั้งค่า](#post-reload)
  - [Gemini API](#post-googlev1betamodelsmgeneratecontent)
  - [Gemini สตรีมมิ่ง](#post-googlev1betamodelsmstreamgeneratecontent)
  - [OpenAI Compatible API](#post-v1chatcompletions)
- [Key Vault API (:56243)](#key-vault-api-56243)
  - [API สาธารณะ](#api-สาธารณะไมตองยืนยันตัวตน)
  - [SSE Event Stream](#get-apievents)
  - [API เฉพาะ Proxy](#api-เฉพาะ-proxyโทเค็นไคลเอ็นต)
  - [Admin API — คีย์](#admin-api--api-คีย)
  - [Admin API — ไคลเอ็นต์](#admin-api--ไคลเอนต)
  - [Admin API — เซอร์วิส](#admin-api--เซอรวิส)
  - [Admin API — รายการโมเดล](#admin-api--รายการโมเดล)
  - [Admin API — สถานะ Proxy](#admin-api--สถานะ-proxy)
- [ประเภทเหตุการณ์ SSE](#ประเภทเหตุการณ-sse)
- [การเราท์ Provider·Model](#การเราทprovidermodel)
- [Data Schema](#data-schema)
- [การตอบกลับข้อผิดพลาด](#การตอบกลับขอผิดพลาด)
- [ตัวอย่าง cURL](#ตัวอยาง-curl)

---

## การยืนยันตัวตน

| ขอบเขต | วิธีการ | Header |
|--------|---------|--------|
| Admin API | Bearer Token | `Authorization: Bearer <admin_token>` |
| Proxy → Vault | Bearer Token | `Authorization: Bearer <client_token>` |
| Proxy API | ไม่มี (โลคอล) | — |

หาก `admin_token` ไม่ได้ตั้งค่า (สตริงว่าง) Admin API ทั้งหมดจะเข้าถึงได้โดยไม่ต้องยืนยันตัวตน

### นโยบายความปลอดภัย

- **Rate Limiting**: เมื่อการยืนยันตัวตน Admin API ล้มเหลวมากกว่า 10 ครั้ง/15 นาที IP นั้นจะถูกบล็อกชั่วคราว (`429 Too Many Requests`)
- **IP Whitelist**: เฉพาะ IP/CIDR ที่ลงทะเบียนในฟิลด์ `ip_whitelist` ของเอเจนต์ (`Client`) เท่านั้นที่สามารถเข้าถึง `/api/keys` ได้ อาร์เรย์ว่างหมายถึงอนุญาตทั้งหมด
- **การป้องกัน theme·lang**: `/admin/theme`, `/admin/lang` ก็ต้องยืนยันตัวตนด้วยโทเค็นแอดมินเช่นกัน

---

## Proxy API (:56244)

เซิร์ฟเวอร์ที่ Proxy ทำงาน พอร์ตเริ่มต้น `56244`

---

### `GET /health`

ตรวจสอบสุขภาพ ส่งกลับ 200 OK เสมอ

**ตัวอย่างการตอบกลับ:**
```json
{
  "status": "ok",
  "version": "v0.1.6.20260314.231308",
  "client": "bot-a"
}
```

---

### `GET /status`

สอบถามสถานะ Proxy โดยละเอียด

**ตัวอย่างการตอบกลับ:**
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

| ฟิลด์ | ประเภท | คำอธิบาย |
|-------|--------|----------|
| `service` | string | เซอร์วิสเริ่มต้นปัจจุบัน |
| `model` | string | โมเดลเริ่มต้นปัจจุบัน |
| `sse` | bool | สถานะการเชื่อมต่อ SSE ของ Vault |
| `filter` | string | โหมดตัวกรองเครื่องมือ |
| `services` | []string | รายการเซอร์วิสที่เปิดใช้งาน |
| `mode` | string | `standalone` \| `distributed` |

---

### `GET /api/models`

สอบถามรายการโมเดลที่ใช้งานได้ ใช้ TTL cache (ค่าเริ่มต้น 10 นาที)

**Query Parameters:**

| พารามิเตอร์ | คำอธิบาย | ตัวอย่าง |
|------------|----------|---------|
| `service` | ตัวกรองเซอร์วิส | `?service=google` |
| `q` | ค้นหา ID/ชื่อโมเดล | `?q=gemini` |

**ตัวอย่างการตอบกลับ:**
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

| ฟิลด์ | ประเภท | คำอธิบาย |
|-------|--------|----------|
| `id` | string | ID โมเดล |
| `name` | string | ชื่อแสดงผลโมเดล |
| `service` | string | `google` \| `openai` \| `anthropic` \| `openrouter` \| `ollama` เป็นต้น |
| `context_length` | int | ขนาด context window |
| `free` | bool | เป็นโมเดลฟรีหรือไม่ (OpenRouter) |

---

### `PUT /api/config/model`

เปลี่ยนเซอร์วิส·โมเดลปัจจุบัน

**Request Body:**
```json
{
  "service": "google",
  "model": "gemini-2.5-flash"
}
```

**การตอบกลับ:**
```json
{
  "status": "ok",
  "service": "google",
  "model": "gemini-2.5-flash"
}
```

> **หมายเหตุ:** ในโหมด distributed แนะนำให้ใช้ `PUT /admin/clients/{id}` ของ Vault แทน API นี้ การเปลี่ยนแปลงใน Vault จะถูกสะท้อนอัตโนมัติผ่าน SSE ภายใน 1-3 วินาที

---

### `PUT /api/config/think-mode`

สลับโหมดคิด (no-op, สำรองสำหรับการขยายในอนาคต)

**การตอบกลับ:**
```json
{"status": "ok"}
```

---

### `POST /reload`

ซิงค์การตั้งค่าไคลเอ็นต์·คีย์จาก Vault ใหม่ทันที

**การตอบกลับ:**
```json
{"status": "reloading"}
```

การซิงค์ใหม่ทำงานแบบ asynchronous จึงเสร็จสิ้นภายใน 1-2 วินาทีหลังได้รับการตอบกลับ

---

### `POST /google/v1beta/models/{model}:generateContent`

Gemini API Proxy (ไม่ใช่สตรีมมิ่ง)

**Path Parameters:**
- `{model}`: ID โมเดล หากมีคำนำหน้า `gemini-` จะเลือกเซอร์วิส Google อัตโนมัติ

**Request Body:** [รูปแบบคำขอ Gemini generateContent](https://ai.google.dev/api/generate-content)

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

**Response Body:** รูปแบบการตอบกลับ Gemini generateContent

**ตัวกรองเครื่องมือ:** เมื่อตั้งค่า `tool_filter: strip_all` อาร์เรย์ `tools` ในคำขอจะถูกลบออกอัตโนมัติ

**Fallback Chain:** เซอร์วิสที่กำหนดล้มเหลว → ฟอลแบ็กตามลำดับเซอร์วิสที่ตั้งค่าไว้ → Ollama (ตัวสุดท้าย)

---

### `POST /google/v1beta/models/{model}:streamGenerateContent`

Gemini API Streaming Proxy รูปแบบคำขอเหมือนกับแบบไม่ใช่สตรีมมิ่ง การตอบกลับเป็น SSE stream:

```
data: {"candidates":[{"content":{"parts":[{"text":"안"}],...},...}]}

data: [DONE]
```

---

### `POST /v1/chat/completions`

OpenAI Compatible API ภายในจะแปลงเป็นรูปแบบ Gemini แล้วประมวลผล

**Request Body:**
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

**การรองรับคำนำหน้า Provider ในฟิลด์ `model` (OpenClaw 3.11+):**

| ตัวอย่างโมเดล | การเราท์ |
|--------------|---------|
| `gemini-2.5-flash` | เซอร์วิสที่ตั้งค่าปัจจุบัน |
| `google/gemini-2.5-pro` | ตรงไปยัง Google |
| `openai/gpt-4o` | ตรงไปยัง OpenAI |
| `anthropic/claude-opus-4-6` | ผ่าน OpenRouter |
| `openrouter/meta-llama/llama-3.3-70b` | ตรงไปยัง OpenRouter |
| `wall-vault/gemini-2.5-flash` | ตรวจจับอัตโนมัติ → Google |
| `wall-vault/claude-opus-4-6` | ตรวจจับอัตโนมัติ → OpenRouter Anthropic |
| `wall-vault/gpt-4o` | ตรวจจับอัตโนมัติ → OpenAI |
| `wall-vault/hunter-alpha` | OpenRouter (ฟรี 1M context) |
| `moonshot/kimi-k2.5` | ผ่าน OpenRouter |
| `opencode-go/model` | ผ่าน OpenRouter |
| `kimi-k2.5:cloud` | ส่วนต่อท้าย `:cloud` → OpenRouter |

ดูรายละเอียดที่ [การเราท์ Provider·Model](#การเราทprovidermodel)

**Response Body:**
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

> **การลบ Model Control Token อัตโนมัติ:** หากการตอบกลับมีตัวคั่น GLM-5 / DeepSeek / ChatML (`<|im_start|>`, `[gMASK]`, `[sop]` เป็นต้น) จะถูกลบออกอัตโนมัติ

---

## Key Vault API (:56243)

เซิร์ฟเวอร์ที่ Key Vault ทำงาน พอร์ตเริ่มต้น `56243`

---

### API สาธารณะ (ไม่ต้องยืนยันตัวตน)

#### `GET /`

Web Dashboard UI เข้าถึงผ่านเบราว์เซอร์

---

#### `GET /api/status`

สอบถามสถานะ Vault

**ตัวอย่างการตอบกลับ:**
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

รายการไคลเอ็นต์ที่ลงทะเบียน (เฉพาะข้อมูลสาธารณะ ไม่รวมโทเค็น)

---

### `GET /api/events`

SSE (Server-Sent Events) สตรีมเหตุการณ์แบบเรียลไทม์

**Headers:**
```
Content-Type: text/event-stream
Cache-Control: no-cache
Connection: keep-alive
```

**รับทันทีเมื่อเชื่อมต่อ:**
```
data: {"type":"connected","clients":2}
```

**ตัวอย่างเหตุการณ์:**
```
data: {"type":"config_change","data":{"client_id":"bot-a","service":"google","model":"gemini-2.5-pro"}}

data: {"type":"key_added","data":{"service":"openrouter"}}

data: {"type":"key_deleted","data":{"service":"google"}}

data: {"type":"service_changed","data":{"action":"updated","id":"ollama"}}

data: {"type":"usage_reset","data":{"time":"2026-03-13T00:00:30Z"}}
```

ดูรายละเอียดประเภทเหตุการณ์ที่ [ประเภทเหตุการณ์ SSE](#ประเภทเหตุการณ-sse)

---

### API เฉพาะ Proxy (โทเค็นไคลเอ็นต์)

ต้องมี header `Authorization: Bearer <client_token>` สามารถยืนยันตัวตนด้วยโทเค็นแอดมินได้เช่นกัน

#### `GET /api/keys`

รายการ API คีย์ที่ถอดรหัสแล้วสำหรับ Proxy

**Query Parameters:**

| พารามิเตอร์ | คำอธิบาย |
|------------|----------|
| `service` | ตัวกรองเซอร์วิส (เช่น: `?service=google`) |

**ตัวอย่างการตอบกลับ:**
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

> **ความปลอดภัย:** ส่งกลับคีย์แบบข้อความธรรมดา เฉพาะคีย์ของเซอร์วิสที่อนุญาตตามการตั้งค่า `allowed_services` ของไคลเอ็นต์เท่านั้นที่จะถูกส่งกลับ

---

#### `GET /api/services`

สอบถามรายการเซอร์วิสที่ Proxy ใช้ ส่งกลับอาร์เรย์ ID ของเซอร์วิสที่มี `proxy_enabled=true`

**ตัวอย่างการตอบกลับ:**
```json
["google", "ollama"]
```

อาร์เรย์ว่างหมายถึง Proxy สามารถใช้เซอร์วิสทั้งหมดโดยไม่จำกัด

---

#### `POST /api/heartbeat`

ส่งสถานะ Proxy (ทำงานอัตโนมัติทุก 20 วินาที)

**Request Body:**
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

| ฟิลด์ | ประเภท | คำอธิบาย |
|-------|--------|----------|
| `client_id` | string | ID ไคลเอ็นต์ |
| `version` | string | เวอร์ชัน Proxy (รวม build timestamp เช่น `v0.1.6.20260314.231308`) |
| `service` | string | เซอร์วิสปัจจุบัน |
| `model` | string | โมเดลปัจจุบัน |
| `sse_connected` | bool | สถานะการเชื่อมต่อ SSE |
| `host` | string | ชื่อโฮสต์ |
| `avatar` | string | base64 data URI of the proxy's local avatar image (sent by proxy, auto-persisted to client record by vault). Set via `WV_AVATAR` env var (relative path under `~/.openclaw/`). |
| `key_usage` | map[string]int | Key ID → successful tokens used today. Vault calls `SetKeyUsage()` (absolute value sync). 429/402/582 errors do **not** increment this. |
| `key_attempts` | map[string]int | Key ID → total API calls today (success + rate-limited). Vault calls `SetKeyAttempts()`. 429/402/582 errors only increment this, not `key_usage`. |
| `key_cooldowns` | map[string]string | Key ID → cooldown end time (RFC3339). Vault calls `SetKeyCooldownIfLater()`. |

**การตอบกลับ:**
```json
{"status": "ok"}
```

---

### Admin API — API คีย์

ต้องมี header `Authorization: Bearer <admin_token>`

#### `GET /admin/keys`

รายการ API คีย์ทั้งหมดที่ลงทะเบียน (ไม่รวมคีย์ข้อความธรรมดา)

**ตัวอย่างการตอบกลับ:**
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

| ฟิลด์ | ประเภท | คำอธิบาย |
|-------|--------|----------|
| `today_usage` | int | จำนวน token คำขอสำเร็จวันนี้ (ไม่รวมข้อผิดพลาด 429/402/582) |
| `today_attempts` | int | จำนวนการเรียก API ทั้งหมดวันนี้ (รวมสำเร็จ + rate-limited) |
| `available` | bool | สามารถใช้งานได้หรือไม่ (ไม่มี cooldown·ไม่ถึงขีดจำกัด) |
| `usage_pct` | int | เปอร์เซ็นต์การใช้งานเทียบกับขีดจำกัดรายวัน (`daily_limit=0` จะเป็น 0) |
| `cooldown_until` | RFC3339 | เวลาสิ้นสุด cooldown (ค่าศูนย์หมายถึงไม่มี) |
| `last_error` | int | โค้ดข้อผิดพลาด HTTP ล่าสุด |

---

#### `POST /admin/keys`

ลงทะเบียน API คีย์ใหม่ ทันทีหลังลงทะเบียนจะกระจายเหตุการณ์ SSE `key_added`

**Request Body:**
```json
{
  "service": "google",
  "key": "AIzaSy...",
  "label": "메인 키",
  "daily_limit": 1000
}
```

| ฟิลด์ | จำเป็น | คำอธิบาย |
|-------|--------|----------|
| `service` | ✅ | `google` \| `openai` \| `anthropic` \| `openrouter` \| `ollama` \| กำหนดเอง |
| `key` | ✅ | API คีย์ข้อความธรรมดา |
| `label` | — | ป้ายกำกับสำหรับระบุ |
| `daily_limit` | — | ขีดจำกัดการใช้งานรายวัน (0 = ไม่จำกัด) |

---

#### `DELETE /admin/keys/{id}`

ลบ API คีย์ หลังลบจะกระจายเหตุการณ์ SSE `key_deleted`

**การตอบกลับ:**
```json
{"status": "deleted"}
```

---

#### `POST /admin/keys/reset`

รีเซ็ตปริมาณการใช้งานรายวันของคีย์ทั้งหมด กระจายเหตุการณ์ SSE `usage_reset`

**การตอบกลับ:**
```json
{
  "status": "reset",
  "time": "2026-03-13T15:00:00Z"
}
```

---

### Admin API — ไคลเอ็นต์

#### `GET /admin/clients`

รายการไคลเอ็นต์ทั้งหมด (รวมโทเค็น)

---

#### `POST /admin/clients`

ลงทะเบียนไคลเอ็นต์ใหม่

**Request Body:**
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

| ฟิลด์ | จำเป็น | คำอธิบาย |
|-------|--------|----------|
| `id` | ✅ | ID เฉพาะของไคลเอ็นต์ |
| `name` | — | ชื่อแสดงผล |
| `token` | — | โทเค็นยืนยันตัวตน (สร้างอัตโนมัติเมื่อไม่ระบุ) |
| `default_service` | — | เซอร์วิสเริ่มต้น |
| `default_model` | — | โมเดลเริ่มต้น |
| `allowed_services` | — | รายการเซอร์วิสที่อนุญาต (อาร์เรย์ว่าง = อนุญาตทั้งหมด) |
| `agent_type` | — | `openclaw` \| `claude-code` \| `cursor` \| `vscode` \| `custom` |
| `work_dir` | — | ไดเรกทอรีทำงานของเอเจนต์ |
| `description` | — | คำอธิบายเอเจนต์ |
| `ip_whitelist` | — | รายการ IP ที่อนุญาต (อาร์เรย์ว่าง = อนุญาตทั้งหมด, รองรับ CIDR) |
| `enabled` | — | เปิดใช้งานหรือไม่ (ค่าเริ่มต้น `true`) |

---

#### `GET /admin/clients/{id}`

สอบถามไคลเอ็นต์เฉพาะ (รวมโทเค็น)

---

#### `PUT /admin/clients/{id}`

เปลี่ยนการตั้งค่าไคลเอ็นต์ **SSE `config_change` กระจาย → สะท้อนไปยัง Proxy ภายใน 1-3 วินาที**

**Request Body (เฉพาะฟิลด์ที่ต้องการเปลี่ยน):**
```json
{
  "default_service": "openrouter",
  "default_model": "anthropic/claude-opus-4-6",
  "enabled": true
}
```

**การตอบกลับ:**
```json
{"status": "updated"}
```

---

#### `DELETE /admin/clients/{id}`

ลบไคลเอ็นต์

---

### Admin API — เซอร์วิส

#### `GET /admin/services`

รายการเซอร์วิสที่ลงทะเบียน

**ตัวอย่างการตอบกลับ:**
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

เซอร์วิสเริ่มต้น 8 รายการ: `google`, `openai`, `anthropic`, `openrouter`, `github-copilot`, `ollama`, `lmstudio`, `vllm`

---

#### `POST /admin/services`

เพิ่มเซอร์วิสกำหนดเอง หลังเพิ่มจะกระจายเหตุการณ์ SSE `service_changed` → **ดรอปดาวน์แดชบอร์ดอัปเดตทันที**

**Request Body:**
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

อัปเดตการตั้งค่าเซอร์วิส หลังเปลี่ยนแปลงจะกระจายเหตุการณ์ SSE `service_changed`

**Request Body:**
```json
{
  "local_url": "http://192.168.x.x:11434",
  "enabled": true
}
```

---

#### `DELETE /admin/services/{id}`

ลบเซอร์วิสกำหนดเอง หลังลบจะกระจายเหตุการณ์ SSE `service_changed`

เมื่อพยายามลบเซอร์วิสเริ่มต้น (`custom: false`):
```json
{"error": "기본 서비스는 삭제할 수 없습니다: google"}
```

---

### Admin API — รายการโมเดล

#### `GET /admin/models`

สอบถามรายการโมเดลตามเซอร์วิส ใช้ TTL cache (10 นาที)

**Query Parameters:**

| พารามิเตอร์ | คำอธิบาย | ตัวอย่าง |
|------------|----------|---------|
| `service` | ตัวกรองเซอร์วิส | `?service=google` |
| `q` | ค้นหาโมเดล | `?q=gemini` |

**วิธีการสอบถามโมเดลตามเซอร์วิส:**

| เซอร์วิส | วิธีการ | จำนวน |
|---------|--------|-------|
| `google` | รายการคงที่ | 8 รายการ (รวม embedding) |
| `openai` | รายการคงที่ | 9 รายการ |
| `anthropic` | รายการคงที่ | 6 รายการ |
| `github-copilot` | รายการคงที่ | 6 รายการ |
| `openrouter` | สอบถามแบบไดนามิกผ่าน API (ฟอลแบ็ก curated 14 รายการเมื่อล้มเหลว) | 340+ รายการ |
| `ollama` | สอบถามแบบไดนามิกจากเซิร์ฟเวอร์โลคอล (แนะนำ 7 รายการเมื่อไม่ตอบกลับ) | ตามจริง |
| `lmstudio` | สอบถามแบบไดนามิกจากเซิร์ฟเวอร์โลคอล | ตามจริง |
| `vllm` | สอบถามแบบไดนามิกจากเซิร์ฟเวอร์โลคอล | ตามจริง |
| กำหนดเอง | OpenAI Compatible `/v1/models` | ตามจริง |

**รายการโมเดลฟอลแบ็ก OpenRouter (เมื่อ API ไม่ตอบกลับ):**

| โมเดล | หมายเหตุ |
|-------|---------|
| `openrouter/hunter-alpha` | ฟรี, 1M context |
| `openrouter/healer-alpha` | ฟรี, omni-modal |
| `moonshot/kimi-k2.5` | 256K context |
| `z-ai/glm-5`, `z-ai/glm-4.7-flash` | — |
| `deepseek/deepseek-r1`, `deepseek/deepseek-chat` | — |
| `qwen/qwen-2.5-72b-instruct` | 131K context |
| `minimax/minimax-m2.5` | — |
| `meta-llama/llama-3.3-70b-instruct` | 131K context |

---

### Admin API — สถานะ Proxy

#### `GET /admin/proxies`

สถานะ Heartbeat ล่าสุดของ Proxy ทั้งหมดที่เชื่อมต่อ

---

## ประเภทเหตุการณ์ SSE

เหตุการณ์ที่รับจากสตรีม `/api/events` ของ Vault:

| `type` | เงื่อนไขการเกิด | เนื้อหา `data` | การตอบสนองของแดชบอร์ด |
|--------|----------------|----------------|----------------------|
| `connected` | ทันทีเมื่อเชื่อมต่อ SSE | `{"clients": N}` | — |
| `config_change` | เปลี่ยนการตั้งค่าไคลเอ็นต์ | `{"client_id","service","model"}` | อัปเดตดรอปดาวน์โมเดลการ์ดเอเจนต์ |
| `key_added` | ลงทะเบียน API คีย์ใหม่ | `{"service": "google"}` | อัปเดตดรอปดาวน์โมเดล |
| `key_deleted` | ลบ API คีย์ | `{"service": "google"}` | อัปเดตดรอปดาวน์โมเดล |
| `service_changed` | เพิ่ม/แก้ไข/ลบเซอร์วิส | `{"action":"added"\|"updated"\|"deleted","id":"...","proxy_services":["google","ollama"]}` | อัปเดต service select + ดรอปดาวน์โมเดลทันที; อัปเดตรายการเซอร์วิส dispatch ของ Proxy แบบเรียลไทม์ |
| `usage_update` | เมื่อรับ Proxy heartbeat (ทุก 20 วินาที) | `{"keys":[{"id","service","today_usage","today_attempts","daily_limit","cooldown_until"},...]}` | อัปเดตแถบ·ตัวเลขการใช้งานคีย์ทันที, เริ่มนับถอยหลัง cooldown ใช้ข้อมูล SSE โดยตรงโดยไม่ต้อง fetch แถบใช้ share-of-total scaling (คีย์ไม่จำกัด) |
| `usage_reset` | รีเซ็ตปริมาณการใช้งานรายวัน | `{"time": "RFC3339"}` | รีโหลดหน้า |

**การประมวลผลเหตุการณ์ที่ Proxy รับ:**

```
config_change รับ
  → เมื่อ client_id ตรงกับตัวเอง
    → อัปเดต service, model ทันที
    → hooksMgr.Fire(EventModelChanged)
```

---

## การเราท์ Provider·Model

การระบุรูปแบบ `provider/model` ในฟิลด์ `model` ของ `/v1/chat/completions` จะเราท์อัตโนมัติ (เข้ากันได้กับ OpenClaw 3.11)

### กฎการเราท์คำนำหน้า

| คำนำหน้า | เป้าหมายการเราท์ | ตัวอย่าง |
|---------|-----------------|---------|
| `google/` | ตรงไปยัง Google | `google/gemini-2.5-pro` |
| `openai/` | ตรงไปยัง OpenAI | `openai/gpt-4o` |
| `anthropic/` | ผ่าน OpenRouter | `anthropic/claude-opus-4-6` |
| `ollama/` | ตรงไปยัง Ollama | `ollama/qwen3.5:35b` |
| `custom/` | แยกวิเคราะห์ซ้ำ (ลบ `custom/` แล้วเราท์ใหม่) | `custom/google/gemini-2.5-flash` → Google |
| `openrouter/` | OpenRouter (คง bare path) | `openrouter/meta-llama/llama-3.3-70b` |
| `opencode/`, `opencode-go/`, `opencode-zen/` | OpenRouter | `opencode-go/model` |
| `moonshot/`, `kimi-coding/` | OpenRouter (คง full path) | `moonshot/kimi-k2.5` |
| `groq/`, `mistral/`, `minimax/`, `cohere/`, `perplexity/` | OpenRouter | `groq/llama-3.1-70b` |
| `together/`, `huggingface/`, `nvidia/`, `venice/` | OpenRouter | — |
| `meta-llama/`, `qwen/`, `deepseek/`, `01-ai/` | OpenRouter (full path) | `deepseek/deepseek-r1` |

### การตรวจจับอัตโนมัติคำนำหน้า `wall-vault/`

คำนำหน้าของ wall-vault เอง ตรวจจับเซอร์วิสจาก model ID อัตโนมัติ

| รูปแบบ Model ID | การเราท์ |
|----------------|---------|
| `wall-vault/gemini-*` | Google |
| `wall-vault/gpt-*`, `wall-vault/o1`, `wall-vault/o3`, `wall-vault/o4*` | OpenAI |
| `wall-vault/claude-*` | OpenRouter (เส้นทาง Anthropic) |
| `wall-vault/hunter-alpha`, `wall-vault/healer-alpha` | OpenRouter (ฟรี 1M ctx) |
| `wall-vault/kimi-*`, `wall-vault/glm-*`, `wall-vault/deepseek-*`, `wall-vault/qwen*` | OpenRouter |
| อื่นๆ | OpenRouter |

### การจัดการส่วนต่อท้าย `:cloud`

ส่วนต่อท้าย `:cloud` ในรูปแบบ Ollama tag จะถูกลบออกอัตโนมัติแล้วเราท์ไปยัง OpenRouter

```
kimi-k2.5:cloud  →  OpenRouter, Model ID: kimi-k2.5
glm-5:cloud      →  OpenRouter, Model ID: glm-5
```

### ตัวอย่างการเชื่อมต่อ OpenClaw openclaw.json

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

คลิก **ปุ่ม 🐾** บนการ์ดเอเจนต์เพื่อคัดลอกสนิปเพ็ตการตั้งค่าสำหรับเอเจนต์นั้นไปยังคลิปบอร์ดอัตโนมัติ

---

## Data Schema

### APIKey

| ฟิลด์ | ประเภท | คำอธิบาย |
|-------|--------|----------|
| `id` | string | ID เฉพาะรูปแบบ UUID |
| `service` | string | `google` \| `openai` \| `anthropic` \| `openrouter` \| `ollama` \| กำหนดเอง |
| `encrypted_key` | string | คีย์เข้ารหัส AES-GCM (Base64) |
| `label` | string | ป้ายกำกับสำหรับระบุ |
| `today_usage` | int | จำนวน token คำขอสำเร็จวันนี้ (ไม่รวมข้อผิดพลาด 429/402/582) |
| `today_attempts` | int | จำนวนการเรียก API ทั้งหมดวันนี้ (สำเร็จ + rate-limited; รีเซ็ตตอนเที่ยงคืน) |
| `daily_limit` | int | ขีดจำกัดรายวัน (0 = ไม่จำกัด) |
| `cooldown_until` | time.Time | เวลาสิ้นสุด cooldown |
| `last_error` | int | โค้ดข้อผิดพลาด HTTP ล่าสุด |
| `created_at` | time.Time | เวลาลงทะเบียน |

**นโยบาย Cooldown:**

| ข้อผิดพลาด HTTP | Cooldown |
|----------------|----------|
| 429 (Too Many Requests) | 30 นาที |
| 402 (Payment Required) | 24 ชั่วโมง |
| 400 / 401 / 403 | 24 ชั่วโมง |
| 582 (Gateway Overload) | 5 นาที |
| ข้อผิดพลาดเครือข่าย | 10 นาที |

> **429·402·582**: ตั้ง cooldown + `today_attempts` เพิ่ม `today_usage` ไม่เปลี่ยนแปลง (นับเฉพาะ token สำเร็จ)
> **Ollama (เซอร์วิสโลคอล)**: `callOllama` ใช้ HTTP client เฉพาะที่มี `Timeout: 0` (ไม่จำกัด) การอนุมานโมเดลขนาดใหญ่อาจใช้เวลาหลายสิบวินาทีถึงหลายนาที จึงไม่ใช้ timeout เริ่มต้น 60 วินาที

### Client

| ฟิลด์ | ประเภท | คำอธิบาย |
|-------|--------|----------|
| `id` | string | ID เฉพาะไคลเอ็นต์ |
| `name` | string | ชื่อแสดงผล |
| `token` | string | โทเค็นยืนยันตัวตน |
| `default_service` | string | เซอร์วิสเริ่มต้น |
| `default_model` | string | โมเดลเริ่มต้น (สามารถใช้รูปแบบ `provider/model`) |
| `allowed_services` | []string | เซอร์วิสที่อนุญาต (อาร์เรย์ว่าง = ทั้งหมด) |
| `agent_type` | string | `openclaw` \| `claude-code` \| `cursor` \| `vscode` \| `custom` |
| `work_dir` | string | ไดเรกทอรีทำงานของเอเจนต์ |
| `description` | string | คำอธิบาย |
| `ip_whitelist` | []string | รายการ IP ที่อนุญาต (รองรับ CIDR) |
| `avatar` | string | Agent avatar — relative path under `~/.openclaw/` (e.g. `workspace/avatar.png`, `workspace/avatars/profile.hpg`) OR base64 data URI (`data:image/png;base64,...`). Supported extensions: `.png`, `.jpg`/`.jpeg`/`.hpg`, `.webp`, `.gif`. |
| `enabled` | bool | เมื่อเป็น `false` จะส่งกลับ `403` เมื่อเข้าถึง `/api/keys` |
| `created_at` | time.Time | เวลาลงทะเบียน |

### ServiceConfig

| ฟิลด์ | ประเภท | คำอธิบาย |
|-------|--------|----------|
| `id` | string | ID เฉพาะเซอร์วิส |
| `name` | string | ชื่อแสดงผล |
| `local_url` | string | URL เซิร์ฟเวอร์โลคอล (Ollama/LMStudio/vLLM/กำหนดเอง) |
| `enabled` | bool | เปิดใช้งานหรือไม่ |
| `custom` | bool | เป็นเซอร์วิสที่ผู้ใช้เพิ่มหรือไม่ |
| `proxy_enabled` | bool | Whether this service is included in proxy dispatch and agent model dropdowns. Controlled by the "프록시 사용" checkbox in the Services card. Only `proxy_enabled: true` services appear in agent service/model selectors. |

### ProxyStatus (Heartbeat)

| ฟิลด์ | ประเภท | คำอธิบาย |
|-------|--------|----------|
| `client_id` | string | ID ไคลเอ็นต์ |
| `version` | string | เวอร์ชัน Proxy (เช่น `v0.1.6.20260314.231308`) |
| `service` | string | เซอร์วิสปัจจุบัน |
| `model` | string | โมเดลปัจจุบัน |
| `sse_connected` | bool | สถานะการเชื่อมต่อ SSE |
| `host` | string | ชื่อโฮสต์ |
| `avatar` | string | base64 data URI of local avatar (auto-synced from proxy via heartbeat) |
| `updated_at` | time.Time | เวลาอัปเดตล่าสุด |
| `vault.today_usage` | int | ปริมาณ token ที่ใช้วันนี้ |
| `vault.daily_limit` | int | ขีดจำกัดรายวัน |
| `vault.key_status` | string | `active` \| `cooldown` \| `exhausted` |

---

## การตอบกลับข้อผิดพลาด

```json
{"error": "오류 메시지"}
```

| โค้ด | ความหมาย |
|------|----------|
| 200 | สำเร็จ |
| 400 | คำขอไม่ถูกต้อง |
| 401 | ยืนยันตัวตนล้มเหลว |
| 403 | ปฏิเสธการเข้าถึง (ไคลเอ็นต์ไม่ได้ใช้งาน, IP ถูกบล็อก) |
| 404 | ไม่พบทรัพยากร |
| 405 | เมธอดไม่ได้รับอนุญาต |
| 429 | เกินขีดจำกัด Rate limit |
| 500 | ข้อผิดพลาดภายในเซิร์ฟเวอร์ |
| 502 | ข้อผิดพลาด Upstream API (ฟอลแบ็กทั้งหมดล้มเหลว) |

---

## ตัวอย่าง cURL

```bash
# ─── Proxy ────────────────────────────────────────────────────────────────────

# ตรวจสอบสุขภาพ
curl https://localhost:56244/health

# สอบถามสถานะ
curl https://localhost:56244/status

# รายการโมเดล (ทั้งหมด)
curl https://localhost:56244/api/models

# เฉพาะโมเดล Google
curl "https://localhost:56244/api/models?service=google"

# ค้นหาโมเดลฟรี
curl "https://localhost:56244/api/models?q=alpha"

# เปลี่ยนโมเดล (โลคอล)
curl -X PUT https://localhost:56244/api/config/model \
  -H "Content-Type: application/json" \
  -d '{"service":"google","model":"gemini-2.5-pro"}'

# รีโหลดการตั้งค่า
curl -X POST https://localhost:56244/reload

# เรียก Gemini API โดยตรง
curl -X POST "https://localhost:56244/google/v1beta/models/gemini-2.5-flash:generateContent" \
  -H "Content-Type: application/json" \
  -d '{"contents":[{"role":"user","parts":[{"text":"안녕하세요"}]}]}'

# OpenAI Compatible (โมเดลเริ่มต้น)
curl -X POST https://localhost:56244/v1/chat/completions \
  -H "Content-Type: application/json" \
  -d '{"model":"gemini-2.5-flash","messages":[{"role":"user","content":"안녕"}]}'

# รูปแบบ OpenClaw provider/model
curl -X POST https://localhost:56244/v1/chat/completions \
  -H "Authorization: Bearer my-agent-token" \
  -H "Content-Type: application/json" \
  -d '{"model":"wall-vault/gemini-2.5-flash","messages":[{"role":"user","content":"안녕"}]}'

# ใช้โมเดล 1M context ฟรี
curl -X POST https://localhost:56244/v1/chat/completions \
  -H "Authorization: Bearer my-agent-token" \
  -H "Content-Type: application/json" \
  -d '{"model":"wall-vault/hunter-alpha","messages":[{"role":"user","content":"안녕"}]}'

# ─── Key Vault (สาธารณะ) ──────────────────────────────────────────────────────

curl https://localhost:56243/api/status
curl https://localhost:56243/api/clients
curl -s https://localhost:56243/api/events --max-time 3

# ─── Key Vault (แอดมิน) ────────────────────────────────────────────────────────

ADMIN="Authorization: Bearer admin-token"

# รายการคีย์
curl -H "$ADMIN" https://localhost:56243/admin/keys

# เพิ่มคีย์ Google
curl -X POST https://localhost:56243/admin/keys \
  -H "$ADMIN" -H "Content-Type: application/json" \
  -d '{"service":"google","key":"AIzaSy...","label":"메인 키","daily_limit":1000}'

# เพิ่มคีย์ OpenAI
curl -X POST https://localhost:56243/admin/keys \
  -H "$ADMIN" -H "Content-Type: application/json" \
  -d '{"service":"openai","key":"sk-...","label":"GPT 키"}'

# เพิ่มคีย์ OpenRouter
curl -X POST https://localhost:56243/admin/keys \
  -H "$ADMIN" -H "Content-Type: application/json" \
  -d '{"service":"openrouter","key":"sk-or-v1-...","label":"OR 키"}'

# ลบคีย์ (SSE key_deleted กระจาย)
curl -X DELETE https://localhost:56243/admin/keys/key-abc123 -H "$ADMIN"

# รีเซ็ตปริมาณการใช้งานรายวัน
curl -X POST https://localhost:56243/admin/keys/reset -H "$ADMIN"

# รายการไคลเอ็นต์
curl -H "$ADMIN" https://localhost:56243/admin/clients

# เพิ่มไคลเอ็นต์ (OpenClaw)
curl -X POST https://localhost:56243/admin/clients \
  -H "$ADMIN" -H "Content-Type: application/json" \
  -d '{"id":"bot-a","name":"봇 A","agent_type":"openclaw","work_dir":"~/.openclaw","default_service":"google","default_model":"gemini-2.5-flash"}'

# เปลี่ยนโมเดลไคลเอ็นต์ (สะท้อนทันทีผ่าน SSE)
curl -X PUT https://localhost:56243/admin/clients/bot-a \
  -H "$ADMIN" -H "Content-Type: application/json" \
  -d '{"default_service":"openrouter","default_model":"wall-vault/hunter-alpha"}'

# ปิดใช้งานไคลเอ็นต์
curl -X PUT https://localhost:56243/admin/clients/my-bot \
  -H "$ADMIN" -H "Content-Type: application/json" \
  -d '{"enabled":false}'

# ลบไคลเอ็นต์
curl -X DELETE https://localhost:56243/admin/clients/my-bot -H "$ADMIN"

# รายการเซอร์วิส
curl -H "$ADMIN" https://localhost:56243/admin/services

# ตั้งค่า Ollama local URL (SSE service_changed กระจาย)
curl -X PUT https://localhost:56243/admin/services/ollama \
  -H "$ADMIN" -H "Content-Type: application/json" \
  -d '{"local_url":"http://192.168.x.x:11434","enabled":true}'

# เปิดใช้งานเซอร์วิส OpenAI
curl -X PUT https://localhost:56243/admin/services/openai \
  -H "$ADMIN" -H "Content-Type: application/json" \
  -d '{"enabled":true}'

# เพิ่มเซอร์วิสกำหนดเอง (SSE service_changed กระจาย)
curl -X POST https://localhost:56243/admin/services \
  -H "$ADMIN" -H "Content-Type: application/json" \
  -d '{"id":"my-llm","name":"사내 LLM","local_url":"http://10.0.0.50:8080","enabled":true}'

# ลบเซอร์วิสกำหนดเอง
curl -X DELETE https://localhost:56243/admin/services/my-llm -H "$ADMIN"

# สอบถามรายการโมเดล
curl -H "$ADMIN" https://localhost:56243/admin/models
curl -H "$ADMIN" "https://localhost:56243/admin/models?service=openrouter"
curl -H "$ADMIN" "https://localhost:56243/admin/models?q=hunter"

# สถานะ Proxy (heartbeat)
curl -H "$ADMIN" https://localhost:56243/admin/proxies

# ─── โหมดกระจาย — Proxy → Vault ──────────────────────────────────────────────

# สอบถามคีย์ที่ถอดรหัสแล้ว
curl https://localhost:56243/api/keys \
  -H "Authorization: Bearer your-bot-a-token"

# ส่ง Heartbeat
curl -X POST https://localhost:56243/api/heartbeat \
  -H "Authorization: Bearer your-bot-a-token" \
  -H "Content-Type: application/json" \
  -d '{"client_id":"bot-a","version":"v0.1.6.20260314.231308","service":"google","model":"gemini-2.5-flash","sse_connected":true}'
```

---

## Middleware

ใช้กับทุกคำขอโดยอัตโนมัติ:

| Middleware | ฟังก์ชัน |
|-----------|---------|
| **Logger** | บันทึกในรูปแบบ `[method] path status latencyms` |
| **CORS** | `Access-Control-Allow-Origin: *` |
| **Recovery** | กู้คืนจาก panic ส่งกลับ 500 |

---

*อัปเดตล่าสุด: 2026-03-16 — v0.1.7: dashboard title renamed to "벽금고(wall-vault) 대시보드", logo moved to sticky topbar, today_attempts tracking, HTTP 582 cooldown (5min), share-of-total bar scaling, custom/ routing fix, Ollama Timeout:0, key_att i18n, avatar heartbeat sync, build timestamp versioning, proxy-only service filter*
