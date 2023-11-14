package test

import (
	"testing"
)

func TestNegate(t *testing.T) {
	assertRetEqZero(t, `
func main()u8{
    return (- 1 + 1) as i32 as u8
}
`)
}

func TestBitNegate(t *testing.T) {
	assertRetEqZero(t, `
func main()u8{
    let i: u8 = 10
    if !i == 245{
        return 0
    }
    return 1
}
`)
}

func TestBoolNegate(t *testing.T) {
	assertRetEqZero(t, `
func main()u8{
    if !true{
		return 1
	}
	return 0
}
`)
}
