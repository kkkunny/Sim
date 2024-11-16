package types

import (
	"fmt"
	"unsafe"

	"github.com/kkkunny/Sim/compiler/hir"
)

// RefType 引用类型
type RefType interface {
	BuildInType
	Mutable() bool
	Pointer() hir.Type
}

func NewRefType(mut bool, p hir.Type) RefType {
	return &_RefType_{
		mut: mut,
		ptr: p,
	}
}

type _RefType_ struct {
	mut bool
	ptr hir.Type
}

func (self *_RefType_) String() string {
	if self.mut {
		return fmt.Sprintf("&mut %s", self.ptr.String())
	} else {
		return fmt.Sprintf("&%s", self.ptr.String())
	}
}

func (self *_RefType_) Equal(dst hir.Type) bool {
	t, ok := As[RefType](dst, true)
	return ok && self.ptr.Equal(t.Pointer())
}

func (self *_RefType_) Pointer() hir.Type {
	return self.ptr
}

func (self *_RefType_) Mutable() bool {
	return self.mut
}

func (self *_RefType_) BuildIn() {}

func (self *_RefType_) Hash() uint64 {
	return uint64(uintptr(unsafe.Pointer(self)))
}
