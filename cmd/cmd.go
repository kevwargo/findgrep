package cmd

import (
	"fmt"
	"strings"

	"github.com/spf13/cobra"
	"github.com/spf13/pflag"

	"github.com/kevwargo/findgrep/config"
)

func Execute() error {
	cfg, err := config.Load(".")
	if err != nil {
		return err
	}

	c := &cobra.Command{
		SilenceErrors: true,
		SilenceUsage:  true,
		RunE: func(_ *cobra.Command, args []string) error {
			return run(cfg, args)
		},
	}

	registerConfigFlags(cfg, c.Flags())

	return c.Execute()
}

func run(cfg *config.Config, patterns []string) error {
	args := []string{"find", "."}

	for _, opt := range cfg.ExcludePaths {
		for _, pattern := range opt.Pattern {
			if !(strings.ContainsAny(pattern, "*/")) {
				pattern = fmt.Sprintf("*/%s/*", pattern)
			}

			args = opt.AppendArgs(args, "!", "-path", pattern)
		}
	}

	for _, opt := range cfg.IgnoreFiles {
		for _, pattern := range opt.Pattern {
			args = opt.AppendArgs(args, "!", "-name", pattern)
		}
	}

	for _, opt := range cfg.SelectFiles {
		for _, pattern := range opt.Pattern {
			args = opt.AppendArgs(args, "-name", pattern)
		}
	}

	args = append(args, "-exec", "grep")
	for _, opt := range cfg.Grep {
		args = opt.AppendArgs(args)
	}
	for _, pattern := range patterns {
		args = append(args, "-e", pattern)
	}
	args = append(args, "{}", "+")

	fmt.Println(args)

	return nil
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
