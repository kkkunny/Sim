package runtime

// #include "types.h"
import "C"

import (
	"strconv"
	"unsafe"
)

type I8 C.i8

func NewI8(v int8) I8 {
	return I8(v)
}

func (self I8) Value() int8 {
	return int8(self)
}

func (self I8) String() string {
	return strconv.FormatInt(int64(self.Value()), 10)
}

type I16 C.i16

func NewI16(v int16) I16 {
	return I16(v)
}

func (self I16) Value() int16 {
	return int16(self)
}

func (self I16) String() string {
	return strconv.FormatInt(int64(self.Value()), 10)
}

type I32 C.i32

func NewI32(v int32) I32 {
	return I32(v)
}

func (self I32) Value() int32 {
	return int32(self)
}

func (self I32) String() string {
	return strconv.FormatInt(int64(self.Value()), 10)
}

type I64 C.i64

func NewI64(v int64) I64 {
	return I64(v)
}

func (self I64) Value() int64 {
	return int64(self)
}

func (self I64) String() string {
	return strconv.FormatInt(self.Value(), 10)
}

type Isize C.isize

func NewIsize(v int) Isize {
	return Isize(v)
}

func (self Isize) Value() int {
	return int(self)
}

func (self Isize) String() string {
	return strconv.FormatInt(int64(self.Value()), 10)
}

type U8 C.u8

func NewU8(v uint8) U8 {
	return U8(v)
}

func (self U8) Value() uint8 {
	return uint8(self)
}

func (self U8) String() string {
	return strconv.FormatUint(uint64(self.Value()), 10)
}

type U16 C.u16

func NewU16(v uint16) U16 {
	return U16(v)
}

func (self U16) Value() uint16 {
	return uint16(self)
}

func (self U16) String() string {
	return strconv.FormatUint(uint64(self.Value()), 10)
}

type U32 C.u32

func NewU32(v uint32) U32 {
	return U32(v)
}

func (self U32) Value() uint32 {
	return uint32(self)
}

func (self U32) String() string {
	return strconv.FormatUint(uint64(self.Value()), 10)
}

type U64 C.u64

func NewU64(v uint64) U64 {
	return U64(v)
}

func (self U64) Value() uint64 {
	return uint64(self)
}

func (self U64) String() string {
	return strconv.FormatUint(self.Value(), 10)
}

type Usize C.usize

func NewUsize(v uint) Usize {
	return Usize(v)
}

func (self Usize) Value() uint {
	return uint(self)
}

func (self Usize) String() string {
	return strconv.FormatUint(uint64(self.Value()), 10)
}

type Bool C.bool

func NewBool(v bool) Bool {
	if v {
		return 1
	} else {
		return 0
	}
}

func (self Bool) Value() bool {
	return self != 0
}

func (self Bool) String() string {
	if self.Value() {
		return "true"
	} else {
		return "false"
	}
}

type Str C.str

func NewStr(v string) Str {
	return *(*Str)(unsafe.Pointer(&v))
}

func (self Str) Data() *byte {
	return (*byte)(self.data)
}

func (self Str) Length() uint {
	return uint(self.len)
}

func (self Str) Value() string {
	return unsafe.String(self.Data(), self.Length())
}

func (self Str) String() string {
	return self.Value()
}
