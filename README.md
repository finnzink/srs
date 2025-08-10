# SRS - Unix-style Spaced Repetition System

A terminal-based spaced repetition system with modular architecture, interactive TUI, and MCP server for AI integration.

## Features

- 🧠 **FSRS Algorithm** - Uses the state-of-the-art Free Spaced Repetition Scheduler
- 📝 **Markdown Cards** - Cards are plain markdown files with syntax highlighting
- 📁 **Folder-based Decks** - Organize cards in directories, any structure you want
- ⌨️ **Interactive TUI** - Clean, full-screen review experience with live editing
- 🤖 **Built-in MCP Server** - AI integration included in single binary
- 🔧 **Modular Architecture** - Shared core library for consistent behavior
- 🧪 **Comprehensive Tests** - Well-tested core logic
- 📊 **Git Integration** - Version control your learning with git
- 🚀 **Cross-platform** - Linux, macOS, Windows

## Architecture

The SRS system is built with a modular architecture:

```
srs/
├── core/          # Shared business logic library
│   ├── card.go    # Card parsing and management
│   ├── deck.go    # Deck operations and configuration
│   ├── scheduler.go # FSRS scheduling logic
│   └── types.go   # Shared types and interfaces
├── tui/           # Terminal UI implementation
├── mcp_simple.go  # Built-in MCP server for AI integration
└── testdata/      # Test fixtures and examples
```

## Installation

### Quick Install (Linux/macOS)
```bash
curl -sSL https://raw.githubusercontent.com/finnzink/srs/main/install.sh | bash
```

### Build from Source
```bash
git clone https://github.com/finnzink/srs
cd srs
go build -o srs .
```

## Quick Start

1. **Configure your base deck:**
   ```bash
   ./srs config  # Set up your deck directory
   ```

2. **Create your first card:**
   ```bash
   cat > ~/flashcards/example.md << 'EOF'
   What is the time complexity of binary search?
   ---
   O(log n) - because we eliminate half the search space with each comparison.
   EOF
   ```

3. **Start reviewing:**
   ```bash
   ./srs review          # Interactive TUI mode
   ./srs review spanish  # Review specific subdeck
   ```

## Usage

### Commands

```bash
./srs review [DECK]    # Start interactive review session
./srs list [DECK]      # Show deck tree with due dates and stats  
./srs config           # Set up base deck directory
./srs mcp              # Start MCP server for AI integration
./srs version          # Show version information
./srs update           # Update to latest version
```

### Review Interface

**Interactive TUI Mode (default):**
- Type your answer before revealing the correct answer
- Rate cards with 1-4 keys
- Edit cards live with 'e' key
- Navigate with arrow keys, quit with 'q'

**Rating Scale:**
- **1** = Again (forgot completely)
- **2** = Hard (recalled with difficulty) 
- **3** = Good (recalled correctly)
- **4** = Easy (recalled easily)

### Card Format

Cards are simple markdown files:

```markdown
What is the time complexity of binary search?
---
O(log n) - because we eliminate half the search space with each comparison.
```

The system automatically manages FSRS metadata:

```markdown
<!-- FSRS: due:2025-01-15T10:30:00Z, stability:2.50, difficulty:5.00, elapsed_days:1, scheduled_days:3, reps:2, lapses:0, state:Review -->

What is the time complexity of binary search?
---
O(log n) - because we eliminate half the search space with each comparison.
```

### Deck Organization

Organize your cards however you like:

```
flashcards/
├── programming/
│   ├── algorithms/
│   │   ├── sorting.md
│   │   └── searching.md
│   └── languages/
│       ├── go.md
│       └── python.md
├── math/
│   ├── calculus.md
│   └── algebra.md
└── languages/
    ├── spanish/
    │   ├── verbs.md
    │   └── vocabulary.md
    └── french/
        └── basics.md
```

## MCP Server Integration

The MCP (Model Context Protocol) server enables AI agents to interact with your flashcards programmatically.

### Start the MCP Server

```bash
./srs mcp
```

### Available MCP Tools

- **`srs/get_due_cards`** - Get cards that are due for review
- **`srs/rate_card`** - Rate a card (1=Again, 2=Hard, 3=Good, 4=Easy)  
- **`srs/get_deck_stats`** - Get statistics for a deck
- **`srs/list_decks`** - List all available decks with statistics

### Example MCP Usage

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

```json
{
  "method": "tools/call", 
  "params": {
    "name": "srs/rate_card",
    "arguments": {
      "file_path": "spanish/verbs.md",
      "rating": 3
    }
  }
}
```

### Configure with Claude Code

```bash
# Add SRS MCP server to Claude Code
claude mcp add srs -- /path/to/srs mcp

# Now AI agents can access your flashcards!
```

### AI Integration Use Cases

- **Automated Review Sessions** - Have AI agents review cards based on performance
- **Content Analysis** - Analyze card difficulty and suggest improvements
- **Progress Tracking** - Generate learning analytics and insights
- **Smart Scheduling** - AI-driven scheduling recommendations
- **Bulk Operations** - Process multiple cards efficiently

## Development

### Running Tests

```bash
# Test core library
cd core && go test -v

# Test all components  
go test -v ./core/...
```

### Project Structure

- **core/**: Shared business logic, fully tested and reusable  
- **tui/**: Bubble Tea-based terminal interface
- **mcp_simple.go**: Built-in MCP server for AI integration
- **main.go**: Unified command-line interface

## Migration from Previous Version

**⚠️ Breaking Changes:**

This version removes the turn-based CLI review functionality in favor of a unified interactive experience.

### What Changed
- ✅ **Removed**: `srs review -r <rating>` turn-based workflow  
- ✅ **New**: All reviews use interactive TUI interface
- ✅ **New**: Built-in MCP server for AI integration (single binary!)
- ✅ **New**: Modular architecture with shared core library
- ✅ **New**: Comprehensive test suite
- ✅ **Improved**: Better error handling and user experience

### What Stayed the Same
- Card format and FSRS metadata (fully compatible)
- Deck organization and configuration
- FSRS scheduling algorithm and behavior
- Command-line interface for non-review operations
- All your existing cards work without changes

### Migration Steps

1. **No action required** - your existing cards and configuration will work
2. **Update workflows** - replace turn-based review scripts with interactive sessions  
3. **Try AI integration** - use `srs mcp` command and configure with Claude Code for enhanced workflows

## Configuration

Configuration is stored as JSON in your home directory:

```json
{
  "base_deck_path": "/home/user/flashcards"
}
```

Environment variables:
- `EDITOR` - Your preferred text editor (default: vim)
- `VISUAL` - Alternative editor variable

## Advanced Usage

### Review Specific Decks

```bash
./srs review programming        # Review programming cards
./srs review spanish/grammar    # Review specific subdeck
```

### Deck Statistics

```bash
./srs list                      # Show all decks with stats
./srs list programming          # Show programming subdeck stats
```

### Integration with Git

```bash
# Version control your learning
git add flashcards/
git commit -m "Add new algorithm cards"

# Review learning progress over time
git log --oneline flashcards/
git diff HEAD~10 flashcards/
```

## Best Practices

### Creating Effective Cards

- **Be concise** - Keep answers brief but complete
- **One concept per card** - Atomic knowledge units
- **Use active recall** - Frame as specific questions
- **Include context** - Add examples or explanations when helpful
- **Test edge cases** - Create cards for common mistakes

### Example Cards

**Programming Concept:**
```markdown
How do you declare a slice in Go?
---
Using `make()`: `s := make([]int, 0, 10)` (length 0, capacity 10)
Or slice literal: `s := []int{1, 2, 3}`
```

**Algorithm:**
```markdown
What are the steps of quicksort?
---
1. Choose a pivot element
2. Partition array around pivot
3. Recursively sort left and right subarrays
Time: O(n log n) average, O(n²) worst case
```

## Why This Architecture?

This unified design provides:

- **Simplicity** - Single binary with all functionality included
- **Consistency** - Same logic across TUI and MCP interfaces  
- **Testability** - Core logic is thoroughly tested
- **AI-Ready** - Built-in MCP server for seamless AI integration
- **Maintainability** - Clean separation of concerns

Perfect for developers who want both interactive review and programmatic access to their learning data.

## License

MIT License - see LICENSE file for details.

## Contributing

Issues and pull requests welcome! 

**Development Guidelines:**
1. Core logic goes in `core/` package with tests
2. Ensure both TUI and MCP server use shared core
3. Follow existing patterns for error handling
4. Add tests for new functionality

This tool follows Unix philosophy: do one thing well, use plain text, and be composable.