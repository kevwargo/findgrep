package config

import (
	"embed"
	"encoding/json"
	"io/fs"
	"strings"
	"testing"

	"github.com/stretchr/testify/require"
	"gopkg.in/yaml.v3"
)

//go:embed all:testdata
var fixtures embed.FS

func TestLoad(t *testing.T) {
	for _, tc := range []struct {
		dir      string
		expected string
	}{
		{
			dir:      "1/2/3",
			expected: "expected-3.yml",
		},
		{
			dir:      "1/2",
			expected: "expected-2.yml",
		},
		{
			dir:      "1",
			expected: "expected-1.yml",
		},
	} {
		t.Run(tc.expected, func(tt *testing.T) {
			expectedBody, err := fixtures.ReadFile("testdata/" + tc.expected)
			require.NoError(tt, err)

			expected := new(Config)
			require.NoError(tt, yaml.Unmarshal(expectedBody, expected))

			actual, err := load("testdata/"+tc.dir, openFixtureNamed)
			require.NoError(tt, err)

			require.Equal(tt, toJson(expected), toJson(actual), "%s %s", tc.dir, tc.expected)
		})
	}
}

func toJson(cfg *Config) string {
	buf := strings.Builder{}
	enc := json.NewEncoder(&buf)
	enc.SetIndent("", "  ")
	if err := enc.Encode(cfg); err != nil {
		panic(err)
	}

	return buf.String()
}

type embeddedNamedFile struct {
	fs.File

	name string
}

func (f *embeddedNamedFile) Name() string {
	return f.name
}

func openFixtureNamed(name string) (namedFile, error) {
	f, err := fixtures.Open(name)

	return &embeddedNamedFile{
		File: f,
		name: name,
	}, err
}
