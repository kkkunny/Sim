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
