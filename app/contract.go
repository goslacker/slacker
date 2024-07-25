package app

// Component 组件。
type Component interface {
	_component()
}

// Initable 表示模块需要被初始化
type Initable interface {
	Init() error
}

// Bootable 表示模块需要被启动
type Bootable interface {
	Boot() error
}

type Serviceable interface {
	//Start 启动服务并阻塞, 框架一般会将这个方法作为协程调用, 报错应打日志记录
	Start()
	//Stop 停止服务并阻塞, 报错应打日志记录
	Stop()
}
