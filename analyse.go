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
	f, r := stlerror.MustWith2(reader.NewReaderFromFile(os.Args[1]))
	defer f.Close()
	stlerror.Must(llvm.InitializeNativeTarget())
	target := stlerror.MustWith(llvm.NativeTarget())
	mean.Usize.Bits = target.PointerSize() * 8
	mean.Isize.Bits = mean.Usize.Bits
	analyser := analyse.New(target, parse.New(lex.New(r)))
	analyser.Analyse().Iterator().Foreach(func(v mean.Global) bool {
		fmt.Println(reflect.TypeOf(v).String())
		return true
	})
}
