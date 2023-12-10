package mir

import (
	"runtime"

	stlbasic "github.com/kkkunny/stl/basic"
	"github.com/kkkunny/stl/container/pair"
	stlmath "github.com/kkkunny/stl/math"
	stlos "github.com/kkkunny/stl/os"
	"github.com/samber/lo"
)

type Target interface {
	Name()string
	Equal(t Target) bool
	AlignOf(t Type)uint64
	SizeOf(t Type)stlos.Size
}

func NewTarget(s string)Target{
	switch s {
	case amd64WindowsTarget{}.Name():
		return new(amd64WindowsTarget)
	default:
		panic("unreachable")
	}
}

func DefaultTarget()Target{
	switch pair.NewPair(runtime.GOARCH, runtime.GOOS) {
	case pair.NewPair("amd64", "windows"):
		return new(amd64WindowsTarget)
	default:
		panic("unreachable")
	}
}

// amd64 && windows
type amd64WindowsTarget struct {}

func (amd64WindowsTarget) Name()string{
	return "x86_64-windows"
}

func (self *amd64WindowsTarget) Equal(t Target) bool{
	return stlbasic.Is[*amd64WindowsTarget](t)
}

func (self *amd64WindowsTarget) AlignOf(obj Type)uint64{
	switch t := obj.(type) {
	case VoidType:
		return 0
	case NumberType:
		if t.Bits()%uint64(stlos.Byte) == 0{
			return t.Bits()/uint64(stlos.Byte)
		}else{
			return t.Bits()/uint64(stlos.Byte)+1
		}
	case PtrType, FuncType:
		return 8
	case ArrayType:
		return self.AlignOf(t.Elem())
	case StructType:
		if len(t.Elems()) == 0{
			return 1
		}
		aligns := lo.Map(t.Elems(), func(item Type, _ int) uint64 {
			return item.Align()
		})
		return lo.Min([]uint64{lo.Max(aligns), 8})
	default:
		panic("unreachable")
	}
}

func (self *amd64WindowsTarget) SizeOf(obj Type)stlos.Size{
	switch t := obj.(type) {
	case VoidType:
		return 0
	case NumberType:
		return stlos.Bit * stlos.Size(t.Bits())
	case PtrType, FuncType:
		return stlos.Byte * 8
	case ArrayType:
		if t.Length() == 0{
			return stlos.Byte * 1
		}
		return self.SizeOf(t.Elem()) * stlos.Size(t.Length())
	case StructType:
		if len(t.Elems()) == 0{
			return stlos.Byte * 1
		}
		offset := stlos.Byte * 0
		for _, e := range t.Elems(){
			validAlign := stlos.Byte * stlos.Size(lo.Min([]uint64{e.Align(), 8}))
			offset = stlmath.RoundTo(offset, validAlign)
			offset += e.Size()
		}
		align := stlos.Size(self.AlignOf(t)) * stlos.Byte
		return stlmath.RoundTo(offset, align)
	default:
		panic("unreachable")
	}
}
