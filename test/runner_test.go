// Package test 是 Sim 编译器的用例驱动器：test/cases/*.sim 每个文件即一个用例——
// 首行注释声明动作（可带 key=value 标志），紧随其后的连续注释块是期望输出，
// 注释块之下是 Sim 代码（注释对编译器不可见，用例文件本身即合法源码）。
//
// 运行：go test ./test/...
// 重写期望：go test ./test/... -update（迭代至稳定，随后请 git diff 审查）
package test

import (
	"bytes"
	"flag"
	"fmt"
	"os"
	"path/filepath"
	"regexp"
	"sort"
	"strings"
	"testing"

	"github.com/gookit/color"

	"github.com/kkkunny/Sim/compiler/analyze"
	"github.com/kkkunny/Sim/compiler/ast"
	"github.com/kkkunny/Sim/compiler/config"
	"github.com/kkkunny/Sim/compiler/hir"
	"github.com/kkkunny/Sim/compiler/lex"
	"github.com/kkkunny/Sim/compiler/parse"
	"github.com/kkkunny/Sim/compiler/reader"
	"github.com/kkkunny/Sim/compiler/report"
	"github.com/kkkunny/Sim/compiler/token"
)

var update = flag.Bool("update", false, "用实际输出重写用例的期望注释块")

// 用例首行允许的标志
var caseFlags = map[string]bool{
	"known-bug": true, // 已知缺陷快照（值对应 docs/known-issues.md 条目编号）
	"skip":      true, // 跳过（值为不含空格的短标识/原因）
}

var repoRoot string

func TestMain(m *testing.M) {
	flag.Parse()
	// 诊断渲染必须无颜色，期望块才能稳定
	color.Disable()
	root, err := findRepoRoot()
	if err != nil {
		fmt.Fprintln(os.Stderr, "error:", err)
		os.Exit(1)
	}
	repoRoot = root
	config.SetSimRoot(root)
	os.Exit(m.Run())
}

func findRepoRoot() (string, error) {
	dir, err := os.Getwd()
	if err != nil {
		return "", err
	}
	for {
		if info, err := os.Stat(filepath.Join(dir, "std", "buildin")); err == nil && info.IsDir() {
			return dir, nil
		}
		parent := filepath.Dir(dir)
		if parent == dir {
			return "", fmt.Errorf("向上未找到仓库根（含 std/buildin 的目录）")
		}
		dir = parent
	}
}

// testCase 一个用例文件的头部与元信息
type testCase struct {
	path     string
	action   string
	flags    map[string]string
	expected string // 期望输出（已剥注释前缀）
	codeFrom int    // 代码起始行（0 基）；介于其间的行均为期望块
}

func parseCase(path string) (*testCase, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		return nil, err
	}
	lines := strings.Split(strings.ReplaceAll(string(data), "\r\n", "\n"), "\n")
	if len(lines) == 0 || !strings.HasPrefix(lines[0], "//") {
		return nil, fmt.Errorf("首行必须是 // <action> [key=value ...]")
	}
	fields := strings.Fields(strings.TrimPrefix(lines[0], "//"))
	if len(fields) == 0 {
		return nil, fmt.Errorf("首行缺少 action")
	}
	tc := &testCase{path: path, action: fields[0], flags: map[string]string{}}
	for _, field := range fields[1:] {
		key, value, ok := strings.Cut(field, "=")
		if !ok || !caseFlags[key] {
			return nil, fmt.Errorf("首行参数 %q 无效（支持 key=value：known-bug/skip）", field)
		}
		tc.flags[key] = value
	}
	var expected []string
	tc.codeFrom = 1
	for ; tc.codeFrom < len(lines); tc.codeFrom++ {
		line := lines[tc.codeFrom]
		if !strings.HasPrefix(line, "//") {
			break
		}
		line = strings.TrimPrefix(line, "//")
		line = strings.TrimPrefix(line, " ")
		expected = append(expected, line)
	}
	tc.expected = strings.Join(expected, "\n")
	return tc, nil
}

func TestCases(t *testing.T) {
	paths, err := filepath.Glob(filepath.Join("cases", "*.sim"))
	if err != nil {
		t.Fatal(err)
	}
	if len(paths) == 0 {
		t.Fatal("cases/ 下没有用例")
	}
	sort.Strings(paths)
	for _, path := range paths {
		path := path
		t.Run(strings.TrimSuffix(filepath.Base(path), ".sim"), func(t *testing.T) {
			t.Parallel()
			tc, err := parseCase(path)
			if err != nil {
				t.Fatalf("用例格式错误: %v", err)
			}
			if reason, ok := tc.flags["skip"]; ok {
				t.Skipf("跳过：%s", reason)
			}
			if *update {
				if err := updateCase(path); err != nil {
					t.Fatalf("更新期望失败: %v", err)
				}
				return
			}
			if id, ok := tc.flags["known-bug"]; ok {
				if err := checkKnownBug(id); err != nil {
					t.Error(err)
				}
			}
			absPath, err := filepath.Abs(path)
			if err != nil {
				t.Fatal(err)
			}
			actual, err := runCase(tc, absPath)
			if err != nil {
				t.Fatalf("用例运行失败: %v", err)
			}
			expected, actual := normalize(tc.expected), normalize(actual)
			if expected == actual {
				return
			}
			msg := fmt.Sprintf("输出与期望不一致：\n%s", diffText(expected, actual))
			if id, ok := tc.flags["known-bug"]; ok {
				msg = fmt.Sprintf("已知缺陷 %s 的行为发生变化（修复或恶化）：请同步更新用例期望与 docs/known-issues.md，修好后移除 known-bug 标志\n%s",
					id, diffText(expected, actual))
			}
			t.Error(msg)
		})
	}
}

// runCase 执行一个用例并返回归一化前的实际输出
func runCase(tc *testCase, path string) (out string, err error) {
	defer func() {
		if r := recover(); r != nil {
			err = fmt.Errorf("panic: %v", r)
		}
	}()
	switch tc.action {
	case "lex":
		src, err := os.ReadFile(path)
		if err != nil {
			return "", err
		}
		return runLex(path, codeSource(src, tc.codeFrom))
	case "parse":
		src, err := os.ReadFile(path)
		if err != nil {
			return "", err
		}
		return runParse(path, src)
	case "analyze":
		return runAnalyze(path)
	default:
		return "", fmt.Errorf("action %q 尚未支持（当前支持 lex/parse/analyze，compile/run 见 plans P1）", tc.action)
	}
}

func runLex(path string, src []byte) (string, error) {
	lexer := lex.New(reader.NewFile(path, bytes.NewReader(src)))
	var sb strings.Builder
	for tok := lexer.Scan(); !tok.Is(token.KindEnum.Eof); tok = lexer.Scan() {
		sb.WriteString(tok.String())
		sb.WriteByte('\n')
	}
	return sb.String(), nil
}

// codeSource 取期望块之后的代码（跳过紧随块的空行）。
// 词法用例只对代码做词法：期望块本身是注释、每行都会产出 Br token，
// 若连块一起扫会与块长度自指（越长 Br 越多，永不收敛）；位置也相对代码块从 1 计。
func codeSource(data []byte, codeFrom int) []byte {
	lines := strings.Split(strings.ReplaceAll(string(data), "\r\n", "\n"), "\n")
	for codeFrom < len(lines) && strings.TrimSpace(lines[codeFrom]) == "" {
		codeFrom++
	}
	if codeFrom >= len(lines) {
		return nil
	}
	return []byte(strings.Join(lines[codeFrom:], "\n"))
}

func runParse(path string, src []byte) (string, error) {
	reporter := report.NewReporter()
	file := parse.New(lex.New(reader.NewFile(path, bytes.NewReader(src))), reporter).Parse()
	if reporter.HasErrors() {
		return reporter.Emit(), nil
	}
	var sb strings.Builder
	ast.Print(&sb, file)
	return sb.String(), nil
}

func runAnalyze(path string) (string, error) {
	reporter := report.NewReporter()
	pkg, err := analyze.AnalyzeWith(path, reporter)
	if reporter.HasErrors() {
		return reporter.Emit(), nil
	} else if err != nil {
		return "", err
	}
	var sb strings.Builder
	hir.Print(&sb, pkg)
	if reporter.HasWarns() {
		sb.WriteString(reporter.Emit())
	}
	return sb.String(), nil
}

var ansiRE = regexp.MustCompile("\x1b\\[[0-9;]*m")

func normalize(s string) string {
	s = ansiRE.ReplaceAllString(s, "")
	s = strings.ReplaceAll(s, "\r\n", "\n")
	return strings.TrimRight(s, "\n")
}

// updateCase 用实际输出重写期望块。期望块长度会影响代码行号（进而影响诊断位置），
// 故需迭代至稳定：块长度变化 → 行号变化 → 再写一次。
func updateCase(path string) error {
	for range 5 {
		tc, err := parseCase(path)
		if err != nil {
			return err
		}
		absPath, err := filepath.Abs(path)
		if err != nil {
			return err
		}
		actual, err := runCase(tc, absPath)
		if err != nil {
			return err
		}
		if normalize(actual) == normalize(tc.expected) {
			return nil
		}
		data, err := os.ReadFile(path)
		if err != nil {
			return err
		}
		lines := strings.Split(strings.ReplaceAll(string(data), "\r\n", "\n"), "\n")
		out := make([]string, 0, len(lines))
		out = append(out, lines[0])
		if actual = normalize(actual); actual != "" {
			for _, line := range strings.Split(actual, "\n") {
				if line == "" {
					out = append(out, "//")
				} else {
					out = append(out, "// "+line)
				}
			}
		}
		out = append(out, lines[tc.codeFrom:]...)
		if err := os.WriteFile(path, []byte(strings.Join(out, "\n")), 0644); err != nil {
			return err
		}
	}
	return fmt.Errorf("期望块更新后未收敛（行号与块长度互相影响，5 次迭代仍在变化）")
}

func checkKnownBug(id string) error {
	data, err := os.ReadFile(filepath.Join(repoRoot, "docs", "known-issues.md"))
	if err != nil {
		return err
	}
	if !strings.Contains(string(data), "### "+id+".") {
		return fmt.Errorf("known-bug=%s 在 docs/known-issues.md 中找不到对应条目（缺陷已修复？请更新期望、移除标志并把用例转正）", id)
	}
	return nil
}

func diffText(expected, actual string) string {
	el, al := strings.Split(expected, "\n"), strings.Split(actual, "\n")
	n := max(len(el), len(al))
	lineAt := func(lines []string, i int) string {
		if i < len(lines) {
			return lines[i]
		}
		return "<无>"
	}
	var sb strings.Builder
	fmt.Fprintf(&sb, "  期望 %d 行 / 实际 %d 行\n", len(el), len(al))
	for i := 0; i < n; i++ {
		if lineAt(el, i) == lineAt(al, i) {
			continue
		}
		fmt.Fprintf(&sb, "  首次不一致在第 %d 行：\n", i+1)
		for j := i; j < i+3 && j < n; j++ {
			fmt.Fprintf(&sb, "    期望: %q\n    实际: %q\n", lineAt(el, j), lineAt(al, j))
		}
		break
	}
	return sb.String()
}
