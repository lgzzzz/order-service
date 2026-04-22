.PHONY: build run test clean docker protobuf wire install-protobuf-tools

# 构建项目
build:
	@echo "Building..."
	@go build -o bin/server ./cmd/server

# 运行服务
run:
	@echo "Starting service..."
	@go run ./cmd/server

# 运行测试
test:
	@echo "Running tests..."
	@go test -v ./...

# 清理
clean:
	@rm -rf bin/
	@go clean

# 启动 Docker 依赖
docker:
	@echo "Starting Docker containers..."
	@docker-compose up -d

# 停止 Docker 依赖
docker-down:
	@echo "Stopping Docker containers..."
	@docker-compose down

# 生成 Protobuf 代码
protobuf:
	@echo "Generating protobuf code..."
	@protoc --go_out=. --go_opt=paths=source_relative \
		--go-grpc_out=. --go-grpc_opt=paths=source_relative \
		api/proto/order/v1/order.proto

# 生成 Wire 依赖注入代码
wire:
	@echo "Generating wire..."
	@cd cmd/server && wire

# 安装 Protobuf 工具
install-protobuf-tools:
	@echo "Installing protobuf tools..."
	@go install google.golang.org/protobuf/cmd/protoc-gen-go@latest
	@go install google.golang.org/grpc/cmd/protoc-gen-go-grpc@latest
	@echo "Please also install protoc binary from: https://github.com/protocolbuffers/protobuf/releases"

# 初始化项目
init:
	@echo "Initializing project..."
	@go mod download
	@go mod tidy
