package config

import (
	"errors"
	"fmt"
	"io"
	"io/fs"
	"os"
	"path/filepath"

	"gopkg.in/yaml.v3"
)

type Config struct {
	ExcludePaths Options     `yaml:"exclude-paths"`
	IgnoreFiles  Options     `yaml:"ignore-files"`
	SelectFiles  Options     `yaml:"select-files"`
	Grep         Options     `yaml:"grep"`
	Misc         MiscOptions `yaml:"misc"`
}

type MiscOptions struct {
	Gzip    *Option `yaml:"gzip"`
	Verbose *Option `yaml:"verbose"`
}

const FileName = ".findgrep.yml"

func Load(dir string) (*Config, error) {
	path, err := filepath.Abs(dir)
	if err != nil {
		return nil, err
	}

	return load(path, func(n string) (namedFile, error) { return os.Open(n) })
}

type namedFile interface {
	fs.File
	Name() string
}

func load(path string, openFn func(string) (namedFile, error)) (*Config, error) {
	cfg, err := loadDefault()
	if err != nil {
		return nil, err
	}

	var files []namedFile
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
			if errors.Is(err, io.EOF) {
				continue
			}

			return nil, fmt.Errorf("parsing config file %q: %w", files[i].Name(), err)
		}
	}

	return cfg, nil
}
