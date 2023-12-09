package mir

type Target interface {
	Equal(t Target) bool
}
