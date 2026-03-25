# wall-vault Benutzerhandbuch
*(Zuletzt aktualisiert: 2026-03-20 — v0.1.15)*

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
- **Automatischer Dienstwechsel (Fallback)**: Antwortet Google nicht, wechselt wall-vault automatisch zu OpenRouter – und wenn das auch nicht klappt, zu Ollama (lokal auf deinem Computer). Die Sitzung bleibt erhalten.
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
        └─ Ollama (lokal auf deinem Computer, letzte Rückfalloption)
```

---

## Installation

### Linux / macOS

Öffne ein Terminal und füge den folgenden Befehl einfach ein:

```bash
# Linux (normaler PC / Server — amd64)
curl -L https://github.com/sookmook/wall-vault/releases/latest/download/wall-vault-linux-amd64 \
  -o ~/.local/bin/wall-vault && chmod +x ~/.local/bin/wall-vault

# macOS Apple Silicon (M1/M2/M3 Mac)
curl -L https://github.com/sookmook/wall-vault/releases/latest/download/wall-vault-darwin-arm64 \
  -o /usr/local/bin/wall-vault && chmod +x /usr/local/bin/wall-vault
```

- `curl -L ...` — Lädt die Datei aus dem Internet herunter.
- `chmod +x` — Macht die heruntergeladene Datei „ausführbar". Ohne diesen Schritt erscheint ein „Keine Berechtigung"-Fehler.

### Windows

Öffne PowerShell (als Administrator) und führe folgende Befehle aus:

```powershell
# Herunterladen
Invoke-WebRequest -Uri `
  "https://github.com/sookmook/wall-vault/releases/latest/download/wall-vault-windows-amd64.exe" `
  -OutFile "$env:LOCALAPPDATA\Programs\wall-vault\wall-vault.exe"

# Zum PATH hinzufügen (nach PowerShell-Neustart aktiv)
$env:PATH += ";$env:LOCALAPPDATA\Programs\wall-vault"
```

> 💡 **Was ist PATH?** PATH ist eine Liste von Ordnern, in denen der Computer nach Programmen sucht. Erst wenn wall-vault im PATH eingetragen ist, kann man `wall-vault` von jedem Ordner aus aufrufen.

### Aus dem Quellcode bauen (für Entwickler)

Dieser Abschnitt ist nur relevant, wenn eine Go-Entwicklungsumgebung installiert ist.

```bash
git clone https://github.com/sookmook/wall-vault
cd wall-vault
make build       # bin/wall-vault (Version: v0.1.6.YYYYMMDD.HHmmss)
make install     # ~/.local/bin/wall-vault
```

> 💡 **Build-Zeitstempel-Version**: Beim Bauen mit `make build` wird die Version automatisch im Format `v0.1.6.20260314.231308` mit Datum und Uhrzeit erzeugt. Beim direkten Bauen mit `go build ./...` wird die Version nur als `"dev"` angezeigt.

---

## Erste Schritte

### Setup-Assistenten starten

Nach der Installation muss beim ersten Mal unbedingt der **Setup-Assistent** ausgeführt werden. Er führt dich Schritt für Schritt durch alle erforderlichen Einstellungen.

```bash
wall-vault setup
```

Der Assistent durchläuft folgende Schritte:

```
1. Sprache auswählen (10 Sprachen, darunter Deutsch)
2. Thema auswählen (light / dark / gold / cherry / ocean)
3. Betriebsmodus — Einzelbetrieb (standalone) oder Mehrgerätebetrieb (distributed)
4. Bot-Name eingeben — wird im Dashboard angezeigt
5. Ports festlegen — Standard: Proxy 56244, Tresor 56243 (einfach Enter drücken, wenn keine Änderung nötig)
6. KI-Dienste auswählen — Google / OpenRouter / Ollama
7. Sicherheitsfilter für Werkzeuge konfigurieren
8. Admin-Token festlegen — Passwort für die Verwaltungsfunktionen im Dashboard; kann auch automatisch generiert werden
9. Verschlüsselungspasswort für API-Schlüssel (optional — für zusätzliche Sicherheit)
10. Speicherort der Konfigurationsdatei
```

> ⚠️ **Merke dir den Admin-Token unbedingt.** Er wird später benötigt, um im Dashboard Schlüssel hinzuzufügen oder Einstellungen zu ändern. Wenn du ihn verlierst, musst du die Konfigurationsdatei direkt bearbeiten.

Nach Abschluss des Assistenten wird die Konfigurationsdatei `wall-vault.yaml` automatisch erstellt.

### Starten

```bash
wall-vault start
```

Damit werden gleichzeitig zwei Server gestartet:

- **Proxy** (`http://localhost:56244`) — Vermittler zwischen OpenClaw und den KI-Diensten
- **Schlüsseltresor** (`http://localhost:56243`) — API-Schlüsselverwaltung und Web-Dashboard

Öffne `http://localhost:56243` im Browser, um das Dashboard aufzurufen.

---

## API-Schlüssel registrieren

Es gibt vier Möglichkeiten, API-Schlüssel zu registrieren. **Für Einsteiger empfehlen wir Methode 1 (Umgebungsvariable).**

### Methode 1: Umgebungsvariable (empfohlen — am einfachsten)

Eine Umgebungsvariable ist ein **vorher gespeicherter Wert**, den ein Programm beim Start ausliest. Gib im Terminal einfach Folgendes ein:

```bash
# Google Gemini-Schlüssel registrieren
export WV_KEY_GOOGLE=AIzaSy...

# OpenRouter-Schlüssel registrieren
export WV_KEY_OPENROUTER=sk-or-v1-...

# Nach der Registrierung starten
wall-vault start
```

Wenn du mehrere Schlüssel hast, verbinde sie mit einem Komma (,). wall-vault verwendet sie dann der Reihe nach automatisch (Round-Robin):

```bash
export WV_KEY_GOOGLE=AIzaSy...,AIzaSy...,AIzaSy...
```

> 💡 **Tipp**: Der `export`-Befehl gilt nur für die aktuelle Terminal-Sitzung. Damit die Einstellung auch nach einem Neustart erhalten bleibt, füge die Zeile in `~/.bashrc` oder `~/.zshrc` ein.

### Methode 2: Dashboard-Oberfläche (per Mausklick)

1. Öffne im Browser `http://localhost:56243`
2. Klicke in der Karte **🔑 API-Schlüssel** auf die Schaltfläche `[+ Hinzufügen]`
3. Gib Dienstart, Schlüsselwert, Bezeichnung (optionaler Name) und Tageslimit ein, dann speichern

### Methode 3: REST API (für Automatisierung und Skripte)

Die REST API ist eine Methode, bei der Programme über HTTP Daten austauschen. Sie eignet sich zur automatisierten Registrierung per Skript.

```bash
curl -X POST http://localhost:56243/admin/keys \
  -H "Authorization: Bearer ADMIN-TOKEN" \
  -H "Content-Type: application/json" \
  -d '{
    "service": "google",
    "key": "AIzaSy...",
    "label": "Hauptschlüssel",
    "daily_limit": 1000
  }'
```

### Methode 4: proxy-Flag (für kurze Tests)

Zum schnellen Testen ohne dauerhafte Registrierung. Der Schlüssel wird verworfen, sobald das Programm beendet wird.

```bash
wall-vault proxy --key-google=AIzaSy... --key-openrouter=sk-or-...
```

---

## Den Proxy verwenden

### Verwendung in OpenClaw (Hauptzweck)

So richtest du OpenClaw ein, damit es über wall-vault eine Verbindung zu KI-Diensten aufbaut.

Öffne die Datei `~/.openclaw/openclaw.json` und füge Folgendes hinzu:

```json5
// ~/.openclaw/openclaw.json
{
  models: {
    providers: {
      "wall-vault": {
        baseUrl: "http://localhost:56244/v1",
        apiKey: "your-agent-token",   // vault-Agent-Token
        api: "openai-completions",
        models: [
          { id: "wall-vault/gemini-2.5-flash" },
          { id: "wall-vault/gemini-2.5-pro" },
          { id: "wall-vault/hunter-alpha" },    // kostenlos, 1M Kontext
          { id: "wall-vault/claude-opus-4-6" }
        ]
      }
    }
  }
}
```

> 💡 **Einfacherer Weg**: Klicke im Dashboard auf die Schaltfläche **🦞 OpenClaw-Konfiguration kopieren** auf der Agentenkarte. Das Snippet mit dem bereits ausgefüllten Token und der Adresse wird in die Zwischenablage kopiert – einfach einfügen und fertig.

**Wohin wird eine Anfrage weitergeleitet, je nach Modellname vor `wall-vault/`?**

wall-vault erkennt anhand des Modellnamens automatisch, welcher KI-Dienst verwendet werden soll:

| Modellformat | Verbundener Dienst |
|-------------|-------------------|
| `wall-vault/gemini-*` | Google Gemini direkt |
| `wall-vault/gpt-*`, `wall-vault/o3`, `wall-vault/o4*` | OpenAI direkt |
| `wall-vault/claude-*` | Anthropic über OpenRouter |
| `wall-vault/hunter-alpha`, `wall-vault/healer-alpha` | OpenRouter (kostenlos, 1M Token Kontext) |
| `wall-vault/kimi-*`, `wall-vault/glm-*`, `wall-vault/deepseek-*` | OpenRouter |
| `google/modellname`, `openai/modellname`, `anthropic/modellname` usw. | Jeweiliger Dienst direkt |
| `custom/google/modellname`, `custom/openai/modellname` usw. | `custom/`-Teil wird entfernt, dann weitergeleitet |
| `modellname:cloud` | `:cloud`-Teil wird entfernt, dann über OpenRouter weitergeleitet |

> 💡 **Was ist Kontext (context)?** Der Kontext ist die Menge an Gesprächsinhalt, die eine KI auf einmal im „Gedächtnis" behalten kann. 1M (eine Million Token) bedeutet, dass sehr lange Gespräche oder sehr lange Dokumente auf einmal verarbeitet werden können.

### Direkte Verbindung im Gemini-API-Format (für bestehende Tools)

Wenn du ein Tool verwendest, das bisher direkt mit der Google Gemini API kommuniziert hat, ersetze einfach die Adresse durch wall-vault:

```bash
export ANTHROPIC_BASE_URL=http://localhost:56244/google
```

Oder wenn du die URL direkt angeben kannst:

```
http://localhost:56244/google/v1beta/models/gemini-2.5-flash:generateContent
```

### Verwendung mit dem OpenAI SDK (Python)

Auch in Python-Code, der KI-Funktionen verwendet, lässt sich wall-vault einbinden. Es genügt, nur `base_url` zu ändern:

```python
from openai import OpenAI

client = OpenAI(
    base_url="http://localhost:56244/v1",
    api_key="not-needed"  # API keys are managed by wall-vault
)

response = client.chat.completions.create(
    model="google/gemini-2.5-flash",   # Format: provider/model
    messages=[{"role": "user", "content": "Hallo!"}]
)
```

### Modell während des Betriebs wechseln

Um das verwendete KI-Modell zu wechseln, während wall-vault bereits läuft:

```bash
# Modell direkt am Proxy ändern
curl -X PUT http://localhost:56244/api/config/model \
  -H "Content-Type: application/json" \
  -d '{"service": "openrouter", "model": "anthropic/claude-3.5-sonnet"}'

# Im verteilten Modus (Multi-Bot): Änderung am Tresor-Server → wird sofort per SSE übernommen
curl -X PUT http://localhost:56243/admin/clients/MEIN-BOT-ID \
  -H "Authorization: Bearer ADMIN-TOKEN" \
  -H "Content-Type: application/json" \
  -d '{"default_service": "google", "default_model": "gemini-2.5-pro"}'
```

### Verfügbare Modelle anzeigen

```bash
# Alle Modelle anzeigen
curl http://localhost:56244/api/models | python3 -m json.tool

# Nur Google-Modelle anzeigen
curl "http://localhost:56244/api/models?service=google"

# Nach Name suchen (z. B. alle Modelle mit „claude")
curl "http://localhost:56244/api/models?q=claude"
```

**Übersicht der wichtigsten Modelle je Dienst:**

| Dienst | Wichtige Modelle |
|--------|-----------------|
| Google | gemini-2.5-pro, gemini-2.5-flash, gemini-2.5-flash-8b, gemini-2.0-flash |
| OpenAI | gpt-4o, gpt-4o-mini, o3, o1, o1-mini |
| OpenRouter | 346+ Modelle (Hunter Alpha 1M Kontext kostenlos, DeepSeek R1/V3, Qwen 2.5 u. v. m.) |
| Ollama | Lokal installierte Modelle werden automatisch erkannt |

---

## Das Schlüsseltresor-Dashboard

Öffne im Browser `http://localhost:56243`, um das Dashboard aufzurufen.

**Aufbau der Oberfläche:**
- **Fixierte Titelleiste (Topbar)**: Logo, Sprach- und Thema-Auswahl, SSE-Verbindungsstatus
- **Kachelraster**: Agenten-, Dienst- und API-Schlüsselkarten sind als Kacheln angeordnet

### API-Schlüssel-Karte

Eine Karte, mit der sich alle registrierten API-Schlüssel auf einen Blick verwalten lassen.

- Zeigt die Schlüsselliste nach Dienst geordnet an.
- `today_usage`: Anzahl der heute erfolgreich verarbeiteten Token (Zeichen, die die KI gelesen und geschrieben hat)
- `today_attempts`: Gesamtanzahl der heutigen Aufrufe (Erfolg + Fehler)
- Mit `[+ Hinzufügen]` neuen Schlüssel registrieren, mit `✕` löschen.

> 💡 **Was ist ein Token?** Ein Token ist die Einheit, in der KI-Systeme Text verarbeiten. Grob gesagt entspricht ein Token etwa einem englischen Wort oder 1–2 Buchstaben. API-Kosten werden in der Regel nach der Anzahl der Token berechnet.

### Agenten-Karte

Zeigt den Status der Bots (Agenten), die mit dem wall-vault-Proxy verbunden sind.

**Der Verbindungsstatus wird in vier Stufen angezeigt:**

| Anzeige | Status | Bedeutung |
|---------|--------|-----------|
| 🟢 | Aktiv | Proxy läuft einwandfrei |
| 🟡 | Verzögert | Antwort kommt, aber langsam |
| 🔴 | Offline | Proxy antwortet nicht |
| ⚫ | Nicht verbunden / Inaktiv | Proxy hat sich noch nie verbunden oder ist deaktiviert |

**Schaltflächen am unteren Rand der Agentenkarte:**

Wenn du beim Registrieren eines Agenten den **Agenten-Typ** angibst, erscheinen automatisch passende Schnellzugriff-Schaltflächen.

---

#### 🔘 Konfiguration kopieren – Verbindungseinstellungen automatisch erstellen

Ein Klick auf die Schaltfläche kopiert ein fertiges Konfigurations-Snippet mit dem Token, der Proxy-Adresse und den Modellinformationen dieses Agenten in die Zwischenablage. Einfach an der richtigen Stelle einfügen – fertig.

| Schaltfläche | Agenten-Typ | Einfügen in |
|-------------|------------|------------|
| 🦞 OpenClaw-Konfiguration kopieren | `openclaw` | `~/.openclaw/openclaw.json` |
| 🦀 NanoClaw-Konfiguration kopieren | `nanoclaw` | `~/.openclaw/openclaw.json` |
| 🟠 Claude Code-Konfiguration kopieren | `claude-code` | `~/.claude/settings.json` |
| ⌨ Cursor-Konfiguration kopieren | `cursor` | Cursor → Settings → AI |
| 💻 VSCode-Konfiguration kopieren | `vscode` | `~/.continue/config.json` |

**Beispiel – bei Typ Claude Code wird Folgendes kopiert:**

```json
// ~/.claude/settings.json
{
  "apiProvider": "openai",
  "baseUrl": "http://192.168.0.6:56244/v1",
  "apiKey": "TOKEN-DIESES-AGENTEN"
}
```

**Beispiel – bei Typ VSCode (Continue):**

```json
// ~/.continue/config.json
{
  "models": [{
    "title": "wall-vault proxy",
    "provider": "openai",
    "model": "gemini-2.5-flash",
    "apiBase": "http://192.168.0.6:56244/v1",
    "apiKey": "TOKEN-DIESES-AGENTEN"
  }]
}
```

**Beispiel – bei Typ Cursor:**

```
Base URL : http://192.168.0.6:56244/v1
API Key  : TOKEN-DIESES-AGENTEN

// Oder als Umgebungsvariable:
OPENAI_BASE_URL=http://192.168.0.6:56244/v1
OPENAI_API_KEY=TOKEN-DIESES-AGENTEN
```

> ⚠️ **Wenn das Kopieren in die Zwischenablage nicht funktioniert**: Browser-Sicherheitsrichtlinien können das Kopieren blockieren. Wenn sich ein Popup mit einem Textfeld öffnet, wähle den gesamten Text mit Strg+A aus und kopiere ihn mit Strg+C.

---

#### ⚡ Automatische Anwendung – ein Klick und die Einrichtung ist erledigt

Wenn der Agenten-Typ `cline`, `claude-code`, `openclaw` oder `nanoclaw` ist, erscheint auf der Agentenkarte die Schaltfläche **⚡ Einstellungen anwenden**. Ein Klick darauf aktualisiert automatisch die lokale Konfigurationsdatei des Agenten.

| Schaltfläche | Agenten-Typ | Zieldatei |
|-------------|------------|-----------|
| ⚡ Cline-Einstellungen anwenden | `cline` | `~/.cline/data/globalState.json` + `secrets.json` |
| ⚡ Claude Code-Einstellungen anwenden | `claude-code` | `~/.claude/settings.json` |
| ⚡ OpenClaw-Einstellungen anwenden | `openclaw` | `~/.openclaw/openclaw.json` |
| ⚡ NanoClaw-Einstellungen anwenden | `nanoclaw` | `~/.openclaw/openclaw.json` |

> ⚠️ Diese Schaltfläche sendet eine Anfrage an **localhost:56244** (lokaler Proxy). Der Proxy muss auf diesem Rechner laufen, damit es funktioniert.

---

#### 🔀 Kartensortierung per Drag & Drop (v0.1.17)

Sie können die Agentenkarten im Dashboard per **Drag & Drop** in die gewünschte Reihenfolge bringen.

1. Greifen Sie eine Agentenkarte mit der Maus und ziehen Sie sie
2. Legen Sie sie auf der Karte an der gewünschten Position ab, um die Reihenfolge zu ändern
3. Die geänderte Reihenfolge wird **sofort auf dem Server gespeichert** und bleibt auch nach dem Neuladen der Seite erhalten

> 💡 Touchgeräte (Mobilgeräte/Tablets) werden noch nicht unterstützt. Bitte verwenden Sie einen Desktop-Browser.

---

#### 🔄 Bidirektionale Modellsynchronisierung (v0.1.16)

Wenn du im Tresor-Dashboard das Modell eines Agenten änderst, wird die lokale Konfiguration dieses Agenten automatisch aktualisiert.

**Bei Cline:**
- Modelländerung im Tresor → SSE-Event → der Proxy aktualisiert das Modellfeld in `globalState.json`
- Aktualisierte Felder: `actModeOpenAiModelId`, `planModeOpenAiModelId`, `openAiModelId`
- `openAiBaseUrl` und der API-Schlüssel werden nicht verändert
- **Ein Neuladen von VS Code ist erforderlich (`Strg+Alt+R` oder `Strg+Umschalt+P` → `Developer: Reload Window`)**
  - Cline liest die Konfigurationsdatei während der Ausführung nicht erneut ein

**Bei Claude Code:**
- Modelländerung im Tresor → SSE-Event → der Proxy aktualisiert das Feld `model` in `settings.json`
- WSL- und Windows-Pfade werden automatisch durchsucht (`~/.claude/`, `/mnt/c/Users/*/.claude/`)

**Umgekehrte Richtung (Agent → Tresor):**
- Wenn ein Agent (Cline, Claude Code usw.) eine Anfrage an den Proxy sendet, nimmt der Proxy die Service- und Modellinformationen des Clients in den Heartbeat auf
- Der aktuell verwendete Dienst und das Modell werden in Echtzeit auf der Agentenkarte im Dashboard angezeigt

> 💡 **Kernpunkt**: Der Proxy identifiziert den Agenten anhand des Authorization-Tokens in der Anfrage und leitet automatisch an den im Tresor konfigurierten Dienst/Modell weiter. Selbst wenn Cline oder Claude Code einen anderen Modellnamen sendet, überschreibt der Proxy ihn mit der Tresor-Konfiguration.

---

### Cline in VS Code verwenden — detaillierte Anleitung

#### Schritt 1: Cline installieren

Installiere **Cline** (ID: `saoudrizwan.claude-dev`) aus dem VS Code Extension Marketplace.

#### Schritt 2: Agenten im Tresor registrieren

1. Öffne das Tresor-Dashboard (`http://Tresor-IP:56243`)
2. Klicke im Abschnitt **Agenten** auf **+ Hinzufügen**
3. Fülle die folgenden Felder aus:

| Feld | Wert | Beschreibung |
|------|------|-------------|
| ID | `mein_cline` | Eindeutiger Bezeichner (Buchstaben, keine Leerzeichen) |
| Name | `Mein Cline` | Im Dashboard angezeigter Name |
| Agenten-Typ | `cline` | ← unbedingt `cline` auswählen |
| Dienst | Gewünschten Dienst auswählen (z. B. `google`) | |
| Modell | Gewünschtes Modell eingeben (z. B. `gemini-2.5-flash`) | |

4. Klicke auf **Speichern** — ein Token wird automatisch generiert

#### Schritt 3: Cline verbinden

**Methode A — Automatische Anwendung (empfohlen)**

1. Stelle sicher, dass der wall-vault-**Proxy** auf diesem Rechner läuft (`localhost:56244`)
2. Klicke auf der Agentenkarte auf **⚡ Cline-Einstellungen anwenden**
3. Wenn die Meldung „Einstellungen angewendet!" erscheint, war es erfolgreich
4. Lade VS Code neu (`Strg+Alt+R`)

**Methode B — Manuelle Konfiguration**

Öffne die Einstellungen (⚙️) in der Cline-Seitenleiste und konfiguriere:
- **API Provider**: `OpenAI Compatible`
- **Base URL**: `http://Proxy-Adresse:56244/v1`
  - Auf demselben Rechner: `http://localhost:56244/v1`
  - Auf einem anderen Gerät (z. B. Mac Mini): `http://192.168.0.6:56244/v1`
- **API Key**: Der vom Tresor ausgestellte Token (von der Agentenkarte kopiert)
- **Model ID**: Das im Tresor eingestellte Modell (z. B. `gemini-2.5-flash`)

#### Schritt 4: Überprüfung

Sende eine beliebige Nachricht im Cline-Chat. Wenn alles funktioniert:
- Auf der Agentenkarte im Dashboard erscheint ein **grüner Punkt (● Aktiv)**
- Auf der Karte werden der aktuelle Dienst und das Modell angezeigt (z. B. `google / gemini-2.5-flash`)

#### Modell wechseln

Um das Modell von Cline zu ändern, ändere es im **Tresor-Dashboard**:

1. Ändere Dienst/Modell im Dropdown-Menü der Agentenkarte
2. Klicke auf **Anwenden**
3. Lade VS Code neu (`Strg+Alt+R`) — der Modellname in der Cline-Fußzeile wird aktualisiert
4. Ab der nächsten Anfrage wird das neue Modell verwendet

> 💡 Tatsächlich identifiziert der Proxy Cline-Anfragen anhand des Tokens und leitet sie an das im Tresor konfigurierte Modell weiter. Selbst ohne VS Code neu zu laden, **ändert sich das tatsächlich verwendete Modell sofort** — das Neuladen dient nur dazu, die Modellanzeige in der Cline-Oberfläche zu aktualisieren.

#### Verbindungsabbruch erkennen

Wenn du VS Code schließt, wechselt die Agentenkarte im Dashboard nach etwa **2–3 Minuten** auf Gelb (Verzögert) und nach **5 Minuten** auf Rot (Offline).

#### Problemlösung

| Symptom | Ursache | Lösung |
|---------|---------|--------|
| Fehler „Verbindung fehlgeschlagen" in Cline | Proxy nicht gestartet oder Adresse falsch | Proxy prüfen mit `curl http://localhost:56244/health` |
| Grüner Punkt erscheint nicht im Tresor | API-Schlüssel (Token) nicht konfiguriert | **⚡ Cline-Einstellungen anwenden** erneut klicken |
| Modellname in der Cline-Fußzeile ändert sich nicht | Cline speichert die Konfiguration im Cache | VS Code neu laden (`Strg+Alt+R`) |
| Ein falscher Modellname wird angezeigt | Alter Bug (behoben in v0.1.16) | Proxy auf Version v0.1.16 oder höher aktualisieren |

---

#### 🟣 Deployment-Befehl kopieren – für die Installation auf einem neuen Gerät

Diese Schaltfläche ist nützlich, wenn wall-vault zum ersten Mal auf einem neuen Computer installiert und mit dem Tresor verbunden werden soll. Ein Klick kopiert das gesamte Installations-Skript. Einfach im Terminal des neuen Computers einfügen und ausführen – folgendes wird dann automatisch erledigt:

1. wall-vault-Binary installieren (wird übersprungen, wenn bereits installiert)
2. systemd-Benutzerdienst automatisch registrieren
3. Dienst starten und automatisch mit dem Tresor verbinden

> 💡 Das Skript enthält bereits den Token und die Tresor-Serveradresse dieses Agenten, sodass nach dem Einfügen keine weitere Anpassung nötig ist.

---

### Dienste-Karte

Eine Karte zum Aktivieren, Deaktivieren und Konfigurieren der KI-Dienste.

- Ein-/Aus-Schalter je Dienst
- Wenn du die Adresse eines lokalen KI-Servers (z. B. Ollama, LM Studio oder vLLM, der auf deinem eigenen Computer läuft) eingibst, werden verfügbare Modelle automatisch erkannt.
- **Verbindungsstatus für lokale Dienste**: Der Punkt ● neben dem Dienstnamen ist **grün**, wenn verbunden, und **grau**, wenn nicht verbunden.
- **Automatische Checkbox-Synchronisierung**: Wenn beim Öffnen der Seite ein lokaler Dienst (z. B. Ollama) läuft, wird er automatisch als aktiv markiert.

> 💡 **Wenn ein lokaler Dienst auf einem anderen Computer läuft**: Gib die IP-Adresse dieses Computers in das URL-Feld des Dienstes ein. Beispiel: `http://192.168.0.6:11434` (Ollama), `http://192.168.0.6:1234` (LM Studio)

### Admin-Token eingeben

Wenn du im Dashboard eine wichtige Aktion (z. B. Schlüssel hinzufügen oder löschen) ausführen möchtest, erscheint ein Popup zur Eingabe des Admin-Tokens. Gib den Token ein, den du im Setup-Assistenten festgelegt hast. Er bleibt bis zum Schließen des Browsers gespeichert.

> ⚠️ **Wenn innerhalb von 15 Minuten mehr als 10 Anmeldeversuche fehlschlagen, wird die betreffende IP-Adresse vorübergehend gesperrt.** Falls du deinen Token vergessen hast, findest du ihn in der `wall-vault.yaml`-Datei unter dem Eintrag `admin_token`.

---

## Verteilter Modus (Multi-Bot)

Wenn du OpenClaw gleichzeitig auf mehreren Computern betreibst, kannst du **einen gemeinsamen Schlüsseltresor** für alle nutzen. Das vereinfacht die Schlüsselverwaltung erheblich, da alles zentral an einem Ort verwaltet wird.

### Beispielaufbau

```
[Schlüsseltresor-Server]
  wall-vault vault    (Tresor :56243, Dashboard)

[WSL Alpha]            [Raspberry Pi Gamma]    [Mac Mini Lokal]
  wall-vault proxy      wall-vault proxy          wall-vault proxy
  openclaw TUI          openclaw TUI              openclaw TUI
  ↕ SSE-Sync            ↕ SSE-Sync                ↕ SSE-Sync
```

Alle Bots richten sich nach dem zentralen Tresor-Server. Wenn du dort ein Modell wechselst oder einen Schlüssel hinzufügst, wird die Änderung sofort auf alle Bots übertragen.

### Schritt 1: Tresor-Server starten

Führe folgenden Befehl auf dem Computer aus, der als Tresor-Server dienen soll:

```bash
wall-vault vault
```

### Schritt 2: Bots (Clients) registrieren

Registriere vorab die Informationen aller Bots, die sich mit dem Tresor-Server verbinden sollen:

```bash
curl -X POST http://localhost:56243/admin/clients \
  -H "Authorization: Bearer ADMIN-TOKEN" \
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

Starte auf jedem Computer, auf dem ein Bot läuft, den Proxy mit der Angabe von Tresor-Serveradresse und Token:

```bash
WV_VAULT_URL=http://192.168.x.x:56243 \
WV_VAULT_TOKEN=bota-secret \
WV_VAULT_CLIENT_ID=botA \
wall-vault proxy
```

> 💡 Ersetze **`192.168.x.x`** durch die tatsächliche lokale IP-Adresse des Tresor-Servers. Diese findest du in den Router-Einstellungen oder mit dem Befehl `ip addr`.

---

## Autostart einrichten

Wenn es lästig ist, wall-vault nach jedem Neustart manuell zu starten, registriere es als Systemdienst. Nach einmaliger Einrichtung startet es automatisch beim Hochfahren.

### Linux — systemd (die meisten Linux-Distributionen)

systemd ist das System, das unter Linux Programme automatisch startet und verwaltet:

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

launchd ist das System für den automatischen Programmstart unter macOS:

```bash
wall-vault doctor deploy launchd
launchctl load ~/Library/LaunchAgents/com.wall-vault.plist
```

### Windows — NSSM

1. Lade NSSM von [nssm.cc](https://nssm.cc/download) herunter und füge es zum PATH hinzu.
2. Öffne PowerShell als Administrator und führe aus:

```powershell
wall-vault doctor deploy windows
```

---

## Doctor-Diagnose

Der `doctor`-Befehl ist ein **Selbstdiagnose- und Reparaturtool**, das prüft, ob wall-vault korrekt eingerichtet ist.

```bash
wall-vault doctor check   # Aktuellen Zustand prüfen (nur lesen, keine Änderungen)
wall-vault doctor fix     # Probleme automatisch beheben
wall-vault doctor all     # Diagnose + automatische Reparatur in einem Schritt
```

> 💡 Wenn etwas nicht zu stimmen scheint, führe zuerst `wall-vault doctor all` aus. Viele Probleme werden damit automatisch behoben.

---

## Umgebungsvariablen – Übersicht

Umgebungsvariablen sind eine Methode, um einem Programm Konfigurationswerte zu übergeben. Du kannst sie im Terminal mit `export VARIABLENNAME=WERT` eingeben oder in die Autostart-Dienstdatei eintragen, damit sie dauerhaft gelten.

| Variable | Beschreibung | Beispielwert |
|----------|-------------|-------------|
| `WV_LANG` | Dashboard-Sprache | `ko`, `en`, `de` |
| `WV_THEME` | Dashboard-Thema | `light`, `dark`, `gold` |
| `WV_KEY_GOOGLE` | Google API-Schlüssel (mehrere mit Komma) | `AIza...,AIza...` |
| `WV_KEY_OPENROUTER` | OpenRouter API-Schlüssel | `sk-or-v1-...` |
| `WV_VAULT_URL` | Tresor-Serveradresse im verteilten Modus | `http://192.168.x.x:56243` |
| `WV_VAULT_TOKEN` | Authentifizierungs-Token für den Client (Bot) | `my-secret-token` |
| `WV_ADMIN_TOKEN` | Admin-Token | `admin-token-here` |
| `WV_MASTER_PASS` | Verschlüsselungspasswort für API-Schlüssel | `my-password` |
| `WV_AVATAR` | Pfad zur Avatar-Bilddatei (relativ zu `~/.openclaw/`) | `workspace/avatars/avatar.png` |
| `OLLAMA_URL` | Adresse des lokalen Ollama-Servers | `http://192.168.x.x:11434` |

---

## Problemlösung

### Proxy startet nicht

Meistens wird der Port bereits von einem anderen Programm verwendet.

```bash
ss -tlnp | grep 56244   # Prüfen, welches Programm Port 56244 belegt
wall-vault proxy --port 8080   # Mit einem anderen Port starten
```

### API-Schlüsselfehler (429, 402, 401, 403, 582)

| Fehlercode | Bedeutung | Lösung |
|-----------|-----------|--------|
| **429** | Zu viele Anfragen (Limit überschritten) | Kurz warten oder weiteren Schlüssel hinzufügen |
| **402** | Zahlung erforderlich oder Guthaben aufgebraucht | Guthaben beim jeweiligen Dienst aufladen |
| **401 / 403** | Schlüssel ungültig oder fehlende Berechtigung | Schlüsselwert prüfen und neu registrieren |
| **582** | Gateway überlastet (Cooldown 5 Minuten) | Wird nach 5 Minuten automatisch aufgehoben |

```bash
# Registrierte Schlüssel und Status anzeigen
curl -H "Authorization: Bearer ADMIN-TOKEN" http://localhost:56243/admin/keys

# Nutzungszähler der Schlüssel zurücksetzen
curl -X POST -H "Authorization: Bearer ADMIN-TOKEN" http://localhost:56243/admin/keys/reset
```

### Agent wird als „Nicht verbunden" angezeigt

„Nicht verbunden" bedeutet, dass der Proxy-Prozess dem Tresor kein Herzschlagsignal (Heartbeat) schickt. **Das heißt nicht, dass Einstellungen verloren gegangen sind.** Der Proxy muss mit der Tresor-Serveradresse und dem Token gestartet werden, damit der Status auf „Verbunden" wechselt.

```bash
# Proxy mit Tresor-Serveradresse, Token und Client-ID starten
WV_VAULT_URL=http://TRESOR-SERVER-ADRESSE:56243 \
WV_VAULT_TOKEN=CLIENT-TOKEN \
WV_VAULT_CLIENT_ID=CLIENT-ID \
wall-vault proxy
```

Nach erfolgreicher Verbindung wird der Status im Dashboard innerhalb von etwa 20 Sekunden auf 🟢 Aktiv wechseln.

### Ollama-Verbindung schlägt fehl

Ollama ist ein Programm, das KI-Modelle direkt auf deinem Computer ausführt. Prüfe zuerst, ob Ollama läuft.

```bash
curl http://localhost:11434/api/tags   # Wenn eine Modellliste erscheint, ist alles in Ordnung
export OLLAMA_URL=http://192.168.x.x:11434   # Falls Ollama auf einem anderen Computer läuft
```

> ⚠️ Wenn Ollama nicht antwortet, starte es zuerst mit dem Befehl `ollama serve`.

> ⚠️ **Große Modelle brauchen Zeit**: Modelle wie `qwen3.5:35b` oder `deepseek-r1` können mehrere Minuten benötigen, um eine Antwort zu generieren. Wenn es so aussieht, als würde nichts passieren, ist die Verarbeitung wahrscheinlich trotzdem im Gange – bitte warte geduldig.

---

*Ausführlichere API-Informationen findest du in [API.md](API.md).*
