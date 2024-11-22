package helpers

import (
	"archive/zip"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"strings"
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

// processZipFile safely extracts files from a ZIP archive, preventing Zip Slip.
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
	// Ensure the temp directory is removed in case of an error.
	defer func() {
		if err != nil {
			os.RemoveAll(tempDir)
		}
	}()

	for _, f := range r.File {
		// Clean the file name to remove any path traversal.
		cleanName := filepath.Clean(f.Name)

		// Prevent absolute paths.
		if filepath.IsAbs(cleanName) {
			return nil, fmt.Errorf("invalid file path %s: absolute paths are not allowed", f.Name)
		}

		// Resolve the absolute path.
		extractedFilePath := filepath.Join(tempDir, cleanName)
		absExtractedPath, err := filepath.Abs(extractedFilePath)
		if err != nil {
			return nil, fmt.Errorf("failed to get absolute path for %s: %w", extractedFilePath, err)
		}

		absTempDir, err := filepath.Abs(tempDir)
		if err != nil {
			return nil, fmt.Errorf("failed to get absolute path for temp directory: %w", err)
		}

		// Ensure that the extracted path is within the temp directory.
		if !strings.HasPrefix(absExtractedPath, absTempDir+string(os.PathSeparator)) {
			return nil, fmt.Errorf("invalid file path %s: outside of the extraction directory", f.Name)
		}

		// Create necessary directories.
		if f.FileInfo().IsDir() {
			if err := os.MkdirAll(absExtractedPath, os.ModePerm); err != nil {
				return nil, fmt.Errorf("failed to create directory %s: %w", absExtractedPath, err)
			}
			continue
		} else {
			if err := os.MkdirAll(filepath.Dir(absExtractedPath), os.ModePerm); err != nil {
				return nil, fmt.Errorf("failed to create directory for file %s: %w", absExtractedPath, err)
			}
		}

		rc, err := f.Open()
		if err != nil {
			return nil, fmt.Errorf("failed to open file %s in zip: %w", f.Name, err)
		}
		defer rc.Close()

		outFile, err := os.Create(absExtractedPath)
		if err != nil {
			return nil, fmt.Errorf("failed to create file %s: %w", absExtractedPath, err)
		}
		defer outFile.Close()

		if _, err := io.Copy(outFile, rc); err != nil {
			return nil, fmt.Errorf("failed to copy file %s: %w", absExtractedPath, err)
		}
	}

	return LoadDataFromPath(tempDir)
}
