# Panduan Pengguna wall-vault

[English](MANUAL.md) · [한국어](MANUAL.ko.md) · [中文](MANUAL.zh.md) · [日本語](MANUAL.ja.md) · [Español](MANUAL.es.md) · [Français](MANUAL.fr.md) · [Deutsch](MANUAL.de.md) · [Português](MANUAL.pt.md) · [العربية](MANUAL.ar.md) · [हिन्दी](MANUAL.hi.md) · **Bahasa Indonesia** · [ภาษาไทย](MANUAL.th.md) · [Kiswahili](MANUAL.sw.md) · [Hausa](MANUAL.ha.md) · [नेपाली](MANUAL.ne.md) · [Монгол](MANUAL.mn.md) · [isiZulu](MANUAL.zu.md)

Panduan ini mencakup pemasangan, konfigurasi, dan pengoperasian wall-vault. Untuk gambaran sekilas, lihat [README](../README.md). Untuk detail HTTP API, lihat [API reference](API.md).

## Daftar Isi

1. [Apa yang dilakukan wall-vault](#apa-yang-dilakukan-wall-vault)
2. [Pemasangan](#pemasangan)
3. [Menjalankan pertama kali dengan setup wizard](#menjalankan-pertama-kali-dengan-setup-wizard)
4. [Mengaktifkan TLS](#mengaktifkan-tls)
5. [Mendaftarkan API key](#mendaftarkan-api-key)
6. [Menghubungkan agen](#menghubungkan-agen)
7. [Dashboard](#dashboard)
8. [Mode terdistribusi](#mode-terdistribusi)
9. [Auto-start](#auto-start)
10. [Plugin yamls](#plugin-yamls)
11. [Doctor](#doctor)
12. [Hooks](#hooks)
13. [Variabel lingkungan](#variabel-lingkungan)
14. [Pemecahan masalah](#pemecahan-masalah)

---

## Apa yang dilakukan wall-vault

wall-vault adalah binary Go tunggal yang menggabungkan dua layanan yang saling bekerja sama:

- **vault** menyimpan API key terenkripsi saat istirahat (AES-GCM dengan kata sandi master), melacak penggunaan dan cooldown per key, menyiarkan perubahan melalui Server-Sent Events (SSE), dan menyajikan dashboard web di `:56243` untuk operator manusia.
- **proxy** mengekspos endpoint Gemini, Anthropic, OpenAI-compatible, dan Ollama-native di `:56244`. Klien AI mana pun yang menunjuk ke proxy akan menggunakan key di vault — klien tidak pernah melihatnya. Ketika satu upstream gagal, dispatch akan beralih ke provider berikutnya secara berurutan.

Ini berguna ketika:

- Anda memiliki key untuk beberapa provider dan menginginkan satu URL yang dihubungi agen.
- Anda ingin key tier-gratis yang sedang cooldown menyingkir tanpa merusak sesi.
- Anda ingin key yang sama menggerakkan beberapa bot, IDE, atau skrip pada LAN yang sama tanpa menyalin kredensial.
- Anda menginginkan dashboard, bukan variabel lingkungan, untuk mengedit key dan mengganti model.
- Anda menginginkan fallback lokal (Ollama, LM Studio, vLLM) saat batas cloud habis.

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

## Pemasangan

### One-liner Linux / macOS

```bash
curl -fsSL https://raw.githubusercontent.com/sookmook/wall-vault/main/install.sh | sh
```

Skrip secara otomatis mendeteksi OS dan arsitektur, mengunduh binary yang tepat ke `~/.local/bin/wall-vault`, dan menjadikannya dapat dieksekusi. Jika `~/.local/bin` tidak ada di `PATH` Anda, tambahkan:

```bash
export PATH="$HOME/.local/bin:$PATH"
```

### Unduhan manual

Binary pre-built dipublikasikan pada setiap rilis di `https://github.com/sookmook/wall-vault/releases`.

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

### Build dari source

Membutuhkan Go 1.25 atau lebih baru.

```bash
git clone https://github.com/sookmook/wall-vault
cd wall-vault
go build -o wall-vault .
```

`make build-all` melakukan cross-compile ke kelima platform yang didukung. Binary akan masuk ke `bin/`.

---

## Menjalankan pertama kali dengan setup wizard

```bash
wall-vault setup
```

Wizard meminta Anda, secara berurutan:

1. **Bahasa** — memilih salah satu dari 17 lokal UI. Dideteksi otomatis dari `$LANG`; wizard tetap menawarkan daftar.
2. **Tema** — `light` (default), `dark`, `cherry`, `ocean`, `gold`, `autumn`, `winter`. Hanya kosmetik.
3. **Mode** — `standalone` (host tunggal, default) atau `distributed` (vault di satu host, proxy di yang lain).
4. **Nama bot** — slug `client_id` bebas. Vault menggunakannya untuk membatasi konfigurasi per-klien (override model, fallback chain).
5. **Port proxy** — default `56244`.
6. **Port vault** — default `56243` (hanya standalone).
7. **Pemilihan layanan** — y/N untuk masing-masing: Google Gemini, OpenRouter, Anthropic, OpenAI, Ollama, LM Studio, vLLM. Banyak pilihan diperbolehkan; masing-masing menulis petunjuk env-var-nya di akhir.
8. **Tool filter** — `strip_all` (default; memblokir semua definisi tool yang masuk untuk keamanan) atau `passthrough` (biarkan tool apa pun lewat).
9. **Token admin** — biarkan kosong untuk dibuat otomatis. Dashboard memerlukan token ini untuk login.
10. **Kata sandi master** — biarkan kosong untuk tanpa enkripsi (TIDAK direkomendasikan); atur nilai untuk mengenkripsi penyimpanan key dengan AES-GCM saat istirahat.
11. **Path penyimpanan** — default `wall-vault.yaml` di direktori saat ini. Loader juga melihat `~/.wall-vault/config.yaml`.

Setelah menyimpan, wizard menjalankan `doctor.FixTrust` sehingga setiap agen yang terpasang secara lokal (OpenClaw, Claude Code, Cline) secara otomatis mendapat CA internal wall-vault yang ditambahkan ke trust store-nya. Jika tidak ada agen seperti itu yang terpasang, langkah ini mencetak `SKIP` dan tidak menulis apa pun.

Kemudian jalankan binary:

```bash
wall-vault start
```

`start` menjalankan vault dan proxy dalam satu proses (mode standalone). Untuk mode terdistribusi gunakan `wall-vault vault` di host vault dan `wall-vault proxy` di setiap host proxy.

Buka `http://localhost:56243` di browser. Login dengan token admin yang dicetak oleh wizard.

---

## Mengaktifkan TLS

Default wizard meninggalkan kedua listener pada HTTP biasa. Sebagian besar agen (OpenClaw, Claude Code, Cursor) bekerja lebih baik dengan satu endpoint HTTPS, jadi TLS direkomendasikan dalam deployment apa pun yang melampaui mesin lokal.

wall-vault dilengkapi dengan CA internalnya sendiri sehingga Anda tidak memerlukan nama DNS publik atau Let's Encrypt.

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

Untuk memperluas trust ke mesin LAN lain, salin `~/.wall-vault/ca.crt` dan jalankan `wall-vault cert install-trust --ca <path>` di masing-masing mesin. Vault juga mengekspos `ca.crt` melalui listener plain-HTTP kecil di `:56247` (**bootstrap port**) untuk kasus catch-22 di mana klien baru memerlukan CA untuk berbicara HTTPS.

### Pendamping HTTP loopback

Beberapa agen — terutama runtime Node yang dibundel OpenClaw — menulis ulang `NODE_EXTRA_CA_CERTS` saat proses spawn, menjatuhkan petunjuk CA apa pun yang disediakan operator. Mereka tidak dapat menghormati CA wall-vault dari dalam daemon, bahkan setelah `cert install-trust`. wall-vault mengatasi ini dengan mengikat **listener plain-HTTP khusus loopback** tambahan di `127.0.0.1:56245` setiap kali TLS diaktifkan. Klien same-host mencapai proxy melalui port itu tanpa TLS sama sekali; klien LAN tetap menggunakan listener TLS.

Nonaktifkan dengan `WV_PROXY_PLAIN_PORT=0` jika Anda tidak membutuhkannya.

### `wall-vault cert list`

Menampilkan setiap sertifikat di bawah `~/.wall-vault/` dengan subject, jendela validitas, dan SAN.

```
$ wall-vault cert list
ca.crt          subject=wall-vault internal CA   not-after=2036-05-05
hostname.crt    subject=hostname                 not-after=2031-05-05   SAN=hostname,localhost,127.0.0.1,192.168.…
```

---

## Mendaftarkan API key

Dua cara: dashboard, atau variabel lingkungan.

### Dashboard (direkomendasikan)

1. Login di `https://localhost:56243` dengan token admin.
2. Klik **+ API key** di kartu keys.
3. Pilih layanan (Google, OpenRouter, Anthropic, OpenAI, …).
4. Tempel key. Simpan.

Beberapa key per layanan diperbolehkan; proxy melakukan round-robin di antara mereka dan melewati yang terkena cooldown per-key.

### Variabel lingkungan (bootstrap satu kali)

```bash
export WV_KEY_GOOGLE="AIzaSyA1...,AIzaSyB2...,AIzaSyC3..."   # comma-separated
export WV_KEY_OPENROUTER="sk-or-v1-…"
export WV_KEY_ANTHROPIC="sk-ant-…"
export WV_KEY_OPENAI="sk-…"
wall-vault start
```

Key yang disediakan dengan cara ini ditulis ke penyimpanan terenkripsi pada peluncuran pertama. Start berikutnya membacanya dari disk; Anda dapat unset variabel lingkungan setelah run pertama.

### Cooldown dan rotasi

Setiap panggilan yang berhasil meningkatkan `usage_count` key dan memperbarui `last_used`. Pada HTTP 429 / 402 / 403, proxy menempatkan key pada **cooldown** (default: 60 menit untuk 429, 24 jam untuk 402, 12 jam untuk 403). Dispatch berikutnya memilih key yang berbeda untuk layanan tersebut. Ketika semua key untuk layanan dalam cooldown, proxy melewati layanan tersebut sepenuhnya dengan cepat dan mencoba provider berikutnya dalam fallback chain.

Cooldown terlihat per-key di dashboard dengan hitungan mundur.

---

## Menghubungkan agen

### OpenClaw

OpenClaw adalah klien target asli. Gunakan modal **+ Add agent** dashboard:

- Atur **Agent type** ke `openclaw` atau `nanoclaw`.
- Atur **Work directory** — untuk OpenClaw ini terisi otomatis sebagai `~/.openclaw`.
- Pilih **preferred service** dan opsional **model override**.
- Klik **Apply**. wall-vault menulis `~/.openclaw/openclaw.json` secara langsung (URL provider, vault token, entri model).

Ketika Anda mengubah model dari dashboard, OpenClaw mengambil perubahan melalui SSE dalam 1-3 detik — tanpa restart.

### Claude Code

```bash
export ANTHROPIC_BASE_URL=https://localhost:56244
export ANTHROPIC_API_KEY=<your-vault-client-token>
claude
```

Ketika kredit upstream Anthropic habis, dispatch fallback ke layanan apa pun yang terdaftar di `fallback_services` klien ini. Secara default, ID model non-Claude yang dikirim ke dispatch anthropic mengembalikan error sehingga misrouting muncul segera. Opt in ke rewrite otomatis:

```yaml
proxy:
  anthropic_fallback_model: "claude-haiku-4-5-20251001"
```

### Cursor

Di Cursor **Settings → AI → OpenAI API**:

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

### HTTP kustom

```bash
curl -X POST https://localhost:56244/v1/chat/completions \
  -H "Authorization: Bearer <your-vault-client-token>" \
  -H "Content-Type: application/json" \
  -d '{
    "model": "gemini-2.5-flash",
    "messages": [{"role": "user", "content": "hello"}]
  }'
```

Endpoint yang sama menerima streaming (`"stream": true`) ketika `proxy.oai_stream_forward: true` diatur.

---

## Dashboard

`https://localhost:56243`. Lima kartu di grid utama:

- **Keys** — setiap API key, dikelompokkan berdasarkan layanan. Tambah, edit, hapus; lihat penggunaan dan cooldown.
- **Services** — Google / OpenRouter / Anthropic / OpenAI / Ollama / LM Studio / vLLM / llama.cpp, ditambah plugin yaml apa pun di `~/.wall-vault/services/`. Atur per-layanan `default_model`, `allowed_models`, base URL, toggle reasoning.
- **Clients (agents)** — setiap klien terdaftar (bot OpenClaw, sesi Claude Code, instance Cursor, …). Tetapkan layanan pilihan, override model, fallback chain.
- **Proxies** — setiap proxy yang telah diautentikasi terhadap vault ini. Status langsung (online/offline), terakhir dilihat, model saat ini.
- **Settings** — token admin, rotasi kata sandi master, tema, bahasa.

Setiap kartu memiliki edit slideover (sisi kanan). Klik di luar atau `Esc` menutupnya. Perubahan didorong ke semua proxy yang terhubung melalui SSE dalam hitungan detik.

**Footer** membawa indikator SSE (hijau = terhubung, oranye = menghubungkan kembali, abu-abu = terputus) dan versi build langsung.

---

## Mode terdistribusi

Ketika Anda memiliki beberapa mesin yang semuanya membutuhkan key yang sama, jalankan vault di satu host dan proxy di masing-masing yang lain.

### Host vault

```bash
WV_VAULT_HOST=0.0.0.0 \
WV_ADMIN_TOKEN=<admin> \
WV_MASTER_PASS=<master> \
wall-vault vault
```

Dashboard sekarang dapat dijangkau di `https://<vault-host>:56243`. Tambahkan agen untuk setiap proxy jarak jauh di kartu **Clients**; masing-masing menghasilkan `vault_token` unik.

### Host proxy

```bash
WV_VAULT_URL=http://<vault-host>:56243 \
WV_VAULT_TOKEN=<that-client-token> \
WV_PROXY_HOST=0.0.0.0 \
wall-vault proxy
```

Proxy mengautentikasi terhadap vault, membuka stream SSE, dan menerapkan konfigurasi apa pun yang diterimanya (layanan pilihan, override model, fallback chain). Edit vault berikutnya mendarat dalam hitungan detik tanpa restart.

Untuk instalasi yang melintasi LAN, aktifkan TLS pada host vault (`WV_VAULT_TLS_ENABLED=1` + variabel lingkungan cert/key) dan jalankan setiap host proxy melalui langkah `wall-vault cert install-trust` yang sama sehingga panggilan HTTPS proxy ke vault dipercaya.

---

## Auto-start

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

Untuk vault di host yang sama, tulis `wall-vault-vault.service` paralel. Untuk mode standalone, satu unit yang memanggil `wall-vault start` sudah cukup.

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

Gunakan `nssm` untuk membungkus `wall-vault.exe start` sebagai layanan Windows, atau entri `schtasks` yang berjalan saat user logon.

---

## Plugin yamls

Backend OpenAI-compatible apa pun dapat ditambahkan tanpa perubahan kode dengan menjatuhkan yaml di bawah `~/.wall-vault/services/`. wall-vault memuatnya pada startup dan mendaftarkan layanan untuk dispatch, set deteksi OAI-compat, dan jembatan Gemini-stream.

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

Set yang dibundel di `configs/services/` (lmstudio, vllm, llamacpp, tgwui, localai, jan, koboldcpp, tabbyapi, mlx-server, litellm-proxy, ollama, google, openrouter) dikirim dinonaktifkan secara default. Salin yang Anda inginkan ke `~/.wall-vault/services/`, atur `enabled: true`, restart.

---

## Doctor

`wall-vault doctor` menjalankan probe kesehatan sekali jalan di seluruh instalasi:

```
✓ vault listener  (https://localhost:56243)
✓ proxy listener  (https://localhost:56244)
✓ master password set
⚠ Google: 2 keys, all on cooldown
✓ Anthropic: 1 key healthy
✗ Ollama: not reachable at http://localhost:11434
```

Setiap baris adalah salah satu dari:

- `✓` — sehat
- `⚠` — terdegradasi tetapi berfungsi (satu key cooldown, kuota rendah, dll.)
- `✗` — rusak
- `SKIP` — tidak dikonfigurasi / tidak berlaku pada host ini

Mode daemon kedua menjalankan probe yang sama setiap `doctor.interval` (default 5 menit) dan menulis hasil ke `doctor.log_file` (default `/tmp/wall-vault-doctor.log`). Ketika `doctor.auto_fix` true, ia juga mencoba memperbaiki drift umum (konfigurasi OpenClaw basi, trust TLS hilang, layanan yang dapat di-restart).

Trigger sekali jalan dari dashboard melalui kartu **Doctor** atau `wall-vault doctor`.

---

## Hooks

Jalankan perintah shell pada event utama:

```yaml
hooks:
  on_model_change:   "logger 'wall-vault: $SERVICE/$MODEL'"
  on_key_exhausted:  "notify-send 'wall-vault' '$SERVICE keys all on cooldown'"
  on_service_down:   "/usr/local/bin/page-oncall.sh $SERVICE '$ERROR'"
  on_doctor_fix:     "echo \"$AGENT: $LEVEL $MSG\" >> ~/wall-vault.audit.log"
  openclaw_socket:   ""    # if set, OpenClaw TUI receives events over this Unix socket
```

Setiap hook mendapat variabel lingkungan spesifik-event (`SERVICE`, `MODEL`, `ERROR`, `AGENT`, `LEVEL`, `MSG`). Hook berjalan async dengan timeout 5 detik — proxy tidak pernah blok pada hook yang lambat.

---

## Variabel lingkungan

| Variabel | Field YAML |
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
| `WV_PROXY_TLS_REQUIRED` | `proxy.tls.required` (refuse to start with TLS off — fails closed when set) |
| `WV_PROXY_ALLOW_CIDRS` | `proxy.allow_cidrs` (comma-separated list, e.g. `192.168.0.0/16,10.0.0.0/8`; loopback always passes) |
| `WV_VAULT_TLS_ENABLED` | `vault.tls.enabled` |
| `WV_VAULT_TLS_CERT` | `vault.tls.cert_file` |
| `WV_VAULT_TLS_KEY` | `vault.tls.key_file` |
| `WV_VAULT_BOOTSTRAP_PORT` | `vault.bootstrap_port` |
| `WV_PROXY_PLAIN_PORT` | `proxy.plain_port` |
| `WV_KEY_GOOGLE` | One-shot import: comma-separated Google keys |
| `WV_KEY_OPENROUTER` | One-shot import: OpenRouter keys |
| `WV_KEY_ANTHROPIC` | One-shot import: Anthropic keys |
| `WV_KEY_OPENAI` | One-shot import: OpenAI keys |
| `WV_OLLAMA_URL` | Per-host Ollama URL override (single instance) |
| `WV_OLLAMA_URLS` | Comma-separated Ollama URLs (multi-instance dispatch) |
| `WV_OLLAMA_KEEP_ALIVE` | `proxy.ollama_keep_alive` |
| `WV_OLLAMA_NUM_CTX` | `proxy.ollama_num_ctx` |
| `WV_LMSTUDIO_URL`, `WV_VLLM_URL`, `WV_LLAMACPP_URL` | Per-backend URL override (single instance) |
| `WV_TOKEN_SENTINEL_FALLBACK` | `proxy.token_sentinel_fallback` |
| `WV_OAI_STREAM_FORWARD` | `proxy.oai_stream_forward` |
| `WV_ANTHROPIC_FALLBACK_MODEL` | `proxy.anthropic_fallback_model` |
| `WV_INJECT_MODEL_IDENTITY` | `proxy.inject_model_identity` (system-message identity guard, off by default) |
| `WV_PROMPT_TOKEN_CAP` | Per-host auto-truncate cap for local OAI-compat prompts (positive int = enable, 0 = off) |
| `WV_DISPATCH_TRACE` | Set to `1` to log every dispatch's resolved service / model with reason (off by default) |
| `WV_ECONOWORLD_MAX_TOKENS` | `proxy.econoworld_max_tokens` |
| `WV_ECONOWORLD_STREAM` | `proxy.econoworld_stream` |
| `WV_ECONOWORLD_REQUEST_TIMEOUT` | `proxy.econoworld_request_timeout` |

Setiap variabel env, ketika diatur, mengalahkan file YAML.

---

## Pemecahan masalah

### `connection refused` di `:56244`

Entah proxy tidak berjalan atau terikat ke host yang berbeda. Periksa:

```bash
ss -lnp | grep 56244
systemctl --user status wall-vault-proxy   # Linux
launchctl list | grep wall-vault           # macOS
```

Jika berjalan di port yang berbeda, konfigurasi Anda memiliki `proxy.port` yang dioverride — periksa `~/.wall-vault/config.yaml`.

### `x509: certificate signed by unknown authority`

Klien tidak mempercayai CA internal wall-vault. Jalankan `wall-vault cert install-trust` di mesin klien. Untuk agen yang runtime-nya mengabaikan trust store OS (mis. Node dengan `NODE_EXTRA_CA_CERTS` hardcoded), gunakan pendamping HTTP loopback di `127.0.0.1:56245` (hanya same-host) atau atur `WV_PROXY_TLS_ENABLED=0` untuk fallback ke HTTP biasa.

### `token not registered with vault`

`Authorization: Bearer <token>` klien tidak cocok dengan klien terdaftar mana pun. Verifikasi token di bawah **Clients** di dashboard. Jika Anda menyalin literal token seperti `proxy-managed`, `dummy`, atau `""` dari konfigurasi basi, ganti dengan token klien yang sebenarnya.

### `Anthropic dispatch needs a Claude model id`

Perilaku default per v0.2.63: ID model non-Claude yang dikirim ke dispatch anthropic mengembalikan error. Entah perbaiki routing (jangan kirim `gemini-2.5-flash` ke anthropic) atau opt in ke rewrite otomatis melalui `proxy.anthropic_fallback_model`.

### `unknown service: <id>`

Dispatch melihat ID layanan yang tidak diklaim oleh plugin yaml mana pun. Periksa:

```bash
ls ~/.wall-vault/services/        # any plugin yaml present?
cat ~/.wall-vault/services/<id>.yaml | grep enabled
```

Jika yaml ada tetapi `enabled: false`, balik. Jika hilang sepenuhnya, salin dari `configs/services/` di source tree.

### Respons kosong pada model reasoning

`qwen3.6`, `deepseek-r1`, dan keluarga GPT-`o1` kadang-kadang hanya memancarkan `reasoning_content` dan meninggalkan `content` kosong. Per v0.2.63 wall-vault fallback ke teks reasoning secara otomatis — jika Anda masih melihat respons kosong, backend tidak mengembalikan field manapun. Periksa log upstream.

Untuk LM Studio dengan qwen3 secara khusus, atur `inline_no_think_for_qwen3: true` di plugin yaml sehingga reasoning dinonaktifkan secara inline. lmstudio.yaml dan ollama.yaml bawaan sudah melakukan ini.

### Dashboard menampilkan "all keys on cooldown" tetapi saya baru saja menambahkan satu

Key baru sehat tetapi jalur dispatch mungkin masih dalam cooldown untuk key yang lebih lama. Coba permintaan baru — proxy round-robin per panggilan, dan key yang sehat akan dipilih berikutnya.

### Vault tidak akan unlock dengan kata sandi master

Kata sandi salah. Tidak ada pemulihan — wall-vault sengaja tidak mengirim backdoor. Jika Anda benar-benar kehilangan kata sandi master, satu-satunya jalan adalah menghapus `~/.wall-vault/data/vault.json`, restart dengan kata sandi baru, dan menambahkan ulang key.

### Batas tier-gratis OpenRouter terkena

Atur `proxy.services` untuk menyertakan `openrouter` dan tambahkan setidaknya satu key OpenRouter. Proxy auto-fallback dari model berbayar ke varian `:free`-nya ketika jalur berbayar mengembalikan 402 / 429.

### `journalctl --user -u wall-vault-proxy` kosong

Log systemd `--user` masuk ke journal user yang menjalankannya. Jika Anda memulai unit sebagai `root` atau via `sudo`, journal ada di instance sistem sebagai gantinya — coba `journalctl -u wall-vault-proxy` tanpa `--user`.

---

## Lainnya

- Referensi HTTP API — lihat [API.md](API.md)
- Source — `https://github.com/sookmook/wall-vault`
- Laporan bug / permintaan fitur — GitHub Issues
- Riwayat rilis — [CHANGELOG.md](../CHANGELOG.md)
