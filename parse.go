//go:build parse

package main

import (
	"bytes"
	"fmt"
	"os"

	stlerror "github.com/kkkunny/stl/error"

	"github.com/kkkunny/Sim/compiler/lex"
	"github.com/kkkunny/Sim/compiler/parse"
)

func main() {
	data := stlerror.MustWith(os.ReadFile(os.Args[1]))
	lexer := lex.New(bytes.NewReader(data))
	parser := parse.New(lexer)
	fmt.Println(parser.Parse())
}
