BINARY = wall-vault
MODULE = github.com/sookmook/wall-vault
VERSION = v0.1.4

LDFLAGS = -ldflags "-X main.version=$(VERSION) -s -w"

# ─── 빌드 ────────────────────────────────────────────────────────────────────

.PHONY: build
build:
	go build $(LDFLAGS) -o bin/$(BINARY) .

.PHONY: build-all
build-all: build-linux-amd64 build-linux-arm64 build-darwin-amd64 build-darwin-arm64 build-windows-amd64

build-linux-amd64:
	GOOS=linux GOARCH=amd64 go build $(LDFLAGS) -o bin/$(BINARY)-linux-amd64 .

build-linux-arm64:
	GOOS=linux GOARCH=arm64 go build $(LDFLAGS) -o bin/$(BINARY)-linux-arm64 .

build-darwin-amd64:
	GOOS=darwin GOARCH=amd64 go build $(LDFLAGS) -o bin/$(BINARY)-darwin-amd64 .

build-darwin-arm64:
	GOOS=darwin GOARCH=arm64 go build $(LDFLAGS) -o bin/$(BINARY)-darwin-arm64 .

build-windows-amd64:
	GOOS=windows GOARCH=amd64 go build $(LDFLAGS) -o bin/$(BINARY)-windows-amd64.exe .

# ─── 실행 ────────────────────────────────────────────────────────────────────

.PHONY: run
run:
	go run . start

.PHONY: run-proxy
run-proxy:
	go run . proxy

.PHONY: run-vault
run-vault:
	go run . vault

# ─── 테스트 ──────────────────────────────────────────────────────────────────

.PHONY: test
test:
	go test ./...

.PHONY: test-verbose
test-verbose:
	go test -v ./...

# ─── 의존성 ──────────────────────────────────────────────────────────────────

.PHONY: deps
deps:
	go mod tidy
	go mod download

# ─── 정리 ────────────────────────────────────────────────────────────────────

.PHONY: clean
clean:
	rm -f bin/$(BINARY) bin/$(BINARY)-*

# ─── 설치 (로컬) ──────────────────────────────────────────────────────────────

.PHONY: install
install: build
	cp bin/$(BINARY) ~/.local/bin/$(BINARY)
	@echo "설치 완료: ~/.local/bin/$(BINARY)"

# ─── 배포 ────────────────────────────────────────────────────────────────────

MINI_HOST = 192.168.x.x
RASPI_HOST = bot-c

.PHONY: deploy-mini
deploy-mini: build-darwin-arm64
	scp bin/$(BINARY)-darwin-arm64 $(MINI_HOST):~/.openclaw/$(BINARY).new
	ssh $(MINI_HOST) 'pkill -f "wall-vault vault" 2>/dev/null || true; sleep 1; \
	  cp ~/.openclaw/$(BINARY).new ~/.openclaw/$(BINARY); \
	  codesign --sign - --force ~/.openclaw/$(BINARY); \
	  launchctl unload ~/Library/LaunchAgents/com.wall-vault.proxy.plist 2>/dev/null || true; \
	  launchctl load ~/Library/LaunchAgents/com.wall-vault.proxy.plist; \
	  WV_ADMIN_TOKEN=dhvmszmffh WV_MASTER_PASS=dhvmszmffh WV_VAULT_PORT=56243 \
	    nohup ~/.openclaw/$(BINARY) vault >> /tmp/wall-vault.err 2>&1 &'
	@echo "미니 배포 완료 (코드 서명 포함)"

.PHONY: deploy-bot-c
deploy-bot-c: build-linux-arm64
	ssh $(RASPI_HOST) 'systemctl --user stop wall-vault-proxy 2>/dev/null || true; sleep 1'
	scp bin/$(BINARY)-linux-arm64 $(RASPI_HOST):~/.openclaw/$(BINARY)
	ssh $(RASPI_HOST) 'systemctl --user start wall-vault-proxy'
	@echo "라즈 배포 완료"

.PHONY: deploy-all
deploy-all: build deploy-bot-c deploy-mini
	systemctl --user stop wall-vault-proxy; \
	  cp bin/$(BINARY) ~/.openclaw/$(BINARY); \
	  systemctl --user start wall-vault-proxy
	@echo "전체 배포 완료"

# ─── 도움말 ──────────────────────────────────────────────────────────────────

.PHONY: help
help:
	@echo "wall-vault $(VERSION) 빌드 도구"
	@echo ""
	@echo "명령:"
	@echo "  make build            현재 OS용 빌드 → bin/wall-vault"
	@echo "  make build-all        전체 플랫폼 크로스컴파일"
	@echo "  make run              wall-vault start 실행"
	@echo "  make run-proxy        프록시만 실행"
	@echo "  make run-vault        키 금고만 실행"
	@echo "  make test             테스트 실행"
	@echo "  make deps             의존성 정리"
	@echo "  make install          ~/.local/bin 에 설치"
	@echo "  make clean            빌드 결과물 삭제"
