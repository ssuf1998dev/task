package js

import (
	"bytes"
	"context"
	"crypto/rand"
	_ "embed"
	"encoding/json"
	"fmt"
	"io"
	"os"

	extism "github.com/extism/go-sdk"
	"github.com/tetratelabs/wazero"
)

//go:embed qjs.wasm
var qjswasm []byte

type JavaScript struct {
	plugin *extism.Plugin
	stdin  *bytes.Buffer
	stdout *bytes.Buffer
	stderr *bytes.Buffer
}

type JSEvalOptions struct {
	Script  string
	Dialect string
	Dir     string
	Env     map[string]string
	Stdin   io.Reader
	Stdout  io.Writer
	Stderr  io.Writer
}

var (
	ctx            context.Context
	cache          wazero.CompilationCache
	compiledPlugin *extism.CompiledPlugin
)

func init() {
	ctx = context.Background()

	cache = wazero.NewCompilationCache()

	mft := extism.Manifest{
		Wasm:   []extism.Wasm{extism.WasmData{Data: qjswasm}},
		Config: map[string]string{},
	}
	config := extism.PluginConfig{
		EnableWasi:    true,
		RuntimeConfig: wazero.NewRuntimeConfig().WithCompilationCache(cache),
	}
	compiledPlugin, _ = extism.NewCompiledPlugin(ctx, mft, config, []extism.HostFunction{})
}

func NewJavaScript() (*JavaScript, error) {
	if compiledPlugin == nil {
		return nil, fmt.Errorf("js: init failed")
	}
	var stdin bytes.Buffer
	var stdout bytes.Buffer
	var stderr bytes.Buffer
	plugin, err := compiledPlugin.Instance(ctx, extism.PluginInstanceConfig{
		ModuleConfig: wazero.NewModuleConfig().
			WithRandSource(rand.Reader).
			WithFSConfig(wazero.NewFSConfig().WithDirMount("/", "/")).
			WithSysNanosleep().
			WithSysNanotime().
			WithSysWalltime().
			WithStdin(&stdin).
			WithStdout(&stdout).
			WithStderr(&stderr),
	})
	if err != nil {
		return nil, err
	}
	_, _, err = plugin.Call("warmup", nil)
	if err != nil {
		plugin.Close(ctx)
		return nil, err
	}

	return &JavaScript{
		plugin: plugin,
		stdin:  &stdin,
		stdout: &stdout,
		stderr: &stderr,
	}, nil
}

func (js *JavaScript) Close() {
	js.stdout.Reset()
	js.stderr.Reset()
	js.plugin.Close(ctx)
}

func (js *JavaScript) Eval(options *JSEvalOptions) (string, error) {
	if options == nil {
		return "", fmt.Errorf("js: nil options given")
	}

	if js.stdin != nil {
		js.stdin.Reset()
	}
	if js.stdout != nil {
		js.stdout.Reset()
	}
	if js.stderr != nil {
		js.stderr.Reset()
	}

	if options.Env != nil {
		if envJson, err := json.Marshal(options.Env); err == nil {
			_, _, _ = js.plugin.Call("setEnv", []byte(envJson))
		}
	}

	dir, _ := os.Getwd()
	if len(options.Dir) != 0 {
		dir = options.Dir
	}
	js.plugin.Config["eval.dir"] = dir

	js.plugin.Config["eval.dialect"] = options.Dialect

	if options.Stdin != nil {
		_, _ = options.Stdin.Read(js.stdin.Bytes())
	}

	exit, _, err := js.plugin.Call("eval", []byte(options.Script))
	if err != nil {
		return "", err
	}
	if exit > 0 {
		return "", fmt.Errorf("js: unknown error, exit with code %d", exit)
	}
	if options.Stdout != nil {
		_, _ = options.Stdout.Write(js.stdout.Bytes())
	}
	if options.Stderr != nil {
		_, _ = options.Stderr.Write(js.stderr.Bytes())
	}
	return js.stdout.String(), nil
}
