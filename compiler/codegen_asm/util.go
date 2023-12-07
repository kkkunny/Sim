package codegen_asm

import (
	"io"
	"os"

	"github.com/kkkunny/go-llvm"
	stlerror "github.com/kkkunny/stl/error"
	stlos "github.com/kkkunny/stl/os"

	"github.com/kkkunny/Sim/codegen_ir"
)

type tempASMFile struct {
	path string
	file *os.File
}

func (self *tempASMFile) Read(p []byte) (n int, err error){
	return self.file.Read(p)
}

func (self *tempASMFile) Close() error{
	if err := stlerror.ErrorWrap(self.file.Close()); err != nil{
		return err
	}
	return stlerror.ErrorWrap(os.Remove(self.path))
}

// CodegenAsm 汇编生成
func CodegenAsm(path string, opt llvm.CodeOptLevel) (io.ReadCloser, stlerror.Error) {
	module, err := codegen_ir.CodegenIr(path)
	if err != nil{
		return nil, err
	}
	if err = stlerror.ErrorWrap(llvm.InitializeNativeAsmPrinter()); err != nil{
		return nil, err
	}
	target, ok := module.GetTarget()
	if !ok{
		panic("unreachable")
	}
	path, err = stlerror.ErrorWith(stlos.RandomTempFilePath("sim"))
	if err != nil{
		return nil, err
	}
	if err = stlerror.ErrorWrap(target.WriteASMToFile(module, path, opt, llvm.RelocModePIC, llvm.CodeModelDefault)); err != nil{
		return nil, err
	}
	file, err := stlerror.ErrorWith(os.Open(path))
	if err != nil{
		return nil, err
	}
	return &tempASMFile{
		path: path,
		file: file,
	}, nil
}
