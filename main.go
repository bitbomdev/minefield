package main

import (
	"fmt"
	"github.com/turrisxyz/BitBom/pkg"
	"log"
	"path/filepath"
)

func main() {
	logger := log.Default()

	filePaths, err := filepath.Glob("test-data/*")
	if err != nil {
		logger.Fatal(fmt.Printf("Failed to find files: %v", err))
	}

	storage := pkg.GetStorageInstance("localhost:6379")

	// Loop through each file path
	for _, sbomPath := range filePaths {

		err := pkg.IngestSBOM(sbomPath, storage)
		if err != nil {
			logger.Fatal(fmt.Printf("Failed to ingest sbom: %v, err: %v", sbomPath, err))
		}
	}

	keys, err := storage.GetAllKeys()
	if err != nil {
		logger.Fatal(fmt.Printf("Failed to get all keys, %v", err))
	}

	for _, key := range keys {
		node, err := storage.GetNode(key)
		if err != nil {
			logger.Fatal(fmt.Printf("Failed to get node, %v", err))
		}

		fmt.Printf("id: %v, child: %+v, parent: %+v\n", node.GetID(), node.GetChild(), node.GetParent())
	}
}
