package sourcetruth

import (
	"os"
	"testing"
)

func TestSimulator_InspectGDrive(t *testing.T) {
	sim := NewSimulator("../.source_of_truth")
	if _, err := os.Stat("../../.source_of_truth"); err == nil {
		sim = NewSimulator("../../.source_of_truth")
	}

	inspection, err := sim.Inspect("https://internal.docs/finance/mrr_spec.pdf")

	if err != nil {
		t.Fatalf("Unexpected error inspecting GDrive document: %v", err)
	}

	if !inspection.Exists {
		t.Errorf("Expected simulated file to exist at %s, but marked as not existing", inspection.LocalPath)
	}

	if inspection.AssetType != TypeGDrive {
		t.Errorf("Expected AssetType %s, got %s", TypeGDrive, inspection.AssetType)
	}

	if inspection.Owner != "human:finance-team" {
		t.Errorf("Expected owner 'human:finance-team', got '%s'", inspection.Owner)
	}
}

func TestSimulator_InspectSpreadsheet(t *testing.T) {
	sim := NewSimulator("../.source_of_truth")
	if _, err := os.Stat("../../.source_of_truth"); err == nil {
		sim = NewSimulator("../../.source_of_truth")
	}

	inspection, err := sim.Inspect("https://docs.google.com/spreadsheets/d/9Z8Y7X6W5V4U3T2S1R0Q/edit")

	if err != nil {
		t.Fatalf("Unexpected error inspecting Spreadsheet: %v", err)
	}

	if !inspection.Exists {
		t.Errorf("Expected spreadsheet file to exist, but got false")
	}

	if len(inspection.Columns) == 0 {
		t.Errorf("Expected columns in spreadsheet metadata, got 0")
	}
}
