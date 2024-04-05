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

	Arg           string        `yaml:"arg"`
	Alias         string        `yaml:"alias"`
	Key           string        `yaml:"key"`
	Default       any           `yaml:"default"`
	CustomType    string        `yaml:"type"`
	AllowedValues []any         `yaml:"allowed-values"`
	MutexGroup    string        `yaml:"mutex-group"`
	Disabled      *bool         `yaml:"disabled"`
	Pattern       stringOrSlice `yaml:"pattern"`
	Target        stringOrSlice `yaml:"target"`
}

func (o *Option) AppendArgs(args []string, values ...string) []string {
	value := o.Value
	if o.isInverted() {
		if value == true {
			value = nil
		} else {
			value = true
		}
	} else if value == nil {
		value = o.Default
	}

	switch value {
	case nil:
		return args
	case true:
		if len(values) == 0 {
			values = o.Target
		}
	default:
		values = nil
		for _, target := range o.Target {
			values = append(values, target, fmt.Sprint(value))
		}
	}

	return append(args, values...)
}

func (o *Option) isBool() bool {
	return o.CustomType == ""
}

func (o *Option) isInverted() bool {
	return o.isBool() && o.Default == true
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

func (o *Option) validate() (err error) {
	if o.AllowedValues != nil {
		switch o.CustomType {
		case "":
			err = errors.New(`an explicit type should be specified along with "allowed-values"`)
		case "str":
			err = validateType[string](o)
		case "int":
			err = validateType[int](o)
		default:
			err = fmt.Errorf("invalid option type: %q", o.CustomType)
		}
	}

	return err
}

func validateType[T any](o *Option) error {
	for _, v := range o.AllowedValues {
		if _, ok := v.(T); !ok {
			return fmt.Errorf(`invalid allowed-value %v(%T) for type %q`, v, v, o.CustomType)
		}
	}

	return nil
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

	for _, opt := range *o {
		if err := opt.validate(); err != nil {
			return fmt.Errorf("invalid option %q: %w", opt.Name, err)
		}
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
