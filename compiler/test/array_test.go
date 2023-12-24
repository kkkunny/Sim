package test

import "testing"

func TestArray(t *testing.T) {
	assertRetEqZero(t, `
func test()[2]u8{
    return [1, 2]
}

func main()u8{
    return test()[0] - 1
}
`)
}

func TestIndex(t *testing.T) {
	assertRetEqZero(t, `
func test()[2]u8{
    return [1, 2]
}

func main()u8{
	let mut a: [2]u8 = [1, 2]
	a[0] = 3
    return a[0] - 3
}
`)
}
