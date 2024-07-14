//go:build parse

package main

import (
	"fmt"
	"os"
	"reflect"

	stliter "github.com/kkkunny/stl/container/iter"
	stlerror "github.com/kkkunny/stl/error"
	stlos "github.com/kkkunny/stl/os"

	"github.com/kkkunny/Sim/compiler/ast"

	"github.com/kkkunny/Sim/compiler/parse"
)

func main() {
	asts := stlerror.MustWith(parse.Parse(stlos.NewFilePath(os.Args[1])))
	stliter.Foreach(asts, func(v ast.Global) bool {
		fmt.Println(reflect.TypeOf(v).String())
		return true
	})
}
