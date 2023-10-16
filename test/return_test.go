package test

import (
	"testing"
)

func TestReturn(t *testing.T) {
	assertRetEqZero(t, `
func main()isize{
    return 0
}
`)
}
