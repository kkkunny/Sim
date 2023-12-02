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

func TestMethod(t *testing.T) {
	assertRetEqZero(t, `
struct T{
    data: u8
}

func (mut T) get_data()u8{
    return self.data
}

func main()u8{
    let v: T = T{data: 1}
    return v.get_data() - 1
}
`)
}
