package test

import "testing"

func TestLocalVariable(t *testing.T) {
	assertRetEqZero(t, `
func main()u8{
    let i: u8 = 1;
    return i - 1
}
`)
}

func TestGlobalVariable(t *testing.T) {
	assertRetEqZero(t, `
let mut i: u8 = 10;
func main()u8{
    i = i - 8
    return i - 2
}
`)
}
