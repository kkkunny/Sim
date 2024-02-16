package extern

// 工具类函数，不可被显示调用

/*
#include <stdlib.h>
*/
import "C"
import (
	"encoding/json"
	"unsafe"

	stlbasic "github.com/kkkunny/stl/basic"
	stlslices "github.com/kkkunny/stl/slices"
	"github.com/samber/lo"

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

// CovertUnionIndex 获取原联合值实际类型在新联合类型中的下标
func CovertUnionIndex(srcStr, dstStr *types.Str, index types.U8) types.U8 {
	var srcTypes, dstTypes []string
	stlbasic.Ignore(json.Unmarshal([]byte(srcStr.String()), &srcTypes))
	stlbasic.Ignore(json.Unmarshal([]byte(dstStr.String()), &dstTypes))
	return types.NewU8(uint8(lo.IndexOf(dstTypes, srcTypes[index.Value()])))
}

// CheckUnionType 检查原联合值实际类型是否属于新联合类型
func CheckUnionType(srcStr, dstStr *types.Str, index types.U8) types.Bool {
	var srcTypes, dstTypes []string
	stlbasic.Ignore(json.Unmarshal([]byte(srcStr.String()), &srcTypes))
	stlbasic.Ignore(json.Unmarshal([]byte(dstStr.String()), &dstTypes))
	return types.NewBool(stlslices.Contain(dstTypes, srcTypes[index.Value()]))
}
