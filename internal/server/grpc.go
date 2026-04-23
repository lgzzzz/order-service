package server

import (
	"order-service/internal/conf"
	"order-service/internal/middleware"

	"github.com/go-kratos/kratos/v2/middleware/recovery"
	"github.com/go-kratos/kratos/v2/transport/grpc"
)

func NewGRPCServer(cfg *conf.Config) *grpc.Server {
	var opts = []grpc.ServerOption{
		grpc.Address(cfg.GRPC.Addr),
		grpc.Middleware(
			recovery.Recovery(),
			middleware.ResponseError(),
		),
	}
	srv := grpc.NewServer(opts...)
	return srv
}
