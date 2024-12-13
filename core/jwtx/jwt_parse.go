package jwtx

import (
	"github.com/golang-jwt/jwt/v5"
	"github.com/goslacker/slacker/core/errx"
	"net/http"
	"strings"
)

func AuthToken(request *http.Request, headerKey string, queryKey string, salt string, check func(claims jwt.MapClaims) error) (claims jwt.MapClaims, err error) {
	var token string
	if t := request.Header.Get(headerKey); t != "" {
		arr := strings.Split(t, " ")
		if len(arr) < 2 {
			err = errx.New("token格式错误")
			return
		}
		token = arr[1]
	} else if t := request.URL.Query().Get(queryKey); t != "" {
		token = t
	} else {
		err = errx.New("无token")
		return
	}

	t, err := Parse(token, salt)
	if err != nil {
		err = errx.Wrap(err)
		return
	}
	if !t.Valid {
		err = errx.New("无效的token")
		return
	}

	claims, ok := t.Claims.(jwt.MapClaims)
	if !ok {
		err = errx.New("解析token失败")
		return
	}

	if check != nil {
		if err = check(claims); err != nil {
			err = errx.Wrap(err)
			return
		}
	}

	return
}
