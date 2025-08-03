package main

import (
	"flag"
	"fmt"
	"os"
	"path/filepath"
	"strconv"
	"strings"
	"time"

	"github.com/open-spaced-repetition/go-fsrs/v3"
)

const usage = `srs - A Unix-style spaced repetition system

USAGE:
    srs [OPTIONS] COMMAND [ARGS...]

COMMANDS:
    review [DECK]       Start reviewing due cards (default: current directory)
    rate CARD RATING    Rate a specific card (1=Again, 2=Hard, 3=Good, 4=Easy)
    list [DECK]         List all cards in deck with due dates
    stats [DECK]        Show deck statistics
    due [DECK]          Show number of due cards

OPTIONS:
    -h, --help          Show this help message

EXAMPLES:
    srs review               # Review due cards in current directory
    srs review ./spanish     # Review cards in spanish directory  
    srs rate math/calc.md 3  # Rate a specific card as "Good"
    srs list                 # List all cards in current directory
    srs stats ./math         # Show statistics for math deck
    srs due                  # Show number of due cards

CARD FORMAT:
    Cards are markdown files with FSRS metadata:
    
    <!-- FSRS: due:2024-01-01T00:00:00Z, stability:1.00, difficulty:5.00, ... -->
    # Question
    What is 2 + 2?
    
    ---
    
    # Answer
    4
`

func main() {
	var help bool
	flag.BoolVar(&help, "h", false, "Show help")
	flag.BoolVar(&help, "help", false, "Show help")
	flag.Usage = func() {
		fmt.Print(usage)
	}
	flag.Parse()

	if help {
		flag.Usage()
		return
	}

	args := flag.Args()
	if len(args) == 0 {
		fmt.Fprintf(os.Stderr, "Error: No command specified\n\n")
		flag.Usage()
		os.Exit(1)
	}

	command := args[0]
	var deckPath string

	if len(args) > 1 {
		deckPath = args[1]
	} else {
		deckPath = "."
	}

	// Convert to absolute path
	absPath, err := filepath.Abs(deckPath)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error: Invalid path %s: %v\n", deckPath, err)
		os.Exit(1)
	}
	deckPath = absPath

	// Check if path exists
	if _, err := os.Stat(deckPath); os.IsNotExist(err) {
		fmt.Fprintf(os.Stderr, "Error: Path %s does not exist\n", deckPath)
		os.Exit(1)
	}

	switch command {
	case "review":
		err := reviewCommand(deckPath)
		if err != nil {
			fmt.Fprintf(os.Stderr, "Error: %v\n", err)
			os.Exit(1)
		}
	case "rate":
		if len(args) < 3 {
			fmt.Fprintf(os.Stderr, "Error: rate command requires card path and rating\nUsage: srs rate CARD RATING\n")
			os.Exit(1)
		}
		err := rateCommand(args[1], args[2])
		if err != nil {
			fmt.Fprintf(os.Stderr, "Error: %v\n", err)
			os.Exit(1)
		}
	case "list":
		err := listCommand(deckPath)
		if err != nil {
			fmt.Fprintf(os.Stderr, "Error: %v\n", err)
			os.Exit(1)
		}
	case "stats":
		err := statsCommand(deckPath)
		if err != nil {
			fmt.Fprintf(os.Stderr, "Error: %v\n", err)
			os.Exit(1)
		}
	case "due":
		err := dueCommand(deckPath)
		if err != nil {
			fmt.Fprintf(os.Stderr, "Error: %v\n", err)
			os.Exit(1)
		}
	default:
		fmt.Fprintf(os.Stderr, "Error: Unknown command '%s'\n\n", command)
		flag.Usage()
		os.Exit(1)
	}
}

func reviewCommand(deckPath string) error {
	cards, err := findCards(deckPath)
	if err != nil {
		return fmt.Errorf("failed to load cards: %v", err)
	}

	dueCards := getDueCards(cards)
	if len(dueCards) == 0 {
		fmt.Printf("No cards are due for review in %s\n", deckPath)
		return nil
	}

	session := NewReviewSession(dueCards)
	return session.Start()
}

func listCommand(deckPath string) error {
	cards, err := findCards(deckPath)
	if err != nil {
		return fmt.Errorf("failed to load cards: %v", err)
	}

	if len(cards) == 0 {
		fmt.Printf("No cards found in %s\n", deckPath)
		return nil
	}

	fmt.Printf("Cards in %s:\n\n", deckPath)
	for _, card := range cards {
		relPath, _ := filepath.Rel(deckPath, card.FilePath)
		due := "Due now"
		if card.FSRSCard.Due.After(time.Now()) {
			due = fmt.Sprintf("Due %s", card.FSRSCard.Due.Format("2006-01-02 15:04"))
		}
		
		questionPreview := strings.ReplaceAll(card.Question, "\n", " ")
		questionPreview = strings.TrimSpace(questionPreview)
		// Remove markdown formatting for preview
		questionPreview = strings.ReplaceAll(questionPreview, "#", "")
		questionPreview = strings.TrimSpace(questionPreview)
		if len(questionPreview) > 50 {
			questionPreview = questionPreview[:50] + "..."
		}
		
		fmt.Printf("%-30s %s\n", relPath, due)
		fmt.Printf("  %s\n\n", questionPreview)
	}

	return nil
}

func statsCommand(deckPath string) error {
	cards, err := findCards(deckPath)
	if err != nil {
		return fmt.Errorf("failed to load cards: %v", err)
	}

	if len(cards) == 0 {
		fmt.Printf("No cards found in %s\n", deckPath)
		return nil
	}

	total := len(cards)
	due := len(getDueCards(cards))
	new := 0
	learning := 0
	review := 0
	relearning := 0

	for _, card := range cards {
		switch StateToString(card.FSRSCard.State) {
		case "New":
			new++
		case "Learning":
			learning++
		case "Review":
			review++
		case "Relearning":
			relearning++
		}
	}

	fmt.Printf("Deck statistics for %s:\n\n", deckPath)
	fmt.Printf("Total cards:    %d\n", total)
	fmt.Printf("Due cards:      %d\n", due)
	fmt.Printf("New cards:      %d\n", new)
	fmt.Printf("Learning cards: %d\n", learning)
	fmt.Printf("Review cards:   %d\n", review)
	fmt.Printf("Relearning:     %d\n", relearning)

	return nil
}

func dueCommand(deckPath string) error {
	cards, err := findCards(deckPath)
	if err != nil {
		return fmt.Errorf("failed to load cards: %v", err)
	}

	dueCards := getDueCards(cards)
	fmt.Printf("%d\n", len(dueCards))
	return nil
}

func rateCommand(cardPath, ratingStr string) error {
	// Convert rating string to int
	rating, err := strconv.Atoi(ratingStr)
	if err != nil {
		return fmt.Errorf("invalid rating '%s': must be 1-4", ratingStr)
	}

	// Validate rating range
	if rating < 1 || rating > 4 {
		return fmt.Errorf("invalid rating %d: must be 1-4", rating)
	}

	// Convert to FSRS rating
	var fsrsRating fsrs.Rating
	switch rating {
	case 1:
		fsrsRating = fsrs.Again
	case 2:
		fsrsRating = fsrs.Hard
	case 3:
		fsrsRating = fsrs.Good
	case 4:
		fsrsRating = fsrs.Easy
	}

	// Get absolute path
	absPath, err := filepath.Abs(cardPath)
	if err != nil {
		return fmt.Errorf("invalid path %s: %v", cardPath, err)
	}

	// Check if file exists
	if _, err := os.Stat(absPath); os.IsNotExist(err) {
		return fmt.Errorf("card file %s does not exist", absPath)
	}

	// Parse the card
	card, err := parseCard(absPath)
	if err != nil {
		return fmt.Errorf("failed to parse card %s: %v", absPath, err)
	}

	// Create a scheduler
	params := fsrs.DefaultParam()
	scheduler := fsrs.NewFSRS(params)

	// Apply the rating
	now := time.Now()
	schedulingCards := scheduler.Repeat(card.FSRSCard, now)
	selectedInfo := schedulingCards[fsrsRating]
	
	card.FSRSCard = selectedInfo.Card
	card.ReviewLog = append(card.ReviewLog, selectedInfo.ReviewLog)

	// Update the card file
	err = card.updateFSRSMetadata()
	if err != nil {
		return fmt.Errorf("failed to update card metadata: %v", err)
	}

	fmt.Printf("Card rated as %s. Next due: %s\n", 
		map[int]string{1: "Again", 2: "Hard", 3: "Good", 4: "Easy"}[rating],
		card.FSRSCard.Due.Format("2006-01-02 15:04"))

	return nil
}