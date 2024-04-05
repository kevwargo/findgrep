package config

import (
	"errors"
	"io/fs"
	"os"
	"path/filepath"

	"gopkg.in/yaml.v3"
)

type Config struct {
	ExcludePaths Options `yaml:"exclude-paths"`
	IgnoreFiles  Options `yaml:"ignore-files"`
	SelectFiles  Options `yaml:"select-files"`
	Grep         Options `yaml:"grep"`
}

const FileName = ".findgrep.yml"

func Load(dir string) (*Config, error) {
	path, err := filepath.Abs(dir)
	if err != nil {
		return nil, err
	}

	return load(path, func(n string) (fs.File, error) { return os.Open(n) })
}

func load(path string, openFn func(string) (fs.File, error)) (*Config, error) {
	cfg, err := loadDefault()
	if err != nil {
		return nil, err
	}

	var files []fs.File
	defer func() {
		for _, f := range files {
			f.Close()
		}
	}()

	for ; path != filepath.Dir(path); path = filepath.Dir(path) {
		f, err := openFn(filepath.Join(path, FileName))
		if err != nil {
			if errors.Is(err, fs.ErrNotExist) {
				continue
			}
			return nil, err
		}

		files = append(files, f)
	}

	for i := len(files) - 1; i >= 0; i-- {
		if err := yaml.NewDecoder(files[i]).Decode(cfg); err != nil {
			return nil, err
		}
	}

	return cfg, nil
}
