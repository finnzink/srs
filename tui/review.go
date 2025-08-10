package tui

import (
	"fmt"
	"strings"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
	"github.com/open-spaced-repetition/go-fsrs/v3"

	"srs/core"
)

type reviewState int

const (
	showingQuestion reviewState = iota
	showingAnswer
)

type ReviewModel struct {
	session     *core.ReviewSession
	currentCard *core.Card
	state       reviewState
	userAnswer  string
	width       int
	height      int
	quitting    bool
	message     string
	scroll      int
}

var (
	questionStyle = lipgloss.NewStyle().
		BorderStyle(lipgloss.RoundedBorder()).
		BorderForeground(lipgloss.Color("62")).
		Padding(0, 1)

	answerStyle = lipgloss.NewStyle().
		BorderStyle(lipgloss.RoundedBorder()).
		BorderForeground(lipgloss.Color("42")).
		Padding(0, 1)

	userAnswerStyle = lipgloss.NewStyle().
		BorderStyle(lipgloss.RoundedBorder()).
		BorderForeground(lipgloss.Color("208")).
		Padding(0, 1)

	promptStyle = lipgloss.NewStyle().
		Foreground(lipgloss.Color("241"))

	helpStyle = lipgloss.NewStyle().
		Foreground(lipgloss.Color("241"))
)

func NewReviewModel(session *core.ReviewSession) (ReviewModel, error) {
	card, err := session.CurrentCard()
	if err != nil {
		return ReviewModel{}, err
	}
	
	return ReviewModel{
		session:     session,
		currentCard: card,
		state:       showingQuestion,
	}, nil
}

func (m ReviewModel) Init() tea.Cmd {
	return nil
}

func (m ReviewModel) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.WindowSizeMsg:
		m.width = msg.Width
		m.height = msg.Height
		return m, nil

	case tea.KeyMsg:
		switch m.state {
		case showingQuestion:
			switch msg.String() {
			case "ctrl+c":
				m.quitting = true
				return m, tea.Quit
			case "enter":
				m.state = showingAnswer
			case "backspace":
				if len(m.userAnswer) > 0 {
					m.userAnswer = m.userAnswer[:len(m.userAnswer)-1]
				}
			case "up":
				if m.scroll > 0 {
					m.scroll--
				}
			case "down":
				m.scroll++
			default:
				if len(msg.String()) == 1 {
					m.userAnswer += msg.String()
				}
			}

		case showingAnswer:
			switch msg.String() {
			case "q", "ctrl+c":
				m.quitting = true
				return m, tea.Quit
			case "1":
				return m.rateCard(fsrs.Again)
			case "2":
				return m.rateCard(fsrs.Hard)
			case "3":
				return m.rateCard(fsrs.Good)
			case "4":
				return m.rateCard(fsrs.Easy)
			case "e", "E":
				m.quitting = true
				m.message = fmt.Sprintf("edit_card:%s:%d", m.userAnswer, int(m.state))
				return m, tea.Quit
			case "up":
				if m.scroll > 0 {
					m.scroll--
				}
			case "down":
				m.scroll++
			}
		}
	}

	return m, nil
}

func (m ReviewModel) rateCard(rating fsrs.Rating) (tea.Model, tea.Cmd) {
	err := m.session.RateCard(rating)
	if err != nil {
		m.message = fmt.Sprintf("Error rating card: %v", err)
		return m, nil
	}

	// Move to next card
	if !m.session.HasNext() {
		m.quitting = true
		return m, tea.Quit
	}

	// Get next card
	nextCard, err := m.session.CurrentCard()
	if err != nil {
		m.message = fmt.Sprintf("Error getting next card: %v", err)
		m.quitting = true
		return m, tea.Quit
	}

	// Reset for next card
	m.currentCard = nextCard
	m.state = showingQuestion
	m.userAnswer = ""
	m.message = ""
	m.scroll = 0

	return m, nil
}

func (m ReviewModel) View() string {
	if m.quitting {
		current, total := m.session.Progress()
		if !m.session.HasNext() {
			return fmt.Sprintf("Session complete! Reviewed %d cards.\n", total)
		}
		return fmt.Sprintf("Session ended. Reviewed %d cards.\n", current-1)
	}

	// Calculate available height for content (leave room for header and help)
	contentHeight := m.height - 4
	if contentHeight < 1 {
		contentHeight = 1
	}

	var content []string

	// Question
	questionText := RenderMarkdown(m.currentCard.Question)
	question := questionStyle.Width(m.width - 4).Render(questionText)
	content = append(content, question)

	// User's answer (if any) - always show between question and answer
	if m.userAnswer != "" {
		userInput := userAnswerStyle.Width(m.width - 4).Render(
			m.userAnswer + func() string {
				if m.state == showingQuestion {
					return "█" // Show cursor when typing
				}
				return ""
			}(),
		)
		content = append(content, userInput)
	}

	// Answer (only in answer state)
	if m.state == showingAnswer {
		answerText := RenderMarkdown(m.currentCard.Answer)
		answer := answerStyle.Width(m.width - 4).Render(answerText)
		content = append(content, answer)
	}

	// Join content and handle scrolling
	fullContent := strings.Join(content, "\n")
	contentLines := strings.Split(fullContent, "\n")

	// Apply scrolling with bounds checking
	if len(contentLines) == 0 {
		contentLines = []string{""}
	}

	startLine := m.scroll
	if startLine < 0 {
		startLine = 0
	}
	if startLine >= len(contentLines) {
		startLine = len(contentLines) - 1
	}

	endLine := startLine + contentHeight
	if endLine > len(contentLines) {
		endLine = len(contentLines)
	}
	if endLine <= startLine {
		endLine = startLine + 1
		if endLine > len(contentLines) {
			endLine = len(contentLines)
		}
	}

	visibleContent := strings.Join(contentLines[startLine:endLine], "\n")

	// Header
	current, total := m.session.Progress()
	progress := fmt.Sprintf("Card %d of %d", current, total)
	header := lipgloss.NewStyle().Bold(true).Render(progress)

	// Help text based on state
	var help string
	switch m.state {
	case showingQuestion:
		if m.userAnswer != "" {
			help = "Enter = show answer • ↑/↓ = scroll • Backspace = delete • Ctrl+C = quit"
		} else {
			help = "Type answer or Enter to skip • ↑/↓ = scroll • Ctrl+C = quit"
		}
	case showingAnswer:
		help = "1 = Again • 2 = Hard • 3 = Good • 4 = Easy • ↑/↓ = scroll\ne = edit • q = quit"
	}

	helpText := helpStyle.Render(help)

	// Show scroll indicator if needed
	scrollIndicator := ""
	if len(contentLines) > contentHeight {
		scrollIndicator = fmt.Sprintf(" (%d/%d)", startLine+1, len(contentLines)-contentHeight+1)
	}

	result := header + scrollIndicator + "\n\n" + visibleContent + "\n\n" + helpText

	// Show message if any
	if m.message != "" {
		messageStyle := lipgloss.NewStyle().Foreground(lipgloss.Color("196"))
		result += "\n" + messageStyle.Render(m.message)
	}

	return result
}