//go:build !(386 || arm)

package js

import "modernc.org/libquickjs"

// var JS_NULL      = libquickjs.TJSValue{Ftag: libquickjs.EJS_TAG_NULL}
var JS_UNDEFINED = libquickjs.TJSValue{Ftag: libquickjs.EJS_TAG_UNDEFINED}

var INT_8_16 = 16

func tag(v libquickjs.TJSValue) int32 {
	return int32(v.Ftag)
}
