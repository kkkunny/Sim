package runtime

import (
	"fmt"
	"os"
	"unsafe"
)

/*
#include <stdint.h>

typedef struct{
	void* data;
	size_t len;
}str;
*/
import "C"

// 内置函数，可显式调用

// sim_runtime_debug 输出字符串
//
//export sim_runtime_debug
func sim_runtime_debug(s C.str) {
	fmt.Println(unsafe.String((*byte)(s.data), uint(s.len)))
}

// sim_runtime_panic 运行时异常抛出
//
//export sim_runtime_panic
func sim_runtime_panic(s C.str) {
	_, _ = fmt.Fprintf(os.Stderr, "panic: %s\n", unsafe.String((*byte)(s.data), uint(s.len)))
	os.Exit(1)
}
