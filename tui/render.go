package tui

import (
	"fmt"
	"os"
	"strings"

	"github.com/charmbracelet/glamour"
)

type MarkdownRenderer struct {
	renderer *glamour.TermRenderer
}

func NewMarkdownRenderer() (*MarkdownRenderer, error) {
	renderer, err := glamour.NewTermRenderer(
		glamour.WithAutoStyle(),
		glamour.WithWordWrap(80),
	)
	if err != nil {
		return nil, err
	}

	return &MarkdownRenderer{
		renderer: renderer,
	}, nil
}

func (mr *MarkdownRenderer) Render(markdown string) (string, error) {
	rendered, err := mr.renderer.Render(markdown)
	if err != nil {
		return "", err
	}
	
	return strings.TrimSpace(rendered), nil
}

// Global renderer instance
var globalRenderer *MarkdownRenderer

func init() {
	var err error
	globalRenderer, err = NewMarkdownRenderer()
	if err != nil {
		fmt.Fprintf(os.Stderr, "Warning: Failed to initialize markdown renderer: %v\n", err)
		globalRenderer = nil
	}
}

func RenderMarkdown(markdown string) string {
	if globalRenderer == nil {
		return markdown
	}
	
	rendered, err := globalRenderer.Render(markdown)
	if err != nil {
		return markdown
	}
	
	return rendered
}