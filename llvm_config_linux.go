//go:build linux

package main

// #cgo CFLAGS: -I/usr/include  -D_GNU_SOURCE -D__STDC_CONSTANT_MACROS -D__STDC_FORMAT_MACROS -D__STDC_LIMIT_MACROS
// #cgo CXXFLAGS: -I/usr/include -std=c++17   -fno-exceptions -D_GNU_SOURCE -D__STDC_CONSTANT_MACROS -D__STDC_FORMAT_MACROS -D__STDC_LIMIT_MACROS
// #cgo LDFLAGS: -L/usr/lib -lLLVM-16
import "C"
