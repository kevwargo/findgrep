package cmd

import (
	"fmt"
	"os/exec"
	"strings"

	"github.com/kevwargo/findgrep/config"
)

func buildCommand(cfg *config.Config, patterns []string) *exec.Cmd {
	args := []string{"."}

	args = append(args, buildExcludePaths(cfg)...)

	args = append(args, "-type", "f")
	args = append(args, buildIgnoreFiles(cfg)...)
	args = append(args, buildSelectFiles(cfg)...)

	grep := grepExecutable
	if cfg.Misc.Gzip.Active() {
		grep = zgrepExecutable
	}
	args = append(args, "-exec", grep)

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
	convertPattern := func(p string) string { return p }
	if cfg.Misc.Gzip.Active() {
		convertPattern = func(p string) string { return p + ".gz" }
	}

	for _, opt := range cfg.SelectFiles {
		if !opt.Active() {
			continue
		}

		switch len(opt.Pattern) {
		case 0:
		case 1:
			args = append(args, "-name", convertPattern(opt.Pattern[0]))
		default:
			args = append(args, "(")
			for idx, pattern := range opt.Pattern {
				if idx > 0 {
					args = append(args, "-o")
				}
				args = append(args, "-name", convertPattern(pattern))
			}
			args = append(args, ")")
		}
	}

	if len(args) == 0 && cfg.Misc.Gzip.Active() {
		args = []string{"-name", "*.gz"}
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

const (
	findExecutable  = "find"
	grepExecutable  = "grep"
	zgrepExecutable = "zgrep"
)
