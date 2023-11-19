package test

import (
	"testing"
)

func TestChar(t *testing.T) {
	assertRetEqZero(t, `
func main()u8{
    return '\0'
}
`)
}
