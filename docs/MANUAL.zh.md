# wall-vault 用户手册
*(Last updated: 2026-04-16 — v0.2.2)*

---

## 目录

1. [什么是 wall-vault？](#什么是-wall-vault)
2. [安装](#安装)
3. [初次启动（setup 向导）](#初次启动)
4. [注册 API 密钥](#注册-api-密钥)
5. [代理使用方法](#代理使用方法)
6. [密钥金库仪表盘](#密钥金库仪表盘)
7. [分布式模式（多机器人）](#分布式模式多机器人)
8. [自动启动设置](#自动启动设置)
9. [Doctor 诊断工具](#doctor-诊断工具)
10. [RTK 令牌节省](#rtk-令牌节省)
11. [环境变量参考](#环境变量参考)
12. [故障排除](#故障排除)

---

## v0.2 升级说明

- `Service` 新增了 `default_model` 和 `allowed_models` 字段。服务级别的默认模型现在直接在服务卡片上设置。
- `Client.default_service` / `default_model` 已重命名并重新解释为 `preferred_service` / `model_override`。如果 override 为空，则使用服务的默认模型。
- 首次启动 v0.2 时，现有的 `vault.json` 会自动迁移，迁移前的状态保存为 `vault.json.pre-v02.{时间戳}.bak`。
- 仪表盘已重构为三个区域：左侧边栏、中央卡片网格和右侧编辑滑动面板。
- Admin API 路径保持不变，但请求/响应正文架构已更新——旧版 CLI 脚本需要相应更新。

---

## v0.2.1 新功能

- **多模态透传（OpenAI → Gemini）**：`/v1/chat/completions` 除了 `text` 之外，现在还支持六种内容片段类型——`input_audio`、`input_video`、`input_image`、`input_file` 以及 `image_url`（支持 data URI 以及 ≤ 5 MB 的外部 http(s) URL）。代理会将每一种类型转换为 Gemini 的 `inlineData`。像 EconoWorld 这样的 OpenAI 兼容客户端可以直接流式传输音频 / 图像 / 视频二进制数据。
- **EconoWorld 代理类型**：使用 `agentType: "econoworld"` 调用 `POST /agent/apply` 会将 wall-vault 设置写入项目的 `analyzer/ai_config.json`。`workDir` 接受以逗号分隔的候选路径列表，并会把 Windows 盘符路径转换为 WSL 挂载路径。
- **仪表盘密钥网格 + CRUD**：11 把密钥以紧凑卡片形式呈现，配有 + 添加 / ✕ 删除滑动面板。
- **服务添加 + 拖拽重新排序**：服务网格新增了 + 添加按钮以及拖拽手柄（`⋮⋮`）。
- **页眉 / 页脚 / 主题动画 / 语言切换器**全部恢复。7 种主题（cherry/dark/light/ocean/gold/autumn/winter）的粒子特效在卡片下方、背景上方的图层播放。
- **滑动面板关闭体验**：点击外部区域或按 `Esc` 即可关闭滑动面板。
- **页脚 `SSE` 状态指示灯**（绿色 = 已连接，橙色 = 重连中，灰色 = 已断开）。

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

## 什么是 wall-vault？

**wall-vault = OpenClaw 专用的 AI 代理（proxy） + API 密钥金库**

使用 AI 服务需要 **API 密钥**。API 密钥就像一张**数字通行证**，用来证明"此人有资格使用该服务"。但是，这种通行证每天的使用次数有限，管理不当还有泄露的风险。

wall-vault 将这些通行证保存在安全的金库中，并在 OpenClaw 和 AI 服务之间充当**代理（proxy）**角色。简单来说，OpenClaw 只需连接 wall-vault，其余复杂的事情由 wall-vault 自动处理。

wall-vault 解决的问题：

- **API 密钥自动轮换**：当某个密钥的使用量达到上限或被临时封锁（冷却）时，会静默切换到下一个密钥。OpenClaw 不会中断，继续正常工作。
- **服务自动回退（fallback）**：如果 Google 没有响应，就切换到 OpenRouter；如果还不行，就自动切换到本地安装的 AI（Ollama、LM Studio、vLLM）。会话不会中断。当原始服务恢复后，下一个请求会自动切回（v0.1.18+，LM Studio/vLLM: v0.1.21+）。
- **实时同步（SSE）**：在金库仪表盘中更改模型后，1-3 秒内即可在 OpenClaw 界面上反映。SSE（Server-Sent Events）是服务器向客户端实时推送变化的技术。
- **实时通知**：密钥耗尽或服务故障等事件会立即显示在 OpenClaw TUI（终端界面）底部。

> :bulb: **Claude Code、Cursor、VS Code** 也可以连接使用，但 wall-vault 的本来目的是配合 OpenClaw 使用。

```
OpenClaw（TUI 终端界面）
        |
        v
  wall-vault 代理 (:56244)   <- 密钥管理、路由、回退、事件
        |
        +-- Google Gemini API
        +-- OpenRouter API（340+ 个模型）
        +-- Ollama / LM Studio / vLLM（本地机器，最后防线）
        +-- OpenAI / Anthropic API
```

---

## 安装

### Linux / macOS

打开终端，直接粘贴以下命令：

```bash
# Linux（普通 PC、服务器 — amd64）
curl -L https://github.com/sookmook/wall-vault/releases/latest/download/wall-vault-linux-amd64 \
  -o ~/.local/bin/wall-vault && chmod +x ~/.local/bin/wall-vault

# macOS Apple Silicon（M1/M2/M3 Mac）
curl -L https://github.com/sookmook/wall-vault/releases/latest/download/wall-vault-darwin-arm64 \
  -o /usr/local/bin/wall-vault && chmod +x /usr/local/bin/wall-vault
```

- `curl -L ...` — 从网络下载文件。
- `chmod +x` — 将下载的文件设为"可执行"。跳过此步骤会出现"权限不足"错误。

### Windows

打开 PowerShell（管理员权限）并执行以下命令：

```powershell
# 下载
Invoke-WebRequest -Uri `
  "https://github.com/sookmook/wall-vault/releases/latest/download/wall-vault-windows-amd64.exe" `
  -OutFile "$env:LOCALAPPDATA\Programs\wall-vault\wall-vault.exe"

# 添加到 PATH（重启 PowerShell 后生效）
$env:PATH += ";$env:LOCALAPPDATA\Programs\wall-vault"
```

> :bulb: **什么是 PATH？** 计算机查找命令的文件夹列表。添加到 PATH 后，可以从任何文件夹输入 `wall-vault` 来运行。

### 从源码构建（开发者用）

仅适用于已安装 Go 语言开发环境的情况。

```bash
git clone https://github.com/sookmook/wall-vault
cd wall-vault
make build       # bin/wall-vault（版本: v0.1.25.YYYYMMDD.HHmmss）
make install     # ~/.local/bin/wall-vault
```

> :bulb: **构建时间戳版本**：使用 `make build` 构建时，版本号会自动生成为 `v0.1.27.20260409` 这样包含日期和时间的格式。使用 `go build ./...` 直接构建时，版本号只会显示 `"dev"`。

---

## 初次启动

### 运行 setup 向导

安装完成后，请务必使用以下命令运行**设置向导**。向导会逐一询问并引导你完成必要的配置项。

```bash
wall-vault setup
```

向导的步骤如下：

```
1. 语言选择（含中文在内的 10 种语言）
2. 主题选择（light / dark / gold / cherry / ocean）
3. 运行模式 — 单机使用（standalone）还是多机共享（distributed）
4. 机器人名称 — 在仪表盘上显示的名称
5. 端口设置 — 默认值：代理 56244，金库 56243（无需更改直接回车）
6. AI 服务选择 — 从 Google / OpenRouter / Ollama / LM Studio / vLLM 中选择
7. 工具安全过滤器设置
8. 管理员令牌 — 用于锁定仪表盘管理功能的密码。支持自动生成
9. API 密钥加密密码 — 用于更安全地存储密钥（可选）
10. 配置文件保存路径
```

> :warning: **请务必记住管理员令牌。** 以后在仪表盘中添加密钥或更改设置时需要用到。如果忘记了，需要直接编辑配置文件。

向导完成后，会自动生成 `wall-vault.yaml` 配置文件。

### 启动

```bash
wall-vault start
```

以下两个服务器同时启动：

- **代理**（`http://localhost:56244`）— 连接 OpenClaw 和 AI 服务的中间人
- **密钥金库**（`http://localhost:56243`）— API 密钥管理及 Web 仪表盘

在浏览器中打开 `http://localhost:56243` 即可访问仪表盘。

---

## 注册 API 密钥

注册 API 密钥有四种方法。**推荐初次使用的用户使用方法 1（环境变量）。**

### 方法 1：环境变量（推荐 — 最简单）

环境变量是程序启动时读取的**预设值**。在终端中输入以下内容：

```bash
# 注册 Google Gemini 密钥
export WV_KEY_GOOGLE=AIzaSy...

# 注册 OpenRouter 密钥
export WV_KEY_OPENROUTER=sk-or-v1-...

# 注册后启动
wall-vault start
```

如果有多个密钥，用逗号（,）分隔。wall-vault 会自动轮流使用（轮询）：

```bash
export WV_KEY_GOOGLE=AIzaSy...,AIzaSy...,AIzaSy...
```

> :bulb: **提示**：`export` 命令仅适用于当前终端会话。要在重启后保持，请将上述行添加到 `~/.bashrc` 或 `~/.zshrc` 文件中。

### 方法 2：仪表盘 UI（鼠标点击）

1. 在浏览器中打开 `http://localhost:56243`
2. 在顶部 **:key: API 密钥** 卡片中点击 `[+ 添加]` 按钮
3. 输入服务类型、密钥值、标签（备注名称）和每日限额后保存

### 方法 3：REST API（自动化/脚本用）

REST API 是程序之间通过 HTTP 交换数据的方式。适用于通过脚本自动注册。

```bash
curl -X POST http://localhost:56243/admin/keys \
  -H "Authorization: Bearer 管理员令牌" \
  -H "Content-Type: application/json" \
  -d '{
    "service": "google",
    "key": "AIzaSy...",
    "label": "主密钥",
    "daily_limit": 1000
  }'
```

### 方法 4：proxy 标志（临时测试用）

无需正式注册，临时插入密钥进行测试。程序退出后密钥消失。

```bash
wall-vault proxy --key-google=AIzaSy... --key-openrouter=sk-or-...
```

---

## 代理使用方法

### 在 OpenClaw 中使用（主要用途）

以下是将 OpenClaw 配置为通过 wall-vault 连接 AI 服务的方法。

打开 `~/.openclaw/openclaw.json` 文件，添加以下内容：

```json5
// ~/.openclaw/openclaw.json
{
  models: {
    providers: {
      "wall-vault": {
        baseUrl: "http://localhost:56244/v1",
        apiKey: "your-agent-token",   // vault 代理令牌
        api: "openai-completions",
        models: [
          { id: "wall-vault/gemini-2.5-flash" },
          { id: "wall-vault/gemini-2.5-pro" },
          { id: "wall-vault/hunter-alpha" },    // 免费 1M 上下文
          { id: "wall-vault/claude-opus-4-6" }
        ]
      }
    }
  }
}
```

> :bulb: **更简单的方法**：点击仪表盘代理卡片上的 **:lobster: 复制 OpenClaw 配置** 按钮，已填好令牌和地址的配置片段会被复制到剪贴板。只需粘贴即可。

**模型名前的 `wall-vault/` 会连接到哪里？**

wall-vault 根据模型名自动判断将请求发送到哪个 AI 服务：

| 模型格式 | 连接的服务 |
|---------|----------|
| `wall-vault/gemini-*` | 直连 Google Gemini |
| `wall-vault/gpt-*`, `wall-vault/o3`, `wall-vault/o4*` | 直连 OpenAI |
| `wall-vault/claude-*` | 通过 OpenRouter 连接 Anthropic |
| `wall-vault/hunter-alpha`, `wall-vault/healer-alpha` | OpenRouter（免费百万令牌上下文） |
| `wall-vault/kimi-*`, `wall-vault/glm-*`, `wall-vault/deepseek-*` | OpenRouter |
| `google/模型名`, `openai/模型名`, `anthropic/模型名` 等 | 直连对应服务 |
| `custom/google/模型名`, `custom/openai/模型名` 等 | 去掉 `custom/` 部分后重新路由 |
| `模型名:cloud` | 去掉 `:cloud` 部分后连接 OpenRouter |

> :bulb: **什么是上下文（context）？** AI 一次能记住的对话量。1M（一百万令牌）意味着可以一次处理非常长的对话或文档。

### 通过 Gemini API 格式直接连接（兼容现有工具）

如果已有直接使用 Google Gemini API 的工具，只需将地址改为 wall-vault：

```bash
export ANTHROPIC_BASE_URL=http://localhost:56244/google
```

或者对于直接指定 URL 的工具：

```
http://localhost:56244/google/v1beta/models/gemini-2.5-flash:generateContent
```

### 在 OpenAI SDK（Python）中使用

在使用 AI 的 Python 代码中也可以连接 wall-vault。只需更改 `base_url`：

```python
from openai import OpenAI

client = OpenAI(
    base_url="http://localhost:56244/v1",
    api_key="not-needed"  # API 密钥由 wall-vault 自动管理
)

response = client.chat.completions.create(
    model="google/gemini-2.5-flash",   # 使用 provider/model 格式
    messages=[{"role": "user", "content": "你好"}]
)
```

### 运行中更改模型

在 wall-vault 已运行的状态下更改 AI 模型：

```bash
# 直接向代理请求更改模型
curl -X PUT http://localhost:56244/api/config/model \
  -H "Content-Type: application/json" \
  -d '{"service": "openrouter", "model": "anthropic/claude-3.5-sonnet"}'

# 分布式模式（多机器人）中在金库服务器更改 -> 通过 SSE 即时同步
curl -X PUT http://localhost:56243/admin/clients/my-bot-id \
  -H "Authorization: Bearer 管理员令牌" \
  -H "Content-Type: application/json" \
  -d '{"default_service": "google", "default_model": "gemini-2.5-pro"}'
```

### 查看可用模型列表

```bash
# 查看完整列表
curl http://localhost:56244/api/models | python3 -m json.tool

# 仅查看 Google 模型
curl "http://localhost:56244/api/models?service=google"

# 按名称搜索（例如：包含"claude"的模型）
curl "http://localhost:56244/api/models?q=claude"
```

**各服务主要模型：**

| 服务 | 主要模型 |
|------|---------|
| Google | gemini-2.5-pro, gemini-2.5-flash, gemini-2.5-flash-8b, gemini-2.0-flash |
| OpenAI | gpt-4o, gpt-4o-mini, o3, o1, o1-mini |
| OpenRouter | 346+（Hunter Alpha 1M 上下文免费、DeepSeek R1/V3、Qwen 2.5 等） |
| Ollama | 从本地安装的服务器自动检测 |
| LM Studio | 本地服务器（端口 1234） |
| vLLM | 本地服务器（端口 8000） |
| llama.cpp | 本地服务器（端口 8080） |

---

## 密钥金库仪表盘

在浏览器中打开 `http://localhost:56243` 即可查看仪表盘。

**界面布局：**
- **顶部固定栏（topbar）**：Logo、语言/主题选择器、SSE 连接状态
- **卡片网格**：代理、服务、API 密钥卡片以磁贴形式排列

### API 密钥卡片

用于一览管理所有已注册 API 密钥的卡片。

- 按服务分类显示密钥列表。
- `today_usage`：今日成功处理的令牌数（AI 读写的字符数）
- `today_attempts`：今日总调用次数（成功 + 失败）
- 用 `[+ 添加]` 按钮注册新密钥，用 `x` 删除密钥。

> :bulb: **什么是令牌（token）？** AI 处理文本时使用的单位。大约相当于一个英文单词，或 1-2 个中文字符。API 费用通常按令牌数计算。

### 代理卡片

显示连接到 wall-vault 代理的机器人（代理）状态的卡片。

**连接状态分为 4 个级别：**

| 指示灯 | 状态 | 含义 |
|--------|------|------|
| :green_circle: | 运行中 | 代理正常工作 |
| :yellow_circle: | 延迟 | 有响应但较慢 |
| :red_circle: | 离线 | 代理无响应 |
| :black_circle: | 未连接/已禁用 | 代理从未连接过金库或已被禁用 |

**代理卡片底部按钮说明：**

注册代理时指定**代理类型**后，对应类型的便捷按钮会自动出现。

---

#### :radio_button: 复制配置按钮 — 自动生成连接设置

点击按钮后，包含该代理的令牌、代理地址、模型信息的配置片段会被复制到剪贴板。只需将复制的内容粘贴到下表所示位置，即可完成连接设置。

| 按钮 | 代理类型 | 粘贴位置 |
|------|---------|---------|
| :lobster: 复制 OpenClaw 配置 | `openclaw` | `~/.openclaw/openclaw.json` |
| :crab: 复制 NanoClaw 配置 | `nanoclaw` | `~/.openclaw/openclaw.json` |
| :orange_circle: 复制 Claude Code 配置 | `claude-code` | `~/.claude/settings.json` |
| :keyboard: 复制 Cursor 配置 | `cursor` | Cursor -> Settings -> AI |
| :computer: 复制 VSCode 配置 | `vscode` | `~/.continue/config.json` |

**示例 — Claude Code 类型时，复制的内容如下：**

```json
// ~/.claude/settings.json
{
  "apiProvider": "openai",
  "baseUrl": "http://192.168.1.20:56244/v1",
  "apiKey": "此代理的令牌"
}
```

**示例 — VSCode（Continue）类型时：**

```yaml
# ~/.continue/config.yaml  <- 粘贴到 config.yaml，而非 config.json
name: My Config
version: 0.0.1
schema: v1

models:
  - name: wall-vault proxy
    provider: openai
    model: gemini-2.5-flash
    apiBase: http://192.168.1.20:56244/v1
    apiKey: 此代理的令牌
    roles:
      - chat
      - edit
      - apply
```

> :warning: **Continue 最新版使用 `config.yaml`。** 如果 `config.yaml` 存在，`config.json` 会被完全忽略。请务必粘贴到 `config.yaml`。

**示例 — Cursor 类型时：**

```
Base URL : http://192.168.1.20:56244/v1
API Key  : 此代理的令牌

// 或使用环境变量：
OPENAI_BASE_URL=http://192.168.1.20:56244/v1
OPENAI_API_KEY=此代理的令牌
```

> :warning: **剪贴板复制不可用时**：浏览器安全策略可能阻止复制。如果弹出文本框，请使用 Ctrl+A 全选后 Ctrl+C 复制。

---

#### :zap: 自动应用按钮 — 一键完成配置

代理类型为 `cline`、`claude-code`、`openclaw` 或 `nanoclaw` 时，代理卡片上会显示 **:zap: 应用配置** 按钮。点击此按钮会自动更新该代理的本地配置文件。

| 按钮 | 代理类型 | 目标文件 |
|------|---------|---------|
| :zap: 应用 Cline 配置 | `cline` | `~/.cline/data/globalState.json` + `secrets.json` |
| :zap: 应用 Claude Code 配置 | `claude-code` | `~/.claude/settings.json` |
| :zap: 应用 OpenClaw 配置 | `openclaw` | `~/.openclaw/openclaw.json` |
| :zap: 应用 NanoClaw 配置 | `nanoclaw` | `~/.openclaw/openclaw.json` |

> :warning: 此按钮向 **localhost:56244**（本地代理）发送请求。该机器上必须有代理在运行。

---

#### :twisted_rightwards_arrows: 拖拽排序卡片（v0.1.17，改进 v0.1.25）

可以**拖拽**仪表盘上的代理卡片来重新排列顺序。

1. 用鼠标抓住卡片左上角的**信号灯（●）**区域进行拖拽
2. 放到目标位置的卡片上即可交换顺序

> :bulb: 卡片主体（输入框、按钮等）不可拖拽。只能从信号灯区域抓取。

#### :orange_circle: 代理进程检测（v0.1.25）

当代理正常工作但本地代理进程（NanoClaw、OpenClaw）已停止时，卡片的信号灯变为**橙色（闪烁）**，并显示"代理进程已停止"消息。

- :green_circle: 绿色：代理 + 进程正常
- :orange_circle: 橙色（闪烁）：代理正常，进程已停止
- :red_circle: 红色：代理离线
3. 更改的顺序会**立即保存到服务器**，刷新页面后仍然保持

> :bulb: 触屏设备（手机/平板）暂不支持。请使用桌面浏览器。

---

#### :arrows_counterclockwise: 双向模型同步（v0.1.16）

在金库仪表盘中更改代理的模型后，该代理的本地配置会自动更新。

**Cline 的情况：**
- 在金库中更改模型 -> SSE 事件 -> 代理更新 `globalState.json` 的模型字段
- 更新对象：`actModeOpenAiModelId`、`planModeOpenAiModelId`、`openAiModelId`
- 不修改 `openAiBaseUrl` 和 API 密钥
- **需要重新加载 VS Code（`Ctrl+Alt+R` 或 `Ctrl+Shift+P` -> `Developer: Reload Window`）**
  - 因为 Cline 在运行中不会重新读取配置文件

**Claude Code 的情况：**
- 在金库中更改模型 -> SSE 事件 -> 代理更新 `settings.json` 的 `model` 字段
- 自动搜索 WSL 和 Windows 两侧路径（`~/.claude/`、`/mnt/c/Users/*/.claude/`）

**反向同步（代理 -> 金库）：**
- 代理（Cline、Claude Code 等）向 proxy 发送请求时，proxy 会在心跳中包含该客户端的服务/模型信息
- 金库仪表盘的代理卡片实时显示当前使用的服务/模型

> :bulb: **关键点**：proxy 通过请求的 Authorization 令牌识别代理，并自动路由到金库中配置的服务/模型。即使 Cline 或 Claude Code 发送了不同的模型名，proxy 也会用金库设置覆盖。

---

### 在 VS Code 中使用 Cline — 详细指南

#### 第 1 步：安装 Cline

从 VS Code 扩展市场安装 **Cline**（ID: `saoudrizwan.claude-dev`）。

#### 第 2 步：在金库中注册代理

1. 打开金库仪表盘（`http://金库IP:56243`）
2. 在**代理**部分点击 **+ 添加**
3. 按以下内容填写：

| 字段 | 值 | 说明 |
|------|---|------|
| ID | `my_cline` | 唯一标识符（英文数字，无空格） |
| 名称 | `My Cline` | 仪表盘上显示的名称 |
| 代理类型 | `cline` | <- 必须选择 `cline` |
| 服务 | 选择要使用的服务（如：`google`） | |
| 模型 | 输入要使用的模型（如：`gemini-2.5-flash`） | |

4. 点击**保存**后令牌会自动生成

#### 第 3 步：连接 Cline

**方法 A — 自动应用（推荐）**

1. 确认该机器上 wall-vault **proxy** 正在运行（`localhost:56244`）
2. 点击仪表盘代理卡片上的 **:zap: 应用 Cline 配置** 按钮
3. 出现"配置应用成功！"通知即表示成功
4. 重新加载 VS Code（`Ctrl+Alt+R`）

**方法 B — 手动设置**

在 Cline 侧边栏打开设置（:gear:）：
- **API Provider**：`OpenAI Compatible`
- **Base URL**：`http://代理地址:56244/v1`
  - 同一台机器：`http://localhost:56244/v1`
  - Mac Mini 等其他机器：`http://192.168.1.20:56244/v1`
- **API Key**：从金库获取的令牌（从代理卡片复制）
- **Model ID**：在金库中设置的模型（如：`gemini-2.5-flash`）

#### 第 4 步：验证

在 Cline 聊天窗口发送任意消息。如果正常：
- 金库仪表盘对应的代理卡片显示**绿色点（● 运行中）**
- 卡片显示当前服务/模型（如：`google / gemini-2.5-flash`）

#### 更改模型

想要更改 Cline 的模型时，请在**金库仪表盘**中操作：

1. 更改代理卡片的服务/模型下拉菜单
2. 点击**应用**
3. 重新加载 VS Code（`Ctrl+Alt+R`）— Cline 页脚的模型名会更新
4. 下一个请求开始使用新模型

> :bulb: 实际上 proxy 通过令牌识别 Cline 的请求，并路由到金库配置的模型。即使不重新加载 VS Code，**实际使用的模型也会立即切换** — 重新加载只是为了更新 Cline UI 的模型显示。

#### 断连检测

关闭 VS Code 后，金库仪表盘的代理卡片约 **90 秒**后变为黄色（延迟），**3 分钟**后变为红色（离线）。（v0.1.18 起，15 秒间隔的状态检查使离线检测更快。）

#### 故障排除

| 症状 | 原因 | 解决 |
|------|------|------|
| Cline 中出现"连接失败"错误 | proxy 未运行或地址错误 | 用 `curl http://localhost:56244/health` 检查 proxy |
| 金库中绿色点不显示 | API 密钥（令牌）未设置 | 再次点击 **:zap: 应用 Cline 配置** 按钮 |
| Cline 页脚模型不更新 | Cline 缓存了设置 | 重新加载 VS Code（`Ctrl+Alt+R`） |
| 显示错误的模型名 | 旧版 bug（v0.1.16 已修复） | 将 proxy 更新到 v0.1.16 以上 |

---

#### :purple_circle: 复制部署命令按钮 — 在新机器上安装时使用

在新电脑上首次安装 wall-vault proxy 并连接金库时使用。点击按钮会复制完整的安装脚本。粘贴到新电脑的终端并执行，以下操作一次完成：

1. wall-vault 二进制文件安装（已安装则跳过）
2. systemd 用户服务自动注册
3. 服务启动和金库自动连接

> :bulb: 脚本中已预填该代理的令牌和金库服务器地址，粘贴后无需任何修改即可直接运行。

---

### 服务卡片

用于启用/禁用和配置 AI 服务的卡片。

- 每个服务的启用/禁用切换开关
- 输入本地 AI 服务器（本机运行的 Ollama、LM Studio、vLLM、llama.cpp 等）的地址后，会自动发现可用模型。
- **本地服务连接状态显示**：服务名旁的 ● 为**绿色**表示已连接，**灰色**表示未连接
- **本地服务自动信号灯**（v0.1.23+）：本地服务（Ollama、LM Studio、vLLM、llama.cpp）根据连接状态自动启用/禁用。服务可达时 15 秒内 ● 变绿并自动勾选；断连时自动取消。与云服务（Google、OpenRouter 等）根据 API 密钥自动切换的方式相同。
- **推理模式开关**（v0.2.17+）：本地服务编辑面板底部会出现**推理模式**复选框。勾选后，代理在发往上游的 chat-completions 请求体中添加 `"reasoning": true`，使得 DeepSeek R1、Qwen QwQ 等支持思维过程输出的模型会额外返回 `<think>…</think>` 块。不认识该字段的服务器会直接忽略，因此即使在混合工作负载中也可以放心保持开启。

> :bulb: **如果本地服务在另一台电脑上运行**：在服务 URL 输入框中输入该电脑的 IP。例如：`http://192.168.1.20:11434`（Ollama）、`http://192.168.1.20:1234`（LM Studio）、`http://192.168.1.20:8080`（llama.cpp）。如果服务绑定在 `127.0.0.1` 而非 `0.0.0.0`，则无法通过外部 IP 访问——请在服务设置中检查绑定地址。

### 管理员令牌输入

在仪表盘中使用添加/删除密钥等重要功能时，会弹出管理员令牌输入框。输入 setup 向导中设置的令牌即可。输入一次后，在关闭浏览器之前一直有效。

> :warning: **如果 15 分钟内认证失败超过 10 次，该 IP 将被临时封锁。** 如果忘记令牌，请查看 `wall-vault.yaml` 文件中的 `admin_token` 字段。

---

## 分布式模式（多机器人）

在多台电脑上同时运行 OpenClaw 时，**共享一个密钥金库**的配置方式。只需在一处管理密钥，非常方便。

### 配置示例

```
[密钥金库服务器]
  wall-vault vault    （密钥金库 :56243，仪表盘）

[WSL Alpha]           [树莓派 Gamma]         [Mac Mini 本地]
  wall-vault proxy      wall-vault proxy        wall-vault proxy
  openclaw TUI          openclaw TUI            openclaw TUI
  <-> SSE 同步          <-> SSE 同步            <-> SSE 同步
```

所有机器人都指向中央金库服务器，因此在金库中更改模型或添加密钥会立即同步到所有机器人。

### 第 1 步：启动密钥金库服务器

在作为金库服务器的电脑上运行：

```bash
wall-vault vault
```

### 第 2 步：注册各机器人（客户端）

预先在金库服务器上注册各机器人的信息：

```bash
curl -X POST http://localhost:56243/admin/clients \
  -H "Authorization: Bearer 管理员令牌" \
  -H "Content-Type: application/json" \
  -d '{
    "id": "botA",
    "name": "机器人A",
    "token": "bota-secret",
    "default_service": "google",
    "default_model": "gemini-2.5-flash"
  }'
```

### 第 3 步：在各机器人机器上启动代理

在安装了机器人的各电脑上，指定金库服务器地址和令牌来运行代理：

```bash
WV_VAULT_URL=http://192.168.x.x:56243 \
WV_VAULT_TOKEN=bota-secret \
WV_VAULT_CLIENT_ID=botA \
wall-vault proxy
```

> :bulb: 请将 **`192.168.x.x`** 替换为金库服务器的实际内网 IP 地址。可通过路由器设置或 `ip addr` 命令查看。

---

## 自动启动设置

如果每次重启电脑都手动启动 wall-vault 太麻烦，可以注册为系统服务。注册一次后，开机时会自动启动。

### Linux — systemd（大多数 Linux 发行版）

systemd 是 Linux 中自动启动和管理程序的系统：

```bash
wall-vault doctor deploy
systemctl --user daemon-reload
systemctl --user enable --now wall-vault
```

查看日志：

```bash
journalctl --user -u wall-vault -f
```

### macOS — launchd

macOS 中负责程序自动启动的系统：

```bash
wall-vault doctor deploy launchd
launchctl load ~/Library/LaunchAgents/com.wall-vault.plist
```

### Windows — NSSM

1. 从 [nssm.cc](https://nssm.cc/download) 下载 NSSM 并添加到 PATH。
2. 在管理员 PowerShell 中：

```powershell
wall-vault doctor deploy windows
```

---

## Doctor 诊断工具

`doctor` 命令是一个能够**自我诊断并修复** wall-vault 配置问题的工具。

```bash
wall-vault doctor check   # 诊断当前状态（只读，不做任何更改）
wall-vault doctor fix     # 自动修复问题
wall-vault doctor all     # 诊断 + 自动修复一步完成
```

> :bulb: 如果觉得有什么不对，先运行 `wall-vault doctor all`。它能自动检测和修复许多问题。

---

## RTK 令牌节省

*(v0.1.24+)*

**RTK（令牌节省工具）** 自动压缩 AI 编程代理（如 Claude Code）执行的 shell 命令输出，减少令牌使用量。例如，`git status` 的 15 行输出可以缩减为 2 行摘要。

### 基本用法

```bash
# 用 wall-vault rtk 包裹命令即可自动过滤输出
wall-vault rtk git status          # 仅显示变更文件列表
wall-vault rtk git diff HEAD~1     # 仅变更行 + 最少上下文
wall-vault rtk git log -10         # 每行仅哈希 + 消息
wall-vault rtk go test ./...       # 仅显示失败的测试
wall-vault rtk ls -la              # 不支持的命令自动截断
```

### 支持的命令和节省效果

| 命令 | 过滤方式 | 节省率 |
|------|---------|--------|
| `git status` | 仅变更文件摘要 | ~87% |
| `git diff` | 变更行 + 3 行上下文 | ~60-94% |
| `git log` | 哈希 + 首行消息 | ~90% |
| `git push/pull/fetch` | 去除进度，仅保留摘要 | ~80% |
| `go test` | 仅显示失败，通过的计数 | ~88-99% |
| `go build/vet` | 仅显示错误 | ~90% |
| 其他所有命令 | 前 50 行 + 后 50 行，最大 32KB | 可变 |

### 三阶段过滤管道

1. **命令级结构过滤器** — 理解 git、go 等的输出格式，只提取有意义的部分
2. **正则表达式后处理** — 去除 ANSI 颜色代码、压缩空行、聚合重复行
3. **直通 + 截断** — 不支持的命令仅保留前/后 50 行

### Claude Code 集成

可以通过 Claude Code 的 `PreToolUse` 钩子，让所有 shell 命令自动经过 RTK。

```bash
# 安装钩子（自动添加到 Claude Code settings.json）
wall-vault rtk hook install
```

或手动添加到 `~/.claude/settings.json`：

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

> :bulb: **退出码保留**：RTK 原样返回原始命令的退出码。命令失败时（exit code != 0），AI 也能准确检测到失败。

> :bulb: **强制英文**：RTK 使用 `LC_ALL=C` 运行命令，无论系统语言设置如何，始终生成英文输出。这是过滤器正确工作的必要条件。

---

## 环境变量参考

环境变量是向程序传递配置值的方式。在终端输入 `export 变量名=值`，或写入自动启动服务文件中即可永久生效。

| 变量 | 说明 | 示例值 |
|------|------|--------|
| `WV_LANG` | 仪表盘语言 | `ko`, `en`, `ja` |
| `WV_THEME` | 仪表盘主题 | `light`, `dark`, `gold` |
| `WV_KEY_GOOGLE` | Google API 密钥（逗号分隔多个） | `AIza...,AIza...` |
| `WV_KEY_OPENROUTER` | OpenRouter API 密钥 | `sk-or-v1-...` |
| `WV_VAULT_URL` | 分布式模式中金库服务器地址 | `http://192.168.x.x:56243` |
| `WV_VAULT_TOKEN` | 客户端（机器人）认证令牌 | `my-secret-token` |
| `WV_ADMIN_TOKEN` | 管理员令牌 | `admin-token-here` |
| `WV_MASTER_PASS` | API 密钥加密密码 | `my-password` |
| `WV_AVATAR` | 头像图片路径（相对于 `~/.openclaw/`） | `workspace/avatars/avatar.png` |
| `OLLAMA_URL` | Ollama 本地服务器地址 | `http://192.168.x.x:11434` |

---

## 故障排除

### 代理无法启动

端口可能已被其他程序占用。

```bash
ss -tlnp | grep 56244   # 查看 56244 端口被谁占用
wall-vault proxy --port 8080   # 使用其他端口启动
```

### API 密钥错误（429, 402, 401, 403, 582）

| 错误码 | 含义 | 处理方法 |
|--------|------|---------|
| **429** | 请求过多（配额超限） | 稍等或添加更多密钥 |
| **402** | 需要付费或额度不足 | 在相应服务中充值 |
| **401 / 403** | 密钥错误或无权限 | 重新确认密钥并重新注册 |
| **582** | 网关过载（5 分钟冷却） | 5 分钟后自动解除 |

```bash
# 查看已注册的密钥列表和状态
curl -H "Authorization: Bearer 管理员令牌" http://localhost:56243/admin/keys

# 重置密钥使用计数器
curl -X POST -H "Authorization: Bearer 管理员令牌" http://localhost:56243/admin/keys/reset
```

### 代理显示"未连接"

"未连接"表示 proxy 进程没有向金库发送心跳信号。**这并不意味着设置没有保存。** proxy 需要在知道金库服务器地址和令牌的状态下运行。

```bash
# 指定金库服务器地址、令牌、客户端 ID 启动 proxy
WV_VAULT_URL=http://金库服务器:56243 \
WV_VAULT_TOKEN=客户端令牌 \
WV_VAULT_CLIENT_ID=客户端ID \
wall-vault proxy
```

连接成功后，约 20 秒内仪表盘会显示 :green_circle: 运行中。

### Ollama 无法连接

Ollama 是在本地电脑上直接运行 AI 的程序。首先检查 Ollama 是否在运行。

```bash
curl http://localhost:11434/api/tags   # 如果出现模型列表则正常
export OLLAMA_URL=http://192.168.x.x:11434   # 如果在其他电脑上运行
```

> :warning: 如果 Ollama 无响应，请先用 `ollama serve` 命令启动 Ollama。

> :warning: **大型模型响应很慢**：像 `qwen3.5:35b`、`deepseek-r1` 这样的大型模型，生成响应可能需要几分钟。即使看起来没有反应，也可能正在处理中——请耐心等待。

---

## 近期变更（v0.1.16 ~ v0.1.27）

### v0.1.27 (2026-04-09)
- **Ollama 回退模型名修复**：从其他服务回退到 Ollama 时，带提供商前缀的模型名（如 `google/gemini-3.1-pro-preview`）会直接传递给 Ollama 的问题已修复。现在自动替换为环境变量/默认模型。
- **冷却时间大幅缩短**：429 限速 30分钟->5分钟，402 付费 1小时->30分钟，401/403 24小时->6小时。防止所有密钥同时进入冷却导致代理完全瘫痪。
- **全部冷却时强制重试**：当所有密钥都在冷却状态时，强制重试最快解除冷却的密钥，防止请求被拒。
- **服务列表显示修复**：`/status` 响应现在显示从 vault 同步的实际服务列表（防止 anthropic 等被遗漏）。

### v0.1.25 (2026-04-08)
- **代理进程检测**：proxy 检测本地代理（NanoClaw/OpenClaw）是否存活，在仪表盘上以橙色信号灯显示。
- **拖拽手柄改进**：卡片排序时只能从信号灯（●）区域抓取。防止从输入框或按钮误触拖拽。

### v0.1.24 (2026-04-06)
- **RTK 令牌节省子命令**：`wall-vault rtk <command>` 自动过滤 shell 命令输出，将 AI 代理的令牌使用量减少 60-90%。内置 git、go 等主要命令的专用过滤器，不支持的命令也会自动截断。通过 Claude Code 的 `PreToolUse` 钩子透明集成。

### v0.1.23 (2026-04-06)
- **Ollama 模型更改修复**：在金库仪表盘中更改 Ollama 模型后不反映到实际 proxy 的问题已修复。之前仅使用环境变量（`OLLAMA_MODEL`），现在优先使用金库设置。
- **本地服务自动信号灯**：Ollama、LM Studio、vLLM 可达时自动启用，断连时自动禁用。与云服务基于密钥的自动切换方式相同。

### v0.1.22 (2026-04-05)
- **空 content 字段遗漏修复**：thinking 模型（gemini-3.1-pro、o1、claude thinking 等）将 max_tokens 配额全部用于 reasoning 而无法生成实际响应时，proxy 通过 `omitempty` 遗漏响应 JSON 的 `content`/`text` 字段，导致 OpenAI/Anthropic SDK 客户端出现 `Cannot read properties of undefined (reading 'trim')` 崩溃的问题已修复。已按官方 API 规范改为始终包含字段。

### v0.1.21 (2026-04-05)
- **Gemma 4 模型支持**：可通过 Google Gemini API 使用 `gemma-4-31b-it`、`gemma-4-26b-a4b-it` 等 Gemma 系列模型。
- **LM Studio / vLLM 正式支持**：之前这些服务在 proxy 路由中缺失，总是回退到 Ollama。现在通过 OpenAI 兼容 API 正常路由。
- **仪表盘服务显示修复**：即使发生回退，仪表盘也始终显示用户配置的服务。
- **本地服务状态显示**：仪表盘加载时，以 ● 点颜色显示本地服务（Ollama、LM Studio、vLLM 等）的连接状态。
- **工具过滤器环境变量**：可通过 `WV_TOOL_FILTER=passthrough` 环境变量设置工具传递模式。

### v0.1.20 (2026-03-28)
- **全面安全加固**：XSS 防护（41 处）、常量时间令牌比较、CORS 限制、请求大小限制、路径穿越防护、SSE 认证、速率限制器加固等 12 项安全改进。

### v0.1.19 (2026-03-27)
- **Claude Code 在线检测**：不经过 proxy 的 Claude Code 也能在仪表盘上显示为在线。

### v0.1.18 (2026-03-26)
- **回退服务粘连修复**：临时错误导致回退到 Ollama 后，原始服务恢复时自动切回。
- **离线检测改进**：15 秒间隔的状态检查使 proxy 停止检测更快。

### v0.1.17 (2026-03-25)
- **拖拽排序卡片**：代理卡片可拖拽重新排序。
- **内联配置应用按钮**：离线代理显示 [:zap: 应用配置] 按钮。
- **新增 cokacdir 代理类型**。

### v0.1.16 (2026-03-25)
- **双向模型同步**：在金库仪表盘中更改 Cline 或 Claude Code 的模型后自动反映。

---

*更详细的 API 信息请参阅 [API.md](API.md)。*
