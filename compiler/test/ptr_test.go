package test

import (
	"testing"
)

func TestPointer(t *testing.T) {
	assertRetEqZero(t, `
func get(v: u8)*u8{
    return &v
}

func main()u8{
    return *get(1)
}
`)
}
