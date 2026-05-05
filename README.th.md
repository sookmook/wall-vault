# wall-vault

> **คลังเก็บคีย์ API + AI proxy ในไบนารี Go เดียว**
> เก็บคีย์ไว้ภายในเครื่องด้วย AES-GCM, หมุนเวียนคีย์ระหว่างผู้ให้บริการต่าง ๆ, สำรองข้อมูลเมื่อมีหนึ่งบริการล้มเหลว และมาพร้อมแดชบอร์ดแบบเรียลไทม์

[![License: GPL v3](https://img.shields.io/badge/License-GPLv3-blue.svg)](LICENSE)
[![Go](https://img.shields.io/badge/Go-1.25-00ADD8.svg)](go.mod)

[English](README.md) · [한국어](README.ko.md) · [中文](README.zh.md) · [日本語](README.ja.md) · [Español](README.es.md) · [Français](README.fr.md) · [Deutsch](README.de.md) · [Português](README.pt.md) · [العربية](README.ar.md) · [हिन्दी](README.hi.md) · [Bahasa Indonesia](README.id.md) · **ภาษาไทย** · [Kiswahili](README.sw.md) · [Hausa](README.ha.md) · [नेपाली](README.ne.md) · [Монгол](README.mn.md) · [isiZulu](README.zu.md)

---

## คืออะไร

wall-vault อยู่ระหว่าง AI agent (OpenClaw, Claude Code, Cursor, Continue, สคริปต์ของคุณเอง) และผู้ให้บริการ AI บนคลาวด์หรือในเครื่องที่มันสื่อสารด้วย สองสิ่งในไบนารีเดียว:

- **Vault** — เก็บคีย์ API ที่เข้ารหัสไว้ขณะพักข้อมูล (AES-GCM พร้อมรหัสผ่านหลัก), หมุนเวียนคีย์, ติดตามการใช้งานและช่วงพักของแต่ละคีย์, กระจายการเปลี่ยนแปลงผ่าน SSE และให้บริการแดชบอร์ดเว็บที่ `:56243`
- **Proxy** — เปิดเผยปลายทางที่เข้ากันได้กับ Gemini, Anthropic และ OpenAI ที่ `:56244`, เลือกคีย์จาก vault, ส่งต่อไปยัง upstream ที่คุณตั้งค่าไว้ และสำรองไปยังผู้ให้บริการรายต่อไปเมื่อรายหนึ่งล้มเหลว

รองรับรูปแบบคำขอสี่แบบ (Gemini `:generateContent`, Anthropic `/v1/messages`, OpenAI `/v1/chat/completions` และ Ollama-native `/api/chat`) และ upstream ห้าหมวด:

| ผู้ให้บริการ | หมายเหตุ |
|----------|-------|
| Google Gemini | API ดั้งเดิม; การหมุนเวียนคีย์ต่อโปรเจกต์ |
| Anthropic | การส่งผ่าน `/v1/messages` ดั้งเดิม |
| OpenAI | `/v1/chat/completions` ดั้งเดิม |
| OpenRouter | โมเดล 340+ รายการ, สำรองอัตโนมัติเป็นรูปแบบ `:free` |
| Ollama / LM Studio / vLLM / llama.cpp / Jan / KoboldCpp / TabbyAPI / mlx-server / LiteLLM | backend ในเครื่องที่เข้ากันได้กับ OpenAI; ติดตั้งได้ทันทีผ่าน plugin yaml |

การเพิ่ม backend ใหม่ที่เข้ากันได้กับ OpenAI ใช้ไฟล์ yaml เดียวภายใต้ `~/.wall-vault/services/` — ไม่ต้องเปลี่ยนโค้ด

## ทำไมคุณอาจต้องการมัน

- คุณกำลังจัดการบริการ AI สามหรือสี่ตัวและต้องการ URL เดียวที่ agent สื่อสารด้วย
- คุณต้องการให้คีย์ free-tier ที่อยู่ในช่วงพักหลีกทางให้คีย์ถัดไปโดยไม่ทำให้เซสชันเสียหาย
- คุณต้องการให้คีย์ชุดเดียวกันขับเคลื่อนหลาย bot / IDE / สคริปต์ใน LAN เดียวกันโดยไม่ต้องคัดลอกข้อมูลรับรอง
- คุณต้องการแดชบอร์ด ไม่ใช่ environment variable สำหรับการแก้ไขคีย์ API
- คุณต้องการตัวเลือกแบบ local-first (Ollama / LM Studio) เมื่อขีดจำกัดของคลาวด์หมด

## เริ่มต้นใช้งานอย่างรวดเร็ว

### ติดตั้ง (Linux / macOS)

```bash
curl -fsSL https://raw.githubusercontent.com/sookmook/wall-vault/main/install.sh | sh
```

หรือดาวน์โหลดไบนารีที่สร้างไว้ล่วงหน้าโดยตรง:

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

### ติดตั้ง (Windows)

```powershell
Invoke-WebRequest -Uri `
  "https://github.com/sookmook/wall-vault/releases/latest/download/wall-vault-windows-amd64.exe" `
  -OutFile wall-vault.exe
```

### รันครั้งแรก

```bash
wall-vault setup    # ตัวช่วยแบบโต้ตอบ — เลือกพอร์ต, บริการ, admin token, รหัสผ่านหลัก
wall-vault start    # เปิดทั้ง vault และ proxy
```

เปิด `http://localhost:56243` (หรือ `https://...` เมื่อเปิด TLS แล้ว — ดูด้านล่าง) ในเบราว์เซอร์ แดชบอร์ดจะถามถึง admin token ที่ `setup` พิมพ์ออกมา จากตรงนั้นคุณสามารถเพิ่มคีย์ API, ลงทะเบียน client และสลับโมเดลได้โดยไม่ต้องรีสตาร์ท

---

## TLS (แนะนำ)

โดยค่าเริ่มต้น `wall-vault setup` จะเขียนการตั้งค่าโดยไม่มี TLS ดังนั้น listener ทั้งสองจึงตอบสนองด้วย HTTP ธรรมดา ตัวอย่าง URL ใน README นี้ใช้ `https://localhost:56244` เพราะ agent ส่วนใหญ่ (OpenClaw, Claude Code, Cursor) ต้องการปลายทางเดียวที่มี TLS อยู่ด้านหน้า ซึ่งจะไม่พังหากคุณย้าย proxy ไปอีก host ในภายหลัง เพื่อให้ตรงกับตัวอย่างเหล่านั้น เปิด TLS ครั้งเดียวด้วย CA ภายในที่มาพร้อมกัน:

```bash
# 1. สร้าง CA ภายในของ wall-vault (ครั้งเดียว, อยู่ใน ~/.wall-vault/ca.{crt,key})
wall-vault cert init

# 2. ออกใบรับรอง host สำหรับเครื่องนี้
#    SAN รวม hostname, localhost, 127.0.0.1 และ LAN IP ใด ๆ ที่ตรวจพบ
wall-vault cert issue $(hostname)

# 3. ไว้วางใจ CA ใน keychain ของระบบปฏิบัติการในเครื่อง
wall-vault cert install-trust

# 4. สลับ listener ไปใช้ TLS
export WV_PROXY_TLS_ENABLED=1
export WV_PROXY_TLS_CERT="$HOME/.wall-vault/$(hostname).crt"
export WV_PROXY_TLS_KEY="$HOME/.wall-vault/$(hostname).key"
export WV_VAULT_TLS_ENABLED=1
export WV_VAULT_TLS_CERT="$HOME/.wall-vault/$(hostname).crt"
export WV_VAULT_TLS_KEY="$HOME/.wall-vault/$(hostname).key"
wall-vault start
```

สำหรับเครื่องอื่นใน LAN ของคุณ: คัดลอก `~/.wall-vault/ca.crt` ไปและเรียกใช้ `wall-vault cert install-trust --ca <path>` ที่นั่น เมื่อ CA ได้รับความไว้วางใจในทุกที่ เครื่องทุกเครื่องในเครือข่ายจะสามารถเข้าถึง proxy ผ่าน `https://<host>:56244` ได้โดยไม่มีคำเตือนใบรับรอง

หากคุณต้องการอยู่กับ HTTP ธรรมดา ปล่อยการตั้งค่าไว้ตามเดิมและแทนที่ `https://` ด้วย `http://` ใน snippet ของ client ด้านล่าง ทั้งสอง scheme ใช้งานได้ ความแตกต่างคือพอร์ตใดตอบสนอง TLS handshake

**Loopback fallback** Client ใน host เดียวกันที่ไม่สามารถปฏิบัติตาม CA ของ wall-vault ได้ (โดยเฉพาะ runtime Node ที่มาพร้อมกับ OpenClaw ซึ่งเขียนทับ `NODE_EXTRA_CA_CERTS` เมื่อ spawn) เข้าถึง proxy ผ่าน companion HTTP ธรรมดาที่ใช้ loopback เท่านั้นที่ `127.0.0.1:56245` wall-vault เปิดใช้งานโดยอัตโนมัติเมื่อ TLS เปิด

---

## การเชื่อมต่อ client

ชี้ AI client ใด ๆ ไปที่ `https://<host>:56244` (หรือ `http://...` หาก TLS ปิด) Proxy ตอบสนองสี่รูปแบบ:

| รูปแบบ | Path | ตัวอย่าง client |
|--------|------|-----------------|
| Gemini | `/google/v1beta/models/<model>:generateContent` | OpenClaw, Gemini CLI, Antigravity |
| Anthropic | `/v1/messages` | Claude Code, Anthropic SDKs |
| OpenAI | `/v1/chat/completions` | Cursor, Continue, สคริปต์ที่กำหนดเอง, แอป LLM ส่วนใหญ่ |
| Ollama-native | `/api/chat` | Ollama client ที่ผ่านโดยตรง |

### Claude Code

```bash
export ANTHROPIC_BASE_URL=https://localhost:56244
export ANTHROPIC_API_KEY=<your-vault-client-token>
claude
```

เมื่อเครดิต Anthropic upstream หมด การส่งต่อจะสำรองไปยังผู้ให้บริการที่คุณตั้งไว้ใน `fallback_services` สำหรับ client นี้ เพื่อเลือกใช้การสำรองแบบไม่ใช่ Claude อย่างชัดเจน:

```yaml
proxy:
  anthropic_fallback_model: "claude-haiku-4-5-20251001"
```

(ค่าเริ่มต้นว่างทำให้การส่งต่อคืนค่าข้อผิดพลาด เพื่อให้การส่งต่อผิดทางปรากฏทันที)

### Cursor / Continue

ใน Cursor **Settings → AI → OpenAI API**:

```
Base URL:  https://localhost:56244
API Key:   <your-vault-client-token>
Model:     gemini-2.5-flash    # หรือโมเดลใด ๆ ที่ wall-vault รู้จัก
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

OpenClaw เป็นเฟรมเวิร์ก agent แบบ TUI ที่ wall-vault สร้างมาเพื่อให้บริการในตอนแรก โมดอล **Add Agent** ของแดชบอร์ดตั้งค่าประเภท agent เป็น `openclaw` (หรือ `nanoclaw`); จากนั้น wall-vault จะเขียน `~/.openclaw/openclaw.json` โดยตรง ซึ่งรวม URL ของผู้ให้บริการ, vault token และรายการโมเดล:

```bash
VAULT_CLIENT_ID=my-bot \
VAULT_URL=http://<vault-host>:56243 \
VAULT_TOKEN=<your-client-token> \
wall-vault proxy
```

### curl / สคริปต์

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

## การกำหนดค่า

`wall-vault setup` เขียน `./wall-vault.yaml` หรือ `~/.wall-vault/config.yaml` แก้ไขด้วยมือสำหรับฟิลด์ที่ตัวช่วยไม่ถาม

```yaml
mode: standalone     # standalone | distributed
lang: en             # en | ko | zh | ja | es | fr | de | pt | ar | hi | id | th | sw | ha | ne | mn | zu
theme: light         # light | dark | cherry | ocean | gold | autumn | winter

proxy:
  port: 56244
  host: ""           # ค่าเริ่มต้น: 127.0.0.1 standalone, 0.0.0.0 distributed
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
  plain_port: 56245              # companion HTTP แบบ loopback เท่านั้นเมื่อ TLS เปิด
  ollama_keep_alive: "30m"       # "-1" ไม่ unload, "0" unload ทันที
  ollama_num_ctx: 8192
  oai_stream_forward: false      # opt-in การส่งผ่าน SSE backend จริง
  anthropic_fallback_model: ""   # opt-in การเขียนใหม่แบบไม่ใช่ Claude บน anthropic dispatch

vault:
  port: 56243
  host: ""
  admin_token: ""
  admin_ip_whitelist: []
  master_password: ""            # รหัสผ่านการเข้ารหัสคีย์ AES-GCM
  data_dir: ~/.wall-vault/data
  services_dir: ~/.wall-vault/services
  tls:
    enabled: false
    cert_file: ""
    key_file: ""
  bootstrap_port: 56247          # listener HTTP ธรรมดาที่ให้บริการเฉพาะ ca.crt

doctor:
  interval: 5m
  auto_fix: true
  log_file: /tmp/wall-vault-doctor.log

hooks:
  on_model_change: ""            # คำสั่ง shell (env: SERVICE, MODEL)
  on_key_exhausted: ""
  on_service_down: ""
  on_doctor_fix: ""
  openclaw_socket: ""
```

### Environment variable

ทุกฟิลด์ YAML มี env override ที่ชนะไฟล์ ที่พบบ่อย:

| ตัวแปร | คำอธิบาย |
|----------|-------------|
| `WV_LANG`, `WV_THEME` | ภาษาและธีม |
| `WV_PROXY_PORT`, `WV_PROXY_HOST` | ที่อยู่ที่ Proxy ฟัง |
| `WV_VAULT_PORT`, `WV_VAULT_HOST` | ที่อยู่ที่ Vault ฟัง |
| `WV_VAULT_URL`, `WV_VAULT_TOKEN` | ปลายทางโหมด distributed |
| `WV_ADMIN_TOKEN`, `WV_MASTER_PASS` | ข้อมูลรับรอง Vault |
| `WV_KEY_GOOGLE`, `WV_KEY_OPENROUTER`, `WV_KEY_ANTHROPIC`, `WV_KEY_OPENAI` | คีย์ API (คั่นด้วยเครื่องหมายจุลภาคสำหรับหลายคีย์) |
| `WV_PROXY_TLS_ENABLED`, `WV_PROXY_TLS_CERT`, `WV_PROXY_TLS_KEY` | Proxy TLS |
| `WV_VAULT_TLS_ENABLED`, `WV_VAULT_TLS_CERT`, `WV_VAULT_TLS_KEY` | Vault TLS |
| `WV_PROXY_PLAIN_PORT` | Companion HTTP แบบ loopback (`0` เพื่อปิด) |
| `WV_VAULT_BOOTSTRAP_PORT` | Listener bootstrap CA (`0` เพื่อปิด) |
| `WV_OLLAMA_URL`, `WV_OLLAMA_KEEP_ALIVE`, `WV_OLLAMA_NUM_CTX` | การปรับแต่ง Ollama |
| `WV_LMSTUDIO_URL`, `WV_VLLM_URL`, `WV_LLAMACPP_URL` | การ override backend ในเครื่อง |
| `WV_TOKEN_SENTINEL_FALLBACK` | การแทนที่ sentinel "proxy-managed" บน loopback |
| `WV_OAI_STREAM_FORWARD` | การส่งผ่าน SSE backend จริงที่เข้ากันได้กับ OpenAI |
| `WV_ANTHROPIC_FALLBACK_MODEL` | Opt-in การเขียนใหม่แบบไม่ใช่ Claude บน anthropic |

---

## โหมด

### Standalone (ค่าเริ่มต้น)

Vault และ proxy ทำงานในกระบวนการเดียวกัน เหมาะสำหรับ host เดียวที่โฮสต์ทั้งคีย์และ agent ฟังเฉพาะ loopback โดยค่าเริ่มต้น

```bash
wall-vault start    # ทำงานทั้งสอง
```

### Distributed

Vault ทำงานบน host เดียว (**vault host**) และเก็บคีย์ทั้งหมด; proxy หลายตัวบน host อื่น ๆ ยืนยันตัวตนด้วย token ต่อ client มีประโยชน์เมื่อหลายเครื่องต้องใช้คีย์ชุดเดียวกันโดยไม่ต้องคัดลอกไปมา

**Vault host:**

```bash
WV_VAULT_HOST=0.0.0.0 wall-vault vault
```

**แต่ละ proxy host:**

```bash
WV_VAULT_URL=http://<vault-host>:56243 \
WV_VAULT_TOKEN=<this-client-token> \
WV_PROXY_HOST=0.0.0.0 \
wall-vault proxy
```

โมดอล **Add Client** ของแดชบอร์ดสร้าง token, ลงทะเบียนประเภท agent และ proxy จะรับการตั้งค่าผ่าน SSE โดยไม่ต้องรีสตาร์ท

---

## Plugin yaml (backend ติดตั้งทันที)

backend ใด ๆ ที่เข้ากันได้กับ OpenAI สามารถเพิ่มเป็น yaml ภายใต้ `~/.wall-vault/services/` wall-vault จะรับมันเมื่อเริ่มทำงาน, ลงทะเบียนเป็นบริการที่กำหนดเส้นทางได้ และการส่งต่อ + ชุดการตรวจจับที่เข้ากันได้กับ OAI + bridge สำหรับสตรีม Gemini ทั้งหมดจะเห็นมันโดยไม่ต้องเปลี่ยนโค้ด

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
inline_no_think_for_qwen3: false   # opt in หาก backend ของคุณตัด marker ออก
```

โทโพโลยี hub (wall-vault หนึ่งอยู่หน้าอีกอัน) รองรับผ่าน `tls_internal_ca: true`, `auth.type: bearer` และ `preserve_model_id: true`

---

## Build จากซอร์ส

```bash
git clone https://github.com/sookmook/wall-vault
cd wall-vault
go build -o wall-vault .
./wall-vault setup
./wall-vault start
```

Cross-compile สำหรับชุดที่รองรับทั้งหมด:

```bash
make build-all   # linux-amd64, linux-arm64, darwin-amd64, darwin-arm64, windows-amd64
```

เวอร์ชันเป็นไปตาม `v{major}.{minor}.{patch}.{YYYYMMDD}.{HHmmss}`; `BASE_VERSION` ใน Makefile ตั้งค่า prefix

### โครงสร้างโปรเจกต์

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

## เอกสารประกอบ

- [คู่มือผู้ใช้](docs/MANUAL.en.md) — การติดตั้ง, แดชบอร์ด, agent, การแก้ไขปัญหา
- [API reference](docs/API.en.md) — ทุกปลายทางพร้อมรูปแบบ request/response
- [CHANGELOG](CHANGELOG.md)

---

## เทคโนโลยีที่ใช้

- Go 1.25, ไบนารีสแตติกเดียว
- [templ](https://templ.guide) สำหรับแดชบอร์ดที่ render ฝั่งเซิร์ฟเวอร์, [HTMX](https://htmx.org) สำหรับการอัปเดตบางส่วน
- AES-GCM (คีย์จาก PBKDF2) สำหรับการเข้ารหัสคีย์ขณะพักข้อมูล
- Server-Sent Events สำหรับการซิงก์การตั้งค่าแบบสดระหว่าง vault และ proxy
- CA ภายในที่ลงนามด้วยตัวเอง + ใบรับรองต่อ host (ไม่ต้องใช้ DNS สาธารณะ / Let's Encrypt)

## License

GPL-3.0 ดู [LICENSE](LICENSE)

## การสนับสนุน

ยินดีรับ pull request ดู [CONTRIBUTING.md](CONTRIBUTING.md) สำหรับการเปลี่ยนแปลงที่ใหญ่กว่า โปรดเปิด issue ก่อนเพื่อหารือเกี่ยวกับการออกแบบ
