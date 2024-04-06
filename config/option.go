package config

import (
	"cmp"
	"errors"
	"fmt"
	"slices"
	"strconv"

	"github.com/spf13/pflag"
	"gopkg.in/yaml.v3"
)

type option struct {
	Name  string `yaml:"-"`
	Value any    `yaml:"-"`

	Arg           string        `yaml:"arg"`
	Alias         string        `yaml:"alias"`
	Key           string        `yaml:"key"`
	Default       any           `yaml:"default"`
	OptType       string        `yaml:"type"`
	AllowedValues []any         `yaml:"allowed-values"`
	MutexGroup    string        `yaml:"mutex-group"`
	Disabled      *bool         `yaml:"disabled"`
	Pattern       stringOrSlice `yaml:"pattern"`
	Target        stringOrSlice `yaml:"target"`

	flag       *pflag.Flag
	setValueFn func(string) (any, error)
}

type (
	Option  option
	Options []*Option
)

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

func (o *Option) UnmarshalYAML(n *yaml.Node) error {
	var opt option
	if err := n.Decode(&opt); err != nil {
		return err
	}
	if err := opt.validate(); err != nil {
		return fmt.Errorf("invalid option %+v at (%d:%d): %w", opt, n.Line, n.Column, err)
	}

	*o = Option(opt)
	fmt.Printf("opttype set to %q after validation\n", o.OptType)

	return nil
}

func (o *Option) Active() bool {
	if o.IsBool() {
		if o.isInverted() {
			return o.Value != true
		}
		return o.Value == true
	}

	if o.Value != nil {
		return true
	}

	return o.Default != nil
}

func (o *Option) Flag() *pflag.Flag {
	return o.flag
}

func (o *Option) IsBool() bool {
	return o.OptType == "bool"
}

func (o *Option) isInverted() bool {
	return o.IsBool() && o.Default == true
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

func (o *option) validate() error {
	switch o.OptType {
	case "":
		o.OptType = "bool"
		fallthrough
	case "bool":
		if o.AllowedValues != nil {
			return errors.New(`an explicit type should be specified along with "allowed-values"`)
		}
		o.setValueFn = func(_ string) (any, error) { return true, nil }
		return validateType[bool](o)
	case "str":
		o.setValueFn = func(raw string) (any, error) { return raw, nil }
		return validateType[string](o)
	case "int":
		o.setValueFn = func(raw string) (any, error) { return strconv.Atoi(raw) }
		return validateType[int](o)
	default:
		return fmt.Errorf("invalid option type: %q", o.OptType)
	}
}

func validateType[T any](o *option) error {
	for _, v := range o.AllowedValues {
		if _, ok := v.(T); !ok {
			return fmt.Errorf(`invalid allowed-value %v(%T) for type %q`, v, v, o.OptType)
		}
	}

	if d := o.Default; d != nil {
		if _, ok := d.(T); !ok {
			return fmt.Errorf("invalid default %v(%T) for type %q", d, d, o.OptType)
		}
	}

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
