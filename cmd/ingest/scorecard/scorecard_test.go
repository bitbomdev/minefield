package scorecard

import (
	"testing"
)

func TestNew(t *testing.T) {
	cmd := New()

	if cmd.Use != "scorecard [path to scorecard file/dir]" {
		t.Errorf("expected Use to be 'scorecard [path to scorecard file/dir]', got %s", cmd.Use)
	}

	if cmd.Short != "Graph scorecard data into the graph, and connect it to existing library nodes" {
		t.Errorf("expected Short to be 'Graph scorecard data into the graph, and connect it to existing library nodes', got %s", cmd.Short)
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
