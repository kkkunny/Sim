//go:build analyse

package main

import (
	"fmt"
	"os"
	"reflect"

	stlerror "github.com/kkkunny/stl/error"

	"github.com/kkkunny/Sim/analyse"
	"github.com/kkkunny/Sim/lex"
	"github.com/kkkunny/Sim/mean"
	"github.com/kkkunny/Sim/parse"
	"github.com/kkkunny/Sim/reader"
)

func main() {
	r := stlerror.MustWith(reader.NewReaderFromFile(os.Args[1]))
	analyser := analyse.New(parse.New(lex.New(r)))
	analyser.Analyse().Iterator().Foreach(func(v mean.Global) bool {
		fmt.Println(reflect.TypeOf(v).String())
		return true
	})
}
