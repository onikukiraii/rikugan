package tui

import (
	"fmt"
	"strings"

	"charm.land/lipgloss/v2"

	"github.com/onikukiraii/rikugan/internal/diff"
)

// SplitModel handles the side-by-side diff view.
type SplitModel struct {
	cursor      int
	offset      int
	height      int
	width       int
	leftPane    []splitLine
	rightPane   []splitLine
	activePane  int // 0 = left (old), 1 = right (new)
	highlighter Highlighter
}

type splitLine struct {
	key      diff.LineKey
	lineNum  int
	content  string
	lineType diff.LineType
	empty    bool          // padding line for alignment
	isHunk   bool          // hunk separator line
	segments []DiffSegment // word-level diff segments (nil = no word diff)
}

// NewSplitModel creates a new split view model.
func NewSplitModel() SplitModel {
	return SplitModel{}
}

// BuildLines constructs the side-by-side view from a DiffFile.
func (m *SplitModel) BuildLines(file diff.DiffFile, fileIdx int) {
	m.leftPane = nil
	m.rightPane = nil
	m.highlighter = NewHighlighter(file.DisplayName())

	for hi, h := range file.Hunks {
		// Add hunk separator
		m.leftPane = append(m.leftPane, splitLine{empty: true, isHunk: true, content: "────"})
		m.rightPane = append(m.rightPane, splitLine{empty: true, isHunk: true, content: "────"})

		li := 0
		for li < len(h.Lines) {
			line := h.Lines[li]
			key := diff.LineKey{FileIndex: fileIdx, HunkIndex: hi, LineIndex: li}

			switch line.Type {
			case diff.LineContext:
				m.leftPane = append(m.leftPane, splitLine{
					key: key, lineNum: line.OldNum, content: line.Content, lineType: line.Type,
				})
				m.rightPane = append(m.rightPane, splitLine{
					key: key, lineNum: line.NewNum, content: line.Content, lineType: line.Type,
				})
				li++

			case diff.LineRemoved:
				// Collect consecutive removed lines
				removed := []int{li}
				for li+1 < len(h.Lines) && h.Lines[li+1].Type == diff.LineRemoved {
					li++
					removed = append(removed, li)
				}
				// Collect consecutive added lines
				var added []int
				for li+1 < len(h.Lines) && h.Lines[li+1].Type == diff.LineAdded {
					li++
					added = append(added, li)
				}
				// Pair them up with word-level diff
				maxLen := max(len(removed), len(added))
				for j := 0; j < maxLen; j++ {
					var leftLine, rightLine splitLine

					if j < len(removed) {
						idx := removed[j]
						k := diff.LineKey{FileIndex: fileIdx, HunkIndex: hi, LineIndex: idx}
						leftLine = splitLine{
							key: k, lineNum: h.Lines[idx].OldNum, content: h.Lines[idx].Content, lineType: diff.LineRemoved,
						}
					} else {
						leftLine = splitLine{empty: true}
					}

					if j < len(added) {
						idx := added[j]
						k := diff.LineKey{FileIndex: fileIdx, HunkIndex: hi, LineIndex: idx}
						rightLine = splitLine{
							key: k, lineNum: h.Lines[idx].NewNum, content: h.Lines[idx].Content, lineType: diff.LineAdded,
						}
					} else {
						rightLine = splitLine{empty: true}
					}

					// Compute word diff for paired lines
					if !leftLine.empty && !rightLine.empty {
						oldSegs, newSegs := ComputeWordDiff(leftLine.content, rightLine.content)
						leftLine.segments = oldSegs
						rightLine.segments = newSegs
					}

					m.leftPane = append(m.leftPane, leftLine)
					m.rightPane = append(m.rightPane, rightLine)
				}
				li++

			case diff.LineAdded:
				m.leftPane = append(m.leftPane, splitLine{empty: true})
				m.rightPane = append(m.rightPane, splitLine{
					key: key, lineNum: line.NewNum, content: line.Content, lineType: line.Type,
				})
				li++
			}
		}
	}
	m.cursor = 0
	m.offset = 0
}

// SetSize updates the viewport dimensions.
func (m *SplitModel) SetSize(w, h int) {
	m.width = w
	m.height = h
}

// MoveUp moves the cursor up.
func (m *SplitModel) MoveUp(n int) {
	m.cursor -= n
	if m.cursor < 0 {
		m.cursor = 0
	}
	m.ensureVisible()
}

// MoveDown moves the cursor down.
func (m *SplitModel) MoveDown(n int) {
	m.cursor += n
	if m.cursor >= len(m.leftPane) {
		m.cursor = len(m.leftPane) - 1
	}
	if m.cursor < 0 {
		m.cursor = 0
	}
	m.ensureVisible()
}

// GoTop moves to the first line.
func (m *SplitModel) GoTop() {
	m.cursor = 0
	m.offset = 0
}

// GoBottom moves to the last line.
func (m *SplitModel) GoBottom() {
	m.cursor = len(m.leftPane) - 1
	m.ensureVisible()
}

// NextHunk moves to the next hunk separator.
func (m *SplitModel) NextHunk() {
	for i := m.cursor + 1; i < len(m.leftPane); i++ {
		if m.leftPane[i].isHunk {
			m.cursor = i
			m.ensureVisible()
			return
		}
	}
}

// PrevHunk moves to the previous hunk separator.
func (m *SplitModel) PrevHunk() {
	for i := m.cursor - 1; i >= 0; i-- {
		if m.leftPane[i].isHunk {
			m.cursor = i
			m.ensureVisible()
			return
		}
	}
}

// TogglePane switches the active pane.
func (m *SplitModel) TogglePane(dir int) {
	m.activePane = dir
}

func (m *SplitModel) ensureVisible() {
	if m.cursor < m.offset {
		m.offset = m.cursor
	}
	if m.cursor >= m.offset+m.height {
		m.offset = m.cursor - m.height + 1
	}
}

// CurrentLineKey returns the LineKey at the cursor.
func (m *SplitModel) CurrentLineKey() (diff.LineKey, bool) {
	if m.cursor < 0 || m.cursor >= len(m.leftPane) {
		return diff.LineKey{}, false
	}
	if m.activePane == 0 {
		line := m.leftPane[m.cursor]
		if line.empty {
			return diff.LineKey{}, false
		}
		return line.key, true
	}
	line := m.rightPane[m.cursor]
	if line.empty {
		return diff.LineKey{}, false
	}
	return line.key, true
}

// View renders the split diff view.
func (m *SplitModel) View(comments map[diff.LineKey]string) string {
	if len(m.leftPane) == 0 {
		return styleContext.Render("  No diff content")
	}

	paneWidth := m.width/2 - 2
	if paneWidth < 10 {
		paneWidth = 10
	}

	var sb strings.Builder
	end := m.offset + m.height
	if end > len(m.leftPane) {
		end = len(m.leftPane)
	}

	for i := m.offset; i < end; i++ {
		left := m.renderSplitLine(m.leftPane[i], paneWidth, i == m.cursor && m.activePane == 0)
		right := m.renderSplitLine(m.rightPane[i], paneWidth, i == m.cursor && m.activePane == 1)

		separator := styleContext.Render("│")
		sb.WriteString(left + separator + right + "\n")

		// Show comments below the line
		for _, pane := range []splitLine{m.leftPane[i], m.rightPane[i]} {
			if !pane.empty {
				if comment, ok := comments[pane.key]; ok {
					indicator := styleCommentIndicator.Render("  ▶ ")
					sb.WriteString(indicator + styleComment.Render(comment) + "\n")
				}
			}
		}
	}

	return sb.String()
}

func (m *SplitModel) renderSplitLine(line splitLine, paneWidth int, isCursor bool) string {
	if line.empty {
		content := strings.Repeat(" ", paneWidth)
		if isCursor {
			return styleCursorLine.Render(content)
		}
		return content
	}

	var numStr string
	if line.lineNum > 0 {
		numStr = fmt.Sprintf("%4d ", line.lineNum)
	} else {
		numStr = "     "
	}

	highlighted := m.highlighter.Highlight(line.content)
	if highlighted == "" {
		highlighted = line.content
	}

	var codeContent string
	if line.segments != nil {
		// Word diff mode: highlight changed parts, syntax highlight unchanged parts
		var changedStyle lipgloss.Style
		if line.lineType == diff.LineAdded {
			changedStyle = styleWordDiffAdded
		} else {
			changedStyle = styleWordDiffRemoved
		}
		var segBuf strings.Builder
		for _, seg := range line.segments {
			if seg.Changed {
				segBuf.WriteString(changedStyle.Render(seg.Text))
			} else {
				segBuf.WriteString(m.highlighter.Highlight(seg.Text))
			}
		}
		codeContent = segBuf.String()
	} else {
		codeContent = highlighted
	}

	var numStyle lipgloss.Style
	switch line.lineType {
	case diff.LineAdded:
		numStyle = styleLineNumAdded
	case diff.LineRemoved:
		numStyle = styleLineNumRemoved
	default:
		numStyle = styleLineNum
	}

	content := numStyle.Render(numStr[:4]) + " " + codeContent

	// Truncate or pad to pane width
	contentWidth := lipgloss.Width(content)
	if contentWidth < paneWidth {
		content += strings.Repeat(" ", paneWidth-contentWidth)
	}

	// Apply diff background tint
	switch line.lineType {
	case diff.LineAdded:
		content = styleBgAdded.Render(content)
	case diff.LineRemoved:
		content = styleBgRemoved.Render(content)
	}

	if isCursor {
		content = styleCursorLine.Render(content)
	}
	return content
}
