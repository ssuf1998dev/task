package task

import (
	"bytes"
	"context"
	"strings"

	"github.com/go-task/task/v3/internal/env"
	"github.com/go-task/task/v3/internal/execext"
	"github.com/go-task/task/v3/taskfile/ast"
)

func (e *Executor) isIf(ctx context.Context, t *ast.Task, value *ast.If) (bool, error) {
	if value == nil {
		return true, nil
	}

	if value.Sh != nil {
		var buff bytes.Buffer
		err := execext.RunCommand(ctx, &execext.RunCommandOptions{
			Command: *value.Sh,
			Dir:     t.Dir,
			Env:     env.Get(t),
			Stdout:  &buff,
		})
		if err != nil {
			return false, err
		}
		return strings.TrimSpace(buff.String()) == "true", nil
	}

	return value.Value == "true", nil
}
