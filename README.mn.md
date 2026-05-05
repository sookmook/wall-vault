# wall-vault

> **API түлхүүрийн сан + AI proxy нэг Go binary дотор.**
> Түлхүүрүүдийг AES-GCM ашиглан дотооддоо хадгалж, нийлүүлэгчдийн хооронд эргэлддэг, нэг нь бүтэлгүйтсэн үед солигддог, мөн бодит цагийн хяналтын самбартай.

[![License: GPL v3](https://img.shields.io/badge/License-GPLv3-blue.svg)](LICENSE)
[![Go](https://img.shields.io/badge/Go-1.25-00ADD8.svg)](go.mod)

[English](README.md) · [한국어](README.ko.md) · [中文](README.zh.md) · [日本語](README.ja.md) · [Español](README.es.md) · [Français](README.fr.md) · [Deutsch](README.de.md) · [Português](README.pt.md) · [العربية](README.ar.md) · [हिन्दी](README.hi.md) · [Bahasa Indonesia](README.id.md) · [ภาษาไทย](README.th.md) · [Kiswahili](README.sw.md) · [Hausa](README.ha.md) · [नेपाली](README.ne.md) · **Монгол** · [isiZulu](README.zu.md)

---

## Энэ юу вэ

wall-vault нь AI агент (OpenClaw, Claude Code, Cursor, Continue, таны өөрийн скрипт) болон үүнтэй харьцдаг үүл эсвэл дотоод AI нийлүүлэгчдийн хооронд байрладаг. Нэг binary дотор хоёр зүйл:

- **Vault** — амарч буй үед шифрлэгдсэн API түлхүүрүүдийг хадгалдаг (AES-GCM мастер нууц үгтэй), эргэлддэг, түлхүүр тус бүрийн хэрэглээ ба амрах хугацааг хянадаг, өөрчлөлтийг SSE-ээр дамжуулдаг, мөн `:56243` дээр вэб хяналтын самбар үйлчлүүлдэг.
- **Proxy** — Gemini, Anthropic, OpenAI-нийцтэй endpoint-уудыг `:56244` дээр гаргадаг, vault-аас түлхүүр сонгодог, таны тохируулсан upstream руу илгээдэг, нэг нь бүтэлгүйтсэн үед дараагийн нийлүүлэгч рүү шилждэг.

Энэ нь хүсэлтийн дөрвөн хэлбэр (Gemini `:generateContent`, Anthropic `/v1/messages`, OpenAI `/v1/chat/completions`, Ollama-native `/api/chat`) болон upstream-ын таван ангиллыг дэмждэг:

| Нийлүүлэгч | Тэмдэглэл |
|----------|-------|
| Google Gemini | Үндсэн API; төсөл тус бүрийн түлхүүр эргэлт |
| Anthropic | Үндсэн `/v1/messages` дамжуулалт |
| OpenAI | Үндсэн `/v1/chat/completions` |
| OpenRouter | 340+ загвар, `:free` хувилбарууд руу автомат шилжилт |
| Ollama / LM Studio / vLLM / llama.cpp / Jan / KoboldCpp / TabbyAPI / mlx-server / LiteLLM | OpenAI-нийцтэй дотоод backend-үүд; plugin yaml-ээр шууд оруулах боломжтой |

OpenAI-нийцтэй шинэ backend нэмэх нь `~/.wall-vault/services/` дотор нэг yaml файл — код өөрчлөх шаардлагагүй.

## Та яагаад хүсэх ёстой вэ

- Та гурав, дөрвөн AI үйлчилгээг зэрэг ашиглаж байгаа бөгөөд агент тантай харьцах ганц URL хүсэж байна.
- Та амрах хугацаатай free-tier түлхүүр сесс эвдэхгүйгээр дараагийнхад зам гаргаж өгөхийг хүсэж байна.
- Та ижил түлхүүрүүдээр нэг LAN дээрх олон bot / IDE / скриптийг итгэмжлэлийг хуулахгүйгээр ажиллуулахыг хүсэж байна.
- Та environment variable биш, харин API түлхүүрийг засах хяналтын самбарыг хүсэж байна.
- Үүлний хязгаар дуусахад та дотоод-нэгдүгээрт сонголт (Ollama / LM Studio) хүсэж байна.

## Хурдан эхлүүлэх

### Суулгах (Linux / macOS)

```bash
curl -fsSL https://raw.githubusercontent.com/sookmook/wall-vault/main/install.sh | sh
```

Эсвэл урьдчилан бүтээгдсэн binary-ийг шууд татаж аваарай:

```bash
# Linux amd64
curl -L https://github.com/sookmook/wall-vault/releases/latest/download/wall-vault-linux-amd64 \
  -o wall-vault && chmod +x wall-vault

# macOS Apple Silicon
curl -L https://github.com/sookmook/wall-vault/releases/latest/download/wall-vault-darwin-arm64 \
  -o wall-vault && chmod +x wall-vault

# Linux arm64 (Raspberry Pi, ARM servers)
curl -L https://github.com/sookmook/wall-vault/releases/latest/download/wall-vault-linux-arm64 \
  -o wall-vault && chmod +x wall-vault
```

### Суулгах (Windows)

```powershell
Invoke-WebRequest -Uri `
  "https://github.com/sookmook/wall-vault/releases/latest/download/wall-vault-windows-amd64.exe" `
  -OutFile wall-vault.exe
```

### Анх удаа ажиллуулах

```bash
wall-vault setup    # харилцах wizard — порт, үйлчилгээ, admin token, мастер нууц үг сонгодог
wall-vault start    # vault болон proxy хоёуланг нь эхлүүлдэг
```

Хөтөч дээр `http://localhost:56243` (эсвэл TLS асаасны дараа `https://...` — доор үзнэ үү) нээгээрэй. Хяналтын самбар нь `setup`-ийн хэвлэсэн admin token-ийг асуудаг. Тэндээс та API түлхүүр нэмж, клиент бүртгэж, дахин эхлүүлэхгүйгээр загвар сольж болно.

---

## TLS (зөвлөмжтэй)

Анхдагчаар `wall-vault setup` нь TLS-гүй тохиргоо бичдэг тул хоёр сонсогч хоёулаа энгийн HTTP-ээр хариулдаг. Энэхүү README-ийн жишээ URL-ууд `https://localhost:56244`-ийг ашигладаг учир нь ихэнх агентууд (OpenClaw, Claude Code, Cursor) дараа нь proxy-г өөр host руу шилжүүлсэн ч эвдэрдэггүй TLS-ээр өмнөх ганц endpoint хүсдэг. Тэдгээр жишээтэй нийцүүлэхийн тулд багцлагдсан дотоод CA-тайгаар нэг удаа TLS-ийг асаа:

```bash
# 1. wall-vault дотоод CA үүсгэх (нэг удаа, ~/.wall-vault/ca.{crt,key} дотор оршдог)
wall-vault cert init

# 2. ЭНЭ машинд зориулсан host гэрчилгээ олгох
#    SAN-д hostname, localhost, 127.0.0.1, болон илрүүлсэн аливаа LAN IP багтдаг
wall-vault cert issue $(hostname)

# 3. Дотоод OS keychain-д CA-д итгэх
wall-vault cert install-trust

# 4. Сонсогчдыг TLS руу шилжүүлэх
export WV_PROXY_TLS_ENABLED=1
export WV_PROXY_TLS_CERT="$HOME/.wall-vault/$(hostname).crt"
export WV_PROXY_TLS_KEY="$HOME/.wall-vault/$(hostname).key"
export WV_VAULT_TLS_ENABLED=1
export WV_VAULT_TLS_CERT="$HOME/.wall-vault/$(hostname).crt"
export WV_VAULT_TLS_KEY="$HOME/.wall-vault/$(hostname).key"
wall-vault start
```

LAN дээрх өөр машинд: `~/.wall-vault/ca.crt`-ийг тийш хуулж аваад `wall-vault cert install-trust --ca <path>`-ийг тэнд ажиллуул. CA нь хаа сайгүй итгэлтэй болсны дараа сүлжээн дэх машин бүр гэрчилгээний анхааруулгагүйгээр `https://<host>:56244` дээгүүр proxy-д хүрч чадна.

Хэрэв та энгийн HTTP дээр үлдэхийг илүүд үзвэл тохиргоог хэвээр үлдээж, доорх client snippet-уудад `https://`-ийг `http://`-ээр солино. Хоёр scheme хоёул ажиллана; ялгаа нь аль порт TLS handshake-д хариу өгөх вэ гэдэгт байна.

**Loopback fallback.** wall-vault CA-д хүндэтгэлтэй хандаж чаддаггүй ижил-host клиентүүд (ялангуяа OpenClaw-ийн багцлагдсан Node runtime, spawn хийхэд `NODE_EXTRA_CA_CERTS`-ийг дахин бичдэг) `127.0.0.1:56245` дээрх loopback-only энгийн-HTTP хамтрагчаар дамжуулан proxy-д хүрдэг. TLS асаалттай үед wall-vault үүнийг автоматаар идэвхжүүлдэг.

---

## Клиентүүдийг холбох

Аливаа AI client-ийг `https://<host>:56244` руу чиглүүл (эсвэл TLS унтраалттай бол `http://...`). Proxy нь дөрвөн хэлбэрт хариу өгдөг:

| Формат | Зам | Жишээ клиент |
|--------|------|-----------------|
| Gemini | `/google/v1beta/models/<model>:generateContent` | OpenClaw, Gemini CLI, Antigravity |
| Anthropic | `/v1/messages` | Claude Code, Anthropic SDKs |
| OpenAI | `/v1/chat/completions` | Cursor, Continue, тусгай скриптүүд, ихэнх LLM апп |
| Ollama-native | `/api/chat` | Ollama клиентүүд дамжин өнгөрдөг |

### Claude Code

```bash
export ANTHROPIC_BASE_URL=https://localhost:56244
export ANTHROPIC_API_KEY=<your-vault-client-token>
claude
```

upstream Anthropic кредит дуусахад dispatch нь энэ client-д зориулж `fallback_services`-д тохируулсан нийлүүлэгчид рүү шилждэг. Не-Claude шилжилтийг тодорхой хүсэхийн тулд:

```yaml
proxy:
  anthropic_fallback_model: "claude-haiku-4-5-20251001"
```

(Хоосон анхдагч нь dispatch-ийг алдаа буцаахаар хийдэг тул буруу чиглүүлэлт шууд харагдана.)

### Cursor / Continue

Cursor-ийн **Settings → AI → OpenAI API**-д:

```
Base URL:  https://localhost:56244
API Key:   <your-vault-client-token>
Model:     gemini-2.5-flash    # эсвэл wall-vault-ын мэддэг ямар ч загвар
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

OpenClaw нь wall-vault-ыг үүсгэхдээ үндсэндээ үйлчилгээ үзүүлэхээр зориулан бүтээгдсэн TUI агентын хүрээ юм. Хяналтын самбарын **Add Agent** modal нь агентын төрлийг `openclaw` (эсвэл `nanoclaw`) болгон тогтоодог; дараа нь wall-vault нь нийлүүлэгчийн URL, vault token, загварын бүртгэлийг оруулсан `~/.openclaw/openclaw.json`-ыг шууд бичдэг:

```bash
VAULT_CLIENT_ID=my-bot \
VAULT_URL=http://<vault-host>:56243 \
VAULT_TOKEN=<your-client-token> \
wall-vault proxy
```

### curl / scripts

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

## Тохиргоо

`wall-vault setup` нь `./wall-vault.yaml` эсвэл `~/.wall-vault/config.yaml`-ыг бичдэг. Wizard-ын асуухгүй талбаруудыг гараар засаарай.

```yaml
mode: standalone     # standalone | distributed
lang: en             # en | ko | zh | ja | es | fr | de | pt | ar | hi | id | th | sw | ha | ne | mn | zu
theme: light         # light | dark | cherry | ocean | gold | autumn | winter

proxy:
  port: 56244
  host: ""           # анхдагч: 127.0.0.1 standalone, 0.0.0.0 distributed
  client_id: my-bot
  vault_url: ""      # distributed: http://vault-host:56243
  vault_token: ""    # distributed: client token
  tool_filter: strip_all   # strip_all | whitelist | passthrough
  allowed_tools: []
  services: [google, openrouter, ollama]
  timeout: 300s
  tls:
    enabled: false
    cert_file: ""
    key_file: ""
  plain_port: 56245              # TLS асаалттай үед loopback-only HTTP хамтрагч
  ollama_keep_alive: "30m"       # "-1" хэзээ ч буулгахгүй, "0" нэн даруй буулгана
  ollama_num_ctx: 8192
  oai_stream_forward: false      # opt-in бодит backend-ын SSE дамжуулалт
  anthropic_fallback_model: ""   # anthropic dispatch дээр opt-in не-Claude дахин бичих

vault:
  port: 56243
  host: ""
  admin_token: ""
  admin_ip_whitelist: []
  master_password: ""            # AES-GCM түлхүүр шифрлэлтийн нууц үг
  data_dir: ~/.wall-vault/data
  services_dir: ~/.wall-vault/services
  tls:
    enabled: false
    cert_file: ""
    key_file: ""
  bootstrap_port: 56247          # зөвхөн ca.crt-ийг үйлчлүүлдэг энгийн-HTTP listener

doctor:
  interval: 5m
  auto_fix: true
  log_file: /tmp/wall-vault-doctor.log

hooks:
  on_model_change: ""            # shell command (env: SERVICE, MODEL)
  on_key_exhausted: ""
  on_service_down: ""
  on_doctor_fix: ""
  openclaw_socket: ""
```

### Environment variables

YAML талбар бүр файлыг ялдаг env override-той. Түгээмэл нь:

| Хувьсагч | Тайлбар |
|----------|-------------|
| `WV_LANG`, `WV_THEME` | Хэл, загвар |
| `WV_PROXY_PORT`, `WV_PROXY_HOST` | Proxy сонсох хаяг |
| `WV_VAULT_PORT`, `WV_VAULT_HOST` | Vault сонсох хаяг |
| `WV_VAULT_URL`, `WV_VAULT_TOKEN` | Distributed-горимын endpoint |
| `WV_ADMIN_TOKEN`, `WV_MASTER_PASS` | Vault итгэмжлэл |
| `WV_KEY_GOOGLE`, `WV_KEY_OPENROUTER`, `WV_KEY_ANTHROPIC`, `WV_KEY_OPENAI` | API түлхүүр (олон бол таслалаар тусгаарлана) |
| `WV_PROXY_TLS_ENABLED`, `WV_PROXY_TLS_CERT`, `WV_PROXY_TLS_KEY` | Proxy TLS |
| `WV_VAULT_TLS_ENABLED`, `WV_VAULT_TLS_CERT`, `WV_VAULT_TLS_KEY` | Vault TLS |
| `WV_PROXY_PLAIN_PORT` | Loopback HTTP хамтрагч (унтраахад `0`) |
| `WV_VAULT_BOOTSTRAP_PORT` | CA bootstrap listener (унтраахад `0`) |
| `WV_OLLAMA_URL`, `WV_OLLAMA_KEEP_ALIVE`, `WV_OLLAMA_NUM_CTX` | Ollama тохируулга |
| `WV_LMSTUDIO_URL`, `WV_VLLM_URL`, `WV_LLAMACPP_URL` | Дотоод backend дарах |
| `WV_TOKEN_SENTINEL_FALLBACK` | Loopback "proxy-managed" sentinel солих |
| `WV_OAI_STREAM_FORWARD` | OpenAI-нийцтэй бодит backend SSE дамжуулалт |
| `WV_ANTHROPIC_FALLBACK_MODEL` | Anthropic дээр opt-in не-Claude дахин бичих |

---

## Горимууд

### Standalone (анхдагч)

Vault болон proxy ижил процесст ажилладаг. Түлхүүр болон агент хоёуланг нь хадгалдаг ганц host-ын хувьд хамгийн сайн. Анхдагчаар зөвхөн loopback дээр сонсдог.

```bash
wall-vault start    # хоёуланг нь ажиллуулна
```

### Distributed

Vault нэг host (**vault host**) дээр ажиллаж бүх түлхүүрүүдийг хадгалдаг; өөр host-ууд дээрх олон proxy тус бүр client тус бүрийн token-оор баталгаажуулдаг. Хэд хэдэн машинд түлхүүрүүдийг хуулахгүйгээр ижил түлхүүр хэрэгтэй үед хэрэгтэй.

**Vault host:**

```bash
WV_VAULT_HOST=0.0.0.0 wall-vault vault
```

**Proxy host бүр:**

```bash
WV_VAULT_URL=http://<vault-host>:56243 \
WV_VAULT_TOKEN=<this-client-token> \
WV_PROXY_HOST=0.0.0.0 \
wall-vault proxy
```

Хяналтын самбарын **Add Client** modal нь token үүсгэж, агентын төрлийг бүртгэх ба proxy нь дахин эхлүүлэхгүйгээр SSE-ээр өөрийн тохиргоог авдаг.

---

## Plugin yaml (drop-in backend)

OpenAI-нийцтэй ямар ч backend-ийг `~/.wall-vault/services/` дотор yaml хэлбэрээр нэмж болно. wall-vault эхлэхэд үүнийг сонгож, чиглүүлэх боломжтой үйлчилгээ болгон бүртгэдэг бөгөөд dispatch + OAI-compat illrүүлэх багц + Gemini-stream гүүр бүгд код өөрчлөхгүйгээр харуулдаг.

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
inline_no_think_for_qwen3: false   # таны backend marker-ийг устгадаг бол opt in хий
```

Hub topology (нэг wall-vault нөгөөгийнхөө өмнө байрладаг) нь `tls_internal_ca: true`, `auth.type: bearer`, `preserve_model_id: true`-ээр дэмжигддэг.

---

## Эх кодоос build хийх

```bash
git clone https://github.com/sookmook/wall-vault
cd wall-vault
go build -o wall-vault .
./wall-vault setup
./wall-vault start
```

Дэмжигдсэн бүх багцад зориулан cross-compile хий:

```bash
make build-all   # linux-amd64, linux-arm64, darwin-amd64, darwin-arm64, windows-amd64
```

Хувилбар `v{major}.{minor}.{patch}.{YYYYMMDD}.{HHmmss}` дагалддаг; Makefile дотрох `BASE_VERSION` нь prefix-ийг тогтоодог.

### Төслийн бүтэц

```
wall-vault/
├── main.go                     # CLI dispatch (start/proxy/vault/setup/cert/doctor)
├── cmd/
│   ├── setup/                  # interactive setup wizard
│   └── cert/                   # internal CA + per-host TLS certificate issuer
├── internal/
│   ├── config/                 # YAML + env loader, plugin loader
│   ├── proxy/                  # request dispatch, key rotation, format converters
│   ├── vault/                  # AES-GCM store, dashboard, SSE broker
│   ├── doctor/                 # health probe + auto-fix
│   ├── hooks/                  # shell-command event triggers
│   └── i18n/                   # 17-language UI strings
├── configs/services/           # bundled plugin yamls (lmstudio, vllm, ollama, …)
└── docs/                       # MANUAL, API reference, 16 locale variants
```

---

## Баримт бичиг

- [Хэрэглэгчийн гарын авлага](docs/MANUAL.en.md) — суулгалт, хяналтын самбар, агентууд, асуудлыг шийдвэрлэх
- [API лавлагаа](docs/API.en.md) — request/response хэлбэр бүхий endpoint бүр
- [CHANGELOG](CHANGELOG.md)

---

## Технологийн стек

- Go 1.25, ганц статик binary
- Сервер дээр render хийгдсэн хяналтын самбарын хувьд [templ](https://templ.guide), хэсэгчилсэн шинэчлэлт хийхэд [HTMX](https://htmx.org)
- Амарч буй түлхүүрийг шифрлэхэд AES-GCM (PBKDF2-аас гаргасан түлхүүр)
- vault болон proxy-ийн хооронд тохиргоог амьдаар нь синхрончлоход Server-Sent Events
- Өөрөө гарын үсэг зурсан дотоод CA + host тус бүрийн гэрчилгээ (нийтийн DNS / Let's Encrypt шаардлагагүй)

## Лиценз

GPL-3.0. [LICENSE](LICENSE)-г үзнэ үү.

## Хувь нэмэр оруулах

Pull request-ыг хүлээж авна. [CONTRIBUTING.md](CONTRIBUTING.md)-г үзнэ үү. Том өөрчлөлтийн хувьд дизайныг хэлэлцэхийн тулд эхлээд issue нээнэ үү.
