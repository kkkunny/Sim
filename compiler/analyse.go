//go:build analyse

package main

import (
	"fmt"
	"os"
	"path/filepath"
	"reflect"

	"github.com/kkkunny/stl/container/iterator"
	stlerror "github.com/kkkunny/stl/error"

	"github.com/kkkunny/Sim/analyse"
	"github.com/kkkunny/Sim/mean"
)

func main() {
	path := stlerror.MustWith(filepath.Abs(os.Args[1]))
	means := stlerror.MustWith(analyse.Analyse(path))
	iterator.Foreach(means, func(v mean.Global) bool {
		fmt.Println(reflect.TypeOf(v).String())
		return true
	})
}
