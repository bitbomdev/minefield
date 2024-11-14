package custom

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
	"github.com/bitbomdev/minefield/cmd/helpers"
	apiv1 "github.com/bitbomdev/minefield/gen/api/v1"
	"github.com/bitbomdev/minefield/gen/api/v1/apiv1connect"
	"github.com/spf13/cobra"
	"google.golang.org/protobuf/types/known/emptypb"
)

func TestRenderTable(t *testing.T) {
	tests := []struct {
		name           string
		queries        []*apiv1.Query
		showInfo       bool
		maxOutput      int
		all            bool
		output         string
		expectedOutput []string
		wantErr        bool
	}{
		{
			name: "basic table without info",
			queries: []*apiv1.Query{
				{
					Node:   &apiv1.Node{Name: "test1", Type: "type1", Id: 1},
					Output: []uint32{1, 2},
				},
				{
					Node:   &apiv1.Node{Name: "test2", Type: "type2", Id: 2},
					Output: []uint32{3, 4},
				},
			},
			showInfo:  false,
			maxOutput: 10,
			all:       false,
			expectedOutput: []string{
				"NAME", "TYPE", "ID", "OUTPUT",
				"test1", "type1", "1", "2", // Output length
				"test2", "type2", "2", "2",
			},
		},
		{
			name: "table with info",
			queries: []*apiv1.Query{
				{
					Node:   &apiv1.Node{Name: "test1", Type: "type1", Id: 1},
					Output: []uint32{1},
				},
			},
			showInfo:  true,
			maxOutput: 10,
			all:       false,
			expectedOutput: []string{
				"NAME", "TYPE", "ID", "OUTPUT", "INFO",
				"test1", "type1", "1", "1", // Output length
				helpers.ComputeAdditionalInfo(&apiv1.Node{Name: "test1", Type: "type1", Id: 1}),
			},
		},
		{
			name: "maxOutput limit",
			queries: []*apiv1.Query{
				{
					Node:   &apiv1.Node{Name: "test1", Type: "type1", Id: 1},
					Output: []uint32{1},
				},
				{
					Node:   &apiv1.Node{Name: "test2", Type: "type2", Id: 2},
					Output: []uint32{2},
				},
			},
			showInfo:  false,
			maxOutput: 1,
			all:       false,
			expectedOutput: []string{
				"NAME", "TYPE", "ID", "OUTPUT",
				"test1", "type1", "1", "1",
			},
		},
		{
			name:      "no queries",
			queries:   []*apiv1.Query{},
			showInfo:  false,
			maxOutput: 10,
			all:       false,
			expectedOutput: []string{
				"No data available",
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Setup
			o := &options{}
			buf := &bytes.Buffer{}

			// Create response
			resp := connect.NewResponse(&apiv1.CustomLeaderboardResponse{
				Queries: tt.queries,
			})

			// Execute
			err := o.renderTable(buf, resp, tt.showInfo, tt.maxOutput, tt.all)

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
	validResponse := connect.NewResponse(&apiv1.CustomLeaderboardResponse{
		Queries: []*apiv1.Query{
			{
				Node:   &apiv1.Node{Name: "test1", Type: "type1", Id: 1},
				Output: []uint32{1},
			},
		},
	})

	tests := []struct {
		name      string
		w         io.Writer
		res       *connect.Response[apiv1.CustomLeaderboardResponse]
		wantError string
	}{
		{
			name:      "nil writer",
			w:         nil,
			res:       validResponse,
			wantError: "writer is nil",
		},
		{
			name:      "nil response",
			w:         validWriter,
			res:       nil,
			wantError: "queries data is invalid",
		},
		{
			name: "nil response Msg",
			w:    validWriter,
			res: &connect.Response[apiv1.CustomLeaderboardResponse]{
				Msg: nil,
			},
			wantError: "queries data is invalid",
		},
		{
			name: "nil response.Msg.Queries",
			w:    validWriter,
			res: connect.NewResponse(&apiv1.CustomLeaderboardResponse{
				Queries: nil,
			}),
			wantError: "queries data is invalid",
		},
		{
			name: "all valid",
			w:    validWriter,
			res:  validResponse,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := o.renderTable(tt.w, tt.res, false, 10, false)
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

	if cmd.Use != "custom [script]" {
		t.Errorf("Expected Use to be 'custom [script]', got '%s'", cmd.Use)
	}

	if cmd.Short == "" {
		t.Error("Short description should not be empty")
	}

	// Check if the RunE function is set
	if cmd.RunE == nil {
		t.Error("RunE function is not set")
	}

	// Verify that flags are properly added
	flags := []string{"all", "max-output", "show-info", "addr", "output"}

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
			name:         "all",
			shorthand:    "",
			defaultValue: false,
			usage:        "show the queries output for each node",
		},
		{
			name:         "max-output",
			shorthand:    "",
			defaultValue: 10,
			usage:        "max number of outputs to display",
		},
		{
			name:         "show-info",
			shorthand:    "",
			defaultValue: true,
			usage:        "display the info column",
		},
		{
			name:         "addr",
			shorthand:    "a",
			defaultValue: "http://localhost:8089",
			usage:        "Address of the Minefield server",
		},
		{
			name:         "output",
			shorthand:    "o",
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

// MockLeaderboardServiceClient is a mock implementation of the LeaderboardServiceClient interface
type MockLeaderboardServiceClient struct {
	AllKeysFunc           func(context.Context, *connect.Request[emptypb.Empty]) (*connect.Response[apiv1.AllKeysResponse], error)
	CustomLeaderboardFunc func(context.Context, *connect.Request[apiv1.CustomLeaderboardRequest]) (*connect.Response[apiv1.CustomLeaderboardResponse], error)
}

func (m *MockLeaderboardServiceClient) AllKeys(ctx context.Context, req *connect.Request[emptypb.Empty]) (*connect.Response[apiv1.AllKeysResponse], error) {
	if m.AllKeysFunc != nil {
		return m.AllKeysFunc(ctx, req)
	}
	return nil, errors.New("AllKeysFunc not implemented")
}

func (m *MockLeaderboardServiceClient) CustomLeaderboard(ctx context.Context, req *connect.Request[apiv1.CustomLeaderboardRequest]) (*connect.Response[apiv1.CustomLeaderboardResponse], error) {
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
		outputFormat  string
		setupClient   func() apiv1connect.LeaderboardServiceClient
		expectedError bool
		checkOutput   func(output string) bool
	}{
		{
			name:         "Successful run with table output",
			args:         []string{"some", "script"},
			outputFormat: "table",
			setupClient: func() apiv1connect.LeaderboardServiceClient {
				return &MockLeaderboardServiceClient{
					CustomLeaderboardFunc: func(ctx context.Context, req *connect.Request[apiv1.CustomLeaderboardRequest]) (*connect.Response[apiv1.CustomLeaderboardResponse], error) {
						queries := []*apiv1.Query{
							{
								Node: &apiv1.Node{
									Name: "Node1",
									Type: "TypeA",
									Id:   1,
								},
								Output: []uint32{1, 2},
							},
							{
								Node: &apiv1.Node{
									Name: "Node2",
									Type: "TypeB",
									Id:   2,
								},
								Output: []uint32{3, 4},
							},
						}
						res := &apiv1.CustomLeaderboardResponse{
							Queries: queries,
						}
						return connect.NewResponse(res), nil
					},
				}
			},
			expectedError: false,
			checkOutput: func(output string) bool {
				return strings.Contains(output, "Node1") && strings.Contains(output, "Node2") &&
					strings.Contains(output, "1") && strings.Contains(output, "2")
			},
		},
		{
			name:         "Successful run with JSON output",
			args:         []string{"soe", "script"},
			outputFormat: "json",
			setupClient: func() apiv1connect.LeaderboardServiceClient {
				return &MockLeaderboardServiceClient{
					CustomLeaderboardFunc: func(ctx context.Context, req *connect.Request[apiv1.CustomLeaderboardRequest]) (*connect.Response[apiv1.CustomLeaderboardResponse], error) {
						queries := []*apiv1.Query{
							{
								Node: &apiv1.Node{
									Name: "Node1",
									Type: "TypeA",
									Id:   1,
								},
								Output: []uint32{1, 2},
							},
							{
								Node: &apiv1.Node{
									Name: "Node2",
									Type: "TypeB",
									Id:   2,
								},
								Output: []uint32{3, 4},
							},
						}
						res := &apiv1.CustomLeaderboardResponse{
							Queries: queries,
						}
						return connect.NewResponse(res), nil
					},
				}
			},
			expectedError: false,
			checkOutput: func(output string) bool {
				return strings.Contains(output, "Node1") && strings.Contains(output, "Node2") &&
					strings.Contains(output, "1") && strings.Contains(output, "2")
			},
		},
		{
			name:         "Error from client",
			args:         []string{"some", "script"},
			outputFormat: "table",
			setupClient: func() apiv1connect.LeaderboardServiceClient {
				return &MockLeaderboardServiceClient{
					CustomLeaderboardFunc: func(ctx context.Context, req *connect.Request[apiv1.CustomLeaderboardRequest]) (*connect.Response[apiv1.CustomLeaderboardResponse], error) {
						return nil, errors.New("mock error")
					},
				}
			},
			expectedError: true,
		},
		{
			name:         "Invalid output format",
			args:         []string{"some", "script"},
			outputFormat: "invalid_format",
			setupClient: func() apiv1connect.LeaderboardServiceClient {
				return &MockLeaderboardServiceClient{}
			},
			expectedError: true,
		},
		{
			name:          "No script provided",
			args:          []string{},
			outputFormat:  "table",
			setupClient:   func() apiv1connect.LeaderboardServiceClient { return &MockLeaderboardServiceClient{} },
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
				output:    tt.outputFormat,
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
