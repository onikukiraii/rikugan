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
	editor     CommentEditor
	filePicker FilePicker
	keys       KeyMap
	width     int
	height    int
	err       error
	showHelp  bool
	copied       bool
	copiedMsg    string
	gPressed bool // for gg detection

	expandedFolds  map[int]map[int][]diff.DiffLine // fileIdx -> foldIdx -> lines
	fileLinesCache map[int][]string                 // fileIdx -> file lines

	loader       Loader
	diffSig      string
	watchEnabled bool
}

// New creates a new Model.
func New(files []diff.DiffFile, loader Loader) Model {
	m := Model{
		files:          files,
		comments:       make(map[diff.LineKey]string),
		editor:         NewCommentEditor(),
		filePicker:     NewFilePicker(),
		keys:           DefaultKeyMap(),
		inline:         NewInlineModel(),
		split:          NewSplitModel(),
		expandedFolds:  make(map[int]map[int][]diff.DiffLine),
		fileLinesCache: make(map[int][]string),
		loader:         loader,
		diffSig:        diffSignature(files),
		watchEnabled:   true,
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
	folds := m.expandedFolds[m.fileIdx]
	totalLines := m.getTotalLines(m.fileIdx)
	m.inline.BuildLines(f, m.fileIdx, folds, totalLines)
	m.split.BuildLines(f, m.fileIdx, folds, totalLines)
}

func (m *Model) getTotalLines(fileIdx int) int {
	if lines, ok := m.fileLinesCache[fileIdx]; ok {
		return len(lines)
	}
	return -1
}

func (m *Model) getFileLines(fileIdx int) ([]string, error) {
	if lines, ok := m.fileLinesCache[fileIdx]; ok {
		return lines, nil
	}
	path := m.files[fileIdx].DisplayName()
	lines, err := diff.ReadFileLines(path)
	if err != nil {
		return nil, err
	}
	m.fileLinesCache[fileIdx] = lines
	return lines, nil
}

// Init implements tea.Model.
func (m Model) Init() tea.Cmd {
	if m.watchEnabled {
		return watchForChanges(m.loader, m.diffSig)
	}
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

	// Handle file picker input
	if m.filePicker.active {
		idx, confirmed, cmd := m.filePicker.Update(msg)
		if confirmed && idx >= 0 {
			m.fileIdx = idx
			m.rebuildLines()
		}
		return m, cmd
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

	case fileCheckMsg:
		if msg.files != nil && msg.sig != m.diffSig {
			if !m.editor.active && !m.filePicker.active {
				m.reloadFiles(msg.files, msg.sig)
			}
		}
		return m, watchForChanges(m.loader, m.diffSig)

	case reloadResultMsg:
		if msg.files != nil && msg.sig != m.diffSig {
			m.reloadFiles(msg.files, msg.sig)
			m.copied = true
			m.copiedMsg = "Reloaded!"
		} else {
			m.copied = true
			m.copiedMsg = "No changes"
		}

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
	case m.keys.Quit, "esc", "ctrl+c":
		if m.showHelp {
			m.showHelp = false
			return m, nil
		}
		return m, tea.Quit

	case m.keys.Down, "down":
		m.currentView().MoveDown(1)
	case m.keys.Up, "up":
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
		} else {
			m.fileIdx = 0
		}
		m.rebuildLines()
	case m.keys.PrevFile:
		if m.fileIdx > 0 {
			m.fileIdx--
		} else {
			m.fileIdx = len(m.files) - 1
		}
		m.rebuildLines()

	case m.keys.NextHunk:
		m.currentView().NextHunk()
	case m.keys.PrevHunk:
		m.currentView().PrevHunk()

	case m.keys.ExpandFold:
		m.expandFold()

	case m.keys.Reload:
		return m, manualReload(m.loader)

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
			m.copiedMsg = "Copied (full)!"
		}

	case m.keys.CopySummary:
		if err := internal.CopyCommentsOnly(m.files, m.comments); err == nil {
			m.copied = true
			m.copiedMsg = "Copied (comments only)!"
		}

	case m.keys.ToggleMode:
		if m.mode == ModeInline {
			m.mode = ModeSplit
		} else {
			m.mode = ModeInline
		}

	case m.keys.PaneLeft, "left":
		if m.mode == ModeSplit {
			m.split.TogglePane(0)
		}
	case m.keys.PaneRight, "right":
		if m.mode == ModeSplit {
			m.split.TogglePane(1)
		}

	case m.keys.Help:
		m.showHelp = !m.showHelp

	case "space":
		cmd := m.filePicker.Open(m.files, m.comments)
		return m, cmd
	}

	return m, nil
}

func (m *Model) reloadFiles(newFiles []diff.DiffFile, sig string) {
	// Preserve current file by name
	currentName := ""
	if m.fileIdx < len(m.files) {
		currentName = m.files[m.fileIdx].DisplayName()
	}

	m.files = newFiles
	m.diffSig = sig
	m.fileIdx = 0

	for i, f := range newFiles {
		if f.DisplayName() == currentName {
			m.fileIdx = i
			break
		}
	}

	// Clear caches since content changed
	m.expandedFolds = make(map[int]map[int][]diff.DiffLine)
	m.fileLinesCache = make(map[int][]string)

	if len(m.files) > 0 {
		m.rebuildLines()
	}
}

func (m *Model) expandFold() {
	// Check if cursor is on a fold line
	var foldIdx int
	var isFold bool

	if m.mode == ModeInline {
		if m.inline.cursor >= 0 && m.inline.cursor < len(m.inline.lines) {
			line := m.inline.lines[m.inline.cursor]
			if line.isFold {
				foldIdx = line.foldIndex
				isFold = true
			}
		}
	} else {
		if m.split.cursor >= 0 && m.split.cursor < len(m.split.leftPane) {
			line := m.split.leftPane[m.split.cursor]
			if line.isFold {
				foldIdx = line.foldIndex
				isFold = true
			}
		}
	}

	if !isFold {
		return
	}

	file := m.files[m.fileIdx]
	if len(file.Hunks) == 0 {
		return
	}

	// Read file lines (cached)
	fileLines, err := m.getFileLines(m.fileIdx)
	if err != nil {
		return
	}

	// Calculate line range for this fold
	var newStart, newEnd int // 1-based, inclusive
	var oldStart int

	if foldIdx == 0 {
		newStart = 1
		newEnd = file.Hunks[0].NewStart - 1
		oldStart = 1
	} else if foldIdx < len(file.Hunks) {
		prevH := file.Hunks[foldIdx-1]
		newStart = prevH.NewStart + prevH.NewCount
		newEnd = file.Hunks[foldIdx].NewStart - 1
		oldStart = prevH.OldStart + prevH.OldCount
	} else {
		// After last hunk
		lastH := file.Hunks[len(file.Hunks)-1]
		newStart = lastH.NewStart + lastH.NewCount
		newEnd = len(fileLines)
		oldStart = lastH.OldStart + lastH.OldCount
	}

	if newStart > newEnd {
		return
	}

	// Create DiffLines for the expanded content
	var lines []diff.DiffLine
	oldNum := oldStart
	for lineNum := newStart; lineNum <= newEnd; lineNum++ {
		idx := lineNum - 1
		content := ""
		if idx < len(fileLines) {
			content = strings.ReplaceAll(fileLines[idx], "\t", "    ")
		}
		lines = append(lines, diff.DiffLine{
			Type:    diff.LineContext,
			Content: content,
			OldNum:  oldNum,
			NewNum:  lineNum,
		})
		oldNum++
	}

	// Store expanded fold
	if m.expandedFolds[m.fileIdx] == nil {
		m.expandedFolds[m.fileIdx] = make(map[int][]diff.DiffLine)
	}
	m.expandedFolds[m.fileIdx][foldIdx] = lines

	// Save cursor positions
	inlineCursor := m.inline.cursor
	splitCursor := m.split.cursor

	m.rebuildLines()

	// Restore cursor positions
	m.inline.cursor = inlineCursor
	if m.inline.cursor >= len(m.inline.lines) {
		m.inline.cursor = len(m.inline.lines) - 1
	}
	m.inline.ensureVisible()
	m.split.cursor = splitCursor
	if m.split.cursor >= len(m.split.leftPane) {
		m.split.cursor = len(m.split.leftPane) - 1
	}
	m.split.ensureVisible()
}

type diffView interface {
	MoveUp(int)
	MoveDown(int)
	GoTop()
	GoBottom()
	NextHunk()
	PrevHunk()
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

	// File picker replaces diff content when active
	if m.filePicker.active {
		sb.WriteString(m.filePicker.View(m.width))
	} else if m.mode == ModeInline {
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
		info += "│ " + m.copiedMsg + " "
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
		"j/k/arrows: up/down  Ctrl+d/u: half page  gg/G: top/bottom  ]/[: next/prev hunk",
		"Tab/Shift+Tab: next/prev file  Space: fuzzy find file",
		"Enter: expand hidden lines  c: comment  d: delete comment",
		"y: copy comments only  Y: copy diff+comments",
		"V: toggle split  h/l/arrows: switch pane (split mode)",
		"r: reload  q: quit  ?: toggle help",
	}
	var sb strings.Builder
	for _, line := range help {
		sb.WriteString(styleHelp.Render("  " + line))
		sb.WriteString("\n")
	}
	return sb.String()
}
