# wall-vault User Manual

[English](MANUAL.md) · [한국어](MANUAL.ko.md) · [中文](MANUAL.zh.md) · [日本語](MANUAL.ja.md) · [Español](MANUAL.es.md) · [Français](MANUAL.fr.md) · [Deutsch](MANUAL.de.md) · [Português](MANUAL.pt.md) · [العربية](MANUAL.ar.md) · [हिन्दी](MANUAL.hi.md) · [Bahasa Indonesia](MANUAL.id.md) · ภาษาไทย · [Kiswahili](MANUAL.sw.md) · [Hausa](MANUAL.ha.md) · [नेपाली](MANUAL.ne.md) · [Монгол](MANUAL.mn.md) · [isiZulu](MANUAL.zu.md)

คู่มือนี้ครอบคลุมการติดตั้ง การกำหนดค่า และการใช้งาน wall-vault สำหรับภาพรวมโดยย่อโปรดดูที่ [README](../README.md) สำหรับรายละเอียด HTTP API ดูที่ [API reference](API.md)

## สารบัญ

1. [wall-vault ทำอะไร](#wall-vault-ทำอะไร)
2. [การติดตั้ง](#การติดตั้ง)
3. [การรันครั้งแรกด้วย setup wizard](#การรันครั้งแรกด้วย-setup-wizard)
4. [การเปิดใช้งาน TLS](#การเปิดใช้งาน-tls)
5. [การลงทะเบียน API key](#การลงทะเบียน-api-key)
6. [การเชื่อมต่อ agent](#การเชื่อมต่อ-agent)
7. [Dashboard](#dashboard)
8. [โหมด distributed](#โหมด-distributed)
9. [การเริ่มต้นอัตโนมัติ](#การเริ่มต้นอัตโนมัติ)
10. [Plugin yamls](#plugin-yamls)
11. [Doctor](#doctor)
12. [Hooks](#hooks)
13. [ตัวแปรสภาพแวดล้อม](#ตัวแปรสภาพแวดล้อม)
14. [การแก้ไขปัญหา](#การแก้ไขปัญหา)

---

## wall-vault ทำอะไร

wall-vault เป็น Go binary ตัวเดียวที่รวมบริการสองอย่างที่ทำงานร่วมกัน:

- **vault** เก็บ API key ที่เข้ารหัสไว้ขณะพัก (AES-GCM พร้อมรหัสผ่านหลัก) ติดตามการใช้งานและ cooldown ต่อ key เผยแพร่การเปลี่ยนแปลงผ่าน Server-Sent Events (SSE) และให้บริการ web dashboard ที่ `:56243` สำหรับผู้ดูแลที่เป็นมนุษย์
- **proxy** เปิด endpoint ที่เข้ากันได้กับ Gemini, Anthropic, OpenAI และ Ollama-native ที่ `:56244` AI client ใดที่ชี้ไปที่ proxy ก็จะใช้ key ที่อยู่ใน vault โดยที่ client จะไม่เห็น key เหล่านั้นเลย เมื่อ upstream หนึ่งล้มเหลว dispatch จะ fall back ไปยัง provider ถัดไปตามลำดับ

สิ่งนี้มีประโยชน์เมื่อ:

- คุณมี key สำหรับ provider หลายราย และต้องการ URL เดียวที่ agent คุยด้วย
- คุณต้องการให้ key ระดับ free-tier ที่อยู่ใน cooldown ถอยออกไปโดยไม่ทำให้ session พัง
- คุณต้องการให้ key ชุดเดียวกันขับเคลื่อนหลาย bot, IDE หรือ script บน LAN เดียวกันโดยไม่ต้องคัดลอกข้อมูลรับรอง
- คุณต้องการ dashboard แทนตัวแปรสภาพแวดล้อมในการแก้ไข key และเปลี่ยนโมเดล
- คุณต้องการ fallback ในเครื่อง (Ollama, LM Studio, vLLM) เมื่อโควต้า cloud หมด

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

## การติดตั้ง

### Linux / macOS one-liner

```bash
curl -fsSL https://raw.githubusercontent.com/sookmook/wall-vault/main/install.sh | sh
```

สคริปต์จะตรวจหา OS และสถาปัตยกรรมโดยอัตโนมัติ ดาวน์โหลด binary ที่ถูกต้องไปยัง `~/.local/bin/wall-vault` และทำให้สามารถ execute ได้ หาก `~/.local/bin` ยังไม่อยู่บน `PATH` ของคุณ ให้เพิ่ม:

```bash
export PATH="$HOME/.local/bin:$PATH"
```

### ดาวน์โหลดด้วยตนเอง

Binary ที่ build ไว้ล่วงหน้าถูกเผยแพร่ใน release ทุกครั้งที่ `https://github.com/sookmook/wall-vault/releases`

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

### Build จาก source

ต้องการ Go 1.25 หรือใหม่กว่า

```bash
git clone https://github.com/sookmook/wall-vault
cd wall-vault
go build -o wall-vault .
```

`make build-all` จะ cross-compile ไปยังทั้งห้าแพลตฟอร์มที่รองรับ Binary จะอยู่ใน `bin/`

---

## การรันครั้งแรกด้วย setup wizard

```bash
wall-vault setup
```

Wizard จะถามคุณตามลำดับ:

1. **ภาษา** — เลือกหนึ่งใน 17 locale ของ UI ตรวจหาโดยอัตโนมัติจาก `$LANG` แต่ wizard ก็เสนอรายการให้เลือกอยู่ดี
2. **ธีม** — `light` (ค่าเริ่มต้น), `dark`, `cherry`, `ocean`, `gold`, `autumn`, `winter` เป็นเรื่องของรูปลักษณ์เท่านั้น
3. **โหมด** — `standalone` (host เดียว ค่าเริ่มต้น) หรือ `distributed` (vault บน host หนึ่ง proxy บน host อื่นๆ)
4. **ชื่อ bot** — slug ของ `client_id` แบบอิสระ vault ใช้สิ่งนี้เพื่อจำกัดขอบเขตการตั้งค่าต่อ client (model override, fallback chain)
5. **Proxy port** — ค่าเริ่มต้น `56244`
6. **Vault port** — ค่าเริ่มต้น `56243` (เฉพาะ standalone)
7. **การเลือกบริการ** — y/N สำหรับแต่ละรายการ: Google Gemini, OpenRouter, Anthropic, OpenAI, Ollama, LM Studio, vLLM เลือกหลายอย่างได้ แต่ละอย่างจะเขียน hint ตัวแปรสภาพแวดล้อมไว้ตอนท้าย
8. **Tool filter** — `strip_all` (ค่าเริ่มต้น บล็อกการกำหนด tool ที่เข้ามาทั้งหมดเพื่อความปลอดภัย) หรือ `passthrough` (ปล่อย tool ใดก็ได้ผ่าน)
9. **Admin token** — เว้นว่างเพื่อสร้างอัตโนมัติ Dashboard ต้องใช้ token นี้เพื่อเข้าสู่ระบบ
10. **รหัสผ่านหลัก** — เว้นว่างเพื่อไม่เข้ารหัส (ไม่แนะนำ) ตั้งค่าเพื่อเข้ารหัส key store ขณะพักด้วย AES-GCM
11. **เส้นทางบันทึก** — ค่าเริ่มต้นเป็น `wall-vault.yaml` ในไดเรกทอรีปัจจุบัน Loader ยังมองหาที่ `~/.wall-vault/config.yaml` ด้วย

หลังจากบันทึก wizard จะรัน `doctor.FixTrust` เพื่อให้ agent ที่ติดตั้งในเครื่อง (OpenClaw, Claude Code, Cline) ได้รับการเพิ่ม CA ภายในของ wall-vault ลงใน trust store โดยอัตโนมัติ หากไม่มี agent ดังกล่าวติดตั้งอยู่ ขั้นตอนนี้จะพิมพ์ `SKIP` และไม่เขียนอะไร

จากนั้นเริ่ม binary:

```bash
wall-vault start
```

`start` รัน vault และ proxy ทั้งคู่ใน process เดียว (โหมด standalone) สำหรับโหมด distributed ใช้ `wall-vault vault` บน vault host และ `wall-vault proxy` บน proxy host แต่ละตัว

เปิด `http://localhost:56243` ใน browser เข้าสู่ระบบด้วย admin token ที่ wizard พิมพ์ออกมา

---

## การเปิดใช้งาน TLS

ค่าเริ่มต้นของ wizard ทำให้ listener ทั้งสองอยู่บน HTTP ธรรมดา agent ส่วนใหญ่ (OpenClaw, Claude Code, Cursor) ทำงานได้ดีกว่ากับ HTTPS endpoint เดียว ดังนั้น TLS จึงแนะนำในทุกการ deploy ที่ครอบคลุมเกินกว่าเครื่องเดียว

wall-vault มาพร้อม CA ภายในของตัวเอง ดังนั้นคุณไม่ต้องการชื่อ DNS สาธารณะหรือ Let's Encrypt

```bash
# 1. สร้าง CA ภายใน — เขียนไปที่ ~/.wall-vault/ca.{crt,key}
#    CA ใช้ได้ 10 ปีโดยค่าเริ่มต้น override ด้วย --ca-years
wall-vault cert init

# 2. ออกใบรับรอง host Subject Alternative Names จะรวมโดยอัตโนมัติ:
#       hostname, "localhost", "127.0.0.1", และ LAN IP ที่ไม่ใช่ loopback ใดๆ ที่ตรวจพบ
#    Override issuer dir ด้วย --dir, validity ด้วย --host-years
wall-vault cert issue $(hostname)

# 3. Trust CA ใน OS keychain ของเครื่องนี้
#    Linux: เขียนไปที่ /etc/ssl/certs/ ผ่าน update-ca-certificates (ต้องการ sudo)
#    macOS: เพิ่มไปยัง System keychain ผ่าน security add-trusted-cert (ต้องการ sudo)
#    Windows: import ไปยัง CurrentUser\Root ผ่าน certutil (ไม่ต้องการ admin)
wall-vault cert install-trust

# 4. เปิดใช้งาน TLS บน listener ทั้งคู่
export WV_PROXY_TLS_ENABLED=1
export WV_PROXY_TLS_CERT="$HOME/.wall-vault/$(hostname).crt"
export WV_PROXY_TLS_KEY="$HOME/.wall-vault/$(hostname).key"
export WV_VAULT_TLS_ENABLED=1
export WV_VAULT_TLS_CERT="$HOME/.wall-vault/$(hostname).crt"
export WV_VAULT_TLS_KEY="$HOME/.wall-vault/$(hostname).key"

wall-vault start
```

หากต้องการขยาย trust ไปยังเครื่อง LAN อื่นๆ ให้คัดลอก `~/.wall-vault/ca.crt` ไปและรัน `wall-vault cert install-trust --ca <path>` บนแต่ละเครื่อง vault ยังเปิด `ca.crt` ผ่าน plain-HTTP listener เล็กๆ ที่ `:56247` (**bootstrap port**) สำหรับกรณี catch-22 ที่ client ใหม่ต้องการ CA เพื่อพูดคุย HTTPS

### Loopback HTTP companion

agent บางตัว — โดยเฉพาะ Node runtime ที่มาพร้อมกับ OpenClaw — เขียนทับ `NODE_EXTRA_CA_CERTS` เมื่อ spawn process โดยทิ้ง CA hint ที่ผู้ดูแลให้ไว้ พวกมันไม่สามารถยอมรับ CA ของ wall-vault จากภายใน daemon ได้ แม้หลังจาก `cert install-trust` แล้ว wall-vault แก้ปัญหานี้โดยการ bind **plain-HTTP listener เฉพาะ loopback** เพิ่มเติมที่ `127.0.0.1:56245` เมื่อใดก็ตามที่ TLS เปิดใช้งาน Client บน host เดียวกันเข้าถึง proxy ผ่าน port นั้นโดยไม่มี TLS เลย LAN client ยังคงใช้ TLS listener

ปิดใช้งานด้วย `WV_PROXY_PLAIN_PORT=0` หากคุณไม่ต้องการ

### `wall-vault cert list`

แสดงทุก cert ภายใต้ `~/.wall-vault/` พร้อม subject, validity window และ SAN

```
$ wall-vault cert list
ca.crt          subject=wall-vault internal CA   not-after=2036-05-05
hostname.crt    subject=hostname                 not-after=2031-05-05   SAN=hostname,localhost,127.0.0.1,192.168.…
```

---

## การลงทะเบียน API key

มีสองวิธี: dashboard หรือตัวแปรสภาพแวดล้อม

### Dashboard (แนะนำ)

1. เข้าสู่ระบบที่ `https://localhost:56243` ด้วย admin token
2. คลิก **+ API key** ในการ์ด keys
3. เลือกบริการ (Google, OpenRouter, Anthropic, OpenAI, …)
4. วาง key บันทึก

หลาย key ต่อบริการก็ใช้ได้ proxy จะ round-robin ระหว่าง key เหล่านั้นและข้าม key ที่ติด cooldown ต่อ key

### ตัวแปรสภาพแวดล้อม (bootstrap แบบครั้งเดียว)

```bash
export WV_KEY_GOOGLE="AIzaSyA1...,AIzaSyB2...,AIzaSyC3..."   # comma-separated
export WV_KEY_OPENROUTER="sk-or-v1-…"
export WV_KEY_ANTHROPIC="sk-ant-…"
export WV_KEY_OPENAI="sk-…"
wall-vault start
```

Key ที่ให้มาด้วยวิธีนี้จะถูกเขียนเข้า store ที่เข้ารหัสในการเปิดใช้งานครั้งแรก การเริ่มครั้งต่อๆ ไปจะอ่านจากดิสก์ คุณสามารถ unset ตัวแปรสภาพแวดล้อมได้หลังจากการรันครั้งแรก

### Cooldown และการหมุนเวียน

ทุกการเรียกที่สำเร็จจะเพิ่ม `usage_count` ของ key และรีเฟรช `last_used` เมื่อได้ HTTP 429 / 402 / 403, proxy จะใส่ key ใน **cooldown** (ค่าเริ่มต้น: 60 นาทีสำหรับ 429, 24 ชั่วโมงสำหรับ 402, 12 ชั่วโมงสำหรับ 403) Dispatch ครั้งถัดไปจะเลือก key อื่นสำหรับบริการนั้น เมื่อ key ทั้งหมดสำหรับบริการอยู่ใน cooldown, proxy จะข้ามบริการนั้นไปอย่างรวดเร็วและลอง provider ถัดไปใน fallback chain

Cooldown สามารถมองเห็นได้ต่อ key ใน dashboard พร้อมเวลานับถอยหลัง

---

## การเชื่อมต่อ agent

### OpenClaw

OpenClaw เป็น client เป้าหมายดั้งเดิม ใช้ modal **+ Add agent** ของ dashboard:

- ตั้ง **Agent type** เป็น `openclaw` หรือ `nanoclaw`
- ตั้ง **Work directory** — สำหรับ OpenClaw จะเติมโดยอัตโนมัติเป็น `~/.openclaw`
- เลือก **preferred service** และอาจเลือก **model override**
- คลิก **Apply** wall-vault จะเขียน `~/.openclaw/openclaw.json` โดยตรง (provider URL, vault token, รายการโมเดล)

เมื่อคุณเปลี่ยนโมเดลจาก dashboard, OpenClaw จะรับการเปลี่ยนแปลงผ่าน SSE ภายใน 1–3 วินาที โดยไม่ต้อง restart

### Claude Code

```bash
export ANTHROPIC_BASE_URL=https://localhost:56244
export ANTHROPIC_API_KEY=<your-vault-client-token>
claude
```

เมื่อเครดิต Anthropic upstream หมด, dispatch จะ fall back ไปยังบริการที่อยู่ใน `fallback_services` ของ client นี้ โดยค่าเริ่มต้น model id ที่ไม่ใช่ Claude ที่ส่งไปยัง anthropic dispatch จะคืนค่า error เพื่อให้การ misroute ปรากฏขึ้นทันที เลือก opt in เพื่อเขียนทับโดยอัตโนมัติ:

```yaml
proxy:
  anthropic_fallback_model: "claude-haiku-4-5-20251001"
```

### Cursor

ใน Cursor **Settings → AI → OpenAI API**:

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

### Custom HTTP

```bash
curl -X POST https://localhost:56244/v1/chat/completions \
  -H "Authorization: Bearer <your-vault-client-token>" \
  -H "Content-Type: application/json" \
  -d '{
    "model": "gemini-2.5-flash",
    "messages": [{"role": "user", "content": "hello"}]
  }'
```

Endpoint เดียวกันรองรับ streaming (`"stream": true`) เมื่อตั้งค่า `proxy.oai_stream_forward: true`

---

## Dashboard

`https://localhost:56243` มีห้าการ์ดบน home grid:

- **Keys** — ทุก API key จัดกลุ่มตามบริการ เพิ่ม แก้ไข ลบ ดูการใช้งานและ cooldown
- **Services** — Google / OpenRouter / Anthropic / OpenAI / Ollama / LM Studio / vLLM / llama.cpp รวมถึง plugin yaml ใดๆ ใน `~/.wall-vault/services/` ตั้งค่า `default_model` ต่อบริการ, `allowed_models`, base URL, การเปิด/ปิด reasoning
- **Clients (agents)** — ทุก client ที่ลงทะเบียน (OpenClaw bot, Claude Code session, Cursor instance, …) กำหนด preferred service, model override, fallback chain
- **Proxies** — ทุก proxy ที่ได้รับการ authenticate กับ vault นี้ สถานะ live (online/offline), ครั้งสุดท้ายที่เห็น, model ปัจจุบัน
- **Settings** — admin token, การหมุนเวียนรหัสผ่านหลัก, ธีม, ภาษา

แต่ละการ์ดมี slideover แก้ไข (ด้านขวา) คลิกข้างนอกหรือ `Esc` เพื่อปิด การเปลี่ยนแปลงจะถูก push ไปยัง proxy ที่เชื่อมต่อทั้งหมดผ่าน SSE ภายในไม่กี่วินาที

**Footer** แสดงตัวบ่งชี้ SSE (เขียว = เชื่อมต่อ, ส้ม = กำลังเชื่อมต่อใหม่, เทา = ตัดการเชื่อมต่อ) และเวอร์ชัน build แบบ live

---

## โหมด distributed

เมื่อคุณมีหลายเครื่องที่ทั้งหมดต้องการ key เดียวกัน ให้รัน vault บน host หนึ่งและ proxy บนแต่ละเครื่องที่เหลือ

### Vault host

```bash
WV_VAULT_HOST=0.0.0.0 \
WV_ADMIN_TOKEN=<admin> \
WV_MASTER_PASS=<master> \
wall-vault vault
```

ตอนนี้ dashboard เข้าถึงได้ที่ `https://<vault-host>:56243` เพิ่ม agent สำหรับแต่ละ proxy ระยะไกลในการ์ด **Clients** แต่ละตัวจะสร้าง `vault_token` ที่ไม่ซ้ำกัน

### Proxy host

```bash
WV_VAULT_URL=http://<vault-host>:56243 \
WV_VAULT_TOKEN=<that-client-token> \
WV_PROXY_HOST=0.0.0.0 \
wall-vault proxy
```

Proxy authenticate กับ vault, เปิด SSE stream และนำการตั้งค่าใดๆ ที่ได้รับมาใช้ (preferred service, model override, fallback chain) การแก้ไข vault ครั้งต่อๆ ไปจะมาถึงในไม่กี่วินาทีโดยไม่ต้อง restart

สำหรับการติดตั้งที่ครอบคลุม LAN ให้เปิดใช้งาน TLS บน vault host (`WV_VAULT_TLS_ENABLED=1` + ตัวแปรสภาพแวดล้อม cert/key) และรันแต่ละ proxy host ผ่านขั้นตอน `wall-vault cert install-trust` เดียวกัน เพื่อให้การเรียก HTTPS ของ proxy ไปยัง vault ได้รับ trust

---

## การเริ่มต้นอัตโนมัติ

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

สำหรับ vault บน host เดียวกัน ให้เขียน `wall-vault-vault.service` คู่ขนาน สำหรับโหมด standalone หนึ่ง unit ที่เรียก `wall-vault start` ก็เพียงพอ

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

ใช้ `nssm` เพื่อ wrap `wall-vault.exe start` เป็น Windows service หรือ entry `schtasks` ที่รันเมื่อ user logon

---

## Plugin yamls

backend ที่เข้ากันได้กับ OpenAI ใดๆ สามารถเพิ่มได้โดยไม่ต้องเปลี่ยนแปลงโค้ดด้วยการวาง yaml ภายใต้ `~/.wall-vault/services/` wall-vault จะโหลดมันเมื่อ startup และลงทะเบียนบริการสำหรับ dispatch, ชุดการตรวจหา OAI-compat และ Gemini-stream bridge

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

ชุดที่มาพร้อมใน `configs/services/` (lmstudio, vllm, llamacpp, tgwui, localai, jan, koboldcpp, tabbyapi, mlx-server, litellm-proxy, ollama, google, openrouter) มาในสถานะ disabled โดยค่าเริ่มต้น คัดลอกตัวที่คุณต้องการไปยัง `~/.wall-vault/services/` ตั้ง `enabled: true` แล้ว restart

---

## Doctor

`wall-vault doctor` รัน health probe ครั้งเดียวทั่วทั้งการติดตั้ง:

```
✓ vault listener  (https://localhost:56243)
✓ proxy listener  (https://localhost:56244)
✓ master password set
⚠ Google: 2 keys, all on cooldown
✓ Anthropic: 1 key healthy
✗ Ollama: not reachable at http://localhost:11434
```

แต่ละบรรทัดเป็นหนึ่งใน:

- `✓` — แข็งแรง
- `⚠` — เสื่อมแต่ยังทำงานได้ (key หนึ่งอยู่ใน cooldown, โควต้าต่ำ ฯลฯ)
- `✗` — เสีย
- `SKIP` — ไม่ได้กำหนดค่า / ไม่เกี่ยวข้องบน host นี้

โหมด daemon ที่สองรัน probe เดียวกันทุก `doctor.interval` (ค่าเริ่มต้น 5 นาที) และเขียนผลลัพธ์ไปยัง `doctor.log_file` (ค่าเริ่มต้น `/tmp/wall-vault-doctor.log`) เมื่อ `doctor.auto_fix` เป็น true มันยังพยายามซ่อม drift ทั่วไป (config OpenClaw ที่เก่า, TLS trust ที่หายไป, บริการที่ restart ได้)

Trigger ครั้งเดียวจาก dashboard ผ่านการ์ด **Doctor** หรือ `wall-vault doctor`

---

## Hooks

รันคำสั่ง shell บน event ของ key:

```yaml
hooks:
  on_model_change:   "logger 'wall-vault: $SERVICE/$MODEL'"
  on_key_exhausted:  "notify-send 'wall-vault' '$SERVICE keys all on cooldown'"
  on_service_down:   "/usr/local/bin/page-oncall.sh $SERVICE '$ERROR'"
  on_doctor_fix:     "echo \"$AGENT: $LEVEL $MSG\" >> ~/wall-vault.audit.log"
  openclaw_socket:   ""    # if set, OpenClaw TUI receives events over this Unix socket
```

แต่ละ hook ได้รับตัวแปรสภาพแวดล้อมเฉพาะ event (`SERVICE`, `MODEL`, `ERROR`, `AGENT`, `LEVEL`, `MSG`) Hook รันแบบ async พร้อม timeout 5 วินาที proxy จะไม่ block บน hook ที่ช้า

---

## ตัวแปรสภาพแวดล้อม

| Variable | YAML field |
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

ทุกตัวแปรสภาพแวดล้อมเมื่อตั้งค่าจะชนะไฟล์ YAML

---

## การแก้ไขปัญหา

### `connection refused` บน `:56244`

อาจเป็นเพราะ proxy ไม่ได้รัน หรือ bind ไปยัง host อื่น ตรวจสอบ:

```bash
ss -lnp | grep 56244
systemctl --user status wall-vault-proxy   # Linux
launchctl list | grep wall-vault           # macOS
```

หากกำลังรันบน port อื่น แสดงว่า config ของคุณมี `proxy.port` ถูก override — ตรวจสอบ `~/.wall-vault/config.yaml`

### `x509: certificate signed by unknown authority`

Client ไม่ trust CA ภายในของ wall-vault รัน `wall-vault cert install-trust` บนเครื่อง client สำหรับ agent ที่ runtime ไม่สนใจ OS trust store (เช่น Node ที่มี `NODE_EXTRA_CA_CERTS` ฮาร์ดโค้ด) ใช้ loopback HTTP companion ที่ `127.0.0.1:56245` (host เดียวกันเท่านั้น) หรือตั้ง `WV_PROXY_TLS_ENABLED=0` เพื่อ fall back เป็น HTTP ธรรมดา

### `token not registered with vault`

`Authorization: Bearer <token>` ของ client ไม่ตรงกับ client ที่ลงทะเบียนใดๆ ตรวจสอบ token ภายใต้ **Clients** ใน dashboard หากคุณคัดลอก token literal เช่น `proxy-managed`, `dummy` หรือ `""` จาก config เก่า ให้แทนที่ด้วย client token จริง

### `Anthropic dispatch needs a Claude model id`

พฤติกรรมเริ่มต้น ณ v0.2.63: model id ที่ไม่ใช่ Claude ที่ส่งไปยัง anthropic dispatch จะคืนค่า error ให้แก้ไข routing (อย่าส่ง `gemini-2.5-flash` ไปยัง anthropic) หรือเลือก opt in เพื่อเขียนทับโดยอัตโนมัติผ่าน `proxy.anthropic_fallback_model`

### `unknown service: <id>`

Dispatch เห็น service id ที่ไม่มี plugin yaml อ้างสิทธิ์ ตรวจสอบ:

```bash
ls ~/.wall-vault/services/        # any plugin yaml present?
cat ~/.wall-vault/services/<id>.yaml | grep enabled
```

หาก yaml มีอยู่แต่ `enabled: false` ให้ flip มัน หากหายไปทั้งหมด ให้คัดลอกจาก `configs/services/` ใน source tree

### Response ว่างเปล่าบนโมเดล reasoning

`qwen3.6`, `deepseek-r1` และตระกูล GPT-`o1` บางครั้ง emit เฉพาะ `reasoning_content` และทิ้ง `content` ว่างเปล่า ณ v0.2.63 wall-vault fall back ไปยังข้อความ reasoning โดยอัตโนมัติ — หากคุณยังเห็น response ว่าง backend กำลังคืนค่าทั้งสอง field ไม่กลับมาเลย ตรวจสอบ log ของ upstream

สำหรับ LM Studio กับ qwen3 โดยเฉพาะ ตั้ง `inline_no_think_for_qwen3: true` ใน plugin yaml เพื่อให้ reasoning ถูกปิดใช้งานแบบ inline lmstudio.yaml และ ollama.yaml ในตัวทำสิ่งนี้แล้ว

### Dashboard แสดง "all keys on cooldown" แต่ฉันเพิ่งเพิ่มหนึ่ง

Key ใหม่แข็งแรงแต่ dispatch path อาจยังอยู่ใน cooldown สำหรับ key เก่า ลอง request ใหม่ — proxy round-robin ต่อการเรียก และ key ที่แข็งแรงจะถูกเลือกถัดไป

### Vault จะไม่ปลดล็อกด้วยรหัสผ่านหลัก

รหัสผ่านผิด ไม่มีการกู้คืน — wall-vault จงใจไม่มี backdoor หากคุณสูญเสียรหัสผ่านหลักจริงๆ ทางเดียวคือลบ `~/.wall-vault/data/vault.json` restart ด้วยรหัสผ่านใหม่ และเพิ่ม key อีกครั้ง

### ขีดจำกัด OpenRouter free-tier ถึง

ตั้ง `proxy.services` ให้รวม `openrouter` และเพิ่ม OpenRouter key อย่างน้อยหนึ่งตัว Proxy จะ auto-fall-back จากโมเดลแบบเสียเงินไปยัง variant `:free` เมื่อ path ที่จ่ายเงินคืนค่า 402 / 429

### `journalctl --user -u wall-vault-proxy` ว่างเปล่า

systemd `--user` log ไปที่ journal ของ user ที่รันมัน หากคุณเริ่ม unit เป็น `root` หรือผ่าน `sudo`, journal อยู่ใน system instance แทน — ลอง `journalctl -u wall-vault-proxy` โดยไม่มี `--user`

---

## เพิ่มเติม

- HTTP API reference — ดูที่ [API.md](API.md)
- Source — `https://github.com/sookmook/wall-vault`
- รายงานบั๊ก / ขอฟีเจอร์ — GitHub Issues
- ประวัติ release — [CHANGELOG.md](../CHANGELOG.md)
