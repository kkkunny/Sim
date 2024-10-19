package types

import (
	"fmt"
	"strings"

	"github.com/kkkunny/stl/container/linkedhashmap"
	stlslices "github.com/kkkunny/stl/container/slices"
)

type EnumField struct {
	name string
	elem Type
}

func NewEnumField(name string, typ ...Type) *Field {
	return &Field{
		name: name,
		typ:  stlslices.Last(typ),
	}
}

func (self *EnumField) Name() string {
	return self.name
}

func (self *EnumField) Elem() (Type, bool) {
	return self.elem, self.elem != nil
}

// EnumType 枚举类型
type EnumType interface {
	Type
	EnumFields() []*EnumField
}

func NewEnumType(fs ...*EnumField) EnumType {
	return &_EnumType_{
		fields: linkedhashmap.StdWith[string, *EnumField](stlslices.FlatMap(fs, func(i int, f *EnumField) []any {
			return []any{f.name, f}
		})...),
	}
}

type _EnumType_ struct {
	fields linkedhashmap.LinkedHashMap[string, *EnumField]
}

func (self *_EnumType_) String() string {
	fields := stlslices.Map(self.fields.Values(), func(i int, f *EnumField) string {
		e, ok := f.Elem()
		if !ok {
			return f.name
		}
		return fmt.Sprintf("%s:%s", f.name, e.String())
	})
	return fmt.Sprintf("enum{%s}", strings.Join(fields, ";"))
}

func (self *_EnumType_) Equal(dst Type) bool {
	t, ok := dst.(EnumType)
	if !ok {
		return false
	}
	dstFields := t.EnumFields()
	if self.fields.Length() != uint(len(dstFields)) {
		return false
	}
	return stlslices.All(self.fields.Values(), func(i int, f1 *EnumField) bool {
		f2 := dstFields[i]
		if f1.name != f2.name {
			return false
		}
		e1, ok1 := f1.Elem()
		e2, ok2 := f2.Elem()
		return ok1 == ok2 && e1.Equal(e2)
	})
}

func (self *_EnumType_) EnumFields() []*EnumField {
	return self.fields.Values()
}
