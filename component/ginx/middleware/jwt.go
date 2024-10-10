package middleware

import (
	"errors"
	"fmt"
	"github.com/gin-gonic/gin"
	"github.com/golang-jwt/jwt/v5"
	"github.com/goslacker/slacker/core/jwtx"
	"net/http"
	"strings"
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
		var token string
		if t := c.Request.Header.Get(j.header); t != "" {
			arr := strings.Split(t, " ")
			if len(arr) < 2 {
				abort = true
				status = http.StatusUnauthorized
				err = errors.New("token格式错误")
				return
			}
			token = arr[1]
		} else if t := c.Query(j.query); t != "" {
			token = t
		} else {
			abort = true
			status = http.StatusUnauthorized
			err = errors.New("无token")
			return
		}

		t, err := jwtx.Parse(token, j.salt)
		if err != nil {
			abort = true
			status = http.StatusUnauthorized
			return
		}
		if !t.Valid {
			abort = true
			status = http.StatusUnauthorized
			err = errors.New("无效的token")
			return
		}

		claims, ok := t.Claims.(jwt.MapClaims)
		if !ok {
			abort = true
			status = http.StatusUnauthorized
			err = errors.New("解析token失败")
			return
		}

		if j.check != nil {
			if err = j.check(claims); err != nil {
				abort = true
				status = http.StatusUnauthorized
				err = fmt.Errorf("未找到用户: %w", err)
				return
			}
		}

		c.Set("claims", claims)

		return
	}
}
