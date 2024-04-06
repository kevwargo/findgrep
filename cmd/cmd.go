package cmd

import (
	"bytes"
	"fmt"
	"os"
	"os/exec"
	"strings"

	"github.com/spf13/cobra"
	"github.com/spf13/pflag"

	"github.com/kevwargo/findgrep/cmd/elisp"
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
				return elisp.Print(cfg)
			}

			findCmd := buildCommand(cfg, patterns)
			if printCmd {
				return printCommand(findCmd.Args...)
			}

			return runCommand(findCmd)
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

	cfg.Misc.Gzip.RegisterFlag(flagSet, "")
}

func runCommand(c *exec.Cmd) error {
	c.Stdout = os.Stdout
	stderr := bytes.Buffer{}
	c.Stderr = &stderr

	err := c.Run()
	if err == nil {
		return nil
	}

	if ee, ok := err.(*exec.ExitError); ok {
		if ee.ExitCode() == 1 && stderr.Len() == 0 {
			return nil
		}
	}

	return fmt.Errorf("%w: %s", err, strings.Trim(stderr.String(), "\n"))
}

func printCommand(args ...string) error {
	for i, arg := range args {
		if strings.ContainsAny(arg, ` "'`) {
			args[i] = fmt.Sprintf("%q", arg)
		}
	}

	fmt.Println(strings.Join(args, " "))

	return nil
}
