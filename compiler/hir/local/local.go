package local

import "github.com/kkkunny/stl/list"

type Local interface {
	setPosition(pos *list.Element[Local])
	position() (*list.Element[Local], bool)
}
