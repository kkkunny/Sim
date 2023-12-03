package test

import (
	"testing"
)

func TestTypeAlias(t *testing.T) {
	assertRetEqZero(t, `
type byte = u8
func main()u8{
	let v: byte = 1
    return v - 1
}
`)
}
