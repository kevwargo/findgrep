package config

import (
	"fmt"
	"strconv"

	"github.com/spf13/pflag"
)

func (o *Option) RegisterFlag(flagSet *pflag.FlagSet, prefix string) {
	arg := o.Arg

	if arg == "" {
		arg = prefix + o.Name
	}

	noOptVal := ""
	if o.CustomType == "" {
		noOptVal = "true"
		if o.Default == true {
			arg = "no-" + arg
		}
	}

	f := flagSet.VarPF(&flag{o}, arg, o.Alias, arg)
	f.NoOptDefVal = noOptVal
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
	case "int":
		x, err := strconv.Atoi(raw)
		if err != nil {
			return err
		}
		f.o.Value = x
	default:
		return fmt.Errorf("invalid option type %q", f.o.CustomType)
	}

	return nil
}

func (f *flag) Type() string {
	if f.o.CustomType == "" {
		return "bool"
	}

	return f.o.CustomType
}
