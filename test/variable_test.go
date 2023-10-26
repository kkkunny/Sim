package test

import "testing"

func TestVariable(t *testing.T) {
	assertRetEqZero(t, `
func main()u8{
    let i: u8 = 1;
    return i - 1
}
`)
}
