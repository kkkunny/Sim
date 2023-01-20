package analyse

import (
	"fmt"
	"github.com/kkkunny/Sim/src/compiler/parse"
	"github.com/kkkunny/Sim/src/compiler/utils"
	stlos "github.com/kkkunny/stl/os"
	"github.com/kkkunny/stl/set"
	"github.com/kkkunny/stl/table"
	"github.com/kkkunny/stl/types"
	"strings"
)

type Type interface {
	fmt.Stringer
	Equal(Type) bool
}

var (
	None = &typeBasic{Name: "none"}

	I8    = &typeBasic{Name: "i8"}
	I16   = &typeBasic{Name: "i16"}
	I32   = &typeBasic{Name: "i32"}
	I64   = &typeBasic{Name: "i64"}
	Isize = &typeBasic{Name: "isize"}

	U8    = &typeBasic{Name: "u8"}
	U16   = &typeBasic{Name: "u16"}
	U32   = &typeBasic{Name: "u32"}
	U64   = &typeBasic{Name: "u64"}
	Usize = &typeBasic{Name: "usize"}

	F32 = &typeBasic{Name: "f32"}
	F64 = &typeBasic{Name: "f64"}

	Bool = &typeBasic{Name: "bool"}
)

// typeBasic 基础类型
type typeBasic struct {
	Name string
}

// IsBasicType 是否是基础类型
func IsBasicType(t Type) bool {
	_, ok := t.(*typeBasic)
	return ok
}

// IsNoneType 是否是空类型
func IsNoneType(t Type) bool {
	return t == None
}

// IsNumberType 是否是数字类型
func IsNumberType(t Type) bool {
	return IsIntType(t) || IsFloatType(t)
}

// IsNumberTypeAndSon 是否是数字类型及其子类型
func IsNumberTypeAndSon(t Type) bool {
	return IsNumberType(GetBaseType(t))
}

// IsIntType 是否是整型
func IsIntType(t Type) bool {
	return IsSintType(t) || IsUintType(t)
}

// IsIntTypeAndSon 是否是整型及其子类型
func IsIntTypeAndSon(t Type) bool {
	return IsIntType(GetBaseType(t))
}

// IsSintType 是否是有符号整型
func IsSintType(t Type) bool {
	return t == I8 || t == I16 || t == I32 || t == I64 || t == Isize
}

// IsSintTypeAndSon 是否是有符号整型及其子类型
func IsSintTypeAndSon(t Type) bool {
	return IsSintType(GetBaseType(t))
}

// IsUintType 是否是无符号整型
func IsUintType(t Type) bool {
	return t == U8 || t == U16 || t == U32 || t == U64 || t == Usize
}

// IsUintTypeAndSon 是否是无符号整型及其子类型
func IsUintTypeAndSon(t Type) bool {
	return IsUintType(GetBaseType(t))
}

// IsFloatType 是否是浮点型
func IsFloatType(t Type) bool {
	return t == F32 || t == F64
}

// IsFloatTypeAndSon 是否是浮点型及其子类型
func IsFloatTypeAndSon(t Type) bool {
	return IsFloatType(GetBaseType(t))
}

// IsBoolType 是否是布尔类型
func IsBoolType(t Type) bool {
	return t == Bool
}

// IsBoolTypeAndSon 是否是布尔类型及其子类型
func IsBoolTypeAndSon(t Type) bool {
	return IsBoolType(GetBaseType(t))
}

func (self typeBasic) String() string {
	return self.Name
}

func (self typeBasic) Equal(t Type) bool {
	if b, ok := t.(*typeBasic); ok {
		return self.Name == b.Name
	}
	return false
}

// TypeFunc 函数类型
type TypeFunc struct {
	Ret    Type
	Params []Type
	VarArg bool
}

// NewFuncType 新建函数类型
func NewFuncType(ret Type, params []Type, varArg bool) *TypeFunc {
	return &TypeFunc{
		Ret:    ret,
		Params: params,
		VarArg: varArg,
	}
}

// IsFuncType 是否是函数类型
func IsFuncType(t Type) bool {
	_, ok := t.(*TypeFunc)
	return ok
}

// IsFuncTypeAndSon 是否是函数类型及其子类型
func IsFuncTypeAndSon(t Type) bool {
	return IsFuncType(GetBaseType(t))
}

func (self TypeFunc) String() string {
	var buf strings.Builder
	buf.WriteString("func(")
	buf.WriteByte(')')
	buf.WriteString(self.Ret.String())
	return buf.String()
}

func (self TypeFunc) Equal(t Type) bool {
	if f, ok := t.(*TypeFunc); ok && self.Ret.Equal(f.Ret) && len(self.Params) == len(f.Params) && self.VarArg == f.VarArg {
		for i, p := range self.Params {
			if !p.Equal(f.Params[i]) {
				return false
			}
		}
		return true
	}
	return false
}

// TypeArray 数组类型
type TypeArray struct {
	Size uint
	Elem Type
}

// NewArrayType 新建数组类型
func NewArrayType(size uint, elem Type) *TypeArray {
	return &TypeArray{
		Size: size,
		Elem: elem,
	}
}

// IsArrayType 是否是数组类型
func IsArrayType(t Type) bool {
	_, ok := t.(*TypeArray)
	return ok
}

// IsArrayTypeAndSon 是否是数组类型及其子类型
func IsArrayTypeAndSon(t Type) bool {
	return IsArrayType(GetBaseType(t))
}

func (self TypeArray) String() string {
	return fmt.Sprintf("[%d]%s", self.Size, self.Elem)
}

func (self TypeArray) Equal(t Type) bool {
	if a, ok := t.(*TypeArray); ok {
		return self.Size == a.Size && self.Elem.Equal(a.Elem)
	}
	return false
}

// TypeTuple 元组类型
type TypeTuple struct {
	Elems []Type
}

// NewTupleType 新建元组类型
func NewTupleType(elems ...Type) *TypeTuple {
	return &TypeTuple{Elems: elems}
}

// IsTupleType 是否是元组类型
func IsTupleType(t Type) bool {
	_, ok := t.(*TypeTuple)
	return ok
}

// IsTupleTypeAndSon 是否是元组类型及其子类型
func IsTupleTypeAndSon(t Type) bool {
	return IsTupleType(GetBaseType(t))
}

func (self TypeTuple) String() string {
	types := make([]string, len(self.Elems))
	for i, t := range self.Elems {
		types[i] = t.String()
	}
	return fmt.Sprintf("(%s)", strings.Join(types, ","))
}

func (self TypeTuple) Equal(t Type) bool {
	if t, ok := t.(*TypeTuple); ok {
		if len(self.Elems) != len(t.Elems) {
			return false
		}
		for i, e := range self.Elems {
			if !e.Equal(t.Elems[i]) {
				return false
			}
		}
		return true
	}
	return false
}

// TypeStruct 结构体类型
type TypeStruct struct {
	Fields *table.LinkedHashMap[string, types.Pair[bool, Type]]
}

// NewStructType 新建结构体类型
func NewStructType(fields *table.LinkedHashMap[string, types.Pair[bool, Type]]) *TypeStruct {
	return &TypeStruct{Fields: fields}
}

// IsStructType 是否是结构体类型
func IsStructType(t Type) bool {
	_, ok := t.(*TypeStruct)
	return ok
}

// IsStructTypeAndSon 是否是结构体类型及其子类型
func IsStructTypeAndSon(t Type) bool {
	return IsStructType(GetBaseType(t))
}

func (self TypeStruct) String() string {
	var buf strings.Builder
	buf.WriteByte('{')
	for iter := self.Fields.Begin(); iter.HasValue(); iter.Next() {
		buf.WriteString(fmt.Sprintf("%s: %s", iter.Key(), iter.Value()))
		if iter.HasNext() {
			buf.WriteString(", ")
		}
	}
	buf.WriteByte('}')
	return buf.String()
}

func (self TypeStruct) Equal(t Type) bool {
	if s, ok := t.(*TypeStruct); ok {
		if self.Fields.Length() != s.Fields.Length() {
			return false
		}
		for iter := self.Fields.Begin(); iter.HasValue(); iter.Next() {
			sk, sv := s.Fields.GetByIndex(iter.Index())
			if iter.Key() != sk || iter.Value().First != sv.First || !iter.Value().Second.Equal(sv.Second) {
				return false
			}
		}
		return true
	}
	return false
}

// TypePtr 指针类型
type TypePtr struct {
	Elem Type
}

// NewPtrType 新建指针类型
func NewPtrType(elem Type) *TypePtr {
	return &TypePtr{
		Elem: elem,
	}
}

// IsPtrType 是否是指针类型
func IsPtrType(t Type) bool {
	_, ok := t.(*TypePtr)
	return ok
}

// IsPtrTypeAndSon 是否是指针类型及其子类型
func IsPtrTypeAndSon(t Type) bool {
	return IsPtrType(GetBaseType(t))
}

func (self TypePtr) String() string {
	return "*" + self.Elem.String()
}

func (self TypePtr) Equal(t Type) bool {
	if a, ok := t.(*TypePtr); ok {
		return self.Elem.Equal(a.Elem)
	}
	return false
}

// Typedef 类型定义
type Typedef struct {
	Pkg     stlos.Path
	Name    string
	Impls   *set.HashSet[*TypeInterface]
	Dst     Type
	Methods map[string]*Function
}

// NewTypedef 新建类型定义
func NewTypedef(pkg stlos.Path, name string, dst Type) *Typedef {
	return &Typedef{
		Pkg:     pkg,
		Name:    name,
		Impls:   set.NewHashSet[*TypeInterface](),
		Dst:     dst,
		Methods: make(map[string]*Function),
	}
}

// IsTypedef 是否是类型定义
func IsTypedef(t Type) bool {
	_, ok := t.(*Typedef)
	return ok
}

func (self Typedef) String() string {
	return self.Pkg.String() + "." + self.Name
}

func (self Typedef) Equal(t Type) bool {
	if td, ok := t.(*Typedef); ok && self.Pkg == td.Pkg && self.Name == td.Name {
		return true
	}
	return false
}

// IsImpl 是否实现了某个接口
func (self Typedef) IsImpl(dst *TypeInterface) bool {
	return self.Impls.Contain(dst)
}

// GetBaseType 获取底层类型
func GetBaseType(t Type) Type {
	switch typ := t.(type) {
	case *Typedef:
		return GetBaseType(typ.Dst)
	default:
		return typ
	}
}

// GetDepthBaseType 获取最底层类型
func GetDepthBaseType(t Type) Type {
	switch typ := t.(type) {
	case *typeBasic:
		return typ
	case *TypeFunc:
		params := make([]Type, len(typ.Params))
		for i, p := range typ.Params {
			params[i] = GetBaseType(p)
		}
		return NewFuncType(GetBaseType(typ.Ret), params, typ.VarArg)
	case *TypePtr:
		return NewPtrType(GetBaseType(typ.Elem))
	case *TypeArray:
		return NewArrayType(typ.Size, GetBaseType(typ.Elem))
	case *TypeTuple:
		elems := make([]Type, len(typ.Elems))
		for i, p := range typ.Elems {
			elems[i] = GetBaseType(p)
		}
		return NewTupleType(elems...)
	case *TypeStruct:
		fields := table.NewLinkedHashMap[string, types.Pair[bool, Type]]()
		for iter := typ.Fields.Begin(); iter.HasValue(); iter.Next() {
			fields.Set(iter.Key(), types.NewPair(iter.Value().First, GetBaseType(iter.Value().Second)))
		}
		return NewStructType(fields)
	case *Typedef:
		return GetBaseType(typ.Dst)
	default:
		panic(fmt.Sprintf("unknown type: %+v", t))
	}
}

// TypeInterface 接口类型
type TypeInterface struct {
	Fields *table.LinkedHashMap[string, *TypeFunc]
}

// NewTypeInterface 新建接口类型
func NewTypeInterface(fields *table.LinkedHashMap[string, *TypeFunc]) *TypeInterface {
	return &TypeInterface{Fields: fields}
}

// IsInterfaceType 是否是结构体类型
func IsInterfaceType(t Type) bool {
	_, ok := t.(*TypeInterface)
	return ok
}

// IsInterfaceTypeAndSon 是否是结构体类型及其子类型
func IsInterfaceTypeAndSon(t Type) bool {
	return IsInterfaceType(GetBaseType(t))
}

func (self TypeInterface) String() string {
	var buf strings.Builder
	buf.WriteString("interface(")
	for iter := self.Fields.Begin(); iter.HasValue(); iter.Next() {
		n, t := iter.Key(), iter.Value()
		buf.WriteString(fmt.Sprintf("%s: %s", n, t))
		if iter.HasNext() {
			buf.WriteString(", ")
		}
	}
	buf.WriteByte(')')
	return buf.String()
}

func (self TypeInterface) Equal(t Type) bool {
	if i, ok := t.(*TypeInterface); ok && self.Fields.Length() == i.Fields.Length() {
		for iter := self.Fields.Begin(); iter.HasValue(); iter.Next() {
			it := i.Fields.Get(iter.Key())
			if it == nil || !iter.Value().Equal(it) {
				return false
			}
		}
		return true
	}
	return false
}

// *********************************************************************************************************************

// 类型
func analyseType(ctx *packageContext, ast parse.Type) (Type, utils.Error) {
	if ast == nil {
		return None, nil
	}
	switch typ := ast.(type) {
	case *parse.TypeIdent:
		return analyseTypeIdent(ctx, typ, false)
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
	case *typeBasic:
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
func analyseTypeIdent(ctx *packageContext, ast *parse.TypeIdent, isImport bool) (Type, utils.Error) {
	if ast.Pkg == nil {
		switch ast.Name.Source {
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
		default:
			// 类型定义
			if td, ok := ctx.typedefs[ast.Name.Source]; ok && (!isImport || td.First) {
				return td.Second, nil
			}
			return nil, utils.Errorf(ast.Position(), "unknown identifier")
		}
	} else {
		pkg := ctx.externs[ast.Pkg.Source]
		if pkg == nil {
			return nil, utils.Errorf(ast.Pkg.Pos, "unknown `%s`", ast.Pkg.Source)
		}
		return analyseTypeIdent(pkg, parse.NewTypeIdent(nil, ast.Name), true)
	}
}
