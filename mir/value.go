package mir

import "fmt"

type Value interface {
	fmt.Stringer
	Type()Type
}
