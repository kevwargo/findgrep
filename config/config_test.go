package config

import (
	"embed"
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

		actual, err := load("testdata/"+tc.dir, fixtures.Open)
		require.NoError(t, err)

		require.Equal(t, expected, actual, "%s %s", tc.dir, tc.expected)
	}
}
