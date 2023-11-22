//go:build parse

package main

import (
	"fmt"
	"os"
	"reflect"

	"github.com/kkkunny/stl/container/iterator"
	stlerror "github.com/kkkunny/stl/error"

	"github.com/kkkunny/Sim/ast"
	"github.com/kkkunny/Sim/parse"
)

func main() {
	asts := stlerror.MustWith(parse.ParseFile(os.Args[1]))
	iterator.Foreach(asts, func(v ast.Global) bool {
		fmt.Println(reflect.TypeOf(v).String())
		return true
	})
}
