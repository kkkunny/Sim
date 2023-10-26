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

func TestIf(t *testing.T) {
	assertRetEqZero(t, `
func main()u8{
    if true{
        return 0
    }
    return 1
}
`)
}
