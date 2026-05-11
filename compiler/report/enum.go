package report

import (
	"reflect"

	"github.com/kkkunny/stl/enum"
)

type ErrorType string

var Errors = enum.New[struct {
	ExpectedToken                ErrorType `enum:"expected token" format:"expected '%s' but got '%s'"`
	UnexpectedToken              ErrorType `enum:"unexpected token" format:"unexpected token '%s'"`
	UnknownIdentifier            ErrorType `enum:"unknown identifier" format:"unknown identifier '%s'"`
	RepeatedIdentifier           ErrorType `enum:"repeated identifier" format:"the identifier '%s' be redefined"`
	UnexpectedExpression         ErrorType `enum:"unexpected expression" format:"expected expression type '%s' but got '%s'"`
	UnexpectedExpressionCategory ErrorType `enum:"unexpected expression category" format:"expected a %s but got '%s'"`
	UnexpectedTypeCategory       ErrorType `enum:"unexpected type category" format:"expected a %s type but got '%s'"`
	InsufficientArguments        ErrorType `enum:"insufficient arguments" format:"expected %d arguments but got %d"`
	ExpectedIntegerConstant      ErrorType `enum:"expected integer constant" format:"expected a integer constant"`
	MustMutable                  ErrorType `enum:"must mutable" format:"the expression must be mutable"`
	MustImmutable                ErrorType `enum:"must immutable" format:"the expression must be immutable"`
	MustNotTemporary             ErrorType `enum:"must not temporary" format:"the expression must be not temporary"`
	MissingType                  ErrorType `enum:"missing type" format:"the expression must have a type"`
	TypeMissingDefaultValue      ErrorType `enum:"missing default value" format:"the type '%s' missing a default value"`
	InvalidType                  ErrorType `enum:"invalid type" format:"expect a valid type"`
	InvalidTypeCovert            ErrorType `enum:"invalid type covert" format:"the type '%s' can not covert to '%s'"`
	InvalidMainFunction          ErrorType `enum:"invalid main function" format:"the global 'main' must be a function"`
	CircularReference            ErrorType `enum:"circular reference" format:"circular reference"`
	UnknownAttribute             ErrorType `enum:"unknown attribute" format:"unknown attribute '%s'"`
	InvalidAttribute             ErrorType `enum:"invalid attribute" format:"the attribute '%s' cannot be used for the '%s'"`
	InvalidChar                  ErrorType `enum:"invalid char" format:"char '%s' is invalid"`
	UnexpectedSelfPosition       ErrorType `enum:"unexpected self position" format:"the parameter 'self' must be in the first position of parameters"`
	UnexpectedSelfType           ErrorType `enum:"unexpected self type" format:"the type of parameter 'self' must be 'Self' or '&Self' or '&mut Self'"`
}]()

var errorFormats = func() map[ErrorType]string {
	v := reflect.ValueOf(Errors)
	t := v.Type()
	res := make(map[ErrorType]string, v.NumField())
	for i := 0; i < v.NumField(); i++ {
		value := v.Field(i).Interface().(ErrorType)
		text, _ := t.Field(i).Tag.Lookup("format")
		res[value] = text
	}
	return res
}()
