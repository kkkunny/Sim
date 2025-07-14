package gc

import "unsafe"

type chunk struct {
	size uint
	ptr  unsafe.Pointer
}
