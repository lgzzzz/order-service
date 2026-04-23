package biz

import (
	"context"
	"fmt"
	"strconv"
	"time"

	"order-service/internal/model"

	"github.com/go-kratos/kratos/v2/log"
	"github.com/google/uuid"
)

const (
	DefaultTimeout = 10 * time.Second
)

type UserRepo interface {
	GetAddress(ctx context.Context, id int64) (string, error)
}

type ProductRepo interface {
	GetProductPrice(ctx context.Context, id int64) (int64, string, error)
}

type InventoryRepo interface {
	LockStock(ctx context.Context, skuID int64, quantity int32) error
	UnlockStock(ctx context.Context, skuID int64, quantity int32) error
}

type CartRepo interface {
	ClearCart(ctx context.Context, userID int64) error
}

type OrderRepo interface {
	Create(ctx context.Context, order *model.Order) error
	GetByID(ctx context.Context, id string) (*model.Order, error)
	UpdateStatus(ctx context.Context, id string, status model.OrderStatus) error
	Exists(ctx context.Context, id string) (bool, error)
	ListByUserID(ctx context.Context, userID int64, limit, offset int) ([]model.Order, error)
	GetByIDempotencyKey(ctx context.Context, key string) (*model.Order, error)
}

// OrderUseCase 订单服务
type OrderUseCase struct {
	repo           OrderRepo
	userRepo       UserRepo
	productRepo    ProductRepo
	inventoryRepo  InventoryRepo
	cartRepo       CartRepo
	log            *log.Helper
	defaultTimeout time.Duration
}

// NewOrderUseCase 创建订单服务实例
func NewOrderUseCase(repo OrderRepo, userRepo UserRepo, productRepo ProductRepo, inventoryRepo InventoryRepo, cartRepo CartRepo, logger log.Logger) *OrderUseCase {
	return &OrderUseCase{
		repo:           repo,
		userRepo:       userRepo,
		productRepo:    productRepo,
		inventoryRepo:  inventoryRepo,
		cartRepo:       cartRepo,
		log:            log.NewHelper(logger),
		defaultTimeout: DefaultTimeout,
	}
}

// CreateOrder 完整下单流程
func (s *OrderUseCase) CreateOrder(ctx context.Context, userID int64, items []model.OrderItem, addressID int64) (*model.Order, error) {
	return s.CreateOrderWithIdempotency(ctx, userID, items, addressID, "")
}

// CreateOrderWithIdempotency 带有幂等性的创建订单
func (s *OrderUseCase) CreateOrderWithIdempotency(ctx context.Context, userID int64, items []model.OrderItem, addressID int64, idempotencyKey string) (*model.Order, error) {
	if len(items) == 0 {
		return nil, fmt.Errorf("order items cannot be empty")
	}
	if userID <= 0 {
		return nil, fmt.Errorf("invalid user ID")
	}
	if addressID <= 0 {
		return nil, fmt.Errorf("invalid address ID")
	}

	// 幂等性检查
	if idempotencyKey != "" {
		existing, err := s.repo.GetByIDempotencyKey(ctx, idempotencyKey)
		if err != nil {
			return nil, err
		}
		if existing != nil {
			return existing, nil
		}
	}

	// 带超时的context
	ctx, cancel := context.WithTimeout(ctx, s.defaultTimeout)
	defer cancel()

	// 1. 获取收货地址
	address, err := s.userRepo.GetAddress(ctx, addressID)
	if err != nil {
		return nil, fmt.Errorf("failed to get address: %w", err)
	}

	var totalAmount int64
	lockedProducts := make([]struct {
		pid      int64
		quantity int32
	}, 0, len(items))

	// 2. 校验商品并计算总价，锁定库存
	for i, item := range items {
		if item.ProductID == "" {
			_ = s.rollbackLockedStock(ctx, lockedProducts)
			return nil, fmt.Errorf("product ID cannot be empty")
		}

		pid, err := strconv.ParseInt(item.ProductID, 10, 64)
		if err != nil {
			_ = s.rollbackLockedStock(ctx, lockedProducts)
			return nil, fmt.Errorf("invalid product ID %s: %w", item.ProductID, err)
		}
		if item.Quantity <= 0 {
			_ = s.rollbackLockedStock(ctx, lockedProducts)
			return nil, fmt.Errorf("quantity must be positive for product %s", item.ProductID)
		}

		price, name, err := s.productRepo.GetProductPrice(ctx, pid)
		if err != nil {
			_ = s.rollbackLockedStock(ctx, lockedProducts)
			return nil, fmt.Errorf("failed to get product %s: %w", item.ProductID, err)
		}
		items[i].Price = price
		items[i].ProductName = name
		totalAmount += price * int64(item.Quantity)

		// 3. 锁定库存
		if err := s.inventoryRepo.LockStock(ctx, pid, item.Quantity); err != nil {
			_ = s.rollbackLockedStock(ctx, lockedProducts)
			return nil, fmt.Errorf("failed to lock stock for product %s: %w", item.ProductID, err)
		}
		lockedProducts = append(lockedProducts, struct {
			pid      int64
			quantity int32
		}{pid: pid, quantity: item.Quantity})
	}

	// 4. 创建订单对象
	order := &model.Order{
		ID:              "ORD-" + uuid.New().String(),
		UserID:          userID,
		Status:          model.OrderStatusCreated,
		Items:           items,
		TotalAmount:     totalAmount,
		ShippingAddress: address,
	}

	// 5. 保存到数据库
	if err := s.repo.Create(ctx, order); err != nil {
		if rbErr := s.rollbackLockedStock(ctx, lockedProducts); rbErr != nil {
			s.log.Errorf("order save failed and rollback also failed: order_err=%v, rollback_err=%v", err, rbErr)
		}
		return nil, fmt.Errorf("failed to create order: %w", err)
	}

	// 6. 清空购物车（异步/可重试，不阻塞主流程）
	go func(uid int64) {
		if err := s.cartRepo.ClearCart(context.Background(), uid); err != nil {
			s.log.Warnf("failed to clear cart for user %d: %v", uid, err)
		}
	}(userID)

	return order, nil
}

func (s *OrderUseCase) rollbackLockedStock(ctx context.Context, lockedProducts []struct {
	pid      int64
	quantity int32
}) error {
	var lastErr error
	for _, p := range lockedProducts {
		if err := s.inventoryRepo.UnlockStock(ctx, p.pid, p.quantity); err != nil {
			s.log.Errorf("failed to rollback stock for product %d: %v", p.pid, err)
			lastErr = err
		}
	}
	return lastErr
}

// HandleOrderCreated 处理订单创建事件
func (s *OrderUseCase) HandleOrderCreated(ctx context.Context, order *model.Order) error {
	s.log.Infof("Processing order created: %s", order.ID)

	// 验证订单数据
	if err := s.validateOrder(order); err != nil {
		return fmt.Errorf("invalid order: %w", err)
	}

	// 保存订单到数据库
	if err := s.repo.Create(ctx, order); err != nil {
		return fmt.Errorf("failed to create order: %w", err)
	}

	s.log.Infof("Order created successfully: %s", order.ID)
	return nil
}

// HandleOrderUpdated 处理订单更新事件
func (s *OrderUseCase) HandleOrderUpdated(ctx context.Context, orderID string, newStatus model.OrderStatus) error {
	s.log.Infof("Processing order updated: %s, new status: %s", orderID, newStatus.String())

	// 检查订单是否存在
	exists, err := s.repo.Exists(ctx, orderID)
	if err != nil {
		return fmt.Errorf("failed to check order: %w", err)
	}
	if !exists {
		return fmt.Errorf("order not found: %s", orderID)
	}

	// 更新订单状态
	if err := s.repo.UpdateStatus(ctx, orderID, newStatus); err != nil {
		return fmt.Errorf("failed to update order: %w", err)
	}

	s.log.Infof("Order updated successfully: %s", orderID)
	return nil
}

// HandleOrderCancelled 处理订单取消事件
func (s *OrderUseCase) HandleOrderCancelled(ctx context.Context, orderID string, reason string) error {
	s.log.Infof("Processing order cancelled: %s, reason: %s", orderID, reason)

	// 更新订单状态为已取消
	if err := s.repo.UpdateStatus(ctx, orderID, model.OrderStatusCancelled); err != nil {
		return fmt.Errorf("failed to cancel order: %w", err)
	}

	s.log.Infof("Order cancelled successfully: %s", orderID)
	return nil
}

// GetOrder 查询订单
func (s *OrderUseCase) GetOrder(ctx context.Context, id string) (*model.Order, error) {
	return s.repo.GetByID(ctx, id)
}

// ListUserOrders 查询用户订单列表
func (s *OrderUseCase) ListUserOrders(ctx context.Context, userID int64, limit, offset int) ([]model.Order, error) {
	return s.repo.ListByUserID(ctx, userID, limit, offset)
}

// validateOrder 验证订单数据
func (s *OrderUseCase) validateOrder(order *model.Order) error {
	if order.ID == "" {
		return fmt.Errorf("order ID is required")
	}
	if order.UserID <= 0 {
		return fmt.Errorf("user ID is required")
	}
	if len(order.Items) == 0 {
		return fmt.Errorf("order items are required")
	}
	if order.TotalAmount <= 0 {
		return fmt.Errorf("total amount must be positive")
	}

	// 验证订单项
	for i, item := range order.Items {
		if item.ProductID == "" {
			return fmt.Errorf("item %d: product ID is required", i)
		}
		if item.Quantity <= 0 {
			return fmt.Errorf("item %d: quantity must be positive", i)
		}
		if item.Price <= 0 {
			return fmt.Errorf("item %d: price must be positive", i)
		}
	}

	return nil
}
