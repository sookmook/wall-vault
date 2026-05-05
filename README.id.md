# wall-vault

> **API key vault + AI proxy dalam satu binary Go.**
> Menyimpan key secara lokal dengan AES-GCM, merotasinya antar provider, beralih ke fallback ketika salah satunya gagal, dan dilengkapi dasbor real-time.

[![License: GPL v3](https://img.shields.io/badge/License-GPLv3-blue.svg)](LICENSE)
[![Go](https://img.shields.io/badge/Go-1.25-00ADD8.svg)](go.mod)

[English](README.md) · [한국어](README.ko.md) · [中文](README.zh.md) · [日本語](README.ja.md) · [Español](README.es.md) · [Français](README.fr.md) · [Deutsch](README.de.md) · [Português](README.pt.md) · [العربية](README.ar.md) · [हिन्दी](README.hi.md) · **Bahasa Indonesia** · [ภาษาไทย](README.th.md) · [Kiswahili](README.sw.md) · [Hausa](README.ha.md) · [नेपाली](README.ne.md) · [Монгол](README.mn.md) · [isiZulu](README.zu.md)

---

## Apa ini

wall-vault duduk di antara AI agent (OpenClaw, Claude Code, Cursor, Continue, atau script Anda sendiri) dan provider AI cloud atau lokal yang menjadi lawan bicaranya. Dua hal dalam satu binary:

- **Vault** — menyimpan API key terenkripsi saat istirahat (AES-GCM dengan master password), merotasinya, melacak penggunaan dan cooldown per-key, menyiarkan perubahan melalui SSE, dan menyajikan dasbor web di `:56243`.
- **Proxy** — mengekspos endpoint yang kompatibel dengan Gemini, Anthropic, dan OpenAI di `:56244`, memilih sebuah key dari vault, mengirim ke upstream yang Anda konfigurasi, dan beralih ke provider berikutnya ketika salah satunya gagal.

Mendukung empat bentuk request (Gemini `:generateContent`, Anthropic `/v1/messages`, OpenAI `/v1/chat/completions`, dan Ollama-native `/api/chat`) serta lima kategori upstream:

| Provider | Catatan |
|----------|-------|
| Google Gemini | API native; rotasi key per project |
| Anthropic | Passthrough native `/v1/messages` |
| OpenAI | `/v1/chat/completions` native |
| OpenRouter | 340+ model, auto-fallback ke varian `:free` |
| Ollama / LM Studio / vLLM / llama.cpp / Jan / KoboldCpp / TabbyAPI / mlx-server / LiteLLM | Backend lokal yang kompatibel dengan OpenAI; drop-in via plugin yaml |

Menambahkan backend kompatibel-OpenAI baru hanya perlu satu file yaml di bawah `~/.wall-vault/services/` — tanpa perubahan kode.

## Mengapa Anda mungkin membutuhkannya

- Anda menjalankan tiga atau empat layanan AI dan menginginkan satu URL untuk diajak bicara oleh agent.
- Anda ingin agar key tier-gratis yang sedang cooldown dapat menyingkir untuk yang berikutnya tanpa merusak sesi.
- Anda ingin key yang sama menggerakkan beberapa bot / IDE / script di LAN yang sama tanpa menyalin kredensial.
- Anda ingin dasbor, bukan environment variable, untuk menyunting API key.
- Anda menginginkan opsi local-first (Ollama / LM Studio) saat batas cloud habis.

## Mulai cepat

### Install (Linux / macOS)

```bash
curl -fsSL https://raw.githubusercontent.com/sookmook/wall-vault/main/install.sh | sh
```

Atau unduh binary pre-built secara langsung:

```bash
# Linux amd64
curl -L https://github.com/sookmook/wall-vault/releases/latest/download/wall-vault-linux-amd64 \
  -o wall-vault && chmod +x wall-vault

# macOS Apple Silicon
curl -L https://github.com/sookmook/wall-vault/releases/latest/download/wall-vault-darwin-arm64 \
  -o wall-vault && chmod +x wall-vault

# Linux arm64 (Raspberry Pi, server ARM)
curl -L https://github.com/sookmook/wall-vault/releases/latest/download/wall-vault-linux-arm64 \
  -o wall-vault && chmod +x wall-vault
```

### Install (Windows)

```powershell
Invoke-WebRequest -Uri `
  "https://github.com/sookmook/wall-vault/releases/latest/download/wall-vault-windows-amd64.exe" `
  -OutFile wall-vault.exe
```

### Menjalankan pertama kali

```bash
wall-vault setup    # wizard interaktif — memilih port, services, admin token, master password
wall-vault start    # meluncurkan vault dan proxy sekaligus
```

Buka `http://localhost:56243` (atau `https://...` setelah TLS aktif — lihat di bawah) di browser. Dasbor meminta admin token yang dicetak oleh `setup`. Dari sana Anda menambahkan API key, mendaftarkan client, dan berganti model tanpa restart.

---

## TLS (direkomendasikan)

Secara default `wall-vault setup` menulis konfigurasi tanpa TLS, sehingga kedua listener menjawab HTTP polos. Contoh URL di README ini menggunakan `https://localhost:56244` karena sebagian besar agent (OpenClaw, Claude Code, Cursor) menginginkan satu endpoint berlapis TLS yang tidak akan rusak jika kelak Anda memindahkan proxy ke host lain. Untuk menyelaraskan dengan contoh-contoh tersebut, aktifkan TLS sekali dengan CA internal yang sudah disediakan:

```bash
# 1. Buat CA internal wall-vault (sekali saja, terletak di ~/.wall-vault/ca.{crt,key})
wall-vault cert init

# 2. Terbitkan host certificate untuk MESIN INI
#    SAN mencakup hostname, localhost, 127.0.0.1, dan setiap LAN IP yang terdeteksi
wall-vault cert issue $(hostname)

# 3. Percayai CA di keychain OS lokal
wall-vault cert install-trust

# 4. Alihkan listener ke TLS
export WV_PROXY_TLS_ENABLED=1
export WV_PROXY_TLS_CERT="$HOME/.wall-vault/$(hostname).crt"
export WV_PROXY_TLS_KEY="$HOME/.wall-vault/$(hostname).key"
export WV_VAULT_TLS_ENABLED=1
export WV_VAULT_TLS_CERT="$HOME/.wall-vault/$(hostname).crt"
export WV_VAULT_TLS_KEY="$HOME/.wall-vault/$(hostname).key"
wall-vault start
```

Untuk mesin lain di LAN Anda: salin `~/.wall-vault/ca.crt` ke sana lalu jalankan `wall-vault cert install-trust --ca <path>` di sana. Setelah CA dipercaya di mana-mana, setiap mesin di jaringan dapat menjangkau proxy melalui `https://<host>:56244` tanpa peringatan sertifikat.

Jika Anda lebih memilih tetap pada HTTP polos, biarkan konfigurasi apa adanya dan ganti `https://` dengan `http://` di cuplikan client di bawah. Kedua skema bekerja; perbedaannya adalah port mana yang menjawab handshake TLS.

**Fallback loopback.** Client di host yang sama yang tidak dapat menghormati CA wall-vault (terutama runtime Node bawaan OpenClaw, yang menulis ulang `NODE_EXTRA_CA_CERTS` saat spawn) menjangkau proxy melalui pendamping HTTP-polos khusus loopback di `127.0.0.1:56245`. wall-vault mengaktifkannya secara otomatis ketika TLS aktif.

---

## Menghubungkan client

Arahkan client AI mana pun ke `https://<host>:56244` (atau `http://...` jika TLS mati). Proxy menjawab empat bentuk:

| Format | Path | Contoh client |
|--------|------|-----------------|
| Gemini | `/google/v1beta/models/<model>:generateContent` | OpenClaw, Gemini CLI, Antigravity |
| Anthropic | `/v1/messages` | Claude Code, Anthropic SDK |
| OpenAI | `/v1/chat/completions` | Cursor, Continue, script kustom, sebagian besar aplikasi LLM |
| Ollama-native | `/api/chat` | Client Ollama yang melewatinya |

### Claude Code

```bash
export ANTHROPIC_BASE_URL=https://localhost:56244
export ANTHROPIC_API_KEY=<your-vault-client-token>
claude
```

Ketika kredit Anthropic upstream habis, dispatch beralih ke provider mana pun yang Anda atur di `fallback_services` untuk client ini. Untuk secara eksplisit memilih ikut serta pada fallback non-Claude:

```yaml
proxy:
  anthropic_fallback_model: "claude-haiku-4-5-20251001"
```

(Default kosong membuat dispatch mengembalikan error agar misrouting muncul segera.)

### Cursor / Continue

Di Cursor **Settings → AI → OpenAI API**:

```
Base URL:  https://localhost:56244
API Key:   <your-vault-client-token>
Model:     gemini-2.5-flash    # atau model apa pun yang dikenali wall-vault
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

OpenClaw adalah kerangka agent TUI yang awalnya menjadi alasan wall-vault dibangun. Modal **Add Agent** di dasbor mengatur tipe agent ke `openclaw` (atau `nanoclaw`); wall-vault kemudian menulis langsung `~/.openclaw/openclaw.json`, termasuk URL provider, vault token, dan entri model:

```bash
VAULT_CLIENT_ID=my-bot \
VAULT_URL=http://<vault-host>:56243 \
VAULT_TOKEN=<your-client-token> \
wall-vault proxy
```

### curl / script

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

## Konfigurasi

`wall-vault setup` menulis salah satu di antara `./wall-vault.yaml` atau `~/.wall-vault/config.yaml`. Sunting secara manual untuk field yang tidak ditanyakan oleh wizard.

```yaml
mode: standalone     # standalone | distributed
lang: en             # en | ko | zh | ja | es | fr | de | pt | ar | hi | id | th | sw | ha | ne | mn | zu
theme: light         # light | dark | cherry | ocean | gold | autumn | winter

proxy:
  port: 56244
  host: ""           # default: 127.0.0.1 untuk standalone, 0.0.0.0 untuk distributed
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
  plain_port: 56245              # pendamping HTTP khusus loopback ketika TLS aktif
  ollama_keep_alive: "30m"       # "-1" jangan pernah unload, "0" unload segera
  ollama_num_ctx: 8192
  oai_stream_forward: false      # opt-in passthrough SSE backend asli
  anthropic_fallback_model: ""   # opt-in rewrite non-Claude pada dispatch anthropic

vault:
  port: 56243
  host: ""
  admin_token: ""
  admin_ip_whitelist: []
  master_password: ""            # password enkripsi key AES-GCM
  data_dir: ~/.wall-vault/data
  services_dir: ~/.wall-vault/services
  tls:
    enabled: false
    cert_file: ""
    key_file: ""
  bootstrap_port: 56247          # listener HTTP-polos yang hanya menyajikan ca.crt

doctor:
  interval: 5m
  auto_fix: true
  log_file: /tmp/wall-vault-doctor.log

hooks:
  on_model_change: ""            # perintah shell (env: SERVICE, MODEL)
  on_key_exhausted: ""
  on_service_down: ""
  on_doctor_fix: ""
  openclaw_socket: ""
```

### Environment variable

Setiap field YAML memiliki override env yang menang atas file. Yang umum:

| Variabel | Deskripsi |
|----------|-------------|
| `WV_LANG`, `WV_THEME` | Bahasa dan tema |
| `WV_PROXY_PORT`, `WV_PROXY_HOST` | Alamat listen proxy |
| `WV_VAULT_PORT`, `WV_VAULT_HOST` | Alamat listen vault |
| `WV_VAULT_URL`, `WV_VAULT_TOKEN` | Endpoint mode distributed |
| `WV_ADMIN_TOKEN`, `WV_MASTER_PASS` | Kredensial vault |
| `WV_KEY_GOOGLE`, `WV_KEY_OPENROUTER`, `WV_KEY_ANTHROPIC`, `WV_KEY_OPENAI` | API key (dipisahkan koma untuk beberapa) |
| `WV_PROXY_TLS_ENABLED`, `WV_PROXY_TLS_CERT`, `WV_PROXY_TLS_KEY` | TLS proxy |
| `WV_VAULT_TLS_ENABLED`, `WV_VAULT_TLS_CERT`, `WV_VAULT_TLS_KEY` | TLS vault |
| `WV_PROXY_PLAIN_PORT` | Pendamping HTTP loopback (`0` untuk menonaktifkan) |
| `WV_VAULT_BOOTSTRAP_PORT` | Listener bootstrap CA (`0` untuk menonaktifkan) |
| `WV_OLLAMA_URL`, `WV_OLLAMA_KEEP_ALIVE`, `WV_OLLAMA_NUM_CTX` | Penyetelan Ollama |
| `WV_LMSTUDIO_URL`, `WV_VLLM_URL`, `WV_LLAMACPP_URL` | Override backend lokal |
| `WV_TOKEN_SENTINEL_FALLBACK` | Substitusi sentinel "proxy-managed" loopback |
| `WV_OAI_STREAM_FORWARD` | Passthrough SSE backend asli kompatibel-OpenAI |
| `WV_ANTHROPIC_FALLBACK_MODEL` | Rewrite non-Claude opt-in pada anthropic |

---

## Mode

### Standalone (default)

Vault dan proxy berjalan dalam proses yang sama. Terbaik untuk satu host yang menampung baik key maupun agent. Mendengarkan hanya pada loopback secara default.

```bash
wall-vault start    # menjalankan keduanya
```

### Distributed

Vault berjalan di satu host (**vault host**) dan menyimpan semua key; beberapa proxy di host lain masing-masing diautentikasi dengan token per-client. Berguna ketika beberapa mesin membutuhkan key yang sama tanpa menyalinnya ke sana-sini.

**Vault host:**

```bash
WV_VAULT_HOST=0.0.0.0 wall-vault vault
```

**Setiap proxy host:**

```bash
WV_VAULT_URL=http://<vault-host>:56243 \
WV_VAULT_TOKEN=<this-client-token> \
WV_PROXY_HOST=0.0.0.0 \
wall-vault proxy
```

Modal **Add Client** di dasbor mencetak token, mendaftarkan tipe agent, dan proxy mengambil konfigurasinya melalui SSE tanpa restart.

---

## Plugin yaml (backend drop-in)

Setiap backend yang kompatibel dengan OpenAI dapat ditambahkan sebagai yaml di bawah `~/.wall-vault/services/`. wall-vault mengambilnya saat start, mendaftarkannya sebagai layanan yang dapat dirutekan, dan dispatch + himpunan deteksi OAI-compat + jembatan Gemini-stream semuanya melihatnya tanpa perubahan kode.

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
inline_no_think_for_qwen3: false   # ikut serta jika backend Anda menanggalkan marker
```

Topologi hub (satu wall-vault menampung wall-vault lain) didukung melalui `tls_internal_ca: true`, `auth.type: bearer`, dan `preserve_model_id: true`.

---

## Build dari source

```bash
git clone https://github.com/sookmook/wall-vault
cd wall-vault
go build -o wall-vault .
./wall-vault setup
./wall-vault start
```

Cross-compile untuk seluruh set yang didukung:

```bash
make build-all   # linux-amd64, linux-arm64, darwin-amd64, darwin-arm64, windows-amd64
```

Versi mengikuti `v{major}.{minor}.{patch}.{YYYYMMDD}.{HHmmss}`; `BASE_VERSION` di Makefile mengatur prefix.

### Tata letak proyek

```
wall-vault/
├── main.go                     # CLI dispatch (start/proxy/vault/setup/cert/doctor)
├── cmd/
│   ├── setup/                  # wizard setup interaktif
│   └── cert/                   # CA internal + penerbit sertifikat TLS per-host
├── internal/
│   ├── config/                 # loader YAML + env, loader plugin
│   ├── proxy/                  # dispatch request, rotasi key, konverter format
│   ├── vault/                  # store AES-GCM, dasbor, broker SSE
│   ├── doctor/                 # probe kesehatan + auto-fix
│   ├── hooks/                  # pemicu event perintah shell
│   └── i18n/                   # string UI 17 bahasa
├── configs/services/           # plugin yaml bawaan (lmstudio, vllm, ollama, …)
└── docs/                       # MANUAL, API reference, 16 varian lokal
```

---

## Dokumentasi

- [Manual pengguna](docs/MANUAL.en.md) — instalasi, dasbor, agent, troubleshooting
- [API reference](docs/API.en.md) — setiap endpoint dengan bentuk request/response
- [CHANGELOG](CHANGELOG.md)

---

## Tech stack

- Go 1.25, satu binary statis tunggal
- [templ](https://templ.guide) untuk dasbor yang dirender di server, [HTMX](https://htmx.org) untuk pembaruan parsial
- AES-GCM (key turunan PBKDF2) untuk enkripsi key saat istirahat
- Server-Sent Events untuk sinkronisasi konfigurasi langsung antara vault dan proxy
- CA internal yang ditandatangani sendiri + sertifikat per-host (tidak perlu DNS publik / Let's Encrypt)

## Lisensi

GPL-3.0. Lihat [LICENSE](LICENSE).

## Kontribusi

Pull request dipersilakan. Lihat [CONTRIBUTING.md](CONTRIBUTING.md). Untuk perubahan yang lebih besar mohon buka issue terlebih dahulu untuk membahas desainnya.
