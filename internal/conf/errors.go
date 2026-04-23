package conf

import "github.com/go-kratos/kratos/v2/errors"

const (
	ErrCodeSuccess           = 0
	ErrCodeInternal          = 30000
	ErrCodeOrderNotFound     = 30001
	ErrCodeInsufficientStock = 30002
	ErrCodeUnauthorized      = 30003
	ErrCodeInvalidRequest    = 30004
	ErrCodeCreateOrderFailed = 30005
	ErrCodeCancelOrderFailed = 30006
	ErrCodeUpdateOrderFailed = 30007
)

const (
	reasonOrderNotFound     = "ORDER_NOT_FOUND"
	reasonInsufficientStock = "INSUFFICIENT_STOCK"
	reasonUnauthorized      = "UNAUTHORIZED"
	reasonInvalidRequest    = "INVALID_REQUEST"
	reasonCreateOrderFailed = "CREATE_ORDER_FAILED"
	reasonCancelOrderFailed = "CANCEL_ORDER_FAILED"
	reasonUpdateOrderFailed = "UPDATE_ORDER_FAILED"
	reasonInternal          = "INTERNAL_ERROR"
)

var (
	ErrOrderNotFound     = errors.NotFound(reasonOrderNotFound, "order not found")
	ErrInsufficientStock = errors.BadRequest(reasonInsufficientStock, "insufficient stock")
	ErrUnauthorized      = errors.Unauthorized(reasonUnauthorized, "unauthorized")
	ErrInternal          = errors.InternalServer(reasonInternal, "internal server error")
)

var reasonToCode = map[string]int32{
	reasonOrderNotFound:     ErrCodeOrderNotFound,
	reasonInsufficientStock: ErrCodeInsufficientStock,
	reasonUnauthorized:      ErrCodeUnauthorized,
	reasonInvalidRequest:    ErrCodeInvalidRequest,
	reasonCreateOrderFailed: ErrCodeCreateOrderFailed,
	reasonCancelOrderFailed: ErrCodeCancelOrderFailed,
	reasonUpdateOrderFailed: ErrCodeUpdateOrderFailed,
	reasonInternal:          ErrCodeInternal,
}

func ErrorCodeFromReason(reason string) int32 {
	if code, ok := reasonToCode[reason]; ok {
		return code
	}
	return ErrCodeInternal
}
