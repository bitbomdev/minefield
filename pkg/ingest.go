package pkg

import (
	"fmt"
	"github.com/RoaringBitmap/roaring"
	"github.com/protobom/protobom/pkg/reader"
	"github.com/protobom/protobom/pkg/sbom"
)

func IngestSBOM(sbomPath string, storage Storage[any]) error {
	// Create a new protobom SBOM reader:
	sbomReader := reader.New()
	document, err := sbomReader.ParseFile(sbomPath)
	if err != nil {
		return fmt.Errorf("failed to parse SBOM file: %w", err)
	}

	protoIDToNodeID := map[string]uint32{}
	var nodesData []struct {
		Type     string
		Metadata any
		Parent   roaring.Bitmap
		Child    roaring.Bitmap
	}

	// Iterate over each node in the SBOM document
	for range document.GetNodeList().GetNodes() {
		parent := roaring.New()
		child := roaring.New()

		// Collect node data
		nodesData = append(nodesData, struct {
			Type     string
			Metadata any
			Parent   roaring.Bitmap
			Child    roaring.Bitmap
		}{
			Type:     "empty_type",
			Metadata: "empty_metadata",
			Parent:   *parent,
			Child:    *child,
		})
	}

	// Create and save nodes in batch
	nodes, err := AddNodes(storage, nodesData)
	if err != nil {
		return fmt.Errorf("failed to add nodes: %w", err)
	}

	// Map proto IDs to node IDs
	for i, node := range document.GetNodeList().GetNodes() {
		protoIDToNodeID[node.Id] = nodes[i].Id
	}

	// Iterate over the dependencies of the node and create a dependency edge
	err2 := addDependency(document, storage, protoIDToNodeID)
	if err2 != nil {
		return err2
	}

	return nil
}

// addDependency iterates over all the edges protobom sbom document and creates a dependency edge between each node in an edge
func addDependency(document *sbom.Document, storage Storage[any], protoIDToNodeID map[string]uint32) error {
	for _, edge := range document.GetNodeList().GetEdges() {
		for _, to := range edge.To {
			fromNode, err := storage.GetNode(protoIDToNodeID[edge.From])
			if err != nil {
				return err
			}

			toNode, err := storage.GetNode(protoIDToNodeID[to])
			if err != nil {
				return err
			}

			err = fromNode.SetDependency(storage, toNode)
			if err != nil {
				return err
			}
		}
	}
	return nil
}
