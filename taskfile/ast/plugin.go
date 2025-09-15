package ast

import (
	"encoding/json"
	"iter"
	"path/filepath"
	"strings"
	"sync"

	"github.com/elliotchance/orderedmap/v3"
	"gopkg.in/yaml.v3"

	"github.com/go-task/task/v3/errors"
	"github.com/go-task/task/v3/internal/deepcopy"
)

type (
	Plugin struct {
		File         string
		Mounts       map[string]string
		SysNanosleep bool `yaml:"sys_nanosleep"`
		SysNanotime  bool `yaml:"sys_nanotime"`
		SysWalltime  bool `yaml:"sys_walltime"`
		Rand         bool
		Stderr       bool
		Stdout       bool
	}
	Plugins struct {
		om    *orderedmap.OrderedMap[string, *Plugin]
		mutex sync.RWMutex
	}
	PluginElement orderedmap.Element[string, *Plugin]
)

func NewPlugins(els ...*PluginElement) *Plugins {
	plugins := &Plugins{
		om: orderedmap.NewOrderedMap[string, *Plugin](),
	}
	for _, el := range els {
		plugins.Set(el.Key, el.Value)
	}
	return plugins
}

func (plugins *Plugins) Len() int {
	if plugins == nil || plugins.om == nil {
		return 0
	}
	defer plugins.mutex.RUnlock()
	plugins.mutex.RLock()
	return plugins.om.Len()
}

func (plugins *Plugins) Get(key string) (*Plugin, bool) {
	if plugins == nil || plugins.om == nil {
		return &Plugin{}, false
	}
	defer plugins.mutex.RUnlock()
	plugins.mutex.RLock()
	return plugins.om.Get(key)
}

func (plugins *Plugins) Set(key string, value *Plugin) bool {
	if plugins == nil {
		plugins = NewPlugins()
	}
	if plugins.om == nil {
		plugins.om = orderedmap.NewOrderedMap[string, *Plugin]()
	}
	defer plugins.mutex.Unlock()
	plugins.mutex.Lock()
	return plugins.om.Set(key, value)
}

func (plugins *Plugins) All() iter.Seq2[string, *Plugin] {
	if plugins == nil || plugins.om == nil {
		return func(yield func(string, *Plugin) bool) {}
	}
	return plugins.om.AllFromFront()
}

func (plugins *Plugins) Keys() iter.Seq[string] {
	if plugins == nil || plugins.om == nil {
		return func(yield func(string) bool) {}
	}
	return plugins.om.Keys()
}

func (plugins *Plugins) Values() iter.Seq[*Plugin] {
	if plugins == nil || plugins.om == nil {
		return func(yield func(*Plugin) bool) {}
	}
	return plugins.om.Values()
}

func (plugins *Plugins) UnmarshalYAML(node *yaml.Node) error {
	if plugins == nil || plugins.om == nil {
		*plugins = *NewPlugins()
	}

	switch node.Kind {
	case yaml.SequenceNode:
		var list []any
		if err := node.Decode(&list); err != nil {
			return errors.NewTaskfileDecodeError(err, node)
		}
		for _, item := range list {
			switch v := item.(type) {
			case string:
				fn := filepath.Base(v)
				name := strings.TrimSuffix(fn, filepath.Ext(fn))
				plugins.Set(name, &Plugin{File: v})
			default:
				marshalled, _ := json.Marshal(v)
				var unmarshalled Plugin
				if err := json.Unmarshal(marshalled, &unmarshalled); err == nil {
					fn := filepath.Base(unmarshalled.File)
					name := strings.TrimSuffix(fn, filepath.Ext(fn))
					plugins.Set(name, &unmarshalled)
				}
			}
		}
		return nil
	case yaml.MappingNode:
		for i := 0; i < len(node.Content); i += 2 {
			keyNode := node.Content[i]
			valueNode := node.Content[i+1]

			var v Plugin
			if err := valueNode.Decode(&v); err != nil {
				return errors.NewTaskfileDecodeError(err, node)
			}

			plugins.Set(keyNode.Value, &v)
		}
		return nil
	}

	return errors.NewTaskfileDecodeError(nil, node).WithTypeMessage("plugins")
}

func (plugin *Plugin) UnmarshalYAML(node *yaml.Node) error {
	switch node.Kind {
	case yaml.ScalarNode:
		var str string
		if err := node.Decode(&str); err != nil {
			return errors.NewTaskfileDecodeError(err, node)
		}
		plugin.File = str
		return nil

	case yaml.MappingNode:
		var v struct {
			File         string
			Mounts       map[string]string
			SysNanosleep bool `yaml:"sys_nanosleep"`
			SysNanotime  bool `yaml:"sys_nanotime"`
			SysWalltime  bool `yaml:"sys_walltime"`
			Rand         bool
			Stderr       bool
			Stdout       bool
		}
		if err := node.Decode(&v); err != nil {
			return errors.NewTaskfileDecodeError(err, node)
		}
		plugin.File = v.File
		plugin.Mounts = v.Mounts
		plugin.SysNanosleep = v.SysNanosleep
		plugin.SysNanotime = v.SysNanotime
		plugin.SysWalltime = v.SysWalltime
		plugin.Rand = v.Rand
		plugin.Stderr = v.Stderr
		plugin.Stdout = v.Stdout
		return nil
	}

	return errors.NewTaskfileDecodeError(nil, node).WithTypeMessage("plugin")
}

func (plugin *Plugin) DeepCopy() *Plugin {
	if plugin == nil {
		return nil
	}
	return &Plugin{
		File:         plugin.File,
		Mounts:       deepcopy.Map(plugin.Mounts),
		SysNanosleep: plugin.SysNanosleep,
		SysNanotime:  plugin.SysNanotime,
		SysWalltime:  plugin.SysWalltime,
		Rand:         plugin.Rand,
		Stderr:       plugin.Stderr,
		Stdout:       plugin.Stdout,
	}
}
