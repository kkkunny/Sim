package extern

// 工具类函数，不可被显示调用

/*
#include <stdlib.h>
*/
import "C"
import (
	"encoding/gob"
	"strings"
	"unsafe"

	"github.com/kkkunny/Sim/runtime/types"
	stlerror "github.com/kkkunny/stl/error"
)

// GCAlloc 在gc上分配堆内存（byte）
func GCAlloc(size types.Usize) types.Ptr {
	ptr := types.NewPtr(C.calloc(1, C.size_t(size)))
	ptr = CheckNull(ptr)
	return ptr
}

// GCFree 在gc上释放堆内存
func GCFree(p types.Ptr){
	C.free(unsafe.Pointer(p))
}

// StrEqStr 字符串比较
func StrEqStr(l, r types.Str) types.Bool {
	return types.NewBool(l.Value() == r.Value())
}

// CheckNull 检查空指针
func CheckNull(p types.Ptr) types.Ptr {
	if p == nil {
		panic("空指针")
	}
	return p
}

// CovertUnionIndex 获取原联合值实际类型在新联合类型中的下标
func CovertUnionIndex(srcStr, dstStr types.Str, index types.U8)types.U8{
	src, dst := new(types.UnionType), new(types.UnionType)

	stlerror.Must(gob.NewDecoder(strings.NewReader(srcStr.String())).Decode(src))
	stlerror.Must(gob.NewDecoder(strings.NewReader(dstStr.String())).Decode(dst))

	if src.Contain(dst){
		return types.NewU8(uint8(src.IndexElem(dst.Elem(uint(index.Value())))))
	}else{
		return types.NewU8(uint8(dst.IndexElem(src.Elem(uint(index.Value())))))
	}
}
