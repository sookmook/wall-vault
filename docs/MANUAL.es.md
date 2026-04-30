# Manual de Usuario de wall-vault
*(Last updated: 2026-04-16 — v0.2.2)*

---

## Tabla de Contenidos

1. [¿Qué es wall-vault?](#qué-es-wall-vault)
2. [Instalación](#instalación)
3. [Primeros Pasos (Asistente de Configuración)](#primeros-pasos)
4. [Registro de Claves API](#registro-de-claves-api)
5. [Uso del Proxy](#uso-del-proxy)
6. [Panel del Almacén de Claves](#panel-del-almacén-de-claves)
7. [Modo Distribuido (Multi-Bot)](#modo-distribuido-multi-bot)
8. [Configuración de Inicio Automático](#configuración-de-inicio-automático)
9. [Doctor (Diagnóstico)](#doctor-diagnóstico)
10. [RTK Ahorro de Tokens](#rtk-ahorro-de-tokens)
11. [Referencia de Variables de Entorno](#referencia-de-variables-de-entorno)
12. [Solución de Problemas](#solución-de-problemas)

---

## ¿Qué es wall-vault?

**wall-vault = Proxy de IA + Almacén de Claves API para OpenClaw**

Para usar servicios de IA, necesitas **claves API**. Una clave API es como un **pase digital** que demuestra que "esta persona está autorizada a usar este servicio". Sin embargo, estos pases tienen límites de uso diario, y siempre existe el riesgo de exposición si no se gestionan correctamente.

wall-vault almacena estos pases en un almacén seguro y actúa como un **proxy (intermediario)** entre OpenClaw y los servicios de IA. En términos simples, OpenClaw solo necesita conectarse a wall-vault, y wall-vault se encarga del resto.

Problemas que resuelve wall-vault:

- **Rotación automática de claves API**: Cuando una clave alcanza su límite de uso o es bloqueada temporalmente (cooldown), el sistema cambia silenciosamente a la siguiente clave. OpenClaw continúa funcionando sin interrupción.
- **Fallback automático de servicio**: Si Google no responde, cambia a OpenRouter. Si eso también falla, cambia automáticamente a la IA instalada localmente (Ollama, LM Studio, vLLM). Tu sesión nunca se interrumpe. Cuando el servicio original se recupera, vuelve automáticamente en la siguiente solicitud (v0.1.18+, LM Studio/vLLM: v0.1.21+).
- **Sincronización en tiempo real (SSE)**: Cuando cambias el modelo en el panel del almacén, se refleja en la pantalla de OpenClaw en 1-3 segundos. SSE (Server-Sent Events) es una tecnología donde el servidor envía cambios a los clientes en tiempo real.
- **Notificaciones en tiempo real**: Eventos como el agotamiento de claves o caídas de servicio se muestran inmediatamente en la parte inferior del TUI de OpenClaw (pantalla de terminal).

> :bulb: **Claude Code, Cursor y VS Code** también se pueden conectar, pero el propósito principal de wall-vault es usarse con OpenClaw.

```
OpenClaw (pantalla terminal TUI)
        |
        v
  wall-vault proxy (:56244)   <- Gestión de claves, enrutamiento, fallback, eventos
        |
        +-- Google Gemini API
        +-- OpenRouter API (340+ modelos)
        +-- Ollama / LM Studio / vLLM (máquina local, último recurso)
        +-- OpenAI / Anthropic API
```

---

## Instalación

### Linux / macOS

Abre una terminal y pega los siguientes comandos:

```bash
# Linux (PC estándar, servidor — amd64)
curl -L https://github.com/sookmook/wall-vault/releases/latest/download/wall-vault-linux-amd64 \
  -o ~/.local/bin/wall-vault && chmod +x ~/.local/bin/wall-vault

# macOS Apple Silicon (M1/M2/M3 Mac)
curl -L https://github.com/sookmook/wall-vault/releases/latest/download/wall-vault-darwin-arm64 \
  -o /usr/local/bin/wall-vault && chmod +x /usr/local/bin/wall-vault
```

- `curl -L ...` — Descarga el archivo desde Internet.
- `chmod +x` — Hace que el archivo descargado sea "ejecutable". Si omites este paso, obtendrás un error de "permiso denegado".

### Windows

Abre PowerShell (como administrador) y ejecuta lo siguiente:

```powershell
# Descarga
Invoke-WebRequest -Uri `
  "https://github.com/sookmook/wall-vault/releases/latest/download/wall-vault-windows-amd64.exe" `
  -OutFile "$env:LOCALAPPDATA\Programs\wall-vault\wall-vault.exe"

# Agregar al PATH (se aplica después de reiniciar PowerShell)
$env:PATH += ";$env:LOCALAPPDATA\Programs\wall-vault"
```

> :bulb: **¿Qué es PATH?** Es una lista de carpetas donde tu computadora busca comandos. Necesitas agregar wall-vault al PATH para poder ejecutar `wall-vault` desde cualquier carpeta.

### Compilación desde código fuente (para desarrolladores)

Solo aplicable si tienes instalado el entorno de desarrollo de Go.

```bash
git clone https://github.com/sookmook/wall-vault
cd wall-vault
make build       # bin/wall-vault (versión: v0.1.25.YYYYMMDD.HHmmss)
make install     # ~/.local/bin/wall-vault
```

> :bulb: **Versión con marca de tiempo de compilación**: Al compilar con `make build`, la versión se genera automáticamente en un formato como `v0.1.27.20260409` que incluye fecha y hora. Si compilas directamente con `go build ./...`, la versión solo mostrará `"dev"`.

---

## Primeros Pasos

### Ejecutar el Asistente de Configuración

Después de la instalación, asegúrate de ejecutar el **asistente de configuración** con el siguiente comando. El asistente te guiará paso a paso a través de los elementos necesarios.

```bash
wall-vault setup
```

El asistente sigue estos pasos:

```
1. Selección de idioma (10 idiomas incluyendo español)
2. Selección de tema (light / dark / gold / cherry / ocean)
3. Modo de operación — standalone (máquina única) o distribuido (múltiples máquinas)
4. Nombre del bot — el nombre que se muestra en el panel
5. Configuración de puertos — por defecto: proxy 56244, almacén 56243 (Enter para mantener valores predeterminados)
6. Selección de servicios de IA — elegir entre Google / OpenRouter / Ollama / LM Studio / vLLM
7. Configuración del filtro de seguridad de herramientas
8. Token de administrador — una contraseña que bloquea las funciones de administración del panel. Generación automática disponible
9. Contraseña de cifrado de claves API — para un almacenamiento más seguro de claves (opcional)
10. Ruta de guardado del archivo de configuración
```

> :warning: **Recuerda tu token de administrador.** Lo necesitarás más tarde para agregar claves o cambiar configuraciones en el panel. Si lo pierdes, tendrás que editar el archivo de configuración manualmente.

Una vez completado el asistente, se crea automáticamente un archivo de configuración `wall-vault.yaml`.

### Inicio

```bash
wall-vault start
```

Los siguientes dos servidores se inician simultáneamente:

- **Proxy** (`https://localhost:56244`) — El intermediario que conecta OpenClaw y los servicios de IA
- **Almacén de claves** (`https://localhost:56243`) — Gestión de claves API y panel web

Abre `https://localhost:56243` en tu navegador para acceder al panel.

---

## Registro de Claves API

Hay cuatro formas de registrar claves API. **Se recomienda el Método 1 (variables de entorno) para principiantes.**

### Método 1: Variables de Entorno (recomendado — el más simple)

Las variables de entorno son **valores preestablecidos** que los programas leen al iniciarse. Solo escribe lo siguiente en tu terminal:

```bash
# Registrar una clave de Google Gemini
export WV_KEY_GOOGLE=AIzaSy...

# Registrar una clave de OpenRouter
export WV_KEY_OPENROUTER=sk-or-v1-...

# Iniciar después del registro
wall-vault start
```

Si tienes múltiples claves, sepáralas con comas. wall-vault las usará en rotación automáticamente (round robin):

```bash
export WV_KEY_GOOGLE=AIzaSy...,AIzaSy...,AIzaSy...
```

> :bulb: **Consejo**: El comando `export` solo se aplica a la sesión de terminal actual. Para persistir después de reinicios, agrega las líneas a tu archivo `~/.bashrc` o `~/.zshrc`.

### Método 2: Interfaz del Panel (apuntar y hacer clic)

1. Abre `https://localhost:56243` en tu navegador
2. Haz clic en el botón `[+ Agregar]` en la tarjeta **:key: Claves API** superior
3. Ingresa el tipo de servicio, valor de la clave, etiqueta (nombre de memo) y límite diario, luego guarda

### Método 3: API REST (para automatización/scripts)

La API REST es un método para que los programas intercambien datos vía HTTP. Útil para el registro automatizado mediante scripts.

```bash
curl -X POST https://localhost:56243/admin/keys \
  -H "Authorization: Bearer token-admin" \
  -H "Content-Type: application/json" \
  -d '{
    "service": "google",
    "key": "AIzaSy...",
    "label": "Clave principal",
    "daily_limit": 1000
  }'
```

### Método 4: Flags del Proxy (para pruebas rápidas)

Usa esto para pruebas temporales sin registro formal. Las claves desaparecen cuando el programa se cierra.

```bash
wall-vault proxy --key-google=AIzaSy... --key-openrouter=sk-or-...
```

---

## Uso del Proxy

### Uso con OpenClaw (propósito principal)

Así se configura OpenClaw para conectarse a servicios de IA a través de wall-vault.

Abre `~/.openclaw/openclaw.json` y agrega lo siguiente:

```json5
// ~/.openclaw/openclaw.json
{
  models: {
    providers: {
      "wall-vault": {
        baseUrl: "https://localhost:56244/v1",
        apiKey: "your-agent-token",   // token de agente del almacén
        api: "openai-completions",
        models: [
          { id: "wall-vault/gemini-2.5-flash" },
          { id: "wall-vault/gemini-2.5-pro" },
          { id: "wall-vault/hunter-alpha" },    // contexto 1M gratuito
          { id: "wall-vault/claude-opus-4-6" }
        ]
      }
    }
  }
}
```

> :bulb: **Método más fácil**: Presiona el botón **:lobster: Copiar config OpenClaw** en la tarjeta de agente del panel. Un snippet con el token y la dirección ya completados se copiará a tu portapapeles. Solo pégalo.

**¿A dónde redirige el prefijo `wall-vault/` en el nombre del modelo?**

wall-vault determina automáticamente a qué servicio de IA enviar la solicitud según el nombre del modelo:

| Formato del modelo | Servicio enrutado |
|-------------------|------------------|
| `wall-vault/gemini-*` | Directo a Google Gemini |
| `wall-vault/gpt-*`, `wall-vault/o3`, `wall-vault/o4*` | Directo a OpenAI |
| `wall-vault/claude-*` | Anthropic vía OpenRouter |
| `wall-vault/hunter-alpha`, `wall-vault/healer-alpha` | OpenRouter (contexto 1M tokens gratuito) |
| `wall-vault/kimi-*`, `wall-vault/glm-*`, `wall-vault/deepseek-*` | OpenRouter |
| `google/nombre-modelo`, `openai/nombre-modelo`, `anthropic/nombre-modelo`, etc. | Directo al servicio correspondiente |
| `custom/google/nombre-modelo`, `custom/openai/nombre-modelo`, etc. | Elimina el prefijo `custom/` y redirige |
| `nombre-modelo:cloud` | Elimina el sufijo `:cloud` y redirige a OpenRouter |

> :bulb: **¿Qué es el contexto?** Es la cantidad de conversación que una IA puede recordar a la vez. 1M (un millón de tokens) significa que puede procesar conversaciones o documentos muy largos en una sola pasada.

### Conexión directa vía formato API Gemini (compatibilidad con herramientas existentes)

Si tienes herramientas que usan directamente la API de Google Gemini, solo cambia la URL a wall-vault:

```bash
export ANTHROPIC_BASE_URL=https://localhost:56244/google
```

O si la herramienta especifica URLs directamente:

```
https://localhost:56244/google/v1beta/models/gemini-2.5-flash:generateContent
```

### Uso con el SDK de OpenAI (Python)

Puedes conectar wall-vault a código Python que usa IA. Solo cambia el `base_url`:

```python
from openai import OpenAI

client = OpenAI(
    base_url="https://localhost:56244/v1",
    api_key="not-needed"  # wall-vault gestiona las claves API por ti
)

response = client.chat.completions.create(
    model="google/gemini-2.5-flash",   # Usa el formato provider/model
    messages=[{"role": "user", "content": "Hola"}]
)
```

### Cambiar modelos en tiempo de ejecución

Para cambiar el modelo de IA mientras wall-vault ya está en ejecución:

```bash
# Cambiar modelo vía solicitud directa al proxy
curl -X PUT https://localhost:56244/api/config/model \
  -H "Content-Type: application/json" \
  -d '{"service": "openrouter", "model": "anthropic/claude-3.5-sonnet"}'

# En modo distribuido (multi-bot), cambiar en el servidor almacén -> sincronizado instantáneamente vía SSE
curl -X PUT https://localhost:56243/admin/clients/mi-bot-id \
  -H "Authorization: Bearer token-admin" \
  -H "Content-Type: application/json" \
  -d '{"default_service": "google", "default_model": "gemini-2.5-pro"}'
```

### Listar modelos disponibles

```bash
# Ver lista completa
curl https://localhost:56244/api/models | python3 -m json.tool

# Ver solo modelos de Google
curl "https://localhost:56244/api/models?service=google"

# Buscar por nombre (ej: modelos que contienen "claude")
curl "https://localhost:56244/api/models?q=claude"
```

**Modelos principales por servicio:**

| Servicio | Modelos principales |
|----------|-------------------|
| Google | gemini-2.5-pro, gemini-2.5-flash, gemini-2.5-flash-8b, gemini-2.0-flash |
| OpenAI | gpt-4o, gpt-4o-mini, o3, o1, o1-mini |
| OpenRouter | 346+ (Hunter Alpha 1M contexto gratuito, DeepSeek R1/V3, Qwen 2.5, etc.) |
| Ollama | Detección automática desde servidor instalado localmente |
| LM Studio | Servidor local (puerto 1234) |
| vLLM | Servidor local (puerto 8000) |
| llama.cpp | Servidor local (puerto 8080) |

---

## Panel del Almacén de Claves

Accede al panel abriendo `https://localhost:56243` en tu navegador.

**Diseño de pantalla:**
- **Barra superior (fija)**: Logo, selector de idioma/tema, estado de conexión SSE
- **Cuadrícula de tarjetas**: Tarjetas de agentes, servicios y claves API en formato de mosaico

### Tarjeta de Claves API

Una tarjeta para gestionar todas las claves API registradas de un vistazo.

- Muestra la lista de claves agrupadas por servicio.
- `today_usage`: Tokens procesados exitosamente hoy (caracteres leídos y escritos por la IA)
- `today_attempts`: Total de llamadas hoy (éxitos + fallos combinados)
- Usa el botón `[+ Agregar]` para registrar nuevas claves, y `x` para eliminar.

> :bulb: **¿Qué son los tokens?** Los tokens son las unidades que la IA usa para procesar texto. Aproximadamente una palabra en inglés, o 1-2 caracteres en español. Los precios de API se calculan generalmente según el número de tokens.

### Tarjeta de Agente

Una tarjeta que muestra el estado de los bots (agentes) conectados al proxy wall-vault.

**El estado de conexión se muestra en 4 niveles:**

| Indicador | Estado | Significado |
|-----------|--------|-------------|
| :green_circle: | En ejecución | El proxy está funcionando normalmente |
| :yellow_circle: | Retrasado | Responde pero lento |
| :red_circle: | Sin conexión | El proxy no responde |
| :black_circle: | No conectado / Deshabilitado | El proxy nunca se ha conectado al almacén o está deshabilitado |

**Guía de botones en la parte inferior de la tarjeta de agente:**

Cuando registras un agente y especificas el **tipo de agente**, aparecen automáticamente botones de conveniencia para ese tipo.

---

#### :radio_button: Botón Copiar Configuración — Genera automáticamente la configuración de conexión

Al hacer clic en el botón, se copia al portapapeles un snippet de configuración con el token del agente, la dirección del proxy y la información del modelo ya completados. Solo pega el contenido copiado en la ubicación mostrada en la tabla a continuación.

| Botón | Tipo de agente | Ubicación para pegar |
|-------|---------------|---------------------|
| :lobster: Copiar config OpenClaw | `openclaw` | `~/.openclaw/openclaw.json` |
| :crab: Copiar config NanoClaw | `nanoclaw` | `~/.openclaw/openclaw.json` |
| :orange_circle: Copiar config Claude Code | `claude-code` | `~/.claude/settings.json` |
| :keyboard: Copiar config Cursor | `cursor` | Cursor -> Settings -> AI |
| :computer: Copiar config VSCode | `vscode` | `~/.continue/config.json` |

**Ejemplo — Para el tipo Claude Code, esto es lo que se copia:**

```json
// ~/.claude/settings.json
{
  "apiProvider": "openai",
  "baseUrl": "http://192.168.1.20:56244/v1",
  "apiKey": "token-de-este-agente"
}
```

**Ejemplo — Para el tipo VSCode (Continue):**

```yaml
# ~/.continue/config.yaml  <- pegar en config.yaml, NO config.json
name: My Config
version: 0.0.1
schema: v1

models:
  - name: wall-vault proxy
    provider: openai
    model: gemini-2.5-flash
    apiBase: http://192.168.1.20:56244/v1
    apiKey: token-de-este-agente
    roles:
      - chat
      - edit
      - apply
```

> :warning: **Las versiones recientes de Continue usan `config.yaml`.** Si `config.yaml` existe, `config.json` se ignora completamente. Asegúrate de pegar en `config.yaml`.

**Ejemplo — Para el tipo Cursor:**

```
Base URL : http://192.168.1.20:56244/v1
API Key  : token-de-este-agente

// O variables de entorno:
OPENAI_BASE_URL=http://192.168.1.20:56244/v1
OPENAI_API_KEY=token-de-este-agente
```

> :warning: **Si la copia al portapapeles no funciona**: Las políticas de seguridad del navegador pueden bloquear la copia. Si aparece un cuadro de texto emergente, selecciona todo con Ctrl+A y copia con Ctrl+C.

---

#### :zap: Botón de Aplicación Automática — Un clic y estás configurado

Para agentes de tipo `cline`, `claude-code`, `openclaw` o `nanoclaw`, aparece un botón **:zap: Aplicar Config** en la tarjeta de agente. Al hacer clic, se actualiza automáticamente el archivo de configuración local del agente.

| Botón | Tipo de agente | Archivo destino |
|-------|---------------|----------------|
| :zap: Aplicar config Cline | `cline` | `~/.cline/data/globalState.json` + `secrets.json` |
| :zap: Aplicar config Claude Code | `claude-code` | `~/.claude/settings.json` |
| :zap: Aplicar config OpenClaw | `openclaw` | `~/.openclaw/openclaw.json` |
| :zap: Aplicar config NanoClaw | `nanoclaw` | `~/.openclaw/openclaw.json` |

> :warning: Este botón envía una solicitud a **localhost:56244** (proxy local). El proxy debe estar en ejecución en esa máquina.

---

#### :twisted_rightwards_arrows: Ordenar tarjetas con arrastrar y soltar (v0.1.17, mejorado v0.1.25)

Puedes **arrastrar** las tarjetas de agentes en el panel para reorganizarlas en cualquier orden.

1. Agarra el área del **semáforo (●)** en la parte superior izquierda de una tarjeta con el ratón y arrastra
2. Suéltala sobre la tarjeta en la posición deseada para intercambiar su orden

> :bulb: El cuerpo de la tarjeta (campos de entrada, botones, etc.) no se puede arrastrar. Solo puedes agarrar desde el área del semáforo.

#### :orange_circle: Detección de Proceso de Agente (v0.1.25)

Cuando el proxy funciona normalmente pero un proceso de agente local (NanoClaw, OpenClaw) ha muerto, el semáforo de la tarjeta cambia a **naranja (parpadeante)** y muestra un mensaje "Proceso de agente detenido".

- :green_circle: Verde: Proxy + agente normal
- :orange_circle: Naranja (parpadeante): Proxy normal, agente muerto
- :red_circle: Rojo: Proxy sin conexión
3. El orden modificado se **guarda en el servidor inmediatamente** y persiste al actualizar la página

> :bulb: Los dispositivos táctiles (móvil/tableta) aún no son compatibles. Usa un navegador de escritorio.

---

#### :arrows_counterclockwise: Sincronización Bidireccional de Modelos (v0.1.16)

Cuando cambias el modelo de un agente en el panel del almacén, la configuración local del agente se actualiza automáticamente.

**Para Cline:**
- Cambio de modelo en el almacén -> evento SSE -> el proxy actualiza los campos de modelo en `globalState.json`
- Campos actualizados: `actModeOpenAiModelId`, `planModeOpenAiModelId`, `openAiModelId`
- `openAiBaseUrl` y la clave API no se tocan
- **Se requiere recargar VS Code (`Ctrl+Alt+R` o `Ctrl+Shift+P` -> `Developer: Reload Window`)**
  - Porque Cline no relee los archivos de configuración mientras se ejecuta

**Para Claude Code:**
- Cambio de modelo en el almacén -> evento SSE -> el proxy actualiza el campo `model` en `settings.json`
- Busca automáticamente en rutas de WSL y Windows (`~/.claude/`, `/mnt/c/Users/*/.claude/`)

**Dirección inversa (agente -> almacén):**
- Cuando los agentes (Cline, Claude Code, etc.) envían solicitudes al proxy, este incluye la información de servicio/modelo del cliente en el heartbeat
- La tarjeta de agente del panel muestra el servicio/modelo actualmente en uso en tiempo real

> :bulb: **Punto clave**: El proxy identifica a los agentes por el token Authorization en las solicitudes y redirige automáticamente al servicio/modelo configurado en el almacén. Incluso si Cline o Claude Code envía un nombre de modelo diferente, el proxy lo sobrescribe con la configuración del almacén.

---

### Usar Cline en VS Code — Guía Detallada

#### Paso 1: Instalar Cline

Instala **Cline** (ID: `saoudrizwan.claude-dev`) desde el Marketplace de Extensiones de VS Code.

#### Paso 2: Registrar Agente en el Almacén

1. Abre el panel del almacén (`http://IP-almacén:56243`)
2. Haz clic en **+ Agregar** en la sección **Agentes**
3. Completa lo siguiente:

| Campo | Valor | Descripción |
|-------|-------|-------------|
| ID | `mi_cline` | Identificador único (alfanumérico, sin espacios) |
| Nombre | `Mi Cline` | Nombre mostrado en el panel |
| Tipo de agente | `cline` | <- Debe seleccionar `cline` |
| Servicio | Elegir servicio (ej: `google`) | |
| Modelo | Ingresar modelo (ej: `gemini-2.5-flash`) | |

4. Haz clic en **Guardar** para generar automáticamente un token

#### Paso 3: Conectar Cline

**Método A — Aplicación automática (recomendado)**

1. Verifica que el **proxy** de wall-vault esté en ejecución en esta máquina (`localhost:56244`)
2. Haz clic en el botón **:zap: Aplicar config Cline** en la tarjeta de agente del panel
3. Éxito cuando veas la notificación "¡Configuración aplicada!"
4. Recarga VS Code (`Ctrl+Alt+R`)

**Método B — Configuración manual**

Abre la configuración (:gear:) en la barra lateral de Cline:
- **API Provider**: `OpenAI Compatible`
- **Base URL**: `http://dirección-proxy:56244/v1`
  - Misma máquina: `https://localhost:56244/v1`
  - Máquina diferente (ej: Mac Mini): `http://192.168.1.20:56244/v1`
- **API Key**: Token emitido por el almacén (copiar de la tarjeta de agente)
- **Model ID**: Modelo configurado en el almacén (ej: `gemini-2.5-flash`)

#### Paso 4: Verificar

Envía cualquier mensaje en la ventana de chat de Cline. Si funciona correctamente:
- La tarjeta de agente correspondiente en el panel muestra un **punto verde (En ejecución)**
- La tarjeta muestra el servicio/modelo actual (ej: `google / gemini-2.5-flash`)

#### Cambiar Modelos

Para cambiar el modelo de Cline, hazlo desde el **panel del almacén**:

1. Cambia el menú desplegable de servicio/modelo en la tarjeta de agente
2. Haz clic en **Aplicar**
3. Recarga VS Code (`Ctrl+Alt+R`) — El nombre del modelo en el pie de página de Cline se actualiza
4. El nuevo modelo se usa desde la siguiente solicitud

> :bulb: En la práctica, el proxy identifica las solicitudes de Cline por el token y las redirige al modelo configurado en el almacén. Incluso sin recargar VS Code, **el modelo realmente usado cambia inmediatamente** — la recarga solo es para actualizar la visualización del modelo en la interfaz de Cline.

#### Detección de Desconexión

Cuando cierras VS Code, la tarjeta de agente en el panel se vuelve amarilla (retrasado) después de unos **90 segundos**, y luego roja (sin conexión) después de **3 minutos**. (Desde v0.1.18, las verificaciones de estado cada 15 segundos han hecho la detección de desconexión más rápida.)

#### Solución de Problemas

| Síntoma | Causa | Solución |
|---------|-------|----------|
| Error "Conexión fallida" en Cline | Proxy no ejecutándose o dirección incorrecta | Verificar proxy con `curl https://localhost:56244/health` |
| El punto verde no aparece en el almacén | Clave API (token) no configurada | Hacer clic en **:zap: Aplicar config Cline** de nuevo |
| El modelo en el pie de página de Cline no cambia | Cline almacena la configuración en caché | Recargar VS Code (`Ctrl+Alt+R`) |
| Se muestra un nombre de modelo incorrecto | Bug antiguo (corregido en v0.1.16) | Actualizar proxy a v0.1.16 o posterior |

---

#### :purple_circle: Botón Copiar Comando de Despliegue — Para instalar en nuevas máquinas

Úsalo al instalar por primera vez el proxy wall-vault en una nueva computadora y conectarlo al almacén. Al hacer clic, se copia el script de instalación completo. Pégalo en la terminal de la nueva computadora y ejecútalo — todo lo siguiente se maneja de una vez:

1. Instalación del binario wall-vault (se omite si ya está instalado)
2. Registro automático del servicio de usuario systemd
3. Inicio del servicio y conexión automática al almacén

> :bulb: El script viene con el token de este agente y la dirección del servidor almacén ya completados, así que puedes ejecutarlo inmediatamente después de pegar sin ninguna modificación.

---

### Tarjeta de Servicio

Una tarjeta para habilitar/deshabilitar y configurar servicios de IA.

- Interruptor de activación/desactivación por servicio
- Ingresa la dirección de servidores de IA locales (Ollama, LM Studio, vLLM, llama.cpp, etc. en tu máquina) y los modelos disponibles se descubren automáticamente.
- **Estado de conexión del servicio local**: El punto junto al nombre del servicio es **verde** si está conectado, **gris** si no
- **Semáforo automático de servicio local** (v0.1.23+): Los servicios locales (Ollama, LM Studio, vLLM, llama.cpp) se activan/desactivan automáticamente según la conectividad. Cuando un servicio se vuelve accesible, cambia a verde y la casilla se activa en 15 segundos; cuando se vuelve inaccesible, se desactiva automáticamente. Funciona de la misma manera que los servicios en la nube (Google, OpenRouter, etc.) que se alternan automáticamente según la disponibilidad de claves API.
- **Interruptor de modo de razonamiento** (v0.2.17+): Aparece una casilla de **modo de razonamiento** en la parte inferior del panel de edición del servicio local. Al activarla, el proxy agrega `"reasoning": true` al cuerpo de chat-completions enviado al upstream, de modo que los modelos que exponen su proceso de pensamiento — como DeepSeek R1 o Qwen QwQ — devuelven un bloque `<think>…</think>` junto con la respuesta. Los servidores que no reconocen este campo simplemente lo ignoran, así que es seguro dejarla activada incluso en cargas de trabajo mixtas.

> :bulb: **Si un servicio local se ejecuta en otra computadora**: Ingresa la IP de esa computadora en el campo URL del servicio. Ejemplo: `http://192.168.1.20:11434` (Ollama), `http://192.168.1.20:1234` (LM Studio), `http://192.168.1.20:8080` (llama.cpp). Si el servicio está vinculado a `127.0.0.1` en lugar de `0.0.0.0`, el acceso por IP externa no funcionará — verifica la dirección de enlace en la configuración del servicio.

### Entrada del Token de Administrador

Cuando intentas usar funciones importantes como agregar o eliminar claves en el panel, aparece una ventana emergente para ingresar el token de administrador. Ingresa el token que configuraste durante el asistente de configuración. Una vez ingresado, persiste hasta que cierres el navegador.

> :warning: **Si los fallos de autenticación superan los 10 en 15 minutos, esa IP se bloquea temporalmente.** Si olvidaste tu token, verifica el campo `admin_token` en tu archivo `wall-vault.yaml`.

---

## Modo Distribuido (Multi-Bot)

Al ejecutar OpenClaw simultáneamente en múltiples computadoras, esta configuración **comparte un único almacén de claves**. Es conveniente porque solo necesitas gestionar las claves en un lugar.

### Ejemplo de Configuración

```
[Servidor de Almacén de Claves]
  wall-vault vault    (almacén de claves :56243, panel)

[WSL Alpha]           [Raspberry Pi Gamma]   [Mac Mini Local]
  wall-vault proxy      wall-vault proxy        wall-vault proxy
  openclaw TUI          openclaw TUI            openclaw TUI
  <-> sync SSE          <-> sync SSE            <-> sync SSE
```

Todos los bots apuntan al servidor de almacén central, así que los cambios de modelo o adiciones de claves en el almacén se reflejan instantáneamente en todos los bots.

### Paso 1: Iniciar el Servidor del Almacén de Claves

Ejecuta esto en la computadora que servirá como servidor del almacén:

```bash
wall-vault vault
```

### Paso 2: Registrar cada Bot (Cliente)

Pre-registra la información de cada bot que se conectará al servidor del almacén:

```bash
curl -X POST https://localhost:56243/admin/clients \
  -H "Authorization: Bearer token-admin" \
  -H "Content-Type: application/json" \
  -d '{
    "id": "botA",
    "name": "Bot A",
    "token": "bota-secret",
    "default_service": "google",
    "default_model": "gemini-2.5-flash"
  }'
```

### Paso 3: Iniciar el Proxy en cada Máquina Bot

En cada máquina bot, inicia el proxy con la dirección del servidor almacén y el token:

```bash
WV_VAULT_URL=http://192.168.x.x:56243 \
WV_VAULT_TOKEN=bota-secret \
WV_VAULT_CLIENT_ID=botA \
wall-vault proxy
```

> :bulb: Reemplaza **`192.168.x.x`** con la dirección IP interna real de la máquina del servidor almacén. Puedes encontrarla en la configuración del router o mediante el comando `ip addr`.

---

## Configuración de Inicio Automático

Si es tedioso iniciar manualmente wall-vault cada vez que reinicias, regístralo como un servicio del sistema. Una vez registrado, se inicia automáticamente al arrancar.

### Linux — systemd (la mayoría de distribuciones Linux)

systemd es el sistema que inicia y gestiona automáticamente los programas en Linux:

```bash
wall-vault doctor deploy
systemctl --user daemon-reload
systemctl --user enable --now wall-vault
```

Consultar registros:

```bash
journalctl --user -u wall-vault -f
```

### macOS — launchd

El sistema responsable del inicio automático de programas en macOS:

```bash
wall-vault doctor deploy launchd
launchctl load ~/Library/LaunchAgents/com.wall-vault.plist
```

### Windows — NSSM

1. Descarga NSSM desde [nssm.cc](https://nssm.cc/download) y agrégalo al PATH.
2. En un PowerShell de administrador:

```powershell
wall-vault doctor deploy windows
```

---

## Doctor (Diagnóstico)

El comando `doctor` es una herramienta que **auto-diagnostica y repara** problemas de configuración de wall-vault.

```bash
wall-vault doctor check   # Diagnosticar estado actual (solo lectura, no cambia nada)
wall-vault doctor fix     # Reparar automáticamente problemas
wall-vault doctor all     # Diagnóstico + reparación automática en un paso
```

> :bulb: Si algo parece andar mal, ejecuta primero `wall-vault doctor all`. Detecta y corrige muchos problemas automáticamente.

---

## RTK Ahorro de Tokens

*(v0.1.24+)*

**RTK (Token Reduction Kit)** comprime automáticamente la salida de comandos shell ejecutados por agentes de codificación IA (como Claude Code), reduciendo el uso de tokens. Por ejemplo, 15 líneas de salida de `git status` pueden reducirse a un resumen de 2 líneas.

### Uso Básico

```bash
# Envuelve comandos con wall-vault rtk para filtrar automáticamente la salida
wall-vault rtk git status          # Muestra solo la lista de archivos modificados
wall-vault rtk git diff HEAD~1     # Solo líneas modificadas + contexto mínimo
wall-vault rtk git log -10         # Hash + mensaje de una línea cada uno
wall-vault rtk go test ./...       # Muestra solo tests fallidos
wall-vault rtk ls -la              # Comandos no soportados se truncan automáticamente
```

### Comandos Soportados y Ahorro

| Comando | Método de filtrado | Ahorro |
|---------|-------------------|--------|
| `git status` | Solo resumen de archivos modificados | ~87% |
| `git diff` | Líneas modificadas + 3 líneas de contexto | ~60-94% |
| `git log` | Hash + primera línea del mensaje | ~90% |
| `git push/pull/fetch` | Progreso eliminado, solo resumen | ~80% |
| `go test` | Solo fallos, éxitos contados | ~88-99% |
| `go build/vet` | Solo errores | ~90% |
| Todos los demás comandos | Primeras 50 + últimas 50 líneas, máx 32KB | Variable |

### Pipeline de Filtrado en 3 Etapas

1. **Filtro estructural específico por comando** — Entiende el formato de salida de git, go, etc. y extrae partes significativas
2. **Post-procesamiento con regex** — Elimina códigos de color ANSI, colapsa líneas vacías, agrega líneas duplicadas
3. **Paso directo + truncamiento** — Comandos no soportados conservan solo las primeras/últimas 50 líneas

### Integración con Claude Code

Puedes configurar el hook `PreToolUse` de Claude Code para pasar automáticamente todos los comandos shell por RTK.

```bash
# Instalar hook (se agrega automáticamente a settings.json de Claude Code)
wall-vault rtk hook install
```

O agregar manualmente a `~/.claude/settings.json`:

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

> :bulb: **Preservación del código de salida**: RTK devuelve el código de salida del comando original sin cambios. Si un comando falla (código de salida != 0), la IA detecta el fallo con precisión.

> :bulb: **Inglés forzado**: RTK ejecuta comandos con `LC_ALL=C`, asegurando salida en inglés independientemente de la configuración de idioma del sistema. Esto es necesario para que los filtros funcionen correctamente.

---

## Referencia de Variables de Entorno

Las variables de entorno son una forma de pasar valores de configuración a los programas. Escribe `export VARIABLE=valor` en la terminal, o colócalas en archivos de servicio de inicio automático para aplicación permanente.

| Variable | Descripción | Ejemplo |
|----------|-------------|---------|
| `WV_LANG` | Idioma del panel | `ko`, `en`, `ja` |
| `WV_THEME` | Tema del panel | `light`, `dark`, `gold` |
| `WV_KEY_GOOGLE` | Clave API de Google (separadas por comas para múltiples) | `AIza...,AIza...` |
| `WV_KEY_OPENROUTER` | Clave API de OpenRouter | `sk-or-v1-...` |
| `WV_VAULT_URL` | Dirección del servidor almacén en modo distribuido | `http://192.168.x.x:56243` |
| `WV_VAULT_TOKEN` | Token de autenticación del cliente (bot) | `my-secret-token` |
| `WV_ADMIN_TOKEN` | Token de administrador | `admin-token-here` |
| `WV_MASTER_PASS` | Contraseña de cifrado de claves API | `my-password` |
| `WV_AVATAR` | Ruta del archivo de imagen de avatar (relativa a `~/.openclaw/`) | `workspace/avatars/avatar.png` |
| `OLLAMA_URL` | Dirección del servidor local de Ollama | `http://192.168.x.x:11434` |

---

## Solución de Problemas

### El Proxy No Inicia

El puerto probablemente ya está en uso por otro programa.

```bash
ss -tlnp | grep 56244   # Verificar qué usa el puerto 56244
wall-vault proxy --port 8080   # Iniciar en un puerto diferente
```

### Errores de Clave API (429, 402, 401, 403, 582)

| Código de Error | Significado | Resolución |
|----------------|-------------|------------|
| **429** | Demasiadas solicitudes (cuota excedida) | Esperar o agregar más claves |
| **402** | Pago requerido o créditos agotados | Recargar créditos en el servicio |
| **401 / 403** | Clave inválida o sin permiso | Verificar valor de la clave y re-registrar |
| **582** | Sobrecarga de gateway (cooldown de 5 minutos) | Se resuelve automáticamente después de 5 minutos |

```bash
# Verificar lista de claves registradas y estado
curl -H "Authorization: Bearer token-admin" https://localhost:56243/admin/keys

# Resetear contadores de uso de claves
curl -X POST -H "Authorization: Bearer token-admin" https://localhost:56243/admin/keys/reset
```

### El Agente Muestra "No Conectado"

"No conectado" significa que el proceso proxy no está enviando heartbeats al almacén. **No significa que las configuraciones no se hayan guardado.** El proxy debe estar en ejecución con la dirección del servidor almacén y el token para entrar en estado conectado.

```bash
# Iniciar proxy con dirección del servidor almacén, token e ID de cliente
WV_VAULT_URL=http://servidor-almacen:56243 \
WV_VAULT_TOKEN=token-cliente \
WV_VAULT_CLIENT_ID=id-cliente \
wall-vault proxy
```

Una vez conectado, el panel muestra :green_circle: En ejecución en aproximadamente 20 segundos.

### Ollama No Se Conecta

Ollama es un programa que ejecuta IA directamente en tu computadora. Primero verifica si Ollama está en ejecución.

```bash
curl http://localhost:11434/api/tags   # Si aparece una lista de modelos, funciona
export OLLAMA_URL=http://192.168.x.x:11434   # Si se ejecuta en otra computadora
```

> :warning: Si Ollama no responde, inícialo primero con el comando `ollama serve`.

> :warning: **Los modelos grandes son lentos**: Modelos grandes como `qwen3.5:35b` o `deepseek-r1` pueden tardar varios minutos en generar una respuesta. Incluso si parece que nada sucede, puede estar procesando — ten paciencia.

---

## Notas de actualización de v0.2

- `Service` ahora cuenta con `default_model` y `allowed_models`. El modelo predeterminado por servicio se establece ahora directamente en la tarjeta de servicio.
- `Client.default_service` / `default_model` han sido renombrados y reinterpretados como `preferred_service` / `model_override`. Si el override está vacío, se utiliza el modelo predeterminado del servicio.
- En el primer inicio de v0.2, el archivo `vault.json` existente se migra automáticamente, y el estado anterior a la migración se preserva como `vault.json.pre-v02.{timestamp}.bak`.
- El panel ha sido reestructurado en tres zonas: una barra lateral izquierda, una cuadrícula de tarjetas central, y un panel deslizable de edición en el lado derecho.
- Las rutas de la API Admin no han cambiado, pero los esquemas de los cuerpos de solicitud/respuesta se han actualizado — los scripts CLI antiguos deberán actualizarse en consecuencia.

---

## Novedades de v0.2.1

- **Paso multimodal (OpenAI → Gemini)**: `/v1/chat/completions` ahora acepta seis tipos de partes de contenido además de `text` — `input_audio`, `input_video`, `input_image`, `input_file` e `image_url` (URIs de datos y URLs externas http(s) de hasta 5 MB). El proxy convierte cada una al formato `inlineData` de Gemini. Clientes compatibles con OpenAI como EconoWorld pueden transmitir directamente blobs de audio, imagen y vídeo.
- **Tipo de agente EconoWorld**: `POST /agent/apply` con `agentType: "econoworld"` escribe la configuración de wall-vault en el archivo `analyzer/ai_config.json` del proyecto. `workDir` acepta una lista separada por comas de rutas candidatas y convierte las rutas de unidad de Windows a rutas de montaje WSL.
- **Cuadrícula de claves en el panel + CRUD**: Las 11 claves se representan como tarjetas compactas con panel deslizable para + añadir / ✕ eliminar.
- **Añadir servicio + reordenar por arrastrar y soltar**: La cuadrícula de servicios incorpora un botón de + añadir y un asa de arrastre (`⋮⋮`).
- **Encabezado / pie de página / animaciones de tema / selector de idioma** restaurados. Los 7 temas (cherry/dark/light/ocean/gold/autumn/winter) reproducen su efecto de partículas en una capa detrás de las tarjetas pero delante del fondo.
- **UX de cierre del panel deslizable**: Un clic fuera o la tecla Esc cierra el panel deslizable.
- **Indicador de estado SSE + temporizador de tiempo de actividad** en la barra superior (topbar), junto al selector de idioma/tema. El contador `⏱ uptime` y el indicador `● SSE` (verde = conectado, naranja = reconectando, gris = desconectado) se muestran juntos (movidos del pie de página al encabezado desde v0.2.18 — estado visible sin hacer scroll).

---

## v0.2.2 Stability & UX Improvements

- **Dispatch fast-skip**: cloud services whose keys are all on cooldown or exhausted are no longer force-retried. Dispatch moves to the next fallback immediately. Per-request tail latency dropped from ~15 s to ~1.5 s.
- **Fallback model swap**: each fallback step now applies the target service's own `default_model`. Previously a `gemini-2.5-flash` request would be handed to Anthropic/Ollama verbatim and rejected (400/404).
- **Anthropic credit-balance handling**: when Anthropic returns HTTP 400 with a "credit balance" body, the proxy promotes it to 402-equivalent and sets a 30 min cooldown so subsequent dispatches skip Anthropic automatically.
- **Service edit default_model dropdown polish**:
  - The server now renders the complete model list (Google 15, OpenRouter 345, etc.) into the `<select>` from the first open — no second round-trip required.
  - `↓ Move to Allowed` button demotes the current default into the allowed_models textarea and clears the default.
  - `✕ Clear` empties the default in place.
  - Collapsible `Custom input` details block lets you type a model ID directly when the dropdown is unreachable.
- **Agent edit/create model_override dropdown**: free text replaced by a `<select>` populated from the preferred service's `default_model` + `allowed_models`. Changing the preferred service auto-repopulates the override options.
- **ClientInput v0.2 fields**: POST `/admin/clients` now accepts v0.2 canonical `preferred_service` / `model_override` alongside legacy `default_service` / `default_model` (legacy is a fallback).

---

## Cambios Recientes (v0.1.16 ~ v0.1.27)

### v0.1.27 (2026-04-09)
- **Corrección del nombre de modelo en fallback a Ollama**: Se corrigió un problema donde los nombres de modelos con prefijo de proveedor (ej: `google/gemini-3.1-pro-preview`) se pasaban directamente a Ollama durante el fallback desde otros servicios. Ahora se reemplaza automáticamente con la variable de entorno/modelo predeterminado.
- **Duración de cooldown significativamente reducida**: 429 límite de tasa 30min->5min, 402 pago 1hora->30min, 401/403 24horas->6horas. Previene la parálisis total del proxy cuando todas las claves entran en cooldown simultáneamente.
- **Reintento forzado en cooldown total**: Cuando todas las claves están en cooldown, se reintenta forzosamente la clave más cercana a expirar para evitar el rechazo de solicitudes.
- **Corrección de visualización de lista de servicios**: La respuesta `/status` ahora muestra la lista real de servicios sincronizados desde el vault (previene la omisión de anthropic etc.).

### v0.1.25 (2026-04-08)
- **Detección de proceso de agente**: El proxy detecta si los agentes locales (NanoClaw/OpenClaw) están vivos y muestra un semáforo naranja en el panel.
- **Mejora del asa de arrastre**: La ordenación de tarjetas ahora solo permite agarrar desde el área del semáforo. Previene el arrastre accidental desde campos de entrada o botones.

### v0.1.24 (2026-04-06)
- **Subcomando RTK de ahorro de tokens**: `wall-vault rtk <command>` filtra automáticamente la salida de comandos shell, reduciendo el uso de tokens de agentes IA en 60-90%. Incluye filtros integrados para comandos principales como git y go, y trunca automáticamente comandos no soportados. Se integra de forma transparente con Claude Code vía hook `PreToolUse`.

### v0.1.23 (2026-04-06)
- **Corrección del cambio de modelo Ollama**: Se corrigió un problema donde cambiar el modelo Ollama en el panel del almacén no se reflejaba en el proxy real. Anteriormente solo se usaba la variable de entorno (`OLLAMA_MODEL`); ahora la configuración del almacén tiene prioridad.
- **Semáforo automático de servicio local**: Ollama, LM Studio y vLLM se activan automáticamente cuando son accesibles y se desactivan cuando se desconectan. Funciona de la misma manera que la alternancia automática basada en claves para servicios en la nube.

### v0.1.22 (2026-04-05)
- **Corrección de omisión del campo content vacío**: Cuando los modelos thinking (gemini-3.1-pro, o1, claude thinking, etc.) usaban todos los max_tokens para reasoning y no podían producir una respuesta real, el proxy omitía los campos `content`/`text` del JSON de respuesta vía `omitempty`, causando que los clientes SDK de OpenAI/Anthropic fallaran con `Cannot read properties of undefined (reading 'trim')`. Corregido para incluir siempre los campos según las especificaciones API oficiales.

### v0.1.21 (2026-04-05)
- **Soporte de modelo Gemma 4**: Los modelos de la serie Gemma como `gemma-4-31b-it` y `gemma-4-26b-a4b-it` ahora se pueden usar a través de la API de Google Gemini.
- **Soporte completo de servicios LM Studio / vLLM**: Anteriormente, estos servicios faltaban en el enrutamiento del proxy y siempre recurrían a Ollama. Ahora se enrutan correctamente vía API compatible con OpenAI.
- **Corrección de visualización de servicio en el panel**: Incluso durante el fallback, el panel siempre muestra el servicio configurado por el usuario.
- **Visualización de estado del servicio local**: Al cargar el panel, el estado de conexión de los servicios locales (Ollama, LM Studio, vLLM, etc.) se muestra mediante el color del punto.
- **Variable de entorno del filtro de herramientas**: Modo de paso de herramientas configurable con la variable de entorno `WV_TOOL_FILTER=passthrough`.

### v0.1.20 (2026-03-28)
- **Fortalecimiento integral de seguridad**: Prevención XSS (41 ubicaciones), comparación de tokens en tiempo constante, restricciones CORS, límites de tamaño de solicitud, prevención de traversal de ruta, autenticación SSE, fortalecimiento del limitador de tasa, y 12 otras mejoras de seguridad.

### v0.1.19 (2026-03-27)
- **Detección en línea de Claude Code**: Las instancias de Claude Code que no pasan por el proxy también se muestran como en línea en el panel.

### v0.1.18 (2026-03-26)
- **Corrección de persistencia del servicio de fallback**: Después de errores temporales que causan fallback a Ollama, retorno automático al servicio original cuando se recupera.
- **Mejora de detección offline**: Verificaciones de estado cada 15 segundos proporcionan detección más rápida de caídas del proxy.

### v0.1.17 (2026-03-25)
- **Ordenar tarjetas con arrastrar y soltar**: Las tarjetas de agentes se pueden arrastrar para reordenar.
- **Botón de aplicación de config en línea**: Los agentes offline muestran un botón [:zap: Aplicar Config].
- **Tipo de agente cokacdir agregado**.

### v0.1.16 (2026-03-25)
- **Sincronización bidireccional de modelos**: Cambiar el modelo de Cline o Claude Code en el panel del almacén se refleja automáticamente.

---

*Para información API más detallada, consulta [API.md](API.md).*
