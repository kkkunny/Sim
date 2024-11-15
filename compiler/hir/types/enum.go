package types

import (
	"fmt"
	"strings"
	"unsafe"

	"github.com/kkkunny/stl/container/linkedhashmap"
	stlslices "github.com/kkkunny/stl/container/slices"
)

type EnumField struct {
	name string
	elem Type
}

func NewEnumField(name string, typ ...Type) *EnumField {
	return &EnumField{
		name: name,
		elem: stlslices.Last(typ),
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
	BuildInType
	EnumFields() linkedhashmap.LinkedHashMap[string, *EnumField]
	Simple() bool
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
	t, ok := As[EnumType](dst, true)
	if !ok {
		return false
	}
	if self.fields.Length() != t.EnumFields().Length() {
		return false
	}
	return stlslices.All(self.fields.Values(), func(i int, f1 *EnumField) bool {
		f2 := t.EnumFields().Get(f1.Name())
		e1, ok1 := f1.Elem()
		e2, ok2 := f2.Elem()
		if !ok1 && !ok2 {
			return true
		}
		return ok1 == ok2 && e1.Equal(e2)
	})
}

func (self *_EnumType_) EqualWithSelf(dst Type, selfs ...Type) bool {
	if dst.Equal(Self) && len(selfs) > 0 {
		dst = stlslices.Last(selfs)
	}

	t, ok := As[EnumType](dst, true)
	if !ok {
		return false
	}
	if self.fields.Length() != t.EnumFields().Length() {
		return false
	}
	return stlslices.All(self.fields.Values(), func(i int, f1 *EnumField) bool {
		f2 := t.EnumFields().Get(f1.Name())
		e1, ok1 := f1.Elem()
		e2, ok2 := f2.Elem()
		return ok1 == ok2 && e1.EqualWithSelf(e2, selfs...)
	})
}

func (self *_EnumType_) EnumFields() linkedhashmap.LinkedHashMap[string, *EnumField] {
	return self.fields
}

func (self *_EnumType_) Simple() bool {
	return stlslices.All(self.fields.Values(), func(_ int, f *EnumField) bool {
		return f.elem == nil
	})
}

func (self *_EnumType_) BuildIn() {}

func (self *_EnumType_) Hash() uint64 {
	return uint64(uintptr(unsafe.Pointer(self)))
}
