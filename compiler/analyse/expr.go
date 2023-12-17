package analyse

import (
	"math/big"

	stlbasic "github.com/kkkunny/stl/basic"
	"github.com/kkkunny/stl/container/hashset"
	stlerror "github.com/kkkunny/stl/error"
	"github.com/samber/lo"

	"github.com/kkkunny/Sim/hir"
	"github.com/kkkunny/Sim/reader"

	"github.com/kkkunny/Sim/ast"
	errors "github.com/kkkunny/Sim/error"
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
		return self.analyseArray(exprNode)
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
	if expect == nil || !stlbasic.Is[hir.NumberType](expect) {
		expect = hir.Isize
	}
	switch t := expect.(type) {
	case hir.IntType:
		value, ok := big.NewInt(0).SetString(node.Value.Source(), 10)
		if !ok {
			panic("unreachable")
		}
		return &hir.Integer{
			Type:  t,
			Value: value,
		}
	case *hir.FloatType:
		value, _ := stlerror.MustWith2(big.ParseFloat(node.Value.Source(), 10, big.MaxPrec, big.ToZero))
		return &hir.Float{
			Type:  t,
			Value: value,
		}
	default:
		panic("unreachable")
	}
}

func (self *Analyser) analyseChar(expect hir.Type, node *ast.Char) hir.Expr {
	if expect == nil || !stlbasic.Is[hir.NumberType](expect) {
		expect = hir.I32
	}
	s := node.Value.Source()
	char := util.ParseEscapeCharacter(s[1:len(s)-1], `\'`, `'`)[0]
	switch t := expect.(type) {
	case hir.IntType:
		value := big.NewInt(int64(char))
		return &hir.Integer{
			Type:  t,
			Value: value,
		}
	case *hir.FloatType:
		value := big.NewFloat(float64(char))
		return &hir.Float{
			Type:  t,
			Value: value,
		}
	default:
		panic("unreachable")
	}
}

func (self *Analyser) analyseFloat(expect hir.Type, node *ast.Float) *hir.Float {
	if expect == nil || !stlbasic.Is[*hir.FloatType](expect) {
		expect = hir.F64
	}
	value, _ := stlerror.MustWith2(big.NewFloat(0).Parse(node.Value.Source(), 10))
	return &hir.Float{
		Type:  expect.(*hir.FloatType),
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
		if lt.Equal(rt) {
			if !left.Mutable() {
				errors.ThrowNotMutableError(node.Left.Position())
			}
			return &hir.Assign{
				Left:  left,
				Right: right,
			}
		}
	case token.AND:
		if lt.Equal(rt) && stlbasic.Is[hir.IntType](lt) {
			return &hir.IntAndInt{
				Left:  left,
				Right: right,
			}
		}
	case token.OR:
		if lt.Equal(rt) && stlbasic.Is[hir.IntType](lt) {
			return &hir.IntOrInt{
				Left:  left,
				Right: right,
			}
		}
	case token.XOR:
		if lt.Equal(rt) && stlbasic.Is[hir.IntType](lt) {
			return &hir.IntXorInt{
				Left:  left,
				Right: right,
			}
		}
	case token.SHL:
		if lt.Equal(rt) && stlbasic.Is[hir.IntType](lt) {
			return &hir.IntShlInt{
				Left:  left,
				Right: right,
			}
		}
	case token.SHR:
		if lt.Equal(rt) && stlbasic.Is[hir.IntType](lt) {
			return &hir.IntShrInt{
				Left:  left,
				Right: right,
			}
		}
	case token.ADD:
		if lt.Equal(rt) && stlbasic.Is[hir.NumberType](lt) {
			return &hir.NumAddNum{
				Left:  left,
				Right: right,
			}
		}
	case token.SUB:
		if lt.Equal(rt) && stlbasic.Is[hir.NumberType](lt) {
			return &hir.NumSubNum{
				Left:  left,
				Right: right,
			}
		}
	case token.MUL:
		if lt.Equal(rt) && stlbasic.Is[hir.NumberType](lt) {
			return &hir.NumMulNum{
				Left:  left,
				Right: right,
			}
		}
	case token.DIV:
		if lt.Equal(rt) && stlbasic.Is[hir.NumberType](lt) {
			return &hir.NumDivNum{
				Left:  left,
				Right: right,
			}
		}
	case token.REM:
		if lt.Equal(rt) && stlbasic.Is[hir.NumberType](lt) {
			return &hir.NumRemNum{
				Left:  left,
				Right: right,
			}
		}
	case token.EQ:
		if lt.Equal(rt) {
			if stlbasic.Is[hir.NumberType](lt) || stlbasic.Is[*hir.BoolType](lt) || stlbasic.Is[*hir.FuncType](lt) ||
			stlbasic.Is[*hir.ArrayType](lt) || stlbasic.Is[*hir.TupleType](lt) || stlbasic.Is[*hir.StructType](lt) ||
			stlbasic.Is[*hir.StringType](lt) || stlbasic.Is[*hir.UnionType](lt) || stlbasic.Is[*hir.PtrType](lt) ||
			stlbasic.Is[*hir.RefType](lt){
				return &hir.Equal{
					Left:  left,
					Right: right,
				}
			}
		}
	case token.NE:
		if stlbasic.Is[hir.NumberType](lt) || stlbasic.Is[*hir.BoolType](lt) || stlbasic.Is[*hir.FuncType](lt) ||
			stlbasic.Is[*hir.ArrayType](lt) || stlbasic.Is[*hir.TupleType](lt) || stlbasic.Is[*hir.StructType](lt) ||
			stlbasic.Is[*hir.StringType](lt) || stlbasic.Is[*hir.UnionType](lt) || stlbasic.Is[*hir.PtrType](lt) ||
			stlbasic.Is[*hir.RefType](lt){
			return &hir.NotEqual{
				Left:  left,
				Right: right,
			}
		}
	case token.LT:
		if lt.Equal(rt) && stlbasic.Is[hir.NumberType](lt) {
			return &hir.NumLtNum{
				Left:  left,
				Right: right,
			}
		}
	case token.GT:
		if lt.Equal(rt) && stlbasic.Is[hir.NumberType](lt) {
			return &hir.NumGtNum{
				Left:  left,
				Right: right,
			}
		}
	case token.LE:
		if lt.Equal(rt) && stlbasic.Is[hir.NumberType](lt) {
			return &hir.NumLeNum{
				Left:  left,
				Right: right,
			}
		}
	case token.GE:
		if lt.Equal(rt) && stlbasic.Is[hir.NumberType](lt) {
			return &hir.NumGeNum{
				Left:  left,
				Right: right,
			}
		}
	case token.LAND:
		if lt.Equal(rt) && stlbasic.Is[*hir.BoolType](lt) {
			return &hir.BoolAndBool{
				Left:  left,
				Right: right,
			}
		}
	case token.LOR:
		if lt.Equal(rt) && stlbasic.Is[*hir.BoolType](lt) {
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
		if stlbasic.Is[*hir.SintType](vt) || stlbasic.Is[*hir.FloatType](vt) {
			return &hir.NumNegate{Value: value}
		}
		errors.ThrowIllegalUnaryError(node.Position(), node.Opera, vt)
		return nil
	case token.NOT:
		value := self.analyseExpr(expect, node.Value)
		vt := value.GetType()
		switch {
		case stlbasic.Is[hir.IntType](vt):
			return &hir.IntBitNegate{Value: value}
		case stlbasic.Is[*hir.BoolType](vt):
			return &hir.BoolNegate{Value: value}
		default:
			errors.ThrowIllegalUnaryError(node.Position(), node.Opera, vt)
			return nil
		}
	case token.AND:
		if expect != nil && stlbasic.Is[*hir.RefType](expect) {
			expect = expect.(*hir.RefType).Elem
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
		v := self.analyseExpr(expect, node.Value)
		vt := v.GetType()
		if !stlbasic.Is[*hir.RefType](vt) {
			errors.ThrowExpectReferenceError(node.Value.Position(), vt)
		}
		return &hir.GetValue{Value: v}
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
	if len(node.GenericArgs) == 0{
		// 普通标识符
		value, ok := self.localScope.GetValue(pkgName, node.Name.Source())
		if !ok {
			errors.ThrowUnknownIdentifierError(node.Name.Position, node.Name)
		}
		return value
	}else{
		// 泛型函数
		gf, ok := self.pkgScope.GetGenericFunction(pkgName, node.Name.Source())
		if !ok {
			errors.ThrowUnknownIdentifierError(node.Name.Position, node.Name)
		}
		if gf.GenericParams.Length() != uint(len(node.GenericArgs)){
			errors.ThrowParameterNumberNotMatchError(node.Name.Position, gf.GenericParams.Length(), uint(len(node.GenericArgs)))
		}
		params := lo.Map(node.GenericArgs, func(item ast.Type, _ int) hir.Type {
			return self.analyseType(item)
		})
		return gf.AddInstance(params...)
	}
}

func (self *Analyser) analyseCall(node *ast.Call) *hir.Call {
	f := self.analyseExpr(nil, node.Func)
	ft, ok := f.GetType().(*hir.FuncType)
	if !ok {
		errors.ThrowNotFunctionError(node.Func.Position(), f.GetType())
	} else if len(ft.Params) != len(node.Args) {
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
	if len(node.Elems) == 1 && (expect == nil || !stlbasic.Is[*hir.TupleType](expect)) {
		return self.analyseExpr(expect, node.Elems[0])
	}

	elemExpects := make([]hir.Type, len(node.Elems))
	if expect != nil {
		if tt, ok := expect.(*hir.TupleType); ok {
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
	if ft.AssignableTo(tt) {
		return self.autoTypeCovert(tt, from)
	}

	switch {
	case stlbasic.Is[hir.NumberType](ft) && stlbasic.Is[hir.NumberType](tt):
		return &hir.Num2Num{
			From: from,
			To:   tt.(hir.NumberType),
		}
	case stlbasic.Is[*hir.UnionType](tt) && tt.(*hir.UnionType).GetElemIndex(ft) >= 0:
		return &hir.Union{
			Type:  tt.(*hir.UnionType),
			Value: from,
		}
	case stlbasic.Is[*hir.UnionType](ft) && ft.(*hir.UnionType).GetElemIndex(tt) >= 0:
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
	if vt := value.GetType(); !vt.AssignableTo(expect) {
		errors.ThrowTypeMismatchError(node.Position(), vt, expect)
	}
	return self.autoTypeCovert(expect, value)
}

// 自动类型转换
func (self *Analyser) autoTypeCovert(expect hir.Type, v hir.Expr) hir.Expr {
	vt := v.GetType()
	if vt.Equal(expect) {
		return v
	}

	switch {
	case stlbasic.Is[*hir.UnionType](expect) && expect.(*hir.UnionType).Contain(vt):
		return &hir.Union{
			Type:  expect.(*hir.UnionType),
			Value: v,
		}
	case stlbasic.Is[*hir.RefType](vt) && vt.(*hir.RefType).ToPtrType().Equal(expect):
		return &hir.WrapWithNull{Value: v}
	default:
		panic("unreachable")
	}
}

func (self *Analyser) analyseArray(node *ast.Array) *hir.Array {
	t := self.analyseArrayType(node.Type)
	elems := make([]hir.Expr, len(node.Elems))
	for i, en := range node.Elems {
		elems[i] = self.expectExpr(t.Elem, en)
	}
	return &hir.Array{
		Type:  t,
		Elems: elems,
	}
}

func (self *Analyser) analyseIndex(node *ast.Index) *hir.Index {
	from := self.analyseExpr(nil, node.From)
	if !stlbasic.Is[*hir.ArrayType](from.GetType()) {
		errors.ThrowNotArrayError(node.From.Position(), from.GetType())
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
	tt, ok := from.GetType().(*hir.TupleType)
	if !ok {
		errors.ThrowNotTupleError(node.From.Position(), from.GetType())
	}

	if index >= uint(len(tt.Elems)) {
		errors.ThrowInvalidIndexError(node.Index.Position, index)
	}
	return &hir.Extract{
		From:  from,
		Index: index,
	}
}

func (self *Analyser) analyseStruct(node *ast.Struct) *hir.Struct {
	st := self.analyseIdentType(node.Type).(*hir.StructType)
	fieldNames := hashset.NewHashSet[string]()
	for iter := st.Fields.Keys().Iterator(); iter.Next(); {
		fieldNames.Add(iter.Value())
	}

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
	if st, ok := from.GetType().(*hir.StructType); ok{
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
	}else{
		errors.ThrowNotStructError(node.From.Position(), from.GetType())
		return nil
	}
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
	case vt.Equal(target):
		return &hir.Boolean{Value: true}
	case stlbasic.Is[*hir.UnionType](vt) && vt.(*hir.UnionType).GetElemIndex(target) >= 0:
		return &hir.UnionTypeJudgment{
			Value: value,
			Type:  target,
		}
	default:
		return &hir.Boolean{Value: false}
	}
}

func (self *Analyser) analyseNull(expect hir.Type, node *ast.Null) *hir.Default {
	if expect == nil || !(stlbasic.Is[*hir.PtrType](expect) || stlbasic.Is[*hir.RefType](expect)) {
		errors.ThrowExpectPointerTypeError(node.Position())
	}
	var t *hir.PtrType
	if stlbasic.Is[*hir.PtrType](expect) {
		t = expect.(*hir.PtrType)
	} else {
		t = expect.(*hir.RefType).ToPtrType()
	}
	return &hir.Default{Type: t}
}

func (self *Analyser) analyseCheckNull(expect hir.Type, node *ast.CheckNull) *hir.CheckNull {
	if expect != nil {
		if pt, ok := expect.(*hir.PtrType); ok {
			expect = pt
		} else if rt, ok := expect.(*hir.RefType); ok {
			expect = rt.ToPtrType()
		}
	}
	value := self.analyseExpr(expect, node.Value)
	vt := value.GetType()
	if !stlbasic.Is[*hir.PtrType](vt) {
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

func (self *Analyser) getTypeDefaultValue(pos reader.Position, t hir.Type) hir.Expr{
	if !t.HasDefault(){
		errors.ThrowCanNotGetDefault(pos, t)
	}
	return &hir.Default{Type: t}
}
