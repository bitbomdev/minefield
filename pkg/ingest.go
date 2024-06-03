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

	// Iterate over each node in the SBOM document
	for _, node := range document.GetNodeList().GetNodes() {
		parent := roaring.New()
		child := roaring.New()

		// Create and add the node to the storage
		// TODO: Add type and metadata
		graphNode, err := AddNode(storage, "empty_type", "empty_metadata", *parent, *child)
		if err != nil {
			return fmt.Errorf("failed to add node: %w", err)
		}
		protoIDToNodeID[node.Id] = graphNode.id
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
