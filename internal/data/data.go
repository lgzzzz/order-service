package data

import (
	"order-service/internal/biz"
	"order-service/internal/conf"

	etcd "github.com/go-kratos/kratos/contrib/registry/etcd/v2"
	"github.com/go-kratos/kratos/v2/log"
	"github.com/go-kratos/kratos/v2/registry"
	"github.com/google/wire"
	clientv3 "go.etcd.io/etcd/client/v3"
	"gorm.io/gorm"
)

// ProviderSet is data providers.
var ProviderSet = wire.NewSet(
	NewData,
	NewOrderRepository,
	ProvideOrderRepo,
	NewUserRepo,
	NewProductRepo,
	NewInventoryRepo,
	NewCartRepo,
	NewDiscovery,
)

// Data .
type Data struct {
	db *gorm.DB
}

// NewData .
func NewData(db *gorm.DB, logger log.Logger) (*Data, func(), error) {
	cleanup := func() {
		log.NewHelper(logger).Info("closing the data resources")
	}
	return &Data{db: db}, cleanup, nil
}

// ProvideOrderRepo 将 OrderRepository 作为 biz.OrderRepo 接口提供
func ProvideOrderRepo(repo *OrderRepository) biz.OrderRepo {
	return repo
}

// NewDiscovery .
func NewDiscovery(conf *conf.Config, logger log.Logger) registry.Discovery {
	client, err := clientv3.New(clientv3.Config{
		Endpoints: conf.Registry.Endpoints,
	})
	if err != nil {
		log.NewHelper(logger).Errorf("failed to create etcd client: %v", err)
		panic(err)
	}
	return etcd.New(client)
}
