package test

import "testing"

func TestCall(t *testing.T) {
	assertRetEqZero(t, `
func main()u8{
    return getFunc()()
}

func get()u8{
    return 0
}

func getFunc()func()u8{
    return get
}
`)
}
