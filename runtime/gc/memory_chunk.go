package gc

import "unsafe"

type memoryChunk struct {
	area   *memoryArea
	offset uint
	busy   bool
	prev   *memoryChunk
	next   *memoryChunk
}

func (b *memoryChunk) ptr() unsafe.Pointer {
	return unsafe.Pointer(uintptr(b.area.address) + uintptr(b.offset))
}

func (b *memoryChunk) size() uint {
	if b.next != nil {
		return b.next.offset - b.offset
	}
	return b.area.size - b.offset
}

func (b *memoryChunk) split(size uint) *memoryChunk {
	newBlock := &memoryChunk{
		area:   b.area,
		offset: b.offset,
		busy:   false,
		prev:   b.prev,
		next:   b,
	}
	b.offset += size
	if b.prev != nil {
		b.prev.next = newBlock
	}
	b.prev = newBlock
	if b.area.chunks == b {
		b.area.chunks = newBlock
	}
	return newBlock
}

func (b *memoryChunk) free() error {
	switch {
	case b.prev != nil && !b.prev.busy:
		b.prev.next = b.next
		if b.next != nil {
			b.next.prev = b.prev
		}
	case b.next != nil && !b.next.busy:
		if b.next.next != nil {
			b.next.next.prev = b
		}
		b.next = b.next.next
		b.busy = false
	default:
		b.busy = false
	}
	return nil
}
