package js

import (
	"fmt"
	"strings"
	"unsafe"

	"github.com/samber/lo"
	"modernc.org/libc"
	"modernc.org/libquickjs"
)

func fp(f any) uintptr {
	type iface [2]uintptr
	return (*iface)(unsafe.Pointer(&f))[1]
}

func (i *QuickJS) jsPrint(tls *libc.TLS, ctx uintptr, thisVal libquickjs.TJSValue, argc int32, argv uintptr) (r libquickjs.TJSValue) {
	outputs := []string{}
	for i := range argc {
		offset := uintptr(i) * uintptr(INT_8_16)

		value := *(*libquickjs.TJSValue)(unsafe.Pointer(argv + offset))
		switch tag(value) {
		case libquickjs.EJS_TAG_STRING, libquickjs.EJS_TAG_STRING_ROPE:
			p := libquickjs.XToCString(tls, ctx, value)
			defer libquickjs.XJS_FreeCString(tls, ctx, p)
			outputs = append(outputs, libc.GoString(p))
		case libquickjs.EJS_TAG_INT:
			outputs = append(outputs, fmt.Sprint(*(*int32)(unsafe.Pointer(&value))))
		case libquickjs.EJS_TAG_BOOL:
			outputs = append(outputs, lo.Ternary(*(*int32)(unsafe.Pointer(&value)) != 0, "true", "false"))
		case libquickjs.EJS_TAG_NULL:
			outputs = append(outputs, "null")
		case libquickjs.EJS_TAG_UNDEFINED:
			outputs = append(outputs, "undefined")
		default:
			json := libquickjs.XJS_JSONStringify(tls, ctx, value, JS_UNDEFINED, JS_UNDEFINED)
			defer libquickjs.XFreeValue(tls, ctx, json)
			p := libquickjs.XToCString(tls, ctx, json)
			defer libquickjs.XJS_FreeCString(tls, ctx, p)
			outputs = append(outputs, libc.GoString(p))
		}
	}

	if len(outputs) > 0 {
		_, _ = i.Stdout.Write([]byte(strings.Join(outputs, "\n") + "\n"))
	}

	return JS_UNDEFINED
}
