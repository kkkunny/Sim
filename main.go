package main

import (
	"fmt"

	stlerror "github.com/kkkunny/stl/error"

	"github.com/kkkunny/Sim/lexer"
	"github.com/kkkunny/Sim/reader"
	"github.com/kkkunny/Sim/token"
)

func main() {
	r := stlerror.MustWith(reader.NewReaderFromFile("example/main.sim"))
	l := lexer.New(r)
	for tok := stlerror.MustWith(l.Scan()); !tok.Is(token.Eof); tok = stlerror.MustWith(l.Scan()) {
		fmt.Println(tok.Kind, tok.Source())
	}
}
