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
