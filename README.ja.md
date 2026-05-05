# wall-vault

> **APIキーボールト + AIプロキシを単一のGoバイナリに。**
> AES-GCMでローカルにキーを保管し、複数プロバイダ間でローテーションし、片方が失敗したらフォールバックし、リアルタイムダッシュボードを同梱。

[![License: GPL v3](https://img.shields.io/badge/License-GPLv3-blue.svg)](LICENSE)
[![Go](https://img.shields.io/badge/Go-1.25-00ADD8.svg)](go.mod)

[English](README.md) · [한국어](README.ko.md) · [中文](README.zh.md) · **日本語** · [Español](README.es.md) · [Français](README.fr.md) · [Deutsch](README.de.md) · [Português](README.pt.md) · [العربية](README.ar.md) · [हिन्दी](README.hi.md) · [Bahasa Indonesia](README.id.md) · [ภาษาไทย](README.th.md) · [Kiswahili](README.sw.md) · [Hausa](README.ha.md) · [नेपाली](README.ne.md) · [Монгол](README.mn.md) · [isiZulu](README.zu.md)

---

## 概要

wall-vault はAIエージェント(OpenClaw、Claude Code、Cursor、Continue、自作スクリプトなど)と、それが対話するクラウドまたはローカルのAIプロバイダの間に位置します。1つのバイナリに2つの機能を備えます。

- **Vault** — APIキーを保存時に暗号化(マスターパスワードによるAES-GCM)し、ローテーションを行い、キーごとの使用状況とクールダウンを記録し、変更をSSEで配信し、`:56243` でWebダッシュボードを提供します。
- **Proxy** — Gemini、Anthropic、OpenAI互換のエンドポイントを `:56244` に公開し、ボールトからキーを選んで設定済みのアップストリームへ振り分け、片方が失敗したら次のプロバイダにフォールバックします。

4種類のリクエスト形式(Gemini `:generateContent`、Anthropic `/v1/messages`、OpenAI `/v1/chat/completions`、Ollama-native `/api/chat`)と、5カテゴリのアップストリームをサポートします。

| プロバイダ | 備考 |
|----------|-------|
| Google Gemini | ネイティブAPI、プロジェクト単位のキーローテーション |
| Anthropic | ネイティブ `/v1/messages` パススルー |
| OpenAI | ネイティブ `/v1/chat/completions` |
| OpenRouter | 340以上のモデル、`:free` バリアントへの自動フォールバック |
| Ollama / LM Studio / vLLM / llama.cpp / Jan / KoboldCpp / TabbyAPI / mlx-server / LiteLLM | OpenAI互換のローカルバックエンド、プラグインyamlで簡単導入 |

新しいOpenAI互換バックエンドの追加は `~/.wall-vault/services/` 配下のyamlファイル1つで完結し、コード変更は不要です。

## 使う理由

- 3〜4個のAIサービスを使い分けていて、エージェントから話しかけるURLを1つにまとめたい。
- クールダウン中の無料枠キーが、セッションを途切れさせずに次のキーに譲ってほしい。
- 同じLAN上の複数のボット/IDE/スクリプトに、認証情報をコピペせずに同じキーを使わせたい。
- 環境変数ではなくダッシュボードでAPIキーを編集したい。
- クラウドの上限を超えたときに、ローカルファースト(Ollama / LM Studio)の選択肢が欲しい。

## クイックスタート

### インストール (Linux / macOS)

```bash
curl -fsSL https://raw.githubusercontent.com/sookmook/wall-vault/main/install.sh | sh
```

ビルド済みバイナリを直接ダウンロードする場合。

```bash
# Linux amd64
curl -L https://github.com/sookmook/wall-vault/releases/latest/download/wall-vault-linux-amd64 \
  -o wall-vault && chmod +x wall-vault

# macOS Apple Silicon
curl -L https://github.com/sookmook/wall-vault/releases/latest/download/wall-vault-darwin-arm64 \
  -o wall-vault && chmod +x wall-vault

# Linux arm64 (Raspberry Pi、ARMサーバ)
curl -L https://github.com/sookmook/wall-vault/releases/latest/download/wall-vault-linux-arm64 \
  -o wall-vault && chmod +x wall-vault
```

### インストール (Windows)

```powershell
Invoke-WebRequest -Uri `
  "https://github.com/sookmook/wall-vault/releases/latest/download/wall-vault-windows-amd64.exe" `
  -OutFile wall-vault.exe
```

### 初回起動

```bash
wall-vault setup    # 対話ウィザード — ポート、サービス、admin token、master password を選択
wall-vault start    # vault と proxy を両方起動
```

ブラウザで `http://localhost:56243` (TLSが有効なら `https://...`、後述参照)を開きます。ダッシュボードは `setup` が出力したadmin tokenを要求します。そこからAPIキーの追加、クライアント登録、再起動なしのモデル切り替えが可能です。

---

## TLS (推奨)

デフォルトでは `wall-vault setup` はTLSなしで設定を書き出すため、両方のリスナーは平文HTTPで応答します。このREADMEのURL例で `https://localhost:56244` を使っているのは、ほとんどのエージェント(OpenClaw、Claude Code、Cursor)が、後でプロキシを別ホストへ移しても壊れないTLSフロントの単一エンドポイントを望むからです。これらの例に合わせるには、同梱の内部CAで一度だけTLSを有効化してください。

```bash
# 1. wall-vault 内部CAを作成 (一度だけ、~/.wall-vault/ca.{crt,key} に格納)
wall-vault cert init

# 2. このマシン用のホスト証明書を発行
#    SANには hostname、localhost、127.0.0.1、検出された任意のLAN IPが含まれる
wall-vault cert issue $(hostname)

# 3. ローカルOSのキーチェーンでCAを信頼
wall-vault cert install-trust

# 4. リスナーをTLSに切り替え
export WV_PROXY_TLS_ENABLED=1
export WV_PROXY_TLS_CERT="$HOME/.wall-vault/$(hostname).crt"
export WV_PROXY_TLS_KEY="$HOME/.wall-vault/$(hostname).key"
export WV_VAULT_TLS_ENABLED=1
export WV_VAULT_TLS_CERT="$HOME/.wall-vault/$(hostname).crt"
export WV_VAULT_TLS_KEY="$HOME/.wall-vault/$(hostname).key"
wall-vault start
```

LAN上の別マシンの場合は、`~/.wall-vault/ca.crt` をコピーしてそこで `wall-vault cert install-trust --ca <path>` を実行してください。CAがどこからでも信頼されるようになれば、ネットワーク上のどのマシンも証明書警告なしに `https://<host>:56244` でプロキシに到達できます。

平文HTTPのまま使いたければ、設定はそのままで、以降のクライアント例では `https://` を `http://` に置き換えてください。両方のスキームで動作します。違いはどちらのポートがTLSハンドシェイクに応答するかだけです。

**ループバックフォールバック。** wall-vault のCAを尊重できない同一ホストクライアント(特にspawn時に `NODE_EXTRA_CA_CERTS` を書き換えるOpenClawの同梱Nodeランタイム)は、`127.0.0.1:56245` のループバック専用平文HTTPコンパニオン経由でプロキシに到達します。TLSが有効になっていると、wall-vault が自動で有効化します。

---

## クライアント接続

任意のAIクライアントを `https://<host>:56244` (TLSがオフなら `http://...`)に向けます。プロキシは4種類の形式に応答します。

| 形式 | パス | クライアント例 |
|--------|------|-----------------|
| Gemini | `/google/v1beta/models/<model>:generateContent` | OpenClaw、Gemini CLI、Antigravity |
| Anthropic | `/v1/messages` | Claude Code、Anthropic SDK |
| OpenAI | `/v1/chat/completions` | Cursor、Continue、自作スクリプト、ほとんどのLLMアプリ |
| Ollama-native | `/api/chat` | パススルーするOllamaクライアント |

### Claude Code

```bash
export ANTHROPIC_BASE_URL=https://localhost:56244
export ANTHROPIC_API_KEY=<your-vault-client-token>
claude
```

アップストリームのAnthropicクレジットが切れると、振り分けはこのクライアントの `fallback_services` に設定したプロバイダへフォールバックします。Claude以外へのフォールバックを明示的にオプトインするには。

```yaml
proxy:
  anthropic_fallback_model: "claude-haiku-4-5-20251001"
```

(デフォルトの空値では振り分けがエラーを返すため、誤ルーティングが即座に表面化します。)

### Cursor / Continue

Cursor の **Settings → AI → OpenAI API**:

```
Base URL:  https://localhost:56244
API Key:   <your-vault-client-token>
Model:     gemini-2.5-flash    # または wall-vault が認識する任意のモデル
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

OpenClaw はTUIエージェントフレームワークで、wall-vault は元々これを支えるために作られました。ダッシュボードの **Add Agent** モーダルでエージェントタイプを `openclaw` (または `nanoclaw`)に設定すると、wall-vault は `~/.openclaw/openclaw.json` を直接書き込みます。プロバイダURL、ボールトトークン、モデルエントリも含まれます。

```bash
VAULT_CLIENT_ID=my-bot \
VAULT_URL=http://<vault-host>:56243 \
VAULT_TOKEN=<your-client-token> \
wall-vault proxy
```

### curl / スクリプト

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

## 設定

`wall-vault setup` は `./wall-vault.yaml` または `~/.wall-vault/config.yaml` のいずれかを書き出します。ウィザードが尋ねないフィールドは手で編集してください。

```yaml
mode: standalone     # standalone | distributed
lang: en             # en | ko | zh | ja | es | fr | de | pt | ar | hi | id | th | sw | ha | ne | mn | zu
theme: light         # light | dark | cherry | ocean | gold | autumn | winter

proxy:
  port: 56244
  host: ""           # デフォルト: standalone は 127.0.0.1、distributed は 0.0.0.0
  client_id: my-bot
  vault_url: ""      # distributed: http://vault-host:56243
  vault_token: ""    # distributed: クライアントトークン
  tool_filter: strip_all   # strip_all | whitelist | passthrough
  allowed_tools: []
  services: [google, openrouter, ollama]
  timeout: 300s
  tls:
    enabled: false
    cert_file: ""
    key_file: ""
  plain_port: 56245              # TLS有効時のループバック専用HTTPコンパニオン
  ollama_keep_alive: "30m"       # "-1" は永久にアンロードしない、"0" は即座にアンロード
  ollama_num_ctx: 8192
  oai_stream_forward: false      # 実バックエンドSSEパススルーをオプトイン
  anthropic_fallback_model: ""   # anthropic 振り分け時の非Claudeリライトをオプトイン

vault:
  port: 56243
  host: ""
  admin_token: ""
  admin_ip_whitelist: []
  master_password: ""            # AES-GCM キー暗号化パスワード
  data_dir: ~/.wall-vault/data
  services_dir: ~/.wall-vault/services
  tls:
    enabled: false
    cert_file: ""
    key_file: ""
  bootstrap_port: 56247          # ca.crt のみを提供する平文HTTPリスナー

doctor:
  interval: 5m
  auto_fix: true
  log_file: /tmp/wall-vault-doctor.log

hooks:
  on_model_change: ""            # シェルコマンド (env: SERVICE, MODEL)
  on_key_exhausted: ""
  on_service_down: ""
  on_doctor_fix: ""
  openclaw_socket: ""
```

### 環境変数

すべてのYAMLフィールドにはファイルより優先される環境変数オーバーライドが用意されています。よく使うもの。

| 変数 | 説明 |
|----------|-------------|
| `WV_LANG`, `WV_THEME` | 言語とテーマ |
| `WV_PROXY_PORT`, `WV_PROXY_HOST` | プロキシのリッスンアドレス |
| `WV_VAULT_PORT`, `WV_VAULT_HOST` | ボールトのリッスンアドレス |
| `WV_VAULT_URL`, `WV_VAULT_TOKEN` | 分散モードのエンドポイント |
| `WV_ADMIN_TOKEN`, `WV_MASTER_PASS` | ボールト認証情報 |
| `WV_KEY_GOOGLE`, `WV_KEY_OPENROUTER`, `WV_KEY_ANTHROPIC`, `WV_KEY_OPENAI` | APIキー (複数指定はカンマ区切り) |
| `WV_PROXY_TLS_ENABLED`, `WV_PROXY_TLS_CERT`, `WV_PROXY_TLS_KEY` | プロキシTLS |
| `WV_VAULT_TLS_ENABLED`, `WV_VAULT_TLS_CERT`, `WV_VAULT_TLS_KEY` | ボールトTLS |
| `WV_PROXY_PLAIN_PORT` | ループバックHTTPコンパニオン (`0` で無効) |
| `WV_VAULT_BOOTSTRAP_PORT` | CAブートストラップリスナー (`0` で無効) |
| `WV_OLLAMA_URL`, `WV_OLLAMA_KEEP_ALIVE`, `WV_OLLAMA_NUM_CTX` | Ollama チューニング |
| `WV_LMSTUDIO_URL`, `WV_VLLM_URL`, `WV_LLAMACPP_URL` | ローカルバックエンドオーバーライド |
| `WV_TOKEN_SENTINEL_FALLBACK` | ループバックの「proxy-managed」センチネル置換 |
| `WV_OAI_STREAM_FORWARD` | OpenAI互換の実バックエンドSSEパススルー |
| `WV_ANTHROPIC_FALLBACK_MODEL` | anthropic での非Claudeリライトをオプトイン |

---

## モード

### Standalone (デフォルト)

ボールトとプロキシが同一プロセス内で動作します。鍵もエージェントも同じホストでホストする場合に最適です。デフォルトではループバックのみリッスンします。

```bash
wall-vault start    # 両方を起動
```

### Distributed

ボールトは1つのホスト(**vault host**)で動作してすべての鍵を保持し、他ホスト上の複数プロキシがそれぞれクライアントごとのトークンで認証します。複数のマシンが同じ鍵を必要とするが、それらをコピーして回りたくない場合に有用です。

**Vault host:**

```bash
WV_VAULT_HOST=0.0.0.0 wall-vault vault
```

**各 proxy host:**

```bash
WV_VAULT_URL=http://<vault-host>:56243 \
WV_VAULT_TOKEN=<this-client-token> \
WV_PROXY_HOST=0.0.0.0 \
wall-vault proxy
```

ダッシュボードの **Add Client** モーダルがトークンを発行してエージェントタイプを登録すると、プロキシは再起動なしにSSE経由で設定を取得します。

---

## プラグイン yaml (ドロップインバックエンド)

任意のOpenAI互換バックエンドは `~/.wall-vault/services/` 配下のyamlとして追加できます。wall-vault は起動時に拾い上げ、振り分け可能なサービスとして登録し、コード変更なしに振り分けロジック、OAI互換検出セット、Geminiストリームブリッジのすべてから見えるようになります。

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
inline_no_think_for_qwen3: false   # マーカーを削るバックエンドならオプトイン
```

ハブトポロジ(wall-vault が別の wall-vault のフロントになる構成)は `tls_internal_ca: true`、`auth.type: bearer`、`preserve_model_id: true` でサポートされます。

---

## ソースからビルド

```bash
git clone https://github.com/sookmook/wall-vault
cd wall-vault
go build -o wall-vault .
./wall-vault setup
./wall-vault start
```

サポート対象一式へのクロスコンパイル。

```bash
make build-all   # linux-amd64, linux-arm64, darwin-amd64, darwin-arm64, windows-amd64
```

バージョンは `v{major}.{minor}.{patch}.{YYYYMMDD}.{HHmmss}` に従います。Makefile の `BASE_VERSION` がプレフィックスを設定します。

### プロジェクト構成

```
wall-vault/
├── main.go                     # CLI ディスパッチ (start/proxy/vault/setup/cert/doctor)
├── cmd/
│   ├── setup/                  # 対話的セットアップウィザード
│   └── cert/                   # 内部CA + ホスト単位 TLS 証明書発行
├── internal/
│   ├── config/                 # YAML + env ローダ、プラグインローダ
│   ├── proxy/                  # リクエスト振り分け、キーローテーション、形式変換
│   ├── vault/                  # AES-GCM ストア、ダッシュボード、SSE ブローカー
│   ├── doctor/                 # ヘルスプローブ + 自動修復
│   ├── hooks/                  # シェルコマンドのイベントトリガ
│   └── i18n/                   # 17言語のUI文字列
├── configs/services/           # 同梱プラグイン yaml (lmstudio, vllm, ollama, …)
└── docs/                       # MANUAL、APIリファレンス、16ロケールバリアント
```

---

## ドキュメント

- [ユーザマニュアル](docs/MANUAL.en.md) — インストール、ダッシュボード、エージェント、トラブルシュート
- [APIリファレンス](docs/API.en.md) — リクエスト/レスポンス形式付きの全エンドポイント
- [CHANGELOG](CHANGELOG.md)

---

## 技術スタック

- Go 1.25、単一の静的バイナリ
- サーバレンダリングダッシュボードに [templ](https://templ.guide)、部分更新に [HTMX](https://htmx.org)
- 保存時のキー暗号化に AES-GCM (PBKDF2 由来の鍵)
- ボールトとプロキシ間のライブ設定同期に Server-Sent Events
- 自己署名内部CA + ホスト単位の証明書 (公開DNSや Let's Encrypt は不要)

## ライセンス

GPL-3.0。[LICENSE](LICENSE) を参照してください。

## コントリビュート

プルリクエスト歓迎。[CONTRIBUTING.md](CONTRIBUTING.md) を参照してください。大きな変更の場合は事前に Issue で設計を相談してください。
