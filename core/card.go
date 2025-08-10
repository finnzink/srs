package core

import (
	"bufio"
	"fmt"
	"os"
	"path/filepath"
	"regexp"
	"strconv"
	"strings"
	"time"

	"github.com/open-spaced-repetition/go-fsrs/v3"
)

// ParseCard reads and parses a markdown card file
func ParseCard(filePath string) (*Card, error) {
	file, err := os.Open(filePath)
	if err != nil {
		return nil, err
	}
	defer file.Close()

	var question, answer strings.Builder
	var fsrsMetadata string
	scanner := bufio.NewScanner(file)
	
	inAnswer := false
	
	for scanner.Scan() {
		line := scanner.Text()
		
		if strings.HasPrefix(line, "<!-- FSRS:") && strings.HasSuffix(line, "-->") {
			fsrsMetadata = strings.TrimSpace(strings.TrimPrefix(strings.TrimSuffix(line, "-->"), "<!-- FSRS:"))
			continue
		}
		
		if line == "---" && !inAnswer {
			inAnswer = true
			continue
		}
		
		if inAnswer {
			answer.WriteString(line + "\n")
		} else {
			question.WriteString(line + "\n")
		}
	}

	if err := scanner.Err(); err != nil {
		return nil, err
	}

	card := &Card{
		Question: strings.TrimSpace(question.String()),
		Answer:   strings.TrimSpace(answer.String()),
		FilePath: filePath,
	}

	if fsrsMetadata != "" {
		card.FSRSCard = parseFSRSMetadata(fsrsMetadata)
	} else {
		card.FSRSCard = fsrs.NewCard()
	}

	fileInfo, err := os.Stat(filePath)
	if err == nil {
		card.LastModified = fileInfo.ModTime()
	}

	return card, nil
}

// FindCards recursively finds all markdown cards in a directory
func FindCards(deckPath string) ([]*Card, error) {
	var cards []*Card
	
	err := filepath.Walk(deckPath, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}
		
		if !info.IsDir() && strings.HasSuffix(strings.ToLower(path), ".md") {
			card, err := ParseCard(path)
			if err != nil {
				fmt.Printf("Warning: failed to parse card %s: %v\n", path, err)
				return nil
			}
			cards = append(cards, card)
		}
		
		return nil
	})
	
	return cards, err
}

// GetDueCards filters cards that are due for review
func GetDueCards(cards []*Card) []*Card {
	now := time.Now()
	var dueCards []*Card
	
	for _, card := range cards {
		if card.FSRSCard.Due.Before(now) || card.FSRSCard.Due.Equal(now) {
			dueCards = append(dueCards, card)
		}
	}
	
	return dueCards
}

// UpdateFSRSMetadata writes the FSRS metadata back to the card file
func (c *Card) UpdateFSRSMetadata() error {
	content, err := os.ReadFile(c.FilePath)
	if err != nil {
		return err
	}

	lines := strings.Split(string(content), "\n")
	
	// Remove existing FSRS metadata
	var filteredLines []string
	for _, line := range lines {
		if !strings.HasPrefix(line, "<!-- FSRS:") {
			filteredLines = append(filteredLines, line)
		}
	}

	// Add new FSRS metadata at the top
	fsrsLine := fmt.Sprintf("<!-- FSRS: due:%s, stability:%.2f, difficulty:%.2f, elapsed_days:%d, scheduled_days:%d, reps:%d, lapses:%d, state:%s -->",
		c.FSRSCard.Due.Format(time.RFC3339),
		c.FSRSCard.Stability,
		c.FSRSCard.Difficulty,
		c.FSRSCard.ElapsedDays,
		c.FSRSCard.ScheduledDays,
		c.FSRSCard.Reps,
		c.FSRSCard.Lapses,
		StateToString(c.FSRSCard.State))

	newContent := fsrsLine + "\n" + strings.Join(filteredLines, "\n")
	
	return os.WriteFile(c.FilePath, []byte(newContent), 0644)
}

func parseFSRSMetadata(metadata string) fsrs.Card {
	card := fsrs.NewCard()
	
	re := regexp.MustCompile(`(\w+):([^,]+)`)
	matches := re.FindAllStringSubmatch(metadata, -1)
	
	for _, match := range matches {
		key := strings.TrimSpace(match[1])
		value := strings.TrimSpace(match[2])
		
		switch key {
		case "due":
			if t, err := time.Parse(time.RFC3339, value); err == nil {
				card.Due = t
			}
		case "stability":
			if f, err := strconv.ParseFloat(value, 64); err == nil {
				card.Stability = f
			}
		case "difficulty":
			if f, err := strconv.ParseFloat(value, 64); err == nil {
				card.Difficulty = f
			}
		case "elapsed_days":
			if i, err := strconv.Atoi(value); err == nil {
				card.ElapsedDays = uint64(i)
			}
		case "scheduled_days":
			if i, err := strconv.Atoi(value); err == nil {
				card.ScheduledDays = uint64(i)
			}
		case "reps":
			if i, err := strconv.Atoi(value); err == nil {
				card.Reps = uint64(i)
			}
		case "lapses":
			if i, err := strconv.Atoi(value); err == nil {
				card.Lapses = uint64(i)
			}
		case "state":
			card.State = StringToState(value)
		}
	}
	
	return card
}

// StateToString converts FSRS state to string
func StateToString(state fsrs.State) string {
	switch state {
	case fsrs.New:
		return "New"
	case fsrs.Learning:
		return "Learning"
	case fsrs.Review:
		return "Review"
	case fsrs.Relearning:
		return "Relearning"
	default:
		return "New"
	}
}

// StringToState converts string to FSRS state
func StringToState(s string) fsrs.State {
	switch s {
	case "New":
		return fsrs.New
	case "Learning":
		return fsrs.Learning
	case "Review":
		return fsrs.Review
	case "Relearning":
		return fsrs.Relearning
	default:
		return fsrs.New
	}
}