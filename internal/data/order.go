package data

import (
	"context"
	"fmt"

	"order-service/internal/model"

	"github.com/go-kratos/kratos/v2/log"
	"gorm.io/gorm"
)

// OrderRepository 订单仓储
type OrderRepository struct {
	db  *gorm.DB
	log *log.Helper
}

// NewOrderRepository 创建订单仓储实例
func NewOrderRepository(db *gorm.DB, logger log.Logger) *OrderRepository {
	return &OrderRepository{
		db:  db,
		log: log.NewHelper(logger),
	}
}

// Create 创建订单
func (r *OrderRepository) Create(ctx context.Context, order *model.Order) error {
	return r.db.WithContext(ctx).Create(order).Error
}

// GetByID 根据 ID 查询订单
func (r *OrderRepository) GetByID(ctx context.Context, id string) (*model.Order, error) {
	var order model.Order
	err := r.db.WithContext(ctx).
		Preload("Items").
		Where("id = ?", id).
		First(&order).Error
	if err != nil {
		return nil, fmt.Errorf("failed to get order: %w", err)
	}
	return &order, nil
}

// Update 更新订单
func (r *OrderRepository) Update(ctx context.Context, order *model.Order) error {
	return r.db.WithContext(ctx).Save(order).Error
}

// UpdateStatus 更新订单状态
func (r *OrderRepository) UpdateStatus(ctx context.Context, id string, status model.OrderStatus) error {
	return r.db.WithContext(ctx).
		Model(&model.Order{}).
		Where("id = ?", id).
		Update("status", status).Error
}

// Delete 删除订单（软删除）
func (r *OrderRepository) Delete(ctx context.Context, id string) error {
	return r.db.WithContext(ctx).Delete(&model.Order{}, id).Error
}

// ListByUserID 查询用户的订单列表
func (r *OrderRepository) ListByUserID(ctx context.Context, userID int64, limit, offset int) ([]model.Order, error) {
	var orders []model.Order
	err := r.db.WithContext(ctx).
		Where("user_id = ?", userID).
		Order("created_at DESC").
		Limit(limit).
		Offset(offset).
		Find(&orders).Error
	if err != nil {
		return nil, fmt.Errorf("failed to list orders: %w", err)
	}
	return orders, nil
}

// Exists 检查订单是否存在
func (r *OrderRepository) Exists(ctx context.Context, id string) (bool, error) {
	var count int64
	err := r.db.WithContext(ctx).
		Model(&model.Order{}).
		Where("id = ? AND deleted_at IS NULL", id).
		Count(&count).Error
	if err != nil {
		return false, err
	}
	return count > 0, nil
}
