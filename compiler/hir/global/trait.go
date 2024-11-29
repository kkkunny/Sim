package global

import (
	"github.com/kkkunny/stl/container/hashmap"
	stlslices "github.com/kkkunny/stl/container/slices"
	stlval "github.com/kkkunny/stl/value"

	"github.com/kkkunny/Sim/compiler/hir"
	"github.com/kkkunny/Sim/compiler/hir/types"
)

type Trait struct {
	pkgGlobalAttr
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

func (self *Trait) HasBeImpled(t hir.Type) bool {
	if self.Package().IsBuildIn() {
		switch self.name {
		case "And", "Or", "Xor", "Shl", "Shr":
			switch {
			case types.Is[types.IntType](t, true):
				return true
			}
		case "Add", "Sub", "Mul", "Div", "Rem", "Lt", "Gt", "Le", "ge":
			switch {
			case types.Is[types.NumType](t, true):
				return true
			}
		case "Land", "Lor":
			switch {
			case types.Is[types.BoolType](t, true):
				return true
			}
		}
	}

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
