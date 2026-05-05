# wall-vault 사용자 매뉴얼

[English](MANUAL.md) · 한국어 · [中文](MANUAL.zh.md) · [日本語](MANUAL.ja.md) · [Español](MANUAL.es.md) · [Français](MANUAL.fr.md) · [Deutsch](MANUAL.de.md) · [Português](MANUAL.pt.md) · [العربية](MANUAL.ar.md) · [हिन्दी](MANUAL.hi.md) · [Bahasa Indonesia](MANUAL.id.md) · [ภาษาไทย](MANUAL.th.md) · [Kiswahili](MANUAL.sw.md) · [Hausa](MANUAL.ha.md) · [नेपाली](MANUAL.ne.md) · [Монгол](MANUAL.mn.md) · [isiZulu](MANUAL.zu.md)

이 매뉴얼은 wall-vault 의 설치, 설정, 운영을 다룹니다. 한눈에 보는 개요는 [README](../README.ko.md), HTTP API 의 자세한 정의는 [API 레퍼런스](API.md) 참조.

## 목차

1. [wall-vault 가 하는 일](#wall-vault-가-하는-일)
2. [설치](#설치)
3. [setup 마법사로 첫 실행](#setup-마법사로-첫-실행)
4. [TLS 켜기](#tls-켜기)
5. [API 키 등록](#api-키-등록)
6. [에이전트 연결](#에이전트-연결)
7. [대시보드](#대시보드)
8. [Distributed 모드](#distributed-모드)
9. [자동 시작](#자동-시작)
10. [플러그인 yaml](#플러그인-yaml)
11. [Doctor](#doctor)
12. [Hooks](#hooks)
13. [환경 변수](#환경-변수)
14. [트러블슈팅](#트러블슈팅)

---

## wall-vault 가 하는 일

wall-vault 는 두 서비스를 한 Go 바이너리에 담은 도구입니다.

- **금고(vault)** — API 키를 저장 시점에 암호화(마스터 비밀번호 기반 AES-GCM)해서 보관하고, 키별 사용량·쿨다운을 추적하고, 변경 사항을 SSE(Server-Sent Events) 로 브로드캐스트하며, 운영자용 웹 대시보드를 `:56243` 에 제공합니다.
- **프록시(proxy)** — `:56244` 에서 Gemini · Anthropic · OpenAI 호환 · Ollama 네이티브 엔드포인트를 노출합니다. 어떤 AI 클라이언트든 프록시를 가리키면 금고 안의 키를 사용하지만 클라이언트는 키를 보지 못합니다. 한쪽 상류가 실패하면 디스패치는 fallback 체인의 다음 제공자로 넘어갑니다.

다음 상황에 유용합니다.

- 여러 제공자의 키를 가지고 있고 에이전트가 보는 URL 은 한 개로 통일하고 싶을 때
- 무료 티어 키가 쿨다운 들어가도 다음 키가 자연스럽게 이어 받아 세션이 안 끊겼으면 할 때
- 같은 LAN 의 여러 봇·IDE·스크립트가 같은 키 묶음을 공유하되 머신마다 자격증명을 복사하지 않고 싶을 때
- 환경 변수 대신 대시보드에서 키를 편집하고 모델을 바꾸고 싶을 때
- 클라우드 한도가 떨어지면 로컬(Ollama / LM Studio / vLLM) 로 자연스럽게 떨어지는 백업 경로가 필요할 때

```
   AI 클라이언트 (OpenClaw, Claude Code, Cursor, …)
            │
            ▼
   wall-vault 프록시  :56244
            │  (키 선택 → 디스패치 → 실패 시 다음 제공자로 fallback)
            ├──► Google Gemini
            ├──► Anthropic
            ├──► OpenAI
            ├──► OpenRouter (340+ 모델, :free 자동 fallback)
            └──► 로컬 OAI 호환 백엔드 (Ollama / LM Studio / vLLM / …)

   금고 (AES-GCM 키 저장소 + 대시보드)  :56243
            ▲
            │  변경 시 SSE 브로드캐스트
   여러 호스트의 프록시들이 한 금고를 공유할 수 있습니다.
```

---

## 설치

### Linux / macOS 원라인 설치

```bash
curl -fsSL https://raw.githubusercontent.com/sookmook/wall-vault/main/install.sh | sh
```

스크립트가 OS 와 아키텍처를 자동 감지해 알맞은 바이너리를 `~/.local/bin/wall-vault` 로 받고 실행 권한을 부여합니다. `~/.local/bin` 이 `PATH` 에 없으면:

```bash
export PATH="$HOME/.local/bin:$PATH"
```

### 수동 다운로드

릴리스마다 사전 빌드 바이너리를 `https://github.com/sookmook/wall-vault/releases` 에 게시합니다.

```bash
# Linux amd64
curl -L https://github.com/sookmook/wall-vault/releases/latest/download/wall-vault-linux-amd64 \
  -o wall-vault && chmod +x wall-vault

# Linux arm64 (Raspberry Pi, ARM 서버)
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

### 소스 빌드

Go 1.25 이상 필요.

```bash
git clone https://github.com/sookmook/wall-vault
cd wall-vault
go build -o wall-vault .
```

`make build-all` 로 5개 플랫폼을 한 번에 크로스 컴파일. 결과물은 `bin/`.

---

## setup 마법사로 첫 실행

```bash
wall-vault setup
```

마법사가 묻는 항목 (순서대로):

1. **언어** — 17개 UI locale 중 하나. `$LANG` 으로 자동 감지하고, 그래도 목록을 보여 줍니다.
2. **테마** — `light` (기본), `dark`, `cherry`, `ocean`, `gold`, `autumn`, `winter`. 외관만 영향.
3. **운용 모드** — `standalone` (단일 호스트, 기본) 또는 `distributed` (금고는 한 호스트, 프록시는 다른 호스트들).
4. **봇 이름** — 자유 형식 `client_id` 슬러그. 금고가 클라이언트별 설정(모델 오버라이드, fallback 체인) 을 이 이름으로 분리합니다.
5. **프록시 포트** — 기본 `56244`.
6. **금고 포트** — 기본 `56243` (standalone 만).
7. **서비스 선택** — Google Gemini, OpenRouter, Anthropic, OpenAI, Ollama, LM Studio, vLLM 각각 y/N. 여러 개 선택해도 됩니다. 끝에서 각 서비스의 환경변수 안내문이 출력됩니다.
8. **도구 필터** — `strip_all` (기본 — 보안상 모든 inbound tool 정의 차단) 또는 `passthrough`.
9. **관리자 토큰** — 빈 입력이면 자동 생성. 대시보드 로그인 시 필요.
10. **마스터 비밀번호** — 빈 입력이면 암호화 없음 (권장 X). 값을 넣으면 키 저장소가 AES-GCM 으로 암호화됩니다.
11. **저장 경로** — 현재 디렉터리의 `wall-vault.yaml` 기본. 로더는 `~/.wall-vault/config.yaml` 도 살핍니다.

저장 후 마법사가 `doctor.FixTrust` 를 실행해 로컬 설치된 에이전트(OpenClaw, Claude Code, Cline) 가 wall-vault 내부 CA 를 자동 신뢰하도록 처리합니다. 해당 에이전트가 없으면 `SKIP` 표시 후 아무것도 쓰지 않습니다.

이후 실행:

```bash
wall-vault start
```

`start` 는 standalone 모드에서 금고와 프록시를 한 프로세스에서 띄웁니다. distributed 모드는 금고 호스트에서 `wall-vault vault`, 프록시 호스트들에서 `wall-vault proxy`.

브라우저로 `http://localhost:56243` 을 엽니다. 마법사가 출력한 관리자 토큰으로 로그인.

---

## TLS 켜기

마법사 기본값은 두 리스너 모두 평문 HTTP. 대부분의 에이전트(OpenClaw, Claude Code, Cursor) 는 단일 HTTPS 엔드포인트를 선호하므로, 단일 머신을 넘어가는 배포에는 TLS 가 권장됩니다.

wall-vault 는 자체 내부 CA 를 함께 제공하므로 공인 DNS 나 Let's Encrypt 가 필요 없습니다.

```bash
# 1. 내부 CA 생성 — ~/.wall-vault/ca.{crt,key} 에 저장.
#    기본 유효기간 10년. --ca-years 로 변경.
wall-vault cert init

# 2. 호스트 인증서 발급. SAN 에 자동 포함:
#       hostname, "localhost", "127.0.0.1", 감지된 비-루프백 LAN IP.
#    --dir 로 저장 폴더, --host-years 로 유효기간 변경.
wall-vault cert issue $(hostname)

# 3. 이 머신의 OS 신뢰 저장소에 CA 등록.
#    Linux: update-ca-certificates 로 /etc/ssl/certs/ 에 (sudo 필요).
#    macOS: security add-trusted-cert 로 시스템 키체인에 (sudo 필요).
#    Windows: certutil 로 CurrentUser\Root 에 (admin 불필요).
wall-vault cert install-trust

# 4. 두 리스너 TLS 활성화.
export WV_PROXY_TLS_ENABLED=1
export WV_PROXY_TLS_CERT="$HOME/.wall-vault/$(hostname).crt"
export WV_PROXY_TLS_KEY="$HOME/.wall-vault/$(hostname).key"
export WV_VAULT_TLS_ENABLED=1
export WV_VAULT_TLS_CERT="$HOME/.wall-vault/$(hostname).crt"
export WV_VAULT_TLS_KEY="$HOME/.wall-vault/$(hostname).key"

wall-vault start
```

LAN 의 다른 머신으로 신뢰를 넓히려면 `~/.wall-vault/ca.crt` 를 옮겨 두고 그쪽에서 `wall-vault cert install-trust --ca <경로>` 를 한 번 실행. 금고는 또한 `:56247` (**bootstrap port**) 의 작은 평문 HTTP 리스너로 `ca.crt` 를 노출합니다 — 새 클라이언트가 HTTPS 로 도달하기 위해 CA 가 필요한 catch-22 상황을 풀기 위해서입니다.

### 루프백 HTTP 보조 리스너

일부 에이전트(특히 OpenClaw 의 번들 Node 런타임)는 프로세스 spawn 시 `NODE_EXTRA_CA_CERTS` 를 덮어써서 운영자가 지정한 CA 힌트를 떨어뜨립니다. `cert install-trust` 후에도 데몬 내부에서는 wall-vault CA 를 신뢰하지 못합니다. wall-vault 는 TLS 가 켜져 있을 때 `127.0.0.1:56245` 에 **루프백 전용 평문 HTTP 리스너** 를 추가로 띄워 이 문제를 우회합니다. 동일 호스트 클라이언트는 그 포트로 TLS 없이 도달하고, LAN 클라이언트는 계속 TLS 리스너 사용.

필요 없으면 `WV_PROXY_PLAIN_PORT=0`.

### `wall-vault cert list`

`~/.wall-vault/` 의 모든 인증서를 subject, 유효 기간, SAN 과 함께 나열.

```
$ wall-vault cert list
ca.crt          subject=wall-vault internal CA   not-after=2036-05-05
hostname.crt    subject=hostname                 not-after=2031-05-05   SAN=hostname,localhost,127.0.0.1,192.168.…
```

---

## API 키 등록

두 가지 경로: 대시보드 또는 환경 변수.

### 대시보드 (권장)

1. `https://localhost:56243` 에 관리자 토큰으로 로그인.
2. 키 카드의 **+ API 키**.
3. 서비스 선택 (Google, OpenRouter, Anthropic, OpenAI …).
4. 키 붙여넣기. 저장.

서비스당 여러 키 OK. 프록시가 round-robin 으로 회전하며 쿨다운 걸린 키는 자동 skip.

### 환경 변수 (일회성 부트스트랩)

```bash
export WV_KEY_GOOGLE="AIzaSyA1...,AIzaSyB2...,AIzaSyC3..."   # 콤마 구분
export WV_KEY_OPENROUTER="sk-or-v1-…"
export WV_KEY_ANTHROPIC="sk-ant-…"
export WV_KEY_OPENAI="sk-…"
wall-vault start
```

이렇게 들어온 키는 첫 실행 시 암호화 저장소에 기록됩니다. 이후 시작은 디스크에서 읽으니 환경 변수는 unset 해도 됩니다.

### 쿨다운과 회전

성공한 호출마다 키의 `usage_count` 가 올라가고 `last_used` 가 갱신됩니다. HTTP 429 / 402 / 403 시 해당 키에 **쿨다운** 이 걸립니다 (기본 — 429: 60분, 402: 24시간, 403: 12시간). 다음 디스패치는 다른 키 선택. 한 서비스의 모든 키가 쿨다운이면 그 서비스를 fast-skip 하고 fallback 체인의 다음 제공자로 넘어갑니다.

쿨다운은 키별 카운트다운으로 대시보드에 표시됩니다.

---

## 에이전트 연결

### OpenClaw

OpenClaw 는 wall-vault 가 본래 보조하기 위해 만들어진 클라이언트. 대시보드의 **+ 에이전트 추가** 모달에서:

- **에이전트 타입** = `openclaw` 또는 `nanoclaw`.
- **작업 디렉터리** — OpenClaw 는 `~/.openclaw` 자동 채움.
- **선호 서비스** + 옵션으로 **모델 오버라이드** 선택.
- **적용**. wall-vault 가 `~/.openclaw/openclaw.json` 을 직접 작성 (제공자 URL, 금고 토큰, 모델 항목).

대시보드에서 모델을 바꾸면 OpenClaw 가 SSE 로 1–3 초 안에 받아옵니다 — 재기동 불필요.

### Claude Code

```bash
export ANTHROPIC_BASE_URL=https://localhost:56244
export ANTHROPIC_API_KEY=<금고-클라이언트-토큰>
claude
```

상류 Anthropic 크레딧이 떨어지면 디스패치가 이 클라이언트의 `fallback_services` 에 따라 fallback. v0.2.63 부터는 anthropic 디스패치에 비-Claude 모델 ID 가 들어오면 명확한 에러 — 잘못된 라우팅이 즉시 드러나도록. 자동 치환을 옵트인하려면:

```yaml
proxy:
  anthropic_fallback_model: "claude-haiku-4-5-20251001"
```

### Cursor

Cursor **Settings → AI → OpenAI API**:

```
Base URL:  https://localhost:56244
API Key:   <금고-클라이언트-토큰>
Model:     gemini-2.5-flash    # 또는 wall-vault 가 알고 있는 어떤 모델
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
      "apiKey": "<금고-클라이언트-토큰>"
    }
  ]
}
```

### 직접 HTTP 호출

```bash
curl -X POST https://localhost:56244/v1/chat/completions \
  -H "Authorization: Bearer <금고-클라이언트-토큰>" \
  -H "Content-Type: application/json" \
  -d '{
    "model": "gemini-2.5-flash",
    "messages": [{"role": "user", "content": "hello"}]
  }'
```

`proxy.oai_stream_forward: true` 가 켜져 있으면 같은 엔드포인트가 `"stream": true` 를 받아 SSE 로 응답합니다.

---

## 대시보드

`https://localhost:56243`. 홈 그리드에 다섯 카드.

- **Keys** — 서비스별 API 키 목록. 추가/편집/삭제, 사용량·쿨다운 표시.
- **Services** — Google / OpenRouter / Anthropic / OpenAI / Ollama / LM Studio / vLLM / llama.cpp 와 `~/.wall-vault/services/` 안의 모든 플러그인 yaml. 서비스별 `default_model`, `allowed_models`, base URL, 추론 토글 설정.
- **Clients (Agents)** — 등록된 모든 클라이언트(OpenClaw 봇, Claude Code 세션, Cursor 인스턴스, …). 선호 서비스, 모델 오버라이드, fallback 체인 지정.
- **Proxies** — 이 금고에 인증된 모든 프록시. 실시간 상태(online/offline), 마지막 접속, 현재 모델.
- **Settings** — 관리자 토큰, 마스터 비밀번호 회전, 테마, 언어.

각 카드는 우측 슬라이드오버에서 편집. 바깥 클릭 또는 `Esc` 로 닫힘. 변경은 SSE 로 모든 연결된 프록시에 수초 안에 푸시.

**푸터** 에 SSE 연결 표시(녹색=연결, 주황=재연결, 회색=끊김) 와 실시간 빌드 버전.

---

## Distributed 모드

여러 머신이 같은 키 묶음을 공유해야 할 때, 한 호스트가 금고를 띄우고 다른 호스트들이 프록시를 띄웁니다.

### 금고 호스트

```bash
WV_VAULT_HOST=0.0.0.0 \
WV_ADMIN_TOKEN=<관리자> \
WV_MASTER_PASS=<마스터> \
wall-vault vault
```

대시보드는 이제 `https://<금고-호스트>:56243`. 원격 프록시마다 **Clients** 카드에서 에이전트를 추가 — 각각 고유 `vault_token` 이 발급됩니다.

### 프록시 호스트들

```bash
WV_VAULT_URL=http://<금고-호스트>:56243 \
WV_VAULT_TOKEN=<해당-클라이언트-토큰> \
WV_PROXY_HOST=0.0.0.0 \
wall-vault proxy
```

프록시는 금고에 인증하고 SSE 스트림을 열어 받은 설정(선호 서비스, 모델 오버라이드, fallback 체인) 을 적용합니다. 이후 금고 편집은 수초 안에 반영 — 재기동 없음.

LAN 을 가로지르는 설치라면 금고 호스트에서 TLS 활성화(`WV_VAULT_TLS_ENABLED=1` + 인증서 env 변수) 후 각 프록시 호스트에서 `wall-vault cert install-trust` 를 실행해 프록시의 HTTPS 호출이 신뢰되도록 합니다.

---

## 자동 시작

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
loginctl enable-linger $USER       # 로그아웃 후에도 유닛 유지
```

같은 호스트의 금고는 `wall-vault-vault.service` 를 평행으로 작성. standalone 모드라면 `wall-vault start` 를 호출하는 단일 유닛이면 충분.

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

`nssm` 으로 `wall-vault.exe start` 를 Windows 서비스로 감싸거나, `schtasks` 로 사용자 로그인 시 실행되도록.

---

## 플러그인 yaml

OpenAI 호환 백엔드는 `~/.wall-vault/services/` 아래 yaml 한 장으로 추가 — 코드 수정 없음. wall-vault 가 시작 시 읽어 디스패치 + OAI 호환 감지 셋 + Gemini 스트림 브리지 모두 등록.

```yaml
# ~/.wall-vault/services/llamacpp.yaml
id: llamacpp                 # 고유 service id
name: llama.cpp              # 사람이 읽는 라벨
enabled: true                # false 면 로딩 시 skip

default_url: http://localhost:8080   # 운영자 오버라이드, env 가 더 우선 (WV_LLAMACPP_URL)
endpoints:
  generate: /v1/chat/completions
  list_models: /v1/models

auth:
  type: none                 # none | bearer | query_param | header
  param: ""                  # query_param 시 파라미터 이름 (예: "key")

request_format: openai       # openai | gemini | ollama | raw

model_fetch:
  enabled: true              # 대시보드가 모델 자동 감지
  dynamic: true              # 대시보드 열 때마다 재조회
  auto_detect_url: true      # 명시 안 돼도 /v1/models 시도

concurrency:
  max: 1                     # 백엔드별 동시 요청 수
  queue_size: 10
  wait_notify: true          # TUI 에이전트에 "큐 대기" 힌트

error_codes:
  503:
    cooldown: 5m
    message: "llama.cpp 미응답"

# 추론 비활성 시 qwen3 family 의 inline /no_think 옵트인.
# 백엔드의 chat 템플릿이 이 마커를 strip 하면 (LM Studio 의 jinja,
# Ollama 의 /v1 layer) 켜세요. 다른 백엔드는 보통 리터럴을 응답에
# 그대로 echo 해서 기본은 false.
inline_no_think_for_qwen3: false

# 허브 토폴로지 — 다른 wall-vault 를 가리킴. 이 플러그인이 원격
# wall-vault 를 앞단에서 받을 때 (수신측이 publisher prefix 로
# 라우팅하도록) 그리고 proxy.vault_token 이 Authorization 으로
# 보내지도록 필요.
preserve_model_id: false
tls_internal_ca: false       # 클라이언트 trust pool 에 ~/.wall-vault/ca.crt 추가
```

번들된 셋(`configs/services/` 의 lmstudio, vllm, llamacpp, tgwui, localai, jan, koboldcpp, tabbyapi, mlx-server, litellm-proxy, ollama, google, openrouter) 은 기본 비활성. 원하는 것을 `~/.wall-vault/services/` 로 복사하고 `enabled: true`, 재기동.

---

## Doctor

`wall-vault doctor` 는 설치 전체에 대해 일회성 헬스 프로브 실행:

```
✓ vault listener  (https://localhost:56243)
✓ proxy listener  (https://localhost:56244)
✓ master password set
⚠ Google: 2 keys, all on cooldown
✓ Anthropic: 1 key healthy
✗ Ollama: not reachable at http://localhost:11434
```

각 라인:

- `✓` — 정상
- `⚠` — 동작은 하지만 저하 (키 한 개 쿨다운, 쿼터 임박 등)
- `✗` — 망가짐
- `SKIP` — 미설정 / 이 호스트에 해당 없음

데몬 모드도 있어 `doctor.interval` (기본 5분) 마다 같은 프로브를 돌리고 `doctor.log_file` (기본 `/tmp/wall-vault-doctor.log`) 에 기록. `doctor.auto_fix: true` 면 흔한 drift (stale OpenClaw 설정, 빠진 TLS 신뢰, 재시작 가능한 서비스) 도 자동 복구 시도.

대시보드의 **Doctor** 카드 또는 `wall-vault doctor` 로 일회성 실행.

---

## Hooks

주요 이벤트에 shell 명령 실행:

```yaml
hooks:
  on_model_change:   "logger 'wall-vault: $SERVICE/$MODEL'"
  on_key_exhausted:  "notify-send 'wall-vault' '$SERVICE keys all on cooldown'"
  on_service_down:   "/usr/local/bin/page-oncall.sh $SERVICE '$ERROR'"
  on_doctor_fix:     "echo \"$AGENT: $LEVEL $MSG\" >> ~/wall-vault.audit.log"
  openclaw_socket:   ""    # 지정 시 OpenClaw TUI 가 이 Unix 소켓으로 이벤트 수신
```

각 hook 은 이벤트별 환경변수 (`SERVICE`, `MODEL`, `ERROR`, `AGENT`, `LEVEL`, `MSG`) 를 받습니다. async 실행 + 5초 timeout — 느린 hook 이 프록시를 막지 않습니다.

---

## 환경 변수

| 변수 | YAML 필드 |
|------|-----------|
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
| `WV_KEY_GOOGLE` | 일회성 임포트: 콤마 구분 Google 키 |
| `WV_KEY_OPENROUTER` | 일회성 임포트: OpenRouter 키 |
| `WV_KEY_ANTHROPIC` | 일회성 임포트: Anthropic 키 |
| `WV_KEY_OPENAI` | 일회성 임포트: OpenAI 키 |
| `WV_OLLAMA_URL` | 호스트별 Ollama URL 오버라이드 |
| `WV_OLLAMA_KEEP_ALIVE` | `proxy.ollama_keep_alive` |
| `WV_OLLAMA_NUM_CTX` | `proxy.ollama_num_ctx` |
| `WV_LMSTUDIO_URL`, `WV_VLLM_URL`, `WV_LLAMACPP_URL` | 백엔드별 URL 오버라이드 |
| `WV_TOKEN_SENTINEL_FALLBACK` | `proxy.token_sentinel_fallback` |
| `WV_OAI_STREAM_FORWARD` | `proxy.oai_stream_forward` |
| `WV_ANTHROPIC_FALLBACK_MODEL` | `proxy.anthropic_fallback_model` |
| `WV_ECONOWORLD_MAX_TOKENS` | `proxy.econoworld_max_tokens` |
| `WV_ECONOWORLD_STREAM` | `proxy.econoworld_stream` |
| `WV_ECONOWORLD_REQUEST_TIMEOUT` | `proxy.econoworld_request_timeout` |

env 변수는 set 되면 YAML 파일을 항상 이깁니다.

---

## 트러블슈팅

### `:56244` 에 `connection refused`

프록시가 안 떠 있거나 다른 호스트에 바인딩됨. 확인:

```bash
ss -lnp | grep 56244
systemctl --user status wall-vault-proxy   # Linux
launchctl list | grep wall-vault           # macOS
```

다른 포트면 설정의 `proxy.port` 가 오버라이드된 것. `~/.wall-vault/config.yaml` 점검.

### `x509: certificate signed by unknown authority`

클라이언트가 wall-vault 내부 CA 를 신뢰 안 함. 클라이언트 머신에서 `wall-vault cert install-trust`. 런타임이 OS 신뢰 저장소를 무시하는 에이전트(예: Node 의 hardcoded `NODE_EXTRA_CA_CERTS`) 는 `127.0.0.1:56245` 의 루프백 HTTP 보조 리스너 (동일 호스트) 사용 또는 `WV_PROXY_TLS_ENABLED=0` 으로 평문 HTTP fallback.

### `token not registered with vault`

클라이언트의 `Authorization: Bearer <token>` 이 등록된 클라이언트와 매치 안 함. 대시보드 **Clients** 에서 토큰 확인. stale config 에서 `proxy-managed`, `dummy`, `""` 같은 리터럴을 복사했다면 실제 클라이언트 토큰으로 교체.

### `Anthropic 디스패치는 Claude 모델 ID 가 필요`

v0.2.63 기본 동작 — anthropic 디스패치에 비-Claude 모델 ID 가 들어오면 에러. 라우팅을 고치거나(예: `gemini-2.5-flash` 를 anthropic 으로 보내지 말기) `proxy.anthropic_fallback_model` 로 자동 치환 옵트인.

### `unknown service: <id>`

디스패치가 어떤 플러그인 yaml 도 주장하지 않은 service id 를 만남. 점검:

```bash
ls ~/.wall-vault/services/        # 플러그인 yaml 존재?
cat ~/.wall-vault/services/<id>.yaml | grep enabled
```

yaml 은 있는데 `enabled: false` 면 켜기. 아예 없으면 소스 트리의 `configs/services/` 에서 복사.

### 추론 모델에서 빈 응답

`qwen3.6`, `deepseek-r1`, GPT `o1` family 가 가끔 `reasoning_content` 만 emit 하고 `content` 는 비웁니다. v0.2.63 부터 wall-vault 가 자동으로 추론 텍스트로 fallback — 그래도 빈 응답이 보이면 백엔드가 둘 다 안 보낸 것. 상류 로그 확인.

LM Studio + qwen3 조합은 플러그인 yaml 의 `inline_no_think_for_qwen3: true` 로 inline 비활성화. 번들된 lmstudio.yaml / ollama.yaml 은 이미 켜져 있습니다.

### "all keys on cooldown" 인데 방금 키 추가했음

새 키는 정상이지만 디스패치 path 가 아직 옛 키의 쿨다운에 잡혀 있을 수 있음. 다시 시도 — 프록시는 호출별 round-robin 이라 정상 키가 다음에 선택됩니다.

### 마스터 비밀번호로 금고 안 풀림

비밀번호 틀림. 복구 경로 없음 — wall-vault 는 의도적으로 backdoor 를 두지 않습니다. 정말 비밀번호를 잃었다면 `~/.wall-vault/data/vault.json` 삭제 → 새 비밀번호로 재시작 → 키 다시 추가.

### OpenRouter 무료 티어 한도 도달

`proxy.services` 에 `openrouter` 포함 + OpenRouter 키 1개 이상 추가. 유료 모델이 402/429 이면 프록시가 자동으로 `:free` 변형으로 fallback.

### `journalctl --user -u wall-vault-proxy` 가 비어 있음

systemd `--user` 로그는 유닛을 실행한 사용자의 journal. `root` 로 또는 `sudo` 로 시작했다면 system instance 의 journal 에 있음 — `journalctl -u wall-vault-proxy` (`--user` 없이).

---

## 더보기

- HTTP API 레퍼런스 — [API.md](API.md)
- 소스 — `https://github.com/sookmook/wall-vault`
- 버그 / 기능 요청 — GitHub Issues
- 릴리스 이력 — [CHANGELOG.md](../CHANGELOG.md)
