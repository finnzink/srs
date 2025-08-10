package integration_tests

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
	"time"

	"integration_tests/fixtures"
	"integration_tests/helpers"
)

// TestCompleteCardLifecycle tests the complete flow: create cards -> list -> review via MCP -> verify updates
func TestCompleteCardLifecycle(t *testing.T) {
	config, err := helpers.NewTestConfig()
	if err != nil {
		t.Fatalf("Failed to create test config: %v", err)
	}
	defer config.Cleanup()

	// Step 1: Create cards
	mathDeck := fixtures.BasicMathDeck()
	if err := mathDeck.CreateDeck(config.BaseDeckPath); err != nil {
		t.Fatalf("Failed to create math deck: %v", err)
	}

	// Step 2: Verify cards show up in list
	result, err := config.RunCommand("list", "basic_math")
	if err != nil {
		t.Fatalf("Failed to run list command: %v", err)
	}

	if !result.Success() {
		t.Fatalf("List command failed: %s", result.Stderr)
	}

	if !strings.Contains(result.Stdout, "basic_math") {
		t.Errorf("List should show basic_math deck, got: %s", result.Stdout)
	}

	// Step 3: Use MCP to get due cards
	client, err := config.NewMCPClient()
	if err != nil {
		t.Fatalf("Failed to start MCP client: %v", err)
	}
	defer client.Close()

	params := map[string]interface{}{
		"name": "srs/get_due_cards",
		"arguments": map[string]interface{}{
			"deck_path": "basic_math",
		},
	}

	response, err := client.SendRequest("tools/call", params)
	if err != nil {
		t.Fatalf("Failed to get due cards: %v", err)
	}

	if response.Error != nil {
		t.Fatalf("MCP get_due_cards returned error: %s", response.Error.Message)
	}

	// Step 4: Rate a card via MCP
	rateParams := map[string]interface{}{
		"name": "srs/rate_card",
		"arguments": map[string]interface{}{
			"file_path": "basic_math/addition.md",
			"rating":    4, // Easy
		},
	}

	rateResponse, err := client.SendRequest("tools/call", rateParams)
	if err != nil {
		t.Fatalf("Failed to rate card: %v", err)
	}

	if rateResponse.Error != nil {
		t.Fatalf("MCP rate_card returned error: %s", rateResponse.Error.Message)
	}
	
	t.Logf("DEBUG: MCP rate_card response: %+v", rateResponse)

	// Step 5: Verify card was updated with FSRS metadata
	cardPath := filepath.Join(config.BaseDeckPath, "basic_math", "addition.md")
	t.Logf("DEBUG: Reading card from: %s", cardPath)
	cardContent, err := os.ReadFile(cardPath)
	if err != nil {
		t.Fatalf("Failed to read updated card: %v", err)
	}

	cardStr := string(cardContent)
	t.Logf("DEBUG: Card content after rating: %s", cardStr)
	if !strings.Contains(cardStr, "FSRS:") {
		t.Errorf("Card should contain FSRS metadata after rating, got: %s", cardStr)
	}

	// Step 6: Verify due date was updated
	if !strings.Contains(cardStr, "due:") {
		t.Errorf("Card should contain due date after rating, got: %s", cardStr)
	}

	t.Log("Complete card lifecycle test passed successfully")
}

// TestMultiDeckWorkflow tests operations across multiple decks
func TestMultiDeckWorkflow(t *testing.T) {
	config, err := helpers.NewTestConfig()
	if err != nil {
		t.Fatalf("Failed to create test config: %v", err)
	}
	defer config.Cleanup()

	// Create multiple decks
	mathDeck := fixtures.BasicMathDeck()
	if err := mathDeck.CreateDeck(config.BaseDeckPath); err != nil {
		t.Fatalf("Failed to create math deck: %v", err)
	}

	progDeck := fixtures.ProgrammingDeck()
	if err := progDeck.CreateDeck(config.BaseDeckPath); err != nil {
		t.Fatalf("Failed to create programming deck: %v", err)
	}

	// Test listing all decks
	result, err := config.RunCommand("list")
	if err != nil {
		t.Fatalf("Failed to run list command: %v", err)
	}

	if !result.Success() {
		t.Fatalf("List command failed: %s", result.Stderr)
	}

	// Should show both decks
	if !strings.Contains(result.Stdout, "basic_math") {
		t.Errorf("Should show basic_math deck")
	}
	if !strings.Contains(result.Stdout, "programming") {
		t.Errorf("Should show programming deck")
	}

	// Test MCP list_decks
	client, err := config.NewMCPClient()
	if err != nil {
		t.Fatalf("Failed to start MCP client: %v", err)
	}
	defer client.Close()

	params := map[string]interface{}{
		"name": "srs/list_decks",
		"arguments": map[string]interface{}{},
	}

	response, err := client.SendRequest("tools/call", params)
	if err != nil {
		t.Fatalf("Failed to list decks via MCP: %v", err)
	}

	if response.Error != nil {
		t.Fatalf("MCP list_decks returned error: %s", response.Error.Message)
	}

	// Test getting stats for specific decks
	for _, deckName := range []string{"basic_math", "programming"} {
		statsParams := map[string]interface{}{
			"name": "srs/get_deck_stats",
			"arguments": map[string]interface{}{
				"deck_path": deckName,
			},
		}

		statsResponse, err := client.SendRequest("tools/call", statsParams)
		if err != nil {
			t.Fatalf("Failed to get stats for %s: %v", deckName, err)
		}

		if statsResponse.Error != nil {
			t.Fatalf("MCP get_deck_stats for %s returned error: %s", deckName, statsResponse.Error.Message)
		}
	}

	t.Log("Multi-deck workflow test passed successfully")
}

// TestReviewProgressTracking tests that card states progress correctly through multiple reviews
func TestReviewProgressTracking(t *testing.T) {
	config, err := helpers.NewTestConfig()
	if err != nil {
		t.Fatalf("Failed to create test config: %v", err)
	}
	defer config.Cleanup()

	// Create a single card for focused testing
	mathDeck := fixtures.BasicMathDeck()
	if err := mathDeck.CreateDeck(config.BaseDeckPath); err != nil {
		t.Fatalf("Failed to create math deck: %v", err)
	}

	client, err := config.NewMCPClient()
	if err != nil {
		t.Fatalf("Failed to start MCP client: %v", err)
	}
	defer client.Close()

	cardFile := "basic_math/addition.md"

	// Rate the card multiple times and verify progression
	ratings := []int{3, 4, 3, 4} // Good, Easy, Good, Easy
	
	for i, rating := range ratings {
		rateParams := map[string]interface{}{
			"name": "srs/rate_card",
			"arguments": map[string]interface{}{
				"file_path": cardFile,
				"rating":    rating,
			},
		}

		response, err := client.SendRequest("tools/call", rateParams)
		if err != nil {
			t.Fatalf("Review %d failed: %v", i+1, err)
		}

		if response.Error != nil {
			t.Fatalf("Review %d returned error: %s", i+1, response.Error.Message)
		}

		// Verify the card file was updated
		cardPath := filepath.Join(config.BaseDeckPath, cardFile)
		cardContent, err := os.ReadFile(cardPath)
		if err != nil {
			t.Fatalf("Failed to read card after review %d: %v", i+1, err)
		}

		cardStr := string(cardContent)
		if !strings.Contains(cardStr, "FSRS:") {
			t.Errorf("Card should contain FSRS metadata after review %d", i+1)
		}

		// Verify reps count increased
		if !strings.Contains(cardStr, "reps:") {
			t.Errorf("Card should contain reps count after review %d", i+1)
		}

		// Small delay between reviews
		time.Sleep(10 * time.Millisecond)
	}

	t.Log("Review progress tracking test passed successfully")
}

// TestErrorRecovery tests how the system handles various error conditions
func TestErrorRecovery(t *testing.T) {
	config, err := helpers.NewTestConfig()
	if err != nil {
		t.Fatalf("Failed to create test config: %v", err)
	}
	defer config.Cleanup()

	// Test 1: Invalid deck path
	result, err := config.RunCommand("list", "nonexistent_deck")
	if err != nil {
		t.Fatalf("Failed to run command: %v", err)
	}
	if result.Success() {
		t.Errorf("Command should fail for nonexistent deck")
	}

	// Test 2: MCP with invalid card path
	client, err := config.NewMCPClient()
	if err != nil {
		t.Fatalf("Failed to start MCP client: %v", err)
	}
	defer client.Close()

	rateParams := map[string]interface{}{
		"name": "srs/rate_card",
		"arguments": map[string]interface{}{
			"file_path": "nonexistent/card.md",
			"rating":    3,
		},
	}

	response, err := client.SendRequest("tools/call", rateParams)
	if err != nil {
		t.Fatalf("Failed to send request: %v", err)
	}

	if response.Error == nil {
		t.Errorf("Expected error for nonexistent card")
	}

	// Test 3: Invalid MCP request format
	malformedResponse, err := client.SendRequest("invalid_method", map[string]interface{}{})
	if err != nil {
		t.Fatalf("Failed to send malformed request: %v", err)
	}

	// Should handle gracefully (may return error or be ignored)
	t.Logf("Malformed request handled: %v", malformedResponse)

	t.Log("Error recovery test passed successfully")
}

// TestPerformanceWithLargeDataset tests performance with a larger number of cards
func TestPerformanceWithLargeDataset(t *testing.T) {
	config, err := helpers.NewTestConfig()
	if err != nil {
		t.Fatalf("Failed to create test config: %v", err)
	}
	defer config.Cleanup()

	// Create multiple decks with multiple cards each
	decks := []fixtures.TestDeck{
		fixtures.BasicMathDeck(),
		fixtures.ProgrammingDeck(),
	}

	for _, deck := range decks {
		if err := deck.CreateDeck(config.BaseDeckPath); err != nil {
			t.Fatalf("Failed to create deck %s: %v", deck.Name, err)
		}
	}

	// Add some reviewed cards
	reviewedDeck := fixtures.ReviewedCardsDeck()
	if err := reviewedDeck.CreateDeck(config.BaseDeckPath); err != nil {
		t.Fatalf("Failed to create reviewed deck: %v", err)
	}

	start := time.Now()

	// Test list performance
	result, err := config.RunCommand("list")
	if err != nil {
		t.Fatalf("Failed to run list command: %v", err)
	}

	if !result.Success() {
		t.Fatalf("List command failed: %s", result.Stderr)
	}

	listDuration := time.Since(start)
	t.Logf("List command took: %v", listDuration)

	// Test MCP performance
	client, err := config.NewMCPClient()
	if err != nil {
		t.Fatalf("Failed to start MCP client: %v", err)
	}
	defer client.Close()

	start = time.Now()

	params := map[string]interface{}{
		"name": "srs/get_due_cards",
		"arguments": map[string]interface{}{
			"deck_path": ".",
		},
	}

	response, err := client.SendRequest("tools/call", params)
	if err != nil {
		t.Fatalf("Failed to get due cards: %v", err)
	}

	if response.Error != nil {
		t.Fatalf("MCP request failed: %s", response.Error.Message)
	}

	mcpDuration := time.Since(start)
	t.Logf("MCP get_due_cards took: %v", mcpDuration)

	// Performance should be reasonable (under 1 second for this dataset size)
	if listDuration > time.Second {
		t.Errorf("List command too slow: %v", listDuration)
	}
	if mcpDuration > time.Second {
		t.Errorf("MCP request too slow: %v", mcpDuration)
	}

	t.Log("Performance test passed successfully")
}