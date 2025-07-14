package util

import stlhash "github.com/kkkunny/stl/hash"

// StringHashFunc 字符串hasher
var StringHashFunc = stlhash.GetMapHashFunc[string]()
