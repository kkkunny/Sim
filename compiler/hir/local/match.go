package local

import (
	stlslices "github.com/kkkunny/stl/container/slices"

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
	cond  values.Value
	cases []*MatchCase
	other *Block
}

func NewMatch(cond values.Value, cases ...*MatchCase) *Match {
	return &Match{
		cond:  cond,
		cases: cases,
	}
}

func (self *Match) local() {
	return
}

func (self *Match) Target() values.Value {
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
	other, ok := self.Other()
	if !ok {
		return BlockEndTypeNone
	}
	if stlslices.Empty(self.cases) {
		return other.BlockEndType()
	}
	minType := BlockEndTypeFuncRet
	for _, c := range self.cases {
		minType = min(minType, c.body.BlockEndType())
	}
	return min(minType, other.BlockEndType())
}

func (self *Match) Cond() values.Value {
	return self.cond
}
