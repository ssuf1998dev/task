package js

import (
	"os"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestBasic(t *testing.T) {
	t.Parallel()

	Setup()
	js1, err := NewJavaScript()
	require.NoError(t, err)
	defer js1.Close()

	out, err := js1.Eval(&JSEvalOptions{
		Script: "print(\"hello, javascript\")",
	})
	require.NoError(t, err)
	assert.Contains(t, out, "hello, javascript")

	js2, err := NewJavaScript()
	require.NoError(t, err)
	defer js2.Close()

	out, err = js2.Eval(&JSEvalOptions{
		Script:  "\"hello, civet\" |> print",
		Dialect: "civet",
	})
	require.NoError(t, err)
	assert.Contains(t, out, "hello, civet")
}

func TestCwd(t *testing.T) {
	t.Parallel()

	Setup()
	js, err := NewJavaScript()
	require.NoError(t, err)
	defer js.Close()

	cwd, _ := os.Getwd()

	out, err := js.Eval(&JSEvalOptions{
		Script: `
		import * as os from "qjs:os";
		print(os.getcwd()[0]);
		os.chdir("~");
		`,
	})
	require.NoError(t, err)
	assert.Equal(t, cwd, strings.TrimSuffix(out, "\n"))

	out, err = js.Eval(&JSEvalOptions{
		Script: `
		import * as os from "qjs:os";
		print(os.getcwd()[0]);
		`,
	})
	require.NoError(t, err)
	assert.Equal(t, cwd, strings.TrimSuffix(out, "\n"))
}
