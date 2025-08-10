package main

import (
	"bufio"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strconv"
	"strings"
	"time"

	"github.com/open-spaced-repetition/go-fsrs/v3"
)

type ReviewSession struct {
	scheduler *fsrs.FSRS
	cards     []*Card
	current   int
}

func NewReviewSession(cards []*Card) *ReviewSession {
	params := fsrs.DefaultParam()
	return &ReviewSession{
		scheduler: fsrs.NewFSRS(params),
		cards:     cards,
		current:   0,
	}
}

func (rs *ReviewSession) reviewCard(card *Card) error {
	fmt.Printf("\n")
	
	PrintMarkdown(card.Question)
	
	reader := bufio.NewReader(os.Stdin)
	var userAnswer string
	
	// Ask if user wants to type an answer
	fmt.Printf("\nType your answer? (y/n or just press Enter to skip): ")
	input, _ := reader.ReadString('\n')
	input = strings.TrimSpace(strings.ToLower(input))
	
	if input == "y" || input == "yes" {
		fmt.Printf("\nYour answer:\n> ")
		userAnswer, _ = reader.ReadString('\n')
		userAnswer = strings.TrimSpace(userAnswer)
		fmt.Printf("\n[Press Enter to show correct answer...]")
		reader.ReadLine()
	} else {
		fmt.Printf("\n[Press Enter to show answer...]")
		reader.ReadLine()
	}
	
	fmt.Printf("\n")
	
	// Show the correct answer
	PrintMarkdown(card.Answer)
	
	fmt.Printf("\n")
	
	// If user typed an answer, show it for comparison
	if userAnswer != "" {
		fmt.Printf("--- Your answer ---\n%s\n\n", userAnswer)
	}
	
	for {
		fmt.Printf("1=Again  2=Hard  3=Good  4=Easy  e=Edit  q=Quit\n> ")
		
		input, _ := reader.ReadString('\n')
		input = strings.TrimSpace(input)
		
		switch input {
		case "1":
			return rs.updateCard(card, fsrs.Again)
		case "2":
			return rs.updateCard(card, fsrs.Hard)
		case "3":
			return rs.updateCard(card, fsrs.Good)
		case "4":
			return rs.updateCard(card, fsrs.Easy)
		case "e", "E":
			// Count lines displayed so far for clearing
			linesDisplayed := rs.countDisplayedLines(card, userAnswer)
			
			err := editCard(card)
			if err != nil {
				fmt.Printf("Error editing card: %v\n", err)
			} else {
				// Reload the card after editing
				updatedCard, err := parseCard(card.FilePath)
				if err != nil {
					fmt.Printf("Error reloading card: %v\n", err)
				} else {
					// Update the card in the session
					rs.cards[rs.current] = updatedCard
					
					// Clear the previous card display and redraw
					rs.clearAndRedrawCard(updatedCard, userAnswer, linesDisplayed)
				}
			}
		case "q":
			return fmt.Errorf("quit")
		default:
			fmt.Printf("Invalid choice.\n")
		}
	}
}

func (rs *ReviewSession) updateCard(card *Card, rating fsrs.Rating) error {
	now := time.Now()
	
	schedulingCards := rs.scheduler.Repeat(card.FSRSCard, now)
	
	selectedInfo := schedulingCards[rating]
	card.FSRSCard = selectedInfo.Card
	
	card.ReviewLog = append(card.ReviewLog, selectedInfo.ReviewLog)
	
	return card.updateFSRSMetadata()
}

func (rs *ReviewSession) Start() error {
	// Use TUI for review sessions
	return rs.StartTUI()
}

func (rs *ReviewSession) StartTurnBased(rating string) error {
	// If rating is provided, rate the current card first
	if rating != "" {
		if rs.current >= len(rs.cards) {
			fmt.Println("No cards available to rate.")
			return nil
		}
		
		// Parse and validate rating
		ratingInt, err := strconv.Atoi(rating)
		if err != nil || ratingInt < 1 || ratingInt > 4 {
			return fmt.Errorf("invalid rating '%s': must be 1-4", rating)
		}
		
		// Convert to FSRS rating
		var fsrsRating fsrs.Rating
		switch ratingInt {
		case 1:
			fsrsRating = fsrs.Again
		case 2:
			fsrsRating = fsrs.Hard
		case 3:
			fsrsRating = fsrs.Good
		case 4:
			fsrsRating = fsrs.Easy
		}
		
		// Rate the current card
		currentCard := rs.cards[rs.current]
		err = rs.updateCard(currentCard, fsrsRating)
		if err != nil {
			return fmt.Errorf("failed to rate card: %v", err)
		}
		
		fmt.Printf("Card rated as %s.\n", 
			map[int]string{1: "Again", 2: "Hard", 3: "Good", 4: "Easy"}[ratingInt])
		
		// Check all cards in the session to see if any have become due
		// and add them to the end of the queue if they're not already in the remaining cards
		now := time.Now()
		remainingCards := rs.cards[rs.current+1:] // Cards we haven't reviewed yet
		
		for i := 0; i <= rs.current; i++ { // Check all cards we've seen so far
			card := rs.cards[i]
			if card.FSRSCard.Due.Before(now) || card.FSRSCard.Due.Equal(now) {
				// Check if this card is already in the remaining queue
				alreadyQueued := false
				for _, remainingCard := range remainingCards {
					if remainingCard.FilePath == card.FilePath {
						alreadyQueued = true
						break
					}
				}
				
				// If not already queued, add it to the end
				if !alreadyQueued {
					rs.cards = append(rs.cards, card)
				}
			}
		}
		
		// Move to next card
		rs.current++
	}
	
	// Show the next due card
	if rs.current >= len(rs.cards) {
		fmt.Println("No more cards due for review!")
		return nil
	}
	
	card := rs.cards[rs.current]
	
	// Display the card
	fmt.Printf("\nCard %d of %d:\n\n", rs.current+1, len(rs.cards))
	PrintMarkdown(card.Question)
	fmt.Printf("\n---\n\n")
	PrintMarkdown(card.Answer)
	
	// Show rating command
	deckPathFromCard := strings.TrimSuffix(card.FilePath, filepath.Base(card.FilePath))
	if deckPathFromCard != "" {
		deckPathFromCard = strings.TrimSuffix(deckPathFromCard, "/")
		// Extract just the subdeck name relative to base deck
		config, _ := loadConfig()
		if config != nil && config.BaseDeckPath != "" {
			if rel, err := filepath.Rel(config.BaseDeckPath, deckPathFromCard); err == nil && rel != "." {
				deckPathFromCard = rel
			} else {
				deckPathFromCard = ""
			}
		} else {
			deckPathFromCard = ""
		}
	}
	
	if deckPathFromCard != "" {
		fmt.Printf("\nTo rate: srs review %s [1-4]\n", deckPathFromCard)
	} else {
		fmt.Printf("\nTo rate: srs review [1-4]\n")
	}
	fmt.Printf("1=Again  2=Hard  3=Good  4=Easy\n")
	
	return nil
}

func (rs *ReviewSession) countDisplayedLines(card *Card, userAnswer string) int {
	lines := 0
	
	// Count question lines (rough estimate)
	lines += strings.Count(card.Question, "\n") + 2 // +2 for extra spacing
	
	// Count answer lines
	lines += strings.Count(card.Answer, "\n") + 2 // +2 for extra spacing
	
	// Count user answer lines if present
	if userAnswer != "" {
		lines += strings.Count(userAnswer, "\n") + 3 // +3 for header and spacing
	}
	
	// Add lines for the rating prompt
	lines += 2
	
	return lines
}

func (rs *ReviewSession) clearAndRedrawCard(card *Card, userAnswer string, linesToClear int) {
	// Move cursor up and clear lines
	for i := 0; i < linesToClear; i++ {
		fmt.Printf("\033[1A\033[K") // Move up one line and clear it
	}
	
	// Redraw the card content
	fmt.Printf("\n")
	PrintMarkdown(card.Question)
	fmt.Printf("\n")
	PrintMarkdown(card.Answer)
	fmt.Printf("\n")
	
	// Show user's answer again if they had one
	if userAnswer != "" {
		fmt.Printf("--- Your answer ---\n%s\n\n", userAnswer)
	}
}

func getDueCards(cards []*Card) []*Card {
	now := time.Now()
	var dueCards []*Card
	
	for _, card := range cards {
		if card.FSRSCard.Due.Before(now) || card.FSRSCard.Due.Equal(now) {
			dueCards = append(dueCards, card)
		}
	}
	
	return dueCards
}

func editCard(card *Card) error {
	// Determine the editor to use
	editor := os.Getenv("EDITOR")
	if editor == "" {
		editor = os.Getenv("VISUAL")
	}
	if editor == "" {
		// Default fallbacks in order of preference
		editors := []string{"vim", "vi", "nano", "emacs"}
		for _, e := range editors {
			if _, err := exec.LookPath(e); err == nil {
				editor = e
				break
			}
		}
	}
	if editor == "" {
		return fmt.Errorf("no editor found. Please set EDITOR or VISUAL environment variable")
	}

	// Create the command
	cmd := exec.Command(editor, card.FilePath)
	cmd.Stdin = os.Stdin
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr

	// Run the editor
	return cmd.Run()
}