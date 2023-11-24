package test

import "testing"

func TestTuple(t *testing.T) {
	assertRetEqZero(t, `
func get()(u8, u8){
    return (1, 2)
}

func main()u8{
    return get().0 - 1
}
`)
}

func TestExtract(t *testing.T) {
	assertRetEqZero(t, `
func main()u8{
	let mut t:(u8, u8) = (1, 2)
	t.0 = 3
    return t.0 - 3
}
`)
}
