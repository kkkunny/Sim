package analyse

import (
	"strings"

	stlbasic "github.com/kkkunny/stl/basic"
	"github.com/kkkunny/stl/container/either"
	"github.com/kkkunny/stl/container/hashset"
	"github.com/kkkunny/stl/container/linkedlist"
	stlerror "github.com/kkkunny/stl/error"
	stlos "github.com/kkkunny/stl/os"
	stlslices "github.com/kkkunny/stl/slices"

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
			if !self.isTypeHasDefault(iter.Value().Type){
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
	case *hir.GenericIdentType:
		return false
	case *hir.GenericStructInst:
		return self.isTypeHasDefault(tt.StructType())
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
			if self.checkTypeCircle(trace, iter.Value().Second.Type){
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

func (self *Analyser) isInDstStructScope(st *hir.StructType)bool{
	if self.selfType == nil{
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

func (self *Analyser) analyseIdent(node *ast.Ident, flag ...bool) util.Option[either.Either[hir.Expr, hir.Type]] {
	var pkgName string
	if pkgToken, ok := node.Pkg.Value(); ok {
		pkgName = pkgToken.Source()
		if !self.pkgScope.externs.ContainKey(pkgName) {
			errors.ThrowUnknownIdentifierError(pkgToken.Position, pkgToken)
		}
	}

	if genericArgs, ok := node.Name.Params.Value(); !ok{
		if len(flag) == 0 || !flag[0]{
			// 类型
			switch name := node.Name.Name.Source(); name {
			case "isize":
				return util.Some(either.Right[hir.Expr, hir.Type](hir.Isize))
			case "i8":
				return util.Some(either.Right[hir.Expr, hir.Type](hir.I8))
			case "i16":
				return util.Some(either.Right[hir.Expr, hir.Type](hir.I16))
			case "i32":
				return util.Some(either.Right[hir.Expr, hir.Type](hir.I32))
			case "i64":
				return util.Some(either.Right[hir.Expr, hir.Type](hir.I64))
			case "usize":
				return util.Some(either.Right[hir.Expr, hir.Type](hir.Usize))
			case "u8":
				return util.Some(either.Right[hir.Expr, hir.Type](hir.U8))
			case "u16":
				return util.Some(either.Right[hir.Expr, hir.Type](hir.U16))
			case "u32":
				return util.Some(either.Right[hir.Expr, hir.Type](hir.U32))
			case "u64":
				return util.Some(either.Right[hir.Expr, hir.Type](hir.U64))
			case "f32":
				return util.Some(either.Right[hir.Expr, hir.Type](hir.F32))
			case "f64":
				return util.Some(either.Right[hir.Expr, hir.Type](hir.F64))
			case "bool":
				return util.Some(either.Right[hir.Expr, hir.Type](hir.Bool))
			case "str":
				return util.Some(either.Right[hir.Expr, hir.Type](hir.Str))
			default:
				// 泛型标识符类型
				if pkgName == ""{
					if to := self.genericIdentMap.Get(name); to != nil {
						return util.Some(either.Right[hir.Expr, hir.Type](to))
					}
				}
				// 类型定义
				if td, ok := self.pkgScope.GetTypeDef(pkgName, name); ok {
					return util.Some(either.Right[hir.Expr, hir.Type](td))
				}
			}
		}

		if len(flag) == 0 || flag[0]{
			// 表达式
			// 标识符表达式
			value, ok := self.localScope.GetValue(pkgName, node.Name.Name.Source())
			if ok {
				return util.Some(either.Left[hir.Expr, hir.Type](value))
			}
		}
	}else{
		if len(flag) == 0 || !flag[0]{
			// 泛型结构体
			if st, ok := self.pkgScope.GetGenericStructDef(pkgName, node.Name.Name.Source()); ok{
				if st.GenericParams.Length() != uint(len(genericArgs.Data)){
					errors.ThrowParameterNumberNotMatchError(genericArgs.Position(), st.GenericParams.Length(), uint(len(genericArgs.Data)))
				}
				params := stlslices.Map(genericArgs.Data, func(_ int, e ast.Type) hir.Type {
					return self.analyseType(e)
				})
				instType := &hir.GenericStructInst{
					Define: st,
					Params: params,
				}

				if self.checkTypeCircle(stlbasic.Ptr(hashset.NewHashSet[hir.Type]()), instType){
					errors.ThrowCircularReference(node.Name.Position(), node.Name.Name)
				}

				return util.Some(either.Right[hir.Expr, hir.Type](instType))
			}
		}

		if len(flag) == 0 || flag[0]{
			// 表达式
			// 泛型标识符表达式实例化
			f, ok := self.pkgScope.GetGenericFuncDef(pkgName, node.Name.Name.Source())
			if ok{
				if f.GenericParams.Length() != uint(len(genericArgs.Data)){
					errors.ThrowParameterNumberNotMatchError(genericArgs.Position(), f.GenericParams.Length(), uint(len(genericArgs.Data)))
				}
				params := stlslices.Map(genericArgs.Data, func(_ int, e ast.Type) hir.Type {
					return self.analyseType(e)
				})
				return util.Some(either.Left[hir.Expr, hir.Type](&hir.GenericFuncInst{
					Define: f,
					Params: params,
				}))
			}
		}
	}
	return util.None[either.Either[hir.Expr, hir.Type]]()
}

func (self *Analyser) analyseFuncBody(node *ast.Block)*hir.Block{
	fn := self.localScope.GetFunc()
	body, jump := self.analyseBlock(node, nil)
	if jump != hir.BlockEofReturn {
		retType := fn.GetFuncType().Ret
		if !hir.IsEmptyType(retType) {
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
