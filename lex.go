//go:build lex

package main

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/kkkunny/Sim/compiler/lex"
	"github.com/kkkunny/Sim/compiler/reader"
	"github.com/kkkunny/Sim/compiler/report"

	"github.com/kkkunny/Sim/compiler/token"
)

func main() {
	defer report.RecoverICE()
	report.SetupConsole()

	file, err := os.Open(os.Args[1])
	if err != nil {
		fmt.Fprintf(os.Stderr, "error: %v\n", err)
		os.Exit(1)
	}
	defer file.Close()
	lexer := lex.New(reader.NewFile(os.Args[1], file))
	wd, err := os.Getwd()
	if err != nil {
		fmt.Fprintf(os.Stderr, "error: %v\n", err)
		os.Exit(1)
	}
	absPath, err := filepath.Abs(os.Args[1])
	if err != nil {
		fmt.Fprintf(os.Stderr, "error: %v\n", err)
		os.Exit(1)
	}
	relpath, err := filepath.Rel(wd, absPath)
	if err != nil || strings.HasPrefix(relpath, "..") {
		relpath = absPath
	}
	for tok := lexer.Scan(); !tok.Is(token.KindEnum.Eof); tok = lexer.Scan() {
		fmt.Printf("%s:", relpath)
		fmt.Println(tok)
	}
}
