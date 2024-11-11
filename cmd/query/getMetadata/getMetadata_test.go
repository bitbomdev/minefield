package getMetadata

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"os"
	"strings"
	"testing"

	"connectrpc.com/connect"
	apiv1 "github.com/bitbomdev/minefield/gen/api/v1"
	"github.com/google/go-cmp/cmp"
	"github.com/spf13/cobra"
	"github.com/zeebo/assert"
)

func TestFormatNodeJSON(t *testing.T) {
	tests := []struct {
		name       string
		node       *apiv1.Node
		want       string
		wantError  bool
		wantErrMsg string
	}{
		{
			name: "valid node with metadata",
			node: &apiv1.Node{
				Name:     "node1",
				Type:     "type1",
				Id:       1,
				Metadata: []byte(`{"key": "value", "nested": {"foo": "bar"}}`),
			},
			want: `{
  "name": "node1",
  "type": "type1",
  "id": "1",
  "metadata": {
    "key": "value",
    "nested": {
      "foo": "bar"
    }
  }
}`,
		},
		{
			name: "valid node with empty metadata",
			node: &apiv1.Node{
				Name:     "node2",
				Type:     "type2",
				Id:       2,
				Metadata: []byte(`{}`),
			},
			want: `{
  "name": "node2",
  "type": "type2",
  "id": "2"
}`,
		},
		{
			name: "valid node with null metadata",
			node: &apiv1.Node{
				Name:     "node3",
				Type:     "type3",
				Id:       3,
				Metadata: nil,
			},
			want: `{
  "name": "node3",
  "type": "type3",
  "id": "3"
}`,
		},
		{
			name:       "nil node",
			node:       nil,
			wantError:  true,
			wantErrMsg: "node cannot be nil",
		},
		{
			name: "invalid metadata json",
			node: &apiv1.Node{
				Name:     "node4",
				Type:     "type4",
				Id:       4,
				Metadata: []byte(`{invalid`),
			},
			wantError:  true,
			wantErrMsg: "failed to unmarshal metadata for node node4",
		},
		{
			name: "metadata with special characters",
			node: &apiv1.Node{
				Name:     "node5",
				Type:     "type5",
				Id:       5,
				Metadata: []byte(`{"special": "!@#$%^&*()\n\t"}`),
			},
			want: `{
  "name": "node5",
  "type": "type5",
  "id": "5",
  "metadata": {
    "special": "!@#$%^&*()\n\t"
  }
}`,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := formatNodeJSON(tt.node)
			if tt.wantError {
				if err == nil {
					t.Error("expected error but got none")
					return
				}
				if !strings.Contains(err.Error(), tt.wantErrMsg) {
					t.Errorf("error message %q does not contain %q", err.Error(), tt.wantErrMsg)
				}
				return
			}

			if err != nil {
				t.Errorf("unexpected error: %v", err)
				return
			}

			// Verify the output is valid JSON
			var gotJSON, wantJSON interface{}
			if err := json.Unmarshal(got, &gotJSON); err != nil {
				t.Errorf("failed to unmarshal got JSON: %v", err)
				return
			}
			if err := json.Unmarshal([]byte(tt.want), &wantJSON); err != nil {
				t.Errorf("failed to unmarshal want JSON: %v", err)
				return
			}

			if diff := cmp.Diff(wantJSON, gotJSON); diff != "" {
				t.Errorf("formatNodeJSON() mismatch (-want +got):\n%s", diff)
			}
		})
	}
}

func TestFormatTable(t *testing.T) {
	node := &apiv1.Node{
		Name: "node1",
		Type: "type1",
		Id:   1,
	}

	buf := &bytes.Buffer{}
	err := formatTable(buf, node)
	if err != nil {
		t.Errorf("formatTable() error = %v", err)
		return
	}

	// Check for both headers and values in the expected order
	expectedRows := [][]string{
		{"NAME", "TYPE", "ID"},  // Headers should be uppercase
		{"node1", "type1", "1"}, // Values
	}

	output := buf.String()
	for _, row := range expectedRows {
		for _, str := range row {
			if !strings.Contains(output, str) {
				t.Errorf("formatTable() output does not contain %q\nGot output:\n%s", str, output)
			}
		}
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
		nodeName            string
		output              string
		outputFile          string
		mockResponse        *apiv1.GetNodeByNameResponse
		mockError           error
		expectError         bool
		expectedErrorString string
	}{
		{
			name:     "valid node with JSON output",
			nodeName: "node1",
			output:   "json",
			mockResponse: &apiv1.GetNodeByNameResponse{
				Node: &apiv1.Node{
					Name:     "node1",
					Type:     "type1",
					Id:       1,
					Metadata: []byte(`{"key": "value"}`),
				},
			},
		},
		{
			name:     "valid node with table output",
			nodeName: "node1",
			output:   "table",
			mockResponse: &apiv1.GetNodeByNameResponse{
				Node: &apiv1.Node{
					Name: "node1",
					Type: "type1",
					Id:   1,
				},
			},
		},
		{
			name:                "node not found",
			nodeName:            "unknown",
			output:              "json",
			mockResponse:        &apiv1.GetNodeByNameResponse{Node: nil},
			expectError:         true,
			expectedErrorString: "node not found: unknown",
		},
		{
			name:                "client error",
			nodeName:            "error",
			output:              "json",
			mockError:           errors.New("client error"),
			expectError:         true,
			expectedErrorString: "query failed: client error",
		},
		{
			name:                "unknown output format",
			nodeName:            "node1",
			output:              "unknown",
			mockResponse:        &apiv1.GetNodeByNameResponse{Node: &apiv1.Node{Name: "node1", Type: "type1", Id: 1}},
			expectError:         true,
			expectedErrorString: "unknown output format: unknown",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mockClient := &mockGraphServiceClient{
				GetNodeByNameFunc: func(ctx context.Context, req *connect.Request[apiv1.GetNodeByNameRequest]) (*connect.Response[apiv1.GetNodeByNameResponse], error) {
					if tt.mockError != nil {
						return nil, tt.mockError
					}
					return connect.NewResponse(tt.mockResponse), nil
				},
			}

			o := &options{
				addr:               "http://localhost:8089",
				output:             tt.output,
				outputFile:         tt.outputFile,
				graphServiceClient: mockClient,
			}

			// Create a cobra command and context
			cmd := &cobra.Command{}
			var outputBuf bytes.Buffer
			cmd.SetOut(&outputBuf)
			cmd.SetErr(&outputBuf)
			cmd.SetContext(context.Background())

			args := []string{tt.nodeName}

			err := o.Run(cmd, args)

			if tt.expectError {
				assert.Error(t, err)
				assert.Equal(t, tt.expectedErrorString, err.Error())
			} else {
				assert.NoError(t, err)
				// If outputFile is set, check if the file was created
				if tt.outputFile != "" {
					_, err := os.Stat(tt.outputFile)
					assert.NoError(t, err)
					// Cleanup test file
					os.Remove(tt.outputFile)
				}
				// Additional assertions can be added here to check the output
			}
		})
	}
}

func TestAddFlags(t *testing.T) {
	o := &options{}
	cmd := &cobra.Command{}
	o.AddFlags(cmd)

	flags := []string{"output-file", "addr", "output"}

	for _, flag := range flags {
		if cmd.Flags().Lookup(flag) == nil {
			t.Errorf("Flag %q not found", flag)
		}
	}
}

func TestNew(t *testing.T) {
	// Create the command
	cmd := New()

	// Test basic command properties
	assert.Equal(t, "get-metadata", cmd.Name())
	assert.Equal(t, "get-metadata [node name]", cmd.Use)
	assert.Equal(t, "Outputs the node with its metadata", cmd.Short)

	// Test that flags were added
	f := cmd.Flags()
	assert.NotNil(t, f.Lookup("output"))
	assert.NotNil(t, f.Lookup("addr"))
	assert.NotNil(t, f.Lookup("output-file"))
}
