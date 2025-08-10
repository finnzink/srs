package core

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
	"time"

	"github.com/open-spaced-repetition/go-fsrs/v3"
)

func TestParseCard(t *testing.T) {
	// Create a temporary test file
	content := `<!-- FSRS: due:2025-01-01T00:00:00Z, stability:2.50, difficulty:5.00, elapsed_days:0, scheduled_days:1, reps:1, lapses:0, state:Learning -->

What is Go?
---
A programming language developed by Google.`

	tmpDir := t.TempDir()
	cardPath := filepath.Join(tmpDir, "test.md")
	
	err := os.WriteFile(cardPath, []byte(content), 0644)
	if err != nil {
		t.Fatalf("Failed to write test file: %v", err)
	}

	card, err := ParseCard(cardPath)
	if err != nil {
		t.Fatalf("ParseCard failed: %v", err)
	}

	if card.Question != "What is Go?" {
		t.Errorf("Expected question 'What is Go?', got '%s'", card.Question)
	}

	if card.Answer != "A programming language developed by Google." {
		t.Errorf("Expected answer 'A programming language developed by Google.', got '%s'", card.Answer)
	}

	if card.FilePath != cardPath {
		t.Errorf("Expected file path '%s', got '%s'", cardPath, card.FilePath)
	}

	// Check FSRS metadata
	expectedDue := time.Date(2025, 1, 1, 0, 0, 0, 0, time.UTC)
	if !card.FSRSCard.Due.Equal(expectedDue) {
		t.Errorf("Expected due date %v, got %v", expectedDue, card.FSRSCard.Due)
	}

	if card.FSRSCard.Stability != 2.50 {
		t.Errorf("Expected stability 2.50, got %.2f", card.FSRSCard.Stability)
	}

	if card.FSRSCard.State != fsrs.Learning {
		t.Errorf("Expected state Learning, got %v", card.FSRSCard.State)
	}
}

func TestParseCardWithoutMetadata(t *testing.T) {
	content := `What is the capital of France?
---
Paris`

	tmpDir := t.TempDir()
	cardPath := filepath.Join(tmpDir, "test.md")
	
	err := os.WriteFile(cardPath, []byte(content), 0644)
	if err != nil {
		t.Fatalf("Failed to write test file: %v", err)
	}

	card, err := ParseCard(cardPath)
	if err != nil {
		t.Fatalf("ParseCard failed: %v", err)
	}

	if card.Question != "What is the capital of France?" {
		t.Errorf("Expected question 'What is the capital of France?', got '%s'", card.Question)
	}

	if card.Answer != "Paris" {
		t.Errorf("Expected answer 'Paris', got '%s'", card.Answer)
	}

	// Should have default FSRS state
	if card.FSRSCard.State != fsrs.New {
		t.Errorf("Expected state New for new card, got %v", card.FSRSCard.State)
	}
}

func TestFindCards(t *testing.T) {
	tmpDir := t.TempDir()
	
	// Create test cards
	cards := map[string]string{
		"card1.md": "Question 1\n---\nAnswer 1",
		"card2.md": "Question 2\n---\nAnswer 2",
		"subdir/card3.md": "Question 3\n---\nAnswer 3",
	}

	for path, content := range cards {
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

	foundCards, err := FindCards(tmpDir)
	if err != nil {
		t.Fatalf("FindCards failed: %v", err)
	}

	if len(foundCards) != 3 {
		t.Errorf("Expected 3 cards, got %d", len(foundCards))
	}

	// Check that all cards were found
	foundPaths := make(map[string]bool)
	for _, card := range foundCards {
		rel, _ := filepath.Rel(tmpDir, card.FilePath)
		foundPaths[rel] = true
	}

	for expectedPath := range cards {
		if !foundPaths[expectedPath] {
			t.Errorf("Expected to find card at %s", expectedPath)
		}
	}
}

func TestGetDueCards(t *testing.T) {
	now := time.Now()
	
	cards := []*Card{
		{
			Question: "Due now",
			FSRSCard: fsrs.Card{Due: now.Add(-1 * time.Hour)}, // Past due
		},
		{
			Question: "Due later",
			FSRSCard: fsrs.Card{Due: now.Add(1 * time.Hour)}, // Future due
		},
		{
			Question: "Due exactly now",
			FSRSCard: fsrs.Card{Due: now}, // Due now
		},
	}

	dueCards := GetDueCards(cards)

	if len(dueCards) != 2 {
		t.Errorf("Expected 2 due cards, got %d", len(dueCards))
	}

	// Check the due cards are correct
	expectedQuestions := map[string]bool{
		"Due now": true,
		"Due exactly now": true,
	}

	for _, card := range dueCards {
		if !expectedQuestions[card.Question] {
			t.Errorf("Unexpected due card: %s", card.Question)
		}
	}
}

func TestUpdateFSRSMetadata(t *testing.T) {
	content := `What is testing?
---
A way to verify code works correctly.`

	tmpDir := t.TempDir()
	cardPath := filepath.Join(tmpDir, "test.md")
	
	err := os.WriteFile(cardPath, []byte(content), 0644)
	if err != nil {
		t.Fatalf("Failed to write test file: %v", err)
	}

	card, err := ParseCard(cardPath)
	if err != nil {
		t.Fatalf("ParseCard failed: %v", err)
	}

	// Modify the FSRS data
	card.FSRSCard.Stability = 3.14
	card.FSRSCard.Difficulty = 6.28
	card.FSRSCard.State = fsrs.Review

	err = card.UpdateFSRSMetadata()
	if err != nil {
		t.Fatalf("UpdateFSRSMetadata failed: %v", err)
	}

	// Read the file back and verify metadata was written
	updatedContent, err := os.ReadFile(cardPath)
	if err != nil {
		t.Fatalf("Failed to read updated file: %v", err)
	}

	contentStr := string(updatedContent)
	if !strings.Contains(contentStr, "stability:3.14") {
		t.Error("Expected stability:3.14 in updated content")
	}

	if !strings.Contains(contentStr, "difficulty:6.28") {
		t.Error("Expected difficulty:6.28 in updated content")
	}

	if !strings.Contains(contentStr, "state:Review") {
		t.Error("Expected state:Review in updated content")
	}

	// Verify the original content is preserved
	if !strings.Contains(contentStr, "What is testing?") {
		t.Error("Original question should be preserved")
	}

	if !strings.Contains(contentStr, "A way to verify code works correctly.") {
		t.Error("Original answer should be preserved")
	}
}

func TestStateConversion(t *testing.T) {
	tests := []struct {
		state    fsrs.State
		expected string
	}{
		{fsrs.New, "New"},
		{fsrs.Learning, "Learning"},
		{fsrs.Review, "Review"},
		{fsrs.Relearning, "Relearning"},
	}

	for _, test := range tests {
		result := StateToString(test.state)
		if result != test.expected {
			t.Errorf("StateToString(%v) = %s, expected %s", test.state, result, test.expected)
		}

		// Test reverse conversion
		backToState := StringToState(result)
		if backToState != test.state {
			t.Errorf("StringToState(%s) = %v, expected %v", result, backToState, test.state)
		}
	}
}