package gc

import "unsafe"

type memoryArea struct {
	size    uint
	address unsafe.Pointer
	chunks  *memoryChunk
}

func newMemoryArea(size uint) (*memoryArea, error) {
	ptr, err := malloc(size)
	if err != nil {
		return nil, err
	}
	area := &memoryArea{
		size:    initSize,
		address: ptr,
	}
	area.chunks = &memoryChunk{area: area}
	return area, nil
}

func (a *memoryArea) foreachChunks(f func(*memoryChunk) bool) {
	for cursor := a.chunks; cursor != nil; cursor = cursor.next {
		if !f(cursor) {
			break
		}
	}
	return
}
