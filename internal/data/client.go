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
	kratosgrpc "github.com/go-kratos/kratos/v2/transport/grpc"
	"google.golang.org/grpc"
)

type userRepo struct {
	client userV1.UserServiceClient
	log    *log.Helper
}

func NewUserRepo(r registry.Discovery, logger log.Logger) biz.UserRepo {
	conn := newGRPCClient(r, "discovery:///user-service")
	return &userRepo{
		client: userV1.NewUserServiceClient(conn),
		log:    log.NewHelper(logger),
	}
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
	log    *log.Helper
}

func NewProductRepo(r registry.Discovery, logger log.Logger) biz.ProductRepo {
	conn := newGRPCClient(r, "discovery:///product-service")
	return &productRepo{
		client: productV1.NewProductServiceClient(conn),
		log:    log.NewHelper(logger),
	}
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
	log    *log.Helper
}

func NewInventoryRepo(r registry.Discovery, logger log.Logger) biz.InventoryRepo {
	conn := newGRPCClient(r, "discovery:///inventory-service")
	return &inventoryRepo{
		client: inventoryV1.NewInventoryClient(conn),
		log:    log.NewHelper(logger),
	}
}

func (r *inventoryRepo) LockStock(ctx context.Context, skuID int64, quantity int32) error {
	_, err := r.client.LockStock(ctx, &inventoryV1.LockStockRequest{
		SkuId:    skuID,
		Quantity: int64(quantity),
	})
	return err
}

func (r *inventoryRepo) UnlockStock(ctx context.Context, skuID int64, quantity int32) error {
	_, err := r.client.UnlockStock(ctx, &inventoryV1.UnlockStockRequest{
		SkuId:    skuID,
		Quantity: int64(quantity),
	})
	return err
}

type cartRepo struct {
	client cartV1.CartServiceClient
	log    *log.Helper
}

func NewCartRepo(r registry.Discovery, logger log.Logger) biz.CartRepo {
	conn := newGRPCClient(r, "discovery:///cart-service")
	return &cartRepo{
		client: cartV1.NewCartServiceClient(conn),
		log:    log.NewHelper(logger),
	}
}

func (r *cartRepo) ClearCart(ctx context.Context, userID int64) error {
	_, err := r.client.ClearCart(ctx, &cartV1.ClearCartRequest{UserId: userID})
	return err
}

func newGRPCClient(r registry.Discovery, endpoint string) *grpc.ClientConn {
	conn, err := kratosgrpc.DialInsecure(
		context.Background(),
		kratosgrpc.WithEndpoint(endpoint),
		kratosgrpc.WithDiscovery(r),
	)
	if err != nil {
		panic(err)
	}
	return conn
}
