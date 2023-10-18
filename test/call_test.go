package test

import "testing"

func TestCall(t *testing.T) {
	assertRetEqZero(t, `
func main()isize{
    return getFunc()()
}

func get()isize{
    return 0
}

func getFunc()func()isize{
    return get
}
`)
}
