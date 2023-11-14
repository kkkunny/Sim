//go:build analyse

package main

import (
	"fmt"
	"os"
	"reflect"

	"github.com/kkkunny/go-llvm"
	"github.com/kkkunny/stl/container/iterator"
	stlerror "github.com/kkkunny/stl/error"

	"github.com/kkkunny/Sim/analyse"
	"github.com/kkkunny/Sim/mean"
	"github.com/kkkunny/Sim/parse"
)

func main() {
	_ = util.Logger.Infof(0, "LLVM VERSION: %s", llvm.Version)
	stlerror.Must(llvm.InitializeNativeTarget())
	target := stlerror.MustWith(llvm.NativeTarget())

	asts := stlerror.MustWith(parse.ParseFile(os.Args[1]))
	analyser := analyse.New(asts, target)
	iterator.Foreach(analyser.Analyse(), func(v mean.Global) bool {
		fmt.Println(reflect.TypeOf(v).String())
		return true
	})
}
