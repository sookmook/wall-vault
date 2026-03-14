BINARY = wall-vault
MODULE = github.com/sookmook/wall-vault
BASE_VERSION = v0.1.6
VERSION = $(BASE_VERSION).$(shell date +%Y%m%d.%H%M%S)

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

# Wait for a process pattern to die (max 10s)
WAIT_DEAD = for i in $$(seq 1 10); do pgrep -f "$(1)" > /dev/null 2>&1 || break; sleep 1; done; \
            pgrep -f "$(1)" > /dev/null 2>&1 && echo "ERROR: process still alive after 10s" && exit 1 || true

.PHONY: deploy-mini
deploy-mini: build-darwin-arm64
	@echo "▶ [mini] uploading binary..."
	scp bin/$(BINARY)-darwin-arm64 $(MINI_HOST):~/.openclaw/$(BINARY).new
	@echo "▶ [mini] stopping services..."
	ssh $(MINI_HOST) '\
	  launchctl unload ~/Library/LaunchAgents/com.wall-vault.vault.plist 2>/dev/null || true; \
	  launchctl unload ~/Library/LaunchAgents/com.wall-vault.proxy.plist 2>/dev/null || true; \
	  pkill -f "wall-vault" 2>/dev/null || true; \
	  for i in $$(seq 1 10); do pgrep -f "wall-vault" > /dev/null 2>&1 || break; sleep 1; done; \
	  pgrep -f "wall-vault" > /dev/null 2>&1 && echo "ERROR: old process still alive" && exit 1 || true'
	@echo "▶ [mini] replacing binary..."
	ssh $(MINI_HOST) '\
	  cp ~/.openclaw/$(BINARY).new ~/.openclaw/$(BINARY) && \
	  codesign --sign - --force ~/.openclaw/$(BINARY) && \
	  rm -f ~/.openclaw/$(BINARY).new && \
	  echo "binary replaced: $$(ls -lh ~/.openclaw/$(BINARY) | awk '"'"'{print $$5, $$6, $$7, $$8}'"'"')"'
	@echo "▶ [mini] starting services..."
	ssh $(MINI_HOST) '\
	  launchctl load ~/Library/LaunchAgents/com.wall-vault.vault.plist && \
	  launchctl load ~/Library/LaunchAgents/com.wall-vault.proxy.plist'
	@echo "▶ [mini] verifying..."
	@sleep 3
	@ssh $(MINI_HOST) '\
	  VER=$$(curl -s --max-time 5 http://localhost:56243/api/status | python3 -c "import sys,json; print(json.load(sys.stdin)[\"version\"])" 2>/dev/null); \
	  if [ "$$VER" = "$(VERSION)" ]; then \
	    echo "✓ mini vault OK: $$VER"; \
	  else \
	    echo "✗ mini vault WRONG VERSION: $$VER (expected $(VERSION))"; exit 1; \
	  fi'

.PHONY: deploy-bot-c
deploy-bot-c: build-linux-arm64
	@echo "▶ [bot-c] uploading binary..."
	scp bin/$(BINARY)-linux-arm64 $(RASPI_HOST):~/.openclaw/$(BINARY).new
	@echo "▶ [bot-c] stopping service..."
	ssh $(RASPI_HOST) 'systemctl --user stop wall-vault-proxy 2>/dev/null || true; \
	  for i in $$(seq 1 8); do pgrep -f "wall-vault" > /dev/null 2>&1 || break; sleep 1; done'
	@echo "▶ [bot-c] replacing binary..."
	ssh $(RASPI_HOST) '\
	  cp ~/.openclaw/$(BINARY).new ~/.openclaw/$(BINARY) && \
	  rm -f ~/.openclaw/$(BINARY).new && \
	  echo "binary replaced: $$(ls -lh ~/.openclaw/$(BINARY) | awk '"'"'{print $$5, $$6, $$7, $$8}'"'"')"'
	@echo "▶ [bot-c] starting service..."
	ssh $(RASPI_HOST) 'systemctl --user start wall-vault-proxy'
	@echo "▶ [bot-c] verifying..."
	@sleep 3
	@ssh $(RASPI_HOST) '\
	  VER=$$(curl -s --max-time 5 http://localhost:56244/health | python3 -c "import sys,json; print(json.load(sys.stdin)[\"version\"])" 2>/dev/null); \
	  if [ "$$VER" = "$(VERSION)" ]; then \
	    echo "✓ bot-c proxy OK: $$VER"; \
	  else \
	    echo "✗ bot-c proxy WRONG VERSION: $$VER (expected $(VERSION))"; exit 1; \
	  fi'

.PHONY: deploy-local
deploy-local: build
	@echo "▶ [local] stopping proxy..."
	systemctl --user stop wall-vault-proxy 2>/dev/null || true
	@for i in $$(seq 1 8); do pgrep -f "wall-vault" > /dev/null 2>&1 || break; sleep 1; done
	@echo "▶ [local] replacing binary..."
	cp bin/$(BINARY) ~/.openclaw/$(BINARY)
	@echo "▶ [local] starting proxy..."
	systemctl --user start wall-vault-proxy
	@sleep 2
	@VER=$$(curl -s --max-time 5 http://localhost:56244/health | python3 -c "import sys,json; print(json.load(sys.stdin)['version'])" 2>/dev/null); \
	  if [ "$$VER" = "$(VERSION)" ]; then \
	    echo "✓ local proxy OK: $$VER"; \
	  else \
	    echo "✗ local proxy WRONG VERSION: $$VER (expected $(VERSION))"; exit 1; \
	  fi

.PHONY: deploy-all
deploy-all: deploy-bot-c deploy-mini deploy-local
	@echo ""
	@echo "✓ all machines deployed $(VERSION)"

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
	@echo "  make deploy-mini      미니(192.168.x.x) 배포 + 버전 검증"
	@echo "  make deploy-bot-c     라즈베리파이 배포 + 버전 검증"
	@echo "  make deploy-local     로컬(WSL) 배포 + 버전 검증"
	@echo "  make deploy-all       세 머신 전체 배포"
