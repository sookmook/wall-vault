# Manual API wall-vault

Dokumen ini menjelaskan secara rinci semua endpoint HTTP API wall-vault.

---

## Daftar Isi

- [Autentikasi](#autentikasi)
- [API Proxy (:56244)](#api-proxy-56244)
  - [Health Check](#get-health)
  - [Kueri Status](#get-status)
  - [Daftar Model](#get-apimodels)
  - [Ubah Model](#put-apiconfigmodel)
  - [Mode Berpikir](#put-apiconfigthink-mode)
  - [Refresh Pengaturan](#post-reload)
  - [API Gemini](#post-googlev1betamodelsmgeneratecontent)
  - [Streaming Gemini](#post-googlev1betamodelsmstreamgeneratecontent)
  - [API Kompatibel OpenAI](#post-v1chatcompletions)
- [API Key Vault (:56243)](#api-key-vault-56243)
  - [API Publik](#api-publik-tanpa-autentikasi)
  - [Stream Event SSE](#get-apievents)
  - [API Khusus Proxy](#api-khusus-proxy-token-klien)
  - [API Admin — Kunci](#api-admin--kunci-api)
  - [API Admin — Klien](#api-admin--klien)
  - [API Admin — Layanan](#api-admin--layanan)
  - [API Admin — Daftar Model](#api-admin--daftar-model)
  - [API Admin — Status Proxy](#api-admin--status-proxy)
- [Tipe Event SSE](#tipe-event-sse)
- [Routing Provider dan Model](#routing-provider-dan-model)
- [Skema Data](#skema-data)
- [Respons Error](#respons-error)
- [Kumpulan Contoh cURL](#kumpulan-contoh-curl)

---

## Autentikasi

| Cakupan | Metode | Header |
|---------|--------|--------|
| API Admin | Token Bearer | `Authorization: Bearer <admin_token>` |
| Proxy → Vault | Token Bearer | `Authorization: Bearer <client_token>` |
| API Proxy | Tidak ada (lokal) | — |

Jika `admin_token` tidak diatur (string kosong), semua API admin dapat diakses tanpa autentikasi.

### Kebijakan Keamanan

- **Rate Limiting**: Jika kegagalan autentikasi API admin melebihi 10 kali/15 menit, IP tersebut diblokir sementara (`429 Too Many Requests`)
- **Whitelist IP**: Hanya IP/CIDR yang terdaftar di field `ip_whitelist` agent (`Client`) yang diizinkan mengakses `/api/keys`. Array kosong mengizinkan semua.
- **Proteksi theme·lang**: `/admin/theme`, `/admin/lang` juga memerlukan autentikasi token admin

---

## API Proxy (:56244)

Server tempat proxy berjalan. Port default `56244`.

---

### `GET /health`

Health check. Selalu mengembalikan 200 OK.

**Contoh respons:**
```json
{
  "status": "ok",
  "version": "v0.1.6.20260314.231308",
  "client": "bot-a"
}
```

---

### `GET /status`

Kueri detail status proxy.

**Contoh respons:**
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

| Field | Tipe | Deskripsi |
|-------|------|-----------|
| `service` | string | Layanan default saat ini |
| `model` | string | Model default saat ini |
| `sse` | bool | Status koneksi SSE ke vault |
| `filter` | string | Mode filter tool |
| `services` | []string | Daftar layanan yang aktif |
| `mode` | string | `standalone` \| `distributed` |

---

### `GET /api/models`

Kueri daftar model yang tersedia. Menggunakan cache TTL (default 10 menit).

**Parameter kueri:**

| Parameter | Deskripsi | Contoh |
|-----------|-----------|--------|
| `service` | Filter layanan | `?service=google` |
| `q` | Pencarian ID/nama model | `?q=gemini` |

**Contoh respons:**
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

| Field | Tipe | Deskripsi |
|-------|------|-----------|
| `id` | string | ID model |
| `name` | string | Nama tampilan model |
| `service` | string | `google` \| `openai` \| `anthropic` \| `openrouter` \| `ollama` dll. |
| `context_length` | int | Ukuran context window |
| `free` | bool | Apakah model gratis (OpenRouter) |

---

### `PUT /api/config/model`

Mengubah layanan dan model saat ini.

**Body permintaan:**
```json
{
  "service": "google",
  "model": "gemini-2.5-flash"
}
```

**Respons:**
```json
{
  "status": "ok",
  "service": "google",
  "model": "gemini-2.5-flash"
}
```

> **Catatan:** Dalam mode distributed, disarankan menggunakan `PUT /admin/clients/{id}` vault daripada API ini. Perubahan vault akan otomatis tercermin dalam 1-3 detik melalui SSE.

---

### `PUT /api/config/think-mode`

Toggle mode berpikir (no-op, untuk ekspansi di masa depan).

**Respons:**
```json
{"status": "ok"}
```

---

### `POST /reload`

Sinkronisasi ulang pengaturan klien dan kunci dari vault secara instan.

**Respons:**
```json
{"status": "reloading"}
```

Sinkronisasi ulang berjalan secara asinkron dan selesai dalam 1-2 detik setelah respons diterima.

---

### `POST /google/v1beta/models/{model}:generateContent`

Proxy API Gemini (non-streaming).

**Parameter path:**
- `{model}`: ID model. Jika memiliki prefiks `gemini-`, layanan Google dipilih secara otomatis.

**Body permintaan:** [Format permintaan Gemini generateContent](https://ai.google.dev/api/generate-content)

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

**Body respons:** Format respons Gemini generateContent

**Filter tool:** Saat pengaturan `tool_filter: strip_all`, array `tools` dalam permintaan akan dihapus secara otomatis.

**Rantai fallback:** Layanan yang ditentukan gagal → fallback sesuai urutan layanan yang dikonfigurasi → Ollama (terakhir).

---

### `POST /google/v1beta/models/{model}:streamGenerateContent`

Proxy streaming API Gemini. Format permintaan sama dengan non-streaming. Respons berupa stream SSE:

```
data: {"candidates":[{"content":{"parts":[{"text":"안"}],...},...}]}

data: [DONE]
```

---

### `POST /v1/chat/completions`

API kompatibel OpenAI. Secara internal dikonversi ke format Gemini lalu diproses.

**Body permintaan:**
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

**Dukungan prefiks provider pada field `model` (OpenClaw 3.11+):**

| Contoh model | Routing |
|-------------|---------|
| `gemini-2.5-flash` | Layanan yang dikonfigurasi saat ini |
| `google/gemini-2.5-pro` | Langsung ke Google |
| `openai/gpt-4o` | Langsung ke OpenAI |
| `anthropic/claude-opus-4-6` | Via OpenRouter |
| `openrouter/meta-llama/llama-3.3-70b` | Langsung ke OpenRouter |
| `wall-vault/gemini-2.5-flash` | Deteksi otomatis → Google |
| `wall-vault/claude-opus-4-6` | Deteksi otomatis → OpenRouter Anthropic |
| `wall-vault/gpt-4o` | Deteksi otomatis → OpenAI |
| `wall-vault/hunter-alpha` | OpenRouter (gratis 1M context) |
| `moonshot/kimi-k2.5` | Via OpenRouter |
| `opencode-go/model` | Via OpenRouter |
| `kimi-k2.5:cloud` | Sufiks `:cloud` → OpenRouter |

Untuk detail lebih lanjut lihat [Routing Provider dan Model](#routing-provider-dan-model).

**Body respons:**
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

> **Penghapusan otomatis token kontrol model:** Jika respons mengandung delimiter GLM-5 / DeepSeek / ChatML (`<|im_start|>`, `[gMASK]`, `[sop]` dll.), token tersebut akan dihapus secara otomatis.

---

## API Key Vault (:56243)

Server tempat key vault berjalan. Port default `56243`.

---

### API Publik (Tanpa Autentikasi)

#### `GET /`

UI dashboard web. Diakses melalui browser.

---

#### `GET /api/status`

Kueri status vault.

**Contoh respons:**
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

Daftar klien yang terdaftar (hanya informasi publik, tanpa token).

---

### `GET /api/events`

Stream event SSE (Server-Sent Events) real-time.

**Header:**
```
Content-Type: text/event-stream
Cache-Control: no-cache
Connection: keep-alive
```

**Diterima segera setelah koneksi:**
```
data: {"type":"connected","clients":2}
```

**Contoh event:**
```
data: {"type":"config_change","data":{"client_id":"bot-a","service":"google","model":"gemini-2.5-pro"}}

data: {"type":"key_added","data":{"service":"openrouter"}}

data: {"type":"key_deleted","data":{"service":"google"}}

data: {"type":"service_changed","data":{"action":"updated","id":"ollama"}}

data: {"type":"usage_reset","data":{"time":"2026-03-13T00:00:30Z"}}
```

Untuk detail tipe event lihat [Tipe Event SSE](#tipe-event-sse).

---

### API Khusus Proxy (Token Klien)

Memerlukan header `Authorization: Bearer <client_token>`. Autentikasi dengan token admin juga dimungkinkan.

#### `GET /api/keys`

Daftar kunci API yang didekripsi untuk proxy.

**Parameter kueri:**

| Parameter | Deskripsi |
|-----------|-----------|
| `service` | Filter layanan (contoh: `?service=google`) |

**Contoh respons:**
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

> **Keamanan:** Mengembalikan kunci dalam teks biasa. Hanya kunci layanan yang diizinkan yang dikembalikan sesuai pengaturan `allowed_services` klien.

---

#### `GET /api/services`

Kueri daftar layanan yang digunakan proxy. Mengembalikan array ID layanan yang `proxy_enabled=true`.

**Contoh respons:**
```json
["google", "ollama"]
```

Jika array kosong, proxy menggunakan semua layanan tanpa batasan.

---

#### `POST /api/heartbeat`

Mengirim status proxy (dijalankan otomatis setiap 20 detik).

**Body permintaan:**
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

| Field | Tipe | Deskripsi |
|-------|------|-----------|
| `client_id` | string | ID klien |
| `version` | string | Versi proxy (termasuk build timestamp, contoh: `v0.1.6.20260314.231308`) |
| `service` | string | Layanan saat ini |
| `model` | string | Model saat ini |
| `sse_connected` | bool | Status koneksi SSE |
| `host` | string | Hostname |
| `avatar` | string | base64 data URI of the proxy's local avatar image (sent by proxy, auto-persisted to client record by vault). Set via `WV_AVATAR` env var (relative path under `~/.openclaw/`). |
| `key_usage` | map[string]int | Key ID → successful tokens used today. Vault calls `SetKeyUsage()` (absolute value sync). 429/402/582 errors do **not** increment this. |
| `key_attempts` | map[string]int | Key ID → total API calls today (success + rate-limited). Vault calls `SetKeyAttempts()`. 429/402/582 errors only increment this, not `key_usage`. |
| `key_cooldowns` | map[string]string | Key ID → cooldown end time (RFC3339). Vault calls `SetKeyCooldownIfLater()`. |

**Respons:**
```json
{"status": "ok"}
```

---

### API Admin — Kunci API

Memerlukan header `Authorization: Bearer <admin_token>`.

#### `GET /admin/keys`

Daftar semua kunci API yang terdaftar (tanpa kunci teks biasa).

**Contoh respons:**
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

| Field | Tipe | Deskripsi |
|-------|------|-----------|
| `today_usage` | int | Jumlah token permintaan berhasil hari ini (tidak termasuk error 429/402/582) |
| `today_attempts` | int | Total panggilan API hari ini (berhasil + rate-limited) |
| `available` | bool | Apakah tersedia untuk digunakan tanpa cooldown atau batas |
| `usage_pct` | int | Persentase penggunaan dari batas harian % (`daily_limit=0` berarti 0) |
| `cooldown_until` | RFC3339 | Waktu berakhirnya cooldown (nilai nol berarti tidak ada) |
| `last_error` | int | Kode error HTTP terakhir |

---

#### `POST /admin/keys`

Mendaftarkan kunci API baru. Event SSE `key_added` langsung di-broadcast saat pendaftaran.

**Body permintaan:**
```json
{
  "service": "google",
  "key": "AIzaSy...",
  "label": "메인 키",
  "daily_limit": 1000
}
```

| Field | Wajib | Deskripsi |
|-------|-------|-----------|
| `service` | ✅ | `google` \| `openai` \| `anthropic` \| `openrouter` \| `ollama` \| kustom |
| `key` | ✅ | Kunci API teks biasa |
| `label` | — | Label identifikasi |
| `daily_limit` | — | Batas penggunaan harian (0 = tidak terbatas) |

---

#### `DELETE /admin/keys/{id}`

Menghapus kunci API. Event SSE `key_deleted` di-broadcast setelah penghapusan.

**Respons:**
```json
{"status": "deleted"}
```

---

#### `POST /admin/keys/reset`

Mereset penggunaan harian semua kunci. Broadcast event SSE `usage_reset`.

**Respons:**
```json
{
  "status": "reset",
  "time": "2026-03-13T15:00:00Z"
}
```

---

### API Admin — Klien

#### `GET /admin/clients`

Daftar semua klien (termasuk token).

---

#### `POST /admin/clients`

Mendaftarkan klien baru.

**Body permintaan:**
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

| Field | Wajib | Deskripsi |
|-------|-------|-----------|
| `id` | ✅ | ID unik klien |
| `name` | — | Nama tampilan |
| `token` | — | Token autentikasi (dibuat otomatis jika tidak diisi) |
| `default_service` | — | Layanan default |
| `default_model` | — | Model default |
| `allowed_services` | — | Daftar layanan yang diizinkan (array kosong = izinkan semua) |
| `agent_type` | — | `openclaw` \| `claude-code` \| `cursor` \| `vscode` \| `custom` |
| `work_dir` | — | Direktori kerja agent |
| `description` | — | Deskripsi agent |
| `ip_whitelist` | — | Daftar IP yang diizinkan (array kosong = izinkan semua, mendukung CIDR) |
| `enabled` | — | Status aktif (default `true`) |

---

#### `GET /admin/clients/{id}`

Kueri klien tertentu (termasuk token).

---

#### `PUT /admin/clients/{id}`

Mengubah pengaturan klien. **Broadcast SSE `config_change` → tercermin di proxy dalam 1-3 detik.**

**Body permintaan (hanya field yang diubah):**
```json
{
  "default_service": "openrouter",
  "default_model": "anthropic/claude-opus-4-6",
  "enabled": true
}
```

**Respons:**
```json
{"status": "updated"}
```

---

#### `DELETE /admin/clients/{id}`

Menghapus klien.

---

### API Admin — Layanan

#### `GET /admin/services`

Daftar layanan yang terdaftar.

**Contoh respons:**
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

8 layanan bawaan: `google`, `openai`, `anthropic`, `openrouter`, `github-copilot`, `ollama`, `lmstudio`, `vllm`

---

#### `POST /admin/services`

Menambahkan layanan kustom. Setelah ditambahkan, event SSE `service_changed` di-broadcast → **dropdown dashboard langsung diperbarui**.

**Body permintaan:**
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

Memperbarui pengaturan layanan. Setelah perubahan, event SSE `service_changed` di-broadcast.

**Body permintaan:**
```json
{
  "local_url": "http://192.168.x.x:11434",
  "enabled": true
}
```

---

#### `DELETE /admin/services/{id}`

Menghapus layanan kustom. Setelah penghapusan, event SSE `service_changed` di-broadcast.

Percobaan menghapus layanan bawaan (`custom: false`):
```json
{"error": "기본 서비스는 삭제할 수 없습니다: google"}
```

---

### API Admin — Daftar Model

#### `GET /admin/models`

Kueri daftar model per layanan. Menggunakan cache TTL (10 menit).

**Parameter kueri:**

| Parameter | Deskripsi | Contoh |
|-----------|-----------|--------|
| `service` | Filter layanan | `?service=google` |
| `q` | Pencarian model | `?q=gemini` |

**Metode kueri model per layanan:**

| Layanan | Metode | Jumlah |
|---------|--------|--------|
| `google` | Daftar tetap | 8 (termasuk embedding) |
| `openai` | Daftar tetap | 9 |
| `anthropic` | Daftar tetap | 6 |
| `github-copilot` | Daftar tetap | 6 |
| `openrouter` | Kueri dinamis via API (fallback ke 14 model terkurasi saat gagal) | 340+ |
| `ollama` | Kueri dinamis dari server lokal (7 rekomendasi saat tidak merespons) | Variabel |
| `lmstudio` | Kueri dinamis dari server lokal | Variabel |
| `vllm` | Kueri dinamis dari server lokal | Variabel |
| Kustom | `/v1/models` kompatibel OpenAI | Variabel |

**Daftar model fallback OpenRouter (saat API tidak merespons):**

| Model | Catatan |
|-------|---------|
| `openrouter/hunter-alpha` | Gratis, 1M context |
| `openrouter/healer-alpha` | Gratis, omni-modal |
| `moonshot/kimi-k2.5` | 256K context |
| `z-ai/glm-5`, `z-ai/glm-4.7-flash` | — |
| `deepseek/deepseek-r1`, `deepseek/deepseek-chat` | — |
| `qwen/qwen-2.5-72b-instruct` | 131K context |
| `minimax/minimax-m2.5` | — |
| `meta-llama/llama-3.3-70b-instruct` | 131K context |

---

### API Admin — Status Proxy

#### `GET /admin/proxies`

Status Heartbeat terakhir dari semua proxy yang terhubung.

---

## Tipe Event SSE

Event yang diterima dari stream `/api/events` vault:

| `type` | Kondisi terjadinya | Isi `data` | Respons dashboard |
|--------|-------------------|------------|-------------------|
| `connected` | Segera setelah koneksi SSE | `{"clients": N}` | — |
| `config_change` | Perubahan pengaturan klien | `{"client_id","service","model"}` | Pembaruan dropdown model kartu agent |
| `key_added` | Pendaftaran kunci API baru | `{"service": "google"}` | Pembaruan dropdown model |
| `key_deleted` | Penghapusan kunci API | `{"service": "google"}` | Pembaruan dropdown model |
| `service_changed` | Penambahan/perubahan/penghapusan layanan | `{"action":"added"\|"updated"\|"deleted","id":"...","proxy_services":["google","ollama"]}` | Pembaruan instan select layanan + dropdown model; pembaruan real-time daftar layanan dispatch proxy |
| `usage_update` | Saat menerima heartbeat proxy (setiap 20 detik) | `{"keys":[{"id","service","today_usage","today_attempts","daily_limit","cooldown_until"},...]}` | Pembaruan instan bar dan angka penggunaan kunci, mulai countdown cooldown. Menggunakan data SSE langsung tanpa fetch. Bar menggunakan penskalaan share-of-total (kunci tidak terbatas). |
| `usage_reset` | Reset penggunaan harian | `{"time": "RFC3339"}` | Refresh halaman |

**Pemrosesan event yang diterima proxy:**

```
config_change diterima
  → Jika client_id cocok dengan milik sendiri
    → Perbarui service, model secara instan
    → hooksMgr.Fire(EventModelChanged)
```

---

## Routing Provider dan Model

Saat menentukan format `provider/model` di field `model` `/v1/chat/completions`, routing otomatis dilakukan (kompatibel dengan OpenClaw 3.11).

### Aturan Routing Prefiks

| Prefiks | Target routing | Contoh |
|---------|---------------|--------|
| `google/` | Langsung ke Google | `google/gemini-2.5-pro` |
| `openai/` | Langsung ke OpenAI | `openai/gpt-4o` |
| `anthropic/` | Via OpenRouter | `anthropic/claude-opus-4-6` |
| `ollama/` | Langsung ke Ollama | `ollama/qwen3.5:35b` |
| `custom/` | Parsing ulang rekursif (hapus `custom/` lalu re-routing) | `custom/google/gemini-2.5-flash` → Google |
| `openrouter/` | OpenRouter (pertahankan bare path) | `openrouter/meta-llama/llama-3.3-70b` |
| `opencode/`, `opencode-go/`, `opencode-zen/` | OpenRouter | `opencode-go/model` |
| `moonshot/`, `kimi-coding/` | OpenRouter (pertahankan full path) | `moonshot/kimi-k2.5` |
| `groq/`, `mistral/`, `minimax/`, `cohere/`, `perplexity/` | OpenRouter | `groq/llama-3.1-70b` |
| `together/`, `huggingface/`, `nvidia/`, `venice/` | OpenRouter | — |
| `meta-llama/`, `qwen/`, `deepseek/`, `01-ai/` | OpenRouter (full path) | `deepseek/deepseek-r1` |

### Deteksi Otomatis Prefiks `wall-vault/`

Prefiks wall-vault sendiri yang secara otomatis menentukan layanan dari ID model.

| Pola ID model | Routing |
|--------------|---------|
| `wall-vault/gemini-*` | Google |
| `wall-vault/gpt-*`, `wall-vault/o1`, `wall-vault/o3`, `wall-vault/o4*` | OpenAI |
| `wall-vault/claude-*` | OpenRouter (path Anthropic) |
| `wall-vault/hunter-alpha`, `wall-vault/healer-alpha` | OpenRouter (gratis 1M ctx) |
| `wall-vault/kimi-*`, `wall-vault/glm-*`, `wall-vault/deepseek-*`, `wall-vault/qwen*` | OpenRouter |
| Lainnya | OpenRouter |

### Pemrosesan Sufiks `:cloud`

Sufiks `:cloud` dalam format tag Ollama secara otomatis dihapus lalu di-routing ke OpenRouter.

```
kimi-k2.5:cloud  →  OpenRouter, ID model: kimi-k2.5
glm-5:cloud      →  OpenRouter, ID model: glm-5
```

### Contoh Integrasi OpenClaw openclaw.json

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

Dengan mengklik tombol **🐾** pada kartu agent, snippet pengaturan untuk agent tersebut akan otomatis disalin ke clipboard.

---

## Skema Data

### APIKey

| Field | Tipe | Deskripsi |
|-------|------|-----------|
| `id` | string | ID unik format UUID |
| `service` | string | `google` \| `openai` \| `anthropic` \| `openrouter` \| `ollama` \| kustom |
| `encrypted_key` | string | Kunci terenkripsi AES-GCM (Base64) |
| `label` | string | Label identifikasi |
| `today_usage` | int | Jumlah token permintaan berhasil hari ini (tidak termasuk error 429/402/582) |
| `today_attempts` | int | Total panggilan API hari ini (berhasil + rate-limited; direset saat tengah malam) |
| `daily_limit` | int | Batas harian (0 = tidak terbatas) |
| `cooldown_until` | time.Time | Waktu berakhirnya cooldown |
| `last_error` | int | Kode error HTTP terakhir |
| `created_at` | time.Time | Waktu pendaftaran |

**Kebijakan cooldown:**

| Error HTTP | Cooldown |
|-----------|----------|
| 429 (Too Many Requests) | 30 menit |
| 402 (Payment Required) | 24 jam |
| 400 / 401 / 403 | 24 jam |
| 582 (Gateway Overload) | 5 menit |
| Error jaringan | 10 menit |

> **429·402·582**: Set cooldown + increment `today_attempts`. `today_usage` tidak berubah (hanya menghitung token berhasil).
> **Ollama (layanan lokal)**: `callOllama` menggunakan HTTP client khusus dengan `Timeout: 0` (tidak terbatas). Inferensi model besar bisa memakan waktu puluhan detik hingga beberapa menit, sehingga timeout default 60 detik tidak diterapkan.

### Client

| Field | Tipe | Deskripsi |
|-------|------|-----------|
| `id` | string | ID unik klien |
| `name` | string | Nama tampilan |
| `token` | string | Token autentikasi |
| `default_service` | string | Layanan default |
| `default_model` | string | Model default (bisa dalam format `provider/model`) |
| `allowed_services` | []string | Layanan yang diizinkan (array kosong = semua) |
| `agent_type` | string | `openclaw` \| `claude-code` \| `cursor` \| `vscode` \| `custom` |
| `work_dir` | string | Direktori kerja agent |
| `description` | string | Deskripsi |
| `ip_whitelist` | []string | Daftar IP yang diizinkan (mendukung CIDR) |
| `avatar` | string | Agent avatar — relative path under `~/.openclaw/` (e.g. `workspace/avatar.png`, `workspace/avatars/profile.hpg`) OR base64 data URI (`data:image/png;base64,...`). Supported extensions: `.png`, `.jpg`/`.jpeg`/`.hpg`, `.webp`, `.gif`. |
| `enabled` | bool | Jika `false` mengembalikan `403` saat mengakses `/api/keys` |
| `created_at` | time.Time | Waktu pendaftaran |

### ServiceConfig

| Field | Tipe | Deskripsi |
|-------|------|-----------|
| `id` | string | ID unik layanan |
| `name` | string | Nama tampilan |
| `local_url` | string | URL server lokal (Ollama/LMStudio/vLLM/kustom) |
| `enabled` | bool | Status aktif |
| `custom` | bool | Apakah layanan ditambahkan pengguna |
| `proxy_enabled` | bool | Whether this service is included in proxy dispatch and agent model dropdowns. Controlled by the "프록시 사용" checkbox in the Services card. Only `proxy_enabled: true` services appear in agent service/model selectors. |

### ProxyStatus (Heartbeat)

| Field | Tipe | Deskripsi |
|-------|------|-----------|
| `client_id` | string | ID klien |
| `version` | string | Versi proxy (contoh: `v0.1.6.20260314.231308`) |
| `service` | string | Layanan saat ini |
| `model` | string | Model saat ini |
| `sse_connected` | bool | Status koneksi SSE |
| `host` | string | Hostname |
| `avatar` | string | base64 data URI of local avatar (auto-synced from proxy via heartbeat) |
| `updated_at` | time.Time | Pembaruan terakhir |
| `vault.today_usage` | int | Penggunaan token hari ini |
| `vault.daily_limit` | int | Batas harian |
| `vault.key_status` | string | `active` \| `cooldown` \| `exhausted` |

---

## Respons Error

```json
{"error": "Pesan error"}
```

| Kode | Arti |
|------|------|
| 200 | Berhasil |
| 400 | Permintaan tidak valid |
| 401 | Gagal autentikasi |
| 403 | Akses ditolak (klien tidak aktif, IP diblokir) |
| 404 | Sumber daya tidak ditemukan |
| 405 | Metode tidak diizinkan |
| 429 | Batas rate terlampaui |
| 500 | Error internal server |
| 502 | Error API upstream (semua fallback gagal) |

---

## Kumpulan Contoh cURL

```bash
# ─── Proxy ────────────────────────────────────────────────────────────────────

# Health check
curl https://localhost:56244/health

# Kueri status
curl https://localhost:56244/status

# Daftar model (semua)
curl https://localhost:56244/api/models

# Hanya model Google
curl "https://localhost:56244/api/models?service=google"

# Pencarian model gratis
curl "https://localhost:56244/api/models?q=alpha"

# Ubah model (lokal)
curl -X PUT https://localhost:56244/api/config/model \
  -H "Content-Type: application/json" \
  -d '{"service":"google","model":"gemini-2.5-pro"}'

# Refresh pengaturan
curl -X POST https://localhost:56244/reload

# Panggilan langsung API Gemini
curl -X POST "https://localhost:56244/google/v1beta/models/gemini-2.5-flash:generateContent" \
  -H "Content-Type: application/json" \
  -d '{"contents":[{"role":"user","parts":[{"text":"안녕하세요"}]}]}'

# Kompatibel OpenAI (model default)
curl -X POST https://localhost:56244/v1/chat/completions \
  -H "Content-Type: application/json" \
  -d '{"model":"gemini-2.5-flash","messages":[{"role":"user","content":"안녕"}]}'

# Format OpenClaw provider/model
curl -X POST https://localhost:56244/v1/chat/completions \
  -H "Authorization: Bearer my-agent-token" \
  -H "Content-Type: application/json" \
  -d '{"model":"wall-vault/gemini-2.5-flash","messages":[{"role":"user","content":"안녕"}]}'

# Menggunakan model gratis 1M context
curl -X POST https://localhost:56244/v1/chat/completions \
  -H "Authorization: Bearer my-agent-token" \
  -H "Content-Type: application/json" \
  -d '{"model":"wall-vault/hunter-alpha","messages":[{"role":"user","content":"안녕"}]}'

# ─── Key Vault (Publik) ──────────────────────────────────────────────────────

curl https://localhost:56243/api/status
curl https://localhost:56243/api/clients
curl -s https://localhost:56243/api/events --max-time 3

# ─── Key Vault (Admin) ───────────────────────────────────────────────────────

ADMIN="Authorization: Bearer admin-token"

# Daftar kunci
curl -H "$ADMIN" https://localhost:56243/admin/keys

# Tambah kunci Google
curl -X POST https://localhost:56243/admin/keys \
  -H "$ADMIN" -H "Content-Type: application/json" \
  -d '{"service":"google","key":"AIzaSy...","label":"메인 키","daily_limit":1000}'

# Tambah kunci OpenAI
curl -X POST https://localhost:56243/admin/keys \
  -H "$ADMIN" -H "Content-Type: application/json" \
  -d '{"service":"openai","key":"sk-...","label":"GPT 키"}'

# Tambah kunci OpenRouter
curl -X POST https://localhost:56243/admin/keys \
  -H "$ADMIN" -H "Content-Type: application/json" \
  -d '{"service":"openrouter","key":"sk-or-v1-...","label":"OR 키"}'

# Hapus kunci (broadcast SSE key_deleted)
curl -X DELETE https://localhost:56243/admin/keys/key-abc123 -H "$ADMIN"

# Reset penggunaan harian
curl -X POST https://localhost:56243/admin/keys/reset -H "$ADMIN"

# Daftar klien
curl -H "$ADMIN" https://localhost:56243/admin/clients

# Tambah klien (OpenClaw)
curl -X POST https://localhost:56243/admin/clients \
  -H "$ADMIN" -H "Content-Type: application/json" \
  -d '{"id":"bot-a","name":"봇 A","agent_type":"openclaw","work_dir":"~/.openclaw","default_service":"google","default_model":"gemini-2.5-flash"}'

# Ubah model klien (tercermin instan via SSE)
curl -X PUT https://localhost:56243/admin/clients/bot-a \
  -H "$ADMIN" -H "Content-Type: application/json" \
  -d '{"default_service":"openrouter","default_model":"wall-vault/hunter-alpha"}'

# Nonaktifkan klien
curl -X PUT https://localhost:56243/admin/clients/my-bot \
  -H "$ADMIN" -H "Content-Type: application/json" \
  -d '{"enabled":false}'

# Hapus klien
curl -X DELETE https://localhost:56243/admin/clients/my-bot -H "$ADMIN"

# Daftar layanan
curl -H "$ADMIN" https://localhost:56243/admin/services

# Set URL lokal Ollama (broadcast SSE service_changed)
curl -X PUT https://localhost:56243/admin/services/ollama \
  -H "$ADMIN" -H "Content-Type: application/json" \
  -d '{"local_url":"http://192.168.x.x:11434","enabled":true}'

# Aktifkan layanan OpenAI
curl -X PUT https://localhost:56243/admin/services/openai \
  -H "$ADMIN" -H "Content-Type: application/json" \
  -d '{"enabled":true}'

# Tambah layanan kustom (broadcast SSE service_changed)
curl -X POST https://localhost:56243/admin/services \
  -H "$ADMIN" -H "Content-Type: application/json" \
  -d '{"id":"my-llm","name":"사내 LLM","local_url":"http://10.0.0.50:8080","enabled":true}'

# Hapus layanan kustom
curl -X DELETE https://localhost:56243/admin/services/my-llm -H "$ADMIN"

# Kueri daftar model
curl -H "$ADMIN" https://localhost:56243/admin/models
curl -H "$ADMIN" "https://localhost:56243/admin/models?service=openrouter"
curl -H "$ADMIN" "https://localhost:56243/admin/models?q=hunter"

# Status proxy (heartbeat)
curl -H "$ADMIN" https://localhost:56243/admin/proxies

# ─── Mode Distributed — Proxy → Vault ────────────────────────────────────────

# Kueri kunci yang didekripsi
curl https://localhost:56243/api/keys \
  -H "Authorization: Bearer your-bot-a-token"

# Kirim Heartbeat
curl -X POST https://localhost:56243/api/heartbeat \
  -H "Authorization: Bearer your-bot-a-token" \
  -H "Content-Type: application/json" \
  -d '{"client_id":"bot-a","version":"v0.1.6.20260314.231308","service":"google","model":"gemini-2.5-flash","sse_connected":true}'
```

---

## Middleware

Diterapkan secara otomatis pada semua permintaan:

| Middleware | Fungsi |
|-----------|--------|
| **Logger** | Logging format `[method] path status latencyms` |
| **CORS** | `Access-Control-Allow-Origin: *` |
| **Recovery** | Pemulihan dari panic, mengembalikan respons 500 |

---

*Terakhir diperbarui: 2026-03-16 — v0.1.7: dashboard title renamed to "벽금고(wall-vault) 대시보드", logo moved to sticky topbar, today_attempts tracking, HTTP 582 cooldown (5min), share-of-total bar scaling, custom/ routing fix, Ollama Timeout:0, key_att i18n, avatar heartbeat sync, build timestamp versioning, proxy-only service filter*
