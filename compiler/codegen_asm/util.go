package codegen_asm

import (
	"io"
	"os"

	"github.com/kkkunny/go-llvm"
	stlerror "github.com/kkkunny/stl/error"
	stlos "github.com/kkkunny/stl/os"

	"github.com/kkkunny/Sim/compiler/codegen_ir"
)

type tempFile struct {
	*os.File
	path stlos.FilePath
}

func (self *tempFile) Close() error {
	if err := self.File.Close(); err != nil {
		return err
	}
	return os.Remove(string(self.path))
}

// CodegenAsm 汇编代码生成
func CodegenAsm(target llvm.Target, path stlos.FilePath) (io.ReadCloser, error) {
	module, err := codegen_ir.CodegenIr(target, path)
	if err != nil {
		return nil, err
	}
	tempPath, err := stlerror.ErrorWith(stlos.RandomTempFilePath("sim"))
	if err != nil {
		return nil, err
	}
	err = stlerror.ErrorWrap(target.WriteASMToFile(module, string(tempPath), llvm.CodeOptLevelDefault, llvm.RelocModePIC, llvm.CodeModelDefault))
	if err != nil {
		return nil, err
	}
	file, err := stlerror.ErrorWith(os.Open(string(tempPath)))
	if err != nil {
		return nil, err
	}
	return &tempFile{
		File: file,
		path: tempPath,
	}, nil
}
