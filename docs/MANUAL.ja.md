# wall-vault ユーザーマニュアル

[English](MANUAL.md) · [한국어](MANUAL.ko.md) · [中文](MANUAL.zh.md) · **日本語** · [Español](MANUAL.es.md) · [Français](MANUAL.fr.md) · [Deutsch](MANUAL.de.md) · [Português](MANUAL.pt.md) · [العربية](MANUAL.ar.md) · [हिन्दी](MANUAL.hi.md) · [Bahasa Indonesia](MANUAL.id.md) · [ภาษาไทย](MANUAL.th.md) · [Kiswahili](MANUAL.sw.md) · [Hausa](MANUAL.ha.md) · [नेपाली](MANUAL.ne.md) · [Монгол](MANUAL.mn.md) · [isiZulu](MANUAL.zu.md)

このマニュアルでは wall-vault のインストール、設定、運用について説明します。概要については [README](../README.md) を参照してください。HTTP API の詳細については [API リファレンス](API.md) を参照してください。

## 目次

1. [wall-vault の機能](#wall-vault-の機能)
2. [インストール](#インストール)
3. [セットアップウィザードによる初回起動](#セットアップウィザードによる初回起動)
4. [TLS の有効化](#tls-の有効化)
5. [API キーの登録](#api-キーの登録)
6. [エージェントの接続](#エージェントの接続)
7. [ダッシュボード](#ダッシュボード)
8. [分散モード](#分散モード)
9. [自動起動](#自動起動)
10. [プラグイン yaml](#プラグイン-yaml)
11. [Doctor](#doctor)
12. [Hooks](#hooks)
13. [環境変数](#環境変数)
14. [トラブルシューティング](#トラブルシューティング)

---

## wall-vault の機能

wall-vault は、協調動作する 2 つのサービスをバンドルした単一の Go バイナリです。

- **vault** は API キーを保管時に暗号化(マスターパスワードによる AES-GCM)して保存し、キーごとの使用状況とクールダウンを追跡し、Server-Sent Events(SSE)で変更をブロードキャストし、人間のオペレーター向けに `:56243` で Web ダッシュボードを提供します。
- **proxy** は `:56244` で Gemini、Anthropic、OpenAI 互換、Ollama ネイティブのエンドポイントを公開します。proxy を指す AI クライアントは vault 内のキーを使用しますが、クライアント側からキーは見えません。1 つの上流が失敗すると、ディスパッチは順番に次のプロバイダーへフォールバックします。

これは次のような場合に便利です。

- 複数のプロバイダーのキーを持っており、エージェントが通信する URL を 1 つにまとめたい場合。
- 無料プランのキーがクールダウン中でも、セッションを中断せずに切り替えたい場合。
- 同じ LAN 上の複数のボット、IDE、スクリプトに認証情報をコピーせずに同じキーを使わせたい場合。
- キーの編集とモデルの切り替えに、環境変数ではなくダッシュボードを使いたい場合。
- クラウドの上限に達したときにローカルなフォールバック(Ollama、LM Studio、vLLM)が欲しい場合。

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

## インストール

### Linux / macOS ワンライナー

```bash
curl -fsSL https://raw.githubusercontent.com/sookmook/wall-vault/main/install.sh | sh
```

このスクリプトは OS とアーキテクチャを自動検出し、適切なバイナリを `~/.local/bin/wall-vault` にダウンロードして実行可能にします。`~/.local/bin` が `PATH` に含まれていない場合は追加してください。

```bash
export PATH="$HOME/.local/bin:$PATH"
```

### 手動ダウンロード

ビルド済みバイナリは、リリースごとに `https://github.com/sookmook/wall-vault/releases` で公開されています。

```bash
# Linux amd64
curl -L https://github.com/sookmook/wall-vault/releases/latest/download/wall-vault-linux-amd64 \
  -o wall-vault && chmod +x wall-vault

# Linux arm64 (Raspberry Pi、ARM サーバー)
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

### ソースからのビルド

Go 1.25 以降が必要です。

```bash
git clone https://github.com/sookmook/wall-vault
cd wall-vault
go build -o wall-vault .
```

`make build-all` はサポートされている 5 つのプラットフォームすべてにクロスコンパイルします。バイナリは `bin/` に生成されます。

---

## セットアップウィザードによる初回起動

```bash
wall-vault setup
```

ウィザードは順番に次の項目を尋ねます。

1. **言語** — 17 種類の UI ロケールから 1 つ選択します。`$LANG` から自動検出されますが、ウィザードはとにかく一覧を表示します。
2. **テーマ** — `light`(デフォルト)、`dark`、`cherry`、`ocean`、`gold`、`autumn`、`winter`。見た目のみ。
3. **モード** — `standalone`(単一ホスト、デフォルト)または `distributed`(vault は 1 つのホスト、proxy はその他のホスト)。
4. **ボット名** — 自由形式の `client_id` スラグ。vault はこれを使ってクライアントごとの設定(モデルオーバーライド、フォールバックチェーン)をスコープ化します。
5. **proxy ポート** — デフォルトは `56244`。
6. **vault ポート** — デフォルトは `56243`(standalone 時のみ)。
7. **サービス選択** — Google Gemini、OpenRouter、Anthropic、OpenAI、Ollama、LM Studio、vLLM のそれぞれに対する y/N。複数選択可。それぞれが env-var ヒントを末尾に書き込みます。
8. **ツールフィルタ** — `strip_all`(デフォルト。セキュリティのために受信ツール定義をすべてブロック)または `passthrough`(任意のツールを通過させる)。
9. **管理トークン** — 空白のままにすると自動生成されます。ダッシュボードへのログインにこのトークンが必要です。
10. **マスターパスワード** — 空白のままにすると暗号化なし(非推奨)、値を設定するとキーストアが保管時に AES-GCM で暗号化されます。
11. **保存パス** — デフォルトは現在のディレクトリの `wall-vault.yaml`。ローダーは `~/.wall-vault/config.yaml` も参照します。

保存後、ウィザードは `doctor.FixTrust` を実行し、ローカルにインストールされているエージェント(OpenClaw、Claude Code、Cline)の信頼ストアに wall-vault 内部 CA を自動的に追加します。該当するエージェントがインストールされていない場合、このステップは `SKIP` を表示し、何も書き込みません。

その後、バイナリを起動します。

```bash
wall-vault start
```

`start` は vault と proxy を 1 つのプロセスで実行します(standalone モード)。distributed モードの場合は、vault ホストで `wall-vault vault` を、各 proxy ホストで `wall-vault proxy` を実行します。

ブラウザで `http://localhost:56243` を開きます。ウィザードが出力した管理トークンでログインします。

---

## TLS の有効化

ウィザードのデフォルトでは、両方のリスナーは平文 HTTP のままです。ほとんどのエージェント(OpenClaw、Claude Code、Cursor)は単一の HTTPS エンドポイントの方が動作が良いため、ローカルマシンを越えるデプロイメントでは TLS を推奨します。

wall-vault は独自の内部 CA を同梱しているため、公開 DNS 名や Let's Encrypt は不要です。

```bash
# 1. 内部 CA を作成 — ~/.wall-vault/ca.{crt,key} に書き込まれる。
#    CA はデフォルトで 10 年有効。--ca-years でオーバーライド可能。
wall-vault cert init

# 2. ホスト証明書を発行。Subject Alternative Names には自動的に以下が含まれる:
#       hostname、"localhost"、"127.0.0.1"、検出された非ループバック LAN IP。
#    発行先ディレクトリは --dir、有効期間は --host-years でオーバーライド可能。
wall-vault cert issue $(hostname)

# 3. このマシンの OS キーチェーンで CA を信頼する。
#    Linux: update-ca-certificates 経由で /etc/ssl/certs/ に書き込む(sudo 必要)。
#    macOS: security add-trusted-cert 経由で System キーチェーンに追加(sudo 必要)。
#    Windows: certutil 経由で CurrentUser\Root にインポート(管理者権限不要)。
wall-vault cert install-trust

# 4. 両方のリスナーで TLS を有効化。
export WV_PROXY_TLS_ENABLED=1
export WV_PROXY_TLS_CERT="$HOME/.wall-vault/$(hostname).crt"
export WV_PROXY_TLS_KEY="$HOME/.wall-vault/$(hostname).key"
export WV_VAULT_TLS_ENABLED=1
export WV_VAULT_TLS_CERT="$HOME/.wall-vault/$(hostname).crt"
export WV_VAULT_TLS_KEY="$HOME/.wall-vault/$(hostname).key"

wall-vault start
```

他の LAN マシンに信頼を拡張するには、`~/.wall-vault/ca.crt` をコピーして、各マシンで `wall-vault cert install-trust --ca <path>` を実行します。vault は、新しいクライアントが HTTPS 通信のために CA を必要とするデッドロック状況のために、`:56247`(**ブートストラップポート**)で小さな平文 HTTP リスナー経由で `ca.crt` も公開します。

### ループバック HTTP コンパニオン

一部のエージェント、特に OpenClaw に同梱されている Node ランタイムは、プロセス起動時に `NODE_EXTRA_CA_CERTS` を書き換えるため、オペレーターが提供した CA ヒントが破棄されます。`cert install-trust` を実行した後でも、デーモン内部から wall-vault CA を尊重できません。wall-vault はこれを回避するため、TLS が有効な場合は `127.0.0.1:56245` に追加の **ループバック専用平文 HTTP リスナー** をバインドします。同一ホストのクライアントはこのポート経由で TLS なしに proxy に接続でき、LAN 上のクライアントは引き続き TLS リスナーを使用します。

不要であれば `WV_PROXY_PLAIN_PORT=0` で無効化できます。

### `wall-vault cert list`

`~/.wall-vault/` 配下のすべての証明書を、subject、有効期間、SAN とともに表示します。

```
$ wall-vault cert list
ca.crt          subject=wall-vault internal CA   not-after=2036-05-05
hostname.crt    subject=hostname                 not-after=2031-05-05   SAN=hostname,localhost,127.0.0.1,192.168.…
```

---

## API キーの登録

2 つの方法があります。ダッシュボード、または環境変数。

### ダッシュボード(推奨)

1. 管理トークンで `https://localhost:56243` にログインします。
2. キーカードの **+ API key** をクリックします。
3. サービスを選択(Google、OpenRouter、Anthropic、OpenAI など)。
4. キーを貼り付けて保存します。

サービスごとに複数のキーを設定できます。proxy はそれらをラウンドロビンし、キーごとのクールダウンに該当するものはスキップします。

### 環境変数(ワンショット bootstrap)

```bash
export WV_KEY_GOOGLE="AIzaSyA1...,AIzaSyB2...,AIzaSyC3..."   # カンマ区切り
export WV_KEY_OPENROUTER="sk-or-v1-…"
export WV_KEY_ANTHROPIC="sk-ant-…"
export WV_KEY_OPENAI="sk-…"
wall-vault start
```

このように指定したキーは、初回起動時に暗号化ストアに書き込まれます。それ以降の起動ではディスクから読み込まれるので、初回起動後は環境変数を解除して構いません。

### クールダウンとローテーション

呼び出しが成功するたびに、そのキーの `usage_count` が増え、`last_used` が更新されます。HTTP 429 / 402 / 403 の場合、proxy はそのキーを **クールダウン** に置きます(デフォルト: 429 は 60 分、402 は 24 時間、403 は 12 時間)。次のディスパッチでは、そのサービスの別のキーが選ばれます。あるサービスのキーがすべてクールダウン中の場合、proxy はそのサービスを完全にスキップして、フォールバックチェーンの次のプロバイダーを試します。

クールダウンは、ダッシュボードでキーごとにカウントダウン付きで表示されます。

---

## エージェントの接続

### OpenClaw

OpenClaw は最初のターゲットクライアントです。ダッシュボードの **+ Add agent** モーダルを使用します。

- **Agent type** を `openclaw` または `nanoclaw` に設定。
- **Work directory** を設定 — OpenClaw の場合は `~/.openclaw` が自動入力されます。
- **preferred service** と任意で **model override** を選択。
- **Apply** をクリック。wall-vault は `~/.openclaw/openclaw.json` に直接書き込みます(プロバイダー URL、vault トークン、モデルエントリ)。

ダッシュボードからモデルを変更すると、OpenClaw は SSE 経由で 1〜3 秒以内に変更を取り込みます — 再起動不要です。

### Claude Code

```bash
export ANTHROPIC_BASE_URL=https://localhost:56244
export ANTHROPIC_API_KEY=<your-vault-client-token>
claude
```

上流の Anthropic クレジットが尽きると、ディスパッチはこのクライアントの `fallback_services` にリストされたサービスへフォールバックします。デフォルトでは、anthropic ディスパッチに送られた非 Claude モデル ID はエラーを返すため、誤ルーティングが即座に表面化します。自動書き換えに参加するには次のように設定します。

```yaml
proxy:
  anthropic_fallback_model: "claude-haiku-4-5-20251001"
```

### Cursor

Cursor の **Settings → AI → OpenAI API**:

```
Base URL:  https://localhost:56244
API Key:   <your-vault-client-token>
Model:     gemini-2.5-flash    # または wall-vault が知っている任意のモデル
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

### カスタム HTTP

```bash
curl -X POST https://localhost:56244/v1/chat/completions \
  -H "Authorization: Bearer <your-vault-client-token>" \
  -H "Content-Type: application/json" \
  -d '{
    "model": "gemini-2.5-flash",
    "messages": [{"role": "user", "content": "hello"}]
  }'
```

`proxy.oai_stream_forward: true` が設定されている場合、同じエンドポイントはストリーミング(`"stream": true`)も受け付けます。

---

## ダッシュボード

`https://localhost:56243`。ホームグリッドには 5 つのカードがあります。

- **Keys** — サービスごとにグループ化されたすべての API キー。追加、編集、削除。使用状況とクールダウンを表示。
- **Services** — Google / OpenRouter / Anthropic / OpenAI / Ollama / LM Studio / vLLM / llama.cpp、および `~/.wall-vault/services/` 内の任意のプラグイン yaml。サービスごとに `default_model`、`allowed_models`、ベース URL、reasoning トグルを設定。
- **Clients (agents)** — 登録済みのすべてのクライアント(OpenClaw ボット、Claude Code セッション、Cursor インスタンスなど)。優先サービス、モデルオーバーライド、フォールバックチェーンを割り当て。
- **Proxies** — この vault に対して認証されたすべての proxy。ライブステータス(オンライン/オフライン)、最終確認日時、現在のモデル。
- **Settings** — 管理トークン、マスターパスワードのローテーション、テーマ、言語。

各カードには編集スライドオーバー(右側)があります。外側をクリックするか `Esc` で閉じます。変更は SSE 経由で接続中のすべての proxy に数秒以内にプッシュされます。

**フッター** には SSE インジケータ(緑 = 接続中、オレンジ = 再接続中、グレー = 切断)とライブビルドバージョンが表示されます。

---

## 分散モード

同じキーを必要とする複数のマシンがある場合、1 つのホストで vault を、それ以外の各ホストで proxy を実行します。

### vault ホスト

```bash
WV_VAULT_HOST=0.0.0.0 \
WV_ADMIN_TOKEN=<admin> \
WV_MASTER_PASS=<master> \
wall-vault vault
```

ダッシュボードは `https://<vault-host>:56243` でアクセス可能になります。**Clients** カードで、リモート proxy ごとにエージェントを追加します。それぞれに固有の `vault_token` が発行されます。

### proxy ホスト

```bash
WV_VAULT_URL=http://<vault-host>:56243 \
WV_VAULT_TOKEN=<that-client-token> \
WV_PROXY_HOST=0.0.0.0 \
wall-vault proxy
```

proxy は vault に対して認証し、SSE ストリームを開き、受信した設定(優先サービス、モデルオーバーライド、フォールバックチェーン)を適用します。それ以降の vault の編集は再起動なしで数秒以内に反映されます。

LAN にまたがるインストールでは、vault ホストで TLS を有効化し(`WV_VAULT_TLS_ENABLED=1` + cert/key の env vars)、proxy の HTTPS 呼び出しが信頼されるように、各 proxy ホストで同じ `wall-vault cert install-trust` ステップを実行します。

---

## 自動起動

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
loginctl enable-linger $USER       # ログアウト後もユニットが動作し続けるように
```

同じホストの vault には、並行して `wall-vault-vault.service` を書きます。standalone モードでは、`wall-vault start` を呼び出す 1 つのユニットで十分です。

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

`nssm` を使って `wall-vault.exe start` を Windows サービスとしてラップするか、ユーザーログイン時に実行する `schtasks` エントリを使用します。

---

## プラグイン yaml

`~/.wall-vault/services/` 配下に yaml をドロップするだけで、コードを変更せずに任意の OpenAI 互換バックエンドを追加できます。wall-vault は起動時にこれを読み込み、ディスパッチ、OAI 互換検出セット、Gemini ストリームブリッジ用にサービスを登録します。

```yaml
# ~/.wall-vault/services/llamacpp.yaml
id: llamacpp                 # 一意のサービス ID
name: llama.cpp              # 人間向けラベル
enabled: true                # 無効化されたプラグインはロード時にスキップ

default_url: http://localhost:8080   # オペレーターのオーバーライド; env が優先 (WV_LLAMACPP_URL)
endpoints:
  generate: /v1/chat/completions
  list_models: /v1/models

auth:
  type: none                 # none | bearer | query_param | header
  param: ""                  # query_param の場合: パラメータ名 (例: "key")

request_format: openai       # openai | gemini | ollama | raw

model_fetch:
  enabled: true              # ダッシュボードがモデルを自動検出することを許可
  dynamic: true              # ダッシュボードが開かれるたびに再フェッチ
  auto_detect_url: true      # 宣言されていなくても /v1/models を試す

concurrency:
  max: 1                     # このバックエンドへの最大同時リクエスト数
  queue_size: 10
  wait_notify: true          # TUI エージェントに "queued" ヒントを表示

error_codes:
  503:
    cooldown: 5m
    message: "llama.cpp not responding"

# reasoning が無効なときに qwen3 系のインライン /no_think ディレクティブをオプトイン。
# バックエンドのチャットテンプレートがマーカーを除去する場合 (LM Studio の
# jinja、Ollama の /v1 レイヤー) は true に設定する。それ以外のバックエンドは
# 通常リテラルテキストをそのまま返すため、yaml ごとにオプトインのままにする。
inline_no_think_for_qwen3: false

# Hub トポロジー — 別の wall-vault を指す。このプラグインがリモートの
# wall-vault のフロントとして動作する場合に必要 (受信側の wall-vault が
# publisher プレフィックスを認識して正しくルーティングできるように)。また、
# proxy.vault_token のベアラートークンを Authorization として送信する。
preserve_model_id: false
tls_internal_ca: false       # ~/.wall-vault/ca.crt をクライアント信頼プールに追加
```

`configs/services/` にバンドルされているセット(lmstudio、vllm、llamacpp、tgwui、localai、jan、koboldcpp、tabbyapi、mlx-server、litellm-proxy、ollama、google、openrouter)はデフォルトで無効になっています。使用したいものを `~/.wall-vault/services/` にコピーし、`enabled: true` に設定して再起動してください。

---

## Doctor

`wall-vault doctor` は、インストール全体に対して 1 回限りのヘルスチェックを実行します。

```
✓ vault listener  (https://localhost:56243)
✓ proxy listener  (https://localhost:56244)
✓ master password set
⚠ Google: 2 keys, all on cooldown
✓ Anthropic: 1 key healthy
✗ Ollama: not reachable at http://localhost:11434
```

各行は次のいずれかです。

- `✓` — 正常
- `⚠` — 機能はしているが劣化(1 つのキーがクールダウン、クォータ低下など)
- `✗` — 故障
- `SKIP` — 未設定 / このホストでは該当なし

セカンドデーモンモードは `doctor.interval` ごと(デフォルト 5 分)に同じプローブを実行し、結果を `doctor.log_file`(デフォルト `/tmp/wall-vault-doctor.log`)に書き込みます。`doctor.auto_fix` が true の場合、よくあるドリフト(古い OpenClaw 設定、不足している TLS 信頼、再起動可能なサービス)も修復しようとします。

ダッシュボードの **Doctor** カードまたは `wall-vault doctor` から 1 回限りで起動できます。

---

## Hooks

主要なイベントでシェルコマンドを実行します。

```yaml
hooks:
  on_model_change:   "logger 'wall-vault: $SERVICE/$MODEL'"
  on_key_exhausted:  "notify-send 'wall-vault' '$SERVICE keys all on cooldown'"
  on_service_down:   "/usr/local/bin/page-oncall.sh $SERVICE '$ERROR'"
  on_doctor_fix:     "echo \"$AGENT: $LEVEL $MSG\" >> ~/wall-vault.audit.log"
  openclaw_socket:   ""    # 設定されている場合、OpenClaw TUI はこの Unix ソケット経由でイベントを受信
```

各 hook はイベント固有の環境変数(`SERVICE`、`MODEL`、`ERROR`、`AGENT`、`LEVEL`、`MSG`)を受け取ります。hook は 5 秒のタイムアウトで非同期に実行されます — proxy が遅い hook でブロックされることはありません。

---

## 環境変数

| 変数 | YAML フィールド |
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
| `WV_VAULT_TLS_ENABLED` | `vault.tls.enabled` |
| `WV_VAULT_TLS_CERT` | `vault.tls.cert_file` |
| `WV_VAULT_TLS_KEY` | `vault.tls.key_file` |
| `WV_VAULT_BOOTSTRAP_PORT` | `vault.bootstrap_port` |
| `WV_PROXY_PLAIN_PORT` | `proxy.plain_port` |
| `WV_KEY_GOOGLE` | ワンショットインポート: カンマ区切りの Google キー |
| `WV_KEY_OPENROUTER` | ワンショットインポート: OpenRouter キー |
| `WV_KEY_ANTHROPIC` | ワンショットインポート: Anthropic キー |
| `WV_KEY_OPENAI` | ワンショットインポート: OpenAI キー |
| `WV_OLLAMA_URL` | ホストごとの Ollama URL オーバーライド |
| `WV_OLLAMA_KEEP_ALIVE` | `proxy.ollama_keep_alive` |
| `WV_OLLAMA_NUM_CTX` | `proxy.ollama_num_ctx` |
| `WV_LMSTUDIO_URL`, `WV_VLLM_URL`, `WV_LLAMACPP_URL` | バックエンドごとの URL オーバーライド |
| `WV_TOKEN_SENTINEL_FALLBACK` | `proxy.token_sentinel_fallback` |
| `WV_OAI_STREAM_FORWARD` | `proxy.oai_stream_forward` |
| `WV_ANTHROPIC_FALLBACK_MODEL` | `proxy.anthropic_fallback_model` |
| `WV_ECONOWORLD_MAX_TOKENS` | `proxy.econoworld_max_tokens` |
| `WV_ECONOWORLD_STREAM` | `proxy.econoworld_stream` |
| `WV_ECONOWORLD_REQUEST_TIMEOUT` | `proxy.econoworld_request_timeout` |

設定されている環境変数は、YAML ファイルより優先されます。

---

## トラブルシューティング

### `:56244` で `connection refused`

proxy が動作していないか、別のホストにバインドされています。確認方法:

```bash
ss -lnp | grep 56244
systemctl --user status wall-vault-proxy   # Linux
launchctl list | grep wall-vault           # macOS
```

別のポートで動作している場合、設定で `proxy.port` がオーバーライドされています。`~/.wall-vault/config.yaml` を確認してください。

### `x509: certificate signed by unknown authority`

クライアントが wall-vault 内部 CA を信頼していません。クライアントマシンで `wall-vault cert install-trust` を実行してください。OS 信頼ストアを無視するランタイムを持つエージェント(例: ハードコードされた `NODE_EXTRA_CA_CERTS` を持つ Node)では、`127.0.0.1:56245`(同一ホストのみ)のループバック HTTP コンパニオンを使うか、`WV_PROXY_TLS_ENABLED=0` を設定して平文 HTTP にフォールバックします。

### `token not registered with vault`

クライアントの `Authorization: Bearer <token>` が、登録済みのクライアントと一致しません。ダッシュボードの **Clients** でトークンを確認してください。`proxy-managed`、`dummy`、`""` などのトークンリテラルを古い設定からコピーした場合は、本物のクライアントトークンに置き換えてください。

### `Anthropic dispatch needs a Claude model id`

v0.2.63 時点のデフォルト動作: anthropic ディスパッチに送られた非 Claude モデル ID はエラーを返します。ルーティングを修正(anthropic に `gemini-2.5-flash` を送らない)するか、`proxy.anthropic_fallback_model` で自動書き換えにオプトインしてください。

### `unknown service: <id>`

ディスパッチが、どのプラグイン yaml も主張しないサービス ID を見ました。確認方法:

```bash
ls ~/.wall-vault/services/        # プラグイン yaml が存在するか?
cat ~/.wall-vault/services/<id>.yaml | grep enabled
```

yaml は存在するが `enabled: false` の場合は、これを反転します。完全に欠落している場合は、ソースツリーの `configs/services/` からコピーします。

### reasoning モデルで応答が空

`qwen3.6`、`deepseek-r1`、GPT-`o1` ファミリーは、`reasoning_content` のみを発行し、`content` を空のままにすることがあります。v0.2.63 時点で wall-vault は自動的に reasoning テキストにフォールバックします — それでも応答が空のままの場合、バックエンドはどちらのフィールドも返していません。上流のログを確認してください。

特に qwen3 を使用する LM Studio の場合、プラグイン yaml で `inline_no_think_for_qwen3: true` を設定して、reasoning がインラインで無効になるようにします。組み込みの lmstudio.yaml と ollama.yaml は既にこれを実施しています。

### キーを追加したばかりなのに、ダッシュボードに「すべてのキーがクールダウン中」と表示される

新しいキーは正常ですが、ディスパッチパスが古いキーのクールダウンにまだ留まっている可能性があります。新しいリクエストを試してください — proxy は呼び出しごとにラウンドロビンを行うため、次に正常なキーが選ばれます。

### マスターパスワードで vault が解錠できない

パスワードが間違っています。リカバリ方法はありません — wall-vault は意図的にバックドアを同梱していません。マスターパスワードを本当に失ってしまった場合、`~/.wall-vault/data/vault.json` を削除し、新しいパスワードで再起動してキーを再追加するしかありません。

### OpenRouter の無料プラン上限に達した

`proxy.services` に `openrouter` を含め、少なくとも 1 つの OpenRouter キーを追加してください。proxy は、有料パスが 402 / 429 を返した場合、有料モデルから `:free` バリアントへ自動フォールバックします。

### `journalctl --user -u wall-vault-proxy` が空

systemd の `--user` ログは、それを実行しているユーザーのジャーナルに送られます。ユニットを `root` または `sudo` で起動した場合、ジャーナルはシステムインスタンスにあるので、`--user` なしで `journalctl -u wall-vault-proxy` を試してください。

---

## さらに

- HTTP API リファレンス — [API.md](API.md) を参照
- ソース — `https://github.com/sookmook/wall-vault`
- バグレポート / 機能リクエスト — GitHub Issues
- リリース履歴 — [CHANGELOG.md](../CHANGELOG.md)
