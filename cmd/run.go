package cmd

import (
	"fmt"
	"os"
	"os/exec"
	"strings"

	"github.com/kevwargo/findgrep/config"
)

func run(c *exec.Cmd) error {
	c.Stdout = os.Stdout
	c.Stderr = os.Stderr

	return c.Run()
}

func buildCommand(cfg *config.Config, patterns []string) *exec.Cmd {
	args := []string{"."}

	for _, opt := range cfg.ExcludePaths {
		for _, pattern := range opt.Pattern {
			if !(strings.ContainsAny(pattern, "*/")) {
				pattern = fmt.Sprintf("*/%s/*", pattern)
			}

			args = opt.AppendArgs(args, "!", "-path", pattern)
		}
	}

	args = append(args, "-type", "f")
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

	return exec.Command(findExecutable, args...)
}

const (
	findExecutable = "find"
)
