package compile

import (
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strings"

	stlerr "github.com/kkkunny/stl/error"
	stlos "github.com/kkkunny/stl/os"

	"github.com/kkkunny/Sim/compiler/codegen"
	"github.com/kkkunny/Sim/compiler/config"
	"github.com/kkkunny/Sim/compiler/hir/stmts"
	"github.com/kkkunny/Sim/compiler/util"
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
	hfile, err := stlerr.ErrorWith(os.Create(headerPath))
	if err != nil {
		return err
	}
	defer hfile.Close()

	relpath, _ := filepath.Rel(config.SimRootPath, pkg.Path)
	headerName := "_SIM_" + strings.ReplaceAll(relpath, string([]rune{filepath.Separator}), "_") + "_H"
	fmt.Fprintf(hfile, "#ifndef %s\n", headerName)
	fmt.Fprintf(hfile, "#define %s 1 \n\n", headerName)
	fmt.Fprintf(hfile, "#include \"include/buildin.h\"\n")
	for _, depPkg := range pkg.Dependencies {
		relpath, _ = filepath.Rel(config.StdPkgPath, depPkg.Path)
		relpath = strings.ReplaceAll(relpath, string([]rune{filepath.Separator}), "")
		fmt.Fprintf(hfile, "#include \"std/%s/%s/%s.h\"\n", relpath, config.CacheDirName, depPkg.Name)
	}

	cir.OutputHeader(hfile)
	fmt.Fprintf(hfile, "\n#endif\n")
	err = stlerr.ErrorWrap(hfile.Sync())
	if err != nil {
		return err
	}

	file, err := stlerr.ErrorWith(stlos.CreateTempFileWithCloser("sim_compile", "c"))
	if err != nil {
		return err
	}
	defer file.Close()

	fmt.Fprintf(file, "#include \"include/buildin.h\"\n")

	cir.Output(file)
	err = stlerr.ErrorWrap(file.Sync())
	if err != nil {
		return err
	}

	execPath, err := util.LookupCCompiler()
	if err != nil {
		return err
	}
	objPath := filepath.Join(cacheDir, pkg.Name+".o")
	cmder := exec.Command(execPath, "-std=c11", "-I", config.SimRootPath, "-c", file.Path(), "-o", objPath)
	cmder.Stdin, cmder.Stdout, cmder.Stderr = os.Stdin, os.Stdout, os.Stderr
	if err = stlerr.ErrorWrap(cmder.Run()); err != nil {
		return err
	}
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

	file, err := stlerr.ErrorWith(stlos.CreateTempFileWithCloser("sim_compile", "c"))
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

	execPath, err := util.LookupCCompiler()
	if err != nil {
		return err
	}
	outPath := filepath.Join(config.WorkPath, "main.out")
	args := []string{"-std=c11", "-I", config.SimRootPath, file.Path()}
	for _, depPkg := range pkg.Dependencies {
		args = append(args, filepath.Join(depPkg.Path, config.CacheDirName, depPkg.Name+".o"))
	}
	args = append(args, "-o", outPath)
	cmder := exec.Command(execPath, args...)
	cmder.Stdin, cmder.Stdout, cmder.Stderr = os.Stdin, os.Stdout, os.Stderr
	if err = stlerr.ErrorWrap(cmder.Run()); err != nil {
		return err
	}
	return nil
}
