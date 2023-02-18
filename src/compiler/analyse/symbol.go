package analyse

import (
	"github.com/kkkunny/Sim/src/compiler/hir"
	stlos "github.com/kkkunny/stl/os"
)

// 符号对象
type symbolObject[T any] struct {
	pub  bool // 是否对外部包公开
	data T    // 数据
}

// 符号表
type symbolTable struct {
	pkg hir.PkgPath  // 包地址
	f   *symbolTable // 父符号表

	types  map[string]symbolObject[*hir.Typedef] // 类型
	values map[string]symbolObject[hir.Ident]    // 值

	inLoop bool     // 是否处于循环中
	ret    hir.Type // 返回值类型
}

// 新建包符号表
func newPkgSymbolTable(pkg stlos.Path) *symbolTable {
	return &symbolTable{
		pkg:    hir.NewPkgPath(pkg),
		types:  make(map[string]symbolObject[*hir.Typedef]),
		values: make(map[string]symbolObject[hir.Ident]),
	}
}

// 新建代码块符号表
func newBlockSymbolTable(f *symbolTable) *symbolTable {
	return &symbolTable{
		pkg:    f.pkg,
		f:      f,
		types:  make(map[string]symbolObject[*hir.Typedef]),
		values: make(map[string]symbolObject[hir.Ident]),
		inLoop: f.inLoop,
		ret:    f.ret,
	}
}

// 是否是包符号表
func (self symbolTable) isPkg() bool {
	return self.f == nil
}

// 声明类型
func (self symbolTable) declType(pub bool, def *hir.Typedef) bool {
	if !self.isPkg() {
		panic("unreachable")
	}

	if _, ok := self.types[def.Name]; ok {
		return false
	}
	self.types[def.Name] = symbolObject[*hir.Typedef]{
		pub:  pub,
		data: def,
	}
	return true
}

// 定义类型
func (self symbolTable) defType(name string, target hir.Type) {
	if !self.isPkg() {
		panic("unreachable")
	}
	self.types[name].data.Target = target
}

// 寻找类型
func (self symbolTable) lookupType(name string) (symbolObject[*hir.Typedef], bool) {
	if t, ok := self.types[name]; ok {
		return t, true
	}
	if !self.isPkg() {
		return self.f.lookupType(name)
	}
	return symbolObject[*hir.Typedef]{}, false
}

// 声明值
func (self symbolTable) defValue(pub bool, name string, v hir.Ident) bool {
	if self.isPkg() {
		if _, ok := self.values[name]; ok {
			return false
		}
	}
	self.values[name] = symbolObject[hir.Ident]{
		pub:  pub,
		data: v,
	}
	return true
}

// 寻找值
func (self symbolTable) lookupValue(name string) (symbolObject[hir.Ident], bool) {
	if v, ok := self.values[name]; ok {
		return v, true
	}
	if !self.isPkg() {
		return self.f.lookupValue(name)
	}
	return symbolObject[hir.Ident]{}, false
}

// 获取包符号表
func (self *symbolTable) getPkgSymbolTable() *symbolTable {
	if self.isPkg() {
		return self
	}
	return self.f.getPkgSymbolTable()
}
