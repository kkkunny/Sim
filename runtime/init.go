package runtime

/*
#include "runtime.h"
*/
import "C"

// Function 外部函数
type Function struct {
	Name string
	C    bool
	To   any
}

// FuncList 外部函数列表
var FuncList = [...]Function{
	{
		Name: "sim_runtime_debug",
		To:   sim_runtime_debug,
	},
	{
		Name: "sim_runtime_panic",
		To:   sim_runtime_panic,
	},
	{
		Name: "sim_runtime_gc_init",
		C:    true,
		To:   C.sim_runtime_gc_init,
	},
	{
		Name: "sim_runtime_gc_alloc",
		C:    true,
		To:   C.sim_runtime_gc_alloc,
	},
	{
		Name: "sim_runtime_gc_gc",
		C:    true,
		To:   C.sim_runtime_gc_gc,
	},
}
