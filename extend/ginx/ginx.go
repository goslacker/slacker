package ginx

import (
	"context"
	"errors"
	"github.com/gin-gonic/gin"
	"github.com/goslacker/slacker/app"
	"github.com/goslacker/slacker/extend/ginx/middleware"
	"github.com/goslacker/slacker/extend/slicex"
	"github.com/spf13/viper"
	"log/slog"
	"net/http"
	"time"
)

func NewGinx() *Ginx {
	return &Ginx{}
}

type Ginx struct {
	app.IsComponent
	router gin.IRouter
	svr    *http.Server
}

func (g *Ginx) Start() {
	c := viper.Sub("ginx")

	g.svr = &http.Server{
		Handler: g.router.(http.Handler),
	}
	g.svr.Addr = c.GetString("addr")

	var err error
	if c.GetBool("lts") {
		err = g.svr.ListenAndServeTLS(c.GetString("certFile"), c.GetString("keyFile"))
	} else {
		err = g.svr.ListenAndServe()
	}
	if err != nil {
		slog.Info("server shutdown", "error", err)
	}
}

func (g *Ginx) Stop() {
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	_ = g.svr.Shutdown(ctx)
}

func (g *Ginx) Init() (err error) {
	c, err := app.Resolve[*viper.Viper]()
	if err != nil {
		return
	}
	c = c.Sub("ginx")
	if c == nil {
		return errors.New("ginx init failed: no config found")
	}
	g.router = gin.Default()
	if c.GetBool("cors") {
		g.router.Use(middleware.CORS)
	}
	g.router.Use(middleware.Options)

	err = app.Bind[Router](g)
	if err != nil {
		return
	}

	return
}

func (g *Ginx) Use(a ...any) Router {
	g.router.Use(convertHandlers(a)...)
	return g
}

func (g *Ginx) Handle(s string, s2 string, a ...any) Router {
	g.router.Handle(s, s2, convertHandlers(a)...)
	return g
}

func (g *Ginx) Any(s string, a ...any) Router {
	g.router.Any(s, convertHandlers(a)...)
	return g
}

func (g *Ginx) GET(s string, a ...any) Router {
	g.router.GET(s, convertHandlers(a)...)
	return g
}

func (g *Ginx) POST(s string, a ...any) Router {
	g.router.POST(s, convertHandlers(a)...)
	return g
}

func (g *Ginx) DELETE(s string, a ...any) Router {
	g.router.DELETE(s, convertHandlers(a)...)
	return g
}

func (g *Ginx) PATCH(s string, a ...any) Router {
	g.router.PATCH(s, convertHandlers(a)...)
	return g
}

func (g *Ginx) PUT(s string, a ...any) Router {
	g.router.PUT(s, convertHandlers(a)...)
	return g
}

func (g *Ginx) OPTIONS(s string, a ...any) Router {
	g.router.OPTIONS(s, convertHandlers(a)...)
	return g
}

func (g *Ginx) HEAD(s string, a ...any) Router {
	g.router.HEAD(s, convertHandlers(a)...)
	return g
}

func (g *Ginx) Match(strings []string, s string, a ...any) Router {
	g.router.Match(strings, s, convertHandlers(a)...)
	return g
}

func (g *Ginx) Group(s string, a ...any) Router {
	return &Ginx{
		router: g.router.Group(s, convertHandlers(a)...),
	}
}

func (g *Ginx) StaticFile(s string, s2 string) Router {
	g.router.StaticFile(s, s2)
	return g
}

func (g *Ginx) StaticFileFS(s string, s2 string, system http.FileSystem) Router {
	g.router.StaticFileFS(s, s2, system)
	return g
}

func (g *Ginx) Static(s string, s2 string) Router {
	g.router.Static(s, s2)
	return g
}

func (g *Ginx) StaticFS(s string, system http.FileSystem) Router {
	g.router.StaticFS(s, system)
	return g
}

func convertHandlers(a []any) []gin.HandlerFunc {
	handlers := make([]gin.HandlerFunc, 0, len(a))
	handlers = append(handlers, slicex.Map(a[:len(a)-1], func(item any) gin.HandlerFunc {
		return WrapMiddleware(item)
	})...)
	handlers = append(handlers, WrapEndpoint(a[len(a)-1]))
	return handlers
}
