//go:build wireinject
// +build wireinject

//go:generate go run -mod=mod github.com/google/wire/cmd/wire
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
	"go.opentelemetry.io/otel/trace"
	"gorm.io/gorm"
)

func initApp(cfg *conf.Config, db *gorm.DB, logger log.Logger, tracer trace.Tracer) (*kratos.App, func(), error) {
	wire.Build(
		data.ProviderSet,
		biz.NewOrderUseCase,
		consumer.NewOrderConsumer,
		ProvideKafkaConfig,
		server.NewGRPCServer,
		service.NewOrderService,
		NewApp,
	)
	return &kratos.App{}, func() {}, nil
}

func ProvideKafkaConfig(cfg *conf.Config) *conf.KafkaConfig {
	return &cfg.Kafka
}

func NewApp(
	cfg *conf.Config,
	grpcServer *transportgrpc.Server,
	orderService *service.OrderService,
	orderConsumer *consumer.OrderConsumer,
) (*kratos.App, error) {
	orderv1.RegisterOrderServiceServer(grpcServer, orderService)

	client, err := clientv3.New(clientv3.Config{
		Endpoints:   cfg.Registry.Endpoints,
		DialTimeout: time.Duration(cfg.Registry.Timeout) * time.Second,
	})
	if err != nil {
		return nil, err
	}

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
