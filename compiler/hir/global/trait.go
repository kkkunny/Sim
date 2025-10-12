package global

import (
	"github.com/kkkunny/stl/container/hashmap"
	stlslices "github.com/kkkunny/stl/container/slices"
	stlval "github.com/kkkunny/stl/value"

	"github.com/kkkunny/Sim/compiler/hir"
	"github.com/kkkunny/Sim/compiler/hir/local"
	"github.com/kkkunny/Sim/compiler/hir/types"
	"github.com/kkkunny/Sim/compiler/hir/values"
)

type traitCovertInfo struct {
	checkFn  func(t hir.Type) bool
	covertFn func(self hir.Value, args ...hir.Value) hir.Value
}

var traitCovertMap = map[string]traitCovertInfo{
	"And": {
		checkFn: func(t hir.Type) bool {
			return types.Is[types.IntType](t)
		},
		covertFn: func(self hir.Value, args ...hir.Value) hir.Value {
			return local.NewAndExpr(self, stlslices.First(args))
		},
	},
	"Or": {
		checkFn: func(t hir.Type) bool {
			return types.Is[types.IntType](t)
		},
		covertFn: func(self hir.Value, args ...hir.Value) hir.Value {
			return local.NewOrExpr(self, stlslices.First(args))
		},
	},
	"Xor": {
		checkFn: func(t hir.Type) bool {
			return types.Is[types.IntType](t)
		},
		covertFn: func(self hir.Value, args ...hir.Value) hir.Value {
			return local.NewXorExpr(self, stlslices.First(args))
		},
	},
	"Shl": {
		checkFn: func(t hir.Type) bool {
			return types.Is[types.IntType](t)
		},
		covertFn: func(self hir.Value, args ...hir.Value) hir.Value {
			return local.NewShlExpr(self, stlslices.First(args))
		},
	},
	"Shr": {
		checkFn: func(t hir.Type) bool {
			return types.Is[types.IntType](t)
		},
		covertFn: func(self hir.Value, args ...hir.Value) hir.Value {
			return local.NewShrExpr(self, stlslices.First(args))
		},
	},
	"Add": {
		checkFn: func(t hir.Type) bool {
			return types.Is[types.NumType](t)
		},
		covertFn: func(self hir.Value, args ...hir.Value) hir.Value {
			return local.NewAddExpr(self, stlslices.First(args))
		},
	},
	"Sub": {
		checkFn: func(t hir.Type) bool {
			return types.Is[types.NumType](t)
		},
		covertFn: func(self hir.Value, args ...hir.Value) hir.Value {
			return local.NewSubExpr(self, stlslices.First(args))
		},
	},
	"Mul": {
		checkFn: func(t hir.Type) bool {
			return types.Is[types.NumType](t)
		},
		covertFn: func(self hir.Value, args ...hir.Value) hir.Value {
			return local.NewMulExpr(self, stlslices.First(args))
		},
	},
	"Div": {
		checkFn: func(t hir.Type) bool {
			return types.Is[types.NumType](t)
		},
		covertFn: func(self hir.Value, args ...hir.Value) hir.Value {
			return local.NewDivExpr(self, stlslices.First(args))
		},
	},
	"Rem": {
		checkFn: func(t hir.Type) bool {
			return types.Is[types.NumType](t)
		},
		covertFn: func(self hir.Value, args ...hir.Value) hir.Value {
			return local.NewRemExpr(self, stlslices.First(args))
		},
	},
	"Lt": {
		checkFn: func(t hir.Type) bool {
			return types.Is[types.NumType](t)
		},
		covertFn: func(self hir.Value, args ...hir.Value) hir.Value {
			return local.NewLtExpr(self, stlslices.First(args))
		},
	},
	"Gt": {
		checkFn: func(t hir.Type) bool {
			return types.Is[types.NumType](t)
		},
		covertFn: func(self hir.Value, args ...hir.Value) hir.Value {
			return local.NewGtExpr(self, stlslices.First(args))
		},
	},
	"Le": {
		checkFn: func(t hir.Type) bool {
			return types.Is[types.NumType](t)
		},
		covertFn: func(self hir.Value, args ...hir.Value) hir.Value {
			return local.NewLeExpr(self, stlslices.First(args))
		},
	},
	"Ge": {
		checkFn: func(t hir.Type) bool {
			return types.Is[types.NumType](t)
		},
		covertFn: func(self hir.Value, args ...hir.Value) hir.Value {
			return local.NewGeExpr(self, stlslices.First(args))
		},
	},
	"Land": {
		checkFn: func(t hir.Type) bool {
			return types.Is[types.BoolType](t)
		},
		covertFn: func(self hir.Value, args ...hir.Value) hir.Value {
			return local.NewLogicAndExpr(self, stlslices.First(args))
		},
	},
	"Lor": {
		checkFn: func(t hir.Type) bool {
			return types.Is[types.BoolType](t)
		},
		covertFn: func(self hir.Value, args ...hir.Value) hir.Value {
			return local.NewLogicOrExpr(self, stlslices.First(args))
		},
	},
	"Eq": {
		checkFn: func(t hir.Type) bool {
			return true
		},
		covertFn: func(self hir.Value, args ...hir.Value) hir.Value {
			if st := self.Type(); types.Is[types.NoThingType](st, true) || types.Is[types.NoReturnType](st, true) {
				return values.NewBoolean(true)
			} else {
				return local.NewEqExpr(self, stlslices.First(args))
			}
		},
	},
	"Neg": {
		checkFn: func(t hir.Type) bool {
			return types.Is[types.NumType](t)
		},
		covertFn: func(self hir.Value, args ...hir.Value) hir.Value {
			return local.NewSubExpr(
				local.NewDefaultExpr(self.Type()),
				self,
			)
		},
	},
	"Not": {
		checkFn: func(t hir.Type) bool {
			return types.Is[types.IntType](t) || types.Is[types.BoolType](t)
		},
		covertFn: func(self hir.Value, args ...hir.Value) hir.Value {
			return local.NewNotExpr(self)
		},
	},
	"Drop": {
		checkFn: func(t hir.Type) bool {
			return true
		},
		covertFn: func(self hir.Value, args ...hir.Value) hir.Value {
			return local.NewDropExpr(self)
		},
	},
}

type Trait struct {
	pkgGlobalAttr
	hir.GenericRestraint
	name    string
	methods hashmap.HashMap[string, *FuncDecl]
}

func NewTrait(name string) *Trait {
	return &Trait{
		name:    name,
		methods: hashmap.StdWith[string, *FuncDecl](),
	}
}

func (self *Trait) GetName() (string, bool) {
	return self.name, self.name != "_"
}

func (self *Trait) AddMethod(m *FuncDecl) bool {
	mn, ok := m.GetName()
	if !ok {
		return true
	} else if self.methods.Contain(mn) {
		return false
	}
	self.methods.Set(mn, m)
	return true
}

func (self *Trait) Methods() []*FuncDecl {
	return self.methods.Values()
}

func (self *Trait) FirstMethod() (*FuncDecl, bool) {
	return stlslices.First(self.methods.Values()), !self.methods.Empty()
}

func (self *Trait) GetMethod(name string) (*FuncDecl, bool) {
	f := self.methods.Get(name)
	return f, f != nil
}

func (self *Trait) GetMethodType(name string) (hir.Type, bool) {
	f, ok := self.GetMethod(name)
	if !ok {
		return nil, false
	}
	return f.Type(), true
}

func (self *Trait) ContainMethod(name string) bool {
	return self.methods.Contain(name)
}

func (self *Trait) GetCovertValue(selfValue hir.Value, args ...hir.Value) (hir.Value, bool) {
	info, ok := traitCovertMap[self.name]
	if !self.Package().IsBuildIn() || !ok {
		return nil, false
	}
	return info.covertFn(selfValue, args...), true
}

func (self *Trait) HasBeImpled(t hir.Type, noBuildin ...bool) (ok bool) {
	defer func() {
		if !ok && self.Package().IsBuildIn() && !stlslices.Last(noBuildin) {
			info, exist := traitCovertMap[self.name]
			if !exist {
				return
			}
			ok = info.checkFn(t)
		}
	}()

	if ct, ok := types.As[CustomTypeDef](t, true); ok {
		return stlslices.All(self.Methods(), func(_ int, dstF *FuncDecl) bool {
			method, ok := ct.GetMethod(stlval.IgnoreWith(dstF.GetName()))
			if !ok {
				return false
			}
			return method.Type().Equal(types.ReplaceVirtualType(hashmap.AnyWith[types.VirtualType, hir.Type](types.Self, t), dstF.Type()))
		})
	} else if gpt, ok := types.As[types.GenericParamType](t, true); ok {
		restraint, ok := gpt.Restraint()
		if !ok {
			return false
		}
		return self == restraint
	}
	return false
}
