package types

import "github.com/kkkunny/Sim/compiler/hir"

// CompileCallType 编译时调用类型
type CompileCallType interface {
	CustomType
	CompileCall()
	Args() []hir.Type
	WithArgs(args []hir.Type) CompileCallType
}
