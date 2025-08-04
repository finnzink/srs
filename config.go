package main

import (
	"bufio"
	"fmt"
	"os"
	"path/filepath"
	"strings"
)

type Config struct {
	BaseDeckPath string
}

const ConfigDirName = "srs"
const ConfigFileName = "config"

func getConfigPath() (string, error) {
	configDir := os.Getenv("XDG_CONFIG_HOME")
	if configDir == "" {
		homeDir, err := os.UserHomeDir()
		if err != nil {
			return "", err
		}
		configDir = filepath.Join(homeDir, ".config")
	}
	
	srsConfigDir := filepath.Join(configDir, ConfigDirName)
	
	// Ensure the config directory exists
	err := os.MkdirAll(srsConfigDir, 0755)
	if err != nil {
		return "", err
	}
	
	return filepath.Join(srsConfigDir, ConfigFileName), nil
}

func loadConfig() (*Config, error) {
	configPath, err := getConfigPath()
	if err != nil {
		return &Config{}, nil
	}

	config := &Config{}

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

		// Parse base_deck=path format
		if strings.HasPrefix(line, "base_deck=") {
			value := strings.TrimSpace(line[10:]) // Remove "base_deck="
			// Expand ~ to home directory
			if strings.HasPrefix(value, "~/") {
				homeDir, err := os.UserHomeDir()
				if err == nil {
					value = filepath.Join(homeDir, value[2:])
				}
			}
			config.BaseDeckPath = value
			break // Only need this one line
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
	fmt.Fprintln(file, "# Base deck path - all subdirectories will be relative to this")
	fmt.Fprintln(file, "")

	// Write base deck path
	if config.BaseDeckPath != "" {
		// Convert absolute paths back to ~ notation for portability
		path := config.BaseDeckPath
		homeDir, err := os.UserHomeDir()
		if err == nil && strings.HasPrefix(path, homeDir) {
			path = "~" + path[len(homeDir):]
		}
		fmt.Fprintf(file, "base_deck=%s\n", path)
	}

	return nil
}

func resolveDeckPath(deckName string, config *Config) (string, error) {
	// If no base deck is configured, return error
	if config.BaseDeckPath == "" {
		return "", fmt.Errorf("no base deck configured - run 'srs config' to set up")
	}

	// If it's an absolute path, use it directly (backwards compatibility)
	if filepath.IsAbs(deckName) {
		return filepath.Abs(deckName)
	}

	// If deckName is "." or empty, use base deck
	if deckName == "." || deckName == "" {
		return filepath.Abs(config.BaseDeckPath)
	}

	// Otherwise, treat it as a subdirectory of the base deck
	return filepath.Abs(filepath.Join(config.BaseDeckPath, deckName))
}


func promptForBaseDeck() error {
	fmt.Println("Welcome to SRS! Let's set up your base deck directory.")
	fmt.Println("This will be the root directory for all your flashcards.")
	fmt.Println()
	
	reader := bufio.NewReader(os.Stdin)
	
	for {
		fmt.Print("Enter the path for your base deck directory (or press Enter for ~/flashcards): ")
		input, err := reader.ReadString('\n')
		if err != nil {
			return fmt.Errorf("failed to read input: %v", err)
		}
		
		input = strings.TrimSpace(input)
		
		// Default to ~/flashcards if no input
		if input == "" {
			homeDir, err := os.UserHomeDir()
			if err != nil {
				return fmt.Errorf("failed to get home directory: %v", err)
			}
			input = filepath.Join(homeDir, "flashcards")
		}
		
		// Expand ~ to home directory
		if strings.HasPrefix(input, "~/") {
			homeDir, err := os.UserHomeDir()
			if err != nil {
				return fmt.Errorf("failed to get home directory: %v", err)
			}
			input = filepath.Join(homeDir, input[2:])
		}
		
		// Get absolute path
		absPath, err := filepath.Abs(input)
		if err != nil {
			fmt.Printf("Invalid path: %v\n", err)
			continue
		}
		
		// Check if directory exists
		if _, err := os.Stat(absPath); os.IsNotExist(err) {
			fmt.Printf("Directory %s does not exist. Create it? (y/n): ", absPath)
			response, err := reader.ReadString('\n')
			if err != nil {
				return fmt.Errorf("failed to read input: %v", err)
			}
			response = strings.TrimSpace(strings.ToLower(response))
			
			if response == "y" || response == "yes" {
				err := os.MkdirAll(absPath, 0755)
				if err != nil {
					fmt.Printf("Failed to create directory: %v\n", err)
					continue
				}
				fmt.Printf("Created directory: %s\n", absPath)
			} else {
				continue
			}
		}
		
		// Save the configuration
		config := &Config{
			BaseDeckPath: absPath,
		}
		
		err = saveConfig(config)
		if err != nil {
			return fmt.Errorf("failed to save config: %v", err)
		}
		
		configPath, _ := getConfigPath()
		fmt.Printf("✅ Base deck configured at: %s\n", absPath)
		fmt.Printf("Configuration saved to: %s\n", configPath)
		fmt.Println()
		fmt.Println("You can now use commands like:")
		fmt.Println("  srs review           # Review cards from base deck")
		fmt.Println("  srs review spanish   # Review cards from spanish subdirectory")
		fmt.Println("  srs review spanish/grammar  # Review from nested subdirectories")
		
		return nil
	}
}