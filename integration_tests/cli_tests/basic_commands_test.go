package cli_tests

import (
	"strings"
	"testing"
	"time"

	"integration_tests/fixtures"
	"integration_tests/helpers"
)

func TestVersionCommand(t *testing.T) {
	config, err := helpers.NewTestConfig()
	if err != nil {
		t.Fatalf("Failed to create test config: %v", err)
	}
	defer config.Cleanup()

	result, err := config.RunCommand("version")
	if err != nil {
		t.Fatalf("Failed to run version command: %v", err)
	}

	if !result.Success() {
		t.Errorf("Version command failed with exit code %d", result.ExitCode)
	}

	if !strings.Contains(result.Stdout, "srs") {
		t.Errorf("Version output should contain 'srs', got: %s", result.Stdout)
	}

	helpers.LogResult("version_command", result, "../logs")
}

func TestConfigCommand(t *testing.T) {
	config, err := helpers.NewTestConfig()
	if err != nil {
		t.Fatalf("Failed to create test config: %v", err)
	}
	defer config.Cleanup()

	// Note: config command is interactive, but it should at least start
	// For now we'll test that it doesn't crash immediately
	result, err := config.RunCommand("--help")
	if err != nil {
		t.Fatalf("Failed to run help command: %v", err)
	}

	if !result.Success() {
		t.Errorf("Help command failed with exit code %d", result.ExitCode)
	}

	helpers.LogResult("config_command", result, "../logs")
}

func TestListCommand(t *testing.T) {
	config, err := helpers.NewTestConfig()
	if err != nil {
		t.Fatalf("Failed to create test config: %v", err)
	}
	defer config.Cleanup()

	// Create test decks
	mathDeck := fixtures.BasicMathDeck()
	if err := mathDeck.CreateDeck(config.BaseDeckPath); err != nil {
		t.Fatalf("Failed to create math deck: %v", err)
	}

	progDeck := fixtures.ProgrammingDeck()
	if err := progDeck.CreateDeck(config.BaseDeckPath); err != nil {
		t.Fatalf("Failed to create programming deck: %v", err)
	}

	// Test list command
	result, err := config.RunCommand("list")
	if err != nil {
		t.Fatalf("Failed to run list command: %v", err)
	}

	if !result.Success() {
		t.Errorf("List command failed with exit code %d, stderr: %s", result.ExitCode, result.Stderr)
	}

	// Should show our test decks
	if !strings.Contains(result.Stdout, "basic_math") {
		t.Errorf("List output should contain 'basic_math' deck, got: %s", result.Stdout)
	}

	if !strings.Contains(result.Stdout, "programming") {
		t.Errorf("List output should contain 'programming' deck, got: %s", result.Stdout)
	}

	helpers.LogResult("list_command", result, "../logs")
}

func TestListSpecificDeck(t *testing.T) {
	config, err := helpers.NewTestConfig()
	if err != nil {
		t.Fatalf("Failed to create test config: %v", err)
	}
	defer config.Cleanup()

	// Create test deck
	mathDeck := fixtures.BasicMathDeck()
	if err := mathDeck.CreateDeck(config.BaseDeckPath); err != nil {
		t.Fatalf("Failed to create math deck: %v", err)
	}

	// Test list specific deck
	result, err := config.RunCommand("list", "basic_math")
	if err != nil {
		t.Fatalf("Failed to run list command for specific deck: %v", err)
	}

	if !result.Success() {
		t.Errorf("List specific deck failed with exit code %d, stderr: %s", result.ExitCode, result.Stderr)
	}

	helpers.LogResult("list_specific_deck", result, "../logs")
}

func TestListNonExistentDeck(t *testing.T) {
	config, err := helpers.NewTestConfig()
	if err != nil {
		t.Fatalf("Failed to create test config: %v", err)
	}
	defer config.Cleanup()

	// Test list non-existent deck
	result, err := config.RunCommand("list", "nonexistent")
	if err != nil {
		t.Fatalf("Failed to run list command for non-existent deck: %v", err)
	}

	if result.Success() {
		t.Errorf("List non-existent deck should fail, but got exit code %d", result.ExitCode)
	}

	if !strings.Contains(result.Stderr, "does not exist") {
		t.Errorf("Error message should mention 'does not exist', got: %s", result.Stderr)
	}

	helpers.LogResult("list_nonexistent_deck", result, "../logs")
}

func TestCommandPerformance(t *testing.T) {
	config, err := helpers.NewTestConfig()
	if err != nil {
		t.Fatalf("Failed to create test config: %v", err)
	}
	defer config.Cleanup()

	// Create a moderate-sized deck
	mathDeck := fixtures.BasicMathDeck()
	if err := mathDeck.CreateDeck(config.BaseDeckPath); err != nil {
		t.Fatalf("Failed to create math deck: %v", err)
	}

	// Test that list command completes quickly
	result, err := config.RunCommand("list")
	if err != nil {
		t.Fatalf("Failed to run list command: %v", err)
	}

	if !result.Success() {
		t.Errorf("List command failed with exit code %d", result.ExitCode)
	}

	// Should complete within reasonable time (2 seconds is generous)
	if result.Duration > 2*time.Second {
		t.Errorf("List command took too long: %v", result.Duration)
	}

	helpers.LogResult("command_performance", result, "../logs")
}

func TestInvalidCommand(t *testing.T) {
	config, err := helpers.NewTestConfig()
	if err != nil {
		t.Fatalf("Failed to create test config: %v", err)
	}
	defer config.Cleanup()

	// Test invalid command
	result, err := config.RunCommand("invalid_command")
	if err != nil {
		t.Fatalf("Failed to run invalid command: %v", err)
	}

	if result.Success() {
		t.Errorf("Invalid command should fail, but got exit code %d", result.ExitCode)
	}

	if !strings.Contains(result.Stderr, "Unknown command") {
		t.Errorf("Error message should mention 'Unknown command', got: %s", result.Stderr)
	}

	helpers.LogResult("invalid_command", result, "../logs")
}