# wall-vault 用户手册
*(最后更新: 2026-04-06 — v0.1.23)*

---

## 目录

1. [什么是 wall-vault？](#什么是-wall-vault)
2. [安装](#安装)
3. [初次启动（setup 向导）](#初次启动)
4. [注册 API 密钥](#注册-api-密钥)
5. [代理使用方法](#代理使用方法)
6. [密钥金库仪表盘](#密钥金库仪表盘)
7. [分布式模式（多机器人）](#分布式模式多机器人)
8. [设置自动启动](#设置自动启动)
9. [Doctor 诊断工具](#doctor-诊断工具)
10. [环境变量参考](#环境变量参考)
11. [故障排除](#故障排除)

---

## 什么是 wall-vault？

**wall-vault = 为 OpenClaw 打造的 AI 代理（代理服务器）+ API 密钥金库**

使用 AI 服务需要 **API 密钥**（API Key）。API 密钥就像"数字通行证"——它证明"这个人有权使用这项服务"。这张通行证每天有使用次数限制，如果管理不当还有泄露的风险。

wall-vault 将这些通行证安全地保存在加密金库中，并在 OpenClaw 与 AI 服务之间充当**代理人（代理服务器）**的角色。简单来说：OpenClaw 只需连接到 wall-vault，其余复杂的事情都由 wall-vault 自动处理。

wall-vault 解决的问题：

- **API 密钥自动轮换**：当某个密钥达到使用上限或暂时被限制（冷却期），会静默地切换到下一个密钥。OpenClaw 不会中断，持续正常运行。
- **服务自动切换（回退）**：如果 Google 没有响应，自动切换到 OpenRouter；如果还是不行，自动切换到本机安装的 Ollama·LM Studio·vLLM（本地 AI）。会话不会断开。当原始服务恢复后，从下一个请求开始自动切换回来（v0.1.18+，LM Studio/vLLM: v0.1.21+）。
- **实时同步（SSE）**：在金库仪表盘中切换模型，1～3 秒内即可在 OpenClaw 界面中生效。SSE（Server-Sent Events）是一种服务器将变化实时推送给客户端的技术。
- **实时通知**：密钥耗尽或服务故障等事件会立即显示在 OpenClaw TUI（终端界面）底部。

> 💡 **Claude Code、Cursor、VS Code** 也可以接入使用，但 wall-vault 的主要用途是配合 OpenClaw 一起使用。

```
OpenClaw（TUI 终端界面）
        │
        ▼
  wall-vault 代理 (:56244)   ← 密钥管理、路由、回退、事件推送
        │
        ├─ Google Gemini API
        ├─ OpenRouter API（340+ 个模型）
        ├─ Ollama / LM Studio / vLLM（本地计算机，最后保障）
        └─ OpenAI / Anthropic API
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

- `curl -L ...` — 从互联网下载文件。
- `chmod +x` — 将下载的文件设为"可执行"。如果跳过此步骤，会出现"权限被拒绝"的错误。

### Windows

以管理员身份打开 PowerShell，运行以下命令：

```powershell
# 下载
Invoke-WebRequest -Uri `
  "https://github.com/sookmook/wall-vault/releases/latest/download/wall-vault-windows-amd64.exe" `
  -OutFile "$env:LOCALAPPDATA\Programs\wall-vault\wall-vault.exe"

# 添加到 PATH（重启 PowerShell 后生效）
$env:PATH += ";$env:LOCALAPPDATA\Programs\wall-vault"
```

> 💡 **什么是 PATH？** PATH 是计算机查找命令的文件夹列表。将 wall-vault 添加到 PATH 后，就可以在任何目录下直接输入 `wall-vault` 来运行。

### 从源码构建（开发者用）

仅适用于已安装 Go 语言开发环境的情况。

```bash
git clone https://github.com/sookmook/wall-vault
cd wall-vault
make build       # bin/wall-vault（版本: v0.1.23.YYYYMMDD.HHmmss）
make install     # ~/.local/bin/wall-vault
```

> 💡 **构建时间戳版本**：使用 `make build` 构建时，版本号会自动生成为 `v0.1.23.20260406.211004` 这样包含日期和时间的格式。如果直接用 `go build ./...` 构建，版本号只会显示 `"dev"`。

---

## 初次启动

### 运行 setup 向导

安装后，请务必先运行以下命令启动**设置向导**。向导会逐步询问并引导您完成必要的配置。

```bash
wall-vault setup
```

向导的步骤如下：

```
1. 语言选择（包括中文在内的 10 种语言）
2. 主题选择（light / dark / gold / cherry / ocean）
3. 运行模式 — 单机使用（standalone）还是多台共享（distributed）
4. 机器人名称 — 仪表盘上显示的名称
5. 端口设置 — 默认值：代理 56244，金库 56243（不需要更改直接回车）
6. AI 服务选择 — 从 Google / OpenRouter / Ollama / LM Studio / vLLM 中选择要使用的服务
7. 工具安全过滤器设置
8. 管理员令牌设置 — 用于锁定仪表盘管理功能的密码，可自动生成
9. API 密钥加密密码设置 — 想要更安全地存储密钥时使用（可选）
10. 配置文件保存路径
```

> ⚠️ **请务必记住管理员令牌。** 以后在仪表盘中添加密钥或更改设置时需要用到。如果忘记了，需要直接编辑配置文件。

向导完成后，`wall-vault.yaml` 配置文件会自动创建。

### 启动

```bash
wall-vault start
```

以下两个服务器会同时启动：

- **代理** (`http://localhost:56244`) — 连接 OpenClaw 与 AI 服务的中间人
- **密钥金库** (`http://localhost:56243`) — API 密钥管理及 Web 仪表盘

在浏览器中打开 `http://localhost:56243` 即可查看仪表盘。

---

## 注册 API 密钥

注册 API 密钥有四种方法。**对于初学者，推荐使用方法 1（环境变量）。**

### 方法 1：环境变量（推荐 — 最简单）

环境变量是程序启动时读取的**预设值**。在终端中输入以下内容即可：

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

> 💡 **提示**：`export` 命令仅对当前终端会话有效。要在重启后仍然生效，请将上述行添加到 `~/.bashrc` 或 `~/.zshrc` 文件中。

### 方法 2：仪表盘 UI（鼠标点击）

1. 在浏览器中打开 `http://localhost:56243`
2. 在顶部 **🔑 API 密钥** 卡片中点击 `[+ 添加]` 按钮
3. 输入服务类型、密钥值、标签（备注名称）和每日限额，然后保存

### 方法 3：REST API（自动化/脚本用）

REST API 是程序之间通过 HTTP 交换数据的方式。在用脚本自动注册时很有用。

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

不正式注册，临时注入密钥进行测试。程序停止后密钥即消失。

```bash
wall-vault proxy --key-google=AIzaSy... --key-openrouter=sk-or-...
```

---

## 代理使用方法

### 在 OpenClaw 中使用（主要用途）

以下是配置 OpenClaw 通过 wall-vault 连接 AI 服务的方法。

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

> 💡 **更简单的方法**：点击仪表盘代理卡片上的 **🦞 复制 OpenClaw 配置** 按钮，会复制已填好令牌和地址的配置片段到剪贴板。直接粘贴即可。

**模型名称前的 `wall-vault/` 会连接到哪里？**

wall-vault 根据模型名称自动判断将请求发送到哪个 AI 服务：

| 模型格式 | 连接目标 |
|---------|---------|
| `wall-vault/gemini-*` | Google Gemini 直连 |
| `wall-vault/gpt-*`, `wall-vault/o3`, `wall-vault/o4*` | OpenAI 直连 |
| `wall-vault/claude-*` | 通过 OpenRouter 连接 Anthropic |
| `wall-vault/hunter-alpha`, `wall-vault/healer-alpha` | OpenRouter（免费 100 万 token 上下文） |
| `wall-vault/kimi-*`, `wall-vault/glm-*`, `wall-vault/deepseek-*` | OpenRouter 连接 |
| `google/模型名`, `openai/模型名`, `anthropic/模型名` 等 | 直连对应服务 |
| `custom/google/模型名`, `custom/openai/模型名` 等 | 去掉 `custom/` 部分后重新路由 |
| `模型名:cloud` | 去掉 `:cloud` 部分后连接 OpenRouter |

> 💡 **什么是上下文（context）？** 上下文是 AI 一次能记住的对话量。1M（100 万 token）意味着可以一次处理非常长的对话或文档。

### 以 Gemini API 格式直连（兼容现有工具）

如果你有直接使用 Google Gemini API 的工具，只需将地址改为 wall-vault：

```bash
export ANTHROPIC_BASE_URL=http://localhost:56244/google
```

或者如果工具需要直接指定 URL：

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

### 运行中更换模型

在 wall-vault 运行中更换 AI 模型：

```bash
# 直接向代理发送请求来更换模型
curl -X PUT http://localhost:56244/api/config/model \
  -H "Content-Type: application/json" \
  -d '{"service": "openrouter", "model": "anthropic/claude-3.5-sonnet"}'

# 分布式模式（多机器人）中在金库服务器上更改 → 通过 SSE 即时同步
curl -X PUT http://localhost:56243/admin/clients/我的机器人ID \
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

# 按名称搜索（例：包含"claude"的模型）
curl "http://localhost:56244/api/models?q=claude"
```

**各服务主要模型：**

| 服务 | 主要模型 |
|------|---------|
| Google | gemini-2.5-pro, gemini-2.5-flash, gemini-2.5-flash-8b, gemini-2.0-flash |
| OpenAI | gpt-4o, gpt-4o-mini, o3, o1, o1-mini |
| OpenRouter | 346 个以上（Hunter Alpha 1M 上下文免费，DeepSeek R1/V3，Qwen 2.5 等） |
| Ollama | 自动检测本机安装的模型 |
| LM Studio | 本地服务器（端口 1234） |
| vLLM | 本地服务器（端口 8000） |

---

## 密钥金库仪表盘

在浏览器中打开 `http://localhost:56243` 即可查看仪表盘。

**界面布局：**
- **顶部固定栏（topbar）**：Logo、语言/主题选择器、SSE 连接状态指示
- **卡片网格**：代理、服务、API 密钥卡片以磁贴形式排列

### API 密钥卡片

一目了然地管理已注册 API 密钥的卡片。

- 按服务分类显示密钥列表。
- `today_usage`：今天成功处理的 token（AI 读写的文字单位）数量
- `today_attempts`：今天的总调用次数（成功 + 失败）
- 用 `[+ 添加]` 按钮注册新密钥，用 `✕` 删除密钥。

> 💡 **什么是 token？** token 是 AI 处理文本时使用的单位。大约相当于一个英文单词，或 1～2 个中文字符。API 费用通常按 token 数量计算。

### 代理卡片

显示连接到 wall-vault 代理的机器人（代理）状态的卡片。

**连接状态分为 4 个等级：**

| 指示 | 状态 | 含义 |
|------|------|------|
| 🟢 | 运行中 | 代理正常工作 |
| 🟡 | 延迟 | 有响应但较慢 |
| 🔴 | 离线 | 代理没有响应 |
| ⚫ | 未连接/已禁用 | 代理从未连接过金库，或已被禁用 |

**代理卡片底部按钮说明：**

注册代理时指定**代理类型**后，会自动显示与该类型对应的便捷按钮。

---

#### 🔘 复制配置按钮 — 自动生成连接配置

点击按钮后，该代理的令牌、代理地址、模型信息已填好的配置片段会被复制到剪贴板。只需将复制的内容粘贴到下表所示位置即可完成连接配置。

| 按钮 | 代理类型 | 粘贴位置 |
|------|---------|---------|
| 🦞 复制 OpenClaw 配置 | `openclaw` | `~/.openclaw/openclaw.json` |
| 🦀 复制 NanoClaw 配置 | `nanoclaw` | `~/.openclaw/openclaw.json` |
| 🟠 复制 Claude Code 配置 | `claude-code` | `~/.claude/settings.json` |
| ⌨ 复制 Cursor 配置 | `cursor` | Cursor → Settings → AI |
| 💻 复制 VSCode 配置 | `vscode` | `~/.continue/config.json` |

**示例 — Claude Code 类型会复制以下内容：**

```json
// ~/.claude/settings.json
{
  "apiProvider": "openai",
  "baseUrl": "http://192.168.0.6:56244/v1",
  "apiKey": "此代理的令牌"
}
```

**示例 — VSCode（Continue）类型：**

```yaml
# ~/.continue/config.yaml  ← 粘贴到 config.yaml，而非 config.json
name: My Config
version: 0.0.1
schema: v1

models:
  - name: wall-vault proxy
    provider: openai
    model: gemini-2.5-flash
    apiBase: http://192.168.0.6:56244/v1
    apiKey: 此代理的令牌
    roles:
      - chat
      - edit
      - apply
```

> ⚠️ **Continue 最新版本使用 `config.yaml`。** 如果 `config.yaml` 存在，`config.json` 会被完全忽略。请务必粘贴到 `config.yaml`。

**示例 — Cursor 类型：**

```
Base URL : http://192.168.0.6:56244/v1
API Key  : 此代理的令牌

// 或者环境变量：
OPENAI_BASE_URL=http://192.168.0.6:56244/v1
OPENAI_API_KEY=此代理的令牌
```

> ⚠️ **如果无法复制到剪贴板**：浏览器安全策略可能会阻止复制。如果弹出文本框，请用 Ctrl+A 全选后 Ctrl+C 复制。

---

#### ⚡ 自动应用按钮 — 一键完成配置

代理类型为 `cline`、`claude-code`、`openclaw` 或 `nanoclaw` 时，代理卡片上会显示 **⚡ 应用配置** 按钮。点击此按钮会自动更新该代理的本地配置文件。

| 按钮 | 代理类型 | 目标文件 |
|------|---------|---------|
| ⚡ 应用 Cline 配置 | `cline` | `~/.cline/data/globalState.json` + `secrets.json` |
| ⚡ 应用 Claude Code 配置 | `claude-code` | `~/.claude/settings.json` |
| ⚡ 应用 OpenClaw 配置 | `openclaw` | `~/.openclaw/openclaw.json` |
| ⚡ 应用 NanoClaw 配置 | `nanoclaw` | `~/.openclaw/openclaw.json` |

> ⚠️ 此按钮会向 **localhost:56244**（本地代理）发送请求。该机器上必须有代理在运行才能生效。

---

#### 🔀 拖拽排序卡片（v0.1.17）

可以通过**拖拽**仪表盘上的代理卡片来重新排列顺序。

1. 用鼠标抓住代理卡片并拖动
2. 放到目标位置的卡片上即可交换位置
3. 更改后的顺序会**立即保存到服务器**，刷新后仍然保持

> 💡 触摸设备（手机/平板）暂不支持。请在桌面浏览器中使用。

---

#### 🔄 双向模型同步（v0.1.16）

在金库仪表盘中更改代理的模型后，该代理的本地配置会自动更新。

**Cline 的情况：**
- 在金库中更改模型 → SSE 事件 → 代理更新 `globalState.json` 中的模型字段
- 更新对象：`actModeOpenAiModelId`、`planModeOpenAiModelId`、`openAiModelId`
- 不会修改 `openAiBaseUrl` 和 API 密钥
- **需要重新加载 VS Code（`Ctrl+Alt+R` 或 `Ctrl+Shift+P` → `Developer: Reload Window`）**
  - 因为 Cline 在运行中不会重新读取配置文件

**Claude Code 的情况：**
- 在金库中更改模型 → SSE 事件 → 代理更新 `settings.json` 中的 `model` 字段
- 自动搜索 WSL 和 Windows 两侧的路径（`~/.claude/`、`/mnt/c/Users/*/.claude/`）

**反向同步（代理 → 金库）：**
- 当代理（Cline、Claude Code 等）向代理发送请求时，代理会在心跳中包含该客户端的服务/模型信息
- 金库仪表盘的代理卡片会实时显示当前正在使用的服务/模型

> 💡 **关键点**：代理通过请求中的 Authorization 令牌识别代理，并自动路由到金库中配置的服务/模型。即使 Cline 或 Claude Code 发送不同的模型名称，代理也会用金库的配置覆盖。

---

### 在 VS Code 中使用 Cline — 详细指南

#### 第 1 步：安装 Cline

在 VS Code 扩展市场中安装 **Cline**（ID：`saoudrizwan.claude-dev`）。

#### 第 2 步：在金库中注册代理

1. 打开金库仪表盘（`http://金库IP:56243`）
2. 在**代理**部分点击 **+ 添加**
3. 按以下内容填写：

| 字段 | 值 | 说明 |
|------|-----|------|
| ID | `my_cline` | 唯一标识符（英文字母数字，不含空格） |
| 名称 | `My Cline` | 仪表盘上显示的名称 |
| 代理类型 | `cline` | ← 必须选择 `cline` |
| 服务 | 选择要使用的服务（例：`google`） | |
| 模型 | 输入要使用的模型（例：`gemini-2.5-flash`） | |

4. 点击**保存**，令牌会自动生成

#### 第 3 步：连接 Cline

**方法 A — 自动应用（推荐）**

1. 确认该机器上 wall-vault **代理**正在运行（`localhost:56244`）
2. 在仪表盘的代理卡片上点击 **⚡ 应用 Cline 配置** 按钮
3. 出现"配置应用成功！"通知即表示成功
4. 重新加载 VS Code（`Ctrl+Alt+R`）

**方法 B — 手动设置**

在 Cline 侧边栏中打开设置（⚙️）：
- **API Provider**：`OpenAI Compatible`
- **Base URL**：`http://代理地址:56244/v1`
  - 同一台机器：`http://localhost:56244/v1`
  - 其他机器（如 Mac Mini）：`http://192.168.0.6:56244/v1`
- **API Key**：金库发放的令牌（从代理卡片复制）
- **Model ID**：金库中设置的模型（例：`gemini-2.5-flash`）

#### 第 4 步：验证

在 Cline 聊天窗口中发送任意消息。正常情况下：
- 金库仪表盘中该代理卡片会显示**绿色圆点（● 运行中）**
- 卡片上会显示当前的服务/模型（例：`google / gemini-2.5-flash`）

#### 更换模型

想要更换 Cline 的模型时，请在**金库仪表盘**中操作：

1. 更改代理卡片上的服务/模型下拉菜单
2. 点击**应用**
3. 重新加载 VS Code（`Ctrl+Alt+R`）—— Cline 底部的模型名称会更新
4. 从下一个请求开始使用新模型

> 💡 实际上，代理通过令牌识别 Cline 的请求，并路由到金库配置的模型。即使不重新加载 VS Code，**实际使用的模型也会立即切换**——重新加载只是为了更新 Cline UI 中的模型显示。

#### 断连检测

关闭 VS Code 后，金库仪表盘中的代理卡片约 **90 秒**后变为黄色（延迟），**3 分钟**后变为红色（离线）。（从 v0.1.18 起，通过 15 秒间隔的状态检查，离线检测更加迅速。）

#### 故障排除

| 症状 | 原因 | 解决方法 |
|------|------|---------|
| Cline 中出现"连接失败"错误 | 代理未运行或地址错误 | 用 `curl http://localhost:56244/health` 检查代理 |
| 金库中绿色圆点不显示 | API 密钥（令牌）未配置 | 再次点击 **⚡ 应用 Cline 配置** 按钮 |
| Cline 底部模型名称不变 | Cline 缓存了设置 | 重新加载 VS Code（`Ctrl+Alt+R`） |
| 显示错误的模型名称 | 旧版 bug（v0.1.16 已修复） | 将代理更新到 v0.1.16 或更高版本 |

---

#### 🟣 复制部署命令按钮 — 在新机器上安装时使用

在新计算机上首次安装 wall-vault 代理并连接到金库时使用。点击按钮会复制完整的安装脚本。将其粘贴到新计算机的终端中运行，以下操作会一次性完成：

1. 安装 wall-vault 二进制文件（已安装则跳过）
2. 自动注册 systemd 用户服务
3. 启动服务并自动连接到金库

> 💡 脚本中已包含此代理的令牌和金库服务器地址，粘贴后无需任何修改即可直接运行。

---

### 服务卡片

用于开关和配置 AI 服务的卡片。

- 每个服务的启用/禁用开关
- 输入本地 AI 服务器（本机运行的 Ollama、LM Studio、vLLM 等）的地址，会自动发现可用模型
- **本地服务连接状态显示**：服务名称旁的 ● 圆点为**绿色**表示已连接，**灰色**表示未连接
- **本地服务自动信号灯**（v0.1.23+）：本地服务（Ollama、LM Studio、vLLM）根据连接可用性自动启用/禁用。当服务可达时，15 秒内 ● 变为绿色且复选框自动勾选；当服务关闭时自动取消。这与云服务（Google、OpenRouter 等）根据 API 密钥有无自动切换的方式相同。

> 💡 **如果本地服务在其他计算机上运行**：在服务 URL 输入框中输入该计算机的 IP。例：`http://192.168.0.6:11434`（Ollama），`http://192.168.0.6:1234`（LM Studio）。如果服务仅绑定到 `127.0.0.1` 而非 `0.0.0.0`，则无法通过外部 IP 访问——请检查服务的绑定地址设置。

### 管理员令牌输入

在仪表盘中尝试使用添加/删除密钥等重要功能时，会弹出管理员令牌输入框。请输入在 setup 向导中设置的令牌。输入一次后在关闭浏览器前一直有效。

> ⚠️ **如果 15 分钟内认证失败超过 10 次，该 IP 将被临时封禁。** 如果忘记了令牌，请查看 `wall-vault.yaml` 文件中的 `admin_token` 字段。

---

## 分布式模式（多机器人）

在多台计算机上同时运行 OpenClaw 时，可以**共享一个密钥金库**。密钥管理只需在一个地方进行，非常方便。

### 配置示例

```
[密钥金库服务器]
  wall-vault vault    (密钥金库 :56243，仪表盘)

[WSL Alpha]           [树莓派 Gamma]          [Mac Mini 本地]
  wall-vault proxy      wall-vault proxy        wall-vault proxy
  openclaw TUI          openclaw TUI            openclaw TUI
  ↕ SSE 同步            ↕ SSE 同步              ↕ SSE 同步
```

所有机器人都指向中央金库服务器，因此在金库中更改模型或添加密钥时，会立即同步到所有机器人。

### 第 1 步：启动密钥金库服务器

在用作金库服务器的计算机上运行：

```bash
wall-vault vault
```

### 第 2 步：注册各机器人（客户端）

预先注册将连接到金库服务器的各机器人信息：

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

### 第 3 步：在各机器人的计算机上启动代理

在安装了机器人的各台计算机上，指定金库服务器地址和令牌来运行代理：

```bash
WV_VAULT_URL=http://192.168.x.x:56243 \
WV_VAULT_TOKEN=bota-secret \
WV_VAULT_CLIENT_ID=botA \
wall-vault proxy
```

> 💡 将 **`192.168.x.x`** 替换为金库服务器计算机的实际内网 IP 地址。可通过路由器设置或 `ip addr` 命令查看。

---

## 设置自动启动

如果每次重启计算机都手动启动 wall-vault 很麻烦，可以将其注册为系统服务。注册一次后，开机时会自动启动。

### Linux — systemd（大多数 Linux）

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

`doctor` 命令是一个**自我诊断并修复** wall-vault 配置问题的工具。

```bash
wall-vault doctor check   # 诊断当前状态（只读，不做任何更改）
wall-vault doctor fix     # 自动修复问题
wall-vault doctor all     # 诊断 + 自动修复一步到位
```

> 💡 如果觉得有什么不对劲，先试试 `wall-vault doctor all`。它能自动捕获并修复很多问题。

---

## 环境变量参考

环境变量是向程序传递配置值的方法。在终端中以 `export 变量名=值` 的形式输入，或者写入自动启动服务文件中即可持久生效。

| 变量 | 说明 | 示例值 |
|------|------|--------|
| `WV_LANG` | 仪表盘语言 | `ko`, `en`, `ja` |
| `WV_THEME` | 仪表盘主题 | `light`, `dark`, `gold` |
| `WV_KEY_GOOGLE` | Google API 密钥（多个用逗号分隔） | `AIza...,AIza...` |
| `WV_KEY_OPENROUTER` | OpenRouter API 密钥 | `sk-or-v1-...` |
| `WV_VAULT_URL` | 分布式模式下金库服务器地址 | `http://192.168.x.x:56243` |
| `WV_VAULT_TOKEN` | 客户端（机器人）认证令牌 | `my-secret-token` |
| `WV_ADMIN_TOKEN` | 管理员令牌 | `admin-token-here` |
| `WV_MASTER_PASS` | API 密钥加密密码 | `my-password` |
| `WV_AVATAR` | 头像图片文件路径（相对于 `~/.openclaw/` 的路径） | `workspace/avatars/avatar.png` |
| `OLLAMA_URL` | Ollama 本地服务器地址 | `http://192.168.x.x:11434` |

---

## 故障排除

### 代理无法启动

端口通常已被其他程序占用。

```bash
ss -tlnp | grep 56244   # 查看谁在使用端口 56244
wall-vault proxy --port 8080   # 使用其他端口启动
```

### API 密钥错误（429, 402, 401, 403, 582）

| 错误码 | 含义 | 处理方法 |
|--------|------|---------|
| **429** | 请求过多（配额超限） | 等待一会儿或添加更多密钥 |
| **402** | 需要付费或额度不足 | 在相应服务中充值 |
| **401 / 403** | 密钥错误或无权限 | 重新确认密钥值后重新注册 |
| **582** | 网关过载（5 分钟冷却） | 5 分钟后自动解除 |

```bash
# 查看已注册密钥列表和状态
curl -H "Authorization: Bearer 管理员令牌" http://localhost:56243/admin/keys

# 重置密钥使用量计数器
curl -X POST -H "Authorization: Bearer 管理员令牌" http://localhost:56243/admin/keys/reset
```

### 代理显示为"未连接"

"未连接"表示代理进程没有向金库发送心跳信号。**这并不意味着配置未保存。** 代理需要在知道金库服务器地址和令牌的情况下运行才能建立连接。

```bash
# 指定金库服务器地址、令牌和客户端 ID 来启动代理
WV_VAULT_URL=http://金库服务器地址:56243 \
WV_VAULT_TOKEN=客户端令牌 \
WV_VAULT_CLIENT_ID=客户端ID \
wall-vault proxy
```

连接成功后，仪表盘会在约 20 秒内显示 🟢 运行中。

### Ollama 连接不上

Ollama 是在本地计算机上直接运行 AI 的程序。首先确认 Ollama 是否在运行。

```bash
curl http://localhost:11434/api/tags   # 如果显示模型列表则正常
export OLLAMA_URL=http://192.168.x.x:11434   # 如果在其他计算机上运行
```

> ⚠️ 如果 Ollama 没有响应，请先用 `ollama serve` 命令启动 Ollama。

> ⚠️ **大型模型响应较慢**：`qwen3.5:35b`、`deepseek-r1` 等大型模型可能需要几分钟才能生成响应。即使看起来没有反应，也可能正在正常处理中，请耐心等待。

---

## 最近更新（v0.1.16 ~ v0.1.23）

### v0.1.23 (2026-04-06)
- **Ollama 模型更改修复**：修复了在金库仪表盘中更改 Ollama 模型但实际代理未生效的问题。以前仅使用环境变量（`OLLAMA_MODEL`），现在优先使用金库设置。
- **本地服务自动信号灯**：Ollama·LM Studio·vLLM 可连接时自动启用，断开时自动禁用。与云服务基于密钥的自动切换方式相同。

### v0.1.22 (2026-04-05)
- **空 content 字段缺失修复**：当 thinking 模型（gemini-3.1-pro、o1、claude thinking 等）将 max_tokens 全部用于 reasoning 而无法生成实际响应时，代理通过 `omitempty` 省略了响应 JSON 中的 `content`/`text` 字段，导致 OpenAI/Anthropic SDK 客户端因 `Cannot read properties of undefined (reading 'trim')` 错误而崩溃。已按照官方 API 规范修改为始终包含这些字段。

### v0.1.21 (2026-04-05)
- **Gemma 4 模型支持**：现在可以通过 Google Gemini API 使用 `gemma-4-31b-it`、`gemma-4-26b-a4b-it` 等 Gemma 系列模型。
- **LM Studio / vLLM 服务正式支持**：以前这些服务在代理路由中被遗漏，总是回退到 Ollama。现在通过 OpenAI 兼容 API 正常路由。
- **仪表盘服务显示修复**：即使发生回退，仪表盘仍始终显示用户配置的服务。
- **本地服务状态显示**：仪表盘加载时以 ● 圆点颜色显示本地服务（Ollama、LM Studio、vLLM 等）的连接状态。
- **工具过滤器环境变量**：可使用 `WV_TOOL_FILTER=passthrough` 环境变量设置工具传递模式。

### v0.1.20 (2026-03-28)
- **全面安全加固**：XSS 防护（41 处）、常量时间令牌比较、CORS 限制、请求大小限制、路径遍历防护、SSE 认证、速率限制器加固等 12 项安全改进。

### v0.1.19 (2026-03-27)
- **Claude Code 在线检测**：不经过代理的 Claude Code 实例也会在仪表盘中显示为在线。

### v0.1.18 (2026-03-26)
- **回退服务粘滞修复**：临时错误导致 Ollama 回退后，原始服务恢复时会自动切回。
- **离线检测改进**：通过 15 秒间隔的状态检查，代理中断检测更加迅速。

### v0.1.17 (2026-03-25)
- **拖拽排序卡片**：代理卡片可通过拖拽来更改顺序。
- **内联配置应用按钮**：离线代理上显示 [⚡ 应用配置] 按钮。
- **新增 cokacdir 代理类型**。

### v0.1.16 (2026-03-25)
- **双向模型同步**：在金库仪表盘中更改 Cline·Claude Code 的模型后会自动同步。

---

*更详细的 API 信息请参阅 [API.md](API.md)。*
