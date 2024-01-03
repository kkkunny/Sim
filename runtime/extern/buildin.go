package extern

import (
	"fmt"

	"github.com/kkkunny/Sim/runtime/types"
)

// 内置函数，可显示调用

// Debug 输出字符串
func Debug(s types.Str) {
	fmt.Println(s)
}
