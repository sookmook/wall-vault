# wall-vault API-Handbuch

Dieses Dokument beschreibt alle HTTP-API-Endpunkte von wall-vault im Detail.

---

## Inhaltsverzeichnis

- [Authentifizierung](#authentifizierung)
- [Proxy-API (:56244)](#proxy-api-56244)
  - [Gesundheitsprüfung](#get-health)
  - [Status](#get-status)
  - [Modellliste](#get-apimodels)
  - [Modell ändern](#put-apiconfigmodel)
  - [Denkmodus](#put-apiconfigthink-mode)
  - [Konfiguration neu laden](#post-reload)
  - [Gemini-API](#post-googlev1betamodelsmgeneratecontent)
  - [Gemini-Streaming](#post-googlev1betamodelsmstreamgeneratecontent)
  - [OpenAI-kompatible API](#post-v1chatcompletions)
- [Schlüsseltresor-API (:56243)](#schlüsseltresor-api-56243)
  - [Öffentliche API](#öffentliche-api-keine-authentifizierung-erforderlich)
  - [SSE-Ereignisstrom](#get-apievents)
  - [Nur-Proxy-API](#nur-proxy-api-client-token)
  - [Admin-API — Schlüssel](#admin-api--api-schlüssel)
  - [Admin-API — Clients](#admin-api--clients)
  - [Admin-API — Dienste](#admin-api--dienste)
  - [Admin-API — Modellliste](#admin-api--modellliste)
  - [Admin-API — Proxy-Status](#admin-api--proxy-status)
- [SSE-Ereignistypen](#sse-ereignistypen)
- [Anbieter-/Modell-Routing](#anbietermodell-routing)
- [Datenschema](#datenschema)
- [Fehlerantworten](#fehlerantworten)
- [cURL-Beispiele](#curl-beispiele)

---

## Authentifizierung

| Bereich | Methode | Header |
|---------|---------|--------|
| Admin-API | Bearer-Token | `Authorization: Bearer <admin_token>` |
| Proxy → Tresor | Bearer-Token | `Authorization: Bearer <client_token>` |
| Proxy-API | Keine (lokal) | — |

Wenn `admin_token` nicht gesetzt ist (leerer String), sind alle Admin-APIs ohne Authentifizierung zugänglich.

### Sicherheitsrichtlinie

- **Rate Limiting**: Bei mehr als 10 fehlgeschlagenen Admin-API-Authentifizierungen innerhalb von 15 Minuten wird die IP vorübergehend gesperrt (`429 Too Many Requests`)
- **IP-Whitelist**: Nur IPs/CIDRs, die im `ip_whitelist`-Feld des Agenten (`Client`) registriert sind, erhalten Zugriff auf `/api/keys`. Bei leerem Array sind alle IPs erlaubt.
- **theme/lang-Schutz**: `/admin/theme` und `/admin/lang` erfordern ebenfalls Admin-Token-Authentifizierung

---

## Proxy-API (:56244)

Der Server, auf dem der Proxy läuft. Standardport `56244`.

---

### `GET /health`

Gesundheitsprüfung. Gibt immer 200 OK zurück.

**Antwortbeispiel:**
```json
{
  "status": "ok",
  "version": "v0.1.6.20260314.231308",
  "client": "bot-a"
}
```

---

### `GET /status`

Detaillierter Proxy-Status.

**Antwortbeispiel:**
```json
{
  "status": "ok",
  "version": "v0.1.6.20260314.231308",
  "client": "bot-a",
  "service": "google",
  "model": "gemini-2.5-flash",
  "sse": true,
  "filter": "strip_all",
  "services": ["google", "openrouter", "ollama"],
  "mode": "distributed"
}
```

| Feld | Typ | Beschreibung |
|------|-----|-------------|
| `service` | string | Aktueller Standarddienst |
| `model` | string | Aktuelles Standardmodell |
| `sse` | bool | Ob die Tresor-SSE-Verbindung besteht |
| `filter` | string | Werkzeugfiltermodus |
| `services` | []string | Liste aktiver Dienste |
| `mode` | string | `standalone` \| `distributed` |

---

### `GET /api/models`

Verfügbare Modelle auflisten. Verwendet TTL-Cache (Standard 10 Minuten).

**Abfrageparameter:**

| Parameter | Beschreibung | Beispiel |
|-----------|-------------|----------|
| `service` | Dienstfilter | `?service=google` |
| `q` | Suche nach Modell-ID/Name | `?q=gemini` |

**Antwortbeispiel:**
```json
{
  "models": [
    {
      "id": "gemini-2.5-pro",
      "name": "Gemini 2.5 Pro",
      "service": "google",
      "context_length": 1048576,
      "free": false
    },
    {
      "id": "openrouter/hunter-alpha",
      "name": "Hunter Alpha (1M ctx, free)",
      "service": "openrouter",
      "context_length": 1048576,
      "free": true
    }
  ],
  "count": 2
}
```

| Feld | Typ | Beschreibung |
|------|-----|-------------|
| `id` | string | Modell-ID |
| `name` | string | Anzeigename des Modells |
| `service` | string | `google` \| `openai` \| `anthropic` \| `openrouter` \| `ollama` usw. |
| `context_length` | int | Kontextfenstergröße |
| `free` | bool | Ob es ein kostenloses Modell ist (OpenRouter) |

---

### `PUT /api/config/model`

Aktuellen Dienst und Modell ändern.

**Anfragekörper:**
```json
{
  "service": "google",
  "model": "gemini-2.5-flash"
}
```

**Antwort:**
```json
{
  "status": "ok",
  "service": "google",
  "model": "gemini-2.5-flash"
}
```

> **Hinweis:** Im verteilten Modus wird empfohlen, `PUT /admin/clients/{id}` des Tresors anstelle dieser API zu verwenden. Tresoränderungen werden automatisch über SSE innerhalb von 1–3 Sekunden übernommen.

---

### `PUT /api/config/think-mode`

Denkmodus umschalten (No-Op, für zukünftige Erweiterung reserviert).

**Antwort:**
```json
{"status": "ok"}
```

---

### `POST /reload`

Client-Einstellungen und Schlüssel sofort vom Tresor neu synchronisieren.

**Antwort:**
```json
{"status": "reloading"}
```

Die Neusynchronisierung läuft asynchron und wird innerhalb von 1–2 Sekunden nach Empfang der Antwort abgeschlossen.

---

### `POST /google/v1beta/models/{model}:generateContent`

Gemini-API-Proxy (nicht-streamend).

**Pfadparameter:**
- `{model}`: Modell-ID. Bei `gemini-`-Präfix wird automatisch der Google-Dienst gewählt.

**Anfragekörper:** [Gemini generateContent-Anfrageformat](https://ai.google.dev/api/generate-content)

```json
{
  "contents": [
    {
      "role": "user",
      "parts": [{"text": "안녕하세요"}]
    }
  ],
  "generationConfig": {
    "temperature": 0.7,
    "maxOutputTokens": 1024
  }
}
```

**Antwortkörper:** Gemini generateContent-Antwortformat

**Werkzeugfilter:** Bei `tool_filter: strip_all` wird das `tools`-Array in der Anfrage automatisch entfernt.

**Fallback-Kette:** Wenn der zugewiesene Dienst fehlschlägt → Fallback in konfigurierter Dienstreihenfolge → Ollama (zuletzt).

---

### `POST /google/v1beta/models/{model}:streamGenerateContent`

Gemini-API-Streaming-Proxy. Das Anfrageformat ist identisch mit dem nicht-streamenden Format. Die Antwort ist ein SSE-Strom:

```
data: {"candidates":[{"content":{"parts":[{"text":"안"}],...},...}]}

data: [DONE]
```

---

### `POST /v1/chat/completions`

OpenAI-kompatible API. Konvertiert intern in das Gemini-Format zur Verarbeitung.

**Anfragekörper:**
```json
{
  "model": "gemini-2.5-flash",
  "messages": [
    {"role": "system", "content": "당신은 도움이 되는 어시스턴트입니다."},
    {"role": "user", "content": "안녕하세요"}
  ],
  "temperature": 0.7,
  "max_tokens": 1024,
  "stream": false
}
```

**Anbieter-Präfix-Unterstützung im `model`-Feld (OpenClaw 3.11+):**

| Modellbeispiel | Routing |
|----------------|---------|
| `gemini-2.5-flash` | Aktuell konfigurierter Dienst |
| `google/gemini-2.5-pro` | Google direkt |
| `openai/gpt-4o` | OpenAI direkt |
| `anthropic/claude-opus-4-6` | Über OpenRouter |
| `openrouter/meta-llama/llama-3.3-70b` | OpenRouter direkt |
| `wall-vault/gemini-2.5-flash` | Automatische Erkennung → Google |
| `wall-vault/claude-opus-4-6` | Automatische Erkennung → OpenRouter Anthropic |
| `wall-vault/gpt-4o` | Automatische Erkennung → OpenAI |
| `wall-vault/hunter-alpha` | OpenRouter (kostenlos, 1M Kontext) |
| `moonshot/kimi-k2.5` | Über OpenRouter |
| `opencode-go/model` | Über OpenRouter |
| `kimi-k2.5:cloud` | `:cloud`-Suffix → OpenRouter |

Weitere Details siehe [Anbieter-/Modell-Routing](#anbietermodell-routing).

**Antwortkörper:**
```json
{
  "choices": [
    {
      "message": {
        "role": "assistant",
        "content": "안녕하세요! 무엇을 도와드릴까요?"
      },
      "finish_reason": "stop",
      "index": 0
    }
  ]
}
```

> **Automatische Entfernung von Modell-Steuerungstoken:** Wenn die Antwort GLM-5 / DeepSeek / ChatML-Trennzeichen enthält (`<|im_start|>`, `[gMASK]`, `[sop]` usw.), werden diese automatisch entfernt.

---

## Schlüsseltresor-API (:56243)

Der Server, auf dem der Schlüsseltresor läuft. Standardport `56243`.

---

### Öffentliche API (Keine Authentifizierung erforderlich)

#### `GET /`

Web-Dashboard-UI. Zugriff über den Browser.

---

#### `GET /api/status`

Tresorstatus.

**Antwortbeispiel:**
```json
{
  "status": "ok",
  "version": "v0.1.6.20260314.231308",
  "keys": 3,
  "clients": 2,
  "sse": 2
}
```

---

#### `GET /api/clients`

Liste registrierter Clients (nur öffentliche Informationen, Token ausgeschlossen).

---

### `GET /api/events`

SSE (Server-Sent Events) Echtzeit-Ereignisstrom.

**Header:**
```
Content-Type: text/event-stream
Cache-Control: no-cache
Connection: keep-alive
```

**Sofort bei Verbindung empfangen:**
```
data: {"type":"connected","clients":2}
```

**Ereignisbeispiele:**
```
data: {"type":"config_change","data":{"client_id":"bot-a","service":"google","model":"gemini-2.5-pro"}}

data: {"type":"key_added","data":{"service":"openrouter"}}

data: {"type":"key_deleted","data":{"service":"google"}}

data: {"type":"service_changed","data":{"action":"updated","id":"ollama"}}

data: {"type":"usage_reset","data":{"time":"2026-03-13T00:00:30Z"}}
```

Detaillierte Ereignistypen siehe [SSE-Ereignistypen](#sse-ereignistypen).

---

### Nur-Proxy-API (Client-Token)

Erfordert den Header `Authorization: Bearer <client_token>`. Admin-Token werden ebenfalls akzeptiert.

#### `GET /api/keys`

Entschlüsselte API-Schlüsselliste für den Proxy.

**Abfrageparameter:**

| Parameter | Beschreibung |
|-----------|-------------|
| `service` | Dienstfilter (z.B. `?service=google`) |

**Antwortbeispiel:**
```json
[
  {
    "id": "key-abc123",
    "service": "google",
    "plain_key": "AIzaSy...",
    "daily_limit": 1000,
    "today_usage": 42,
    "today_attempts": 45
  }
]
```

> **Sicherheit:** Gibt Klartextschlüssel zurück. Es werden nur Schlüssel für Dienste zurückgegeben, die durch die `allowed_services`-Einstellung des Clients erlaubt sind.

---

#### `GET /api/services`

Dienstliste für den Proxy. Gibt ein Array von Dienst-IDs zurück, bei denen `proxy_enabled=true` ist.

**Antwortbeispiel:**
```json
["google", "ollama"]
```

Bei leerem Array nutzt der Proxy alle Dienste ohne Einschränkung.

---

#### `POST /api/heartbeat`

Proxy-Status senden (wird automatisch alle 20 Sekunden ausgeführt).

**Anfragekörper:**
```json
{
  "client_id": "bot-a",
  "version": "v0.1.6.20260314.231308",
  "service": "google",
  "model": "gemini-2.5-flash",
  "sse_connected": true,
  "host": "bot-a-host",
  "avatar": "data:image/png;base64,...",
  "key_usage":     {"key-abc123": 42, "key-def456": 0},
  "key_attempts":  {"key-abc123": 45, "key-def456": 3},
  "key_cooldowns": {"key-abc123": "2026-03-15T14:30:00Z"}
}
```

| Feld | Typ | Beschreibung |
|------|-----|-------------|
| `client_id` | string | Client-ID |
| `version` | string | Proxy-Version (enthält Build-Zeitstempel, z.B. `v0.1.6.20260314.231308`) |
| `service` | string | Aktueller Dienst |
| `model` | string | Aktuelles Modell |
| `sse_connected` | bool | Ob SSE verbunden ist |
| `host` | string | Hostname |
| `avatar` | string | base64 data URI of the proxy's local avatar image (sent by proxy, auto-persisted to client record by vault). Set via `WV_AVATAR` env var (relative path under `~/.openclaw/`). |
| `key_usage` | map[string]int | Key ID → successful tokens used today. Vault calls `SetKeyUsage()` (absolute value sync). 429/402/582 errors do **not** increment this. |
| `key_attempts` | map[string]int | Key ID → total API calls today (success + rate-limited). Vault calls `SetKeyAttempts()`. 429/402/582 errors only increment this, not `key_usage`. |
| `key_cooldowns` | map[string]string | Key ID → cooldown end time (RFC3339). Vault calls `SetKeyCooldownIfLater()`. |

**Antwort:**
```json
{"status": "ok"}
```

---

### Admin-API — API-Schlüssel

Erfordert den Header `Authorization: Bearer <admin_token>`.

#### `GET /admin/keys`

Alle registrierten API-Schlüssel auflisten (Klartextschlüssel ausgeschlossen).

**Antwortbeispiel:**
```json
[
  {
    "id": "key-abc123",
    "service": "google",
    "label": "메인 키",
    "today_usage": 42,
    "today_attempts": 45,
    "daily_limit": 1000,
    "cooldown_until": "0001-01-01T00:00:00Z",
    "last_error": 0,
    "created_at": "2026-03-13T12:00:00Z",
    "available": true,
    "usage_pct": 4
  }
]
```

| Feld | Typ | Beschreibung |
|------|-----|-------------|
| `today_usage` | int | Erfolgreiche Anfrage-Token heute (enthält keine 429/402/582-Fehler) |
| `today_attempts` | int | Gesamte API-Aufrufe heute (Erfolg + Rate-Limited) |
| `available` | bool | Ob ohne Abklingzeit oder Limit verfügbar |
| `usage_pct` | int | Nutzungsprozentsatz des Tageslimits (`daily_limit=0` → 0) |
| `cooldown_until` | RFC3339 | Ende der Abklingzeit (Nullwert bedeutet keine) |
| `last_error` | int | Letzter HTTP-Fehlercode |

---

#### `POST /admin/keys`

Neuen API-Schlüssel registrieren. Ein SSE-`key_added`-Ereignis wird sofort bei der Registrierung gesendet.

**Anfragekörper:**
```json
{
  "service": "google",
  "key": "AIzaSy...",
  "label": "메인 키",
  "daily_limit": 1000
}
```

| Feld | Erforderlich | Beschreibung |
|------|-------------|-------------|
| `service` | ✅ | `google` \| `openai` \| `anthropic` \| `openrouter` \| `ollama` \| benutzerdefiniert |
| `key` | ✅ | API-Schlüssel im Klartext |
| `label` | — | Identifikationslabel |
| `daily_limit` | — | Tägliches Nutzungslimit (0 = unbegrenzt) |

---

#### `DELETE /admin/keys/{id}`

API-Schlüssel löschen. Nach dem Löschen wird ein SSE-`key_deleted`-Ereignis gesendet.

**Antwort:**
```json
{"status": "deleted"}
```

---

#### `POST /admin/keys/reset`

Tägliche Nutzung aller Schlüssel zurücksetzen. SSE-`usage_reset`-Ereignis wird gesendet.

**Antwort:**
```json
{
  "status": "reset",
  "time": "2026-03-13T15:00:00Z"
}
```

---

### Admin-API — Clients

#### `GET /admin/clients`

Alle Clients auflisten (einschließlich Token).

---

#### `POST /admin/clients`

Neuen Client registrieren.

**Anfragekörper:**
```json
{
  "id": "my-bot",
  "name": "내 봇",
  "token": "my-secret-token",
  "default_service": "google",
  "default_model": "gemini-2.5-flash",
  "allowed_services": ["google", "openrouter"],
  "agent_type": "openclaw",
  "work_dir": "~/.openclaw",
  "description": "OpenClaw 에이전트",
  "ip_whitelist": ["10.0.0.1", "10.0.0.0/24"],
  "enabled": true
}
```

| Feld | Erforderlich | Beschreibung |
|------|-------------|-------------|
| `id` | ✅ | Eindeutige Client-ID |
| `name` | — | Anzeigename |
| `token` | — | Authentifizierungstoken (wird automatisch generiert, wenn weggelassen) |
| `default_service` | — | Standarddienst |
| `default_model` | — | Standardmodell |
| `allowed_services` | — | Erlaubte Dienstliste (leeres Array = alle erlaubt) |
| `agent_type` | — | `openclaw` \| `claude-code` \| `cursor` \| `vscode` \| `custom` |
| `work_dir` | — | Arbeitsverzeichnis des Agenten |
| `description` | — | Agentenbeschreibung |
| `ip_whitelist` | — | Erlaubte IP-Liste (leeres Array = alle erlaubt, CIDR unterstützt) |
| `enabled` | — | Ob aktiviert (Standard `true`) |

---

#### `GET /admin/clients/{id}`

Bestimmten Client abrufen (einschließlich Token).

---

#### `PUT /admin/clients/{id}`

Client-Einstellungen aktualisieren. **SSE-`config_change`-Broadcast → wird innerhalb von 1–3 Sekunden auf dem Proxy übernommen.**

**Anfragekörper (nur zu ändernde Felder angeben):**
```json
{
  "default_service": "openrouter",
  "default_model": "anthropic/claude-opus-4-6",
  "enabled": true
}
```

**Antwort:**
```json
{"status": "updated"}
```

---

#### `DELETE /admin/clients/{id}`

Client löschen.

---

### Admin-API — Dienste

#### `GET /admin/services`

Registrierte Dienste auflisten.

**Antwortbeispiel:**
```json
[
  {"id": "google",      "name": "Google Gemini",   "enabled": true,  "custom": false},
  {"id": "openai",      "name": "OpenAI",          "enabled": true,  "custom": false},
  {"id": "anthropic",   "name": "Anthropic",       "enabled": false, "custom": false},
  {"id": "openrouter",  "name": "OpenRouter",      "enabled": true,  "custom": false},
  {"id": "ollama",      "name": "Ollama (Local)",  "enabled": true,  "custom": false,
   "local_url": "http://localhost:11434"},
  {"id": "lmstudio",    "name": "LM Studio",       "enabled": false, "custom": false},
  {"id": "vllm",        "name": "vLLM",            "enabled": false, "custom": false},
  {"id": "github-copilot","name":"GitHub Copilot", "enabled": false, "custom": false}
]
```

8 integrierte Dienste: `google`, `openai`, `anthropic`, `openrouter`, `github-copilot`, `ollama`, `lmstudio`, `vllm`

---

#### `POST /admin/services`

Benutzerdefinierten Dienst hinzufügen. SSE-`service_changed`-Ereignis wird nach dem Hinzufügen gesendet → **Dashboard-Dropdowns werden sofort aktualisiert**.

**Anfragekörper:**
```json
{
  "id": "my-llm",
  "name": "사내 LLM 서버",
  "local_url": "http://10.0.0.50:8080",
  "enabled": true
}
```

---

#### `PUT /admin/services/{id}`

Diensteinstellungen aktualisieren. SSE-`service_changed`-Ereignis wird nach Änderungen gesendet.

**Anfragekörper:**
```json
{
  "local_url": "http://192.168.x.x:11434",
  "enabled": true
}
```

---

#### `DELETE /admin/services/{id}`

Benutzerdefinierten Dienst löschen. SSE-`service_changed`-Ereignis wird nach dem Löschen gesendet.

Versuch, einen integrierten Dienst zu löschen (`custom: false`):
```json
{"error": "기본 서비스는 삭제할 수 없습니다: google"}
```

---

### Admin-API — Modellliste

#### `GET /admin/models`

Modelle nach Dienst auflisten. Verwendet TTL-Cache (10 Minuten).

**Abfrageparameter:**

| Parameter | Beschreibung | Beispiel |
|-----------|-------------|----------|
| `service` | Dienstfilter | `?service=google` |
| `q` | Modellsuche | `?q=gemini` |

**Modellabfrage nach Dienst:**

| Dienst | Methode | Anzahl |
|--------|---------|--------|
| `google` | Statische Liste | 8 (inkl. Embedding) |
| `openai` | Statische Liste | 9 |
| `anthropic` | Statische Liste | 6 |
| `github-copilot` | Statische Liste | 6 |
| `openrouter` | Dynamische API-Abfrage (Fallback auf 14 kuratierte Modelle bei Fehler) | 340+ |
| `ollama` | Dynamische lokale Serverabfrage (7 empfohlene bei Nicht-Antwort) | Variabel |
| `lmstudio` | Dynamische lokale Serverabfrage | Variabel |
| `vllm` | Dynamische lokale Serverabfrage | Variabel |
| Benutzerdefiniert | OpenAI-kompatibel `/v1/models` | Variabel |

**OpenRouter-Fallback-Modelle (bei nicht erreichbarer API):**

| Modell | Hinweise |
|--------|----------|
| `openrouter/hunter-alpha` | Kostenlos, 1M Kontext |
| `openrouter/healer-alpha` | Kostenlos, omni-modal |
| `moonshot/kimi-k2.5` | 256K Kontext |
| `z-ai/glm-5`, `z-ai/glm-4.7-flash` | — |
| `deepseek/deepseek-r1`, `deepseek/deepseek-chat` | — |
| `qwen/qwen-2.5-72b-instruct` | 131K Kontext |
| `minimax/minimax-m2.5` | — |
| `meta-llama/llama-3.3-70b-instruct` | 131K Kontext |

---

### Admin-API — Proxy-Status

#### `GET /admin/proxies`

Letzter Heartbeat-Status aller verbundenen Proxys.

---

## SSE-Ereignistypen

Ereignisse, die vom Tresor-`/api/events`-Strom empfangen werden:

| `type` | Auslöser | `data`-Inhalt | Dashboard-Reaktion |
|--------|----------|---------------|-------------------|
| `connected` | Sofort bei SSE-Verbindung | `{"clients": N}` | — |
| `config_change` | Client-Einstellungen geändert | `{"client_id","service","model"}` | Agentenkarte Modell-Dropdown aktualisiert |
| `key_added` | Neuer API-Schlüssel registriert | `{"service": "google"}` | Modell-Dropdown aktualisiert |
| `key_deleted` | API-Schlüssel gelöscht | `{"service": "google"}` | Modell-Dropdown aktualisiert |
| `service_changed` | Dienst hinzugefügt/aktualisiert/gelöscht | `{"action":"added"\|"updated"\|"deleted","id":"...","proxy_services":["google","ollama"]}` | Dienst-Select + Modell-Dropdown sofort aktualisiert; Dispatch-Dienstliste des Proxys in Echtzeit aktualisiert |
| `usage_update` | Bei Proxy-Heartbeat (alle 20s) | `{"keys":[{"id","service","today_usage","today_attempts","daily_limit","cooldown_until"},...]}` | Schlüsselnutzungsbalken und -zahlen sofort aktualisiert, Abklingzeit-Countdown startet. SSE-Daten werden direkt ohne Fetch verwendet. Balken verwenden Anteil-am-Gesamt-Skalierung (für unbegrenzte Schlüssel). |
| `usage_reset` | Tägliche Nutzung zurückgesetzt | `{"time": "RFC3339"}` | Seitenaktualisierung |

**Ereignisverarbeitung auf der Proxy-Seite:**

```
config_change empfangen
  → Wenn client_id mit eigener ID übereinstimmt
    → service, model sofort aktualisiert
    → hooksMgr.Fire(EventModelChanged)
```

---

## Anbieter-/Modell-Routing

Bei Angabe eines `provider/model`-Formats im `model`-Feld von `/v1/chat/completions` wird automatisches Routing angewendet (OpenClaw 3.11-kompatibel).

### Präfix-Routing-Regeln

| Präfix | Routing-Ziel | Beispiel |
|--------|-------------|----------|
| `google/` | Google direkt | `google/gemini-2.5-pro` |
| `openai/` | OpenAI direkt | `openai/gpt-4o` |
| `anthropic/` | Über OpenRouter | `anthropic/claude-opus-4-6` |
| `ollama/` | Ollama direkt | `ollama/qwen3.5:35b` |
| `custom/` | Rekursives Re-Parsing (`custom/` entfernen und erneut routen) | `custom/google/gemini-2.5-flash` → Google |
| `openrouter/` | OpenRouter (Bare-Pfad beibehalten) | `openrouter/meta-llama/llama-3.3-70b` |
| `opencode/`, `opencode-go/`, `opencode-zen/` | OpenRouter | `opencode-go/model` |
| `moonshot/`, `kimi-coding/` | OpenRouter (vollständiger Pfad beibehalten) | `moonshot/kimi-k2.5` |
| `groq/`, `mistral/`, `minimax/`, `cohere/`, `perplexity/` | OpenRouter | `groq/llama-3.1-70b` |
| `together/`, `huggingface/`, `nvidia/`, `venice/` | OpenRouter | — |
| `meta-llama/`, `qwen/`, `deepseek/`, `01-ai/` | OpenRouter (vollständiger Pfad) | `deepseek/deepseek-r1` |

### `wall-vault/`-Präfix automatische Erkennung

Das wall-vault-eigene Präfix ermittelt den Dienst automatisch aus der Modell-ID.

| Modell-ID-Muster | Routing |
|------------------|---------|
| `wall-vault/gemini-*` | Google |
| `wall-vault/gpt-*`, `wall-vault/o1`, `wall-vault/o3`, `wall-vault/o4*` | OpenAI |
| `wall-vault/claude-*` | OpenRouter (Anthropic-Pfad) |
| `wall-vault/hunter-alpha`, `wall-vault/healer-alpha` | OpenRouter (kostenlos, 1M ctx) |
| `wall-vault/kimi-*`, `wall-vault/glm-*`, `wall-vault/deepseek-*`, `wall-vault/qwen*` | OpenRouter |
| Sonstige | OpenRouter |

### `:cloud`-Suffix-Behandlung

Das Ollama-Tag-Format `:cloud`-Suffix wird automatisch entfernt und an OpenRouter geroutet.

```
kimi-k2.5:cloud  →  OpenRouter, Modell-ID: kimi-k2.5
glm-5:cloud      →  OpenRouter, Modell-ID: glm-5
```

### OpenClaw openclaw.json Integrationsbeispiel

```json5
// ~/.openclaw/openclaw.json
{
  models: {
    providers: {
      "wall-vault": {
        baseUrl: "https://localhost:56244/v1",
        apiKey: "YOUR_AGENT_TOKEN",
        api: "openai-completions",
        models: [
          { id: "wall-vault/gemini-2.5-flash" },
          { id: "wall-vault/hunter-alpha" },
          { id: "wall-vault/claude-opus-4-6" }
        ]
      }
    }
  },
  agents: {
    defaults: {
      model: {
        primary: "wall-vault/gemini-2.5-flash",
        fallbacks: ["wall-vault/hunter-alpha"]
      }
    }
  }
}
```

Klicken Sie auf die **🐾-Schaltfläche** auf einer Agentenkarte, um das Konfigurationssnippet für diesen Agenten automatisch in die Zwischenablage zu kopieren.

---

## Datenschema

### APIKey

| Feld | Typ | Beschreibung |
|------|-----|-------------|
| `id` | string | Eindeutige ID im UUID-Format |
| `service` | string | `google` \| `openai` \| `anthropic` \| `openrouter` \| `ollama` \| benutzerdefiniert |
| `encrypted_key` | string | AES-GCM-verschlüsselter Schlüssel (Base64) |
| `label` | string | Identifikationslabel |
| `today_usage` | int | Erfolgreiche Anfrage-Token heute (enthält keine 429/402/582-Fehler) |
| `today_attempts` | int | Gesamte API-Aufrufe heute (Erfolg + Rate-Limited; Zurücksetzung um Mitternacht) |
| `daily_limit` | int | Tageslimit (0 = unbegrenzt) |
| `cooldown_until` | time.Time | Ende der Abklingzeit |
| `last_error` | int | Letzter HTTP-Fehlercode |
| `created_at` | time.Time | Registrierungszeitpunkt |

**Abklingzeit-Richtlinie:**

| HTTP-Fehler | Abklingzeit |
|-------------|-------------|
| 429 (Too Many Requests) | 30 Minuten |
| 402 (Payment Required) | 24 Stunden |
| 400 / 401 / 403 | 24 Stunden |
| 582 (Gateway Overload) | 5 Minuten |
| Netzwerkfehler | 10 Minuten |

> **429/402/582**: Abklingzeit wird gesetzt + `today_attempts` wird erhöht. `today_usage` bleibt unverändert (nur erfolgreiche Token werden gezählt).
> **Ollama (lokaler Dienst)**: `callOllama` verwendet einen dedizierten HTTP-Client mit `Timeout: 0` (unbegrenzt). Inferenz großer Modelle kann zehn Sekunden bis Minuten dauern, daher wird das Standard-60-Sekunden-Timeout nicht angewendet.

### Client

| Feld | Typ | Beschreibung |
|------|-----|-------------|
| `id` | string | Eindeutige Client-ID |
| `name` | string | Anzeigename |
| `token` | string | Authentifizierungstoken |
| `default_service` | string | Standarddienst |
| `default_model` | string | Standardmodell (kann im `provider/model`-Format sein) |
| `allowed_services` | []string | Erlaubte Dienste (leeres Array = alle) |
| `agent_type` | string | `openclaw` \| `claude-code` \| `cursor` \| `vscode` \| `custom` |
| `work_dir` | string | Arbeitsverzeichnis des Agenten |
| `description` | string | Beschreibung |
| `ip_whitelist` | []string | Erlaubte IP-Liste (CIDR unterstützt) |
| `avatar` | string | Agent avatar — relative path under `~/.openclaw/` (e.g. `workspace/avatar.png`, `workspace/avatars/profile.hpg`) OR base64 data URI (`data:image/png;base64,...`). Supported extensions: `.png`, `.jpg`/`.jpeg`/`.hpg`, `.webp`, `.gif`. |
| `enabled` | bool | Bei `false` wird `403` beim Zugriff auf `/api/keys` zurückgegeben |
| `created_at` | time.Time | Registrierungszeitpunkt |

### ServiceConfig

| Feld | Typ | Beschreibung |
|------|-----|-------------|
| `id` | string | Eindeutige Dienst-ID |
| `name` | string | Anzeigename |
| `local_url` | string | Lokale Server-URL (Ollama/LMStudio/vLLM/benutzerdefiniert) |
| `enabled` | bool | Ob aktiviert |
| `custom` | bool | Ob benutzerdefiniert hinzugefügter Dienst |
| `proxy_enabled` | bool | Whether this service is included in proxy dispatch and agent model dropdowns. Controlled by the "프록시 사용" checkbox in the Services card. Only `proxy_enabled: true` services appear in agent service/model selectors. |

### ProxyStatus (Heartbeat)

| Feld | Typ | Beschreibung |
|------|-----|-------------|
| `client_id` | string | Client-ID |
| `version` | string | Proxy-Version (z.B. `v0.1.6.20260314.231308`) |
| `service` | string | Aktueller Dienst |
| `model` | string | Aktuelles Modell |
| `sse_connected` | bool | Ob SSE verbunden ist |
| `host` | string | Hostname |
| `avatar` | string | base64 data URI of local avatar (auto-synced from proxy via heartbeat) |
| `updated_at` | time.Time | Letzte Aktualisierung |
| `vault.today_usage` | int | Heutige Token-Nutzung |
| `vault.daily_limit` | int | Tageslimit |
| `vault.key_status` | string | `active` \| `cooldown` \| `exhausted` |

---

## Fehlerantworten

```json
{"error": "오류 메시지"}
```

| Code | Bedeutung |
|------|-----------|
| 200 | Erfolg |
| 400 | Ungültige Anfrage |
| 401 | Authentifizierung fehlgeschlagen |
| 403 | Zugriff verweigert (deaktivierter Client, IP gesperrt) |
| 404 | Ressource nicht gefunden |
| 405 | Methode nicht erlaubt |
| 429 | Rate-Limit überschritten |
| 500 | Interner Serverfehler |
| 502 | Upstream-API-Fehler (alle Fallbacks fehlgeschlagen) |

---

## cURL-Beispiele

```bash
# ─── Proxy ────────────────────────────────────────────────────────────────────

# Gesundheitsprüfung
curl https://localhost:56244/health

# Status
curl https://localhost:56244/status

# Modellliste (alle)
curl https://localhost:56244/api/models

# Nur Google-Modelle
curl "https://localhost:56244/api/models?service=google"

# Kostenlose Modelle suchen
curl "https://localhost:56244/api/models?q=alpha"

# Modell ändern (lokal)
curl -X PUT https://localhost:56244/api/config/model \
  -H "Content-Type: application/json" \
  -d '{"service":"google","model":"gemini-2.5-pro"}'

# Konfiguration neu laden
curl -X POST https://localhost:56244/reload

# Direkter Gemini-API-Aufruf
curl -X POST "https://localhost:56244/google/v1beta/models/gemini-2.5-flash:generateContent" \
  -H "Content-Type: application/json" \
  -d '{"contents":[{"role":"user","parts":[{"text":"안녕하세요"}]}]}'

# OpenAI-kompatibel (Standardmodell)
curl -X POST https://localhost:56244/v1/chat/completions \
  -H "Content-Type: application/json" \
  -d '{"model":"gemini-2.5-flash","messages":[{"role":"user","content":"안녕"}]}'

# OpenClaw provider/model Format
curl -X POST https://localhost:56244/v1/chat/completions \
  -H "Authorization: Bearer my-agent-token" \
  -H "Content-Type: application/json" \
  -d '{"model":"wall-vault/gemini-2.5-flash","messages":[{"role":"user","content":"안녕"}]}'

# Kostenloses 1M-Kontext-Modell
curl -X POST https://localhost:56244/v1/chat/completions \
  -H "Authorization: Bearer my-agent-token" \
  -H "Content-Type: application/json" \
  -d '{"model":"wall-vault/hunter-alpha","messages":[{"role":"user","content":"안녕"}]}'

# ─── Schlüsseltresor (Öffentlich) ─────────────────────────────────────────────

curl https://localhost:56243/api/status
curl https://localhost:56243/api/clients
curl -s https://localhost:56243/api/events --max-time 3

# ─── Schlüsseltresor (Admin) ──────────────────────────────────────────────────

ADMIN="Authorization: Bearer admin-token"

# Schlüsselliste
curl -H "$ADMIN" https://localhost:56243/admin/keys

# Google-Schlüssel hinzufügen
curl -X POST https://localhost:56243/admin/keys \
  -H "$ADMIN" -H "Content-Type: application/json" \
  -d '{"service":"google","key":"AIzaSy...","label":"메인 키","daily_limit":1000}'

# OpenAI-Schlüssel hinzufügen
curl -X POST https://localhost:56243/admin/keys \
  -H "$ADMIN" -H "Content-Type: application/json" \
  -d '{"service":"openai","key":"sk-...","label":"GPT 키"}'

# OpenRouter-Schlüssel hinzufügen
curl -X POST https://localhost:56243/admin/keys \
  -H "$ADMIN" -H "Content-Type: application/json" \
  -d '{"service":"openrouter","key":"sk-or-v1-...","label":"OR 키"}'

# Schlüssel löschen (SSE key_deleted Broadcast)
curl -X DELETE https://localhost:56243/admin/keys/key-abc123 -H "$ADMIN"

# Tägliche Nutzung zurücksetzen
curl -X POST https://localhost:56243/admin/keys/reset -H "$ADMIN"

# Client-Liste
curl -H "$ADMIN" https://localhost:56243/admin/clients

# Client hinzufügen (OpenClaw)
curl -X POST https://localhost:56243/admin/clients \
  -H "$ADMIN" -H "Content-Type: application/json" \
  -d '{"id":"bot-a","name":"봇 A","agent_type":"openclaw","work_dir":"~/.openclaw","default_service":"google","default_model":"gemini-2.5-flash"}'

# Client-Modell ändern (SSE sofortige Aktualisierung)
curl -X PUT https://localhost:56243/admin/clients/bot-a \
  -H "$ADMIN" -H "Content-Type: application/json" \
  -d '{"default_service":"openrouter","default_model":"wall-vault/hunter-alpha"}'

# Client deaktivieren
curl -X PUT https://localhost:56243/admin/clients/my-bot \
  -H "$ADMIN" -H "Content-Type: application/json" \
  -d '{"enabled":false}'

# Client löschen
curl -X DELETE https://localhost:56243/admin/clients/my-bot -H "$ADMIN"

# Dienstliste
curl -H "$ADMIN" https://localhost:56243/admin/services

# Ollama lokale URL setzen (SSE service_changed Broadcast)
curl -X PUT https://localhost:56243/admin/services/ollama \
  -H "$ADMIN" -H "Content-Type: application/json" \
  -d '{"local_url":"http://192.168.x.x:11434","enabled":true}'

# OpenAI-Dienst aktivieren
curl -X PUT https://localhost:56243/admin/services/openai \
  -H "$ADMIN" -H "Content-Type: application/json" \
  -d '{"enabled":true}'

# Benutzerdefinierten Dienst hinzufügen (SSE service_changed Broadcast)
curl -X POST https://localhost:56243/admin/services \
  -H "$ADMIN" -H "Content-Type: application/json" \
  -d '{"id":"my-llm","name":"사내 LLM","local_url":"http://10.0.0.50:8080","enabled":true}'

# Benutzerdefinierten Dienst löschen
curl -X DELETE https://localhost:56243/admin/services/my-llm -H "$ADMIN"

# Modellliste
curl -H "$ADMIN" https://localhost:56243/admin/models
curl -H "$ADMIN" "https://localhost:56243/admin/models?service=openrouter"
curl -H "$ADMIN" "https://localhost:56243/admin/models?q=hunter"

# Proxy-Status (Heartbeat)
curl -H "$ADMIN" https://localhost:56243/admin/proxies

# ─── Verteilter Modus — Proxy → Tresor ───────────────────────────────────────

# Entschlüsselte Schlüssel abrufen
curl https://localhost:56243/api/keys \
  -H "Authorization: Bearer your-bot-a-token"

# Heartbeat senden
curl -X POST https://localhost:56243/api/heartbeat \
  -H "Authorization: Bearer your-bot-a-token" \
  -H "Content-Type: application/json" \
  -d '{"client_id":"bot-a","version":"v0.1.6.20260314.231308","service":"google","model":"gemini-2.5-flash","sse_connected":true}'
```

---

## Middleware

Automatisch auf alle Anfragen angewendet:

| Middleware | Funktion |
|-----------|----------|
| **Logger** | Protokolliert im Format `[method] path status latencyms` |
| **CORS** | `Access-Control-Allow-Origin: *` |
| **Recovery** | Panic-Wiederherstellung, gibt 500-Antwort zurück |

---

*Zuletzt aktualisiert: 16.03.2026 — v0.1.7: dashboard title renamed to "벽금고(wall-vault) 대시보드", logo moved to sticky topbar, today_attempts tracking, HTTP 582 cooldown (5min), share-of-total bar scaling, custom/ routing fix, Ollama Timeout:0, key_att i18n, avatar heartbeat sync, build timestamp versioning, proxy-only service filter*
