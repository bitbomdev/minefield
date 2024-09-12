package osv

import (
	"archive/zip"
	"fmt"
	"io"
	"os"
	"path/filepath"

	"github.com/bit-bom/minefield/pkg/graph"
	"github.com/bit-bom/minefield/pkg/tools"
	"github.com/bit-bom/minefield/pkg/tools/ingest"
	"github.com/spf13/cobra"
)

type options struct {
	storage graph.Storage
}

func (o *options) AddFlags(_ *cobra.Command) {}

func (o *options) Run(_ *cobra.Command, args []string) error {
	// load vuln data into storage
	progress := func(count int, path string) {
		fmt.Printf("\r\033[K%s", printProgress(count, path))
	}
	if err := ingest.Vulnerabilities(o.storage, progress); err != nil {
		return fmt.Errorf("failed to load vuln data: %w", err)
	}
	fmt.Println("\nVulnerabilities loaded successfully")

	return nil
}

func New(storage graph.Storage) *cobra.Command {
	o := &options{
		storage: storage,
	}
	cmd := &cobra.Command{
		Use:               "load [zip file or vuln dir or vuln file]",
		Short:             "Load vuln data into storage",
		RunE:              o.Run,
		Args:              cobra.ExactArgs(1),
		DisableAutoGenTag: true,
	}
	o.AddFlags(cmd)

	return cmd
}

func printProgress(count int, path string) string {
	return fmt.Sprintf("\033[1;36mIngested %d vulnerabilities\033[0m | \033[1;34mCurrent: %s\033[0m", count, tools.TruncateString(path, 50))
}

// VulnerabilitiesToStorage takes in a dir or file path and dumps the vulns into the storage. The vulns can either be json files or zip files
func (o *options) VulnerabilitiesToStorage(path string, progress func(count int, id string)) (int, error) {
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
			subCount, err := o.VulnerabilitiesToStorage(entryPath, progress)
			count += subCount
			if progress != nil {
				progress(count, entryPath)
			}
			if err != nil {
				errors = append(errors, fmt.Errorf("failed to ingest vulnerabilities from path %s: %w", entryPath, err))
			}
		}
	} else {
		if filepath.Ext(path) == ".zip" {
			subCount, err := o.processZipFile(path, progress)
			if err != nil {
				errors = append(errors, fmt.Errorf("failed to process zip file %s: %w", path, err))
			} else {
				count += subCount
				if progress != nil {
					progress(count, path)
				}
			}
		} else if filepath.Ext(path) == ".json" {
			if err := o.processJSONFile(path); err != nil {
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
		return 0, fmt.Errorf("errors occurred during vulnerabilities ingestion: %v", errors)
	}

	return count, nil
}

func (o *options) processZipFile(filePath string, progress func(count int, id string)) (int, error) {
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

	return o.VulnerabilitiesToStorage(tempDir, progress)
}

func (o *options) processJSONFile(filePath string) error {
	content, err := os.ReadFile(filePath)
	if err != nil {
		return fmt.Errorf("failed to read JSON file %s: %w", filePath, err)
	}
	return ingest.LoadVulnerabilities(o.storage, content)
}
