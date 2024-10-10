package jwtx

import (
	"context"
	"fmt"
	"github.com/golang-jwt/jwt/v5"
	"github.com/goslacker/slacker/core/tool/convert"
)

const (
	JwtID        = "jti"
	JwtIssuer    = "iss"
	JwtSubject   = "sub"
	JwtAudience  = "aud"
	JwtExpiresAt = "exp"
	JwtNotBefore = "nbf"
	JwtIssuedAt  = "iat"
)

func NewJwtTokenBuilder() *jwtTokenBuilder {
	return &jwtTokenBuilder{
		method: jwt.SigningMethodHS256,
	}
}

type jwtTokenBuilder struct {
	jwt.MapClaims
	key    []byte
	method jwt.SigningMethod
}

func (j *jwtTokenBuilder) WithClaim(key string, value interface{}) *jwtTokenBuilder {
	if j.MapClaims == nil {
		j.MapClaims = make(jwt.MapClaims)
	}
	j.MapClaims[key] = value
	return j
}

func (j *jwtTokenBuilder) WithClaims(claims map[string]interface{}) *jwtTokenBuilder {
	if j.MapClaims == nil {
		j.MapClaims = make(jwt.MapClaims)
	}
	for key, value := range claims {
		j.MapClaims[key] = value
	}
	return j
}

func (j *jwtTokenBuilder) WithKey(key string) *jwtTokenBuilder {
	j.key = []byte(key)
	return j
}

func (j *jwtTokenBuilder) WithMethod(method jwt.SigningMethod) *jwtTokenBuilder {
	j.method = method
	return j
}

func (j jwtTokenBuilder) BuildToken() (string, error) {
	t := jwt.NewWithClaims(j.method, j.MapClaims)
	return t.SignedString(j.key)
}

// Parse token and fill claims
func Parse(token string, key string) (t *jwt.Token, err error) {
	t, err = jwt.Parse(token, func(token *jwt.Token) (any, error) {
		return []byte(key), nil
	})

	return
}

type claimsKey string

const ClaimsKey claimsKey = "claims"

func NewContextWithClaims(ctx context.Context, claims jwt.Claims) context.Context {
	return context.WithValue(ctx, ClaimsKey, claims)
}

func FieldFromContext[T any](ctx context.Context, field string) (result T, err error) {
	tmp := ctx.Value(ClaimsKey)
	if tmp == nil {
		err = fmt.Errorf("claims not found")
		return
	}
	claims := tmp.(jwt.MapClaims)
	v := claims[field]
	if v == nil {
		err = fmt.Errorf("field %s not found", field)
		return
	}
	return convert.To[T](v)
}
