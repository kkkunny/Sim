package output

import "github.com/kkkunny/Sim/mir"

type Output interface {
	Codegen(module *mir.Module)
}
