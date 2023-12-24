package test

import "testing"

func TestLoop(t *testing.T) {
	assertRetEqZero(t, `
func main()u8{
	let mut a: u8 = 10
    loop {
		if a == 0{
            break
        }else if a < 5{
			a = a - 1
			continue
		}
		a = a - 2
	}
	return a
}
`)
}

func TestFor(t *testing.T) {
	assertRetEqZero(t, `
func main()u8{
	let mut a: [2]u8 = [1, 2]
	let mut res: u8 = 3 
    for e in a{
		res = res - e
	}
	return res
}
`)
}
