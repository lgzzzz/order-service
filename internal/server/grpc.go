package server

import (
	"time"

	"order-service/internal/conf"
	"order-service/internal/middleware"

	"github.com/go-kratos/kratos/v2/middleware/recovery"
	"github.com/go-kratos/kratos/v2/transport/grpc"
	"github.com/lgzzzz/mall-tracing/grpcutil"
	tracingmiddleware "github.com/lgzzzz/mall-tracing/middleware"
	"go.opentelemetry.io/otel/trace"
)

func NewGRPCServer(cfg *conf.Config, tracer trace.Tracer) *grpc.Server {
	return grpcutil.NewServerBuilder().
		WithAddress(cfg.GRPC.Addr).
		WithTimeout(time.Duration(cfg.GRPC.Timeout) * time.Second).
		WithMiddleware(
			recovery.Recovery(),
			tracingmiddleware.ServerMiddleware(tracer),
			middleware.ResponseError(),
		).
		Build()
}