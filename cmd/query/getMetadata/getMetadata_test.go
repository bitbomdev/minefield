package getMetadata

import (
	"bytes"
	"context"
	"errors"
	"os"
	"strings"
	"testing"

	"connectrpc.com/connect"
	apiv1 "github.com/bitbomdev/minefield/gen/api/v1"
	"github.com/spf13/cobra"
	"github.com/zeebo/assert"
)

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
