package config

import (
	"github.com/spf13/pflag"
)

const (
	FileName = ".findgrep.yml"

	TypeBool   = "bool"
	TypeString = "str"
	TypeInt    = "int"

	MiscGzip    = "gzip"
	MiscVerbose = "verbose"
)

type Config struct {
	ExcludePaths *OptionGroup `yaml:"exclude-paths"`
	IgnoreFiles  *OptionGroup `yaml:"ignore-files"`
	SelectFiles  *OptionGroup `yaml:"select-files"`
	Grep         *OptionGroup `yaml:"grep"`
	Misc         *OptionGroup `yaml:"misc"`
}

type OptionGroup struct {
	optmap  map[string]*Option
	ordered []*Option
}

type Option struct {
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

	flag *pflag.Flag
}

func (c *Config) OptionGroups() []*OptionGroup {
	return []*OptionGroup{c.ExcludePaths, c.IgnoreFiles, c.SelectFiles, c.Grep, c.Misc}
}

func (o *OptionGroup) All() []*Option {
	return o.ordered
}

func (o *OptionGroup) IsSet(name string) bool {
	if opt := o.optmap[name]; opt != nil {
		return opt.IsSet()
	}

	return false
}

func (o *Option) IsSet() bool {
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
	return o.OptType == TypeBool
}

func (o *Option) isInverted() bool {
	return o.IsBool() && o.Default == true
}
