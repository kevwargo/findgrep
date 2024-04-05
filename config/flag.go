package config

import (
	"fmt"
	"slices"
	"strconv"

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
	switch f.o.CustomType {
	case "":
		if raw != "true" {
			return fmt.Errorf("internal error: unexpected bool value: %q", raw)
		}
		f.o.Value = true
	case "str":
		f.o.Value = raw
	case "int":
		x, err := strconv.Atoi(raw)
		if err != nil {
			return err
		}
		f.o.Value = x
	default:
		// shouldn't happen, as the types are validated when unmarshalling
		return fmt.Errorf("invalid option type %q", f.o.CustomType)
	}

	if f.o.AllowedValues != nil {
		if !slices.Contains(f.o.AllowedValues, f.o.Value) {
			return fmt.Errorf("%q is not one of the allowed %v", f.o.Value, f.o.AllowedValues)
		}
	}

	return nil
}

func (f *flag) Type() string {
	if f.o.CustomType == "" {
		return "bool"
	}

	return f.o.CustomType
}
