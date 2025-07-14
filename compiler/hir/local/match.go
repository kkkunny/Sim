package local

import (
	stlslices "github.com/kkkunny/stl/container/slices"
	stlval "github.com/kkkunny/stl/value"

	"github.com/kkkunny/Sim/compiler/hir"
	"github.com/kkkunny/Sim/compiler/hir/types"
	"github.com/kkkunny/Sim/compiler/hir/values"
)

type MatchCase struct {
	name string
	val  values.VarDecl
	body *Block
}

func NewMatchCase(name string, val values.VarDecl, body *Block) *MatchCase {
	return &MatchCase{
		name: name,
		val:  val,
		body: body,
	}
}

func (self *MatchCase) Body() *Block {
	return self.body
}

func (self *MatchCase) Var() (values.VarDecl, bool) {
	return self.val, self.val != nil
}

func (self *MatchCase) Name() string {
	return self.name
}

// Match 枚举匹配
type Match struct {
	cond  hir.Value
	cases []*MatchCase
	other *Block
}

func NewMatch(cond hir.Value, cases ...*MatchCase) *Match {
	return &Match{
		cond:  cond,
		cases: cases,
	}
}

func (self *Match) Local() {
	return
}

func (self *Match) Target() hir.Value {
	return self.cond
}

func (self *Match) Cases() []*MatchCase {
	return self.cases
}

func (self *Match) SetOther(block *Block) {
	self.other = block
}

func (self *Match) Other() (*Block, bool) {
	return self.other, self.other != nil
}

func (self *Match) BlockEndType() BlockEndType {
	if stlslices.Empty(self.cases) && self.other == nil {
		return BlockEndTypeNone
	} else if stlslices.Empty(self.cases) {
		return self.other.BlockEndType()
	}

	minType := BlockEndTypeFuncRet
	for _, c := range self.cases {
		minType = min(minType, c.body.BlockEndType())
	}

	et := stlval.IgnoreWith(types.As[types.EnumType](self.cond.Type()))
	if int(et.EnumFields().Length()) == len(self.cases) {
		return minType
	} else if self.other != nil {
		return min(minType, self.other.BlockEndType())
	}

	return BlockEndTypeNone
}

func (self *Match) Cond() hir.Value {
	return self.cond
}
