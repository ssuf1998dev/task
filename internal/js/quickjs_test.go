package js

import (
	"bytes"
	"testing"
	"unsafe"

	"github.com/samber/lo"
	"github.com/stretchr/testify/assert"
	"modernc.org/libc"
	"modernc.org/libquickjs"
)

func TestBasic(t *testing.T) {
	t.Parallel()

	qjs := lo.Must(NewQuickJS())
	defer qjs.Close()

	val := qjs.Eval("1+1")
	assert.Equal(t, int32(libquickjs.EJS_TAG_INT), tag(val))
	assert.Equal(t, 2, int(*(*int32)(unsafe.Pointer(&val))))

	val = qjs.Eval("(async()=>1+1)()", QJSEvalAwait(true))
	assert.Equal(t, int32(libquickjs.EJS_TAG_INT), tag(val))
	assert.Equal(t, 2, int(*(*int32)(unsafe.Pointer(&val))))
}

// func TestInterrupt(t *testing.T) {
// 	qjs, _ := NewQuickJS()
// 	defer qjs.Close()

// 	qjs.SetTimeoutInterrupt(3)

// 	var val = qjs.Eval("while(true){}")
// 	assert.Equal(t, tag(val), int32(libquickjs.EJS_TAG_EXCEPTION))
// }

func TestProcessEnv(t *testing.T) {
	t.Parallel()

	qjs, _ := NewQuickJS()
	defer qjs.Close()

	qjs.ProcessEnv(map[string]string{
		"foo": "bar",
	})
	val := qjs.Eval("process.env")
	assert.Equal(t, int32(libquickjs.EJS_TAG_OBJECT), tag(val))

	val = qjs.Eval("process.env.foo")
	assert.Equal(t, int32(libquickjs.EJS_TAG_STRING), tag(val))
	ptr := libquickjs.XToCString(qjs.tls, qjs.ctx, val)
	defer libquickjs.XJS_FreeCString(qjs.tls, qjs.ctx, ptr)
	assert.Equal(t, "bar", libc.GoString(ptr))

	qjs.ProcessEnv(map[string]string{
		"bar": "foo",
	})
	val = qjs.Eval("process.env.bar")
	assert.Equal(t, int32(libquickjs.EJS_TAG_STRING), tag(val))
	ptr = libquickjs.XToCString(qjs.tls, qjs.ctx, val)
	defer libquickjs.XJS_FreeCString(qjs.tls, qjs.ctx, ptr)
	assert.Equal(t, libc.GoString(ptr), "foo")
}

func TestLoadModule(t *testing.T) {
	t.Parallel()

	qjs, _ := NewQuickJS()
	defer qjs.Close()

	var buff bytes.Buffer
	qjs.Stdout = &buff

	qjs.LoadModule("export const foo = 'bar'", "foo")
	qjs.Eval("import {foo} from 'foo';print(foo);", QJSEvalAwait(true))
	assert.Contains(t, buff.String(), "bar")
}

func TestStd(t *testing.T) {
	t.Parallel()

	qjs, _ := NewQuickJS()
	defer qjs.Close()

	val := qjs.Eval("console.log")
	assert.Equal(t, int32(libquickjs.EJS_TAG_OBJECT), tag(val))

	val = qjs.Eval("print")
	assert.Equal(t, int32(libquickjs.EJS_TAG_OBJECT), tag(val))

	val = qjs.Eval("import { setTimeout } from 'os';setTimeout", QJSEvalFlagModule(true))
	assert.Equal(t, int32(libquickjs.EJS_TAG_OBJECT), tag(val))

	var buff bytes.Buffer
	qjs.Stdout = &buff
	qjs.Eval("print('foo')")
	assert.Contains(t, buff.String(), "foo")
}
