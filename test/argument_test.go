package test

import "testing"

func TestArgument(t *testing.T) {
	assertRetEqZero(t, `
func add(x:u8, y:u8)u8{
    return x + y
}

func sub(x:u8, y:u8)u8{
    return x - y
}

func main()u8{
    return sub(add(1, 2), 3)
}
`)
}
