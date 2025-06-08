package ast_test

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"gopkg.in/yaml.v3"

	"github.com/go-task/task/v3/taskfile/ast"
)

func TestPluginsParse(t *testing.T) {
	t.Parallel()

	tests := []struct {
		content  string
		v        any
		expected any
	}{
		{
			`
a: a.wasm
b:
  file: b.wasm
`,
			&ast.Plugins{},
			ast.NewPlugins(
				&ast.PluginElement{Key: "a", Value: &ast.Plugin{File: "a.wasm"}},
				&ast.PluginElement{Key: "b", Value: &ast.Plugin{File: "b.wasm"}},
			),
		},
		{
			`
- a.wasm
- b.wasm
`,
			&ast.Plugins{},
			ast.NewPlugins(
				&ast.PluginElement{Key: "a", Value: &ast.Plugin{File: "a.wasm"}},
				&ast.PluginElement{Key: "b", Value: &ast.Plugin{File: "b.wasm"}},
			),
		},
	}
	for _, test := range tests {
		err := yaml.Unmarshal([]byte(test.content), test.v)
		require.NoError(t, err)
		assert.Equal(t, test.expected, test.v)
	}
}
