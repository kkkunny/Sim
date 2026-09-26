//go:build analyze

package main

import (
	"errors"
	"flag"
	"fmt"
	"os"

	"github.com/kkkunny/Sim/compiler/analyze"
	"github.com/kkkunny/Sim/compiler/config"
	"github.com/kkkunny/Sim/compiler/hir"
	"github.com/kkkunny/Sim/compiler/report"
)

func main() {
	defer report.RecoverICE()
	report.SetupConsole()

	rootPath := flag.String("root", "", "Sim 根目录（包含 std/）；缺省依次取 $SIM_ROOT、可执行文件旁、当前工作目录")
	flag.Parse()
	if *rootPath != "" {
		config.SetSimRoot(*rootPath)
	}
	if flag.NArg() < 1 {
		fmt.Fprintln(os.Stderr, "usage: analyze [-root <dir>] <source.sim|dir>")
		os.Exit(2)
	}

	pkg, err := analyze.Analyze(flag.Arg(0))
	if err != nil {
		if !errors.Is(err, report.ErrReported) {
			fmt.Fprintf(os.Stderr, "error: %v\n", err)
		}
		os.Exit(1)
	}
	hir.Print(os.Stdout, pkg)
}
