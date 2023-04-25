package codegen

import (
	"github.com/kkkunny/Sim/src/compiler/mir"
	"github.com/kkkunny/llvm"
	stlutil "github.com/kkkunny/stl/util"
)

// CodeGenerator 代码生成器
type CodeGenerator struct {
	targetMachine llvm.TargetMachine
	targetData    llvm.TargetData
	ctx           llvm.Context
	module        llvm.Module
	builder       llvm.Builder
	function      llvm.Value

	globals map[mir.Global]llvm.Value
	types   map[*mir.Alias]llvm.Type
	vars    map[mir.Value]llvm.Value
	blocks  map[*mir.Block]llvm.BasicBlock

	// init fini
	inits, finis []llvm.Value
}

// NewCodeGenerator 新建代码生成器
func NewCodeGenerator(targetStr string, release bool) (*CodeGenerator, error) {
	if defaultTarget := llvm.DefaultTargetTriple(); targetStr == "" || targetStr == defaultTarget {
		targetStr = defaultTarget
		if err := llvm.InitializeNativeTarget(); err != nil {
			return nil, err
		}
		if err := llvm.InitializeNativeAsmPrinter(); err != nil {
			return nil, err
		}
	} else {
		llvm.InitializeAllTargets()
		llvm.InitializeAllAsmPrinters()
	}

	ctx := llvm.NewContext()
	module := ctx.NewModule("")
	module.SetTarget(targetStr)
	target, err := llvm.GetTargetFromTriple(targetStr)
	if err != nil {
		return nil, err
	}
	tm := target.CreateTargetMachine(
		targetStr,
		"generic",
		"",
		stlutil.Ternary(release, llvm.CodeGenLevelAggressive, llvm.CodeGenLevelNone),
		llvm.RelocPIC,
		llvm.CodeModelDefault,
	)
	module.SetDataLayout(tm.CreateTargetData().String())
	return &CodeGenerator{
		targetMachine: tm,
		targetData:    tm.CreateTargetData(),
		ctx:           ctx,
		module:        module,
		builder:       ctx.NewBuilder(),
		globals:       make(map[mir.Global]llvm.Value),
		types:         make(map[*mir.Alias]llvm.Type),
	}, nil
}

// Codegen 代码生成
func (self *CodeGenerator) Codegen(pkg mir.Package) (llvm.Module, llvm.TargetMachine) {
	// 类型声明
	for cursor := pkg.Globals.Front(); cursor != nil; cursor = cursor.Next() {
		if alias, ok := cursor.Value().(*mir.Alias); ok {
			self.codegenAliasDecl(alias)
		}
	}
	// 类型定义
	for cursor := pkg.Globals.Front(); cursor != nil; cursor = cursor.Next() {
		if alias, ok := cursor.Value().(*mir.Alias); ok {
			self.codegenAliasDef(alias)
		}
	}
	// 声明
	for cursor := pkg.Globals.Front(); cursor != nil; cursor = cursor.Next() {
		switch global := cursor.Value().(type) {
		case *mir.Alias:
		case *mir.Function:
			fn := self.codegenFunctionDecl(global)
			if global.GetInit() {
				self.inits = append(self.inits, fn)
			}
			if global.GetFini() {
				self.finis = append(self.finis, fn)
			}
		case *mir.Variable:
			self.codegenVariableDecl(global)
		default:
			panic("unreachable")
		}
	}
	// 定义
	for cursor := pkg.Globals.Front(); cursor != nil; cursor = cursor.Next() {
		switch global := cursor.Value().(type) {
		case *mir.Alias:
		case *mir.Function:
			self.codegenFunctionDef(global)
		case *mir.Variable:
			self.codegenVariableDef(global)
		default:
			panic("unreachable")
		}
	}
	// init && fini
	self.codegenInitAndFini()
	return self.module, self.targetMachine
}

// init && fini
func (self *CodeGenerator) codegenInitAndFini() {
	structType := self.ctx.StructType(
		[]llvm.Type{
			self.ctx.Int32Type(),
			llvm.PointerType(llvm.FunctionType(self.ctx.VoidType(), nil, false), 0),
			llvm.PointerType(self.ctx.Int8Type(), 0),
		},
		false,
	)
	// init
	if len(self.inits) > 0 {
		init := llvm.AddGlobal(self.module, llvm.ArrayType(structType, len(self.inits)), "llvm.global_ctors")
		init.SetLinkage(llvm.AppendingLinkage)
		values := make([]llvm.Value, len(self.inits))
		for i, f := range self.inits {
			values[i] = llvm.ConstStruct(
				[]llvm.Value{
					llvm.ConstInt(structType.StructElementTypes()[0], 65535, true),
					f,
					llvm.ConstZero(structType.StructElementTypes()[2]),
				},
				false,
			)
		}
		init.SetInitializer(llvm.ConstArray(structType, values))
	}
	// fini
	if len(self.finis) > 0 {
		fini := llvm.AddGlobal(self.module, llvm.ArrayType(structType, len(self.finis)), "llvm.global_dtors")
		fini.SetLinkage(llvm.AppendingLinkage)
		values := make([]llvm.Value, len(self.finis))
		for i, f := range self.finis {
			values[len(self.finis)-i-1] = llvm.ConstStruct(
				[]llvm.Value{
					llvm.ConstInt(structType.StructElementTypes()[0], 65535, true),
					f,
					llvm.ConstZero(structType.StructElementTypes()[2]),
				},
				false,
			)
		}
		fini.SetInitializer(llvm.ConstArray(structType, values))
	}
}
