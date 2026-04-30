# Manual do Usuário wall-vault
*(Last updated: 2026-04-16 — v0.2.2)*

---

## Índice

1. [O que é wall-vault?](#o-que-é-wall-vault)
2. [Instalação](#instalação)
3. [Primeiros passos (assistente setup)](#primeiros-passos)
4. [Registro de API keys](#registro-de-api-keys)
5. [Como usar o proxy](#como-usar-o-proxy)
6. [Dashboard do cofre de chaves](#dashboard-do-cofre-de-chaves)
7. [Modo distribuído (multi-bot)](#modo-distribuído-multi-bot)
8. [Configuração de inicialização automática](#configuração-de-inicialização-automática)
9. [Doctor — diagnóstico](#doctor-diagnóstico)
10. [RTK — economia de tokens](#rtk-economia-de-tokens)
11. [Referência de variáveis de ambiente](#referência-de-variáveis-de-ambiente)
12. [Solução de problemas](#solução-de-problemas)

---

## Notas de atualização v0.2

- `Service` ganhou `default_model` e `allowed_models`. O modelo padrão por serviço agora é definido diretamente no cartão do serviço.
- `Client.default_service` / `default_model` foram renomeados e reinterpretados como `preferred_service` / `model_override`. Se o override estiver vazio, o modelo padrão do serviço é usado.
- Na primeira inicialização da v0.2, o `vault.json` existente é migrado automaticamente, e o estado anterior à migração é preservado como `vault.json.pre-v02.{timestamp}.bak`.
- O dashboard foi reestruturado em três zonas: uma barra lateral esquerda, uma grade de cartões no centro e um painel de edição deslizável no lado direito.
- Os caminhos da Admin API permanecem inalterados, mas os esquemas do corpo da solicitação/resposta foram atualizados — scripts CLI antigos precisarão ser atualizados de acordo.

---

## Novos recursos da v0.2.1

- **Pass-through multimodal (OpenAI → Gemini)**: `/v1/chat/completions` agora aceita seis tipos de partes de conteúdo além de `text` — `input_audio`, `input_video`, `input_image`, `input_file` e `image_url` (data URIs e URLs http(s) externas ≤ 5 MB). O proxy converte cada um para `inlineData` do Gemini. Clientes compatíveis com OpenAI, como o EconoWorld, podem transmitir blobs de áudio / imagem / vídeo diretamente.
- **Tipo de agente EconoWorld**: `POST /agent/apply` com `agentType: "econoworld"` grava as configurações do wall-vault em `analyzer/ai_config.json` do projeto. `workDir` aceita uma lista separada por vírgulas de caminhos candidatos e converte caminhos de unidade do Windows em caminhos de montagem do WSL.
- **Grade de chaves + CRUD no dashboard**: 11 chaves são renderizadas como cartões compactos com slideover de + adicionar / ✕ excluir.
- **Adicionar serviço + reordenação por arrastar e soltar**: a grade de serviços ganha um botão + adicionar e uma alça de arrastar (`⋮⋮`).
- **Cabeçalho / rodapé / animações de tema / seletor de idioma** restaurados. Os 7 temas (cherry/dark/light/ocean/gold/autumn/winter) exibem seu efeito de partículas em uma camada atrás dos cartões, mas acima do fundo.
- **UX de dispensa do slideover**: clique fora ou Esc fecha o slideover.
- **Indicador de status SSE + cronômetro de tempo de atividade** na barra superior (topbar), ao lado do seletor de idioma/tema. O contador `⏱ uptime` e o indicador `● SSE` (verde = conectado, laranja = reconectando, cinza = desconectado) ficam juntos (movidos do rodapé para o cabeçalho a partir de v0.2.18 — status visível sem precisar rolar).

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

## O que é wall-vault?

**wall-vault = Proxy (agente intermediário) de IA + Cofre de API keys para o OpenClaw**

Para usar serviços de IA, você precisa de **API keys**. Uma API key é como um **crachá digital** que prova "esta pessoa tem permissão para usar este serviço". Porém, esses crachás têm um limite diário de uso e há risco de exposição se mal gerenciados.

O wall-vault guarda esses crachás em um cofre seguro e atua como **proxy (agente intermediário)** entre o OpenClaw e os serviços de IA. Em outras palavras, o OpenClaw só precisa se conectar ao wall-vault, e o wall-vault cuida de todo o resto.

Problemas que o wall-vault resolve:

- **Rotação automática de API keys**: Quando o uso de uma chave atinge o limite ou ela é temporariamente bloqueada (cooldown), o sistema muda silenciosamente para a próxima chave. O OpenClaw continua funcionando sem interrupções.
- **Troca automática de serviço (fallback)**: Se o Google não responde, muda para OpenRouter; se isso também não funcionar, muda automaticamente para Ollama, LM Studio ou vLLM (IA local) instalados no seu computador. A sessão não é interrompida. Quando o serviço original se recupera, o sistema volta automaticamente a partir da próxima solicitação (v0.1.18+, LM Studio/vLLM: v0.1.21+).
- **Sincronização em tempo real (SSE)**: Quando você altera o modelo no dashboard do cofre, a mudança é refletida na tela do OpenClaw em 1 a 3 segundos. SSE (Server-Sent Events) é uma tecnologia em que o servidor envia atualizações em tempo real para o cliente.
- **Notificações em tempo real**: Eventos como esgotamento de chaves ou falhas de serviço são exibidos imediatamente na parte inferior da TUI (tela do terminal) do OpenClaw.

> 💡 **Claude Code, Cursor e VS Code** também podem ser conectados, mas o propósito principal do wall-vault é ser usado com o OpenClaw.

```
OpenClaw (tela TUI do terminal)
        │
        ▼
  Proxy wall-vault (:56244)   ← gestão de chaves, roteamento, fallback, eventos
        │
        ├─ Google Gemini API
        ├─ OpenRouter API (mais de 340 modelos)
        ├─ Ollama / LM Studio / vLLM (seu computador, último recurso)
        └─ OpenAI / Anthropic API
```

---

## Instalação

### Linux / macOS

Abra o terminal e cole os comandos abaixo:

```bash
# Linux (PC comum, servidor — amd64)
curl -L https://github.com/sookmook/wall-vault/releases/latest/download/wall-vault-linux-amd64 \
  -o ~/.local/bin/wall-vault && chmod +x ~/.local/bin/wall-vault

# macOS Apple Silicon (Mac M1/M2/M3)
curl -L https://github.com/sookmook/wall-vault/releases/latest/download/wall-vault-darwin-arm64 \
  -o /usr/local/bin/wall-vault && chmod +x /usr/local/bin/wall-vault
```

- `curl -L ...` — Baixa o arquivo da internet.
- `chmod +x` — Torna o arquivo baixado "executável". Se pular esta etapa, ocorrerá um erro de "permissão negada".

### Windows

Abra o PowerShell (como administrador) e execute:

```powershell
# Download
Invoke-WebRequest -Uri `
  "https://github.com/sookmook/wall-vault/releases/latest/download/wall-vault-windows-amd64.exe" `
  -OutFile "$env:LOCALAPPDATA\Programs\wall-vault\wall-vault.exe"

# Adicionar ao PATH (aplica-se após reiniciar o PowerShell)
$env:PATH += ";$env:LOCALAPPDATA\Programs\wall-vault"
```

> 💡 **O que é PATH?** É a lista de pastas onde o computador procura por comandos. Ao adicioná-lo ao PATH, você pode executar `wall-vault` de qualquer pasta.

### Compilar a partir do código-fonte (para desenvolvedores)

Aplicável apenas se você tem o ambiente de desenvolvimento Go instalado.

```bash
git clone https://github.com/sookmook/wall-vault
cd wall-vault
make build       # bin/wall-vault (versão: v0.1.25.YYYYMMDD.HHmmss)
make install     # ~/.local/bin/wall-vault
```

> 💡 **Versão com timestamp de compilação**: Ao compilar com `make build`, a versão é gerada automaticamente no formato `v0.1.27.20260409` incluindo data e hora. Se compilar diretamente com `go build ./...`, a versão mostrará apenas `"dev"`.

---

## Primeiros passos

### Executar o assistente setup

Após a instalação, execute obrigatoriamente o **assistente de configuração** com o comando abaixo. O assistente irá guiá-lo perguntando cada item necessário.

```bash
wall-vault setup
```

Etapas do assistente:

```
1. Seleção de idioma (10 idiomas incluindo português)
2. Seleção de tema (light / dark / gold / cherry / ocean)
3. Modo de operação — usar sozinho (standalone) ou em vários computadores (distributed)
4. Nome do bot — nome que será exibido no dashboard
5. Configuração de portas — padrão: proxy 56244, cofre 56243 (pressione Enter se não quiser alterar)
6. Seleção de serviços de IA — Google / OpenRouter / Ollama / LM Studio / vLLM
7. Configuração do filtro de ferramentas de segurança
8. Token de administrador — senha para bloquear funções de gerenciamento do dashboard. Geração automática disponível
9. Senha de criptografia de API keys — para armazenar chaves de forma mais segura (opcional)
10. Caminho de salvamento do arquivo de configuração
```

> ⚠️ **Lembre-se do token de administrador.** Será necessário para adicionar chaves ou alterar configurações no dashboard. Se esquecê-lo, terá que editar o arquivo de configuração manualmente.

Ao concluir o assistente, o arquivo de configuração `wall-vault.yaml` será gerado automaticamente.

### Execução

```bash
wall-vault start
```

Dois servidores são iniciados simultaneamente:

- **Proxy** (`https://localhost:56244`) — Agente intermediário entre o OpenClaw e os serviços de IA
- **Cofre de chaves** (`https://localhost:56243`) — Gerenciamento de API keys e dashboard web

Abra `https://localhost:56243` no navegador para acessar o dashboard.

---

## Registro de API keys

Existem quatro formas de registrar API keys. **Para iniciantes, recomendamos o método 1 (variáveis de ambiente)**.

### Método 1: Variáveis de ambiente (recomendado — mais simples)

Variáveis de ambiente são **valores pré-configurados** que o programa lê ao iniciar. Digite no terminal:

```bash
# Registrar chave do Google Gemini
export WV_KEY_GOOGLE=AIzaSy...

# Registrar chave do OpenRouter
export WV_KEY_OPENROUTER=sk-or-v1-...

# Iniciar após o registro
wall-vault start
```

Se tiver várias chaves, separe-as com vírgula (,). O wall-vault as utilizará automaticamente em rotação (round-robin):

```bash
export WV_KEY_GOOGLE=AIzaSy...,AIzaSy...,AIzaSy...
```

> 💡 **Dica**: O comando `export` aplica-se apenas à sessão atual do terminal. Para que persista após reiniciar o computador, adicione a linha ao arquivo `~/.bashrc` ou `~/.zshrc`.

### Método 2: Dashboard UI (clique com o mouse)

1. Acesse `https://localhost:56243` no navegador
2. No card **🔑 API Keys** no topo, clique no botão `[+ Adicionar]`
3. Insira o tipo de serviço, valor da chave, rótulo (nome para referência) e limite diário, depois salve

### Método 3: REST API (para automação/scripts)

REST API é a forma de programas trocarem dados via HTTP. Útil para registro automatizado via script.

```bash
curl -X POST https://localhost:56243/admin/keys \
  -H "Authorization: Bearer SEU_TOKEN_ADMIN" \
  -H "Content-Type: application/json" \
  -d '{
    "service": "google",
    "key": "AIzaSy...",
    "label": "Chave principal",
    "daily_limit": 1000
  }'
```

### Método 4: Flag do proxy (para testes rápidos)

Para testar temporariamente com uma chave sem registro formal. A chave desaparece ao encerrar o programa.

```bash
wall-vault proxy --key-google=AIzaSy... --key-openrouter=sk-or-...
```

---

## Como usar o proxy

### Uso com OpenClaw (propósito principal)

Veja como configurar o OpenClaw para se conectar aos serviços de IA através do wall-vault.

Abra o arquivo `~/.openclaw/openclaw.json` e adicione o seguinte:

```json5
// ~/.openclaw/openclaw.json
{
  models: {
    providers: {
      "wall-vault": {
        baseUrl: "https://localhost:56244/v1",
        apiKey: "your-agent-token",   // token do agente vault
        api: "openai-completions",
        models: [
          { id: "wall-vault/gemini-2.5-flash" },
          { id: "wall-vault/gemini-2.5-pro" },
          { id: "wall-vault/hunter-alpha" },    // 1M context gratuito
          { id: "wall-vault/claude-opus-4-6" }
        ]
      }
    }
  }
}
```

> 💡 **Forma mais fácil**: Clique no botão **🦞 Copiar config OpenClaw** no card do agente no dashboard. Um snippet com o token e endereço já preenchidos será copiado para a área de transferência. Basta colar.

**Para onde o prefixo `wall-vault/` no nome do modelo direciona?**

O wall-vault determina automaticamente para qual serviço de IA enviar a solicitação com base no nome do modelo:

| Formato do modelo | Serviço conectado |
|----------|--------------|
| `wall-vault/gemini-*` | Conexão direta com Google Gemini |
| `wall-vault/gpt-*`, `wall-vault/o3`, `wall-vault/o4*` | Conexão direta com OpenAI |
| `wall-vault/claude-*` | Anthropic via OpenRouter |
| `wall-vault/hunter-alpha`, `wall-vault/healer-alpha` | OpenRouter (1 milhão de tokens de contexto gratuito) |
| `wall-vault/kimi-*`, `wall-vault/glm-*`, `wall-vault/deepseek-*` | Conexão via OpenRouter |
| `google/modelo`, `openai/modelo`, `anthropic/modelo` etc. | Conexão direta com o serviço correspondente |
| `custom/google/modelo`, `custom/openai/modelo` etc. | Remove o prefixo `custom/` e redireciona |
| `modelo:cloud` | Remove o sufixo `:cloud` e conecta via OpenRouter |

> 💡 **O que é contexto (context)?** É o volume de conversa que a IA pode "lembrar" de uma vez. 1M (um milhão de tokens) significa que conversas muito longas ou documentos extensos podem ser processados de uma vez.

### Conexão direta no formato Gemini API (compatibilidade com ferramentas existentes)

Se você tem uma ferramenta que usa a API do Google Gemini diretamente, basta alterar o endereço para o wall-vault:

```bash
export ANTHROPIC_BASE_URL=https://localhost:56244/google
```

Ou, se a ferramenta especifica a URL diretamente:

```
https://localhost:56244/google/v1beta/models/gemini-2.5-flash:generateContent
```

### Uso com OpenAI SDK (Python)

Você pode conectar o wall-vault em código Python que utiliza IA. Basta alterar o `base_url`:

```python
from openai import OpenAI

client = OpenAI(
    base_url="https://localhost:56244/v1",
    api_key="not-needed"  # O wall-vault gerencia as API keys automaticamente
)

response = client.chat.completions.create(
    model="google/gemini-2.5-flash",   # Formato provider/model
    messages=[{"role": "user", "content": "Olá"}]
)
```

### Alterar modelo durante a execução

Para alterar o modelo de IA com o wall-vault já em execução:

```bash
# Alterar modelo solicitando diretamente ao proxy
curl -X PUT https://localhost:56244/api/config/model \
  -H "Content-Type: application/json" \
  -d '{"service": "openrouter", "model": "anthropic/claude-3.5-sonnet"}'

# No modo distribuído (multi-bot), altere no servidor do cofre → refletido instantaneamente via SSE
curl -X PUT https://localhost:56243/admin/clients/MEU-BOT-ID \
  -H "Authorization: Bearer SEU_TOKEN_ADMIN" \
  -H "Content-Type: application/json" \
  -d '{"default_service": "google", "default_model": "gemini-2.5-pro"}'
```

### Verificar modelos disponíveis

```bash
# Ver lista completa
curl https://localhost:56244/api/models | python3 -m json.tool

# Ver apenas modelos do Google
curl "https://localhost:56244/api/models?service=google"

# Pesquisar por nome (ex.: modelos contendo "claude")
curl "https://localhost:56244/api/models?q=claude"
```

**Resumo dos principais modelos por serviço:**

| Serviço | Principais modelos |
|--------|----------|
| Google | gemini-2.5-pro, gemini-2.5-flash, gemini-2.5-flash-8b, gemini-2.0-flash |
| OpenAI | gpt-4o, gpt-4o-mini, o3, o1, o1-mini |
| OpenRouter | Mais de 346 (Hunter Alpha 1M contexto gratuito, DeepSeek R1/V3, Qwen 2.5, etc.) |
| Ollama | Detecção automática de servidores locais instalados no seu computador |
| LM Studio | Servidor local no seu computador (porta 1234) |
| vLLM | Servidor local no seu computador (porta 8000) |
| llama.cpp | Servidor local no seu computador (porta 8080) |

---

## Dashboard do cofre de chaves

Acesse `https://localhost:56243` no navegador para ver o dashboard.

**Layout da tela:**
- **Barra superior fixa (topbar)**: Logo, seletor de idioma/tema, indicador de conexão SSE
- **Grade de cards**: Cards de agentes, serviços e API keys organizados em formato de tiles

### Card de API keys

Card para gerenciar todas as API keys registradas de forma visual.

- Mostra a lista de chaves separadas por serviço.
- `today_usage`: Tokens processados com sucesso hoje (quantidade de caracteres lidos e escritos pela IA)
- `today_attempts`: Total de chamadas hoje (sucesso + falha)
- Botão `[+ Adicionar]` para registrar novas chaves, `✕` para excluir chaves.

> 💡 **O que é um token?** É a unidade que a IA usa para processar texto. Corresponde aproximadamente a uma palavra em inglês ou 1-2 caracteres em outros idiomas. O custo da API geralmente é calculado com base na quantidade de tokens.

### Card de agentes

Card que mostra o status dos bots (agentes) conectados ao proxy wall-vault.

**O status de conexão é exibido em 4 níveis:**

| Indicador | Status | Significado |
|------|------|------|
| 🟢 | Em execução | O proxy está funcionando normalmente |
| 🟡 | Atraso | Responde, mas com lentidão |
| 🔴 | Offline | O proxy não responde |
| ⚫ | Desconectado/Inativo | O proxy nunca se conectou ao cofre ou está desativado |

**Guia dos botões na parte inferior do card do agente:**

Ao registrar um agente, se você especificar o **tipo de agente**, botões de conveniência correspondentes aparecerão automaticamente.

---

#### 🔘 Botão Copiar Configuração — cria a configuração de conexão automaticamente

Ao clicar no botão, um snippet de configuração com o token, endereço do proxy e informações do modelo já preenchidos é copiado para a área de transferência. Basta colar no local indicado na tabela abaixo para completar a configuração.

| Botão | Tipo de agente | Onde colar |
|------|-------------|-------------|
| 🦞 Copiar config OpenClaw | `openclaw` | `~/.openclaw/openclaw.json` |
| 🦀 Copiar config NanoClaw | `nanoclaw` | `~/.openclaw/openclaw.json` |
| 🟠 Copiar config Claude Code | `claude-code` | `~/.claude/settings.json` |
| ⌨ Copiar config Cursor | `cursor` | Cursor → Settings → AI |
| 💻 Copiar config VSCode | `vscode` | `~/.continue/config.json` |

**Exemplo — Para o tipo Claude Code, este conteúdo é copiado:**

```json
// ~/.claude/settings.json
{
  "apiProvider": "openai",
  "baseUrl": "http://192.168.1.20:56244/v1",
  "apiKey": "token-deste-agente"
}
```

**Exemplo — Para o tipo VSCode (Continue):**

```yaml
# ~/.continue/config.yaml  ← Cole em config.yaml, não em config.json
name: My Config
version: 0.0.1
schema: v1

models:
  - name: wall-vault proxy
    provider: openai
    model: gemini-2.5-flash
    apiBase: http://192.168.1.20:56244/v1
    apiKey: token-deste-agente
    roles:
      - chat
      - edit
      - apply
```

> ⚠️ **As versões mais recentes do Continue usam `config.yaml`.** Se `config.yaml` existir, `config.json` será completamente ignorado. Certifique-se de colar em `config.yaml`.

**Exemplo — Para o tipo Cursor:**

```
Base URL : http://192.168.1.20:56244/v1
API Key  : token-deste-agente

// Ou variáveis de ambiente:
OPENAI_BASE_URL=http://192.168.1.20:56244/v1
OPENAI_API_KEY=token-deste-agente
```

> ⚠️ **Se a cópia para a área de transferência não funcionar**: A política de segurança do navegador pode bloquear a cópia. Se aparecer um popup com caixa de texto, selecione tudo com Ctrl+A e copie com Ctrl+C.

---

#### ⚡ Botão de Aplicação Automática — um clique e a configuração está pronta

Para agentes do tipo `cline`, `claude-code`, `openclaw` ou `nanoclaw`, o card do agente exibe o botão **⚡ Aplicar configuração**. Ao clicar, o arquivo de configuração local do agente é atualizado automaticamente.

| Botão | Tipo de agente | Arquivo alvo |
|------|-------------|-------------|
| ⚡ Aplicar config Cline | `cline` | `~/.cline/data/globalState.json` + `secrets.json` |
| ⚡ Aplicar config Claude Code | `claude-code` | `~/.claude/settings.json` |
| ⚡ Aplicar config OpenClaw | `openclaw` | `~/.openclaw/openclaw.json` |
| ⚡ Aplicar config NanoClaw | `nanoclaw` | `~/.openclaw/openclaw.json` |

> ⚠️ Este botão envia solicitações para **localhost:56244** (proxy local). O proxy deve estar em execução naquela máquina para funcionar.

---

#### 🔀 Ordenação de cards por arrastar e soltar (v0.1.17, melhorado v0.1.25)

Você pode **arrastar** os cards de agentes no dashboard para reorganizá-los na ordem desejada.

1. Segure e arraste a área do **semáforo (●)** no canto superior esquerdo do card
2. Solte sobre o card na posição desejada para trocar a ordem

> 💡 O corpo do card (campos de entrada, botões, etc.) não pode ser arrastado. Só é possível segurar pela área do semáforo.

#### 🟠 Detecção de processo do agente (v0.1.25)

Quando o proxy está funcionando normalmente, mas o processo do agente local (NanoClaw, OpenClaw) morreu, o semáforo do card muda para **laranja (piscando)** e exibe a mensagem "Processo do agente parado".

- 🟢 Verde: Proxy + agente normais
- 🟠 Laranja (piscando): Proxy normal, agente morto
- 🔴 Vermelho: Proxy offline
3. A ordem alterada é **salva imediatamente no servidor** e persiste mesmo após atualizar a página

> 💡 Em dispositivos touch (mobile/tablet), isso ainda não é suportado. Use um navegador desktop.

---

#### 🔄 Sincronização bidirecional de modelos (v0.1.16)

Quando você altera o modelo de um agente no dashboard do cofre, a configuração local daquele agente é atualizada automaticamente.

**Para o Cline:**
- Alterar modelo no cofre → evento SSE → proxy atualiza o campo de modelo no `globalState.json`
- Campos atualizados: `actModeOpenAiModelId`, `planModeOpenAiModelId`, `openAiModelId`
- `openAiBaseUrl` e API key não são modificados
- **É necessário recarregar o VS Code (`Ctrl+Alt+R` ou `Ctrl+Shift+P` → `Developer: Reload Window`)**
  - O Cline não relê o arquivo de configuração durante a execução

**Para o Claude Code:**
- Alterar modelo no cofre → evento SSE → proxy atualiza o campo `model` no `settings.json`
- Busca automática de caminhos em WSL e Windows (`~/.claude/`, `/mnt/c/Users/*/.claude/`)

**Direção inversa (agente → cofre):**
- Quando um agente (Cline, Claude Code, etc.) envia uma solicitação ao proxy, o proxy inclui as informações de serviço/modelo daquele cliente no heartbeat
- O card do agente no dashboard do cofre mostra o serviço/modelo em uso em tempo real

> 💡 **Ponto-chave**: O proxy identifica o agente pelo token de Authorization da solicitação e faz o roteamento automático para o serviço/modelo configurado no cofre. Mesmo que o Cline ou Claude Code envie um nome de modelo diferente, o proxy sobrepõe com a configuração do cofre.

---

### Usando o Cline no VS Code — Guia detalhado

#### Etapa 1: Instalar o Cline

Instale o **Cline** (ID: `saoudrizwan.claude-dev`) no marketplace de extensões do VS Code.

#### Etapa 2: Registrar o agente no cofre

1. Abra o dashboard do cofre (`http://IP-DO-COFRE:56243`)
2. Na seção **Agentes**, clique em **+ Adicionar**
3. Preencha da seguinte forma:

| Campo | Valor | Descrição |
|------|----|------|
| ID | `meu_cline` | Identificador único (alfanumérico, sem espaços) |
| Nome | `Meu Cline` | Nome exibido no dashboard |
| Tipo de agente | `cline` | ← Selecione obrigatoriamente `cline` |
| Serviço | Selecione o serviço (ex.: `google`) | |
| Modelo | Digite o modelo (ex.: `gemini-2.5-flash`) | |

4. Ao clicar em **Salvar**, o token será gerado automaticamente

#### Etapa 3: Conectar ao Cline

**Método A — Aplicação automática (recomendado)**

1. Verifique se o **proxy** wall-vault está em execução nesta máquina (`localhost:56244`)
2. Clique no botão **⚡ Aplicar config Cline** no card do agente no dashboard
3. Sucesso quando aparecer a notificação "Configuração aplicada com sucesso!"
4. Recarregue o VS Code (`Ctrl+Alt+R`)

**Método B — Configuração manual**

Na barra lateral do Cline, abra configurações (⚙️):
- **API Provider**: `OpenAI Compatible`
- **Base URL**: `http://ENDEREÇO-DO-PROXY:56244/v1`
  - Na mesma máquina: `https://localhost:56244/v1`
  - Em outra máquina (ex.: servidor Mini): `http://192.168.1.20:56244/v1`
- **API Key**: Token emitido pelo cofre (copie do card do agente)
- **Model ID**: Modelo configurado no cofre (ex.: `gemini-2.5-flash`)

#### Etapa 4: Verificação

Envie qualquer mensagem no chat do Cline. Se estiver funcionando:
- O card do agente no dashboard do cofre mostrará **ponto verde (● Em execução)**
- O card exibirá o serviço/modelo atual (ex.: `google / gemini-2.5-flash`)

#### Alterar modelo

Para alterar o modelo do Cline, faça pelo **dashboard do cofre**:

1. Altere o dropdown de serviço/modelo no card do agente
2. Clique em **Aplicar**
3. Recarregue o VS Code (`Ctrl+Alt+R`) — o nome do modelo no rodapé do Cline será atualizado
4. O novo modelo será usado a partir da próxima solicitação

> 💡 Na prática, o proxy identifica as solicitações do Cline pelo token e faz o roteamento para o modelo configurado no cofre. Mesmo sem recarregar o VS Code, **o modelo realmente utilizado muda imediatamente** — o reload é para atualizar a exibição do modelo na interface do Cline.

#### Detecção de desconexão

Ao fechar o VS Code, no dashboard do cofre, o card do agente muda para amarelo (atraso) após aproximadamente **90 segundos** e para vermelho (offline) após **3 minutos**. (A partir do v0.1.18, a detecção de offline ficou mais rápida com verificações de status a cada 15 segundos.)

#### Solução de problemas

| Sintoma | Causa | Solução |
|------|------|------|
| Erro "Falha na conexão" no Cline | Proxy não está em execução ou endereço incorreto | Verifique o proxy com `curl https://localhost:56244/health` |
| Ponto verde não aparece no cofre | API key (token) não configurada | Clique novamente no botão **⚡ Aplicar config Cline** |
| Modelo no rodapé do Cline não muda | Cline está com configuração em cache | Recarregue o VS Code (`Ctrl+Alt+R`) |
| Nome de modelo incorreto exibido | Bug antigo (corrigido no v0.1.16) | Atualize o proxy para v0.1.16 ou superior |

---

#### 🟣 Botão Copiar Comando de Deploy — para instalar em uma nova máquina

Usado ao instalar o proxy wall-vault pela primeira vez em um novo computador e conectá-lo ao cofre. Ao clicar no botão, todo o script de instalação é copiado. Cole no terminal do novo computador e execute para que tudo seja processado de uma vez:

1. Instalação do binário wall-vault (pula se já instalado)
2. Registro automático do serviço systemd do usuário
3. Início do serviço e conexão automática ao cofre

> 💡 O script já vem com o token deste agente e o endereço do servidor do cofre preenchidos, então pode ser executado imediatamente após colar, sem modificações adicionais.

---

### Card de serviços

Card para ativar/desativar ou configurar os serviços de IA.

- Switch toggle para ativar/desativar cada serviço
- Ao inserir o endereço de servidores de IA locais (Ollama, LM Studio, vLLM, llama.cpp, etc., executados no seu computador), os modelos disponíveis são detectados automaticamente.
- **Indicador de status de conexão de serviço local**: O ponto ● ao lado do nome do serviço fica **verde** quando conectado, **cinza** quando não conectado
- **Semáforo automático de serviço local** (v0.1.23+): Serviços locais (Ollama, LM Studio, vLLM, llama.cpp) são ativados/desativados automaticamente com base na disponibilidade de conexão. Ao ativar um serviço, em até 15 segundos o ponto ● fica verde e a checkbox é ativada; ao desativar o serviço, ele é desligado automaticamente. Funciona da mesma forma que o toggle automático de serviços cloud (Google, OpenRouter, etc.) baseado na presença de API keys.
- **Toggle do modo de raciocínio** (v0.2.17+): Uma checkbox **modo de raciocínio** aparece na parte inferior do formulário de edição dos serviços locais. Quando ativada, o proxy adiciona `"reasoning": true` ao corpo das chat-completions enviado ao upstream, permitindo que modelos compatíveis com a saída do processo de raciocínio — como DeepSeek R1 ou Qwen QwQ — retornem também blocos `<think>…</think>`. Servidores que não conhecem esse campo simplesmente o ignoram, então você pode deixá-la ativada com segurança mesmo em workloads mistas.

> 💡 **Se o serviço local está rodando em outro computador**: Insira o IP daquele computador no campo de URL do serviço. Ex.: `http://192.168.1.20:11434` (Ollama), `http://192.168.1.20:1234` (LM Studio), `http://192.168.1.20:8080` (llama.cpp). Se o serviço estiver vinculado apenas a `127.0.0.1` e não a `0.0.0.0`, o acesso por IP externo não funcionará — verifique o endereço de vinculação nas configurações do serviço.

### Entrada do token de administrador

Quando tentar usar funções importantes como adicionar/excluir chaves no dashboard, aparecerá um popup solicitando o token de administrador. Insira o token configurado no assistente setup. Uma vez inserido, permanece válido até fechar o navegador.

> ⚠️ **Se houver mais de 10 falhas de autenticação em 15 minutos, o IP será temporariamente bloqueado.** Se esqueceu o token, verifique o item `admin_token` no arquivo `wall-vault.yaml`.

---

## Modo distribuído (multi-bot)

Quando vários computadores executam o OpenClaw simultaneamente, esta é a configuração para **compartilhar um único cofre de chaves**. É conveniente pois o gerenciamento de chaves é feito em um só lugar.

### Exemplo de configuração

```
[Servidor do cofre de chaves]
  wall-vault vault    (cofre de chaves :56243, dashboard)

[WSL Alpha]            [Raspberry Pi Gamma]    [Mac Mini Local]
  wall-vault proxy      wall-vault proxy        wall-vault proxy
  openclaw TUI          openclaw TUI            openclaw TUI
  ↕ Sinc. SSE           ↕ Sinc. SSE             ↕ Sinc. SSE
```

Todos os bots olham para o servidor do cofre central, então ao alterar modelos ou adicionar chaves no cofre, as mudanças são refletidas instantaneamente em todos os bots.

### Etapa 1: Iniciar o servidor do cofre de chaves

Execute no computador que será o servidor do cofre:

```bash
wall-vault vault
```

### Etapa 2: Registrar cada bot (cliente)

Registre previamente as informações de cada bot que se conectará ao servidor do cofre:

```bash
curl -X POST https://localhost:56243/admin/clients \
  -H "Authorization: Bearer SEU_TOKEN_ADMIN" \
  -H "Content-Type: application/json" \
  -d '{
    "id": "botA",
    "name": "Bot A",
    "token": "bota-secret",
    "default_service": "google",
    "default_model": "gemini-2.5-flash"
  }'
```

### Etapa 3: Iniciar o proxy em cada computador do bot

Execute o proxy em cada computador onde o bot está instalado, especificando o endereço do servidor do cofre e o token:

```bash
WV_VAULT_URL=http://192.168.x.x:56243 \
WV_VAULT_TOKEN=bota-secret \
WV_VAULT_CLIENT_ID=botA \
wall-vault proxy
```

> 💡 Substitua **`192.168.x.x`** pelo endereço IP interno real do computador do servidor do cofre. Você pode verificá-lo nas configurações do roteador ou com o comando `ip addr`.

---

## Configuração de inicialização automática

Se for incômodo ligar o wall-vault manualmente toda vez que reiniciar o computador, registre-o como serviço do sistema. Uma vez registrado, ele inicia automaticamente ao ligar.

### Linux — systemd (maioria das distribuições Linux)

O systemd é o sistema que gerencia a inicialização e operação automática de programas no Linux:

```bash
wall-vault doctor deploy
systemctl --user daemon-reload
systemctl --user enable --now wall-vault
```

Ver logs:

```bash
journalctl --user -u wall-vault -f
```

### macOS — launchd

É o sistema responsável pela execução automática de programas no macOS:

```bash
wall-vault doctor deploy launchd
launchctl load ~/Library/LaunchAgents/com.wall-vault.plist
```

### Windows — NSSM

1. Baixe o NSSM de [nssm.cc](https://nssm.cc/download) e adicione ao PATH.
2. No PowerShell como administrador:

```powershell
wall-vault doctor deploy windows
```

---

## Doctor — diagnóstico

O comando `doctor` é uma ferramenta que **diagnostica e corrige automaticamente** se o wall-vault está configurado corretamente.

```bash
wall-vault doctor check   # Diagnostica o estado atual (apenas leitura, não altera nada)
wall-vault doctor fix     # Repara problemas automaticamente
wall-vault doctor all     # Diagnóstico + reparo automático de uma vez
```

> 💡 Se algo parecer estranho, execute `wall-vault doctor all` primeiro. Ele resolve muitos problemas automaticamente.

---

## RTK — economia de tokens

*(v0.1.24+)*

**RTK (ferramenta de economia de tokens)** comprime automaticamente a saída de comandos shell executados por agentes de codificação IA (como Claude Code), reduzindo o consumo de tokens. Por exemplo, 15 linhas de saída do `git status` são resumidas em 2 linhas.

### Uso básico

```bash
# Envolva o comando com wall-vault rtk para filtrar a saída automaticamente
wall-vault rtk git status          # Mostra apenas a lista de arquivos alterados
wall-vault rtk git diff HEAD~1     # Apenas linhas alteradas + contexto mínimo
wall-vault rtk git log -10         # Uma linha por commit: hash + mensagem
wall-vault rtk go test ./...       # Mostra apenas testes que falharam
wall-vault rtk ls -la              # Comandos não suportados são automaticamente truncados
```

### Comandos suportados e economia

| Comando | Método de filtragem | Taxa de economia |
|------|----------|--------|
| `git status` | Apenas resumo de arquivos alterados | ~87% |
| `git diff` | Linhas alteradas + 3 linhas de contexto | ~60-94% |
| `git log` | Hash + primeira linha da mensagem | ~90% |
| `git push/pull/fetch` | Remove progresso, apenas resumo | ~80% |
| `go test` | Mostra apenas falhas, conta aprovados | ~88-99% |
| `go build/vet` | Mostra apenas erros | ~90% |
| Todos os outros comandos | Primeiras 50 linhas + últimas 50 linhas, máximo 32KB | Variável |

### Pipeline de filtro em 3 etapas

1. **Filtro estrutural por comando** — Entende o formato de saída de git, go, etc. e extrai apenas as partes significativas
2. **Pós-processamento com regex** — Remove códigos de cor ANSI, comprime linhas vazias, agrega linhas duplicadas
3. **Passthrough + truncamento** — Comandos não suportados mantêm apenas as primeiras/últimas 50 linhas

### Integração com Claude Code

Você pode configurar o hook `PreToolUse` do Claude Code para que todos os comandos shell passem automaticamente pelo RTK.

```bash
# Instalar hook (adicionado automaticamente ao settings.json do Claude Code)
wall-vault rtk hook install
```

Ou adicione manualmente em `~/.claude/settings.json`:

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

> 💡 **Preservação do exit code**: O RTK retorna o exit code original do comando. Se o comando falhar (exit code ≠ 0), a IA detecta a falha corretamente.

> 💡 **Saída forçada em inglês**: O RTK executa comandos com `LC_ALL=C` para sempre gerar saída em inglês, independente das configurações de idioma do sistema. Isso garante que os filtros funcionem corretamente.

---

## Referência de variáveis de ambiente

Variáveis de ambiente são uma forma de passar valores de configuração para o programa. Digite `export NOME=valor` no terminal ou coloque no arquivo de serviço de inicialização automática para que seja sempre aplicado.

| Variável | Descrição | Valor exemplo |
|------|------|---------|
| `WV_LANG` | Idioma do dashboard | `ko`, `en`, `ja` |
| `WV_THEME` | Tema do dashboard | `light`, `dark`, `gold` |
| `WV_KEY_GOOGLE` | API key do Google (múltiplas separadas por vírgula) | `AIza...,AIza...` |
| `WV_KEY_OPENROUTER` | API key do OpenRouter | `sk-or-v1-...` |
| `WV_VAULT_URL` | Endereço do servidor do cofre no modo distribuído | `http://192.168.x.x:56243` |
| `WV_VAULT_TOKEN` | Token de autenticação do cliente (bot) | `my-secret-token` |
| `WV_ADMIN_TOKEN` | Token de administrador | `admin-token-here` |
| `WV_MASTER_PASS` | Senha de criptografia de API keys | `my-password` |
| `WV_AVATAR` | Caminho do arquivo de avatar (caminho relativo a `~/.openclaw/`) | `workspace/avatars/avatar.png` |
| `OLLAMA_URL` | Endereço do servidor local Ollama | `http://192.168.x.x:11434` |

---

## Solução de problemas

### Quando o proxy não inicia

Na maioria dos casos, a porta já está sendo usada por outro programa.

```bash
ss -tlnp | grep 56244   # Verificar quem está usando a porta 56244
wall-vault proxy --port 8080   # Iniciar com outro número de porta
```

### Quando ocorrem erros de API key (429, 402, 401, 403, 582)

| Código de erro | Significado | Como resolver |
|----------|------|----------|
| **429** | Muitas solicitações (cota excedida) | Aguarde ou adicione outra chave |
| **402** | Pagamento necessário ou créditos insuficientes | Recarregue créditos no serviço |
| **401 / 403** | Chave incorreta ou sem permissão | Verifique o valor da chave e registre novamente |
| **582** | Sobrecarga do gateway (cooldown de 5 min) | Libera automaticamente após 5 minutos |

```bash
# Verificar lista de chaves registradas e status
curl -H "Authorization: Bearer SEU_TOKEN_ADMIN" https://localhost:56243/admin/keys

# Resetar contadores de uso das chaves
curl -X POST -H "Authorization: Bearer SEU_TOKEN_ADMIN" https://localhost:56243/admin/keys/reset
```

### Quando o agente aparece como "Desconectado"

"Desconectado" significa que o processo do proxy não está enviando sinais (heartbeat) ao cofre. **Não significa que as configurações não foram salvas.** O proxy precisa estar em execução conhecendo o endereço do servidor do cofre e o token para mudar para o status conectado.

```bash
# Iniciar proxy especificando endereço do servidor do cofre, token e ID do cliente
WV_VAULT_URL=http://ENDEREÇO-DO-COFRE:56243 \
WV_VAULT_TOKEN=TOKEN-DO-CLIENTE \
WV_VAULT_CLIENT_ID=ID-DO-CLIENTE \
wall-vault proxy
```

Se a conexão for bem-sucedida, em aproximadamente 20 segundos o dashboard mostrará 🟢 Em execução.

### Quando o Ollama não conecta

Ollama é um programa que executa IA diretamente no seu computador. Primeiro, verifique se o Ollama está em execução.

```bash
curl http://localhost:11434/api/tags   # Se mostrar a lista de modelos, está normal
export OLLAMA_URL=http://192.168.x.x:11434   # Se estiver rodando em outro computador
```

> ⚠️ Se o Ollama não responder, inicie-o primeiro com o comando `ollama serve`.

> ⚠️ **Modelos grandes são lentos**: Modelos grandes como `qwen3.5:35b` e `deepseek-r1` podem levar vários minutos para gerar uma resposta. Mesmo que pareça que não há resposta, pode estar processando normalmente — aguarde.

---

## Alterações recentes (v0.1.16 ~ v0.1.27)

### v0.1.27 (2026-04-09)
- **Correção do nome do modelo no fallback para Ollama**: Corrigido problema em que nomes de modelo com prefixo de provider (ex.: `google/gemini-3.1-pro-preview`) eram passados diretamente para o Ollama durante o fallback de outros serviços. Agora é substituído automaticamente pela variável de ambiente/modelo padrão.
- **Redução significativa dos tempos de cooldown**: Rate limit 429: 30min→5min, pagamento 402: 1h→30min, 401/403: 24h→6h. Previne situação de paralisia total do proxy quando todas as chaves entram em cooldown simultâneo.
- **Retry forçado em cooldown total**: Quando todas as chaves estão em cooldown, a chave que se libera mais cedo é forçada a tentar novamente, evitando rejeição de solicitações.
- **Correção da lista de serviços**: A resposta `/status` exibe a lista real de serviços sincronizados do vault (evita omissão de anthropic, etc.).

### v0.1.25 (2026-04-08)
- **Detecção de processo do agente**: O proxy detecta a sobrevivência de agentes locais (NanoClaw/OpenClaw) e exibe semáforo laranja no dashboard.
- **Melhoria do handle de arraste**: Na ordenação de cards, só é possível arrastar pela área do semáforo (●). Previne arraste acidental em campos de entrada ou botões.

### v0.1.24 (2026-04-06)
- **Subcomando RTK de economia de tokens**: `wall-vault rtk <command>` filtra automaticamente a saída de comandos shell, reduzindo o consumo de tokens do agente de IA em 60-90%. Inclui filtros dedicados para comandos principais como git e go, e trunca automaticamente comandos não suportados. Integra-se de forma transparente via hook `PreToolUse` do Claude Code.

### v0.1.23 (2026-04-06)
- **Correção de alteração de modelo Ollama**: Corrigido problema em que alterar o modelo Ollama no dashboard do cofre não era refletido no proxy. Anteriormente usava apenas variável de ambiente (`OLLAMA_MODEL`), agora prioriza a configuração do cofre.
- **Semáforo automático de serviço local**: Ollama, LM Studio e vLLM são ativados automaticamente quando conectáveis e desativados quando desconectados. Funciona da mesma forma que o toggle automático baseado em chaves dos serviços cloud.

### v0.1.22 (2026-04-05)
- **Correção de campo content vazio ausente**: Quando modelos thinking (gemini-3.1-pro, o1, claude thinking, etc.) gastam todo o limite max_tokens em reasoning e não conseguem gerar resposta real, o proxy omitia os campos `content`/`text` da resposta JSON com `omitempty`, causando crash dos clientes SDK OpenAI/Anthropic com `Cannot read properties of undefined (reading 'trim')`. Alterado para sempre incluir os campos conforme a especificação oficial da API.

### v0.1.21 (2026-04-05)
- **Suporte a modelos Gemma 4**: Modelos da família Gemma como `gemma-4-31b-it`, `gemma-4-26b-a4b-it` podem ser usados via API do Google Gemini.
- **Suporte oficial a serviços LM Studio / vLLM**: Anteriormente, esses serviços eram omitidos no roteamento do proxy e sempre substituídos pelo Ollama. Agora são roteados corretamente via API compatível com OpenAI.
- **Correção da exibição de serviço no dashboard**: Mesmo quando ocorre fallback, o dashboard sempre mostra o serviço configurado pelo usuário.
- **Exibição de status de serviço local**: Ao carregar o dashboard, o status de conexão dos serviços locais (Ollama, LM Studio, vLLM, etc.) é exibido pela cor do ponto ●.
- **Variável de ambiente para filtro de ferramentas**: `WV_TOOL_FILTER=passthrough` permite configurar o modo de passagem de ferramentas (tools) via variável de ambiente.

### v0.1.20 (2026-03-28)
- **Reforço abrangente de segurança**: Prevenção de XSS (41 pontos), comparação de tokens em tempo constante, restrição de CORS, limites de tamanho de solicitação, prevenção de path traversal, autenticação SSE, reforço de rate limiting, entre 12 itens de segurança melhorados.

### v0.1.19 (2026-03-27)
- **Detecção de Claude Code online**: Claude Code que não passa pelo proxy também aparece como online no dashboard.

### v0.1.18 (2026-03-26)
- **Correção de fixação no serviço de fallback**: Após fallback temporário para Ollama, o sistema volta automaticamente ao serviço original quando recuperado.
- **Melhoria na detecção de offline**: Verificação de status a cada 15 segundos acelera a detecção de parada do proxy.

### v0.1.17 (2026-03-25)
- **Ordenação de cards por arrastar e soltar**: Cards de agentes podem ser reorganizados arrastando.
- **Botão inline de aplicação de configuração**: Botão [⚡ Aplicar configuração] exibido em agentes offline.
- **Tipo de agente cokacdir adicionado**.

### v0.1.16 (2026-03-25)
- **Sincronização bidirecional de modelos**: Alterar o modelo do Cline ou Claude Code no dashboard do cofre é aplicado automaticamente.

---

*Para informações mais detalhadas da API, consulte [API.md](API.md).*
