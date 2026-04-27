//go:build codegen

package main

import (
	"fmt"
	"os"

	"github.com/kkkunny/Sim/compiler/analyze"
	"github.com/kkkunny/Sim/compiler/codegen"
)

func main() {
	pkg, err := analyze.Analyze(os.Args[1])
	if err != nil {
		panic(err)
	}
	builder := codegen.New().Generate(pkg)
	fmt.Println(builder)
}
