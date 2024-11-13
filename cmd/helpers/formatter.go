package helpers

import (
	"encoding/json"
	"fmt"
	"strconv"

	v1 "github.com/bitbomdev/minefield/gen/api/v1"
)

type nodeOutput struct {
	Name     string                 `json:"name"`
	Type     string                 `json:"type"`
	ID       string                 `json:"id"`
	Metadata map[string]interface{} `json:"metadata,omitempty"`
}

// FormatNodeJSON formats the nodes as JSON.
func FormatNodeJSON(nodes []*v1.Node) ([]byte, error) {
	if nodes == nil {
		return nil, fmt.Errorf("nodes cannot be nil")
	}

	if len(nodes) == 0 {
		return nil, fmt.Errorf("no nodes found")
	}

	outputs := make([]nodeOutput, 0, len(nodes))
	for _, node := range nodes {
		var metadata map[string]interface{}
		if len(node.Metadata) > 0 {
			if err := json.Unmarshal(node.Metadata, &metadata); err != nil {
				return nil, fmt.Errorf("failed to unmarshal metadata for node %s: %w", node.Name, err)
			}
		}

		outputs = append(outputs, nodeOutput{
			Name:     node.Name,
			Type:     node.Type,
			ID:       strconv.FormatUint(uint64(node.Id), 10),
			Metadata: metadata,
		})
	}

	return json.MarshalIndent(outputs, "", "  ")
}
