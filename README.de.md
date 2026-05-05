# wall-vault

> **API-Schlüssel-Tresor + KI-Proxy in einer einzigen Go-Binärdatei.**
> Speichert Schlüssel lokal mit AES-GCM, rotiert sie über Anbieter hinweg, fällt zurück, wenn einer ausfällt, und wird mit einem Echtzeit-Dashboard ausgeliefert.

[![License: GPL v3](https://img.shields.io/badge/License-GPLv3-blue.svg)](LICENSE)
[![Go](https://img.shields.io/badge/Go-1.25-00ADD8.svg)](go.mod)

[English](README.md) · [한국어](README.ko.md) · [中文](README.zh.md) · [日本語](README.ja.md) · [Español](README.es.md) · [Français](README.fr.md) · **Deutsch** · [Português](README.pt.md) · [العربية](README.ar.md) · [हिन्दी](README.hi.md) · [Bahasa Indonesia](README.id.md) · [ภาษาไทย](README.th.md) · [Kiswahili](README.sw.md) · [Hausa](README.ha.md) · [नेपाली](README.ne.md) · [Монгол](README.mn.md) · [isiZulu](README.zu.md)

---

## Was es ist

wall-vault sitzt zwischen einem KI-Agenten (OpenClaw, Claude Code, Cursor, Continue, Ihrem eigenen Skript) und den Cloud- oder lokalen KI-Anbietern, mit denen er spricht. Zwei Dinge in einer Binärdatei:

- **Vault** — speichert API-Schlüssel verschlüsselt im Ruhezustand (AES-GCM mit einem Master-Passwort), rotiert sie, verfolgt Nutzung und Cooldowns pro Schlüssel, sendet Änderungen über SSE und stellt ein Web-Dashboard auf `:56243` bereit.
- **Proxy** — stellt auf `:56244` Gemini-, Anthropic- und OpenAI-kompatible Endpunkte bereit, wählt einen Schlüssel aus dem Tresor, leitet an das konfigurierte Upstream weiter und fällt auf den nächsten Anbieter zurück, wenn einer fehlschlägt.

Er unterstützt vier Anfrageformen (Gemini `:generateContent`, Anthropic `/v1/messages`, OpenAI `/v1/chat/completions` und Ollama-nativ `/api/chat`) und fünf Upstream-Kategorien:

| Anbieter | Hinweise |
|----------|-------|
| Google Gemini | Native API; Schlüsselrotation pro Projekt |
| Anthropic | Nativer `/v1/messages`-Passthrough |
| OpenAI | Natives `/v1/chat/completions` |
| OpenRouter | 340+ Modelle, automatischer Fallback auf `:free`-Varianten |
| Ollama / LM Studio / vLLM / llama.cpp / Jan / KoboldCpp / TabbyAPI / mlx-server / LiteLLM | Lokale OpenAI-kompatible Backends; Drop-in über Plugin-yaml |

Ein neues OpenAI-kompatibles Backend hinzuzufügen ist eine yaml-Datei unter `~/.wall-vault/services/` — keine Codeänderung nötig.

## Warum Sie es vielleicht möchten

- Sie jonglieren mit drei oder vier KI-Diensten und möchten eine URL, mit der der Agent spricht.
- Sie möchten, dass ein Free-Tier-Schlüssel im Cooldown zur Seite tritt, ohne die Sitzung zu unterbrechen.
- Sie möchten, dass dieselben Schlüssel mehrere Bots / IDEs / Skripte im selben LAN versorgen, ohne Anmeldedaten zu kopieren.
- Sie wollen ein Dashboard, keine Umgebungsvariablen, zum Bearbeiten von API-Schlüsseln.
- Sie wollen eine Local-First-Option (Ollama / LM Studio), wenn Cloud-Limits ausgeschöpft sind.

## Schnellstart

### Installation (Linux / macOS)

```bash
curl -fsSL https://raw.githubusercontent.com/sookmook/wall-vault/main/install.sh | sh
```

Oder laden Sie eine vorgefertigte Binärdatei direkt herunter:

```bash
# Linux amd64
curl -L https://github.com/sookmook/wall-vault/releases/latest/download/wall-vault-linux-amd64 \
  -o wall-vault && chmod +x wall-vault

# macOS Apple Silicon
curl -L https://github.com/sookmook/wall-vault/releases/latest/download/wall-vault-darwin-arm64 \
  -o wall-vault && chmod +x wall-vault

# Linux arm64 (Raspberry Pi, ARM-Server)
curl -L https://github.com/sookmook/wall-vault/releases/latest/download/wall-vault-linux-arm64 \
  -o wall-vault && chmod +x wall-vault
```

### Installation (Windows)

```powershell
Invoke-WebRequest -Uri `
  "https://github.com/sookmook/wall-vault/releases/latest/download/wall-vault-windows-amd64.exe" `
  -OutFile wall-vault.exe
```

### Erster Start

```bash
wall-vault setup    # interaktiver Assistent — wählt Port, Dienste, Admin-Token, Master-Passwort
wall-vault start    # startet sowohl Vault als auch Proxy
```

Öffnen Sie `http://localhost:56243` (oder `https://...`, sobald TLS aktiviert ist — siehe unten) im Browser. Das Dashboard fragt nach dem von `setup` ausgegebenen Admin-Token. Von dort aus fügen Sie API-Schlüssel hinzu, registrieren Clients und wechseln Modelle ohne Neustart.

---

## TLS (empfohlen)

Standardmäßig schreibt `wall-vault setup` eine Konfiguration ohne TLS, sodass beide Listener mit einfachem HTTP antworten. Die Beispiel-URLs in dieser README verwenden `https://localhost:56244`, weil die meisten Agenten (OpenClaw, Claude Code, Cursor) einen einzigen TLS-vorgelagerten Endpunkt wollen, der nicht zerbricht, wenn Sie den Proxy später auf einen anderen Host verlagern. Um diesen Beispielen zu entsprechen, aktivieren Sie TLS einmalig mit der mitgelieferten internen CA:

```bash
# 1. Erstellen Sie die wall-vault interne CA (einmalig, liegt in ~/.wall-vault/ca.{crt,key})
wall-vault cert init

# 2. Stellen Sie ein Host-Zertifikat für DIESE Maschine aus
#    SANs umfassen Hostname, localhost, 127.0.0.1 und jede erkannte LAN-IP
wall-vault cert issue $(hostname)

# 3. Vertrauen Sie der CA im lokalen OS-Schlüsselbund
wall-vault cert install-trust

# 4. Schalten Sie die Listener auf TLS um
export WV_PROXY_TLS_ENABLED=1
export WV_PROXY_TLS_CERT="$HOME/.wall-vault/$(hostname).crt"
export WV_PROXY_TLS_KEY="$HOME/.wall-vault/$(hostname).key"
export WV_VAULT_TLS_ENABLED=1
export WV_VAULT_TLS_CERT="$HOME/.wall-vault/$(hostname).crt"
export WV_VAULT_TLS_KEY="$HOME/.wall-vault/$(hostname).key"
wall-vault start
```

Für eine andere Maschine in Ihrem LAN: Kopieren Sie `~/.wall-vault/ca.crt` hinüber und führen Sie dort `wall-vault cert install-trust --ca <pfad>` aus. Sobald die CA überall vertraut wird, kann jede Maschine im Netzwerk den Proxy über `https://<host>:56244` ohne Zertifikatswarnungen erreichen.

Wenn Sie lieber bei einfachem HTTP bleiben möchten, lassen Sie die Konfiguration so wie sie ist und ersetzen Sie `https://` durch `http://` in den Client-Snippets unten. Beide Schemata funktionieren; der Unterschied ist, welcher Port auf einen TLS-Handshake antwortet.

**Loopback-Fallback.** Clients auf demselben Host, die die wall-vault-CA nicht ehren können (insbesondere die mit OpenClaw gebündelte Node-Runtime, die `NODE_EXTRA_CA_CERTS` beim Spawnen überschreibt), erreichen den Proxy über einen reinen Loopback-Plain-HTTP-Begleiter auf `127.0.0.1:56245`. wall-vault aktiviert ihn automatisch, wenn TLS eingeschaltet ist.

---

## Clients verbinden

Richten Sie einen beliebigen KI-Client auf `https://<host>:56244` aus (oder `http://...`, wenn TLS aus ist). Der Proxy antwortet in vier Formen:

| Format | Pfad | Beispiel-Clients |
|--------|------|-----------------|
| Gemini | `/google/v1beta/models/<model>:generateContent` | OpenClaw, Gemini CLI, Antigravity |
| Anthropic | `/v1/messages` | Claude Code, Anthropic-SDKs |
| OpenAI | `/v1/chat/completions` | Cursor, Continue, eigene Skripte, die meisten LLM-Apps |
| Ollama-nativ | `/api/chat` | Durchgereichte Ollama-Clients |

### Claude Code

```bash
export ANTHROPIC_BASE_URL=https://localhost:56244
export ANTHROPIC_API_KEY=<ihr-vault-client-token>
claude
```

Wenn die Anthropic-Credits im Upstream aufgebraucht sind, fällt das Dispatch auf die Anbieter zurück, die Sie in `fallback_services` für diesen Client festgelegt haben. Um explizit auf Nicht-Claude-Fallback umzusteigen:

```yaml
proxy:
  anthropic_fallback_model: "claude-haiku-4-5-20251001"
```

(Der leere Standard sorgt dafür, dass das Dispatch einen Fehler zurückgibt, sodass Fehlrouting sofort sichtbar wird.)

### Cursor / Continue

In Cursor **Settings → AI → OpenAI API**:

```
Base URL:  https://localhost:56244
API Key:   <ihr-vault-client-token>
Model:     gemini-2.5-flash    # oder jedes Modell, das wall-vault kennt
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
      "apiKey": "<ihr-vault-client-token>"
    }
  ]
}
```

### OpenClaw

OpenClaw ist ein TUI-Agenten-Framework, das wall-vault ursprünglich bedienen sollte. Das **Add Agent**-Modal des Dashboards setzt den Agententyp auf `openclaw` (oder `nanoclaw`); wall-vault schreibt dann direkt `~/.openclaw/openclaw.json`, einschließlich Provider-URLs, Vault-Token und Modelleinträgen:

```bash
VAULT_CLIENT_ID=my-bot \
VAULT_URL=http://<vault-host>:56243 \
VAULT_TOKEN=<ihr-client-token> \
wall-vault proxy
```

### curl / Skripte

```bash
curl -X POST https://localhost:56244/v1/chat/completions \
  -H "Authorization: Bearer <ihr-vault-client-token>" \
  -H "Content-Type: application/json" \
  -d '{
    "model": "gemini-2.5-flash",
    "messages": [{"role": "user", "content": "ping"}]
  }'
```

---

## Konfiguration

`wall-vault setup` schreibt entweder `./wall-vault.yaml` oder `~/.wall-vault/config.yaml`. Bearbeiten Sie von Hand für Felder, nach denen der Assistent nicht fragt.

```yaml
mode: standalone     # standalone | distributed
lang: en             # en | ko | zh | ja | es | fr | de | pt | ar | hi | id | th | sw | ha | ne | mn | zu
theme: light         # light | dark | cherry | ocean | gold | autumn | winter

proxy:
  port: 56244
  host: ""           # Standard: 127.0.0.1 standalone, 0.0.0.0 distributed
  client_id: my-bot
  vault_url: ""      # distributed: http://vault-host:56243
  vault_token: ""    # distributed: Client-Token
  tool_filter: strip_all   # strip_all | whitelist | passthrough
  allowed_tools: []
  services: [google, openrouter, ollama]
  timeout: 300s
  tls:
    enabled: false
    cert_file: ""
    key_file: ""
  plain_port: 56245              # Loopback-only HTTP-Begleiter, wenn TLS aktiv
  ollama_keep_alive: "30m"       # "-1" niemals entladen, "0" sofort entladen
  ollama_num_ctx: 8192
  oai_stream_forward: false      # opt-in echter Backend-SSE-Passthrough
  anthropic_fallback_model: ""   # opt-in Nicht-Claude-Umschreibung beim Anthropic-Dispatch

vault:
  port: 56243
  host: ""
  admin_token: ""
  admin_ip_whitelist: []
  master_password: ""            # AES-GCM-Schlüsselverschlüsselungspasswort
  data_dir: ~/.wall-vault/data
  services_dir: ~/.wall-vault/services
  tls:
    enabled: false
    cert_file: ""
    key_file: ""
  bootstrap_port: 56247          # Plain-HTTP-Listener, der nur ca.crt ausliefert

doctor:
  interval: 5m
  auto_fix: true
  log_file: /tmp/wall-vault-doctor.log

hooks:
  on_model_change: ""            # Shell-Befehl (env: SERVICE, MODEL)
  on_key_exhausted: ""
  on_service_down: ""
  on_doctor_fix: ""
  openclaw_socket: ""
```

### Umgebungsvariablen

Jedes YAML-Feld hat eine Env-Override, die Vorrang vor der Datei hat. Häufig verwendete:

| Variable | Beschreibung |
|----------|-------------|
| `WV_LANG`, `WV_THEME` | Sprache und Theme |
| `WV_PROXY_PORT`, `WV_PROXY_HOST` | Proxy-Lauschadresse |
| `WV_VAULT_PORT`, `WV_VAULT_HOST` | Vault-Lauschadresse |
| `WV_VAULT_URL`, `WV_VAULT_TOKEN` | Endpunkte für verteilten Modus |
| `WV_ADMIN_TOKEN`, `WV_MASTER_PASS` | Vault-Anmeldedaten |
| `WV_KEY_GOOGLE`, `WV_KEY_OPENROUTER`, `WV_KEY_ANTHROPIC`, `WV_KEY_OPENAI` | API-Schlüssel (kommagetrennt für mehrere) |
| `WV_PROXY_TLS_ENABLED`, `WV_PROXY_TLS_CERT`, `WV_PROXY_TLS_KEY` | Proxy-TLS |
| `WV_VAULT_TLS_ENABLED`, `WV_VAULT_TLS_CERT`, `WV_VAULT_TLS_KEY` | Vault-TLS |
| `WV_PROXY_PLAIN_PORT` | Loopback-HTTP-Begleiter (`0` zum Deaktivieren) |
| `WV_VAULT_BOOTSTRAP_PORT` | CA-Bootstrap-Listener (`0` zum Deaktivieren) |
| `WV_OLLAMA_URL`, `WV_OLLAMA_KEEP_ALIVE`, `WV_OLLAMA_NUM_CTX` | Ollama-Tuning |
| `WV_LMSTUDIO_URL`, `WV_VLLM_URL`, `WV_LLAMACPP_URL` | Lokale Backend-Overrides |
| `WV_TOKEN_SENTINEL_FALLBACK` | Loopback-„proxy-managed“-Sentinel-Substitution |
| `WV_OAI_STREAM_FORWARD` | OpenAI-kompat echter Backend-SSE-Passthrough |
| `WV_ANTHROPIC_FALLBACK_MODEL` | Opt-in Nicht-Claude-Umschreibung bei Anthropic |

---

## Modi

### Standalone (Standard)

Vault und Proxy laufen im selben Prozess. Am besten für einen einzelnen Host, der sowohl die Schlüssel als auch den Agenten beherbergt. Standardmäßig nur Loopback.

```bash
wall-vault start    # führt beide aus
```

### Distributed

Der Vault läuft auf einem Host (dem **Vault-Host**) und speichert alle Schlüssel; mehrere Proxies auf anderen Hosts authentifizieren sich jeweils mit einem Per-Client-Token. Nützlich, wenn mehrere Maschinen dieselben Schlüssel benötigen, ohne sie kopieren zu müssen.

**Vault-Host:**

```bash
WV_VAULT_HOST=0.0.0.0 wall-vault vault
```

**Jeder Proxy-Host:**

```bash
WV_VAULT_URL=http://<vault-host>:56243 \
WV_VAULT_TOKEN=<dieser-client-token> \
WV_PROXY_HOST=0.0.0.0 \
wall-vault proxy
```

Das **Add Client**-Modal des Dashboards prägt einen Token, registriert einen Agententyp, und der Proxy holt seine Konfiguration über SSE ohne Neustart ab.

---

## Plugin-yaml (Drop-in-Backend)

Jedes OpenAI-kompatible Backend kann als yaml unter `~/.wall-vault/services/` hinzugefügt werden. wall-vault erkennt es beim Start, registriert es als routbaren Dienst, und das Dispatch + die OAI-Kompat-Erkennungsmenge + die Gemini-Stream-Brücke sehen es alle ohne Codeänderungen.

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
inline_no_think_for_qwen3: false   # einschalten, wenn Ihr Backend den Marker entfernt
```

Hub-Topologie (ein wall-vault vor einem anderen) wird über `tls_internal_ca: true`, `auth.type: bearer` und `preserve_model_id: true` unterstützt.

---

## Aus Quellen bauen

```bash
git clone https://github.com/sookmook/wall-vault
cd wall-vault
go build -o wall-vault .
./wall-vault setup
./wall-vault start
```

Cross-Compile für die gesamte unterstützte Auswahl:

```bash
make build-all   # linux-amd64, linux-arm64, darwin-amd64, darwin-arm64, windows-amd64
```

Versionen folgen `v{major}.{minor}.{patch}.{YYYYMMDD}.{HHmmss}`; `BASE_VERSION` im Makefile setzt das Präfix.

### Projektaufbau

```
wall-vault/
├── main.go                     # CLI-Dispatch (start/proxy/vault/setup/cert/doctor)
├── cmd/
│   ├── setup/                  # interaktiver Setup-Assistent
│   └── cert/                   # interne CA + TLS-Zertifikataussteller pro Host
├── internal/
│   ├── config/                 # YAML- + Env-Loader, Plugin-Loader
│   ├── proxy/                  # Anfrage-Dispatch, Schlüsselrotation, Formatkonverter
│   ├── vault/                  # AES-GCM-Store, Dashboard, SSE-Broker
│   ├── doctor/                 # Gesundheits-Probe + Auto-Fix
│   ├── hooks/                  # Shell-Befehl-Ereignisauslöser
│   └── i18n/                   # UI-Strings in 17 Sprachen
├── configs/services/           # mitgelieferte Plugin-yamls (lmstudio, vllm, ollama, …)
└── docs/                       # MANUAL, API-Referenz, 16 Locale-Varianten
```

---

## Dokumentation

- [Benutzerhandbuch](docs/MANUAL.en.md) — Installation, Dashboard, Agenten, Fehlerbehebung
- [API-Referenz](docs/API.en.md) — jeder Endpunkt mit Anfrage-/Antwortformen
- [CHANGELOG](CHANGELOG.md)

---

## Tech-Stack

- Go 1.25, eine einzelne statische Binärdatei
- [templ](https://templ.guide) für serverseitig gerendertes Dashboard, [HTMX](https://htmx.org) für Teilaktualisierungen
- AES-GCM (PBKDF2-abgeleiteter Schlüssel) für Schlüsselverschlüsselung im Ruhezustand
- Server-Sent Events für Live-Konfigurationssynchronisation zwischen Vault und Proxies
- Selbstsignierte interne CA + Per-Host-Zertifikate (kein öffentliches DNS / Let's Encrypt erforderlich)

## Lizenz

GPL-3.0. Siehe [LICENSE](LICENSE).

## Mitwirken

Pull Requests sind willkommen. Siehe [CONTRIBUTING.md](CONTRIBUTING.md). Bei größeren Änderungen öffnen Sie bitte zuerst ein Issue, um das Design zu besprechen.
