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
	"path/filepath"
	"slices"
	"sync"

	extism "github.com/extism/go-sdk"
	"github.com/tetratelabs/wazero"
)

//go:embed qjs.wasm
var qjswasm []byte

var (
	ctx            context.Context
	cache          wazero.CompilationCache
	compiledPlugin *extism.CompiledPlugin
	once           sync.Once
)

func setup() {
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

func Setup() {
	once.Do(setup)
}

type JavaScript struct {
	plugin *extism.Plugin
	stdin  *bytes.Buffer
	stdout *bytes.Buffer
	stderr *bytes.Buffer
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

type JSEvalOptions struct {
	Script  string
	Dialect string
	Dir     string
	Env     map[string]string
	Stdin   io.Reader
	Stdout  io.Writer
	Stderr  io.Writer
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
	js.plugin.Config = map[string]string{}
	return js.stdout.String(), nil
}

type JSEvalFileOptions struct {
	File    string
	Dialect string
	Dir     string
	Env     map[string]string
	Args    []string
	Stdin   io.Reader
	Stdout  io.Writer
	Stderr  io.Writer
}

func (js *JavaScript) EvalFile(options *JSEvalFileOptions) (string, error) {
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
	js.plugin.Config["evalFile.dir"] = dir

	js.plugin.Config["evalFile.argv0"] = filepath.ToSlash(os.Args[0])

	options.Args = slices.Insert(options.Args, 0, options.File)
	json, err := json.Marshal(options.Args)
	if err == nil {
		js.plugin.Config["evalFile.scriptArgs"] = string(json)
	}

	js.plugin.Config["evalFile.dialect"] = options.Dialect

	if options.Stdin != nil {
		_, _ = options.Stdin.Read(js.stdin.Bytes())
	}

	exit, _, err := js.plugin.Call("evalFile", []byte(options.File))
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
	js.plugin.Config = map[string]string{}
	return js.stdout.String(), nil
}
