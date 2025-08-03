package main

import (
	"bufio"
	"fmt"
	"os"
	"os/exec"
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
	
	fmt.Printf("\n[Press Enter to show answer...] ")
	
	reader := bufio.NewReader(os.Stdin)
	reader.ReadLine()
	
	fmt.Printf("\n")
	
	PrintMarkdown(card.Answer)
	
	fmt.Printf("\n")
	
	for {
		fmt.Printf("\n1=Again  2=Hard  3=Good  4=Easy  e=Edit  q=Quit\n> ")
		
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
					// Show the updated question and answer
					fmt.Printf("\n--- Card updated ---\n\n")
					PrintMarkdown(updatedCard.Question)
					fmt.Printf("\n")
					PrintMarkdown(updatedCard.Answer)
					fmt.Printf("\n")
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
	if len(rs.cards) == 0 {
		fmt.Println("No cards to review!")
		return nil
	}
	
	fmt.Printf("Starting review session with %d cards\n", len(rs.cards))
	
	for rs.current < len(rs.cards) {
		card := rs.cards[rs.current]
		
		err := rs.reviewCard(card)
		if err != nil {
			if err.Error() == "quit" {
				fmt.Printf("\nSession ended. Reviewed %d cards.\n", rs.current)
				return nil
			}
			return err
		}
		
		rs.current++
	}
	
	fmt.Printf("\nSession complete! Reviewed %d cards.\n", len(rs.cards))
	return nil
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