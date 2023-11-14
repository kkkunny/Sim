package test

import (
	"testing"
)

func TestBool(t *testing.T) {
	assertRetEqZero(t, `
func main()u8{
	if false{
		return 1
	}
    return 0
}
`)
}
