package ast_test

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"gopkg.in/yaml.v3"

	"github.com/go-task/task/v3/taskfile/ast"
)

func TestSshParse(t *testing.T) {
	t.Parallel()

	tests := []struct {
		content  string
		v        any
		expected any
	}{
		{
			"//127.0.0.1:22",
			&ast.Ssh{},
			&ast.Ssh{
				Addr:       "127.0.0.1:22",
				User:       "",
				Password:   "",
				PrivateKey: "",
				Insecure:   false,
			},
		},
		{
			"//foo:bar@127.0.0.1:22",
			&ast.Ssh{},
			&ast.Ssh{
				Addr:       "127.0.0.1:22",
				User:       "foo",
				Password:   "bar",
				PrivateKey: "",
				Insecure:   false,
			},
		},
		{
			"//foo@127.0.0.1:22",
			&ast.Ssh{},
			&ast.Ssh{
				Addr:       "127.0.0.1:22",
				User:       "foo",
				Password:   "",
				PrivateKey: "",
				Insecure:   false,
			},
		},
		{
			`
addr: 127.0.0.1:22
user: foo
password: bar
`,
			&ast.Ssh{},
			&ast.Ssh{
				Addr:       "127.0.0.1:22",
				User:       "foo",
				Password:   "bar",
				PrivateKey: "",
				Insecure:   false,
			},
		},
		{
			"addr: 127.0.0.1:22",
			&ast.Ssh{},
			&ast.Ssh{
				Addr:       "127.0.0.1:22",
				User:       "",
				Password:   "",
				PrivateKey: "",
				Insecure:   false,
			},
		},
	}
	for _, test := range tests {
		err := yaml.Unmarshal([]byte(test.content), test.v)
		require.NoError(t, err)
		assert.Equal(t, test.expected, test.v)
	}
}
