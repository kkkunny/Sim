package test

import (
	"testing"
)

func TestUnion(t *testing.T) {
	assertRetEqZero(t, `
func main()u8{
    let i: i8|u8 = 0
    if i is i8{
        return 0
    }
    return 1
}
`)
}

func TestUnionTypeCovert(t *testing.T) {
	assertRetEqZero(t, `
func main()u8{
    let i: i8|u8 = 2 as u8
    if i is u8{
        return i as u8 - 2
    }
    return 1
}
`)
}
