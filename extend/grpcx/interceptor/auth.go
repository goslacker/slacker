package interceptor

import (
	"context"
	"log/slog"
	"strings"

	"github.com/golang-jwt/jwt/v5"
	"github.com/goslacker/slacker/extend/jwtx"
	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/metadata"
	"google.golang.org/grpc/status"
)

func NewJWTAuth() *JWTAuth {
	return &JWTAuth{
		whiteList: map[string]struct{}{},
	}
}

type JWTAuth struct {
	whiteList map[string]struct{}
	check     func(ctx context.Context, data jwt.MapClaims) error
}

func (a *JWTAuth) RegisterToWhiteList(whiteList ...string) {
	for _, v := range whiteList {
		a.whiteList[v] = struct{}{}
	}
}

func (a *JWTAuth) SetCheck(check func(ctx context.Context, data jwt.MapClaims) error) {
	a.check = check
}

func (a *JWTAuth) InWhiteList(token string) bool {
	_, ok := a.whiteList[token]
	return ok
}

func (a *JWTAuth) UnaryAuthInterceptor(ctx context.Context, req any, info *grpc.UnaryServerInfo, handler grpc.UnaryHandler) (result any, err error) {
	md, _ := metadata.FromIncomingContext(ctx)

	if !a.InWhiteList(info.FullMethod) {
		var token string
		if t, ok := md["authorization"]; ok {
			token = t[0]
		} else {
			err = status.New(codes.Unauthenticated, "").Err()
			slog.Debug("token not found")
			return
		}

		arr := strings.Split(token, " ")
		if len(arr) < 2 {
			err = status.New(codes.Unauthenticated, "").Err()
			slog.Debug("token format error", "token", token)
			return
		}
		token = arr[1]

		var t *jwt.Token
		t, err = jwtx.Parse(token, "") //TODO: support salt
		if err != nil {
			slog.Debug("parse token failed", "error", err, "token", token)
			err = status.New(codes.Unauthenticated, "").Err()
			return
		}
		if !t.Valid {
			err = status.New(codes.Unauthenticated, "").Err()
			slog.Debug("token is invalid", "token", token)
			return
		}

		claims, ok := t.Claims.(jwt.MapClaims)
		if !ok {
			err = status.New(codes.Unauthenticated, "").Err()
			slog.Debug("claims type error", "claims", claims)
			return
		}

		if a.check != nil {
			if err = a.check(ctx, claims); err != nil {
				slog.Debug("check claims failed", "error", err)
				err = status.New(codes.Unauthenticated, "").Err()
				return
			}
		}

		ctx = context.WithValue(ctx, "claims", claims)
	}

	// 调用被拦截的方法
	return handler(ctx, req)
}
