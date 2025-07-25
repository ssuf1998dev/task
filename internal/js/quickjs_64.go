//go:build !(386 || arm)

package js

import "modernc.org/libquickjs"

// var JS_NULL      = libquickjs.TJSValue{Ftag: libquickjs.EJS_TAG_NULL}
var JS_UNDEFINED = libquickjs.TJSValue{Ftag: libquickjs.EJS_TAG_UNDEFINED}

func tag(v libquickjs.TJSValue) int32 {
	return int32(v.Ftag)
}
