package module

type Registrat interface {
	//注册
	Register(module Module)(bool, error)
	UnRegister(mid MID)(bool, error)
	//获取一个指定类型的组件的实例，该函数基于负载均衡策略
	Get(moduleType Type)(Module, error)
	GetAllByType(moduleType Type)(map[MID]Module, error)
	GetAll()map[MID]Module
	Clear()
}