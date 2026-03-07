package tui

import (
	"fmt"
	"strings"

	"charm.land/bubbles/v2/textinput"
	tea "charm.land/bubbletea/v2"

	"github.com/onikukiraii/rikugan/internal/diff"
)

// FilePicker provides fuzzy file search and selection.
type FilePicker struct {
	input    textinput.Model
	active   bool
	files    []diff.DiffFile
	filtered []int // indices into files
	cursor   int
	comments map[diff.LineKey]string
}

// NewFilePicker creates a new file picker.
func NewFilePicker() FilePicker {
	ti := textinput.New()
	ti.Prompt = "  file> "
	ti.CharLimit = 128
	return FilePicker{input: ti}
}

// Open activates the file picker.
func (fp *FilePicker) Open(files []diff.DiffFile, comments map[diff.LineKey]string) tea.Cmd {
	fp.active = true
	fp.files = files
	fp.comments = comments
	fp.input.SetValue("")
	fp.cursor = 0
	fp.updateFilter()
	return fp.input.Focus()
}

// Close deactivates the file picker.
func (fp *FilePicker) Close() {
	fp.active = false
	fp.input.Blur()
	fp.input.SetValue("")
}

// Update handles input. Returns (selectedIndex, confirmed, cmd).
func (fp *FilePicker) Update(msg tea.Msg) (int, bool, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.KeyPressMsg:
		switch msg.String() {
		case "enter":
			if len(fp.filtered) > 0 {
				idx := fp.filtered[fp.cursor]
				fp.Close()
				return idx, true, nil
			}
			fp.Close()
			return -1, false, nil
		case "esc":
			fp.Close()
			return -1, false, nil
		case "down", "ctrl+n":
			if fp.cursor < len(fp.filtered)-1 {
				fp.cursor++
			}
			return -1, false, nil
		case "up", "ctrl+p":
			if fp.cursor > 0 {
				fp.cursor--
			}
			return -1, false, nil
		}
	}

	prevVal := fp.input.Value()
	var cmd tea.Cmd
	fp.input, cmd = fp.input.Update(msg)
	if fp.input.Value() != prevVal {
		fp.updateFilter()
	}
	return -1, false, cmd
}

func (fp *FilePicker) updateFilter() {
	query := strings.ToLower(fp.input.Value())
	fp.filtered = nil
	for i, f := range fp.files {
		name := strings.ToLower(f.DisplayName())
		if query == "" || fuzzyMatch(name, query) {
			fp.filtered = append(fp.filtered, i)
		}
	}
	if fp.cursor >= len(fp.filtered) {
		fp.cursor = max(0, len(fp.filtered)-1)
	}
}

// fuzzyMatch checks if all characters in pattern appear in s in order.
func fuzzyMatch(s, pattern string) bool {
	pi := 0
	for i := 0; i < len(s) && pi < len(pattern); i++ {
		if s[i] == pattern[pi] {
			pi++
		}
	}
	return pi == len(pattern)
}

// View renders the file picker overlay.
func (fp *FilePicker) View(width int) string {
	if !fp.active {
		return ""
	}

	var sb strings.Builder
	sb.WriteString(fp.input.View())
	sb.WriteString("\n")

	maxShow := 10
	if len(fp.filtered) < maxShow {
		maxShow = len(fp.filtered)
	}

	for i := 0; i < maxShow; i++ {
		idx := fp.filtered[i]
		name := fp.files[idx].DisplayName()

		commentCount := 0
		for key := range fp.comments {
			if key.FileIndex == idx {
				commentCount++
			}
		}

		line := name
		if commentCount > 0 {
			line += fmt.Sprintf(" (%d)", commentCount)
		}

		if i == fp.cursor {
			sb.WriteString(styleFileTabActive.Render("  > " + line))
		} else {
			sb.WriteString(styleFileTab.Render("    " + line))
		}
		sb.WriteString("\n")
	}

	if len(fp.filtered) == 0 {
		sb.WriteString(styleContext.Render("    No matching files"))
		sb.WriteString("\n")
	}

	return sb.String()
}
