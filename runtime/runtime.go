package runtime

/*
#include "runtime.h"
*/
import "C"
import "unsafe"

var StrEqStr unsafe.Pointer = C.sim_runtime_str_eq_str
