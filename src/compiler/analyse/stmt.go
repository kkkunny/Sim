package analyse

import (
	"github.com/kkkunny/Sim/src/compiler/parse"
	"github.com/kkkunny/Sim/src/compiler/utils"
	"github.com/kkkunny/stl/types"
)

// Block 代码块
type Block struct {
	Stmts []Stmt
}

func (self Block) stmt() {}

// Stmt 语句
type Stmt interface {
	stmt()
}

// Return 函数返回
type Return struct {
	Value Expr
}

func (self Return) stmt() {}

// Variable 变量
type Variable struct {
	Type  Type
	Value Expr
}

func (self Variable) stmt() {}

func (self Variable) ident() {}

func (self Variable) GetType() Type {
	return self.Type
}

func (self Variable) GetMut() bool {
	return true
}

func (self Variable) IsTemporary() bool {
	return false
}

func (self Variable) IsConst() bool {
	return false
}

// IfElse 条件分支
type IfElse struct {
	Cond        Expr
	True, False *Block
}

func (self IfElse) stmt() {}

// Loop 循环
type Loop struct {
	Cond Expr
	Body *Block
}

func (self Loop) stmt() {}

// LoopControl 循环
type LoopControl struct {
	Type string
}

func (self LoopControl) stmt() {}

// Switch 分支
type Switch struct {
	From    Expr
	Cases   []types.Pair[Expr, *Block]
	Default *Block // 可能为空
}

func (self Switch) stmt() {}

// *********************************************************************************************************************

// 代码块
func analyseBlock(ctx localContext, ast *parse.Block, inLoop bool) (*blockContext, *Block, utils.Error) {
	bctx := newBlockContext(ctx, inLoop)

	var stmts []Stmt

	var errors []utils.Error
	for iter := ast.Stmts.Iterator(); iter.HasValue(); iter.Next() {
		if bctx.IsEnd() {
			break
		}
		stmt, err := analyseStmt(bctx, iter.Value())
		if err != nil {
			errors = append(errors, err)
		} else {
			stmts = append(stmts, stmt)
		}
	}

	block := &Block{Stmts: stmts}
	if len(errors) == 0 {
		return bctx, block, nil
	} else if len(errors) == 1 {
		return nil, nil, errors[0]
	} else {
		return nil, nil, utils.NewMultiError(errors...)
	}
}

// 语句
func analyseStmt(ctx *blockContext, ast parse.Stmt) (Stmt, utils.Error) {
	switch stmt := ast.(type) {
	case *parse.Return:
		res, err := analyseReturn(ctx, stmt)
		ctx.SetEnd()
		return res, err
	case *parse.Variable:
		return analyseVariable(ctx, stmt)
	case parse.Expr:
		return analyseExpr(ctx, nil, stmt)
	case *parse.Block:
		bctx, res, err := analyseBlock(ctx, stmt, false)
		if err != nil {
			return nil, err
		}
		if bctx.IsEnd() {
			ctx.SetEnd()
		}
		return res, nil
	case *parse.IfElse:
		res, err, end := analyseIfElse(ctx, stmt)
		if err != nil {
			return nil, err
		}
		if end {
			ctx.SetEnd()
		}
		return res, nil
	case *parse.Loop:
		return analyseFor(ctx, stmt)
	case *parse.LoopControl:
		if !ctx.IsInLoop() {
			return nil, utils.Errorf(stmt.Position(), "must in a loop")
		}
		ctx.SetEnd()
		return &LoopControl{Type: stmt.Kind.Source}, nil
	case *parse.Switch:
		return analyseSwitch(ctx, stmt)
	default:
		panic("unknown stmt")
	}
}

// 函数返回
func analyseReturn(ctx *blockContext, ast *parse.Return) (*Return, utils.Error) {
	ret := ctx.GetRetType()
	if ast.Value == nil {
		if ret.Equal(None) {
			return &Return{}, nil
		} else {
			return nil, utils.Errorf(ast.Position(), "expect a return value")
		}
	} else {
		if ret.Equal(None) {
			return nil, utils.Errorf(ast.Position(), "not expect a return value")
		} else {
			value, err := expectExpr(ctx, ret, ast.Value)
			if err != nil {
				return nil, err
			}
			return &Return{Value: value}, nil
		}
	}
}

// 变量
func analyseVariable(ctx *blockContext, ast *parse.Variable) (*Variable, utils.Error) {
	if ast.Type == nil && ast.Value == nil {
		return nil, utils.Errorf(ast.Name.Pos, "expect a type or a value")
	}

	var typ Type
	var err utils.Error
	if ast.Type != nil {
		typ, err = analyseType(ctx.GetPackageContext(), ast.Type)
		if err != nil {
			return nil, err
		}
	}

	var value Expr
	if ast.Type != nil && ast.Value != nil {
		value, err = expectExpr(ctx, typ, ast.Value)
		if err != nil {
			return nil, err
		}
	} else if ast.Type == nil && ast.Value != nil {
		value, err = analyseExpr(ctx, nil, ast.Value)
		if err != nil {
			return nil, err
		} else if IsNoneType(value.GetType()) {
			return nil, utils.Errorf(ast.Value.Position(), "expect a value")
		}
		typ = value.GetType()
	} else {
		value = getDefaultExprByType(typ)
	}

	v := &Variable{
		Type:  typ,
		Value: value,
	}
	if !ctx.AddValue(ast.Name.Source, v) {
		return nil, utils.Errorf(ast.Name.Pos, "duplicate identifier")
	}
	return v, nil
}

// 条件分支
func analyseIfElse(ctx *blockContext, ast *parse.IfElse) (*IfElse, utils.Error, bool) {
	cond, err := expectExprAndSon(ctx, Bool, ast.Cond)
	if err != nil {
		return nil, err, false
	}

	tctx, tb, te := analyseBlock(ctx, ast.Body, false)

	if ast.Next == nil {
		if te != nil {
			return nil, te, false
		}
		return &IfElse{
			Cond: cond,
			True: tb,
		}, nil, false
	} else if ast.Next != nil && ast.Next.Cond == nil {
		fctx, fb, fe := analyseBlock(ctx, ast.Next.Body, false)
		if te != nil && fe != nil {
			return nil, utils.NewMultiError(te, fe), false
		} else if te != nil {
			return nil, te, false
		} else if fe != nil {
			return nil, fe, false
		}
		return &IfElse{
			Cond:  cond,
			True:  tb,
			False: fb,
		}, nil, tctx.IsEnd() && fctx.IsEnd()
	} else {
		nb, ne, nret := analyseIfElse(ctx, ast.Next)
		if te != nil && ne != nil {
			return nil, utils.NewMultiError(te, ne), false
		} else if te != nil {
			return nil, te, false
		} else if ne != nil {
			return nil, ne, false
		}
		return &IfElse{
			Cond:  cond,
			True:  tb,
			False: &Block{Stmts: []Stmt{nb}},
		}, nil, tctx.IsEnd() && nret
	}
}

// 循环
func analyseFor(ctx *blockContext, ast *parse.Loop) (*Loop, utils.Error) {
	cond, err := expectExprAndSon(ctx, Bool, ast.Cond)
	if err != nil {
		return nil, err
	}

	_, body, err := analyseBlock(ctx, ast.Body, true)
	if err != nil {
		return nil, err
	}

	return &Loop{
		Cond: cond,
		Body: body,
	}, nil
}

// 分支
func analyseSwitch(ctx *blockContext, ast *parse.Switch) (*Switch, utils.Error) {
	from, err := analyseExpr(ctx, nil, ast.From)
	if err != nil {
		return nil, err
	}
	ft := from.GetType()

	cases := make([]types.Pair[Expr, *Block], len(ast.Cases))
	end := ast.Default != nil
	var errs []utils.Error
	for i, c := range ast.Cases {
		cv, err := expectExpr(ctx, ft, c.First)
		if err != nil {
			errs = append(errs, err)
			continue
		}
		cbCtx, cb, err := analyseBlock(ctx, c.Second, false)
		if err != nil {
			errs = append(errs, err)
			continue
		}
		end = end && cbCtx.end
		cases[i] = types.NewPair(cv, cb)
	}
	var de *Block
	if ast.Default != nil {
		var deCtx *blockContext
		deCtx, de, err = analyseBlock(ctx, ast.Default, false)
		if err != nil {
			errs = append(errs, err)
		} else {
			end = end && deCtx.end
		}
	}
	if len(errs) == 0 {
		return &Switch{
			From:    from,
			Cases:   cases,
			Default: de,
		}, nil
	} else if len(errs) == 1 {
		return nil, errs[0]
	} else {
		return nil, utils.NewMultiError(errs...)
	}
}
