# wall-vault

> AI 프록시 + 키 금고 통합 시스템

단독 봇부터 멀티 봇 분산 구성까지, 하나의 바이너리로 AI API 프록시와 키 관리를 해결합니다.

[![License: MIT](https://img.shields.io/badge/License-MIT-blue.svg)](LICENSE)
[![Go Version](https://img.shields.io/badge/Go-1.22+-00ADD8.svg)](https://go.dev)

---

## 빠른 시작 (5분 설치)

```bash
# 1. 다운로드 (Linux)
curl -L https://github.com/sookmook/wall-vault/releases/latest/download/wall-vault-linux-amd64 \
  -o wall-vault && chmod +x wall-vault

# 2. 실행
./wall-vault start
```

브라우저에서 `http://localhost:56243` 열면 키 금고 대시보드가 열립니다.

---

## 기능

| 기능 | 설명 |
|------|------|
| **프록시** | Google Gemini / OpenRouter / Ollama API 프록시 |
| **키 금고** | API 키 관리, 사용량 모니터링, 라운드 로빈 |
| **SSE 동기화** | 키 금고 ↔ 프록시 실시간 설정 동기화 |
| **도구 보안 필터** | 외부 function calling 차단 (strip_all / whitelist) |
| **Ollama 자동 조회** | 로컬 Ollama 서버 모델 자동 검색 |
| **주치의** | 서비스 헬스체크 및 자동 복구 |
| **다국어** | 한국어 / English / 日本語 |
| **테마** | 벚꽃 🌸 / 다크 / 라이트 / 오션 |

---

## 사용법

```bash
wall-vault start             # 모든 서비스 시작 (초보자용)
wall-vault proxy             # 프록시만 실행
wall-vault vault             # 키 금고만 실행
wall-vault doctor status     # 상태 확인
wall-vault doctor fix        # 자동 복구
```

---

## 설정

```bash
# 초보자용 설정 파일 복사
cp configs/example-standalone.yaml wall-vault.yaml

# 멀티 봇 구성
cp configs/example-distributed.yaml wall-vault.yaml
```

### 주요 설정 항목 (`wall-vault.yaml`)

```yaml
mode: standalone   # standalone | distributed
lang: ko           # ko | en | ja
theme: sakura      # sakura | dark | light | ocean

proxy:
  port: 56244
  client_id: my-bot
  tool_filter: strip_all
  services: [google, openrouter, ollama]

vault:
  port: 56243
```

### 환경변수

| 변수 | 설명 |
|------|------|
| `WV_LANG` | 언어 (ko/en/ja) |
| `WV_THEME` | 테마 |
| `WV_VAULT_URL` | 키 금고 URL (분산 모드) |
| `WV_VAULT_TOKEN` | 프록시 인증 토큰 |
| `WV_ADMIN_TOKEN` | 관리자 토큰 |
| `OLLAMA_URL` | Ollama 서버 URL |

---

## 서비스 플러그인

새 AI 서비스는 코드 수정 없이 YAML로 추가:

```yaml
# configs/services/my-service.yaml
id: my-service
name: My AI Service
enabled: true
endpoints:
  generate: https://api.example.com/v1/chat
auth:
  type: bearer
request_format: openai
```

---

## 모드

### Standalone (단독 봇)
```
[wall-vault] 프록시(:56244) + 금고(:56243) 한 기기에서 실행
```

### Distributed (멀티 봇)
```
[키 금고 서버 :56243]
    ├── SSE → [봇 A :56244]
    ├── SSE → [봇 B :56244]
    └── SSE → [봇 C :56244]
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
```

---

## 프로젝트 구조

```
wall-vault/
├── main.go                    # 진입점 + 서브커맨드 라우터
├── cmd/
│   ├── proxy/                 # 프록시 서버
│   ├── vault/                 # 키 금고 서버
│   └── doctor/                # 헬스체크
├── internal/
│   ├── config/                # 설정 로드·저장
│   ├── models/                # 모델 레지스트리 (자동 조회)
│   ├── theme/                 # UI 테마
│   ├── i18n/                  # 다국어
│   └── hooks/                 # OpenClaw 연동 훅
├── configs/
│   ├── services/              # 서비스 플러그인 YAML
│   ├── example-standalone.yaml
│   └── example-distributed.yaml
└── bin/                       # 빌드 결과물
```

---

## 라이선스

MIT License — 자유롭게 사용, 수정, 배포 가능합니다.
