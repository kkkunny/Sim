package analyse

import (
	"github.com/kkkunny/Sim/src/compiler/hir"
	"github.com/kkkunny/Sim/src/compiler/lex"
	"github.com/kkkunny/Sim/src/compiler/parse"
	"github.com/kkkunny/Sim/src/compiler/utils"
	stlutil "github.com/kkkunny/stl/util"
)

// 获取类型默认值
func (self *Analyser) getDefaultValue(t hir.Type) (hir.Expr, utils.Error) {
	switch {
	case t.IsBool():
		return hir.NewBoolean(t, false), nil
	case t.IsInt():
		return hir.NewInteger(t, 0), nil
	case t.IsFloat():
		return hir.NewFloat(t, 0), nil
	case t.IsPtr():
		return hir.NewEmptyPtr(t), nil
	case t.IsFunc():
		return hir.NewEmptyFunc(t), nil
	case t.IsArray():
		return hir.NewEmptyArray(t), nil
	case t.IsTuple():
		return hir.NewEmptyTuple(t), nil
	case t.IsStruct():
		return hir.NewEmptyStruct(t), nil
	case t.IsEnum():
		return hir.NewEmptyEnum(t), nil
	default:
		panic("unreachable")
	}
}

// 表达式
func (self *Analyser) analyseExpr(expect *hir.Type, astObj parse.Expr) (hir.Expr, utils.Error) {
	switch ast := astObj.(type) {
	case *parse.Ident:
		return self.analyseIdent(*ast)
	case *parse.Array:
		return self.analyseArray(expect, *ast)
	case *parse.TupleOrExpr:
		return self.analyseTuple(expect, *ast)
	case *parse.Struct:
		return self.analyseStruct(expect, *ast)
	case *parse.Bool:
		return self.analyseBool(expect, *ast)
	case *parse.Int:
		return self.analyseInt(expect, *ast)
	case *parse.Float:
		return self.analyseFloat(expect, *ast)
	case *parse.Char:
		return self.analyseChar(expect, *ast)
	case *parse.String:
		return self.analyseString(expect, *ast)
	case *parse.Null:
		return self.analyseNull(expect, *ast)
	case *parse.Unary:
		return self.analyseUnary(expect, *ast)
	case *parse.Binary:
		return self.analyseBinary(expect, *ast)
	case *parse.Ternary:
		return self.analyseTernary(expect, *ast)
	case *parse.Call:
		return self.analyseCall(*ast)
	case *parse.Covert:
		return self.analyseCovert(*ast)
	case *parse.Dot:
		return self.analyseDot(*ast)
	case *parse.Index:
		return self.analyseIndex(*ast)
	default:
		panic("unreachable")
	}
}

// 期待指定类型的表达式
func (self *Analyser) expectExpr(expect hir.Type, ast parse.Expr) (hir.Expr, utils.Error) {
	v, err := self.analyseExpr(&expect, ast)
	if err != nil {
		return nil, err
	}
	vt := v.Type()
	if !vt.Equal(expect) {
		return nil, utils.Errorf(ast.Position(), "expect type `%s` but there is `%s`", expect, vt)
	}
	return v, nil
}

// 期待近似指定类型的表达式
func (self *Analyser) expectLikeExpr(expect hir.Type, ast parse.Expr) (hir.Expr, utils.Error) {
	v, err := self.analyseExpr(&expect, ast)
	if err != nil {
		return nil, err
	}
	vt := v.Type()
	if !vt.Like(expect) {
		return nil, utils.Errorf(ast.Position(), "expect type `%s` but there is `%s`", expect, vt)
	}
	return v, nil
}

// 标识符
func (self *Analyser) analyseIdent(ast parse.Ident) (hir.Ident, utils.Error) {
	for _, astPkg := range ast.Pkgs {
		symbol := self.symbols[astPkg]
		if symbol == self.symbol.getPkgSymbolTable() {
			symbol = self.symbol
		}
		if v, ok := symbol.lookupValue(ast.Name.Source); ok {
			return v.data, nil
		}
	}
	return nil, utils.Errorf(ast.Name.Pos, errUnknownIdentifier)
}

// 数组
func (self *Analyser) analyseArray(expect *hir.Type, ast parse.Array) (hir.Expr, utils.Error) {
	var expectElemType hir.Type
	if expect != nil && expect.IsArray() {
		if expect.GetArraySize() == uint(len(ast.Elems)) {
			expectElemType = expect.GetArrayElem()
		} else if len(ast.Elems) == 0 {
			return hir.NewEmptyArray(*expect), nil
		}
	} else if len(ast.Elems) == 0 {
		return nil, utils.Errorf(ast.Position(), "expect a array type")
	}

	// 元素
	var _hasFirst bool // 是否至少有一个可以确定类型的元素
	elems := make([]hir.Expr, len(ast.Elems))
	var errs []utils.Error
	for i, e := range ast.Elems {
		var err utils.Error
		if !_hasFirst {
			elems[i], err = self.analyseExpr(&expectElemType, e)
		} else {
			elems[i], err = self.expectExpr(expectElemType, e)
		}
		if err != nil {
			errs = append(errs, err)
		} else {
			_hasFirst = true
			expectElemType = elems[i].Type()
		}
	}
	if len(errs) == 1 {
		return nil, errs[0]
	} else if len(errs) > 1 {
		return nil, utils.NewMultiError(errs...)
	}

	// 类型
	var expectType hir.Type
	if expect != nil && expect.IsArray() && expect.GetArraySize() == uint(len(ast.Elems)) && expect.GetArrayElem().Equal(expectElemType) {
		expectType = *expect
	} else {
		expectType = hir.NewTypeArray(uint(len(elems)), expectElemType)
	}

	return hir.NewArray(expectType, elems...), nil
}

// 元组
func (self *Analyser) analyseTuple(expect *hir.Type, ast parse.TupleOrExpr) (hir.Expr, utils.Error) {
	expectElemTypes := make([]hir.Type, len(ast.Elems))
	if expect != nil && expect.IsTuple() {
		if len(expect.GetTupleElems()) == len(ast.Elems) {
			expectElemTypes = expect.GetTupleElems()
		} else if len(ast.Elems) == 0 {
			return hir.NewEmptyTuple(*expect), nil
		}
	}
	if len(ast.Elems) == 1 {
		// 括号表达式
		return self.analyseExpr(expect, ast.Elems[0])
	}

	// 元素
	var isExpect bool // 是否是期待的类型
	elems := make([]hir.Expr, len(ast.Elems))
	var errs []utils.Error
	for i, e := range ast.Elems {
		var err utils.Error
		elems[i], err = self.analyseExpr(
			stlutil.Ternary[*hir.Type](
				expectElemTypes[i].IsNone(),
				nil,
				&expectElemTypes[i],
			), e,
		)
		if err != nil {
			errs = append(errs, err)
		} else {
			et := elems[i].Type()
			isExpect = isExpect && expectElemTypes[i].Equal(et)
			expectElemTypes[i] = et
		}
	}
	if len(errs) == 1 {
		return nil, errs[0]
	} else if len(errs) > 1 {
		return nil, utils.NewMultiError(errs...)
	}

	// 类型
	var expectType hir.Type
	if isExpect {
		expectType = *expect
	} else {
		expectType = hir.NewTypeTuple(expectElemTypes...)
	}

	return hir.NewTuple(expectType, elems...), nil
}

// 结构体
func (self *Analyser) analyseStruct(expect *hir.Type, ast parse.Struct) (*hir.Struct, utils.Error) {
	var st hir.Type
	if expect != nil && expect.IsStruct() {
		st = expect.GetDeepTypedefTarget()
	} else {
		return nil, utils.Errorf(ast.Position(), "expect a struct type")
	}
	fieldDefs := st.GetStructFields()

	// 用户给出的字段值
	giveFieldMap := make(map[string]hir.Expr)
	var errs []utils.Error
	for i, f := range ast.Fields {
		// 字段是否已经赋过值
		if _, ok := giveFieldMap[f.First.Source]; ok {
			errs = append(errs, utils.Errorf(f.First.Pos, errDuplicateDeclaration))
			continue
		}
		// 获取字段下标
		index := -1
		for i, def := range fieldDefs {
			if def.Second == f.First.Source {
				index = i
				break
			}
		}
		if index < 0 {
			errs = append(errs, utils.Errorf(f.First.Pos, errUnknownIdentifier))
			continue
		}
		// 类型检查
		field, err := self.expectExpr(fieldDefs[i].Third, f.Second)
		if err != nil {
			errs = append(errs, err)
			continue
		}
		giveFieldMap[f.First.Source] = field
	}
	if len(errs) == 1 {
		return nil, errs[0]
	} else if len(errs) > 1 {
		return nil, utils.NewMultiError(errs...)
	}

	// 填充字段默认值
	fields := make([]hir.Expr, len(fieldDefs))
	for i, def := range fieldDefs {
		if v, ok := giveFieldMap[def.Second]; ok {
			fields[i] = v
			continue
		}
		v, err := self.getDefaultValue(def.Third)
		if err != nil {
			errs = append(errs, err)
			continue
		}
		fields[i] = v
	}
	if len(errs) == 1 {
		return nil, errs[0]
	} else if len(errs) > 1 {
		return nil, utils.NewMultiError(errs...)
	}

	return hir.NewStruct(*expect, fields...), nil
}

// 枚举
func (self *Analyser) analyseEnum(t hir.Type, nameToken lex.Token, argAsts []parse.Expr) (*hir.Enum, utils.Error) {
	// 查找枚举
	fields := t.GetEnumFields()
	index := -1
	for i, f := range fields {
		if f.Second == nameToken.Source {
			index = i
			break
		}
	}
	if index < 0 || (t.IsTypedef() && !t.GetTypedef().Pkg.Equal(self.symbol.pkg) && !fields[index].First) {
		return nil, utils.Errorf(nameToken.Pos, errUnknownIdentifier)
	}
	field := fields[index]

	// 值数量
	valueCount := stlutil.Ternary(field.Third == nil, 0, 1)
	if len(argAsts) != valueCount {
		return nil, utils.Errorf(nameToken.Pos, "expect `%d` arguments but there is `%d`", valueCount, len(argAsts))
	}

	// 值
	var value hir.Expr
	if len(argAsts) == 1 {
		var err utils.Error
		value, err = self.expectExpr(*field.Third, argAsts[0])
		if err != nil {
			return nil, err
		}
	}

	return hir.NewEnum(t, field.Second, value), nil
}

// 布尔
func (self *Analyser) analyseBool(expect *hir.Type, ast parse.Bool) (*hir.Boolean, utils.Error) {
	var expectType hir.Type
	if expect != nil && expect.IsBool() {
		expectType = *expect
	} else {
		expectType = hir.NewTypeBool()
	}

	return hir.NewBoolean(expectType, ast.Value), nil
}

// 整数
func (self *Analyser) analyseInt(expect *hir.Type, ast parse.Int) (*hir.Integer, utils.Error) {
	var expectType hir.Type
	if expect != nil && expect.IsInt() {
		expectType = *expect
	} else {
		expectType = hir.NewTypeIsize()
	}

	return hir.NewInteger(expectType, ast.Value), nil
}

// 浮点数
func (self *Analyser) analyseFloat(expect *hir.Type, ast parse.Float) (*hir.Float, utils.Error) {
	var expectType hir.Type
	if expect != nil && expect.IsFloat() {
		expectType = *expect
	} else {
		expectType = hir.NewTypeF64()
	}

	return hir.NewFloat(expectType, ast.Value), nil
}

// 字符
func (self *Analyser) analyseChar(expect *hir.Type, ast parse.Char) (*hir.Integer, utils.Error) {
	var expectType hir.Type
	if expect != nil && expect.IsInt() {
		expectType = *expect
	} else {
		expectType = hir.NewTypeI8()
	}

	return hir.NewInteger(expectType, int64(ast.Value)), nil
}

// 字符串
func (self *Analyser) analyseString(expect *hir.Type, ast parse.String) (*hir.String, utils.Error) {
	expectType := hir.NewTypePtr(hir.NewTypeI8())
	if expect != nil && expect.Like(expectType) {
		expectType = *expect
	}

	return hir.NewString(expectType, ast.Value), nil
}

// null
func (self *Analyser) analyseNull(expect *hir.Type, ast parse.Null) (hir.Expr, utils.Error) {
	switch {
	case expect != nil && expect.IsPtr():
		return hir.NewEmptyPtr(*expect), nil
	case expect != nil && expect.IsFunc():
		return hir.NewEmptyFunc(*expect), nil
	default:
		return nil, utils.Errorf(ast.Position(), "expect a pointer type")
	}
}

// 一元表达式
func (self *Analyser) analyseUnary(expect *hir.Type, ast parse.Unary) (hir.Expr, utils.Error) {
	switch ast.Opera.Kind {
	case lex.SUB:
		v, err := self.analyseExpr(expect, ast.Value)
		if err != nil {
			return nil, err
		} else if v.Type().IsNumber() {
			return nil, utils.Errorf(ast.Value.Position(), errExpectNumber)
		}
		var left hir.Expr
		if v.Type().IsInt() {
			left = hir.NewInteger(v.Type(), 0)
		} else {
			left = hir.NewFloat(v.Type(), 0)
		}
		return hir.NewSub(left, v), nil
	case lex.NEG:
		v, err := self.analyseExpr(expect, ast.Value)
		if err != nil {
			return nil, err
		} else if v.Type().IsInt() {
			return nil, utils.Errorf(ast.Value.Position(), errExpectInteger)
		}
		return hir.NewXor(v, hir.NewInteger(v.Type(), -1)), nil
	case lex.NOT:
		v, err := self.analyseExpr(expect, ast.Value)
		if err != nil {
			return nil, err
		} else if !v.Type().IsBool() {
			return nil, utils.Errorf(ast.Value.Position(), errExpectBoolean)
		}
		return hir.NewNot(v), nil
	case lex.AND:
		if expect != nil && expect.IsPtr() {
			_elemType := expect.GetPtr()
			expect = &_elemType
		}
		v, err := self.analyseExpr(expect, ast.Value)
		if err != nil {
			return nil, err
		} else if v.Immediate() {
			return nil, utils.Errorf(ast.Value.Position(), errNotExpectImmediateValue)
		}
		return hir.NewGetPointer(v), nil
	case lex.MUL:
		if expect != nil {
			_elemType := hir.NewTypePtr(*expect)
			expect = &_elemType
		}
		v, err := self.analyseExpr(expect, ast.Value)
		if err != nil {
			return nil, err
		}
		return hir.NewGetValue(v), nil
	case lex.LEN:
		v, err := self.analyseExpr(nil, ast.Value)
		if err != nil {
			return nil, err
		}
		vt := v.Type()
		switch {
		case vt.IsArray():
			return hir.NewInteger(hir.NewTypeUsize(), int64(vt.GetArraySize())), nil
		case vt.IsTuple():
			return hir.NewInteger(hir.NewTypeUsize(), int64(len(vt.GetTupleElems()))), nil
		default:
			return nil, utils.Errorf(ast.Value.Position(), "expect a array or a tuple")
		}
	case lex.TYPENAME:
		v, err := self.analyseExpr(nil, ast.Value)
		if err != nil {
			return nil, err
		}
		return hir.NewString(hir.NewTypePtr(hir.NewTypeI8()), v.Type().String()), nil
	case lex.SIZEOF:
		v, err := self.analyseExpr(nil, ast.Value)
		if err != nil {
			return nil, err
		}
		return hir.NewInteger(hir.NewTypeUsize(), int64(v.Type().Size())), nil
	default:
		panic("unreachable")
	}
}

// 二元表达式
func (self *Analyser) analyseBinary(expect *hir.Type, ast parse.Binary) (hir.Binary, utils.Error) {
	var errs []utils.Error
	left, err := self.analyseExpr(expect, ast.Left)
	if err != nil {
		errs = append(errs, err)
	}
	var right hir.Expr
	if err == nil {
		right, err = self.expectExpr(left.Type(), ast.Right)
	} else {
		right, err = self.analyseExpr(expect, ast.Right)
	}
	if err != nil {
		errs = append(errs, err)
	}

	if len(errs) == 1 {
		return nil, errs[0]
	} else if len(errs) > 1 {
		return nil, utils.NewMultiError(errs...)
	}

	lt := left.Type()

	switch ast.Opera.Kind {
	case lex.ASS:
		if !left.Mutable() {
			return nil, utils.Errorf(ast.Left.Position(), errExpectMutableValue)
		}
		return hir.NewAssign(left, right), nil
	case lex.ADS:
		if !left.Mutable() {
			return nil, utils.Errorf(ast.Left.Position(), errExpectMutableValue)
		} else if !lt.IsNumber() {
			return nil, utils.Errorf(ast.Left.Position(), errExpectNumber)
		}
		return hir.NewAssign(left, hir.NewAdd(left, right)), nil
	case lex.SUS:
		if !left.Mutable() {
			return nil, utils.Errorf(ast.Left.Position(), errExpectMutableValue)
		} else if !lt.IsNumber() {
			return nil, utils.Errorf(ast.Left.Position(), errExpectNumber)
		}
		return hir.NewAssign(left, hir.NewSub(left, right)), nil
	case lex.MUS:
		if !left.Mutable() {
			return nil, utils.Errorf(ast.Left.Position(), errExpectMutableValue)
		} else if !lt.IsNumber() {
			return nil, utils.Errorf(ast.Left.Position(), errExpectNumber)
		}
		return hir.NewAssign(left, hir.NewMul(left, right)), nil
	case lex.DIS:
		if !left.Mutable() {
			return nil, utils.Errorf(ast.Left.Position(), errExpectMutableValue)
		} else if !lt.IsNumber() {
			return nil, utils.Errorf(ast.Left.Position(), errExpectNumber)
		}
		return hir.NewAssign(left, hir.NewDiv(left, right)), nil
	case lex.MOS:
		if !left.Mutable() {
			return nil, utils.Errorf(ast.Left.Position(), errExpectMutableValue)
		} else if !lt.IsNumber() {
			return nil, utils.Errorf(ast.Left.Position(), errExpectNumber)
		}
		return hir.NewAssign(left, hir.NewMod(left, right)), nil
	case lex.ANS:
		if !left.Mutable() {
			return nil, utils.Errorf(ast.Left.Position(), errExpectMutableValue)
		} else if !lt.IsInt() {
			return nil, utils.Errorf(ast.Left.Position(), errExpectInteger)
		}
		return hir.NewAssign(left, hir.NewAnd(left, right)), nil
	case lex.ORS:
		if !left.Mutable() {
			return nil, utils.Errorf(ast.Left.Position(), errExpectMutableValue)
		} else if !lt.IsInt() {
			return nil, utils.Errorf(ast.Left.Position(), errExpectInteger)
		}
		return hir.NewAssign(left, hir.NewOr(left, right)), nil
	case lex.XOS:
		if !left.Mutable() {
			return nil, utils.Errorf(ast.Left.Position(), errExpectMutableValue)
		} else if !lt.IsInt() {
			return nil, utils.Errorf(ast.Left.Position(), errExpectInteger)
		}
		return hir.NewAssign(left, hir.NewXor(left, right)), nil
	case lex.SLS:
		if !left.Mutable() {
			return nil, utils.Errorf(ast.Left.Position(), errExpectMutableValue)
		} else if !lt.IsInt() {
			return nil, utils.Errorf(ast.Left.Position(), errExpectInteger)
		}
		return hir.NewAssign(left, hir.NewShl(left, right)), nil
	case lex.SRS:
		if !left.Mutable() {
			return nil, utils.Errorf(ast.Left.Position(), errExpectMutableValue)
		} else if !lt.IsInt() {
			return nil, utils.Errorf(ast.Left.Position(), errExpectInteger)
		}
		return hir.NewAssign(left, hir.NewShr(left, right)), nil
	case lex.LAN:
		if !lt.IsBool() {
			return nil, utils.Errorf(ast.Left.Position(), errExpectBoolean)
		}
		return hir.NewLogicAnd(left, right), nil
	case lex.LOR:
		if !lt.IsBool() {
			return nil, utils.Errorf(ast.Left.Position(), errExpectBoolean)
		}
		return hir.NewLogicOr(left, right), nil
	case lex.AND:
		if !lt.IsInt() {
			return nil, utils.Errorf(ast.Left.Position(), errExpectInteger)
		}
		return hir.NewAnd(left, right), nil
	case lex.OR:
		if !lt.IsInt() {
			return nil, utils.Errorf(ast.Left.Position(), errExpectInteger)
		}
		return hir.NewOr(left, right), nil
	case lex.XOR:
		if !lt.IsInt() {
			return nil, utils.Errorf(ast.Left.Position(), errExpectInteger)
		}
		return hir.NewXor(left, right), nil
	case lex.SHL:
		if !lt.IsInt() {
			return nil, utils.Errorf(ast.Left.Position(), errExpectInteger)
		}
		return hir.NewShl(left, right), nil
	case lex.SHR:
		if !lt.IsInt() {
			return nil, utils.Errorf(ast.Left.Position(), errExpectInteger)
		}
		return hir.NewShr(left, right), nil
	case lex.EQ:
		typ := hir.NewTypeBool()
		if expect != nil && expect.IsBool() {
			typ = *expect
		}
		return hir.NewEqual(typ, left, right), nil
	case lex.NE:
		typ := hir.NewTypeBool()
		if expect != nil && expect.IsBool() {
			typ = *expect
		}
		return hir.NewNotEqual(typ, left, right), nil
	case lex.LT:
		if !lt.IsNumber() {
			return nil, utils.Errorf(ast.Left.Position(), errExpectNumber)
		}
		typ := hir.NewTypeBool()
		if expect != nil && expect.IsBool() {
			typ = *expect
		}
		return hir.NewLt(typ, left, right), nil
	case lex.LE:
		if !lt.IsNumber() {
			return nil, utils.Errorf(ast.Left.Position(), errExpectNumber)
		}
		typ := hir.NewTypeBool()
		if expect != nil && expect.IsBool() {
			typ = *expect
		}
		return hir.NewLe(typ, left, right), nil
	case lex.GT:
		if !lt.IsNumber() {
			return nil, utils.Errorf(ast.Left.Position(), errExpectNumber)
		}
		typ := hir.NewTypeBool()
		if expect != nil && expect.IsBool() {
			typ = *expect
		}
		return hir.NewLt(typ, right, left), nil
	case lex.GE:
		if !lt.IsNumber() {
			return nil, utils.Errorf(ast.Left.Position(), errExpectNumber)
		}
		typ := hir.NewTypeBool()
		if expect != nil && expect.IsBool() {
			typ = *expect
		}
		return hir.NewLe(typ, right, left), nil
	case lex.ADD:
		if !lt.IsNumber() {
			return nil, utils.Errorf(ast.Left.Position(), errExpectNumber)
		}
		return hir.NewAdd(left, right), nil
	case lex.SUB:
		if !lt.IsNumber() {
			return nil, utils.Errorf(ast.Left.Position(), errExpectNumber)
		}
		return hir.NewSub(left, right), nil
	case lex.MUL:
		if !lt.IsNumber() {
			return nil, utils.Errorf(ast.Left.Position(), errExpectNumber)
		}
		return hir.NewMul(left, right), nil
	case lex.DIV:
		if !lt.IsNumber() {
			return nil, utils.Errorf(ast.Left.Position(), errExpectNumber)
		}
		return hir.NewDiv(left, right), nil
	case lex.MOD:
		if !lt.IsNumber() {
			return nil, utils.Errorf(ast.Left.Position(), errExpectNumber)
		}
		return hir.NewMod(left, right), nil
	default:
		panic("unreachable")
	}
}

// 三元表达式
func (self *Analyser) analyseTernary(expect *hir.Type, ast parse.Ternary) (*hir.Ternary, utils.Error) {
	var errs []utils.Error

	cond, err := self.expectLikeExpr(hir.NewTypeBool(), ast.Cond)
	if err != nil {
		errs = append(errs, err)
	}

	tv, err := self.analyseExpr(expect, ast.True)
	if err != nil {
		errs = append(errs, err)
	}

	var fv hir.Expr
	if err == nil {
		fv, err = self.expectExpr(tv.Type(), ast.False)
	} else {
		fv, err = self.analyseExpr(expect, ast.False)
	}
	if err != nil {
		errs = append(errs, err)
	}

	if len(errs) == 1 {
		return nil, errs[0]
	} else if len(errs) > 1 {
		return nil, utils.NewMultiError(errs...)
	}

	return hir.NewTernary(cond, tv, fv), nil
}

// 调用
func (self *Analyser) analyseCall(ast parse.Call) (hir.Expr, utils.Error) {
	if dot, ok := ast.Func.(*parse.Dot); ok {
		// 枚举值
		if ident, ok := dot.Front.(*parse.Ident); ok {
			t, err := self.analyseTypeIdent(*parse.NewTypeIdent(ident.Pkgs, ident.Name))
			if err == nil && t.IsEnum() {
				return self.analyseEnum(t, dot.End, ast.Args)
			}
		}
		// 方法调用
		selfExpr, err := self.analyseExpr(nil, dot.Front)
		if err != nil {
			return nil, err
		}
		if selfExpr.Type().IsTypedef() || (selfExpr.Type().Kind == hir.TPtr && selfExpr.Type().GetPtr().IsTypedef()) {
			return self.analyseCallMethod(selfExpr, dot.End, ast.Args)
		}
	}

	// 函数
	f, err := self.analyseExpr(nil, ast.Func)
	if err != nil {
		return nil, err
	}
	ft := f.Type()

	// 参数数量
	paramTypes := ft.GetFuncParams()
	if len(ast.Args) < len(paramTypes) || (!ft.GetFuncVarArg() && len(ast.Args) > len(paramTypes)) {
		return nil, utils.Errorf(
			ast.Func.Position(),
			"expect `%d` arguments but there is `%d`",
			len(paramTypes),
			len(ast.Args),
		)
	}

	// 实参
	var errs []utils.Error
	args := make([]hir.Expr, len(ast.Args))
	for i, a := range ast.Args {
		if i < len(paramTypes) {
			args[i], err = self.expectExpr(paramTypes[i], a)
		} else {
			args[i], err = self.analyseExpr(nil, a)
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

	return hir.NewFuncCall(f, args...), nil
}

// 调用方法
func (self *Analyser) analyseCallMethod(
	selfExpr hir.Expr, methodNameToken lex.Token, argAsts []parse.Expr,
) (*hir.MethodCall, utils.Error) {
	// 方法
	var def *hir.Typedef
	if selfExpr.Type().IsPtr() {
		def = selfExpr.Type().GetPtr().GetTypedef()
	} else {
		def = selfExpr.Type().GetTypedef()
	}
	m, ok := def.LookupMethod(methodNameToken.Source)
	if !ok {
		return nil, utils.Errorf(methodNameToken.Pos, errUnknownIdentifier)
	}
	ft := m.Type()

	// 参数数量
	paramTypes := ft.GetFuncParams()
	if len(argAsts) < len(paramTypes) || (!ft.GetFuncVarArg() && len(argAsts) > len(paramTypes)) {
		return nil, utils.Errorf(
			methodNameToken.Pos,
			"expect `%d` arguments but there is `%d`",
			len(paramTypes),
			len(argAsts),
		)
	}

	// 实参
	var errs []utils.Error
	args := make([]hir.Expr, len(argAsts))
	for i, a := range argAsts {
		var err utils.Error
		if i < len(paramTypes) {
			args[i], err = self.expectExpr(paramTypes[i], a)
		} else {
			args[i], err = self.analyseExpr(nil, a)
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

	return hir.NewMethodCall(m, selfExpr, args...), nil
}

// 点
func (self *Analyser) analyseDot(ast parse.Dot) (hir.Expr, utils.Error) {
	// 枚举值
	if ident, ok := ast.Front.(*parse.Ident); ok {
		t, err := self.analyseTypeIdent(*parse.NewTypeIdent(ident.Pkgs, ident.Name))
		if err == nil && t.IsEnum() {
			return self.analyseEnum(t, ast.End, nil)
		}
	}

	// from值
	from, err := self.analyseExpr(nil, ast.Front)
	if err != nil {
		return nil, err
	}
	ft := from.Type()

	switch {
	case ft.IsStruct():
		fields := ft.GetStructFields()
		for _, f := range fields {
			if f.Second == ast.End.Source {
				if (ft.IsTypedef() && (ft.GetTypedef().Pkg.Equal(self.symbol.pkg) || f.First)) || !ft.IsTypedef() {
					return hir.NewGetStructField(from, ast.End.Source), nil
				} else {
					break
				}
			}
		}
		return nil, utils.Errorf(ast.End.Pos, errUnknownIdentifier)
	case ft.IsPtr() && ft.GetPtr().IsStruct():
		st := ft.GetPtr()
		fields := st.GetStructFields()
		for _, f := range fields {
			if f.Second == ast.End.Source {
				if (st.IsTypedef() && (st.GetTypedef().Pkg.Equal(self.symbol.pkg) || f.First)) || !st.IsTypedef() {
					return hir.NewGetStructField(hir.NewGetValue(from), ast.End.Source), nil
				} else {
					break
				}
			}
		}
		return nil, utils.Errorf(ast.End.Pos, errUnknownIdentifier)
	case ft.IsEnum():
		fields := ft.GetEnumFields()
		index := -1
		for i, f := range fields {
			if f.Second == ast.End.Source {
				index = i
				break
			}
		}
		if index < 0 || (ft.IsTypedef() && !ft.GetTypedef().Pkg.Equal(self.symbol.pkg) && !fields[index].First) {
			return nil, utils.Errorf(ast.End.Pos, errUnknownIdentifier)
		}
		return hir.NewGetEnumField(from, ast.End.Source), nil
	case ft.IsPtr() && ft.GetPtr().IsEnum():
		et := ft.GetPtr()
		fields := et.GetEnumFields()
		index := -1
		for i, f := range fields {
			if f.Second == ast.End.Source {
				index = i
				break
			}
		}
		if index < 0 || (et.IsTypedef() && !et.GetTypedef().Pkg.Equal(self.symbol.pkg) && !fields[index].First) {
			return nil, utils.Errorf(ast.End.Pos, errUnknownIdentifier)
		}
		return hir.NewGetEnumField(hir.NewGetValue(from), ast.End.Source), nil
	default:
		return nil, utils.Errorf(ast.Front.Position(), "expect a struct or a pointer of struct")
	}
}

// 索引
func (self *Analyser) analyseIndex(ast parse.Index) (hir.Index, utils.Error) {
	from, err := self.analyseExpr(nil, ast.Front)
	if err != nil {
		return nil, err
	}
	ft := from.Type()

	switch {
	case ft.IsArray():
		index, err := self.expectExpr(hir.NewTypeUsize(), ast.Index)
		if err != nil {
			return nil, err
		}
		return hir.NewArrayIndex(from, index), nil
	case ft.IsTuple():
		literal, ok := ast.Index.(*parse.Int)
		if !ok {
			return nil, utils.Errorf(ast.Index.Position(), "expect a literal integer")
		} else if literal.Value < 0 || literal.Value > int64(len(ft.GetTupleElems())) {
			return nil, utils.Errorf(
				ast.Index.Position(), "expect a integer between `%d` and `%d`", 0,
				len(ft.GetTupleElems()),
			)
		}
		return hir.NewTupleIndex(from, uint(literal.Value)), nil
	case ft.IsPtr():
		index, err := self.expectExpr(hir.NewTypeUsize(), ast.Index)
		if err != nil {
			return nil, err
		}
		return hir.NewPointerIndex(from, index), nil
	default:
		return nil, utils.Errorf(ast.Front.Position(), "expect a array or a tuple or a pointer")
	}
}

// 类型转换
func (self *Analyser) analyseCovert(ast parse.Covert) (hir.Expr, utils.Error) {
	to, err := self.analyseType(ast.To)
	if err != nil {
		return nil, err
	}

	from, err := self.analyseExpr(&to, ast.From)
	if err != nil {
		return nil, err
	}
	ft := from.Type()

	switch {
	case ft.Equal(to):
		return from, nil
	case ft.Like(to):
		return hir.NewWrapCovert(from, to), nil
	case ft.IsInt() && to.IsInt():
		return hir.NewInt2Int(from, to), nil
	case ft.IsFloat() && to.IsFloat():
		return hir.NewFloat2Float(from, to), nil
	case ft.IsInt() && to.IsFloat():
		return hir.NewInt2Float(from, to), nil
	case ft.IsFloat() && to.IsInt():
		return hir.NewFloat2Int(from, to), nil
	case (ft.IsPtr() || ft.IsFunc()) && (to.IsPtr() || to.IsFunc()):
		return hir.NewPtr2Ptr(from, to), nil
	case ft.IsUsize() && (to.IsPtr() || to.IsFunc()):
		return hir.NewUsize2Ptr(from, to), nil
	case (ft.IsPtr() || ft.IsFunc()) && to.IsUsize():
		return hir.NewPtr2Usize(from, to), nil
	default:
		return nil, utils.Errorf(ast.To.Position(), "can not covert to this type")
	}
}
