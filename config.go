package main

import (
	"bufio"
	"fmt"
	"os"
	"path/filepath"
	"strings"
)

type Config struct {
	DefaultDeck string
	Decks       map[string]string // alias -> path
}

const ConfigFileName = ".srsrc"

func getConfigPath() (string, error) {
	homeDir, err := os.UserHomeDir()
	if err != nil {
		return "", err
	}
	return filepath.Join(homeDir, ConfigFileName), nil
}

func loadConfig() (*Config, error) {
	configPath, err := getConfigPath()
	if err != nil {
		return &Config{Decks: make(map[string]string)}, nil
	}

	config := &Config{
		Decks: make(map[string]string),
	}

	file, err := os.Open(configPath)
	if err != nil {
		// If config file doesn't exist, return empty config
		if os.IsNotExist(err) {
			return config, nil
		}
		return nil, err
	}
	defer file.Close()

	scanner := bufio.NewScanner(file)
	for scanner.Scan() {
		line := strings.TrimSpace(scanner.Text())
		
		// Skip empty lines and comments
		if line == "" || strings.HasPrefix(line, "#") {
			continue
		}

		// Parse key=value format
		if strings.Contains(line, "=") {
			parts := strings.SplitN(line, "=", 2)
			key := strings.TrimSpace(parts[0])
			value := strings.TrimSpace(parts[1])

			if key == "default" {
				config.DefaultDeck = value
			} else {
				// Expand ~ to home directory
				if strings.HasPrefix(value, "~/") {
					homeDir, err := os.UserHomeDir()
					if err == nil {
						value = filepath.Join(homeDir, value[2:])
					}
				}
				config.Decks[key] = value
			}
		}
	}

	return config, scanner.Err()
}

func saveConfig(config *Config) error {
	configPath, err := getConfigPath()
	if err != nil {
		return err
	}

	file, err := os.Create(configPath)
	if err != nil {
		return err
	}
	defer file.Close()

	// Write header
	fmt.Fprintln(file, "# SRS Configuration")
	fmt.Fprintln(file, "# Format: alias=path or default=alias")
	fmt.Fprintln(file, "")

	// Write default deck
	if config.DefaultDeck != "" {
		fmt.Fprintf(file, "default=%s\n", config.DefaultDeck)
		fmt.Fprintln(file, "")
	}

	// Write deck aliases
	fmt.Fprintln(file, "# Deck aliases")
	for alias, path := range config.Decks {
		// Convert absolute paths back to ~ notation for portability
		homeDir, err := os.UserHomeDir()
		if err == nil && strings.HasPrefix(path, homeDir) {
			path = "~" + path[len(homeDir):]
		}
		fmt.Fprintf(file, "%s=%s\n", alias, path)
	}

	return nil
}

func resolveDeckPath(deckName string, config *Config) (string, error) {
	// If it's an absolute or relative path, use it directly
	if filepath.IsAbs(deckName) || strings.Contains(deckName, "/") {
		return filepath.Abs(deckName)
	}

	// Check if it's a deck alias
	if path, exists := config.Decks[deckName]; exists {
		return filepath.Abs(path)
	}

	// If no deck specified and we have a default, use it
	if deckName == "." && config.DefaultDeck != "" {
		if path, exists := config.Decks[config.DefaultDeck]; exists {
			return filepath.Abs(path)
		}
		// Default deck might be a path
		return filepath.Abs(config.DefaultDeck)
	}

	// Fall back to treating it as a path
	return filepath.Abs(deckName)
}

func createDefaultConfig() error {
	config := &Config{
		Decks: map[string]string{
			"example": "./example_deck",
		},
		DefaultDeck: "example",
	}

	err := saveConfig(config)
	if err != nil {
		return err
	}

	configPath, _ := getConfigPath()
	fmt.Printf("Created default config at %s\n", configPath)
	fmt.Println("Edit this file to add your deck locations and aliases.")
	
	return nil
}