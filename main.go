package main

import (
	"fmt"

	stlerror "github.com/kkkunny/stl/error"

	"github.com/kkkunny/Sim/lexer"
	"github.com/kkkunny/Sim/parser"
	"github.com/kkkunny/Sim/reader"
)

func main() {
	r := stlerror.MustWith(reader.NewReaderFromFile("example/main.sim"))
	p := parser.New(lexer.New(r))
	asts := p.Parse()
	fmt.Println(asts)
}
