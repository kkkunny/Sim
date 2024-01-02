package extern

// Function 外部函数
type Function struct {
	Name string
	To any
}

// FuncList 外部函数列表
var FuncList = [...]Function{
	// runtime
	{
		Name: "sim_runtime_str_eq_str",
		To:   StrEqStr,
	},
	{
		Name: "sim_runtime_debug",
		To:   Debug,
	},
	{
		Name: "sim_runtime_check_null",
		To:   CheckNull,
	},
}
