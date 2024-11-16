package ast

import (
	"fmt"
	"io"
	"strings"

	stlerr "github.com/kkkunny/stl/error"
)

func outputDepth(w io.Writer, depth uint) error {
	_, err := stlerr.ErrorWith(fmt.Fprintf(w, strings.Repeat("     ", int(depth))))
	return err
}

func outputf(w io.Writer, format string, a ...any) error {
	_, err := stlerr.ErrorWith(fmt.Fprintf(w, format, a...))
	return err
}
