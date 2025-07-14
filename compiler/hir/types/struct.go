package types

import (
	"fmt"
	"strings"
	"unsafe"

	"github.com/kkkunny/stl/container/linkedhashmap"
	stlslices "github.com/kkkunny/stl/container/slices"

	"github.com/kkkunny/Sim/compiler/hir"
)

type Field struct {
	pub  bool
	mut  bool
	name string
	typ  hir.Type
}

func NewField(pub bool, mut bool, name string, typ hir.Type) *Field {
	return &Field{
		pub:  pub,
		mut:  mut,
		name: name,
		typ:  typ,
	}
}

func (self *Field) Name() string {
	return self.name
}

func (self *Field) Type() hir.Type {
	return self.typ
}

func (self *Field) Public() bool {
	return self.pub
}

func (self *Field) Mutable() bool {
	return self.mut
}

// StructType 结构体类型
type StructType interface {
	hir.BuildInType
	Fields() linkedhashmap.LinkedHashMap[string, *Field]
}

func NewStructType(fs ...*Field) StructType {
	lhm := linkedhashmap.StdWithCap[string, *Field](uint(len(fs)))
	for _, f := range fs {
		lhm.Set(f.Name(), f)
	}
	return &_StructType_{
		fields: lhm,
	}
}

type _StructType_ struct {
	fields linkedhashmap.LinkedHashMap[string, *Field]
}

func (self *_StructType_) String() string {
	fields := stlslices.Map(self.fields.Values(), func(i int, f *Field) string {
		return fmt.Sprintf("%s:%s", f.name, f.typ.String())
	})
	return fmt.Sprintf("struct{%s}", strings.Join(fields, ";"))
}

func (self *_StructType_) Equal(dst hir.Type) bool {
	t, ok := As[StructType](dst, true)
	if !ok || self.fields.Length() != t.Fields().Length() {
		return false
	}
	fields2 := t.Fields().Values()
	return stlslices.All(self.fields.Values(), func(i int, f1 *Field) bool {
		f2 := fields2[i]
		return f1.pub == f2.pub &&
			f1.mut == f2.mut &&
			f1.name == f2.name &&
			f1.typ.Equal(f2.typ)
	})
}

func (self *_StructType_) Fields() linkedhashmap.LinkedHashMap[string, *Field] {
	return self.fields
}

func (self *_StructType_) BuildIn() {}

func (self *_StructType_) Hash() uint64 {
	return uint64(uintptr(unsafe.Pointer(self)))
}
