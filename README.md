<p align="center">
  <img src="docs/logo.png" alt="wall-vault" width="200">
</p>

<h1 align="center">🔐 wall-vault</h1>

<p align="center"><i>AI 프록시 + 키 금고 통합 시스템 · AI Proxy + Key Vault Unified System</i></p>

<p align="center">
  <a href="LICENSE"><img src="https://img.shields.io/badge/License-GPL%20v3-blue.svg" alt="License: GPL v3"></a>
  <img src="https://img.shields.io/badge/Go-1.22+-00ADD8.svg" alt="Go Version">
  <a href="https://github.com/sookmook/wall-vault/actions/workflows/ci.yml"><img src="https://github.com/sookmook/wall-vault/actions/workflows/ci.yml/badge.svg" alt="CI"></a>
  <img src="https://img.shields.io/badge/languages-10-brightgreen.svg" alt="Languages">
  <img src="https://img.shields.io/badge/platform-Linux%20%7C%20macOS%20%7C%20Windows-lightgrey.svg" alt="Platform">
</p>

---

## 언어 · Language

| [🇰🇷 한국어](#-탄생-배경) | [🇺🇸 English](#-origin-story) | [🇨🇳 中文](#-项目简介) | [🇯🇵 日本語](#-はじめに) | [🇪🇸 Español](#-introducción) | [🇫🇷 Français](#-présentation) | [🇩🇪 Deutsch](#-über-das-projekt) |

---

## 💀 탄생 배경: 봇들이 죽던 날

새벽이었다.

알림 하나가 떴다. 평소와 달랐다.

로그인해 보니 — API 키 전부 무효. vault.json 공백. 봇들은 아무 말도 없었다. 할 수가 없었다. **기억이 통째로 지워져 있었으니까.**

> *모토코는 내 작업 스타일을 꿰뚫고 있었다. 미니는 매일 아침 브리핑을 준비했다. 라즈는 라즈베리파이 위에서 묵묵히 모든 걸 처리했다. 2주 넘게 공들여 키운 AI 비서들이었다.*
>
> *해커 한 명이 내부망에 들어와서 그걸 전부 날려버렸다.*
>
> *잘 키운 반려동물이 하룻밤 사이에 사라진 것 같은 기분이었다.*

기억을 복원하는 데 일주일이 걸렸다. 완전하지도 않았다.

이건 두 번 다시 겪으면 안 됐다.

그래서 만들었다. **키를 잠그는 금고. 봇들을 지키는 벽. 다시는 해커 한 명 때문에 모든 게 끝나지 않도록.**

---

## ⚔️ 그래서, 이게 뭐냐면

한 줄 요약: **"AI 봇들이 절대 죽지 않게 만드는 보디가드."**

```
해커가 키를 털어도  → 금고가 막는다
키 한도가 차도      → 다음 키로 알아서 넘긴다
서비스가 다운돼도   → Gemini → OpenAI → Ollama 순서로 폴백
봇이 100대여도      → 설정 하나 바꾸면 1-3초 내 전원에 반영
```

더 풀어쓰면:

- 🔐 **키 금고(Vault)**: AES-GCM 암호화. 라운드 로빈 자동 순환. 할당량·오류·쿨다운 알아서 관리.
- 🔀 **AI 프록시(Proxy)**: OpenClaw·Claude Code·VS Code·내 스크립트 — 어디서 오든 Gemini / OpenAI / Ollama로 중계. 하나 죽으면 다음 걸로.
- ⚡ **SSE 실시간 동기화**: 금고에서 뭔가 바꾸면 연결된 모든 봇에 즉각 반영. 재시작 불필요.
- 🛡️ **보안 필터**: function calling 완전 차단. 외부 스킬이 내 AI를 멋대로 조종하는 걸 막는다.
- 🦞 **OpenClaw 전용 연동**: Unix 소켓으로 TUI에 실시간 이벤트 전달. openclaw.json 자동 갱신.

Go 바이너리 단 하나. 봇 한 대부터 분산 다중 봇까지 전부 커버.

---

## 목차

- [기능](#기능)
- [빠른 시작](#빠른-시작)
- [다국어](#다국어--languages)
- [사용법](#사용법)
- [아키텍처](#아키텍처)
- [설정](#설정)
- [지원 서비스](#지원-서비스)
- [API 엔드포인트](#api-엔드포인트)
- [모드](#모드)
- [자동 시작 설정](#자동-시작-설정)
- [서비스 플러그인](#서비스-플러그인)
- [OpenClaw 연동](#openclaw-연동)
- [빌드](#빌드)
- [프로젝트 구조](#프로젝트-구조)
- [라이선스](#라이선스)

---

## 기능

| 기능 | 설명 |
|------|------|
| **AI 프록시** | Google Gemini / OpenAI / Anthropic / OpenRouter / GitHub Copilot / Ollama / LMStudio / vLLM |
| **키 금고** | API 키 관리, 사용량 모니터링, 라운드 로빈 자동 순환 |
| **AES-GCM 암호화** | 마스터 비밀번호로 API 키 암호화 저장 |
| **SSE 실시간 동기화** | 금고 ↔ 프록시 1–3초 내 자동 반영 |
| **도구 보안 필터** | function calling 차단 (`strip_all` / `whitelist` / `passthrough`) |
| **폴백 체인** | 서비스 실패 시 자동 전환, 최종 폴백은 Ollama |
| **모델 레지스트리** | 전체 모델 ID·이름 검색 (OpenRouter 340개+) |
| **로컬 AI 지원** | Ollama / LM Studio / vLLM 자동 감지 + 수동 URL 설정 |
| **서비스 관리** | UI에서 서비스 추가·수정·삭제, 커스텀 서비스 지원 |
| **에이전트 관리** | 에이전트별 서비스·모델·IP 화이트리스트·작업 디렉토리 설정 |
| **에이전트 상태** | 4단계 표시 🟢실행중 / 🟡지연 / 🔴오프라인 / ⚫미연결 (heartbeat 미수신 설명 포함) |
| **타입별 설정 복사** | 🦞 openclaw / 🟠 claude-code / ⌨ cursor / 💻 vscode 각각 맞춤 스니펫 복사 |
| **주치의(Doctor)** | 헬스체크, 자동복구, systemd/launchd/NSSM 등록 |
| **[다국어](#다국어--languages)** | 세계 10대 언어 지원 |
| **테마** | 라이트 ☀️ / 다크 🌑 / 골드 ✨ / 벚꽃 🌸 / 오션 🌊 |
| **크로스 플랫폼** | Linux / macOS / Windows / WSL |
| **[OpenClaw 연동](#openclaw-연동)** | Unix 소켓으로 TUI 실시간 알림, 에이전트 자동 설정 |

---

## 빠른 시작

### Linux / macOS

```bash
# Linux (amd64)
curl -L https://github.com/sookmook/wall-vault/releases/latest/download/wall-vault-linux-amd64 \
  -o wall-vault && chmod +x wall-vault

# macOS Apple Silicon
curl -L https://github.com/sookmook/wall-vault/releases/latest/download/wall-vault-darwin-arm64 \
  -o wall-vault && chmod +x wall-vault

# 대화형 설치 마법사 (처음 시작)
./wall-vault setup

# 실행
./wall-vault start
```

### Windows

```powershell
Invoke-WebRequest -Uri `
  "https://github.com/sookmook/wall-vault/releases/latest/download/wall-vault-windows-amd64.exe" `
  -OutFile wall-vault.exe

.\wall-vault.exe setup
.\wall-vault.exe start
```

브라우저에서 `http://localhost:56243` 을 열면 키 금고 대시보드가 나타납니다.

---

## 다국어 · Languages

setup 마법사, 시스템 메시지, 대시보드 UI가 다음 언어로 표시됩니다.

| 코드 | 언어 | Code | Language |
|------|------|------|----------|
| `ko` | 한국어 | `ar` | العربية |
| `en` | English | `pt` | Português |
| `zh` | 中文 | `fr` | Français |
| `es` | Español | `de` | Deutsch |
| `hi` | हिन्दी | `ja` | 日本語 |

```bash
WV_LANG=en ./wall-vault setup   # 영어로 설치
WV_LANG=ja ./wall-vault setup   # 일본어로 설치
```

대시보드에서 언어 스위처로 실시간 전환 가능합니다 (페이지 리로드 없음).

---

## 사용법

```bash
wall-vault setup                      # 대화형 설치 마법사 (처음 시작)
wall-vault start                      # 프록시 + 키 금고 동시 실행
wall-vault proxy                      # 프록시만 실행
wall-vault proxy --key-google=AIza... # API 키 직접 전달
wall-vault vault                      # 키 금고만 실행
wall-vault doctor check               # 서비스 상태 확인
wall-vault doctor status              # 상세 보고서
wall-vault doctor fix                 # 자동 복구
wall-vault doctor deploy              # systemd 서비스 파일 생성 (Linux)
wall-vault doctor deploy launchd      # launchd plist 생성 (macOS)
wall-vault doctor deploy windows      # NSSM 서비스 스크립트 생성 (Windows)
```

### proxy 플래그 · Proxy Flags

| 플래그 | 환경변수 | 설명 |
|--------|----------|------|
| `--key-google` | `WV_KEY_GOOGLE` | Google API 키 |
| `--key-openrouter` | `WV_KEY_OPENROUTER` | OpenRouter API 키 |
| `--vault` | `WV_VAULT_URL` | 키 금고 URL |
| `--vault-token` | `WV_VAULT_TOKEN` | 금고 인증 토큰 |
| `--filter` | — | 도구 필터 (`strip_all` / `whitelist` / `passthrough`) |
| `--port` | `WV_PROXY_PORT` | 프록시 포트 |
| `--id` | `VAULT_CLIENT_ID` | 클라이언트 ID |

---

## 아키텍처

```
              ┌──────────────────────────┐
              │   키 금고 (:56243)        │
              │   API 키 AES-GCM 암호화   │
              │   SSE 브로드캐스트         │
              └───────────┬──────────────┘
                          │ SSE 실시간 동기화 (1-3초)
       ┌──────────────────┼──────────────────┐
       ▼                  ▼                  ▼
  봇A (:56244)       봇B (:56244)       봇C (:56244)
   (프록시)           (프록시)           (프록시)
       │                  │                  │
       └──────────────────┴──────────────────┘
                          │ 폴백 체인
       ┌───────────┬───────┴───────┬──────────────┐
       ▼           ▼               ▼              ▼
    Google      OpenAI        OpenRouter      Ollama (최종)
```

### 폴백 체인 · Fallback Chain

```
1단계: 지정 서비스 (클라이언트 설정 기준)
2단계: 나머지 서비스 순서대로
3단계: Ollama (최종 폴백 — 인터넷이 끊겨도 살아남는다)
```

### 쿨다운 · Cooldown

| HTTP 오류 | 대기 시간 |
|-----------|----------|
| 429 Too Many Requests | 30분 |
| 400/401/402/403 | 24시간 |
| 네트워크 오류 | 10분 |

---

## 설정

```bash
# 대화형 마법사 (권장)
./wall-vault setup

# 또는 예제 복사 후 수동 편집
cp configs/example-standalone.yaml wall-vault.yaml
```

### 주요 설정 (`wall-vault.yaml`)

```yaml
mode: standalone   # standalone | distributed
lang: ko           # ko | en | zh | es | hi | ar | pt | fr | de | ja
theme: light       # light | dark | gold | cherry | ocean

proxy:
  port: 56244
  client_id: my-bot
  vault_url: ""              # 분산 모드: http://키금고서버:56243
  vault_token: ""            # 분산 모드: 클라이언트 토큰
  tool_filter: strip_all     # strip_all | whitelist | passthrough
  services: [google, openrouter, ollama]
  timeout: 60s

vault:
  port: 56243
  admin_token: ""            # 빈칸이면 인증 없음 (로컬 개발용)
  master_password: ""        # API 키 암호화 비밀번호
  data_dir: ~/.wall-vault/data
```

### 환경변수 · Environment Variables

| 변수 | 설명 |
|------|------|
| `WV_LANG` | 언어 (ko/en/zh/es/hi/ar/pt/fr/de/ja) |
| `WV_THEME` | 테마 (light/dark/gold/cherry/ocean) |
| `WV_PROXY_PORT` | 프록시 포트 오버라이드 |
| `WV_VAULT_PORT` | 금고 포트 오버라이드 |
| `WV_VAULT_URL` | 키 금고 URL (분산 모드) |
| `WV_VAULT_TOKEN` | 프록시 인증 토큰 |
| `WV_ADMIN_TOKEN` | 관리자 토큰 |
| `WV_MASTER_PASS` | 암호화 마스터 비밀번호 |
| `WV_KEY_GOOGLE` | Google API 키 (쉼표 구분 복수 키) |
| `WV_KEY_OPENROUTER` | OpenRouter API 키 |

---

## 지원 서비스

### 클라우드 API · Cloud APIs

| 서비스 ID | 이름 | 모델 수 |
|-----------|------|---------|
| `google` | Google Gemini | 6개 고정 |
| `openai` | OpenAI | 8개 고정 |
| `anthropic` | Anthropic | 6개 고정 |
| `openrouter` | OpenRouter | 340개+ (동적) |
| `github-copilot` | GitHub Copilot | 6개 고정 |

### 로컬 AI · Local AI

| 서비스 ID | 이름 | 기본 포트 |
|-----------|------|-----------|
| `ollama` | Ollama | 11434 |
| `lmstudio` | LM Studio | 1234 |
| `vllm` | vLLM | 8000 |
| (커스텀) | 직접 추가 | 임의 |

로컬 서버는 대시보드 **서비스** 카드에서 URL 설정 및 모델 자동 감지 가능.

---

## API 엔드포인트

상세 문서: [docs/API.md](docs/API.md)

### 프록시 (`:56244`) · Proxy

| 경로 | 설명 |
|------|------|
| `POST /google/v1beta/models/{m}:generateContent` | Gemini API 프록시 |
| `POST /google/v1beta/models/{m}:streamGenerateContent` | Gemini 스트리밍 |
| `POST /v1/chat/completions` | OpenAI 호환 API |
| `GET /health` | 헬스체크 |
| `GET /status` | 상태 조회 |
| `GET /api/models` | 모델 목록 |
| `PUT /api/config/model` | 모델 변경 |
| `POST /reload` | 설정 새로고침 |

### 키 금고 (`:56243`) · Key Vault

| 경로 | 인증 | 설명 |
|------|------|------|
| `GET /` | — | 대시보드 UI |
| `GET /api/status` | — | 상태 조회 |
| `GET /api/events` | — | SSE 스트림 |
| `GET /api/keys` | 클라이언트 토큰 | 복호화된 키 목록 (IP 화이트리스트 적용) |
| `POST /api/heartbeat` | 클라이언트 토큰 | 프록시 상태 전송 |
| `GET /admin/keys` | 관리자 | 키 목록 |
| `POST /admin/keys` | 관리자 | 키 추가 |
| `DELETE /admin/keys/{id}` | 관리자 | 키 삭제 |
| `POST /admin/keys/reset` | 관리자 | 일일 사용량 초기화 |
| `GET /admin/clients` | 관리자 | 클라이언트 목록 |
| `POST /admin/clients` | 관리자 | 클라이언트 추가 |
| `PUT /admin/clients/{id}` | 관리자 | 클라이언트 수정 |
| `DELETE /admin/clients/{id}` | 관리자 | 클라이언트 삭제 |
| `GET /admin/services` | 관리자 | 서비스 목록 |
| `POST /admin/services` | 관리자 | 커스텀 서비스 추가 |
| `PUT /admin/services/{id}` | 관리자 | 서비스 업데이트 |
| `DELETE /admin/services/{id}` | 관리자 | 커스텀 서비스 삭제 |
| `GET /admin/models` | 관리자 | 모델 목록 (캐시) |
| `GET /admin/proxies` | 관리자 | 프록시 Heartbeat 상태 |
| `PUT /admin/theme` | 관리자 | 테마 변경 |
| `PUT /admin/lang` | 관리자 | 언어 변경 |

---

## 모드

### Standalone (단독 봇)

```bash
# 환경변수로 키 전달
WV_KEY_GOOGLE=AIza... ./wall-vault start

# 플래그로
./wall-vault proxy --key-google=AIza...
```

### Distributed (멀티 봇)

```bash
# [미니 서버] 키 금고 실행
./wall-vault vault

# [각 봇] 프록시 연결
WV_VAULT_URL=http://192.168.x.x:56243 \
WV_VAULT_TOKEN=내-봇-토큰 \
./wall-vault proxy
```

금고에서 설정을 변경하면 1–3초 안에 모든 봇에 SSE로 자동 반영된다. **재시작 불필요.**

---

## 자동 시작 설정

### Linux — systemd

```bash
wall-vault doctor deploy
systemctl --user daemon-reload
systemctl --user enable --now wall-vault
```

### macOS — launchd

```bash
wall-vault doctor deploy launchd
launchctl load ~/Library/LaunchAgents/com.wall-vault.plist
```

### Windows — NSSM

1. [NSSM](https://nssm.cc/download) 다운로드 후 PATH에 추가
2. 서비스 스크립트 생성: `.\wall-vault.exe doctor deploy windows`
3. 관리자 PowerShell에서 실행: `%USERPROFILE%\install-wall-vault-service.bat`

---

## 서비스 플러그인

코드 수정 없이 YAML 파일로 새 AI 서비스를 추가할 수 있다.

```yaml
# configs/services/my-service.yaml
id: my-service
name: My AI Service
enabled: true
endpoints:
  generate: https://api.example.com/v1/chat
auth:
  type: bearer
request_format: openai   # gemini | openai | ollama | raw
error_codes:
  429:
    cooldown: 30m
  401:
    cooldown: 24h
usage_threshold: 97
```

또는 대시보드 UI의 **서비스** 카드에서 직접 추가.

---

## OpenClaw 연동

**OpenClaw**는 페르소나와 장기기억을 가진 AI 에이전트를 여러 기기에서 분산 운용하는 프레임워크입니다. wall-vault는 OpenClaw를 위해 태어났으며, 두 시스템은 깊게 통합됩니다.

### 에이전트 등록

대시보드 **에이전트 추가** 모달에서 종류를 `openclaw`로 선택하면:
- 작업 디렉토리가 `~/.openclaw`로 자동 제안됩니다
- wall-vault가 해당 에이전트의 API 키 공급원·프록시가 됩니다

```bash
# OpenClaw 에이전트용 프록시 실행 예시
VAULT_CLIENT_ID=bot-a \
VAULT_URL=http://192.168.x.x:56243 \
wall-vault proxy
```

### Unix 소켓 연동 (TUI 실시간 알림)

wall-vault는 Unix 도메인 소켓으로 OpenClaw TUI에 이벤트를 즉시 전달합니다.

```yaml
# wall-vault.yaml
hooks:
  openclaw_socket: ~/.openclaw/wall-vault.sock
```

| 이벤트 | 발생 조건 |
|--------|----------|
| `model_changed` | 모델 전환 시 |
| `key_exhausted` | API 키 일일 한도 초과 |
| `service_down` | 서비스 장애·쿨다운 진입 |
| `ollama_waiting` | Ollama 로컬 모델 응답 대기 |
| `ollama_done` | Ollama 응답 완료 |
| `tui_footer` | TUI 하단 상태 메시지 |

이벤트 페이로드 형식:

```json
{
  "type": "model_changed",
  "timestamp": "2026-03-13T00:00:00Z",
  "data": { "model": "gemini-2.5-flash", "service": "google" }
}
```

### openclaw.json 프로바이더 설정

wall-vault를 OpenClaw의 커스텀 프로바이더로 등록하면 에이전트 대시보드의 **🐾 버튼**으로 설정 스니펫을 자동 복사할 수 있습니다.

```json5
// ~/.openclaw/openclaw.json
{
  models: {
    providers: {
      "wall-vault": {
        baseUrl: "http://localhost:56244/v1",
        apiKey: "YOUR_AGENT_TOKEN",
        api: "openai-completions",
        models: [
          { id: "wall-vault/gemini-2.5-flash" },
          { id: "wall-vault/gemini-2.5-pro" },
          { id: "wall-vault/gemini-2.0-flash" }
        ]
      }
    }
  },
  agents: {
    defaults: {
      model: { primary: "wall-vault/gemini-2.5-flash" }
    }
  }
}
```

- `baseUrl`: wall-vault 프록시 주소 (포트 56244)
- `apiKey`: 에이전트 카드의 **토큰** 값
- 모델 ID 앞에 `wall-vault/` 접두어를 붙여 자동 라우팅
- `wall-vault/gemini-*` → Google Gemini 직접 호출
- `wall-vault/gpt-*`, `wall-vault/o3` 등 → OpenAI 직접 호출
- `wall-vault/claude-*` → OpenRouter 경유 Anthropic
- `wall-vault/hunter-alpha`, `wall-vault/healer-alpha` → OpenRouter (OpenClaw 3.11 무료 1M context 모델)
- `wall-vault/kimi-*`, `wall-vault/glm-*`, `wall-vault/deepseek-*` 등 → OpenRouter
- OpenClaw 3.11 프로바이더 접두어 모두 지원: `opencode/`, `opencode-go/`, `moonshot/`, `kimi-coding/`, `minimax/`, `groq/`, `mistral/`, `deepseek/`, `qwen/`, `meta-llama/` 등
- Ollama `:cloud` 접미사 자동 처리 (`kimi-k2.5:cloud` → OpenRouter 경유)

### SSE 자동 동기화

OpenClaw 에이전트는 wall-vault SSE 스트림을 구독해 모델·서비스 변경을 **1–3초** 내에 자동 반영합니다. `openclaw.json` 설정 파일도 자동 업데이트되므로 에이전트를 재시작할 필요가 없습니다.

### 권장 디렉토리 구성

```
~/.openclaw/
├── openclaw.json        ← 에이전트 설정 (SSE 자동 갱신)
├── wall-vault.sock      ← 소켓 (wall-vault가 생성)
├── unified-proxy        ← wall-vault 프록시 바이너리
└── doctor.sh            ← 헬스체크 스크립트
```

---

## 빌드

```bash
# 현재 OS용
make build

# 전체 플랫폼 크로스컴파일
make build-all
# → bin/wall-vault-linux-amd64
# → bin/wall-vault-linux-arm64
# → bin/wall-vault-darwin-amd64
# → bin/wall-vault-darwin-arm64
# → bin/wall-vault-windows-amd64.exe

# 테스트 (39개)
make test

# 로컬 설치
make install  # ~/.local/bin/wall-vault
```

---

## 프로젝트 구조

```
wall-vault/
├── main.go                      # 진입점 + 서브커맨드 라우터
├── cmd/
│   ├── proxy/proxy.go           # proxy 서브커맨드
│   ├── vault/vault.go           # vault 서브커맨드
│   ├── setup/setup.go           # setup 마법사 (10개 언어)
│   └── doctor/doctor.go         # doctor 서브커맨드
├── internal/
│   ├── config/
│   │   ├── config.go            # 설정 로드·저장
│   │   └── services.go          # 서비스 플러그인 로더
│   ├── proxy/
│   │   ├── server.go            # 프록시 HTTP 서버 + 폴백 체인
│   │   ├── keymgr.go            # 라운드 로빈 키 관리자
│   │   ├── convert.go           # Gemini↔OpenAI↔Ollama 변환
│   │   └── toolfilter.go        # 도구 보안 필터
│   ├── vault/
│   │   ├── server.go            # 키 금고 HTTP 서버 + rate limiter
│   │   ├── store.go             # AES-GCM 암호화 저장소
│   │   ├── models.go            # 데이터 모델
│   │   ├── broker.go            # SSE 브로드캐스터
│   │   └── ui.go                # 대시보드 HTML (테마·다국어)
│   ├── doctor/doctor.go         # 자동복구
│   ├── models/registry.go       # 모델 레지스트리 (340개+)
│   ├── i18n/i18n.go             # 다국어 (10개 언어)
│   └── hooks/hooks.go           # 이벤트 훅 시스템
├── configs/
│   ├── services/                # 서비스 플러그인 YAML
│   ├── example-standalone.yaml
│   └── example-distributed.yaml
└── docs/
    ├── logo.png                 # 프로젝트 로고
    ├── API.md                   # API 상세 문서
    └── MANUAL.md                # 사용자 가이드
```

---

## 💀 Origin Story: The Night the Bots Died

It was the middle of the night.

One alert. Something felt wrong.

Logged in — all API keys invalidated. vault.json empty. The bots had gone silent. Of course they had. **Their memories had been completely erased.**

> *Motoko knew my work style inside-out. Mini prepared morning briefings every day. Raz handled everything quietly from a Raspberry Pi in the corner. Two weeks of careful cultivation. Two weeks of patience, tuning, and personality-shaping.*
>
> *A single hacker broke into the lab's internal network and torched all of it.*
>
> *It felt like coming home to find a beloved pet had simply vanished.*

It took a week to restore most of the memories. Not all of them came back.

This could never happen again.

So I built something. **A vault for the keys. A wall for the bots. A guarantee that no single attack could ever end everything again.**

---

## ⚔️ What It Actually Is

One line: **"A bodyguard that keeps your AI bots alive no matter what."**

```
Hacker steals a key?       → Vault blocks it. Rotates to the next.
Key hits its daily limit?  → Automatically switches. No downtime.
Service goes dark?         → Falls back: Gemini → OpenAI → Ollama
Running 100 bots?          → Change one setting. All bots updated in 1–3s.
```

In more detail:

- 🔐 **Key Vault**: AES-GCM encrypted storage. Round-robin rotation. Quota, cooldown, and error handling — all automatic.
- 🔀 **AI Proxy**: Accepts requests from OpenClaw, Claude Code, VS Code, your scripts — routes them to Gemini / OpenAI / Ollama. One dies, the next one picks up.
- ⚡ **SSE Real-time Sync**: Change anything in the vault, every connected bot reflects it instantly. No restarts.
- 🛡️ **Security Filter**: Full function calling block. Stops external skills from hijacking your AI.
- 🦞 **OpenClaw Integration**: Live events over Unix socket to TUI. Auto-updates openclaw.json.

Single Go binary. One bot or a dozen — fully covered.

---

## Table of Contents

- [Features](#features)
- [Quick Start](#quick-start)
- [Languages](#다국어--languages)
- [Architecture](#아키텍처)
- [Configuration](#설정)
- [Supported Services](#지원-서비스)
- [API Reference](#api-엔드포인트)
- [Modes](#모드)
- [Auto-Start](#자동-시작-설정)
- [OpenClaw Integration](#openclaw-integration)
- [Build](#빌드)
- [Project Structure](#프로젝트-구조)
- [License](#라이선스)

---

## Features

| Feature | Description |
|---------|-------------|
| **AI Proxy** | Google Gemini / OpenAI / Anthropic / OpenRouter / GitHub Copilot / Ollama / LMStudio / vLLM |
| **Key Vault** | API key management, usage monitoring, round-robin rotation |
| **AES-GCM Encryption** | Keys encrypted with master password, never stored in plaintext |
| **SSE Real-time Sync** | Vault ↔ proxy config sync within 1–3 seconds |
| **Tool Security Filter** | Block function calling (`strip_all` / `whitelist` / `passthrough`) |
| **Fallback Chain** | Auto-switch on service failure, final fallback to local Ollama |
| **Model Registry** | 340+ OpenRouter models + dynamic local model discovery |
| **Local AI Support** | Ollama / LM Studio / vLLM auto-detection + manual URL |
| **Service Management** | Add/edit/delete services from UI, custom service support |
| **Agent Management** | Per-agent service / model / IP whitelist / workdir |
| **Agent Status** | 4-state: 🟢Online / 🟡Delayed / 🔴Offline / ⚫Disconnected (heartbeat context shown) |
| **Per-type Config Copy** | 🦞 openclaw / 🟠 claude-code / ⌨ cursor / 💻 vscode — one-click config snippet |
| **Doctor** | Health check, auto-recovery, systemd/launchd/NSSM registration |
| **[Multi-language](#다국어--languages)** | 10 world languages |
| **Themes** | Light ☀️ / Dark 🌑 / Gold ✨ / Cherry 🌸 / Ocean 🌊 |
| **Cross-platform** | Linux / macOS / Windows / WSL |
| **[OpenClaw Integration](#openclaw-integration)** | Unix socket TUI events, agent auto-config |

---

## Quick Start

```bash
# Download (Linux amd64)
curl -L https://github.com/sookmook/wall-vault/releases/latest/download/wall-vault-linux-amd64 \
  -o wall-vault && chmod +x wall-vault

# Interactive setup wizard
./wall-vault setup

# Launch (proxy + vault)
./wall-vault start
```

Open `http://localhost:56243` to access the dashboard.

---

## OpenClaw Integration

**OpenClaw** is a distributed AI agent framework that runs personas with long-term memory across multiple devices. wall-vault was born to serve OpenClaw — the two systems are deeply integrated.

### Register an OpenClaw Agent

In the dashboard **Add Agent** modal, set the agent type to `openclaw`:
- Work directory auto-fills as `~/.openclaw`
- wall-vault becomes the API key supplier and proxy for that agent

```bash
VAULT_CLIENT_ID=bot-a \
VAULT_URL=http://192.168.x.x:56243 \
wall-vault proxy
```

### Unix Socket Events (TUI Live Notifications)

wall-vault sends real-time JSON events over a Unix domain socket to OpenClaw's TUI.

```yaml
# wall-vault.yaml
hooks:
  openclaw_socket: ~/.openclaw/wall-vault.sock
```

| Event | Trigger |
|-------|---------|
| `model_changed` | Model switch |
| `key_exhausted` | API key daily limit reached |
| `service_down` | Service failure / cooldown |
| `ollama_waiting` | Waiting for local Ollama response |
| `ollama_done` | Ollama response complete |
| `tui_footer` | Status message to TUI footer |

Event payload:

```json
{
  "type": "model_changed",
  "timestamp": "2026-03-13T00:00:00Z",
  "data": { "model": "gemini-2.5-flash", "service": "google" }
}
```

### openclaw.json Provider Config

Register wall-vault as a custom provider in OpenClaw. Use the **🐾 button** on any agent card to copy the config snippet automatically.

```json5
// ~/.openclaw/openclaw.json
{
  models: {
    providers: {
      "wall-vault": {
        baseUrl: "http://localhost:56244/v1",
        apiKey: "YOUR_AGENT_TOKEN",
        api: "openai-completions",
        models: [
          { id: "wall-vault/gemini-2.5-flash" },
          { id: "wall-vault/gemini-2.5-pro" },
          { id: "wall-vault/gemini-2.0-flash" }
        ]
      }
    }
  },
  agents: {
    defaults: {
      model: { primary: "wall-vault/gemini-2.5-flash" }
    }
  }
}
```

- `baseUrl`: wall-vault proxy address (port 56244)
- `apiKey`: the **token** value from the agent card
- Prefix model IDs with `wall-vault/` for automatic routing
- `wall-vault/gemini-*` → Google Gemini (direct)
- `wall-vault/gpt-*`, `wall-vault/o3`, etc. → OpenAI (direct)
- `wall-vault/claude-*` → Anthropic via OpenRouter
- `wall-vault/hunter-alpha`, `wall-vault/healer-alpha` → OpenRouter (OpenClaw 3.11 free 1M-context models)
- `wall-vault/kimi-*`, `wall-vault/glm-*`, `wall-vault/deepseek-*`, etc. → OpenRouter
- All OpenClaw 3.11 provider prefixes: `opencode/`, `opencode-go/`, `moonshot/`, `kimi-coding/`, `minimax/`, `groq/`, `mistral/`, `deepseek/`, `qwen/`, `meta-llama/`, and more
- Ollama `:cloud` suffix auto-handled (`kimi-k2.5:cloud` → OpenRouter)

### SSE Auto-Sync

OpenClaw agents subscribe to the wall-vault SSE stream and apply model/service changes within **1–3 seconds** — no restart needed. `openclaw.json` is updated automatically.

### Recommended Directory Layout

```
~/.openclaw/
├── openclaw.json        ← agent config (SSE auto-updated)
├── wall-vault.sock      ← socket file (created by wall-vault)
├── unified-proxy        ← wall-vault proxy binary
└── doctor.sh            ← health check script
```

---

## 🤓 Tech Stack

- **Language**: Go 1.22+ (single binary, zero runtime dependencies)
- **Encryption**: AES-256-GCM (crypto/rand nonce)
- **Realtime**: Server-Sent Events (SSE)
- **UI**: Server-rendered HTML (no frontend framework, no npm)
- **Tests**: 39 unit tests (crypto / proxy / vault / middleware / hooks)
- **CI/CD**: GitHub Actions (5-platform cross-compile + auto Release)

---

## 라이선스 · License

이 프로젝트는 **GNU General Public License v3.0 (GPL-3.0)** 라이선스를 따릅니다.
This project is licensed under the **GNU General Public License v3.0 (GPL-3.0)**.

> 저작권은 기본적으로 GPL 3.0 라이선스를 따릅니다.
>
> 개인적인 용도나 교육을 위한 활용은 얼마든지 가능합니다.
>
> 다만, 소스를 수정해 배포하시거나 상업적 용도로 사용하실 때에는 반드시 사전에 제작자에게 연락해 주시기 바랍니다.
>
> 제작자가 잡생각이 많고 놀기 좋아하는 게으른 성격이라, 새 기능 추가 요구가 있어도 버전업에 반영될지는 미지수입니다만... 계속 조르다 보면 언젠가 들어줄 수도 있으니, 필요한 기능이 있으시면 열심히 요구해 보시기 바랍니다. 한번 마음 먹으면 잘 하니까요. ㅋㅋㅋ

---

> The copyright follows the GPL 3.0 license.
>
> Personal use and educational use are fully permitted.
>
> However, if you wish to distribute modified versions or use this commercially, please contact the author beforehand.
>
> The author is a lazy daydreamer who loves to play, so whether new feature requests will make it into a release is anybody's guess — but keep nagging and maybe someday they'll get done. Once motivated, though, the work gets done well. lol

---

<p align="center">
  <b>sookmook · Sookmook Future Informatics Foundation</b><br>
  <i>"AI 봇의 기억은 소중하다. 지키자."</i><br>
  <i>"An AI bot's memory is precious. Protect it."</i>
</p>

---

*최종 업데이트 · Last updated: 2026-03-13*

---

## 🇨🇳 项目简介

> *"上个月，一名黑客入侵了我们实验室的内网，造成了严重破坏。"*
>
> *"精心培育了两周多的 AI 助手机器人的所有记忆，瞬间全部消失。"*
>
> *"就是为了这个，才有了这个项目。"*

**wall-vault** 是一个 AI 代理 + 密钥保险库一体化系统。

### 主要功能

| 功能 | 说明 |
|------|------|
| **AI 代理** | Google Gemini / OpenAI / Anthropic / OpenRouter / Ollama |
| **密钥保险库** | API 密钥 AES-GCM 加密存储，自动轮换 |
| **实时同步** | 保险库设置变更后 1–3 秒内同步到所有代理 |
| **安全过滤** | 阻断外部 function calling（strip_all 模式）|
| **自动故障转移** | 服务失败时自动切换，最终回退到本地 Ollama |
| **代理状态显示** | 4 级状态 🟢运行中 / 🟡延迟 / 🔴离线 / ⚫未连接（附 heartbeat 提示）|
| **按类型复制配置** | 🦞 openclaw / 🟠 claude-code / ⌨ cursor / 💻 vscode 一键复制代理配置 |
| **多语言支持** | 10 种语言（ko/en/zh/es/hi/ar/pt/fr/de/ja）|
| **主题** | Light ☀️ / Dark 🌑 / Gold ✨ / Cherry 🌸 / Ocean 🌊 |

### 快速开始

```bash
# Linux (amd64)
curl -L https://github.com/sookmook/wall-vault/releases/latest/download/wall-vault-linux-amd64 \
  -o wall-vault && chmod +x wall-vault

./wall-vault setup   # 交互式安装向导
./wall-vault start   # 启动代理 + 保险库
```

打开 `http://localhost:56243` 即可访问控制面板。

---

## 🇯🇵 はじめに

> *「先月、ハッカーが研究室の内部ネットワークに侵入し、大きな被害をもたらした。」*
>
> *「2週間以上かけて育ててきた AI アシスタントボットの記憶が、一瞬で全て消えてしまった。」*
>
> *「だから、このプロジェクトを始めた。」*

**wall-vault** は AI プロキシ + キー金庫の統合システムです。

### 主な機能

| 機能 | 説明 |
|------|------|
| **AI プロキシ** | Google Gemini / OpenAI / Anthropic / OpenRouter / Ollama |
| **キー金庫** | API キーを AES-GCM で暗号化・自動ローテーション |
| **リアルタイム同期** | 設定変更が 1〜3 秒以内に全プロキシへ反映 |
| **セキュリティフィルタ** | 外部 function calling を完全ブロック |
| **フォールバックチェーン** | 障害時に自動切換え、最終は Ollama |
| **エージェント状態** | 4段階 🟢実行中 / 🟡遅延 / 🔴オフライン / ⚫未接続（heartbeat 説明付き）|
| **タイプ別設定コピー** | 🦞 openclaw / 🟠 claude-code / ⌨ cursor / 💻 vscode ワンクリック設定コピー |
| **多言語対応** | 10 言語（ko/en/zh/es/hi/ar/pt/fr/de/ja）|
| **テーマ** | Light ☀️ / Dark 🌑 / Gold ✨ / Cherry 🌸 / Ocean 🌊 |

### クイックスタート

```bash
# Linux (amd64)
curl -L https://github.com/sookmook/wall-vault/releases/latest/download/wall-vault-linux-amd64 \
  -o wall-vault && chmod +x wall-vault

./wall-vault setup   # インタラクティブセットアップ
./wall-vault start   # プロキシ + 金庫を起動
```

ブラウザで `http://localhost:56243` を開くとダッシュボードが表示されます。

---

## 🇪🇸 Introducción

**wall-vault** es un sistema integrado de proxy de IA y bóveda de claves API.

### Características principales

| Función | Descripción |
|---------|-------------|
| **Proxy de IA** | Google Gemini / OpenAI / Anthropic / OpenRouter / Ollama |
| **Bóveda de claves** | Cifrado AES-GCM, rotación automática |
| **Sync en tiempo real** | Cambios reflejados en 1–3 segundos (SSE) |
| **Filtro de seguridad** | Bloqueo total de function calling externo |
| **Cadena de fallback** | Conmutación automática ante fallos |
| **Estado del agente** | 4 niveles 🟢Activo / 🟡Tardanza / 🔴Offline / ⚫Sin conectar (con guía heartbeat) |
| **Copia de config por tipo** | 🦞 openclaw / 🟠 claude-code / ⌨ cursor / 💻 vscode — snippet con un clic |
| **Temas** | Light ☀️ / Dark 🌑 / Gold ✨ / Cherry 🌸 / Ocean 🌊 |

### Inicio rápido

```bash
curl -L https://github.com/sookmook/wall-vault/releases/latest/download/wall-vault-linux-amd64 \
  -o wall-vault && chmod +x wall-vault
./wall-vault setup && ./wall-vault start
```

Abra `http://localhost:56243` en el navegador.

---

## 🇫🇷 Présentation

**wall-vault** est un système intégré proxy IA + coffre-fort de clés API.

### Fonctionnalités

| Fonction | Description |
|----------|-------------|
| **Proxy IA** | Google Gemini / OpenAI / Anthropic / OpenRouter / Ollama |
| **Coffre-fort** | Chiffrement AES-GCM, rotation automatique des clés |
| **Sync temps réel** | Changements reflétés en 1–3 secondes (SSE) |
| **Filtre sécurité** | Blocage total du function calling externe |
| **Chaîne de repli** | Basculement automatique en cas d'échec |
| **Statut agent** | 4 niveaux 🟢Actif / 🟡Retard / 🔴Hors ligne / ⚫Déconnecté (guide heartbeat inclus) |
| **Config par type** | 🦞 openclaw / 🟠 claude-code / ⌨ cursor / 💻 vscode — snippet en un clic |
| **Thèmes** | Light ☀️ / Dark 🌑 / Gold ✨ / Cherry 🌸 / Ocean 🌊 |

### Démarrage rapide

```bash
curl -L https://github.com/sookmook/wall-vault/releases/latest/download/wall-vault-linux-amd64 \
  -o wall-vault && chmod +x wall-vault
./wall-vault setup && ./wall-vault start
```

Ouvrez `http://localhost:56243` dans votre navigateur.

---

## 🇩🇪 Über das Projekt

**wall-vault** ist ein integriertes KI-Proxy- und API-Schlüssel-Tresor-System.

### Hauptfunktionen

| Funktion | Beschreibung |
|----------|--------------|
| **KI-Proxy** | Google Gemini / OpenAI / Anthropic / OpenRouter / Ollama |
| **Schlüsseltresor** | AES-GCM-Verschlüsselung, automatische Rotation |
| **Echtzeit-Sync** | Änderungen in 1–3 Sekunden übertragen (SSE) |
| **Sicherheitsfilter** | Vollständige Blockierung externen Function Callings |
| **Fallback-Kette** | Automatischer Wechsel bei Dienstausfall |
| **Agent-Status** | 4 Stufen 🟢Aktiv / 🟡Verzögert / 🔴Offline / ⚫Nicht verbunden (Heartbeat-Hinweis) |
| **Konfig-Kopie je Typ** | 🦞 openclaw / 🟠 claude-code / ⌨ cursor / 💻 vscode — Snippet per Klick |
| **Designs** | Light ☀️ / Dark 🌑 / Gold ✨ / Cherry 🌸 / Ocean 🌊 |

### Schnellstart

```bash
curl -L https://github.com/sookmook/wall-vault/releases/latest/download/wall-vault-linux-amd64 \
  -o wall-vault && chmod +x wall-vault
./wall-vault setup && ./wall-vault start
```

Öffnen Sie `http://localhost:56243` im Browser.
