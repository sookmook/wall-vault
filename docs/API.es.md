# Manual de API de wall-vault

Este documento describe en detalle todos los endpoints de la API HTTP de wall-vault.

---

## Tabla de contenidos

- [Autenticación](#autenticación)
- [API del Proxy (:56244)](#api-del-proxy-56244)
  - [Verificación de salud](#get-health)
  - [Estado](#get-status)
  - [Lista de modelos](#get-apimodels)
  - [Cambiar modelo](#put-apiconfigmodel)
  - [Modo pensamiento](#put-apiconfigthink-mode)
  - [Recargar configuración](#post-reload)
  - [API Gemini](#post-googlev1betamodelsmgeneratecontent)
  - [Streaming Gemini](#post-googlev1betamodelsmstreamgeneratecontent)
  - [API compatible con OpenAI](#post-v1chatcompletions)
- [API de la Bóveda de claves (:56243)](#api-de-la-bóveda-de-claves-56243)
  - [API pública](#api-pública-sin-autenticación-requerida)
  - [Flujo de eventos SSE](#get-apievents)
  - [API exclusiva del proxy](#api-exclusiva-del-proxy-token-de-cliente)
  - [API Admin — Claves](#api-admin--claves-api)
  - [API Admin — Clientes](#api-admin--clientes)
  - [API Admin — Servicios](#api-admin--servicios)
  - [API Admin — Lista de modelos](#api-admin--lista-de-modelos)
  - [API Admin — Estado del proxy](#api-admin--estado-del-proxy)
- [Tipos de eventos SSE](#tipos-de-eventos-sse)
- [Enrutamiento de proveedor/modelo](#enrutamiento-de-proveedormodelo)
- [Esquema de datos](#esquema-de-datos)
- [Respuestas de error](#respuestas-de-error)
- [Ejemplos cURL](#ejemplos-curl)

---

## Autenticación

| Ámbito | Método | Encabezado |
|--------|--------|------------|
| API Admin | Token Bearer | `Authorization: Bearer <admin_token>` |
| Proxy → Bóveda | Token Bearer | `Authorization: Bearer <client_token>` |
| API del Proxy | Ninguno (local) | — |

Si `admin_token` no está configurado (cadena vacía), todas las APIs de administración son accesibles sin autenticación.

### Política de seguridad

- **Limitación de velocidad**: Si la autenticación de la API admin falla más de 10 veces en 15 minutos, la IP se bloquea temporalmente (`429 Too Many Requests`)
- **Lista blanca de IP**: Solo las IP/CIDR registradas en el campo `ip_whitelist` del agente (`Client`) tienen acceso a `/api/keys`. Si el array está vacío, todas las IP están permitidas.
- **Protección de theme/lang**: `/admin/theme` y `/admin/lang` también requieren autenticación con token de administrador

---

## API del Proxy (:56244)

El servidor donde se ejecuta el proxy. Puerto predeterminado `56244`.

---

### `GET /health`

Verificación de salud. Siempre devuelve 200 OK.

**Ejemplo de respuesta:**
```json
{
  "status": "ok",
  "version": "v0.1.6.20260314.231308",
  "client": "bot-a"
}
```

---

### `GET /status`

Estado detallado del proxy.

**Ejemplo de respuesta:**
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

| Campo | Tipo | Descripción |
|-------|------|-------------|
| `service` | string | Servicio predeterminado actual |
| `model` | string | Modelo predeterminado actual |
| `sse` | bool | Si la conexión SSE con la bóveda está establecida |
| `filter` | string | Modo de filtrado de herramientas |
| `services` | []string | Lista de servicios activos |
| `mode` | string | `standalone` \| `distributed` |

---

### `GET /api/models`

Listar modelos disponibles. Usa caché TTL (10 minutos por defecto).

**Parámetros de consulta:**

| Parámetro | Descripción | Ejemplo |
|-----------|-------------|---------|
| `service` | Filtro por servicio | `?service=google` |
| `q` | Búsqueda por ID/nombre de modelo | `?q=gemini` |

**Ejemplo de respuesta:**
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

| Campo | Tipo | Descripción |
|-------|------|-------------|
| `id` | string | ID del modelo |
| `name` | string | Nombre para mostrar del modelo |
| `service` | string | `google` \| `openai` \| `anthropic` \| `openrouter` \| `ollama` etc. |
| `context_length` | int | Tamaño de la ventana de contexto |
| `free` | bool | Si es un modelo gratuito (OpenRouter) |

---

### `PUT /api/config/model`

Cambiar el servicio y modelo actuales.

**Cuerpo de la solicitud:**
```json
{
  "service": "google",
  "model": "gemini-2.5-flash"
}
```

**Respuesta:**
```json
{
  "status": "ok",
  "service": "google",
  "model": "gemini-2.5-flash"
}
```

> **Nota:** En modo distribuido, se recomienda usar `PUT /admin/clients/{id}` de la bóveda en lugar de esta API. Los cambios en la bóveda se propagan automáticamente a través de SSE en 1 a 3 segundos.

---

### `PUT /api/config/think-mode`

Alternar modo pensamiento (no-op, reservado para expansión futura).

**Respuesta:**
```json
{"status": "ok"}
```

---

### `POST /reload`

Resincronizar inmediatamente la configuración del cliente y las claves desde la bóveda.

**Respuesta:**
```json
{"status": "reloading"}
```

La resincronización se ejecuta de forma asíncrona y se completa en 1 a 2 segundos después de recibir la respuesta.

---

### `POST /google/v1beta/models/{model}:generateContent`

Proxy de API Gemini (sin streaming).

**Parámetro de ruta:**
- `{model}`: ID del modelo. Si tiene el prefijo `gemini-`, el servicio de Google se selecciona automáticamente.

**Cuerpo de la solicitud:** [Formato de solicitud Gemini generateContent](https://ai.google.dev/api/generate-content)

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

**Cuerpo de la respuesta:** Formato de respuesta Gemini generateContent

**Filtro de herramientas:** Cuando `tool_filter: strip_all` está configurado, el array `tools` de la solicitud se elimina automáticamente.

**Cadena de respaldo:** Si el servicio designado falla → respaldo en el orden de servicios configurado → Ollama (último recurso).

---

### `POST /google/v1beta/models/{model}:streamGenerateContent`

Proxy de streaming de API Gemini. El formato de solicitud es idéntico al no streaming. La respuesta es un flujo SSE:

```
data: {"candidates":[{"content":{"parts":[{"text":"안"}],...},...}]}

data: [DONE]
```

---

### `POST /v1/chat/completions`

API compatible con OpenAI. Convierte internamente al formato Gemini para su procesamiento.

**Cuerpo de la solicitud:**
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

**Soporte de prefijo de proveedor en el campo `model` (OpenClaw 3.11+):**

| Ejemplo de modelo | Enrutamiento |
|-------------------|-------------|
| `gemini-2.5-flash` | Servicio configurado actual |
| `google/gemini-2.5-pro` | Google directo |
| `openai/gpt-4o` | OpenAI directo |
| `anthropic/claude-opus-4-6` | Vía OpenRouter |
| `openrouter/meta-llama/llama-3.3-70b` | OpenRouter directo |
| `wall-vault/gemini-2.5-flash` | Detección automática → Google |
| `wall-vault/claude-opus-4-6` | Detección automática → OpenRouter Anthropic |
| `wall-vault/gpt-4o` | Detección automática → OpenAI |
| `wall-vault/hunter-alpha` | OpenRouter (gratuito, 1M contexto) |
| `moonshot/kimi-k2.5` | Vía OpenRouter |
| `opencode-go/model` | Vía OpenRouter |
| `kimi-k2.5:cloud` | Sufijo `:cloud` → OpenRouter |

Para más detalles, consulte [Enrutamiento de proveedor/modelo](#enrutamiento-de-proveedormodelo).

**Cuerpo de la respuesta:**
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

> **Eliminación automática de tokens de control de modelo:** Si la respuesta contiene delimitadores GLM-5 / DeepSeek / ChatML (`<|im_start|>`, `[gMASK]`, `[sop]`, etc.), se eliminan automáticamente.

---

## API de la Bóveda de claves (:56243)

El servidor donde se ejecuta la bóveda de claves. Puerto predeterminado `56243`.

---

### API pública (Sin autenticación requerida)

#### `GET /`

Interfaz web del panel de control. Acceso a través del navegador.

---

#### `GET /api/status`

Estado de la bóveda.

**Ejemplo de respuesta:**
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

Lista de clientes registrados (solo información pública, tokens excluidos).

---

### `GET /api/events`

Flujo de eventos SSE (Server-Sent Events) en tiempo real.

**Encabezados:**
```
Content-Type: text/event-stream
Cache-Control: no-cache
Connection: keep-alive
```

**Recibido inmediatamente al conectar:**
```
data: {"type":"connected","clients":2}
```

**Ejemplos de eventos:**
```
data: {"type":"config_change","data":{"client_id":"bot-a","service":"google","model":"gemini-2.5-pro"}}

data: {"type":"key_added","data":{"service":"openrouter"}}

data: {"type":"key_deleted","data":{"service":"google"}}

data: {"type":"service_changed","data":{"action":"updated","id":"ollama"}}

data: {"type":"usage_reset","data":{"time":"2026-03-13T00:00:30Z"}}
```

Para tipos de eventos detallados, consulte [Tipos de eventos SSE](#tipos-de-eventos-sse).

---

### API exclusiva del proxy (Token de cliente)

Requiere el encabezado `Authorization: Bearer <client_token>`. Los tokens de administrador también son aceptados.

#### `GET /api/keys`

Lista de claves API descifradas proporcionadas al proxy.

**Parámetros de consulta:**

| Parámetro | Descripción |
|-----------|-------------|
| `service` | Filtro por servicio (ej.: `?service=google`) |

**Ejemplo de respuesta:**
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

> **Seguridad:** Devuelve claves en texto plano. Solo se devuelven claves para servicios permitidos según la configuración `allowed_services` del cliente.

---

#### `GET /api/services`

Lista de servicios para el proxy. Devuelve un array de IDs de servicios donde `proxy_enabled=true`.

**Ejemplo de respuesta:**
```json
["google", "ollama"]
```

Si el array está vacío, el proxy usa todos los servicios sin restricción.

---

#### `POST /api/heartbeat`

Enviar estado del proxy (ejecutado automáticamente cada 20 segundos).

**Cuerpo de la solicitud:**
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

| Campo | Tipo | Descripción |
|-------|------|-------------|
| `client_id` | string | ID del cliente |
| `version` | string | Versión del proxy (incluye marca de tiempo de compilación, ej. `v0.1.6.20260314.231308`) |
| `service` | string | Servicio actual |
| `model` | string | Modelo actual |
| `sse_connected` | bool | Si SSE está conectado |
| `host` | string | Nombre del host |
| `avatar` | string | base64 data URI of the proxy's local avatar image (sent by proxy, auto-persisted to client record by vault). Set via `WV_AVATAR` env var (relative path under `~/.openclaw/`). |
| `key_usage` | map[string]int | Key ID → successful tokens used today. Vault calls `SetKeyUsage()` (absolute value sync). 429/402/582 errors do **not** increment this. |
| `key_attempts` | map[string]int | Key ID → total API calls today (success + rate-limited). Vault calls `SetKeyAttempts()`. 429/402/582 errors only increment this, not `key_usage`. |
| `key_cooldowns` | map[string]string | Key ID → cooldown end time (RFC3339). Vault calls `SetKeyCooldownIfLater()`. |

**Respuesta:**
```json
{"status": "ok"}
```

---

### API Admin — Claves API

Requiere el encabezado `Authorization: Bearer <admin_token>`.

#### `GET /admin/keys`

Listar todas las claves API registradas (claves en texto plano excluidas).

**Ejemplo de respuesta:**
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

| Campo | Tipo | Descripción |
|-------|------|-------------|
| `today_usage` | int | Tokens de solicitudes exitosas hoy (no incluye errores 429/402/582) |
| `today_attempts` | int | Total de llamadas API hoy (éxito + limitadas por velocidad) |
| `available` | bool | Si está disponible sin tiempo de enfriamiento ni límite |
| `usage_pct` | int | Porcentaje de uso del límite diario (`daily_limit=0` → 0) |
| `cooldown_until` | RFC3339 | Fin del tiempo de enfriamiento (valor cero significa ninguno) |
| `last_error` | int | Último código de error HTTP |

---

#### `POST /admin/keys`

Registrar una nueva clave API. Un evento SSE `key_added` se transmite inmediatamente tras el registro.

**Cuerpo de la solicitud:**
```json
{
  "service": "google",
  "key": "AIzaSy...",
  "label": "메인 키",
  "daily_limit": 1000
}
```

| Campo | Requerido | Descripción |
|-------|-----------|-------------|
| `service` | ✅ | `google` \| `openai` \| `anthropic` \| `openrouter` \| `ollama` \| personalizado |
| `key` | ✅ | Clave API en texto plano |
| `label` | — | Etiqueta de identificación |
| `daily_limit` | — | Límite de uso diario (0 = ilimitado) |

---

#### `DELETE /admin/keys/{id}`

Eliminar una clave API. Un evento SSE `key_deleted` se transmite después de la eliminación.

**Respuesta:**
```json
{"status": "deleted"}
```

---

#### `POST /admin/keys/reset`

Restablecer el uso diario de todas las claves. Se transmite el evento SSE `usage_reset`.

**Respuesta:**
```json
{
  "status": "reset",
  "time": "2026-03-13T15:00:00Z"
}
```

---

### API Admin — Clientes

#### `GET /admin/clients`

Listar todos los clientes (tokens incluidos).

---

#### `POST /admin/clients`

Registrar un nuevo cliente.

**Cuerpo de la solicitud:**
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

| Campo | Requerido | Descripción |
|-------|-----------|-------------|
| `id` | ✅ | ID único del cliente |
| `name` | — | Nombre para mostrar |
| `token` | — | Token de autenticación (se genera automáticamente si se omite) |
| `default_service` | — | Servicio predeterminado |
| `default_model` | — | Modelo predeterminado |
| `allowed_services` | — | Lista de servicios permitidos (array vacío = todos permitidos) |
| `agent_type` | — | `openclaw` \| `claude-code` \| `cursor` \| `vscode` \| `custom` |
| `work_dir` | — | Directorio de trabajo del agente |
| `description` | — | Descripción del agente |
| `ip_whitelist` | — | Lista de IP permitidas (array vacío = todas permitidas, CIDR soportado) |
| `enabled` | — | Si está habilitado (por defecto `true`) |

---

#### `GET /admin/clients/{id}`

Obtener un cliente específico (token incluido).

---

#### `PUT /admin/clients/{id}`

Actualizar la configuración de un cliente. **Transmisión SSE `config_change` → reflejado en el proxy en 1 a 3 segundos.**

**Cuerpo de la solicitud (incluir solo los campos a modificar):**
```json
{
  "default_service": "openrouter",
  "default_model": "anthropic/claude-opus-4-6",
  "enabled": true
}
```

**Respuesta:**
```json
{"status": "updated"}
```

---

#### `DELETE /admin/clients/{id}`

Eliminar un cliente.

---

### API Admin — Servicios

#### `GET /admin/services`

Listar servicios registrados.

**Ejemplo de respuesta:**
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

8 servicios integrados: `google`, `openai`, `anthropic`, `openrouter`, `github-copilot`, `ollama`, `lmstudio`, `vllm`

---

#### `POST /admin/services`

Agregar un servicio personalizado. El evento SSE `service_changed` se transmite después de agregarlo → **los menús desplegables del panel se actualizan inmediatamente**.

**Cuerpo de la solicitud:**
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

Actualizar la configuración de un servicio. El evento SSE `service_changed` se transmite después de los cambios.

**Cuerpo de la solicitud:**
```json
{
  "local_url": "http://192.168.x.x:11434",
  "enabled": true
}
```

---

#### `DELETE /admin/services/{id}`

Eliminar un servicio personalizado. El evento SSE `service_changed` se transmite después de la eliminación.

Intento de eliminar un servicio integrado (`custom: false`):
```json
{"error": "기본 서비스는 삭제할 수 없습니다: google"}
```

---

### API Admin — Lista de modelos

#### `GET /admin/models`

Listar modelos por servicio. Usa caché TTL (10 minutos).

**Parámetros de consulta:**

| Parámetro | Descripción | Ejemplo |
|-----------|-------------|---------|
| `service` | Filtro por servicio | `?service=google` |
| `q` | Búsqueda de modelo | `?q=gemini` |

**Obtención de modelos por servicio:**

| Servicio | Método | Cantidad |
|----------|--------|----------|
| `google` | Lista estática | 8 (incluyendo embedding) |
| `openai` | Lista estática | 9 |
| `anthropic` | Lista estática | 6 |
| `github-copilot` | Lista estática | 6 |
| `openrouter` | Consulta dinámica a API (respaldo a 14 modelos seleccionados en caso de fallo) | 340+ |
| `ollama` | Consulta dinámica a servidor local (7 recomendados si no responde) | Variable |
| `lmstudio` | Consulta dinámica a servidor local | Variable |
| `vllm` | Consulta dinámica a servidor local | Variable |
| Personalizado | Compatible con OpenAI `/v1/models` | Variable |

**Modelos de respaldo de OpenRouter (cuando la API no responde):**

| Modelo | Notas |
|--------|-------|
| `openrouter/hunter-alpha` | Gratuito, 1M contexto |
| `openrouter/healer-alpha` | Gratuito, omni-modal |
| `moonshot/kimi-k2.5` | 256K contexto |
| `z-ai/glm-5`, `z-ai/glm-4.7-flash` | — |
| `deepseek/deepseek-r1`, `deepseek/deepseek-chat` | — |
| `qwen/qwen-2.5-72b-instruct` | 131K contexto |
| `minimax/minimax-m2.5` | — |
| `meta-llama/llama-3.3-70b-instruct` | 131K contexto |

---

### API Admin — Estado del proxy

#### `GET /admin/proxies`

Último estado de heartbeat de todos los proxys conectados.

---

## Tipos de eventos SSE

Eventos recibidos del flujo `/api/events` de la bóveda:

| `type` | Disparador | Contenido `data` | Reacción del panel |
|--------|-----------|------------------|-------------------|
| `connected` | Inmediatamente al conectar SSE | `{"clients": N}` | — |
| `config_change` | Configuración del cliente cambiada | `{"client_id","service","model"}` | Menú desplegable de modelo de la tarjeta de agente actualizado |
| `key_added` | Nueva clave API registrada | `{"service": "google"}` | Menú desplegable de modelo actualizado |
| `key_deleted` | Clave API eliminada | `{"service": "google"}` | Menú desplegable de modelo actualizado |
| `service_changed` | Servicio agregado/actualizado/eliminado | `{"action":"added"\|"updated"\|"deleted","id":"...","proxy_services":["google","ollama"]}` | Selector de servicio + menú desplegable de modelo actualizados inmediatamente; lista de servicios de despacho del proxy actualizada en tiempo real |
| `usage_update` | Al recibir heartbeat del proxy (cada 20s) | `{"keys":[{"id","service","today_usage","today_attempts","daily_limit","cooldown_until"},...]}` | Barras y números de uso de claves actualizados instantáneamente, cuenta regresiva del tiempo de enfriamiento iniciada. Datos SSE usados directamente sin fetch. Las barras usan escalado proporcional al total (para claves ilimitadas). |
| `usage_reset` | Restablecimiento del uso diario | `{"time": "RFC3339"}` | Recarga de página |

**Procesamiento de eventos en el lado del proxy:**

```
config_change recibido
  → Si client_id coincide con el propio ID
    → service, model actualizados inmediatamente
    → hooksMgr.Fire(EventModelChanged)
```

---

## Enrutamiento de proveedor/modelo

Al especificar un formato `provider/model` en el campo `model` de `/v1/chat/completions`, se aplica enrutamiento automático (compatible con OpenClaw 3.11).

### Reglas de enrutamiento por prefijo

| Prefijo | Destino de enrutamiento | Ejemplo |
|---------|------------------------|---------|
| `google/` | Google directo | `google/gemini-2.5-pro` |
| `openai/` | OpenAI directo | `openai/gpt-4o` |
| `anthropic/` | Vía OpenRouter | `anthropic/claude-opus-4-6` |
| `ollama/` | Ollama directo | `ollama/qwen3.5:35b` |
| `custom/` | Re-análisis recursivo (eliminar `custom/` y re-enrutar) | `custom/google/gemini-2.5-flash` → Google |
| `openrouter/` | OpenRouter (ruta base conservada) | `openrouter/meta-llama/llama-3.3-70b` |
| `opencode/`, `opencode-go/`, `opencode-zen/` | OpenRouter | `opencode-go/model` |
| `moonshot/`, `kimi-coding/` | OpenRouter (ruta completa conservada) | `moonshot/kimi-k2.5` |
| `groq/`, `mistral/`, `minimax/`, `cohere/`, `perplexity/` | OpenRouter | `groq/llama-3.1-70b` |
| `together/`, `huggingface/`, `nvidia/`, `venice/` | OpenRouter | — |
| `meta-llama/`, `qwen/`, `deepseek/`, `01-ai/` | OpenRouter (ruta completa) | `deepseek/deepseek-r1` |

### Detección automática del prefijo `wall-vault/`

El prefijo propio de wall-vault determina automáticamente el servicio a partir del ID del modelo.

| Patrón de ID de modelo | Enrutamiento |
|------------------------|-------------|
| `wall-vault/gemini-*` | Google |
| `wall-vault/gpt-*`, `wall-vault/o1`, `wall-vault/o3`, `wall-vault/o4*` | OpenAI |
| `wall-vault/claude-*` | OpenRouter (ruta Anthropic) |
| `wall-vault/hunter-alpha`, `wall-vault/healer-alpha` | OpenRouter (gratuito, 1M ctx) |
| `wall-vault/kimi-*`, `wall-vault/glm-*`, `wall-vault/deepseek-*`, `wall-vault/qwen*` | OpenRouter |
| Otros | OpenRouter |

### Manejo del sufijo `:cloud`

El sufijo `:cloud` en formato de etiqueta Ollama se elimina automáticamente y se enruta a OpenRouter.

```
kimi-k2.5:cloud  →  OpenRouter, ID de modelo: kimi-k2.5
glm-5:cloud      →  OpenRouter, ID de modelo: glm-5
```

### Ejemplo de integración OpenClaw openclaw.json

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

Haga clic en el **botón 🐾** de una tarjeta de agente para copiar automáticamente el fragmento de configuración para ese agente al portapapeles.

---

## Esquema de datos

### APIKey

| Campo | Tipo | Descripción |
|-------|------|-------------|
| `id` | string | ID único en formato UUID |
| `service` | string | `google` \| `openai` \| `anthropic` \| `openrouter` \| `ollama` \| personalizado |
| `encrypted_key` | string | Clave cifrada AES-GCM (Base64) |
| `label` | string | Etiqueta de identificación |
| `today_usage` | int | Tokens de solicitudes exitosas hoy (no incluye errores 429/402/582) |
| `today_attempts` | int | Total de llamadas API hoy (éxito + limitadas por velocidad; se restablece a medianoche) |
| `daily_limit` | int | Límite diario (0 = ilimitado) |
| `cooldown_until` | time.Time | Fin del tiempo de enfriamiento |
| `last_error` | int | Último código de error HTTP |
| `created_at` | time.Time | Fecha de registro |

**Política de tiempo de enfriamiento:**

| Error HTTP | Tiempo de enfriamiento |
|------------|----------------------|
| 429 (Too Many Requests) | 30 minutos |
| 402 (Payment Required) | 24 horas |
| 400 / 401 / 403 | 24 horas |
| 582 (Gateway Overload) | 5 minutos |
| Error de red | 10 minutos |

> **429/402/582**: Se establece el tiempo de enfriamiento + `today_attempts` se incrementa. `today_usage` no cambia (solo se cuentan los tokens exitosos).
> **Ollama (servicio local)**: `callOllama` usa un cliente HTTP dedicado con `Timeout: 0` (ilimitado). La inferencia de modelos grandes puede tomar desde decenas de segundos hasta minutos, por lo que no se aplica el timeout predeterminado de 60 segundos.

### Client

| Campo | Tipo | Descripción |
|-------|------|-------------|
| `id` | string | ID único del cliente |
| `name` | string | Nombre para mostrar |
| `token` | string | Token de autenticación |
| `default_service` | string | Servicio predeterminado |
| `default_model` | string | Modelo predeterminado (puede estar en formato `provider/model`) |
| `allowed_services` | []string | Servicios permitidos (array vacío = todos) |
| `agent_type` | string | `openclaw` \| `claude-code` \| `cursor` \| `vscode` \| `custom` |
| `work_dir` | string | Directorio de trabajo del agente |
| `description` | string | Descripción |
| `ip_whitelist` | []string | Lista de IP permitidas (CIDR soportado) |
| `avatar` | string | Agent avatar — relative path under `~/.openclaw/` (e.g. `workspace/avatar.png`, `workspace/avatars/profile.hpg`) OR base64 data URI (`data:image/png;base64,...`). Supported extensions: `.png`, `.jpg`/`.jpeg`/`.hpg`, `.webp`, `.gif`. |
| `enabled` | bool | Si es `false`, devuelve `403` al acceder a `/api/keys` |
| `created_at` | time.Time | Fecha de registro |

### ServiceConfig

| Campo | Tipo | Descripción |
|-------|------|-------------|
| `id` | string | ID único del servicio |
| `name` | string | Nombre para mostrar |
| `local_url` | string | URL del servidor local (Ollama/LMStudio/vLLM/personalizado) |
| `enabled` | bool | Si está habilitado |
| `custom` | bool | Si es un servicio agregado por el usuario |
| `proxy_enabled` | bool | Whether this service is included in proxy dispatch and agent model dropdowns. Controlled by the "프록시 사용" checkbox in the Services card. Only `proxy_enabled: true` services appear in agent service/model selectors. |

### ProxyStatus (Heartbeat)

| Campo | Tipo | Descripción |
|-------|------|-------------|
| `client_id` | string | ID del cliente |
| `version` | string | Versión del proxy (ej. `v0.1.6.20260314.231308`) |
| `service` | string | Servicio actual |
| `model` | string | Modelo actual |
| `sse_connected` | bool | Si SSE está conectado |
| `host` | string | Nombre del host |
| `avatar` | string | base64 data URI of local avatar (auto-synced from proxy via heartbeat) |
| `updated_at` | time.Time | Última actualización |
| `vault.today_usage` | int | Uso de tokens hoy |
| `vault.daily_limit` | int | Límite diario |
| `vault.key_status` | string | `active` \| `cooldown` \| `exhausted` |

---

## Respuestas de error

```json
{"error": "오류 메시지"}
```

| Código | Significado |
|--------|-------------|
| 200 | Éxito |
| 400 | Solicitud incorrecta |
| 401 | Fallo de autenticación |
| 403 | Acceso denegado (cliente deshabilitado, IP bloqueada) |
| 404 | Recurso no encontrado |
| 405 | Método no permitido |
| 429 | Límite de velocidad excedido |
| 500 | Error interno del servidor |
| 502 | Error de API upstream (todos los respaldos fallaron) |

---

## Ejemplos cURL

```bash
# ─── Proxy ────────────────────────────────────────────────────────────────────

# Verificación de salud
curl http://localhost:56244/health

# Estado
curl http://localhost:56244/status

# Lista de modelos (todos)
curl http://localhost:56244/api/models

# Solo modelos de Google
curl "http://localhost:56244/api/models?service=google"

# Buscar modelos gratuitos
curl "http://localhost:56244/api/models?q=alpha"

# Cambiar modelo (local)
curl -X PUT http://localhost:56244/api/config/model \
  -H "Content-Type: application/json" \
  -d '{"service":"google","model":"gemini-2.5-pro"}'

# Recargar configuración
curl -X POST http://localhost:56244/reload

# Llamada directa a API Gemini
curl -X POST "http://localhost:56244/google/v1beta/models/gemini-2.5-flash:generateContent" \
  -H "Content-Type: application/json" \
  -d '{"contents":[{"role":"user","parts":[{"text":"안녕하세요"}]}]}'

# Compatible con OpenAI (modelo predeterminado)
curl -X POST http://localhost:56244/v1/chat/completions \
  -H "Content-Type: application/json" \
  -d '{"model":"gemini-2.5-flash","messages":[{"role":"user","content":"안녕"}]}'

# Formato OpenClaw provider/model
curl -X POST http://localhost:56244/v1/chat/completions \
  -H "Authorization: Bearer my-agent-token" \
  -H "Content-Type: application/json" \
  -d '{"model":"wall-vault/gemini-2.5-flash","messages":[{"role":"user","content":"안녕"}]}'

# Modelo gratuito de 1M de contexto
curl -X POST http://localhost:56244/v1/chat/completions \
  -H "Authorization: Bearer my-agent-token" \
  -H "Content-Type: application/json" \
  -d '{"model":"wall-vault/hunter-alpha","messages":[{"role":"user","content":"안녕"}]}'

# ─── Bóveda de claves (Público) ───────────────────────────────────────────────

curl http://localhost:56243/api/status
curl http://localhost:56243/api/clients
curl -s http://localhost:56243/api/events --max-time 3

# ─── Bóveda de claves (Admin) ────────────────────────────────────────────────

ADMIN="Authorization: Bearer admin-token"

# Lista de claves
curl -H "$ADMIN" http://localhost:56243/admin/keys

# Agregar clave de Google
curl -X POST http://localhost:56243/admin/keys \
  -H "$ADMIN" -H "Content-Type: application/json" \
  -d '{"service":"google","key":"AIzaSy...","label":"메인 키","daily_limit":1000}'

# Agregar clave de OpenAI
curl -X POST http://localhost:56243/admin/keys \
  -H "$ADMIN" -H "Content-Type: application/json" \
  -d '{"service":"openai","key":"sk-...","label":"GPT 키"}'

# Agregar clave de OpenRouter
curl -X POST http://localhost:56243/admin/keys \
  -H "$ADMIN" -H "Content-Type: application/json" \
  -d '{"service":"openrouter","key":"sk-or-v1-...","label":"OR 키"}'

# Eliminar clave (transmisión SSE key_deleted)
curl -X DELETE http://localhost:56243/admin/keys/key-abc123 -H "$ADMIN"

# Restablecer uso diario
curl -X POST http://localhost:56243/admin/keys/reset -H "$ADMIN"

# Lista de clientes
curl -H "$ADMIN" http://localhost:56243/admin/clients

# Agregar cliente (OpenClaw)
curl -X POST http://localhost:56243/admin/clients \
  -H "$ADMIN" -H "Content-Type: application/json" \
  -d '{"id":"bot-a","name":"봇 A","agent_type":"openclaw","work_dir":"~/.openclaw","default_service":"google","default_model":"gemini-2.5-flash"}'

# Cambiar modelo del cliente (actualización SSE instantánea)
curl -X PUT http://localhost:56243/admin/clients/bot-a \
  -H "$ADMIN" -H "Content-Type: application/json" \
  -d '{"default_service":"openrouter","default_model":"wall-vault/hunter-alpha"}'

# Deshabilitar cliente
curl -X PUT http://localhost:56243/admin/clients/my-bot \
  -H "$ADMIN" -H "Content-Type: application/json" \
  -d '{"enabled":false}'

# Eliminar cliente
curl -X DELETE http://localhost:56243/admin/clients/my-bot -H "$ADMIN"

# Lista de servicios
curl -H "$ADMIN" http://localhost:56243/admin/services

# Configurar URL local de Ollama (transmisión SSE service_changed)
curl -X PUT http://localhost:56243/admin/services/ollama \
  -H "$ADMIN" -H "Content-Type: application/json" \
  -d '{"local_url":"http://192.168.x.x:11434","enabled":true}'

# Habilitar servicio OpenAI
curl -X PUT http://localhost:56243/admin/services/openai \
  -H "$ADMIN" -H "Content-Type: application/json" \
  -d '{"enabled":true}'

# Agregar servicio personalizado (transmisión SSE service_changed)
curl -X POST http://localhost:56243/admin/services \
  -H "$ADMIN" -H "Content-Type: application/json" \
  -d '{"id":"my-llm","name":"사내 LLM","local_url":"http://10.0.0.50:8080","enabled":true}'

# Eliminar servicio personalizado
curl -X DELETE http://localhost:56243/admin/services/my-llm -H "$ADMIN"

# Lista de modelos
curl -H "$ADMIN" http://localhost:56243/admin/models
curl -H "$ADMIN" "http://localhost:56243/admin/models?service=openrouter"
curl -H "$ADMIN" "http://localhost:56243/admin/models?q=hunter"

# Estado del proxy (heartbeat)
curl -H "$ADMIN" http://localhost:56243/admin/proxies

# ─── Modo distribuido — Proxy → Bóveda ───────────────────────────────────────

# Obtener claves descifradas
curl http://localhost:56243/api/keys \
  -H "Authorization: Bearer your-bot-a-token"

# Enviar heartbeat
curl -X POST http://localhost:56243/api/heartbeat \
  -H "Authorization: Bearer your-bot-a-token" \
  -H "Content-Type: application/json" \
  -d '{"client_id":"bot-a","version":"v0.1.6.20260314.231308","service":"google","model":"gemini-2.5-flash","sse_connected":true}'
```

---

## Middleware

Aplicado automáticamente a todas las solicitudes:

| Middleware | Función |
|-----------|---------|
| **Logger** | Registra en formato `[method] path status latencyms` |
| **CORS** | `Access-Control-Allow-Origin: *` |
| **Recovery** | Recuperación de pánico, devuelve respuesta 500 |

---

*Última actualización: 16/03/2026 — v0.1.7: dashboard title renamed to "벽금고(wall-vault) 대시보드", logo moved to sticky topbar, today_attempts tracking, HTTP 582 cooldown (5min), share-of-total bar scaling, custom/ routing fix, Ollama Timeout:0, key_att i18n, avatar heartbeat sync, build timestamp versioning, proxy-only service filter*
