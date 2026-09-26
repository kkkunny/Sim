package lex

import (
	"github.com/kkkunny/Sim/compiler/reader"
)

// Error 词法分析期间读取源码失败（IO 错误 / 非法 UTF-8 等）。
// 由调用方（Parser）捕获并转换为诊断。
type Error struct {
	Pos reader.Position
	Err error
}

func (e *Error) Error() string {
	return e.Err.Error()
}

func (e *Error) Unwrap() error {
	return e.Err
}
