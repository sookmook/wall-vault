# Manual do Usuario wall-vault
*(Last updated: 2026-04-06 — v0.1.24)*

---

## Indice

1. [O que e wall-vault?](#o-que-e-wall-vault)
2. [Instalacao](#instalacao)
3. [Primeiros passos (assistente de setup)](#primeiros-passos)
4. [Registro de chaves de API](#registro-de-chaves-de-api)
5. [Como usar o proxy](#como-usar-o-proxy)
6. [Painel do cofre de chaves](#painel-do-cofre-de-chaves)
7. [Modo distribuido (multiplos bots)](#modo-distribuido-multiplos-bots)
8. [Configuracao de inicializacao automatica](#configuracao-de-inicializacao-automatica)
9. [Doctor (diagnostico)](#doctor-diagnostico)
10. [RTK Economia de tokens](#rtk-economia-de-tokens)
11. [Referencia de variaveis de ambiente](#referencia-de-variaveis-de-ambiente)
12. [Solucao de problemas](#solucao-de-problemas)

---

## O que e wall-vault?

**wall-vault = Proxy de IA + Cofre de chaves de API para o OpenClaw**

Para usar servicos de IA, voce precisa de uma **chave de API**. Uma chave de API e como um **cracha digital** que comprova que "esta pessoa esta autorizada a usar este servico". Porem, esse cracha tem um limite de uso diario e pode ser exposto se mal gerenciado.

O wall-vault armazena esses crachas em um cofre seguro e atua como **proxy (intermediario)** entre o OpenClaw e os servicos de IA. Em resumo, o OpenClaw so precisa se conectar ao wall-vault, e o wall-vault cuida de todo o resto.

Problemas que o wall-vault resolve:

- **Rotacao automatica de chaves de API**: Quando uma chave atinge seu limite de uso ou e temporariamente bloqueada (cooldown), ele silenciosamente muda para a proxima chave. O OpenClaw continua funcionando sem interrupcao.
- **Substituicao automatica de servico (fallback)**: Se o Google nao responder, ele muda para o OpenRouter; se isso tambem falhar, muda automaticamente para IA local instalada no seu computador (Ollama, LM Studio, vLLM). A sessao nao e interrompida. Quando o servico original se recupera, ele volta automaticamente a partir da proxima requisicao (v0.1.18+, LM Studio/vLLM: v0.1.21+).
- **Sincronizacao em tempo real (SSE)**: Quando voce muda o modelo no painel do cofre, a mudanca e refletida na tela do OpenClaw em 1 a 3 segundos. SSE (Server-Sent Events) e uma tecnologia que permite ao servidor enviar atualizacoes em tempo real para o cliente.
- **Notificacoes em tempo real**: Eventos como esgotamento de chaves ou falhas de servico sao exibidos imediatamente na parte inferior do TUI (interface de terminal) do OpenClaw.

> 💡 **Claude Code, Cursor e VS Code** tambem podem ser conectados, mas o proposito original do wall-vault e ser usado com o OpenClaw.

```
OpenClaw (interface TUI no terminal)
        │
        ▼
  wall-vault proxy (:56244)   ← gerenciamento de chaves, roteamento, fallback, eventos
        │
        ├─ Google Gemini API
        ├─ OpenRouter API (mais de 340 modelos)
        ├─ Ollama / LM Studio / vLLM (seu computador, ultimo recurso)
        └─ OpenAI / Anthropic API
```

---

## Instalacao

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
- `chmod +x` — Torna o arquivo baixado "executavel". Se pular este passo, ocorrera um erro de "permissao negada".

### Windows

Abra o PowerShell (como administrador) e execute os comandos abaixo:

```powershell
# Download
Invoke-WebRequest -Uri `
  "https://github.com/sookmook/wall-vault/releases/latest/download/wall-vault-windows-amd64.exe" `
  -OutFile "$env:LOCALAPPDATA\Programs\wall-vault\wall-vault.exe"

# Adicionar ao PATH (aplica apos reiniciar o PowerShell)
$env:PATH += ";$env:LOCALAPPDATA\Programs\wall-vault"
```

> 💡 **O que e PATH?** E a lista de pastas onde o computador procura por comandos. Ao adicionar ao PATH, voce pode executar `wall-vault` de qualquer pasta simplesmente digitando o nome.

### Compilar a partir do codigo-fonte (para desenvolvedores)

Aplicavel apenas se voce tiver o ambiente de desenvolvimento Go instalado.

```bash
git clone https://github.com/sookmook/wall-vault
cd wall-vault
make build       # bin/wall-vault (versao: v0.1.24.YYYYMMDD.HHmmss)
make install     # ~/.local/bin/wall-vault
```

> 💡 **Versao com timestamp de compilacao**: Ao compilar com `make build`, a versao e gerada automaticamente em formato que inclui data e hora, como `v0.1.24.20260406.225957`. Se voce compilar diretamente com `go build ./...`, a versao sera exibida apenas como `"dev"`.

---

## Primeiros passos

### Executar o assistente de setup

Apos a instalacao, execute obrigatoriamente o **assistente de configuracao** com o comando abaixo. O assistente ira guia-lo perguntando cada item necessario.

```bash
wall-vault setup
```

As etapas que o assistente percorre sao:

```
1. Selecao de idioma (10 idiomas incluindo coreano)
2. Selecao de tema (light / dark / gold / cherry / ocean)
3. Modo de operacao — uso individual (standalone) ou compartilhado em varias maquinas (distributed)
4. Nome do bot — nome exibido no painel
5. Configuracao de portas — padrao: proxy 56244, cofre 56243 (pressione Enter se nao precisar alterar)
6. Selecao de servicos de IA — Google / OpenRouter / Ollama / LM Studio / vLLM
7. Configuracao do filtro de seguranca de ferramentas
8. Configuracao do token de administrador — senha que bloqueia funcoes de gerenciamento do painel. Geracao automatica disponivel
9. Configuracao da senha de criptografia de chaves de API — para armazenamento mais seguro (opcional)
10. Caminho para salvar o arquivo de configuracao
```

> ⚠️ **Lembre-se de guardar o token de administrador.** Ele sera necessario mais tarde para adicionar chaves ou alterar configuracoes no painel. Se voce perde-lo, tera que editar o arquivo de configuracao manualmente.

Apos concluir o assistente, o arquivo de configuracao `wall-vault.yaml` sera gerado automaticamente.

### Execucao

```bash
wall-vault start
```

Os dois servidores abaixo iniciam simultaneamente:

- **Proxy** (`http://localhost:56244`) — intermediario que conecta o OpenClaw aos servicos de IA
- **Cofre de chaves** (`http://localhost:56243`) — gerenciamento de chaves de API e painel web

Abra `http://localhost:56243` no navegador para acessar o painel imediatamente.

---

## Registro de chaves de API

Existem quatro maneiras de registrar chaves de API. **Para iniciantes, recomendamos o metodo 1 (variaveis de ambiente).**

### Metodo 1: Variaveis de ambiente (recomendado — mais simples)

Variaveis de ambiente sao **valores pre-configurados** que o programa le ao iniciar. Basta digitar no terminal como abaixo:

```bash
# Registrar chave do Google Gemini
export WV_KEY_GOOGLE=AIzaSy...

# Registrar chave do OpenRouter
export WV_KEY_OPENROUTER=sk-or-v1-...

# Executar apos o registro
wall-vault start
```

Se voce tiver varias chaves, separe-as com virgula (,). O wall-vault as usara automaticamente em rotacao (round robin):

```bash
export WV_KEY_GOOGLE=AIzaSy...,AIzaSy...,AIzaSy...
```

> 💡 **Dica**: O comando `export` se aplica apenas a sessao atual do terminal. Para que persista apos reiniciar o computador, adicione a linha acima ao arquivo `~/.bashrc` ou `~/.zshrc`.

### Metodo 2: Interface do painel (clique com o mouse)

1. Acesse `http://localhost:56243` no navegador
2. Clique no botao `[+ Adicionar]` no card **🔑 Chaves de API** no topo
3. Insira o tipo de servico, valor da chave, rotulo (nome para referencia) e limite diario, depois salve

### Metodo 3: REST API (para automacao/scripts)

REST API e uma forma de programas trocarem dados via HTTP. Util para registro automatico via script.

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

### Metodo 4: Flag do proxy (para testes rapidos)

Use para testar temporariamente inserindo uma chave sem registro formal. A chave desaparece quando o programa e encerrado.

```bash
wall-vault proxy --key-google=AIzaSy... --key-openrouter=sk-or-...
```

---

## Como usar o proxy

### Uso com OpenClaw (proposito principal)

Veja como configurar o OpenClaw para se conectar aos servicos de IA atraves do wall-vault.

Abra o arquivo `~/.openclaw/openclaw.json` e adicione o seguinte conteudo:

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

> 💡 **Maneira mais facil**: Pressione o botao **🦞 Copiar configuracao OpenClaw** no card do agente no painel e um snippet com o token e endereco ja preenchidos sera copiado para a area de transferencia. Basta colar.

**Para onde o `wall-vault/` no inicio do nome do modelo direciona?**

O wall-vault determina automaticamente para qual servico de IA enviar a requisicao com base no nome do modelo:

| Formato do modelo | Servico conectado |
|----------|--------------|
| `wall-vault/gemini-*` | Conexao direta com Google Gemini |
| `wall-vault/gpt-*`, `wall-vault/o3`, `wall-vault/o4*` | Conexao direta com OpenAI |
| `wall-vault/claude-*` | Conexao com Anthropic via OpenRouter |
| `wall-vault/hunter-alpha`, `wall-vault/healer-alpha` | OpenRouter (1 milhao de tokens de contexto gratuitos) |
| `wall-vault/kimi-*`, `wall-vault/glm-*`, `wall-vault/deepseek-*` | Conexao via OpenRouter |
| `google/nome-modelo`, `openai/nome-modelo`, `anthropic/nome-modelo` etc. | Conexao direta com o servico correspondente |
| `custom/google/nome-modelo`, `custom/openai/nome-modelo` etc. | Remove a parte `custom/` e redireciona |
| `nome-modelo:cloud` | Remove a parte `:cloud` e conecta via OpenRouter |

> 💡 **O que e contexto?** E a quantidade de conversa que a IA consegue "lembrar" de uma vez. 1M (um milhao de tokens) permite processar conversas muito longas ou documentos extensos de uma so vez.

### Conexao direta no formato da API Gemini (compatibilidade com ferramentas existentes)

Se voce ja tinha ferramentas que usavam a API do Google Gemini diretamente, basta trocar o endereco para o wall-vault:

```bash
export ANTHROPIC_BASE_URL=http://localhost:56244/google
```

Ou, para ferramentas que especificam a URL diretamente:

```
http://localhost:56244/google/v1beta/models/gemini-2.5-flash:generateContent
```

### Uso com OpenAI SDK (Python)

Voce pode conectar o wall-vault tambem em codigo Python que utiliza IA. Basta alterar o `base_url`:

```python
from openai import OpenAI

client = OpenAI(
    base_url="http://localhost:56244/v1",
    api_key="not-needed"  # As chaves de API sao gerenciadas pelo wall-vault
)

response = client.chat.completions.create(
    model="google/gemini-2.5-flash",   # Insira no formato provider/model
    messages=[{"role": "user", "content": "Ola"}]
)
```

### Trocar o modelo durante a execucao

Para trocar o modelo de IA enquanto o wall-vault ja esta em execucao:

```bash
# Alterar o modelo fazendo requisicao direta ao proxy
curl -X PUT http://localhost:56244/api/config/model \
  -H "Content-Type: application/json" \
  -d '{"service": "openrouter", "model": "anthropic/claude-3.5-sonnet"}'

# No modo distribuido (multiplos bots), altere no servidor do cofre → refletido instantaneamente via SSE
curl -X PUT http://localhost:56243/admin/clients/meu-bot-id \
  -H "Authorization: Bearer token-admin" \
  -H "Content-Type: application/json" \
  -d '{"default_service": "google", "default_model": "gemini-2.5-pro"}'
```

### Verificar lista de modelos disponiveis

```bash
# Ver lista completa
curl http://localhost:56244/api/models | python3 -m json.tool

# Ver apenas modelos do Google
curl "http://localhost:56244/api/models?service=google"

# Pesquisar por nome (ex: modelos contendo "claude")
curl "http://localhost:56244/api/models?q=claude"
```

**Resumo dos principais modelos por servico:**

| Servico | Principais modelos |
|--------|----------|
| Google | gemini-2.5-pro, gemini-2.5-flash, gemini-2.5-flash-8b, gemini-2.0-flash |
| OpenAI | gpt-4o, gpt-4o-mini, o3, o1, o1-mini |
| OpenRouter | Mais de 346 (Hunter Alpha 1M contexto gratuito, DeepSeek R1/V3, Qwen 2.5 etc.) |
| Ollama | Deteccao automatica de servidor local instalado no computador |
| LM Studio | Servidor local no computador (porta 1234) |
| vLLM | Servidor local no computador (porta 8000) |

---

## Painel do cofre de chaves

Acesse `http://localhost:56243` no navegador para ver o painel.

**Estrutura da tela:**
- **Barra superior fixa (topbar)**: Logo, seletores de idioma e tema, indicador de status da conexao SSE
- **Grade de cards**: Cards de agentes, servicos e chaves de API dispostos em formato de blocos

### Card de chaves de API

Card onde voce pode gerenciar todas as chaves de API registradas de uma so vez.

- Mostra a lista de chaves separadas por servico.
- `today_usage`: Tokens (numero de caracteres lidos e escritos pela IA) processados com sucesso hoje
- `today_attempts`: Total de chamadas hoje (incluindo sucesso + falha)
- Botao `[+ Adicionar]` para registrar novas chaves e `✕` para exclui-las.

> 💡 **O que e um token?** E a unidade usada pela IA para processar texto. Corresponde aproximadamente a uma palavra em ingles ou 1-2 caracteres em coreano. As tarifas de API geralmente sao calculadas com base neste numero de tokens.

### Card de agente

Card que mostra o status dos bots (agentes) conectados ao proxy do wall-vault.

**O status da conexao e exibido em 4 niveis:**

| Indicador | Status | Significado |
|------|------|------|
| 🟢 | Em execucao | O proxy esta funcionando normalmente |
| 🟡 | Atrasado | Responde, mas com lentidao |
| 🔴 | Offline | O proxy nao esta respondendo |
| ⚫ | Nao conectado/Desativado | O proxy nunca se conectou ao cofre ou esta desativado |

**Guia dos botoes na parte inferior do card de agente:**

Ao registrar um agente, se voce especificar o **tipo de agente**, botoes de conveniencia correspondentes ao tipo aparecem automaticamente.

---

#### 🔘 Botao copiar configuracao — Gera automaticamente a configuracao de conexao

Ao clicar no botao, um snippet de configuracao com o token, endereco do proxy e informacoes do modelo ja preenchidos e copiado para a area de transferencia. Basta colar o conteudo copiado no local indicado na tabela abaixo para completar a configuracao de conexao.

| Botao | Tipo de agente | Local para colar |
|------|-------------|-------------|
| 🦞 Copiar config OpenClaw | `openclaw` | `~/.openclaw/openclaw.json` |
| 🦀 Copiar config NanoClaw | `nanoclaw` | `~/.openclaw/openclaw.json` |
| 🟠 Copiar config Claude Code | `claude-code` | `~/.claude/settings.json` |
| ⌨ Copiar config Cursor | `cursor` | Cursor → Settings → AI |
| 💻 Copiar config VSCode | `vscode` | `~/.continue/config.json` |

**Exemplo — Se o tipo for Claude Code, o seguinte conteudo e copiado:**

```json
// ~/.claude/settings.json
{
  "apiProvider": "openai",
  "baseUrl": "http://192.168.1.20:56244/v1",
  "apiKey": "token-deste-agente"
}
```

**Exemplo — Se o tipo for VSCode (Continue):**

```yaml
# ~/.continue/config.yaml  ← Colar em config.yaml, nao config.json
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

> ⚠️ **A versao mais recente do Continue usa `config.yaml`.** Se `config.yaml` existir, `config.json` e completamente ignorado. Certifique-se de colar em `config.yaml`.

**Exemplo — Se o tipo for Cursor:**

```
Base URL : http://192.168.1.20:56244/v1
API Key  : token-deste-agente

// Ou via variaveis de ambiente:
OPENAI_BASE_URL=http://192.168.1.20:56244/v1
OPENAI_API_KEY=token-deste-agente
```

> ⚠️ **Quando a copia para a area de transferencia nao funcionar**: A politica de seguranca do navegador pode bloquear a copia. Se uma caixa de texto aparecer em popup, selecione tudo com Ctrl+A e copie com Ctrl+C.

---

#### ⚡ Botao de aplicacao automatica — Uma vez pressionado, a configuracao esta pronta

Para agentes do tipo `cline`, `claude-code`, `openclaw` ou `nanoclaw`, um botao **⚡ Aplicar configuracao** e exibido no card do agente. Ao pressionar este botao, o arquivo de configuracao local do agente e automaticamente atualizado.

| Botao | Tipo de agente | Arquivo alvo |
|------|-------------|-------------|
| ⚡ Aplicar config Cline | `cline` | `~/.cline/data/globalState.json` + `secrets.json` |
| ⚡ Aplicar config Claude Code | `claude-code` | `~/.claude/settings.json` |
| ⚡ Aplicar config OpenClaw | `openclaw` | `~/.openclaw/openclaw.json` |
| ⚡ Aplicar config NanoClaw | `nanoclaw` | `~/.openclaw/openclaw.json` |

> ⚠️ Este botao envia requisicoes para **localhost:56244** (proxy local). O proxy precisa estar em execucao nessa maquina para funcionar.

---

#### 🔀 Ordenacao de cards por arrastar e soltar (v0.1.17)

Voce pode **arrastar** os cards de agente no painel para reordena-los como desejar.

1. Clique e segure o card do agente com o mouse e arraste
2. Solte sobre o card na posicao desejada para trocar a ordem
3. A nova ordem e **salva instantaneamente no servidor** e persiste mesmo ao atualizar a pagina

> 💡 Em dispositivos touch (celular/tablet) ainda nao e suportado. Use em um navegador desktop.

---

#### 🔄 Sincronizacao bidirecional de modelo (v0.1.16)

Quando voce muda o modelo de um agente no painel do cofre, a configuracao local do agente e automaticamente atualizada.

**No caso do Cline:**
- Mudanca de modelo no cofre → evento SSE → proxy atualiza o campo de modelo no `globalState.json`
- Campos atualizados: `actModeOpenAiModelId`, `planModeOpenAiModelId`, `openAiModelId`
- `openAiBaseUrl` e chave de API nao sao alterados
- **E necessario recarregar o VS Code (`Ctrl+Alt+R` ou `Ctrl+Shift+P` → `Developer: Reload Window`)**
  - Porque o Cline nao rele o arquivo de configuracao durante a execucao

**No caso do Claude Code:**
- Mudanca de modelo no cofre → evento SSE → proxy atualiza o campo `model` no `settings.json`
- Busca automatica em ambos os caminhos WSL e Windows (`~/.claude/`, `/mnt/c/Users/*/.claude/`)

**Direcao oposta (agente → cofre):**
- Quando um agente (Cline, Claude Code etc.) envia uma requisicao ao proxy, o proxy inclui as informacoes de servico/modelo do cliente no heartbeat
- O servico/modelo atualmente em uso e exibido em tempo real no card do agente no painel do cofre

> 💡 **Ponto-chave**: O proxy identifica o agente pelo token de Authorization da requisicao e faz o roteamento automatico para o servico/modelo configurado no cofre. Mesmo que o Cline ou Claude Code envie um nome de modelo diferente, o proxy o substitui pela configuracao do cofre.

---

### Usando Cline no VS Code — Guia detalhado

#### Etapa 1: Instalar o Cline

Instale o **Cline** (ID: `saoudrizwan.claude-dev`) no marketplace de extensoes do VS Code.

#### Etapa 2: Registrar o agente no cofre

1. Abra o painel do cofre (`http://IP-do-cofre:56243`)
2. Na secao **Agentes**, clique em **+ Adicionar**
3. Preencha da seguinte forma:

| Campo | Valor | Descricao |
|------|----|------|
| ID | `meu_cline` | Identificador unico (em ingles, sem espacos) |
| Nome | `Meu Cline` | Nome exibido no painel |
| Tipo de agente | `cline` | ← Deve selecionar `cline` obrigatoriamente |
| Servico | Selecione o servico desejado (ex: `google`) | |
| Modelo | Insira o modelo desejado (ex: `gemini-2.5-flash`) | |

4. Ao clicar em **Salvar**, o token e gerado automaticamente

#### Etapa 3: Conectar ao Cline

**Metodo A — Aplicacao automatica (recomendado)**

1. Confirme que o **proxy** do wall-vault esta em execucao na maquina (`localhost:56244`)
2. Clique no botao **⚡ Aplicar config Cline** no card do agente no painel
3. Se aparecer a notificacao "Configuracao aplicada com sucesso!", foi bem-sucedido
4. Recarregue o VS Code (`Ctrl+Alt+R`)

**Metodo B — Configuracao manual**

Abra as configuracoes (⚙️) na barra lateral do Cline:
- **API Provider**: `OpenAI Compatible`
- **Base URL**: `http://endereco-do-proxy:56244/v1`
  - Na mesma maquina: `http://localhost:56244/v1`
  - Em outra maquina como servidor mini: `http://192.168.1.20:56244/v1`
- **API Key**: Token emitido pelo cofre (copie do card do agente)
- **Model ID**: Modelo configurado no cofre (ex: `gemini-2.5-flash`)

#### Etapa 4: Verificacao

Envie qualquer mensagem no chat do Cline. Se estiver normal:
- O card do agente correspondente no painel do cofre mostrara um **ponto verde (● Em execucao)**
- O servico/modelo atual sera exibido no card (ex: `google / gemini-2.5-flash`)

#### Alterar o modelo

Quando quiser trocar o modelo do Cline, faca a alteracao no **painel do cofre**:

1. Altere o dropdown de servico/modelo no card do agente
2. Clique em **Aplicar**
3. Recarregue o VS Code (`Ctrl+Alt+R`) — o nome do modelo no rodape do Cline sera atualizado
4. O novo modelo sera usado a partir da proxima requisicao

> 💡 Na verdade, o proxy identifica a requisicao do Cline pelo token e roteia para o modelo configurado no cofre. Mesmo sem recarregar o VS Code, **o modelo efetivamente usado muda imediatamente** — o recarregamento e apenas para atualizar a exibicao do modelo na UI do Cline.

#### Deteccao de desconexao

Ao fechar o VS Code, o card do agente no painel do cofre muda para amarelo (atrasado) apos aproximadamente **90 segundos** e para vermelho (offline) apos **3 minutos**. (A partir da v0.1.18, a verificacao de status a cada 15 segundos tornou a deteccao de offline mais rapida.)

#### Solucao de problemas

| Sintoma | Causa | Solucao |
|------|------|------|
| Erro "Falha na conexao" no Cline | Proxy nao esta em execucao ou endereco incorreto | Verifique o proxy com `curl http://localhost:56244/health` |
| Ponto verde nao aparece no cofre | Chave de API (token) nao configurada | Clique novamente no botao **⚡ Aplicar config Cline** |
| Modelo no rodape do Cline nao muda | Cline armazena a configuracao em cache | Recarregue o VS Code (`Ctrl+Alt+R`) |
| Nome de modelo errado e exibido | Bug antigo (corrigido na v0.1.16) | Atualize o proxy para v0.1.16 ou superior |

---

#### 🟣 Botao copiar comando de deploy — Use ao instalar em uma nova maquina

Use ao instalar o proxy do wall-vault pela primeira vez em um novo computador e conecta-lo ao cofre. Ao clicar no botao, o script de instalacao completo e copiado. Cole no terminal do novo computador e execute para processar tudo de uma vez:

1. Instalacao do binario wall-vault (pula se ja estiver instalado)
2. Registro automatico do servico de usuario systemd
3. Iniciar o servico e conexao automatica ao cofre

> 💡 O script ja contem o token deste agente e o endereco do servidor do cofre preenchidos, entao pode ser executado imediatamente apos colar, sem modificacoes adicionais.

---

### Card de servico

Card para ativar/desativar e configurar servicos de IA.

- Interruptores de ativar/desativar por servico
- Ao inserir o endereco de servidores de IA local (Ollama, LM Studio, vLLM etc. executando no seu computador), os modelos disponiveis sao detectados automaticamente.
- **Indicador de status de conexao de servico local**: O ponto ● ao lado do nome do servico e **verde** quando conectado e **cinza** quando desconectado
- **Semaforo automatico de servico local** (v0.1.23+): Servicos locais (Ollama, LM Studio, vLLM) sao automaticamente ativados/desativados conforme a disponibilidade de conexao. Ao ativar o servico, ele muda para ● verde em ate 15 segundos e a caixa de selecao e marcada; ao desativar, e desligado automaticamente. Funciona da mesma forma que os servicos de nuvem (Google, OpenRouter etc.) sao alternados automaticamente com base na presenca de chaves de API.

> 💡 **Se o servico local estiver em execucao em outro computador**: Insira o IP desse computador no campo de URL do servico. Ex: `http://192.168.1.20:11434` (Ollama), `http://192.168.1.20:1234` (LM Studio). Se o servico estiver vinculado apenas a `127.0.0.1` em vez de `0.0.0.0`, nao sera acessivel pelo IP externo — verifique o endereco de binding nas configuracoes do servico.

### Entrada do token de administrador

Quando voce tenta usar funcoes importantes no painel, como adicionar/excluir chaves, um popup de entrada do token de administrador aparece. Insira o token que foi configurado no assistente de setup. Uma vez inserido, ele permanece valido ate fechar o navegador.

> ⚠️ **Se houver mais de 10 tentativas de autenticacao falhadas em 15 minutos, o IP sera temporariamente bloqueado.** Se voce esqueceu o token, verifique o item `admin_token` no arquivo `wall-vault.yaml`.

---

## Modo distribuido (multiplos bots)

Quando voce opera o OpenClaw simultaneamente em varias maquinas, esta e a configuracao para **compartilhar um unico cofre de chaves**. E conveniente porque o gerenciamento de chaves e feito em um so lugar.

### Exemplo de configuracao

```
[Servidor do cofre de chaves]
  wall-vault vault    (cofre de chaves :56243, painel)

[WSL Alpha]           [Raspberry Pi Gamma]    [Mac Mini Local]
  wall-vault proxy      wall-vault proxy        wall-vault proxy
  openclaw TUI          openclaw TUI            openclaw TUI
  ↕ Sincronizacao SSE   ↕ Sincronizacao SSE     ↕ Sincronizacao SSE
```

Todos os bots apontam para o servidor do cofre central, entao quando voce muda o modelo ou adiciona chaves no cofre, a mudanca e refletida instantaneamente em todos os bots.

### Etapa 1: Iniciar o servidor do cofre de chaves

Execute no computador que sera usado como servidor do cofre:

```bash
wall-vault vault
```

### Etapa 2: Registrar cada bot (cliente)

Registre previamente as informacoes de cada bot que se conectara ao servidor do cofre:

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

Em cada computador onde o bot esta instalado, execute o proxy especificando o endereco e token do servidor do cofre:

```bash
WV_VAULT_URL=http://192.168.x.x:56243 \
WV_VAULT_TOKEN=bota-secret \
WV_VAULT_CLIENT_ID=botA \
wall-vault proxy
```

> 💡 **`192.168.x.x`** deve ser substituido pelo endereco IP interno real do computador do servidor do cofre. Voce pode verifica-lo nas configuracoes do roteador ou com o comando `ip addr`.

---

## Configuracao de inicializacao automatica

Se for inconveniente iniciar manualmente o wall-vault toda vez que reiniciar o computador, registre-o como servico do sistema. Uma vez registrado, ele inicia automaticamente na inicializacao.

### Linux — systemd (maioria das distribuicoes Linux)

systemd e o sistema que inicia e gerencia programas automaticamente no Linux:

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

Sistema responsavel pela execucao automatica de programas no macOS:

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

## Doctor (diagnostico)

O comando `doctor` e uma ferramenta que **diagnostica e corrige automaticamente** se o wall-vault esta configurado corretamente.

```bash
wall-vault doctor check   # Diagnosticar estado atual (apenas leitura, nada e alterado)
wall-vault doctor fix     # Corrigir problemas automaticamente
wall-vault doctor all     # Diagnostico + correcao automatica de uma vez
```

> 💡 Se algo parecer estranho, execute `wall-vault doctor all` primeiro. Ele detecta e corrige muitos problemas automaticamente.

---

## RTK Economia de tokens

*(v0.1.24+)*

**RTK (ferramenta de economia de tokens)** comprime automaticamente a saida de comandos shell executados por agentes de codificacao de IA (como Claude Code), reduzindo o consumo de tokens. Por exemplo, a saida de 15 linhas de `git status` e reduzida a um resumo de 2 linhas.

### Uso basico

```bash
# Envolva o comando com wall-vault rtk e a saida sera filtrada automaticamente
wall-vault rtk git status          # Mostra apenas a lista de arquivos alterados
wall-vault rtk git diff HEAD~1     # Apenas linhas alteradas + contexto minimo
wall-vault rtk git log -10         # Hash + mensagem em uma linha cada
wall-vault rtk go test ./...       # Mostra apenas testes que falharam
wall-vault rtk ls -la              # Comandos nao suportados sao truncados automaticamente
```

### Comandos suportados e economia

| Comando | Metodo de filtro | Economia |
|------|----------|--------|
| `git status` | Apenas resumo de arquivos alterados | ~87% |
| `git diff` | Linhas alteradas + 3 linhas de contexto | ~60-94% |
| `git log` | Hash + primeira linha da mensagem | ~90% |
| `git push/pull/fetch` | Remove progresso, apenas resumo | ~80% |
| `go test` | Mostra apenas falhas, conta aprovados | ~88-99% |
| `go build/vet` | Mostra apenas erros | ~90% |
| Todos os outros comandos | Primeiras 50 linhas + ultimas 50 linhas, maximo 32KB | Variavel |

### Pipeline de filtro em 3 etapas

1. **Filtro estrutural por comando** — Entende o formato de saida de git, go etc. e extrai apenas partes significativas
2. **Pos-processamento com regex** — Remove codigos de cor ANSI, reduz linhas vazias, agrega linhas duplicadas
3. **Passthrough + truncamento** — Comandos nao suportados mantem apenas as primeiras/ultimas 50 linhas

### Integracao com Claude Code

Voce pode configurar um hook `PreToolUse` do Claude Code para que todos os comandos shell passem automaticamente pelo RTK.

```bash
# Instalar hook (adicionado automaticamente ao settings.json do Claude Code)
wall-vault rtk hook install
```

Ou adicionar manualmente a `~/.claude/settings.json`:

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

> 💡 **Preservacao do exit code**: O RTK retorna o codigo de saida original do comando. Se o comando falhar (exit code ≠ 0), a IA tambem detecta a falha corretamente.

> 💡 **Saida forcada em ingles**: O RTK executa os comandos com `LC_ALL=C` para sempre gerar saida em ingles, independentemente das configuracoes de idioma do sistema. Isso e necessario para que os filtros funcionem corretamente.

---

## Referencia de variaveis de ambiente

Variaveis de ambiente sao uma forma de passar valores de configuracao para o programa. Insira no terminal no formato `export NOME_VARIAVEL=valor` ou coloque no arquivo de servico de inicializacao automatica para aplicacao permanente.

| Variavel | Descricao | Valor exemplo |
|------|------|---------|
| `WV_LANG` | Idioma do painel | `ko`, `en`, `ja` |
| `WV_THEME` | Tema do painel | `light`, `dark`, `gold` |
| `WV_KEY_GOOGLE` | Chave de API do Google (multiplas separadas por virgula) | `AIza...,AIza...` |
| `WV_KEY_OPENROUTER` | Chave de API do OpenRouter | `sk-or-v1-...` |
| `WV_VAULT_URL` | Endereco do servidor do cofre no modo distribuido | `http://192.168.x.x:56243` |
| `WV_VAULT_TOKEN` | Token de autenticacao do cliente (bot) | `my-secret-token` |
| `WV_ADMIN_TOKEN` | Token de administrador | `admin-token-here` |
| `WV_MASTER_PASS` | Senha de criptografia de chaves de API | `my-password` |
| `WV_AVATAR` | Caminho do arquivo de imagem do avatar (caminho relativo a partir de `~/.openclaw/`) | `workspace/avatars/avatar.png` |
| `OLLAMA_URL` | Endereco do servidor local Ollama | `http://192.168.x.x:11434` |

---

## Solucao de problemas

### Quando o proxy nao inicia

Geralmente a porta ja esta sendo usada por outro programa.

```bash
ss -tlnp | grep 56244   # Verificar quem esta usando a porta 56244
wall-vault proxy --port 8080   # Iniciar com outro numero de porta
```

### Quando ocorrem erros de chave de API (429, 402, 401, 403, 582)

| Codigo de erro | Significado | Acao |
|----------|------|----------|
| **429** | Muitas requisicoes (limite de uso excedido) | Aguarde um momento ou adicione outra chave |
| **402** | Pagamento necessario ou creditos insuficientes | Recarregue creditos no servico correspondente |
| **401 / 403** | Chave incorreta ou sem permissao | Verifique novamente o valor da chave e re-registre |
| **582** | Sobrecarga no gateway (cooldown de 5 minutos) | Liberado automaticamente apos 5 minutos |

```bash
# Verificar lista e status das chaves registradas
curl -H "Authorization: Bearer token-admin" http://localhost:56243/admin/keys

# Resetar contadores de uso das chaves
curl -X POST -H "Authorization: Bearer token-admin" http://localhost:56243/admin/keys/reset
```

### Quando o agente aparece como "nao conectado"

"Nao conectado" e o estado em que o processo do proxy nao esta enviando sinais (heartbeat) para o cofre. **Nao significa que as configuracoes nao foram salvas.** O proxy precisa estar em execucao com o endereco e token do servidor do cofre para mudar para o estado conectado.

```bash
# Iniciar o proxy especificando o endereco do servidor do cofre, token e ID do cliente
WV_VAULT_URL=http://endereco-do-cofre:56243 \
WV_VAULT_TOKEN=token-do-cliente \
WV_VAULT_CLIENT_ID=id-do-cliente \
wall-vault proxy
```

Se a conexao for bem-sucedida, o status mudara para 🟢 Em execucao no painel em aproximadamente 20 segundos.

### Quando a conexao com o Ollama nao funciona

Ollama e um programa que executa IA diretamente no seu computador. Primeiro, verifique se o Ollama esta ativo.

```bash
curl http://localhost:11434/api/tags   # Se a lista de modelos aparecer, esta normal
export OLLAMA_URL=http://192.168.x.x:11434   # Se estiver em execucao em outro computador
```

> ⚠️ Se o Ollama nao responder, inicie-o primeiro com o comando `ollama serve`.

> ⚠️ **Modelos grandes sao lentos para responder**: Modelos grandes como `qwen3.5:35b`, `deepseek-r1` podem levar varios minutos para gerar uma resposta. Mesmo que pareca que nao ha resposta, pode estar processando normalmente — aguarde.

---

## Alteracoes recentes (v0.1.16 ~ v0.1.24)

### v0.1.24 (2026-04-06)
- **Subcomando RTK de economia de tokens**: `wall-vault rtk <command>` filtra automaticamente a saida de comandos shell para reduzir o consumo de tokens de agentes de IA em 60-90%. Inclui filtros dedicados para comandos principais como git e go, e trunca automaticamente comandos nao suportados. Integra-se de forma transparente com o hook `PreToolUse` do Claude Code.

### v0.1.23 (2026-04-06)
- **Correcao da mudanca de modelo Ollama**: Corrigido o problema onde a mudanca de modelo Ollama no painel do cofre nao era refletida no proxy. Anteriormente, apenas a variavel de ambiente (`OLLAMA_MODEL`) era usada, agora a configuracao do cofre tem prioridade.
- **Semaforo automatico de servico local**: Ollama, LM Studio e vLLM sao automaticamente ativados quando conectaveis e desativados quando desconectados. Funciona da mesma forma que o toggle automatico baseado em chaves dos servicos de nuvem.

### v0.1.22 (2026-04-05)
- **Correcao de campo content vazio ausente**: Quando modelos thinking (gemini-3.1-pro, o1, claude thinking etc.) usavam todo o limite de max_tokens no reasoning sem gerar resposta real, o proxy omitia os campos `content`/`text` do JSON de resposta com `omitempty`, causando crash nos clientes SDK OpenAI/Anthropic com o erro `Cannot read properties of undefined (reading 'trim')`. Alterado para sempre incluir os campos conforme a especificacao oficial da API.

### v0.1.21 (2026-04-05)
- **Suporte a modelos Gemma 4**: Modelos da familia Gemma como `gemma-4-31b-it`, `gemma-4-26b-a4b-it` podem ser usados via API do Google Gemini.
- **Suporte oficial a LM Studio / vLLM**: Anteriormente, esses servicos eram omitidos do roteamento do proxy e sempre substituidos pelo Ollama. Agora sao roteados corretamente via API compativel com OpenAI.
- **Correcao da exibicao de servico no painel**: Mesmo quando ocorre fallback, o painel sempre exibe o servico configurado pelo usuario.
- **Indicador de status de servico local**: Ao carregar o painel, o status de conexao dos servicos locais (Ollama, LM Studio, vLLM etc.) e exibido pela cor do ponto ●.
- **Variavel de ambiente do filtro de ferramentas**: O modo de passagem de ferramentas (tools) pode ser configurado com a variavel de ambiente `WV_TOOL_FILTER=passthrough`.

### v0.1.20 (2026-03-28)
- **Reforco abrangente de seguranca**: 12 itens de seguranca melhorados, incluindo prevencao de XSS (41 pontos), comparacao de tokens em tempo constante, restricao de CORS, limites de tamanho de requisicao, prevencao de travessia de caminho, autenticacao SSE, reforco de limitacao de taxa etc.

### v0.1.19 (2026-03-27)
- **Deteccao online do Claude Code**: Claude Code que nao passa pelo proxy tambem e exibido como online no painel.

### v0.1.18 (2026-03-26)
- **Correcao de fixacao de servico de fallback**: Apos fallback para Ollama devido a erro temporario, retorna automaticamente quando o servico original se recupera.
- **Melhoria na deteccao offline**: A deteccao de interrupcao do proxy ficou mais rapida com verificacao de status a cada 15 segundos.

### v0.1.17 (2026-03-25)
- **Ordenacao de cards por arrastar e soltar**: Cards de agente podem ser reordenados arrastando.
- **Botao de aplicacao de configuracao inline**: O botao [⚡ Aplicar configuracao] e exibido em agentes offline.
- **Adicionado tipo de agente cokacdir**.

### v0.1.16 (2026-03-25)
- **Sincronizacao bidirecional de modelo**: Quando o modelo do Cline ou Claude Code e alterado no painel do cofre, e refletido automaticamente.

---

*Para informacoes mais detalhadas sobre a API, consulte [API.md](API.md).*
