# Manual de Usuario de wall-vault
*(Última actualización: 2026-04-08 — v0.1.25)*

---

## Tabla de contenidos

1. [¿Qué es wall-vault?](#qué-es-wall-vault)
2. [Instalación](#instalación)
3. [Primeros pasos (asistente de configuración)](#primeros-pasos)
4. [Registro de claves API](#registro-de-claves-api)
5. [Uso del proxy](#uso-del-proxy)
6. [Panel de control de la bóveda](#panel-de-control-de-la-bóveda)
7. [Modo distribuido (multi-bot)](#modo-distribuido-multi-bot)
8. [Configuración de inicio automático](#configuración-de-inicio-automático)
9. [Doctor (herramienta de diagnóstico)](#doctor-herramienta-de-diagnóstico)
10. [RTK Ahorro de tokens](#rtk-ahorro-de-tokens)
11. [Referencia de variables de entorno](#referencia-de-variables-de-entorno)
12. [Solución de problemas](#solución-de-problemas)

---

## ¿Qué es wall-vault?

**wall-vault = Proxy de IA + bóveda de claves API para OpenClaw**

Para usar servicios de IA, necesitas **claves API**. Una clave API es como un **pase digital** que demuestra que "esta persona está autorizada para usar este servicio". Sin embargo, estos pases tienen límites de uso diario, y si se gestionan mal, corren el riesgo de ser expuestos.

wall-vault almacena estos pases en una bóveda segura y actúa como **proxy (intermediario)** entre OpenClaw y los servicios de IA. En términos sencillos, OpenClaw solo necesita conectarse a wall-vault, y wall-vault se encarga automáticamente de todo lo demás.

Problemas que wall-vault resuelve:

- **Rotación automática de claves API**: Cuando una clave alcanza su límite de uso o se bloquea temporalmente (enfriamiento), cambia silenciosamente a la siguiente clave. OpenClaw continúa funcionando sin interrupciones.
- **Conmutación automática de servicios (fallback)**: Si Google no responde, cambia a OpenRouter; si eso también falla, cambia a Ollama, LM Studio o vLLM (IA local) instalados en tu computadora. Las sesiones nunca se interrumpen. Cuando el servicio original se recupera, vuelve automáticamente en la siguiente solicitud (v0.1.18+, LM Studio/vLLM: v0.1.21+).
- **Sincronización en tiempo real (SSE)**: Cuando cambias un modelo en el panel de la bóveda, se refleja en la pantalla de OpenClaw en 1-3 segundos. SSE (Server-Sent Events) es una tecnología donde el servidor envía cambios a los clientes en tiempo real.
- **Notificaciones en tiempo real**: Eventos como el agotamiento de claves o caídas de servicio se muestran inmediatamente en la parte inferior de la interfaz TUI de OpenClaw (pantalla de terminal).

> 💡 **Claude Code, Cursor y VS Code** también pueden conectarse, pero el propósito principal de wall-vault es usarse con OpenClaw.

```
OpenClaw (interfaz TUI de terminal)
        │
        ▼
  wall-vault proxy (:56244)   ← gestión de claves, enrutamiento, fallback, eventos
        │
        ├─ Google Gemini API
        ├─ OpenRouter API (340+ modelos)
        ├─ Ollama / LM Studio / vLLM (tu computadora, último recurso)
        └─ OpenAI / Anthropic API
```

---

## Instalación

### Linux / macOS

Abre una terminal y pega los siguientes comandos.

```bash
# Linux (PC estándar, servidor — amd64)
curl -L https://github.com/sookmook/wall-vault/releases/latest/download/wall-vault-linux-amd64 \
  -o ~/.local/bin/wall-vault && chmod +x ~/.local/bin/wall-vault

# macOS Apple Silicon (M1/M2/M3 Mac)
curl -L https://github.com/sookmook/wall-vault/releases/latest/download/wall-vault-darwin-arm64 \
  -o /usr/local/bin/wall-vault && chmod +x /usr/local/bin/wall-vault
```

- `curl -L ...` — Descarga un archivo de Internet.
- `chmod +x` — Hace que el archivo descargado sea "ejecutable". Si omites este paso, obtendrás un error de "permiso denegado".

### Windows

Abre PowerShell (como administrador) y ejecuta los siguientes comandos.

```powershell
# Descarga
Invoke-WebRequest -Uri `
  "https://github.com/sookmook/wall-vault/releases/latest/download/wall-vault-windows-amd64.exe" `
  -OutFile "$env:LOCALAPPDATA\Programs\wall-vault\wall-vault.exe"

# Agregar al PATH (se aplica después de reiniciar PowerShell)
$env:PATH += ";$env:LOCALAPPDATA\Programs\wall-vault"
```

> 💡 **¿Qué es el PATH?** Es la lista de carpetas donde tu computadora busca comandos. Necesitas agregar wall-vault al PATH para poder ejecutar `wall-vault` desde cualquier carpeta.

### Compilar desde el código fuente (para desarrolladores)

Solo aplicable si tienes un entorno de desarrollo Go instalado.

```bash
git clone https://github.com/sookmook/wall-vault
cd wall-vault
make build       # bin/wall-vault (versión: v0.1.25.YYYYMMDD.HHmmss)
make install     # ~/.local/bin/wall-vault
```

> 💡 **Versión con marca de tiempo de compilación**: Al compilar con `make build`, la versión se genera automáticamente en un formato como `v0.1.25.20260408.022325` que incluye fecha y hora. Si compilas directamente con `go build ./...`, la versión solo mostrará `"dev"`.

---

## Primeros pasos

### Ejecutar el asistente de configuración

Después de la instalación, debes ejecutar primero el **asistente de configuración** con el siguiente comando. El asistente te guía paso a paso por los ajustes necesarios.

```bash
wall-vault setup
```

El asistente recorre los siguientes pasos:

```
1. Selección de idioma (10 idiomas incluyendo coreano)
2. Selección de tema (light / dark / gold / cherry / ocean)
3. Modo de operación — uso individual (standalone) o compartido (distributed)
4. Nombre del bot — el nombre mostrado en el panel de control
5. Configuración de puertos — por defecto: proxy 56244, bóveda 56243 (Enter para mantener valores predeterminados)
6. Selección de servicios de IA — Google / OpenRouter / Ollama / LM Studio / vLLM
7. Configuración del filtro de seguridad de herramientas
8. Token de administrador — una contraseña para bloquear las funciones de administración del panel. Se puede generar automáticamente
9. Contraseña de cifrado de claves API — para almacenamiento extra seguro de claves (opcional)
10. Ubicación de guardado del archivo de configuración
```

> ⚠️ **Asegúrate de recordar tu token de administrador.** Lo necesitarás más tarde para agregar claves o cambiar configuraciones en el panel de control. Si lo pierdes, tendrás que editar directamente el archivo de configuración.

Una vez completado el asistente, se genera automáticamente un archivo de configuración `wall-vault.yaml`.

### Ejecución

```bash
wall-vault start
```

Dos servidores inician simultáneamente:

- **Proxy** (`http://localhost:56244`) — el intermediario que conecta OpenClaw con los servicios de IA
- **Bóveda de claves** (`http://localhost:56243`) — gestión de claves API y panel de control web

Abre `http://localhost:56243` en tu navegador para acceder al panel de control.

---

## Registro de claves API

Hay cuatro formas de registrar claves API. **Se recomienda el método 1 (variables de entorno) para principiantes.**

### Método 1: Variables de entorno (recomendado — el más simple)

Las variables de entorno son **valores preconfigurados** que los programas leen al iniciar. Introdúcelas en tu terminal así:

```bash
# Registrar una clave de Google Gemini
export WV_KEY_GOOGLE=AIzaSy...

# Registrar una clave de OpenRouter
export WV_KEY_OPENROUTER=sk-or-v1-...

# Iniciar después del registro
wall-vault start
```

Si tienes varias claves, conéctalas con comas. wall-vault las rotará automáticamente (round-robin):

```bash
export WV_KEY_GOOGLE=AIzaSy...,AIzaSy...,AIzaSy...
```

> 💡 **Consejo**: El comando `export` solo se aplica a la sesión de terminal actual. Para que persista después de reinicios, agrega las líneas a tu archivo `~/.bashrc` o `~/.zshrc`.

### Método 2: Interfaz del panel de control (apuntar y hacer clic)

1. Abre `http://localhost:56243` en tu navegador
2. Haz clic en `[+ Agregar]` en la tarjeta **🔑 Claves API** en la parte superior
3. Introduce el tipo de servicio, valor de la clave, etiqueta (nombre descriptivo) y límite diario, luego guarda

### Método 3: API REST (para automatización/scripts)

La API REST es un método para que los programas intercambien datos por HTTP. Útil para el registro automatizado mediante scripts.

```bash
curl -X POST http://localhost:56243/admin/keys \
  -H "Authorization: Bearer TU_TOKEN_ADMIN" \
  -H "Content-Type: application/json" \
  -d '{
    "service": "google",
    "key": "AIzaSy...",
    "label": "Clave principal",
    "daily_limit": 1000
  }'
```

### Método 4: Flags del proxy (para pruebas rápidas)

Para pruebas temporales sin registro formal. Las claves se pierden cuando el programa se cierra.

```bash
wall-vault proxy --key-google=AIzaSy... --key-openrouter=sk-or-...
```

---

## Uso del proxy

### Uso con OpenClaw (propósito principal)

Así se configura OpenClaw para conectarse a servicios de IA a través de wall-vault.

Abre `~/.openclaw/openclaw.json` y agrega lo siguiente:

```json5
// ~/.openclaw/openclaw.json
{
  models: {
    providers: {
      "wall-vault": {
        baseUrl: "http://localhost:56244/v1",
        apiKey: "your-agent-token",   // token de agente de la bóveda
        api: "openai-completions",
        models: [
          { id: "wall-vault/gemini-2.5-flash" },
          { id: "wall-vault/gemini-2.5-pro" },
          { id: "wall-vault/hunter-alpha" },    // 1M contexto gratis
          { id: "wall-vault/claude-opus-4-6" }
        ]
      }
    }
  }
}
```

> 💡 **Método más fácil**: Haz clic en el botón **🦞 Copiar configuración de OpenClaw** en la tarjeta de agente del panel de control. Se copiará al portapapeles un fragmento con el token y la dirección ya completados. Solo pégalo.

**¿A dónde redirige el prefijo `wall-vault/` en los nombres de modelo?**

wall-vault determina automáticamente a qué servicio de IA enviar las solicitudes basándose en el nombre del modelo:

| Formato del modelo | Servicio destino |
|-------------------|-----------------|
| `wall-vault/gemini-*` | Directo a Google Gemini |
| `wall-vault/gpt-*`, `wall-vault/o3`, `wall-vault/o4*` | Directo a OpenAI |
| `wall-vault/claude-*` | Anthropic vía OpenRouter |
| `wall-vault/hunter-alpha`, `wall-vault/healer-alpha` | OpenRouter (1M tokens contexto gratis) |
| `wall-vault/kimi-*`, `wall-vault/glm-*`, `wall-vault/deepseek-*` | OpenRouter |
| `google/modelo`, `openai/modelo`, `anthropic/modelo`, etc. | Directo al servicio correspondiente |
| `custom/google/modelo`, `custom/openai/modelo`, etc. | Elimina el prefijo `custom/` y redirige |
| `modelo:cloud` | Elimina el sufijo `:cloud` y redirige a OpenRouter |

> 💡 **¿Qué es el contexto?** Es la cantidad de conversación que una IA puede recordar de una vez. 1M (un millón de tokens) significa que puede procesar conversaciones o documentos muy largos en una sola pasada.

### Conexión directa en formato Gemini API (compatibilidad con herramientas existentes)

Si tienes herramientas que usaban directamente la API de Google Gemini, simplemente cambia la dirección a wall-vault:

```bash
export ANTHROPIC_BASE_URL=http://localhost:56244/google
```

O si la herramienta especifica URLs directamente:

```
http://localhost:56244/google/v1beta/models/gemini-2.5-flash:generateContent
```

### Uso con el SDK de OpenAI (Python)

También puedes conectar wall-vault desde código Python que use IA. Solo cambia la `base_url`:

```python
from openai import OpenAI

client = OpenAI(
    base_url="http://localhost:56244/v1",
    api_key="not-needed"  # wall-vault gestiona las claves API por ti
)

response = client.chat.completions.create(
    model="google/gemini-2.5-flash",   # usar formato provider/model
    messages=[{"role": "user", "content": "Hola"}]
)
```

### Cambiar modelos durante la ejecución

Para cambiar el modelo de IA mientras wall-vault está en ejecución:

```bash
# Cambiar modelo solicitando directamente al proxy
curl -X PUT http://localhost:56244/api/config/model \
  -H "Content-Type: application/json" \
  -d '{"service": "openrouter", "model": "anthropic/claude-3.5-sonnet"}'

# En modo distribuido (multi-bot), cambiar desde el servidor bóveda → sincronizado al instante vía SSE
curl -X PUT http://localhost:56243/admin/clients/mi-bot-id \
  -H "Authorization: Bearer TU_TOKEN_ADMIN" \
  -H "Content-Type: application/json" \
  -d '{"default_service": "google", "default_model": "gemini-2.5-pro"}'
```

### Consultar modelos disponibles

```bash
# Ver lista completa
curl http://localhost:56244/api/models | python3 -m json.tool

# Ver solo modelos de Google
curl "http://localhost:56244/api/models?service=google"

# Buscar por nombre (ej.: modelos que contienen "claude")
curl "http://localhost:56244/api/models?q=claude"
```

**Principales modelos por servicio:**

| Servicio | Principales modelos |
|----------|-------------------|
| Google | gemini-2.5-pro, gemini-2.5-flash, gemini-2.5-flash-8b, gemini-2.0-flash |
| OpenAI | gpt-4o, gpt-4o-mini, o3, o1, o1-mini |
| OpenRouter | 346+ (Hunter Alpha 1M contexto gratis, DeepSeek R1/V3, Qwen 2.5, etc.) |
| Ollama | Detecta automáticamente los modelos de tu servidor local |
| LM Studio | Servidor local (puerto 1234) |
| vLLM | Servidor local (puerto 8000) |

---

## Panel de control de la bóveda

Abre `http://localhost:56243` en tu navegador para acceder al panel de control.

**Diseño:**
- **Barra superior (fija)**: Logo, selector de idioma/tema, estado de conexión SSE
- **Cuadrícula de tarjetas**: Tarjetas de agentes, servicios y claves API dispuestas en mosaico

### Tarjeta de claves API

Una tarjeta para gestionar de un vistazo todas las claves API registradas.

- Muestra listas de claves organizadas por servicio.
- `today_usage`: Número de tokens (caracteres leídos/escritos por la IA) procesados exitosamente hoy
- `today_attempts`: Número total de llamadas hoy (incluyendo éxitos y fallos)
- Botón `[+ Agregar]` para registrar nuevas claves, `✕` para eliminarlas.

> 💡 **¿Qué es un token?** Es una unidad que la IA usa para procesar texto. Equivale aproximadamente a una palabra en inglés, o 1-2 caracteres en español. Los precios de las API se calculan normalmente según el número de tokens.

### Tarjeta de agente

Una tarjeta que muestra el estado de los bots (agentes) conectados al proxy wall-vault.

**El estado de conexión se muestra en 4 niveles:**

| Indicador | Estado | Significado |
|-----------|--------|------------|
| 🟢 | En ejecución | El proxy funciona normalmente |
| 🟡 | Retrasado | Responde pero lento |
| 🔴 | Sin conexión | El proxy no responde |
| ⚫ | No conectado / Deshabilitado | El proxy nunca se conectó a la bóveda o está deshabilitado |

**Guía de botones en la parte inferior de la tarjeta de agente:**

Cuando registras un agente y especificas el **tipo de agente**, aparecen automáticamente botones de conveniencia que coinciden con ese tipo.

---

#### 🔘 Botón Copiar configuración — genera automáticamente los ajustes de conexión

Al hacer clic en el botón, se copia al portapapeles un fragmento de configuración con el token, la dirección del proxy y la información del modelo del agente ya completados. Simplemente pega el contenido copiado en la ubicación indicada en la tabla para completar la configuración de conexión.

| Botón | Tipo de agente | Ubicación de pegado |
|-------|---------------|-------------------|
| 🦞 Copiar config de OpenClaw | `openclaw` | `~/.openclaw/openclaw.json` |
| 🦀 Copiar config de NanoClaw | `nanoclaw` | `~/.openclaw/openclaw.json` |
| 🟠 Copiar config de Claude Code | `claude-code` | `~/.claude/settings.json` |
| ⌨ Copiar config de Cursor | `cursor` | Cursor → Settings → AI |
| 💻 Copiar config de VSCode | `vscode` | `~/.continue/config.json` |

**Ejemplo — para el tipo Claude Code, se copia este contenido:**

```json
// ~/.claude/settings.json
{
  "apiProvider": "openai",
  "baseUrl": "http://192.168.0.6:56244/v1",
  "apiKey": "token-de-este-agente"
}
```

**Ejemplo — para el tipo VSCode (Continue):**

```yaml
# ~/.continue/config.yaml  ← pegar en config.yaml, no en config.json
name: My Config
version: 0.0.1
schema: v1

models:
  - name: wall-vault proxy
    provider: openai
    model: gemini-2.5-flash
    apiBase: http://192.168.0.6:56244/v1
    apiKey: token-de-este-agente
    roles:
      - chat
      - edit
      - apply
```

> ⚠️ **La última versión de Continue usa `config.yaml`.** Si `config.yaml` existe, `config.json` se ignora completamente. Asegúrate de pegar en `config.yaml`.

**Ejemplo — para el tipo Cursor:**

```
Base URL : http://192.168.0.6:56244/v1
API Key  : token-de-este-agente

// O variables de entorno:
OPENAI_BASE_URL=http://192.168.0.6:56244/v1
OPENAI_API_KEY=token-de-este-agente
```

> ⚠️ **Si la copia al portapapeles no funciona**: Las políticas de seguridad del navegador pueden bloquear la copia. Si aparece un cuadro de texto emergente, usa Ctrl+A para seleccionar todo, luego Ctrl+C para copiar.

---

#### ⚡ Botón de aplicación automática — un clic para completar la configuración

Para los tipos de agente `cline`, `claude-code`, `openclaw` o `nanoclaw`, aparece un botón **⚡ Aplicar configuración** en la tarjeta de agente. Al hacer clic, se actualiza automáticamente el archivo de configuración local del agente.

| Botón | Tipo de agente | Archivo destino |
|-------|---------------|----------------|
| ⚡ Aplicar config de Cline | `cline` | `~/.cline/data/globalState.json` + `secrets.json` |
| ⚡ Aplicar config de Claude Code | `claude-code` | `~/.claude/settings.json` |
| ⚡ Aplicar config de OpenClaw | `openclaw` | `~/.openclaw/openclaw.json` |
| ⚡ Aplicar config de NanoClaw | `nanoclaw` | `~/.openclaw/openclaw.json` |

> ⚠️ Este botón envía solicitudes a **localhost:56244** (proxy local). El proxy debe estar en ejecución en esa máquina para que funcione.

---

#### 🔀 Ordenar tarjetas con arrastrar y soltar (v0.1.17, mejorado v0.1.25)

Puedes **arrastrar** las tarjetas de agente del panel de control para reorganizarlas en el orden que prefieras.

1. Agarra el área del **semáforo (●)** en la esquina superior izquierda de la tarjeta con el ratón y arrastra
2. Suéltala sobre otra tarjeta para intercambiar posiciones

> 💡 El cuerpo de la tarjeta (campos de entrada, botones, etc.) no es arrastrable. Solo puedes agarrar desde el área del semáforo.

#### 🟠 Detección de proceso de agente (v0.1.25)

Cuando el proxy funciona normalmente pero el proceso del agente local (NanoClaw, OpenClaw) ha muerto, el semáforo de la tarjeta cambia a **naranja (parpadeando)** y muestra un mensaje de "Proceso de agente detenido".

- 🟢 Verde: Proxy + agente ambos normales
- 🟠 Naranja (parpadeando): Proxy normal, agente muerto
- 🔴 Rojo: Proxy sin conexión
3. El orden modificado se **guarda inmediatamente en el servidor** y persiste después de actualizar

> 💡 Los dispositivos táctiles (móvil/tablet) aún no son compatibles. Por favor, usa un navegador de escritorio.

---

#### 🔄 Sincronización bidireccional de modelos (v0.1.16)

Cuando cambias el modelo de un agente en el panel de la bóveda, la configuración local del agente se actualiza automáticamente.

**Para Cline:**
- Cambiar modelo en la bóveda → evento SSE → el proxy actualiza los campos de modelo en `globalState.json`
- Campos actualizados: `actModeOpenAiModelId`, `planModeOpenAiModelId`, `openAiModelId`
- `openAiBaseUrl` y la clave API no se modifican
- **Se requiere recargar VS Code (`Ctrl+Alt+R` o `Ctrl+Shift+P` → `Developer: Reload Window`)**
  - Porque Cline no relee los archivos de configuración durante la ejecución

**Para Claude Code:**
- Cambiar modelo en la bóveda → evento SSE → el proxy actualiza el campo `model` en `settings.json`
- Busca automáticamente rutas tanto de WSL como de Windows (`~/.claude/`, `/mnt/c/Users/*/.claude/`)

**Dirección inversa (agente → bóveda):**
- Cuando un agente (Cline, Claude Code, etc.) envía una solicitud al proxy, el proxy incluye la información de servicio/modelo de ese cliente en el heartbeat
- La tarjeta de agente en el panel de la bóveda muestra el servicio/modelo actualmente en uso en tiempo real

> 💡 **Punto clave**: El proxy identifica a los agentes por el token de autorización en las solicitudes y enruta automáticamente al servicio/modelo configurado en la bóveda. Incluso si Cline o Claude Code envía un nombre de modelo diferente, el proxy lo anula con la configuración de la bóveda.

---

### Usar Cline en VS Code — Guía detallada

#### Paso 1: Instalar Cline

Instala **Cline** (ID: `saoudrizwan.claude-dev`) desde el marketplace de extensiones de VS Code.

#### Paso 2: Registrar el agente en la bóveda

1. Abre el panel de control de la bóveda (`http://IP_BOVEDA:56243`)
2. Haz clic en **+ Agregar** en la sección de **Agentes**
3. Introduce lo siguiente:

| Campo | Valor | Descripción |
|-------|-------|-------------|
| ID | `mi_cline` | Identificador único (alfanumérico, sin espacios) |
| Nombre | `Mi Cline` | Nombre mostrado en el panel de control |
| Tipo de agente | `cline` | ← debe seleccionar `cline` |
| Servicio | Seleccionar el servicio deseado (ej.: `google`) | |
| Modelo | Introducir el modelo deseado (ej.: `gemini-2.5-flash`) | |

4. Haz clic en **Guardar** — se genera un token automáticamente

#### Paso 3: Conectar a Cline

**Método A — Aplicación automática (recomendado)**

1. Verifica que el **proxy** wall-vault está en ejecución en esa máquina (`localhost:56244`)
2. Haz clic en el botón **⚡ Aplicar config de Cline** en la tarjeta de agente del panel de control
3. Éxito cuando aparezca la notificación "¡Configuración aplicada!"
4. Recarga VS Code (`Ctrl+Alt+R`)

**Método B — Configuración manual**

Abre los ajustes (⚙️) en la barra lateral de Cline:
- **API Provider**: `OpenAI Compatible`
- **Base URL**: `http://DIRECCION_PROXY:56244/v1`
  - Misma máquina: `http://localhost:56244/v1`
  - Otra máquina (ej.: Mac Mini): `http://192.168.0.6:56244/v1`
- **API Key**: Token emitido por la bóveda (copiar desde la tarjeta de agente)
- **Model ID**: Modelo configurado en la bóveda (ej.: `gemini-2.5-flash`)

#### Paso 4: Verificar

Envía cualquier mensaje en el chat de Cline. Si todo funciona:
- La tarjeta de agente en el panel de control muestra un **punto verde (● En ejecución)**
- La tarjeta muestra el servicio/modelo actual (ej.: `google / gemini-2.5-flash`)

#### Cambiar modelos

Cuando quieras cambiar el modelo de Cline, cámbialo desde el **panel de control de la bóveda**:

1. Cambia el menú desplegable de servicio/modelo en la tarjeta de agente
2. Haz clic en **Aplicar**
3. Recarga VS Code (`Ctrl+Alt+R`) — el nombre del modelo en el pie de página de Cline se actualiza
4. El nuevo modelo se usa a partir de la siguiente solicitud

> 💡 En la práctica, el proxy identifica las solicitudes de Cline por el token y las enruta al modelo configurado en la bóveda. Incluso sin recargar VS Code, **el modelo realmente usado cambia inmediatamente** — la recarga solo sirve para actualizar la visualización del modelo en la interfaz de Cline.

#### Detección de desconexión

Cuando cierras VS Code, la tarjeta de agente en el panel de control se vuelve amarilla (retrasado) después de unos **90 segundos**, luego roja (sin conexión) después de **3 minutos**. (Desde v0.1.18, las comprobaciones de estado cada 15 segundos hacen más rápida la detección de desconexión.)

#### Solución de problemas

| Síntoma | Causa | Solución |
|---------|-------|----------|
| Error "Conexión fallida" en Cline | Proxy no iniciado o dirección incorrecta | Verificar proxy con `curl http://localhost:56244/health` |
| El punto verde no aparece en la bóveda | Clave API (token) no configurada | Hacer clic de nuevo en **⚡ Aplicar config de Cline** |
| El modelo en el pie de Cline no cambia | Cline tiene la configuración en caché | Recargar VS Code (`Ctrl+Alt+R`) |
| Se muestra nombre de modelo incorrecto | Bug antiguo (corregido en v0.1.16) | Actualizar proxy a v0.1.16+ |

---

#### 🟣 Botón Copiar comando de despliegue — para instalar en máquinas nuevas

Se usa al instalar por primera vez el proxy wall-vault en una computadora nueva y conectarlo a la bóveda. Al hacer clic, se copia el script de instalación completo. Pégalo en la terminal de la nueva computadora y ejecútalo — lo siguiente se procesa de una sola vez:

1. Instalar el binario wall-vault (se omite si ya está instalado)
2. Registrar automáticamente el servicio de usuario systemd
3. Iniciar el servicio y conectar automáticamente a la bóveda

> 💡 El script ya contiene el token y la dirección del servidor bóveda de este agente, así que puedes ejecutarlo inmediatamente después de pegarlo sin ninguna modificación.

---

### Tarjeta de servicio

Una tarjeta para activar/desactivar y configurar servicios de IA.

- Interruptores para activar/desactivar cada servicio
- Introduce la dirección de un servidor de IA local (Ollama, LM Studio, vLLM, etc. en tu computadora) para descubrir automáticamente los modelos disponibles.
- **Visualización del estado de conexión de servicios locales**: El punto ● junto al nombre del servicio es **verde** cuando está conectado, **gris** cuando no lo está
- **Semáforo automático de servicios locales** (v0.1.23+): Los servicios locales (Ollama, LM Studio, vLLM) se activan/desactivan automáticamente según la disponibilidad de conexión. Cuando un servicio se conecta, el punto ● se vuelve verde y la casilla se marca en 15 segundos; cuando se desconecta, se desactiva automáticamente. Funciona de la misma manera que la alternancia automática de servicios en la nube (Google, OpenRouter, etc.) basada en la disponibilidad de claves API.

> 💡 **Si tu servicio local se ejecuta en otra computadora**: Introduce la IP de esa computadora en el campo de URL del servicio. Ejemplo: `http://192.168.0.6:11434` (Ollama), `http://192.168.0.6:1234` (LM Studio). Si el servicio solo está vinculado a `127.0.0.1` en lugar de `0.0.0.0`, el acceso por IP externa no funcionará — verifica la dirección de vinculación en la configuración del servicio.

### Ingreso del token de administrador

Cuando intentas usar funciones importantes como agregar o eliminar claves en el panel de control, aparece un popup de ingreso del token de administrador. Introduce el token que configuraste durante el asistente de configuración. Una vez introducido, persiste hasta que cierres el navegador.

> ⚠️ **Si los fallos de autenticación superan 10 en 15 minutos, la IP se bloquea temporalmente.** Si olvidaste el token, verifica el campo `admin_token` en el archivo `wall-vault.yaml`.

---

## Modo distribuido (multi-bot)

Una configuración para **compartir una única bóveda de claves** cuando se ejecuta OpenClaw en múltiples computadoras simultáneamente. Es conveniente porque la gestión de claves se hace en un solo lugar.

### Ejemplo de configuración

```
[Servidor de bóveda de claves]
  wall-vault vault    (bóveda de claves :56243, panel de control)

[WSL Alpha]          [Raspberry Pi Gamma]    [Mac Mini Local]
  wall-vault proxy      wall-vault proxy        wall-vault proxy
  openclaw TUI          openclaw TUI            openclaw TUI
  ↕ sincronización SSE  ↕ sincronización SSE    ↕ sincronización SSE
```

Todos los bots apuntan al servidor central de la bóveda, así que cuando cambias un modelo o agregas una clave en la bóveda, se refleja instantáneamente en todos los bots.

### Paso 1: Iniciar el servidor de bóveda de claves

Ejecuta esto en la computadora que servirá como servidor de la bóveda:

```bash
wall-vault vault
```

### Paso 2: Registrar cada bot (cliente)

Pre-registra la información de cada bot que se conectará al servidor de la bóveda:

```bash
curl -X POST http://localhost:56243/admin/clients \
  -H "Authorization: Bearer TU_TOKEN_ADMIN" \
  -H "Content-Type: application/json" \
  -d '{
    "id": "botA",
    "name": "Bot A",
    "token": "bota-secret",
    "default_service": "google",
    "default_model": "gemini-2.5-flash"
  }'
```

### Paso 3: Iniciar el proxy en cada computadora bot

En cada computadora con un bot, ejecuta el proxy con la dirección del servidor de la bóveda y el token:

```bash
WV_VAULT_URL=http://192.168.x.x:56243 \
WV_VAULT_TOKEN=bota-secret \
WV_VAULT_CLIENT_ID=botA \
wall-vault proxy
```

> 💡 Reemplaza **`192.168.x.x`** con la dirección IP interna real de la computadora del servidor de la bóveda. Puedes verificarla en la configuración de tu router o con el comando `ip addr`.

---

## Configuración de inicio automático

Si es tedioso iniciar manualmente wall-vault cada vez que reinicias tu computadora, regístralo como servicio del sistema. Una vez registrado, se inicia automáticamente al arrancar.

### Linux — systemd (la mayoría de las distribuciones Linux)

systemd es el sistema que inicia y gestiona programas automáticamente en Linux:

```bash
wall-vault doctor deploy
systemctl --user daemon-reload
systemctl --user enable --now wall-vault
```

Ver registros:

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

1. Descarga NSSM de [nssm.cc](https://nssm.cc/download) y agrégalo al PATH.
2. En un PowerShell de administrador:

```powershell
wall-vault doctor deploy windows
```

---

## Doctor (herramienta de diagnóstico)

El comando `doctor` es una herramienta que **auto-diagnostica y repara** la configuración de wall-vault.

```bash
wall-vault doctor check   # Diagnosticar el estado actual (solo lectura, no cambia nada)
wall-vault doctor fix     # Corregir problemas automáticamente
wall-vault doctor all     # Diagnóstico + corrección automática en un paso
```

> 💡 Si algo parece estar mal, prueba primero `wall-vault doctor all`. Detecta automáticamente muchos problemas.

---

## RTK Ahorro de tokens

*(v0.1.24+)*

**RTK (Herramienta de ahorro de tokens)** comprime automáticamente la salida de comandos shell ejecutados por agentes de codificación IA (como Claude Code), reduciendo el uso de tokens. Por ejemplo, 15 líneas de salida de `git status` se comprimen a un resumen de 2 líneas.

### Uso básico

```bash
# Envolver comandos con wall-vault rtk para filtrar automáticamente la salida
wall-vault rtk git status          # muestra solo la lista de archivos modificados
wall-vault rtk git diff HEAD~1     # solo líneas modificadas + contexto mínimo
wall-vault rtk git log -10         # hash + mensaje de una línea
wall-vault rtk go test ./...       # muestra solo las pruebas fallidas
wall-vault rtk ls -la              # los comandos no compatibles se truncan automáticamente
```

### Comandos compatibles y ahorro

| Comando | Método de filtrado | Ahorro |
|---------|-------------------|--------|
| `git status` | Solo resumen de archivos modificados | ~87% |
| `git diff` | Líneas modificadas + 3 líneas de contexto | ~60-94% |
| `git log` | Hash + primera línea del mensaje | ~90% |
| `git push/pull/fetch` | Eliminar progreso, solo resumen | ~80% |
| `go test` | Solo mostrar fallos, contar éxitos | ~88-99% |
| `go build/vet` | Solo mostrar errores | ~90% |
| Todos los demás comandos | Primeras 50 + últimas 50 líneas, máx 32KB | Variable |

### Pipeline de filtrado en 3 etapas

1. **Filtro estructural específico del comando** — Comprende los formatos de salida de git, go, etc. y extrae solo las partes significativas
2. **Post-procesamiento con regex** — Elimina códigos de color ANSI, comprime líneas vacías, agrega líneas duplicadas
3. **Paso directo + truncado** — Los comandos no compatibles solo mantienen las primeras/últimas 50 líneas

### Integración con Claude Code

Puedes configurar todos los comandos shell para que pasen automáticamente por RTK usando el hook `PreToolUse` de Claude Code.

```bash
# Instalar hook (se agrega automáticamente a Claude Code settings.json)
wall-vault rtk hook install
```

O agregar manualmente en `~/.claude/settings.json`:

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

> 💡 **Preservación del código de salida**: RTK devuelve el código de salida del comando original tal cual. Si un comando falla (código de salida ≠ 0), la IA detecta correctamente el fallo.

> 💡 **Salida forzada en inglés**: RTK ejecuta comandos con `LC_ALL=C` para producir siempre salida en inglés independientemente de la configuración de idioma del sistema. Esto asegura que los filtros funcionen correctamente.

---

## Referencia de variables de entorno

Las variables de entorno son una forma de pasar valores de configuración a los programas. Introdúcelas en tu terminal como `export VARIABLE=valor`, o agrégalas a tu archivo de servicio de inicio automático para efecto permanente.

| Variable | Descripción | Valor de ejemplo |
|----------|-------------|-----------------|
| `WV_LANG` | Idioma del panel de control | `ko`, `en`, `ja` |
| `WV_THEME` | Tema del panel de control | `light`, `dark`, `gold` |
| `WV_KEY_GOOGLE` | Clave API de Google (coma para múltiples) | `AIza...,AIza...` |
| `WV_KEY_OPENROUTER` | Clave API de OpenRouter | `sk-or-v1-...` |
| `WV_VAULT_URL` | Dirección del servidor bóveda en modo distribuido | `http://192.168.x.x:56243` |
| `WV_VAULT_TOKEN` | Token de autenticación del cliente (bot) | `my-secret-token` |
| `WV_ADMIN_TOKEN` | Token de administrador | `admin-token-here` |
| `WV_MASTER_PASS` | Contraseña de cifrado de claves API | `my-password` |
| `WV_AVATAR` | Ruta del archivo de imagen del avatar (relativa a `~/.openclaw/`) | `workspace/avatars/avatar.png` |
| `OLLAMA_URL` | Dirección del servidor local Ollama | `http://192.168.x.x:11434` |

---

## Solución de problemas

### Cuando el proxy no inicia

El puerto suele estar en uso por otro programa.

```bash
ss -tlnp | grep 56244   # Verificar qué está usando el puerto 56244
wall-vault proxy --port 8080   # Iniciar con un número de puerto diferente
```

### Errores de clave API (429, 402, 401, 403, 582)

| Código de error | Significado | Solución |
|----------------|------------|----------|
| **429** | Demasiadas solicitudes (límite de uso excedido) | Esperar un momento o agregar más claves |
| **402** | Pago requerido o créditos agotados | Recargar créditos en el servicio correspondiente |
| **401 / 403** | Clave inválida o sin permiso | Verificar el valor de la clave y re-registrar |
| **582** | Sobrecarga del gateway (enfriamiento de 5 minutos) | Se resuelve automáticamente después de 5 minutos |

```bash
# Verificar lista de claves registradas y su estado
curl -H "Authorization: Bearer TU_TOKEN_ADMIN" http://localhost:56243/admin/keys

# Restablecer contadores de uso de claves
curl -X POST -H "Authorization: Bearer TU_TOKEN_ADMIN" http://localhost:56243/admin/keys/reset
```

### Cuando el agente muestra "No conectado"

"No conectado" significa que el proceso del proxy no está enviando señales de heartbeat a la bóveda. **No significa que la configuración no se haya guardado.** El proxy debe estar en ejecución con la dirección del servidor bóveda y el token para mostrarse como conectado.

```bash
# Iniciar proxy con dirección del servidor bóveda, token e ID de cliente
WV_VAULT_URL=http://SERVIDOR_BOVEDA:56243 \
WV_VAULT_TOKEN=token-cliente \
WV_VAULT_CLIENT_ID=id-cliente \
wall-vault proxy
```

Una vez conectado exitosamente, el panel de control muestra 🟢 En ejecución en unos 20 segundos.

### Cuando Ollama no se conecta

Ollama es un programa que ejecuta IA directamente en tu computadora. Primero verifica si Ollama está en ejecución.

```bash
curl http://localhost:11434/api/tags   # Si aparece la lista de modelos, funciona
export OLLAMA_URL=http://192.168.x.x:11434   # Si se ejecuta en otra computadora
```

> ⚠️ Si Ollama no responde, inícialo primero con el comando `ollama serve`.

> ⚠️ **Los modelos grandes son lentos**: Modelos grandes como `qwen3.5:35b` o `deepseek-r1` pueden tardar varios minutos en generar una respuesta. Aunque parezca que no hay respuesta, puede estar procesando normalmente — por favor espera.

---

## Cambios recientes (v0.1.16 ~ v0.1.25)

### v0.1.25 (2026-04-08)
- **Detección de proceso de agente**: El proxy detecta si los agentes locales (NanoClaw/OpenClaw) están activos y muestra un semáforo naranja en el panel de control.
- **Mejora del asa de arrastre**: La ordenación de tarjetas ahora solo funciona desde el área del semáforo (●). Previene el arrastre accidental desde campos de entrada o botones.

### v0.1.24 (2026-04-06)
- **Subcomando RTK de ahorro de tokens**: `wall-vault rtk <command>` filtra automáticamente la salida de comandos shell, reduciendo el uso de tokens de agentes IA en un 60-90%. Filtros integrados para comandos principales como git y go, con truncado automático para comandos no compatibles. Se integra de forma transparente con los hooks `PreToolUse` de Claude Code.

### v0.1.23 (2026-04-06)
- **Corrección del cambio de modelo Ollama**: Se corrigió el problema donde cambiar modelos de Ollama en el panel de la bóveda no se reflejaba realmente en el proxy. Anteriormente solo se usaba la variable de entorno (`OLLAMA_MODEL`), ahora la configuración de la bóveda tiene prioridad.
- **Semáforo automático de servicios locales**: Ollama, LM Studio y vLLM se activan automáticamente cuando están conectables y se desactivan automáticamente al desconectarse. Mismo mecanismo que la alternancia automática de servicios en la nube basada en claves.

### v0.1.22 (2026-04-05)
- **Corrección del campo content vacío**: Se corrigió el problema donde los modelos de pensamiento (gemini-3.1-pro, o1, claude thinking, etc.) que usaban todos los max_tokens en razonamiento sin producir respuestas reales causaban que el proxy omitiera campos `content`/`text` vía `omitempty`, haciendo que los clientes SDK de OpenAI/Anthropic fallaran con errores `Cannot read properties of undefined (reading 'trim')`. Cambiado para siempre incluir campos según la especificación oficial de la API.

### v0.1.21 (2026-04-05)
- **Soporte de modelos Gemma 4**: Modelos Gemma como `gemma-4-31b-it` y `gemma-4-26b-a4b-it` ahora pueden usarse a través de la API de Google Gemini.
- **Soporte oficial de LM Studio / vLLM**: Anteriormente estos servicios faltaban en el enrutamiento del proxy y siempre recurrían a Ollama. Ahora se enrutan correctamente a través de la API compatible con OpenAI.
- **Corrección de visualización de servicios en el panel**: El panel siempre muestra el servicio configurado por el usuario incluso cuando ocurre fallback.
- **Visualización del estado de servicios locales**: Muestra el estado de conexión de servicios locales (Ollama, LM Studio, vLLM, etc.) mediante el color del punto ● al cargar el panel.
- **Variable de entorno del filtro de herramientas**: El modo de paso de herramientas puede configurarse con la variable de entorno `WV_TOOL_FILTER=passthrough`.

### v0.1.20 (2026-03-28)
- **Endurecimiento integral de seguridad**: Prevención XSS (41 puntos), comparación de tokens en tiempo constante, restricción CORS, límites de tamaño de solicitud, prevención de traversal de rutas, autenticación SSE, endurecimiento del limitador de velocidad y 12 mejoras de seguridad en total.

### v0.1.19 (2026-03-27)
- **Detección en línea de Claude Code**: Claude Code ejecutándose sin pasar por el proxy ahora se muestra como en línea en el panel de control.

### v0.1.18 (2026-03-26)
- **Corrección del bloqueo del servicio de fallback**: Después del fallback temporal a Ollama, retorno automático al servicio original cuando se recupera.
- **Mejora de la detección de desconexión**: Las comprobaciones de estado cada 15 segundos hacen más rápida la detección de parada del proxy.

### v0.1.17 (2026-03-25)
- **Ordenar tarjetas con arrastrar y soltar**: Las tarjetas de agente pueden reordenarse arrastrándolas.
- **Botones de aplicación de configuración en línea**: Aparecen botones [⚡ Aplicar configuración] en las tarjetas de agentes sin conexión.
- **Tipo de agente cokacdir agregado**.

### v0.1.16 (2026-03-25)
- **Sincronización bidireccional de modelos**: Cambiar modelos de Cline o Claude Code desde el panel de la bóveda se refleja automáticamente.

---

*Para información más detallada sobre la API, consulta [API.md](API.md).*
