package types

import (
	"fmt"
	"strings"

	stlslices "github.com/kkkunny/stl/container/slices"
)

// FuncType 函数类型
type FuncType interface {
	BuildInType
	CallableType
	Func()
}

type _FuncType_ struct {
	ret    Type
	params []Type
}

func NewFuncType(ret Type, ps ...Type) FuncType {
	return &_FuncType_{
		ret:    ret,
		params: ps,
	}
}

func (self *_FuncType_) String() string {
	params := stlslices.Map(self.params, func(i int, p Type) string { return p.String() })
	return fmt.Sprintf("func(%s)%s", strings.Join(params, ", "), self.ret.String())
}

func (self *_FuncType_) Equal(dst Type, selfs ...Type) bool {
	if dst.Equal(Self) && len(selfs) > 0 {
		dst = stlslices.Last(selfs)
	}

	t, ok := dst.(FuncType)
	if !ok || len(self.params) != len(t.Params()) || !self.ret.Equal(t.Ret(), selfs...) {
		return false
	}
	return stlslices.All(self.params, func(i int, p Type) bool {
		return p.Equal(t.Params()[i], selfs...)
	})
}

func (self *_FuncType_) Ret() Type {
	return self.ret
}

func (self *_FuncType_) Params() []Type {
	return self.params
}

func (self *_FuncType_) Func() {}

func (self *_FuncType_) BuildIn() {}
