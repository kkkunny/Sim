package config

import (
	"github.com/kkkunny/go-llvm"

	"github.com/kkkunny/Sim/util"
)

func init() {
	util.Logger.Infof(0, "LLVM VERSION: %s", llvm.Version)
}
