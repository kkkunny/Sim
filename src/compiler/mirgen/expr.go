package mirgen

import (
	"github.com/kkkunny/Sim/src/compiler/hir"
	"github.com/kkkunny/Sim/src/compiler/mir"
	"github.com/kkkunny/stl/util"
)

// 表达式
func (self *MirGenerator) genExpr(ir hir.Expr, getValue bool) mir.Value {
	switch expr := ir.(type) {
	case *hir.Integer, *hir.Float, *hir.Boolean, *hir.String, *hir.EmptyFunc, *hir.EmptyPtr, *hir.EmptyStruct, *hir.EmptyArray, *hir.EmptyTuple, *hir.EmptyEnum:
		return self.genConstantExpr(ir)
	case hir.Ident:
		switch ident := expr.(type) {
		case *hir.Param, *hir.Variable, *hir.GlobalValue:
			v := self.vars[ident]
			if getValue {
				return self.block.NewLoad(v)
			}
			return v
		case *hir.Function, *hir.Method:
			return self.vars[ident]
		default:
			panic("unreachable")
		}
	case hir.Binary:
		switch binary := expr.(type) {
		case *hir.Assign:
			l, r := self.genExpr(binary.Left, false), self.genExpr(binary.Right, true)
			self.block.NewStore(r, l)
			return nil
		case *hir.Equal:
			l, r := self.genExpr(binary.Left, true), self.genExpr(binary.Right, true)
			nb, v := self.block.NewEqual(l, r)
			self.block = nb
			return self.block.NewSelect(v, mir.NewInt(t_i8, 1), mir.NewInt(t_i8, 0))
		case *hir.NotEqual:
			l, r := self.genExpr(binary.Left, true), self.genExpr(binary.Right, true)
			nb, v := self.block.NewEqual(l, r)
			self.block = nb
			return self.block.NewSelect(v, mir.NewInt(t_i8, 0), mir.NewInt(t_i8, 1))
		case *hir.Lt:
			l, r := self.genExpr(binary.Left, true), self.genExpr(binary.Right, true)
			return self.block.NewSelect(self.block.NewLt(l, r), mir.NewInt(t_i8, 1), mir.NewInt(t_i8, 0))
		case *hir.Le:
			l, r := self.genExpr(binary.Left, true), self.genExpr(binary.Right, true)
			return self.block.NewSelect(self.block.NewLe(l, r), mir.NewInt(t_i8, 1), mir.NewInt(t_i8, 0))
		case *hir.Add:
			l, r := self.genExpr(binary.Left, true), self.genExpr(binary.Right, true)
			return self.block.NewAdd(l, r)
		case *hir.Sub:
			l, r := self.genExpr(binary.Left, true), self.genExpr(binary.Right, true)
			return self.block.NewSub(l, r)
		case *hir.Mul:
			l, r := self.genExpr(binary.Left, true), self.genExpr(binary.Right, true)
			return self.block.NewMul(l, r)
		case *hir.Div:
			l, r := self.genExpr(binary.Left, true), self.genExpr(binary.Right, true)
			return self.block.NewDiv(l, r)
		case *hir.Mod:
			l, r := self.genExpr(binary.Left, true), self.genExpr(binary.Right, true)
			return self.block.NewMod(l, r)
		case *hir.And:
			l, r := self.genExpr(binary.Left, true), self.genExpr(binary.Right, true)
			return self.block.NewAnd(l, r)
		case *hir.Or:
			l, r := self.genExpr(binary.Left, true), self.genExpr(binary.Right, true)
			return self.block.NewOr(l, r)
		case *hir.Xor:
			l, r := self.genExpr(binary.Left, true), self.genExpr(binary.Right, true)
			return self.block.NewXor(l, r)
		case *hir.Shl:
			l, r := self.genExpr(binary.Left, true), self.genExpr(binary.Right, true)
			return self.block.NewShl(l, r)
		case *hir.Shr:
			l, r := self.genExpr(binary.Left, true), self.genExpr(binary.Right, true)
			return self.block.NewShr(l, r)
		case *hir.LogicAnd:
			nb, eb := self.block.Belong.NewBlock(), self.block.Belong.NewBlock()

			left := self.genExpr(binary.Left, true)
			self.block.NewCondJmp(self.block.NewEq(left, mir.NewInt(left.GetType(), 1)), nb, eb)
			pb := self.block

			self.block = nb
			right := self.genExpr(binary.Right, true)
			self.block.NewJmp(eb)
			nb = self.block

			self.block = eb
			return self.block.NewPhi([]*mir.Block{pb, nb}, []mir.Value{mir.NewInt(t_i8, 0), right})
		case *hir.LogicOr:
			nb, eb := self.block.Belong.NewBlock(), self.block.Belong.NewBlock()

			left := self.genExpr(binary.Left, true)
			self.block.NewCondJmp(self.block.NewEq(left, mir.NewInt(left.GetType(), 1)), eb, nb)
			pb := self.block

			self.block = nb
			right := self.genExpr(binary.Right, true)
			self.block.NewJmp(eb)
			nb = self.block

			self.block = eb
			return self.block.NewPhi([]*mir.Block{pb, nb}, []mir.Value{mir.NewInt(t_i8, 1), right})
		default:
			panic("unreachable")
		}
	case hir.Call:
		switch call := expr.(type) {
		case *hir.FuncCall:
			f := self.genExpr(call.Func, true)
			args := make([]mir.Value, len(call.Args))
			for i, a := range call.Args {
				args[i] = self.genExpr(a, true)
			}
			return self.block.NewCall(f, args...)
		case *hir.MethodCall:
			f := self.vars[call.Method]
			args := make([]mir.Value, len(call.Args)+1)
			if call.Self.Type().IsPtr() {
				args[0] = self.genExpr(call.Self, true)
			} else if !call.Self.Immediate() {
				args[0] = self.genExpr(call.Self, false)
			} else {
				selfArg := self.genExpr(call.Self, true)
				args[0] = self.block.NewAlloc(self.genType(call.Self.Type()))
				self.block.NewStore(selfArg, args[0])
			}
			for i, a := range call.Args {
				args[i+1] = self.genExpr(a, true)
			}
			v := self.block.NewCall(f, args...)
			return v
		default:
			panic("unreachable")
		}
	case hir.Unary:
		switch unary := expr.(type) {
		case *hir.Not:
			v := self.genExpr(unary.Value, true)
			return self.block.NewXor(v, mir.NewInt(v.GetType(), 1))
		case *hir.GetPointer:
			return self.genExpr(unary.Value, false)
		case *hir.GetValue:
			value := self.genExpr(unary.Value, true)
			if getValue {
				return self.block.NewLoad(value)
			}
			return value
		default:
			panic("unreachable")
		}
	case hir.Index:
		switch index := expr.(type) {
		case *hir.ArrayIndex:
			from, value := self.genExpr(index.From, false), self.genExpr(index.Index, true)
			v := self.block.NewArrayIndex(from, value)
			if getValue && v.IsPtr() {
				return self.block.NewLoad(v)
			}
			return v
		case *hir.PointerIndex:
			from, value := self.genExpr(index.From, true), self.genExpr(index.Index, true)
			v := self.block.NewPointerIndex(from, value)
			if getValue {
				return self.block.NewLoad(v)
			}
			return v
		case *hir.TupleIndex:
			from := self.genExpr(index.From, false)
			v := self.block.NewStructIndex(from, index.Index)
			if getValue && v.IsPtr() {
				return self.block.NewLoad(v)
			}
			return v
		default:
			panic("unreachable")
		}
	case *hir.Ternary:
		_cond := self.genExpr(expr.Cond, true)
		cond := self.block.NewEq(_cond, mir.NewInt(_cond.GetType(), 1))
		tb, fb, eb := self.block.Belong.NewBlock(), self.block.Belong.NewBlock(), self.block.Belong.NewBlock()
		self.block.NewCondJmp(cond, tb, fb)

		self.block = tb
		tv := self.genExpr(expr.True, getValue)
		self.block.NewJmp(eb)

		self.block = fb
		fv := self.genExpr(expr.False, getValue)
		self.block.NewJmp(eb)

		self.block = eb
		return self.block.NewPhi([]*mir.Block{tb, fb}, []mir.Value{tv, fv})
	case *hir.Array:
		isConst := true
		elems := make([]mir.Value, len(expr.Elems))
		consts := make([]mir.Constant, len(expr.Elems))
		for i, e := range expr.Elems {
			elems[i] = self.genExpr(e, true)
			if c, ok := elems[i].(mir.Constant); !ok {
				isConst = false
			} else {
				consts[i] = c
			}
		}
		if isConst {
			return mir.NewArray(consts...)
		} else {
			tmp := self.block.NewAlloc(self.genType(expr.Type()))
			for i, e := range elems {
				index := self.block.NewArrayIndex(tmp, mir.NewUint(t_usize, uint64(i)))
				self.block.NewStore(e, index)
			}
			v := self.block.NewLoad(tmp)
			return v
		}
	case *hir.Tuple:
		isConst := true
		elems := make([]mir.Value, len(expr.Elems))
		consts := make([]mir.Constant, len(expr.Elems))
		for i, e := range expr.Elems {
			elems[i] = self.genExpr(e, true)
			if c, ok := elems[i].(mir.Constant); !ok {
				isConst = false
			} else {
				consts[i] = c
			}
		}
		if isConst {
			return mir.NewStruct(consts...)
		} else {
			tmp := self.block.NewAlloc(self.genType(expr.Type()))
			for i, e := range elems {
				index := self.block.NewStructIndex(tmp, uint(i))
				self.block.NewStore(e, index)
			}
			v := self.block.NewLoad(tmp)
			return v
		}
	case *hir.Struct:
		isConst := true
		elems := make([]mir.Value, len(expr.Fields))
		consts := make([]mir.Constant, len(expr.Fields))
		for i, e := range expr.Fields {
			elems[i] = self.genExpr(e, true)
			if c, ok := elems[i].(mir.Constant); !ok {
				isConst = false
			} else {
				consts[i] = c
			}
		}
		if isConst {
			return mir.NewStruct(consts...)
		} else {
			tmp := self.block.NewAlloc(self.genType(expr.Type()))
			for i, e := range elems {
				index := self.block.NewStructIndex(tmp, uint(i))
				self.block.NewStore(e, index)
			}
			v := self.block.NewLoad(tmp)
			return v
		}
	case *hir.Enum:
		index := expr.GetFieldIndex()
		tmp := self.block.NewAlloc(self.genType(expr.Type()))
		self.block.NewStore(mir.NewUint(t_usize, uint64(index)), self.block.NewStructIndex(tmp, 0))
		elemTypeHir := expr.Type().GetEnumFields()[index].Third
		if elemTypeHir != nil {
			ptr := self.block.NewStructIndex(tmp, 1)
			v := self.block.NewWrapUnion(ptr.GetType().GetPtr(), self.genExpr(expr.Value, true))
			self.block.NewStore(v, ptr)
		}
		return self.block.NewLoad(tmp)
	case *hir.GetStructField:
		from := self.genExpr(expr.From, false)
		v := self.block.NewStructIndex(from, expr.GetFieldIndex())
		if getValue && v.IsPtr() {
			return self.block.NewLoad(v)
		}
		return v
	case *hir.GetEnumField:
		f := self.genExpr(expr.From, false)
		v := self.block.NewUnwrapUnion(self.genType(expr.Type()), self.block.NewStructIndex(f, 1))
		if getValue && v.IsPtr() {
			return self.block.NewLoad(v)
		}
		return v
	case hir.Covert:
		switch covert := expr.(type) {
		case *hir.WrapCovert:
			return self.genExpr(covert.From, true)
		case *hir.Int2Int, *hir.Float2Float, *hir.Int2Float, *hir.Float2Int:
			from := self.genExpr(covert.GetFrom(), true)
			return self.block.NewNumber2Number(from, self.genType(covert.Type()))
		case *hir.Usize2Ptr:
			from := self.genExpr(covert.From, true)
			return self.block.NewUint2Ptr(from, self.genType(covert.Type()))
		case *hir.Ptr2Usize:
			from := self.genExpr(covert.From, true)
			return self.block.NewPtr2Uint(from, self.genType(covert.Type()))
		case *hir.Ptr2Ptr:
			from := self.genExpr(covert.From, true)
			return self.block.NewPtr2Ptr(from, self.genType(covert.Type()))
		default:
			panic("unreachable")
		}
	case *hir.Size:
		targetType, resType := self.genType(expr.Typ), self.genType(expr.Type())
		return mir.NewInt(resType, int64(targetType.Size()))
	case *hir.Align:
		targetType, resType := self.genType(expr.Typ), self.genType(expr.Type())
		return mir.NewInt(resType, int64(targetType.Align()))
	default:
		panic("unreachable")
	}
}

// 常量表达式
func (self *MirGenerator) genConstantExpr(ir hir.Expr) mir.Constant {
	switch expr := ir.(type) {
	case *hir.Integer:
		return mir.NewInt(self.genType(expr.Typ), expr.Value)
	case *hir.Float:
		return mir.NewFloat(self.genType(expr.Typ), expr.Value)
	case *hir.Boolean:
		return util.Ternary(expr.Value, mir.NewSint(t_bool, 1), mir.NewSint(t_bool, 0))
	case *hir.EmptyFunc:
		return mir.NewEmptyFunc(self.genType(expr.Typ))
	case *hir.EmptyPtr:
		return mir.NewEmptyPtr(self.genType(expr.Typ))
	case *hir.EmptyArray:
		return mir.NewEmptyArray(self.genType(expr.Typ))
	case *hir.EmptyTuple, *hir.EmptyStruct, *hir.EmptyEnum:
		return mir.NewEmptyStruct(self.genType(expr.Type()))
	case *hir.Array:
		elems := make([]mir.Constant, len(expr.Elems))
		for i, e := range expr.Elems {
			elems[i] = self.genConstantExpr(e)
		}
		return mir.NewArray(elems...)
	case *hir.Tuple:
		elems := make([]mir.Constant, len(expr.Elems))
		for i, e := range expr.Elems {
			elems[i] = self.genConstantExpr(e)
		}
		return mir.NewStruct(elems...)
	case *hir.Struct:
		elems := make([]mir.Constant, len(expr.Fields))
		for i, e := range expr.Fields {
			elems[i] = self.genConstantExpr(e)
		}
		return mir.NewStruct(elems...)
	case *hir.String:
		v, ok := self.stringPool[expr.Value]
		if !ok {
			elems := make([]mir.Constant, len(expr.Value)+1)
			for i, e := range []byte(expr.Value) {
				elems[i] = mir.NewSint(t_i8, int64(e))
			}
			elems[len(elems)-1] = mir.NewSint(t_i8, 0)
			init := mir.NewArray(elems...)

			v = self.pkg.NewVariable(init.GetType(), "", init)
			self.stringPool[expr.Value] = v
		}
		return mir.NewArrayIndexConst(v, 0)
	default:
		panic("unreachable")
	}
}
