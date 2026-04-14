package compiler

import (
	"bytes"
	"os"
	"os/exec"

	"github.com/kkkunny/Sim/compiler/codegen"
	"github.com/kkkunny/Sim/compiler/lex"
	"github.com/kkkunny/Sim/compiler/parser"
)

func Compile(sourcePath string, outputPath string) error {
	data, err := os.ReadFile(sourcePath)
	if err != nil {
		return err
	}

	lexer := lex.New(bytes.NewReader(data))
	parser := parser.New(lexer)
	ast := parser.Parse()

	codegen := codegen.New()
	cCode := codegen.Generate(ast)

	tmpFile := sourcePath + ".c"
	err = os.WriteFile(tmpFile, []byte(cCode), 0644)
	if err != nil {
		return err
	}
	defer os.Remove(tmpFile)

	cmd := exec.Command("clang", tmpFile, "-o", outputPath)
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	return cmd.Run()
}
