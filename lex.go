//go:build lex

package main

import (
	"fmt"
	"os"

	stlerror "github.com/kkkunny/stl/error"

	"github.com/kkkunny/Sim/lex"
	"github.com/kkkunny/Sim/reader"
	"github.com/kkkunny/Sim/token"
)

func main() {
	r := stlerror.MustWith(reader.NewReaderFromFile(os.Args[1]))
	lexer := lex.New(r)
	for tok := lexer.Scan(); !tok.Is(token.EOF); tok = lexer.Scan() {
		fmt.Println(tok)
	}
}
