package extern

// 工具类函数，不可被显示调用

/*
#include <stdlib.h>
*/
import "C"
import (
	"unsafe"

	stlbasic "github.com/kkkunny/stl/basic"

	"github.com/kkkunny/Sim/runtime/types"
)

// GCAlloc 在gc上分配堆内存（byte）
func GCAlloc(size types.Usize) types.Ptr {
	ptr := types.NewPtr(C.calloc(1, C.size_t(size)))
	ptr = CheckNull(ptr)
	return ptr
}

// GCFree 在gc上释放堆内存
func GCFree(p types.Ptr) {
	C.free(unsafe.Pointer(p))
}

// StrEqStr 字符串比较
func StrEqStr(l, r types.Str) types.Bool {
	return types.NewBool(l.Value() == r.Value())
}

// CheckNull 检查空指针
func CheckNull(p types.Ptr) types.Ptr {
	if p == nil {
		Panic(stlbasic.Ptr(types.NewStr("zero exception")))
	}
	return p
}
