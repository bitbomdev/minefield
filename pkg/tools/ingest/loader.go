package ingest

import (
	"archive/zip"
	"fmt"
	"io"
	"os"
	"path/filepath"

	"github.com/bitbomdev/minefield/pkg/graph"
)

// LoadDataFromPath takes in a directory or file path and processes the data into the storage.
// The data can either be JSON files or ZIP files containing JSON files.
func LoadDataFromPath(storage graph.Storage, path string, processJSON func(storage graph.Storage, content []byte) error, progress func(count int, id string)) (int, error) {
	info, err := os.Stat(path)
	if err != nil {
		return 0, fmt.Errorf("error accessing path %s: %w", path, err)
	}

	var errors []error
	count := 0

	if info.IsDir() {
		entries, err := os.ReadDir(path)
		if err != nil {
			return 0, fmt.Errorf("failed to read directory %s: %w", path, err)
		}
		for _, entry := range entries {
			entryPath := filepath.Join(path, entry.Name())
			subCount, err := LoadDataFromPath(storage, entryPath, processJSON, progress)
			count += subCount
			if progress != nil {
				progress(count, entryPath)
			}
			if err != nil {
				errors = append(errors, fmt.Errorf("failed to ingest data from path %s: %w", entryPath, err))
			}
		}
	} else {
		if filepath.Ext(path) == ".zip" {
			subCount, err := processZipFile(storage, path, processJSON, progress)
			if err != nil {
				errors = append(errors, fmt.Errorf("failed to process zip file %s: %w", path, err))
			} else {
				count += subCount
				if progress != nil {
					progress(count, path)
				}
			}
		} else if filepath.Ext(path) == ".json" {
			if err := processJSONFile(storage, path, processJSON); err != nil {
				errors = append(errors, fmt.Errorf("failed to process JSON file %s: %w", path, err))
			} else {
				count++
				if progress != nil {
					progress(count, path)
				}
			}
		}
	}

	if len(errors) > 0 {
		return 0, fmt.Errorf("errors occurred during data ingestion: %v", errors)
	}

	return count, nil
}

func processZipFile(storage graph.Storage, filePath string, processJSON func(storage graph.Storage, content []byte) error, progress func(count int, id string)) (int, error) {
	r, err := zip.OpenReader(filePath)
	if err != nil {
		return 0, fmt.Errorf("failed to open zip file %s: %w", filePath, err)
	}
	defer r.Close()

	tempDir, err := os.MkdirTemp("", "unzipped")
	if err != nil {
		return 0, fmt.Errorf("failed to create temp directory: %w", err)
	}
	defer os.RemoveAll(tempDir)

	for _, f := range r.File {
		rc, err := f.Open()
		if err != nil {
			return 0, fmt.Errorf("failed to open file %s in zip: %w", f.Name, err)
		}
		defer rc.Close()

		extractedFilePath := filepath.Join(tempDir, f.Name)
		outFile, err := os.Create(extractedFilePath)
		if err != nil {
			return 0, fmt.Errorf("failed to create file %s: %w", extractedFilePath, err)
		}
		defer outFile.Close()
		_, err = io.Copy(outFile, rc)
		if err != nil {
			return 0, fmt.Errorf("failed to copy file %s: %w", extractedFilePath, err)
		}
	}

	return LoadDataFromPath(storage, tempDir, processJSON, progress)
}

func processJSONFile(storage graph.Storage, filePath string, processJSON func(storage graph.Storage, content []byte) error) error {
	content, err := os.ReadFile(filePath)
	if err != nil {
		return fmt.Errorf("failed to read JSON file %s: %w", filePath, err)
	}
	return processJSON(storage, content)
}
