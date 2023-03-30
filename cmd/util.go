package cmd

import (
	"errors"
	"fmt"
	"io"
	"math/rand"
	"os"
	"os/exec"
	"strings"
	"time"

	"github.com/kkkunny/Sim/src/compiler/analyse"
	"github.com/kkkunny/Sim/src/compiler/codegen"
	"github.com/kkkunny/Sim/src/compiler/mir/pass"
	"github.com/kkkunny/Sim/src/compiler/mirgen"
	"github.com/kkkunny/Sim/src/compiler/parse"
	"github.com/kkkunny/llvm"
	stlos "github.com/kkkunny/stl/os"
)

// LookupCmd 查找命令
func LookupCmd(cmd ...string) (string, *exec.Cmd) {
	for _, c := range cmd {
		p, err := exec.LookPath(c)
		if err != nil {
			continue
		}
		e := exec.Command(p)
		e.Stdin = os.Stdin
		e.Stdout = os.Stdout
		e.Stderr = os.Stderr
		return c, e
	}
	return "", nil
}

// RandomString 随机字符串
func RandomString(n uint8) string {
	rand.Seed(time.Now().Unix())
	var buf strings.Builder
	for i := uint8(0); i < n; i++ {
		n := rand.Intn(62)
		if n < 26 {
			buf.WriteByte('a' + byte(n))
		} else if n < 52 {
			buf.WriteByte('A' + byte(n) - 26)
		} else {
			buf.WriteByte('0' + byte(n) - 52)
		}
	}
	return buf.String()
}

// 编译到ir
func compileToLLVM(config *buildConfig, from stlos.Path) (llvm.Module, llvm.TargetMachine, error) {
	var err error
	ast, err := parse.Parse(from)
	if err != nil {
		return llvm.Module{}, llvm.TargetMachine{}, err
	}
	hirs, err := analyse.NewAnalyser().Analyse(ast)
	if err != nil {
		return llvm.Module{}, llvm.TargetMachine{}, err
	}
	mirs := mirgen.NewMirGenerator().Generate(*hirs)
	mirs = pass.WalkPass(mirs, pass.MUST)
	codegener, err := codegen.NewCodeGenerator("", config.Release)
	if err != nil {
		return llvm.Module{}, llvm.TargetMachine{}, err
	}
	module, machine := codegener.Codegen(*mirs)
	return module, machine, nil
}

// 输出ir
func outputIr(module llvm.Module, to stlos.Path) error {
	file, err := os.OpenFile(to.String(), os.O_WRONLY|os.O_CREATE|os.O_TRUNC, 0644)
	if err != nil {
		return err
	}
	defer file.Close()

	_, err = io.Copy(file, strings.NewReader(module.String()))
	return err
}

// 输出bc
func outputBc(module llvm.Module, to stlos.Path) error {
	file, err := os.OpenFile(to.String(), os.O_WRONLY|os.O_CREATE|os.O_TRUNC, 0644)
	if err != nil {
		return err
	}
	defer file.Close()

	return llvm.WriteBitcodeToFile(module, file)
}

// 输出汇编
func outputAsm(module llvm.Module, targetMachine llvm.TargetMachine, to stlos.Path) (stlos.Path, error) {
	if to == "" {
		for {
			to = stlos.Path(os.TempDir()).Join(stlos.Path(RandomString(6) + ".s"))
			if !to.IsExist() {
				break
			}
		}
	}

	if err := targetMachine.EmitToFile(module, to.String(), llvm.AssemblyFile); err != nil {
		return "", err
	}
	return to, nil
}

// 输出目标文件
func outputObject(from, to stlos.Path, links []stlos.Path) (stlos.Path, error) {
	if to == "" {
		for {
			to = stlos.Path(os.TempDir()).Join(stlos.Path(RandomString(6) + ".o"))
			if !to.IsExist() {
				break
			}
		}
	}

	_, compiler := LookupCmd("clang", "gcc")
	if compiler == nil {
		return "", errors.New("can not found a c compiler")
	}

	compiler.Args = append(compiler.Args, "-c", "-o", to.String(), from.String())
	for _, link := range links {
		compiler.Args = append(compiler.Args, link.String())
	}
	return to, compiler.Run()
}

// 输出动态库
func outputSharedLibrary(from, to stlos.Path, libraries, libraryPaths []string) (stlos.Path, error) {
	if to == "" {
		for {
			to = stlos.Path(os.TempDir()).Join("lib" + stlos.Path(RandomString(6)) + ".so")
			if !to.IsExist() {
				break
			}
		}
	}

	_, compiler := LookupCmd("clang", "gcc")
	if compiler == nil {
		return "", errors.New("can not found a c compiler")
	}

	compiler.Args = append(compiler.Args, "-fPIC", "-shared", "-o", to.String(), from.String())
	for _, l := range libraries {
		compiler.Args = append(compiler.Args, fmt.Sprintf("-l%s", l))
	}
	for _, L := range libraryPaths {
		compiler.Args = append(compiler.Args, fmt.Sprintf("-L%s", L))
	}
	return to, compiler.Run()
}

// 输出静态库
func outputStaticLibrary(from, to stlos.Path) (stlos.Path, error) {
	if to == "" {
		for {
			to = stlos.Path(os.TempDir()).Join("lib" + stlos.Path(RandomString(6)) + ".a")
			if !to.IsExist() {
				break
			}
		}
	}

	_, compiler := LookupCmd("ar")
	if compiler == nil {
		return "", errors.New("can not found a file for packaging")
	}

	compiler.Args = append(compiler.Args, "-rcs", to.String(), from.String())
	return to, compiler.Run()
}

// 输出可执行文件
func outputExecutableFile(from, to stlos.Path, static bool, libraries, libraryPaths []string) (stlos.Path, error) {
	if to == "" {
		for {
			to = stlos.Path(os.TempDir()).Join(stlos.Path(RandomString(6))) + ".out"
			if !to.IsExist() {
				break
			}
		}
	}

	_, compiler := LookupCmd("clang", "gcc")
	if compiler == nil {
		return "", errors.New("can not found a c compiler")
	}

	compiler.Args = append(compiler.Args, "-fPIC")
	if static {
		compiler.Args = append(compiler.Args, "-static")
	}
	compiler.Args = append(compiler.Args, "-o", to.String(), from.String())
	for _, l := range libraries {
		compiler.Args = append(compiler.Args, fmt.Sprintf("-l%s", l))
	}
	for _, L := range libraryPaths {
		compiler.Args = append(compiler.Args, fmt.Sprintf("-L%s", L))
	}
	return to, compiler.Run()
}
