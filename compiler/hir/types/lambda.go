package types

import (
	"fmt"
	"strings"
	"unsafe"

	stlslices "github.com/kkkunny/stl/container/slices"
	stlval "github.com/kkkunny/stl/value"
)

// LambdaType 匿名函数类型
type LambdaType interface {
	BuildInType
	CallableType
	Lambda()
}

type _LambdaType_ struct {
	ret    Type
	params []Type
}

func NewLambdaType(ret Type, ps ...Type) LambdaType {
	return &_LambdaType_{
		ret:    ret,
		params: ps,
	}
}

func (self *_LambdaType_) String() string {
	params := stlslices.Map(self.params, func(i int, p Type) string { return p.String() })
	ret := stlval.Ternary(Is[NoThingType](self.ret, true), "void", self.ret.String())
	return fmt.Sprintf("(%s)->%s", strings.Join(params, ", "), ret)
}

func (self *_LambdaType_) Equal(dst Type) bool {
	t, ok := As[LambdaType](dst, true)
	if !ok || len(self.params) != len(t.Params()) || !self.ret.Equal(t.Ret()) {
		return false
	}
	return stlslices.All(self.params, func(i int, p Type) bool {
		return p.Equal(t.Params()[i])
	})
}

func (self *_LambdaType_) EqualWithSelf(dst Type, selfs ...Type) bool {
	if dst.Equal(Self) && len(selfs) > 0 {
		dst = stlslices.Last(selfs)
	}

	t, ok := As[LambdaType](dst, true)
	if !ok || len(self.params) != len(t.Params()) || !self.ret.EqualWithSelf(t.Ret(), selfs...) {
		return false
	}
	return stlslices.All(self.params, func(i int, p Type) bool {
		return p.EqualWithSelf(t.Params()[i], selfs...)
	})
}

func (self *_LambdaType_) Ret() Type {
	return self.ret
}

func (self *_LambdaType_) Params() []Type {
	return self.params
}

func (self *_LambdaType_) ToFunc() FuncType {
	return NewFuncType(self.ret, self.params...)
}

func (self *_LambdaType_) Lambda() {}

func (self *_LambdaType_) BuildIn() {}

func (self *_LambdaType_) Hash() uint64 {
	return uint64(uintptr(unsafe.Pointer(self)))
}
