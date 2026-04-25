package cir

type Namer interface {
	SetName(s string)
	GetName() string
}
