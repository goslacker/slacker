package ginx

import (
	"github.com/gin-gonic/gin"
	"net/http"
)

type IFile interface {
	Name() string
	MimeType() string
	Content() []byte
}

type IMeta interface {
	GetMeta() map[string]any
}

type Meta struct {
	CurrentPage uint
	Total       uint
	LastPage    uint
	PerPage     uint
}

func (m *Meta) GetMeta() (result map[string]any) {
	result = make(map[string]any)
	if m.CurrentPage != 0 {
		result["currentPage"] = m.CurrentPage
	}
	result["total"] = m.Total
	if m.LastPage != 0 {
		result["lastPage"] = m.LastPage
	}
	if m.PerPage != 0 {
		result["perPage"] = m.PerPage
	}
	return
}

func WithHttpStatusCode(code int) func(*Response) {
	return func(response *Response) {
		response.code = code
	}
}

func WithEnableAbort() func(*Response) {
	return func(response *Response) {
		response.abort = true
	}
}

func NewSuccessResponse(data any, meta IMeta, opts ...func(response *Response)) *Response {
	r := &Response{
		Data: data,
		code: http.StatusOK,
	}
	if meta != nil {
		r.Meta = meta.GetMeta()
	}
	for _, opt := range opts {
		opt(r)
	}
	return r
}

func NewFileResponse(file IFile, opts ...func(response *Response)) *Response {
	r := &Response{
		file: file,
		code: http.StatusOK,
	}
	for _, opt := range opts {
		opt(r)
	}
	return r
}

func NewErrorResponse(err error, opts ...func(response *Response)) *Response {
	r := &Response{
		err:  err,
		code: http.StatusInternalServerError,
	}
	for _, opt := range opts {
		opt(r)
	}
	return r
}

type Response struct {
	Data    any            `json:"data"`
	Message string         `json:"message"`
	Meta    map[string]any `json:"meta,omitempty"`
	file    IFile
	meta    IMeta
	err     error
	code    int
	abort   bool
}

func (r *Response) Do(ctx *gin.Context) {
	if r.err != nil {
		r.Data = nil
		r.Message = r.err.Error()
		if r.abort {
			ctx.AbortWithStatusJSON(r.code, r)
		} else {
			ctx.JSON(r.code, r)
		}
		return
	}

	if r.file != nil {
		ctx.Status(http.StatusOK)
		if r.file.MimeType() != "" {
			ctx.Writer.Header().Add("Content-Type", r.file.MimeType())
		} else {
			ctx.Writer.Header().Add("Content-Type", "application/octet-stream")
		}
		if r.file.Name() != "" {
			ctx.Writer.Header().Add("Content-Disposition", "attachment; filename=\""+r.file.Name()+"\"")
		}
		_, _ = ctx.Writer.Write(r.file.Content())
		return
	}

	ctx.JSON(r.code, r)
}
