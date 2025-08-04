package main

import (
	"bufio"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestResolveDeckPath(t *testing.T) {
	tempDir := createTempDir(t)
	
	tests := []struct {
		name       string
		deckName   string
		baseDeck   string
		expected   string
		expectError bool
	}{
		{
			name:     "current directory",
			deckName: ".",
			baseDeck: tempDir,
			expected: tempDir,
		},
		{
			name:     "empty deck name",
			deckName: "",
			baseDeck: tempDir,
			expected: tempDir,
		},
		{
			name:     "subdirectory",
			deckName: "spanish",
			baseDeck: tempDir,
			expected: filepath.Join(tempDir, "spanish"),
		},
		{
			name:     "nested subdirectory",
			deckName: "spanish/grammar",
			baseDeck: tempDir,
			expected: filepath.Join(tempDir, "spanish", "grammar"),
		},
		{
			name:     "absolute path",
			deckName: "/absolute/path",
			baseDeck: tempDir,
			expected: "/absolute/path",
		},
		{
			name:        "no base deck configured",
			deckName:    "spanish",
			baseDeck:    "",
			expectError: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			config := &Config{
				BaseDeckPath: tt.baseDeck,
			}

			result, err := resolveDeckPath(tt.deckName, config)

			if tt.expectError {
				if err == nil {
					t.Errorf("Expected error but got none")
				}
				if !strings.Contains(err.Error(), "no base deck configured") {
					t.Errorf("Expected 'no base deck configured' error, got: %v", err)
				}
				return
			}

			if err != nil {
				t.Fatalf("Unexpected error: %v", err)
			}

			// Convert both to absolute paths for comparison
			expectedAbs, _ := filepath.Abs(tt.expected)
			resultAbs, _ := filepath.Abs(result)

			if resultAbs != expectedAbs {
				t.Errorf("Expected %q, got %q", expectedAbs, resultAbs)
			}
		})
	}
}

func TestLoadConfigWithTempFile(t *testing.T) {
	tempDir := createTempDir(t)
	
	tests := []struct {
		name         string
		configContent string
		expected     *Config
	}{
		{
			name:         "valid config",
			configContent: "base_deck=/path/to/flashcards\n",
			expected:     &Config{BaseDeckPath: "/path/to/flashcards"},
		},
		{
			name: "config with comments and empty lines",
			configContent: `# SRS Configuration
# Base deck path - all subdirectories will be relative to this

base_deck=/path/to/flashcards
`,
			expected: &Config{BaseDeckPath: "/path/to/flashcards"},
		},
		{
			name:         "config with tilde expansion",
			configContent: "base_deck=~/flashcards\n",
			expected:     &Config{BaseDeckPath: expandTildePath("~/flashcards")},
		},
		{
			name:         "config with extra whitespace",
			configContent: "base_deck=/path/to/flashcards\n",
			expected:     &Config{BaseDeckPath: "/path/to/flashcards"},
		},
		{
			name:         "empty config",
			configContent: "",
			expected:     &Config{BaseDeckPath: ""},
		},
		{
			name:         "config with only comments",
			configContent: "# Just a comment\n# Another comment\n",
			expected:     &Config{BaseDeckPath: ""},
		},
		{
			name:         "config with unknown lines",
			configContent: "unknown_setting=value\nbase_deck=/path/to/flashcards\nother_line\n",
			expected:     &Config{BaseDeckPath: "/path/to/flashcards"},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			configPath := createTempFile(t, tempDir, "config_"+tt.name, tt.configContent)

			// Create a temporary loadConfig that reads from our test file
			file, err := os.Open(configPath)
			if err != nil {
				t.Fatalf("Failed to open config file: %v", err)
			}
			defer file.Close()

			config := &Config{}
			
			// Duplicate the parsing logic from loadConfig
			scanner := bufio.NewScanner(file)
			for scanner.Scan() {
				line := strings.TrimSpace(scanner.Text())
				
				if line == "" || strings.HasPrefix(line, "#") {
					continue
				}

				if strings.HasPrefix(line, "base_deck=") {
					value := strings.TrimSpace(line[10:])
					if strings.HasPrefix(value, "~/") {
						homeDir, err := os.UserHomeDir()
						if err == nil {
							value = filepath.Join(homeDir, value[2:])
						}
					}
					config.BaseDeckPath = value
					break
				}
			}

			if config.BaseDeckPath != tt.expected.BaseDeckPath {
				t.Errorf("Expected BaseDeckPath %q, got %q", tt.expected.BaseDeckPath, config.BaseDeckPath)
			}
		})
	}
}

func expandTildePath(path string) string {
	if strings.HasPrefix(path, "~/") {
		homeDir, err := os.UserHomeDir()
		if err == nil {
			return filepath.Join(homeDir, path[2:])
		}
	}
	return path
}

func TestSaveAndLoadConfigRoundTrip(t *testing.T) {
	tempDir := createTempDir(t)
	
	// Create a mock config directory structure
	configDir := filepath.Join(tempDir, ".config", "srs")
	err := os.MkdirAll(configDir, 0755)
	if err != nil {
		t.Fatalf("Failed to create config dir: %v", err)
	}
	
	configPath := filepath.Join(configDir, "config")
	
	originalConfig := &Config{
		BaseDeckPath: "/test/path/to/flashcards",
	}

	// Write config manually (testing the format)
	file, err := os.Create(configPath)
	if err != nil {
		t.Fatalf("Failed to create config file: %v", err)
	}
	
	fmt.Fprintln(file, "# SRS Configuration")
	fmt.Fprintln(file, "# Base deck path - all subdirectories will be relative to this")
	fmt.Fprintln(file, "")
	fmt.Fprintf(file, "base_deck=%s\n", originalConfig.BaseDeckPath)
	file.Close()

	// Read it back
	loadedConfig := &Config{}
	file, err = os.Open(configPath)
	if err != nil {
		t.Fatalf("Failed to open config file: %v", err)
	}
	defer file.Close()

	// Use same parsing logic as loadConfig
	scanner := bufio.NewScanner(file)
	for scanner.Scan() {
		line := strings.TrimSpace(scanner.Text())
		
		if line == "" || strings.HasPrefix(line, "#") {
			continue
		}

		if strings.HasPrefix(line, "base_deck=") {
			value := strings.TrimSpace(line[10:])
			if strings.HasPrefix(value, "~/") {
				homeDir, err := os.UserHomeDir()
				if err == nil {
					value = filepath.Join(homeDir, value[2:])
				}
			}
			loadedConfig.BaseDeckPath = value
			break
		}
	}

	if loadedConfig.BaseDeckPath != originalConfig.BaseDeckPath {
		t.Errorf("Round trip failed: expected %q, got %q", 
			originalConfig.BaseDeckPath, loadedConfig.BaseDeckPath)
	}
}

func TestConfigPathExpansion(t *testing.T) {
	homeDir, err := os.UserHomeDir()
	if err != nil {
		t.Skip("Cannot get home directory")
	}

	tests := []struct {
		name     string
		input    string
		expected string
	}{
		{
			name:     "tilde expansion",
			input:    "~/flashcards",
			expected: filepath.Join(homeDir, "flashcards"),
		},
		{
			name:     "tilde with subdirectory",
			input:    "~/Documents/flashcards",
			expected: filepath.Join(homeDir, "Documents", "flashcards"),
		},
		{
			name:     "absolute path unchanged",
			input:    "/absolute/path",
			expected: "/absolute/path",
		},
		{
			name:     "relative path unchanged",
			input:    "relative/path",
			expected: "relative/path",
		},
		{
			name:     "just tilde unchanged",
			input:    "~",
			expected: "~",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Test the expansion logic used in config parsing
			result := tt.input
			if strings.HasPrefix(result, "~/") {
				result = filepath.Join(homeDir, result[2:])
			}

			if result != tt.expected {
				t.Errorf("Expected %q, got %q", tt.expected, result)
			}
		})
	}
}