package compile

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"

	stlerr "github.com/kkkunny/stl/error"

	"github.com/kkkunny/Sim/compiler/codegen"
	"github.com/kkkunny/Sim/compiler/config"
	"github.com/kkkunny/Sim/compiler/hir/stmts"
)

type Compiler struct{}

func NewCompiler() *Compiler {
	return &Compiler{}
}

func (c *Compiler) Compile(pkg *stmts.Package) error {
	return c.compileMainPkg(pkg)
}

// 编译依赖包
func (c *Compiler) compileDepPkg(ctx *codegen.Context, pkg *stmts.Package) error {
	for _, depPkg := range pkg.Dependencies {
		err := c.compileDepPkg(ctx, depPkg)
		if err != nil {
			return err
		}
	}

	cacheDir := filepath.Join(pkg.Path, config.CacheDirName)
	err := stlerr.ErrorWrap(os.RemoveAll(cacheDir))
	if err != nil {
		return err
	}

	err = stlerr.ErrorWrap(os.Mkdir(cacheDir, 0755))
	if err != nil {
		return err
	}

	cir := codegen.New(ctx, pkg).Generate()

	headerPath := filepath.Join(cacheDir, pkg.Name+".h")
	err = stlerr.ErrorWrap(os.RemoveAll(headerPath))
	if err != nil {
		return err
	}
	file, err := stlerr.ErrorWith(os.Create(headerPath))
	if err != nil {
		return err
	}
	defer file.Close()

	fmt.Fprintf(file, "#include \"include/buildin.c\"\n")
	for _, depPkg := range pkg.Dependencies {
		relpath, _ := filepath.Rel(config.StdPkgPath, depPkg.Path)
		relpath = strings.ReplaceAll(relpath, string([]rune{filepath.Separator}), "")
		fmt.Fprintf(file, "#include \"std/%s/%s/%s.h\"\n", relpath, config.CacheDirName, depPkg.Name)
	}

	cir.OutputHeader(file)
	err = stlerr.ErrorWrap(file.Sync())
	if err != nil {
		return err
	}

	sourcePath := filepath.Join(cacheDir, pkg.Name+".c")
	err = stlerr.ErrorWrap(os.RemoveAll(sourcePath))
	if err != nil {
		return err
	}
	// file, err := stlerr.ErrorWith(stlos.CreateTempFileWithCloser("sim_compile", "c"))
	file, err = stlerr.ErrorWith(os.Create(sourcePath))
	if err != nil {
		return err
	}
	defer file.Close()

	cir.Output(file)
	err = stlerr.ErrorWrap(file.Sync())
	if err != nil {
		return err
	}

	// execPath, err := util.LookupCCompiler()
	// if err != nil {
	// 	return err
	// }
	// objPath := filepath.Join(cacheDir, pkg.Name+".o")
	// cmder := exec.Command(execPath, "-std=c11", "-I", config.StdPkgPath, "-c", sourcePath, "-o", objPath)
	// cmder.Stdin, cmder.Stdout, cmder.Stderr = os.Stdin, os.Stdout, os.Stderr
	// if err = stlerr.ErrorWrap(cmder.Run()); err != nil {
	// 	return err
	// }
	return nil
}

// 编译主包
func (c *Compiler) compileMainPkg(pkg *stmts.Package) error {
	ctx := codegen.NewContext()
	for _, depPkg := range pkg.Dependencies {
		err := c.compileDepPkg(ctx, depPkg)
		if err != nil {
			return err
		}
	}

	cacheDir := filepath.Join(pkg.Path, config.CacheDirName)
	err := stlerr.ErrorWrap(os.RemoveAll(cacheDir))
	if err != nil {
		return err
	}

	err = stlerr.ErrorWrap(os.Mkdir(cacheDir, 0755))
	if err != nil {
		return err
	}

	sourcePath := filepath.Join(pkg.Path, pkg.Name+".c")
	err = stlerr.ErrorWrap(os.RemoveAll(sourcePath))
	if err != nil {
		return err
	}
	// file, err := stlerr.ErrorWith(stlos.CreateTempFileWithCloser("sim_compile", "c"))
	file, err := stlerr.ErrorWith(os.Create(sourcePath))
	if err != nil {
		return err
	}
	defer file.Close()

	for _, depPkg := range pkg.Dependencies {
		relpath, _ := filepath.Rel(config.StdPkgPath, depPkg.Path)
		relpath = strings.ReplaceAll(relpath, string([]rune{filepath.Separator}), "")
		fmt.Fprintf(file, "#include \"std/%s/%s/%s.h\"\n", relpath, config.CacheDirName, depPkg.Name)
	}

	file.Write([]byte("#include \"include/buildin.c\"\n"))
	codegen.New(ctx, pkg).Generate().Output(file)
	err = stlerr.ErrorWrap(file.Sync())
	if err != nil {
		return err
	}

	// execPath, err := util.LookupCCompiler()
	// if err != nil {
	// 	return err
	// }
	// outPath := filepath.Join(config.WorkPath, "main.out")
	// cmder := exec.Command(execPath, "-std=c11", "-I", config.StdPkgPath, sourcePath, "-o", outPath)
	// cmder.Stdin, cmder.Stdout, cmder.Stderr = os.Stdin, os.Stdout, os.Stderr
	// if err = stlerr.ErrorWrap(cmder.Run()); err != nil {
	// 	return err
	// }
	return nil
}
