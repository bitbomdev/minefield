package custom

import (
	"bytes"
	"context"
	"errors"
	"strconv"
	"strings"
	"testing"

	"connectrpc.com/connect"
	apiv1 "github.com/bitbomdev/minefield/gen/api/v1"
	"github.com/spf13/cobra"
	"github.com/stretchr/testify/assert"
)

func TestOptions_AddFlags(t *testing.T) {
	o := &options{}
	cmd := &cobra.Command{}

	o.AddFlags(cmd)

	// Test "max-output" flag
	maxOutputFlag := cmd.Flags().Lookup("max-output")
	if maxOutputFlag == nil {
		t.Error("Expected 'max-output' flag to be defined")
	} else {
		if maxOutputFlag.DefValue != "10" {
			t.Errorf("Expected default value of 'max-output' to be '10', got '%s'", maxOutputFlag.DefValue)
		}
	}

	// Test "show-info" flag
	showInfoFlag := cmd.Flags().Lookup("show-info")
	if showInfoFlag == nil {
		t.Error("Expected 'show-info' flag to be defined")
	} else {
		if showInfoFlag.DefValue != "true" {
			t.Errorf("Expected default value of 'show-info' to be 'true', got '%s'", showInfoFlag.DefValue)
		}
	}

	// Test "addr" flag
	addrFlag := cmd.Flags().Lookup("addr")
	if addrFlag == nil {
		t.Error("Expected 'addr' flag to be defined")
	} else {
		if addrFlag.DefValue != "http://localhost:8089" {
			t.Errorf("Expected default value of 'addr' to be 'http://localhost:8089', got '%s'", addrFlag.DefValue)
		}
	}

	// Test "output" flag
	outputFlag := cmd.Flags().Lookup("output")
	if outputFlag == nil {
		t.Error("Expected 'output' flag to be defined")
	} else {
		if outputFlag.DefValue != "table" {
			t.Errorf("Expected default value of 'output' to be 'table', got '%s'", outputFlag.DefValue)
		}
	}
}

func TestNewCommand(t *testing.T) {
	cmd := New()

	if cmd == nil {
		t.Fatal("Expected New() to return a non-nil command")
	}

	// Check the command's basic properties
	if cmd.Use != "custom [script]" {
		t.Errorf("Expected Use: 'custom [script]', got: '%s'", cmd.Use)
	}

	if cmd.Short != "Execute a custom query script" {
		t.Errorf("Expected Short: 'Execute a custom query script', got: '%s'", cmd.Short)
	}

	expectedLong := "Execute a custom query script to perform tailored queries against the project's dependencies and dependents."
	if cmd.Long != expectedLong {
		t.Errorf("Expected Long: '%s', got: '%s'", expectedLong, cmd.Long)
	}

	if cmd.Args == nil {
		t.Error("Expected Args to be defined")
	}

	if cmd.RunE == nil {
		t.Error("Expected RunE to be defined")
	}

	// Verify that the flags are added to the command
	flags := []struct {
		name      string
		shorthand string
		defValue  string
	}{
		{name: "max-output", defValue: "10"},
		{name: "show-info", defValue: "true"},
		{name: "addr", defValue: "http://localhost:8089"},
		{name: "output", defValue: "table"},
	}

	for _, flag := range flags {
		f := cmd.Flags().Lookup(flag.name)
		if f == nil {
			t.Errorf("Expected flag '%s' to be defined", flag.name)
		} else {
			if f.DefValue != flag.defValue {
				t.Errorf("Expected default value of flag '%s' to be '%s', got '%s'", flag.name, flag.defValue, f.DefValue)
			}
		}
	}
}

// TestFormatTable tests the formatTable function
func TestFormatTable(t *testing.T) {
	nodes := []*apiv1.Node{
		{
			Name: "Node1",
			Type: "TypeA",
			Id:   1,
		},
		{
			Name: "Node2",
			Type: "TypeB",
			Id:   2,
		},
		{
			Name: "Node3",
			Type: "TypeC",
			Id:   3,
		},
	}

	t.Run("WithShowInfoTrue", func(t *testing.T) {
		var buf bytes.Buffer
		err := formatTable(&buf, nodes, 10, true)
		assert.NoError(t, err)

		output := buf.String()

		// Check if the output contains expected headers
		assert.Contains(t, output, "NAME")
		assert.Contains(t, output, "TYPE")
		assert.Contains(t, output, "ID")
		assert.Contains(t, output, "INFO")

		// Check if the output contains expected node data
		for _, node := range nodes {
			assert.Contains(t, output, node.Name)
			assert.Contains(t, output, node.Type)
			assert.Contains(t, output, strconv.FormatUint(uint64(node.Id), 10))
		}
	})

	t.Run("MaxOutputLimit", func(t *testing.T) {
		var buf bytes.Buffer
		err := formatTable(&buf, nodes, 2, true)
		assert.NoError(t, err)

		output := buf.String()

		// Only the first two nodes should be present
		assert.Contains(t, output, "Node1")
		assert.Contains(t, output, "Node2")
		assert.NotContains(t, output, "Node3")
	})
}

// Mock implementation of QueryServiceClient
type mockQueryServiceClient struct {
	QueryFunc func(ctx context.Context, req *connect.Request[apiv1.QueryRequest]) (*connect.Response[apiv1.QueryResponse], error)
}

func (m *mockQueryServiceClient) Query(ctx context.Context, req *connect.Request[apiv1.QueryRequest]) (*connect.Response[apiv1.QueryResponse], error) {
	return m.QueryFunc(ctx, req)
}

func TestRun(t *testing.T) {
	tests := []struct {
		name                string
		script              string
		output              string
		maxOutput           int
		showInfo            bool
		mockResponse        *apiv1.QueryResponse
		mockError           error
		expectError         bool
		expectedErrorString string
		expectedOutput      []string
	}{
		{
			name:      "valid response with nodes in json output",
			script:    "match (n) return n",
			output:    "json",
			maxOutput: 10,
			showInfo:  true,
			mockResponse: &apiv1.QueryResponse{
				Nodes: []*apiv1.Node{
					{Name: "node1", Type: "type1", Id: 1},
					{Name: "node2", Type: "type2", Id: 2},
				},
			},
			expectedOutput: []string{
				`"name": "node1"`, `"type": "type1"`, `"id": "1"`,
				`"name": "node2"`, `"type": "type2"`, `"id": "2"`,
			},
		},
		{
			name:      "valid response with nodes in table output",
			script:    "match (n) return n",
			output:    "table",
			maxOutput: 10,
			showInfo:  true,
			mockResponse: &apiv1.QueryResponse{
				Nodes: []*apiv1.Node{
					{Name: "node1", Type: "type1", Id: 1},
					{Name: "node2", Type: "type2", Id: 2},
				},
			},
			expectedOutput: []string{
				"NAME", "TYPE", "ID", "INFO",
				"node1", "type1", "1",
				"node2", "type2", "2",
			},
		},
		{
			name:                "no nodes found",
			script:              "match (n) where n.name = 'unknown' return n",
			output:              "table",
			maxOutput:           10,
			showInfo:            true,
			mockResponse:        &apiv1.QueryResponse{Nodes: []*apiv1.Node{}},
			expectError:         true,
			expectedErrorString: "no nodes found for script: match (n) where n.name = 'unknown' return n",
		},
		{
			name:                "client error",
			script:              "bad script",
			output:              "json",
			maxOutput:           10,
			showInfo:            true,
			mockError:           errors.New("client error"),
			expectError:         true,
			expectedErrorString: "query failed: client error",
		},
		{
			name:                "unknown output format",
			script:              "match (n) return n",
			output:              "unknown",
			maxOutput:           10,
			showInfo:            true,
			mockResponse:        &apiv1.QueryResponse{Nodes: []*apiv1.Node{{Name: "node1", Type: "type1", Id: 1}}},
			expectError:         true,
			expectedErrorString: "unknown output format: unknown",
		},
		{
			name:                "empty script",
			script:              "   ",
			output:              "json",
			maxOutput:           10,
			showInfo:            true,
			expectError:         true,
			expectedErrorString: "script cannot be empty",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Mock QueryServiceClient
			mockClient := &mockQueryServiceClient{
				QueryFunc: func(ctx context.Context, req *connect.Request[apiv1.QueryRequest]) (*connect.Response[apiv1.QueryResponse], error) {
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
				showInfo:           tt.showInfo,
				queryServiceClient: mockClient,
			}

			// Create a cobra command and context
			cmd := &cobra.Command{}
			cmd.SetOut(new(bytes.Buffer)) // Capture output for assertions
			outputBuf := &bytes.Buffer{}
			cmd.SetOut(outputBuf)
			cmd.SetContext(context.Background())

			args := []string{tt.script}

			err := o.Run(cmd, args)

			if tt.expectError {
				assert.Error(t, err)
				assert.Equal(t, tt.expectedErrorString, err.Error())
			} else {
				assert.NoError(t, err)
				outputStr := outputBuf.String()
				for _, expected := range tt.expectedOutput {
					if !strings.Contains(outputStr, expected) {
						t.Errorf("Output does not contain expected string: %s", expected)
					}
				}
			}
		})
	}
}
