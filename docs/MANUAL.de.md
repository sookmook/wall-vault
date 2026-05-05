# wall-vault Benutzerhandbuch

[English](MANUAL.md) · [한국어](MANUAL.ko.md) · [中文](MANUAL.zh.md) · [日本語](MANUAL.ja.md) · [Español](MANUAL.es.md) · [Français](MANUAL.fr.md) · **Deutsch** · [Português](MANUAL.pt.md) · [العربية](MANUAL.ar.md) · [हिन्दी](MANUAL.hi.md) · [Bahasa Indonesia](MANUAL.id.md) · [ภาษาไทย](MANUAL.th.md) · [Kiswahili](MANUAL.sw.md) · [Hausa](MANUAL.ha.md) · [नेपाली](MANUAL.ne.md) · [Монгол](MANUAL.mn.md) · [isiZulu](MANUAL.zu.md)

Dieses Handbuch behandelt Installation, Konfiguration und Betrieb von wall-vault. Eine schnelle Übersicht finden Sie in der [README](../README.md). Details zur HTTP-API finden Sie in der [API-Referenz](API.md).

## Inhalt

1. [Was wall-vault tut](#was-wall-vault-tut)
2. [Installation](#installation)
3. [Erster Start mit dem Setup-Assistenten](#erster-start-mit-dem-setup-assistenten)
4. [TLS aktivieren](#tls-aktivieren)
5. [API-Schlüssel registrieren](#api-schlüssel-registrieren)
6. [Agenten verbinden](#agenten-verbinden)
7. [Das Dashboard](#das-dashboard)
8. [Verteilter Modus](#verteilter-modus)
9. [Auto-Start](#auto-start)
10. [Plugin-yamls](#plugin-yamls)
11. [Doctor](#doctor)
12. [Hooks](#hooks)
13. [Umgebungsvariablen](#umgebungsvariablen)
14. [Fehlerbehebung](#fehlerbehebung)

---

## Was wall-vault tut

wall-vault ist eine einzelne Go-Binärdatei, die zwei zusammenarbeitende Dienste bündelt:

- **Der Tresor (vault)** speichert API-Schlüssel verschlüsselt im Ruhezustand (AES-GCM mit einem Master-Passwort), verfolgt Nutzung und Cooldowns pro Schlüssel, sendet Änderungen über Server-Sent Events (SSE) und stellt unter `:56243` ein Web-Dashboard für menschliche Operatoren bereit.
- **Der Proxy** stellt unter `:56244` Endpunkte für Gemini, Anthropic, OpenAI-kompatibel und Ollama-nativ bereit. Jeder KI-Client, der auf den Proxy zeigt, verwendet die Schlüssel im Tresor — Clients sehen sie nie. Wenn ein Upstream fehlschlägt, fällt der Versand auf den nächsten Anbieter in der Reihenfolge zurück.

Dies ist nützlich, wenn:

- Sie Schlüssel für mehrere Anbieter haben und eine einzige URL möchten, mit der der Agent kommuniziert.
- Sie möchten, dass ein Free-Tier-Schlüssel im Cooldown beiseitetritt, ohne die Sitzung zu unterbrechen.
- Sie möchten, dass dieselben Schlüssel mehrere Bots, IDEs oder Skripte im selben LAN antreiben, ohne Anmeldedaten zu kopieren.
- Sie ein Dashboard anstelle von Umgebungsvariablen zum Bearbeiten von Schlüsseln und Wechseln von Modellen wünschen.
- Sie einen lokalen Fallback (Ollama, LM Studio, vLLM) möchten, wenn die Cloud-Limits erschöpft sind.

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

## Installation

### Linux / macOS Einzeiler

```bash
curl -fsSL https://raw.githubusercontent.com/sookmook/wall-vault/main/install.sh | sh
```

Das Skript erkennt Betriebssystem und Architektur automatisch, lädt die richtige Binärdatei nach `~/.local/bin/wall-vault` herunter und macht sie ausführbar. Falls `~/.local/bin` nicht in Ihrem `PATH` ist, fügen Sie es hinzu:

```bash
export PATH="$HOME/.local/bin:$PATH"
```

### Manueller Download

Vorgefertigte Binärdateien werden bei jedem Release auf `https://github.com/sookmook/wall-vault/releases` veröffentlicht.

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

### Aus dem Quellcode bauen

Erfordert Go 1.25 oder neuer.

```bash
git clone https://github.com/sookmook/wall-vault
cd wall-vault
go build -o wall-vault .
```

`make build-all` kompiliert für alle fünf unterstützten Plattformen quer. Die Binärdateien landen in `bin/`.

---

## Erster Start mit dem Setup-Assistenten

```bash
wall-vault setup
```

Der Assistent fragt Sie der Reihe nach:

1. **Sprache** — wählt eine von 17 UI-Sprachen. Wird automatisch aus `$LANG` erkannt; der Assistent bietet trotzdem eine Liste an.
2. **Theme** — `light` (Standard), `dark`, `cherry`, `ocean`, `gold`, `autumn`, `winter`. Nur kosmetisch.
3. **Modus** — `standalone` (Single-Host, Standard) oder `distributed` (Vault auf einem Host, Proxies auf anderen).
4. **Bot-Name** — ein freier `client_id`-Slug. Der Tresor nutzt diesen, um die Konfiguration pro Client einzugrenzen (Modell-Overrides, Fallback-Ketten).
5. **Proxy-Port** — Standard `56244`.
6. **Vault-Port** — Standard `56243` (nur standalone).
7. **Dienst-Auswahl** — ein y/N für jeden von: Google Gemini, OpenRouter, Anthropic, OpenAI, Ollama, LM Studio, vLLM. Mehrere Auswahlen sind in Ordnung; jede schreibt am Ende ihren Umgebungsvariablen-Hinweis.
8. **Tool-Filter** — `strip_all` (Standard; blockiert aus Sicherheitsgründen alle eingehenden Tool-Definitionen) oder `passthrough` (lässt alle Tools durch).
9. **Admin-Token** — leer lassen für automatische Generierung. Das Dashboard erfordert dieses Token zum Anmelden.
10. **Master-Passwort** — leer lassen für keine Verschlüsselung (NICHT empfohlen); Wert setzen, um den Schlüsselspeicher mit AES-GCM im Ruhezustand zu verschlüsseln.
11. **Speicherpfad** — Standard ist `wall-vault.yaml` im aktuellen Verzeichnis. Der Loader sucht auch in `~/.wall-vault/config.yaml`.

Nach dem Speichern führt der Assistent `doctor.FixTrust` aus, sodass jeder lokal installierte Agent (OpenClaw, Claude Code, Cline) die wall-vault interne CA automatisch in seinen Trust-Store erhält. Wenn kein solcher Agent installiert ist, gibt der Schritt `SKIP` aus und schreibt nichts.

Dann starten Sie die Binärdatei:

```bash
wall-vault start
```

`start` führt sowohl den Tresor als auch den Proxy in einem Prozess aus (Standalone-Modus). Für den verteilten Modus verwenden Sie `wall-vault vault` auf dem Tresor-Host und `wall-vault proxy` auf jedem Proxy-Host.

Öffnen Sie `http://localhost:56243` in einem Browser. Melden Sie sich mit dem Admin-Token an, das der Assistent ausgegeben hat.

---

## TLS aktivieren

Die Standardeinstellungen des Assistenten lassen beide Listener auf reinem HTTP. Die meisten Agenten (OpenClaw, Claude Code, Cursor) funktionieren besser gegen einen einzigen HTTPS-Endpunkt, daher wird TLS in jeder Bereitstellung empfohlen, die über die lokale Maschine hinausgeht.

wall-vault bringt seine eigene interne CA mit, sodass Sie keinen öffentlichen DNS-Namen oder Let's Encrypt benötigen.

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

Um das Vertrauen auf andere LAN-Maschinen auszudehnen, kopieren Sie `~/.wall-vault/ca.crt` und führen Sie auf jeder `wall-vault cert install-trust --ca <path>` aus. Der Tresor stellt `ca.crt` außerdem über einen winzigen reinen HTTP-Listener auf `:56247` (dem **Bootstrap-Port**) bereit, für den Catch-22-Fall, in dem ein neuer Client die CA benötigt, um HTTPS zu sprechen.

### HTTP-Begleiter im Loopback

Einige Agenten — insbesondere die mitgelieferte Node-Laufzeit von OpenClaw — überschreiben `NODE_EXTRA_CA_CERTS` beim Prozess-Spawn und verwerfen jeden vom Operator bereitgestellten CA-Hinweis. Sie können die wall-vault CA aus dem Daemon heraus nicht ehren, auch nach `cert install-trust` nicht. wall-vault umgeht dies, indem es bei aktiviertem TLS einen zusätzlichen **nur-Loopback reinen HTTP-Listener** auf `127.0.0.1:56245` bindet. Same-Host-Clients erreichen den Proxy über diesen Port ganz ohne TLS; LAN-Clients verwenden weiterhin den TLS-Listener.

Mit `WV_PROXY_PLAIN_PORT=0` deaktivieren, falls nicht benötigt.

### `wall-vault cert list`

Zeigt jedes Zertifikat unter `~/.wall-vault/` mit Subject, Gültigkeitsfenster und SANs.

```
$ wall-vault cert list
ca.crt          subject=wall-vault internal CA   not-after=2036-05-05
hostname.crt    subject=hostname                 not-after=2031-05-05   SAN=hostname,localhost,127.0.0.1,192.168.…
```

---

## API-Schlüssel registrieren

Zwei Wege: das Dashboard oder Umgebungsvariablen.

### Dashboard (empfohlen)

1. Melden Sie sich auf `https://localhost:56243` mit dem Admin-Token an.
2. Klicken Sie auf **+ API key** in der Schlüsselkarte.
3. Wählen Sie einen Dienst (Google, OpenRouter, Anthropic, OpenAI, …).
4. Fügen Sie den Schlüssel ein. Speichern.

Mehrere Schlüssel pro Dienst sind in Ordnung; der Proxy macht Round-Robin zwischen ihnen und überspringt jene, die einen Cooldown pro Schlüssel erreicht haben.

### Umgebungsvariablen (einmaliger Bootstrap)

```bash
export WV_KEY_GOOGLE="AIzaSyA1...,AIzaSyB2...,AIzaSyC3..."   # comma-separated
export WV_KEY_OPENROUTER="sk-or-v1-…"
export WV_KEY_ANTHROPIC="sk-ant-…"
export WV_KEY_OPENAI="sk-…"
wall-vault start
```

Auf diese Weise bereitgestellte Schlüssel werden beim ersten Start in den verschlüsselten Speicher geschrieben. Nachfolgende Starts lesen sie von der Festplatte; Sie können die Umgebungsvariablen nach dem ersten Lauf entsetzen.

### Cooldowns und Rotation

Jeder erfolgreiche Aufruf erhöht die `usage_count` des Schlüssels und aktualisiert `last_used`. Bei HTTP 429 / 402 / 403 setzt der Proxy den Schlüssel in einen **Cooldown** (Standard: 60 Minuten für 429, 24 Stunden für 402, 12 Stunden für 403). Der nächste Versand wählt einen anderen Schlüssel für diesen Dienst. Wenn alle Schlüssel für einen Dienst im Cooldown sind, überspringt der Proxy diesen Dienst schnell vollständig und versucht den nächsten Anbieter in der Fallback-Kette.

Cooldowns sind pro Schlüssel im Dashboard mit einem Countdown sichtbar.

---

## Agenten verbinden

### OpenClaw

OpenClaw ist der ursprüngliche Zielclient. Verwenden Sie das **+ Add agent** Modal des Dashboards:

- Setzen Sie **Agent type** auf `openclaw` oder `nanoclaw`.
- Setzen Sie **Work directory** — bei OpenClaw wird dies automatisch mit `~/.openclaw` gefüllt.
- Wählen Sie einen **preferred service** und optional einen **model override**.
- Klicken Sie auf **Apply**. wall-vault schreibt direkt `~/.openclaw/openclaw.json` (Anbieter-URLs, Vault-Token, Modelleinträge).

Wenn Sie das Modell aus dem Dashboard ändern, übernimmt OpenClaw die Änderung über SSE innerhalb von 1–3 Sekunden — kein Neustart.

### Claude Code

```bash
export ANTHROPIC_BASE_URL=https://localhost:56244
export ANTHROPIC_API_KEY=<your-vault-client-token>
claude
```

Wenn die Upstream-Anthropic-Credits aufgebraucht sind, fällt der Versand auf jene Dienste zurück, die in `fallback_services` dieses Clients aufgeführt sind. Standardmäßig gibt eine an den anthropic-Dispatch gesendete Nicht-Claude-Modell-ID einen Fehler zurück, sodass Fehlroutings sofort erkennbar sind. Aktivieren Sie das automatische Umschreiben:

```yaml
proxy:
  anthropic_fallback_model: "claude-haiku-4-5-20251001"
```

### Cursor

In Cursor **Settings → AI → OpenAI API**:

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

### Eigenes HTTP

```bash
curl -X POST https://localhost:56244/v1/chat/completions \
  -H "Authorization: Bearer <your-vault-client-token>" \
  -H "Content-Type: application/json" \
  -d '{
    "model": "gemini-2.5-flash",
    "messages": [{"role": "user", "content": "hello"}]
  }'
```

Derselbe Endpunkt akzeptiert Streaming (`"stream": true`), wenn `proxy.oai_stream_forward: true` gesetzt ist.

---

## Das Dashboard

`https://localhost:56243`. Fünf Karten im Home-Grid:

- **Keys** — jeder API-Schlüssel, gruppiert nach Dienst. Hinzufügen, bearbeiten, löschen; Nutzung und Cooldown sehen.
- **Services** — Google / OpenRouter / Anthropic / OpenAI / Ollama / LM Studio / vLLM / llama.cpp, plus jedes Plugin-yaml in `~/.wall-vault/services/`. Pro Dienst `default_model`, `allowed_models`, Basis-URL, Reasoning-Toggle setzen.
- **Clients (agents)** — jeder registrierte Client (OpenClaw-Bot, Claude Code-Sitzung, Cursor-Instanz, …). Bevorzugten Dienst, Modell-Override, Fallback-Kette zuweisen.
- **Proxies** — jeder Proxy, der sich gegen diesen Tresor authentifiziert hat. Live-Status (online/offline), zuletzt gesehen, aktuelles Modell.
- **Settings** — Admin-Token, Master-Passwort-Rotation, Theme, Sprache.

Jede Karte hat einen Edit-Slideover (rechte Seite). Klick außerhalb oder `Esc` schließt ihn. Änderungen werden innerhalb von Sekunden über SSE an alle verbundenen Proxies übertragen.

Der **Footer** trägt einen SSE-Indikator (grün = verbunden, orange = wieder verbinden, grau = getrennt) und die Live-Build-Version.

---

## Verteilter Modus

Wenn Sie mehrere Maschinen haben, die alle dieselben Schlüssel benötigen, lassen Sie den Tresor auf einem Host und Proxies auf jedem der anderen laufen.

### Tresor-Host

```bash
WV_VAULT_HOST=0.0.0.0 \
WV_ADMIN_TOKEN=<admin> \
WV_MASTER_PASS=<master> \
wall-vault vault
```

Das Dashboard ist nun unter `https://<vault-host>:56243` erreichbar. Fügen Sie für jeden entfernten Proxy einen Agenten in der **Clients**-Karte hinzu; jeder erzeugt ein eindeutiges `vault_token`.

### Proxy-Hosts

```bash
WV_VAULT_URL=http://<vault-host>:56243 \
WV_VAULT_TOKEN=<that-client-token> \
WV_PROXY_HOST=0.0.0.0 \
wall-vault proxy
```

Der Proxy authentifiziert sich gegen den Tresor, öffnet einen SSE-Stream und wendet jede Konfiguration an, die er erhält (bevorzugter Dienst, Modell-Override, Fallback-Kette). Nachfolgende Tresor-Änderungen kommen in Sekunden ohne Neustart an.

Für LAN-übergreifende Installationen aktivieren Sie TLS auf dem Tresor-Host (`WV_VAULT_TLS_ENABLED=1` + die cert/key-Umgebungsvariablen) und führen Sie jeden Proxy-Host durch denselben `wall-vault cert install-trust`-Schritt, sodass die HTTPS-Aufrufe des Proxy in den Tresor vertraut werden.

---

## Auto-Start

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

Für den Tresor auf demselben Host schreiben Sie eine parallele `wall-vault-vault.service`. Für den Standalone-Modus genügt eine Unit, die `wall-vault start` aufruft.

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

Verwenden Sie `nssm`, um `wall-vault.exe start` als Windows-Dienst zu umhüllen, oder einen `schtasks`-Eintrag, der bei der Benutzeranmeldung läuft.

---

## Plugin-yamls

Jedes OpenAI-kompatible Backend kann ohne Codeänderungen hinzugefügt werden, indem ein yaml unter `~/.wall-vault/services/` abgelegt wird. wall-vault lädt es beim Start und registriert den Dienst für den Versand, das OAI-Compat-Erkennungsset und die Gemini-Stream-Bridge.

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

Das mitgelieferte Set in `configs/services/` (lmstudio, vllm, llamacpp, tgwui, localai, jan, koboldcpp, tabbyapi, mlx-server, litellm-proxy, ollama, google, openrouter) wird standardmäßig deaktiviert ausgeliefert. Kopieren Sie das gewünschte nach `~/.wall-vault/services/`, setzen Sie `enabled: true`, starten Sie neu.

---

## Doctor

`wall-vault doctor` führt eine einmalige Health-Probe über die gesamte Installation aus:

```
✓ vault listener  (https://localhost:56243)
✓ proxy listener  (https://localhost:56244)
✓ master password set
⚠ Google: 2 keys, all on cooldown
✓ Anthropic: 1 key healthy
✗ Ollama: not reachable at http://localhost:11434
```

Jede Zeile ist eine von:

- `✓` — gesund
- `⚠` — beeinträchtigt, aber funktionsfähig (ein Schlüssel im Cooldown, niedrige Quote, etc.)
- `✗` — kaputt
- `SKIP` — nicht konfiguriert / auf diesem Host nicht zutreffend

Ein zweiter Daemon-Modus führt dieselbe Probe alle `doctor.interval` (Standard 5 Minuten) aus und schreibt die Ergebnisse in `doctor.log_file` (Standard `/tmp/wall-vault-doctor.log`). Wenn `doctor.auto_fix` true ist, versucht er auch, gängige Drift zu reparieren (veraltete OpenClaw-Konfiguration, fehlendes TLS-Trust, neustartbare Dienste).

Lösen Sie eine einmalige Ausführung über die **Doctor**-Karte des Dashboards oder `wall-vault doctor` aus.

---

## Hooks

Führen Sie einen Shell-Befehl bei Schlüsselereignissen aus:

```yaml
hooks:
  on_model_change:   "logger 'wall-vault: $SERVICE/$MODEL'"
  on_key_exhausted:  "notify-send 'wall-vault' '$SERVICE keys all on cooldown'"
  on_service_down:   "/usr/local/bin/page-oncall.sh $SERVICE '$ERROR'"
  on_doctor_fix:     "echo \"$AGENT: $LEVEL $MSG\" >> ~/wall-vault.audit.log"
  openclaw_socket:   ""    # if set, OpenClaw TUI receives events over this Unix socket
```

Jeder Hook erhält ereignisspezifische Umgebungsvariablen (`SERVICE`, `MODEL`, `ERROR`, `AGENT`, `LEVEL`, `MSG`). Hooks laufen asynchron mit einem 5-Sekunden-Timeout — der Proxy blockiert nie auf einem langsamen Hook.

---

## Umgebungsvariablen

| Variable | YAML-Feld |
|----------|-----------|
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
| `WV_VAULT_TLS_ENABLED` | `vault.tls.enabled` |
| `WV_VAULT_TLS_CERT` | `vault.tls.cert_file` |
| `WV_VAULT_TLS_KEY` | `vault.tls.key_file` |
| `WV_VAULT_BOOTSTRAP_PORT` | `vault.bootstrap_port` |
| `WV_PROXY_PLAIN_PORT` | `proxy.plain_port` |
| `WV_KEY_GOOGLE` | Einmaliger Import: kommagetrennte Google-Schlüssel |
| `WV_KEY_OPENROUTER` | Einmaliger Import: OpenRouter-Schlüssel |
| `WV_KEY_ANTHROPIC` | Einmaliger Import: Anthropic-Schlüssel |
| `WV_KEY_OPENAI` | Einmaliger Import: OpenAI-Schlüssel |
| `WV_OLLAMA_URL` | Pro-Host Ollama-URL-Override |
| `WV_OLLAMA_KEEP_ALIVE` | `proxy.ollama_keep_alive` |
| `WV_OLLAMA_NUM_CTX` | `proxy.ollama_num_ctx` |
| `WV_LMSTUDIO_URL`, `WV_VLLM_URL`, `WV_LLAMACPP_URL` | Pro-Backend URL-Override |
| `WV_TOKEN_SENTINEL_FALLBACK` | `proxy.token_sentinel_fallback` |
| `WV_OAI_STREAM_FORWARD` | `proxy.oai_stream_forward` |
| `WV_ANTHROPIC_FALLBACK_MODEL` | `proxy.anthropic_fallback_model` |
| `WV_ECONOWORLD_MAX_TOKENS` | `proxy.econoworld_max_tokens` |
| `WV_ECONOWORLD_STREAM` | `proxy.econoworld_stream` |
| `WV_ECONOWORLD_REQUEST_TIMEOUT` | `proxy.econoworld_request_timeout` |

Jede Umgebungsvariable, wenn gesetzt, gewinnt gegen die YAML-Datei.

---

## Fehlerbehebung

### `connection refused` auf `:56244`

Entweder läuft der Proxy nicht oder er ist an einen anderen Host gebunden. Prüfen:

```bash
ss -lnp | grep 56244
systemctl --user status wall-vault-proxy   # Linux
launchctl list | grep wall-vault           # macOS
```

Wenn er auf einem anderen Port läuft, hat Ihre Konfiguration `proxy.port` überschrieben — prüfen Sie `~/.wall-vault/config.yaml`.

### `x509: certificate signed by unknown authority`

Der Client vertraut der wall-vault internen CA nicht. Führen Sie `wall-vault cert install-trust` auf der Client-Maschine aus. Für Agenten, deren Laufzeit den OS-Trust-Store ignoriert (z. B. Node mit hartcodiertem `NODE_EXTRA_CA_CERTS`), verwenden Sie den Loopback-HTTP-Begleiter auf `127.0.0.1:56245` (nur Same-Host) oder setzen Sie `WV_PROXY_TLS_ENABLED=0`, um auf reines HTTP zurückzufallen.

### `token not registered with vault`

Das `Authorization: Bearer <token>` des Clients passt zu keinem registrierten Client. Überprüfen Sie das Token unter **Clients** im Dashboard. Wenn Sie ein Token-Literal wie `proxy-managed`, `dummy` oder `""` aus einer veralteten Konfiguration kopiert haben, ersetzen Sie es durch das echte Client-Token.

### `Anthropic dispatch needs a Claude model id`

Standardverhalten ab v0.2.63: eine an den anthropic-Dispatch gesendete Nicht-Claude-Modell-ID gibt einen Fehler zurück. Entweder das Routing korrigieren (senden Sie nicht `gemini-2.5-flash` an anthropic) oder das automatische Umschreiben über `proxy.anthropic_fallback_model` aktivieren.

### `unknown service: <id>`

Der Versand sah eine Service-ID, die kein Plugin-yaml beanspruchte. Prüfen:

```bash
ls ~/.wall-vault/services/        # any plugin yaml present?
cat ~/.wall-vault/services/<id>.yaml | grep enabled
```

Wenn das yaml existiert, aber `enabled: false` ist, kippen Sie es. Wenn es ganz fehlt, kopieren Sie aus `configs/services/` im Quellbaum.

### Leere Antwort bei einem Reasoning-Modell

`qwen3.6`, `deepseek-r1` und die GPT-`o1`-Familie geben manchmal nur `reasoning_content` aus und lassen `content` leer. Ab v0.2.63 fällt wall-vault automatisch auf den Reasoning-Text zurück — wenn Sie immer noch leere Antworten sehen, gibt das Backend keines der beiden Felder zurück. Prüfen Sie die Upstream-Logs.

Speziell für LM Studio mit qwen3 setzen Sie `inline_no_think_for_qwen3: true` im Plugin-yaml, damit Reasoning inline deaktiviert wird. Die mitgelieferten lmstudio.yaml und ollama.yaml tun dies bereits.

### Dashboard zeigt „alle Schlüssel im Cooldown", aber ich habe gerade einen hinzugefügt

Der neue Schlüssel ist gesund, aber der Versandpfad kann sich noch im Cooldown für einen älteren Schlüssel befinden. Versuchen Sie eine neue Anfrage — der Proxy macht Round-Robin pro Aufruf, und ein gesunder Schlüssel wird als nächster gewählt.

### Tresor öffnet nicht mit dem Master-Passwort

Falsches Passwort. Es gibt keine Wiederherstellung — wall-vault liefert absichtlich keine Hintertür. Wenn Sie das Master-Passwort wirklich verloren haben, ist der einzige Weg, `~/.wall-vault/data/vault.json` zu löschen, mit einem neuen Passwort neu zu starten und die Schlüssel wieder hinzuzufügen.

### Free-Tier OpenRouter-Limits erreicht

Setzen Sie `proxy.services` so, dass es `openrouter` einschließt, und fügen Sie mindestens einen OpenRouter-Schlüssel hinzu. Der Proxy fällt automatisch von einem bezahlten Modell auf seine `:free`-Variante zurück, wenn der bezahlte Pfad 402 / 429 zurückgibt.

### `journalctl --user -u wall-vault-proxy` ist leer

systemd `--user`-Logs gehen in das Journal des Benutzers, der es ausführt. Wenn Sie die Unit als `root` oder über `sudo` gestartet haben, ist das Journal stattdessen in der System-Instanz — versuchen Sie `journalctl -u wall-vault-proxy` ohne `--user`.

---

## Mehr

- HTTP-API-Referenz — siehe [API.md](API.md)
- Quellcode — `https://github.com/sookmook/wall-vault`
- Bug-Reports / Feature-Wünsche — GitHub Issues
- Release-Historie — [CHANGELOG.md](../CHANGELOG.md)
