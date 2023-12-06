package test

import "testing"

func TestGenericFunction(t *testing.T) {
	assertRetEqZero(t, `
func get<T>(v: T)T{
    return v
}

func main()u8{
    let i: u8 = get::<u8>(1)
    let j: isize = get::<isize>(1)
    return i - j as u8
}
`)
}
