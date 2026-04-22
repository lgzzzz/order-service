package service

import (
	"context"
	"strconv"

	orderv1 "order-service/api/proto/order/v1"
	"order-service/internal/biz"
	"order-service/internal/model"

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

	items := make([]model.OrderItem, 0, len(req.Items))
	for _, item := range req.Items {
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

	// 构建响应
	pbOrder := modelToProto(order)
	return &orderv1.CreateOrderResponse{
		OrderId: order.ID,
		Order:   pbOrder,
	}, nil
}

// GetOrder 查询订单
func (s *OrderService) GetOrder(ctx context.Context, req *orderv1.GetOrderRequest) (*orderv1.GetOrderResponse, error) {
	s.log.Infof("gRPC GetOrder called: order_id=%s", req.OrderId)

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
	s.log.Infof("gRPC ListOrders called: user_id=%d, page_size=%d", req.UserId, req.PageSize)

	pageSize := int(req.PageSize)
	if pageSize <= 0 {
		pageSize = 20
	}
	offset := int(req.PageToken)

	userID, _ := strconv.ParseInt(req.UserId, 10, 64)
	orders, err := s.orderUseCase.ListUserOrders(ctx, userID, pageSize, offset)
	if err != nil {
		s.log.Errorf("Failed to list orders: %v", err)
		return nil, err
	}

	// 转换为 Protobuf 格式
	pbOrders := make([]*orderv1.Order, 0, len(orders))
	for _, order := range orders {
		pbOrders = append(pbOrders, modelToProto(&order))
	}

	nextPageToken := offset + pageSize
	if nextPageToken >= len(orders) {
		nextPageToken = 0
	}

	return &orderv1.ListOrdersResponse{
		Orders:        pbOrders,
		NextPageToken: int32(nextPageToken),
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
