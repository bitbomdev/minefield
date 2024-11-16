package sbom

import (
	"testing"
)

func TestNew(t *testing.T) {
	cmd := New()

	if cmd.Use != "sbom [path to sbom file/dir]" {
		t.Errorf("expected Use to be 'sbom [path to sbom file/dir]', got %s", cmd.Use)
	}

	if cmd.Short != "Ingest an sbom into the graph " {
		t.Errorf("expected Short to be 'Ingest an sbom into the graph ', got %s", cmd.Short)
	}

	if cmd.Args == nil || cmd.Args(nil, []string{"arg1"}) != nil {
		t.Errorf("expected Args to be cobra.ExactArgs(1)")
	}

	if cmd.DisableAutoGenTag != true {
		t.Errorf("expected DisableAutoGenTag to be true")
	}

	if cmd.RunE == nil {
		t.Errorf("expected RunE to be set")
	}
}
