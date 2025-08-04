package main

import (
	"fmt"
	"path/filepath"
	"sort"
	"strings"
	"time"
)

type DeckNode struct {
	Name     string
	Path     string
	IsDir    bool
	Cards    []*Card
	Children []*DeckNode
	Parent   *DeckNode
}

func buildDeckTree(deckPath string) (*DeckNode, error) {
	// Get all cards in the deck
	cards, err := findCards(deckPath)
	if err != nil {
		return nil, err
	}

	// Create root node
	root := &DeckNode{
		Name:     filepath.Base(deckPath),
		Path:     deckPath,
		IsDir:    true,
		Children: make([]*DeckNode, 0),
	}

	// Group cards by directory
	dirCards := make(map[string][]*Card)
	for _, card := range cards {
		relPath, err := filepath.Rel(deckPath, card.FilePath)
		if err != nil {
			continue
		}
		
		dir := filepath.Dir(relPath)
		if dir == "." {
			dir = ""
		}
		
		dirCards[dir] = append(dirCards[dir], card)
	}

	// Build tree structure
	dirNodes := make(map[string]*DeckNode)
	dirNodes[""] = root

	// Create directory nodes
	for dir := range dirCards {
		if dir == "" {
			continue
		}
		
		parts := strings.Split(dir, string(filepath.Separator))
		currentPath := ""
		
		for i, part := range parts {
			if i > 0 {
				currentPath += string(filepath.Separator)
			}
			currentPath += part
			
			if _, exists := dirNodes[currentPath]; !exists {
				parentPath := filepath.Dir(currentPath)
				if parentPath == "." {
					parentPath = ""
				}
				
				parent := dirNodes[parentPath]
				node := &DeckNode{
					Name:     part,
					Path:     filepath.Join(deckPath, currentPath),
					IsDir:    true,
					Children: make([]*DeckNode, 0),
					Parent:   parent,
				}
				
				parent.Children = append(parent.Children, node)
				dirNodes[currentPath] = node
			}
		}
	}

	// Add cards to their directories
	for dir, cards := range dirCards {
		if node, exists := dirNodes[dir]; exists {
			node.Cards = cards
		}
	}

	// Sort children and cards
	sortNode(root)
	
	return root, nil
}

func sortNode(node *DeckNode) {
	// Sort children alphabetically
	sort.Slice(node.Children, func(i, j int) bool {
		return node.Children[i].Name < node.Children[j].Name
	})
	
	// Sort cards alphabetically
	sort.Slice(node.Cards, func(i, j int) bool {
		return filepath.Base(node.Cards[i].FilePath) < filepath.Base(node.Cards[j].FilePath)
	})
	
	// Recursively sort children
	for _, child := range node.Children {
		sortNode(child)
	}
}

func printDeckTree(node *DeckNode, prefix string, isLast bool) {
	if node.Parent != nil { // Don't print the root deck name
		// Print current node
		connector := "├── "
		if isLast {
			connector = "└── "
		}
		
		if node.IsDir {
			fmt.Printf("%s%s%s/\n", prefix, connector, node.Name)
		}
		
		// Update prefix for children
		if isLast {
			prefix = prefix + "    "
		} else {
			prefix = prefix + "│   "
		}
	}
	
	// Print cards in this directory
	for i, card := range node.Cards {
		isLastCard := i == len(node.Cards)-1 && len(node.Children) == 0
		
		connector := "├── "
		if isLastCard {
			connector = "└── "
		}
		
		cardName := strings.TrimSuffix(filepath.Base(card.FilePath), ".md")
		statusInfo := getCardStatusInfo(card)
		
		fmt.Printf("%s%s%s %s\n", prefix, connector, cardName, statusInfo)
	}
	
	// Print subdirectories
	for i, child := range node.Children {
		isLastChild := i == len(node.Children)-1
		printDeckTree(child, prefix, isLastChild)
	}
}

func getCardStatusInfo(card *Card) string {
	now := time.Now()
	
	// ANSI color codes
	const (
		Red    = "\033[31m"
		Yellow = "\033[33m"
		Green  = "\033[32m"
		Blue   = "\033[34m"
		Gray   = "\033[37m"
		Reset  = "\033[0m"
	)
	
	if card.FSRSCard.Due.Before(now) || card.FSRSCard.Due.Equal(now) {
		return Red + "due now" + Reset
	}
	
	// Calculate time until due
	timeUntil := card.FSRSCard.Due.Sub(now)
	
	if timeUntil < 24*time.Hour {
		hours := int(timeUntil.Hours())
		return fmt.Sprintf(Yellow+"due in %dh"+Reset, hours)
	} else if timeUntil < 7*24*time.Hour {
		days := int(timeUntil.Hours() / 24)
		return fmt.Sprintf(Green+"due in %dd"+Reset, days)
	} else if timeUntil < 30*24*time.Hour {
		weeks := int(timeUntil.Hours() / (24 * 7))
		return fmt.Sprintf(Blue+"due in %dw"+Reset, weeks)
	} else {
		months := int(timeUntil.Hours() / (24 * 30))
		return fmt.Sprintf(Gray+"due in %dmo"+Reset, months)
	}
}

func statusCommand(deckPath string) error {
	// Build the tree structure
	tree, err := buildDeckTree(deckPath)
	if err != nil {
		return fmt.Errorf("failed to build deck tree: %v", err)
	}
	
	// Get all cards for detailed stats
	cards, err := findCards(deckPath)
	if err != nil {
		return fmt.Errorf("failed to load cards: %v", err)
	}
	
	// Count totals and states
	totalCards, dueCards := countCards(tree)
	new, learning, review, relearning := countCardStates(cards)
	
	// Print header with comprehensive stats
	fmt.Printf("Deck: %s\n", deckPath)
	fmt.Printf("Cards: %d total, %d due | %d new, %d learning, %d review, %d relearning\n\n", 
		totalCards, dueCards, new, learning, review, relearning)
	
	// Print the tree
	if totalCards == 0 {
		fmt.Println("No cards found in this deck.")
		return nil
	}
	
	printDeckTree(tree, "", true)
	
	return nil
}

func countCardStates(cards []*Card) (new, learning, review, relearning int) {
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
	return new, learning, review, relearning
}

func countCards(node *DeckNode) (total, due int) {
	now := time.Now()
	
	// Count cards in this node
	total += len(node.Cards)
	for _, card := range node.Cards {
		if card.FSRSCard.Due.Before(now) || card.FSRSCard.Due.Equal(now) {
			due++
		}
	}
	
	// Count cards in children
	for _, child := range node.Children {
		childTotal, childDue := countCards(child)
		total += childTotal
		due += childDue
	}
	
	return total, due
}