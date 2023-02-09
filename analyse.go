//go:build test && analyse

package main

import (
	"encoding/json"
	"fmt"
	"os"

	"github.com/kkkunny/Sim/src/compiler/analyse"
	"github.com/kkkunny/Sim/src/compiler/parse"
	stlos "github.com/kkkunny/stl/os"
	"github.com/kkkunny/stl/util"
)

func main() {
	ast := util.MustValue(parse.Parse(stlos.Path(os.Args[1])))
	mean := util.MustValue(analyse.Analyse(ast))
	out := string(util.MustValue(json.MarshalIndent(mean, "", "  ")))
	fmt.Println(out)
}
