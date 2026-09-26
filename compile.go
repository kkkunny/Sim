//go:build compile

package main

import (
	"errors"
	"fmt"
	"os"

	"github.com/kkkunny/Sim/compiler/analyze"
	"github.com/kkkunny/Sim/compiler/compile"
	"github.com/kkkunny/Sim/compiler/report"
)

func main() {
	defer report.RecoverICE()
	report.SetupConsole()

	pkg, err := analyze.Analyze(os.Args[1])
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
}
