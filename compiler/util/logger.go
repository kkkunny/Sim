package util

import (
	stllog "github.com/kkkunny/stl/log"

	"github.com/kkkunny/Sim/compiler/config"
)

// Logger 日志管理器
var Logger = stllog.DefaultLogger(config.Debug)
