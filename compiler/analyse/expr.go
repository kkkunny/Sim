package analyse

import (
	"math/big"

	stlbasic "github.com/kkkunny/stl/basic"
	"github.com/kkkunny/stl/container/hashset"
	"github.com/kkkunny/stl/container/iterator"
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
	case *ast.Boolean:
		return self.analyseBool(exprNode)
	case *ast.Binary:
		return self.analyseBinary(expect, exprNode)
	case *ast.Unary:
		return self.analyseUnary(expect, exprNode)
	case *ast.Ident:
		return self.analyseIdent(exprNode)
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
	case *ast.Field:
		return self.analyseField(exprNode)
	case *ast.String:
		return self.analyseString(exprNode)
	case *ast.Judgment:
		return self.analyseJudgment(exprNode)
	case *ast.Null:
		return self.analyseNull(expect, exprNode)
	case *ast.CheckNull:
		return self.analyseCheckNull(expect, exprNode)
	case *ast.SelfValue:
		return self.analyseSelfValue(exprNode)
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

func (self *Analyser) analyseBool(node *ast.Boolean) *hir.Boolean {
	return &hir.Boolean{Value: node.Value.Is(token.TRUE)}
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
			return &hir.NumDivNum{
				Left:  left,
				Right: right,
			}
		}
	case token.REM:
		if lt.EqualTo(rt) && hir.IsNumberType(lt) {
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
		}
		return &hir.GetPtr{Value: value}
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

func (self *Analyser) analyseIdent(node *ast.Ident) hir.Expr {
	var pkgName string
	if pkgToken, ok := node.Pkg.Value(); ok {
		pkgName = pkgToken.Source()
		if !self.pkgScope.externs.ContainKey(pkgName) {
			errors.ThrowUnknownIdentifierError(node.Position(), node.Name)
		}
	}
	// 普通标识符
	value, ok := self.localScope.GetValue(pkgName, node.Name.Source())
	if !ok {
		errors.ThrowUnknownIdentifierError(node.Name.Position, node.Name)
	}
	return value
}

func (self *Analyser) analyseCall(node *ast.Call) *hir.Call {
	f := self.analyseExpr(nil, node.Func)
	if !hir.IsFuncType(f.GetType()){
		errors.ThrowExpectFunctionError(node.Func.Position(), f.GetType())
	}
	ft := hir.AsFuncType(f.GetType())
	if len(ft.Params) != len(node.Args) {
		errors.ThrowParameterNumberNotMatchError(node.Position(), uint(len(ft.Params)), uint(len(node.Args)))
	}
	args := lo.Map(node.Args, func(item ast.Expr, index int) hir.Expr {
		return self.analyseExpr(ft.Params[index], item)
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
	case hir.IsUnionType(ft) && hir.AsUnionType(ft).GetElemIndex(tt) >= 0:
		// <i8,u8> -> i8
		return &hir.UnUnion{
			Type:  tt,
			Value: from,
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
		// i8 -> <i8,u8>
		return &hir.Union{
			Type:  expect,
			Value: v,
		}, true
	case hir.IsRefType(vt) && hir.AsRefType(vt).ToPtrType().EqualTo(expect):
		// *i8 -> *?i8
		return &hir.WrapWithNull{Value: v}, true
	default:
		return v, false
	}
}

func (self *Analyser) analyseArray(expect hir.Type, node *ast.Array) *hir.Array {
	var expectElem hir.Type
	if expect != nil && hir.IsArrayType(expect){
		expectElem = hir.AsArrayType(expect).Elem
	}
	if expectElem == nil && len(node.Elems) == 0{
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
	return &hir.Array{Elems: elems}
}

func (self *Analyser) analyseIndex(node *ast.Index) *hir.Index {
	from := self.analyseExpr(nil, node.From)
	if !hir.IsArrayType(from.GetType()) {
		errors.ThrowExpectArrayError(node.From.Position(), from.GetType())
	}
	index := self.expectExpr(hir.Usize, node.Index)
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
	stObj := self.analyseIdentType(node.Type)
	if !hir.IsStructType(stObj){
		errors.ThrowExpectStructTypeError(node.Type.Position(), stObj)
	}
	st := hir.AsStructType(stObj)
	fieldNames := iterator.Map[string, string, hashset.HashSet[string]](st.Fields.Keys(), func(s string) string {return s})

	existedFields := make(map[string]hir.Expr)
	for _, nf := range node.Fields {
		fn := nf.First.Source()
		if !fieldNames.Contain(fn) {
			errors.ThrowIdentifierDuplicationError(nf.First.Position, nf.First)
		}
		existedFields[fn] = self.expectExpr(st.Fields.Get(fn).Second, nf.Second)
	}

	fields := make([]hir.Expr, st.Fields.Length())
	var i int
	for iter := st.Fields.Iterator(); iter.Next(); i++ {
		fn, ft := iter.Value().First, iter.Value().Second.Second
		if fv, ok := existedFields[fn]; ok {
			fields[i] = fv
		} else {
			fields[i] = self.getTypeDefaultValue(node.Type.Position(), ft)
		}
	}

	return &hir.Struct{
		Type:   st,
		Fields: fields,
	}
}

func (self *Analyser) analyseField(node *ast.Field) hir.Expr {
	from := self.analyseExpr(nil, node.From)
	fieldName := node.Index.Source()
	if !hir.IsStructType(from.GetType()){
		errors.ThrowExpectStructError(node.From.Position(), from.GetType())
	}
	st := hir.AsStructType(from.GetType())

	// 方法
	if method := st.Methods.Get(fieldName); method != nil && (method.Public || st.Pkg == self.pkgScope.pkg){
		return &hir.Method{
			Self:   from,
			Define: method,
		}
	}

	// 字段
	var i int
	for iter := st.Fields.Iterator(); iter.Next(); i++ {
		field := iter.Value()
		if field.First == fieldName && (field.Second.First || st.Pkg == self.pkgScope.pkg) {
			return &hir.Field{
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
	case hir.IsUnionType(vt) && hir.AsUnionType(vt).GetElemIndex(target) >= 0:
		return &hir.UnionTypeJudgment{
			Value: value,
			Type:  target,
		}
	default:
		return &hir.Boolean{Value: false}
	}
}

func (self *Analyser) analyseNull(expect hir.Type, node *ast.Null) *hir.Default {
	if expect == nil {
		errors.ThrowExpectPointerTypeError(node.Position(), hir.Empty)
	}else if !hir.IsPtrType(expect) && !hir.IsFuncType(expect){
		errors.ThrowExpectPointerTypeError(node.Position(), expect)
	}
	return &hir.Default{Type: expect}
}

func (self *Analyser) analyseCheckNull(expect hir.Type, node *ast.CheckNull) *hir.CheckNull {
	if expect != nil {
		if hir.IsPtrType(expect){
			expect = hir.AsPtrType(expect)
		} else if hir.IsRefType(expect) {
			expect = hir.AsRefType(expect).ToPtrType()
		}
	}
	value := self.analyseExpr(expect, node.Value)
	vt := value.GetType()
	if !hir.IsPtrType(vt) {
		errors.ThrowExpectPointerError(node.Value.Position(), vt)
	}
	return &hir.CheckNull{Value: value}
}

func (self *Analyser) analyseSelfValue(node *ast.SelfValue)*hir.Param{
	if self.selfValue == nil{
		errors.ThrowUnknownIdentifierError(node.Position(), node.Token)
	}
	return self.selfValue
}
