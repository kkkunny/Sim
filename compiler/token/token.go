package token

import (
	"fmt"

	"github.com/kkkunny/Sim/compiler/reader"
)

// Token token
type Token struct {
	Position   reader.Position // 位置
	Kind       Kind            // 种类
	OriginText string          // 原始文本
}

// Is 种类是否是
func (t Token) Is(k Kind) bool {
	return t.Kind == k
}

func (t Token) String() string {
	switch t.Kind {
	case KindEnum.Illegal, KindEnum.Ident, KindEnum.Integer, KindEnum.Char, KindEnum.String:
		return fmt.Sprintf("%s\t[%s]%s", t.Position.String(), t.Kind, t.OriginText)
	default:
		return fmt.Sprintf("%s\t%s", t.Position.String(), t.Kind.String())
	}
}
