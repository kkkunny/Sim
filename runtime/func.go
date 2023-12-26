package runtime

// RuntimeFunction 运行时函数
type RuntimeFunction struct {
	Name string
	To any
}

// RuntimeFuncList 运行时函数列表
var RuntimeFuncList = [...]RuntimeFunction{
	// runtime
	{
		Name: "sim_runtime_str_eq_str",
		To: SimRuntimeStrEqStr,
	},
	{
		Name: "sim_runtime_debug",
		To: SimRuntimeDebug,
	},
	{
		Name: "sim_runtime_check_null",
		To: SimRuntimeCheckNull,
	},
	// c
	{
		Name: "exit",
		To: Exit,
	},
}
