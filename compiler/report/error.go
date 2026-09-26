package report

import (
	"errors"
	"fmt"
	"os"
	"runtime/debug"

	"github.com/gookit/color"
)

// ErrReported 表示诊断已经输出，调用方无需重复报告，直接以失败结束即可。
var ErrReported = errors.New("diagnostics already reported")

// RecoverICE 供各入口 defer 调用：把未预期的 panic 转换为内部编译器错误（ICE）诊断。
// 解析/分析阶段用于恢复的 panic（parseAbort/analyzeAbort）必须在此之前被 recovered。
func RecoverICE() {
	r := recover()
	if r == nil {
		return
	}
	fmt.Fprintf(os.Stderr, "internal compiler error: %v\n", r)
	fmt.Fprintln(os.Stderr, "this is a bug in the Sim compiler; please report it with the code that triggered it")
	fmt.Fprintln(os.Stderr, "--- stack trace ---")
	os.Stderr.Write(debug.Stack())
	os.Exit(1)
}

// SetupConsole 在 stderr 被重定向时关闭颜色，避免 ANSI 转义污染输出。
func SetupConsole() {
	fi, err := os.Stderr.Stat()
	if err != nil || fi.Mode()&os.ModeCharDevice == 0 {
		color.Disable()
	}
}
