package gc

import (
	"encoding/binary"
	"errors"
	"unsafe"

	stlval "github.com/kkkunny/stl/value"
)

const initSize = 1024

type GarbageCollector struct {
	stackBegin uintptr
	areas      []*memoryArea
}

func newGarbageCollector(stackBegin uintptr) (*GarbageCollector, error) {
	area, err := newMemoryArea(initSize)
	if err != nil {
		return nil, err
	}
	return &GarbageCollector{
		stackBegin: stackBegin,
		areas:      []*memoryArea{area},
	}, nil
}

// 扫描并且分配空闲内容
func (c *GarbageCollector) scanAndThenAllocate(size uint) (unsafe.Pointer, bool, error) {
	if size == 0 {
		return nil, false, errors.New("can not allocate size 0")
	}

	var freeChunk *memoryChunk
	for _, area := range c.areas {
		if freeChunk != nil {
			break
		}
		area.foreachChunks(func(chunk *memoryChunk) bool {
			if chunk.busy {
				return true
			}
			if chunk.size() == size {
				freeChunk = chunk
				return false
			} else if chunk.size() > size {
				freeChunk = chunk.split(size)
				return false
			}
			return true
		})
	}
	if freeChunk == nil {
		return nil, false, nil
	}

	freeChunk.busy = true
	return memset(freeChunk.ptr(), 0, freeChunk.size()), true, nil
}

// 分配内存
func (c *GarbageCollector) allocate(size uint, stackPos uintptr) (unsafe.Pointer, error) {
	ptr, ok, err := c.scanAndThenAllocate(size)
	if err != nil {
		return nil, err
	}
	if ok {
		return ptr, nil
	}

	// 找不到空位就gc
	err = c.gc(stackPos)
	if err != nil {
		return nil, err
	}

	ptr, ok, err = c.scanAndThenAllocate(size)
	if err != nil {
		return nil, err
	}
	if ok {
		return ptr, nil
	}

	// 还找不到空位就扩容
	err = c.expansion(size)
	if err != nil {
		return nil, err
	}

	ptr, ok, err = c.scanAndThenAllocate(size)
	if err != nil {
		return nil, err
	}
	if ok {
		return ptr, nil
	}
	return nil, errors.New("no space for allocation")
}

// 垃圾回收
func (c *GarbageCollector) gc(stackPos uintptr) error {
	busyBlocks := make(map[unsafe.Pointer]*memoryChunk)
	for _, area := range c.areas {
		area.foreachChunks(func(chunk *memoryChunk) bool {
			if !chunk.busy {
				return true
			}
			busyBlocks[chunk.ptr()] = chunk
			return true
		})
	}

	var endian binary.ByteOrder = binary.BigEndian
	if IsLittleEndian() {
		endian = binary.LittleEndian
	}

	refBlocks := make(map[unsafe.Pointer]*memoryChunk, len(busyBlocks))
	var scanAddressFn func(begin, end uintptr)
	scanAddressFn = func(begin, end uintptr) {
		for cursor := begin; cursor <= end; cursor++ {
			bytes := *(*[8]byte)(unsafe.Pointer(cursor))
			scanPtr := unsafe.Pointer(uintptr(endian.Uint64(bytes[:])))
			block, ok := busyBlocks[scanPtr]
			if !ok {
				continue
			}
			if _, ok := refBlocks[scanPtr]; ok {
				continue
			}
			refBlocks[scanPtr] = block
			scanAddressFn(uintptr(block.ptr()), uintptr(block.ptr())+uintptr(block.size())-1)
		}
	}

	// 标记
	// 栈
	if stackPos >= c.stackBegin {
		scanAddressFn(c.stackBegin, stackPos)
	} else {
		scanAddressFn(stackPos, c.stackBegin)
	}

	// 清除
	unusedBlocks := make(map[unsafe.Pointer]*memoryChunk, len(busyBlocks))
	for ptr, block := range busyBlocks {
		if _, ok := refBlocks[ptr]; ok {
			continue
		}
		unusedBlocks[ptr] = block
	}
	for _, block := range unusedBlocks {
		err := block.free()
		if err != nil {
			return err
		}
	}
	return nil
}

func (c *GarbageCollector) size() (size uint) {
	for _, area := range c.areas {
		size += area.size
	}
	return size
}

// 扩容
func (c *GarbageCollector) expansion(expectSize uint) error {
	size := c.size()
	size = max(size+stlval.ValueOr(size/2, 1), expectSize)
	area, err := newMemoryArea(size)
	if err != nil {
		return err
	}
	c.areas = append(c.areas, area)
	return nil
}
