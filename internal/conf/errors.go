package conf

import "github.com/go-kratos/kratos/v2/errors"

const (
	reasonOrderNotFound    = "ORDER_NOT_FOUND"
	reasonInsufficientStock = "INSUFFICIENT_STOCK"
	reasonUnauthorized     = "UNAUTHORIZED"
)

var (
	ErrOrderNotFound     = errors.NotFound(reasonOrderNotFound, "order not found")
	ErrInsufficientStock = errors.BadRequest(reasonInsufficientStock, "insufficient stock")
	ErrUnauthorized      = errors.Unauthorized(reasonUnauthorized, "unauthorized")
)
