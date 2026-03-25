# Manual de Usuario de wall-vault
*(Última actualización: 2026-03-20 — v0.1.15)*

---

## Tabla de Contenidos

1. [¿Qué es wall-vault?](#qué-es-wall-vault)
2. [Instalación](#instalación)
3. [Primeros pasos (asistente setup)](#primeros-pasos)
4. [Registro de claves API](#registro-de-claves-api)
5. [Uso del proxy](#uso-del-proxy)
6. [Panel de control del almacén de claves](#panel-de-control-del-almacén-de-claves)
7. [Modo distribuido (multi-bot)](#modo-distribuido-multi-bot)
8. [Inicio automático](#inicio-automático)
9. [Doctor: diagnóstico automático](#doctor-diagnóstico-automático)
10. [Variables de entorno](#variables-de-entorno)
11. [Solución de problemas](#solución-de-problemas)

---

## ¿Qué es wall-vault?

**wall-vault = proxy de IA para OpenClaw + almacén seguro de claves API**

Para usar servicios de IA necesitas una **clave API** (en español: "pase digital"). Una clave API es como un carnet que le dice al servicio "esta persona tiene derecho a usarme". Estas claves tienen un límite de uso diario y, si no se gestionan bien, pueden quedar expuestas.

wall-vault guarda esas claves en un almacén cifrado y actúa como **intermediario (proxy)** entre OpenClaw y los servicios de IA. En pocas palabras: OpenClaw solo necesita conectarse a wall-vault, y wall-vault se encarga de todo lo demás.

Problemas que wall-vault resuelve:

- **Rotación automática de claves**: si una clave alcanza su límite o entra en enfriamiento (cooldown), wall-vault pasa silenciosamente a la siguiente. OpenClaw no se interrumpe.
- **Cambio automático de servicio (fallback)**: si Google no responde, cambia a OpenRouter; si tampoco responde, cambia a Ollama (IA local en tu ordenador). La sesión no se corta.
- **Sincronización en tiempo real (SSE)**: si cambias el modelo en el panel de control, se refleja en OpenClaw en 1–3 segundos. SSE (Server-Sent Events) es una tecnología que permite al servidor enviar cambios al cliente en tiempo real.
- **Notificaciones en tiempo real**: eventos como claves agotadas o fallos de servicio aparecen directamente en la parte inferior de la pantalla TUI (terminal) de OpenClaw.

> 💡 También puedes conectar **Claude Code, Cursor y VS Code**, pero el propósito principal de wall-vault es usarse junto con OpenClaw.

```
OpenClaw (pantalla de terminal TUI)
        │
        ▼
  proxy wall-vault (:56244)   ← gestión de claves, enrutamiento, fallback, eventos
        │
        ├─ Google Gemini API
        ├─ OpenRouter API (más de 340 modelos)
        └─ Ollama (tu ordenador, último recurso)
```

---

## Instalación

### Linux / macOS

Abre una terminal y pega el siguiente comando tal cual.

```bash
# Linux (PC o servidor estándar — amd64)
curl -L https://github.com/sookmook/wall-vault/releases/latest/download/wall-vault-linux-amd64 \
  -o ~/.local/bin/wall-vault && chmod +x ~/.local/bin/wall-vault

# macOS Apple Silicon (Mac M1/M2/M3)
curl -L https://github.com/sookmook/wall-vault/releases/latest/download/wall-vault-darwin-arm64 \
  -o /usr/local/bin/wall-vault && chmod +x /usr/local/bin/wall-vault
```

- `curl -L ...` — descarga el archivo desde internet.
- `chmod +x` — marca el archivo descargado como "ejecutable". Si omites este paso, obtendrás un error de "permiso denegado".

### Windows

Abre PowerShell (con permisos de administrador) y ejecuta el siguiente comando.

```powershell
# Descargar
Invoke-WebRequest -Uri `
  "https://github.com/sookmook/wall-vault/releases/latest/download/wall-vault-windows-amd64.exe" `
  -OutFile "$env:LOCALAPPDATA\Programs\wall-vault\wall-vault.exe"

# Añadir al PATH (se aplica al reiniciar PowerShell)
$env:PATH += ";$env:LOCALAPPDATA\Programs\wall-vault"
```

> 💡 **¿Qué es el PATH?** Es la lista de carpetas donde el sistema operativo busca los comandos. Si añades wall-vault al PATH, podrás escribir `wall-vault` en cualquier carpeta y ejecutarlo directamente.

### Compilar desde el código fuente (para desarrolladores)

Solo es necesario si tienes instalado el entorno de desarrollo de Go.

```bash
git clone https://github.com/sookmook/wall-vault
cd wall-vault
make build       # bin/wall-vault (versión: v0.1.6.YYYYMMDD.HHmmss)
make install     # ~/.local/bin/wall-vault
```

> 💡 **Versión con sello de tiempo**: al compilar con `make build`, la versión incluye automáticamente la fecha y hora, como `v0.1.6.20260314.231308`. Si compilas directamente con `go build ./...`, la versión aparecerá simplemente como `"dev"`.

---

## Primeros pasos

### Ejecutar el asistente de configuración (setup)

La primera vez que uses wall-vault, ejecuta el **asistente de configuración** con el siguiente comando. El asistente te guiará paso a paso preguntándote cada opción necesaria.

```bash
wall-vault setup
```

Los pasos que recorre el asistente son los siguientes:

```
1. Selección de idioma (10 idiomas disponibles, incluido español)
2. Selección de tema (light / dark / gold / cherry / ocean)
3. Modo de operación — elegir entre uso individual (standalone) o en varios equipos (distributed)
4. Nombre del bot — el nombre que aparecerá en el panel de control
5. Configuración de puertos — por defecto: proxy 56244, almacén 56243 (pulsa Enter para dejarlo así)
6. Selección de servicios de IA — Google / OpenRouter / Ollama
7. Configuración del filtro de seguridad de herramientas
8. Token de administrador — contraseña para proteger las funciones de gestión del panel. Se puede generar automáticamente
9. Contraseña de cifrado de claves API — para mayor seguridad en el almacenamiento (opcional)
10. Ruta donde se guardará el archivo de configuración
```

> ⚠️ **Recuerda bien el token de administrador.** Lo necesitarás más adelante para añadir claves o cambiar la configuración desde el panel. Si lo pierdes, tendrás que editar el archivo de configuración manualmente.

Al completar el asistente se genera automáticamente el archivo de configuración `wall-vault.yaml`.

### Iniciar

```bash
wall-vault start
```

Se inician simultáneamente dos servidores:

- **Proxy** (`http://localhost:56244`) — intermediario entre OpenClaw y los servicios de IA
- **Almacén de claves** (`http://localhost:56243`) — gestión de claves API y panel de control web

Abre `http://localhost:56243` en tu navegador para acceder al panel de control.

---

## Registro de claves API

Hay cuatro formas de registrar una clave API. **Si estás empezando, te recomendamos el método 1 (variables de entorno)**.

### Método 1: Variables de entorno (recomendado — el más sencillo)

Una variable de entorno es un **valor preconfigurado** que el programa lee al iniciarse. Escribe lo siguiente en la terminal:

```bash
# Registrar clave de Google Gemini
export WV_KEY_GOOGLE=AIzaSy...

# Registrar clave de OpenRouter
export WV_KEY_OPENROUTER=sk-or-v1-...

# Iniciar tras registrar
wall-vault start
```

Si tienes varias claves, sepáralas con comas (`,`). wall-vault las usará de forma rotativa automática (round robin):

```bash
export WV_KEY_GOOGLE=AIzaSy...,AIzaSy...,AIzaSy...
```

> 💡 **Consejo**: el comando `export` solo se aplica a la sesión de terminal actual. Para que persista al reiniciar el ordenador, añade esa línea a tu archivo `~/.bashrc` o `~/.zshrc`.

### Método 2: Panel de control web (con el ratón)

1. Abre `http://localhost:56243` en el navegador
2. En la tarjeta **🔑 API Keys**, haz clic en el botón `[+ Añadir]`
3. Introduce el tipo de servicio, el valor de la clave, una etiqueta (nombre descriptivo) y el límite diario, luego guarda

### Método 3: API REST (para automatización y scripts)

La API REST es una forma de que los programas intercambien datos mediante HTTP. Es útil para registrar claves automáticamente mediante scripts.

```bash
curl -X POST http://localhost:56243/admin/keys \
  -H "Authorization: Bearer tu-token-de-administrador" \
  -H "Content-Type: application/json" \
  -d '{
    "service": "google",
    "key": "AIzaSy...",
    "label": "Clave principal",
    "daily_limit": 1000
  }'
```

### Método 4: Flag del proxy (para pruebas rápidas)

Útil cuando quieres probar una clave de forma temporal sin registrarla. Desaparece al cerrar el programa.

```bash
wall-vault proxy --key-google=AIzaSy... --key-openrouter=sk-or-...
```

---

## Uso del proxy

### Uso con OpenClaw (propósito principal)

Así se configura OpenClaw para conectarse a los servicios de IA a través de wall-vault.

Abre el archivo `~/.openclaw/openclaw.json` y añade lo siguiente:

```json5
// ~/.openclaw/openclaw.json
{
  models: {
    providers: {
      "wall-vault": {
        baseUrl: "http://localhost:56244/v1",
        apiKey: "your-agent-token",   // token del agente en el almacén
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

> 💡 **Manera más fácil**: pulsa el botón **🦞 Copiar configuración de OpenClaw** en la tarjeta del agente en el panel. Se copiará al portapapeles un fragmento ya relleno con tu token y dirección. Solo tienes que pegarlo.

**¿A dónde se dirige cada modelo según el prefijo `wall-vault/`?**

wall-vault determina automáticamente a qué servicio de IA enviar la solicitud según el nombre del modelo:

| Formato del modelo | Servicio de destino |
|-------------------|---------------------|
| `wall-vault/gemini-*` | Conexión directa a Google Gemini |
| `wall-vault/gpt-*`, `wall-vault/o3`, `wall-vault/o4*` | Conexión directa a OpenAI |
| `wall-vault/claude-*` | Anthropic a través de OpenRouter |
| `wall-vault/hunter-alpha`, `wall-vault/healer-alpha` | OpenRouter (contexto 1M de tokens gratuito) |
| `wall-vault/kimi-*`, `wall-vault/glm-*`, `wall-vault/deepseek-*` | Conexión a OpenRouter |
| `google/nombre-modelo`, `openai/nombre-modelo`, `anthropic/nombre-modelo`, etc. | Conexión directa al servicio indicado |
| `custom/google/nombre-modelo`, `custom/openai/nombre-modelo`, etc. | Elimina el prefijo `custom/` y redirige |
| `nombre-modelo:cloud` | Elimina el sufijo `:cloud` y conecta a OpenRouter |

> 💡 **¿Qué es el contexto?** Es la cantidad de conversación o texto que la IA puede "recordar" de una vez. Con 1M (un millón de tokens) se pueden procesar conversaciones muy largas o documentos extensos en una sola sesión.

### Conexión directa con formato de API Gemini (compatibilidad con herramientas existentes)

Si ya usas alguna herramienta que se conecta directamente a Google Gemini API, basta con cambiar la dirección a wall-vault:

```bash
export ANTHROPIC_BASE_URL=http://localhost:56244/google
```

O, si la herramienta permite especificar la URL directamente:

```
http://localhost:56244/google/v1beta/models/gemini-2.5-flash:generateContent
```

### Uso con el SDK de OpenAI (Python)

También puedes conectar wall-vault desde código Python que usa IA. Solo cambia `base_url`:

```python
from openai import OpenAI

client = OpenAI(
    base_url="http://localhost:56244/v1",
    api_key="not-needed"  # wall-vault gestiona las claves API automáticamente
)

response = client.chat.completions.create(
    model="google/gemini-2.5-flash",   # formato proveedor/modelo
    messages=[{"role": "user", "content": "Hola"}]
)
```

### Cambiar el modelo en tiempo de ejecución

Si wall-vault ya está en ejecución y quieres cambiar el modelo de IA que utiliza:

```bash
# Cambiar el modelo enviando una solicitud directamente al proxy
curl -X PUT http://localhost:56244/api/config/model \
  -H "Content-Type: application/json" \
  -d '{"service": "openrouter", "model": "anthropic/claude-3.5-sonnet"}'

# En modo distribuido (multi-bot), cámbialo desde el almacén → se aplica al instante vía SSE
curl -X PUT http://localhost:56243/admin/clients/id-de-mi-bot \
  -H "Authorization: Bearer tu-token-de-administrador" \
  -H "Content-Type: application/json" \
  -d '{"default_service": "google", "default_model": "gemini-2.5-pro"}'
```

### Consultar la lista de modelos disponibles

```bash
# Ver la lista completa
curl http://localhost:56244/api/models | python3 -m json.tool

# Ver solo los modelos de Google
curl "http://localhost:56244/api/models?service=google"

# Buscar por nombre (ejemplo: modelos que contienen "claude")
curl "http://localhost:56244/api/models?q=claude"
```

**Resumen de modelos principales por servicio:**

| Servicio | Modelos principales |
|----------|---------------------|
| Google | gemini-2.5-pro, gemini-2.5-flash, gemini-2.5-flash-8b, gemini-2.0-flash |
| OpenAI | gpt-4o, gpt-4o-mini, o3, o1, o1-mini |
| OpenRouter | Más de 346 modelos (Hunter Alpha 1M contexto gratis, DeepSeek R1/V3, Qwen 2.5, etc.) |
| Ollama | Detecta automáticamente el servidor local instalado en tu ordenador |

---

## Panel de control del almacén de claves

Accede al panel visitando `http://localhost:56243` en tu navegador.

**Estructura de la pantalla:**
- **Barra superior fija (topbar)**: logo, selector de idioma y tema, indicador de estado de conexión SSE
- **Cuadrícula de tarjetas**: tarjetas de agentes, servicios y claves API dispuestas en forma de mosaico

### Tarjeta de claves API

Una tarjeta que te permite gestionar de un vistazo todas las claves API registradas.

- Muestra la lista de claves agrupadas por servicio.
- `today_usage`: número de tokens procesados con éxito hoy (texto leído y generado por la IA)
- `today_attempts`: número total de llamadas realizadas hoy (éxitos + fallos)
- El botón `[+ Añadir]` registra una nueva clave; `✕` la elimina.

> 💡 **¿Qué es un token?** Es la unidad que usa la IA para procesar texto. Corresponde aproximadamente a una palabra en inglés, o a uno o dos caracteres en otros idiomas. El coste de la API suele calcularse en función del número de tokens.

### Tarjeta de agentes

Una tarjeta que muestra el estado de los bots (agentes) conectados al proxy de wall-vault.

**El estado de conexión se muestra en 4 niveles:**

| Indicador | Estado | Significado |
|-----------|--------|-------------|
| 🟢 | En ejecución | El proxy funciona con normalidad |
| 🟡 | Con retraso | Responde pero lentamente |
| 🔴 | Sin conexión | El proxy no responde |
| ⚫ | Sin conectar / inactivo | El proxy nunca se ha conectado al almacén, o está desactivado |

**Guía de botones en la parte inferior de la tarjeta del agente:**

Cuando registras un agente, si especificas su **tipo**, aparecerán automáticamente los botones de acceso rápido correspondientes.

---

#### 🔘 Botón "Copiar configuración" — genera la configuración de conexión automáticamente

Al hacer clic, se copia al portapapeles un fragmento de configuración ya relleno con el token, la dirección del proxy y los datos del modelo de ese agente. Solo tienes que pegarlo en la ubicación indicada en la tabla siguiente para completar la configuración.

| Botón | Tipo de agente | Dónde pegar |
|-------|---------------|-------------|
| 🦞 Copiar configuración de OpenClaw | `openclaw` | `~/.openclaw/openclaw.json` |
| 🦀 Copiar configuración de NanoClaw | `nanoclaw` | `~/.openclaw/openclaw.json` |
| 🟠 Copiar configuración de Claude Code | `claude-code` | `~/.claude/settings.json` |
| ⌨ Copiar configuración de Cursor | `cursor` | Cursor → Settings → AI |
| 💻 Copiar configuración de VSCode | `vscode` | `~/.continue/config.json` |

**Ejemplo — si el tipo es Claude Code, se copia algo así:**

```json
// ~/.claude/settings.json
{
  "apiProvider": "openai",
  "baseUrl": "http://192.168.1.20:56244/v1",
  "apiKey": "token-de-este-agente"
}
```

**Ejemplo — si el tipo es VSCode (Continue):**

```json
// ~/.continue/config.json
{
  "models": [{
    "title": "wall-vault proxy",
    "provider": "openai",
    "model": "gemini-2.5-flash",
    "apiBase": "http://192.168.1.20:56244/v1",
    "apiKey": "token-de-este-agente"
  }]
}
```

**Ejemplo — si el tipo es Cursor:**

```
Base URL : http://192.168.1.20:56244/v1
API Key  : token-de-este-agente

// O como variables de entorno:
OPENAI_BASE_URL=http://192.168.1.20:56244/v1
OPENAI_API_KEY=token-de-este-agente
```

> ⚠️ **Si el portapapeles no funciona**: la política de seguridad del navegador puede bloquear la copia. Si aparece un cuadro de texto emergente, selecciona todo con Ctrl+A y copia con Ctrl+C.

---

#### ⚡ Botón de aplicación automática — configuración completada con un solo clic

Cuando el tipo de agente es `cline`, `claude-code`, `openclaw` o `nanoclaw`, aparece un botón **⚡ Aplicar configuración** en la tarjeta del agente. Al pulsarlo, el archivo de configuración local del agente correspondiente se actualiza automáticamente.

| Botón | Tipo de agente | Archivo de destino |
|-------|---------------|-------------------|
| ⚡ Aplicar configuración de Cline | `cline` | `~/.cline/data/globalState.json` + `secrets.json` |
| ⚡ Aplicar configuración de Claude Code | `claude-code` | `~/.claude/settings.json` |
| ⚡ Aplicar configuración de OpenClaw | `openclaw` | `~/.openclaw/openclaw.json` |
| ⚡ Aplicar configuración de NanoClaw | `nanoclaw` | `~/.openclaw/openclaw.json` |

> ⚠️ Este botón envía una solicitud a **localhost:56244** (el proxy local). Solo funciona si el proxy está en ejecución en esa máquina.

---

#### 🔀 Ordenar tarjetas con arrastrar y soltar (v0.1.17)

Puedes **arrastrar** las tarjetas de agentes en el panel para reorganizarlas en el orden que desees.

1. Toma una tarjeta de agente con el ratón y arrástrala
2. Suéltala sobre otra tarjeta para intercambiar sus posiciones
3. El nuevo orden se **guarda inmediatamente en el servidor** y se mantiene después de actualizar la página

> 💡 Los dispositivos táctiles (móvil/tableta) aún no son compatibles. Por favor, usa un navegador de escritorio.

---

#### 🔄 Sincronización bidireccional de modelos (v0.1.16)

Cuando cambias el modelo de un agente en el panel de la bóveda, la configuración local de ese agente se actualiza automáticamente.

**En el caso de Cline:**
- Cambio de modelo en la bóveda → evento SSE → el proxy actualiza los campos de modelo en `globalState.json`
- Campos actualizados: `actModeOpenAiModelId`, `planModeOpenAiModelId`, `openAiModelId`
- No se modifican `openAiBaseUrl` ni la clave API
- **Es necesario recargar VS Code (`Ctrl+Alt+R` o `Ctrl+Shift+P` → `Developer: Reload Window`)**
  - Cline no relee el archivo de configuración mientras está en ejecución

**En el caso de Claude Code:**
- Cambio de modelo en la bóveda → evento SSE → el proxy actualiza el campo `model` en `settings.json`
- Busca automáticamente en las rutas de WSL y Windows (`~/.claude/`, `/mnt/c/Users/*/.claude/`)

**En dirección inversa (agente → bóveda):**
- Cuando un agente (Cline, Claude Code, etc.) envía una solicitud al proxy, este incluye la información de servicio y modelo del cliente en el heartbeat
- La tarjeta del agente en el panel de la bóveda muestra en tiempo real el servicio/modelo en uso

> 💡 **Punto clave**: el proxy identifica al agente mediante el token Authorization de la solicitud y lo enruta automáticamente al servicio/modelo configurado en la bóveda. Aunque Cline o Claude Code envíen un nombre de modelo diferente, el proxy lo sobrescribe con la configuración de la bóveda.

---

### Usar Cline en VS Code — guía detallada

#### Paso 1: Instalar Cline

Instala **Cline** (ID: `saoudrizwan.claude-dev`) desde el marketplace de extensiones de VS Code.

#### Paso 2: Registrar el agente en la bóveda

1. Abre el panel de la bóveda (`http://IP-de-la-bóveda:56243`)
2. En la sección **Agentes**, haz clic en **+ Añadir**
3. Rellena los campos así:

| Campo | Valor | Descripción |
|-------|-------|-------------|
| ID | `my_cline` | Identificador único (alfanumérico, sin espacios) |
| Nombre | `My Cline` | Nombre que aparecerá en el panel |
| Tipo de agente | `cline` | ← Debes seleccionar `cline` obligatoriamente |
| Servicio | Selecciona el servicio que quieras usar (ej.: `google`) | |
| Modelo | Introduce el modelo que quieras usar (ej.: `gemini-2.5-flash`) | |

4. Al pulsar **Guardar**, se genera automáticamente un token

#### Paso 3: Conectar con Cline

**Opción A — Aplicación automática (recomendada)**

1. Comprueba que el **proxy** de wall-vault esté en ejecución en esa máquina (`localhost:56244`)
2. En la tarjeta del agente del panel, haz clic en **⚡ Aplicar configuración de Cline**
3. Si aparece la notificación "¡Configuración aplicada!", ha funcionado
4. Recarga VS Code (`Ctrl+Alt+R`)

**Opción B — Configuración manual**

Abre la configuración (⚙️) en la barra lateral de Cline:
- **API Provider**: `OpenAI Compatible`
- **Base URL**: `http://dirección-del-proxy:56244/v1`
  - En la misma máquina: `http://localhost:56244/v1`
  - En otra máquina (como un Mini server): `http://192.168.1.20:56244/v1`
- **API Key**: el token emitido en la bóveda (cópialo desde la tarjeta del agente)
- **Model ID**: el modelo configurado en la bóveda (ej.: `gemini-2.5-flash`)

#### Paso 4: Verificación

Envía cualquier mensaje en la ventana de chat de Cline. Si todo funciona correctamente:
- En el panel de la bóveda, la tarjeta del agente mostrará un **punto verde (● en ejecución)**
- La tarjeta mostrará el servicio/modelo actual (ej.: `google / gemini-2.5-flash`)

#### Cambiar el modelo

Cuando quieras cambiar el modelo de Cline, hazlo desde el **panel de la bóveda**:

1. Cambia el desplegable de servicio/modelo en la tarjeta del agente
2. Haz clic en **Aplicar**
3. Recarga VS Code (`Ctrl+Alt+R`) — el nombre del modelo en el pie de Cline se actualizará
4. A partir de la siguiente solicitud se usará el nuevo modelo

> 💡 En realidad, el proxy identifica las solicitudes de Cline por su token y las enruta al modelo de la configuración de la bóveda. Aunque no recargues VS Code, **el modelo que realmente se usa cambia de inmediato** — la recarga solo sirve para actualizar el nombre del modelo que muestra la interfaz de Cline.

#### Detección de desconexión

Al cerrar VS Code, la tarjeta del agente en el panel de la bóveda cambiará a amarillo (con retraso) tras unos **2–3 minutos**, y a rojo (sin conexión) tras **5 minutos**.

#### Solución de problemas

| Síntoma | Causa | Solución |
|---------|-------|----------|
| Error de "conexión fallida" en Cline | Proxy no ejecutándose o dirección incorrecta | Comprueba el proxy con `curl http://localhost:56244/health` |
| No aparece el punto verde en la bóveda | Clave API (token) no configurada | Haz clic de nuevo en **⚡ Aplicar configuración de Cline** |
| El modelo en el pie de Cline no cambia | Cline tiene la configuración en caché | Recarga VS Code (`Ctrl+Alt+R`) |
| Se muestra un nombre de modelo incorrecto | Bug antiguo (corregido en v0.1.16) | Actualiza el proxy a v0.1.16 o superior |

---

#### 🟣 Botón "Copiar comando de despliegue" — para instalar en una nueva máquina

Se usa cuando quieres instalar el proxy de wall-vault en un nuevo ordenador y conectarlo al almacén. Al hacer clic, se copia el script de instalación completo. Pégalo en la terminal del nuevo ordenador y ejecútalo; se procesará todo en un solo paso:

1. Instalar el binario de wall-vault (se omite si ya está instalado)
2. Registrar automáticamente el servicio de usuario en systemd
3. Iniciar el servicio y conectarlo automáticamente al almacén

> 💡 El script ya incluye el token de este agente y la dirección del servidor del almacén, por lo que puedes ejecutarlo directamente sin modificar nada.

---

### Tarjeta de servicios

Una tarjeta para activar o desactivar los servicios de IA y configurar sus opciones.

- Interruptores de activación/desactivación por servicio
- Si introduces la dirección de un servidor de IA local (Ollama, LM Studio, vLLM, etc. corriendo en tu ordenador), encontrará automáticamente los modelos disponibles.
- **Indicador de estado del servicio local**: el punto ● junto al nombre del servicio es **verde** si está conectado, **gris** si no lo está.
- **Sincronización automática de casillas**: si un servicio local (como Ollama) está en ejecución al abrir la página, la casilla se marcará automáticamente.

> 💡 **Si el servicio local corre en otro ordenador**: introduce la IP de ese ordenador en el campo de URL del servicio. Por ejemplo: `http://192.168.1.20:11434` (Ollama), `http://192.168.1.20:1234` (LM Studio)

### Introducción del token de administrador

Cuando intentas usar funciones importantes desde el panel, como añadir o eliminar claves, aparecerá una ventana emergente pidiendo el token de administrador. Introduce el token que configuraste con el asistente setup. Una vez introducido, permanece activo hasta que cierres el navegador.

> ⚠️ **Si introduces el token de forma incorrecta más de 10 veces en 15 minutos, esa IP quedará bloqueada temporalmente.** Si olvidaste el token, búscalo en la entrada `admin_token` del archivo `wall-vault.yaml`.

---

## Modo distribuido (multi-bot)

Cuando quieres ejecutar OpenClaw simultáneamente en varios ordenadores, puedes configurarlos para que **compartan un único almacén de claves**. Así solo necesitas gestionar las claves en un lugar.

### Ejemplo de configuración

```
[Servidor del almacén de claves]
  wall-vault vault    (almacén :56243, panel de control)

[WSL Alpha]              [Raspberry Pi Gamma]    [Mac Mini Local]
  wall-vault proxy         wall-vault proxy          wall-vault proxy
  openclaw TUI             openclaw TUI              openclaw TUI
  ↕ sincronización SSE     ↕ sincronización SSE      ↕ sincronización SSE
```

Todos los bots apuntan al servidor del almacén central, de modo que si cambias el modelo o añades una clave en el almacén, se aplica inmediatamente a todos los bots.

### Paso 1: Iniciar el servidor del almacén de claves

Ejecuta esto en el ordenador que actuará como servidor del almacén:

```bash
wall-vault vault
```

### Paso 2: Registrar cada bot (cliente)

Registra de antemano la información de cada bot que se conectará al servidor del almacén:

```bash
curl -X POST http://localhost:56243/admin/clients \
  -H "Authorization: Bearer tu-token-de-administrador" \
  -H "Content-Type: application/json" \
  -d '{
    "id": "botA",
    "name": "Bot A",
    "token": "bota-secret",
    "default_service": "google",
    "default_model": "gemini-2.5-flash"
  }'
```

### Paso 3: Iniciar el proxy en cada ordenador bot

En cada ordenador bot, inicia el proxy especificando la dirección y el token del servidor del almacén:

```bash
WV_VAULT_URL=http://192.168.x.x:56243 \
WV_VAULT_TOKEN=bota-secret \
WV_VAULT_CLIENT_ID=botA \
wall-vault proxy
```

> 💡 Sustituye **`192.168.x.x`** por la IP interna real del ordenador que hace de servidor del almacén. Puedes consultarla en la configuración de tu router o con el comando `ip addr`.

---

## Inicio automático

Si te resulta tedioso iniciar wall-vault manualmente cada vez que reinicias el ordenador, regístralo como servicio del sistema. Una vez registrado, se iniciará automáticamente al arrancar.

### Linux — systemd (la mayoría de distribuciones Linux)

systemd es el sistema de Linux que inicia y gestiona programas automáticamente:

```bash
wall-vault doctor deploy
systemctl --user daemon-reload
systemctl --user enable --now wall-vault
```

Ver los registros (logs):

```bash
journalctl --user -u wall-vault -f
```

### macOS — launchd

El sistema de macOS encargado de la ejecución automática de programas:

```bash
wall-vault doctor deploy launchd
launchctl load ~/Library/LaunchAgents/com.wall-vault.plist
```

### Windows — NSSM

1. Descarga NSSM desde [nssm.cc](https://nssm.cc/download) y añádelo al PATH.
2. En PowerShell con permisos de administrador:

```powershell
wall-vault doctor deploy windows
```

---

## Doctor: diagnóstico automático

El comando `doctor` es una **herramienta de autodiagnóstico y reparación** que comprueba si wall-vault está configurado correctamente.

```bash
wall-vault doctor check   # Diagnostica el estado actual (solo lectura, no modifica nada)
wall-vault doctor fix     # Repara los problemas automáticamente
wall-vault doctor all     # Diagnóstico + reparación automática en un solo paso
```

> 💡 Si algo parece ir mal, empieza ejecutando `wall-vault doctor all`. Resuelve muchos problemas de forma automática.

---

## Variables de entorno

Las variables de entorno son una forma de pasar valores de configuración a un programa. Puedes introducirlas en la terminal con el formato `export NOMBRE_VARIABLE=valor`, o incluirlas en el archivo del servicio de inicio automático para que siempre estén activas.

| Variable | Descripción | Valor de ejemplo |
|----------|-------------|-----------------|
| `WV_LANG` | Idioma del panel de control | `ko`, `en`, `ja` |
| `WV_THEME` | Tema del panel de control | `light`, `dark`, `gold` |
| `WV_KEY_GOOGLE` | Clave(s) API de Google (separadas por comas) | `AIza...,AIza...` |
| `WV_KEY_OPENROUTER` | Clave API de OpenRouter | `sk-or-v1-...` |
| `WV_VAULT_URL` | Dirección del servidor del almacén en modo distribuido | `http://192.168.x.x:56243` |
| `WV_VAULT_TOKEN` | Token de autenticación del cliente (bot) | `my-secret-token` |
| `WV_ADMIN_TOKEN` | Token de administrador | `admin-token-here` |
| `WV_MASTER_PASS` | Contraseña de cifrado de las claves API | `my-password` |
| `WV_AVATAR` | Ruta del archivo de imagen de avatar (relativa a `~/.openclaw/`) | `workspace/avatars/avatar.png` |
| `OLLAMA_URL` | Dirección del servidor local de Ollama | `http://192.168.x.x:11434` |

---

## Solución de problemas

### El proxy no arranca

Es frecuente que el puerto ya esté siendo usado por otro programa.

```bash
ss -tlnp | grep 56244   # Comprueba qué proceso está usando el puerto 56244
wall-vault proxy --port 8080   # Inicia en un puerto diferente
```

### Errores de clave API (429, 402, 401, 403, 582)

| Código de error | Significado | Cómo resolverlo |
|----------------|-------------|-----------------|
| **429** | Demasiadas solicitudes (límite de uso superado) | Espera un momento o añade otra clave |
| **402** | Pago requerido o créditos insuficientes | Recarga créditos en el servicio correspondiente |
| **401 / 403** | Clave incorrecta o sin permisos | Verifica el valor de la clave y vuelve a registrarla |
| **582** | Sobrecarga en la pasarela (cooldown de 5 minutos) | Se libera automáticamente tras 5 minutos |

```bash
# Ver la lista de claves registradas y su estado
curl -H "Authorization: Bearer tu-token-de-administrador" http://localhost:56243/admin/keys

# Reiniciar el contador de uso de las claves
curl -X POST -H "Authorization: Bearer tu-token-de-administrador" http://localhost:56243/admin/keys/reset
```

### El agente aparece como "sin conectar"

"Sin conectar" significa que el proceso del proxy no está enviando señales (heartbeat) al almacén. **No significa que la configuración se haya perdido.** El proxy debe conocer la dirección del servidor del almacén y el token para aparecer como conectado.

```bash
# Inicia el proxy especificando la dirección del almacén, el token y el ID de cliente
WV_VAULT_URL=http://direccion-del-almacen:56243 \
WV_VAULT_TOKEN=token-del-cliente \
WV_VAULT_CLIENT_ID=id-del-cliente \
wall-vault proxy
```

Si la conexión tiene éxito, en unos 20 segundos el panel mostrará el estado 🟢 en ejecución.

### Ollama no se conecta

Ollama es un programa que ejecuta IA directamente en tu ordenador. Primero comprueba si Ollama está encendido.

```bash
curl http://localhost:11434/api/tags   # Si aparece la lista de modelos, todo va bien
export OLLAMA_URL=http://192.168.x.x:11434   # Si Ollama corre en otro ordenador
```

> ⚠️ Si Ollama no responde, inícialo primero con el comando `ollama serve`.

> ⚠️ **Los modelos grandes son lentos**: modelos como `qwen3.5:35b` o `deepseek-r1` pueden tardar varios minutos en generar una respuesta. Si parece que no hay respuesta, puede que el proceso esté funcionando con normalidad; ten paciencia.

---

*Para información más detallada sobre la API, consulta [API.md](API.md).*
