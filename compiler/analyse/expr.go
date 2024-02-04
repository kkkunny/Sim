package analyse

import (
	"math/big"

	stlbasic "github.com/kkkunny/stl/basic"
	"github.com/kkkunny/stl/container/either"
	stlerror "github.com/kkkunny/stl/error"
	"github.com/samber/lo"

	"github.com/kkkunny/Sim/ast"
	errors "github.com/kkkunny/Sim/error"
	"github.com/kkkunny/Sim/hir"
	"github.com/kkkunny/Sim/token"
	"github.com/kkkunny/Sim/util"
)

func (self *Analyser) analyseExpr(expect hir.Type, node ast.Expr) hir.Expr {
	switch exprNode := node.(type) {
	case *ast.Integer:
		return self.analyseInteger(expect, exprNode)
	case *ast.Char:
		return self.analyseChar(expect, exprNode)
	case *ast.Float:
		return self.analyseFloat(expect, exprNode)
	case *ast.Binary:
		return self.analyseBinary(expect, exprNode)
	case *ast.Unary:
		return self.analyseUnary(expect, exprNode)
	case *ast.IdentExpr:
		return self.analyseIdentExpr(exprNode)
	case *ast.Call:
		return self.analyseCall(exprNode)
	case *ast.Tuple:
		return self.analyseTuple(expect, exprNode)
	case *ast.Covert:
		return self.analyseCovert(exprNode)
	case *ast.Array:
		return self.analyseArray(expect, exprNode)
	case *ast.Index:
		return self.analyseIndex(exprNode)
	case *ast.Extract:
		return self.analyseExtract(expect, exprNode)
	case *ast.Struct:
		return self.analyseStruct(exprNode)
	case *ast.GetField:
		return self.analyseGetField(exprNode)
	case *ast.String:
		return self.analyseString(exprNode)
	case *ast.Judgment:
		return self.analyseJudgment(exprNode)
	case *ast.Null:
		return self.analyseNull(expect, exprNode)
	case *ast.StaticMethod:
		return self.analyseStaticMethod(exprNode)
	default:
		panic("unreachable")
	}
}

func (self *Analyser) analyseInteger(expect hir.Type, node *ast.Integer) hir.Expr {
	if expect == nil || !hir.IsNumberType(expect) {
		expect = hir.Isize
	}
	switch {
	case hir.IsIntType(expect):
		value, ok := big.NewInt(0).SetString(node.Value.Source(), 10)
		if !ok {
			panic("unreachable")
		}
		return &hir.Integer{
			Type:  expect,
			Value: value,
		}
	case hir.IsFloatType(expect):
		value, _ := stlerror.MustWith2(big.ParseFloat(node.Value.Source(), 10, big.MaxPrec, big.ToZero))
		return &hir.Float{
			Type:  expect,
			Value: value,
		}
	default:
		panic("unreachable")
	}
}

func (self *Analyser) analyseChar(expect hir.Type, node *ast.Char) hir.Expr {
	if expect == nil || !hir.IsNumberType(expect) {
		expect = hir.I32
	}
	s := node.Value.Source()
	char := util.ParseEscapeCharacter(s[1:len(s)-1], `\'`, `'`)[0]
	switch {
	case hir.IsIntType(expect):
		value := big.NewInt(int64(char))
		return &hir.Integer{
			Type:  expect,
			Value: value,
		}
	case hir.IsFloatType(expect):
		value := big.NewFloat(float64(char))
		return &hir.Float{
			Type:  expect,
			Value: value,
		}
	default:
		panic("unreachable")
	}
}

func (self *Analyser) analyseFloat(expect hir.Type, node *ast.Float) *hir.Float {
	if expect == nil || !hir.IsFloatType(expect) {
		expect = hir.F64
	}
	value, _ := stlerror.MustWith2(big.NewFloat(0).Parse(node.Value.Source(), 10))
	return &hir.Float{
		Type:  expect,
		Value: value,
	}
}

func (self *Analyser) analyseBinary(expect hir.Type, node *ast.Binary) hir.Binary {
	left := self.analyseExpr(expect, node.Left)
	lt := left.GetType()
	right := self.analyseExpr(lt, node.Right)
	rt := right.GetType()

	switch node.Opera.Kind {
	case token.ASS:
		if lt.EqualTo(rt) {
			if !left.Mutable() {
				errors.ThrowNotMutableError(node.Left.Position())
			}
			return &hir.Assign{
				Left:  left,
				Right: right,
			}
		}
	case token.AND:
		if lt.EqualTo(rt) && hir.IsIntType(lt) {
			return &hir.IntAndInt{
				Left:  left,
				Right: right,
			}
		}
	case token.OR:
		if lt.EqualTo(rt) && hir.IsIntType(lt) {
			return &hir.IntOrInt{
				Left:  left,
				Right: right,
			}
		}
	case token.XOR:
		if lt.EqualTo(rt) && hir.IsIntType(lt) {
			return &hir.IntXorInt{
				Left:  left,
				Right: right,
			}
		}
	case token.SHL:
		if lt.EqualTo(rt) && hir.IsIntType(lt) {
			return &hir.IntShlInt{
				Left:  left,
				Right: right,
			}
		}
	case token.SHR:
		if lt.EqualTo(rt) && hir.IsIntType(lt) {
			return &hir.IntShrInt{
				Left:  left,
				Right: right,
			}
		}
	case token.ADD:
		if lt.EqualTo(rt) && hir.IsNumberType(lt) {
			return &hir.NumAddNum{
				Left:  left,
				Right: right,
			}
		}
	case token.SUB:
		if lt.EqualTo(rt) && hir.IsNumberType(lt) {
			return &hir.NumSubNum{
				Left:  left,
				Right: right,
			}
		}
	case token.MUL:
		if lt.EqualTo(rt) && hir.IsNumberType(lt) {
			return &hir.NumMulNum{
				Left:  left,
				Right: right,
			}
		}
	case token.DIV:
		if lt.EqualTo(rt) && hir.IsNumberType(lt) {
			if (stlbasic.Is[*hir.Integer](right) && right.(*hir.Integer).Value.Cmp(big.NewInt(0))==0) ||
				(stlbasic.Is[*hir.Float](right) && right.(*hir.Float).Value.Cmp(big.NewFloat(0))==0){
				errors.ThrowDivZero(node.Right.Position())
			}
			return &hir.NumDivNum{
				Left:  left,
				Right: right,
			}
		}
	case token.REM:
		if lt.EqualTo(rt) && hir.IsNumberType(lt) {
			if (stlbasic.Is[*hir.Integer](right) && right.(*hir.Integer).Value.Cmp(big.NewInt(0))==0) ||
				(stlbasic.Is[*hir.Float](right) && right.(*hir.Float).Value.Cmp(big.NewFloat(0))==0){
				errors.ThrowDivZero(node.Right.Position())
			}
			return &hir.NumRemNum{
				Left:  left,
				Right: right,
			}
		}
	case token.EQ:
		if lt.EqualTo(rt) {
			if !hir.IsEmptyType(lt){
				return &hir.Equal{
					Left:  left,
					Right: right,
				}
			}
		}
	case token.NE:
		if !hir.IsEmptyType(lt){
			return &hir.NotEqual{
				Left:  left,
				Right: right,
			}
		}
	case token.LT:
		if lt.EqualTo(rt) && hir.IsNumberType(lt) {
			return &hir.NumLtNum{
				Left:  left,
				Right: right,
			}
		}
	case token.GT:
		if lt.EqualTo(rt) && hir.IsNumberType(lt) {
			return &hir.NumGtNum{
				Left:  left,
				Right: right,
			}
		}
	case token.LE:
		if lt.EqualTo(rt) && hir.IsNumberType(lt) {
			return &hir.NumLeNum{
				Left:  left,
				Right: right,
			}
		}
	case token.GE:
		if lt.EqualTo(rt) && hir.IsNumberType(lt) {
			return &hir.NumGeNum{
				Left:  left,
				Right: right,
			}
		}
	case token.LAND:
		if lt.EqualTo(rt) && hir.IsBoolType(lt) {
			return &hir.BoolAndBool{
				Left:  left,
				Right: right,
			}
		}
	case token.LOR:
		if lt.EqualTo(rt) && hir.IsBoolType(lt) {
			return &hir.BoolOrBool{
				Left:  left,
				Right: right,
			}
		}
	default:
		panic("unreachable")
	}

	errors.ThrowIllegalBinaryError(node.Position(), node.Opera, left, right)
	return nil
}

func (self *Analyser) analyseUnary(expect hir.Type, node *ast.Unary) hir.Unary {
	switch node.Opera.Kind {
	case token.SUB:
		value := self.analyseExpr(expect, node.Value)
		vt := value.GetType()
		if hir.IsSintType(vt) || hir.IsFloatType(vt) {
			return &hir.NumNegate{Value: value}
		}
		errors.ThrowIllegalUnaryError(node.Position(), node.Opera, vt)
		return nil
	case token.NOT:
		value := self.analyseExpr(expect, node.Value)
		vt := value.GetType()
		switch {
		case hir.IsIntType(vt):
			return &hir.IntBitNegate{Value: value}
		case hir.IsBoolType(vt):
			return &hir.BoolNegate{Value: value}
		default:
			errors.ThrowIllegalUnaryError(node.Position(), node.Opera, vt)
			return nil
		}
	case token.AND:
		if expect != nil && hir.IsRefType(expect) {
			expect = hir.AsRefType(expect).Elem
		}
		value := self.analyseExpr(expect, node.Value)
		if !stlbasic.Is[hir.Ident](value) {
			errors.ThrowCanNotGetPointer(node.Value.Position())
		}else if localVarDef, ok := value.(*hir.LocalVarDef); ok{
			localVarDef.Escaped = true
		}
		return &hir.GetRef{Value: value}
	case token.MUL:
		if expect != nil {
			expect = &hir.RefType{Elem: expect}
		}
		value := self.analyseExpr(expect, node.Value)
		vt := value.GetType()
		if !hir.IsRefType(vt) {
			errors.ThrowExpectReferenceError(node.Value.Position(), vt)
		}
		return &hir.GetValue{Value: value}
	default:
		panic("unreachable")
	}
}

func (self *Analyser) analyseIdentExpr(node *ast.IdentExpr) hir.Expr {
	expr := self.analyseIdent((*ast.Ident)(node), true)
	if expr.IsNone(){
		errors.ThrowUnknownIdentifierError(node.Name.Position, node.Name)
	}
	value, _ := expr.MustValue().Left()
	return value
}

func (self *Analyser) analyseCall(node *ast.Call) *hir.Call {
	f := self.analyseExpr(nil, node.Func)
	if !hir.IsFuncType(f.GetType()){
		errors.ThrowExpectFunctionError(node.Func.Position(), f.GetType())
	}
	vararg := stlbasic.Is[*hir.FuncDef](f) && f.(*hir.FuncDef).VarArg
	ft := hir.AsFuncType(f.GetType())
	if (!vararg && len(node.Args) != len(ft.Params)) || (vararg && len(node.Args) < len(ft.Params)) {
		errors.ThrowParameterNumberNotMatchError(node.Position(), uint(len(ft.Params)), uint(len(node.Args)))
	}
	args := lo.Map(node.Args, func(item ast.Expr, index int) hir.Expr {
		if vararg && index >= len(ft.Params){
			return self.analyseExpr(nil, item)
		}else{
			return self.expectExpr(ft.Params[index], item)
		}
	})
	return &hir.Call{
		Func: f,
		Args: args,
	}
}

func (self *Analyser) analyseTuple(expect hir.Type, node *ast.Tuple) hir.Expr {
	if len(node.Elems) == 1 && (expect == nil || !hir.IsTupleType(expect)) {
		return self.analyseExpr(expect, node.Elems[0])
	}

	elemExpects := make([]hir.Type, len(node.Elems))
	if expect != nil {
		if hir.IsTupleType(expect) {
			tt := hir.AsTupleType(expect)
			if len(tt.Elems) < len(node.Elems) {
				copy(elemExpects, tt.Elems)
			} else if len(tt.Elems) > len(node.Elems) {
				elemExpects = tt.Elems[:len(node.Elems)]
			} else {
				elemExpects = tt.Elems
			}
		}
	}
	elems := lo.Map(node.Elems, func(item ast.Expr, index int) hir.Expr {
		return self.analyseExpr(elemExpects[index], item)
	})
	return &hir.Tuple{Elems: elems}
}

func (self *Analyser) analyseCovert(node *ast.Covert) hir.Expr {
	tt := self.analyseType(node.Type)
	from := self.analyseExpr(tt, node.Value)
	ft := from.GetType()
	if v, ok := self.autoTypeCovert(tt, from); ok {
		return v
	}

	switch {
	case hir.IsNumberType(ft) && hir.IsNumberType(tt):
		// i8 -> u8
		return &hir.Num2Num{
			From: from,
			To:   tt,
		}
	case hir.IsUnionType(ft) && hir.AsUnionType(ft).Contain(tt):
		if hir.IsUnionType(tt){
			// <i8,u8> -> <i8>
			return &hir.ShrinkUnion{
				Type:  tt,
				Value: from,
			}
		}else{
			// <i8,u8> -> i8
			return &hir.UnUnion{
				Type:  tt,
				Value: from,
			}
		}
	case hir.IsPointer(ft) && hir.IsPointer(tt):
		// *u8 | *?u8 | func() -> *u8 | *?u8 | func()
		return &hir.Pointer2Pointer{
			From: from,
			To:   tt,
		}
	case hir.IsPointer(ft) && tt.EqualTo(hir.Usize):
		// *u8 | *?u8 | func() -> usize
		return &hir.Pointer2Usize{
			From: from,
		}
	case ft.EqualTo(hir.Usize) && hir.IsPointer(tt):
		// usize -> *u8 | *?u8 | func()
		return &hir.Usize2Pointer{
			From: from,
			To: tt,
		}
	default:
		errors.ThrowIllegalCovertError(node.Position(), ft, tt)
		return nil
	}
}

func (self *Analyser) expectExpr(expect hir.Type, node ast.Expr) hir.Expr {
	value := self.analyseExpr(expect, node)
	newValue, ok := self.autoTypeCovert(expect, value)
	if !ok{
		errors.ThrowTypeMismatchError(node.Position(), value.GetType(), expect)
	}
	return newValue
}

// 自动类型转换
func (self *Analyser) autoTypeCovert(expect hir.Type, v hir.Expr) (hir.Expr, bool) {
	vt := v.GetType()
	if vt.EqualTo(expect) {
		return v, true
	}

	switch {
	case hir.IsUnionType(expect) && hir.AsUnionType(expect).Contain(vt):
		if hir.IsUnionType(vt){
			// <i8> -> <i8,u8>
			return &hir.ExpandUnion{
				Type:  expect,
				Value: v,
			}, true
		}else{
			// i8 -> <i8,u8>
			return &hir.Union{
				Type:  expect,
				Value: v,
			}, true
		}
	default:
		return v, false
	}
}

func (self *Analyser) analyseArray(expect hir.Type, node *ast.Array) *hir.Array {
	var expectArray *hir.ArrayType
	var expectElem hir.Type
	if expect != nil && hir.IsArrayType(expect){
		expectArray = hir.AsArrayType(expect)
		expectElem = expectArray.Elem
	}
	if expectArray == nil && len(node.Elems) == 0{
		errors.ThrowExpectArrayTypeError(node.Position(), hir.Empty)
	}
	elems := make([]hir.Expr, len(node.Elems))
	for i, elemNode := range node.Elems {
		elems[i] = stlbasic.TernaryAction(i==0, func() hir.Expr {
			return self.analyseExpr(expectElem, elemNode)
		}, func() hir.Expr {
			return self.expectExpr(elems[0].GetType(), elemNode)
		})
	}
	return &hir.Array{
		Type: stlbasic.TernaryAction(len(elems)!=0, func() hir.Type {
			return &hir.ArrayType{
				Size: uint(len(elems)),
				Elem: elems[0].GetType(),
			}
		}, func() hir.Type {return expectArray}),
		Elems: elems,
	}
}

func (self *Analyser) analyseIndex(node *ast.Index) *hir.Index {
	from := self.analyseExpr(nil, node.From)
	if !hir.IsArrayType(from.GetType()) {
		errors.ThrowExpectArrayError(node.From.Position(), from.GetType())
	}
	at := hir.AsArrayType(from.GetType())
	index := self.expectExpr(hir.Usize, node.Index)
	if stlbasic.Is[*hir.Integer](index) && index.(*hir.Integer).Value.Cmp(big.NewInt(int64(at.Size))) >= 0{
		errors.ThrowIndexOutOfRange(node.Index.Position())
	}
	return &hir.Index{
		From:  from,
		Index: index,
	}
}

func (self *Analyser) analyseExtract(expect hir.Type, node *ast.Extract) *hir.Extract {
	indexValue, ok := big.NewInt(0).SetString(node.Index.Source(), 10)
	if !ok {
		panic("unreachable")
	}
	if !indexValue.IsUint64() {
		panic("unreachable")
	}
	index := uint(indexValue.Uint64())

	expectFrom := &hir.TupleType{Elems: make([]hir.Type, index+1)}
	expectFrom.Elems[index] = expect

	from := self.analyseExpr(expectFrom, node.From)
	if !hir.IsTupleType(from.GetType()){
		errors.ThrowExpectTupleError(node.From.Position(), from.GetType())
	}

	tt := hir.AsTupleType(from.GetType())
	if index >= uint(len(tt.Elems)) {
		errors.ThrowInvalidIndexError(node.Index.Position, index)
	}
	return &hir.Extract{
		From:  from,
		Index: index,
	}
}

func (self *Analyser) analyseStruct(node *ast.Struct) *hir.Struct {
	stObj := self.analyseType(node.Type)
	if !hir.IsStructType(stObj){
		errors.ThrowExpectStructTypeError(node.Type.Position(), stObj)
	}
	st := hir.AsStructType(stObj)

	existedFields := make(map[string]hir.Expr)
	for _, nf := range node.Fields {
		fn := nf.First.Source()
		if !st.Fields.ContainKey(fn) || (!self.pkgScope.pkg.Equal(st.Pkg) && !st.Fields.Get(fn).Public) {
			errors.ThrowUnknownIdentifierError(nf.First.Position, nf.First)
		}
		existedFields[fn] = self.expectExpr(st.Fields.Get(fn).Type, nf.Second)
	}

	fields := make([]hir.Expr, st.Fields.Length())
	var i int
	for iter := st.Fields.Iterator(); iter.Next(); i++ {
		fn, ft := iter.Value().First, iter.Value().Second.Type
		if fv, ok := existedFields[fn]; ok {
			fields[i] = fv
		} else {
			fields[i] = self.getTypeDefaultValue(node.Type.Position(), ft)
		}
	}

	return &hir.Struct{
		Type:   stObj,
		Fields: fields,
	}
}

func (self *Analyser) analyseGetField(node *ast.GetField) hir.Expr {
	fieldName := node.Index.Source()

	var from hir.Expr
	if fromObj := self.analyseExpr(nil, node.From); hir.IsStructType(fromObj.GetType()){
		from = fromObj
	}else if ft := fromObj.GetType(); hir.IsRefType(ft) && hir.IsStructType(hir.AsRefType(ft).Elem){
		from = &hir.GetValue{Value: fromObj}
	}else{
		errors.ThrowExpectStructError(node.From.Position(), ft)
	}
	st := hir.AsStructType(from.GetType())

	// 方法 or 泛型方法
	if methodObj := st.Methods.Get(fieldName); methodObj != nil && !methodObj.IsStatic() && (methodObj.GetPublic() || st.Pkg == self.pkgScope.pkg){
		switch method := methodObj.(type) {
		case *hir.MethodDef:
			return &hir.Method{
				Self:   either.Left[hir.Expr, *hir.StructType](from),
				Define: method,
			}
		default:
			panic("unreachable")
		}
	}

	// 字段
	for i, iter := 0, st.Fields.Iterator(); iter.Next(); i++ {
		field := iter.Value()
		if field.First == fieldName && (field.Second.Public || self.pkgScope.pkg.Equal(st.Pkg)) {
			return &hir.GetField{
				Internal: self.pkgScope.pkg.Equal(st.Pkg),
				From:  from,
				Index: uint(i),
			}
		}
	}

	errors.ThrowUnknownIdentifierError(node.Index.Position, node.Index)
	return nil
}

func (self *Analyser) analyseString(node *ast.String) *hir.String {
	s := node.Value.Source()
	s = util.ParseEscapeCharacter(s[1:len(s)-1], `\"`, `"`)
	return &hir.String{Value: s}
}

func (self *Analyser) analyseJudgment(node *ast.Judgment) hir.Expr {
	target := self.analyseType(node.Type)
	value := self.analyseExpr(target, node.Value)
	vt := value.GetType()

	switch {
	case vt.EqualTo(target):
		return &hir.Boolean{Value: true}
	default:
		return &hir.TypeJudgment{
			Value: value,
			Type:  target,
		}
	}
}

func (self *Analyser) analyseNull(expect hir.Type, node *ast.Null) *hir.Default {
	if expect == nil {
		errors.ThrowExpectPointerTypeError(node.Position(), hir.Empty)
	}else if !hir.IsPointer(expect){
		errors.ThrowExpectPointerTypeError(node.Position(), expect)
	}
	return &hir.Default{Type: expect}
}

func (self *Analyser) analyseStaticMethod(node *ast.StaticMethod)hir.Expr{
	stObj := self.analyseType(node.Type)
	if !hir.IsStructType(stObj){
		errors.ThrowExpectStructError(node.Type.Position(), stObj)
	}
	st := hir.AsStructType(stObj)

	f := st.Methods.Get(node.Name.Source())
	if f == nil || (!f.GetPublic() && !self.pkgScope.pkg.Equal(f.GetPackage())){
		errors.ThrowUnknownIdentifierError(node.Name.Position, node.Name)
	}

	switch method := f.(type) {
	case *hir.MethodDef:
		return &hir.Method{
			Self:   either.Right[hir.Expr, *hir.StructType](st),
			Define: method,
		}
	default:
		panic("unreachable")
	}
}
