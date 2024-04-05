package cmd

import (
	"fmt"

	"github.com/spf13/cobra"
	"github.com/spf13/pflag"

	"github.com/kevwargo/findgrep/config"
)

func Execute() error {
	cfg, err := config.Load(".")
	if err != nil {
		return err
	}

	var printCmd, printElispTransient bool

	c := &cobra.Command{
		SilenceErrors: true,
		SilenceUsage:  true,
		RunE: func(_ *cobra.Command, patterns []string) error {
			if printElispTransient {
				return printElisp(cfg)
			}

			findCmd := buildCommand(cfg, patterns)
			if !printCmd {
				return run(findCmd)
			}

			fmt.Printf("%q\n", findCmd.Args)
			return nil
		},
	}

	c.Flags().BoolVar(&printCmd, "print-cmd", false, "Print the find command without actually executing")
	c.Flags().BoolVar(&printElispTransient, "print-elisp-transient", false, "Print the Emacs Lisp transient config")

	registerConfigFlags(cfg, c.Flags())

	return c.Execute()
}

func registerConfigFlags(cfg *config.Config, flagSet *pflag.FlagSet) {
	for _, opt := range cfg.ExcludePaths {
		opt.RegisterFlag(flagSet, "exclude-")
	}
	for _, opt := range cfg.IgnoreFiles {
		opt.RegisterFlag(flagSet, "ignore-")
	}
	for _, opt := range cfg.SelectFiles {
		opt.RegisterFlag(flagSet, "only-")
	}
	for _, opt := range cfg.Grep {
		opt.RegisterFlag(flagSet, "")
	}
}
