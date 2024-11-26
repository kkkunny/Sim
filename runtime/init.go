package runtime

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
		To:   sim_runtime_debug,
	},
	{
		Name: "sim_runtime_panic",
		To:   sim_runtime_panic,
	},
	{
		Name: "sim_runtime_malloc",
		To:   sim_runtime_malloc,
	},
}
