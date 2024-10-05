package ginx

import (
	"errors"
	"fmt"
	"github.com/gin-gonic/gin"
	"github.com/gin-gonic/gin/binding"
	"github.com/golang-jwt/jwt/v5"
	"github.com/goslacker/slacker/app"
	"github.com/goslacker/slacker/container"
	"github.com/goslacker/slacker/extend/reflectx"
	"net/http"
	"reflect"
)

type handlerOpts struct {
	NotAbortWhenErr bool
}

func WithDisableAbortWhenErr() func(opts *handlerOpts) {
	return func(opts *handlerOpts) {
		opts.NotAbortWhenErr = true
	}
}

func WrapMiddleware(f any, opts ...func(*handlerOpts)) gin.HandlerFunc {
	opt := &handlerOpts{}
	for _, f := range opts {
		f(opt)
	}

	fType := reflect.TypeOf(f)
	if fType.Kind() != reflect.Func {
		panic(errors.New("func only"))
	}

	return func(ctx *gin.Context) {
		params, err := buildParams(fType, ctx)
		if err != nil {
			response := &ErrorJsonResponse{
				Message:    fmt.Errorf("build params failed: %w", err).Error(),
				StatusCode: http.StatusInternalServerError,
			}
			if opt.NotAbortWhenErr {
				response.Abort = true
			}
			response.Do(ctx)
			return
		}

		results := reflect.ValueOf(f).Call(params)

		response := fromResults(results)
		if _, ok := response.(*ErrorJsonResponse); ok {
			response.Do(ctx)
			return
		}

		ctx.Next()
		return
	}
}

func WrapEndpoint(f any, opts ...func(*handlerOpts)) gin.HandlerFunc {
	opt := &handlerOpts{}
	for _, f := range opts {
		f(opt)
	}

	fType := reflect.TypeOf(f)
	if fType.Kind() != reflect.Func {
		panic(errors.New("func only"))
	}

	return func(ctx *gin.Context) {
		params, err := buildParams(fType, ctx)
		if err != nil {
			response := &ErrorJsonResponse{
				Message:    fmt.Errorf("build params failed: %w", err).Error(),
				StatusCode: http.StatusInternalServerError,
			}
			if opt.NotAbortWhenErr {
				response.Abort = true
			}
			response.Do(ctx)
			return
		}

		results := reflect.ValueOf(f).Call(params)

		response := fromResults(results)
		response.Do(ctx)
	}
}

func buildParams(fType reflect.Type, ctx *gin.Context) (params []reflect.Value, err error) {
	params = make([]reflect.Value, fType.NumIn())
	valueCtx := reflect.ValueOf(ctx)
	for i := 0; i < fType.NumIn(); i++ {
		switch fType.In(i) {
		case valueCtx.Type():
			params[i] = valueCtx
		case reflect.TypeOf((*Page)(nil)):
			params[i] = reflect.ValueOf(NewPageFromCtx(ctx, 15))
		case reflect.TypeOf((*Filters)(nil)):
			params[i] = reflect.ValueOf(NewFiltersFromCtx(ctx))
		case reflect.TypeOf((*Sorts)(nil)):
			params[i] = reflect.ValueOf(NewSortsFromCtx(ctx))
		case reflect.TypeOf((jwt.MapClaims)(nil)):
			claims, ok := ctx.Get("claims")
			if ok {
				params[i] = reflect.ValueOf(claims)
			} else {
				params[i] = reflect.ValueOf((jwt.MapClaims)(nil))
			}
		default:
			params[i], err = parseParam(ctx, fType.In(i))
			if err != nil {
				return
			}
		}
	}

	return
}

func parseParam(ctx *gin.Context, t reflect.Type) (p reflect.Value, err error) {
	p, err = app.GetContainer().Resolve(t, "")

	if err != nil && !errors.Is(err, container.ErrNotFound) {
		return
	}

	if err != nil || !p.IsValid() {
		psrc := reflectx.NewIndirectType(t)

		if len(ctx.Request.URL.Query()) > 0 {
			_ = ctx.ShouldBindQuery(psrc.Interface())
		}
		if len(ctx.Params) > 0 {
			_ = ctx.ShouldBindUri(psrc.Interface())
		}
		_ = ctx.ShouldBind(psrc.Interface())

		err = binding.Validator.ValidateStruct(psrc.Interface())
		if err != nil {
			return
		}

		pp := reflect.New(t)
		err = reflectx.SetValue(pp, psrc)
		if err != nil {
			return
		}
		p = pp.Elem()
	}

	return
}
