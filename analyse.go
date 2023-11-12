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
	"github.com/kkkunny/Sim/lex"
	"github.com/kkkunny/Sim/mean"
	"github.com/kkkunny/Sim/parse"
	"github.com/kkkunny/Sim/reader"
	"github.com/kkkunny/Sim/util"
)

func main() {
	util.InitLLVM()
	stlerror.Must(llvm.InitializeNativeTarget())
	target := stlerror.MustWith(llvm.NativeTarget())
	mean.Usize.Bits = target.PointerSize() * 8
	mean.Isize.Bits = mean.Usize.Bits
	f, r := stlerror.MustWith2(reader.NewReaderFromFile(os.Args[1]))
	defer f.Close()
	analyser := analyse.New(parse.New(lex.New(r)))
	iterator.Foreach(analyser.Analyse(), func(v mean.Global) bool {
		fmt.Println(reflect.TypeOf(v).String())
		return true
	})
}
