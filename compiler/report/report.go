package report

import (
	"errors"
	"fmt"
	"io"
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
	Error level `enum:"report"`
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
		levelPrefix(r.Level),
		r.Title,
		r.Message))
	buf.WriteString("\n")

	// 位置
	path := r.Position.Reader.Path()
	if wd, err := os.Getwd(); err == nil {
		if relpath, err := filepath.Rel(wd, path); err == nil {
			path = relpath
		}
	}
	buf.WriteString(fmt.Sprintf("  --> %s:%d:%d\n", path, r.Position.BeginRow, r.Position.BeginCol))

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
	buf.WriteString("   |\n")
	lines := strings.Split(code, "\n")
	for i, line := range lines {
		buf.WriteString(fmt.Sprintf("%2d |", r.Position.BeginRow+int64(i)))
		var beginPos, endPos int
		if len(lines) == 1 {
			beginPos, endPos = int(r.Position.BeginCol)-1, int(r.Position.EndCol)
		} else if i == 0 {
			beginPos = int(r.Position.BeginCol) - 1
			endPos = beginPos
		} else if i == len(lines)-1 {
			beginPos = int(r.Position.EndCol) - 1
			endPos = beginPos
		}
		prev, mid, next := line[:beginPos], line[beginPos:endPos], line[endPos:]
		if len(lines) == 1 {
			mid = r.Level.CodeColor().Sprintf(mid)
		} else if i == 0 {
			next = r.Level.CodeColor().Sprintf(next)
		} else if i == len(lines)-1 {
			prev = r.Level.CodeColor().Sprintf(prev)
		} else {
			prev = r.Level.CodeColor().Sprintf(prev)
			next = r.Level.CodeColor().Sprintf(next)
		}
		buf.WriteString(prev)
		buf.WriteString(mid)
		buf.WriteString(next)
		buf.WriteString("\n")
	}
	buf.WriteString("   |")

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
