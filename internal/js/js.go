package js

import (
	"bytes"
	"context"
	"crypto/rand"
	_ "embed"
	"fmt"
	"os"
	"sync"

	extism "github.com/extism/go-sdk"
	"github.com/tetratelabs/wazero"
)

const (
	DIALECT_NONE  = 0
	DIALECT_CIVET = 1
)

//go:embed qjs.wasm
var qjswasm []byte

type JavaScriptWorker struct {
	ctx    context.Context
	plugin *extism.Plugin
	stdout *bytes.Buffer
	stderr *bytes.Buffer
}

func (w *JavaScriptWorker) Eval(input string, dialect int) (string, string, error) {
	cwd, _ := os.Getwd()
	w.plugin.Call("eval", fmt.Appendf(nil, "import * as os from 'qjs:os';os.chdir('%s');", cwd))

	switch dialect {
	case DIALECT_CIVET:
		w.plugin.Config["eval.dialect"] = "civet"
	default:
		w.plugin.Config["eval.dialect"] = "javascript"
	}
	exit, out, err := w.plugin.Call("eval", []byte(input))
	if err != nil {
		return "", "", err
	}
	if exit > 0 {
		return "", "", fmt.Errorf("exit with code %d", exit)
	}
	return string(out), w.stdout.String(), nil
}

type JavaScriptPool struct {
	ctx context.Context
	sync.Pool
	plugin *extism.CompiledPlugin
	cache  *wazero.CompilationCache
}

func NewJavaScriptPool() (*JavaScriptPool, error) {
	ctx := context.Background()

	cache := wazero.NewCompilationCache()

	mft := extism.Manifest{
		Wasm:   []extism.Wasm{extism.WasmData{Data: qjswasm}},
		Config: map[string]string{},
	}
	config := extism.PluginConfig{
		EnableWasi:    true,
		RuntimeConfig: wazero.NewRuntimeConfig().WithCompilationCache(cache),
	}
	plugin, err := extism.NewCompiledPlugin(ctx, mft, config, []extism.HostFunction{})
	if err != nil {
		return nil, err
	}

	p := &JavaScriptPool{
		ctx: ctx,
		Pool: sync.Pool{
			New: func() any {
				var stdout bytes.Buffer
				var stderr bytes.Buffer
				inst, err := plugin.Instance(ctx, extism.PluginInstanceConfig{
					ModuleConfig: wazero.NewModuleConfig().
						WithRandSource(rand.Reader).
						WithFSConfig(wazero.NewFSConfig().WithDirMount("/", "/")).
						WithSysNanosleep().
						WithSysNanotime().
						WithSysWalltime().
						WithStdout(&stdout).
						WithStderr(&stderr),
				})
				if err != nil {
					return err
				}
				_, _, err = inst.Call("warmup", nil)
				if err != nil {
					inst.Close(ctx)
					return err
				}
				return &JavaScriptWorker{
					ctx:    ctx,
					plugin: inst,
					stdout: &stdout,
					stderr: &stderr,
				}
			},
		},
		plugin: plugin,
		cache:  &cache,
	}

	return p, nil
}

func (p *JavaScriptPool) Close() {
	(*p.cache).Close(p.ctx)
	p.plugin.Close(p.ctx)
}
