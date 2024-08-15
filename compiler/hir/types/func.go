package types

import (
	"fmt"
	"strings"

	stlslices "github.com/kkkunny/stl/container/slices"
)

// FuncType 函数类型
type FuncType struct {
	ret    Type
	params []Type
}

func NewFuncType(ret Type, ps ...Type) *FuncType {
	return &FuncType{
		ret:    ret,
		params: ps,
	}
}

func (self *FuncType) String() string {
	params := stlslices.Map(self.params, func(i int, p Type) string { return p.String() })
	return fmt.Sprintf("func(%s)%s", strings.Join(params, ", "), self.ret.String())
}

func (self *FuncType) Equal(dst Type) bool {
	t, ok := dst.(*FuncType)
	if !ok || len(self.params) != len(t.params) || !self.ret.Equal(t.ret) {
		return false
	}
	return stlslices.All(self.params, func(i int, p Type) bool {
		return p.Equal(t.params[i])
	})
}

func (self *FuncType) Ret() Type {
	return self.ret
}

func (self *FuncType) Params() []Type {
	return self.params
}
