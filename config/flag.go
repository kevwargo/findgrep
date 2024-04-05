package config

import (
	"fmt"
	"slices"

	"github.com/spf13/cobra"
	"github.com/spf13/pflag"
)

func (o *Option) RegisterFlag(flagSet *pflag.FlagSet, prefix string) {
	name := o.Arg

	if name == "" {
		name = prefix + o.Name
	}

	noOptVal := ""
	if o.isBool() {
		noOptVal = "true"
	}
	if o.isInverted() {
		name = "no-" + name
	}

	f := flagSet.VarPF(&flag{o}, name, o.Alias, name)
	f.NoOptDefVal = noOptVal
	if o.AllowedValues != nil && o.Default == nil {
		cobra.MarkFlagRequired(flagSet, name)
	}
}

type flag struct {
	o *Option
}

func (f *flag) String() string {
	return fmt.Sprintf("%v", f.o.Value)
}

func (f *flag) Set(raw string) error {
	v, err := f.o.setValueFn(raw)
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

func (f *flag) Type() string {
	return f.o.CustomType
}
