# Manual do Usuário wall-vault
*(Last updated: 2026-04-08 — v0.1.25)*

---

## Índice

1. [O que é wall-vault?](#o-que-é-wall-vault)
2. [Instalação](#instalação)
3. [Primeiros passos (assistente de setup)](#primeiros-passos)
4. [Cadastro de chaves API](#cadastro-de-chaves-api)
5. [Como usar o proxy](#como-usar-o-proxy)
6. [Dashboard do cofre de chaves](#dashboard-do-cofre-de-chaves)
7. [Modo distribuído (multi-bot)](#modo-distribuído-multi-bot)
8. [Inicialização automática](#inicialização-automática)
9. [Doctor (diagnóstico)](#doctor-diagnóstico)
10. [RTK Economia de tokens](#rtk-economia-de-tokens)
11. [Referência de variáveis de ambiente](#referência-de-variáveis-de-ambiente)
12. [Solução de problemas](#solução-de-problemas)

---

## O que é wall-vault?

**wall-vault = proxy de IA + cofre de chaves API para o OpenClaw**

Para usar serviços de IA, você precisa de **chaves API**. Uma chave API é como um **crachá de acesso digital** que prova "esta pessoa tem permissão para usar este serviço". No entanto, esses crachás têm um número limitado de usos por dia e podem ser expostos se mal gerenciados.

wall-vault armazena esses crachás em um cofre seguro e atua como **proxy** entre o OpenClaw e os serviços de IA. Em termos simples, o OpenClaw só precisa se conectar ao wall-vault, e o wall-vault cuida de todo o resto.

Problemas que o wall-vault resolve:

- **Rotação automática de chaves API**: Quando uma chave atinge o limite de uso ou é temporariamente bloqueada (cooldown), muda silenciosamente para a próxima chave. O OpenClaw continua funcionando sem interrupção.
- **Troca automática de serviço (fallback)**: Se o Google não responder, muda automaticamente para o OpenRouter; se isso também falhar, muda para o Ollama/LM Studio/vLLM (IA local) instalado no seu computador. A sessão não é interrompida. Quando o serviço original se recupera, retorna automaticamente a partir da próxima requisição (v0.1.18+, LM Studio/vLLM: v0.1.21+).
- **Sincronização em tempo real (SSE)**: Quando você altera o modelo no dashboard do cofre, a mudança é refletida na tela do OpenClaw em 1-3 segundos. SSE (Server-Sent Events) é uma tecnologia onde o servidor envia atualizações em tempo real para o cliente.
- **Notificações em tempo real**: Eventos como esgotamento de chaves ou falhas de serviço são exibidos imediatamente na parte inferior do TUI (tela do terminal) do OpenClaw.

> 💡 **Claude Code, Cursor, VS Code** também podem ser conectados, mas o objetivo principal do wall-vault é ser usado em conjunto com o OpenClaw.

```
OpenClaw (tela TUI do terminal)
        │
        ▼
  proxy wall-vault (:56244)   ← gerenciamento de chaves, roteamento, fallback, eventos
        │
        ├─ Google Gemini API
        ├─ OpenRouter API (mais de 340 modelos)
        ├─ Ollama / LM Studio / vLLM (computador local, último recurso)
        └─ OpenAI / Anthropic API
```

---

## Instalação

### Linux / macOS

Abra o terminal e cole os comandos abaixo.

```bash
# Linux (PC comum, servidor — amd64)
curl -L https://github.com/sookmook/wall-vault/releases/latest/download/wall-vault-linux-amd64 \
  -o ~/.local/bin/wall-vault && chmod +x ~/.local/bin/wall-vault

# macOS Apple Silicon (M1/M2/M3 Mac)
curl -L https://github.com/sookmook/wall-vault/releases/latest/download/wall-vault-darwin-arm64 \
  -o /usr/local/bin/wall-vault && chmod +x /usr/local/bin/wall-vault
```

- `curl -L ...` — Faz download do arquivo da internet.
- `chmod +x` — Torna o arquivo baixado "executável". Pular esta etapa causará erro de "permissão negada".

### Windows

Abra o PowerShell (como administrador) e execute os comandos abaixo.

```powershell
# Download
Invoke-WebRequest -Uri `
  "https://github.com/sookmook/wall-vault/releases/latest/download/wall-vault-windows-amd64.exe" `
  -OutFile "$env:LOCALAPPDATA\Programs\wall-vault\wall-vault.exe"

# Adicionar ao PATH (aplicado após reiniciar o PowerShell)
$env:PATH += ";$env:LOCALAPPDATA\Programs\wall-vault"
```

> 💡 **O que é PATH?** É a lista de pastas onde o computador procura por comandos. Adicionando ao PATH, você pode executar `wall-vault` de qualquer pasta.

### Compilar a partir do código-fonte (para desenvolvedores)

Aplicável apenas se você tiver o ambiente de desenvolvimento Go instalado.

```bash
git clone https://github.com/sookmook/wall-vault
cd wall-vault
make build       # bin/wall-vault (versão: v0.1.25.YYYYMMDD.HHmmss)
make install     # ~/.local/bin/wall-vault
```

> 💡 **Versão com timestamp de build**: Ao compilar com `make build`, a versão é gerada automaticamente em formato como `v0.1.25.20260408.022325`, incluindo data e hora. Se compilar diretamente com `go build ./...`, a versão aparecerá apenas como `"dev"`.

---

## Primeiros passos

### Executar o assistente de setup

Após a instalação, é obrigatório executar o **assistente de configuração** com o comando abaixo pela primeira vez. O assistente guiará você, perguntando os itens necessários um por um.

```bash
wall-vault setup
```

As etapas que o assistente percorre são:

```
1. Seleção de idioma (10 idiomas incluindo coreano)
2. Seleção de tema (light / dark / gold / cherry / ocean)
3. Modo de operação — usar sozinho (standalone) ou em vários computadores (distributed)
4. Nome do bot — o nome que aparecerá no dashboard
5. Configuração de portas — padrão: proxy 56244, cofre 56243 (pressione Enter se não precisar alterar)
6. Seleção de serviços de IA — Google / OpenRouter / Ollama / LM Studio / vLLM
7. Configuração do filtro de ferramentas de segurança
8. Configuração do token de administrador — senha para bloquear funções de gerenciamento do dashboard. Geração automática também é possível
9. Configuração de senha de criptografia de chaves API — para armazenar chaves com mais segurança (opcional)
10. Caminho para salvar o arquivo de configuração
```

> ⚠️ **Lembre-se do token de administrador.** Será necessário depois para adicionar chaves ou alterar configurações no dashboard. Se perdê-lo, terá que editar o arquivo de configuração manualmente.

Ao concluir o assistente, o arquivo de configuração `wall-vault.yaml` é gerado automaticamente.

### Execução

```bash
wall-vault start
```

Dois servidores iniciam simultaneamente:

- **Proxy** (`http://localhost:56244`) — O intermediário que conecta o OpenClaw aos serviços de IA
- **Cofre de chaves** (`http://localhost:56243`) — Gerenciamento de chaves API e dashboard web

Abra `http://localhost:56243` no navegador para ver o dashboard imediatamente.

---

## Cadastro de chaves API

Existem quatro formas de cadastrar chaves API. **Para iniciantes, recomendamos o Método 1 (variáveis de ambiente)**.

### Método 1: Variáveis de ambiente (recomendado — mais simples)

Variáveis de ambiente são **valores pré-configurados** que o programa lê ao iniciar. Basta digitar no terminal conforme abaixo.

```bash
# Cadastrar chave Google Gemini
export WV_KEY_GOOGLE=AIzaSy...

# Cadastrar chave OpenRouter
export WV_KEY_OPENROUTER=sk-or-v1-...

# Executar após o cadastro
wall-vault start
```

Se tiver várias chaves, separe-as com vírgulas(,). O wall-vault as usará automaticamente em rodízio (round robin):

```bash
export WV_KEY_GOOGLE=AIzaSy...,AIzaSy...,AIzaSy...
```

> 💡 **Dica**: O comando `export` se aplica apenas à sessão atual do terminal. Para persistir após reiniciar o computador, adicione as linhas acima ao arquivo `~/.bashrc` ou `~/.zshrc`.

### Método 2: Interface do Dashboard (clique com o mouse)

1. Acesse `http://localhost:56243` no navegador
2. Clique no botão `[+ Adicionar]` no card **🔑 Chaves API** na parte superior
3. Insira o tipo de serviço, valor da chave, rótulo (nome de referência) e limite diário, depois salve

### Método 3: API REST (para automação/scripts)

API REST é uma forma de programas trocarem dados via HTTP. Útil para cadastro automático via scripts.

```bash
curl -X POST http://localhost:56243/admin/keys \
  -H "Authorization: Bearer TOKEN_ADMIN" \
  -H "Content-Type: application/json" \
  -d '{
    "service": "google",
    "key": "AIzaSy...",
    "label": "Chave principal",
    "daily_limit": 1000
  }'
```

### Método 4: Flag do proxy (para testes rápidos)

Usado para testar temporariamente inserindo uma chave sem cadastro formal. Desaparece ao encerrar o programa.

```bash
wall-vault proxy --key-google=AIzaSy... --key-openrouter=sk-or-...
```

---

## Como usar o proxy

### Uso com OpenClaw (objetivo principal)

Veja como configurar o OpenClaw para se conectar aos serviços de IA através do wall-vault.

Abra o arquivo `~/.openclaw/openclaw.json` e adicione o conteúdo abaixo:

```json5
// ~/.openclaw/openclaw.json
{
  models: {
    providers: {
      "wall-vault": {
        baseUrl: "http://localhost:56244/v1",
        apiKey: "your-agent-token",   // token do agente do vault
        api: "openai-completions",
        models: [
          { id: "wall-vault/gemini-2.5-flash" },
          { id: "wall-vault/gemini-2.5-pro" },
          { id: "wall-vault/hunter-alpha" },    // contexto de 1M grátis
          { id: "wall-vault/claude-opus-4-6" }
        ]
      }
    }
  }
}
```

> 💡 **Forma mais fácil**: Clique no botão **🦞 Copiar config OpenClaw** no card do agente no dashboard. Um snippet com o token e endereço já preenchidos será copiado para a área de transferência. Basta colar.

**Para onde o `wall-vault/` no início do nome do modelo aponta?**

O wall-vault determina automaticamente para qual serviço de IA enviar a requisição com base no nome do modelo:

| Formato do modelo | Serviço conectado |
|----------|--------------|
| `wall-vault/gemini-*` | Conexão direta com Google Gemini |
| `wall-vault/gpt-*`, `wall-vault/o3`, `wall-vault/o4*` | Conexão direta com OpenAI |
| `wall-vault/claude-*` | Conexão com Anthropic via OpenRouter |
| `wall-vault/hunter-alpha`, `wall-vault/healer-alpha` | OpenRouter (1 milhão de tokens de contexto grátis) |
| `wall-vault/kimi-*`, `wall-vault/glm-*`, `wall-vault/deepseek-*` | Conexão com OpenRouter |
| `google/modelo`, `openai/modelo`, `anthropic/modelo` etc. | Conexão direta com o respectivo serviço |
| `custom/google/modelo`, `custom/openai/modelo` etc. | Remove a parte `custom/` e redireciona |
| `modelo:cloud` | Remove a parte `:cloud` e conecta via OpenRouter |

> 💡 **O que é contexto?** É a quantidade de conversa que a IA consegue lembrar de uma vez. 1M (um milhão de tokens) permite processar conversas muito longas ou documentos extensos de uma vez.

### Conexão direta no formato da API Gemini (compatibilidade com ferramentas existentes)

Se você tiver ferramentas que usam a API do Google Gemini diretamente, basta alterar o endereço para o wall-vault:

```bash
export ANTHROPIC_BASE_URL=http://localhost:56244/google
```

Ou se a ferramenta permite especificar a URL diretamente:

```
http://localhost:56244/google/v1beta/models/gemini-2.5-flash:generateContent
```

### Uso com OpenAI SDK (Python)

Você também pode conectar o wall-vault em código Python que utiliza IA. Basta alterar o `base_url`:

```python
from openai import OpenAI

client = OpenAI(
    base_url="http://localhost:56244/v1",
    api_key="not-needed"  # As chaves API são gerenciadas pelo wall-vault
)

response = client.chat.completions.create(
    model="google/gemini-2.5-flash",   # formato provider/model
    messages=[{"role": "user", "content": "Olá"}]
)
```

### Alterar modelo durante a execução

Para alterar o modelo de IA enquanto o wall-vault está em execução:

```bash
# Alterar modelo requisitando diretamente ao proxy
curl -X PUT http://localhost:56244/api/config/model \
  -H "Content-Type: application/json" \
  -d '{"service": "openrouter", "model": "anthropic/claude-3.5-sonnet"}'

# No modo distribuído (multi-bot), altere no servidor do cofre → refletido instantaneamente via SSE
curl -X PUT http://localhost:56243/admin/clients/meu-bot-id \
  -H "Authorization: Bearer TOKEN_ADMIN" \
  -H "Content-Type: application/json" \
  -d '{"default_service": "google", "default_model": "gemini-2.5-pro"}'
```

### Verificar lista de modelos disponíveis

```bash
# Ver lista completa
curl http://localhost:56244/api/models | python3 -m json.tool

# Ver apenas modelos do Google
curl "http://localhost:56244/api/models?service=google"

# Pesquisar por nome (ex: modelos contendo "claude")
curl "http://localhost:56244/api/models?q=claude"
```

**Resumo dos principais modelos por serviço:**

| Serviço | Principais modelos |
|--------|----------|
| Google | gemini-2.5-pro, gemini-2.5-flash, gemini-2.5-flash-8b, gemini-2.0-flash |
| OpenAI | gpt-4o, gpt-4o-mini, o3, o1, o1-mini |
| OpenRouter | Mais de 346 (Hunter Alpha com 1M de contexto grátis, DeepSeek R1/V3, Qwen 2.5 etc.) |
| Ollama | Detecção automática do servidor local instalado no computador |
| LM Studio | Servidor local no computador (porta 1234) |
| vLLM | Servidor local no computador (porta 8000) |

---

## Dashboard do cofre de chaves

Acesse `http://localhost:56243` no navegador para ver o dashboard.

**Layout da tela:**
- **Barra superior fixa (topbar)**: Logo, seletores de idioma/tema, indicador de status da conexão SSE
- **Grade de cards**: Cards de agentes, serviços e chaves API dispostos em formato de mosaico

### Card de chaves API

Card para gerenciar as chaves API cadastradas de forma visual.

- Exibe a lista de chaves separadas por serviço.
- `today_usage`: Número de tokens (caracteres lidos e escritos pela IA) processados com sucesso hoje
- `today_attempts`: Número total de chamadas hoje (incluindo sucesso + falhas)
- Cadastre novas chaves com o botão `[+ Adicionar]` e exclua com `✕`.

> 💡 **O que é um token?** É a unidade que a IA usa para processar texto. Equivale aproximadamente a uma palavra em inglês, ou 1-2 caracteres em coreano. As tarifas da API geralmente são calculadas com base no número de tokens.

### Card de agente

Card que mostra o status dos bots (agentes) conectados ao proxy wall-vault.

**O status da conexão é exibido em 4 níveis:**

| Indicador | Status | Significado |
|------|------|------|
| 🟢 | Em execução | Proxy funcionando normalmente |
| 🟡 | Atrasado | Respondendo, mas lentamente |
| 🔴 | Offline | Proxy não está respondendo |
| ⚫ | Desconectado/Inativo | Proxy nunca se conectou ao cofre ou está desativado |

**Guia dos botões na parte inferior do card de agente:**

Ao cadastrar um agente, especifique o **tipo de agente** e os botões de conveniência correspondentes aparecerão automaticamente.

---

#### 🔘 Botão Copiar Configuração — Gera automaticamente a configuração de conexão

Ao clicar no botão, um snippet de configuração com o token do agente, endereço do proxy e informações do modelo já preenchidos é copiado para a área de transferência. Basta colar o conteúdo copiado no local indicado na tabela abaixo para completar a configuração da conexão.

| Botão | Tipo de agente | Onde colar |
|------|-------------|-------------|
| 🦞 Copiar config OpenClaw | `openclaw` | `~/.openclaw/openclaw.json` |
| 🦀 Copiar config NanoClaw | `nanoclaw` | `~/.openclaw/openclaw.json` |
| 🟠 Copiar config Claude Code | `claude-code` | `~/.claude/settings.json` |
| ⌨ Copiar config Cursor | `cursor` | Cursor → Settings → AI |
| 💻 Copiar config VSCode | `vscode` | `~/.continue/config.json` |

**Exemplo — Para tipo Claude Code, o conteúdo copiado será:**

```json
// ~/.claude/settings.json
{
  "apiProvider": "openai",
  "baseUrl": "http://192.168.0.6:56244/v1",
  "apiKey": "token-deste-agente"
}
```

**Exemplo — Para tipo VSCode (Continue):**

```yaml
# ~/.continue/config.yaml  ← colar em config.yaml, não config.json
name: My Config
version: 0.0.1
schema: v1

models:
  - name: wall-vault proxy
    provider: openai
    model: gemini-2.5-flash
    apiBase: http://192.168.0.6:56244/v1
    apiKey: token-deste-agente
    roles:
      - chat
      - edit
      - apply
```

> ⚠️ **A versão mais recente do Continue usa `config.yaml`.** Se `config.yaml` existir, `config.json` é completamente ignorado. Certifique-se de colar em `config.yaml`.

**Exemplo — Para tipo Cursor:**

```
Base URL : http://192.168.0.6:56244/v1
API Key  : token-deste-agente

// Ou variáveis de ambiente:
OPENAI_BASE_URL=http://192.168.0.6:56244/v1
OPENAI_API_KEY=token-deste-agente
```

> ⚠️ **Quando a cópia para a área de transferência não funcionar**: A política de segurança do navegador pode bloquear a cópia. Se uma caixa de texto pop-up aparecer, selecione tudo com Ctrl+A e copie com Ctrl+C.

---

#### ⚡ Botão de aplicação automática — Um clique e a configuração está pronta

Para agentes do tipo `cline`, `claude-code`, `openclaw` ou `nanoclaw`, um botão **⚡ Aplicar Configuração** aparece no card do agente. Ao clicar neste botão, o arquivo de configuração local do agente é automaticamente atualizado.

| Botão | Tipo de agente | Arquivo de destino |
|------|-------------|-------------|
| ⚡ Aplicar config Cline | `cline` | `~/.cline/data/globalState.json` + `secrets.json` |
| ⚡ Aplicar config Claude Code | `claude-code` | `~/.claude/settings.json` |
| ⚡ Aplicar config OpenClaw | `openclaw` | `~/.openclaw/openclaw.json` |
| ⚡ Aplicar config NanoClaw | `nanoclaw` | `~/.openclaw/openclaw.json` |

> ⚠️ Este botão envia uma requisição para **localhost:56244** (proxy local). O proxy deve estar em execução naquela máquina para funcionar.

---

#### 🔀 Arrastar e soltar cards para reordenar (v0.1.17, melhorado v0.1.25)

Você pode **arrastar** os cards de agente no dashboard para reorganizá-los na ordem desejada.

1. Clique e segure a área do **semáforo (●)** no canto superior esquerdo do card e arraste
2. Solte sobre o card na posição desejada e a ordem será alterada

> 💡 O corpo do card (campos de entrada, botões, etc.) não pode ser arrastado. Só é possível arrastar pela área do semáforo.

#### 🟠 Detecção de processo do agente (v0.1.25)

Quando o proxy está funcionando normalmente, mas o processo do agente local (NanoClaw, OpenClaw) parou, o semáforo do card muda para **laranja (piscando)** e a mensagem "Processo do agente parado" é exibida.

- 🟢 Verde: Proxy + agente normais
- 🟠 Laranja (piscando): Proxy normal, agente parado
- 🔴 Vermelho: Proxy offline
3. A ordem alterada é **salva imediatamente no servidor** e mantida mesmo ao atualizar a página

> 💡 Dispositivos touch (celular/tablet) ainda não são suportados. Use em navegadores desktop.

---

#### 🔄 Sincronização bidirecional de modelos (v0.1.16)

Quando você altera o modelo de um agente no dashboard do cofre, a configuração local do agente é automaticamente atualizada.

**Para Cline:**
- Alteração de modelo no cofre → evento SSE → proxy atualiza o campo de modelo em `globalState.json`
- Campos atualizados: `actModeOpenAiModelId`, `planModeOpenAiModelId`, `openAiModelId`
- `openAiBaseUrl` e a chave API não são alterados
- **É necessário recarregar o VS Code (`Ctrl+Alt+R` ou `Ctrl+Shift+P` → `Developer: Reload Window`)**
  - Cline não relê o arquivo de configuração durante a execução

**Para Claude Code:**
- Alteração de modelo no cofre → evento SSE → proxy atualiza o campo `model` em `settings.json`
- Busca automática em caminhos WSL e Windows (`~/.claude/`, `/mnt/c/Users/*/.claude/`)

**Direção inversa (agente → cofre):**
- Quando um agente (Cline, Claude Code, etc.) envia uma requisição ao proxy, o proxy inclui as informações de serviço/modelo daquele cliente no heartbeat
- O serviço/modelo atualmente em uso é exibido em tempo real no card do agente no dashboard do cofre

> 💡 **Ponto-chave**: O proxy identifica o agente pelo token de Authorization da requisição e faz o roteamento automático para o serviço/modelo configurado no cofre. Mesmo que o Cline ou Claude Code envie um nome de modelo diferente, o proxy sobrescreve com a configuração do cofre.

---

### Usando Cline no VS Code — Guia detalhado

#### Etapa 1: Instalar o Cline

Instale o **Cline** (ID: `saoudrizwan.claude-dev`) no marketplace de extensões do VS Code.

#### Etapa 2: Cadastrar agente no cofre

1. Abra o dashboard do cofre (`http://IP-do-cofre:56243`)
2. Na seção **Agentes**, clique em **+ Adicionar**
3. Preencha da seguinte forma:

| Campo | Valor | Descrição |
|------|----|------|
| ID | `meu_cline` | Identificador único (alfanumérico, sem espaços) |
| Nome | `Meu Cline` | Nome exibido no dashboard |
| Tipo de agente | `cline` | ← Selecione obrigatoriamente `cline` |
| Serviço | Selecione o serviço desejado (ex: `google`) | |
| Modelo | Insira o modelo desejado (ex: `gemini-2.5-flash`) | |

4. Ao clicar em **Salvar**, o token é gerado automaticamente

#### Etapa 3: Conectar ao Cline

**Método A — Aplicação automática (recomendado)**

1. Verifique se o **proxy** wall-vault está em execução naquela máquina (`localhost:56244`)
2. Clique no botão **⚡ Aplicar config Cline** no card do agente no dashboard
3. Se aparecer a notificação "Configuração aplicada com sucesso!", foi bem-sucedido
4. Recarregue o VS Code (`Ctrl+Alt+R`)

**Método B — Configuração manual**

Abra as configurações (⚙️) na barra lateral do Cline:
- **API Provider**: `OpenAI Compatible`
- **Base URL**: `http://endereço-do-proxy:56244/v1`
  - Se for a mesma máquina: `http://localhost:56244/v1`
  - Se for outra máquina (ex: servidor Mini): `http://192.168.0.6:56244/v1`
- **API Key**: Token emitido pelo cofre (copie do card do agente)
- **Model ID**: Modelo configurado no cofre (ex: `gemini-2.5-flash`)

#### Etapa 4: Verificação

Envie qualquer mensagem no chat do Cline. Se estiver normal:
- Um **ponto verde (● Em execução)** aparecerá no card do agente correspondente no dashboard do cofre
- O serviço/modelo atual será exibido no card (ex: `google / gemini-2.5-flash`)

#### Alterar modelo

Quando quiser alterar o modelo do Cline, altere no **dashboard do cofre**:

1. Altere o dropdown de serviço/modelo no card do agente
2. Clique em **Aplicar**
3. Recarregue o VS Code (`Ctrl+Alt+R`) — o nome do modelo no rodapé do Cline será atualizado
4. O novo modelo será usado a partir da próxima requisição

> 💡 Na prática, o proxy identifica a requisição do Cline pelo token e roteia para o modelo configurado no cofre. Mesmo sem recarregar o VS Code, **o modelo realmente usado muda imediatamente** — o recarregamento é apenas para atualizar a exibição do modelo na interface do Cline.

#### Detecção de desconexão

Quando o VS Code é fechado, o card do agente no dashboard do cofre muda para amarelo (atrasado) após cerca de **90 segundos** e para vermelho (offline) após **3 minutos**. (A partir do v0.1.18, a detecção de offline ficou mais rápida com verificação de status a cada 15 segundos.)

#### Solução de problemas

| Sintoma | Causa | Solução |
|------|------|------|
| Erro "Falha na conexão" no Cline | Proxy não está em execução ou endereço incorreto | Verifique o proxy com `curl http://localhost:56244/health` |
| Ponto verde não aparece no cofre | Token (chave API) não configurado | Clique novamente no botão **⚡ Aplicar config Cline** |
| Nome do modelo no rodapé do Cline não muda | Cline mantém cache da configuração | Recarregue o VS Code (`Ctrl+Alt+R`) |
| Nome de modelo errado exibido | Bug antigo (corrigido no v0.1.16) | Atualize o proxy para v0.1.16 ou superior |

---

#### 🟣 Botão Copiar comando de deploy — Para instalação em novas máquinas

Usado ao instalar o proxy wall-vault pela primeira vez em um novo computador e conectá-lo ao cofre. Ao clicar no botão, o script de instalação completo é copiado. Cole e execute no terminal do novo computador e o seguinte será processado de uma vez:

1. Instalação do binário wall-vault (pula se já estiver instalado)
2. Registro automático do serviço systemd do usuário
3. Iniciar serviço e conectar automaticamente ao cofre

> 💡 O script já contém o token deste agente e o endereço do servidor do cofre, então pode ser executado imediatamente após colar, sem modificações.

---

### Card de serviços

Card para ativar, desativar ou configurar os serviços de IA que serão usados.

- Switch toggle para ativar/desativar cada serviço
- Insira o endereço do servidor de IA local (Ollama, LM Studio, vLLM, etc. rodando no seu computador) para que os modelos disponíveis sejam detectados automaticamente.
- **Indicador de status de conexão do serviço local**: O ponto ● ao lado do nome do serviço fica **verde** se conectado e **cinza** se desconectado
- **Semáforo automático para serviços locais** (v0.1.23+): Serviços locais (Ollama, LM Studio, vLLM) são automaticamente ativados/desativados conforme a disponibilidade da conexão. Ao ativar um serviço, dentro de 15 segundos o ponto ● fica verde e o checkbox é marcado; ao desativar, é automaticamente desmarcado. Funciona da mesma forma que o toggle automático de serviços cloud (Google, OpenRouter, etc.) baseado na presença de chaves API.

> 💡 **Se o serviço local está rodando em outro computador**: Insira o IP daquele computador no campo de URL do serviço. Ex: `http://192.168.0.6:11434` (Ollama), `http://192.168.0.6:1234` (LM Studio). Se o serviço estiver vinculado apenas a `127.0.0.1` ao invés de `0.0.0.0`, o acesso via IP externo não funcionará, então verifique o endereço de binding nas configurações do serviço.

### Inserção do token de administrador

Ao tentar usar funções importantes no dashboard como adicionar/excluir chaves, um pop-up para inserir o token de administrador aparecerá. Insira o token que configurou no assistente de setup. Uma vez inserido, permanece válido até fechar o navegador.

> ⚠️ **Se houver mais de 10 tentativas de autenticação falhadas em 15 minutos, o IP será temporariamente bloqueado.** Se esqueceu o token, verifique o item `admin_token` no arquivo `wall-vault.yaml`.

---

## Modo distribuído (multi-bot)

É uma configuração onde **um único cofre de chaves é compartilhado** ao operar o OpenClaw em vários computadores simultaneamente. É conveniente pois o gerenciamento de chaves é feito em um só lugar.

### Exemplo de configuração

```
[Servidor do cofre de chaves]
  wall-vault vault    (cofre de chaves :56243, dashboard)

[WSL Alpha]            [Raspberry Pi Gamma]    [Mac Mini Local]
  wall-vault proxy      wall-vault proxy        wall-vault proxy
  openclaw TUI          openclaw TUI            openclaw TUI
  ↕ sinc. SSE           ↕ sinc. SSE             ↕ sinc. SSE
```

Todos os bots apontam para o servidor central do cofre, então quando você altera o modelo ou adiciona chaves no cofre, a mudança é refletida instantaneamente em todos os bots.

### Etapa 1: Iniciar o servidor do cofre de chaves

Execute no computador que será usado como servidor do cofre:

```bash
wall-vault vault
```

### Etapa 2: Cadastrar cada bot (cliente)

Cadastre previamente as informações de cada bot que se conectará ao servidor do cofre:

```bash
curl -X POST http://localhost:56243/admin/clients \
  -H "Authorization: Bearer TOKEN_ADMIN" \
  -H "Content-Type: application/json" \
  -d '{
    "id": "botA",
    "name": "BotA",
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

> 💡 **Substitua `192.168.x.x`** pelo endereço IP interno real do computador do servidor do cofre. Pode ser verificado nas configurações do roteador ou com o comando `ip addr`.

---

## Inicialização automática

Se for incômodo iniciar o wall-vault manualmente toda vez que reiniciar o computador, registre-o como serviço do sistema. Uma vez registrado, iniciará automaticamente na inicialização.

### Linux — systemd (maioria das distribuições Linux)

systemd é o sistema de gerenciamento automático de programas no Linux:

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

1. Baixe o NSSM em [nssm.cc](https://nssm.cc/download) e adicione ao PATH.
2. No PowerShell como administrador:

```powershell
wall-vault doctor deploy windows
```

---

## Doctor (diagnóstico)

O comando `doctor` é uma ferramenta que **diagnostica e corrige automaticamente** se o wall-vault está configurado corretamente.

```bash
wall-vault doctor check   # Diagnóstico do estado atual (apenas leitura, nada é alterado)
wall-vault doctor fix     # Corrigir problemas automaticamente
wall-vault doctor all     # Diagnóstico + correção automática de uma vez
```

> 💡 Se algo parecer errado, execute `wall-vault doctor all` primeiro. Muitos problemas são resolvidos automaticamente.

---

## RTK Economia de tokens

*(v0.1.24+)*

**RTK (ferramenta de economia de tokens)** comprime automaticamente a saída dos comandos shell executados por agentes de IA (como Claude Code), reduzindo o consumo de tokens. Por exemplo, a saída de 15 linhas do `git status` é resumida em 2 linhas.

### Uso básico

```bash
# Encapsule comandos com wall-vault rtk para filtragem automática da saída
wall-vault rtk git status          # Mostra apenas a lista de arquivos alterados
wall-vault rtk git diff HEAD~1     # Apenas linhas alteradas + contexto mínimo
wall-vault rtk git log -10         # Uma linha por commit: hash + mensagem
wall-vault rtk go test ./...       # Mostra apenas testes que falharam
wall-vault rtk ls -la              # Comandos não suportados são truncados automaticamente
```

### Comandos suportados e economia

| Comando | Método de filtragem | Economia |
|------|----------|--------|
| `git status` | Apenas resumo de arquivos alterados | ~87% |
| `git diff` | Linhas alteradas + 3 linhas de contexto | ~60-94% |
| `git log` | Hash + primeira linha da mensagem | ~90% |
| `git push/pull/fetch` | Remove progresso, apenas resumo | ~80% |
| `go test` | Mostra apenas falhas, conta aprovações | ~88-99% |
| `go build/vet` | Mostra apenas erros | ~90% |
| Todos os outros comandos | Primeiras 50 + últimas 50 linhas, máximo 32KB | Variável |

### Pipeline de 3 etapas de filtragem

1. **Filtro estrutural por comando** — Entende o formato de saída de git, go, etc. e extrai apenas as partes significativas
2. **Pós-processamento por regex** — Remove códigos de cor ANSI, reduz linhas vazias, agrega linhas duplicadas
3. **Passthrough + truncamento** — Comandos não suportados mantêm apenas as primeiras/últimas 50 linhas

### Integração com Claude Code

Pode ser configurado para que todos os comandos shell passem automaticamente pelo RTK usando o hook `PreToolUse` do Claude Code.

```bash
# Instalar hook (adiciona automaticamente ao settings.json do Claude Code)
wall-vault rtk hook install
```

Ou adicione manualmente ao `~/.claude/settings.json`:

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

> 💡 **Preservação do exit code**: O RTK retorna o código de saída original do comando. Se o comando falhar (exit code ≠ 0), a IA detecta a falha corretamente.

> 💡 **Saída em inglês forçada**: O RTK executa comandos com `LC_ALL=C` para gerar sempre saída em inglês, independentemente da configuração de idioma do sistema. Isso garante que os filtros funcionem corretamente.

---

## Referência de variáveis de ambiente

Variáveis de ambiente são uma forma de passar valores de configuração para o programa. Insira no terminal no formato `export VARIAVEL=valor` ou coloque no arquivo de serviço de inicialização automática para aplicação permanente.

| Variável | Descrição | Valor de exemplo |
|------|------|---------|
| `WV_LANG` | Idioma do dashboard | `ko`, `en`, `ja` |
| `WV_THEME` | Tema do dashboard | `light`, `dark`, `gold` |
| `WV_KEY_GOOGLE` | Chave(s) API do Google (separadas por vírgula) | `AIza...,AIza...` |
| `WV_KEY_OPENROUTER` | Chave API do OpenRouter | `sk-or-v1-...` |
| `WV_VAULT_URL` | Endereço do servidor do cofre no modo distribuído | `http://192.168.x.x:56243` |
| `WV_VAULT_TOKEN` | Token de autenticação do cliente (bot) | `my-secret-token` |
| `WV_ADMIN_TOKEN` | Token de administrador | `admin-token-here` |
| `WV_MASTER_PASS` | Senha de criptografia de chaves API | `my-password` |
| `WV_AVATAR` | Caminho do arquivo de avatar (relativo a `~/.openclaw/`) | `workspace/avatars/avatar.png` |
| `OLLAMA_URL` | Endereço do servidor local Ollama | `http://192.168.x.x:11434` |

---

## Solução de problemas

### Quando o proxy não inicia

Frequentemente, a porta já está sendo usada por outro programa.

```bash
ss -tlnp | grep 56244   # Verificar quem está usando a porta 56244
wall-vault proxy --port 8080   # Iniciar com outro número de porta
```

### Erros de chave API (429, 402, 401, 403, 582)

| Código de erro | Significado | Ação |
|----------|------|----------|
| **429** | Muitas requisições (limite de uso excedido) | Aguarde ou adicione outra chave |
| **402** | Pagamento necessário ou créditos insuficientes | Recarregue créditos no serviço correspondente |
| **401 / 403** | Chave incorreta ou sem permissão | Verifique o valor da chave e recadastre |
| **582** | Sobrecarga do gateway (cooldown de 5 min) | Será liberado automaticamente após 5 minutos |

```bash
# Verificar lista de chaves cadastradas e status
curl -H "Authorization: Bearer TOKEN_ADMIN" http://localhost:56243/admin/keys

# Resetar contadores de uso das chaves
curl -X POST -H "Authorization: Bearer TOKEN_ADMIN" http://localhost:56243/admin/keys/reset
```

### Quando o agente aparece como "Desconectado"

"Desconectado" significa que o processo do proxy não está enviando sinais (heartbeat) para o cofre. **Não significa que as configurações não foram salvas.** O proxy precisa estar em execução com o endereço do servidor do cofre e o token para que o status de conexão mude.

```bash
# Iniciar o proxy especificando endereço do servidor do cofre, token e ID do cliente
WV_VAULT_URL=http://endereco-do-cofre:56243 \
WV_VAULT_TOKEN=token-do-cliente \
WV_VAULT_CLIENT_ID=id-do-cliente \
wall-vault proxy
```

Se a conexão for bem-sucedida, o status mudará para 🟢 Em execução no dashboard dentro de cerca de 20 segundos.

### Quando a conexão com Ollama não funciona

Ollama é um programa que executa IA diretamente no seu computador. Primeiro, verifique se o Ollama está ligado.

```bash
curl http://localhost:11434/api/tags   # Se a lista de modelos aparecer, está normal
export OLLAMA_URL=http://192.168.x.x:11434   # Se estiver rodando em outro computador
```

> ⚠️ Se o Ollama não responder, inicie-o primeiro com o comando `ollama serve`.

> ⚠️ **Modelos grandes são lentos**: Modelos grandes como `qwen3.5:35b` e `deepseek-r1` podem levar vários minutos para gerar respostas. Mesmo que pareça que não há resposta, pode estar processando normalmente, então aguarde.

---

## Alterações recentes (v0.1.16 ~ v0.1.25)

### v0.1.25 (2026-04-08)
- **Detecção de processo do agente**: O proxy detecta se o agente local (NanoClaw/OpenClaw) está vivo e exibe um semáforo laranja no dashboard.
- **Melhoria do handle de arrastar**: Ao reordenar cards, só é possível arrastar pela área do semáforo (●). Não é mais possível arrastar acidentalmente de campos de entrada ou botões.

### v0.1.24 (2026-04-06)
- **Subcomando RTK de economia de tokens**: `wall-vault rtk <command>` filtra automaticamente a saída de comandos shell, reduzindo o consumo de tokens de agentes de IA em 60-90%. Inclui filtros dedicados para comandos principais como git e go, e trunca automaticamente comandos não suportados. Integra-se transparentemente com o hook `PreToolUse` do Claude Code.

### v0.1.23 (2026-04-06)
- **Correção da mudança de modelo Ollama**: Corrigido problema onde mudar o modelo Ollama no dashboard do cofre não era refletido no proxy real. Anteriormente usava apenas a variável de ambiente (`OLLAMA_MODEL`), agora prioriza a configuração do cofre.
- **Semáforo automático para serviços locais**: Ollama/LM Studio/vLLM são automaticamente ativados quando conectáveis e desativados quando desconectados. Funciona da mesma forma que o toggle automático de serviços cloud baseado em chaves.

### v0.1.22 (2026-04-05)
- **Correção de campo content vazio ausente**: Quando modelos thinking (gemini-3.1-pro, o1, claude thinking, etc.) usam todo o limite de max_tokens em reasoning e não conseguem gerar resposta real, o proxy omitia os campos `content`/`text` da resposta JSON via `omitempty`, causando crashes em clientes SDK OpenAI/Anthropic com erro `Cannot read properties of undefined (reading 'trim')`. Corrigido para sempre incluir os campos conforme a especificação oficial da API.

### v0.1.21 (2026-04-05)
- **Suporte a modelos Gemma 4**: Modelos da família Gemma como `gemma-4-31b-it` e `gemma-4-26b-a4b-it` podem ser usados via Google Gemini API.
- **Suporte oficial para serviços LM Studio / vLLM**: Anteriormente, esses serviços eram ignorados no roteamento do proxy e sempre substituídos pelo Ollama. Agora são roteados corretamente via API compatível com OpenAI.
- **Correção da exibição de serviço no dashboard**: Mesmo quando ocorre fallback, o dashboard sempre exibe o serviço configurado pelo usuário.
- **Indicador de status do serviço local**: Ao carregar o dashboard, exibe o status de conexão de serviços locais (Ollama, LM Studio, vLLM, etc.) através da cor do ponto ●.
- **Variável de ambiente do filtro de ferramentas**: O modo de passagem de ferramentas (tools) pode ser configurado via variável de ambiente `WV_TOOL_FILTER=passthrough`.

### v0.1.20 (2026-03-28)
- **Reforço de segurança abrangente**: 12 itens de segurança melhorados incluindo prevenção de XSS (41 pontos), comparação de tokens em tempo constante, restrição CORS, limites de tamanho de requisição, prevenção de travessia de diretório, autenticação SSE e reforço de rate limiter.

### v0.1.19 (2026-03-27)
- **Detecção online do Claude Code**: Claude Code que não passa pelo proxy também é exibido como online no dashboard.

### v0.1.18 (2026-03-26)
- **Correção de serviço fallback fixado**: Após fallback para Ollama devido a erro temporário, retorna automaticamente quando o serviço original se recupera.
- **Melhoria na detecção offline**: Detecção mais rápida de parada do proxy com verificação de status a cada 15 segundos.

### v0.1.17 (2026-03-25)
- **Reordenação de cards por arrastar e soltar**: Os cards de agente podem ser reorganizados arrastando-os.
- **Botão inline de aplicação de configuração**: O botão [⚡ Aplicar config] é exibido em agentes offline.
- **Adição do tipo de agente cokacdir**.

### v0.1.16 (2026-03-25)
- **Sincronização bidirecional de modelos**: Alterar o modelo de Cline/Claude Code no dashboard do cofre é refletido automaticamente.

---

*Para informações mais detalhadas sobre a API, consulte [API.md](API.md).*
