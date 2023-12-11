package mir

import (
	"fmt"
	"runtime"

	stlbasic "github.com/kkkunny/stl/basic"
	"github.com/kkkunny/stl/container/pair"
	stlmath "github.com/kkkunny/stl/math"
	stlos "github.com/kkkunny/stl/os"
	"github.com/samber/lo"
)

type OS string

const (
	OSWindows OS = "windows"
	OSLinux OS = "linux"
)

type Arch string

const (
	ArchX8664 Arch = "x86_64"
)

func PackTargetName(arch Arch, os OS)string{
	return fmt.Sprintf("%s-%s", arch, os)
}

type Target interface {
	Name()string
	Equal(t Target) bool
	AlignOf(t Type)uint64
	SizeOf(t Type)stlos.Size
}

func NewTarget(s string)Target{
	switch s {
	case x8664WindowsTarget{}.Name():
		return new(x8664WindowsTarget)
	default:
		panic("unreachable")
	}
}

func DefaultTarget()Target{
	switch pair.NewPair(runtime.GOARCH, runtime.GOOS) {
	case pair.NewPair("amd64", "windows"):
		return new(x8664WindowsTarget)
	case pair.NewPair("amd64", "linux"):
		return new(x8664LinuxTarget)
	default:
		panic("unreachable")
	}
}

// x86_64 && windows
type x8664WindowsTarget struct {}

func (self x8664WindowsTarget) Name()string{
	return PackTargetName(self.Arch(), self.OS())
}

func (x8664WindowsTarget) OS()OS{
	return OSWindows
}

func (x8664WindowsTarget) Arch()Arch{
	return ArchX8664
}

func (self *x8664WindowsTarget) Equal(t Target) bool{
	return stlbasic.Is[*x8664WindowsTarget](t)
}

func (self *x8664WindowsTarget) AlignOf(obj Type)uint64{
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

func (self *x8664WindowsTarget) SizeOf(obj Type)stlos.Size{
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

// x86_64 && linux
type x8664LinuxTarget struct {}

func (self x8664LinuxTarget) Name()string{
	return PackTargetName(self.Arch(), self.OS())
}

func (x8664LinuxTarget) OS()OS{
	return OSLinux
}

func (x8664LinuxTarget) Arch()Arch{
	return ArchX8664
}

func (self *x8664LinuxTarget) Equal(t Target) bool{
	return stlbasic.Is[*x8664LinuxTarget](t)
}

func (self *x8664LinuxTarget) AlignOf(obj Type)uint64{
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

func (self *x8664LinuxTarget) SizeOf(obj Type)stlos.Size{
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
