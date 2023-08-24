package ctxpool

import (
	"context"
	"sync"
	"testing"
)

const (
	mapUint64PoolKey1024   = "test_map[uint64]uint64_1024"
	mapUint64PoolKey512    = "test_map[uint64]uint64_512"
	sliceUint64PoolKey1024 = "test_[]uint64_1024"
	commonObjPoolKey       = "test_*CommonObj"
)

var benchInitOnce = &sync.Once{}

func BeforeBenchMark() {
	registerMap()
	registerSlice()
	registerCommonObj()
}

func BenchmarkMapWithPool(b *testing.B) {
	benchInitOnce.Do(func() {
		BeforeBenchMark()
	})
	ctx := context.Background()
	for n := 0; n < b.N; n++ {
		getMapWithPool(ctx, mapUint64PoolKey1024)
	}
}

func BenchmarkMapWithOutPool(b *testing.B) {
	benchInitOnce.Do(func() {
		BeforeBenchMark()
	})
	for i := 0; i < b.N; i++ {
		getMapWithOutPool()
	}
}

func BenchmarkCommonObjWithPool(b *testing.B) {
	benchInitOnce.Do(func() {
		BeforeBenchMark()
	})
	ctx := context.Background()
	for n := 0; n < b.N; n++ {
		getObjWithPool(ctx, commonObjPoolKey)
	}
}

func BenchmarkCommonObjWithOutPool(b *testing.B) {
	benchInitOnce.Do(func() {
		BeforeBenchMark()
	})
	for i := 0; i < b.N; i++ {
		getObjWithOutPool()
	}
}
func BenchmarkSliceWithPool(b *testing.B) {
	benchInitOnce.Do(func() {
		BeforeBenchMark()
	})
	ctx := context.Background()
	for n := 0; n < b.N; n++ {
		getSliceWithPool(ctx, sliceUint64PoolKey1024)

	}
}

func BenchmarkSliceWithOutPool(b *testing.B) {
	benchInitOnce.Do(func() {
		BeforeBenchMark()
	})
	for i := 0; i < b.N; i++ {
		getSliceWithOutPool()
	}
}

func getMapWithPool(ctx context.Context, poolKey string) {
	ctx, _ = injectCtxPool(ctx, nil)
	m := GetObjFromCtxPool(ctx, poolKey)
	res := m.(map[uint64]uint64)
	fillMap(res)
	ctx = resetCtxPool(ctx, nil, nil)
}

func getMapWithOutPool() {
	resMap := make(map[uint64]uint64, 1000)
	fillMap(resMap)
}

func fillMap(m map[uint64]uint64) {
	for i := 0; i < 1000; i++ {
		i := uint64(i)
		m[i] = i
	}
}

func fillSlice(m []uint64) {
	for i := 0; i < 1000; i++ {
		m = append(m, uint64(i))
	}
}

func fillCommonObj(obj *CommonObj) {
	obj.Name = "test"
	obj.Age = 10
	obj.data = []byte("xx")
}

func getSliceWithPool(ctx context.Context, poolKey string) {
	ctx, _ = injectCtxPool(ctx, nil)
	m := GetObjFromCtxPool(ctx, poolKey)
	res := m.([]uint64)
	fillSlice(res)
	resetCtxPool(ctx, nil, nil)

}

func getSliceWithOutPool() {
	resMap := make([]uint64, 0, 1000)
	fillSlice(resMap)
}

type CommonObj struct {
	Name string
	Age  int
	data []byte
}

func getObjWithPool(ctx context.Context, poolKey string) {
	ctx, _ = injectCtxPool(context.Background(), nil)
	m := GetObjFromCtxPool(ctx, poolKey)
	res := m.(*CommonObj)
	fillCommonObj(res)
	ctx = resetCtxPool(ctx, nil, nil)
}

func getObjWithOutPool() {
	for i := 0; i < 100000; i++ {
		obj := &CommonObj{}
		fillCommonObj(obj)
	}
}
func registerMap() {
	RegisterObjToCtxPool(&CtxObjRegisterMeta{
		PoolKey: mapUint64PoolKey1024,
		CreateFn: func() interface{} {
			return make(map[uint64]uint64, 1024)
		},
		ResetFn: func(res interface{}) {
			resMap := res.(map[uint64]uint64)
			for k := range resMap {
				resMap[k] = 0
			}
		},
	})
}
func registerSlice() {
	RegisterObjToCtxPool(&CtxObjRegisterMeta{
		PoolKey: sliceUint64PoolKey1024,
		CreateFn: func() interface{} {
			return make([]uint64, 0, 1000)
		},
		ResetFn: func(val interface{}) {
			resMap := val.([]uint64)
			resMap = resMap[:0]
		},
	})
}

func registerCommonObj() {
	RegisterObjToCtxPool(&CtxObjRegisterMeta{
		PoolKey: commonObjPoolKey,
		CreateFn: func() interface{} {
			return &CommonObj{}
		},
		ResetFn: func(val interface{}) {
			resMap := val.(*CommonObj)
			resMap.Name = ""
			resMap.Age = 0
			resMap.data = nil
		},
	})
}
