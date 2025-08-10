package tui

import (
	"fmt"
	"os"
	"os/exec"
	"strings"

	tea "github.com/charmbracelet/bubbletea"

	"srs/core"
)

// StartTUI starts the TUI review session
func StartTUI(cards []*core.Card) error {
	if len(cards) == 0 {
		fmt.Println("No cards to review!")
		return nil
	}

	session := core.NewReviewSession(cards)

	for {
		model, err := NewReviewModel(session)
		if err != nil {
			return fmt.Errorf("failed to create review model: %v", err)
		}
		
		program := tea.NewProgram(model, tea.WithAltScreen())
		
		finalModel, err := program.Run()
		if err != nil {
			return fmt.Errorf("TUI error: %v", err)
		}

		final := finalModel.(ReviewModel)
		
		// Check if user wanted to edit
		if strings.HasPrefix(final.message, "edit_card:") {
			// Parse the state information
			parts := strings.Split(final.message, ":")
			var savedUserAnswer string
			var savedState reviewState
			if len(parts) >= 3 {
				savedUserAnswer = parts[1]
				if stateInt := 0; len(parts[2]) > 0 {
					fmt.Sscanf(parts[2], "%d", &stateInt)
					savedState = reviewState(stateInt)
				}
			}
			
			// Edit the current card
			err := editCard(final.currentCard)
			if err != nil {
				fmt.Printf("Error editing card: %v\n", err)
				return nil
			}
			
			// Reload the card
			updatedCard, err := core.ParseCard(final.currentCard.FilePath)
			if err != nil {
				fmt.Printf("Error reloading card: %v\n", err)
				return nil
			}
			
			// Update the session
			final.session.UpdateCurrentCard(updatedCard)
			
			// Create new model with restored state
			model, err := NewReviewModel(final.session)
			if err != nil {
				return fmt.Errorf("failed to create review model after edit: %v", err)
			}
			model.userAnswer = savedUserAnswer
			model.state = savedState
			
			// Continue with restored state
			program := tea.NewProgram(model, tea.WithAltScreen())
			finalModel, err := program.Run()
			if err != nil {
				return fmt.Errorf("TUI error: %v", err)
			}
			
			// Update the final model for the next iteration
			final = finalModel.(ReviewModel)
			continue
		}
		
		// Check if we completed the session or user quit normally
		current, total := final.session.Progress()
		if !final.session.HasNext() {
			fmt.Printf("Session complete! Reviewed %d cards.\n", total)
		} else {
			fmt.Printf("Session ended. Reviewed %d cards.\n", current-1)
		}
		
		break
	}
	return nil
}

func editCard(card *core.Card) error {
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