//go:build wireinject
// +build wireinject

package main

import (
	orderv1 "order-service/api/proto/order/v1"
	"order-service/internal/biz"
	"order-service/internal/conf"
	"order-service/internal/consumer"
	"order-service/internal/data"
	"order-service/internal/server"
	"order-service/internal/service"
	"time"

	"github.com/go-kratos/kratos/contrib/registry/etcd/v2"
	"github.com/go-kratos/kratos/v2"
	"github.com/go-kratos/kratos/v2/log"
	transportgrpc "github.com/go-kratos/kratos/v2/transport/grpc"
	"github.com/google/wire"
	clientv3 "go.etcd.io/etcd/client/v3"
	"gorm.io/gorm"
)

// initApp 初始化应用
func initApp(cfg *conf.Config, db *gorm.DB, logger log.Logger) (*kratos.App, error) {
	wire.Build(
		// 数据层与远程客户端
		data.ProviderSet,
		// 业务层
		biz.NewOrderUseCase,
		// 消费者
		consumer.NewOrderConsumer,
		ProvideKafkaConfig,
		// gRPC 服务器
		server.NewGRPCServer,
		service.NewOrderService,
		// 应用
		NewApp,
	)
	return &kratos.App{}, nil
}

func ProvideKafkaConfig(cfg *conf.Config) *conf.KafkaConfig {
	return &cfg.Kafka
}

// NewApp 创建应用实例
func NewApp(
	cfg *conf.Config,
	grpcServer *transportgrpc.Server,
	orderService *service.OrderService,
	orderConsumer *consumer.OrderConsumer,
) (*kratos.App, error) {
	// 注册 gRPC 服务
	orderv1.RegisterOrderServiceServer(grpcServer, orderService)

	// 配置 etcd 客户端
	client, err := clientv3.New(clientv3.Config{
		Endpoints:   cfg.Registry.Endpoints,
		DialTimeout: time.Duration(cfg.Registry.Timeout) * time.Second,
	})
	if err != nil {
		return nil, err
	}

	// 创建 etcd 注册中心
	reg := etcd.New(client)

	return kratos.New(
		kratos.Name("order-service"),
		kratos.Version("1.0.0"),
		kratos.Server(
			grpcServer,
			orderConsumer,
		),
		kratos.Registrar(reg),
	), nil
}

func provideConfig(cfg *conf.Config) conf.Config {
	return *cfg
}

func provideKafkaConfig(cfg conf.Config) *conf.KafkaConfig {
	return &cfg.Kafka
}
