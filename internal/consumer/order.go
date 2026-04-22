package consumer

import (
	"context"
	"fmt"

	orderv1 "order-service/api/proto/order/v1"
	"order-service/internal/biz"
	"order-service/internal/conf"
	"order-service/internal/model"

	"github.com/go-kratos/kratos/v2/log"
	"github.com/segmentio/kafka-go"
	"google.golang.org/protobuf/encoding/protojson"
)

// OrderConsumer 订单消息消费者
type OrderConsumer struct {
	reader  *kafka.Reader
	service *biz.OrderUseCase
	config  *conf.KafkaConfig
	log     *log.Helper
}

// NewOrderConsumer 创建订单消费者实例
func NewOrderConsumer(cfg *conf.KafkaConfig, orderService *biz.OrderUseCase, logger log.Logger) *OrderConsumer {
	// 创建 Kafka reader
	reader := kafka.NewReader(kafka.ReaderConfig{
		Brokers:  cfg.Brokers,
		GroupID:  cfg.GroupID,
		Topic:    cfg.OrderTopic,
		MinBytes: cfg.MinBytes,
		MaxBytes: cfg.MaxBytes,
	})

	return &OrderConsumer{
		reader:  reader,
		service: orderService,
		config:  cfg,
		log:     log.NewHelper(logger),
	}
}

// Start 启动消费循环
func (c *OrderConsumer) Start(ctx context.Context) error {
	c.log.Infof("Starting order consumer, topic: %s, group: %s", c.config.OrderTopic, c.config.GroupID)

	// 启动多个 worker
	for i := 0; i < c.config.Workers; i++ {
		go c.worker(ctx, i)
	}

	// 等待上下文取消
	<-ctx.Done()
	return c.reader.Close()
}

// worker 消费 worker
func (c *OrderConsumer) worker(ctx context.Context, id int) {
	c.log.Infof("Worker %d started", id)

	for {
		select {
		case <-ctx.Done():
			c.log.Infof("Worker %d stopping", id)
			return
		default:
			// 读取消息
			msg, err := c.reader.ReadMessage(ctx)
			if err != nil {
				c.log.Errorf("Worker %d failed to read message: %v", id, err)
				continue
			}

			// 处理消息
			if err := c.handleMessage(ctx, msg); err != nil {
				c.log.Errorf("Worker %d failed to handle message: %v", id, err)
				// 这里可以实现重试逻辑或死信队列
				continue
			}

			// 提交偏移量
			if err := c.reader.CommitMessages(ctx, msg); err != nil {
				c.log.Errorf("Worker %d failed to commit offset: %v", id, err)
			}
		}
	}
}

// handleMessage 处理单条消息
func (c *OrderConsumer) handleMessage(ctx context.Context, msg kafka.Message) error {
	c.log.Infof("Received message from topic: %s, partition: %d, offset: %d",
		msg.Topic, msg.Partition, msg.Offset)

	// 反序列化 OrderEvent
	var event orderv1.OrderEvent
	err := protojson.Unmarshal(msg.Value, &event)
	if err != nil {
		return fmt.Errorf("failed to unmarshal order event: %w", err)
	}

	c.log.Infof("Processing event: %s, type: %s", event.EventId, event.EventType.String())

	switch event.EventType {
	case orderv1.OrderEventType_ORDER_EVENT_TYPE_CREATED:
		if created := event.GetCreated(); created != nil && created.Order != nil {
			order := protoToModel(created.Order)
			return c.service.HandleOrderCreated(ctx, order)
		}
	case orderv1.OrderEventType_ORDER_EVENT_TYPE_UPDATED:
		if updated := event.GetUpdated(); updated != nil {
			return c.service.HandleOrderUpdated(ctx, updated.OrderId, model.OrderStatus(updated.NewStatus))
		}
	case orderv1.OrderEventType_ORDER_EVENT_TYPE_CANCELLED:
		if cancelled := event.GetCancelled(); cancelled != nil {
			return c.service.HandleOrderCancelled(ctx, cancelled.OrderId, cancelled.Reason)
		}
	default:
		c.log.Warnf("Unknown event type: %v", event.EventType)
	}

	return nil
}

// protoToModel 将 Protobuf 订单转换为内部模型
func protoToModel(pbOrder *orderv1.Order) *model.Order {
	order := &model.Order{
		ID:              pbOrder.Id,
		UserID:          pbOrder.UserId,
		Status:          model.OrderStatus(pbOrder.Status),
		TotalAmount:     pbOrder.TotalAmount,
		ShippingAddress: pbOrder.ShippingAddress,
	}

	if pbOrder.CreatedAt != nil {
		order.CreatedAt = pbOrder.CreatedAt.AsTime()
	}
	if pbOrder.UpdatedAt != nil {
		order.UpdatedAt = pbOrder.UpdatedAt.AsTime()
	}

	for _, item := range pbOrder.Items {
		order.Items = append(order.Items, model.OrderItem{
			ID:          item.Id,
			ProductID:   item.ProductId,
			ProductName: item.ProductName,
			Quantity:    item.Quantity,
			Price:       item.Price,
		})
	}

	return order
}

// Stop 停止消费循环
func (c *OrderConsumer) Stop(ctx context.Context) error {
	c.log.Info("Stopping order consumer...")
	return c.reader.Close()
}
