# wall-vault

> **API 密钥保险库 + AI 代理,合二为一的 Go 单文件程序。**
> 使用 AES-GCM 在本地保存密钥,在多家服务商之间轮换,在某一家失败时自动回退,并自带实时仪表盘。

[![License: GPL v3](https://img.shields.io/badge/License-GPLv3-blue.svg)](LICENSE)
[![Go](https://img.shields.io/badge/Go-1.25-00ADD8.svg)](go.mod)

[English](README.md) · [한국어](README.ko.md) · **中文** · [日本語](README.ja.md) · [Español](README.es.md) · [Français](README.fr.md) · [Deutsch](README.de.md) · [Português](README.pt.md) · [العربية](README.ar.md) · [हिन्दी](README.hi.md) · [Bahasa Indonesia](README.id.md) · [ภาษาไทย](README.th.md) · [Kiswahili](README.sw.md) · [Hausa](README.ha.md) · [नेपाली](README.ne.md) · [Монгол](README.mn.md) · [isiZulu](README.zu.md)

---

## 这是什么

wall-vault 介于 AI 代理(OpenClaw、Claude Code、Cursor、Continue,或你自己写的脚本)和它要调用的云端或本地 AI 服务商之间。一个二进制里有两件事。

- **Vault** — 把 API 密钥静态加密保存(由主密码派生的 AES-GCM),进行轮换,记录每把密钥的用量和冷却时间,通过 SSE 广播变更,并在 `:56243` 提供 Web 仪表盘。
- **Proxy** — 在 `:56244` 暴露 Gemini、Anthropic 和 OpenAI 兼容端点,从保险库选取一把密钥,转发到你配置的上游,在某一家失败时回退到下一家服务商。

它支持四种请求格式(Gemini `:generateContent`、Anthropic `/v1/messages`、OpenAI `/v1/chat/completions`、Ollama 原生 `/api/chat`)和五类上游。

| 服务商 | 说明 |
|----------|-------|
| Google Gemini | 原生 API,按项目轮换密钥 |
| Anthropic | 原生 `/v1/messages` 透传 |
| OpenAI | 原生 `/v1/chat/completions` |
| OpenRouter | 340+ 种模型,自动回退到 `:free` 变体 |
| Ollama / LM Studio / vLLM / llama.cpp / Jan / KoboldCpp / TabbyAPI / mlx-server / LiteLLM | OpenAI 兼容的本地后端,通过插件 yaml 即插即用 |

新增一个 OpenAI 兼容后端只需在 `~/.wall-vault/services/` 下放一个 yaml 文件,无需改动代码。

## 你为什么会需要它

- 你同时在用三四家 AI 服务,希望让代理只对一个 URL 说话。
- 你希望免费额度密钥在冷却时让位给下一把,而不会打断当前会话。
- 你希望同一局域网内的多个机器人 / IDE / 脚本共用同一套密钥,而不必到处复制凭据。
- 你希望用仪表盘而不是环境变量来编辑 API 密钥。
- 当云端额度用尽时,你希望有一个本地优先的备选(Ollama / LM Studio)。

## 快速开始

### 安装 (Linux / macOS)

```bash
curl -fsSL https://raw.githubusercontent.com/sookmook/wall-vault/main/install.sh | sh
```

或者直接下载预编译二进制。

```bash
# Linux amd64
curl -L https://github.com/sookmook/wall-vault/releases/latest/download/wall-vault-linux-amd64 \
  -o wall-vault && chmod +x wall-vault

# macOS Apple Silicon
curl -L https://github.com/sookmook/wall-vault/releases/latest/download/wall-vault-darwin-arm64 \
  -o wall-vault && chmod +x wall-vault

# Linux arm64 (Raspberry Pi、ARM 服务器)
curl -L https://github.com/sookmook/wall-vault/releases/latest/download/wall-vault-linux-arm64 \
  -o wall-vault && chmod +x wall-vault
```

### 安装 (Windows)

```powershell
Invoke-WebRequest -Uri `
  "https://github.com/sookmook/wall-vault/releases/latest/download/wall-vault-windows-amd64.exe" `
  -OutFile wall-vault.exe
```

### 首次运行

```bash
wall-vault setup    # 交互式向导 — 选择端口、服务、admin token、master password
wall-vault start    # 同时启动 vault 与 proxy
```

在浏览器打开 `http://localhost:56243` (启用 TLS 后则是 `https://...`,详见下文)。仪表盘会要求输入 `setup` 打印出的 admin token。从那里你可以添加 API 密钥、注册客户端,并且无需重启就能切换模型。

---

## TLS (推荐)

默认情况下 `wall-vault setup` 写出的配置不启用 TLS,因此两个监听器都以明文 HTTP 应答。本 README 的示例 URL 使用 `https://localhost:56244`,是因为大多数代理(OpenClaw、Claude Code、Cursor)更希望对接一个 TLS 前置的统一端点,这样以后把代理迁到别的主机也不会失效。要匹配这些示例,只需用内置的内部 CA 启用一次 TLS。

```bash
# 1. 创建 wall-vault 内部 CA (一次性,保存于 ~/.wall-vault/ca.{crt,key})
wall-vault cert init

# 2. 为本机签发主机证书
#    SAN 包含主机名、localhost、127.0.0.1 以及检测到的任何局域网 IP
wall-vault cert issue $(hostname)

# 3. 将 CA 信任写入本机 OS 钥匙串
wall-vault cert install-trust

# 4. 将监听器切换到 TLS
export WV_PROXY_TLS_ENABLED=1
export WV_PROXY_TLS_CERT="$HOME/.wall-vault/$(hostname).crt"
export WV_PROXY_TLS_KEY="$HOME/.wall-vault/$(hostname).key"
export WV_VAULT_TLS_ENABLED=1
export WV_VAULT_TLS_CERT="$HOME/.wall-vault/$(hostname).crt"
export WV_VAULT_TLS_KEY="$HOME/.wall-vault/$(hostname).key"
wall-vault start
```

要让局域网内的另一台机器使用,把 `~/.wall-vault/ca.crt` 拷过去,在那台机器上执行 `wall-vault cert install-trust --ca <path>`。一旦 CA 在所有机器上都被信任,网络中的每台机器都能以 `https://<host>:56244` 访问代理而不出现证书警告。

如果你宁愿继续使用明文 HTTP,保持配置不变,把后面客户端片段中的 `https://` 换成 `http://` 即可。两种模式都能工作,区别只在于哪个端口接受 TLS 握手。

**回环回退。** 同主机上无法尊重 wall-vault CA 的客户端(尤其是 OpenClaw 内置的 Node 运行时,它在 spawn 时会重写 `NODE_EXTRA_CA_CERTS`),通过 `127.0.0.1:56245` 上仅回环的明文 HTTP 伴侣端口访问代理。当 TLS 启用时,wall-vault 会自动启用它。

---

## 接入客户端

让任何 AI 客户端指向 `https://<host>:56244` (若关闭 TLS 则用 `http://...`)。代理回应四种格式。

| 格式 | 路径 | 客户端示例 |
|--------|------|-----------------|
| Gemini | `/google/v1beta/models/<model>:generateContent` | OpenClaw、Gemini CLI、Antigravity |
| Anthropic | `/v1/messages` | Claude Code、Anthropic SDK |
| OpenAI | `/v1/chat/completions` | Cursor、Continue、自定义脚本、绝大多数 LLM 应用 |
| Ollama 原生 | `/api/chat` | 透传的 Ollama 客户端 |

### Claude Code

```bash
export ANTHROPIC_BASE_URL=https://localhost:56244
export ANTHROPIC_API_KEY=<your-vault-client-token>
claude
```

当上游 Anthropic 额度用尽,转发会回退到你为该客户端在 `fallback_services` 中设置的服务商。要明确开启对非 Claude 模型的回退。

```yaml
proxy:
  anthropic_fallback_model: "claude-haiku-4-5-20251001"
```

(默认空值会让转发返回错误,以便错路由立刻浮出水面。)

### Cursor / Continue

在 Cursor 的 **Settings → AI → OpenAI API** 中:

```
Base URL:  https://localhost:56244
API Key:   <your-vault-client-token>
Model:     gemini-2.5-flash    # 或 wall-vault 已知的任意模型
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
      "apiKey": "<your-vault-client-token>"
    }
  ]
}
```

### OpenClaw

OpenClaw 是一个 TUI 代理框架,wall-vault 最初就是为了服务它而构建。仪表盘的 **Add Agent** 模态框把代理类型设为 `openclaw` (或 `nanoclaw`)后,wall-vault 会直接写入 `~/.openclaw/openclaw.json`,包含服务商 URL、保险库 token 以及模型条目。

```bash
VAULT_CLIENT_ID=my-bot \
VAULT_URL=http://<vault-host>:56243 \
VAULT_TOKEN=<your-client-token> \
wall-vault proxy
```

### curl / 脚本

```bash
curl -X POST https://localhost:56244/v1/chat/completions \
  -H "Authorization: Bearer <your-vault-client-token>" \
  -H "Content-Type: application/json" \
  -d '{
    "model": "gemini-2.5-flash",
    "messages": [{"role": "user", "content": "ping"}]
  }'
```

---

## 配置

`wall-vault setup` 会写入 `./wall-vault.yaml` 或 `~/.wall-vault/config.yaml` 之一。向导不会询问的字段请手动编辑。

```yaml
mode: standalone     # standalone | distributed
lang: en             # en | ko | zh | ja | es | fr | de | pt | ar | hi | id | th | sw | ha | ne | mn | zu
theme: light         # light | dark | cherry | ocean | gold | autumn | winter

proxy:
  port: 56244
  host: ""           # 默认: standalone 为 127.0.0.1,distributed 为 0.0.0.0
  client_id: my-bot
  vault_url: ""      # distributed: http://vault-host:56243
  vault_token: ""    # distributed: 客户端 token
  tool_filter: strip_all   # strip_all | whitelist | passthrough
  allowed_tools: []
  services: [google, openrouter, ollama]
  timeout: 300s
  tls:
    enabled: false
    cert_file: ""
    key_file: ""
  plain_port: 56245              # TLS 启用时的回环专用 HTTP 伴侣
  ollama_keep_alive: "30m"       # "-1" 永不卸载,"0" 立即卸载
  ollama_num_ctx: 8192
  oai_stream_forward: false      # 选择性启用真后端 SSE 透传
  anthropic_fallback_model: ""   # 选择性启用 anthropic 转发时改写为非 Claude 模型

vault:
  port: 56243
  host: ""
  admin_token: ""
  admin_ip_whitelist: []
  master_password: ""            # AES-GCM 密钥加密口令
  data_dir: ~/.wall-vault/data
  services_dir: ~/.wall-vault/services
  tls:
    enabled: false
    cert_file: ""
    key_file: ""
  bootstrap_port: 56247          # 仅提供 ca.crt 的明文 HTTP 监听器

doctor:
  interval: 5m
  auto_fix: true
  log_file: /tmp/wall-vault-doctor.log

hooks:
  on_model_change: ""            # shell 命令 (env: SERVICE, MODEL)
  on_key_exhausted: ""
  on_service_down: ""
  on_doctor_fix: ""
  openclaw_socket: ""
```

### 环境变量

每个 YAML 字段都有一个会覆盖文件值的环境变量。常用的几个。

| 变量 | 描述 |
|----------|-------------|
| `WV_LANG`, `WV_THEME` | 语言与主题 |
| `WV_PROXY_PORT`, `WV_PROXY_HOST` | 代理监听地址 |
| `WV_VAULT_PORT`, `WV_VAULT_HOST` | 保险库监听地址 |
| `WV_VAULT_URL`, `WV_VAULT_TOKEN` | 分布式模式端点 |
| `WV_ADMIN_TOKEN`, `WV_MASTER_PASS` | 保险库凭据 |
| `WV_KEY_GOOGLE`, `WV_KEY_OPENROUTER`, `WV_KEY_ANTHROPIC`, `WV_KEY_OPENAI` | API 密钥 (多个用逗号分隔) |
| `WV_PROXY_TLS_ENABLED`, `WV_PROXY_TLS_CERT`, `WV_PROXY_TLS_KEY` | 代理 TLS |
| `WV_VAULT_TLS_ENABLED`, `WV_VAULT_TLS_CERT`, `WV_VAULT_TLS_KEY` | 保险库 TLS |
| `WV_PROXY_PLAIN_PORT` | 回环 HTTP 伴侣 (`0` 表示禁用) |
| `WV_VAULT_BOOTSTRAP_PORT` | CA 引导监听器 (`0` 表示禁用) |
| `WV_OLLAMA_URL`, `WV_OLLAMA_KEEP_ALIVE`, `WV_OLLAMA_NUM_CTX` | Ollama 调优 |
| `WV_LMSTUDIO_URL`, `WV_VLLM_URL`, `WV_LLAMACPP_URL` | 本地后端覆盖 |
| `WV_TOKEN_SENTINEL_FALLBACK` | 回环 "proxy-managed" 哨兵替换 |
| `WV_OAI_STREAM_FORWARD` | OpenAI 兼容真后端 SSE 透传 |
| `WV_ANTHROPIC_FALLBACK_MODEL` | 选择性启用 anthropic 转发改写为非 Claude |

---

## 模式

### Standalone (默认)

保险库和代理在同一进程中运行。最适合密钥与代理在同一台主机上的场景。默认仅监听回环。

```bash
wall-vault start    # 同时运行两者
```

### Distributed

保险库运行在一台主机(**vault host**)上,保存所有密钥;其他主机上的多个代理各自用一份客户端 token 进行认证。在多台机器需要共享同一批密钥而又不想到处复制的场景中很有用。

**Vault host:**

```bash
WV_VAULT_HOST=0.0.0.0 wall-vault vault
```

**每台 proxy host:**

```bash
WV_VAULT_URL=http://<vault-host>:56243 \
WV_VAULT_TOKEN=<this-client-token> \
WV_PROXY_HOST=0.0.0.0 \
wall-vault proxy
```

仪表盘的 **Add Client** 模态框会签发 token 并注册代理类型,代理会通过 SSE 拾取配置而无需重启。

---

## 插件 yaml (即插即用后端)

任何 OpenAI 兼容后端都可以通过在 `~/.wall-vault/services/` 下放一个 yaml 文件来添加。wall-vault 会在启动时拾取并注册为可路由服务,转发逻辑、OAI 兼容检测集合以及 Gemini 流桥接都能感知到它,无需改动代码。

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
inline_no_think_for_qwen3: false   # 如果你的后端会把标记剥掉就启用
```

通过 `tls_internal_ca: true`、`auth.type: bearer` 和 `preserve_model_id: true`,可以支持 hub 拓扑(一个 wall-vault 作为另一个 wall-vault 的前置)。

---

## 从源码构建

```bash
git clone https://github.com/sookmook/wall-vault
cd wall-vault
go build -o wall-vault .
./wall-vault setup
./wall-vault start
```

为整套支持平台交叉编译。

```bash
make build-all   # linux-amd64, linux-arm64, darwin-amd64, darwin-arm64, windows-amd64
```

版本号遵循 `v{major}.{minor}.{patch}.{YYYYMMDD}.{HHmmss}`;Makefile 中的 `BASE_VERSION` 设置前缀。

### 项目结构

```
wall-vault/
├── main.go                     # CLI 分发 (start/proxy/vault/setup/cert/doctor)
├── cmd/
│   ├── setup/                  # 交互式安装向导
│   └── cert/                   # 内部 CA + 按主机 TLS 证书签发
├── internal/
│   ├── config/                 # YAML + env 加载器、插件加载器
│   ├── proxy/                  # 请求转发、密钥轮换、格式转换
│   ├── vault/                  # AES-GCM 存储、仪表盘、SSE broker
│   ├── doctor/                 # 健康探针 + 自动修复
│   ├── hooks/                  # shell 命令事件触发器
│   └── i18n/                   # 17 种语言 UI 文本
├── configs/services/           # 内置插件 yaml (lmstudio, vllm, ollama, …)
└── docs/                       # MANUAL、API 参考、16 个语言变体
```

---

## 文档

- [用户手册](docs/MANUAL.en.md) — 安装、仪表盘、代理、故障排查
- [API 参考](docs/API.en.md) — 每个端点的请求/响应格式
- [CHANGELOG](CHANGELOG.md)

---

## 技术栈

- Go 1.25,单个静态二进制
- 服务端渲染仪表盘使用 [templ](https://templ.guide),局部更新使用 [HTMX](https://htmx.org)
- 静态密钥加密使用 AES-GCM (PBKDF2 派生密钥)
- 保险库与代理之间的实时配置同步使用 Server-Sent Events
- 自签内部 CA + 按主机证书 (无需公网 DNS / Let's Encrypt)

## 许可证

GPL-3.0,详见 [LICENSE](LICENSE)。

## 贡献

欢迎提交 Pull Request,详见 [CONTRIBUTING.md](CONTRIBUTING.md)。如果是较大的改动,请先开一个 Issue 讨论设计方案。
