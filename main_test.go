package main

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/open-spaced-repetition/go-fsrs/v3"
)

// Test helper functions
func createTempDir(t *testing.T) string {
	dir, err := os.MkdirTemp("", "srs_test")
	if err != nil {
		t.Fatalf("Failed to create temp dir: %v", err)
	}
	t.Cleanup(func() {
		os.RemoveAll(dir)
	})
	return dir
}

func createTempFile(t *testing.T, dir, filename, content string) string {
	path := filepath.Join(dir, filename)
	if err := os.MkdirAll(filepath.Dir(path), 0755); err != nil {
		t.Fatalf("Failed to create dir for %s: %v", path, err)
	}
	if err := os.WriteFile(path, []byte(content), 0644); err != nil {
		t.Fatalf("Failed to write file %s: %v", path, err)
	}
	return path
}

// String utilities tests
func TestStateToString(t *testing.T) {
	tests := []struct {
		state    fsrs.State
		expected string
	}{
		{fsrs.New, "New"},
		{fsrs.Learning, "Learning"},
		{fsrs.Review, "Review"},
		{fsrs.Relearning, "Relearning"},
		{fsrs.State(99), "Unknown"}, // Invalid state
	}

	for _, tt := range tests {
		result := StateToString(tt.state)
		if result != tt.expected {
			t.Errorf("StateToString(%v) = %q, want %q", tt.state, result, tt.expected)
		}
	}
}

func TestStringToState(t *testing.T) {
	tests := []struct {
		input    string
		expected fsrs.State
	}{
		{"New", fsrs.New},
		{"Learning", fsrs.Learning},
		{"Review", fsrs.Review},
		{"Relearning", fsrs.Relearning},
		{"Unknown", fsrs.New},     // Default fallback
		{"invalid", fsrs.New},     // Default fallback
		{"", fsrs.New},            // Default fallback
		{"learning", fsrs.New},    // Case sensitive
	}

	for _, tt := range tests {
		result := StringToState(tt.input)
		if result != tt.expected {
			t.Errorf("StringToState(%q) = %v, want %v", tt.input, result, tt.expected)
		}
	}
}

func TestStateToStringRoundTrip(t *testing.T) {
	states := []fsrs.State{fsrs.New, fsrs.Learning, fsrs.Review, fsrs.Relearning}
	
	for _, state := range states {
		str := StateToString(state)
		roundTrip := StringToState(str)
		if roundTrip != state {
			t.Errorf("Round trip failed for %v: %v -> %q -> %v", state, state, str, roundTrip)
		}
	}
}