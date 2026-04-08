# Panduan Pengguna wall-vault
*(Last updated: 2026-04-08 — v0.1.25)*

---

## Daftar Isi

1. [Apa itu wall-vault?](#apa-itu-wall-vault)
2. [Instalasi](#instalasi)
3. [Memulai pertama kali (wizard setup)](#memulai-pertama-kali)
4. [Pendaftaran kunci API](#pendaftaran-kunci-api)
5. [Cara menggunakan proxy](#cara-menggunakan-proxy)
6. [Dashboard brankas kunci](#dashboard-brankas-kunci)
7. [Mode terdistribusi (multi-bot)](#mode-terdistribusi-multi-bot)
8. [Pengaturan mulai otomatis](#pengaturan-mulai-otomatis)
9. [Doctor (diagnostik)](#doctor-diagnostik)
10. [RTK Penghematan token](#rtk-penghematan-token)
11. [Referensi variabel lingkungan](#referensi-variabel-lingkungan)
12. [Pemecahan masalah](#pemecahan-masalah)

---

## Apa itu wall-vault?

**wall-vault = proxy AI + brankas kunci API untuk OpenClaw**

Untuk menggunakan layanan AI, Anda memerlukan **kunci API**. Kunci API seperti **kartu akses digital** yang membuktikan "orang ini memiliki izin untuk menggunakan layanan ini". Namun, kartu akses ini memiliki batas penggunaan harian dan bisa terekspos jika tidak dikelola dengan baik.

wall-vault menyimpan kartu akses ini dalam brankas yang aman dan berperan sebagai **proxy (perantara)** antara OpenClaw dan layanan AI. Sederhananya, OpenClaw hanya perlu terhubung ke wall-vault, dan wall-vault yang menangani semua hal rumit lainnya.

Masalah yang diselesaikan wall-vault:

- **Rotasi otomatis kunci API**: Ketika penggunaan suatu kunci mencapai batas atau diblokir sementara (cooldown), secara diam-diam beralih ke kunci berikutnya. OpenClaw terus berjalan tanpa gangguan.
- **Pergantian layanan otomatis (fallback)**: Jika Google tidak merespons, otomatis beralih ke OpenRouter; jika itu juga gagal, beralih ke Ollama/LM Studio/vLLM (AI lokal) yang terinstal di komputer Anda. Sesi tidak terputus. Ketika layanan asli pulih, otomatis kembali dari permintaan berikutnya (v0.1.18+, LM Studio/vLLM: v0.1.21+).
- **Sinkronisasi real-time (SSE)**: Ketika Anda mengubah model di dashboard brankas, perubahan tercermin di layar OpenClaw dalam 1-3 detik. SSE (Server-Sent Events) adalah teknologi di mana server mendorong pembaruan secara real-time ke klien.
- **Notifikasi real-time**: Event seperti habisnya kunci atau kegagalan layanan langsung ditampilkan di bagian bawah TUI (layar terminal) OpenClaw.

> 💡 **Claude Code, Cursor, VS Code** juga bisa dihubungkan, tetapi tujuan utama wall-vault adalah untuk digunakan bersama OpenClaw.

```
OpenClaw (layar TUI terminal)
        │
        ▼
  proxy wall-vault (:56244)   ← manajemen kunci, routing, fallback, event
        │
        ├─ Google Gemini API
        ├─ OpenRouter API (340+ model)
        ├─ Ollama / LM Studio / vLLM (komputer lokal, pilihan terakhir)
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

# macOS Apple Silicon (M1/M2/M3 Mac)
curl -L https://github.com/sookmook/wall-vault/releases/latest/download/wall-vault-darwin-arm64 \
  -o /usr/local/bin/wall-vault && chmod +x /usr/local/bin/wall-vault
```

- `curl -L ...` — Mengunduh file dari internet.
- `chmod +x` — Membuat file yang diunduh "dapat dieksekusi". Melewatkan langkah ini akan menyebabkan error "izin ditolak".

### Windows

Buka PowerShell (sebagai administrator) dan jalankan perintah di bawah ini.

```powershell
# Unduh
Invoke-WebRequest -Uri `
  "https://github.com/sookmook/wall-vault/releases/latest/download/wall-vault-windows-amd64.exe" `
  -OutFile "$env:LOCALAPPDATA\Programs\wall-vault\wall-vault.exe"

# Tambahkan ke PATH (berlaku setelah restart PowerShell)
$env:PATH += ";$env:LOCALAPPDATA\Programs\wall-vault"
```

> 💡 **Apa itu PATH?** PATH adalah daftar folder di mana komputer mencari perintah. Dengan menambahkan ke PATH, Anda dapat menjalankan `wall-vault` dari folder mana pun.

### Build dari kode sumber (untuk developer)

Hanya berlaku jika Anda memiliki lingkungan pengembangan bahasa Go yang terinstal.

```bash
git clone https://github.com/sookmook/wall-vault
cd wall-vault
make build       # bin/wall-vault (versi: v0.1.25.YYYYMMDD.HHmmss)
make install     # ~/.local/bin/wall-vault
```

> 💡 **Versi dengan timestamp build**: Ketika di-build dengan `make build`, versi otomatis dibuat dalam format seperti `v0.1.25.20260408.022325` yang menyertakan tanggal dan waktu. Jika di-build langsung dengan `go build ./...`, versi akan ditampilkan sebagai `"dev"` saja.

---

## Memulai pertama kali

### Menjalankan wizard setup

Setelah instalasi, Anda wajib menjalankan **wizard setup** dengan perintah di bawah ini untuk pertama kali. Wizard akan memandu Anda dengan menanyakan item yang diperlukan satu per satu.

```bash
wall-vault setup
```

Langkah-langkah yang dilalui wizard:

```
1. Pemilihan bahasa (10 bahasa termasuk Korea)
2. Pemilihan tema (light / dark / gold / cherry / ocean)
3. Mode operasi — sendiri (standalone) atau bersama di beberapa komputer (distributed)
4. Nama bot — nama yang akan ditampilkan di dashboard
5. Pengaturan port — default: proxy 56244, brankas 56243 (tekan Enter jika tidak perlu diubah)
6. Pemilihan layanan AI — Google / OpenRouter / Ollama / LM Studio / vLLM
7. Pengaturan filter keamanan tool
8. Pengaturan token admin — kata sandi untuk mengunci fungsi manajemen dashboard. Bisa dibuat otomatis
9. Pengaturan kata sandi enkripsi kunci API — untuk menyimpan kunci lebih aman (opsional)
10. Lokasi penyimpanan file konfigurasi
```

> ⚠️ **Pastikan mengingat token admin.** Akan diperlukan nanti untuk menambah kunci atau mengubah pengaturan di dashboard. Jika lupa, Anda harus mengedit file konfigurasi secara manual.

Setelah wizard selesai, file konfigurasi `wall-vault.yaml` dibuat secara otomatis.

### Menjalankan

```bash
wall-vault start
```

Dua server dimulai secara bersamaan:

- **Proxy** (`http://localhost:56244`) — Perantara yang menghubungkan OpenClaw dengan layanan AI
- **Brankas kunci** (`http://localhost:56243`) — Manajemen kunci API dan dashboard web

Buka `http://localhost:56243` di browser untuk melihat dashboard langsung.

---

## Pendaftaran kunci API

Ada empat cara mendaftarkan kunci API. **Untuk pemula, kami merekomendasikan Cara 1 (variabel lingkungan)**.

### Cara 1: Variabel lingkungan (direkomendasikan — paling sederhana)

Variabel lingkungan adalah **nilai yang telah diatur sebelumnya** yang dibaca program saat memulai. Cukup ketik di terminal seperti di bawah ini.

```bash
# Daftarkan kunci Google Gemini
export WV_KEY_GOOGLE=AIzaSy...

# Daftarkan kunci OpenRouter
export WV_KEY_OPENROUTER=sk-or-v1-...

# Jalankan setelah pendaftaran
wall-vault start
```

Jika Anda memiliki beberapa kunci, hubungkan dengan koma(,). wall-vault akan menggunakannya secara bergiliran otomatis (round robin):

```bash
export WV_KEY_GOOGLE=AIzaSy...,AIzaSy...,AIzaSy...
```

> 💡 **Tips**: Perintah `export` hanya berlaku untuk sesi terminal saat ini. Agar tetap setelah restart komputer, tambahkan baris di atas ke file `~/.bashrc` atau `~/.zshrc`.

### Cara 2: UI Dashboard (klik dengan mouse)

1. Akses `http://localhost:56243` di browser
2. Klik tombol `[+ Tambah]` di kartu **🔑 Kunci API** di bagian atas
3. Masukkan jenis layanan, nilai kunci, label (nama referensi), dan batas harian, lalu simpan

### Cara 3: REST API (untuk otomasi/skrip)

REST API adalah cara program bertukar data melalui HTTP. Berguna untuk pendaftaran otomatis via skrip.

```bash
curl -X POST http://localhost:56243/admin/keys \
  -H "Authorization: Bearer TOKEN_ADMIN" \
  -H "Content-Type: application/json" \
  -d '{
    "service": "google",
    "key": "AIzaSy...",
    "label": "Kunci utama",
    "daily_limit": 1000
  }'
```

### Cara 4: Flag proxy (untuk tes cepat)

Digunakan untuk tes sementara dengan memasukkan kunci tanpa pendaftaran resmi. Hilang saat program dihentikan.

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
        apiKey: "your-agent-token",   // token agen vault
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

> 💡 **Cara lebih mudah**: Klik tombol **🦞 Salin config OpenClaw** di kartu agen di dashboard. Snippet dengan token dan alamat sudah terisi akan disalin ke clipboard. Tinggal tempel saja.

**Ke mana `wall-vault/` di depan nama model mengarah?**

wall-vault secara otomatis menentukan layanan AI mana yang akan menerima permintaan berdasarkan nama model:

| Format model | Layanan yang terhubung |
|----------|--------------|
| `wall-vault/gemini-*` | Koneksi langsung ke Google Gemini |
| `wall-vault/gpt-*`, `wall-vault/o3`, `wall-vault/o4*` | Koneksi langsung ke OpenAI |
| `wall-vault/claude-*` | Koneksi ke Anthropic via OpenRouter |
| `wall-vault/hunter-alpha`, `wall-vault/healer-alpha` | OpenRouter (1 juta token context gratis) |
| `wall-vault/kimi-*`, `wall-vault/glm-*`, `wall-vault/deepseek-*` | Koneksi ke OpenRouter |
| `google/nama_model`, `openai/nama_model`, `anthropic/nama_model` dll. | Koneksi langsung ke layanan terkait |
| `custom/google/nama_model`, `custom/openai/nama_model` dll. | Menghapus bagian `custom/` dan me-route ulang |
| `nama_model:cloud` | Menghapus bagian `:cloud` dan menghubungkan via OpenRouter |

> 💡 **Apa itu context?** Jumlah percakapan yang bisa diingat AI sekaligus. 1M (satu juta token) berarti percakapan yang sangat panjang atau dokumen besar bisa diproses sekaligus.

### Koneksi langsung format Gemini API (kompatibilitas dengan tool yang ada)

Jika Anda memiliki tool yang menggunakan Google Gemini API secara langsung, cukup ubah alamat ke wall-vault:

```bash
export ANTHROPIC_BASE_URL=http://localhost:56244/google
```

Atau jika tool memungkinkan penentuan URL secara langsung:

```
http://localhost:56244/google/v1beta/models/gemini-2.5-flash:generateContent
```

### Penggunaan dengan OpenAI SDK (Python)

Anda juga bisa menghubungkan wall-vault di kode Python yang menggunakan AI. Cukup ubah `base_url`:

```python
from openai import OpenAI

client = OpenAI(
    base_url="http://localhost:56244/v1",
    api_key="not-needed"  # Kunci API dikelola oleh wall-vault
)

response = client.chat.completions.create(
    model="google/gemini-2.5-flash",   # format provider/model
    messages=[{"role": "user", "content": "Halo"}]
)
```

### Mengubah model saat berjalan

Untuk mengubah model AI saat wall-vault sedang berjalan:

```bash
# Ubah model dengan permintaan langsung ke proxy
curl -X PUT http://localhost:56244/api/config/model \
  -H "Content-Type: application/json" \
  -d '{"service": "openrouter", "model": "anthropic/claude-3.5-sonnet"}'

# Dalam mode terdistribusi (multi-bot), ubah di server brankas → langsung tercermin via SSE
curl -X PUT http://localhost:56243/admin/clients/id-bot-saya \
  -H "Authorization: Bearer TOKEN_ADMIN" \
  -H "Content-Type: application/json" \
  -d '{"default_service": "google", "default_model": "gemini-2.5-pro"}'
```

### Memeriksa daftar model yang tersedia

```bash
# Lihat daftar lengkap
curl http://localhost:56244/api/models | python3 -m json.tool

# Lihat model Google saja
curl "http://localhost:56244/api/models?service=google"

# Cari berdasarkan nama (contoh: model yang mengandung "claude")
curl "http://localhost:56244/api/models?q=claude"
```

**Ringkasan model utama per layanan:**

| Layanan | Model utama |
|--------|----------|
| Google | gemini-2.5-pro, gemini-2.5-flash, gemini-2.5-flash-8b, gemini-2.0-flash |
| OpenAI | gpt-4o, gpt-4o-mini, o3, o1, o1-mini |
| OpenRouter | 346+ (Hunter Alpha dengan context 1M gratis, DeepSeek R1/V3, Qwen 2.5 dll.) |
| Ollama | Deteksi otomatis server lokal yang terinstal di komputer |
| LM Studio | Server lokal komputer (port 1234) |
| vLLM | Server lokal komputer (port 8000) |

---

## Dashboard brankas kunci

Akses `http://localhost:56243` di browser untuk melihat dashboard.

**Layout layar:**
- **Bar atas tetap (topbar)**: Logo, pemilih bahasa/tema, indikator status koneksi SSE
- **Grid kartu**: Kartu agen, layanan, dan kunci API disusun dalam format ubin

### Kartu kunci API

Kartu untuk mengelola kunci API yang terdaftar secara visual.

- Menampilkan daftar kunci yang dikelompokkan per layanan.
- `today_usage`: Jumlah token (karakter yang dibaca dan ditulis AI) yang berhasil diproses hari ini
- `today_attempts`: Jumlah total panggilan hari ini (termasuk sukses + gagal)
- Daftarkan kunci baru dengan tombol `[+ Tambah]` dan hapus dengan `✕`.

> 💡 **Apa itu token?** Token adalah unit yang digunakan AI untuk memproses teks. Kira-kira setara dengan satu kata dalam bahasa Inggris, atau 1-2 karakter Korea. Tarif API biasanya dihitung berdasarkan jumlah token ini.

### Kartu agen

Kartu yang menampilkan status bot (agen) yang terhubung ke proxy wall-vault.

**Status koneksi ditampilkan dalam 4 level:**

| Indikator | Status | Arti |
|------|------|------|
| 🟢 | Berjalan | Proxy berfungsi normal |
| 🟡 | Tertunda | Merespons tapi lambat |
| 🔴 | Offline | Proxy tidak merespons |
| ⚫ | Tidak terhubung/Nonaktif | Proxy belum pernah terhubung ke brankas atau dinonaktifkan |

**Panduan tombol di bagian bawah kartu agen:**

Saat mendaftarkan agen, tentukan **jenis agen** dan tombol kenyamanan yang sesuai akan muncul secara otomatis.

---

#### 🔘 Tombol salin konfigurasi — Membuat pengaturan koneksi secara otomatis

Saat tombol diklik, snippet konfigurasi dengan token agen, alamat proxy, dan informasi model sudah terisi akan disalin ke clipboard. Cukup tempel konten yang disalin di lokasi yang ditunjukkan pada tabel di bawah untuk menyelesaikan pengaturan koneksi.

| Tombol | Jenis agen | Tempat tempel |
|------|-------------|-------------|
| 🦞 Salin config OpenClaw | `openclaw` | `~/.openclaw/openclaw.json` |
| 🦀 Salin config NanoClaw | `nanoclaw` | `~/.openclaw/openclaw.json` |
| 🟠 Salin config Claude Code | `claude-code` | `~/.claude/settings.json` |
| ⌨ Salin config Cursor | `cursor` | Cursor → Settings → AI |
| 💻 Salin config VSCode | `vscode` | `~/.continue/config.json` |

**Contoh — Untuk tipe Claude Code, konten yang disalin:**

```json
// ~/.claude/settings.json
{
  "apiProvider": "openai",
  "baseUrl": "http://192.168.0.6:56244/v1",
  "apiKey": "token-agen-ini"
}
```

**Contoh — Untuk tipe VSCode (Continue):**

```yaml
# ~/.continue/config.yaml  ← tempel di config.yaml, bukan config.json
name: My Config
version: 0.0.1
schema: v1

models:
  - name: wall-vault proxy
    provider: openai
    model: gemini-2.5-flash
    apiBase: http://192.168.0.6:56244/v1
    apiKey: token-agen-ini
    roles:
      - chat
      - edit
      - apply
```

> ⚠️ **Versi terbaru Continue menggunakan `config.yaml`.** Jika `config.yaml` ada, `config.json` akan diabaikan sepenuhnya. Pastikan untuk menempel di `config.yaml`.

**Contoh — Untuk tipe Cursor:**

```
Base URL : http://192.168.0.6:56244/v1
API Key  : token-agen-ini

// Atau variabel lingkungan:
OPENAI_BASE_URL=http://192.168.0.6:56244/v1
OPENAI_API_KEY=token-agen-ini
```

> ⚠️ **Ketika salin ke clipboard tidak berfungsi**: Kebijakan keamanan browser mungkin memblokir penyalinan. Jika muncul popup textbox, pilih semua dengan Ctrl+A dan salin dengan Ctrl+C.

---

#### ⚡ Tombol terapkan otomatis — Satu klik, pengaturan selesai

Untuk agen berjenis `cline`, `claude-code`, `openclaw`, atau `nanoclaw`, tombol **⚡ Terapkan Pengaturan** ditampilkan di kartu agen. Mengklik tombol ini akan otomatis memperbarui file konfigurasi lokal agen.

| Tombol | Jenis agen | File target |
|------|-------------|-------------|
| ⚡ Terapkan config Cline | `cline` | `~/.cline/data/globalState.json` + `secrets.json` |
| ⚡ Terapkan config Claude Code | `claude-code` | `~/.claude/settings.json` |
| ⚡ Terapkan config OpenClaw | `openclaw` | `~/.openclaw/openclaw.json` |
| ⚡ Terapkan config NanoClaw | `nanoclaw` | `~/.openclaw/openclaw.json` |

> ⚠️ Tombol ini mengirim permintaan ke **localhost:56244** (proxy lokal). Proxy harus berjalan di mesin tersebut agar berfungsi.

---

#### 🔀 Drag and drop kartu untuk mengurutkan ulang (v0.1.17, ditingkatkan v0.1.25)

Anda bisa **men-drag** kartu agen di dashboard untuk menata ulang sesuai urutan yang diinginkan.

1. Klik dan tahan area **lampu lalu lintas (●)** di sudut kiri atas kartu dan drag
2. Lepaskan di atas kartu pada posisi yang diinginkan dan urutan akan berubah

> 💡 Badan kartu (field input, tombol, dll.) tidak bisa di-drag. Hanya bisa ditangkap dari area lampu lalu lintas.

#### 🟠 Deteksi proses agen (v0.1.25)

Ketika proxy berfungsi normal tetapi proses agen lokal (NanoClaw, OpenClaw) berhenti, lampu lalu lintas kartu berubah menjadi **oranye (berkedip)** dan pesan "Proses agen berhenti" ditampilkan.

- 🟢 Hijau: Proxy + agen normal
- 🟠 Oranye (berkedip): Proxy normal, agen berhenti
- 🔴 Merah: Proxy offline
3. Urutan yang diubah **langsung disimpan ke server** dan tetap setelah refresh halaman

> 💡 Perangkat sentuh (ponsel/tablet) belum didukung. Gunakan browser desktop.

---

#### 🔄 Sinkronisasi model dua arah (v0.1.16)

Ketika Anda mengubah model agen di dashboard brankas, pengaturan lokal agen otomatis diperbarui.

**Untuk Cline:**
- Mengubah model di brankas → event SSE → proxy memperbarui field model di `globalState.json`
- Field yang diperbarui: `actModeOpenAiModelId`, `planModeOpenAiModelId`, `openAiModelId`
- `openAiBaseUrl` dan kunci API tidak diubah
- **Perlu reload VS Code (`Ctrl+Alt+R` atau `Ctrl+Shift+P` → `Developer: Reload Window`)**
  - Cline tidak membaca ulang file konfigurasi saat berjalan

**Untuk Claude Code:**
- Mengubah model di brankas → event SSE → proxy memperbarui field `model` di `settings.json`
- Pencarian otomatis di jalur WSL dan Windows (`~/.claude/`, `/mnt/c/Users/*/.claude/`)

**Arah sebaliknya (agen → brankas):**
- Ketika agen (Cline, Claude Code, dll.) mengirim permintaan ke proxy, proxy menyertakan informasi layanan/model klien tersebut dalam heartbeat
- Layanan/model yang sedang digunakan ditampilkan secara real-time di kartu agen di dashboard brankas

> 💡 **Poin utama**: Proxy mengidentifikasi agen melalui token Authorization dari permintaan dan melakukan routing otomatis ke layanan/model yang dikonfigurasi di brankas. Meskipun Cline atau Claude Code mengirim nama model yang berbeda, proxy menimpa dengan konfigurasi brankas.

---

### Menggunakan Cline di VS Code — Panduan detail

#### Langkah 1: Instal Cline

Instal **Cline** (ID: `saoudrizwan.claude-dev`) dari marketplace ekstensi VS Code.

#### Langkah 2: Daftarkan agen di brankas

1. Buka dashboard brankas (`http://IP-brankas:56243`)
2. Di bagian **Agen**, klik **+ Tambah**
3. Isi sebagai berikut:

| Field | Nilai | Keterangan |
|------|----|------|
| ID | `my_cline` | Pengidentifikasi unik (alfanumerik, tanpa spasi) |
| Nama | `My Cline` | Nama yang ditampilkan di dashboard |
| Jenis agen | `cline` | ← harus memilih `cline` |
| Layanan | Pilih layanan yang diinginkan (contoh: `google`) | |
| Model | Masukkan model yang diinginkan (contoh: `gemini-2.5-flash`) | |

4. Klik **Simpan** dan token akan dibuat secara otomatis

#### Langkah 3: Hubungkan ke Cline

**Cara A — Terapkan otomatis (direkomendasikan)**

1. Pastikan **proxy** wall-vault berjalan di mesin tersebut (`localhost:56244`)
2. Klik tombol **⚡ Terapkan config Cline** di kartu agen di dashboard
3. Jika muncul notifikasi "Pengaturan berhasil diterapkan!", berarti sukses
4. Reload VS Code (`Ctrl+Alt+R`)

**Cara B — Pengaturan manual**

Buka pengaturan (⚙️) di sidebar Cline:
- **API Provider**: `OpenAI Compatible`
- **Base URL**: `http://alamat-proxy:56244/v1`
  - Jika di mesin yang sama: `http://localhost:56244/v1`
  - Jika di mesin lain (misal server Mini): `http://192.168.0.6:56244/v1`
- **API Key**: Token yang diterbitkan dari brankas (salin dari kartu agen)
- **Model ID**: Model yang dikonfigurasi di brankas (contoh: `gemini-2.5-flash`)

#### Langkah 4: Verifikasi

Kirim pesan apa pun di chat Cline. Jika normal:
- **Titik hijau (● Berjalan)** akan muncul di kartu agen terkait di dashboard brankas
- Layanan/model saat ini akan ditampilkan di kartu (contoh: `google / gemini-2.5-flash`)

#### Mengubah model

Ketika ingin mengubah model Cline, ubah di **dashboard brankas**:

1. Ubah dropdown layanan/model di kartu agen
2. Klik **Terapkan**
3. Reload VS Code (`Ctrl+Alt+R`) — nama model di footer Cline akan diperbarui
4. Model baru akan digunakan dari permintaan berikutnya

> 💡 Sebenarnya, proxy mengidentifikasi permintaan Cline berdasarkan token dan me-route ke model yang dikonfigurasi di brankas. Bahkan tanpa reload VS Code, **model yang sebenarnya digunakan langsung berubah** — reload hanya untuk memperbarui tampilan model di UI Cline.

#### Deteksi pemutusan koneksi

Ketika VS Code ditutup, kartu agen di dashboard brankas berubah menjadi kuning (tertunda) setelah sekitar **90 detik** dan merah (offline) setelah **3 menit**. (Sejak v0.1.18, deteksi offline menjadi lebih cepat dengan pemeriksaan status setiap 15 detik.)

#### Pemecahan masalah

| Gejala | Penyebab | Solusi |
|------|------|------|
| Error "Koneksi gagal" di Cline | Proxy tidak berjalan atau alamat salah | Periksa proxy dengan `curl http://localhost:56244/health` |
| Titik hijau tidak muncul di brankas | Kunci API (token) belum diatur | Klik tombol **⚡ Terapkan config Cline** lagi |
| Nama model di footer Cline tidak berubah | Cline meng-cache pengaturan | Reload VS Code (`Ctrl+Alt+R`) |
| Nama model yang salah ditampilkan | Bug lama (diperbaiki di v0.1.16) | Perbarui proxy ke v0.1.16 atau lebih baru |

---

#### 🟣 Tombol salin perintah deploy — Untuk instalasi di mesin baru

Digunakan saat menginstal proxy wall-vault pertama kali di komputer baru dan menghubungkannya ke brankas. Mengklik tombol akan menyalin seluruh skrip instalasi. Tempel dan jalankan di terminal komputer baru dan hal berikut akan diproses sekaligus:

1. Instalasi biner wall-vault (dilewati jika sudah terinstal)
2. Pendaftaran otomatis layanan systemd user
3. Mulai layanan dan koneksi otomatis ke brankas

> 💡 Skrip sudah berisi token agen ini dan alamat server brankas, jadi bisa langsung dijalankan setelah ditempel tanpa modifikasi.

---

### Kartu layanan

Kartu untuk mengaktifkan, menonaktifkan, atau mengkonfigurasi layanan AI yang akan digunakan.

- Toggle switch untuk mengaktifkan/menonaktifkan setiap layanan
- Masukkan alamat server AI lokal (Ollama, LM Studio, vLLM, dll. yang berjalan di komputer Anda) dan model yang tersedia akan terdeteksi secara otomatis.
- **Indikator status koneksi layanan lokal**: Titik ● di samping nama layanan berwarna **hijau** jika terhubung dan **abu-abu** jika tidak terhubung
- **Lampu lalu lintas otomatis layanan lokal** (v0.1.23+): Layanan lokal (Ollama, LM Studio, vLLM) otomatis diaktifkan/dinonaktifkan berdasarkan ketersediaan koneksi. Saat mengaktifkan layanan, dalam 15 detik titik ● menjadi hijau dan checkbox aktif; saat menonaktifkan, otomatis mati. Bekerja sama seperti toggle otomatis layanan cloud (Google, OpenRouter, dll.) berdasarkan keberadaan kunci API.

> 💡 **Jika layanan lokal berjalan di komputer lain**: Masukkan IP komputer tersebut di field URL layanan. Contoh: `http://192.168.0.6:11434` (Ollama), `http://192.168.0.6:1234` (LM Studio). Jika layanan hanya terikat ke `127.0.0.1` bukan `0.0.0.0`, akses melalui IP eksternal tidak akan berfungsi, jadi periksa alamat binding di pengaturan layanan.

### Input token admin

Saat mencoba menggunakan fungsi penting di dashboard seperti menambah/menghapus kunci, popup input token admin akan muncul. Masukkan token yang diatur di wizard setup. Setelah dimasukkan, tetap berlaku sampai browser ditutup.

> ⚠️ **Jika kegagalan autentikasi melebihi 10 kali dalam 15 menit, IP tersebut akan diblokir sementara.** Jika lupa token, periksa item `admin_token` di file `wall-vault.yaml`.

---

## Mode terdistribusi (multi-bot)

Konfigurasi di mana **satu brankas kunci dibagikan** saat menjalankan OpenClaw di beberapa komputer secara bersamaan. Praktis karena manajemen kunci dilakukan di satu tempat.

### Contoh konfigurasi

```
[Server brankas kunci]
  wall-vault vault    (brankas kunci :56243, dashboard)

[WSL Alpha]            [Raspberry Pi Gamma]    [Mac Mini Lokal]
  wall-vault proxy      wall-vault proxy        wall-vault proxy
  openclaw TUI          openclaw TUI            openclaw TUI
  ↕ sinkr. SSE          ↕ sinkr. SSE            ↕ sinkr. SSE
```

Semua bot mengarah ke server brankas pusat, jadi ketika Anda mengubah model atau menambah kunci di brankas, perubahan langsung tercermin ke semua bot.

### Langkah 1: Mulai server brankas kunci

Jalankan di komputer yang akan digunakan sebagai server brankas:

```bash
wall-vault vault
```

### Langkah 2: Daftarkan setiap bot (klien)

Daftarkan terlebih dahulu informasi setiap bot yang akan terhubung ke server brankas:

```bash
curl -X POST http://localhost:56243/admin/clients \
  -H "Authorization: Bearer TOKEN_ADMIN" \
  -H "Content-Type: application/json" \
  -d '{
    "id": "botA",
    "name": "BotA",
    "token": "bota-secret",
    "default_service": "google",
    "default_model": "gemini-2.5-flash"
  }'
```

### Langkah 3: Mulai proxy di setiap komputer bot

Jalankan proxy di setiap komputer tempat bot terinstal dengan menentukan alamat server brankas dan token:

```bash
WV_VAULT_URL=http://192.168.x.x:56243 \
WV_VAULT_TOKEN=bota-secret \
WV_VAULT_CLIENT_ID=botA \
wall-vault proxy
```

> 💡 **Ganti `192.168.x.x`** dengan alamat IP internal sebenarnya dari komputer server brankas. Bisa diperiksa di pengaturan router atau dengan perintah `ip addr`.

---

## Pengaturan mulai otomatis

Jika merasa repot harus memulai wall-vault secara manual setiap kali komputer direstart, daftarkan sebagai layanan sistem. Setelah terdaftar, akan otomatis dimulai saat boot.

### Linux — systemd (sebagian besar Linux)

systemd adalah sistem untuk memulai dan mengelola program secara otomatis di Linux:

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
2. Di PowerShell sebagai administrator:

```powershell
wall-vault doctor deploy windows
```

---

## Doctor (diagnostik)

Perintah `doctor` adalah alat yang **mendiagnosis dan memperbaiki secara otomatis** apakah wall-vault dikonfigurasi dengan benar.

```bash
wall-vault doctor check   # Diagnosis status saat ini (hanya baca, tidak mengubah apa pun)
wall-vault doctor fix     # Perbaiki masalah secara otomatis
wall-vault doctor all     # Diagnosis + perbaikan otomatis sekaligus
```

> 💡 Jika ada yang tampak salah, jalankan `wall-vault doctor all` terlebih dahulu. Banyak masalah diperbaiki secara otomatis.

---

## RTK Penghematan token

*(v0.1.24+)*

**RTK (alat penghematan token)** secara otomatis mengompresi output perintah shell yang dijalankan oleh agen AI coding (seperti Claude Code), mengurangi penggunaan token. Misalnya, output `git status` 15 baris diringkas menjadi 2 baris.

### Penggunaan dasar

```bash
# Bungkus perintah dengan wall-vault rtk untuk penyaringan output otomatis
wall-vault rtk git status          # Hanya menampilkan daftar file yang diubah
wall-vault rtk git diff HEAD~1     # Hanya baris yang diubah + context minimal
wall-vault rtk git log -10         # Satu baris per commit: hash + pesan
wall-vault rtk go test ./...       # Hanya menampilkan tes yang gagal
wall-vault rtk ls -la              # Perintah tidak didukung otomatis dipotong
```

### Perintah yang didukung dan penghematan

| Perintah | Metode penyaringan | Penghematan |
|------|----------|--------|
| `git status` | Hanya ringkasan file yang diubah | ~87% |
| `git diff` | Baris yang diubah + 3 baris context | ~60-94% |
| `git log` | Hash + baris pertama pesan | ~90% |
| `git push/pull/fetch` | Hapus progres, hanya ringkasan | ~80% |
| `go test` | Hanya tampilkan yang gagal, hitung yang lulus | ~88-99% |
| `go build/vet` | Hanya tampilkan error | ~90% |
| Semua perintah lain | 50 baris pertama + 50 terakhir, maks 32KB | Bervariasi |

### Pipeline penyaringan 3 tahap

1. **Filter struktural per perintah** — Memahami format output git, go, dll. dan mengekstrak hanya bagian yang bermakna
2. **Post-processing regex** — Menghapus kode warna ANSI, mengurangi baris kosong, mengagregasi baris duplikat
3. **Passthrough + pemotongan** — Perintah tidak didukung hanya menyimpan 50 baris pertama/terakhir

### Integrasi Claude Code

Bisa dikonfigurasi agar semua perintah shell otomatis melewati RTK menggunakan hook `PreToolUse` Claude Code.

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

> 💡 **Preservasi exit code**: RTK mengembalikan exit code asli perintah. Jika perintah gagal (exit code ≠ 0), AI mendeteksi kegagalan dengan akurat.

> 💡 **Output Inggris paksa**: RTK menjalankan perintah dengan `LC_ALL=C` untuk selalu menghasilkan output Inggris terlepas dari pengaturan bahasa sistem. Ini memastikan filter bekerja dengan benar.

---

## Referensi variabel lingkungan

Variabel lingkungan adalah cara meneruskan nilai konfigurasi ke program. Masukkan di terminal dalam format `export NAMA_VARIABEL=nilai` atau taruh di file layanan mulai otomatis untuk penerapan permanen.

| Variabel | Keterangan | Contoh nilai |
|------|------|---------|
| `WV_LANG` | Bahasa dashboard | `ko`, `en`, `ja` |
| `WV_THEME` | Tema dashboard | `light`, `dark`, `gold` |
| `WV_KEY_GOOGLE` | Kunci API Google (beberapa dipisah koma) | `AIza...,AIza...` |
| `WV_KEY_OPENROUTER` | Kunci API OpenRouter | `sk-or-v1-...` |
| `WV_VAULT_URL` | Alamat server brankas dalam mode terdistribusi | `http://192.168.x.x:56243` |
| `WV_VAULT_TOKEN` | Token autentikasi klien (bot) | `my-secret-token` |
| `WV_ADMIN_TOKEN` | Token admin | `admin-token-here` |
| `WV_MASTER_PASS` | Kata sandi enkripsi kunci API | `my-password` |
| `WV_AVATAR` | Path file avatar (relatif terhadap `~/.openclaw/`) | `workspace/avatars/avatar.png` |
| `OLLAMA_URL` | Alamat server lokal Ollama | `http://192.168.x.x:11434` |

---

## Pemecahan masalah

### Ketika proxy tidak bisa dimulai

Sering kali, port sudah digunakan oleh program lain.

```bash
ss -tlnp | grep 56244   # Periksa siapa yang menggunakan port 56244
wall-vault proxy --port 8080   # Mulai dengan nomor port lain
```

### Error kunci API (429, 402, 401, 403, 582)

| Kode error | Arti | Tindakan |
|----------|------|----------|
| **429** | Terlalu banyak permintaan (batas penggunaan terlampaui) | Tunggu sebentar atau tambah kunci lain |
| **402** | Pembayaran diperlukan atau kredit tidak cukup | Isi ulang kredit di layanan terkait |
| **401 / 403** | Kunci salah atau tidak punya izin | Periksa kembali nilai kunci dan daftarkan ulang |
| **582** | Overload gateway (cooldown 5 menit) | Otomatis terbuka setelah 5 menit |

```bash
# Periksa daftar kunci terdaftar dan status
curl -H "Authorization: Bearer TOKEN_ADMIN" http://localhost:56243/admin/keys

# Reset counter penggunaan kunci
curl -X POST -H "Authorization: Bearer TOKEN_ADMIN" http://localhost:56243/admin/keys/reset
```

### Ketika agen ditampilkan sebagai "Tidak terhubung"

"Tidak terhubung" berarti proses proxy tidak mengirim sinyal (heartbeat) ke brankas. **Ini bukan berarti pengaturan tidak tersimpan.** Proxy harus berjalan dengan alamat server brankas dan token agar status koneksi berubah.

```bash
# Mulai proxy dengan menentukan alamat server brankas, token, dan ID klien
WV_VAULT_URL=http://alamat-server-brankas:56243 \
WV_VAULT_TOKEN=token-klien \
WV_VAULT_CLIENT_ID=id-klien \
wall-vault proxy
```

Jika koneksi berhasil, status akan berubah menjadi 🟢 Berjalan di dashboard dalam sekitar 20 detik.

### Ketika koneksi Ollama tidak berfungsi

Ollama adalah program yang menjalankan AI langsung di komputer Anda. Pertama, periksa apakah Ollama sudah berjalan.

```bash
curl http://localhost:11434/api/tags   # Jika daftar model muncul, berarti normal
export OLLAMA_URL=http://192.168.x.x:11434   # Jika berjalan di komputer lain
```

> ⚠️ Jika Ollama tidak merespons, mulai dulu dengan perintah `ollama serve`.

> ⚠️ **Model besar lambat**: Model besar seperti `qwen3.5:35b` dan `deepseek-r1` bisa memakan waktu beberapa menit untuk menghasilkan respons. Meskipun tampak tidak ada respons, mungkin sedang diproses secara normal, jadi mohon tunggu.

---

## Perubahan terbaru (v0.1.16 ~ v0.1.25)

### v0.1.25 (2026-04-08)
- **Deteksi proses agen**: Proxy mendeteksi apakah agen lokal (NanoClaw/OpenClaw) masih hidup dan menampilkan lampu lalu lintas oranye di dashboard.
- **Perbaikan handle drag**: Saat mengurutkan ulang kartu, hanya bisa di-drag dari area lampu lalu lintas (●). Tidak lagi bisa ter-drag secara tidak sengaja dari field input atau tombol.

### v0.1.24 (2026-04-06)
- **Subperintah RTK penghematan token**: `wall-vault rtk <command>` otomatis menyaring output perintah shell, mengurangi penggunaan token agen AI sebesar 60-90%. Termasuk filter khusus untuk perintah utama seperti git dan go, dan otomatis memotong perintah tidak didukung. Terintegrasi secara transparan dengan hook `PreToolUse` Claude Code.

### v0.1.23 (2026-04-06)
- **Perbaikan perubahan model Ollama**: Diperbaiki masalah di mana mengubah model Ollama di dashboard brankas tidak tercermin di proxy sebenarnya. Sebelumnya hanya menggunakan variabel lingkungan (`OLLAMA_MODEL`), sekarang memprioritaskan konfigurasi brankas.
- **Lampu lalu lintas otomatis layanan lokal**: Ollama/LM Studio/vLLM otomatis diaktifkan saat bisa terhubung dan dinonaktifkan saat terputus. Bekerja sama seperti toggle otomatis layanan cloud berdasarkan kunci.

### v0.1.22 (2026-04-05)
- **Perbaikan field content kosong yang hilang**: Ketika model thinking (gemini-3.1-pro, o1, claude thinking, dll.) menggunakan seluruh batas max_tokens untuk reasoning dan tidak bisa menghasilkan respons nyata, proxy menghilangkan field `content`/`text` dari JSON respons via `omitempty`, menyebabkan klien SDK OpenAI/Anthropic crash dengan error `Cannot read properties of undefined (reading 'trim')`. Diperbaiki agar selalu menyertakan field sesuai spesifikasi API resmi.

### v0.1.21 (2026-04-05)
- **Dukungan model Gemma 4**: Model keluarga Gemma seperti `gemma-4-31b-it` dan `gemma-4-26b-a4b-it` bisa digunakan via Google Gemini API.
- **Dukungan resmi layanan LM Studio / vLLM**: Sebelumnya layanan ini terlewat dari routing proxy dan selalu diganti Ollama. Sekarang di-route dengan benar via API yang kompatibel dengan OpenAI.
- **Perbaikan tampilan layanan dashboard**: Meskipun terjadi fallback, dashboard selalu menampilkan layanan yang dikonfigurasi pengguna.
- **Indikator status layanan lokal**: Saat memuat dashboard, menampilkan status koneksi layanan lokal (Ollama, LM Studio, vLLM, dll.) melalui warna titik ●.
- **Variabel lingkungan filter tool**: Mode penerusan tool bisa dikonfigurasi via variabel lingkungan `WV_TOOL_FILTER=passthrough`.

### v0.1.20 (2026-03-28)
- **Penguatan keamanan menyeluruh**: 12 item keamanan ditingkatkan termasuk pencegahan XSS (41 titik), perbandingan token waktu konstan, pembatasan CORS, batas ukuran permintaan, pencegahan traversal jalur, autentikasi SSE, dan penguatan rate limiter.

### v0.1.19 (2026-03-27)
- **Deteksi online Claude Code**: Claude Code yang tidak melewati proxy juga ditampilkan sebagai online di dashboard.

### v0.1.18 (2026-03-26)
- **Perbaikan layanan fallback yang terjebak**: Setelah fallback ke Ollama karena error sementara, otomatis kembali saat layanan asli pulih.
- **Peningkatan deteksi offline**: Deteksi berhentinya proxy lebih cepat dengan pemeriksaan status setiap 15 detik.

### v0.1.17 (2026-03-25)
- **Pengurutan kartu drag and drop**: Kartu agen bisa diurutkan ulang dengan men-drag-nya.
- **Tombol terapkan pengaturan inline**: Tombol [⚡ Terapkan Pengaturan] ditampilkan di agen offline.
- **Penambahan jenis agen cokacdir**.

### v0.1.16 (2026-03-25)
- **Sinkronisasi model dua arah**: Mengubah model Cline/Claude Code di dashboard brankas otomatis tercermin.

---

*Untuk informasi API yang lebih detail, lihat [API.md](API.md).*
