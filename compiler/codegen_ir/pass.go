package codegen_ir

var modulePasses = []string{
	"constmerge", // 合并常量
	"globaldce",  // 死全局变量消除
}

var functionPasses = []string{
	"dce",           // 死代码消除
	"gvn",           // 删除冗余指令
	"instcombine",   // 合并冗余指令
	"always-inline", // 函数内联
}
