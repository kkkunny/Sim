package report

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/gookit/color"
	"github.com/kkkunny/stl/enum"

	"github.com/kkkunny/Sim/compiler/reader"
)

type level string

var levelEnum = enum.New[struct {
	Warn  level `enum:"warn"`
	Error level `enum:"error"`
}]()

var levelColorMap = map[level]color.Style{
	levelEnum.Warn:  color.New(color.FgYellow),
	levelEnum.Error: color.New(color.FgRed),
}

var levelCodeColorMap = map[level]color.Style{
	levelEnum.Warn:  color.New(color.BgYellow),
	levelEnum.Error: color.New(color.BgRed),
}

func (l level) Color() color.Style {
	return levelColorMap[l]
}

func (l level) CodeColor() color.Style {
	return levelCodeColorMap[l]
}

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
	var buf strings.Builder

	// 错误级别和标题
	buf.WriteString(r.Level.Color().Sprintf("%s[%s]: %s",
		r.Level,
		r.Title,
		r.Message))
	buf.WriteString("\n")

	// 位置：相对路径逃逸工作目录（以 .. 开头）时保留绝对路径
	path := r.Position.Reader.Path()
	if wd, err := os.Getwd(); err == nil {
		if relpath, err := filepath.Rel(wd, path); err == nil && !strings.HasPrefix(relpath, "..") {
			path = relpath
		}
	}
	buf.WriteString(fmt.Sprintf("  --> %s:%d:%d\n", path, r.Position.BeginRow, r.Position.BeginCol))

	// 代码
	lines := r.sourceLines()
	buf.WriteString("   |\n")
	for i, line := range lines {
		buf.WriteString(fmt.Sprintf("%2d |", r.Position.BeginRow+int64(i)))
		begin, end := r.highlightRange(i, len(lines), len(line))
		buf.WriteString(line[:begin])
		buf.WriteString(r.Level.CodeColor().Sprintf("%s", line[begin:end]))
		buf.WriteString(line[end:])
		buf.WriteString("\n")
	}
	buf.WriteString("   |")

	return buf.String()
}

// sourceLines 读取诊断位置覆盖的完整源码行；读取失败时返回空，诊断渲染不能因此崩溃。
func (r *report) sourceLines() []string {
	pos := r.Position

	// 起始行首：BeginCol 从 1 开始计数；0 或非法列按当前位置处理
	beginOffset := pos.BeginOffset - (pos.BeginCol - 1)
	if pos.BeginCol <= 0 || beginOffset < 0 {
		beginOffset = pos.BeginOffset
	}
	if beginOffset < 0 {
		beginOffset = 0
	}
	endOffset := pos.EndOffset
	if endOffset < beginOffset {
		endOffset = beginOffset
	}

	code, err := ReadFromTo(pos.Reader, beginOffset, endOffset+1)
	if err != nil {
		return nil
	}
	// 补足到行尾
	for {
		c, _, err := pos.Reader.ReadRune()
		if err != nil || c == '\n' {
			break
		}
		code += string(c)
	}
	if code == "" {
		return nil
	}
	return strings.Split(code, "\n")
}

// highlightRange 计算第 i 行需要高亮的字节区间 [begin, end)，越界一律收缩到合法范围。
func (r *report) highlightRange(i, total, lineLen int) (int, int) {
	begin, end := 0, lineLen
	switch {
	case total <= 1:
		begin = clampIndex(int(r.Position.BeginCol)-1, 0, lineLen)
		end = clampIndex(int(r.Position.EndCol), begin, lineLen)
	case i == 0:
		begin = clampIndex(int(r.Position.BeginCol)-1, 0, lineLen)
	case i == total-1:
		end = clampIndex(int(r.Position.EndCol), 0, lineLen)
	}
	if end < begin {
		end = begin
	}
	return begin, end
}

func clampIndex(v, min, max int) int {
	if v < min {
		return min
	}
	if v > max {
		return max
	}
	return v
}
