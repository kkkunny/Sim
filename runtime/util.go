package runtime

// 工具类函数，不可被显示调用

/*
#include <stdlib.h>
#include <stdint.h>
*/
import "C"
import (
	"fmt"
	"os"
	"unsafe"
)

// sim_runtime_malloc 在gc上分配堆内存（byte）
//
//export sim_runtime_malloc
func sim_runtime_malloc(size C.size_t) *C.void {
	ptr := C.calloc(1, C.size_t(size))
	if ptr == nil {
		_, _ = fmt.Fprintf(os.Stderr, "panic: %s\n", "zero exception")
		os.Exit(1)
	}
	return (*C.void)(ptr)
}

// sim_runtime_free 在gc上释放堆内存
//
//export sim_runtime_free
func sim_runtime_free(ptr *C.void) {
	C.free(unsafe.Pointer(ptr))
}
