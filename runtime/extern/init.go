package extern

// Function 外部函数
type Function struct {
	Name string
	To   any
}

// FuncList 外部函数列表
var FuncList = [...]Function{
	// runtime
	{
		Name: "sim_runtime_debug",
		To:   Debug,
	},
	{
		Name: "sim_runtime_panic",
		To:   Panic,
	},
	{
		Name: "sim_runtime_malloc",
		To:   GCAlloc,
	},
}
