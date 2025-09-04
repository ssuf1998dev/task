package ast

import (
	"gopkg.in/yaml.v3"

	"github.com/go-task/task/v3/errors"
	"github.com/go-task/task/v3/internal/deepcopy"
)

type Ssh struct {
	Addr       string
	User       string
	Password   string
	PrivateKey string
	KnownHosts []string
	Insecure   bool
	Raw        string
}

func (s *Ssh) DeepCopy() *Ssh {
	if s == nil {
		return nil
	}
	return &Ssh{
		Addr:       s.Addr,
		User:       s.User,
		Password:   s.Password,
		PrivateKey: s.PrivateKey,
		KnownHosts: deepcopy.Slice(s.KnownHosts),
		Insecure:   s.Insecure,
		Raw:        s.Raw,
	}
}

func (s *Ssh) UnmarshalYAML(node *yaml.Node) error {
	switch node.Kind {
	case yaml.ScalarNode:
		var ssh string
		err := node.Decode(&ssh)
		if err != nil {
			return errors.NewTaskfileDecodeError(err, node)
		}
		s.Raw = ssh
		return nil
	case yaml.MappingNode:
		var ssh struct {
			Addr       string
			User       string
			Password   string
			PrivateKey string
			KnownHosts []string
			Insecure   bool
			Raw        string
		}
		if err := node.Decode(&ssh); err != nil {
			return errors.NewTaskfileDecodeError(err, node)
		}
		*s = ssh
		s.Raw = ""
		return nil
	}
	return errors.NewTaskfileDecodeError(nil, node).WithTypeMessage("ssh")
}
