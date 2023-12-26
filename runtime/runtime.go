package runtime

import (
	"fmt"
)

// SimRuntimeStrEqStr 字符串比较
func SimRuntimeStrEqStr(l, r Str) Bool {
	return NewBool(l.Value() == r.Value())
}

// SimRuntimeDebug 输出字符串
func SimRuntimeDebug(s Str) {
	fmt.Println(s)
}

// SimRuntimeCheckNull 检查空指针
func SimRuntimeCheckNull(p Ptr) Ptr {
	if p == nil {
		panic("空指针")
	}
	return p
}
