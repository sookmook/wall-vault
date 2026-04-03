BINARY = wall-vault
MODULE = github.com/sookmook/wall-vault
BASE_VERSION = v0.1.21
VERSION := $(BASE_VERSION).$(shell date +%Y%m%d.%H%M%S)

LDFLAGS = -ldflags "-X main.version=$(VERSION) -s -w"

# ─── 빌드 ────────────────────────────────────────────────────────────────────

.PHONY: build
build:
	~/go/bin/go build $(LDFLAGS) -o bin/$(BINARY) .

.PHONY: build-all
build-all: build-linux-amd64 build-linux-arm64 build-darwin-amd64 build-darwin-arm64 build-windows-amd64

build-linux-amd64:
	GOOS=linux GOARCH=amd64 ~/go/bin/go build $(LDFLAGS) -o bin/$(BINARY)-linux-amd64 .

build-linux-arm64:
	GOOS=linux GOARCH=arm64 ~/go/bin/go build $(LDFLAGS) -o bin/$(BINARY)-linux-arm64 .

build-darwin-amd64:
	GOOS=darwin GOARCH=amd64 ~/go/bin/go build $(LDFLAGS) -o bin/$(BINARY)-darwin-amd64 .

build-darwin-arm64:
	GOOS=darwin GOARCH=arm64 ~/go/bin/go build $(LDFLAGS) -o bin/$(BINARY)-darwin-arm64 .

build-windows-amd64:
	GOOS=windows GOARCH=amd64 ~/go/bin/go build $(LDFLAGS) -o bin/$(BINARY)-windows-amd64.exe .

# ─── 실행 ────────────────────────────────────────────────────────────────────

.PHONY: run
run:
	~/go/bin/go run . start

.PHONY: run-proxy
run-proxy:
	~/go/bin/go run . proxy

.PHONY: run-vault
run-vault:
	~/go/bin/go run . vault

# ─── 테스트 ──────────────────────────────────────────────────────────────────

.PHONY: test
test:
	~/go/bin/go test ./...

.PHONY: test-verbose
test-verbose:
	~/go/bin/go test -v ./...

# ─── 의존성 ──────────────────────────────────────────────────────────────────

.PHONY: deps
deps:
	~/go/bin/go mod tidy
	~/go/bin/go mod download

# ─── 정리 ────────────────────────────────────────────────────────────────────

.PHONY: clean
clean:
	rm -f bin/$(BINARY) bin/$(BINARY)-*

# ─── 설치 (로컬) ──────────────────────────────────────────────────────────────

.PHONY: install
install: build
	cp bin/$(BINARY) ~/.local/bin/$(BINARY)
	@echo "설치 완료: ~/.local/bin/$(BINARY)"

# ─── 개인 배포 설정 (Makefile.local 에서 로드) ───────────────────────────────
# Copy Makefile.local.example → Makefile.local and fill in your host details.
# Makefile.local is gitignored and never committed.
-include Makefile.local

# ─── 도움말 ──────────────────────────────────────────────────────────────────

.PHONY: help
help:
	@echo "wall-vault $(VERSION) build tool"
	@echo ""
	@echo "Commands:"
	@echo "  make build            build for current OS → bin/wall-vault"
	@echo "  make build-all        cross-compile all platforms"
	@echo "  make run              run wall-vault start"
	@echo "  make run-proxy        run proxy only"
	@echo "  make run-vault        run vault only"
	@echo "  make test             run tests"
	@echo "  make deps             tidy dependencies"
	@echo "  make install          install to ~/.local/bin"
	@echo "  make clean            remove build artifacts"
	@echo ""
	@echo "Deploy (requires Makefile.local — see Makefile.local.example):"
	@echo "  make deploy-mini      deploy to macOS machine"
	@echo "  make deploy-bot-c     deploy to Raspberry Pi"
	@echo "  make deploy-local     deploy locally (WSL)"
	@echo "  make deploy-all       deploy to all machines"
