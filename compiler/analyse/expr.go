package analyse

import (
	"math/big"

	stlbasic "github.com/kkkunny/stl/basic"
	"github.com/kkkunny/stl/container/hashset"
	stlerror "github.com/kkkunny/stl/error"
	"github.com/samber/lo"

	"github.com/kkkunny/Sim/mean"
	"github.com/kkkunny/Sim/reader"

	"github.com/kkkunny/Sim/ast"
	errors "github.com/kkkunny/Sim/error"
	"github.com/kkkunny/Sim/token"
	"github.com/kkkunny/Sim/util"
)

func (self *Analyser) analyseExpr(expect mean.Type, node ast.Expr) mean.Expr {
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

func (self *Analyser) analyseInteger(expect mean.Type, node *ast.Integer) mean.Expr {
	if expect == nil || !stlbasic.Is[mean.NumberType](expect) {
		expect = mean.Isize
	}
	switch t := expect.(type) {
	case mean.IntType:
		value, ok := big.NewInt(0).SetString(node.Value.Source(), 10)
		if !ok {
			panic("unreachable")
		}
		return &mean.Integer{
			Type:  t,
			Value: *value,
		}
	case *mean.FloatType:
		value, _ := stlerror.MustWith2(big.ParseFloat(node.Value.Source(), 10, big.MaxPrec, big.ToZero))
		return &mean.Float{
			Type:  t,
			Value: *value,
		}
	default:
		panic("unreachable")
	}
}

func (self *Analyser) analyseChar(expect mean.Type, node *ast.Char) mean.Expr {
	if expect == nil || !stlbasic.Is[mean.NumberType](expect) {
		expect = mean.I32
	}
	s := node.Value.Source()
	char := util.ParseEscapeCharacter(s[1:len(s)-1], `\'`, `'`)[0]
	switch t := expect.(type) {
	case mean.IntType:
		value := big.NewInt(int64(char))
		return &mean.Integer{
			Type:  t,
			Value: *value,
		}
	case *mean.FloatType:
		value := big.NewFloat(float64(char))
		return &mean.Float{
			Type:  t,
			Value: *value,
		}
	default:
		panic("unreachable")
	}
}

func (self *Analyser) analyseFloat(expect mean.Type, node *ast.Float) *mean.Float {
	if expect == nil || !stlbasic.Is[*mean.FloatType](expect) {
		expect = mean.F64
	}
	value, _ := stlerror.MustWith2(big.NewFloat(0).Parse(node.Value.Source(), 10))
	return &mean.Float{
		Type:  expect.(*mean.FloatType),
		Value: *value,
	}
}

func (self *Analyser) analyseBool(node *ast.Boolean) *mean.Boolean {
	return &mean.Boolean{Value: node.Value.Is(token.TRUE)}
}

func (self *Analyser) analyseBinary(expect mean.Type, node *ast.Binary) mean.Binary {
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
			return &mean.Assign{
				Left:  left,
				Right: right,
			}
		}
	case token.AND:
		if lt.Equal(rt) && stlbasic.Is[mean.IntType](lt) {
			return &mean.IntAndInt{
				Left:  left,
				Right: right,
			}
		}
	case token.OR:
		if lt.Equal(rt) && stlbasic.Is[mean.IntType](lt) {
			return &mean.IntOrInt{
				Left:  left,
				Right: right,
			}
		}
	case token.XOR:
		if lt.Equal(rt) && stlbasic.Is[mean.IntType](lt) {
			return &mean.IntXorInt{
				Left:  left,
				Right: right,
			}
		}
	case token.SHL:
		if lt.Equal(rt) && stlbasic.Is[mean.IntType](lt) {
			return &mean.IntShlInt{
				Left:  left,
				Right: right,
			}
		}
	case token.SHR:
		if lt.Equal(rt) && stlbasic.Is[mean.IntType](lt) {
			return &mean.IntShrInt{
				Left:  left,
				Right: right,
			}
		}
	case token.ADD:
		if lt.Equal(rt) && stlbasic.Is[mean.NumberType](lt) {
			return &mean.NumAddNum{
				Left:  left,
				Right: right,
			}
		}
	case token.SUB:
		if lt.Equal(rt) && stlbasic.Is[mean.NumberType](lt) {
			return &mean.NumSubNum{
				Left:  left,
				Right: right,
			}
		}
	case token.MUL:
		if lt.Equal(rt) && stlbasic.Is[mean.NumberType](lt) {
			return &mean.NumMulNum{
				Left:  left,
				Right: right,
			}
		}
	case token.DIV:
		if lt.Equal(rt) && stlbasic.Is[mean.NumberType](lt) {
			return &mean.NumDivNum{
				Left:  left,
				Right: right,
			}
		}
	case token.REM:
		if lt.Equal(rt) && stlbasic.Is[mean.NumberType](lt) {
			return &mean.NumRemNum{
				Left:  left,
				Right: right,
			}
		}
	case token.EQ:
		if lt.Equal(rt) {
			switch {
			case stlbasic.Is[mean.NumberType](lt):
				return &mean.NumEqNum{
					Left:  left,
					Right: right,
				}
			case stlbasic.Is[*mean.BoolType](lt):
				return &mean.BoolEqBool{
					Left:  left,
					Right: right,
				}
			case stlbasic.Is[*mean.FuncType](lt):
				return &mean.FuncEqFunc{
					Left:  left,
					Right: right,
				}
			case stlbasic.Is[*mean.ArrayType](lt):
				return &mean.ArrayEqArray{
					Left:  left,
					Right: right,
				}
			case stlbasic.Is[*mean.TupleType](lt):
				return &mean.TupleEqTuple{
					Left:  left,
					Right: right,
				}
			case stlbasic.Is[*mean.StructType](lt):
				return &mean.StructEqStruct{
					Left:  left,
					Right: right,
				}
			case stlbasic.Is[*mean.StringType](lt):
				return &mean.StringEqString{
					Left:  left,
					Right: right,
				}
			case stlbasic.Is[*mean.UnionType](lt):
				return &mean.UnionEqUnion{
					Left:  left,
					Right: right,
				}
			case stlbasic.Is[*mean.StringType](lt):
				return &mean.StringEqString{
					Left:  left,
					Right: right,
				}
			case stlbasic.Is[*mean.UnionType](lt):
				return &mean.UnionEqUnion{
					Left:  left,
					Right: right,
				}
			}
		}
	case token.NE:
		if lt.Equal(rt) {
			switch {
			case stlbasic.Is[mean.NumberType](lt):
				return &mean.NumNeNum{
					Left:  left,
					Right: right,
				}
			case stlbasic.Is[*mean.BoolType](lt):
				return &mean.BoolNeBool{
					Left:  left,
					Right: right,
				}
			case stlbasic.Is[*mean.FuncType](lt):
				return &mean.FuncNeFunc{
					Left:  left,
					Right: right,
				}
			case stlbasic.Is[*mean.ArrayType](lt):
				return &mean.ArrayNeArray{
					Left:  left,
					Right: right,
				}
			case stlbasic.Is[*mean.TupleType](lt):
				return &mean.TupleNeTuple{
					Left:  left,
					Right: right,
				}
			case stlbasic.Is[*mean.StructType](lt):
				return &mean.StructNeStruct{
					Left:  left,
					Right: right,
				}
			case stlbasic.Is[*mean.StringType](lt):
				return &mean.StringNeString{
					Left:  left,
					Right: right,
				}
			case stlbasic.Is[*mean.UnionType](lt):
				return &mean.UnionNeUnion{
					Left:  left,
					Right: right,
				}
			case stlbasic.Is[*mean.StringType](lt):
				return &mean.StringNeString{
					Left:  left,
					Right: right,
				}
			case stlbasic.Is[*mean.UnionType](lt):
				return &mean.UnionNeUnion{
					Left:  left,
					Right: right,
				}
			}
		}
	case token.LT:
		if lt.Equal(rt) && stlbasic.Is[mean.NumberType](lt) {
			return &mean.NumLtNum{
				Left:  left,
				Right: right,
			}
		}
	case token.GT:
		if lt.Equal(rt) && stlbasic.Is[mean.NumberType](lt) {
			return &mean.NumGtNum{
				Left:  left,
				Right: right,
			}
		}
	case token.LE:
		if lt.Equal(rt) && stlbasic.Is[mean.NumberType](lt) {
			return &mean.NumLeNum{
				Left:  left,
				Right: right,
			}
		}
	case token.GE:
		if lt.Equal(rt) && stlbasic.Is[mean.NumberType](lt) {
			return &mean.NumGeNum{
				Left:  left,
				Right: right,
			}
		}
	case token.LAND:
		if lt.Equal(rt) && stlbasic.Is[*mean.BoolType](lt) {
			return &mean.BoolAndBool{
				Left:  left,
				Right: right,
			}
		}
	case token.LOR:
		if lt.Equal(rt) && stlbasic.Is[*mean.BoolType](lt) {
			return &mean.BoolOrBool{
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

func (self *Analyser) analyseUnary(expect mean.Type, node *ast.Unary) mean.Unary {
	switch node.Opera.Kind {
	case token.SUB:
		value := self.analyseExpr(expect, node.Value)
		vt := value.GetType()
		if stlbasic.Is[*mean.SintType](vt) || stlbasic.Is[*mean.FloatType](vt) {
			return &mean.NumNegate{Value: value}
		}
		errors.ThrowIllegalUnaryError(node.Position(), node.Opera, vt)
		return nil
	case token.NOT:
		value := self.analyseExpr(expect, node.Value)
		vt := value.GetType()
		switch {
		case stlbasic.Is[mean.IntType](vt):
			return &mean.IntBitNegate{Value: value}
		case stlbasic.Is[*mean.BoolType](vt):
			return &mean.BoolNegate{Value: value}
		default:
			errors.ThrowIllegalUnaryError(node.Position(), node.Opera, vt)
			return nil
		}
	case token.AND:
		if expect != nil && stlbasic.Is[*mean.RefType](expect) {
			expect = expect.(*mean.RefType).Elem
		}
		value := self.analyseExpr(expect, node.Value)
		if !stlbasic.Is[mean.Ident](value) {
			errors.ThrowCanNotGetPointer(node.Value.Position())
		}
		return &mean.GetPtr{Value: value}
	case token.MUL:
		if expect != nil {
			expect = &mean.RefType{Elem: expect}
		}
		v := self.analyseExpr(expect, node.Value)
		vt := v.GetType()
		if !stlbasic.Is[*mean.RefType](vt) {
			errors.ThrowExpectReferenceError(node.Value.Position(), vt)
		}
		return &mean.GetValue{Value: v}
	default:
		panic("unreachable")
	}
}

func (self *Analyser) analyseIdent(node *ast.Ident) mean.Ident {
	var pkgName string
	if pkgToken, ok := node.Pkg.Value(); ok {
		pkgName = pkgToken.Source()
		if !self.pkgScope.externs.ContainKey(pkgName) {
			errors.ThrowUnknownIdentifierError(node.Position(), node.Name)
		}
	}
	value, ok := self.localScope.GetValue(pkgName, node.Name.Source())
	if !ok {
		errors.ThrowUnknownIdentifierError(node.Position(), node.Name)
	}
	return value
}

func (self *Analyser) analyseCall(node *ast.Call) *mean.Call {
	f := self.analyseExpr(nil, node.Func)
	ft, ok := f.GetType().(*mean.FuncType)
	if !ok {
		errors.ThrowNotFunctionError(node.Func.Position(), f.GetType())
	} else if len(ft.Params) != len(node.Args) {
		errors.ThrowParameterNumberNotMatchError(node.Position(), uint(len(ft.Params)), uint(len(node.Args)))
	}
	args := lo.Map(node.Args, func(item ast.Expr, index int) mean.Expr {
		return self.analyseExpr(ft.Params[index], item)
	})
	return &mean.Call{
		Func: f,
		Args: args,
	}
}

func (self *Analyser) analyseTuple(expect mean.Type, node *ast.Tuple) mean.Expr {
	if len(node.Elems) == 1 && (expect == nil || !stlbasic.Is[*mean.TupleType](expect)) {
		return self.analyseExpr(expect, node.Elems[0])
	}

	elemExpects := make([]mean.Type, len(node.Elems))
	if expect != nil {
		if tt, ok := expect.(*mean.TupleType); ok {
			if len(tt.Elems) < len(node.Elems) {
				copy(elemExpects, tt.Elems)
			} else if len(tt.Elems) > len(node.Elems) {
				elemExpects = tt.Elems[:len(node.Elems)]
			} else {
				elemExpects = tt.Elems
			}
		}
	}
	elems := lo.Map(node.Elems, func(item ast.Expr, index int) mean.Expr {
		return self.analyseExpr(elemExpects[index], item)
	})
	return &mean.Tuple{Elems: elems}
}

func (self *Analyser) analyseCovert(node *ast.Covert) mean.Expr {
	tt := self.analyseType(node.Type)
	from := self.analyseExpr(tt, node.Value)
	ft := from.GetType()
	if ft.AssignableTo(tt) {
		return self.autoTypeCovert(tt, from)
	}

	switch {
	case stlbasic.Is[mean.NumberType](ft) && stlbasic.Is[mean.NumberType](tt):
		return &mean.Num2Num{
			From: from,
			To:   tt.(mean.NumberType),
		}
	case stlbasic.Is[*mean.UnionType](tt) && tt.(*mean.UnionType).GetElemIndex(ft) >= 0:
		return &mean.Union{
			Type:  tt.(*mean.UnionType),
			Value: from,
		}
	case stlbasic.Is[*mean.UnionType](ft) && ft.(*mean.UnionType).GetElemIndex(tt) >= 0:
		return &mean.UnUnion{
			Type:  tt,
			Value: from,
		}
	default:
		errors.ThrowIllegalCovertError(node.Position(), ft, tt)
		return nil
	}
}

func (self *Analyser) expectExpr(expect mean.Type, node ast.Expr) mean.Expr {
	value := self.analyseExpr(expect, node)
	if vt := value.GetType(); !vt.AssignableTo(expect) {
		errors.ThrowTypeMismatchError(node.Position(), vt, expect)
	}
	return self.autoTypeCovert(expect, value)
}

// 自动类型转换
func (self *Analyser) autoTypeCovert(expect mean.Type, v mean.Expr) mean.Expr {
	vt := v.GetType()
	if vt.Equal(expect) {
		return v
	}

	switch {
	case stlbasic.Is[*mean.UnionType](expect) && expect.(*mean.UnionType).Elems.ContainKey(vt.String()):
		return &mean.Union{
			Type:  expect.(*mean.UnionType),
			Value: v,
		}
	case stlbasic.Is[*mean.RefType](vt) && vt.(*mean.RefType).ToPtrType().Equal(expect):
		return &mean.WrapWithNull{Value: v}
	default:
		panic("unreachable")
	}
}

func (self *Analyser) analyseArray(node *ast.Array) *mean.Array {
	t := self.analyseArrayType(node.Type)
	elems := make([]mean.Expr, len(node.Elems))
	for i, en := range node.Elems {
		elems[i] = self.expectExpr(t.Elem, en)
	}
	return &mean.Array{
		Type:  t,
		Elems: elems,
	}
}

func (self *Analyser) analyseIndex(node *ast.Index) *mean.Index {
	from := self.analyseExpr(nil, node.From)
	if !stlbasic.Is[*mean.ArrayType](from.GetType()) {
		errors.ThrowNotArrayError(node.From.Position(), from.GetType())
	}
	index := self.expectExpr(mean.Usize, node.Index)
	return &mean.Index{
		From:  from,
		Index: index,
	}
}

func (self *Analyser) analyseExtract(expect mean.Type, node *ast.Extract) *mean.Extract {
	indexValue, ok := big.NewInt(0).SetString(node.Index.Source(), 10)
	if !ok {
		panic("unreachable")
	}
	if !indexValue.IsUint64() {
		panic("unreachable")
	}
	index := uint(indexValue.Uint64())

	expectFrom := &mean.TupleType{Elems: make([]mean.Type, index+1)}
	expectFrom.Elems[index] = expect

	from := self.analyseExpr(expectFrom, node.From)
	tt, ok := from.GetType().(*mean.TupleType)
	if !ok {
		errors.ThrowNotTupleError(node.From.Position(), from.GetType())
	}

	if index >= uint(len(tt.Elems)) {
		errors.ThrowInvalidIndexError(node.Index.Position, index)
	}
	return &mean.Extract{
		From:  from,
		Index: index,
	}
}

func (self *Analyser) analyseStruct(node *ast.Struct) *mean.Struct {
	st := self.analyseIdentType(node.Type).(*mean.StructType)
	fieldNames := hashset.NewHashSet[string]()
	for iter := st.Fields.Keys().Iterator(); iter.Next(); {
		fieldNames.Add(iter.Value())
	}

	existedFields := make(map[string]mean.Expr)
	for _, nf := range node.Fields {
		fn := nf.First.Source()
		if !fieldNames.Contain(fn) {
			errors.ThrowIdentifierDuplicationError(nf.First.Position, nf.First)
		}
		existedFields[fn] = self.expectExpr(st.Fields.Get(fn), nf.Second)
	}

	fields := make([]mean.Expr, st.Fields.Length())
	var i int
	for iter := st.Fields.Iterator(); iter.Next(); i++ {
		fn, ft := iter.Value().First, iter.Value().Second
		if fv, ok := existedFields[fn]; ok {
			fields[i] = fv
		} else {
			fields[i] = self.getTypeDefaultValue(node.Type.Position(), ft)
		}
	}

	return &mean.Struct{
		Type:   st,
		Fields: fields,
	}
}

func (self *Analyser) analyseField(node *ast.Field) mean.Expr {
	from := self.analyseExpr(nil, node.From)
	fieldName := node.Index.Source()
	st, ok := from.GetType().(*mean.StructType)
	if !ok {
		errors.ThrowNotStructError(node.From.Position(), from.GetType())
	}

	// 方法
	if method := st.Methods.Get(fieldName); method != nil{
		return &mean.Method{
			Self: from,
			Method: method,
		}
	}

	// 字段
	if !st.Fields.ContainKey(fieldName) {
		errors.ThrowUnknownIdentifierError(node.Index.Position, node.Index)
	}
	var i int
	for iter := st.Fields.Keys().Iterator(); iter.Next(); i++ {
		if iter.Value() == fieldName {
			break
		}
	}
	return &mean.Field{
		From:  from,
		Index: uint(i),
	}
}

func (self *Analyser) analyseString(node *ast.String) *mean.String {
	s := node.Value.Source()
	s = util.ParseEscapeCharacter(s[1:len(s)-1], `\"`, `"`)
	return &mean.String{Value: s}
}

func (self *Analyser) analyseJudgment(node *ast.Judgment) mean.Expr {
	target := self.analyseType(node.Type)
	value := self.analyseExpr(target, node.Value)
	vt := value.GetType()

	switch {
	case vt.Equal(target):
		return &mean.Boolean{Value: true}
	case stlbasic.Is[*mean.UnionType](vt) && vt.(*mean.UnionType).GetElemIndex(target) >= 0:
		return &mean.UnionTypeJudgment{
			Value: value,
			Type:  target,
		}
	default:
		return &mean.Boolean{Value: false}
	}
}

func (self *Analyser) analyseNull(expect mean.Type, node *ast.Null) *mean.Zero {
	if expect == nil || !(stlbasic.Is[*mean.PtrType](expect) || stlbasic.Is[*mean.RefType](expect)) {
		errors.ThrowExpectPointerTypeError(node.Position())
	}
	var t *mean.PtrType
	if stlbasic.Is[*mean.PtrType](expect) {
		t = expect.(*mean.PtrType)
	} else {
		t = expect.(*mean.RefType).ToPtrType()
	}
	return &mean.Zero{Type: t}
}

func (self *Analyser) analyseCheckNull(expect mean.Type, node *ast.CheckNull) *mean.CheckNull {
	if expect != nil {
		if pt, ok := expect.(*mean.PtrType); ok {
			expect = pt
		} else if rt, ok := expect.(*mean.RefType); ok {
			expect = rt.ToPtrType()
		}
	}
	value := self.analyseExpr(expect, node.Value)
	vt := value.GetType()
	if !stlbasic.Is[*mean.PtrType](vt) {
		errors.ThrowExpectPointerError(node.Value.Position(), vt)
	}
	return &mean.CheckNull{Value: value}
}

func (self *Analyser) analyseSelfValue(node *ast.SelfValue)*mean.Param{
	if self.selfValue == nil{
		errors.ThrowUnknownIdentifierError(node.Position(), node.Token)
	}
	return self.selfValue
}

func (self *Analyser) getTypeDefaultValue(pos reader.Position, t mean.Type)mean.Expr{
	if !t.HasDefault(){
		errors.ThrowCanNotGetDefault(pos, t)
	}
	return &mean.Zero{Type: t}
}
