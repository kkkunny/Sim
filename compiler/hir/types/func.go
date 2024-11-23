package types

import (
	"fmt"
	"strings"
	"unsafe"

	stlslices "github.com/kkkunny/stl/container/slices"

	"github.com/kkkunny/Sim/compiler/hir"
)

// FuncType 函数类型
type FuncType interface {
	hir.BuildInType
	CallableType
	Func()
}

type _FuncType_ struct {
	ret    hir.Type
	params []hir.Type
}

func NewFuncType(ret hir.Type, ps ...hir.Type) FuncType {
	return &_FuncType_{
		ret:    ret,
		params: ps,
	}
}

func (self *_FuncType_) String() string {
	params := stlslices.Map(self.params, func(i int, p hir.Type) string { return p.String() })
	return fmt.Sprintf("func(%s)%s", strings.Join(params, ", "), self.ret.String())
}

func (self *_FuncType_) Equal(dst hir.Type) bool {
	t, ok := As[FuncType](dst, true)
	if !ok || len(self.params) != len(t.Params()) || !self.ret.Equal(t.Ret()) {
		return false
	}
	return stlslices.All(self.params, func(i int, p hir.Type) bool {
		return p.Equal(t.Params()[i])
	})
}

func (self *_FuncType_) Ret() hir.Type {
	return self.ret
}

func (self *_FuncType_) Params() []hir.Type {
	return self.params
}

func (self *_FuncType_) Func() {}

func (self *_FuncType_) BuildIn() {}

func (self *_FuncType_) ToFunc() FuncType {
	return self
}

func (self *_FuncType_) Hash() uint64 {
	return uint64(uintptr(unsafe.Pointer(self)))
}
