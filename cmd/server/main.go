package main

import (
	internalconfig "order-service/internal/conf"
	"order-service/internal/model"

	"os"

	"github.com/go-kratos/kratos/v2/config"
	"github.com/go-kratos/kratos/v2/config/file"
	"github.com/go-kratos/kratos/v2/log"
	"gorm.io/driver/mysql"
	"gorm.io/gorm"

	etcdconfig "github.com/go-kratos/kratos/contrib/config/etcd/v2"
	clientv3 "go.etcd.io/etcd/client/v3"
)

func main() {
	logger := log.With(log.NewStdLogger(os.Stdout),
		"ts", log.DefaultTimestamp,
		"caller", log.DefaultCaller,
		"service.name", "order-service",
		"service.version", "1.0.0",
	)
	h := log.NewHelper(logger)

	// 1. 加载本地引导配置（包含 etcd 地址等）
	c := config.New(
		config.WithSource(
			file.NewSource("configs/config.yaml"),
		),
	)
	if err := c.Load(); err != nil {
		h.Fatalf("failed to load config: %v", err)
	}

	var bc internalconfig.Config
	if err := c.Scan(&bc); err != nil {
		h.Fatalf("failed to scan config: %v", err)
	}

	// 2. 初始化 etcd 客户端
	client, err := clientv3.New(clientv3.Config{
		Endpoints: bc.Registry.Endpoints,
	})
	if err != nil {
		h.Fatalf("failed to create etcd client: %v", err)
	}

	// 3. 加载远程 etcd 配置
	source, err := etcdconfig.New(client, etcdconfig.WithPath("/order-service/config"))
	if err != nil {
		h.Fatalf("failed to create etcd source: %v", err)
	}

	remoteConfig := config.New(
		config.WithSource(source),
	)
	if err := remoteConfig.Load(); err != nil {
		h.Fatalf("failed to load remote config: %v", err)
	}

	var cfg internalconfig.Config
	if err := remoteConfig.Scan(&cfg); err != nil {
		h.Fatalf("failed to scan remote config: %v", err)
	}

	// 应用默认值
	applyDefaults(&cfg)

	// 初始化数据库
	db := initDatabase(cfg.Database, logger)

	// 使用 Wire 进行依赖注入
	app, err := initApp(&cfg, db, logger)
	if err != nil {
		h.Fatalf("failed to init app: %v", err)
	}

	// 启动应用
	if err := app.Run(); err != nil {
		h.Fatalf("failed to run application: %v", err)
	}

	h.Info("Application stopped gracefully")
}

// applyDefaults 设置默认值
func applyDefaults(cfg *internalconfig.Config) {
	if cfg.Kafka.Workers == 0 {
		cfg.Kafka.Workers = 3
	}
	if cfg.Kafka.MinBytes == 0 {
		cfg.Kafka.MinBytes = 10e3 // 10KB
	}
	if cfg.Kafka.MaxBytes == 0 {
		cfg.Kafka.MaxBytes = 10e6 // 10MB
	}
	if cfg.Kafka.ReadTimeout == 0 {
		cfg.Kafka.ReadTimeout = 10 // 10秒
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

// initDatabase 初始化数据库连接
func initDatabase(cfg internalconfig.DatabaseConfig, logger log.Logger) *gorm.DB {
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

	// 自动迁移表结构
	if err := autoMigrate(db); err != nil {
		h.Fatalf("Failed to migrate database: %v", err)
	}

	h.Info("Database connected and migrated successfully")
	return db
}

// autoMigrate 自动迁移数据库表
func autoMigrate(db *gorm.DB) error {
	return db.AutoMigrate(
		&model.Order{},
		&model.OrderItem{},
	)
}
