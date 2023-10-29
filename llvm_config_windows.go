//go:build windows

package main

// #cgo CFLAGS: -IC:/msys64/clang64/include  -D_FILE_OFFSET_BITS=64 -D__STDC_CONSTANT_MACROS -D__STDC_FORMAT_MACROS -D__STDC_LIMIT_MACROS
// #cgo CXXFLAGS: -IC:/msys64/clang64/include -std=c++17 -stdlib=libc++  -fno-exceptions -D_FILE_OFFSET_BITS=64 -D__STDC_CONSTANT_MACROS -D__STDC_FORMAT_MACROS -D__STDC_LIMIT_MACROS
// #cgo LDFLAGS: -LC:/msys64/clang64/lib -lLLVM-17
import "C"