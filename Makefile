# Heimdall API Makefile

.PHONY: help build test clean admin public deps fmt lint docker

# 默认目标
help:
	@echo "Available commands:"
	@echo "  build    - 构建所有服务"
	@echo "  admin    - 启动管理服务 (端口: 8080)"
	@echo "  public   - 启动公开服务 (端口: 8081)"
	@echo "  test     - 运行所有测试"
	@echo "  deps     - 整理依赖"
	@echo "  fmt      - 格式化代码"
	@echo "  lint     - 代码检查"
	@echo "  clean    - 清理构建文件"
	@echo "  docker   - 构建Docker镜像"

# 构建所有服务
build:
	@echo "构建所有服务..."
	go build -o bin/admin-api ./admin-api/admin
	go build -o bin/public-api ./public-api/public
	@echo "构建完成"

# 启动管理服务
admin:
	@echo "启动管理服务 (端口: 8080)..."
	go run ./admin-api/admin

# 启动公开服务
public:
	@echo "启动公开服务 (端口: 8081)..."
	go run ./public-api/public

# 运行测试
test:
	@echo "运行所有测试..."
	go test ./... -v

# 整理依赖
deps:
	@echo "整理Go模块依赖..."
	go mod tidy
	go mod download

# 格式化代码
fmt:
	@echo "格式化Go代码..."
	go fmt ./...

# 代码检查 (需要安装golangci-lint)
lint:
	@echo "执行代码检查..."
	@if command -v golangci-lint >/dev/null 2>&1; then \
		golangci-lint run; \
	else \
		echo "golangci-lint 未安装，跳过代码检查"; \
	fi

# 清理构建文件
clean:
	@echo "清理构建文件..."
	rm -rf bin/
	go clean ./...

# 创建目录
bin:
	mkdir -p bin

# 构建Docker镜像
docker:
	@echo "构建Docker镜像..."
	docker build -t heimdall-admin-api -f docker/Dockerfile.admin .
	docker build -t heimdall-public-api -f docker/Dockerfile.public .

# 开发环境启动 (需要先启动数据库服务)
dev: deps
	@echo "启动开发环境..."
	@echo "请确保MongoDB和Redis已启动"
	@echo "管理服务: http://localhost:8080"
	@echo "公开服务: http://localhost:8081"

# 生成API代码
generate:
	@echo "重新生成API代码..."
	cd admin-api/admin && goctl api go -api admin.api -dir . --style=goZero
	cd public-api/public && goctl api go -api public.api -dir . --style=goZero
	@echo "代码生成完成"

# 安装开发工具
install-tools:
	@echo "安装开发工具..."
	go install github.com/golangci/golangci-lint/cmd/golangci-lint@latest
	go install github.com/zeromicro/go-zero/tools/goctl@latest
	@echo "工具安装完成" 