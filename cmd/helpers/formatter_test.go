package helpers

import (
	"bytes"
	"encoding/json"
	"strings"
	"testing"

	v1 "github.com/bitbomdev/minefield/gen/api/v1"
)

// TestFormatNodeJSON tests the FormatNodeJSON function for 100% code coverage.
func TestFormatNodeJSON(t *testing.T) {
	tests := []struct {
		name        string
		nodes       []*v1.Node
		expectErr   bool
		expectedErr string
		expected    string
	}{
		{
			name:        "Nil nodes slice",
			nodes:       nil,
			expectErr:   true,
			expectedErr: "nodes cannot be nil",
		},
		{
			name:        "Empty nodes slice",
			nodes:       []*v1.Node{},
			expectErr:   true,
			expectedErr: "no nodes found",
		},
		{
			name: "Valid node without metadata",
			nodes: []*v1.Node{
				{
					Name:     "Node1",
					Type:     "TypeA",
					Id:       1,
					Metadata: nil,
				},
			},
			expectErr: false,
			expected: `[
  {
    "name": "Node1",
    "type": "TypeA",
    "id": "1"
  }
]`,
		},
		{
			name: "Valid node with valid metadata",
			nodes: []*v1.Node{
				{
					Name:     "Node2",
					Type:     "TypeB",
					Id:       2,
					Metadata: []byte(`{"key":"value"}`),
				},
			},
			expectErr: false,
			expected: `[
  {
    "name": "Node2",
    "type": "TypeB",
    "id": "2",
    "metadata": {
      "key": "value"
    }
  }
]`,
		},
		{
			name: "Valid node with invalid metadata",
			nodes: []*v1.Node{
				{
					Name:     "Node3",
					Type:     "TypeC",
					Id:       3,
					Metadata: []byte(`invalid json`),
				},
			},
			expectErr:   true,
			expectedErr: "failed to unmarshal metadata for node Node3",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			output, err := FormatNodeJSON(tt.nodes)

			if tt.expectErr {
				if err == nil {
					t.Errorf("Expected error but got none")
				} else if !strings.Contains(err.Error(), tt.expectedErr) {
					t.Errorf("Expected error to contain '%s', but got '%s'", tt.expectedErr, err.Error())
				}
			} else {
				if err != nil {
					t.Errorf("Unexpected error: %v", err)
				}

				var expectedBuffer bytes.Buffer
				if err := json.Compact(&expectedBuffer, []byte(tt.expected)); err != nil {
					t.Fatalf("Invalid expected JSON: %v", err)
				}

				var outputBuffer bytes.Buffer
				if err := json.Compact(&outputBuffer, output); err != nil {
					t.Fatalf("Invalid output JSON: %v", err)
				}

				if expectedBuffer.String() != outputBuffer.String() {
					t.Errorf("Expected output %s, got %s", expectedBuffer.String(), outputBuffer.String())
				}
			}
		})
	}
}
