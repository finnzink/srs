# SRS MCP Server

An MCP (Model Context Protocol) server for the SRS spaced repetition system. This allows AI models to interact with your flashcard decks programmatically.

## Installation

```bash
cd mcp
go mod tidy
go build -o srs-mcp-server
```

## Usage

The server provides the following tools:

### `srs/get_due_cards`
Get cards that are due for review.

Parameters:
- `deck_path` (optional): Path to deck relative to base deck path (defaults to ".")

### `srs/rate_card`  
Rate a card and update its scheduling.

Parameters:
- `file_path` (required): Path to the card file
- `rating` (required): Rating (1=Again, 2=Hard, 3=Good, 4=Easy)

### `srs/get_deck_stats`
Get statistics for a deck.

Parameters:
- `deck_path` (optional): Path to deck relative to base deck path (defaults to ".")

### `srs/list_decks`
List all available decks with their statistics.

No parameters required.

## Example Usage

```json
{
  "method": "tools/call",
  "params": {
    "name": "srs/get_due_cards",
    "arguments": {
      "deck_path": "spanish"
    }
  }
}
```

## Configuration

Make sure to run `srs config` to set up your base deck path before using the MCP server.