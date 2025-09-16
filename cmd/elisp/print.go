package elisp

import (
	"fmt"

	"github.com/kevwargo/findgrep/config"
)

func Print(cfg *config.Config) error {
	if err := resolveKeys(cfg); err != nil {
		return err
	}

	for _, opt := range cfg.SelectFiles.All() {
		opt.MutexGroup = "select"
	}

	groups := collectGroups(cfg)
	groupsCount := len(groups)

	fmt.Print("(")
	for idx, g := range groups {
		printGroup(g.title, g.group, idx == 0, idx == groupsCount-1)
	}
	fmt.Println(")")

	return nil
}

type optionGroup struct {
	title string
	group *config.OptionGroup
}

func collectGroups(cfg *config.Config) (groups []optionGroup) {
	if !cfg.ExcludePaths.Empty() {
		groups = append(groups, optionGroup{title: "Exclude paths", group: cfg.ExcludePaths})
	}
	if !cfg.IgnoreFiles.Empty() {
		groups = append(groups, optionGroup{title: "Ignore files", group: cfg.IgnoreFiles})
	}
	if !cfg.SelectFiles.Empty() {
		groups = append(groups, optionGroup{title: "Select files", group: cfg.SelectFiles})
	}
	if !cfg.Grep.Empty() {
		groups = append(groups, optionGroup{title: "Grep options", group: cfg.Grep})
	}
	if !cfg.Misc.Empty() {
		groups = append(groups, optionGroup{title: "Misc options", group: cfg.Misc})
	}

	return
}

func printGroup(name string, optionGroup *config.OptionGroup, first, last bool) {
	if !first {
		fmt.Print(" ")
	}
	fmt.Printf("[%q\n", name)

	options := optionGroup.All()
	count := len(options)
	for i := range count {
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
