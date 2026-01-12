package main

import "testing"

// TestValidateAWSRegion verifies region validation behavior.
func TestValidateAWSRegion(t *testing.T) {
	// Validate a known-good region.
	valid, err := validateAWSRegion("us-east-1")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if !valid {
		t.Fatalf("expected region to be valid")
	}

	// Validate a known-bad region.
	valid, err = validateAWSRegion("invalid")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if valid {
		t.Fatalf("expected region to be invalid")
	}
}

// TestParseDebugValue verifies debug parsing behavior.
func TestParseDebugValue(t *testing.T) {
	// Validate true variants.
	if !parseDebugValue("true") {
		t.Fatalf("expected true to be parsed as true")
	}
	if !parseDebugValue("yes") {
		t.Fatalf("expected yes to be parsed as true")
	}
	if !parseDebugValue("t") {
		t.Fatalf("expected t to be parsed as true")
	}

	// Validate false variants.
	if parseDebugValue("false") {
		t.Fatalf("expected false to be parsed as false")
	}
	if parseDebugValue("no") {
		t.Fatalf("expected no to be parsed as false")
	}
}
