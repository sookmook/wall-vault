# Panduan Pengguna wall-vault
*(Terakhir diperbarui: 2026-03-20 — v0.1.15)*

---

## Daftar Isi

1. [Apa itu wall-vault?](#apa-itu-wall-vault)
2. [Instalasi](#instalasi)
3. [Memulai Pertama Kali (Wizard setup)](#memulai-pertama-kali)
4. [Mendaftarkan Kunci API](#mendaftarkan-kunci-api)
5. [Cara Menggunakan Proxy](#cara-menggunakan-proxy)
6. [Dashboard Brankas Kunci](#dashboard-brankas-kunci)
7. [Mode Terdistribusi (Multi-Bot)](#mode-terdistribusi-multi-bot)
8. [Pengaturan Mulai Otomatis](#pengaturan-mulai-otomatis)
9. [Doctor — Alat Diagnostik](#doctor--alat-diagnostik)
10. [Referensi Variabel Lingkungan](#referensi-variabel-lingkungan)
11. [Pemecahan Masalah](#pemecahan-masalah)

---

## Apa itu wall-vault?

**wall-vault = Proxy AI + Brankas Kunci API untuk OpenClaw**

Untuk menggunakan layanan AI, Anda membutuhkan **kunci API** — semacam **tanda masuk digital** yang membuktikan bahwa Anda berhak menggunakan layanan tersebut. Kunci ini memiliki batas penggunaan harian, dan jika tidak dijaga dengan baik, bisa bocor ke pihak yang tidak bertanggung jawab.

wall-vault menyimpan semua tanda masuk digital Anda di dalam brankas yang aman, lalu bertindak sebagai **perantara (proxy)** antara OpenClaw dan layanan AI. Singkatnya, OpenClaw hanya perlu terhubung ke wall-vault, dan wall-vault yang mengurus semua hal rumit di balik layar.

Masalah yang diselesaikan wall-vault:

- **Rotasi kunci API otomatis**: Jika satu kunci mencapai batasnya atau sedang diblokir sementara (cooldown), wall-vault secara diam-diam beralih ke kunci berikutnya. OpenClaw terus bekerja tanpa gangguan.
- **Pergantian layanan otomatis (fallback)**: Jika Google tidak merespons, otomatis beralih ke OpenRouter. Jika itu pun gagal, beralih ke Ollama (AI lokal di komputer Anda sendiri). Sesi tidak terputus.
- **Sinkronisasi real-time (SSE)**: Jika Anda mengganti model AI di dashboard brankas, perubahan akan tercermin di layar OpenClaw dalam 1–3 detik. SSE (Server-Sent Events) adalah teknologi di mana server mendorong perubahan ke klien secara real-time.
- **Notifikasi real-time**: Kejadian seperti kunci habis atau gangguan layanan langsung ditampilkan di bagian bawah layar TUI (tampilan terminal) OpenClaw.

> 💡 **Claude Code, Cursor, VS Code** juga bisa dihubungkan ke wall-vault, namun tujuan utama wall-vault adalah digunakan bersama OpenClaw.

```
OpenClaw (tampilan terminal TUI)
        │
        ▼
  wall-vault proxy (:56244)   ← manajemen kunci, routing, fallback, notifikasi
        │
        ├─ Google Gemini API
        ├─ OpenRouter API (lebih dari 340 model)
        └─ Ollama (komputer Anda sendiri, pilihan terakhir)
```

---

## Instalasi

### Linux / macOS

Buka terminal dan tempelkan perintah berikut persis seperti ini.

```bash
# Linux (PC biasa, server — amd64)
curl -L https://github.com/sookmook/wall-vault/releases/latest/download/wall-vault-linux-amd64 \
  -o ~/.local/bin/wall-vault && chmod +x ~/.local/bin/wall-vault

# macOS Apple Silicon (Mac M1/M2/M3)
curl -L https://github.com/sookmook/wall-vault/releases/latest/download/wall-vault-darwin-arm64 \
  -o /usr/local/bin/wall-vault && chmod +x /usr/local/bin/wall-vault
```

- `curl -L ...` — mengunduh file dari internet.
- `chmod +x` — membuat file yang diunduh menjadi "dapat dijalankan". Jika langkah ini dilewati, akan muncul error "izin ditolak".

### Windows

Buka PowerShell (sebagai Administrator) dan jalankan perintah berikut.

```powershell
# Unduh
Invoke-WebRequest -Uri `
  "https://github.com/sookmook/wall-vault/releases/latest/download/wall-vault-windows-amd64.exe" `
  -OutFile "$env:LOCALAPPDATA\Programs\wall-vault\wall-vault.exe"

# Tambahkan ke PATH (berlaku setelah PowerShell dimulai ulang)
$env:PATH += ";$env:LOCALAPPDATA\Programs\wall-vault"
```

> 💡 **Apa itu PATH?** PATH adalah daftar folder tempat komputer mencari perintah. Dengan menambahkan wall-vault ke PATH, Anda bisa mengetik `wall-vault` dari folder mana pun dan langsung menjalankannya.

### Build dari Kode Sumber (untuk Developer)

Hanya berlaku jika Anda sudah menginstal lingkungan pengembangan bahasa Go.

```bash
git clone https://github.com/sookmook/wall-vault
cd wall-vault
make build       # bin/wall-vault (versi: v0.1.6.YYYYMMDD.HHmmss)
make install     # ~/.local/bin/wall-vault
```

> 💡 **Versi dengan timestamp build**: Jika di-build menggunakan `make build`, versi akan dibuat otomatis dalam format seperti `v0.1.6.20260314.231308` yang menyertakan tanggal dan waktu. Jika di-build dengan `go build ./...` secara langsung, versi akan ditampilkan sebagai `"dev"`.

---

## Memulai Pertama Kali

### Menjalankan Wizard setup

Setelah instalasi, pertama kali Anda harus menjalankan **wizard konfigurasi** dengan perintah berikut. Wizard akan memandu Anda satu per satu melalui semua pengaturan yang diperlukan.

```bash
wall-vault setup
```

Langkah-langkah yang dilakukan wizard:

```
1. Pilih bahasa (10 bahasa tersedia, termasuk Indonesia)
2. Pilih tema (light / dark / gold / cherry / ocean)
3. Mode operasi — pilih standalone (sendiri) atau distributed (beberapa komputer bersama)
4. Masukkan nama bot — nama yang akan ditampilkan di dashboard
5. Konfigurasi port — default: proxy 56244, brankas 56243 (tekan Enter jika tidak ingin mengubah)
6. Pilih layanan AI — pilih layanan yang ingin digunakan: Google / OpenRouter / Ollama
7. Konfigurasi filter keamanan alat
8. Konfigurasi token admin — kata sandi untuk mengunci fitur manajemen dashboard. Bisa dibuat otomatis
9. Konfigurasi kata sandi enkripsi kunci API — untuk menyimpan kunci dengan lebih aman (opsional)
10. Lokasi penyimpanan file konfigurasi
```

> ⚠️ **Harap ingat token admin Anda.** Token ini dibutuhkan nanti saat menambahkan kunci atau mengubah pengaturan di dashboard. Jika lupa, Anda harus mengedit file konfigurasi secara manual.

Setelah wizard selesai, file konfigurasi `wall-vault.yaml` akan dibuat secara otomatis.

### Menjalankan

```bash
wall-vault start
```

Dua server berikut akan dimulai secara bersamaan:

- **Proxy** (`http://localhost:56244`) — perantara antara OpenClaw dan layanan AI
- **Brankas Kunci** (`http://localhost:56243`) — manajemen kunci API dan dashboard web

Buka `http://localhost:56243` di browser untuk langsung melihat dashboard.

---

## Mendaftarkan Kunci API

Ada empat cara untuk mendaftarkan kunci API. **Bagi pemula, cara 1 (variabel lingkungan) sangat direkomendasikan.**

### Cara 1: Variabel Lingkungan (Direkomendasikan — Paling Mudah)

Variabel lingkungan (environment variable) adalah **nilai yang telah ditetapkan sebelumnya** yang dibaca saat program dimulai. Masukkan perintah berikut di terminal:

```bash
# Daftarkan kunci Google Gemini
export WV_KEY_GOOGLE=AIzaSy...

# Daftarkan kunci OpenRouter
export WV_KEY_OPENROUTER=sk-or-v1-...

# Jalankan setelah mendaftarkan
wall-vault start
```

Jika Anda memiliki beberapa kunci, hubungkan dengan koma (,). wall-vault akan menggunakannya secara bergantian secara otomatis (round robin):

```bash
export WV_KEY_GOOGLE=AIzaSy...,AIzaSy...,AIzaSy...
```

> 💡 **Tips**: Perintah `export` hanya berlaku untuk sesi terminal saat ini. Agar tetap berlaku setelah komputer dinyalakan ulang, tambahkan baris tersebut ke file `~/.bashrc` atau `~/.zshrc`.

### Cara 2: Antarmuka Dashboard (Klik dengan Mouse)

1. Buka `http://localhost:56243` di browser
2. Klik tombol `[+ Tambah]` di kartu **🔑 API Keys** bagian atas
3. Masukkan jenis layanan, nilai kunci, label (nama untuk catatan), dan batas harian, lalu simpan

### Cara 3: REST API (untuk Otomasi dan Skrip)

REST API adalah cara program saling bertukar data melalui HTTP. Berguna untuk pendaftaran otomatis melalui skrip.

```bash
curl -X POST http://localhost:56243/admin/keys \
  -H "Authorization: Bearer TOKEN_ADMIN" \
  -H "Content-Type: application/json" \
  -d '{
    "service": "google",
    "key": "AIzaSy...",
    "label": "Kunci Utama",
    "daily_limit": 1000
  }'
```

### Cara 4: Flag proxy (untuk Pengujian Cepat)

Digunakan untuk memasukkan kunci sementara tanpa pendaftaran resmi. Kunci akan hilang saat program ditutup.

```bash
wall-vault proxy --key-google=AIzaSy... --key-openrouter=sk-or-...
```

---

## Cara Menggunakan Proxy

### Menggunakan dengan OpenClaw (Tujuan Utama)

Berikut cara mengonfigurasi OpenClaw agar terhubung ke layanan AI melalui wall-vault.

Buka file `~/.openclaw/openclaw.json` dan tambahkan konten berikut:

```json5
// ~/.openclaw/openclaw.json
{
  models: {
    providers: {
      "wall-vault": {
        baseUrl: "http://localhost:56244/v1",
        apiKey: "your-agent-token",   // token agen vault
        api: "openai-completions",
        models: [
          { id: "wall-vault/gemini-2.5-flash" },
          { id: "wall-vault/gemini-2.5-pro" },
          { id: "wall-vault/hunter-alpha" },    // gratis, konteks 1M token
          { id: "wall-vault/claude-opus-4-6" }
        ]
      }
    }
  }
}
```

> 💡 **Cara lebih mudah**: Tekan tombol **🦞 Salin Konfigurasi OpenClaw** di kartu agen pada dashboard. Potongan konfigurasi yang sudah terisi token dan alamatnya akan disalin ke clipboard. Tinggal tempel saja.

**Ke mana `wall-vault/` di depan nama model akan diarahkan?**

wall-vault secara otomatis menentukan layanan AI mana yang akan menerima permintaan berdasarkan nama model:

| Format Model | Layanan yang Dihubungi |
|-------------|----------------------|
| `wall-vault/gemini-*` | Google Gemini langsung |
| `wall-vault/gpt-*`, `wall-vault/o3`, `wall-vault/o4*` | OpenAI langsung |
| `wall-vault/claude-*` | Anthropic melalui OpenRouter |
| `wall-vault/hunter-alpha`, `wall-vault/healer-alpha` | OpenRouter (gratis, konteks 1 juta token) |
| `wall-vault/kimi-*`, `wall-vault/glm-*`, `wall-vault/deepseek-*` | OpenRouter |
| `google/nama-model`, `openai/nama-model`, `anthropic/nama-model`, dll. | Layanan terkait langsung |
| `custom/google/nama-model`, `custom/openai/nama-model`, dll. | Hapus bagian `custom/` lalu diteruskan |
| `nama-model:cloud` | Hapus bagian `:cloud` lalu diarahkan ke OpenRouter |

> 💡 **Apa itu konteks (context)?** Konteks adalah jumlah percakapan yang bisa diingat AI dalam satu sesi. Konteks 1M (satu juta token) berarti AI dapat memproses percakapan sangat panjang atau dokumen besar sekaligus.

### Koneksi Langsung Format Gemini API (Kompatibilitas dengan Alat Lama)

Jika Anda memiliki alat yang sebelumnya menggunakan Google Gemini API secara langsung, cukup ubah alamatnya ke wall-vault:

```bash
export ANTHROPIC_BASE_URL=http://localhost:56244/google
```

Atau jika alatnya mendukung URL langsung:

```
http://localhost:56244/google/v1beta/models/gemini-2.5-flash:generateContent
```

### Menggunakan dengan OpenAI SDK (Python)

Kode Python yang memanfaatkan AI juga bisa dihubungkan ke wall-vault. Cukup ubah `base_url`-nya:

```python
from openai import OpenAI

client = OpenAI(
    base_url="http://localhost:56244/v1",
    api_key="not-needed"  # Kunci API dikelola oleh wall-vault
)

response = client.chat.completions.create(
    model="google/gemini-2.5-flash",   # format: provider/model
    messages=[{"role": "user", "content": "Halo"}]
)
```

### Mengganti Model Saat Sedang Berjalan

Untuk mengganti model AI yang digunakan saat wall-vault sudah berjalan:

```bash
# Ubah model dengan mengirim permintaan langsung ke proxy
curl -X PUT http://localhost:56244/api/config/model \
  -H "Content-Type: application/json" \
  -d '{"service": "openrouter", "model": "anthropic/claude-3.5-sonnet"}'

# Dalam mode terdistribusi (multi-bot), ubah dari server brankas → langsung tersinkronisasi via SSE
curl -X PUT http://localhost:56243/admin/clients/ID-BOT-SAYA \
  -H "Authorization: Bearer TOKEN_ADMIN" \
  -H "Content-Type: application/json" \
  -d '{"default_service": "google", "default_model": "gemini-2.5-pro"}'
```

### Melihat Daftar Model yang Tersedia

```bash
# Lihat semua daftar
curl http://localhost:56244/api/models | python3 -m json.tool

# Hanya model Google
curl "http://localhost:56244/api/models?service=google"

# Cari berdasarkan nama (contoh: model yang mengandung "claude")
curl "http://localhost:56244/api/models?q=claude"
```

**Ringkasan model utama per layanan:**

| Layanan | Model Utama |
|---------|------------|
| Google | gemini-2.5-pro, gemini-2.5-flash, gemini-2.5-flash-8b, gemini-2.0-flash |
| OpenAI | gpt-4o, gpt-4o-mini, o3, o1, o1-mini |
| OpenRouter | Lebih dari 346 model (Hunter Alpha 1M konteks gratis, DeepSeek R1/V3, Qwen 2.5, dll.) |
| Ollama | Deteksi otomatis server lokal yang terinstal di komputer Anda |

---

## Dashboard Brankas Kunci

Buka `http://localhost:56243` di browser untuk melihat dashboard.

**Tata letak layar:**
- **Bar atas tetap (topbar)**: Logo, pemilih bahasa & tema, indikator status koneksi SSE
- **Grid kartu**: Kartu agen, layanan, dan kunci API tersusun dalam bentuk ubin

### Kartu Kunci API

Kartu untuk mengelola semua kunci API yang terdaftar dalam satu tampilan.

- Menampilkan daftar kunci yang dikelompokkan per layanan.
- `today_usage`: jumlah token yang berhasil diproses hari ini (jumlah kata yang dibaca dan ditulis AI)
- `today_attempts`: total jumlah permintaan hari ini (termasuk yang berhasil dan gagal)
- Gunakan tombol `[+ Tambah]` untuk mendaftarkan kunci baru, dan `✕` untuk menghapus kunci.

> 💡 **Apa itu token?** Token adalah satuan yang digunakan AI saat memproses teks. Kira-kira setara dengan satu kata bahasa Inggris, atau 1–2 karakter bahasa lainnya. Biaya API biasanya dihitung berdasarkan jumlah token ini.

### Kartu Agen

Kartu yang menampilkan status bot (agen) yang terhubung ke proxy wall-vault.

**Status koneksi ditampilkan dalam 4 tahap:**

| Tampilan | Status | Arti |
|---------|--------|------|
| 🟢 | Berjalan | Proxy beroperasi dengan normal |
| 🟡 | Lambat | Merespons tapi lambat |
| 🔴 | Offline | Proxy tidak merespons |
| ⚫ | Tidak Terhubung / Nonaktif | Proxy belum pernah terhubung ke brankas atau telah dinonaktifkan |

**Panduan tombol di bawah kartu agen:**

Saat mendaftarkan agen, tentukan **jenis agen**-nya, dan tombol pintasan yang sesuai akan muncul secara otomatis.

---

#### 🔘 Tombol Salin Konfigurasi — Membuat pengaturan koneksi secara otomatis

Klik tombol ini untuk menyalin potongan konfigurasi yang sudah terisi token agen, alamat proxy, dan informasi model ke clipboard. Cukup tempel isi tersebut ke lokasi yang ditunjukkan pada tabel berikut untuk menyelesaikan pengaturan koneksi.

| Tombol | Jenis Agen | Lokasi untuk Ditempel |
|--------|-----------|----------------------|
| 🦞 Salin Konfigurasi OpenClaw | `openclaw` | `~/.openclaw/openclaw.json` |
| 🦀 Salin Konfigurasi NanoClaw | `nanoclaw` | `~/.openclaw/openclaw.json` |
| 🟠 Salin Konfigurasi Claude Code | `claude-code` | `~/.claude/settings.json` |
| ⌨ Salin Konfigurasi Cursor | `cursor` | Cursor → Settings → AI |
| 💻 Salin Konfigurasi VSCode | `vscode` | `~/.continue/config.json` |

**Contoh — jika tipe Claude Code, ini yang akan disalin:**

```json
// ~/.claude/settings.json
{
  "apiProvider": "openai",
  "baseUrl": "http://192.168.1.20:56244/v1",
  "apiKey": "token-agen-ini"
}
```

**Contoh — jika tipe VSCode (Continue):**

```json
// ~/.continue/config.json
{
  "models": [{
    "title": "wall-vault proxy",
    "provider": "openai",
    "model": "gemini-2.5-flash",
    "apiBase": "http://192.168.1.20:56244/v1",
    "apiKey": "token-agen-ini"
  }]
}
```

**Contoh — jika tipe Cursor:**

```
Base URL : http://192.168.1.20:56244/v1
API Key  : token-agen-ini

// Atau menggunakan variabel lingkungan:
OPENAI_BASE_URL=http://192.168.1.20:56244/v1
OPENAI_API_KEY=token-agen-ini
```

> ⚠️ **Jika salin ke clipboard tidak berhasil**: Kebijakan keamanan browser terkadang memblokir aksi ini. Jika muncul kotak teks di popup, tekan Ctrl+A untuk memilih semua teks, lalu Ctrl+C untuk menyalin.

---

#### ⚡ Tombol Terapkan Otomatis — Satu kali tekan, pengaturan selesai

Jika jenis agen adalah `cline`, `claude-code`, `openclaw`, atau `nanoclaw`, tombol **⚡ Terapkan Pengaturan** akan muncul di kartu agen. Menekan tombol ini akan memperbarui file pengaturan lokal agen tersebut secara otomatis.

| Tombol | Jenis Agen | File yang Diperbarui |
|--------|-----------|---------------------|
| ⚡ Terapkan Pengaturan Cline | `cline` | `~/.cline/data/globalState.json` + `secrets.json` |
| ⚡ Terapkan Pengaturan Claude Code | `claude-code` | `~/.claude/settings.json` |
| ⚡ Terapkan Pengaturan OpenClaw | `openclaw` | `~/.openclaw/openclaw.json` |
| ⚡ Terapkan Pengaturan NanoClaw | `nanoclaw` | `~/.openclaw/openclaw.json` |

> ⚠️ Tombol ini mengirim permintaan ke **localhost:56244** (proxy lokal). Proxy harus berjalan di mesin ini agar tombol berfungsi.

---

#### 🔀 Pengurutan Kartu dengan Seret dan Lepas (v0.1.17)

Anda dapat **menyeret** kartu agen di dashboard untuk menyusun ulang sesuai urutan yang diinginkan.

1. Klik dan seret kartu agen dengan mouse
2. Lepaskan di atas kartu pada posisi yang diinginkan untuk menukar urutan
3. Urutan baru **langsung disimpan ke server** dan tetap bertahan meskipun halaman di-refresh

> 💡 Perangkat sentuh (ponsel/tablet) belum didukung. Silakan gunakan browser desktop.

---

#### 🔄 Sinkronisasi Model Dua Arah (v0.1.16)

Saat Anda mengubah model agen dari dashboard brankas, pengaturan lokal agen tersebut akan diperbarui secara otomatis.

**Untuk Cline:**
- Ubah model di brankas → event SSE → proxy memperbarui field model di `globalState.json`
- Field yang diperbarui: `actModeOpenAiModelId`, `planModeOpenAiModelId`, `openAiModelId`
- `openAiBaseUrl` dan kunci API tidak diubah
- **Diperlukan reload VS Code (`Ctrl+Alt+R` atau `Ctrl+Shift+P` → `Developer: Reload Window`)**
  - Karena Cline tidak membaca ulang file pengaturan selama berjalan

**Untuk Claude Code:**
- Ubah model di brankas → event SSE → proxy memperbarui field `model` di `settings.json`
- Secara otomatis mencari di kedua path WSL dan Windows (`~/.claude/`, `/mnt/c/Users/*/.claude/`)

**Arah sebaliknya (agen → brankas):**
- Saat agen (Cline, Claude Code, dll.) mengirim permintaan ke proxy, proxy menyertakan informasi layanan dan model klien tersebut dalam heartbeat
- Layanan/model yang sedang digunakan ditampilkan secara real-time di kartu agen pada dashboard brankas

> 💡 **Inti**: Proxy mengidentifikasi agen dari token Authorization di permintaan, lalu secara otomatis mengarahkan ke layanan/model yang sudah diatur di brankas. Meskipun Cline atau Claude Code mengirimkan nama model yang berbeda, proxy akan menimpa dengan pengaturan brankas.

---

### Menggunakan Cline di VS Code — Panduan Lengkap

#### Langkah 1: Instal Cline

Instal **Cline** (ID: `saoudrizwan.claude-dev`) dari marketplace ekstensi VS Code.

#### Langkah 2: Daftarkan Agen di Brankas

1. Buka dashboard brankas (`http://IP-brankas:56243`)
2. Di bagian **Agen**, klik **+ Tambah**
3. Masukkan informasi berikut:

| Field | Nilai | Keterangan |
|-------|-------|-----------|
| ID | `my_cline` | Pengenal unik (huruf Latin, tanpa spasi) |
| Nama | `My Cline` | Nama yang ditampilkan di dashboard |
| Jenis Agen | `cline` | ← harus pilih `cline` |
| Layanan | Pilih layanan yang ingin digunakan (contoh: `google`) | |
| Model | Masukkan model yang ingin digunakan (contoh: `gemini-2.5-flash`) | |

4. Tekan **Simpan** dan token akan dibuat secara otomatis

#### Langkah 3: Hubungkan ke Cline

**Cara A — Terapkan Otomatis (Direkomendasikan)**

1. Pastikan **proxy** wall-vault berjalan di mesin ini (`localhost:56244`)
2. Klik tombol **⚡ Terapkan Pengaturan Cline** di kartu agen pada dashboard
3. Jika muncul notifikasi "Pengaturan berhasil diterapkan!", berarti sukses
4. Reload VS Code (`Ctrl+Alt+R`)

**Cara B — Pengaturan Manual**

Buka pengaturan (⚙️) Cline dari sidebar, lalu isi:
- **API Provider**: `OpenAI Compatible`
- **Base URL**: `http://alamat-proxy:56244/v1`
  - Jika di mesin yang sama: `http://localhost:56244/v1`
  - Jika di mesin lain (misalnya Mac Mini): `http://192.168.1.20:56244/v1`
- **API Key**: Token yang diterbitkan dari brankas (salin dari kartu agen)
- **Model ID**: Model yang diatur di brankas (contoh: `gemini-2.5-flash`)

#### Langkah 4: Verifikasi

Kirim pesan apa saja di jendela chat Cline. Jika semuanya normal:
- **Titik hijau (● Berjalan)** akan muncul di kartu agen pada dashboard brankas
- Layanan/model yang sedang digunakan akan ditampilkan di kartu (contoh: `google / gemini-2.5-flash`)

#### Mengganti Model

Untuk mengganti model Cline, ubah dari **dashboard brankas**:

1. Ubah dropdown layanan/model di kartu agen
2. Klik **Terapkan**
3. Reload VS Code (`Ctrl+Alt+R`) — nama model di footer Cline akan diperbarui
4. Mulai dari permintaan berikutnya, model baru yang akan digunakan

> 💡 Sebenarnya, proxy mengidentifikasi permintaan Cline berdasarkan token dan mengarahkannya ke model sesuai pengaturan brankas. Meskipun VS Code tidak di-reload, **model yang benar-benar digunakan langsung berubah** — reload hanya untuk memperbarui tampilan nama model di antarmuka Cline.

#### Mendeteksi Pemutusan Koneksi

Saat VS Code ditutup, kartu agen di dashboard brankas akan berubah menjadi kuning (lambat) setelah sekitar **2–3 menit**, dan merah (offline) setelah **5 menit**.

#### Pemecahan Masalah

| Gejala | Penyebab | Solusi |
|--------|----------|-------|
| Error "koneksi gagal" di Cline | Proxy tidak berjalan atau alamat salah | Periksa dengan `curl http://localhost:56244/health` |
| Titik hijau tidak muncul di brankas | Kunci API (token) belum diatur | Klik tombol **⚡ Terapkan Pengaturan Cline** sekali lagi |
| Nama model di footer Cline tidak berubah | Cline meng-cache pengaturan | Reload VS Code (`Ctrl+Alt+R`) |
| Nama model yang salah ditampilkan | Bug lama (diperbaiki di v0.1.16) | Perbarui proxy ke v0.1.16 atau lebih baru |

---

#### 🟣 Tombol Salin Perintah Deploy — Digunakan saat menginstal di mesin baru

Digunakan saat pertama kali menginstal proxy wall-vault di komputer baru dan menghubungkannya ke brankas. Klik tombol ini untuk menyalin seluruh skrip instalasi. Tempel dan jalankan di terminal komputer baru, dan hal-hal berikut akan diurus sekaligus:

1. Instal binary wall-vault (dilewati jika sudah terinstal)
2. Daftarkan layanan pengguna systemd secara otomatis
3. Mulai layanan dan hubungkan ke brankas secara otomatis

> 💡 Skrip sudah berisi token agen ini dan alamat server brankas, sehingga Anda bisa langsung menjalankannya setelah ditempel tanpa perubahan apa pun.

---

### Kartu Layanan

Kartu untuk mengaktifkan atau menonaktifkan dan mengonfigurasi layanan AI yang ingin digunakan.

- Sakelar toggle aktifkan/nonaktifkan per layanan
- Masukkan alamat server AI lokal (Ollama, LM Studio, vLLM, dll. yang berjalan di komputer Anda) untuk menemukan model yang tersedia secara otomatis.
- **Indikator status koneksi layanan lokal**: Titik ● di sebelah nama layanan berwarna **hijau** berarti terhubung, **abu-abu** berarti tidak terhubung.
- **Sinkronisasi otomatis checkbox**: Jika layanan lokal (seperti Ollama) sedang berjalan saat halaman dibuka, statusnya akan otomatis menjadi tercentang.

> 💡 **Jika layanan lokal berjalan di komputer lain**: Masukkan IP komputer tersebut di kolom URL layanan. Contoh: `http://192.168.1.20:11434` (Ollama), `http://192.168.1.20:1234` (LM Studio)

### Memasukkan Token Admin

Saat Anda mencoba menggunakan fitur penting di dashboard seperti menambah atau menghapus kunci, popup input token admin akan muncul. Masukkan token yang sudah Anda buat saat wizard setup. Setelah dimasukkan sekali, token akan tetap berlaku sampai browser ditutup.

> ⚠️ **Jika autentikasi gagal lebih dari 10 kali dalam 15 menit, IP tersebut akan diblokir sementara.** Jika Anda lupa token, periksa entri `admin_token` di file `wall-vault.yaml`.

---

## Mode Terdistribusi (Multi-Bot)

Konfigurasi ini digunakan saat menjalankan OpenClaw secara bersamaan di beberapa komputer, dengan **satu brankas kunci bersama**. Manajemen kunci hanya perlu dilakukan di satu tempat sehingga lebih praktis.

### Contoh Konfigurasi

```
[Server Brankas Kunci]
  wall-vault vault    (brankas :56243, dashboard)

[WSL Alpha]            [Raspberry Pi Gamma]    [Mac Mini Lokal]
  wall-vault proxy      wall-vault proxy          wall-vault proxy
  openclaw TUI          openclaw TUI              openclaw TUI
  ↕ Sinkronisasi SSE    ↕ Sinkronisasi SSE        ↕ Sinkronisasi SSE
```

Semua bot mengacu pada server brankas di tengah, sehingga perubahan model atau penambahan kunci di brankas langsung tercermin ke semua bot.

### Langkah 1: Mulai Server Brankas Kunci

Jalankan di komputer yang akan dijadikan server brankas:

```bash
wall-vault vault
```

### Langkah 2: Daftarkan Setiap Bot (Klien)

Daftarkan informasi setiap bot yang akan terhubung ke server brankas terlebih dahulu:

```bash
curl -X POST http://localhost:56243/admin/clients \
  -H "Authorization: Bearer TOKEN_ADMIN" \
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

Jalankan proxy di setiap komputer bot dengan menentukan alamat server brankas dan tokennya:

```bash
WV_VAULT_URL=http://192.168.x.x:56243 \
WV_VAULT_TOKEN=bota-secret \
WV_VAULT_CLIENT_ID=botA \
wall-vault proxy
```

> 💡 Ganti bagian **`192.168.x.x`** dengan alamat IP internal sebenarnya dari komputer server brankas. Anda bisa menemukannya di pengaturan router atau dengan perintah `ip addr`.

---

## Pengaturan Mulai Otomatis

Jika Anda merasa kerepotan harus menyalakan wall-vault secara manual setiap kali komputer dinyalakan ulang, daftarkan sebagai layanan sistem. Setelah didaftarkan, wall-vault akan mulai otomatis saat booting.

### Linux — systemd (Sebagian Besar Linux)

systemd adalah sistem di Linux yang secara otomatis menjalankan dan mengelola program:

```bash
wall-vault doctor deploy
systemctl --user daemon-reload
systemctl --user enable --now wall-vault
```

Melihat log:

```bash
journalctl --user -u wall-vault -f
```

### macOS — launchd

Sistem yang menangani jalannya program otomatis di macOS:

```bash
wall-vault doctor deploy launchd
launchctl load ~/Library/LaunchAgents/com.wall-vault.plist
```

### Windows — NSSM

1. Unduh NSSM dari [nssm.cc](https://nssm.cc/download) dan tambahkan ke PATH.
2. Di PowerShell dengan hak Administrator:

```powershell
wall-vault doctor deploy windows
```

---

## Doctor — Alat Diagnostik

Perintah `doctor` adalah **alat yang mendiagnosis sendiri dan memperbaiki** apakah wall-vault dikonfigurasi dengan benar.

```bash
wall-vault doctor check   # Diagnosis status saat ini (hanya baca, tidak mengubah apa pun)
wall-vault doctor fix     # Perbaiki masalah secara otomatis
wall-vault doctor all     # Diagnosis + perbaikan otomatis sekaligus
```

> 💡 Jika terasa ada yang tidak beres, coba jalankan `wall-vault doctor all` terlebih dahulu. Banyak masalah yang bisa ditangani secara otomatis.

---

## Referensi Variabel Lingkungan

Variabel lingkungan (environment variable) adalah cara meneruskan nilai konfigurasi ke program. Masukkan dalam format `export NAMA_VARIABEL=nilai` di terminal, atau simpan di file layanan mulai otomatis agar selalu berlaku.

| Variabel | Keterangan | Contoh Nilai |
|----------|-----------|-------------|
| `WV_LANG` | Bahasa dashboard | `ko`, `en`, `ja` |
| `WV_THEME` | Tema dashboard | `light`, `dark`, `gold` |
| `WV_KEY_GOOGLE` | Kunci API Google (beberapa kunci dipisahkan koma) | `AIza...,AIza...` |
| `WV_KEY_OPENROUTER` | Kunci API OpenRouter | `sk-or-v1-...` |
| `WV_VAULT_URL` | Alamat server brankas dalam mode terdistribusi | `http://192.168.x.x:56243` |
| `WV_VAULT_TOKEN` | Token autentikasi klien (bot) | `my-secret-token` |
| `WV_ADMIN_TOKEN` | Token admin | `admin-token-here` |
| `WV_MASTER_PASS` | Kata sandi enkripsi kunci API | `my-password` |
| `WV_AVATAR` | Path file gambar avatar (relatif terhadap `~/.openclaw/`) | `workspace/avatars/avatar.png` |
| `OLLAMA_URL` | Alamat server lokal Ollama | `http://192.168.x.x:11434` |

---

## Pemecahan Masalah

### Proxy Tidak Mau Mulai

Ini sering terjadi karena port sudah digunakan oleh program lain.

```bash
ss -tlnp | grep 56244   # Cek siapa yang menggunakan port 56244
wall-vault proxy --port 8080   # Mulai dengan nomor port berbeda
```

### Error Kunci API (429, 402, 401, 403, 582)

| Kode Error | Arti | Cara Mengatasi |
|-----------|------|---------------|
| **429** | Terlalu banyak permintaan (kuota terlampaui) | Tunggu sebentar atau tambahkan kunci lain |
| **402** | Perlu pembayaran atau kredit tidak cukup | Isi ulang kredit di layanan yang bersangkutan |
| **401 / 403** | Kunci salah atau tidak memiliki izin | Periksa ulang nilai kunci dan daftarkan kembali |
| **582** | Gateway kelebihan beban (cooldown 5 menit) | Akan otomatis pulih setelah 5 menit |

```bash
# Lihat daftar dan status kunci yang terdaftar
curl -H "Authorization: Bearer TOKEN_ADMIN" http://localhost:56243/admin/keys

# Reset penghitung penggunaan kunci
curl -X POST -H "Authorization: Bearer TOKEN_ADMIN" http://localhost:56243/admin/keys/reset
```

### Agen Ditampilkan sebagai "Tidak Terhubung"

"Tidak Terhubung" berarti proses proxy tidak mengirim sinyal (heartbeat) ke brankas. **Ini bukan berarti konfigurasi tidak tersimpan.** Proxy harus dijalankan dengan mengetahui alamat server brankas dan tokennya agar statusnya berubah menjadi terhubung.

```bash
# Mulai proxy dengan menentukan alamat server brankas, token, dan ID klien
WV_VAULT_URL=http://ALAMAT-SERVER-BRANKAS:56243 \
WV_VAULT_TOKEN=TOKEN-KLIEN \
WV_VAULT_CLIENT_ID=ID-KLIEN \
wall-vault proxy
```

Jika koneksi berhasil, dalam sekitar 20 detik status akan berubah menjadi 🟢 Berjalan di dashboard.

### Ollama Tidak Bisa Terhubung

Ollama adalah program yang menjalankan AI langsung di komputer Anda. Pertama, pastikan Ollama sedang berjalan.

```bash
curl http://localhost:11434/api/tags   # Jika daftar model muncul, berarti normal
export OLLAMA_URL=http://192.168.x.x:11434   # Jika berjalan di komputer lain
```

> ⚠️ Jika Ollama tidak merespons, mulai Ollama terlebih dahulu dengan perintah `ollama serve`.

> ⚠️ **Model besar merespons lambat**: Model besar seperti `qwen3.5:35b` atau `deepseek-r1` bisa membutuhkan beberapa menit untuk menghasilkan respons. Meskipun terlihat seperti tidak ada respons, mungkin sebenarnya sedang diproses dengan normal — mohon bersabar.

---

*Untuk informasi API yang lebih detail, lihat [API.md](API.md).*
