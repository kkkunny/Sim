package mir

import (
	"fmt"
	"strings"

	"github.com/bahlo/generic-list-go"
)

// Block 代码块
type Block struct {
	Belong *Function
	Insts  *list.List[Inst]
}

func (self *Function) NewBlock() *Block {
	b := &Block{
		Belong: self,
		Insts:  list.New[Inst](),
	}
	self.Blocks.PushBack(b)
	return b
}
func (self *Block) GetName() string {
	return fmt.Sprintf("b%p", self)
}
func (self Block) String() string {
	var buf strings.Builder
	buf.WriteString(self.GetName())
	buf.WriteString(":\n")
	for cursor := self.Insts.Front(); cursor != nil; cursor = cursor.Next() {
		buf.WriteByte('\t')
		buf.WriteString(cursor.Value.String())
		if cursor.Next() != nil {
			buf.WriteByte('\n')
		}
	}
	return buf.String()
}
func (self *Block) NewEqual(left, right Value) (*Block, Value) {
	t := left.GetType()
	switch {
	case t.IsNumber() || t.IsPtr() || t.IsFunc() || t.IsBool():
		return self, self.NewEq(left, right)
	case t.IsArray():
		if t.GetArraySize() == 0 {
			return self, NewBool(true)
		}
		i := self.NewAlloc(NewTypeSint(0))
		self.NewStore(NewSint(i.Type, 0), i)
		cb := self.Belong.NewBlock()
		self.NewJmp(cb)

		iv := cb.NewLoad(i)
		lb, eb := self.Belong.NewBlock(), self.Belong.NewBlock()
		lt := cb.NewLt(iv, NewSint(i.Type, int64(t.GetArraySize())))
		cb.NewCondJmp(lt, lb, eb)

		var lv, rv Value
		l := lb.NewArrayIndex(left, iv)
		if l.IsPtr() {
			lv = lb.NewLoad(l)
		} else {
			lv = l
		}
		r := lb.NewArrayIndex(right, iv)
		if r.IsPtr() {
			rv = lb.NewLoad(r)
		} else {
			rv = r
		}
		lb.NewStore(lb.NewAdd(iv, NewSint(i.Type, 1)), i)
		nlb, elemEqual := lb.NewEqual(lv, rv)
		lb.NewCondJmp(elemEqual, cb, eb)

		return eb, eb.NewPhi([]*Block{cb, nlb}, []Value{NewBool(true), NewBool(false)})
	case t.IsStruct():
		elems := t.GetStructElems()
		if len(elems) == 0 {
			return self, NewBool(true)
		}
		blocks := make([]*Block, len(elems))
		values := make([]Value, len(elems))
		lb, eb := self, self.Belong.NewBlock()
		for i := range elems {
			var lv, rv Value
			l := lb.NewStructIndex(left, uint(i))
			if l.IsPtr() {
				lv = lb.NewLoad(l)
			} else {
				lv = l
			}
			r := lb.NewStructIndex(right, uint(i))
			if r.IsPtr() {
				rv = lb.NewLoad(r)
			} else {
				rv = r
			}
			nlb, v := lb.NewEqual(lv, rv)
			blocks[i], values[i] = nlb, v
			if i < len(elems)-1 {
				nb := self.Belong.NewBlock()
				nlb.NewCondJmp(v, nb, eb)
				lb = nb
			} else {
				nlb.NewJmp(eb)
				lb = eb
			}
		}
		return eb, eb.NewPhi(blocks, values)
	default:
		panic("unreachable")
	}
}
func (self *Block) NewInt2Int(f Value, t Type) Value {
	ft := f.GetType()
	if ft.IsSint() && t.IsSint() {
		return self.NewSint2Sint(f, t)
	} else if ft.IsSint() && t.IsUint() {
		return self.NewSint2Uint(f, t)
	} else if ft.IsUint() && t.IsSint() {
		return self.NewUint2Sint(f, t)
	} else if ft.IsUint() && t.IsUint() {
		return self.NewUint2Uint(f, t)
	} else {
		panic("unreachable")
	}
}
func (self *Block) NewNumber2Number(f Value, t Type) Value {
	ft := f.GetType()
	if ft.IsInteger() && t.IsInteger() {
		return self.NewInt2Int(f, t)
	} else if ft.IsSint() && t.IsFloat() {
		return self.NewSint2Float(f, t)
	} else if ft.IsUint() && t.IsFloat() {
		return self.NewUint2Float(f, t)
	} else if ft.IsFloat() && t.IsSint() {
		return self.NewFloat2Sint(f, t)
	} else if ft.IsFloat() && t.IsUint() {
		return self.NewFloat2Uint(f, t)
	} else if ft.IsFloat() && t.IsFloat() {
		return self.NewFloat2Float(f, t)
	} else {
		panic("unreachable")
	}
}
func (self *Block) NewInt2Ptr(f Value, t Type) Value {
	ft := f.GetType()
	if ft.IsSint() && t.IsPtr() {
		return self.NewSint2Ptr(f, t)
	} else if ft.IsUint() && t.IsPtr() {
		return self.NewUint2Ptr(f, t)
	} else {
		panic("unreachable")
	}
}
func (self *Block) NewPtr2Int(f Value, t Type) Value {
	ft := f.GetType()
	if ft.IsPtr() && t.IsSint() {
		return self.NewPtr2Sint(f, t)
	} else if ft.IsPtr() && t.IsUint() {
		return self.NewPtr2Uint(f, t)
	} else {
		panic("unreachable")
	}
}
