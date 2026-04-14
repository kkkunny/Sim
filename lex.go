//go:build lex

package main

import (
	"bytes"
	"fmt"
	"os"

	stlerror "github.com/kkkunny/stl/error"

	"github.com/kkkunny/Sim/compiler/lex"

	"github.com/kkkunny/Sim/compiler/token"
)

func main() {
	data := stlerror.MustWith(os.ReadFile(os.Args[1]))
	lexer := lex.New(bytes.NewReader(data))
	for tok := lexer.Scan(); !tok.Is(token.KindEnum.Eof); tok = lexer.Scan() {
		fmt.Printf("%s:", os.Args[1])
		fmt.Println(tok)
	}
}
