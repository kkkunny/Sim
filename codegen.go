//go:build codegen

package main

import (
	"fmt"
	"os"

	stlerror "github.com/kkkunny/stl/error"

	"github.com/kkkunny/Sim/analyse"
	"github.com/kkkunny/Sim/codegen"
	"github.com/kkkunny/Sim/lex"
	"github.com/kkkunny/Sim/parse"
	"github.com/kkkunny/Sim/reader"
)

func main() {
	r := stlerror.MustWith(reader.NewReaderFromFile(os.Args[1]))
	generator := codegen.New(analyse.New(parse.New(lex.New(r))))
	module := generator.Codegen()
	fmt.Println(module)
}