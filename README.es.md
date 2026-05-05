# wall-vault

> **Bóveda de claves API + proxy de IA en un solo binario Go.**
> Guarda las claves localmente con AES-GCM, las rota entre proveedores, recurre a otra cuando una falla y se entrega con un panel en tiempo real.

[![License: GPL v3](https://img.shields.io/badge/License-GPLv3-blue.svg)](LICENSE)
[![Go](https://img.shields.io/badge/Go-1.25-00ADD8.svg)](go.mod)

[English](README.md) · [한국어](README.ko.md) · [中文](README.zh.md) · [日本語](README.ja.md) · **Español** · [Français](README.fr.md) · [Deutsch](README.de.md) · [Português](README.pt.md) · [العربية](README.ar.md) · [हिन्दी](README.hi.md) · [Bahasa Indonesia](README.id.md) · [ภาษาไทย](README.th.md) · [Kiswahili](README.sw.md) · [Hausa](README.ha.md) · [नेपाली](README.ne.md) · [Монгол](README.mn.md) · [isiZulu](README.zu.md)

---

## Qué es

wall-vault se sitúa entre un agente de IA (OpenClaw, Claude Code, Cursor, Continue, tu propio script) y los proveedores de IA, en la nube o locales, con los que habla. Dos cosas en un mismo binario.

- **Vault** — almacena las claves API cifradas en reposo (AES-GCM con una contraseña maestra), las rota, registra el uso y los enfriamientos por clave, difunde los cambios por SSE y sirve un panel web en `:56243`.
- **Proxy** — expone endpoints compatibles con Gemini, Anthropic y OpenAI en `:56244`, escoge una clave de la bóveda, despacha al upstream que has configurado y, cuando uno falla, recurre al siguiente proveedor.

Soporta cuatro formas de petición (Gemini `:generateContent`, Anthropic `/v1/messages`, OpenAI `/v1/chat/completions` y la nativa de Ollama `/api/chat`) y cinco categorías de upstream.

| Proveedor | Notas |
|----------|-------|
| Google Gemini | API nativa; rotación de claves por proyecto |
| Anthropic | Passthrough nativo de `/v1/messages` |
| OpenAI | Nativo `/v1/chat/completions` |
| OpenRouter | 340+ modelos, fallback automático a las variantes `:free` |
| Ollama / LM Studio / vLLM / llama.cpp / Jan / KoboldCpp / TabbyAPI / mlx-server / LiteLLM | Backends locales compatibles con OpenAI; integración directa mediante un yaml de plugin |

Añadir un nuevo backend compatible con OpenAI es un único yaml en `~/.wall-vault/services/`, sin cambios de código.

## Por qué podrías quererlo

- Estás haciendo malabares con tres o cuatro servicios de IA y quieres que el agente hable con una sola URL.
- Quieres que una clave de plan gratuito en enfriamiento ceda el sitio a la siguiente sin romper la sesión.
- Quieres que las mismas claves alimenten varios bots / IDE / scripts en la misma LAN sin copiar credenciales.
- Quieres un panel, no variables de entorno, para editar tus claves API.
- Quieres una opción local-first (Ollama / LM Studio) cuando se agotan los límites de la nube.

## Inicio rápido

### Instalación (Linux / macOS)

```bash
curl -fsSL https://raw.githubusercontent.com/sookmook/wall-vault/main/install.sh | sh
```

O descarga directamente un binario precompilado.

```bash
# Linux amd64
curl -L https://github.com/sookmook/wall-vault/releases/latest/download/wall-vault-linux-amd64 \
  -o wall-vault && chmod +x wall-vault

# macOS Apple Silicon
curl -L https://github.com/sookmook/wall-vault/releases/latest/download/wall-vault-darwin-arm64 \
  -o wall-vault && chmod +x wall-vault

# Linux arm64 (Raspberry Pi, servidores ARM)
curl -L https://github.com/sookmook/wall-vault/releases/latest/download/wall-vault-linux-arm64 \
  -o wall-vault && chmod +x wall-vault
```

### Instalación (Windows)

```powershell
Invoke-WebRequest -Uri `
  "https://github.com/sookmook/wall-vault/releases/latest/download/wall-vault-windows-amd64.exe" `
  -OutFile wall-vault.exe
```

### Primera ejecución

```bash
wall-vault setup    # asistente interactivo: elige puerto, servicios, admin token, master password
wall-vault start    # arranca vault y proxy
```

Abre `http://localhost:56243` (o `https://...` cuando actives TLS, ver más abajo) en un navegador. El panel pide el admin token impreso por `setup`. Desde allí añades claves API, registras clientes y cambias de modelo sin reiniciar.

---

## TLS (recomendado)

Por defecto, `wall-vault setup` escribe una configuración sin TLS, así que ambos listeners responden en HTTP plano. Las URL de ejemplo de este README usan `https://localhost:56244` porque la mayoría de los agentes (OpenClaw, Claude Code, Cursor) prefieren un único endpoint con TLS por delante que no se rompa si más adelante mueves el proxy a otra máquina. Para encajar con esos ejemplos, activa TLS una vez con la CA interna que viene incluida.

```bash
# 1. Crea la CA interna de wall-vault (una sola vez, vive en ~/.wall-vault/ca.{crt,key})
wall-vault cert init

# 2. Emite un certificado de host para ESTA máquina
#    Los SAN incluyen el hostname, localhost, 127.0.0.1 y cualquier IP de LAN detectada
wall-vault cert issue $(hostname)

# 3. Confía en la CA en el llavero del SO local
wall-vault cert install-trust

# 4. Cambia los listeners a TLS
export WV_PROXY_TLS_ENABLED=1
export WV_PROXY_TLS_CERT="$HOME/.wall-vault/$(hostname).crt"
export WV_PROXY_TLS_KEY="$HOME/.wall-vault/$(hostname).key"
export WV_VAULT_TLS_ENABLED=1
export WV_VAULT_TLS_CERT="$HOME/.wall-vault/$(hostname).crt"
export WV_VAULT_TLS_KEY="$HOME/.wall-vault/$(hostname).key"
wall-vault start
```

Para otra máquina en tu LAN, copia `~/.wall-vault/ca.crt` y ejecuta allí `wall-vault cert install-trust --ca <path>`. Una vez la CA esté confiada en todas partes, cualquier máquina de la red podrá llegar al proxy por `https://<host>:56244` sin avisos de certificado.

Si prefieres seguir con HTTP en plano, deja la configuración como está y sustituye `https://` por `http://` en los fragmentos de cliente más abajo. Ambos esquemas funcionan; la única diferencia es qué puerto responde a un handshake TLS.

**Fallback de loopback.** Los clientes en el mismo host que no pueden honrar la CA de wall-vault (notablemente el runtime Node empaquetado con OpenClaw, que reescribe `NODE_EXTRA_CA_CERTS` al arrancar) llegan al proxy a través de un compañero HTTP en plano, solo loopback, en `127.0.0.1:56245`. wall-vault lo activa automáticamente cuando TLS está encendido.

---

## Conectar clientes

Apunta cualquier cliente de IA a `https://<host>:56244` (o `http://...` si TLS está apagado). El proxy responde a cuatro formas.

| Formato | Ruta | Clientes de ejemplo |
|--------|------|-----------------|
| Gemini | `/google/v1beta/models/<model>:generateContent` | OpenClaw, Gemini CLI, Antigravity |
| Anthropic | `/v1/messages` | Claude Code, SDK de Anthropic |
| OpenAI | `/v1/chat/completions` | Cursor, Continue, scripts propios, la mayoría de apps LLM |
| Nativo Ollama | `/api/chat` | Clientes Ollama en passthrough |

### Claude Code

```bash
export ANTHROPIC_BASE_URL=https://localhost:56244
export ANTHROPIC_API_KEY=<your-vault-client-token>
claude
```

Cuando se acaben los créditos del Anthropic upstream, el despacho recurre a los proveedores que hayas indicado en `fallback_services` para este cliente. Para activar explícitamente un fallback fuera de Claude.

```yaml
proxy:
  anthropic_fallback_model: "claude-haiku-4-5-20251001"
```

(El valor por defecto vacío hace que el despacho devuelva un error, así un mal enrutado sale a la luz al instante.)

### Cursor / Continue

En Cursor, **Settings → AI → OpenAI API**:

```
Base URL:  https://localhost:56244
API Key:   <your-vault-client-token>
Model:     gemini-2.5-flash    # o cualquier modelo que wall-vault conozca
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
      "apiKey": "<your-vault-client-token>"
    }
  ]
}
```

### OpenClaw

OpenClaw es un framework de agentes TUI al que wall-vault servía originalmente. La modal **Add Agent** del panel pone el tipo de agente en `openclaw` (o `nanoclaw`); wall-vault entonces escribe `~/.openclaw/openclaw.json` directamente, incluyendo URL de proveedores, el token de la bóveda y las entradas de modelos.

```bash
VAULT_CLIENT_ID=my-bot \
VAULT_URL=http://<vault-host>:56243 \
VAULT_TOKEN=<your-client-token> \
wall-vault proxy
```

### curl / scripts

```bash
curl -X POST https://localhost:56244/v1/chat/completions \
  -H "Authorization: Bearer <your-vault-client-token>" \
  -H "Content-Type: application/json" \
  -d '{
    "model": "gemini-2.5-flash",
    "messages": [{"role": "user", "content": "ping"}]
  }'
```

---

## Configuración

`wall-vault setup` escribe `./wall-vault.yaml` o `~/.wall-vault/config.yaml`. Edita a mano los campos que el asistente no pregunta.

```yaml
mode: standalone     # standalone | distributed
lang: en             # en | ko | zh | ja | es | fr | de | pt | ar | hi | id | th | sw | ha | ne | mn | zu
theme: light         # light | dark | cherry | ocean | gold | autumn | winter

proxy:
  port: 56244
  host: ""           # por defecto: 127.0.0.1 standalone, 0.0.0.0 distributed
  client_id: my-bot
  vault_url: ""      # distributed: http://vault-host:56243
  vault_token: ""    # distributed: token de cliente
  tool_filter: strip_all   # strip_all | whitelist | passthrough
  allowed_tools: []
  services: [google, openrouter, ollama]
  timeout: 300s
  tls:
    enabled: false
    cert_file: ""
    key_file: ""
  plain_port: 56245              # compañero HTTP solo loopback cuando TLS está activo
  ollama_keep_alive: "30m"       # "-1" nunca descarga, "0" descarga inmediato
  ollama_num_ctx: 8192
  oai_stream_forward: false      # passthrough opt-in del SSE real del backend
  anthropic_fallback_model: ""   # opt-in para reescribir a no-Claude en despacho anthropic

vault:
  port: 56243
  host: ""
  admin_token: ""
  admin_ip_whitelist: []
  master_password: ""            # contraseña de cifrado AES-GCM
  data_dir: ~/.wall-vault/data
  services_dir: ~/.wall-vault/services
  tls:
    enabled: false
    cert_file: ""
    key_file: ""
  bootstrap_port: 56247          # listener HTTP plano que solo sirve ca.crt

doctor:
  interval: 5m
  auto_fix: true
  log_file: /tmp/wall-vault-doctor.log

hooks:
  on_model_change: ""            # comando shell (env: SERVICE, MODEL)
  on_key_exhausted: ""
  on_service_down: ""
  on_doctor_fix: ""
  openclaw_socket: ""
```

### Variables de entorno

Cada campo YAML tiene una variable de entorno que lo sobreescribe y gana sobre el archivo. Las habituales.

| Variable | Descripción |
|----------|-------------|
| `WV_LANG`, `WV_THEME` | Idioma y tema |
| `WV_PROXY_PORT`, `WV_PROXY_HOST` | Dirección de escucha del proxy |
| `WV_VAULT_PORT`, `WV_VAULT_HOST` | Dirección de escucha de la bóveda |
| `WV_VAULT_URL`, `WV_VAULT_TOKEN` | Endpoints de modo distributed |
| `WV_ADMIN_TOKEN`, `WV_MASTER_PASS` | Credenciales de la bóveda |
| `WV_KEY_GOOGLE`, `WV_KEY_OPENROUTER`, `WV_KEY_ANTHROPIC`, `WV_KEY_OPENAI` | Claves API (separadas por coma para varias) |
| `WV_PROXY_TLS_ENABLED`, `WV_PROXY_TLS_CERT`, `WV_PROXY_TLS_KEY` | TLS del proxy |
| `WV_VAULT_TLS_ENABLED`, `WV_VAULT_TLS_CERT`, `WV_VAULT_TLS_KEY` | TLS de la bóveda |
| `WV_PROXY_PLAIN_PORT` | Compañero HTTP loopback (`0` para deshabilitar) |
| `WV_VAULT_BOOTSTRAP_PORT` | Listener de bootstrap de CA (`0` para deshabilitar) |
| `WV_OLLAMA_URL`, `WV_OLLAMA_KEEP_ALIVE`, `WV_OLLAMA_NUM_CTX` | Ajustes de Ollama |
| `WV_LMSTUDIO_URL`, `WV_VLLM_URL`, `WV_LLAMACPP_URL` | Sustituciones de backends locales |
| `WV_TOKEN_SENTINEL_FALLBACK` | Sustitución del centinela "proxy-managed" en loopback |
| `WV_OAI_STREAM_FORWARD` | Passthrough del SSE real del backend en OpenAI-compat |
| `WV_ANTHROPIC_FALLBACK_MODEL` | Opt-in de reescritura a no-Claude en anthropic |

---

## Modos

### Standalone (por defecto)

Bóveda y proxy corren en el mismo proceso. Lo ideal para un único host que aloja claves y agente. Por defecto solo escucha en loopback.

```bash
wall-vault start    # corre ambos
```

### Distributed

La bóveda corre en un host (el **vault host**) y guarda todas las claves; varios proxies en otros hosts se autentican cada uno con un token por cliente. Útil cuando varias máquinas necesitan las mismas claves sin tener que copiarlas por todas partes.

**Vault host:**

```bash
WV_VAULT_HOST=0.0.0.0 wall-vault vault
```

**Cada proxy host:**

```bash
WV_VAULT_URL=http://<vault-host>:56243 \
WV_VAULT_TOKEN=<this-client-token> \
WV_PROXY_HOST=0.0.0.0 \
wall-vault proxy
```

La modal **Add Client** del panel acuña un token, registra un tipo de agente y el proxy recoge su configuración por SSE sin reiniciarse.

---

## Plugin yaml (backend drop-in)

Cualquier backend compatible con OpenAI puede añadirse como un yaml en `~/.wall-vault/services/`. wall-vault lo recoge al arrancar, lo registra como servicio enrutable, y el despacho, el conjunto de detección de OAI-compat y el puente Gemini-stream lo ven todos sin tocar código.

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
inline_no_think_for_qwen3: false   # actívalo si tu backend elimina el marcador
```

La topología hub (un wall-vault delante de otro) se soporta vía `tls_internal_ca: true`, `auth.type: bearer` y `preserve_model_id: true`.

---

## Compilar desde el código

```bash
git clone https://github.com/sookmook/wall-vault
cd wall-vault
go build -o wall-vault .
./wall-vault setup
./wall-vault start
```

Compilar de forma cruzada para todo el set soportado.

```bash
make build-all   # linux-amd64, linux-arm64, darwin-amd64, darwin-arm64, windows-amd64
```

Las versiones siguen `v{major}.{minor}.{patch}.{YYYYMMDD}.{HHmmss}`; `BASE_VERSION` en el Makefile fija el prefijo.

### Estructura del proyecto

```
wall-vault/
├── main.go                     # despacho CLI (start/proxy/vault/setup/cert/doctor)
├── cmd/
│   ├── setup/                  # asistente interactivo de instalación
│   └── cert/                   # CA interna + emisor de certificados TLS por host
├── internal/
│   ├── config/                 # cargador YAML + env, cargador de plugins
│   ├── proxy/                  # despacho de peticiones, rotación de claves, conversores de formato
│   ├── vault/                  # almacén AES-GCM, panel, broker SSE
│   ├── doctor/                 # sonda de salud + auto-fix
│   ├── hooks/                  # disparadores de eventos vía comando shell
│   └── i18n/                   # cadenas UI en 17 idiomas
├── configs/services/           # yamls de plugin incluidos (lmstudio, vllm, ollama, …)
└── docs/                       # MANUAL, referencia API, 16 variantes de idioma
```

---

## Documentación

- [Manual de usuario](docs/MANUAL.en.md) — instalación, panel, agentes, resolución de problemas
- [Referencia de la API](docs/API.en.md) — cada endpoint con sus formas de petición/respuesta
- [CHANGELOG](CHANGELOG.md)

---

## Stack técnico

- Go 1.25, un único binario estático
- [templ](https://templ.guide) para el panel renderizado en servidor, [HTMX](https://htmx.org) para actualizaciones parciales
- AES-GCM (clave derivada con PBKDF2) para el cifrado de claves en reposo
- Server-Sent Events para sincronización en vivo de la configuración entre la bóveda y los proxies
- CA interna autofirmada + certificados por host (sin DNS público / Let's Encrypt requerido)

## Licencia

GPL-3.0. Ver [LICENSE](LICENSE).

## Contribuir

Pull requests bienvenidas. Ver [CONTRIBUTING.md](CONTRIBUTING.md). Para cambios mayores, por favor abre primero un issue para discutir el diseño.
