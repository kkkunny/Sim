package parse

import (
	"os"
	"path/filepath"

	"github.com/kkkunny/stl/container/linkedlist"
	stlerror "github.com/kkkunny/stl/error"

	"github.com/kkkunny/Sim/ast"
	"github.com/kkkunny/Sim/lex"
	"github.com/kkkunny/Sim/reader"
	"github.com/kkkunny/Sim/token"
)

func loopParseWithUtil[T any](self *Parser, sem, end token.Kind, f func() T) (res []T) {
	for self.skipSEM(); !self.nextIs(end); self.skipSEM() {
		res = append(res, f())
		if !self.skipNextIs(sem) {
			break
		}
	}
	self.skipSEM()
	return res
}

// ParseFile 语法解析目标文件
func ParseFile(path string) (linkedlist.LinkedList[ast.Global], stlerror.Error) {
	_, r, err := reader.NewReaderFromFile(path)
	if err != nil {
		return linkedlist.LinkedList[ast.Global]{}, err
	}
	defer func() {
		// HACK: 打开的文件需要关闭
		// _ = closer.Close()
	}()
	return New(lex.New(r)).Parse(), nil
}

// ParseDir 语法解析目标目录
func ParseDir(path string) (linkedlist.LinkedList[ast.Global], stlerror.Error) {
	entries, err := stlerror.ErrorWith(os.ReadDir(path))
	if err != nil {
		return linkedlist.LinkedList[ast.Global]{}, err
	}
	var asts linkedlist.LinkedList[ast.Global]
	for _, entry := range entries {
		fileAst, err := ParseFile(filepath.Join(path, entry.Name()))
		if err != nil {
			return linkedlist.LinkedList[ast.Global]{}, err
		}
		asts.Append(fileAst)
	}
	return asts, nil
}
