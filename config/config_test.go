package config_test

import (
	"embed"
	"testing"

	"github.com/stretchr/testify/assert"
	"gopkg.in/yaml.v3"

	"github.com/kevwargo/findgrep/config"
)

//go:embed testdata/*.yml
var fixtures embed.FS

func TestMerge(t *testing.T) {
	var merged []byte
	var sources [][]byte

	entries, err := fixtures.ReadDir("testdata")
	if err != nil {
		panic(err)
	}

	for _, entry := range entries {
		contents, err := fixtures.ReadFile("testdata/" + entry.Name())
		assert.NoError(t, err, "reading fixture", entry.Name())

		if entry.Name() == "merged.yml" {
			merged = contents
		} else {
			sources = append(sources, contents)
		}
	}

	if merged == nil {
		t.Fatal("merged config fixture not found")
	}

	var expected, actual config.Config
	assert.NoError(t, yaml.Unmarshal(merged, &expected))
	for _, src := range sources {
		assert.NoError(t, yaml.Unmarshal(src, &actual))
	}

	assert.Equal(t, expected, actual)
}
