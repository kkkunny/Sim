package types

type wrapper interface {
	Wrap(inner Type) BuildInType
}

func wrap[To Type](from Type, inner To) To {
	return from.(wrapper).Wrap(inner).(To)
}
