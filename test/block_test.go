package test

import "testing"

func TestBlock(t *testing.T) {
	assertRetEqZero(t, `
func main()u8{
    {
        let i: u8 = 1;
        return i - 1
    }
}
`)
}

func TestIf(t *testing.T) {
	assertRetEqZero(t, `
func main()u8{
    if true{
        return 0
    }
    return 1
}
`)
}

func TestIfElse(t *testing.T) {
	assertRetEqZero(t, `
func main()u8{
    if true{
        return 0
    }else{
		return 1
	}
}
`)
}

func TestElseIf(t *testing.T) {
	assertRetEqZero(t, `
func main()u8{
    if false{
        return 1
    }else if true{
		return 0
	}else{
		return 1
	}
}
`)
}

func TestElseIfElseIf(t *testing.T) {
	assertRetEqZero(t, `
func main()u8{
    if false{
        return 1
    }else if false{
		return 1
	}else if true{
		return 0
	}else{
		return 1
	}
}
`)
}

func TestNestedIfElse(t *testing.T) {
	assertRetEqZero(t, `
func main()u8{
    if false{
        return 1
    }else if true{
		if false{
			return 1
		}else if true{
			if true{
				return 0
			}
			return 1
		}else{
			return 1
		}
	}else{
		return 1
	}
}
`)
}
