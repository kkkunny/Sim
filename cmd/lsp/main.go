package main

import (
	"github.com/kkkunny/Sim/lsp"
)

func main() {
	err := lsp.StartServer()
	if err != nil {
		panic(err)
	}
}
