## chlog — Makefile

APP     := chlog
MAIN_DIR := ./cmd/chlog
ROOT_OUT := ../..
GOEXE = $(shell go env GOEXE)
BINARY  := $(APP)$(GOEXE)

# Build metadata
BUILD_TIME := $(shell date +%Y-%m-%dT%H:%M:%S)
GIT_HASH  := $(shell git rev-parse --short=8 HEAD 2>/dev/null || echo "unknown")
VERSION ?= $(shell git describe --tags --always --dirty 2>/dev/null | sed 's/^v//' || echo "dev-$(GIT_HASH)")

LDFLAGS := -s -w \
	-X main.Version=$(VERSION) \
	-X main.GitHash=$(GIT_HASH) \
	-X 'main.BuildTime=$(BUILD_TIME)'

.PHONY: all build backend clean help latest

## all: build (default)
all: build

## build: build Go binary (current platform)
build:
	@echo "🐹 Building Go binary ($(VERSION) @ $(GIT_HASH))..."
	go -C $(MAIN_DIR) build -ldflags "$(LDFLAGS)" -o $(ROOT_OUT)/$(BINARY) .
	@echo "📦 Compressing binary..."
	@upx -6 --no-progress $(BINARY)
	@echo "✅ Binary: $(BINARY) ($$(du -sh $(BINARY) | cut -f1))"

## install: install Go binary to $GOPATH/bin
install:
	go -C $(MAIN_DIR) install -ldflags "$(LDFLAGS)" .
	upx -6 --no-progress $(GOPATH)/bin/$(BINARY)
	@echo "✅ Installed to GOPATH/bin"

## run: build and run with current directory
run: build
	./$(BINARY)

# ─── Cross Compilation ────────────────────────────────────────────────────────

DIST_DIR := dist

## build-all: cross-compile for all platforms
build-all: build-linux build-linux-arm64 build-darwin build-darwin-arm64 build-windows latest-yaml

## latest-yaml: generate latest.yaml release metadata
latest-yaml:
	@mkdir -p $(DIST_DIR)
	@{ \
		echo "name: $(APP)"; \
		echo "version: $(VERSION)"; \
		echo "released_at: $(BUILD_TIME)"; \
	} > $(DIST_DIR)/latest.yaml
	@echo "   → $(DIST_DIR)/latest.yaml"

## build-linux: compile for Linux amd64
build-linux:
	@echo "🐧 linux/amd64..."
	@mkdir -p $(DIST_DIR)
	@GOOS=linux GOARCH=amd64 go -C $(MAIN_DIR) build -ldflags "$(LDFLAGS)" -o $(ROOT_OUT)/$(DIST_DIR)/$(APP)-linux-amd64 .
	upx -6 --no-progress $(DIST_DIR)/$(APP)-linux-amd64
	chmod +x $(DIST_DIR)/$(APP)-linux-amd64
	@echo "   → $(DIST_DIR)/$(APP)-linux-amd64"

## build-linux-arm64: compile for Linux arm64
build-linux-arm64:
	@echo "🐧 linux/arm64..."
	@mkdir -p $(DIST_DIR)
	@GOOS=linux GOARCH=arm64 go -C $(MAIN_DIR) build -ldflags "$(LDFLAGS)" -o $(ROOT_OUT)/$(DIST_DIR)/$(APP)-linux-arm64 .
	upx -6 --no-progress $(DIST_DIR)/$(APP)-linux-arm64
	chmod +x $(DIST_DIR)/$(APP)-linux-arm64
	@echo "   → $(DIST_DIR)/$(APP)-linux-arm64"

## build-darwin: compile for macOS amd64
build-darwin:
	@echo "🍎 darwin/amd64..."
	@mkdir -p $(DIST_DIR)
	@GOOS=darwin GOARCH=amd64 go -C $(MAIN_DIR) build -ldflags "$(LDFLAGS)" -o $(ROOT_OUT)/$(DIST_DIR)/$(APP)-darwin-amd64 .
	@echo "   → $(DIST_DIR)/$(APP)-darwin-amd64"

## build-darwin-arm64: compile for macOS Apple Silicon
build-darwin-arm64:
	@echo "🍎 darwin/arm64..."
	@mkdir -p $(DIST_DIR)
	@GOOS=darwin GOARCH=arm64 go -C $(MAIN_DIR) build -ldflags "$(LDFLAGS)" -o $(ROOT_OUT)/$(DIST_DIR)/$(APP)-darwin-arm64 .
	# upx -6 --no-progress $(DIST_DIR)/$(APP)-darwin-arm64 # 压缩有问题在 macos 12+
	@echo "   → $(DIST_DIR)/$(APP)-darwin-arm64"

## build-windows: compile for Windows amd64
build-windows:
	@echo "🪟 windows/amd64..."
	@mkdir -p $(DIST_DIR)
	@GOOS=windows GOARCH=amd64 go -C $(MAIN_DIR) build -ldflags "$(LDFLAGS)" -o $(ROOT_OUT)/$(DIST_DIR)/$(APP)-windows-amd64.exe .
	upx -6 --no-progress $(DIST_DIR)/$(APP)-windows-amd64.exe
	@echo "   → $(DIST_DIR)/$(APP)-windows-amd64.exe"

.PHONY: release
release: build-all ## Create release archives for all platforms TODO 还未启用的
	@echo "Creating release archives..."
	@mkdir -p release
	@cd $(DIST_DIR) && \
	tar -czf ../release/$(APP)-linux-amd64.tar.gz $(APP)-linux-amd64; \
	tar -czf ../release/$(APP)-linux-arm64.tar.gz $(APP)-linux-arm64; \
	tar -czf ../release/$(APP)-darwin-amd64.tar.gz $(APP)-darwin-amd64; \
	tar -czf ../release/$(APP)-darwin-arm64.tar.gz $(APP)-darwin-arm64; \
	zip -czf ../release/$(APP)-windows-amd64.zip $(APP)-windows-amd64.exe; \
	# zip ../release/$(APP)-windows-arm64.zip $(APP)-windows-arm64.exe; \
	@echo "Release archives created in release/"

## clean: remove build artifacts
clean:
	@rm -f $(BINARY)
	@rm -rf $(DIST_DIR)
	@echo "🧹 Cleaned"

## help: show this help
help:
	@echo "Skillc Build System"
	@echo ""
	@grep -E '^## ' $(MAKEFILE_LIST) | sed 's/## /  /'
