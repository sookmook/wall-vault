# wall-vault ユーザーマニュアル
*(Last updated: 2026-04-09 — v0.1.27)*

---

## 目次

1. [wall-vaultとは？](#wall-vaultとは)
2. [インストール](#インストール)
3. [はじめに（セットアップウィザード）](#はじめに)
4. [APIキーの登録](#apiキーの登録)
5. [プロキシの使い方](#プロキシの使い方)
6. [キー金庫ダッシュボード](#キー金庫ダッシュボード)
7. [分散モード（マルチボット）](#分散モードマルチボット)
8. [自動起動設定](#自動起動設定)
9. [Doctor（ドクター）](#doctorドクター)
10. [RTKトークン節約](#rtkトークン節約)
11. [環境変数リファレンス](#環境変数リファレンス)
12. [トラブルシューティング](#トラブルシューティング)

---

## wall-vaultとは？

**wall-vault = OpenClaw用のAIプロキシ + APIキー金庫**

AIサービスを利用するには**APIキー**が必要です。APIキーとは「この人はこのサービスを使う資格がある」ことを証明する**デジタル入館証**のようなものです。しかし、この入館証には1日の利用回数に制限があり、管理を誤ると漏洩のリスクもあります。

wall-vaultはこれらの入館証を安全な金庫に保管し、OpenClawとAIサービスの間で**プロキシ（代理人）**の役割を果たします。つまり、OpenClawはwall-vaultに接続するだけでよく、残りの複雑な処理はwall-vaultが自動的に行います。

wall-vaultが解決する問題：

- **APIキーの自動ローテーション**：あるキーの使用量が上限に達したり一時的にブロックされた（クールダウン）場合、静かに次のキーに切り替えます。OpenClawは中断なく動作し続けます。
- **サービスの自動フォールバック**：Googleが応答しなければOpenRouterへ、それもダメならローカルにインストールされたAI（Ollama、LM Studio、vLLM）へ自動的に切り替わります。セッションは途切れません。元のサービスが復旧すると、次のリクエストから自動的に戻ります（v0.1.18+、LM Studio/vLLM: v0.1.21+）。
- **リアルタイム同期（SSE）**：金庫ダッシュボードでモデルを変更すると、1〜3秒以内にOpenClawの画面に反映されます。SSE（Server-Sent Events）とは、サーバーがクライアントにリアルタイムで変更をプッシュする技術です。
- **リアルタイム通知**：キーの枯渇やサービス障害などのイベントが、OpenClawのTUI（ターミナル画面）下部にすぐ表示されます。

> :bulb: **Claude Code、Cursor、VS Code**も接続して使えますが、wall-vaultの本来の目的はOpenClawと一緒に使うことです。

```
OpenClaw（TUIターミナル画面）
        |
        v
  wall-vault プロキシ (:56244)   <- キー管理、ルーティング、フォールバック、イベント
        |
        +-- Google Gemini API
        +-- OpenRouter API（340以上のモデル）
        +-- Ollama / LM Studio / vLLM（ローカルマシン、最後の砦）
        +-- OpenAI / Anthropic API
```

---

## インストール

### Linux / macOS

ターミナルを開き、以下のコマンドをそのまま貼り付けてください。

```bash
# Linux（一般的なPC、サーバー — amd64）
curl -L https://github.com/sookmook/wall-vault/releases/latest/download/wall-vault-linux-amd64 \
  -o ~/.local/bin/wall-vault && chmod +x ~/.local/bin/wall-vault

# macOS Apple Silicon（M1/M2/M3 Mac）
curl -L https://github.com/sookmook/wall-vault/releases/latest/download/wall-vault-darwin-arm64 \
  -o /usr/local/bin/wall-vault && chmod +x /usr/local/bin/wall-vault
```

- `curl -L ...` — インターネットからファイルをダウンロードします。
- `chmod +x` — ダウンロードしたファイルを「実行可能」にします。この手順を省略すると「権限がありません」エラーが出ます。

### Windows

PowerShell（管理者権限）を開き、以下のコマンドを実行してください。

```powershell
# ダウンロード
Invoke-WebRequest -Uri `
  "https://github.com/sookmook/wall-vault/releases/latest/download/wall-vault-windows-amd64.exe" `
  -OutFile "$env:LOCALAPPDATA\Programs\wall-vault\wall-vault.exe"

# PATHに追加（PowerShell再起動後に適用）
$env:PATH += ";$env:LOCALAPPDATA\Programs\wall-vault"
```

> :bulb: **PATHとは？** コンピューターがコマンドを探すフォルダのリストです。PATHに追加することで、どのフォルダからでも`wall-vault`と入力して実行できるようになります。

### ソースからビルド（開発者向け）

Go言語の開発環境がインストールされている場合のみ該当します。

```bash
git clone https://github.com/sookmook/wall-vault
cd wall-vault
make build       # bin/wall-vault（バージョン: v0.1.25.YYYYMMDD.HHmmss）
make install     # ~/.local/bin/wall-vault
```

> :bulb: **ビルドタイムスタンプバージョン**：`make build`でビルドすると、`v0.1.27.20260409`のように日付・時刻を含む形式でバージョンが自動生成されます。`go build ./...`で直接ビルドすると、バージョンは`"dev"`とだけ表示されます。

---

## はじめに

### セットアップウィザードの実行

インストール後、最初に必ず以下のコマンドで**セットアップウィザード**を実行してください。ウィザードが必要な項目を一つずつ質問しながら案内してくれます。

```bash
wall-vault setup
```

ウィザードが進行するステップは以下の通りです：

```
1. 言語選択（日本語を含む10言語）
2. テーマ選択（light / dark / gold / cherry / ocean）
3. 運用モード — スタンドアロン（1台で使用）またはディストリビューテッド（複数台で共有）
4. ボット名 — ダッシュボードに表示される名前
5. ポート設定 — デフォルト: プロキシ 56244、金庫 56243（変更不要ならそのままEnter）
6. AIサービス選択 — Google / OpenRouter / Ollama / LM Studio / vLLMから選択
7. ツールセキュリティフィルター設定
8. 管理者トークン — ダッシュボードの管理機能をロックするパスワード。自動生成も可能
9. APIキー暗号化パスワード — キーをより安全に保存したい場合（任意）
10. 設定ファイルの保存先
```

> :warning: **管理者トークンは必ず覚えておいてください。** 後でダッシュボードでキーを追加したり設定を変更する際に必要です。忘れた場合は設定ファイルを直接編集する必要があります。

ウィザードが完了すると、`wall-vault.yaml`設定ファイルが自動的に作成されます。

### 起動

```bash
wall-vault start
```

以下の2つのサーバーが同時に起動します：

- **プロキシ**（`http://localhost:56244`）— OpenClawとAIサービスを繋ぐ代理人
- **キー金庫**（`http://localhost:56243`）— APIキー管理およびWebダッシュボード

ブラウザで`http://localhost:56243`を開くと、ダッシュボードにすぐアクセスできます。

---

## APIキーの登録

APIキーを登録する方法は4つあります。**初めての方には方法1（環境変数）をお勧めします。**

### 方法1：環境変数（推奨 — 最も簡単）

環境変数とは、プログラムの起動時に読み込まれる**事前設定された値**です。ターミナルで以下のように入力します。

```bash
# Google Geminiキーを登録
export WV_KEY_GOOGLE=AIzaSy...

# OpenRouterキーを登録
export WV_KEY_OPENROUTER=sk-or-v1-...

# 登録後に起動
wall-vault start
```

キーを複数持っている場合は、カンマ（,）で区切ってください。wall-vaultがキーを順番に自動使用します（ラウンドロビン）：

```bash
export WV_KEY_GOOGLE=AIzaSy...,AIzaSy...,AIzaSy...
```

> :bulb: **ヒント**：`export`コマンドは現在のターミナルセッションにのみ適用されます。再起動後も維持したい場合は、`~/.bashrc`または`~/.zshrc`ファイルに上記の行を追加してください。

### 方法2：ダッシュボードUI（マウスでクリック）

1. ブラウザで`http://localhost:56243`にアクセス
2. 上部の**:key: APIキー**カードで`[+ 追加]`ボタンをクリック
3. サービス種類、キー値、ラベル（メモ用の名前）、日次制限を入力して保存

### 方法3：REST API（自動化・スクリプト用）

REST APIとは、プログラム同士がHTTPでデータをやり取りする方式です。スクリプトで自動登録する際に便利です。

```bash
curl -X POST http://localhost:56243/admin/keys \
  -H "Authorization: Bearer 管理者トークン" \
  -H "Content-Type: application/json" \
  -d '{
    "service": "google",
    "key": "AIzaSy...",
    "label": "メインキー",
    "daily_limit": 1000
  }'
```

### 方法4：proxyフラグ（一時的なテスト用）

正式登録なしに一時的にキーを入れてテストしたい場合に使います。プログラムを終了すると消えます。

```bash
wall-vault proxy --key-google=AIzaSy... --key-openrouter=sk-or-...
```

---

## プロキシの使い方

### OpenClawでの使用（主目的）

OpenClawがwall-vaultを通じてAIサービスに接続するための設定方法です。

`~/.openclaw/openclaw.json`ファイルを開き、以下の内容を追加してください：

```json5
// ~/.openclaw/openclaw.json
{
  models: {
    providers: {
      "wall-vault": {
        baseUrl: "http://localhost:56244/v1",
        apiKey: "your-agent-token",   // vault エージェントトークン
        api: "openai-completions",
        models: [
          { id: "wall-vault/gemini-2.5-flash" },
          { id: "wall-vault/gemini-2.5-pro" },
          { id: "wall-vault/hunter-alpha" },    // 無料 1M コンテキスト
          { id: "wall-vault/claude-opus-4-6" }
        ]
      }
    }
  }
}
```

> :bulb: **もっと簡単な方法**：ダッシュボードのエージェントカードにある**:lobster: OpenClaw設定コピー**ボタンを押すと、トークンとアドレスが入力済みのスニペットがクリップボードにコピーされます。貼り付けるだけで完了です。

**モデル名の前の`wall-vault/`はどこに接続されるのか？**

モデル名を見て、wall-vaultがどのAIサービスにリクエストを送るかを自動判断します：

| モデル形式 | 接続先サービス |
|-----------|--------------|
| `wall-vault/gemini-*` | Google Geminiに直接接続 |
| `wall-vault/gpt-*`, `wall-vault/o3`, `wall-vault/o4*` | OpenAIに直接接続 |
| `wall-vault/claude-*` | OpenRouter経由でAnthropicに接続 |
| `wall-vault/hunter-alpha`, `wall-vault/healer-alpha` | OpenRouter（無料100万トークンコンテキスト） |
| `wall-vault/kimi-*`, `wall-vault/glm-*`, `wall-vault/deepseek-*` | OpenRouterに接続 |
| `google/モデル名`, `openai/モデル名`, `anthropic/モデル名`など | 該当サービスに直接接続 |
| `custom/google/モデル名`, `custom/openai/モデル名`など | `custom/`部分を除去して再ルーティング |
| `モデル名:cloud` | `:cloud`部分を除去してOpenRouterに接続 |

> :bulb: **コンテキストとは？** AIが一度に記憶できる会話の量です。1M（100万トークン）なら、非常に長い会話や長い文書も一度に処理できます。

### Gemini API形式での直接接続（既存ツール互換）

Google Gemini APIを直接使っているツールがある場合、URLをwall-vaultに変更するだけです：

```bash
export ANTHROPIC_BASE_URL=http://localhost:56244/google
```

またはURLを直接指定するツールの場合：

```
http://localhost:56244/google/v1beta/models/gemini-2.5-flash:generateContent
```

### OpenAI SDK（Python）での使用

PythonでAIを活用するコードでもwall-vaultを接続できます。`base_url`を変更するだけです：

```python
from openai import OpenAI

client = OpenAI(
    base_url="http://localhost:56244/v1",
    api_key="not-needed"  # APIキーはwall-vaultが自動管理します
)

response = client.chat.completions.create(
    model="google/gemini-2.5-flash",   # provider/model形式で入力
    messages=[{"role": "user", "content": "こんにちは"}]
)
```

### 実行中にモデルを変更する

wall-vaultが既に実行中の状態でAIモデルを変更するには：

```bash
# プロキシに直接リクエストしてモデル変更
curl -X PUT http://localhost:56244/api/config/model \
  -H "Content-Type: application/json" \
  -d '{"service": "openrouter", "model": "anthropic/claude-3.5-sonnet"}'

# 分散モード（マルチボット）では金庫サーバーで変更 → SSEで即時反映
curl -X PUT http://localhost:56243/admin/clients/my-bot-id \
  -H "Authorization: Bearer 管理者トークン" \
  -H "Content-Type: application/json" \
  -d '{"default_service": "google", "default_model": "gemini-2.5-pro"}'
```

### 利用可能なモデル一覧の確認

```bash
# 全リスト表示
curl http://localhost:56244/api/models | python3 -m json.tool

# Googleモデルのみ表示
curl "http://localhost:56244/api/models?service=google"

# 名前で検索（例：「claude」を含むモデル）
curl "http://localhost:56244/api/models?q=claude"
```

**サービス別の主要モデル：**

| サービス | 主要モデル |
|---------|----------|
| Google | gemini-2.5-pro, gemini-2.5-flash, gemini-2.5-flash-8b, gemini-2.0-flash |
| OpenAI | gpt-4o, gpt-4o-mini, o3, o1, o1-mini |
| OpenRouter | 346以上（Hunter Alpha 1Mコンテキスト無料、DeepSeek R1/V3、Qwen 2.5など） |
| Ollama | ローカルにインストールされたサーバーから自動検出 |
| LM Studio | ローカルサーバー（ポート1234） |
| vLLM | ローカルサーバー（ポート8000） |

---

## キー金庫ダッシュボード

ブラウザで`http://localhost:56243`にアクセスすると、ダッシュボードが表示されます。

**画面構成：**
- **上部固定バー（topbar）**：ロゴ、言語・テーマセレクター、SSE接続状態表示
- **カードグリッド**：エージェント・サービス・APIキーカードがタイル形式で配置

### APIキーカード

登録されたAPIキーを一覧で管理できるカードです。

- サービス別にキーリストを分類表示します。
- `today_usage`：今日正常に処理されたトークン（AIが読み書きした文字数）の数
- `today_attempts`：今日の総呼び出し回数（成功 + 失敗を含む）
- `[+ 追加]`ボタンで新しいキーを登録し、`x`でキーを削除します。

> :bulb: **トークンとは？** AIがテキストを処理する際に使用する単位です。おおよそ英語1単語、または日本語1〜2文字に相当します。API料金は通常このトークン数に基づいて計算されます。

### エージェントカード

wall-vaultプロキシに接続されたボット（エージェント）の状態を表示するカードです。

**接続状態は4段階で表示されます：**

| 表示 | 状態 | 意味 |
|------|------|------|
| :green_circle: | 実行中 | プロキシが正常に動作中 |
| :yellow_circle: | 遅延 | 応答はあるが遅い |
| :red_circle: | オフライン | プロキシが応答していない |
| :black_circle: | 未接続・無効 | プロキシが金庫に接続されたことがない、または無効化されている |

**エージェントカード下部のボタン案内：**

エージェントを登録する際に**エージェント種類**を指定すると、その種類に応じた便利ボタンが自動的に表示されます。

---

#### :radio_button: 設定コピーボタン — 接続設定を自動生成します

ボタンをクリックすると、そのエージェントのトークン、プロキシアドレス、モデル情報が入力済みの設定スニペットがクリップボードにコピーされます。コピーした内容を以下の表の場所に貼り付けるだけで接続設定が完了します。

| ボタン | エージェント種類 | 貼り付け先 |
|--------|---------------|-----------|
| :lobster: OpenClaw設定コピー | `openclaw` | `~/.openclaw/openclaw.json` |
| :crab: NanoClaw設定コピー | `nanoclaw` | `~/.openclaw/openclaw.json` |
| :orange_circle: Claude Code設定コピー | `claude-code` | `~/.claude/settings.json` |
| :keyboard: Cursor設定コピー | `cursor` | Cursor -> Settings -> AI |
| :computer: VSCode設定コピー | `vscode` | `~/.continue/config.json` |

**例 — Claude Codeタイプの場合、以下の内容がコピーされます：**

```json
// ~/.claude/settings.json
{
  "apiProvider": "openai",
  "baseUrl": "http://192.168.0.6:56244/v1",
  "apiKey": "このエージェントのトークン"
}
```

**例 — VSCode（Continue）タイプの場合：**

```yaml
# ~/.continue/config.yaml  <- config.jsonではなくconfig.yamlに貼り付け
name: My Config
version: 0.0.1
schema: v1

models:
  - name: wall-vault proxy
    provider: openai
    model: gemini-2.5-flash
    apiBase: http://192.168.0.6:56244/v1
    apiKey: このエージェントのトークン
    roles:
      - chat
      - edit
      - apply
```

> :warning: **Continueの最新バージョンは`config.yaml`を使用します。** `config.yaml`が存在する場合、`config.json`は完全に無視されます。必ず`config.yaml`に貼り付けてください。

**例 — Cursorタイプの場合：**

```
Base URL : http://192.168.0.6:56244/v1
API Key  : このエージェントのトークン

// または環境変数:
OPENAI_BASE_URL=http://192.168.0.6:56244/v1
OPENAI_API_KEY=このエージェントのトークン
```

> :warning: **クリップボードにコピーできない場合**：ブラウザのセキュリティポリシーによりコピーがブロックされることがあります。ポップアップでテキストボックスが表示されたら、Ctrl+Aで全選択してCtrl+Cでコピーしてください。

---

#### :zap: 自動適用ボタン — ワンクリックで設定完了

エージェント種類が`cline`、`claude-code`、`openclaw`、`nanoclaw`の場合、エージェントカードに**:zap: 設定適用**ボタンが表示されます。このボタンを押すと、そのエージェントのローカル設定ファイルが自動的に更新されます。

| ボタン | エージェント種類 | 適用先ファイル |
|--------|---------------|-------------|
| :zap: Cline設定適用 | `cline` | `~/.cline/data/globalState.json` + `secrets.json` |
| :zap: Claude Code設定適用 | `claude-code` | `~/.claude/settings.json` |
| :zap: OpenClaw設定適用 | `openclaw` | `~/.openclaw/openclaw.json` |
| :zap: NanoClaw設定適用 | `nanoclaw` | `~/.openclaw/openclaw.json` |

> :warning: このボタンは**localhost:56244**（ローカルプロキシ）にリクエストを送信します。そのマシンでプロキシが実行中である必要があります。

---

#### :twisted_rightwards_arrows: ドラッグ&ドロップカード並べ替え（v0.1.17、改善 v0.1.25）

ダッシュボードのエージェントカードを**ドラッグ**して、好きな順序に並べ替えられます。

1. カード左上の**信号機（●）**エリアをマウスで掴んでドラッグします
2. 目的の位置のカードの上にドロップすると順序が入れ替わります

> :bulb: カード本体（入力フィールド、ボタンなど）はドラッグできません。信号機エリアでのみ掴めます。

#### :orange_circle: エージェントプロセス検出（v0.1.25）

プロキシは正常に動作しているが、ローカルエージェントプロセス（NanoClaw、OpenClaw）が終了した場合、カードの信号機が**オレンジ色（点滅）**に変わり、「エージェントプロセス停止」メッセージが表示されます。

- :green_circle: 緑：プロキシ + エージェント正常
- :orange_circle: オレンジ（点滅）：プロキシ正常、エージェント停止
- :red_circle: 赤：プロキシオフライン
3. 変更された順序は**即座にサーバーに保存**され、ページを更新しても維持されます

> :bulb: タッチデバイス（モバイル/タブレット）はまだサポートされていません。デスクトップブラウザでご利用ください。

---

#### :arrows_counterclockwise: 双方向モデル同期（v0.1.16）

金庫ダッシュボードでエージェントのモデルを変更すると、そのエージェントのローカル設定が自動的に更新されます。

**Clineの場合：**
- 金庫でモデルを変更 → SSEイベント → プロキシが`globalState.json`のモデルフィールドを更新
- 更新対象：`actModeOpenAiModelId`、`planModeOpenAiModelId`、`openAiModelId`
- `openAiBaseUrl`とAPIキーは変更されません
- **VS Codeのリロードが必要です（`Ctrl+Alt+R`または`Ctrl+Shift+P` → `Developer: Reload Window`）**
  - Clineは実行中に設定ファイルを再読み込みしないため

**Claude Codeの場合：**
- 金庫でモデルを変更 → SSEイベント → プロキシが`settings.json`の`model`フィールドを更新
- WSLとWindowsの両方のパスを自動探索（`~/.claude/`、`/mnt/c/Users/*/.claude/`）

**逆方向（エージェント → 金庫）：**
- エージェント（Cline、Claude Codeなど）がプロキシにリクエストを送ると、プロキシがハートビートにそのクライアントのサービス・モデル情報を含めます
- 金庫ダッシュボードのエージェントカードに現在使用中のサービス/モデルがリアルタイムで表示されます

> :bulb: **ポイント**：プロキシはリクエストのAuthorizationトークンでエージェントを識別し、金庫に設定されたサービス/モデルに自動ルーティングします。ClineやClaude Codeが異なるモデル名を送っても、プロキシが金庫の設定でオーバーライドします。

---

### VS CodeでClineを使う — 詳細ガイド

#### ステップ1：Clineのインストール

VS Code拡張マーケットプレイスから**Cline**（ID: `saoudrizwan.claude-dev`）をインストールします。

#### ステップ2：金庫にエージェントを登録

1. 金庫ダッシュボード（`http://金庫IP:56243`）を開きます
2. **エージェント**セクションで**+ 追加**をクリック
3. 以下のように入力します：

| フィールド | 値 | 説明 |
|-----------|---|------|
| ID | `my_cline` | 一意の識別子（英数字、スペースなし） |
| 名前 | `My Cline` | ダッシュボードに表示される名前 |
| エージェント種類 | `cline` | ← 必ず`cline`を選択 |
| サービス | 使用するサービスを選択（例：`google`） | |
| モデル | 使用するモデルを入力（例：`gemini-2.5-flash`） | |

4. **保存**を押すとトークンが自動生成されます

#### ステップ3：Clineに接続

**方法A — 自動適用（推奨）**

1. このマシンでwall-vault**プロキシ**が実行中か確認（`localhost:56244`）
2. ダッシュボードのエージェントカードで**:zap: Cline設定適用**ボタンをクリック
3. 「設定適用完了！」通知が出れば成功
4. VS Codeをリロード（`Ctrl+Alt+R`）

**方法B — 手動設定**

Clineサイドバーで設定（:gear:）を開き：
- **API Provider**：`OpenAI Compatible`
- **Base URL**：`http://プロキシアドレス:56244/v1`
  - 同じマシンなら`http://localhost:56244/v1`
  - Macミニなど別のマシンなら`http://192.168.0.6:56244/v1`
- **API Key**：金庫で発行されたトークン（エージェントカードからコピー）
- **Model ID**：金庫で設定したモデル（例：`gemini-2.5-flash`）

#### ステップ4：確認

Clineのチャットウィンドウに何かメッセージを送ってみます。正常なら：
- 金庫ダッシュボードの該当エージェントカードに**緑の点（● 実行中）**が表示されます
- カードに現在のサービス/モデルが表示されます（例：`google / gemini-2.5-flash`）

#### モデルの変更

Clineのモデルを変えたい場合は**金庫ダッシュボード**で変更してください：

1. エージェントカードのサービス/モデルドロップダウンを変更
2. **適用**をクリック
3. VS Codeをリロード（`Ctrl+Alt+R`）— Clineフッターのモデル名が更新されます
4. 次のリクエストから新しいモデルが使用されます

> :bulb: 実際にはプロキシがClineのリクエストをトークンで識別し、金庫設定のモデルにルーティングします。VS Codeをリロードしなくても**実際に使用されるモデルは即座に変わります** — リロードはCline UIのモデル表示を更新するためのものです。

#### 切断検出

VS Codeを閉じると、金庫ダッシュボードのエージェントカードは約**90秒**後に黄色（遅延）に、**3分**後に赤色（オフライン）に変わります。（v0.1.18から15秒間隔の状態チェックによりオフライン検出が速くなりました。）

#### トラブルシューティング

| 症状 | 原因 | 解決 |
|------|------|------|
| Clineで「接続失敗」エラー | プロキシ未実行またはアドレス間違い | `curl http://localhost:56244/health`でプロキシを確認 |
| 金庫で緑の点が出ない | APIキー（トークン）が未設定 | **:zap: Cline設定適用**ボタンを再クリック |
| Clineフッターのモデルが変わらない | Clineが設定をキャッシュしている | VS Codeをリロード（`Ctrl+Alt+R`） |
| 間違ったモデル名が表示される | 旧バグ（v0.1.16で修正済み） | プロキシをv0.1.16以上にアップデート |

---

#### :purple_circle: デプロイコマンドコピーボタン — 新しいマシンにインストールする際に使用

新しいコンピューターにwall-vaultプロキシを初めてインストールして金庫に接続する際に使います。ボタンをクリックすると、インストールスクリプト全体がコピーされます。新しいコンピューターのターミナルに貼り付けて実行すると、以下が一度に処理されます：

1. wall-vaultバイナリのインストール（既にインストール済みならスキップ）
2. systemdユーザーサービスの自動登録
3. サービスの起動と金庫への自動接続

> :bulb: スクリプトにはこのエージェントのトークンと金庫サーバーアドレスが既に入力されているため、貼り付け後に別途修正なしですぐに実行できます。

---

### サービスカード

使用するAIサービスのオン/オフ切り替えや設定を行うカードです。

- サービスごとの有効化・無効化トグルスイッチ
- ローカルAIサーバー（自分のコンピューターで動かすOllama、LM Studio、vLLMなど）のアドレスを入力すると、利用可能なモデルを自動検出します。
- **ローカルサービス接続状態表示**：サービス名の横の●が**緑色**なら接続中、**灰色**なら未接続
- **ローカルサービス自動信号機**（v0.1.23+）：ローカルサービス（Ollama、LM Studio、vLLM）は接続可否に応じて自動的に有効化/無効化されます。サービスが到達可能になると15秒以内に●が緑色に変わりチェックボックスがオンになり、接続が切れると自動でオフになります。クラウドサービス（Google、OpenRouterなど）がAPIキーの有無で自動トグルされるのと同じ方式です。

> :bulb: **ローカルサービスが別のコンピューターで実行中の場合**：サービスURL入力欄にそのコンピューターのIPを入力してください。例：`http://192.168.0.6:11434`（Ollama）、`http://192.168.0.6:1234`（LM Studio）。サービスが`0.0.0.0`ではなく`127.0.0.1`にのみバインドされている場合、外部IPからのアクセスはできないため、サービス設定でバインドアドレスを確認してください。

### 管理者トークンの入力

ダッシュボードでキーの追加・削除のような重要な機能を使おうとすると、管理者トークン入力ポップアップが表示されます。セットアップウィザードで設定したトークンを入力してください。一度入力すると、ブラウザを閉じるまで維持されます。

> :warning: **認証失敗が15分以内に10回を超えると、そのIPが一時的にブロックされます。** トークンを忘れた場合は、`wall-vault.yaml`ファイルの`admin_token`項目を確認してください。

---

## 分散モード（マルチボット）

複数のコンピューターでOpenClawを同時に運用する際に、**1つのキー金庫を共有**する構成です。キー管理を1か所で行えるため便利です。

### 構成例

```
[キー金庫サーバー]
  wall-vault vault    （キー金庫 :56243、ダッシュボード）

[WSLアルファ]          [ラズベリーパイ ガンマ]  [Macミニ ローカル]
  wall-vault proxy      wall-vault proxy        wall-vault proxy
  openclaw TUI          openclaw TUI            openclaw TUI
  <-> SSE同期           <-> SSE同期             <-> SSE同期
```

すべてのボットが中央の金庫サーバーを参照しているため、金庫でモデルを変更したりキーを追加すると、すべてのボットに即座に反映されます。

### ステップ1：キー金庫サーバーの起動

金庫サーバーとして使うコンピューターで実行します：

```bash
wall-vault vault
```

### ステップ2：各ボット（クライアント）の登録

金庫サーバーに接続する各ボットの情報を事前に登録します：

```bash
curl -X POST http://localhost:56243/admin/clients \
  -H "Authorization: Bearer 管理者トークン" \
  -H "Content-Type: application/json" \
  -d '{
    "id": "botA",
    "name": "ボットA",
    "token": "bota-secret",
    "default_service": "google",
    "default_model": "gemini-2.5-flash"
  }'
```

### ステップ3：各ボットマシンでプロキシを起動

ボットがインストールされた各コンピューターで、金庫サーバーのアドレスとトークンを指定してプロキシを実行します：

```bash
WV_VAULT_URL=http://192.168.x.x:56243 \
WV_VAULT_TOKEN=bota-secret \
WV_VAULT_CLIENT_ID=botA \
wall-vault proxy
```

> :bulb: **`192.168.x.x`**の部分は、金庫サーバーマシンの実際の内部IPアドレスに置き換えてください。ルーター設定または`ip addr`コマンドで確認できます。

---

## 自動起動設定

コンピューターを再起動するたびに手動でwall-vaultを起動するのが面倒なら、システムサービスとして登録しましょう。一度登録すれば起動時に自動的に開始されます。

### Linux — systemd（ほとんどのLinuxディストリビューション）

systemdはLinuxでプログラムを自動的に起動・管理するシステムです：

```bash
wall-vault doctor deploy
systemctl --user daemon-reload
systemctl --user enable --now wall-vault
```

ログ確認：

```bash
journalctl --user -u wall-vault -f
```

### macOS — launchd

macOSでプログラムの自動実行を担当するシステムです：

```bash
wall-vault doctor deploy launchd
launchctl load ~/Library/LaunchAgents/com.wall-vault.plist
```

### Windows — NSSM

1. [nssm.cc](https://nssm.cc/download)からNSSMをダウンロードしてPATHに追加します。
2. 管理者権限のPowerShellで：

```powershell
wall-vault doctor deploy windows
```

---

## Doctor（ドクター）

`doctor`コマンドは、wall-vaultの設定が正しいかどうかを**自己診断して修復してくれるツール**です。

```bash
wall-vault doctor check   # 現在の状態を診断（読み取りのみ、何も変更しません）
wall-vault doctor fix     # 問題を自動修復
wall-vault doctor all     # 診断 + 自動修復を一度に実行
```

> :bulb: 何かおかしいと思ったら、まず`wall-vault doctor all`を実行してみてください。多くの問題を自動的に検出して修復します。

---

## RTKトークン節約

*(v0.1.24+)*

**RTK（トークン節約ツール）**は、AIコーディングエージェント（Claude Codeなど）が実行するシェルコマンドの出力を自動的に圧縮し、トークン使用量を削減します。例えば、`git status`の15行の出力が2行の要約に短縮されます。

### 基本的な使い方

```bash
# コマンドをwall-vault rtkで囲むと出力が自動フィルタリングされます
wall-vault rtk git status          # 変更されたファイルリストのみ表示
wall-vault rtk git diff HEAD~1     # 変更行 + 最小コンテキストのみ
wall-vault rtk git log -10         # ハッシュ + 1行メッセージずつ
wall-vault rtk go test ./...       # 失敗したテストのみ表示
wall-vault rtk ls -la              # 非対応コマンドは自動切り詰め
```

### 対応コマンドと節約効果

| コマンド | フィルター方式 | 節約率 |
|---------|-------------|--------|
| `git status` | 変更ファイルの要約のみ | ~87% |
| `git diff` | 変更行 + 3行コンテキスト | ~60-94% |
| `git log` | ハッシュ + 1行目メッセージ | ~90% |
| `git push/pull/fetch` | 進捗表示除去、要約のみ | ~80% |
| `go test` | 失敗のみ表示、成功はカウント | ~88-99% |
| `go build/vet` | エラーのみ表示 | ~90% |
| その他すべてのコマンド | 先頭50行 + 末尾50行、最大32KB | 可変 |

### 3段階フィルターパイプライン

1. **コマンド別構造フィルター** — git、goなどの出力形式を理解し、意味のある部分のみ抽出
2. **正規表現後処理** — ANSIカラーコード除去、空行圧縮、重複行集計
3. **パススルー + 切り詰め** — 非対応コマンドは先頭/末尾50行のみ保持

### Claude Code連携

Claude Codeの`PreToolUse`フックで、すべてのシェルコマンドを自動的にRTKを経由するように設定できます。

```bash
# フックインストール（Claude Code settings.jsonに自動追加）
wall-vault rtk hook install
```

または手動で`~/.claude/settings.json`に追加：

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

> :bulb: **終了コード保持**：RTKは元のコマンドの終了コードをそのまま返します。コマンドが失敗した場合（exit code != 0）、AIも正確に失敗を検出します。

> :bulb: **英語強制**：RTKは`LC_ALL=C`でコマンドを実行し、システム言語設定に関係なく常に英語の出力を生成します。フィルターが正確に動作するために必要です。

---

## 環境変数リファレンス

環境変数は、プログラムに設定値を渡す方法です。`export 変数名=値`の形式でターミナルに入力するか、自動起動サービスファイルに記述しておけば常時適用されます。

| 変数 | 説明 | 例 |
|------|------|---|
| `WV_LANG` | ダッシュボード言語 | `ko`, `en`, `ja` |
| `WV_THEME` | ダッシュボードテーマ | `light`, `dark`, `gold` |
| `WV_KEY_GOOGLE` | Google APIキー（カンマ区切りで複数） | `AIza...,AIza...` |
| `WV_KEY_OPENROUTER` | OpenRouter APIキー | `sk-or-v1-...` |
| `WV_VAULT_URL` | 分散モードでの金庫サーバーアドレス | `http://192.168.x.x:56243` |
| `WV_VAULT_TOKEN` | クライアント（ボット）認証トークン | `my-secret-token` |
| `WV_ADMIN_TOKEN` | 管理者トークン | `admin-token-here` |
| `WV_MASTER_PASS` | APIキー暗号化パスワード | `my-password` |
| `WV_AVATAR` | アバター画像ファイルパス（`~/.openclaw/`基準の相対パス） | `workspace/avatars/avatar.png` |
| `OLLAMA_URL` | Ollamaローカルサーバーアドレス | `http://192.168.x.x:11434` |

---

## トラブルシューティング

### プロキシが起動しない場合

ポートが既に別のプログラムに使用されている場合が多いです。

```bash
ss -tlnp | grep 56244   # ポート56244を使っているのは何か確認
wall-vault proxy --port 8080   # 別のポートで起動
```

### APIキーエラーが出る場合（429, 402, 401, 403, 582）

| エラーコード | 意味 | 対処法 |
|------------|------|--------|
| **429** | リクエスト過多（使用量超過） | しばらく待つか、他のキーを追加 |
| **402** | 支払い必要またはクレジット不足 | 該当サービスでクレジットをチャージ |
| **401 / 403** | キーが間違っているか権限なし | キー値を再確認して再登録 |
| **582** | ゲートウェイ過負荷（5分間クールダウン） | 5分後に自動解除 |

```bash
# 登録されたキーリストとステータスを確認
curl -H "Authorization: Bearer 管理者トークン" http://localhost:56243/admin/keys

# キー使用量カウンターをリセット
curl -X POST -H "Authorization: Bearer 管理者トークン" http://localhost:56243/admin/keys/reset
```

### エージェントが「未接続」と表示される場合

「未接続」とは、プロキシプロセスが金庫にハートビート（信号）を送っていない状態です。**設定が保存されていないという意味ではありません。** プロキシが金庫サーバーのアドレスとトークンを知った状態で実行されている必要があります。

```bash
# 金庫サーバーアドレス、トークン、クライアントIDを指定してプロキシを起動
WV_VAULT_URL=http://金庫サーバー:56243 \
WV_VAULT_TOKEN=クライアントトークン \
WV_VAULT_CLIENT_ID=クライアントID \
wall-vault proxy
```

接続に成功すると、約20秒以内にダッシュボードで:green_circle: 実行中に変わります。

### Ollamaに接続できない場合

Ollamaは自分のコンピューターでAIを直接実行するプログラムです。まずOllamaが起動しているか確認してください。

```bash
curl http://localhost:11434/api/tags   # モデルリストが表示されれば正常
export OLLAMA_URL=http://192.168.x.x:11434   # 別のコンピューターで実行中の場合
```

> :warning: Ollamaが応答しない場合は、まず`ollama serve`コマンドでOllamaを起動してください。

> :warning: **大型モデルは応答が遅いです**：`qwen3.5:35b`や`deepseek-r1`のような大きなモデルは、応答生成に数分かかることがあります。応答がないように見えても処理中の可能性があるので、お待ちください。

---

## 最近の変更点（v0.1.16 ~ v0.1.27）

### v0.1.27 (2026-04-09)
- **Ollamaフォールバックモデル名修正**：他のサービスからOllamaへのフォールバック時に、プロバイダー接頭辞付きモデル名（例：`google/gemini-3.1-pro-preview`）がOllamaにそのまま渡されていた問題を修正。環境変数/デフォルトモデルに自動置換されるようになりました。
- **クールダウン時間を大幅短縮**：429レート制限30分→5分、402支払い1時間→30分、401/403 24時間→6時間。すべてのキーが同時にクールダウンに入りプロキシが完全に麻痺する状況を防止。
- **全キークールダウン時の強制再試行**：すべてのキーがクールダウン状態の場合、最も早く解除されるキーを強制的に再試行して、リクエスト拒否を防止します。
- **サービスリスト表示修正**：`/status`レスポンスがvaultから同期された実際のサービスリストを表示するようになりました（anthropicなどの表示漏れを防止）。

### v0.1.25 (2026-04-08)
- **エージェントプロセス検出**：プロキシがローカルエージェント（NanoClaw/OpenClaw）の生存を検出し、ダッシュボードにオレンジの信号機で表示します。
- **ドラッグハンドル改善**：カード並べ替え時に信号機（●）エリアでのみ掴めるように変更。入力フィールドやボタンからの誤ドラッグを防止。

### v0.1.24 (2026-04-06)
- **RTKトークン節約サブコマンド**：`wall-vault rtk <command>`でシェルコマンドの出力を自動フィルタリングし、AIエージェントのトークン使用量を60-90%削減。git、goなどの主要コマンド用の専用フィルターを内蔵し、非対応コマンドも自動切り詰め。Claude Codeの`PreToolUse`フックで透過的に連携。

### v0.1.23 (2026-04-06)
- **Ollamaモデル変更修正**：金庫ダッシュボードでOllamaモデルを変更しても実際のプロキシに反映されなかった問題を修正。以前は環境変数（`OLLAMA_MODEL`）のみ使用していましたが、金庫設定を優先するようになりました。
- **ローカルサービス自動信号機**：Ollama、LM Studio、vLLMが到達可能な場合は自動有効化、切断時は自動無効化。クラウドサービスのキーベース自動トグルと同じ方式。

### v0.1.22 (2026-04-05)
- **空のcontentフィールド欠落修正**：thinkingモデル（gemini-3.1-pro、o1、claude thinkingなど）がmax_tokensの上限をreasoningに使い切り実際の応答を生成できない場合、プロキシが応答JSONの`content`/`text`フィールドを`omitempty`で欠落させ、OpenAI/Anthropic SDKクライアントが`Cannot read properties of undefined (reading 'trim')`エラーでクラッシュする問題を修正。公式APIスペック通り常にフィールドを含むように変更。

### v0.1.21 (2026-04-05)
- **Gemma 4モデルサポート**：Google Gemini API経由で`gemma-4-31b-it`、`gemma-4-26b-a4b-it`などのGemmaシリーズモデルが使用可能になりました。
- **LM Studio / vLLMサービス正式サポート**：以前はこれらのサービスがプロキシルーティングから漏れており、常にOllamaに代替されていました。OpenAI互換APIで正常にルーティングされるようになりました。
- **ダッシュボードサービス表示修正**：フォールバックが発生しても、ダッシュボードには常にユーザーが設定したサービスが表示されるようになりました。
- **ローカルサービスステータス表示**：ダッシュボード読み込み時に、ローカルサービス（Ollama、LM Studio、vLLMなど）の接続状態を●の色で表示。
- **ツールフィルター環境変数**：`WV_TOOL_FILTER=passthrough`環境変数でツールのパススルーモードを設定可能。

### v0.1.20 (2026-03-28)
- **包括的セキュリティ強化**：XSS防止（41か所）、定数時間トークン比較、CORS制限、リクエストサイズ制限、パストラバーサル防止、SSE認証、レートリミッター強化など12項目のセキュリティ改善。

### v0.1.19 (2026-03-27)
- **Claude Codeオンライン検出**：プロキシを経由しないClaude Codeもダッシュボードでオンラインとして表示されるようになりました。

### v0.1.18 (2026-03-26)
- **フォールバックサービス固着修正**：一時的なエラーでOllamaにフォールバックした後、元のサービスが復旧すると自動復帰。
- **オフライン検出改善**：15秒間隔の状態チェックでプロキシの停止検出が速くなりました。

### v0.1.17 (2026-03-25)
- **ドラッグ&ドロップカード並べ替え**：エージェントカードをドラッグして順序を変更可能。
- **インライン設定適用ボタン**：オフラインエージェントに[:zap: 設定適用]ボタンが表示されます。
- **cokacdir エージェントタイプ追加**。

### v0.1.16 (2026-03-25)
- **双方向モデル同期**：金庫ダッシュボードでCline・Claude Codeのモデルを変更すると自動反映されます。

---

*詳しいAPI情報は[API.md](API.md)をご覧ください。*
