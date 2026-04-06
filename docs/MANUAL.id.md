# Panduan Pengguna wall-vault
*(Last updated: 2026-04-06 — v0.1.24)*

---

## Daftar Isi

1. [Apa itu wall-vault?](#apa-itu-wall-vault)
2. [Instalasi](#instalasi)
3. [Memulai pertama kali (wizard setup)](#memulai-pertama-kali)
4. [Pendaftaran kunci API](#pendaftaran-kunci-api)
5. [Cara menggunakan proxy](#cara-menggunakan-proxy)
6. [Dashboard brankas kunci](#dashboard-brankas-kunci)
7. [Mode terdistribusi (multi bot)](#mode-terdistribusi-multi-bot)
8. [Pengaturan start otomatis](#pengaturan-start-otomatis)
9. [Doctor (diagnosis)](#doctor-diagnosis)
10. [RTK Penghematan token](#rtk-penghematan-token)
11. [Referensi variabel lingkungan](#referensi-variabel-lingkungan)
12. [Pemecahan masalah](#pemecahan-masalah)

---

## Apa itu wall-vault?

**wall-vault = Proxy AI + Brankas kunci API untuk OpenClaw**

Untuk menggunakan layanan AI, Anda membutuhkan **kunci API**. Kunci API seperti **kartu akses digital** yang membuktikan bahwa "orang ini berhak menggunakan layanan ini". Namun, kartu akses ini memiliki batas penggunaan harian dan berisiko terekspos jika tidak dikelola dengan baik.

wall-vault menyimpan kartu-kartu akses ini di brankas yang aman dan berperan sebagai **proxy (perantara)** antara OpenClaw dan layanan AI. Sederhananya, OpenClaw hanya perlu terhubung ke wall-vault, dan wall-vault yang mengurus sisanya.

Masalah yang diselesaikan wall-vault:

- **Rotasi kunci API otomatis**: Ketika satu kunci mencapai batas penggunaan atau diblokir sementara (cooldown), ia diam-diam beralih ke kunci berikutnya. OpenClaw terus berjalan tanpa gangguan.
- **Penggantian layanan otomatis (fallback)**: Jika Google tidak merespons, beralih ke OpenRouter; jika itu juga gagal, otomatis beralih ke AI lokal yang terinstal di komputer Anda (Ollama, LM Studio, vLLM). Sesi tidak terputus. Ketika layanan asli pulih, otomatis kembali dari permintaan berikutnya (v0.1.18+, LM Studio/vLLM: v0.1.21+).
- **Sinkronisasi real-time (SSE)**: Ketika Anda mengubah model di dashboard brankas, perubahan tercermin di layar OpenClaw dalam 1-3 detik. SSE (Server-Sent Events) adalah teknologi yang memungkinkan server mengirimkan pembaruan ke klien secara real-time.
- **Notifikasi real-time**: Event seperti kehabisan kunci atau gangguan layanan langsung ditampilkan di bagian bawah TUI (layar terminal) OpenClaw.

> 💡 **Claude Code, Cursor, VS Code** juga bisa dihubungkan, tetapi tujuan utama wall-vault adalah digunakan bersama OpenClaw.

```
OpenClaw (antarmuka TUI di terminal)
        │
        ▼
  wall-vault proxy (:56244)   ← manajemen kunci, routing, fallback, event
        │
        ├─ Google Gemini API
        ├─ OpenRouter API (lebih dari 340 model)
        ├─ Ollama / LM Studio / vLLM (komputer Anda, pilihan terakhir)
        └─ OpenAI / Anthropic API
```

---

## Instalasi

### Linux / macOS

Buka terminal dan tempelkan perintah berikut:

```bash
# Linux (PC biasa, server — amd64)
curl -L https://github.com/sookmook/wall-vault/releases/latest/download/wall-vault-linux-amd64 \
  -o ~/.local/bin/wall-vault && chmod +x ~/.local/bin/wall-vault

# macOS Apple Silicon (M1/M2/M3 Mac)
curl -L https://github.com/sookmook/wall-vault/releases/latest/download/wall-vault-darwin-arm64 \
  -o /usr/local/bin/wall-vault && chmod +x /usr/local/bin/wall-vault
```

- `curl -L ...` — Mengunduh file dari internet.
- `chmod +x` — Membuat file yang diunduh menjadi "dapat dieksekusi". Jika langkah ini dilewati, akan muncul error "izin ditolak".

### Windows

Buka PowerShell (sebagai administrator) dan jalankan perintah berikut:

```powershell
# Unduh
Invoke-WebRequest -Uri `
  "https://github.com/sookmook/wall-vault/releases/latest/download/wall-vault-windows-amd64.exe" `
  -OutFile "$env:LOCALAPPDATA\Programs\wall-vault\wall-vault.exe"

# Tambahkan ke PATH (berlaku setelah restart PowerShell)
$env:PATH += ";$env:LOCALAPPDATA\Programs\wall-vault"
```

> 💡 **Apa itu PATH?** Ini adalah daftar folder tempat komputer mencari perintah. Dengan menambahkan ke PATH, Anda dapat menjalankan `wall-vault` dari folder mana pun hanya dengan mengetikkan namanya.

### Build dari source code (untuk developer)

Hanya berlaku jika lingkungan pengembangan bahasa Go sudah terinstal.

```bash
git clone https://github.com/sookmook/wall-vault
cd wall-vault
make build       # bin/wall-vault (versi: v0.1.24.YYYYMMDD.HHmmss)
make install     # ~/.local/bin/wall-vault
```

> 💡 **Versi dengan timestamp build**: Ketika build dengan `make build`, versi otomatis dibuat dalam format yang menyertakan tanggal dan waktu, seperti `v0.1.24.20260406.225957`. Jika build langsung dengan `go build ./...`, versi hanya akan ditampilkan sebagai `"dev"`.

---

## Memulai pertama kali

### Menjalankan wizard setup

Setelah instalasi, wajib jalankan **wizard pengaturan** dengan perintah berikut. Wizard akan memandu Anda dengan menanyakan setiap item yang diperlukan satu per satu.

```bash
wall-vault setup
```

Langkah-langkah yang dilalui wizard:

```
1. Pemilihan bahasa (10 bahasa termasuk Korea)
2. Pemilihan tema (light / dark / gold / cherry / ocean)
3. Mode operasi — penggunaan sendiri (standalone) atau berbagi di beberapa mesin (distributed)
4. Nama bot — nama yang ditampilkan di dashboard
5. Pengaturan port — default: proxy 56244, brankas 56243 (tekan Enter jika tidak perlu mengubah)
6. Pemilihan layanan AI — Google / OpenRouter / Ollama / LM Studio / vLLM
7. Pengaturan filter keamanan tools
8. Pengaturan token admin — kata sandi yang mengunci fungsi manajemen dashboard. Pembuatan otomatis tersedia
9. Pengaturan kata sandi enkripsi kunci API — untuk penyimpanan yang lebih aman (opsional)
10. Path penyimpanan file konfigurasi
```

> ⚠️ **Pastikan untuk mengingat token admin.** Ini akan diperlukan nanti untuk menambahkan kunci atau mengubah pengaturan di dashboard. Jika hilang, Anda harus mengedit file konfigurasi secara manual.

Setelah wizard selesai, file konfigurasi `wall-vault.yaml` akan dibuat secara otomatis.

### Menjalankan

```bash
wall-vault start
```

Dua server berikut dimulai secara bersamaan:

- **Proxy** (`http://localhost:56244`) — perantara yang menghubungkan OpenClaw dengan layanan AI
- **Brankas kunci** (`http://localhost:56243`) — manajemen kunci API dan dashboard web

Buka `http://localhost:56243` di browser untuk langsung mengakses dashboard.

---

## Pendaftaran kunci API

Ada empat cara untuk mendaftarkan kunci API. **Untuk pemula, disarankan metode 1 (variabel lingkungan).**

### Metode 1: Variabel lingkungan (disarankan — paling sederhana)

Variabel lingkungan adalah **nilai yang sudah dikonfigurasi sebelumnya** yang dibaca program saat dimulai. Masukkan di terminal seperti berikut:

```bash
# Mendaftarkan kunci Google Gemini
export WV_KEY_GOOGLE=AIzaSy...

# Mendaftarkan kunci OpenRouter
export WV_KEY_OPENROUTER=sk-or-v1-...

# Jalankan setelah pendaftaran
wall-vault start
```

Jika Anda memiliki beberapa kunci, hubungkan dengan koma (,). wall-vault akan menggunakannya secara otomatis bergantian (round robin):

```bash
export WV_KEY_GOOGLE=AIzaSy...,AIzaSy...,AIzaSy...
```

> 💡 **Tips**: Perintah `export` hanya berlaku untuk sesi terminal saat ini. Agar tetap ada setelah restart komputer, tambahkan baris di atas ke file `~/.bashrc` atau `~/.zshrc`.

### Metode 2: UI Dashboard (klik dengan mouse)

1. Akses `http://localhost:56243` di browser
2. Klik tombol `[+ Tambah]` di kartu **🔑 Kunci API** di bagian atas
3. Masukkan jenis layanan, nilai kunci, label (nama referensi), dan batas harian, lalu simpan

### Metode 3: REST API (untuk otomasi/skrip)

REST API adalah cara program bertukar data melalui HTTP. Berguna untuk pendaftaran otomatis melalui skrip.

```bash
curl -X POST http://localhost:56243/admin/keys \
  -H "Authorization: Bearer token-admin" \
  -H "Content-Type: application/json" \
  -d '{
    "service": "google",
    "key": "AIzaSy...",
    "label": "Kunci utama",
    "daily_limit": 1000
  }'
```

### Metode 4: Flag proxy (untuk tes cepat)

Gunakan untuk menguji sementara dengan memasukkan kunci tanpa pendaftaran formal. Kunci akan hilang saat program ditutup.

```bash
wall-vault proxy --key-google=AIzaSy... --key-openrouter=sk-or-...
```

---

## Cara menggunakan proxy

### Penggunaan dengan OpenClaw (tujuan utama)

Berikut cara mengatur OpenClaw agar terhubung ke layanan AI melalui wall-vault.

Buka file `~/.openclaw/openclaw.json` dan tambahkan konten berikut:

```json5
// ~/.openclaw/openclaw.json
{
  models: {
    providers: {
      "wall-vault": {
        baseUrl: "http://localhost:56244/v1",
        apiKey: "your-agent-token",   // token agen brankas
        api: "openai-completions",
        models: [
          { id: "wall-vault/gemini-2.5-flash" },
          { id: "wall-vault/gemini-2.5-pro" },
          { id: "wall-vault/hunter-alpha" },    // context 1M gratis
          { id: "wall-vault/claude-opus-4-6" }
        ]
      }
    }
  }
}
```

> 💡 **Cara lebih mudah**: Tekan tombol **🦞 Salin konfigurasi OpenClaw** di kartu agen di dashboard dan snippet dengan token dan alamat yang sudah terisi akan disalin ke clipboard. Tinggal tempel saja.

**Ke mana `wall-vault/` di awal nama model mengarah?**

wall-vault secara otomatis menentukan layanan AI mana yang akan menerima permintaan berdasarkan nama model:

| Format model | Layanan yang terhubung |
|----------|--------------|
| `wall-vault/gemini-*` | Koneksi langsung ke Google Gemini |
| `wall-vault/gpt-*`, `wall-vault/o3`, `wall-vault/o4*` | Koneksi langsung ke OpenAI |
| `wall-vault/claude-*` | Koneksi ke Anthropic melalui OpenRouter |
| `wall-vault/hunter-alpha`, `wall-vault/healer-alpha` | OpenRouter (1 juta token context gratis) |
| `wall-vault/kimi-*`, `wall-vault/glm-*`, `wall-vault/deepseek-*` | Koneksi melalui OpenRouter |
| `google/nama-model`, `openai/nama-model`, `anthropic/nama-model` dll. | Koneksi langsung ke layanan terkait |
| `custom/google/nama-model`, `custom/openai/nama-model` dll. | Menghapus bagian `custom/` dan merutekan ulang |
| `nama-model:cloud` | Menghapus bagian `:cloud` dan menghubungkan melalui OpenRouter |

> 💡 **Apa itu context?** Ini adalah jumlah percakapan yang dapat "diingat" AI sekaligus. 1M (satu juta token) berarti percakapan yang sangat panjang atau dokumen besar dapat diproses sekaligus.

### Koneksi langsung dalam format Gemini API (kompatibilitas dengan tools yang ada)

Jika Anda memiliki tools yang sudah menggunakan Google Gemini API secara langsung, cukup ubah alamatnya ke wall-vault:

```bash
export ANTHROPIC_BASE_URL=http://localhost:56244/google
```

Atau untuk tools yang menentukan URL secara langsung:

```
http://localhost:56244/google/v1beta/models/gemini-2.5-flash:generateContent
```

### Penggunaan dengan OpenAI SDK (Python)

Anda juga dapat menghubungkan wall-vault dalam kode Python yang menggunakan AI. Cukup ubah `base_url`:

```python
from openai import OpenAI

client = OpenAI(
    base_url="http://localhost:56244/v1",
    api_key="not-needed"  # Kunci API dikelola otomatis oleh wall-vault
)

response = client.chat.completions.create(
    model="google/gemini-2.5-flash",   # Masukkan dalam format provider/model
    messages=[{"role": "user", "content": "Halo"}]
)
```

### Mengubah model saat berjalan

Untuk mengubah model AI saat wall-vault sudah berjalan:

```bash
# Ubah model dengan mengirim permintaan langsung ke proxy
curl -X PUT http://localhost:56244/api/config/model \
  -H "Content-Type: application/json" \
  -d '{"service": "openrouter", "model": "anthropic/claude-3.5-sonnet"}'

# Dalam mode terdistribusi (multi bot), ubah dari server brankas → langsung tercermin melalui SSE
curl -X PUT http://localhost:56243/admin/clients/id-bot-saya \
  -H "Authorization: Bearer token-admin" \
  -H "Content-Type: application/json" \
  -d '{"default_service": "google", "default_model": "gemini-2.5-pro"}'
```

### Memeriksa daftar model yang tersedia

```bash
# Lihat daftar lengkap
curl http://localhost:56244/api/models | python3 -m json.tool

# Lihat hanya model Google
curl "http://localhost:56244/api/models?service=google"

# Cari berdasarkan nama (contoh: model yang mengandung "claude")
curl "http://localhost:56244/api/models?q=claude"
```

**Ringkasan model utama per layanan:**

| Layanan | Model utama |
|--------|----------|
| Google | gemini-2.5-pro, gemini-2.5-flash, gemini-2.5-flash-8b, gemini-2.0-flash |
| OpenAI | gpt-4o, gpt-4o-mini, o3, o1, o1-mini |
| OpenRouter | Lebih dari 346 (Hunter Alpha 1M context gratis, DeepSeek R1/V3, Qwen 2.5 dll.) |
| Ollama | Deteksi otomatis server lokal yang terinstal di komputer |
| LM Studio | Server lokal di komputer (port 1234) |
| vLLM | Server lokal di komputer (port 8000) |

---

## Dashboard brankas kunci

Akses `http://localhost:56243` di browser untuk melihat dashboard.

**Struktur layar:**
- **Bar atas tetap (topbar)**: Logo, pemilih bahasa dan tema, indikator status koneksi SSE
- **Grid kartu**: Kartu agen, layanan, dan kunci API disusun dalam bentuk ubin

### Kartu kunci API

Kartu untuk mengelola semua kunci API yang terdaftar dalam satu pandangan.

- Menampilkan daftar kunci yang dikelompokkan berdasarkan layanan.
- `today_usage`: Token (jumlah karakter yang dibaca dan ditulis AI) yang berhasil diproses hari ini
- `today_attempts`: Total jumlah panggilan hari ini (termasuk sukses + gagal)
- Tombol `[+ Tambah]` untuk mendaftarkan kunci baru dan `✕` untuk menghapusnya.

> 💡 **Apa itu token?** Ini adalah unit yang digunakan AI untuk memproses teks. Kira-kira setara dengan satu kata dalam bahasa Inggris atau 1-2 karakter Korea. Biaya API biasanya dihitung berdasarkan jumlah token ini.

### Kartu agen

Kartu yang menampilkan status bot (agen) yang terhubung ke proxy wall-vault.

**Status koneksi ditampilkan dalam 4 level:**

| Indikator | Status | Arti |
|------|------|------|
| 🟢 | Berjalan | Proxy berjalan normal |
| 🟡 | Tertunda | Merespons tapi lambat |
| 🔴 | Offline | Proxy tidak merespons |
| ⚫ | Tidak terhubung/Nonaktif | Proxy belum pernah terhubung ke brankas atau nonaktif |

**Panduan tombol di bagian bawah kartu agen:**

Saat mendaftarkan agen, jika Anda menentukan **jenis agen**, tombol kenyamanan yang sesuai dengan jenis tersebut akan muncul secara otomatis.

---

#### 🔘 Tombol salin konfigurasi — Membuat pengaturan koneksi secara otomatis

Saat tombol diklik, snippet konfigurasi dengan token, alamat proxy, dan informasi model yang sudah terisi disalin ke clipboard. Cukup tempel konten yang disalin di lokasi yang ditunjukkan dalam tabel di bawah untuk menyelesaikan pengaturan koneksi.

| Tombol | Jenis agen | Lokasi tempel |
|------|-------------|-------------|
| 🦞 Salin konfigurasi OpenClaw | `openclaw` | `~/.openclaw/openclaw.json` |
| 🦀 Salin konfigurasi NanoClaw | `nanoclaw` | `~/.openclaw/openclaw.json` |
| 🟠 Salin konfigurasi Claude Code | `claude-code` | `~/.claude/settings.json` |
| ⌨ Salin konfigurasi Cursor | `cursor` | Cursor → Settings → AI |
| 💻 Salin konfigurasi VSCode | `vscode` | `~/.continue/config.json` |

**Contoh — Jika jenisnya Claude Code, konten berikut yang disalin:**

```json
// ~/.claude/settings.json
{
  "apiProvider": "openai",
  "baseUrl": "http://192.168.1.20:56244/v1",
  "apiKey": "token-agen-ini"
}
```

**Contoh — Jika jenisnya VSCode (Continue):**

```yaml
# ~/.continue/config.yaml  ← Tempel di config.yaml, bukan config.json
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

> ⚠️ **Versi terbaru Continue menggunakan `config.yaml`.** Jika `config.yaml` ada, `config.json` akan sepenuhnya diabaikan. Pastikan menempel di `config.yaml`.

**Contoh — Jika jenisnya Cursor:**

```
Base URL : http://192.168.1.20:56244/v1
API Key  : token-agen-ini

// Atau melalui variabel lingkungan:
OPENAI_BASE_URL=http://192.168.1.20:56244/v1
OPENAI_API_KEY=token-agen-ini
```

> ⚠️ **Ketika salin ke clipboard tidak berfungsi**: Kebijakan keamanan browser mungkin memblokir penyalinan. Jika muncul kotak teks di popup, pilih semua dengan Ctrl+A dan salin dengan Ctrl+C.

---

#### ⚡ Tombol terapkan otomatis — Tekan sekali dan pengaturan selesai

Untuk agen dengan jenis `cline`, `claude-code`, `openclaw`, atau `nanoclaw`, tombol **⚡ Terapkan konfigurasi** ditampilkan di kartu agen. Menekan tombol ini akan memperbarui file konfigurasi lokal agen secara otomatis.

| Tombol | Jenis agen | File target |
|------|-------------|-------------|
| ⚡ Terapkan konfigurasi Cline | `cline` | `~/.cline/data/globalState.json` + `secrets.json` |
| ⚡ Terapkan konfigurasi Claude Code | `claude-code` | `~/.claude/settings.json` |
| ⚡ Terapkan konfigurasi OpenClaw | `openclaw` | `~/.openclaw/openclaw.json` |
| ⚡ Terapkan konfigurasi NanoClaw | `nanoclaw` | `~/.openclaw/openclaw.json` |

> ⚠️ Tombol ini mengirim permintaan ke **localhost:56244** (proxy lokal). Proxy harus berjalan di mesin tersebut agar berfungsi.

---

#### 🔀 Pengurutan kartu drag and drop (v0.1.17)

Anda dapat **menyeret** kartu agen di dashboard untuk menyusun ulang sesuai keinginan.

1. Klik dan tahan kartu agen dengan mouse lalu seret
2. Lepaskan di atas kartu pada posisi yang diinginkan untuk menukar urutan
3. Urutan yang diubah **langsung disimpan di server** dan tetap ada meski halaman di-refresh

> 💡 Perangkat sentuh (ponsel/tablet) belum didukung. Gunakan di browser desktop.

---

#### 🔄 Sinkronisasi model dua arah (v0.1.16)

Ketika Anda mengubah model agen di dashboard brankas, pengaturan lokal agen diperbarui secara otomatis.

**Untuk Cline:**
- Ubah model di brankas → event SSE → proxy memperbarui field model di `globalState.json`
- Field yang diperbarui: `actModeOpenAiModelId`, `planModeOpenAiModelId`, `openAiModelId`
- `openAiBaseUrl` dan kunci API tidak diubah
- **Perlu reload VS Code (`Ctrl+Alt+R` atau `Ctrl+Shift+P` → `Developer: Reload Window`)**
  - Karena Cline tidak membaca ulang file konfigurasi saat berjalan

**Untuk Claude Code:**
- Ubah model di brankas → event SSE → proxy memperbarui field `model` di `settings.json`
- Pencarian otomatis di kedua path WSL dan Windows (`~/.claude/`, `/mnt/c/Users/*/.claude/`)

**Arah sebaliknya (agen → brankas):**
- Ketika agen (Cline, Claude Code dll.) mengirim permintaan ke proxy, proxy menyertakan informasi layanan/model klien dalam heartbeat
- Layanan/model yang sedang digunakan ditampilkan secara real-time di kartu agen di dashboard brankas

> 💡 **Poin utama**: Proxy mengidentifikasi agen dari token Authorization dalam permintaan dan secara otomatis merutekan ke layanan/model yang dikonfigurasi di brankas. Meskipun Cline atau Claude Code mengirim nama model yang berbeda, proxy menimpanya dengan pengaturan brankas.

---

### Menggunakan Cline di VS Code — Panduan detail

#### Langkah 1: Instal Cline

Instal **Cline** (ID: `saoudrizwan.claude-dev`) dari marketplace ekstensi VS Code.

#### Langkah 2: Daftarkan agen di brankas

1. Buka dashboard brankas (`http://IP-brankas:56243`)
2. Di bagian **Agen**, klik **+ Tambah**
3. Masukkan seperti berikut:

| Field | Nilai | Deskripsi |
|------|----|------|
| ID | `my_cline` | Pengenal unik (dalam bahasa Inggris, tanpa spasi) |
| Nama | `My Cline` | Nama yang ditampilkan di dashboard |
| Jenis agen | `cline` | ← Harus memilih `cline` |
| Layanan | Pilih layanan yang diinginkan (contoh: `google`) | |
| Model | Masukkan model yang diinginkan (contoh: `gemini-2.5-flash`) | |

4. Klik **Simpan** dan token akan dibuat secara otomatis

#### Langkah 3: Hubungkan ke Cline

**Metode A — Terapkan otomatis (disarankan)**

1. Pastikan **proxy** wall-vault berjalan di mesin tersebut (`localhost:56244`)
2. Klik tombol **⚡ Terapkan konfigurasi Cline** di kartu agen di dashboard
3. Jika muncul notifikasi "Konfigurasi berhasil diterapkan!", berarti berhasil
4. Reload VS Code (`Ctrl+Alt+R`)

**Metode B — Pengaturan manual**

Buka pengaturan (⚙️) di sidebar Cline:
- **API Provider**: `OpenAI Compatible`
- **Base URL**: `http://alamat-proxy:56244/v1`
  - Di mesin yang sama: `http://localhost:56244/v1`
  - Di mesin lain seperti server mini: `http://192.168.1.20:56244/v1`
- **API Key**: Token yang diterbitkan dari brankas (salin dari kartu agen)
- **Model ID**: Model yang dikonfigurasi di brankas (contoh: `gemini-2.5-flash`)

#### Langkah 4: Verifikasi

Kirim pesan apa saja di jendela chat Cline. Jika normal:
- Kartu agen yang bersangkutan di dashboard brankas akan menampilkan **titik hijau (● Berjalan)**
- Layanan/model saat ini akan ditampilkan di kartu (contoh: `google / gemini-2.5-flash`)

#### Mengubah model

Ketika ingin mengubah model Cline, lakukan perubahan dari **dashboard brankas**:

1. Ubah dropdown layanan/model di kartu agen
2. Klik **Terapkan**
3. Reload VS Code (`Ctrl+Alt+R`) — nama model di footer Cline akan diperbarui
4. Model baru akan digunakan mulai permintaan berikutnya

> 💡 Sebenarnya, proxy mengidentifikasi permintaan Cline berdasarkan token dan merutekan ke model yang dikonfigurasi di brankas. Bahkan tanpa reload VS Code, **model yang sebenarnya digunakan langsung berubah** — reload hanya untuk memperbarui tampilan model di UI Cline.

#### Deteksi pemutusan koneksi

Saat VS Code ditutup, kartu agen di dashboard brankas berubah menjadi kuning (tertunda) setelah sekitar **90 detik** dan merah (offline) setelah **3 menit**. (Sejak v0.1.18, pengecekan status setiap 15 detik membuat deteksi offline lebih cepat.)

#### Pemecahan masalah

| Gejala | Penyebab | Solusi |
|------|------|------|
| Error "Koneksi gagal" di Cline | Proxy tidak berjalan atau alamat salah | Periksa proxy dengan `curl http://localhost:56244/health` |
| Titik hijau tidak muncul di brankas | Kunci API (token) belum diatur | Klik lagi tombol **⚡ Terapkan konfigurasi Cline** |
| Model di footer Cline tidak berubah | Cline meng-cache pengaturan | Reload VS Code (`Ctrl+Alt+R`) |
| Nama model yang salah ditampilkan | Bug lama (diperbaiki di v0.1.16) | Perbarui proxy ke v0.1.16 atau lebih baru |

---

#### 🟣 Tombol salin perintah deploy — Gunakan saat menginstal di mesin baru

Gunakan saat pertama kali menginstal proxy wall-vault di komputer baru dan menghubungkannya ke brankas. Saat tombol diklik, seluruh skrip instalasi disalin. Tempel di terminal komputer baru dan jalankan untuk memproses semuanya sekaligus:

1. Instalasi binary wall-vault (dilewati jika sudah terinstal)
2. Pendaftaran otomatis layanan pengguna systemd
3. Memulai layanan dan koneksi otomatis ke brankas

> 💡 Skrip sudah berisi token agen ini dan alamat server brankas yang sudah terisi, sehingga dapat langsung dijalankan setelah ditempel tanpa modifikasi tambahan.

---

### Kartu layanan

Kartu untuk mengaktifkan/menonaktifkan dan mengonfigurasi layanan AI.

- Sakelar aktifkan/nonaktifkan per layanan
- Memasukkan alamat server AI lokal (Ollama, LM Studio, vLLM dll. yang berjalan di komputer Anda) akan mendeteksi model yang tersedia secara otomatis.
- **Indikator status koneksi layanan lokal**: Titik ● di samping nama layanan berwarna **hijau** saat terhubung dan **abu-abu** saat tidak terhubung
- **Lampu lalu lintas otomatis layanan lokal** (v0.1.23+): Layanan lokal (Ollama, LM Studio, vLLM) secara otomatis diaktifkan/dinonaktifkan berdasarkan ketersediaan koneksi. Saat layanan diaktifkan, berubah menjadi ● hijau dalam 15 detik dan kotak centang menyala; saat dinonaktifkan, otomatis mati. Bekerja dengan cara yang sama seperti layanan cloud (Google, OpenRouter dll.) yang di-toggle otomatis berdasarkan keberadaan kunci API.

> 💡 **Jika layanan lokal berjalan di komputer lain**: Masukkan IP komputer tersebut di kolom URL layanan. Contoh: `http://192.168.1.20:11434` (Ollama), `http://192.168.1.20:1234` (LM Studio). Jika layanan hanya terikat ke `127.0.0.1` bukan `0.0.0.0`, tidak dapat diakses melalui IP eksternal — periksa alamat binding di pengaturan layanan.

### Input token admin

Saat mencoba menggunakan fungsi penting di dashboard seperti menambah/menghapus kunci, popup input token admin muncul. Masukkan token yang diatur di wizard setup. Setelah dimasukkan sekali, tetap berlaku sampai browser ditutup.

> ⚠️ **Jika percobaan autentikasi gagal melebihi 10 kali dalam 15 menit, IP tersebut akan diblokir sementara.** Jika Anda lupa tokennya, periksa item `admin_token` di file `wall-vault.yaml`.

---

## Mode terdistribusi (multi bot)

Ketika mengoperasikan OpenClaw secara bersamaan di beberapa mesin, ini adalah konfigurasi untuk **berbagi satu brankas kunci**. Praktis karena manajemen kunci dilakukan di satu tempat.

### Contoh konfigurasi

```
[Server brankas kunci]
  wall-vault vault    (brankas kunci :56243, dashboard)

[WSL Alpha]           [Raspberry Pi Gamma]    [Mac Mini Lokal]
  wall-vault proxy      wall-vault proxy        wall-vault proxy
  openclaw TUI          openclaw TUI            openclaw TUI
  ↕ Sinkronisasi SSE    ↕ Sinkronisasi SSE      ↕ Sinkronisasi SSE
```

Semua bot mengarah ke server brankas pusat, sehingga saat Anda mengubah model atau menambah kunci di brankas, perubahan langsung tercermin di semua bot.

### Langkah 1: Mulai server brankas kunci

Jalankan di komputer yang akan digunakan sebagai server brankas:

```bash
wall-vault vault
```

### Langkah 2: Daftarkan setiap bot (klien)

Daftarkan terlebih dahulu informasi setiap bot yang akan terhubung ke server brankas:

```bash
curl -X POST http://localhost:56243/admin/clients \
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

### Langkah 3: Mulai proxy di setiap komputer bot

Di setiap komputer tempat bot terinstal, jalankan proxy dengan menentukan alamat dan token server brankas:

```bash
WV_VAULT_URL=http://192.168.x.x:56243 \
WV_VAULT_TOKEN=bota-secret \
WV_VAULT_CLIENT_ID=botA \
wall-vault proxy
```

> 💡 **`192.168.x.x`** harus diganti dengan alamat IP internal sebenarnya dari komputer server brankas. Dapat diperiksa dari pengaturan router atau perintah `ip addr`.

---

## Pengaturan start otomatis

Jika merepotkan untuk memulai wall-vault secara manual setiap kali restart komputer, daftarkan sebagai layanan sistem. Setelah didaftarkan, otomatis dimulai saat boot.

### Linux — systemd (sebagian besar Linux)

systemd adalah sistem yang memulai dan mengelola program secara otomatis di Linux:

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

Sistem yang bertanggung jawab atas eksekusi program otomatis di macOS:

```bash
wall-vault doctor deploy launchd
launchctl load ~/Library/LaunchAgents/com.wall-vault.plist
```

### Windows — NSSM

1. Unduh NSSM dari [nssm.cc](https://nssm.cc/download) dan tambahkan ke PATH.
2. Di PowerShell sebagai administrator:

```powershell
wall-vault doctor deploy windows
```

---

## Doctor (diagnosis)

Perintah `doctor` adalah alat yang **mendiagnosis dan memperbaiki secara otomatis** apakah wall-vault dikonfigurasi dengan benar.

```bash
wall-vault doctor check   # Diagnosis kondisi saat ini (hanya baca, tidak mengubah apa pun)
wall-vault doctor fix     # Perbaiki masalah secara otomatis
wall-vault doctor all     # Diagnosis + perbaikan otomatis sekaligus
```

> 💡 Jika ada yang terasa aneh, jalankan `wall-vault doctor all` terlebih dahulu. Ini mendeteksi dan memperbaiki banyak masalah secara otomatis.

---

## RTK Penghematan token

*(v0.1.24+)*

**RTK (alat penghematan token)** secara otomatis mengompresi output perintah shell yang dijalankan oleh agen coding AI (seperti Claude Code), mengurangi konsumsi token. Misalnya, output 15 baris dari `git status` diringkas menjadi 2 baris.

### Penggunaan dasar

```bash
# Bungkus perintah dengan wall-vault rtk dan output akan difilter secara otomatis
wall-vault rtk git status          # Hanya menampilkan daftar file yang berubah
wall-vault rtk git diff HEAD~1     # Hanya baris yang berubah + konteks minimal
wall-vault rtk git log -10         # Hash + pesan satu baris
wall-vault rtk go test ./...       # Hanya menampilkan tes yang gagal
wall-vault rtk ls -la              # Perintah tidak didukung dipotong otomatis
```

### Perintah yang didukung dan penghematan

| Perintah | Metode filter | Penghematan |
|------|----------|--------|
| `git status` | Hanya ringkasan file yang berubah | ~87% |
| `git diff` | Baris yang berubah + 3 baris konteks | ~60-94% |
| `git log` | Hash + baris pertama pesan | ~90% |
| `git push/pull/fetch` | Hapus progres, hanya ringkasan | ~80% |
| `go test` | Hanya tampilkan kegagalan, hitung yang lulus | ~88-99% |
| `go build/vet` | Hanya tampilkan error | ~90% |
| Semua perintah lain | 50 baris pertama + 50 baris terakhir, maks 32KB | Variabel |

### Pipeline filter 3 tahap

1. **Filter struktural per perintah** — Memahami format output git, go dll. dan mengekstrak hanya bagian yang bermakna
2. **Post-processing regex** — Hapus kode warna ANSI, kurangi baris kosong, agregasi baris duplikat
3. **Passthrough + pemotongan** — Perintah tidak didukung hanya mempertahankan 50 baris pertama/terakhir

### Integrasi Claude Code

Anda dapat mengonfigurasi hook `PreToolUse` Claude Code agar semua perintah shell secara otomatis melewati RTK.

```bash
# Instal hook (ditambahkan otomatis ke settings.json Claude Code)
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

> 💡 **Preservasi exit code**: RTK mengembalikan kode keluar asli dari perintah. Jika perintah gagal (exit code ≠ 0), AI juga mendeteksi kegagalan dengan benar.

> 💡 **Output dipaksa bahasa Inggris**: RTK menjalankan perintah dengan `LC_ALL=C` untuk selalu menghasilkan output bahasa Inggris, terlepas dari pengaturan bahasa sistem. Ini diperlukan agar filter bekerja dengan benar.

---

## Referensi variabel lingkungan

Variabel lingkungan adalah cara untuk meneruskan nilai konfigurasi ke program. Masukkan di terminal dalam format `export NAMA_VARIABEL=nilai` atau letakkan di file layanan start otomatis untuk penerapan permanen.

| Variabel | Deskripsi | Contoh nilai |
|------|------|---------|
| `WV_LANG` | Bahasa dashboard | `ko`, `en`, `ja` |
| `WV_THEME` | Tema dashboard | `light`, `dark`, `gold` |
| `WV_KEY_GOOGLE` | Kunci API Google (beberapa dipisahkan koma) | `AIza...,AIza...` |
| `WV_KEY_OPENROUTER` | Kunci API OpenRouter | `sk-or-v1-...` |
| `WV_VAULT_URL` | Alamat server brankas dalam mode terdistribusi | `http://192.168.x.x:56243` |
| `WV_VAULT_TOKEN` | Token autentikasi klien (bot) | `my-secret-token` |
| `WV_ADMIN_TOKEN` | Token admin | `admin-token-here` |
| `WV_MASTER_PASS` | Kata sandi enkripsi kunci API | `my-password` |
| `WV_AVATAR` | Path file gambar avatar (path relatif dari `~/.openclaw/`) | `workspace/avatars/avatar.png` |
| `OLLAMA_URL` | Alamat server lokal Ollama | `http://192.168.x.x:11434` |

---

## Pemecahan masalah

### Ketika proxy tidak mau mulai

Biasanya port sudah digunakan oleh program lain.

```bash
ss -tlnp | grep 56244   # Periksa siapa yang menggunakan port 56244
wall-vault proxy --port 8080   # Mulai dengan nomor port lain
```

### Ketika terjadi error kunci API (429, 402, 401, 403, 582)

| Kode error | Arti | Tindakan |
|----------|------|----------|
| **429** | Terlalu banyak permintaan (batas penggunaan terlampaui) | Tunggu sebentar atau tambahkan kunci lain |
| **402** | Pembayaran diperlukan atau kredit tidak cukup | Isi ulang kredit di layanan terkait |
| **401 / 403** | Kunci salah atau tidak ada izin | Periksa ulang nilai kunci dan daftarkan ulang |
| **582** | Overload gateway (cooldown 5 menit) | Otomatis hilang setelah 5 menit |

```bash
# Periksa daftar dan status kunci yang terdaftar
curl -H "Authorization: Bearer token-admin" http://localhost:56243/admin/keys

# Reset counter penggunaan kunci
curl -X POST -H "Authorization: Bearer token-admin" http://localhost:56243/admin/keys/reset
```

### Ketika agen ditampilkan sebagai "tidak terhubung"

"Tidak terhubung" berarti proses proxy tidak mengirim sinyal (heartbeat) ke brankas. **Ini bukan berarti pengaturan tidak tersimpan.** Proxy harus berjalan dengan alamat dan token server brankas agar berubah ke status terhubung.

```bash
# Mulai proxy dengan menentukan alamat server brankas, token, dan ID klien
WV_VAULT_URL=http://alamat-server-brankas:56243 \
WV_VAULT_TOKEN=token-klien \
WV_VAULT_CLIENT_ID=id-klien \
wall-vault proxy
```

Jika koneksi berhasil, status akan berubah menjadi 🟢 Berjalan di dashboard dalam sekitar 20 detik.

### Ketika koneksi Ollama tidak berfungsi

Ollama adalah program yang menjalankan AI langsung di komputer Anda. Pertama, periksa apakah Ollama aktif.

```bash
curl http://localhost:11434/api/tags   # Jika daftar model muncul, berarti normal
export OLLAMA_URL=http://192.168.x.x:11434   # Jika berjalan di komputer lain
```

> ⚠️ Jika Ollama tidak merespons, mulai dulu dengan perintah `ollama serve`.

> ⚠️ **Model besar lambat merespons**: Model besar seperti `qwen3.5:35b`, `deepseek-r1` mungkin memerlukan beberapa menit untuk menghasilkan respons. Meskipun tampak tidak ada respons, mungkin sedang diproses secara normal — harap tunggu.

---

## Perubahan terbaru (v0.1.16 ~ v0.1.24)

### v0.1.24 (2026-04-06)
- **Subperintah RTK penghematan token**: `wall-vault rtk <command>` secara otomatis memfilter output perintah shell untuk mengurangi konsumsi token agen AI sebesar 60-90%. Menyertakan filter khusus untuk perintah utama seperti git dan go, dan secara otomatis memotong perintah yang tidak didukung. Terintegrasi secara transparan dengan hook `PreToolUse` Claude Code.

### v0.1.23 (2026-04-06)
- **Perbaikan perubahan model Ollama**: Diperbaiki masalah di mana perubahan model Ollama di dashboard brankas tidak tercermin di proxy. Sebelumnya hanya menggunakan variabel lingkungan (`OLLAMA_MODEL`), sekarang pengaturan brankas diprioritaskan.
- **Lampu lalu lintas otomatis layanan lokal**: Ollama, LM Studio, dan vLLM secara otomatis diaktifkan saat dapat terhubung dan dinonaktifkan saat terputus. Bekerja dengan cara yang sama seperti toggle otomatis berbasis kunci untuk layanan cloud.

### v0.1.22 (2026-04-05)
- **Perbaikan field content kosong yang hilang**: Ketika model thinking (gemini-3.1-pro, o1, claude thinking dll.) menggunakan semua batas max_tokens untuk reasoning tanpa menghasilkan respons aktual, proxy menghilangkan field `content`/`text` dari JSON respons dengan `omitempty`, menyebabkan klien SDK OpenAI/Anthropic crash dengan error `Cannot read properties of undefined (reading 'trim')`. Diubah agar selalu menyertakan field sesuai spesifikasi API resmi.

### v0.1.21 (2026-04-05)
- **Dukungan model Gemma 4**: Model keluarga Gemma seperti `gemma-4-31b-it`, `gemma-4-26b-a4b-it` dapat digunakan melalui Google Gemini API.
- **Dukungan resmi LM Studio / vLLM**: Sebelumnya layanan ini dihilangkan dari routing proxy dan selalu diganti dengan Ollama. Sekarang dirutekan dengan benar melalui API yang kompatibel dengan OpenAI.
- **Perbaikan tampilan layanan di dashboard**: Meskipun terjadi fallback, dashboard selalu menampilkan layanan yang dikonfigurasi pengguna.
- **Indikator status layanan lokal**: Saat dashboard dimuat, status koneksi layanan lokal (Ollama, LM Studio, vLLM dll.) ditampilkan berdasarkan warna titik ●.
- **Variabel lingkungan filter tools**: Mode penerusan tools dapat diatur dengan variabel lingkungan `WV_TOOL_FILTER=passthrough`.

### v0.1.20 (2026-03-28)
- **Penguatan keamanan komprehensif**: 12 item keamanan ditingkatkan termasuk pencegahan XSS (41 titik), perbandingan token waktu-konstan, pembatasan CORS, batas ukuran permintaan, pencegahan path traversal, autentikasi SSE, penguatan rate limiting dll.

### v0.1.19 (2026-03-27)
- **Deteksi online Claude Code**: Claude Code yang tidak melalui proxy juga ditampilkan sebagai online di dashboard.

### v0.1.18 (2026-03-26)
- **Perbaikan perekatan layanan fallback**: Setelah fallback ke Ollama karena error sementara, otomatis kembali saat layanan asli pulih.
- **Peningkatan deteksi offline**: Deteksi penghentian proxy menjadi lebih cepat dengan pengecekan status setiap 15 detik.

### v0.1.17 (2026-03-25)
- **Pengurutan kartu drag and drop**: Kartu agen dapat disusun ulang dengan menyeret.
- **Tombol terapkan konfigurasi inline**: Tombol [⚡ Terapkan konfigurasi] ditampilkan pada agen offline.
- **Ditambahkan jenis agen cokacdir**.

### v0.1.16 (2026-03-25)
- **Sinkronisasi model dua arah**: Ketika model Cline atau Claude Code diubah di dashboard brankas, otomatis tercermin.

---

*Untuk informasi API yang lebih detail, lihat [API.md](API.md).*
