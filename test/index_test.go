package test

import "testing"

func TestIndex(t *testing.T) {
	// TODO: error
	assertRetEqZero(t, `
func test()[2]u8{
    return [2]u8{0, 2}
}

func main()u8{
    return test()[0]
}
`)
}
