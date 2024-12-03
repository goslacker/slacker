package grpcx

import (
	"context"
	"log/slog"
	"strings"

	"github.com/goslacker/slacker/core/jwtx"
	"github.com/goslacker/slacker/core/tool/convert"

	"github.com/golang-jwt/jwt/v5"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/metadata"
	"google.golang.org/grpc/status"
)

func NewJWTAuth(check func(ctx context.Context, data jwt.MapClaims) error) *JWTAuthTool {
	return &JWTAuthTool{
		check: check,
	}
}

type JWTAuthTool struct {
	check func(ctx context.Context, data jwt.MapClaims) error
}

func (a *JWTAuthTool) FieldU64(ctx context.Context, field string) (u64 uint64, err error) {
	claims, err := a.Auth(ctx)
	if err != nil {
		return
	}
	return convert.To[uint64](claims[field])
}

func (a *JWTAuthTool) Auth(ctx context.Context) (claims jwt.MapClaims, err error) {
	md, _ := metadata.FromIncomingContext(ctx)
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

	return
}
