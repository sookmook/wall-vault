<p align="center">
  <img src="docs/logo.png" alt="wall-vault" width="800">
</p>

# wall-vault

> **Cofre de chaves de API + proxy de IA num único binário Go.**
> Armazena as chaves localmente com AES-GCM, faz rotação entre fornecedores, recorre a alternativas quando alguma falha, e inclui um painel em tempo real.

[![License: GPL v3](https://img.shields.io/badge/License-GPLv3-blue.svg)](LICENSE)
[![Go](https://img.shields.io/badge/Go-1.25-00ADD8.svg)](go.mod)

[English](README.md) · [한국어](README.ko.md) · [中文](README.zh.md) · [日本語](README.ja.md) · [Español](README.es.md) · [Français](README.fr.md) · [Deutsch](README.de.md) · **Português** · [العربية](README.ar.md) · [हिन्दी](README.hi.md) · [Bahasa Indonesia](README.id.md) · [ภาษาไทย](README.th.md) · [Kiswahili](README.sw.md) · [Hausa](README.ha.md) · [नेपाली](README.ne.md) · [Монгол](README.mn.md) · [isiZulu](README.zu.md)

---

## O que é

O wall-vault fica entre um agente de IA (OpenClaw, Claude Code, Cursor, Continue, o seu próprio script) e os fornecedores de IA na nuvem ou locais com os quais ele conversa. Duas coisas num só binário:

- **Vault** — armazena chaves de API encriptadas em repouso (AES-GCM com uma palavra-passe mestre), faz a sua rotação, regista a utilização e cooldowns por chave, transmite alterações via SSE, e disponibiliza um painel web em `:56243`.
- **Proxy** — expõe endpoints compatíveis com Gemini, Anthropic e OpenAI em `:56244`, escolhe uma chave do cofre, despacha para o upstream configurado, e recorre ao próximo fornecedor quando algum falha.

Suporta quatro formatos de pedido (Gemini `:generateContent`, Anthropic `/v1/messages`, OpenAI `/v1/chat/completions`, e Ollama nativo `/api/chat`) e cinco categorias de upstream:

| Fornecedor | Notas |
|----------|-------|
| Google Gemini | API nativa; rotação de chaves por projeto |
| Anthropic | Passthrough nativo `/v1/messages` |
| OpenAI | `/v1/chat/completions` nativo |
| OpenRouter | 340+ modelos, fallback automático para variantes `:free` |
| Ollama / LM Studio / vLLM / llama.cpp / Jan / KoboldCpp / TabbyAPI / mlx-server / LiteLLM | Backends locais compatíveis com OpenAI; integração via plugin yaml |

Adicionar um novo backend compatível com OpenAI é um único ficheiro yaml em `~/.wall-vault/services/` — sem alterar código.

## Porque é que pode querer usar

- Está a fazer malabarismo com três ou quatro serviços de IA e quer um único URL ao qual o agente fala.
- Quer que uma chave free-tier em cooldown ceda lugar à seguinte sem quebrar a sessão.
- Quer que as mesmas chaves alimentem vários bots / IDEs / scripts na mesma rede local sem copiar credenciais.
- Quer um painel, e não variáveis de ambiente, para editar chaves de API.
- Quer uma opção local-first (Ollama / LM Studio) quando os limites na nuvem se esgotam.

## Início rápido

### Instalação (Linux / macOS)

```bash
curl -fsSL https://raw.githubusercontent.com/sookmook/wall-vault/main/install.sh | sh
```

Ou descarregue um binário pré-compilado diretamente:

```bash
# Linux amd64
curl -L https://github.com/sookmook/wall-vault/releases/latest/download/wall-vault-linux-amd64 \
  -o wall-vault && chmod +x wall-vault

# macOS Apple Silicon
curl -L https://github.com/sookmook/wall-vault/releases/latest/download/wall-vault-darwin-arm64 \
  -o wall-vault && chmod +x wall-vault

# Linux arm64 (Raspberry Pi, servidores ARM)
curl -L https://github.com/sookmook/wall-vault/releases/latest/download/wall-vault-linux-arm64 \
  -o wall-vault && chmod +x wall-vault
```

### Instalação (Windows)

```powershell
Invoke-WebRequest -Uri `
  "https://github.com/sookmook/wall-vault/releases/latest/download/wall-vault-windows-amd64.exe" `
  -OutFile wall-vault.exe
```

### Primeira execução

```bash
wall-vault setup    # assistente interativo — escolhe porta, serviços, token de admin, palavra-passe mestre
wall-vault start    # arranca o vault e o proxy
```

Abra `http://localhost:56243` (ou `https://...` assim que o TLS estiver ativo — ver abaixo) num navegador. O painel pede o token de admin impresso pelo `setup`. A partir daí adiciona chaves de API, regista clientes e troca de modelos sem reiniciar.

---

## TLS (recomendado)

Por predefinição, o `wall-vault setup` escreve uma configuração sem TLS, pelo que ambos os listeners respondem em HTTP simples. Os URLs de exemplo neste README usam `https://localhost:56244` porque a maioria dos agentes (OpenClaw, Claude Code, Cursor) prefere um único endpoint frontalizado por TLS que não se quebre se mais tarde mover o proxy para outro host. Para coincidir com esses exemplos, ative o TLS uma vez com a CA interna fornecida:

```bash
# 1. Crie a CA interna do wall-vault (uma única vez, fica em ~/.wall-vault/ca.{crt,key})
wall-vault cert init

# 2. Emita um certificado de host para ESTA máquina
#    Os SANs incluem hostname, localhost, 127.0.0.1, e qualquer IP de LAN detetado
wall-vault cert issue $(hostname)

# 3. Confie na CA no chaveiro do sistema operativo local
wall-vault cert install-trust

# 4. Mude os listeners para TLS
export WV_PROXY_TLS_ENABLED=1
export WV_PROXY_TLS_CERT="$HOME/.wall-vault/$(hostname).crt"
export WV_PROXY_TLS_KEY="$HOME/.wall-vault/$(hostname).key"
export WV_VAULT_TLS_ENABLED=1
export WV_VAULT_TLS_CERT="$HOME/.wall-vault/$(hostname).crt"
export WV_VAULT_TLS_KEY="$HOME/.wall-vault/$(hostname).key"
wall-vault start
```

Para outra máquina na sua LAN: copie `~/.wall-vault/ca.crt` para lá e execute aí `wall-vault cert install-trust --ca <caminho>`. Quando a CA for confiável em todo o lado, qualquer máquina da rede pode aceder ao proxy via `https://<host>:56244` sem avisos de certificado.

Se preferir manter HTTP simples, deixe a configuração tal como está e troque `https://` por `http://` nos excertos de cliente abaixo. Ambos os esquemas funcionam; a diferença é qual a porta que responde a um handshake TLS.

**Fallback de loopback.** Os clientes no mesmo host que não conseguem honrar a CA do wall-vault (em particular o runtime Node empacotado com o OpenClaw, que reescreve `NODE_EXTRA_CA_CERTS` ao iniciar) chegam ao proxy através de um companheiro HTTP simples só de loopback em `127.0.0.1:56245`. O wall-vault ativa-o automaticamente quando o TLS está ligado.

---

## Ligação de clientes

Aponte qualquer cliente de IA para `https://<host>:56244` (ou `http://...` se o TLS estiver desligado). O proxy responde em quatro formatos:

| Formato | Caminho | Exemplos de clientes |
|--------|------|-----------------|
| Gemini | `/google/v1beta/models/<model>:generateContent` | OpenClaw, Gemini CLI, Antigravity |
| Anthropic | `/v1/messages` | Claude Code, SDKs Anthropic |
| OpenAI | `/v1/chat/completions` | Cursor, Continue, scripts personalizados, a maioria das apps LLM |
| Ollama nativo | `/api/chat` | Clientes Ollama em passthrough |

### Claude Code

```bash
export ANTHROPIC_BASE_URL=https://localhost:56244
export ANTHROPIC_API_KEY=<o-seu-token-de-cliente-do-vault>
claude
```

Quando os créditos Anthropic upstream se esgotam, o despacho recorre aos fornecedores que tiver definido em `fallback_services` para este cliente. Para ativar explicitamente um fallback não-Claude:

```yaml
proxy:
  anthropic_fallback_model: "claude-haiku-4-5-20251001"
```

(O valor predefinido vazio faz com que o despacho devolva um erro, para que qualquer encaminhamento errado apareça imediatamente.)

### Cursor / Continue

No Cursor **Settings → AI → OpenAI API**:

```
Base URL:  https://localhost:56244
API Key:   <o-seu-token-de-cliente-do-vault>
Model:     gemini-2.5-flash    # ou qualquer modelo conhecido pelo wall-vault
```

Continue (`config.json`):

```json
{
  "models": [
    {
      "title": "wall-vault",
      "provider": "openai",
      "model": "gemini-2.5-flash",
      "apiBase": "https://localhost:56244/v1",
      "apiKey": "<o-seu-token-de-cliente-do-vault>"
    }
  ]
}
```

### OpenClaw

O OpenClaw é uma framework de agentes TUI que o wall-vault foi originalmente concebido para servir. O modal **Add Agent** do painel define o tipo de agente como `openclaw` (ou `nanoclaw`); o wall-vault escreve então diretamente `~/.openclaw/openclaw.json`, incluindo URLs dos fornecedores, o token do vault e as entradas de modelo:

```bash
VAULT_CLIENT_ID=my-bot \
VAULT_URL=http://<vault-host>:56243 \
VAULT_TOKEN=<o-seu-token-de-cliente> \
wall-vault proxy
```

### curl / scripts

```bash
curl -X POST https://localhost:56244/v1/chat/completions \
  -H "Authorization: Bearer <o-seu-token-de-cliente-do-vault>" \
  -H "Content-Type: application/json" \
  -d '{
    "model": "gemini-2.5-flash",
    "messages": [{"role": "user", "content": "ping"}]
  }'
```

---

## Configuração

O `wall-vault setup` escreve `./wall-vault.yaml` ou `~/.wall-vault/config.yaml`. Edite à mão os campos sobre os quais o assistente não pergunta.

```yaml
mode: standalone     # standalone | distributed
lang: en             # en | ko | zh | ja | es | fr | de | pt | ar | hi | id | th | sw | ha | ne | mn | zu
theme: light         # light | dark | cherry | ocean | gold | autumn | winter

proxy:
  port: 56244
  host: ""           # predefinição: 127.0.0.1 standalone, 0.0.0.0 distributed
  client_id: my-bot
  vault_url: ""      # distributed: http://vault-host:56243
  vault_token: ""    # distributed: token de cliente
  tool_filter: strip_all   # strip_all | whitelist | passthrough
  allowed_tools: []
  services: [google, openrouter, ollama]
  timeout: 300s
  tls:
    enabled: false
    cert_file: ""
    key_file: ""
  plain_port: 56245              # companheiro HTTP só de loopback quando TLS está ligado
  ollama_keep_alive: "30m"       # "-1" nunca descarregar, "0" descarregar imediatamente
  ollama_num_ctx: 8192
  oai_stream_forward: false      # passthrough opcional de SSE do backend real
  anthropic_fallback_model: ""   # reescrita opcional não-Claude no despacho anthropic

vault:
  port: 56243
  host: ""
  admin_token: ""
  admin_ip_whitelist: []
  master_password: ""            # palavra-passe de encriptação da chave AES-GCM
  data_dir: ~/.wall-vault/data
  services_dir: ~/.wall-vault/services
  tls:
    enabled: false
    cert_file: ""
    key_file: ""
  bootstrap_port: 56247          # listener HTTP simples que serve apenas ca.crt

doctor:
  interval: 5m
  auto_fix: true
  log_file: /tmp/wall-vault-doctor.log

hooks:
  on_model_change: ""            # comando shell (env: SERVICE, MODEL)
  on_key_exhausted: ""
  on_service_down: ""
  on_doctor_fix: ""
  openclaw_socket: ""
```

### Variáveis de ambiente

Cada campo YAML tem uma sobreposição via env que prevalece sobre o ficheiro. As mais comuns:

| Variável | Descrição |
|----------|-------------|
| `WV_LANG`, `WV_THEME` | Idioma e tema |
| `WV_PROXY_PORT`, `WV_PROXY_HOST` | Endereço de escuta do proxy |
| `WV_VAULT_PORT`, `WV_VAULT_HOST` | Endereço de escuta do vault |
| `WV_VAULT_URL`, `WV_VAULT_TOKEN` | Endpoints em modo distribuído |
| `WV_ADMIN_TOKEN`, `WV_MASTER_PASS` | Credenciais do vault |
| `WV_KEY_GOOGLE`, `WV_KEY_OPENROUTER`, `WV_KEY_ANTHROPIC`, `WV_KEY_OPENAI` | Chaves de API (separadas por vírgulas para várias) |
| `WV_PROXY_TLS_ENABLED`, `WV_PROXY_TLS_CERT`, `WV_PROXY_TLS_KEY` | TLS do proxy |
| `WV_VAULT_TLS_ENABLED`, `WV_VAULT_TLS_CERT`, `WV_VAULT_TLS_KEY` | TLS do vault |
| `WV_PROXY_PLAIN_PORT` | Companheiro HTTP de loopback (`0` para desativar) |
| `WV_VAULT_BOOTSTRAP_PORT` | Listener de bootstrap da CA (`0` para desativar) |
| `WV_OLLAMA_URL`, `WV_OLLAMA_KEEP_ALIVE`, `WV_OLLAMA_NUM_CTX` | Afinação do Ollama |
| `WV_LMSTUDIO_URL`, `WV_VLLM_URL`, `WV_LLAMACPP_URL` | Sobreposições de backend local |
| `WV_TOKEN_SENTINEL_FALLBACK` | Substituição da sentinela "proxy-managed" em loopback |
| `WV_OAI_STREAM_FORWARD` | Passthrough SSE do backend real compatível com OpenAI |
| `WV_ANTHROPIC_FALLBACK_MODEL` | Reescrita opcional não-Claude no anthropic |

---

## Modos

### Standalone (predefinição)

Vault e proxy correm no mesmo processo. Ideal para um único host que aloja tanto as chaves como o agente. Por predefinição escuta apenas em loopback.

```bash
wall-vault start    # corre os dois
```

### Distributed

O vault corre num host (o **vault host**) e armazena todas as chaves; vários proxies em outros hosts autenticam-se cada um com um token por cliente. Útil quando várias máquinas precisam das mesmas chaves sem ter de as copiar de um lado para o outro.

**Vault host:**

```bash
WV_VAULT_HOST=0.0.0.0 wall-vault vault
```

**Cada host de proxy:**

```bash
WV_VAULT_URL=http://<vault-host>:56243 \
WV_VAULT_TOKEN=<este-token-de-cliente> \
WV_PROXY_HOST=0.0.0.0 \
wall-vault proxy
```

O modal **Add Client** do painel cunha um token, regista um tipo de agente, e o proxy recebe a sua configuração via SSE sem reiniciar.

---

## Plugin yaml (backend pronto a usar)

Qualquer backend compatível com OpenAI pode ser adicionado como yaml em `~/.wall-vault/services/`. O wall-vault deteta-o no arranque, regista-o como serviço encaminhável, e o despacho + o conjunto de deteção compatível com OAI + a ponte de stream do Gemini veem-no todos sem alterações de código.

```yaml
# ~/.wall-vault/services/llamacpp.yaml
id: llamacpp
name: llama.cpp
enabled: true
default_url: http://localhost:8080
endpoints:
  generate: /v1/chat/completions
  list_models: /v1/models
auth:
  type: none
request_format: openai
inline_no_think_for_qwen3: false   # ative se o seu backend remover o marcador
```

A topologia hub (um wall-vault à frente de outro) é suportada via `tls_internal_ca: true`, `auth.type: bearer` e `preserve_model_id: true`.

---

## Compilar a partir do código-fonte

```bash
git clone https://github.com/sookmook/wall-vault
cd wall-vault
go build -o wall-vault .
./wall-vault setup
./wall-vault start
```

Compilação cruzada para todo o conjunto suportado:

```bash
make build-all   # linux-amd64, linux-arm64, darwin-amd64, darwin-arm64, windows-amd64
```

As versões seguem `v{major}.{minor}.{patch}.{YYYYMMDD}.{HHmmss}`; `BASE_VERSION` no Makefile define o prefixo.

### Estrutura do projeto

```
wall-vault/
├── main.go                     # despacho de CLI (start/proxy/vault/setup/cert/doctor)
├── cmd/
│   ├── setup/                  # assistente de configuração interativo
│   └── cert/                   # CA interna + emissor de certificado TLS por host
├── internal/
│   ├── config/                 # carregador YAML + env, carregador de plugins
│   ├── proxy/                  # despacho de pedidos, rotação de chaves, conversores de formato
│   ├── vault/                  # store AES-GCM, painel, broker SSE
│   ├── doctor/                 # sonda de saúde + auto-fix
│   ├── hooks/                  # acionadores de eventos por comando shell
│   └── i18n/                   # strings de UI em 17 idiomas
├── configs/services/           # plugins yaml incluídos (lmstudio, vllm, ollama, …)
└── docs/                       # MANUAL, referência da API, 16 variantes locais
```

---

## Documentação

- [Manual do utilizador](docs/MANUAL.en.md) — instalação, painel, agentes, resolução de problemas
- [Referência da API](docs/API.en.md) — cada endpoint com formatos de pedido/resposta
- [CHANGELOG](CHANGELOG.md)

---

## Pilha tecnológica

- Go 1.25, único binário estático
- [templ](https://templ.guide) para o painel renderizado no servidor, [HTMX](https://htmx.org) para atualizações parciais
- AES-GCM (chave derivada por PBKDF2) para encriptação de chaves em repouso
- Server-Sent Events para sincronização de configuração ao vivo entre vault e proxies
- CA interna autoassinada + certificados por host (sem necessidade de DNS público / Let's Encrypt)

## Licença

GPL-3.0. Veja [LICENSE](LICENSE).

## Contribuir

Pull requests são bem-vindos. Veja [CONTRIBUTING.md](CONTRIBUTING.md). Para alterações maiores, por favor abra primeiro uma issue para discutir o desenho.
