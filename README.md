# wall-vault

> AI 프록시 + 키 금고 통합 시스템

단독 봇부터 멀티 봇 분산 구성까지, 하나의 Go 바이너리로 AI API 프록시와 키 관리를 해결합니다.

[![License: MIT](https://img.shields.io/badge/License-MIT-blue.svg)](LICENSE)
[![Go Version](https://img.shields.io/badge/Go-1.22+-00ADD8.svg)](https://go.dev)
[![CI](https://github.com/sookmook/wall-vault/actions/workflows/ci.yml/badge.svg)](https://github.com/sookmook/wall-vault/actions/workflows/ci.yml)
[![Languages](https://img.shields.io/badge/languages-10-brightgreen.svg)](#다국어)
[![Platform](https://img.shields.io/badge/platform-Linux%20%7C%20macOS%20%7C%20Windows-lightgrey.svg)](#설치)

---

## 빠른 시작

### Linux / macOS

```bash
# 1. 다운로드
curl -L https://github.com/sookmook/wall-vault/releases/latest/download/wall-vault-linux-amd64 \
  -o wall-vault && chmod +x wall-vault

# macOS Apple Silicon
curl -L https://github.com/sookmook/wall-vault/releases/latest/download/wall-vault-darwin-arm64 \
  -o wall-vault && chmod +x wall-vault

# 2. 대화형 설치 마법사
./wall-vault setup

# 3. 실행
./wall-vault start
```

### Windows

```powershell
# PowerShell — 다운로드
Invoke-WebRequest -Uri `
  "https://github.com/sookmook/wall-vault/releases/latest/download/wall-vault-windows-amd64.exe" `
  -OutFile wall-vault.exe

# 실행
.\wall-vault.exe setup
.\wall-vault.exe start
```

### WSL (Windows Subsystem for Linux)

```bash
# WSL에서 Linux 바이너리 사용 (권장)
curl -L https://github.com/sookmook/wall-vault/releases/latest/download/wall-vault-linux-amd64 \
  -o wall-vault && chmod +x wall-vault
./wall-vault setup
```

브라우저에서 `http://localhost:56243` 을 열면 키 금고 대시보드가 나타납니다.

---

## 기능

| 기능 | 설명 |
|------|------|
| **프록시** | Google Gemini / OpenRouter / Ollama API 프록시 |
| **키 금고** | API 키 관리, 사용량 모니터링, 라운드 로빈 |
| **AES-GCM 암호화** | 마스터 비밀번호로 API 키 암호화 저장 |
| **SSE 동기화** | 키 금고 ↔ 프록시 실시간 설정 동기화 (1–3초) |
| **도구 보안 필터** | function calling 차단 (strip_all / whitelist / passthrough) |
| **폴백 체인** | Google → OpenRouter → Ollama 자동 전환 |
| **모델 검색** | 전체 모델 ID·이름 검색 (346개+ OpenRouter 지원) |
| **주치의** | 헬스체크, 자동복구, systemd/launchd/NSSM 등록 |
| **서비스 플러그인** | YAML 파일로 코드 수정 없이 새 서비스 추가 |
| **다국어** | 세계 10대 언어 지원 |
| **테마** | 벚꽃 🌸 / 다크 / 라이트 / 오션 |
| **크로스 플랫폼** | Linux / macOS / Windows / WSL |

---

## 다국어

setup 마법사와 시스템 메시지가 다음 언어로 표시됩니다:

| 코드 | 언어 | 코드 | 언어 |
|------|------|------|------|
| `ko` | 한국어 | `ar` | العربية |
| `en` | English | `pt` | Português |
| `zh` | 中文 | `fr` | Français |
| `es` | Español | `de` | Deutsch |
| `hi` | हिन्दी | `ja` | 日本語 |

언어 자동 감지 (`LANG` 환경변수) 또는 수동 지정:

```bash
WV_LANG=ja ./wall-vault setup
```

---

## 사용법

```bash
wall-vault setup                    # 대화형 설치 마법사 (처음 시작)
wall-vault start                    # 프록시 + 키 금고 동시 실행
wall-vault proxy                    # 프록시만 실행
wall-vault proxy --key-google=AIza... # API 키 직접 전달
wall-vault vault                    # 키 금고만 실행
wall-vault doctor check             # 서비스 상태 확인
wall-vault doctor status            # 상세 보고서
wall-vault doctor fix               # 자동 복구
wall-vault doctor deploy            # systemd 서비스 파일 생성 (Linux)
wall-vault doctor deploy launchd    # launchd plist 생성 (macOS)
wall-vault doctor deploy windows    # NSSM 서비스 스크립트 생성 (Windows)
```

### proxy 플래그

| 플래그 | 환경변수 | 설명 |
|--------|----------|------|
| `--key-google` | `WV_KEY_GOOGLE` | Google API 키 |
| `--key-openrouter` | `WV_KEY_OPENROUTER` | OpenRouter API 키 |
| `--vault` | `WV_VAULT_URL` | 키 금고 URL |
| `--vault-token` | `WV_VAULT_TOKEN` | 금고 인증 토큰 |
| `--filter` | — | 도구 필터 (strip_all/whitelist/passthrough) |
| `--port` | `WV_PROXY_PORT` | 프록시 포트 |
| `--id` | `VAULT_CLIENT_ID` | 클라이언트 ID |

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
```

### 주요 설정 (`wall-vault.yaml`)

```yaml
mode: standalone   # standalone | distributed
lang: ko           # ko | en | zh | es | hi | ar | pt | fr | de | ja
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
| `WV_LANG` | 언어 (ko/en/zh/es/hi/ar/pt/fr/de/ja) |
| `WV_THEME` | 테마 |
| `WV_PROXY_PORT` | 프록시 포트 오버라이드 |
| `WV_VAULT_PORT` | 금고 포트 오버라이드 |
| `WV_VAULT_URL` | 키 금고 URL (분산 모드) |
| `WV_VAULT_TOKEN` | 프록시 인증 토큰 |
| `WV_ADMIN_TOKEN` | 관리자 토큰 |
| `WV_MASTER_PASS` | 암호화 마스터 비밀번호 |
| `WV_KEY_GOOGLE` | Google API 키 (쉼표 구분) |
| `WV_KEY_OPENROUTER` | OpenRouter API 키 |

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
2. 서비스 스크립트 생성:
   ```powershell
   .\wall-vault.exe doctor deploy windows
   ```
3. 생성된 스크립트를 관리자 권한으로 실행:
   ```powershell
   # 관리자 PowerShell
   %USERPROFILE%\install-wall-vault-service.bat
   ```
4. 서비스 관리:
   ```powershell
   nssm start wall-vault
   nssm stop wall-vault
   nssm restart wall-vault
   ```

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
| `GET /api/models` | 모델 목록 (346개+) |
| `GET /api/models?q=gemini` | 모델 검색 |
| `PUT /api/config/model` | 모델 변경 |
| `POST /reload` | 설정 새로고침 |

### 키 금고 (`:56243`)

| 경로 | 설명 |
|------|------|
| `GET /` | 대시보드 UI |
| `GET /api/status` | 상태 조회 |
| `GET /api/events` | SSE 스트림 |
| `GET /api/keys` | 복호화된 키 목록 (클라이언트 인증) |
| `POST /admin/keys` | 키 추가 |
| `DELETE /admin/keys/{id}` | 키 삭제 |
| `POST /admin/clients` | 클라이언트 추가 |
| `PUT /admin/clients/{id}` | 클라이언트 수정 |
| `DELETE /admin/clients/{id}` | 클라이언트 삭제 |

---

## 모드

### Standalone (단독 봇)

```bash
WV_KEY_GOOGLE=AIza... ./wall-vault start
# 또는 플래그로
./wall-vault proxy --key-google=AIza...
```

### Distributed (멀티 봇)

```bash
# 키 금고 서버
./wall-vault vault

# 각 봇에서
WV_VAULT_URL=http://192.168.0.6:56243 \
WV_VAULT_TOKEN=my-bot-token \
./wall-vault proxy
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

# 테스트
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
│   │   ├── config.go            # 설정 로드·저장 (Windows APPDATA 지원)
│   │   └── services.go          # 서비스 플러그인 로더
│   ├── proxy/
│   │   ├── server.go            # 프록시 HTTP 서버 + 폴백 체인
│   │   ├── keymgr.go            # 라운드 로빈 키 관리자
│   │   ├── convert.go           # Gemini↔OpenAI↔Ollama 변환
│   │   └── toolfilter.go        # 도구 보안 필터
│   ├── vault/
│   │   ├── server.go            # 키 금고 HTTP 서버
│   │   ├── store.go             # AES-GCM 암호화 저장소
│   │   ├── broker.go            # SSE 브로드캐스터
│   │   └── ui.go                # 대시보드 HTML (테마 지원)
│   ├── doctor/doctor.go         # 자동복구 (systemd/launchd/NSSM)
│   ├── models/registry.go       # 모델 레지스트리 + 검색
│   ├── i18n/i18n.go             # 다국어 (10개 언어)
│   └── hooks/hooks.go           # 이벤트 훅 시스템
├── configs/
│   ├── services/                # 서비스 플러그인 YAML
│   ├── example-standalone.yaml
│   └── example-distributed.yaml
└── .github/workflows/
    ├── ci.yml                   # 테스트 + 5플랫폼 빌드
    └── release.yml              # 태그 → GitHub Release 자동 생성
```

---

## 라이선스

MIT License — 자유롭게 사용, 수정, 배포 가능합니다.
