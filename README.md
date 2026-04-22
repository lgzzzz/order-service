# Order Service - 订单处理微服务

基于 go-kratos 和 Kafka 的消息驱动订单处理微服务模板，支持 gRPC 调用。

## 技术栈

- **框架**: [go-kratos](https://go-kratos.dev/) v2
- **消息队列**: Kafka (使用 [segmentio/kafka-go](https://github.com/segmentio/kafka-go))
- **RPC**: gRPC + Protocol Buffers
- **数据库**: MySQL + [GORM](https://gorm.io/)
- **依赖注入**: [google/wire](https://github.com/google/wire)
- **配置**: YAML

## 项目结构

```
order-service/
├── api/
│   └── proto/
│       └── order/v1/
│           └── order.proto # Protobuf 定义
├── cmd/
│   └── server/
│       ├── main.go         # 应用入口
│       ├── wire.go         # 依赖注入配置
│       └── wire_gen.go     # Wire 生成的代码
├── configs/
│   └── config.yaml         # 配置文件
├── internal/
│   ├── biz/
│   │   └── order.go        # 业务逻辑层 (UseCase)
│   ├── conf/
│   │   ├── config.go       # 配置结构定义
│   │   └── grpc.go         # gRPC 配置定义
│   ├── consumer/
│   │   └── order.go        # Kafka 消费者
│   ├── data/
│   │   └── order.go        # 数据访问层 (Repository)
│   ├── model/
│   │   └── order.go        # 数据模型 (Entity/GORM)
│   ├── server/
│   │   └── grpc.go         # gRPC 服务器初始化
│   └── service/
│       └── order.go        # API 服务实现 (DTO 转换与业务分发)
├── docs/
│   └── 2026-04-14-order-service-design.md
├── docker-compose.yml      # Docker 配置
├── Makefile                # 构建脚本
├── go.mod
└── README.md
```

## 快速开始

### 1. 环境要求

- Go 1.21+
- Docker & Docker Compose
- Kafka
- MySQL

### 2. 启动依赖服务

```bash
make docker
```

这将启动 Kafka、Zookeeper 和 MySQL 容器。

### 3. 安装依赖

```bash
make init
```

### 4. 生成代码

生成 Protobuf 代码：
```bash
make protobuf
```

生成 Wire 依赖注入代码：
```bash
make wire
```

### 5. 运行服务

```bash
make run
```

或者构建后运行：
```bash
make build
./bin/server
```

## 配置

编辑 `configs/config.yaml` 文件来修改配置：

- **Kafka**: 修改 brokers、topic、消费者组等
- **数据库**: 修改 DSN、连接池大小等

## 消息格式

使用 Protobuf 定义订单消息格式，位于 `api/proto/order/v1/order.proto`：

- `OrderCreatedEvent`: 订单创建事件
- `OrderUpdatedEvent`: 订单更新事件
- `OrderCancelledEvent`: 订单取消事件
- `OrderEvent`: 联合消息，包含所有事件类型

## gRPC 服务

服务启动后，可以通过 gRPC 调用以下接口：

### 服务地址

默认地址：`localhost:50051`

### 可用接口

| 方法 | 说明 |
|------|------|
| `CreateOrder` | 创建订单 |
| `GetOrder` | 查询订单 |
| `UpdateOrderStatus` | 更新订单状态 |
| `CancelOrder` | 取消订单 |
| `ListOrders` | 查询订单列表 |

### 调用示例

使用 `grpcurl` 调用：

```bash
# 查询订单
grpcurl -plaintext -d '{"order_id": "order-123"}' \
  localhost:50051 order.v1.OrderService/GetOrder

# 创建订单
grpcurl -plaintext -d '{
  "user_id": "user-123",
  "shipping_address": "北京市朝阳区",
  "items": [
    {"product_id": "prod-1", "product_name": "商品1", "quantity": 2, "price": 1000}
  ]
}' localhost:50051 order.v1.OrderService/CreateOrder
```

## 扩展

### 添加新的消费者

1. 在 `internal/consumer/` 创建新的消费者文件
2. 在 `cmd/server/wire.go` 中注册依赖
3. 在 `cmd/server/main.go` 中启动消费者

### 添加新的业务逻辑

1. 在 `internal/biz/` 中定义新的 `UseCase` 方法
2. 在 `internal/service/` 中添加对应的 gRPC 服务方法进行调用

### 添加新的数据模型

1. 在 `internal/model/` 中定义模型
2. 在 `internal/data/` 中实现仓储 (Repository)
3. 在 `cmd/server/main.go` 的 `autoMigrate` 中注册模型

## 测试

```bash
make test
```

## 清理

```bash
make clean
```

停止 Docker 容器：
```bash
make docker-down
```

## 开发工具

### 安装 protoc 编译器

```bash
# macOS
brew install protobuf

# Ubuntu/Debian
sudo apt install protobuf-compiler

# 或者从 https://github.com/protocolbuffers/protobuf/releases 下载
```

### 安装 Go 插件

```bash
go install google.golang.org/protobuf/cmd/protoc-gen-go@latest
go install google.golang.org/grpc/cmd/protoc-gen-go-grpc@latest
go install github.com/google/wire/cmd/wire@latest
```

## 许可证

MIT
