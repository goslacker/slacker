package ginx

import (
	"github.com/gin-gonic/gin"
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
			r := &ErrorJsonResponse{
				Message:    last.Interface().(error).Error(),
				StatusCode: http.StatusInternalServerError,
			}
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
		} else if result.Kind() == reflect.Struct ||
			result.Kind() == reflect.Map ||
			(result.Kind() == reflect.Pointer && result.Elem().Kind() == reflect.Map) ||
			(result.Kind() == reflect.Pointer && result.Elem().Kind() == reflect.Struct) {
			sr.Data = result.Interface()
		} else if result.Kind() == reflect.Int {
			sr.StatusCode = result.Interface().(int)
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

type ErrorJsonResponse struct {
	Message    string `json:"message"`
	StatusCode int    `json:"-"`
	Abort      bool   `json:"-"`
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

//
//type IMeta interface {
//	GetMeta() map[string]any
//}
//
//type Meta struct {
//	CurrentPage uint
//	Total       uint
//	LastPage    uint
//	PerPage     uint
//}
//
//func (m *Meta) GetMeta() (result map[string]any) {
//	result = make(map[string]any)
//	if m.CurrentPage != 0 {
//		result["currentPage"] = m.CurrentPage
//	}
//	result["total"] = m.Total
//	if m.LastPage != 0 {
//		result["lastPage"] = m.LastPage
//	}
//	if m.PerPage != 0 {
//		result["perPage"] = m.PerPage
//	}
//	return
//}
//
//func WithHttpStatusCode(code int) func(*Response) {
//	return func(response *Response) {
//		response.code = code
//	}
//}
//
//func WithEnableAbort() func(*Response) {
//	return func(response *Response) {
//		response.abort = true
//	}
//}
//
//func NewSuccessResponse(data any, meta IMeta, opts ...func(response *Response)) *Response {
//	r := &Response{
//		Data: data,
//		code: http.StatusOK,
//	}
//	if meta != nil {
//		r.Meta = meta.GetMeta()
//	}
//	for _, opt := range opts {
//		opt(r)
//	}
//	return r
//}
//
//func NewFileResponse(file IFile, opts ...func(response *Response)) *Response {
//	r := &Response{
//		file: file,
//		code: http.StatusOK,
//	}
//	for _, opt := range opts {
//		opt(r)
//	}
//	return r
//}
//
//func NewErrorResponse(err error, opts ...func(response *Response)) *Response {
//	r := &Response{
//		err:  err,
//		code: http.StatusInternalServerError,
//	}
//	for _, opt := range opts {
//		opt(r)
//	}
//	return r
//}
//
//type Response struct {
//	Data    any            `json:"data"`
//	Message string         `json:"message"`
//	Meta    map[string]any `json:"meta,omitempty"`
//	file    IFile
//	meta    IMeta
//	err     error
//	code    int
//	abort   bool
//}
//
//func (r *Response) Do(ctx *gin.Context) {
//	if r.err != nil {
//		r.Data = nil
//		r.Message = r.err.Error()
//		if r.abort {
//			ctx.AbortWithStatusJSON(r.code, r)
//		} else {
//			ctx.JSON(r.code, r)
//		}
//		return
//	}
//
//	if r.file != nil {
//		ctx.Status(http.StatusOK)
//		if r.file.MimeType() != "" {
//			ctx.Writer.Header().Add("Content-Type", r.file.MimeType())
//		} else {
//			ctx.Writer.Header().Add("Content-Type", "application/octet-stream")
//		}
//		if r.file.Name() != "" {
//			ctx.Writer.Header().Add("Content-Disposition", "attachment; filename=\""+r.file.Name()+"\"")
//		}
//		_, _ = ctx.Writer.Write(r.file.Content())
//		return
//	}
//
//	ctx.JSON(r.code, r)
//}
