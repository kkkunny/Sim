package analyse

import (
	. "github.com/kkkunny/Sim/src/compiler/hir"
	stlos "github.com/kkkunny/stl/os"
	"github.com/kkkunny/stl/types"
)

// CompilerContext 编译环境
type CompilerContext struct {
	Links map[stlos.Path]struct{}
	Libs  map[string]struct{}
}

// 新建编译环境
func newCompilerContext() *CompilerContext {
	return &CompilerContext{
		Links: make(map[stlos.Path]struct{}),
		Libs:  make(map[string]struct{}),
	}
}

// ProgramContext 程序环境
type ProgramContext struct {
	*CompilerContext
	Pkgs    map[stlos.Path]*packageContext
	Globals []Global
}

// 新建程序环境
func newProgramContext() *ProgramContext {
	return &ProgramContext{
		Pkgs:            make(map[stlos.Path]*packageContext),
		CompilerContext: newCompilerContext(),
	}
}

// 包环境
type packageContext struct {
	f    *ProgramContext
	path stlos.Path

	globals  map[string]types.Pair[bool, Ident]
	typedefs map[string]types.Pair[bool, *Typedef]
}

// 新建包环境
func newPackageContext(f *ProgramContext, path stlos.Path) *packageContext {
	return &packageContext{
		f:        f,
		path:     path,
		globals:  make(map[string]types.Pair[bool, Ident]),
		typedefs: make(map[string]types.Pair[bool, *Typedef]),
	}
}

// GetProgramContext 获取程序环境
func (self packageContext) GetProgramContext() *ProgramContext {
	return self.f
}

func (self packageContext) GetValue(name string) types.Pair[bool, Ident] {
	if v, ok := self.globals[name]; ok {
		return v
	}
	return types.NewPair[bool, Ident](false, nil)
}

func (self *packageContext) AddValue(pub bool, name string, value Ident) bool {
	if _, ok := self.globals[name]; ok {
		return false
	}
	self.globals[name] = types.NewPair(pub, value)
	return true
}

// 本地环境
type localContext interface {
	AddValue(name string, value Ident) bool
	GetValue(name string) Ident
	GetRetType() Type
	GetPackageContext() *packageContext
	SetEnd()
	IsEnd() bool
	GetPackagePath() stlos.Path
}

// 函数环境
type functionContext struct {
	f      *packageContext
	ret    Type
	params map[string]*Param
	end    bool
}

// 新建函数环境
func newFunctionContext(f *packageContext, ret Type) *functionContext {
	return &functionContext{
		f:      f,
		ret:    ret,
		params: make(map[string]*Param),
	}
}

func (self functionContext) GetRetType() Type {
	return self.ret
}

func (self functionContext) GetValue(name string) Ident {
	param, ok := self.params[name]
	if ok {
		return param
	}
	return self.f.GetValue(name).Second
}

func (self *functionContext) AddValue(name string, value Ident) bool {
	if _, ok := self.params[name]; ok {
		return false
	}
	self.params[name] = value.(*Param)
	return true
}

func (self *functionContext) GetPackageContext() *packageContext {
	return self.f
}

func (self *functionContext) SetEnd() {
	self.end = true
}

func (self functionContext) IsEnd() bool {
	return self.end
}

func (self functionContext) GetPackagePath() stlos.Path {
	return self.GetPackageContext().path
}

// 代码块环境
type blockContext struct {
	f      localContext
	inLoop bool
	locals map[string]*Variable
	end    bool
}

// 代码块环境
func newBlockContext(f localContext, inLoop bool) *blockContext {
	return &blockContext{
		f:      f,
		inLoop: inLoop,
		locals: make(map[string]*Variable),
	}
}

func (self blockContext) GetRetType() Type {
	return self.f.GetRetType()
}

func (self blockContext) GetValue(name string) Ident {
	local, ok := self.locals[name]
	if ok {
		return local
	}
	return self.f.GetValue(name)
}

func (self *blockContext) AddValue(name string, value Ident) bool {
	self.locals[name] = value.(*Variable)
	return true
}

func (self *blockContext) IsInLoop() bool {
	if self.inLoop {
		return true
	} else if self.f != nil {
		if fb, ok := self.f.(*blockContext); ok {
			return fb.IsInLoop()
		}
	}
	return false
}

func (self *blockContext) GetPackageContext() *packageContext {
	return self.f.GetPackageContext()
}

func (self *blockContext) SetEnd() {
	self.end = true
}

func (self blockContext) IsEnd() bool {
	return self.end
}

func (self blockContext) GetPackagePath() stlos.Path {
	return self.GetPackageContext().path
}
