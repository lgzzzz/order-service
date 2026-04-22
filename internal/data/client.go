package data

import (
	"context"
	"fmt"

	cartV1 "cart-service/api/cart/v1"
	inventoryV1 "inventory-service/api/inventory/v1"
	productV1 "product-service/api/product/v1"
	userV1 "user-service/api/user/v1"

	"order-service/internal/biz"

	"github.com/go-kratos/kratos/v2/log"
	"github.com/go-kratos/kratos/v2/registry"
	"github.com/go-kratos/kratos/v2/transport/grpc"
)

type userRepo struct {
	client userV1.UserServiceClient
}

func NewUserRepo(r registry.Discovery, logger log.Logger) biz.UserRepo {
	conn, err := grpc.DialInsecure(
		context.Background(),
		grpc.WithEndpoint("discovery:///user-service"),
		grpc.WithDiscovery(r),
	)
	if err != nil {
		panic(err)
	}
	return &userRepo{client: userV1.NewUserServiceClient(conn)}
}

func (r *userRepo) GetAddress(ctx context.Context, id int64) (string, error) {
	reply, err := r.client.GetAddress(ctx, &userV1.GetAddressRequest{Id: id})
	if err != nil {
		return "", err
	}
	return fmt.Sprintf("%s %s %s %s", reply.Province, reply.City, reply.District, reply.Detail), nil
}

type productRepo struct {
	client productV1.ProductServiceClient
}

func NewProductRepo(r registry.Discovery, logger log.Logger) biz.ProductRepo {
	conn, err := grpc.DialInsecure(
		context.Background(),
		grpc.WithEndpoint("discovery:///product-service"),
		grpc.WithDiscovery(r),
	)
	if err != nil {
		panic(err)
	}
	return &productRepo{client: productV1.NewProductServiceClient(conn)}
}

func (r *productRepo) GetProductPrice(ctx context.Context, id int64) (int64, string, error) {
	reply, err := r.client.GetProduct(ctx, &productV1.GetProductRequest{Id: id})
	if err != nil {
		return 0, "", err
	}
	return reply.Price, reply.Name, nil
}

type inventoryRepo struct {
	client inventoryV1.InventoryClient
}

func NewInventoryRepo(r registry.Discovery, logger log.Logger) biz.InventoryRepo {
	conn, err := grpc.DialInsecure(
		context.Background(),
		grpc.WithEndpoint("discovery:///inventory-service"),
		grpc.WithDiscovery(r),
	)
	if err != nil {
		panic(err)
	}
	return &inventoryRepo{client: inventoryV1.NewInventoryClient(conn)}
}

func (r *inventoryRepo) LockStock(ctx context.Context, skuID int64, quantity int32) error {
	_, err := r.client.LockStock(ctx, &inventoryV1.LockStockRequest{
		SkuId:    skuID,
		Quantity: int64(quantity),
	})
	return err
}

type cartRepo struct {
	client cartV1.CartServiceClient
}

func NewCartRepo(r registry.Discovery, logger log.Logger) biz.CartRepo {
	conn, err := grpc.DialInsecure(
		context.Background(),
		grpc.WithEndpoint("discovery:///cart-service"),
		grpc.WithDiscovery(r),
	)
	if err != nil {
		panic(err)
	}
	return &cartRepo{client: cartV1.NewCartServiceClient(conn)}
}

func (r *cartRepo) ClearCart(ctx context.Context, userID int64) error {
	_, err := r.client.ClearCart(ctx, &cartV1.ClearCartRequest{UserId: userID})
	return err
}
