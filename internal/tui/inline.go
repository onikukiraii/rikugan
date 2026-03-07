package tui

import (
	"fmt"
	"strings"

	"github.com/onikukiraii/rikugan/internal/diff"
)

// InlineModel handles the inline diff view.
type InlineModel struct {
	cursor      int
	offset      int
	height      int
	width       int
	lines       []renderedLine
	highlighter Highlighter
}

type renderedLine struct {
	key     diff.LineKey
	isHunk  bool
	content string // pre-rendered content (without cursor highlight)
	rawLine *diff.DiffLine
}

// NewInlineModel creates a new inline view model.
func NewInlineModel() InlineModel {
	return InlineModel{}
}

// BuildLines constructs the renderable lines from a DiffFile.
func (m *InlineModel) BuildLines(file diff.DiffFile, fileIdx int) {
	m.lines = nil
	m.highlighter = NewHighlighter(file.DisplayName())
	for hi, h := range file.Hunks {
		header := fmt.Sprintf("@@ -%d,%d +%d,%d @@", h.OldStart, h.OldCount, h.NewStart, h.NewCount)
		if h.Header != "" {
			header += " " + h.Header
		}
		m.lines = append(m.lines, renderedLine{
			isHunk:  true,
			content: styleHunkHeader.Render(header),
		})

		for li, line := range h.Lines {
			key := diff.LineKey{FileIndex: fileIdx, HunkIndex: hi, LineIndex: li}
			m.lines = append(m.lines, renderedLine{
				key:     key,
				rawLine: &file.Hunks[hi].Lines[li],
				content: m.renderDiffLine(line),
			})
		}
	}
	m.cursor = 0
	m.offset = 0
}

func (m *InlineModel) renderDiffLine(line diff.DiffLine) string {
	var oldNum, newNum string
	if line.OldNum > 0 {
		oldNum = fmt.Sprintf("%4d", line.OldNum)
	} else {
		oldNum = "    "
	}
	if line.NewNum > 0 {
		newNum = fmt.Sprintf("%4d", line.NewNum)
	} else {
		newNum = "    "
	}

	highlighted := m.highlighter.Highlight(line.Content)
	if highlighted == "" {
		highlighted = line.Content
	}

	switch line.Type {
	case diff.LineAdded:
		nums := styleLineNumAdded.Render(oldNum) + " " + styleLineNumAdded.Render(newNum) + " "
		return styleBgAdded.Render(nums + highlighted)
	case diff.LineRemoved:
		nums := styleLineNumRemoved.Render(oldNum) + " " + styleLineNumRemoved.Render(newNum) + " "
		return styleBgRemoved.Render(nums + highlighted)
	default:
		nums := styleLineNum.Render(oldNum) + " " + styleLineNum.Render(newNum) + " "
		return nums + highlighted
	}
}

// SetSize updates the viewport dimensions.
func (m *InlineModel) SetSize(w, h int) {
	m.width = w
	m.height = h
}

// MoveUp moves the cursor up.
func (m *InlineModel) MoveUp(n int) {
	m.cursor -= n
	if m.cursor < 0 {
		m.cursor = 0
	}
	m.ensureVisible()
}

// MoveDown moves the cursor down.
func (m *InlineModel) MoveDown(n int) {
	m.cursor += n
	if m.cursor >= len(m.lines) {
		m.cursor = len(m.lines) - 1
	}
	if m.cursor < 0 {
		m.cursor = 0
	}
	m.ensureVisible()
}

// GoTop moves to the first line.
func (m *InlineModel) GoTop() {
	m.cursor = 0
	m.offset = 0
}

// GoBottom moves to the last line.
func (m *InlineModel) GoBottom() {
	m.cursor = len(m.lines) - 1
	m.ensureVisible()
}

// NextHunk moves to the next hunk header.
func (m *InlineModel) NextHunk() {
	for i := m.cursor + 1; i < len(m.lines); i++ {
		if m.lines[i].isHunk {
			m.cursor = i
			m.ensureVisible()
			return
		}
	}
}

// PrevHunk moves to the previous hunk header.
func (m *InlineModel) PrevHunk() {
	for i := m.cursor - 1; i >= 0; i-- {
		if m.lines[i].isHunk {
			m.cursor = i
			m.ensureVisible()
			return
		}
	}
}

func (m *InlineModel) ensureVisible() {
	if m.cursor < m.offset {
		m.offset = m.cursor
	}
	if m.cursor >= m.offset+m.height {
		m.offset = m.cursor - m.height + 1
	}
}

// CurrentLineKey returns the LineKey at the cursor, if available.
func (m *InlineModel) CurrentLineKey() (diff.LineKey, bool) {
	if m.cursor < 0 || m.cursor >= len(m.lines) {
		return diff.LineKey{}, false
	}
	line := m.lines[m.cursor]
	if line.isHunk {
		return diff.LineKey{}, false
	}
	return line.key, true
}

// View renders the inline diff view.
func (m *InlineModel) View(comments map[diff.LineKey]string) string {
	if len(m.lines) == 0 {
		return styleContext.Render("  No diff content")
	}

	var sb strings.Builder
	end := m.offset + m.height
	if end > len(m.lines) {
		end = len(m.lines)
	}

	for i := m.offset; i < end; i++ {
		line := m.lines[i]
		content := line.content

		if i == m.cursor {
			content = styleCursorLine.Render(content)
		}

		sb.WriteString(content)
		sb.WriteString("\n")

		if !line.isHunk {
			if comment, ok := comments[line.key]; ok {
				indicator := styleCommentIndicator.Render("  ▶ ")
				sb.WriteString(indicator + styleComment.Render(comment) + "\n")
			}
		}
	}

	return sb.String()
}
