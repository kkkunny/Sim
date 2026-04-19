package compile

import (
	"errors"
	"os"
	"os/exec"
	"path/filepath"

	"github.com/kkkunny/stl/container/optional"
	stlerr "github.com/kkkunny/stl/error"
	stlos "github.com/kkkunny/stl/os"

	"github.com/kkkunny/Sim/compiler/cir"
	"github.com/kkkunny/Sim/compiler/config"
)

type Compiler struct {
	ir *cir.Builder
}

func NewCompiler(ir *cir.Builder) *Compiler {
	return &Compiler{ir: ir}
}

func (c *Compiler) Compile() error {
	// main函数
	mainF := c.ir.BuildFuncDecl("main", cir.I32, nil)
	mainF.Body = optional.Some(c.ir.BuildBlock())

	// c代码临时文件
	file, err := stlerr.ErrorWith(stlos.CreateTempFileWithCloser("sim_compile", "c"))
	if err != nil {
		return err
	}
	defer file.Close()

	// 输出
	c.ir.Output(file)
	err = stlerr.ErrorWrap(file.Sync())
	if err != nil {
		return err
	}

	// 查找c编译器
	execPath, err := stlerr.ErrorWith(exec.LookPath("clang"))
	if err != nil {
		execPath, err = stlerr.ErrorWith(exec.LookPath("gcc"))
		if err != nil {
			return err
		}
	}

	// 编译
	outPath := filepath.Join(config.WorkPath, "main.out")
	cmder := exec.Command(execPath, "-std=c11", "-I", config.IncludePath, file.Path(), "-o", outPath)
	cmder.Stdin, cmder.Stdout, cmder.Stderr = os.Stdin, os.Stdout, os.Stderr
	if err = stlerr.ErrorWrap(cmder.Run()); err != nil {
		var exitErr *exec.ExitError
		if !errors.As(err, &exitErr) {
			return err
		}
	}
	// defer os.Remove(outPath)

	// // 运行
	// cmder = exec.Command(outPath)
	// cmder.Stdin, cmder.Stdout, cmder.Stderr = os.Stdin, os.Stdout, os.Stderr
	// if err = stlerr.ErrorWrap(cmder.Run()); err != nil {
	// 	var exitErr *exec.ExitError
	// 	if !errors.As(err, &exitErr) {
	// 		return err
	// 	}
	// }
	return nil
}
