package tests

import (
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/kkkunny/go-llvm"
	stlerror "github.com/kkkunny/stl/error"
	stlos "github.com/kkkunny/stl/os"
	stltest "github.com/kkkunny/stl/test"

	"github.com/kkkunny/Sim/compiler/codegen_ir"

	"github.com/kkkunny/Sim/compiler/interpret"
)

func runPath(t *testing.T, path stlos.FilePath) {
	infos, err := os.ReadDir(path.String())
	if err != nil {
		t.Fatal(err)
	}
	for _, info := range infos {
		if info.IsDir() {
			runPath(t, path.Join(info.Name()))
		} else if filepath.Ext(info.Name()) == ".sim" {
			t.Run(strings.TrimSuffix(info.Name(), ".sim"), func(t *testing.T) {
				module, err := codegen_ir.CodegenIr(stlerror.MustWith(llvm.NativeTarget()), path.Join(info.Name()))
				if err != nil {
					t.Fatal(err)
				}
				ret, err := interpret.Interpret(module)
				if err != nil {
					t.Fatal(err)
				}
				stltest.AssertEq(t, ret, 0)
			})
		}
	}
}

func TestExitWithZero(t *testing.T) {
	stlerror.Must(llvm.InitializeNativeTarget())
	runPath(t, stlos.NewFilePath("."))
}
