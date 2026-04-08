# wall-vault Benutzerhandbuch
*(Zuletzt aktualisiert: 2026-04-08 — v0.1.25)*

---

## Inhaltsverzeichnis

1. [Was ist wall-vault?](#was-ist-wall-vault)
2. [Installation](#installation)
3. [Erste Schritte (Setup-Assistent)](#erste-schritte)
4. [API-Schlüssel registrieren](#api-schlüssel-registrieren)
5. [Proxy-Nutzung](#proxy-nutzung)
6. [Schlüsseltresor-Dashboard](#schlüsseltresor-dashboard)
7. [Verteilter Modus (Multi-Bot)](#verteilter-modus-multi-bot)
8. [Autostart-Konfiguration](#autostart-konfiguration)
9. [Doctor (Diagnosetool)](#doctor-diagnosetool)
10. [RTK Token-Einsparung](#rtk-token-einsparung)
11. [Umgebungsvariablen-Referenz](#umgebungsvariablen-referenz)
12. [Fehlerbehebung](#fehlerbehebung)

---

## Was ist wall-vault?

**wall-vault = AI-Proxy + API-Schlüsseltresor für OpenClaw**

Um AI-Dienste zu nutzen, benötigen Sie **API-Schlüssel**. Ein API-Schlüssel ist wie ein **digitaler Ausweis**, der beweist: „Diese Person ist berechtigt, diesen Dienst zu nutzen." Allerdings haben diese Ausweise ein tägliches Nutzungslimit, und bei unsachgemäßer Verwaltung besteht ein Risiko der Offenlegung.

wall-vault verwahrt diese Ausweise in einem sicheren Tresor und fungiert als **Proxy (Stellvertreter)** zwischen OpenClaw und AI-Diensten. Einfach gesagt: OpenClaw muss sich nur mit wall-vault verbinden, und wall-vault erledigt den Rest automatisch.

Probleme, die wall-vault löst:

- **Automatische API-Schlüssel-Rotation**: Wenn ein Schlüssel sein Nutzungslimit erreicht oder vorübergehend blockiert wird (Cooldown), wechselt es lautlos zum nächsten Schlüssel. OpenClaw arbeitet ohne Unterbrechung weiter.
- **Automatischer Service-Fallback**: Wenn Google nicht antwortet, wird auf OpenRouter gewechselt; wenn das auch nicht geht, wird auf lokal installiertes Ollama, LM Studio oder vLLM (lokale AI) umgeschaltet. Sitzungen brechen nicht ab. Wenn der ursprüngliche Dienst sich erholt, wird ab der nächsten Anfrage automatisch zurückgewechselt (v0.1.18+, LM Studio/vLLM: v0.1.21+).
- **Echtzeit-Synchronisation (SSE)**: Wenn Sie ein Modell im Tresor-Dashboard ändern, wird dies innerhalb von 1-3 Sekunden auf dem OpenClaw-Bildschirm angezeigt. SSE (Server-Sent Events) ist eine Technologie, bei der der Server Änderungen in Echtzeit an Clients übermittelt.
- **Echtzeit-Benachrichtigungen**: Ereignisse wie Schlüsselerschöpfung oder Dienstausfälle werden sofort am unteren Rand des OpenClaw TUI (Terminalbildschirm) angezeigt.

> 💡 **Claude Code, Cursor und VS Code** können ebenfalls verbunden werden, aber der eigentliche Zweck von wall-vault ist die Verwendung mit OpenClaw.

```
OpenClaw (TUI-Terminalbildschirm)
        │
        ▼
  wall-vault Proxy (:56244)   ← Schlüsselverwaltung, Routing, Fallback, Events
        │
        ├─ Google Gemini API
        ├─ OpenRouter API (340+ Modelle)
        ├─ Ollama / LM Studio / vLLM (lokaler PC, letzte Instanz)
        └─ OpenAI / Anthropic API
```

---

## Installation

### Linux / macOS

Öffnen Sie ein Terminal und fügen Sie die folgenden Befehle ein.

```bash
# Linux (normaler PC, Server — amd64)
curl -L https://github.com/sookmook/wall-vault/releases/latest/download/wall-vault-linux-amd64 \
  -o ~/.local/bin/wall-vault && chmod +x ~/.local/bin/wall-vault

# macOS Apple Silicon (M1/M2/M3 Mac)
curl -L https://github.com/sookmook/wall-vault/releases/latest/download/wall-vault-darwin-arm64 \
  -o /usr/local/bin/wall-vault && chmod +x /usr/local/bin/wall-vault
```

- `curl -L ...` — Lädt eine Datei aus dem Internet herunter.
- `chmod +x` — Macht die heruntergeladene Datei „ausführbar". Wenn Sie diesen Schritt überspringen, erhalten Sie einen „Berechtigung verweigert"-Fehler.

### Windows

Öffnen Sie PowerShell (als Administrator) und führen Sie die folgenden Befehle aus.

```powershell
# Download
Invoke-WebRequest -Uri `
  "https://github.com/sookmook/wall-vault/releases/latest/download/wall-vault-windows-amd64.exe" `
  -OutFile "$env:LOCALAPPDATA\Programs\wall-vault\wall-vault.exe"

# Zum PATH hinzufügen (gilt nach PowerShell-Neustart)
$env:PATH += ";$env:LOCALAPPDATA\Programs\wall-vault"
```

> 💡 **Was ist PATH?** Es ist die Liste der Ordner, in denen Ihr Computer nach Befehlen sucht. Sie müssen wall-vault zum PATH hinzufügen, damit Sie `wall-vault` von jedem Ordner aus ausführen können.

### Aus dem Quellcode bauen (für Entwickler)

Nur relevant, wenn eine Go-Entwicklungsumgebung installiert ist.

```bash
git clone https://github.com/sookmook/wall-vault
cd wall-vault
make build       # bin/wall-vault (Version: v0.1.25.YYYYMMDD.HHmmss)
make install     # ~/.local/bin/wall-vault
```

> 💡 **Build-Zeitstempel-Version**: Beim Bauen mit `make build` wird die Version automatisch in einem Format wie `v0.1.25.20260408.022325` mit Datum und Uhrzeit generiert. Beim direkten Bauen mit `go build ./...` wird die Version nur als `"dev"` angezeigt.

---

## Erste Schritte

### Setup-Assistent ausführen

Nach der Installation müssen Sie zunächst den **Setup-Assistenten** mit dem folgenden Befehl ausführen. Der Assistent führt Sie Schritt für Schritt durch die erforderlichen Einstellungen.

```bash
wall-vault setup
```

Der Assistent führt durch folgende Schritte:

```
1. Sprachauswahl (10 Sprachen einschließlich Koreanisch)
2. Theme-Auswahl (light / dark / gold / cherry / ocean)
3. Betriebsmodus — Einzelnutzung (standalone) oder gemeinsame Nutzung (distributed)
4. Bot-Name — der im Dashboard angezeigte Name
5. Port-Einstellungen — Standard: Proxy 56244, Tresor 56243 (Enter für Standard)
6. AI-Dienst-Auswahl — Google / OpenRouter / Ollama / LM Studio / vLLM
7. Tool-Sicherheitsfilter-Einstellungen
8. Admin-Token — ein Passwort zum Sperren der Dashboard-Verwaltung. Automatische Generierung möglich
9. API-Schlüssel-Verschlüsselungspasswort — für extra-sichere Schlüsselspeicherung (optional)
10. Speicherort der Konfigurationsdatei
```

> ⚠️ **Merken Sie sich unbedingt Ihren Admin-Token.** Sie benötigen ihn später zum Hinzufügen von Schlüsseln oder Ändern von Einstellungen im Dashboard. Wenn Sie ihn verlieren, müssen Sie die Konfigurationsdatei direkt bearbeiten.

Nach Abschluss des Assistenten wird die Konfigurationsdatei `wall-vault.yaml` automatisch erstellt.

### Ausführung

```bash
wall-vault start
```

Zwei Server starten gleichzeitig:

- **Proxy** (`http://localhost:56244`) — der Vermittler zwischen OpenClaw und AI-Diensten
- **Schlüsseltresor** (`http://localhost:56243`) — API-Schlüsselverwaltung und Web-Dashboard

Öffnen Sie `http://localhost:56243` in Ihrem Browser, um das Dashboard aufzurufen.

---

## API-Schlüssel registrieren

Es gibt vier Möglichkeiten, API-Schlüssel zu registrieren. **Methode 1 (Umgebungsvariablen) wird für Einsteiger empfohlen.**

### Methode 1: Umgebungsvariablen (empfohlen — am einfachsten)

Umgebungsvariablen sind **voreingestellte Werte**, die Programme beim Start lesen. Geben Sie sie im Terminal wie folgt ein:

```bash
# Google Gemini-Schlüssel registrieren
export WV_KEY_GOOGLE=AIzaSy...

# OpenRouter-Schlüssel registrieren
export WV_KEY_OPENROUTER=sk-or-v1-...

# Nach der Registrierung starten
wall-vault start
```

Wenn Sie mehrere Schlüssel haben, verbinden Sie sie mit Kommas. wall-vault rotiert automatisch durch die Schlüssel (Round-Robin):

```bash
export WV_KEY_GOOGLE=AIzaSy...,AIzaSy...,AIzaSy...
```

> 💡 **Tipp**: Der `export`-Befehl gilt nur für die aktuelle Terminalsitzung. Um ihn über Neustarts hinweg beizubehalten, fügen Sie die Zeilen in Ihre `~/.bashrc`- oder `~/.zshrc`-Datei ein.

### Methode 2: Dashboard-UI (per Mausklick)

1. Öffnen Sie `http://localhost:56243` im Browser
2. Klicken Sie auf `[+ Hinzufügen]` in der **🔑 API-Schlüssel**-Karte oben
3. Geben Sie Diensttyp, Schlüsselwert, Label (beschreibender Name) und Tageslimit ein und speichern Sie

### Methode 3: REST API (für Automatisierung/Skripte)

REST API ist eine Methode für Programme, Daten über HTTP auszutauschen. Nützlich für automatisierte Registrierung per Skript.

```bash
curl -X POST http://localhost:56243/admin/keys \
  -H "Authorization: Bearer IHR_ADMIN_TOKEN" \
  -H "Content-Type: application/json" \
  -d '{
    "service": "google",
    "key": "AIzaSy...",
    "label": "Hauptschlüssel",
    "daily_limit": 1000
  }'
```

### Methode 4: Proxy-Flags (für schnelle Tests)

Für temporäre Tests ohne formale Registrierung. Schlüssel gehen beim Beenden des Programms verloren.

```bash
wall-vault proxy --key-google=AIzaSy... --key-openrouter=sk-or-...
```

---

## Proxy-Nutzung

### Verwendung mit OpenClaw (Hauptzweck)

So konfigurieren Sie OpenClaw für die Verbindung mit AI-Diensten über wall-vault.

Öffnen Sie `~/.openclaw/openclaw.json` und fügen Sie Folgendes hinzu:

```json5
// ~/.openclaw/openclaw.json
{
  models: {
    providers: {
      "wall-vault": {
        baseUrl: "http://localhost:56244/v1",
        apiKey: "your-agent-token",   // Vault-Agenten-Token
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

> 💡 **Einfachere Methode**: Klicken Sie auf die Schaltfläche **🦞 OpenClaw-Konfiguration kopieren** auf der Dashboard-Agentenkarte. Ein Snippet mit vorausgefülltem Token und Adresse wird in die Zwischenablage kopiert. Einfach einfügen.

**Wohin leitet das `wall-vault/`-Präfix im Modellnamen?**

wall-vault bestimmt automatisch anhand des Modellnamens, an welchen AI-Dienst Anfragen gesendet werden:

| Modellformat | Zieldienst |
|-------------|-----------|
| `wall-vault/gemini-*` | Direkt zu Google Gemini |
| `wall-vault/gpt-*`, `wall-vault/o3`, `wall-vault/o4*` | Direkt zu OpenAI |
| `wall-vault/claude-*` | Anthropic über OpenRouter |
| `wall-vault/hunter-alpha`, `wall-vault/healer-alpha` | OpenRouter (kostenloser 1M-Token-Kontext) |
| `wall-vault/kimi-*`, `wall-vault/glm-*`, `wall-vault/deepseek-*` | OpenRouter |
| `google/Modell`, `openai/Modell`, `anthropic/Modell` usw. | Direkt zum entsprechenden Dienst |
| `custom/google/Modell`, `custom/openai/Modell` usw. | Entfernt `custom/`-Präfix und leitet um |
| `Modell:cloud` | Entfernt `:cloud`-Suffix und leitet zu OpenRouter |

> 💡 **Was ist Kontext?** Die Menge an Konversation, die eine AI auf einmal erinnern kann. 1M (eine Million Token) bedeutet, dass sehr lange Gespräche oder Dokumente in einem Durchgang verarbeitet werden können.

### Direkte Gemini-API-Format-Verbindung (für bestehende Tool-Kompatibilität)

Wenn Sie Tools haben, die die Google Gemini API direkt verwendeten, ändern Sie einfach die Adresse auf wall-vault:

```bash
export ANTHROPIC_BASE_URL=http://localhost:56244/google
```

Oder wenn das Tool URLs direkt angibt:

```
http://localhost:56244/google/v1beta/models/gemini-2.5-flash:generateContent
```

### Verwendung mit OpenAI SDK (Python)

Sie können wall-vault auch aus Python-Code verbinden, der AI nutzt. Ändern Sie einfach die `base_url`:

```python
from openai import OpenAI

client = OpenAI(
    base_url="http://localhost:56244/v1",
    api_key="not-needed"  # API-Schlüssel werden von wall-vault verwaltet
)

response = client.chat.completions.create(
    model="google/gemini-2.5-flash",   # im provider/model-Format eingeben
    messages=[{"role": "user", "content": "Hallo"}]
)
```

### Modell während der Laufzeit ändern

Um das AI-Modell zu ändern, während wall-vault bereits läuft:

```bash
# Modell durch direkte Proxy-Anfrage ändern
curl -X PUT http://localhost:56244/api/config/model \
  -H "Content-Type: application/json" \
  -d '{"service": "openrouter", "model": "anthropic/claude-3.5-sonnet"}'

# Im verteilten Modus (Multi-Bot) vom Tresor-Server ändern → sofortige SSE-Synchronisation
curl -X PUT http://localhost:56243/admin/clients/mein-bot-id \
  -H "Authorization: Bearer IHR_ADMIN_TOKEN" \
  -H "Content-Type: application/json" \
  -d '{"default_service": "google", "default_model": "gemini-2.5-pro"}'
```

### Verfügbare Modelle anzeigen

```bash
# Vollständige Liste anzeigen
curl http://localhost:56244/api/models | python3 -m json.tool

# Nur Google-Modelle anzeigen
curl "http://localhost:56244/api/models?service=google"

# Nach Namen suchen (z.B. Modelle mit "claude")
curl "http://localhost:56244/api/models?q=claude"
```

**Wichtigste Modelle nach Dienst:**

| Dienst | Wichtigste Modelle |
|--------|-------------------|
| Google | gemini-2.5-pro, gemini-2.5-flash, gemini-2.5-flash-8b, gemini-2.0-flash |
| OpenAI | gpt-4o, gpt-4o-mini, o3, o1, o1-mini |
| OpenRouter | 346+ (Hunter Alpha 1M-Kontext kostenlos, DeepSeek R1/V3, Qwen 2.5 usw.) |
| Ollama | Automatische Erkennung der Modelle auf Ihrem lokalen Server |
| LM Studio | Lokaler Server (Port 1234) |
| vLLM | Lokaler Server (Port 8000) |

---

## Schlüsseltresor-Dashboard

Öffnen Sie `http://localhost:56243` in Ihrem Browser, um das Dashboard aufzurufen.

**Layout:**
- **Obere Leiste (fixiert)**: Logo, Sprach-/Theme-Auswahl, SSE-Verbindungsstatus
- **Karten-Raster**: Agenten-, Dienst- und API-Schlüssel-Karten in Kachelansicht

### API-Schlüssel-Karte

Eine Karte zur Verwaltung aller registrierten API-Schlüssel auf einen Blick.

- Zeigt Schlüssellisten nach Dienst sortiert an.
- `today_usage`: Anzahl der heute erfolgreich verarbeiteten Token (von der AI gelesene/geschriebene Zeichen)
- `today_attempts`: Gesamtzahl der heutigen Aufrufe (einschließlich Erfolge und Fehlschläge)
- `[+ Hinzufügen]`-Schaltfläche zum Registrieren neuer Schlüssel, `✕` zum Löschen.

> 💡 **Was ist ein Token?** Eine Einheit, die AI zur Textverarbeitung verwendet. Ein Token entspricht ungefähr einem englischen Wort oder 1-2 deutschen Zeichen. API-Preise werden normalerweise basierend auf der Token-Anzahl berechnet.

### Agenten-Karte

Eine Karte, die den Status der mit dem wall-vault-Proxy verbundenen Bots (Agenten) anzeigt.

**Verbindungsstatus wird in 4 Stufen angezeigt:**

| Anzeige | Status | Bedeutung |
|---------|--------|-----------|
| 🟢 | Aktiv | Proxy arbeitet normal |
| 🟡 | Verzögert | Antwortet, aber langsam |
| 🔴 | Offline | Proxy antwortet nicht |
| ⚫ | Nicht verbunden / Deaktiviert | Proxy hat sich nie mit dem Tresor verbunden oder ist deaktiviert |

**Schaltflächen am unteren Rand der Agenten-Karte:**

Wenn Sie bei der Registrierung eines Agenten den **Agententyp** angeben, erscheinen automatisch passende Komfortschaltflächen.

---

#### 🔘 Konfiguration-kopieren-Schaltfläche — erstellt automatisch Verbindungseinstellungen

Beim Klicken wird ein Konfigurationssnippet mit vorausgefülltem Token, Proxy-Adresse und Modellinformationen des Agenten in die Zwischenablage kopiert. Fügen Sie den kopierten Inhalt einfach an der in der Tabelle angegebenen Stelle ein, um die Verbindungseinrichtung abzuschließen.

| Schaltfläche | Agententyp | Einfügeort |
|-------------|-----------|-----------|
| 🦞 OpenClaw-Konfiguration kopieren | `openclaw` | `~/.openclaw/openclaw.json` |
| 🦀 NanoClaw-Konfiguration kopieren | `nanoclaw` | `~/.openclaw/openclaw.json` |
| 🟠 Claude Code-Konfiguration kopieren | `claude-code` | `~/.claude/settings.json` |
| ⌨ Cursor-Konfiguration kopieren | `cursor` | Cursor → Settings → AI |
| 💻 VSCode-Konfiguration kopieren | `vscode` | `~/.continue/config.json` |

**Beispiel — für Claude Code-Typ wird folgender Inhalt kopiert:**

```json
// ~/.claude/settings.json
{
  "apiProvider": "openai",
  "baseUrl": "http://192.168.0.6:56244/v1",
  "apiKey": "Token-dieses-Agenten"
}
```

**Beispiel — für VSCode (Continue)-Typ:**

```yaml
# ~/.continue/config.yaml  ← in config.yaml einfügen, nicht config.json
name: My Config
version: 0.0.1
schema: v1

models:
  - name: wall-vault proxy
    provider: openai
    model: gemini-2.5-flash
    apiBase: http://192.168.0.6:56244/v1
    apiKey: Token-dieses-Agenten
    roles:
      - chat
      - edit
      - apply
```

> ⚠️ **Die neueste Version von Continue verwendet `config.yaml`.** Wenn `config.yaml` existiert, wird `config.json` vollständig ignoriert. Stellen Sie sicher, dass Sie in `config.yaml` einfügen.

**Beispiel — für Cursor-Typ:**

```
Base URL : http://192.168.0.6:56244/v1
API Key  : Token-dieses-Agenten

// Oder Umgebungsvariablen:
OPENAI_BASE_URL=http://192.168.0.6:56244/v1
OPENAI_API_KEY=Token-dieses-Agenten
```

> ⚠️ **Wenn das Kopieren in die Zwischenablage nicht funktioniert**: Browser-Sicherheitsrichtlinien können das Kopieren blockieren. Wenn ein Popup-Textfeld erscheint, verwenden Sie Strg+A zum Auswählen und Strg+C zum Kopieren.

---

#### ⚡ Auto-Anwenden-Schaltfläche — ein Klick für die fertige Einrichtung

Für die Agententypen `cline`, `claude-code`, `openclaw` oder `nanoclaw` erscheint eine **⚡ Konfiguration anwenden**-Schaltfläche auf der Agentenkarte. Ein Klick aktualisiert automatisch die lokale Konfigurationsdatei des Agenten.

| Schaltfläche | Agententyp | Zieldatei |
|-------------|-----------|----------|
| ⚡ Cline-Konfiguration anwenden | `cline` | `~/.cline/data/globalState.json` + `secrets.json` |
| ⚡ Claude Code-Konfiguration anwenden | `claude-code` | `~/.claude/settings.json` |
| ⚡ OpenClaw-Konfiguration anwenden | `openclaw` | `~/.openclaw/openclaw.json` |
| ⚡ NanoClaw-Konfiguration anwenden | `nanoclaw` | `~/.openclaw/openclaw.json` |

> ⚠️ Diese Schaltfläche sendet Anfragen an **localhost:56244** (lokaler Proxy). Der Proxy muss auf diesem Rechner laufen, damit es funktioniert.

---

#### 🔀 Drag-and-Drop-Kartensortierung (v0.1.17, verbessert v0.1.25)

Sie können Dashboard-Agentenkarten per **Drag-and-Drop** in Ihre bevorzugte Reihenfolge bringen.

1. Greifen Sie den **Ampel-Bereich (●)** oben links auf der Karte mit der Maus und ziehen Sie
2. Lassen Sie sie auf einer anderen Karte fallen, um die Positionen zu tauschen

> 💡 Der Karteninhalt (Eingabefelder, Schaltflächen usw.) ist nicht ziehbar. Sie können nur vom Ampel-Bereich aus greifen.

#### 🟠 Agenten-Prozess-Erkennung (v0.1.25)

Wenn der Proxy normal läuft, aber der lokale Agentenprozess (NanoClaw, OpenClaw) gestorben ist, wird die Kartenampel **orange (blinkend)** und zeigt eine „Agentenprozess gestoppt"-Meldung.

- 🟢 Grün: Proxy + Agent beide normal
- 🟠 Orange (blinkend): Proxy normal, Agent gestorben
- 🔴 Rot: Proxy offline
3. Die geänderte Reihenfolge wird **sofort auf dem Server gespeichert** und bleibt nach dem Aktualisieren erhalten

> 💡 Touch-Geräte (Mobiltelefone/Tablets) werden noch nicht unterstützt. Bitte verwenden Sie einen Desktop-Browser.

---

#### 🔄 Bidirektionale Modell-Synchronisation (v0.1.16)

Wenn Sie das Modell eines Agenten im Tresor-Dashboard ändern, wird die lokale Konfiguration des Agenten automatisch aktualisiert.

**Für Cline:**
- Modell im Tresor ändern → SSE-Event → Proxy aktualisiert Modellfelder in `globalState.json`
- Aktualisierte Felder: `actModeOpenAiModelId`, `planModeOpenAiModelId`, `openAiModelId`
- `openAiBaseUrl` und API-Schlüssel werden nicht berührt
- **VS Code-Reload (`Strg+Alt+R` oder `Strg+Umschalt+P` → `Developer: Reload Window`) ist erforderlich**
  - Da Cline die Konfigurationsdateien während der Ausführung nicht erneut liest

**Für Claude Code:**
- Modell im Tresor ändern → SSE-Event → Proxy aktualisiert das `model`-Feld in `settings.json`
- Durchsucht automatisch sowohl WSL- als auch Windows-Pfade (`~/.claude/`, `/mnt/c/Users/*/.claude/`)

**Umgekehrte Richtung (Agent → Tresor):**
- Wenn ein Agent (Cline, Claude Code usw.) eine Anfrage an den Proxy sendet, enthält der Proxy die Dienst-/Modellinformationen dieses Clients im Heartbeat
- Die Agentenkarte im Tresor-Dashboard zeigt den aktuell verwendeten Dienst/das Modell in Echtzeit an

> 💡 **Kernpunkt**: Der Proxy identifiziert Agenten anhand des Authorization-Tokens in Anfragen und leitet automatisch zum im Tresor konfigurierten Dienst/Modell weiter. Selbst wenn Cline oder Claude Code einen anderen Modellnamen senden, überschreibt der Proxy ihn mit der Tresor-Konfiguration.

---

### Cline in VS Code verwenden — Detaillierte Anleitung

#### Schritt 1: Cline installieren

Installieren Sie **Cline** (ID: `saoudrizwan.claude-dev`) vom VS Code-Erweiterungsmarktplatz.

#### Schritt 2: Agent im Tresor registrieren

1. Öffnen Sie das Tresor-Dashboard (`http://TRESOR_IP:56243`)
2. Klicken Sie auf **+ Hinzufügen** im Abschnitt **Agenten**
3. Geben Sie Folgendes ein:

| Feld | Wert | Beschreibung |
|------|------|-------------|
| ID | `mein_cline` | Eindeutiger Bezeichner (alphanumerisch, ohne Leerzeichen) |
| Name | `Mein Cline` | Im Dashboard angezeigter Name |
| Agententyp | `cline` | ← muss `cline` ausgewählt werden |
| Dienst | Gewünschten Dienst wählen (z.B. `google`) | |
| Modell | Gewünschtes Modell eingeben (z.B. `gemini-2.5-flash`) | |

4. Klicken Sie auf **Speichern** — ein Token wird automatisch generiert

#### Schritt 3: Mit Cline verbinden

**Methode A — Auto-Anwenden (empfohlen)**

1. Stellen Sie sicher, dass der wall-vault-**Proxy** auf diesem Rechner läuft (`localhost:56244`)
2. Klicken Sie auf die Schaltfläche **⚡ Cline-Konfiguration anwenden** auf der Agentenkarte
3. Erfolgreich bei Anzeige der Benachrichtigung „Konfiguration angewendet!"
4. VS Code neu laden (`Strg+Alt+R`)

**Methode B — Manuelle Einrichtung**

Öffnen Sie die Einstellungen (⚙️) in der Cline-Seitenleiste:
- **API Provider**: `OpenAI Compatible`
- **Base URL**: `http://PROXY_ADRESSE:56244/v1`
  - Gleicher Rechner: `http://localhost:56244/v1`
  - Anderer Rechner (z.B. Mac Mini): `http://192.168.0.6:56244/v1`
- **API Key**: Im Tresor ausgestellter Token (von Agentenkarte kopieren)
- **Model ID**: Im Tresor eingestelltes Modell (z.B. `gemini-2.5-flash`)

#### Schritt 4: Überprüfen

Senden Sie eine beliebige Nachricht im Cline-Chat. Bei Erfolg:
- Die Agentenkarte im Tresor-Dashboard zeigt einen **grünen Punkt (● Aktiv)**
- Die Karte zeigt den aktuellen Dienst/das Modell (z.B. `google / gemini-2.5-flash`)

#### Modell ändern

Wenn Sie Clines Modell ändern möchten, ändern Sie es im **Tresor-Dashboard**:

1. Ändern Sie das Dienst-/Modell-Dropdown auf der Agentenkarte
2. Klicken Sie auf **Anwenden**
3. VS Code neu laden (`Strg+Alt+R`) — der Modellname in der Cline-Fußzeile wird aktualisiert
4. Ab der nächsten Anfrage wird das neue Modell verwendet

> 💡 In der Praxis identifiziert der Proxy Clines Anfragen anhand des Tokens und leitet zum im Tresor konfigurierten Modell weiter. Auch ohne VS Code-Reload **ändert sich das tatsächlich verwendete Modell sofort** — der Reload dient nur der Aktualisierung der Modellanzeige in der Cline-Benutzeroberfläche.

#### Trennungserkennung

Wenn Sie VS Code schließen, wird die Agentenkarte im Tresor-Dashboard nach etwa **90 Sekunden** gelb (verzögert), dann nach **3 Minuten** rot (offline). (Ab v0.1.18 sorgen 15-Sekunden-Statusprüfungen für schnellere Offline-Erkennung.)

#### Fehlerbehebung

| Symptom | Ursache | Lösung |
|---------|---------|--------|
| „Verbindung fehlgeschlagen"-Fehler in Cline | Proxy nicht gestartet oder falsche Adresse | Proxy prüfen mit `curl http://localhost:56244/health` |
| Grüner Punkt erscheint nicht im Tresor | API-Schlüssel (Token) nicht konfiguriert | **⚡ Cline-Konfiguration anwenden** erneut klicken |
| Cline-Fußzeilen-Modell ändert sich nicht | Cline cachet die Konfiguration | VS Code neu laden (`Strg+Alt+R`) |
| Falscher Modellname wird angezeigt | Alter Bug (in v0.1.16 behoben) | Proxy auf v0.1.16+ aktualisieren |

---

#### 🟣 Deploy-Befehl-kopieren-Schaltfläche — für die Installation auf neuen Rechnern

Wird verwendet, wenn Sie den wall-vault-Proxy erstmals auf einem neuen Computer installieren und mit dem Tresor verbinden. Durch Klicken wird das gesamte Installationsskript kopiert. Fügen Sie es in das Terminal des neuen Computers ein und führen Sie es aus — Folgendes wird in einem Schritt erledigt:

1. wall-vault-Binary installieren (übersprungen wenn bereits installiert)
2. systemd-Benutzerdienst automatisch registrieren
3. Dienst starten und automatisch mit dem Tresor verbinden

> 💡 Das Skript enthält bereits das Token und die Tresor-Server-Adresse dieses Agenten, sodass Sie es nach dem Einfügen sofort ohne Änderungen ausführen können.

---

### Dienst-Karte

Eine Karte zum Ein-/Ausschalten und Konfigurieren von AI-Diensten.

- Umschalter zum Aktivieren/Deaktivieren einzelner Dienste
- Geben Sie die Adresse eines lokalen AI-Servers (Ollama, LM Studio, vLLM usw. auf Ihrem Computer) ein, um verfügbare Modelle automatisch zu erkennen.
- **Lokaler Dienst-Verbindungsstatus**: Der ●-Punkt neben dem Dienstnamen ist **grün** wenn verbunden, **grau** wenn nicht verbunden
- **Lokaler Dienst-Ampel-Automatik** (v0.1.23+): Lokale Dienste (Ollama, LM Studio, vLLM) werden basierend auf der Verbindungsverfügbarkeit automatisch aktiviert/deaktiviert. Bei Verbindung wird der ●-Punkt innerhalb von 15 Sekunden grün und das Kontrollkästchen aktiviert; bei Trennung wird automatisch deaktiviert. Funktioniert gleich wie die schlüsselbasierte Auto-Umschaltung bei Cloud-Diensten (Google, OpenRouter usw.).

> 💡 **Wenn Ihr lokaler Dienst auf einem anderen Computer läuft**: Geben Sie die IP dieses Computers im Dienst-URL-Feld ein. Beispiel: `http://192.168.0.6:11434` (Ollama), `http://192.168.0.6:1234` (LM Studio). Wenn der Dienst nur an `127.0.0.1` statt `0.0.0.0` gebunden ist, funktioniert der Zugriff über externe IP nicht — überprüfen Sie die Bindungsadresse in den Diensteinstellungen.

### Admin-Token-Eingabe

Wenn Sie im Dashboard wichtige Funktionen wie das Hinzufügen oder Löschen von Schlüsseln nutzen möchten, erscheint ein Admin-Token-Eingabe-Popup. Geben Sie den Token ein, den Sie im Setup-Assistenten festgelegt haben. Nach einmaliger Eingabe bleibt er bis zum Schließen des Browsers gültig.

> ⚠️ **Wenn die Authentifizierungsfehler innerhalb von 15 Minuten 10 überschreiten, wird die IP vorübergehend gesperrt.** Wenn Sie den Token vergessen haben, prüfen Sie das `admin_token`-Feld in der Datei `wall-vault.yaml`.

---

## Verteilter Modus (Multi-Bot)

Eine Konfiguration zum **Teilen eines einzelnen Schlüsseltresors**, wenn OpenClaw auf mehreren Computern gleichzeitig betrieben wird. Praktisch, da die Schlüsselverwaltung nur an einem Ort erfolgt.

### Konfigurationsbeispiel

```
[Schlüsseltresor-Server]
  wall-vault vault    (Schlüsseltresor :56243, Dashboard)

[WSL Alpha]          [Raspberry Pi Gamma]    [Mac Mini Lokal]
  wall-vault proxy      wall-vault proxy        wall-vault proxy
  openclaw TUI          openclaw TUI            openclaw TUI
  ↕ SSE-Sync            ↕ SSE-Sync              ↕ SSE-Sync
```

Alle Bots zeigen auf den zentralen Tresor-Server, sodass Änderungen an Modellen oder hinzugefügte Schlüssel sofort auf alle Bots übertragen werden.

### Schritt 1: Schlüsseltresor-Server starten

Führen Sie dies auf dem Computer aus, der als Tresor-Server dienen soll:

```bash
wall-vault vault
```

### Schritt 2: Jeden Bot (Client) registrieren

Registrieren Sie vorab Informationen für jeden Bot, der sich mit dem Tresor-Server verbinden wird:

```bash
curl -X POST http://localhost:56243/admin/clients \
  -H "Authorization: Bearer IHR_ADMIN_TOKEN" \
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

Starten Sie auf jedem Computer mit einem Bot den Proxy mit Tresor-Server-Adresse und Token:

```bash
WV_VAULT_URL=http://192.168.x.x:56243 \
WV_VAULT_TOKEN=bota-secret \
WV_VAULT_CLIENT_ID=botA \
wall-vault proxy
```

> 💡 Ersetzen Sie **`192.168.x.x`** durch die tatsächliche interne IP-Adresse des Tresor-Server-Computers. Sie können sie über Ihre Router-Einstellungen oder den Befehl `ip addr` herausfinden.

---

## Autostart-Konfiguration

Wenn es lästig ist, wall-vault bei jedem Neustart manuell zu starten, registrieren Sie es als Systemdienst. Nach der Registrierung startet es automatisch beim Hochfahren.

### Linux — systemd (die meisten Linux-Distributionen)

systemd ist das System, das Programme unter Linux automatisch startet und verwaltet:

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

Das System für den automatischen Programmstart unter macOS:

```bash
wall-vault doctor deploy launchd
launchctl load ~/Library/LaunchAgents/com.wall-vault.plist
```

### Windows — NSSM

1. Laden Sie NSSM von [nssm.cc](https://nssm.cc/download) herunter und fügen Sie es zum PATH hinzu.
2. In einer Administrator-PowerShell:

```powershell
wall-vault doctor deploy windows
```

---

## Doctor (Diagnosetool)

Der `doctor`-Befehl ist ein Tool, das wall-vaults Konfiguration **selbst diagnostiziert und repariert**.

```bash
wall-vault doctor check   # Aktuellen Status diagnostizieren (nur lesen, nichts ändern)
wall-vault doctor fix     # Probleme automatisch beheben
wall-vault doctor all     # Diagnose + automatische Reparatur in einem Schritt
```

> 💡 Wenn etwas nicht stimmt, führen Sie zuerst `wall-vault doctor all` aus. Es erkennt viele Probleme automatisch.

---

## RTK Token-Einsparung

*(v0.1.24+)*

**RTK (Token-Einsparungstool)** komprimiert automatisch die Ausgabe von Shell-Befehlen, die von AI-Coding-Agenten (wie Claude Code) ausgeführt werden, und reduziert den Token-Verbrauch. Zum Beispiel werden 15 Zeilen `git status`-Ausgabe auf eine 2-Zeilen-Zusammenfassung komprimiert.

### Grundlegende Nutzung

```bash
# Befehle mit wall-vault rtk umschließen für automatische Ausgabefilterung
wall-vault rtk git status          # zeigt nur geänderte Dateiliste
wall-vault rtk git diff HEAD~1     # nur geänderte Zeilen + minimaler Kontext
wall-vault rtk git log -10         # Hash + Einzeilennachricht
wall-vault rtk go test ./...       # zeigt nur fehlgeschlagene Tests
wall-vault rtk ls -la              # nicht unterstützte Befehle werden automatisch gekürzt
```

### Unterstützte Befehle und Einsparungen

| Befehl | Filtermethode | Einsparung |
|--------|-------------|-----------|
| `git status` | Nur Zusammenfassung geänderter Dateien | ~87% |
| `git diff` | Geänderte Zeilen + 3 Zeilen Kontext | ~60-94% |
| `git log` | Hash + erste Nachrichtenzeile | ~90% |
| `git push/pull/fetch` | Fortschritt entfernen, nur Zusammenfassung | ~80% |
| `go test` | Nur Fehler anzeigen, Erfolge zählen | ~88-99% |
| `go build/vet` | Nur Fehler anzeigen | ~90% |
| Alle anderen Befehle | Erste 50 + letzte 50 Zeilen, max 32KB | Variabel |

### 3-Stufen-Filter-Pipeline

1. **Befehlsspezifischer Strukturfilter** — Versteht Ausgabeformate von git, go usw. und extrahiert nur bedeutungsvolle Teile
2. **Regex-Nachverarbeitung** — Entfernt ANSI-Farbcodes, komprimiert Leerzeilen, fasst Duplikate zusammen
3. **Durchleitung + Kürzung** — Nicht unterstützte Befehle behalten nur erste/letzte 50 Zeilen

### Claude Code-Integration

Sie können alle Shell-Befehle über Claude Codes `PreToolUse`-Hook automatisch durch RTK leiten.

```bash
# Hook installieren (automatisch zur Claude Code settings.json hinzugefügt)
wall-vault rtk hook install
```

Oder manuell in `~/.claude/settings.json` hinzufügen:

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

> 💡 **Exit-Code-Erhaltung**: RTK gibt den Exit-Code des Originalbefehls unverändert zurück. Wenn ein Befehl fehlschlägt (Exit-Code ≠ 0), erkennt die AI den Fehler korrekt.

> 💡 **Erzwungene englische Ausgabe**: RTK führt Befehle mit `LC_ALL=C` aus, um unabhängig von den Systemspracheinstellungen immer englische Ausgabe zu erzeugen. Dies stellt sicher, dass Filter korrekt funktionieren.

---

## Umgebungsvariablen-Referenz

Umgebungsvariablen sind eine Methode, Konfigurationswerte an Programme zu übergeben. Geben Sie sie im Terminal als `export VARIABLE=Wert` ein, oder fügen Sie sie in Ihre Autostart-Dienstdatei ein.

| Variable | Beschreibung | Beispielwert |
|----------|-------------|-------------|
| `WV_LANG` | Dashboard-Sprache | `ko`, `en`, `ja` |
| `WV_THEME` | Dashboard-Theme | `light`, `dark`, `gold` |
| `WV_KEY_GOOGLE` | Google API-Schlüssel (kommagetrennt für mehrere) | `AIza...,AIza...` |
| `WV_KEY_OPENROUTER` | OpenRouter API-Schlüssel | `sk-or-v1-...` |
| `WV_VAULT_URL` | Tresor-Server-Adresse im verteilten Modus | `http://192.168.x.x:56243` |
| `WV_VAULT_TOKEN` | Client-(Bot-)Authentifizierungstoken | `my-secret-token` |
| `WV_ADMIN_TOKEN` | Admin-Token | `admin-token-here` |
| `WV_MASTER_PASS` | API-Schlüssel-Verschlüsselungspasswort | `my-password` |
| `WV_AVATAR` | Avatar-Bilddateipfad (relativ zu `~/.openclaw/`) | `workspace/avatars/avatar.png` |
| `OLLAMA_URL` | Ollama-Server-Adresse (lokal) | `http://192.168.x.x:11434` |

---

## Fehlerbehebung

### Wenn der Proxy nicht startet

Der Port wird oft bereits von einem anderen Programm verwendet.

```bash
ss -tlnp | grep 56244   # Prüfen, wer Port 56244 verwendet
wall-vault proxy --port 8080   # Mit anderer Portnummer starten
```

### API-Schlüssel-Fehler (429, 402, 401, 403, 582)

| Fehlercode | Bedeutung | Lösung |
|-----------|---------|--------|
| **429** | Zu viele Anfragen (Nutzungslimit überschritten) | Kurz warten oder mehr Schlüssel hinzufügen |
| **402** | Zahlung erforderlich oder Guthaben aufgebraucht | Guthaben beim jeweiligen Dienst aufladen |
| **401 / 403** | Ungültiger Schlüssel oder keine Berechtigung | Schlüsselwert erneut prüfen und neu registrieren |
| **582** | Gateway-Überlastung (5-Minuten-Cooldown) | Löst sich nach 5 Minuten automatisch |

```bash
# Registrierte Schlüsselliste und Status prüfen
curl -H "Authorization: Bearer IHR_ADMIN_TOKEN" http://localhost:56243/admin/keys

# Schlüsselnutzungszähler zurücksetzen
curl -X POST -H "Authorization: Bearer IHR_ADMIN_TOKEN" http://localhost:56243/admin/keys/reset
```

### Wenn der Agent als „Nicht verbunden" angezeigt wird

„Nicht verbunden" bedeutet, dass der Proxy-Prozess keine Heartbeat-Signale an den Tresor sendet. **Es bedeutet nicht, dass Einstellungen nicht gespeichert sind.** Der Proxy muss mit der Tresor-Server-Adresse und dem Token laufen, um als verbunden angezeigt zu werden.

```bash
# Proxy mit Tresor-Server-Adresse, Token und Client-ID starten
WV_VAULT_URL=http://TRESOR_SERVER:56243 \
WV_VAULT_TOKEN=client-token \
WV_VAULT_CLIENT_ID=client-id \
wall-vault proxy
```

Nach erfolgreicher Verbindung zeigt das Dashboard innerhalb von etwa 20 Sekunden 🟢 Aktiv an.

### Wenn Ollama keine Verbindung herstellt

Ollama ist ein Programm, das AI direkt auf Ihrem Computer ausführt. Prüfen Sie zunächst, ob Ollama läuft.

```bash
curl http://localhost:11434/api/tags   # Wenn die Modellliste erscheint, funktioniert es
export OLLAMA_URL=http://192.168.x.x:11434   # Wenn auf einem anderen Computer
```

> ⚠️ Wenn Ollama nicht antwortet, starten Sie es zuerst mit dem Befehl `ollama serve`.

> ⚠️ **Große Modelle sind langsam**: Große Modelle wie `qwen3.5:35b` oder `deepseek-r1` können mehrere Minuten zur Antwortgenerierung benötigen. Auch wenn es so aussieht, als gäbe es keine Antwort, verarbeitet es möglicherweise normal — bitte warten Sie.

---

## Letzte Änderungen (v0.1.16 ~ v0.1.25)

### v0.1.25 (2026-04-08)
- **Agenten-Prozess-Erkennung**: Proxy erkennt, ob lokale Agenten (NanoClaw/OpenClaw) am Leben sind, und zeigt eine orange Ampel im Dashboard an.
- **Drag-Handle-Verbesserung**: Kartensortierung funktioniert nur noch über den Ampel-Bereich (●). Verhindert versehentliches Ziehen von Eingabefeldern oder Schaltflächen.

### v0.1.24 (2026-04-06)
- **RTK Token-Einsparungs-Unterbefehl**: `wall-vault rtk <command>` filtert Shell-Befehlsausgaben automatisch und reduziert den Token-Verbrauch von AI-Agenten um 60-90%. Eingebaute Filter für Hauptbefehle wie git und go, mit automatischer Kürzung für nicht unterstützte Befehle. Transparente Integration mit Claude Code `PreToolUse`-Hooks.

### v0.1.23 (2026-04-06)
- **Ollama-Modelländerung behoben**: Problem behoben, bei dem Änderungen des Ollama-Modells im Tresor-Dashboard nicht im Proxy wirksam wurden. Zuvor wurde nur die Umgebungsvariable (`OLLAMA_MODEL`) verwendet, jetzt haben Tresor-Einstellungen Vorrang.
- **Lokaler Dienst-Ampel-Automatik**: Ollama, LM Studio und vLLM werden bei Verfügbarkeit automatisch aktiviert und bei Trennung automatisch deaktiviert. Gleicher Mechanismus wie die schlüsselbasierte Auto-Umschaltung bei Cloud-Diensten.

### v0.1.22 (2026-04-05)
- **Leeres content-Feld behoben**: Problem behoben, bei dem Thinking-Modelle (gemini-3.1-pro, o1, Claude Thinking usw.), die alle max_tokens für Reasoning verbrauchten, ohne tatsächliche Antworten zu erzeugen, dazu führten, dass der Proxy `content`/`text`-Felder über `omitempty` ausließ und OpenAI/Anthropic SDK-Clients mit `Cannot read properties of undefined (reading 'trim')`-Fehlern abstürzten. Geändert, um Felder gemäß offizieller API-Spezifikation immer einzuschließen.

### v0.1.21 (2026-04-05)
- **Gemma 4-Modellunterstützung**: Gemma-Modelle wie `gemma-4-31b-it` und `gemma-4-26b-a4b-it` können jetzt über die Google Gemini API verwendet werden.
- **LM Studio / vLLM offizielle Unterstützung**: Zuvor fehlten diese Dienste im Proxy-Routing und fielen immer auf Ollama zurück. Jetzt korrekt über OpenAI-kompatible API geroutet.
- **Dashboard-Dienstanzeige behoben**: Dashboard zeigt immer den vom Benutzer konfigurierten Dienst, auch bei Fallback.
- **Lokaler Dienst-Statusanzeige**: Zeigt den Verbindungsstatus lokaler Dienste (Ollama, LM Studio, vLLM usw.) über ●-Punktfarbe beim Dashboard-Laden.
- **Tool-Filter-Umgebungsvariable**: Tool-Durchleitungsmodus kann mit der Umgebungsvariable `WV_TOOL_FILTER=passthrough` eingestellt werden.

### v0.1.20 (2026-03-28)
- **Umfassende Sicherheitshärtung**: XSS-Prävention (41 Stellen), Konstantzeit-Token-Vergleich, CORS-Einschränkung, Anfragegrößenlimits, Pfad-Traversal-Prävention, SSE-Authentifizierung, Rate-Limiter-Härtung und insgesamt 12 Sicherheitsverbesserungen.

### v0.1.19 (2026-03-27)
- **Claude Code Online-Erkennung**: Claude Code, das nicht über den Proxy läuft, wird jetzt im Dashboard als online angezeigt.

### v0.1.18 (2026-03-26)
- **Fallback-Dienst-Festkleben behoben**: Nach temporärem Fehler-Fallback auf Ollama erfolgt automatische Rückkehr zum Originaldienst bei Erholung.
- **Offline-Erkennung verbessert**: 15-Sekunden-Statusprüfungen machen die Proxy-Ausfallserkennung schneller.

### v0.1.17 (2026-03-25)
- **Drag-and-Drop-Kartensortierung**: Agentenkarten können per Drag-and-Drop neu sortiert werden.
- **Inline-Konfiguration-Anwenden-Schaltflächen**: [⚡ Konfiguration anwenden]-Schaltflächen erscheinen auf Offline-Agentenkarten.
- **cokacdir-Agententyp hinzugefügt**.

### v0.1.16 (2026-03-25)
- **Bidirektionale Modell-Synchronisation**: Änderungen des Cline- oder Claude Code-Modells im Tresor-Dashboard werden automatisch übernommen.

---

*Für detailliertere API-Informationen siehe [API.md](API.md).*
