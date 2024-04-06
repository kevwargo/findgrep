package cmd

import (
	"bytes"
	"fmt"
	"os"
	"os/exec"
	"strings"

	"github.com/kevwargo/findgrep/config"
)

func run(c *exec.Cmd) error {
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

func buildCommand(cfg *config.Config, patterns []string) *exec.Cmd {
	args := []string{"."}

	args = append(args, buildExcludePaths(cfg)...)

	args = append(args, "-type", "f")
	args = append(args, buildIgnoreFiles(cfg)...)
	args = append(args, buildSelectFiles(cfg)...)

	args = append(args, "-exec", "grep")
	args = append(args, buildGrep(cfg, patterns)...)

	args = append(args, "{}", "+")

	return exec.Command(findExecutable, args...)
}

func buildExcludePaths(cfg *config.Config) (args []string) {
	for _, opt := range cfg.ExcludePaths {
		if !opt.Active() {
			continue
		}
		for _, pattern := range opt.Pattern {
			if !(strings.ContainsAny(pattern, "*/")) {
				pattern = fmt.Sprintf("*/%s/*", pattern)
			}

			args = append(args, "!", "-path", pattern)
		}
	}

	return
}

func buildIgnoreFiles(cfg *config.Config) (args []string) {
	for _, opt := range cfg.IgnoreFiles {
		if !opt.Active() {
			continue
		}
		for _, pattern := range opt.Pattern {
			args = append(args, "!", "-name", pattern)
		}
	}

	return args
}

func buildSelectFiles(cfg *config.Config) (args []string) {
	for _, opt := range cfg.SelectFiles {
		if !opt.Active() {
			continue
		}
		switch len(opt.Pattern) {
		case 0:
		case 1:
			args = append(args, "-name", opt.Pattern[0])
		default:
			args = append(args, "(")
			for idx, pattern := range opt.Pattern {
				if idx > 0 {
					args = append(args, "-o")
				}
				args = append(args, "-name", pattern)
			}
			args = append(args, ")")
		}

	}

	return args
}

func buildGrep(cfg *config.Config, patterns []string) (args []string) {
	for _, opt := range cfg.Grep {
		if !opt.Active() {
			continue
		}

		value := opt.Value
		if value == nil {
			value = opt.Default
		}

		if value == true {
			args = append(args, opt.Target...)
			continue
		}

		for _, target := range opt.Target {
			v := fmt.Sprint(value)
			if strings.HasSuffix(target, "=") {
				args = append(args, target+v)
			} else {
				args = append(args, target, v)
			}
		}
	}
	for _, pattern := range patterns {
		args = append(args, "-e", pattern)
	}

	return args
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

const (
	findExecutable = "find"
)
