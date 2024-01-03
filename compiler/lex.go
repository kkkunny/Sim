//go:build lex

package main

import (
	"fmt"
	"os"

	stlerror "github.com/kkkunny/stl/error"
	stlos "github.com/kkkunny/stl/os"

	"github.com/kkkunny/Sim/lex"
	"github.com/kkkunny/Sim/reader"
	"github.com/kkkunny/Sim/token"
)

func main() {
	f, r := stlerror.MustWith2(reader.NewReaderFromFile(stlos.NewFilePath(os.Args[1])))
	defer f.Close()
	lexer := lex.New(r)
	for tok := lexer.Scan(); !tok.Is(token.EOF); tok = lexer.Scan() {
		fmt.Println(tok)
	}
}
