package analyse

import (
	"strconv"
	"strings"

	stlbasic "github.com/kkkunny/stl/basic"
	"github.com/kkkunny/stl/container/either"
	"github.com/kkkunny/stl/container/hashset"
	"github.com/kkkunny/stl/container/linkedlist"
	stlerror "github.com/kkkunny/stl/error"
	stlos "github.com/kkkunny/stl/os"

	"github.com/kkkunny/Sim/ast"
	errors "github.com/kkkunny/Sim/error"
	"github.com/kkkunny/Sim/hir"
	"github.com/kkkunny/Sim/parse"
	"github.com/kkkunny/Sim/reader"
	"github.com/kkkunny/Sim/util"
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
	name = stlbasic.Ternary(name != "", name, pkg.GetPackageName())

	if !importAll && self.pkgScope.externs.ContainKey(name) {
		return linkedlist.LinkedList[hir.Global]{}, importPackageErrorDuplication
	}

	var scope *_PkgScope
	defer func() {
		if err != importPackageErrorNone {
			return
		}
		// 关联包
		if importAll {
			self.pkgScope.links.Add(scope)
		} else {
			self.pkgScope.externs.Set(name, scope)
		}
	}()

	if scope = self.pkgs.Get(pkg); self.pkgs.ContainKey(pkg) && scope == nil {
		// 循环导入
		return linkedlist.LinkedList[hir.Global]{}, importPackageErrorCircular
	} else if self.pkgs.ContainKey(pkg) && scope != nil {
		// 导入过该包，有缓存，不再追加该包的语句
		return linkedlist.LinkedList[hir.Global]{}, importPackageErrorNone
	}

	// 开始分析该包，放一个占位符，以便检测循环导入
	self.pkgs.Set(pkg, nil)
	// 分析并追加该包的语句
	var analyseError stlerror.Error
	hirs, scope, analyseError = analyseSonPackage(self, pkg)
	if analyseError != nil && hirs.Empty() {
		return linkedlist.LinkedList[hir.Global]{}, importPackageErrorInvalid
	}
	self.pkgs.Set(pkg, scope)
	return hirs, importPackageErrorNone
}

func (self *Analyser) setSelfType(td hir.GlobalType) (callback func()) {
	bk := self.selfType
	self.selfType = td
	return func() {
		self.selfType = bk
	}
}

// 获取类型默认值
func (self *Analyser) getTypeDefaultValue(pos reader.Position, t hir.Type) *hir.Default {
	if !t.HasDefault() && !t.EqualTo(hir.NewRefType(false, self.pkgScope.Str())) {
		errors.ThrowCanNotGetDefault(pos, t)
	}
	return &hir.Default{Type: t}
}

// 检查类型定义是否循环
func (self *Analyser) checkTypeDefCircle(trace *hashset.HashSet[hir.Type], t hir.Type) bool {
	if trace.Contain(t) {
		return true
	}
	trace.Add(t)
	defer func() {
		trace.Remove(t)
	}()

	switch typ := t.(type) {
	case *hir.NoThingType, *hir.SintType, *hir.UintType, *hir.FloatType, *hir.FuncType, *hir.RefType, *hir.NoReturnType:
	case *hir.ArrayType:
		return self.checkTypeDefCircle(trace, typ.Elem)
	case *hir.TupleType:
		for _, e := range typ.Elems {
			if self.checkTypeDefCircle(trace, e) {
				return true
			}
		}
	case *hir.StructType:
		for iter := typ.Fields.Iterator(); iter.Next(); {
			if self.checkTypeDefCircle(trace, iter.Value().Second.Type) {
				return true
			}
		}
	case *hir.UnionType:
		for _, e := range typ.Elems {
			if self.checkTypeDefCircle(trace, e) {
				return true
			}
		}
	case *hir.SelfType:
		if self.checkTypeDefCircle(trace, typ.Self.MustValue()) {
			return true
		}
	case *hir.AliasType:
		if self.checkTypeDefCircle(trace, typ.Target) {
			return true
		}
	case *hir.CustomType:
		if self.checkTypeDefCircle(trace, typ.Target) {
			return true
		}
	default:
		panic("unreachable")
	}
	return false
}

// 检查类型别名是否循环
func (self *Analyser) checkTypeAliasCircle(trace *hashset.HashSet[hir.Type], t hir.Type) bool {
	if trace.Contain(t) {
		return true
	}
	trace.Add(t)
	defer func() {
		trace.Remove(t)
	}()

	switch typ := t.(type) {
	case *hir.NoThingType, *hir.SintType, *hir.UintType, *hir.FloatType, *hir.CustomType, *hir.NoReturnType:
	case *hir.FuncType:
		for _, p := range typ.Params {
			if self.checkTypeAliasCircle(trace, p) {
				return true
			}
		}
		return self.checkTypeAliasCircle(trace, typ.Ret)
	case *hir.RefType:
		return self.checkTypeAliasCircle(trace, typ.Elem)
	case *hir.ArrayType:
		return self.checkTypeAliasCircle(trace, typ.Elem)
	case *hir.TupleType:
		for _, e := range typ.Elems {
			if self.checkTypeAliasCircle(trace, e) {
				return true
			}
		}
	case *hir.StructType:
		for iter := typ.Fields.Iterator(); iter.Next(); {
			if self.checkTypeAliasCircle(trace, iter.Value().Second.Type) {
				return true
			}
		}
	case *hir.UnionType:
		for _, e := range typ.Elems {
			if self.checkTypeAliasCircle(trace, e) {
				return true
			}
		}
	case *hir.SelfType:
		if self.checkTypeAliasCircle(trace, typ.Self.MustValue()) {
			return true
		}
	case *hir.AliasType:
		if self.checkTypeAliasCircle(trace, typ.Target) {
			return true
		}
	default:
		panic("unreachable")
	}
	return false
}

func (self *Analyser) isInDstStructScope(st *hir.CustomType) bool {
	if self.selfType == nil {
		return false
	}
	selfName := stlbasic.TernaryAction(!strings.Contains(self.selfType.GetName(), "::"), func() string {
		return self.selfType.GetName()
	}, func() string {
		return strings.Split(self.selfType.GetName(), "::")[0]
	})
	stName := stlbasic.TernaryAction(!strings.Contains(st.GetName(), "::"), func() string {
		return st.GetName()
	}, func() string {
		return strings.Split(st.GetName(), "::")[0]
	})
	return self.selfType.GetPackage().Equal(st.GetPackage()) && selfName == stName
}

// 分析标识符，表达式优先
func (self *Analyser) analyseIdent(node *ast.Ident, flag ...bool) util.Option[either.Either[hir.Ident, hir.Type]] {
	var pkgName string
	if pkgToken, ok := node.Pkg.Value(); ok {
		pkgName = pkgToken.Source()
		if !self.pkgScope.externs.ContainKey(pkgName) {
			errors.ThrowUnknownIdentifierError(pkgToken.Position, pkgToken)
		}
	}

	if len(flag) == 0 || flag[0] {
		// 表达式
		// 标识符表达式
		value, ok := self.localScope.GetValue(pkgName, node.Name.Source())
		if ok {
			return util.Some(either.Left[hir.Ident, hir.Type](value))
		}
	}

	if len(flag) == 0 || !flag[0] {
		// 类型
		name := node.Name.Source()
		// 内置类型
		if name == "X" {
			return util.Some(either.Right[hir.Ident, hir.Type](hir.NoReturn))
		} else if strings.HasPrefix(name, "__buildin_i") {
			bits, err := strconv.ParseUint(name[len("__buildin_i"):], 10, 8)
			if err == nil && bits > 0 && bits <= 128 {
				return util.Some(either.Right[hir.Ident, hir.Type](hir.NewSintType(uint8(bits))))
			}
		} else if strings.HasPrefix(name, "__buildin_u") {
			bits, err := strconv.ParseUint(name[len("__buildin_u"):], 10, 8)
			if err == nil && bits > 0 && bits <= 128 {
				return util.Some(either.Right[hir.Ident, hir.Type](hir.NewUintType(uint8(bits))))
			}
		} else if strings.HasPrefix(name, "__buildin_f") {
			bits, err := strconv.ParseUint(name[len("__buildin_f"):], 10, 8)
			if err == nil && (bits == 16 || bits == 32 || bits == 64 || bits == 128) {
				return util.Some(either.Right[hir.Ident, hir.Type](hir.NewFloatType(uint8(bits))))
			}
		}
		// 类型定义
		if td, ok := self.pkgScope.GetTypeDef(pkgName, name); ok {
			return util.Some(either.Right[hir.Ident, hir.Type](td))
		}
	}
	return util.None[either.Either[hir.Ident, hir.Type]]()
}

func (self *Analyser) analyseFuncBody(node *ast.Block) *hir.Block {
	fn := self.localScope.GetFunc()
	body, jump := self.analyseBlock(node, nil)
	if jump != hir.BlockEofReturn {
		retType := fn.GetFuncType().Ret
		if !hir.IsType[*hir.NoThingType](retType) {
			errors.ThrowMissingReturnValueError(node.Position(), retType)
		}
		body.Stmts.PushBack(&hir.Return{
			Func:  fn,
			Value: util.None[hir.Expr](),
		})
	}
	return body
}

// Analyse 语义分析
func Analyse(path stlos.FilePath) (*hir.Result, stlerror.Error) {
	asts, err := parse.Parse(path)
	if err != nil {
		return nil, err
	}
	return New(asts).Analyse(), nil
}

// 语义分析子包
func analyseSonPackage(parent *Analyser, pkg hir.Package) (linkedlist.LinkedList[hir.Global], *_PkgScope, stlerror.Error) {
	asts, err := parse.Parse(pkg.Path())
	if err != nil {
		return linkedlist.LinkedList[hir.Global]{}, nil, err
	}
	analyser := newSon(parent, asts)
	return analyser.Analyse().Globals, analyser.pkgScope, nil
}
