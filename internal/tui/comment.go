package tui

import (
	"charm.land/bubbles/v2/textinput"
	tea "charm.land/bubbletea/v2"

	"github.com/onikukiraii/rikugan/internal/diff"
)

// CommentEditor manages the inline comment input.
type CommentEditor struct {
	input   textinput.Model
	active  bool
	lineKey diff.LineKey
}

// NewCommentEditor creates a new comment editor.
func NewCommentEditor() CommentEditor {
	ti := textinput.New()
	ti.Prompt = "  comment> "
	ti.CharLimit = 256
	return CommentEditor{input: ti}
}

// Open starts editing a comment for the given line.
func (c *CommentEditor) Open(key diff.LineKey, existing string) tea.Cmd {
	c.active = true
	c.lineKey = key
	c.input.SetValue(existing)
	return c.input.Focus()
}

// Close stops editing.
func (c *CommentEditor) Close() {
	c.active = false
	c.input.Blur()
	c.input.SetValue("")
}

// Update handles messages while the editor is active.
func (c *CommentEditor) Update(msg tea.Msg) (string, bool, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.KeyPressMsg:
		switch msg.String() {
		case "enter":
			val := c.input.Value()
			c.Close()
			return val, true, nil
		case "esc":
			c.Close()
			return "", false, nil
		}
	}
	var cmd tea.Cmd
	c.input, cmd = c.input.Update(msg)
	return "", false, cmd
}

// View renders the comment input.
func (c *CommentEditor) View() string {
	if !c.active {
		return ""
	}
	return c.input.View()
}
