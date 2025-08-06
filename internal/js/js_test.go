package js

import (
	"os"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestBasic(t *testing.T) {
	p, err := NewJavaScriptPool()
	if err != nil {
		t.Error(err)
	}
	defer p.Close()

	w1 := p.Get().(*JavaScriptWorker)
	_, out, err := w1.Eval("print(\"hello, javascript\")", DIALECT_NONE)
	if err != nil {
		t.Error(err)
	}
	assert.Contains(t, out, "hello, javascript")
	p.Put(w1)

	w2 := p.Get().(*JavaScriptWorker)
	_, out, err = w2.Eval("\"hello, civet\" |> print", DIALECT_CIVET)
	if err != nil {
		t.Error(err)
	}
	assert.Contains(t, out, "hello, civet")
}

func TestCwd(t *testing.T) {
	cwd, _ := os.Getwd()

	p, err := NewJavaScriptPool()
	if err != nil {
		t.Error(err)
	}
	defer p.Close()

	w := p.Get().(*JavaScriptWorker)
	_, out, err := w.Eval(`
	import * as os from "qjs:os";
	print(os.getcwd()[0]);
	`, DIALECT_NONE)
	if err != nil {
		t.Error(err)
	}
	assert.Equal(t, cwd, strings.TrimSuffix(out, "\n"))
	p.Put(w)
}
