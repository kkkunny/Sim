package global

type Global interface {
	Package() *Package
	Public() bool
}
