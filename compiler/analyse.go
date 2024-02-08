//go:build analyse

package main

import (
	"fmt"
	"os"
	"reflect"

	stliter "github.com/kkkunny/stl/container/iter"
	stlerror "github.com/kkkunny/stl/error"
	stlos "github.com/kkkunny/stl/os"

	"github.com/kkkunny/Sim/analyse"
	"github.com/kkkunny/Sim/hir"
)

func main() {
	res := stlerror.MustWith(analyse.Analyse(stlos.NewFilePath(os.Args[1])))
	stliter.Foreach(res.Globals, func(v hir.Global) bool {
		fmt.Println(reflect.TypeOf(v).String())
		return true
	})
}
