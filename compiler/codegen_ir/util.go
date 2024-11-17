package codegen_ir

import (
	"fmt"

	"github.com/kkkunny/go-llvm"
	"github.com/kkkunny/stl/container/hashmap"
	"github.com/kkkunny/stl/container/queue"
	stlslices "github.com/kkkunny/stl/container/slices"
	"github.com/kkkunny/stl/container/tuple"
	stlerr "github.com/kkkunny/stl/error"
	stlmath "github.com/kkkunny/stl/math"
	stlos "github.com/kkkunny/stl/os"
	stlval "github.com/kkkunny/stl/value"

	"github.com/heimdalr/dag"

	"github.com/kkkunny/Sim/compiler/analyse"
	"github.com/kkkunny/Sim/compiler/hir"
	"github.com/kkkunny/Sim/compiler/hir/global"
	"github.com/kkkunny/Sim/compiler/hir/local"
	"github.com/kkkunny/Sim/compiler/hir/types"
	"github.com/kkkunny/Sim/compiler/hir/values"
)

func (self *CodeGenerator) getIdentName(ir values.Ident) string {
	switch ident := ir.(type) {
	case *global.FuncDef:
		name := stlslices.First(stlslices.FlatMap(ident.Attrs(), func(_ int, attr global.FuncAttr) []string {
			nameAttr, ok := attr.(*global.FuncAttrLinkName)
			if !ok {
				return nil
			}
			return []string{nameAttr.Name()}
		}))
		if name != "" {
			return name
		}
		return fmt.Sprintf("%s::%s", ident.Package().String(), stlval.IgnoreWith(ident.GetName()))
	case *global.MethodDef:
		name := stlslices.First(stlslices.FlatMap(ident.Attrs(), func(_ int, attr global.FuncAttr) []string {
			nameAttr, ok := attr.(*global.FuncAttrLinkName)
			if !ok {
				return nil
			}
			return []string{nameAttr.Name()}
		}))
		if name != "" {
			return name
		}
		return fmt.Sprintf("%s::%s::%s", ident.Package().String(), stlval.IgnoreWith(ident.From().GetName()), stlval.IgnoreWith(ident.GetName()))
	case *global.VarDef:
		name := stlslices.First(stlslices.FlatMap(ident.Attrs(), func(_ int, attr global.VarAttr) []string {
			nameAttr, ok := attr.(*global.VarAttrLinkName)
			if !ok {
				return nil
			}
			return []string{nameAttr.Name()}
		}))
		if name != "" {
			return name
		}
		return fmt.Sprintf("%s::%s", ident.Package().String(), stlval.IgnoreWith(ident.GetName()))
	case *local.SingleVarDef, values.VarDecl, *local.Param:
		return ""
	default:
		panic("unreachable")
	}
}

func (self *CodeGenerator) typeIsStruct(t hir.Type) bool {
	switch t := t.(type) {
	case types.NoThingType, types.NoReturnType, types.NumType, types.BoolType, types.RefType, types.ArrayType, types.FuncType:
		return false
	case types.StrType, types.TupleType, types.LambdaType, types.StructType:
		return true
	case types.EnumType:
		return !t.Simple()
	case types.TypeDef:
		return self.typeIsStruct(t.Target())
	default:
		panic("unreachable")
	}
}

func (self *CodeGenerator) buildCopy(tIr hir.Type, v llvm.Value) llvm.Value {
	switch tIr := tIr.(type) {
	case types.NoThingType, types.NoReturnType:
		panic("unreachable")
	case types.RefType, types.CallableType, types.NumType:
		return v
	case types.ArrayType:
		t := self.codegenType(tIr).(llvm.ArrayType)
		if t.Capacity() == 0 {
			return v
		}

		newArrayPtr := self.builder.CreateAlloca("", t)
		indexPtr := self.builder.CreateAlloca("", self.builder.Isize())
		self.builder.CreateStore(self.builder.ConstZero(self.builder.Isize()), indexPtr)
		condBlock := self.builder.CurrentFunction().NewBlock("")
		self.builder.CreateBr(condBlock)

		// cond
		self.builder.MoveToAfter(condBlock)
		index := self.builder.CreateLoad("", self.builder.Isize(), indexPtr)
		var cond llvm.Value = self.builder.CreateIntCmp("", llvm.IntULT, index, self.builder.ConstIsize(int64(t.Capacity())))
		bodyBlock, outBlock := self.builder.CurrentFunction().NewBlock(""), self.builder.CurrentFunction().NewBlock("")
		self.builder.CreateCondBr(cond, bodyBlock, outBlock)

		// body
		self.builder.MoveToAfter(bodyBlock)
		self.builder.CreateStore(self.buildCopy(tIr.Elem(), self.builder.CreateArrayIndex(t, v, index, false)), self.builder.CreateArrayIndex(t, newArrayPtr, index, true))
		actionBlock := self.builder.CurrentFunction().NewBlock("")
		self.builder.CreateBr(actionBlock)

		// action
		self.builder.MoveToAfter(actionBlock)
		self.builder.CreateStore(self.builder.CreateUAdd("", index, self.builder.ConstIsize(1)), indexPtr)
		self.builder.CreateBr(condBlock)

		// out
		self.builder.MoveToAfter(outBlock)
		return self.builder.CreateLoad("", t, newArrayPtr)
	case types.TupleType, types.StructType:
		st := self.codegenType(tIr).(llvm.StructType)
		if len(st.Elems()) == 0 {
			return v
		}

		var fields []hir.Type
		if ttIr, ok := types.As[types.TupleType](tIr); ok {
			fields = ttIr.Elems()
		} else {
			fields = stlslices.Map(stlval.IgnoreWith(types.As[types.StructType](tIr)).Fields().Values(), func(_ int, field *types.Field) hir.Type {
				return field.Type()
			})
		}

		newStructPtr := self.builder.CreateAlloca("", st)
		for i := range st.Elems() {
			self.builder.CreateStore(self.buildCopy(fields[i], self.builder.CreateStructIndex(st, v, uint(i), false)), self.builder.CreateStructIndex(st, newStructPtr, uint(i), true))
		}
		return self.builder.CreateLoad("", st, newStructPtr)
	case types.EnumType:
		if tIr.Simple() {
			return v
		}

		t := self.codegenType(tIr).(llvm.StructType)
		beginBlock := self.builder.CurrentBlock()
		index := self.builder.CreateStructIndex(t, v, 1, false)

		values := make([]llvm.Value, tIr.EnumFields().Length())
		indexBlockPairs := stlslices.Map(tIr.EnumFields().Values(), func(i int, field *types.EnumField) struct {
			Value llvm.Value
			Block llvm.Block
		} {
			if _, ok := field.Elem(); !ok {
				return struct {
					Value llvm.Value
					Block llvm.Block
				}{Value: self.builder.ConstInteger(index.Type().(llvm.IntegerType), int64(i)), Block: beginBlock}
			}

			filedBlock := beginBlock.Belong().NewBlock("")
			self.builder.MoveToAfter(filedBlock)

			data := self.builder.CreateStructIndex(t, v, 0, false)
			ptr := self.builder.CreateAlloca("", t)
			self.builder.CreateStore(self.buildCopy(stlval.IgnoreWith(field.Elem()), data), self.builder.CreateStructIndex(t, ptr, 0, true))
			self.builder.CreateStore(index, self.builder.CreateStructIndex(t, ptr, 1, true))
			values[i] = self.builder.CreateLoad("", t, ptr)
			return struct {
				Value llvm.Value
				Block llvm.Block
			}{Value: self.builder.ConstInteger(index.Type().(llvm.IntegerType), int64(i)), Block: filedBlock}
		})

		endBlock := beginBlock.Belong().NewBlock("")
		for i, p := range indexBlockPairs {
			if p.Block != beginBlock {
				continue
			}
			p.Block = endBlock
			indexBlockPairs[i] = p
		}
		self.builder.MoveToAfter(beginBlock)
		self.builder.CreateSwitch(index, endBlock, indexBlockPairs...)
		self.builder.MoveToAfter(endBlock)

		phi := self.builder.CreatePHI("", v.Type())
		phi.AddIncomings(struct {
			Value llvm.Value
			Block llvm.Block
		}{Value: v, Block: beginBlock})
		for i, p := range indexBlockPairs {
			if p.Block == endBlock {
				continue
			}
			self.builder.MoveToAfter(p.Block)
			self.builder.CreateBr(endBlock)
			phi.AddIncomings(struct {
				Value llvm.Value
				Block llvm.Block
			}{Value: values[i], Block: p.Block})
			self.builder.MoveToAfter(endBlock)
		}
		return phi
	case types.TypeDef:
		return self.buildCopy(tIr.Target(), v)
	default:
		panic("unreachable")
	}
}

func (self *CodeGenerator) buildCheckZero(v llvm.Value) {
	var cond llvm.Value
	if stlval.Is[llvm.FloatType](v.Type()) {
		cond = self.builder.CreateFloatCmp("", llvm.FloatOEQ, v, self.builder.ConstZero(v.Type()))
	} else {
		cond = self.builder.CreateIntCmp("", llvm.IntEQ, v, self.builder.ConstZero(v.Type()))
	}
	f := self.builder.CurrentFunction()
	panicBlock, endBlock := f.NewBlock(""), f.NewBlock("")
	self.builder.CreateCondBr(cond, panicBlock, endBlock)

	self.builder.MoveToAfter(panicBlock)
	self.buildPanic("zero exception")

	self.builder.MoveToAfter(endBlock)
}

func (self *CodeGenerator) buildPanic(s string) {
	fn := self.builder.GetExternFunction("sim_runtime_panic", self.builder.FunctionType(false, self.builder.VoidType(), self.builder.OpaquePointerType()))
	self.builder.CreateCall("", fn.FunctionType(), fn, self.builder.ConstString(s))
	self.builder.CreateUnreachable()
}

func (self *CodeGenerator) buildCheckIndex(index llvm.Value, rangev uint64) {
	cond := self.builder.CreateIntCmp("", llvm.IntUGE, index, self.builder.ConstInteger(index.Type().(llvm.IntegerType), int64(rangev)))
	f := self.builder.CurrentFunction()
	panicBlock, endBlock := f.NewBlock(""), f.NewBlock("")
	self.builder.CreateCondBr(cond, panicBlock, endBlock)

	self.builder.MoveToAfter(panicBlock)
	self.buildPanic("index out of range")

	self.builder.MoveToAfter(endBlock)
}

func (self *CodeGenerator) buildMalloc(t llvm.Type) llvm.Value {
	fn := self.builder.GetExternFunction("sim_runtime_malloc", self.builder.FunctionType(false, self.builder.OpaquePointerType(), self.builder.Isize()))
	size := stlmath.RoundTo(self.builder.GetStoreSizeOfType(t), self.builder.GetABIAlignOfType(t))
	return self.builder.CreateCall("", fn.FunctionType(), fn, self.builder.ConstIsize(int64(size)))
}

func (self *CodeGenerator) buildEqual(tIr hir.Type, l, r llvm.Value, not bool) llvm.Value {
	switch tIr := tIr.(type) {
	case types.IntType, types.RefType, types.BoolType:
		return self.builder.CreateIntCmp("", stlval.Ternary(!not, llvm.IntEQ, llvm.IntNE), l, r)
	case types.FloatType:
		return self.builder.CreateFloatCmp("", stlval.Ternary(!not, llvm.FloatOEQ, llvm.FloatUNE), l, r)
	case types.ArrayType:
		t := self.codegenType(tIr).(llvm.ArrayType)
		if t.Capacity() == 0 {
			return self.builder.ConstBoolean(true)
		}

		indexPtr := self.builder.CreateAlloca("", self.builder.Isize())
		self.builder.CreateStore(self.builder.ConstZero(self.builder.Isize()), indexPtr)
		condBlock := self.builder.CurrentFunction().NewBlock("")
		self.builder.CreateBr(condBlock)

		// cond
		self.builder.MoveToAfter(condBlock)
		index := self.builder.CreateLoad("", self.builder.Isize(), indexPtr)
		var cond llvm.Value = self.builder.CreateIntCmp("", llvm.IntULT, index, self.builder.ConstIsize(int64(t.Capacity())))
		bodyBlock, outBlock := self.builder.CurrentFunction().NewBlock(""), self.builder.CurrentFunction().NewBlock("")
		self.builder.CreateCondBr(cond, bodyBlock, outBlock)

		// body
		self.builder.MoveToAfter(bodyBlock)
		cond = self.buildEqual(tIr.Elem(), self.builder.CreateArrayIndex(t, l, index, false), self.builder.CreateArrayIndex(t, r, index, false), false)
		bodyEndBlock := self.builder.CurrentBlock()
		actionBlock := bodyEndBlock.Belong().NewBlock("")
		self.builder.CreateCondBr(cond, actionBlock, outBlock)

		// action
		self.builder.MoveToAfter(actionBlock)
		self.builder.CreateStore(self.builder.CreateUAdd("", index, self.builder.ConstIsize(1)), indexPtr)
		self.builder.CreateBr(condBlock)

		// out
		self.builder.MoveToAfter(outBlock)
		phi := self.builder.CreatePHI(
			"",
			self.builder.BooleanType(),
			struct {
				Value llvm.Value
				Block llvm.Block
			}{Value: self.builder.ConstBoolean(true), Block: condBlock},
			struct {
				Value llvm.Value
				Block llvm.Block
			}{Value: self.builder.ConstBoolean(false), Block: bodyEndBlock},
		)
		if not {
			return self.builder.CreateNot("", phi)
		}
		return phi
	case types.TupleType, types.StructType:
		st := self.codegenType(tIr).(llvm.StructType)
		if len(st.Elems()) == 0 {
			return self.builder.ConstBoolean(true)
		}

		var fields []hir.Type
		if ttIr, ok := types.As[types.TupleType](tIr); ok {
			fields = ttIr.Elems()
		} else {
			fields = stlslices.Map(stlval.IgnoreWith(types.As[types.StructType](tIr)).Fields().Values(), func(_ int, field *types.Field) hir.Type {
				return field.Type()
			})
		}

		beginBlock := self.builder.CurrentBlock()
		nextBlocks := make([]tuple.Tuple2[llvm.Value, llvm.Block], uint(len(st.Elems())))
		srcBlocks := make([]llvm.Block, uint(len(st.Elems())))
		for i := uint(0); i < uint(len(st.Elems())); i++ {
			cond := self.buildEqual(fields[i], self.builder.CreateStructIndex(st, l, i, false), self.builder.CreateStructIndex(st, r, i, false), false)
			nextBlock := beginBlock.Belong().NewBlock("")
			nextBlocks[i] = tuple.Pack2(cond, nextBlock)
			srcBlocks[i] = self.builder.CurrentBlock()
			self.builder.MoveToAfter(nextBlock)
		}
		self.builder.MoveToAfter(beginBlock)

		endBlock := stlslices.Last(nextBlocks).E2()
		for i, p := range nextBlocks {
			if i != len(nextBlocks)-1 {
				self.builder.CreateCondBr(p.E1(), p.E2(), endBlock)
			} else {
				self.builder.CreateBr(endBlock)
			}
			self.builder.MoveToAfter(p.E2())
		}

		phi := self.builder.CreatePHI("", self.builder.BooleanType())
		for i, b := range srcBlocks {
			phi.AddIncomings(struct {
				Value llvm.Value
				Block llvm.Block
			}{Value: stlval.Ternary[llvm.Value](i != len(srcBlocks)-1, self.builder.ConstBoolean(false), stlslices.Last(nextBlocks).E1()), Block: b})
		}
		if not {
			return self.builder.CreateNot("", phi)
		}
		return phi
	case types.LambdaType:
		st := self.codegenType(tIr).(llvm.StructType)
		f1p := self.builder.CreateIntCmp("", stlval.Ternary(!not, llvm.IntEQ, llvm.IntNE), self.builder.CreateStructIndex(st, l, 0, false), self.builder.CreateStructIndex(st, r, 0, false))
		f2p := self.builder.CreateIntCmp("", stlval.Ternary(!not, llvm.IntEQ, llvm.IntNE), self.builder.CreateStructIndex(st, l, 1, false), self.builder.CreateStructIndex(st, r, 1, false))
		pp := self.builder.CreateIntCmp("", stlval.Ternary(!not, llvm.IntEQ, llvm.IntNE), self.builder.CreateStructIndex(st, l, 2, false), self.builder.CreateStructIndex(st, r, 2, false))
		return stlval.TernaryAction(!not, func() llvm.Value {
			return self.builder.CreateAnd("", self.builder.CreateAnd("", f1p, f2p), pp)
		}, func() llvm.Value {
			return self.builder.CreateOr("", self.builder.CreateOr("", f1p, f2p), pp)
		})
	case types.EnumType:
		if tIr.Simple() {
			return self.builder.CreateIntCmp("", stlval.Ternary(!not, llvm.IntEQ, llvm.IntNE), l, r)
		}

		t := self.codegenType(tIr).(llvm.StructType)
		beginBlock := self.builder.CurrentBlock()
		li, ri := self.builder.CreateStructIndex(t, l, 1, false), self.builder.CreateStructIndex(t, r, 1, false)
		indexEq := self.builder.CreateIntCmp("", llvm.IntEQ, li, ri)
		bodyBlock := beginBlock.Belong().NewBlock("")
		self.builder.MoveToAfter(bodyBlock)

		values := make([]llvm.Value, tIr.EnumFields().Length())
		indexBlockPairs := stlslices.Map(tIr.EnumFields().Values(), func(i int, field *types.EnumField) struct {
			Value llvm.Value
			Block llvm.Block
		} {
			if _, ok := field.Elem(); !ok {
				return struct {
					Value llvm.Value
					Block llvm.Block
				}{Value: self.builder.ConstInteger(li.Type().(llvm.IntegerType), int64(i)), Block: bodyBlock}
			}

			filedBlock := beginBlock.Belong().NewBlock("")
			self.builder.MoveToAfter(filedBlock)

			ldp, rdp := self.builder.CreateStructIndex(t, l, 0, true), self.builder.CreateStructIndex(t, r, 0, true)
			ftIr := stlval.IgnoreWith(field.Elem())
			ft := self.codegenType(ftIr)
			values[i] = self.buildEqual(ftIr, self.builder.CreateLoad("", ft, ldp), self.builder.CreateLoad("", ft, rdp), false)
			return struct {
				Value llvm.Value
				Block llvm.Block
			}{Value: self.builder.ConstInteger(li.Type().(llvm.IntegerType), int64(i)), Block: filedBlock}
		})

		endBlock := beginBlock.Belong().NewBlock("")
		for i, p := range indexBlockPairs {
			if p.Block != bodyBlock {
				continue
			}
			p.Block = endBlock
			indexBlockPairs[i] = p
		}
		self.builder.MoveToAfter(beginBlock)
		self.builder.CreateCondBr(indexEq, bodyBlock, endBlock)

		self.builder.MoveToAfter(bodyBlock)
		self.builder.CreateSwitch(li, endBlock, indexBlockPairs...)
		self.builder.MoveToAfter(endBlock)

		phi := self.builder.CreatePHI("", self.builder.BooleanType())
		phi.AddIncomings(struct {
			Value llvm.Value
			Block llvm.Block
		}{Value: self.builder.ConstBoolean(false), Block: beginBlock})
		phi.AddIncomings(struct {
			Value llvm.Value
			Block llvm.Block
		}{Value: self.builder.ConstBoolean(true), Block: bodyBlock})
		for i, p := range indexBlockPairs {
			if p.Block == endBlock {
				continue
			}
			self.builder.MoveToAfter(p.Block)
			self.builder.CreateBr(endBlock)
			phi.AddIncomings(struct {
				Value llvm.Value
				Block llvm.Block
			}{Value: values[i], Block: p.Block})
			self.builder.MoveToAfter(endBlock)
		}
		if not {
			return self.builder.CreateNot("", phi)
		}
		return phi
	case types.CustomType:
		return self.buildEqual(tIr.Target(), l, r, not)
	case types.AliasType:
		return self.buildEqual(tIr.Target(), l, r, not)
	default:
		panic("unreachable")
	}
}

func (self *CodeGenerator) codegenDefault(ir hir.Type) llvm.Value {
	switch ir := ir.(type) {
	case types.RefType:
		if ir.Pointer().Equal(types.Str) {
			return self.builder.ConstString("")
		}
		panic("unreachable")
	case types.NumType:
		return self.builder.ConstZero(self.codegenType(ir))
	case types.BoolType:
		return self.builder.ConstBoolean(false)
	case types.ArrayType:
		at := self.codegenArrayType(ir)
		if ir.Size() == 0 {
			return self.builder.ConstZero(at)
		}

		key := fmt.Sprintf("default:%s", ir.String())
		var fn llvm.Function
		if !self.funcCache.Contain(key) {
			curBlock := self.builder.CurrentBlock()
			ft := self.builder.FunctionType(false, at)
			fn = self.builder.NewFunction("", ft)
			self.funcCache.Set(key, fn)
			self.builder.MoveToAfter(fn.NewBlock(""))

			arrayPtr := self.builder.CreateAlloca("", at)
			indexPtr := self.builder.CreateAlloca("", self.builder.Isize())
			self.builder.CreateStore(self.builder.ConstZero(self.builder.Isize()), indexPtr)
			condBlock := fn.NewBlock("")
			self.builder.CreateBr(condBlock)

			self.builder.MoveToAfter(condBlock)
			index := self.builder.CreateLoad("", self.builder.Isize(), indexPtr)
			cond := self.builder.CreateIntCmp("", llvm.IntULT, index, self.builder.ConstIsize(int64(ir.Size())))
			loopBlock, endBlock := fn.NewBlock(""), fn.NewBlock("")
			self.builder.CreateCondBr(cond, loopBlock, endBlock)

			self.builder.MoveToAfter(loopBlock)
			elemPtr := self.builder.CreateArrayIndex(at, arrayPtr, index, true)
			self.builder.CreateStore(self.codegenDefault(ir.Elem()), elemPtr)
			self.builder.CreateStore(self.builder.CreateUAdd("", index, self.builder.ConstIsize(1)), indexPtr)
			self.builder.CreateBr(condBlock)

			self.builder.MoveToAfter(endBlock)
			self.builder.CreateRet(stlval.Ptr[llvm.Value](self.builder.CreateLoad("", at, arrayPtr)))

			self.builder.MoveToAfter(curBlock)
		} else {
			fn = self.funcCache.Get(key)
		}
		return self.builder.CreateCall("", fn.FunctionType(), fn)
	case types.TupleType:
		elems := stlslices.Map(ir.Elems(), func(_ int, elem hir.Type) llvm.Value {
			return self.codegenDefault(elem)
		})
		return self.builder.CreateStruct(self.codegenTupleType(ir), elems...)
	case types.StructType:
		fields := stlslices.Map(ir.Fields().Values(), func(_ int, field *types.Field) llvm.Value {
			return self.codegenDefault(field.Type())
		})
		return self.builder.CreateStruct(self.codegenStructType(ir), fields...)
	case types.FuncType:
		ft := self.codegenFuncType(ir)
		key := fmt.Sprintf("default:%s", ir.String())
		var fn llvm.Function
		if !self.funcCache.Contain(key) {
			curBlock := self.builder.CurrentBlock()
			fn = self.builder.NewFunction("", ft)
			self.funcCache.Set(key, fn)
			self.builder.MoveToAfter(fn.NewBlock(""))
			if ft.ReturnType().Equal(self.builder.VoidType()) {
				self.builder.CreateRet(nil)
			} else {
				self.builder.CreateRet(stlval.Ptr(self.buildCopy(ir.Ret(), self.codegenDefault(ir.Ret()))))
			}
			self.builder.MoveToAfter(curBlock)
		} else {
			fn = self.funcCache.Get(key)
		}
		return fn
	case types.LambdaType:
		lts := self.codegenLambdaType()
		fn := self.codegenDefault(ir.ToFunc())
		return self.builder.CreateStruct(lts, fn, self.builder.ConstZero(lts.Elems()[1]))
	case types.EnumType:
		if ir.Simple() {
			return self.builder.ConstInteger(self.codegenType(ir).(llvm.IntegerType), 0)
		}

		f := func(elem hir.Type, hasElem bool) (v llvm.Value, ok bool) {
			defer func() {
				if err := recover(); err != nil {
					ok = false
				}
			}()
			if !hasElem {
				return nil, true
			}
			return self.codegenDefault(elem), true
		}
		var index int
		var data llvm.Value
		var ok bool
		for i, elem := range ir.EnumFields().Values() {
			data, ok = f(elem.Elem())
			if ok {
				index = i
				break
			}
		}
		if !ok {
			panic("unreachable")
		}

		ut := self.codegenType(ir).(llvm.StructType)
		ptr := self.builder.CreateAlloca("", ut)
		if data != nil {
			self.builder.CreateStore(data, self.builder.CreateStructIndex(ut, ptr, 0, true))
		}
		self.builder.CreateStore(
			self.builder.ConstInteger(ut.GetElem(1).(llvm.IntegerType), int64(index)),
			self.builder.CreateStructIndex(ut, ptr, 1, true),
		)
		return self.builder.CreateLoad("", ut, ptr)
	case types.TypeDef:
		return self.codegenDefault(ir.Target())
	default:
		panic("unreachable")
	}
}

// CodegenIr 中间代码生成
func CodegenIr(target llvm.Target, path stlos.FilePath) (llvm.Module, error) {
	entryPkg, err := analyse.Analyse(path)
	if err != nil {
		return llvm.Module{}, err
	}

	var buildinPkgTaskID string
	existPkgs := hashmap.StdWith[*hir.Package, string]()
	pkgs := queue.New[*hir.Package](entryPkg)
	dager := dag.NewDAG()
	for !pkgs.Empty() {
		pkg := pkgs.Pop()

		var pkgTaskID string
		if existPkgs.Contain(pkg) {
			pkgTaskID = existPkgs.Get(pkg)
		} else {
			pkgTaskID, err = stlerr.ErrorWith(dager.AddVertex(pkg))
			if err != nil {
				return llvm.Module{}, err
			}
			existPkgs.Set(pkg, pkgTaskID)
		}

		if pkg.IsBuildIn() {
			buildinPkgTaskID = pkgTaskID
		}

		for _, depPkg := range pkg.GetDependencyPackages() {
			var depTaskID string
			if existPkgs.Contain(depPkg) {
				depTaskID = existPkgs.Get(depPkg)
			} else {
				pkgs.Push(depPkg)
				depTaskID, err = stlerr.ErrorWith(dager.AddVertex(depPkg))
				if err != nil {
					return llvm.Module{}, err
				}
				existPkgs.Set(depPkg, depTaskID)
			}
			if err = stlerr.ErrorWrap(dager.AddEdge(depTaskID, pkgTaskID)); err != nil {
				return llvm.Module{}, err
			}
		}
	}

	moduleCh := make(chan llvm.Module, 1)
	go func() {
		defer close(moduleCh)
		var resList []dag.FlowResult
		resList, err = stlerr.ErrorWith(dager.DescendantsFlow(buildinPkgTaskID, nil, func(d *dag.DAG, id string, depResults []dag.FlowResult) (interface{}, error) {
			_, err := stlslices.MapError(depResults, func(_ int, res dag.FlowResult) (any, error) {
				if res.Error != nil {
					return nil, res.Error
				}
				return nil, nil
			})
			if err != nil {
				return nil, err
			}

			pkgObj, err := stlerr.ErrorWith(d.GetVertex(id))
			if err != nil {
				return nil, err
			}
			pkg := pkgObj.(*hir.Package)
			module := New(target, pkg).Codegen()
			moduleCh <- module

			return nil, nil
		}))
		if err == nil {
			_, err = stlslices.MapError(resList, func(_ int, res dag.FlowResult) (any, error) {
				if res.Error != nil {
					return nil, res.Error
				}
				return nil, nil
			})
		}
	}()
	if err != nil {
		return llvm.Module{}, err
	}

	var module llvm.Module
	var existBaseModule bool
	for backModule := range moduleCh {
		if !existBaseModule {
			module, existBaseModule = backModule, true
		} else {
			err = stlerr.ErrorWrap(backModule.Link(module))
			if err != nil {
				return llvm.Module{}, err
			}
			module = backModule
		}
	}

	passOption := llvm.NewPassOption()
	defer passOption.Free()
	err = stlerr.ErrorWrap(module.RunPasses(passOption, append(modulePasses, functionPasses...)...))
	return module, nil
}
