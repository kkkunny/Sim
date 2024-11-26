package codegen_ir

var modulePasses = []string{
	"codegenprepare", // 优化代码生成
	"da",             // 依赖关系分析
	"constmerge",     // 合并重复的全局常量
	"constprop",      // 简单的常量传播
	"deadtypeelim",   // 死类型消除
	"globaldce",      // 死全局变量消除
	"globalopt",      // 全局变量优化
	"die",            // 死指令消除
	"dse",            // 死存储消除
	"dce",            // 死代码消除
	"gvn",            // 全局值编号
	"simplifycfg",    // 简化CFG
	"instcombine",    // 合并冗余指令
	"indvars",        // 规范化归纳变量
	// "always_inline",  // 内联
	// "inline",         // 内联
}
