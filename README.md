# wall-vault

> AI 프록시 + 키 금고 통합 시스템

단독 봇부터 멀티 봇 분산 구성까지, 하나의 바이너리로 AI API 프록시와 키 관리를 해결합니다.

[![License: MIT](https://img.shields.io/badge/License-MIT-blue.svg)](LICENSE)
[![Go Version](https://img.shields.io/badge/Go-1.22+-00ADD8.svg)](https://go.dev)

---

## 빠른 시작 (5분 설치)

```bash
# 1. 다운로드 (Linux amd64)
curl -L https://github.com/sookmook/wall-vault/releases/latest/download/wall-vault-linux-amd64 \
  -o wall-vault && chmod +x wall-vault

# 2. 대화형 설정 마법사 (처음 한 번)
./wall-vault setup

# 3. 실행
./wall-vault start
```

브라우저에서 `http://localhost:56243` 열면 키 금고 대시보드가 열립니다.

---

## 기능

| 기능 | 설명 |
|------|------|
| **프록시** | Google Gemini / OpenRouter / Ollama API 프록시 |
| **키 금고** | API 키 관리, 사용량 모니터링, 라운드 로빈 |
| **AES-GCM 암호화** | 마스터 비밀번호로 API 키 암호화 저장 |
| **SSE 동기화** | 키 금고 ↔ 프록시 실시간 설정 동기화 |
| **도구 보안 필터** | 외부 function calling 차단 (strip_all / whitelist / passthrough) |
| **폴백 체인** | Google → OpenRouter → Ollama 자동 전환 |
| **Ollama 자동 조회** | 로컬 Ollama 서버 모델 자동 검색 |
| **주치의** | 서비스 헬스체크, 자동복구, systemd/launchd 등록 |
| **서비스 플러그인** | YAML 파일로 코드 수정 없이 새 서비스 추가 |
| **다국어** | 한국어 / English / 日本語 |
| **테마** | 벚꽃 🌸 / 다크 / 라이트 / 오션 |

---

## 사용법

```bash
wall-vault setup             # 대화형 설치 마법사 (처음 시작)
wall-vault start             # 모든 서비스 시작 (proxy + vault)
wall-vault proxy             # 프록시만 실행
wall-vault vault             # 키 금고만 실행
wall-vault doctor check      # 상태 확인
wall-vault doctor status     # 상세 보고서
wall-vault doctor fix        # 자동 복구
wall-vault doctor deploy     # systemd 서비스 파일 생성
wall-vault doctor deploy launchd  # macOS launchd plist 생성
```

---

## 아키텍처

```
                  ┌────────────────────────┐
                  │  키 금고 (:56243)       │
                  │  API 키 암호화 저장      │
                  │  SSE 브로드캐스트        │
                  └──────────┬─────────────┘
                             │ SSE 실시간 동기화
           ┌─────────────────┼─────────────────┐
           ▼                 ▼                 ▼
      봇A (:56244)      봇B (:56244)      봇C (:56244)
       (프록시)          (프록시)          (프록시)
           │                 │                 │
           └─────────────────┴─────────────────┘
                             │ 폴백 체인
                  ┌──────────┼──────────┐
                  ▼          ▼          ▼
               Google   OpenRouter   Ollama
```

### 폴백 체인

```
1단계: 지정 서비스 (설정 기준)
2단계: 나머지 서비스 순서대로
3단계: Ollama (최종 폴백)
```

---

## 설정

```bash
# 대화형 마법사 (권장)
./wall-vault setup

# 또는 예제 설정 복사
cp configs/example-standalone.yaml wall-vault.yaml
cp configs/example-distributed.yaml wall-vault.yaml
```

### 주요 설정 (`wall-vault.yaml`)

```yaml
mode: standalone   # standalone | distributed
lang: ko           # ko | en | ja
theme: sakura      # sakura | dark | light | ocean

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
  admin_token: ""            # 빈칸이면 인증 없음
  master_password: ""        # API 키 암호화 비밀번호
  data_dir: ~/.wall-vault/data
  services_dir: configs/services  # 서비스 플러그인 YAML 폴더
```

### 환경변수

| 변수 | 설명 |
|------|------|
| `WV_LANG` | 언어 (ko/en/ja) |
| `WV_THEME` | 테마 |
| `WV_PROXY_PORT` | 프록시 포트 오버라이드 |
| `WV_VAULT_PORT` | 금고 포트 오버라이드 |
| `WV_VAULT_URL` | 키 금고 URL (분산 모드) |
| `WV_VAULT_TOKEN` | 프록시 인증 토큰 |
| `WV_ADMIN_TOKEN` | 관리자 토큰 |
| `WV_MASTER_PASS` | 암호화 마스터 비밀번호 |
| `OLLAMA_URL` | Ollama 서버 URL |
| `WV_KEY_GOOGLE` | Google API 키 (쉼표 구분) |
| `WV_KEY_OPENROUTER` | OpenRouter API 키 |

---

## 서비스 플러그인

코드 수정 없이 YAML 파일로 새 AI 서비스 추가:

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
model_fetch:
  enabled: false
error_codes:
  429:
    cooldown: 30m
  401:
    cooldown: 24h
usage_threshold: 97
```

---

## API 엔드포인트

### 프록시 (`:56244`)

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

### 키 금고 (`:56243`)

| 경로 | 설명 |
|------|------|
| `GET /` | 대시보드 UI |
| `GET /api/status` | 상태 조회 |
| `GET /api/events` | SSE 스트림 |
| `GET /api/clients` | 클라이언트 목록 (공개) |
| `GET /api/keys` | 복호화된 키 목록 (클라이언트 인증) |
| `POST /api/heartbeat` | 프록시 상태 전송 |
| `GET /admin/keys` | 키 목록 (관리자) |
| `POST /admin/keys` | 키 추가 |
| `DELETE /admin/keys/{id}` | 키 삭제 |
| `POST /admin/keys/reset` | 일일 사용량 초기화 |
| `GET /admin/clients` | 클라이언트 목록 (관리자) |
| `POST /admin/clients` | 클라이언트 추가 |
| `PUT /admin/clients/{id}` | 클라이언트 수정 |
| `DELETE /admin/clients/{id}` | 클라이언트 삭제 |

---

## 모드

### Standalone (단독 봇)

```
[wall-vault] 프록시(:56244) + 금고(:56243) 한 기기에서 실행
```

```bash
./wall-vault start
# 환경변수로 키 주입
WV_KEY_GOOGLE=AIza... ./wall-vault start
```

### Distributed (멀티 봇)

```
[키 금고 서버 :56243]
    ├── SSE → [봇 A :56244]
    ├── SSE → [봇 B :56244]
    └── SSE → [봇 C :56244]
```

```bash
# 키 금고 서버 (미니)
./wall-vault vault

# 프록시 각 봇에서
WV_VAULT_URL=http://192.168.0.6:56243 \
WV_VAULT_TOKEN=my-bot-token \
./wall-vault proxy
```

---

## 주치의 (Doctor)

```bash
# 상태 확인
wall-vault doctor check

# 상세 보고서
wall-vault doctor status

# 자동 복구 (systemd → launchd → 직접 시작 순서)
wall-vault doctor fix

# systemd 서비스 등록
wall-vault doctor deploy
systemctl --user daemon-reload
systemctl --user enable --now wall-vault

# macOS launchd 등록
wall-vault doctor deploy launchd
launchctl load ~/Library/LaunchAgents/com.wall-vault.plist
```

---

## 빌드

```bash
# 현재 OS용
make build

# 전체 플랫폼 크로스컴파일
make build-all
# → bin/wall-vault-linux-amd64  (6.3MB)
# → bin/wall-vault-linux-arm64  (6.1MB)
# → bin/wall-vault-darwin-amd64 (6.5MB)
# → bin/wall-vault-darwin-arm64 (6.2MB)

# 테스트
make test

# 로컬 설치
make install  # → ~/.local/bin/wall-vault
```

---

## 프로젝트 구조

```
wall-vault/
├── main.go                      # 진입점 + 서브커맨드 라우터
├── cmd/
│   ├── proxy/proxy.go           # proxy 서브커맨드
│   ├── vault/vault.go           # vault 서브커맨드
│   ├── setup/setup.go           # setup 마법사
│   └── doctor/doctor.go         # doctor 서브커맨드
├── internal/
│   ├── config/
│   │   ├── config.go            # 설정 로드·저장
│   │   └── services.go          # 서비스 플러그인 로더
│   ├── proxy/
│   │   ├── server.go            # 프록시 HTTP 서버 + 폴백 체인
│   │   ├── stream.go            # 스트리밍 핸들러 (SSE)
│   │   ├── keymgr.go            # 키 관리자 (라운드 로빈·쿨다운)
│   │   ├── heartbeat.go         # 금고 Heartbeat 전송
│   │   ├── sseconn.go           # SSE 클라이언트 (금고 연결)
│   │   ├── convert.go           # Gemini↔OpenAI↔Ollama 변환
│   │   ├── toolfilter.go        # 도구 보안 필터
│   │   └── models.go            # 요청/응답 타입 정의
│   ├── vault/
│   │   ├── server.go            # 키 금고 HTTP 서버
│   │   ├── store.go             # 스레드 안전 저장소 (JSON)
│   │   ├── crypto.go            # AES-GCM 암호화
│   │   ├── broker.go            # SSE 브로드캐스터
│   │   ├── ui.go                # 대시보드 HTML 생성
│   │   └── models.go            # APIKey·Client·ProxyStatus 정의
│   ├── doctor/doctor.go         # 헬스체크·자동복구·서비스 파일 생성
│   ├── middleware/middleware.go  # Logger·CORS·Recovery·Chain
│   ├── models/registry.go       # 모델 레지스트리 (Google/OpenRouter/Ollama)
│   ├── theme/theme.go           # UI 테마 (sakura/dark/light/ocean)
│   ├── i18n/i18n.go             # 다국어 (ko/en/ja)
│   └── hooks/hooks.go           # OpenClaw 연동 훅
├── configs/
│   ├── services/                # 서비스 플러그인 YAML
│   │   ├── google.yaml
│   │   ├── openrouter.yaml
│   │   └── ollama.yaml
│   ├── example-standalone.yaml
│   └── example-distributed.yaml
└── bin/                         # 빌드 결과물
```

---

## 라이선스

MIT License — 자유롭게 사용, 수정, 배포 가능합니다.
