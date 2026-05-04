//go:build debug

package main

import (
	"errors"
	"os"
	"os/exec"
	"path/filepath"

	stlerr "github.com/kkkunny/stl/error"

	"github.com/kkkunny/Sim/compiler/analyze"
	"github.com/kkkunny/Sim/compiler/compile"
	"github.com/kkkunny/Sim/compiler/config"
)

func main() {
	testFilePath := "examples/main.sim"

	pkg, err := analyze.Analyze(testFilePath)
	if err != nil {
		panic(err)
	}
	err = compile.NewCompiler().Compile(pkg)
	if err != nil {
		panic(err)
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
