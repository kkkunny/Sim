package cir

type Global interface {
	global()
}

type FuncDecl struct {
	Name       string
	Params     []*ParamDecl
	ReturnType Type
	Body       *Block
}

func (*FuncDecl) global() {}
