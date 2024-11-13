package keys

import (
	"bytes"
	"context"
	"errors"
	"fmt"
	"io"
	"strconv"
	"strings"
	"testing"

	"connectrpc.com/connect"
	v1 "github.com/bitbomdev/minefield/gen/api/v1"
	"github.com/bitbomdev/minefield/gen/api/v1/apiv1connect"
	"github.com/spf13/cobra"
	"google.golang.org/protobuf/types/known/emptypb"
)

func TestRenderTable(t *testing.T) {
	tests := []struct {
		name           string
		nodes          []*v1.Node
		showInfo       bool
		maxOutput      int
		expectedOutput []string
		wantErr        bool
	}{
		{
			name: "basic table without info",
			nodes: []*v1.Node{
				{Name: "test1", Type: "type1", Id: 1},
				{Name: "test2", Type: "type2", Id: 2},
			},
			showInfo:  false,
			maxOutput: 10,
			expectedOutput: []string{
				"NAME", "TYPE", "ID",
				"test1", "type1", "1",
				"test2", "type2", "2",
			},
		},
		{
			name: "table with info",
			nodes: []*v1.Node{
				{Name: "test1", Type: "type1", Id: 1},
			},
			showInfo:  true,
			maxOutput: 10,
			expectedOutput: []string{
				"NAME", "TYPE", "ID", "INFO",
				"test1", "type1", "1",
			},
		},
		{
			name: "respects maxOutput",
			nodes: []*v1.Node{
				{Name: "test1", Type: "type1", Id: 1},
				{Name: "test2", Type: "type2", Id: 2},
			},
			showInfo:  false,
			maxOutput: 1,
			expectedOutput: []string{
				"NAME", "TYPE", "ID",
				"test1", "type1", "1",
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Setup
			o := &options{}
			buf := &bytes.Buffer{}

			// Create response
			resp := connect.NewResponse(&v1.AllKeysResponse{
				Nodes: tt.nodes,
			})
			// Execute
			err := o.renderTable(buf, resp, tt.showInfo, tt.maxOutput)

			// Assert
			if tt.wantErr {
				if err == nil {
					t.Errorf("expected error, got nil")
				}
				return
			}
			if err != nil {
				t.Errorf("unexpected error: %v", err)
			}

			// Check output contains expected strings
			output := buf.String()
			for _, expected := range tt.expectedOutput {
				if !strings.Contains(output, expected) {
					t.Errorf("output does not contain expected string %q", expected)
				}
			}
		})
	}
}

func TestRenderTable_NilConditions(t *testing.T) {
	o := &options{}
	validWriter := &bytes.Buffer{}
	validResponse := connect.NewResponse(&v1.AllKeysResponse{
		Nodes: []*v1.Node{
			{Name: "test1", Type: "type1", Id: 1},
		},
	})

	tests := []struct {
		name      string
		w         io.Writer
		nodes     *connect.Response[v1.AllKeysResponse]
		wantError string
	}{
		{
			name:      "nil writer",
			w:         nil,
			nodes:     validResponse,
			wantError: "writer is nil",
		},
		{
			name:      "nil nodes",
			w:         validWriter,
			nodes:     nil,
			wantError: "nodes data is invalid",
		},
		{
			name: "nil nodes.Msg",
			w:    validWriter,
			nodes: &connect.Response[v1.AllKeysResponse]{
				Msg: nil,
			},
			wantError: "nodes data is invalid",
		},
		{
			name: "nil nodes.Msg.Nodes",
			w:    validWriter,
			nodes: connect.NewResponse(&v1.AllKeysResponse{
				Nodes: nil,
			}),
			wantError: "nodes data is invalid",
		},
		{
			name:  "all valid",
			w:     validWriter,
			nodes: validResponse,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := o.renderTable(tt.w, tt.nodes, false, 10)
			if tt.wantError != "" {
				if err == nil {
					t.Errorf("expected error %q, got nil", tt.wantError)
				} else if err.Error() != tt.wantError {
					t.Errorf("expected error %q, got error %q", tt.wantError, err.Error())
				}
			} else {
				if err != nil {
					t.Errorf("unexpected error: %v", err)
				}
			}
		})
	}
}

func TestNew(t *testing.T) {
	cmd := New()

	if cmd == nil {
		t.Fatal("New() returned nil")
	}

	if cmd.Use != "allKeys" {
		t.Errorf("Expected Use to be 'allKeys', got '%s'", cmd.Use)
	}

	if cmd.Short != "Retrieve and display all leaderboard keys" {
		t.Errorf("Unexpected Short description: '%s'", cmd.Short)
	}

	if cmd.Long == "" {
		t.Error("Long description should not be empty")
	}

	// Check if the RunE function is set
	if cmd.RunE == nil {
		t.Error("RunE function is not set")
	}

	// Verify that flags are properly added
	flags := []string{"max-output", "show-info", "addr", "output"}

	for _, flag := range flags {
		if cmd.Flags().Lookup(flag) == nil {
			t.Errorf("Expected flag '%s' to be added", flag)
		}
	}
}

func TestOptions_AddFlags(t *testing.T) {
	o := &options{}
	cmd := &cobra.Command{}

	o.AddFlags(cmd)

	flags := cmd.Flags()

	tests := []struct {
		name         string
		shorthand    string
		defaultValue interface{}
		usage        string
	}{
		{
			name:         "max-output",
			shorthand:    "m",
			defaultValue: 10,
			usage:        "Specify the maximum number of keys to display",
		},
		{
			name:         "show-info",
			shorthand:    "i",
			defaultValue: true,
			usage:        "Toggle display of additional information for each key",
		},
		{
			name:         "addr",
			shorthand:    "a",
			defaultValue: "http://localhost:8089",
			usage:        "Address of the Minefield server",
		},
		{
			name:         "output",
			shorthand:    "",
			defaultValue: "table",
			usage:        "Output format (table or json)",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			f := flags.Lookup(tt.name)
			if f == nil {
				t.Errorf("Flag %q not found", tt.name)
				return
			}

			if f.Shorthand != tt.shorthand {
				t.Errorf("Expected shorthand %q for flag %q, got %q", tt.shorthand, tt.name, f.Shorthand)
			}

			if f.DefValue != toString(tt.defaultValue) {
				t.Errorf("Expected default value %q for flag %q, got %q", toString(tt.defaultValue), tt.name, f.DefValue)
			}

			if f.Usage != tt.usage {
				t.Errorf("Expected usage %q for flag %q, got %q", tt.usage, tt.name, f.Usage)
			}
		})
	}
}

func toString(value interface{}) string {
	switch v := value.(type) {
	case string:
		return v
	case int:
		return strconv.Itoa(v)
	case bool:
		return strconv.FormatBool(v)
	default:
		return fmt.Sprintf("%v", v)
	}
}

type MockLeaderboardServiceClient struct {
	AllKeysFunc           func(context.Context, *connect.Request[emptypb.Empty]) (*connect.Response[v1.AllKeysResponse], error)
	CustomLeaderboardFunc func(context.Context, *connect.Request[v1.CustomLeaderboardRequest]) (*connect.Response[v1.CustomLeaderboardResponse], error)
}

func (m *MockLeaderboardServiceClient) AllKeys(ctx context.Context, req *connect.Request[emptypb.Empty]) (*connect.Response[v1.AllKeysResponse], error) {
	if m.AllKeysFunc != nil {
		return m.AllKeysFunc(ctx, req)
	}
	return nil, errors.New("AllKeysFunc not implemented")
}

func (m *MockLeaderboardServiceClient) CustomLeaderboard(ctx context.Context, req *connect.Request[v1.CustomLeaderboardRequest]) (*connect.Response[v1.CustomLeaderboardResponse], error) {
	if m.CustomLeaderboardFunc != nil {
		return m.CustomLeaderboardFunc(ctx, req)
	}
	return nil, errors.New("CustomLeaderboardFunc not implemented")
}

// TestOptions_Run tests the Run method.
func TestOptions_Run(t *testing.T) {
	tests := []struct {
		name          string
		args          []string
		setupClient   func() apiv1connect.LeaderboardServiceClient
		expectedError bool
		checkOutput   func(output string) bool
	}{
		{
			name: "Successful run with mock data",
			args: []string{},
			setupClient: func() apiv1connect.LeaderboardServiceClient {
				return &MockLeaderboardServiceClient{
					AllKeysFunc: func(ctx context.Context, req *connect.Request[emptypb.Empty]) (*connect.Response[v1.AllKeysResponse], error) {
						nodes := []*v1.Node{
							{Name: "Node1", Type: "TypeA", Id: 1},
							{Name: "Node2", Type: "TypeB", Id: 2},
						}
						res := &v1.AllKeysResponse{
							Nodes: nodes,
						}
						return connect.NewResponse(res), nil
					},
				}
			},
			expectedError: false,
			checkOutput: func(output string) bool {
				return strings.Contains(output, "Node1") && strings.Contains(output, "Node2")
			},
		},
		{
			name: "Error from client",
			args: []string{},
			setupClient: func() apiv1connect.LeaderboardServiceClient {
				return &MockLeaderboardServiceClient{
					AllKeysFunc: func(ctx context.Context, req *connect.Request[emptypb.Empty]) (*connect.Response[v1.AllKeysResponse], error) {
						return nil, errors.New("mock error")
					},
				}
			},
			expectedError: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			o := &options{
				maxOutput: 10,
				showInfo:  true,
				addr:      "http://localhost:8089",
				client:    tt.setupClient(),
				output:    "table",
			}

			// Create a buffer to capture output
			var buf bytes.Buffer

			// Redirect the command's output to the buffer
			cmd := &cobra.Command{}
			cmd.SetOut(&buf)

			// Run the command
			err := o.Run(cmd, tt.args)

			// Check for expected error
			if (err != nil) != tt.expectedError {
				t.Fatalf("expected error: %v, got: %v", tt.expectedError, err)
			}

			// Check output if needed
			if tt.checkOutput != nil {
				output := buf.String()
				if !tt.checkOutput(output) {
					t.Errorf("output check failed, output: %s", output)
				}
			}
		})
	}
}
