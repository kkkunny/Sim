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

func TestGenericStruct(t *testing.T) {
	t.Run("简单", func(t *testing.T) {
		assertRetEqZero(t, `
struct pair<T, F>{
    F: T,
    S: F
}

func main()u8{
    let a = pair::<u8, u16>{F: 0, S: 1}
    return a.F
}
`)
	})
	t.Run("嵌套", func(t *testing.T) {
		assertRetEqZero(t, `
struct either<T, F>{
    isLeft: bool,
    Left: T,
    Right: F
}

struct pair<T, F>{
    Value: either::<T, F>,
}

func main()u8{
    let a = pair::<u8, u16>{Value: either::<u8, u16>{isLeft: true, Left: 1, Right: 2}}
    return a.Value.Left - 1
}
`)
	})
	t.Run("混合结构体和函数", func(t *testing.T) {
		assertRetEqZero(t, `
struct either<T, F>{
    isLeft: bool,
    Left: T,
    Right: F
}

struct pair<T, F>{
    Value: either::<T, F>,
}

func new_pair<T, F>(x: T, y: F)pair::<T, F>{
    return pair::<T, F>{Value: either::<T, F>{isLeft: true, Left: x, Right: y}}
}

func main()u8{
    let a = new_pair::<u8, i16>(1, 2)
    return a.Value.Left - 1
}
`)
	})
}