package errors

import (
	"fmt"
	"os"
	"strings"

	"github.com/kkkunny/stl/container/dynarray"
	"github.com/kkkunny/stl/container/iterator"

	"github.com/kkkunny/Sim/ast"
	"github.com/kkkunny/Sim/hir"
	"github.com/kkkunny/Sim/reader"
	"github.com/kkkunny/Sim/token"
)

// ThrowError 抛出异常
func ThrowError(pos reader.Position, format string, args ...any) {
	_, err := fmt.Fprintln(os.Stderr, fmt.Sprintf("%s:%d:%d: %s", pos.Reader.Path(), pos.BeginRow, pos.BeginCol, fmt.Sprintf(format, args...)))
	if err != nil {
		panic(err)
	}
	panic("compiler error")
}

// ThrowCanNotGetDefault 不能获得默认值
func ThrowCanNotGetDefault(pos reader.Position, t hir.Type) {
	ThrowError(pos, "can not get the default value for type `%s`", t)
}

// ThrowExpectAttribute 期待属性
func ThrowExpectAttribute(pos reader.Position, attr ast.Attr) {
	ThrowError(pos, "expect attribute `%s`", attr.AttrName())
}

// ThrowTypeMismatchError 类型不匹配
func ThrowTypeMismatchError(pos reader.Position, t1, t2 hir.Type) {
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

// ThrowNotMutableError 值必须是可变的
func ThrowNotMutableError(pos reader.Position) {
	ThrowError(pos, "the value must be mutable")
}

// ThrowIllegalBinaryError 非法的二元运算
func ThrowIllegalBinaryError(pos reader.Position, op token.Token, left, right hir.Expr) {
	ThrowError(pos, "illegal binary operation with type `%s` `%s` `%s`", left.GetType(), op.Source(), right.GetType())
}

// ThrowIllegalUnaryError 非法的一元运算
func ThrowIllegalUnaryError(pos reader.Position, op token.Token, t hir.Type) {
	ThrowError(pos, "illegal unary operation with `%s` type `%s`", op.Source(), t)
}

// ThrowNotFunctionError 必须是函数
func ThrowNotFunctionError(pos reader.Position, t hir.Type) {
	ThrowError(pos, "expect function type but there is `%s`", t)
}

// ThrowParameterNumberNotMatchError 参数数量不匹配
func ThrowParameterNumberNotMatchError(pos reader.Position, expect, now uint) {
	ThrowError(pos, "parameter number does not match, expect `%d`, but there has `%d`", expect, now)
}

// ThrowIllegalCovertError 非法的类型转换
func ThrowIllegalCovertError(pos reader.Position, from, to hir.Type) {
	ThrowError(pos, "type `%s` can not covert to `%s`", from, to)
}

// ThrowNotArrayError 必须是数组
func ThrowNotArrayError(pos reader.Position, t hir.Type) {
	ThrowError(pos, "expect array type but there is `%s`", t)
}

// ThrowExpectPointerTypeError 期待指针类型
func ThrowExpectPointerTypeError(pos reader.Position) {
	ThrowError(pos, "expect a pointer type")
}

// ThrowExpectPointerError 期待指针
func ThrowExpectPointerError(pos reader.Position, t hir.Type) {
	ThrowError(pos, "expect pointer type but there is `%s`", t)
}

// ThrowExpectReferenceError 期待引用
func ThrowExpectReferenceError(pos reader.Position, t hir.Type) {
	ThrowError(pos, "expect reference type but there is `%s`", t)
}

// ThrowNotTupleError 必须是元组
func ThrowNotTupleError(pos reader.Position, t hir.Type) {
	ThrowError(pos, "expect tuple type but there is `%s`", t)
}

// ThrowNotStructError 必须是结构体
func ThrowNotStructError(pos reader.Position, t hir.Type) {
	ThrowError(pos, "expect struct type but there is `%s`", t)
}

// ThrowInvalidIndexError 超出下标
func ThrowInvalidIndexError(pos reader.Position, index uint) {
	ThrowError(pos, "invalid index with `%d`", index)
}

// ThrowLoopControlError 非法的循环控制
func ThrowLoopControlError(pos reader.Position) {
	ThrowError(pos, "must in a loop")
}

// ThrowMissingReturnValueError 缺失返回值
func ThrowMissingReturnValueError(pos reader.Position, t hir.Type) {
	ThrowError(pos, "missing a return value type `%s`", t)
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
func ThrowInvalidPackage(pos reader.Position, paths dynarray.DynArray[token.Token]) {
	pathStrs := iterator.Map[token.Token, string, dynarray.DynArray[string]](paths, func(v token.Token) string {
		return v.Source()
	})
	ThrowError(pos, "package `%s` is invalid", strings.Join(pathStrs.ToSlice(), "."))
}

// ThrowCanNotGetPointer 不能取指针
func ThrowCanNotGetPointer(pos reader.Position) {
	ThrowError(pos, "can not get the pointer of this expression")
}

// ThrowCircularReference 循环引用
func ThrowCircularReference(pos reader.Position, path token.Token) {
	ThrowError(pos, "circular reference `%s`", path.Source())
}
