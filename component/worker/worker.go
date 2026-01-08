package worker

import (
	"context"
	"fmt"
	"log/slog"
	"runtime/debug"
	"sync"
	"time"
	"unsafe"
)

func NewManager() *Manager {
	return &Manager{
		workers: make(map[string]*worker),
	}
}

type Worker func(ctx context.Context)

type worker struct {
	f      Worker
	ctx    context.Context
	cancel context.CancelFunc
	keep   bool
}

type Manager struct {
	workers map[string]*worker
	lock    sync.Mutex
	ctx     context.Context
	cancel  context.CancelFunc
}

type option struct {
	Name string
	Keep bool
}

type Opt func(opt *option)

func WithName(name string) Opt {
	return func(opt *option) {
		opt.Name = name
	}
}

func WithNotKeepAlive() Opt {
	return func(opt *option) {
		opt.Keep = false
	}
}

func (m *Manager) Register(w Worker, opts ...Opt) {
	opt := &option{
		Keep: true,
	}
	for _, o := range opts {
		o(opt)
	}

	if opt.Name == "" {
		opt.Name = fmt.Sprintf("%x", unsafe.Pointer(&w))
	}

	m.lock.Lock()
	defer m.lock.Unlock()
	m.workers[opt.Name] = &worker{
		f:    w,
		keep: opt.Keep,
	}
}

func (m *Manager) runWorker(name string, w *worker) {
	defer func() {
		if r := recover(); r != nil {
			switch x := r.(type) {
			case interface{ String() string }:
				slog.Warn("worker panic recover", "error", x.String(), "name", name)
			default:
				slog.Warn("worker panic recover", "error", r, "name", name, "stack", string(debug.Stack()))
			}
			if w.keep {
				m.lock.Lock()
				w.ctx = nil
				w.cancel = nil
				m.lock.Unlock()
				return
			}
		}
		m.unregister(name)
	}()

	w.f(w.ctx)
}

func (m *Manager) stopWorker(name string) {
	m.lock.Lock()
	defer m.lock.Unlock()
	if w, ok := m.workers[name]; ok {
		w.cancel()
	}
	return
}

func (m *Manager) unregister(name string) {
	m.lock.Lock()
	defer m.lock.Unlock()
	mp := make(map[string]*worker, len(m.workers))
	for k, v := range m.workers {
		if name != k {
			mp[k] = v
		}
	}
	m.workers = mp
	return
}

func (m *Manager) Start(ctx context.Context) {
	if ctx == nil {
		ctx = context.Background()
	}
	m.ctx, m.cancel = context.WithCancel(ctx)
	ticker := time.NewTicker(time.Second * 5)
	for range ticker.C {
		select {
		case <-m.ctx.Done():
			return
		default:
		}

		m.lock.Lock()
		for name, w := range m.workers {
			if w.cancel == nil {
				w.ctx, w.cancel = context.WithCancel(m.ctx)
				go m.runWorker(name, w)
			}
		}
		m.lock.Unlock()
	}
}

func (m *Manager) Stop(opts ...Opt) {
	opt := &option{}
	for _, o := range opts {
		o(opt)
	}

	if opt.Name != "" {
		m.stopWorker(opt.Name)
	} else {
		m.cancel()
	}
}
