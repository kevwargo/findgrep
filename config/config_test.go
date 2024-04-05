package config

import (
	"embed"
	"io/fs"
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
		expectedBody, err := fixtures.ReadFile("testdata/" + tc.expected)
		require.NoError(t, err)

		expected := new(Config)
		require.NoError(t, yaml.Unmarshal(expectedBody, expected))

		actual, err := load("testdata/"+tc.dir, openFixtureNamed)
		require.NoError(t, err)

		require.Equal(t, expected, actual, "%s %s", tc.dir, tc.expected)
	}
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
