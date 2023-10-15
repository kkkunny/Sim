package main

import (
	"fmt"

	stlerror "github.com/kkkunny/stl/error"

	"github.com/kkkunny/Sim/analyse"
	"github.com/kkkunny/Sim/codegen"
	"github.com/kkkunny/Sim/lex"
	"github.com/kkkunny/Sim/parse"
	"github.com/kkkunny/Sim/reader"
)

func main() {
	r := stlerror.MustWith(reader.NewReaderFromFile("example/main.sim"))
	generator := codegen.New(analyse.New(parse.New(lex.New(r))))
	fmt.Println(generator.Codegen())
}
