package test

import (
	"testing"
)

func TestGenericFunction(t *testing.T) {
	t.Run("简单", func(t *testing.T) {
		assertRetEqZero(t, `
func get<T>(x: T)T{
    return x
}

func main()u8{
    return get::<u8>(0)
}
`)
	})
	t.Run("多个参数", func(t *testing.T) {
		assertRetEqZero(t, `
func get<T, F>(x: T, y: F)T{
    return x
}

func main()u8{
    return get::<u8, i8>(0, 1)
}
`)
	})
	t.Run("嵌套", func(t *testing.T) {
		assertRetEqZero(t, `
func get2<F>(y: F)F{
    return y
}

func get1<T>(x: T)T{
    return get2::<T>(x)
}

func main()u8{
    return get1::<u8>(1) - 1
}
`)
	})
}