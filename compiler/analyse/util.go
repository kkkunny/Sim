package analyse

import (
	stlbasic "github.com/kkkunny/stl/basic"
	"github.com/kkkunny/stl/container/hashset"
	"github.com/kkkunny/stl/container/linkedlist"
	stlerror "github.com/kkkunny/stl/error"
	stlos "github.com/kkkunny/stl/os"

	errors "github.com/kkkunny/Sim/error"
	"github.com/kkkunny/Sim/hir"
	"github.com/kkkunny/Sim/parse"
	"github.com/kkkunny/Sim/reader"
)

type importPackageErrorKind uint8

const (
	importPackageErrorNone importPackageErrorKind = iota
	// 循环导入
	importPackageErrorCircular
	// 包名冲突
	importPackageErrorDuplication
	// 无效包
	importPackageErrorInvalid
)

// importPackage 导入包
func (self *Analyser) importPackage(pkg hir.Package, name string, importAll bool) (hirs linkedlist.LinkedList[hir.Global], err importPackageErrorKind) {
	name = stlbasic.Ternary(name!="", name, pkg.GetPackageName())

	if !importAll && self.pkgScope.externs.ContainKey(name){
		return linkedlist.LinkedList[hir.Global]{}, importPackageErrorDuplication
	}

	var scope *_PkgScope
	defer func() {
		if err != importPackageErrorNone{
			return
		}
		// 关联包
		if importAll {
			self.pkgScope.links.Add(scope)
		} else {
			self.pkgScope.externs.Set(name, scope)
		}
	}()

	if scope = self.pkgs.Get(pkg); self.pkgs.ContainKey(pkg) && scope == nil{
		// 循环导入
		return linkedlist.LinkedList[hir.Global]{}, importPackageErrorCircular
	} else if self.pkgs.ContainKey(pkg) && scope != nil{
		// 导入过该包，有缓存，不再追加该包的语句
		return linkedlist.LinkedList[hir.Global]{}, importPackageErrorNone
	}

	// 开始分析该包，放一个占位符，以便检测循环导入
	self.pkgs.Set(pkg, nil)
	// 分析并追加该包的语句
	var analyseError stlerror.Error
	hirs, scope, analyseError = analyseSonPackage(self, pkg)
	if analyseError != nil && hirs.Empty(){
		return linkedlist.LinkedList[hir.Global]{}, importPackageErrorInvalid
	}
	self.pkgs.Set(pkg, scope)
	return hirs, importPackageErrorNone
}

func (self *Analyser) setSelfValue(v *hir.Param)(callback func()){
	bk := self.selfValue
	self.selfValue = v
	return func() {
		self.selfValue = bk
	}
}

func (self *Analyser) setSelfType(td hir.TypeDef)(callback func()){
	bk := self.selfType
	self.selfType = td
	return func() {
		self.selfType = bk
	}
}

// 获取类型默认值
func (self *Analyser) getTypeDefaultValue(pos reader.Position, t hir.Type) hir.Expr{
	if !self.isTypeHasDefault(t){
		errors.ThrowCanNotGetDefault(pos, t)
	}
	return &hir.Default{Type: t}
}

// 类型是否有默认值
func (self *Analyser) isTypeHasDefault(t hir.Type)bool{
	switch tt := t.(type) {
	case *hir.EmptyType, *hir.RefType, *hir.FuncType:
		return false
	case *hir.SintType, *hir.UintType, *hir.FloatType, *hir.BoolType, *hir.StringType, *hir.PtrType:
		return true
	case *hir.ArrayType:
		return self.isTypeHasDefault(tt.Elem)
	case *hir.TupleType:
		for _, e := range tt.Elems{
			if !self.isTypeHasDefault(e){
				return false
			}
		}
		return true
	case *hir.StructType:
		for iter:=tt.Fields.Values().Iterator(); iter.Next(); {
			if !self.isTypeHasDefault(iter.Value().Second){
				return false
			}
		}
		return true
	case *hir.UnionType:
		for _, e := range tt.Elems {
			if !self.isTypeHasDefault(e){
				return false
			}
		}
		return true
	case *hir.SelfType:
		return self.isTypeHasDefault(tt.Self)
	case *hir.AliasType:
		return self.isTypeHasDefault(tt.Target)
	default:
		panic("unreachable")
	}
}

// 检查类型是否循环
func (self *Analyser) checkTypeCircle(trace *hashset.HashSet[hir.Type], t hir.Type)bool{
	if trace.Contain(t){
		return true
	}
	trace.Add(t)
	defer func() {
		trace.Remove(t)
	}()

	switch typ := t.(type) {
	case *hir.EmptyType, *hir.SintType, *hir.UintType, *hir.FloatType, *hir.FuncType, *hir.BoolType, *hir.StringType, *hir.PtrType, *hir.RefType, *hir.GenericIdentType, *hir.GenericStructInst:
	case *hir.ArrayType:
		return self.checkTypeCircle(trace, typ.Elem)
	case *hir.TupleType:
		for _, e := range typ.Elems{
			if self.checkTypeCircle(trace, e){
				return true
			}
		}
	case *hir.StructType:
		for iter:=typ.Fields.Iterator(); iter.Next(); {
			if self.checkTypeCircle(trace, iter.Value().Second.Second){
				return true
			}
		}
	case *hir.UnionType:
		for _, e := range typ.Elems {
			if self.checkTypeCircle(trace, e){
				return true
			}
		}
	case *hir.SelfType:
		if self.checkTypeCircle(trace, typ.Self){
			return true
		}
	case *hir.AliasType:
		if self.checkTypeCircle(trace, typ.Target){
			return true
		}
	default:
		panic("unreachable")
	}
	return false
}

// Analyse 语义分析
func Analyse(path stlos.FilePath) (linkedlist.LinkedList[hir.Global], stlerror.Error) {
	asts, err := parse.Parse(path)
	if err != nil{
		return linkedlist.LinkedList[hir.Global]{}, err
	}
	return New(asts).Analyse(), nil
}

// 语义分析子包
func analyseSonPackage(parent *Analyser, pkg hir.Package) (linkedlist.LinkedList[hir.Global], *_PkgScope, stlerror.Error) {
	asts, err := parse.Parse(pkg.Path())
	if err != nil{
		return linkedlist.LinkedList[hir.Global]{}, nil, err
	}
	analyser := newSon(parent, asts)
	return analyser.Analyse(), analyser.pkgScope, nil
}
