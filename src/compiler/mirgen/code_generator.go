package mirgen

import (
	"github.com/kkkunny/Sim/src/compiler/hir"
	"github.com/kkkunny/Sim/src/compiler/mir"
)

// MirGenerator 中级中间代码生成器
type MirGenerator struct {
	pkg   *mir.Package
	block *mir.Block

	vars  map[hir.Ident]mir.Value
	types map[string]mir.Type

	// loop
	cb, eb *mir.Block
	// string
	stringPool map[string]*mir.Variable
	// global
	globalDefFunc   *mir.Function
	globalLastBlock *mir.Block
	// init fini
	inits, finis []*mir.Function
}

// NewMirGenerator 新建中级中间代码生成器
func NewMirGenerator() *MirGenerator {
	return &MirGenerator{
		pkg:        mir.NewPackage(),
		vars:       make(map[hir.Ident]mir.Value),
		types:      make(map[string]mir.Type),
		stringPool: make(map[string]*mir.Variable),
	}
}

// Generate 生成
func (self *MirGenerator) Generate(pkg hir.Package) *mir.Package {
	// 声明
	self.genDecl(pkg)
	// 定义
	self.genDef(pkg)
	// init && fini
	self.genInit()
	self.genFini()
	return self.pkg
}

// 声明
func (self *MirGenerator) genDecl(pkg hir.Package) {
	for iter := pkg.Globals.Iterator(); iter.HasValue(); iter.Next() {
		switch global := iter.Value().(type) {
		case *hir.Typedef:
		case *hir.Function:
			self.genFunctionDecl(global)
		case *hir.Method:
			self.genMethodDecl(global)
		case *hir.GlobalValue:
			self.genGlobalValueDecl(global)
		default:
			panic("unreachable")
		}
	}
}

// 定义
func (self *MirGenerator) genDef(pkg hir.Package) {
	for iter := pkg.Globals.Iterator(); iter.HasValue(); iter.Next() {
		switch global := iter.Value().(type) {
		case *hir.Typedef:
		case *hir.Function:
			self.genFunctionDef(global)
		case *hir.Method:
			self.genMethodDef(global)
		case *hir.GlobalValue:
			self.genGlobalValueDef(global)
		default:
			panic("unreachable")
		}
	}
	if self.globalLastBlock != nil {
		self.globalLastBlock.NewReturn(nil)
	}
}

// init
func (self *MirGenerator) genInit() {
	if len(self.inits) > 0 {
		fn := self.pkg.NewFunction(
			mir.NewTypeFunc(false, mir.NewTypeVoid()),
			"",
		)
		fn.SetInit(true)
		block := fn.NewBlock()
		for _, f := range self.inits {
			block.NewCall(f)
		}
		block.NewReturn(nil)
	}
}

// fini
func (self *MirGenerator) genFini() {
	if len(self.finis) > 0 {
		fn := self.pkg.NewFunction(
			mir.NewTypeFunc(false, mir.NewTypeVoid()),
			"",
		)
		fn.SetFini(true)
		block := fn.NewBlock()
		for _, f := range self.finis {
			block.NewCall(f)
		}
		block.NewReturn(nil)
	}
}
