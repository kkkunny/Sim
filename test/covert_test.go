package test

import "testing"

func TestCovert(t *testing.T) {
	assertRetEqZero(t, `
func main()u8{
    return get() as u8 - 1
}

func get()isize{
    return 1
}
`)
}
