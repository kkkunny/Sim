package analyse

import (
	"fmt"

	"github.com/kkkunny/Sim/src/compiler/hir"
	"github.com/kkkunny/Sim/src/compiler/parse"
	stlos "github.com/kkkunny/stl/os"
	"github.com/kkkunny/stl/queue"
)

// 符号对象
type symbolObject[T any] struct {
	pub  bool // 是否对外部包公开
	data T    // 数据
}

// 泛型对象
type symbolGenericObject[T any] struct {
	pub  bool // 是否对外部包公开
	data T    // 数据
}

// 符号表
type symbolTable struct {
	pkg hir.PkgPath  // 包地址
	f   *symbolTable // 父符号表

	types          map[string]symbolObject[*hir.Typedef]          // 类型
	values         map[string]symbolObject[hir.Ident]             // 值
	genericFuncs   map[string]symbolGenericObject[parse.Function] // 泛型函数
	genericTypeMap *queue.Queue[map[string]hir.Type]              // 泛型类型映射
	genericTypes   map[string]symbolGenericObject[parse.TypeDef]  // 泛型类型
	genericMethods map[string]symbolGenericObject[parse.Method]   // 泛型方法

	inLoop bool     // 是否处于循环中
	ret    hir.Type // 返回值类型
}

// 新建包符号表
func newPkgSymbolTable(pkg stlos.Path) *symbolTable {
	return &symbolTable{
		pkg:            hir.NewPkgPath(pkg),
		types:          make(map[string]symbolObject[*hir.Typedef]),
		values:         make(map[string]symbolObject[hir.Ident]),
		genericFuncs:   make(map[string]symbolGenericObject[parse.Function]),
		genericTypeMap: queue.NewQueue[map[string]hir.Type](),
		genericTypes:   make(map[string]symbolGenericObject[parse.TypeDef]),
		genericMethods: make(map[string]symbolGenericObject[parse.Method]),
	}
}

// 新建代码块符号表
func newBlockSymbolTable(f *symbolTable) *symbolTable {
	return &symbolTable{
		pkg:            f.pkg,
		f:              f,
		values:         make(map[string]symbolObject[hir.Ident]),
		genericTypeMap: queue.NewQueue[map[string]hir.Type](),
		inLoop:         f.inLoop,
		ret:            f.ret,
	}
}

// 是否是包符号表
func (self symbolTable) isPkg() bool {
	return self.f == nil
}

// 声明类型
func (self *symbolTable) declType(pub bool, def *hir.Typedef) bool {
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
func (self *symbolTable) defType(name string, target hir.Type) {
	if !self.isPkg() {
		panic("unreachable")
	}
	self.types[name].data.Target = target
}

// 寻找类型
func (self symbolTable) lookupType(name string) (symbolObject[*hir.Typedef], bool) {
	if !self.isPkg() {
		return self.f.lookupType(name)
	}
	if t, ok := self.types[name]; ok {
		return t, true
	}
	return symbolObject[*hir.Typedef]{}, false
}

// 声明值
func (self *symbolTable) defValue(pub bool, name string, v hir.Ident) bool {
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

// 声明泛型函数
func (self *symbolTable) defGenericFunc(pub bool, name string, ast parse.Function) bool {
	if !self.isPkg() {
		panic("unreachable")
	}
	if _, ok := self.genericFuncs[name]; ok {
		return false
	}
	self.genericFuncs[name] = symbolGenericObject[parse.Function]{
		pub:  pub,
		data: ast,
	}
	return true
}

// 寻找泛型函数
func (self symbolTable) lookupGenericFunc(name string) (symbolGenericObject[parse.Function], bool) {
	if !self.isPkg() {
		return self.f.lookupGenericFunc(name)
	}
	if v, ok := self.genericFuncs[name]; ok {
		return v, true
	}
	return symbolGenericObject[parse.Function]{}, false
}

// 声明泛型类型
func (self *symbolTable) defGenericType(pub bool, name string, ast parse.TypeDef) bool {
	if !self.isPkg() {
		panic("unreachable")
	}
	if _, ok := self.genericTypes[name]; ok {
		return false
	}
	self.genericTypes[name] = symbolGenericObject[parse.TypeDef]{
		pub:  pub,
		data: ast,
	}
	return true
}

// 寻找泛型类型
func (self symbolTable) lookupGenericType(name string) (symbolGenericObject[parse.TypeDef], bool) {
	if !self.isPkg() {
		return self.f.lookupGenericType(name)
	}
	if v, ok := self.genericTypes[name]; ok {
		return v, true
	}
	return symbolGenericObject[parse.TypeDef]{}, false
}

// 添加泛型类型映射
func (self *symbolTable) addGenericTypeMap(m map[string]hir.Type) {
	self.genericTypeMap.Push(m)
}

// 删除泛型类型映射
func (self *symbolTable) removeGenericTypeMap() {
	self.genericTypeMap.Pop()
}

// 查找泛型类型映射
func (self symbolTable) lookupGenericTypeMap(name string) (hir.Type, bool) {
	if self.genericTypeMap.Empty() && !self.isPkg() {
		return self.f.lookupGenericTypeMap(name)
	}
	if self.genericTypeMap.Empty() {
		return hir.Type{}, false
	}
	maps := self.genericTypeMap.Peek()
	if t, ok := maps[name]; ok {
		return t, true
	}
	if !self.isPkg() {
		return self.f.lookupGenericTypeMap(name)
	}
	return hir.Type{}, false
}

// 声明泛型方法
func (self *symbolTable) defGenericMethod(pub bool, tname, name string, ast parse.Method) bool {
	if !self.isPkg() {
		panic("unreachable")
	}
	realName := fmt.Sprintf("%s::%s", tname, name)
	if _, ok := self.genericMethods[realName]; ok {
		return false
	}
	self.genericMethods[realName] = symbolGenericObject[parse.Method]{
		pub:  pub,
		data: ast,
	}
	return true
}

// 寻找泛型方法
func (self symbolTable) lookupGenericMethod(tname, name string) (symbolGenericObject[parse.Method], bool) {
	if !self.isPkg() {
		return self.f.lookupGenericMethod(tname, name)
	}
	realName := fmt.Sprintf("%s::%s", tname, name)
	if v, ok := self.genericMethods[realName]; ok {
		return v, true
	}
	return symbolGenericObject[parse.Method]{}, false
}
