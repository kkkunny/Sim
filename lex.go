//go:build lex

package main

import (
	"fmt"
	"os"
	"path/filepath"

	stlerror "github.com/kkkunny/stl/error"

	"github.com/kkkunny/Sim/compiler/lex"
	"github.com/kkkunny/Sim/compiler/reader"

	"github.com/kkkunny/Sim/compiler/token"
)

func main() {
	file := stlerror.MustWith(os.Open(os.Args[1]))
	defer file.Close()
	lexer := lex.New(reader.NewFile(os.Args[1], file))
	wd := stlerror.MustWith(os.Getwd())
	relpath := stlerror.MustWith(filepath.Rel(wd, os.Args[1]))
	for tok := lexer.Scan(); !tok.Is(token.KindEnum.Eof); tok = lexer.Scan() {
		fmt.Printf("%s:", relpath)
		fmt.Println(tok)
	}
}
