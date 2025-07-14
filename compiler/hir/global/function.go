package global

import (
	stlslices "github.com/kkkunny/stl/container/slices"
	stlval "github.com/kkkunny/stl/value"

	"github.com/kkkunny/Sim/compiler/hir"
	"github.com/kkkunny/Sim/compiler/hir/local"
	"github.com/kkkunny/Sim/compiler/hir/types"
)

// FuncDecl 函数声明
type FuncDecl struct {
	typ    types.FuncType
	name   string
	params []*local.Param
}

func NewFuncDecl(typ types.FuncType, name string, params ...*local.Param) *FuncDecl {
	return &FuncDecl{typ: typ, name: name, params: params}
}

func (self *FuncDecl) CallableType() types.CallableType {
	return self.typ
}

func (self *FuncDecl) GetName() (string, bool) {
	return self.name, self.name != "_"
}

func (self *FuncDecl) Params() []*local.Param {
	return self.params
}

func (self *FuncDecl) Type() hir.Type {
	return self.CallableType()
}

func (self *FuncDecl) Ident() {}

func (self *FuncDecl) GetSelfParam(selfType hir.Type) (*local.Param, bool) {
	if len(self.params) == 0 {
		return nil, false
	}
	if ctd, ok := types.As[CustomTypeDef](selfType, true); ok && len(ctd.GenericParams()) > 0 {
		selfType = NewGenericCustomTypeDef(
			ctd,
			stlslices.Map(ctd.GenericParams(), func(_ int, gp types.GenericParamType) hir.Type {
				return gp
			})...,
		)
	}
	firstParam := stlslices.First(self.params)
	firstParamType := stlval.TernaryAction(types.Is[types.RefType](firstParam.Type(), true), func() hir.Type {
		rt, ok := types.As[types.RefType](firstParam.Type(), true)
		if !ok {
			panic("unreachable")
		}
		return rt.Pointer()
	}, func() hir.Type {
		return firstParam.Type()
	})
	if !firstParamType.Equal(selfType) && !firstParamType.Equal(types.Self) {
		return nil, false
	}
	return firstParam, true
}

// FuncAttr 函数属性
type FuncAttr interface {
	funcAttr()
}

// FuncAttrLinkName 链接名
type FuncAttrLinkName struct {
	name string
}

func WithLinkNameFuncAttr(name string) FuncAttr {
	return &FuncAttrLinkName{name: name}
}
func (self *FuncAttrLinkName) funcAttr() {}
func (self *FuncAttrLinkName) Name() string {
	return self.name
}

// FuncAttrInline 内联
type FuncAttrInline struct {
	inline bool
}

func WithInlineFuncAttr(inline bool) FuncAttr {
	return &FuncAttrInline{inline: inline}
}
func (self *FuncAttrInline) funcAttr() {}
func (self *FuncAttrInline) Inline() bool {
	return self.inline
}

// FuncAttrVararg 可变参数
type FuncAttrVararg struct{}

func WithVarargFuncAttr() FuncAttr {
	return &FuncAttrVararg{}
}
func (self *FuncAttrVararg) funcAttr() {}
