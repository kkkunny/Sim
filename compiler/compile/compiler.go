package compile

import (
	"os"
	"os/exec"
	"path/filepath"

	"github.com/heimdalr/dag"
	"github.com/kkkunny/go-llvm/target"
	stlerr "github.com/kkkunny/stl/error"
	stlos "github.com/kkkunny/stl/os"

	"github.com/kkkunny/Sim/compiler/config"
	"github.com/kkkunny/Sim/compiler/hir/globals"
	"github.com/kkkunny/Sim/compiler/llgen"
	"github.com/kkkunny/Sim/compiler/util"
)

type Compiler struct {
	ctx *llgen.Context
	tm  *target.TargetMachine
}

func NewCompiler() *Compiler {
	return &Compiler{
		ctx: llgen.NewContext(),
	}
}

func (c *Compiler) Compile(pkg *globals.Package) error {
	if err := c.initTarget(); err != nil {
		return err
	}
	defer func() {
		stlerr.Must(c.tm.Close())
		stlerr.Must(c.ctx.Close())
	}()

	dagger := dag.NewDAG()
	pkg2Vertex := make(map[*globals.Package]string)
	var buildDAG func(pkg *globals.Package) (string, error)
	buildDAG = func(pkg *globals.Package) (string, error) {
		if v, ok := pkg2Vertex[pkg]; ok {
			return v, nil
		}
		v, err := stlerr.ErrorWith(dagger.AddVertex(pkg))
		if err != nil {
			return "", err
		}
		pkg2Vertex[pkg] = v
		for _, dep := range pkg.Dependencies {
			depV, err := buildDAG(dep)
			if err != nil {
				return "", err
			}
			err = stlerr.ErrorWrap(dagger.AddEdge(depV, v))
			if err != nil {
				return "", err
			}
		}
		return v, nil
	}
	_, err := buildDAG(pkg)
	if err != nil {
		return err
	}

	dagger.OrderedWalk(c)
	return nil
}

func (c *Compiler) Visit(v dag.Vertexer) {
	_, obj := v.Vertex()
	pkg := obj.(*globals.Package)
	if pkg.Name == "main" {
		err := c.compileMainPkg(pkg)
		if err != nil {
			panic(err)
		}
	} else {
		err := c.compileDepPkg(pkg)
		if err != nil {
			panic(err)
		}
	}
}

func (c *Compiler) initTarget() error {
	if err := target.InitNative(); err != nil {
		return err
	}
	native, err := target.NativeTarget()
	if err != nil {
		return err
	}
	tm, err := stlerr.ErrorWith(target.NewTargetMachine(
		native,
		target.DefaultTriple(),
		target.HostCPUName(),
		target.HostCPUFeatures(),
		target.OptAggressive,
		target.RelocPIC,
		target.CodeModelDefault,
	))
	if err != nil {
		return err
	}
	c.tm = tm
	return nil
}

// 编译依赖包
func (c *Compiler) compileDepPkg(pkg *globals.Package) error {
	// 创建目录
	cacheDir := filepath.Join(pkg.Path, config.CacheDir)
	err := stlerr.ErrorWrap(os.MkdirAll(cacheDir, 0755))
	if err != nil {
		return err
	}

	// 加锁
	locker, err := newCacheLock(pkg.Path)
	if err != nil {
		return err
	}
	defer func() {
		stlerr.Must(locker.Close())
	}()
	if err = locker.Lock(); err != nil {
		return err
	}

	// 即使命中缓存也要生成代码，以填充共享的 llgen.Context（idents/typeCache）
	g := llgen.New(c.ctx, pkg)
	defer func() {
		stlerr.Must(g.Close())
	}()
	c.tm.ApplyTo(g.Model())
	model := g.Generate()
	if valid, err := isCacheValid(pkg.Path, pkg.Name); err != nil {
		return err
	} else if valid {
		return nil
	}

	objPath := filepath.Join(cacheDir, pkg.Name+".o")
	if err = c.tm.EmitToFile(model, objPath, target.ObjectFile); err != nil {
		return err
	}
	llPath := filepath.Join(cacheDir, pkg.Name+".ll")
	if err = model.WriteToFile(llPath); err != nil {
		return err
	}
	return writeCacheBackend(cacheDir)
}

// 编译主包
func (c *Compiler) compileMainPkg(pkg *globals.Package) error {
	g := llgen.New(c.ctx, pkg)
	defer func() {
		stlerr.Must(g.Close())
	}()
	c.tm.ApplyTo(g.Model())
	model := g.Generate()

	file, err := stlerr.ErrorWith(stlos.CreateTempFileWithCloser("sim_compile", "o"))
	if err != nil {
		return err
	}
	defer func() {
		stlerr.Must(file.Close())
	}()
	if err = c.tm.EmitToFile(model, file.Path(), target.ObjectFile); err != nil {
		return err
	}

	execPath, err := util.LookupLinker()
	if err != nil {
		return err
	}
	outPath := filepath.Join(config.WorkPath, "main.out")
	args := []string{file.Path()}
	for _, depPkg := range pkg.Dependencies {
		args = append(args, filepath.Join(depPkg.Path, config.CacheDir, depPkg.Name+".o"))
	}
	args = append(args, "-lm", "-o", outPath)
	cmder := exec.Command(execPath, args...)
	cmder.Stdin, cmder.Stdout, cmder.Stderr = os.Stdin, os.Stdout, os.Stderr
	if err = stlerr.ErrorWrap(cmder.Run()); err != nil {
		return err
	}
	return nil
}
