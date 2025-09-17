package ast

import (
	"fmt"

	"gopkg.in/yaml.v3"

	"github.com/go-task/task/v3/errors"
)

type If struct {
	Value string
	Sh    *string
}

func (v *If) UnmarshalYAML(node *yaml.Node) error {
	switch node.Kind {
	case yaml.MappingNode:
		key := "<none>"
		if len(node.Content) > 0 {
			key = node.Content[0].Value
		}
		switch key {
		case "sh":
			var m struct {
				Sh *string
			}
			if err := node.Decode(&m); err != nil {
				return errors.NewTaskfileDecodeError(err, node)
			}
			v.Sh = m.Sh
			return nil
		default:
			return errors.NewTaskfileDecodeError(nil, node).WithMessage(`%q is not a valid variable type. Try "sh" using a scalar value`, key)
		}
	default:
		var raw any
		if err := node.Decode(&raw); err != nil {
			return errors.NewTaskfileDecodeError(err, node)
		}
		switch value := raw.(type) {
		case string:
			v.Value = value
			return nil
		case bool:
			v.Value = fmt.Sprintf("%t", value)
			return nil
		default:
			return errors.NewTaskfileDecodeError(nil, node).WithMessage("invalid scalar value type")
		}
	}
}

func (v *If) DeepCopy() *If {
	if v == nil {
		return nil
	}
	return &If{
		Value: v.Value,
		Sh:    v.Sh,
	}
}
