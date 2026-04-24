package main

import (
	"context"
	"flag"
	"os"
	"time"

	"order-service/internal/conf"
	"order-service/internal/model"

	"github.com/go-kratos/kratos/v2/config"
	"github.com/go-kratos/kratos/v2/config/file"
	etcdConfig "github.com/go-kratos/kratos/contrib/config/etcd/v2"
	"github.com/go-kratos/kratos/v2/log"
	"github.com/lgzzzz/mall-tracing/tracing"
	clientv3 "go.etcd.io/etcd/client/v3"
	"go.opentelemetry.io/otel/trace"
	"gorm.io/driver/mysql"
	"gorm.io/gorm"
)

// go build -ldflags "-X main.Version=x.y.z"
var (
	Version  string
	confPath string
)

func init() {
	flag.StringVar(&confPath, "conf", "configs/config.yaml", "config path, eg: -conf configs/config.yaml")
	flag.Parse()
}

func main() {
	logger := log.With(log.NewStdLogger(os.Stdout),
		"ts", log.DefaultTimestamp,
		"caller", log.DefaultCaller,
		"service.name", "order-service",
		"service.version", Version,
	)
	h := log.NewHelper(logger)

	c := config.New(
		config.WithSource(
			file.NewSource(confPath),
		),
	)
	defer c.Close()

	if err := c.Load(); err != nil {
		h.Fatalf("failed to load config: %v", err)
	}

	var bc conf.Bootstrap
	if err := c.Scan(&bc); err != nil {
		h.Fatalf("failed to scan config: %v", err)
	}

	if bc.ConfigCenter != nil && len(bc.ConfigCenter.Endpoints) > 0 {
		client, err := clientv3.New(clientv3.Config{
			Endpoints: bc.ConfigCenter.Endpoints,
		})
		if err != nil {
			h.Errorf("failed to create etcd client: %v", err)
		} else {
			source, err := etcdConfig.New(client, etcdConfig.WithPath(bc.ConfigCenter.Key))
			if err != nil {
				h.Errorf("failed to create etcd config source: %v", err)
			} else {
				remoteConfig := config.New(
					config.WithSource(source),
				)
				defer remoteConfig.Close()

				if err := remoteConfig.Load(); err != nil {
					h.Errorf("failed to load config from etcd: %v", err)
				} else {
					if err := remoteConfig.Scan(&bc); err != nil {
						h.Errorf("failed to scan config from etcd: %v", err)
					}
				}
			}
		}
	}

	applyDefaults(&bc)

	db := initDatabase(bc.Database, logger)

	var tp trace.TracerProvider
	tracer := tracing.NewTracer("order-service")
	if bc.Tracing.Enabled {
		var err error
		tp, err = tracing.Init(tracing.Config{
			ServiceName:  "order-service",
			Version:      Version,
			OTLPEndpoint: bc.Tracing.Endpoint,
			SampleRatio:  bc.Tracing.SampleRatio,
			Insecure:     true,
		})
		if err != nil {
			h.Errorf("failed to init tracer provider: %v", err)
		} else {
			h.Info("Tracer provider initialized")
		}
	}

	app, cleanup, err := initApp(&bc, db, logger, tracer)
	if err != nil {
		h.Fatalf("failed to init app: %v", err)
	}

	if err := app.Run(); err != nil {
		h.Fatalf("failed to run application: %v", err)
	}

	if tp != nil {
		shutdownCtx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
		defer cancel()
		if err := tracing.Shutdown(shutdownCtx, tp); err != nil {
			h.Errorf("failed to shutdown tracer provider: %v", err)
		}
	}
	cleanup()
	h.Info("Application stopped gracefully")
}

func applyDefaults(cfg *conf.Bootstrap) {
	if cfg.Kafka.Workers == 0 {
		cfg.Kafka.Workers = 3
	}
	if cfg.Kafka.MinBytes == 0 {
		cfg.Kafka.MinBytes = 10e3
	}
	if cfg.Kafka.MaxBytes == 0 {
		cfg.Kafka.MaxBytes = 10e6
	}
	if cfg.Kafka.ReadTimeout == 0 {
		cfg.Kafka.ReadTimeout = 10
	}
	if cfg.Database.MaxIdleConns == 0 {
		cfg.Database.MaxIdleConns = 10
	}
	if cfg.Database.MaxOpenConns == 0 {
		cfg.Database.MaxOpenConns = 100
	}
	if cfg.GRPC.Addr == "" {
		cfg.GRPC.Addr = ":50051"
	}
	if cfg.GRPC.MaxConnections == 0 {
		cfg.GRPC.MaxConnections = 1000
	}
	if len(cfg.Registry.Endpoints) == 0 {
		cfg.Registry.Endpoints = []string{"127.0.0.1:2379"}
	}
	if cfg.Registry.Timeout == 0 {
		cfg.Registry.Timeout = 5
	}
}

func initDatabase(cfg conf.DatabaseConfig, logger log.Logger) *gorm.DB {
	h := log.NewHelper(logger)
	db, err := gorm.Open(mysql.Open(cfg.DSN), &gorm.Config{})
	if err != nil {
		h.Fatalf("Failed to connect to database: %v", err)
	}

	sqlDB, err := db.DB()
	if err != nil {
		h.Fatalf("Failed to get underlying sql.DB: %v", err)
	}

	sqlDB.SetMaxIdleConns(cfg.MaxIdleConns)
	sqlDB.SetMaxOpenConns(cfg.MaxOpenConns)

	if err := autoMigrate(db); err != nil {
		h.Fatalf("Failed to migrate database: %v", err)
	}

	h.Info("Database connected and migrated successfully")
	return db
}

func autoMigrate(db *gorm.DB) error {
	return db.AutoMigrate(
		&model.Order{},
		&model.OrderItem{},
	)
}