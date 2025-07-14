package gc

import (
	"fmt"
	"testing"
	"unsafe"
)

var usedPtr unsafe.Pointer

func TestAlloca(t *testing.T) {
	var err error
	usedPtr, err = Alloca(10)
	if err != nil {
		t.FailNow()
	}
	err = GC()
	if err != nil {
		t.FailNow()
	}
	fmt.Println(usedPtr)
}
