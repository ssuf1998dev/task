package interpreter

import (
	_ "embed"
	"errors"
	"fmt"
	"io"

	"github.com/dop251/goja"
)

var ErrNilOptions = errors.New("interpreter: nil options given")

type InterpretJSOptions struct {
	Script string
	Dir    string
	Env    map[string]string
	Stdin  io.Reader
	Stdout io.Writer
	Stderr io.Writer
}

// func chdirScript(dir string) string {
// 	if len(dir) <= 0 {
// 		return ""
// 	}
// 	return fmt.Sprintf("(await import('os')).chdir('%s');", escape(dir))
// }

func exposeProcess(rt *goja.Runtime, env map[string]string) {
	type process struct {
		Env map[string]string `json:"env"`
	}

	p := process{Env: env}

	rt.Set("process", &p)
}

func InterpretJS(opts *InterpretJSOptions) error {
	if opts == nil {
		return ErrNilOptions
	}

	rt := goja.New()
	rt.SetFieldNameMapper(goja.TagFieldNameMapper("json", true))

	exposeProcess(rt, opts.Env)

	if v, err := rt.RunString(
		fmt.Sprintf("(async()=>{%s})().then(JSON.stringify)", opts.Script),
	); err == nil {
		p, _ := v.Export().(*goja.Promise)

		switch p.State() {
		case goja.PromiseStateFulfilled:
			if result, ok := p.Result().Export().(string); ok {
				opts.Stdout.Write([]byte(result + "\n"))
			} else {
				opts.Stdout.Write([]byte("undefined\n"))
			}
		default:
			opts.Stderr.Write([]byte(p.Result().String() + "\n"))
			return err
		}
	} else {
		opts.Stderr.Write([]byte(err.Error() + "\n"))
		return err
	}

	return nil
}
