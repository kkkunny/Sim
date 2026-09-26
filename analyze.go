//go:build analyze

package main

import (
	"errors"
	"fmt"
	"os"

	"github.com/kkkunny/Sim/compiler/analyze"
	"github.com/kkkunny/Sim/compiler/hir"
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
	hir.Print(os.Stdout, pkg)
}
