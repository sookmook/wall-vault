# Manual do Usuário wall-vault
*(Last updated: 2026-04-06 — v0.1.23)*

---

## Índice

1. [O que é wall-vault?](#o-que-é-wall-vault)
2. [Instalação](#instalação)
3. [Primeiros passos (assistente de setup)](#primeiros-passos)
4. [Registro de chaves de API](#registro-de-chaves-de-api)
5. [Como usar o proxy](#como-usar-o-proxy)
6. [Painel do cofre de chaves](#painel-do-cofre-de-chaves)
7. [Modo distribuído (múltiplos bots)](#modo-distribuído-múltiplos-bots)
8. [Configuração de inicialização automática](#configuração-de-inicialização-automática)
9. [Doctor (diagnóstico)](#doctor-diagnóstico)
10. [Referência de variáveis de ambiente](#referência-de-variáveis-de-ambiente)
11. [Solução de problemas](#solução-de-problemas)

---

## O que é wall-vault?

**wall-vault = Proxy de IA + Cofre de chaves de API para o OpenClaw**

Para usar serviços de IA, você precisa de uma **chave de API**. Uma chave de API é como um **crachá digital** que comprova que "esta pessoa está autorizada a usar este serviço". Porém, esse crachá tem um limite de uso diário e pode ser exposto se mal gerenciado.

O wall-vault armazena esses crachás em um cofre seguro e atua como **proxy (intermediário)** entre o OpenClaw e os serviços de IA. Em resumo, o OpenClaw só precisa se conectar ao wall-vault, e o wall-vault cuida de todo o resto.

Problemas que o wall-vault resolve:

- **Rotação automática de chaves de API**: Quando uma chave atinge seu limite de uso ou é temporariamente bloqueada (cooldown), ele silenciosamente muda para a próxima chave. O OpenClaw continua funcionando sem interrupção.
- **Substituição automática de serviço (fallback)**: Se o Google não responder, ele muda para o OpenRouter; se isso também falhar, muda automaticamente para IA local instalada no seu computador (Ollama, LM Studio, vLLM). A sessão não é interrompida. Quando o serviço original se recupera, ele volta automaticamente a partir da próxima requisição (v0.1.18+, LM Studio/vLLM: v0.1.21+).
- **Sincronização em tempo real (SSE)**: Quando você muda o modelo no painel do cofre, a mudança é refletida na tela do OpenClaw em 1 a 3 segundos. SSE (Server-Sent Events) é uma tecnologia que permite ao servidor enviar atualizações em tempo real para o cliente.
- **Notificações em tempo real**: Eventos como esgotamento de chaves ou falhas de serviço são exibidos imediatamente na parte inferior do TUI (interface de terminal) do OpenClaw.

> 💡 **Claude Code, Cursor e VS Code** também podem ser conectados, mas o propósito original do wall-vault é ser usado com o OpenClaw.

```
OpenClaw (interface TUI no terminal)
        │
        ▼
  wall-vault proxy (:56244)   ← gerenciamento de chaves, roteamento, fallback, eventos
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

# macOS Apple Silicon (M1/M2/M3 Mac)
curl -L https://github.com/sookmook/wall-vault/releases/latest/download/wall-vault-darwin-arm64 \
  -o /usr/local/bin/wall-vault && chmod +x /usr/local/bin/wall-vault
```

- `curl -L ...` — Baixa o arquivo da internet.
- `chmod +x` — Torna o arquivo baixado "executável". Se pular este passo, ocorrerá um erro de "permissão negada".

### Windows

Abra o PowerShell (como administrador) e execute os comandos abaixo:

```powershell
# Download
Invoke-WebRequest -Uri `
  "https://github.com/sookmook/wall-vault/releases/latest/download/wall-vault-windows-amd64.exe" `
  -OutFile "$env:LOCALAPPDATA\Programs\wall-vault\wall-vault.exe"

# Adicionar ao PATH (aplica após reiniciar o PowerShell)
$env:PATH += ";$env:LOCALAPPDATA\Programs\wall-vault"
```

> 💡 **O que é PATH?** É a lista de pastas onde o computador procura por comandos. Ao adicionar ao PATH, você pode executar `wall-vault` de qualquer pasta simplesmente digitando o nome.

### Compilar a partir do código-fonte (para desenvolvedores)

Aplicável apenas se você tiver o ambiente de desenvolvimento Go instalado.

```bash
git clone https://github.com/sookmook/wall-vault
cd wall-vault
make build       # bin/wall-vault (versão: v0.1.23.YYYYMMDD.HHmmss)
make install     # ~/.local/bin/wall-vault
```

> 💡 **Versão com timestamp de compilação**: Ao compilar com `make build`, a versão é gerada automaticamente em formato que inclui data e hora, como `v0.1.23.20260406.211004`. Se você compilar diretamente com `go build ./...`, a versão será exibida apenas como `"dev"`.

---

## Primeiros passos

### Executar o assistente de setup

Após a instalação, execute obrigatoriamente o **assistente de configuração** com o comando abaixo. O assistente irá guiá-lo perguntando cada item necessário.

```bash
wall-vault setup
```

As etapas que o assistente percorre são:

```
1. Seleção de idioma (10 idiomas incluindo coreano)
2. Seleção de tema (light / dark / gold / cherry / ocean)
3. Modo de operação — uso individual (standalone) ou compartilhado em várias máquinas (distributed)
4. Nome do bot — nome exibido no painel
5. Configuração de portas — padrão: proxy 56244, cofre 56243 (pressione Enter se não precisar alterar)
6. Seleção de serviços de IA — Google / OpenRouter / Ollama / LM Studio / vLLM
7. Configuração do filtro de segurança de ferramentas
8. Configuração do token de administrador — senha que bloqueia funções de gerenciamento do painel. Geração automática disponível
9. Configuração da senha de criptografia de chaves de API — para armazenamento mais seguro (opcional)
10. Caminho para salvar o arquivo de configuração
```

> ⚠️ **Lembre-se de guardar o token de administrador.** Ele será necessário mais tarde para adicionar chaves ou alterar configurações no painel. Se você perdê-lo, terá que editar o arquivo de configuração manualmente.

Após concluir o assistente, o arquivo de configuração `wall-vault.yaml` será gerado automaticamente.

### Execução

```bash
wall-vault start
```

Os dois servidores abaixo iniciam simultaneamente:

- **Proxy** (`http://localhost:56244`) — intermediário que conecta o OpenClaw aos serviços de IA
- **Cofre de chaves** (`http://localhost:56243`) — gerenciamento de chaves de API e painel web

Abra `http://localhost:56243` no navegador para acessar o painel imediatamente.

---

## Registro de chaves de API

Existem quatro maneiras de registrar chaves de API. **Para iniciantes, recomendamos o método 1 (variáveis de ambiente).**

### Método 1: Variáveis de ambiente (recomendado — mais simples)

Variáveis de ambiente são **valores pré-configurados** que o programa lê ao iniciar. Basta digitar no terminal como abaixo:

```bash
# Registrar chave do Google Gemini
export WV_KEY_GOOGLE=AIzaSy...

# Registrar chave do OpenRouter
export WV_KEY_OPENROUTER=sk-or-v1-...

# Executar após o registro
wall-vault start
```

Se você tiver várias chaves, separe-as com vírgula (,). O wall-vault as usará automaticamente em rotação (round robin):

```bash
export WV_KEY_GOOGLE=AIzaSy...,AIzaSy...,AIzaSy...
```

> 💡 **Dica**: O comando `export` se aplica apenas à sessão atual do terminal. Para que persista após reiniciar o computador, adicione a linha acima ao arquivo `~/.bashrc` ou `~/.zshrc`.

### Método 2: Interface do painel (clique com o mouse)

1. Acesse `http://localhost:56243` no navegador
2. Clique no botão `[+ Adicionar]` no card **🔑 Chaves de API** no topo
3. Insira o tipo de serviço, valor da chave, rótulo (nome para referência) e limite diário, depois salve

### Método 3: REST API (para automação/scripts)

REST API é uma forma de programas trocarem dados via HTTP. Útil para registro automático via script.

```bash
curl -X POST http://localhost:56243/admin/keys \
  -H "Authorization: Bearer token-admin" \
  -H "Content-Type: application/json" \
  -d '{
    "service": "google",
    "key": "AIzaSy...",
    "label": "Chave principal",
    "daily_limit": 1000
  }'
```

### Método 4: Flag do proxy (para testes rápidos)

Use para testar temporariamente inserindo uma chave sem registro formal. A chave desaparece quando o programa é encerrado.

```bash
wall-vault proxy --key-google=AIzaSy... --key-openrouter=sk-or-...
```

---

## Como usar o proxy

### Uso com OpenClaw (propósito principal)

Veja como configurar o OpenClaw para se conectar aos serviços de IA através do wall-vault.

Abra o arquivo `~/.openclaw/openclaw.json` e adicione o seguinte conteúdo:

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
          { id: "wall-vault/hunter-alpha" },    // 1M context gratuito
          { id: "wall-vault/claude-opus-4-6" }
        ]
      }
    }
  }
}
```

> 💡 **Maneira mais fácil**: Pressione o botão **🦞 Copiar configuração OpenClaw** no card do agente no painel e um snippet com o token e endereço já preenchidos será copiado para a área de transferência. Basta colar.

**Para onde o `wall-vault/` no início do nome do modelo direciona?**

O wall-vault determina automaticamente para qual serviço de IA enviar a requisição com base no nome do modelo:

| Formato do modelo | Serviço conectado |
|----------|--------------|
| `wall-vault/gemini-*` | Conexão direta com Google Gemini |
| `wall-vault/gpt-*`, `wall-vault/o3`, `wall-vault/o4*` | Conexão direta com OpenAI |
| `wall-vault/claude-*` | Conexão com Anthropic via OpenRouter |
| `wall-vault/hunter-alpha`, `wall-vault/healer-alpha` | OpenRouter (1 milhão de tokens de contexto gratuitos) |
| `wall-vault/kimi-*`, `wall-vault/glm-*`, `wall-vault/deepseek-*` | Conexão via OpenRouter |
| `google/nome-modelo`, `openai/nome-modelo`, `anthropic/nome-modelo` etc. | Conexão direta com o serviço correspondente |
| `custom/google/nome-modelo`, `custom/openai/nome-modelo` etc. | Remove a parte `custom/` e redireciona |
| `nome-modelo:cloud` | Remove a parte `:cloud` e conecta via OpenRouter |

> 💡 **O que é contexto?** É a quantidade de conversa que a IA consegue "lembrar" de uma vez. 1M (um milhão de tokens) permite processar conversas muito longas ou documentos extensos de uma só vez.

### Conexão direta no formato da API Gemini (compatibilidade com ferramentas existentes)

Se você já tinha ferramentas que usavam a API do Google Gemini diretamente, basta trocar o endereço para o wall-vault:

```bash
export ANTHROPIC_BASE_URL=http://localhost:56244/google
```

Ou, para ferramentas que especificam a URL diretamente:

```
http://localhost:56244/google/v1beta/models/gemini-2.5-flash:generateContent
```

### Uso com OpenAI SDK (Python)

Você pode conectar o wall-vault também em código Python que utiliza IA. Basta alterar o `base_url`:

```python
from openai import OpenAI

client = OpenAI(
    base_url="http://localhost:56244/v1",
    api_key="not-needed"  # As chaves de API são gerenciadas pelo wall-vault
)

response = client.chat.completions.create(
    model="google/gemini-2.5-flash",   # Insira no formato provider/model
    messages=[{"role": "user", "content": "Olá"}]
)
```

### Trocar o modelo durante a execução

Para trocar o modelo de IA enquanto o wall-vault já está em execução:

```bash
# Alterar o modelo fazendo requisição direta ao proxy
curl -X PUT http://localhost:56244/api/config/model \
  -H "Content-Type: application/json" \
  -d '{"service": "openrouter", "model": "anthropic/claude-3.5-sonnet"}'

# No modo distribuído (múltiplos bots), altere no servidor do cofre → refletido instantaneamente via SSE
curl -X PUT http://localhost:56243/admin/clients/meu-bot-id \
  -H "Authorization: Bearer token-admin" \
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
| OpenRouter | Mais de 346 (Hunter Alpha 1M contexto gratuito, DeepSeek R1/V3, Qwen 2.5 etc.) |
| Ollama | Detecção automática de servidor local instalado no computador |
| LM Studio | Servidor local no computador (porta 1234) |
| vLLM | Servidor local no computador (porta 8000) |

---

## Painel do cofre de chaves

Acesse `http://localhost:56243` no navegador para ver o painel.

**Estrutura da tela:**
- **Barra superior fixa (topbar)**: Logo, seletores de idioma e tema, indicador de status da conexão SSE
- **Grade de cards**: Cards de agentes, serviços e chaves de API dispostos em formato de blocos

### Card de chaves de API

Card onde você pode gerenciar todas as chaves de API registradas de uma só vez.

- Mostra a lista de chaves separadas por serviço.
- `today_usage`: Tokens (número de caracteres lidos e escritos pela IA) processados com sucesso hoje
- `today_attempts`: Total de chamadas hoje (incluindo sucesso + falha)
- Botão `[+ Adicionar]` para registrar novas chaves e `✕` para excluí-las.

> 💡 **O que é um token?** É a unidade usada pela IA para processar texto. Corresponde aproximadamente a uma palavra em inglês ou 1-2 caracteres em coreano. As tarifas de API geralmente são calculadas com base neste número de tokens.

### Card de agente

Card que mostra o status dos bots (agentes) conectados ao proxy do wall-vault.

**O status da conexão é exibido em 4 níveis:**

| Indicador | Status | Significado |
|------|------|------|
| 🟢 | Em execução | O proxy está funcionando normalmente |
| 🟡 | Atrasado | Responde, mas com lentidão |
| 🔴 | Offline | O proxy não está respondendo |
| ⚫ | Não conectado/Desativado | O proxy nunca se conectou ao cofre ou está desativado |

**Guia dos botões na parte inferior do card de agente:**

Ao registrar um agente, se você especificar o **tipo de agente**, botões de conveniência correspondentes ao tipo aparecem automaticamente.

---

#### 🔘 Botão copiar configuração — Gera automaticamente a configuração de conexão

Ao clicar no botão, um snippet de configuração com o token, endereço do proxy e informações do modelo já preenchidos é copiado para a área de transferência. Basta colar o conteúdo copiado no local indicado na tabela abaixo para completar a configuração de conexão.

| Botão | Tipo de agente | Local para colar |
|------|-------------|-------------|
| 🦞 Copiar config OpenClaw | `openclaw` | `~/.openclaw/openclaw.json` |
| 🦀 Copiar config NanoClaw | `nanoclaw` | `~/.openclaw/openclaw.json` |
| 🟠 Copiar config Claude Code | `claude-code` | `~/.claude/settings.json` |
| ⌨ Copiar config Cursor | `cursor` | Cursor → Settings → AI |
| 💻 Copiar config VSCode | `vscode` | `~/.continue/config.json` |

**Exemplo — Se o tipo for Claude Code, o seguinte conteúdo é copiado:**

```json
// ~/.claude/settings.json
{
  "apiProvider": "openai",
  "baseUrl": "http://192.168.0.6:56244/v1",
  "apiKey": "token-deste-agente"
}
```

**Exemplo — Se o tipo for VSCode (Continue):**

```yaml
# ~/.continue/config.yaml  ← Colar em config.yaml, não config.json
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

**Exemplo — Se o tipo for Cursor:**

```
Base URL : http://192.168.0.6:56244/v1
API Key  : token-deste-agente

// Ou via variáveis de ambiente:
OPENAI_BASE_URL=http://192.168.0.6:56244/v1
OPENAI_API_KEY=token-deste-agente
```

> ⚠️ **Quando a cópia para a área de transferência não funcionar**: A política de segurança do navegador pode bloquear a cópia. Se uma caixa de texto aparecer em popup, selecione tudo com Ctrl+A e copie com Ctrl+C.

---

#### ⚡ Botão de aplicação automática — Uma vez pressionado, a configuração está pronta

Para agentes do tipo `cline`, `claude-code`, `openclaw` ou `nanoclaw`, um botão **⚡ Aplicar configuração** é exibido no card do agente. Ao pressionar este botão, o arquivo de configuração local do agente é automaticamente atualizado.

| Botão | Tipo de agente | Arquivo alvo |
|------|-------------|-------------|
| ⚡ Aplicar config Cline | `cline` | `~/.cline/data/globalState.json` + `secrets.json` |
| ⚡ Aplicar config Claude Code | `claude-code` | `~/.claude/settings.json` |
| ⚡ Aplicar config OpenClaw | `openclaw` | `~/.openclaw/openclaw.json` |
| ⚡ Aplicar config NanoClaw | `nanoclaw` | `~/.openclaw/openclaw.json` |

> ⚠️ Este botão envia requisições para **localhost:56244** (proxy local). O proxy precisa estar em execução nessa máquina para funcionar.

---

#### 🔀 Ordenação de cards por arrastar e soltar (v0.1.17)

Você pode **arrastar** os cards de agente no painel para reordená-los como desejar.

1. Clique e segure o card do agente com o mouse e arraste
2. Solte sobre o card na posição desejada para trocar a ordem
3. A nova ordem é **salva instantaneamente no servidor** e persiste mesmo ao atualizar a página

> 💡 Em dispositivos touch (celular/tablet) ainda não é suportado. Use em um navegador desktop.

---

#### 🔄 Sincronização bidirecional de modelo (v0.1.16)

Quando você muda o modelo de um agente no painel do cofre, a configuração local do agente é automaticamente atualizada.

**No caso do Cline:**
- Mudança de modelo no cofre → evento SSE → proxy atualiza o campo de modelo no `globalState.json`
- Campos atualizados: `actModeOpenAiModelId`, `planModeOpenAiModelId`, `openAiModelId`
- `openAiBaseUrl` e chave de API não são alterados
- **É necessário recarregar o VS Code (`Ctrl+Alt+R` ou `Ctrl+Shift+P` → `Developer: Reload Window`)**
  - Porque o Cline não relê o arquivo de configuração durante a execução

**No caso do Claude Code:**
- Mudança de modelo no cofre → evento SSE → proxy atualiza o campo `model` no `settings.json`
- Busca automática em ambos os caminhos WSL e Windows (`~/.claude/`, `/mnt/c/Users/*/.claude/`)

**Direção oposta (agente → cofre):**
- Quando um agente (Cline, Claude Code etc.) envia uma requisição ao proxy, o proxy inclui as informações de serviço/modelo do cliente no heartbeat
- O serviço/modelo atualmente em uso é exibido em tempo real no card do agente no painel do cofre

> 💡 **Ponto-chave**: O proxy identifica o agente pelo token de Authorization da requisição e faz o roteamento automático para o serviço/modelo configurado no cofre. Mesmo que o Cline ou Claude Code envie um nome de modelo diferente, o proxy o substitui pela configuração do cofre.

---

### Usando Cline no VS Code — Guia detalhado

#### Etapa 1: Instalar o Cline

Instale o **Cline** (ID: `saoudrizwan.claude-dev`) no marketplace de extensões do VS Code.

#### Etapa 2: Registrar o agente no cofre

1. Abra o painel do cofre (`http://IP-do-cofre:56243`)
2. Na seção **Agentes**, clique em **+ Adicionar**
3. Preencha da seguinte forma:

| Campo | Valor | Descrição |
|------|----|------|
| ID | `meu_cline` | Identificador único (em inglês, sem espaços) |
| Nome | `Meu Cline` | Nome exibido no painel |
| Tipo de agente | `cline` | ← Deve selecionar `cline` obrigatoriamente |
| Serviço | Selecione o serviço desejado (ex: `google`) | |
| Modelo | Insira o modelo desejado (ex: `gemini-2.5-flash`) | |

4. Ao clicar em **Salvar**, o token é gerado automaticamente

#### Etapa 3: Conectar ao Cline

**Método A — Aplicação automática (recomendado)**

1. Confirme que o **proxy** do wall-vault está em execução na máquina (`localhost:56244`)
2. Clique no botão **⚡ Aplicar config Cline** no card do agente no painel
3. Se aparecer a notificação "Configuração aplicada com sucesso!", foi bem-sucedido
4. Recarregue o VS Code (`Ctrl+Alt+R`)

**Método B — Configuração manual**

Abra as configurações (⚙️) na barra lateral do Cline:
- **API Provider**: `OpenAI Compatible`
- **Base URL**: `http://endereço-do-proxy:56244/v1`
  - Na mesma máquina: `http://localhost:56244/v1`
  - Em outra máquina como servidor mini: `http://192.168.0.6:56244/v1`
- **API Key**: Token emitido pelo cofre (copie do card do agente)
- **Model ID**: Modelo configurado no cofre (ex: `gemini-2.5-flash`)

#### Etapa 4: Verificação

Envie qualquer mensagem no chat do Cline. Se estiver normal:
- O card do agente correspondente no painel do cofre mostrará um **ponto verde (● Em execução)**
- O serviço/modelo atual será exibido no card (ex: `google / gemini-2.5-flash`)

#### Alterar o modelo

Quando quiser trocar o modelo do Cline, faça a alteração no **painel do cofre**:

1. Altere o dropdown de serviço/modelo no card do agente
2. Clique em **Aplicar**
3. Recarregue o VS Code (`Ctrl+Alt+R`) — o nome do modelo no rodapé do Cline será atualizado
4. O novo modelo será usado a partir da próxima requisição

> 💡 Na verdade, o proxy identifica a requisição do Cline pelo token e roteia para o modelo configurado no cofre. Mesmo sem recarregar o VS Code, **o modelo efetivamente usado muda imediatamente** — o recarregamento é apenas para atualizar a exibição do modelo na UI do Cline.

#### Detecção de desconexão

Ao fechar o VS Code, o card do agente no painel do cofre muda para amarelo (atrasado) após aproximadamente **90 segundos** e para vermelho (offline) após **3 minutos**. (A partir da v0.1.18, a verificação de status a cada 15 segundos tornou a detecção de offline mais rápida.)

#### Solução de problemas

| Sintoma | Causa | Solução |
|------|------|------|
| Erro "Falha na conexão" no Cline | Proxy não está em execução ou endereço incorreto | Verifique o proxy com `curl http://localhost:56244/health` |
| Ponto verde não aparece no cofre | Chave de API (token) não configurada | Clique novamente no botão **⚡ Aplicar config Cline** |
| Modelo no rodapé do Cline não muda | Cline armazena a configuração em cache | Recarregue o VS Code (`Ctrl+Alt+R`) |
| Nome de modelo errado é exibido | Bug antigo (corrigido na v0.1.16) | Atualize o proxy para v0.1.16 ou superior |

---

#### 🟣 Botão copiar comando de deploy — Use ao instalar em uma nova máquina

Use ao instalar o proxy do wall-vault pela primeira vez em um novo computador e conectá-lo ao cofre. Ao clicar no botão, o script de instalação completo é copiado. Cole no terminal do novo computador e execute para processar tudo de uma vez:

1. Instalação do binário wall-vault (pula se já estiver instalado)
2. Registro automático do serviço de usuário systemd
3. Iniciar o serviço e conexão automática ao cofre

> 💡 O script já contém o token deste agente e o endereço do servidor do cofre preenchidos, então pode ser executado imediatamente após colar, sem modificações adicionais.

---

### Card de serviço

Card para ativar/desativar e configurar serviços de IA.

- Interruptores de ativar/desativar por serviço
- Ao inserir o endereço de servidores de IA local (Ollama, LM Studio, vLLM etc. executando no seu computador), os modelos disponíveis são detectados automaticamente.
- **Indicador de status de conexão de serviço local**: O ponto ● ao lado do nome do serviço é **verde** quando conectado e **cinza** quando desconectado
- **Semáforo automático de serviço local** (v0.1.23+): Serviços locais (Ollama, LM Studio, vLLM) são automaticamente ativados/desativados conforme a disponibilidade de conexão. Ao ativar o serviço, ele muda para ● verde em até 15 segundos e a caixa de seleção é marcada; ao desativar, é desligado automaticamente. Funciona da mesma forma que os serviços de nuvem (Google, OpenRouter etc.) são alternados automaticamente com base na presença de chaves de API.

> 💡 **Se o serviço local estiver em execução em outro computador**: Insira o IP desse computador no campo de URL do serviço. Ex: `http://192.168.0.6:11434` (Ollama), `http://192.168.0.6:1234` (LM Studio). Se o serviço estiver vinculado apenas a `127.0.0.1` em vez de `0.0.0.0`, não será acessível pelo IP externo — verifique o endereço de binding nas configurações do serviço.

### Entrada do token de administrador

Quando você tenta usar funções importantes no painel, como adicionar/excluir chaves, um popup de entrada do token de administrador aparece. Insira o token que foi configurado no assistente de setup. Uma vez inserido, ele permanece válido até fechar o navegador.

> ⚠️ **Se houver mais de 10 tentativas de autenticação falhadas em 15 minutos, o IP será temporariamente bloqueado.** Se você esqueceu o token, verifique o item `admin_token` no arquivo `wall-vault.yaml`.

---

## Modo distribuído (múltiplos bots)

Quando você opera o OpenClaw simultaneamente em várias máquinas, esta é a configuração para **compartilhar um único cofre de chaves**. É conveniente porque o gerenciamento de chaves é feito em um só lugar.

### Exemplo de configuração

```
[Servidor do cofre de chaves]
  wall-vault vault    (cofre de chaves :56243, painel)

[WSL Alpha]           [Raspberry Pi Gamma]    [Mac Mini Local]
  wall-vault proxy      wall-vault proxy        wall-vault proxy
  openclaw TUI          openclaw TUI            openclaw TUI
  ↕ Sincronização SSE   ↕ Sincronização SSE     ↕ Sincronização SSE
```

Todos os bots apontam para o servidor do cofre central, então quando você muda o modelo ou adiciona chaves no cofre, a mudança é refletida instantaneamente em todos os bots.

### Etapa 1: Iniciar o servidor do cofre de chaves

Execute no computador que será usado como servidor do cofre:

```bash
wall-vault vault
```

### Etapa 2: Registrar cada bot (cliente)

Registre previamente as informações de cada bot que se conectará ao servidor do cofre:

```bash
curl -X POST http://localhost:56243/admin/clients \
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

### Etapa 3: Iniciar o proxy em cada computador do bot

Em cada computador onde o bot está instalado, execute o proxy especificando o endereço e token do servidor do cofre:

```bash
WV_VAULT_URL=http://192.168.x.x:56243 \
WV_VAULT_TOKEN=bota-secret \
WV_VAULT_CLIENT_ID=botA \
wall-vault proxy
```

> 💡 **`192.168.x.x`** deve ser substituído pelo endereço IP interno real do computador do servidor do cofre. Você pode verificá-lo nas configurações do roteador ou com o comando `ip addr`.

---

## Configuração de inicialização automática

Se for inconveniente iniciar manualmente o wall-vault toda vez que reiniciar o computador, registre-o como serviço do sistema. Uma vez registrado, ele inicia automaticamente na inicialização.

### Linux — systemd (maioria das distribuições Linux)

systemd é o sistema que inicia e gerencia programas automaticamente no Linux:

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

Sistema responsável pela execução automática de programas no macOS:

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
wall-vault doctor check   # Diagnosticar estado atual (apenas leitura, nada é alterado)
wall-vault doctor fix     # Corrigir problemas automaticamente
wall-vault doctor all     # Diagnóstico + correção automática de uma vez
```

> 💡 Se algo parecer estranho, execute `wall-vault doctor all` primeiro. Ele detecta e corrige muitos problemas automaticamente.

---

## Referência de variáveis de ambiente

Variáveis de ambiente são uma forma de passar valores de configuração para o programa. Insira no terminal no formato `export NOME_VARIAVEL=valor` ou coloque no arquivo de serviço de inicialização automática para aplicação permanente.

| Variável | Descrição | Valor exemplo |
|------|------|---------|
| `WV_LANG` | Idioma do painel | `ko`, `en`, `ja` |
| `WV_THEME` | Tema do painel | `light`, `dark`, `gold` |
| `WV_KEY_GOOGLE` | Chave de API do Google (múltiplas separadas por vírgula) | `AIza...,AIza...` |
| `WV_KEY_OPENROUTER` | Chave de API do OpenRouter | `sk-or-v1-...` |
| `WV_VAULT_URL` | Endereço do servidor do cofre no modo distribuído | `http://192.168.x.x:56243` |
| `WV_VAULT_TOKEN` | Token de autenticação do cliente (bot) | `my-secret-token` |
| `WV_ADMIN_TOKEN` | Token de administrador | `admin-token-here` |
| `WV_MASTER_PASS` | Senha de criptografia de chaves de API | `my-password` |
| `WV_AVATAR` | Caminho do arquivo de imagem do avatar (caminho relativo a partir de `~/.openclaw/`) | `workspace/avatars/avatar.png` |
| `OLLAMA_URL` | Endereço do servidor local Ollama | `http://192.168.x.x:11434` |

---

## Solução de problemas

### Quando o proxy não inicia

Geralmente a porta já está sendo usada por outro programa.

```bash
ss -tlnp | grep 56244   # Verificar quem está usando a porta 56244
wall-vault proxy --port 8080   # Iniciar com outro número de porta
```

### Quando ocorrem erros de chave de API (429, 402, 401, 403, 582)

| Código de erro | Significado | Ação |
|----------|------|----------|
| **429** | Muitas requisições (limite de uso excedido) | Aguarde um momento ou adicione outra chave |
| **402** | Pagamento necessário ou créditos insuficientes | Recarregue créditos no serviço correspondente |
| **401 / 403** | Chave incorreta ou sem permissão | Verifique novamente o valor da chave e re-registre |
| **582** | Sobrecarga no gateway (cooldown de 5 minutos) | Liberado automaticamente após 5 minutos |

```bash
# Verificar lista e status das chaves registradas
curl -H "Authorization: Bearer token-admin" http://localhost:56243/admin/keys

# Resetar contadores de uso das chaves
curl -X POST -H "Authorization: Bearer token-admin" http://localhost:56243/admin/keys/reset
```

### Quando o agente aparece como "não conectado"

"Não conectado" é o estado em que o processo do proxy não está enviando sinais (heartbeat) para o cofre. **Não significa que as configurações não foram salvas.** O proxy precisa estar em execução com o endereço e token do servidor do cofre para mudar para o estado conectado.

```bash
# Iniciar o proxy especificando o endereço do servidor do cofre, token e ID do cliente
WV_VAULT_URL=http://endereco-do-cofre:56243 \
WV_VAULT_TOKEN=token-do-cliente \
WV_VAULT_CLIENT_ID=id-do-cliente \
wall-vault proxy
```

Se a conexão for bem-sucedida, o status mudará para 🟢 Em execução no painel em aproximadamente 20 segundos.

### Quando a conexão com o Ollama não funciona

Ollama é um programa que executa IA diretamente no seu computador. Primeiro, verifique se o Ollama está ativo.

```bash
curl http://localhost:11434/api/tags   # Se a lista de modelos aparecer, está normal
export OLLAMA_URL=http://192.168.x.x:11434   # Se estiver em execução em outro computador
```

> ⚠️ Se o Ollama não responder, inicie-o primeiro com o comando `ollama serve`.

> ⚠️ **Modelos grandes são lentos para responder**: Modelos grandes como `qwen3.5:35b`, `deepseek-r1` podem levar vários minutos para gerar uma resposta. Mesmo que pareça que não há resposta, pode estar processando normalmente — aguarde.

---

## Alterações recentes (v0.1.16 ~ v0.1.23)

### v0.1.23 (2026-04-06)
- **Correção da mudança de modelo Ollama**: Corrigido o problema onde a mudança de modelo Ollama no painel do cofre não era refletida no proxy. Anteriormente, apenas a variável de ambiente (`OLLAMA_MODEL`) era usada, agora a configuração do cofre tem prioridade.
- **Semáforo automático de serviço local**: Ollama, LM Studio e vLLM são automaticamente ativados quando conectáveis e desativados quando desconectados. Funciona da mesma forma que o toggle automático baseado em chaves dos serviços de nuvem.

### v0.1.22 (2026-04-05)
- **Correção de campo content vazio ausente**: Quando modelos thinking (gemini-3.1-pro, o1, claude thinking etc.) usavam todo o limite de max_tokens no reasoning sem gerar resposta real, o proxy omitia os campos `content`/`text` do JSON de resposta com `omitempty`, causando crash nos clientes SDK OpenAI/Anthropic com o erro `Cannot read properties of undefined (reading 'trim')`. Alterado para sempre incluir os campos conforme a especificação oficial da API.

### v0.1.21 (2026-04-05)
- **Suporte a modelos Gemma 4**: Modelos da família Gemma como `gemma-4-31b-it`, `gemma-4-26b-a4b-it` podem ser usados via API do Google Gemini.
- **Suporte oficial a LM Studio / vLLM**: Anteriormente, esses serviços eram omitidos do roteamento do proxy e sempre substituídos pelo Ollama. Agora são roteados corretamente via API compatível com OpenAI.
- **Correção da exibição de serviço no painel**: Mesmo quando ocorre fallback, o painel sempre exibe o serviço configurado pelo usuário.
- **Indicador de status de serviço local**: Ao carregar o painel, o status de conexão dos serviços locais (Ollama, LM Studio, vLLM etc.) é exibido pela cor do ponto ●.
- **Variável de ambiente do filtro de ferramentas**: O modo de passagem de ferramentas (tools) pode ser configurado com a variável de ambiente `WV_TOOL_FILTER=passthrough`.

### v0.1.20 (2026-03-28)
- **Reforço abrangente de segurança**: 12 itens de segurança melhorados, incluindo prevenção de XSS (41 pontos), comparação de tokens em tempo constante, restrição de CORS, limites de tamanho de requisição, prevenção de travessia de caminho, autenticação SSE, reforço de limitação de taxa etc.

### v0.1.19 (2026-03-27)
- **Detecção online do Claude Code**: Claude Code que não passa pelo proxy também é exibido como online no painel.

### v0.1.18 (2026-03-26)
- **Correção de fixação de serviço de fallback**: Após fallback para Ollama devido a erro temporário, retorna automaticamente quando o serviço original se recupera.
- **Melhoria na detecção offline**: A detecção de interrupção do proxy ficou mais rápida com verificação de status a cada 15 segundos.

### v0.1.17 (2026-03-25)
- **Ordenação de cards por arrastar e soltar**: Cards de agente podem ser reordenados arrastando.
- **Botão de aplicação de configuração inline**: O botão [⚡ Aplicar configuração] é exibido em agentes offline.
- **Adicionado tipo de agente cokacdir**.

### v0.1.16 (2026-03-25)
- **Sincronização bidirecional de modelo**: Quando o modelo do Cline ou Claude Code é alterado no painel do cofre, é refletido automaticamente.

---

*Para informações mais detalhadas sobre a API, consulte [API.md](API.md).*
