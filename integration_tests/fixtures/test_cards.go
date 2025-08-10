package fixtures

import (
	"fmt"
	"os"
	"path/filepath"
	"time"
)

// TestDeck represents a test deck structure
type TestDeck struct {
	Name  string
	Cards []TestCard
}

// TestCard represents a test card
type TestCard struct {
	Filename string
	Question string
	Answer   string
	WithFSRS bool // Whether to include FSRS metadata
}

// BasicMathDeck returns a simple math deck for testing
func BasicMathDeck() TestDeck {
	return TestDeck{
		Name: "basic_math",
		Cards: []TestCard{
			{
				Filename: "addition.md",
				Question: "What is 2 + 2?",
				Answer:   "4",
				WithFSRS: false,
			},
			{
				Filename: "multiplication.md", 
				Question: "What is 3 × 4?",
				Answer:   "12",
				WithFSRS: false,
			},
			{
				Filename: "division.md",
				Question: "What is 15 ÷ 3?",
				Answer:   "5",
				WithFSRS: false,
			},
		},
	}
}

// ProgrammingDeck returns a programming concepts deck
func ProgrammingDeck() TestDeck {
	return TestDeck{
		Name: "programming",
		Cards: []TestCard{
			{
				Filename: "time_complexity.md",
				Question: "What is the time complexity of binary search?",
				Answer:   "O(log n) - because we eliminate half the search space with each comparison.",
				WithFSRS: false,
			},
			{
				Filename: "sorting.md",
				Question: "What sorting algorithm has O(n log n) average case?",
				Answer:   "Merge sort, heap sort, and quicksort (average case).",
				WithFSRS: false,
			},
		},
	}
}

// ReviewedCardsDeck returns cards with FSRS metadata (as if they've been reviewed)
func ReviewedCardsDeck() TestDeck {
	return TestDeck{
		Name: "reviewed",
		Cards: []TestCard{
			{
				Filename: "easy_card.md",
				Question: "What is 1 + 1?",
				Answer:   "2",
				WithFSRS: true,
			},
			{
				Filename: "hard_card.md",
				Question: "Explain the CAP theorem in distributed systems.",
				Answer:   "Consistency, Availability, and Partition tolerance - you can only guarantee two out of three.",
				WithFSRS: true,
			},
		},
	}
}

// CreateDeck creates a test deck on the filesystem
func (td TestDeck) CreateDeck(basePath string) error {
	deckPath := filepath.Join(basePath, td.Name)
	if err := os.MkdirAll(deckPath, 0755); err != nil {
		return fmt.Errorf("failed to create deck directory: %v", err)
	}

	for _, card := range td.Cards {
		cardPath := filepath.Join(deckPath, card.Filename)
		content := fmt.Sprintf("%s\n---\n%s", card.Question, card.Answer)
		
		if card.WithFSRS {
			// Add sample FSRS metadata
			due := time.Now().Add(24 * time.Hour).Format("2006-01-02T15:04:05Z")
			fsrsLine := fmt.Sprintf("<!-- FSRS: due:%s, stability:2.50, difficulty:5.00, elapsed_days:1, scheduled_days:1, reps:1, lapses:0, state:Review -->\n\n", due)
			content = fsrsLine + content
		}

		if err := os.WriteFile(cardPath, []byte(content), 0644); err != nil {
			return fmt.Errorf("failed to write card %s: %v", cardPath, err)
		}
	}

	return nil
}

// CleanupDeck removes a test deck from the filesystem
func (td TestDeck) CleanupDeck(basePath string) error {
	deckPath := filepath.Join(basePath, td.Name)
	return os.RemoveAll(deckPath)
}