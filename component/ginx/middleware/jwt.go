package middleware

import (
	"github.com/gin-gonic/gin"
	"github.com/golang-jwt/jwt/v5"
	"github.com/goslacker/slacker/core/jwtx"
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

func (j *JwtAuthMiddlewareBuilder) Build() func(*gin.Context) (bool, int, error) {
	return func(c *gin.Context) (abort bool, status int, err error) {
		claims, err := jwtx.AuthToken(c.Request, j.header, j.query, j.salt, j.check)
		if err != nil {
			abort = true
			status = http.StatusUnauthorized
			return
		}
		c.Set("claims", claims)

		return
	}
}
