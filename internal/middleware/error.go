package middleware

import (
	"context"
	"reflect"

	orderv1 "order-service/api/proto/order/v1"
	"order-service/internal/conf"

	"github.com/go-kratos/kratos/v2/errors"
	kratosmiddleware "github.com/go-kratos/kratos/v2/middleware"
)

func ResponseError() kratosmiddleware.Middleware {
	return func(handler kratosmiddleware.Handler) kratosmiddleware.Handler {
		return func(ctx context.Context, req interface{}) (interface{}, error) {
			reply, err := handler(ctx, req)
			if reply == nil {
				return reply, err
			}

			if setStatusByReflection(reply, err) {
				return reply, nil
			}

			return reply, err
		}
	}
}

func setStatusByReflection(reply interface{}, err error) bool {
	v := reflect.ValueOf(reply)
	if v.Kind() != reflect.Ptr || v.IsNil() {
		return false
	}

	statusField := v.Elem().FieldByName("Status")
	if !statusField.IsValid() || !statusField.CanSet() {
		return false
	}

	if err == nil {
		statusField.Set(reflect.Zero(statusField.Type()))
		return true
	}

	if kratosErr := errors.FromError(err); kratosErr != nil {
		status := &orderv1.ResponseStatus{
			ErrorCode:    conf.ErrorCodeFromReason(kratosErr.Reason),
			ErrorMessage: kratosErr.Message,
		}
		statusField.Set(reflect.ValueOf(status))
		return true
	}

	return false
}
