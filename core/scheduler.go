package core

import (
	"fmt"
	"time"

	"github.com/open-spaced-repetition/go-fsrs/v3"
)

// NewReviewSession creates a new review session with the given cards
func NewReviewSession(cards []*Card) *ReviewSession {
	params := fsrs.DefaultParam()
	return &ReviewSession{
		scheduler: fsrs.NewFSRS(params),
		cards:     cards,
		current:   0,
	}
}

// CurrentCard returns the current card in the session
func (rs *ReviewSession) CurrentCard() (*Card, error) {
	if rs.current >= len(rs.cards) {
		return nil, fmt.Errorf("no more cards in session")
	}
	return rs.cards[rs.current], nil
}

// HasNext returns whether there are more cards in the session
func (rs *ReviewSession) HasNext() bool {
	return rs.current < len(rs.cards)
}

// Progress returns current position and total cards in the session
func (rs *ReviewSession) Progress() (current, total int) {
	return rs.current + 1, len(rs.cards)
}

// RateCard rates the current card and updates its scheduling
func (rs *ReviewSession) RateCard(rating fsrs.Rating) error {
	if rs.current >= len(rs.cards) {
		return fmt.Errorf("no cards available to rate")
	}
	
	card := rs.cards[rs.current]
	now := time.Now()
	
	schedulingCards := rs.scheduler.Repeat(card.FSRSCard, now)
	selectedInfo := schedulingCards[rating]
	card.FSRSCard = selectedInfo.Card
	
	card.ReviewLog = append(card.ReviewLog, selectedInfo.ReviewLog)
	
	err := card.UpdateFSRSMetadata()
	if err != nil {
		return fmt.Errorf("failed to update card metadata: %v", err)
	}
	
	// Check all cards in the session to see if any have become due
	// and add them to the end of the queue if they're not already in the remaining cards
	remainingCards := rs.cards[rs.current+1:] // Cards we haven't reviewed yet
	
	for i := 0; i <= rs.current; i++ { // Check all cards we've seen so far
		checkCard := rs.cards[i]
		if checkCard.FSRSCard.Due.Before(now) || checkCard.FSRSCard.Due.Equal(now) {
			// Check if this card is already in the remaining queue
			alreadyQueued := false
			for _, remainingCard := range remainingCards {
				if remainingCard.FilePath == checkCard.FilePath {
					alreadyQueued = true
					break
				}
			}
			
			// If not already queued, add it to the end
			if !alreadyQueued {
				rs.cards = append(rs.cards, checkCard)
			}
		}
	}
	
	// Move to next card
	rs.current++
	
	return nil
}

// RatingFromInt converts integer rating (1-4) to FSRS Rating
func RatingFromInt(rating int) (fsrs.Rating, error) {
	switch rating {
	case 1:
		return fsrs.Again, nil
	case 2:
		return fsrs.Hard, nil
	case 3:
		return fsrs.Good, nil
	case 4:
		return fsrs.Easy, nil
	default:
		return fsrs.Again, fmt.Errorf("invalid rating %d: must be 1-4", rating)
	}
}

// UpdateCurrentCard updates the current card in the session (e.g., after editing)
func (rs *ReviewSession) UpdateCurrentCard(card *Card) {
	if rs.current < len(rs.cards) {
		rs.cards[rs.current] = card
	}
}

// RatingToString converts FSRS Rating to string
func RatingToString(rating fsrs.Rating) string {
	switch rating {
	case fsrs.Again:
		return "Again"
	case fsrs.Hard:
		return "Hard" 
	case fsrs.Good:
		return "Good"
	case fsrs.Easy:
		return "Easy"
	default:
		return "Again"
	}
}