package ingest

import (
	"bytes"
	"errors"
	"fmt"

	"github.com/bitbomdev/minefield/pkg/graph"
	"github.com/protobom/protobom/pkg/reader"
)

func SBOM(storage graph.Storage, data []byte) error {
	if len(data) == 0 {
		return fmt.Errorf("data is empty")
	}
	// Create a new protobom reader
	r := reader.New()

	// Parse the SBOM file
	document, err := r.ParseStream(bytes.NewReader(data))
	if err != nil {
		return fmt.Errorf("failed to parse SBOM file: %w", err)
	}

	// Get the node list from the document
	nodeList := document.GetNodeList()
	if nodeList == nil {
		return nil
	}

	// Process each node in the SBOM

	nameToId := map[string]uint32{}

	for _, node := range nodeList.GetNodes() {
		purl := string(node.Purl())
		if purl == "" {
			purl = fmt.Sprintf("pkg:%s@%s", node.GetName(), node.GetVersion())
		}

		graphNode, err := graph.AddNode(storage, "library", node, purl)
		if err != nil {
			if errors.Is(err, graph.ErrNodeAlreadyExists) {
				// log.Printf("Skipping node %s: %s\n", node.GetName(), err)
			} else {
				return fmt.Errorf("failed to add node: %w", err)
			}
		}

		nameToId[node.Id] = graphNode.ID
	}

	for _, edge := range nodeList.Edges {
		fromNode, err := storage.GetNode(nameToId[edge.From])
		if err != nil {
			return fmt.Errorf("failed to get from node %s: %w", edge.From, err)
		}

		for _, to := range edge.To {

			toNode, err := storage.GetNode(nameToId[to])
			if err != nil {
				return fmt.Errorf("failed to to get node %s: %w", edge.To, err)
			}

			if fromNode.ID != toNode.ID {
				if err := fromNode.SetDependency(storage, toNode); err != nil {
					return fmt.Errorf("failed to add edge %s -> %s: %w", edge.From, to, err)
				}
			}

		}
	}

	return nil
}
