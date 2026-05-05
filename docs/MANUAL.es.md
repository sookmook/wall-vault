# Manual de usuario de wall-vault

[English](MANUAL.md) · [한국어](MANUAL.ko.md) · [中文](MANUAL.zh.md) · [日本語](MANUAL.ja.md) · **Español** · [Français](MANUAL.fr.md) · [Deutsch](MANUAL.de.md) · [Português](MANUAL.pt.md) · [العربية](MANUAL.ar.md) · [हिन्दी](MANUAL.hi.md) · [Bahasa Indonesia](MANUAL.id.md) · [ภาษาไทย](MANUAL.th.md) · [Kiswahili](MANUAL.sw.md) · [Hausa](MANUAL.ha.md) · [नेपाली](MANUAL.ne.md) · [Монгол](MANUAL.mn.md) · [isiZulu](MANUAL.zu.md)

Este manual cubre la instalación, configuración y operación de wall-vault. Para una visión general rápida, consulta el [README](../README.md). Para detalles de la API HTTP, consulta la [referencia de API](API.md).

## Contenido

1. [Qué hace wall-vault](#qué-hace-wall-vault)
2. [Instalación](#instalación)
3. [Primer arranque con el asistente de configuración](#primer-arranque-con-el-asistente-de-configuración)
4. [Habilitar TLS](#habilitar-tls)
5. [Registrar claves de API](#registrar-claves-de-api)
6. [Conectar agentes](#conectar-agentes)
7. [El panel](#el-panel)
8. [Modo distribuido](#modo-distribuido)
9. [Inicio automático](#inicio-automático)
10. [Plugins yaml](#plugins-yaml)
11. [Doctor](#doctor)
12. [Hooks](#hooks)
13. [Variables de entorno](#variables-de-entorno)
14. [Solución de problemas](#solución-de-problemas)

---

## Qué hace wall-vault

wall-vault es un único binario Go que agrupa dos servicios cooperativos:

- **El vault** almacena claves de API cifradas en reposo (AES-GCM con una contraseña maestra), rastrea el uso y los enfriamientos por clave, transmite los cambios mediante Server-Sent Events (SSE) y sirve un panel web en `:56243` para los operadores humanos.
- **El proxy** expone endpoints Gemini, Anthropic, compatibles con OpenAI y nativos de Ollama en `:56244`. Cualquier cliente de IA que apunte al proxy usa las claves del vault — los clientes nunca las ven. Cuando un upstream falla, el dispatch hace fallback al siguiente proveedor en orden.

Esto es útil cuando:

- Tienes claves para varios proveedores y quieres una única URL con la que el agente se comunique.
- Quieres que una clave del nivel gratuito que está en enfriamiento se aparte sin interrumpir la sesión.
- Quieres que las mismas claves alimenten varios bots, IDEs o scripts en la misma LAN sin copiar credenciales.
- Quieres un panel, no variables de entorno, para editar claves y cambiar de modelo.
- Quieres un fallback local (Ollama, LM Studio, vLLM) cuando se agoten los límites de la nube.

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

## Instalación

### Línea única para Linux / macOS

```bash
curl -fsSL https://raw.githubusercontent.com/sookmook/wall-vault/main/install.sh | sh
```

El script detecta automáticamente el sistema operativo y la arquitectura, descarga el binario adecuado en `~/.local/bin/wall-vault` y lo hace ejecutable. Si `~/.local/bin` no está en tu `PATH`, añádelo:

```bash
export PATH="$HOME/.local/bin:$PATH"
```

### Descarga manual

Los binarios precompilados se publican en cada release en `https://github.com/sookmook/wall-vault/releases`.

```bash
# Linux amd64
curl -L https://github.com/sookmook/wall-vault/releases/latest/download/wall-vault-linux-amd64 \
  -o wall-vault && chmod +x wall-vault

# Linux arm64 (Raspberry Pi, servidores ARM)
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

### Compilar desde el código fuente

Requiere Go 1.25 o más reciente.

```bash
git clone https://github.com/sookmook/wall-vault
cd wall-vault
go build -o wall-vault .
```

`make build-all` compila de forma cruzada para las cinco plataformas soportadas. Los binarios se generan en `bin/`.

---

## Primer arranque con el asistente de configuración

```bash
wall-vault setup
```

El asistente te pide en orden:

1. **Idioma** — elige uno de los 17 locales de UI. Se detecta automáticamente desde `$LANG`; el asistente ofrece la lista igualmente.
2. **Tema** — `light` (predeterminado), `dark`, `cherry`, `ocean`, `gold`, `autumn`, `winter`. Solo cosmético.
3. **Modo** — `standalone` (un solo host, predeterminado) o `distributed` (vault en un host, proxies en otros).
4. **Nombre del bot** — un slug `client_id` de forma libre. El vault lo usa para acotar la configuración por cliente (overrides de modelo, cadenas de fallback).
5. **Puerto del proxy** — predeterminado `56244`.
6. **Puerto del vault** — predeterminado `56243` (solo standalone).
7. **Selección de servicios** — un y/N para cada uno de: Google Gemini, OpenRouter, Anthropic, OpenAI, Ollama, LM Studio, vLLM. Múltiples opciones son válidas; cada una escribe su pista de variable de entorno al final.
8. **Filtro de herramientas** — `strip_all` (predeterminado; bloquea todas las definiciones de herramientas entrantes por seguridad) o `passthrough` (deja pasar cualquier herramienta).
9. **Token de admin** — déjalo en blanco para autogenerarlo. El panel requiere este token para iniciar sesión.
10. **Contraseña maestra** — déjala en blanco para no usar cifrado (NO recomendado); establece un valor para cifrar el almacén de claves en reposo con AES-GCM.
11. **Ruta de guardado** — predeterminado `wall-vault.yaml` en el directorio actual. El cargador también busca en `~/.wall-vault/config.yaml`.

Después de guardar, el asistente ejecuta `doctor.FixTrust` para que cualquier agente instalado localmente (OpenClaw, Claude Code, Cline) reciba automáticamente el CA interno de wall-vault añadido a su almacén de confianza. Si no hay ningún agente de ese tipo instalado, el paso imprime `SKIP` y no escribe nada.

Luego inicia el binario:

```bash
wall-vault start
```

`start` ejecuta tanto el vault como el proxy en un solo proceso (modo standalone). Para el modo distributed, usa `wall-vault vault` en el host del vault y `wall-vault proxy` en cada host de proxy.

Abre `http://localhost:56243` en un navegador. Inicia sesión con el token de admin que el asistente imprimió.

---

## Habilitar TLS

Los valores predeterminados del asistente dejan ambos listeners en HTTP plano. La mayoría de los agentes (OpenClaw, Claude Code, Cursor) funcionan mejor contra un único endpoint HTTPS, por lo que se recomienda TLS en cualquier despliegue que abarque más allá de la máquina local.

wall-vault viene con su propio CA interno, por lo que no necesitas un nombre DNS público ni Let's Encrypt.

```bash
# 1. Crear el CA interno — escrito en ~/.wall-vault/ca.{crt,key}.
#    El CA es válido por 10 años por defecto; sobreescribir con --ca-years.
wall-vault cert init

# 2. Emitir un certificado de host. Los Subject Alternative Names incluyen automáticamente:
#       hostname, "localhost", "127.0.0.1" y cualquier IP LAN no de loopback detectada.
#    Sobreescribir el directorio del emisor con --dir, la validez con --host-years.
wall-vault cert issue $(hostname)

# 3. Confiar en el CA en el llavero del SO de esta máquina.
#    Linux: escribe en /etc/ssl/certs/ vía update-ca-certificates (necesita sudo).
#    macOS: añade al llavero System vía security add-trusted-cert (necesita sudo).
#    Windows: importa en CurrentUser\Root vía certutil (no necesita admin).
wall-vault cert install-trust

# 4. Habilitar TLS en ambos listeners.
export WV_PROXY_TLS_ENABLED=1
export WV_PROXY_TLS_CERT="$HOME/.wall-vault/$(hostname).crt"
export WV_PROXY_TLS_KEY="$HOME/.wall-vault/$(hostname).key"
export WV_VAULT_TLS_ENABLED=1
export WV_VAULT_TLS_CERT="$HOME/.wall-vault/$(hostname).crt"
export WV_VAULT_TLS_KEY="$HOME/.wall-vault/$(hostname).key"

wall-vault start
```

Para extender la confianza a otras máquinas LAN, copia `~/.wall-vault/ca.crt` y ejecuta `wall-vault cert install-trust --ca <path>` en cada una. El vault también expone `ca.crt` mediante un pequeño listener HTTP plano en `:56247` (el **puerto bootstrap**) para el caso en que un cliente nuevo necesita el CA para hablar HTTPS.

### Compañero HTTP de loopback

Algunos agentes — en particular el runtime Node empaquetado con OpenClaw — reescriben `NODE_EXTRA_CA_CERTS` en el spawn del proceso, descartando cualquier pista de CA proporcionada por el operador. No pueden honrar el CA de wall-vault desde dentro del daemon, ni siquiera después de `cert install-trust`. wall-vault soluciona esto vinculando un **listener HTTP plano de solo loopback** adicional en `127.0.0.1:56245` siempre que TLS esté habilitado. Los clientes en el mismo host alcanzan el proxy a través de ese puerto sin TLS en absoluto; los clientes LAN siguen usando el listener TLS.

Desactiva con `WV_PROXY_PLAIN_PORT=0` si no lo necesitas.

### `wall-vault cert list`

Muestra todos los certificados bajo `~/.wall-vault/` con subject, ventana de validez y SANs.

```
$ wall-vault cert list
ca.crt          subject=wall-vault internal CA   not-after=2036-05-05
hostname.crt    subject=hostname                 not-after=2031-05-05   SAN=hostname,localhost,127.0.0.1,192.168.…
```

---

## Registrar claves de API

Dos formas: el panel o las variables de entorno.

### Panel (recomendado)

1. Inicia sesión en `https://localhost:56243` con el token de admin.
2. Haz clic en **+ API key** en la tarjeta de claves.
3. Elige un servicio (Google, OpenRouter, Anthropic, OpenAI, …).
4. Pega la clave. Guardar.

Múltiples claves por servicio están bien; el proxy hace round-robin entre ellas y omite las que están en enfriamiento por clave.

### Variables de entorno (bootstrap único)

```bash
export WV_KEY_GOOGLE="AIzaSyA1...,AIzaSyB2...,AIzaSyC3..."   # separadas por coma
export WV_KEY_OPENROUTER="sk-or-v1-…"
export WV_KEY_ANTHROPIC="sk-ant-…"
export WV_KEY_OPENAI="sk-…"
wall-vault start
```

Las claves proporcionadas de esta forma se escriben en el almacén cifrado en el primer arranque. Los inicios subsiguientes las leen del disco; puedes desestablecer las variables de entorno tras la primera ejecución.

### Enfriamientos y rotación

Cada llamada exitosa incrementa el `usage_count` de la clave y refresca `last_used`. En HTTP 429 / 402 / 403, el proxy pone la clave en **enfriamiento** (predeterminado: 60 minutos para 429, 24 horas para 402, 12 horas para 403). El siguiente dispatch elige una clave diferente para ese servicio. Cuando todas las claves de un servicio están en enfriamiento, el proxy salta rápidamente ese servicio por completo y prueba el siguiente proveedor en la cadena de fallback.

Los enfriamientos son visibles por clave en el panel con una cuenta atrás.

---

## Conectar agentes

### OpenClaw

OpenClaw es el cliente objetivo original. Usa el modal **+ Add agent** del panel:

- Establece **Agent type** a `openclaw` o `nanoclaw`.
- Establece **Work directory** — para OpenClaw se autocompleta como `~/.openclaw`.
- Elige un **preferred service** y opcionalmente un **model override**.
- Haz clic en **Apply**. wall-vault escribe directamente en `~/.openclaw/openclaw.json` (URLs de proveedores, token del vault, entradas de modelo).

Cuando cambias el modelo desde el panel, OpenClaw recoge el cambio mediante SSE en 1 a 3 segundos — sin reinicio.

### Claude Code

```bash
export ANTHROPIC_BASE_URL=https://localhost:56244
export ANTHROPIC_API_KEY=<your-vault-client-token>
claude
```

Cuando se agotan los créditos upstream de Anthropic, el dispatch hace fallback a los servicios listados en `fallback_services` de este cliente. Por defecto, un model id no-Claude enviado al dispatch anthropic devuelve un error para que el enrutamiento incorrecto salga inmediatamente a la superficie. Activa la reescritura automática:

```yaml
proxy:
  anthropic_fallback_model: "claude-haiku-4-5-20251001"
```

### Cursor

En **Settings → AI → OpenAI API** de Cursor:

```
Base URL:  https://localhost:56244
API Key:   <your-vault-client-token>
Model:     gemini-2.5-flash    # o cualquier modelo que wall-vault conozca
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

### HTTP personalizado

```bash
curl -X POST https://localhost:56244/v1/chat/completions \
  -H "Authorization: Bearer <your-vault-client-token>" \
  -H "Content-Type: application/json" \
  -d '{
    "model": "gemini-2.5-flash",
    "messages": [{"role": "user", "content": "hello"}]
  }'
```

El mismo endpoint acepta streaming (`"stream": true`) cuando `proxy.oai_stream_forward: true` está establecido.

---

## El panel

`https://localhost:56243`. Cinco tarjetas en la cuadrícula de inicio:

- **Keys** — todas las claves de API, agrupadas por servicio. Añadir, editar, eliminar; ver uso y enfriamiento.
- **Services** — Google / OpenRouter / Anthropic / OpenAI / Ollama / LM Studio / vLLM / llama.cpp, además de cualquier yaml de plugin en `~/.wall-vault/services/`. Establece por servicio `default_model`, `allowed_models`, URL base, toggle de reasoning.
- **Clients (agents)** — cada cliente registrado (bot OpenClaw, sesión Claude Code, instancia Cursor, …). Asigna servicio preferido, override de modelo, cadena de fallback.
- **Proxies** — cada proxy que se ha autenticado contra este vault. Estado en vivo (online/offline), última vez visto, modelo actual.
- **Settings** — token de admin, rotación de contraseña maestra, tema, idioma.

Cada tarjeta tiene un slideover de edición (lado derecho). Clic fuera o `Esc` lo cierra. Los cambios se envían a todos los proxies conectados mediante SSE en segundos.

El **footer** lleva un indicador SSE (verde = conectado, naranja = reconectando, gris = desconectado) y la versión de build en vivo.

---

## Modo distribuido

Cuando tienes varias máquinas que necesitan las mismas claves, ejecuta el vault en un host y proxies en cada uno de los demás.

### Host del vault

```bash
WV_VAULT_HOST=0.0.0.0 \
WV_ADMIN_TOKEN=<admin> \
WV_MASTER_PASS=<master> \
wall-vault vault
```

El panel ahora es alcanzable en `https://<vault-host>:56243`. Añade un agente para cada proxy remoto en la tarjeta **Clients**; cada uno acuña un `vault_token` único.

### Hosts de proxy

```bash
WV_VAULT_URL=http://<vault-host>:56243 \
WV_VAULT_TOKEN=<that-client-token> \
WV_PROXY_HOST=0.0.0.0 \
wall-vault proxy
```

El proxy se autentica contra el vault, abre un stream SSE y aplica cualquier configuración que reciba (servicio preferido, override de modelo, cadena de fallback). Las ediciones subsiguientes del vault llegan en segundos sin reinicio.

Para instalaciones que abarcan toda la LAN, habilita TLS en el host del vault (`WV_VAULT_TLS_ENABLED=1` + las variables de entorno de cert/key) y ejecuta cada host de proxy a través del mismo paso `wall-vault cert install-trust` para que las llamadas HTTPS del proxy al vault sean confiables.

---

## Inicio automático

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
loginctl enable-linger $USER       # para que la unidad siga corriendo tras logout
```

Para el vault en el mismo host, escribe un `wall-vault-vault.service` paralelo. Para el modo standalone, una unidad que llame a `wall-vault start` es suficiente.

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

Usa `nssm` para envolver `wall-vault.exe start` como un servicio de Windows, o una entrada `schtasks` que se ejecute al inicio de sesión del usuario.

---

## Plugins yaml

Cualquier backend compatible con OpenAI puede añadirse sin cambios de código simplemente colocando un yaml en `~/.wall-vault/services/`. wall-vault lo carga al inicio y registra el servicio para el dispatch, el conjunto de detección OAI-compat y el bridge de stream Gemini.

```yaml
# ~/.wall-vault/services/llamacpp.yaml
id: llamacpp                 # id de servicio único
name: llama.cpp              # etiqueta legible
enabled: true                # los plugins desactivados se omiten al cargar

default_url: http://localhost:8080   # override del operador; env gana (WV_LLAMACPP_URL)
endpoints:
  generate: /v1/chat/completions
  list_models: /v1/models

auth:
  type: none                 # none | bearer | query_param | header
  param: ""                  # para query_param: el nombre del parámetro (p. ej. "key")

request_format: openai       # openai | gemini | ollama | raw

model_fetch:
  enabled: true              # deja que el panel detecte modelos automáticamente
  dynamic: true              # re-fetch en cada apertura del panel
  auto_detect_url: true      # prueba /v1/models incluso si no se declara

concurrency:
  max: 1                     # peticiones concurrentes máximas a este backend
  queue_size: 10
  wait_notify: true          # muestra hint "queued" a agentes TUI

error_codes:
  503:
    cooldown: 5m
    message: "llama.cpp not responding"

# Opt-in para la directiva inline /no_think de la familia qwen3 cuando reasoning está apagado.
# Establecer true si la plantilla de chat de tu backend elimina el marcador (jinja
# de LM Studio, capa /v1 de Ollama). Otros backends suelen devolver el texto literal,
# por lo que esto se mantiene como opt-in por yaml.
inline_no_think_for_qwen3: false

# Topología Hub — apunta a otro wall-vault. Requerido cuando este plugin
# está al frente de un wall-vault remoto (para que el wall-vault receptor
# vea el prefijo del publicador y enrute correctamente) y para que el
# bearer token en proxy.vault_token se envíe como Authorization.
preserve_model_id: false
tls_internal_ca: false       # añade ~/.wall-vault/ca.crt al pool de confianza del cliente
```

El conjunto empaquetado en `configs/services/` (lmstudio, vllm, llamacpp, tgwui, localai, jan, koboldcpp, tabbyapi, mlx-server, litellm-proxy, ollama, google, openrouter) viene desactivado por defecto. Copia el que quieras a `~/.wall-vault/services/`, establece `enabled: true`, reinicia.

---

## Doctor

`wall-vault doctor` ejecuta una sonda de salud única en toda la instalación:

```
✓ vault listener  (https://localhost:56243)
✓ proxy listener  (https://localhost:56244)
✓ master password set
⚠ Google: 2 keys, all on cooldown
✓ Anthropic: 1 key healthy
✗ Ollama: not reachable at http://localhost:11434
```

Cada línea es uno de:

- `✓` — saludable
- `⚠` — degradado pero funcional (una clave en enfriamiento, baja cuota, etc.)
- `✗` — roto
- `SKIP` — no configurado / no aplicable en este host

Un segundo modo daemon ejecuta la misma sonda cada `doctor.interval` (predeterminado 5 minutos) y escribe los resultados en `doctor.log_file` (predeterminado `/tmp/wall-vault-doctor.log`). Cuando `doctor.auto_fix` es true, también intenta reparar las desviaciones comunes (configuración OpenClaw obsoleta, falta de confianza TLS, servicios reiniciables).

Activa una sonda única desde el panel mediante la tarjeta **Doctor** o `wall-vault doctor`.

---

## Hooks

Ejecuta un comando shell en eventos clave:

```yaml
hooks:
  on_model_change:   "logger 'wall-vault: $SERVICE/$MODEL'"
  on_key_exhausted:  "notify-send 'wall-vault' '$SERVICE keys all on cooldown'"
  on_service_down:   "/usr/local/bin/page-oncall.sh $SERVICE '$ERROR'"
  on_doctor_fix:     "echo \"$AGENT: $LEVEL $MSG\" >> ~/wall-vault.audit.log"
  openclaw_socket:   ""    # si está establecido, OpenClaw TUI recibe eventos por este socket Unix
```

Cada hook recibe variables de entorno específicas del evento (`SERVICE`, `MODEL`, `ERROR`, `AGENT`, `LEVEL`, `MSG`). Los hooks se ejecutan async con un timeout de 5 segundos — el proxy nunca bloquea por un hook lento.

---

## Variables de entorno

| Variable | Campo YAML |
|----------|------------|
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
| `WV_KEY_GOOGLE` | Importación única: claves Google separadas por coma |
| `WV_KEY_OPENROUTER` | Importación única: claves OpenRouter |
| `WV_KEY_ANTHROPIC` | Importación única: claves Anthropic |
| `WV_KEY_OPENAI` | Importación única: claves OpenAI |
| `WV_OLLAMA_URL` | Override de URL de Ollama por host |
| `WV_OLLAMA_KEEP_ALIVE` | `proxy.ollama_keep_alive` |
| `WV_OLLAMA_NUM_CTX` | `proxy.ollama_num_ctx` |
| `WV_LMSTUDIO_URL`, `WV_VLLM_URL`, `WV_LLAMACPP_URL` | Override de URL por backend |
| `WV_TOKEN_SENTINEL_FALLBACK` | `proxy.token_sentinel_fallback` |
| `WV_OAI_STREAM_FORWARD` | `proxy.oai_stream_forward` |
| `WV_ANTHROPIC_FALLBACK_MODEL` | `proxy.anthropic_fallback_model` |
| `WV_ECONOWORLD_MAX_TOKENS` | `proxy.econoworld_max_tokens` |
| `WV_ECONOWORLD_STREAM` | `proxy.econoworld_stream` |
| `WV_ECONOWORLD_REQUEST_TIMEOUT` | `proxy.econoworld_request_timeout` |

Cada variable de entorno, cuando está establecida, gana sobre el archivo YAML.

---

## Solución de problemas

### `connection refused` en `:56244`

O bien el proxy no está corriendo o está vinculado a un host diferente. Comprueba:

```bash
ss -lnp | grep 56244
systemctl --user status wall-vault-proxy   # Linux
launchctl list | grep wall-vault           # macOS
```

Si está corriendo en un puerto diferente, tu configuración tiene `proxy.port` sobreescrito — comprueba `~/.wall-vault/config.yaml`.

### `x509: certificate signed by unknown authority`

El cliente no confía en el CA interno de wall-vault. Ejecuta `wall-vault cert install-trust` en la máquina cliente. Para agentes cuyo runtime ignora el almacén de confianza del SO (p. ej. Node con un `NODE_EXTRA_CA_CERTS` codificado), usa el compañero HTTP de loopback en `127.0.0.1:56245` (solo mismo host) o establece `WV_PROXY_TLS_ENABLED=0` para hacer fallback a HTTP plano.

### `token not registered with vault`

El `Authorization: Bearer <token>` del cliente no coincide con ningún cliente registrado. Verifica el token bajo **Clients** en el panel. Si copiaste un token literal como `proxy-managed`, `dummy` o `""` de una configuración obsoleta, reemplázalo con el token de cliente real.

### `Anthropic dispatch needs a Claude model id`

Comportamiento predeterminado a partir de v0.2.63: un model id no-Claude enviado al dispatch anthropic devuelve un error. O bien arregla el enrutamiento (no envíes `gemini-2.5-flash` a anthropic) o activa la reescritura automática mediante `proxy.anthropic_fallback_model`.

### `unknown service: <id>`

El dispatch vio un id de servicio que ningún yaml de plugin reclamó. Comprueba:

```bash
ls ~/.wall-vault/services/        # ¿hay algún yaml de plugin presente?
cat ~/.wall-vault/services/<id>.yaml | grep enabled
```

Si el yaml existe pero está `enabled: false`, cámbialo. Si falta por completo, copia desde `configs/services/` en el árbol de fuentes.

### Respuesta vacía en un modelo de reasoning

`qwen3.6`, `deepseek-r1` y la familia GPT-`o1` a veces emiten solo `reasoning_content` y dejan `content` vacío. A partir de v0.2.63, wall-vault hace fallback al texto de reasoning automáticamente — si aún ves respuestas vacías, el backend no está devolviendo ninguno de los dos campos. Comprueba los logs del upstream.

Para LM Studio con qwen3 específicamente, establece `inline_no_think_for_qwen3: true` en el yaml del plugin para que reasoning se desactive inline. Los lmstudio.yaml y ollama.yaml integrados ya lo hacen.

### El panel muestra "all keys on cooldown" pero acabo de añadir una

La nueva clave es saludable pero la ruta de dispatch puede seguir en el enfriamiento de una clave más antigua. Intenta una petición nueva — el proxy hace round-robin por llamada, y se elegirá una clave saludable a continuación.

### El vault no se desbloquea con la contraseña maestra

Contraseña incorrecta. No hay recuperación — wall-vault deliberadamente no incluye una puerta trasera. Si has perdido genuinamente la contraseña maestra, el único camino es eliminar `~/.wall-vault/data/vault.json`, reiniciar con una nueva contraseña y volver a añadir las claves.

### Se han alcanzado los límites del nivel gratuito de OpenRouter

Establece `proxy.services` para incluir `openrouter` y añade al menos una clave OpenRouter. El proxy hace fallback automático de un modelo de pago a su variante `:free` cuando la ruta de pago devuelve 402 / 429.

### `journalctl --user -u wall-vault-proxy` está vacío

Los logs de systemd `--user` van al journal del usuario que lo ejecuta. Si iniciaste la unidad como `root` o vía `sudo`, el journal está en la instancia del sistema en su lugar — prueba `journalctl -u wall-vault-proxy` sin `--user`.

---

## Más

- Referencia de la API HTTP — consulta [API.md](API.md)
- Código fuente — `https://github.com/sookmook/wall-vault`
- Reportes de bugs / solicitudes de funcionalidad — GitHub Issues
- Historial de releases — [CHANGELOG.md](../CHANGELOG.md)
