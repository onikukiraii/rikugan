package tui

import (
	"fmt"
	"strings"

	tea "charm.land/bubbletea/v2"
	"charm.land/lipgloss/v2"

	"github.com/onikukiraii/rikugan/internal"
	"github.com/onikukiraii/rikugan/internal/diff"
)

// ViewMode represents the diff display mode.
type ViewMode int

const (
	ModeInline ViewMode = iota
	ModeSplit
)

// Model is the root Bubble Tea model.
type Model struct {
	files     []diff.DiffFile
	fileIdx   int
	mode      ViewMode
	inline    InlineModel
	split     SplitModel
	comments  map[diff.LineKey]string
	editor    CommentEditor
	keys      KeyMap
	width     int
	height    int
	err       error
	showHelp  bool
	copied    bool
	gPressed  bool // for gg detection
}

// New creates a new Model.
func New(files []diff.DiffFile) Model {
	m := Model{
		files:    files,
		comments: make(map[diff.LineKey]string),
		editor:   NewCommentEditor(),
		keys:     DefaultKeyMap(),
		inline:   NewInlineModel(),
		split:    NewSplitModel(),
	}
	if len(files) > 0 {
		m.rebuildLines()
	}
	return m
}

// NewError creates a model displaying an error.
func NewError(err error) Model {
	return Model{err: err}
}

func (m *Model) rebuildLines() {
	if m.fileIdx >= len(m.files) {
		return
	}
	f := m.files[m.fileIdx]
	m.inline.BuildLines(f, m.fileIdx)
	m.split.BuildLines(f, m.fileIdx)
}

// Init implements tea.Model.
func (m Model) Init() tea.Cmd {
	return nil
}

// Update implements tea.Model.
func (m Model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	if m.err != nil {
		if msg, ok := msg.(tea.KeyPressMsg); ok {
			if msg.String() == "q" || msg.String() == "esc" {
				return m, tea.Quit
			}
		}
		return m, nil
	}

	// Handle comment editor input
	if m.editor.active {
		val, committed, cmd := m.editor.Update(msg)
		if committed && val != "" {
			m.comments[m.editor.lineKey] = val
		}
		return m, cmd
	}

	switch msg := msg.(type) {
	case tea.WindowSizeMsg:
		m.width = msg.Width
		m.height = msg.Height
		contentHeight := m.height - 4 // header + status + help
		m.inline.SetSize(m.width, contentHeight)
		m.split.SetSize(m.width, contentHeight)
		return m, nil

	case tea.KeyPressMsg:
		return m.handleKey(msg)
	}

	return m, nil
}

func (m Model) handleKey(msg tea.KeyPressMsg) (tea.Model, tea.Cmd) {
	key := msg.String()

	// Reset copied flag on any key
	if m.copied {
		m.copied = false
	}

	// Handle gg (go to top)
	if m.gPressed {
		m.gPressed = false
		if key == "g" {
			m.currentView().GoTop()
			return m, nil
		}
	}

	switch key {
	case m.keys.Quit, "esc":
		if m.showHelp {
			m.showHelp = false
			return m, nil
		}
		return m, tea.Quit

	case m.keys.Down:
		m.currentView().MoveDown(1)
	case m.keys.Up:
		m.currentView().MoveUp(1)
	case m.keys.HalfPageDn:
		m.currentView().MoveDown(m.height / 2)
	case m.keys.HalfPageUp:
		m.currentView().MoveUp(m.height / 2)

	case m.keys.Top:
		m.gPressed = true

	case m.keys.Bottom:
		m.currentView().GoBottom()

	case m.keys.NextFile:
		if m.fileIdx < len(m.files)-1 {
			m.fileIdx++
			m.rebuildLines()
		}
	case m.keys.PrevFile:
		if m.fileIdx > 0 {
			m.fileIdx--
			m.rebuildLines()
		}

	case m.keys.Comment:
		if key, ok := m.currentView().CurrentLineKey(); ok {
			existing := m.comments[key]
			cmd := m.editor.Open(key, existing)
			return m, cmd
		}

	case m.keys.DelComment:
		if key, ok := m.currentView().CurrentLineKey(); ok {
			delete(m.comments, key)
		}

	case m.keys.Copy:
		if err := internal.CopyToClipboard(m.files, m.comments); err == nil {
			m.copied = true
		}

	case m.keys.ToggleMode:
		if m.mode == ModeInline {
			m.mode = ModeSplit
		} else {
			m.mode = ModeInline
		}

	case m.keys.PaneLeft:
		if m.mode == ModeSplit {
			m.split.TogglePane(0)
		}
	case m.keys.PaneRight:
		if m.mode == ModeSplit {
			m.split.TogglePane(1)
		}

	case m.keys.Help:
		m.showHelp = !m.showHelp
	}

	return m, nil
}

type diffView interface {
	MoveUp(int)
	MoveDown(int)
	GoTop()
	GoBottom()
	CurrentLineKey() (diff.LineKey, bool)
}

func (m *Model) currentView() diffView {
	if m.mode == ModeSplit {
		return &m.split
	}
	return &m.inline
}

// View implements tea.Model.
func (m Model) View() tea.View {
	if m.err != nil {
		content := styleRemoved.Render("Error: "+m.err.Error()) + "\n\n" +
			styleContext.Render("Press q to quit")
		v := tea.NewView(content)
		v.AltScreen = true
		return v
	}

	if len(m.files) == 0 {
		content := styleContext.Render("No diff to display") + "\n\n" +
			styleContext.Render("Press q to quit")
		v := tea.NewView(content)
		v.AltScreen = true
		return v
	}

	var sb strings.Builder

	// File tabs
	sb.WriteString(m.renderFileTabs())
	sb.WriteString("\n")

	// Diff content
	if m.mode == ModeInline {
		sb.WriteString(m.inline.View(m.comments))
	} else {
		sb.WriteString(m.split.View(m.comments))
	}

	// Comment editor
	if m.editor.active {
		sb.WriteString(m.editor.View())
		sb.WriteString("\n")
	}

	// Status bar
	sb.WriteString(m.renderStatusBar())

	// Help overlay
	if m.showHelp {
		sb.WriteString("\n" + m.renderHelp())
	}

	v := tea.NewView(sb.String())
	v.AltScreen = true
	return v
}

func (m Model) renderFileTabs() string {
	var tabs []string
	for i, f := range m.files {
		name := f.DisplayName()
		commentCount := 0
		for key := range m.comments {
			if key.FileIndex == i {
				commentCount++
			}
		}
		if commentCount > 0 {
			name += fmt.Sprintf(" (%d)", commentCount)
		}

		if i == m.fileIdx {
			tabs = append(tabs, styleFileTabActive.Render(name))
		} else {
			tabs = append(tabs, styleFileTab.Render(name))
		}
	}
	return strings.Join(tabs, styleContext.Render("│"))
}

func (m Model) renderStatusBar() string {
	modeStr := "INLINE"
	if m.mode == ModeSplit {
		modeStr = "SPLIT"
	}

	left := styleStatusBar.Render(fmt.Sprintf(" rikugan │ %s ", modeStr))

	info := fmt.Sprintf(" %d/%d files │ %d comments ",
		m.fileIdx+1, len(m.files), len(m.comments))

	if m.copied {
		info += "│ Copied! "
	}

	right := styleStatusBarSection.Render(info)

	gap := m.width - lipgloss.Width(left) - lipgloss.Width(right)
	if gap < 0 {
		gap = 0
	}

	return left + strings.Repeat(" ", gap) + right
}

func (m Model) renderHelp() string {
	help := []string{
		"j/k: up/down  Ctrl+d/u: half page  gg/G: top/bottom",
		"Tab/Shift+Tab: next/prev file  V: toggle split",
		"c: comment  d: delete comment  y: copy to clipboard",
		"h/l: switch pane (split mode)  q: quit  ?: toggle help",
	}
	var sb strings.Builder
	for _, line := range help {
		sb.WriteString(styleHelp.Render("  " + line))
		sb.WriteString("\n")
	}
	return sb.String()
}
