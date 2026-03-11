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

---

## 라이선스

기여하신 코드는 [MIT License](LICENSE) 하에 배포됩니다.
