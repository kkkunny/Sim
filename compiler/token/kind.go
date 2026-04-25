package token

import (
	"reflect"

	stlmaps "github.com/kkkunny/stl/container/maps"
	"github.com/kkkunny/stl/enum"
)

// Kind token类型
type Kind uint8

var KindEnum = enum.New[struct {
	Illegal Kind `text:"illegal"`
	Eof     Kind `text:"eof"`
	Br      Kind `text:"br"`

	Ident   Kind `text:"ident"`
	Integer Kind `text:"integer"`

	Add Kind `text:"+"`
	Sub Kind `text:"-"`
	Mul Kind `text:"*"`
	Quo Kind `text:"/"`
	Rem Kind `text:"%"`

	And Kind `text:"&"`
	Or  Kind `text:"|"`
	Xor Kind `text:"^"`
	Not Kind `text:"!"`
	Shl Kind `text:"<<"`
	Shr Kind `text:">>"`

	Eq  Kind `text:"=="`
	Neq Kind `text:"!="`
	Lt  Kind `text:"<"`
	Lte Kind `text:"<="`
	Gt  Kind `text:">"`
	Gte Kind `text:">="`

	Assign    Kind `text:"="`
	AddAssign Kind `text:"+="`
	SubAssign Kind `text:"-="`
	MulAssign Kind `text:"*="`
	QuoAssign Kind `text:"/="`
	RemAssign Kind `text:"%="`
	AndAssign Kind `text:"&="`
	OrAssign  Kind `text:"|="`
	XorAssign Kind `text:"^="`
	ShlAssign Kind `text:"<<="`
	ShrAssign Kind `text:">>="`

	Lpa      Kind `text:"("`
	Rpa      Kind `text:")"`
	Lba      Kind `text:"["`
	Rba      Kind `text:"]"`
	Lbr      Kind `text:"{"`
	Rbr      Kind `text:"}"`
	Col      Kind `text:":"`
	Comma    Kind `text:","`
	Question Kind `text:"?"`
	Sem      Kind `text:";"`
	Arrow    Kind `text:"->"`

	Let    Kind `text:"let"`
	Return Kind `text:"return"`
	Mut    Kind `text:"mut"`
	As     Kind `text:"as"`
	True   Kind `text:"true"`
	False  Kind `text:"false"`
	If     Kind `text:"if"`
	Else   Kind `text:"else"`
	For    Kind `text:"for"`
	In     Kind `text:"in"`
}]()

var kind2Text = func() map[Kind]string {
	v := reflect.ValueOf(KindEnum)
	t := v.Type()
	res := make(map[Kind]string, v.NumField())
	for i := 0; i < v.NumField(); i++ {
		value := v.Field(i).Interface().(Kind)
		text, _ := t.Field(i).Tag.Lookup("text")
		res[value] = text
	}
	return res
}()

var keyword2Kind = func() map[string]Kind {
	res := make(map[string]Kind, len(kind2Text))
	keywordKey2Text := stlmaps.Filter(kind2Text, func(k Kind, v string) bool {
		return k >= KindEnum.Let
	})
	for k, v := range keywordKey2Text {
		res[v] = k
	}
	return res
}()

// Lookup 区分标识符和关键字
func Lookup(s string) Kind {
	keyword, ok := keyword2Kind[s]
	if ok {
		return keyword
	}
	return KindEnum.Ident
}

func (k Kind) String() string {
	return kind2Text[k]
}

func (k Kind) Priority() int {
	switch k {
	case KindEnum.Mul, KindEnum.Quo, KindEnum.Rem:
		return 9
	case KindEnum.Add, KindEnum.Sub:
		return 8
	case KindEnum.Shl, KindEnum.Shr:
		return 7
	case KindEnum.Lt, KindEnum.Lte, KindEnum.Gt, KindEnum.Gte:
		return 6
	case KindEnum.Eq, KindEnum.Neq:
		return 5
	case KindEnum.And:
		return 4
	case KindEnum.Xor:
		return 3
	case KindEnum.Or:
		return 2
	case KindEnum.Assign, KindEnum.AddAssign, KindEnum.SubAssign, KindEnum.MulAssign, KindEnum.QuoAssign,
		KindEnum.RemAssign, KindEnum.AndAssign, KindEnum.OrAssign, KindEnum.XorAssign, KindEnum.ShlAssign, KindEnum.ShrAssign:
		return 1
	default:
		return -1
	}
}
