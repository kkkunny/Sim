package analyse

import (
	stlbasic "github.com/kkkunny/stl/basic"
	"github.com/kkkunny/stl/container/linkedlist"
	stlerror "github.com/kkkunny/stl/error"
	stlos "github.com/kkkunny/stl/os"

	"github.com/kkkunny/Sim/hir"
	"github.com/kkkunny/Sim/parse"
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
