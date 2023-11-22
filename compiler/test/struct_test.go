package test

import "testing"

func TestStruct(t *testing.T) {
	assertRetEqZero(t, `
struct T{
    data: u8
}

func get()T{
    return T{data: 1}
}

func main()u8{
    return get().data - 1
}
`)
}

func TestField(t *testing.T) {
	assertRetEqZero(t, `
struct T{
    data: u8
}

func main()u8{
	let mut s: T = T{data: 1}
	s.data = 3
    return s.data - 3
}
`)
}
