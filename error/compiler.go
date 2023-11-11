package errors

import (
	"fmt"
	"os"

	"github.com/kkkunny/Sim/mean"
	"github.com/kkkunny/Sim/reader"
	"github.com/kkkunny/Sim/token"
)

// ThrowError 抛出异常
func ThrowError(pos reader.Position, format string, args ...any) {
	_, err := fmt.Fprintf(os.Stderr, format, args...)
	if err != nil {
		panic(err)
	}
	panic("compiler error")
}

// ThrowTypeMismatchError 类型不匹配
func ThrowTypeMismatchError(pos reader.Position, t1, t2 mean.Type) {
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
func ThrowIllegalBinaryError(pos reader.Position, op token.Token, left, right mean.Expr) {
	ThrowError(pos, "illegal binary operation with type `%s` `%s` `%s`", left.GetType(), op.Source(), right.GetType())
}

// ThrowIllegalUnaryError 非法的一元运算
func ThrowIllegalUnaryError(pos reader.Position, op token.Token, value mean.Expr) {
	ThrowError(pos, "illegal unary operation with `%s` type `%s`", op.Source(), value.GetType())
}

// ThrowNotFunctionError 必须是函数
func ThrowNotFunctionError(pos reader.Position, t mean.Type) {
	ThrowError(pos, "expect function type but there is `%s`", t)
}

// ThrowParameterNumberNotMatchError 参数数量不匹配
func ThrowParameterNumberNotMatchError(pos reader.Position, expect, now uint) {
	ThrowError(pos, "parameter number does not match, expect `%d`, but there has `%d`", expect, now)
}

// ThrowIllegalCovertError 非法的类型转换
func ThrowIllegalCovertError(pos reader.Position, from, to mean.Type) {
	ThrowError(pos, "type `%s` can not covert to `%s`", from, to)
}

// ThrowNotArrayError 必须是数组
func ThrowNotArrayError(pos reader.Position, t mean.Type) {
	ThrowError(pos, "expect array type but there is `%s`", t)
}

// ThrowNotTupleError 必须是元组
func ThrowNotTupleError(pos reader.Position, t mean.Type) {
	ThrowError(pos, "expect tuple type but there is `%s`", t)
}

// ThrowNotStructError 必须是结构体
func ThrowNotStructError(pos reader.Position, t mean.Type) {
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
func ThrowMissingReturnValueError(pos reader.Position, t mean.Type) {
	ThrowError(pos, "missing a return value type `%s`", t)
}
