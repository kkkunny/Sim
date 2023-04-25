package analyse

import (
	"fmt"
	"math"
	"strings"

	"github.com/kkkunny/Sim/src/compiler/hir"
	"github.com/kkkunny/Sim/src/compiler/parse"
	"github.com/kkkunny/Sim/src/compiler/utils"
	"github.com/kkkunny/stl/set"
	"github.com/kkkunny/stl/table"
	"github.com/kkkunny/stl/types"
)

// 类型
func (self *Analyser) analyseType(ast parse.Type) (hir.Type, utils.Error) {
	if ast == nil {
		return hir.NewTypeNone(), nil
	}

	switch typ := ast.(type) {
	case *parse.TypeIdent:
		return self.analyseTypeIdent(*typ)
	case *parse.TypePtr:
		return self.analyseTypePtr(*typ)
	case *parse.TypeFunc:
		return self.analyseTypeFunc(*typ)
	case *parse.TypeArray:
		return self.analyseTypeArray(*typ)
	case *parse.TypeTuple:
		return self.analyseTypeTuple(*typ)
	case *parse.TypeStruct:
		return self.analyseTypeStruct(*typ)
	case *parse.TypeUnion:
		return self.analyseTypeUnion(*typ)
	default:
		panic("unreachable")
	}
}

// 标识符类型
func (self *Analyser) analyseTypeIdent(ast parse.TypeIdent) (hir.Type, utils.Error) {
	// 泛型类型
	if len(ast.GenericArgs) != 0 {
		for _, pkg := range ast.Pkgs {
			symbol := self.symbols[pkg]
			t, ok := symbol.lookupGenericType(ast.Name.Source)
			if !ok {
				continue
			}
			if self.symbol.getPkgSymbolTable() == symbol || t.pub {
				td, err := self.instantiateGenericType(symbol, t.data, ast)
				if err != nil {
					return hir.Type{}, err
				}
				return hir.NewTypeTypedef(td), nil
			}
		}
	} else {
		// 内置类型
		switch ast.Name.Source {
		case "bool":
			return hir.NewTypeBool(), nil
		case "i8":
			return hir.NewTypeI8(), nil
		case "u8":
			return hir.NewTypeU8(), nil
		case "i16":
			return hir.NewTypeI16(), nil
		case "u16":
			return hir.NewTypeU16(), nil
		case "i32":
			return hir.NewTypeI32(), nil
		case "u32":
			return hir.NewTypeU32(), nil
		case "i64":
			return hir.NewTypeI64(), nil
		case "u64":
			return hir.NewTypeU64(), nil
		case "isize":
			return hir.NewTypeIsize(), nil
		case "usize":
			return hir.NewTypeUsize(), nil
		case "f32":
			return hir.NewTypeF32(), nil
		case "f64":
			return hir.NewTypeF64(), nil
		}

		// 泛型类型映射
		if self.symbol.getPkgSymbolTable().pkg.Path == ast.Pkgs[0].Path {
			if t, ok := self.symbol.lookupGenericTypeMap(ast.Name.Source); ok {
				return t, nil
			}
		}

		// 自定义类型
		for _, pkg := range ast.Pkgs {
			symbol := self.symbols[pkg]
			t, ok := symbol.lookupType(ast.Name.Source)
			if !ok {
				continue
			}
			if self.symbol.getPkgSymbolTable() == symbol || t.pub {
				return hir.NewTypeTypedef(t.data), nil
			}
		}
	}

	return hir.Type{}, utils.Errorf(ast.Name.Pos, errUnknownIdentifier)
}

// 实例化泛型类型
func (self *Analyser) instantiateGenericType(symbol *symbolTable, ast parse.TypeDef, ident parse.TypeIdent) (
	*hir.Typedef, utils.Error,
) {
	// 泛型参数
	if len(ident.GenericArgs) != len(ast.GenericParams) {
		return nil, utils.Errorf(
			ident.Position(),
			"expect `%d` type but there is `%d`",
			len(ast.GenericParams),
			len(ident.GenericArgs),
		)
	}
	args := make([]hir.Type, len(ident.GenericArgs))
	argStrs := make([]string, len(ident.GenericArgs))
	errs := make([]utils.Error, 0, len(args))
	for i, a := range ident.GenericArgs {
		at, err := self.analyseType(a)
		if err != nil {
			errs = append(errs, err)
		} else {
			args[i] = at
			argStrs[i] = at.String()
		}
	}
	if len(errs) == 1 {
		return nil, errs[0]
	} else if len(errs) > 1 {
		return nil, utils.NewMultiError(errs...)
	}

	// 切换符号表
	proSymbol := self.symbol
	self.symbol = symbol
	defer func() {
		self.symbol = proSymbol
	}()

	// 类型名重定义
	name := fmt.Sprintf("%s::<%s>", ident.Name.Source, strings.Join(argStrs, ", "))
	if t, ok := self.symbol.lookupType(name); ok {
		return t.data, nil
	}
	ast.Name.Source = name
	defer func() {
		ast.Name.Source = ident.Name.Source
	}()

	// 类型映射
	maps := make(map[string]hir.Type)
	for i, p := range ast.GenericParams {
		maps[p.Source] = args[i]
	}
	symbol.addGenericTypeMap(maps)
	defer func() {
		symbol.removeGenericTypeMap()
	}()

	// 类型声明
	if !self.symbol.declType(
		ast.Public,
		hir.NewTypedef(self.symbol.pkg, ast.Name.Source, hir.NewTypeNone()),
	) {
		return nil, utils.Errorf(ast.Name.Pos, errDuplicateDeclaration)
	}
	// 类型定义
	def, err := self.analyseTypedef(ast)
	if err != nil {
		return nil, err
	}
	def.GenericArgs = table.NewLinkedHashMap[string, hir.Type]()
	for i, a := range args {
		def.GenericArgs.Set(ast.GenericParams[i].Source, a)
	}
	// 循环检测
	if self.checkTypeCircle(set.NewLinkedHashSet[*hir.Typedef](), hir.NewTypeTypedef(def)) {
		errs = append(errs, utils.Errorf(ast.Name.Pos, errCircularReference))
	}
	return def, nil
}

// 指针类型
func (self *Analyser) analyseTypePtr(ast parse.TypePtr) (hir.Type, utils.Error) {
	elem, err := self.analyseType(ast.Elem)
	if err != nil {
		return hir.Type{}, err
	}
	return hir.NewTypePtr(elem), nil
}

// 函数类型
func (self *Analyser) analyseTypeFunc(ast parse.TypeFunc) (hir.Type, utils.Error) {
	// 返回值
	ret, err := self.analyseType(ast.Ret)
	if err != nil {
		return hir.Type{}, nil
	}
	// 参数
	var errs []utils.Error
	params := make([]hir.Type, len(ast.Params))
	for i, p := range ast.Params {
		params[i], err = self.analyseType(p)
		if err != nil {
			errs = append(errs, err)
		}
	}
	if len(errs) == 1 {
		return hir.Type{}, errs[0]
	} else if len(errs) > 1 {
		return hir.Type{}, utils.NewMultiError(errs...)
	}

	return hir.NewTypeFunc(ast.VarArg, ret, params...), nil
}

// 数组类型
func (self *Analyser) analyseTypeArray(ast parse.TypeArray) (hir.Type, utils.Error) {
	// 大小
	size := ast.Size.Value
	if size < 0 || size > math.MaxInt {
		return hir.Type{}, utils.Errorf(ast.Size.Position(), errDataOverflow)
	}
	// 元组类型
	elem, err := self.analyseType(ast.Elem)
	if err != nil {
		return hir.Type{}, nil
	}
	return hir.NewTypeArray(uint(size), elem), nil
}

// 元组类型
func (self *Analyser) analyseTypeTuple(ast parse.TypeTuple) (hir.Type, utils.Error) {
	var errs []utils.Error
	elems := make([]hir.Type, len(ast.Elems))
	for i, e := range ast.Elems {
		elem, err := self.analyseType(e)
		if err != nil {
			errs = append(errs, err)
		} else {
			elems[i] = elem
		}
	}
	if len(errs) == 1 {
		return hir.Type{}, errs[0]
	} else if len(errs) > 1 {
		return hir.Type{}, utils.NewMultiError(errs...)
	}

	return hir.NewTypeTuple(elems...), nil
}

// 结构体类型
func (self *Analyser) analyseTypeStruct(ast parse.TypeStruct) (hir.Type, utils.Error) {
	var errs []utils.Error
	fieldSet := make(map[string]struct{})
	fields := make([]types.ThreePair[bool, string, hir.Type], len(ast.Fields))
	for i, f := range ast.Fields {
		if _, ok := fieldSet[f.Second.Name.Source]; ok {
			errs = append(errs, utils.Errorf(f.Second.Name.Pos, errDuplicateDeclaration))
			continue
		}
		fieldSet[f.Second.Name.Source] = struct{}{}
		field, err := self.analyseType(f.Second.Type)
		if err != nil {
			errs = append(errs, err)
		} else {
			fields[i] = types.NewThreePair(f.First, f.Second.Name.Source, field)
		}
	}
	if len(errs) == 1 {
		return hir.Type{}, errs[0]
	} else if len(errs) > 1 {
		return hir.Type{}, utils.NewMultiError(errs...)
	}

	return hir.NewTypeStruct(fields...), nil
}

// 联合类型
func (self *Analyser) analyseTypeUnion(ast parse.TypeUnion) (hir.Type, utils.Error) {
	elems := make([]hir.Type, len(ast.Elems))
	errs := make([]utils.Error, 0, len(ast.Elems))
	for i, e := range ast.Elems {
		elem, err := self.analyseType(e)
		if err != nil {
			errs = append(errs, err)
		} else {
			elems[i] = elem
		}
	}
	if len(errs) == 1 {
		return hir.Type{}, errs[0]
	} else if len(errs) > 1 {
		return hir.Type{}, utils.NewMultiError(errs...)
	}
	return hir.NewTypeUnion(elems...), nil
}

// 类型循环检测
func (self *Analyser) checkTypeCircle(temp *set.LinkedHashSet[*hir.Typedef], t hir.Type) bool {
	switch t.Kind {
	case hir.TNone, hir.TBool, hir.TI8, hir.TU8, hir.TI16, hir.TU16, hir.TI32, hir.TU32, hir.TI64, hir.TU64, hir.TIsize, hir.TUsize, hir.TF32, hir.TF64:
		return false
	case hir.TPtr:
		// TODO: 允许任何指针类型循环引用
		lastTarget := temp.Last().Target
		if lastTarget.IsTuple() || lastTarget.IsStruct() {
			return false
		}
		return self.checkTypeCircle(temp, t.GetPtr())
	case hir.TFunc:
		// TODO: 允许任何指针类型循环引用
		lastTarget := temp.Last().Target
		if lastTarget.IsTuple() || lastTarget.IsStruct() {
			return false
		}
		if self.checkTypeCircle(temp, t.GetFuncRet()) {
			return true
		}
		for _, p := range t.GetFuncParams() {
			if self.checkTypeCircle(temp, p) {
				return true
			}
		}
		return false
	case hir.TArray:
		return self.checkTypeCircle(temp, t.GetArrayElem())
	case hir.TTuple:
		for _, e := range t.GetTupleElems() {
			if self.checkTypeCircle(temp, e) {
				return true
			}
		}
		return false
	case hir.TStruct:
		for _, f := range t.GetStructFields() {
			if self.checkTypeCircle(temp, f.Third) {
				return true
			}
		}
		return false
	case hir.TUnion:
		for _, e := range t.GetUnionElems() {
			if self.checkTypeCircle(temp, e) {
				return true
			}
		}
		return false
	case hir.TTypedef:
		def := t.GetTypedef()
		if !temp.Add(def) {
			return true
		}
		defer func() {
			temp.Remove(def)
		}()
		return self.checkTypeCircle(temp, def.Target)
	default:
		panic("unreachable")
	}
}
