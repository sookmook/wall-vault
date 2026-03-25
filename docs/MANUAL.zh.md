# wall-vault 用户手册
*(最后更新: 2026-03-20 — v0.1.15)*

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
- **服务自动切换（回退）**：如果 Google 没有响应，自动切换到 OpenRouter；如果还是不行，自动切换到本机安装的 Ollama（本地 AI）。会话不会断开。
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
        └─ Ollama（本地计算机，最后保障）
```

---

## 安装

### Linux / macOS

打开终端，将以下命令直接粘贴运行。

```bash
# Linux（普通 PC、服务器 — amd64）
curl -L https://github.com/sookmook/wall-vault/releases/latest/download/wall-vault-linux-amd64 \
  -o ~/.local/bin/wall-vault && chmod +x ~/.local/bin/wall-vault

# macOS Apple Silicon（M1/M2/M3 Mac）
curl -L https://github.com/sookmook/wall-vault/releases/latest/download/wall-vault-darwin-arm64 \
  -o /usr/local/bin/wall-vault && chmod +x /usr/local/bin/wall-vault
```

- `curl -L ...` — 从网络下载文件。
- `chmod +x` — 将下载的文件设置为"可执行"。如果跳过这一步，会出现"权限不足"的错误。

### Windows

以管理员权限打开 PowerShell，执行以下命令。

```powershell
# 下载
Invoke-WebRequest -Uri `
  "https://github.com/sookmook/wall-vault/releases/latest/download/wall-vault-windows-amd64.exe" `
  -OutFile "$env:LOCALAPPDATA\Programs\wall-vault\wall-vault.exe"

# 添加到 PATH（重启 PowerShell 后生效）
$env:PATH += ";$env:LOCALAPPDATA\Programs\wall-vault"
```

> 💡 **什么是 PATH？** PATH 是计算机查找命令的文件夹列表。将路径添加到 PATH 后，无论在哪个文件夹中，都可以直接输入 `wall-vault` 来运行程序。

### 从源码编译（开发者专用）

仅适用于已安装 Go 语言开发环境的情况。

```bash
git clone https://github.com/sookmook/wall-vault
cd wall-vault
make build       # bin/wall-vault（版本：v0.1.6.YYYYMMDD.HHmmss）
make install     # ~/.local/bin/wall-vault
```

> 💡 **构建时间戳版本**：使用 `make build` 编译时，版本号会自动生成为包含日期和时间的格式，例如 `v0.1.6.20260314.231308`。如果直接使用 `go build ./...` 编译，版本号只会显示为 `"dev"`。

---

## 初次启动

### 运行 setup 向导

安装完成后，第一次使用时必须运行以下命令启动**设置向导**。向导会逐步询问所需的配置项并引导您完成设置。

```bash
wall-vault setup
```

向导的步骤如下：

```
1. 选择语言（包括中文在内的 10 种语言）
2. 选择主题（light / dark / gold / cherry / ocean）
3. 运行模式 — 选择单机使用（standalone）还是多机器共用（distributed）
4. 输入机器人名称 — 显示在仪表盘上的名称
5. 端口设置 — 默认：代理 56244，金库 56243（不需要修改就直接按回车）
6. 选择 AI 服务 — Google / OpenRouter / Ollama 中选择要使用的服务
7. 工具安全过滤器设置
8. 管理员令牌设置 — 锁定仪表盘管理功能的密码，也可以自动生成
9. API 密钥加密密码设置 — 如果希望更安全地存储密钥（可选）
10. 配置文件保存路径
```

> ⚠️ **请务必记住管理员令牌。** 之后在仪表盘中添加密钥或修改设置时需要用到。如果忘记了，需要直接修改配置文件。

向导完成后，`wall-vault.yaml` 配置文件会自动生成。

### 启动

```bash
wall-vault start
```

以下两个服务器会同时启动：

- **代理服务器**（`http://localhost:56244`）— 连接 OpenClaw 与 AI 服务的中间人
- **密钥金库**（`http://localhost:56243`）— API 密钥管理及网页仪表盘

在浏览器中打开 `http://localhost:56243` 即可查看仪表盘。

---

## 注册 API 密钥

注册 API 密钥有四种方式。**对于初次使用的用户，推荐使用方式 1（环境变量）**。

### 方式 1：环境变量（推荐 — 最简单）

环境变量是程序启动时读取的**预设值**。在终端中按如下方式输入即可。

```bash
# 注册 Google Gemini 密钥
export WV_KEY_GOOGLE=AIzaSy...

# 注册 OpenRouter 密钥
export WV_KEY_OPENROUTER=sk-or-v1-...

# 注册后启动
wall-vault start
```

如果有多个密钥，用逗号（,）连接。wall-vault 会自动轮流使用这些密钥（轮询）：

```bash
export WV_KEY_GOOGLE=AIzaSy...,AIzaSy...,AIzaSy...
```

> 💡 **提示**：`export` 命令只对当前终端会话有效。要在重启计算机后依然生效，请将上述行添加到 `~/.bashrc` 或 `~/.zshrc` 文件中。

### 方式 2：仪表盘 UI（鼠标点击操作）

1. 在浏览器中访问 `http://localhost:56243`
2. 在顶部 **🔑 API 密钥** 卡片中点击 `[+ 添加]` 按钮
3. 输入服务类型、密钥值、标签（备注名称）和每日限额，然后保存

### 方式 3：REST API（自动化/脚本使用）

REST API 是程序之间通过 HTTP 互相传递数据的方式。适合用脚本自动批量注册密钥。

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

### 方式 4：proxy 标志（临时测试使用）

不正式注册，临时填入密钥进行测试时使用。程序退出后密钥即消失。

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

> 💡 **更简单的方法**：点击仪表盘代理卡片上的 **🦞 OpenClaw 设置复制** 按钮，包含令牌和地址的配置片段会自动复制到剪贴板。直接粘贴即可。

**模型名称前的 `wall-vault/` 会连接到哪里？**

wall-vault 根据模型名称自动判断应将请求发送到哪个 AI 服务：

| 模型格式 | 连接的服务 |
|----------|-----------|
| `wall-vault/gemini-*` | 直连 Google Gemini |
| `wall-vault/gpt-*`、`wall-vault/o3`、`wall-vault/o4*` | 直连 OpenAI |
| `wall-vault/claude-*` | 通过 OpenRouter 连接 Anthropic |
| `wall-vault/hunter-alpha`、`wall-vault/healer-alpha` | OpenRouter（免费 100 万 Token 上下文） |
| `wall-vault/kimi-*`、`wall-vault/glm-*`、`wall-vault/deepseek-*` | 连接 OpenRouter |
| `google/模型名`、`openai/模型名`、`anthropic/模型名` 等 | 直连对应服务 |
| `custom/google/模型名`、`custom/openai/模型名` 等 | 去除 `custom/` 前缀后重新路由 |
| `模型名:cloud` | 去除 `:cloud` 后连接 OpenRouter |

> 💡 **什么是上下文（context）？** AI 一次能记住的对话量。1M（百万 Token）意味着非常长的对话或长文档都可以一次性处理。

### 以 Gemini API 格式直接连接（兼容现有工具）

如果您有直接使用 Google Gemini API 的工具，只需将地址改为 wall-vault 即可：

```bash
export ANTHROPIC_BASE_URL=http://localhost:56244/google
```

或者对于可以直接指定 URL 的工具：

```
http://localhost:56244/google/v1beta/models/gemini-2.5-flash:generateContent
```

### 在 OpenAI SDK（Python）中使用

在 Python 代码中使用 AI 时，同样可以接入 wall-vault。只需修改 `base_url` 即可：

```python
from openai import OpenAI

client = OpenAI(
    base_url="http://localhost:56244/v1",
    api_key="not-needed"  # API 密钥由 wall-vault 自动管理
)

response = client.chat.completions.create(
    model="google/gemini-2.5-flash",   # 使用 provider/model 格式输入
    messages=[{"role": "user", "content": "你好"}]
)
```

### 运行中切换模型

在 wall-vault 已经运行的状态下切换 AI 模型：

```bash
# 直接向代理发起请求切换模型
curl -X PUT http://localhost:56244/api/config/model \
  -H "Content-Type: application/json" \
  -d '{"service": "openrouter", "model": "anthropic/claude-3.5-sonnet"}'

# 分布式模式（多机器人）下在金库服务器修改 → 通过 SSE 即时生效
curl -X PUT http://localhost:56243/admin/clients/我的机器人-id \
  -H "Authorization: Bearer 管理员令牌" \
  -H "Content-Type: application/json" \
  -d '{"default_service": "google", "default_model": "gemini-2.5-pro"}'
```

### 查看可用模型列表

```bash
# 查看全部列表
curl http://localhost:56244/api/models | python3 -m json.tool

# 只查看 Google 模型
curl "http://localhost:56244/api/models?service=google"

# 按名称搜索（例如：包含"claude"的模型）
curl "http://localhost:56244/api/models?q=claude"
```

**各服务主要模型汇总：**

| 服务 | 主要模型 |
|------|---------|
| Google | gemini-2.5-pro、gemini-2.5-flash、gemini-2.5-flash-8b、gemini-2.0-flash |
| OpenAI | gpt-4o、gpt-4o-mini、o3、o1、o1-mini |
| OpenRouter | 346 个以上（Hunter Alpha 1M 上下文免费、DeepSeek R1/V3、Qwen 2.5 等） |
| Ollama | 自动检测本机安装的本地服务器 |

---

## 密钥金库仪表盘

在浏览器中访问 `http://localhost:56243` 即可查看仪表盘。

**界面组成：**
- **顶部固定栏（topbar）**：Logo、语言和主题选择器、SSE 连接状态显示
- **卡片网格**：代理、服务、API 密钥卡片以磁贴形式排列

### API 密钥卡片

一目了然地管理已注册 API 密钥的卡片。

- 按服务分类显示密钥列表。
- `today_usage`：今日成功处理的 Token 数（AI 读取和生成的字符数）
- `today_attempts`：今日总调用次数（包含成功和失败）
- 点击 `[+ 添加]` 按钮注册新密钥，点击 `✕` 删除密钥。

> 💡 **什么是 Token（令牌）？** Token 是 AI 处理文本时使用的计量单位。大约相当于一个英文单词，或 1～2 个中文字符。API 费用通常按 Token 数量计算。

### 代理卡片

显示已连接到 wall-vault 代理的机器人（代理）状态的卡片。

**连接状态分为 4 个等级：**

| 显示 | 状态 | 含义 |
|------|------|------|
| 🟢 | 运行中 | 代理正常运行 |
| 🟡 | 延迟 | 有响应但速度慢 |
| 🔴 | 离线 | 代理无响应 |
| ⚫ | 未连接/未激活 | 代理从未连接到金库或已被停用 |

**代理卡片底部按钮说明：**

注册代理时指定**代理类型**后，会自动显示该类型对应的快捷按钮。

---

#### 🔘 设置复制按钮 — 自动生成连接配置

点击按钮后，包含该代理的令牌、代理地址和模型信息的配置片段会复制到剪贴板。将复制的内容粘贴到下表对应位置，即可完成连接配置。

| 按钮 | 代理类型 | 粘贴位置 |
|------|---------|---------|
| 🦞 OpenClaw 设置复制 | `openclaw` | `~/.openclaw/openclaw.json` |
| 🦀 NanoClaw 设置复制 | `nanoclaw` | `~/.openclaw/openclaw.json` |
| 🟠 Claude Code 设置复制 | `claude-code` | `~/.claude/settings.json` |
| ⌨ Cursor 设置复制 | `cursor` | Cursor → Settings → AI |
| 💻 VSCode 设置复制 | `vscode` | `~/.continue/config.json` |

**示例 — Claude Code 类型时复制的内容：**

```json
// ~/.claude/settings.json
{
  "apiProvider": "openai",
  "baseUrl": "http://192.168.0.6:56244/v1",
  "apiKey": "此代理的令牌"
}
```

**示例 — VSCode（Continue）类型时：**

```json
// ~/.continue/config.json
{
  "models": [{
    "title": "wall-vault proxy",
    "provider": "openai",
    "model": "gemini-2.5-flash",
    "apiBase": "http://192.168.0.6:56244/v1",
    "apiKey": "此代理的令牌"
  }]
}
```

**示例 — Cursor 类型时：**

```
Base URL : http://192.168.0.6:56244/v1
API Key  : 此代理的令牌

// 或使用环境变量：
OPENAI_BASE_URL=http://192.168.0.6:56244/v1
OPENAI_API_KEY=此代理的令牌
```

> ⚠️ **剪贴板复制失败时**：浏览器安全策略有时会阻止复制操作。如果弹出文本框，请用 Ctrl+A 全选后再按 Ctrl+C 复制。

---

#### ⚡ 自动应用按钮 — 一键完成设置

当代理类型为 `cline`、`claude-code`、`openclaw`、`nanoclaw` 时，代理卡片上会显示 **⚡ 应用设置** 按钮。点击该按钮后，对应代理的本地配置文件会自动更新。

| 按钮 | 代理类型 | 应用目标文件 |
|------|---------|------------|
| ⚡ 应用 Cline 设置 | `cline` | `~/.cline/data/globalState.json` + `secrets.json` |
| ⚡ 应用 Claude Code 设置 | `claude-code` | `~/.claude/settings.json` |
| ⚡ 应用 OpenClaw 设置 | `openclaw` | `~/.openclaw/openclaw.json` |
| ⚡ 应用 NanoClaw 设置 | `nanoclaw` | `~/.openclaw/openclaw.json` |

> ⚠️ 该按钮会向 **localhost:56244**（本地代理）发送请求。只有在该机器上代理正在运行时才有效。

---

#### 🔀 拖拽排序代理卡片 (v0.1.17)

您可以通过**拖拽**仪表盘上的代理卡片，将它们重新排列为您想要的顺序。

1. 用鼠标按住代理卡片并拖动
2. 将其放到目标位置的卡片上，即可交换顺序
3. 更改后的顺序会**立即保存到服务器**，刷新页面后仍然保持

> 💡 触屏设备（手机/平板）暂不支持此功能，请使用桌面浏览器。

---

#### 🔄 双向模型同步 (v0.1.16)

在金库仪表盘中更改代理的模型后，对应代理的本地设置会自动更新。

**Cline 的情况：**
- 在金库中更改模型 → SSE 事件 → 代理更新 `globalState.json` 中的模型字段
- 更新对象：`actModeOpenAiModelId`、`planModeOpenAiModelId`、`openAiModelId`
- 不会修改 `openAiBaseUrl` 和 API 密钥
- **需要重新加载 VS Code（`Ctrl+Alt+R` 或 `Ctrl+Shift+P` → `Developer: Reload Window`）**
  - 因为 Cline 在运行期间不会重新读取配置文件

**Claude Code 的情况：**
- 在金库中更改模型 → SSE 事件 → 代理更新 `settings.json` 中的 `model` 字段
- 自动搜索 WSL 和 Windows 两侧的路径（`~/.claude/`、`/mnt/c/Users/*/.claude/`）

**反向同步（代理 → 金库）：**
- 当代理（Cline、Claude Code 等）向代理服务器发送请求时，代理服务器会在心跳中包含该客户端的服务和模型信息
- 金库仪表盘中的代理卡片会实时显示当前使用的服务/模型

> 💡 **要点**：代理服务器通过请求中的 Authorization 令牌识别代理，并自动路由到金库中设定的服务/模型。即使 Cline 或 Claude Code 发送了不同的模型名称，代理服务器也会用金库设置进行覆盖。

---

### 在 VS Code 中使用 Cline — 详细指南

#### 第 1 步：安装 Cline

在 VS Code 扩展市场中搜索并安装 **Cline**（ID：`saoudrizwan.claude-dev`）。

#### 第 2 步：在金库中注册代理

1. 打开金库仪表盘（`http://金库IP:56243`）
2. 在 **代理** 部分点击 **+ 添加**
3. 按如下内容填写：

| 字段 | 值 | 说明 |
|------|----|------|
| ID | `my_cline` | 唯一标识符（英文字母，不含空格） |
| 名称 | `My Cline` | 显示在仪表盘上的名称 |
| 代理类型 | `cline` | ← 必须选择 `cline` |
| 服务 | 选择要使用的服务（例：`google`） | |
| 模型 | 输入要使用的模型（例：`gemini-2.5-flash`） | |

4. 点击 **保存** 后令牌会自动生成

#### 第 3 步：连接到 Cline

**方式 A — 自动应用（推荐）**

1. 确认该机器上的 wall-vault **代理** 正在运行（`localhost:56244`）
2. 在仪表盘的代理卡片中点击 **⚡ 应用 Cline 设置** 按钮
3. 出现"设置应用完成！"提示即为成功
4. 重新加载 VS Code（`Ctrl+Alt+R`）

**方式 B — 手动配置**

打开 Cline 侧边栏的设置（⚙️）：
- **API Provider**：`OpenAI Compatible`
- **Base URL**：`http://代理地址:56244/v1`
  - 同一台机器：`http://localhost:56244/v1`
  - Mini 服务器等其他机器：`http://192.168.0.6:56244/v1`
- **API Key**：金库中颁发的令牌（从代理卡片复制）
- **Model ID**：金库中设置的模型（例：`gemini-2.5-flash`）

#### 第 4 步：验证

在 Cline 聊天窗口中发送任意消息。如果一切正常：
- 金库仪表盘中该代理卡片会显示 **绿色圆点（● 运行中）**
- 卡片上会显示当前服务/模型（例：`google / gemini-2.5-flash`）

#### 更改模型

如果要更改 Cline 的模型，请在 **金库仪表盘** 中操作：

1. 更改代理卡片中的服务/模型下拉菜单
2. 点击 **应用**
3. 重新加载 VS Code（`Ctrl+Alt+R`）— Cline 页脚中的模型名称会更新
4. 之后的请求将使用新模型

> 💡 实际上代理服务器通过令牌识别 Cline 的请求，并路由到金库设置中指定的模型。即使不重新加载 VS Code，**实际使用的模型也会立即切换** — 重新加载只是为了更新 Cline 界面中显示的模型名称。

#### 断线检测

关闭 VS Code 后，金库仪表盘中的代理卡片会在约 **90 秒** 后变为黄色（延迟），**3 分钟** 后变为红色（离线）。（自 v0.1.18 起，每 15 秒状态检查间隔使离线检测更加迅速。）

#### 故障排除

| 症状 | 原因 | 解决方法 |
|------|------|---------|
| Cline 中出现"连接失败"错误 | 代理未运行或地址有误 | 使用 `curl http://localhost:56244/health` 检查代理 |
| 金库中未显示绿色圆点 | 未设置 API 密钥（令牌） | 再次点击 **⚡ 应用 Cline 设置** 按钮 |
| Cline 页脚的模型未更新 | Cline 缓存了设置 | 重新加载 VS Code（`Ctrl+Alt+R`） |
| 显示了错误的模型名称 | 旧版 Bug（v0.1.16 已修复） | 将代理更新到 v0.1.16 或更高版本 |

---

#### 🟣 部署命令复制按钮 — 在新机器上安装时使用

在新计算机上首次安装 wall-vault 代理并连接到金库时使用。点击按钮后，完整的安装脚本会被复制。在新计算机的终端中粘贴并运行，将一次性完成以下操作：

1. 安装 wall-vault 二进制文件（如果已安装则跳过）
2. 自动注册 systemd 用户服务
3. 启动服务并自动连接到金库

> 💡 脚本中已预填了该代理的令牌和金库服务器地址，粘贴后无需额外修改即可直接运行。

---

### 服务卡片

用于开启/关闭或配置要使用的 AI 服务的卡片。

- 每个服务的启用/禁用切换开关
- 输入本地 AI 服务器（运行在本机的 Ollama、LM Studio、vLLM 等）的地址后，会自动发现可用模型。
- **本地服务连接状态显示**：服务名称旁边的 ● 圆点**绿色**表示已连接，**灰色**表示未连接
- **复选框自动同步**：打开页面时，如果本地服务（Ollama 等）正在运行，复选框会自动勾选。

> 💡 **如果本地服务运行在其他计算机上**：在服务 URL 输入框中输入那台计算机的 IP。例如：`http://192.168.0.6:11434`（Ollama）、`http://192.168.0.6:1234`（LM Studio）

### 管理员令牌输入

在仪表盘中使用添加/删除密钥等重要功能时，会弹出管理员令牌输入框。请输入在 setup 向导中设置的令牌。输入一次后，在关闭浏览器之前会持续保持认证状态。

> ⚠️ **15 分钟内认证失败超过 10 次，该 IP 将被临时封禁。** 如果忘记了令牌，请在 `wall-vault.yaml` 文件中查看 `admin_token` 配置项。

---

## 分布式模式（多机器人）

在多台计算机上同时运行 OpenClaw 时，采用**共享同一个密钥金库**的配置。只需在一处管理密钥，非常方便。

### 配置示例

```
[密钥金库服务器]
  wall-vault vault    （密钥金库 :56243，仪表盘）

[WSL Alpha]            [树莓派 Gamma]         [Mac Mini 本地]
  wall-vault proxy      wall-vault proxy        wall-vault proxy
  openclaw TUI          openclaw TUI            openclaw TUI
  ↕ SSE 同步            ↕ SSE 同步              ↕ SSE 同步
```

所有机器人都指向中央金库服务器，在金库中切换模型或添加密钥后，所有机器人会立即生效。

### 第 1 步：启动密钥金库服务器

在作为金库服务器的计算机上运行：

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

### 第 3 步：在各机器人计算机上启动代理

在安装了机器人的各台计算机上，指定金库服务器地址和令牌来运行代理：

```bash
WV_VAULT_URL=http://192.168.x.x:56243 \
WV_VAULT_TOKEN=bota-secret \
WV_VAULT_CLIENT_ID=botA \
wall-vault proxy
```

> 💡 请将 **`192.168.x.x`** 替换为金库服务器计算机的实际内网 IP 地址。可通过路由器设置或 `ip addr` 命令查看。

---

## 设置自动启动

如果每次重启计算机都要手动启动 wall-vault 觉得麻烦，可以将其注册为系统服务。注册一次后，每次开机会自动启动。

### Linux — systemd（大多数 Linux 发行版）

systemd 是 Linux 上自动启动和管理程序的系统：

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

macOS 上负责程序自动运行的系统：

```bash
wall-vault doctor deploy launchd
launchctl load ~/Library/LaunchAgents/com.wall-vault.plist
```

### Windows — NSSM

1. 从 [nssm.cc](https://nssm.cc/download) 下载 NSSM 并添加到 PATH。
2. 在管理员权限的 PowerShell 中运行：

```powershell
wall-vault doctor deploy windows
```

---

## Doctor 诊断工具

`doctor` 命令是一个**自我诊断并修复** wall-vault 配置的工具。

```bash
wall-vault doctor check   # 诊断当前状态（只读，不做任何修改）
wall-vault doctor fix     # 自动修复问题
wall-vault doctor all     # 诊断 + 自动修复，一步完成
```

> 💡 如果感觉有些不对劲，请先运行 `wall-vault doctor all`。很多问题都能自动解决。

---

## 环境变量参考

环境变量是向程序传递配置值的方式。在终端中以 `export 变量名=值` 的形式输入，或写入自动启动服务文件中，即可持续生效。

| 变量 | 说明 | 示例值 |
|------|------|--------|
| `WV_LANG` | 仪表盘语言 | `ko`、`en`、`ja` |
| `WV_THEME` | 仪表盘主题 | `light`、`dark`、`gold` |
| `WV_KEY_GOOGLE` | Google API 密钥（多个用逗号分隔） | `AIza...,AIza...` |
| `WV_KEY_OPENROUTER` | OpenRouter API 密钥 | `sk-or-v1-...` |
| `WV_VAULT_URL` | 分布式模式中金库服务器地址 | `http://192.168.x.x:56243` |
| `WV_VAULT_TOKEN` | 客户端（机器人）认证令牌 | `my-secret-token` |
| `WV_ADMIN_TOKEN` | 管理员令牌 | `admin-token-here` |
| `WV_MASTER_PASS` | API 密钥加密密码 | `my-password` |
| `WV_AVATAR` | 头像图片文件路径（相对于 `~/.openclaw/` 的相对路径） | `workspace/avatars/avatar.png` |
| `OLLAMA_URL` | Ollama 本地服务器地址 | `http://192.168.x.x:11434` |

---

## 故障排除

### 代理无法启动时

通常是端口已被其他程序占用。

```bash
ss -tlnp | grep 56244   # 查看是谁在使用 56244 端口
wall-vault proxy --port 8080   # 使用其他端口号启动
```

### 出现 API 密钥错误时（429、402、401、403、582）

| 错误代码 | 含义 | 处理方法 |
|---------|------|---------|
| **429** | 请求过多（超出使用量） | 稍等片刻或添加其他密钥 |
| **402** | 需要付款或余额不足 | 在对应服务中充值 |
| **401 / 403** | 密钥错误或权限不足 | 重新确认密钥值后重新注册 |
| **582** | 网关过载（冷却 5 分钟） | 5 分钟后自动解除 |

```bash
# 查看已注册的密钥列表及状态
curl -H "Authorization: Bearer 管理员令牌" http://localhost:56243/admin/keys

# 重置密钥使用量计数器
curl -X POST -H "Authorization: Bearer 管理员令牌" http://localhost:56243/admin/keys/reset
```

### 代理显示"未连接"时

"未连接"表示代理进程没有向金库发送心跳信号（heartbeat）。**这不代表设置没有保存。** 代理必须知道金库服务器地址和令牌并运行，才能变为连接状态。

```bash
# 指定金库服务器地址、令牌和客户端 ID 来启动代理
WV_VAULT_URL=http://金库服务器地址:56243 \
WV_VAULT_TOKEN=客户端令牌 \
WV_VAULT_CLIENT_ID=客户端ID \
wall-vault proxy
```

连接成功后，约 20 秒内仪表盘中会变为 🟢 运行中。

### Ollama 无法连接时

Ollama 是直接在本地计算机上运行 AI 的程序。请先确认 Ollama 是否已启动。

```bash
curl http://localhost:11434/api/tags   # 如果显示模型列表则正常
export OLLAMA_URL=http://192.168.x.x:11434   # 如果运行在其他计算机上
```

> ⚠️ 如果 Ollama 没有响应，请先用 `ollama serve` 命令启动 Ollama。

> ⚠️ **大型模型响应较慢**：`qwen3.5:35b`、`deepseek-r1` 等大型模型可能需要数分钟才能生成响应。即使看起来没有响应，也可能正在正常处理中，请耐心等待。

---

*更详细的 API 信息请参考 [API.md](API.md)。*
