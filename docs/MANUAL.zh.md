# wall-vault 用户手册

[English](MANUAL.md) · [한국어](MANUAL.ko.md) · **中文** · [日本語](MANUAL.ja.md) · [Español](MANUAL.es.md) · [Français](MANUAL.fr.md) · [Deutsch](MANUAL.de.md) · [Português](MANUAL.pt.md) · [العربية](MANUAL.ar.md) · [हिन्दी](MANUAL.hi.md) · [Bahasa Indonesia](MANUAL.id.md) · [ภาษาไทย](MANUAL.th.md) · [Kiswahili](MANUAL.sw.md) · [Hausa](MANUAL.ha.md) · [नेपाली](MANUAL.ne.md) · [Монгол](MANUAL.mn.md) · [isiZulu](MANUAL.zu.md)

本手册涵盖 wall-vault 的安装、配置和操作。如需概览,请参阅 [README](../README.md)。如需 HTTP API 详细信息,请参阅 [API 参考](API.md)。

## 目录

1. [wall-vault 的功能](#wall-vault-的功能)
2. [安装](#安装)
3. [使用安装向导首次运行](#使用安装向导首次运行)
4. [启用 TLS](#启用-tls)
5. [注册 API 密钥](#注册-api-密钥)
6. [连接代理](#连接代理)
7. [仪表板](#仪表板)
8. [分布式模式](#分布式模式)
9. [自启动](#自启动)
10. [插件 yaml](#插件-yaml)
11. [Doctor](#doctor)
12. [Hooks](#hooks)
13. [环境变量](#环境变量)
14. [故障排查](#故障排查)

---

## wall-vault 的功能

wall-vault 是一个单一 Go 二进制文件,捆绑了两个协同运行的服务:

- **vault** 在静态状态下加密存储 API 密钥(使用主密码进行 AES-GCM 加密),按密钥跟踪使用情况和冷却时间,通过 Server-Sent Events(SSE)广播变更,并在 `:56243` 上为操作员提供 Web 仪表板。
- **proxy** 在 `:56244` 上公开 Gemini、Anthropic、OpenAI 兼容和 Ollama 原生端点。任何指向 proxy 的 AI 客户端都使用 vault 中的密钥——客户端永远看不到这些密钥。当一个上游失败时,调度会按顺序回退到下一个提供商。

这在以下情况下非常有用:

- 你拥有多个提供商的密钥,并希望代理只与一个 URL 通信。
- 你希望免费层级密钥在冷却期间退场,而不打断会话。
- 你希望在同一 LAN 上的多个机器人、IDE 或脚本使用相同的密钥,而无需复制凭证。
- 你希望使用仪表板而非环境变量来编辑密钥和切换模型。
- 你希望在云端额度用尽时使用本地回退(Ollama、LM Studio、vLLM)。

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

## 安装

### Linux / macOS 一行命令

```bash
curl -fsSL https://raw.githubusercontent.com/sookmook/wall-vault/main/install.sh | sh
```

该脚本自动检测 OS 和架构,将正确的二进制文件下载到 `~/.local/bin/wall-vault`,并使其可执行。如果 `~/.local/bin` 不在你的 `PATH` 中,请添加:

```bash
export PATH="$HOME/.local/bin:$PATH"
```

### 手动下载

每次发布都会在 `https://github.com/sookmook/wall-vault/releases` 上发布预构建的二进制文件。

```bash
# Linux amd64
curl -L https://github.com/sookmook/wall-vault/releases/latest/download/wall-vault-linux-amd64 \
  -o wall-vault && chmod +x wall-vault

# Linux arm64 (Raspberry Pi、ARM 服务器)
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

### 从源码构建

需要 Go 1.25 或更新版本。

```bash
git clone https://github.com/sookmook/wall-vault
cd wall-vault
go build -o wall-vault .
```

`make build-all` 会交叉编译到所有五个支持的平台。二进制文件会生成在 `bin/` 中。

---

## 使用安装向导首次运行

```bash
wall-vault setup
```

向导会按顺序提示你以下内容:

1. **语言** — 从 17 个 UI 语言中选择一个。会自动从 `$LANG` 检测;但向导依然提供列表。
2. **主题** — `light`(默认)、`dark`、`cherry`、`ocean`、`gold`、`autumn`、`winter`。仅外观影响。
3. **模式** — `standalone`(单主机,默认)或 `distributed`(vault 在一个主机,proxy 在其他主机)。
4. **机器人名称** — 自由形式的 `client_id` 标识。vault 用其来限定每个客户端的配置(模型覆盖、回退链)。
5. **proxy 端口** — 默认 `56244`。
6. **vault 端口** — 默认 `56243`(仅 standalone)。
7. **服务选择** — 对以下每项给出 y/N:Google Gemini、OpenRouter、Anthropic、OpenAI、Ollama、LM Studio、vLLM。可多选;每一项都会在末尾写入其环境变量提示。
8. **工具过滤器** — `strip_all`(默认;出于安全考虑屏蔽所有传入工具定义)或 `passthrough`(允许任何工具通过)。
9. **管理令牌** — 留空以自动生成。仪表板需要此令牌登录。
10. **主密码** — 留空表示不加密(不推荐);设置一个值则会用 AES-GCM 在静态状态下加密密钥存储。
11. **保存路径** — 默认是当前目录中的 `wall-vault.yaml`。加载器也会检查 `~/.wall-vault/config.yaml`。

保存后,向导会运行 `doctor.FixTrust`,以便任何本地安装的代理(OpenClaw、Claude Code、Cline)将 wall-vault 内部 CA 自动添加到其信任存储。如果未安装此类代理,该步骤会显示 `SKIP`,不写入任何内容。

然后启动二进制文件:

```bash
wall-vault start
```

`start` 在一个进程中同时运行 vault 和 proxy(standalone 模式)。对于 distributed 模式,在 vault 主机上使用 `wall-vault vault`,在每个 proxy 主机上使用 `wall-vault proxy`。

在浏览器中打开 `http://localhost:56243`。使用向导打印的管理令牌登录。

---

## 启用 TLS

向导的默认设置将两个监听器都保留为纯 HTTP。大多数代理(OpenClaw、Claude Code、Cursor)在使用单一 HTTPS 端点时表现更好,因此在跨越本地机器以外的任何部署中都建议使用 TLS。

wall-vault 自带内部 CA,因此你不需要公共 DNS 名称或 Let's Encrypt。

```bash
# 1. 创建内部 CA — 写入 ~/.wall-vault/ca.{crt,key}。
#    CA 默认 10 年有效;使用 --ca-years 覆盖。
wall-vault cert init

# 2. 颁发主机证书。Subject Alternative Names 自动包含:
#       hostname、"localhost"、"127.0.0.1" 以及检测到的任何非环回 LAN IP。
#    使用 --dir 覆盖颁发目录,使用 --host-years 覆盖有效期。
wall-vault cert issue $(hostname)

# 3. 在本机的 OS 密钥环中信任 CA。
#    Linux:通过 update-ca-certificates 写入 /etc/ssl/certs/(需要 sudo)。
#    macOS:通过 security add-trusted-cert 添加到 System 密钥环(需要 sudo)。
#    Windows:通过 certutil 导入到 CurrentUser\Root(无需管理员权限)。
wall-vault cert install-trust

# 4. 在两个监听器上启用 TLS。
export WV_PROXY_TLS_ENABLED=1
export WV_PROXY_TLS_CERT="$HOME/.wall-vault/$(hostname).crt"
export WV_PROXY_TLS_KEY="$HOME/.wall-vault/$(hostname).key"
export WV_VAULT_TLS_ENABLED=1
export WV_VAULT_TLS_CERT="$HOME/.wall-vault/$(hostname).crt"
export WV_VAULT_TLS_KEY="$HOME/.wall-vault/$(hostname).key"

wall-vault start
```

要将信任扩展到其他 LAN 机器,复制 `~/.wall-vault/ca.crt`,在每台机器上运行 `wall-vault cert install-trust --ca <path>`。vault 还会通过 `:56247`(**bootstrap 端口**)上的小型纯 HTTP 监听器公开 `ca.crt`,以应对全新客户端需要 CA 才能进行 HTTPS 通信的死锁情况。

### 环回 HTTP 伴侣

某些代理——特别是 OpenClaw 自带的 Node 运行时——会在进程启动时重写 `NODE_EXTRA_CA_CERTS`,丢弃任何由操作员提供的 CA 提示。即使在 `cert install-trust` 之后,它们也无法在守护进程内部识别 wall-vault CA。wall-vault 通过在启用 TLS 时绑定一个额外的 **仅环回纯 HTTP 监听器** `127.0.0.1:56245` 来解决此问题。同主机的客户端通过该端口完全不使用 TLS 即可连接 proxy;LAN 客户端继续使用 TLS 监听器。

如果不需要,使用 `WV_PROXY_PLAIN_PORT=0` 禁用。

### `wall-vault cert list`

显示 `~/.wall-vault/` 下所有证书的 subject、有效期窗口和 SAN。

```
$ wall-vault cert list
ca.crt          subject=wall-vault internal CA   not-after=2036-05-05
hostname.crt    subject=hostname                 not-after=2031-05-05   SAN=hostname,localhost,127.0.0.1,192.168.…
```

---

## 注册 API 密钥

有两种方式:仪表板,或环境变量。

### 仪表板(推荐)

1. 使用管理令牌登录到 `https://localhost:56243`。
2. 在密钥卡中点击 **+ API key**。
3. 选择服务(Google、OpenRouter、Anthropic、OpenAI 等)。
4. 粘贴密钥。保存。

每个服务允许多个密钥;proxy 会在它们之间轮询,并跳过命中按密钥冷却的密钥。

### 环境变量(一次性 bootstrap)

```bash
export WV_KEY_GOOGLE="AIzaSyA1...,AIzaSyB2...,AIzaSyC3..."   # 逗号分隔
export WV_KEY_OPENROUTER="sk-or-v1-…"
export WV_KEY_ANTHROPIC="sk-ant-…"
export WV_KEY_OPENAI="sk-…"
wall-vault start
```

通过这种方式提供的密钥会在首次启动时写入加密存储。后续启动会从磁盘读取它们;首次运行后可以取消设置环境变量。

### 冷却和轮换

每次成功调用都会增加该密钥的 `usage_count` 并刷新 `last_used`。在 HTTP 429 / 402 / 403 时,proxy 将该密钥置于 **冷却** 状态(默认值:429 为 60 分钟,402 为 24 小时,403 为 12 小时)。下一次调度会为该服务选择不同的密钥。当某个服务的所有密钥都处于冷却状态时,proxy 会快速跳过该服务,并尝试回退链中的下一个提供商。

冷却时间会在仪表板中按密钥显示,带倒计时。

---

## 连接代理

### OpenClaw

OpenClaw 是最初的目标客户端。使用仪表板的 **+ Add agent** 模态框:

- 将 **Agent type** 设置为 `openclaw` 或 `nanoclaw`。
- 设置 **Work directory** — 对于 OpenClaw 会自动填充为 `~/.openclaw`。
- 选择一个 **preferred service**,可选地选择 **model override**。
- 点击 **Apply**。wall-vault 会直接写入 `~/.openclaw/openclaw.json`(提供商 URL、vault 令牌、模型条目)。

当你从仪表板更改模型时,OpenClaw 会通过 SSE 在 1-3 秒内接收变更——无需重启。

### Claude Code

```bash
export ANTHROPIC_BASE_URL=https://localhost:56244
export ANTHROPIC_API_KEY=<your-vault-client-token>
claude
```

当上游 Anthropic 信用额度用尽时,调度会回退到该客户端的 `fallback_services` 中列出的服务。默认情况下,发送到 anthropic 调度的非 Claude 模型 ID 会返回错误,使错误路由立即显现。要选择启用自动重写:

```yaml
proxy:
  anthropic_fallback_model: "claude-haiku-4-5-20251001"
```

### Cursor

在 Cursor 的 **Settings → AI → OpenAI API** 中:

```
Base URL:  https://localhost:56244
API Key:   <your-vault-client-token>
Model:     gemini-2.5-flash    # 或 wall-vault 已知的任何模型
```

### Continue (VS Code、JetBrains)

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

### 自定义 HTTP

```bash
curl -X POST https://localhost:56244/v1/chat/completions \
  -H "Authorization: Bearer <your-vault-client-token>" \
  -H "Content-Type: application/json" \
  -d '{
    "model": "gemini-2.5-flash",
    "messages": [{"role": "user", "content": "hello"}]
  }'
```

当设置了 `proxy.oai_stream_forward: true` 时,同一端点也接受流式传输(`"stream": true`)。

---

## 仪表板

`https://localhost:56243`。主页网格上有五张卡片:

- **Keys** — 按服务分组的所有 API 密钥。添加、编辑、删除;查看用量和冷却。
- **Services** — Google / OpenRouter / Anthropic / OpenAI / Ollama / LM Studio / vLLM / llama.cpp,加上 `~/.wall-vault/services/` 中的任何插件 yaml。设置每个服务的 `default_model`、`allowed_models`、基础 URL 和 reasoning 开关。
- **Clients (agents)** — 每个已注册的客户端(OpenClaw 机器人、Claude Code 会话、Cursor 实例等)。分配首选服务、模型覆盖和回退链。
- **Proxies** — 每个对该 vault 进行了认证的 proxy。实时状态(在线/离线)、最后可见时间、当前模型。
- **Settings** — 管理令牌、主密码轮换、主题、语言。

每张卡片都有一个编辑滑出框(右侧)。点击外部或按 `Esc` 关闭。变更通过 SSE 在数秒内推送到所有已连接的 proxy。

**页脚** 包含一个 SSE 指示器(绿色 = 已连接、橙色 = 重连中、灰色 = 已断开)和实时构建版本。

---

## 分布式模式

当你有多台机器都需要相同的密钥时,在一台主机上运行 vault,在其余每台主机上运行 proxy。

### vault 主机

```bash
WV_VAULT_HOST=0.0.0.0 \
WV_ADMIN_TOKEN=<admin> \
WV_MASTER_PASS=<master> \
wall-vault vault
```

仪表板现在可以从 `https://<vault-host>:56243` 访问。在 **Clients** 卡中为每个远程 proxy 添加一个代理;每个代理都会铸造一个唯一的 `vault_token`。

### proxy 主机

```bash
WV_VAULT_URL=http://<vault-host>:56243 \
WV_VAULT_TOKEN=<that-client-token> \
WV_PROXY_HOST=0.0.0.0 \
wall-vault proxy
```

proxy 会对 vault 进行认证、打开 SSE 流,并应用接收到的任何配置(首选服务、模型覆盖、回退链)。后续的 vault 编辑会在数秒内生效,无需重启。

对于跨 LAN 的安装,在 vault 主机上启用 TLS(`WV_VAULT_TLS_ENABLED=1` + cert/key 环境变量),并通过相同的 `wall-vault cert install-trust` 步骤运行每个 proxy 主机,以便 proxy 对 vault 的 HTTPS 调用受到信任。

---

## 自启动

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
loginctl enable-linger $USER       # 注销后单元继续运行
```

对于同一主机上的 vault,编写一个并行的 `wall-vault-vault.service`。对于 standalone 模式,一个调用 `wall-vault start` 的单元就足够了。

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

使用 `nssm` 将 `wall-vault.exe start` 包装为 Windows 服务,或使用一个在用户登录时运行的 `schtasks` 条目。

---

## 插件 yaml

任何 OpenAI 兼容的后端都可以通过将 yaml 放入 `~/.wall-vault/services/` 来添加,无需更改代码。wall-vault 在启动时加载它,并将该服务注册到调度、OAI 兼容检测集和 Gemini 流桥接。

```yaml
# ~/.wall-vault/services/llamacpp.yaml
id: llamacpp                 # 唯一服务 ID
name: llama.cpp              # 人类可读标签
enabled: true                # 已禁用的插件在加载时被跳过

default_url: http://localhost:8080   # 操作员覆盖;env 优先 (WV_LLAMACPP_URL)
endpoints:
  generate: /v1/chat/completions
  list_models: /v1/models

auth:
  type: none                 # none | bearer | query_param | header
  param: ""                  # 对于 query_param:参数名 (例如 "key")

request_format: openai       # openai | gemini | ollama | raw

model_fetch:
  enabled: true              # 让仪表板自动检测模型
  dynamic: true              # 仪表板每次打开时重新获取
  auto_detect_url: true      # 即使未声明也尝试 /v1/models

concurrency:
  max: 1                     # 此后端的最大并发请求数
  queue_size: 10
  wait_notify: true          # 向 TUI 代理显示 "queued" 提示

error_codes:
  503:
    cooldown: 5m
    message: "llama.cpp not responding"

# 在 reasoning 关闭时为 qwen3 系列启用内联 /no_think 指令。
# 如果你的后端的聊天模板会去除该标记 (LM Studio 的 jinja、Ollama 的 /v1
# 层),则设置为 true。其他后端通常会原样回显字面文本,因此每个 yaml 默认
# 保持 opt-in。
inline_no_think_for_qwen3: false

# Hub 拓扑 — 指向另一个 wall-vault。当此插件作为远程 wall-vault 的前端时
# 必需 (这样接收方 wall-vault 可以看到 publisher 前缀并正确路由),并且使
# proxy.vault_token 中的 bearer token 作为 Authorization 发送。
preserve_model_id: false
tls_internal_ca: false       # 将 ~/.wall-vault/ca.crt 添加到客户端信任池
```

`configs/services/` 中捆绑的集合(lmstudio、vllm、llamacpp、tgwui、localai、jan、koboldcpp、tabbyapi、mlx-server、litellm-proxy、ollama、google、openrouter)默认禁用。将你想要的复制到 `~/.wall-vault/services/`,设置 `enabled: true`,然后重启。

---

## Doctor

`wall-vault doctor` 对整个安装运行一次性健康探测:

```
✓ vault listener  (https://localhost:56243)
✓ proxy listener  (https://localhost:56244)
✓ master password set
⚠ Google: 2 keys, all on cooldown
✓ Anthropic: 1 key healthy
✗ Ollama: not reachable at http://localhost:11434
```

每行是以下之一:

- `✓` — 健康
- `⚠` — 性能下降但功能正常(一个密钥处于冷却、配额低等)
- `✗` — 已损坏
- `SKIP` — 未配置 / 在此主机上不适用

第二种守护进程模式每隔 `doctor.interval`(默认 5 分钟)运行相同的探测,并将结果写入 `doctor.log_file`(默认 `/tmp/wall-vault-doctor.log`)。当 `doctor.auto_fix` 为 true 时,它还会尝试修复常见的偏差(过时的 OpenClaw 配置、缺失的 TLS 信任、可重启的服务)。

通过仪表板的 **Doctor** 卡片或 `wall-vault doctor` 触发一次性运行。

---

## Hooks

在关键事件上运行 shell 命令:

```yaml
hooks:
  on_model_change:   "logger 'wall-vault: $SERVICE/$MODEL'"
  on_key_exhausted:  "notify-send 'wall-vault' '$SERVICE keys all on cooldown'"
  on_service_down:   "/usr/local/bin/page-oncall.sh $SERVICE '$ERROR'"
  on_doctor_fix:     "echo \"$AGENT: $LEVEL $MSG\" >> ~/wall-vault.audit.log"
  openclaw_socket:   ""    # 如果设置,OpenClaw TUI 通过此 Unix 套接字接收事件
```

每个 hook 都会获得事件特定的环境变量(`SERVICE`、`MODEL`、`ERROR`、`AGENT`、`LEVEL`、`MSG`)。Hook 以 5 秒超时异步运行——proxy 永远不会因慢速 hook 而阻塞。

---

## 环境变量

| 变量 | YAML 字段 |
|------|------------|
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
| `WV_PROXY_TLS_REQUIRED` | `proxy.tls.required` (TLS 关闭时拒绝启动 — 防止明文回退) |
| `WV_PROXY_ALLOW_CIDRS` | `proxy.allow_cidrs` (逗号分隔列表, 例如 `192.168.0.0/16,10.0.0.0/8`; 回环地址始终通过) |
| `WV_VAULT_TLS_ENABLED` | `vault.tls.enabled` |
| `WV_VAULT_TLS_CERT` | `vault.tls.cert_file` |
| `WV_VAULT_TLS_KEY` | `vault.tls.key_file` |
| `WV_VAULT_BOOTSTRAP_PORT` | `vault.bootstrap_port` |
| `WV_PROXY_PLAIN_PORT` | `proxy.plain_port` |
| `WV_KEY_GOOGLE` | 一次性导入:逗号分隔的 Google 密钥 |
| `WV_KEY_OPENROUTER` | 一次性导入:OpenRouter 密钥 |
| `WV_KEY_ANTHROPIC` | 一次性导入:Anthropic 密钥 |
| `WV_KEY_OPENAI` | 一次性导入:OpenAI 密钥 |
| `WV_OLLAMA_URL` | 每主机 Ollama URL 覆盖（单实例） |
| `WV_OLLAMA_URLS` | 逗号分隔的 Ollama URL 列表（多实例分发） |
| `WV_OLLAMA_KEEP_ALIVE` | `proxy.ollama_keep_alive` |
| `WV_OLLAMA_NUM_CTX` | `proxy.ollama_num_ctx` |
| `WV_LMSTUDIO_URL`, `WV_VLLM_URL`, `WV_LLAMACPP_URL` | 每后端 URL 覆盖（单实例） |
| `WV_TOKEN_SENTINEL_FALLBACK` | `proxy.token_sentinel_fallback` |
| `WV_OAI_STREAM_FORWARD` | `proxy.oai_stream_forward` |
| `WV_ANTHROPIC_FALLBACK_MODEL` | `proxy.anthropic_fallback_model` |
| `WV_INJECT_MODEL_IDENTITY` | `proxy.inject_model_identity`（系统消息身份守卫，默认关闭） |
| `WV_PROMPT_TOKEN_CAP` | 每主机本地 OAI 兼容提示自动截断阈值（正整数 = 启用，0 = 关闭） |
| `WV_DISPATCH_TRACE` | 设为 `1` 以记录每个 dispatch 的解析服务/模型和原因（默认关闭） |
| `WV_ECONOWORLD_MAX_TOKENS` | `proxy.econoworld_max_tokens` |
| `WV_ECONOWORLD_STREAM` | `proxy.econoworld_stream` |
| `WV_ECONOWORLD_REQUEST_TIMEOUT` | `proxy.econoworld_request_timeout` |

设置的每个环境变量都会胜过 YAML 文件。

---

## 故障排查

### `:56244` 上的 `connection refused`

要么 proxy 未运行,要么它绑定到了不同的主机。请检查:

```bash
ss -lnp | grep 56244
systemctl --user status wall-vault-proxy   # Linux
launchctl list | grep wall-vault           # macOS
```

如果它在不同的端口上运行,你的配置中 `proxy.port` 被覆盖了——请检查 `~/.wall-vault/config.yaml`。

### `x509: certificate signed by unknown authority`

客户端不信任 wall-vault 内部 CA。在客户端机器上运行 `wall-vault cert install-trust`。对于运行时忽略 OS 信任存储的代理(例如带有硬编码 `NODE_EXTRA_CA_CERTS` 的 Node),请使用 `127.0.0.1:56245` 上的环回 HTTP 伴侣(仅同主机),或设置 `WV_PROXY_TLS_ENABLED=0` 以回退到纯 HTTP。

### `token not registered with vault`

客户端的 `Authorization: Bearer <token>` 与任何已注册的客户端都不匹配。请在仪表板的 **Clients** 下验证令牌。如果你从过时的配置中复制了像 `proxy-managed`、`dummy` 或 `""` 这样的字面令牌,请用真正的客户端令牌替换它。

### `Anthropic dispatch needs a Claude model id`

v0.2.63 起的默认行为:发送到 anthropic 调度的非 Claude 模型 ID 会返回错误。要么修复路由(不要将 `gemini-2.5-flash` 发送到 anthropic),要么通过 `proxy.anthropic_fallback_model` 选择启用自动重写。

### `unknown service: <id>`

调度看到了一个没有插件 yaml 声明的服务 ID。请检查:

```bash
ls ~/.wall-vault/services/        # 是否存在任何插件 yaml?
cat ~/.wall-vault/services/<id>.yaml | grep enabled
```

如果 yaml 存在但 `enabled: false`,请翻转它。如果完全缺失,请从源码树中的 `configs/services/` 复制。

### reasoning 模型上响应为空

`qwen3.6`、`deepseek-r1` 和 GPT-`o1` 系列有时只发出 `reasoning_content` 而留空 `content`。从 v0.2.63 起,wall-vault 会自动回退到 reasoning 文本——如果你仍然看到空响应,则后端两个字段都没有返回。检查上游的日志。

特别是对于使用 qwen3 的 LM Studio,请在插件 yaml 中设置 `inline_no_think_for_qwen3: true`,以便内联禁用 reasoning。内置的 lmstudio.yaml 和 ollama.yaml 已经这样做了。

### 仪表板显示"所有密钥都在冷却",但我刚刚添加了一个

新密钥是健康的,但调度路径可能仍处于较旧密钥的冷却中。请尝试一个新请求——proxy 按调用轮询,接下来会选择一个健康的密钥。

### vault 无法用主密码解锁

密码错误。没有恢复方法——wall-vault 故意不附带后门。如果你确实丢失了主密码,唯一的方法是删除 `~/.wall-vault/data/vault.json`,使用新密码重新启动,然后重新添加密钥。

### 达到了免费层级 OpenRouter 限额

将 `proxy.services` 设置为包含 `openrouter`,并至少添加一个 OpenRouter 密钥。当付费路径返回 402 / 429 时,proxy 会自动从付费模型回退到其 `:free` 变体。

### `journalctl --user -u wall-vault-proxy` 为空

systemd 的 `--user` 日志会发送到运行它的用户的日志中。如果你以 `root` 或通过 `sudo` 启动该单元,日志会在系统实例中——请尝试不带 `--user` 的 `journalctl -u wall-vault-proxy`。

---

## 更多

- HTTP API 参考 — 请参阅 [API.md](API.md)
- 源码 — `https://github.com/sookmook/wall-vault`
- Bug 报告 / 功能请求 — GitHub Issues
- 发布历史 — [CHANGELOG.md](../CHANGELOG.md)
