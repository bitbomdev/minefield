package ingest

import (
	"archive/zip"
	"fmt"
	"io"
	"os"
	"path/filepath"
)

// LoadDataFromPath takes in a directory or file path and processes the data into the storage.
// The data can either be JSON files or ZIP files containing JSON files.

type Data struct {
	Path string
	Data []byte
}

func LoadDataFromPath(path string) ([]Data, error) {
	info, err := os.Stat(path)
	if err != nil {
		return nil, fmt.Errorf("error accessing path %s: %w", path, err)
	}

	result := []Data{}

	var errors []error

	if info.IsDir() {
		entries, err := os.ReadDir(path)
		if err != nil {
			return nil, fmt.Errorf("failed to read directory %s: %w", path, err)
		}
		for _, entry := range entries {
			entryPath := filepath.Join(path, entry.Name())
			subResult, err := LoadDataFromPath(entryPath)
			if err != nil {
				errors = append(errors, fmt.Errorf("failed to load data from path %s: %w", entryPath, err))
			} else {
				result = append(result, subResult...)
			}
			if err != nil {
				errors = append(errors, fmt.Errorf("failed to ingest data from path %s: %w", entryPath, err))
			}
		}
	} else {
		switch filepath.Ext(path) {
		case ".zip":
			subResult, err := processZipFile(path)
			if err != nil {
				errors = append(errors, fmt.Errorf("failed to process zip file %s: %w", path, err))
			} else {
				result = subResult
			}
		case ".json":
			data, err := os.ReadFile(path)
			if err != nil {
				return nil, fmt.Errorf("failed to read JSON file %s: %w", path, err)
			}
			result = append(result, Data{Path: path, Data: data})
		}
	}

	if len(errors) > 0 {
		return nil, fmt.Errorf("errors occurred during data ingestion: %v", errors)
	}

	return result, nil
}

func processZipFile(filePath string) ([]Data, error) {
	r, err := zip.OpenReader(filePath)
	if err != nil {
		return nil, fmt.Errorf("failed to open zip file %s: %w", filePath, err)
	}
	defer r.Close()

	tempDir, err := os.MkdirTemp("", "unzipped")
	if err != nil {
		return nil, fmt.Errorf("failed to create temp directory: %w", err)
	}
	defer os.RemoveAll(tempDir)

	for _, f := range r.File {
		rc, err := f.Open()
		if err != nil {
			return nil, fmt.Errorf("failed to open file %s in zip: %w", f.Name, err)
		}
		defer rc.Close()

		extractedFilePath := filepath.Join(tempDir, f.Name)
		outFile, err := os.Create(extractedFilePath)
		if err != nil {
			return nil, fmt.Errorf("failed to create file %s: %w", extractedFilePath, err)
		}
		defer outFile.Close()
		_, err = io.Copy(outFile, rc)
		if err != nil {
			return nil, fmt.Errorf("failed to copy file %s: %w", extractedFilePath, err)
		}
	}

	return LoadDataFromPath(tempDir)
}
