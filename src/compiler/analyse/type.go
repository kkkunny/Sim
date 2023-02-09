package analyse

import (
	. "github.com/kkkunny/Sim/src/compiler/hir"
	"github.com/kkkunny/Sim/src/compiler/lex"
	"github.com/kkkunny/Sim/src/compiler/parse"
	"github.com/kkkunny/Sim/src/compiler/utils"
	stlos "github.com/kkkunny/stl/os"
	"github.com/kkkunny/stl/set"
	"github.com/kkkunny/stl/table"
	"github.com/kkkunny/stl/types"
)

// 类型
func analyseType(ctx *packageContext, ast parse.Type) (Type, utils.Error) {
	if ast == nil {
		return None, nil
	}
	switch typ := ast.(type) {
	case *parse.TypeIdent:
		return analyseTypeIdent(ctx, typ.Pkg.Path, typ.Name)
	case *parse.TypeFunc:
		ret, err := analyseType(ctx, typ.Ret)
		if err != nil {
			return nil, err
		}
		params := make([]Type, len(typ.Params))
		var errors []utils.Error
		for i, p := range typ.Params {
			param, err := analyseType(ctx, p)
			if err != nil {
				errors = append(errors, err)
			} else {
				params[i] = param
			}
		}
		if len(errors) == 0 {
			return NewFuncType(ret, params, typ.VarArg), nil
		} else if len(errors) == 1 {
			return nil, errors[0]
		} else {
			return nil, utils.NewMultiError(errors...)
		}
	case *parse.TypeArray:
		elem, err := analyseType(ctx, typ.Elem)
		if err != nil {
			return nil, err
		}
		return NewArrayType(uint(typ.Size.Value), elem), nil
	case *parse.TypeTuple:
		elems := make([]Type, len(typ.Elems))
		var errors []utils.Error
		for i, e := range typ.Elems {
			elem, err := analyseType(ctx, e)
			if err != nil {
				errors = append(errors, err)
			} else {
				elems[i] = elem
			}
		}
		if len(errors) == 0 {
			return NewTupleType(elems...), nil
		} else if len(errors) == 1 {
			return nil, errors[0]
		} else {
			return nil, utils.NewMultiError(errors...)
		}
	case *parse.TypeStruct:
		fields := table.NewLinkedHashMap[string, types.Pair[bool, Type]]()
		var errors []utils.Error
		for _, f := range typ.Fields {
			ft, err := analyseType(ctx, f.Second.Type)
			if err != nil {
				errors = append(errors, err)
			} else if fields.ContainKey(f.Second.Name.Source) {
				errors = append(errors, utils.Errorf(f.Second.Name.Pos, "duplicate identifier"))
			} else {
				fields.Set(f.Second.Name.Source, types.NewPair(f.First, ft))
			}
		}
		if len(errors) == 0 {
			return NewStructType(fields), nil
		} else if len(errors) == 1 {
			return nil, errors[0]
		} else {
			return nil, utils.NewMultiError(errors...)
		}
	case *parse.TypePtr:
		elem, err := analyseType(ctx, typ.Elem)
		if err != nil {
			return nil, err
		}
		return NewPtrType(elem), nil
	case *parse.TypeInterface:
		fields := table.NewLinkedHashMap[string, *TypeFunc]()
		var errs []utils.Error
		for _, f := range typ.Fields {
			t, err := analyseType(ctx, f.Type)
			if err != nil {
				errs = append(errs, err)
			} else if fields.ContainKey(f.Name.Source) {
				errs = append(errs, utils.Errorf(f.Name.Pos, "duplicate identifier"))
			} else if ft, ok := t.(*TypeFunc); !ok {
				errs = append(errs, utils.Errorf(f.Name.Pos, "expect a function type"))
			} else {
				fields.Set(f.Name.Source, ft)
			}
		}
		if len(errs) == 0 {
			return NewTypeInterface(fields), nil
		} else if len(errs) == 1 {
			return nil, errs[0]
		} else {
			return nil, utils.NewMultiError(errs...)
		}
	default:
		panic("")
	}
}

// 检查类型循环引用
// 只允许元组和结构体循环引用指针
func checkTypeCircle(tmp *set.LinkedHashSet[*Typedef], t Type) bool {
	if t == nil {
		return false
	}
	switch typ := t.(type) {
	case *TypeBasic:
		return false
	case *TypeFunc:
		if IsTupleType(tmp.Last().Dst) || IsStructType(tmp.Last().Dst) {
			return false
		}
		if checkTypeCircle(tmp, typ.Ret) {
			return true
		}
		for _, p := range typ.Params {
			if checkTypeCircle(tmp, p) {
				return true
			}
		}
		return false
	case *TypePtr:
		if IsTupleType(tmp.Last().Dst) || IsStructType(tmp.Last().Dst) {
			return false
		}
		return checkTypeCircle(tmp, typ.Elem)
	case *TypeArray:
		return checkTypeCircle(tmp, typ.Elem)
	case *TypeTuple:
		for _, e := range typ.Elems {
			if checkTypeCircle(tmp, e) {
				return true
			}
		}
		return false
	case *TypeStruct:
		for iter := typ.Fields.Begin(); iter.HasValue(); iter.Next() {
			if checkTypeCircle(tmp, iter.Value().Second) {
				return true
			}
		}
		return false
	case *Typedef:
		if !tmp.Add(typ) {
			return true
		}
		defer func() {
			tmp.Remove(typ)
		}()
		return checkTypeCircle(tmp, typ.Dst)
	case *TypeInterface:
		for iter := typ.Fields.Begin(); iter.HasValue(); iter.Next() {
			if checkTypeCircle(tmp, iter.Value()) {
				return true
			}
		}
		return false
	default:
		panic("")
	}
}

// 标识符类型
func analyseTypeIdent(ctx *packageContext, pkgPath stlos.Path, name lex.Token) (Type, utils.Error) {
	if pkgPath == ctx.path {
		switch name.Source {
		case "i8":
			return I8, nil
		case "i16":
			return I16, nil
		case "i32":
			return I32, nil
		case "i64":
			return I64, nil
		case "isize":
			return Isize, nil
		case "u8":
			return U8, nil
		case "u16":
			return U16, nil
		case "u32":
			return U32, nil
		case "u64":
			return U64, nil
		case "usize":
			return Usize, nil
		case "f32":
			return F32, nil
		case "f64":
			return F64, nil
		case "bool":
			return Bool, nil
		}
	}

	// 类型定义
	pkg := ctx.f.Pkgs[pkgPath]
	if td, ok := pkg.typedefs[name.Source]; ok && (pkgPath == ctx.path || td.First) {
		return td.Second, nil
	}
	return nil, utils.Errorf(name.Pos, "unknown identifier")
}
