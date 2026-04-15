//go:build lex

package main

import (
	"bytes"
	"fmt"
	"os"
	"path/filepath"

	stlerror "github.com/kkkunny/stl/error"

	"github.com/kkkunny/Sim/compiler/lex"

	"github.com/kkkunny/Sim/compiler/token"
)

func main() {
	data := stlerror.MustWith(os.ReadFile(os.Args[1]))
	lexer := lex.New(bytes.NewReader(data))
	wd := stlerror.MustWith(os.Getwd())
	relpath := stlerror.MustWith(filepath.Rel(wd, os.Args[1]))
	for tok := lexer.Scan(); !tok.Is(token.KindEnum.Eof); tok = lexer.Scan() {
		fmt.Printf("%s:", relpath)
		fmt.Println(tok)
	}
}
