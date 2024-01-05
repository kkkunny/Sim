package traits

import (
	"github.com/kkkunny/stl/container/hashmap"
	stliter "github.com/kkkunny/stl/container/iter"
	"github.com/kkkunny/stl/container/pair"

	"github.com/kkkunny/Sim/runtime/types"
)

// Trait 特性
type Trait struct {
	Name string
	Methods hashmap.HashMap[string, *types.FuncType]
}

func newTrait(name string, methods map[string]*types.FuncType)*Trait{
	newMethods := hashmap.NewHashMapWithCapacity[string, *types.FuncType](uint(len(methods)))
	for k, v := range methods{
		newMethods.Set(k, v)
	}
	return &Trait{
		Name: name,
		Methods: newMethods,
	}
}

func (self Trait) copyAndReplaceSelfType(selfType types.Type)*Trait{
	return &Trait{
		Name: self.Name,
		Methods: stliter.Map[pair.Pair[string, *types.FuncType], pair.Pair[string, *types.FuncType], hashmap.HashMap[string, *types.FuncType]](self.Methods, func(p pair.Pair[string, *types.FuncType]) pair.Pair[string, *types.FuncType] {
			return pair.NewPair(p.First, replaceSelfType(p.Second, selfType).(*types.FuncType))
		}),
	}
}

// IsInst 是否是实例
func (self Trait) IsInst(st *types.StructType)bool{
	methodInst := hashmap.NewHashMapWithCapacity[string, *types.FuncType](self.Methods.Length())
	for _, method := range st.Methods{
		if self.Methods.ContainKey(method.Name){
			methodInst.Set(method.Name, method.Type)
		}
	}
	if methodInst.Length() != self.Methods.Length(){
		return false
	}
	for iter:=methodInst.Iterator(); iter.Next(); {
		if !iter.Value().Second.Equal(self.Methods.Get(iter.Value().First)){
			return false
		}
	}
	return true
}
