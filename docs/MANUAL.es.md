# Manual de Usuario de wall-vault
*(Last updated: 2026-04-06 — v0.1.24)*

---

## Tabla de contenidos

1. [Que es wall-vault?](#que-es-wall-vault)
2. [Instalacion](#instalacion)
3. [Primeros pasos (asistente de configuracion)](#primeros-pasos)
4. [Registro de claves API](#registro-de-claves-api)
5. [Uso del proxy](#uso-del-proxy)
6. [Panel de control de la boveda](#panel-de-control-de-la-boveda)
7. [Modo distribuido (multi-bot)](#modo-distribuido-multi-bot)
8. [Configuracion de inicio automatico](#configuracion-de-inicio-automatico)
9. [Doctor — Herramienta de autodiagnostico](#doctor--herramienta-de-autodiagnostico)
10. [RTK Ahorro de tokens](#rtk-ahorro-de-tokens)
11. [Referencia de variables de entorno](#referencia-de-variables-de-entorno)
12. [Solucion de problemas](#solucion-de-problemas)

---

## Que es wall-vault?

**wall-vault = un proxy de IA + boveda de claves API para OpenClaw**

Para usar servicios de IA, necesitas **claves API** — piensa en ellas como un **pase digital** que prueba que estas autorizado para usar un servicio en particular. Estos pases tienen limites de uso diario y pueden ser comprometidos si se gestionan incorrectamente.

wall-vault guarda tus pases de forma segura en una boveda cifrada y actua como **proxy (intermediario)** entre OpenClaw y los servicios de IA. En resumen, OpenClaw solo necesita comunicarse con wall-vault — wall-vault maneja todo lo complicado entre bastidores.

Esto es lo que wall-vault hace por ti:

- **Rotacion automatica de claves**: Cuando una clave alcanza su limite o se bloquea temporalmente (enfriamiento), wall-vault cambia silenciosamente a la siguiente clave. OpenClaw sigue funcionando sin interrupcion.
- **Cambio automatico de servicio (fallback)**: Si Google no responde, cambia a OpenRouter. Si eso tambien falla, cambia automaticamente a Ollama, LM Studio o vLLM (IA local) en tu maquina. Tu sesion nunca se interrumpe. Cuando el servicio original se recupera, vuelve automaticamente desde la siguiente solicitud (v0.1.18+, LM Studio/vLLM: v0.1.21+).
- **Sincronizacion en tiempo real (SSE)**: Cambia el modelo en el panel de la boveda y OpenClaw lo refleja en 1-3 segundos. SSE (Server-Sent Events) es una tecnologia donde el servidor envia actualizaciones a los clientes en tiempo real.
- **Notificaciones en tiempo real**: Eventos como agotamiento de claves o caidas de servicio aparecen inmediatamente en la parte inferior del TUI (interfaz de terminal) de OpenClaw.

> 💡 **Claude Code, Cursor y VS Code** tambien pueden conectarse, pero el proposito principal de wall-vault es trabajar junto con OpenClaw.

```
OpenClaw (terminal TUI)
        │
        ▼
  wall-vault proxy (:56244)   ← gestion de claves, enrutamiento, fallback, eventos
        │
        ├─ Google Gemini API
        ├─ OpenRouter API (340+ modelos)
        ├─ Ollama / LM Studio / vLLM (maquina local, ultimo recurso)
        └─ OpenAI / Anthropic API
```

---

## Instalacion

### Linux / macOS

Abre una terminal y pega los siguientes comandos:

```bash
# Linux (PC estandar, servidor — amd64)
curl -L https://github.com/sookmook/wall-vault/releases/latest/download/wall-vault-linux-amd64 \
  -o ~/.local/bin/wall-vault && chmod +x ~/.local/bin/wall-vault

# macOS Apple Silicon (M1/M2/M3 Mac)
curl -L https://github.com/sookmook/wall-vault/releases/latest/download/wall-vault-darwin-arm64 \
  -o /usr/local/bin/wall-vault && chmod +x /usr/local/bin/wall-vault
```

- `curl -L ...` — Descarga un archivo de Internet.
- `chmod +x` — Hace el archivo descargado ejecutable. Si omites este paso, obtendras un error de "permiso denegado".

### Windows

Abre PowerShell (como administrador) y ejecuta los siguientes comandos:

```powershell
# Descargar
Invoke-WebRequest -Uri `
  "https://github.com/sookmook/wall-vault/releases/latest/download/wall-vault-windows-amd64.exe" `
  -OutFile "$env:LOCALAPPDATA\Programs\wall-vault\wall-vault.exe"

# Agregar al PATH (toma efecto despues de reiniciar PowerShell)
$env:PATH += ";$env:LOCALAPPDATA\Programs\wall-vault"
```

> 💡 **Que es el PATH?** Es una lista de carpetas donde tu computadora busca comandos. Agregar wall-vault al PATH te permite ejecutar `wall-vault` desde cualquier directorio.

### Compilacion desde el codigo fuente (para desarrolladores)

Esto solo aplica si tienes un entorno de desarrollo Go instalado.

```bash
git clone https://github.com/sookmook/wall-vault
cd wall-vault
make build       # bin/wall-vault (version: v0.1.24.YYYYMMDD.HHmmss)
make install     # ~/.local/bin/wall-vault
```

> 💡 **Version con marca de tiempo de compilacion**: Cuando compilas con `make build`, la version se genera automaticamente en un formato como `v0.1.24.20260406.225957` que incluye fecha y hora. Si compilas directamente con `go build ./...`, la version simplemente mostrara `"dev"`.

---

## Primeros pasos

### Ejecutar el asistente de configuracion

Despues de la instalacion, asegurate de ejecutar primero el **asistente de configuracion**. El asistente te guiara paso a paso, pidiendo la informacion necesaria.

```bash
wall-vault setup
```

Estos son los pasos del asistente:

```
1. Seleccion de idioma (10 idiomas incluyendo espanol)
2. Seleccion de tema (light / dark / gold / cherry / ocean)
3. Modo de operacion — autonomo (usuario unico) o distribuido (multiples maquinas)
4. Nombre del bot — el nombre mostrado en el panel de control
5. Configuracion de puertos — por defecto: proxy 56244, boveda 56243 (presiona Enter para mantener los valores por defecto)
6. Seleccion de servicios de IA — Google / OpenRouter / Ollama / LM Studio / vLLM
7. Configuracion del filtro de seguridad de herramientas
8. Token de administrador — una contrasena para bloquear las funciones de administracion del panel; puede generarse automaticamente
9. Contrasena de cifrado de claves API — para almacenamiento extra seguro de claves (opcional)
10. Ubicacion de guardado del archivo de configuracion
```

> ⚠️ **Asegurate de recordar tu token de administrador.** Lo necesitaras mas tarde para agregar claves o cambiar configuraciones en el panel de control. Si lo olvidas, tendras que editar manualmente el archivo de configuracion.

Una vez completado el asistente, se crea automaticamente un archivo de configuracion `wall-vault.yaml`.

### Inicio

```bash
wall-vault start
```

Dos servidores se inician simultaneamente:

- **Proxy** (`http://localhost:56244`) — el intermediario que conecta OpenClaw con los servicios de IA
- **Boveda de claves** (`http://localhost:56243`) — gestion de claves API y panel de control web

Abre `http://localhost:56243` en tu navegador para ver el panel de control inmediatamente.

---

## Registro de claves API

Hay cuatro formas de registrar claves API. **Para principiantes, se recomienda el Metodo 1 (variables de entorno).**

### Metodo 1: Variables de entorno (recomendado — el mas simple)

Las variables de entorno son **valores preestablecidos** que un programa lee al iniciar. Simplemente escribe lo siguiente en tu terminal:

```bash
# Registrar una clave de Google Gemini
export WV_KEY_GOOGLE=AIzaSy...

# Registrar una clave de OpenRouter
export WV_KEY_OPENROUTER=sk-or-v1-...

# Iniciar despues del registro
wall-vault start
```

Si tienes multiples claves, separalas con comas. wall-vault las rotara automaticamente (round-robin):

```bash
export WV_KEY_GOOGLE=AIzaSy...,AIzaSy...,AIzaSy...
```

> 💡 **Consejo**: El comando `export` solo aplica a la sesion de terminal actual. Para persistirlo entre reinicios, agrega la linea a tu archivo `~/.bashrc` o `~/.zshrc`.

### Metodo 2: Interfaz del panel de control (apuntar y hacer clic)

1. Abre `http://localhost:56243` en tu navegador
2. Haz clic en el boton `[+ Agregar]` en la tarjeta superior **🔑 Claves API**
3. Ingresa el tipo de servicio, valor de la clave, etiqueta (nombre memo) y limite diario, luego guarda

### Metodo 3: API REST (para automatizacion/scripts)

La API REST es una forma en que los programas intercambian datos via HTTP. Es util para registro automatizado mediante scripts.

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

Usa esto para inyectar temporalmente una clave para pruebas sin registro formal. La clave desaparece cuando el programa se detiene.

```bash
wall-vault proxy --key-google=AIzaSy... --key-openrouter=sk-or-...
```

---

## Uso del proxy

### Uso con OpenClaw (proposito principal)

Aqui se muestra como configurar OpenClaw para conectarse a servicios de IA a traves de wall-vault.

Abre `~/.openclaw/openclaw.json` y agrega lo siguiente:

```json5
// ~/.openclaw/openclaw.json
{
  models: {
    providers: {
      "wall-vault": {
        baseUrl: "http://localhost:56244/v1",
        apiKey: "your-agent-token",   // token de agente de la boveda
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

> 💡 **Metodo mas facil**: Haz clic en el boton **🦞 Copiar configuracion de OpenClaw** en la tarjeta de agente del panel de control — copia un fragmento con el token y la direccion ya completados. Solo pegalo.

**A donde redirige el prefijo `wall-vault/` en los nombres de modelos?**

wall-vault determina automaticamente a que servicio de IA enviar la solicitud basandose en el nombre del modelo:

| Formato del modelo | Redirige a |
|-------------------|-----------|
| `wall-vault/gemini-*` | Google Gemini directo |
| `wall-vault/gpt-*`, `wall-vault/o3`, `wall-vault/o4*` | OpenAI directo |
| `wall-vault/claude-*` | Anthropic via OpenRouter |
| `wall-vault/hunter-alpha`, `wall-vault/healer-alpha` | OpenRouter (contexto 1M tokens gratuito) |
| `wall-vault/kimi-*`, `wall-vault/glm-*`, `wall-vault/deepseek-*` | OpenRouter |
| `google/nombre-modelo`, `openai/nombre-modelo`, `anthropic/nombre-modelo`, etc. | Directo a ese servicio |
| `custom/google/nombre-modelo`, `custom/openai/nombre-modelo`, etc. | Elimina el prefijo `custom/` y redirige |
| `nombre-modelo:cloud` | Elimina el sufijo `:cloud` y redirige a OpenRouter |

> 💡 **Que es el contexto?** Es la cantidad de conversacion que una IA puede recordar a la vez. 1M (un millon de tokens) significa que puede procesar conversaciones o documentos muy largos en una sola sesion.

### Formato directo de API Gemini (para compatibilidad con herramientas existentes)

Si tienes herramientas que ya usan directamente la API de Google Gemini, simplemente cambia la direccion a wall-vault:

```bash
export ANTHROPIC_BASE_URL=http://localhost:56244/google
```

O si la herramienta acepta una URL directa:

```
http://localhost:56244/google/v1beta/models/gemini-2.5-flash:generateContent
```

### Uso con el SDK de OpenAI (Python)

Tambien puedes conectar wall-vault a codigo Python que usa IA. Solo cambia la `base_url`:

```python
from openai import OpenAI

client = OpenAI(
    base_url="http://localhost:56244/v1",
    api_key="not-needed"  # wall-vault gestiona las claves API por ti
)

response = client.chat.completions.create(
    model="google/gemini-2.5-flash",   # usa el formato provider/model
    messages=[{"role": "user", "content": "Hola"}]
)
```

### Cambiar modelos en tiempo de ejecucion

Para cambiar el modelo de IA mientras wall-vault ya esta en ejecucion:

```bash
# Cambiar modelo enviando una solicitud al proxy
curl -X PUT http://localhost:56244/api/config/model \
  -H "Content-Type: application/json" \
  -d '{"service": "openrouter", "model": "anthropic/claude-3.5-sonnet"}'

# En modo distribuido (multi-bot), cambiar en el servidor de la boveda → sincronizado instantaneamente via SSE
curl -X PUT http://localhost:56243/admin/clients/mi-bot-id \
  -H "Authorization: Bearer tu-token-admin" \
  -H "Content-Type: application/json" \
  -d '{"default_service": "google", "default_model": "gemini-2.5-pro"}'
```

### Verificar modelos disponibles

```bash
# Ver lista completa
curl http://localhost:56244/api/models | python3 -m json.tool

# Ver solo modelos de Google
curl "http://localhost:56244/api/models?service=google"

# Buscar por nombre (ej: modelos que contengan "claude")
curl "http://localhost:56244/api/models?q=claude"
```

**Modelos clave por servicio:**

| Servicio | Modelos clave |
|----------|-------------|
| Google | gemini-2.5-pro, gemini-2.5-flash, gemini-2.5-flash-8b, gemini-2.0-flash |
| OpenAI | gpt-4o, gpt-4o-mini, o3, o1, o1-mini |
| OpenRouter | 346+ (Hunter Alpha contexto 1M gratuito, DeepSeek R1/V3, Qwen 2.5, etc.) |
| Ollama | Detecta automaticamente modelos instalados localmente |
| LM Studio | Servidor local (puerto 1234) |
| vLLM | Servidor local (puerto 8000) |

---

## Panel de control de la boveda

Abre `http://localhost:56243` en tu navegador para ver el panel de control.

**Disposicion:**
- **Barra superior (fija)**: Logo, selectores de idioma/tema, indicador de estado de conexion SSE
- **Cuadricula de tarjetas**: Tarjetas de agentes, servicios y claves API dispuestas en mosaico

### Tarjetas de claves API

Estas tarjetas ofrecen una vista rapida de tus claves API registradas.

- Las claves estan organizadas por servicio.
- `today_usage`: Numero de tokens (unidades de texto que la IA lee/escribe) procesados exitosamente hoy
- `today_attempts`: Numero total de llamadas hoy (exitosas + fallidas)
- Usa el boton `[+ Agregar]` para registrar nuevas claves y `✕` para eliminarlas.

> 💡 **Que es un token?** Es la unidad que la IA usa para procesar texto. Aproximadamente una palabra en ingles, o 1-2 caracteres en espanol. Los precios de API se basan tipicamente en el recuento de tokens.

### Tarjetas de agentes

Estas tarjetas muestran el estado de los bots (agentes) conectados al proxy wall-vault.

**El estado de conexion tiene 4 niveles:**

| Indicador | Estado | Significado |
|-----------|--------|------------|
| 🟢 | En ejecucion | El proxy esta operando normalmente |
| 🟡 | Retrasado | Responde pero lento |
| 🔴 | Fuera de linea | El proxy no responde |
| ⚫ | No conectado / Deshabilitado | El proxy nunca se ha conectado a la boveda o esta deshabilitado |

**Botones en la parte inferior de las tarjetas de agentes:**

Cuando registras un agente con un **tipo de agente** especifico, aparecen automaticamente botones de conveniencia correspondientes.

---

#### 🔘 Boton Copiar configuracion — genera automaticamente la configuracion de conexion

Al hacer clic en este boton, se copia al portapapeles un fragmento de configuracion con el token, direccion del proxy e informacion del modelo del agente ya completados. Solo pegalo en la ubicacion mostrada en la tabla para completar la configuracion de conexion.

| Boton | Tipo de agente | Donde pegar |
|-------|---------------|------------|
| 🦞 Copiar config OpenClaw | `openclaw` | `~/.openclaw/openclaw.json` |
| 🦀 Copiar config NanoClaw | `nanoclaw` | `~/.openclaw/openclaw.json` |
| 🟠 Copiar config Claude Code | `claude-code` | `~/.claude/settings.json` |
| ⌨ Copiar config Cursor | `cursor` | Cursor → Settings → AI |
| 💻 Copiar config VSCode | `vscode` | `~/.continue/config.json` |

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

**Ejemplo — Para el tipo Cursor:**

```
Base URL : http://192.168.1.20:56244/v1
API Key  : token-de-este-agente

// O como variables de entorno:
OPENAI_BASE_URL=http://192.168.1.20:56244/v1
OPENAI_API_KEY=token-de-este-agente
```

> ⚠️ **Si copiar al portapapeles no funciona**: Las politicas de seguridad del navegador pueden bloquear la copia. Si aparece un popup con un cuadro de texto, usa Ctrl+A para seleccionar todo, luego Ctrl+C para copiar.

---

#### ⚡ Boton de aplicacion automatica — un clic y listo

Para agentes de tipo `cline`, `claude-code`, `openclaw` o `nanoclaw`, la tarjeta de agente muestra un boton **⚡ Aplicar configuracion**. Al hacer clic, se actualiza automaticamente el archivo de configuracion local del agente.

| Boton | Tipo de agente | Archivo destino |
|-------|---------------|----------------|
| ⚡ Aplicar config Cline | `cline` | `~/.cline/data/globalState.json` + `secrets.json` |
| ⚡ Aplicar config Claude Code | `claude-code` | `~/.claude/settings.json` |
| ⚡ Aplicar config OpenClaw | `openclaw` | `~/.openclaw/openclaw.json` |
| ⚡ Aplicar config NanoClaw | `nanoclaw` | `~/.openclaw/openclaw.json` |

> ⚠️ Este boton envia una solicitud a **localhost:56244** (proxy local). El proxy debe estar en ejecucion en esa maquina para que funcione.

---

#### 🔀 Ordenar tarjetas con arrastrar y soltar (v0.1.17)

Puedes **arrastrar** las tarjetas de agentes en el panel de control para reorganizarlas en cualquier orden.

1. Agarra una tarjeta de agente con el raton y arrastala
2. Sueltala sobre otra tarjeta para intercambiar posiciones
3. El nuevo orden se **guarda inmediatamente en el servidor** y persiste despues de refrescar

> 💡 Los dispositivos tactiles (movil/tablet) aun no son compatibles. Usa un navegador de escritorio.

---

#### 🔄 Sincronizacion bidireccional de modelos (v0.1.16)

Cuando cambias el modelo de un agente en el panel de la boveda, la configuracion local del agente se actualiza automaticamente.

**Para Cline:**
- Cambio de modelo en la boveda → evento SSE → el proxy actualiza el campo del modelo en `globalState.json`
- Campos actualizados: `actModeOpenAiModelId`, `planModeOpenAiModelId`, `openAiModelId`
- `openAiBaseUrl` y la clave API no se tocan
- **Se requiere recargar VS Code (`Ctrl+Alt+R` o `Ctrl+Shift+P` → `Developer: Reload Window`)**
  - Porque Cline no relee el archivo de configuracion durante la ejecucion

**Para Claude Code:**
- Cambio de modelo en la boveda → evento SSE → el proxy actualiza el campo `model` en `settings.json`
- Busca automaticamente en rutas WSL y Windows (`~/.claude/`, `/mnt/c/Users/*/.claude/`)

**Direccion inversa (agente → boveda):**
- Cuando un agente (Cline, Claude Code, etc.) envia una solicitud al proxy, el proxy incluye la informacion de servicio/modelo de ese cliente en el heartbeat
- La tarjeta de agente en el panel de la boveda muestra el servicio/modelo actualmente en uso en tiempo real

> 💡 **Punto clave**: El proxy identifica agentes por su token de Authorization en las solicitudes y redirige automaticamente al servicio/modelo configurado en la boveda. Incluso si Cline o Claude Code envia un nombre de modelo diferente, el proxy lo sobreescribe con la configuracion de la boveda.

---

### Usar Cline con VS Code — Guia detallada

#### Paso 1: Instalar Cline

Instala **Cline** (ID: `saoudrizwan.claude-dev`) desde el Marketplace de Extensiones de VS Code.

#### Paso 2: Registrar el agente en la boveda

1. Abre el panel de la boveda (`http://IP-de-la-boveda:56243`)
2. Haz clic en **+ Agregar** en la seccion **Agentes**
3. Completa lo siguiente:

| Campo | Valor | Descripcion |
|-------|-------|------------|
| ID | `my_cline` | Identificador unico (alfanumerico, sin espacios) |
| Nombre | `My Cline` | Nombre mostrado en el panel de control |
| Tipo de agente | `cline` | ← Debe seleccionar `cline` |
| Servicio | Seleccionar el servicio a usar (ej: `google`) | |
| Modelo | Ingresar el modelo a usar (ej: `gemini-2.5-flash`) | |

4. Haz clic en **Guardar** — se genera automaticamente un token

#### Paso 3: Conectar a Cline

**Metodo A — Aplicacion automatica (recomendado)**

1. Asegurate de que el **proxy** de wall-vault este en ejecucion en esa maquina (`localhost:56244`)
2. Haz clic en el boton **⚡ Aplicar config Cline** en la tarjeta de agente del panel de control
3. Si ves la notificacion "Configuracion aplicada exitosamente!", funciono
4. Recarga VS Code (`Ctrl+Alt+R`)

**Metodo B — Configuracion manual**

Abre Configuracion (⚙️) en la barra lateral de Cline:
- **API Provider**: `OpenAI Compatible`
- **Base URL**: `http://direccion-del-proxy:56244/v1`
  - Misma maquina: `http://localhost:56244/v1`
  - Maquina diferente (ej: Mac Mini): `http://192.168.1.20:56244/v1`
- **API Key**: El token emitido por la boveda (copiar desde la tarjeta de agente)
- **Model ID**: El modelo configurado en la boveda (ej: `gemini-2.5-flash`)

#### Paso 4: Verificar

Envia cualquier mensaje en la ventana de chat de Cline. Si funciona:
- La tarjeta de agente en el panel de la boveda muestra un **punto verde (● En ejecucion)**
- La tarjeta muestra el servicio/modelo actual (ej: `google / gemini-2.5-flash`)

#### Cambiar el modelo

Cuando quieras cambiar el modelo de Cline, hazlo desde el **panel de la boveda**:

1. Cambia el desplegable de servicio/modelo en la tarjeta de agente
2. Haz clic en **Aplicar**
3. Recarga VS Code (`Ctrl+Alt+R`) — el nombre del modelo en el pie de pagina de Cline se actualizara
4. El nuevo modelo se usa a partir de la siguiente solicitud

> 💡 En la practica, el proxy identifica las solicitudes de Cline por el token y las redirige al modelo configurado en la boveda. Incluso sin recargar VS Code, **el modelo realmente utilizado cambia inmediatamente** — la recarga es solo para actualizar la visualizacion del modelo en la interfaz de Cline.

#### Deteccion de desconexion

Cuando VS Code se cierra, la tarjeta de agente en el panel de la boveda se vuelve amarilla (retrasado) despues de unos **90 segundos**, y roja (fuera de linea) despues de **3 minutos**. (Desde v0.1.18, la deteccion fuera de linea es mas rapida gracias a las verificaciones de estado cada 15 segundos.)

#### Solucion de problemas

| Sintoma | Causa | Solucion |
|---------|-------|----------|
| Error "Conexion fallida" en Cline | Proxy no esta en ejecucion o direccion incorrecta | Verificar proxy con `curl http://localhost:56244/health` |
| El punto verde no aparece en la boveda | Clave API (token) no configurada | Hacer clic nuevamente en el boton **⚡ Aplicar config Cline** |
| El modelo del pie de Cline no cambia | Cline cachea la configuracion | Recargar VS Code (`Ctrl+Alt+R`) |
| Se muestra un nombre de modelo incorrecto | Bug antiguo (corregido en v0.1.16) | Actualizar el proxy a v0.1.16 o posterior |

---

#### 🟣 Boton Copiar comando de despliegue — para instalar en una nueva maquina

Usa esto cuando instales por primera vez el proxy wall-vault en una nueva computadora y lo conectes a la boveda. Al hacer clic en el boton se copia el script de instalacion completo. Pegalo en la terminal de la nueva computadora y ejecutalo para:

1. Instalar el binario wall-vault (se omite si ya esta instalado)
2. Registrar automaticamente un servicio de usuario systemd
3. Iniciar el servicio y conectarse automaticamente a la boveda

> 💡 El script ya contiene el token y la direccion del servidor de la boveda de este agente, por lo que puedes ejecutarlo inmediatamente despues de pegar sin ninguna modificacion.

---

### Tarjetas de servicios

Estas tarjetas te permiten habilitar/deshabilitar y configurar servicios de IA.

- Interruptor para habilitar/deshabilitar cada servicio
- Ingresa la direccion de un servidor de IA local (Ollama, LM Studio, vLLM, etc. en tu computadora) para descubrir automaticamente los modelos disponibles
- **Estado de conexion del servicio local**: Un punto ● junto al nombre del servicio es **verde** si esta conectado, **gris** si no
- **Senalizacion automatica de servicios locales** (v0.1.23+): Los servicios locales (Ollama, LM Studio, vLLM) se habilitan/deshabilitan automaticamente segun la disponibilidad de conexion. Cuando un servicio se vuelve accesible, cambia a ● verde y la casilla de verificacion se activa en 15 segundos; cuando el servicio se cae, se deshabilita automaticamente. Esto funciona igual que el cambio automatico de servicios en la nube (Google, OpenRouter, etc.) basado en la disponibilidad de claves API.

> 💡 **Si el servicio local esta en otra computadora**: Ingresa la IP de esa computadora en el campo de URL del servicio. Ejemplo: `http://192.168.1.20:11434` (Ollama), `http://192.168.1.20:1234` (LM Studio). Si el servicio esta vinculado solo a `127.0.0.1` en lugar de `0.0.0.0`, el acceso por IP externa no funcionara — verifica la configuracion de direccion de enlace del servicio.

### Ingreso del token de administrador

Cuando intentes usar funciones importantes como agregar o eliminar claves en el panel de control, aparecera un popup de ingreso del token de administrador. Ingresa el token que configuraste durante el asistente de configuracion. Una vez ingresado, permanece valido hasta que cierres el navegador.

> ⚠️ **Si la autenticacion falla mas de 10 veces en 15 minutos, esa IP sera bloqueada temporalmente.** Si olvidaste tu token, verifica el campo `admin_token` en el archivo `wall-vault.yaml`.

---

## Modo distribuido (multi-bot)

Cuando ejecutas OpenClaw simultaneamente en multiples computadoras, puedes **compartir una sola boveda de claves**. Esto es conveniente porque solo necesitas gestionar claves en un solo lugar.

### Ejemplo de configuracion

```
[Servidor de boveda de claves]
  wall-vault vault    (boveda :56243, panel de control)

[WSL Alpha]           [Raspberry Pi Gamma]    [Mac Mini Local]
  wall-vault proxy      wall-vault proxy        wall-vault proxy
  openclaw TUI          openclaw TUI            openclaw TUI
  ↕ sync SSE            ↕ sync SSE              ↕ sync SSE
```

Todos los bots apuntan al servidor de boveda central. Cuando cambias un modelo o agregas una clave en la boveda, se refleja inmediatamente en todos los bots.

### Paso 1: Iniciar el servidor de boveda

Ejecuta esto en la computadora que servira como servidor de boveda:

```bash
wall-vault vault
```

### Paso 2: Registrar cada bot (cliente)

Registra la informacion de cada bot que se conectara al servidor de boveda:

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

### Paso 3: Iniciar el proxy en cada computadora bot

En cada computadora donde hay un bot instalado, ejecuta el proxy con la direccion del servidor de boveda y el token:

```bash
WV_VAULT_URL=http://192.168.x.x:56243 \
WV_VAULT_TOKEN=bota-secret \
WV_VAULT_CLIENT_ID=botA \
wall-vault proxy
```

> 💡 Reemplaza **`192.168.x.x`** con la direccion IP interna real de la computadora del servidor de boveda. Puedes encontrarla a traves de la configuracion de tu router o el comando `ip addr`.

---

## Configuracion de inicio automatico

Si es tedioso iniciar manualmente wall-vault cada vez que reinicias tu computadora, registralo como servicio del sistema. Una vez registrado, se inicia automaticamente al encender.

### Linux — systemd (la mayoria de las distribuciones Linux)

systemd es el sistema que inicia y gestiona automaticamente programas en Linux:

```bash
wall-vault doctor deploy
systemctl --user daemon-reload
systemctl --user enable --now wall-vault
```

Verificar logs:

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

1. Descarga NSSM desde [nssm.cc](https://nssm.cc/download) y agregalo a tu PATH.
2. En un PowerShell de administrador:

```powershell
wall-vault doctor deploy windows
```

---

## Doctor — Herramienta de autodiagnostico

El comando `doctor` es una herramienta que **diagnostica y repara** automaticamente la configuracion de wall-vault.

```bash
wall-vault doctor check   # Diagnosticar estado actual (solo lectura, no cambia nada)
wall-vault doctor fix     # Reparar problemas automaticamente
wall-vault doctor all     # Diagnosticar + reparar automaticamente en un paso
```

> 💡 Si algo parece mal, intenta ejecutar `wall-vault doctor all` primero. Detecta y corrige muchos problemas automaticamente.

---

## RTK Ahorro de tokens

*(v0.1.24+)*

**RTK (Herramienta de ahorro de tokens)** comprime automaticamente la salida de comandos shell ejecutados por agentes de codificacion IA (como Claude Code), reduciendo el uso de tokens. Por ejemplo, la salida de 15 lineas de `git status` se condensa a un resumen de 2 lineas.

### Uso basico

```bash
# Envuelve comandos con wall-vault rtk para filtrar automaticamente la salida
wall-vault rtk git status          # muestra solo la lista de archivos modificados
wall-vault rtk git diff HEAD~1     # solo lineas modificadas + contexto minimo
wall-vault rtk git log -10         # hash + mensaje en una linea cada uno
wall-vault rtk go test ./...       # muestra solo las pruebas fallidas
wall-vault rtk ls -la              # comandos no soportados se truncan automaticamente
```

### Comandos soportados y ahorros

| Comando | Metodo de filtrado | Ahorro |
|---------|-------------------|--------|
| `git status` | Solo resumen de archivos modificados | ~87% |
| `git diff` | Lineas modificadas + 3 lineas de contexto | ~60-94% |
| `git log` | Hash + primera linea del mensaje | ~90% |
| `git push/pull/fetch` | Eliminar progreso, solo resumen | ~80% |
| `go test` | Solo mostrar fallos, contar aprobados | ~88-99% |
| `go build/vet` | Solo mostrar errores | ~90% |
| Todos los demas comandos | Primeras 50 + ultimas 50 lineas, max 32KB | Variable |

### Pipeline de filtrado en 3 etapas

1. **Filtro estructural por comando** — Entiende los formatos de salida de git, go, etc. y extrae solo las partes significativas
2. **Post-procesamiento con expresiones regulares** — Elimina codigos de color ANSI, colapsa lineas vacias, agrega lineas duplicadas
3. **Paso directo + truncamiento** — Comandos no soportados mantienen solo las primeras/ultimas 50 lineas

### Integracion con Claude Code

Puedes configurar un hook `PreToolUse` de Claude Code para enrutar automaticamente todos los comandos shell a traves de RTK.

```bash
# Instalar hook (se agrega automaticamente al settings.json de Claude Code)
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

> 💡 **Preservacion del codigo de salida**: RTK devuelve el codigo de salida del comando original tal cual. Si un comando falla (codigo de salida ≠ 0), la IA detecta correctamente el fallo.

> 💡 **Salida forzada en ingles**: RTK ejecuta comandos con `LC_ALL=C`, produciendo salida en ingles independientemente de la configuracion de idioma del sistema. Esto asegura que los filtros funcionen correctamente.

---

## Referencia de variables de entorno

Las variables de entorno son una forma de pasar valores de configuracion a un programa. Ingresalas en la terminal con `export VARIABLE=valor`, o agregalas a tu archivo de servicio de inicio automatico para aplicacion permanente.

| Variable | Descripcion | Valor de ejemplo |
|----------|-------------|-----------------|
| `WV_LANG` | Idioma del panel de control | `ko`, `en`, `ja` |
| `WV_THEME` | Tema del panel de control | `light`, `dark`, `gold` |
| `WV_KEY_GOOGLE` | Clave API de Google (separadas por comas) | `AIza...,AIza...` |
| `WV_KEY_OPENROUTER` | Clave API de OpenRouter | `sk-or-v1-...` |
| `WV_VAULT_URL` | Direccion del servidor de boveda en modo distribuido | `http://192.168.x.x:56243` |
| `WV_VAULT_TOKEN` | Token de autenticacion del cliente (bot) | `my-secret-token` |
| `WV_ADMIN_TOKEN` | Token de administrador | `admin-token-here` |
| `WV_MASTER_PASS` | Contrasena de cifrado de claves API | `my-password` |
| `WV_AVATAR` | Ruta del archivo de imagen de avatar (relativa a `~/.openclaw/`) | `workspace/avatars/avatar.png` |
| `OLLAMA_URL` | Direccion del servidor local de Ollama | `http://192.168.x.x:11434` |

---

## Solucion de problemas

### El proxy no inicia

El puerto a menudo ya esta en uso por otro programa.

```bash
ss -tlnp | grep 56244   # Verificar que esta usando el puerto 56244
wall-vault proxy --port 8080   # Iniciar en un puerto diferente
```

### Errores de claves API (429, 402, 401, 403, 582)

| Codigo de error | Significado | Que hacer |
|----------------|------------|-----------|
| **429** | Demasiadas solicitudes (cuota excedida) | Esperar un momento o agregar mas claves |
| **402** | Pago requerido o creditos agotados | Recargar creditos en ese servicio |
| **401 / 403** | Clave invalida o sin permiso | Verificar el valor de la clave y re-registrar |
| **582** | Sobrecarga de gateway (enfriamiento de 5 minutos) | Se resuelve automaticamente despues de 5 minutos |

```bash
# Verificar lista de claves registradas y estado
curl -H "Authorization: Bearer tu-token-admin" http://localhost:56243/admin/keys

# Reiniciar contadores de uso de claves
curl -X POST -H "Authorization: Bearer tu-token-admin" http://localhost:56243/admin/keys/reset
```

### Agente muestra "No conectado"

"No conectado" significa que el proceso proxy no esta enviando heartbeats a la boveda. **No significa que la configuracion no se haya guardado.** El proxy necesita estar ejecutandose con la direccion del servidor de boveda y el token para establecer una conexion.

```bash
# Iniciar proxy con direccion del servidor de boveda, token e ID de cliente
WV_VAULT_URL=http://servidor-boveda:56243 \
WV_VAULT_TOKEN=token-cliente \
WV_VAULT_CLIENT_ID=id-cliente \
wall-vault proxy
```

Una vez conectado, el panel de control mostrara 🟢 En ejecucion en unos 20 segundos.

### Problemas de conexion con Ollama

Ollama es un programa que ejecuta IA directamente en tu computadora. Primero, asegurate de que Ollama este ejecutandose.

```bash
curl http://localhost:11434/api/tags   # Si aparece una lista de modelos, esta funcionando
export OLLAMA_URL=http://192.168.x.x:11434   # Si se ejecuta en otra computadora
```

> ⚠️ Si Ollama no responde, inicialo primero con `ollama serve`.

> ⚠️ **Los modelos grandes son lentos para responder**: Modelos grandes como `qwen3.5:35b` o `deepseek-r1` pueden tardar varios minutos en generar una respuesta. Aunque parezca que no pasa nada, puede estar procesando — por favor ten paciencia.

---

## Cambios recientes (v0.1.16 ~ v0.1.24)

### v0.1.24 (2026-04-06)
- **Subcomando RTK de ahorro de tokens**: `wall-vault rtk <command>` filtra automaticamente la salida de comandos shell para reducir el uso de tokens de agentes IA en un 60-90%. Incluye filtros integrados para comandos clave como git y go, y trunca automaticamente comandos no soportados. Se integra transparentemente con Claude Code via hook `PreToolUse`.

### v0.1.23 (2026-04-06)
- **Correccion del cambio de modelo Ollama**: Se corrigio un problema donde cambiar el modelo de Ollama en el panel de la boveda no se reflejaba en el proxy real. Anteriormente solo se usaba la variable de entorno (`OLLAMA_MODEL`), ahora la configuracion de la boveda tiene prioridad.
- **Senalizacion automatica de servicios locales**: Ollama, LM Studio y vLLM se habilitan automaticamente cuando son accesibles y se deshabilitan cuando no lo son. Funciona igual que el cambio automatico basado en claves para servicios en la nube.

### v0.1.22 (2026-04-05)
- **Correccion del campo content vacio**: Cuando los modelos de pensamiento (gemini-3.1-pro, o1, claude thinking, etc.) usan todos los max_tokens en razonamiento y no pueden producir una respuesta real, el proxy omitia los campos `content`/`text` del JSON de respuesta via `omitempty`, causando que los clientes SDK OpenAI/Anthropic fallaran con `Cannot read properties of undefined (reading 'trim')`. Corregido para siempre incluir los campos segun la especificacion API oficial.

### v0.1.21 (2026-04-05)
- **Soporte del modelo Gemma 4**: Los modelos de la familia Gemma como `gemma-4-31b-it` y `gemma-4-26b-a4b-it` ahora se pueden usar via la API Google Gemini.
- **Soporte de servicios LM Studio / vLLM**: Anteriormente estos servicios faltaban en el enrutamiento del proxy y siempre recaian en Ollama. Ahora se enrutan correctamente via API compatible con OpenAI.
- **Correccion de la visualizacion de servicios del panel**: Incluso cuando ocurre un fallback, el panel siempre muestra el servicio configurado por el usuario.
- **Visualizacion del estado de servicios locales**: Muestra el estado de conexion de servicios locales (Ollama, LM Studio, vLLM, etc.) con colores de puntos ● al cargar el panel.
- **Variable de entorno del filtro de herramientas**: Usa la variable de entorno `WV_TOOL_FILTER=passthrough` para configurar el modo de paso de herramientas.

### v0.1.20 (2026-03-28)
- **Fortalecimiento integral de seguridad**: Prevencion XSS (41 ubicaciones), comparacion de tokens en tiempo constante, restricciones CORS, limites de tamano de solicitud, prevencion de recorrido de ruta, autenticacion SSE, fortalecimiento del limitador de velocidad, y 12 otras mejoras de seguridad.

### v0.1.19 (2026-03-27)
- **Deteccion en linea de Claude Code**: Las instancias de Claude Code que no pasan por el proxy ahora se muestran como en linea en el panel de control.

### v0.1.18 (2026-03-26)
- **Correccion del servicio de fallback atascado**: Despues de un error temporal que causa un fallback a Ollama, regresa automaticamente al servicio original cuando se recupera.
- **Mejora de deteccion fuera de linea**: Las verificaciones de estado cada 15 segundos hacen que la deteccion de caidas del proxy sea mas rapida.

### v0.1.17 (2026-03-25)
- **Ordenar tarjetas con arrastrar y soltar**: Las tarjetas de agentes se pueden arrastrar y soltar para cambiar su orden.
- **Boton de aplicar configuracion en linea**: El boton [⚡ Aplicar configuracion] se muestra en tarjetas de agentes fuera de linea.
- **Tipo de agente cokacdir agregado**.

### v0.1.16 (2026-03-25)
- **Sincronizacion bidireccional de modelos**: Cambiar un modelo Cline o Claude Code en el panel de la boveda se refleja automaticamente.

---

*Para informacion API mas detallada, consulta [API.md](API.md).*
