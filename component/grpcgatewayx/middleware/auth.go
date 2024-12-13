package middleware

import (
	"context"
	"github.com/golang-jwt/jwt/v5"
	"github.com/goslacker/slacker/core/jwtx"
	"github.com/grpc-ecosystem/grpc-gateway/v2/runtime"
	"log/slog"
	"net/http"
)

func NewJwtAuthMiddlewareBuilder() *JwtAuthMiddlewareBuilder {
	return &JwtAuthMiddlewareBuilder{
		header: "Authorization",
		query:  "token",
	}
}

type JwtAuthMiddlewareBuilder struct {
	header string                           //头字段
	query  string                           //querystring字段
	salt   string                           //盐
	check  func(claims jwt.MapClaims) error //判断是否通过
}

func (j *JwtAuthMiddlewareBuilder) SetHeader(header string) *JwtAuthMiddlewareBuilder {
	j.header = header
	return j
}

func (j *JwtAuthMiddlewareBuilder) SetQuery(queryField string) *JwtAuthMiddlewareBuilder {
	j.query = queryField
	return j
}

func (j *JwtAuthMiddlewareBuilder) SetSalt(salt string) *JwtAuthMiddlewareBuilder {
	j.salt = salt
	return j
}

func (j *JwtAuthMiddlewareBuilder) SetUserIdentifier(f func(claims jwt.MapClaims) error) *JwtAuthMiddlewareBuilder {
	j.check = f
	return j
}

func (j *JwtAuthMiddlewareBuilder) Build(next runtime.HandlerFunc) runtime.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request, pathParams map[string]string) {
		claims, err := jwtx.AuthToken(r, j.header, j.query, j.salt, j.check)
		if err != nil {
			slog.Debug("parse auth token failed", "error", err)
			//w.WriteHeader(http.StatusUnauthorized)
			//return
		} else {
			r = r.WithContext(context.WithValue(r.Context(), "claims", claims))
		}
		next(w, r, pathParams)
	}
}
