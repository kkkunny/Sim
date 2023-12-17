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
    return *get(1) - 1
}
`)
}

func TestReference(t *testing.T) {
	assertRetEqZero(t, `
func get(v: *?u8)*?u8{
    return v
}

func main()u8{
    let v: u8 = 1
    let vr: *u8 = &v
    let vp: *?u8 = vr
    return *vp! - 1
}
`)
}
