//go:build analyse

package main

import (
	"fmt"
	"os"
	"path/filepath"
	"reflect"

	"github.com/kkkunny/go-llvm"
	"github.com/kkkunny/stl/container/iterator"
	stlerror "github.com/kkkunny/stl/error"

	"github.com/kkkunny/Sim/analyse"
	"github.com/kkkunny/Sim/mean"
	"github.com/kkkunny/Sim/parse"
	"github.com/kkkunny/Sim/util"
)

func main() {
	_ = util.Logger.Infof(0, "LLVM VERSION: %s", llvm.Version)
	stlerror.Must(llvm.InitializeNativeTarget())
	target := stlerror.MustWith(llvm.NativeTarget())

	path := stlerror.MustWith(filepath.Abs(os.Args[1]))
	asts := stlerror.MustWith(parse.ParseFile(path))
	analyser := analyse.New(asts, target)
	iterator.Foreach(analyser.Analyse(), func(v mean.Global) bool {
		fmt.Println(reflect.TypeOf(v).String())
		return true
	})
}
