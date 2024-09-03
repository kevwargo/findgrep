package config

import (
	"fmt"
	"slices"
	"strconv"

	"github.com/spf13/cobra"
	"github.com/spf13/pflag"
)

func (o OptionGroup) RegisterFlag(flagSet *pflag.FlagSet, prefix string) {
	for _, opt := range o.ordered {
		opt.RegisterFlag(flagSet, prefix)
	}
}

func (o *Option) RegisterFlag(flagSet *pflag.FlagSet, prefix string) {
	name := o.Arg

	if name == "" {
		name = prefix + o.Name
	}

	noOptVal := ""
	if o.IsBool() {
		noOptVal = "true"
	}
	if o.isInverted() {
		name = "no-" + name
	}

	f := flagSet.VarPF(&optionFlag{o}, name, o.Alias, name)
	f.NoOptDefVal = noOptVal
	if o.AllowedValues != nil && o.Default == nil {
		cobra.MarkFlagRequired(flagSet, name)
	}
	o.flag = f
}

type optionFlag struct {
	o *Option
}

func (f *optionFlag) String() string {
	return fmt.Sprintf("%v", f.o.Value)
}

func (f *optionFlag) Set(raw string) (err error) {
	var v any
	switch f.o.OptType {
	case TypeBool:
		v = true
	case TypeString:
		v = raw
	case TypeInt:
		v, err = strconv.Atoi(raw)
	}

	if err != nil {
		return err
	}

	if f.o.AllowedValues != nil {
		if !slices.Contains(f.o.AllowedValues, v) {
			return fmt.Errorf("%q is not one of the allowed %v", v, f.o.AllowedValues)
		}
	}

	f.o.Value = v

	return nil
}

func (f *optionFlag) Type() string {
	return f.o.OptType
}
