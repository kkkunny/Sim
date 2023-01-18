package analyse

import (
	"github.com/kkkunny/Sim/src/compiler/lex"
	"github.com/kkkunny/Sim/src/compiler/parse"
	"github.com/kkkunny/Sim/src/compiler/utils"
)

// Expr 表达式
type Expr interface {
	Stmt
	GetType() Type
	GetMut() bool
	IsTemporary() bool
	IsConst() bool
}

// Ident 标识符
type Ident interface {
	Expr
	ident()
}

// Integer 整数
type Integer struct {
	Type  Type
	Value int64
}

func (self Integer) stmt() {}

func (self Integer) GetType() Type {
	return self.Type
}

func (self Integer) GetMut() bool {
	return false
}

func (self Integer) IsTemporary() bool {
	return true
}

func (self Integer) IsConst() bool {
	return true
}

// Float 浮点数
type Float struct {
	Type  Type
	Value float64
}

func (self Float) stmt() {}

func (self Float) GetType() Type {
	return self.Type
}

func (self Float) GetMut() bool {
	return false
}

func (self Float) IsTemporary() bool {
	return true
}

func (self Float) IsConst() bool {
	return true
}

// Boolean 布尔数
type Boolean struct {
	Type  Type
	Value bool
}

func (self Boolean) stmt() {}

func (self Boolean) GetType() Type {
	return self.Type
}

func (self Boolean) GetMut() bool {
	return false
}

func (self Boolean) IsTemporary() bool {
	return true
}

func (self Boolean) IsConst() bool {
	return true
}

// String 字符串
type String struct {
	Type  Type
	Value string
}

func (self String) stmt() {}

func (self String) GetType() Type {
	return self.Type
}

func (self String) GetMut() bool {
	return false
}

func (self String) IsTemporary() bool {
	return true
}

func (self String) IsConst() bool {
	return true
}

// Null 空指针
type Null struct {
	Type Type
}

func (self Null) stmt() {}

func (self Null) GetType() Type {
	return self.Type
}

func (self Null) GetMut() bool {
	return false
}

func (self Null) IsTemporary() bool {
	return true
}

func (self Null) IsConst() bool {
	return true
}

// Binary 二元表达式
type Binary struct {
	Opera       string
	Left, Right Expr
}

func (self Binary) stmt() {}

func (self Binary) GetType() Type {
	return self.Left.GetType()
}

func (self Binary) GetMut() bool {
	return false
}

func (self Binary) IsTemporary() bool {
	return true
}

func (self Binary) IsConst() bool {
	return false
}

// FuncCall 函数调用
type FuncCall struct {
	Func Expr
	Args []Expr
}

func (self FuncCall) stmt() {}

func (self FuncCall) GetType() Type {
	return GetBaseType(self.Func.GetType()).(*TypeFunc).Ret
}

func (self FuncCall) GetMut() bool {
	return false
}

func (self FuncCall) IsTemporary() bool {
	return true
}

func (self FuncCall) IsConst() bool {
	return false
}

// MethodCall 方法调用
type MethodCall struct {
	Method *Method
	Args   []Expr
}

func (self MethodCall) stmt() {}

func (self MethodCall) GetType() Type {
	return self.Method.GetMethodType().Ret
}

func (self MethodCall) GetMut() bool {
	return false
}

func (self MethodCall) IsTemporary() bool {
	return true
}

func (self MethodCall) IsConst() bool {
	return false
}

// Param 参数
type Param struct {
	Type Type
}

func (self Param) stmt() {}

func (self Param) ident() {}

func (self Param) GetType() Type {
	return self.Type
}

func (self Param) GetMut() bool {
	return true
}

func (self Param) IsTemporary() bool {
	return false
}

func (self Param) IsConst() bool {
	return false
}

// Array 数组
type Array struct {
	Type  Type
	Elems []Expr
}

func (self Array) stmt() {}

func (self Array) GetType() Type {
	return self.Type
}

func (self Array) GetMut() bool {
	return false
}

func (self Array) IsTemporary() bool {
	return true
}

func (self Array) IsConst() bool {
	for _, e := range self.Elems {
		if !e.IsConst() {
			return false
		}
	}
	return true
}

// EmptyArray 空数组
type EmptyArray struct {
	Type Type
}

func (self EmptyArray) stmt() {}

func (self EmptyArray) GetType() Type {
	return self.Type
}

func (self EmptyArray) GetMut() bool {
	return false
}

func (self EmptyArray) IsTemporary() bool {
	return true
}

func (self EmptyArray) IsConst() bool {
	return true
}

// Assign 赋值
type Assign struct {
	Opera       string
	Left, Right Expr
}

func (self Assign) stmt() {}

func (self Assign) GetType() Type {
	return None
}

func (self Assign) GetMut() bool {
	return false
}

func (self Assign) IsTemporary() bool {
	return true
}

func (self Assign) IsConst() bool {
	return false
}

// Equal 赋值
type Equal struct {
	Opera       string
	Left, Right Expr
}

func (self Equal) stmt() {}

func (self Equal) GetType() Type {
	return Bool
}

func (self Equal) GetMut() bool {
	return false
}

func (self Equal) IsTemporary() bool {
	return true
}

func (self Equal) IsConst() bool {
	return false
}

// Unary 一元表达式
type Unary struct {
	Mutable bool
	Type    Type
	Opera   string
	Value   Expr
}

func (self Unary) stmt() {}

func (self Unary) GetType() Type {
	return self.Type
}

func (self Unary) GetMut() bool {
	return self.Mutable
}

func (self Unary) IsTemporary() bool {
	return true
}

func (self Unary) IsConst() bool {
	return false
}

// Index 索引
type Index struct {
	Type        Type
	From, Index Expr
}

func (self Index) stmt() {}

func (self Index) GetType() Type {
	return self.Type
}

func (self Index) GetMut() bool {
	return self.From.GetMut()
}

func (self Index) IsTemporary() bool {
	return self.From.IsTemporary()
}

func (self Index) IsConst() bool {
	return false
}

// Select 选择
type Select struct {
	Cond, True, False Expr
}

func (self Select) stmt() {}

func (self Select) GetType() Type {
	return self.True.GetType()
}

func (self Select) GetMut() bool {
	return self.True.GetMut() && self.False.GetMut()
}

func (self Select) IsTemporary() bool {
	return self.True.IsTemporary() || self.False.IsTemporary()
}

func (self Select) IsConst() bool {
	return false
}

// Tuple 元组
type Tuple struct {
	Type  Type
	Elems []Expr
}

func (self Tuple) stmt() {}

func (self Tuple) GetType() Type {
	return self.Type
}

func (self Tuple) GetMut() bool {
	return false
}

func (self Tuple) IsTemporary() bool {
	return true
}

func (self Tuple) IsConst() bool {
	for _, e := range self.Elems {
		if !e.IsConst() {
			return false
		}
	}
	return true
}

// EmptyTuple 空元组
type EmptyTuple struct {
	Type Type
}

func (self EmptyTuple) stmt() {}

func (self EmptyTuple) GetType() Type {
	return self.Type
}

func (self EmptyTuple) GetMut() bool {
	return false
}

func (self EmptyTuple) IsTemporary() bool {
	return true
}

func (self EmptyTuple) IsConst() bool {
	return true
}

// Struct 结构体
type Struct struct {
	Type   Type
	Fields []Expr
}

func (self Struct) stmt() {}

func (self Struct) GetType() Type {
	return self.Type
}

func (self Struct) GetMut() bool {
	return false
}

func (self Struct) IsTemporary() bool {
	return true
}

func (self Struct) IsConst() bool {
	for _, e := range self.Fields {
		if !e.IsConst() {
			return false
		}
	}
	return true
}

// EmptyStruct 空结构体
type EmptyStruct struct {
	Type Type
}

func (self EmptyStruct) stmt() {}

func (self EmptyStruct) GetType() Type {
	return self.Type
}

func (self EmptyStruct) GetMut() bool {
	return false
}

func (self EmptyStruct) IsTemporary() bool {
	return true
}

func (self EmptyStruct) IsConst() bool {
	return true
}

// GetField 获取成员
type GetField struct {
	From  Expr
	Index string
}

func (self GetField) stmt() {}

func (self GetField) GetType() Type {
	return GetBaseType(self.From.GetType()).(*TypeStruct).Fields.Get(self.Index).Second
}

func (self GetField) GetMut() bool {
	return self.From.GetMut()
}

func (self GetField) IsTemporary() bool {
	return self.From.IsTemporary()
}

func (self GetField) IsConst() bool {
	return false
}

// Covert 类型转换
type Covert struct {
	From Expr
	To   Type
}

func (self Covert) stmt() {}

func (self Covert) GetType() Type {
	return self.To
}

func (self Covert) GetMut() bool {
	return false
}

func (self Covert) IsTemporary() bool {
	return true
}

func (self Covert) IsConst() bool {
	return false
}

// Method 方法
type Method struct {
	Self Expr // 类型定义 || 类型定义指针
	Func *Function
}

func (self Method) stmt() {}

func (self Method) GetType() Type {
	return self.Func.GetType()
}

func (self Method) GetMethodType() *TypeFunc {
	return self.Func.GetMethodType()
}

func (self Method) GetMut() bool {
	return false
}

func (self Method) IsTemporary() bool {
	return true
}

func (self Method) IsConst() bool {
	return false
}

// Alloc 栈内存分配
type Alloc struct {
	Size Expr
}

func (self Alloc) stmt() {}

func (self Alloc) GetType() Type {
	return NewPtrType(Usize)
}

func (self Alloc) GetMut() bool {
	return false
}

func (self Alloc) IsTemporary() bool {
	return true
}

func (self Alloc) IsConst() bool {
	return false
}

// GetInterfaceField 获取接口成员
type GetInterfaceField struct {
	From  Expr
	Index string
}

func (self GetInterfaceField) stmt() {}

func (self GetInterfaceField) GetType() Type {
	ft := self.GetType().(*TypeFunc)
	ft.Params = append([]Type{NewPtrType(Usize)}, ft.Params...)
	return ft
}

func (self GetInterfaceField) GetMethodType() *TypeFunc {
	return GetBaseType(self.From.GetType()).(*TypeInterface).Fields.Get(self.Index)
}

func (self GetInterfaceField) GetMut() bool {
	return false
}

func (self GetInterfaceField) IsTemporary() bool {
	return true
}

func (self GetInterfaceField) IsConst() bool {
	return false
}

// InterfaceFieldCall 接口成员调用
type InterfaceFieldCall struct {
	Field *GetInterfaceField
	Args  []Expr
}

func (self InterfaceFieldCall) stmt() {}

func (self InterfaceFieldCall) GetType() Type {
	return self.Field.GetMethodType().Ret
}

func (self InterfaceFieldCall) GetMut() bool {
	return false
}

func (self InterfaceFieldCall) IsTemporary() bool {
	return true
}

func (self InterfaceFieldCall) IsConst() bool {
	return false
}

// *********************************************************************************************************************

// 表达式
func analyseExpr(ctx *blockContext, expect Type, ast parse.Expr) (Expr, utils.Error) {
	switch expr := ast.(type) {
	case *parse.Int:
		if expect == nil || !IsNumberTypeAndSon(expect) {
			expect = Isize
		}
		if IsIntTypeAndSon(expect) {
			return &Integer{
				Type:  expect,
				Value: expr.Value,
			}, nil
		} else {
			return &Float{
				Type:  expect,
				Value: float64(expr.Value),
			}, nil
		}
	case *parse.Float:
		if expect == nil || !IsFloatTypeAndSon(expect) {
			expect = F64
		}
		return &Float{
			Type:  expect,
			Value: expr.Value,
		}, nil
	case *parse.Bool:
		if expect == nil || !IsBoolTypeAndSon(expect) {
			expect = Bool
		}
		return &Boolean{
			Type:  expect,
			Value: expr.Value,
		}, nil
	case *parse.Char:
		if expect == nil || !IsNumberTypeAndSon(expect) {
			expect = I32
		}
		return &Integer{
			Type:  expect,
			Value: int64(expr.Value),
		}, nil
	case *parse.String:
		if expect == nil || !GetDepthBaseType(expect).Equal(NewPtrType(I8)) {
			expect = NewPtrType(I8)
		}
		return &String{
			Type:  expect,
			Value: expr.Value,
		}, nil
	case *parse.Null:
		if expect == nil || (!IsPtrTypeAndSon(expect) && !IsFuncTypeAndSon(expect)) {
			return nil, utils.Errorf(expr.Position(), "expect a pointer type")
		}
		return &Null{Type: expect}, nil
	case *parse.Ident:
		return analyseIdent(ctx, expr)
	case *parse.Array:
		if len(expr.Elems) == 0 {
			if expect == nil || !IsArrayTypeAndSon(expect) {
				return nil, utils.Errorf(expr.Position(), "expect a array type")
			}
			return &EmptyArray{Type: expect}, nil
		}
		if expect != nil {
			if at, ok := GetBaseType(expect).(*TypeArray); ok && at.Size == uint(len(expr.Elems)) {
				expect = at.Elem
			}
		}

		elems := make([]Expr, len(expr.Elems))
		var errs []utils.Error
		for i, e := range expr.Elems {
			var err utils.Error
			if elems[0] == nil {
				elems[i], err = analyseExpr(ctx, expect, e)
			} else {
				elems[i], err = expectExpr(ctx, elems[0].GetType(), e)
			}
			if err != nil {
				errs = append(errs, err)
			}
		}
		if len(errs) == 1 {
			return nil, errs[0]
		} else if len(errs) > 1 {
			return nil, utils.NewMultiError(errs...)
		}

		var rt Type = NewArrayType(uint(len(elems)), elems[0].GetType())
		if expect != nil && GetDepthBaseType(expect).Equal(GetDepthBaseType(rt)) {
			rt = expect
		}
		return &Array{
			Type:  rt,
			Elems: elems,
		}, nil
	case *parse.TupleOrExpr:
		if len(expr.Elems) == 0 {
			if expect == nil || !IsTupleTypeAndSon(expect) {
				return nil, utils.Errorf(expr.Position(), "expect a tuple type")
			}
			return &EmptyTuple{Type: expect}, nil
		} else if len(expr.Elems) == 1 && (expect == nil || !IsTupleTypeAndSon(expect) || len(GetBaseType(expect).(*TypeTuple).Elems) != 1) {
			return analyseExpr(ctx, expect, expr.Elems[0])
		}
		expects := make([]Type, len(expr.Elems))
		if expect != nil {
			if tt, ok := GetBaseType(expect).(*TypeTuple); ok && len(tt.Elems) == len(expr.Elems) {
				for i := range expects {
					expects[i] = tt.Elems[i]
				}
			}
		}
		elems, err := analyseExprList(ctx, expects, expr.Elems)
		if err != nil {
			return nil, err
		}
		for i, e := range elems {
			expects[i] = e.GetType()
		}
		var rt Type = NewTupleType(expects...)
		if expect != nil && GetDepthBaseType(expect).Equal(GetDepthBaseType(rt)) {
			rt = expect
		}
		return &Tuple{
			Type:  rt,
			Elems: elems,
		}, nil
	case *parse.Struct:
		if len(expr.Fields) == 0 {
			if expect == nil || !IsStructTypeAndSon(expect) {
				return nil, utils.Errorf(expr.Position(), "expect a struct type")
			}
			return &EmptyStruct{Type: expect}, nil
		}
		if expect == nil || !IsStructTypeAndSon(expect) {
			return nil, utils.Errorf(expr.Position(), "expect a struct type")
		} else if GetBaseType(expect).(*TypeStruct).Fields.Length() != len(expr.Fields) {
			return nil, utils.Errorf(expr.Position(), "expect `%d` fields", len(expr.Fields))
		}
		expects := make([]Type, len(expr.Fields))
		for iter := GetBaseType(expect).(*TypeStruct).Fields.Begin(); iter.HasValue(); iter.Next() {
			expects[iter.Index()] = iter.Value().Second
		}
		fields, err := analyseExprList(ctx, expects, expr.Fields)
		if err != nil {
			return nil, err
		}
		for i, e := range fields {
			expects[i] = e.GetType()
		}
		return &Struct{
			Type:   expect,
			Fields: fields,
		}, nil
	case *parse.Unary:
		switch expr.Opera.Kind {
		case lex.SUB:
			value, err := analyseExpr(ctx, expect, expr.Value)
			if err != nil {
				return nil, err
			}
			if !IsNumberTypeAndSon(value.GetType()) {
				return nil, utils.Errorf(expr.Value.Position(), "expect a number")
			}
			left, err := getDefaultExprByType(expr.Value.Position(), value.GetType())
			if err != nil {
				return nil, err
			}
			return &Binary{
				Opera: "-",
				Left:  left,
				Right: value,
			}, nil
		case lex.NEG:
			value, err := analyseExpr(ctx, expect, expr.Value)
			if err != nil {
				return nil, err
			}
			if !IsSintTypeAndSon(value.GetType()) {
				return nil, utils.Errorf(expr.Value.Position(), "expect a signed integer")
			}
			return &Binary{
				Opera: "^",
				Left:  value,
				Right: &Integer{
					Type:  value.GetType(),
					Value: -1,
				},
			}, nil
		case lex.NOT:
			if expect == nil || !GetBaseType(expect).Equal(Bool) {
				expect = Bool
			}
			value, err := expectExprAndSon(ctx, expect, expr.Value)
			if err != nil {
				return nil, err
			}
			return &Unary{
				Mutable: false,
				Type:    value.GetType(),
				Opera:   "!",
				Value:   value,
			}, nil
		case lex.AND:
			if expect != nil && IsPtrTypeAndSon(expect) {
				expect = GetBaseType(expect).(*TypePtr).Elem
			}
			value, err := analyseExpr(ctx, expect, expr.Value)
			if err != nil {
				return nil, err
			}
			if value.IsTemporary() {
				return nil, utils.Errorf(expr.Value.Position(), "not expect a temporary value")
			}
			return &Unary{
				Mutable: false,
				Type:    NewPtrType(value.GetType()),
				Opera:   "&",
				Value:   value,
			}, nil
		case lex.MUL:
			if expect != nil {
				expect = NewPtrType(expect)
			}
			value, err := analyseExpr(ctx, expect, expr.Value)
			if err != nil {
				return nil, err
			}
			vt := value.GetType()
			if !IsPtrTypeAndSon(vt) {
				return nil, utils.Errorf(expr.Value.Position(), "expect a pointer")
			}
			return &Unary{
				Mutable: value.GetMut(),
				Type:    GetBaseType(vt).(*TypePtr).Elem,
				Opera:   "*",
				Value:   value,
			}, nil
		default:
			panic("")
		}
	case *parse.Binary:
		left, err := analyseExpr(ctx, nil, expr.Left)
		if err != nil {
			return nil, err
		}
		lt := left.GetType()
		right, err := expectExpr(ctx, lt, expr.Right)
		if err != nil {
			return nil, err
		}
		switch expr.Opera.Kind {
		case lex.ASS, lex.ADS, lex.SUS, lex.MUS, lex.DIS, lex.MOS, lex.ANS, lex.ORS, lex.XOS, lex.SLS, lex.SRS:
			if !left.GetMut() {
				return nil, utils.Errorf(expr.Left.Position(), "expect a mutable value")
			}
			switch expr.Opera.Kind {
			case lex.ASS:
			case lex.ADS, lex.SUS, lex.MUS, lex.DIS, lex.MOS:
				if !IsNumberTypeAndSon(lt) {
					return nil, utils.Errorf(expr.Left.Position(), "expect a number")
				}
			case lex.ANS, lex.ORS, lex.XOS, lex.SLS, lex.SRS:
				if !IsIntTypeAndSon(lt) {
					return nil, utils.Errorf(expr.Left.Position(), "expect a integer")
				}
			default:
				panic("unknown binary")
			}
			return &Assign{
				Opera: expr.Opera.Source,
				Left:  left,
				Right: right,
			}, nil
		case lex.LAN, lex.LOR:
			if !IsBoolTypeAndSon(lt) {
				return nil, utils.Errorf(expr.Left.Position(), "expect a boolean")
			}
		case lex.EQ, lex.NE:
			return &Equal{
				Opera: expr.Opera.Source,
				Left:  left,
				Right: right,
			}, nil
		case lex.LT, lex.LE, lex.GT, lex.GE:
			if !IsNumberTypeAndSon(lt) {
				return nil, utils.Errorf(expr.Left.Position(), "expect a number")
			}
			return &Equal{
				Opera: expr.Opera.Source,
				Left:  left,
				Right: right,
			}, nil
		case lex.ADD, lex.SUB, lex.MUL, lex.DIV, lex.MOD:
			if !IsNumberTypeAndSon(lt) {
				return nil, utils.Errorf(expr.Left.Position(), "expect a number")
			}
		case lex.AND, lex.OR, lex.XOR, lex.SHL, lex.SHR:
			if !IsIntTypeAndSon(lt) {
				return nil, utils.Errorf(expr.Left.Position(), "expect a integer")
			}
		default:
			panic("unknown binary")
		}
		return &Binary{
			Opera: expr.Opera.Source,
			Left:  left,
			Right: right,
		}, nil
	case *parse.Ternary:
		cond, err := expectExprAndSon(ctx, Bool, expr.Cond)
		if err != nil {
			return nil, err
		}
		tv, err := analyseExpr(ctx, expect, expr.True)
		if err != nil {
			return nil, err
		}
		fv, err := expectExpr(ctx, tv.GetType(), expr.False)
		if err != nil {
			return nil, err
		}
		return &Select{
			Cond:  cond,
			True:  tv,
			False: fv,
		}, nil
	case *parse.Call:
		f, err := analyseExpr(ctx, nil, expr.Func)
		if err != nil {
			if ident, ok := expr.Func.(*parse.Ident); ok && ident.Pkg == nil {
				return analyseBuildInFuncCall(ctx, ident, expr.Args)
			}
			return nil, err
		}
		if method, ok := f.(*Method); ok {
			ft := method.GetMethodType()
			if len(ft.Params) != len(expr.Args) {
				return nil, utils.Errorf(expr.Func.Position(), "expect %d arguments", len(ft.Params))
			}

			args := make([]Expr, len(expr.Args))
			var errs []utils.Error
			for i, pt := range ft.Params {
				var err utils.Error
				args[i], err = expectExpr(ctx, pt, expr.Args[i])
				if err != nil {
					errs = append(errs, err)
				}
			}
			if len(errs) == 1 {
				return nil, errs[0]
			} else if len(errs) > 1 {
				return nil, utils.NewMultiError(errs...)
			}

			return &MethodCall{
				Method: method,
				Args:   args,
			}, nil
		} else if im, ok := f.(*GetInterfaceField); ok {
			ft := im.GetMethodType()
			if len(ft.Params) != len(expr.Args) {
				return nil, utils.Errorf(expr.Func.Position(), "expect %d arguments", len(ft.Params))
			}

			args := make([]Expr, len(expr.Args))
			var errs []utils.Error
			for i, pt := range ft.Params {
				var err utils.Error
				args[i], err = expectExpr(ctx, pt, expr.Args[i])
				if err != nil {
					errs = append(errs, err)
				}
			}
			if len(errs) == 1 {
				return nil, errs[0]
			} else if len(errs) > 1 {
				return nil, utils.NewMultiError(errs...)
			}

			return &InterfaceFieldCall{
				Field: im,
				Args:  args,
			}, nil
		} else {
			ft, ok := GetBaseType(f.GetType()).(*TypeFunc)
			if !ok {
				return nil, utils.Errorf(expr.Func.Position(), "expect a function")
			} else if len(ft.Params) != len(expr.Args) {
				return nil, utils.Errorf(expr.Func.Position(), "expect %d arguments", len(ft.Params))
			}

			args := make([]Expr, len(expr.Args))
			var errs []utils.Error
			for i, pt := range ft.Params {
				var err utils.Error
				args[i], err = expectExpr(ctx, pt, expr.Args[i])
				if err != nil {
					errs = append(errs, err)
				}
			}
			if len(errs) == 1 {
				return nil, errs[0]
			} else if len(errs) > 1 {
				return nil, utils.NewMultiError(errs...)
			}

			return &FuncCall{
				Func: f,
				Args: args,
			}, nil
		}
	case *parse.Dot:
		prefix, err := analyseExpr(ctx, nil, expr.Front)
		if err != nil {
			return nil, err
		}

		// 方法
		prefixType := prefix.GetType()
		if IsTypedef(prefixType) || (IsPtrType(prefixType) && IsTypedef(prefixType.(*TypePtr).Elem)) {
			var _selfType *Typedef
			if td, ok := prefixType.(*Typedef); ok {
				_selfType = td
			} else {
				_selfType = prefixType.(*TypePtr).Elem.(*Typedef)
			}

			if fun, ok := _selfType.Methods[expr.End.Source]; ok {
				return &Method{
					Self: prefix,
					Func: fun,
				}, nil
			}
		}

		// 属性
		switch t := GetBaseType(prefixType).(type) {
		case *TypeStruct:
			if !t.Fields.ContainKey(expr.End.Source) {
				return nil, utils.Errorf(expr.End.Pos, "unknown identifier")
			} else if td, ok := prefixType.(*Typedef); ok && ctx.GetPackageContext().path != td.Pkg && !t.Fields.Get(expr.End.Source).First {
				return nil, utils.Errorf(expr.End.Pos, "unknown identifier")
			}
			return &GetField{
				From:  prefix,
				Index: expr.End.Source,
			}, nil
		case *TypeInterface:
			if !t.Fields.ContainKey(expr.End.Source) {
				return nil, utils.Errorf(expr.End.Pos, "unknown identifier")
			}
			return &GetInterfaceField{
				From:  prefix,
				Index: expr.End.Source,
			}, nil
		case *TypePtr:
			if st, ok := GetBaseType(t.Elem).(*TypeStruct); ok {
				if !st.Fields.ContainKey(expr.End.Source) {
					return nil, utils.Errorf(expr.End.Pos, "unknown identifier")
				} else if td, ok := t.Elem.(*Typedef); ok && ctx.GetPackageContext().path != td.Pkg && !st.Fields.Get(expr.End.Source).First {
					return nil, utils.Errorf(expr.End.Pos, "unknown identifier")
				}
				return &GetField{
					From: &Unary{
						Mutable: prefix.GetMut(),
						Type:    t.Elem,
						Opera:   "*",
						Value:   prefix,
					},
					Index: expr.End.Source,
				}, nil
			} else if st, ok := GetBaseType(t.Elem).(*TypeInterface); ok {
				if !st.Fields.ContainKey(expr.End.Source) {
					return nil, utils.Errorf(expr.End.Pos, "unknown identifier")
				}
				return &GetInterfaceField{
					From: &Unary{
						Mutable: prefix.GetMut(),
						Type:    t.Elem,
						Opera:   "*",
						Value:   prefix,
					},
					Index: expr.End.Source,
				}, nil
			}
		}
		return nil, utils.Errorf(expr.Front.Position(), "expect a struct")
	case *parse.Index:
		prefix, err := analyseExpr(ctx, nil, expr.Front)
		if err != nil {
			return nil, err
		}
		switch pt := GetBaseType(prefix.GetType()).(type) {
		case *TypeArray:
			index, err := autoExpectExpr(ctx, Usize, expr.Index)
			if err != nil {
				return nil, err
			}
			return &Index{
				Type:  pt.Elem,
				From:  prefix,
				Index: index,
			}, nil
		case *TypePtr:
			index, err := autoExpectExpr(ctx, Usize, expr.Index)
			if err != nil {
				return nil, err
			}
			return &Index{
				Type:  pt.Elem,
				From:  prefix,
				Index: index,
			}, nil
		case *TypeTuple:
			index, err := analyseExpr(ctx, Usize, expr.Index)
			if err != nil {
				return nil, err
			}
			literal, ok := index.(*Integer)
			if !ok {
				return nil, utils.Errorf(expr.Index.Position(), "expect a integer literal")
			}
			return &Index{
				Type:  pt.Elems[literal.Value],
				From:  prefix,
				Index: literal,
			}, nil
		default:
			return nil, utils.Errorf(expr.Front.Position(), "expect a array or tuple")
		}
	case *parse.Covert:
		to, err := analyseType(ctx.GetPackageContext(), expr.To)
		if err != nil {
			return nil, err
		}
		from, err := analyseExpr(ctx, to, expr.From)
		if err != nil {
			return nil, err
		}
		res := analyseCovert(from, to)
		if res == nil {
			return nil, utils.Errorf(expr.From.Position(), "can not covert to type `%s`", to)
		}
		return res, nil
	default:
		panic("unknown expression")
	}
}

// 期待指定类型的表达式
func expectExprWithType(pos utils.Position, expect Type, expr Expr) (Expr, utils.Error) {
	if vv := analyseAutoCovert(expr, expect); vv != nil {
		return vv, nil
	}
	exprType := expr.GetType()
	if !exprType.Equal(expect) {
		return nil, utils.Errorf(pos, "expect type `%s` but there is `%s`", expect, exprType)
	}
	return expr, nil
}

// 期待指定类型的表达式及其子类型
func expectExprWithTypeAndSon(pos utils.Position, expect Type, expr Expr) (Expr, utils.Error) {
	if vv := analyseAutoCovert(expr, expect); vv != nil {
		return vv, nil
	}
	exprType := expr.GetType()
	if !GetDepthBaseType(exprType).Equal(GetDepthBaseType(expect)) {
		return nil, utils.Errorf(pos, "expect type `%s` but there is `%s`", expect, exprType)
	}
	return expr, nil
}

// 期待指定类型的表达式
func expectExpr(ctx *blockContext, expect Type, ast parse.Expr) (Expr, utils.Error) {
	expr, err := analyseExpr(ctx, expect, ast)
	if err != nil {
		return nil, err
	}
	return expectExprWithType(ast.Position(), expect, expr)
}

// 期待指定类型的表达式及其子类型
func expectExprAndSon(ctx *blockContext, expect Type, ast parse.Expr) (Expr, utils.Error) {
	expr, err := analyseExpr(ctx, expect, ast)
	if err != nil {
		return nil, err
	}
	return expectExprWithTypeAndSon(ast.Position(), expect, expr)
}

// 自动转换成期待的类型
func autoExpectExpr(ctx *blockContext, expect Type, ast parse.Expr) (Expr, utils.Error) {
	expr, err := analyseExpr(ctx, expect, ast)
	if err != nil {
		return nil, err
	}
	v := analyseCovert(expr, expect)
	if v == nil {
		return nil, utils.Errorf(ast.Position(), "expect type `%s` but there is `%s`", expect, expr.GetType())
	}
	return v, nil
}

// 获取类型默认值
func getDefaultExprByType(pos utils.Position, t Type) (Expr, utils.Error) {
	switch GetBaseType(t).(type) {
	case *typeBasic:
		switch {
		case IsNoneType(t):
			panic("")
		case IsIntType(t):
			return &Integer{
				Type:  t,
				Value: 0,
			}, nil
		case IsFloatType(t):
			return &Float{
				Type:  t,
				Value: 0,
			}, nil
		case IsBoolType(t):
			return &Boolean{
				Type:  t,
				Value: false,
			}, nil
		default:
			panic("")
		}
	case *TypeFunc:
		return &Null{Type: t}, nil
	case *TypeArray:
		return &EmptyArray{Type: t}, nil
	case *TypeTuple:
		return &EmptyTuple{Type: t}, nil
	case *TypeStruct:
		return &EmptyStruct{Type: t}, nil
	case *TypePtr:
		return &Null{Type: t}, nil
	case *TypeInterface:
		return nil, utils.Errorf(pos, "can not get default value for this type")
	default:
		panic("")
	}
}

// 表达式列表
func analyseExprList(ctx *blockContext, expects []Type, asts []parse.Expr) ([]Expr, utils.Error) {
	exprs := make([]Expr, len(asts))
	var errors []utils.Error
	for i, e := range asts {
		var expect Type
		if len(expects) == len(asts) {
			expect = expects[i]
		}
		expr, err := analyseExpr(ctx, expect, e)
		if err != nil {
			errors = append(errors, err)
		} else {
			exprs[i] = expr
		}
	}

	if len(errors) == 0 {
		return exprs, nil
	} else if len(errors) == 1 {
		return nil, errors[0]
	} else {
		return nil, utils.NewMultiError(errors...)
	}
}

// 标识符
func analyseIdent(ctx *blockContext, ast *parse.Ident) (Expr, utils.Error) {
	if ast.Pkg == nil {
		v := ctx.GetValue(ast.Name.Source)
		if v == nil {
			return nil, utils.Errorf(ast.Position(), "unknown identifier")
		}
		return v, nil
	} else {
		pkg := ctx.GetPackageContext().externs[ast.Pkg.Source]
		if pkg == nil {
			return nil, utils.Errorf(ast.Pkg.Pos, "unknown identifier")
		}
		value := pkg.GetValue(ast.Name.Source)
		if !value.First || value.Second == nil {
			return nil, utils.Errorf(ast.Name.Pos, "unknown identifier")
		}
		return value.Second, nil
	}
}

// 内置函数调用
func analyseBuildInFuncCall(ctx *blockContext, ident *parse.Ident, paramAsts []parse.Expr) (Expr, utils.Error) {
	switch ident.Name.Source {
	case "len":
		if len(paramAsts) != 1 {
			return nil, utils.Errorf(ident.Position(), "expect 1 arguments")
		}
		param, err := analyseExpr(ctx, nil, paramAsts[0])
		if err != nil {
			return nil, err
		}
		pt := param.GetType()
		array, ok := GetBaseType(pt).(*TypeArray)
		if ok {
			return &Integer{
				Type:  Usize,
				Value: int64(array.Size),
			}, nil
		}
		return nil, utils.Errorf(paramAsts[0].Position(), "expect a array")
	case "typename":
		if len(paramAsts) != 1 {
			return nil, utils.Errorf(ident.Position(), "expect 1 arguments")
		}
		param, err := analyseExpr(ctx, nil, paramAsts[0])
		if err != nil {
			return nil, err
		}
		return &String{
			Type:  NewPtrType(I8),
			Value: param.GetType().String(),
		}, nil
	case "alloc":
		if len(paramAsts) != 1 {
			return nil, utils.Errorf(ident.Position(), "expect 1 arguments")
		}
		param, err := expectExpr(ctx, Usize, paramAsts[0])
		if err != nil {
			return nil, err
		}
		return &Alloc{Size: param}, nil
	default:
		return nil, utils.Errorf(ident.Position(), "unknown identifier")
	}
}

// 类型转换
func analyseCovert(v Expr, t Type) *Covert {
	if vv := analyseAutoCovert(v, t); vv != nil {
		return vv
	}

	ft := v.GetType()
	switch {
	case GetDepthBaseType(ft).Equal(GetDepthBaseType(t)):
	case IsNumberTypeAndSon(ft) && IsNumberTypeAndSon(t):
	case GetBaseType(ft).Equal(Usize) && (IsPtrTypeAndSon(t) || IsFuncTypeAndSon(t)):
	case (IsPtrTypeAndSon(ft) || IsFuncTypeAndSon(ft)) && GetBaseType(t).Equal(Usize):
	case (IsPtrTypeAndSon(ft) || IsFuncTypeAndSon(ft)) && (IsPtrTypeAndSon(t) || IsFuncTypeAndSon(t)):
	default:
		return nil
	}

	return &Covert{
		From: v,
		To:   t,
	}
}

// 自动类型转换
func analyseAutoCovert(v Expr, t Type) *Covert {
	ft := v.GetType()
	switch {
	case IsPtrTypeAndSon(ft) && IsTypedef(GetBaseType(ft).(*TypePtr).Elem) && IsInterfaceTypeAndSon(t) && GetBaseType(ft).(*TypePtr).Elem.(*Typedef).IsImpl(GetBaseType(t).(*TypeInterface)):
		// 类型定义指针 --> 接口类型定义
	default:
		return nil
	}

	return &Covert{
		From: v,
		To:   t,
	}
}
