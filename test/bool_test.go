package test

import (
	"testing"
)

func TestBool(t *testing.T) {
	assertRetEqZero(t, `
func main()bool{
    return false
}
`)
}
