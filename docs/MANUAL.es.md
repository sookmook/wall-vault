# Manual de Usuario de wall-vault
*(Ultima actualizacion: 2026-04-06 — v0.1.23)*

---

## Tabla de Contenidos

1. [Que es wall-vault?](#que-es-wall-vault)
2. [Instalacion](#instalacion)
3. [Primeros pasos (asistente setup)](#primeros-pasos)
4. [Registro de claves API](#registro-de-claves-api)
5. [Uso del proxy](#uso-del-proxy)
6. [Panel de control del almacen de claves](#panel-de-control-del-almacen-de-claves)
7. [Modo distribuido (multi-bot)](#modo-distribuido-multi-bot)
8. [Inicio automatico](#inicio-automatico)
9. [Doctor: diagnostico automatico](#doctor-diagnostico-automatico)
10. [Variables de entorno](#variables-de-entorno)
11. [Solucion de problemas](#solucion-de-problemas)

---

## Que es wall-vault?

**wall-vault = proxy de IA para OpenClaw + almacen seguro de claves API**

Para usar servicios de IA necesitas una **clave API** (un "pase digital"). Una clave API es como un carnet que le dice al servicio "esta persona tiene derecho a usarme". Estas claves tienen un limite de uso diario y, si no se gestionan bien, pueden quedar expuestas.

wall-vault guarda esas claves en un almacen cifrado y actua como **intermediario (proxy)** entre OpenClaw y los servicios de IA. En pocas palabras: OpenClaw solo necesita conectarse a wall-vault, y wall-vault se encarga de todo lo demas.

Problemas que wall-vault resuelve:

- **Rotacion automatica de claves**: si una clave alcanza su limite o entra en enfriamiento (cooldown), wall-vault pasa silenciosamente a la siguiente. OpenClaw no se interrumpe.
- **Cambio automatico de servicio (fallback)**: si Google no responde, cambia a OpenRouter; si tampoco responde, cambia a Ollama, LM Studio o vLLM (IA local en tu ordenador). La sesion no se corta. Cuando el servicio original se recupera, vuelve a el automaticamente a partir de la siguiente solicitud (v0.1.18+, LM Studio/vLLM: v0.1.21+).
- **Sincronizacion en tiempo real (SSE)**: si cambias el modelo en el panel de control, se refleja en OpenClaw en 1-3 segundos. SSE (Server-Sent Events) es una tecnologia que permite al servidor enviar cambios al cliente en tiempo real.
- **Notificaciones en tiempo real**: eventos como claves agotadas o fallos de servicio aparecen directamente en la parte inferior de la pantalla TUI (terminal) de OpenClaw.

> 💡 Tambien puedes conectar **Claude Code, Cursor y VS Code**, pero el proposito principal de wall-vault es usarse junto con OpenClaw.

```
OpenClaw (pantalla de terminal TUI)
        │
        ▼
  proxy wall-vault (:56244)   ← gestion de claves, enrutamiento, fallback, eventos
        │
        ├─ Google Gemini API
        ├─ OpenRouter API (mas de 340 modelos)
        ├─ Ollama / LM Studio / vLLM (tu ordenador, ultimo recurso)
        └─ OpenAI / Anthropic API
```

---

## Instalacion

### Linux / macOS

Abre un terminal y pega los siguientes comandos:

```bash
# Linux (PC estandar, servidor — amd64)
curl -L https://github.com/sookmook/wall-vault/releases/latest/download/wall-vault-linux-amd64 \
  -o ~/.local/bin/wall-vault && chmod +x ~/.local/bin/wall-vault

# macOS Apple Silicon (M1/M2/M3 Mac)
curl -L https://github.com/sookmook/wall-vault/releases/latest/download/wall-vault-darwin-arm64 \
  -o /usr/local/bin/wall-vault && chmod +x /usr/local/bin/wall-vault
```

- `curl -L ...` — Descarga un archivo de Internet.
- `chmod +x` — Hace que el archivo descargado sea ejecutable. Si omites este paso, obtendras un error de "permiso denegado".

### Windows

Abre PowerShell (como administrador) y ejecuta los siguientes comandos:

```powershell
# Descargar
Invoke-WebRequest -Uri `
  "https://github.com/sookmook/wall-vault/releases/latest/download/wall-vault-windows-amd64.exe" `
  -OutFile "$env:LOCALAPPDATA\Programs\wall-vault\wall-vault.exe"

# Agregar al PATH (se aplica tras reiniciar PowerShell)
$env:PATH += ";$env:LOCALAPPDATA\Programs\wall-vault"
```

> 💡 **Que es el PATH?** Es la lista de carpetas donde tu ordenador busca los comandos. Al agregar wall-vault al PATH, puedes ejecutar `wall-vault` desde cualquier directorio.

### Compilar desde el codigo fuente (para desarrolladores)

Solo aplica si tienes un entorno de desarrollo Go instalado.

```bash
git clone https://github.com/sookmook/wall-vault
cd wall-vault
make build       # bin/wall-vault (version: v0.1.23.YYYYMMDD.HHmmss)
make install     # ~/.local/bin/wall-vault
```

> 💡 **Version con marca de tiempo**: al compilar con `make build`, la version se genera automaticamente en formato `v0.1.23.20260406.211004` que incluye fecha y hora. Si compilas directamente con `go build ./...`, la version mostrara simplemente `"dev"`.

---

## Primeros pasos

### Ejecutar el asistente de configuracion

Despues de la instalacion, asegurate de ejecutar primero el **asistente de configuracion**. El asistente te guiara paso a paso, preguntandote la informacion necesaria.

```bash
wall-vault setup
```

Estos son los pasos que sigue el asistente:

```
1. Seleccion de idioma (10 idiomas incluido el espanol)
2. Seleccion de tema (light / dark / gold / cherry / ocean)
3. Modo de operacion — individual (standalone) o multiples maquinas (distributed)
4. Nombre del bot — el nombre que se mostrara en el panel de control
5. Configuracion de puertos — por defecto: proxy 56244, almacen 56243 (pulsa Enter para mantener)
6. Seleccion de servicios IA — Google / OpenRouter / Ollama / LM Studio / vLLM
7. Configuracion del filtro de seguridad de herramientas
8. Token de administrador — una contrasena para bloquear las funciones de administracion; puede generarse automaticamente
9. Contrasena de cifrado de claves API — para almacenamiento extra seguro (opcional)
10. Ubicacion del archivo de configuracion
```

> ⚠️ **Asegurate de recordar tu token de administrador.** Lo necesitaras mas adelante para agregar claves o cambiar configuraciones en el panel de control. Si lo olvidas, tendras que editar manualmente el archivo de configuracion.

Al completar el asistente, se crea automaticamente un archivo de configuracion `wall-vault.yaml`.

### Iniciar

```bash
wall-vault start
```

Se inician dos servidores simultaneamente:

- **Proxy** (`http://localhost:56244`) — el intermediario que conecta OpenClaw con los servicios de IA
- **Almacen de claves** (`http://localhost:56243`) — gestion de claves API y panel de control web

Abre `http://localhost:56243` en tu navegador para ver el panel de control inmediatamente.

---

## Registro de claves API

Hay cuatro formas de registrar claves API. **Para principiantes, se recomienda el metodo 1 (variables de entorno).**

### Metodo 1: Variables de entorno (recomendado — el mas simple)

Las variables de entorno son **valores preconfigurados** que un programa lee al iniciarse. Simplemente escribe lo siguiente en tu terminal:

```bash
# Registrar una clave de Google Gemini
export WV_KEY_GOOGLE=AIzaSy...

# Registrar una clave de OpenRouter
export WV_KEY_OPENROUTER=sk-or-v1-...

# Iniciar tras el registro
wall-vault start
```

Si tienes multiples claves, separalas con comas. wall-vault las rotara automaticamente (round-robin):

```bash
export WV_KEY_GOOGLE=AIzaSy...,AIzaSy...,AIzaSy...
```

> 💡 **Consejo**: el comando `export` solo se aplica a la sesion de terminal actual. Para que persista tras reinicios, agrega la linea a tu archivo `~/.bashrc` o `~/.zshrc`.

### Metodo 2: Interfaz del panel de control (clic de raton)

1. Abre `http://localhost:56243` en tu navegador
2. Haz clic en el boton `[+ Agregar]` en la tarjeta **🔑 Claves API** superior
3. Introduce el tipo de servicio, el valor de la clave, una etiqueta (nombre de nota) y el limite diario, luego guarda

### Metodo 3: API REST (para automatizacion/scripts)

La API REST es una forma en que los programas intercambian datos via HTTP. Util para registro automatizado mediante scripts.

```bash
curl -X POST http://localhost:56243/admin/keys \
  -H "Authorization: Bearer tu-token-admin" \
  -H "Content-Type: application/json" \
  -d '{
    "service": "google",
    "key": "AIzaSy...",
    "label": "Clave principal",
    "daily_limit": 1000
  }'
```

### Metodo 4: Flags del proxy (para pruebas rapidas)

Sirve para inyectar temporalmente una clave para pruebas sin registro formal. La clave desaparece cuando se detiene el programa.

```bash
wall-vault proxy --key-google=AIzaSy... --key-openrouter=sk-or-...
```

---

## Uso del proxy

### Uso con OpenClaw (proposito principal)

Asi es como se configura OpenClaw para conectarse a los servicios de IA a traves de wall-vault.

Abre el archivo `~/.openclaw/openclaw.json` y agrega lo siguiente:

```json5
// ~/.openclaw/openclaw.json
{
  models: {
    providers: {
      "wall-vault": {
        baseUrl: "http://localhost:56244/v1",
        apiKey: "tu-token-agente",   // token del agente del almacen
        api: "openai-completions",
        models: [
          { id: "wall-vault/gemini-2.5-flash" },
          { id: "wall-vault/gemini-2.5-pro" },
          { id: "wall-vault/hunter-alpha" },    // 1M de contexto gratuito
          { id: "wall-vault/claude-opus-4-6" }
        ]
      }
    }
  }
}
```

> 💡 **Metodo mas facil**: haz clic en el boton **🦞 Copiar config OpenClaw** en la tarjeta de agente del panel de control — copia un fragmento con el token y la direccion ya completados. Solo pega.

**A donde se dirige el prefijo `wall-vault/` en los nombres de modelo?**

wall-vault determina automaticamente a que servicio de IA enviar la solicitud segun el nombre del modelo:

| Formato del modelo | Se dirige a |
|-------------------|-------------|
| `wall-vault/gemini-*` | Google Gemini directo |
| `wall-vault/gpt-*`, `wall-vault/o3`, `wall-vault/o4*` | OpenAI directo |
| `wall-vault/claude-*` | Anthropic via OpenRouter |
| `wall-vault/hunter-alpha`, `wall-vault/healer-alpha` | OpenRouter (contexto gratuito de 1M tokens) |
| `wall-vault/kimi-*`, `wall-vault/glm-*`, `wall-vault/deepseek-*` | OpenRouter |
| `google/nombre-modelo`, `openai/nombre-modelo`, `anthropic/nombre-modelo`, etc. | Directamente al servicio correspondiente |
| `custom/google/nombre-modelo`, `custom/openai/nombre-modelo`, etc. | Elimina el prefijo `custom/` y reenviar |
| `nombre-modelo:cloud` | Elimina el sufijo `:cloud` y envia a OpenRouter |

> 💡 **Que es el contexto?** Es la cantidad de conversacion que una IA puede recordar de una vez. 1M (un millon de tokens) significa que puede procesar conversaciones o documentos muy largos en una sola sesion.

### Conexion directa en formato Gemini API (compatibilidad con herramientas existentes)

Si tienes herramientas que ya usan directamente la API de Google Gemini, simplemente cambia la direccion a wall-vault:

```bash
export ANTHROPIC_BASE_URL=http://localhost:56244/google
```

O si la herramienta acepta una URL directa:

```
http://localhost:56244/google/v1beta/models/gemini-2.5-flash:generateContent
```

### Uso con el SDK de OpenAI (Python)

Tambien puedes conectar wall-vault a codigo Python que use IA. Solo cambia el `base_url`:

```python
from openai import OpenAI

client = OpenAI(
    base_url="http://localhost:56244/v1",
    api_key="not-needed"  # wall-vault gestiona las claves API por ti
)

response = client.chat.completions.create(
    model="google/gemini-2.5-flash",   # formato provider/model
    messages=[{"role": "user", "content": "Hola"}]
)
```

### Cambiar de modelo en ejecucion

Para cambiar el modelo de IA mientras wall-vault esta en ejecucion:

```bash
# Cambiar modelo enviando una solicitud al proxy
curl -X PUT http://localhost:56244/api/config/model \
  -H "Content-Type: application/json" \
  -d '{"service": "openrouter", "model": "anthropic/claude-3.5-sonnet"}'

# En modo distribuido (multi-bot), cambiar en el servidor del almacen → sincronizacion SSE instantanea
curl -X PUT http://localhost:56243/admin/clients/mi-bot-id \
  -H "Authorization: Bearer tu-token-admin" \
  -H "Content-Type: application/json" \
  -d '{"default_service": "google", "default_model": "gemini-2.5-pro"}'
```

### Consultar la lista de modelos disponibles

```bash
# Ver la lista completa
curl http://localhost:56244/api/models | python3 -m json.tool

# Ver solo modelos de Google
curl "http://localhost:56244/api/models?service=google"

# Buscar por nombre (ej: modelos que contengan "claude")
curl "http://localhost:56244/api/models?q=claude"
```

**Modelos principales por servicio:**

| Servicio | Modelos principales |
|----------|-------------------|
| Google | gemini-2.5-pro, gemini-2.5-flash, gemini-2.5-flash-8b, gemini-2.0-flash |
| OpenAI | gpt-4o, gpt-4o-mini, o3, o1, o1-mini |
| OpenRouter | 346+ (Hunter Alpha 1M contexto gratuito, DeepSeek R1/V3, Qwen 2.5, etc.) |
| Ollama | Deteccion automatica de modelos instalados localmente |
| LM Studio | Servidor local (puerto 1234) |
| vLLM | Servidor local (puerto 8000) |

---

## Panel de control del almacen de claves

Abre `http://localhost:56243` en tu navegador para ver el panel de control.

**Disposicion:**
- **Barra superior (fija)**: logo, selectores de idioma/tema, indicador de conexion SSE
- **Cuadricula de tarjetas**: tarjetas de agentes, servicios y claves API dispuestas en mosaico

### Tarjetas de Claves API

Tarjetas que te permiten gestionar tus claves API registradas de un vistazo.

- Las claves se organizan por servicio.
- `today_usage`: numero de tokens (unidades de texto que la IA lee/escribe) procesados con exito hoy
- `today_attempts`: numero total de llamadas hoy (exitosas + fallidas)
- Usa el boton `[+ Agregar]` para registrar nuevas claves y `✕` para eliminarlas.

> 💡 **Que es un token?** Es la unidad que usa la IA para procesar texto. Aproximadamente una palabra en ingles, o 1-2 caracteres en espanol. Los precios de API generalmente se calculan segun el numero de tokens.

### Tarjetas de Agentes

Tarjetas que muestran el estado de los bots (agentes) conectados al proxy wall-vault.

**El estado de conexion se muestra en 4 niveles:**

| Indicador | Estado | Significado |
|-----------|--------|-------------|
| 🟢 | En ejecucion | El proxy funciona normalmente |
| 🟡 | Retrasado | Responde pero lento |
| 🔴 | Sin conexion | El proxy no responde |
| ⚫ | No conectado / Desactivado | El proxy nunca se ha conectado al almacen, o esta desactivado |

**Botones en la parte inferior de las tarjetas de agentes:**

Cuando registras un agente con un **tipo de agente** especifico, aparecen automaticamente botones de conveniencia adaptados a ese tipo.

---

#### 🔘 Boton Copiar configuracion — genera automaticamente los ajustes de conexion

Al hacer clic, se copia al portapapeles un fragmento de configuracion con el token del agente, la direccion del proxy y la informacion del modelo ya completados. Solo pega el contenido en la ubicacion indicada en la tabla para completar la configuracion.

| Boton | Tipo de agente | Donde pegar |
|-------|---------------|-------------|
| 🦞 Copiar config OpenClaw | `openclaw` | `~/.openclaw/openclaw.json` |
| 🦀 Copiar config NanoClaw | `nanoclaw` | `~/.openclaw/openclaw.json` |
| 🟠 Copiar config Claude Code | `claude-code` | `~/.claude/settings.json` |
| ⌨ Copiar config Cursor | `cursor` | Cursor → Settings → AI |
| 💻 Copiar config VSCode | `vscode` | `~/.continue/config.json` |

**Ejemplo — Para tipo Claude Code, se copia lo siguiente:**

```json
// ~/.claude/settings.json
{
  "apiProvider": "openai",
  "baseUrl": "http://192.168.1.20:56244/v1",
  "apiKey": "token-de-este-agente"
}
```

**Ejemplo — Para tipo VSCode (Continue):**

```yaml
# ~/.continue/config.yaml  ← pegar en config.yaml, NO en config.json
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

> ⚠️ **La ultima version de Continue usa `config.yaml`.** Si `config.yaml` existe, `config.json` se ignora completamente. Asegurate de pegar en `config.yaml`.

**Ejemplo — Para tipo Cursor:**

```
Base URL : http://192.168.1.20:56244/v1
API Key  : token-de-este-agente

// O como variables de entorno:
OPENAI_BASE_URL=http://192.168.1.20:56244/v1
OPENAI_API_KEY=token-de-este-agente
```

> ⚠️ **Si la copia al portapapeles no funciona**: las politicas de seguridad del navegador pueden bloquear la copia. Si aparece un popup con un cuadro de texto, selecciona todo con Ctrl+A y luego copia con Ctrl+C.

---

#### ⚡ Boton de aplicacion automatica — un clic y listo

Para agentes de tipo `cline`, `claude-code`, `openclaw` o `nanoclaw`, la tarjeta de agente muestra un boton **⚡ Aplicar config**. Al hacer clic, se actualiza automaticamente el archivo de configuracion local del agente.

| Boton | Tipo de agente | Archivo destino |
|-------|---------------|-----------------|
| ⚡ Aplicar config Cline | `cline` | `~/.cline/data/globalState.json` + `secrets.json` |
| ⚡ Aplicar config Claude Code | `claude-code` | `~/.claude/settings.json` |
| ⚡ Aplicar config OpenClaw | `openclaw` | `~/.openclaw/openclaw.json` |
| ⚡ Aplicar config NanoClaw | `nanoclaw` | `~/.openclaw/openclaw.json` |

> ⚠️ Este boton envia una solicitud a **localhost:56244** (proxy local). El proxy debe estar en ejecucion en esa maquina para que funcione.

---

#### 🔀 Ordenar tarjetas arrastrando y soltando (v0.1.17)

Puedes **arrastrar** las tarjetas de agentes en el panel de control para reorganizarlas en el orden que desees.

1. Agarra una tarjeta de agente con el raton y arrastrala
2. Sueltala sobre otra tarjeta para intercambiar posiciones
3. El nuevo orden se **guarda en el servidor inmediatamente** y persiste tras actualizar la pagina

> 💡 Los dispositivos tactiles (movil/tablet) aun no estan soportados. Usa un navegador de escritorio.

---

#### 🔄 Sincronizacion bidireccional de modelos (v0.1.16)

Cuando cambias el modelo de un agente en el panel de control del almacen, la configuracion local del agente se actualiza automaticamente.

**Para Cline:**
- Cambio de modelo en el almacen → evento SSE → el proxy actualiza el campo de modelo en `globalState.json`
- Campos actualizados: `actModeOpenAiModelId`, `planModeOpenAiModelId`, `openAiModelId`
- `openAiBaseUrl` y la clave API no se modifican
- **Se requiere recargar VS Code (`Ctrl+Alt+R` o `Ctrl+Shift+P` → `Developer: Reload Window`)**
  - Porque Cline no vuelve a leer el archivo de configuracion durante la ejecucion

**Para Claude Code:**
- Cambio de modelo en el almacen → evento SSE → el proxy actualiza el campo `model` en `settings.json`
- Busca automaticamente en rutas WSL y Windows (`~/.claude/`, `/mnt/c/Users/*/.claude/`)

**Direccion inversa (agente → almacen):**
- Cuando un agente (Cline, Claude Code, etc.) envia una solicitud al proxy, este incluye la informacion de servicio/modelo del cliente en el heartbeat
- La tarjeta de agente en el panel de control muestra el servicio/modelo actualmente en uso en tiempo real

> 💡 **Punto clave**: el proxy identifica a los agentes por su token de Authorization en las solicitudes, y los redirige automaticamente al servicio/modelo configurado en el almacen. Incluso si Cline o Claude Code envia un nombre de modelo diferente, el proxy lo sobreescribe con la configuracion del almacen.

---

### Usar Cline con VS Code — Guia detallada

#### Paso 1: Instalar Cline

Instala **Cline** (ID: `saoudrizwan.claude-dev`) desde el marketplace de extensiones de VS Code.

#### Paso 2: Registrar el agente en el almacen

1. Abre el panel de control del almacen (`http://IP-almacen:56243`)
2. Haz clic en **+ Agregar** en la seccion **Agentes**
3. Completa lo siguiente:

| Campo | Valor | Descripcion |
|-------|-------|-------------|
| ID | `mi_cline` | Identificador unico (alfanumerico, sin espacios) |
| Nombre | `Mi Cline` | Nombre que se muestra en el panel de control |
| Tipo de agente | `cline` | ← Debes seleccionar `cline` |
| Servicio | Selecciona el servicio a usar (ej: `google`) | |
| Modelo | Introduce el modelo a usar (ej: `gemini-2.5-flash`) | |

4. Haz clic en **Guardar** — se genera un token automaticamente

#### Paso 3: Conectar Cline

**Metodo A — Aplicacion automatica (recomendado)**

1. Asegurate de que el **proxy** wall-vault esta en ejecucion en esa maquina (`localhost:56244`)
2. Haz clic en el boton **⚡ Aplicar config Cline** en la tarjeta de agente del panel de control
3. Si aparece la notificacion "Configuracion aplicada con exito!", funciono
4. Recarga VS Code (`Ctrl+Alt+R`)

**Metodo B — Configuracion manual**

Abre Configuracion (⚙️) en la barra lateral de Cline:
- **API Provider**: `OpenAI Compatible`
- **Base URL**: `http://direccion-proxy:56244/v1`
  - Misma maquina: `http://localhost:56244/v1`
  - Otra maquina (ej: Mac Mini): `http://192.168.1.20:56244/v1`
- **API Key**: el token emitido por el almacen (copia desde la tarjeta de agente)
- **Model ID**: el modelo configurado en el almacen (ej: `gemini-2.5-flash`)

#### Paso 4: Verificar

Envia cualquier mensaje en la ventana de chat de Cline. Si funciona:
- La tarjeta de agente en el panel de control muestra un **punto verde (● En ejecucion)**
- La tarjeta muestra el servicio/modelo actual (ej: `google / gemini-2.5-flash`)

#### Cambiar de modelo

Cuando quieras cambiar el modelo de Cline, hazlo desde el **panel de control del almacen**:

1. Cambia el desplegable de servicio/modelo en la tarjeta de agente
2. Haz clic en **Aplicar**
3. Recarga VS Code (`Ctrl+Alt+R`) — el nombre del modelo en el pie de pagina de Cline se actualizara
4. El nuevo modelo se usara a partir de la siguiente solicitud

> 💡 En la practica, el proxy identifica las solicitudes de Cline por el token y las dirige al modelo configurado en el almacen. Incluso sin recargar VS Code, **el modelo real utilizado cambia inmediatamente** — la recarga solo sirve para actualizar la visualizacion del modelo en la interfaz de Cline.

#### Deteccion de desconexion

Cuando VS Code se cierra, la tarjeta de agente en el panel de control cambia a amarillo (retrasado) despues de unos **90 segundos**, y a rojo (sin conexion) despues de **3 minutos**. (Desde v0.1.18, la deteccion sin conexion es mas rapida gracias a comprobaciones de estado cada 15 segundos.)

#### Solucion de problemas

| Sintoma | Causa | Solucion |
|---------|-------|----------|
| Error "conexion fallida" en Cline | Proxy no iniciado o direccion incorrecta | Verifica el proxy con `curl http://localhost:56244/health` |
| El punto verde no aparece en el almacen | Clave API (token) no configurada | Haz clic de nuevo en el boton **⚡ Aplicar config Cline** |
| El modelo en el pie de Cline no cambia | Cline tiene la configuracion en cache | Recarga VS Code (`Ctrl+Alt+R`) |
| Se muestra un nombre de modelo incorrecto | Bug antiguo (corregido en v0.1.16) | Actualiza el proxy a v0.1.16 o posterior |

---

#### 🟣 Boton Copiar comando de despliegue — para instalar en una nueva maquina

Usa este boton cuando instales por primera vez el proxy wall-vault en un nuevo ordenador y lo conectes al almacen. Al hacer clic se copia el script de instalacion completo. Pegalo en el terminal del nuevo ordenador y ejecutalo para realizar todo de una vez:

1. Instalar el binario wall-vault (se omite si ya esta instalado)
2. Registrar automaticamente un servicio de usuario systemd
3. Iniciar el servicio y conectarse automaticamente al almacen

> 💡 El script ya contiene el token de este agente y la direccion del servidor del almacen, asi que puedes ejecutarlo directamente tras pegarlo sin ninguna modificacion.

---

### Tarjetas de Servicios

Tarjetas para activar/desactivar y configurar los servicios de IA.

- Interruptor de activacion/desactivacion para cada servicio
- Introduce la direccion de un servidor de IA local (Ollama, LM Studio, vLLM, etc. en tu ordenador) para descubrir automaticamente los modelos disponibles
- **Indicador de estado de conexion del servicio local**: un punto ● junto al nombre del servicio es **verde** si esta conectado, **gris** si no
- **Senalizacion automatica de servicios locales** (v0.1.23+): los servicios locales (Ollama, LM Studio, vLLM) se activan/desactivan automaticamente segun la disponibilidad de conexion. Cuando un servicio se vuelve accesible, el punto ● cambia a verde y la casilla se activa en menos de 15 segundos; cuando el servicio se cae, se desactiva automaticamente. Funciona de la misma manera que los servicios en la nube (Google, OpenRouter, etc.) se alternan automaticamente segun la disponibilidad de claves API.

> 💡 **Si el servicio local se ejecuta en otro ordenador**: introduce la IP de ese ordenador en el campo URL del servicio. Ejemplo: `http://192.168.1.20:11434` (Ollama), `http://192.168.1.20:1234` (LM Studio). Si el servicio solo esta enlazado a `127.0.0.1` en lugar de `0.0.0.0`, el acceso desde una IP externa no funcionara — verifica la direccion de enlace en la configuracion del servicio.

### Introduccion del token de administrador

Cuando intentas usar funciones importantes como agregar o eliminar claves en el panel de control, aparece un popup de introduccion del token de administrador. Introduce el token que configuraste en el asistente de configuracion. Una vez introducido, permanece valido hasta que cierres el navegador.

> ⚠️ **Si la autenticacion falla mas de 10 veces en 15 minutos, esa IP sera bloqueada temporalmente.** Si has olvidado tu token, consulta el campo `admin_token` en el archivo `wall-vault.yaml`.

---

## Modo distribuido (multi-bot)

Cuando ejecutas OpenClaw en multiples ordenadores simultaneamente, puedes **compartir un unico almacen de claves**. Es practico porque solo necesitas gestionar las claves en un lugar.

### Ejemplo de configuracion

```
[Servidor del Almacen de Claves]
  wall-vault vault    (almacen de claves :56243, panel de control)

[WSL Alpha]           [Raspberry Pi Gamma]    [Mac Mini Local]
  wall-vault proxy      wall-vault proxy        wall-vault proxy
  openclaw TUI          openclaw TUI            openclaw TUI
  ↕ sync SSE            ↕ sync SSE              ↕ sync SSE
```

Todos los bots apuntan al servidor central del almacen, asi que cuando cambias un modelo o agregas una clave en el almacen, se refleja inmediatamente en todos los bots.

### Paso 1: Iniciar el servidor del almacen de claves

Ejecuta esto en el ordenador que servira como servidor del almacen:

```bash
wall-vault vault
```

### Paso 2: Registrar cada bot (cliente)

Registra la informacion de cada bot que se conectara al servidor del almacen:

```bash
curl -X POST http://localhost:56243/admin/clients \
  -H "Authorization: Bearer tu-token-admin" \
  -H "Content-Type: application/json" \
  -d '{
    "id": "botA",
    "name": "Bot A",
    "token": "bota-secret",
    "default_service": "google",
    "default_model": "gemini-2.5-flash"
  }'
```

### Paso 3: Iniciar el proxy en cada ordenador del bot

En cada ordenador donde hay un bot instalado, ejecuta el proxy con la direccion del servidor del almacen y el token:

```bash
WV_VAULT_URL=http://192.168.x.x:56243 \
WV_VAULT_TOKEN=bota-secret \
WV_VAULT_CLIENT_ID=botA \
wall-vault proxy
```

> 💡 Reemplaza **`192.168.x.x`** con la direccion IP interna real del ordenador servidor del almacen. Puedes encontrarla en la configuracion de tu router o con el comando `ip addr`.

---

## Inicio automatico

Si es tedioso iniciar manualmente wall-vault cada vez que reinicias tu ordenador, registralo como servicio del sistema. Una vez registrado, se inicia automaticamente al arrancar.

### Linux — systemd (la mayoria de distribuciones Linux)

systemd es el sistema que inicia y gestiona automaticamente los programas en Linux:

```bash
wall-vault doctor deploy
systemctl --user daemon-reload
systemctl --user enable --now wall-vault
```

Consultar logs:

```bash
journalctl --user -u wall-vault -f
```

### macOS — launchd

El sistema responsable del inicio automatico de programas en macOS:

```bash
wall-vault doctor deploy launchd
launchctl load ~/Library/LaunchAgents/com.wall-vault.plist
```

### Windows — NSSM

1. Descarga NSSM desde [nssm.cc](https://nssm.cc/download) y agregalo al PATH.
2. En PowerShell como administrador:

```powershell
wall-vault doctor deploy windows
```

---

## Doctor: diagnostico automatico

El comando `doctor` es una herramienta que **diagnostica y repara** automaticamente la configuracion de wall-vault.

```bash
wall-vault doctor check   # Diagnosticar estado actual (solo lectura, no cambia nada)
wall-vault doctor fix     # Reparar problemas automaticamente
wall-vault doctor all     # Diagnostico + reparacion automatica en un solo paso
```

> 💡 Si algo parece ir mal, prueba primero `wall-vault doctor all`. Detecta y corrige muchos problemas automaticamente.

---

## Variables de entorno

Las variables de entorno son una forma de pasar valores de configuracion a un programa. Introduzlas en el terminal con `export VARIABLE=valor`, o agregalas en tu archivo de servicio de inicio automatico para aplicacion permanente.

| Variable | Descripcion | Valor de ejemplo |
|----------|-------------|-----------------|
| `WV_LANG` | Idioma del panel de control | `ko`, `en`, `ja` |
| `WV_THEME` | Tema del panel de control | `light`, `dark`, `gold` |
| `WV_KEY_GOOGLE` | Clave API de Google (separadas por comas para multiples) | `AIza...,AIza...` |
| `WV_KEY_OPENROUTER` | Clave API de OpenRouter | `sk-or-v1-...` |
| `WV_VAULT_URL` | Direccion del servidor del almacen en modo distribuido | `http://192.168.x.x:56243` |
| `WV_VAULT_TOKEN` | Token de autenticacion del cliente (bot) | `my-secret-token` |
| `WV_ADMIN_TOKEN` | Token de administrador | `admin-token-here` |
| `WV_MASTER_PASS` | Contrasena de cifrado de claves API | `my-password` |
| `WV_AVATAR` | Ruta del archivo de imagen avatar (relativa a `~/.openclaw/`) | `workspace/avatars/avatar.png` |
| `OLLAMA_URL` | Direccion del servidor local Ollama | `http://192.168.x.x:11434` |

---

## Solucion de problemas

### El proxy no se inicia

El puerto suele estar en uso por otro programa.

```bash
ss -tlnp | grep 56244   # Comprobar quien usa el puerto 56244
wall-vault proxy --port 8080   # Iniciar en otro puerto
```

### Errores de clave API (429, 402, 401, 403, 582)

| Codigo de error | Significado | Que hacer |
|----------------|-------------|-----------|
| **429** | Demasiadas solicitudes (cuota excedida) | Esperar un rato o agregar mas claves |
| **402** | Pago requerido o creditos agotados | Recargar creditos en ese servicio |
| **401 / 403** | Clave invalida o sin permiso | Verificar el valor de la clave y volver a registrar |
| **582** | Sobrecarga del gateway (enfriamiento de 5 minutos) | Se resuelve automaticamente en 5 minutos |

```bash
# Consultar la lista de claves registradas y su estado
curl -H "Authorization: Bearer tu-token-admin" http://localhost:56243/admin/keys

# Reiniciar contadores de uso de claves
curl -X POST -H "Authorization: Bearer tu-token-admin" http://localhost:56243/admin/keys/reset
```

### El agente aparece como "No conectado"

"No conectado" significa que el proceso proxy no esta enviando senales (heartbeat) al almacen. **No significa que la configuracion no se haya guardado.** El proxy necesita estar en ejecucion con la direccion del servidor del almacen y el token para establecer una conexion.

```bash
# Iniciar proxy con direccion del servidor del almacen, token e ID de cliente
WV_VAULT_URL=http://servidor-almacen:56243 \
WV_VAULT_TOKEN=token-cliente \
WV_VAULT_CLIENT_ID=id-cliente \
wall-vault proxy
```

Una vez conectado, el panel de control mostrara 🟢 En ejecucion en unos 20 segundos.

### Problemas de conexion con Ollama

Ollama es un programa que ejecuta IA directamente en tu ordenador. Primero, asegurate de que Ollama esta en ejecucion.

```bash
curl http://localhost:11434/api/tags   # Si aparece una lista de modelos, funciona
export OLLAMA_URL=http://192.168.x.x:11434   # Si se ejecuta en otro ordenador
```

> ⚠️ Si Ollama no responde, inicialo primero con `ollama serve`.

> ⚠️ **Los modelos grandes son lentos**: modelos grandes como `qwen3.5:35b` o `deepseek-r1` pueden tardar varios minutos en generar una respuesta. Aunque parezca que no pasa nada, puede estar procesando normalmente — ten paciencia.

---

## Cambios recientes (v0.1.16 ~ v0.1.23)

### v0.1.23 (2026-04-06)
- **Correccion del cambio de modelo Ollama**: se corrigio un problema donde cambiar el modelo Ollama en el panel de control del almacen no se reflejaba en el proxy real. Anteriormente solo se usaba la variable de entorno (`OLLAMA_MODEL`), pero ahora la configuracion del almacen tiene prioridad.
- **Senalizacion automatica de servicios locales**: Ollama, LM Studio y vLLM se activan automaticamente cuando estan accesibles y se desactivan cuando no lo estan. Funciona igual que la alternancia automatica basada en claves para servicios en la nube.

### v0.1.22 (2026-04-05)
- **Correccion del campo content vacio**: cuando los modelos de razonamiento (gemini-3.1-pro, o1, claude thinking, etc.) usan todos los max_tokens en el razonamiento y no pueden producir una respuesta real, el proxy omitia los campos `content`/`text` del JSON de respuesta via `omitempty`, causando que los clientes SDK de OpenAI/Anthropic se bloquearan con `Cannot read properties of undefined (reading 'trim')`. Corregido para siempre incluir los campos segun la especificacion oficial de la API.

### v0.1.21 (2026-04-05)
- **Soporte de modelos Gemma 4**: los modelos de la familia Gemma como `gemma-4-31b-it` y `gemma-4-26b-a4b-it` ahora pueden usarse a traves de la API de Google Gemini.
- **Soporte de servicios LM Studio / vLLM**: anteriormente estos servicios no estaban en el enrutamiento del proxy y siempre volvian a Ollama. Ahora se enrutan correctamente via la API compatible con OpenAI.
- **Correccion de visualizacion de servicios en el panel**: incluso cuando ocurre un fallback, el panel siempre muestra el servicio configurado por el usuario.
- **Visualizacion del estado de servicios locales**: al cargar el panel, el estado de conexion de los servicios locales (Ollama, LM Studio, vLLM, etc.) se muestra con el color del punto ●.
- **Variable de entorno para filtro de herramientas**: usa `WV_TOOL_FILTER=passthrough` para configurar el modo de paso de herramientas.

### v0.1.20 (2026-03-28)
- **Fortalecimiento integral de seguridad**: prevencion XSS (41 ubicaciones), comparacion de tokens en tiempo constante, restricciones CORS, limites de tamano de solicitud, prevencion de travesia de rutas, autenticacion SSE, endurecimiento del limitador de velocidad y 12 mejoras de seguridad adicionales.

### v0.1.19 (2026-03-27)
- **Deteccion en linea de Claude Code**: las instancias de Claude Code que no pasan por el proxy ahora se muestran como en linea en el panel de control.

### v0.1.18 (2026-03-26)
- **Correccion del servicio de fallback pegajoso**: despues de un error temporal que causa un fallback a Ollama, se vuelve automaticamente al servicio original cuando se recupera.
- **Mejora en la deteccion sin conexion**: las comprobaciones de estado cada 15 segundos hacen que la deteccion de caidas del proxy sea mas rapida.

### v0.1.17 (2026-03-25)
- **Ordenar tarjetas arrastrando y soltando**: las tarjetas de agentes se pueden arrastrar y soltar para cambiar su orden.
- **Boton de aplicacion de configuracion en linea**: el boton [⚡ Aplicar config] se muestra en las tarjetas de agentes sin conexion.
- **Tipo de agente cokacdir agregado**.

### v0.1.16 (2026-03-25)
- **Sincronizacion bidireccional de modelos**: al cambiar un modelo de Cline o Claude Code en el panel de control del almacen, se refleja automaticamente.

---

*Para informacion mas detallada sobre la API, consulta [API.md](API.md).*
