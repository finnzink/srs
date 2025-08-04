package main

import (
	"flag"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strconv"
	"time"

	"github.com/open-spaced-repetition/go-fsrs/v3"
)

const usage = `srs - A Unix-style spaced repetition system

USAGE:
    srs [OPTIONS] COMMAND [ARGS...]

COMMANDS:
    review [SUBDECK]    Start reviewing due cards from base deck or subdirectory
    rate CARD RATING    Rate a specific card (1=Again, 2=Hard, 3=Good, 4=Easy)
    list [SUBDECK]      Show deck tree with due dates and cards
    stats [SUBDECK]     Show deck statistics
    due [SUBDECK]       Show number of due cards
    config              Set up base deck directory
    update              Update to the latest version
    version             Show version information

OPTIONS:
    -h, --help          Show this help message
    -v, --version       Show version information

EXAMPLES:
    srs config                    # Set up your base deck directory
    srs review                    # Review all cards from base deck
    srs review spanish            # Review cards from spanish subdirectory
    srs review spanish/grammar    # Review from nested subdirectories
    srs list                      # Show tree for entire base deck
    srs list spanish              # Show tree for spanish subdirectory
    srs stats spanish/grammar     # Show statistics for grammar subdirectory
    srs due                       # Show due cards count for entire deck
    srs rate spanish/verb.md 3    # Rate a specific card as "Good"

CARD FORMAT:
    Cards are markdown files:
    
    What is the capital of Paris?
    ---
    France

Guidelines for creating excellent flashcards:
• Be EXTREMELY concise - answers should be 1-2 sentences maximum!
• Focus on core concepts, relationships, and techniques rather than trivia or isolated facts
• Break complex ideas into smaller, atomic concepts
• Ensure each card tests one specific idea (atomic)
• Front of card should ask a specific question that prompts recall
• Back of card should provide the shortest possible complete answer
• CRITICAL: Keep answers as brief as possible while maintaining accuracy - aim for 10-25 words max
• When referencing the author or source, use their specific name rather than general phrases like "the author" or "this text" which won't make sense months later when the user is reviewing the cards
• Try to cite the author or the source when discussing something that is not an established concept but rather a new take or theory or prediction. 
• The questions should be precise and unambiguously exclude alternative correct answers
• The questions should encode ideas from multiple angles
• Avoid yes/no question, or, in general, questions that admit a binary answer
• Avoid unordered lists of items (especially if they contain many items)
• If quantities are involved, they should be relative, or the unit of measure should be specified in the question
`

func main() {
	var help, version bool
	flag.BoolVar(&help, "h", false, "Show help")
	flag.BoolVar(&help, "help", false, "Show help")
	flag.BoolVar(&version, "v", false, "Show version")
	flag.BoolVar(&version, "version", false, "Show version")
	flag.Usage = func() {
		fmt.Print(usage)
		
		// Try to show current deck structure if configured
		config, err := loadConfig()
		if err == nil && config.BaseDeckPath != "" {
			fmt.Printf("\nCURRENT DECK:\n")
			err := statusCommand(config.BaseDeckPath)
			if err != nil {
				fmt.Printf("(Unable to load deck: %v)\n", err)
			}
		}
	}
	flag.Parse()

	if help {
		flag.Usage()
		return
	}

	if version {
		printVersion()
		return
	}

	args := flag.Args()
	if len(args) == 0 {
		fmt.Fprintf(os.Stderr, "Error: No command specified\n\n")
		flag.Usage()
		os.Exit(1)
	}

	command := args[0]
	
	// Load config
	config, err := loadConfig()
	if err != nil {
		fmt.Fprintf(os.Stderr, "Warning: Failed to load config: %v\n", err)
		config = &Config{}
	}
	
	// Check if this is first run (no base deck configured) and command needs it
	if config.BaseDeckPath == "" && command != "config" && command != "version" && command != "update" {
		fmt.Println("No base deck configured. Let's set one up first!")
		err := promptForBaseDeck()
		if err != nil {
			fmt.Fprintf(os.Stderr, "Error setting up base deck: %v\n", err)
			os.Exit(1)
		}
		// Reload config after setup
		config, err = loadConfig()
		if err != nil {
			fmt.Fprintf(os.Stderr, "Error reloading config: %v\n", err)
			os.Exit(1)
		}
	}
	
	var deckPath string
	if len(args) > 1 {
		deckPath = args[1]
	} else {
		deckPath = "."
	}

	// Resolve deck path using config (unless it's a command that doesn't need a deck)
	if command != "config" && command != "version" && command != "update" {
		resolvedPath, err := resolveDeckPath(deckPath, config)
		if err != nil {
			fmt.Fprintf(os.Stderr, "Error: Invalid path %s: %v\n", deckPath, err)
			os.Exit(1)
		}
		deckPath = resolvedPath

		// Check if path exists
		if _, err := os.Stat(deckPath); os.IsNotExist(err) {
			fmt.Fprintf(os.Stderr, "Error: Path %s does not exist\n", deckPath)
			os.Exit(1)
		}
	}

	switch command {
	case "review":
		// Check for updates before starting review (non-blocking)
		go checkForUpdates()
		
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
		err := statusCommand(deckPath)
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
	case "config":
		err := promptForBaseDeck()
		if err != nil {
			fmt.Fprintf(os.Stderr, "Error: %v\n", err)
			os.Exit(1)
		}
	case "version":
		printVersion()
	case "update":
		err := updateCommand()
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

func updateCommand() error {
	fmt.Println("Updating SRS to the latest version...")
	
	// Download and run the install script
	cmd := exec.Command("bash", "-c", 
		"curl -sSL https://raw.githubusercontent.com/finnzink/srs/main/install.sh | bash")
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	
	err := cmd.Run()
	if err != nil {
		return fmt.Errorf("update failed: %v", err)
	}
	
	fmt.Println("✅ Update completed successfully!")
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