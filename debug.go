//go:build debug

package main

import (
	"errors"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"

	stlerr "github.com/kkkunny/stl/error"

	"github.com/kkkunny/Sim/compiler/analyze"
	"github.com/kkkunny/Sim/compiler/compile"
	"github.com/kkkunny/Sim/compiler/config"
	"github.com/kkkunny/Sim/compiler/report"
)

func main() {
	defer report.RecoverICE()
	report.SetupConsole()

	testFilePath := "examples/main.sim"

	pkg, err := analyze.Analyze(testFilePath)
	if err != nil {
		if !errors.Is(err, report.ErrReported) {
			fmt.Fprintf(os.Stderr, "error: %v\n", err)
		}
		os.Exit(1)
	}
	err = compile.NewCompiler().Compile(pkg)
	if err != nil {
		fmt.Fprintf(os.Stderr, "error: %v\n", err)
		os.Exit(1)
	}

	outPath := filepath.Join(config.WorkPath, "main.out")
	defer os.Remove(outPath)
	cmder := exec.Command(outPath)
	cmder.Stdin, cmder.Stdout, cmder.Stderr = os.Stdin, os.Stdout, os.Stderr
	if err = stlerr.ErrorWrap(cmder.Run()); err != nil {
		var exitErr *exec.ExitError
		if !errors.As(err, &exitErr) {
			fmt.Fprintf(os.Stderr, "error: %v\n", err)
			os.Exit(1)
		} else {
			os.Exit(exitErr.ExitCode())
		}
	}
}
