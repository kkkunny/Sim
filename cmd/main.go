package main

import (
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strings"

	"github.com/kkkunny/go-llvm"
	stlslices "github.com/kkkunny/stl/container/slices"
	stlerr "github.com/kkkunny/stl/error"
	stlval "github.com/kkkunny/stl/value"
	"github.com/spf13/cobra"

	"github.com/kkkunny/Sim/compiler/codegen_ir"
	"github.com/kkkunny/Sim/compiler/interpret"
	"github.com/kkkunny/Sim/compiler/util"
)

var (
	outputPath string
)

func main() {
	rootCmd.AddCommand(buildCmd)
	buildCmd.PersistentFlags().StringVarP(&outputPath, "output", "o", "", "output path")
	rootCmd.AddCommand(runCmd)
	if err := rootCmd.Execute(); err != nil {
		_, _ = fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
}

var rootCmd = &cobra.Command{
	Use:   "sim",
	Short: "a compiler for the Sim language",
}

var buildCmd = &cobra.Command{
	Use:              "build PATH",
	Short:            "compile the sim file",
	TraverseChildren: true,
	Run: func(cmd *cobra.Command, args []string) {
		input := stlslices.First(args)
		input = stlerr.MustWith(filepath.Abs(input))
		oOutputPath := filepath.Join(filepath.Dir(input), strings.ReplaceAll(filepath.Base(input), filepath.Ext(input), ".obj"))

		llvm.EnablePrettyStackTrace()
		target := stlerr.MustWith(util.GetLLVMTarget())
		module := stlerr.MustWith(codegen_ir.CodegenIr(target, input))
		stlerr.Must(stlerr.ErrorWrap(target.WriteOBJToFile(module, string(oOutputPath), llvm.CodeOptLevelDefault, llvm.RelocModePIC, llvm.CodeModelDefault)))
		defer os.Remove(oOutputPath)

		binOutputPath := stlval.IfLazy(outputPath == "", func() string {
			return filepath.Join(filepath.Dir(oOutputPath), strings.ReplaceAll(filepath.Base(oOutputPath), filepath.Ext(oOutputPath), ".out"))
		}, func() string {
			return stlerr.MustWith(filepath.Abs(outputPath))
		})
		cmder := exec.Command("clang", string(oOutputPath), "-L.", "-lsim", "-o", string(binOutputPath))
		cmder.Stdout = os.Stdout
		cmder.Stderr = os.Stderr
		stlval.Ignore(cmder.Run())
	},
}

var runCmd = &cobra.Command{
	Use:              "run PATH",
	Short:            "run the sim file",
	TraverseChildren: true,
	Run: func(cmd *cobra.Command, args []string) {
		input := stlslices.First(args)
		input = stlerr.MustWith(filepath.Abs(input))
		llvm.EnablePrettyStackTrace()
		target := stlerr.MustWith(util.GetLLVMTarget())
		module := stlerr.MustWith(codegen_ir.CodegenIr(target, input))
		ret := stlerr.MustWith(interpret.Interpret(module))
		os.Exit(int(ret))
	},
}
