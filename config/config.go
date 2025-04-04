package config

import (
	"cmp"
	"errors"
	"fmt"
	"slices"

	"gopkg.in/yaml.v3"
)

type Config struct {
	ExcludePaths Options `yaml:"exclude-paths" json:"exclude-paths,omitempty"`
	ExcludeFiles Options `yaml:"exclude-files" json:"exclude-files,omitempty"`
	SelectFiles  Options `yaml:"select-files" json:"select-files,omitempty"`
	Grep         Options `yaml:"grep" json:"grep,omitempty"`
}

type Option struct {
	Name       string        `yaml:"-" json:"name,omitempty"`
	Arg        string        `yaml:"arg" json:"arg,omitempty"`
	Alias      string        `yaml:"alias" json:"alias,omitempty"`
	Key        string        `yaml:"key" json:"key,omitempty"`
	Default    any           `yaml:"default" json:"default,omitempty"`
	MutexGroup string        `yaml:"mutex-group" json:"mutex-group,omitempty"`
	Disabled   *bool         `yaml:"disabled" json:"disabled,omitempty"`
	Pattern    stringOrSlice `yaml:"pattern" json:"pattern,omitempty"`
	Target     stringOrSlice `yaml:"target" json:"target,omitempty"`
}

type Options []Option

func (o *Options) UnmarshalYAML(n *yaml.Node) error {
	var newOptions map[string]Option
	if err := n.Decode(&newOptions); err != nil {
		return fmt.Errorf("decoding optoins map: %w", err)
	}

	for idx := range *o {
		name := (*o)[idx].Name
		if newOpt, ok := newOptions[name]; ok {
			(*o)[idx].merge(newOpt)
			delete(newOptions, name)
		}
	}

	for name, newOpt := range newOptions {
		newOpt.Name = name
		*o = append(*o, newOpt)
	}

	slices.SortFunc(*o, func(opt1, opt2 Option) int {
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

func yamlKind(k yaml.Kind) string {
	switch k {
	case yaml.DocumentNode:
		return "document"
	case yaml.SequenceNode:
		return "sequence"
	case yaml.MappingNode:
		return "mapping"
	case yaml.ScalarNode:
		return "scalar"
	case yaml.AliasNode:
		return "alias"
	default:
		return "<unknown>"
	}
}
