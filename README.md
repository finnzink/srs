# SRS - Unix-style Spaced Repetition System

A terminal-based spaced repetition system that follows Unix principles. Cards are stored as plain markdown files with FSRS metadata, making them git-friendly and easily editable.

## Features

- 🧠 **FSRS Algorithm** - Uses the state-of-the-art Free Spaced Repetition Scheduler
- 📝 **Markdown Cards** - Cards are plain markdown files with syntax highlighting
- 📁 **Folder-based Decks** - Organize cards in directories, any structure you want
- ⌨️ **Terminal Interface** - Clean, distraction-free review experience
- ✏️ **Live Editing** - Edit cards during review with your preferred editor
- 🔧 **Unix Philosophy** - Simple commands, plain text files, composable tools
- 📊 **Git Integration** - Version control your learning with git
- 🚀 **Cross-platform** - Linux, macOS, Windows

## Installation

### Quick Install (Linux/macOS)
```bash
curl -sSL https://raw.githubusercontent.com/finnzink/srs/main/install.sh | bash
```

### Manual Download
Download the binary for your platform from [releases](https://github.com/finnzink/srs/releases):

**Linux:**
```bash
curl -L https://github.com/finnzink/srs/releases/latest/download/srs-linux-amd64 -o srs
chmod +x srs
sudo mv srs /usr/local/bin/
```

**macOS:**
```bash
curl -L https://github.com/finnzink/srs/releases/latest/download/srs-darwin-arm64 -o srs
chmod +x srs
sudo mv srs /usr/local/bin/
```

### Build from Source
```bash
git clone https://github.com/finnzink/srs
cd srs
go build -o srs
```

## Quick Start

1. **Create a deck directory:**
   ```bash
   mkdir my-deck
   ```

2. **Create your first card:**
   ```bash
   cat > my-deck/example.md << 'EOF'
   # What is the capital of France?
   
   ---
   
   # Paris
   
   The capital and largest city of France.
   EOF
   ```

3. **Configure your base deck:**
   ```bash
   srs config  # Set my-deck as base directory
   ```

4. **Start reviewing:**
   ```bash
   srs review          # Turn-based mode (default)
   srs -i review       # Interactive TUI mode
   ```

## Usage

### Commands

```bash
srs review [DECK] [RATING]  # Show next card (turn-based) or rate current card
srs list [DECK]             # Show deck tree with due dates and stats
srs -i review [DECK]        # Interactive TUI mode (full-screen interface)
srs config                  # Set up base deck directory
```

### Card Format

Cards are markdown files with optional FSRS metadata:

```markdown
<!-- FSRS: due:2024-01-01T00:00:00Z, stability:1.00, difficulty:5.00, ... -->
# Question

What is the time complexity of binary search?

---

# Answer

**O(log n)**

Binary search eliminates half of the remaining elements in each step:

```python
def binary_search(arr, target):
    left, right = 0, len(arr) - 1
    while left <= right:
        mid = (left + right) // 2
        if arr[mid] == target:
            return mid
        elif arr[mid] < target:
            left = mid + 1
        else:
            right = mid - 1
    return -1
```
```

### Review Modes

**Turn-based Mode (Default):**
```bash
srs review          # Show next due card
srs review 3        # Rate current card as "Good" and show next
srs review spanish  # Review cards from spanish subdeck
srs review spanish 2  # Rate card in spanish subdeck as "Hard"
```

**Interactive TUI Mode:**
```bash
srs -i review       # Full-screen interface with live typing
```

**Rating Scale:**
- **1** = Again (forgot completely)
- **2** = Hard (recalled with difficulty) 
- **3** = Good (recalled correctly)
- **4** = Easy (recalled easily)

**Interactive Mode Keys:**
- **1-4** = Rate card
- **e** = Edit card in your editor
- **q** = Quit session

### Deck Organization

Organize your cards however you like:

```
my-decks/
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

## Examples

### Programming Card
```markdown
# How do you declare a slice in Go?

---

## Slice Declaration

Using make:
```go
s := make([]int, 0, 10)  // length 0, capacity 10
```

Slice literal:
```go
s := []int{1, 2, 3, 4, 5}
```

From array:
```go
arr := [5]int{1, 2, 3, 4, 5}
s := arr[1:4]  // elements 1, 2, 3
```
```

### Batch Review with Scripts
```bash
# Review all cards in turn-based mode
srs review

# Review cards from specific subdeck
srs review programming

# Rate current card and continue (in a script)
srs review programming 3  # Rate as "Good"
srs review programming 4  # Rate next as "Easy"
```

### Integration with Git
```bash
# Version control your learning
git add my-deck/
git commit -m "Add new algorithm cards"

# Review what changed
git log --oneline my-deck/

# See FSRS progress over time
git diff HEAD~10 my-deck/
```

## Configuration

The app respects standard Unix environment variables:

- `EDITOR` - Your preferred text editor (default: vim)
- `VISUAL` - Alternative editor variable

## Why SRS?

This tool was built for developers who want:

- **Simple text files** instead of proprietary databases
- **Git integration** for versioning and syncing
- **Terminal workflow** that fits into existing development environment
- **Extensibility** through Unix pipes and scripts
- **No vendor lock-in** - your data is portable markdown

Perfect for learning on remote development servers, SSH sessions, or anywhere you have a terminal.

## License

MIT License - see LICENSE file for details.

## Contributing

Issues and pull requests welcome! This tool follows Unix philosophy: do one thing well, use plain text, be composable.