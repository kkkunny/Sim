package test

import (
	"testing"
)

func TestPointer(t *testing.T) {
	assertRetEqZero(t, `
func test(p: *u8)u8{
    return *p
}

func main()u8{
    let i: u8 = 0
    return test(&i)
}
`)
}
