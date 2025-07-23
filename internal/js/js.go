package js

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

var ErrNilOptions = errors.New("js: nil options given")

type JSOptions struct {
	Script  string
	Dialect string
	Dir     string
	Env     map[string]string
	Stdin   io.Reader
	Stdout  io.Writer
	Stderr  io.Writer
}

type JavaScript struct {
	qjs         *QuickJS
	civetLoaded bool
}

func (j *JavaScript) escape(s string) string {
	return string(regexp.MustCompile("'").ReplaceAll([]byte(s), []byte("\\'")))
}

func (j *JavaScript) chdirScript(dir string) string {
	if len(dir) <= 0 {
		return ""
	}
	return fmt.Sprintf("(await import('os')).chdir('%s');", j.escape(dir))
}

func NewJavaScript() *JavaScript {
	qjs, _ := NewQuickJS()
	return &JavaScript{qjs: qjs}
}

func (j *JavaScript) Interpret(opts *JSOptions) error {
	if opts == nil {
		return ErrNilOptions
	}

	if dir, err := os.Getwd(); err == nil {
		defer (func() {
			_ = os.Chdir(dir)
		})()
	}

	j.qjs.ProcessEnv(opts.Env)

	script := opts.Script

	switch opts.Dialect {
	case "civet":
		if !j.civetLoaded {
			mod := j.qjs.LoadModule(fmt.Sprintf("export{compile};\n%s", civetJs), "civet")
			if tag(mod) == libquickjs.EJS_TAG_EXCEPTION {
				err := j.qjs.ExceptionToError()
				_, _ = opts.Stderr.Write([]byte(err.Error() + "\n"))
				return err
			}
			defer libquickjs.XFreeValue(j.qjs.tls, j.qjs.ctx, mod)
			j.civetLoaded = true
		}

		code := j.escape(script)

		js := j.qjs.Eval(fmt.Sprintf(
			"(async()=>(await import('civet')).compile(`%s`,{js:true}))()",
			code,
		), QJSEvalAwait(true))
		defer libquickjs.XFreeValue(j.qjs.tls, j.qjs.ctx, js)
		if tag(js) == libquickjs.EJS_TAG_EXCEPTION {
			err := j.qjs.ExceptionToError()
			_, _ = opts.Stderr.Write([]byte(err.Error() + "\n"))
			return err
		}

		jsPtr := libquickjs.XToCString(j.qjs.tls, j.qjs.ctx, js)
		defer libquickjs.XJS_FreeCString(j.qjs.tls, j.qjs.ctx, jsPtr)
		script = libc.GoString(jsPtr)
	default:
	}

	result := j.qjs.Eval(
		fmt.Sprintf("(async()=>{%s%s})()", j.chdirScript(opts.Dir), script),
		QJSEvalAwait(true),
	)
	json := libquickjs.XJS_JSONStringify(j.qjs.tls, j.qjs.ctx, result, JS_UNDEFINED, JS_UNDEFINED)
	defer libquickjs.XFreeValue(j.qjs.tls, j.qjs.ctx, json)
	if tag(json) == libquickjs.EJS_TAG_EXCEPTION {
		err := j.qjs.ExceptionToError()
		_, _ = opts.Stderr.Write([]byte(err.Error() + "\n"))
		return err
	}

	jsonPtr := libquickjs.XToCString(j.qjs.tls, j.qjs.ctx, json)
	defer libquickjs.XJS_FreeCString(j.qjs.tls, j.qjs.ctx, jsonPtr)
	_, _ = opts.Stdout.Write([]byte(libc.GoString(jsonPtr) + "\n"))

	return nil
}

func (j *JavaScript) Close() {
	if j.qjs != nil {
		j.qjs.Close()
	}
}
