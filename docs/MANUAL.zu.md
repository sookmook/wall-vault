# wall-vault User Manual

[English](MANUAL.md) · [한국어](MANUAL.ko.md) · [中文](MANUAL.zh.md) · [日本語](MANUAL.ja.md) · [Español](MANUAL.es.md) · [Français](MANUAL.fr.md) · [Deutsch](MANUAL.de.md) · [Português](MANUAL.pt.md) · [العربية](MANUAL.ar.md) · [हिन्दी](MANUAL.hi.md) · [Bahasa Indonesia](MANUAL.id.md) · [ภาษาไทย](MANUAL.th.md) · [Kiswahili](MANUAL.sw.md) · [Hausa](MANUAL.ha.md) · [नेपाली](MANUAL.ne.md) · [Монгол](MANUAL.mn.md) · isiZulu

Le ncwadi-mhlahlandlela ihlanganisa ukufaka, ukulungiselela, kanye nokusebenzisa i-wall-vault. Ngomvunge oqondile bheka i-[README](../README.md). Ngemininingwane ye-HTTP API bheka i-[API reference](API.md).

## Okuqukethwe

1. [Lokho i-wall-vault ekwenzayo](#lokho-i-wall-vault-ekwenzayo)
2. [Ukufaka](#ukufaka)
3. [Ukusebenzisa kokuqala nge-setup wizard](#ukusebenzisa-kokuqala-nge-setup-wizard)
4. [Ukunika amandla i-TLS](#ukunika-amandla-i-tls)
5. [Ukubhalisa ama-API key](#ukubhalisa-ama-api-key)
6. [Ukuxhuma ama-agent](#ukuxhuma-ama-agent)
7. [I-Dashboard](#i-dashboard)
8. [Imodi ye-Distributed](#imodi-ye-distributed)
9. [Ukuqala ngokuzenzakalelayo](#ukuqala-ngokuzenzakalelayo)
10. [Ama-Plugin yamls](#ama-plugin-yamls)
11. [Doctor](#doctor)
12. [Hooks](#hooks)
13. [Okuguquguqukayo kwendawo](#okuguquguqukayo-kwendawo)
14. [Ukuxazulula izinkinga](#ukuxazulula-izinkinga)

---

## Lokho i-wall-vault ekwenzayo

I-wall-vault iyi-Go binary eyodwa ehlanganisa izinsizakalo ezimbili ezisebenzisana:

- **I-vault** igcina ama-API key e-encrypted lapho ephumule (i-AES-GCM ngephasiwedi enkulu), ilandelela ukusetshenziswa kanye ne-cooldown nge-key ngalinye, isakaza izinguquko nge-Server-Sent Events (SSE), futhi inikeze i-web dashboard ku-`:56243` kubasebenzisi abangabantu.
- **I-proxy** iveza ama-endpoint ahambisanayo no-Gemini, Anthropic, OpenAI, kanye no-Ollama-native ku-`:56244`. Yiliphi i-AI client ekhomba i-proxy isebenzisa ama-key avault — amaclient awakuboni neze. Lapho upstream eyodwa ihluleka, i-dispatch iwela ku-provider olandelayo ngokulandelana.

Lokhu kuyasiza lapho:

- Une-key zama-provider amaningi futhi ufuna i-URL eyodwa lapho i-agent ikhuluma khona.
- Ufuna i-key ye-free-tier ekwi-cooldown ihambe ngaphandle kokuphula i-session.
- Ufuna ama-key afanayo aqhube ama-bot, ama-IDE, noma ama-script amaningi ku-LAN efanayo ngaphandle kokukopisha ama-credentials.
- Ufuna i-dashboard, hhayi okuguquguqukayo kwendawo, ukuze uhlele ama-key futhi ushintshe ama-model.
- Ufuna i-fallback yendawo (Ollama, LM Studio, vLLM) lapho imikhawulo ye-cloud iphela.

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

## Ukufaka

### Linux / macOS one-liner

```bash
curl -fsSL https://raw.githubusercontent.com/sookmook/wall-vault/main/install.sh | sh
```

I-script ithola ngokuzenzakalelayo i-OS ne-architecture, idawunilodi i-binary efanele ku-`~/.local/bin/wall-vault`, futhi iyenze ikwazi ukusebenza. Uma `~/.local/bin` ingekho ku-`PATH` yakho, yengeze:

```bash
export PATH="$HOME/.local/bin:$PATH"
```

### Idawunilodi ngesandla

Ama-binary akhiwe kusengaphambili ashicilelwe ku-release ngalinye ku-`https://github.com/sookmook/wall-vault/releases`.

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

### Yakha kusukela kumthombo

Idinga i-Go 1.25 noma entsha kakhulu.

```bash
git clone https://github.com/sookmook/wall-vault
cd wall-vault
go build -o wall-vault .
```

`make build-all` i-cross-compile kuzo zonke izinhlelo ezinhlanu ezisekelwayo. Ama-binary afika ku-`bin/`.

---

## Ukusebenzisa kokuqala nge-setup wizard

```bash
wall-vault setup
```

I-wizard ikubuza, ngokulandelana:

1. **Ulimi** — ikhetha okukodwa kokungu-17 wama-locale e-UI. Itholakala ngokuzenzakalelayo kusukela ku-`$LANG`; i-wizard inikeza uhlu noma kunjalo.
2. **Itimu** — `light` (okuzenzakalelayo), `dark`, `cherry`, `ocean`, `gold`, `autumn`, `winter`. Kungumkhangiso kuphela.
3. **Imodi** — `standalone` (i-host eyodwa, okuzenzakalelayo) noma `distributed` (i-vault ku-host eyodwa, ama-proxy kwabanye).
4. **Igama le-bot** — i-`client_id` slug yamahhala. I-vault iyisebenzisa ukukhawula i-config nge-client ngayinye (ama-model overrides, ama-fallback chains).
5. **I-Proxy port** — okuzenzakalelayo `56244`.
6. **I-Vault port** — okuzenzakalelayo `56243` (i-standalone kuphela).
7. **Ukukhetha izinsizakalo** — i-y/N ngalinye lalokhu: Google Gemini, OpenRouter, Anthropic, OpenAI, Ollama, LM Studio, vLLM. Ukukhetha okuningi kuhle; ngalinye libhala uphawu lwalo lwe-env-var ekupheleni.
8. **I-Tool filter** — `strip_all` (okuzenzakalelayo; ivimba zonke izincazelo ze-tool ezingenayo ngokuphepha) noma `passthrough` (idedela noma yiyiphi i-tool idlule).
9. **I-Admin token** — shiya ize ukuze izenzelekelwe ngokuzenzakalelayo. I-dashboard idinga le-token ukuze ungene.
10. **Iphasiwedi enkulu** — shiya ize ngokungekho ku-encryption (AKUTUSWA); setha ivelu ukuze i-AES-GCM i-encrypt i-key store lapho iphumule.
11. **Indlela yokulondoloza** — okuzenzakalelayo ku-`wall-vault.yaml` kuhla yamanje. I-loader iphinde ibheke ku-`~/.wall-vault/config.yaml`.

Ngemva kokulondoloza, i-wizard isebenzisa i-`doctor.FixTrust` ukuze noma yiyiphi i-agent efakwe endaweni (OpenClaw, Claude Code, Cline) ithole i-wall-vault internal CA yengezwe ku-trust store yayo ngokuzenzakalelayo. Uma kungekho i-agent enjalo efakiwe, isinyathelo siphrinta `SKIP` futhi asibhali lutho.

Bese qala i-binary:

```bash
wall-vault start
```

I-`start` isebenzisa zombili i-vault kanye ne-proxy enqubweni eyodwa (imodi ye-standalone). Kwimodi ye-distributed sebenzisa i-`wall-vault vault` ku-vault host kanye ne-`wall-vault proxy` ku-proxy host ngayinye.

Vula i-`http://localhost:56243` kusiphequluli. Ngena ne-admin token ekuthi i-wizard ihlolisise.

---

## Ukunika amandla i-TLS

Okuzenzakalelayo kwe-wizard kushiya bobabili ababukeli ku-HTTP elula. Iningi lama-agent (OpenClaw, Claude Code, Cursor) lisebenza kangcono mayelana ne-HTTPS endpoint eyodwa, ngakho i-TLS ituswa kunoma yikuphi ukufakwa okudlulele kumshini wendawo kuphela.

I-wall-vault iza ne-CA yayo yangaphakathi ngakho awudingi igama le-DNS lomphakathi noma i-Let's Encrypt.

```bash
# 1. Dala i-CA yangaphakathi — ibhalwe ku- ~/.wall-vault/ca.{crt,key}.
#    I-CA inhle iminyaka eyi-10 ngokuzenzakalelayo; chitha nge --ca-years.
wall-vault cert init

# 2. Khipha i-host certificate. I-Subject Alternative Names ihlanganisa ngokuzenzakalelayo:
#       hostname, "localhost", "127.0.0.1", kanye nanoma yiluphi i-LAN IP elingelona elwe-loopback elitholakele.
#    Chitha umqondisi we-issuer nge --dir, ubuvithi nge --host-years.
wall-vault cert issue $(hostname)

# 3. Thembela ku-CA ku-OS keychain yalo mshini.
#    Linux: ibhala ku-/etc/ssl/certs/ ngokwe-update-ca-certificates (idinga i-sudo).
#    macOS: yengeza ku-System keychain ngokwe-security add-trusted-cert (idinga i-sudo).
#    Windows: ingenisa ku-CurrentUser\Root ngokwe-certutil (akudingeki i-admin).
wall-vault cert install-trust

# 4. Nika amandla i-TLS kubo bobabili ababukeli.
export WV_PROXY_TLS_ENABLED=1
export WV_PROXY_TLS_CERT="$HOME/.wall-vault/$(hostname).crt"
export WV_PROXY_TLS_KEY="$HOME/.wall-vault/$(hostname).key"
export WV_VAULT_TLS_ENABLED=1
export WV_VAULT_TLS_CERT="$HOME/.wall-vault/$(hostname).crt"
export WV_VAULT_TLS_KEY="$HOME/.wall-vault/$(hostname).key"

wall-vault start
```

Ukunwebela ukwethemba kwabanye omishini be-LAN, kopisha i-`~/.wall-vault/ca.crt` futhi usebenzise i-`wall-vault cert install-trust --ca <path>` kungayinye. I-vault iphinde iveze i-`ca.crt` ngokuncane kwe-plain-HTTP listener ku-`:56247` (**i-bootstrap port**) ngecala lika-catch-22 lapho i-client entsha idinga i-CA ukukhuluma i-HTTPS.

### Loopback HTTP companion

Amanye ama-agent — ikakhulukazi i-Node runtime ehlanganiswe ne-OpenClaw — ibhala kabusha i-`NODE_EXTRA_CA_CERTS` ku-process spawn, ilahla noma yiluphi uphawu lwe-CA olunikezelwe ngumsebenzisi. Awakwazi ukuhlonipha i-wall-vault CA esuka phakathi kwe-daemon, ngisho nangemva kwe-`cert install-trust`. I-wall-vault isebenza nakukho ngokubopha **i-loopback-only plain-HTTP listener** eyengeziwe ku-`127.0.0.1:56245` noma kunini i-TLS inikwe amandla. Ama-Same-host clients afinyelela i-proxy ngakuleyo port ngaphandle kwe-TLS neze; ama-LAN clients aqhubeka esebenzisa i-TLS listener.

Khubaza nge-`WV_PROXY_PLAIN_PORT=0` uma ungayidingi.

### `wall-vault cert list`

Ikhombisa wonke i-cert ngaphansi kwe-`~/.wall-vault/` nge-subject, ividothi yokuvumelana, kanye nama-SANs.

```
$ wall-vault cert list
ca.crt          subject=wall-vault internal CA   not-after=2036-05-05
hostname.crt    subject=hostname                 not-after=2031-05-05   SAN=hostname,localhost,127.0.0.1,192.168.…
```

---

## Ukubhalisa ama-API key

Izindlela ezimbili: i-dashboard, noma okuguquguqukayo kwendawo.

### I-Dashboard (etuswayo)

1. Ngena ku-`https://localhost:56243` ne-admin token.
2. Chofoza i-**+ API key** ku-keys card.
3. Khetha insizakalo (Google, OpenRouter, Anthropic, OpenAI, …).
4. Namathisela i-key. Londoloza.

Ama-key amaningi ngenkonzo kuhle; i-proxy yenza i-round-robin phakathi kwawo futhi yeqe lawo aqala ku-cooldown nge-key.

### Okuguquguqukayo kwendawo (i-bootstrap eyodwa)

```bash
export WV_KEY_GOOGLE="AIzaSyA1...,AIzaSyB2...,AIzaSyC3..."   # comma-separated
export WV_KEY_OPENROUTER="sk-or-v1-…"
export WV_KEY_ANTHROPIC="sk-ant-…"
export WV_KEY_OPENAI="sk-…"
wall-vault start
```

Ama-key anikezelwe ngale ndlela abhalwa ku-encrypted store ekuqaleni okokuqala. Ukuqala okulandelayo kuyawafunda ku-disk; ungakhipha ama-env vars ngemva kokuqala kokuqala.

### Ama-Cooldowns kanye nokujikeleza

Ikholi ngalinye eliphumelelayo lengeza i-`usage_count` ye-key futhi livuselele i-`last_used`. Ku-HTTP 429 / 402 / 403, i-proxy ifaka i-key ku-**cooldown** (okuzenzakalelayo: imizuzu eyi-60 ku-429, amahora ayi-24 ku-402, amahora ayi-12 ku-403). I-dispatch elandelayo ikhetha i-key ehlukile yalowo msebenzi. Lapho wonke ama-key omsebenzi ekwi-cooldown, i-proxy iyaqhubeka iyeke loyo msebenzi ngokushesha futhi izame i-provider olandelayo ku-fallback chain.

Ama-cooldowns ayabonakala nge-key ku-dashboard ngokubala kukheli.

---

## Ukuxhuma ama-agent

### OpenClaw

I-OpenClaw yi-client okuqondiswe kuyo kuqala. Sebenzisa i-modal ye-**+ Add agent** ye-dashboard:

- Setha i-**Agent type** ku-`openclaw` noma `nanoclaw`.
- Setha i-**Work directory** — ye-OpenClaw lokhu kugcwala ngokuzenzakalelayo njenge `~/.openclaw`.
- Khetha **i-preferred service** futhi mhlawumbe **i-model override**.
- Chofoza i-**Apply**. I-wall-vault ibhala i-`~/.openclaw/openclaw.json` ngqo (ama-provider URLs, i-vault token, ama-model entries).

Lapho ushintsha i-model kusuka ku-dashboard, i-OpenClaw ithatha ushintsho ngokwe-SSE phakathi kwemizuzwana eyi-1–3 — akukho ukuqala kabusha.

### Claude Code

```bash
export ANTHROPIC_BASE_URL=https://localhost:56244
export ANTHROPIC_API_KEY=<your-vault-client-token>
claude
```

Lapho ama-credit e-upstream Anthropic ephela, i-dispatch iwela kunoma yiziphi izinsizakalo ezisohlwini lwe-`fallback_services` ye-client. Ngokuzenzakalelayo, i-model id elingewona elika-Claude elithunyelwe ku-anthropic dispatch libuyisela iphutha ukuze ukuxhumeka kabi kuvele ngokushesha. Khetha ukuvuma ku-rewrite ezenzakalelayo:

```yaml
proxy:
  anthropic_fallback_model: "claude-haiku-4-5-20251001"
```

### Cursor

Ku-Cursor **Settings → AI → OpenAI API**:

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

I-endpoint efanayo yamukela ukugobhoza (`"stream": true`) lapho i-`proxy.oai_stream_forward: true` isethiwe.

---

## I-Dashboard

`https://localhost:56243`. Amakhadi amahlanu ku-home grid:

- **Keys** — i-API key ngalinye, ehlanganiswe ngokomsebenzi. Yengeza, hlela, susa; bona ukusetshenziswa kanye ne-cooldown.
- **Services** — Google / OpenRouter / Anthropic / OpenAI / Ollama / LM Studio / vLLM / llama.cpp, kanye nanoma yiluphi i-plugin yaml ku-`~/.wall-vault/services/`. Setha nge-msebenzi `default_model`, `allowed_models`, base URL, ushintsho lokucabanga.
- **Clients (agents)** — i-client ngayinye ebhalisiwe (OpenClaw bot, Claude Code session, Cursor instance, …). Yabela insizakalo etuswayo, i-model override, i-fallback chain.
- **Proxies** — i-proxy ngayinye etuswile kule i-vault. Isimo esiphilayo (online/offline), okugcina kubonwe, i-model yamanje.
- **Settings** — i-admin token, ukujikeleza kwephasiwedi enkulu, itimu, ulimi.

Ikhadi ngalinye line-edit slideover (uhlangothi lwesokudla). Ukuchofoza ngaphandle noma i-`Esc` kuvala. Izinguquko zisukezwa kuwo wonke ama-proxy axhumene nge-SSE phakathi kwemizuzwana.

**I-footer** iphethe inkomba ye-SSE (luhlaza = ixhunyiwe, i-orange = ixhuma futhi, mpunga = idluliwe) kanye nenguqulo ye-build ephilayo.

---

## Imodi ye-Distributed

Lapho unemishini eminingana okufanele yonke idinge ama-key afanayo, sebenzisa i-vault ku-host eyodwa kanye nama-proxy kuwo wonke amanye.

### Vault host

```bash
WV_VAULT_HOST=0.0.0.0 \
WV_ADMIN_TOKEN=<admin> \
WV_MASTER_PASS=<master> \
wall-vault vault
```

I-dashboard manje iyatholakala ku-`https://<vault-host>:56243`. Yengeza i-agent ye-proxy ngayinye yesibuko ku-card ye-**Clients**; ngalinye lakha i-`vault_token` eyingqayizivele.

### Proxy hosts

```bash
WV_VAULT_URL=http://<vault-host>:56243 \
WV_VAULT_TOKEN=<that-client-token> \
WV_PROXY_HOST=0.0.0.0 \
wall-vault proxy
```

I-proxy ifakaza ku-vault, ivula i-SSE stream, futhi isebenzisa noma yiluphi i-config etholayo (insizakalo etuswayo, i-model override, i-fallback chain). Ukuhlela kwe-vault okulandelayo kufika phakathi kwemizuzwana ngaphandle kokuqala kabusha.

Ngokufakwa okudlula i-LAN, nika amandla i-TLS ku-vault host (`WV_VAULT_TLS_ENABLED=1` + ama-cert/key env vars) futhi sebenzisa i-proxy host ngayinye ngesinyathelo esifanayo se-`wall-vault cert install-trust` ukuze izinkomba ze-HTTPS ze-proxy ku-vault zithenjwe.

---

## Ukuqala ngokuzenzakalelayo

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

Yokwasi ku-host efanayo, bhala i-`wall-vault-vault.service` ehambisanayo. Kwimodi ye-standalone, i-unit eyodwa ebiza i-`wall-vault start` yanele.

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

Sebenzisa i-`nssm` ukusongela i-`wall-vault.exe start` njenge-Windows service, noma i-entry ye-`schtasks` esebenza ku-user logon.

---

## Ama-Plugin yamls

Noma yiluphi i-OpenAI-compatible backend lingangezwa ngaphandle kokushintsha ikhodi ngokulahla i-yaml ngaphansi kwe-`~/.wall-vault/services/`. I-wall-vault iyilayisha ekuqaleni futhi ibhalisa insizakalo ye-dispatch, i-OAI-compat detection set, kanye ne-Gemini-stream bridge.

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

Iqembu elihlanganisiwe ku-`configs/services/` (lmstudio, vllm, llamacpp, tgwui, localai, jan, koboldcpp, tabbyapi, mlx-server, litellm-proxy, ollama, google, openrouter) lifika likhutshazwe ngokuzenzakalelayo. Kopisha leyo oyifunayo ku-`~/.wall-vault/services/`, setha i-`enabled: true`, qala kabusha.

---

## Doctor

I-`wall-vault doctor` isebenzisa i-one-shot health probe kuyo yonke ifakwayo:

```
✓ vault listener  (https://localhost:56243)
✓ proxy listener  (https://localhost:56244)
✓ master password set
⚠ Google: 2 keys, all on cooldown
✓ Anthropic: 1 key healthy
✗ Ollama: not reachable at http://localhost:11434
```

Ulayini ngamunye ungukukodwa kwalokhu:

- `✓` — kunempilo
- `⚠` — kuncipiswe kodwa kuyasebenza (i-key eyodwa epholile, i-quota ephansi, njll.)
- `✗` — kuphukile
- `SKIP` — akulinganiselwa / akufaneleki kuleyo host

Imodi yesibili ye-daemon isebenzisa i-probe efanayo ku-`doctor.interval` ngayinye (okuzenzakalelayo imizuzu eyi-5) futhi ibhala imiphumela ku-`doctor.log_file` (okuzenzakalelayo `/tmp/wall-vault-doctor.log`). Lapho i-`doctor.auto_fix` iyiqiniso, iphinde izame ukulungisa i-drift evamile (i-OpenClaw config endala, i-TLS trust elahlekile, izinsizakalo ezikwazi ukuqalwa kabusha).

Vusela i-one-shot kusuka ku-dashboard ngekhadi le-**Doctor** noma i-`wall-vault doctor`.

---

## Hooks

Sebenzisa i-shell command ezenzakalweni ze-key:

```yaml
hooks:
  on_model_change:   "logger 'wall-vault: $SERVICE/$MODEL'"
  on_key_exhausted:  "notify-send 'wall-vault' '$SERVICE keys all on cooldown'"
  on_service_down:   "/usr/local/bin/page-oncall.sh $SERVICE '$ERROR'"
  on_doctor_fix:     "echo \"$AGENT: $LEVEL $MSG\" >> ~/wall-vault.audit.log"
  openclaw_socket:   ""    # if set, OpenClaw TUI receives events over this Unix socket
```

I-hook ngayinye ithola okuguquguqukayo kwendawo okuthize kwemicimbi (`SERVICE`, `MODEL`, `ERROR`, `AGENT`, `LEVEL`, `MSG`). Ama-hook asebenza i-async ne-timeout yemizuzwana eyi-5 — i-proxy ayikaze ivimbe ku-hook elingenamandla.

---

## Okuguquguqukayo kwendawo

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

I-env var ngayinye, lapho ibekiwe, inqoba i-YAML file.

---

## Ukuxazulula izinkinga

### `connection refused` ku-`:56244`

Mhlawumbe i-proxy ayisebenzi noma iboshelwe ku-host ehlukile. Hlola:

```bash
ss -lnp | grep 56244
systemctl --user status wall-vault-proxy   # Linux
launchctl list | grep wall-vault           # macOS
```

Uma isebenza ku-port ehlukile, i-config yakho ine-`proxy.port` echithwe — hlola i-`~/.wall-vault/config.yaml`.

### `x509: certificate signed by unknown authority`

I-client ayithembi i-wall-vault internal CA. Sebenzisa i-`wall-vault cert install-trust` kumshini we-client. Yokwasi ama-agent ane-runtime engayilaleli i-OS trust store (isib. i-Node ene-`NODE_EXTRA_CA_CERTS` ehlukaniswe ngokomzimba), sebenzisa i-loopback HTTP companion ku-`127.0.0.1:56245` (i-host eyodwa kuphela) noma setha i-`WV_PROXY_TLS_ENABLED=0` ukuze uwele ku-HTTP elula.

### `token not registered with vault`

I-`Authorization: Bearer <token>` ye-client ayifani nanoma yiyiphi i-client ebhalisiwe. Qinisekisa i-token ngaphansi kwe-**Clients** ku-dashboard. Uma ukopishe i-token literal njenge-`proxy-managed`, `dummy`, noma `""` kusuka ku-config endala, yibuyisele nge-token yeqiniso ye-client.

### `Anthropic dispatch needs a Claude model id`

Ukuziphatha okuzenzakalelayo kusukela ku-v0.2.63: i-model id elingelona elika-Claude elithunyelwe ku-anthropic dispatch libuyisela iphutha. Mhlawumbe lungisa i-routing (ungathumeli `gemini-2.5-flash` ku-anthropic) noma khetha ukuvuma ku-rewrite ezenzakalelayo nge-`proxy.anthropic_fallback_model`.

### `unknown service: <id>`

I-Dispatch ibone i-service id okungekho i-plugin yaml efaka isicelo. Hlola:

```bash
ls ~/.wall-vault/services/        # any plugin yaml present?
cat ~/.wall-vault/services/<id>.yaml | grep enabled
```

Uma i-yaml ikhona kodwa i-`enabled: false`, yiphendule. Uma ingekho ngempela, kopisha kusuka ku-`configs/services/` ku-source tree.

### Impendulo engenalutho ku-reasoning model

I-`qwen3.6`, i-`deepseek-r1`, kanye nomdeni we-GPT-`o1` ngezinye izikhathi bakhipha kuphela i-`reasoning_content` futhi bashiye i-`content` ingenalutho. Kusukela ku-v0.2.63 i-wall-vault iwela kumbhalo wokucabanga ngokuzenzakalelayo — uma usalokhu ubona izimpendulo ezingenalutho, i-backend ayibuyiseli okunye okuye field. Hlola amalogi e-upstream.

Ye-LM Studio ne-qwen3 ngokukhethekile, setha i-`inline_no_think_for_qwen3: true` ku-plugin yaml ukuze ukucabanga kukhutshazwe inline. I-Built-in lmstudio.yaml kanye ne-ollama.yaml zenza kakade lokhu.

### I-Dashboard ikhombisa "all keys on cooldown" kodwa ngisanda kungeza eyodwa

I-key entsha inempilo kodwa indlela ye-dispatch ingahle ibe iseku-cooldown ye-key endala. Zama isicelo esisha — i-proxy yenza i-round-robins nge-call, futhi i-key enempilo izokhethwa olandelayo.

### I-Vault ayivuli nge-master password

Iphasiwedi engalungile. Akukho ukubuyiselwa — i-wall-vault ngamabomu ayithumeli i-backdoor. Uma ngempela ulahlekelwe iphasiwedi enkulu, indlela kuphela ukususa i-`~/.wall-vault/data/vault.json`, qala kabusha ngephasiwedi entsha, futhi ungeze ama-key futhi.

### Imikhawulo ye-Free-tier OpenRouter ifikile

Setha i-`proxy.services` ukuze ihlanganise i-`openrouter` futhi yengeze okungenani i-OpenRouter key eyodwa. I-proxy iyazenzakalisa ngokuziwela kusuka ku-model okhokhelwayo kuya kwe-`:free` variant yawo lapho indlela ekhokhelwayo ibuyisela i-402 / 429.

### `journalctl --user -u wall-vault-proxy` ingenalutho

Amalogi we-systemd `--user` aya ku-journal yomsebenzisi oyiqhubayo. Uma uqalile i-unit njenge-`root` noma nge-`sudo`, i-journal isenkambeni ye-system esikhundleni — zama i-`journalctl -u wall-vault-proxy` ngaphandle kwe-`--user`.

---

## Okuningi

- HTTP API reference — bheka [API.md](API.md)
- Source — `https://github.com/sookmook/wall-vault`
- Imibiko yamabug / izicelo zezici — GitHub Issues
- Umlando we-release — [CHANGELOG.md](../CHANGELOG.md)
