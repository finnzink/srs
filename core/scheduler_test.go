package core

import (
	"os"
	"testing"

	"github.com/open-spaced-repetition/go-fsrs/v3"
)

func TestNewReviewSession(t *testing.T) {
	cards := []*Card{
		{Question: "Q1", Answer: "A1", FSRSCard: fsrs.NewCard()},
		{Question: "Q2", Answer: "A2", FSRSCard: fsrs.NewCard()},
	}

	session := NewReviewSession(cards)

	if session == nil {
		t.Fatal("NewReviewSession returned nil")
	}

	if len(session.cards) != 2 {
		t.Errorf("Expected 2 cards in session, got %d", len(session.cards))
	}

	if session.current != 0 {
		t.Errorf("Expected current index 0, got %d", session.current)
	}
}

func TestCurrentCard(t *testing.T) {
	cards := []*Card{
		{Question: "Q1", Answer: "A1", FSRSCard: fsrs.NewCard()},
		{Question: "Q2", Answer: "A2", FSRSCard: fsrs.NewCard()},
	}

	session := NewReviewSession(cards)

	// Test getting current card
	card, err := session.CurrentCard()
	if err != nil {
		t.Fatalf("CurrentCard failed: %v", err)
	}

	if card.Question != "Q1" {
		t.Errorf("Expected question 'Q1', got '%s'", card.Question)
	}

	// Test when no more cards
	session.current = 2 // Beyond array bounds
	_, err = session.CurrentCard()
	if err == nil {
		t.Error("Expected error when no more cards, got nil")
	}
}

func TestHasNext(t *testing.T) {
	cards := []*Card{
		{Question: "Q1", Answer: "A1", FSRSCard: fsrs.NewCard()},
		{Question: "Q2", Answer: "A2", FSRSCard: fsrs.NewCard()},
	}

	session := NewReviewSession(cards)

	if !session.HasNext() {
		t.Error("Expected HasNext() to return true initially")
	}

	session.current = 1
	if !session.HasNext() {
		t.Error("Expected HasNext() to return true for second card")
	}

	session.current = 2
	if session.HasNext() {
		t.Error("Expected HasNext() to return false when past end")
	}
}

func TestProgress(t *testing.T) {
	cards := []*Card{
		{Question: "Q1", Answer: "A1", FSRSCard: fsrs.NewCard()},
		{Question: "Q2", Answer: "A2", FSRSCard: fsrs.NewCard()},
		{Question: "Q3", Answer: "A3", FSRSCard: fsrs.NewCard()},
	}

	session := NewReviewSession(cards)

	// Test initial progress
	current, total := session.Progress()
	if current != 1 || total != 3 {
		t.Errorf("Expected progress (1, 3), got (%d, %d)", current, total)
	}

	// Test after advancing
	session.current = 1
	current, total = session.Progress()
	if current != 2 || total != 3 {
		t.Errorf("Expected progress (2, 3), got (%d, %d)", current, total)
	}
}

func TestRateCard(t *testing.T) {
	// Create a test card file
	tmpDir := t.TempDir()
	card := &Card{
		Question: "Test question",
		Answer: "Test answer",
		FilePath: tmpDir + "/test.md",
		FSRSCard: fsrs.NewCard(),
	}

	// Write initial card content
	content := `Test question
---
Test answer`
	
	err := writeFile(card.FilePath, content)
	if err != nil {
		t.Fatalf("Failed to write test file: %v", err)
	}

	session := NewReviewSession([]*Card{card})

	originalState := card.FSRSCard.State
	originalReps := card.FSRSCard.Reps

	err = session.RateCard(fsrs.Good)
	if err != nil {
		t.Fatalf("RateCard failed: %v", err)
	}

	// Verify card was updated
	if card.FSRSCard.State == originalState && card.FSRSCard.Reps == originalReps {
		t.Error("Card FSRS data should have been updated after rating")
	}

	// Verify session moved to next card
	if session.current != 1 {
		t.Errorf("Expected current index 1 after rating, got %d", session.current)
	}

	// Test rating when no cards available
	err = session.RateCard(fsrs.Good)
	if err == nil {
		t.Error("Expected error when rating with no cards available")
	}
}

func TestRatingFromInt(t *testing.T) {
	tests := []struct {
		input    int
		expected fsrs.Rating
		hasError bool
	}{
		{1, fsrs.Again, false},
		{2, fsrs.Hard, false},
		{3, fsrs.Good, false},
		{4, fsrs.Easy, false},
		{0, fsrs.Again, true},  // Invalid
		{5, fsrs.Again, true},  // Invalid
		{-1, fsrs.Again, true}, // Invalid
	}

	for _, test := range tests {
		result, err := RatingFromInt(test.input)
		
		if test.hasError {
			if err == nil {
				t.Errorf("RatingFromInt(%d) expected error, got nil", test.input)
			}
		} else {
			if err != nil {
				t.Errorf("RatingFromInt(%d) unexpected error: %v", test.input, err)
			}
			if result != test.expected {
				t.Errorf("RatingFromInt(%d) = %v, expected %v", test.input, result, test.expected)
			}
		}
	}
}

func TestRatingToString(t *testing.T) {
	tests := []struct {
		rating   fsrs.Rating
		expected string
	}{
		{fsrs.Again, "Again"},
		{fsrs.Hard, "Hard"},
		{fsrs.Good, "Good"},
		{fsrs.Easy, "Easy"},
	}

	for _, test := range tests {
		result := RatingToString(test.rating)
		if result != test.expected {
			t.Errorf("RatingToString(%v) = %s, expected %s", test.rating, result, test.expected)
		}
	}
}

func TestUpdateCurrentCard(t *testing.T) {
	originalCard := &Card{Question: "Original", Answer: "Original", FSRSCard: fsrs.NewCard()}
	updatedCard := &Card{Question: "Updated", Answer: "Updated", FSRSCard: fsrs.NewCard()}

	session := NewReviewSession([]*Card{originalCard})

	// Update the current card
	session.UpdateCurrentCard(updatedCard)

	// Verify the card was updated
	current, err := session.CurrentCard()
	if err != nil {
		t.Fatalf("CurrentCard failed: %v", err)
	}

	if current.Question != "Updated" {
		t.Errorf("Expected updated question 'Updated', got '%s'", current.Question)
	}
}

// Helper function to write file content
func writeFile(path, content string) error {
	return os.WriteFile(path, []byte(content), 0644)
}