package osv

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/spf13/pflag"
)

func TestNew(t *testing.T) {
	tests := []struct {
		name          string
		wantUse       string
		wantShort     string
		wantFlagCount int
	}{
		{
			name:          "creates command with correct configuration",
			wantUse:       "osv [path to vulnerability file/dir]",
			wantShort:     "Graph vulnerability data into the graph, and connect it to existing library nodes",
			wantFlagCount: 1, // Should have the "addr" flag
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			cmd := New()

			// Test basic command properties
			assert.Equal(t, tt.wantUse, cmd.Use)
			assert.Equal(t, tt.wantShort, cmd.Short)
			assert.True(t, cmd.DisableAutoGenTag)
			assert.NotNil(t, cmd.RunE)

			// Test flags
			flags := cmd.Flags()

			// Count the number of defined flags
			flagCount := 0
			flags.VisitAll(func(*pflag.Flag) { flagCount++ })
			assert.Equal(t, tt.wantFlagCount, flagCount)

			// Test addr flag specifically
			addrFlag := flags.Lookup("addr")
			assert.NotNil(t, addrFlag)
			assert.Equal(t, "string", addrFlag.Value.Type())
			assert.Equal(t, DefaultAddr, addrFlag.DefValue)
		})
	}
}
