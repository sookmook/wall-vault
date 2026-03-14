# Contributing to wall-vault

Thank you for contributing to wall-vault!
*(Last updated: 2026-03-15)*

---

## How to Contribute

### Bug Reports

Open a new issue on [GitHub Issues](https://github.com/sookmook/wall-vault/issues).
Including the following helps resolve it faster:

- OS and version (e.g. Ubuntu 22.04 / macOS 14 / Windows 11 WSL2)
- wall-vault version (`wall-vault --version` or `/health` response)
- Steps to reproduce
- Expected behavior vs. actual behavior
- Relevant logs

### Feature Requests

Open an issue with the `enhancement` label.

### Code Contributions

1. Fork the repository
2. Create a feature branch: `git checkout -b feat/my-feature`
3. Make your changes
4. Add tests and make sure they pass: `make test`
5. Send a PR

---

## Development Setup

```bash
# Go 1.22+ required
go version

# Install dependencies
make deps

# Build
make build

# Test
make test
make test-verbose

# Local install
make install
```

---

## Project Structure

```
wall-vault/
├── cmd/              # subcommand entry points
├── internal/         # core logic (not exported)
│   ├── config/       # config load / save / plugins
│   ├── proxy/        # AI API proxy server
│   ├── vault/        # key vault server
│   ├── doctor/       # health check / auto-recovery
│   ├── i18n/         # multi-language support
│   ├── middleware/   # HTTP middleware
│   ├── models/       # model registry
│   ├── hooks/        # event hook system
│   └── theme/        # UI themes
├── configs/          # example configs, service plugins
├── docs/             # documentation
└── .github/          # GitHub Actions
```

---

## Coding Style

- **Go standard style** (`gofmt`, `go vet` must pass)
- Add Go doc comments to new exported functions
- Wrap errors: `fmt.Errorf("context: %w", err)`
- Tests in the same package as `_test.go` files
- Unit tests required for new features
- **Commit messages in English** — no Korean in commit messages
- **Source code comments in English**

---

## Writing Tests

```go
// Function name: Test{Feature}_{Scenario}
func TestHandleKeys_Unauthorized(t *testing.T) {
    // Arrange
    srv := newTestServer(t)

    // Act
    req := httptest.NewRequest("GET", "/api/keys", nil)
    w := httptest.NewRecorder()
    srv.Handler().ServeHTTP(w, req)

    // Assert
    if w.Code != http.StatusUnauthorized {
        t.Errorf("status = %d, want 401", w.Code)
    }
}
```

- Mock external services (Google, OpenRouter, Ollama) with `httptest.NewServer`
- Use `t.TempDir()` for file I/O
- Synchronize time-dependent tests with short `time.Sleep` or channels

---

## PR Checklist

- [ ] `go vet ./...` passes
- [ ] `go test ./...` all pass
- [ ] Tests added for new features
- [ ] New subcommands / flags: update `docs/API.md` or `docs/MANUAL.md`
- [ ] Major changes: add entry to `CHANGELOG.md` under `[Unreleased]`
- [ ] No personal API keys, tokens, or passwords in the commit
- [ ] `wall-vault.yaml` not accidentally included (it's in `.gitignore`)

---

## License

Contributions are released under the [GPL-3.0 License](LICENSE).

---
---

# wall-vault 기여 가이드

wall-vault에 기여해 주셔서 감사합니다!

---

## 기여 방법

### 버그 리포트

[GitHub Issues](https://github.com/sookmook/wall-vault/issues) 에서 새 이슈를 열어주세요.
다음 정보를 포함하면 빠른 해결에 도움이 됩니다:

- OS 및 버전 (예: Ubuntu 22.04 / macOS 14 / Windows 11 WSL2)
- wall-vault 버전 (`wall-vault --version` 또는 `/health` 응답)
- 재현 단계
- 예상 동작 vs 실제 동작
- 관련 로그

### 기능 제안

Issues에서 `enhancement` 레이블로 제안해 주세요.

### 코드 기여

1. 저장소를 fork합니다
2. feature 브랜치를 만듭니다: `git checkout -b feat/my-feature`
3. 변경 사항을 작성합니다
4. 테스트를 추가하고 통과시킵니다: `make test`
5. PR을 보냅니다

---

## 개발 환경

```bash
# Go 1.22 이상 필요
go version

# 의존성 설치
make deps

# 빌드
make build

# 테스트
make test
make test-verbose

# 로컬 설치
make install
```

---

## 프로젝트 구조

```
wall-vault/
├── cmd/              # 서브커맨드 진입점
├── internal/         # 핵심 로직 (패키지 공개 금지)
│   ├── config/       # 설정 로드·저장·플러그인
│   ├── proxy/        # AI API 프록시 서버
│   ├── vault/        # 키 금고 서버
│   ├── doctor/       # 헬스체크·자동복구
│   ├── i18n/         # 다국어 지원
│   ├── middleware/   # HTTP 미들웨어
│   ├── models/       # 모델 레지스트리
│   ├── hooks/        # 이벤트 훅 시스템
│   └── theme/        # UI 테마
├── configs/          # 예제 설정·서비스 플러그인
├── docs/             # 문서
└── .github/          # GitHub Actions
```

---

## 코딩 스타일

- **Go 표준 스타일** (`gofmt`, `go vet` 통과 필수)
- 새 패키지 공개 함수에는 Go doc 주석 추가
- 에러는 `fmt.Errorf("컨텍스트: %w", err)` 형식으로 래핑
- 테스트는 같은 패키지에 `_test.go` 파일로 작성
- 새 기능에는 단위 테스트 필수
- **커밋 메시지는 영어** — 한국어 커밋 메시지 금지
- **소스코드 주석은 영어**

---

## 테스트 작성 가이드

```go
// 테스트 함수명: Test{기능}_{시나리오}
func TestHandleKeys_Unauthorized(t *testing.T) {
    // Arrange
    srv := newTestServer(t)

    // Act
    req := httptest.NewRequest("GET", "/api/keys", nil)
    w := httptest.NewRecorder()
    srv.Handler().ServeHTTP(w, req)

    // Assert
    if w.Code != http.StatusUnauthorized {
        t.Errorf("status = %d, want 401", w.Code)
    }
}
```

- 외부 서비스(Google, OpenRouter, Ollama)는 `httptest.NewServer`로 mock
- 파일 I/O는 `t.TempDir()` 사용
- 시간 의존적 테스트는 짧은 `time.Sleep` 또는 채널로 동기화

---

## PR 체크리스트

- [ ] `go vet ./...` 통과
- [ ] `go test ./...` 전부 통과
- [ ] 새 기능에 테스트 추가
- [ ] 새 서브커맨드·플래그는 `docs/API.md` 또는 `docs/MANUAL.md` 업데이트
- [ ] 주요 변경은 `CHANGELOG.md` [Unreleased] 섹션에 기록
- [ ] 개인 API 키, 토큰, 비밀번호가 커밋에 포함되지 않았는지 확인
- [ ] `wall-vault.yaml` 파일이 실수로 포함되지 않았는지 확인 (`.gitignore`에 있음)

---

## 라이선스

기여하신 코드는 [GPL-3.0 License](LICENSE) 하에 배포됩니다.
