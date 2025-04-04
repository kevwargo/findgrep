package cmd

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
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

	var printCmd, printElispTransient, printJSON bool

	c := &cobra.Command{
		SilenceErrors: true,
		SilenceUsage:  true,
		RunE: func(_ *cobra.Command, patterns []string) error {
			if printJSON {
				return printConfigJSON(cfg)
			}

			if printElispTransient {
				return elisp.Print(cfg)
			}

			findCmd := buildCommand(cfg, patterns)

			if printCmd {
				printCommand(os.Stdout, findCmd.Args...)
				return nil
			}

			if cfg.Misc.IsSet(config.MiscVerbose) {
				printCommand(os.Stderr, findCmd.Args...)
			}
			return runCommand(findCmd)
		},
	}

	c.Flags().BoolVar(&printCmd, "print-cmd", false, "Print the find command without actually executing")
	c.Flags().BoolVar(&printElispTransient, "print-elisp-transient", false, "Print the Emacs Lisp transient config")
	c.Flags().BoolVar(&printJSON, "print-json", false, "Print config in JSON format")

	registerConfigFlags(cfg, c.Flags())

	return c.Execute()
}

func printConfigJSON(cfg *config.Config) error {
	enc := json.NewEncoder(os.Stdout)
	enc.SetIndent("", "  ")
	return enc.Encode(cfg)
}

func registerConfigFlags(cfg *config.Config, flagSet *pflag.FlagSet) {
	cfg.ExcludePaths.RegisterFlag(flagSet, "exclude-")
	cfg.IgnoreFiles.RegisterFlag(flagSet, "ignore-")
	cfg.SelectFiles.RegisterFlag(flagSet, "only-")
	cfg.Grep.RegisterFlag(flagSet, "")
	cfg.Misc.RegisterFlag(flagSet, "")
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

func printCommand(w io.Writer, args ...string) {
	for i, arg := range args {
		if strings.ContainsAny(arg, ` "'`) {
			args[i] = fmt.Sprintf("%q", arg)
		}
	}

	fmt.Fprintln(w, strings.Join(args, " "))
}
