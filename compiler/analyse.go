//go:build analyse

package main

import (
	"fmt"
	"os"
	"reflect"

	"github.com/kkkunny/stl/container/iterator"
	stlerror "github.com/kkkunny/stl/error"
	stlos "github.com/kkkunny/stl/os"

	"github.com/kkkunny/Sim/analyse"
	"github.com/kkkunny/Sim/hir"
)

func main() {
	means := stlerror.MustWith(analyse.Analyse(stlos.NewFilePath(os.Args[1])))
	iterator.Foreach(means, func(v hir.Global) bool {
		fmt.Println(reflect.TypeOf(v).String())
		return true
	})
}
