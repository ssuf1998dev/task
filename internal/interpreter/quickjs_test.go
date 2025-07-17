package interpreter

import (
	"testing"
	"unsafe"

	"github.com/samber/lo"
	"github.com/stretchr/testify/assert"
	"modernc.org/libc"
	"modernc.org/libquickjs"
)

func TestBasic(t *testing.T) {
	qjs := lo.Must(NewQuickJSInterpreter())
	defer qjs.Close()

	val := qjs.Eval("1+1")
	assert.Equal(t, tag(val), int32(libquickjs.EJS_TAG_INT))
	assert.Equal(t, int(*(*int32)(unsafe.Pointer(&val))), 2)

	val = qjs.Eval("(async()=>1+1)()", QJSEvalAwait(true))
	assert.Equal(t, tag(val), int32(libquickjs.EJS_TAG_INT))
	assert.Equal(t, int(*(*int32)(unsafe.Pointer(&val))), 2)
}

// func TestInterrupt(t *testing.T) {
// 	qjs, _ := NewQuickJSInterpreter()
// 	defer qjs.Close()

// 	qjs.SetTimeoutInterrupt(3)

// 	var val = qjs.Eval("while(true){}")
// 	assert.Equal(t, tag(val), int32(libquickjs.EJS_TAG_EXCEPTION))
// }

func TestProcessEnv(t *testing.T) {
	qjs, _ := NewQuickJSInterpreter()
	defer qjs.Close()

	qjs.ProcessEnv(map[string]string{
		"foo": "bar",
	})
	val := qjs.Eval("process.env")
	assert.Equal(t, tag(val), int32(libquickjs.EJS_TAG_OBJECT))

	val = qjs.Eval("process.env.foo")
	assert.Equal(t, tag(val), int32(libquickjs.EJS_TAG_STRING))
	ptr := libquickjs.XToCString(qjs.tls, qjs.ctx, val)
	defer libquickjs.XJS_FreeCString(qjs.tls, qjs.ctx, ptr)
	assert.Equal(t, libc.GoString(ptr), "bar")

	qjs.ProcessEnv(map[string]string{
		"bar": "foo",
	})
	val = qjs.Eval("process.env.bar")
	assert.Equal(t, tag(val), int32(libquickjs.EJS_TAG_STRING))
	ptr = libquickjs.XToCString(qjs.tls, qjs.ctx, val)
	defer libquickjs.XJS_FreeCString(qjs.tls, qjs.ctx, ptr)
	assert.Equal(t, libc.GoString(ptr), "foo")
}

func TestLoadModule(t *testing.T) {
	qjs, _ := NewQuickJSInterpreter()
	defer qjs.Close()

	qjs.LoadModule("export const foo = 'bar'", "foo")
	val := qjs.Eval("(async()=>{const {foo} = await import('foo');return foo;})()", QJSEvalAwait(true))
	assert.Equal(t, tag(val), int32(libquickjs.EJS_TAG_STRING))
	ptr := libquickjs.XToCString(qjs.tls, qjs.ctx, val)
	defer libquickjs.XJS_FreeCString(qjs.tls, qjs.ctx, ptr)
	assert.Equal(t, libc.GoString(ptr), "bar")
}
