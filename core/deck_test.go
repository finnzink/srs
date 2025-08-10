package core

import (
	"os"
	"path/filepath"
	"testing"
	"time"

	"github.com/open-spaced-repetition/go-fsrs/v3"
)

func TestLoadConfig(t *testing.T) {
	// Test loading config when no config exists
	config, err := LoadConfig()
	if err != nil {
		t.Fatalf("LoadConfig failed: %v", err)
	}

	if config.BaseDeckPath != "" {
		t.Error("Expected empty BaseDeckPath for non-existent config")
	}
}

func TestSaveAndLoadConfig(t *testing.T) {
	// Create a temporary config for testing
	originalHomeDir := os.Getenv("HOME")
	tmpDir := t.TempDir()
	os.Setenv("HOME", tmpDir)
	defer os.Setenv("HOME", originalHomeDir)

	config := &Config{
		BaseDeckPath: "/test/path/to/deck",
	}

	err := SaveConfig(config)
	if err != nil {
		t.Fatalf("SaveConfig failed: %v", err)
	}

	loadedConfig, err := LoadConfig()
	if err != nil {
		t.Fatalf("LoadConfig failed: %v", err)
	}

	if loadedConfig.BaseDeckPath != config.BaseDeckPath {
		t.Errorf("Expected BaseDeckPath '%s', got '%s'", 
			config.BaseDeckPath, loadedConfig.BaseDeckPath)
	}
}

func TestResolveDeckPath(t *testing.T) {
	config := &Config{
		BaseDeckPath: "/home/user/flashcards",
	}

	tests := []struct {
		input    string
		expected string
	}{
		{".", "/home/user/flashcards"},
		{"", "/home/user/flashcards"},
		{"spanish", "/home/user/flashcards/spanish"},
		{"spanish/grammar", "/home/user/flashcards/spanish/grammar"},
	}

	for _, test := range tests {
		result, err := ResolveDeckPath(test.input, config)
		if err != nil {
			t.Errorf("ResolveDeckPath(%s) failed: %v", test.input, err)
			continue
		}

		// Convert to absolute path for comparison
		expected, _ := filepath.Abs(test.expected)
		result, _ = filepath.Abs(result)

		if result != expected {
			t.Errorf("ResolveDeckPath(%s) = %s, expected %s", 
				test.input, result, expected)
		}
	}

	// Test error case - no config
	_, err := ResolveDeckPath("test", nil)
	if err == nil {
		t.Error("Expected error when config is nil")
	}

	// Test error case - empty config
	emptyConfig := &Config{}
	_, err = ResolveDeckPath("test", emptyConfig)
	if err == nil {
		t.Error("Expected error when BaseDeckPath is empty")
	}
}

func TestGetDeckStats(t *testing.T) {
	now := time.Now()
	
	cards := []*Card{
		// New card
		{FSRSCard: fsrs.Card{State: fsrs.New, Due: now.Add(time.Hour)}},
		// Learning card that's due
		{FSRSCard: fsrs.Card{State: fsrs.Learning, Due: now.Add(-time.Hour)}},
		// Review card that's not due
		{FSRSCard: fsrs.Card{State: fsrs.Review, Due: now.Add(time.Hour)}},
		// Relearning card that's due (counts as learning)
		{FSRSCard: fsrs.Card{State: fsrs.Relearning, Due: now.Add(-time.Hour)}},
		// Another new card that's due
		{FSRSCard: fsrs.Card{State: fsrs.New, Due: now.Add(-time.Hour)}},
	}

	stats := GetDeckStats(cards)

	if stats.TotalCards != 5 {
		t.Errorf("Expected TotalCards 5, got %d", stats.TotalCards)
	}

	if stats.DueCards != 3 {
		t.Errorf("Expected DueCards 3, got %d", stats.DueCards)
	}

	if stats.NewCards != 2 {
		t.Errorf("Expected NewCards 2, got %d", stats.NewCards)
	}

	if stats.LearningCards != 2 {
		t.Errorf("Expected LearningCards 2, got %d", stats.LearningCards)
	}

	if stats.ReviewCards != 1 {
		t.Errorf("Expected ReviewCards 1, got %d", stats.ReviewCards)
	}
}

func TestGetDeckTree(t *testing.T) {
	tmpDir := t.TempDir()
	
	// Create a test deck structure
	testStructure := map[string]string{
		"root1.md": "Q1\n---\nA1",
		"root2.md": "Q2\n---\nA2",
		"spanish/vocab.md": "Q3\n---\nA3",
		"spanish/grammar/verbs.md": "Q4\n---\nA4",
		"math/algebra.md": "Q5\n---\nA5",
	}

	// Create the test files
	for path, content := range testStructure {
		fullPath := filepath.Join(tmpDir, path)
		dir := filepath.Dir(fullPath)
		
		err := os.MkdirAll(dir, 0755)
		if err != nil {
			t.Fatalf("Failed to create directory %s: %v", dir, err)
		}
		
		err = os.WriteFile(fullPath, []byte(content), 0644)
		if err != nil {
			t.Fatalf("Failed to write test file %s: %v", fullPath, err)
		}
	}

	deckTree, err := GetDeckTree(tmpDir)
	if err != nil {
		t.Fatalf("GetDeckTree failed: %v", err)
	}

	// Check root directory stats
	rootStats, exists := deckTree[""]
	if !exists {
		t.Fatal("Expected root directory stats")
	}

	if rootStats.TotalCards != 2 {
		t.Errorf("Expected 2 cards in root, got %d", rootStats.TotalCards)
	}

	// Check spanish subdirectory stats
	spanishStats, exists := deckTree["spanish"]
	if !exists {
		t.Fatal("Expected spanish directory stats")
	}

	if spanishStats.TotalCards != 1 {
		t.Errorf("Expected 1 card in spanish dir, got %d", spanishStats.TotalCards)
	}

	// Check that nested directories are handled correctly
	grammarPath := filepath.Join("spanish", "grammar")
	grammarStats, exists := deckTree[grammarPath]
	if !exists {
		t.Fatal("Expected spanish/grammar directory stats")
	}

	if grammarStats.TotalCards != 1 {
		t.Errorf("Expected 1 card in grammar dir, got %d", grammarStats.TotalCards)
	}
}

func TestGetCardsInDirectory(t *testing.T) {
	tmpDir := t.TempDir()
	
	// Create test structure with nested directories
	testFiles := map[string]string{
		"card1.md": "Q1\n---\nA1",
		"card2.md": "Q2\n---\nA2",
		"subdir/card3.md": "Q3\n---\nA3", // Should not be included
	}

	// Create the test files
	for path, content := range testFiles {
		fullPath := filepath.Join(tmpDir, path)
		dir := filepath.Dir(fullPath)
		
		err := os.MkdirAll(dir, 0755)
		if err != nil {
			t.Fatalf("Failed to create directory %s: %v", dir, err)
		}
		
		err = os.WriteFile(fullPath, []byte(content), 0644)
		if err != nil {
			t.Fatalf("Failed to write test file %s: %v", fullPath, err)
		}
	}

	cards, err := getCardsInDirectory(tmpDir)
	if err != nil {
		t.Fatalf("getCardsInDirectory failed: %v", err)
	}

	// Should only include cards directly in the directory, not in subdirectories
	if len(cards) != 2 {
		t.Errorf("Expected 2 cards in directory, got %d", len(cards))
	}

	// Verify the correct cards were found
	foundQuestions := make(map[string]bool)
	for _, card := range cards {
		foundQuestions[card.Question] = true
	}

	expectedQuestions := []string{"Q1", "Q2"}
	for _, expected := range expectedQuestions {
		if !foundQuestions[expected] {
			t.Errorf("Expected to find question '%s'", expected)
		}
	}

	// Q3 should not be found since it's in a subdirectory
	if foundQuestions["Q3"] {
		t.Error("Should not have found Q3 from subdirectory")
	}
}