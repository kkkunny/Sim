package values

import "github.com/kkkunny/Sim/compiler/hir/types"

// VarDecl 变量声明
type VarDecl struct {
	mut  bool
	name string
	typ  types.Type
}

func NewVarDecl(mut bool, name string, t types.Type) *VarDecl {
	return &VarDecl{
		mut:  mut,
		name: name,
		typ:  t,
	}
}

func (self *VarDecl) Type() types.Type {
	return self.typ
}

func (self *VarDecl) Mutable() bool {
	return self.mut
}

func (self *VarDecl) GetName() (string, bool) {
	return self.name, self.name != "_"
}

func (self *VarDecl) Storable() bool {
	return true
}
