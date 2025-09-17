package ast

import (
	"gopkg.in/yaml.v3"

	"github.com/go-task/task/v3/errors"
	"github.com/go-task/task/v3/internal/deepcopy"
)

// Cmd is a task command
type Cmd struct {
	Cmd         string
	Task        string
	ThisSsh     bool   `yaml:"this_ssh"`
	StdoutFile  string `yaml:"stdout_file"`
	If          *If
	For         *For
	Silent      bool
	Set         []string
	Shopt       []string
	Vars        *Vars
	IgnoreError bool
	Defer       bool
	Platforms   []*Platform
	Interp      string
}

func (c *Cmd) DeepCopy() *Cmd {
	if c == nil {
		return nil
	}
	return &Cmd{
		Cmd:         c.Cmd,
		Task:        c.Task,
		ThisSsh:     c.ThisSsh,
		StdoutFile:  c.StdoutFile,
		If:          c.If.DeepCopy(),
		For:         c.For.DeepCopy(),
		Silent:      c.Silent,
		Set:         deepcopy.Slice(c.Set),
		Shopt:       deepcopy.Slice(c.Shopt),
		Vars:        c.Vars.DeepCopy(),
		IgnoreError: c.IgnoreError,
		Defer:       c.Defer,
		Platforms:   deepcopy.Slice(c.Platforms),
		Interp:      c.Interp,
	}
}

func (c *Cmd) UnmarshalYAML(node *yaml.Node) error {
	switch node.Kind {

	case yaml.ScalarNode:
		var cmd string
		if err := node.Decode(&cmd); err != nil {
			return errors.NewTaskfileDecodeError(err, node)
		}
		c.Cmd = cmd
		c.ThisSsh = true
		return nil

	case yaml.MappingNode:
		var cmdStruct struct {
			Cmd         string
			Task        string
			ThisSsh     *bool  `yaml:"this_ssh"`
			StdoutFile  string `yaml:"stdout_file"`
			If          *If
			For         *For
			Silent      bool
			Set         []string
			Shopt       []string
			Vars        *Vars
			IgnoreError bool `yaml:"ignore_error"`
			Defer       *Defer
			Platforms   []*Platform
			Ssh         *Ssh
			Interp      string
		}
		if err := node.Decode(&cmdStruct); err != nil {
			return errors.NewTaskfileDecodeError(err, node)
		}
		if cmdStruct.Defer != nil {

			// A deferred command
			if cmdStruct.Defer.Cmd != "" {
				c.Defer = true
				c.Cmd = cmdStruct.Defer.Cmd
				c.StdoutFile = cmdStruct.StdoutFile
				if cmdStruct.ThisSsh == nil {
					c.ThisSsh = true
				} else {
					c.ThisSsh = *cmdStruct.ThisSsh
				}
				c.If = cmdStruct.If
				c.Silent = cmdStruct.Silent
				return nil
			}

			// A deferred task call
			if cmdStruct.Defer.Task != "" {
				c.Defer = true
				c.Task = cmdStruct.Defer.Task
				if cmdStruct.ThisSsh == nil {
					c.ThisSsh = false
				} else {
					c.ThisSsh = *cmdStruct.ThisSsh
				}
				c.Vars = cmdStruct.Defer.Vars
				c.Silent = cmdStruct.Defer.Silent
				return nil
			}
			return nil
		}

		// A task call
		if cmdStruct.Task != "" {
			c.Task = cmdStruct.Task
			if cmdStruct.ThisSsh == nil {
				c.ThisSsh = false
			} else {
				c.ThisSsh = *cmdStruct.ThisSsh
			}
			c.Vars = cmdStruct.Vars
			c.For = cmdStruct.For
			c.Silent = cmdStruct.Silent
			return nil
		}

		// A command with additional options
		if cmdStruct.Cmd != "" {
			c.Cmd = cmdStruct.Cmd
			c.StdoutFile = cmdStruct.StdoutFile
			if cmdStruct.ThisSsh == nil {
				c.ThisSsh = true
			} else {
				c.ThisSsh = *cmdStruct.ThisSsh
			}
			c.If = cmdStruct.If
			c.For = cmdStruct.For
			c.Silent = cmdStruct.Silent
			c.Set = cmdStruct.Set
			c.Shopt = cmdStruct.Shopt
			c.IgnoreError = cmdStruct.IgnoreError
			c.Platforms = cmdStruct.Platforms
			c.Interp = cmdStruct.Interp
			return nil
		}

		return errors.NewTaskfileDecodeError(nil, node).WithMessage("invalid keys in command")
	}

	return errors.NewTaskfileDecodeError(nil, node).WithTypeMessage("command")
}
