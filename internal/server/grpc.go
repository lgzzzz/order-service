package server

import (
	"order-service/internal/conf"

	"github.com/go-kratos/kratos/v2/transport/grpc"
)

// NewGRPCServer 创建 Kratos gRPC 服务器实例
func NewGRPCServer(cfg *conf.Config) *grpc.Server {
	var opts = []grpc.ServerOption{
		grpc.Address(cfg.GRPC.Addr),
	}
	// 可以根据需要添加更多配置，如 Timeout, Middleware 等
	srv := grpc.NewServer(opts...)
	return srv
}
