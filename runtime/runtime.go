package runtime

// #include "runtime.h"
import "C"
import "fmt"

// StrEqStr 字符串比较
var StrEqStr = C.sim_runtime_str_eq_str

//export sim_runtime_str_eq_str
func sim_runtime_str_eq_str(l, r Str) Bool {
	return NewBool(l.Value() == r.Value())
}

// Debug 输出字符串
var Debug = C.sim_runtime_debug

//export sim_runtime_debug
func sim_runtime_debug(s Str) {
	fmt.Println(s)
}
