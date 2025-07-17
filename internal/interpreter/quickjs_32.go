//go:build 386 || arm

package interpreter

import (
	"modernc.org/libc"
	"modernc.org/libquickjs"
)

// var JS_NULL      = uint64(libquickjs.EJS_TAG_NULL << 32)
var JS_UNDEFINED = uint64(libquickjs.EJS_TAG_UNDEFINED << 32)

func tag(v libquickjs.TJSValue) (r int32) {
	if r = int32(v >> 32); uint32(r)-libc.Uint32FromInt32(libquickjs.EJS_TAG_FIRST) >= libquickjs.EJS_TAG_FLOAT64-libquickjs.EJS_TAG_FIRST {
		r = libquickjs.EJS_TAG_FLOAT64
	}
	return r
}
