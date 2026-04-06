# wall-vault Benutzerhandbuch
*(Zuletzt aktualisiert: 2026-04-06 — v0.1.23)*

---

## Inhaltsverzeichnis

1. [Was ist wall-vault?](#was-ist-wall-vault)
2. [Installation](#installation)
3. [Erste Schritte (Setup-Assistent)](#erste-schritte)
4. [API-Schlüssel registrieren](#api-schlüssel-registrieren)
5. [Den Proxy verwenden](#den-proxy-verwenden)
6. [Das Schlüsseltresor-Dashboard](#das-schlüsseltresor-dashboard)
7. [Verteilter Modus (Multi-Bot)](#verteilter-modus-multi-bot)
8. [Autostart einrichten](#autostart-einrichten)
9. [Doctor-Diagnose](#doctor-diagnose)
10. [Umgebungsvariablen – Übersicht](#umgebungsvariablen--übersicht)
11. [Problemlösung](#problemlösung)

---

## Was ist wall-vault?

**wall-vault = KI-Proxy (Vermittler) + API-Schlüsseltresor für OpenClaw**

Um KI-Dienste zu nutzen, benötigt man einen **API-Schlüssel** – das ist eine Art **digitaler Ausweis**, der bestätigt: „Diese Person darf diesen Dienst verwenden." Solche Ausweise haben jedoch ein tägliches Nutzungslimit und können bei schlechter Verwaltung in falsche Hände geraten.

wall-vault bewahrt diese Ausweise sicher in einem verschlüsselten Tresor auf und fungiert als **Vermittler (Proxy)** zwischen OpenClaw und den KI-Diensten. Einfach gesagt: OpenClaw verbindet sich nur mit wall-vault – den Rest erledigt wall-vault automatisch im Hintergrund.

Probleme, die wall-vault für dich löst:

- **Automatische Schlüsselrotation**: Wenn ein Schlüssel sein Limit erreicht oder vorübergehend gesperrt ist (Cooldown), wechselt wall-vault still und leise zum nächsten. OpenClaw läuft ohne Unterbrechung weiter.
- **Automatischer Dienstwechsel (Fallback)**: Antwortet Google nicht, wechselt wall-vault automatisch zu OpenRouter – und wenn das auch nicht klappt, zu Ollama, LM Studio oder vLLM (lokale KI auf deinem Computer). Die Sitzung bleibt erhalten. Wenn der ursprüngliche Dienst wiederhergestellt ist, wird ab der nächsten Anfrage automatisch zurückgewechselt (v0.1.18+, LM Studio/vLLM: v0.1.21+).
- **Echtzeit-Synchronisierung (SSE)**: Wenn du im Dashboard ein Modell wechselst, wird die Änderung innerhalb von 1–3 Sekunden in OpenClaw übernommen. SSE (Server-Sent Events) ist eine Technologie, bei der der Server Änderungen sofort an den Client schickt – ohne dass der Client nachfragen muss.
- **Echtzeit-Benachrichtigungen**: Ereignisse wie ein erschöpfter Schlüssel oder ein Dienstausfall werden sofort in der TUI-Oberfläche (Terminal-Anzeige) von OpenClaw angezeigt.

> 💡 **Claude Code, Cursor und VS Code** können ebenfalls angebunden werden, aber der ursprüngliche Zweck von wall-vault ist die gemeinsame Nutzung mit OpenClaw.

```
OpenClaw (TUI – Terminal-Oberfläche)
        │
        ▼
  wall-vault Proxy (:56244)   ← Schlüsselverwaltung, Routing, Fallback, Ereignisse
        │
        ├─ Google Gemini API
        ├─ OpenRouter API (340+ Modelle)
        ├─ Ollama / LM Studio / vLLM (lokal auf deinem Computer, letzte Rückfalloption)
        └─ OpenAI / Anthropic API
```

---

## Installation

### Linux / macOS

Öffne ein Terminal und füge die folgenden Befehle ein:

```bash
# Linux (normaler PC, Server — amd64)
curl -L https://github.com/sookmook/wall-vault/releases/latest/download/wall-vault-linux-amd64 \
  -o ~/.local/bin/wall-vault && chmod +x ~/.local/bin/wall-vault

# macOS Apple Silicon (M1/M2/M3 Mac)
curl -L https://github.com/sookmook/wall-vault/releases/latest/download/wall-vault-darwin-arm64 \
  -o /usr/local/bin/wall-vault && chmod +x /usr/local/bin/wall-vault
```

- `curl -L ...` — Lädt eine Datei aus dem Internet herunter.
- `chmod +x` — Macht die heruntergeladene Datei ausführbar. Wird dieser Schritt übersprungen, erscheint ein „Berechtigung verweigert"-Fehler.

### Windows

Öffne PowerShell (als Administrator) und führe die folgenden Befehle aus:

```powershell
# Download
Invoke-WebRequest -Uri `
  "https://github.com/sookmook/wall-vault/releases/latest/download/wall-vault-windows-amd64.exe" `
  -OutFile "$env:LOCALAPPDATA\Programs\wall-vault\wall-vault.exe"

# Zum PATH hinzufügen (wird nach Neustart von PowerShell wirksam)
$env:PATH += ";$env:LOCALAPPDATA\Programs\wall-vault"
```

> 💡 **Was ist PATH?** Eine Liste von Ordnern, in denen der Computer nach Befehlen sucht. Wenn du wall-vault zum PATH hinzufügst, kannst du `wall-vault` von jedem Verzeichnis aus starten.

### Aus dem Quellcode bauen (für Entwickler)

Dies gilt nur, wenn eine Go-Entwicklungsumgebung installiert ist.

```bash
git clone https://github.com/sookmook/wall-vault
cd wall-vault
make build       # bin/wall-vault (Version: v0.1.23.YYYYMMDD.HHmmss)
make install     # ~/.local/bin/wall-vault
```

> 💡 **Build-Zeitstempel-Version**: Beim Bauen mit `make build` wird die Version automatisch im Format `v0.1.23.20260406.211004` (inkl. Datum und Uhrzeit) generiert. Baut man direkt mit `go build ./...`, wird nur `"dev"` als Version angezeigt.

---

## Erste Schritte

### Den Setup-Assistenten starten

Führe nach der Installation unbedingt zuerst den **Setup-Assistenten** aus. Er führt dich Schritt für Schritt durch die Konfiguration.

```bash
wall-vault setup
```

Der Assistent durchläuft folgende Schritte:

```
1. Sprachauswahl (10 Sprachen, darunter Deutsch)
2. Design-Auswahl (light / dark / gold / cherry / ocean)
3. Betriebsmodus — Einzelnutzung (standalone) oder Mehrbetrieb (distributed)
4. Bot-Name — der im Dashboard angezeigte Name
5. Port-Einstellungen — Standard: Proxy 56244, Tresor 56243 (bei Bedarf einfach Enter drücken)
6. KI-Dienste wählen — Google / OpenRouter / Ollama / LM Studio / vLLM
7. Tool-Sicherheitsfilter konfigurieren
8. Admin-Token festlegen — ein Passwort zum Absichern der Dashboard-Verwaltung; kann automatisch generiert werden
9. API-Schlüssel-Verschlüsselungspasswort — für besonders sichere Schlüsselablage (optional)
10. Speicherort der Konfigurationsdatei
```

> ⚠️ **Merke dir unbedingt den Admin-Token.** Du brauchst ihn später, um im Dashboard Schlüssel hinzuzufügen oder Einstellungen zu ändern. Falls du ihn vergisst, musst du die Konfigurationsdatei manuell bearbeiten.

Nach Abschluss des Assistenten wird automatisch eine `wall-vault.yaml`-Konfigurationsdatei erstellt.

### Starten

```bash
wall-vault start
```

Es werden gleichzeitig zwei Server gestartet:

- **Proxy** (`http://localhost:56244`) — der Vermittler zwischen OpenClaw und den KI-Diensten
- **Schlüsseltresor** (`http://localhost:56243`) — API-Schlüsselverwaltung und Web-Dashboard

Öffne `http://localhost:56243` im Browser, um das Dashboard sofort zu sehen.

---

## API-Schlüssel registrieren

Es gibt vier Methoden, um API-Schlüssel zu registrieren. **Für Einsteiger empfehlen wir Methode 1 (Umgebungsvariablen).**

### Methode 1: Umgebungsvariablen (empfohlen – am einfachsten)

Umgebungsvariablen sind **voreingestellte Werte**, die ein Programm beim Start ausliest. Gib einfach Folgendes im Terminal ein:

```bash
# Google Gemini-Schlüssel registrieren
export WV_KEY_GOOGLE=AIzaSy...

# OpenRouter-Schlüssel registrieren
export WV_KEY_OPENROUTER=sk-or-v1-...

# Nach der Registrierung starten
wall-vault start
```

Wenn du mehrere Schlüssel hast, trenne sie mit Kommas. wall-vault nutzt sie automatisch abwechselnd (Round-Robin):

```bash
export WV_KEY_GOOGLE=AIzaSy...,AIzaSy...,AIzaSy...
```

> 💡 **Tipp**: Der Befehl `export` gilt nur für die aktuelle Terminal-Sitzung. Um ihn dauerhaft zu machen, füge die Zeile in `~/.bashrc` oder `~/.zshrc` ein.

### Methode 2: Dashboard-UI (per Mausklick)

1. Öffne `http://localhost:56243` im Browser
2. Klicke in der oberen **🔑 API-Schlüssel**-Karte auf `[+ Hinzufügen]`
3. Gib den Diensttyp, den Schlüsselwert, ein Label (Notizname) und das Tageslimit ein und speichere

### Methode 3: REST API (für Automatisierung/Skripte)

REST API ist eine Methode, mit der Programme über HTTP Daten austauschen. Nützlich für automatische Registrierung per Skript.

```bash
curl -X POST http://localhost:56243/admin/keys \
  -H "Authorization: Bearer dein-admin-token" \
  -H "Content-Type: application/json" \
  -d '{
    "service": "google",
    "key": "AIzaSy...",
    "label": "Hauptschlüssel",
    "daily_limit": 1000
  }'
```

### Methode 4: Proxy-Flags (für kurze Tests)

Hiermit kannst du temporär einen Schlüssel zum Testen übergeben, ohne ihn formal zu registrieren. Der Schlüssel verschwindet beim Beenden des Programms.

```bash
wall-vault proxy --key-google=AIzaSy... --key-openrouter=sk-or-...
```

---

## Den Proxy verwenden

### Verwendung mit OpenClaw (Hauptzweck)

So konfigurierst du OpenClaw, um über wall-vault mit KI-Diensten zu kommunizieren.

Öffne die Datei `~/.openclaw/openclaw.json` und füge Folgendes hinzu:

```json5
// ~/.openclaw/openclaw.json
{
  models: {
    providers: {
      "wall-vault": {
        baseUrl: "http://localhost:56244/v1",
        apiKey: "dein-agent-token",   // Vault-Agent-Token
        api: "openai-completions",
        models: [
          { id: "wall-vault/gemini-2.5-flash" },
          { id: "wall-vault/gemini-2.5-pro" },
          { id: "wall-vault/hunter-alpha" },    // kostenloser 1M-Kontext
          { id: "wall-vault/claude-opus-4-6" }
        ]
      }
    }
  }
}
```

> 💡 **Einfachere Methode**: Klicke auf den Button **🦞 OpenClaw-Konfiguration kopieren** auf der Agenten-Karte im Dashboard – damit wird ein Snippet mit bereits eingetragenem Token und Adresse in die Zwischenablage kopiert. Einfach einfügen.

**Wohin leitet das `wall-vault/`-Präfix im Modellnamen?**

wall-vault erkennt am Modellnamen automatisch, an welchen KI-Dienst die Anfrage weitergeleitet wird:

| Modellformat | Weiterleitung an |
|-------------|-----------------|
| `wall-vault/gemini-*` | Google Gemini direkt |
| `wall-vault/gpt-*`, `wall-vault/o3`, `wall-vault/o4*` | OpenAI direkt |
| `wall-vault/claude-*` | Anthropic über OpenRouter |
| `wall-vault/hunter-alpha`, `wall-vault/healer-alpha` | OpenRouter (kostenloser 1M-Token-Kontext) |
| `wall-vault/kimi-*`, `wall-vault/glm-*`, `wall-vault/deepseek-*` | OpenRouter |
| `google/Modellname`, `openai/Modellname`, `anthropic/Modellname` usw. | Direkt zum jeweiligen Dienst |
| `custom/google/Modellname`, `custom/openai/Modellname` usw. | `custom/`-Präfix entfernen und neu routen |
| `Modellname:cloud` | `:cloud`-Suffix entfernen und über OpenRouter leiten |

> 💡 **Was ist Kontext?** Die Gesprächsmenge, die eine KI auf einmal im Gedächtnis behalten kann. 1M (eine Million Token) bedeutet, dass auch sehr lange Gespräche oder Dokumente in einer einzigen Sitzung verarbeitet werden können.

### Direktverbindung im Gemini-API-Format (Kompatibilität mit bestehenden Tools)

Wenn du Tools hast, die bereits direkt die Google Gemini API nutzen, ändere einfach die Adresse auf wall-vault:

```bash
export ANTHROPIC_BASE_URL=http://localhost:56244/google
```

Oder wenn das Tool eine direkte URL erwartet:

```
http://localhost:56244/google/v1beta/models/gemini-2.5-flash:generateContent
```

### Verwendung mit dem OpenAI SDK (Python)

Du kannst wall-vault auch in Python-Code einbinden, der KI nutzt. Ändere einfach die `base_url`:

```python
from openai import OpenAI

client = OpenAI(
    base_url="http://localhost:56244/v1",
    api_key="not-needed"  # API-Schlüssel werden von wall-vault verwaltet
)

response = client.chat.completions.create(
    model="google/gemini-2.5-flash",   # im Format provider/model eingeben
    messages=[{"role": "user", "content": "Hallo"}]
)
```

### Modell während der Laufzeit wechseln

Um das KI-Modell zu wechseln, während wall-vault bereits läuft:

```bash
# Modell direkt über den Proxy ändern
curl -X PUT http://localhost:56244/api/config/model \
  -H "Content-Type: application/json" \
  -d '{"service": "openrouter", "model": "anthropic/claude-3.5-sonnet"}'

# Im verteilten Modus (Multi-Bot) auf dem Tresor-Server ändern → sofortige SSE-Synchronisierung
curl -X PUT http://localhost:56243/admin/clients/mein-bot-id \
  -H "Authorization: Bearer dein-admin-token" \
  -H "Content-Type: application/json" \
  -d '{"default_service": "google", "default_model": "gemini-2.5-pro"}'
```

### Verfügbare Modelle anzeigen

```bash
# Vollständige Liste anzeigen
curl http://localhost:56244/api/models | python3 -m json.tool

# Nur Google-Modelle anzeigen
curl "http://localhost:56244/api/models?service=google"

# Nach Namen suchen (z. B. Modelle mit „claude")
curl "http://localhost:56244/api/models?q=claude"
```

**Wichtige Modelle pro Dienst:**

| Dienst | Wichtige Modelle |
|--------|-----------------|
| Google | gemini-2.5-pro, gemini-2.5-flash, gemini-2.5-flash-8b, gemini-2.0-flash |
| OpenAI | gpt-4o, gpt-4o-mini, o3, o1, o1-mini |
| OpenRouter | 346+ (Hunter Alpha 1M-Kontext kostenlos, DeepSeek R1/V3, Qwen 2.5 usw.) |
| Ollama | Automatische Erkennung lokal installierter Modelle |
| LM Studio | Lokaler Server (Port 1234) |
| vLLM | Lokaler Server (Port 8000) |

---

## Das Schlüsseltresor-Dashboard

Öffne `http://localhost:56243` im Browser, um das Dashboard zu sehen.

**Aufbau der Oberfläche:**
- **Obere Leiste (fixiert)**: Logo, Sprach-/Design-Auswahl, SSE-Verbindungsstatus
- **Kartengitter**: Agenten-, Dienst- und API-Schlüssel-Karten als Kacheln angeordnet

### API-Schlüssel-Karten

Karten, die dir einen Überblick über deine registrierten API-Schlüssel geben.

- Schlüssel werden nach Dienst gruppiert angezeigt.
- `today_usage`: Heute erfolgreich verarbeitete Token (Texteinheiten, die die KI liest/schreibt)
- `today_attempts`: Gesamtanzahl der Aufrufe heute (Erfolg + Fehler)
- Mit `[+ Hinzufügen]` neue Schlüssel registrieren, mit `✕` löschen.

> 💡 **Was ist ein Token?** Die Einheit, in der KI Text verarbeitet. Ungefähr ein englisches Wort oder 1–2 deutsche Buchstaben. API-Kosten werden typischerweise nach Token-Anzahl berechnet.

### Agenten-Karten

Karten, die den Status der mit dem wall-vault-Proxy verbundenen Bots (Agenten) zeigen.

**Der Verbindungsstatus wird in 4 Stufen angezeigt:**

| Anzeige | Status | Bedeutung |
|---------|--------|-----------|
| 🟢 | Läuft | Proxy funktioniert normal |
| 🟡 | Verzögert | Antwortet, aber langsam |
| 🔴 | Offline | Proxy antwortet nicht |
| ⚫ | Nicht verbunden / Deaktiviert | Proxy war noch nie mit dem Tresor verbunden oder ist deaktiviert |

**Buttons am unteren Rand der Agenten-Karten:**

Wenn du beim Registrieren eines Agenten einen **Agententyp** angibst, erscheinen automatisch passende Komfortbuttons.

---

#### 🔘 Konfiguration-Kopieren-Button — erstellt automatisch die Verbindungseinstellungen

Beim Klicken wird ein Konfigurations-Snippet mit dem Token, der Proxy-Adresse und den Modellinformationen des Agenten in die Zwischenablage kopiert. Füge den kopierten Inhalt an der in der Tabelle genannten Stelle ein, um die Verbindung einzurichten.

| Button | Agententyp | Einfügen in |
|--------|-----------|-------------|
| 🦞 OpenClaw-Konfiguration kopieren | `openclaw` | `~/.openclaw/openclaw.json` |
| 🦀 NanoClaw-Konfiguration kopieren | `nanoclaw` | `~/.openclaw/openclaw.json` |
| 🟠 Claude Code-Konfiguration kopieren | `claude-code` | `~/.claude/settings.json` |
| ⌨ Cursor-Konfiguration kopieren | `cursor` | Cursor → Settings → AI |
| 💻 VSCode-Konfiguration kopieren | `vscode` | `~/.continue/config.json` |

**Beispiel — Bei Claude Code-Typ wird Folgendes kopiert:**

```json
// ~/.claude/settings.json
{
  "apiProvider": "openai",
  "baseUrl": "http://192.168.1.20:56244/v1",
  "apiKey": "token-dieses-agenten"
}
```

**Beispiel — Bei VSCode (Continue)-Typ:**

```yaml
# ~/.continue/config.yaml  ← in config.yaml einfügen, NICHT in config.json
name: My Config
version: 0.0.1
schema: v1

models:
  - name: wall-vault proxy
    provider: openai
    model: gemini-2.5-flash
    apiBase: http://192.168.1.20:56244/v1
    apiKey: token-dieses-agenten
    roles:
      - chat
      - edit
      - apply
```

> ⚠️ **Die neueste Version von Continue verwendet `config.yaml`.** Wenn `config.yaml` existiert, wird `config.json` vollständig ignoriert. Unbedingt in `config.yaml` einfügen.

**Beispiel — Bei Cursor-Typ:**

```
Base URL : http://192.168.1.20:56244/v1
API Key  : token-dieses-agenten

// Oder als Umgebungsvariablen:
OPENAI_BASE_URL=http://192.168.1.20:56244/v1
OPENAI_API_KEY=token-dieses-agenten
```

> ⚠️ **Falls das Kopieren in die Zwischenablage nicht funktioniert**: Browser-Sicherheitsrichtlinien können das Kopieren blockieren. Wenn ein Popup mit einer Textbox erscheint, wähle mit Ctrl+A alles aus und kopiere mit Ctrl+C.

---

#### ⚡ Automatisch-Anwenden-Button — ein Klick und die Konfiguration steht

Bei Agenten vom Typ `cline`, `claude-code`, `openclaw` oder `nanoclaw` wird auf der Agenten-Karte ein **⚡ Konfiguration anwenden**-Button angezeigt. Ein Klick aktualisiert automatisch die lokale Konfigurationsdatei des Agenten.

| Button | Agententyp | Zieldatei |
|--------|-----------|-----------|
| ⚡ Cline-Konfiguration anwenden | `cline` | `~/.cline/data/globalState.json` + `secrets.json` |
| ⚡ Claude Code-Konfiguration anwenden | `claude-code` | `~/.claude/settings.json` |
| ⚡ OpenClaw-Konfiguration anwenden | `openclaw` | `~/.openclaw/openclaw.json` |
| ⚡ NanoClaw-Konfiguration anwenden | `nanoclaw` | `~/.openclaw/openclaw.json` |

> ⚠️ Dieser Button sendet eine Anfrage an **localhost:56244** (lokaler Proxy). Der Proxy muss auf dieser Maschine laufen, damit es funktioniert.

---

#### 🔀 Drag & Drop – Karten umsortieren (v0.1.17)

Agenten-Karten im Dashboard können per **Drag & Drop** in eine gewünschte Reihenfolge gebracht werden.

1. Greife eine Agenten-Karte mit der Maus und ziehe sie
2. Lasse sie über einer anderen Karte los, um die Positionen zu tauschen
3. Die neue Reihenfolge wird **sofort auf dem Server gespeichert** und bleibt auch nach dem Neuladen erhalten

> 💡 Touchgeräte (Mobiltelefon/Tablet) werden noch nicht unterstützt. Bitte einen Desktop-Browser verwenden.

---

#### 🔄 Bidirektionale Modellsynchronisierung (v0.1.16)

Wenn du im Dashboard das Modell eines Agenten änderst, wird dessen lokale Konfiguration automatisch aktualisiert.

**Bei Cline:**
- Modelländerung im Tresor → SSE-Event → Proxy aktualisiert das Modellfeld in `globalState.json`
- Aktualisierte Felder: `actModeOpenAiModelId`, `planModeOpenAiModelId`, `openAiModelId`
- `openAiBaseUrl` und API-Schlüssel bleiben unberührt
- **VS Code muss neu geladen werden (`Ctrl+Alt+R` oder `Ctrl+Shift+P` → `Developer: Reload Window`)**
  - Weil Cline die Konfigurationsdatei während der Laufzeit nicht erneut liest

**Bei Claude Code:**
- Modelländerung im Tresor → SSE-Event → Proxy aktualisiert das `model`-Feld in `settings.json`
- Automatische Suche auf WSL- und Windows-Pfaden (`~/.claude/`, `/mnt/c/Users/*/.claude/`)

**Umgekehrte Richtung (Agent → Tresor):**
- Wenn ein Agent (Cline, Claude Code usw.) eine Anfrage an den Proxy sendet, fügt der Proxy die Dienst-/Modellinformationen dieses Clients in den Heartbeat ein
- Die Agenten-Karte im Dashboard zeigt den aktuell verwendeten Dienst/Modell in Echtzeit an

> 💡 **Kernpunkt**: Der Proxy identifiziert Agenten anhand des Authorization-Tokens in den Anfragen und leitet automatisch an den im Tresor konfigurierten Dienst/Modell weiter. Selbst wenn Cline oder Claude Code einen anderen Modellnamen senden, überschreibt der Proxy mit der Tresor-Konfiguration.

---

### Cline mit VS Code verwenden — Detaillierte Anleitung

#### Schritt 1: Cline installieren

Installiere **Cline** (ID: `saoudrizwan.claude-dev`) aus dem VS Code Extensions Marketplace.

#### Schritt 2: Agent im Tresor registrieren

1. Öffne das Tresor-Dashboard (`http://Tresor-IP:56243`)
2. Klicke im Bereich **Agenten** auf **+ Hinzufügen**
3. Fülle folgende Felder aus:

| Feld | Wert | Beschreibung |
|------|------|-------------|
| ID | `mein_cline` | Eindeutiger Bezeichner (alphanumerisch, ohne Leerzeichen) |
| Name | `Mein Cline` | Im Dashboard angezeigter Name |
| Agententyp | `cline` | ← Unbedingt `cline` wählen |
| Dienst | Den zu verwendenden Dienst wählen (z. B. `google`) | |
| Modell | Das zu verwendende Modell eingeben (z. B. `gemini-2.5-flash`) | |

4. Klicke auf **Speichern** — ein Token wird automatisch generiert

#### Schritt 3: Mit Cline verbinden

**Methode A — Automatisch anwenden (empfohlen)**

1. Stelle sicher, dass der wall-vault-**Proxy** auf dieser Maschine läuft (`localhost:56244`)
2. Klicke auf den **⚡ Cline-Konfiguration anwenden**-Button auf der Agenten-Karte
3. Wenn die Meldung „Konfiguration erfolgreich angewendet!" erscheint, hat es funktioniert
4. Lade VS Code neu (`Ctrl+Alt+R`)

**Methode B — Manuelle Einrichtung**

Öffne die Einstellungen (⚙️) in der Cline-Seitenleiste:
- **API Provider**: `OpenAI Compatible`
- **Base URL**: `http://Proxy-Adresse:56244/v1`
  - Gleiche Maschine: `http://localhost:56244/v1`
  - Andere Maschine (z. B. Mac Mini): `http://192.168.1.20:56244/v1`
- **API Key**: Das vom Tresor ausgestellte Token (von der Agenten-Karte kopieren)
- **Model ID**: Das im Tresor konfigurierte Modell (z. B. `gemini-2.5-flash`)

#### Schritt 4: Überprüfen

Sende eine beliebige Nachricht im Cline-Chatfenster. Wenn es funktioniert:
- Die Agenten-Karte im Dashboard zeigt einen **grünen Punkt (● Läuft)** an
- Auf der Karte wird der aktuelle Dienst/Modell angezeigt (z. B. `google / gemini-2.5-flash`)

#### Modell wechseln

Wenn du Clines Modell wechseln möchtest, tue dies im **Tresor-Dashboard**:

1. Ändere das Dienst-/Modell-Dropdown auf der Agenten-Karte
2. Klicke auf **Anwenden**
3. Lade VS Code neu (`Ctrl+Alt+R`) — der Modellname in Clines Fußzeile wird aktualisiert
4. Ab der nächsten Anfrage wird das neue Modell verwendet

> 💡 In der Praxis identifiziert der Proxy Clines Anfragen anhand des Tokens und leitet sie zum Tresor-konfigurierten Modell weiter. Selbst ohne VS Code-Neuladen **wechselt das tatsächlich verwendete Modell sofort** — das Neuladen dient nur dazu, die Modellanzeige in Clines UI zu aktualisieren.

#### Verbindungsabbruch erkennen

Wenn VS Code geschlossen wird, wechselt die Agenten-Karte im Dashboard nach etwa **90 Sekunden** auf Gelb (verzögert) und nach **3 Minuten** auf Rot (offline). (Ab v0.1.18 ist die Offline-Erkennung dank 15-Sekunden-Intervallprüfungen schneller.)

#### Problemlösung

| Symptom | Ursache | Lösung |
|---------|---------|--------|
| „Verbindung fehlgeschlagen"-Fehler in Cline | Proxy läuft nicht oder falsche Adresse | Proxy prüfen mit `curl http://localhost:56244/health` |
| Grüner Punkt erscheint nicht im Tresor | API-Schlüssel (Token) nicht konfiguriert | **⚡ Cline-Konfiguration anwenden**-Button erneut klicken |
| Cline-Fußzeile zeigt altes Modell | Cline hat die Einstellungen gecacht | VS Code neu laden (`Ctrl+Alt+R`) |
| Falscher Modellname angezeigt | Alter Bug (behoben in v0.1.16) | Proxy auf v0.1.16 oder neuer aktualisieren |

---

#### 🟣 Deploy-Befehl-Kopieren-Button — für die Installation auf neuen Maschinen

Verwende diesen Button, wenn du den wall-vault-Proxy erstmals auf einem neuen Computer installierst und mit dem Tresor verbindest. Beim Klicken wird das gesamte Installationsskript kopiert. Füge es ins Terminal des neuen Computers ein und führe es aus, um Folgendes auf einmal zu erledigen:

1. wall-vault-Binary installieren (wird übersprungen, falls bereits installiert)
2. Automatische Registrierung als systemd-Benutzerdienst
3. Dienst starten und automatisch mit dem Tresor verbinden

> 💡 Im Skript sind bereits das Token und die Tresor-Serveradresse dieses Agenten eingetragen, sodass es direkt nach dem Einfügen ohne Änderungen ausgeführt werden kann.

---

### Dienst-Karten

Karten zum Ein-/Ausschalten und Konfigurieren der KI-Dienste.

- Ein-/Aus-Schalter pro Dienst
- Adresse eines lokalen KI-Servers (Ollama, LM Studio, vLLM usw. auf deinem Computer) eingeben, um verfügbare Modelle automatisch zu erkennen
- **Lokaler Dienst-Verbindungsstatus**: Ein ● Punkt neben dem Dienstnamen ist **grün** bei Verbindung, **grau** bei Nicht-Verbindung
- **Automatische Signalanzeige für lokale Dienste** (v0.1.23+): Lokale Dienste (Ollama, LM Studio, vLLM) werden je nach Verbindungsverfügbarkeit automatisch aktiviert/deaktiviert. Wenn ein Dienst erreichbar wird, wechselt der ● Punkt innerhalb von 15 Sekunden auf Grün und das Kontrollkästchen wird aktiviert; wenn der Dienst ausfällt, wird er automatisch deaktiviert. Dies funktioniert genauso wie die automatische Umschaltung von Cloud-Diensten (Google, OpenRouter usw.) basierend auf API-Schlüsselverfügbarkeit.

> 💡 **Falls der lokale Dienst auf einem anderen Computer läuft**: Gib die IP dieses Computers im Dienst-URL-Feld ein. Beispiel: `http://192.168.1.20:11434` (Ollama), `http://192.168.1.20:1234` (LM Studio). Wenn der Dienst nur an `127.0.0.1` statt `0.0.0.0` gebunden ist, funktioniert der Zugriff über eine externe IP nicht — überprüfe die Bindungsadresse in den Diensteinstellungen.

### Admin-Token-Eingabe

Wenn du im Dashboard wichtige Funktionen wie das Hinzufügen oder Löschen von Schlüsseln nutzen möchtest, erscheint ein Admin-Token-Eingabe-Popup. Gib den Token ein, den du beim Setup-Assistenten festgelegt hast. Nach einmaliger Eingabe bleibt er gültig, bis du den Browser schließt.

> ⚠️ **Schlägt die Authentifizierung innerhalb von 15 Minuten mehr als 10 Mal fehl, wird diese IP vorübergehend gesperrt.** Falls du deinen Token vergessen hast, schaue im Feld `admin_token` in der Datei `wall-vault.yaml` nach.

---

## Verteilter Modus (Multi-Bot)

Wenn du OpenClaw auf mehreren Computern gleichzeitig betreibst, kannst du **einen einzigen Schlüsseltresor teilen**. Das ist praktisch, weil die Schlüsselverwaltung nur an einer Stelle stattfindet.

### Beispielkonfiguration

```
[Schlüsseltresor-Server]
  wall-vault vault    (Schlüsseltresor :56243, Dashboard)

[WSL Alpha]           [Raspberry Pi Gamma]    [Mac Mini Lokal]
  wall-vault proxy      wall-vault proxy        wall-vault proxy
  openclaw TUI          openclaw TUI            openclaw TUI
  ↕ SSE-Sync            ↕ SSE-Sync              ↕ SSE-Sync
```

Alle Bots zeigen auf den zentralen Tresor-Server, sodass Modelländerungen oder hinzugefügte Schlüssel sofort auf allen Bots wirksam werden.

### Schritt 1: Schlüsseltresor-Server starten

Auf dem Computer, der als Tresor-Server dienen soll:

```bash
wall-vault vault
```

### Schritt 2: Jeden Bot (Client) registrieren

Registriere die Informationen jedes Bots, der sich mit dem Tresor verbindet:

```bash
curl -X POST http://localhost:56243/admin/clients \
  -H "Authorization: Bearer dein-admin-token" \
  -H "Content-Type: application/json" \
  -d '{
    "id": "botA",
    "name": "Bot A",
    "token": "bota-secret",
    "default_service": "google",
    "default_model": "gemini-2.5-flash"
  }'
```

### Schritt 3: Proxy auf jedem Bot-Computer starten

Auf jedem Computer, auf dem ein Bot installiert ist, starte den Proxy mit der Tresor-Adresse und dem Token:

```bash
WV_VAULT_URL=http://192.168.x.x:56243 \
WV_VAULT_TOKEN=bota-secret \
WV_VAULT_CLIENT_ID=botA \
wall-vault proxy
```

> 💡 Ersetze **`192.168.x.x`** durch die tatsächliche interne IP-Adresse des Tresor-Server-Computers. Du findest sie in den Router-Einstellungen oder über den Befehl `ip addr`.

---

## Autostart einrichten

Wenn es lästig ist, wall-vault bei jedem Neustart manuell zu starten, registriere es als Systemdienst. Einmal registriert, startet es automatisch beim Hochfahren.

### Linux — systemd (die meisten Linux-Distributionen)

systemd ist das System unter Linux, das Programme automatisch startet und verwaltet:

```bash
wall-vault doctor deploy
systemctl --user daemon-reload
systemctl --user enable --now wall-vault
```

Logs anzeigen:

```bash
journalctl --user -u wall-vault -f
```

### macOS — launchd

Das System unter macOS für den automatischen Programmstart:

```bash
wall-vault doctor deploy launchd
launchctl load ~/Library/LaunchAgents/com.wall-vault.plist
```

### Windows — NSSM

1. Lade NSSM von [nssm.cc](https://nssm.cc/download) herunter und füge es zum PATH hinzu.
2. In einer PowerShell mit Administratorrechten:

```powershell
wall-vault doctor deploy windows
```

---

## Doctor-Diagnose

Der `doctor`-Befehl ist ein Tool, das die wall-vault-Konfiguration **selbst diagnostiziert und repariert**.

```bash
wall-vault doctor check   # Aktuellen Zustand diagnostizieren (nur lesen, ändert nichts)
wall-vault doctor fix     # Probleme automatisch reparieren
wall-vault doctor all     # Diagnose + automatische Reparatur in einem Schritt
```

> 💡 Wenn etwas nicht stimmt, versuche zuerst `wall-vault doctor all`. Es erkennt und behebt viele Probleme automatisch.

---

## Umgebungsvariablen – Übersicht

Umgebungsvariablen sind eine Methode, um einem Programm Konfigurationswerte zu übergeben. Gib sie im Terminal mit `export VARIABLE=Wert` ein, oder trage sie in die Autostart-Dienstdatei ein, damit sie dauerhaft wirksam sind.

| Variable | Beschreibung | Beispielwert |
|----------|-------------|--------------|
| `WV_LANG` | Dashboard-Sprache | `ko`, `en`, `ja` |
| `WV_THEME` | Dashboard-Design | `light`, `dark`, `gold` |
| `WV_KEY_GOOGLE` | Google API-Schlüssel (Komma-getrennt für mehrere) | `AIza...,AIza...` |
| `WV_KEY_OPENROUTER` | OpenRouter API-Schlüssel | `sk-or-v1-...` |
| `WV_VAULT_URL` | Tresor-Serveradresse im verteilten Modus | `http://192.168.x.x:56243` |
| `WV_VAULT_TOKEN` | Client (Bot) Auth-Token | `my-secret-token` |
| `WV_ADMIN_TOKEN` | Admin-Token | `admin-token-here` |
| `WV_MASTER_PASS` | API-Schlüssel-Verschlüsselungspasswort | `my-password` |
| `WV_AVATAR` | Avatar-Bilddateipfad (relativ zu `~/.openclaw/`) | `workspace/avatars/avatar.png` |
| `OLLAMA_URL` | Ollama lokale Serveradresse | `http://192.168.x.x:11434` |

---

## Problemlösung

### Proxy startet nicht

Häufig wird der Port bereits von einem anderen Programm verwendet.

```bash
ss -tlnp | grep 56244   # Prüfen, wer Port 56244 nutzt
wall-vault proxy --port 8080   # Mit einem anderen Port starten
```

### API-Schlüssel-Fehler (429, 402, 401, 403, 582)

| Fehlercode | Bedeutung | Abhilfe |
|-----------|-----------|---------|
| **429** | Zu viele Anfragen (Kontingent überschritten) | Warten oder weitere Schlüssel hinzufügen |
| **402** | Zahlung erforderlich oder Guthaben aufgebraucht | Guthaben beim betreffenden Dienst aufladen |
| **401 / 403** | Schlüssel falsch oder keine Berechtigung | Schlüsselwert überprüfen und neu registrieren |
| **582** | Gateway-Überlastung (5-Minuten-Cooldown) | Wird nach 5 Minuten automatisch freigegeben |

```bash
# Registrierte Schlüsselliste und Status anzeigen
curl -H "Authorization: Bearer dein-admin-token" http://localhost:56243/admin/keys

# Schlüssel-Nutzungszähler zurücksetzen
curl -X POST -H "Authorization: Bearer dein-admin-token" http://localhost:56243/admin/keys/reset
```

### Agent wird als „Nicht verbunden" angezeigt

„Nicht verbunden" bedeutet, dass der Proxy-Prozess kein Signal (Heartbeat) an den Tresor sendet. **Es bedeutet nicht, dass die Konfiguration nicht gespeichert wurde.** Der Proxy muss mit der Tresor-Adresse und dem Token gestartet werden, um eine Verbindung herzustellen.

```bash
# Proxy mit Tresor-Adresse, Token und Client-ID starten
WV_VAULT_URL=http://Tresor-Server:56243 \
WV_VAULT_TOKEN=Client-Token \
WV_VAULT_CLIENT_ID=Client-ID \
wall-vault proxy
```

Nach erfolgreicher Verbindung wechselt das Dashboard innerhalb von etwa 20 Sekunden auf 🟢 Läuft.

### Ollama-Verbindungsprobleme

Ollama ist ein Programm, das KI direkt auf deinem Computer ausführt. Stelle zunächst sicher, dass Ollama läuft.

```bash
curl http://localhost:11434/api/tags   # Wenn eine Modellliste erscheint, funktioniert es
export OLLAMA_URL=http://192.168.x.x:11434   # Falls auf einem anderen Computer
```

> ⚠️ Wenn Ollama nicht antwortet, starte es zuerst mit `ollama serve`.

> ⚠️ **Große Modelle antworten langsam**: Große Modelle wie `qwen3.5:35b` oder `deepseek-r1` können mehrere Minuten zur Antwortgenerierung benötigen. Auch wenn es so aussieht, als würde nichts passieren, wird möglicherweise noch verarbeitet — bitte habe Geduld.

---

## Aktuelle Änderungen (v0.1.16 ~ v0.1.23)

### v0.1.23 (2026-04-06)
- **Ollama-Modellwechsel behoben**: Ein Problem wurde behoben, bei dem das Ändern des Ollama-Modells im Tresor-Dashboard nicht im tatsächlichen Proxy wirksam wurde. Zuvor wurde nur die Umgebungsvariable (`OLLAMA_MODEL`) verwendet, jetzt haben Tresor-Einstellungen Vorrang.
- **Automatische Signalanzeige für lokale Dienste**: Ollama, LM Studio und vLLM werden bei Erreichbarkeit automatisch aktiviert und bei Nichterreichbarkeit deaktiviert. Funktioniert genauso wie die schlüsselbasierte automatische Umschaltung für Cloud-Dienste.

### v0.1.22 (2026-04-05)
- **Leeres Content-Feld behoben**: Wenn Thinking-Modelle (gemini-3.1-pro, o1, claude thinking usw.) alle max_tokens für Reasoning verbrauchen und keine tatsächliche Antwort erzeugen können, hat der Proxy die `content`/`text`-Felder aus der JSON-Antwort via `omitempty` weggelassen, wodurch OpenAI/Anthropic-SDK-Clients mit `Cannot read properties of undefined (reading 'trim')` abstürzten. Behoben, sodass Felder gemäß offizieller API-Spezifikation immer enthalten sind.

### v0.1.21 (2026-04-05)
- **Gemma 4-Modellunterstützung**: Gemma-Familienmodelle wie `gemma-4-31b-it` und `gemma-4-26b-a4b-it` können jetzt über die Google Gemini API verwendet werden.
- **LM Studio / vLLM-Dienstunterstützung**: Zuvor fehlten diese Dienste im Proxy-Routing und fielen immer auf Ollama zurück. Jetzt korrekt über OpenAI-kompatible API geroutet.
- **Dashboard-Dienstanzeige behoben**: Auch bei Fallback zeigt das Dashboard immer den vom Benutzer konfigurierten Dienst an.
- **Lokaler Dienststatus**: Beim Laden des Dashboards wird der Verbindungsstatus lokaler Dienste (Ollama, LM Studio, vLLM usw.) durch die Farbe des ● Punkts angezeigt.
- **Tool-Filter-Umgebungsvariable**: Mit der Umgebungsvariable `WV_TOOL_FILTER=passthrough` kann der Tool-Durchleitungsmodus eingestellt werden.

### v0.1.20 (2026-03-28)
- **Umfassende Sicherheitshärtung**: XSS-Schutz (41 Stellen), konstante Token-Vergleichszeit, CORS-Beschränkungen, Anfragegrößenlimits, Pfad-Traversal-Schutz, SSE-Authentifizierung, Rate-Limiter-Härtung und 12 weitere Sicherheitsverbesserungen.

### v0.1.19 (2026-03-27)
- **Claude Code-Online-Erkennung**: Claude-Code-Instanzen, die nicht über den Proxy gehen, werden jetzt im Dashboard als online angezeigt.

### v0.1.18 (2026-03-26)
- **Fallback-Dienst-Klebeeffekt behoben**: Nach einem temporären Fehler mit Ollama-Fallback wird automatisch zum ursprünglichen Dienst zurückgewechselt, sobald dieser wiederhergestellt ist.
- **Verbesserte Offline-Erkennung**: 15-Sekunden-Intervallprüfungen ermöglichen eine schnellere Erkennung von Proxy-Ausfällen.

### v0.1.17 (2026-03-25)
- **Drag & Drop-Kartenordnung**: Agenten-Karten können per Drag & Drop umsortiert werden.
- **Inline-Konfigurationsanwendungs-Button**: Der [⚡ Konfiguration anwenden]-Button wird auf Offline-Agenten-Karten angezeigt.
- **Agententyp cokacdir hinzugefügt**.

### v0.1.16 (2026-03-25)
- **Bidirektionale Modellsynchronisierung**: Modelländerungen für Cline oder Claude Code im Tresor-Dashboard werden automatisch übernommen.

---

*Detaillierte API-Informationen findest du in [API.md](API.md).*
