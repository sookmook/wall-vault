# دليل واجهة برمجة التطبيقات (API) لـ wall-vault

يصف هذا المستند جميع نقاط نهاية HTTP API الخاصة بـ wall-vault بالتفصيل.

---

## جدول المحتويات

- [المصادقة](#المصادقة)
- [واجهة البروكسي (:56244)](#واجهة-البروكسي-56244)
  - [فحص الصحة](#get-health)
  - [استعلام الحالة](#get-status)
  - [قائمة النماذج](#get-apimodels)
  - [تغيير النموذج](#put-apiconfigmodel)
  - [وضع التفكير](#put-apiconfigthink-mode)
  - [تحديث الإعدادات](#post-reload)
  - [واجهة Gemini](#post-googlev1betamodelsmgeneratecontent)
  - [بث Gemini](#post-googlev1betamodelsmstreamgeneratecontent)
  - [واجهة متوافقة مع OpenAI](#post-v1chatcompletions)
- [واجهة خزنة المفاتيح (:56243)](#واجهة-خزنة-المفاتيح-56243)
  - [الواجهة العامة](#الواجهة-العامة-بدون-مصادقة)
  - [بث أحداث SSE](#get-apievents)
  - [واجهة خاصة بالبروكسي](#واجهة-خاصة-بالبروكسي-رمز-العميل)
  - [واجهة المسؤول — المفاتيح](#واجهة-المسؤول--مفاتيح-api)
  - [واجهة المسؤول — العملاء](#واجهة-المسؤول--العملاء)
  - [واجهة المسؤول — الخدمات](#واجهة-المسؤول--الخدمات)
  - [واجهة المسؤول — قائمة النماذج](#واجهة-المسؤول--قائمة-النماذج)
  - [واجهة المسؤول — حالة البروكسي](#واجهة-المسؤول--حالة-البروكسي)
- [أنواع أحداث SSE](#أنواع-أحداث-sse)
- [توجيه المزود والنموذج](#توجيه-المزودوالنموذج)
- [مخطط البيانات](#مخطط-البيانات)
- [استجابات الخطأ](#استجابات-الخطأ)
- [مجموعة أمثلة cURL](#مجموعة-أمثلة-curl)

---

## المصادقة

| النطاق | الطريقة | الترويسة |
|--------|---------|----------|
| واجهة المسؤول | رمز Bearer | `Authorization: Bearer <admin_token>` |
| البروكسي → الخزنة | رمز Bearer | `Authorization: Bearer <client_token>` |
| واجهة البروكسي | بدون (محلي) | — |

إذا لم يتم تعيين `admin_token` (سلسلة فارغة)، يمكن الوصول إلى جميع واجهات المسؤول بدون مصادقة.

### سياسة الأمان

- **تقييد المعدل**: عند تجاوز 10 محاولات فاشلة لمصادقة واجهة المسؤول خلال 15 دقيقة، يتم حظر عنوان IP مؤقتاً (`429 Too Many Requests`)
- **قائمة IP المسموح بها**: يُسمح فقط لعناوين IP/CIDR المسجلة في حقل `ip_whitelist` الخاص بالعميل (`Client`) بالوصول إلى `/api/keys`. المصفوفة الفارغة تسمح للجميع.
- **حماية theme·lang**: نقاط `/admin/theme` و `/admin/lang` تتطلب أيضاً مصادقة رمز المسؤول

---

## واجهة البروكسي (:56244)

الخادم الذي يعمل عليه البروكسي. المنفذ الافتراضي `56244`.

---

### `GET /health`

فحص الصحة. يُرجع دائماً 200 OK.

**مثال على الاستجابة:**
```json
{
  "status": "ok",
  "version": "v0.1.6.20260314.231308",
  "client": "bot-a"
}
```

---

### `GET /status`

استعلام تفصيلي عن حالة البروكسي.

**مثال على الاستجابة:**
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

| الحقل | النوع | الوصف |
|-------|-------|-------|
| `service` | string | الخدمة الافتراضية الحالية |
| `model` | string | النموذج الافتراضي الحالي |
| `sse` | bool | حالة اتصال SSE بالخزنة |
| `filter` | string | وضع فلتر الأدوات |
| `services` | []string | قائمة الخدمات المفعّلة |
| `mode` | string | `standalone` \| `distributed` |

---

### `GET /api/models`

استعلام قائمة النماذج المتاحة. يستخدم ذاكرة تخزين مؤقت TTL (افتراضي 10 دقائق).

**معلمات الاستعلام:**

| المعلمة | الوصف | مثال |
|---------|-------|------|
| `service` | فلتر الخدمة | `?service=google` |
| `q` | بحث بمعرف/اسم النموذج | `?q=gemini` |

**مثال على الاستجابة:**
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

| الحقل | النوع | الوصف |
|-------|-------|-------|
| `id` | string | معرف النموذج |
| `name` | string | اسم العرض للنموذج |
| `service` | string | `google` \| `openai` \| `anthropic` \| `openrouter` \| `ollama` إلخ |
| `context_length` | int | حجم نافذة السياق |
| `free` | bool | ما إذا كان النموذج مجانياً (OpenRouter) |

---

### `PUT /api/config/model`

تغيير الخدمة والنموذج الحاليين.

**نص الطلب:**
```json
{
  "service": "google",
  "model": "gemini-2.5-flash"
}
```

**الاستجابة:**
```json
{
  "status": "ok",
  "service": "google",
  "model": "gemini-2.5-flash"
}
```

> **ملاحظة:** في الوضع الموزع، يُوصى باستخدام `PUT /admin/clients/{id}` الخاص بالخزنة بدلاً من هذه الواجهة. تنعكس تغييرات الخزنة تلقائياً خلال 1-3 ثوانٍ عبر SSE.

---

### `PUT /api/config/think-mode`

تبديل وضع التفكير (no-op، للتوسع المستقبلي).

**الاستجابة:**
```json
{"status": "ok"}
```

---

### `POST /reload`

إعادة مزامنة إعدادات العميل والمفاتيح من الخزنة فوراً.

**الاستجابة:**
```json
{"status": "reloading"}
```

تعمل إعادة المزامنة بشكل غير متزامن وتكتمل خلال 1-2 ثانية بعد تلقي الاستجابة.

---

### `POST /google/v1beta/models/{model}:generateContent`

بروكسي واجهة Gemini (بدون بث).

**معلمة المسار:**
- `{model}`: معرف النموذج. إذا كان يحتوي على بادئة `gemini-`، يتم اختيار خدمة Google تلقائياً.

**نص الطلب:** [صيغة طلب Gemini generateContent](https://ai.google.dev/api/generate-content)

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

**نص الاستجابة:** صيغة استجابة Gemini generateContent

**فلتر الأدوات:** عند تعيين `tool_filter: strip_all`، يتم إزالة مصفوفة `tools` من الطلب تلقائياً.

**سلسلة التراجع:** فشل الخدمة المحددة → التراجع بترتيب الخدمات المُعدّة → Ollama (أخيراً).

---

### `POST /google/v1beta/models/{model}:streamGenerateContent`

بروكسي بث واجهة Gemini. صيغة الطلب مطابقة لوضع عدم البث. الاستجابة هي بث SSE:

```
data: {"candidates":[{"content":{"parts":[{"text":"안"}],...},...}]}

data: [DONE]
```

---

### `POST /v1/chat/completions`

واجهة متوافقة مع OpenAI. يتم تحويلها داخلياً إلى صيغة Gemini ثم معالجتها.

**نص الطلب:**
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

**دعم بادئة المزود في حقل `model` (OpenClaw 3.11+):**

| مثال النموذج | التوجيه |
|-------------|---------|
| `gemini-2.5-flash` | الخدمة المُعدّة حالياً |
| `google/gemini-2.5-pro` | مباشرة إلى Google |
| `openai/gpt-4o` | مباشرة إلى OpenAI |
| `anthropic/claude-opus-4-6` | عبر OpenRouter |
| `openrouter/meta-llama/llama-3.3-70b` | مباشرة إلى OpenRouter |
| `wall-vault/gemini-2.5-flash` | اكتشاف تلقائي → Google |
| `wall-vault/claude-opus-4-6` | اكتشاف تلقائي → OpenRouter Anthropic |
| `wall-vault/gpt-4o` | اكتشاف تلقائي → OpenAI |
| `wall-vault/hunter-alpha` | OpenRouter (مجاني 1M context) |
| `moonshot/kimi-k2.5` | عبر OpenRouter |
| `opencode-go/model` | عبر OpenRouter |
| `kimi-k2.5:cloud` | لاحقة `:cloud` → OpenRouter |

لمزيد من التفاصيل راجع [توجيه المزود والنموذج](#توجيه-المزودوالنموذج).

**نص الاستجابة:**
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

> **إزالة رموز التحكم بالنموذج تلقائياً:** إذا احتوت الاستجابة على فواصل GLM-5 / DeepSeek / ChatML مثل (`<|im_start|>`, `[gMASK]`, `[sop]` إلخ) فسيتم إزالتها تلقائياً.

---

## واجهة خزنة المفاتيح (:56243)

الخادم الذي تعمل عليه خزنة المفاتيح. المنفذ الافتراضي `56243`.

---

### الواجهة العامة (بدون مصادقة)

#### `GET /`

واجهة لوحة التحكم. يتم الوصول إليها عبر المتصفح.

---

#### `GET /api/status`

استعلام حالة الخزنة.

**مثال على الاستجابة:**
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

قائمة العملاء المسجلين (معلومات عامة فقط، بدون رموز المصادقة).

---

### `GET /api/events`

بث أحداث SSE (Server-Sent Events) في الوقت الحقيقي.

**الترويسات:**
```
Content-Type: text/event-stream
Cache-Control: no-cache
Connection: keep-alive
```

**يتم استلامها فور الاتصال:**
```
data: {"type":"connected","clients":2}
```

**أمثلة على الأحداث:**
```
data: {"type":"config_change","data":{"client_id":"bot-a","service":"google","model":"gemini-2.5-pro"}}

data: {"type":"key_added","data":{"service":"openrouter"}}

data: {"type":"key_deleted","data":{"service":"google"}}

data: {"type":"service_changed","data":{"action":"updated","id":"ollama"}}

data: {"type":"usage_reset","data":{"time":"2026-03-13T00:00:30Z"}}
```

لمزيد من تفاصيل أنواع الأحداث راجع [أنواع أحداث SSE](#أنواع-أحداث-sse).

---

### واجهة خاصة بالبروكسي (رمز العميل)

تتطلب ترويسة `Authorization: Bearer <client_token>`. يمكن أيضاً المصادقة برمز المسؤول.

#### `GET /api/keys`

قائمة مفاتيح API المفكّكة التشفير المقدمة للبروكسي.

**معلمات الاستعلام:**

| المعلمة | الوصف |
|---------|-------|
| `service` | فلتر الخدمة (مثال: `?service=google`) |

**مثال على الاستجابة:**
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

> **الأمان:** تُرجع المفاتيح بنص عادي. يتم إرجاع مفاتيح الخدمات المسموح بها فقط وفقاً لإعداد `allowed_services` الخاص بالعميل.

---

#### `GET /api/services`

استعلام قائمة الخدمات التي يستخدمها البروكسي. تُرجع مصفوفة معرفات الخدمات التي `proxy_enabled=true`.

**مثال على الاستجابة:**
```json
["google", "ollama"]
```

إذا كانت المصفوفة فارغة، يستخدم البروكسي جميع الخدمات بدون قيود.

---

#### `POST /api/heartbeat`

إرسال حالة البروكسي (يتم تنفيذه تلقائياً كل 20 ثانية).

**نص الطلب:**
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

| الحقل | النوع | الوصف |
|-------|-------|-------|
| `client_id` | string | معرف العميل |
| `version` | string | إصدار البروكسي (يتضمن طابع وقت البناء، مثال: `v0.1.6.20260314.231308`) |
| `service` | string | الخدمة الحالية |
| `model` | string | النموذج الحالي |
| `sse_connected` | bool | حالة اتصال SSE |
| `host` | string | اسم المضيف |
| `avatar` | string | base64 data URI of the proxy's local avatar image (sent by proxy, auto-persisted to client record by vault). Set via `WV_AVATAR` env var (relative path under `~/.openclaw/`). |
| `key_usage` | map[string]int | Key ID → successful tokens used today. Vault calls `SetKeyUsage()` (absolute value sync). 429/402/582 errors do **not** increment this. |
| `key_attempts` | map[string]int | Key ID → total API calls today (success + rate-limited). Vault calls `SetKeyAttempts()`. 429/402/582 errors only increment this, not `key_usage`. |
| `key_cooldowns` | map[string]string | Key ID → cooldown end time (RFC3339). Vault calls `SetKeyCooldownIfLater()`. |

**الاستجابة:**
```json
{"status": "ok"}
```

---

### واجهة المسؤول — مفاتيح API

تتطلب ترويسة `Authorization: Bearer <admin_token>`.

#### `GET /admin/keys`

قائمة جميع مفاتيح API المسجلة (بدون المفاتيح بنص عادي).

**مثال على الاستجابة:**
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

| الحقل | النوع | الوصف |
|-------|-------|-------|
| `today_usage` | int | عدد رموز الطلبات الناجحة اليوم (لا تشمل أخطاء 429/402/582) |
| `today_attempts` | int | إجمالي عدد استدعاءات API اليوم (الناجحة + المقيدة بالمعدل) |
| `available` | bool | ما إذا كان متاحاً للاستخدام بدون تبريد أو حد |
| `usage_pct` | int | نسبة الاستخدام من الحد اليومي % (`daily_limit=0` تعني 0) |
| `cooldown_until` | RFC3339 | وقت انتهاء التبريد (القيمة الصفرية تعني لا يوجد) |
| `last_error` | int | آخر رمز خطأ HTTP |

---

#### `POST /admin/keys`

تسجيل مفتاح API جديد. يتم بث حدث SSE `key_added` فوراً عند التسجيل.

**نص الطلب:**
```json
{
  "service": "google",
  "key": "AIzaSy...",
  "label": "메인 키",
  "daily_limit": 1000
}
```

| الحقل | مطلوب | الوصف |
|-------|-------|-------|
| `service` | ✅ | `google` \| `openai` \| `anthropic` \| `openrouter` \| `ollama` \| مخصص |
| `key` | ✅ | مفتاح API بنص عادي |
| `label` | — | تسمية للتعريف |
| `daily_limit` | — | حد الاستخدام اليومي (0 = غير محدود) |

---

#### `DELETE /admin/keys/{id}`

حذف مفتاح API. يتم بث حدث SSE `key_deleted` بعد الحذف.

**الاستجابة:**
```json
{"status": "deleted"}
```

---

#### `POST /admin/keys/reset`

إعادة تعيين الاستخدام اليومي لجميع المفاتيح. بث حدث SSE `usage_reset`.

**الاستجابة:**
```json
{
  "status": "reset",
  "time": "2026-03-13T15:00:00Z"
}
```

---

### واجهة المسؤول — العملاء

#### `GET /admin/clients`

قائمة جميع العملاء (تتضمن الرموز).

---

#### `POST /admin/clients`

تسجيل عميل جديد.

**نص الطلب:**
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

| الحقل | مطلوب | الوصف |
|-------|-------|-------|
| `id` | ✅ | معرف العميل الفريد |
| `name` | — | اسم العرض |
| `token` | — | رمز المصادقة (يتم إنشاؤه تلقائياً إذا لم يُحدد) |
| `default_service` | — | الخدمة الافتراضية |
| `default_model` | — | النموذج الافتراضي |
| `allowed_services` | — | قائمة الخدمات المسموح بها (مصفوفة فارغة = السماح للجميع) |
| `agent_type` | — | `openclaw` \| `claude-code` \| `cursor` \| `vscode` \| `custom` |
| `work_dir` | — | دليل عمل الوكيل |
| `description` | — | وصف الوكيل |
| `ip_whitelist` | — | قائمة عناوين IP المسموح بها (مصفوفة فارغة = السماح للجميع، يدعم CIDR) |
| `enabled` | — | حالة التفعيل (الافتراضي `true`) |

---

#### `GET /admin/clients/{id}`

استعلام عميل محدد (يتضمن الرمز).

---

#### `PUT /admin/clients/{id}`

تغيير إعدادات العميل. **بث SSE `config_change` → ينعكس على البروكسي خلال 1-3 ثوانٍ.**

**نص الطلب (الحقول المراد تغييرها فقط):**
```json
{
  "default_service": "openrouter",
  "default_model": "anthropic/claude-opus-4-6",
  "enabled": true
}
```

**الاستجابة:**
```json
{"status": "updated"}
```

---

#### `DELETE /admin/clients/{id}`

حذف عميل.

---

### واجهة المسؤول — الخدمات

#### `GET /admin/services`

قائمة الخدمات المسجلة.

**مثال على الاستجابة:**
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

8 خدمات مدمجة: `google`, `openai`, `anthropic`, `openrouter`, `github-copilot`, `ollama`, `lmstudio`, `vllm`

---

#### `POST /admin/services`

إضافة خدمة مخصصة. بعد الإضافة يتم بث حدث SSE `service_changed` → **تحديث فوري للقوائم المنسدلة في لوحة التحكم**.

**نص الطلب:**
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

تحديث إعدادات الخدمة. بعد التغيير يتم بث حدث SSE `service_changed`.

**نص الطلب:**
```json
{
  "local_url": "http://192.168.x.x:11434",
  "enabled": true
}
```

---

#### `DELETE /admin/services/{id}`

حذف خدمة مخصصة. بعد الحذف يتم بث حدث SSE `service_changed`.

محاولة حذف خدمة مدمجة (`custom: false`):
```json
{"error": "기본 서비스는 삭제할 수 없습니다: google"}
```

---

### واجهة المسؤول — قائمة النماذج

#### `GET /admin/models`

استعلام قائمة النماذج حسب الخدمة. يستخدم ذاكرة تخزين مؤقت TTL (10 دقائق).

**معلمات الاستعلام:**

| المعلمة | الوصف | مثال |
|---------|-------|------|
| `service` | فلتر الخدمة | `?service=google` |
| `q` | بحث النماذج | `?q=gemini` |

**طريقة استعلام النماذج حسب الخدمة:**

| الخدمة | الطريقة | العدد |
|--------|---------|-------|
| `google` | قائمة ثابتة | 8 (تشمل embedding) |
| `openai` | قائمة ثابتة | 9 |
| `anthropic` | قائمة ثابتة | 6 |
| `github-copilot` | قائمة ثابتة | 6 |
| `openrouter` | استعلام ديناميكي عبر API (تراجع إلى 14 نموذج مختار عند الفشل) | أكثر من 340 |
| `ollama` | استعلام ديناميكي من الخادم المحلي (7 مقترحات عند عدم الاستجابة) | متغير |
| `lmstudio` | استعلام ديناميكي من الخادم المحلي | متغير |
| `vllm` | استعلام ديناميكي من الخادم المحلي | متغير |
| مخصص | `/v1/models` متوافق مع OpenAI | متغير |

**قائمة نماذج OpenRouter الاحتياطية (عند عدم استجابة API):**

| النموذج | ملاحظات |
|---------|---------|
| `openrouter/hunter-alpha` | مجاني، 1M context |
| `openrouter/healer-alpha` | مجاني، omni-modal |
| `moonshot/kimi-k2.5` | 256K context |
| `z-ai/glm-5`, `z-ai/glm-4.7-flash` | — |
| `deepseek/deepseek-r1`, `deepseek/deepseek-chat` | — |
| `qwen/qwen-2.5-72b-instruct` | 131K context |
| `minimax/minimax-m2.5` | — |
| `meta-llama/llama-3.3-70b-instruct` | 131K context |

---

### واجهة المسؤول — حالة البروكسي

#### `GET /admin/proxies`

آخر حالة Heartbeat لجميع البروكسيات المتصلة.

---

## أنواع أحداث SSE

الأحداث المستلمة من بث `/api/events` الخاص بالخزنة:

| `type` | شرط الحدوث | محتوى `data` | استجابة لوحة التحكم |
|--------|------------|--------------|---------------------|
| `connected` | فور اتصال SSE | `{"clients": N}` | — |
| `config_change` | تغيير إعدادات العميل | `{"client_id","service","model"}` | تحديث القائمة المنسدلة لنموذج بطاقة الوكيل |
| `key_added` | تسجيل مفتاح API جديد | `{"service": "google"}` | تحديث القائمة المنسدلة للنماذج |
| `key_deleted` | حذف مفتاح API | `{"service": "google"}` | تحديث القائمة المنسدلة للنماذج |
| `service_changed` | إضافة/تعديل/حذف خدمة | `{"action":"added"\|"updated"\|"deleted","id":"...","proxy_services":["google","ollama"]}` | تحديث فوري لقائمة الخدمات + القائمة المنسدلة للنماذج؛ تحديث فوري لقائمة خدمات dispatch البروكسي |
| `usage_update` | عند استلام heartbeat من البروكسي (كل 20 ثانية) | `{"keys":[{"id","service","today_usage","today_attempts","daily_limit","cooldown_until"},...]}` | تحديث فوري لأشرطة وأرقام استخدام المفاتيح، بدء العد التنازلي للتبريد. استخدام مباشر لبيانات SSE بدون fetch. الأشرطة تستخدم تحجيم share-of-total (مفاتيح غير محدودة). |
| `usage_reset` | إعادة تعيين الاستخدام اليومي | `{"time": "RFC3339"}` | تحديث الصفحة |

**معالجة الأحداث المستلمة من البروكسي:**

```
config_change مستلم
  → إذا تطابق client_id مع العميل الحالي
    → تحديث فوري للخدمة والنموذج
    → hooksMgr.Fire(EventModelChanged)
```

---

## توجيه المزود والنموذج

عند تحديد صيغة `provider/model` في حقل `model` الخاص بـ `/v1/chat/completions`، يتم التوجيه التلقائي (متوافق مع OpenClaw 3.11).

### قواعد التوجيه بالبادئة

| البادئة | هدف التوجيه | مثال |
|---------|-------------|------|
| `google/` | مباشرة إلى Google | `google/gemini-2.5-pro` |
| `openai/` | مباشرة إلى OpenAI | `openai/gpt-4o` |
| `anthropic/` | عبر OpenRouter | `anthropic/claude-opus-4-6` |
| `ollama/` | مباشرة إلى Ollama | `ollama/qwen3.5:35b` |
| `custom/` | إعادة تحليل تكرارية (إزالة `custom/` وإعادة التوجيه) | `custom/google/gemini-2.5-flash` → Google |
| `openrouter/` | OpenRouter (الحفاظ على المسار الأساسي) | `openrouter/meta-llama/llama-3.3-70b` |
| `opencode/`, `opencode-go/`, `opencode-zen/` | OpenRouter | `opencode-go/model` |
| `moonshot/`, `kimi-coding/` | OpenRouter (الحفاظ على المسار الكامل) | `moonshot/kimi-k2.5` |
| `groq/`, `mistral/`, `minimax/`, `cohere/`, `perplexity/` | OpenRouter | `groq/llama-3.1-70b` |
| `together/`, `huggingface/`, `nvidia/`, `venice/` | OpenRouter | — |
| `meta-llama/`, `qwen/`, `deepseek/`, `01-ai/` | OpenRouter (المسار الكامل) | `deepseek/deepseek-r1` |

### اكتشاف تلقائي لبادئة `wall-vault/`

بادئة wall-vault الخاصة التي تحدد الخدمة تلقائياً من معرف النموذج.

| نمط معرف النموذج | التوجيه |
|-----------------|---------|
| `wall-vault/gemini-*` | Google |
| `wall-vault/gpt-*`, `wall-vault/o1`, `wall-vault/o3`, `wall-vault/o4*` | OpenAI |
| `wall-vault/claude-*` | OpenRouter (مسار Anthropic) |
| `wall-vault/hunter-alpha`, `wall-vault/healer-alpha` | OpenRouter (مجاني 1M ctx) |
| `wall-vault/kimi-*`, `wall-vault/glm-*`, `wall-vault/deepseek-*`, `wall-vault/qwen*` | OpenRouter |
| أخرى | OpenRouter |

### معالجة لاحقة `:cloud`

يتم إزالة لاحقة `:cloud` بصيغة وسم Ollama تلقائياً ثم التوجيه إلى OpenRouter.

```
kimi-k2.5:cloud  →  OpenRouter, معرف النموذج: kimi-k2.5
glm-5:cloud      →  OpenRouter, معرف النموذج: glm-5
```

### مثال ربط OpenClaw openclaw.json

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

بالنقر على زر **🐾** في بطاقة الوكيل، يتم نسخ مقتطف الإعدادات الخاص بذلك الوكيل تلقائياً إلى الحافظة.

---

## مخطط البيانات

### APIKey

| الحقل | النوع | الوصف |
|-------|-------|-------|
| `id` | string | معرف فريد بصيغة UUID |
| `service` | string | `google` \| `openai` \| `anthropic` \| `openrouter` \| `ollama` \| مخصص |
| `encrypted_key` | string | مفتاح مشفر بـ AES-GCM (Base64) |
| `label` | string | تسمية للتعريف |
| `today_usage` | int | عدد رموز الطلبات الناجحة اليوم (لا تشمل أخطاء 429/402/582) |
| `today_attempts` | int | إجمالي استدعاءات API اليوم (الناجحة + المقيدة بالمعدل؛ يُعاد التعيين عند منتصف الليل) |
| `daily_limit` | int | الحد اليومي (0 = غير محدود) |
| `cooldown_until` | time.Time | وقت انتهاء التبريد |
| `last_error` | int | آخر رمز خطأ HTTP |
| `created_at` | time.Time | وقت التسجيل |

**سياسة التبريد:**

| خطأ HTTP | التبريد |
|----------|---------|
| 429 (Too Many Requests) | 30 دقيقة |
| 402 (Payment Required) | 24 ساعة |
| 400 / 401 / 403 | 24 ساعة |
| 582 (Gateway Overload) | 5 دقائق |
| خطأ في الشبكة | 10 دقائق |

> **429·402·582**: تعيين تبريد + زيادة `today_attempts`. لا تغيير في `today_usage` (يتم حساب الرموز الناجحة فقط).
> **Ollama (خدمة محلية)**: يستخدم `callOllama` عميل HTTP خاص بـ `Timeout: 0` (غير محدود). قد يستغرق استدلال النماذج الكبيرة عشرات الثواني إلى عدة دقائق، لذا لا يتم تطبيق المهلة الافتراضية البالغة 60 ثانية.

### Client

| الحقل | النوع | الوصف |
|-------|-------|-------|
| `id` | string | معرف العميل الفريد |
| `name` | string | اسم العرض |
| `token` | string | رمز المصادقة |
| `default_service` | string | الخدمة الافتراضية |
| `default_model` | string | النموذج الافتراضي (يمكن أن يكون بصيغة `provider/model`) |
| `allowed_services` | []string | الخدمات المسموح بها (مصفوفة فارغة = الكل) |
| `agent_type` | string | `openclaw` \| `claude-code` \| `cursor` \| `vscode` \| `custom` |
| `work_dir` | string | دليل عمل الوكيل |
| `description` | string | الوصف |
| `ip_whitelist` | []string | قائمة عناوين IP المسموح بها (يدعم CIDR) |
| `avatar` | string | Agent avatar — relative path under `~/.openclaw/` (e.g. `workspace/avatar.png`, `workspace/avatars/profile.hpg`) OR base64 data URI (`data:image/png;base64,...`). Supported extensions: `.png`, `.jpg`/`.jpeg`/`.hpg`, `.webp`, `.gif`. |
| `enabled` | bool | إذا كان `false` يُرجع `403` عند الوصول إلى `/api/keys` |
| `created_at` | time.Time | وقت التسجيل |

### ServiceConfig

| الحقل | النوع | الوصف |
|-------|-------|-------|
| `id` | string | معرف الخدمة الفريد |
| `name` | string | اسم العرض |
| `local_url` | string | عنوان URL للخادم المحلي (Ollama/LMStudio/vLLM/مخصص) |
| `enabled` | bool | حالة التفعيل |
| `custom` | bool | ما إذا كانت خدمة مضافة من المستخدم |
| `proxy_enabled` | bool | Whether this service is included in proxy dispatch and agent model dropdowns. Controlled by the "프록시 사용" checkbox in the Services card. Only `proxy_enabled: true` services appear in agent service/model selectors. |

### ProxyStatus (Heartbeat)

| الحقل | النوع | الوصف |
|-------|-------|-------|
| `client_id` | string | معرف العميل |
| `version` | string | إصدار البروكسي (مثال: `v0.1.6.20260314.231308`) |
| `service` | string | الخدمة الحالية |
| `model` | string | النموذج الحالي |
| `sse_connected` | bool | حالة اتصال SSE |
| `host` | string | اسم المضيف |
| `avatar` | string | base64 data URI of local avatar (auto-synced from proxy via heartbeat) |
| `updated_at` | time.Time | آخر تحديث |
| `vault.today_usage` | int | استخدام الرموز اليوم |
| `vault.daily_limit` | int | الحد اليومي |
| `vault.key_status` | string | `active` \| `cooldown` \| `exhausted` |

---

## استجابات الخطأ

```json
{"error": "رسالة الخطأ"}
```

| الرمز | المعنى |
|-------|--------|
| 200 | نجاح |
| 400 | طلب غير صالح |
| 401 | فشل المصادقة |
| 403 | رفض الوصول (عميل غير نشط، حظر IP) |
| 404 | المورد غير موجود |
| 405 | طريقة غير مسموح بها |
| 429 | تجاوز حد المعدل |
| 500 | خطأ داخلي في الخادم |
| 502 | خطأ في واجهة المنبع (فشل جميع البدائل) |

---

## مجموعة أمثلة cURL

```bash
# ─── البروكسي ─────────────────────────────────────────────────────────────────

# فحص الصحة
curl https://localhost:56244/health

# استعلام الحالة
curl https://localhost:56244/status

# قائمة النماذج (الكل)
curl https://localhost:56244/api/models

# نماذج Google فقط
curl "https://localhost:56244/api/models?service=google"

# بحث النماذج المجانية
curl "https://localhost:56244/api/models?q=alpha"

# تغيير النموذج (محلي)
curl -X PUT https://localhost:56244/api/config/model \
  -H "Content-Type: application/json" \
  -d '{"service":"google","model":"gemini-2.5-pro"}'

# تحديث الإعدادات
curl -X POST https://localhost:56244/reload

# استدعاء مباشر لواجهة Gemini
curl -X POST "https://localhost:56244/google/v1beta/models/gemini-2.5-flash:generateContent" \
  -H "Content-Type: application/json" \
  -d '{"contents":[{"role":"user","parts":[{"text":"안녕하세요"}]}]}'

# متوافق مع OpenAI (النموذج الافتراضي)
curl -X POST https://localhost:56244/v1/chat/completions \
  -H "Content-Type: application/json" \
  -d '{"model":"gemini-2.5-flash","messages":[{"role":"user","content":"안녕"}]}'

# صيغة OpenClaw provider/model
curl -X POST https://localhost:56244/v1/chat/completions \
  -H "Authorization: Bearer my-agent-token" \
  -H "Content-Type: application/json" \
  -d '{"model":"wall-vault/gemini-2.5-flash","messages":[{"role":"user","content":"안녕"}]}'

# استخدام نموذج مجاني 1M context
curl -X POST https://localhost:56244/v1/chat/completions \
  -H "Authorization: Bearer my-agent-token" \
  -H "Content-Type: application/json" \
  -d '{"model":"wall-vault/hunter-alpha","messages":[{"role":"user","content":"안녕"}]}'

# ─── خزنة المفاتيح (عامة) ─────────────────────────────────────────────────────

curl https://localhost:56243/api/status
curl https://localhost:56243/api/clients
curl -s https://localhost:56243/api/events --max-time 3

# ─── خزنة المفاتيح (المسؤول) ───────────────────────────────────────────────────

ADMIN="Authorization: Bearer admin-token"

# قائمة المفاتيح
curl -H "$ADMIN" https://localhost:56243/admin/keys

# إضافة مفتاح Google
curl -X POST https://localhost:56243/admin/keys \
  -H "$ADMIN" -H "Content-Type: application/json" \
  -d '{"service":"google","key":"AIzaSy...","label":"메인 키","daily_limit":1000}'

# إضافة مفتاح OpenAI
curl -X POST https://localhost:56243/admin/keys \
  -H "$ADMIN" -H "Content-Type: application/json" \
  -d '{"service":"openai","key":"sk-...","label":"GPT 키"}'

# إضافة مفتاح OpenRouter
curl -X POST https://localhost:56243/admin/keys \
  -H "$ADMIN" -H "Content-Type: application/json" \
  -d '{"service":"openrouter","key":"sk-or-v1-...","label":"OR 키"}'

# حذف مفتاح (بث SSE key_deleted)
curl -X DELETE https://localhost:56243/admin/keys/key-abc123 -H "$ADMIN"

# إعادة تعيين الاستخدام اليومي
curl -X POST https://localhost:56243/admin/keys/reset -H "$ADMIN"

# قائمة العملاء
curl -H "$ADMIN" https://localhost:56243/admin/clients

# إضافة عميل (OpenClaw)
curl -X POST https://localhost:56243/admin/clients \
  -H "$ADMIN" -H "Content-Type: application/json" \
  -d '{"id":"bot-a","name":"봇 A","agent_type":"openclaw","work_dir":"~/.openclaw","default_service":"google","default_model":"gemini-2.5-flash"}'

# تغيير نموذج العميل (انعكاس فوري عبر SSE)
curl -X PUT https://localhost:56243/admin/clients/bot-a \
  -H "$ADMIN" -H "Content-Type: application/json" \
  -d '{"default_service":"openrouter","default_model":"wall-vault/hunter-alpha"}'

# تعطيل عميل
curl -X PUT https://localhost:56243/admin/clients/my-bot \
  -H "$ADMIN" -H "Content-Type: application/json" \
  -d '{"enabled":false}'

# حذف عميل
curl -X DELETE https://localhost:56243/admin/clients/my-bot -H "$ADMIN"

# قائمة الخدمات
curl -H "$ADMIN" https://localhost:56243/admin/services

# تعيين عنوان URL المحلي لـ Ollama (بث SSE service_changed)
curl -X PUT https://localhost:56243/admin/services/ollama \
  -H "$ADMIN" -H "Content-Type: application/json" \
  -d '{"local_url":"http://192.168.x.x:11434","enabled":true}'

# تفعيل خدمة OpenAI
curl -X PUT https://localhost:56243/admin/services/openai \
  -H "$ADMIN" -H "Content-Type: application/json" \
  -d '{"enabled":true}'

# إضافة خدمة مخصصة (بث SSE service_changed)
curl -X POST https://localhost:56243/admin/services \
  -H "$ADMIN" -H "Content-Type: application/json" \
  -d '{"id":"my-llm","name":"사내 LLM","local_url":"http://10.0.0.50:8080","enabled":true}'

# حذف خدمة مخصصة
curl -X DELETE https://localhost:56243/admin/services/my-llm -H "$ADMIN"

# استعلام قائمة النماذج
curl -H "$ADMIN" https://localhost:56243/admin/models
curl -H "$ADMIN" "https://localhost:56243/admin/models?service=openrouter"
curl -H "$ADMIN" "https://localhost:56243/admin/models?q=hunter"

# حالة البروكسي (heartbeat)
curl -H "$ADMIN" https://localhost:56243/admin/proxies

# ─── الوضع الموزع — البروكسي → الخزنة ───────────────────────────────────────

# استعلام المفاتيح المفكّكة التشفير
curl https://localhost:56243/api/keys \
  -H "Authorization: Bearer your-bot-a-token"

# إرسال Heartbeat
curl -X POST https://localhost:56243/api/heartbeat \
  -H "Authorization: Bearer your-bot-a-token" \
  -H "Content-Type: application/json" \
  -d '{"client_id":"bot-a","version":"v0.1.6.20260314.231308","service":"google","model":"gemini-2.5-flash","sse_connected":true}'
```

---

## البرمجيات الوسيطة

تُطبق تلقائياً على جميع الطلبات:

| البرمجية الوسيطة | الوظيفة |
|------------------|---------|
| **Logger** | تسجيل بصيغة `[method] path status latencyms` |
| **CORS** | `Access-Control-Allow-Origin: *` |
| **Recovery** | استعادة من الذعر، إرجاع استجابة 500 |

---

*آخر تحديث: 2026-03-16 — v0.1.7: dashboard title renamed to "벽금고(wall-vault) 대시보드", logo moved to sticky topbar, today_attempts tracking, HTTP 582 cooldown (5min), share-of-total bar scaling, custom/ routing fix, Ollama Timeout:0, key_att i18n, avatar heartbeat sync, build timestamp versioning, proxy-only service filter*
