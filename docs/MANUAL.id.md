# Panduan Pengguna wall-vault
*(Last updated: 2026-04-16 — v0.2.2)*

---

## Daftar Isi

1. [Apa itu wall-vault?](#apa-itu-wall-vault)
2. [Instalasi](#instalasi)
3. [Memulai (wizard setup)](#memulai)
4. [Pendaftaran API Key](#pendaftaran-api-key)
5. [Penggunaan Proxy](#penggunaan-proxy)
6. [Dashboard Brankas Kunci](#dashboard-brankas-kunci)
7. [Mode Terdistribusi (Multi Bot)](#mode-terdistribusi-multi-bot)
8. [Pengaturan Auto-Start](#pengaturan-auto-start)
9. [Doctor (Dokter)](#doctor-dokter)
10. [Penghematan Token RTK](#penghematan-token-rtk)
11. [Referensi Variabel Lingkungan](#referensi-variabel-lingkungan)
12. [Pemecahan Masalah](#pemecahan-masalah)

---

## Catatan Peningkatan v0.2

- `Service` mendapatkan `default_model` dan `allowed_models`. Model default per-layanan kini diatur langsung pada kartu layanan.
- `Client.default_service` / `default_model` telah diubah nama dan diinterpretasi ulang menjadi `preferred_service` / `model_override`. Jika override kosong, model default dari layanan digunakan.
- Saat startup v0.2 pertama kali, `vault.json` yang ada secara otomatis dimigrasikan, dan keadaan pre-migrasi disimpan sebagai `vault.json.pre-v02.{timestamp}.bak`.
- Dashboard telah distruktur ulang menjadi tiga zona: bilah samping kiri, grid kartu pusat, dan panel edit geser di sisi kanan.
- Jalur Admin API tidak berubah, tetapi skema badan permintaan/respons telah diperbarui — skrip CLI lama akan perlu diperbarui sesuai kebutuhan.

---

## Fitur Baru v0.2.1

- **Pass-through multimoda (OpenAI → Gemini)**: `/v1/chat/completions` kini menerima enam jenis bagian konten selain `text` — `input_audio`, `input_video`, `input_image`, `input_file`, dan `image_url` (data URI dan URL http(s) eksternal ≤ 5 MB). Proxy mengonversi masing-masing menjadi `inlineData` Gemini. Klien yang kompatibel dengan OpenAI seperti EconoWorld dapat melakukan streaming blob audio / gambar / video secara langsung.
- **Jenis agen EconoWorld**: `POST /agent/apply` dengan `agentType: "econoworld"` menulis pengaturan wall-vault ke dalam `analyzer/ai_config.json` proyek. `workDir` menerima daftar jalur kandidat yang dipisahkan koma dan mengonversi jalur drive Windows menjadi jalur mount WSL.
- **Grid kunci dashboard + CRUD**: 11 kunci dirender sebagai kartu ringkas dengan slideover + tambah / ✕ hapus.
- **Penambahan layanan + penyusunan ulang seret-dan-lepas**: grid layanan mendapatkan tombol + tambah dan pegangan seret (`⋮⋮`).
- **Header / footer / animasi tema / pengalih bahasa** dipulihkan. Ketujuh tema (cherry/dark/light/ocean/gold/autumn/winter) memainkan efek partikelnya pada lapisan di belakang kartu namun di atas latar belakang.
- **UX penutupan slideover**: klik di luar atau Esc menutup slideover.
- **Indikator status SSE + pengatur waktu aktif**: di bar atas (topbar), di samping pemilih bahasa/tema, terdapat penghitung `⏱ uptime` dan indikator `● SSE` (hijau = terhubung, oranye = menyambung ulang, abu-abu = terputus) yang diletakkan berdampingan (dipindahkan dari footer ke header sejak v0.2.18 — status bisa dilihat tanpa menggulir).

---

## v0.2.2 Stability & UX Improvements

- **Dispatch fast-skip**: cloud services whose keys are all on cooldown or exhausted are no longer force-retried. Dispatch moves to the next fallback immediately. Per-request tail latency dropped from ~15 s to ~1.5 s.
- **Fallback model swap**: each fallback step now applies the target service's own `default_model`. Previously a `gemini-2.5-flash` request would be handed to Anthropic/Ollama verbatim and rejected (400/404).
- **Anthropic credit-balance handling**: when Anthropic returns HTTP 400 with a "credit balance" body, the proxy promotes it to 402-equivalent and sets a 30 min cooldown so subsequent dispatches skip Anthropic automatically.
- **Service edit default_model dropdown polish**:
  - The server now renders the complete model list (Google 15, OpenRouter 345, etc.) into the `<select>` from the first open — no second round-trip required.
  - `↓ Move to Allowed` button demotes the current default into the allowed_models textarea and clears the default.
  - `✕ Clear` empties the default in place.
  - Collapsible `Custom input` details block lets you type a model ID directly when the dropdown is unreachable.
- **Agent edit/create model_override dropdown**: free text replaced by a `<select>` populated from the preferred service's `default_model` + `allowed_models`. Changing the preferred service auto-repopulates the override options.
- **ClientInput v0.2 fields**: POST `/admin/clients` now accepts v0.2 canonical `preferred_service` / `model_override` alongside legacy `default_service` / `default_model` (legacy is a fallback).

---

## Apa itu wall-vault?

**wall-vault = Proxy AI + Brankas API Key untuk OpenClaw**

Untuk menggunakan layanan AI, Anda memerlukan **API key**. API key ibarat **kartu akses digital** yang membuktikan bahwa "orang ini berhak menggunakan layanan ini." Namun, kartu akses ini memiliki batas penggunaan harian, dan jika tidak dikelola dengan baik, ada risiko bocor.

wall-vault menyimpan kartu-kartu akses ini dalam brankas yang aman dan berperan sebagai **proxy (perantara)** antara OpenClaw dan layanan AI. Singkatnya, OpenClaw cukup terhubung ke wall-vault, dan sisanya wall-vault yang mengurus semuanya.

Masalah yang diselesaikan wall-vault:

- **Rotasi API key otomatis**: Ketika penggunaan suatu key mencapai batas atau diblokir sementara (cooldown), secara diam-diam beralih ke key berikutnya. OpenClaw terus berjalan tanpa gangguan.
- **Fallback layanan otomatis**: Jika Google tidak merespons, otomatis beralih ke OpenRouter, jika itu juga tidak bisa maka ke Ollama/LM Studio/vLLM (AI lokal) yang terinstal di komputer Anda. Sesi tidak terputus. Saat layanan asli pulih, otomatis kembali mulai dari permintaan berikutnya (v0.1.18+, LM Studio/vLLM: v0.1.21+).
- **Sinkronisasi real-time (SSE)**: Saat mengganti model di dashboard brankas, dalam 1-3 detik langsung tercermin di layar OpenClaw. SSE (Server-Sent Events) adalah teknologi di mana server mendorong pembaruan secara real-time ke klien.
- **Notifikasi real-time**: Event seperti kehabisan key atau gangguan layanan langsung ditampilkan di bagian bawah OpenClaw TUI (layar terminal).

> 💡 **Claude Code, Cursor, VS Code** juga bisa dihubungkan, tapi tujuan utama wall-vault adalah digunakan bersama OpenClaw.

```
OpenClaw (layar terminal TUI)
        │
        ▼
  wall-vault proxy (:56244)   ← manajemen key, routing, fallback, event
        │
        ├─ Google Gemini API
        ├─ OpenRouter API (340+ model)
        ├─ Ollama / LM Studio / vLLM (komputer Anda, pilihan terakhir)
        └─ OpenAI / Anthropic API
```

---

## Instalasi

### Linux / macOS

Buka terminal dan tempel perintah di bawah ini.

```bash
# Linux (PC biasa, server — amd64)
curl -L https://github.com/sookmook/wall-vault/releases/latest/download/wall-vault-linux-amd64 \
  -o ~/.local/bin/wall-vault && chmod +x ~/.local/bin/wall-vault

# macOS Apple Silicon (Mac M1/M2/M3)
curl -L https://github.com/sookmook/wall-vault/releases/latest/download/wall-vault-darwin-arm64 \
  -o /usr/local/bin/wall-vault && chmod +x /usr/local/bin/wall-vault
```

- `curl -L ...` — Mengunduh file dari internet.
- `chmod +x` — Membuat file yang diunduh "dapat dieksekusi". Jika langkah ini dilewati, akan muncul error "izin ditolak".

### Windows

Buka PowerShell (Administrator) dan jalankan perintah berikut.

```powershell
# Unduh
Invoke-WebRequest -Uri `
  "https://github.com/sookmook/wall-vault/releases/latest/download/wall-vault-windows-amd64.exe" `
  -OutFile "$env:LOCALAPPDATA\Programs\wall-vault\wall-vault.exe"

# Tambahkan ke PATH (berlaku setelah PowerShell di-restart)
$env:PATH += ";$env:LOCALAPPDATA\Programs\wall-vault"
```

> 💡 **Apa itu PATH?** Daftar folder tempat komputer mencari perintah. Dengan menambahkan ke PATH, Anda bisa menjalankan `wall-vault` dari folder mana saja.

### Build dari Sumber (untuk developer)

Hanya berlaku jika environment pengembangan bahasa Go sudah terinstal.

```bash
git clone https://github.com/sookmook/wall-vault
cd wall-vault
make build       # bin/wall-vault (versi: v0.1.25.YYYYMMDD.HHmmss)
make install     # ~/.local/bin/wall-vault
```

> 💡 **Versi timestamp build**: Saat build dengan `make build`, versi otomatis dihasilkan dalam format yang menyertakan tanggal dan waktu seperti `v0.1.27.20260409`. Jika build langsung dengan `go build ./...`, versi hanya menampilkan `"dev"`.

---

## Memulai

### Menjalankan wizard setup

Setelah instalasi, pertama kali Anda **wajib** menjalankan **wizard setup** dengan perintah berikut. Wizard akan menanyakan item yang diperlukan satu per satu dan memandu Anda.

```bash
wall-vault setup
```

Langkah-langkah yang dilakukan wizard:

```
1. Pilih bahasa (10 bahasa termasuk bahasa Indonesia)
2. Pilih tema (light / dark / gold / cherry / ocean)
3. Mode operasi — penggunaan sendiri (standalone) atau beberapa mesin bersama (distributed)
4. Nama bot — nama yang ditampilkan di dashboard
5. Pengaturan port — default: proxy 56244, brankas 56243 (tekan enter jika tidak perlu diubah)
6. Pilih layanan AI — Google / OpenRouter / Ollama / LM Studio / vLLM
7. Pengaturan filter keamanan tool
8. Pengaturan token admin — kata sandi untuk mengunci fitur manajemen dashboard. Bisa di-auto-generate
9. Kata sandi enkripsi API key — untuk menyimpan key lebih aman (opsional)
10. Path file konfigurasi
```

> ⚠️ **Pastikan Anda mengingat token admin.** Diperlukan nanti saat menambahkan key atau mengubah pengaturan di dashboard. Jika lupa, Anda harus mengedit file konfigurasi secara manual.

Setelah wizard selesai, file konfigurasi `wall-vault.yaml` akan dibuat secara otomatis.

### Menjalankan

```bash
wall-vault start
```

Dua server berikut akan dimulai bersamaan:

- **Proxy** (`https://localhost:56244`) — Perantara yang menghubungkan OpenClaw dan layanan AI
- **Brankas Kunci** (`https://localhost:56243`) — Manajemen API key dan dashboard web

Buka `https://localhost:56243` di browser untuk langsung melihat dashboard.

---

## Pendaftaran API Key

Ada empat cara mendaftarkan API key. **Untuk pemula, disarankan Metode 1 (variabel lingkungan).**

### Metode 1: Variabel Lingkungan (Disarankan — Paling Sederhana)

Variabel lingkungan adalah **nilai yang telah ditetapkan sebelumnya** yang dibaca program saat dimulai. Ketik di terminal seperti berikut.

```bash
# Daftarkan Google Gemini key
export WV_KEY_GOOGLE=AIzaSy...

# Daftarkan OpenRouter key
export WV_KEY_OPENROUTER=sk-or-v1-...

# Jalankan setelah pendaftaran
wall-vault start
```

Jika Anda memiliki beberapa key, hubungkan dengan koma (,). wall-vault akan menggunakan key secara bergantian otomatis (round robin):

```bash
export WV_KEY_GOOGLE=AIzaSy...,AIzaSy...,AIzaSy...
```

> 💡 **Tips**: Perintah `export` hanya berlaku untuk sesi terminal saat ini. Agar tetap berlaku setelah komputer di-restart, tambahkan baris di atas ke file `~/.bashrc` atau `~/.zshrc`.

### Metode 2: Dashboard UI (Klik dengan Mouse)

1. Buka `https://localhost:56243` di browser
2. Klik tombol `[+ Tambah]` di kartu **🔑 API Key** di bagian atas
3. Masukkan jenis layanan, nilai key, label (nama memo), batas harian, lalu simpan

### Metode 3: REST API (untuk Otomasi/Skrip)

REST API adalah cara program bertukar data melalui HTTP. Berguna untuk pendaftaran otomatis via skrip.

```bash
curl -X POST https://localhost:56243/admin/keys \
  -H "Authorization: Bearer token-admin" \
  -H "Content-Type: application/json" \
  -d '{
    "service": "google",
    "key": "AIzaSy...",
    "label": "Key utama",
    "daily_limit": 1000
  }'
```

### Metode 4: Flag proxy (untuk Tes Singkat)

Untuk memasukkan key sementara dan menguji tanpa pendaftaran resmi. Hilang saat program ditutup.

```bash
wall-vault proxy --key-google=AIzaSy... --key-openrouter=sk-or-...
```

---

## Penggunaan Proxy

### Penggunaan di OpenClaw (Tujuan Utama)

Cara mengatur agar OpenClaw terhubung ke layanan AI melalui wall-vault:

Buka file `~/.openclaw/openclaw.json` dan tambahkan konten berikut:

```json5
// ~/.openclaw/openclaw.json
{
  models: {
    providers: {
      "wall-vault": {
        baseUrl: "https://localhost:56244/v1",
        apiKey: "your-agent-token",   // token agen vault
        api: "openai-completions",
        models: [
          { id: "wall-vault/gemini-2.5-flash" },
          { id: "wall-vault/gemini-2.5-pro" },
          { id: "wall-vault/hunter-alpha" },    // gratis 1M context
          { id: "wall-vault/claude-opus-4-6" }
        ]
      }
    }
  }
}
```

> 💡 **Cara lebih mudah**: Tekan tombol **🦞 Salin Pengaturan OpenClaw** di kartu agen dashboard, snippet dengan token dan alamat yang sudah terisi akan disalin ke clipboard. Tinggal tempel saja.

**`wall-vault/` di depan nama model akan terhubung ke mana?**

Dengan melihat nama model, wall-vault secara otomatis menentukan ke layanan AI mana permintaan dikirim:

| Format Model | Layanan Terhubung |
|----------|--------------|
| `wall-vault/gemini-*` | Google Gemini langsung |
| `wall-vault/gpt-*`, `wall-vault/o3`, `wall-vault/o4*` | OpenAI langsung |
| `wall-vault/claude-*` | Anthropic melalui OpenRouter |
| `wall-vault/hunter-alpha`, `wall-vault/healer-alpha` | OpenRouter (gratis 1 juta token context) |
| `wall-vault/kimi-*`, `wall-vault/glm-*`, `wall-vault/deepseek-*` | Koneksi OpenRouter |
| `google/nama-model`, `openai/nama-model`, `anthropic/nama-model` dll | Koneksi langsung ke layanan terkait |
| `custom/google/nama-model`, `custom/openai/nama-model` dll | Hapus bagian `custom/` dan re-routing |
| `nama-model:cloud` | Hapus bagian `:cloud` dan koneksi OpenRouter |

> 💡 **Apa itu context?** Jumlah percakapan yang bisa diingat AI sekaligus. Dengan 1M (satu juta token), percakapan sangat panjang atau dokumen panjang pun bisa diproses sekaligus.

### Koneksi Langsung Format Gemini API (Kompatibilitas Tool yang Ada)

Jika Anda memiliki tool yang langsung menggunakan Google Gemini API, cukup ubah alamatnya ke wall-vault:

```bash
export ANTHROPIC_BASE_URL=https://localhost:56244/google
```

Atau untuk tool yang menentukan URL secara langsung:

```
https://localhost:56244/google/v1beta/models/gemini-2.5-flash:generateContent
```

### Penggunaan di OpenAI SDK (Python)

wall-vault juga bisa dihubungkan dalam kode Python yang menggunakan AI. Cukup ubah `base_url`:

```python
from openai import OpenAI

client = OpenAI(
    base_url="https://localhost:56244/v1",
    api_key="not-needed"  # API key dikelola wall-vault secara otomatis
)

response = client.chat.completions.create(
    model="google/gemini-2.5-flash",   # masukkan dalam format provider/model
    messages=[{"role": "user", "content": "Halo"}]
)
```

### Mengganti Model Saat Berjalan

Untuk mengganti model AI yang digunakan saat wall-vault sudah berjalan:

```bash
# Minta langsung ke proxy untuk mengganti model
curl -X PUT https://localhost:56244/api/config/model \
  -H "Content-Type: application/json" \
  -d '{"service": "openrouter", "model": "anthropic/claude-3.5-sonnet"}'

# Dalam mode terdistribusi (multi bot), ubah dari server brankas → langsung diterapkan via SSE
curl -X PUT https://localhost:56243/admin/clients/id-bot-saya \
  -H "Authorization: Bearer token-admin" \
  -H "Content-Type: application/json" \
  -d '{"default_service": "google", "default_model": "gemini-2.5-pro"}'
```

### Melihat Daftar Model yang Tersedia

```bash
# Lihat daftar lengkap
curl https://localhost:56244/api/models | python3 -m json.tool

# Hanya model Google
curl "https://localhost:56244/api/models?service=google"

# Cari berdasarkan nama (contoh: model yang mengandung "claude")
curl "https://localhost:56244/api/models?q=claude"
```

**Ringkasan Model Utama per Layanan:**

| Layanan | Model Utama |
|--------|----------|
| Google | gemini-2.5-pro, gemini-2.5-flash, gemini-2.5-flash-8b, gemini-2.0-flash |
| OpenAI | gpt-4o, gpt-4o-mini, o3, o1, o1-mini |
| OpenRouter | 346+ (Hunter Alpha 1M context gratis, DeepSeek R1/V3, Qwen 2.5 dll) |
| Ollama | Auto-deteksi server lokal yang terinstal di komputer Anda |
| LM Studio | Server lokal komputer Anda (port 1234) |
| vLLM | Server lokal komputer Anda (port 8000) |
| llama.cpp | Server lokal komputer Anda (port 8080) |

---

## Dashboard Brankas Kunci

Buka `https://localhost:56243` di browser untuk melihat dashboard.

**Tata Letak Layar:**
- **Bar atas tetap (topbar)**: Logo, pemilih bahasa/tema, status koneksi SSE
- **Grid Kartu**: Kartu agen, layanan, API key tersusun dalam bentuk tile

### Kartu API Key

Kartu untuk mengelola API key yang terdaftar dalam satu pandangan.

- Menampilkan daftar key berdasarkan layanan.
- `today_usage`: Jumlah token (karakter yang dibaca dan ditulis AI) yang berhasil diproses hari ini
- `today_attempts`: Total jumlah panggilan hari ini (termasuk sukses + gagal)
- Daftarkan key baru dengan tombol `[+ Tambah]`, hapus key dengan `✕`.

> 💡 **Apa itu token?** Unit yang digunakan AI untuk memproses teks. Kira-kira setara dengan satu kata bahasa Inggris, atau 1-2 karakter bahasa Indonesia. Biaya API biasanya dihitung berdasarkan jumlah token ini.

### Kartu Agen

Kartu yang menampilkan status bot (agen) yang terhubung ke proxy wall-vault.

**Status koneksi ditampilkan dalam 4 tingkat:**

| Simbol | Status | Arti |
|------|------|------|
| 🟢 | Berjalan | Proxy beroperasi normal |
| 🟡 | Lambat | Respons datang tapi lambat |
| 🔴 | Offline | Proxy tidak merespons |
| ⚫ | Belum terhubung/Nonaktif | Proxy belum pernah terhubung ke brankas atau dinonaktifkan |

**Panduan Tombol Bawah Kartu Agen:**

Saat mendaftarkan agen, jika **jenis agen** ditentukan, tombol praktis yang sesuai dengan jenis tersebut otomatis muncul.

---

#### 🔘 Tombol Salin Pengaturan — Membuat pengaturan koneksi secara otomatis

Saat tombol diklik, snippet pengaturan dengan token agen, alamat proxy, informasi model yang sudah terisi disalin ke clipboard. Tempel konten yang disalin ke lokasi dalam tabel di bawah untuk menyelesaikan pengaturan koneksi.

| Tombol | Jenis Agen | Lokasi Tempel |
|------|-------------|-------------|
| 🦞 Salin Pengaturan OpenClaw | `openclaw` | `~/.openclaw/openclaw.json` |
| 🦀 Salin Pengaturan NanoClaw | `nanoclaw` | `~/.openclaw/openclaw.json` |
| 🟠 Salin Pengaturan Claude Code | `claude-code` | `~/.claude/settings.json` |
| ⌨ Salin Pengaturan Cursor | `cursor` | Cursor → Settings → AI |
| 💻 Salin Pengaturan VSCode | `vscode` | `~/.continue/config.json` |

**Contoh — Jika tipe Claude Code, konten seperti ini yang disalin:**

```json
// ~/.claude/settings.json
{
  "apiProvider": "openai",
  "baseUrl": "http://192.168.1.20:56244/v1",
  "apiKey": "token-agen-ini"
}
```

**Contoh — Jika tipe VSCode (Continue):**

```yaml
# ~/.continue/config.yaml  ← tempel di config.yaml, bukan config.json
name: My Config
version: 0.0.1
schema: v1

models:
  - name: wall-vault proxy
    provider: openai
    model: gemini-2.5-flash
    apiBase: http://192.168.1.20:56244/v1
    apiKey: token-agen-ini
    roles:
      - chat
      - edit
      - apply
```

> ⚠️ **Versi terbaru Continue menggunakan `config.yaml`.** Jika `config.yaml` ada, `config.json` sepenuhnya diabaikan. Pastikan tempel di `config.yaml`.

**Contoh — Jika tipe Cursor:**

```
Base URL : http://192.168.1.20:56244/v1
API Key  : token-agen-ini

// atau variabel lingkungan:
OPENAI_BASE_URL=http://192.168.1.20:56244/v1
OPENAI_API_KEY=token-agen-ini
```

> ⚠️ **Jika salin clipboard tidak berfungsi**: Kebijakan keamanan browser mungkin memblokir penyalinan. Jika popup textbox muncul, pilih semua dengan Ctrl+A lalu salin dengan Ctrl+C.

---

#### ⚡ Tombol Auto-Apply — Satu klik, pengaturan selesai

Jika jenis agen adalah `cline`, `claude-code`, `openclaw`, `nanoclaw`, tombol **⚡ Terapkan Pengaturan** ditampilkan di kartu agen. Menekan tombol ini secara otomatis memperbarui file pengaturan lokal agen tersebut.

| Tombol | Jenis Agen | File yang Diterapkan |
|------|-------------|-------------|
| ⚡ Terapkan Pengaturan Cline | `cline` | `~/.cline/data/globalState.json` + `secrets.json` |
| ⚡ Terapkan Pengaturan Claude Code | `claude-code` | `~/.claude/settings.json` |
| ⚡ Terapkan Pengaturan OpenClaw | `openclaw` | `~/.openclaw/openclaw.json` |
| ⚡ Terapkan Pengaturan NanoClaw | `nanoclaw` | `~/.openclaw/openclaw.json` |

> ⚠️ Tombol ini mengirim permintaan ke **localhost:56244** (proxy lokal). Proxy harus berjalan di mesin tersebut agar berfungsi.

---

#### 🔀 Pengurutan Kartu Drag and Drop (v0.1.17, ditingkatkan v0.1.25)

Kartu agen di dashboard dapat **di-drag** untuk diatur ulang sesuai urutan yang diinginkan.

1. Pegang area **lampu lalu lintas (●)** di kiri atas kartu dengan mouse dan drag
2. Lepaskan di atas kartu posisi yang diinginkan, urutan akan berubah

> 💡 Body kartu (field input, tombol dll) tidak bisa di-drag. Hanya bisa dipegang dari area lampu lalu lintas.

#### 🟠 Deteksi Proses Agen (v0.1.25)

Saat proxy beroperasi normal tapi proses agen lokal (NanoClaw, OpenClaw) mati, lampu lalu lintas kartu berubah **oranye (berkedip)** dan pesan "Proses agen berhenti" ditampilkan.

- 🟢 Hijau: Proxy + agen normal
- 🟠 Oranye (berkedip): Proxy normal, agen mati
- 🔴 Merah: Proxy offline
3. Urutan yang diubah **langsung disimpan di server** dan tetap ada setelah refresh

> 💡 Belum didukung di perangkat sentuh (mobile/tablet). Gunakan browser desktop.

---

#### 🔄 Sinkronisasi Model Dua Arah (v0.1.16)

Saat mengganti model agen di dashboard brankas, pengaturan lokal agen tersebut otomatis diperbarui.

**Untuk Cline:**
- Ubah model di brankas → event SSE → proxy memperbarui field model di `globalState.json`
- Target pembaruan: `actModeOpenAiModelId`, `planModeOpenAiModelId`, `openAiModelId`
- `openAiBaseUrl` dan API key tidak disentuh
- **Diperlukan reload VS Code (`Ctrl+Alt+R` atau `Ctrl+Shift+P` → `Developer: Reload Window`)**
  - Karena Cline tidak membaca ulang file pengaturan saat berjalan

**Untuk Claude Code:**
- Ubah model di brankas → event SSE → proxy memperbarui field `model` di `settings.json`
- Otomatis mencari path WSL dan Windows (`~/.claude/`, `/mnt/c/Users/*/.claude/`)

**Arah sebaliknya (agen → brankas):**
- Agen (Cline, Claude Code dll) mengirim permintaan ke proxy, proxy menyertakan informasi layanan/model klien dalam heartbeat
- Layanan/model yang sedang digunakan ditampilkan secara real-time di kartu agen dashboard brankas

> 💡 **Poin penting**: Proxy mengidentifikasi agen dari token Authorization permintaan dan otomatis me-routing ke layanan/model yang diatur di brankas. Meskipun Cline atau Claude Code mengirim nama model yang berbeda, proxy meng-override dengan pengaturan brankas.

---

### Menggunakan Cline di VS Code — Panduan Detail

#### Langkah 1: Instal Cline

Instal **Cline** (ID: `saoudrizwan.claude-dev`) dari VS Code Extension Marketplace.

#### Langkah 2: Daftarkan Agen di Brankas

1. Buka dashboard brankas (`http://IP-brankas:56243`)
2. Klik **+ Tambah** di bagian **Agen**
3. Masukkan sebagai berikut:

| Field | Nilai | Keterangan |
|------|----|------|
| ID | `cline_saya` | Identifier unik (huruf Inggris, tanpa spasi) |
| Nama | `Cline Saya` | Nama yang ditampilkan di dashboard |
| Jenis Agen | `cline` | ← wajib pilih `cline` |
| Layanan | Pilih layanan yang akan digunakan (contoh: `google`) | |
| Model | Masukkan model yang akan digunakan (contoh: `gemini-2.5-flash`) | |

4. Tekan **Simpan**, token otomatis dihasilkan

#### Langkah 3: Hubungkan ke Cline

**Metode A — Auto-apply (Disarankan)**

1. Pastikan **proxy** wall-vault berjalan di mesin tersebut (`localhost:56244`)
2. Klik tombol **⚡ Terapkan Pengaturan Cline** di kartu agen dashboard
3. Jika muncul notifikasi "Pengaturan berhasil diterapkan!", berarti sukses
4. Reload VS Code (`Ctrl+Alt+R`)

**Metode B — Pengaturan Manual**

Buka pengaturan (⚙️) di sidebar Cline:
- **API Provider**: `OpenAI Compatible`
- **Base URL**: `http://alamat-proxy:56244/v1`
  - Jika di mesin yang sama: `https://localhost:56244/v1`
  - Jika di mesin lain seperti server mini: `http://192.168.1.20:56244/v1`
- **API Key**: Token yang diterbitkan dari brankas (salin dari kartu agen)
- **Model ID**: Model yang diatur di brankas (contoh: `gemini-2.5-flash`)

#### Langkah 4: Verifikasi

Kirim pesan apa saja di jendela chat Cline. Jika normal:
- Kartu agen tersebut di dashboard brankas akan menampilkan **titik hijau (● Berjalan)**
- Layanan/model saat ini ditampilkan di kartu (contoh: `google / gemini-2.5-flash`)

#### Mengganti Model

Saat ingin mengganti model Cline, ubah dari **dashboard brankas**:

1. Ubah dropdown layanan/model di kartu agen
2. Klik **Terapkan**
3. Reload VS Code (`Ctrl+Alt+R`) — nama model di footer Cline akan diperbarui
4. Permintaan berikutnya akan menggunakan model baru

> 💡 Sebenarnya proxy mengidentifikasi permintaan Cline berdasarkan token dan me-routing ke model pengaturan brankas. Meskipun tidak reload VS Code, **model yang benar-benar digunakan langsung berubah** — reload hanya untuk memperbarui tampilan model di UI Cline.

#### Deteksi Pemutusan Koneksi

Saat VS Code ditutup, di dashboard brankas sekitar **90 detik** kemudian kartu agen menjadi kuning (lambat), **3 menit** kemudian merah (offline). (Sejak v0.1.18, pengecekan status interval 15 detik membuat deteksi offline lebih cepat.)

#### Pemecahan Masalah

| Gejala | Penyebab | Solusi |
|------|------|------|
| Error "koneksi gagal" di Cline | Proxy tidak berjalan atau alamat salah | Periksa proxy dengan `curl https://localhost:56244/health` |
| Titik hijau tidak muncul di brankas | API key (token) belum diatur | Klik tombol **⚡ Terapkan Pengaturan Cline** lagi |
| Model footer Cline tidak berubah | Cline meng-cache pengaturan | Reload VS Code (`Ctrl+Alt+R`) |
| Nama model yang salah ditampilkan | Bug lama (diperbaiki di v0.1.16) | Perbarui proxy ke v0.1.16+ |

---

#### 🟣 Tombol Salin Perintah Deploy — Untuk menginstal di mesin baru

Digunakan saat pertama kali menginstal proxy wall-vault di komputer baru dan menghubungkan ke brankas. Saat tombol diklik, seluruh skrip instalasi disalin. Tempel di terminal komputer baru dan jalankan, yang berikut ini diproses sekaligus:

1. Instal binary wall-vault (dilewati jika sudah terinstal)
2. Pendaftaran otomatis layanan user systemd
3. Mulai layanan dan koneksi otomatis ke brankas

> 💡 Token agen dan alamat server brankas sudah terisi dalam skrip, jadi setelah ditempel bisa langsung dijalankan tanpa modifikasi.

---

### Kartu Layanan

Kartu untuk mengaktifkan/menonaktifkan atau mengatur layanan AI yang digunakan.

- Toggle switch aktif/nonaktif per layanan
- Jika alamat server AI lokal (Ollama, LM Studio, vLLM, llama.cpp dll yang berjalan di komputer Anda) dimasukkan, model yang tersedia ditemukan secara otomatis.
- **Tampilan status koneksi layanan lokal**: Titik ● di samping nama layanan berwarna **hijau** jika terhubung, **abu-abu** jika tidak terhubung
- **Lampu lalu lintas otomatis layanan lokal** (v0.1.23+): Layanan lokal (Ollama, LM Studio, vLLM, llama.cpp) otomatis aktif saat bisa terhubung, otomatis nonaktif saat terputus. Cara yang sama seperti auto-toggle berbasis API key pada layanan cloud (Google, OpenRouter dll).
- **Toggle mode reasoning** (v0.2.17+): Checkbox **mode reasoning** muncul di bagian bawah jendela edit layanan lokal. Jika diaktifkan, proxy menambahkan `"reasoning": true` ke body chat-completions yang dikirim ke upstream, sehingga model yang mendukung output proses berpikir seperti DeepSeek R1, Qwen QwQ mengembalikan blok `<think>…</think>` bersama respons. Server yang tidak mengenali field ini akan mengabaikannya, jadi aman untuk tetap diaktifkan bahkan pada workload campuran.

> 💡 **Jika layanan lokal berjalan di komputer lain**: Masukkan IP komputer tersebut di kolom URL layanan. Contoh: `http://192.168.1.20:11434` (Ollama), `http://192.168.1.20:1234` (LM Studio), `http://192.168.1.20:8080` (llama.cpp). Jika layanan hanya bind ke `127.0.0.1` bukan `0.0.0.0`, koneksi via IP eksternal tidak akan berfungsi, periksa alamat bind di pengaturan layanan.

### Input Token Admin

Saat mencoba menggunakan fitur penting seperti menambah/menghapus key di dashboard, popup input token admin muncul. Masukkan token yang diatur di wizard setup. Setelah dimasukkan sekali, tetap berlaku sampai browser ditutup.

> ⚠️ **Jika kegagalan autentikasi melebihi 10 kali dalam 15 menit, IP tersebut akan diblokir sementara.** Jika lupa token, periksa item `admin_token` di file `wall-vault.yaml`.

---

## Mode Terdistribusi (Multi Bot)

Konfigurasi untuk **berbagi satu brankas kunci** saat menjalankan OpenClaw di beberapa komputer secara bersamaan. Nyaman karena manajemen kunci cukup dari satu tempat.

### Contoh Konfigurasi

```
[Server Brankas Kunci]
  wall-vault vault    (brankas kunci :56243, dashboard)

[WSL Alpha]            [Raspberry Pi Gamma]    [Mac Mini Lokal]
  wall-vault proxy      wall-vault proxy        wall-vault proxy
  openclaw TUI          openclaw TUI            openclaw TUI
  ↕ sinkronisasi SSE    ↕ sinkronisasi SSE      ↕ sinkronisasi SSE
```

Semua bot melihat server brankas di tengah, jadi saat mengganti model atau menambah key di brankas, langsung diterapkan ke semua bot.

### Langkah 1: Mulai Server Brankas Kunci

Jalankan di komputer yang akan digunakan sebagai server brankas:

```bash
wall-vault vault
```

### Langkah 2: Daftarkan Setiap Bot (Klien)

Daftarkan informasi setiap bot yang terhubung ke server brankas terlebih dahulu:

```bash
curl -X POST https://localhost:56243/admin/clients \
  -H "Authorization: Bearer token-admin" \
  -H "Content-Type: application/json" \
  -d '{
    "id": "botA",
    "name": "Bot A",
    "token": "bota-secret",
    "default_service": "google",
    "default_model": "gemini-2.5-flash"
  }'
```

### Langkah 3: Mulai Proxy di Setiap Komputer Bot

Jalankan proxy dengan menentukan alamat server brankas dan token di setiap komputer tempat bot terinstal:

```bash
WV_VAULT_URL=http://192.168.x.x:56243 \
WV_VAULT_TOKEN=bota-secret \
WV_VAULT_CLIENT_ID=botA \
wall-vault proxy
```

> 💡 Ganti bagian **`192.168.x.x`** dengan alamat IP internal aktual komputer server brankas. Bisa diperiksa dari pengaturan router atau perintah `ip addr`.

---

## Pengaturan Auto-Start

Jika merepotkan harus menjalankan wall-vault secara manual setiap kali komputer di-restart, daftarkan sebagai system service. Setelah didaftarkan sekali, otomatis dimulai saat boot.

### Linux — systemd (Kebanyakan Linux)

systemd adalah sistem yang mengelola dan memulai program secara otomatis di Linux:

```bash
wall-vault doctor deploy
systemctl --user daemon-reload
systemctl --user enable --now wall-vault
```

Lihat log:

```bash
journalctl --user -u wall-vault -f
```

### macOS — launchd

Sistem yang bertanggung jawab atas eksekusi otomatis program di macOS:

```bash
wall-vault doctor deploy launchd
launchctl load ~/Library/LaunchAgents/com.wall-vault.plist
```

### Windows — NSSM

1. Unduh NSSM dari [nssm.cc](https://nssm.cc/download) dan tambahkan ke PATH.
2. Di PowerShell Administrator:

```powershell
wall-vault doctor deploy windows
```

---

## Doctor (Dokter)

Perintah `doctor` adalah alat untuk **mendiagnosis dan memperbaiki sendiri** apakah wall-vault dikonfigurasi dengan benar.

```bash
wall-vault doctor check   # Diagnosis status saat ini (hanya baca, tidak mengubah apa pun)
wall-vault doctor fix     # Perbaiki masalah secara otomatis
wall-vault doctor all     # Diagnosis + perbaikan otomatis sekaligus
```

> 💡 Jika ada yang terasa aneh, jalankan `wall-vault doctor all` terlebih dahulu. Banyak masalah terdeteksi secara otomatis.

---

## Penghematan Token RTK

*(v0.1.24+)*

**RTK (Alat Penghematan Token)** secara otomatis mengompresi output perintah shell yang dijalankan oleh agen coding AI (Claude Code dll) untuk mengurangi penggunaan token. Misalnya, output 15 baris dari `git status` diringkas menjadi 2 baris.

### Penggunaan Dasar

```bash
# Bungkus perintah dengan wall-vault rtk, output otomatis difilter
wall-vault rtk git status          # Hanya daftar file yang berubah
wall-vault rtk git diff HEAD~1     # Baris yang berubah + context minimal
wall-vault rtk git log -10         # Hash + pesan satu baris
wall-vault rtk go test ./...       # Hanya tampilkan test yang gagal
wall-vault rtk ls -la              # Perintah yang tidak didukung otomatis dipotong
```

### Perintah yang Didukung dan Efek Penghematan

| Perintah | Metode Filter | Tingkat Penghematan |
|------|----------|--------|
| `git status` | Hanya ringkasan file yang berubah | ~87% |
| `git diff` | Baris yang berubah + 3 baris context | ~60-94% |
| `git log` | Hash + pesan baris pertama | ~90% |
| `git push/pull/fetch` | Hapus progress, hanya ringkasan | ~80% |
| `go test` | Hanya tampilkan gagal, hitung yang lulus | ~88-99% |
| `go build/vet` | Hanya error | ~90% |
| Semua perintah lainnya | 50 baris pertama + 50 baris terakhir, maks 32KB | Bervariasi |

### Pipeline Filter 3 Tahap

1. **Filter struktur per perintah** — Memahami format output git, go dll dan mengekstrak bagian yang bermakna
2. **Post-processing regex** — Hapus kode warna ANSI, kurangi baris kosong, kumpulkan baris duplikat
3. **Passthrough + Pemotongan** — Untuk perintah yang tidak didukung, simpan 50 baris pertama/terakhir

### Integrasi Claude Code

Semua perintah shell bisa diatur untuk otomatis melewati RTK menggunakan hook `PreToolUse` Claude Code.

```bash
# Instal hook (otomatis ditambahkan ke settings.json Claude Code)
wall-vault rtk hook install
```

Atau tambahkan secara manual ke `~/.claude/settings.json`:

```json
{
  "hooks": {
    "PreToolUse": [{
      "matcher": "Bash",
      "command": "wall-vault rtk rewrite"
    }]
  }
}
```

> 💡 **Preservasi exit code**: RTK mengembalikan exit code perintah asli apa adanya. Jika perintah gagal (exit code ≠ 0), AI juga mendeteksi kegagalan dengan akurat.

> 💡 **Paksa bahasa Inggris**: RTK menjalankan perintah dengan `LC_ALL=C` agar selalu menghasilkan output bahasa Inggris terlepas dari pengaturan bahasa sistem. Ini agar filter bekerja dengan akurat.

---

## Referensi Variabel Lingkungan

Variabel lingkungan adalah cara meneruskan nilai pengaturan ke program. Ketik di terminal dalam format `export nama-variabel=nilai`, atau letakkan di file layanan auto-start agar selalu diterapkan.

| Variabel | Keterangan | Contoh Nilai |
|------|------|---------|
| `WV_LANG` | Bahasa dashboard | `ko`, `en`, `ja` |
| `WV_THEME` | Tema dashboard | `light`, `dark`, `gold` |
| `WV_KEY_GOOGLE` | API key Google (beberapa dengan koma) | `AIza...,AIza...` |
| `WV_KEY_OPENROUTER` | API key OpenRouter | `sk-or-v1-...` |
| `WV_VAULT_URL` | Alamat server brankas dalam mode terdistribusi | `http://192.168.x.x:56243` |
| `WV_VAULT_TOKEN` | Token autentikasi klien (bot) | `my-secret-token` |
| `WV_ADMIN_TOKEN` | Token admin | `admin-token-here` |
| `WV_MASTER_PASS` | Kata sandi enkripsi API key | `my-password` |
| `WV_AVATAR` | Path file gambar avatar (path relatif dari `~/.openclaw/`) | `workspace/avatars/avatar.png` |
| `OLLAMA_URL` | Alamat server lokal Ollama | `http://192.168.x.x:11434` |

---

## Pemecahan Masalah

### Proxy Tidak Bisa Dimulai

Penyebab umum adalah port sudah digunakan oleh program lain.

```bash
ss -tlnp | grep 56244   # Periksa siapa yang menggunakan port 56244
wall-vault proxy --port 8080   # Mulai dengan nomor port lain
```

### Error API Key (429, 402, 401, 403, 582)

| Kode Error | Arti | Penanganan |
|----------|------|----------|
| **429** | Terlalu banyak permintaan (batas penggunaan terlampaui) | Tunggu sebentar atau tambah key lain |
| **402** | Pembayaran diperlukan atau kredit habis | Isi ulang kredit di layanan terkait |
| **401 / 403** | Key salah atau tidak ada izin | Periksa ulang nilai key dan daftarkan kembali |
| **582** | Gateway overload (cooldown 5 menit) | Otomatis dilepas setelah 5 menit |

```bash
# Periksa daftar dan status key terdaftar
curl -H "Authorization: Bearer token-admin" https://localhost:56243/admin/keys

# Reset counter penggunaan key
curl -X POST -H "Authorization: Bearer token-admin" https://localhost:56243/admin/keys/reset
```

### Agen Menampilkan "Belum Terhubung"

"Belum terhubung" berarti proses proxy tidak mengirim sinyal (heartbeat) ke brankas. **Ini bukan berarti pengaturan tidak tersimpan.** Proxy harus berjalan dengan mengetahui alamat server brankas dan token agar menjadi status terhubung.

```bash
# Mulai proxy dengan menentukan alamat server brankas, token, ID klien
WV_VAULT_URL=http://alamat-server-brankas:56243 \
WV_VAULT_TOKEN=token-klien \
WV_VAULT_CLIENT_ID=ID-klien \
wall-vault proxy
```

Jika koneksi berhasil, dalam sekitar 20 detik akan muncul 🟢 Berjalan di dashboard.

### Ollama Tidak Bisa Terhubung

Ollama adalah program yang menjalankan AI langsung di komputer Anda. Pertama periksa apakah Ollama menyala.

```bash
curl http://localhost:11434/api/tags   # Jika daftar model muncul, normal
export OLLAMA_URL=http://192.168.x.x:11434   # Jika berjalan di komputer lain
```

> ⚠️ Jika Ollama tidak merespons, mulai Ollama terlebih dahulu dengan perintah `ollama serve`.

> ⚠️ **Model besar lambat**: Model besar seperti `qwen3.5:35b`, `deepseek-r1` bisa memakan waktu beberapa menit untuk menghasilkan respons. Meskipun terlihat tidak ada respons, proses mungkin masih berjalan normal, harap tunggu.

---

## Perubahan Terbaru (v0.1.16 ~ v0.1.27)

### v0.1.27 (2026-04-09)
- **Perbaikan nama model fallback Ollama**: Saat fallback dari layanan lain ke Ollama, nama model dengan prefix provider (contoh: `google/gemini-3.1-pro-preview`) dikirim apa adanya ke Ollama. Sekarang otomatis diganti dengan variabel lingkungan/model default.
- **Pengurangan drastis waktu cooldown**: 429 rate limit 30 menit→5 menit, 402 pembayaran 1 jam→30 menit, 401/403 24 jam→6 jam. Mencegah situasi di mana semua key cooldown bersamaan dan proxy lumpuh total.
- **Coba paksa saat semua cooldown**: Saat semua key dalam status cooldown, coba paksa dengan key yang paling cepat selesai cooldown untuk mencegah penolakan permintaan.
- **Perbaikan tampilan daftar layanan**: Respons `/status` menampilkan daftar layanan aktual yang disinkronisasi dari vault (mencegah absennya anthropic dll).

### v0.1.25 (2026-04-08)
- **Deteksi proses agen**: Proxy mendeteksi apakah agen lokal (NanoClaw/OpenClaw) masih hidup dan menampilkan lampu lalu lintas oranye di dashboard.
- **Perbaikan drag handle**: Saat mengurutkan kartu, hanya bisa dipegang dari area lampu lalu lintas (●). Tidak ada drag tidak sengaja dari field input atau tombol.

### v0.1.24 (2026-04-06)
- **Subperintah penghematan token RTK**: `wall-vault rtk <command>` otomatis memfilter output perintah shell, mengurangi penggunaan token agen AI sebesar 60-90%. Filter bawaan untuk perintah utama seperti git, go, perintah yang tidak didukung juga otomatis dipotong. Integrasi transparan melalui hook `PreToolUse` Claude Code.

### v0.1.23 (2026-04-06)
- **Perbaikan perubahan model Ollama**: Model Ollama yang diubah dari dashboard brankas tidak benar-benar diterapkan ke proxy. Sebelumnya hanya menggunakan variabel lingkungan (`OLLAMA_MODEL`), sekarang pengaturan brankas yang diprioritaskan.
- **Lampu lalu lintas otomatis layanan lokal**: Ollama/LM Studio/vLLM otomatis aktif saat bisa terhubung, otomatis nonaktif saat terputus. Sama seperti auto-toggle berbasis key layanan cloud.

### v0.1.22 (2026-04-05)
- **Perbaikan field content kosong yang hilang**: Saat model thinking (gemini-3.1-pro, o1, claude thinking dll) menghabiskan batas max_tokens pada reasoning dan tidak bisa membuat respons aktual, proxy menghilangkan field `content`/`text` di JSON respons dengan `omitempty` menyebabkan klien SDK OpenAI/Anthropic crash dengan error `Cannot read properties of undefined (reading 'trim')`. Diubah agar selalu menyertakan field sesuai spesifikasi API resmi.

### v0.1.21 (2026-04-05)
- **Dukungan model Gemma 4**: Model seri Gemma seperti `gemma-4-31b-it`, `gemma-4-26b-a4b-it` dll bisa digunakan melalui Google Gemini API.
- **Dukungan resmi layanan LM Studio / vLLM**: Sebelumnya layanan ini terlewat di routing proxy dan selalu digantikan Ollama. Sekarang routing normal via API kompatibel OpenAI.
- **Perbaikan tampilan layanan dashboard**: Meskipun terjadi fallback, dashboard selalu menampilkan layanan yang diatur pengguna.
- **Tampilan status layanan lokal**: Saat dashboard dimuat, status koneksi layanan lokal (Ollama, LM Studio, vLLM dll) ditampilkan dengan warna titik ●.
- **Variabel lingkungan filter tool**: Mode penerusan tool bisa diatur dengan variabel lingkungan `WV_TOOL_FILTER=passthrough`.

### v0.1.20 (2026-03-28)
- **Penguatan keamanan komprehensif**: Pencegahan XSS (41 lokasi), perbandingan token waktu konstan, pembatasan CORS, batas ukuran permintaan, pencegahan traversal path, autentikasi SSE, penguatan rate limiter dll 12 item keamanan ditingkatkan.

### v0.1.19 (2026-03-27)
- **Deteksi Claude Code online**: Claude Code yang tidak melalui proxy juga ditampilkan online di dashboard.

### v0.1.18 (2026-03-26)
- **Perbaikan layanan fallback macet**: Setelah fallback ke Ollama karena error sementara, otomatis kembali saat layanan asli pulih.
- **Peningkatan deteksi offline**: Pengecekan status interval 15 detik membuat deteksi penghentian proxy lebih cepat.

### v0.1.17 (2026-03-25)
- **Pengurutan kartu drag and drop**: Kartu agen bisa diseret untuk mengubah urutan.
- **Tombol apply pengaturan inline**: Tombol [⚡ Terapkan Pengaturan] ditampilkan di agen offline.
- **Jenis agen cokacdir ditambahkan.**

### v0.1.16 (2026-03-25)
- **Sinkronisasi model dua arah**: Mengubah model Cline/Claude Code dari dashboard brankas otomatis diterapkan.

---

*Untuk informasi API lebih detail, lihat [API.md](API.md).*
