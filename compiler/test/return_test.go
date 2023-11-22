package test

import (
	"testing"
)

func TestReturn(t *testing.T) {
	assertRetEqZero(t, `
func main()u8{
    return 0
}
`)
}
