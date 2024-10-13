package types

import (
	"fmt"
	"strings"

	stlslices "github.com/kkkunny/stl/container/slices"
)

// LambdaType 匿名函数类型
type LambdaType interface {
	Type
	Ret() Type
	Params() []Type
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
	return fmt.Sprintf("(%s)->%s", strings.Join(params, ", "), self.ret.String())
}

func (self *_LambdaType_) Equal(dst Type) bool {
	t, ok := dst.(LambdaType)
	if !ok || len(self.params) != len(t.Params()) || !self.ret.Equal(t.Ret()) {
		return false
	}
	return stlslices.All(self.params, func(i int, p Type) bool {
		return p.Equal(t.Params()[i])
	})
}

func (self *_LambdaType_) Ret() Type {
	return self.ret
}

func (self *_LambdaType_) Params() []Type {
	return self.params
}
