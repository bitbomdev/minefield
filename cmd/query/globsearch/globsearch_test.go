package globsearch

import (
	"bytes"
	"context"
	"errors"
	"fmt"
	"io"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"connectrpc.com/connect"
	apiv1 "github.com/bit-bom/minefield/gen/api/v1"
	"github.com/spf13/cobra"
	"github.com/zeebo/assert"
)

func TestFormatTable(t *testing.T) {
	tests := []struct {
		name      string
		nodes     []*apiv1.Node
		maxOutput int
		want      []string // Strings that should appear in the output
	}{
		{
			name: "basic output",
			nodes: []*apiv1.Node{
				{Name: "node1", Type: "type1", Id: 1},
				{Name: "node2", Type: "type2", Id: 2},
			},
			maxOutput: 10,
			want: []string{
				"node1", "type1", "1",
				"node2", "type2", "2",
			},
		},
		{
			name: "respects maxOutput",
			nodes: []*apiv1.Node{
				{Name: "node1", Type: "type1", Id: 1},
				{Name: "node2", Type: "type2", Id: 2},
				{Name: "node3", Type: "type3", Id: 3},
			},
			maxOutput: 2,
			want: []string{
				"node1", "type1", "1",
				"node2", "type2", "2",
			},
		},
		{
			name:      "empty nodes",
			nodes:     []*apiv1.Node{},
			maxOutput: 10,
			want:      []string{"NAME", "TYPE", "ID"}, // Should only contain headers
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			buf := &bytes.Buffer{}
			err := formatTable(buf, tt.nodes, tt.maxOutput)
			if err != nil {
				t.Errorf("formatTable() error = %v", err)
				return
			}

			got := buf.String()
			for _, want := range tt.want {
				if !strings.Contains(got, want) {
					t.Errorf("formatTable() output doesn't contain %q\nGot:\n%s", want, got)
				}
			}
		})
	}
}

func TestNew_Flags(t *testing.T) {
	// Start test server
	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// Mock response if needed
		w.WriteHeader(http.StatusOK)
	}))
	defer ts.Close()

	tests := []struct {
		name    string
		args    []string
		checkFn func(*cobra.Command) error
	}{
		{
			name: "default flags",
			args: []string{"pattern"},
			checkFn: func(cmd *cobra.Command) error {
				maxOutput, _ := cmd.Flags().GetInt("max-output")
				if maxOutput != 10 {
					return fmt.Errorf("max-output = %v, want %v", maxOutput, 10)
				}

				addr, _ := cmd.Flags().GetString("addr")
				if addr != "http://localhost:8089" {
					return fmt.Errorf("addr = %v, want %v", addr, "http://localhost:8089")
				}

				output, _ := cmd.Flags().GetString("output")
				if output != "table" {
					return fmt.Errorf("output = %v, want %v", output, "table")
				}

				return nil
			},
		},
		{
			name: "custom flags",
			args: []string{
				"pattern",
				"--max-output=20",
				fmt.Sprintf("--addr=%s", ts.URL),
				"--output=json",
			},
			checkFn: func(cmd *cobra.Command) error {
				maxOutput, _ := cmd.Flags().GetInt("max-output")
				if maxOutput != 20 {
					return fmt.Errorf("max-output = %v, want %v", maxOutput, 20)
				}

				addr, _ := cmd.Flags().GetString("addr")
				if addr != ts.URL {
					return fmt.Errorf("addr = %v, want %v", addr, ts.URL)
				}

				output, _ := cmd.Flags().GetString("output")
				if output != "json" {
					return fmt.Errorf("output = %v, want %v", output, "json")
				}

				return nil
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			cmd := New()
			cmd.SetArgs(tt.args)

			// Parse flags
			if err := cmd.ParseFlags(tt.args); err != nil {
				t.Errorf("cmd.ParseFlags() error = %v", err)
				return
			}

			// Check flag values
			if err := tt.checkFn(cmd); err != nil {
				t.Error(err)
			}
		})
	}
}

type mockGraphServiceClient struct {
	GetNodesByGlobFunc func(ctx context.Context, req *connect.Request[apiv1.GetNodesByGlobRequest]) (*connect.Response[apiv1.GetNodesByGlobResponse], error)
	GetNodeFunc        func(ctx context.Context, req *connect.Request[apiv1.GetNodeRequest]) (*connect.Response[apiv1.GetNodeResponse], error)
	GetNodeByNameFunc  func(ctx context.Context, req *connect.Request[apiv1.GetNodeByNameRequest]) (*connect.Response[apiv1.GetNodeByNameResponse], error)
}

func (m *mockGraphServiceClient) GetNodesByGlob(ctx context.Context, req *connect.Request[apiv1.GetNodesByGlobRequest]) (*connect.Response[apiv1.GetNodesByGlobResponse], error) {
	return m.GetNodesByGlobFunc(ctx, req)
}

func (m *mockGraphServiceClient) GetNode(ctx context.Context, req *connect.Request[apiv1.GetNodeRequest]) (*connect.Response[apiv1.GetNodeResponse], error) {
	return m.GetNodeFunc(ctx, req)
}

func (m *mockGraphServiceClient) GetNodeByName(ctx context.Context, req *connect.Request[apiv1.GetNodeByNameRequest]) (*connect.Response[apiv1.GetNodeByNameResponse], error) {
	return m.GetNodeByNameFunc(ctx, req)
}

func TestRun(t *testing.T) {
	tests := []struct {
		name                string
		pattern             string
		output              string
		maxOutput           int
		mockResponse        *apiv1.GetNodesByGlobResponse
		mockError           error
		expectError         bool
		expectedErrorString string
	}{
		{
			name:      "valid response with nodes",
			pattern:   "node*",
			output:    "json",
			maxOutput: 10,
			mockResponse: &apiv1.GetNodesByGlobResponse{
				Nodes: []*apiv1.Node{
					{Name: "node1", Type: "type1", Id: 1},
					{Name: "node2", Type: "type2", Id: 2},
				},
			},
		},
		{
			name:      "metadata",
			pattern:   "node*",
			output:    "json",
			maxOutput: 10,
			mockResponse: &apiv1.GetNodesByGlobResponse{
				Nodes: []*apiv1.Node{
					{Name: "node1", Type: "type1", Id: 1, Metadata: []byte(`{"field": "value1"}`)},
				},
			},
		},
		{
			name:                "no nodes found",
			pattern:             "unknown*",
			output:              "table",
			maxOutput:           10,
			mockResponse:        &apiv1.GetNodesByGlobResponse{Nodes: []*apiv1.Node{}},
			expectError:         true,
			expectedErrorString: "no nodes found matching pattern: unknown*",
		},
		{
			name:                "client error",
			pattern:             "error*",
			output:              "json",
			maxOutput:           10,
			mockError:           errors.New("client error"),
			expectError:         true,
			expectedErrorString: "query failed: client error",
		},
		{
			name:                "unknown output format",
			pattern:             "node*",
			output:              "unknown",
			maxOutput:           10,
			mockResponse:        &apiv1.GetNodesByGlobResponse{Nodes: []*apiv1.Node{{Name: "node1", Type: "type1", Id: 1}}},
			expectError:         true,
			expectedErrorString: "unknown output format: unknown",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mockClient := &mockGraphServiceClient{
				GetNodesByGlobFunc: func(ctx context.Context, req *connect.Request[apiv1.GetNodesByGlobRequest]) (*connect.Response[apiv1.GetNodesByGlobResponse], error) {
					if tt.mockError != nil {
						return nil, tt.mockError
					}
					return connect.NewResponse(tt.mockResponse), nil
				},
			}

			o := &options{
				addr:               "http://localhost:8089",
				output:             tt.output,
				maxOutput:          tt.maxOutput,
				graphServiceClient: mockClient,
			}

			// Create a cobra command and context
			cmd := &cobra.Command{}
			cmd.SetOut(io.Discard) // Discard output during testing
			cmd.SetContext(context.Background())

			args := []string{tt.pattern}

			err := o.Run(cmd, args)

			if tt.expectError {
				assert.Error(t, err)
				assert.Equal(t, tt.expectedErrorString, err.Error())
			} else {
				assert.NoError(t, err)
			}
		})
	}
}
