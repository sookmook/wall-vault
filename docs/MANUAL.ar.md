# دليل مستخدم wall-vault

[English](MANUAL.md) · [한국어](MANUAL.ko.md) · [中文](MANUAL.zh.md) · [日本語](MANUAL.ja.md) · [Español](MANUAL.es.md) · [Français](MANUAL.fr.md) · [Deutsch](MANUAL.de.md) · [Português](MANUAL.pt.md) · **العربية** · [हिन्दी](MANUAL.hi.md) · [Bahasa Indonesia](MANUAL.id.md) · [ภาษาไทย](MANUAL.th.md) · [Kiswahili](MANUAL.sw.md) · [Hausa](MANUAL.ha.md) · [नेपाली](MANUAL.ne.md) · [Монгол](MANUAL.mn.md) · [isiZulu](MANUAL.zu.md)

يغطي هذا الدليل تثبيت wall-vault وتكوينه وتشغيله. للحصول على نظرة عامة سريعة، راجع [README](../README.md). لمعرفة تفاصيل HTTP API، راجع [API reference](API.md).

## المحتويات

1. [ما يفعله wall-vault](#ما-يفعله-wall-vault)
2. [التثبيت](#التثبيت)
3. [التشغيل الأول مع معالج الإعداد](#التشغيل-الأول-مع-معالج-الإعداد)
4. [تفعيل TLS](#تفعيل-tls)
5. [تسجيل مفاتيح API](#تسجيل-مفاتيح-api)
6. [توصيل الوكلاء](#توصيل-الوكلاء)
7. [لوحة التحكم](#لوحة-التحكم)
8. [الوضع الموزع](#الوضع-الموزع)
9. [التشغيل التلقائي](#التشغيل-التلقائي)
10. [ملفات yaml للإضافات](#ملفات-yaml-للإضافات)
11. [Doctor](#doctor)
12. [Hooks](#hooks)
13. [متغيرات البيئة](#متغيرات-البيئة)
14. [حل المشاكل](#حل-المشاكل)

---

## ما يفعله wall-vault

wall-vault هو ملف Go ثنائي واحد يجمع خدمتين متعاونتين:

- **الخزنة (vault)** تخزن مفاتيح API مشفرة عند الراحة (AES-GCM بكلمة مرور رئيسية)، وتتتبع الاستخدام وفترات التهدئة لكل مفتاح، وتبث التغييرات عبر Server-Sent Events (SSE)، وتقدم لوحة تحكم ويب على `:56243` للمشغلين البشريين.
- **الوسيط (proxy)** يعرض نقاط نهاية متوافقة مع Gemini و Anthropic و OpenAI، وكذلك Ollama الأصلية على `:56244`. أي عميل ذكاء اصطناعي يشير إلى الوسيط فإنه يستخدم المفاتيح الموجودة في الخزنة — لا يراها العملاء أبداً. عندما يفشل أحد المزودين، ينتقل التوزيع إلى المزود التالي بالترتيب.

هذا مفيد عندما:

- لديك مفاتيح لعدة مزودين وتريد عنوان URL واحدًا يتحدث إليه الوكيل.
- تريد أن يتنحى مفتاح المستوى المجاني الموجود في فترة تهدئة دون كسر الجلسة.
- تريد أن تشغل نفس المفاتيح روبوتات متعددة أو IDEs أو سكربتات على نفس LAN دون نسخ الاعتمادات.
- تريد لوحة تحكم، وليس متغيرات بيئة، لتحرير المفاتيح وتبديل النماذج.
- تريد بديلاً محلياً (Ollama، LM Studio، vLLM) عند نفاد حدود السحابة.

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

## التثبيت

### سطر واحد لـ Linux / macOS

```bash
curl -fsSL https://raw.githubusercontent.com/sookmook/wall-vault/main/install.sh | sh
```

يكتشف السكربت تلقائياً نظام التشغيل والمعمارية، وينزل الملف الثنائي الصحيح إلى `~/.local/bin/wall-vault`، ويجعله قابلاً للتنفيذ. إذا لم يكن `~/.local/bin` موجوداً في `PATH`، أضفه:

```bash
export PATH="$HOME/.local/bin:$PATH"
```

### التنزيل اليدوي

تُنشر الملفات الثنائية المُجمَّعة مسبقاً في كل إصدار على `https://github.com/sookmook/wall-vault/releases`.

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

### البناء من المصدر

يتطلب Go 1.25 أو أحدث.

```bash
git clone https://github.com/sookmook/wall-vault
cd wall-vault
go build -o wall-vault .
```

`make build-all` يقوم بالتجميع المتقاطع لجميع المنصات الخمس المدعومة. تظهر الملفات الثنائية في `bin/`.

---

## التشغيل الأول مع معالج الإعداد

```bash
wall-vault setup
```

يطلب منك المعالج، بالترتيب:

1. **اللغة** — يختار واحدة من 17 لغة لواجهة المستخدم. يتم اكتشافها تلقائياً من `$LANG`؛ ومع ذلك يقدم المعالج قائمة.
2. **المظهر (Theme)** — `light` (الافتراضي)، `dark`، `cherry`، `ocean`، `gold`، `autumn`، `winter`. للتجميل فقط.
3. **الوضع** — `standalone` (مضيف واحد، الافتراضي) أو `distributed` (الخزنة على مضيف واحد، الوسطاء على آخرين).
4. **اسم الروبوت** — معرف `client_id` على شكل slug حر. تستخدم الخزنة هذا لتحديد نطاق التكوين لكل عميل (تجاوزات النموذج، سلاسل التراجع).
5. **منفذ الوسيط** — الافتراضي `56244`.
6. **منفذ الخزنة** — الافتراضي `56243` (للوضع المستقل فقط).
7. **اختيار الخدمة** — y/N لكل من: Google Gemini، OpenRouter، Anthropic، OpenAI، Ollama، LM Studio، vLLM. الاختيارات المتعددة مقبولة؛ كل واحدة تكتب تلميح متغير البيئة الخاص بها في النهاية.
8. **مرشح الأدوات** — `strip_all` (الافتراضي؛ يحظر جميع تعريفات الأدوات الواردة لأغراض الأمان) أو `passthrough` (السماح بمرور أي أداة).
9. **رمز المسؤول** — اتركه فارغاً للتوليد التلقائي. تتطلب لوحة التحكم هذا الرمز لتسجيل الدخول.
10. **كلمة المرور الرئيسية** — اتركها فارغة لعدم التشفير (غير مُوصى به)؛ عيِّن قيمة لتشفير AES-GCM لمخزن المفاتيح عند الراحة.
11. **مسار الحفظ** — يكون افتراضياً `wall-vault.yaml` في المجلد الحالي. ينظر المُحمِّل أيضاً في `~/.wall-vault/config.yaml`.

بعد الحفظ، يقوم المعالج بتشغيل `doctor.FixTrust` بحيث يحصل أي وكيل مثبت محلياً (OpenClaw، Claude Code، Cline) تلقائياً على CA الداخلي لـ wall-vault مضافاً إلى مخزن الثقة الخاص به. إذا لم يكن أي وكيل من هذا القبيل مثبتاً، تطبع الخطوة `SKIP` ولا تكتب شيئاً.

ثم ابدأ تشغيل الملف الثنائي:

```bash
wall-vault start
```

`start` يشغل كلاً من الخزنة والوسيط في عملية واحدة (الوضع المستقل). للوضع الموزع استخدم `wall-vault vault` على مضيف الخزنة و `wall-vault proxy` على كل مضيف وسيط.

افتح `http://localhost:56243` في المتصفح. سجِّل الدخول برمز المسؤول الذي طبعه المعالج.

---

## تفعيل TLS

تترك الإعدادات الافتراضية للمعالج كلا المُستمعَين على HTTP عادي. تعمل معظم الوكلاء (OpenClaw، Claude Code، Cursor) بشكل أفضل مقابل نقطة نهاية HTTPS واحدة، لذا يُوصى باستخدام TLS في أي توزيع يمتد إلى أكثر من الجهاز المحلي.

يأتي wall-vault مع CA داخلي خاص به، لذا لا تحتاج إلى اسم DNS عام أو Let's Encrypt.

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

لتمديد الثقة إلى أجهزة LAN أخرى، انسخ `~/.wall-vault/ca.crt` وقم بتشغيل `wall-vault cert install-trust --ca <path>` على كل واحدة. تعرض الخزنة أيضاً `ca.crt` عبر مُستمع HTTP عادي صغير على `:56247` (**منفذ التمهيد**) لحالة catch-22 حيث يحتاج عميل جديد إلى CA للتحدث عبر HTTPS.

### رفيق HTTP loopback

بعض الوكلاء — لا سيما وقت تشغيل Node المرفق مع OpenClaw — تعيد كتابة `NODE_EXTRA_CA_CERTS` عند إنشاء العملية، مما يسقط أي تلميح CA قدمه المشغل. لا يمكنها احترام CA الخاص بـ wall-vault من داخل الـ daemon، حتى بعد `cert install-trust`. يحل wall-vault هذه المشكلة من خلال ربط **مُستمع HTTP عادي خاص بـ loopback فقط** على `127.0.0.1:56245` كلما تم تمكين TLS. يصل العملاء على نفس المضيف إلى الوسيط من خلال هذا المنفذ دون TLS على الإطلاق؛ بينما تستمر عملاء LAN في استخدام مُستمع TLS.

عطّله بـ `WV_PROXY_PLAIN_PORT=0` إذا لم تكن بحاجة إليه.

### `wall-vault cert list`

يعرض كل شهادة تحت `~/.wall-vault/` مع الموضوع ونافذة الصلاحية وأسماء SANs.

```
$ wall-vault cert list
ca.crt          subject=wall-vault internal CA   not-after=2036-05-05
hostname.crt    subject=hostname                 not-after=2031-05-05   SAN=hostname,localhost,127.0.0.1,192.168.…
```

---

## تسجيل مفاتيح API

طريقتان: لوحة التحكم، أو متغيرات البيئة.

### لوحة التحكم (مُوصى بها)

1. سجل الدخول على `https://localhost:56243` برمز المسؤول.
2. انقر **+ API key** في بطاقة المفاتيح.
3. اختر خدمة (Google، OpenRouter، Anthropic، OpenAI، …).
4. الصق المفتاح. احفظ.

مفاتيح متعددة لكل خدمة جيدة؛ يقوم الوسيط بالتدوير بينها ويتخطى تلك التي وصلت إلى فترة تهدئة لكل مفتاح.

### متغيرات البيئة (تمهيد لمرة واحدة)

```bash
export WV_KEY_GOOGLE="AIzaSyA1...,AIzaSyB2...,AIzaSyC3..."   # comma-separated
export WV_KEY_OPENROUTER="sk-or-v1-…"
export WV_KEY_ANTHROPIC="sk-ant-…"
export WV_KEY_OPENAI="sk-…"
wall-vault start
```

تكتب المفاتيح المقدمة بهذه الطريقة في المخزن المشفر عند أول تشغيل. تقرأ التشغيلات اللاحقة هذه من القرص؛ يمكنك إلغاء تعيين متغيرات البيئة بعد التشغيل الأول.

### فترات التهدئة والتدوير

كل مكالمة ناجحة تزيد `usage_count` للمفتاح وتحدّث `last_used`. عند HTTP 429 / 402 / 403، يضع الوسيط المفتاح في **فترة تهدئة** (الافتراضيات: 60 دقيقة لـ 429، و24 ساعة لـ 402، و12 ساعة لـ 403). ينتقي التوزيع التالي مفتاحاً مختلفاً لتلك الخدمة. عندما تكون جميع مفاتيح الخدمة في فترة تهدئة، يتخطى الوسيط تلك الخدمة بسرعة بالكامل ويحاول المزود التالي في سلسلة التراجع.

تكون فترات التهدئة مرئية لكل مفتاح في لوحة التحكم مع عدّ تنازلي.

---

## توصيل الوكلاء

### OpenClaw

OpenClaw هو العميل المستهدف الأصلي. استخدم النافذة المنبثقة **+ Add agent** في لوحة التحكم:

- عيّن **Agent type** إلى `openclaw` أو `nanoclaw`.
- عيّن **Work directory** — لـ OpenClaw يتم ملؤه تلقائياً كـ `~/.openclaw`.
- اختر **preferred service** واختياريا **model override**.
- انقر **Apply**. يكتب wall-vault `~/.openclaw/openclaw.json` مباشرة (عناوين URL للمزود، رمز الخزنة، إدخالات النموذج).

عندما تغير النموذج من لوحة التحكم، يلتقط OpenClaw التغيير عبر SSE في غضون 1-3 ثوانٍ — دون إعادة تشغيل.

### Claude Code

```bash
export ANTHROPIC_BASE_URL=https://localhost:56244
export ANTHROPIC_API_KEY=<your-vault-client-token>
claude
```

عندما ينفد رصيد Anthropic الأعلى، يتراجع التوزيع إلى أي خدمات مدرجة في `fallback_services` لهذا العميل. افتراضياً، يُرجع معرف نموذج غير Claude مرسل إلى توزيع anthropic خطأ بحيث يظهر التوجيه الخاطئ على الفور. اختر الكتابة التلقائية:

```yaml
proxy:
  anthropic_fallback_model: "claude-haiku-4-5-20251001"
```

### Cursor

في **Settings → AI → OpenAI API** الخاص بـ Cursor:

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

### HTTP مخصص

```bash
curl -X POST https://localhost:56244/v1/chat/completions \
  -H "Authorization: Bearer <your-vault-client-token>" \
  -H "Content-Type: application/json" \
  -d '{
    "model": "gemini-2.5-flash",
    "messages": [{"role": "user", "content": "hello"}]
  }'
```

تقبل نقطة النهاية نفسها التدفق (`"stream": true`) عند تعيين `proxy.oai_stream_forward: true`.

---

## لوحة التحكم

`https://localhost:56243`. خمس بطاقات على شبكة الصفحة الرئيسية:

- **Keys** — كل مفتاح API، مجموع حسب الخدمة. أضف، حرر، احذف؛ شاهد الاستخدام وفترة التهدئة.
- **Services** — Google / OpenRouter / Anthropic / OpenAI / Ollama / LM Studio / vLLM / llama.cpp، بالإضافة إلى أي ملف yaml إضافي في `~/.wall-vault/services/`. عيّن `default_model`، `allowed_models`، عنوان URL الأساسي، تبديل التفكير لكل خدمة.
- **Clients (agents)** — كل عميل مسجل (روبوت OpenClaw، جلسة Claude Code، نسخة Cursor، …). عيّن الخدمة المفضلة، تجاوز النموذج، سلسلة التراجع.
- **Proxies** — كل وسيط تم مصادقته على هذه الخزنة. الحالة الحية (online/offline)، آخر ظهور، النموذج الحالي.
- **Settings** — رمز المسؤول، تدوير كلمة المرور الرئيسية، المظهر، اللغة.

كل بطاقة لها slideover للتحرير (الجانب الأيمن). النقر خارجاً أو `Esc` يغلقه. تُدفع التغييرات إلى جميع الوسطاء المتصلين عبر SSE في غضون ثوانٍ.

يحمل **التذييل** مؤشر SSE (أخضر = متصل، برتقالي = إعادة الاتصال، رمادي = غير متصل) وإصدار البناء الحي.

---

## الوضع الموزع

عندما يكون لديك عدة أجهزة كلها تحتاج إلى نفس المفاتيح، شغّل الخزنة على مضيف واحد والوسطاء على كل من البقية.

### مضيف الخزنة

```bash
WV_VAULT_HOST=0.0.0.0 \
WV_ADMIN_TOKEN=<admin> \
WV_MASTER_PASS=<master> \
wall-vault vault
```

لوحة التحكم متاحة الآن على `https://<vault-host>:56243`. أضف وكيلاً لكل وسيط بعيد في بطاقة **Clients**؛ كل واحد يصدر `vault_token` فريداً.

### مضيفو الوسيط

```bash
WV_VAULT_URL=http://<vault-host>:56243 \
WV_VAULT_TOKEN=<that-client-token> \
WV_PROXY_HOST=0.0.0.0 \
wall-vault proxy
```

يصادق الوسيط على الخزنة، ويفتح تدفق SSE، ويطبق أي تكوين يستلمه (الخدمة المفضلة، تجاوز النموذج، سلسلة التراجع). تصل تعديلات الخزنة اللاحقة في ثوانٍ بدون إعادة تشغيل.

لتثبيتات تمتد عبر LAN، فعّل TLS على مضيف الخزنة (`WV_VAULT_TLS_ENABLED=1` + متغيرات بيئة الشهادة/المفتاح) وقم بتشغيل كل مضيف وسيط من خلال نفس خطوة `wall-vault cert install-trust` بحيث يتم الوثوق بمكالمات HTTPS الخاصة بالوسيط إلى الخزنة.

---

## التشغيل التلقائي

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

للخزنة على نفس المضيف، اكتب `wall-vault-vault.service` موازياً. للوضع المستقل، وحدة واحدة تستدعي `wall-vault start` كافية.

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

استخدم `nssm` لتغليف `wall-vault.exe start` كخدمة Windows، أو إدخال `schtasks` يعمل عند تسجيل دخول المستخدم.

---

## ملفات yaml للإضافات

يمكن إضافة أي backend متوافق مع OpenAI بدون تغييرات في الكود عن طريق إسقاط ملف yaml تحت `~/.wall-vault/services/`. يُحمل wall-vault الملف عند بدء التشغيل ويسجل الخدمة للتوزيع، ومجموعة كشف OAI-compat، وجسر تدفق Gemini.

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

تأتي المجموعة المرفقة في `configs/services/` (lmstudio، vllm، llamacpp، tgwui، localai، jan، koboldcpp، tabbyapi، mlx-server، litellm-proxy، ollama، google، openrouter) معطلة افتراضياً. انسخ الذي تريده إلى `~/.wall-vault/services/`، عيّن `enabled: true`، وأعد التشغيل.

---

## Doctor

`wall-vault doctor` يُجري فحص صحة لمرة واحدة عبر التثبيت بأكمله:

```
✓ vault listener  (https://localhost:56243)
✓ proxy listener  (https://localhost:56244)
✓ master password set
⚠ Google: 2 keys, all on cooldown
✓ Anthropic: 1 key healthy
✗ Ollama: not reachable at http://localhost:11434
```

كل سطر هو واحد من:

- `✓` — صحي
- `⚠` — متدهور لكن يعمل (مفتاح واحد في فترة تهدئة، حصة منخفضة، إلخ.)
- `✗` — معطل
- `SKIP` — غير مكوّن / غير قابل للتطبيق على هذا المضيف

يقوم وضع daemon ثانٍ بتشغيل نفس الفحص كل `doctor.interval` (الافتراضي 5 دقائق) ويكتب النتائج إلى `doctor.log_file` (الافتراضي `/tmp/wall-vault-doctor.log`). عندما يكون `doctor.auto_fix` صحيحاً، يحاول أيضاً إصلاح الانحراف الشائع (تكوين OpenClaw قديم، ثقة TLS مفقودة، خدمات قابلة لإعادة التشغيل).

شغل عملية واحدة من لوحة التحكم عبر بطاقة **Doctor** أو `wall-vault doctor`.

---

## Hooks

شغّل أمر shell عند الأحداث الرئيسية:

```yaml
hooks:
  on_model_change:   "logger 'wall-vault: $SERVICE/$MODEL'"
  on_key_exhausted:  "notify-send 'wall-vault' '$SERVICE keys all on cooldown'"
  on_service_down:   "/usr/local/bin/page-oncall.sh $SERVICE '$ERROR'"
  on_doctor_fix:     "echo \"$AGENT: $LEVEL $MSG\" >> ~/wall-vault.audit.log"
  openclaw_socket:   ""    # if set, OpenClaw TUI receives events over this Unix socket
```

يحصل كل hook على متغيرات بيئة خاصة بالحدث (`SERVICE`، `MODEL`، `ERROR`، `AGENT`، `LEVEL`، `MSG`). تعمل الـ hooks بشكل غير متزامن مع مهلة 5 ثوانٍ — لا يحجب الوسيط أبداً عن hook بطيء.

---

## متغيرات البيئة

| المتغير | حقل YAML |
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

كل متغير بيئة، عند تعيينه، يتغلب على ملف YAML.

---

## حل المشاكل

### `connection refused` على `:56244`

إما أن الوسيط لا يعمل أو أنه مرتبط بمضيف مختلف. تحقق من:

```bash
ss -lnp | grep 56244
systemctl --user status wall-vault-proxy   # Linux
launchctl list | grep wall-vault           # macOS
```

إذا كان يعمل على منفذ مختلف، فإن تكوينك يحتوي على تجاوز لـ `proxy.port` — تحقق من `~/.wall-vault/config.yaml`.

### `x509: certificate signed by unknown authority`

العميل لا يثق بـ CA الداخلي لـ wall-vault. شغل `wall-vault cert install-trust` على جهاز العميل. للوكلاء التي يتجاهل وقت تشغيلها مخزن ثقة OS (مثل Node مع `NODE_EXTRA_CA_CERTS` مُرمَّز بشكل صلب)، استخدم رفيق HTTP loopback على `127.0.0.1:56245` (نفس المضيف فقط) أو عيّن `WV_PROXY_TLS_ENABLED=0` للرجوع إلى HTTP عادي.

### `token not registered with vault`

`Authorization: Bearer <token>` للعميل لا يطابق أي عميل مسجل. تحقق من الرمز تحت **Clients** في لوحة التحكم. إذا نسخت رمزاً حرفياً مثل `proxy-managed`، `dummy`، أو `""` من تكوين قديم، استبدله برمز العميل الحقيقي.

### `Anthropic dispatch needs a Claude model id`

السلوك الافتراضي اعتباراً من v0.2.63: يُرجع معرف نموذج غير Claude مرسل إلى توزيع anthropic خطأ. إما إصلاح التوجيه (لا ترسل `gemini-2.5-flash` إلى anthropic) أو الموافقة على الكتابة التلقائية عبر `proxy.anthropic_fallback_model`.

### `unknown service: <id>`

رأى التوزيع معرف خدمة لم تطالب به أي ملف yaml إضافي. تحقق من:

```bash
ls ~/.wall-vault/services/        # any plugin yaml present?
cat ~/.wall-vault/services/<id>.yaml | grep enabled
```

إذا كان ملف yaml موجوداً ولكنه `enabled: false`، اقلبه. إذا كان مفقوداً تماماً، انسخ من `configs/services/` في شجرة المصدر.

### استجابة فارغة على نموذج تفكير

`qwen3.6`، `deepseek-r1`، وعائلة GPT-`o1` تُصدر أحياناً `reasoning_content` فقط وتترك `content` فارغاً. اعتباراً من v0.2.63 يتراجع wall-vault إلى نص التفكير تلقائياً — إذا كنت لا تزال ترى استجابات فارغة، فإن backend لا يُرجع أي حقل. تحقق من سجلات upstream.

لـ LM Studio مع qwen3 على وجه التحديد، عيّن `inline_no_think_for_qwen3: true` في ملف yaml الإضافي بحيث يتم تعطيل التفكير inline. lmstudio.yaml و ollama.yaml المدمجة تفعل هذا بالفعل.

### تظهر لوحة التحكم "all keys on cooldown" لكنني للتو أضفت واحداً

المفتاح الجديد صحي لكن مسار التوزيع قد لا يزال في فترة تهدئة لمفتاح أقدم. جرب طلباً جديداً — يقوم الوسيط بالتدوير لكل مكالمة، وسيتم انتقاء مفتاح صحي تالياً.

### الخزنة لن تفتح بكلمة المرور الرئيسية

كلمة مرور خاطئة. لا يوجد استرداد — wall-vault لا يشحن باب خلفي عمداً. إذا فقدت بصدق كلمة المرور الرئيسية، فالطريق الوحيد هو حذف `~/.wall-vault/data/vault.json`، وإعادة التشغيل بكلمة مرور جديدة، وإعادة إضافة المفاتيح.

### ضربت حدود OpenRouter للمستوى المجاني

عيّن `proxy.services` ليتضمن `openrouter` وأضف مفتاح OpenRouter واحد على الأقل. يتراجع الوسيط تلقائياً من النموذج المدفوع إلى متغيره `:free` عندما يُرجع المسار المدفوع 402 / 429.

### `journalctl --user -u wall-vault-proxy` فارغ

سجلات systemd `--user` تذهب إلى journal للمستخدم الذي يشغلها. إذا بدأت الوحدة كـ `root` أو عبر `sudo`، فإن journal موجود في النسخة الخاصة بالنظام بدلاً من ذلك — جرب `journalctl -u wall-vault-proxy` بدون `--user`.

---

## المزيد

- مرجع HTTP API — انظر [API.md](API.md)
- المصدر — `https://github.com/sookmook/wall-vault`
- تقارير الأخطاء / طلبات الميزات — GitHub Issues
- سجل الإصدارات — [CHANGELOG.md](../CHANGELOG.md)
