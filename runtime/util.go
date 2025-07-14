package runtime

/*
#include <stdlib.h>
#include <stdint.h>
*/
import "C"
import (
	"fmt"
	"os"
	"unsafe"

	"github.com/kkkunny/Sim/runtime/gc"
)

// 工具类函数，不可被显示调用

// sim_runtime_gc_init 初始化gc
//
//export sim_runtime_gc__init
func sim_runtime_gc__init(stackBegin, dataBegin, dataEnd *C.void) {
	err := gc.Init(uintptr(unsafe.Pointer(stackBegin)), uintptr(unsafe.Pointer(dataBegin)), uintptr(unsafe.Pointer(dataEnd)))
	if err != nil {
		_, _ = fmt.Fprintf(os.Stderr, "panic: %s\n", err)
		os.Exit(1)
	}
}

// sim_runtime_gc__alloc 在gc上分配堆内存（byte）
//
//export sim_runtime_gc__alloc
func sim_runtime_gc__alloc(size C.size_t, stackPos C.size_t) *C.void {
	ptr, err := gc.Alloca(uint(size), uintptr(stackPos))
	if err != nil {
		_, _ = fmt.Fprintf(os.Stderr, "panic: %s\n", err)
		os.Exit(1)
	}
	return (*C.void)(ptr)
}

// sim_runtime_gc__gc 立即进行垃圾回收
//
//export sim_runtime_gc__gc
func sim_runtime_gc__gc(stackPos C.size_t) {
	err := gc.GC(uintptr(stackPos))
	if err != nil {
		_, _ = fmt.Fprintf(os.Stderr, "panic: %s\n", err)
		os.Exit(1)
	}
}

// // sim_runtime_free 在gc上释放堆内存
// //
// //export sim_runtime_free
// func sim_runtime_free(ptr *C.void) {
// 	C.free(unsafe.Pointer(ptr))
// }
