package test

import (
	"testing"
)

func TestAnd(t *testing.T) {
	assertRetEqZero(t, `
func main()isize{
    return 0&1
}
`)
}

func TestOr(t *testing.T) {
	assertRetEqZero(t, `
func main()isize{
    return 0|0
}
`)
}

func TestXor(t *testing.T) {
	assertRetEqZero(t, `
func main()isize{
    return 0^0
}
`)
}

func TestAdd(t *testing.T) {
	assertRetEqZero(t, `
func main()isize{
    return -1+1
}
`)
}

func TestSub(t *testing.T) {
	assertRetEqZero(t, `
func main()isize{
    return 2-2
}
`)
}

func TestMul(t *testing.T) {
	assertRetEqZero(t, `
func main()isize{
    return 10*0
}
`)
}

func TestDiv(t *testing.T) {
	assertRetEqZero(t, `
func main()isize{
    return 0/10
}
`)
}

func TestRem(t *testing.T) {
	assertRetEqZero(t, `
func main()isize{
    return 10%2
}
`)
}
