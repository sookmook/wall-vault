# คู่มือผู้ใช้ wall-vault
*(อัปเดตล่าสุด: 2026-04-05 — v0.1.21)*

---

## สารบัญ

1. [wall-vault คืออะไร?](#wall-vault-คืออะไร)
2. [การติดตั้ง](#การติดตั้ง)
3. [เริ่มต้นใช้งานครั้งแรก (setup wizard)](#เริ่มต้นใช้งานครั้งแรก)
4. [การลงทะเบียน API Key](#การลงทะเบียน-api-key)
5. [วิธีใช้งาน Proxy](#วิธีใช้งาน-proxy)
6. [แดชบอร์ดคลังกุญแจ](#แดชบอร์ดคลังกุญแจ)
7. [โหมดกระจาย (Multi-Bot)](#โหมดกระจาย-multi-bot)
8. [การตั้งค่าเริ่มต้นอัตโนมัติ](#การตั้งค่าเริ่มต้นอัตโนมัติ)
9. [Doctor ผู้ช่วยวินิจฉัย](#doctor-ผู้ช่วยวินิจฉัย)
10. [ตัวแปรสภาพแวดล้อม (Environment Variables)](#ตัวแปรสภาพแวดล้อม)
11. [การแก้ปัญหา](#การแก้ปัญหา)

---

## wall-vault คืออะไร?

**wall-vault = ตัวแทน AI (Proxy) + คลังกุญแจ API สำหรับ OpenClaw**

ในการใช้บริการ AI คุณจำเป็นต้องมี **API Key** (บัตรผ่านดิจิทัล) ซึ่งเป็นหลักฐานที่ยืนยันว่า "บุคคลนี้มีสิทธิ์ใช้บริการนี้" อย่างไรก็ตาม บัตรผ่านนี้มีจำนวนครั้งที่ใช้ได้จำกัดต่อวัน และหากจัดการไม่ดีก็อาจมีความเสี่ยงที่จะถูกเปิดเผยได้

wall-vault เก็บรักษาบัตรผ่านเหล่านี้ไว้ในคลังที่ปลอดภัย และทำหน้าที่เป็น **ตัวแทน (Proxy)** ระหว่าง OpenClaw กับบริการ AI กล่าวง่ายๆ คือ OpenClaw เชื่อมต่อกับ wall-vault เพียงจุดเดียว แล้ว wall-vault จะจัดการเรื่องที่ซับซ้อนทั้งหมดให้โดยอัตโนมัติ

ปัญหาที่ wall-vault ช่วยแก้ไข:

- **หมุนเวียน API Key อัตโนมัติ**: เมื่อ Key หนึ่งถึงขีดจำกัดการใช้งานหรือถูกพักชั่วคราว (cooldown) ระบบจะเปลี่ยนไปใช้ Key ถัดไปอย่างเงียบๆ OpenClaw ทำงานต่อเนื่องโดยไม่สะดุด
- **สลับบริการอัตโนมัติ (Fallback)**: หาก Google ไม่ตอบสนอง จะเปลี่ยนไปใช้ OpenRouter โดยอัตโนมัติ และหากยังไม่ได้ผล จะเปลี่ยนไปใช้ Ollama (AI ในเครื่องของคุณ) เซสชันจะไม่ขาดหาย เมื่อบริการเดิมกลับมาใช้งานได้ ระบบจะสลับกลับโดยอัตโนมัติตั้งแต่คำขอถัดไป (v0.1.18+)
- **ซิงค์แบบเรียลไทม์ (SSE)**: เมื่อคุณเปลี่ยนโมเดลในแดชบอร์ดคลังกุญแจ การเปลี่ยนแปลงจะปรากฏบนหน้าจอ OpenClaw ภายใน 1–3 วินาที SSE (Server-Sent Events) คือเทคโนโลยีที่เซิร์ฟเวอร์ส่งการเปลี่ยนแปลงไปยัง Client แบบเรียลไทม์
- **การแจ้งเตือนแบบเรียลไทม์**: เหตุการณ์ต่างๆ เช่น Key หมดหรือบริการล่ม จะแสดงทันทีที่ส่วนล่างของ TUI (หน้าจอเทอร์มินัล) ของ OpenClaw

> 💡 **Claude Code, Cursor, VS Code** สามารถเชื่อมต่อและใช้งานได้เช่นกัน แต่จุดประสงค์หลักของ wall-vault คือการใช้งานร่วมกับ OpenClaw

```
OpenClaw (หน้าจอเทอร์มินัล TUI)
        │
        ▼
  wall-vault proxy (:56244)   ← จัดการ Key, routing, fallback, events
        │
        ├─ Google Gemini API
        ├─ OpenRouter API (มากกว่า 340 โมเดล)
        └─ Ollama (เครื่องของคุณ, ทางเลือกสุดท้าย)
```

---

## การติดตั้ง

### Linux / macOS

เปิดเทอร์มินัลแล้ววางคำสั่งด้านล่างได้เลย

```bash
# Linux (PC ทั่วไป, เซิร์ฟเวอร์ — amd64)
curl -L https://github.com/sookmook/wall-vault/releases/latest/download/wall-vault-linux-amd64 \
  -o ~/.local/bin/wall-vault && chmod +x ~/.local/bin/wall-vault

# macOS Apple Silicon (Mac M1/M2/M3)
curl -L https://github.com/sookmook/wall-vault/releases/latest/download/wall-vault-darwin-arm64 \
  -o /usr/local/bin/wall-vault && chmod +x /usr/local/bin/wall-vault
```

- `curl -L ...` — ดาวน์โหลดไฟล์จากอินเทอร์เน็ต
- `chmod +x` — ทำให้ไฟล์ที่ดาวน์โหลดมา "สามารถรันได้" หากข้ามขั้นตอนนี้จะเกิดข้อผิดพลาด "Permission denied"

### Windows

เปิด PowerShell (สิทธิ์ผู้ดูแลระบบ) แล้วรันคำสั่งด้านล่าง

```powershell
# ดาวน์โหลด
Invoke-WebRequest -Uri `
  "https://github.com/sookmook/wall-vault/releases/latest/download/wall-vault-windows-amd64.exe" `
  -OutFile "$env:LOCALAPPDATA\Programs\wall-vault\wall-vault.exe"

# เพิ่มลงใน PATH (มีผลหลังรีสตาร์ท PowerShell)
$env:PATH += ";$env:LOCALAPPDATA\Programs\wall-vault"
```

> 💡 **PATH คืออะไร?** คือรายการโฟลเดอร์ที่คอมพิวเตอร์ใช้ค้นหาคำสั่ง การเพิ่มลงใน PATH ทำให้คุณพิมพ์ `wall-vault` จากโฟลเดอร์ไหนก็ได้

### บิลด์จากซอร์สโค้ด (สำหรับนักพัฒนา)

ใช้เฉพาะเมื่อมีสภาพแวดล้อมการพัฒนาภาษา Go ติดตั้งแล้วเท่านั้น

```bash
git clone https://github.com/sookmook/wall-vault
cd wall-vault
make build       # bin/wall-vault (เวอร์ชัน: v0.1.6.YYYYMMDD.HHmmss)
make install     # ~/.local/bin/wall-vault
```

> 💡 **เวอร์ชันพร้อม timestamp**: เมื่อบิลด์ด้วย `make build` เวอร์ชันจะถูกสร้างโดยอัตโนมัติในรูปแบบที่มีวันที่และเวลา เช่น `v0.1.6.20260314.231308` แต่หากบิลด์โดยตรงด้วย `go build ./...` เวอร์ชันจะแสดงเพียง `"dev"` เท่านั้น

---

## เริ่มต้นใช้งานครั้งแรก

### รัน setup wizard

หลังการติดตั้ง ครั้งแรกให้รัน **Setup Wizard** ด้วยคำสั่งด้านล่างเสมอ Wizard จะถามทีละขั้นตอนเพื่อตั้งค่าสิ่งที่จำเป็น

```bash
wall-vault setup
```

ขั้นตอนที่ Wizard จะดำเนินการ:

```
1. เลือกภาษา (10 ภาษา รวมถึงภาษาไทย)
2. เลือกธีม (light / dark / gold / cherry / ocean)
3. เลือกโหมดการทำงาน — ใช้คนเดียว (standalone) หรือใช้หลายเครื่องพร้อมกัน (distributed)
4. ใส่ชื่อ Bot — ชื่อที่จะแสดงในแดชบอร์ด
5. ตั้งค่า Port — ค่าเริ่มต้น: proxy 56244, vault 56243 (กด Enter ถ้าไม่ต้องการเปลี่ยน)
6. เลือกบริการ AI — Google / OpenRouter / Ollama
7. ตั้งค่าตัวกรองความปลอดภัยของเครื่องมือ
8. ตั้งค่า Admin Token — รหัสผ่านสำหรับล็อกฟีเจอร์จัดการแดชบอร์ด สามารถให้ระบบสร้างให้อัตโนมัติได้
9. ตั้งค่ารหัสผ่านเข้ารหัส API Key — สำหรับเก็บ Key อย่างปลอดภัยยิ่งขึ้น (ไม่บังคับ)
10. ระบุที่บันทึกไฟล์ config
```

> ⚠️ **อย่าลืม Admin Token** คุณจะต้องใช้มันในภายหลังเมื่อต้องการเพิ่ม Key หรือเปลี่ยนการตั้งค่าในแดชบอร์ด หากลืมจะต้องแก้ไขไฟล์ config โดยตรง

เมื่อ Wizard เสร็จสิ้น ไฟล์ config `wall-vault.yaml` จะถูกสร้างขึ้นอัตโนมัติ

### การรัน

```bash
wall-vault start
```

เซิร์ฟเวอร์สองตัวจะเริ่มทำงานพร้อมกัน:

- **Proxy** (`http://localhost:56244`) — ตัวแทนที่เชื่อมต่อ OpenClaw กับบริการ AI
- **Key Vault** (`http://localhost:56243`) — จัดการ API Key และแดชบอร์ดเว็บ

เปิดเบราว์เซอร์ไปที่ `http://localhost:56243` เพื่อดูแดชบอร์ดได้ทันที

---

## การลงทะเบียน API Key

มีสี่วิธีในการลงทะเบียน API Key **สำหรับผู้เริ่มต้น แนะนำวิธีที่ 1 (Environment Variable)**

### วิธีที่ 1: Environment Variable (แนะนำ — ง่ายที่สุด)

Environment Variable (ตัวแปรสภาพแวดล้อม) คือ **ค่าที่กำหนดไว้ล่วงหน้า** ซึ่งโปรแกรมจะอ่านเมื่อเริ่มต้น ให้พิมพ์คำสั่งด้านล่างในเทอร์มินัล

```bash
# ลงทะเบียน Google Gemini Key
export WV_KEY_GOOGLE=AIzaSy...

# ลงทะเบียน OpenRouter Key
export WV_KEY_OPENROUTER=sk-or-v1-...

# รันหลังลงทะเบียน
wall-vault start
```

หากมีหลาย Key ให้คั่นด้วยเครื่องหมายจุลภาค (,) wall-vault จะวนใช้ Key แต่ละตัวโดยอัตโนมัติ (round-robin):

```bash
export WV_KEY_GOOGLE=AIzaSy...,AIzaSy...,AIzaSy...
```

> 💡 **เคล็ดลับ**: คำสั่ง `export` มีผลเฉพาะเซสชันเทอร์มินัลปัจจุบันเท่านั้น หากต้องการให้คงอยู่หลังรีสตาร์ทเครื่อง ให้เพิ่มบรรทัดดังกล่าวลงในไฟล์ `~/.bashrc` หรือ `~/.zshrc`

### วิธีที่ 2: แดชบอร์ด UI (คลิกด้วยเมาส์)

1. เปิดเบราว์เซอร์ไปที่ `http://localhost:56243`
2. คลิกปุ่ม `[+ เพิ่ม]` ในการ์ด **🔑 API Key** ที่ด้านบน
3. กรอกประเภทบริการ, ค่า Key, Label (ชื่อสำหรับจดจำ), และขีดจำกัดรายวัน แล้วบันทึก

### วิธีที่ 3: REST API (สำหรับระบบอัตโนมัติ/สคริปต์)

REST API คือวิธีที่โปรแกรมต่างๆ รับส่งข้อมูลกันผ่าน HTTP เหมาะสำหรับการลงทะเบียนอัตโนมัติด้วยสคริปต์

```bash
curl -X POST http://localhost:56243/admin/keys \
  -H "Authorization: Bearer 관리자토큰" \
  -H "Content-Type: application/json" \
  -d '{
    "service": "google",
    "key": "AIzaSy...",
    "label": "메인 키",
    "daily_limit": 1000
  }'
```

### วิธีที่ 4: Proxy flag (สำหรับทดสอบชั่วคราว)

ใช้เมื่อต้องการทดสอบโดยใส่ Key ชั่วคราวโดยไม่ต้องลงทะเบียนอย่างเป็นทางการ Key จะหายไปเมื่อปิดโปรแกรม

```bash
wall-vault proxy --key-google=AIzaSy... --key-openrouter=sk-or-...
```

---

## วิธีใช้งาน Proxy

### การใช้งานกับ OpenClaw (จุดประสงค์หลัก)

วิธีตั้งค่าให้ OpenClaw เชื่อมต่อบริการ AI ผ่าน wall-vault

เปิดไฟล์ `~/.openclaw/openclaw.json` แล้วเพิ่มเนื้อหาด้านล่าง:

```json5
// ~/.openclaw/openclaw.json
{
  models: {
    providers: {
      "wall-vault": {
        baseUrl: "http://localhost:56244/v1",
        apiKey: "your-agent-token",   // vault agent token
        api: "openai-completions",
        models: [
          { id: "wall-vault/gemini-2.5-flash" },
          { id: "wall-vault/gemini-2.5-pro" },
          { id: "wall-vault/hunter-alpha" },    // context ฟรี 1M
          { id: "wall-vault/claude-opus-4-6" }
        ]
      }
    }
  }
}
```

> 💡 **วิธีที่ง่ายกว่า**: กดปุ่ม **🦞 คัดลอก config OpenClaw** ในการ์ด Agent ของแดชบอร์ด จะได้ snippet ที่มี token และที่อยู่กรอกไว้แล้ว แค่วางลงไปได้เลย

**`wall-vault/` ด้านหน้าชื่อโมเดลส่งไปที่ไหน?**

wall-vault จะตัดสินใจโดยอัตโนมัติว่าจะส่ง request ไปยังบริการ AI ใดตามชื่อโมเดล:

| รูปแบบโมเดล | บริการที่เชื่อมต่อ |
|-------------|-----------------|
| `wall-vault/gemini-*` | เชื่อมต่อ Google Gemini โดยตรง |
| `wall-vault/gpt-*`, `wall-vault/o3`, `wall-vault/o4*` | เชื่อมต่อ OpenAI โดยตรง |
| `wall-vault/claude-*` | เชื่อมต่อ Anthropic ผ่าน OpenRouter |
| `wall-vault/hunter-alpha`, `wall-vault/healer-alpha` | OpenRouter (context ฟรี 1 ล้าน token) |
| `wall-vault/kimi-*`, `wall-vault/glm-*`, `wall-vault/deepseek-*` | เชื่อมต่อ OpenRouter |
| `google/ชื่อโมเดล`, `openai/ชื่อโมเดล`, `anthropic/ชื่อโมเดล` เป็นต้น | เชื่อมต่อบริการนั้นโดยตรง |
| `custom/google/ชื่อโมเดล`, `custom/openai/ชื่อโมเดล` เป็นต้น | ลบส่วน `custom/` แล้วส่งต่อ |
| `ชื่อโมเดล:cloud` | ลบส่วน `:cloud` แล้วเชื่อมต่อ OpenRouter |

> 💡 **Context คืออะไร?** คือปริมาณการสนทนาที่ AI จำได้ในครั้งเดียว 1M (หนึ่งล้าน token) หมายความว่าสามารถประมวลผลบทสนทนายาวมากหรือเอกสารขนาดใหญ่ได้ในครั้งเดียว

### เชื่อมต่อโดยตรงในรูปแบบ Gemini API (ความเข้ากันได้กับเครื่องมือเดิม)

หากมีเครื่องมือที่ใช้ Google Gemini API โดยตรงอยู่แล้ว เพียงเปลี่ยนที่อยู่มาที่ wall-vault:

```bash
export ANTHROPIC_BASE_URL=http://localhost:56244/google
```

หรือหากเครื่องมือรองรับการระบุ URL โดยตรง:

```
http://localhost:56244/google/v1beta/models/gemini-2.5-flash:generateContent
```

### การใช้งานกับ OpenAI SDK (Python)

สามารถเชื่อมต่อ wall-vault จากโค้ด Python ที่ใช้ AI ได้ เพียงเปลี่ยน `base_url`:

```python
from openai import OpenAI

client = OpenAI(
    base_url="http://localhost:56244/v1",
    api_key="not-needed"  # wall-vault จัดการ API Key ให้อัตโนมัติ
)

response = client.chat.completions.create(
    model="google/gemini-2.5-flash",   # ระบุในรูปแบบ provider/model
    messages=[{"role": "user", "content": "สวัสดีครับ"}]
)
```

### เปลี่ยนโมเดลขณะรันอยู่

หากต้องการเปลี่ยนโมเดล AI ที่ใช้งานขณะที่ wall-vault กำลังทำงานอยู่:

```bash
# ส่ง request โดยตรงไปยัง proxy เพื่อเปลี่ยนโมเดล
curl -X PUT http://localhost:56244/api/config/model \
  -H "Content-Type: application/json" \
  -d '{"service": "openrouter", "model": "anthropic/claude-3.5-sonnet"}'

# ในโหมดกระจาย (multi-bot) เปลี่ยนที่ vault server → จะสะท้อนทันทีผ่าน SSE
curl -X PUT http://localhost:56243/admin/clients/내-봇-id \
  -H "Authorization: Bearer 관리자토큰" \
  -H "Content-Type: application/json" \
  -d '{"default_service": "google", "default_model": "gemini-2.5-pro"}'
```

### ดูรายชื่อโมเดลที่ใช้ได้

```bash
# ดูรายการทั้งหมด
curl http://localhost:56244/api/models | python3 -m json.tool

# ดูเฉพาะโมเดล Google
curl "http://localhost:56244/api/models?service=google"

# ค้นหาตามชื่อ (เช่น โมเดลที่มีคำว่า "claude")
curl "http://localhost:56244/api/models?q=claude"
```

**สรุปโมเดลหลักแต่ละบริการ:**

| บริการ | โมเดลหลัก |
|--------|-----------|
| Google | gemini-2.5-pro, gemini-2.5-flash, gemini-2.5-flash-8b, gemini-2.0-flash |
| OpenAI | gpt-4o, gpt-4o-mini, o3, o1, o1-mini |
| OpenRouter | มากกว่า 346 โมเดล (Hunter Alpha context 1M ฟรี, DeepSeek R1/V3, Qwen 2.5 เป็นต้น) |
| Ollama | ตรวจจับ local server ที่ติดตั้งในเครื่องอัตโนมัติ |

---

## แดชบอร์ดคลังกุญแจ

เปิดเบราว์เซอร์ไปที่ `http://localhost:56243` เพื่อดูแดชบอร์ด

**โครงสร้างหน้าจอ:**
- **แถบด้านบน (topbar)**: โลโก้, ตัวเลือกภาษาและธีม, สถานะการเชื่อมต่อ SSE
- **Grid การ์ด**: การ์ด Agent, บริการ, และ API Key จัดเรียงในรูปแบบ tile

### การ์ด API Key

การ์ดสำหรับจัดการ API Key ที่ลงทะเบียนไว้ทั้งหมดในที่เดียว

- แสดงรายการ Key แยกตามบริการ
- `today_usage`: จำนวน token (หน่วยนับข้อความของ AI) ที่ประมวลผลสำเร็จวันนี้
- `today_attempts`: จำนวนครั้งที่เรียกใช้ทั้งหมดวันนี้ (รวมสำเร็จ + ล้มเหลว)
- ใช้ปุ่ม `[+ เพิ่ม]` เพื่อลงทะเบียน Key ใหม่ และ `✕` เพื่อลบ Key

> 💡 **Token คืออะไร?** คือหน่วยที่ AI ใช้ประมวลผลข้อความ โดยประมาณเท่ากับหนึ่งคำภาษาอังกฤษ หรือตัวอักษรไทย 1–2 ตัว ค่าบริการ API มักคิดตามจำนวน token นี้

### การ์ด Agent

การ์ดที่แสดงสถานะของ Bot (Agent) ที่เชื่อมต่อกับ wall-vault proxy

**สถานะการเชื่อมต่อแสดง 4 ระดับ:**

| สัญลักษณ์ | สถานะ | ความหมาย |
|----------|-------|---------|
| 🟢 | กำลังทำงาน | Proxy ทำงานปกติ |
| 🟡 | ล่าช้า | ได้รับ response แต่ช้า |
| 🔴 | ออฟไลน์ | Proxy ไม่ตอบสนอง |
| ⚫ | ไม่เชื่อมต่อ/ปิดใช้งาน | Proxy ไม่เคยเชื่อมต่อกับ vault หรือถูกปิดใช้งาน |

**คำแนะนำปุ่มด้านล่างการ์ด Agent:**

เมื่อลงทะเบียน Agent คุณสามารถระบุ **ประเภท Agent** ได้ และปุ่มสะดวกที่เหมาะกับประเภทนั้นจะปรากฏขึ้นโดยอัตโนมัติ

---

#### 🔘 ปุ่มคัดลอก Config — สร้างการตั้งค่าการเชื่อมต่อให้อัตโนมัติ

เมื่อกดปุ่ม snippet การตั้งค่าที่มี token, ที่อยู่ proxy, และข้อมูลโมเดลของ agent นั้นกรอกไว้ครบแล้วจะถูกคัดลอกไปยัง clipboard เพียงวางลงในตำแหน่งตามตารางด้านล่างเพื่อตั้งค่าการเชื่อมต่อให้เสร็จสมบูรณ์

| ปุ่ม | ประเภท Agent | ตำแหน่งที่วาง |
|------|------------|-------------|
| 🦞 คัดลอก config OpenClaw | `openclaw` | `~/.openclaw/openclaw.json` |
| 🦀 คัดลอก config NanoClaw | `nanoclaw` | `~/.openclaw/openclaw.json` |
| 🟠 คัดลอก config Claude Code | `claude-code` | `~/.claude/settings.json` |
| ⌨ คัดลอก config Cursor | `cursor` | Cursor → Settings → AI |
| 💻 คัดลอก config VSCode | `vscode` | `~/.continue/config.json` |

**ตัวอย่าง — หากเป็นประเภท Claude Code จะคัดลอกเนื้อหาแบบนี้:**

```json
// ~/.claude/settings.json
{
  "apiProvider": "openai",
  "baseUrl": "http://192.168.0.6:56244/v1",
  "apiKey": "이-에이전트의-토큰"
}
```

**ตัวอย่าง — หากเป็นประเภท VSCode (Continue):**

```json
// ~/.continue/config.json
{
  "models": [{
    "title": "wall-vault proxy",
    "provider": "openai",
    "model": "gemini-2.5-flash",
    "apiBase": "http://192.168.0.6:56244/v1",
    "apiKey": "이-에이전트의-토큰"
  }]
}
```

**ตัวอย่าง — หากเป็นประเภท Cursor:**

```
Base URL : http://192.168.0.6:56244/v1
API Key  : 이-에이전트의-토큰

// หรือ environment variable:
OPENAI_BASE_URL=http://192.168.0.6:56244/v1
OPENAI_API_KEY=이-에이전트의-토큰
```

> ⚠️ **เมื่อการคัดลอกไปยัง clipboard ไม่ทำงาน**: นโยบายความปลอดภัยของเบราว์เซอร์อาจบล็อกการคัดลอก หากมี popup กล่องข้อความเปิดขึ้น ให้กด Ctrl+A เพื่อเลือกทั้งหมด แล้วกด Ctrl+C เพื่อคัดลอก

---

#### ⚡ ปุ่มตั้งค่าอัตโนมัติ — กดครั้งเดียวเสร็จสิ้น

หากประเภท agent เป็น `cline`, `claude-code`, `openclaw`, `nanoclaw` จะมีปุ่ม **⚡ ตั้งค่า** ปรากฏบนการ์ด agent เมื่อกดปุ่มนี้ ไฟล์ตั้งค่าในเครื่องของ agent นั้นจะถูกอัปเดตอัตโนมัติ

| ปุ่ม | ประเภท Agent | ไฟล์ที่ตั้งค่า |
|------|-------------|-------------|
| ⚡ ตั้งค่า Cline | `cline` | `~/.cline/data/globalState.json` + `secrets.json` |
| ⚡ ตั้งค่า Claude Code | `claude-code` | `~/.claude/settings.json` |
| ⚡ ตั้งค่า OpenClaw | `openclaw` | `~/.openclaw/openclaw.json` |
| ⚡ ตั้งค่า NanoClaw | `nanoclaw` | `~/.openclaw/openclaw.json` |

> ⚠️ ปุ่มนี้ส่งคำขอไปยัง **localhost:56244** (proxy ในเครื่อง) ดังนั้น proxy ต้องทำงานอยู่บนเครื่องนั้นจึงจะใช้งานได้

---

#### 🔀 ลากและวางเพื่อจัดเรียงการ์ด (v0.1.17)

คุณสามารถ**ลาก**การ์ดเอเจนต์บนแดชบอร์ดเพื่อจัดเรียงตามลำดับที่ต้องการ

1. คลิกค้างที่การ์ดเอเจนต์แล้วลาก
2. วางบนการ์ดในตำแหน่งที่ต้องการ ลำดับจะเปลี่ยนทันที
3. ลำดับที่เปลี่ยนแปลง**จะถูกบันทึกลงเซิร์ฟเวอร์ทันที** และจะคงอยู่แม้รีเฟรชหน้า

> 💡 อุปกรณ์ระบบสัมผัส (มือถือ/แท็บเล็ต) ยังไม่รองรับในขณะนี้ กรุณาใช้เบราว์เซอร์เดสก์ท็อป

---

#### 🔄 การซิงค์โมเดลสองทิศทาง (v0.1.16)

เมื่อเปลี่ยนโมเดลของ agent บนแดชบอร์ด vault การตั้งค่าในเครื่องของ agent นั้นจะถูกอัปเดตอัตโนมัติ

**กรณี Cline:**
- เปลี่ยนโมเดลบน vault → SSE event → proxy อัปเดตฟิลด์โมเดลใน `globalState.json`
- ฟิลด์ที่อัปเดต: `actModeOpenAiModelId`, `planModeOpenAiModelId`, `openAiModelId`
- `openAiBaseUrl` และ API key จะไม่ถูกเปลี่ยนแปลง
- **ต้อง reload VS Code (`Ctrl+Alt+R` หรือ `Ctrl+Shift+P` → `Developer: Reload Window`)**
  - เพราะ Cline ไม่ได้อ่านไฟล์ตั้งค่าใหม่ในระหว่างที่ทำงาน

**กรณี Claude Code:**
- เปลี่ยนโมเดลบน vault → SSE event → proxy อัปเดตฟิลด์ `model` ใน `settings.json`
- ค้นหา path อัตโนมัติทั้ง WSL และ Windows (`~/.claude/`, `/mnt/c/Users/*/.claude/`)

**ทิศทางตรงข้าม (agent → vault):**
- เมื่อ agent (Cline, Claude Code เป็นต้น) ส่งคำขอผ่าน proxy จะรวมข้อมูลบริการ·โมเดลของ client ไว้ใน heartbeat
- การ์ด agent บนแดชบอร์ด vault จะแสดงบริการ/โมเดลที่ใช้อยู่ในขณะนั้นแบบเรียลไทม์

> 💡 **สรุป**: proxy ระบุ agent จาก Authorization token ในคำขอ และเราท์ไปยังบริการ/โมเดลที่ตั้งค่าไว้บน vault โดยอัตโนมัติ แม้ Cline หรือ Claude Code จะส่งชื่อโมเดลอื่นมา proxy จะ override ด้วยการตั้งค่าของ vault

---

### การใช้ Cline ใน VS Code — คู่มือฉบับละเอียด

#### ขั้นตอนที่ 1: ติดตั้ง Cline

ติดตั้ง **Cline** (ID: `saoudrizwan.claude-dev`) จาก VS Code Extension Marketplace

#### ขั้นตอนที่ 2: ลงทะเบียน agent บน vault

1. เปิดแดชบอร์ด vault (`http://IP_vault:56243`)
2. คลิก **+ เพิ่ม** ในส่วน **agent**
3. กรอกข้อมูลดังนี้:

| ฟิลด์ | ค่า | คำอธิบาย |
|------|----|------|
| ID | `my_cline` | ตัวระบุเฉพาะ (ภาษาอังกฤษ ไม่มีช่องว่าง) |
| ชื่อ | `My Cline` | ชื่อที่แสดงบนแดชบอร์ด |
| ประเภท Agent | `cline` | ← ต้องเลือก `cline` |
| บริการ | เลือกบริการที่จะใช้ (เช่น `google`) | |
| โมเดล | ป้อนโมเดลที่จะใช้ (เช่น `gemini-2.5-flash`) | |

4. กด **บันทึก** แล้ว token จะถูกสร้างขึ้นอัตโนมัติ

#### ขั้นตอนที่ 3: เชื่อมต่อ Cline

**วิธี A — ตั้งค่าอัตโนมัติ (แนะนำ)**

1. ตรวจสอบว่า wall-vault **proxy** กำลังทำงานบนเครื่องนั้น (`localhost:56244`)
2. กดปุ่ม **⚡ ตั้งค่า Cline** บนการ์ด agent ในแดชบอร์ด
3. เมื่อเห็นข้อความ "ตั้งค่าสำเร็จ!" แสดงว่าเรียบร้อย
4. Reload VS Code (`Ctrl+Alt+R`)

**วิธี B — ตั้งค่าด้วยตนเอง**

เปิดการตั้งค่า (⚙️) ใน sidebar ของ Cline:
- **API Provider**: `OpenAI Compatible`
- **Base URL**: `http://ที่อยู่proxy:56244/v1`
  - เครื่องเดียวกัน: `http://localhost:56244/v1`
  - เครื่องอื่น เช่น Mac Mini: `http://192.168.0.6:56244/v1`
- **API Key**: token ที่ได้จาก vault (คัดลอกจากการ์ด agent)
- **Model ID**: โมเดลที่ตั้งค่าบน vault (เช่น `gemini-2.5-flash`)

#### ขั้นตอนที่ 4: ยืนยัน

ส่งข้อความอะไรก็ได้ในแชทของ Cline หากทำงานปกติ:
- การ์ด agent บนแดชบอร์ด vault จะแสดง **จุดสีเขียว (● กำลังทำงาน)**
- การ์ดจะแสดงบริการ/โมเดลปัจจุบัน (เช่น `google / gemini-2.5-flash`)

#### การเปลี่ยนโมเดล

หากต้องการเปลี่ยนโมเดลของ Cline ให้เปลี่ยนที่ **แดชบอร์ด vault**:

1. เปลี่ยน dropdown บริการ/โมเดลบนการ์ด agent
2. กด **ตกลง**
3. Reload VS Code (`Ctrl+Alt+R`) — ชื่อโมเดลใน footer ของ Cline จะอัปเดต
4. คำขอถัดไปจะใช้โมเดลใหม่

> 💡 ในความเป็นจริง proxy จะระบุคำขอของ Cline ด้วย token และเราท์ไปยังโมเดลที่ตั้งค่าบน vault แม้ไม่ reload VS Code **โมเดลจริงที่ใช้งานจะเปลี่ยนทันที** — การ reload เพื่ออัปเดตชื่อโมเดลที่แสดงบน UI ของ Cline เท่านั้น

#### การตรวจจับการตัดการเชื่อมต่อ

เมื่อปิด VS Code การ์ด agent บนแดชบอร์ด vault จะเปลี่ยนเป็นสีเหลือง (ล่าช้า) ภายใน **90 วินาที** และเปลี่ยนเป็นสีแดง (ออฟไลน์) ภายใน **3 นาที** (ตั้งแต่ v0.1.18 การตรวจสอบสถานะทุก 15 วินาทีทำให้ตรวจจับออฟไลน์ได้เร็วขึ้น)

#### การแก้ปัญหา

| อาการ | สาเหตุ | วิธีแก้ |
|------|------|------|
| Cline แสดง "เชื่อมต่อล้มเหลว" | proxy ไม่ทำงานหรือที่อยู่ผิด | ตรวจสอบ proxy ด้วย `curl http://localhost:56244/health` |
| ไม่มีจุดสีเขียวบน vault | ยังไม่ได้ตั้งค่า API key (token) | กดปุ่ม **⚡ ตั้งค่า Cline** อีกครั้ง |
| ชื่อโมเดลใน footer ของ Cline ไม่เปลี่ยน | Cline แคชการตั้งค่า | Reload VS Code (`Ctrl+Alt+R`) |
| แสดงชื่อโมเดลผิด | บั๊กเก่า (แก้ไขแล้วใน v0.1.16) | อัปเดต proxy เป็น v0.1.16 ขึ้นไป |

---

#### 🟣 ปุ่มคัดลอกคำสั่ง Deploy — ใช้เมื่อติดตั้งบนเครื่องใหม่

ใช้เมื่อต้องการติดตั้ง wall-vault proxy บนคอมพิวเตอร์ใหม่และเชื่อมต่อกับ vault เมื่อกดปุ่ม สคริปต์ติดตั้งทั้งหมดจะถูกคัดลอก นำไปวางและรันในเทอร์มินัลของเครื่องใหม่ เพื่อดำเนินการต่อไปนี้พร้อมกัน:

1. ติดตั้ง binary ของ wall-vault (ข้ามหากติดตั้งแล้ว)
2. ลงทะเบียน systemd user service อัตโนมัติ
3. เริ่มบริการและเชื่อมต่อกับ vault อัตโนมัติ

> 💡 สคริปต์มี token ของ agent นี้และที่อยู่ vault server กรอกไว้แล้ว จึงสามารถรันได้ทันทีหลังวางโดยไม่ต้องแก้ไขอะไรเพิ่ม

---

### การ์ดบริการ

การ์ดสำหรับเปิด/ปิดหรือตั้งค่าบริการ AI ที่จะใช้งาน

- สวิตช์เปิด/ปิดแต่ละบริการ
- หากป้อนที่อยู่ของ AI server ในเครื่อง (Ollama, LM Studio, vLLM ที่รันในเครื่องของคุณ) ระบบจะค้นหาโมเดลที่ใช้ได้โดยอัตโนมัติ
- **แสดงสถานะการเชื่อมต่อ local service**: จุด ● ข้างชื่อบริการ **สีเขียว** = เชื่อมต่ออยู่, **สีเทา** = ไม่ได้เชื่อมต่อ
- **แสดงสถานะ local service**: เมื่อเปิดหน้า หาก local service (เช่น Ollama) กำลังทำงาน จุด ● จะเปลี่ยนเป็นสีเขียว แต่สถานะ checkbox จะไม่ถูกเปลี่ยน

> 💡 **หาก local service รันอยู่บนเครื่องอื่น**: ให้ป้อน IP ของเครื่องนั้นในช่อง URL ของบริการ เช่น `http://192.168.0.6:11434` (Ollama), `http://192.168.0.6:1234` (LM Studio)

### การป้อน Admin Token

เมื่อต้องการใช้ฟีเจอร์สำคัญในแดชบอร์ด เช่น เพิ่มหรือลบ Key จะมี popup ให้ป้อน Admin Token ให้ป้อน token ที่ตั้งค่าไว้ใน setup wizard เมื่อป้อนครั้งหนึ่งแล้วจะคงอยู่จนกว่าจะปิดเบราว์เซอร์

> ⚠️ **หากการยืนยันตัวตนล้มเหลวเกิน 10 ครั้งภายใน 15 นาที IP นั้นจะถูกบล็อกชั่วคราว** หากลืม token ให้ตรวจสอบรายการ `admin_token` ในไฟล์ `wall-vault.yaml`

---

## โหมดกระจาย (Multi-Bot)

เมื่อต้องการใช้งาน OpenClaw บนหลายเครื่องพร้อมกัน ให้ใช้การตั้งค่าที่ **แชร์ vault กุญแจเดียว** คุณจัดการ Key ที่จุดเดียวเท่านั้นจึงสะดวกมาก

### ตัวอย่างการตั้งค่า

```
[Vault Server]
  wall-vault vault    (Key vault :56243, แดชบอร์ด)

[WSL Alpha]            [Raspberry Pi Gamma]    [Mac Mini Local]
  wall-vault proxy      wall-vault proxy        wall-vault proxy
  openclaw TUI          openclaw TUI            openclaw TUI
  ↕ SSE sync            ↕ SSE sync              ↕ SSE sync
```

Bot ทุกตัวมองไปที่ vault server ตรงกลาง ดังนั้นเมื่อเปลี่ยนโมเดลหรือเพิ่ม Key ใน vault จะสะท้อนไปยัง Bot ทุกตัวทันที

### ขั้นตอนที่ 1: เริ่ม vault server

รันบนเครื่องที่จะใช้เป็น vault server:

```bash
wall-vault vault
```

### ขั้นตอนที่ 2: ลงทะเบียน Bot แต่ละตัว (client)

ลงทะเบียนข้อมูล Bot ที่จะเชื่อมต่อกับ vault server ล่วงหน้า:

```bash
curl -X POST http://localhost:56243/admin/clients \
  -H "Authorization: Bearer 관리자토큰" \
  -H "Content-Type: application/json" \
  -d '{
    "id": "botA",
    "name": "봇A",
    "token": "bota-secret",
    "default_service": "google",
    "default_model": "gemini-2.5-flash"
  }'
```

### ขั้นตอนที่ 3: เริ่ม proxy บนเครื่องของแต่ละ Bot

บนเครื่องแต่ละเครื่องที่ติดตั้ง Bot รัน proxy โดยระบุที่อยู่ vault server และ token:

```bash
WV_VAULT_URL=http://192.168.x.x:56243 \
WV_VAULT_TOKEN=bota-secret \
WV_VAULT_CLIENT_ID=botA \
wall-vault proxy
```

> 💡 แทนที่ **`192.168.x.x`** ด้วย IP ภายในจริงของเครื่อง vault server ตรวจสอบได้จากการตั้งค่า router หรือคำสั่ง `ip addr`

---

## การตั้งค่าเริ่มต้นอัตโนมัติ

หากการเปิด wall-vault ด้วยตัวเองทุกครั้งที่รีสตาร์ทเครื่องเป็นเรื่องยุ่งยาก ให้ลงทะเบียนเป็น system service เมื่อลงทะเบียนครั้งหนึ่งแล้ว ระบบจะเริ่มต้นอัตโนมัติเมื่อบูต

### Linux — systemd (Linux ส่วนใหญ่)

systemd คือระบบที่จัดการการเริ่มต้นและบริหารโปรแกรมใน Linux โดยอัตโนมัติ:

```bash
wall-vault doctor deploy
systemctl --user daemon-reload
systemctl --user enable --now wall-vault
```

ดู log:

```bash
journalctl --user -u wall-vault -f
```

### macOS — launchd

ระบบที่จัดการการรันโปรแกรมอัตโนมัติใน macOS:

```bash
wall-vault doctor deploy launchd
launchctl load ~/Library/LaunchAgents/com.wall-vault.plist
```

### Windows — NSSM

1. ดาวน์โหลด NSSM จาก [nssm.cc](https://nssm.cc/download) แล้วเพิ่มลงใน PATH
2. รันใน PowerShell สิทธิ์ผู้ดูแลระบบ:

```powershell
wall-vault doctor deploy windows
```

---

## Doctor ผู้ช่วยวินิจฉัย

คำสั่ง `doctor` คือ **เครื่องมือที่วินิจฉัยและแก้ไขตัวเอง** เพื่อตรวจสอบว่า wall-vault ตั้งค่าถูกต้องหรือไม่

```bash
wall-vault doctor check   # วินิจฉัยสถานะปัจจุบัน (อ่านอย่างเดียว ไม่เปลี่ยนแปลงใดๆ)
wall-vault doctor fix     # แก้ไขปัญหาอัตโนมัติ
wall-vault doctor all     # วินิจฉัย + แก้ไขอัตโนมัติในครั้งเดียว
```

> 💡 หากรู้สึกว่ามีบางอย่างผิดปกติ ให้ลองรัน `wall-vault doctor all` ก่อน มันสามารถจับปัญหาส่วนใหญ่ได้อัตโนมัติ

---

## ตัวแปรสภาพแวดล้อม

ตัวแปรสภาพแวดล้อม (Environment Variable) คือวิธีส่งค่าการตั้งค่าให้โปรแกรม พิมพ์ในรูปแบบ `export ชื่อตัวแปร=ค่า` ในเทอร์มินัล หรือใส่ไว้ในไฟล์ service สำหรับเริ่มต้นอัตโนมัติเพื่อให้มีผลตลอดเวลา

| ตัวแปร | คำอธิบาย | ตัวอย่างค่า |
|--------|---------|-----------|
| `WV_LANG` | ภาษาของแดชบอร์ด | `ko`, `en`, `ja` |
| `WV_THEME` | ธีมของแดชบอร์ด | `light`, `dark`, `gold` |
| `WV_KEY_GOOGLE` | Google API Key (คั่นด้วยจุลภาคสำหรับหลาย Key) | `AIza...,AIza...` |
| `WV_KEY_OPENROUTER` | OpenRouter API Key | `sk-or-v1-...` |
| `WV_VAULT_URL` | ที่อยู่ vault server ในโหมดกระจาย | `http://192.168.x.x:56243` |
| `WV_VAULT_TOKEN` | Token สำหรับยืนยันตัวตน client (Bot) | `my-secret-token` |
| `WV_ADMIN_TOKEN` | Admin Token | `admin-token-here` |
| `WV_MASTER_PASS` | รหัสผ่านเข้ารหัส API Key | `my-password` |
| `WV_AVATAR` | path ไฟล์รูปอวตาร (relative path จาก `~/.openclaw/`) | `workspace/avatars/avatar.png` |
| `OLLAMA_URL` | ที่อยู่ Ollama local server | `http://192.168.x.x:11434` |

---

## การแก้ปัญหา

### เมื่อ Proxy ไม่เริ่มต้น

มักเกิดจาก port ถูกโปรแกรมอื่นใช้อยู่แล้ว

```bash
ss -tlnp | grep 56244   # ตรวจสอบว่าใครใช้ port 56244 อยู่
wall-vault proxy --port 8080   # เริ่มต้นด้วย port อื่น
```

### เมื่อเกิดข้อผิดพลาด API Key (429, 402, 401, 403, 582)

| รหัสข้อผิดพลาด | ความหมาย | วิธีแก้ไข |
|--------------|---------|---------|
| **429** | Request มากเกินไป (เกินขีดจำกัดการใช้งาน) | รอสักครู่หรือเพิ่ม Key ใหม่ |
| **402** | ต้องชำระเงินหรือเครดิตไม่พอ | เติมเครดิตในบริการนั้น |
| **401 / 403** | Key ไม่ถูกต้องหรือไม่มีสิทธิ์ | ตรวจสอบค่า Key แล้วลงทะเบียนใหม่ |
| **582** | Gateway โหลดหนักเกิน (cooldown 5 นาที) | จะปลดล็อกอัตโนมัติหลัง 5 นาที |

```bash
# ตรวจสอบรายการและสถานะ Key ที่ลงทะเบียนไว้
curl -H "Authorization: Bearer 관리자토큰" http://localhost:56243/admin/keys

# รีเซ็ต counter การใช้งาน Key
curl -X POST -H "Authorization: Bearer 관리자토큰" http://localhost:56243/admin/keys/reset
```

### เมื่อ Agent แสดงว่า "ไม่เชื่อมต่อ"

"ไม่เชื่อมต่อ" หมายความว่า proxy process ไม่ได้ส่งสัญญาณ (heartbeat) ไปยัง vault **ไม่ได้หมายความว่าการตั้งค่าไม่ได้ถูกบันทึก** Proxy ต้องรู้ที่อยู่ vault server และ token จึงจะเปลี่ยนเป็นสถานะเชื่อมต่อได้

```bash
# เริ่ม proxy โดยระบุที่อยู่ vault server, token, และ client ID
WV_VAULT_URL=http://금고서버주소:56243 \
WV_VAULT_TOKEN=클라이언트토큰 \
WV_VAULT_CLIENT_ID=클라이언트ID \
wall-vault proxy
```

เมื่อเชื่อมต่อสำเร็จ แดชบอร์ดจะเปลี่ยนเป็น 🟢 กำลังทำงาน ภายในประมาณ 20 วินาที

### เมื่อ Ollama เชื่อมต่อไม่ได้

Ollama คือโปรแกรมที่รัน AI โดยตรงบนเครื่องของคุณ ให้ตรวจสอบก่อนว่า Ollama เปิดอยู่หรือไม่

```bash
curl http://localhost:11434/api/tags   # หากแสดงรายการโมเดล แปลว่าปกติ
export OLLAMA_URL=http://192.168.x.x:11434   # หากรันอยู่บนเครื่องอื่น
```

> ⚠️ หาก Ollama ไม่ตอบสนอง ให้เริ่ม Ollama ก่อนด้วยคำสั่ง `ollama serve`

> ⚠️ **โมเดลขนาดใหญ่ตอบสนองช้า**: โมเดลใหญ่อย่าง `qwen3.5:35b` หรือ `deepseek-r1` อาจใช้เวลาหลายนาทีในการสร้าง response แม้จะดูเหมือนไม่มีการตอบสนอง แต่อาจกำลังประมวลผลอยู่ตามปกติ กรุณารอ

---

## การเปลี่ยนแปลงล่าสุด (v0.1.16 ~ v0.1.21)

### v0.1.21 (2026-04-05)
- **รองรับโมเดล Gemma 4**: โมเดล Gemma (gemma-4-31b-it, gemma-4-26b-a4b-it) ถูกส่งผ่าน Google Gemini API แล้ว
- **รองรับ LM Studio / vLLM**: บริการในเครื่องเหล่านี้ถูกส่งต่ออย่างถูกต้องแทนที่จะ fallback ไปยัง Ollama
- **แก้ไข Dashboard**: แสดงบริการที่กำหนดค่าไว้เสมอ ไม่ใช่บริการ fallback
- **คง checkbox ของ local service**: Dashboard ไม่ปิด local service อัตโนมัติเมื่อโหลดหน้าอีกต่อไป
- **ตัวแปรสภาพแวดล้อม tool filter**: รองรับ `WV_TOOL_FILTER=passthrough`

### v0.1.20 (2026-03-28)
- **เสริมความปลอดภัยอย่างครอบคลุม**: ป้องกัน XSS (41 จุด), เปรียบเทียบ token แบบเวลาคงที่, จำกัด CORS, จำกัดขนาดคำขอ และอื่นๆ

### v0.1.19 (2026-03-27)
- **ตรวจจับ Claude Code ออนไลน์**: Claude Code แสดงเป็นออนไลน์บน dashboard แม้จะ bypass proxy

### v0.1.18 (2026-03-26)
- **แก้ไขการกู้คืน fallback**: กู้คืนอัตโนมัติไปยังบริการที่ต้องการเมื่อพร้อมใช้งาน
- **ตรวจจับออฟไลน์ที่ดีขึ้น**: polling สถานะทุก 15 วินาที

### v0.1.17 (2026-03-25)
- **จัดเรียงการ์ดด้วยการลากและวาง**
- **ปุ่ม apply แบบ inline สำหรับ agent ที่ไม่ได้เชื่อมต่อ**
- **เพิ่มประเภท agent cokacdir**

### v0.1.16 (2026-03-25)
- **ซิงค์โมเดลสองทิศทาง** สำหรับ Cline และ Claude Code

---

*สำหรับข้อมูล API เพิ่มเติม โปรดดูที่ [API.md](API.md)*
