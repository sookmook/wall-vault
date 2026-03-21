# wall-vault 사용자 매뉴얼
*(Last updated: 2026-03-20 — v0.1.15)*

---

## 목차

1. [wall-vault란?](#wall-vault란)
2. [설치](#설치)
3. [처음 시작하기 (setup 마법사)](#처음-시작하기)
4. [API 키 등록](#api-키-등록)
5. [프록시 사용법](#프록시-사용법)
6. [키 금고 대시보드](#키-금고-대시보드)
7. [분산 모드 (멀티 봇)](#분산-모드-멀티-봇)
8. [자동 시작 설정](#자동-시작-설정)
9. [Doctor 주치의](#doctor-주치의)
10. [환경변수 참고](#환경변수-참고)
11. [문제 해결](#문제-해결)

---

## wall-vault란?

**wall-vault = 오픈클로(OpenClaw)를 위한 AI 대리인(프록시) + API 키 금고**

AI 서비스를 사용하려면 **API 키**가 필요합니다. API 키란 "이 사람은 이 서비스를 쓸 자격이 있다"는 것을 증명하는 **디지털 출입증** 같은 것입니다. 그런데 이 출입증은 하루에 쓸 수 있는 횟수가 정해져 있고, 잘못 관리하면 노출될 위험도 있습니다.

wall-vault는 이 출입증들을 안전한 금고에 보관하고, 오픈클로(OpenClaw)와 AI 서비스 사이에서 **대리인(프록시)** 역할을 합니다. 쉽게 말해, 오픈클로는 wall-vault에만 연결하면 되고, 나머지 복잡한 일은 wall-vault가 알아서 처리해 줍니다.

wall-vault가 해결해 주는 문제들:

- **API 키 자동 순환**: 한 키의 사용량이 한도에 달하거나 잠시 막히면(쿨다운), 조용히 다음 키로 넘어갑니다. 오픈클로는 중단 없이 계속 작동합니다.
- **서비스 자동 교체(폴백)**: Google이 응답하지 않으면 OpenRouter로, 그것도 안 되면 내 컴퓨터에 설치된 Ollama(로컬 AI)로 자동 전환됩니다. 세션이 끊기지 않습니다.
- **실시간 동기화(SSE)**: 금고 대시보드에서 모델을 바꾸면 1~3초 안에 오픈클로 화면에 반영됩니다. SSE(Server-Sent Events)란 서버가 변화를 실시간으로 클라이언트에 밀어주는 기술입니다.
- **실시간 알림**: 키 소진이나 서비스 장애 같은 이벤트가 오픈클로 TUI(터미널 화면) 하단에 바로 표시됩니다.

> 💡 **Claude Code, Cursor, VS Code**도 연결해서 쓸 수 있지만, wall-vault의 본래 목적은 오픈클로와 함께 쓰는 것입니다.

```
오픈클로 (TUI 터미널 화면)
        │
        ▼
  wall-vault 프록시 (:56244)   ← 키 관리, 라우팅, 폴백, 이벤트
        │
        ├─ Google Gemini API
        ├─ OpenRouter API (340개 이상 모델)
        └─ Ollama (내 컴퓨터, 최후 보루)
```

---

## 설치

### Linux / macOS

터미널을 열고 아래 명령어를 그대로 붙여넣으세요.

```bash
# Linux (일반 PC, 서버 — amd64)
curl -L https://github.com/sookmook/wall-vault/releases/latest/download/wall-vault-linux-amd64 \
  -o ~/.local/bin/wall-vault && chmod +x ~/.local/bin/wall-vault

# macOS Apple Silicon (M1/M2/M3 맥)
curl -L https://github.com/sookmook/wall-vault/releases/latest/download/wall-vault-darwin-arm64 \
  -o /usr/local/bin/wall-vault && chmod +x /usr/local/bin/wall-vault
```

- `curl -L ...` — 파일을 인터넷에서 다운로드합니다.
- `chmod +x` — 다운로드한 파일을 "실행 가능"하게 만듭니다. 이 과정을 빠뜨리면 "권한 없음" 오류가 납니다.

### Windows

PowerShell(관리자 권한)을 열고 아래 명령어를 실행하세요.

```powershell
# 다운로드
Invoke-WebRequest -Uri `
  "https://github.com/sookmook/wall-vault/releases/latest/download/wall-vault-windows-amd64.exe" `
  -OutFile "$env:LOCALAPPDATA\Programs\wall-vault\wall-vault.exe"

# PATH 추가 (PowerShell 재시작 후 적용)
$env:PATH += ";$env:LOCALAPPDATA\Programs\wall-vault"
```

> 💡 **PATH란?** 컴퓨터가 명령어를 찾아보는 폴더 목록입니다. PATH에 추가해야 어느 폴더에서든 `wall-vault`라고 입력해서 실행할 수 있습니다.

### 소스에서 직접 빌드 (개발자용)

Go 언어 개발 환경이 설치된 경우에만 해당됩니다.

```bash
git clone https://github.com/sookmook/wall-vault
cd wall-vault
make build       # bin/wall-vault (버전: v0.1.6.YYYYMMDD.HHmmss)
make install     # ~/.local/bin/wall-vault
```

> 💡 **빌드 타임스탬프 버전**: `make build`로 빌드하면 버전이 `v0.1.6.20260314.231308`처럼 날짜·시각이 포함된 형식으로 자동 생성됩니다. `go build ./...`로 직접 빌드하면 버전이 `"dev"`로만 표시됩니다.

---

## 처음 시작하기

### setup 마법사 실행

설치 후 처음에는 반드시 아래 명령어로 **설정 마법사**를 실행하세요. 마법사가 필요한 항목들을 하나씩 물어보면서 안내해 줍니다.

```bash
wall-vault setup
```

마법사가 진행하는 단계는 다음과 같습니다:

```
1. 언어 선택 (한국어 포함 10개 언어)
2. 테마 선택 (light / dark / gold / cherry / ocean)
3. 운용 모드 — 혼자 쓸지(standalone), 여러 대에서 함께 쓸지(distributed) 선택
4. 봇 이름 입력 — 대시보드에 표시될 이름
5. 포트 설정 — 기본값: 프록시 56244, 금고 56243 (바꿀 필요 없으면 그냥 엔터)
6. AI 서비스 선택 — Google / OpenRouter / Ollama 중 쓸 서비스
7. 도구 보안 필터 설정
8. 관리자 토큰 설정 — 대시보드 관리 기능을 잠그는 비밀번호. 자동 생성도 가능
9. API 키 암호화 비밀번호 설정 — 키를 더 안전하게 저장하고 싶을 때 (선택 사항)
10. 설정 파일 저장 경로
```

> ⚠️ **관리자 토큰은 꼭 기억해 두세요.** 나중에 대시보드에서 키를 추가하거나 설정을 바꿀 때 필요합니다. 잃어버리면 설정 파일을 직접 수정해야 합니다.

마법사를 완료하면 `wall-vault.yaml` 설정 파일이 자동으로 생성됩니다.

### 실행

```bash
wall-vault start
```

아래 두 서버가 동시에 시작됩니다:

- **프록시** (`http://localhost:56244`) — 오픈클로와 AI 서비스 사이를 연결하는 대리인
- **키 금고** (`http://localhost:56243`) — API 키 관리 및 웹 대시보드

브라우저에서 `http://localhost:56243`을 열면 대시보드를 바로 확인할 수 있습니다.

---

## API 키 등록

API 키를 등록하는 방법은 네 가지입니다. **처음 시작하는 분께는 방법 1(환경변수)을 권장**합니다.

### 방법 1: 환경변수 (권장 — 가장 간단)

환경변수란 프로그램이 시작될 때 읽어 오는 **미리 설정해 둔 값**입니다. 터미널에서 아래처럼 입력하면 됩니다.

```bash
# Google Gemini 키 등록
export WV_KEY_GOOGLE=AIzaSy...

# OpenRouter 키 등록
export WV_KEY_OPENROUTER=sk-or-v1-...

# 등록 후 실행
wall-vault start
```

키를 여러 개 가지고 있다면 쉼표(,)로 연결하세요. wall-vault가 키들을 돌아가며 자동으로 사용합니다(라운드 로빈):

```bash
export WV_KEY_GOOGLE=AIzaSy...,AIzaSy...,AIzaSy...
```

> 💡 **팁**: `export` 명령은 현재 터미널 세션에만 적용됩니다. 컴퓨터를 재시작해도 유지되게 하려면 `~/.bashrc` 또는 `~/.zshrc` 파일에 위 줄을 추가하세요.

### 방법 2: 대시보드 UI (마우스로 클릭)

1. 브라우저에서 `http://localhost:56243` 접속
2. 상단 **🔑 API 키** 카드에서 `[+ 추가]` 버튼 클릭
3. 서비스 종류, 키 값, 레이블(메모용 이름), 일일 한도를 입력한 뒤 저장

### 방법 3: REST API (자동화·스크립트용)

REST API란 프로그램끼리 HTTP로 데이터를 주고받는 방식입니다. 스크립트로 자동 등록할 때 유용합니다.

```bash
curl -X POST http://localhost:56243/admin/keys \
  -H "Authorization: Bearer 관리자토큰" \
  -H "Content-Type: application/json" \
  -d '{
    "service": "google",
    "key": "AIzaSy...",
    "label": "메인 키",
    "daily_limit": 1000
  }'
```

### 방법 4: proxy 플래그 (잠깐 테스트할 때)

정식 등록 없이 임시로 키를 넣어서 테스트해 볼 때 씁니다. 프로그램을 종료하면 사라집니다.

```bash
wall-vault proxy --key-google=AIzaSy... --key-openrouter=sk-or-...
```

---

## 프록시 사용법

### OpenClaw에서 사용 (주목적)

오픈클로가 wall-vault를 통해 AI 서비스에 연결되도록 설정하는 방법입니다.

`~/.openclaw/openclaw.json` 파일을 열고 아래 내용을 추가하세요:

```json5
// ~/.openclaw/openclaw.json
{
  models: {
    providers: {
      "wall-vault": {
        baseUrl: "http://localhost:56244/v1",
        apiKey: "your-agent-token",   // vault 에이전트 토큰
        api: "openai-completions",
        models: [
          { id: "wall-vault/gemini-2.5-flash" },
          { id: "wall-vault/gemini-2.5-pro" },
          { id: "wall-vault/hunter-alpha" },    // 무료 1M context
          { id: "wall-vault/claude-opus-4-6" }
        ]
      }
    }
  }
}
```

> 💡 **더 쉬운 방법**: 대시보드 에이전트 카드의 **🦞 OpenClaw 설정 복사** 버튼을 누르면 토큰과 주소가 이미 채워진 스니펫이 클립보드에 복사됩니다. 그냥 붙여넣기만 하면 됩니다.

**모델 이름 앞의 `wall-vault/`는 어디로 연결될까요?**

모델 이름을 보면 wall-vault가 어떤 AI 서비스로 요청을 보낼지 자동으로 판단합니다:

| 모델 형식 | 연결되는 서비스 |
|----------|--------------|
| `wall-vault/gemini-*` | Google Gemini 직접 연결 |
| `wall-vault/gpt-*`, `wall-vault/o3`, `wall-vault/o4*` | OpenAI 직접 연결 |
| `wall-vault/claude-*` | OpenRouter를 통해 Anthropic 연결 |
| `wall-vault/hunter-alpha`, `wall-vault/healer-alpha` | OpenRouter (무료 100만 토큰 컨텍스트) |
| `wall-vault/kimi-*`, `wall-vault/glm-*`, `wall-vault/deepseek-*` | OpenRouter 연결 |
| `google/모델명`, `openai/모델명`, `anthropic/모델명` 등 | 해당 서비스 직접 연결 |
| `custom/google/모델명`, `custom/openai/모델명` 등 | `custom/` 부분을 제거하고 재라우팅 |
| `모델명:cloud` | `:cloud` 부분을 제거하고 OpenRouter 연결 |

> 💡 **컨텍스트(context)란?** AI가 한 번에 기억할 수 있는 대화 분량입니다. 1M(백만 토큰)이면 매우 긴 대화나 긴 문서도 한 번에 처리할 수 있습니다.

### Gemini API 형식으로 직접 연결 (기존 도구 호환)

Google Gemini API를 직접 쓰던 도구가 있다면, 주소만 wall-vault로 바꿔 주세요:

```bash
export ANTHROPIC_BASE_URL=http://localhost:56244/google
```

또는 URL을 직접 지정하는 도구라면:

```
http://localhost:56244/google/v1beta/models/gemini-2.5-flash:generateContent
```

### OpenAI SDK(파이썬)에서 사용

파이썬으로 AI를 활용하는 코드에서도 wall-vault를 연결할 수 있습니다. `base_url`만 바꾸면 됩니다:

```python
from openai import OpenAI

client = OpenAI(
    base_url="http://localhost:56244/v1",
    api_key="not-needed"  # API 키는 wall-vault가 알아서 관리합니다
)

response = client.chat.completions.create(
    model="google/gemini-2.5-flash",   # provider/model 형식으로 입력
    messages=[{"role": "user", "content": "안녕하세요"}]
)
```

### 실행 중에 모델 바꾸기

wall-vault가 이미 실행 중인 상태에서 사용할 AI 모델을 바꾸려면:

```bash
# 프록시에 직접 요청하여 모델 변경
curl -X PUT http://localhost:56244/api/config/model \
  -H "Content-Type: application/json" \
  -d '{"service": "openrouter", "model": "anthropic/claude-3.5-sonnet"}'

# 분산 모드(멀티 봇)에서는 금고 서버에서 변경 → SSE로 즉시 반영됨
curl -X PUT http://localhost:56243/admin/clients/내-봇-id \
  -H "Authorization: Bearer 관리자토큰" \
  -H "Content-Type: application/json" \
  -d '{"default_service": "google", "default_model": "gemini-2.5-pro"}'
```

### 사용 가능한 모델 목록 확인

```bash
# 전체 목록 보기
curl http://localhost:56244/api/models | python3 -m json.tool

# Google 모델만 보기
curl "http://localhost:56244/api/models?service=google"

# 이름으로 검색 (예: "claude"가 포함된 모델)
curl "http://localhost:56244/api/models?q=claude"
```

**서비스별 주요 모델 요약:**

| 서비스 | 주요 모델 |
|--------|----------|
| Google | gemini-2.5-pro, gemini-2.5-flash, gemini-2.5-flash-8b, gemini-2.0-flash |
| OpenAI | gpt-4o, gpt-4o-mini, o3, o1, o1-mini |
| OpenRouter | 346개 이상 (Hunter Alpha 1M 컨텍스트 무료, DeepSeek R1/V3, Qwen 2.5 등) |
| Ollama | 내 컴퓨터에 설치된 로컬 서버 자동 감지 |

---

## 키 금고 대시보드

브라우저에서 `http://localhost:56243`에 접속하면 대시보드를 볼 수 있습니다.

**화면 구성:**
- **상단 고정 바(topbar)**: 로고, 언어·테마 선택기, SSE 연결 상태 표시
- **카드 그리드**: 에이전트·서비스·API 키 카드들이 타일 형태로 배치

### API 키 카드

등록된 API 키들을 한눈에 관리할 수 있는 카드입니다.

- 서비스별로 구분해서 키 목록을 보여 줍니다.
- `today_usage`: 오늘 성공적으로 처리된 토큰(AI가 읽고 쓴 글자 수) 수
- `today_attempts`: 오늘 총 호출 횟수 (성공 + 실패 포함)
- `[+ 추가]` 버튼으로 새 키를 등록하고, `✕`로 키를 삭제합니다.

> 💡 **토큰(token)이란?** AI가 텍스트를 처리할 때 사용하는 단위입니다. 대략 영어 단어 하나, 또는 한글 1~2글자 정도에 해당합니다. API 요금은 보통 이 토큰 수에 따라 계산됩니다.

### 에이전트 카드

wall-vault 프록시에 연결된 봇(에이전트)들의 상태를 보여 주는 카드입니다.

**연결 상태는 4단계로 표시됩니다:**

| 표시 | 상태 | 의미 |
|------|------|------|
| 🟢 | 실행 중 | 프록시가 정상적으로 동작하고 있음 |
| 🟡 | 지연 | 응답은 오지만 느림 |
| 🔴 | 오프라인 | 프록시가 응답하지 않음 |
| ⚫ | 미연결·비활성화 | 프록시가 금고에 연결된 적 없거나 비활성화됨 |

**에이전트 카드 하단 버튼 안내:**

에이전트를 등록할 때 **에이전트 종류**를 지정하면, 해당 종류에 맞는 편의 버튼이 자동으로 나타납니다.

---

#### 🔘 설정 복사 버튼 — 연결 설정을 자동으로 만들어 줍니다

버튼을 클릭하면 해당 에이전트의 토큰, 프록시 주소, 모델 정보가 이미 채워진 설정 스니펫이 클립보드에 복사됩니다. 복사한 내용을 아래 표의 위치에 붙여넣기만 하면 연결 설정이 완료됩니다.

| 버튼 | 에이전트 종류 | 붙여넣을 위치 |
|------|-------------|-------------|
| 🦞 OpenClaw 설정 복사 | `openclaw` | `~/.openclaw/openclaw.json` |
| 🦀 NanoClaw 설정 복사 | `nanoclaw` | `~/.openclaw/openclaw.json` |
| 🟠 Claude Code 설정 복사 | `claude-code` | `~/.claude/settings.json` |
| ⌨ Cursor 설정 복사 | `cursor` | Cursor → Settings → AI |
| 💻 VSCode 설정 복사 | `vscode` | `~/.continue/config.json` |

**예시 — Claude Code 타입이면 이런 내용이 복사됩니다:**

```json
// ~/.claude/settings.json
{
  "apiProvider": "openai",
  "baseUrl": "http://192.168.1.20:56244/v1",
  "apiKey": "이-에이전트의-토큰"
}
```

**예시 — VSCode (Continue) 타입이면:**

```yaml
# ~/.continue/config.yaml  ← config.json이 아닌 config.yaml 에 붙여넣기
name: My Config
version: 0.0.1
schema: v1

models:
  - name: wall-vault proxy
    provider: openai
    model: gemini-2.5-flash
    apiBase: http://192.168.1.20:56244/v1
    apiKey: 이-에이전트의-토큰
    roles:
      - chat
      - edit
      - apply
```

> ⚠️ **Continue 최신 버전은 `config.yaml`을 사용합니다.** `config.yaml`이 존재하면 `config.json`은 완전히 무시됩니다. 반드시 `config.yaml`에 붙여넣으세요.

**예시 — Cursor 타입이면:**

```
Base URL : http://192.168.1.20:56244/v1
API Key  : 이-에이전트의-토큰

// 또는 환경변수:
OPENAI_BASE_URL=http://192.168.1.20:56244/v1
OPENAI_API_KEY=이-에이전트의-토큰
```

> ⚠️ **클립보드 복사가 안 될 때**: 브라우저 보안 정책으로 복사가 막히는 경우가 있습니다. 팝업으로 텍스트박스가 열리면 Ctrl+A로 전체 선택 후 Ctrl+C로 복사하세요.

---

#### 🟣 배포 명령어 복사 버튼 — 새 머신에 설치할 때 씁니다

새 컴퓨터에 wall-vault 프록시를 처음 설치하고 금고에 연결할 때 사용합니다. 버튼을 클릭하면 설치 스크립트 전체가 복사됩니다. 새 컴퓨터의 터미널에 붙여넣고 실행하면 다음이 한 번에 처리됩니다:

1. wall-vault 바이너리 설치 (이미 설치되어 있으면 건너뜀)
2. systemd 사용자 서비스 자동 등록
3. 서비스 시작 및 금고 자동 연결

> 💡 스크립트 안에 이 에이전트의 토큰과 금고 서버 주소가 이미 채워져 있으므로, 붙여넣기 후 별도 수정 없이 바로 실행할 수 있습니다.

---

### 서비스 카드

사용할 AI 서비스를 켜고 끄거나 설정하는 카드입니다.

- 서비스별 활성화·비활성화 토글 스위치
- 로컬 AI 서버(내 컴퓨터에서 돌리는 Ollama, LM Studio, vLLM 등)의 주소를 입력하면 사용 가능한 모델을 자동으로 찾아 줍니다.
- **로컬 서비스 연결 상태 표시**: 서비스 이름 옆 ● 점이 **초록색**이면 연결됨, **회색**이면 연결 안 됨
- **체크박스 자동 동기화**: 페이지를 열 때 로컬 서비스(Ollama 등)가 실행 중이면 자동으로 체크 상태가 됩니다.

> 💡 **로컬 서비스가 다른 컴퓨터에서 실행 중이라면**: 서비스 URL 입력란에 그 컴퓨터의 IP를 입력하세요. 예: `http://192.168.1.20:11434` (Ollama), `http://192.168.1.20:1234` (LM Studio)

### 관리자 토큰 입력

대시보드에서 키 추가·삭제처럼 중요한 기능을 사용하려고 하면 관리자 토큰 입력 팝업이 나타납니다. setup 마법사에서 설정했던 토큰을 입력하세요. 한 번 입력하면 브라우저를 닫기 전까지 유지됩니다.

> ⚠️ **인증 실패가 15분 내 10회를 초과하면 해당 IP가 일시 차단됩니다.** 토큰을 잊으셨다면 `wall-vault.yaml` 파일에서 `admin_token` 항목을 확인하세요.

---

## 분산 모드 (멀티 봇)

여러 대의 컴퓨터에서 오픈클로를 동시에 운영할 때, **하나의 키 금고를 공유**하는 구성입니다. 키 관리를 한 곳에서만 하면 되므로 편리합니다.

### 구성 예시

```
[키 금고 서버]
  wall-vault vault    (키 금고 :56243, 대시보드)

[WSL 알파]            [라즈베리파이 감마]    [맥미니 로컬]
  wall-vault proxy      wall-vault proxy        wall-vault proxy
  openclaw TUI          openclaw TUI            openclaw TUI
  ↕ SSE 동기화          ↕ SSE 동기화            ↕ SSE 동기화
```

모든 봇이 가운데 금고 서버를 바라보고 있어서, 금고에서 모델을 바꾸거나 키를 추가하면 모든 봇에 즉시 반영됩니다.

### 1단계: 키 금고 서버 시작

금고 서버로 쓸 컴퓨터에서 실행합니다:

```bash
wall-vault vault
```

### 2단계: 각 봇(클라이언트) 등록

금고 서버에 접속하는 각 봇의 정보를 미리 등록해 둡니다:

```bash
curl -X POST http://localhost:56243/admin/clients \
  -H "Authorization: Bearer 관리자토큰" \
  -H "Content-Type: application/json" \
  -d '{
    "id": "botA",
    "name": "봇A",
    "token": "bota-secret",
    "default_service": "google",
    "default_model": "gemini-2.5-flash"
  }'
```

### 3단계: 각 봇 컴퓨터에서 프록시 시작

봇이 설치된 각 컴퓨터에서 금고 서버 주소와 토큰을 지정해 프록시를 실행합니다:

```bash
WV_VAULT_URL=http://192.168.x.x:56243 \
WV_VAULT_TOKEN=bota-secret \
WV_VAULT_CLIENT_ID=botA \
wall-vault proxy
```

> 💡 **`192.168.x.x`** 부분은 금고 서버 컴퓨터의 실제 내부 IP 주소로 바꾸세요. 공유기 설정 또는 `ip addr` 명령어로 확인할 수 있습니다.

---

## 자동 시작 설정

컴퓨터를 재시작할 때마다 수동으로 wall-vault를 켜는 것이 번거롭다면, 시스템 서비스로 등록해 두세요. 한 번 등록하면 부팅 시 자동으로 시작됩니다.

### Linux — systemd (대부분의 리눅스)

systemd는 리눅스에서 프로그램을 자동으로 시작·관리하는 시스템입니다:

```bash
wall-vault doctor deploy
systemctl --user daemon-reload
systemctl --user enable --now wall-vault
```

로그 확인:

```bash
journalctl --user -u wall-vault -f
```

### macOS — launchd

macOS에서 프로그램 자동 실행을 담당하는 시스템입니다:

```bash
wall-vault doctor deploy launchd
launchctl load ~/Library/LaunchAgents/com.wall-vault.plist
```

### Windows — NSSM

1. [nssm.cc](https://nssm.cc/download)에서 NSSM을 다운로드해 PATH에 추가합니다.
2. 관리자 권한 PowerShell에서:

```powershell
wall-vault doctor deploy windows
```

---

## Doctor 주치의

`doctor` 명령은 wall-vault가 올바르게 설정되어 있는지 **스스로 진단하고 고쳐 주는 도구**입니다.

```bash
wall-vault doctor check   # 현재 상태 진단 (읽기만 함, 아무것도 변경하지 않음)
wall-vault doctor fix     # 문제를 자동으로 복구
wall-vault doctor all     # 진단 + 자동 복구 한 번에
```

> 💡 뭔가 이상한 것 같다면 `wall-vault doctor all`을 먼저 실행해 보세요. 많은 문제를 자동으로 잡아 줍니다.

---

## 환경변수 참고

환경변수는 프로그램에 설정값을 전달하는 방법입니다. `export 변수명=값` 형태로 터미널에 입력하거나, 자동 시작 서비스 파일에 넣어 두면 항상 적용됩니다.

| 변수 | 설명 | 예시 값 |
|------|------|---------|
| `WV_LANG` | 대시보드 언어 | `ko`, `en`, `ja` |
| `WV_THEME` | 대시보드 테마 | `light`, `dark`, `gold` |
| `WV_KEY_GOOGLE` | Google API 키 (쉼표로 여러 개) | `AIza...,AIza...` |
| `WV_KEY_OPENROUTER` | OpenRouter API 키 | `sk-or-v1-...` |
| `WV_VAULT_URL` | 분산 모드에서 금고 서버 주소 | `http://192.168.x.x:56243` |
| `WV_VAULT_TOKEN` | 클라이언트(봇) 인증 토큰 | `my-secret-token` |
| `WV_ADMIN_TOKEN` | 관리자 토큰 | `admin-token-here` |
| `WV_MASTER_PASS` | API 키 암호화 비밀번호 | `my-password` |
| `WV_AVATAR` | 아바타 이미지 파일 경로 (`~/.openclaw/` 기준 상대경로) | `workspace/avatars/avatar.png` |
| `OLLAMA_URL` | Ollama 로컬 서버 주소 | `http://192.168.x.x:11434` |

---

## 문제 해결

### 프록시가 시작되지 않을 때

포트가 이미 다른 프로그램에 의해 사용 중인 경우가 많습니다.

```bash
ss -tlnp | grep 56244   # 56244 포트를 누가 쓰고 있는지 확인
wall-vault proxy --port 8080   # 다른 포트 번호로 시작
```

### API 키 오류가 날 때 (429, 402, 401, 403, 582)

| 오류 코드 | 의미 | 대처 방법 |
|----------|------|----------|
| **429** | 요청이 너무 많음 (사용량 초과) | 잠시 기다리거나 다른 키 추가 |
| **402** | 결제 필요 또는 크레딧 부족 | 해당 서비스에서 크레딧 충전 |
| **401 / 403** | 키가 잘못되었거나 권한 없음 | 키 값 재확인 후 재등록 |
| **582** | 게이트웨이 과부하 (쿨다운 5분) | 5분 후 자동 해제됨 |

```bash
# 등록된 키 목록 및 상태 확인
curl -H "Authorization: Bearer 관리자토큰" http://localhost:56243/admin/keys

# 키 사용량 카운터 초기화
curl -X POST -H "Authorization: Bearer 관리자토큰" http://localhost:56243/admin/keys/reset
```

### 에이전트가 "미연결"로 표시될 때

"미연결"은 프록시 프로세스가 금고에 신호(heartbeat)를 보내지 않는 상태입니다. **설정이 저장되지 않았다는 뜻이 아닙니다.** 프록시가 금고 서버 주소와 토큰을 알고 실행되어야 연결 상태로 바뀝니다.

```bash
# 금고 서버 주소, 토큰, 클라이언트 ID를 지정해서 프록시 시작
WV_VAULT_URL=http://금고서버주소:56243 \
WV_VAULT_TOKEN=클라이언트토큰 \
WV_VAULT_CLIENT_ID=클라이언트ID \
wall-vault proxy
```

연결에 성공하면 약 20초 안에 대시보드에서 🟢 실행 중으로 바뀝니다.

### Ollama 연결이 안 될 때

Ollama는 내 컴퓨터에서 직접 AI를 실행하는 프로그램입니다. 먼저 Ollama가 켜져 있는지 확인하세요.

```bash
curl http://localhost:11434/api/tags   # 모델 목록이 나오면 정상
export OLLAMA_URL=http://192.168.x.x:11434   # 다른 컴퓨터에서 실행 중인 경우
```

> ⚠️ Ollama가 응답하지 않으면 `ollama serve` 명령으로 먼저 Ollama를 시작하세요.

> ⚠️ **대형 모델은 응답이 느립니다**: `qwen3.5:35b`, `deepseek-r1` 같은 큰 모델은 응답 생성까지 수 분이 걸릴 수 있습니다. 응답이 없는 것처럼 보여도 정상 처리 중일 수 있으니 기다려 주세요.

---

*더 자세한 API 정보는 [API.md](API.md)를 참고하세요.*
