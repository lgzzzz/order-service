package model

import (
	"time"

	"gorm.io/gorm"
)

// OrderStatus 订单状态
type OrderStatus int

const (
	OrderStatusUnspecified OrderStatus = 0
	OrderStatusCreated     OrderStatus = 1 // 已创建
	OrderStatusPaid        OrderStatus = 2 // 已支付
	OrderStatusShipped     OrderStatus = 3 // 已发货
	OrderStatusCompleted   OrderStatus = 4 // 已完成
	OrderStatusCancelled   OrderStatus = 5 // 已取消
)

func (s OrderStatus) String() string {
	names := map[OrderStatus]string{
		OrderStatusUnspecified: "未指定",
		OrderStatusCreated:     "已创建",
		OrderStatusPaid:        "已支付",
		OrderStatusShipped:     "已发货",
		OrderStatusCompleted:   "已完成",
		OrderStatusCancelled:   "已取消",
	}
	return names[s]
}

// Order 订单模型
type Order struct {
	ID              string      `gorm:"type:varchar(64);primaryKey" json:"id"`
	UserID          int64       `gorm:"type:bigint;index" json:"user_id"`
	Status          OrderStatus `gorm:"type:int;index" json:"status"`
	TotalAmount     int64       `gorm:"type:bigint" json:"total_amount"`
	ShippingAddress string      `gorm:"type:varchar(512)" json:"shipping_address"`
	CreatedAt       time.Time   `json:"created_at"`
	UpdatedAt       time.Time   `json:"updated_at"`
	DeletedAt       gorm.DeletedAt `gorm:"index" json:"-"`

	// 关联
	Items []OrderItem `gorm:"foreignKey:OrderID" json:"items"`
}

// TableName 指定表名
func (Order) TableName() string {
	return "orders"
}

// OrderItem 订单项模型
type OrderItem struct {
	ID        string `gorm:"type:varchar(64);primaryKey" json:"id"`
	OrderID   string `gorm:"type:varchar(64);index" json:"order_id"`
	ProductID string `gorm:"type:varchar(64)" json:"product_id"`
	ProductName string `gorm:"type:varchar(256)" json:"product_name"`
	Quantity  int32  `gorm:"type:int" json:"quantity"`
	Price     int64  `gorm:"type:bigint" json:"price"` // 单价（分）
	CreatedAt time.Time `json:"created_at"`
}

// TableName 指定表名
func (OrderItem) TableName() string {
	return "order_items"
}
