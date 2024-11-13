//go:build analyse

package main

import (
	"os"

	stlerror "github.com/kkkunny/stl/error"
	stlos "github.com/kkkunny/stl/os"

	"github.com/kkkunny/Sim/compiler/analyse"
)

func main() {
	stlerror.MustWith(analyse.Analyse(stlos.NewFilePath(os.Args[1])))
	// stliter.Foreach(res.Globals, func(v oldhir.Global) bool {
	// 	fmt.Println(reflect.TypeOf(v).String())
	// 	return true
	// })
}
