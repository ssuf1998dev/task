package js

import (
	_ "embed"
	"fmt"
	"io"
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
	Env     map[string]string
	Stdout  io.Writer
	Stderr  io.Writer
}

type JavaScript struct {
	qjs *QuickJS
}

func (j *JavaScript) escape(s string) string {
	return string(regexp.MustCompile("'").ReplaceAll([]byte(s), []byte("\\'")))
}

func NewJavaScript() (*JavaScript, error) {
	qjs, err := NewQuickJS()
	if err != nil {
		return nil, err
	}
	return &JavaScript{qjs: qjs}, nil
}

func NewJavaScriptInterpret(opts *JSOptions) error {
	j, err := NewJavaScript()
	if err != nil {
		return err
	}
	defer j.Close()
	return j.Interpret(opts)
}

func (j *JavaScript) Interpret(opts *JSOptions) error {
	if opts == nil {
		return ErrNilOptions
	}

	if opts.Stdout != nil {
		j.qjs.Stdout = opts.Stdout
	}
	if opts.Stderr != nil {
		j.qjs.Stderr = opts.Stderr
	}

	j.qjs.ProcessEnv(opts.Env)

	script := opts.Script

	switch opts.Dialect {
	case "civet":
		mod := j.qjs.LoadModule(fmt.Sprintf("export{compile};\n%s", civetJs), "civet")
		if tag(mod) == libquickjs.EJS_TAG_EXCEPTION {
			err := j.qjs.ExceptionToError()
			return err
		}
		defer libquickjs.XFreeValue(j.qjs.tls, j.qjs.ctx, mod)

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

	result := j.qjs.Eval(script, QJSEvalAwait(true))
	if tag(result) == libquickjs.EJS_TAG_EXCEPTION {
		err := j.qjs.ExceptionToError()
		_, _ = opts.Stderr.Write([]byte(err.Error() + "\n"))
		return err
	}

	return nil
}

func (j *JavaScript) Close() {
	if j.qjs != nil {
		j.qjs.Close()
	}
}
