package errors

import (
	"fmt"
	"os"
	"strings"

	stlslices "github.com/kkkunny/stl/container/slices"
	stlos "github.com/kkkunny/stl/os"

	"github.com/kkkunny/Sim/compiler/ast"
	"github.com/kkkunny/Sim/compiler/hir/types"
	"github.com/kkkunny/Sim/compiler/hir/values"

	"github.com/kkkunny/Sim/compiler/reader"

	"github.com/kkkunny/Sim/compiler/token"
)

// ThrowError 抛出异常
func ThrowError(pos reader.Position, format string, args ...any) {
	_, err := fmt.Fprintln(os.Stderr, fmt.Sprintf("%s:%d:%d: %s", pos.Reader.Path(), pos.BeginRow, pos.BeginCol, fmt.Sprintf(format, args...)))
	if err != nil {
		panic(err)
	}
	panic("compiler error")
}

// ThrowCanNotGetDefaultV2 不能获得默认值
func ThrowCanNotGetDefaultV2(pos reader.Position, t types.Type) {
	ThrowError(pos, "can not get the default value for type `%s`", t)
}

// ThrowExpectAttribute 期待属性
func ThrowExpectAttribute(pos reader.Position, attr ast.Attr) {
	ThrowError(pos, "expect attribute `%s`", attr.AttrName())
}

// ThrowTypeMismatchErrorV2 类型不匹配
func ThrowTypeMismatchErrorV2(pos reader.Position, t1, t2 types.Type) {
	ThrowError(pos, "type `%s` does not match type `%s`", t1, t2)
}

// ThrowIdentifierDuplicationError 标识符重复
func ThrowIdentifierDuplicationError(pos reader.Position, ident token.Token) {
	ThrowError(pos, "identifier `%s` is repeated", ident.Source())
}

// ThrowUnknownIdentifierError 未知的标识符
func ThrowUnknownIdentifierError(pos reader.Position, ident token.Token) {
	ThrowError(pos, "identifier `%s` could not be found", ident.Source())
}

// ThrowExprTemporaryError 值必须不是临时的
func ThrowExprTemporaryError(pos reader.Position) {
	ThrowError(pos, "the value must not be temporary")
}

// ThrowNotMutableError 值必须是可变的
func ThrowNotMutableError(pos reader.Position) {
	ThrowError(pos, "the value must be mutable")
}

// ThrowIllegalBinaryErrorV2 非法的二元运算
func ThrowIllegalBinaryErrorV2(pos reader.Position, op token.Token, left, right values.Value) {
	ThrowError(pos, "illegal binary operation with type `%s` `%s` `%s`", left.Type(), op.Source(), right.Type())
}

// ThrowIllegalUnaryErrorV2 非法的一元运算
func ThrowIllegalUnaryErrorV2(pos reader.Position, op token.Token, t types.Type) {
	ThrowError(pos, "illegal unary operation with `%s` type `%s`", op.Source(), t)
}

// ThrowParameterNumberNotMatchError 参数数量不匹配
func ThrowParameterNumberNotMatchError(pos reader.Position, expect, now uint) {
	ThrowError(pos, "parameter number does not match, expect `%d`, but there has `%d`", expect, now)
}

// ThrowIllegalCovertErrorV2 非法的类型转换
func ThrowIllegalCovertErrorV2(pos reader.Position, from, to types.Type) {
	ThrowError(pos, "type `%s` can not covert to `%s`", from, to)
}

// ThrowExpectStructTypeErrorV2 期待结构体类型
func ThrowExpectStructTypeErrorV2(pos reader.Position, t types.Type) {
	ThrowError(pos, "expect a struct type but there is type `%s`", t)
}

// ThrowExpectArrayTypeErrorV2 期待数组类型
func ThrowExpectArrayTypeErrorV2(pos reader.Position, t types.Type) {
	ThrowError(pos, "expect a array type but there is type `%s`", t)
}

// ThrowExpectEnumTypeErrorV2 期待枚举类型
func ThrowExpectEnumTypeErrorV2(pos reader.Position, t types.Type) {
	ThrowError(pos, "expect a enum type but there is type `%s`", t)
}

// ThrowExpectCallableErrorV2 期待一个可调用的
func ThrowExpectCallableErrorV2(pos reader.Position, t types.Type) {
	ThrowError(pos, "expect a callable but there is type `%s`", t)
}

// ThrowExpectReferenceErrorV2 期待一个引用
func ThrowExpectReferenceErrorV2(pos reader.Position, t types.Type) {
	ThrowError(pos, "expect a reference but there is type `%s`", t)
}

// ThrowExpectArrayErrorV2 期待一个数组
func ThrowExpectArrayErrorV2(pos reader.Position, t types.Type) {
	ThrowError(pos, "expect a array but there is type `%s`", t)
}

// ThrowExpectTupleErrorV2 期待一个元组
func ThrowExpectTupleErrorV2(pos reader.Position, t types.Type) {
	ThrowError(pos, "expect a tuple but there is type `%s`", t)
}

// ThrowExpectStructErrorV2 期待一个结构体
func ThrowExpectStructErrorV2(pos reader.Position, t types.Type) {
	ThrowError(pos, "expect a struct but there is type `%s`", t)
}

// ThrowInvalidIndexError 超出下标
func ThrowInvalidIndexError(pos reader.Position, index uint) {
	ThrowError(pos, "invalid index with `%d`", index)
}

// ThrowLoopControlError 非法的循环控制
func ThrowLoopControlError(pos reader.Position) {
	ThrowError(pos, "must in a loop")
}

// ThrowNotExpectToken 不期待的token
func ThrowNotExpectToken(pos reader.Position, expect token.Kind, token token.Token) {
	ThrowError(pos, "expect next token is `%s`, but there is `%s`", expect, token.Kind)
}

// ThrowIllegalInteger 非法的整数
func ThrowIllegalInteger(pos reader.Position, token token.Token) {
	ThrowError(pos, "illegal integer `%s`", token.Source())
}

// ThrowIllegalExpression 非法的表达式
func ThrowIllegalExpression(pos reader.Position) {
	ThrowError(pos, "illegal expression")
}

// ThrowIllegalGlobal 非法的全局
func ThrowIllegalGlobal(pos reader.Position) {
	ThrowError(pos, "illegal global")
}

// ThrowIllegalAttr 非法的属性
func ThrowIllegalAttr(pos reader.Position) {
	ThrowError(pos, "illegal attribute")
}

// ThrowUnExpectAttr 不期待的属性
func ThrowUnExpectAttr(pos reader.Position) {
	ThrowError(pos, "unexpected attribute for this global")
}

// ThrowIllegalType 非法的类型
func ThrowIllegalType(pos reader.Position) {
	ThrowError(pos, "illegal type")
}

// ThrowInvalidPackage 无效包
func ThrowInvalidPackage(pos reader.Position, paths []token.Token) {
	pathStrs := stlslices.Map(paths, func(_ int, v token.Token) string {
		return v.Source()
	})
	ThrowError(pos, "package `%s` is invalid", strings.Join(pathStrs, "."))
}

// ThrowCircularReference 循环引用
func ThrowCircularReference(pos reader.Position, path token.Token) {
	ThrowError(pos, "circular reference `%s`", path.Source())
}

// ThrowIndexOutOfRange 超出下标
func ThrowIndexOutOfRange(pos reader.Position) {
	ThrowError(pos, "index out of range")
}

// ThrowExpectMoreCaseV2 期待更多case
func ThrowExpectMoreCaseV2(pos reader.Position, et types.Type, now, expect uint) {
	ThrowError(pos, "type `%s` has `%d` case but there is `%d`", et, expect, now)
}

// ThrowPackageCircularReference 包循环引用
func ThrowPackageCircularReference(pos reader.Position, pkgChain []stlos.FilePath) {
	pkgChain = append(pkgChain, pkgChain[0])
	ThrowError(pos, "package circular reference: %s->", pkgChain[len(pkgChain)-1], strings.Join(stlslices.Map(pkgChain, func(_ int, pkg stlos.FilePath) string {
		return string(pkg)
	}), "->"))
}
