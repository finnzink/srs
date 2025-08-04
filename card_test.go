package main

import (
	"strings"
	"testing"
	"time"

	"github.com/open-spaced-repetition/go-fsrs/v3"
)

func TestParseFSRSMetadata(t *testing.T) {
	tests := []struct {
		name     string
		metadata string
		validate func(*testing.T, fsrs.Card)
	}{
		{
			name:     "empty metadata",
			metadata: "",
			validate: func(t *testing.T, card fsrs.Card) {
				// Should be equivalent to a new card
				newCard := fsrs.NewCard()
				if card.State != newCard.State {
					t.Errorf("Expected state %v, got %v", newCard.State, card.State)
				}
			},
		},
		{
			name:     "complete metadata",
			metadata: "due:2024-01-15T10:30:00Z, stability:2.50, difficulty:6.25, elapsed_days:5, scheduled_days:3, reps:10, lapses:2, state:Review",
			validate: func(t *testing.T, card fsrs.Card) {
				expectedDue, _ := time.Parse(time.RFC3339, "2024-01-15T10:30:00Z")
				if !card.Due.Equal(expectedDue) {
					t.Errorf("Expected due %v, got %v", expectedDue, card.Due)
				}
				if card.Stability != 2.50 {
					t.Errorf("Expected stability 2.50, got %f", card.Stability)
				}
				if card.Difficulty != 6.25 {
					t.Errorf("Expected difficulty 6.25, got %f", card.Difficulty)
				}
				if card.ElapsedDays != 5 {
					t.Errorf("Expected elapsed_days 5, got %d", card.ElapsedDays)
				}
				if card.ScheduledDays != 3 {
					t.Errorf("Expected scheduled_days 3, got %d", card.ScheduledDays)
				}
				if card.Reps != 10 {
					t.Errorf("Expected reps 10, got %d", card.Reps)
				}
				if card.Lapses != 2 {
					t.Errorf("Expected lapses 2, got %d", card.Lapses)
				}
				if card.State != fsrs.Review {
					t.Errorf("Expected state Review, got %v", card.State)
				}
			},
		},
		{
			name:     "partial metadata",
			metadata: "stability:1.75, difficulty:4.50, reps:3",
			validate: func(t *testing.T, card fsrs.Card) {
				if card.Stability != 1.75 {
					t.Errorf("Expected stability 1.75, got %f", card.Stability)
				}
				if card.Difficulty != 4.50 {
					t.Errorf("Expected difficulty 4.50, got %f", card.Difficulty)
				}
				if card.Reps != 3 {
					t.Errorf("Expected reps 3, got %d", card.Reps)
				}
				// Other fields should be defaults from NewCard()
				newCard := fsrs.NewCard()
				if card.ElapsedDays != newCard.ElapsedDays {
					t.Errorf("Expected default elapsed_days %d, got %d", newCard.ElapsedDays, card.ElapsedDays)
				}
			},
		},
		{
			name:     "malformed values",
			metadata: "stability:invalid, difficulty:6.25, elapsed_days:not_a_number, reps:5",
			validate: func(t *testing.T, card fsrs.Card) {
				// Invalid stability should remain default
				newCard := fsrs.NewCard()
				if card.Stability != newCard.Stability {
					t.Errorf("Expected default stability %f for invalid input, got %f", newCard.Stability, card.Stability)
				}
				// Valid difficulty should be parsed
				if card.Difficulty != 6.25 {
					t.Errorf("Expected difficulty 6.25, got %f", card.Difficulty)
				}
				// Invalid elapsed_days should remain default
				if card.ElapsedDays != newCard.ElapsedDays {
					t.Errorf("Expected default elapsed_days %d for invalid input, got %d", newCard.ElapsedDays, card.ElapsedDays)
				}
				// Valid reps should be parsed
				if card.Reps != 5 {
					t.Errorf("Expected reps 5, got %d", card.Reps)
				}
			},
		},
		{
			name:     "invalid date format",
			metadata: "due:not-a-date, stability:2.0",
			validate: func(t *testing.T, card fsrs.Card) {
				// Invalid date should remain default
				newCard := fsrs.NewCard()
				if !card.Due.Equal(newCard.Due) {
					t.Errorf("Expected default due date for invalid input, got %v", card.Due)
				}
				// Valid stability should be parsed
				if card.Stability != 2.0 {
					t.Errorf("Expected stability 2.0, got %f", card.Stability)
				}
			},
		},
		{
			name:     "unknown state",
			metadata: "state:UnknownState, reps:1",
			validate: func(t *testing.T, card fsrs.Card) {
				// Unknown state should default to New
				if card.State != fsrs.New {
					t.Errorf("Expected state New for unknown input, got %v", card.State)
				}
				if card.Reps != 1 {
					t.Errorf("Expected reps 1, got %d", card.Reps)
				}
			},
		},
		{
			name:     "extra whitespace and formatting",
			metadata: " due: 2024-01-15T10:30:00Z , stability: 2.50 , difficulty: 6.25 ",
			validate: func(t *testing.T, card fsrs.Card) {
				expectedDue, _ := time.Parse(time.RFC3339, "2024-01-15T10:30:00Z")
				if !card.Due.Equal(expectedDue) {
					t.Errorf("Expected due %v, got %v", expectedDue, card.Due)
				}
				if card.Stability != 2.50 {
					t.Errorf("Expected stability 2.50, got %f", card.Stability)
				}
				if card.Difficulty != 6.25 {
					t.Errorf("Expected difficulty 6.25, got %f", card.Difficulty)
				}
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			card := parseFSRSMetadata(tt.metadata)
			tt.validate(t, card)
		})
	}
}

func TestParseCard(t *testing.T) {
	tempDir := createTempDir(t)

	tests := []struct {
		name         string
		content      string
		expectedCard func(*testing.T, *Card)
		expectError  bool
	}{
		{
			name: "basic card",
			content: `# Question
What is 2 + 2?

---

# Answer
4`,
			expectedCard: func(t *testing.T, card *Card) {
				if !strings.Contains(card.Question, "What is 2 + 2?") {
					t.Errorf("Expected question to contain 'What is 2 + 2?', got %q", card.Question)
				}
				if !strings.Contains(card.Answer, "4") {
					t.Errorf("Expected answer to contain '4', got %q", card.Answer)
				}
				// Should be a new card since no FSRS metadata
				if card.FSRSCard.State != fsrs.New {
					t.Errorf("Expected new card state, got %v", card.FSRSCard.State)
				}
			},
		},
		{
			name: "card with FSRS metadata",
			content: `<!-- FSRS: due:2024-01-15T10:30:00Z, stability:2.50, difficulty:6.25, reps:5, state:Review -->
# Question
What is the capital of France?

---

# Answer
Paris`,
			expectedCard: func(t *testing.T, card *Card) {
				if !strings.Contains(card.Question, "What is the capital of France?") {
					t.Errorf("Expected question to contain 'What is the capital of France?', got %q", card.Question)
				}
				if !strings.Contains(card.Answer, "Paris") {
					t.Errorf("Expected answer to contain 'Paris', got %q", card.Answer)
				}
				if card.FSRSCard.State != fsrs.Review {
					t.Errorf("Expected Review state, got %v", card.FSRSCard.State)
				}
				if card.FSRSCard.Stability != 2.50 {
					t.Errorf("Expected stability 2.50, got %f", card.FSRSCard.Stability)
				}
				if card.FSRSCard.Reps != 5 {
					t.Errorf("Expected reps 5, got %d", card.FSRSCard.Reps)
				}
			},
		},
		{
			name: "card with multiple separators",
			content: `# Question
First part
---
Middle part (should be in answer)
---
Final part`,
			expectedCard: func(t *testing.T, card *Card) {
				if !strings.Contains(card.Question, "First part") {
					t.Errorf("Expected question to contain 'First part', got %q", card.Question)
				}
				// Everything after first --- should be in answer
				if !strings.Contains(card.Answer, "Middle part") || !strings.Contains(card.Answer, "Final part") {
					t.Errorf("Expected answer to contain both middle and final parts, got %q", card.Answer)
				}
			},
		},
		{
			name: "card without separator",
			content: `# Question only
This card has no answer section`,
			expectedCard: func(t *testing.T, card *Card) {
				if !strings.Contains(card.Question, "This card has no answer section") {
					t.Errorf("Expected question to contain content, got %q", card.Question)
				}
				if card.Answer != "" {
					t.Errorf("Expected empty answer, got %q", card.Answer)
				}
			},
		},
		{
			name: "empty card",
			content: ``,
			expectedCard: func(t *testing.T, card *Card) {
				if card.Question != "" {
					t.Errorf("Expected empty question, got %q", card.Question)
				}
				if card.Answer != "" {
					t.Errorf("Expected empty answer, got %q", card.Answer)
				}
			},
		},
		{
			name: "card with only separator",
			content: `---`,
			expectedCard: func(t *testing.T, card *Card) {
				if card.Question != "" {
					t.Errorf("Expected empty question, got %q", card.Question)
				}
				if card.Answer != "" {
					t.Errorf("Expected empty answer, got %q", card.Answer)
				}
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			filename := tt.name + ".md"
			filePath := createTempFile(t, tempDir, filename, tt.content)

			card, err := parseCard(filePath)
			
			if tt.expectError {
				if err == nil {
					t.Errorf("Expected error but got none")
				}
				return
			}

			if err != nil {
				t.Fatalf("Unexpected error: %v", err)
			}

			if card.FilePath != filePath {
				t.Errorf("Expected file path %q, got %q", filePath, card.FilePath)
			}

			tt.expectedCard(t, card)
		})
	}
}

func TestFindCards(t *testing.T) {
	tempDir := createTempDir(t)

	// Create test structure
	createTempFile(t, tempDir, "card1.md", "# Q\nQuestion 1\n---\n# A\nAnswer 1")
	createTempFile(t, tempDir, "card2.md", "# Q\nQuestion 2\n---\n# A\nAnswer 2")
	createTempFile(t, tempDir, "not_markdown.txt", "This should be ignored")
	createTempFile(t, tempDir, "subdir/card3.md", "# Q\nQuestion 3\n---\n# A\nAnswer 3")
	createTempFile(t, tempDir, "subdir/nested/card4.MD", "# Q\nQuestion 4\n---\n# A\nAnswer 4") // Test case insensitive

	cards, err := findCards(tempDir)
	if err != nil {
		t.Fatalf("Unexpected error: %v", err)
	}

	if len(cards) != 4 {
		t.Errorf("Expected 4 cards, got %d", len(cards))
	}

	// Verify all cards are markdown files
	for _, card := range cards {
		if !strings.HasSuffix(strings.ToLower(card.FilePath), ".md") {
			t.Errorf("Non-markdown file found: %s", card.FilePath)
		}
	}

	// Verify cards contain expected content
	questions := make(map[string]bool)
	for _, card := range cards {
		questions[card.Question] = true
	}

	expectedQuestions := []string{"# Q\nQuestion 1", "# Q\nQuestion 2", "# Q\nQuestion 3", "# Q\nQuestion 4"}
	for _, expected := range expectedQuestions {
		if !questions[expected] {
			t.Errorf("Expected question %q not found", expected)
		}
	}
}

func TestMarkdownFileDetection(t *testing.T) {
	tests := []struct {
		filename string
		expected bool
	}{
		{"card.md", true},
		{"card.MD", true},
		{"card.Md", true},
		{"card.mD", true},
		{"card.txt", false},
		{"card", false},
		{"card.markdown", false}, // Only .md extension supported
		{"README.md", true},
		{".md", true},
		{"file.md.backup", false},
		{"file.mdx", false},
	}

	for _, tt := range tests {
		t.Run(tt.filename, func(t *testing.T) {
			// Test the logic used in findCards for file detection
			isMarkdown := strings.HasSuffix(strings.ToLower(tt.filename), ".md")
			if isMarkdown != tt.expected {
				t.Errorf("File %q: expected markdown=%v, got %v", tt.filename, tt.expected, isMarkdown)
			}
		})
	}
}

func TestFindCardsEmptyDirectory(t *testing.T) {
	tempDir := createTempDir(t)

	cards, err := findCards(tempDir)
	if err != nil {
		t.Fatalf("Unexpected error: %v", err)
	}

	if len(cards) != 0 {
		t.Errorf("Expected 0 cards in empty directory, got %d", len(cards))
	}
}

func TestFindCardsNonexistentDirectory(t *testing.T) {
	nonexistentPath := "/this/path/should/not/exist"

	cards, err := findCards(nonexistentPath)
	if err == nil {
		t.Errorf("Expected error for nonexistent directory, but got none")
	}
	if cards != nil {
		t.Errorf("Expected nil cards for error case, got %v", cards)
	}
}