package main

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"os"
	"path/filepath"
	"strconv"

	"github.com/mark3labs/mcp-go"
	"github.com/mark3labs/mcp-go/mcp"
	"github.com/mark3labs/mcp-go/server"

	"srs/core"
)

type SRSServer struct {
	config *core.Config
}

func NewSRSServer() (*SRSServer, error) {
	config, err := core.LoadConfig()
	if err != nil {
		return nil, fmt.Errorf("failed to load config: %v", err)
	}
	
	if config.BaseDeckPath == "" {
		return nil, fmt.Errorf("no base deck path configured. Please run 'srs config' first")
	}
	
	return &SRSServer{config: config}, nil
}

func (s *SRSServer) handleGetDueCards(ctx context.Context, args map[string]interface{}) (*mcp.CallToolResult, error) {
	deckPath := "."
	if path, ok := args["deck_path"].(string); ok && path != "" {
		deckPath = path
	}
	
	resolvedPath, err := core.ResolveDeckPath(deckPath, s.config)
	if err != nil {
		return &mcp.CallToolResult{
			IsError: true,
			Content: []interface{}{
				map[string]interface{}{
					"type": "text",
					"text": fmt.Sprintf("Error resolving deck path: %v", err),
				},
			},
		}, nil
	}
	
	cards, err := core.FindCards(resolvedPath)
	if err != nil {
		return &mcp.CallToolResult{
			IsError: true,
			Content: []interface{}{
				map[string]interface{}{
					"type": "text",
					"text": fmt.Sprintf("Error loading cards: %v", err),
				},
			},
		}, nil
	}
	
	dueCards := core.GetDueCards(cards)
	
	result := map[string]interface{}{
		"deck_path": deckPath,
		"total_cards": len(cards),
		"due_count": len(dueCards),
		"due_cards": make([]map[string]interface{}, len(dueCards)),
	}
	
	for i, card := range dueCards {
		result["due_cards"].([]map[string]interface{})[i] = map[string]interface{}{
			"file_path": card.FilePath,
			"question": card.Question,
			"answer": card.Answer,
			"due": card.FSRSCard.Due,
			"state": core.StateToString(card.FSRSCard.State),
			"reps": card.FSRSCard.Reps,
			"difficulty": card.FSRSCard.Difficulty,
			"stability": card.FSRSCard.Stability,
		}
	}
	
	resultJSON, _ := json.MarshalIndent(result, "", "  ")
	
	return &mcp.CallToolResult{
		Content: []interface{}{
			map[string]interface{}{
				"type": "text",
				"text": string(resultJSON),
			},
		},
	}, nil
}

func (s *SRSServer) handleRateCard(ctx context.Context, args map[string]interface{}) (*mcp.CallToolResult, error) {
	filePath, ok := args["file_path"].(string)
	if !ok || filePath == "" {
		return &mcp.CallToolResult{
			IsError: true,
			Content: []interface{}{
				map[string]interface{}{
					"type": "text", 
					"text": "file_path is required",
				},
			},
		}, nil
	}
	
	ratingStr, ok := args["rating"].(string)
	if !ok {
		if ratingFloat, ok := args["rating"].(float64); ok {
			ratingStr = strconv.Itoa(int(ratingFloat))
		} else {
			return &mcp.CallToolResult{
				IsError: true,
				Content: []interface{}{
					map[string]interface{}{
						"type": "text",
						"text": "rating is required (1-4)",
					},
				},
			}, nil
		}
	}
	
	rating, err := strconv.Atoi(ratingStr)
	if err != nil || rating < 1 || rating > 4 {
		return &mcp.CallToolResult{
			IsError: true,
			Content: []interface{}{
				map[string]interface{}{
					"type": "text",
					"text": "rating must be an integer between 1-4",
				},
			},
		}, nil
	}
	
	// Resolve file path if it's relative
	if !filepath.IsAbs(filePath) {
		filePath = filepath.Join(s.config.BaseDeckPath, filePath)
	}
	
	card, err := core.ParseCard(filePath)
	if err != nil {
		return &mcp.CallToolResult{
			IsError: true,
			Content: []interface{}{
				map[string]interface{}{
					"type": "text",
					"text": fmt.Sprintf("Error parsing card: %v", err),
				},
			},
		}, nil
	}
	
	fsrsRating, err := core.RatingFromInt(rating)
	if err != nil {
		return &mcp.CallToolResult{
			IsError: true,
			Content: []interface{}{
				map[string]interface{}{
					"type": "text",
					"text": fmt.Sprintf("Invalid rating: %v", err),
				},
			},
		}, nil
	}
	
	// Create a temporary session to rate the card
	session := core.NewReviewSession([]*core.Card{card})
	err = session.RateCard(fsrsRating)
	if err != nil {
		return &mcp.CallToolResult{
			IsError: true,
			Content: []interface{}{
				map[string]interface{}{
					"type": "text",
					"text": fmt.Sprintf("Error rating card: %v", err),
				},
			},
		}, nil
	}
	
	result := map[string]interface{}{
		"success": true,
		"card_path": filePath,
		"rating": core.RatingToString(fsrsRating),
		"new_due_date": card.FSRSCard.Due,
		"new_state": core.StateToString(card.FSRSCard.State),
		"reps": card.FSRSCard.Reps,
		"difficulty": card.FSRSCard.Difficulty,
		"stability": card.FSRSCard.Stability,
	}
	
	resultJSON, _ := json.MarshalIndent(result, "", "  ")
	
	return &mcp.CallToolResult{
		Content: []interface{}{
			map[string]interface{}{
				"type": "text",
				"text": string(resultJSON),
			},
		},
	}, nil
}

func (s *SRSServer) handleGetDeckStats(ctx context.Context, args map[string]interface{}) (*mcp.CallToolResult, error) {
	deckPath := "."
	if path, ok := args["deck_path"].(string); ok && path != "" {
		deckPath = path
	}
	
	resolvedPath, err := core.ResolveDeckPath(deckPath, s.config)
	if err != nil {
		return &mcp.CallToolResult{
			IsError: true,
			Content: []interface{}{
				map[string]interface{}{
					"type": "text",
					"text": fmt.Sprintf("Error resolving deck path: %v", err),
				},
			},
		}, nil
	}
	
	cards, err := core.FindCards(resolvedPath)
	if err != nil {
		return &mcp.CallToolResult{
			IsError: true,
			Content: []interface{}{
				map[string]interface{}{
					"type": "text",
					"text": fmt.Sprintf("Error loading cards: %v", err),
				},
			},
		}, nil
	}
	
	stats := core.GetDeckStats(cards)
	
	result := map[string]interface{}{
		"deck_path": deckPath,
		"total_cards": stats.TotalCards,
		"due_cards": stats.DueCards,
		"new_cards": stats.NewCards,
		"learning_cards": stats.LearningCards,
		"review_cards": stats.ReviewCards,
	}
	
	resultJSON, _ := json.MarshalIndent(result, "", "  ")
	
	return &mcp.CallToolResult{
		Content: []interface{}{
			map[string]interface{}{
				"type": "text",
				"text": string(resultJSON),
			},
		},
	}, nil
}

func (s *SRSServer) handleListDecks(ctx context.Context, args map[string]interface{}) (*mcp.CallToolResult, error) {
	deckTree, err := core.GetDeckTree(s.config.BaseDeckPath)
	if err != nil {
		return &mcp.CallToolResult{
			IsError: true,
			Content: []interface{}{
				map[string]interface{}{
					"type": "text",
					"text": fmt.Sprintf("Error getting deck tree: %v", err),
				},
			},
		}, nil
	}
	
	result := map[string]interface{}{
		"base_path": s.config.BaseDeckPath,
		"decks": deckTree,
	}
	
	resultJSON, _ := json.MarshalIndent(result, "", "  ")
	
	return &mcp.CallToolResult{
		Content: []interface{}{
			map[string]interface{}{
				"type": "text",
				"text": string(resultJSON),
			},
		},
	}, nil
}

func main() {
	srsServer, err := NewSRSServer()
	if err != nil {
		log.Fatalf("Failed to create SRS server: %v", err)
	}
	
	s := server.NewStdioServer(
		"srs-mcp-server",
		"1.0.0",
		server.WithRequestLogger(os.Stderr, true),
	)
	
	// Register tools
	s.AddTool("srs/get_due_cards", "Get cards that are due for review", map[string]interface{}{
		"type": "object",
		"properties": map[string]interface{}{
			"deck_path": map[string]interface{}{
				"type": "string",
				"description": "Path to deck (relative to base deck path, defaults to '.')",
			},
		},
	}, srsServer.handleGetDueCards)
	
	s.AddTool("srs/rate_card", "Rate a card and update its scheduling", map[string]interface{}{
		"type": "object",
		"properties": map[string]interface{}{
			"file_path": map[string]interface{}{
				"type": "string",
				"description": "Path to the card file",
			},
			"rating": map[string]interface{}{
				"type": "integer",
				"description": "Rating (1=Again, 2=Hard, 3=Good, 4=Easy)",
			},
		},
		"required": []string{"file_path", "rating"},
	}, srsServer.handleRateCard)
	
	s.AddTool("srs/get_deck_stats", "Get statistics for a deck", map[string]interface{}{
		"type": "object",
		"properties": map[string]interface{}{
			"deck_path": map[string]interface{}{
				"type": "string",
				"description": "Path to deck (relative to base deck path, defaults to '.')",
			},
		},
	}, srsServer.handleGetDeckStats)
	
	s.AddTool("srs/list_decks", "List all available decks", map[string]interface{}{
		"type": "object",
		"properties": map[string]interface{}{},
	}, srsServer.handleListDecks)
	
	if err := s.Serve(); err != nil {
		log.Fatalf("Server error: %v", err)
	}
}