package main

import (
	"bufio"
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"github.com/open-spaced-repetition/go-fsrs/v3"
)

// MCP Protocol types
type MCPRequest struct {
	ID     *json.RawMessage `json:"id"`
	Method string           `json:"method"`
	Params json.RawMessage  `json:"params,omitempty"`
}

type MCPResponse struct {
	ID     *json.RawMessage `json:"id"`
	Result interface{}      `json:"result,omitempty"`
	Error  *MCPError        `json:"error,omitempty"`
}

type MCPError struct {
	Code    int    `json:"code"`
	Message string `json:"message"`
}

type ToolCallParams struct {
	Name      string                 `json:"name"`
	Arguments map[string]interface{} `json:"arguments"`
}

// Tool implementations
func handleGetDueCards(config *Config, args map[string]interface{}) (interface{}, error) {
	deckPath := "."
	if path, ok := args["deck_path"].(string); ok && path != "" {
		deckPath = path
	}
	
	resolvedPath, err := resolveDeckPath(deckPath, config)
	if err != nil {
		return nil, fmt.Errorf("error resolving deck path: %v", err)
	}
	
	cards, err := findCards(resolvedPath)
	if err != nil {
		return nil, fmt.Errorf("error loading cards: %v", err)
	}
	
	dueCards := getDueCards(cards)
	
	result := map[string]interface{}{
		"deck_path":   deckPath,
		"total_cards": len(cards),
		"due_count":   len(dueCards),
		"due_cards":   make([]map[string]interface{}, len(dueCards)),
	}
	
	for i, card := range dueCards {
		result["due_cards"].([]map[string]interface{})[i] = map[string]interface{}{
			"file_path":  card.FilePath,
			"question":   card.Question,
			"answer":     card.Answer,
			"due":        card.FSRSCard.Due.Format("2006-01-02T15:04:05Z"),
			"state":      stateString(card.FSRSCard.State),
			"reps":       card.FSRSCard.Reps,
			"difficulty": card.FSRSCard.Difficulty,
			"stability":  card.FSRSCard.Stability,
		}
	}
	
	return result, nil
}

func handleRateCard(config *Config, args map[string]interface{}) (interface{}, error) {
	filePath, ok := args["file_path"].(string)
	if !ok || filePath == "" {
		return nil, fmt.Errorf("file_path is required")
	}
	
	ratingFloat, ok := args["rating"].(float64)
	if !ok {
		return nil, fmt.Errorf("rating is required (1-4)")
	}
	
	rating := int(ratingFloat)
	if rating < 1 || rating > 4 {
		return nil, fmt.Errorf("rating must be an integer between 1-4")
	}
	
	// Resolve file path if it's relative
	if !filepath.IsAbs(filePath) {
		filePath = filepath.Join(config.BaseDeckPath, filePath)
	}
	
	card, err := parseCard(filePath)
	if err != nil {
		return nil, fmt.Errorf("error parsing card: %v", err)
	}
	
	// Convert rating and update card using existing review logic
	err = rateCard(card, rating)
	if err != nil {
		return nil, fmt.Errorf("error rating card: %v", err)
	}
	
	result := map[string]interface{}{
		"success":      true,
		"card_path":    filePath,
		"rating":       fmt.Sprintf("%d", rating),
		"new_due_date": card.FSRSCard.Due.Format("2006-01-02T15:04:05Z"),
		"new_state":    stateString(card.FSRSCard.State),
		"reps":         card.FSRSCard.Reps,
		"difficulty":   card.FSRSCard.Difficulty,
		"stability":    card.FSRSCard.Stability,
	}
	
	return result, nil
}

func handleGetDeckStats(config *Config, args map[string]interface{}) (interface{}, error) {
	deckPath := "."
	if path, ok := args["deck_path"].(string); ok && path != "" {
		deckPath = path
	}
	
	resolvedPath, err := resolveDeckPath(deckPath, config)
	if err != nil {
		return nil, fmt.Errorf("error resolving deck path: %v", err)
	}
	
	cards, err := findCards(resolvedPath)
	if err != nil {
		return nil, fmt.Errorf("error loading cards: %v", err)
	}
	
	stats := getSimpleDeckStats(cards)
	
	result := map[string]interface{}{
		"deck_path":   deckPath,
		"total_cards": stats.TotalCards,
		"due_cards":   stats.DueCards,
	}
	
	return result, nil
}

func handleListDecks(config *Config, args map[string]interface{}) (interface{}, error) {
	deckTree, err := getSimpleDeckTree(config.BaseDeckPath)
	if err != nil {
		return nil, fmt.Errorf("error getting deck tree: %v", err)
	}
	
	result := map[string]interface{}{
		"base_path": config.BaseDeckPath,
		"decks":     deckTree,
	}
	
	return result, nil
}

// Helper functions for deck operations
type SimpleDeckStats struct {
	TotalCards int
	DueCards   int
}

func getSimpleDeckStats(cards []*Card) SimpleDeckStats {
	dueCards := getDueCards(cards)
	return SimpleDeckStats{
		TotalCards: len(cards),
		DueCards:   len(dueCards),
	}
}

func getSimpleDeckTree(basePath string) (map[string]SimpleDeckStats, error) {
	result := make(map[string]SimpleDeckStats)
	
	cards, err := findCards(basePath)
	if err != nil {
		return nil, err
	}
	
	// For now, just return the root deck stats
	stats := getSimpleDeckStats(cards)
	result["."] = stats
	
	return result, nil
}

// Simple card rating function
func rateCard(card *Card, rating int) error {
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
	default:
		return fmt.Errorf("invalid rating: %d (must be 1-4)", rating)
	}
	
	// Create a session and update the card
	session := NewReviewSession([]*Card{card})
	return session.updateCard(card, fsrsRating)
}

// Helper functions for FSRS types
func stateString(state interface{}) string {
	// This is a simplified state string conversion
	return fmt.Sprintf("%v", state)
}

// Simple MCP server implementation
func mcpSimpleCommand() error {
	config, err := loadConfig()
	if err != nil {
		return fmt.Errorf("failed to load config: %v", err)
	}
	
	if config.BaseDeckPath == "" {
		return fmt.Errorf("no base deck path configured. Please run 'srs config' first")
	}
	
	scanner := bufio.NewScanner(os.Stdin)
	
	// Send initialization message
	initResp := MCPResponse{
		ID: nil,
		Result: map[string]interface{}{
			"protocol_version": "1.0",
			"capabilities": map[string]interface{}{
				"tools": map[string]interface{}{
					"srs/get_due_cards": map[string]interface{}{
						"description": "Get cards that are due for review",
						"inputSchema": map[string]interface{}{
							"type": "object",
							"properties": map[string]interface{}{
								"deck_path": map[string]interface{}{
									"type":        "string",
									"description": "Path to deck (relative to base deck path, defaults to '.')",
								},
							},
						},
					},
					"srs/rate_card": map[string]interface{}{
						"description": "Rate a card and update its scheduling",
						"inputSchema": map[string]interface{}{
							"type": "object",
							"properties": map[string]interface{}{
								"file_path": map[string]interface{}{
									"type":        "string",
									"description": "Path to the card file",
								},
								"rating": map[string]interface{}{
									"type":        "number",
									"description": "Rating (1=Again, 2=Hard, 3=Good, 4=Easy)",
								},
							},
							"required": []string{"file_path", "rating"},
						},
					},
					"srs/get_deck_stats": map[string]interface{}{
						"description": "Get statistics for a deck",
						"inputSchema": map[string]interface{}{
							"type": "object",
							"properties": map[string]interface{}{
								"deck_path": map[string]interface{}{
									"type":        "string",
									"description": "Path to deck (relative to base deck path, defaults to '.')",
								},
							},
						},
					},
					"srs/list_decks": map[string]interface{}{
						"description": "List all available decks",
						"inputSchema": map[string]interface{}{
							"type":       "object",
							"properties": map[string]interface{}{},
						},
					},
				},
			},
		},
	}
	
	respBytes, _ := json.Marshal(initResp)
	fmt.Println(string(respBytes))
	
	// Process requests
	for scanner.Scan() {
		line := scanner.Text()
		if line == "" {
			continue
		}
		
		var req MCPRequest
		if err := json.Unmarshal([]byte(line), &req); err != nil {
			continue
		}
		
		var resp MCPResponse
		resp.ID = req.ID
		
		if req.Method == "tools/call" {
			var params ToolCallParams
			if err := json.Unmarshal(req.Params, &params); err != nil {
				resp.Error = &MCPError{Code: -32602, Message: "Invalid params"}
			} else {
				var result interface{}
				var err error
				
				switch params.Name {
				case "srs/get_due_cards":
					result, err = handleGetDueCards(config, params.Arguments)
				case "srs/rate_card":
					result, err = handleRateCard(config, params.Arguments)
				case "srs/get_deck_stats":
					result, err = handleGetDeckStats(config, params.Arguments)
				case "srs/list_decks":
					result, err = handleListDecks(config, params.Arguments)
				default:
					err = fmt.Errorf("unknown tool: %s", params.Name)
				}
				
				if err != nil {
					resp.Error = &MCPError{Code: -32603, Message: err.Error()}
				} else {
					resp.Result = map[string]interface{}{
						"content": []map[string]interface{}{
							{
								"type": "text",
								"text": fmt.Sprintf("%v", result),
							},
						},
					}
				}
			}
		}
		
		respBytes, _ := json.Marshal(resp)
		fmt.Println(string(respBytes))
	}
	
	return nil
}