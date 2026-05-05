<p align="center">
  <img src="docs/logo.png" alt="wall-vault" width="800">
</p>

# wall-vault

> **Indawo yokugcina okhiye be-API + AI proxy ku-Go binary eyodwa.**
> Igcina okhiye ngaphakathi kwemishini nge-AES-GCM, iyabazungezisa phakathi kwabahlinzeki, iguqukela kwabanye uma omunye ehluleka, futhi iza nedeshibhodi yesikhathi sangempela.

[![License: GPL v3](https://img.shields.io/badge/License-GPLv3-blue.svg)](LICENSE)
[![Go](https://img.shields.io/badge/Go-1.25-00ADD8.svg)](go.mod)

[English](README.md) · [한국어](README.ko.md) · [中文](README.zh.md) · [日本語](README.ja.md) · [Español](README.es.md) · [Français](README.fr.md) · [Deutsch](README.de.md) · [Português](README.pt.md) · [العربية](README.ar.md) · [हिन्दी](README.hi.md) · [Bahasa Indonesia](README.id.md) · [ภาษาไทย](README.th.md) · [Kiswahili](README.sw.md) · [Hausa](README.ha.md) · [नेपाली](README.ne.md) · [Монгол](README.mn.md) · **isiZulu**

---

## Yini lokho

wall-vault ihlala phakathi kwe-AI agent (OpenClaw, Claude Code, Cursor, Continue, iskripthi sakho) nabahlinzeki be-AI besefini noma bangaphakathi ekhuluma nabo. Izinto ezimbili ku-binary eyodwa:

- **Vault** — igcina okhiye be-API abafihlwe ngenkathi bephumule (AES-GCM ngephasiwedi enkulu), ibazungezise, ilandelele ukusetshenziswa kwekey ngakukey nezikhathi zokuphumula, isakaze izinguquko nge-SSE, futhi inikeze ideshibhodi yewebhu ku-`:56243`.
- **Proxy** — yembule izinhloko ezihambisanayo ze-Gemini, Anthropic, ne-OpenAI ku-`:56244`, ikhethe ikhiye kwi-vault, ithumele ku-upstream oyilungisile, futhi iguqukele kumhlinzeki olandelayo lapho omunye ehluleka.

Isekela izinhlobo ezine zezicelo (Gemini `:generateContent`, Anthropic `/v1/messages`, OpenAI `/v1/chat/completions`, ne-Ollama-native `/api/chat`) nezigaba ezinhlanu ze-upstream:

| Umhlinzeki | Amaphuzu |
|----------|-------|
| Google Gemini | I-API yangokwemvelo; ukuzungeza okhiye ngephrojekthi ngayinye |
| Anthropic | Ukudlula okwemvelo `/v1/messages` |
| OpenAI | Eyemvelo `/v1/chat/completions` |
| OpenRouter | Amamodeli angu-340+, ukubuyela ngokuzenzakalelayo kwizinguqulo `:free` |
| Ollama / LM Studio / vLLM / llama.cpp / Jan / KoboldCpp / TabbyAPI / mlx-server / LiteLLM | Ama-backend angaphakathi ahambisana ne-OpenAI; afakwe kalula ngokusebenzisa i-plugin yaml |

Ukungeza i-backend entsha ehambisana ne-OpenAI kungumfayela owodwa we-yaml ngaphansi kwe-`~/.wall-vault/services/` — ngaphandle koshintsho lwekhodi.

## Kungani ungase uyifune

- Uphathe izinkonzo ze-AI ezintathu noma ezine futhi ufuna i-URL eyodwa lapho i-agent ikhuluma khona.
- Ufuna ukuthi ikhiye le-free-tier esekuphumeleni inikeze indlela kuyo elandelayo ngaphandle kokwaphula isikhathi.
- Ufuna ukuthi okhiye abafanayo banikeze amandla kuma-bots / ama-IDE / amaskripthi amaningi ku-LAN efanayo ngaphandle kokukopisha izithombe.
- Ufuna ideshibhodi, hhayi okuhlukahlukile kwendawo, ukuhlela okhiye be-API.
- Ufuna ukukhetha kokuqala-okwasendaweni (Ollama / LM Studio) lapho imikhawulo yefu iphela.

## Ukuqala okusheshayo

### Faka (Linux / macOS)

```bash
curl -fsSL https://raw.githubusercontent.com/sookmook/wall-vault/main/install.sh | sh
```

Noma landa i-binary esiyakhiwe ngokuqondile:

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

### Faka (Windows)

```powershell
Invoke-WebRequest -Uri `
  "https://github.com/sookmook/wall-vault/releases/latest/download/wall-vault-windows-amd64.exe" `
  -OutFile wall-vault.exe
```

### Ukusebenzisa kokuqala

```bash
wall-vault setup    # i-wizard ehlanganyelayo — ikhetha ichweba, izinkonzo, admin token, iphasiwedi enkulu
wall-vault start    # iqala kokubili i-vault ne-proxy
```

Vula `http://localhost:56243` (noma `https://...` uma i-TLS isivunyelwe — bheka ngezansi) kusiphequluli. Ideshibhodi inxusa i-admin token ephrintwe yi-`setup`. Kusukela lapho ungangeza okhiye be-API, ubhalise amakhasimende, futhi ushintshe amamodeli ngaphandle kokuqalisa kabusha.

---

## TLS (kuphakanyiswa)

Ngokuzenzakalelayo `wall-vault setup` ibhala ukucushwa ngaphandle kwe-TLS, ngakho-ke abalalele bobabili baphendula nge-HTTP elula. Izibonelo ze-URL kule-README zisebenzisa `https://localhost:56244` ngoba ama-agent amaningi (OpenClaw, Claude Code, Cursor) afuna i-endpoint eyodwa enobuphongolo be-TLS engeke iphukise uma ngemuva uyisa i-proxy kweminye i-host. Ukufanisa lezo zibonelo, vumela i-TLS kanye ngokusebenzisa i-CA yangaphakathi efakiwe:

```bash
# 1. Yakha i-CA yangaphakathi yewall-vault (kanye nje, ihlala ku-~/.wall-vault/ca.{crt,key})
wall-vault cert init

# 2. Khipha isitifiketi se-host kule mishini
#    Ama-SAN ahlanganisa i-hostname, localhost, 127.0.0.1, nanoma yi-IP ye-LAN etholiwe
wall-vault cert issue $(hostname)

# 3. Themba i-CA ku-keychain ye-OS yendawo
wall-vault cert install-trust

# 4. Shintshela abalaleli ku-TLS
export WV_PROXY_TLS_ENABLED=1
export WV_PROXY_TLS_CERT="$HOME/.wall-vault/$(hostname).crt"
export WV_PROXY_TLS_KEY="$HOME/.wall-vault/$(hostname).key"
export WV_VAULT_TLS_ENABLED=1
export WV_VAULT_TLS_CERT="$HOME/.wall-vault/$(hostname).crt"
export WV_VAULT_TLS_KEY="$HOME/.wall-vault/$(hostname).key"
wall-vault start
```

Kwenye imishini ku-LAN yakho: kopisha `~/.wall-vault/ca.crt` lapho bese usebenzisa `wall-vault cert install-trust --ca <path>` lapho. Lapho i-CA isithenjwa yonke indawo, wonke umshini onethiwekhi ungafinyelela i-proxy nge-`https://<host>:56244` ngaphandle kwezixwayiso zesitifiketi.

Uma ungathanda ukuhlala ku-HTTP elula, shiya ukucushwa njengoba kunjalo bese ushintsha `https://` nge-`http://` kuziqephu zekhasimende ngezansi. Womabili amasakhema asebenza; umehluko uthi ichweba liphi eliphendula ukuxhawulana kwe-TLS.

**I-Loopback fallback.** Amakhasimende ahosti efanayo angakwazi ukuhlonipha i-CA ye-wall-vault (ikakhulukazi i-Node runtime efakwe ne-OpenClaw, ebhala kabusha i-`NODE_EXTRA_CA_CERTS` lapho i-spawning) afinyelela i-proxy ngomngani we-HTTP elula we-loopback-kuphela ku-`127.0.0.1:56245`. wall-vault iyivumela ngokuzenzakalelayo lapho i-TLS ivuliwe.

---

## Ukuxhuma amakhasimende

Khombisa noma yiluphi i-AI client ku-`https://<host>:56244` (noma `http://...` uma i-TLS icishiwe). I-Proxy iphendula izinhlobo ezine:

| Ifomethi | Indlela | Izibonelo zamakhasimende |
|--------|------|-----------------|
| Gemini | `/google/v1beta/models/<model>:generateContent` | OpenClaw, Gemini CLI, Antigravity |
| Anthropic | `/v1/messages` | Claude Code, Anthropic SDKs |
| OpenAI | `/v1/chat/completions` | Cursor, Continue, amaskripthi athile, izinhlelo eziningi ze-LLM |
| Ollama-native | `/api/chat` | Amakhasimende e-Ollama adlulayo |

### Claude Code

```bash
export ANTHROPIC_BASE_URL=https://localhost:56244
export ANTHROPIC_API_KEY=<your-vault-client-token>
claude
```

Lapho ama-credit e-Anthropic upstream ephela, ukuthumela kuwela kunoma yibaphi abahlinzeki obeke ku-`fallback_services` kuleli khasimende. Ukuze ukhethe ukubuyela kwangaphandle kwe-Claude ngokucacile:

```yaml
proxy:
  anthropic_fallback_model: "claude-haiku-4-5-20251001"
```

(Okuzenzakalelayo okungenalutho kwenza ukuthumela kubuyele iphutha ukuze ukungahambisani kwendlela kuvele ngokushesha.)

### Cursor / Continue

Ku-Cursor **Settings → AI → OpenAI API**:

```
Base URL:  https://localhost:56244
API Key:   <your-vault-client-token>
Model:     gemini-2.5-flash    # noma yiliphi imodeli i-wall-vault eyaziyo
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

OpenClaw uhlaka lwe-agent ye-TUI okwakhelwe wall-vault ekuqaleni ukunikeza inkonzo. I-modal ye-**Add Agent** yedeshibhodi isetha uhlobo lwe-agent ku-`openclaw` (noma `nanoclaw`); bese i-wall-vault ibhala `~/.openclaw/openclaw.json` ngokuqondile, kuhlanganise nama-URL abahlinzeki, vault token, namabhayisikobho amamodeli:

```bash
VAULT_CLIENT_ID=my-bot \
VAULT_URL=http://<vault-host>:56243 \
VAULT_TOKEN=<your-client-token> \
wall-vault proxy
```

### curl / amaskripthi

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

## Ukucushwa

`wall-vault setup` ibhala noma `./wall-vault.yaml` noma `~/.wall-vault/config.yaml`. Hlela ngesandla amakhomba i-wizard engawabuzi.

```yaml
mode: standalone     # standalone | distributed
lang: en             # en | ko | zh | ja | es | fr | de | pt | ar | hi | id | th | sw | ha | ne | mn | zu
theme: light         # light | dark | cherry | ocean | gold | autumn | winter

proxy:
  port: 56244
  host: ""           # okuzenzakalelayo: 127.0.0.1 standalone, 0.0.0.0 distributed
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
  plain_port: 56245              # umngani we-HTTP we-loopback-kuphela uma i-TLS ivuliwe
  ollama_keep_alive: "30m"       # "-1" ungalokothi ulayishe phansi, "0" layisha phansi ngokushesha
  ollama_num_ctx: 8192
  oai_stream_forward: false      # opt-in ukudlula kwangempela kwe-backend SSE
  anthropic_fallback_model: ""   # opt-in ukubhalwa kabusha okungeyona i-Claude ku-anthropic dispatch

vault:
  port: 56243
  host: ""
  admin_token: ""
  admin_ip_whitelist: []
  master_password: ""            # iphasiwedi yokufihla ikey ye-AES-GCM
  data_dir: ~/.wall-vault/data
  services_dir: ~/.wall-vault/services
  tls:
    enabled: false
    cert_file: ""
    key_file: ""
  bootstrap_port: 56247          # umlaleli we-HTTP olula okhonza kuphela i-ca.crt

doctor:
  interval: 5m
  auto_fix: true
  log_file: /tmp/wall-vault-doctor.log

hooks:
  on_model_change: ""            # umyalo we-shell (env: SERVICE, MODEL)
  on_key_exhausted: ""
  on_service_down: ""
  on_doctor_fix: ""
  openclaw_socket: ""
```

### Okuhlukahluke kwendawo

Wonke ukugcwala kwe-YAML kunokuhlukile kwe-env okunqoba ifayela. Okujwayelekile:

| Okuhlukile | Incazelo |
|----------|-------------|
| `WV_LANG`, `WV_THEME` | Ulimi netimu |
| `WV_PROXY_PORT`, `WV_PROXY_HOST` | Ikheli lokulalela leProxy |
| `WV_VAULT_PORT`, `WV_VAULT_HOST` | Ikheli lokulalela leVault |
| `WV_VAULT_URL`, `WV_VAULT_TOKEN` | Amaphuzu omugqa we-Distributed |
| `WV_ADMIN_TOKEN`, `WV_MASTER_PASS` | Imininingwane ye-Vault |
| `WV_KEY_GOOGLE`, `WV_KEY_OPENROUTER`, `WV_KEY_ANTHROPIC`, `WV_KEY_OPENAI` | Okhiye be-API (abahlukaniswe ngekhefu uma kuningi) |
| `WV_PROXY_TLS_ENABLED`, `WV_PROXY_TLS_CERT`, `WV_PROXY_TLS_KEY` | Proxy TLS |
| `WV_VAULT_TLS_ENABLED`, `WV_VAULT_TLS_CERT`, `WV_VAULT_TLS_KEY` | Vault TLS |
| `WV_PROXY_PLAIN_PORT` | Umngani we-Loopback HTTP (`0` ukuvimba) |
| `WV_VAULT_BOOTSTRAP_PORT` | Umlaleli we-CA bootstrap (`0` ukuvimba) |
| `WV_OLLAMA_URL`, `WV_OLLAMA_KEEP_ALIVE`, `WV_OLLAMA_NUM_CTX` | Ukushuna kwe-Ollama |
| `WV_LMSTUDIO_URL`, `WV_VLLM_URL`, `WV_LLAMACPP_URL` | Ukungavumelani kwe-backend wendawo |
| `WV_TOKEN_SENTINEL_FALLBACK` | Ukufakelwa kwe-Loopback "proxy-managed" sentinel |
| `WV_OAI_STREAM_FORWARD` | Ukudlula kwe-SSE kwe-backend wangempela ohambisana ne-OpenAI |
| `WV_ANTHROPIC_FALLBACK_MODEL` | Opt-in ukubhalwa kabusha okungeyona i-Claude ku-anthropic |

---

## Amamodi

### Standalone (okuzenzakalelayo)

I-Vault ne-proxy ziqhuba enqubweni efanayo. Kungcono ku-host eyodwa enqumela kokubili okhiye ne-agent. Ilalela kuphela ku-loopback ngokuzenzakalelayo.

```bash
wall-vault start    # iqhuba kokubili
```

### Distributed

I-vault iqhuba ku-host eyodwa (**i-vault host**) futhi igcina bonke okhiye; ama-proxy amaningi kwamanye ama-host ngalinye lifakaza ngephuzu lwekhasimende ngalinye. Kuwusizo lapho izimishini eziningana zidinga okhiye abafanayo ngaphandle kokubakopisha.

**I-Vault host:**

```bash
WV_VAULT_HOST=0.0.0.0 wall-vault vault
```

**Yonke i-host yeproxy:**

```bash
WV_VAULT_URL=http://<vault-host>:56243 \
WV_VAULT_TOKEN=<this-client-token> \
WV_PROXY_HOST=0.0.0.0 \
wall-vault proxy
```

I-modal ye-**Add Client** yedeshibhodi yenza i-token, ibhalise uhlobo lwe-agent, futhi i-proxy ithatha ukucushwa kwayo nge-SSE ngaphandle kokuqalisa kabusha.

---

## Plugin yaml (i-backend yokufaka)

Noma yiyiphi i-backend ehambisana ne-OpenAI ingangezwa njenge-yaml ngaphansi kwe-`~/.wall-vault/services/`. wall-vault iyithatha lapho iqala, iyibhalise njengomsebenzi ongahanjiswe, futhi ukuthumela + isethi yokuthola i-OAI-compat + ibhuloho le-Gemini-stream konke kuyibona ngaphandle koshintsho lwekhodi.

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
inline_no_think_for_qwen3: false   # khetha uma i-backend yakho ihlubula imaka
```

Ukuhleleka kwe-Hub (i-wall-vault eyodwa imi phambi kwenye) kusekelwa nge-`tls_internal_ca: true`, `auth.type: bearer`, ne-`preserve_model_id: true`.

---

## Yakha kusukela emthonjeni

```bash
git clone https://github.com/sookmook/wall-vault
cd wall-vault
go build -o wall-vault .
./wall-vault setup
./wall-vault start
```

Yakha i-cross-compile kuyisethi yonke esekelwa:

```bash
make build-all   # linux-amd64, linux-arm64, darwin-amd64, darwin-arm64, windows-amd64
```

Izinguqulo zilandela `v{major}.{minor}.{patch}.{YYYYMMDD}.{HHmmss}`; `BASE_VERSION` ku-Makefile isetha isiqalo.

### Ukuhlelwa kwephrojekthi

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

## Amabhuku

- [Imanuwali yomsebenzisi](docs/MANUAL.en.md) — ukufaka, ideshibhodi, ama-agents, ukuxazulula izinkinga
- [Inkomba ye-API](docs/API.en.md) — wonke umnyombo onezimo zesicelo/zempendulo
- [CHANGELOG](CHANGELOG.md)

---

## Ubuchwepheshe

- Go 1.25, i-binary eyodwa engaguquki
- [templ](https://templ.guide) yedeshibhodi ehlanzwe ngeseva, [HTMX](https://htmx.org) yokuvuselelwa kwengxenye
- AES-GCM (ikey etholakala nge-PBKDF2) yokufihlwa kwekey ngenkathi iphumule
- Server-Sent Events yokuvumelaniswa kokucushwa okuphilayo phakathi kwe-vault nama-proxy
- I-CA yangaphakathi ezisayinele ngokwayo + izitifiketi ngeyihosti (akudingeki i-DNS yomphakathi / Let's Encrypt)

## Ilayisense

GPL-3.0. Bheka [LICENSE](LICENSE).

## Ukufaka isandla

Ama-pull request ayemukelwa. Bheka [CONTRIBUTING.md](CONTRIBUTING.md). Kwizinguquko ezinkulu, sicela uvule i-issue kuqala ukuze uxoxisane ngedizayini.
