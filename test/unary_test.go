package test

import (
	"testing"
)

func TestNegate(t *testing.T) {
	assertRetEqZero(t, `
func main()isize{
    return - 1 + 1
}
`)
}

func TestBitNegate(t *testing.T) {
	assertRetEqZero(t, `
func main()i32{
    let i: u8 = 10
    if !i == 245{
        return 0
    }
    return 1
}
`)
}
