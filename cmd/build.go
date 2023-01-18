package cmd

import (
	"errors"
	"fmt"
	stlos "github.com/kkkunny/stl/os"
	"github.com/kkkunny/stl/util"
	"github.com/spf13/cobra"
	"os"
)

type buildConfig struct {
	Target       stlos.Path   // 目标地址
	Output       stlos.Path   // 输出地址
	End          string       // 输出文件类型
	Release      bool         // 是否是release模式
	Linkages     []stlos.Path // 链接
	Libraries    []string     // 链接库
	LibraryPaths []string     // 链接库地址
}

func BuildCmd() *cobra.Command {
	var conf buildConfig
	cmd := &cobra.Command{
		Use:   "build",
		Short: "compiler a sim source file",
		Args: func(cmd *cobra.Command, args []string) error {
			if err := cobra.ExactArgs(1)(cmd, args); err != nil {
				return err
			}
			target := stlos.Path(args[0])
			if !target.IsExist() {
				return errors.New("expect a sim source file path")
			}
			target, err := target.GetAbsolute()
			if err != nil {
				return err
			}
			conf.Target = target
			return nil
		},
		RunE: func(cmd *cobra.Command, args []string) error {
			util.Must(build(conf))
			return nil
		},
	}
	// output path
	cmd.Flags().StringVarP((*string)(&conf.Output), "output", "o", "", "output path")
	// output file type
	cmd.Flags().StringVar(&conf.End, "end", "exe", "output file type, in (\"ir\", \"bc\", \"asm\", \"obj\", \"shared-lib\", \"static-lib\", \"exe\", \"static\")")
	// release
	cmd.Flags().BoolVar(&conf.Release, "release", false, "release mode")
	// lib
	cmd.Flags().StringSliceVarP(&conf.Libraries, "lib", "l", nil, "linkage extern library")
	cmd.Flags().StringSliceVarP(&conf.LibraryPaths, "lib_path", "L", nil, "library path")
	return cmd
}

func build(conf buildConfig) error {
	// 输出类型
	switch conf.End {
	case "ir", "bc", "asm", "obj", "shared-lib", "static-lib", "exe", "static":
	default:
		return fmt.Errorf("unknwon output file type")
	}

	// 输出地址
	if conf.Output == "" {
		if !conf.Target.IsDir() {
			switch conf.End {
			case "ir":
				conf.Output = conf.Target.WithExtension("ll")
			case "bc":
				conf.Output = conf.Target.WithExtension("bc")
			case "asm":
				conf.Output = conf.Target.WithExtension("s")
			case "obj":
				conf.Output = conf.Target.WithExtension("o")
			case "shared-lib":
				conf.Output = conf.Target.GetParent().Join("lib" + conf.Target.GetBase().WithExtension("so"))
			case "static-lib":
				conf.Output = conf.Target.GetParent().Join("lib" + conf.Target.GetBase().WithExtension("a"))
			case "exe", "static":
				conf.Output = conf.Target.WithExtension("out")
			}
		} else {
			switch conf.End {
			case "ir":
				conf.Output = conf.Target.Join(conf.Target.GetBase().WithExtension("ll"))
			case "bc":
				conf.Output = conf.Target.Join(conf.Target.GetBase().WithExtension("bc"))
			case "asm":
				conf.Output = conf.Target.Join(conf.Target.GetBase().WithExtension("s"))
			case "obj":
				conf.Output = conf.Target.Join(conf.Target.GetBase().WithExtension("o"))
			case "shared-lib":
				conf.Output = conf.Target.Join("lib" + conf.Target.GetBase().WithExtension("so"))
			case "static-lib":
				conf.Output = conf.Target.Join("lib" + conf.Target.GetBase().WithExtension("a"))
			case "exe", "static":
				conf.Output = conf.Target.Join(conf.Target.GetBase().WithExtension("out"))
			}
		}
	}

	// 编译
	module, targetMachine, err := compileToLLVM(&conf, conf.Target)
	if err != nil {
		return err
	}

	// llvm
	if conf.End == "ir" {
		return outputIr(module, conf.Output)
	}

	// bc
	if conf.End == "bc" {
		return outputBc(module, conf.Output)
	}

	// 汇编
	var asmPath stlos.Path
	if conf.End == "asm" {
		asmPath = conf.Output
		_, err = outputAsm(module, targetMachine, conf.Output)
		return err
	} else {
		asmPath, err = outputAsm(module, targetMachine, "")
		if err != nil {
			return err
		}
	}
	defer os.Remove(asmPath.String())

	// 链接
	var objectPath stlos.Path
	if conf.End == "obj" {
		objectPath = conf.Output
		_, err = outputObject(asmPath, conf.Output, conf.Linkages)
		return err
	} else {
		objectPath, err = outputObject(asmPath, "", conf.Linkages)
		if err != nil {
			return err
		}
	}
	defer os.Remove(objectPath.String())

	switch conf.End {
	case "shared-lib":
		// 动态库
		_, err = outputSharedLibrary(objectPath, conf.Output, conf.Libraries, conf.LibraryPaths)
		return err
	case "static-lib":
		// 静态库
		_, err = outputStaticLibrary(objectPath, conf.Output)
		return err
	case "exe":
		// 可执行文件
		_, err = outputExecutableFile(objectPath, conf.Output, false, conf.Libraries, conf.LibraryPaths)
		return err
	case "static":
		// 静态可执行文件
		_, err = outputExecutableFile(objectPath, conf.Output, true, conf.Libraries, conf.LibraryPaths)
		return err
	default:
		panic("")
	}
}
