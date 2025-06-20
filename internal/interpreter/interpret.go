package interpreter

import (
	_ "embed"
	"errors"
	"fmt"
	"io"
	"regexp"

	"github.com/buke/quickjs-go"
)

//go:embed civet/Civet/dist/quickjs.mjs
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

func escape(s string) string {
	return string(regexp.MustCompile("'").ReplaceAll([]byte(s), []byte("\\'")))
}

func chdirScript(dir string) string {
	if len(dir) <= 0 {
		return ""
	}
	return fmt.Sprintf("(await import('os')).chdir('%s');", escape(dir))
}

func exposeProcess(ctx *quickjs.Context, env map[string]string) {
	proc := ctx.Object()

	procEnv := ctx.Object()
	proc.Set("env", procEnv)
	for k, v := range env {
		procEnv.Set(k, ctx.String(v))
	}

	ctx.Globals().Set("process", proc)
}

func InterpretJS(opts *InterpretJSOptions) error {
	if opts == nil {
		return ErrNilOptions
	}

	rt := quickjs.NewRuntime()
	defer rt.Close()

	ctx := rt.NewContext()
	defer ctx.Close()

	exposeProcess(ctx, opts.Env)

	var script = opts.Script

	switch opts.Dialect {
	case "civet":
		mod, err := ctx.LoadModule(fmt.Sprintf("export{compile};\n%s", civetJs), "civet")
		defer mod.Free()
		if err != nil {
			return err
		}

		code := escape(script)

		js, err := ctx.Eval(fmt.Sprintf(
			"(async()=>(await import('civet')).compile(`%s`,{js:true}))()",
			code,
		), quickjs.EvalAwait(true))
		defer js.Free()
		if err != nil {
			return err
		}

		script = js.ToString()
	default:
	}

	if result, err := ctx.Eval(
		fmt.Sprintf("(async()=>{%s%s})()", chdirScript(opts.Dir), script),
		quickjs.EvalAwait(true),
	); err == nil {
		defer result.Free()
		opts.Stdout.Write([]byte(result.JSONStringify() + "\n"))
	} else {
		opts.Stderr.Write([]byte(err.Error() + "\n"))
		return err
	}

	return nil
}
