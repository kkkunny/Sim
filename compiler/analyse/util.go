package analyse

import (
	"github.com/kkkunny/stl/container/linkedlist"
	stlerror "github.com/kkkunny/stl/error"

	"github.com/kkkunny/Sim/mean"
	"github.com/kkkunny/Sim/parse"
)

// Analyse 语义分析
func Analyse(path string) (linkedlist.LinkedList[mean.Global], stlerror.Error) {
	asts, err := parse.Parse(path)
	if err != nil{
		return linkedlist.LinkedList[mean.Global]{}, err
	}
	return New(asts).Analyse(), nil
}

// 语义分析子包
func analyseSonPackage(parent *Analyser, path string) (linkedlist.LinkedList[mean.Global], *_PkgScope, stlerror.Error) {
	asts, err := parse.Parse(path)
	if err != nil{
		return linkedlist.LinkedList[mean.Global]{}, nil, err
	}
	analyser := newSon(parent, asts)
	return analyser.Analyse(), analyser.pkgScope, nil
}
