package hir

type Type interface {
	printWriter
	typ()
}

var Unit = &UnitType{}

type UnitType struct{}

func (*UnitType) typ() {}

func (t *UnitType) print(p *printer) {
	p.WriteString("unit")
}

var (
	I32 = &IntType{}
)

type IntType struct{}

func (*IntType) typ() {}

func (t *IntType) print(p *printer) {
	p.WriteString("i32")
}

type FuncType struct {
	Return Type
	Params []Type
}

func NewFuncType(ret Type, params ...Type) *FuncType {
	return &FuncType{
		Return: ret,
		Params: params,
	}
}

func (*FuncType) typ() {}

func (t *FuncType) print(p *printer) {
	p.WriteString("(")
	for _, param := range t.Params {
		p.WriteBy(param)
		if param != t.Params[len(t.Params)-1] {
			p.WriteString(", ")
		}
	}
	p.WriteString(")")
	if t.Return != Unit {
		p.WriteString(" -> ")
		p.WriteBy(t.Return)
	}
}
