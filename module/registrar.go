package module

import (
	"sync"
	"gopcpv2-web-spider/errors"
	"fmt"
)

type Registrar interface {
	//注册
	Register(module Module)(bool, error)
	UnRegister(mid MID)(bool, error)
	//获取一个指定类型的组件的实例，该函数基于负载均衡策略
	Get(moduleType Type)(Module, error)
	GetAllByType(moduleType Type)(map[MID]Module, error)
	GetAll()map[MID]Module
	Clear()
}

type myRegistrar struct {
	moduleTypeMap map[Type]map[MID]Module
	rwlock sync.RWMutex
}

func NewRegistrar()Registrar{
	return &myRegistrar{
		moduleTypeMap: map[Type]map[MID]Module{},
	}
}

func(registrar *myRegistrar)Register(module Module)(bool, error){
	if module == nil{
		return false, errors.NewIllegalParameterError("组件实例是nil")
	}
	mid := module.ID()
	parts, err := SplitMID(mid)
	if err != nil{
		return false, err
	}
	moduleType := legalLetterTypeMap[parts[0]]
	if !CheckType(moduleType, module){
		return false, errors.NewIllegalParameterError(fmt.Sprintf("无效的模块类型：%s", moduleType))
	}
	registrar.rwlock.Lock()
	defer registrar.rwlock.Unlock()
	modules := registrar.moduleTypeMap[moduleType]
	if modules == nil{
		modules = map[MID]Module{}
	}
	if _, ok := modules[mid]; ok{
		return false, nil
	}
	modules[mid] = module
	registrar.moduleTypeMap[moduleType] = modules
	return true, nil

}

func (registrar *myRegistrar)UnRegister(mid MID)(bool, error){
	parts, err := SplitMID(mid)
	if err != nil{
		return false, err
	}
	moduleType := legalLetterTypeMap[parts[0]]
	var deleted bool
	registrar.rwlock.Lock()
	defer registrar.rwlock.Unlock()
	if modules, ok := registrar.moduleTypeMap[moduleType]; ok{
		if _, ok := modules[mid]; ok{
			delete(modules, mid)
			deleted = true
		}
	}
	return deleted, nil
}

//用于获取一个指定类型的组件的实例。
//本函数会基于负载均衡策略返回实例。
func (registrar *myRegistrar)Get(moduleType Type)(Module, error){
	modules, err := registrar.GetAllByType(moduleType)
	if err != nil{
		return nil, err
	}
	minScore := uint64(0)
	var selectedModule Module
	for _, module := range modules{
		SetScore(module)
		if err != nil{
			return nil, err
		}
		score := module.Score()
		if minScore == 0 || score < minScore{
			selectedModule = module
			minScore = score
		}
	}
	return selectedModule, nil

}

//用于获取指定类型的所有组件实例。
func (registrar *myRegistrar)GetAllByType(moduleType Type)(map[MID]Module, error){
	if !LegalType(moduleType){
		return nil, errors.NewIllegalParameterError(fmt.Sprintf("无效的模块类型：%s", moduleType))
	}
	registrar.rwlock.RLock()
	defer registrar.rwlock.RUnlock()
	modules := registrar.moduleTypeMap[moduleType]
	if len(modules) == 0{
		return nil, ErrNotFoundModuleInstance
	}
	result := map[MID]Module{}
	for mid, module := range modules{
		result[mid] = module
	}
	return result, nil
}

//用于获取所有组件实例。
func (registrar *myRegistrar)GetAll()map[MID]Module{
	result := map[MID]Module{}
	registrar.rwlock.RLock()
	defer registrar.rwlock.RUnlock()
	for _, modules := range registrar.moduleTypeMap{
		for mid, module := range modules{
			result[mid] = module
		}
	}
	return result
}

//会清除所有的组件注册记录。
func (registrar *myRegistrar)Clear(){
	registrar.rwlock.Lock()
	defer registrar.rwlock.Unlock()
	registrar.moduleTypeMap = map[Type]map[MID]Module{}
}