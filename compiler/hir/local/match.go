package local

import (
	stlslices "github.com/kkkunny/stl/container/slices"
	"github.com/kkkunny/stl/list"

	"github.com/kkkunny/Sim/compiler/hir/values"
)

type MatchCase struct {
	name string
	elem *values.VarDecl
	body *Block
}

func NewMatchCase(name string, e *values.VarDecl) *MatchCase {
	return &MatchCase{
		name: name,
		elem: e,
	}
}

func (self *MatchCase) Body() *Block {
	return self.body
}

func (self *MatchCase) Elem() *values.VarDecl {
	return self.elem
}

func (self *MatchCase) Name() string {
	return self.name
}

// Match 枚举匹配
type Match struct {
	pos    *list.Element[Local]
	target values.Value
	cases  []*MatchCase
	other  *Block
}

func NewMatch(p *Block, t values.Value, cases ...*MatchCase) *Match {
	for _, c := range cases {
		c.body = NewBlock(p)
	}
	return &Match{
		target: t,
		cases:  cases,
	}
}

func (self *Match) setPosition(pos *list.Element[Local]) {
	self.pos = pos
}

func (self *Match) position() (*list.Element[Local], bool) {
	return self.pos, self.pos != nil
}

func (self *Match) Target() values.Value {
	return self.target
}

func (self *Match) Cases() []*MatchCase {
	return self.cases
}

func (self *Match) NewOther(p *Block) *Block {
	self.other = NewBlock(p)
	return self.other
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
