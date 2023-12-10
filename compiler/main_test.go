package main

import (
	"path/filepath"
	"testing"

	stlerror "github.com/kkkunny/stl/error"

	"github.com/kkkunny/Sim/codegen_ir"
	"github.com/kkkunny/Sim/mir"
)

func TestDebug(t *testing.T) {
	path := stlerror.MustWith(filepath.Abs("example/main.sim"))
	stlerror.MustWith(codegen_ir.CodegenIr(mir.DefaultTarget(), path))
}