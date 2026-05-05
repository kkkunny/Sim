package report

import (
	"reflect"

	"github.com/kkkunny/stl/enum"
)

type Error string

var Errors = enum.New[struct {
	ExpectedToken            Error `enum:"expected token" format:"expected '%s' but got '%s'"`
	UnexpectedToken          Error `enum:"unexpected token" format:"unexpected token '%s'"`
	UnknownIdentifier        Error `enum:"unknown identifier" format:"unknown identifier '%s'"`
	RepeatedIdentifier       Error `enum:"repeated identifier" format:"the identifier '%s' be redefined"`
	UnexpectedExpression     Error `enum:"unexpected expression" format:"expected expression type '%s' but got '%s'"`
	UnexpectedExpressionType Error `enum:"unexpected expression type" format:"expected a %s expression but got '%s'"`
	InsufficientArguments    Error `enum:"insufficient arguments" format:"expected %d arguments but got %d"`
	ExpectedIntegerConstant  Error `enum:"expected integer constant" format:"expected a integer constant"`
	MustMutable              Error `enum:"must mutable" format:"the expression must be mutable"`
	MustImmutable            Error `enum:"must immutable" format:"the expression must be immutable"`
	MustNotTemporary         Error `enum:"must not temporary" format:"the expression must be not temporary"`
	MissingType              Error `enum:"missing type" format:"the expression must have a type"`
	TypeMissingDefaultValue  Error `enum:"missing default value" format:"the type '%s' missing a default value"`
	InvalidType              Error `enum:"invalid type" format:"expect a valid type"`
	InvalidTypeCovert        Error `enum:"invalid type covert" format:"the type '%s' can not covert to '%s'"`
	InvalidMainFunction      Error `enum:"invalid main function" format:"the global 'main' must be a function"`
	InvalidRecursionType     Error `enum:"invalid recursion type" format:"the type is invalid because of recursion"`
	UnknownAttribute         Error `enum:"unknown attribute" format:"unknown attribute '%s'"`
	InvalidAttribute         Error `enum:"invalid attribute" format:"the attribute '%s' cannot be used for the '%s'"`
	InvalidChar              Error `enum:"invalid char" format:"char '%s' is invalid"`
}]()

var errorFormats = func() map[Error]string {
	v := reflect.ValueOf(Errors)
	t := v.Type()
	res := make(map[Error]string, v.NumField())
	for i := 0; i < v.NumField(); i++ {
		value := v.Field(i).Interface().(Error)
		text, _ := t.Field(i).Tag.Lookup("format")
		res[value] = text
	}
	return res
}()
