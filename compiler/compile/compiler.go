package compile

import (
	"os"
	"os/exec"
	"path/filepath"

	"github.com/heimdalr/dag"
	"github.com/kkkunny/go-llvm/target"
	stlerr "github.com/kkkunny/stl/error"
	stlos "github.com/kkkunny/stl/os"

	"github.com/kkkunny/Sim/compiler/codegen"
	"github.com/kkkunny/Sim/compiler/config"
	"github.com/kkkunny/Sim/compiler/hir/globals"
	"github.com/kkkunny/Sim/compiler/util"
)

type Compiler struct {
	ctx *codegen.Context

	err error // 遍历中首次出现的错误
}

func NewCompiler() *Compiler {
	return &Compiler{
		ctx: codegen.NewContext(),
	}
}

func (c *Compiler) Compile(pkg *globals.Package) error {
	defer c.ctx.Close()

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
		// 依赖边去重：Dependencies 理论上已去重，这里防御重复 import 等回归
		//（heimdalr/dag 对重复边返回 EdgeDuplicateError，错误信息只有裸 UUID）
		addedEdges := make(map[string]bool, len(pkg.Dependencies))
		for _, dep := range pkg.Dependencies {
			depV, err := buildDAG(dep)
			if err != nil {
				return "", err
			}
			if addedEdges[depV] {
				continue
			}
			addedEdges[depV] = true
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
	return c.err
}

func (c *Compiler) Visit(v dag.Vertexer) {
	if c.err != nil {
		return
	}
	_, obj := v.Vertex()
	pkg := obj.(*globals.Package)
	if pkg.Name == "main" {
		c.err = c.compileMainPkg(pkg)
	} else {
		c.err = c.compileDepPkg(pkg)
	}
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
	defer locker.Close()
	if err = locker.Lock(); err != nil {
		return err
	}

	// 即使命中缓存也要生成代码，以填充共享的 codegen.Context（idents/typeCache）
	gen := codegen.New(c.ctx, pkg)
	defer gen.Close()
	gen.Generate()

	if valid, err := isCacheValid(pkg); err != nil {
		return err
	} else if valid {
		return nil
	}

	objPath := filepath.Join(cacheDir, pkg.Name+".o")
	err = c.ctx.TargetMachine().EmitToFile(gen.Module(), objPath, target.ObjectFile)
	if err != nil {
		return err
	}
	// 清理旧 C 后端遗留的头文件
	_ = os.Remove(filepath.Join(cacheDir, pkg.Name+".h"))
	return writeBackendMarker(cacheDir)
}

// 编译主包
func (c *Compiler) compileMainPkg(pkg *globals.Package) error {
	gen := codegen.New(c.ctx, pkg)
	defer gen.Close()
	gen.Generate()

	objPath, objFile, err := stlerr.ErrorWith2(stlos.CreateTempFile("sim_main", "o"))
	if err != nil {
		return err
	}
	defer os.Remove(objPath)

	err = c.ctx.TargetMachine().EmitToFile(gen.Module(), objPath, target.ObjectFile)
	if err != nil {
		objFile.Close()
		return err
	}
	if err = stlerr.ErrorWrap(objFile.Close()); err != nil {
		return err
	}

	// 收集所有（含传递）依赖包的目标文件
	var depObjs []string
	seen := make(map[*globals.Package]bool)
	var collectDeps func(p *globals.Package)
	collectDeps = func(p *globals.Package) {
		for _, dep := range p.Dependencies {
			if seen[dep] {
				continue
			}
			seen[dep] = true
			depObjs = append(depObjs, filepath.Join(dep.Path, config.CacheDir, dep.Name+".o"))
			collectDeps(dep)
		}
	}
	collectDeps(pkg)

	execPath, err := util.LookupCCompiler()
	if err != nil {
		return err
	}
	outPath := filepath.Join(config.WorkPath, "main.out")
	args := append([]string{objPath}, depObjs...)
	args = append(args, "-lm", "-o", outPath)
	cmder := exec.Command(execPath, args...)
	cmder.Stdin, cmder.Stdout, cmder.Stderr = os.Stdin, os.Stdout, os.Stderr
	return stlerr.ErrorWrap(cmder.Run())
}
