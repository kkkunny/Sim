package types

import (
	"github.com/kkkunny/Sim/compiler/hir"
)

// Invalid 类型分析失败时的占位类型：与任意类型比较都视为相等，用于抑制级联报错。
// 存在诊断错误时不会进入代码生成。
var Invalid = InvalidType{}

type InvalidType struct{}

func (t InvalidType) Print(p *hir.Printer) {
	p.WriteString(t.String())
}

func (InvalidType) String() string {
	return "<invalid>"
}

func (InvalidType) Equal(hir.Type) bool {
	return true
}
