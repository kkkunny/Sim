package mir

import (
	"fmt"
	"io"
	"strings"

	"github.com/kkkunny/containers/hashmap"
)

type Stringer struct {
	pkg *Package

	globals *hashmap.HashMap[Global, uint]
	blocks  *hashmap.HashMap[*Block, uint]
	values  *hashmap.HashMap[Value, uint]
}

func NewStringer(pkg *Package) *Stringer {
	return &Stringer{
		pkg:     pkg,
		globals: hashmap.NewHashMap[Global, uint](),
		blocks:  hashmap.NewHashMap[*Block, uint](),
		values:  hashmap.NewHashMap[Value, uint](),
	}
}
func (self Stringer) Output(w io.Writer) {
	var aliasCounter uint
	for cursor := self.pkg.Globals.Front(); cursor != nil; cursor = cursor.Next() {
		global := cursor.Value()
		if _, ok := global.(*Alias); ok {
			self.globals.Set(global, aliasCounter)
			aliasCounter++
		}
	}

	var globalValueCounter uint
	for cursor := self.pkg.Globals.Front(); cursor != nil; cursor = cursor.Next() {
		global := cursor.Value()
		if _, ok := global.(Value); ok {
			self.globals.Set(global, globalValueCounter)
			globalValueCounter++
		}
	}

	for cursor := self.pkg.Globals.Front(); cursor != nil; cursor = cursor.Next() {
		_, _ = io.WriteString(w, self.globalString(cursor.Value()))
		if cursor.Next() != nil {
			_, _ = io.WriteString(w, "\n")
		}
	}
}
func (self *Stringer) globalString(ir Global) string {
	switch global := ir.(type) {
	case *Alias:
		return fmt.Sprintf("type %s = %s", self.globalName(global), self.typeString(global.Target))
	case *Function:
		self.blocks.Clear()
		self.values.Clear()
		var buf strings.Builder
		buf.WriteString("func ")
		buf.WriteString(self.globalName(global))
		buf.WriteByte('(')
		for i, p := range global.Params {
			self.values.Set(p, uint(i))
			buf.WriteString(self.valueName(p))
			buf.WriteByte(' ')
			buf.WriteString(self.typeString(p.Type))
			if i < len(global.Params)-1 {
				buf.WriteString(", ")
			}
		}
		if global.Type.GetFuncVarArg() {
			buf.WriteString(", ...")
		}
		buf.WriteByte(')')
		buf.WriteString(self.typeString(global.Type.GetFuncRet()))
		if global.Extern {
			buf.WriteString(" #extern")
		}
		if global.inline {
			buf.WriteString(" #inline")
		}
		if global.noInline {
			buf.WriteString(" #noinline")
		}
		if global.NoReturn {
			buf.WriteString(" #noreturn")
		}
		if global.init {
			buf.WriteString(" #init")
		}
		if global.fini {
			buf.WriteString(" #fini")
		}
		if global.Blocks.Length() != 0 {
			buf.WriteByte('\n')
			var blockCounter uint
			for cursor := global.Blocks.Front(); cursor != nil; cursor = cursor.Next() {
				self.blocks.Set(cursor.Value(), blockCounter)
				blockCounter++
			}
			varCounter := uint(len(global.Params))
			for cursor := global.Blocks.Front(); cursor != nil; cursor = cursor.Next() {
				buf.WriteString(self.blockString(cursor.Value(), &varCounter))
				buf.WriteByte('\n')
			}
		}
		return buf.String()
	case *Variable:
		var buf strings.Builder
		buf.WriteString("var ")
		buf.WriteString(self.globalName(global))
		buf.WriteByte(' ')
		buf.WriteString(self.typeString(global.Type))
		if global.Value != nil {
			buf.WriteString(" = ")
			buf.WriteString(self.valueName(global.Value))
		}
		if global.Extern {
			buf.WriteString(" #extern")
		}
		return buf.String()
	default:
		panic("unreachable")
	}
}
func (self *Stringer) globalName(ir Global) string {
	switch global := ir.(type) {
	case *Alias:
		if global.Name != "" {
			return global.Name
		}
		n, _ := self.globals.Get(global)
		return fmt.Sprintf("t%d", n)
	case *Function:
		if global.Name != "" {
			return global.Name
		}
		n, _ := self.globals.Get(global)
		return fmt.Sprintf("g%d", n)
	case *Variable:
		if global.Name != "" {
			return global.Name
		}
		n, _ := self.globals.Get(global)
		return fmt.Sprintf("g%d", n)
	default:
		panic("unreachable")
	}
}
func (self *Stringer) blockString(ir *Block, varCounter *uint) string {
	var buf strings.Builder
	buf.WriteString(self.blockName(ir))
	buf.WriteString(":\n")
	for cursor := ir.Insts.Front(); cursor != nil; cursor = cursor.Next() {
		if v, ok := cursor.Value().(Value); ok {
			self.values.Set(v, *varCounter)
			*varCounter++
		}
	}
	for cursor := ir.Insts.Front(); cursor != nil; cursor = cursor.Next() {
		buf.WriteByte('\t')
		buf.WriteString(self.instString(cursor.Value()))
		if cursor.Next() != nil {
			buf.WriteByte('\n')
		}
	}
	return buf.String()
}
func (self *Stringer) blockName(ir *Block) string {
	n, _ := self.blocks.Get(ir)
	return fmt.Sprintf("b%d", n)
}
func (self *Stringer) instString(ir Inst) string {
	switch inst := ir.(type) {
	case *Alloc:
		return fmt.Sprintf(
			"%s %s = alloc %s",
			self.typeString(inst.GetType()),
			self.valueName(inst),
			self.typeString(inst.Type),
		)
	case *Load:
		return fmt.Sprintf(
			"%s %s = load %s",
			self.typeString(inst.GetType()),
			self.valueName(inst),
			self.valueName(inst.Value),
		)
	case *Store:
		return fmt.Sprintf("store %s to %s", self.valueName(inst.From), self.valueName(inst.To))
	case *Unreachable:
		return "unreachable"
	case *Return:
		if inst.Value == nil {
			return "ret void"
		}
		return fmt.Sprintf("ret %s", self.valueName(inst.Value))
	case *Jmp:
		return fmt.Sprintf("jmp %s", self.blockName(inst.Dst))
	case *CondJmp:
		return fmt.Sprintf(
			"if %s jmp %s or %s",
			self.valueName(inst.Cond),
			self.blockName(inst.True),
			self.blockName(inst.False),
		)
	case *Eq:
		return fmt.Sprintf(
			"%s %s = eq %s with %s",
			self.typeString(inst.GetType()),
			self.valueName(inst),
			self.valueName(inst.Left),
			self.valueName(inst.Right),
		)
	case *Ne:
		return fmt.Sprintf(
			"%s %s = ne %s with %s",
			self.typeString(inst.GetType()),
			self.valueName(inst),
			self.valueName(inst.Left),
			self.valueName(inst.Right),
		)
	case *Lt:
		return fmt.Sprintf(
			"%s %s = lt %s with %s",
			self.typeString(inst.GetType()),
			self.valueName(inst),
			self.valueName(inst.Left),
			self.valueName(inst.Right),
		)
	case *Le:
		return fmt.Sprintf(
			"%s %s = le %s with %s",
			self.typeString(inst.GetType()),
			self.valueName(inst),
			self.valueName(inst.Left),
			self.valueName(inst.Right),
		)
	case *Gt:
		return fmt.Sprintf(
			"%s %s = gt %s with %s",
			self.typeString(inst.GetType()),
			self.valueName(inst),
			self.valueName(inst.Left),
			self.valueName(inst.Right),
		)
	case *Ge:
		return fmt.Sprintf(
			"%s %s = ge %s with %s",
			self.typeString(inst.GetType()),
			self.valueName(inst),
			self.valueName(inst.Left),
			self.valueName(inst.Right),
		)
	case *ArrayIndex:
		return fmt.Sprintf(
			"%s %s = index %s %s",
			self.typeString(inst.GetType()),
			self.valueName(inst),
			self.valueName(inst.From),
			self.valueName(inst.Index),
		)
	case *StructIndex:
		return fmt.Sprintf(
			"%s %s = index %s %d",
			self.typeString(inst.GetType()),
			self.valueName(inst),
			self.valueName(inst.From),
			inst.Index,
		)
	case *PointerIndex:
		return fmt.Sprintf(
			"%s %s = index %s %s",
			self.typeString(inst.GetType()),
			self.valueName(inst),
			self.valueName(inst.From),
			self.valueName(inst.Index),
		)
	case *UnwrapUnion:
		return fmt.Sprintf(
			"%s %s = unwrap union %s",
			self.typeString(inst.GetType()),
			self.valueName(inst),
			self.valueName(inst.Value),
		)
	case *Add:
		return fmt.Sprintf(
			"%s %s = add %s %s",
			self.typeString(inst.GetType()),
			self.valueName(inst),
			self.valueName(inst.Left),
			self.valueName(inst.Right),
		)
	case *Sub:
		return fmt.Sprintf(
			"%s %s = sub %s %s",
			self.typeString(inst.GetType()),
			self.valueName(inst),
			self.valueName(inst.Left),
			self.valueName(inst.Right),
		)
	case *Mul:
		return fmt.Sprintf(
			"%s %s = mul %s %s",
			self.typeString(inst.GetType()),
			self.valueName(inst),
			self.valueName(inst.Left),
			self.valueName(inst.Right),
		)
	case *Div:
		return fmt.Sprintf(
			"%s %s = div %s %s",
			self.typeString(inst.GetType()),
			self.valueName(inst),
			self.valueName(inst.Left),
			self.valueName(inst.Right),
		)
	case *Mod:
		return fmt.Sprintf(
			"%s %s = mod %s %s",
			self.typeString(inst.GetType()),
			self.valueName(inst),
			self.valueName(inst.Left),
			self.valueName(inst.Right),
		)
	case *And:
		return fmt.Sprintf(
			"%s %s = and %s %s",
			self.typeString(inst.GetType()),
			self.valueName(inst),
			self.valueName(inst.Left),
			self.valueName(inst.Right),
		)
	case *Or:
		return fmt.Sprintf(
			"%s %s = or %s %s",
			self.typeString(inst.GetType()),
			self.valueName(inst),
			self.valueName(inst.Left),
			self.valueName(inst.Right),
		)
	case *Xor:
		return fmt.Sprintf(
			"%s %s = xor %s %s",
			self.typeString(inst.GetType()),
			self.valueName(inst),
			self.valueName(inst.Left),
			self.valueName(inst.Right),
		)
	case *Shl:
		return fmt.Sprintf(
			"%s %s = shl %s %s",
			self.typeString(inst.GetType()),
			self.valueName(inst),
			self.valueName(inst.Left),
			self.valueName(inst.Right),
		)
	case *Shr:
		return fmt.Sprintf(
			"%s %s = shr %s %s",
			self.typeString(inst.GetType()),
			self.valueName(inst),
			self.valueName(inst.Left),
			self.valueName(inst.Right),
		)
	case *Phi:
		var buf strings.Builder
		buf.WriteString(self.typeString(inst.GetType()))
		buf.WriteByte(' ')
		buf.WriteString(self.valueName(inst))
		buf.WriteString(" = phi [")
		for i, b := range inst.Froms {
			buf.WriteString(self.blockName(b))
			buf.WriteByte(' ')
			buf.WriteString(self.valueName(inst.Values[i]))
			if i < len(inst.Froms)-1 {
				buf.WriteString(", ")
			}
		}
		buf.WriteByte(']')
		return buf.String()
	case *Sint2Sint:
		return fmt.Sprintf(
			"%s %s = %s to %s",
			self.typeString(inst.GetType()),
			self.valueName(inst),
			self.valueName(inst.From),
			self.typeString(inst.To),
		)
	case *Sint2Uint:
		return fmt.Sprintf(
			"%s %s = %s to %s",
			self.typeString(inst.GetType()),
			self.valueName(inst),
			self.valueName(inst.From),
			self.typeString(inst.To),
		)
	case *Sint2Float:
		return fmt.Sprintf(
			"%s %s = %s to %s",
			self.typeString(inst.GetType()),
			self.valueName(inst),
			self.valueName(inst.From),
			self.typeString(inst.To),
		)
	case *Sint2Ptr:
		return fmt.Sprintf(
			"%s %s = %s to %s",
			self.typeString(inst.GetType()),
			self.valueName(inst),
			self.valueName(inst.From),
			self.typeString(inst.To),
		)
	case *Uint2Uint:
		return fmt.Sprintf(
			"%s %s = %s to %s",
			self.typeString(inst.GetType()),
			self.valueName(inst),
			self.valueName(inst.From),
			self.typeString(inst.To),
		)
	case *Uint2Sint:
		return fmt.Sprintf(
			"%s %s = %s to %s",
			self.typeString(inst.GetType()),
			self.valueName(inst),
			self.valueName(inst.From),
			self.typeString(inst.To),
		)
	case *Uint2Float:
		return fmt.Sprintf(
			"%s %s = %s to %s",
			self.typeString(inst.GetType()),
			self.valueName(inst),
			self.valueName(inst.From),
			self.typeString(inst.To),
		)
	case *Uint2Ptr:
		return fmt.Sprintf(
			"%s %s = %s to %s",
			self.typeString(inst.GetType()),
			self.valueName(inst),
			self.valueName(inst.From),
			self.typeString(inst.To),
		)
	case *Float2Sint:
		return fmt.Sprintf(
			"%s %s = %s to %s",
			self.typeString(inst.GetType()),
			self.valueName(inst),
			self.valueName(inst.From),
			self.typeString(inst.To),
		)
	case *Float2Uint:
		return fmt.Sprintf(
			"%s %s = %s to %s",
			self.typeString(inst.GetType()),
			self.valueName(inst),
			self.valueName(inst.From),
			self.typeString(inst.To),
		)
	case *Float2Float:
		return fmt.Sprintf(
			"%s %s = %s to %s",
			self.typeString(inst.GetType()),
			self.valueName(inst),
			self.valueName(inst.From),
			self.typeString(inst.To),
		)
	case *Ptr2Ptr:
		return fmt.Sprintf(
			"%s %s = %s to %s",
			self.typeString(inst.GetType()),
			self.valueName(inst),
			self.valueName(inst.From),
			self.typeString(inst.To),
		)
	case *Ptr2Sint:
		return fmt.Sprintf(
			"%s %s = %s to %s",
			self.typeString(inst.GetType()),
			self.valueName(inst),
			self.valueName(inst.From),
			self.typeString(inst.To),
		)
	case *Ptr2Uint:
		return fmt.Sprintf(
			"%s %s = %s to %s",
			self.typeString(inst.GetType()),
			self.valueName(inst),
			self.valueName(inst.From),
			self.typeString(inst.To),
		)
	case *Select:
		return fmt.Sprintf(
			"%s %s = if %s then %s or %s",
			self.typeString(inst.GetType()),
			self.valueName(inst),
			self.valueName(inst.Cond),
			self.valueName(inst.True),
			self.valueName(inst.False),
		)
	case *Not:
		return fmt.Sprintf(
			"%s %s = not %s",
			self.typeString(inst.GetType()),
			self.valueName(inst),
			self.valueName(inst.Value),
		)
	case *LogicNot:
		return fmt.Sprintf(
			"%s %s = logic not %s",
			self.typeString(inst.GetType()),
			self.valueName(inst),
			self.valueName(inst.Value),
		)
	case *LogicAnd:
		return fmt.Sprintf(
			"%s %s = logic and %s with %s",
			self.typeString(inst.GetType()),
			self.valueName(inst),
			self.valueName(inst.Left),
			self.valueName(inst.Right),
		)
	case *LogicOr:
		return fmt.Sprintf(
			"%s %s = logic or %s with %s",
			self.typeString(inst.GetType()),
			self.valueName(inst),
			self.valueName(inst.Left),
			self.valueName(inst.Right),
		)
	case *Call:
		var buf strings.Builder
		if !inst.GetType().IsVoid() {
			buf.WriteString(self.typeString(inst.GetType()))
			buf.WriteByte(' ')
			buf.WriteString(self.valueName(inst))
			buf.WriteString(" = ")
		}
		buf.WriteString("call ")
		buf.WriteString(self.valueName(inst.Func))
		buf.WriteByte('(')
		for i, a := range inst.Args {
			buf.WriteString(self.valueName(a))
			if i < len(inst.Args)-1 {
				buf.WriteString(", ")
			}
		}
		buf.WriteByte(')')
		return buf.String()
	case *WrapUnion:
		return fmt.Sprintf(
			"%s %s = union %s",
			self.typeString(inst.GetType()),
			self.valueName(inst),
			self.valueName(inst.Value),
		)
	default:
		panic("unreachable")
	}
}
func (self *Stringer) valueName(ir Value) string {
	switch value := ir.(type) {
	case Global:
		return self.globalName(value)
	case Constant:
		switch constant := value.(type) {
		case *Bool:
			if constant.Value {
				return "true"
			}
			return "false"
		case *Sint:
			return constant.Value.String()
		case *Uint:
			return constant.Value.String()
		case *Float:
			return constant.Value.String()
		case *EmptyPtr, *EmptyFunc, *EmptyArray, *EmptyStruct, *EmptyUnion:
			return "empty"
		case *Array:
			var buf strings.Builder
			buf.WriteByte('(')
			for i, e := range constant.Elems {
				buf.WriteString(self.valueName(e))
				if i < len(constant.Elems)-1 {
					buf.WriteByte(',')
				}
			}
			buf.WriteByte(')')
			return buf.String()
		case *Struct:
			var buf strings.Builder
			buf.WriteByte('{')
			for i, e := range constant.Elems {
				buf.WriteString(self.valueName(e))
				if i < len(constant.Elems)-1 {
					buf.WriteByte(',')
				}
			}
			buf.WriteByte('}')
			return buf.String()
		case *ArrayIndexConst:
			return fmt.Sprintf(
				"index %s %d",
				self.valueName(constant.From),
				constant.Index,
			)
		default:
			panic("unreachable")
		}
	case Inst, *Param:
		n, _ := self.values.Get(value)
		return fmt.Sprintf("v%d", n)
	default:
		panic("unreachable")
	}
}
func (self *Stringer) typeString(ir Type) string {
	switch ir.Kind {
	case TVoid:
		return "void"
	case TBool:
		return "bool"
	case TSint:
		if ir.width == 0 {
			return "sint"
		}
		return fmt.Sprintf("i%d", ir.width*8)
	case TUint:
		if ir.width == 0 {
			return "uint"
		}
		return fmt.Sprintf("u%d", ir.width*8)
	case TFloat:
		return fmt.Sprintf("f%d", ir.width*8)
	case TPtr:
		return "*" + self.typeString(ir.elems[0])
	case TFunc:
		var buf strings.Builder
		buf.WriteString("func(")
		for i, p := range ir.elems[1:] {
			buf.WriteString(self.typeString(p))
			if i < len(ir.elems)-2 {
				buf.WriteString(",")
			}
		}
		buf.WriteByte(')')
		if !ir.elems[0].IsVoid() {
			buf.WriteString(self.typeString(ir.elems[0]))
		}
		return buf.String()
	case TArray:
		return fmt.Sprintf("[%d]%s", ir.width, self.typeString(ir.elems[0]))
	case TStruct:
		var buf strings.Builder
		buf.WriteByte('{')
		for i, e := range ir.elems {
			buf.WriteString(self.typeString(e))
			if i < len(ir.elems)-1 {
				buf.WriteString(",")
			}
		}
		buf.WriteByte('}')
		return buf.String()
	case TUnion:
		var buf strings.Builder
		buf.WriteByte('<')
		for i, e := range ir.elems {
			buf.WriteString(self.typeString(e))
			if i < len(ir.elems)-1 {
				buf.WriteString(",")
			}
		}
		buf.WriteByte('>')
		return buf.String()
	case TAlias:
		return self.globalName(ir.alias)
	default:
		panic("unreachable")
	}
}
