# wall-vault User Manual

[English](MANUAL.md) · [한국어](MANUAL.ko.md) · [中文](MANUAL.zh.md) · [日本語](MANUAL.ja.md) · [Español](MANUAL.es.md) · [Français](MANUAL.fr.md) · [Deutsch](MANUAL.de.md) · [Português](MANUAL.pt.md) · [العربية](MANUAL.ar.md) · [हिन्दी](MANUAL.hi.md) · [Bahasa Indonesia](MANUAL.id.md) · [ภาษาไทย](MANUAL.th.md) · [Kiswahili](MANUAL.sw.md) · [Hausa](MANUAL.ha.md) · [नेपाली](MANUAL.ne.md) · Монгол · [isiZulu](MANUAL.zu.md)

Энэ гарын авлага нь wall-vault-ыг суулгах, тохируулах, ажиллуулах талаар тайлбарладаг. Товч тоймыг [README](../README.md)-ээс үзнэ үү. HTTP API-ийн дэлгэрэнгүй мэдээллийг [API reference](API.md)-ээс үзнэ үү.

## Агуулга

1. [wall-vault юу хийдэг вэ](#wall-vault-юу-хийдэг-вэ)
2. [Суулгац](#суулгац)
3. [setup wizard-аар анх удаа ажиллуулах](#setup-wizard-аар-анх-удаа-ажиллуулах)
4. [TLS-ийг идэвхжүүлэх](#tls-ийг-идэвхжүүлэх)
5. [API key бүртгэх](#api-key-бүртгэх)
6. [Agent холбох](#agent-холбох)
7. [Dashboard](#dashboard)
8. [Distributed горим](#distributed-горим)
9. [Авто-эхлэлт](#авто-эхлэлт)
10. [Plugin yamls](#plugin-yamls)
11. [Doctor](#doctor)
12. [Hooks](#hooks)
13. [Орчны хувьсагчид](#орчны-хувьсагчид)
14. [Алдааг засах](#алдааг-засах)

---

## wall-vault юу хийдэг вэ

wall-vault бол хамтран ажилладаг хоёр үйлчилгээг нэгтгэсэн нэг Go binary юм:

- **Vault** API key-уудыг амрах үед нь шифрлэн хадгалдаг (мастер нууц үгтэй AES-GCM), key тус бүрийн хэрэглээ болон cooldown-ыг хянадаг, өөрчлөлтүүдийг Server-Sent Events (SSE)-ээр дамжуулдаг, мөн хүний оператор нарт зориулсан вэб dashboard-ыг `:56243` дээр үйлчилгээгээр хангадаг.
- **Proxy** Gemini, Anthropic, OpenAI-нийцтэй, Ollama-native endpoint-уудыг `:56244` дээр илчилнэ. Proxy руу заасан AI client нь vault дахь key-уудыг ашигладаг — client-ууд тэдгээрийг хэзээ ч харахгүй. Нэг upstream бүтэлгүйтэх үед dispatch нь дарааллын дагуу дараагийн provider руу шилждэг.

Энэ нь дараах тохиолдолд хэрэгтэй:

- Танд хэд хэдэн provider-ийн key-ууд байгаа бөгөөд agent ярилцах нэг URL хэрэгтэй бол.
- Та cooldown дээр байгаа free-tier key-г session эвдэхгүйгээр зайлуулахыг хүсэж байгаа бол.
- Та credentials хуулахгүйгээр нэг LAN дээр олон bot, IDE, эсвэл script-уудыг ижил key-уудаар тэжээхийг хүсэж байгаа бол.
- Та key-уудыг засах, model-уудыг солихын тулд environment variable биш dashboard ашиглахыг хүсэж байгаа бол.
- Cloud хязгаар дуусахад орон нутгийн fallback (Ollama, LM Studio, vLLM) хэрэгтэй бол.

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

## Суулгац

### Linux / macOS one-liner

```bash
curl -fsSL https://raw.githubusercontent.com/sookmook/wall-vault/main/install.sh | sh
```

Скрипт нь OS болон architecture-ыг автоматаар илрүүлж, зөв binary-г `~/.local/bin/wall-vault` руу татаж аваад, executable болгоно. Хэрэв `~/.local/bin` таны `PATH` дээр байхгүй бол нэмнэ үү:

```bash
export PATH="$HOME/.local/bin:$PATH"
```

### Гар татах

Урьдчилан бэлдсэн binary-уудыг release бүрт `https://github.com/sookmook/wall-vault/releases` дээр нийтэлдэг.

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

### Source-оос build хийх

Go 1.25 эсвэл түүнээс шинэ хувилбар шаардлагатай.

```bash
git clone https://github.com/sookmook/wall-vault
cd wall-vault
go build -o wall-vault .
```

`make build-all` нь дэмжигдсэн таван platform руу cross-compile хийнэ. Binary-ууд `bin/`-д үлдэнэ.

---

## setup wizard-аар анх удаа ажиллуулах

```bash
wall-vault setup
```

Wizard нь танаас дарааллын дагуу асууна:

1. **Хэл** — 17 UI locale-ээс нэгийг сонгоно. `$LANG`-аас автоматаар илрүүлдэг; wizard ямартай ч жагсаалтыг санал болгоно.
2. **Theme** — `light` (өгөгдмөл), `dark`, `cherry`, `ocean`, `gold`, `autumn`, `winter`. Зөвхөн гадаад харагдац.
3. **Горим** — `standalone` (нэг host, өгөгдмөл) эсвэл `distributed` (нэг host дээр vault, бусад дээр proxy-нууд).
4. **Bot нэр** — чөлөөт `client_id` slug. Vault нь үүнийг client-тус бүрийн config (model overrides, fallback chains) хязгаарлахад ашигладаг.
5. **Proxy port** — өгөгдмөл `56244`.
6. **Vault port** — өгөгдмөл `56243` (зөвхөн standalone).
7. **Үйлчилгээ сонголт** — дараах тус бүрд y/N: Google Gemini, OpenRouter, Anthropic, OpenAI, Ollama, LM Studio, vLLM. Олон сонголт зүгээр; тус бүр эцэст нь өөрийн env-var зөвлөмжийг бичнэ.
8. **Tool filter** — `strip_all` (өгөгдмөл; аюулгүй байдлын үүднээс ирж буй бүх tool definition-уудыг хаадаг) эсвэл `passthrough` (ямар ч tool-ыг нэвтрүүлдэг).
9. **Admin token** — автоматаар үүсгэхийн тулд хоосон үлдээ. Dashboard-д нэвтрэхэд энэ token шаардлагатай.
10. **Мастер нууц үг** — шифрлэхгүй бол хоосон үлдээ (зөвлөдөггүй); амрах үед key store-ыг AES-GCM шифрлэхийн тулд утга тохируул.
11. **Хадгалах зам** — одоогийн directory дахь `wall-vault.yaml` руу өгөгдмөл. Loader нь `~/.wall-vault/config.yaml`-ыг ч мөн харна.

Хадгалсны дараа wizard нь `doctor.FixTrust`-ыг ажиллуулдаг ингэснээр орон нутагт суулгасан ямар ч agent (OpenClaw, Claude Code, Cline) wall-vault-ийн дотоод CA-г trust store-доо автоматаар нэмж авах болно. Хэрэв ийм agent суулгаагүй бол энэ алхам `SKIP` хэвлээд юу ч бичихгүй.

Дараа нь binary-г эхлүүл:

```bash
wall-vault start
```

`start` нь vault болон proxy-г нэг process-д хамтад нь ажиллуулна (standalone горим). Distributed горимын хувьд vault host дээр `wall-vault vault`, proxy host тус бүр дээр `wall-vault proxy` ашигла.

Browser дээр `http://localhost:56243`-ыг нээ. Wizard-ийн хэвлэсэн admin token-оор нэвтэр.

---

## TLS-ийг идэвхжүүлэх

Wizard-ийн өгөгдмөл нь хоёр listener-ыг энгийн HTTP дээр үлдээдэг. Ихэнх agent (OpenClaw, Claude Code, Cursor) нэг HTTPS endpoint-той илүү сайн ажилладаг, тиймээс орон нутгийн машинаас илүү тэлэх deployment-д TLS-ийг зөвлөдөг.

wall-vault нь өөрийн дотоод CA-тай хамт ирдэг тул нийтийн DNS нэр эсвэл Let's Encrypt шаардагдахгүй.

```bash
# 1. Дотоод CA үүсгэх — ~/.wall-vault/ca.{crt,key}-д бичигдэнэ.
#    CA нь өгөгдмөлөөр 10 жил хүчинтэй; --ca-years-аар override хийнэ.
wall-vault cert init

# 2. Host certificate гарга. Subject Alternative Names-д автоматаар орно:
#       hostname, "localhost", "127.0.0.1", илрүүлсэн loopback бус LAN IP-нүүд.
#    Issuer dir-ийг --dir-аар, validity-г --host-years-аар override хийнэ.
wall-vault cert issue $(hostname)

# 3. Энэ машины OS keychain-д CA-г trust хий.
#    Linux: update-ca-certificates-аар /etc/ssl/certs/ руу бичнэ (sudo шаардлагатай).
#    macOS: security add-trusted-cert-аар System keychain-д нэмнэ (sudo шаардлагатай).
#    Windows: certutil-аар CurrentUser\Root руу import хийнэ (admin шаардлагагүй).
wall-vault cert install-trust

# 4. Хоёр listener дээр TLS-г идэвхжүүлнэ.
export WV_PROXY_TLS_ENABLED=1
export WV_PROXY_TLS_CERT="$HOME/.wall-vault/$(hostname).crt"
export WV_PROXY_TLS_KEY="$HOME/.wall-vault/$(hostname).key"
export WV_VAULT_TLS_ENABLED=1
export WV_VAULT_TLS_CERT="$HOME/.wall-vault/$(hostname).crt"
export WV_VAULT_TLS_KEY="$HOME/.wall-vault/$(hostname).key"

wall-vault start
```

Бусад LAN машинд trust-ыг тэлэхийн тулд `~/.wall-vault/ca.crt`-ыг хуулж, тус бүр дээр `wall-vault cert install-trust --ca <path>`-ыг ажиллуул. Vault мөн `:56247` (**bootstrap port**) дээр жижиг plain-HTTP listener-ээр `ca.crt`-г илчилдэг шинэ client-д HTTPS ярихын тулд CA хэрэгтэй болсон catch-22 тохиолдлуудад зориулж.

### Loopback HTTP companion

Зарим agent — ялангуяа OpenClaw-ийн багцлагдсан Node runtime — process spawn-д `NODE_EXTRA_CA_CERTS`-ыг дахин бичдэг бөгөөд оператороос өгсөн CA hint-ийг хаядаг. Тэд `cert install-trust` хийсний дараа ч daemon дотроос wall-vault CA-г хүндлэж чадахгүй. wall-vault нь TLS идэвхжсэн үед `127.0.0.1:56245` дээр **зөвхөн loopback plain-HTTP listener**-ыг нэмж холбосноор үүнийг тойрч ажилладаг. Same-host client-ууд тэр port-оор TLS-гүйгээр proxy-д хүрнэ; LAN client-ууд TLS listener-ыг үргэлжлүүлэн ашиглана.

Хэрэгцээгүй бол `WV_PROXY_PLAIN_PORT=0`-аар идэвхгүй болго.

### `wall-vault cert list`

`~/.wall-vault/`-ын доорх cert бүрийг subject, validity window, SAN-уудтай нь харуулна.

```
$ wall-vault cert list
ca.crt          subject=wall-vault internal CA   not-after=2036-05-05
hostname.crt    subject=hostname                 not-after=2031-05-05   SAN=hostname,localhost,127.0.0.1,192.168.…
```

---

## API key бүртгэх

Хоёр арга: dashboard, эсвэл environment variable-ууд.

### Dashboard (зөвлөдөг)

1. Admin token-оор `https://localhost:56243`-д нэвтэр.
2. Keys card-д **+ API key** дар.
3. Үйлчилгээ сонго (Google, OpenRouter, Anthropic, OpenAI, …).
4. Key-г paste хий. Хадгал.

Үйлчилгээ тус бүрд олон key байж болно; proxy тэдгээрийн хооронд round-robin хийж, per-key cooldown-д орсон key-уудыг алгасдаг.

### Environment variables (нэг удаагийн bootstrap)

```bash
export WV_KEY_GOOGLE="AIzaSyA1...,AIzaSyB2...,AIzaSyC3..."   # comma-separated
export WV_KEY_OPENROUTER="sk-or-v1-…"
export WV_KEY_ANTHROPIC="sk-ant-…"
export WV_KEY_OPENAI="sk-…"
wall-vault start
```

Ийнхүү өгсөн key-уудыг анхны launch-ын үед encrypted store руу бичнэ. Дараагийн start-ууд нь тэдгээрийг disk-ээс уншина; та эхний run-ийн дараа env var-уудыг unset хийж болно.

### Cooldown болон солих

Амжилттай дуудлага бүр key-ийн `usage_count`-ыг нэмэгдүүлж, `last_used`-ийг шинэчилнэ. HTTP 429 / 402 / 403 дээр proxy нь key-г **cooldown** руу оруулдаг (өгөгдмөл: 429-д 60 минут, 402-д 24 цаг, 403-д 12 цаг). Дараагийн dispatch тэр үйлчилгээний өөр key-г сонгоно. Үйлчилгээний бүх key cooldown-д орсон үед proxy тэр үйлчилгээг бүхэлд нь хурдан алгасч fallback chain-ий дараагийн provider-ыг туршина.

Cooldown-ууд dashboard дээр key-тус бүрд буурах тоологчтойгоор харагдана.

---

## Agent холбох

### OpenClaw

OpenClaw анхны зорилтот client. Dashboard-ын **+ Add agent** modal-ыг ашигла:

- **Agent type**-ыг `openclaw` эсвэл `nanoclaw` болгож тохируул.
- **Work directory**-г тохируул — OpenClaw-ын хувьд энэ нь `~/.openclaw` болж автоматаар бөглөгдөнө.
- **preferred service** болон сонголтоор **model override** сонго.
- **Apply** дар. wall-vault `~/.openclaw/openclaw.json`-ыг шууд бичнэ (provider URL, vault token, model entry-үүд).

Та dashboard-аас model-ыг солих үед OpenClaw 1–3 секундийн дотор SSE-ээр өөрчлөлтийг авна — restart хийхгүй.

### Claude Code

```bash
export ANTHROPIC_BASE_URL=https://localhost:56244
export ANTHROPIC_API_KEY=<your-vault-client-token>
claude
```

Upstream Anthropic credit дуусахад dispatch нь энэ client-ын `fallback_services`-д жагсаасан үйлчилгээ рүү шилждэг. Өгөгдмөлөөр anthropic dispatch руу илгээсэн Claude бус model id error буцаадаг ингэснээр misrouting шууд гарч ирнэ. Автомат дахин бичилт рүү opt in хий:

```yaml
proxy:
  anthropic_fallback_model: "claude-haiku-4-5-20251001"
```

### Cursor

Cursor **Settings → AI → OpenAI API**-д:

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

`proxy.oai_stream_forward: true` тохируулагдсан үед мөн endpoint streaming (`"stream": true`)-ыг хүлээн авна.

---

## Dashboard

`https://localhost:56243`. Home grid дээр таван card:

- **Keys** — API key бүрийг үйлчилгээгээр бүлэглэсэн. Нэмэх, засах, устгах; хэрэглээ, cooldown-ыг харах.
- **Services** — Google / OpenRouter / Anthropic / OpenAI / Ollama / LM Studio / vLLM / llama.cpp, мөн `~/.wall-vault/services/`-ын ямар ч plugin yaml. Үйлчилгээ тус бүрд `default_model`, `allowed_models`, base URL, reasoning toggle тохируулах.
- **Clients (agents)** — бүртгэгдсэн client бүр (OpenClaw bot, Claude Code session, Cursor instance, …). Preferred service, model override, fallback chain хуваарилах.
- **Proxies** — энэ vault руу authenticate хийсэн proxy бүр. Шууд төлөв (online/offline), сүүлчийн харсан, одоогийн model.
- **Settings** — admin token, мастер нууц үг солих, theme, хэл.

Card бүр edit slideover-той (баруун тал). Гадуур дарах эсвэл `Esc` дарвал хаагдана. Өөрчлөлтүүд секундын дотор SSE-ээр холбогдсон бүх proxy руу push хийгддэг.

**Footer** нь SSE indicator (ногоон = холбогдсон, улбар шар = дахин холбогдож байна, саарал = тасарсан) болон шууд build хувилбарыг агуулдаг.

---

## Distributed горим

Танд бүгд адил key хэрэгтэй хэд хэдэн машин байгаа үед нэг host дээр vault, бусад тус бүр дээр proxy-нуудыг ажиллуул.

### Vault host

```bash
WV_VAULT_HOST=0.0.0.0 \
WV_ADMIN_TOKEN=<admin> \
WV_MASTER_PASS=<master> \
wall-vault vault
```

Dashboard одоо `https://<vault-host>:56243`-д хүртээмжтэй боллоо. **Clients** card-д алсын proxy тус бүрд agent нэмэх; тус бүр өвөрмөц `vault_token` үүсгэнэ.

### Proxy host-ууд

```bash
WV_VAULT_URL=http://<vault-host>:56243 \
WV_VAULT_TOKEN=<that-client-token> \
WV_PROXY_HOST=0.0.0.0 \
wall-vault proxy
```

Proxy нь vault руу authenticate хийж, SSE stream нээж, хүлээн авсан ямар ч config-ыг (preferred service, model override, fallback chain) хэрэгжүүлдэг. Дараагийн vault засварууд секундын дотор restart хийхгүйгээр гарч ирнэ.

LAN-тэлсэн суулгацын хувьд vault host дээр TLS идэвхжүүл (`WV_VAULT_TLS_ENABLED=1` + cert/key env var-ууд) ба proxy host тус бүрийг ижил `wall-vault cert install-trust` алхмаар ажиллуул ингэснээр proxy-ийн vault руу хийх HTTPS дуудлагууд найдвартай байх болно.

---

## Авто-эхлэлт

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

Ижил host дээрх vault-ын хувьд параллел `wall-vault-vault.service` бичнэ. Standalone горимын хувьд `wall-vault start` дуудах нэг unit хангалттай.

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

`wall-vault.exe start`-г Windows service болгож wrap хийхэд `nssm` ашигла, эсвэл хэрэглэгч logon хийх үед ажиллах `schtasks` бичлэг ашигла.

---

## Plugin yamls

Ямар ч OpenAI-нийцтэй backend-ыг `~/.wall-vault/services/`-ын доор yaml хаяснаар код өөрчлөлгүйгээр нэмж болно. wall-vault нь үүнийг startup-д ачаалж, dispatch, OAI-compat detection set, Gemini-stream bridge-д үйлчилгээг бүртгэдэг.

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

`configs/services/`-д багцлагдсан багц (lmstudio, vllm, llamacpp, tgwui, localai, jan, koboldcpp, tabbyapi, mlx-server, litellm-proxy, ollama, google, openrouter) өгөгдмөлөөр идэвхгүйгээр ирдэг. Хэрэгтэйг нь `~/.wall-vault/services/` руу хуулж, `enabled: true` болгож, restart хий.

---

## Doctor

`wall-vault doctor` нь бүх суулгац дээр нэг удаагийн health probe ажиллуулдаг:

```
✓ vault listener  (https://localhost:56243)
✓ proxy listener  (https://localhost:56244)
✓ master password set
⚠ Google: 2 keys, all on cooldown
✓ Anthropic: 1 key healthy
✗ Ollama: not reachable at http://localhost:11434
```

Мөр бүр дараахын аль нэг:

- `✓` — эрүүл
- `⚠` — доройтсон ч ажиллаж байгаа (нэг key cooldown-д орсон, бага quota гэх мэт)
- `✗` — эвдэрсэн
- `SKIP` — тохируулаагүй / энэ host-д хамаагүй

Хоёр дахь daemon горим нь ижил probe-ыг `doctor.interval` бүр (өгөгдмөл 5 минут) ажиллуулж, үр дүнг `doctor.log_file`-д (өгөгдмөл `/tmp/wall-vault-doctor.log`) бичнэ. `doctor.auto_fix` true үед энэ нь нийтлэг drift-ийг засахыг (хуучин OpenClaw config, дутуу TLS trust, дахин эхлүүлэх боломжтой үйлчилгээнүүд) оролдоно.

Dashboard-аас **Doctor** card эсвэл `wall-vault doctor`-аар нэг удаагийн trigger хий.

---

## Hooks

Key event-үүд дээр shell command ажиллуул:

```yaml
hooks:
  on_model_change:   "logger 'wall-vault: $SERVICE/$MODEL'"
  on_key_exhausted:  "notify-send 'wall-vault' '$SERVICE keys all on cooldown'"
  on_service_down:   "/usr/local/bin/page-oncall.sh $SERVICE '$ERROR'"
  on_doctor_fix:     "echo \"$AGENT: $LEVEL $MSG\" >> ~/wall-vault.audit.log"
  openclaw_socket:   ""    # if set, OpenClaw TUI receives events over this Unix socket
```

Hook бүр event-тусгайлан environment variable-ууд (`SERVICE`, `MODEL`, `ERROR`, `AGENT`, `LEVEL`, `MSG`) авдаг. Hook-ууд 5 секундын timeout-той async ажилладаг — proxy удаан hook дээр хэзээ ч block хийхгүй.

---

## Орчны хувьсагчид

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

Тохируулагдсан env var бүр YAML file-аас давамгайлдаг.

---

## Алдааг засах

### `:56244` дээр `connection refused`

Эсвэл proxy ажиллахгүй байгаа, эсвэл өөр host-д холбогдсон байна. Шалга:

```bash
ss -lnp | grep 56244
systemctl --user status wall-vault-proxy   # Linux
launchctl list | grep wall-vault           # macOS
```

Хэрэв энэ нь өөр port дээр ажиллаж байгаа бол, таны config-д `proxy.port` override хийгдсэн байна — `~/.wall-vault/config.yaml`-г шалга.

### `x509: certificate signed by unknown authority`

Client wall-vault дотоод CA-г trust хийдэггүй. Client машин дээр `wall-vault cert install-trust` ажиллуул. OS trust store-ыг үл тоох runtime-тай agent-ын хувьд (жишээ нь `NODE_EXTRA_CA_CERTS` нь hardcoded байгаа Node), `127.0.0.1:56245` (зөвхөн нэг host) дахь loopback HTTP companion-ыг ашигла, эсвэл энгийн HTTP руу шилжихийн тулд `WV_PROXY_TLS_ENABLED=0` тохируул.

### `token not registered with vault`

Client-ын `Authorization: Bearer <token>` бүртгэгдсэн ямар ч client-тай таарахгүй байна. Dashboard-ын **Clients** доор token-ыг шалга. Хэрэв та хуучин config-аас `proxy-managed`, `dummy`, `""` зэрэг token-ыг шууд хуулсан бол үүнийг жинхэнэ client token-оор солино.

### `Anthropic dispatch needs a Claude model id`

v0.2.63-аас хойшхи өгөгдмөл зан үйл: anthropic dispatch руу илгээсэн Claude бус model id error буцаадаг. Routing-аа засах (anthropic руу `gemini-2.5-flash` бүү илгээ) эсвэл `proxy.anthropic_fallback_model`-ээр автомат дахин бичилт рүү opt in хий.

### `unknown service: <id>`

Dispatch ямар ч plugin yaml эзэмшээгүй service id-г харсан. Шалга:

```bash
ls ~/.wall-vault/services/        # any plugin yaml present?
cat ~/.wall-vault/services/<id>.yaml | grep enabled
```

Хэрэв yaml байгаа боловч `enabled: false` бол түүнийг сольж тохируул. Хэрэв энэ нь бүхэлдээ дутуу бол source tree-ийн `configs/services/`-аас хуул.

### Reasoning model дээр хоосон хариу

`qwen3.6`, `deepseek-r1`, GPT-`o1` гэр бүл нь заримдаа зөвхөн `reasoning_content` гаргаж, `content`-ыг хоосон үлдээдэг. v0.2.63-аас хойш wall-vault автоматаар reasoning text руу шилжидэг — хэрэв та одоо ч хоосон хариу харж байгаа бол backend ямар ч field буцаахгүй байна. Upstream-ийн log-уудыг шалга.

LM Studio-той qwen3-ын хувьд тусгайлан plugin yaml-д `inline_no_think_for_qwen3: true` тохируул ингэснээр reasoning inline идэвхгүй болно. Built-in lmstudio.yaml болон ollama.yaml аль хэдийн үүнийг хийсэн байдаг.

### Dashboard "all keys on cooldown" гэж харуулдаг ч би сая нэгийг нэмлээ

Шинэ key эрүүл боловч dispatch path хуучин key-ийн cooldown-д хэвээр байж магадгүй. Шинэ хүсэлт хий — proxy дуудлага тус бүрд round-robin хийдэг бөгөөд эрүүл key дараагийнх нь сонгогдоно.

### Vault мастер нууц үгээр түгжээ тайлахгүй

Буруу нууц үг. Сэргээх боломжгүй — wall-vault зориудаар backdoor илгээдэггүй. Хэрэв та үнэхээр мастер нууц үгээ алдсан бол цорын ганц зам нь `~/.wall-vault/data/vault.json`-ыг устгаад, шинэ нууц үгээр restart хийж, key-уудыг дахин нэмэх юм.

### Free-tier OpenRouter хязгаар хүрсэн

`proxy.services`-д `openrouter` оруулахаар тохируулж, дор хаяж нэг OpenRouter key нэм. Paid path 402 / 429 буцаах үед proxy нь paid model-аас түүний `:free` хувилбар руу автоматаар шилждэг.

### `journalctl --user -u wall-vault-proxy` хоосон байна

systemd `--user` log-ууд нь түүнийг ажиллуулж буй хэрэглэгчийн journal руу очдог. Хэрэв та unit-ыг `root`-оор эсвэл `sudo`-р эхлүүлсэн бол journal нь оронд нь system instance-д байна — `--user`-гүйгээр `journalctl -u wall-vault-proxy`-г оролдоно уу.

---

## Илүү

- HTTP API reference — [API.md](API.md)-г үзнэ үү
- Source — `https://github.com/sookmook/wall-vault`
- Bug report / feature request — GitHub Issues
- Release түүх — [CHANGELOG.md](../CHANGELOG.md)
