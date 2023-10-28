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

func TestFlaotEq(t *testing.T) {
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
