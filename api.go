package ctxpool

import (
	"context"
	"fmt"
	"sync"
)

//RegisterObjToCtxPool 注册对象，在init方法里面执行，重复注册会panic,最好只注册一次，且不要并发注册，为了考虑性能，不加锁
func RegisterObjToCtxPool(req *CtxObjRegisterMeta) {
	if _, ok := globalCtxObjPools[req.PoolKey]; ok {
		panic(fmt.Sprintf("key:%v already register", req.PoolKey))
	}
	globalCtxObjPools[req.PoolKey] = &sync.Pool{
		New: func() interface{} {
			return req.CreateFn()
		},
	}
	ctxObjMetaMap[req.PoolKey] = req
}

//====================API Call 场景下使用，会在afterCall hook里自动回收对象

// GetObjFromCtxPool 从ctx中获取对象，不同类型的对象用不同的poolKey，poolKey不允许重复
// poolKey Example: "[]uint64_256","[]uint64_512","map[uint64]uint64_1024","ctxpool.CommonObj"
// 注意：会在afterCall hook里回收分配的对象，如果接口在返回后还有其他异步逻辑在执行，需要在完成后调用一下ResetCtxPool方法，否则可能导致内存泄露
func GetObjFromCtxPool(ctx context.Context, poolKey string) interface{} {
	factory := ctx.Value(objFactoryKey).(*ObjFactory)
	if factory == nil {
		return nil
	}
	return factory.GetObj(poolKey)
}

//======================以下两个方法，在非API Call的场景下，提供的ctxPool的使用能力

func ResetCtxPool(ctx context.Context) {
	factory := ctx.Value(objFactoryKey).(*ObjFactory)
	factory.resetToPool()
}

func InjectCtxPool(ctx context.Context) context.Context {
	return context.WithValue(ctx, objFactoryKey, objFactoryPool.Get())
}
