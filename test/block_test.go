package test

import "testing"

func TestBlock(t *testing.T) {
	assertRetEqZero(t, `
func main()u8{
    {
        let i: u8 = 1;
        return i - 1
    }
}
`)
}
