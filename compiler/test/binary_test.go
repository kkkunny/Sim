package test

import (
	"testing"
)

func TestBinary1(t *testing.T) {
	assertRetEqZero(t, `
func main()u8{
    return (0&1)|0^0
}
`)
}

func TestBinary2(t *testing.T) {
	assertRetEqZero(t, `
func main()u8{
    return (5+1-2/2*2)%2
}
`)
}

func TestLt(t *testing.T) {
	assertRetEqZero(t, `
func main()u8{
    if 2<1{
        return 1
    }
    return 0
}
`)
}

func TestGt(t *testing.T) {
	assertRetEqZero(t, `
func main()u8{
    if 1>2{
        return 1
    }
    return 0
}
`)
}

func TestLe(t *testing.T) {
	assertRetEqZero(t, `
func main()u8{
    if 2<=1{
        return 1
    }
    return 0
}
`)
}

func TestGe(t *testing.T) {
	assertRetEqZero(t, `
func main()u8{
    if 1>=2{
        return 1
    }
    return 0
}
`)
}

func TestIntEq(t *testing.T) {
	assertRetEqZero(t, `
func main()u8{
    if 1==2{
        return 1
    }
    return 0
}
`)
}

func TestFloatEq(t *testing.T) {
	assertRetEqZero(t, `
func main()u8{
    let i: f64 = 1.2
    if i==2.2{
        return 1
    }
    return 0
}
`)
}

func TestBoolEq(t *testing.T) {
	assertRetEqZero(t, `
func main()u8{
    if true == false{
        return 1
    }
    return 0
}
`)
}

func TestArrayEq(t *testing.T) {
	assertRetEqZero(t, `
func main()u8{
	let a: [2]i32 = [2]i32{1, 2}
	let b: [2]i32 = [2]i32{3, 2}
    if a == b{
        return 1
    }
    return 0
}
`)
}

func TestTupleEq(t *testing.T) {
	assertRetEqZero(t, `
func main()u8{
	let a: (i32, i32) = (1, 2)
	let b: (i32, i32) = (3, 2)
    if a == b{
        return 1
    }
    return 0
}
`)
}

func TestStructEq(t *testing.T) {
	assertRetEqZero(t, `
struct A{
	f1: i32,
	f2: i32
}

func main()u8{
	let a: A = A{f1: 1, f2: 2}
    if a != a{
        return 1
    }
    return 0
}
`)
}

func TestStringEq(t *testing.T) {
	assertRetEqZero(t, `
func main()u8{
    let s1: str = "123"
    let s2: str = "234"
    if s1 == s2{
        return 1
    }
    return 0
}
`)
}

func TestUnionEq(t *testing.T) {
	assertRetEqZero(t, `
func main()u8{
    let i:u8 = 1;
    let j:<u8,i8> = i;
    if j == j{
        return 0
    }else{
        return 1
    }
}
`)
}

func TestShl(t *testing.T) {
	assertRetEqZero(t, `
func main()u8{
    if (11 << 2) != 44{
        return 1
    }
    return 0
}
`)
}

func TestShr(t *testing.T) {
	assertRetEqZero(t, `
func main()u8{
    if (11 >> 2) != 2{
        return 1
    }
    return 0
}
`)
}

func TestLogicAnd(t *testing.T) {
	assertRetEqZero(t, `
func main()u8{
    if true && false{
        return 1
    }
    return 0
}
`)
}

func TestLogicOr(t *testing.T) {
	assertRetEqZero(t, `
func main()u8{
    if false || false{
        return 1
    }
    return 0
}
`)
}

func TestAssign(t *testing.T) {
	assertRetEqZero(t, `
func main()u8{
	let mut i: u8 = 10
	i = i - 10
    return i
}
`)
}

func TestUnTuple(t *testing.T) {
	assertRetEqZero(t, `
func main()u8{
	let mut i: u8 = 0
	let mut j: u8 = 0
	let mut m: u8 = 0
	let n: ((u8, u8), u8) = ((1, 2), 3)
	((i, j), m) = n
    return m - 3
}
`)
}
