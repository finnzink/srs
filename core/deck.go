package core

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/open-spaced-repetition/go-fsrs/v3"
)

// LoadConfig loads the application configuration
func LoadConfig() (*Config, error) {
	homeDir, err := os.UserHomeDir()
	if err != nil {
		return nil, err
	}
	
	configPath := filepath.Join(homeDir, ".srs_config.json")
	
	if _, err := os.Stat(configPath); os.IsNotExist(err) {
		return &Config{}, nil
	}
	
	data, err := os.ReadFile(configPath)
	if err != nil {
		return nil, err
	}
	
	var config Config
	err = json.Unmarshal(data, &config)
	if err != nil {
		return nil, err
	}
	
	return &config, nil
}

// SaveConfig saves the application configuration
func SaveConfig(config *Config) error {
	homeDir, err := os.UserHomeDir()
	if err != nil {
		return err
	}
	
	configPath := filepath.Join(homeDir, ".srs_config.json")
	
	data, err := json.MarshalIndent(config, "", "  ")
	if err != nil {
		return err
	}
	
	return os.WriteFile(configPath, data, 0644)
}

// ResolveDeckPath resolves a relative deck path using the config
func ResolveDeckPath(deckPath string, config *Config) (string, error) {
	if filepath.IsAbs(deckPath) {
		return deckPath, nil
	}
	
	if config == nil || config.BaseDeckPath == "" {
		return "", fmt.Errorf("no base deck path configured")
	}
	
	if deckPath == "." {
		return config.BaseDeckPath, nil
	}
	
	return filepath.Join(config.BaseDeckPath, deckPath), nil
}

// GetDeckStats calculates statistics for a deck
func GetDeckStats(cards []*Card) DeckStats {
	stats := DeckStats{
		TotalCards: len(cards),
	}
	
	dueCards := GetDueCards(cards)
	stats.DueCards = len(dueCards)
	
	for _, card := range cards {
		switch card.FSRSCard.State {
		case fsrs.New:
			stats.NewCards++
		case fsrs.Learning, fsrs.Relearning:
			stats.LearningCards++
		case fsrs.Review:
			stats.ReviewCards++
		}
	}
	
	return stats
}

// GetDeckTree builds a tree structure of decks and their stats
func GetDeckTree(basePath string) (map[string]DeckStats, error) {
	deckStats := make(map[string]DeckStats)
	
	err := filepath.Walk(basePath, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}
		
		if info.IsDir() {
			relPath, err := filepath.Rel(basePath, path)
			if err != nil {
				return err
			}
			
			if relPath == "." {
				relPath = ""
			}
			
			// Get cards for this specific directory (not recursive)
			cards, err := getCardsInDirectory(path)
			if err != nil {
				return err
			}
			
			if len(cards) > 0 {
				deckStats[relPath] = GetDeckStats(cards)
			}
		}
		
		return nil
	})
	
	return deckStats, err
}

// getCardsInDirectory gets cards only in the specified directory (non-recursive)
func getCardsInDirectory(dirPath string) ([]*Card, error) {
	var cards []*Card
	
	entries, err := os.ReadDir(dirPath)
	if err != nil {
		return nil, err
	}
	
	for _, entry := range entries {
		if !entry.IsDir() && strings.HasSuffix(strings.ToLower(entry.Name()), ".md") {
			cardPath := filepath.Join(dirPath, entry.Name())
			card, err := ParseCard(cardPath)
			if err != nil {
				fmt.Printf("Warning: failed to parse card %s: %v\n", cardPath, err)
				continue
			}
			cards = append(cards, card)
		}
	}
	
	return cards, nil
}