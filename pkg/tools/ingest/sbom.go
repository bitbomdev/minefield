package ingest

import (
	"errors"
	"fmt"
	"os"
	"path/filepath"

	"github.com/bit-bom/minefield/pkg/graph"
	"github.com/protobom/protobom/pkg/reader"
)

// SBOM ingests a SBOM file or directory into the storage backend.
func SBOM(sbomPath string, storage graph.Storage) error {
	info, err := os.Stat(sbomPath)
	if err != nil {
		return fmt.Errorf("error accessing path %s: %w", sbomPath, err)
	}

	if info.IsDir() {
		entries, err := os.ReadDir(sbomPath)
		if err != nil {
			return fmt.Errorf("failed to read directory %s: %w", sbomPath, err)
		}
		for _, entry := range entries {
			entryPath := filepath.Join(sbomPath, entry.Name())
			if err := SBOM(entryPath, storage); err != nil {
				return fmt.Errorf("failed to ingest SBOM from path %s: %w", entryPath, err)
			}
		}
	} else {
		return processSBOMFile(sbomPath, storage)
	}

	return nil
}

func processSBOMFile(filePath string, storage graph.Storage) error {
	if filePath == "" {
		return fmt.Errorf("file path is empty")
	}

	file, err := os.Open(filePath)
	if err != nil {
		return fmt.Errorf("failed to open file %s: %w", filePath, err)
	}

	// Create a new protobom reader
	r := reader.New()

	// Parse the SBOM file
	document, err := r.ParseFile(filePath)
	if err != nil {
		return fmt.Errorf("failed to parse SBOM file %s: %w", filePath, err)
	}

	// Get the node list from the document
	nodeList := document.GetNodeList()
	if nodeList == nil {
		return nil
	}

	// Process each node in the SBOM

	nameToId := map[string]uint32{}

	for _, node := range nodeList.GetNodes() {
		purl := fmt.Sprintf("pkg:generic/%s", node.GetName())

		graphNode, err := graph.AddNode(storage, "library", file, purl)
		if err != nil {
			if errors.Is(err, graph.ErrNodeAlreadyExists) {
				// TODO: Add a logger
				fmt.Println("Skipping...")
				continue
			}
			return fmt.Errorf("failed to add node: %w", err)
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
