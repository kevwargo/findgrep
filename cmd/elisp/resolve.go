package elisp

import (
	"fmt"
	"strings"
	"unicode"

	"github.com/kevwargo/findgrep/config"
)

func resolveKeys(optionGroups ...config.OptionGroup) error {
	s, err := initState(optionGroups)
	if err != nil {
		return err
	}

	return s.resolveAll()
}

type state struct {
	used            map[string]*config.Option
	unresolved      []*config.Option
	unresolvedCount int
	keyPool         []string
	allowed         map[string]bool
}

func initState(optionGroups []config.OptionGroup) (*state, error) {
	used := make(map[string]*config.Option)
	var unresolved []*config.Option

	for _, group := range optionGroups {
		for _, opt := range group.All() {
			if k := opt.Key; k != "" {
				if other := used[k]; other != nil {
					return nil, fmt.Errorf("key %q used by both %q and %q", k, opt.Flag().Name, other.Flag().Name)
				}
				used[k] = opt
			} else {
				unresolved = append(unresolved, opt)
			}
		}
	}

	keyPool := generateKeyPool()
	allowed := make(map[string]bool, len(keyPool))
	for _, k := range keyPool {
		allowed[k] = used[k] == nil
	}

	return &state{
		used:            used,
		unresolved:      unresolved,
		unresolvedCount: len(unresolved),
		keyPool:         keyPool,
		allowed:         allowed,
	}, nil
}

func (s *state) resolveAll() error {
	s.useAliases()

	for pos := 0; s.unresolvedCount > 0; pos++ {
		for i, opt := range s.unresolved {
			if opt == nil {
				continue
			}

			if !s.resolveFromName(i, pos) && !s.resolveFromPool(i) {
				return fmt.Errorf("could not find a key for %q", opt.Name)
			}
		}
	}

	return nil
}

func (s *state) resolveFromName(idx int, pos int) bool {
	if pos >= len(s.unresolved[idx].Name) {
		return false
	}

	if k := string(s.unresolved[idx].Name[pos]); s.allowed[k] {
		return s.resolve(idx, k) || s.resolve(idx, swapCase(k))
	}

	return false
}

func (s *state) resolveFromPool(idx int) bool {
	for _, k := range s.keyPool {
		if s.resolve(idx, k) {
			return true
		}
	}

	return false
}

func (s *state) useAliases() {
	for i, opt := range s.unresolved {
		if opt.Alias != "" {
			s.resolve(i, opt.Alias)
		}
	}
}

func (s *state) resolve(idx int, key string) bool {
	if s.used[key] != nil {
		return false
	}

	s.unresolved[idx].Key = key
	s.used[key] = s.unresolved[idx]
	s.unresolved[idx] = nil
	s.unresolvedCount--

	return true
}

func swapCase(c string) string {
	if unicode.IsUpper(rune(c[0])) {
		return strings.ToLower(c)
	}
	return strings.ToUpper(c)
}

func generateKeyPool() (keys []string) {
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
