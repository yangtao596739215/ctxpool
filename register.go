package ctxpool

import (
	"sync"
)

var globalCtxObjPools = map[string]*sync.Pool{}

type CtxObjRegisterMeta struct {
	PoolKey  string
	CreateFn func() interface{}
	ResetFn  func(val interface{})
}

var ctxObjMetaMap = map[string]*CtxObjRegisterMeta{}

func getResetFn(poolKey string) func(val interface{}) {
	return ctxObjMetaMap[poolKey].ResetFn
}
