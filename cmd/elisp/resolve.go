package elisp

import (
	"fmt"
	"strings"
	"unicode"

	"github.com/kevwargo/findgrep/config"
)

func resolveKeys(optionSets ...config.Options) error {
	if allowedKeys == nil {
		allowedKeys = generateAllowedKeys()
	}

	s, err := initState(optionSets)
	if err != nil {
		return err
	}

	return s.resolveAll()
}

type state struct {
	used       map[string]*config.Option
	unresolved config.Options
}

func initState(optionSets []config.Options) (*state, error) {
	used := make(map[string]*config.Option)
	var unresolved config.Options

	for _, options := range optionSets {
		for _, opt := range options {
			if k := opt.Key; k != "" {
				if o := used[k]; o != nil {
					return nil, fmt.Errorf("key %q used by both %q and %q", k, opt.Flag().Name, o.Flag().Name)
				}
				used[k] = opt
			} else {
				unresolved = append(unresolved, opt)
			}
		}
	}

	return &state{
		used:       used,
		unresolved: unresolved,
	}, nil
}

func (s *state) resolveAll() error {
	if s.useAliases() {
		return nil
	}

	for resolved, pos := false, 0; !resolved; pos++ {
		resolved = true
		for i, opt := range s.unresolved {
			if opt == nil {
				continue
			}

			resolved = false

			if pos < len(opt.Name) {
				if c := string(opt.Name[pos]); !s.resolve(i, c) {
					s.resolve(i, swapCase(c))
				}
			} else {
				for _, k := range allowedKeys {
					if resolved = s.resolve(i, k); resolved {
						break
					}
				}

				if !resolved {
					return fmt.Errorf("could not find a key for %q", opt.Name)
				}
			}
		}
	}

	return nil
}

func (s *state) useAliases() bool {
	resolvedAll := true

	for i, opt := range s.unresolved {
		if opt.Alias == "" {
			resolvedAll = false
		} else if s.used[opt.Alias] != nil {
			resolvedAll = false
		} else {
			opt.Key = opt.Alias
			s.unresolved[i] = nil
			s.used[opt.Key] = opt
		}
	}

	return resolvedAll
}

func (s *state) resolve(idx int, key string) bool {
	if s.used[key] != nil {
		return false
	}

	s.unresolved[idx].Key = key
	s.used[key] = s.unresolved[idx]
	s.unresolved[idx] = nil

	return true
}

func swapCase(c string) string {
	if unicode.IsUpper(rune(c[0])) {
		return strings.ToLower(c)
	}
	return strings.ToUpper(c)
}

var allowedKeys []string

func generateAllowedKeys() (keys []string) {
	for r := 'a'; r <= 'z'; r++ {
		if r != 'q' {
			keys = append(keys, string(r))
		}
	}
	for r := 'A'; r <= 'Z'; r++ {
		keys = append(keys, string(r))
	}
	for _, k := range keys {
		keys = append(keys, "M-"+k)
	}

	for r := '0'; r <= '9'; r++ {
		keys = append(keys, string(r))
	}
	for r := '0'; r <= '9'; r++ {
		keys = append(keys, fmt.Sprintf("M-%c", r))
	}

	return
}
