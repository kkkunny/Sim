package parse

import (
	"errors"
	"fmt"

	"github.com/kkkunny/Sim/src/compiler/lex"
	"github.com/kkkunny/Sim/src/compiler/utils"
	stlos "github.com/kkkunny/stl/os"
)

// Global 全局
type Global interface {
	Ast
	Global()
}

// 包导入
type importAst struct {
	Pos         utils.Position
	Packages    []lex.Token
	Alias       *lex.Token // 别名（可能为空，与IsImportAll互斥）
	IsImportAll *lex.Token // 是否导入全部（可能为空，与Alias互斥）
}

func newImportAst(pos utils.Position, pkgs []lex.Token, alias *lex.Token, isImportAll *lex.Token) *importAst {
	return &importAst{
		Pos:         pos,
		Packages:    pkgs,
		Alias:       alias,
		IsImportAll: isImportAll,
	}
}

func (self importAst) Position() utils.Position {
	return self.Pos
}

func (self importAst) Global() {}

// TypeDef 类型定义
type TypeDef struct {
	Pos           utils.Position
	Public        bool
	Name          lex.Token
	GenericParams []lex.Token
	Target        Type
}

func NewTypeDef(
	pos utils.Position, pub bool, name lex.Token, genericParams []lex.Token, target Type,
) *TypeDef {
	return &TypeDef{
		Pos:           pos,
		Public:        pub,
		Name:          name,
		GenericParams: genericParams,
		Target:        target,
	}
}

func (self TypeDef) Position() utils.Position {
	return self.Pos
}

func (self TypeDef) Global() {}

// Function 函数
type Function struct {
	Pos           utils.Position
	Attrs         []Attr
	Public        bool
	Ret           Type
	Name          lex.Token
	GenericParams []lex.Token
	Params        []*Param
	VarArg        bool
	Body          *Block // （可能为空）
}

func NewFunction(
	pos utils.Position, attrs []Attr, pub bool, ret Type, name lex.Token, genericParams []lex.Token, params []*Param,
	varArg bool,
	body *Block,
) *Function {
	return &Function{
		Pos:           pos,
		Attrs:         attrs,
		Public:        pub,
		Ret:           ret,
		Name:          name,
		GenericParams: genericParams,
		Params:        params,
		VarArg:        varArg,
		Body:          body,
	}
}

func (self Function) Position() utils.Position {
	return self.Pos
}

func (self Function) Global() {}

// Method 方法
type Method struct {
	Pos               utils.Position
	Attrs             []Attr
	Public            bool
	Mutable           bool
	Self              lex.Token
	SelfGenericParams []lex.Token
	Ret               Type
	Name              lex.Token
	GenericParams     []lex.Token
	Params            []*Param
	VarArg            bool
	Body              *Block
}

func NewMethod(
	pos utils.Position, attrs []Attr, pub bool, mut bool, self lex.Token, selfGenericParams []lex.Token, ret Type,
	name lex.Token, genericParams []lex.Token, params []*Param,
	varArg bool, body *Block,
) *Method {
	return &Method{
		Pos:               pos,
		Attrs:             attrs,
		Public:            pub,
		Mutable:           mut,
		Self:              self,
		SelfGenericParams: selfGenericParams,
		Ret:               ret,
		Name:              name,
		GenericParams:     genericParams,
		Params:            params,
		VarArg:            varArg,
		Body:              body,
	}
}

func (self Method) Position() utils.Position {
	return self.Pos
}

func (self Method) Global() {}

// GlobalValue 全局变量
type GlobalValue struct {
	Pos     utils.Position
	Attrs   []Attr
	Public  bool
	Mutable bool // 是否可变
	Type    Type
	Name    lex.Token
	Value   Expr // 可能为空
}

func NewGlobalValue(pos utils.Position, attrs []Attr, pub bool, mut bool, t Type, name lex.Token, v Expr) *GlobalValue {
	return &GlobalValue{
		Pos:     pos,
		Attrs:   attrs,
		Public:  pub,
		Mutable: mut,
		Type:    t,
		Name:    name,
		Value:   v,
	}
}

func (self GlobalValue) Position() utils.Position {
	return self.Pos
}

func (self GlobalValue) Global() {}

// ****************************************************************

var (
	errStrUnknownGlobal = "unknown global"
	errStrCanNotUseAttr = "can not use this attribute"
)

// 全局
func (self *parser) parseGlobal() Global {
	var pub *lex.Token
	if self.skipNextIs(lex.PUB) {
		pub = &self.curTok
	}

	switch self.nextTok.Kind {
	case lex.IMPORT, lex.TYPE:
		return self.parseGlobalWithNoAttr(pub)
	case lex.Attr, lex.FUNC, lex.LET:
		return self.parseGlobalWithAttr(pub)
	default:
		fmt.Println(self.nextTok.Source)
		self.throwErrorf(self.nextTok.Pos, errStrUnknownGlobal)
		return nil
	}
}

// 全局（不带属性）
func (self *parser) parseGlobalWithNoAttr(pub *lex.Token) Global {
	switch self.nextTok.Kind {
	case lex.IMPORT:
		if pub != nil {
			self.throwErrorf(self.nextTok.Pos, errStrUnknownGlobal)
		}
		self.parseImport()
		return nil
	case lex.TYPE:
		return self.parseTypeDef(pub)
	default:
		self.throwErrorf(self.nextTok.Pos, errStrUnknownGlobal)
		return nil
	}
}

// 全局（带属性）
func (self *parser) parseGlobalWithAttr(pub *lex.Token) Global {
	var attrs []Attr
	if pub == nil {
		for self.nextIs(lex.Attr) {
			if len(attrs) == 0 && pub != nil {
				self.throwErrorf(self.nextTok.Pos, errStrUnknownGlobal)
			}
			attrs = append(attrs, self.parseAttr())
			self.expectNextIs(lex.SEM)
		}

		if self.skipNextIs(lex.PUB) {
			pub = &self.curTok
		}
	}

	switch self.nextTok.Kind {
	case lex.FUNC:
		return self.parseFunction(pub, attrs)
	case lex.LET:
		return self.parseGlobalValue(pub, attrs)
	default:
		self.throwErrorf(self.nextTok.Pos, errStrUnknownGlobal)
		return nil
	}
}

// 包导入
func (self *parser) parseImport() {
	// 语法分析
	begin := self.expectNextIs(lex.IMPORT).Pos
	paths := self.parseTokenListAtLeastOne(lex.CLL)
	var alias *lex.Token
	var isImportAll *lex.Token
	if self.skipNextIs(lex.AS) {
		if self.skipNextIs(lex.MUL) {
			name := self.curTok
			isImportAll = &name
		} else {
			name := self.expectNextIs(lex.IDENT)
			alias = &name
		}
	}
	ast := newImportAst(utils.MixPosition(begin, paths[len(paths)-1].Pos), paths, alias, isImportAll)

	// 导入包
	self.importImport(ast)
}

// 导入包
func (self *parser) importImport(ast *importAst) *Package {
	// 获取包绝对路径
	pathPos := utils.MixPosition(ast.Packages[0].Pos, ast.Packages[len(ast.Packages)-1].Pos)
	var path stlos.Path
	for _, p := range ast.Packages {
		path = path.Join(stlos.Path(p.Source))
	}
	path = utils.StdPath.Join(path)
	if absPath, err := path.GetAbsolute(); err != nil || !path.IsExist() {
		self.throwErrorf(pathPos, "unknown package")
	} else {
		path = absPath
	}

	// 解析包
	pkg, err := self.importPath(path)
	if err != nil {
		var parseError utils.Error
		if errors.As(err, &parseError) {
			panic(parseError)
		} else {
			self.throwErrorf(pathPos, err.Error())
		}
	}

	// 建立包关联
	if ast.IsImportAll == nil {
		// 包名
		var name string
		var namePos utils.Position
		if ast.Alias != nil {
			name = ast.Alias.Source
			namePos = ast.Alias.Pos
		} else {
			name = path.GetBase().String()
			namePos = ast.Packages[len(ast.Packages)-1].Pos
		}

		// 包名冲突
		if pkgTmp, ok := self.pkg.importMap[name]; ok && pkgTmp.Path != pkg.Path {
			self.throwErrorf(namePos, "duplicate package name")
		}

		self.pkg.importMap[name] = pkg
	} else {
		if self.pkg.includeMap.Contain(pkg) {
			self.pkg.includeMap.Remove(pkg)
		}
		self.pkg.includeMap.Add(pkg)
	}

	return pkg
}

// 导入地址
func (self *parser) importPath(path stlos.Path) (*Package, error) {
	// 缓存
	if pkg, ok := pkgs[path]; ok {
		if pkg == nil {
			return nil, fmt.Errorf("circular reference package")
		}
		return pkg, nil
	}

	// 解析
	return parsePackage(path)
}

// 类型定义
func (self *parser) parseTypeDef(pub *lex.Token) *TypeDef {
	begin := self.expectNextIs(lex.TYPE).Pos
	name := self.expectNextIs(lex.IDENT)
	genericParams := self.parseGenericParamList()
	target := self.parseType()
	return NewTypeDef(utils.MixPosition(begin, target.Position()), pub != nil, name, genericParams, target)
}

// 函数
func (self *parser) parseFunction(pub *lex.Token, attrs []Attr) Global {
	var isExtern bool
	for _, attr := range attrs {
		switch attr.(type) {
		case *AttrExtern:
			isExtern = true
		case *AttrNoReturn, *AttrInline, *AttrInit, *AttrFini:
		default:
			self.throwErrorf(attr.Position(), errStrCanNotUseAttr)
			return nil
		}
	}

	self.expectNextIs(lex.FUNC)
	var begin utils.Position
	if len(attrs) > 0 {
		begin = attrs[0].Position()
	} else if pub != nil {
		begin = pub.Pos
	} else {
		begin = self.curTok.Pos
	}

	if self.nextIs(lex.LPA) {
		return self.parseMethod(begin, pub != nil, attrs)
	}

	name := self.expectNextIs(lex.IDENT)
	genericParams := self.parseGenericParamList()
	self.expectNextIs(lex.LPA)
	params, varArg := self.parseParamList(lex.RPA)
	self.expectNextIs(lex.RPA)
	ret := self.parseTypeOrNil()
	var body *Block
	if !isExtern || self.nextIs(lex.LBR) || len(genericParams) != 0 {
		body = self.parseBlock()
	}

	pos := utils.MixPosition(begin, self.curTok.Pos)
	if len(params) != 0 || ret != nil || len(genericParams) != 0 {
		for _, attr := range attrs {
			switch attr.(type) {
			case *AttrInit, *AttrFini:
				self.throwErrorf(attr.Position(), errStrCanNotUseAttr)
				return nil
			}
		}
	}
	if isExtern && body == nil && len(genericParams) == 0 {
		for _, attr := range attrs {
			switch attr.(type) {
			case *AttrExtern, *AttrNoReturn, *AttrInit, *AttrFini:
			default:
				self.throwErrorf(attr.Position(), errStrCanNotUseAttr)
				return nil
			}
		}
		return NewFunction(pos, attrs, pub != nil, ret, name, nil, params, varArg, nil)
	} else if len(genericParams) != 0 {
		for _, attr := range attrs {
			switch attr.(type) {
			case *AttrNoReturn, *AttrInline:
			default:
				self.throwErrorf(attr.Position(), errStrCanNotUseAttr)
				return nil
			}
		}
		return NewFunction(pos, attrs, pub != nil, ret, name, genericParams, params, varArg, body)
	} else {
		for _, attr := range attrs {
			switch attr.(type) {
			case *AttrExtern, *AttrNoReturn, *AttrInline, *AttrInit, *AttrFini:
			default:
				self.throwErrorf(attr.Position(), errStrCanNotUseAttr)
				return nil
			}
		}
		return NewFunction(pos, attrs, pub != nil, ret, name, genericParams, params, varArg, body)
	}
}

// 方法
func (self *parser) parseMethod(begin utils.Position, pub bool, attrs []Attr) *Method {
	for _, attr := range attrs {
		switch attr.(type) {
		case *AttrNoReturn, *AttrInline:
		default:
			self.throwErrorf(attr.Position(), errStrCanNotUseAttr)
			return nil
		}
	}

	self.expectNextIs(lex.LPA)
	mut := self.skipNextIs(lex.MUT)
	selfTok := self.expectNextIs(lex.IDENT)
	selfGenericParams := self.parseGenericParamList()
	self.expectNextIs(lex.RPA)

	name := self.expectNextIs(lex.IDENT)
	genericParams := self.parseGenericParamList()
	self.expectNextIs(lex.LPA)
	params, varArg := self.parseParamList(lex.RPA)
	self.expectNextIs(lex.RPA)
	ret := self.parseTypeOrNil()
	var body *Block
	if self.nextIs(lex.LBR) {
		body = self.parseBlock()
	}
	return NewMethod(
		utils.MixPosition(begin, self.curTok.Pos),
		attrs,
		pub,
		mut,
		selfTok,
		selfGenericParams,
		ret,
		name,
		genericParams,
		params,
		varArg,
		body,
	)
}

// 全局变量
func (self *parser) parseGlobalValue(pub *lex.Token, attrs []Attr) *GlobalValue {
	for _, attr := range attrs {
		switch attr.(type) {
		case *AttrExtern:
		default:
			self.throwErrorf(attr.Position(), errStrCanNotUseAttr)
			return nil
		}
	}

	var begin utils.Position
	self.expectNextIs(lex.LET)
	if len(attrs) > 0 {
		begin = attrs[0].Position()
	} else if pub != nil {
		begin = pub.Pos
	} else {
		begin = self.curTok.Pos
	}

	mut := self.skipNextIs(lex.MUT)

	name := self.expectNextIs(lex.IDENT)

	self.expectNextIs(lex.COL)
	t := self.parseType()

	var v Expr
	if self.skipNextIs(lex.ASS) {
		v = self.parseExpr()
	}
	return NewGlobalValue(utils.MixPosition(begin, self.curTok.Pos), attrs, pub != nil, mut, t, name, v)
}

// 泛型形参
func (self *parser) parseGenericParamList() []lex.Token {
	if !self.skipNextIs(lex.LT) {
		return nil
	}
	var params []lex.Token
	for !self.nextIs(lex.GT) {
		params = append(params, self.expectNextIs(lex.IDENT))
		if !self.skipNextIs(lex.COM) {
			break
		}
	}
	self.expectNextIs(lex.GT)
	return params
}
