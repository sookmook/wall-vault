# Manual do Usuário do wall-vault

[English](MANUAL.md) · [한국어](MANUAL.ko.md) · [中文](MANUAL.zh.md) · [日本語](MANUAL.ja.md) · [Español](MANUAL.es.md) · [Français](MANUAL.fr.md) · [Deutsch](MANUAL.de.md) · **Português** · [العربية](MANUAL.ar.md) · [हिन्दी](MANUAL.hi.md) · [Bahasa Indonesia](MANUAL.id.md) · [ภาษาไทย](MANUAL.th.md) · [Kiswahili](MANUAL.sw.md) · [Hausa](MANUAL.ha.md) · [नेपाली](MANUAL.ne.md) · [Монгол](MANUAL.mn.md) · [isiZulu](MANUAL.zu.md)

Este manual cobre a instalação, configuração e operação do wall-vault. Para uma visão geral rápida, consulte o [README](../README.md). Para detalhes da API HTTP, consulte a [referência da API](API.md).

## Conteúdo

1. [O que o wall-vault faz](#o-que-o-wall-vault-faz)
2. [Instalação](#instalação)
3. [Primeira execução com o assistente de configuração](#primeira-execução-com-o-assistente-de-configuração)
4. [Habilitando TLS](#habilitando-tls)
5. [Registrando chaves de API](#registrando-chaves-de-api)
6. [Conectando agentes](#conectando-agentes)
7. [O painel](#o-painel)
8. [Modo distribuído](#modo-distribuído)
9. [Início automático](#início-automático)
10. [yamls de plugin](#yamls-de-plugin)
11. [Doctor](#doctor)
12. [Hooks](#hooks)
13. [Variáveis de ambiente](#variáveis-de-ambiente)
14. [Solução de problemas](#solução-de-problemas)

---

## O que o wall-vault faz

O wall-vault é um único binário Go que agrupa dois serviços cooperantes:

- **O cofre (vault)** armazena chaves de API criptografadas em repouso (AES-GCM com uma senha mestra), monitora o uso e os cooldowns por chave, transmite alterações via Server-Sent Events (SSE) e disponibiliza um painel web em `:56243` para operadores humanos.
- **O proxy** expõe endpoints Gemini, Anthropic, compatíveis com OpenAI e nativos do Ollama em `:56244`. Qualquer cliente de IA que aponte para o proxy está usando as chaves do cofre — os clientes nunca as veem. Quando um upstream falha, o despacho recorre ao próximo provedor na ordem.

Isso é útil quando:

- Você tem chaves para vários provedores e quer uma única URL com a qual o agente fala.
- Você quer que uma chave free-tier em cooldown se afaste sem interromper a sessão.
- Você quer que as mesmas chaves alimentem vários bots, IDEs ou scripts na mesma LAN sem copiar credenciais.
- Você quer um painel, e não variáveis de ambiente, para editar chaves e trocar de modelo.
- Você quer um fallback local (Ollama, LM Studio, vLLM) quando os limites de nuvem se esgotam.

```
   AI client (OpenClaw, Claude Code, Cursor, …)
            │
            ▼
   wall-vault proxy  :56244
            │  (selects key, dispatches, falls back on failure)
            ├──► Google Gemini
            ├──► Anthropic
            ├──► OpenAI
            ├──► OpenRouter (340+ models, auto :free fallback)
            └──► Local OAI-compat backends (Ollama / LM Studio / vLLM / …)

   vault (AES-GCM key store + dashboard)  :56243
            ▲
            │  SSE broadcast on change
   Multiple proxies on different hosts can share one vault.
```

---

## Instalação

### Linux / macOS one-liner

```bash
curl -fsSL https://raw.githubusercontent.com/sookmook/wall-vault/main/install.sh | sh
```

O script detecta automaticamente o sistema operacional e a arquitetura, baixa o binário correto para `~/.local/bin/wall-vault` e o torna executável. Se `~/.local/bin` não estiver no seu `PATH`, adicione-o:

```bash
export PATH="$HOME/.local/bin:$PATH"
```

### Download manual

Binários pré-compilados são publicados em cada release em `https://github.com/sookmook/wall-vault/releases`.

```bash
# Linux amd64
curl -L https://github.com/sookmook/wall-vault/releases/latest/download/wall-vault-linux-amd64 \
  -o wall-vault && chmod +x wall-vault

# Linux arm64 (Raspberry Pi, ARM servers)
curl -L https://github.com/sookmook/wall-vault/releases/latest/download/wall-vault-linux-arm64 \
  -o wall-vault && chmod +x wall-vault

# macOS Apple Silicon
curl -L https://github.com/sookmook/wall-vault/releases/latest/download/wall-vault-darwin-arm64 \
  -o wall-vault && chmod +x wall-vault

# macOS Intel
curl -L https://github.com/sookmook/wall-vault/releases/latest/download/wall-vault-darwin-amd64 \
  -o wall-vault && chmod +x wall-vault
```

### Windows

```powershell
Invoke-WebRequest -Uri `
  "https://github.com/sookmook/wall-vault/releases/latest/download/wall-vault-windows-amd64.exe" `
  -OutFile wall-vault.exe
```

### Compilando a partir do código-fonte

Requer Go 1.25 ou mais recente.

```bash
git clone https://github.com/sookmook/wall-vault
cd wall-vault
go build -o wall-vault .
```

`make build-all` faz cross-compile para todas as cinco plataformas suportadas. Os binários são gerados em `bin/`.

---

## Primeira execução com o assistente de configuração

```bash
wall-vault setup
```

O assistente solicita, em ordem:

1. **Idioma** — escolhe uma das 17 localidades de UI. Detectado automaticamente a partir de `$LANG`; o assistente oferece uma lista mesmo assim.
2. **Tema** — `light` (padrão), `dark`, `cherry`, `ocean`, `gold`, `autumn`, `winter`. Apenas cosmético.
3. **Modo** — `standalone` (host único, padrão) ou `distributed` (cofre em um host, proxies em outros).
4. **Nome do bot** — um slug `client_id` livre. O cofre usa isso para escopar a configuração por cliente (substituições de modelo, cadeias de fallback).
5. **Porta do proxy** — padrão `56244`.
6. **Porta do cofre** — padrão `56243` (somente standalone).
7. **Seleção de serviços** — um y/N para cada um de: Google Gemini, OpenRouter, Anthropic, OpenAI, Ollama, LM Studio, vLLM. Múltiplas escolhas são aceitas; cada uma escreve sua dica de variável de ambiente no final.
8. **Filtro de ferramentas** — `strip_all` (padrão; bloqueia todas as definições de ferramentas recebidas por segurança) ou `passthrough` (deixa qualquer ferramenta passar).
9. **Token admin** — deixe em branco para gerar automaticamente. O painel exige este token para o login.
10. **Senha mestra** — deixe em branco para nenhuma criptografia (NÃO recomendado); defina um valor para criptografar o armazenamento de chaves em repouso com AES-GCM.
11. **Caminho de salvamento** — padrão `wall-vault.yaml` no diretório atual. O carregador também consulta `~/.wall-vault/config.yaml`.

Após salvar, o assistente executa `doctor.FixTrust` para que qualquer agente instalado localmente (OpenClaw, Claude Code, Cline) receba automaticamente a CA interna do wall-vault em sua trust store. Se nenhum agente desse tipo estiver instalado, o passo imprime `SKIP` e não escreve nada.

Em seguida, inicie o binário:

```bash
wall-vault start
```

`start` executa tanto o cofre quanto o proxy em um único processo (modo standalone). Para o modo distribuído, use `wall-vault vault` no host do cofre e `wall-vault proxy` em cada host de proxy.

Abra `http://localhost:56243` em um navegador. Faça login com o token admin que o assistente imprimiu.

---

## Habilitando TLS

Os padrões do assistente deixam ambos os listeners em HTTP simples. A maioria dos agentes (OpenClaw, Claude Code, Cursor) funciona melhor contra um único endpoint HTTPS, então TLS é recomendado em qualquer implantação que vá além da máquina local.

O wall-vault vem com sua própria CA interna, então você não precisa de um nome DNS público nem de Let's Encrypt.

```bash
# 1. Create the internal CA — written to ~/.wall-vault/ca.{crt,key}.
#    The CA is good for 10 years by default; override with --ca-years.
wall-vault cert init

# 2. Issue a host certificate. Subject Alternative Names automatically include:
#       hostname, "localhost", "127.0.0.1", and any non-loopback LAN IP detected.
#    Override the issuer dir with --dir, validity with --host-years.
wall-vault cert issue $(hostname)

# 3. Trust the CA in this machine's OS keychain.
#    Linux: writes to /etc/ssl/certs/ via update-ca-certificates (needs sudo).
#    macOS: adds to the System keychain via security add-trusted-cert (needs sudo).
#    Windows: imports into CurrentUser\Root via certutil (no admin needed).
wall-vault cert install-trust

# 4. Enable TLS on both listeners.
export WV_PROXY_TLS_ENABLED=1
export WV_PROXY_TLS_CERT="$HOME/.wall-vault/$(hostname).crt"
export WV_PROXY_TLS_KEY="$HOME/.wall-vault/$(hostname).key"
export WV_VAULT_TLS_ENABLED=1
export WV_VAULT_TLS_CERT="$HOME/.wall-vault/$(hostname).crt"
export WV_VAULT_TLS_KEY="$HOME/.wall-vault/$(hostname).key"

wall-vault start
```

Para estender a confiança a outras máquinas da LAN, copie `~/.wall-vault/ca.crt` e execute `wall-vault cert install-trust --ca <path>` em cada uma. O cofre também expõe `ca.crt` por meio de um pequeno listener HTTP simples em `:56247` (a **porta de bootstrap**) para o caso do dilema em que um cliente novo precisa da CA para falar HTTPS.

### Companheiro HTTP em loopback

Alguns agentes — notavelmente o runtime Node empacotado do OpenClaw — reescrevem `NODE_EXTRA_CA_CERTS` no spawn do processo, descartando qualquer dica de CA fornecida pelo operador. Eles não conseguem honrar a CA do wall-vault de dentro do daemon, mesmo após `cert install-trust`. O wall-vault contorna isso vinculando um **listener HTTP simples adicional restrito ao loopback** em `127.0.0.1:56245` sempre que TLS está habilitado. Clientes do mesmo host alcançam o proxy por meio dessa porta sem TLS algum; clientes da LAN continuam usando o listener TLS.

Desabilite com `WV_PROXY_PLAIN_PORT=0` se você não precisar.

### `wall-vault cert list`

Mostra cada certificado em `~/.wall-vault/` com sujeito, janela de validade e SANs.

```
$ wall-vault cert list
ca.crt          subject=wall-vault internal CA   not-after=2036-05-05
hostname.crt    subject=hostname                 not-after=2031-05-05   SAN=hostname,localhost,127.0.0.1,192.168.…
```

---

## Registrando chaves de API

Duas formas: o painel ou variáveis de ambiente.

### Painel (recomendado)

1. Faça login em `https://localhost:56243` com o token admin.
2. Clique em **+ API key** no card de chaves.
3. Escolha um serviço (Google, OpenRouter, Anthropic, OpenAI, …).
4. Cole a chave. Salve.

Múltiplas chaves por serviço são permitidas; o proxy faz round-robin entre elas e ignora aquelas que atingiram um cooldown por chave.

### Variáveis de ambiente (bootstrap único)

```bash
export WV_KEY_GOOGLE="AIzaSyA1...,AIzaSyB2...,AIzaSyC3..."   # comma-separated
export WV_KEY_OPENROUTER="sk-or-v1-…"
export WV_KEY_ANTHROPIC="sk-ant-…"
export WV_KEY_OPENAI="sk-…"
wall-vault start
```

Chaves fornecidas dessa forma são gravadas no armazenamento criptografado na primeira inicialização. Inicializações subsequentes as leem do disco; você pode desativar as variáveis de ambiente após a primeira execução.

### Cooldowns e rotação

Cada chamada bem-sucedida incrementa o `usage_count` da chave e atualiza `last_used`. Em HTTP 429 / 402 / 403, o proxy coloca a chave em **cooldown** (padrões: 60 minutos para 429, 24 horas para 402, 12 horas para 403). O próximo despacho escolhe outra chave para esse serviço. Quando todas as chaves de um serviço estão em cooldown, o proxy pula esse serviço inteiramente e tenta o próximo provedor na cadeia de fallback.

Os cooldowns são visíveis por chave no painel com uma contagem regressiva.

---

## Conectando agentes

### OpenClaw

OpenClaw é o cliente alvo original. Use o modal **+ Add agent** do painel:

- Defina **Agent type** como `openclaw` ou `nanoclaw`.
- Defina **Work directory** — para OpenClaw isso é preenchido automaticamente como `~/.openclaw`.
- Escolha um **preferred service** e opcionalmente um **model override**.
- Clique em **Apply**. O wall-vault grava `~/.openclaw/openclaw.json` diretamente (URLs de provedor, token do cofre, entradas de modelo).

Quando você muda o modelo no painel, o OpenClaw capta a mudança via SSE em 1–3 segundos — sem reiniciar.

### Claude Code

```bash
export ANTHROPIC_BASE_URL=https://localhost:56244
export ANTHROPIC_API_KEY=<your-vault-client-token>
claude
```

Quando os créditos upstream da Anthropic acabarem, o despacho recorre aos serviços listados em `fallback_services` deste cliente. Por padrão, um id de modelo não-Claude enviado ao despacho anthropic retorna um erro para que erros de roteamento apareçam imediatamente. Habilite a reescrita automática:

```yaml
proxy:
  anthropic_fallback_model: "claude-haiku-4-5-20251001"
```

### Cursor

No Cursor **Settings → AI → OpenAI API**:

```
Base URL:  https://localhost:56244
API Key:   <your-vault-client-token>
Model:     gemini-2.5-flash    # or any model wall-vault knows
```

### Continue (VS Code, JetBrains)

`config.json`:

```json
{
  "models": [
    {
      "title": "wall-vault",
      "provider": "openai",
      "model": "gemini-2.5-flash",
      "apiBase": "https://localhost:56244/v1",
      "apiKey": "<your-vault-client-token>"
    }
  ]
}
```

### HTTP personalizado

```bash
curl -X POST https://localhost:56244/v1/chat/completions \
  -H "Authorization: Bearer <your-vault-client-token>" \
  -H "Content-Type: application/json" \
  -d '{
    "model": "gemini-2.5-flash",
    "messages": [{"role": "user", "content": "hello"}]
  }'
```

O mesmo endpoint aceita streaming (`"stream": true`) quando `proxy.oai_stream_forward: true` está definido.

---

## O painel

`https://localhost:56243`. Cinco cards na grade inicial:

- **Keys** — cada chave de API, agrupada por serviço. Adicionar, editar, excluir; ver uso e cooldown.
- **Services** — Google / OpenRouter / Anthropic / OpenAI / Ollama / LM Studio / vLLM / llama.cpp, mais qualquer plugin yaml em `~/.wall-vault/services/`. Definir `default_model`, `allowed_models` por serviço, URL base e alternância de raciocínio.
- **Clients (agents)** — cada cliente registrado (bot OpenClaw, sessão Claude Code, instância Cursor, …). Atribuir serviço preferido, substituição de modelo, cadeia de fallback.
- **Proxies** — cada proxy que se autenticou contra este cofre. Status ao vivo (online/offline), última visualização, modelo atual.
- **Settings** — token admin, rotação de senha mestra, tema, idioma.

Cada card tem um slideover de edição (lado direito). Clique externo ou `Esc` o fecha. As alterações são enviadas a todos os proxies conectados via SSE em segundos.

O **rodapé** carrega um indicador SSE (verde = conectado, laranja = reconectando, cinza = desconectado) e a versão de build ao vivo.

---

## Modo distribuído

Quando você tem várias máquinas que precisam todas das mesmas chaves, execute o cofre em um host e proxies em cada um dos outros.

### Host do cofre

```bash
WV_VAULT_HOST=0.0.0.0 \
WV_ADMIN_TOKEN=<admin> \
WV_MASTER_PASS=<master> \
wall-vault vault
```

O painel agora está acessível em `https://<vault-host>:56243`. Adicione um agente para cada proxy remoto no card **Clients**; cada um cunha um `vault_token` único.

### Hosts de proxy

```bash
WV_VAULT_URL=http://<vault-host>:56243 \
WV_VAULT_TOKEN=<that-client-token> \
WV_PROXY_HOST=0.0.0.0 \
wall-vault proxy
```

O proxy se autentica contra o cofre, abre um stream SSE e aplica qualquer configuração que receber (serviço preferido, substituição de modelo, cadeia de fallback). Edições subsequentes do cofre chegam em segundos sem reiniciar.

Para instalações que abrangem a LAN, habilite TLS no host do cofre (`WV_VAULT_TLS_ENABLED=1` + as variáveis de ambiente cert/key) e passe cada host de proxy pelo mesmo passo `wall-vault cert install-trust` para que as chamadas HTTPS do proxy ao cofre sejam confiáveis.

---

## Início automático

### systemd (Linux)

```ini
# ~/.config/systemd/user/wall-vault-proxy.service
[Unit]
Description=wall-vault proxy
After=network-online.target

[Service]
Type=simple
ExecStart=%h/.local/bin/wall-vault proxy
Restart=on-failure
RestartSec=5

[Install]
WantedBy=default.target
```

```bash
systemctl --user enable --now wall-vault-proxy
loginctl enable-linger $USER       # so the unit keeps running after logout
```

Para o cofre no mesmo host, escreva um `wall-vault-vault.service` paralelo. Para o modo standalone, uma única unidade chamando `wall-vault start` é suficiente.

### launchd (macOS)

```xml
<!-- ~/Library/LaunchAgents/com.wall-vault.proxy.plist -->
<?xml version="1.0" encoding="UTF-8"?>
<!DOCTYPE plist PUBLIC "-//Apple//DTD PLIST 1.0//EN"
  "http://www.apple.com/DTDs/PropertyList-1.0.dtd">
<plist version="1.0">
<dict>
  <key>Label</key><string>com.wall-vault.proxy</string>
  <key>ProgramArguments</key>
  <array><string>/usr/local/bin/wall-vault</string><string>proxy</string></array>
  <key>RunAtLoad</key><true/>
  <key>KeepAlive</key><true/>
  <key>StandardOutPath</key><string>/tmp/wall-vault.proxy.log</string>
  <key>StandardErrorPath</key><string>/tmp/wall-vault.proxy.err</string>
</dict>
</plist>
```

```bash
launchctl load ~/Library/LaunchAgents/com.wall-vault.proxy.plist
```

### Windows

Use `nssm` para envolver `wall-vault.exe start` como um serviço Windows, ou uma entrada `schtasks` que execute no logon do usuário.

---

## yamls de plugin

Qualquer backend compatível com OpenAI pode ser adicionado sem alterações de código colocando um yaml em `~/.wall-vault/services/`. O wall-vault o carrega na inicialização e registra o serviço para despacho, o conjunto de detecção OAI-compat e a ponte Gemini-stream.

```yaml
# ~/.wall-vault/services/llamacpp.yaml
id: llamacpp                 # unique service id
name: llama.cpp              # human label
enabled: true                # disabled plugins are skipped at load

default_url: http://localhost:8080   # operator override; env wins (WV_LLAMACPP_URL)
endpoints:
  generate: /v1/chat/completions
  list_models: /v1/models

auth:
  type: none                 # none | bearer | query_param | header
  param: ""                  # for query_param: the param name (e.g. "key")

request_format: openai       # openai | gemini | ollama | raw

model_fetch:
  enabled: true              # let the dashboard auto-detect models
  dynamic: true              # re-fetch on every dashboard open
  auto_detect_url: true      # try /v1/models even when not declared

concurrency:
  max: 1                     # max concurrent requests to this backend
  queue_size: 10
  wait_notify: true          # show "queued" hint to TUI agents

error_codes:
  503:
    cooldown: 5m
    message: "llama.cpp not responding"

# Opt in to qwen3-family inline /no_think directive when reasoning is off.
# Set true if your backend's chat template strips the marker (LM Studio's
# jinja, Ollama's /v1 layer). Other backends typically echo the literal
# text back, so this stays opt-in per yaml.
inline_no_think_for_qwen3: false

# Hub topology — point at another wall-vault. Required when this plugin
# fronts a remote wall-vault (so the receiving wall-vault sees the
# publisher prefix and routes correctly) and so the bearer token in
# proxy.vault_token is sent as Authorization.
preserve_model_id: false
tls_internal_ca: false       # add ~/.wall-vault/ca.crt to client trust pool
```

O conjunto empacotado em `configs/services/` (lmstudio, vllm, llamacpp, tgwui, localai, jan, koboldcpp, tabbyapi, mlx-server, litellm-proxy, ollama, google, openrouter) vem desativado por padrão. Copie o que você quiser para `~/.wall-vault/services/`, defina `enabled: true`, reinicie.

---

## Doctor

`wall-vault doctor` executa uma sondagem de saúde única em toda a instalação:

```
✓ vault listener  (https://localhost:56243)
✓ proxy listener  (https://localhost:56244)
✓ master password set
⚠ Google: 2 keys, all on cooldown
✓ Anthropic: 1 key healthy
✗ Ollama: not reachable at http://localhost:11434
```

Cada linha é uma de:

- `✓` — saudável
- `⚠` — degradado mas funcional (uma chave em cooldown, cota baixa, etc.)
- `✗` — quebrado
- `SKIP` — não configurado / não aplicável neste host

Um segundo modo daemon executa a mesma sondagem a cada `doctor.interval` (padrão 5 minutos) e grava os resultados em `doctor.log_file` (padrão `/tmp/wall-vault-doctor.log`). Quando `doctor.auto_fix` é true, ele também tenta reparar drift comum (configuração OpenClaw obsoleta, trust TLS ausente, serviços reiniciáveis).

Acione uma execução única no painel pelo card **Doctor** ou `wall-vault doctor`.

---

## Hooks

Execute um comando shell em eventos chave:

```yaml
hooks:
  on_model_change:   "logger 'wall-vault: $SERVICE/$MODEL'"
  on_key_exhausted:  "notify-send 'wall-vault' '$SERVICE keys all on cooldown'"
  on_service_down:   "/usr/local/bin/page-oncall.sh $SERVICE '$ERROR'"
  on_doctor_fix:     "echo \"$AGENT: $LEVEL $MSG\" >> ~/wall-vault.audit.log"
  openclaw_socket:   ""    # if set, OpenClaw TUI receives events over this Unix socket
```

Cada hook recebe variáveis de ambiente específicas do evento (`SERVICE`, `MODEL`, `ERROR`, `AGENT`, `LEVEL`, `MSG`). Hooks rodam de forma assíncrona com um timeout de 5 segundos — o proxy nunca bloqueia em um hook lento.

---

## Variáveis de ambiente

| Variável | Campo YAML |
|----------|------------|
| `WV_LANG` | `lang` |
| `WV_THEME` | `theme` |
| `WV_PROXY_PORT` | `proxy.port` |
| `WV_PROXY_HOST` | `proxy.host` |
| `WV_VAULT_PORT` | `vault.port` |
| `WV_VAULT_HOST` | `vault.host` |
| `WV_VAULT_URL` | `proxy.vault_url` (distributed) |
| `WV_VAULT_TOKEN` | `proxy.vault_token` |
| `WV_ADMIN_TOKEN` | `vault.admin_token` |
| `WV_MASTER_PASS` | `vault.master_password` |
| `WV_AVATAR` | `proxy.avatar` |
| `WV_TOOL_FILTER` | `proxy.tool_filter` |
| `WV_CC_CLIENT_ID` | `proxy.claude_code_client_id` |
| `WV_PROXY_TLS_ENABLED` | `proxy.tls.enabled` |
| `WV_PROXY_TLS_CERT` | `proxy.tls.cert_file` |
| `WV_PROXY_TLS_KEY` | `proxy.tls.key_file` |
| `WV_PROXY_TLS_REQUIRED` | `proxy.tls.required` (recusa iniciar com TLS desativado — bloqueia o fallback em texto plano) |
| `WV_PROXY_ALLOW_CIDRS` | `proxy.allow_cidrs` (lista separada por vírgulas, p. ex. `192.168.0.0/16,10.0.0.0/8`; loopback sempre passa) |
| `WV_VAULT_TLS_ENABLED` | `vault.tls.enabled` |
| `WV_VAULT_TLS_CERT` | `vault.tls.cert_file` |
| `WV_VAULT_TLS_KEY` | `vault.tls.key_file` |
| `WV_VAULT_BOOTSTRAP_PORT` | `vault.bootstrap_port` |
| `WV_PROXY_PLAIN_PORT` | `proxy.plain_port` |
| `WV_KEY_GOOGLE` | Importação única: chaves Google separadas por vírgula |
| `WV_KEY_OPENROUTER` | Importação única: chaves OpenRouter |
| `WV_KEY_ANTHROPIC` | Importação única: chaves Anthropic |
| `WV_KEY_OPENAI` | Importação única: chaves OpenAI |
| `WV_OLLAMA_URL` | Substituição de URL Ollama por host (instância única) |
| `WV_OLLAMA_URLS` | URLs Ollama separadas por vírgula (dispatch multi-instância) |
| `WV_OLLAMA_KEEP_ALIVE` | `proxy.ollama_keep_alive` |
| `WV_OLLAMA_NUM_CTX` | `proxy.ollama_num_ctx` |
| `WV_LMSTUDIO_URL`, `WV_VLLM_URL`, `WV_LLAMACPP_URL` | Substituição de URL por backend (instância única) |
| `WV_TOKEN_SENTINEL_FALLBACK` | `proxy.token_sentinel_fallback` |
| `WV_OAI_STREAM_FORWARD` | `proxy.oai_stream_forward` |
| `WV_ANTHROPIC_FALLBACK_MODEL` | `proxy.anthropic_fallback_model` |
| `WV_INJECT_MODEL_IDENTITY` | `proxy.inject_model_identity` (guarda de identidade por mensagem de sistema, desligado por padrão) |
| `WV_PROMPT_TOKEN_CAP` | Limite de auto-truncamento por host para prompts OAI-compat locais (inteiro positivo = ativar, 0 = off) |
| `WV_DISPATCH_TRACE` | Definir `1` para registrar serviço/modelo resolvido e razão de cada dispatch (off por padrão) |
| `WV_ECONOWORLD_MAX_TOKENS` | `proxy.econoworld_max_tokens` |
| `WV_ECONOWORLD_STREAM` | `proxy.econoworld_stream` |
| `WV_ECONOWORLD_REQUEST_TIMEOUT` | `proxy.econoworld_request_timeout` |

Cada variável de ambiente, quando definida, prevalece sobre o arquivo YAML.

---

## Solução de problemas

### `connection refused` em `:56244`

Ou o proxy não está rodando ou está vinculado a um host diferente. Verifique:

```bash
ss -lnp | grep 56244
systemctl --user status wall-vault-proxy   # Linux
launchctl list | grep wall-vault           # macOS
```

Se ele estiver rodando em uma porta diferente, sua configuração tem `proxy.port` substituído — verifique `~/.wall-vault/config.yaml`.

### `x509: certificate signed by unknown authority`

O cliente não confia na CA interna do wall-vault. Execute `wall-vault cert install-trust` na máquina cliente. Para agentes cujo runtime ignora o trust store do SO (por exemplo, Node com um `NODE_EXTRA_CA_CERTS` codificado), use o companheiro HTTP em loopback em `127.0.0.1:56245` (apenas mesmo host) ou defina `WV_PROXY_TLS_ENABLED=0` para retornar para HTTP simples.

### `token not registered with vault`

O `Authorization: Bearer <token>` do cliente não corresponde a nenhum cliente registrado. Verifique o token em **Clients** no painel. Se você copiou um literal de token como `proxy-managed`, `dummy` ou `""` de uma configuração obsoleta, substitua-o pelo token de cliente real.

### `Anthropic dispatch needs a Claude model id`

Comportamento padrão a partir da v0.2.63: um id de modelo não-Claude enviado ao despacho anthropic retorna um erro. Ou corrija o roteamento (não envie `gemini-2.5-flash` para anthropic) ou habilite a reescrita automática via `proxy.anthropic_fallback_model`.

### `unknown service: <id>`

O despacho viu um id de serviço que nenhum plugin yaml reivindicou. Verifique:

```bash
ls ~/.wall-vault/services/        # any plugin yaml present?
cat ~/.wall-vault/services/<id>.yaml | grep enabled
```

Se o yaml existe mas é `enabled: false`, ative-o. Se está totalmente ausente, copie de `configs/services/` na árvore de fontes.

### Resposta vazia em um modelo de raciocínio

`qwen3.6`, `deepseek-r1` e a família GPT-`o1` às vezes emitem apenas `reasoning_content` e deixam `content` vazio. A partir da v0.2.63 o wall-vault recorre automaticamente ao texto de raciocínio — se você ainda vê respostas vazias, o backend não está retornando nenhum dos campos. Verifique os logs do upstream.

Para LM Studio com qwen3 especificamente, defina `inline_no_think_for_qwen3: true` no plugin yaml para que o raciocínio seja desativado inline. Os arquivos lmstudio.yaml e ollama.yaml integrados já fazem isso.

### O painel mostra "todas as chaves em cooldown" mas eu acabei de adicionar uma

A nova chave está saudável mas o caminho de despacho ainda pode estar no cooldown de uma chave mais antiga. Tente uma nova requisição — o proxy faz round-robin por chamada, e uma chave saudável será escolhida em seguida.

### O cofre não abre com a senha mestra

Senha errada. Não há recuperação — o wall-vault deliberadamente não inclui um backdoor. Se você realmente perdeu a senha mestra, o único caminho é deletar `~/.wall-vault/data/vault.json`, reiniciar com uma nova senha e adicionar as chaves novamente.

### Limites do free-tier OpenRouter atingidos

Defina `proxy.services` para incluir `openrouter` e adicione pelo menos uma chave OpenRouter. O proxy automaticamente recorre de um modelo pago para sua variante `:free` quando o caminho pago retorna 402 / 429.

### `journalctl --user -u wall-vault-proxy` está vazio

Os logs systemd `--user` vão para o journal do usuário que o executa. Se você iniciou a unit como `root` ou via `sudo`, o journal está na instância do sistema — tente `journalctl -u wall-vault-proxy` sem `--user`.

---

## Mais

- Referência da API HTTP — veja [API.md](API.md)
- Código fonte — `https://github.com/sookmook/wall-vault`
- Relatórios de bug / pedidos de feature — GitHub Issues
- Histórico de releases — [CHANGELOG.md](../CHANGELOG.md)
