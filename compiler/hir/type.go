package hir

type Type interface {
	printWriter
	typ()
}

var (
	I32 = &IntType{}
)

type IntType struct{}

func (*IntType) typ() {}

func (t *IntType) print(p *printer) {
	p.WriteString("i32")
}

var Unit = &UnitType{}

type UnitType struct{}

func (*UnitType) typ() {}

func (t *UnitType) print(p *printer) {
	p.WriteString("unit")
}
