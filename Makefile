.PHONY: build clean build-all package

# 项目名称
PROJECT_NAME := cbi
VERSION := $(shell git describe --tags --always --dirty 2>/dev/null | sed 's/^v//' || echo "0.1.0")
BUILD_DIR := ./bin

# 构建参数
LDFLAGS := -ldflags "-X github.com/CreatiBI/cli/cmd.Version=$(VERSION)"

# 默认目标
all: build

# 构建当前平台二进制文件
build:
	go build $(LDFLAGS) -o $(BUILD_DIR)/$(PROJECT_NAME) .

# 清理构建产物
clean:
	rm -rf $(BUILD_DIR)

# 跨平台编译
build-all: clean
	@echo "Building for all platforms..."

	# macOS amd64
	@mkdir -p $(BUILD_DIR)/darwin-amd64
	GOOS=darwin GOARCH=amd64 go build $(LDFLAGS) -o $(BUILD_DIR)/darwin-amd64/$(PROJECT_NAME) .
	@tar -czf $(BUILD_DIR)/darwin-amd64/$(PROJECT_NAME).tar.gz -C $(BUILD_DIR)/darwin-amd64 $(PROJECT_NAME)

	# macOS arm64 (Apple Silicon)
	@mkdir -p $(BUILD_DIR)/darwin-arm64
	GOOS=darwin GOARCH=arm64 go build $(LDFLAGS) -o $(BUILD_DIR)/darwin-arm64/$(PROJECT_NAME) .
	@tar -czf $(BUILD_DIR)/darwin-arm64/$(PROJECT_NAME).tar.gz -C $(BUILD_DIR)/darwin-arm64 $(PROJECT_NAME)

	# Linux amd64
	@mkdir -p $(BUILD_DIR)/linux-amd64
	GOOS=linux GOARCH=amd64 go build $(LDFLAGS) -o $(BUILD_DIR)/linux-amd64/$(PROJECT_NAME) .
	@tar -czf $(BUILD_DIR)/linux-amd64/$(PROJECT_NAME).tar.gz -C $(BUILD_DIR)/linux-amd64 $(PROJECT_NAME)

	# Linux arm64
	@mkdir -p $(BUILD_DIR)/linux-arm64
	GOOS=linux GOARCH=arm64 go build $(LDFLAGS) -o $(BUILD_DIR)/linux-arm64/$(PROJECT_NAME) .
	@tar -czf $(BUILD_DIR)/linux-arm64/$(PROJECT_NAME).tar.gz -C $(BUILD_DIR)/linux-arm64 $(PROJECT_NAME)

	# Windows amd64
	@mkdir -p $(BUILD_DIR)/windows-amd64
	GOOS=windows GOARCH=amd64 go build $(LDFLAGS) -o $(BUILD_DIR)/windows-amd64/$(PROJECT_NAME).exe .
	@cd $(BUILD_DIR)/windows-amd64 && zip -q $(PROJECT_NAME).zip $(PROJECT_NAME).exe

	# Windows arm64
	@mkdir -p $(BUILD_DIR)/windows-arm64
	GOOS=windows GOARCH=arm64 go build $(LDFLAGS) -o $(BUILD_DIR)/windows-arm64/$(PROJECT_NAME).exe .
	@cd $(BUILD_DIR)/windows-arm64 && zip -q $(PROJECT_NAME).zip $(PROJECT_NAME).exe

	@echo "✓ All platforms built successfully"
	@ls -la $(BUILD_DIR)/*

# 安装到系统
install: build
	cp $(BUILD_DIR)/$(PROJECT_NAME) /usr/local/bin/

# 初始化依赖
deps:
	go mod download
	go mod tidy

# 运行 CLI（开发模式）
run:
	go run .

# 格式化代码
fmt:
	go fmt ./...

# 检查代码
lint:
	go vet ./...

# 测试
test:
	go test -v ./...

# 发布到 npm
package: build-all
	npm pack