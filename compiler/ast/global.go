package ast

import "fmt"

type Global interface {
	global()
	fmt.Stringer
}

type FuncDecl struct {
	Name       string
	Params     []*ParamDecl
	ReturnType Type
}

func (f *FuncDecl) global() {}

func (f *FuncDecl) String() string {
	params := ""
	for i, p := range f.Params {
		if i > 0 {
			params += ", "
		}
		params += p.String()
	}
	ret := ""
	if f.ReturnType != nil {
		ret = ": " + f.ReturnType.String()
	}
	return fmt.Sprintf("FuncDecl(Name: %s, Params: [%s]%s)", f.Name, params, ret)
}
