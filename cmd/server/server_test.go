package server

import (
	"testing"

	"github.com/bitbomdev/minefield/pkg/graph"
	"github.com/spf13/cobra"
	"github.com/stretchr/testify/assert"
)

func TestOptions_AddFlags(t *testing.T) {
	tests := []struct {
		name           string
		initialOptions *options
		expectedAddr   string
		expectedConc   int32
		flagValues     map[string]string
		shouldSetFlags bool
	}{
		{
			name:           "default values",
			initialOptions: &options{},
			expectedAddr:   "localhost:8089",
			expectedConc:   10,
			shouldSetFlags: false,
		},
		{
			name:           "custom values",
			initialOptions: &options{},
			expectedAddr:   "0.0.0.0:9000",
			expectedConc:   20,
			flagValues: map[string]string{
				"addr":        "0.0.0.0:9000",
				"concurrency": "20",
			},
			shouldSetFlags: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			cmd := &cobra.Command{}
			tt.initialOptions.AddFlags(cmd)

			// If we should set flags, do so
			if tt.shouldSetFlags {
				for flag, value := range tt.flagValues {
					err := cmd.Flags().Set(flag, value)
					assert.NoError(t, err)
				}
			}

			// Get the flags and verify their values
			addr, err := cmd.Flags().GetString("addr")
			assert.NoError(t, err)
			assert.Equal(t, tt.expectedAddr, addr)

			conc, err := cmd.Flags().GetInt32("concurrency")
			assert.NoError(t, err)
			assert.Equal(t, tt.expectedConc, conc)
		})
	}
}

type mockStorage struct {
	graph.Storage
}

func TestNew(t *testing.T) {
	tests := []struct {
		name    string
		storage graph.Storage
		want    struct {
			use   string
			short string
		}
	}{
		{
			name:    "creates server command with correct properties",
			storage: &mockStorage{},
			want: struct {
				use   string
				short string
			}{
				use:   "server",
				short: "Start the minefield server for graph operations and queries",
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			cmd := New()

			assert.NotNil(t, cmd)
			assert.Equal(t, tt.want.use, cmd.Use)
			assert.Equal(t, tt.want.short, cmd.Short)
			assert.True(t, cmd.DisableAutoGenTag)

			// Verify flags are added
			flags := cmd.Flags()
			concurrencyFlag := flags.Lookup("concurrency")
			assert.NotNil(t, concurrencyFlag)
			assert.Equal(t, "10", concurrencyFlag.DefValue)

			addrFlag := flags.Lookup("addr")
			assert.NotNil(t, addrFlag)
			assert.Equal(t, "localhost:8089", addrFlag.DefValue)
		})
	}
}
func TestSetupServer(t *testing.T) {
	o := &options{
		storage:     &mockStorage{},
		concurrency: 10,
		addr:        "localhost:8089",
	}

	srv, err := o.setupServer()
	if err != nil {
		t.Fatalf("Expected no error, got %v", err)
	}

	if srv.Addr != "localhost:8089" {
		t.Errorf("Expected address 'localhost:8089', got '%s'", srv.Addr)
	}

	if srv.Handler == nil {
		t.Error("Expected handler to be set, got nil")
	}
}
