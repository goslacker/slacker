package ginx

import (
	"github.com/gin-gonic/gin"
	"github.com/goslacker/slacker/core/slicex"
	"github.com/goslacker/slacker/core/urlx"
	"github.com/stoewer/go-strcase"
)

func NewPageFromCtx(ctx *gin.Context, defaultSize int) *Page {
	m := &Page{}
	ctx.ShouldBindQuery(m)
	if m.Page == 0 {
		m.Page = 1
	}
	if m.Size == 0 {
		m.Size = defaultSize
	}
	return m
}

type Page struct {
	Page int `form:"page"`
	Size int `form:"size"`
}

func NewFiltersFromCtx(ctx *gin.Context) *Filters {
	tmp, _ := urlx.ParseQuery(ctx.Request.URL.Query(), "filters")
	for k, v := range tmp {
		delete(tmp, k)
		var t []string
		for _, item := range v {
			if item != "" {
				t = append(t, item)
			}
		}
		if len(t) > 0 {
			tmp[strcase.SnakeCase(k)] = t
		}
	}
	return &Filters{
		M: tmp,
	}
}

type Filters struct {
	M map[string][]string
}

func NewSortsFromCtx(ctx *gin.Context) *Sorts {
	tmp := ctx.QueryMap("sorts")
	for k, v := range tmp {
		delete(tmp, k)
		if v != "" && slicex.Contains(v, []string{"desc", "DESC", "asc", "ASC"}) {
			tmp[strcase.SnakeCase(k)] = v
		}
	}
	return &Sorts{
		m: tmp,
	}
}

type Sorts struct {
	m map[string]string
}
