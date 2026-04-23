package service

import (
	"context"
	"strconv"

	orderv1 "order-service/api/proto/order/v1"
	"order-service/internal/biz"
	"order-service/internal/model"

	"github.com/go-kratos/kratos/v2/errors"
	"github.com/go-kratos/kratos/v2/log"
	"google.golang.org/protobuf/types/known/timestamppb"
)

// OrderService gRPC 订单服务实现
type OrderService struct {
	orderv1.UnimplementedOrderServiceServer

	orderUseCase *biz.OrderUseCase
	log          *log.Helper
}

// NewOrderService 创建 gRPC 订单服务实例
func NewOrderService(orderUseCase *biz.OrderUseCase, logger log.Logger) *OrderService {
	return &OrderService{
		orderUseCase: orderUseCase,
		log:          log.NewHelper(logger),
	}
}

// CreateOrder 创建订单
func (s *OrderService) CreateOrder(ctx context.Context, req *orderv1.CreateOrderRequest) (*orderv1.CreateOrderResponse, error) {
	s.log.Infof("gRPC CreateOrder called: user_id=%d", req.UserId)

	if req.UserId <= 0 {
		return nil, errors.BadRequest("INVALID_REQUEST", "user_id is required")
	}
	if req.AddressId <= 0 {
		return nil, errors.BadRequest("INVALID_REQUEST", "address_id is required")
	}
	if len(req.Items) == 0 {
		return nil, errors.BadRequest("INVALID_REQUEST", "items cannot be empty")
	}

	items := make([]model.OrderItem, 0, len(req.Items))
	for _, item := range req.Items {
		if item.ProductId == "" {
			return nil, errors.BadRequest("INVALID_REQUEST", "product_id is required")
		}
		if item.Quantity <= 0 {
			return nil, errors.BadRequest("INVALID_REQUEST", "quantity must be positive")
		}
		items = append(items, model.OrderItem{
			ProductID: item.ProductId,
			Quantity:  item.Quantity,
		})
	}

	order, err := s.orderUseCase.CreateOrder(ctx, req.UserId, items, req.AddressId)
	if err != nil {
		s.log.Errorf("Failed to create order: %v", err)
		return nil, err
	}

	pbOrder := modelToProto(order)
	return &orderv1.CreateOrderResponse{
		OrderId: order.ID,
		Order:   pbOrder,
	}, nil
}

// GetOrder 查询订单
func (s *OrderService) GetOrder(ctx context.Context, req *orderv1.GetOrderRequest) (*orderv1.GetOrderResponse, error) {
	s.log.Infof("gRPC GetOrder called: order_id=%s", req.OrderId)

	if req.OrderId == "" {
		return nil, errors.BadRequest("INVALID_REQUEST", "order_id is required")
	}

	order, err := s.orderUseCase.GetOrder(ctx, req.OrderId)
	if err != nil {
		s.log.Errorf("Failed to get order: %v", err)
		return nil, err
	}

	pbOrder := modelToProto(order)
	return &orderv1.GetOrderResponse{
		Order: pbOrder,
	}, nil
}

// UpdateOrderStatus 更新订单状态
func (s *OrderService) UpdateOrderStatus(ctx context.Context, req *orderv1.UpdateOrderStatusRequest) (*orderv1.UpdateOrderStatusResponse, error) {
	s.log.Infof("gRPC UpdateOrderStatus called: order_id=%s, status=%s", req.OrderId, req.Status.String())

	if req.OrderId == "" {
		return nil, errors.BadRequest("INVALID_REQUEST", "order_id is required")
	}

	newStatus := model.OrderStatus(req.Status)
	if err := s.orderUseCase.HandleOrderUpdated(ctx, req.OrderId, newStatus); err != nil {
		s.log.Errorf("Failed to update order status: %v", err)
		return &orderv1.UpdateOrderStatusResponse{
			Success: false,
		}, err
	}

	return &orderv1.UpdateOrderStatusResponse{
		Success: true,
	}, nil
}

// CancelOrder 取消订单
func (s *OrderService) CancelOrder(ctx context.Context, req *orderv1.CancelOrderRequest) (*orderv1.CancelOrderResponse, error) {
	s.log.Infof("gRPC CancelOrder called: order_id=%s, reason=%s", req.OrderId, req.Reason)

	if req.OrderId == "" {
		return nil, errors.BadRequest("INVALID_REQUEST", "order_id is required")
	}

	if err := s.orderUseCase.HandleOrderCancelled(ctx, req.OrderId, req.Reason); err != nil {
		s.log.Errorf("Failed to cancel order: %v", err)
		return &orderv1.CancelOrderResponse{
			Success: false,
		}, err
	}

	return &orderv1.CancelOrderResponse{
		Success: true,
	}, nil
}

// ListOrders 查询订单列表
func (s *OrderService) ListOrders(ctx context.Context, req *orderv1.ListOrdersRequest) (*orderv1.ListOrdersResponse, error) {
	s.log.Infof("gRPC ListOrders called: user_id=%s, page_size=%d", req.UserId, req.PageSize)

	if req.UserId == "" {
		return nil, errors.BadRequest("INVALID_REQUEST", "user_id is required")
	}

	pageSize := int(req.PageSize)
	if pageSize <= 0 {
		pageSize = 20
	}
	if pageSize > 100 {
		pageSize = 100
	}
	offset := int(req.PageToken)

	userID, err := strconv.ParseInt(req.UserId, 10, 64)
	if err != nil {
		return nil, errors.BadRequest("INVALID_REQUEST", "invalid user_id format")
	}

	orders, err := s.orderUseCase.ListUserOrders(ctx, userID, pageSize, offset)
	if err != nil {
		s.log.Errorf("Failed to list orders: %v", err)
		return nil, err
	}

	pbOrders := make([]*orderv1.Order, 0, len(orders))
	for _, order := range orders {
		pbOrders = append(pbOrders, modelToProto(&order))
	}

	hasMore := len(orders) == pageSize
	nextPageToken := int32(0)
	if hasMore {
		nextPageToken = int32(offset + pageSize)
	}

	return &orderv1.ListOrdersResponse{
		Orders:        pbOrders,
		NextPageToken: nextPageToken,
		TotalCount:    int32(len(orders)),
	}, nil
}

// modelToProto 将模型转换为 Protobuf 格式
func modelToProto(order *model.Order) *orderv1.Order {
	pbItems := make([]*orderv1.OrderItem, 0, len(order.Items))
	for _, item := range order.Items {
		pbItems = append(pbItems, &orderv1.OrderItem{
			Id:          item.ID,
			ProductId:   item.ProductID,
			ProductName: item.ProductName,
			Quantity:    item.Quantity,
			Price:       item.Price,
		})
	}

	return &orderv1.Order{
		Id:              order.ID,
		UserId:          order.UserID,
		Status:          orderv1.OrderStatus(order.Status),
		Items:           pbItems,
		TotalAmount:     order.TotalAmount,
		ShippingAddress: order.ShippingAddress,
		CreatedAt:       timestamppb.New(order.CreatedAt),
		UpdatedAt:       timestamppb.New(order.UpdatedAt),
	}
}
