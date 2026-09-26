//go:build codegen

package main

import (
	"errors"
	"flag"
	"fmt"
	"os"

	"github.com/kkkunny/Sim/compiler/analyze"
	"github.com/kkkunny/Sim/compiler/codegen"
	"github.com/kkkunny/Sim/compiler/config"
	"github.com/kkkunny/Sim/compiler/hir/globals"
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
		fmt.Fprintln(os.Stderr, "usage: codegen [-root <dir>] <source.sim|dir>")
		os.Exit(2)
	}

	pkg, err := analyze.Analyze(flag.Arg(0))
	if err != nil {
		if !errors.Is(err, report.ErrReported) {
			fmt.Fprintf(os.Stderr, "error: %v\n", err)
		}
		os.Exit(1)
	}

	ctx := codegen.NewContext()
	defer ctx.Close()

	for _, p := range collectPkgs(pkg) {
		gen := codegen.New(ctx, p)
		gen.Generate()
		fmt.Println(gen.Module())
		if err := gen.Close(); err != nil {
			panic(err)
		}
	}
}

// collectPkgs 后序收集包（依赖优先）
func collectPkgs(pkg *globals.Package) []*globals.Package {
	var pkgs []*globals.Package
	seen := make(map[*globals.Package]bool)
	var walk func(p *globals.Package)
	walk = func(p *globals.Package) {
		if seen[p] {
			return
		}
		seen[p] = true
		for _, dep := range p.Dependencies {
			walk(dep)
		}
		pkgs = append(pkgs, p)
	}
	walk(pkg)
	return pkgs
}
