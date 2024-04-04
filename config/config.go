package config

import (
	"cmp"
	"errors"
	"fmt"
	"io/fs"
	"os"
	"path/filepath"
	"slices"

	"gopkg.in/yaml.v3"
)

type Config struct {
	ExcludePaths Options `yaml:"exclude-paths"`
	ExcludeFiles Options `yaml:"exclude-files"`
	SelectFiles  Options `yaml:"select-files"`
	Grep         Options `yaml:"grep"`
}

const FileName = ".findgrep.yml"

func Load(dir string) (*Config, error) {
	path, err := filepath.Abs(dir)
	if err != nil {
		return nil, err
	}

	return load(path, func(n string) (fs.File, error) { return os.Open(n) })
}

func load(path string, openFn func(string) (fs.File, error)) (*Config, error) {
	var files []fs.File
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

	var cfg Config
	for i := len(files) - 1; i >= 0; i-- {
		if err := yaml.NewDecoder(files[i]).Decode(&cfg); err != nil {
			return nil, err
		}
	}

	return &cfg, nil
}

type Option struct {
	Name       string        `yaml:"-"`
	Arg        string        `yaml:"arg"`
	Alias      string        `yaml:"alias"`
	Key        string        `yaml:"key"`
	Default    any           `yaml:"default"`
	Type       string        `yaml:"type"`
	MutexGroup string        `yaml:"mutex-group"`
	Disabled   *bool         `yaml:"disabled"`
	Pattern    stringOrSlice `yaml:"pattern"`
	Target     stringOrSlice `yaml:"target"`
}

type Options []Option

type stringOrSlice []string

func (o *Options) UnmarshalYAML(n *yaml.Node) error {
	var newOptions map[string]Option
	if err := n.Decode(&newOptions); err != nil {
		return fmt.Errorf("decoding options map: %w", err)
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

func (o *Option) merge(src Option) {
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
