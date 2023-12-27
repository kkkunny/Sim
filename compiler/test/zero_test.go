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

func TestExitWithZero(t *testing.T) {
	testdir := stlos.NewFilePath(".")
	infos, err := os.ReadDir(testdir.String())
	if err != nil{
		t.Fatal(err)
	}
	for _, info := range infos{
		if info.IsDir() || filepath.Ext(info.Name()) != ".sim"{
			continue
		}
		path := testdir.Join(info.Name())
		t.Run(strings.TrimSuffix(info.Name(), ".sim"), func(t *testing.T) {
			module, err := codegen_ir.CodegenIr(mir.DefaultTarget(), path)
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
