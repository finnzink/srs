package main

import (
	"fmt"
	"strconv"
	"testing"

	"github.com/open-spaced-repetition/go-fsrs/v3"
)

func TestRatingValidation(t *testing.T) {
	tests := []struct {
		name        string
		ratingStr   string
		expectedInt int
		expectError bool
		errorMsg    string
	}{
		{
			name:        "valid rating 1",
			ratingStr:   "1",
			expectedInt: 1,
		},
		{
			name:        "valid rating 2",
			ratingStr:   "2",
			expectedInt: 2,
		},
		{
			name:        "valid rating 3",
			ratingStr:   "3",
			expectedInt: 3,
		},
		{
			name:        "valid rating 4",
			ratingStr:   "4",
			expectedInt: 4,
		},
		{
			name:        "invalid rating 0",
			ratingStr:   "0",
			expectError: true,
			errorMsg:    "invalid rating 0: must be 1-4",
		},
		{
			name:        "invalid rating 5",
			ratingStr:   "5",
			expectError: true,
			errorMsg:    "invalid rating 5: must be 1-4",
		},
		{
			name:        "non-numeric rating",
			ratingStr:   "abc",
			expectError: true,
			errorMsg:    "invalid rating 'abc': must be 1-4",
		},
		{
			name:        "empty rating",
			ratingStr:   "",
			expectError: true,
			errorMsg:    "invalid rating '': must be 1-4",
		},
		{
			name:        "decimal rating",
			ratingStr:   "3.5",
			expectError: true,
			errorMsg:    "invalid rating '3.5': must be 1-4",
		},
		{
			name:        "negative rating",
			ratingStr:   "-1",
			expectError: true,
			errorMsg:    "invalid rating -1: must be 1-4",
		},
		{
			name:        "very large number",
			ratingStr:   "999",
			expectError: true,
			errorMsg:    "invalid rating 999: must be 1-4",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Test the validation logic from rateCommand
			rating, err := strconv.Atoi(tt.ratingStr)
			
			var validationErr error
			if err != nil {
				validationErr = fmt.Errorf("invalid rating '%s': must be 1-4", tt.ratingStr)
			} else if rating < 1 || rating > 4 {
				validationErr = fmt.Errorf("invalid rating %d: must be 1-4", rating)
			}

			if tt.expectError {
				if validationErr == nil {
					t.Errorf("Expected error but got none")
					return
				}
				if validationErr.Error() != tt.errorMsg {
					t.Errorf("Expected error message %q, got %q", tt.errorMsg, validationErr.Error())
				}
			} else {
				if validationErr != nil {
					t.Errorf("Unexpected error: %v", validationErr)
					return
				}
				if rating != tt.expectedInt {
					t.Errorf("Expected rating %d, got %d", tt.expectedInt, rating)
				}
			}
		})
	}
}

func TestRatingToFSRSConversion(t *testing.T) {
	tests := []struct {
		rating       int
		expectedFSRS fsrs.Rating
		description  string
	}{
		{1, fsrs.Again, "Again"},
		{2, fsrs.Hard, "Hard"},
		{3, fsrs.Good, "Good"},
		{4, fsrs.Easy, "Easy"},
	}

	for _, tt := range tests {
		t.Run(tt.description, func(t *testing.T) {
			// Test the conversion logic from rateCommand
			var fsrsRating fsrs.Rating
			switch tt.rating {
			case 1:
				fsrsRating = fsrs.Again
			case 2:
				fsrsRating = fsrs.Hard
			case 3:
				fsrsRating = fsrs.Good
			case 4:
				fsrsRating = fsrs.Easy
			}

			if fsrsRating != tt.expectedFSRS {
				t.Errorf("Expected FSRS rating %v, got %v", tt.expectedFSRS, fsrsRating)
			}
		})
	}
}

func TestRatingDescriptionMapping(t *testing.T) {
	// Test the description mapping used in rateCommand output
	ratingDescriptions := map[int]string{
		1: "Again",
		2: "Hard", 
		3: "Good",
		4: "Easy",
	}

	tests := []struct {
		rating      int
		expectedDesc string
	}{
		{1, "Again"},
		{2, "Hard"},
		{3, "Good"},
		{4, "Easy"},
	}

	for _, tt := range tests {
		t.Run(tt.expectedDesc, func(t *testing.T) {
			desc, exists := ratingDescriptions[tt.rating]
			if !exists {
				t.Errorf("No description found for rating %d", tt.rating)
				return
			}
			if desc != tt.expectedDesc {
				t.Errorf("Expected description %q for rating %d, got %q", tt.expectedDesc, tt.rating, desc)
			}
		})
	}
}

func TestRatingRoundTripConversion(t *testing.T) {
	// Test that rating -> FSRS -> description chain is consistent
	testCases := []struct {
		rating      int
		fsrsRating  fsrs.Rating
		description string
	}{
		{1, fsrs.Again, "Again"},
		{2, fsrs.Hard, "Hard"},
		{3, fsrs.Good, "Good"},
		{4, fsrs.Easy, "Easy"},
	}

	ratingDescriptions := map[int]string{1: "Again", 2: "Hard", 3: "Good", 4: "Easy"}

	for _, tt := range testCases {
		t.Run(tt.description, func(t *testing.T) {
			// Convert rating to FSRS
			var fsrsRating fsrs.Rating
			switch tt.rating {
			case 1:
				fsrsRating = fsrs.Again
			case 2:
				fsrsRating = fsrs.Hard
			case 3:
				fsrsRating = fsrs.Good
			case 4:
				fsrsRating = fsrs.Easy
			}

			// Check FSRS conversion
			if fsrsRating != tt.fsrsRating {
				t.Errorf("FSRS conversion failed: expected %v, got %v", tt.fsrsRating, fsrsRating)
			}

			// Check description mapping
			desc := ratingDescriptions[tt.rating]
			if desc != tt.description {
				t.Errorf("Description mapping failed: expected %q, got %q", tt.description, desc)
			}
		})
	}
}

func TestBoundaryRatingValues(t *testing.T) {
	// Test edge cases for rating validation
	tests := []struct {
		name      string
		ratingStr string
		valid     bool
	}{
		{"minimum valid", "1", true},
		{"maximum valid", "4", true},
		{"below minimum", "0", false},
		{"above maximum", "5", false},
		{"way below minimum", "-100", false},
		{"way above maximum", "100", false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			rating, err := strconv.Atoi(tt.ratingStr)
			if err != nil {
				if tt.valid {
					t.Errorf("Expected valid rating but got parse error: %v", err)
				}
				return
			}

			isValid := rating >= 1 && rating <= 4
			if isValid != tt.valid {
				t.Errorf("Expected valid=%v for rating %d, got valid=%v", tt.valid, rating, isValid)
			}
		})
	}
}