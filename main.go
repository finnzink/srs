package main

import (
	"flag"
	"fmt"
	"os"
	"os/exec"
)

const usage = `srs - A Unix-style spaced repetition system

USAGE:
    srs [OPTIONS] COMMAND [ARGS...]

COMMANDS:
    review [SUBDECK] [RATING]  Show next card (turn-based) or rate current card
    list [SUBDECK]             Show deck tree with due dates and stats
    config                     Set up base deck directory
    update                     Update to the latest version
    version                    Show version information

OPTIONS:
    -i, --interactive          Use interactive TUI mode for review
    -h, --help                 Show this help message
    -v, --version              Show version information

EXAMPLES:
    srs config                 # Set up your base deck directory
    srs review                 # Show next due card (turn-based)
    srs review spanish         # Show next due card from spanish subdirectory
    srs review spanish 3       # Rate current card as "Good" and show next
    srs review -i              # Start interactive TUI review mode
    srs list                   # Show tree with due dates and deck stats
    srs list spanish           # Show tree for spanish subdirectory

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
	var help, version, interactive bool
	flag.BoolVar(&help, "h", false, "Show help")
	flag.BoolVar(&help, "help", false, "Show help")
	flag.BoolVar(&version, "v", false, "Show version")
	flag.BoolVar(&version, "version", false, "Show version")
	flag.BoolVar(&interactive, "i", false, "Use interactive TUI mode for review")
	flag.BoolVar(&interactive, "interactive", false, "Use interactive TUI mode for review")
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
	var rating string
	
	// Special handling for review command to distinguish between subdeck and rating
	if command == "review" {
		if len(args) == 3 {
			// srs review [subdeck] [rating]
			deckPath = args[1]
			rating = args[2]
		} else if len(args) == 2 {
			// Check if second arg is a rating (1-4) or a subdeck path
			if args[1] == "1" || args[1] == "2" || args[1] == "3" || args[1] == "4" {
				// srs review [rating] - no subdeck specified
				deckPath = "."
				rating = args[1]
			} else {
				// srs review [subdeck] - no rating specified
				deckPath = args[1]
			}
		} else {
			// srs review - no subdeck or rating specified
			deckPath = "."
		}
	} else {
		// For other commands, use normal parsing
		if len(args) > 1 {
			deckPath = args[1]
		} else {
			deckPath = "."
		}
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
		
		err := reviewCommand(deckPath, rating, interactive)
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

func reviewCommand(deckPath, rating string, interactive bool) error {
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
	
	if interactive {
		// Use TUI mode
		return session.Start()
	}
	
	// Turn-based mode
	return session.StartTurnBased(rating)
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

