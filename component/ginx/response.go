package ginx

import (
	"fmt"
	"github.com/gin-gonic/gin"
	"github.com/goslacker/slacker/core/errx"
	"github.com/goslacker/slacker/core/tool"
	"net/http"
	"reflect"
)

type Response interface {
	Do(ctx *gin.Context)
}

type Meta map[string]any

func fromResults(results []reflect.Value) Response {
	//处理空返回
	if len(results) == 0 {
		return &SuccessJsonResponse{
			StatusCode: http.StatusOK,
		}
	}

	//处理error
	last := results[len(results)-1]
	if last.IsValid() && !last.IsZero() {
		if last.Type().Implements(reflect.TypeOf((*error)(nil)).Elem()) {
			r := ResponseFromError(last.Interface().(error))

			for _, result := range results {
				if result.Kind() == reflect.Int {
					r.StatusCode = result.Interface().(int)
				} else if result.Kind() == reflect.Bool {
					r.Abort = result.Interface().(bool)
				}
			}

			return r
		}
	}

	//处理直接返回Response接口
	first := results[0]
	if first.Type().Implements(reflect.TypeOf((*Response)(nil)).Elem()) {
		return first.Interface().(Response)
	}

	//处理返回文件
	if first.Type().Implements(reflect.TypeOf((*File)(nil)).Elem()) {
		r := &FileResponse{
			File:       first.Interface().(File),
			StatusCode: http.StatusOK,
		}
		for _, result := range results {
			if result.Kind() == reflect.Int {
				r.StatusCode = result.Interface().(int)
				break
			}
		}
		return r
	}

	sr := &SuccessJsonResponse{
		StatusCode: http.StatusOK,
	}
	for _, result := range results {
		if !result.IsValid() {
			continue
		}
		if result.Type() == reflect.TypeOf(Meta{}) {
			sr.Meta = result.Interface().(Meta)
		} else if result.Kind() == reflect.Int {
			sr.StatusCode = result.Interface().(int)
		} else {
			if sr.Data == nil {
				sr.Data = result.Interface()
			}
		}
	}
	return sr
}

type FileResponse struct {
	File       File
	StatusCode int
}

func (fr *FileResponse) Do(ctx *gin.Context) {
	if fr.File.MimeType() != "" {
		ctx.Writer.Header().Set("Content-Type", fr.File.MimeType())
	} else {
		ctx.Writer.Header().Set("Content-Type", "application/octet-stream")
	}
	if fr.File.Name() != "" {
		ctx.Writer.Header().Set("Content-Disposition", "attachment; filename=\""+fr.File.Name()+"\"")
	}
	ctx.Status(fr.StatusCode)
	_, err := ctx.Writer.Write(fr.File.Content())
	if err != nil {
		panic(err)
	}
}

type SuccessJsonResponse struct {
	Data       any            `json:"data"`
	Meta       map[string]any `json:"meta,omitempty"`
	StatusCode int            `json:"-"`
}

func (sr *SuccessJsonResponse) Do(ctx *gin.Context) {
	ctx.JSON(sr.StatusCode, sr)
}

func ResponseFromError(err error) *ErrorJsonResponse {
	r := &ErrorJsonResponse{
		StatusCode: http.StatusInternalServerError,
	}
	switch x := err.(type) {
	case *errx.Error:
		r.Message = x.Error()
		if http.StatusText(x.Code) == "" {
			r.Code = tool.Reference(x.Code)
		} else {
			r.StatusCode = x.Code
		}
		if x.Detail != nil && len(x.Detail) > 0 {
			r.Detail = x.Detail
		}
	case error:
		r.Message = x.Error()
	default:
		panic(fmt.Errorf("not a error: %#v", err))
	}

	return r
}

type ErrorJsonResponse struct {
	Message    string         `json:"message"`
	Detail     map[string]any `json:"detail,omitempty"`
	Code       *int           `json:"code,omitempty"`
	StatusCode int            `json:"-"`
	Abort      bool           `json:"-"`
}

func (er *ErrorJsonResponse) Do(ctx *gin.Context) {
	if er.Abort {
		ctx.AbortWithStatusJSON(er.StatusCode, er)
	} else {
		ctx.JSON(er.StatusCode, er)
	}
}

type NormalFile struct {
	name     string
	mimeType string
	content  []byte
}

func (n NormalFile) Name() string {
	return n.name
}

func (n NormalFile) MimeType() string {
	return n.mimeType
}

func (n NormalFile) Content() []byte {
	return n.content
}

type File interface {
	Name() string
	MimeType() string
	Content() []byte
}
