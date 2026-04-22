package middleware

import (
	"context"
	"strings"

	"order-service/internal/conf"

	"github.com/go-kratos/kratos/v2/metadata"
	"github.com/go-kratos/kratos/v2/middleware"
	"github.com/golang-jwt/jwt/v5"
)

type authKey struct{}

type AuthInfo struct {
	UserID   int64
	Username string
	Role     string
}

func ServerAuth(secret string) middleware.Middleware {
	return func(handler middleware.Handler) middleware.Handler {
		return func(ctx context.Context, req interface{}) (interface{}, error) {
			tokenStr := extractToken(ctx)
			if tokenStr == "" {
				return nil, conf.ErrUnauthorized
			}
			token, err := jwt.Parse(tokenStr, func(t *jwt.Token) (interface{}, error) {
				return []byte(secret), nil
			})
			if err != nil || !token.Valid {
				return nil, conf.ErrUnauthorized
			}
			claims, ok := token.Claims.(jwt.MapClaims)
			if !ok {
				return nil, conf.ErrUnauthorized
			}
			info := &AuthInfo{}
			if uid, ok := claims["user_id"].(float64); ok {
				info.UserID = int64(uid)
			}
			if username, ok := claims["username"].(string); ok {
				info.Username = username
			}
			if role, ok := claims["role"].(string); ok {
				info.Role = role
			}
			ctx = context.WithValue(ctx, authKey{}, info)
			return handler(ctx, req)
		}
	}
}

func GetAuthInfo(ctx context.Context) (*AuthInfo, bool) {
	info, ok := ctx.Value(authKey{}).(*AuthInfo)
	return info, ok
}

func extractToken(ctx context.Context) string {
	if md, ok := metadata.FromServerContext(ctx); ok {
		auth := md.Get("authorization")
		if auth != "" {
			return strings.TrimPrefix(auth, "Bearer ")
		}
	}
	return ""
}
