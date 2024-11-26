//go:build parse

package main

import (
	"fmt"
	"os"

	stlerror "github.com/kkkunny/stl/error"
	stlos "github.com/kkkunny/stl/os"

	"github.com/kkkunny/Sim/compiler/parse"
)

func main() {
	astsList := stlerror.MustWith(parse.Parse(stlos.NewFilePath(os.Args[1])))
	for j, asts := range astsList {
		for i, iter := 0, asts.Iterator(); iter.Next(); i++ {
			stlerror.Must(iter.Value().Output(os.Stdout, 0))
			if i < int(asts.Length())-1 {
				stlerror.MustWith(fmt.Fprintf(os.Stdout, "\n\n"))
			}
		}
		if j < len(astsList)-1 {
			stlerror.MustWith(fmt.Fprintf(os.Stdout, "\n\n"))
		}
	}
}
