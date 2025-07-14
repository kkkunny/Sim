package gc

/*
#include <stdlib.h>
#include <stdint.h>
#include <string.h>
*/
import "C"
import (
	"errors"
	"unsafe"
)

func malloc(size uint) (unsafe.Pointer, error) {
	ptr := unsafe.Pointer(C.malloc(C.size_t(size)))
	if ptr == nil {
		return nil, errors.New("can not allocate in heap")
	}
	return ptr, nil
}

func calloc(count uint, size uint) (unsafe.Pointer, error) {
	ptr := unsafe.Pointer(C.calloc(C.size_t(count), C.size_t(size)))
	if ptr == nil {
		return nil, errors.New("can not allocate in heap")
	}
	return ptr, nil
}

func realloc(ptr unsafe.Pointer, size uint) (unsafe.Pointer, error) {
	ptr = unsafe.Pointer(C.realloc(ptr, C.size_t(size)))
	if ptr == nil {
		return nil, errors.New("can not allocate in heap")
	}
	return ptr, nil
}

func free(ptr unsafe.Pointer) error {
	if ptr == nil {
		return errors.New("can not free empty pointer")
	}
	C.free(ptr)
	return nil
}

func memset(ptr unsafe.Pointer, value uint8, num uint) unsafe.Pointer {
	return unsafe.Pointer(C.memset(ptr, C.int(value), C.size_t(num)))
}
