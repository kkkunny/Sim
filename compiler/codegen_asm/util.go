package codegen_asm

import (
	"io"
	"os"

	llvm2 "github.com/kkkunny/go-llvm"
	stlerror "github.com/kkkunny/stl/error"
	stlos "github.com/kkkunny/stl/os"

	"github.com/kkkunny/Sim/compiler/codegen_ir"
	"github.com/kkkunny/Sim/mir"
	"github.com/kkkunny/Sim/mir/output/llvm"
)

type tempFile struct {
	*os.File
	path stlos.FilePath
}

func (self *tempFile) Close() error {
	if err := self.File.Close(); err != nil {
		return err
	}
	return os.Remove(self.path.String())
}

// CodegenAsm 汇编代码生成
func CodegenAsm(target mir.Target, path stlos.FilePath) (io.ReadCloser, stlerror.Error) {
	irModule, err := codegen_ir.CodegenIr(target, path)
	if err != nil {
		return nil, err
	}
	outputer := llvm.NewLLVMOutputer()
	outputer.Codegen(irModule)
	tempPath, err := stlerror.ErrorWith(stlos.RandomTempFilePath("sim"))
	if err != nil {
		return nil, err
	}
	err = stlerror.ErrorWrap(outputer.Target().WriteASMToFile(outputer.Module(), tempPath.String(), llvm2.CodeOptLevelDefault, llvm2.RelocModePIC, llvm2.CodeModelDefault))
	if err != nil {
		return nil, err
	}
	file, err := stlerror.ErrorWith(os.Open(tempPath.String()))
	if err != nil {
		return nil, err
	}
	return &tempFile{
		File: file,
		path: tempPath,
	}, nil
}
