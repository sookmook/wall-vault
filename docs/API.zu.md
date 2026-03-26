# Umhlahlandlela we-API ye-wall-vault

Le dokhumenti ichaza ngokuningiliziwe wonke ama-endpoint e-HTTP API e-wall-vault.

---

## Okuqukethwe

- [Ukuqinisekisa](#ukuqinisekisa)
- [I-API ye-Proxy (:56244)](#i-api-ye-proxy-56244)
  - [Ukuhlola impilo](#get-health)
  - [Isimo esiningiliziwe](#get-status)
  - [Uhlu lwamamodeli](#get-apimodels)
  - [Ukushintsha imodeli](#put-apiconfigmodel)
  - [Imodi yokucabanga](#put-apiconfigthink-mode)
  - [Ukuhlaziya izilungiselelo](#post-reload)
  - [Gemini API](#post-googlev1betamodelsmgeneratecontent)
  - [Gemini usakazo](#post-googlev1betamodelsmstreamgeneratecontent)
  - [I-API ehambisanayo ne-OpenAI](#post-v1chatcompletions)
- [I-API Yokugcina Okhiye (:56243)](#i-api-yokugcina-okhiye-56243)
  - [I-API yomphakathi](#i-api-yomphakathi-akudingeki-ukuqinisekisa)
  - [Umfudlana wezigameko ze-SSE](#get-apievents)
  - [I-API ye-proxy kuphela](#i-api-ye-proxy-kuphela-ithokheni-yeklayenti)
  - [I-API yomphathi — Okhiye](#i-api-yomphathi--okhiye-be-api)
  - [I-API yomphathi — Amaklayenti](#i-api-yomphathi--amaklayenti)
  - [I-API yomphathi — Amasevisi](#i-api-yomphathi--amasevisi)
  - [I-API yomphathi — Uhlu lwamamodeli](#i-api-yomphathi--uhlu-lwamamodeli)
  - [I-API yomphathi — Isimo se-proxy](#i-api-yomphathi--isimo-se-proxy)
- [Izinhlobo zezigameko ze-SSE](#izinhlobo-zezigameko-ze-sse)
- [Ukulayisha komhlinzeki·wemodeli](#ukulayisha-komhlinzekiwemodeli)
- [Isikimu sedatha](#isikimu-sedatha)
- [Izimpendulo zamaphutha](#izimpendulo-zamaphutha)
- [Iqoqo lezibonelo ze-cURL](#iqoqo-lezibonelo-ze-curl)

---

## Ukuqinisekisa

| Indawo | Indlela | Isihloko |
|------|------|------|
| I-API yomphathi | Ithokheni ye-Bearer | `Authorization: Bearer <admin_token>` |
| I-Proxy → Isigcini | Ithokheni ye-Bearer | `Authorization: Bearer <client_token>` |
| I-API ye-Proxy | Ayikho (yendawo) | — |

Uma `admin_token` ingabekwanga (uchungechunge olungenalutho) yonke i-API yomphathi ingatholakala ngaphandle kokuqinisekisa.

### Inqubomgomo yokuphepha

- **Ukukhawulela Isilinganiso**: Uma ukuqinisekisa kwe-API yomphathi kuhluleka izikhathi eziyi-10/imizuzu eyi-15, lelo IP lizovalelwa okwesikhashana (`429 Too Many Requests`)
- **Uhlu Lwama-IP Avunyelwe**: Ama-IP/CIDR abhaliswe esigabeni se-`ip_whitelist` se-ejenti (`Client`) kuphela angafinyelela `/api/keys`. Uma uhlu lungenalutho, bonke bavunyelwe.
- **Ukuvikelwa kwe-theme·lang**: `/admin/theme`, `/admin/lang` nazo zidinga ukuqinisekiswa kwethokheni yomphathi

---

## I-API ye-Proxy (:56244)

Iseva lapho i-proxy isebenza khona. Inombolo yetheku elivamile `56244`.

---

### `GET /health`

Ukuhlola impilo. Ihlale ibuyisa 200 OK.

**Isibonelo sempendulo:**
```json
{
  "status": "ok",
  "version": "v0.1.6.20260314.231308",
  "client": "bot-a"
}
```

---

### `GET /status`

Ukuhlola okununingiliziwe kwesimo se-proxy.

**Isibonelo sempendulo:**
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

| Isigaba | Uhlobo | Incazelo |
|------|------|------|
| `service` | string | Isevisi yamanje evamile |
| `model` | string | Imodeli yamanje evamile |
| `sse` | bool | Isimo sokuxhumana kwe-SSE sesigcini |
| `filter` | string | Imodi yesihlungi samathuluzi |
| `services` | []string | Uhlu lwamasevisi asebenzayo |
| `mode` | string | `standalone` \| `distributed` |

---

### `GET /api/models`

Uhlu lwamamodeli atholakalayo. Isebenzisa i-TTL cache (imizuzu eyi-10 evamile).

**Amapharamitha ombuzo:**

| Ipharamitha | Incazelo | Isibonelo |
|---------|------|------|
| `service` | Isihlungi sesevisi | `?service=google` |
| `q` | Ukusesha i-ID/igama lemodeli | `?q=gemini` |

**Isibonelo sempendulo:**
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

| Isigaba | Uhlobo | Incazelo |
|------|------|------|
| `id` | string | I-ID yemodeli |
| `name` | string | Igama lokubonisa lemodeli |
| `service` | string | `google` \| `openai` \| `anthropic` \| `openrouter` \| `ollama` njll. |
| `context_length` | int | Ubukhulu bewindi lomongo |
| `free` | bool | Noma imodeli yamahhala yini (OpenRouter) |

---

### `PUT /api/config/model`

Ukushintsha isevisi·nemodeli yamanje.

**Umzimba wesicelo:**
```json
{
  "service": "google",
  "model": "gemini-2.5-flash"
}
```

**Impendulo:**
```json
{
  "status": "ok",
  "service": "google",
  "model": "gemini-2.5-flash"
}
```

> **Qaphela:** Kwimodi ye-distributed, kunconywa ukusebenzisa `PUT /admin/clients/{id}` yesigcini esikhundleni sale API. Izinguquko ezenziwe esigcinini zibonakala ngokuzenzakalelayo nge-SSE phakathi kwamasekhondi ayi-1–3.

---

### `PUT /api/config/think-mode`

Ukushintsha imodi yokucabanga (no-op, yokwengezelwa kwesikhathi esizayo).

**Impendulo:**
```json
{"status": "ok"}
```

---

### `POST /reload`

Ukuvumelanisa kabusha izilungiselelo zeklayenti·nokhiye ngokushesha esigcinini.

**Impendulo:**
```json
{"status": "reloading"}
```

Ukuvumelanisa kabusha kusebenza ngendlela engahlanganyeli, ngakho kuqedwa phakathi kwamasekhondi ayi-1–2 ngemva kokuthola impendulo.

---

### `POST /google/v1beta/models/{model}:generateContent`

I-proxy ye-Gemini API (engeyona eyosakazo).

**Ipharamitha yendlela:**
- `{model}`: I-ID yemodeli. Uma inesijobelelo sangaphambili se-`gemini-` isevisi ye-Google ikhetha ngokuzenzakalelayo.

**Umzimba wesicelo:** [Isakhiwo sesicelo se-Gemini generateContent](https://ai.google.dev/api/generate-content)

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

**Umzimba wempendulo:** Isakhiwo sempendulo ye-Gemini generateContent

**Isihlungi samathuluzi:** Uma `tool_filter: strip_all` ibekiwe, uhlu lwama-`tools` lwesicelo lususwa ngokuzenzakalelayo.

**Umcwayo wokuhlehla:** Isevisi ekhethiwe iyehluleka → ukuhlehla ngokulandelana kwezevisi ezibekiwe → Ollama (yokugcina).

---

### `POST /google/v1beta/models/{model}:streamGenerateContent`

I-proxy yosakazo ye-Gemini API. Isakhiwo sesicelo sifana nesokungabi nosakazo. Impendulo iza njengomfudlana we-SSE:

```
data: {"candidates":[{"content":{"parts":[{"text":"안"}],...},...}]}

data: [DONE]
```

---

### `POST /v1/chat/completions`

I-API ehambisanayo ne-OpenAI. Ngaphakathi iguqulelwa esakhiweni se-Gemini bese icutshungulwa.

**Umzimba wesicelo:**
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

**Ukwesekwa kwesijobelelo sangaphambili somhlinzeki esigabeni se-`model` (OpenClaw 3.11+):**

| Isibonelo semodeli | Ukulayisha |
|-----------|--------|
| `gemini-2.5-flash` | Isevisi yezilungiselelo zamanje |
| `google/gemini-2.5-pro` | Google ngokuqondile |
| `openai/gpt-4o` | OpenAI ngokuqondile |
| `anthropic/claude-opus-4-6` | Nge-OpenRouter |
| `openrouter/meta-llama/llama-3.3-70b` | OpenRouter ngokuqondile |
| `wall-vault/gemini-2.5-flash` | Ukuthola ngokuzenzakalelayo → Google |
| `wall-vault/claude-opus-4-6` | Ukuthola ngokuzenzakalelayo → OpenRouter Anthropic |
| `wall-vault/gpt-4o` | Ukuthola ngokuzenzakalelayo → OpenAI |
| `wall-vault/hunter-alpha` | OpenRouter (mahhala 1M context) |
| `moonshot/kimi-k2.5` | Nge-OpenRouter |
| `opencode-go/model` | Nge-OpenRouter |
| `kimi-k2.5:cloud` | Isijobelelo sangemuva `:cloud` → OpenRouter |

Ukuthola imininingwane engeziwe bheka [Ukulayisha komhlinzeki·wemodeli](#ukulayisha-komhlinzekiwemodeli).

**Umzimba wempendulo:**
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

> **Ukususwa ngokuzenzakalelayo kwamathokheni okulawula imodeli:** Uma impendulo iqukethe izihlukanisi ze-GLM-5 / DeepSeek / ChatML (`<|im_start|>`, `[gMASK]`, `[sop]` njll.) zisuswa ngokuzenzakalelayo.

---

## I-API Yokugcina Okhiye (:56243)

Iseva lapho isigcini sokhiye sisebenza khona. Inombolo yetheku elivamile `56243`.

---

### I-API yomphakathi (akudingeki ukuqinisekisa)

#### `GET /`

I-UI yedashibhodi yewebhu. Finyelela ngesiphequluli.

---

#### `GET /api/status`

Ukuhlola isimo sesigcini.

**Isibonelo sempendulo:**
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

Uhlu lwamaklayenti abhaliswe (ulwazi lomphakathi kuphela, ngaphandle kwethokheni).

---

### `GET /api/events`

Umfudlana wezigameko ze-SSE (Server-Sent Events) ngesikhathi sangempela.

**Izihloko:**
```
Content-Type: text/event-stream
Cache-Control: no-cache
Connection: keep-alive
```

**Utholakala masinyane ngemva kokuxhuma:**
```
data: {"type":"connected","clients":2}
```

**Izibonelo zezigameko:**
```
data: {"type":"config_change","data":{"client_id":"bot-a","service":"google","model":"gemini-2.5-pro"}}

data: {"type":"key_added","data":{"service":"openrouter"}}

data: {"type":"key_deleted","data":{"service":"google"}}

data: {"type":"service_changed","data":{"action":"updated","id":"ollama"}}

data: {"type":"usage_reset","data":{"time":"2026-03-13T00:00:30Z"}}
```

Ukuthola izinhlobo ezigcwele zezigameko bheka [Izinhlobo zezigameko ze-SSE](#izinhlobo-zezigameko-ze-sse).

---

### I-API ye-proxy kuphela (ithokheni yeklayenti)

Isihloko se-`Authorization: Bearer <client_token>` siyadingeka. Ukuqinisekisa ngethokheni yomphathi nako kuyenzeka.

#### `GET /api/keys`

Uhlu lokhiye be-API abavulwe esifihliwe abahlinzekwa i-proxy.

**Amapharamitha ombuzo:**

| Ipharamitha | Incazelo |
|---------|------|
| `service` | Isihlungi sesevisi (isb.: `?service=google`) |

**Isibonelo sempendulo:**
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

> **Ukuphepha:** Ibuyisa okhiye ngombhalo ovulekile. Okhiye bamasevisi avunyelwe kuphela ababuyiswa ngokwezilungiselelo ze-`allowed_services` zeklayenti.

---

#### `GET /api/services`

Uhlu lwamasevisi asetshenziwa yi-proxy. Ibuyisa uhlu lwama-ID amasevisi lapho `proxy_enabled=true`.

**Isibonelo sempendulo:**
```json
["google", "ollama"]
```

Uma uhlu lungenalutho, i-proxy isebenzisa wonke amasevisi ngaphandle kwemikhawulo.

---

#### `POST /api/heartbeat`

Ukuthumela isimo se-proxy (kuzenzakalelayo njalo ngamasekhondi angu-20).

**Umzimba wesicelo:**
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

| Isigaba | Uhlobo | Incazelo |
|------|------|------|
| `client_id` | string | I-ID yeklayenti |
| `version` | string | Inguqulo ye-proxy (ne-build timestamp, isb. `v0.1.6.20260314.231308`) |
| `service` | string | Isevisi yamanje |
| `model` | string | Imodeli yamanje |
| `sse_connected` | bool | Isimo sokuxhumana kwe-SSE |
| `host` | string | Igama lomsingathi |
| `avatar` | string | base64 data URI of the proxy's local avatar image (sent by proxy, auto-persisted to client record by vault). Set via `WV_AVATAR` env var (relative path under `~/.openclaw/`). |
| `key_usage` | map[string]int | Key ID → successful tokens used today. Vault calls `SetKeyUsage()` (absolute value sync). 429/402/582 errors do **not** increment this. |
| `key_attempts` | map[string]int | Key ID → total API calls today (success + rate-limited). Vault calls `SetKeyAttempts()`. 429/402/582 errors only increment this, not `key_usage`. |
| `key_cooldowns` | map[string]string | Key ID → cooldown end time (RFC3339). Vault calls `SetKeyCooldownIfLater()`. |

**Impendulo:**
```json
{"status": "ok"}
```

---

### I-API yomphathi — Okhiye be-API

Isihloko se-`Authorization: Bearer <admin_token>` siyadingeka.

#### `GET /admin/keys`

Uhlu lwawo wonke okhiye be-API ababhaliswe (ngaphandle kokhiye bombhalo ovulekile).

**Isibonelo sempendulo:**
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

| Isigaba | Uhlobo | Incazelo |
|------|------|------|
| `today_usage` | int | Inani lamathokheni ezicelo eziphumelele namuhla (amaphutha e-429/402/582 awafakiwe) |
| `today_attempts` | int | Isamba samakholi e-API namuhla (aphumelele + akhawulelwe isilinganiso) |
| `available` | bool | Itholakalayo yokusebenziswa ngaphandle kokupholisa·umkhawulo |
| `usage_pct` | int | I-% yokusebenzisa ngokuqhathaniswa nomkhawulo wosuku (`daily_limit=0` ingu-0) |
| `cooldown_until` | RFC3339 | Isikhathi sokuphela kokupholisa (uma inani lize akukho) |
| `last_error` | int | Ikhodi yokugcina yephutha le-HTTP |

---

#### `POST /admin/keys`

Ukubhalisa ukhiye omusha we-API. Masinyane ngemva kokubhalisa isigameko se-SSE `key_added` sisakazwa.

**Umzimba wesicelo:**
```json
{
  "service": "google",
  "key": "AIzaSy...",
  "label": "메인 키",
  "daily_limit": 1000
}
```

| Isigaba | Iyadingeka | Incazelo |
|------|------|------|
| `service` | ✅ | `google` \| `openai` \| `anthropic` \| `openrouter` \| `ollama` \| ngokwezifiso |
| `key` | ✅ | Ukhiye we-API ngombhalo ovulekile |
| `label` | — | Ilebula yokukhomba |
| `daily_limit` | — | Umkhawulo wokusebenzisa usuku ngalunye (0 = akunamkhawulo) |

---

#### `DELETE /admin/keys/{id}`

Ukususa ukhiye we-API. Ngemva kokususa isigameko se-SSE `key_deleted` sisakazwa.

**Impendulo:**
```json
{"status": "deleted"}
```

---

#### `POST /admin/keys/reset`

Ukusetha kabusha inani lokusebenzisa losuku lwonke okhiye. Isigameko se-SSE `usage_reset` sisakazwa.

**Impendulo:**
```json
{
  "status": "reset",
  "time": "2026-03-13T15:00:00Z"
}
```

---

### I-API yomphathi — Amaklayenti

#### `GET /admin/clients`

Uhlu lwawo wonke amaklayenti (nethokheni).

---

#### `POST /admin/clients`

Ukubhalisa iklayenti elisha.

**Umzimba wesicelo:**
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

| Isigaba | Iyadingeka | Incazelo |
|------|------|------|
| `id` | ✅ | I-ID ehlukile yeklayenti |
| `name` | — | Igama lokubonisa |
| `token` | — | Ithokheni yokuqinisekisa (uma ishiyiwe iyazalwa ngokuzenzakalelayo) |
| `default_service` | — | Isevisi evamile |
| `default_model` | — | Imodeli evamile |
| `allowed_services` | — | Uhlu lwamasevisi avunyelwe (uhlu olungenalutho = konke) |
| `agent_type` | — | `openclaw` \| `claude-code` \| `cursor` \| `vscode` \| `custom` |
| `work_dir` | — | Uhla lomsebenzi we-ejenti |
| `description` | — | Incazelo ye-ejenti |
| `ip_whitelist` | — | Uhlu lwama-IP avunyelwe (uhlu olungenalutho = konke, i-CIDR iyasekelwa) |
| `enabled` | — | Isimo sokusebenza (okuvamile `true`) |

---

#### `GET /admin/clients/{id}`

Ukuhlola iklayenti ethile (nethokheni).

---

#### `PUT /admin/clients/{id}`

Ukushintsha izilungiselelo zeklayenti. **Isigameko se-SSE `config_change` sisakazwa → sibonakala ku-proxy phakathi kwamasekhondi ayi-1–3.**

**Umzimba wesicelo (izigaba zokushintsha kuphela):**
```json
{
  "default_service": "openrouter",
  "default_model": "anthropic/claude-opus-4-6",
  "enabled": true
}
```

**Impendulo:**
```json
{"status": "updated"}
```

---

#### `DELETE /admin/clients/{id}`

Ukususa iklayenti.

---

### I-API yomphathi — Amasevisi

#### `GET /admin/services`

Uhlu lwamasevisi abhaliswe.

**Isibonelo sempendulo:**
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

Amasevisi angu-8 ahlinzekwa ngokuvamile: `google`, `openai`, `anthropic`, `openrouter`, `github-copilot`, `ollama`, `lmstudio`, `vllm`

---

#### `POST /admin/services`

Ukwengeza isevisi ngokwezifiso. Ngemva kokwengeza isigameko se-SSE `service_changed` sisakazwa → **i-dropdown yedashibhodi ibuyekezwa masinyane**.

**Umzimba wesicelo:**
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

Ukubuyekeza izilungiselelo zesevisi. Ngemva kwezinguquko isigameko se-SSE `service_changed` sisakazwa.

**Umzimba wesicelo:**
```json
{
  "local_url": "http://192.168.x.x:11434",
  "enabled": true
}
```

---

#### `DELETE /admin/services/{id}`

Ukususa isevisi ngokwezifiso. Ngemva kokususa isigameko se-SSE `service_changed` sisakazwa.

Umzamo wokususa isevisi evamile (`custom: false`):
```json
{"error": "기본 서비스는 삭제할 수 없습니다: google"}
```

---

### I-API yomphathi — Uhlu lwamamodeli

#### `GET /admin/models`

Uhlu lwamamodeli ngesevisi ngayinye. Isebenzisa i-TTL cache (imizuzu eyi-10).

**Amapharamitha ombuzo:**

| Ipharamitha | Incazelo | Isibonelo |
|---------|------|------|
| `service` | Isihlungi sesevisi | `?service=google` |
| `q` | Ukusesha imodeli | `?q=gemini` |

**Indlela yokuthola amamodeli ngesevisi:**

| Isevisi | Indlela | Inani |
|--------|------|------|
| `google` | Uhlu olungashintshi | 8 (ne-embedding) |
| `openai` | Uhlu olungashintshi | 9 |
| `anthropic` | Uhlu olungashintshi | 6 |
| `github-copilot` | Uhlu olungashintshi | 6 |
| `openrouter` | Umbuzo we-API oguquguqukayo (uma wehluleka curated fallback 14) | 340+ |
| `ollama` | Umbuzo weseva yendawo oguquguqukayo (uma ingaphenduli izincomo 7) | iyaguquka |
| `lmstudio` | Umbuzo weseva yendawo oguquguqukayo | iyaguquka |
| `vllm` | Umbuzo weseva yendawo oguquguqukayo | iyaguquka |
| ngokwezifiso | OpenAI ehambisanayo `/v1/models` | iyaguquka |

**Uhlu lwamamodeli okuhlehla e-OpenRouter (uma i-API ingaphenduli):**

| Imodeli | Amanothi akhethekile |
|------|----------|
| `openrouter/hunter-alpha` | mahhala, 1M context |
| `openrouter/healer-alpha` | mahhala, omni-modal |
| `moonshot/kimi-k2.5` | 256K context |
| `z-ai/glm-5`, `z-ai/glm-4.7-flash` | — |
| `deepseek/deepseek-r1`, `deepseek/deepseek-chat` | — |
| `qwen/qwen-2.5-72b-instruct` | 131K context |
| `minimax/minimax-m2.5` | — |
| `meta-llama/llama-3.3-70b-instruct` | 131K context |

---

### I-API yomphathi — Isimo se-proxy

#### `GET /admin/proxies`

Isimo se-Heartbeat sokugcina sawo wonke ama-proxy axhunyiwe.

---

## Izinhlobo zezigameko ze-SSE

Izigameko ezitholakala kumfudlana we-`/api/events` wesigcini:

| `type` | Isimo sokwenzeka | Okuqukethwe ku-`data` | Impendulo yedashibhodi |
|--------|-----------|-------------|--------------|
| `connected` | Masinyane ngemva kokuxhuma kwe-SSE | `{"clients": N}` | — |
| `config_change` | Ukushintsha kwezilungiselelo zeklayenti | `{"client_id","service","model"}` | I-dropdown yemodeli yekhadi le-ejenti ibuyekezwa |
| `key_added` | Ukubhaliswa kokhiye omusha we-API | `{"service": "google"}` | I-dropdown yemodeli ibuyekezwa |
| `key_deleted` | Ukhiye we-API ususwe | `{"service": "google"}` | I-dropdown yemodeli ibuyekezwa |
| `service_changed` | Ukwengeza/ukuhlela/ukususa isevisi | `{"action":"added"\|"updated"\|"deleted","id":"...","proxy_services":["google","ollama"]}` | I-select yesevisi + i-dropdown yemodeli ibuyekezwa masinyane; uhlu lwamasevisi okuthumela lwe-proxy lubuyekezwa ngesikhathi sangempela |
| `usage_update` | Uma i-heartbeat ye-proxy itholakala (njalo ngamasekhondi angu-20) | `{"keys":[{"id","service","today_usage","today_attempts","daily_limit","cooldown_until"},...]}` | Ibha·izinombolo zokusebenzisa zokhiye zibuyekezwa masinyane, ukubala kokupholisa kuqala. Idatha ye-SSE isetshenziswa ngokuqondile ngaphandle kwe-fetch. Ibha isebenzisa ukukala kwesabelo-sesamba (okhiye abangenamkhawulo). |
| `usage_reset` | Inani lokusebenzisa losuku lisetwa kabusha | `{"time": "RFC3339"}` | Ikhasi libuyekezwa |

**Ukucutshungulwa kwezigameko ezitholwa yi-proxy:**

```
config_change itholakele
  → uma client_id ilingana neyayo
    → service, model ibuyekezwa masinyane
    → hooksMgr.Fire(EventModelChanged)
```

---

## Ukulayisha komhlinzeki·wemodeli

Uma ubeke isakhiwo se-`provider/model` esigabeni se-`model` se-`/v1/chat/completions` ukulayisha okuzenzakalelayo kwenziwa (kuhambisana ne-OpenClaw 3.11).

### Imithetho yokulayisha ngesijobelelo sangaphambili

| Isijobelelo sangaphambili | Lapho kuya khona | Isibonelo |
|--------|------------|------|
| `google/` | Google ngokuqondile | `google/gemini-2.5-pro` |
| `openai/` | OpenAI ngokuqondile | `openai/gpt-4o` |
| `anthropic/` | Nge-OpenRouter | `anthropic/claude-opus-4-6` |
| `ollama/` | Ollama ngokuqondile | `ollama/qwen3.5:35b` |
| `custom/` | Ukuhlaziya kabusha ngokuphindaphinda (susa `custom/` bese ulayisha kabusha) | `custom/google/gemini-2.5-flash` → Google |
| `openrouter/` | OpenRouter (indlela esobala igcinwa) | `openrouter/meta-llama/llama-3.3-70b` |
| `opencode/`, `opencode-go/`, `opencode-zen/` | OpenRouter | `opencode-go/model` |
| `moonshot/`, `kimi-coding/` | OpenRouter (indlela egcwele igcinwa) | `moonshot/kimi-k2.5` |
| `groq/`, `mistral/`, `minimax/`, `cohere/`, `perplexity/` | OpenRouter | `groq/llama-3.1-70b` |
| `together/`, `huggingface/`, `nvidia/`, `venice/` | OpenRouter | — |
| `meta-llama/`, `qwen/`, `deepseek/`, `01-ai/` | OpenRouter (indlela egcwele) | `deepseek/deepseek-r1` |

### Ukuthola ngokuzenzakalelayo kwesijobelelo sangaphambili se-`wall-vault/`

i-wall-vault ibona isevisi ngokuzenzakalelayo ku-ID yemodeli ngesojobelelo sayo.

| Isakhiwo se-ID yemodeli | Ukulayisha |
|-------------|--------|
| `wall-vault/gemini-*` | Google |
| `wall-vault/gpt-*`, `wall-vault/o1`, `wall-vault/o3`, `wall-vault/o4*` | OpenAI |
| `wall-vault/claude-*` | OpenRouter (indlela ye-Anthropic) |
| `wall-vault/hunter-alpha`, `wall-vault/healer-alpha` | OpenRouter (mahhala 1M ctx) |
| `wall-vault/kimi-*`, `wall-vault/glm-*`, `wall-vault/deepseek-*`, `wall-vault/qwen*` | OpenRouter |
| Okunye | OpenRouter |

### Ukucutshungulwa kwesijobelelo sangemuva se-`:cloud`

Isijobelelo sangemuva se-`:cloud` sesakhiwo se-tag ye-Ollama sisuswa ngokuzenzakalelayo bese kulayishwa ku-OpenRouter.

```
kimi-k2.5:cloud  →  OpenRouter, I-ID yemodeli: kimi-k2.5
glm-5:cloud      →  OpenRouter, I-ID yemodeli: glm-5
```

### Isibonelo sokuxhumanisa i-OpenClaw openclaw.json

```json5
// ~/.openclaw/openclaw.json
{
  models: {
    providers: {
      "wall-vault": {
        baseUrl: "http://localhost:56244/v1",
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

Ukuchofoza inkinobho ye-**🐾** ekhadini le-ejenti kukopisha ngokuzenzakalelayo i-snippet yezilungiselelo ze-ejenti leyo kubhodi lokunamathisela.

---

## Isikimu sedatha

### APIKey

| Isigaba | Uhlobo | Incazelo |
|------|------|------|
| `id` | string | I-ID ehlukile yesakhiwo se-UUID |
| `service` | string | `google` \| `openai` \| `anthropic` \| `openrouter` \| `ollama` \| ngokwezifiso |
| `encrypted_key` | string | Ukhiye obethwe nge-AES-GCM (Base64) |
| `label` | string | Ilebula yokukhomba |
| `today_usage` | int | Inani lamathokheni ezicelo eziphumelele namuhla (amaphutha e-429/402/582 awafakiwe) |
| `today_attempts` | int | Isamba samakholi e-API namuhla (aphumelele + akhawulelwe isilinganiso; lisetwa kabusha phakathi kwamabili) |
| `daily_limit` | int | Umkhawulo wosuku (0 = akunamkhawulo) |
| `cooldown_until` | time.Time | Isikhathi sokuphela kokupholisa |
| `last_error` | int | Ikhodi yokugcina yephutha le-HTTP |
| `created_at` | time.Time | Isikhathi sokubhalisa |

**Inqubomgomo yokupholisa:**

| Iphutha le-HTTP | Ukupholisa |
|-----------|--------|
| 429 (Too Many Requests) | Imizuzu engu-30 |
| 402 (Payment Required) | Amahora angu-24 |
| 400 / 401 / 403 | Amahora angu-24 |
| 582 (Gateway Overload) | Imizuzu engu-5 |
| Iphutha lenethiwekhi | Imizuzu engu-10 |

> **429·402·582**: Ukupholisa kubekwa + `today_attempts` iyenyuka. `today_usage` ayishintshi (amathokheni aphumelele kuphela abalwa).
> **Ollama (isevisi yendawo)**: `callOllama` isebenzisa iklayenti ye-HTTP ezinikele `Timeout: 0` (akunamkhawulo). Ukuqagela kwamamodeli amakhulu kungathatha amasekhondi amashumi kuya emizuzwini eminingana, ngakho i-timeout evamile yamasekhondi angu-60 ayisebenziswa.

### Client

| Isigaba | Uhlobo | Incazelo |
|------|------|------|
| `id` | string | I-ID ehlukile yeklayenti |
| `name` | string | Igama lokubonisa |
| `token` | string | Ithokheni yokuqinisekisa |
| `default_service` | string | Isevisi evamile |
| `default_model` | string | Imodeli evamile (isakhiwo se-`provider/model` siyenzeka) |
| `allowed_services` | []string | Amasevisi avunyelwe (uhlu olungenalutho = konke) |
| `agent_type` | string | `openclaw` \| `claude-code` \| `cursor` \| `vscode` \| `custom` |
| `work_dir` | string | Uhla lomsebenzi we-ejenti |
| `description` | string | Incazelo |
| `ip_whitelist` | []string | Uhlu lwama-IP avunyelwe (i-CIDR iyasekelwa) |
| `avatar` | string | Agent avatar — relative path under `~/.openclaw/` (e.g. `workspace/avatar.png`, `workspace/avatars/profile.hpg`) OR base64 data URI (`data:image/png;base64,...`). Supported extensions: `.png`, `.jpg`/`.jpeg`/`.hpg`, `.webp`, `.gif`. |
| `enabled` | bool | Uma ingu-`false` ukufinyelela `/api/keys` kubuyisa `403` |
| `created_at` | time.Time | Isikhathi sokubhalisa |

### ServiceConfig

| Isigaba | Uhlobo | Incazelo |
|------|------|------|
| `id` | string | I-ID ehlukile yesevisi |
| `name` | string | Igama lokubonisa |
| `local_url` | string | I-URL yeseva yendawo (Ollama/LMStudio/vLLM/ngokwezifiso) |
| `enabled` | bool | Isimo sokusebenza |
| `custom` | bool | Noma isevisi yengezwe ngumsebenzisi yini |
| `proxy_enabled` | bool | Whether this service is included in proxy dispatch and agent model dropdowns. Controlled by the "프록시 사용" checkbox in the Services card. Only `proxy_enabled: true` services appear in agent service/model selectors. |

### ProxyStatus (Heartbeat)

| Isigaba | Uhlobo | Incazelo |
|------|------|------|
| `client_id` | string | I-ID yeklayenti |
| `version` | string | Inguqulo ye-proxy (isb. `v0.1.6.20260314.231308`) |
| `service` | string | Isevisi yamanje |
| `model` | string | Imodeli yamanje |
| `sse_connected` | bool | Isimo sokuxhumana kwe-SSE |
| `host` | string | Igama lomsingathi |
| `avatar` | string | base64 data URI of local avatar (auto-synced from proxy via heartbeat) |
| `updated_at` | time.Time | Ukubuyekezwa kokugcina |
| `vault.today_usage` | int | Ukusetshenziswa kwamathokheni kwanamuhla |
| `vault.daily_limit` | int | Umkhawulo wosuku |
| `vault.key_status` | string | `active` \| `cooldown` \| `exhausted` |

---

## Izimpendulo zamaphutha

```json
{"error": "오류 메시지"}
```

| Ikhodi | Incazelo |
|------|------|
| 200 | Kuphumelele |
| 400 | Isicelo esingalungile |
| 401 | Ukuqinisekisa kuhlulekile |
| 403 | Ukufinyelela kwenqatshiwe (iklayenti engasebenzi, i-IP evalelwe) |
| 404 | Umthombo awutholakalanga |
| 405 | Indlela engavunyelwe |
| 429 | Umkhawulo wesilinganiso weqiwe |
| 500 | Iphutha langaphakathi leseva |
| 502 | Iphutha le-API yasenhla (konke ukuhlehla kuhlulekile) |

---

## Iqoqo lezibonelo ze-cURL

```bash
# ─── I-Proxy ───────────────────────────────────────────────────────────────────

# Ukuhlola impilo
curl http://localhost:56244/health

# Ukuhlola isimo
curl http://localhost:56244/status

# Uhlu lwamamodeli (konke)
curl http://localhost:56244/api/models

# Amamodeli e-Google kuphela
curl "http://localhost:56244/api/models?service=google"

# Ukusesha amamodeli amahhala
curl "http://localhost:56244/api/models?q=alpha"

# Ukushintsha imodeli (yendawo)
curl -X PUT http://localhost:56244/api/config/model \
  -H "Content-Type: application/json" \
  -d '{"service":"google","model":"gemini-2.5-pro"}'

# Ukuhlaziya izilungiselelo
curl -X POST http://localhost:56244/reload

# Ukushayela i-Gemini API ngokuqondile
curl -X POST "http://localhost:56244/google/v1beta/models/gemini-2.5-flash:generateContent" \
  -H "Content-Type: application/json" \
  -d '{"contents":[{"role":"user","parts":[{"text":"안녕하세요"}]}]}'

# Ehambisanayo ne-OpenAI (imodeli evamile)
curl -X POST http://localhost:56244/v1/chat/completions \
  -H "Content-Type: application/json" \
  -d '{"model":"gemini-2.5-flash","messages":[{"role":"user","content":"안녕"}]}'

# Isakhiwo se-provider/model se-OpenClaw
curl -X POST http://localhost:56244/v1/chat/completions \
  -H "Authorization: Bearer my-agent-token" \
  -H "Content-Type: application/json" \
  -d '{"model":"wall-vault/gemini-2.5-flash","messages":[{"role":"user","content":"안녕"}]}'

# Ukusebenzisa imodeli ye-1M context yamahhala
curl -X POST http://localhost:56244/v1/chat/completions \
  -H "Authorization: Bearer my-agent-token" \
  -H "Content-Type: application/json" \
  -d '{"model":"wall-vault/hunter-alpha","messages":[{"role":"user","content":"안녕"}]}'

# ─── Isigcini sokhiye (umphakathi) ───────────────────────────────────────────────────────────

curl http://localhost:56243/api/status
curl http://localhost:56243/api/clients
curl -s http://localhost:56243/api/events --max-time 3

# ─── Isigcini sokhiye (umphathi) ─────────────────────────────────────────────────────────

ADMIN="Authorization: Bearer admin-token"

# Uhlu lokhiye
curl -H "$ADMIN" http://localhost:56243/admin/keys

# Ukwengeza ukhiye we-Google
curl -X POST http://localhost:56243/admin/keys \
  -H "$ADMIN" -H "Content-Type: application/json" \
  -d '{"service":"google","key":"AIzaSy...","label":"메인 키","daily_limit":1000}'

# Ukwengeza ukhiye we-OpenAI
curl -X POST http://localhost:56243/admin/keys \
  -H "$ADMIN" -H "Content-Type: application/json" \
  -d '{"service":"openai","key":"sk-...","label":"GPT 키"}'

# Ukwengeza ukhiye we-OpenRouter
curl -X POST http://localhost:56243/admin/keys \
  -H "$ADMIN" -H "Content-Type: application/json" \
  -d '{"service":"openrouter","key":"sk-or-v1-...","label":"OR 키"}'

# Ukususa ukhiye (SSE key_deleted iyasakazwa)
curl -X DELETE http://localhost:56243/admin/keys/key-abc123 -H "$ADMIN"

# Ukusetha kabusha inani lokusebenzisa losuku
curl -X POST http://localhost:56243/admin/keys/reset -H "$ADMIN"

# Uhlu lwamaklayenti
curl -H "$ADMIN" http://localhost:56243/admin/clients

# Ukwengeza iklayenti (OpenClaw)
curl -X POST http://localhost:56243/admin/clients \
  -H "$ADMIN" -H "Content-Type: application/json" \
  -d '{"id":"bot-a","name":"봇 A","agent_type":"openclaw","work_dir":"~/.openclaw","default_service":"google","default_model":"gemini-2.5-flash"}'

# Ukushintsha imodeli yeklayenti (SSE masinyane)
curl -X PUT http://localhost:56243/admin/clients/bot-a \
  -H "$ADMIN" -H "Content-Type: application/json" \
  -d '{"default_service":"openrouter","default_model":"wall-vault/hunter-alpha"}'

# Ukukhubaza iklayenti
curl -X PUT http://localhost:56243/admin/clients/my-bot \
  -H "$ADMIN" -H "Content-Type: application/json" \
  -d '{"enabled":false}'

# Ukususa iklayenti
curl -X DELETE http://localhost:56243/admin/clients/my-bot -H "$ADMIN"

# Uhlu lwamasevisi
curl -H "$ADMIN" http://localhost:56243/admin/services

# Ukusetha i-URL yendawo ye-Ollama (SSE service_changed iyasakazwa)
curl -X PUT http://localhost:56243/admin/services/ollama \
  -H "$ADMIN" -H "Content-Type: application/json" \
  -d '{"local_url":"http://192.168.x.x:11434","enabled":true}'

# Ukuvula isevisi ye-OpenAI
curl -X PUT http://localhost:56243/admin/services/openai \
  -H "$ADMIN" -H "Content-Type: application/json" \
  -d '{"enabled":true}'

# Ukwengeza isevisi ngokwezifiso (SSE service_changed iyasakazwa)
curl -X POST http://localhost:56243/admin/services \
  -H "$ADMIN" -H "Content-Type: application/json" \
  -d '{"id":"my-llm","name":"사내 LLM","local_url":"http://10.0.0.50:8080","enabled":true}'

# Ukususa isevisi ngokwezifiso
curl -X DELETE http://localhost:56243/admin/services/my-llm -H "$ADMIN"

# Ukuhlola uhlu lwamamodeli
curl -H "$ADMIN" http://localhost:56243/admin/models
curl -H "$ADMIN" "http://localhost:56243/admin/models?service=openrouter"
curl -H "$ADMIN" "http://localhost:56243/admin/models?q=hunter"

# Isimo se-proxy (heartbeat)
curl -H "$ADMIN" http://localhost:56243/admin/proxies

# ─── Imodi yokusabalalisa — I-Proxy → Isigcini ───────────────────────────────────────────────

# Ukuhlola okhiye abavulwe esifihliwe
curl http://localhost:56243/api/keys \
  -H "Authorization: Bearer your-bot-a-token"

# Ukuthumela i-Heartbeat
curl -X POST http://localhost:56243/api/heartbeat \
  -H "Authorization: Bearer your-bot-a-token" \
  -H "Content-Type: application/json" \
  -d '{"client_id":"bot-a","version":"v0.1.6.20260314.231308","service":"google","model":"gemini-2.5-flash","sse_connected":true}'
```

---

## I-Middleware

Isetshenziswa ngokuzenzakalelayo kuzo zonke izicelo:

| I-Middleware | Umsebenzi |
|---------|------|
| **Logger** | Ukubhala ngomkhakha we-`[method] path status latencyms` |
| **CORS** | `Access-Control-Allow-Origin: *` |
| **Recovery** | Ukuvuselela ekwesabeni, ukubuyisa impendulo ye-500 |

---

*Ibuyekezwe kokugcina: 2026-03-16 — v0.1.7: dashboard title renamed to "벽금고(wall-vault) 대시보드", logo moved to sticky topbar, today_attempts tracking, HTTP 582 cooldown (5min), share-of-total bar scaling, custom/ routing fix, Ollama Timeout:0, key_att i18n, avatar heartbeat sync, build timestamp versioning, proxy-only service filter*
