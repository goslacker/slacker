package worker

import (
	"context"
	"time"

	"github.com/goslacker/slacker/core/app"
)

func NewComponent() *Component {
	return &Component{}
}

type Component struct {
	app.Component
	*Manager
	ctx    context.Context
	cancel context.CancelFunc
}

func (c *Component) Init() error {
	return app.Bind[*Manager](func() *Manager {
		c.Manager = NewManager()
		return c.Manager
	})
}

// Start 启动服务并阻塞, 框架一般会将这个方法作为协程调用, 报错应打日志记录
func (c *Component) Start() {
	time.Sleep(time.Second * 5)
	c.ctx, c.cancel = context.WithCancel(context.Background())
	c.Manager.Start(c.ctx)
}

// Stop 停止服务并阻塞, 报错应打日志记录
func (c *Component) Stop() {
	c.cancel()
}
