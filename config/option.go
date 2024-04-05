package config

import (
	"cmp"
	"errors"
	"fmt"
	"slices"

	"gopkg.in/yaml.v3"
)

type Option struct {
	Name  string `yaml:"-"`
	Value any    `yaml:"-"`

	Arg        string        `yaml:"arg"`
	Alias      string        `yaml:"alias"`
	Key        string        `yaml:"key"`
	Default    any           `yaml:"default"`
	CustomType string        `yaml:"type"`
	MutexGroup string        `yaml:"mutex-group"`
	Disabled   *bool         `yaml:"disabled"`
	Pattern    stringOrSlice `yaml:"pattern"`
	Target     stringOrSlice `yaml:"target"`
}

func (o *Option) AppendArgs(args []string, values ...string) []string {
	if o.skip() {
		return args
	}

	if len(values) == 0 {
		switch o.Value.(type) {
		case nil, bool:
			values = o.Target
		default:
			for _, target := range o.Target {
				values = append(values, target, fmt.Sprint(o.Value))
			}
		}
	}

	return append(args, values...)
}

func (o *Option) skip() bool {
	valueEmpty := false
	switch o.Value {
	case nil, false, "":
		valueEmpty = true
	}

	defaultEmpty := false
	switch o.Default {
	case nil, false, "":
		defaultEmpty = true
	}

	return valueEmpty == defaultEmpty
}

func (o *Option) merge(src *Option) {
	if src.Arg != "" {
		o.Arg = src.Arg
	}
	if src.Alias != "" {
		o.Alias = src.Alias
	}
	if src.Key != "" {
		o.Key = src.Key
	}
	if src.Default != nil {
		o.Default = src.Default
	}
	if src.Disabled != nil {
		o.Disabled = src.Disabled
	}
}

type Options []*Option

func (o *Options) UnmarshalYAML(n *yaml.Node) error {
	var newOptions map[string]*Option
	if err := n.Decode(&newOptions); err != nil {
		return fmt.Errorf("decoding options map: %w", err)
	}

	for idx := range *o {
		name := (*o)[idx].Name
		if newOpt, ok := newOptions[name]; ok {
			if newOpt != nil {
				(*o)[idx].merge(newOpt)
			}
			delete(newOptions, name)
		}
	}

	for name, newOpt := range newOptions {
		newOpt.Name = name
		*o = append(*o, newOpt)
	}

	slices.SortFunc(*o, func(opt1, opt2 *Option) int {
		return cmp.Compare(opt1.Name, opt2.Name)
	})

	return nil
}

type stringOrSlice []string

func (s *stringOrSlice) UnmarshalYAML(n *yaml.Node) error {
	var sliceErr error
	var slice []string
	if sliceErr = n.Decode(&slice); sliceErr == nil {
		*s = slice
		return nil
	}

	var str string
	if err := n.Decode(&str); err != nil {
		return errors.Join(sliceErr, err)
	}

	*s = []string{str}

	return nil
}
