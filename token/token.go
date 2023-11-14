package token

import (
	"fmt"

	"github.com/kkkunny/Sim/reader"
)

// Token token
type Token struct {
	Position reader.Position // 位置
	Kind     Kind            // 种类
}

// Is 种类是否是
func (self Token) Is(k Kind) bool {
	return self.Kind == k
}

// Source 源码
func (self Token) Source() string {
	return self.Position.Text()
}

func (self Token) String() string {
	switch self.Kind {
	case ILLEGAL, IDENT, INTEGER, FLOAT:
		return fmt.Sprintf("%s(`%s`)", self.Kind, self.Source())
	default:
		return self.Kind.String()
	}
}
