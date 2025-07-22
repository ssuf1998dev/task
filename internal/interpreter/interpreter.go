package interpreter

import (
	_ "embed"
	"fmt"
	"io"
	"os"
	"regexp"

	"modernc.org/libc"
	"modernc.org/libquickjs"

	"github.com/go-task/task/v3/errors"
)

//go:embed civet/Civet/dist/quickjs.min.mjs
var civetJs string

var ErrNilOptions = errors.New("interpreter: nil options given")

type InterpretJSOptions struct {
	Script  string
	Dialect string
	Dir     string
	Env     map[string]string
	Stdin   io.Reader
	Stdout  io.Writer
	Stderr  io.Writer
}

type JSInterpreter struct {
	qjs         *QuickJSInterpreter
	civetLoaded bool
}

func (i *JSInterpreter) escape(s string) string {
	return string(regexp.MustCompile("'").ReplaceAll([]byte(s), []byte("\\'")))
}

func (i *JSInterpreter) chdirScript(dir string) string {
	if len(dir) <= 0 {
		return ""
	}
	return fmt.Sprintf("(await import('os')).chdir('%s');", i.escape(dir))
}

func NewJSInterpreter() *JSInterpreter {
	qjs, _ := NewQuickJSInterpreter()
	return &JSInterpreter{qjs: qjs}
}

func (i *JSInterpreter) Interpret(opts *InterpretJSOptions) error {
	if opts == nil {
		return ErrNilOptions
	}

	if dir, err := os.Getwd(); err == nil {
		defer (func() {
			_ = os.Chdir(dir)
		})()
	}

	i.qjs.ProcessEnv(opts.Env)

	script := opts.Script

	switch opts.Dialect {
	case "civet":
		if !i.civetLoaded {
			mod := i.qjs.LoadModule(fmt.Sprintf("export{compile};\n%s", civetJs), "civet")
			if tag(mod) == libquickjs.EJS_TAG_EXCEPTION {
				err := i.qjs.ExceptionToError()
				_, _ = opts.Stderr.Write([]byte(err.Error() + "\n"))
				return err
			}
			defer libquickjs.XFreeValue(i.qjs.tls, i.qjs.ctx, mod)
			i.civetLoaded = true
		}

		code := i.escape(script)

		js := i.qjs.Eval(fmt.Sprintf(
			"(async()=>(await import('civet')).compile(`%s`,{js:true}))()",
			code,
		), QJSEvalAwait(true))
		defer libquickjs.XFreeValue(i.qjs.tls, i.qjs.ctx, js)
		if tag(js) == libquickjs.EJS_TAG_EXCEPTION {
			err := i.qjs.ExceptionToError()
			_, _ = opts.Stderr.Write([]byte(err.Error() + "\n"))
			return err
		}

		jsPtr := libquickjs.XToCString(i.qjs.tls, i.qjs.ctx, js)
		defer libquickjs.XJS_FreeCString(i.qjs.tls, i.qjs.ctx, jsPtr)
		script = libc.GoString(jsPtr)
	default:
	}

	result := i.qjs.Eval(
		fmt.Sprintf("(async()=>{%s%s})()", i.chdirScript(opts.Dir), script),
		QJSEvalAwait(true),
	)
	json := libquickjs.XJS_JSONStringify(i.qjs.tls, i.qjs.ctx, result, JS_UNDEFINED, JS_UNDEFINED)
	defer libquickjs.XFreeValue(i.qjs.tls, i.qjs.ctx, json)
	if tag(json) == libquickjs.EJS_TAG_EXCEPTION {
		err := i.qjs.ExceptionToError()
		_, _ = opts.Stderr.Write([]byte(err.Error() + "\n"))
		return err
	}

	jsonPtr := libquickjs.XToCString(i.qjs.tls, i.qjs.ctx, json)
	defer libquickjs.XJS_FreeCString(i.qjs.tls, i.qjs.ctx, jsonPtr)
	_, _ = opts.Stdout.Write([]byte(libc.GoString(jsonPtr) + "\n"))

	return nil
}

func (i *JSInterpreter) Close() {
	if i.qjs != nil {
		i.qjs.Close()
	}
}
