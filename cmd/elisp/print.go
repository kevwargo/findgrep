package elisp

import (
	"fmt"

	"github.com/kevwargo/findgrep/config"
)

func Print(cfg *config.Config) error {
	if err := resolveKeys(cfg.SelectFiles, cfg.Grep, cfg.ExcludePaths, cfg.IgnoreFiles, cfg.Misc); err != nil {
		return err
	}

	for _, opt := range cfg.SelectFiles.All() {
		opt.MutexGroup = "select"
	}

	fmt.Print("(")
	printGroup("Exclude paths", cfg.ExcludePaths, true, false)
	printGroup("Ignore files", cfg.IgnoreFiles, false, false)
	printGroup("Select files", cfg.SelectFiles, false, false)
	printGroup("Grep options", cfg.Grep, false, false)
	printGroup("Misc options", cfg.Misc, false, true)
	fmt.Println(")")

	return nil
}

func printGroup(name string, optionGroup config.OptionGroup, first, last bool) {
	if !first {
		fmt.Print(" ")
	}
	fmt.Printf("[%q\n", name)

	options := optionGroup.All()
	count := len(options)
	for i := 0; i < count; i++ {
		printOption(options[i], i == count-1)
	}

	fmt.Print("]")
	if !last {
		fmt.Println()
	}
}

func printOption(opt *config.Option, last bool) {
	longArg := "--" + opt.Flag().Name
	class := "findgrep-switch"
	if !opt.IsBool() {
		class = "findgrep-option"
		longArg += "="
	}

	var mutex string
	if m := opt.MutexGroup; m != "" {
		mutex = fmt.Sprintf(" :mutex-group %s", m)
	}

	name := opt.Name
	if name == "" {
		name = opt.Flag().Name
	}

	fmt.Printf("  (%q %q %q :class %s%s)", opt.Key, name, longArg, class, mutex)
	if !last {
		fmt.Println()
	}
}
