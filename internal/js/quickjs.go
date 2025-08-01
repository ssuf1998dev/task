package js

import (
	"fmt"
	"io"
	"os"
	"unsafe"

	"github.com/samber/lo"
	"modernc.org/libc"
	"modernc.org/libquickjs"
)

type QuickJS struct {
	tls    *libc.TLS
	rt     uintptr
	ctx    uintptr
	Stdout io.Writer
	Stderr io.Writer
}

func (i *QuickJS) prepare() {
	libquickjs.Xjs_std_add_helpers(i.tls, i.ctx, -1, 0)

	g := libquickjs.XJS_GetGlobalObject(i.tls, i.ctx)
	for k, f := range map[string]any{
		"print": i.jsPrint,
	} {
		libquickjs.XJS_SetPropertyStr(
			i.tls, i.ctx, g,
			lo.Must(libc.CString(k)),
			libquickjs.XJS_NewCFunction2(
				i.tls, i.ctx, fp(f), lo.Must(libc.CString(k)), int32(1), int32(libquickjs.EJS_CFUNC_generic), 0,
			),
		)
	}

	libquickjs.Xjs_init_module_std(i.tls, i.ctx, lo.Must(libc.CString("std")))
	libquickjs.Xjs_init_module_os(i.tls, i.ctx, lo.Must(libc.CString("os")))
}

func NewQuickJS() (*QuickJS, error) {
	tls := libc.NewTLS()
	rt := libquickjs.XJS_NewRuntime(tls)
	if rt == 0 {
		tls.Close()
		return nil, fmt.Errorf("failed to create with empty JavaScript runtime")
	}

	libquickjs.Xjs_std_init_handlers(tls, rt)
	ctx := libquickjs.XJS_NewContext(tls, rt)
	if ctx == 0 {
		libquickjs.XJS_FreeRuntime(tls, rt)
		tls.Close()
		return nil, fmt.Errorf("failed to create with empty JavaScript context")
	}

	libquickjs.XJS_SetMemoryLimit(tls, rt, libquickjs.Tsize_t(32*1024*1024))

	result := &QuickJS{
		tls:    tls,
		rt:     rt,
		ctx:    ctx,
		Stdout: os.Stdout,
		Stderr: os.Stderr,
	}

	result.prepare()

	// TODO
	// libquickjs.XJS_SetInterruptHandler(tls, rt, fp(interruptHandler), result.interrupting)

	return result, nil
}

func (i *QuickJS) Close() {
	libquickjs.XJS_FreeContext(i.tls, i.ctx)
	libquickjs.XJS_FreeRuntime(i.tls, i.rt)
	i.tls.Close()
}

type QJSEvalOptions struct {
	js_eval_type_global       bool
	js_eval_type_module       bool
	js_eval_flag_strict       bool
	js_eval_flag_compile_only bool
	filename                  string
	await                     bool
	load_only                 bool
}

type QJSEvalOption func(*QJSEvalOptions)

func QJSEvalFlagGlobal(global bool) QJSEvalOption {
	return func(flags *QJSEvalOptions) {
		flags.js_eval_type_global = global
	}
}

func QJSEvalFlagModule(module bool) QJSEvalOption {
	return func(flags *QJSEvalOptions) {
		flags.js_eval_type_module = module
	}
}

func QJSEvalFlagStrict(strict bool) QJSEvalOption {
	return func(flags *QJSEvalOptions) {
		flags.js_eval_flag_strict = strict
	}
}

func QJSEvalFlagCompileOnly(compileOnly bool) QJSEvalOption {
	return func(flags *QJSEvalOptions) {
		flags.js_eval_flag_compile_only = compileOnly
	}
}

func QJSEvalFileName(filename string) QJSEvalOption {
	return func(flags *QJSEvalOptions) {
		flags.filename = filename
	}
}

func QJSEvalAwait(await bool) QJSEvalOption {
	return func(flags *QJSEvalOptions) {
		flags.await = await
	}
}

func QJSEvalLoadOnly(loadOnly bool) QJSEvalOption {
	return func(flags *QJSEvalOptions) {
		flags.load_only = loadOnly
	}
}

func (i *QuickJS) Eval(code string, opts ...QJSEvalOption) libquickjs.TJSValue {
	options := QJSEvalOptions{
		js_eval_type_global: true,
		filename:            "<eval>",
		await:               false,
	}
	for _, fn := range opts {
		fn(&options)
	}

	flag := 0
	if options.js_eval_type_global {
		flag |= libquickjs.MJS_EVAL_TYPE_GLOBAL
	}
	if options.js_eval_type_module {
		flag |= libquickjs.MJS_EVAL_TYPE_MODULE
	}
	if options.js_eval_flag_strict {
		flag |= libquickjs.MJS_EVAL_FLAG_STRICT
	}
	if options.js_eval_flag_compile_only {
		flag |= libquickjs.MJS_EVAL_FLAG_COMPILE_ONLY
	}

	codePtr := lo.Must(libc.CString(code))
	defer libc.Xfree(i.tls, codePtr)

	filenamePtr := lo.Must(libc.CString(options.filename))
	defer libc.Xfree(i.tls, filenamePtr)

	if libquickjs.XJS_DetectModule(i.tls, codePtr, libquickjs.Tsize_t(len(code))) != 0 {
		flag |= libquickjs.MJS_EVAL_TYPE_MODULE
	}

	var val libquickjs.TJSValue
	if options.await {
		val = libquickjs.Xjs_std_await(i.tls, i.ctx, libquickjs.XJS_Eval(i.tls, i.ctx, codePtr, libquickjs.Tsize_t(len(code)), filenamePtr, int32(flag)))
	} else {
		val = libquickjs.XJS_Eval(i.tls, i.ctx, codePtr, libquickjs.Tsize_t(len(code)), filenamePtr, int32(flag))
	}

	return val
}

func (i *QuickJS) ProcessEnv(env map[string]string) {
	g := libquickjs.XJS_GetGlobalObject(i.tls, i.ctx)

	process := libquickjs.XJS_NewObject(i.tls, i.ctx)
	processEnv := libquickjs.XJS_NewObject(i.tls, i.ctx)

	for k, v := range env {
		kPtr := lo.Must(libc.CString(k))
		defer libc.Xfree(i.tls, kPtr)
		vPtr := lo.Must(libc.CString(v))
		defer libc.Xfree(i.tls, vPtr)
		libquickjs.XJS_SetPropertyStr(
			i.tls,
			i.ctx,
			processEnv,
			kPtr,
			libquickjs.XJS_NewStringLen(i.tls, i.ctx, vPtr, libquickjs.Tsize_t(len(v))),
		)
	}

	processEnvNamePtr := lo.Must(libc.CString("env"))
	defer libc.Xfree(i.tls, processEnvNamePtr)
	libquickjs.XJS_SetPropertyStr(i.tls, i.ctx, process, processEnvNamePtr, processEnv)

	processNamePtr := lo.Must(libc.CString("process"))
	defer libc.Xfree(i.tls, processNamePtr)
	libquickjs.XJS_SetPropertyStr(i.tls, i.ctx, g, processNamePtr, process)
}

func (i *QuickJS) LoadModule(code string, moduleName string, cache *[]byte, opts ...QJSEvalOption) libquickjs.TJSValue {
	options := QJSEvalOptions{
		load_only: false,
	}
	for _, fn := range opts {
		fn(&options)
	}

	if cache != nil {
		return i.LoadModuleBytecode(*cache, QJSEvalLoadOnly(options.load_only))
	}

	ptr := lo.Must(libc.CString(code))
	defer libc.Xfree(i.tls, ptr)

	if libquickjs.XJS_DetectModule(i.tls, ptr, libquickjs.Tsize_t(len(code))) == 0 {
		msgPtr := lo.Must(libc.CString(fmt.Sprintf("not a module: %s", moduleName)))
		defer libc.Xfree(i.tls, msgPtr)
		return libquickjs.XJS_ThrowSyntaxError(i.tls, i.ctx, msgPtr, 0)
	}

	codeByte, err := i.Compile(code, QJSEvalFlagModule(true), QJSEvalFlagCompileOnly(true), QJSEvalFileName(moduleName))
	if err != nil {
		msgPtr := lo.Must(libc.CString(err.Error()))
		defer libc.Xfree(i.tls, msgPtr)
		return libquickjs.XJS_ThrowInternalError(i.tls, i.ctx, msgPtr, 0)
	}

	return i.LoadModuleBytecode(codeByte, QJSEvalLoadOnly(options.load_only))
}

func (i *QuickJS) LoadModuleBytecode(buf []byte, opts ...QJSEvalOption) libquickjs.TJSValue {
	if len(buf) == 0 {
		msgPtr := lo.Must(libc.CString("empty bytecode"))
		defer libc.Xfree(i.tls, msgPtr)
		return libquickjs.XJS_ThrowSyntaxError(i.tls, i.ctx, msgPtr, 0)
	}

	obj := libquickjs.XJS_ReadObject(
		i.tls,
		i.ctx,
		uintptr(unsafe.Pointer(&buf[0])),
		libquickjs.Tsize_t(len(buf)),
		libc.Int32(libquickjs.MJS_READ_OBJ_BYTECODE),
	)

	if tag(obj) == int32(libquickjs.EJS_TAG_EXCEPTION) {
		return obj
	}

	options := QJSEvalOptions{}
	for _, fn := range opts {
		fn(&options)
	}

	if options.load_only {
		if tag(obj) == int32(libquickjs.EJS_TAG_MODULE) {
			libquickjs.Xjs_module_set_import_meta(i.tls, i.ctx, obj, 0, 0)
		}
		return obj
	} else {
		if tag(obj) == int32(libquickjs.EJS_TAG_MODULE) {
			if libquickjs.XJS_ResolveModule(i.tls, i.ctx, obj) < 0 {
				libquickjs.XFreeValue(i.tls, i.ctx, obj)

				msgPtr := lo.Must(libc.CString("can not resolve this module"))
				defer libc.Xfree(i.tls, msgPtr)
				return libquickjs.XJS_ThrowSyntaxError(i.tls, i.ctx, msgPtr, 0)
			}

			libquickjs.Xjs_module_set_import_meta(i.tls, i.ctx, obj, 0, 0)
			return libquickjs.Xjs_std_await(i.tls, i.ctx, libquickjs.XJS_EvalFunction(i.tls, i.ctx, obj))
		} else {
			return libquickjs.XJS_EvalFunction(i.tls, i.ctx, obj)
		}
	}
}

func (i *QuickJS) Compile(code string, opts ...QJSEvalOption) ([]byte, error) {
	opts = append(opts, QJSEvalFlagCompileOnly(true))
	val := i.Eval(code, opts...)
	defer libquickjs.XFreeValue(i.tls, i.ctx, val)

	kSize := 0
	ptr := libquickjs.XJS_WriteObject(i.tls, i.ctx, uintptr(unsafe.Pointer(&kSize)), val, libc.Int32(libquickjs.MJS_WRITE_OBJ_BYTECODE))
	if ptr == 0 {
		return nil, fmt.Errorf("OOM")
	}

	defer libquickjs.Xjs_free(i.tls, i.ctx, ptr)

	b := make([]byte, kSize)
	if kSize > 0 {
		copy(b, libc.GoBytes(ptr, kSize))
	}

	return b, nil
}

func (i *QuickJS) ExceptionToError() error {
	e := libquickjs.XJS_GetException(i.tls, i.ctx)
	defer libquickjs.XFreeValue(i.tls, i.ctx, e)
	p := libquickjs.XToCString(i.tls, i.ctx, e)
	defer libquickjs.XJS_FreeCString(i.tls, i.ctx, p)
	return fmt.Errorf("%s", libc.GoString(p))
}
