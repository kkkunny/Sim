package report

import (
	"errors"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"strings"

	"github.com/kkkunny/stl/enum"

	"github.com/kkkunny/Sim/compiler/reader"
)

type level string

var levelEnum = enum.New[struct {
	Warn  level `enum:"warn"`
	Error level `enum:"report"`
}]()

type report struct {
	Level    level
	Title    string
	Message  string
	Position reader.Position
}

func newReport(level level, title, message string, pos reader.Position) *report {
	return &report{
		Level:    level,
		Title:    title,
		Message:  message,
		Position: pos,
	}
}

// Format 格式化错误报告为字符串
func (r *report) Format() string {
	if r.Position.BeginRow != r.Position.EndRow {
		// TODO: 多行
		panic("todo")
		return ""
	}

	var buf strings.Builder

	// 错误级别和标题
	buf.WriteString(fmt.Sprintf("%s[%s]: %s\n",
		levelPrefix(r.Level),
		r.Title,
		r.Message))

	// 位置
	path := r.Position.Reader.Path()
	if wd, err := os.Getwd(); err == nil {
		if relpath, err := filepath.Rel(wd, path); err == nil {
			path = relpath
		}
	}
	buf.WriteString(fmt.Sprintf("  --> %s:%d:%d\n", path, r.Position.BeginRow, r.Position.BeginCol))
	buf.WriteString("   |\n")

	// 代码
	beginOffset := r.Position.BeginOffset - (r.Position.BeginCol - 1)
	endOffset := r.Position.EndOffset
	code, _ := ReadFromTo(r.Position.Reader, beginOffset, endOffset+1)
	for {
		c, _, err := r.Position.Reader.ReadRune()
		if err != nil && !errors.Is(err, io.EOF) {
			panic(err)
		} else if err != nil || c == '\n' {
			break
		}
		code += string(c)
	}
	// code = strings.ReplaceAll(strings.ReplaceAll(code, "\t", "    "), "\n", " ")
	buf.WriteString(fmt.Sprintf("%2d | %s\n", r.Position.BeginRow, code))

	// 箭头
	arrowPos := int(r.Position.BeginCol)
	if arrowPos > len(code) {
		arrowPos = len(code)
	}
	buf.WriteString("   |")
	buf.WriteString(strings.Repeat(" ", arrowPos))
	buf.WriteString("^")
	if r.Position.EndCol > r.Position.BeginCol {
		buf.WriteString(strings.Repeat("~", int(r.Position.EndCol-r.Position.BeginCol)))
	}
	buf.WriteString("\n")

	return buf.String()
}

// levelPrefix 返回级别的前缀颜色代码（这里暂时不使用颜色，后续可以扩展）
func levelPrefix(level level) string {
	switch level {
	case levelEnum.Error:
		return "report"
	case levelEnum.Warn:
		return "warning"
	default:
		return ""
	}
}
