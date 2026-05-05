# wall-vault

> **خزنة مفاتيح API + وكيل ذكاء اصطناعي في ملف Go ثنائي واحد.**
> يخزّن المفاتيح محلياً باستخدام AES-GCM، ويدوّرها بين المزوّدين، ويتحوّل إلى البديل عند فشل أحدهم، ويأتي مع لوحة تحكّم لحظية.

[![License: GPL v3](https://img.shields.io/badge/License-GPLv3-blue.svg)](LICENSE)
[![Go](https://img.shields.io/badge/Go-1.25-00ADD8.svg)](go.mod)

[English](README.md) · [한국어](README.ko.md) · [中文](README.zh.md) · [日本語](README.ja.md) · [Español](README.es.md) · [Français](README.fr.md) · [Deutsch](README.de.md) · [Português](README.pt.md) · **العربية** · [हिन्दी](README.hi.md) · [Bahasa Indonesia](README.id.md) · [ภาษาไทย](README.th.md) · [Kiswahili](README.sw.md) · [Hausa](README.ha.md) · [नेपाली](README.ne.md) · [Монгол](README.mn.md) · [isiZulu](README.zu.md)

---

## ما هو

يقع wall-vault بين وكيل الذكاء الاصطناعي (OpenClaw، Claude Code، Cursor، Continue، أو السكربت الخاص بك) ومزوّدي الذكاء الاصطناعي السحابيين أو المحليّين الذين يتواصل معهم. شيئان في ثنائي واحد:

- **Vault** — يخزّن مفاتيح API مشفّرة عند الراحة (AES-GCM بكلمة سر رئيسية)، يدوّرها، يتتبّع استخدام كل مفتاح وفترات تهدئته، يبثّ التغييرات عبر SSE، ويقدّم لوحة تحكّم ويب على `:56243`.
- **Proxy** — يكشف نقاط نهاية متوافقة مع Gemini وAnthropic وOpenAI على `:56244`، يختار مفتاحاً من الـ vault، يرسل الطلب إلى المصدر الذي عيّنته، ويتحوّل إلى المزوّد التالي عند فشل أحدهم.

يدعم أربعة أشكال للطلبات (Gemini `:generateContent`، Anthropic `/v1/messages`، OpenAI `/v1/chat/completions`، وOllama-native `/api/chat`) وخمس فئات من المصادر:

| المزوّد | ملاحظات |
|----------|-------|
| Google Gemini | API الأصلي؛ تدوير المفاتيح حسب المشروع |
| Anthropic | تمرير أصلي عبر `/v1/messages` |
| OpenAI | `/v1/chat/completions` الأصلي |
| OpenRouter | أكثر من 340 نموذجاً، تحوّل تلقائي إلى متغيّرات `:free` |
| Ollama / LM Studio / vLLM / llama.cpp / Jan / KoboldCpp / TabbyAPI / mlx-server / LiteLLM | خوادم محليّة متوافقة مع OpenAI؛ إضافة جاهزة عبر ملف yaml للإضافة |

إضافة خادم خلفي جديد متوافق مع OpenAI تتطلّب ملف yaml واحداً تحت `~/.wall-vault/services/` — دون أي تغيير في الكود.

## لماذا قد ترغب فيه

- أنت تتعامل مع ثلاث أو أربع خدمات ذكاء اصطناعي وتريد عنواناً واحداً يتواصل معه الوكيل.
- تريد لمفتاح من الفئة المجانيّة في فترة تهدئة أن يفسح المجال للمفتاح التالي دون كسر الجلسة.
- تريد للمفاتيح ذاتها أن تشغّل عدة بوتات / IDE / سكربتات على الشبكة المحليّة نفسها دون نسخ بيانات الاعتماد.
- تريد لوحة تحكّم لتعديل مفاتيح API بدلاً من متغيّرات البيئة.
- تريد خياراً يعمل أولاً محلياً (Ollama / LM Studio) عند نفاد حدود السحابة.

## بداية سريعة

### التثبيت (Linux / macOS)

```bash
curl -fsSL https://raw.githubusercontent.com/sookmook/wall-vault/main/install.sh | sh
```

أو نزّل ثنائياً مبنيّاً مسبقاً مباشرةً:

```bash
# Linux amd64
curl -L https://github.com/sookmook/wall-vault/releases/latest/download/wall-vault-linux-amd64 \
  -o wall-vault && chmod +x wall-vault

# macOS Apple Silicon
curl -L https://github.com/sookmook/wall-vault/releases/latest/download/wall-vault-darwin-arm64 \
  -o wall-vault && chmod +x wall-vault

# Linux arm64 (Raspberry Pi، خوادم ARM)
curl -L https://github.com/sookmook/wall-vault/releases/latest/download/wall-vault-linux-arm64 \
  -o wall-vault && chmod +x wall-vault
```

### التثبيت (Windows)

```powershell
Invoke-WebRequest -Uri `
  "https://github.com/sookmook/wall-vault/releases/latest/download/wall-vault-windows-amd64.exe" `
  -OutFile wall-vault.exe
```

### التشغيل الأول

```bash
wall-vault setup    # معالج تفاعلي — يختار المنفذ والخدمات ورمز المسؤول وكلمة السر الرئيسيّة
wall-vault start    # يشغّل كلاً من vault وproxy
```

افتح `http://localhost:56243` (أو `https://...` بعد تفعيل TLS — انظر أدناه) في المتصفّح. تطلب اللوحة رمز المسؤول الذي طبعه `setup`. من هناك تضيف مفاتيح API، تسجّل العملاء، وتبدّل النماذج دون إعادة تشغيل.

---

## TLS (موصى به)

افتراضياً يكتب `wall-vault setup` تكوينه دون TLS، لذا يستجيب المستمعان كلاهما عبر HTTP عادي. عناوين URL في هذا الملف تستخدم `https://localhost:56244` لأن أكثر الوكلاء (OpenClaw، Claude Code، Cursor) يفضّلون نقطة نهاية واحدة محميّة بـ TLS لا تنكسر إذا نقلت الـ proxy لاحقاً إلى مضيف آخر. لمطابقة هذه الأمثلة، فعّل TLS مرّة واحدة باستخدام الـ CA الداخلي المرفق:

```bash
# 1. إنشاء CA الداخلي لـ wall-vault (مرّة واحدة، يقع في ~/.wall-vault/ca.{crt,key})
wall-vault cert init

# 2. إصدار شهادة مضيف لهذه الآلة
#    SANs تشمل اسم المضيف، localhost، 127.0.0.1، وأي عنوان IP محلي مكتشف
wall-vault cert issue $(hostname)

# 3. الوثوق بـ CA في keychain نظام التشغيل المحلي
wall-vault cert install-trust

# 4. تحويل المستمعين إلى TLS
export WV_PROXY_TLS_ENABLED=1
export WV_PROXY_TLS_CERT="$HOME/.wall-vault/$(hostname).crt"
export WV_PROXY_TLS_KEY="$HOME/.wall-vault/$(hostname).key"
export WV_VAULT_TLS_ENABLED=1
export WV_VAULT_TLS_CERT="$HOME/.wall-vault/$(hostname).crt"
export WV_VAULT_TLS_KEY="$HOME/.wall-vault/$(hostname).key"
wall-vault start
```

لآلة أخرى على الشبكة المحليّة: انسخ `~/.wall-vault/ca.crt` إليها وشغّل `wall-vault cert install-trust --ca <path>` هناك. بمجرّد الوثوق بـ CA في كل مكان، يمكن لكل آلة على الشبكة الوصول إلى الـ proxy عبر `https://<host>:56244` دون تحذيرات شهادة.

إذا فضّلت البقاء على HTTP العادي، اترك التكوين كما هو واستبدل `https://` بـ `http://` في مقتطفات العميل أدناه. كلا المخطّطين يعملان؛ الفرق هو أيّ منفذ يستجيب لمصافحة TLS.

**احتياطي حلقة الاسترجاع.** العملاء على المضيف نفسه الذين لا يستطيعون احترام CA الخاص بـ wall-vault (لا سيما وقت التشغيل المضمّن لـ Node في OpenClaw، الذي يعيد كتابة `NODE_EXTRA_CA_CERTS` عند الإطلاق) يصلون إلى الـ proxy عبر مرافق HTTP عادي على حلقة الاسترجاع فقط على `127.0.0.1:56245`. يفعّله wall-vault تلقائياً عند تفعيل TLS.

---

## ربط العملاء

وجّه أي عميل ذكاء اصطناعي إلى `https://<host>:56244` (أو `http://...` إذا كان TLS معطّلاً). يستجيب الـ proxy لأربعة أشكال:

| الصيغة | المسار | عملاء مثاليّون |
|--------|------|-----------------|
| Gemini | `/google/v1beta/models/<model>:generateContent` | OpenClaw، Gemini CLI، Antigravity |
| Anthropic | `/v1/messages` | Claude Code، حزم Anthropic SDK |
| OpenAI | `/v1/chat/completions` | Cursor، Continue، السكربتات المخصّصة، أكثر تطبيقات LLM |
| Ollama-native | `/api/chat` | عملاء Ollama الذين يمرّرون عبره |

### Claude Code

```bash
export ANTHROPIC_BASE_URL=https://localhost:56244
export ANTHROPIC_API_KEY=<your-vault-client-token>
claude
```

عند نفاد رصيد Anthropic في المصدر، ينتقل التوجيه إلى المزوّدين الذين عيّنتهم في `fallback_services` لهذا العميل. للاشتراك الصريح في احتياطي غير-Claude:

```yaml
proxy:
  anthropic_fallback_model: "claude-haiku-4-5-20251001"
```

(القيمة الافتراضية الفارغة تجعل التوجيه يُرجع خطأً ليظهر سوء التوجيه فوراً.)

### Cursor / Continue

في Cursor **Settings → AI → OpenAI API**:

```
Base URL:  https://localhost:56244
API Key:   <your-vault-client-token>
Model:     gemini-2.5-flash    # أو أي نموذج يعرفه wall-vault
```

Continue (`config.json`):

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

### OpenClaw

OpenClaw هو إطار وكيل TUI الذي بُني wall-vault أصلاً لخدمته. مودال **Add Agent** في اللوحة يضبط نوع الوكيل إلى `openclaw` (أو `nanoclaw`)؛ ثم يكتب wall-vault `~/.openclaw/openclaw.json` مباشرةً، شاملاً عناوين URL للمزوّدين ورمز vault وإدخالات النماذج:

```bash
VAULT_CLIENT_ID=my-bot \
VAULT_URL=http://<vault-host>:56243 \
VAULT_TOKEN=<your-client-token> \
wall-vault proxy
```

### curl / السكربتات

```bash
curl -X POST https://localhost:56244/v1/chat/completions \
  -H "Authorization: Bearer <your-vault-client-token>" \
  -H "Content-Type: application/json" \
  -d '{
    "model": "gemini-2.5-flash",
    "messages": [{"role": "user", "content": "ping"}]
  }'
```

---

## التكوين

يكتب `wall-vault setup` إمّا `./wall-vault.yaml` أو `~/.wall-vault/config.yaml`. عدّل يدوياً الحقول التي لا يسأل عنها المعالج.

```yaml
mode: standalone     # standalone | distributed
lang: en             # en | ko | zh | ja | es | fr | de | pt | ar | hi | id | th | sw | ha | ne | mn | zu
theme: light         # light | dark | cherry | ocean | gold | autumn | winter

proxy:
  port: 56244
  host: ""           # افتراضياً: 127.0.0.1 في standalone، 0.0.0.0 في distributed
  client_id: my-bot
  vault_url: ""      # distributed: http://vault-host:56243
  vault_token: ""    # distributed: رمز العميل
  tool_filter: strip_all   # strip_all | whitelist | passthrough
  allowed_tools: []
  services: [google, openrouter, ollama]
  timeout: 300s
  tls:
    enabled: false
    cert_file: ""
    key_file: ""
  plain_port: 56245              # مرافق HTTP على حلقة الاسترجاع فقط عند تفعيل TLS
  ollama_keep_alive: "30m"       # "-1" لا يُفرَغ مطلقاً، "0" يُفرَغ فوراً
  ollama_num_ctx: 8192
  oai_stream_forward: false      # تمرير SSE حقيقي للخادم الخلفي بالاشتراك
  anthropic_fallback_model: ""   # إعادة كتابة غير-Claude بالاشتراك على توجيه anthropic

vault:
  port: 56243
  host: ""
  admin_token: ""
  admin_ip_whitelist: []
  master_password: ""            # كلمة سر تشفير المفاتيح بـ AES-GCM
  data_dir: ~/.wall-vault/data
  services_dir: ~/.wall-vault/services
  tls:
    enabled: false
    cert_file: ""
    key_file: ""
  bootstrap_port: 56247          # مستمع HTTP عادي يقدّم ca.crt فقط

doctor:
  interval: 5m
  auto_fix: true
  log_file: /tmp/wall-vault-doctor.log

hooks:
  on_model_change: ""            # أمر shell (env: SERVICE, MODEL)
  on_key_exhausted: ""
  on_service_down: ""
  on_doctor_fix: ""
  openclaw_socket: ""
```

### متغيّرات البيئة

كل حقل YAML له بديل بمتغيّر بيئة يفوز على الملف. الشائعة منها:

| المتغيّر | الوصف |
|----------|-------------|
| `WV_LANG`, `WV_THEME` | اللغة والسمة |
| `WV_PROXY_PORT`, `WV_PROXY_HOST` | عنوان استماع الـ proxy |
| `WV_VAULT_PORT`, `WV_VAULT_HOST` | عنوان استماع الـ vault |
| `WV_VAULT_URL`, `WV_VAULT_TOKEN` | نقاط نهاية وضع distributed |
| `WV_ADMIN_TOKEN`, `WV_MASTER_PASS` | بيانات اعتماد الـ vault |
| `WV_KEY_GOOGLE`, `WV_KEY_OPENROUTER`, `WV_KEY_ANTHROPIC`, `WV_KEY_OPENAI` | مفاتيح API (مفصولة بفاصلة لمتعدّدة) |
| `WV_PROXY_TLS_ENABLED`, `WV_PROXY_TLS_CERT`, `WV_PROXY_TLS_KEY` | TLS الـ proxy |
| `WV_VAULT_TLS_ENABLED`, `WV_VAULT_TLS_CERT`, `WV_VAULT_TLS_KEY` | TLS الـ vault |
| `WV_PROXY_PLAIN_PORT` | مرافق HTTP على حلقة الاسترجاع (`0` للتعطيل) |
| `WV_VAULT_BOOTSTRAP_PORT` | مستمع bootstrap لـ CA (`0` للتعطيل) |
| `WV_OLLAMA_URL`, `WV_OLLAMA_KEEP_ALIVE`, `WV_OLLAMA_NUM_CTX` | ضبط Ollama |
| `WV_LMSTUDIO_URL`, `WV_VLLM_URL`, `WV_LLAMACPP_URL` | تجاوز الخادم الخلفي المحلي |
| `WV_TOKEN_SENTINEL_FALLBACK` | استبدال خفير "proxy-managed" على حلقة الاسترجاع |
| `WV_OAI_STREAM_FORWARD` | تمرير SSE حقيقي للخادم الخلفي المتوافق مع OpenAI |
| `WV_ANTHROPIC_FALLBACK_MODEL` | إعادة كتابة غير-Claude بالاشتراك على anthropic |

---

## الأوضاع

### Standalone (افتراضي)

يعمل الـ vault والـ proxy في العمليّة ذاتها. الأفضل لمضيف واحد يستضيف كلاً من المفاتيح والوكيل. يستمع على حلقة الاسترجاع فقط افتراضياً.

```bash
wall-vault start    # يشغّل كليهما
```

### Distributed

يعمل الـ vault على مضيف واحد (**vault host**) ويخزّن جميع المفاتيح؛ بينما تستوثق عدّة proxies على مضيفين آخرين، كل منها برمز خاص بكل عميل. مفيد عندما تحتاج عدّة آلات إلى المفاتيح ذاتها دون نسخها.

**Vault host:**

```bash
WV_VAULT_HOST=0.0.0.0 wall-vault vault
```

**كل proxy host:**

```bash
WV_VAULT_URL=http://<vault-host>:56243 \
WV_VAULT_TOKEN=<this-client-token> \
WV_PROXY_HOST=0.0.0.0 \
wall-vault proxy
```

مودال **Add Client** في اللوحة يصكّ رمزاً، ويسجّل نوع وكيل، ويلتقط الـ proxy تكوينه عبر SSE دون إعادة تشغيل.

---

## ملف yaml للإضافة (خادم خلفي جاهز)

أي خادم خلفي متوافق مع OpenAI يمكن إضافته كـ yaml تحت `~/.wall-vault/services/`. يلتقطه wall-vault عند البدء، يسجّله كخدمة قابلة للتوجيه، ويراه التوجيه + مجموعة كشف المتوافقة مع OAI + جسر Gemini-stream دون تغييرات في الكود.

```yaml
# ~/.wall-vault/services/llamacpp.yaml
id: llamacpp
name: llama.cpp
enabled: true
default_url: http://localhost:8080
endpoints:
  generate: /v1/chat/completions
  list_models: /v1/models
auth:
  type: none
request_format: openai
inline_no_think_for_qwen3: false   # اشترك إذا كان خادمك الخلفي يجرّد العلامة
```

طوبولوجيا hub (wall-vault واحد يواجه آخر) مدعومة عبر `tls_internal_ca: true`، `auth.type: bearer`، و`preserve_model_id: true`.

---

## البناء من المصدر

```bash
git clone https://github.com/sookmook/wall-vault
cd wall-vault
go build -o wall-vault .
./wall-vault setup
./wall-vault start
```

ترجمة عبر منصّات للمجموعة المدعومة بأكملها:

```bash
make build-all   # linux-amd64, linux-arm64, darwin-amd64, darwin-arm64, windows-amd64
```

تتبع الإصدارات `v{major}.{minor}.{patch}.{YYYYMMDD}.{HHmmss}`؛ يضبط `BASE_VERSION` في الـ Makefile البادئة.

### تنظيم المشروع

```
wall-vault/
├── main.go                     # توجيه CLI (start/proxy/vault/setup/cert/doctor)
├── cmd/
│   ├── setup/                  # معالج الإعداد التفاعلي
│   └── cert/                   # CA داخلي + مُصدِر شهادات TLS لكل مضيف
├── internal/
│   ├── config/                 # محمّل YAML + env، محمّل الإضافات
│   ├── proxy/                  # توجيه الطلبات، تدوير المفاتيح، محوّلات الصيغة
│   ├── vault/                  # متجر AES-GCM، اللوحة، وسيط SSE
│   ├── doctor/                 # فحص الصحّة + الإصلاح التلقائي
│   ├── hooks/                  # محرّكات أحداث أوامر shell
│   └── i18n/                   # سلاسل واجهة بـ 17 لغة
├── configs/services/           # إضافات yaml المرفقة (lmstudio, vllm, ollama, …)
└── docs/                       # MANUAL، مرجع API، 16 متغيّر محلي
```

---

## التوثيق

- [دليل المستخدم](docs/MANUAL.en.md) — التثبيت، اللوحة، الوكلاء، استكشاف الأخطاء
- [مرجع API](docs/API.en.md) — كل نقطة نهاية مع أشكال الطلب/الاستجابة
- [CHANGELOG](CHANGELOG.md)

---

## التقنيّة

- Go 1.25، ثنائي ساكن واحد
- [templ](https://templ.guide) للوحة التي تُرسَم على الخادم، [HTMX](https://htmx.org) للتحديثات الجزئيّة
- AES-GCM (مفتاح مشتقّ من PBKDF2) لتشفير المفاتيح عند الراحة
- Server-Sent Events للمزامنة المباشرة للتكوين بين الـ vault والـ proxies
- CA داخلي ذاتي التوقيع + شهادات لكل مضيف (لا حاجة إلى DNS عام / Let's Encrypt)

## الترخيص

GPL-3.0. انظر [LICENSE](LICENSE).

## المساهمة

طلبات السحب مرحّب بها. انظر [CONTRIBUTING.md](CONTRIBUTING.md). للتغييرات الكبيرة الرجاء فتح issue أوّلاً لمناقشة التصميم.
