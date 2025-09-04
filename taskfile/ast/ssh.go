package ast

import (
	"net/url"

	"gopkg.in/yaml.v3"

	"github.com/go-task/task/v3/errors"
)

type Ssh struct {
	Addr       string
	User       string
	Password   string
	PrivateKey string
	Insecure   bool
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
		Insecure:   s.Insecure,
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
		parsed, err := url.Parse(ssh)
		if err != nil {
			return errors.NewTaskfileDecodeError(err, node)
		}
		s.Addr = parsed.Host
		s.User = parsed.User.Username()
		s.Password, _ = parsed.User.Password()
		s.Insecure = parsed.Query().Has("insecure")
		return nil
	case yaml.MappingNode:
		var ssh struct {
			Addr       string
			User       string
			Password   string
			PrivateKey string
			Insecure   bool
		}
		if err := node.Decode(&ssh); err != nil {
			return errors.NewTaskfileDecodeError(err, node)
		}
		*s = ssh
		return nil
	}
	return errors.NewTaskfileDecodeError(nil, node).WithTypeMessage("ssh")
}
