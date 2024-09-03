package config

import (
	"cmp"
	"errors"
	"fmt"
	"io"
	"io/fs"
	"os"
	"path/filepath"
	"slices"

	"gopkg.in/yaml.v3"
)

func Load(dir string) (*Config, error) {
	path, err := filepath.Abs(dir)
	if err != nil {
		return nil, err
	}

	return load(path, func(n string) (namedFile, error) { return os.Open(n) })
}

type namedFile interface {
	fs.File
	Name() string
}

func load(path string, openFn func(string) (namedFile, error)) (*Config, error) {
	cfg, err := loadDefault()
	if err != nil {
		return nil, err
	}

	var files []namedFile
	defer func() {
		for _, f := range files {
			f.Close()
		}
	}()

	for ; path != filepath.Dir(path); path = filepath.Dir(path) {
		f, err := openFn(filepath.Join(path, FileName))
		if err != nil {
			if errors.Is(err, fs.ErrNotExist) {
				continue
			}
			return nil, err
		}

		files = append(files, f)
	}

	for i := len(files) - 1; i >= 0; i-- {
		if err := yaml.NewDecoder(files[i]).Decode(cfg); err != nil {
			if errors.Is(err, io.EOF) {
				continue
			}

			return nil, fmt.Errorf("loading config file %q: %w", files[i].Name(), err)
		}
	}

	return cfg.stripDisabled(), nil
}

func (c *Config) stripDisabled() *Config {
	for _, group := range c.OptionGroups() {
		group.ordered = slices.DeleteFunc(group.ordered, func(opt *Option) bool {
			if opt.Disabled == nil || !*opt.Disabled {
				return false
			}

			delete(group.optmap, opt.Name)
			return true
		})
	}

	return c
}

func (o *OptionGroup) UnmarshalYAML(n *yaml.Node) error {
	var newOptions map[string]*Option
	if err := n.Decode(&newOptions); err != nil {
		return fmt.Errorf("decoding options map: %w", err)
	}
	if len(newOptions) == 0 {
		return nil
	}

	if o.optmap == nil {
		o.optmap = make(map[string]*Option, len(newOptions))
	}
	for name, newOpt := range newOptions {
		old := o.optmap[name]
		if old != nil {
			old.merge(newOpt)
			newOpt = old
		} else {
			newOpt.Name = name
		}

		if err := newOpt.validate(); err != nil {
			return fmt.Errorf("invalid option %+v: %w", newOpt, err)
		}

		if old == nil {
			o.optmap[name] = newOpt
			o.ordered = append(o.ordered, newOpt)
		}
	}

	slices.SortFunc(o.ordered, func(o1, o2 *Option) int {
		return cmp.Compare(o1.Name, o2.Name)
	})

	return nil
}

func (o *Option) merge(newOpt *Option) {
	if newOpt.Arg != "" {
		o.Arg = newOpt.Arg
	}
	if newOpt.Alias != "" {
		o.Alias = newOpt.Alias
	}
	if newOpt.Key != "" {
		o.Key = newOpt.Key
	}
	if newOpt.Default != nil {
		o.Default = newOpt.Default
	}
	if newOpt.Disabled != nil {
		o.Disabled = newOpt.Disabled
	}
}

func (o *Option) validate() error {
	switch o.OptType {
	case "":
		o.OptType = TypeBool
		fallthrough
	case TypeBool:
		if o.AllowedValues != nil {
			return errors.New(`an explicit type should be specified along with "allowed-values"`)
		}
		return validateType[bool](o)
	case TypeString:
		return validateType[string](o)
	case TypeInt:
		return validateType[int](o)
	default:
		return fmt.Errorf("invalid option type: %q", o.OptType)
	}
}

func validateType[T any](o *Option) error {
	for _, v := range o.AllowedValues {
		if _, ok := v.(T); !ok {
			return fmt.Errorf("invalid allowed value %v(%T) for type %q", v, v, o.OptType)
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
