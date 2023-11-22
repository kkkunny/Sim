package analyse

import (
	"math/big"

	stlbasic "github.com/kkkunny/stl/basic"
	"github.com/kkkunny/stl/container/hashset"
	stlerror "github.com/kkkunny/stl/error"
	"github.com/samber/lo"

	"github.com/kkkunny/Sim/ast"
	errors "github.com/kkkunny/Sim/error"
	. "github.com/kkkunny/Sim/mean"
	"github.com/kkkunny/Sim/token"
	"github.com/kkkunny/Sim/util"
)

func (self *Analyser) analyseExpr(expect Type, node ast.Expr) Expr {
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
	default:
		panic("unreachable")
	}
}

func (self *Analyser) analyseInteger(expect Type, node *ast.Integer) Expr {
	if expect == nil || !TypeIs[NumberType](expect) {
		expect = Isize
	}
	switch t := expect.(type) {
	case IntType:
		value, ok := big.NewInt(0).SetString(node.Value.Source(), 10)
		if !ok {
			panic("unreachable")
		}
		return &Integer{
			Type:  t,
			Value: *value,
		}
	case *FloatType:
		value, _ := stlerror.MustWith2(big.ParseFloat(node.Value.Source(), 10, big.MaxPrec, big.ToZero))
		return &Float{
			Type:  t,
			Value: *value,
		}
	default:
		panic("unreachable")
	}
}

func (self *Analyser) analyseChar(expect Type, node *ast.Char) Expr {
	if expect == nil || !TypeIs[NumberType](expect) {
		expect = I32
	}
	s := node.Value.Source()
	char := util.ParseEscapeCharacter(s[1:len(s)-1], `\'`, `'`)[0]
	switch t := expect.(type) {
	case IntType:
		value := big.NewInt(int64(char))
		return &Integer{
			Type:  t,
			Value: *value,
		}
	case *FloatType:
		value := big.NewFloat(float64(char))
		return &Float{
			Type:  t,
			Value: *value,
		}
	default:
		panic("unreachable")
	}
}

func (self *Analyser) analyseFloat(expect Type, node *ast.Float) *Float {
	if expect == nil || !TypeIs[*FloatType](expect) {
		expect = F64
	}
	value, _ := stlerror.MustWith2(big.NewFloat(0).Parse(node.Value.Source(), 10))
	return &Float{
		Type:  expect.(*FloatType),
		Value: *value,
	}
}

func (self *Analyser) analyseBool(node *ast.Boolean) *Boolean {
	return &Boolean{Value: node.Value.Is(token.TRUE)}
}

func (self *Analyser) analyseBinary(expect Type, node *ast.Binary) Binary {
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
			return &Assign{
				Left:  left,
				Right: right,
			}
		}
	case token.AND:
		if lt.Equal(rt) && TypeIs[IntType](lt) {
			return &IntAndInt{
				Left:  left,
				Right: right,
			}
		}
	case token.OR:
		if lt.Equal(rt) && TypeIs[IntType](lt) {
			return &IntOrInt{
				Left:  left,
				Right: right,
			}
		}
	case token.XOR:
		if lt.Equal(rt) && TypeIs[IntType](lt) {
			return &IntXorInt{
				Left:  left,
				Right: right,
			}
		}
	case token.SHL:
		if lt.Equal(rt) && TypeIs[IntType](lt) {
			return &IntShlInt{
				Left:  left,
				Right: right,
			}
		}
	case token.SHR:
		if lt.Equal(rt) && TypeIs[IntType](lt) {
			return &IntShrInt{
				Left:  left,
				Right: right,
			}
		}
	case token.ADD:
		if lt.Equal(rt) && TypeIs[NumberType](lt) {
			return &NumAddNum{
				Left:  left,
				Right: right,
			}
		}
	case token.SUB:
		if lt.Equal(rt) && TypeIs[NumberType](lt) {
			return &NumSubNum{
				Left:  left,
				Right: right,
			}
		}
	case token.MUL:
		if lt.Equal(rt) && TypeIs[NumberType](lt) {
			return &NumMulNum{
				Left:  left,
				Right: right,
			}
		}
	case token.DIV:
		if lt.Equal(rt) && TypeIs[NumberType](lt) {
			return &NumDivNum{
				Left:  left,
				Right: right,
			}
		}
	case token.REM:
		if lt.Equal(rt) && TypeIs[NumberType](lt) {
			return &NumRemNum{
				Left:  left,
				Right: right,
			}
		}
	case token.EQ:
		if lt.Equal(rt) {
			switch {
			case TypeIs[NumberType](lt):
				return &NumEqNum{
					Left:  left,
					Right: right,
				}
			case TypeIs[*BoolType](lt):
				return &BoolEqBool{
					Left:  left,
					Right: right,
				}
			case TypeIs[*FuncType](lt):
				return &FuncEqFunc{
					Left:  left,
					Right: right,
				}
			case TypeIs[*ArrayType](lt):
				return &ArrayEqArray{
					Left:  left,
					Right: right,
				}
			case TypeIs[*TupleType](lt):
				return &TupleEqTuple{
					Left:  left,
					Right: right,
				}
			case TypeIs[*StructType](lt):
				return &StructEqStruct{
					Left:  left,
					Right: right,
				}
			case TypeIs[*StringType](lt):
				return &StringEqString{
					Left:  left,
					Right: right,
				}
			case TypeIs[*UnionType](lt):
				return &UnionEqUnion{
					Left:  left,
					Right: right,
				}
			}
		}
	case token.NE:
		if lt.Equal(rt) {
			switch {
			case TypeIs[NumberType](lt):
				return &NumNeNum{
					Left:  left,
					Right: right,
				}
			case TypeIs[*BoolType](lt):
				return &BoolNeBool{
					Left:  left,
					Right: right,
				}
			case TypeIs[*FuncType](lt):
				return &FuncNeFunc{
					Left:  left,
					Right: right,
				}
			case TypeIs[*ArrayType](lt):
				return &ArrayNeArray{
					Left:  left,
					Right: right,
				}
			case TypeIs[*TupleType](lt):
				return &TupleNeTuple{
					Left:  left,
					Right: right,
				}
			case TypeIs[*StructType](lt):
				return &StructNeStruct{
					Left:  left,
					Right: right,
				}
			case TypeIs[*StringType](lt):
				return &StringNeString{
					Left:  left,
					Right: right,
				}
			case TypeIs[*UnionType](lt):
				return &UnionNeUnion{
					Left:  left,
					Right: right,
				}
			}
		}
	case token.LT:
		if lt.Equal(rt) && TypeIs[NumberType](lt) {
			return &NumLtNum{
				Left:  left,
				Right: right,
			}
		}
	case token.GT:
		if lt.Equal(rt) && TypeIs[NumberType](lt) {
			return &NumGtNum{
				Left:  left,
				Right: right,
			}
		}
	case token.LE:
		if lt.Equal(rt) && TypeIs[NumberType](lt) {
			return &NumLeNum{
				Left:  left,
				Right: right,
			}
		}
	case token.GE:
		if lt.Equal(rt) && TypeIs[NumberType](lt) {
			return &NumGeNum{
				Left:  left,
				Right: right,
			}
		}
	case token.LAND:
		if lt.Equal(rt) && TypeIs[*BoolType](lt) {
			return &BoolAndBool{
				Left:  left,
				Right: right,
			}
		}
	case token.LOR:
		if lt.Equal(rt) && TypeIs[*BoolType](lt) {
			return &BoolOrBool{
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

func (self *Analyser) analyseUnary(expect Type, node *ast.Unary) Unary {
	switch node.Opera.Kind {
	case token.SUB:
		value := self.analyseExpr(expect, node.Value)
		vt := value.GetType()
		if TypeIs[*SintType](vt) || TypeIs[*FloatType](vt) {
			return &NumNegate{Value: value}
		}
		errors.ThrowIllegalUnaryError(node.Position(), node.Opera, vt)
		return nil
	case token.NOT:
		value := self.analyseExpr(expect, node.Value)
		vt := value.GetType()
		switch {
		case TypeIs[IntType](vt):
			return &IntBitNegate{Value: value}
		case TypeIs[*BoolType](vt):
			return &BoolNegate{Value: value}
		default:
			errors.ThrowIllegalUnaryError(node.Position(), node.Opera, vt)
			return nil
		}
	case token.AND:
		if expect != nil && TypeIs[*PtrType](expect) {
			expect = expect.(*PtrType).Elem
		}
		value := self.analyseExpr(expect, node.Value)
		if !stlbasic.Is[Ident](value) {
			errors.ThrowCanNotGetPointer(node.Value.Position())
		}
		return &GetPtr{Value: value}
	case token.MUL:
		if expect != nil {
			expect = &PtrType{Elem: expect}
		}
		return &GetValue{Value: self.analyseExpr(expect, node.Value)}
	default:
		panic("unreachable")
	}
}

func (self *Analyser) analyseIdent(node *ast.Ident) Ident {
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

func (self *Analyser) analyseCall(node *ast.Call) *Call {
	f := self.analyseExpr(nil, node.Func)
	ft, ok := f.GetType().(*FuncType)
	if !ok {
		errors.ThrowNotFunctionError(node.Func.Position(), f.GetType())
	} else if len(ft.Params) != len(node.Args) {
		errors.ThrowParameterNumberNotMatchError(node.Position(), uint(len(ft.Params)), uint(len(node.Args)))
	}
	args := lo.Map(node.Args, func(item ast.Expr, index int) Expr {
		return self.analyseExpr(ft.Params[index], item)
	})
	return &Call{
		Func: f,
		Args: args,
	}
}

func (self *Analyser) analyseTuple(expect Type, node *ast.Tuple) Expr {
	if len(node.Elems) == 1 && (expect == nil || !TypeIs[*TupleType](expect)) {
		return self.analyseExpr(expect, node.Elems[0])
	}

	elemExpects := make([]Type, len(node.Elems))
	if expect != nil {
		if tt, ok := expect.(*TupleType); ok {
			if len(tt.Elems) < len(node.Elems) {
				copy(elemExpects, tt.Elems)
			} else if len(tt.Elems) > len(node.Elems) {
				elemExpects = tt.Elems[:len(node.Elems)]
			} else {
				elemExpects = tt.Elems
			}
		}
	}
	elems := lo.Map(node.Elems, func(item ast.Expr, index int) Expr {
		return self.analyseExpr(elemExpects[index], item)
	})
	return &Tuple{Elems: elems}
}

func (self *Analyser) analyseCovert(node *ast.Covert) Expr {
	tt := self.analyseType(node.Type)
	from := self.analyseExpr(tt, node.Value)
	ft := from.GetType()
	if ft.AssignableTo(tt) {
		return from
	}

	switch {
	case TypeIs[NumberType](ft) && TypeIs[NumberType](tt):
		return &Num2Num{
			From: from,
			To:   tt.(NumberType),
		}
	case TypeIs[*UnionType](tt) && tt.(*UnionType).GetElemIndex(ft) >= 0:
		return &Union{
			Type:  tt.(*UnionType),
			Value: from,
		}
	case TypeIs[*UnionType](ft) && ft.(*UnionType).GetElemIndex(tt) >= 0:
		return &UnUnion{
			Type:  tt,
			Value: from,
		}
	default:
		errors.ThrowIllegalCovertError(node.Position(), ft, tt)
		return nil
	}
}

func (self *Analyser) expectExpr(expect Type, node ast.Expr) Expr {
	value := self.analyseExpr(expect, node)
	if vt := value.GetType(); !vt.AssignableTo(expect) {
		errors.ThrowTypeMismatchError(node.Position(), vt, expect)
	}
	return self.autoTypeCovert(expect, value)
}

// 自动类型转换
func (self *Analyser) autoTypeCovert(expect Type, v Expr) Expr {
	vt := v.GetType()
	if vt.Equal(expect) {
		return v
	}

	switch {
	case TypeIs[*UnionType](expect):
		return &Union{
			Type:  expect.(*UnionType),
			Value: v,
		}
	default:
		panic("unreachable")
	}
}

func (self *Analyser) analyseArray(node *ast.Array) *Array {
	t := self.analyseArrayType(node.Type)
	elems := make([]Expr, len(node.Elems))
	for i, en := range node.Elems {
		elems[i] = self.expectExpr(t.Elem, en)
	}
	return &Array{
		Type:  t,
		Elems: elems,
	}
}

func (self *Analyser) analyseIndex(node *ast.Index) *Index {
	from := self.analyseExpr(nil, node.From)
	if !TypeIs[*ArrayType](from.GetType()) {
		errors.ThrowNotArrayError(node.From.Position(), from.GetType())
	}
	index := self.expectExpr(Usize, node.Index)
	return &Index{
		From:  from,
		Index: index,
	}
}

func (self *Analyser) analyseExtract(expect Type, node *ast.Extract) *Extract {
	indexValue, ok := big.NewInt(0).SetString(node.Index.Source(), 10)
	if !ok {
		panic("unreachable")
	}
	if !indexValue.IsUint64() {
		panic("unreachable")
	}
	index := uint(indexValue.Uint64())

	expectFrom := &TupleType{Elems: make([]Type, index+1)}
	expectFrom.Elems[index] = expect

	from := self.analyseExpr(expectFrom, node.From)
	tt, ok := from.GetType().(*TupleType)
	if !ok {
		errors.ThrowNotTupleError(node.From.Position(), from.GetType())
	}

	if index >= uint(len(tt.Elems)) {
		errors.ThrowInvalidIndexError(node.Index.Position, index)
	}
	return &Extract{
		From:  from,
		Index: index,
	}
}

func (self *Analyser) analyseStruct(node *ast.Struct) *Struct {
	st := self.analyseIdentType(node.Type).(*StructType)
	fieldNames := hashset.NewHashSet[string]()
	for iter := st.Fields.Keys().Iterator(); iter.Next(); {
		fieldNames.Add(iter.Value())
	}

	existedFields := make(map[string]Expr)
	for _, nf := range node.Fields {
		fn := nf.First.Source()
		if !fieldNames.Contain(fn) {
			errors.ThrowIdentifierDuplicationError(nf.First.Position, nf.First)
		}
		existedFields[fn] = self.expectExpr(st.Fields.Get(fn), nf.Second)
	}

	fields := make([]Expr, st.Fields.Length())
	var i int
	for iter := st.Fields.Iterator(); iter.Next(); i++ {
		fn, ft := iter.Value().First, iter.Value().Second
		if fv, ok := existedFields[fn]; ok {
			fields[i] = fv
		} else {
			fields[i] = &Zero{Type: ft}
		}
	}

	return &Struct{
		Type:   st,
		Fields: fields,
	}
}

func (self *Analyser) analyseField(node *ast.Field) *Field {
	from := self.analyseExpr(nil, node.From)
	fieldName := node.Index.Source()
	st, ok := from.GetType().(*StructType)
	if !ok {
		errors.ThrowNotStructError(node.From.Position(), from.GetType())
	} else if !st.Fields.ContainKey(fieldName) {
		errors.ThrowUnknownIdentifierError(node.Index.Position, node.Index)
	}
	var i int
	for iter := st.Fields.Keys().Iterator(); iter.Next(); i++ {
		if iter.Value() == fieldName {
			break
		}
	}
	return &Field{
		From:  from,
		Index: uint(i),
	}
}

func (self *Analyser) analyseString(node *ast.String) *String {
	s := node.Value.Source()
	s = util.ParseEscapeCharacter(s[1:len(s)-1], `\"`, `"`)
	return &String{Value: s}
}

func (self *Analyser) analyseJudgment(node *ast.Judgment) Expr {
	target := self.analyseType(node.Type)
	value := self.analyseExpr(target, node.Value)
	vt := value.GetType()

	switch {
	case vt.Equal(target):
		return &Boolean{Value: true}
	case TypeIs[*UnionType](vt) && vt.(*UnionType).GetElemIndex(target) >= 0:
		return &UnionTypeJudgment{
			Value: value,
			Type:  target,
		}
	default:
		return &Boolean{Value: false}
	}
}
