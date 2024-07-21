package codegen_ir

var modulePasses = []string{
	"constmerge",    // 合并常量
	"globaldce",     // 死全局变量消除
	"always-inline", // 函数内联
}

var functionPasses = []string{
	"dce", // 死代码消除
}
