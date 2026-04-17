package report

import (
	"reflect"

	"github.com/kkkunny/stl/enum"
)

type Error string

var Errors = enum.New[struct {
	ExpectedToken     Error `enum:"expected token" format:"expected '%s' but got '%s'"`
	UnexpectedToken   Error `enum:"unexpected token" format:"unexpected token '%s'"`
	UnknownIdentifier Error `enum:"unknown identifier" format:"unknown identifier '%s'"`
	UnexpectedType    Error `enum:"unexpected type" format:"expected type '%s' but got '%s'"`
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
