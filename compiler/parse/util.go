package parse

import (
	"os"
	"path/filepath"
	"reflect"

	"github.com/kkkunny/stl/container/linkedlist"
	"github.com/kkkunny/stl/container/optional"
	stlerror "github.com/kkkunny/stl/error"
	stlos "github.com/kkkunny/stl/os"
	"github.com/samber/lo"

	"github.com/kkkunny/Sim/compiler/ast"

	"github.com/kkkunny/Sim/compiler/lex"

	"github.com/kkkunny/Sim/compiler/reader"

	errors "github.com/kkkunny/Sim/compiler/error"

	"github.com/kkkunny/Sim/compiler/token"
)

func loopParseWithUtil[T any](self *Parser, sem, end token.Kind, f func() T, atLeastOne ...bool) (res []T) {
	atLeastOneVal := len(atLeastOne) > 0 && atLeastOne[0]
	for self.skipSEM(); (len(res) == 0 && atLeastOneVal) || !self.nextIs(end); self.skipSEM() {
		res = append(res, f())
		if !self.skipNextIs(sem) {
			break
		}
	}
	self.skipSEM()
	return res
}

func (self *Parser) parseExprList(end token.Kind, atLeaseOne ...bool) (res []ast.Expr) {
	return loopParseWithUtil(self, token.COM, end, func() ast.Expr {
		return self.mustExpr(self.parseOptionExpr(true))
	}, atLeaseOne...)
}

func (self *Parser) parseTypeList(end token.Kind, atLeastOne ...bool) (res []ast.Type) {
	return loopParseWithUtil(self, token.COM, end, func() ast.Type {
		return self.parseType()
	}, atLeastOne...)
}

func (self *Parser) parseParamList(end token.Kind) (res []ast.Param) {
	return loopParseWithUtil(self, token.COM, end, func() ast.Param {
		return self.parseParam()
	})
}

func (self *Parser) parseParam() ast.Param {
	if self.skipNextIs(token.MUT) {
		mut := self.curTok
		name := self.expectNextIs(token.IDENT)
		self.expectNextIs(token.COL)
		typ := self.parseType()
		return ast.Param{
			Mutable: optional.Some(mut),
			Name:    optional.Some(name),
			Type:    typ,
		}
	} else {
		var name optional.Optional[token.Token]
		typ := self.parseType()
		if ident, ok := typ.(*ast.IdentType); ok && ident.Pkg.IsNone() && ident.GenericArgs.IsNone() && self.skipNextIs(token.COL) {
			name = optional.Some(ident.Name)
			typ = self.parseType()
		}
		return ast.Param{
			Mutable: optional.None[token.Token](),
			Name:    name,
			Type:    typ,
		}
	}
}

func (self *Parser) parseFieldList(end token.Kind) (res []ast.Field) {
	return loopParseWithUtil(self, token.COM, end, func() ast.Field {
		pub := self.skipNextIs(token.PUBLIC)
		mut := self.skipNextIs(token.MUT)
		pn := self.expectNextIs(token.IDENT)
		self.expectNextIs(token.COL)
		pt := self.parseType()
		return ast.Field{
			Public:  pub,
			Mutable: mut,
			Name:    pn,
			Type:    pt,
		}
	})
}

func expectAttrIn(attrs []ast.Attr, expectAttr ...ast.Attr) {
	expectAttrTypes := lo.Map(expectAttr, func(item ast.Attr, _ int) reflect.Type {
		return reflect.ValueOf(item).Type()
	})
loop:
	for _, attr := range attrs {
		if len(expectAttrTypes) == 0 {
			errors.ThrowUnExpectAttr(attr.Position())
		}
		attrType := reflect.ValueOf(attr).Type()
		for _, expectAttrType := range expectAttrTypes {
			if attrType.AssignableTo(expectAttrType) {
				continue loop
			}
		}
		errors.ThrowUnExpectAttr(attr.Position())
	}
}

func (self *Parser) parseIdent() *ast.Ident {
	var pkg optional.Optional[token.Token]
	var name token.Token
	var genericArgs optional.Optional[*ast.GenericArgList]

	pkgOrName := self.expectNextIs(token.IDENT)
	if !self.skipNextIs(token.SCOPE) {
		name = pkgOrName
	} else {
		if !self.nextIs(token.LT) {
			pkg, name = optional.Some(pkgOrName), self.expectNextIs(token.IDENT)
			if self.skipNextIs(token.SCOPE) {
				genericArgs = optional.Some(self.parseGenericArgList())
			}
		} else {
			name = pkgOrName
			genericArgs = optional.Some(self.parseGenericArgList())
		}
	}
	return &ast.Ident{
		Pkg:         pkg,
		Name:        name,
		GenericArgs: genericArgs,
	}
}

func (self *Parser) parseGenericParamList() optional.Optional[*ast.GenericParamList] {
	if !self.skipNextIs(token.LT) {
		return optional.None[*ast.GenericParamList]()
	}
	begin := self.curTok.Position
	params := loopParseWithUtil(self, token.COM, token.GT, func() token.Token {
		return self.expectNextIs(token.IDENT)
	}, true)
	end := self.expectNextIs(token.GT).Position
	return optional.Some(&ast.GenericParamList{
		Begin:  begin,
		Params: params,
		End:    end,
	})
}

func (self *Parser) parseGenericArgList() *ast.GenericArgList {
	begin := self.expectNextIs(token.LT).Position
	args := loopParseWithUtil(self, token.COM, token.GT, func() ast.Type {
		return self.parseType()
	}, true)
	end := self.expectNextIs(token.GT).Position
	return &ast.GenericArgList{
		Begin: begin,
		Args:  args,
		End:   end,
	}
}

func (self *Parser) parseFuncDecl(beforeName, afterName func()) ast.FuncDecl {
	begin := self.expectNextIs(token.FUNC).Position
	if beforeName != nil {
		beforeName()
	}
	name := self.expectNextIs(token.IDENT)
	if afterName != nil {
		afterName()
	}
	self.expectNextIs(token.LPA)
	params := self.parseParamList(token.RPA)
	self.expectNextIs(token.RPA)
	ret := self.parseOptionType()
	return ast.FuncDecl{
		Begin:  begin,
		Name:   name,
		Params: params,
		Ret:    ret,
		End:    self.curTok.Position,
	}
}

// 语法解析目标文件
func parseFile(path stlos.FilePath) (linkedlist.LinkedList[ast.Global], error) {
	_, r, err := reader.NewReaderFromFile(path)
	if err != nil {
		return linkedlist.LinkedList[ast.Global]{}, err
	}
	return New(lex.New(r)).Parse(), nil
}

// 语法解析目标目录
func parseDir(path stlos.FilePath) ([]linkedlist.LinkedList[ast.Global], error) {
	entries, err := stlerror.ErrorWith(os.ReadDir(string(path)))
	if err != nil {
		return nil, err
	}
	var astsList []linkedlist.LinkedList[ast.Global]
	for _, entry := range entries {
		if entry.IsDir() || filepath.Ext(entry.Name()) != ".sim" {
			continue
		}
		fileAst, err := parseFile(path.Join(entry.Name()))
		if err != nil {
			return nil, err
		}
		astsList = append(astsList, fileAst)
	}
	return astsList, nil
}

// Parse 语法解析
func Parse(path stlos.FilePath) ([]linkedlist.LinkedList[ast.Global], error) {
	fs, err := stlerror.ErrorWith(os.Stat(string(path)))
	if err != nil {
		return nil, err
	}
	if fs.IsDir() {
		return parseDir(path)
	} else {
		asts, err := parseFile(path)
		if err != nil {
			return nil, err
		}
		return []linkedlist.LinkedList[ast.Global]{asts}, nil
	}
}
