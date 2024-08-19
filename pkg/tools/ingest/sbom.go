package ingest

import (
	"errors"
	"fmt"
	"os"
	"path/filepath"

	cdx "github.com/CycloneDX/cyclonedx-go"
	"github.com/bit-bom/minefield/pkg/graph"
)

// IngestSBOM ingests a SBOM file or directory into the storages backend.
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

	_, err := os.Stat(filePath)
	if err != nil {
		return fmt.Errorf("failed to stat file %s: %w", filePath, err)
	}

	file, err := os.Open(filePath)
	if err != nil {
		return err
	}
	defer file.Close()

	bom := new(cdx.BOM)
	decoder := cdx.NewBOMDecoder(file, cdx.BOMFileFormatJSON)
	if err = decoder.Decode(bom); err != nil {
		return fmt.Errorf("failed to decode BOM: %w", err)
	}

	if bom.Metadata == nil || bom.Metadata.Component == nil {
		return nil
	}
	mainBomNodes := []cdx.Component{*bom.Metadata.Component}

	stack := []cdx.Component{*bom.Metadata.Component}

	for len(stack) > 0 {
		comp := stack[0]
		stack = stack[1:]

		if comp.Components != nil && len(*comp.Components) > 0 {
			stack = append(stack, *comp.Components...)
			mainBomNodes = append(mainBomNodes, *comp.Components...)
		}
	}

	var mainPurls []string

	for _, mainBomNode := range mainBomNodes {
		mainPurl := fmt.Sprintf("pkg:generic/%s", mainBomNode.Name)

		mainPurls = append(mainPurls, mainPurl)
	}

	var mainGraphNodes []*graph.Node

	for i := range mainPurls {
		mainGraphNode, err := graph.AddNode(storage, "library", bom, mainPurls[i])
		if err != nil {
			return fmt.Errorf("failed to parse SBOM file %s: %w", filePath, err)
		}
		mainGraphNodes = append(mainGraphNodes, mainGraphNode)
	}

	for _, node := range *bom.Components {

		purl := fmt.Sprintf("pkg:generic/%s", node.Name)

		graphNode, err := graph.AddNode(storage, "library", any(node), purl)
		if err != nil {
			if errors.Is(err, graph.ErrNodeAlreadyExists) {
				// TODO: Add a logger
				fmt.Println("Skipping...")
				continue
			}
			return fmt.Errorf("failed to add node: %w", err)
		}

		for _, mainGraphNode := range mainGraphNodes {
			if mainGraphNode.ID == graphNode.ID {
				continue
			}
			if err := mainGraphNode.SetDependency(storage, graphNode); err != nil {
				return fmt.Errorf("failed to add dependencies: %w", err)
			}
		}
	}

	return nil
}
