<p align="center">
  <img src="docs/logo.png" alt="wall-vault" width="200">
</p>

# wall-vault

> **단일 Go 바이너리에 담긴 API 키 금고 + AI 프록시.**
> 키를 AES-GCM 으로 로컬 보관하고, 여러 제공자에 걸쳐 회전시키며, 한쪽이 실패하면 다음으로 페일오버하고, 실시간 대시보드를 함께 제공합니다.

[![License: GPL v3](https://img.shields.io/badge/License-GPLv3-blue.svg)](LICENSE)
[![Go](https://img.shields.io/badge/Go-1.25-00ADD8.svg)](go.mod)

[English](README.md) · 한국어 · [中文](README.zh.md) · [日本語](README.ja.md) · [Español](README.es.md) · [Français](README.fr.md) · [Deutsch](README.de.md) · [Português](README.pt.md) · [العربية](README.ar.md) · [हिन्दी](README.hi.md) · [Bahasa Indonesia](README.id.md) · [ภาษาไทย](README.th.md) · [Kiswahili](README.sw.md) · [Hausa](README.ha.md) · [नेपाली](README.ne.md) · [Монгол](README.mn.md) · [isiZulu](README.zu.md)

---

## 무엇인가

wall-vault 는 AI 에이전트(OpenClaw, Claude Code, Cursor, Continue, 직접 만든 스크립트)와 그 에이전트가 호출하는 클라우드·로컬 AI 제공자 사이에 자리잡습니다. 한 바이너리에 두 가지가 들어 있습니다.

- **금고(vault)** — API 키를 저장 시점에 암호화(마스터 비밀번호 기반 AES-GCM)해서 보관하고, 회전시키고, 키별 사용량·쿨다운을 추적하고, 변경 사항을 SSE 로 브로드캐스트하며, `:56243` 에 웹 대시보드를 띄웁니다.
- **프록시(proxy)** — `:56244` 에서 Gemini · Anthropic · OpenAI 호환 엔드포인트를 노출하고, 금고에서 키를 골라 설정한 상류로 디스패치하며, 한쪽이 실패하면 다음 제공자로 페일오버합니다.

네 가지 요청 형태(Gemini `:generateContent`, Anthropic `/v1/messages`, OpenAI `/v1/chat/completions`, Ollama 네이티브 `/api/chat`)와 다섯 부류의 상류를 지원합니다.

| 제공자 | 비고 |
|--------|------|
| Google Gemini | 네이티브 API, 프로젝트별 키 회전 |
| Anthropic | 네이티브 `/v1/messages` 패스스루 |
| OpenAI | 네이티브 `/v1/chat/completions` |
| OpenRouter | 340+ 모델, `:free` 변형 자동 폴백 |
| Ollama / LM Studio / vLLM / llama.cpp / Jan / KoboldCpp / TabbyAPI / mlx-server / LiteLLM | OpenAI 호환 로컬 백엔드, 플러그인 yaml 로 드롭인 |

OpenAI 호환 백엔드는 `~/.wall-vault/services/` 아래에 yaml 한 장만 두면 됩니다 — 코드 수정 없음.

## 어떤 사용자에게 맞는가

- AI 서비스를 서너 개 동시에 쓰면서 에이전트가 보는 URL 은 한 개로 통일하고 싶을 때
- 무료 티어 키가 쿨다운 들어가도 다음 키가 자연스럽게 이어 받아 세션이 안 끊겼으면 할 때
- 같은 LAN 의 여러 봇·IDE·스크립트가 같은 키 묶음을 공유하되 개별 머신에 자격 증명을 복사해 두지 않고 싶을 때
- 환경 변수 대신 대시보드에서 API 키를 편집하고 싶을 때
- 클라우드 한도가 떨어지면 로컬(Ollama / LM Studio) 로 자연스럽게 떨어지는 백업 경로가 필요할 때

## 빠른 시작

### 설치 (Linux / macOS)

```bash
curl -fsSL https://raw.githubusercontent.com/sookmook/wall-vault/main/install.sh | sh
```

또는 사전 빌드 바이너리를 직접 받기:

```bash
# Linux amd64
curl -L https://github.com/sookmook/wall-vault/releases/latest/download/wall-vault-linux-amd64 \
  -o wall-vault && chmod +x wall-vault

# macOS Apple Silicon
curl -L https://github.com/sookmook/wall-vault/releases/latest/download/wall-vault-darwin-arm64 \
  -o wall-vault && chmod +x wall-vault

# Linux arm64 (Raspberry Pi, ARM 서버)
curl -L https://github.com/sookmook/wall-vault/releases/latest/download/wall-vault-linux-arm64 \
  -o wall-vault && chmod +x wall-vault
```

### 설치 (Windows)

```powershell
Invoke-WebRequest -Uri `
  "https://github.com/sookmook/wall-vault/releases/latest/download/wall-vault-windows-amd64.exe" `
  -OutFile wall-vault.exe
```

### 첫 실행

```bash
wall-vault setup    # 대화형 마법사 — 포트, 서비스, 관리자 토큰, 마스터 비밀번호 결정
wall-vault start    # 금고와 프록시를 함께 띄움
```

브라우저에서 `http://localhost:56243` 을 엽니다(아래 TLS 켜기 절차 후엔 `https://...`). 대시보드가 `setup` 이 출력한 관리자 토큰을 묻고, 그 후엔 재기동 없이 API 키 추가, 클라이언트 등록, 모델 전환을 할 수 있습니다.

---

## TLS 설정 (권장)

`wall-vault setup` 의 기본값은 TLS 미적용 — 두 리스너가 평문 HTTP 로 답합니다. 이 README 의 예시 URL 은 `https://localhost:56244` 인데, 대부분의 에이전트(OpenClaw, Claude Code, Cursor)는 호스트가 바뀌어도 안 깨지는 단일 TLS 엔드포인트를 선호하기 때문입니다. 예시와 맞추려면 내장 내부 CA 로 TLS 를 한 번 켜 두면 됩니다.

```bash
# 1. 내부 CA 생성 (한 번만, ~/.wall-vault/ca.{crt,key} 에 저장)
wall-vault cert init

# 2. 이 머신용 호스트 인증서 발급
#    SAN 에 hostname, localhost, 127.0.0.1, 감지된 LAN IP 가 포함됩니다
wall-vault cert issue $(hostname)

# 3. 이 머신의 OS 신뢰 저장소에 CA 등록
wall-vault cert install-trust

# 4. 리스너를 TLS 모드로 전환
export WV_PROXY_TLS_ENABLED=1
export WV_PROXY_TLS_CERT="$HOME/.wall-vault/$(hostname).crt"
export WV_PROXY_TLS_KEY="$HOME/.wall-vault/$(hostname).key"
export WV_VAULT_TLS_ENABLED=1
export WV_VAULT_TLS_CERT="$HOME/.wall-vault/$(hostname).crt"
export WV_VAULT_TLS_KEY="$HOME/.wall-vault/$(hostname).key"
wall-vault start
```

LAN 의 다른 머신에서도 쓰려면 `~/.wall-vault/ca.crt` 를 옮기고 그쪽에서 `wall-vault cert install-trust --ca <경로>` 를 한 번 실행합니다. CA 가 모든 머신에 신뢰되면 네트워크 상의 어느 머신에서든 인증서 경고 없이 `https://<host>:56244` 로 프록시에 도달합니다.

평문 HTTP 그대로 쓰고 싶다면 설정을 그대로 두고 아래 클라이언트 예시의 `https://` 를 `http://` 로 바꾸기만 하면 됩니다. 둘 다 작동하며, 차이는 어느 포트가 TLS 핸드셰이크에 응답하느냐 뿐입니다.

**루프백 보조 리스너.** wall-vault CA 를 신뢰하기 어려운 동일 호스트 클라이언트(특히 OpenClaw 의 번들 Node 런타임이 `NODE_EXTRA_CA_CERTS` 를 spawn 시에 덮어쓰는 케이스)는 `127.0.0.1:56245` 의 루프백 전용 평문 HTTP 보조 리스너로 우회합니다. TLS 가 켜져 있을 때 자동 활성화됩니다.

---

## 클라이언트 연결

`https://<host>:56244` (TLS 미적용 시 `http://...`) 에 어떤 AI 클라이언트든 연결할 수 있습니다. 프록시는 네 가지 형태에 응답합니다.

| 형태 | 경로 | 예시 클라이언트 |
|------|------|-----------------|
| Gemini | `/google/v1beta/models/<model>:generateContent` | OpenClaw, Gemini CLI, Antigravity |
| Anthropic | `/v1/messages` | Claude Code, Anthropic SDK |
| OpenAI | `/v1/chat/completions` | Cursor, Continue, 직접 만든 스크립트, 대부분의 LLM 앱 |
| Ollama 네이티브 | `/api/chat` | Ollama 클라이언트 패스스루 |

### Claude Code

```bash
export ANTHROPIC_BASE_URL=https://localhost:56244
export ANTHROPIC_API_KEY=<금고-클라이언트-토큰>
claude
```

상류 Anthropic 크레딧이 떨어지면 디스패치는 이 클라이언트의 `fallback_services` 에 설정한 제공자로 떨어집니다. 비-Claude 모델 자동 치환을 명시적으로 옵트인하려면:

```yaml
proxy:
  anthropic_fallback_model: "claude-haiku-4-5-20251001"
```

(빈 기본값은 디스패치가 명확한 에러를 반환하게 합니다 — 잘못된 라우팅이 즉시 드러나도록.)

### Cursor / Continue

Cursor **Settings → AI → OpenAI API**:

```
Base URL:  https://localhost:56244
API Key:   <금고-클라이언트-토큰>
Model:     gemini-2.5-flash    # 또는 wall-vault 가 알고 있는 어떤 모델이든
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
      "apiKey": "<금고-클라이언트-토큰>"
    }
  ]
}
```

### OpenClaw

OpenClaw 는 wall-vault 가 본래 보조하기 위해 만들어진 TUI 에이전트 프레임워크입니다. 대시보드의 **Add Agent** 모달에서 에이전트 타입을 `openclaw` (또는 `nanoclaw`) 로 지정하면, wall-vault 가 `~/.openclaw/openclaw.json` 을 직접 작성합니다 — 제공자 URL, 금고 토큰, 모델 항목 모두.

```bash
VAULT_CLIENT_ID=my-bot \
VAULT_URL=http://<vault-host>:56243 \
VAULT_TOKEN=<클라이언트-토큰> \
wall-vault proxy
```

### curl / 스크립트

```bash
curl -X POST https://localhost:56244/v1/chat/completions \
  -H "Authorization: Bearer <금고-클라이언트-토큰>" \
  -H "Content-Type: application/json" \
  -d '{
    "model": "gemini-2.5-flash",
    "messages": [{"role": "user", "content": "ping"}]
  }'
```

---

## 설정

`wall-vault setup` 은 `./wall-vault.yaml` 또는 `~/.wall-vault/config.yaml` 에 저장합니다. 마법사가 묻지 않는 항목은 직접 편집하면 됩니다.

```yaml
mode: standalone     # standalone | distributed
lang: ko             # ko | en | zh | ja | es | fr | de | pt | ar | hi | id | th | sw | ha | ne | mn | zu
theme: light         # light | dark | cherry | ocean | gold | autumn | winter

proxy:
  port: 56244
  host: ""           # 기본: 127.0.0.1 standalone, 0.0.0.0 distributed
  client_id: my-bot
  vault_url: ""      # distributed: http://vault-host:56243
  vault_token: ""    # distributed: 클라이언트 토큰
  tool_filter: strip_all   # strip_all | whitelist | passthrough
  allowed_tools: []
  services: [google, openrouter, ollama]
  timeout: 300s
  tls:
    enabled: false
    cert_file: ""
    key_file: ""
  plain_port: 56245              # TLS 켜졌을 때만 활성, 루프백 전용 HTTP 보조
  ollama_keep_alive: "30m"       # "-1" 절대 unload 안 함, "0" 즉시 unload
  ollama_num_ctx: 8192
  oai_stream_forward: false      # 백엔드 SSE 패스스루 옵트인
  anthropic_fallback_model: ""   # anthropic 디스패치 비-Claude 자동 치환 옵트인

vault:
  port: 56243
  host: ""
  admin_token: ""
  admin_ip_whitelist: []
  master_password: ""            # AES-GCM 키 암호화 비밀번호
  data_dir: ~/.wall-vault/data
  services_dir: ~/.wall-vault/services
  tls:
    enabled: false
    cert_file: ""
    key_file: ""
  bootstrap_port: 56247          # ca.crt 만 노출하는 평문 HTTP 부트스트랩

doctor:
  interval: 5m
  auto_fix: true
  log_file: /tmp/wall-vault-doctor.log

hooks:
  on_model_change: ""            # shell 명령 (env: SERVICE, MODEL)
  on_key_exhausted: ""
  on_service_down: ""
  on_doctor_fix: ""
  openclaw_socket: ""
```

### 환경 변수

모든 YAML 필드에 환경 변수 오버라이드가 있습니다(env 가 파일보다 우선). 자주 쓰는 항목:

| 변수 | 설명 |
|------|------|
| `WV_LANG`, `WV_THEME` | 언어와 테마 |
| `WV_PROXY_PORT`, `WV_PROXY_HOST` | 프록시 리스너 |
| `WV_VAULT_PORT`, `WV_VAULT_HOST` | 금고 리스너 |
| `WV_VAULT_URL`, `WV_VAULT_TOKEN` | distributed 모드 엔드포인트 |
| `WV_ADMIN_TOKEN`, `WV_MASTER_PASS` | 금고 자격증명 |
| `WV_KEY_GOOGLE`, `WV_KEY_OPENROUTER`, `WV_KEY_ANTHROPIC`, `WV_KEY_OPENAI` | API 키 (여러 개는 콤마 구분) |
| `WV_PROXY_TLS_ENABLED`, `WV_PROXY_TLS_CERT`, `WV_PROXY_TLS_KEY` | 프록시 TLS |
| `WV_VAULT_TLS_ENABLED`, `WV_VAULT_TLS_CERT`, `WV_VAULT_TLS_KEY` | 금고 TLS |
| `WV_PROXY_PLAIN_PORT` | 루프백 HTTP 보조 (`0` 비활성) |
| `WV_VAULT_BOOTSTRAP_PORT` | CA 부트스트랩 리스너 (`0` 비활성) |
| `WV_OLLAMA_URL`, `WV_OLLAMA_KEEP_ALIVE`, `WV_OLLAMA_NUM_CTX` | Ollama 튜닝 |
| `WV_LMSTUDIO_URL`, `WV_VLLM_URL`, `WV_LLAMACPP_URL` | 로컬 백엔드 URL 오버라이드 |
| `WV_TOKEN_SENTINEL_FALLBACK` | 루프백 "proxy-managed" 센티넬 치환 |
| `WV_OAI_STREAM_FORWARD` | OpenAI 호환 백엔드 SSE 패스스루 |
| `WV_ANTHROPIC_FALLBACK_MODEL` | anthropic 디스패치 비-Claude 자동 치환 |

---

## 운용 모드

### Standalone (기본)

금고와 프록시가 같은 프로세스에서 동작합니다. 키와 에이전트가 한 호스트에 함께 있을 때 가장 단순. 기본은 루프백만 바인딩.

```bash
wall-vault start    # 둘 다 실행
```

### Distributed

한 호스트(**금고 호스트**) 가 모든 키를 보관하고, 다른 호스트의 여러 프록시가 각자 클라이언트별 토큰으로 인증합니다. 여러 머신이 같은 키 묶음이 필요한데 머신 간에 자격증명을 복사하지 않고 싶을 때 유용.

**금고 호스트:**

```bash
WV_VAULT_HOST=0.0.0.0 wall-vault vault
```

**각 프록시 호스트:**

```bash
WV_VAULT_URL=http://<vault-host>:56243 \
WV_VAULT_TOKEN=<해당-클라이언트-토큰> \
WV_PROXY_HOST=0.0.0.0 \
wall-vault proxy
```

대시보드의 **Add Client** 모달이 토큰을 발급하고 에이전트 타입을 등록합니다. 프록시는 SSE 로 자기 설정을 받아 재기동 없이 적용합니다.

---

## 플러그인 yaml (드롭인 백엔드)

OpenAI 호환 백엔드는 `~/.wall-vault/services/` 아래의 yaml 한 장으로 추가됩니다. wall-vault 가 시작 시 읽어 등록하고, 디스패치 + OAI 호환 감지 셋 + Gemini 스트림 브리지 모두 코드 수정 없이 인식합니다.

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
inline_no_think_for_qwen3: false   # 백엔드가 마커를 strip 하면 옵트인
```

허브 토폴로지(한 wall-vault 가 다른 wall-vault 를 앞단에서 받음) 는 `tls_internal_ca: true`, `auth.type: bearer`, `preserve_model_id: true` 로 지원됩니다.

---

## 소스에서 빌드

```bash
git clone https://github.com/sookmook/wall-vault
cd wall-vault
go build -o wall-vault .
./wall-vault setup
./wall-vault start
```

지원 플랫폼 전체 크로스 컴파일:

```bash
make build-all   # linux-amd64, linux-arm64, darwin-amd64, darwin-arm64, windows-amd64
```

버전은 `v{major}.{minor}.{patch}.{YYYYMMDD}.{HHmmss}` 형식. `BASE_VERSION` 이 Makefile 의 접두사를 결정합니다.

### 프로젝트 구조

```
wall-vault/
├── main.go                     # CLI 디스패치 (start/proxy/vault/setup/cert/doctor)
├── cmd/
│   ├── setup/                  # 대화형 setup 마법사
│   └── cert/                   # 내부 CA + 호스트별 TLS 인증서 발급
├── internal/
│   ├── config/                 # YAML + env 로더, 플러그인 로더
│   ├── proxy/                  # 요청 디스패치, 키 회전, 형식 변환
│   ├── vault/                  # AES-GCM 저장소, 대시보드, SSE 브로커
│   ├── doctor/                 # 헬스 프로브 + 자동 복구
│   ├── hooks/                  # shell 명령 이벤트 트리거
│   └── i18n/                   # 17개 언어 UI 문자열
├── configs/services/           # 번들 플러그인 yaml (lmstudio, vllm, ollama, …)
└── docs/                       # MANUAL, API 레퍼런스, 16개 locale 변형
```

---

## 문서

- [사용자 매뉴얼](docs/MANUAL.md) — 설치, 대시보드, 에이전트, 트러블슈팅
- [API 레퍼런스](docs/API.md) — 모든 엔드포인트의 요청/응답 형태
- [CHANGELOG](CHANGELOG.md)

---

## 기술 스택

- Go 1.25, 단일 정적 바이너리
- 서버 렌더 대시보드는 [templ](https://templ.guide), 부분 갱신은 [HTMX](https://htmx.org)
- 키 저장 시 암호화: PBKDF2 유도 키로 AES-GCM
- 금고와 프록시 간 실시간 설정 동기화: Server-Sent Events
- 자체 서명 내부 CA + 호스트별 인증서(공인 DNS / Let's Encrypt 불필요)

## 라이선스

GPL-3.0. [LICENSE](LICENSE) 참조.

## 기여

Pull request 환영합니다. [CONTRIBUTING.md](CONTRIBUTING.md) 참조. 큰 변경은 먼저 이슈를 열어 설계를 논의해 주세요.
