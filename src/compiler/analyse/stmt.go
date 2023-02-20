package analyse

import (
	"github.com/kkkunny/Sim/src/compiler/hir"
	"github.com/kkkunny/Sim/src/compiler/lex"
	"github.com/kkkunny/Sim/src/compiler/parse"
	"github.com/kkkunny/Sim/src/compiler/utils"
	"github.com/kkkunny/stl/list"
)

// 语句
func (self *Analyser) analyseStmt(astObj parse.Stmt) (hir.Stmt, bool, utils.Error) {
	switch ast := astObj.(type) {
	case *parse.Block:
		return self.analyseBlock(*ast)
	case *parse.IfElse:
		return self.analyseIfElse(*ast)
	case *parse.Loop:
		stmt, err := self.analyseLoop(*ast)
		return stmt, false, err
	case *parse.LoopControl:
		stmt, err := self.analyseLoopControl(*ast)
		return stmt, false, err
	case *parse.Return:
		stmt, err := self.analyseReturn(*ast)
		return stmt, true, err
	case *parse.Switch:
		return self.analyseSwitch(*ast)
	case *parse.Variable:
		stmt, err := self.analyseVariable(*ast)
		return stmt, false, err
	case parse.Expr:
		expr, err := self.analyseExpr(nil, ast)
		return expr, false, err
	default:
		panic("unreachable")
	}
}

// 代码块
func (self *Analyser) analyseBlock(ast parse.Block, pro ...func(symbol *symbolTable) utils.Error) (
	*hir.Block, bool, utils.Error,
) {
	self.symbol = newBlockSymbolTable(self.symbol)
	defer func() {
		self.symbol = self.symbol.f
	}()

	for _, f := range pro {
		if err := f(self.symbol); err != nil {
			return nil, false, err
		}
	}

	stmts := list.NewSingleLinkedList[hir.Stmt]()
	var ret bool

loop:
	for iter := ast.Stmts.Iterator(); iter.HasValue(); iter.Next() {
		stmt, stmtRet, err := self.analyseStmt(iter.Value())
		if err != nil {
			return nil, false, err
		}
		stmts.PushBack(stmt)
		ret = ret || stmtRet

		switch stmt.(type) {
		case *hir.Break, *hir.Continue, *hir.Return:
			break loop
		}
	}

	return hir.NewBlock(stmts), ret, nil
}

// 条件分支
func (self *Analyser) analyseIfElse(ast parse.IfElse) (*hir.IfElse, bool, utils.Error) {
	// 条件
	cond, err := self.expectLikeExpr(hir.NewTypeBool(), ast.Cond)
	if err != nil {
		return nil, false, err
	}

	// true
	tb, tr, err := self.analyseBlock(*ast.Body)
	if err != nil {
		return nil, false, err
	}

	if ast.Next == nil {
		return hir.NewIfElse(cond, tb, nil), false, nil
	}

	// else
	if ast.Next.Cond == nil {
		fb, fr, err := self.analyseBlock(*ast.Next.Body)
		if err != nil {
			return nil, false, err
		}
		return hir.NewIfElse(cond, tb, fb), tr && fr, nil
	}

	// else if
	fb, fr, err := self.analyseIfElse(*ast.Next)
	if err != nil {
		return nil, false, err
	}
	return hir.NewIfElse(cond, tb, hir.NewBlock(list.NewSingleLinkedList[hir.Stmt](fb))), tr && fr, nil
}

// 循环
func (self *Analyser) analyseLoop(ast parse.Loop) (*hir.Loop, utils.Error) {
	// TODO: 循环体内函数返回判断
	// 条件
	cond, err := self.expectLikeExpr(hir.NewTypeBool(), ast.Cond)
	if err != nil {
		return nil, err
	}

	// 循环体
	body, _, err := self.analyseBlock(
		*ast.Body, func(symbol *symbolTable) utils.Error {
			symbol.inLoop = true
			return nil
		},
	)
	if err != nil {
		return nil, err
	}

	return hir.NewLoop(cond, body), nil
}

// 循环控制
func (self *Analyser) analyseLoopControl(ast parse.LoopControl) (hir.LoopControl, utils.Error) {
	if !self.symbol.inLoop {
		return nil, utils.Errorf(ast.Position(), "must in a loop")
	}
	switch ast.Kind.Kind {
	case lex.BREAK:
		return hir.NewBreak(), nil
	case lex.CONTINUE:
		return hir.NewContinue(), nil
	default:
		panic("unreachable")
	}
}

// 函数返回
func (self *Analyser) analyseReturn(ast parse.Return) (*hir.Return, utils.Error) {
	if ast.Value != nil && !self.symbol.ret.IsNone() {
		value, err := self.expectExpr(self.symbol.ret, ast.Value)
		if err != nil {
			return nil, err
		}
		return hir.NewReturn(value), nil
	} else if ast.Value != nil {
		return nil, utils.Errorf(ast.Value.Position(), "do not expect any value")
	} else if !self.symbol.ret.IsNone() {
		return nil, utils.Errorf(ast.Position(), "expect a value")
	} else {
		return hir.NewReturn(nil), nil
	}
}

// 分支
func (self *Analyser) analyseSwitch(ast parse.Switch) (hir.Stmt, bool, utils.Error) {
	// 判断值
	from, err := self.analyseExpr(nil, ast.From)
	if err != nil {
		return nil, false, err
	}
	fromType := from.Type()

	// 枚举匹配
	if fromType.IsEnum() {
		return self.analyseMatch(from, ast)
	}

	// 分支
	var errs []utils.Error
	cas, cbs := make([]hir.Expr, len(ast.Cases)), make([]*hir.Block, len(ast.Cases))
	for i, c := range ast.Cases {
		v, err := self.expectExpr(fromType, c.First)
		if err != nil {
			errs = append(errs, err)
			continue
		}
		b, _, err := self.analyseBlock(*c.Second)
		if err != nil {
			errs = append(errs, err)
			continue
		}
		cas[i], cbs[i] = v, b
	}

	// 默认分支
	var db *hir.Block
	var ret bool
	if ast.Default != nil {
		db, ret, err = self.analyseBlock(*ast.Default)
		if err != nil {
			errs = append(errs, err)
		}
	}

	if len(errs) == 1 {
		return nil, false, errs[0]
	} else if len(errs) > 1 {
		return nil, false, utils.NewMultiError(errs...)
	}

	return hir.NewSwitch(from, cas, cbs, db), ret, nil
}

// 枚举匹配
func (self *Analyser) analyseMatch(from hir.Expr, ast parse.Switch) (*hir.Match, bool, utils.Error) {
	fromType := from.Type()

	// 分支
	var errs []utils.Error
	caSet := make(map[string]struct{})
	cas, cbs := make([]string, len(ast.Cases)), make([]*hir.Block, len(ast.Cases))
	for i, c := range ast.Cases {
		ident, ok := c.First.(*parse.Ident)
		if !ok || ident.Pkgs[0].Path != self.symbol.pkg.Path {
			errs = append(errs, utils.Errorf(c.First.Position(), "expect enum field"))
			continue
		}

		field, ok := fromType.GetEnumFieldByName(ident.Name.Source)
		if !ok || (fromType.IsTypedef() && !fromType.GetTypedef().Pkg.Equal(self.symbol.pkg) && !field.First) {
			errs = append(errs, utils.Errorf(ident.Name.Pos, errUnknownIdentifier))
			continue
		}

		if _, ok := caSet[ident.Name.Source]; ok {
			errs = append(errs, utils.Errorf(ident.Name.Pos, errDuplicateDeclaration))
			continue
		}
		caSet[ident.Name.Source] = struct{}{}

		body, _, err := self.analyseBlock(*c.Second)
		if err != nil {
			errs = append(errs, err)
			continue
		}

		cas[i], cbs[i] = field.Second, body
	}

	// 默认分支
	var db *hir.Block
	var ret bool
	if ast.Default != nil {
		var err utils.Error
		db, ret, err = self.analyseBlock(*ast.Default)
		if err != nil {
			errs = append(errs, err)
		}
	}

	if len(errs) == 1 {
		return nil, false, errs[0]
	} else if len(errs) > 1 {
		return nil, false, utils.NewMultiError(errs...)
	}

	return hir.NewMatch(from, cas, cbs, db), ret, nil
}

// 变量定义
func (self *Analyser) analyseVariable(ast parse.Variable) (*hir.Variable, utils.Error) {
	// 类型
	var typ hir.Type
	var err utils.Error
	if ast.Type != nil {
		typ, err = self.analyseType(ast.Type)
	}
	if err != nil {
		return nil, err
	}

	// 值
	var value hir.Expr
	if ast.Value != nil && ast.Type != nil {
		value, err = self.expectExpr(typ, ast.Value)
	} else if ast.Value != nil {
		value, err = self.analyseExpr(nil, ast.Value)
		if err == nil {
			typ = value.Type()
		}
	} else {
		value, err = self.getDefaultValue(typ)
	}
	if err != nil {
		return nil, err
	}

	// 定义
	ident := hir.NewVariable(typ, value)
	self.symbol.defValue(false, ast.Name.Source, ident)

	return ident, nil
}
