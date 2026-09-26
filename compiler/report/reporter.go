package report

import (
	"fmt"
	"io"
	"strings"

	"github.com/kkkunny/Sim/compiler/reader"
)

// Reporter 诊断报告器：累积错误与警告，由调用方决定何时输出、何时终止。
type Reporter struct {
	reports []*report
}

// NewReporter 创建错误报告器
func NewReporter() *Reporter {
	return &Reporter{}
}

func (r *Reporter) Warn(title, message string, pos reader.Position) {
	r.reports = append(r.reports, newReport(levelEnum.Warn, title, message, pos))
}

func (r *Reporter) Warnf(pos reader.Position, err ErrorType, args ...any) {
	r.Warn(string(err), formatError(err, args...), pos)
}

func (r *Reporter) Error(title, message string, pos reader.Position) {
	r.reports = append(r.reports, newReport(levelEnum.Error, title, message, pos))
}

func (r *Reporter) Errorf(pos reader.Position, err ErrorType, args ...any) {
	r.Error(string(err), formatError(err, args...), pos)
}

// HasErrors 检查是否有错误
func (r *Reporter) HasErrors() bool {
	for _, report := range r.reports {
		if report.Level == levelEnum.Error {
			return true
		}
	}
	return false
}

// HasWarns 检查是否有警告
func (r *Reporter) HasWarns() bool {
	for _, report := range r.reports {
		if report.Level == levelEnum.Warn {
			return true
		}
	}
	return false
}

// Emit 输出所有报告为字符串
func (r *Reporter) Emit() string {
	var sb strings.Builder

	for _, report := range r.reports {
		sb.WriteString(report.Format())
		sb.WriteString("\n")
	}

	// 统计信息
	errorCount := 0
	warningCount := 0
	for _, report := range r.reports {
		if report.Level == levelEnum.Error {
			errorCount++
		} else if report.Level == levelEnum.Warn {
			warningCount++
		}
	}

	if errorCount > 0 || warningCount > 0 {
		sb.WriteString("Summary: ")
		if errorCount > 0 {
			sb.WriteString(fmt.Sprintf("%d error(s)", errorCount))
		}
		if warningCount > 0 {
			if errorCount > 0 {
				sb.WriteString(", ")
			}
			sb.WriteString(fmt.Sprintf("%d warning(s)", warningCount))
		}
		sb.WriteString("\n")
	}
	if errorCount > 0 {
		sb.WriteString(fmt.Sprintf("error: aborting due to %d previous error(s)\n", errorCount))
	}

	return sb.String()
}

// Print 打印所有报告
func (r *Reporter) Print(w io.Writer) {
	fmt.Fprint(w, r.Emit())
}

func formatError(err ErrorType, args ...any) string {
	format, ok := errorFormats[err]
	if !ok {
		return string(err)
	}
	return fmt.Sprintf(format, args...)
}
