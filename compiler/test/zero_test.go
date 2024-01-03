package test

import (
	"os"
	"path/filepath"
	"strings"
	"testing"

	stlos "github.com/kkkunny/stl/os"
	stltest "github.com/kkkunny/stl/test"

	"github.com/kkkunny/Sim/codegen_ir"
	"github.com/kkkunny/Sim/interpret"
	"github.com/kkkunny/Sim/mir"
)

func runPath(t *testing.T, path stlos.FilePath){
	infos, err := os.ReadDir(path.String())
	if err != nil{
		t.Fatal(err)
	}
	for _, info := range infos{
		if info.IsDir(){
			runPath(t, path.Join(info.Name()))
		}else if filepath.Ext(info.Name()) == ".sim"{
			t.Run(strings.TrimSuffix(info.Name(), ".sim"), func(t *testing.T) {
				module, err := codegen_ir.CodegenIr(mir.DefaultTarget(), path.Join(info.Name()))
				if err != nil{
					t.Fatal(err)
				}
				ret, err := interpret.Interpret(module)
				if err != nil{
					t.Fatal(err)
				}
				stltest.AssertEq(t, ret, 0)
			})
		}
	}
}

func TestExitWithZero(t *testing.T) {
	runPath(t, stlos.NewFilePath("."))
}
