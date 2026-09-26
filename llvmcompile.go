//go:build llvmcompile

package main

import (
	"fmt"
	"os"
	"os/exec"
	"path/filepath"

	"github.com/kkkunny/go-llvm/target"
	stlerr "github.com/kkkunny/stl/error"

	"github.com/kkkunny/Sim/compiler/analyze"
	"github.com/kkkunny/Sim/compiler/config"
	"github.com/kkkunny/Sim/compiler/hir/globals"
	"github.com/kkkunny/Sim/compiler/llgen"
	"github.com/kkkunny/Sim/compiler/util"
)

// 迁移期临时驱动：analyze → llgen → 产出 .o → clang 链接 main.out
func main() {
	pkg, err := analyze.Analyze(os.Args[1])
	if err != nil {
		panic(err)
	}

	ctx := llgen.NewContext()
	defer ctx.Close()

	// 依赖优先（后序遍历）
	var pkgs []*globals.Package
	seen := make(map[*globals.Package]bool)
	var collect func(p *globals.Package)
	collect = func(p *globals.Package) {
		if seen[p] {
			return
		}
		seen[p] = true
		for _, dep := range p.Dependencies {
			collect(dep)
		}
		pkgs = append(pkgs, p)
	}
	collect(pkg)

	tmpDir := stlerr.MustWith(os.MkdirTemp("", "sim-llvm"))
	defer os.RemoveAll(tmpDir)

	var objs []string
	for _, p := range pkgs {
		gen := llgen.New(ctx, p)
		gen.Generate()
		objPath := filepath.Join(tmpDir, p.Name+".o")
		if err := ctx.TargetMachine().EmitToFile(gen.Module(), objPath, target.ObjectFile); err != nil {
			panic(err)
		}
		if err := gen.Close(); err != nil {
			panic(err)
		}
		objs = append(objs, objPath)
	}

	execPath := stlerr.MustWith(util.LookupCCompiler())
	outPath := filepath.Join(config.WorkPath, "main.out")
	args := append(append([]string{}, objs...), "-o", outPath, "-lm")
	cmder := exec.Command(execPath, args...)
	cmder.Stdin, cmder.Stdout, cmder.Stderr = os.Stdin, os.Stdout, os.Stderr
	if err := stlerr.ErrorWrap(cmder.Run()); err != nil {
		panic(err)
	}
	fmt.Println("wrote " + outPath)
}
