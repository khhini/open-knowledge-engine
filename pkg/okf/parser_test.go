package okf

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
	"time"
)

func TestParseConcept(t *testing.T) {
	tmpDir, err := os.MkdirTemp("", "okf-parser-test-*")
	if err != nil {
		t.Fatalf("failed to create temp dir: %v", err)
	}
	defer os.RemoveAll(tmpDir)

	// Case 1: Valid Concept
	validContent := `---
type: concept
title: Test Concept
description: A test description
status: stable
stale_after: 2026-12-31
verified:
  - actor: human:tester
    at: 2026-08-17T12:00:00Z
    notes: Tested OK
---
This is the body. [[another-concept]]`

	validPath := filepath.Join(tmpDir, "test-concept.md")
	if err := os.WriteFile(validPath, []byte(validContent), 0644); err != nil {
		t.Fatalf("failed to write test file: %v", err)
	}

	concept, err := ParseConcept(validPath)
	if err != nil {
		t.Fatalf("unexpected error parsing valid concept: %v", err)
	}

	if concept.Frontmatter.Type != "concept" {
		t.Errorf("expected type 'concept', got '%s'", concept.Frontmatter.Type)
	}
	if concept.Frontmatter.Title != "Test Concept" {
		t.Errorf("expected title 'Test Concept', got '%s'", concept.Frontmatter.Title)
	}
	if len(concept.Frontmatter.Verified) != 1 || concept.Frontmatter.Verified[0].Actor != "human:tester" {
		t.Errorf("expected verified actor 'human:tester'")
	}
	if concept.IsStale {
		t.Errorf("expected concept to not be stale at evaluation since stale_after is 2026-12-31")
	}

	// Case 2: Missing Type Key
	invalidContent := `---
title: Missing Type
---
Body content.`
	invalidPath := filepath.Join(tmpDir, "invalid-concept.md")
	if err := os.WriteFile(invalidPath, []byte(invalidContent), 0644); err != nil {
		t.Fatalf("failed to write test file: %v", err)
	}

	_, err = ParseConcept(invalidPath)
	if err == nil {
		t.Error("expected error parsing concept without type key, got nil")
	}
}

func TestEvaluateTrustTier(t *testing.T) {
	now := time.Date(2026, 8, 17, 12, 0, 0, 0, time.UTC)

	// Stale Case
	fmStale := &Frontmatter{
		Type:       "concept",
		StaleAfter: "2026-08-16",
	}
	if tier := EvaluateTrustTier(fmStale, now); tier != TierStale {
		t.Errorf("expected TierStale, got %s", tier)
	}

	// Human-Reviewed Case
	fmHuman := &Frontmatter{
		Type: "concept",
		Verified: []Verification{
			{Actor: "human:reviewer", At: now},
		},
	}
	if tier := EvaluateTrustTier(fmHuman, now); tier != TierHumanReviewed {
		t.Errorf("expected TierHumanReviewed, got %s", tier)
	}

	// Machine-Confirmed Case
	fmMachine := &Frontmatter{
		Type: "concept",
		Verified: []Verification{
			{Actor: "process:validator", At: now},
		},
	}
	if tier := EvaluateTrustTier(fmMachine, now); tier != TierMachineConfirmed {
		t.Errorf("expected TierMachineConfirmed, got %s", tier)
	}

	// Unverified Case
	fmUnverified := &Frontmatter{
		Type: "concept",
	}
	if tier := EvaluateTrustTier(fmUnverified, now); tier != TierUnverified {
		t.Errorf("expected TierUnverified, got %s", tier)
	}
}

func TestFormatConcept(t *testing.T) {
	fm := &Frontmatter{
		Type:  "concept",
		Title: "Formatted Title",
	}
	body := "Formatted body content."

	bytes, err := FormatConcept(fm, body)
	if err != nil {
		t.Fatalf("unexpected error formatting concept: %v", err)
	}

	// Verify formatted output contains key parts
	outStr := string(bytes)
	if !strings.Contains(outStr, "type: concept") {
		t.Errorf("expected output to contain type, got:\n%s", outStr)
	}
	if !strings.Contains(outStr, "Formatted body content.") {
		t.Errorf("expected output to contain body, got:\n%s", outStr)
	}
}
