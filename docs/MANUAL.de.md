# wall-vault Benutzerhandbuch
*(Last updated: 2026-04-09 — v0.1.27)*

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
9. [Doctor (Diagnose)](#doctor-diagnose)
10. [RTK Token-Einsparung](#rtk-token-einsparung)
11. [Umgebungsvariablen-Referenz](#umgebungsvariablen-referenz)
12. [Fehlerbehebung](#fehlerbehebung)

---

## v0.2 Upgrade-Hinweise

- `Service` hat jetzt `default_model` und `allowed_models` gewonnen. Das dienstspezifische Standardmodell wird jetzt direkt auf der Service-Karte festgelegt.
- `Client.default_service` / `default_model` wurden in `preferred_service` / `model_override` umbenannt und neu interpretiert. Wenn die Überschreibung leer ist, wird das Standardmodell des Dienstes verwendet.
- Beim ersten v0.2-Start wird die vorhandene `vault.json` automatisch migriert, und der Zustand vor der Migration wird als `vault.json.pre-v02.{timestamp}.bak` gespeichert.
- Das Dashboard wurde in drei Zonen umstrukturiert: eine linke Seitenleiste, ein zentrales Kartengitter und ein rechtseitiger Edit-Slideover.
- Admin-API-Pfade sind unverändert, aber die Request-/Response-Body-Schemas wurden aktualisiert – alte CLI-Skripte müssen entsprechend aktualisiert werden.

---

## Was ist wall-vault?

**wall-vault = KI-Proxy + API-Schlüsseltresor für OpenClaw**

Um KI-Dienste zu nutzen, benötigen Sie **API-Schlüssel**. Ein API-Schlüssel ist wie ein **digitaler Ausweis**, der beweist: "Diese Person ist berechtigt, diesen Dienst zu nutzen." Allerdings haben diese Ausweise tägliche Nutzungslimits, und bei unsachgemäßer Verwaltung besteht das Risiko einer Offenlegung.

wall-vault bewahrt diese Ausweise in einem sicheren Tresor auf und fungiert als **Proxy (Stellvertreter)** zwischen OpenClaw und den KI-Diensten. Einfach gesagt: OpenClaw muss sich nur mit wall-vault verbinden, und wall-vault erledigt den Rest.

Probleme, die wall-vault löst:

- **Automatische API-Schlüssel-Rotation**: Wenn ein Schlüssel sein Nutzungslimit erreicht oder vorübergehend gesperrt wird (Cooldown), wechselt das System lautlos zum nächsten Schlüssel. OpenClaw arbeitet ohne Unterbrechung weiter.
- **Automatischer Service-Fallback**: Wenn Google nicht antwortet, wird auf OpenRouter umgeschaltet. Wenn das auch nicht funktioniert, wird automatisch auf lokal installierte KI (Ollama, LM Studio, vLLM) gewechselt. Ihre Sitzung wird nicht unterbrochen. Wenn der ursprüngliche Dienst wiederhergestellt ist, wird bei der nächsten Anfrage automatisch zurückgewechselt (v0.1.18+, LM Studio/vLLM: v0.1.21+).
- **Echtzeit-Synchronisation (SSE)**: Wenn Sie das Modell im Tresor-Dashboard ändern, wird es innerhalb von 1-3 Sekunden auf dem OpenClaw-Bildschirm übernommen. SSE (Server-Sent Events) ist eine Technologie, bei der der Server Änderungen in Echtzeit an Clients pusht.
- **Echtzeit-Benachrichtigungen**: Ereignisse wie erschöpfte Schlüssel oder Dienstausfälle werden sofort am unteren Rand des OpenClaw-TUI (Terminal-Bildschirm) angezeigt.

> :bulb: **Claude Code, Cursor und VS Code** können ebenfalls verbunden werden, aber der Hauptzweck von wall-vault ist die Verwendung mit OpenClaw.

```
OpenClaw (TUI-Terminal-Bildschirm)
        |
        v
  wall-vault Proxy (:56244)   <- Schlüsselverwaltung, Routing, Fallback, Events
        |
        +-- Google Gemini API
        +-- OpenRouter API (340+ Modelle)
        +-- Ollama / LM Studio / vLLM (lokaler Rechner, letzte Instanz)
        +-- OpenAI / Anthropic API
```

---

## Installation

### Linux / macOS

Öffnen Sie ein Terminal und fügen Sie die folgenden Befehle ein:

```bash
# Linux (Standard-PC, Server — amd64)
curl -L https://github.com/sookmook/wall-vault/releases/latest/download/wall-vault-linux-amd64 \
  -o ~/.local/bin/wall-vault && chmod +x ~/.local/bin/wall-vault

# macOS Apple Silicon (M1/M2/M3 Mac)
curl -L https://github.com/sookmook/wall-vault/releases/latest/download/wall-vault-darwin-arm64 \
  -o /usr/local/bin/wall-vault && chmod +x /usr/local/bin/wall-vault
```

- `curl -L ...` — Lädt die Datei aus dem Internet herunter.
- `chmod +x` — Macht die heruntergeladene Datei "ausführbar". Wenn Sie diesen Schritt auslassen, erhalten Sie einen "Zugriff verweigert"-Fehler.

### Windows

Öffnen Sie PowerShell (als Administrator) und führen Sie folgende Befehle aus:

```powershell
# Download
Invoke-WebRequest -Uri `
  "https://github.com/sookmook/wall-vault/releases/latest/download/wall-vault-windows-amd64.exe" `
  -OutFile "$env:LOCALAPPDATA\Programs\wall-vault\wall-vault.exe"

# Zum PATH hinzufügen (wird nach Neustart von PowerShell wirksam)
$env:PATH += ";$env:LOCALAPPDATA\Programs\wall-vault"
```

> :bulb: **Was ist PATH?** Eine Liste von Ordnern, in denen Ihr Computer nach Befehlen sucht. Durch Hinzufügen zum PATH können Sie `wall-vault` von jedem Ordner aus ausführen.

### Aus Quellcode kompilieren (für Entwickler)

Nur relevant, wenn die Go-Entwicklungsumgebung installiert ist.

```bash
git clone https://github.com/sookmook/wall-vault
cd wall-vault
make build       # bin/wall-vault (Version: v0.1.25.YYYYMMDD.HHmmss)
make install     # ~/.local/bin/wall-vault
```

> :bulb: **Build-Zeitstempel-Version**: Beim Kompilieren mit `make build` wird die Version automatisch in einem Format wie `v0.1.27.20260409` mit Datum und Uhrzeit generiert. Beim direkten Kompilieren mit `go build ./...` wird die Version nur als `"dev"` angezeigt.

---

## Erste Schritte

### Setup-Assistent ausführen

Führen Sie nach der Installation unbedingt den **Setup-Assistenten** mit folgendem Befehl aus. Der Assistent führt Sie Schritt für Schritt durch die notwendigen Einstellungen.

```bash
wall-vault setup
```

Der Assistent durchläuft folgende Schritte:

```
1. Sprachauswahl (10 Sprachen inkl. Deutsch)
2. Theme-Auswahl (light / dark / gold / cherry / ocean)
3. Betriebsmodus — Standalone (Einzelrechner) oder Distributed (mehrere Rechner)
4. Bot-Name — Der auf dem Dashboard angezeigte Name
5. Port-Konfiguration — Standard: Proxy 56244, Tresor 56243 (Enter für Standardwerte)
6. KI-Service-Auswahl — Google / OpenRouter / Ollama / LM Studio / vLLM
7. Tool-Sicherheitsfilter-Einstellungen
8. Admin-Token — Ein Passwort zum Sperren der Dashboard-Verwaltungsfunktionen. Automatische Generierung möglich
9. API-Schlüssel-Verschlüsselungspasswort — Für sicherere Schlüsselspeicherung (optional)
10. Speicherpfad der Konfigurationsdatei
```

> :warning: **Merken Sie sich Ihr Admin-Token.** Sie benötigen es später, um Schlüssel hinzuzufügen oder Einstellungen im Dashboard zu ändern. Wenn Sie es verlieren, müssen Sie die Konfigurationsdatei manuell bearbeiten.

Nach Abschluss des Assistenten wird automatisch eine `wall-vault.yaml`-Konfigurationsdatei erstellt.

### Starten

```bash
wall-vault start
```

Die folgenden zwei Server starten gleichzeitig:

- **Proxy** (`http://localhost:56244`) — Der Vermittler zwischen OpenClaw und KI-Diensten
- **Schlüsseltresor** (`http://localhost:56243`) — API-Schlüsselverwaltung und Web-Dashboard

Öffnen Sie `http://localhost:56243` in Ihrem Browser, um auf das Dashboard zuzugreifen.

---

## API-Schlüssel registrieren

Es gibt vier Methoden zur Registrierung von API-Schlüsseln. **Für Anfänger wird Methode 1 (Umgebungsvariablen) empfohlen.**

### Methode 1: Umgebungsvariablen (empfohlen — am einfachsten)

Umgebungsvariablen sind **voreingestellte Werte**, die Programme beim Start einlesen. Geben Sie einfach folgendes im Terminal ein:

```bash
# Google Gemini-Schlüssel registrieren
export WV_KEY_GOOGLE=AIzaSy...

# OpenRouter-Schlüssel registrieren
export WV_KEY_OPENROUTER=sk-or-v1-...

# Nach der Registrierung starten
wall-vault start
```

Wenn Sie mehrere Schlüssel haben, trennen Sie diese mit Kommas. wall-vault verwendet sie automatisch im Wechsel (Round Robin):

```bash
export WV_KEY_GOOGLE=AIzaSy...,AIzaSy...,AIzaSy...
```

> :bulb: **Tipp**: Der `export`-Befehl gilt nur für die aktuelle Terminal-Sitzung. Damit es auch nach einem Neustart bestehen bleibt, fügen Sie die Zeilen zu Ihrer `~/.bashrc`- oder `~/.zshrc`-Datei hinzu.

### Methode 2: Dashboard-UI (Mausklick)

1. Öffnen Sie `http://localhost:56243` im Browser
2. Klicken Sie im oberen **:key: API-Schlüssel**-Bereich auf `[+ Hinzufügen]`
3. Geben Sie Servicetyp, Schlüsselwert, Bezeichnung (Memo-Name) und Tageslimit ein und speichern Sie

### Methode 3: REST-API (für Automatisierung/Skripte)

REST-API ist eine Methode, mit der Programme Daten über HTTP austauschen. Nützlich für automatisierte Registrierung per Skript.

```bash
curl -X POST http://localhost:56243/admin/keys \
  -H "Authorization: Bearer admin-token" \
  -H "Content-Type: application/json" \
  -d '{
    "service": "google",
    "key": "AIzaSy...",
    "label": "Hauptschlüssel",
    "daily_limit": 1000
  }'
```

### Methode 4: Proxy-Flags (für schnelle Tests)

Verwenden Sie dies für temporäre Tests ohne formale Registrierung. Die Schlüssel verschwinden beim Beenden des Programms.

```bash
wall-vault proxy --key-google=AIzaSy... --key-openrouter=sk-or-...
```

---

## Proxy-Nutzung

### Verwendung mit OpenClaw (Hauptzweck)

So konfigurieren Sie OpenClaw, um KI-Dienste über wall-vault zu nutzen.

Öffnen Sie `~/.openclaw/openclaw.json` und fügen Sie folgendes hinzu:

```json5
// ~/.openclaw/openclaw.json
{
  models: {
    providers: {
      "wall-vault": {
        baseUrl: "http://localhost:56244/v1",
        apiKey: "your-agent-token",   // Vault-Agent-Token
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

> :bulb: **Einfacherer Weg**: Klicken Sie auf die **:lobster: OpenClaw-Konfiguration kopieren**-Schaltfläche auf der Dashboard-Agentenkarte. Ein Snippet mit vorausgefülltem Token und Adresse wird in die Zwischenablage kopiert. Einfach einfügen.

**Wohin leitet das `wall-vault/`-Präfix im Modellnamen weiter?**

wall-vault bestimmt automatisch anhand des Modellnamens, an welchen KI-Dienst die Anfrage gesendet wird:

| Modellformat | Angesteuerter Dienst |
|-------------|---------------------|
| `wall-vault/gemini-*` | Direkt zu Google Gemini |
| `wall-vault/gpt-*`, `wall-vault/o3`, `wall-vault/o4*` | Direkt zu OpenAI |
| `wall-vault/claude-*` | Anthropic über OpenRouter |
| `wall-vault/hunter-alpha`, `wall-vault/healer-alpha` | OpenRouter (kostenloser 1M-Token-Kontext) |
| `wall-vault/kimi-*`, `wall-vault/glm-*`, `wall-vault/deepseek-*` | OpenRouter |
| `google/modellname`, `openai/modellname`, `anthropic/modellname` usw. | Direkt zum jeweiligen Dienst |
| `custom/google/modellname`, `custom/openai/modellname` usw. | Entfernt `custom/`-Präfix und leitet um |
| `modellname:cloud` | Entfernt `:cloud`-Suffix und leitet zu OpenRouter |

> :bulb: **Was ist Kontext?** Die Menge an Konversation, die eine KI auf einmal behalten kann. 1M (eine Million Token) bedeutet, dass auch sehr lange Gespräche oder Dokumente in einem Durchgang verarbeitet werden können.

### Direktverbindung über Gemini-API-Format (Kompatibilität mit bestehenden Tools)

Wenn Sie Tools haben, die die Google Gemini API direkt verwenden, ändern Sie einfach die URL auf wall-vault:

```bash
export ANTHROPIC_BASE_URL=http://localhost:56244/google
```

Oder wenn das Tool URLs direkt angibt:

```
http://localhost:56244/google/v1beta/models/gemini-2.5-flash:generateContent
```

### Verwendung mit dem OpenAI SDK (Python)

Sie können wall-vault auch mit Python-Code verbinden, der KI nutzt. Ändern Sie einfach die `base_url`:

```python
from openai import OpenAI

client = OpenAI(
    base_url="http://localhost:56244/v1",
    api_key="not-needed"  # API-Schlüssel werden von wall-vault verwaltet
)

response = client.chat.completions.create(
    model="google/gemini-2.5-flash",   # Verwenden Sie das Provider/Modell-Format
    messages=[{"role": "user", "content": "Hallo"}]
)
```

### Modell zur Laufzeit ändern

Um das KI-Modell zu ändern, während wall-vault bereits läuft:

```bash
# Modell direkt über Proxy-Anfrage ändern
curl -X PUT http://localhost:56244/api/config/model \
  -H "Content-Type: application/json" \
  -d '{"service": "openrouter", "model": "anthropic/claude-3.5-sonnet"}'

# Im verteilten Modus (Multi-Bot): Änderung am Tresor-Server -> sofortige Synchronisation via SSE
curl -X PUT http://localhost:56243/admin/clients/mein-bot-id \
  -H "Authorization: Bearer admin-token" \
  -H "Content-Type: application/json" \
  -d '{"default_service": "google", "default_model": "gemini-2.5-pro"}'
```

### Verfügbare Modelle auflisten

```bash
# Vollständige Liste anzeigen
curl http://localhost:56244/api/models | python3 -m json.tool

# Nur Google-Modelle anzeigen
curl "http://localhost:56244/api/models?service=google"

# Nach Name suchen (z.B. Modelle mit "claude")
curl "http://localhost:56244/api/models?q=claude"
```

**Wichtige Modelle nach Service:**

| Dienst | Wichtige Modelle |
|--------|-----------------|
| Google | gemini-2.5-pro, gemini-2.5-flash, gemini-2.5-flash-8b, gemini-2.0-flash |
| OpenAI | gpt-4o, gpt-4o-mini, o3, o1, o1-mini |
| OpenRouter | 346+ (Hunter Alpha 1M-Kontext kostenlos, DeepSeek R1/V3, Qwen 2.5 usw.) |
| Ollama | Automatisch erkannt vom lokal installierten Server |
| LM Studio | Lokaler Server (Port 1234) |
| vLLM | Lokaler Server (Port 8000) |

---

## Schlüsseltresor-Dashboard

Öffnen Sie `http://localhost:56243` in Ihrem Browser, um das Dashboard anzuzeigen.

**Bildschirmaufbau:**
- **Obere Leiste (fixiert)**: Logo, Sprach-/Theme-Auswahl, SSE-Verbindungsstatus
- **Karten-Raster**: Agenten-, Service- und API-Schlüsselkarten in Kachelform

### API-Schlüsselkarte

Eine Karte zur übersichtlichen Verwaltung aller registrierten API-Schlüssel.

- Zeigt die Schlüsselliste nach Service gruppiert an.
- `today_usage`: Heute erfolgreich verarbeitete Token (von der KI gelesene und geschriebene Zeichen)
- `today_attempts`: Gesamtanzahl der heutigen Aufrufe (Erfolge + Fehlschläge)
- Verwenden Sie `[+ Hinzufügen]` zum Registrieren neuer Schlüssel und `x` zum Löschen.

> :bulb: **Was sind Token?** Token sind die Einheiten, die KI zur Textverarbeitung verwendet. Ungefähr ein englisches Wort oder 1-2 deutsche Zeichen. API-Preise werden in der Regel nach Token-Anzahl berechnet.

### Agentenkarte

Eine Karte, die den Status der mit dem wall-vault-Proxy verbundenen Bots (Agenten) anzeigt.

**Der Verbindungsstatus wird in 4 Stufen angezeigt:**

| Anzeige | Status | Bedeutung |
|---------|--------|-----------|
| :green_circle: | Aktiv | Proxy arbeitet normal |
| :yellow_circle: | Verzögert | Antwortet, aber langsam |
| :red_circle: | Offline | Proxy antwortet nicht |
| :black_circle: | Nicht verbunden / Deaktiviert | Proxy war nie mit dem Tresor verbunden oder ist deaktiviert |

**Agentenkarten-Schaltflächen:**

Wenn Sie beim Registrieren eines Agenten den **Agententyp** angeben, erscheinen automatisch passende Komfort-Schaltflächen.

---

#### :radio_button: Konfiguration-kopieren-Schaltfläche — Erstellt automatisch Verbindungseinstellungen

Durch Klicken wird ein Konfigurations-Snippet mit vorausgefülltem Token, Proxy-Adresse und Modellinformationen in die Zwischenablage kopiert. Fügen Sie den kopierten Inhalt einfach an der in der Tabelle angegebenen Stelle ein.

| Schaltfläche | Agententyp | Einfügeort |
|-------------|-----------|------------|
| :lobster: OpenClaw-Konfiguration kopieren | `openclaw` | `~/.openclaw/openclaw.json` |
| :crab: NanoClaw-Konfiguration kopieren | `nanoclaw` | `~/.openclaw/openclaw.json` |
| :orange_circle: Claude Code-Konfiguration kopieren | `claude-code` | `~/.claude/settings.json` |
| :keyboard: Cursor-Konfiguration kopieren | `cursor` | Cursor -> Settings -> AI |
| :computer: VSCode-Konfiguration kopieren | `vscode` | `~/.continue/config.json` |

**Beispiel — Für Claude Code-Typ wird folgendes kopiert:**

```json
// ~/.claude/settings.json
{
  "apiProvider": "openai",
  "baseUrl": "http://192.168.1.20:56244/v1",
  "apiKey": "token-dieses-agenten"
}
```

**Beispiel — Für VSCode (Continue)-Typ:**

```yaml
# ~/.continue/config.yaml  <- In config.yaml einfügen, NICHT config.json
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

> :warning: **Aktuelle Continue-Versionen verwenden `config.yaml`.** Wenn `config.yaml` existiert, wird `config.json` komplett ignoriert. Fügen Sie unbedingt in `config.yaml` ein.

**Beispiel — Für Cursor-Typ:**

```
Base URL : http://192.168.1.20:56244/v1
API Key  : token-dieses-agenten

// Oder Umgebungsvariablen:
OPENAI_BASE_URL=http://192.168.1.20:56244/v1
OPENAI_API_KEY=token-dieses-agenten
```

> :warning: **Wenn das Kopieren in die Zwischenablage nicht funktioniert**: Browser-Sicherheitsrichtlinien können das Kopieren blockieren. Wenn ein Popup-Textfeld erscheint, wählen Sie alles mit Ctrl+A aus und kopieren Sie mit Ctrl+C.

---

#### :zap: Automatische-Anwendung-Schaltfläche — Ein Klick und die Konfiguration steht

Bei Agenten vom Typ `cline`, `claude-code`, `openclaw` oder `nanoclaw` erscheint eine **:zap: Konfiguration anwenden**-Schaltfläche auf der Agentenkarte. Ein Klick aktualisiert automatisch die lokale Konfigurationsdatei des Agenten.

| Schaltfläche | Agententyp | Zieldatei |
|-------------|-----------|-----------|
| :zap: Cline-Konfiguration anwenden | `cline` | `~/.cline/data/globalState.json` + `secrets.json` |
| :zap: Claude Code-Konfiguration anwenden | `claude-code` | `~/.claude/settings.json` |
| :zap: OpenClaw-Konfiguration anwenden | `openclaw` | `~/.openclaw/openclaw.json` |
| :zap: NanoClaw-Konfiguration anwenden | `nanoclaw` | `~/.openclaw/openclaw.json` |

> :warning: Diese Schaltfläche sendet eine Anfrage an **localhost:56244** (lokaler Proxy). Der Proxy muss auf diesem Rechner laufen.

---

#### :twisted_rightwards_arrows: Drag-and-Drop Kartensortierung (v0.1.17, verbessert v0.1.25)

Sie können Agentenkarten im Dashboard per **Drag-and-Drop** in beliebiger Reihenfolge anordnen.

1. Greifen Sie den **Ampel-Bereich (●)** oben links auf der Karte mit der Maus und ziehen Sie
2. Lassen Sie auf der Karte an der gewünschten Position los, um die Reihenfolge zu tauschen

> :bulb: Der Karteninhalt (Eingabefelder, Schaltflächen usw.) kann nicht gezogen werden. Nur im Ampelbereich greifen.

#### :orange_circle: Agentenprozess-Erkennung (v0.1.25)

Wenn der Proxy normal funktioniert, aber ein lokaler Agentenprozess (NanoClaw, OpenClaw) gestorben ist, wechselt die Kartenampel auf **orange (blinkend)** und zeigt eine "Agentenprozess gestoppt"-Meldung an.

- :green_circle: Grün: Proxy + Agent normal
- :orange_circle: Orange (blinkend): Proxy normal, Agent gestoppt
- :red_circle: Rot: Proxy offline
3. Die geänderte Reihenfolge wird **sofort auf dem Server gespeichert** und bleibt nach Seitenaktualisierung bestehen

> :bulb: Touch-Geräte (Mobiltelefone/Tablets) werden noch nicht unterstützt. Verwenden Sie einen Desktop-Browser.

---

#### :arrows_counterclockwise: Bidirektionale Modell-Synchronisation (v0.1.16)

Wenn Sie das Modell eines Agenten im Tresor-Dashboard ändern, wird die lokale Konfiguration des Agenten automatisch aktualisiert.

**Für Cline:**
- Modelländerung im Tresor -> SSE-Event -> Proxy aktualisiert Modellfelder in `globalState.json`
- Aktualisierte Felder: `actModeOpenAiModelId`, `planModeOpenAiModelId`, `openAiModelId`
- `openAiBaseUrl` und API-Schlüssel werden nicht verändert
- **VS Code-Neustart erforderlich (`Ctrl+Alt+R` oder `Ctrl+Shift+P` -> `Developer: Reload Window`)**
  - Weil Cline Konfigurationsdateien zur Laufzeit nicht neu einliest

**Für Claude Code:**
- Modelländerung im Tresor -> SSE-Event -> Proxy aktualisiert `model`-Feld in `settings.json`
- Durchsucht automatisch sowohl WSL- als auch Windows-Pfade (`~/.claude/`, `/mnt/c/Users/*/.claude/`)

**Umgekehrte Richtung (Agent -> Tresor):**
- Wenn Agenten (Cline, Claude Code usw.) Anfragen an den Proxy senden, enthält der Proxy im Heartbeat die Service/Modell-Informationen des Clients
- Die Agentenkarte im Tresor-Dashboard zeigt den aktuell verwendeten Service/Modell in Echtzeit an

> :bulb: **Kernpunkt**: Der Proxy identifiziert Agenten anhand des Authorization-Tokens in Anfragen und routet automatisch zum im Tresor konfigurierten Service/Modell. Selbst wenn Cline oder Claude Code einen anderen Modellnamen senden, überschreibt der Proxy ihn mit der Tresor-Einstellung.

---

### Cline in VS Code verwenden — Detaillierter Leitfaden

#### Schritt 1: Cline installieren

Installieren Sie **Cline** (ID: `saoudrizwan.claude-dev`) aus dem VS Code Extension Marketplace.

#### Schritt 2: Agent im Tresor registrieren

1. Öffnen Sie das Tresor-Dashboard (`http://tresor-IP:56243`)
2. Klicken Sie im **Agenten**-Bereich auf **+ Hinzufügen**
3. Füllen Sie folgendes aus:

| Feld | Wert | Beschreibung |
|------|------|-------------|
| ID | `mein_cline` | Eindeutige Kennung (alphanumerisch, keine Leerzeichen) |
| Name | `Mein Cline` | Auf dem Dashboard angezeigter Name |
| Agententyp | `cline` | <- Muss `cline` ausgewählt werden |
| Service | Gewünschten Service auswählen (z.B. `google`) | |
| Modell | Gewünschtes Modell eingeben (z.B. `gemini-2.5-flash`) | |

4. Klicken Sie auf **Speichern**, um ein Token automatisch zu generieren

#### Schritt 3: Cline verbinden

**Methode A — Automatische Anwendung (empfohlen)**

1. Stellen Sie sicher, dass der wall-vault-**Proxy** auf diesem Rechner läuft (`localhost:56244`)
2. Klicken Sie auf die **:zap: Cline-Konfiguration anwenden**-Schaltfläche auf der Dashboard-Agentenkarte
3. Erfolg bei der Meldung "Konfiguration angewendet!"
4. Laden Sie VS Code neu (`Ctrl+Alt+R`)

**Methode B — Manuelle Einrichtung**

Öffnen Sie die Einstellungen (:gear:) in der Cline-Seitenleiste:
- **API Provider**: `OpenAI Compatible`
- **Base URL**: `http://proxy-adresse:56244/v1`
  - Gleicher Rechner: `http://localhost:56244/v1`
  - Anderer Rechner (z.B. Mac Mini): `http://192.168.1.20:56244/v1`
- **API Key**: Vom Tresor ausgegebenes Token (von der Agentenkarte kopieren)
- **Model ID**: Im Tresor konfiguriertes Modell (z.B. `gemini-2.5-flash`)

#### Schritt 4: Überprüfung

Senden Sie eine beliebige Nachricht im Cline-Chat-Fenster. Bei korrekter Funktion:
- Die entsprechende Agentenkarte im Tresor-Dashboard zeigt einen **grünen Punkt (● Aktiv)**
- Die Karte zeigt den aktuellen Service/Modell (z.B. `google / gemini-2.5-flash`)

#### Modell ändern

Wenn Sie das Cline-Modell ändern möchten, tun Sie dies über das **Tresor-Dashboard**:

1. Ändern Sie das Service/Modell-Dropdown auf der Agentenkarte
2. Klicken Sie auf **Anwenden**
3. Laden Sie VS Code neu (`Ctrl+Alt+R`) — Der Modellname in der Cline-Fußzeile wird aktualisiert
4. Das neue Modell wird ab der nächsten Anfrage verwendet

> :bulb: Tatsächlich identifiziert der Proxy Clines Anfragen anhand des Tokens und routet sie zum im Tresor konfigurierten Modell. Auch ohne VS Code-Neustart **ändert sich das tatsächlich verwendete Modell sofort** — der Neustart dient nur dazu, die Modellanzeige in der Cline-UI zu aktualisieren.

#### Trennungserkennung

Wenn Sie VS Code schließen, wird die Agentenkarte im Tresor-Dashboard nach etwa **90 Sekunden** gelb (verzögert) und nach **3 Minuten** rot (offline). (Seit v0.1.18 ermöglichen 15-Sekunden-Statusprüfungen eine schnellere Offline-Erkennung.)

#### Fehlerbehebung

| Symptom | Ursache | Lösung |
|---------|--------|--------|
| "Verbindung fehlgeschlagen"-Fehler in Cline | Proxy nicht gestartet oder falsche Adresse | Proxy mit `curl http://localhost:56244/health` prüfen |
| Grüner Punkt erscheint nicht im Tresor | API-Schlüssel (Token) nicht konfiguriert | **:zap: Cline-Konfiguration anwenden**-Schaltfläche erneut klicken |
| Modell in der Cline-Fußzeile ändert sich nicht | Cline speichert Einstellungen im Cache | VS Code neu laden (`Ctrl+Alt+R`) |
| Falscher Modellname wird angezeigt | Alter Bug (behoben in v0.1.16) | Proxy auf v0.1.16 oder höher aktualisieren |

---

#### :purple_circle: Deploy-Befehl-kopieren-Schaltfläche — Zur Installation auf neuen Rechnern

Verwenden Sie dies bei der Erstinstallation des wall-vault-Proxy auf einem neuen Computer und der Verbindung zum Tresor. Durch Klicken wird das gesamte Installationsskript kopiert. Fügen Sie es im Terminal des neuen Computers ein und führen Sie es aus — folgendes wird auf einmal erledigt:

1. wall-vault-Binary-Installation (wird übersprungen, wenn bereits installiert)
2. Automatische systemd-Benutzer-Service-Registrierung
3. Service-Start und automatische Tresor-Verbindung

> :bulb: Das Skript enthält bereits das Token dieses Agenten und die Tresor-Server-Adresse, sodass Sie es nach dem Einfügen ohne weitere Änderungen sofort ausführen können.

---

### Servicekarte

Eine Karte zum Ein-/Ausschalten und Konfigurieren von KI-Diensten.

- Ein/Aus-Schalter pro Service
- Geben Sie die Adresse lokaler KI-Server ein (Ollama, LM Studio, vLLM usw. auf Ihrem Rechner), und verfügbare Modelle werden automatisch erkannt.
- **Lokaler Service-Verbindungsstatus**: Der Punkt neben dem Servicenamen ist **grün** bei Verbindung, **grau** wenn nicht verbunden
- **Lokale Service-Automatik** (v0.1.23+): Lokale Dienste (Ollama, LM Studio, vLLM) werden basierend auf der Erreichbarkeit automatisch aktiviert/deaktiviert. Wenn ein Dienst erreichbar wird, wechselt er innerhalb von 15 Sekunden auf grün und das Kontrollkästchen wird aktiviert; bei Verbindungsabbruch wird automatisch deaktiviert. Funktioniert wie die schlüsselbasierte Automatik bei Cloud-Diensten (Google, OpenRouter usw.).

> :bulb: **Wenn ein lokaler Dienst auf einem anderen Computer läuft**: Geben Sie die IP dieses Computers in das Service-URL-Feld ein. Beispiel: `http://192.168.1.20:11434` (Ollama), `http://192.168.1.20:1234` (LM Studio). Wenn der Dienst nur an `127.0.0.1` statt `0.0.0.0` gebunden ist, funktioniert der Zugriff über externe IP nicht — prüfen Sie die Bindungsadresse in den Diensteinstellungen.

### Admin-Token-Eingabe

Wenn Sie im Dashboard wichtige Funktionen wie das Hinzufügen oder Löschen von Schlüsseln verwenden möchten, erscheint ein Admin-Token-Eingabe-Popup. Geben Sie das im Setup-Assistenten gesetzte Token ein. Nach einmaliger Eingabe bleibt es gültig, bis Sie den Browser schließen.

> :warning: **Wenn Authentifizierungsfehler innerhalb von 15 Minuten 10 Mal überschritten werden, wird diese IP vorübergehend gesperrt.** Wenn Sie Ihr Token vergessen haben, prüfen Sie den `admin_token`-Eintrag in Ihrer `wall-vault.yaml`-Datei.

---

## Verteilter Modus (Multi-Bot)

Wenn OpenClaw gleichzeitig auf mehreren Computern betrieben wird, **teilen sich alle einen einzigen Schlüsseltresor**. Da die Schlüsselverwaltung nur an einer Stelle erfolgt, ist dies sehr praktisch.

### Beispielkonfiguration

```
[Schlüsseltresor-Server]
  wall-vault vault    (Schlüsseltresor :56243, Dashboard)

[WSL Alpha]           [Raspberry Pi Gamma]   [Mac Mini Lokal]
  wall-vault proxy      wall-vault proxy        wall-vault proxy
  openclaw TUI          openclaw TUI            openclaw TUI
  <-> SSE-Sync          <-> SSE-Sync            <-> SSE-Sync
```

Alle Bots zeigen auf den zentralen Tresor-Server, sodass Modelländerungen oder Schlüsselhinzufügungen im Tresor sofort bei allen Bots übernommen werden.

### Schritt 1: Schlüsseltresor-Server starten

Führen Sie auf dem Computer, der als Tresor-Server dient, folgenden Befehl aus:

```bash
wall-vault vault
```

### Schritt 2: Jeden Bot (Client) registrieren

Registrieren Sie die Informationen jedes Bots, der sich mit dem Tresor-Server verbinden wird:

```bash
curl -X POST http://localhost:56243/admin/clients \
  -H "Authorization: Bearer admin-token" \
  -H "Content-Type: application/json" \
  -d '{
    "id": "botA",
    "name": "Bot A",
    "token": "bota-secret",
    "default_service": "google",
    "default_model": "gemini-2.5-flash"
  }'
```

### Schritt 3: Proxy auf jedem Bot-Rechner starten

Starten Sie auf jedem Bot-Rechner den Proxy mit der Tresor-Server-Adresse und dem Token:

```bash
WV_VAULT_URL=http://192.168.x.x:56243 \
WV_VAULT_TOKEN=bota-secret \
WV_VAULT_CLIENT_ID=botA \
wall-vault proxy
```

> :bulb: Ersetzen Sie **`192.168.x.x`** durch die tatsächliche interne IP-Adresse des Tresor-Server-Rechners. Sie finden sie in den Router-Einstellungen oder über den Befehl `ip addr`.

---

## Autostart-Konfiguration

Wenn es umständlich ist, wall-vault bei jedem Neustart manuell zu starten, registrieren Sie es als Systemdienst. Nach einmaliger Registrierung startet es automatisch beim Booten.

### Linux — systemd (die meisten Linux-Distributionen)

systemd ist das System, das Programme unter Linux automatisch startet und verwaltet:

```bash
wall-vault doctor deploy
systemctl --user daemon-reload
systemctl --user enable --now wall-vault
```

Protokolle anzeigen:

```bash
journalctl --user -u wall-vault -f
```

### macOS — launchd

Das System für automatisches Starten von Programmen unter macOS:

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

## Doctor (Diagnose)

Der `doctor`-Befehl ist ein Tool, das wall-vault-Konfigurationsprobleme **selbst diagnostiziert und repariert**.

```bash
wall-vault doctor check   # Aktuellen Zustand diagnostizieren (nur lesen, nichts ändern)
wall-vault doctor fix     # Probleme automatisch reparieren
wall-vault doctor all     # Diagnose + automatische Reparatur in einem Schritt
```

> :bulb: Wenn etwas nicht zu stimmen scheint, führen Sie zuerst `wall-vault doctor all` aus. Es erkennt und behebt viele Probleme automatisch.

---

## RTK Token-Einsparung

*(v0.1.24+)*

**RTK (Token Reduction Kit)** komprimiert automatisch die Ausgabe von Shell-Befehlen, die KI-Coding-Agenten (wie Claude Code) ausführen, und reduziert so den Token-Verbrauch. Beispielsweise werden 15 Zeilen `git status`-Ausgabe auf eine 2-Zeilen-Zusammenfassung reduziert.

### Grundlegende Verwendung

```bash
# Befehle mit wall-vault rtk umhüllen, um die Ausgabe automatisch zu filtern
wall-vault rtk git status          # Zeigt nur die Liste geänderter Dateien
wall-vault rtk git diff HEAD~1     # Nur geänderte Zeilen + minimaler Kontext
wall-vault rtk git log -10         # Hash + einzeilige Nachricht je Commit
wall-vault rtk go test ./...       # Zeigt nur fehlgeschlagene Tests
wall-vault rtk ls -la              # Nicht unterstützte Befehle werden automatisch gekürzt
```

### Unterstützte Befehle und Einsparungen

| Befehl | Filtermethode | Einsparung |
|--------|-------------|------------|
| `git status` | Nur Zusammenfassung geänderter Dateien | ~87% |
| `git diff` | Geänderte Zeilen + 3 Zeilen Kontext | ~60-94% |
| `git log` | Hash + erste Zeile Nachricht | ~90% |
| `git push/pull/fetch` | Fortschritt entfernt, nur Zusammenfassung | ~80% |
| `go test` | Nur Fehler, Erfolge werden gezählt | ~88-99% |
| `go build/vet` | Nur Fehler | ~90% |
| Alle anderen Befehle | Erste 50 + letzte 50 Zeilen, max. 32 KB | Variabel |

### 3-Stufen-Filter-Pipeline

1. **Befehlsspezifischer Strukturfilter** — Versteht das Ausgabeformat von git, go usw. und extrahiert nur bedeutungsvolle Teile
2. **Regex-Nachbearbeitung** — Entfernt ANSI-Farbcodes, komprimiert Leerzeilen, fasst doppelte Zeilen zusammen
3. **Durchleitung + Kürzung** — Nicht unterstützte Befehle behalten nur die ersten/letzten 50 Zeilen

### Claude Code-Integration

Sie können den `PreToolUse`-Hook von Claude Code konfigurieren, um alle Shell-Befehle automatisch durch RTK zu leiten.

```bash
# Hook installieren (wird automatisch zu Claude Code settings.json hinzugefügt)
wall-vault rtk hook install
```

Oder manuell zu `~/.claude/settings.json` hinzufügen:

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

> :bulb: **Exit-Code-Erhaltung**: RTK gibt den Exit-Code des ursprünglichen Befehls unverändert zurück. Wenn ein Befehl fehlschlägt (Exit-Code != 0), erkennt die KI den Fehler korrekt.

> :bulb: **Englisch erzwungen**: RTK führt Befehle mit `LC_ALL=C` aus und erzeugt unabhängig von den Systemspracheinstellungen immer englische Ausgabe. Dies ist notwendig, damit die Filter korrekt funktionieren.

---

## Umgebungsvariablen-Referenz

Umgebungsvariablen sind eine Möglichkeit, Konfigurationswerte an Programme zu übergeben. Geben Sie `export VARIABLE=wert` im Terminal ein oder tragen Sie sie in Autostart-Service-Dateien ein.

| Variable | Beschreibung | Beispielwert |
|----------|-------------|-------------|
| `WV_LANG` | Dashboard-Sprache | `ko`, `en`, `ja` |
| `WV_THEME` | Dashboard-Theme | `light`, `dark`, `gold` |
| `WV_KEY_GOOGLE` | Google-API-Schlüssel (kommagetrennt für mehrere) | `AIza...,AIza...` |
| `WV_KEY_OPENROUTER` | OpenRouter-API-Schlüssel | `sk-or-v1-...` |
| `WV_VAULT_URL` | Tresor-Server-Adresse im verteilten Modus | `http://192.168.x.x:56243` |
| `WV_VAULT_TOKEN` | Client-(Bot-)Authentifizierungstoken | `my-secret-token` |
| `WV_ADMIN_TOKEN` | Admin-Token | `admin-token-here` |
| `WV_MASTER_PASS` | API-Schlüssel-Verschlüsselungspasswort | `my-password` |
| `WV_AVATAR` | Avatar-Bilddateipfad (relativ zu `~/.openclaw/`) | `workspace/avatars/avatar.png` |
| `OLLAMA_URL` | Ollama lokale Serveradresse | `http://192.168.x.x:11434` |

---

## Fehlerbehebung

### Proxy startet nicht

Der Port wird wahrscheinlich bereits von einem anderen Programm verwendet.

```bash
ss -tlnp | grep 56244   # Prüfen, wer Port 56244 verwendet
wall-vault proxy --port 8080   # Mit einem anderen Port starten
```

### API-Schlüsselfehler (429, 402, 401, 403, 582)

| Fehlercode | Bedeutung | Lösung |
|------------|----------|--------|
| **429** | Zu viele Anfragen (Kontingent überschritten) | Warten oder weitere Schlüssel hinzufügen |
| **402** | Zahlung erforderlich oder Guthaben aufgebraucht | Guthaben beim entsprechenden Dienst aufladen |
| **401 / 403** | Schlüssel falsch oder keine Berechtigung | Schlüsselwert überprüfen und neu registrieren |
| **582** | Gateway-Überlastung (5 Minuten Cooldown) | Wird nach 5 Minuten automatisch aufgehoben |

```bash
# Registrierte Schlüsselliste und Status prüfen
curl -H "Authorization: Bearer admin-token" http://localhost:56243/admin/keys

# Schlüsselverbrauchszähler zurücksetzen
curl -X POST -H "Authorization: Bearer admin-token" http://localhost:56243/admin/keys/reset
```

### Agent zeigt "Nicht verbunden" an

"Nicht verbunden" bedeutet, dass der Proxy-Prozess keine Heartbeats an den Tresor sendet. **Das bedeutet nicht, dass Einstellungen nicht gespeichert wurden.** Der Proxy muss mit der Tresor-Server-Adresse und dem Token laufen.

```bash
# Proxy mit Tresor-Server-Adresse, Token und Client-ID starten
WV_VAULT_URL=http://tresor-server:56243 \
WV_VAULT_TOKEN=client-token \
WV_VAULT_CLIENT_ID=client-id \
wall-vault proxy
```

Nach erfolgreicher Verbindung zeigt das Dashboard innerhalb von etwa 20 Sekunden :green_circle: Aktiv an.

### Ollama verbindet sich nicht

Ollama ist ein Programm, das KI direkt auf Ihrem Computer ausführt. Prüfen Sie zuerst, ob Ollama läuft.

```bash
curl http://localhost:11434/api/tags   # Wenn eine Modellliste erscheint, funktioniert es
export OLLAMA_URL=http://192.168.x.x:11434   # Wenn auf einem anderen Computer
```

> :warning: Wenn Ollama nicht antwortet, starten Sie es zuerst mit dem Befehl `ollama serve`.

> :warning: **Große Modelle sind langsam**: Große Modelle wie `qwen3.5:35b` oder `deepseek-r1` können mehrere Minuten für eine Antwort benötigen. Auch wenn es so aussieht, als würde nichts passieren, wird möglicherweise noch verarbeitet — bitte haben Sie Geduld.

---

## Letzte Änderungen (v0.1.16 ~ v0.1.27)

### v0.1.27 (2026-04-09)
- **Ollama-Fallback-Modellname behoben**: Bei Fallback von anderen Diensten zu Ollama wurden Provider-präfixierte Modellnamen (z.B. `google/gemini-3.1-pro-preview`) direkt an Ollama weitergegeben. Wird jetzt automatisch durch Umgebungsvariable/Standardmodell ersetzt.
- **Cooldown-Dauer drastisch reduziert**: 429 Rate-Limit 30Min->5Min, 402 Zahlung 1Std->30Min, 401/403 24Std->6Std. Verhindert vollständige Proxy-Lähmung, wenn alle Schlüssel gleichzeitig in Cooldown gehen.
- **Erzwungener Retry bei vollständigem Cooldown**: Wenn alle Schlüssel im Cooldown sind, wird der Schlüssel mit dem frühesten Ablauf erzwungen erneut versucht, um Anfrageverweigerungen zu verhindern.
- **Service-Liste-Anzeige behoben**: Die `/status`-Antwort zeigt jetzt die tatsächlich vom Vault synchronisierte Service-Liste (verhindert Auslassung von anthropic usw.).

### v0.1.25 (2026-04-08)
- **Agentenprozess-Erkennung**: Der Proxy erkennt, ob lokale Agenten (NanoClaw/OpenClaw) noch laufen, und zeigt eine orangefarbene Ampel im Dashboard.
- **Drag-Handle-Verbesserung**: Kartensortierung nur noch über den Ampelbereich (●) möglich. Verhindert versehentliches Ziehen von Eingabefeldern oder Schaltflächen.

### v0.1.24 (2026-04-06)
- **RTK Token-Einsparung-Unterbefehl**: `wall-vault rtk <command>` filtert Shell-Befehlsausgaben automatisch und reduziert den Token-Verbrauch von KI-Agenten um 60-90%. Enthält eingebaute Filter für Hauptbefehle wie git und go und kürzt nicht unterstützte Befehle automatisch. Transparente Integration mit Claude Code über `PreToolUse`-Hook.

### v0.1.23 (2026-04-06)
- **Ollama-Modellwechsel behoben**: Die Änderung des Ollama-Modells im Tresor-Dashboard wurde nicht auf den tatsächlichen Proxy übertragen. Zuvor wurde nur die Umgebungsvariable (`OLLAMA_MODEL`) verwendet; jetzt haben Tresor-Einstellungen Vorrang.
- **Lokale Service-Automatik**: Ollama, LM Studio und vLLM werden bei Erreichbarkeit automatisch aktiviert und bei Verbindungsabbruch deaktiviert. Funktioniert wie die schlüsselbasierte Automatik bei Cloud-Diensten.

### v0.1.22 (2026-04-05)
- **Leeres content-Feld behoben**: Wenn Thinking-Modelle (gemini-3.1-pro, o1, claude thinking usw.) alle max_tokens für Reasoning verbraucht und keine tatsächliche Antwort erzeugt haben, hat der Proxy `content`/`text`-Felder in der Antwort-JSON über `omitempty` ausgelassen, was OpenAI/Anthropic SDK-Clients mit `Cannot read properties of undefined (reading 'trim')` zum Absturz brachte. Behoben: Felder werden jetzt gemäß offizieller API-Spezifikation immer einbezogen.

### v0.1.21 (2026-04-05)
- **Gemma 4 Modell-Support**: Gemma-Serienmodelle wie `gemma-4-31b-it` und `gemma-4-26b-a4b-it` können jetzt über die Google Gemini API verwendet werden.
- **LM Studio / vLLM vollständiger Service-Support**: Zuvor fehlten diese Dienste im Proxy-Routing und fielen immer auf Ollama zurück. Jetzt korrekt über OpenAI-kompatible API geroutet.
- **Dashboard-Service-Anzeige behoben**: Auch bei Fallback zeigt das Dashboard immer den vom Benutzer konfigurierten Service an.
- **Lokaler Service-Status**: Beim Laden des Dashboards wird der Verbindungsstatus lokaler Dienste (Ollama, LM Studio, vLLM usw.) über die Punktfarbe angezeigt.
- **Tool-Filter-Umgebungsvariable**: Tool-Durchleitungsmodus mit `WV_TOOL_FILTER=passthrough`-Umgebungsvariable einstellbar.

### v0.1.20 (2026-03-28)
- **Umfassende Sicherheitshärtung**: XSS-Prävention (41 Stellen), Konstantzeit-Token-Vergleich, CORS-Einschränkungen, Anfragegrößen-Limits, Pfadtraversal-Prävention, SSE-Authentifizierung, Rate-Limiter-Härtung und 12 weitere Sicherheitsverbesserungen.

### v0.1.19 (2026-03-27)
- **Claude Code Online-Erkennung**: Claude Code-Instanzen, die nicht über den Proxy laufen, werden ebenfalls als online im Dashboard angezeigt.

### v0.1.18 (2026-03-26)
- **Fallback-Service-Festsetzung behoben**: Nach vorübergehenden Fehlern mit Ollama-Fallback erfolgt automatische Rückkehr zum ursprünglichen Dienst nach dessen Wiederherstellung.
- **Offline-Erkennung verbessert**: 15-Sekunden-Statusprüfungen ermöglichen schnellere Proxy-Ausfallerkennung.

### v0.1.17 (2026-03-25)
- **Drag-and-Drop-Kartensortierung**: Agentenkarten können per Drag-and-Drop umsortiert werden.
- **Inline-Konfigurationsanwendung**: Offline-Agenten zeigen eine [:zap: Konfiguration anwenden]-Schaltfläche.
- **cokacdir-Agententyp hinzugefügt**.

### v0.1.16 (2026-03-25)
- **Bidirektionale Modell-Synchronisation**: Modellwechsel für Cline oder Claude Code im Tresor-Dashboard werden automatisch übernommen.

---

*Für detailliertere API-Informationen siehe [API.md](API.md).*
