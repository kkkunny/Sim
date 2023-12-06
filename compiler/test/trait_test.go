package test

import "testing"

func TestDefaultTrait(t *testing.T) {
	assertRetEqZero(t, `
struct A{
    data: u8
}

func (A) default()Self{
    return A{data: 1}
}

func default<T: Default>()T{
    let v: T
    return v
}

func main()u8{
    let i: A = default::<A>()
    return i.data - 1
}
`)
}
