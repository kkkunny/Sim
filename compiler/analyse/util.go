package analyse

import (
	"strconv"
	"strings"

	"github.com/kkkunny/stl/container/either"
	stliter "github.com/kkkunny/stl/container/iter"
	"github.com/kkkunny/stl/container/linkedlist"
	"github.com/kkkunny/stl/container/optional"
	"github.com/kkkunny/stl/container/set"
	stlslices "github.com/kkkunny/stl/container/slices"
	"github.com/kkkunny/stl/container/tuple"
	stlos "github.com/kkkunny/stl/os"
	stlval "github.com/kkkunny/stl/value"

	"github.com/kkkunny/Sim/compiler/ast"
	"github.com/kkkunny/Sim/compiler/oldhir"

	"github.com/kkkunny/Sim/compiler/parse"

	"github.com/kkkunny/Sim/compiler/reader"

	errors "github.com/kkkunny/Sim/compiler/error"
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
func (self *Analyser) importPackage(pkg oldhir.Package, name string, importAll bool) (hirs linkedlist.LinkedList[oldhir.Global], err importPackageErrorKind) {
	name = stlval.Ternary(name != "", name, pkg.GetPackageName())

	if !importAll && self.pkgScope.externs.Contain(name) {
		return linkedlist.LinkedList[oldhir.Global]{}, importPackageErrorDuplication
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

	if scope = self.pkgs.Get(pkg); self.pkgs.Contain(pkg) && scope == nil {
		// 循环导入
		return linkedlist.LinkedList[oldhir.Global]{}, importPackageErrorCircular
	} else if self.pkgs.Contain(pkg) && scope != nil {
		// 导入过该包，有缓存，不再追加该包的语句
		return linkedlist.LinkedList[oldhir.Global]{}, importPackageErrorNone
	}

	// 开始分析该包，放一个占位符，以便检测循环导入
	self.pkgs.Set(pkg, nil)
	// 分析并追加该包的语句
	var analyseError error
	hirs, scope, analyseError = analyseSonPackage(self, pkg)
	if analyseError != nil && hirs.Empty() {
		return linkedlist.LinkedList[oldhir.Global]{}, importPackageErrorInvalid
	}
	self.pkgs.Set(pkg, scope)
	return hirs, importPackageErrorNone
}

func (self *Analyser) setSelfType(td *oldhir.CustomType) (callback func()) {
	bk := self.selfType
	self.selfType = td
	return func() {
		self.selfType = bk
	}
}

// 获取类型默认值
func (self *Analyser) getTypeDefaultValue(pos reader.Position, t oldhir.Type) *oldhir.Default {
	if !self.hasTypeDefault(t) {
		errors.ThrowCanNotGetDefault(pos, t)
	}
	return &oldhir.Default{Type: t}
}

// 检查类型定义是否循环
func (self *Analyser) checkTypeDefCircle(trace set.Set[oldhir.Type], t oldhir.Type) bool {
	if trace.Contain(t) {
		return true
	}
	trace.Add(t)
	defer func() {
		trace.Remove(t)
	}()

	switch typ := t.(type) {
	case *oldhir.NoThingType, *oldhir.SintType, *oldhir.UintType, *oldhir.FloatType, *oldhir.FuncType, *oldhir.RefType, *oldhir.NoReturnType, *oldhir.LambdaType:
	case *oldhir.ArrayType:
		return self.checkTypeDefCircle(trace, typ.Elem)
	case *oldhir.TupleType:
		for _, e := range typ.Elems {
			if self.checkTypeDefCircle(trace, e) {
				return true
			}
		}
	case *oldhir.StructType:
		for iter := typ.Fields.Iterator(); iter.Next(); {
			if self.checkTypeDefCircle(trace, iter.Value().E2().Type) {
				return true
			}
		}
	case *oldhir.AliasType:
		if self.checkTypeDefCircle(trace, typ.Target) {
			return true
		}
	case *oldhir.CustomType:
		if self.checkTypeDefCircle(trace, typ.Target) {
			return true
		}
	case *oldhir.EnumType:
		for iter := typ.Fields.Iterator(); iter.Next(); {
			elemOp := iter.Value().E2().Elem
			if elem, ok := elemOp.Value(); ok {
				if self.checkTypeDefCircle(trace, elem) {
					return true
				}
			}
		}
	default:
		panic("unreachable")
	}
	return false
}

// 检查类型别名是否循环
func (self *Analyser) checkTypeAliasCircle(trace set.Set[oldhir.Type], t oldhir.Type) bool {
	if trace.Contain(t) {
		return true
	}
	trace.Add(t)
	defer func() {
		trace.Remove(t)
	}()

	switch typ := t.(type) {
	case *oldhir.NoThingType, *oldhir.SintType, *oldhir.UintType, *oldhir.FloatType, *oldhir.CustomType, *oldhir.NoReturnType:
	case *oldhir.FuncType:
		for _, p := range typ.Params {
			if self.checkTypeAliasCircle(trace, p) {
				return true
			}
		}
		return self.checkTypeAliasCircle(trace, typ.Ret)
	case *oldhir.LambdaType:
		for _, p := range typ.Params {
			if self.checkTypeAliasCircle(trace, p) {
				return true
			}
		}
		return self.checkTypeAliasCircle(trace, typ.Ret)
	case *oldhir.RefType:
		return self.checkTypeAliasCircle(trace, typ.Elem)
	case *oldhir.ArrayType:
		return self.checkTypeAliasCircle(trace, typ.Elem)
	case *oldhir.TupleType:
		for _, e := range typ.Elems {
			if self.checkTypeAliasCircle(trace, e) {
				return true
			}
		}
	case *oldhir.StructType:
		for iter := typ.Fields.Iterator(); iter.Next(); {
			if self.checkTypeAliasCircle(trace, iter.Value().E2().Type) {
				return true
			}
		}
	case *oldhir.AliasType:
		if self.checkTypeAliasCircle(trace, typ.Target) {
			return true
		}
	case *oldhir.EnumType:
		for iter := typ.Fields.Iterator(); iter.Next(); {
			elemOp := iter.Value().E2().Elem
			if elem, ok := elemOp.Value(); ok {
				if self.checkTypeDefCircle(trace, elem) {
					return true
				}
			}
		}
	default:
		panic("unreachable")
	}
	return false
}

func (self *Analyser) isInDstStructScope(st *oldhir.CustomType) bool {
	if self.selfType == nil {
		return false
	}
	selfName := stlval.TernaryAction(!strings.Contains(self.selfType.GetName(), "::"), func() string {
		return self.selfType.GetName()
	}, func() string {
		return strings.Split(self.selfType.GetName(), "::")[0]
	})
	stName := stlval.TernaryAction(!strings.Contains(st.GetName(), "::"), func() string {
		return st.GetName()
	}, func() string {
		return strings.Split(st.GetName(), "::")[0]
	})
	return self.selfType.GetPackage().Equal(st.GetPackage()) && selfName == stName
}

// 分析标识符，表达式优先
func (self *Analyser) analyseIdent(node *ast.Ident, flag ...bool) optional.Optional[either.Either[oldhir.Ident, oldhir.Type]] {
	var pkgName string
	if pkgToken, ok := node.Pkg.Value(); ok {
		pkgName = pkgToken.Source()
		if !self.pkgScope.externs.Contain(pkgName) {
			errors.ThrowUnknownIdentifierError(pkgToken.Position, pkgToken)
		}
	}

	if len(flag) == 0 || flag[0] {
		// 表达式
		// 标识符表达式
		value := stlval.TernaryAction(self.localScope == nil, func() tuple.Tuple2[oldhir.Ident, bool] {
			return tuple.Pack2[oldhir.Ident, bool](self.pkgScope.GetValue(pkgName, node.Name.Source()))
		}, func() tuple.Tuple2[oldhir.Ident, bool] {
			return tuple.Pack2[oldhir.Ident, bool](self.localScope.GetValue(pkgName, node.Name.Source()))
		})
		if value.E2() {
			return optional.Some(either.Left[oldhir.Ident, oldhir.Type](value.E1()))
		}
	}

	if len(flag) == 0 || !flag[0] {
		// 类型
		name := node.Name.Source()
		// 内置类型
		if self.pkgScope.pkg.Equal(oldhir.BuildInPackage) && strings.HasPrefix(name, "__buildin_i") {
			bits, err := strconv.ParseUint(name[len("__buildin_i"):], 10, 8)
			if err == nil && bits > 0 && bits <= 128 {
				return optional.Some(either.Right[oldhir.Ident, oldhir.Type](oldhir.NewSintType(uint8(bits))))
			}
		} else if self.pkgScope.pkg.Equal(oldhir.BuildInPackage) && strings.HasPrefix(name, "__buildin_u") {
			bits, err := strconv.ParseUint(name[len("__buildin_u"):], 10, 8)
			if err == nil && bits > 0 && bits <= 128 {
				return optional.Some(either.Right[oldhir.Ident, oldhir.Type](oldhir.NewUintType(uint8(bits))))
			}
		} else if self.pkgScope.pkg.Equal(oldhir.BuildInPackage) && strings.HasPrefix(name, "__buildin_f") {
			bits, err := strconv.ParseUint(name[len("__buildin_f"):], 10, 8)
			if err == nil && (bits == 16 || bits == 32 || bits == 64 || bits == 128) {
				return optional.Some(either.Right[oldhir.Ident, oldhir.Type](oldhir.NewFloatType(uint8(bits))))
			}
		}
		// 类型定义
		if td, ok := self.pkgScope.GetTypeDef(pkgName, name); ok {
			return optional.Some(either.Right[oldhir.Ident, oldhir.Type](td))
		}
	}
	return optional.None[either.Either[oldhir.Ident, oldhir.Type]]()
}

func (self *Analyser) analyseFuncBody(node *ast.Block) *oldhir.Block {
	fn := self.localScope.GetFunc()
	body, jump := self.analyseBlock(node, nil)
	if jump != oldhir.BlockEofReturn {
		retType := fn.GetFuncType().Ret
		if !retType.EqualTo(oldhir.NoThing) {
			errors.ThrowMissingReturnValueError(node.Position(), retType)
		}
		body.Stmts.PushBack(oldhir.NewReturn(fn))
	}
	return body
}

func (self *Analyser) analyseFuncDecl(node ast.FuncDecl) oldhir.FuncDecl {
	paramNameSet := set.StdHashSetWith[string]()
	params := stlslices.Map(node.Params, func(_ int, e ast.Param) *oldhir.Param {
		param := self.analyseParam(e)
		if param.Name.IsSome() && !paramNameSet.Add(param.Name.MustValue()) {
			errors.ThrowIdentifierDuplicationError(e.Name.MustValue().Position, e.Name.MustValue())
		}
		return param
	})
	return oldhir.FuncDecl{
		Name:   node.Name.Source(),
		Params: params,
		Ret:    self.analyseOptionTypeWith(node.Ret, noReturnTypeAnalyser),
	}
}

// 类型是否有默认值
func (self *Analyser) hasTypeDefault(t oldhir.Type) bool {
	switch tt := t.(type) {
	case *oldhir.NoThingType, *oldhir.NoReturnType, *oldhir.RefType:
		return false
	case *oldhir.SintType, *oldhir.UintType, *oldhir.FloatType:
		return true
	case *oldhir.FuncType:
		if tt.Ret.EqualTo(oldhir.NoThing) {
			return true
		}
		return self.hasTypeDefault(tt.Ret)
	case *oldhir.LambdaType:
		if tt.Ret.EqualTo(oldhir.NoThing) {
			return true
		}
		return self.hasTypeDefault(tt.Ret)
	case *oldhir.CustomType:
		if self.pkgScope.Default().HasBeImpled(t) {
			return true
		}
		return self.hasTypeDefault(tt.Target)
	case *oldhir.AliasType:
		return self.hasTypeDefault(tt.Target)
	case *oldhir.ArrayType:
		return self.hasTypeDefault(tt.Elem)
	case *oldhir.TupleType:
		return stlslices.All(tt.Elems, func(_ int, e oldhir.Type) bool {
			return self.hasTypeDefault(e)
		})
	case *oldhir.StructType:
		return stliter.All(tt.Fields, func(e tuple.Tuple2[string, oldhir.Field]) bool {
			return self.hasTypeDefault(e.E2().Type)
		})
	case *oldhir.EnumType:
		return stliter.Any(tt.Fields, func(p tuple.Tuple2[string, oldhir.EnumField]) bool {
			return p.E2().Elem.IsNone() || self.hasTypeDefault(p.E2().Elem.MustValue())
		})
	default:
		panic("unreachable")
	}
}

// Analyse 语义分析
func Analyse(path stlos.FilePath) (*oldhir.Result, error) {
	asts, err := parse.Parse(path)
	if err != nil {
		return nil, err
	}
	return New(asts).Analyse(), nil
}

// 语义分析子包
func analyseSonPackage(parent *Analyser, pkg oldhir.Package) (linkedlist.LinkedList[oldhir.Global], *_PkgScope, error) {
	asts, err := parse.Parse(pkg.Path())
	if err != nil {
		return linkedlist.LinkedList[oldhir.Global]{}, nil, err
	}
	analyser := newSon(parent, asts)
	return analyser.Analyse().Globals, analyser.pkgScope, nil
}
