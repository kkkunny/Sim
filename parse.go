//go:build parse

package main

import (
	"fmt"
	"os"
	"reflect"

	"github.com/kkkunny/stl/container/iterator"
	stlerror "github.com/kkkunny/stl/error"

	"github.com/kkkunny/Sim/ast"
	"github.com/kkkunny/Sim/lex"
	"github.com/kkkunny/Sim/parse"
	"github.com/kkkunny/Sim/reader"
)

func main() {
	f, r := stlerror.MustWith2(reader.NewReaderFromFile(os.Args[1]))
	defer f.Close()
	parser := parse.New(lex.New(r))
	iterator.Foreach(parser.Parse(), func(v ast.Global) bool {
		fmt.Println(reflect.TypeOf(v).String())
		return true
	})
}
