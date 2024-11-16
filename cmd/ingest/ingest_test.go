package ingest

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestNewCommand(t *testing.T) {
	cmd := New()
	assert.NotNil(t, cmd, "Ingest command should not be nil")

	assert.Equal(t, "ingest", cmd.Use, "Command 'Use' should be 'ingest'")
	assert.Equal(t, "ingest metadata into the graph", cmd.Short, "Command 'Short' description should match")

	subcommands := cmd.Commands()
	subcommandUses := []string{}
	for _, subcmd := range subcommands {
		subcommandUses = append(subcommandUses, subcmd.Use)
	}

	expectedSubcommands := []string{
		"osv [path to vulnerability file/dir]",
		"sbom [path to sbom file/dir]",
		"scorecard [path to scorecard file/dir]",
	}
	assert.ElementsMatch(t, expectedSubcommands, subcommandUses, "Subcommands should match expected list")
}

func TestWireDependencyResolution(t *testing.T) {
	cmd := New()
	assert.NotNil(t, cmd, "Ingest command should initialize without error")
}
