package test

import (
	"testing"
)

func TestComment(t *testing.T) {
	assertRetEqZero(t, `
func main()u8{
    // 这是一个注释
    let i = 1  // 这还是一个注释
    return 0
}
`)
}
