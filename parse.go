//go:build parse

package main

import (
	"fmt"
	"os"
	"reflect"

	stlerror "github.com/kkkunny/stl/error"

	"github.com/kkkunny/Sim/ast"
	"github.com/kkkunny/Sim/lex"
	"github.com/kkkunny/Sim/parse"
	"github.com/kkkunny/Sim/reader"
)

func main() {
	r := stlerror.MustWith(reader.NewReaderFromFile(os.Args[1]))
	parser := parse.New(lex.New(r))
	parser.Parse().Iterator().Foreach(func(v ast.Global) bool {
		fmt.Println(reflect.TypeOf(v).String())
		return true
	})
}
