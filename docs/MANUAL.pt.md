# Manual do Usuário wall-vault
*(Última atualização: 2026-04-05 — v0.1.21)*

---

## Índice

1. [O que é o wall-vault?](#o-que-é-o-wall-vault)
2. [Instalação](#instalação)
3. [Primeiros passos (assistente de configuração)](#primeiros-passos)
4. [Cadastro de chaves de API](#cadastro-de-chaves-de-api)
5. [Como usar o proxy](#como-usar-o-proxy)
6. [Painel do cofre de chaves](#painel-do-cofre-de-chaves)
7. [Modo distribuído (múltiplos bots)](#modo-distribuído-múltiplos-bots)
8. [Configuração de inicialização automática](#configuração-de-inicialização-automática)
9. [Doctor — diagnóstico automático](#doctor--diagnóstico-automático)
10. [Variáveis de ambiente](#variáveis-de-ambiente)
11. [Solução de problemas](#solução-de-problemas)

---

## O que é o wall-vault?

**wall-vault = proxy de IA + cofre de chaves de API para o OpenClaw**

Para usar serviços de inteligência artificial, você precisa de uma **chave de API** — pense nela como um **crachá digital** que prova que você tem permissão para usar aquele serviço. Esse crachá tem um limite de uso diário e pode ser comprometido se não for guardado com cuidado.

O wall-vault guarda esses crachás em um cofre seguro e atua como **intermediário (proxy)** entre o OpenClaw e os serviços de IA. Na prática, o OpenClaw só precisa falar com o wall-vault; todo o resto — autenticação, roteamento, fallback — é resolvido automaticamente.

Problemas que o wall-vault resolve para você:

- **Rotação automática de chaves**: quando uma chave atinge o limite de uso ou fica temporariamente bloqueada (cooldown), o wall-vault passa silenciosamente para a próxima chave. O OpenClaw continua funcionando sem interrupção.
- **Troca automática de serviço (fallback)**: se o Google não responder, o wall-vault tenta o OpenRouter; se esse também falhar, usa o Ollama (IA local na sua máquina). A sessão não cai. Quando o serviço original se recupera, a troca de volta acontece automaticamente a partir da próxima requisição (v0.1.18+).
- **Sincronização em tempo real (SSE)**: se você mudar o modelo no painel do cofre, a mudança aparece no OpenClaw em 1 a 3 segundos. SSE (Server-Sent Events) é uma tecnologia em que o servidor empurra atualizações diretamente para o cliente em tempo real.
- **Notificações em tempo real**: eventos como esgotamento de chave ou falha de serviço aparecem imediatamente na linha de status do TUI (tela de terminal) do OpenClaw.

> 💡 **Claude Code, Cursor e VS Code** também podem ser conectados ao wall-vault, mas o objetivo principal é funcionar junto com o OpenClaw.

```
OpenClaw (tela de terminal TUI)
        │
        ▼
  wall-vault proxy (:56244)   ← gerenciamento de chaves, roteamento, fallback, eventos
        │
        ├─ Google Gemini API
        ├─ OpenRouter API (mais de 340 modelos)
        └─ Ollama (na sua máquina, último recurso)
```

---

## Instalação

### Linux / macOS

Abra o terminal e cole o comando abaixo exatamente como está.

```bash
# Linux (PC comum, servidor — amd64)
curl -L https://github.com/sookmook/wall-vault/releases/latest/download/wall-vault-linux-amd64 \
  -o ~/.local/bin/wall-vault && chmod +x ~/.local/bin/wall-vault

# macOS Apple Silicon (Mac M1/M2/M3)
curl -L https://github.com/sookmook/wall-vault/releases/latest/download/wall-vault-darwin-arm64 \
  -o /usr/local/bin/wall-vault && chmod +x /usr/local/bin/wall-vault
```

- `curl -L ...` — baixa o arquivo da internet.
- `chmod +x` — torna o arquivo baixado "executável". Se pular esse passo, você verá um erro de "permissão negada".

### Windows

Abra o PowerShell como administrador e execute o comando abaixo.

```powershell
# Download
Invoke-WebRequest -Uri `
  "https://github.com/sookmook/wall-vault/releases/latest/download/wall-vault-windows-amd64.exe" `
  -OutFile "$env:LOCALAPPDATA\Programs\wall-vault\wall-vault.exe"

# Adicionar ao PATH (aplicado após reiniciar o PowerShell)
$env:PATH += ";$env:LOCALAPPDATA\Programs\wall-vault"
```

> 💡 **O que é PATH?** É a lista de pastas onde o computador procura por comandos. Ao adicionar o wall-vault ao PATH, você pode digitá-lo em qualquer pasta do terminal e ele será encontrado automaticamente.

### Compilar a partir do código-fonte (para desenvolvedores)

Apenas se você tiver o ambiente de desenvolvimento Go instalado.

```bash
git clone https://github.com/sookmook/wall-vault
cd wall-vault
make build       # bin/wall-vault (versão: v0.1.6.YYYYMMDD.HHmmss)
make install     # ~/.local/bin/wall-vault
```

> 💡 **Versão com timestamp de build**: ao compilar com `make build`, a versão é gerada automaticamente no formato `v0.1.6.20260314.231308`, incluindo data e hora. Se compilar diretamente com `go build ./...`, a versão aparecerá apenas como `"dev"`.

---

## Primeiros passos

### Executar o assistente de configuração

Após a instalação, execute obrigatoriamente o **assistente de configuração** com o comando abaixo. Ele vai guiá-lo passo a passo por todas as configurações necessárias.

```bash
wall-vault setup
```

O assistente percorre as seguintes etapas:

```
1. Escolha do idioma (10 idiomas disponíveis, incluindo português)
2. Escolha do tema (light / dark / gold / cherry / ocean)
3. Modo de operação — sozinho (standalone) ou com múltiplas máquinas (distributed)
4. Nome do bot — o nome exibido no painel
5. Configuração de portas — padrão: proxy 56244, cofre 56243 (pressione Enter para manter)
6. Serviços de IA — escolha quais usar: Google / OpenRouter / Ollama
7. Configuração do filtro de segurança de ferramentas
8. Token de administrador — senha para proteger as funções de gerenciamento do painel (pode ser gerado automaticamente)
9. Senha de criptografia das chaves de API — para armazenamento ainda mais seguro (opcional)
10. Caminho para salvar o arquivo de configuração
```

> ⚠️ **Guarde bem o token de administrador.** Você vai precisar dele para adicionar chaves ou alterar configurações no painel. Se perdê-lo, será necessário editar o arquivo de configuração manualmente.

Ao concluir o assistente, o arquivo `wall-vault.yaml` será criado automaticamente.

### Iniciar o wall-vault

```bash
wall-vault start
```

Dois servidores serão iniciados ao mesmo tempo:

- **Proxy** (`http://localhost:56244`) — intermediário entre o OpenClaw e os serviços de IA
- **Cofre de chaves** (`http://localhost:56243`) — gerenciamento de chaves de API e painel web

Abra `http://localhost:56243` no navegador para acessar o painel imediatamente.

---

## Cadastro de chaves de API

Há quatro formas de cadastrar uma chave de API. **Para quem está começando, recomendamos o Método 1 (variável de ambiente).**

### Método 1: Variável de ambiente (recomendado — mais simples)

Variáveis de ambiente são **valores pré-configurados** que o programa lê ao iniciar. No terminal, basta digitar:

```bash
# Cadastrar chave do Google Gemini
export WV_KEY_GOOGLE=AIzaSy...

# Cadastrar chave do OpenRouter
export WV_KEY_OPENROUTER=sk-or-v1-...

# Iniciar após cadastrar
wall-vault start
```

Se você tiver várias chaves, separe-as por vírgula (,). O wall-vault vai usá-las em rodízio automático (round-robin):

```bash
export WV_KEY_GOOGLE=AIzaSy...,AIzaSy...,AIzaSy...
```

> 💡 **Dica**: o comando `export` só é válido para a sessão de terminal atual. Para que persista após reiniciar o computador, adicione essa linha ao arquivo `~/.bashrc` ou `~/.zshrc`.

### Método 2: Interface do painel (via mouse)

1. Acesse `http://localhost:56243` no navegador
2. No cartão **🔑 Chaves de API** no topo, clique no botão `[+ Adicionar]`
3. Preencha o tipo de serviço, valor da chave, rótulo (nome para referência) e limite diário, depois salve

### Método 3: API REST (para automação e scripts)

A API REST é uma forma de programas se comunicarem via HTTP. Útil para cadastro automatizado via script.

```bash
curl -X POST http://localhost:56243/admin/keys \
  -H "Authorization: Bearer SEU_TOKEN_ADMIN" \
  -H "Content-Type: application/json" \
  -d '{
    "service": "google",
    "key": "AIzaSy...",
    "label": "chave principal",
    "daily_limit": 1000
  }'
```

### Método 4: Flag do proxy (para testes rápidos)

Útil para inserir uma chave temporariamente sem cadastro formal. A chave some quando o programa é encerrado.

```bash
wall-vault proxy --key-google=AIzaSy... --key-openrouter=sk-or-...
```

---

## Como usar o proxy

### Usando com o OpenClaw (uso principal)

Veja como configurar o OpenClaw para se conectar aos serviços de IA através do wall-vault.

Abra o arquivo `~/.openclaw/openclaw.json` e adicione o seguinte conteúdo:

```json5
// ~/.openclaw/openclaw.json
{
  models: {
    providers: {
      "wall-vault": {
        baseUrl: "http://localhost:56244/v1",
        apiKey: "your-agent-token",   // token do agente no cofre
        api: "openai-completions",
        models: [
          { id: "wall-vault/gemini-2.5-flash" },
          { id: "wall-vault/gemini-2.5-pro" },
          { id: "wall-vault/hunter-alpha" },    // contexto gratuito de 1M tokens
          { id: "wall-vault/claude-opus-4-6" }
        ]
      }
    }
  }
}
```

> 💡 **Forma mais fácil**: clique no botão **🦞 Copiar configuração do OpenClaw** no cartão do agente no painel. Um trecho de configuração já preenchido com o token e o endereço será copiado para a área de transferência. Basta colar.

**Para onde o prefixo `wall-vault/` direciona cada modelo?**

O wall-vault determina automaticamente qual serviço de IA usar com base no nome do modelo:

| Formato do modelo | Serviço utilizado |
|-------------------|------------------|
| `wall-vault/gemini-*` | Google Gemini (conexão direta) |
| `wall-vault/gpt-*`, `wall-vault/o3`, `wall-vault/o4*` | OpenAI (conexão direta) |
| `wall-vault/claude-*` | Anthropic via OpenRouter |
| `wall-vault/hunter-alpha`, `wall-vault/healer-alpha` | OpenRouter (contexto gratuito de 1M tokens) |
| `wall-vault/kimi-*`, `wall-vault/glm-*`, `wall-vault/deepseek-*` | OpenRouter |
| `google/nome-do-modelo`, `openai/nome-do-modelo`, `anthropic/nome-do-modelo` etc. | Conexão direta com o serviço indicado |
| `custom/google/nome-do-modelo`, `custom/openai/nome-do-modelo` etc. | Remove o prefixo `custom/` e redireciona |
| `nome-do-modelo:cloud` | Remove `:cloud` e conecta via OpenRouter |

> 💡 **O que é contexto?** É a quantidade de conversa que a IA consegue lembrar de uma só vez. Com 1M (um milhão de tokens), é possível processar conversas muito longas ou documentos extensos de uma só vez.

### Conexão direta no formato da API Gemini (compatibilidade com ferramentas existentes)

Se você já usa alguma ferramenta que se conecta diretamente à API do Google Gemini, basta trocar o endereço para o wall-vault:

```bash
export ANTHROPIC_BASE_URL=http://localhost:56244/google
```

Ou, se a ferramenta permite especificar a URL diretamente:

```
http://localhost:56244/google/v1beta/models/gemini-2.5-flash:generateContent
```

### Usando com o SDK OpenAI (Python)

Você também pode conectar o wall-vault em código Python que usa IA. Basta alterar o `base_url`:

```python
from openai import OpenAI

client = OpenAI(
    base_url="http://localhost:56244/v1",
    api_key="not-needed"  # as chaves de API são gerenciadas pelo wall-vault
)

response = client.chat.completions.create(
    model="google/gemini-2.5-flash",   # formato: provedor/modelo
    messages=[{"role": "user", "content": "Olá!"}]
)
```

### Trocar o modelo com o serviço em execução

Para alterar o modelo de IA enquanto o wall-vault já está rodando:

```bash
# Alterar modelo enviando requisição diretamente ao proxy
curl -X PUT http://localhost:56244/api/config/model \
  -H "Content-Type: application/json" \
  -d '{"service": "openrouter", "model": "anthropic/claude-3.5-sonnet"}'

# No modo distribuído (múltiplos bots), altere pelo servidor do cofre → propagado via SSE imediatamente
curl -X PUT http://localhost:56243/admin/clients/ID-DO-SEU-BOT \
  -H "Authorization: Bearer SEU_TOKEN_ADMIN" \
  -H "Content-Type: application/json" \
  -d '{"default_service": "google", "default_model": "gemini-2.5-pro"}'
```

### Consultar a lista de modelos disponíveis

```bash
# Ver lista completa
curl http://localhost:56244/api/models | python3 -m json.tool

# Ver apenas modelos do Google
curl "http://localhost:56244/api/models?service=google"

# Buscar por nome (exemplo: modelos com "claude" no nome)
curl "http://localhost:56244/api/models?q=claude"
```

**Resumo dos principais modelos por serviço:**

| Serviço | Principais modelos |
|---------|-------------------|
| Google | gemini-2.5-pro, gemini-2.5-flash, gemini-2.5-flash-8b, gemini-2.0-flash |
| OpenAI | gpt-4o, gpt-4o-mini, o3, o1, o1-mini |
| OpenRouter | mais de 346 modelos (Hunter Alpha 1M contexto gratuito, DeepSeek R1/V3, Qwen 2.5, etc.) |
| Ollama | detecta automaticamente o servidor local instalado na sua máquina |

---

## Painel do cofre de chaves

Acesse `http://localhost:56243` no navegador para ver o painel.

**Estrutura da tela:**
- **Barra superior fixa (topbar)**: logotipo, seletor de idioma e tema, indicador de conexão SSE
- **Grade de cartões**: cartões de agentes, serviços e chaves de API organizados em mosaico

### Cartão de chaves de API

Cartão para gerenciar todas as chaves de API registradas de forma centralizada.

- Exibe a lista de chaves separadas por serviço.
- `today_usage`: quantidade de tokens processados com sucesso hoje (unidade de texto lida/gerada pela IA)
- `today_attempts`: total de chamadas feitas hoje (sucessos + falhas)
- Use o botão `[+ Adicionar]` para cadastrar novas chaves e `✕` para excluir.

> 💡 **O que é um token?** É a unidade que a IA usa para processar texto. Equivale aproximadamente a uma palavra em inglês, ou a 1 a 2 caracteres em português. O custo da API geralmente é calculado com base nessa contagem de tokens.

### Cartão de agentes

Cartão que exibe o status dos bots (agentes) conectados ao proxy do wall-vault.

**O status de conexão é mostrado em 4 níveis:**

| Indicador | Status | Significado |
|-----------|--------|-------------|
| 🟢 | Em execução | O proxy está funcionando normalmente |
| 🟡 | Com atraso | Respondendo, mas com lentidão |
| 🔴 | Offline | O proxy não está respondendo |
| ⚫ | Não conectado / inativo | O proxy nunca se conectou ao cofre, ou está desativado |

**Guia dos botões na parte inferior do cartão de agente:**

Ao registrar um agente com um **tipo de agente** específico, os botões de ação correspondentes aparecem automaticamente.

---

#### 🔘 Botão Copiar configuração — gera a configuração de conexão automaticamente

Ao clicar no botão, um trecho de configuração já preenchido com o token, endereço do proxy e informações do modelo é copiado para a área de transferência. Basta colar no local indicado na tabela abaixo para concluir a configuração de conexão.

| Botão | Tipo de agente | Onde colar |
|-------|---------------|------------|
| 🦞 Copiar configuração do OpenClaw | `openclaw` | `~/.openclaw/openclaw.json` |
| 🦀 Copiar configuração do NanoClaw | `nanoclaw` | `~/.openclaw/openclaw.json` |
| 🟠 Copiar configuração do Claude Code | `claude-code` | `~/.claude/settings.json` |
| ⌨ Copiar configuração do Cursor | `cursor` | Cursor → Settings → AI |
| 💻 Copiar configuração do VSCode | `vscode` | `~/.continue/config.json` |

**Exemplo — tipo Claude Code, o seguinte conteúdo é copiado:**

```json
// ~/.claude/settings.json
{
  "apiProvider": "openai",
  "baseUrl": "http://192.168.0.6:56244/v1",
  "apiKey": "token-deste-agente"
}
```

**Exemplo — tipo VSCode (Continue):**

```json
// ~/.continue/config.json
{
  "models": [{
    "title": "wall-vault proxy",
    "provider": "openai",
    "model": "gemini-2.5-flash",
    "apiBase": "http://192.168.0.6:56244/v1",
    "apiKey": "token-deste-agente"
  }]
}
```

**Exemplo — tipo Cursor:**

```
Base URL : http://192.168.0.6:56244/v1
API Key  : token-deste-agente

// Ou via variáveis de ambiente:
OPENAI_BASE_URL=http://192.168.0.6:56244/v1
OPENAI_API_KEY=token-deste-agente
```

> ⚠️ **Se a cópia para área de transferência não funcionar**: políticas de segurança do navegador podem bloquear essa ação. Se uma caixa de texto aparecer em um popup, selecione tudo com Ctrl+A e copie com Ctrl+C.

---

#### ⚡ Botão de aplicação automática — um clique e a configuração está pronta

Quando o tipo de agente é `cline`, `claude-code`, `openclaw` ou `nanoclaw`, o botão **⚡ Aplicar configuração** aparece no cartão do agente. Ao clicar, o arquivo de configuração local do agente é atualizado automaticamente.

| Botão | Tipo de agente | Arquivo alvo |
|-------|---------------|--------------|
| ⚡ Aplicar configuração do Cline | `cline` | `~/.cline/data/globalState.json` + `secrets.json` |
| ⚡ Aplicar configuração do Claude Code | `claude-code` | `~/.claude/settings.json` |
| ⚡ Aplicar configuração do OpenClaw | `openclaw` | `~/.openclaw/openclaw.json` |
| ⚡ Aplicar configuração do NanoClaw | `nanoclaw` | `~/.openclaw/openclaw.json` |

> ⚠️ Este botão envia uma requisição para **localhost:56244** (proxy local). O proxy precisa estar em execução nesta máquina para que funcione.

---

#### 🔀 Ordenação de cartões por arrastar e soltar (v0.1.17)

Você pode **arrastar** os cartões de agentes no painel para reorganizá-los na ordem desejada.

1. Segure um cartão de agente com o mouse e arraste-o
2. Solte-o sobre o cartão na posição desejada para alterar a ordem
3. A ordem alterada é **salva imediatamente no servidor** e mantida mesmo após atualizar a página

> 💡 Dispositivos com tela sensível ao toque (celular/tablet) ainda não são suportados. Use um navegador de desktop.

---

#### 🔄 Sincronização bidirecional de modelos (v0.1.16)

Quando você altera o modelo de um agente no painel do cofre, a configuração local desse agente é atualizada automaticamente.

**Para o Cline:**
- Alteração do modelo no cofre → evento SSE → o proxy atualiza o campo de modelo em `globalState.json`
- Campos atualizados: `actModeOpenAiModelId`, `planModeOpenAiModelId`, `openAiModelId`
- `openAiBaseUrl` e a chave de API não são alterados
- **É necessário recarregar o VS Code (`Ctrl+Alt+R` ou `Ctrl+Shift+P` → `Developer: Reload Window`)**
  - O Cline não relê o arquivo de configuração durante a execução

**Para o Claude Code:**
- Alteração do modelo no cofre → evento SSE → o proxy atualiza o campo `model` em `settings.json`
- Os caminhos WSL e Windows são verificados automaticamente (`~/.claude/`, `/mnt/c/Users/*/.claude/`)

**Direção inversa (agente → cofre):**
- Quando um agente (Cline, Claude Code etc.) envia uma requisição ao proxy, o proxy inclui as informações de serviço e modelo do cliente no heartbeat
- O serviço e modelo atualmente em uso são exibidos em tempo real no cartão do agente no painel

> 💡 **Ponto-chave**: o proxy identifica o agente pelo token de autorização da requisição e roteia automaticamente para o serviço/modelo configurado no cofre. Mesmo que o Cline ou Claude Code envie um nome de modelo diferente, o proxy substitui pela configuração do cofre.

---

### Usando o Cline no VS Code — guia detalhado

#### Passo 1: Instalar o Cline

Instale o **Cline** (ID: `saoudrizwan.claude-dev`) pelo marketplace de extensões do VS Code.

#### Passo 2: Registrar o agente no cofre

1. Abra o painel do cofre (`http://IP-do-cofre:56243`)
2. Na seção **Agentes**, clique em **+ Adicionar**
3. Preencha os seguintes campos:

| Campo | Valor | Descrição |
|-------|-------|-----------|
| ID | `meu_cline` | Identificador único (letras, sem espaços) |
| Nome | `Meu Cline` | Nome exibido no painel |
| Tipo de agente | `cline` | ← selecione obrigatoriamente `cline` |
| Serviço | Selecione o serviço desejado (ex.: `google`) | |
| Modelo | Insira o modelo desejado (ex.: `gemini-2.5-flash`) | |

4. Clique em **Salvar** — um token será gerado automaticamente

#### Passo 3: Conectar o Cline

**Método A — Aplicação automática (recomendado)**

1. Verifique se o **proxy** do wall-vault está em execução nesta máquina (`localhost:56244`)
2. Clique no botão **⚡ Aplicar configuração do Cline** no cartão do agente
3. Se a mensagem "Configuração aplicada!" aparecer, foi bem-sucedido
4. Recarregue o VS Code (`Ctrl+Alt+R`)

**Método B — Configuração manual**

Abra as configurações (⚙️) na barra lateral do Cline e configure:
- **API Provider**: `OpenAI Compatible`
- **Base URL**: `http://endereco-do-proxy:56244/v1`
  - Na mesma máquina: `http://localhost:56244/v1`
  - Em outro dispositivo (ex.: Mac Mini): `http://192.168.0.6:56244/v1`
- **API Key**: o token emitido pelo cofre (copiado do cartão do agente)
- **Model ID**: o modelo configurado no cofre (ex.: `gemini-2.5-flash`)

#### Passo 4: Verificação

Envie qualquer mensagem no chat do Cline. Se estiver funcionando:
- Um **ponto verde (● Em execução)** aparecerá no cartão do agente no painel
- O serviço e modelo atuais serão exibidos no cartão (ex.: `google / gemini-2.5-flash`)

#### Trocar o modelo

Para trocar o modelo do Cline, altere-o no **painel do cofre**:

1. Altere o serviço/modelo no menu suspenso do cartão do agente
2. Clique em **Aplicar**
3. Recarregue o VS Code (`Ctrl+Alt+R`) — o nome do modelo no rodapé do Cline será atualizado
4. A partir da próxima requisição, o novo modelo será utilizado

> 💡 Na prática, o proxy identifica as requisições do Cline pelo token e as roteia para o modelo configurado no cofre. Mesmo sem recarregar o VS Code, **o modelo realmente utilizado muda imediatamente** — o recarregamento serve apenas para atualizar a exibição na interface do Cline.

#### Detecção de desconexão

Ao fechar o VS Code, o cartão do agente no painel muda para amarelo (com atraso) após aproximadamente **90 segundos** e para vermelho (offline) após **3 minutos**. (Desde a v0.1.18, o intervalo de verificação de status de 15 segundos permite uma detecção de offline mais rápida.)

#### Solução de problemas

| Sintoma | Causa | Solução |
|---------|-------|---------|
| Erro "falha na conexão" no Cline | Proxy não iniciado ou endereço incorreto | Verifique o proxy com `curl http://localhost:56244/health` |
| Ponto verde não aparece no cofre | Chave de API (token) não configurada | Clique novamente em **⚡ Aplicar configuração do Cline** |
| Nome do modelo no rodapé do Cline não muda | O Cline mantém a configuração em cache | Recarregue o VS Code (`Ctrl+Alt+R`) |
| Um nome de modelo incorreto é exibido | Bug antigo (corrigido na v0.1.16) | Atualize o proxy para a versão v0.1.16 ou superior |

---

#### 🟣 Botão Copiar comando de implantação — para instalar em uma nova máquina

Use este botão ao instalar o proxy do wall-vault em um novo computador e conectá-lo ao cofre pela primeira vez. Ao clicar, o script de instalação completo é copiado para a área de transferência. Cole-o no terminal do novo computador e execute — as seguintes etapas serão realizadas de uma só vez:

1. Instalação do binário do wall-vault (pulada se já estiver instalado)
2. Registro automático como serviço de usuário do systemd
3. Inicialização do serviço e conexão automática com o cofre

> 💡 O script já vem preenchido com o token deste agente e o endereço do servidor do cofre. Não é necessário fazer nenhuma alteração antes de executar.

---

### Cartão de serviços

Cartão para ativar, desativar ou configurar os serviços de IA disponíveis.

- Chave de ativação/desativação por serviço
- Ao informar o endereço de um servidor de IA local (Ollama, LM Studio, vLLM etc. rodando na sua máquina), os modelos disponíveis são detectados automaticamente.
- **Indicador de status do serviço local**: o ponto ● ao lado do nome do serviço fica **verde** quando conectado e **cinza** quando desconectado.
- **Indicador de status do serviço local**: ao abrir a página, se um serviço local (como Ollama) estiver em execução, o ponto ● fica verde — mas o estado da caixa de seleção não é alterado.

> 💡 **Se o serviço local estiver rodando em outro computador**: insira o IP daquela máquina no campo de URL do serviço. Exemplos: `http://192.168.0.6:11434` (Ollama), `http://192.168.0.6:1234` (LM Studio)

### Inserção do token de administrador

Ao tentar usar funções importantes no painel — como adicionar ou excluir chaves — um popup solicitará o token de administrador. Insira o token definido durante o assistente de configuração. Uma vez inserido, ele permanece ativo até você fechar o navegador.

> ⚠️ **Se houver mais de 10 falhas de autenticação em 15 minutos, o IP correspondente será temporariamente bloqueado.** Se esquecer o token, verifique o item `admin_token` no arquivo `wall-vault.yaml`.

---

## Modo distribuído (múltiplos bots)

Quando você opera o OpenClaw em várias máquinas simultaneamente, é possível **compartilhar um único cofre de chaves** entre todas elas. Assim, o gerenciamento de chaves fica centralizado em um só lugar.

### Exemplo de configuração

```
[Servidor do cofre]
  wall-vault vault    (cofre de chaves :56243, painel)

[WSL Alpha]             [Raspberry Pi Gamma]    [Mac Mini Local]
  wall-vault proxy        wall-vault proxy         wall-vault proxy
  openclaw TUI            openclaw TUI             openclaw TUI
  ↕ sync via SSE          ↕ sync via SSE           ↕ sync via SSE
```

Todos os bots apontam para o mesmo servidor do cofre. Qualquer mudança de modelo ou adição de chave no cofre é propagada imediatamente para todos os bots.

### Passo 1: Iniciar o servidor do cofre

Execute no computador que será o servidor do cofre:

```bash
wall-vault vault
```

### Passo 2: Registrar cada bot (cliente)

Cadastre antecipadamente as informações de cada bot que se conectará ao servidor do cofre:

```bash
curl -X POST http://localhost:56243/admin/clients \
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

### Passo 3: Iniciar o proxy em cada máquina do bot

Em cada computador onde o bot está instalado, inicie o proxy especificando o endereço e o token do servidor do cofre:

```bash
WV_VAULT_URL=http://192.168.x.x:56243 \
WV_VAULT_TOKEN=bota-secret \
WV_VAULT_CLIENT_ID=botA \
wall-vault proxy
```

> 💡 Substitua **`192.168.x.x`** pelo endereço IP interno real do servidor do cofre. Você pode verificá-lo nas configurações do roteador ou com o comando `ip addr`.

---

## Configuração de inicialização automática

Se for cansativo iniciar o wall-vault manualmente toda vez que o computador liga, registre-o como serviço do sistema. Uma vez registrado, ele inicia automaticamente na inicialização.

### Linux — systemd (maioria das distribuições Linux)

O systemd é o sistema que gerencia a inicialização e execução automática de programas no Linux:

```bash
wall-vault doctor deploy
systemctl --user daemon-reload
systemctl --user enable --now wall-vault
```

Verificar os logs:

```bash
journalctl --user -u wall-vault -f
```

### macOS — launchd

O launchd é o sistema responsável pela execução automática de programas no macOS:

```bash
wall-vault doctor deploy launchd
launchctl load ~/Library/LaunchAgents/com.wall-vault.plist
```

### Windows — NSSM

1. Baixe o NSSM em [nssm.cc](https://nssm.cc/download) e adicione-o ao PATH.
2. No PowerShell com permissões de administrador:

```powershell
wall-vault doctor deploy windows
```

---

## Doctor — diagnóstico automático

O comando `doctor` é uma **ferramenta de autodiagnóstico e correção** que verifica se o wall-vault está configurado corretamente.

```bash
wall-vault doctor check   # Diagnostica o estado atual (somente leitura, sem alterações)
wall-vault doctor fix     # Corrige problemas automaticamente
wall-vault doctor all     # Diagnóstico + correção automática em uma só etapa
```

> 💡 Se algo parecer errado, execute `wall-vault doctor all` primeiro. Ele resolve muitos problemas automaticamente.

---

## Variáveis de ambiente

Variáveis de ambiente são uma forma de passar configurações para o programa. Você pode digitá-las no terminal como `export NOME_DA_VARIAVEL=valor`, ou incluí-las no arquivo do serviço de inicialização automática para que sejam aplicadas sempre.

| Variável | Descrição | Exemplo de valor |
|----------|-----------|-----------------|
| `WV_LANG` | Idioma do painel | `pt`, `en`, `ko`, `ja` |
| `WV_THEME` | Tema do painel | `light`, `dark`, `gold` |
| `WV_KEY_GOOGLE` | Chave(s) de API do Google (separe por vírgula) | `AIza...,AIza...` |
| `WV_KEY_OPENROUTER` | Chave de API do OpenRouter | `sk-or-v1-...` |
| `WV_VAULT_URL` | Endereço do servidor do cofre no modo distribuído | `http://192.168.x.x:56243` |
| `WV_VAULT_TOKEN` | Token de autenticação do cliente (bot) | `my-secret-token` |
| `WV_ADMIN_TOKEN` | Token de administrador | `admin-token-here` |
| `WV_MASTER_PASS` | Senha de criptografia das chaves de API | `my-password` |
| `WV_AVATAR` | Caminho do arquivo de avatar (relativo a `~/.openclaw/`) | `workspace/avatars/avatar.png` |
| `OLLAMA_URL` | Endereço do servidor local do Ollama | `http://192.168.x.x:11434` |

---

## Solução de problemas

### O proxy não inicia

O problema mais comum é a porta já estar em uso por outro programa.

```bash
ss -tlnp | grep 56244   # verifica qual programa está usando a porta 56244
wall-vault proxy --port 8080   # inicia em outra porta
```

### Erros de chave de API (429, 402, 401, 403, 582)

| Código de erro | Significado | O que fazer |
|----------------|-------------|-------------|
| **429** | Muitas requisições (limite de uso atingido) | Aguarde um momento ou adicione outra chave |
| **402** | Pagamento necessário ou créditos insuficientes | Recarregue créditos no serviço correspondente |
| **401 / 403** | Chave inválida ou sem permissão | Verifique o valor da chave e recadastre |
| **582** | Sobrecarga no gateway (cooldown de 5 minutos) | Liberado automaticamente após 5 minutos |

```bash
# Verificar lista de chaves cadastradas e seus status
curl -H "Authorization: Bearer SEU_TOKEN_ADMIN" http://localhost:56243/admin/keys

# Resetar contadores de uso das chaves
curl -X POST -H "Authorization: Bearer SEU_TOKEN_ADMIN" http://localhost:56243/admin/keys/reset
```

### O agente aparece como "não conectado"

"Não conectado" significa que o processo do proxy não está enviando sinais (heartbeat) para o cofre. **Isso não significa que as configurações foram perdidas.** O proxy precisa saber o endereço do servidor do cofre e o token para que o status mude para conectado.

```bash
# Iniciar o proxy especificando o endereço do cofre, token e ID do cliente
WV_VAULT_URL=http://ENDERECO-DO-COFRE:56243 \
WV_VAULT_TOKEN=TOKEN-DO-CLIENTE \
WV_VAULT_CLIENT_ID=ID-DO-CLIENTE \
wall-vault proxy
```

Se a conexão for bem-sucedida, o painel mostrará 🟢 Em execução em cerca de 20 segundos.

### Ollama não conecta

O Ollama é um programa que roda IA diretamente na sua máquina. Primeiro, verifique se o Ollama está em execução.

```bash
curl http://localhost:11434/api/tags   # se retornar a lista de modelos, está funcionando
export OLLAMA_URL=http://192.168.x.x:11434   # se estiver rodando em outra máquina
```

> ⚠️ Se o Ollama não responder, inicie-o primeiro com o comando `ollama serve`.

> ⚠️ **Modelos grandes são lentos**: modelos como `qwen3.5:35b` e `deepseek-r1` podem levar vários minutos para gerar uma resposta. Se parecer que não está respondendo, pode ser processamento normal — aguarde com paciência.

---

## Alterações recentes (v0.1.16 ~ v0.1.21)

### v0.1.21 (2026-04-05)
- **Suporte a modelos Gemma 4**: modelos Gemma (gemma-4-31b-it, gemma-4-26b-a4b-it) agora são roteados pela API Google Gemini.
- **Suporte a LM Studio / vLLM**: esses serviços locais agora são despachados corretamente em vez de cair para o Ollama.
- **Correção do painel**: sempre exibe o serviço configurado, não o serviço de fallback.
- **Caixa de seleção do serviço local preservada**: o painel não desativa mais automaticamente os serviços locais ao carregar a página.
- **Variável de ambiente de filtro de ferramentas**: suporte a `WV_TOOL_FILTER=passthrough`.

### v0.1.20 (2026-03-28)
- **Reforço abrangente de segurança**: prevenção de XSS (41 pontos), comparação de tokens em tempo constante, restrição CORS, limites de tamanho de requisição e mais.

### v0.1.19 (2026-03-27)
- **Detecção online do Claude Code**: Claude Code aparece como online no painel mesmo quando está contornando o proxy.

### v0.1.18 (2026-03-26)
- **Correção de recuperação de fallback**: recupera automaticamente para o serviço preferido quando disponível.
- **Detecção offline aprimorada**: polling de status a cada 15 segundos.

### v0.1.17 (2026-03-25)
- **Reordenação de cartões por arrastar e soltar**.
- **Botões de aplicação inline para agentes desconectados**.
- **Tipo de agente cokacdir adicionado**.

### v0.1.16 (2026-03-25)
- **Sincronização bidirecional de modelos** para Cline e Claude Code.

---

*Para informações detalhadas sobre a API, consulte [API.md](API.md).*
