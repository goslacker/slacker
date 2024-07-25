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
	"github.com/goslacker/slacker/tool/convert"
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

	if fType.NumOut() > 1 {
		panic(errors.New("middleware should return only err"))
	}
	return func(ctx *gin.Context) {
		params, err := buildParams(fType, ctx)
		if err != nil {
			var response *Response
			if opt.NotAbortWhenErr {
				response = NewErrorResponse(fmt.Errorf("build params failed: %w", err))
			} else {
				response = NewErrorResponse(fmt.Errorf("build params failed: %w", err), WithEnableAbort())
			}
			response.Do(ctx)
			return
		}

		results := reflect.ValueOf(f).Call(params)

		if len(results) == 0 || results[0].IsNil() {
			ctx.Next()
			return
		}

		var response *Response
		err = results[0].Interface().(error)
		if opt.NotAbortWhenErr {
			response = NewErrorResponse(fmt.Errorf("build params failed: %w", err))
		} else {
			response = NewErrorResponse(fmt.Errorf("build params failed: %w", err), WithEnableAbort())
		}
		response.Do(ctx)
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

	if fType.NumOut() > 3 {
		panic(errors.New("should not return results more than 3"))
	}

	switch fType.NumOut() {
	case 2:
		if !fType.Out(1).Implements(reflect.TypeOf((*error)(nil)).Elem()) {
			panic(errors.New("second one of results should be error"))
		}
	case 3:
		if !fType.Out(2).Implements(reflect.TypeOf((*error)(nil)).Elem()) {
			panic(errors.New("third one of results should be error"))
		}
		if !isNumber(fType.Out(0)) && fType.Out(0).Name() != "IMeta" {
			panic(errors.New("first one of results should be number(total of records) or IMeta"))
		}
	}

	return func(ctx *gin.Context) {
		params, err := buildParams(fType, ctx)
		if err != nil {
			var response *Response
			if opt.NotAbortWhenErr {
				response = NewErrorResponse(fmt.Errorf("build params failed: %w", err))
			} else {
				response = NewErrorResponse(fmt.Errorf("build params failed: %w", err), WithEnableAbort())
			}
			response.Do(ctx)
			return
		}

		results := reflect.ValueOf(f).Call(params)

		parseToResponse(!opt.NotAbortWhenErr, results...).Do(ctx)
	}
}

func parseToResponse(shouldAbort bool, results ...reflect.Value) (resp *Response) {
	switch len(results) {
	case 0:
		resp = NewSuccessResponse(nil, nil)
	case 1:
		r := results[0].Interface()
		if r == nil {
			resp = NewSuccessResponse(nil, nil)
			break
		}
		switch x := r.(type) {
		case error:
			if shouldAbort {
				resp = NewErrorResponse(x, WithEnableAbort())
			} else {
				resp = NewErrorResponse(x)
			}
		case IFile:
			resp = NewFileResponse(x)
		default:
			resp = NewSuccessResponse(x, nil)
		}
	case 2:
		if err, ok := results[1].Interface().(error); ok && err != nil {
			if shouldAbort {
				resp = NewErrorResponse(err, WithEnableAbort())
			} else {
				resp = NewErrorResponse(err)
			}
			break
		}
		switch x := results[0].Interface().(type) {
		case IFile:
			resp = NewFileResponse(x)
		default:
			resp = NewSuccessResponse(x, nil)
		}
	case 3:
		if err, ok := results[1].Interface().(error); ok && err != nil {
			if shouldAbort {
				resp = NewErrorResponse(err, WithEnableAbort())
			} else {
				resp = NewErrorResponse(err)
			}
			break
		}

		if isNumber(results[0]) {
			tmp, err := convert.ToKindValue(results[0], reflect.Int)
			if err != nil {
				panic(err)
			}
			resp = NewSuccessResponse(results[1].Interface(), &Meta{Total: uint(tmp.Interface().(int))})
		} else if m, ok := results[0].Interface().(IMeta); ok {
			resp = NewSuccessResponse(results[1].Interface(), m)
		} else {
			panic(errors.New("three results only for pagination"))
		}
	}

	return
}

func isNumber(v interface{ Kind() reflect.Kind }) bool {
	switch v.Kind() {
	case reflect.Int, reflect.Int8, reflect.Int16, reflect.Int32,
		reflect.Int64, reflect.Uint, reflect.Uint8, reflect.Uint16,
		reflect.Uint32, reflect.Uint64, reflect.Float32, reflect.Float64:
		return true
	default:
		return false
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
