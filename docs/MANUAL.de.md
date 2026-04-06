# wall-vault Benutzerhandbuch
*(Zuletzt aktualisiert: 2026-04-06 — v0.1.24)*

---

## Inhaltsverzeichnis

1. [Was ist wall-vault?](#was-ist-wall-vault)
2. [Installation](#installation)
3. [Erste Schritte (Setup-Assistent)](#erste-schritte)
4. [API-Schlüssel registrieren](#api-schlüssel-registrieren)
5. [Proxy-Nutzung](#proxy-nutzung)
6. [Schlüsseltresor-Dashboard](#schlüsseltresor-dashboard)
7. [Verteilter Modus (Multi-Bot)](#verteilter-modus-multi-bot)
8. [Autostart-Einrichtung](#autostart-einrichtung)
9. [Doctor — Selbstdiagnose-Tool](#doctor--selbstdiagnose-tool)
10. [RTK Token-Einsparung](#rtk-token-einsparung)
11. [Umgebungsvariablen-Referenz](#umgebungsvariablen-referenz)
12. [Fehlerbehebung](#fehlerbehebung)

---

## Was ist wall-vault?

**wall-vault = ein AI-Proxy + API-Schlüsseltresor für OpenClaw**

Um AI-Dienste zu nutzen, benötigen Sie **API-Schlüssel** — betrachten Sie sie als **digitalen Ausweis**, der beweist, dass Sie berechtigt sind, einen bestimmten Dienst zu nutzen. Diese Ausweise haben tägliche Nutzungslimits und können bei unsachgemäßer Handhabung kompromittiert werden.

wall-vault bewahrt Ihre Ausweise sicher in einem verschlüsselten Tresor auf und fungiert als **Proxy (Vermittler)** zwischen OpenClaw und den AI-Diensten. Kurz gesagt: OpenClaw muss nur mit wall-vault kommunizieren — wall-vault erledigt den ganzen komplizierten Rest im Hintergrund.

Folgendes übernimmt wall-vault für Sie:

- **Automatische Schlüsselrotation**: Wenn ein Schlüssel sein Limit erreicht oder vorübergehend gesperrt wird (Cooldown), wechselt wall-vault still zum nächsten Schlüssel. OpenClaw arbeitet ohne Unterbrechung weiter.
- **Automatischer Service-Fallback**: Wenn Google nicht antwortet, wird auf OpenRouter ausgewichen. Wenn auch das nicht funktioniert, wird automatisch auf Ollama, LM Studio oder vLLM (lokale AI) auf Ihrem Rechner umgeschaltet. Ihre Sitzung wird nie unterbrochen. Wenn der ursprüngliche Dienst wiederhergestellt ist, wird ab der nächsten Anfrage automatisch zurückgewechselt (v0.1.18+, LM Studio/vLLM: v0.1.21+).
- **Echtzeit-Synchronisation (SSE)**: Ändern Sie das Modell im Tresor-Dashboard, und OpenClaw spiegelt es innerhalb von 1–3 Sekunden wider. SSE (Server-Sent Events) ist eine Technologie, bei der der Server Aktualisierungen in Echtzeit an Clients sendet.
- **Echtzeit-Benachrichtigungen**: Ereignisse wie Schlüsselerschöpfung oder Dienstausfälle erscheinen sofort im unteren Bereich von OpenClaws TUI (Terminal-Benutzeroberfläche).

> 💡 **Claude Code, Cursor und VS Code** können ebenfalls verbunden werden, aber der Hauptzweck von wall-vault ist die Zusammenarbeit mit OpenClaw.

```
OpenClaw (TUI-Terminal)
        │
        ▼
  wall-vault Proxy (:56244)   ← Schlüsselverwaltung, Routing, Fallback, Events
        │
        ├─ Google Gemini API
        ├─ OpenRouter API (340+ Modelle)
        ├─ Ollama / LM Studio / vLLM (lokaler Rechner, letzter Ausweg)
        └─ OpenAI / Anthropic API
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

- `curl -L ...` — Lädt eine Datei aus dem Internet herunter.
- `chmod +x` — Macht die heruntergeladene Datei ausführbar. Wenn Sie diesen Schritt überspringen, erhalten Sie einen Fehler „Berechtigung verweigert".

### Windows

Öffnen Sie PowerShell (als Administrator) und führen Sie die folgenden Befehle aus:

```powershell
# Herunterladen
Invoke-WebRequest -Uri `
  "https://github.com/sookmook/wall-vault/releases/latest/download/wall-vault-windows-amd64.exe" `
  -OutFile "$env:LOCALAPPDATA\Programs\wall-vault\wall-vault.exe"

# Zum PATH hinzufügen (wirksam nach Neustart von PowerShell)
$env:PATH += ";$env:LOCALAPPDATA\Programs\wall-vault"
```

> 💡 **Was ist PATH?** Eine Liste von Ordnern, in denen Ihr Computer nach Befehlen sucht. Wenn Sie wall-vault zum PATH hinzufügen, können Sie `wall-vault` aus jedem Verzeichnis ausführen.

### Aus Quellcode bauen (für Entwickler)

Dies gilt nur, wenn eine Go-Entwicklungsumgebung installiert ist.

```bash
git clone https://github.com/sookmook/wall-vault
cd wall-vault
make build       # bin/wall-vault (Version: v0.1.24.YYYYMMDD.HHmmss)
make install     # ~/.local/bin/wall-vault
```

> 💡 **Build-Zeitstempel-Version**: Wenn Sie mit `make build` bauen, wird die Version automatisch in einem Format wie `v0.1.24.20260406.225957` generiert, das Datum und Uhrzeit enthält. Wenn Sie direkt mit `go build ./...` bauen, wird die Version einfach als `"dev"` angezeigt.

---

## Erste Schritte

### Ausführen des Setup-Assistenten

Stellen Sie nach der Installation sicher, dass Sie zuerst den **Setup-Assistenten** ausführen. Der Assistent führt Sie Schritt für Schritt durch die Konfiguration.

```bash
wall-vault setup
```

Hier sind die Schritte, die der Assistent durchläuft:

```
1. Sprachauswahl (10 Sprachen einschließlich Deutsch)
2. Theme-Auswahl (light / dark / gold / cherry / ocean)
3. Betriebsmodus — Standalone (einzelner Benutzer) oder verteilt (mehrere Rechner)
4. Bot-Name — der im Dashboard angezeigte Name
5. Port-Einstellungen — Standards: Proxy 56244, Tresor 56243 (Enter drücken für Standards)
6. AI-Dienstauswahl — Google / OpenRouter / Ollama / LM Studio / vLLM
7. Tool-Sicherheitsfilter-Einstellungen
8. Admin-Token — ein Passwort zum Sperren der Dashboard-Verwaltungsfunktionen; kann automatisch generiert werden
9. API-Schlüssel-Verschlüsselungspasswort — für besonders sichere Schlüsselspeicherung (optional)
10. Speicherort der Konfigurationsdatei
```

> ⚠️ **Merken Sie sich unbedingt Ihr Admin-Token.** Sie benötigen es später, um Schlüssel hinzuzufügen oder Einstellungen im Dashboard zu ändern. Wenn Sie es vergessen, müssen Sie die Konfigurationsdatei manuell bearbeiten.

Nach Abschluss des Assistenten wird automatisch eine `wall-vault.yaml`-Konfigurationsdatei erstellt.

### Starten

```bash
wall-vault start
```

Zwei Server starten gleichzeitig:

- **Proxy** (`http://localhost:56244`) — der Vermittler, der OpenClaw mit AI-Diensten verbindet
- **Schlüsseltresor** (`http://localhost:56243`) — API-Schlüsselverwaltung und Web-Dashboard

Öffnen Sie `http://localhost:56243` in Ihrem Browser, um das Dashboard sofort zu sehen.

---

## API-Schlüssel registrieren

Es gibt vier Möglichkeiten, API-Schlüssel zu registrieren. **Für Anfänger wird Methode 1 (Umgebungsvariablen) empfohlen.**

### Methode 1: Umgebungsvariablen (empfohlen — am einfachsten)

Umgebungsvariablen sind **voreingestellte Werte**, die ein Programm beim Start liest. Geben Sie einfach Folgendes in Ihr Terminal ein:

```bash
# Google Gemini-Schlüssel registrieren
export WV_KEY_GOOGLE=AIzaSy...

# OpenRouter-Schlüssel registrieren
export WV_KEY_OPENROUTER=sk-or-v1-...

# Nach der Registrierung starten
wall-vault start
```

Wenn Sie mehrere Schlüssel haben, trennen Sie sie mit Kommas. wall-vault rotiert automatisch durch sie (Round-Robin):

```bash
export WV_KEY_GOOGLE=AIzaSy...,AIzaSy...,AIzaSy...
```

> 💡 **Tipp**: Der `export`-Befehl gilt nur für die aktuelle Terminal-Sitzung. Um ihn über Neustarts hinweg beizubehalten, fügen Sie die Zeile Ihrer `~/.bashrc`- oder `~/.zshrc`-Datei hinzu.

### Methode 2: Dashboard-UI (Klicken mit der Maus)

1. Öffnen Sie `http://localhost:56243` in Ihrem Browser
2. Klicken Sie auf die Schaltfläche `[+ Hinzufügen]` in der oberen **🔑 API-Schlüssel**-Karte
3. Geben Sie den Diensttyp, den Schlüsselwert, ein Label (Memo-Name) und das Tageslimit ein und speichern Sie

### Methode 3: REST API (für Automatisierung/Skripte)

REST API ist eine Methode für Programme, Daten über HTTP auszutauschen. Nützlich für automatische Registrierung per Skript.

```bash
curl -X POST http://localhost:56243/admin/keys \
  -H "Authorization: Bearer Ihr-Admin-Token" \
  -H "Content-Type: application/json" \
  -d '{
    "service": "google",
    "key": "AIzaSy...",
    "label": "Hauptschlüssel",
    "daily_limit": 1000
  }'
```

### Methode 4: Proxy-Flags (für schnelle Tests)

Verwenden Sie dies, um vorübergehend einen Schlüssel zum Testen einzufügen, ohne ihn formal zu registrieren. Der Schlüssel verschwindet, wenn das Programm beendet wird.

```bash
wall-vault proxy --key-google=AIzaSy... --key-openrouter=sk-or-...
```

---

## Proxy-Nutzung

### Verwendung mit OpenClaw (Hauptzweck)

So konfigurieren Sie OpenClaw, um sich über wall-vault mit AI-Diensten zu verbinden.

Öffnen Sie `~/.openclaw/openclaw.json` und fügen Sie Folgendes hinzu:

```json5
// ~/.openclaw/openclaw.json
{
  models: {
    providers: {
      "wall-vault": {
        baseUrl: "http://localhost:56244/v1",
        apiKey: "your-agent-token",   // Tresor-Agenten-Token
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

> 💡 **Einfachere Methode**: Klicken Sie auf die Schaltfläche **🦞 OpenClaw-Konfiguration kopieren** auf der Agenten-Karte im Dashboard — sie kopiert einen Snippet mit bereits ausgefülltem Token und Adresse. Einfach einfügen.

**Wohin leitet das Präfix `wall-vault/` in Modellnamen weiter?**

wall-vault bestimmt automatisch, an welchen AI-Dienst die Anfrage gesendet wird, basierend auf dem Modellnamen:

| Modellformat | Weiterleitung an |
|-------------|-----------------|
| `wall-vault/gemini-*` | Google Gemini direkt |
| `wall-vault/gpt-*`, `wall-vault/o3`, `wall-vault/o4*` | OpenAI direkt |
| `wall-vault/claude-*` | Anthropic über OpenRouter |
| `wall-vault/hunter-alpha`, `wall-vault/healer-alpha` | OpenRouter (kostenloser 1M-Token-Kontext) |
| `wall-vault/kimi-*`, `wall-vault/glm-*`, `wall-vault/deepseek-*` | OpenRouter |
| `google/Modellname`, `openai/Modellname`, `anthropic/Modellname` usw. | Direkt zum jeweiligen Dienst |
| `custom/google/Modellname`, `custom/openai/Modellname` usw. | Entfernt `custom/`-Präfix und leitet um |
| `Modellname:cloud` | Entfernt `:cloud`-Suffix und leitet an OpenRouter |

> 💡 **Was ist Kontext?** Die Menge an Konversation, die eine AI auf einmal speichern kann. 1M (eine Million Token) bedeutet, dass sehr lange Gespräche oder Dokumente in einer einzigen Sitzung verarbeitet werden können.

### Direktes Gemini-API-Format (für bestehende Tool-Kompatibilität)

Wenn Sie Tools haben, die Google's Gemini API direkt nutzen, ändern Sie einfach die Adresse auf wall-vault:

```bash
export ANTHROPIC_BASE_URL=http://localhost:56244/google
```

Oder wenn das Tool eine direkte URL akzeptiert:

```
http://localhost:56244/google/v1beta/models/gemini-2.5-flash:generateContent
```

### Verwendung mit dem OpenAI SDK (Python)

Sie können wall-vault auch mit Python-Code verbinden, der AI nutzt. Ändern Sie einfach die `base_url`:

```python
from openai import OpenAI

client = OpenAI(
    base_url="http://localhost:56244/v1",
    api_key="not-needed"  # wall-vault verwaltet API-Schlüssel für Sie
)

response = client.chat.completions.create(
    model="google/gemini-2.5-flash",   # verwenden Sie das provider/model-Format
    messages=[{"role": "user", "content": "Hallo"}]
)
```

### Modell zur Laufzeit ändern

Um das AI-Modell zu ändern, während wall-vault bereits läuft:

```bash
# Modell durch Senden einer Anfrage an den Proxy ändern
curl -X PUT http://localhost:56244/api/config/model \
  -H "Content-Type: application/json" \
  -d '{"service": "openrouter", "model": "anthropic/claude-3.5-sonnet"}'

# Im verteilten Modus (Multi-Bot) auf dem Tresor-Server ändern → sofort per SSE synchronisiert
curl -X PUT http://localhost:56243/admin/clients/mein-bot-id \
  -H "Authorization: Bearer Ihr-Admin-Token" \
  -H "Content-Type: application/json" \
  -d '{"default_service": "google", "default_model": "gemini-2.5-pro"}'
```

### Verfügbare Modelle prüfen

```bash
# Vollständige Liste anzeigen
curl http://localhost:56244/api/models | python3 -m json.tool

# Nur Google-Modelle anzeigen
curl "http://localhost:56244/api/models?service=google"

# Nach Name suchen (z.B. Modelle mit "claude")
curl "http://localhost:56244/api/models?q=claude"
```

**Wichtige Modelle nach Dienst:**

| Dienst | Wichtige Modelle |
|--------|-----------------|
| Google | gemini-2.5-pro, gemini-2.5-flash, gemini-2.5-flash-8b, gemini-2.0-flash |
| OpenAI | gpt-4o, gpt-4o-mini, o3, o1, o1-mini |
| OpenRouter | 346+ (Hunter Alpha 1M-Kontext kostenlos, DeepSeek R1/V3, Qwen 2.5 usw.) |
| Ollama | Erkennt lokal installierte Modelle automatisch |
| LM Studio | Lokaler Server (Port 1234) |
| vLLM | Lokaler Server (Port 8000) |

---

## Schlüsseltresor-Dashboard

Öffnen Sie `http://localhost:56243` in Ihrem Browser, um das Dashboard zu sehen.

**Layout:**
- **Obere Leiste (fixiert)**: Logo, Sprach-/Theme-Auswahl, SSE-Verbindungsstatusanzeige
- **Kartenraster**: Agenten-, Dienst- und API-Schlüssel-Karten in Kachelform angeordnet

### API-Schlüssel-Karten

Diese Karten bieten einen Überblick über Ihre registrierten API-Schlüssel.

- Schlüssel sind nach Dienst organisiert.
- `today_usage`: Anzahl der heute erfolgreich verarbeiteten Token (Texteinheiten, die die AI liest/schreibt)
- `today_attempts`: Gesamtzahl der heutigen Aufrufe (erfolgreich + fehlgeschlagen)
- Verwenden Sie die `[+ Hinzufügen]`-Schaltfläche, um neue Schlüssel zu registrieren, und `✕` zum Löschen.

> 💡 **Was ist ein Token?** Die Einheit, die AI zur Textverarbeitung verwendet. Ungefähr ein englisches Wort oder 1–2 koreanische/deutsche Zeichen. API-Preise basieren typischerweise auf der Token-Anzahl.

### Agenten-Karten

Diese Karten zeigen den Status von Bots (Agenten), die mit dem wall-vault-Proxy verbunden sind.

**Verbindungsstatus hat 4 Stufen:**

| Anzeige | Status | Bedeutung |
|---------|--------|-----------|
| 🟢 | Läuft | Proxy arbeitet normal |
| 🟡 | Verzögert | Antwortet, aber langsam |
| 🔴 | Offline | Proxy antwortet nicht |
| ⚫ | Nicht verbunden / Deaktiviert | Proxy hat sich nie mit dem Tresor verbunden oder ist deaktiviert |

**Schaltflächen am unteren Rand der Agenten-Karten:**

Wenn Sie einen Agenten mit einem bestimmten **Agententyp** registrieren, erscheinen automatisch passende Komfort-Schaltflächen.

---

#### 🔘 Konfiguration-kopieren-Schaltfläche — generiert automatisch Verbindungseinstellungen

Ein Klick auf diese Schaltfläche kopiert einen Konfigurationsschnipsel mit bereits ausgefülltem Token, Proxy-Adresse und Modellinformationen des Agenten in die Zwischenablage. Fügen Sie ihn einfach an der in der Tabelle gezeigten Stelle ein, um die Verbindungseinrichtung abzuschließen.

| Schaltfläche | Agententyp | Einfügen in |
|-------------|-----------|-------------|
| 🦞 OpenClaw-Konfiguration kopieren | `openclaw` | `~/.openclaw/openclaw.json` |
| 🦀 NanoClaw-Konfiguration kopieren | `nanoclaw` | `~/.openclaw/openclaw.json` |
| 🟠 Claude Code-Konfiguration kopieren | `claude-code` | `~/.claude/settings.json` |
| ⌨ Cursor-Konfiguration kopieren | `cursor` | Cursor → Settings → AI |
| 💻 VSCode-Konfiguration kopieren | `vscode` | `~/.continue/config.json` |

**Beispiel — Für den Claude Code-Typ wird Folgendes kopiert:**

```json
// ~/.claude/settings.json
{
  "apiProvider": "openai",
  "baseUrl": "http://192.168.0.6:56244/v1",
  "apiKey": "Token-dieses-Agenten"
}
```

**Beispiel — Für den VSCode (Continue)-Typ:**

```yaml
# ~/.continue/config.yaml  ← in config.yaml einfügen, NICHT config.json
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

**Beispiel — Für den Cursor-Typ:**

```
Base URL : http://192.168.0.6:56244/v1
API Key  : Token-dieses-Agenten

// Oder als Umgebungsvariablen:
OPENAI_BASE_URL=http://192.168.0.6:56244/v1
OPENAI_API_KEY=Token-dieses-Agenten
```

> ⚠️ **Wenn das Kopieren in die Zwischenablage nicht funktioniert**: Browser-Sicherheitsrichtlinien können das Kopieren blockieren. Wenn ein Popup mit einem Textfeld erscheint, verwenden Sie Strg+A zum Alles auswählen und dann Strg+C zum Kopieren.

---

#### ⚡ Auto-Apply-Schaltfläche — ein Klick und fertig

Für Agenten vom Typ `cline`, `claude-code`, `openclaw` oder `nanoclaw` zeigt die Agenten-Karte eine **⚡ Konfiguration anwenden**-Schaltfläche an. Ein Klick darauf aktualisiert automatisch die lokale Konfigurationsdatei des Agenten.

| Schaltfläche | Agententyp | Zieldatei |
|-------------|-----------|-----------|
| ⚡ Cline-Konfiguration anwenden | `cline` | `~/.cline/data/globalState.json` + `secrets.json` |
| ⚡ Claude Code-Konfiguration anwenden | `claude-code` | `~/.claude/settings.json` |
| ⚡ OpenClaw-Konfiguration anwenden | `openclaw` | `~/.openclaw/openclaw.json` |
| ⚡ NanoClaw-Konfiguration anwenden | `nanoclaw` | `~/.openclaw/openclaw.json` |

> ⚠️ Diese Schaltfläche sendet eine Anfrage an **localhost:56244** (lokaler Proxy). Der Proxy muss auf diesem Rechner laufen, damit es funktioniert.

---

#### 🔀 Drag & Drop Kartensortierung (v0.1.17)

Sie können Agenten-Karten im Dashboard per **Drag & Drop** in beliebiger Reihenfolge anordnen.

1. Greifen Sie eine Agenten-Karte mit der Maus und ziehen Sie sie
2. Lassen Sie sie auf einer anderen Karte los, um die Positionen zu tauschen
3. Die neue Reihenfolge wird **sofort auf dem Server gespeichert** und bleibt nach dem Aktualisieren erhalten

> 💡 Touch-Geräte (Mobilgeräte/Tablets) werden noch nicht unterstützt. Verwenden Sie einen Desktop-Browser.

---

#### 🔄 Bidirektionale Modellsynchronisation (v0.1.16)

Wenn Sie das Modell eines Agenten im Tresor-Dashboard ändern, wird die lokale Konfiguration des Agenten automatisch aktualisiert.

**Für Cline:**
- Modelländerung im Tresor → SSE-Event → Proxy aktualisiert das Modellfeld in `globalState.json`
- Aktualisierte Felder: `actModeOpenAiModelId`, `planModeOpenAiModelId`, `openAiModelId`
- `openAiBaseUrl` und API-Schlüssel werden nicht berührt
- **VS Code Neuladen erforderlich (`Strg+Alt+R` oder `Strg+Umschalt+P` → `Developer: Reload Window`)**
  - Weil Cline die Konfigurationsdatei während der Ausführung nicht neu liest

**Für Claude Code:**
- Modelländerung im Tresor → SSE-Event → Proxy aktualisiert das `model`-Feld in `settings.json`
- Durchsucht automatisch sowohl WSL- als auch Windows-Pfade (`~/.claude/`, `/mnt/c/Users/*/.claude/`)

**Umgekehrte Richtung (Agent → Tresor):**
- Wenn ein Agent (Cline, Claude Code usw.) eine Anfrage an den Proxy sendet, fügt der Proxy die Dienst-/Modellinformationen dieses Clients in den Heartbeat ein
- Die Agenten-Karte im Tresor-Dashboard zeigt den aktuell verwendeten Dienst/das Modell in Echtzeit an

> 💡 **Kernpunkt**: Der Proxy identifiziert Agenten anhand ihres Authorization-Tokens in Anfragen und leitet automatisch zum im Tresor konfigurierten Dienst/Modell weiter. Selbst wenn Cline oder Claude Code einen anderen Modellnamen sendet, überschreibt der Proxy ihn mit der Tresor-Konfiguration.

---

### Cline mit VS Code verwenden — Detaillierte Anleitung

#### Schritt 1: Cline installieren

Installieren Sie **Cline** (ID: `saoudrizwan.claude-dev`) aus dem VS Code Extensions Marketplace.

#### Schritt 2: Agent im Tresor registrieren

1. Öffnen Sie das Tresor-Dashboard (`http://Tresor-IP:56243`)
2. Klicken Sie auf **+ Hinzufügen** im Abschnitt **Agenten**
3. Füllen Sie Folgendes aus:

| Feld | Wert | Beschreibung |
|------|------|-------------|
| ID | `my_cline` | Eindeutiger Identifikator (alphanumerisch, keine Leerzeichen) |
| Name | `My Cline` | Im Dashboard angezeigter Name |
| Agententyp | `cline` | ← Muss `cline` auswählen |
| Dienst | Gewünschten Dienst auswählen (z.B. `google`) | |
| Modell | Gewünschtes Modell eingeben (z.B. `gemini-2.5-flash`) | |

4. Klicken Sie auf **Speichern** — ein Token wird automatisch generiert

#### Schritt 3: Mit Cline verbinden

**Methode A — Automatisches Anwenden (empfohlen)**

1. Stellen Sie sicher, dass der wall-vault-**Proxy** auf diesem Rechner läuft (`localhost:56244`)
2. Klicken Sie auf die **⚡ Cline-Konfiguration anwenden**-Schaltfläche auf der Agenten-Karte im Dashboard
3. Wenn die Benachrichtigung „Konfiguration erfolgreich angewendet!" erscheint, hat es funktioniert
4. VS Code neu laden (`Strg+Alt+R`)

**Methode B — Manuelle Einrichtung**

Öffnen Sie Einstellungen (⚙️) in der Cline-Seitenleiste:
- **API Provider**: `OpenAI Compatible`
- **Base URL**: `http://Proxy-Adresse:56244/v1`
  - Gleicher Rechner: `http://localhost:56244/v1`
  - Anderer Rechner (z.B. Mac Mini): `http://192.168.0.6:56244/v1`
- **API Key**: Das vom Tresor ausgestellte Token (von der Agenten-Karte kopieren)
- **Model ID**: Das im Tresor konfigurierte Modell (z.B. `gemini-2.5-flash`)

#### Schritt 4: Überprüfen

Senden Sie eine beliebige Nachricht im Cline-Chat-Fenster. Wenn es funktioniert:
- Die Agenten-Karte im Tresor-Dashboard zeigt einen **grünen Punkt (● Läuft)**
- Die Karte zeigt den aktuellen Dienst/das Modell (z.B. `google / gemini-2.5-flash`)

#### Modell ändern

Wenn Sie Clines Modell ändern möchten, tun Sie dies im **Tresor-Dashboard**:

1. Ändern Sie das Dienst-/Modell-Dropdown auf der Agenten-Karte
2. Klicken Sie auf **Anwenden**
3. VS Code neu laden (`Strg+Alt+R`) — der Modellname in Clines Fußzeile wird aktualisiert
4. Das neue Modell wird ab der nächsten Anfrage verwendet

> 💡 In der Praxis identifiziert der Proxy Clines Anfragen anhand des Tokens und leitet sie an das im Tresor konfigurierte Modell weiter. Auch ohne VS Code-Neuladen **ändert sich das tatsächlich verwendete Modell sofort** — das Neuladen dient nur zur Aktualisierung der Modellanzeige in Clines UI.

#### Verbindungsunterbrechung erkennen

Wenn VS Code geschlossen wird, wird die Agenten-Karte im Tresor-Dashboard nach etwa **90 Sekunden** gelb (verzögert) und nach **3 Minuten** rot (offline). (Ab v0.1.18 ist die Offline-Erkennung dank 15-Sekunden-Intervall-Statusprüfungen schneller.)

#### Fehlerbehebung

| Symptom | Ursache | Lösung |
|---------|---------|--------|
| „Verbindung fehlgeschlagen"-Fehler in Cline | Proxy läuft nicht oder falsche Adresse | Proxy mit `curl http://localhost:56244/health` prüfen |
| Grüner Punkt erscheint nicht im Tresor | API-Schlüssel (Token) nicht konfiguriert | **⚡ Cline-Konfiguration anwenden**-Schaltfläche erneut klicken |
| Cline-Fußzeile-Modell ändert sich nicht | Cline cacht Einstellungen | VS Code neu laden (`Strg+Alt+R`) |
| Falscher Modellname angezeigt | Alter Bug (in v0.1.16 behoben) | Proxy auf v0.1.16 oder höher aktualisieren |

---

#### 🟣 Deploy-Befehl-kopieren-Schaltfläche — für die Installation auf einem neuen Rechner

Verwenden Sie dies, wenn Sie den wall-vault-Proxy erstmals auf einem neuen Computer installieren und ihn mit dem Tresor verbinden. Ein Klick auf die Schaltfläche kopiert das gesamte Installationsskript. Fügen Sie es in das Terminal des neuen Computers ein und führen Sie es aus:

1. wall-vault-Binary installieren (wird übersprungen, wenn bereits installiert)
2. Automatisch einen systemd-Benutzerdienst registrieren
3. Dienst starten und automatisch mit dem Tresor verbinden

> 💡 Das Skript enthält bereits das Token und die Tresor-Server-Adresse dieses Agenten, sodass Sie es nach dem Einfügen sofort ohne Änderungen ausführen können.

---

### Service-Karten

Diese Karten ermöglichen das Aktivieren/Deaktivieren und Konfigurieren von AI-Diensten.

- Umschalter zum Aktivieren/Deaktivieren jedes Dienstes
- Geben Sie die Adresse eines lokalen AI-Servers (Ollama, LM Studio, vLLM usw. auf Ihrem Computer) ein, um verfügbare Modelle automatisch zu erkennen
- **Status der lokalen Dienstverbindung**: Ein ● Punkt neben dem Dienstnamen ist **grün**, wenn verbunden, **grau**, wenn nicht
- **Automatische Signalisierung lokaler Dienste** (v0.1.23+): Lokale Dienste (Ollama, LM Studio, vLLM) werden basierend auf der Verbindungsverfügbarkeit automatisch aktiviert/deaktiviert. Wenn ein Dienst erreichbar wird, wechselt er innerhalb von 15 Sekunden auf ● grün und das Kontrollkästchen wird aktiviert; wenn der Dienst ausfällt, wird er automatisch deaktiviert. Dies funktioniert genauso wie das automatische Umschalten von Cloud-Diensten (Google, OpenRouter usw.) basierend auf API-Schlüssel-Verfügbarkeit.

> 💡 **Wenn der lokale Dienst auf einem anderen Computer läuft**: Geben Sie die IP dieses Computers im URL-Feld des Dienstes ein. Beispiel: `http://192.168.0.6:11434` (Ollama), `http://192.168.0.6:1234` (LM Studio). Wenn der Dienst nur an `127.0.0.1` statt `0.0.0.0` gebunden ist, funktioniert der Zugriff über externe IP nicht — überprüfen Sie die Bindungsadress-Einstellung des Dienstes.

### Admin-Token-Eingabe

Wenn Sie wichtige Funktionen wie das Hinzufügen oder Löschen von Schlüsseln im Dashboard verwenden möchten, erscheint ein Admin-Token-Eingabe-Popup. Geben Sie das Token ein, das Sie während des Setup-Assistenten festgelegt haben. Einmal eingegeben, bleibt es gültig, bis Sie den Browser schließen.

> ⚠️ **Wenn die Authentifizierung innerhalb von 15 Minuten mehr als 10 Mal fehlschlägt, wird diese IP vorübergehend gesperrt.** Wenn Sie Ihr Token vergessen haben, überprüfen Sie das Feld `admin_token` in der Datei `wall-vault.yaml`.

---

## Verteilter Modus (Multi-Bot)

Wenn Sie OpenClaw gleichzeitig auf mehreren Computern betreiben, können Sie **einen einzelnen Schlüsseltresor gemeinsam nutzen**. Dies ist praktisch, da Sie Schlüssel nur an einer Stelle verwalten müssen.

### Beispielkonfiguration

```
[Schlüsseltresor-Server]
  wall-vault vault    (Schlüsseltresor :56243, Dashboard)

[WSL Alpha]           [Raspberry Pi Gamma]    [Mac Mini Lokal]
  wall-vault proxy      wall-vault proxy        wall-vault proxy
  openclaw TUI          openclaw TUI            openclaw TUI
  ↕ SSE-Sync            ↕ SSE-Sync              ↕ SSE-Sync
```

Alle Bots zeigen auf den zentralen Tresor-Server. Wenn Sie ein Modell ändern oder einen Schlüssel im Tresor hinzufügen, wird dies sofort auf allen Bots widergespiegelt.

### Schritt 1: Schlüsseltresor-Server starten

Führen Sie dies auf dem Computer aus, der als Tresor-Server dienen soll:

```bash
wall-vault vault
```

### Schritt 2: Jeden Bot (Client) registrieren

Registrieren Sie die Informationen für jeden Bot, der sich mit dem Tresor-Server verbinden wird:

```bash
curl -X POST http://localhost:56243/admin/clients \
  -H "Authorization: Bearer Ihr-Admin-Token" \
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

Führen Sie auf jedem Computer, auf dem ein Bot installiert ist, den Proxy mit der Tresor-Server-Adresse und dem Token aus:

```bash
WV_VAULT_URL=http://192.168.x.x:56243 \
WV_VAULT_TOKEN=bota-secret \
WV_VAULT_CLIENT_ID=botA \
wall-vault proxy
```

> 💡 Ersetzen Sie **`192.168.x.x`** durch die tatsächliche interne IP-Adresse des Tresor-Server-Computers. Sie können sie über Ihre Router-Einstellungen oder den Befehl `ip addr` herausfinden.

---

## Autostart-Einrichtung

Wenn es lästig ist, wall-vault bei jedem Neustart Ihres Computers manuell zu starten, registrieren Sie es als Systemdienst. Einmal registriert, startet es automatisch beim Hochfahren.

### Linux — systemd (die meisten Linux-Distributionen)

systemd ist das System, das Programme unter Linux automatisch startet und verwaltet:

```bash
wall-vault doctor deploy
systemctl --user daemon-reload
systemctl --user enable --now wall-vault
```

Logs prüfen:

```bash
journalctl --user -u wall-vault -f
```

### macOS — launchd

Das System, das für den automatischen Programmstart unter macOS zuständig ist:

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

## Doctor — Selbstdiagnose-Tool

Der `doctor`-Befehl ist ein Tool, das wall-vaults Konfiguration automatisch **diagnostiziert und repariert**.

```bash
wall-vault doctor check   # Aktuellen Zustand diagnostizieren (nur lesen, nichts ändern)
wall-vault doctor fix     # Probleme automatisch reparieren
wall-vault doctor all     # Diagnose + automatische Reparatur in einem Schritt
```

> 💡 Wenn etwas nicht stimmt, versuchen Sie zuerst `wall-vault doctor all` auszuführen. Es erkennt und behebt viele Probleme automatisch.

---

## RTK Token-Einsparung

*(v0.1.24+)*

**RTK (Token-Einsparungs-Tool)** komprimiert automatisch die Ausgabe von Shell-Befehlen, die von AI-Coding-Agenten (wie Claude Code) ausgeführt werden, und reduziert so den Token-Verbrauch. Zum Beispiel wird die 15-zeilige Ausgabe von `git status` auf eine 2-zeilige Zusammenfassung komprimiert.

### Grundlegende Verwendung

```bash
# Befehle mit wall-vault rtk umschließen, um die Ausgabe automatisch zu filtern
wall-vault rtk git status          # zeigt nur die Liste geänderter Dateien
wall-vault rtk git diff HEAD~1     # nur geänderte Zeilen + minimaler Kontext
wall-vault rtk git log -10         # Hash + einzeilige Nachricht
wall-vault rtk go test ./...       # zeigt nur fehlgeschlagene Tests
wall-vault rtk ls -la              # nicht unterstützte Befehle werden automatisch gekürzt
```

### Unterstützte Befehle und Einsparungen

| Befehl | Filtermethode | Einsparung |
|--------|-------------|-----------|
| `git status` | Nur Zusammenfassung geänderter Dateien | ~87% |
| `git diff` | Geänderte Zeilen + 3 Zeilen Kontext | ~60-94% |
| `git log` | Hash + erste Zeile der Nachricht | ~90% |
| `git push/pull/fetch` | Fortschritt entfernen, nur Zusammenfassung | ~80% |
| `go test` | Nur Fehler anzeigen, Erfolge zählen | ~88-99% |
| `go build/vet` | Nur Fehler anzeigen | ~90% |
| Alle anderen Befehle | Erste 50 + letzte 50 Zeilen, max. 32KB | Variabel |

### 3-Stufen-Filter-Pipeline

1. **Befehlsspezifischer Strukturfilter** — Versteht Ausgabeformate von git, go usw. und extrahiert nur bedeutungsvolle Teile
2. **Regex-Nachbearbeitung** — Entfernt ANSI-Farbcodes, komprimiert Leerzeilen, aggregiert doppelte Zeilen
3. **Durchleitung + Kürzung** — Nicht unterstützte Befehle behalten nur die ersten/letzten 50 Zeilen

### Claude Code Integration

Sie können einen Claude Code `PreToolUse`-Hook einrichten, um alle Shell-Befehle automatisch durch RTK zu leiten.

```bash
# Hook installieren (wird automatisch zur Claude Code settings.json hinzugefügt)
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

> 💡 **Exit-Code-Beibehaltung**: RTK gibt den Exit-Code des ursprünglichen Befehls unverändert zurück. Wenn ein Befehl fehlschlägt (Exit-Code ≠ 0), erkennt die AI den Fehler korrekt.

> 💡 **Erzwungene englische Ausgabe**: RTK führt Befehle mit `LC_ALL=C` aus und erzeugt unabhängig von den Systemspracheinstellungen immer englische Ausgaben. Dies stellt sicher, dass die Filter korrekt funktionieren.

---

## Umgebungsvariablen-Referenz

Umgebungsvariablen sind eine Methode, um Konfigurationswerte an ein Programm zu übergeben. Geben Sie sie im Terminal mit `export VARIABLE=Wert` ein oder fügen Sie sie Ihrer Autostart-Dienstdatei hinzu, damit sie dauerhaft angewendet werden.

| Variable | Beschreibung | Beispielwert |
|----------|-------------|-------------|
| `WV_LANG` | Dashboard-Sprache | `ko`, `en`, `ja` |
| `WV_THEME` | Dashboard-Theme | `light`, `dark`, `gold` |
| `WV_KEY_GOOGLE` | Google API-Schlüssel (kommagetrennt für mehrere) | `AIza...,AIza...` |
| `WV_KEY_OPENROUTER` | OpenRouter API-Schlüssel | `sk-or-v1-...` |
| `WV_VAULT_URL` | Tresor-Server-Adresse im verteilten Modus | `http://192.168.x.x:56243` |
| `WV_VAULT_TOKEN` | Client (Bot) Auth-Token | `my-secret-token` |
| `WV_ADMIN_TOKEN` | Admin-Token | `admin-token-here` |
| `WV_MASTER_PASS` | API-Schlüssel-Verschlüsselungspasswort | `my-password` |
| `WV_AVATAR` | Avatar-Bilddateipfad (relativ zu `~/.openclaw/`) | `workspace/avatars/avatar.png` |
| `OLLAMA_URL` | Ollama lokale Server-Adresse | `http://192.168.x.x:11434` |

---

## Fehlerbehebung

### Proxy startet nicht

Der Port wird oft bereits von einem anderen Programm verwendet.

```bash
ss -tlnp | grep 56244   # Prüfen, was Port 56244 verwendet
wall-vault proxy --port 8080   # Auf einem anderen Port starten
```

### API-Schlüssel-Fehler (429, 402, 401, 403, 582)

| Fehlercode | Bedeutung | Was tun |
|-----------|---------|--------|
| **429** | Zu viele Anfragen (Kontingent überschritten) | Warten oder weitere Schlüssel hinzufügen |
| **402** | Zahlung erforderlich oder Guthaben erschöpft | Guthaben beim jeweiligen Dienst aufladen |
| **401 / 403** | Ungültiger Schlüssel oder keine Berechtigung | Schlüsselwert überprüfen und neu registrieren |
| **582** | Gateway-Überlastung (5-Minuten-Cooldown) | Löst sich nach 5 Minuten automatisch |

```bash
# Registrierte Schlüsselliste und Status prüfen
curl -H "Authorization: Bearer Ihr-Admin-Token" http://localhost:56243/admin/keys

# Schlüssel-Nutzungszähler zurücksetzen
curl -X POST -H "Authorization: Bearer Ihr-Admin-Token" http://localhost:56243/admin/keys/reset
```

### Agent wird als „Nicht verbunden" angezeigt

„Nicht verbunden" bedeutet, dass der Proxy-Prozess keine Heartbeats an den Tresor sendet. **Es bedeutet nicht, dass die Konfiguration nicht gespeichert wurde.** Der Proxy muss mit der Tresor-Server-Adresse und dem Token laufen, um eine Verbindung herzustellen.

```bash
# Proxy mit Tresor-Server-Adresse, Token und Client-ID starten
WV_VAULT_URL=http://Tresor-Server:56243 \
WV_VAULT_TOKEN=Client-Token \
WV_VAULT_CLIENT_ID=Client-ID \
wall-vault proxy
```

Nach erfolgreicher Verbindung zeigt das Dashboard innerhalb von etwa 20 Sekunden 🟢 Läuft an.

### Ollama-Verbindungsprobleme

Ollama ist ein Programm, das AI direkt auf Ihrem Computer ausführt. Stellen Sie zunächst sicher, dass Ollama läuft.

```bash
curl http://localhost:11434/api/tags   # Wenn eine Modellliste erscheint, funktioniert es
export OLLAMA_URL=http://192.168.x.x:11434   # Wenn auf einem anderen Computer ausgeführt
```

> ⚠️ Wenn Ollama nicht antwortet, starten Sie es zuerst mit `ollama serve`.

> ⚠️ **Große Modelle antworten langsam**: Große Modelle wie `qwen3.5:35b` oder `deepseek-r1` können mehrere Minuten benötigen, um eine Antwort zu generieren. Auch wenn es so aussieht, als würde nichts passieren, wird möglicherweise noch verarbeitet — bitte haben Sie Geduld.

---

## Letzte Änderungen (v0.1.16 ~ v0.1.24)

### v0.1.24 (2026-04-06)
- **RTK Token-Einsparungs-Unterbefehl**: `wall-vault rtk <command>` filtert automatisch Shell-Befehlsausgaben, um den Token-Verbrauch von AI-Agenten um 60-90% zu reduzieren. Enthält eingebaute Filter für wichtige Befehle wie git und go und kürzt nicht unterstützte Befehle automatisch. Integration mit Claude Code über `PreToolUse`-Hook transparent möglich.

### v0.1.23 (2026-04-06)
- **Ollama-Modelländerungs-Fix**: Behebt ein Problem, bei dem die Änderung des Ollama-Modells im Tresor-Dashboard nicht im eigentlichen Proxy widergespiegelt wurde. Zuvor wurde nur die Umgebungsvariable (`OLLAMA_MODEL`) verwendet, jetzt haben Tresor-Einstellungen Vorrang.
- **Automatische Signalisierung lokaler Dienste**: Ollama, LM Studio und vLLM werden automatisch aktiviert, wenn erreichbar, und deaktiviert, wenn nicht erreichbar. Funktioniert genauso wie das schlüsselbasierte automatische Umschalten bei Cloud-Diensten.

### v0.1.22 (2026-04-05)
- **Fix für leeres Content-Feld**: Wenn Thinking-Modelle (gemini-3.1-pro, o1, claude thinking usw.) alle max_tokens für Reasoning aufbrauchen und keine eigentliche Antwort erzeugen können, hat der Proxy die `content`/`text`-Felder aus dem Antwort-JSON per `omitempty` ausgelassen, was dazu führte, dass OpenAI/Anthropic SDK-Clients mit `Cannot read properties of undefined (reading 'trim')` abstürzten. Behoben, um die Felder gemäß der offiziellen API-Spezifikation immer einzuschließen.

### v0.1.21 (2026-04-05)
- **Gemma 4 Modell-Unterstützung**: Gemma-Familienmodelle wie `gemma-4-31b-it` und `gemma-4-26b-a4b-it` können jetzt über die Google Gemini API verwendet werden.
- **LM Studio / vLLM Service-Unterstützung**: Zuvor fehlten diese Dienste im Proxy-Routing und fielen immer auf Ollama zurück. Jetzt korrekt über OpenAI-kompatible API geroutet.
- **Dashboard-Dienstanzeige-Fix**: Auch bei Fallback zeigt das Dashboard immer den vom Benutzer konfigurierten Dienst an.
- **Lokale Dienst-Statusanzeige**: Zeigt den Verbindungsstatus lokaler Dienste (Ollama, LM Studio, vLLM usw.) mit ●-Punkt-Farben beim Laden des Dashboards.
- **Tool-Filter-Umgebungsvariable**: Verwenden Sie die Umgebungsvariable `WV_TOOL_FILTER=passthrough`, um den Tool-Durchleitungsmodus einzustellen.

### v0.1.20 (2026-03-28)
- **Umfassende Sicherheitshärtung**: XSS-Prävention (41 Stellen), Konstantzeit-Token-Vergleich, CORS-Einschränkungen, Anfragegrößenlimits, Path-Traversal-Prävention, SSE-Authentifizierung, Rate-Limiter-Härtung und 12 weitere Sicherheitsverbesserungen.

### v0.1.19 (2026-03-27)
- **Claude Code Online-Erkennung**: Claude Code-Instanzen, die nicht über den Proxy laufen, werden jetzt im Dashboard als online angezeigt.

### v0.1.18 (2026-03-26)
- **Fallback-Service-Hängen-Fix**: Nach einem vorübergehenden Fehler, der einen Ollama-Fallback verursacht, kehrt es automatisch zum ursprünglichen Dienst zurück, wenn dieser wiederhergestellt wird.
- **Verbesserte Offline-Erkennung**: 15-Sekunden-Intervall-Statusprüfungen machen die Erkennung von Proxy-Ausfällen schneller.

### v0.1.17 (2026-03-25)
- **Drag & Drop Kartensortierung**: Agenten-Karten können per Drag & Drop umsortiert werden.
- **Inline-Konfiguration-anwenden-Schaltfläche**: Die [⚡ Konfiguration anwenden]-Schaltfläche wird auf Offline-Agenten-Karten angezeigt.
- **cokacdir Agententyp hinzugefügt**.

### v0.1.16 (2026-03-25)
- **Bidirektionale Modellsynchronisation**: Die Änderung eines Cline- oder Claude Code-Modells im Tresor-Dashboard wird automatisch widergespiegelt.

---

*Für detailliertere API-Informationen siehe [API.md](API.md).*
