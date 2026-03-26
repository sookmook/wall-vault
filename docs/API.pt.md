# Manual de API do wall-vault

Este documento descreve em detalhes todos os endpoints HTTP API do wall-vault.

---

## Sumário

- [Autenticação](#autenticação)
- [API do Proxy (:56244)](#api-do-proxy-56244)
  - [Health Check](#get-health)
  - [Consulta de Status](#get-status)
  - [Lista de Modelos](#get-apimodels)
  - [Alteração de Modelo](#put-apiconfigmodel)
  - [Modo de Pensamento](#put-apiconfigthink-mode)
  - [Atualização de Configurações](#post-reload)
  - [API Gemini](#post-googlev1betamodelsmgeneratecontent)
  - [Streaming Gemini](#post-googlev1betamodelsmstreamgeneratecontent)
  - [API Compatível com OpenAI](#post-v1chatcompletions)
- [API do Key Vault (:56243)](#api-do-key-vault-56243)
  - [API Pública](#api-pública-sem-autenticação)
  - [Stream de Eventos SSE](#get-apievents)
  - [API Exclusiva do Proxy](#api-exclusiva-do-proxy-token-do-cliente)
  - [API Admin — Chaves](#api-admin--chaves-api)
  - [API Admin — Clientes](#api-admin--clientes)
  - [API Admin — Serviços](#api-admin--serviços)
  - [API Admin — Lista de Modelos](#api-admin--lista-de-modelos)
  - [API Admin — Status do Proxy](#api-admin--status-do-proxy)
- [Tipos de Eventos SSE](#tipos-de-eventos-sse)
- [Roteamento de Provedor e Modelo](#roteamento-de-provedor-e-modelo)
- [Esquema de Dados](#esquema-de-dados)
- [Respostas de Erro](#respostas-de-erro)
- [Coleção de Exemplos cURL](#coleção-de-exemplos-curl)

---

## Autenticação

| Escopo | Método | Header |
|--------|--------|--------|
| API Admin | Token Bearer | `Authorization: Bearer <admin_token>` |
| Proxy → Vault | Token Bearer | `Authorization: Bearer <client_token>` |
| API do Proxy | Nenhum (local) | — |

Se o `admin_token` não estiver definido (string vazia), todas as APIs admin ficam acessíveis sem autenticação.

### Política de Segurança

- **Rate Limiting**: Quando as falhas de autenticação da API admin excedem 10 vezes em 15 minutos, o IP é bloqueado temporariamente (`429 Too Many Requests`)
- **Whitelist de IP**: Apenas IPs/CIDRs registrados no campo `ip_whitelist` do agente (`Client`) podem acessar `/api/keys`. Array vazio permite todos.
- **Proteção theme·lang**: `/admin/theme`, `/admin/lang` também requerem autenticação por token admin

---

## API do Proxy (:56244)

Servidor onde o proxy é executado. Porta padrão `56244`.

---

### `GET /health`

Health check. Sempre retorna 200 OK.

**Exemplo de resposta:**
```json
{
  "status": "ok",
  "version": "v0.1.6.20260314.231308",
  "client": "bot-a"
}
```

---

### `GET /status`

Consulta detalhada do status do proxy.

**Exemplo de resposta:**
```json
{
  "status": "ok",
  "version": "v0.1.6.20260314.231308",
  "client": "bot-a",
  "service": "google",
  "model": "gemini-2.5-flash",
  "sse": true,
  "filter": "strip_all",
  "services": ["google", "openrouter", "ollama"],
  "mode": "distributed"
}
```

| Campo | Tipo | Descrição |
|-------|------|-----------|
| `service` | string | Serviço padrão atual |
| `model` | string | Modelo padrão atual |
| `sse` | bool | Status da conexão SSE com o vault |
| `filter` | string | Modo de filtro de ferramentas |
| `services` | []string | Lista de serviços ativos |
| `mode` | string | `standalone` \| `distributed` |

---

### `GET /api/models`

Consulta da lista de modelos disponíveis. Utiliza cache TTL (padrão 10 minutos).

**Parâmetros de consulta:**

| Parâmetro | Descrição | Exemplo |
|-----------|-----------|---------|
| `service` | Filtro de serviço | `?service=google` |
| `q` | Pesquisa por ID/nome do modelo | `?q=gemini` |

**Exemplo de resposta:**
```json
{
  "models": [
    {
      "id": "gemini-2.5-pro",
      "name": "Gemini 2.5 Pro",
      "service": "google",
      "context_length": 1048576,
      "free": false
    },
    {
      "id": "openrouter/hunter-alpha",
      "name": "Hunter Alpha (1M ctx, free)",
      "service": "openrouter",
      "context_length": 1048576,
      "free": true
    }
  ],
  "count": 2
}
```

| Campo | Tipo | Descrição |
|-------|------|-----------|
| `id` | string | ID do modelo |
| `name` | string | Nome de exibição do modelo |
| `service` | string | `google` \| `openai` \| `anthropic` \| `openrouter` \| `ollama` etc. |
| `context_length` | int | Tamanho da janela de contexto |
| `free` | bool | Se o modelo é gratuito (OpenRouter) |

---

### `PUT /api/config/model`

Alteração do serviço e modelo atuais.

**Corpo da requisição:**
```json
{
  "service": "google",
  "model": "gemini-2.5-flash"
}
```

**Resposta:**
```json
{
  "status": "ok",
  "service": "google",
  "model": "gemini-2.5-flash"
}
```

> **Nota:** No modo distribuído, é recomendado usar `PUT /admin/clients/{id}` do vault ao invés desta API. As alterações do vault são refletidas automaticamente em 1-3 segundos via SSE.

---

### `PUT /api/config/think-mode`

Alternância do modo de pensamento (no-op, para expansão futura).

**Resposta:**
```json
{"status": "ok"}
```

---

### `POST /reload`

Ressincronização instantânea das configurações do cliente e chaves a partir do vault.

**Resposta:**
```json
{"status": "reloading"}
```

A ressincronização é executada de forma assíncrona e é concluída em 1-2 segundos após o recebimento da resposta.

---

### `POST /google/v1beta/models/{model}:generateContent`

Proxy da API Gemini (sem streaming).

**Parâmetro de path:**
- `{model}`: ID do modelo. Se tiver o prefixo `gemini-`, o serviço Google é selecionado automaticamente.

**Corpo da requisição:** [Formato de requisição Gemini generateContent](https://ai.google.dev/api/generate-content)

```json
{
  "contents": [
    {
      "role": "user",
      "parts": [{"text": "안녕하세요"}]
    }
  ],
  "generationConfig": {
    "temperature": 0.7,
    "maxOutputTokens": 1024
  }
}
```

**Corpo da resposta:** Formato de resposta Gemini generateContent

**Filtro de ferramentas:** Com a configuração `tool_filter: strip_all`, o array `tools` da requisição é removido automaticamente.

**Cadeia de fallback:** Falha do serviço especificado → fallback na ordem dos serviços configurados → Ollama (último recurso).

---

### `POST /google/v1beta/models/{model}:streamGenerateContent`

Proxy de streaming da API Gemini. O formato de requisição é idêntico ao modo sem streaming. A resposta é um stream SSE:

```
data: {"candidates":[{"content":{"parts":[{"text":"안"}],...},...}]}

data: [DONE]
```

---

### `POST /v1/chat/completions`

API compatível com OpenAI. Internamente é convertida para o formato Gemini e processada.

**Corpo da requisição:**
```json
{
  "model": "gemini-2.5-flash",
  "messages": [
    {"role": "system", "content": "당신은 도움이 되는 어시스턴트입니다."},
    {"role": "user", "content": "안녕하세요"}
  ],
  "temperature": 0.7,
  "max_tokens": 1024,
  "stream": false
}
```

**Suporte a prefixo de provedor no campo `model` (OpenClaw 3.11+):**

| Exemplo de modelo | Roteamento |
|-------------------|------------|
| `gemini-2.5-flash` | Serviço configurado atualmente |
| `google/gemini-2.5-pro` | Direto para Google |
| `openai/gpt-4o` | Direto para OpenAI |
| `anthropic/claude-opus-4-6` | Via OpenRouter |
| `openrouter/meta-llama/llama-3.3-70b` | Direto para OpenRouter |
| `wall-vault/gemini-2.5-flash` | Detecção automática → Google |
| `wall-vault/claude-opus-4-6` | Detecção automática → OpenRouter Anthropic |
| `wall-vault/gpt-4o` | Detecção automática → OpenAI |
| `wall-vault/hunter-alpha` | OpenRouter (gratuito 1M context) |
| `moonshot/kimi-k2.5` | Via OpenRouter |
| `opencode-go/model` | Via OpenRouter |
| `kimi-k2.5:cloud` | Sufixo `:cloud` → OpenRouter |

Para mais detalhes consulte [Roteamento de Provedor e Modelo](#roteamento-de-provedor-e-modelo).

**Corpo da resposta:**
```json
{
  "choices": [
    {
      "message": {
        "role": "assistant",
        "content": "안녕하세요! 무엇을 도와드릴까요?"
      },
      "finish_reason": "stop",
      "index": 0
    }
  ]
}
```

> **Remoção automática de tokens de controle do modelo:** Se a resposta contiver delimitadores GLM-5 / DeepSeek / ChatML (`<|im_start|>`, `[gMASK]`, `[sop]` etc.), eles são removidos automaticamente.

---

## API do Key Vault (:56243)

Servidor onde o key vault é executado. Porta padrão `56243`.

---

### API Pública (Sem Autenticação)

#### `GET /`

Interface web do dashboard. Acessada pelo navegador.

---

#### `GET /api/status`

Consulta do status do vault.

**Exemplo de resposta:**
```json
{
  "status": "ok",
  "version": "v0.1.6.20260314.231308",
  "keys": 3,
  "clients": 2,
  "sse": 2
}
```

---

#### `GET /api/clients`

Lista de clientes registrados (apenas informações públicas, sem tokens).

---

### `GET /api/events`

Stream de eventos SSE (Server-Sent Events) em tempo real.

**Headers:**
```
Content-Type: text/event-stream
Cache-Control: no-cache
Connection: keep-alive
```

**Recebido imediatamente após a conexão:**
```
data: {"type":"connected","clients":2}
```

**Exemplos de eventos:**
```
data: {"type":"config_change","data":{"client_id":"bot-a","service":"google","model":"gemini-2.5-pro"}}

data: {"type":"key_added","data":{"service":"openrouter"}}

data: {"type":"key_deleted","data":{"service":"google"}}

data: {"type":"service_changed","data":{"action":"updated","id":"ollama"}}

data: {"type":"usage_reset","data":{"time":"2026-03-13T00:00:30Z"}}
```

Para detalhes dos tipos de eventos consulte [Tipos de Eventos SSE](#tipos-de-eventos-sse).

---

### API Exclusiva do Proxy (Token do Cliente)

Requer header `Authorization: Bearer <client_token>`. Autenticação com token admin também é possível.

#### `GET /api/keys`

Lista de chaves API descriptografadas fornecidas ao proxy.

**Parâmetros de consulta:**

| Parâmetro | Descrição |
|-----------|-----------|
| `service` | Filtro de serviço (exemplo: `?service=google`) |

**Exemplo de resposta:**
```json
[
  {
    "id": "key-abc123",
    "service": "google",
    "plain_key": "AIzaSy...",
    "daily_limit": 1000,
    "today_usage": 42,
    "today_attempts": 45
  }
]
```

> **Segurança:** Retorna chaves em texto puro. Apenas chaves de serviços permitidos são retornadas conforme a configuração `allowed_services` do cliente.

---

#### `GET /api/services`

Consulta da lista de serviços utilizados pelo proxy. Retorna array de IDs de serviços com `proxy_enabled=true`.

**Exemplo de resposta:**
```json
["google", "ollama"]
```

Se o array estiver vazio, o proxy usa todos os serviços sem restrições.

---

#### `POST /api/heartbeat`

Envio do status do proxy (executado automaticamente a cada 20 segundos).

**Corpo da requisição:**
```json
{
  "client_id": "bot-a",
  "version": "v0.1.6.20260314.231308",
  "service": "google",
  "model": "gemini-2.5-flash",
  "sse_connected": true,
  "host": "bot-a-host",
  "avatar": "data:image/png;base64,...",
  "key_usage":     {"key-abc123": 42, "key-def456": 0},
  "key_attempts":  {"key-abc123": 45, "key-def456": 3},
  "key_cooldowns": {"key-abc123": "2026-03-15T14:30:00Z"}
}
```

| Campo | Tipo | Descrição |
|-------|------|-----------|
| `client_id` | string | ID do cliente |
| `version` | string | Versão do proxy (inclui timestamp de build, ex: `v0.1.6.20260314.231308`) |
| `service` | string | Serviço atual |
| `model` | string | Modelo atual |
| `sse_connected` | bool | Status da conexão SSE |
| `host` | string | Hostname |
| `avatar` | string | base64 data URI of the proxy's local avatar image (sent by proxy, auto-persisted to client record by vault). Set via `WV_AVATAR` env var (relative path under `~/.openclaw/`). |
| `key_usage` | map[string]int | Key ID → successful tokens used today. Vault calls `SetKeyUsage()` (absolute value sync). 429/402/582 errors do **not** increment this. |
| `key_attempts` | map[string]int | Key ID → total API calls today (success + rate-limited). Vault calls `SetKeyAttempts()`. 429/402/582 errors only increment this, not `key_usage`. |
| `key_cooldowns` | map[string]string | Key ID → cooldown end time (RFC3339). Vault calls `SetKeyCooldownIfLater()`. |

**Resposta:**
```json
{"status": "ok"}
```

---

### API Admin — Chaves API

Requer header `Authorization: Bearer <admin_token>`.

#### `GET /admin/keys`

Lista de todas as chaves API registradas (sem chaves em texto puro).

**Exemplo de resposta:**
```json
[
  {
    "id": "key-abc123",
    "service": "google",
    "label": "메인 키",
    "today_usage": 42,
    "today_attempts": 45,
    "daily_limit": 1000,
    "cooldown_until": "0001-01-01T00:00:00Z",
    "last_error": 0,
    "created_at": "2026-03-13T12:00:00Z",
    "available": true,
    "usage_pct": 4
  }
]
```

| Campo | Tipo | Descrição |
|-------|------|-----------|
| `today_usage` | int | Número de tokens de requisições bem-sucedidas hoje (não inclui erros 429/402/582) |
| `today_attempts` | int | Total de chamadas API hoje (bem-sucedidas + rate-limited) |
| `available` | bool | Se está disponível para uso sem cooldown ou limite |
| `usage_pct` | int | Porcentagem de uso em relação ao limite diário % (`daily_limit=0` significa 0) |
| `cooldown_until` | RFC3339 | Hora de término do cooldown (valor zero significa nenhum) |
| `last_error` | int | Último código de erro HTTP |

---

#### `POST /admin/keys`

Registro de nova chave API. O evento SSE `key_added` é transmitido imediatamente após o registro.

**Corpo da requisição:**
```json
{
  "service": "google",
  "key": "AIzaSy...",
  "label": "메인 키",
  "daily_limit": 1000
}
```

| Campo | Obrigatório | Descrição |
|-------|-------------|-----------|
| `service` | ✅ | `google` \| `openai` \| `anthropic` \| `openrouter` \| `ollama` \| personalizado |
| `key` | ✅ | Chave API em texto puro |
| `label` | — | Rótulo de identificação |
| `daily_limit` | — | Limite de uso diário (0 = ilimitado) |

---

#### `DELETE /admin/keys/{id}`

Exclusão de chave API. O evento SSE `key_deleted` é transmitido após a exclusão.

**Resposta:**
```json
{"status": "deleted"}
```

---

#### `POST /admin/keys/reset`

Reset do uso diário de todas as chaves. Transmissão do evento SSE `usage_reset`.

**Resposta:**
```json
{
  "status": "reset",
  "time": "2026-03-13T15:00:00Z"
}
```

---

### API Admin — Clientes

#### `GET /admin/clients`

Lista de todos os clientes (incluindo tokens).

---

#### `POST /admin/clients`

Registro de novo cliente.

**Corpo da requisição:**
```json
{
  "id": "my-bot",
  "name": "내 봇",
  "token": "my-secret-token",
  "default_service": "google",
  "default_model": "gemini-2.5-flash",
  "allowed_services": ["google", "openrouter"],
  "agent_type": "openclaw",
  "work_dir": "~/.openclaw",
  "description": "OpenClaw 에이전트",
  "ip_whitelist": ["10.0.0.1", "10.0.0.0/24"],
  "enabled": true
}
```

| Campo | Obrigatório | Descrição |
|-------|-------------|-----------|
| `id` | ✅ | ID único do cliente |
| `name` | — | Nome de exibição |
| `token` | — | Token de autenticação (gerado automaticamente se omitido) |
| `default_service` | — | Serviço padrão |
| `default_model` | — | Modelo padrão |
| `allowed_services` | — | Lista de serviços permitidos (array vazio = permitir todos) |
| `agent_type` | — | `openclaw` \| `claude-code` \| `cursor` \| `vscode` \| `custom` |
| `work_dir` | — | Diretório de trabalho do agente |
| `description` | — | Descrição do agente |
| `ip_whitelist` | — | Lista de IPs permitidos (array vazio = permitir todos, suporte a CIDR) |
| `enabled` | — | Status de ativação (padrão `true`) |

---

#### `GET /admin/clients/{id}`

Consulta de cliente específico (incluindo token).

---

#### `PUT /admin/clients/{id}`

Alteração das configurações do cliente. **Transmissão SSE `config_change` → refletido no proxy em 1-3 segundos.**

**Corpo da requisição (apenas campos a serem alterados):**
```json
{
  "default_service": "openrouter",
  "default_model": "anthropic/claude-opus-4-6",
  "enabled": true
}
```

**Resposta:**
```json
{"status": "updated"}
```

---

#### `DELETE /admin/clients/{id}`

Exclusão de cliente.

---

### API Admin — Serviços

#### `GET /admin/services`

Lista de serviços registrados.

**Exemplo de resposta:**
```json
[
  {"id": "google",      "name": "Google Gemini",   "enabled": true,  "custom": false},
  {"id": "openai",      "name": "OpenAI",          "enabled": true,  "custom": false},
  {"id": "anthropic",   "name": "Anthropic",       "enabled": false, "custom": false},
  {"id": "openrouter",  "name": "OpenRouter",      "enabled": true,  "custom": false},
  {"id": "ollama",      "name": "Ollama (Local)",  "enabled": true,  "custom": false,
   "local_url": "http://localhost:11434"},
  {"id": "lmstudio",    "name": "LM Studio",       "enabled": false, "custom": false},
  {"id": "vllm",        "name": "vLLM",            "enabled": false, "custom": false},
  {"id": "github-copilot","name":"GitHub Copilot", "enabled": false, "custom": false}
]
```

8 serviços integrados: `google`, `openai`, `anthropic`, `openrouter`, `github-copilot`, `ollama`, `lmstudio`, `vllm`

---

#### `POST /admin/services`

Adição de serviço personalizado. Após a adição, o evento SSE `service_changed` é transmitido → **dropdowns do dashboard atualizados instantaneamente**.

**Corpo da requisição:**
```json
{
  "id": "my-llm",
  "name": "사내 LLM 서버",
  "local_url": "http://10.0.0.50:8080",
  "enabled": true
}
```

---

#### `PUT /admin/services/{id}`

Atualização das configurações do serviço. Após a alteração, o evento SSE `service_changed` é transmitido.

**Corpo da requisição:**
```json
{
  "local_url": "http://192.168.x.x:11434",
  "enabled": true
}
```

---

#### `DELETE /admin/services/{id}`

Exclusão de serviço personalizado. Após a exclusão, o evento SSE `service_changed` é transmitido.

Tentativa de excluir serviço integrado (`custom: false`):
```json
{"error": "기본 서비스는 삭제할 수 없습니다: google"}
```

---

### API Admin — Lista de Modelos

#### `GET /admin/models`

Consulta da lista de modelos por serviço. Utiliza cache TTL (10 minutos).

**Parâmetros de consulta:**

| Parâmetro | Descrição | Exemplo |
|-----------|-----------|---------|
| `service` | Filtro de serviço | `?service=google` |
| `q` | Pesquisa de modelo | `?q=gemini` |

**Método de consulta de modelos por serviço:**

| Serviço | Método | Quantidade |
|---------|--------|------------|
| `google` | Lista fixa | 8 (incluindo embedding) |
| `openai` | Lista fixa | 9 |
| `anthropic` | Lista fixa | 6 |
| `github-copilot` | Lista fixa | 6 |
| `openrouter` | Consulta dinâmica via API (fallback para 14 modelos selecionados em caso de falha) | 340+ |
| `ollama` | Consulta dinâmica do servidor local (7 recomendações quando não responde) | Variável |
| `lmstudio` | Consulta dinâmica do servidor local | Variável |
| `vllm` | Consulta dinâmica do servidor local | Variável |
| Personalizado | `/v1/models` compatível com OpenAI | Variável |

**Lista de modelos fallback do OpenRouter (quando a API não responde):**

| Modelo | Observações |
|--------|-------------|
| `openrouter/hunter-alpha` | Gratuito, 1M context |
| `openrouter/healer-alpha` | Gratuito, omni-modal |
| `moonshot/kimi-k2.5` | 256K context |
| `z-ai/glm-5`, `z-ai/glm-4.7-flash` | — |
| `deepseek/deepseek-r1`, `deepseek/deepseek-chat` | — |
| `qwen/qwen-2.5-72b-instruct` | 131K context |
| `minimax/minimax-m2.5` | — |
| `meta-llama/llama-3.3-70b-instruct` | 131K context |

---

### API Admin — Status do Proxy

#### `GET /admin/proxies`

Último status de Heartbeat de todos os proxies conectados.

---

## Tipos de Eventos SSE

Eventos recebidos do stream `/api/events` do vault:

| `type` | Condição de disparo | Conteúdo do `data` | Resposta do dashboard |
|--------|--------------------|--------------------|----------------------|
| `connected` | Imediatamente após conexão SSE | `{"clients": N}` | — |
| `config_change` | Alteração de configuração do cliente | `{"client_id","service","model"}` | Atualização do dropdown de modelo do cartão do agente |
| `key_added` | Registro de nova chave API | `{"service": "google"}` | Atualização do dropdown de modelos |
| `key_deleted` | Exclusão de chave API | `{"service": "google"}` | Atualização do dropdown de modelos |
| `service_changed` | Adição/modificação/exclusão de serviço | `{"action":"added"\|"updated"\|"deleted","id":"...","proxy_services":["google","ollama"]}` | Atualização instantânea do select de serviços + dropdown de modelos; atualização em tempo real da lista de serviços de dispatch do proxy |
| `usage_update` | Ao receber heartbeat do proxy (a cada 20 segundos) | `{"keys":[{"id","service","today_usage","today_attempts","daily_limit","cooldown_until"},...]}` | Atualização instantânea de barras e números de uso das chaves, início da contagem regressiva de cooldown. Uso direto dos dados SSE sem fetch. Barras usam escalonamento share-of-total (chaves ilimitadas). |
| `usage_reset` | Reset do uso diário | `{"time": "RFC3339"}` | Atualização da página |

**Processamento de eventos recebidos pelo proxy:**

```
config_change recebido
  → Se client_id corresponde ao próprio
    → Atualiza service, model instantaneamente
    → hooksMgr.Fire(EventModelChanged)
```

---

## Roteamento de Provedor e Modelo

Ao especificar o formato `provider/model` no campo `model` de `/v1/chat/completions`, o roteamento automático é realizado (compatível com OpenClaw 3.11).

### Regras de Roteamento por Prefixo

| Prefixo | Destino de roteamento | Exemplo |
|---------|----------------------|---------|
| `google/` | Direto para Google | `google/gemini-2.5-pro` |
| `openai/` | Direto para OpenAI | `openai/gpt-4o` |
| `anthropic/` | Via OpenRouter | `anthropic/claude-opus-4-6` |
| `ollama/` | Direto para Ollama | `ollama/qwen3.5:35b` |
| `custom/` | Reparse recursivo (remove `custom/` e redireciona) | `custom/google/gemini-2.5-flash` → Google |
| `openrouter/` | OpenRouter (mantém o bare path) | `openrouter/meta-llama/llama-3.3-70b` |
| `opencode/`, `opencode-go/`, `opencode-zen/` | OpenRouter | `opencode-go/model` |
| `moonshot/`, `kimi-coding/` | OpenRouter (mantém o full path) | `moonshot/kimi-k2.5` |
| `groq/`, `mistral/`, `minimax/`, `cohere/`, `perplexity/` | OpenRouter | `groq/llama-3.1-70b` |
| `together/`, `huggingface/`, `nvidia/`, `venice/` | OpenRouter | — |
| `meta-llama/`, `qwen/`, `deepseek/`, `01-ai/` | OpenRouter (full path) | `deepseek/deepseek-r1` |

### Detecção Automática do Prefixo `wall-vault/`

Prefixo próprio do wall-vault que determina automaticamente o serviço a partir do ID do modelo.

| Padrão do ID do modelo | Roteamento |
|------------------------|------------|
| `wall-vault/gemini-*` | Google |
| `wall-vault/gpt-*`, `wall-vault/o1`, `wall-vault/o3`, `wall-vault/o4*` | OpenAI |
| `wall-vault/claude-*` | OpenRouter (path Anthropic) |
| `wall-vault/hunter-alpha`, `wall-vault/healer-alpha` | OpenRouter (gratuito 1M ctx) |
| `wall-vault/kimi-*`, `wall-vault/glm-*`, `wall-vault/deepseek-*`, `wall-vault/qwen*` | OpenRouter |
| Outros | OpenRouter |

### Processamento do Sufixo `:cloud`

O sufixo `:cloud` no formato de tag Ollama é removido automaticamente e roteado para o OpenRouter.

```
kimi-k2.5:cloud  →  OpenRouter, ID do modelo: kimi-k2.5
glm-5:cloud      →  OpenRouter, ID do modelo: glm-5
```

### Exemplo de Integração OpenClaw openclaw.json

```json5
// ~/.openclaw/openclaw.json
{
  models: {
    providers: {
      "wall-vault": {
        baseUrl: "http://localhost:56244/v1",
        apiKey: "YOUR_AGENT_TOKEN",
        api: "openai-completions",
        models: [
          { id: "wall-vault/gemini-2.5-flash" },
          { id: "wall-vault/hunter-alpha" },
          { id: "wall-vault/claude-opus-4-6" }
        ]
      }
    }
  },
  agents: {
    defaults: {
      model: {
        primary: "wall-vault/gemini-2.5-flash",
        fallbacks: ["wall-vault/hunter-alpha"]
      }
    }
  }
}
```

Ao clicar no botão **🐾** no cartão do agente, o snippet de configuração para aquele agente é copiado automaticamente para a área de transferência.

---

## Esquema de Dados

### APIKey

| Campo | Tipo | Descrição |
|-------|------|-----------|
| `id` | string | ID único no formato UUID |
| `service` | string | `google` \| `openai` \| `anthropic` \| `openrouter` \| `ollama` \| personalizado |
| `encrypted_key` | string | Chave criptografada com AES-GCM (Base64) |
| `label` | string | Rótulo de identificação |
| `today_usage` | int | Número de tokens de requisições bem-sucedidas hoje (não inclui erros 429/402/582) |
| `today_attempts` | int | Total de chamadas API hoje (bem-sucedidas + rate-limited; resetado à meia-noite) |
| `daily_limit` | int | Limite diário (0 = ilimitado) |
| `cooldown_until` | time.Time | Hora de término do cooldown |
| `last_error` | int | Último código de erro HTTP |
| `created_at` | time.Time | Hora de registro |

**Política de cooldown:**

| Erro HTTP | Cooldown |
|-----------|----------|
| 429 (Too Many Requests) | 30 minutos |
| 402 (Payment Required) | 24 horas |
| 400 / 401 / 403 | 24 horas |
| 582 (Gateway Overload) | 5 minutos |
| Erro de rede | 10 minutos |

> **429·402·582**: Define cooldown + incrementa `today_attempts`. `today_usage` não é alterado (apenas tokens bem-sucedidos são contabilizados).
> **Ollama (serviço local)**: `callOllama` usa um cliente HTTP dedicado com `Timeout: 0` (ilimitado). Inferência de modelos grandes pode levar dezenas de segundos a vários minutos, portanto o timeout padrão de 60 segundos não é aplicado.

### Client

| Campo | Tipo | Descrição |
|-------|------|-----------|
| `id` | string | ID único do cliente |
| `name` | string | Nome de exibição |
| `token` | string | Token de autenticação |
| `default_service` | string | Serviço padrão |
| `default_model` | string | Modelo padrão (pode estar no formato `provider/model`) |
| `allowed_services` | []string | Serviços permitidos (array vazio = todos) |
| `agent_type` | string | `openclaw` \| `claude-code` \| `cursor` \| `vscode` \| `custom` |
| `work_dir` | string | Diretório de trabalho do agente |
| `description` | string | Descrição |
| `ip_whitelist` | []string | Lista de IPs permitidos (suporte a CIDR) |
| `avatar` | string | Agent avatar — relative path under `~/.openclaw/` (e.g. `workspace/avatar.png`, `workspace/avatars/profile.hpg`) OR base64 data URI (`data:image/png;base64,...`). Supported extensions: `.png`, `.jpg`/`.jpeg`/`.hpg`, `.webp`, `.gif`. |
| `enabled` | bool | Se `false`, retorna `403` ao acessar `/api/keys` |
| `created_at` | time.Time | Hora de registro |

### ServiceConfig

| Campo | Tipo | Descrição |
|-------|------|-----------|
| `id` | string | ID único do serviço |
| `name` | string | Nome de exibição |
| `local_url` | string | URL do servidor local (Ollama/LMStudio/vLLM/personalizado) |
| `enabled` | bool | Status de ativação |
| `custom` | bool | Se é um serviço adicionado pelo usuário |
| `proxy_enabled` | bool | Whether this service is included in proxy dispatch and agent model dropdowns. Controlled by the "프록시 사용" checkbox in the Services card. Only `proxy_enabled: true` services appear in agent service/model selectors. |

### ProxyStatus (Heartbeat)

| Campo | Tipo | Descrição |
|-------|------|-----------|
| `client_id` | string | ID do cliente |
| `version` | string | Versão do proxy (ex: `v0.1.6.20260314.231308`) |
| `service` | string | Serviço atual |
| `model` | string | Modelo atual |
| `sse_connected` | bool | Status da conexão SSE |
| `host` | string | Hostname |
| `avatar` | string | base64 data URI of local avatar (auto-synced from proxy via heartbeat) |
| `updated_at` | time.Time | Última atualização |
| `vault.today_usage` | int | Uso de tokens hoje |
| `vault.daily_limit` | int | Limite diário |
| `vault.key_status` | string | `active` \| `cooldown` \| `exhausted` |

---

## Respostas de Erro

```json
{"error": "Mensagem de erro"}
```

| Código | Significado |
|--------|-------------|
| 200 | Sucesso |
| 400 | Requisição inválida |
| 401 | Falha na autenticação |
| 403 | Acesso negado (cliente inativo, IP bloqueado) |
| 404 | Recurso não encontrado |
| 405 | Método não permitido |
| 429 | Limite de rate excedido |
| 500 | Erro interno do servidor |
| 502 | Erro de API upstream (todos os fallbacks falharam) |

---

## Coleção de Exemplos cURL

```bash
# ─── Proxy ────────────────────────────────────────────────────────────────────

# Health check
curl http://localhost:56244/health

# Consulta de status
curl http://localhost:56244/status

# Lista de modelos (todos)
curl http://localhost:56244/api/models

# Apenas modelos Google
curl "http://localhost:56244/api/models?service=google"

# Pesquisa de modelos gratuitos
curl "http://localhost:56244/api/models?q=alpha"

# Alteração de modelo (local)
curl -X PUT http://localhost:56244/api/config/model \
  -H "Content-Type: application/json" \
  -d '{"service":"google","model":"gemini-2.5-pro"}'

# Atualização de configurações
curl -X POST http://localhost:56244/reload

# Chamada direta da API Gemini
curl -X POST "http://localhost:56244/google/v1beta/models/gemini-2.5-flash:generateContent" \
  -H "Content-Type: application/json" \
  -d '{"contents":[{"role":"user","parts":[{"text":"안녕하세요"}]}]}'

# Compatível com OpenAI (modelo padrão)
curl -X POST http://localhost:56244/v1/chat/completions \
  -H "Content-Type: application/json" \
  -d '{"model":"gemini-2.5-flash","messages":[{"role":"user","content":"안녕"}]}'

# Formato OpenClaw provider/model
curl -X POST http://localhost:56244/v1/chat/completions \
  -H "Authorization: Bearer my-agent-token" \
  -H "Content-Type: application/json" \
  -d '{"model":"wall-vault/gemini-2.5-flash","messages":[{"role":"user","content":"안녕"}]}'

# Uso de modelo gratuito 1M context
curl -X POST http://localhost:56244/v1/chat/completions \
  -H "Authorization: Bearer my-agent-token" \
  -H "Content-Type: application/json" \
  -d '{"model":"wall-vault/hunter-alpha","messages":[{"role":"user","content":"안녕"}]}'

# ─── Key Vault (Público) ─────────────────────────────────────────────────────

curl http://localhost:56243/api/status
curl http://localhost:56243/api/clients
curl -s http://localhost:56243/api/events --max-time 3

# ─── Key Vault (Admin) ───────────────────────────────────────────────────────

ADMIN="Authorization: Bearer admin-token"

# Lista de chaves
curl -H "$ADMIN" http://localhost:56243/admin/keys

# Adicionar chave Google
curl -X POST http://localhost:56243/admin/keys \
  -H "$ADMIN" -H "Content-Type: application/json" \
  -d '{"service":"google","key":"AIzaSy...","label":"메인 키","daily_limit":1000}'

# Adicionar chave OpenAI
curl -X POST http://localhost:56243/admin/keys \
  -H "$ADMIN" -H "Content-Type: application/json" \
  -d '{"service":"openai","key":"sk-...","label":"GPT 키"}'

# Adicionar chave OpenRouter
curl -X POST http://localhost:56243/admin/keys \
  -H "$ADMIN" -H "Content-Type: application/json" \
  -d '{"service":"openrouter","key":"sk-or-v1-...","label":"OR 키"}'

# Excluir chave (broadcast SSE key_deleted)
curl -X DELETE http://localhost:56243/admin/keys/key-abc123 -H "$ADMIN"

# Reset do uso diário
curl -X POST http://localhost:56243/admin/keys/reset -H "$ADMIN"

# Lista de clientes
curl -H "$ADMIN" http://localhost:56243/admin/clients

# Adicionar cliente (OpenClaw)
curl -X POST http://localhost:56243/admin/clients \
  -H "$ADMIN" -H "Content-Type: application/json" \
  -d '{"id":"bot-a","name":"봇 A","agent_type":"openclaw","work_dir":"~/.openclaw","default_service":"google","default_model":"gemini-2.5-flash"}'

# Alterar modelo do cliente (reflexo instantâneo via SSE)
curl -X PUT http://localhost:56243/admin/clients/bot-a \
  -H "$ADMIN" -H "Content-Type: application/json" \
  -d '{"default_service":"openrouter","default_model":"wall-vault/hunter-alpha"}'

# Desativar cliente
curl -X PUT http://localhost:56243/admin/clients/my-bot \
  -H "$ADMIN" -H "Content-Type: application/json" \
  -d '{"enabled":false}'

# Excluir cliente
curl -X DELETE http://localhost:56243/admin/clients/my-bot -H "$ADMIN"

# Lista de serviços
curl -H "$ADMIN" http://localhost:56243/admin/services

# Definir URL local do Ollama (broadcast SSE service_changed)
curl -X PUT http://localhost:56243/admin/services/ollama \
  -H "$ADMIN" -H "Content-Type: application/json" \
  -d '{"local_url":"http://192.168.x.x:11434","enabled":true}'

# Ativar serviço OpenAI
curl -X PUT http://localhost:56243/admin/services/openai \
  -H "$ADMIN" -H "Content-Type: application/json" \
  -d '{"enabled":true}'

# Adicionar serviço personalizado (broadcast SSE service_changed)
curl -X POST http://localhost:56243/admin/services \
  -H "$ADMIN" -H "Content-Type: application/json" \
  -d '{"id":"my-llm","name":"사내 LLM","local_url":"http://10.0.0.50:8080","enabled":true}'

# Excluir serviço personalizado
curl -X DELETE http://localhost:56243/admin/services/my-llm -H "$ADMIN"

# Consulta da lista de modelos
curl -H "$ADMIN" http://localhost:56243/admin/models
curl -H "$ADMIN" "http://localhost:56243/admin/models?service=openrouter"
curl -H "$ADMIN" "http://localhost:56243/admin/models?q=hunter"

# Status do proxy (heartbeat)
curl -H "$ADMIN" http://localhost:56243/admin/proxies

# ─── Modo Distribuído — Proxy → Vault ────────────────────────────────────────

# Consulta de chaves descriptografadas
curl http://localhost:56243/api/keys \
  -H "Authorization: Bearer your-bot-a-token"

# Envio de Heartbeat
curl -X POST http://localhost:56243/api/heartbeat \
  -H "Authorization: Bearer your-bot-a-token" \
  -H "Content-Type: application/json" \
  -d '{"client_id":"bot-a","version":"v0.1.6.20260314.231308","service":"google","model":"gemini-2.5-flash","sse_connected":true}'
```

---

## Middleware

Aplicado automaticamente a todas as requisições:

| Middleware | Função |
|-----------|--------|
| **Logger** | Log no formato `[method] path status latencyms` |
| **CORS** | `Access-Control-Allow-Origin: *` |
| **Recovery** | Recuperação de panic, retorna resposta 500 |

---

*Última atualização: 2026-03-16 — v0.1.7: dashboard title renamed to "벽금고(wall-vault) 대시보드", logo moved to sticky topbar, today_attempts tracking, HTTP 582 cooldown (5min), share-of-total bar scaling, custom/ routing fix, Ollama Timeout:0, key_att i18n, avatar heartbeat sync, build timestamp versioning, proxy-only service filter*
