package extern

import (
	"fmt"
	"os"

	"github.com/kkkunny/Sim/runtime/types"
)

// 内置函数，可显式调用

// Debug 输出字符串
func Debug(s *types.Str) {
	fmt.Println(s.String())
}

// Panic 运行时异常抛出
func Panic(s *types.Str){
	_, _ = fmt.Fprintf(os.Stderr, "panic: %s\n", s.String())
	os.Exit(1)
}
