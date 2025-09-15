package ast

import (
	"fmt"
	"strings"

	"gopkg.in/yaml.v3"

	"github.com/go-task/task/v3/errors"
	"github.com/go-task/task/v3/internal/deepcopy"
)

type SshUpload struct {
	Source string
	Target string
}

func (u *SshUpload) DeepCopy() *SshUpload {
	if u == nil {
		return nil
	}
	return &SshUpload{Source: u.Source, Target: u.Target}
}

func (u *SshUpload) UnmarshalYAML(node *yaml.Node) error {
	switch node.Kind {
	case yaml.ScalarNode:
		var item string
		err := node.Decode(&item)
		if err != nil {
			return errors.NewTaskfileDecodeError(err, node)
		}
		parts := strings.SplitN(item, ":", 2)
		if len(parts) < 2 {
			return errors.NewTaskfileDecodeError(fmt.Errorf("must be <source>:<target>"), node)
		}
		u.Source = parts[0]
		u.Target = parts[1]
		return nil
	case yaml.MappingNode:
		var item struct {
			Source string
			Target string
		}
		if err := node.Decode(&item); err != nil {
			return errors.NewTaskfileDecodeError(err, node)
		}
		*u = item
		return nil
	}
	return errors.NewTaskfileDecodeError(nil, node).WithTypeMessage("upload")
}

type Ssh struct {
	Url        string
	Addr       string
	User       string
	Password   string
	Key        string
	KeyPath    string
	KnownHosts []string
	Timeout    int
	Insecure   bool
	Uploads    []SshUpload
	Disabled   bool
}

func (s *Ssh) DeepCopy() *Ssh {
	if s == nil {
		return nil
	}
	return &Ssh{
		Url:        s.Url,
		Addr:       s.Addr,
		User:       s.User,
		Password:   s.Password,
		Key:        s.Key,
		KeyPath:    s.KeyPath,
		KnownHosts: deepcopy.Slice(s.KnownHosts),
		Timeout:    s.Timeout,
		Insecure:   s.Insecure,
		Uploads:    deepcopy.Slice(s.Uploads),
		Disabled:   s.Disabled,
	}
}

func (s *Ssh) UnmarshalYAML(node *yaml.Node) error {
	switch node.Kind {
	case yaml.ScalarNode:
		var v any
		err := node.Decode(&v)
		if err != nil {
			return errors.NewTaskfileDecodeError(err, node)
		}
		if url, ok := v.(string); ok {
			s.Url = url
			return nil
		}
		if enable, ok := v.(bool); ok {
			s.Disabled = !enable
			return nil
		}
		return errors.NewTaskfileDecodeError(err, node)
	case yaml.MappingNode:
		var ssh struct {
			Url        string
			Addr       string
			User       string
			Password   string
			Key        string
			KeyPath    string
			KnownHosts []string
			Timeout    int
			Insecure   bool
			Uploads    []SshUpload
			Disabled   bool
		}
		if err := node.Decode(&ssh); err != nil {
			return errors.NewTaskfileDecodeError(err, node)
		}
		*s = ssh
		return nil
	}
	return errors.NewTaskfileDecodeError(nil, node).WithTypeMessage("ssh")
}
