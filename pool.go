package ctxpool

import (
	"context"
	"log"
	"sync"
	"time"
)

type BeforeCall func(context.Context, interface{}) (context.Context, error)

// AfterCall functions executed after api handler called.
type AfterCall func(context.Context, interface{}, error) context.Context

func AddHook(name string, hook interface{}) {

}

func init() {
	AddHook("server", BeforeCall(injectCtxPool))
	AddHook("server", AfterCall(resetCtxPool))
}

type ResetMeta struct {
	val     interface{}
	resetFn func(val interface{}) //重制函数
	pool    *sync.Pool
}

func (r *ResetMeta) reset() {
	r.val = nil
	r.resetFn = nil
	r.pool = nil
	resetMetaPool.Put(r)
}

var objFactoryPool = &sync.Pool{
	New: func() interface{} {
		return &ObjFactory{}
	},
}

var resetMetaPool = sync.Pool{
	New: func() interface{} {
		return &ResetMeta{}
	},
}

//ObjFactory 对象工厂，每个请求一个
type ObjFactory struct {
	mu         sync.Mutex
	resetMetas []*ResetMeta
}

func (of *ObjFactory) GetObj(poolKey string) interface{} {
	return of.getObj(
		poolKey,
	)
}

func (of *ObjFactory) getObj(poolKey string) interface{} {
	pool, ok := globalCtxObjPools[poolKey]
	if !ok {
		panic("get unRegistered obj")
	}
	res := pool.Get()
	resetMeta := resetMetaPool.Get().(*ResetMeta)
	resetMeta.resetFn = func(val interface{}) {
		getResetFn(poolKey)(val)
	}
	resetMeta.val = res
	resetMeta.pool = pool
	of.mu.Lock()
	of.resetMetas = append(of.resetMetas, resetMeta)
	of.mu.Unlock()
	return res
}

func (of *ObjFactory) resetToPool() {
	of.mu.Lock()
	of.mu.Unlock()
	for _, v := range of.resetMetas {
		//执行reset逻辑
		v.resetFn(v.val)
		//放回pool
		v.pool.Put(v.val)
		v.reset()
	}
	of.resetMetas = nil
	objFactoryPool.Put(of)
}

var objFactoryKey = struct {
}{}

func injectCtxPool(ctx context.Context, req interface{}) (context.Context, error) {
	return context.WithValue(ctx, objFactoryKey, objFactoryPool.Get()), nil
}

func resetCtxPool(ctx context.Context, req interface{}, err error) context.Context {
	//rpc框架没提供数据写完后的回调，所以只能检测ctx是否Done了，为了不阻塞现场，只能异步，可以考虑用goroutine-pool来优化，其实这样也够用，新版本的go在创建g的时候做了优化
	go func() {
		select {
		case <-ctx.Done():
			factory := ctx.Value(objFactoryKey).(*ObjFactory)
			factory.resetToPool()
		case <-time.After(20 * time.Second):
			log.Printf("resetCtxPool wait for 20s")
		}
	}()
	return ctx
}
