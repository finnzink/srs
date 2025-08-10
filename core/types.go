package core

import (
	"time"

	"github.com/open-spaced-repetition/go-fsrs/v3"
)

// Card represents a flashcard with its content and scheduling information
type Card struct {
	Question     string
	Answer       string
	FilePath     string
	FSRSCard     fsrs.Card
	ReviewLog    []fsrs.ReviewLog
	LastModified time.Time
}

// DeckStats contains statistics about a deck
type DeckStats struct {
	TotalCards   int
	DueCards     int
	NewCards     int
	LearningCards int
	ReviewCards  int
}

// ReviewSession manages a review session for multiple cards
type ReviewSession struct {
	scheduler *fsrs.FSRS
	cards     []*Card
	current   int
}

// Config holds application configuration
type Config struct {
	BaseDeckPath string `json:"base_deck_path"`
}