//go:build debug

package main

import (
	"errors"
	"os"
	"os/exec"
	"path/filepath"

	stlerr "github.com/kkkunny/stl/error"

	"github.com/kkkunny/Sim/compiler/analyze"
	"github.com/kkkunny/Sim/compiler/codegen"
	"github.com/kkkunny/Sim/compiler/compile"
	"github.com/kkkunny/Sim/compiler/config"
	"github.com/kkkunny/Sim/compiler/parse"
	simerr "github.com/kkkunny/Sim/compiler/report"
)

func main() {
	testFilePath := "examples/main.sim"

	reporter := simerr.NewReporter()
	ast := stlerr.MustWith(parse.Parse(testFilePath, reporter))
	if reporter.HasErrors() {
		reporter.Print()
		os.Exit(1)
	}

	node := analyze.NewAnalyzer(reporter).Analyze(ast)
	if reporter.HasErrors() {
		reporter.Print()
		os.Exit(1)
	}

	builder := codegen.New().Generate(node)
	err := compile.NewCompiler(builder).Compile()
	if err != nil {
		var exitErr *exec.ExitError
		if !errors.As(err, &exitErr) {
			panic(err)
		} else {
			os.Exit(exitErr.ExitCode())
		}
	}
	outPath := filepath.Join(config.WorkPath, "main.out")
	defer os.Remove(outPath)
	cmder := exec.Command(outPath)
	cmder.Stdin, cmder.Stdout, cmder.Stderr = os.Stdin, os.Stdout, os.Stderr
	if err = stlerr.ErrorWrap(cmder.Run()); err != nil {
		var exitErr *exec.ExitError
		if !errors.As(err, &exitErr) {
			panic(err)
		} else {
			os.Exit(exitErr.ExitCode())
		}
	}
}
