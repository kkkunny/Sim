package test

import (
	"testing"
)

func TestNegate(t *testing.T) {
	assertRetEqZero(t, `
func main()isize{
    return - 1 + 1
}
`)
}
