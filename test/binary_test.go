package test

import (
	"testing"
)

func TestBinary1(t *testing.T) {
	assertRetEqZero(t, `
func main()isize{
    return (0&1)|0^0
}
`)
}

func TestBinary2(t *testing.T) {
	assertRetEqZero(t, `
func main()isize{
    return (5+1-2/2*2)%2
}
`)
}

func TestLt(t *testing.T) {
	assertRetEqZero(t, `
func main()bool{
    return 2<1
}
`)
}

func TestGt(t *testing.T) {
	assertRetEqZero(t, `
func main()bool{
    return 1>2
}
`)
}

func TestLe(t *testing.T) {
	assertRetEqZero(t, `
func main()bool{
    return 2<=1
}
`)
}

func TestGe(t *testing.T) {
	assertRetEqZero(t, `
func main()bool{
    return 1>=2
}
`)
}

func TestIntEq(t *testing.T) {
	assertRetEqZero(t, `
func main()bool{
    return 1==2
}
`)
}

func TestFloatEq(t *testing.T) {
	assertRetEqZero(t, `
func main()bool{
    let i: f64 = 1.2
    return i==2.2
}
`)
}

func TestBoolEq(t *testing.T) {
	assertRetEqZero(t, `
func main()bool{
    return true == false
}
`)
}

func TestArrayEq(t *testing.T) {
	assertRetEqZero(t, `
func main()bool{
	let a: [2]i32 = [2]i32{1, 2}
	let b: [2]i32 = [2]i32{3, 2}
    return a == b
}
`)
}

func TestTupleEq(t *testing.T) {
	assertRetEqZero(t, `
func main()bool{
	let a: (i32, i32) = (1, 2)
	let b: (i32, i32) = (3, 2)
    return a == b
}
`)
}

func TestStructEq(t *testing.T) {
	assertRetEqZero(t, `
struct A{
	f1: i32,
	f2: i32
}

func main()bool{
	let a: A = A{f1: 1, f2: 2}
    return a != a
}
`)
}

func TestShl(t *testing.T) {
	assertRetEqZero(t, `
func main()bool{
    return (11 << 2) != 44
}
`)
}

func TestShr(t *testing.T) {
	assertRetEqZero(t, `
func main()bool{
    return (11 >> 2) != 2
}
`)
}

func TestLogicAnd(t *testing.T) {
	assertRetEqZero(t, `
func main()bool{
    return true && false
}
`)
}

func TestLogicOr(t *testing.T) {
	assertRetEqZero(t, `
func main()bool{
    return false || false
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

func TestUnpack(t *testing.T) {
	assertRetEqZero(t, `
func main()i32{
	let mut i: i32 = 0
	let mut j: i32 = 0
	let mut m: i32 = 0
	let n: ((i32, i32), i32) = ((1, 2), 3)
	((i, j), m) = n
    return m - 3
}
`)
}
